package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/invitation"
	"github.com/oscar/oscar/internal/domain/tenant"
	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/pkg/crypto"
	"github.com/oscar/oscar/pkg/errs"
	"github.com/oscar/oscar/pkg/validator"
)

type InvitationHandler struct {
	invitationRepo invitation.Repository
	userRepo       user.Repository
	roleRepo       user.RoleRepository
	tenantRepo     tenant.Repository
	crypto         *crypto.Crypto
	emailSender    EmailSender
}

type EmailSender interface {
	SendInvitationEmail(to, inviterName, orgName, token string) error
	SendVerificationEmail(to, token string) error
}

type MockEmailSender struct{}

func (m *MockEmailSender) SendInvitationEmail(to, inviterName, orgName, token string) error {
	return nil
}

func (m *MockEmailSender) SendVerificationEmail(to, token string) error {
	return nil
}

type CreateInvitationRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	RoleName  string `json:"role_name" validate:"required,oneof=Admin Member Read Only Sales Manager"`
}

type ValidateInvitationResponse struct {
	Valid     bool   `json:"valid"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	OrgName   string `json:"org_name"`
	RoleName  string `json:"role_name"`
	ExpiresAt string `json:"expires_at"`
	Expired   bool   `json:"expired"`
	Accepted  bool   `json:"accepted"`
}

func NewInvitationHandler(
	invitationRepo invitation.Repository,
	userRepo user.Repository,
	roleRepo user.RoleRepository,
	tenantRepo tenant.Repository,
	cryptoSvc *crypto.Crypto,
	emailSender EmailSender,
) *InvitationHandler {
	return &InvitationHandler{
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		tenantRepo:     tenantRepo,
		crypto:         cryptoSvc,
		emailSender:    emailSender,
	}
}

func (h *InvitationHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateInvitationRequest
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

	existingUser, _ := h.userRepo.GetByEmail(c.Request().Context(), tenantID, req.Email)
	if existingUser != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "EMAIL_EXISTS",
				"message": "A user with this email already exists in this organization",
			},
		})
	}

	existingInvitation, _ := h.invitationRepo.GetByEmail(c.Request().Context(), tenantID, req.Email)
	if existingInvitation != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INVITATION_EXISTS",
				"message": "An invitation has already been sent to this email",
			},
		})
	}

	token, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	inv, err := h.invitationRepo.Create(c.Request().Context(), tenantID, userID, &invitation.CreateInvitationRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleName:  req.RoleName,
	}, token, expiresAt)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	t, _ := h.tenantRepo.GetByID(c.Request().Context(), tenantID)
	orgName := "oscar"
	if t != nil {
		orgName = t.Name
	}

	inviterUser, _ := h.userRepo.GetByID(c.Request().Context(), userID)
	inviterName := "A team member"
	if inviterUser != nil {
		inviterName = inviterUser.FullName()
	}

	_ = h.emailSender.SendInvitationEmail(req.Email, inviterName, orgName, token)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": map[string]interface{}{
			"id":         inv.ID,
			"email":      inv.Email,
			"first_name": inv.FirstName,
			"last_name":  inv.LastName,
			"role_name":  inv.RoleName,
			"expires_at": inv.ExpiresAt,
			"invite_url": "/invite/" + token,
		},
	})
}

func (h *InvitationHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	limit := 50
	offset := 0

	invitations, total, err := h.invitationRepo.ListByTenant(c.Request().Context(), tenantID, limit, offset)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": invitations,
		"meta": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *InvitationHandler) Delete(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid invitation ID").HTTPError(c)
	}

	inv, err := h.invitationRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Invitation not found").HTTPError(c)
	}

	if inv.TenantID != tenantID {
		return errs.Forbidden("You do not have permission to delete this invitation").HTTPError(c)
	}

	if err := h.invitationRepo.Delete(c.Request().Context(), id); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Invitation deleted successfully",
	})
}

func (h *InvitationHandler) Validate(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return errs.BadRequest("Token is required").HTTPError(c)
	}

	inv, err := h.invitationRepo.GetByToken(c.Request().Context(), token)
	if err != nil {
		return errs.NotFound("Invitation not found").HTTPError(c)
	}

	t, _ := h.tenantRepo.GetByID(c.Request().Context(), inv.TenantID)
	orgName := "oscar"
	if t != nil {
		orgName = t.Name
	}

	return c.JSON(http.StatusOK, &ValidateInvitationResponse{
		Valid:     inv.IsValid(),
		Email:     inv.Email,
		FirstName: inv.FirstName,
		LastName:  inv.LastName,
		OrgName:   orgName,
		RoleName:  inv.RoleName,
		ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
		Expired:   inv.IsExpired(),
		Accepted:  inv.IsAccepted(),
	})
}
