package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/pkg/crypto"
	"github.com/oscar/oscar/pkg/errs"
)

type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyTenantID  contextKey = "tenant_id"
	ContextKeyPayload   contextKey = "payload"
	ContextKeyRoles     contextKey = "roles"
	ContextKeyRequestID contextKey = "request_id"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			c.Response().Header().Set("X-Request-ID", requestID)
			c.Set(string(ContextKeyRequestID), requestID)
			return next(c)
		}
	}
}

func Auth(tokenManager *crypto.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.ErrUnauthorized
			}

			token := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}

			payload, err := tokenManager.ValidateToken(token)
			if err != nil {
				return echo.ErrUnauthorized
			}

			c.Set(string(ContextKeyPayload), payload)

			userID, _ := uuid.Parse(payload.UserID)
			c.Set(string(ContextKeyUserID), userID)

			tenantID, _ := uuid.Parse(payload.TenantID)
			c.Set(string(ContextKeyTenantID), tenantID)

			c.Set(string(ContextKeyRoles), payload.Roles)

			return next(c)
		}
	}
}

func TenantResolver(tenantRepo TenantRepository, pool TenantPool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := c.Get(string(ContextKeyTenantID)).(uuid.UUID)
			if tenantID == uuid.Nil {
				return echo.ErrUnauthorized
			}

			_, err := tenantRepo.GetByID(context.Background(), tenantID)
			if err != nil {
				return echo.ErrForbidden
			}

			ctxWithTx, tx, err := pool.SetTenantContext(c.Request().Context(), tenantID)
			if err != nil {
				return errs.Internal(err).HTTPError(c)
			}
			c.SetRequest(c.Request().WithContext(ctxWithTx))

			err = next(c)
			if err != nil {
				tx.Rollback(ctxWithTx)
				return err
			}
			tx.Commit(ctxWithTx)
			return nil
		}
	}
}

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles, ok := c.Get(string(ContextKeyRoles)).([]string)
			if !ok {
				return echo.ErrForbidden
			}

			for _, required := range roles {
				for _, userRole := range userRoles {
					if userRole == required {
						return next(c)
					}
				}
			}

			return echo.ErrForbidden
		}
	}
}

type rateLimitEntry struct {
	count    int64
	expireAt time.Time
}

type InMemoryRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	maxReqs int
	window  time.Duration
	redis   RedisClient
	redisOK bool
}

func NewRateLimiter(redis RedisClient, maxRequests int, window time.Duration) *InMemoryRateLimiter {
	rl := &InMemoryRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		maxReqs: maxRequests,
		window:  window,
		redis:   redis,
		redisOK: redis != nil,
	}

	go rl.cleanupExpired()

	return rl
}

func (rl *InMemoryRateLimiter) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.After(entry.expireAt) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *InMemoryRateLimiter) Allow(clientIP string) (bool, int64, error) {
	if rl.redis != nil && rl.redisOK {
		key := "ratelimit:" + clientIP
		count, err := rl.redis.Incr(context.Background(), key).Result()
		if err != nil {
			rl.redisOK = false
			fmt.Printf("[RateLimit] Redis error, switching to in-memory: %v\n", err)
			return rl.allowInMemory(clientIP)
		}

		if count == 1 {
			rl.redis.Expire(context.Background(), key, rl.window)
		}

		if count > int64(rl.maxReqs) {
			return false, count, nil
		}
		return true, count, nil
	}

	return rl.allowInMemory(clientIP)
}

func (rl *InMemoryRateLimiter) allowInMemory(clientIP string) (bool, int64, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := "ratelimit:" + clientIP
	now := time.Now()

	entry, exists := rl.entries[key]
	if !exists || now.After(entry.expireAt) {
		rl.entries[key] = &rateLimitEntry{
			count:    1,
			expireAt: now.Add(rl.window),
		}
		return true, 1, nil
	}

	entry.count++
	return entry.count <= int64(rl.maxReqs), entry.count, nil
}

func RateLimitMiddleware(limiter *InMemoryRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()

			allowed, count, err := limiter.Allow(clientIP)
			if err != nil {
				fmt.Printf("[RateLimit] Error: %v\n", err)
				return next(c)
			}

			remaining := int64(limiter.maxReqs) - count
			if remaining < 0 {
				remaining = 0
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.maxReqs))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Response().Header().Set("X-RateLimit-Window", limiter.window.String())

			if !allowed {
				c.Response().Header().Set("Retry-After", "60")
				return echo.ErrTooManyRequests
			}

			return next(c)
		}
	}
}

func RateLimit(redis RedisClient, maxRequests int, window time.Duration) echo.MiddlewareFunc {
	limiter := NewRateLimiter(redis, maxRequests, window)
	return RateLimitMiddleware(limiter)
}

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					c.Error(echo.ErrInternalServerError)
				}
			}()
			return next(c)
		}
	}
}

type TenantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error)
}

type TenantPool interface {
	SetTenantContext(ctx context.Context, tenantID uuid.UUID) (context.Context, pgx.Tx, error)
}

type RedisClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}
