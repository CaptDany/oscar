package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/invitation"
	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/internal/email"
	"github.com/oscar/oscar/pkg/crypto"
	"github.com/oscar/oscar/pkg/errs"
	"github.com/oscar/oscar/pkg/validator"
)

type AuthHandler struct {
	userRepo       user.Repository
	tenantRepo     tenant.Repository
	roleRepo       user.RoleRepository
	invitationRepo invitation.Repository
	crypto         *crypto.Crypto
	tokenManager   *crypto.TokenManager
	emailClient    *email.EmailClient
	baseURL        string
	frontendURL    string
}

type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	FirstName       string `json:"first_name" validate:"required,min=1"`
	LastName        string `json:"last_name"`
	TenantName      string `json:"tenant_name" validate:"omitempty,min=2"`
	TenantSlug      string `json:"tenant_slug" validate:"omitempty,min=2,max=63"`
	InvitationToken string `json:"invitation_token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int64         `json:"expires_at"`
	TokenType    string        `json:"token_type"`
	User         *UserResponse `json:"user"`
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

func NewAuthHandlerWithInvitations(
	userRepo user.Repository,
	tenantRepo tenant.Repository,
	roleRepo user.RoleRepository,
	invitationRepo invitation.Repository,
	cryptoSvc *crypto.Crypto,
	tokenManager *crypto.TokenManager,
	emailClient *email.EmailClient,
	baseURL string,
	frontendURL string,
) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		roleRepo:       roleRepo,
		invitationRepo: invitationRepo,
		crypto:         cryptoSvc,
		tokenManager:   tokenManager,
		emailClient:    emailClient,
		baseURL:        baseURL,
		frontendURL:    frontendURL,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_FAILED",
				"message": "Validation failed",
				"details": validationErrors,
			},
		})
	}

	var t *tenant.Tenant
	var err error
	var isNewTenant bool

	if req.InvitationToken != "" {
		return h.registerWithInvitation(c, req)
	}

	if req.TenantSlug == "" || req.TenantName == "" {
		return errs.BadRequest("Tenant name and slug are required for new registrations").HTTPError(c)
	}

	existingTenant, err := h.tenantRepo.GetBySlug(c.Request().Context(), req.TenantSlug)
	if err != nil {
		createTenantReq := &tenant.CreateTenantRequest{
			Slug: req.TenantSlug,
			Name: req.TenantName,
		}
		t, err = h.tenantRepo.Create(c.Request().Context(), createTenantReq)
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
		isNewTenant = true

		if err := h.tenantRepo.SeedRoles(c.Request().Context(), t.ID); err != nil {
			return errs.Internal(err).HTTPError(c)
		}

		if err := h.tenantRepo.SeedPipeline(c.Request().Context(), t.ID); err != nil {
			return errs.Internal(err).HTTPError(c)
		}
	} else {
		t = existingTenant
	}

	passwordHash, err := h.crypto.HashPassword(req.Password)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	createUserReq := &user.CreateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	u, err := h.userRepo.Create(c.Request().Context(), t.ID, createUserReq, passwordHash)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	var defaultRoleName string
	if isNewTenant {
		defaultRoleName = "Owner"
	} else {
		existingUsers, _, _ := h.userRepo.List(c.Request().Context(), t.ID, 1, 0)
		if len(existingUsers) == 0 {
			defaultRoleName = "Owner"
		} else {
			defaultRoleName = "Read Only"
		}
	}

	defaultRole, err := h.roleRepo.GetByName(c.Request().Context(), t.ID, defaultRoleName)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.roleRepo.AssignToUser(c.Request().Context(), u.ID, []uuid.UUID{defaultRole.ID})

	verificationToken, _ := crypto.GenerateSecureToken(32)

	verifyURL := fmt.Sprintf("%s/verify-email/%s", h.frontendURL, verificationToken)

	if h.emailClient != nil {
		go func() {
			ctx := context.Background()
			if err := h.userRepo.SetEmailVerificationToken(ctx, u.ID, verificationToken); err != nil {
				fmt.Printf("Failed to set verification token: %v\n", err)
				return
			}
			if err := h.emailClient.SendEmailVerification(u.Email, req.FirstName, verifyURL); err != nil {
				fmt.Printf("Failed to send verification email to %s: %v\n", u.Email, err)
				return
			}
			fmt.Printf("Verification email sent to %s\n", u.Email)
		}()
	}

	payload := crypto.TokenPayload{
		UserID:   u.ID.String(),
		TenantID: t.ID.String(),
		Email:    u.Email,
		Roles:    []string{defaultRole.Name},
	}

	tokens, err := h.tokenManager.GenerateTokenPair(payload, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, &AuthResponse{
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
			Roles:     []string{defaultRole.Name},
		},
	})
}

func (h *AuthHandler) registerWithInvitation(c echo.Context, req RegisterRequest) error {
	if h.invitationRepo == nil {
		return errs.Internal(nil).HTTPError(c)
	}

	inv, err := h.invitationRepo.GetByToken(c.Request().Context(), req.InvitationToken)
	if err != nil {
		return errs.NotFound("Invitation not found").HTTPError(c)
	}

	if !inv.IsValid() {
		if inv.IsExpired() {
			return errs.BadRequest("Invitation has expired").HTTPError(c)
		}
		if inv.IsAccepted() {
			return errs.BadRequest("Invitation has already been used").HTTPError(c)
		}
		return errs.BadRequest("Invitation is no longer valid").HTTPError(c)
	}

	if inv.Email != req.Email {
		return errs.BadRequest("Email does not match invitation").HTTPError(c)
	}

	t, err := h.tenantRepo.GetByID(c.Request().Context(), inv.TenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	passwordHash, err := h.crypto.HashPassword(req.Password)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	createUserReq := &user.CreateUserRequest{
		Email: req.Email,
	}
	u, err := h.userRepo.Create(c.Request().Context(), inv.TenantID, createUserReq, passwordHash)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	assignedRole, err := h.roleRepo.GetByName(c.Request().Context(), inv.TenantID, inv.RoleName)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	_ = h.roleRepo.AssignToUser(c.Request().Context(), u.ID, []uuid.UUID{assignedRole.ID})

	_ = h.invitationRepo.MarkAccepted(c.Request().Context(), inv.ID)

	payload := crypto.TokenPayload{
		UserID:   u.ID.String(),
		TenantID: t.ID.String(),
		Email:    u.Email,
		Roles:    []string{assignedRole.Name},
	}

	tokens, err := h.tokenManager.GenerateTokenPair(payload, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		TokenType:    tokens.TokenType,
		User: &UserResponse{
			ID:        u.ID,
			TenantID:  u.TenantID,
			Email:     u.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Roles:     []string{assignedRole.Name},
		},
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_FAILED",
				"message": "Validation failed",
				"details": validationErrors,
			},
		})
	}

	u, err := h.userRepo.GetByEmail(c.Request().Context(), uuid.Nil, req.Email)
	if err != nil {
		return errs.Unauthorized("Invalid credentials").HTTPError(c)
	}

	if !h.crypto.VerifyPassword(req.Password, u.PasswordHash) {
		return errs.Unauthorized("Invalid credentials").HTTPError(c)
	}

	if u.EmailVerifiedAt == nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "EMAIL_NOT_VERIFIED",
				"message": "Please verify your email address before logging in",
			},
		})
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
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
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

func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return errs.BadRequest("Verification token is required").HTTPError(c)
	}

	u, err := h.userRepo.GetByVerificationToken(c.Request().Context(), token)
	if err != nil {
		return errs.BadRequest("Invalid or expired verification link").HTTPError(c)
	}

	if u.EmailVerifiedAt != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Email already verified",
		})
	}

	if err := h.userRepo.VerifyEmail(c.Request().Context(), u.ID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
	})
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *AuthHandler) ResendVerification(c echo.Context) error {
	var req ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_FAILED",
				"message": "Validation failed",
				"details": validationErrors,
			},
		})
	}

	u, err := h.userRepo.GetByEmail(c.Request().Context(), uuid.Nil, req.Email)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "If an account with that email exists, a verification email has been sent",
		})
	}

	if u.EmailVerifiedAt != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Email already verified",
		})
	}

	verificationToken, _ := crypto.GenerateSecureToken(32)
	verifyURL := fmt.Sprintf("%s/verify-email/%s", h.frontendURL, verificationToken)

	if h.emailClient != nil {
		go func() {
			ctx := context.Background()
			if err := h.userRepo.SetEmailVerificationToken(ctx, u.ID, verificationToken); err != nil {
				fmt.Printf("Failed to set verification token: %v\n", err)
				return
			}
			if err := h.emailClient.SendEmailVerification(u.Email, u.FirstName, verifyURL); err != nil {
				fmt.Printf("Failed to send verification email to %s: %v\n", u.Email, err)
				return
			}
			fmt.Printf("Verification email sent to %s\n", u.Email)
		}()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "If an account with that email exists, a verification email has been sent",
	})
}
