package handlers

import (
	"bytes"
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

type UploadHandler struct {
	storage  *storage.R2Client
	userRepo user.Repository
}

func NewUploadHandler(storage *storage.R2Client, userRepo user.Repository) *UploadHandler {
	return &UploadHandler{
		storage:  storage,
		userRepo: userRepo,
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
