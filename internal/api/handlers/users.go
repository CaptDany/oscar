package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/internal/storage"
	"github.com/oscar/oscar/pkg/errs"
)

type UserHandler struct {
	userRepo user.Repository
	roleRepo user.RoleRepository
	storage  *storage.MinIOClient
}

func NewUserHandler(userRepo user.Repository, roleRepo user.RoleRepository, storage *storage.MinIOClient) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		roleRepo: roleRepo,
		storage:  storage,
	}
}

type UpdateUserRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" validate:"required,min=1"`
}

func (h *UserHandler) getAvatarURL(ctx echo.Context, objectKey *string) *string {
	if objectKey == nil || *objectKey == "" {
		return nil
	}
	url, err := h.storage.GetPresignedURL(ctx.Request().Context(), *objectKey, 7*24*time.Hour)
	if err != nil {
		return nil
	}
	return &url
}

func (h *UserHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	limit := 20
	offset := 0

	users, total, err := h.userRepo.List(c.Request().Context(), tenantID, limit, offset)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	var response []map[string]interface{}
	for _, u := range users {
		roles, _ := h.roleRepo.GetUserRoles(c.Request().Context(), u.ID)
		var roleNames []string
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
		}

		response = append(response, map[string]interface{}{
			"id":         u.ID,
			"tenant_id":  u.TenantID,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"avatar_url": h.getAvatarURL(c, u.AvatarURL),
			"is_active":  u.IsActive,
			"created_at": u.CreatedAt,
			"roles":      roleNames,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   response,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) Get(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	u, err := h.userRepo.GetByID(c.Request().Context(), userID)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	roles, _ := h.roleRepo.GetUserRoles(c.Request().Context(), u.ID)
	var roleNames []string
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         u.ID,
		"tenant_id":  u.TenantID,
		"email":      u.Email,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"avatar_url": h.getAvatarURL(c, u.AvatarURL),
		"timezone":   u.Timezone,
		"locale":     u.Locale,
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"roles":      roleNames,
	})
}

func (h *UserHandler) UpdateRoles(c echo.Context) error {
	requestingUserID := c.Get("user_id").(uuid.UUID)
	requestingRoles := c.Get("roles").([]string)

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	var req UpdateUserRolesRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	targetUser, err := h.userRepo.GetByID(c.Request().Context(), targetUserID)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	isOwnerOrAdmin := false
	for _, role := range requestingRoles {
		if role == "Owner" || role == "Admin" {
			isOwnerOrAdmin = true
			break
		}
	}

	if !isOwnerOrAdmin {
		return errs.Forbidden("Only Owner or Admin can update user roles").HTTPError(c)
	}

	if err := h.roleRepo.SetUserRoles(c.Request().Context(), targetUser.ID, req.RoleIDs); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	updatedRoles, _ := h.roleRepo.GetUserRoles(c.Request().Context(), targetUser.ID)
	var roleNames []string
	for _, r := range updatedRoles {
		roleNames = append(roleNames, r.Name)
	}

	_ = requestingUserID

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         targetUser.ID,
		"tenant_id":  targetUser.TenantID,
		"email":      targetUser.Email,
		"first_name": targetUser.FirstName,
		"last_name":  targetUser.LastName,
		"roles":      roleNames,
	})
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url"`
	Timezone  *string `json:"timezone"`
	Locale    *string `json:"locale"`
	IsActive  *bool   `json:"is_active"`
}

func (h *UserHandler) Update(c echo.Context) error {
	requestingUserID := c.Get("user_id").(uuid.UUID)
	requestingRoles := c.Get("roles").([]string)

	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	_, err = h.userRepo.GetByID(c.Request().Context(), targetUserID)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	isOwnerOrAdmin := false
	for _, role := range requestingRoles {
		if role == "Owner" || role == "Admin" {
			isOwnerOrAdmin = true
			break
		}
	}

	isSelfEdit := requestingUserID == targetUserID

	if !isSelfEdit && !isOwnerOrAdmin {
		return errs.Forbidden("You don't have permission to edit this user").HTTPError(c)
	}

	domainReq := &user.UpdateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarURL: req.AvatarURL,
		Timezone:  req.Timezone,
		Locale:    req.Locale,
	}

	if isOwnerOrAdmin && !isSelfEdit {
		domainReq.IsActive = req.IsActive
	}

	updatedUser, err := h.userRepo.Update(c.Request().Context(), targetUserID, domainReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	updatedRoles, _ := h.roleRepo.GetUserRoles(c.Request().Context(), updatedUser.ID)
	var roleNames []string
	for _, r := range updatedRoles {
		roleNames = append(roleNames, r.Name)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":         updatedUser.ID,
		"tenant_id":  updatedUser.TenantID,
		"email":      updatedUser.Email,
		"first_name": updatedUser.FirstName,
		"last_name":  updatedUser.LastName,
		"avatar_url": h.getAvatarURL(c, updatedUser.AvatarURL),
		"timezone":   updatedUser.Timezone,
		"locale":     updatedUser.Locale,
		"is_active":  updatedUser.IsActive,
		"roles":      roleNames,
	})
}
