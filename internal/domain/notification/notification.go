package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	UserID     uuid.UUID  `json:"user_id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	EntityType *string    `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateNotificationRequest struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	Type       string    `json:"type" validate:"required"`
	Title      string    `json:"title" validate:"required"`
	Body       string    `json:"body"`
	EntityType *string   `json:"entity_type"`
	EntityID   *uuid.UUID `json:"entity_id"`
}

type ListNotificationsFilter struct {
	UnreadOnly bool
	Cursor     string
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateNotificationRequest) (*Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	List(ctx context.Context, tenantID, userID uuid.UUID, filter *ListNotificationsFilter) ([]*Notification, string, int, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) (*Notification, error)
	MarkAllAsRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
