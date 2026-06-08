package apikey

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	UserID     uuid.UUID  `json:"user_id"`
	KeyPrefix  string     `json:"key_prefix,omitempty"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required,min=1,max=100"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type CreatedAPIKey struct {
	APIKey
	FullKey string `json:"full_key"`
}

type Repository interface {
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *CreateAPIKeyRequest, keyHash, keyPrefix string) (*APIKey, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
