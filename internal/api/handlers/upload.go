package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/user"
	"github.com/oscar/oscar/internal/storage"
	"github.com/oscar/oscar/pkg/errs"
)

type BrandAssetsUpdater interface {
	UpdateBrandAssets(ctx context.Context, tenantID uuid.UUID, logoLightURL, logoDarkURL, faviconURL *string) error
}

type UploadHandler struct {
	storage  *storage.R2Client
	userRepo user.Repository
	brandRepo BrandAssetsUpdater
}

func NewUploadHandler(storage *storage.R2Client, userRepo user.Repository, brandRepo BrandAssetsUpdater) *UploadHandler {
	return &UploadHandler{
		storage:  storage,
		userRepo: userRepo,
		brandRepo: brandRepo,
	}
}

type GetPresignedURLRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type GetPresignedURLResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	FinalURL  string `json:"final_url"`
}

func (h *UploadHandler) GetAvatarPresignedURL(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return errs.Unauthorized("User not authenticated").HTTPError(c)
	}

	var req GetPresignedURLRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".heif" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return errs.BadRequest("Invalid file type. Allowed: jpg, jpeg, heif, png, gif, webp").HTTPError(c)
	}

	timestamp := time.Now().UnixMilli()
	objectKey := fmt.Sprintf("avatars/%s/%d%s", userID.String(), timestamp, ".jpg")

	uploadURL, err := h.storage.GetPresignedPutURL(c.Request().Context(), objectKey, 5*time.Minute)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	finalURL, err := h.storage.GetPresignedURL(c.Request().Context(), objectKey, 24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"upload_url": uploadURL,
		"object_key": objectKey,
		"final_url":  finalURL,
	})
}

type ConfirmAvatarRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
}

func (h *UploadHandler) GetAvatarURL(c echo.Context) error {
	fmt.Printf("GetAvatarURL: called for path %s\n", c.Request().URL.Path)
	userIDParam := c.Param("user_id")

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	user, err := h.userRepo.GetByID(c.Request().Context(), userID)
	if err != nil {
		return errs.NotFound("User not found").HTTPError(c)
	}

	if user.AvatarURL == nil {
		return errs.NotFound("Avatar not set").HTTPError(c)
	}

	avatarKey := *user.AvatarURL
	if avatarKey == "" {
		return errs.NotFound("Avatar not set").HTTPError(c)
	}

	presignedURL, err := h.storage.GetPresignedURL(c.Request().Context(), avatarKey, 24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.Redirect(http.StatusTemporaryRedirect, presignedURL)
}

func (h *UploadHandler) ConfirmAvatarUpload(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	if userID == uuid.Nil {
		return errs.Unauthorized("User not authenticated").HTTPError(c)
	}

	var req ConfirmAvatarRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	reader, err := h.storage.Download(c.Request().Context(), req.ObjectKey)
	if err != nil {
		return errs.BadRequest("Failed to download uploaded file").HTTPError(c)
	}
	defer reader.Close()

	processedImage, err := storage.CropAndResizeToSquare(reader, storage.AvatarSize)
	if err != nil {
		fmt.Printf("Error processing image: %v\n", err)
		return errs.Internal(err).HTTPError(c)
	}

	buf := new(bytes.Buffer)
	processedImageSize, err := io.Copy(buf, processedImage)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	parts := strings.Split(req.ObjectKey, "/")
	if len(parts) < 3 {
		return errs.BadRequest("Invalid object key").HTTPError(c)
	}
	filename := parts[len(parts)-1]
	finalKey := fmt.Sprintf("avatars/%s/processed_%s", userID.String(), filename)

	err = h.storage.Upload(c.Request().Context(), finalKey, bytes.NewReader(buf.Bytes()), processedImageSize, "image/jpeg")
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	err = h.userRepo.UpdateAvatar(c.Request().Context(), userID, finalKey)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"avatar_key": finalKey,
	})
}

type GetBrandingAssetPresignedRequest struct {
	AssetType string `json:"asset_type" validate:"required,oneof=logo_light logo_dark favicon"`
	Filename string `json:"filename" validate:"required"`
}

func (h *UploadHandler) GetBrandingAssetPresignedURL(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	if tenantID == uuid.Nil {
		return errs.Unauthorized("Tenant not authenticated").HTTPError(c)
	}

	var req GetBrandingAssetPresignedRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext != ".svg" {
		return errs.BadRequest("Branding assets must be SVG files").HTTPError(c)
	}

	timestamp := time.Now().UnixMilli()
	var objectKey string
	switch req.AssetType {
	case "logo_light":
		objectKey = fmt.Sprintf("branding/%s/logo_light_%d.svg", tenantID.String(), timestamp)
	case "logo_dark":
		objectKey = fmt.Sprintf("branding/%s/logo_dark_%d.svg", tenantID.String(), timestamp)
	case "favicon":
		objectKey = fmt.Sprintf("branding/%s/favicon_%d.svg", tenantID.String(), timestamp)
	}

	uploadURL, err := h.storage.GetPresignedPutURL(c.Request().Context(), objectKey, 5*time.Minute)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	finalURL, err := h.storage.GetPresignedURL(c.Request().Context(), objectKey, 24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"upload_url": uploadURL,
		"object_key": objectKey,
		"final_url":  finalURL,
	})
}

type ConfirmBrandingAssetRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
	AssetType string `json:"asset_type" validate:"required,oneof=logo_light logo_dark favicon"`
}

func (h *UploadHandler) ConfirmBrandingAssetUpload(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	if tenantID == uuid.Nil {
		return errs.Unauthorized("Tenant not authenticated").HTTPError(c)
	}

	var req ConfirmBrandingAssetRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	if h.brandRepo == nil {
		return errs.Internal(fmt.Errorf("branding repository not configured")).HTTPError(c)
	}

	_, err := h.storage.GetPresignedURL(c.Request().Context(), req.ObjectKey, 24*time.Hour)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	var logoLightURL, logoDarkURL, faviconURL *string
	switch req.AssetType {
	case "logo_light":
		logoLightURL = &req.ObjectKey
	case "logo_dark":
		logoDarkURL = &req.ObjectKey
	case "favicon":
		faviconURL = &req.ObjectKey
	}

	err = h.brandRepo.UpdateBrandAssets(c.Request().Context(), tenantID, logoLightURL, logoDarkURL, faviconURL)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"asset_type": req.AssetType,
		"object_key": req.ObjectKey,
	})
}
