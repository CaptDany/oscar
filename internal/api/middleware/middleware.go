package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/pkg/crypto"
)

type contextKey string

const (
	ContextKeyUserID  contextKey = "user_id"
	ContextKeyTenantID contextKey = "tenant_id"
	ContextKeyPayload  contextKey = "payload"
	ContextKeyRoles    contextKey = "roles"
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

func TenantResolver(tenantRepo TenantRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := c.Get(string(ContextKeyTenantID)).(uuid.UUID)
			if tenantID == uuid.Nil {
				return echo.ErrUnauthorized
			}

			ctx := c.Request().Context()

			_, err := tenantRepo.GetByID(ctx, tenantID)
			if err != nil {
				return echo.ErrForbidden
			}

			return next(c)
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

func RateLimit(redis RedisClient, maxRequests int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()
			key := "ratelimit:" + clientIP

			count, err := redis.Incr(c.Request().Context(), key)
			if err != nil {
				return next(c)
			}

			if count == 1 {
				redis.Expire(c.Request().Context(), key, window)
			}

			if count > int64(maxRequests) {
				return echo.ErrTooManyRequests
			}

			c.Response().Header().Set("X-RateLimit-Limit", string(rune(maxRequests)))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(maxRequests-int(count))))

			return next(c)
		}
	}
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
	GetByID(ctx interface{ Context() interface{} }, id uuid.UUID) (interface{}, error)
}

type RedisClient interface {
	Incr(ctx interface{ Context() interface{} }, key string) (int64, error)
	Expire(ctx interface{ Context() interface{} }, key string, duration time.Duration) error
}
