package handlers

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/attachment"
	"github.com/oscar/oscar/internal/storage"
	"github.com/oscar/oscar/pkg/errs"
)

type AttachmentHandler struct {
	repo    attachment.Repository
	storage *storage.R2Client
}

func NewAttachmentHandler(repo attachment.Repository, storage *storage.R2Client) *AttachmentHandler {
	return &AttachmentHandler{repo: repo, storage: storage}
}

type GetAttachmentPresignedURLRequest struct {
	EntityType string    `json:"entity_type" validate:"required"`
	EntityID   uuid.UUID `json:"entity_id" validate:"required"`
	FileName   string    `json:"file_name" validate:"required"`
	FileSize   int64     `json:"file_size"`
	MimeType   string    `json:"mime_type"`
}

type GetAttachmentPresignedURLResponse struct {
	PresignedURL string `json:"presigned_url"`
	AttachmentID string `json:"attachment_id"`
	PublicURL    string `json:"public_url"`
}

func (h *AttachmentHandler) GetPresignedURL(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var req GetAttachmentPresignedURLRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if req.MimeType == "" {
		req.MimeType = "application/octet-stream"
	}

	ext := path.Ext(req.FileName)
	attachmentID := uuid.New()
	storagePath := fmt.Sprintf("attachments/%s/%s/%s%s", tenantID.String(), req.EntityType, attachmentID.String(), ext)

	presignedURL, err := h.storage.GetPresignedPutURL(c.Request().Context(), storagePath, 15*time.Minute)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	domainReq := &attachment.CreateAttachmentRequest{
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		MimeType:   req.MimeType,
	}

	result, err := h.repo.Create(c.Request().Context(), tenantID, domainReq, storagePath, userID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	publicURL := h.storage.GetPublicURL(storagePath)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": GetAttachmentPresignedURLResponse{
			PresignedURL: presignedURL,
			AttachmentID: result.ID.String(),
			PublicURL:    publicURL,
		},
	})
}

func (h *AttachmentHandler) ListByEntity(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	entityType := c.Param("entity_type")
	entityID, err := uuid.Parse(c.Param("entity_id"))
	if err != nil {
		return errs.BadRequest("Invalid entity ID").HTTPError(c)
	}

	validTypes := map[string]bool{"person": true, "company": true, "deal": true, "activity": true}
	if !validTypes[entityType] {
		return errs.BadRequest("Invalid entity type. Must be one of: person, company, deal, activity").HTTPError(c)
	}

	atts, err := h.repo.ListByEntity(c.Request().Context(), tenantID, entityType, entityID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	type attachmentResponse struct {
		ID         string    `json:"id"`
		FileName   string    `json:"file_name"`
		FileSize   int64     `json:"file_size"`
		MimeType   string    `json:"mime_type"`
		PublicURL  string    `json:"public_url"`
		UploadedBy string    `json:"uploaded_by"`
		CreatedAt  time.Time `json:"created_at"`
	}

	resp := make([]attachmentResponse, 0, len(atts))
	for _, a := range atts {
		resp = append(resp, attachmentResponse{
			ID:         a.ID.String(),
			FileName:   a.FileName,
			FileSize:   a.FileSize,
			MimeType:   a.MimeType,
			PublicURL:  h.storage.GetPublicURL(a.StoragePath),
			UploadedBy: a.UploadedBy.String(),
			CreatedAt:  a.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": resp,
	})
}

func (h *AttachmentHandler) Delete(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	attachmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid attachment ID").HTTPError(c)
	}

	att, err := h.repo.GetByID(c.Request().Context(), attachmentID)
	if err != nil {
		return errs.NotFound("Attachment not found").HTTPError(c)
	}

	if err := h.storage.Delete(c.Request().Context(), att.StoragePath); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), attachmentID, tenantID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "Attachment deleted"},
	})
}


