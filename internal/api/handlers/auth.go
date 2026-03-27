package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/opencrm/opencrm/pkg/crypto"
	"github.com/opencrm/opencrm/pkg/errs"
)

type AuthHandler struct {
	userRepo        UserRepository
	tenantRepo      TenantRepository
	roleRepo        RoleRepository
	crypto          *crypto.Crypto
	tokenManager    *crypto.TokenManager
	redis           RedisClient
}

type UserRepository interface {
	Create(ctx interface{ Context() interface{} }, tenantID uuid.UUID, email, passwordHash, firstName, lastName string) (interface{}, error)
	GetByEmail(ctx interface{ Context() interface{} }, tenantID uuid.UUID, email string) (interface{}, error)
	UpdateLastLogin(ctx interface{ Context() interface{} }, id uuid.UUID) error
}

type TenantRepository interface {
	Create(ctx interface{ Context() interface{} }, slug, name string) (interface{}, error)
	GetBySlug(ctx interface{ Context() interface{} }, slug string) (interface{}, error)
	SeedRoles(ctx interface{ Context() interface{} }, tenantID uuid.UUID) error
	SeedPipeline(ctx interface{ Context() interface{} }, tenantID uuid.UUID) error
}

type RoleRepository interface {
	GetByName(ctx interface{ Context() interface{} }, tenantID uuid.UUID, name string) (interface{}, error)
	AssignToUser(ctx interface{ Context() interface{} }, userID uuid.UUID, roleIDs []uuid.UUID) error
}

type RedisClient interface {
	Set(ctx interface{ Context() interface{} }, key string, value interface{}, expiration time.Duration) error
	Get(ctx interface{ Context() interface{} }, key string) (string, error)
	Del(ctx interface{ Context() interface{} }, keys ...string) error
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1"`
	LastName  string `json:"last_name" validate:"required,min=1"`
	TenantName string `json:"tenant_name" validate:"required,min=2"`
	TenantSlug string `json:"tenant_slug" validate:"required,min=2,max=63"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt   int64  `json:"expires_at"`
	TokenType   string `json:"token_type"`
	User        *UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Roles     []string  `json:"roles"`
}

func NewAuthHandler(
	userRepo UserRepository,
	tenantRepo TenantRepository,
	roleRepo RoleRepository,
	crypto *crypto.Crypto,
	tokenManager *crypto.TokenManager,
	redis RedisClient,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		tenantRepo:   tenantRepo,
		roleRepo:     roleRepo,
		crypto:       crypto,
		tokenManager: tokenManager,
		redis:        redis,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	exists, _ := h.tenantRepo.GetBySlug(c.Request().Context(), req.TenantSlug)
	if exists != nil {
		return errs.Conflict("Tenant slug already taken").HTTPError(c)
	}

	tenant, err := h.tenantRepo.Create(c.Request().Context(), req.TenantSlug, req.TenantName)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	tenantID := tenant.(interface{ GetID() uuid.UUID }).GetID()

	if err := h.tenantRepo.SeedRoles(c.Request().Context(), tenantID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.tenantRepo.SeedPipeline(c.Request().Context(), tenantID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	ownerRole, err := h.roleRepo.GetByName(c.Request().Context(), tenantID, "Owner")
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}
	roleID := ownerRole.(interface{ GetID() uuid.UUID }).GetID()

	user, err := h.userRepo.Create(c.Request().Context(), tenantID, req.Email, "", req.FirstName, req.LastName)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.roleRepo.AssignToUser(c.Request().Context(), user.(interface{ GetID() uuid.UUID }).GetID(), []uuid.UUID{roleID})

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Registration successful",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	user, err := h.userRepo.GetByEmail(c.Request().Context(), uuid.Nil, req.Email)
	if err != nil {
		return errs.Unauthorized("Invalid credentials")
	}

	u := user.(interface {
		GetID() uuid.UUID
		GetTenantID() uuid.UUID
		GetPasswordHash() string
		GetEmail() string
		GetFirstName() string
		GetLastName() string
	})

	if !h.crypto.VerifyPassword(req.Password, u.GetPasswordHash()) {
		return errs.Unauthorized("Invalid credentials")
	}

	tenant, err := h.tenantRepo.GetBySlug(c.Request().Context(), "")
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}
	_ = tenant

	roles, _ := h.roleRepo.GetByName(c.Request().Context(), uuid.Nil, "")
	roleNames := []string{}
	if roles != nil {
		roleNames = append(roleNames, roles.(interface{ GetName() string }).GetName())
	}

	payload := crypto.TokenPayload{
		UserID:   u.GetID().String(),
		TenantID: u.GetTenantID().String(),
		Email:    u.GetEmail(),
		Roles:    roleNames,
	}

	tokens, err := h.tokenManager.GenerateTokenPair(payload, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.redis.Set(c.Request().Context(), "refresh:"+u.GetID().String(), tokens.RefreshToken, 7*24*time.Hour); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.userRepo.UpdateLastLogin(c.Request().Context(), u.GetID())

	return c.JSON(http.StatusOK, &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:   tokens.ExpiresAt,
		TokenType:   tokens.TokenType,
		User: &UserResponse{
			ID:        u.GetID(),
			TenantID:  u.GetTenantID(),
			Email:     u.GetEmail(),
			FirstName: u.GetFirstName(),
			LastName:  u.GetLastName(),
			Roles:     roleNames,
		},
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	payload, err := h.tokenManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return errs.Unauthorized("Invalid refresh token")
	}

	userID, _ := uuid.Parse(payload.UserID)
	storedToken, err := h.redis.Get(c.Request().Context(), "refresh:"+userID.String())
	if err != nil || storedToken != req.RefreshToken {
		return errs.Unauthorized("Invalid refresh token")
	}

	tokens, err := h.tokenManager.RefreshTokens(req.RefreshToken)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.redis.Set(c.Request().Context(), "refresh:"+userID.String(), tokens.RefreshToken, 7*24*time.Hour); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:   tokens.ExpiresAt,
		TokenType:   tokens.TokenType,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	payload := c.Get("payload").(*crypto.TokenPayload)
	if payload == nil {
		return errs.Unauthorized("Not authenticated").HTTPError(c)
	}

	userID, _ := uuid.Parse(payload.UserID)
	_ = h.redis.Del(c.Request().Context(), "refresh:"+userID.String())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	payload := c.Get("payload").(*crypto.TokenPayload)
	if payload == nil {
		return errs.Unauthorized("Not authenticated").HTTPError(c)
	}

	userID, _ := uuid.Parse(payload.UserID)
	user, err := h.userRepo.GetByEmail(c.Request().Context(), uuid.Nil, payload.Email)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	u := user.(interface {
		GetID() uuid.UUID
		GetTenantID() uuid.UUID
		GetEmail() string
		GetFirstName() string
		GetLastName() string
	})

	return c.JSON(http.StatusOK, &UserResponse{
		ID:        u.GetID(),
		TenantID:  u.GetTenantID(),
		Email:     u.GetEmail(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Roles:     payload.Roles,
	})
}

func (e *Error) HTTPError(c echo.Context) error {
	return c.JSON(e.HTTPStatus(), map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
			"details": e.Details,
		},
	})
}
