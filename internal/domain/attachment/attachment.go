package attachment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	StoragePath string    `json:"storage_path"`
	UploadedBy  uuid.UUID `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateAttachmentRequest struct {
	EntityType  string    `json:"entity_type" validate:"required"`
	EntityID    uuid.UUID `json:"entity_id" validate:"required"`
	FileName    string    `json:"file_name" validate:"required"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateAttachmentRequest, storagePath string, uploadedBy uuid.UUID) (*Attachment, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*Attachment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Attachment, error)
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}
