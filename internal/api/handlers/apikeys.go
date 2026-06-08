package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/apikey"
	"github.com/oscar/oscar/internal/db/repositories"
	"github.com/oscar/oscar/pkg/errs"
)

type APIKeyHandler struct {
	repo apikey.Repository
}

func NewAPIKeyHandler(repo apikey.Repository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

type CreateAPIKeyRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	ExpiresAt string `json:"expires_at"`
}

func (h *APIKeyHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	keys, err := h.repo.List(c.Request().Context(), tenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if keys == nil {
		keys = []*apikey.APIKey{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": keys,
	})
}

func (h *APIKeyHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	fullKey, keyHash, keyPrefix, err := repositories.GenerateAPIKey()
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return errs.BadRequest("Invalid expires_at format, use RFC3339").HTTPError(c)
		}
		expiresAt = &t
	}

	domainReq := &apikey.CreateAPIKeyRequest{
		Name:      req.Name,
		ExpiresAt: expiresAt,
	}

	key, err := h.repo.Create(c.Request().Context(), tenantID, userID, domainReq, keyHash, keyPrefix)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": map[string]interface{}{
			"id":         key.ID,
			"name":       key.Name,
			"key_prefix": keyPrefix,
			"full_key":   fullKey,
			"created_at": key.CreatedAt,
			"expires_at": key.ExpiresAt,
		},
	})
}

func (h *APIKeyHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid API key ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "API key revoked"},
	})
}
