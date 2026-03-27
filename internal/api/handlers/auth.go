package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/pkg/crypto"
	"github.com/oscar/oscar/pkg/errs"
)

type AuthHandler struct {
	userRepo     user.Repository
	tenantRepo   tenant.Repository
	roleRepo     user.RoleRepository
	crypto       *crypto.Crypto
	tokenManager *crypto.TokenManager
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required,min=1"`
	LastName    string `json:"last_name" validate:"required,min=1"`
	TenantName  string `json:"tenant_name" validate:"required,min=2"`
	TenantSlug  string `json:"tenant_slug" validate:"required,min=2,max=63"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    int64           `json:"expires_at"`
	TokenType    string          `json:"token_type"`
	User         *UserResponse   `json:"user"`
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
	userRepo user.Repository,
	tenantRepo tenant.Repository,
	roleRepo user.RoleRepository,
	cryptoSvc *crypto.Crypto,
	tokenManager *crypto.TokenManager,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		tenantRepo:   tenantRepo,
		roleRepo:     roleRepo,
		crypto:       cryptoSvc,
		tokenManager: tokenManager,
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

	_, err := h.tenantRepo.GetBySlug(c.Request().Context(), req.TenantSlug)
	if err == nil {
		return errs.Conflict("Tenant slug already taken").HTTPError(c)
	}

	createTenantReq := &tenant.CreateTenantRequest{
		Slug: req.TenantSlug,
		Name: req.TenantName,
	}
	t, err := h.tenantRepo.Create(c.Request().Context(), createTenantReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.tenantRepo.SeedRoles(c.Request().Context(), t.ID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.tenantRepo.SeedPipeline(c.Request().Context(), t.ID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	ownerRole, err := h.roleRepo.GetByName(c.Request().Context(), t.ID, "Owner")
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	createUserReq := &user.CreateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	u, err := h.userRepo.Create(c.Request().Context(), t.ID, createUserReq, "")
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.roleRepo.AssignToUser(c.Request().Context(), u.ID, []uuid.UUID{ownerRole.ID})

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

	u, err := h.userRepo.GetByEmail(c.Request().Context(), uuid.Nil, req.Email)
	if err != nil {
		return errs.Unauthorized("Invalid credentials").HTTPError(c)
	}

	if !h.crypto.VerifyPassword(req.Password, u.PasswordHash) {
		return errs.Unauthorized("Invalid credentials").HTTPError(c)
	}

	roleNames, _ := h.roleRepo.GetUserRoleNames(c.Request().Context(), u.ID)

	payload := crypto.TokenPayload{
		UserID:   u.ID.String(),
		TenantID: u.TenantID.String(),
		Email:    u.Email,
		Roles:    roleNames,
	}

	tokens, err := h.tokenManager.GenerateTokenPair(payload, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.userRepo.UpdateLastLogin(c.Request().Context(), u.ID)

	return c.JSON(http.StatusOK, &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		TokenType:    tokens.TokenType,
		User: &UserResponse{
			ID:        u.ID,
			TenantID:  u.TenantID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
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

	tokens, err := h.tokenManager.RefreshTokens(req.RefreshToken)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	userID, _ := uuid.Parse(payload.UserID)
	_ = userID

	_ = tokens

	return c.JSON(http.StatusOK, &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		TokenType:    tokens.TokenType,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	payload := c.Get("payload").(*crypto.TokenPayload)
	if payload == nil {
		return errs.Unauthorized("Not authenticated").HTTPError(c)
	}

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
	u, err := h.userRepo.GetByID(c.Request().Context(), userID)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, &UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Roles:     payload.Roles,
	})
}
