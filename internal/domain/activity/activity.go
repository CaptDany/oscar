package activity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ActivityType string

const (
	ActivityTypeNote     ActivityType = "note"
	ActivityTypeCall     ActivityType = "call"
	ActivityTypeEmail    ActivityType = "email"
	ActivityTypeMeeting  ActivityType = "meeting"
	ActivityTypeTask     ActivityType = "task"
	ActivityTypeWhatsapp ActivityType = "whatsapp"
	ActivityTypeSMS      ActivityType = "sms"
)

type ActivityStatus string

const (
	ActivityStatusPlanned   ActivityStatus = "planned"
	ActivityStatusCompleted ActivityStatus = "completed"
	ActivityStatusCancelled ActivityStatus = "cancelled"
)

type ActivityDirection string

const (
	ActivityDirectionInbound  ActivityDirection = "inbound"
	ActivityDirectionOutbound ActivityDirection = "outbound"
)

type EntityType string

const (
	EntityTypePerson  EntityType = "person"
	EntityTypeCompany EntityType = "company"
	EntityTypeDeal    EntityType = "deal"
)

type Activity struct {
	ID              uuid.UUID          `json:"id"`
	TenantID        uuid.UUID          `json:"tenant_id"`
	Type            ActivityType       `json:"type"`
	Subject         string             `json:"subject"`
	Body            *string            `json:"body"`
	Outcome         *string            `json:"outcome"`
	Direction       *ActivityDirection `json:"direction"`
	Status          ActivityStatus     `json:"status"`
	DueAt           *time.Time         `json:"due_at"`
	CompletedAt     *time.Time         `json:"completed_at"`
	DurationSeconds *int               `json:"duration_seconds"`
	OwnerID         *uuid.UUID         `json:"owner_id"`
	CreatedBy       *uuid.UUID         `json:"created_by"`
	CustomFields    interface{}        `json:"custom_fields"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       *time.Time         `json:"-"`
}

type ActivityAssociation struct {
	ID         uuid.UUID  `json:"id"`
	ActivityID uuid.UUID  `json:"activity_id"`
	EntityType EntityType `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TimelineEntry struct {
	Activity
	Associations []ActivityAssociation `json:"associations"`
}

type CreateActivityRequest struct {
	Type            ActivityType       `json:"type" validate:"required"`
	Subject         string             `json:"subject" validate:"required,min=1,max=255"`
	Body            *string            `json:"body"`
	Outcome         *string            `json:"outcome"`
	Direction       *ActivityDirection `json:"direction"`
	Status          ActivityStatus     `json:"status"`
	DueAt           *time.Time         `json:"due_at"`
	DurationSeconds *int               `json:"duration_seconds"`
	OwnerID         *uuid.UUID         `json:"owner_id"`
	CreatedBy       *uuid.UUID         `json:"created_by"`
	CustomFields    interface{}        `json:"custom_fields"`
	EntityType      *EntityType        `json:"entity_type"`
	EntityID        *uuid.UUID         `json:"entity_id"`
}

type UpdateActivityRequest struct {
	Type            *ActivityType      `json:"type"`
	Subject         *string            `json:"subject"`
	Body            *string            `json:"body"`
	Outcome         *string            `json:"outcome"`
	Direction       *ActivityDirection `json:"direction"`
	Status          *ActivityStatus    `json:"status"`
	DueAt           *time.Time         `json:"due_at"`
	CompletedAt     *time.Time         `json:"completed_at"`
	DurationSeconds *int               `json:"duration_seconds"`
	OwnerID         *uuid.UUID         `json:"owner_id"`
	CustomFields    interface{}        `json:"custom_fields"`
}

type ListActivitiesFilter struct {
	Type       ActivityType
	Status     ActivityStatus
	OwnerID    *uuid.UUID
	EntityType EntityType
	EntityID   *uuid.UUID
	DueBefore  *time.Time
	Cursor     string
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateActivityRequest) (*Activity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Activity, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateActivityRequest) (*Activity, error)
	Complete(ctx context.Context, id uuid.UUID) (*Activity, error)
	Uncomplete(ctx context.Context, id uuid.UUID) (*Activity, error)
	SoftDelete(ctx context.Context, id uuid.UUID) (*Activity, error)
	List(ctx context.Context, tenantID uuid.UUID, filter *ListActivitiesFilter) ([]*Activity, string, int, error)
	GetPendingReminders(ctx context.Context, tenantID uuid.UUID) ([]*Activity, error)
	Count(ctx context.Context, tenantID uuid.UUID, filter *ListActivitiesFilter) (int, error)
	CountByType(ctx context.Context, tenantID uuid.UUID) (map[ActivityType]int, error)
}

type AssociationRepository interface {
	Create(ctx context.Context, activityID uuid.UUID, entityType EntityType, entityID uuid.UUID) (*ActivityAssociation, error)
	ListByActivity(ctx context.Context, activityID uuid.UUID) ([]ActivityAssociation, error)
	ListTimeline(ctx context.Context, entityType EntityType, entityID uuid.UUID, limit, offset int) ([]*TimelineEntry, error)
	DeleteByActivity(ctx context.Context, activityID uuid.UUID) error
}
