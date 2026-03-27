package audit_log

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID       `json:"id"`
	TenantID   uuid.UUID       `json:"tenant_id"`
	UserID     *uuid.UUID      `json:"user_id,omitempty"`
	UserEmail  *string         `json:"user_email,omitempty"`
	FirstName  *string         `json:"first_name,omitempty"`
	LastName   *string         `json:"last_name,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id"`
	Diff       json.RawMessage `json:"diff,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	UserAgent  *string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ListAuditLogsFilter struct {
	EntityType *string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	Cursor     string
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, entityType string, entityID uuid.UUID, diff json.RawMessage, ipAddress, userAgent *string) (*AuditLog, error)
	List(ctx context.Context, tenantID uuid.UUID, filter *ListAuditLogsFilter) ([]*AuditLog, string, int, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	Count(ctx context.Context, tenantID uuid.UUID) (int, error)
}
