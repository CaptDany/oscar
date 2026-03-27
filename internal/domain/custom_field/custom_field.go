package custom_field

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDate     FieldType = "date"
	FieldTypeSelect   FieldType = "select"
	FieldTypeMultiSelect FieldType = "multiselect"
)

type EntityType string

const (
	EntityTypePerson   EntityType = "person"
	EntityTypeCompany  EntityType = "company"
	EntityTypeDeal     EntityType = "deal"
	EntityTypeActivity EntityType = "activity"
)

type CustomFieldDefinition struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	EntityType   EntityType     `json:"entity_type"`
	FieldKey     string         `json:"field_key"`
	Label        string         `json:"label"`
	FieldType    FieldType      `json:"field_type"`
	Options      interface{}    `json:"options,omitempty"`
	IsRequired   bool           `json:"is_required"`
	ShowInList   bool           `json:"show_in_list"`
	ShowInCard   bool           `json:"show_in_card"`
	Position     int            `json:"position"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CreateCustomFieldRequest struct {
	EntityType   EntityType   `json:"entity_type" validate:"required"`
	FieldKey     string       `json:"field_key" validate:"required,min=1,max=50"`
	Label        string       `json:"label" validate:"required,min=1,max=100"`
	FieldType    FieldType    `json:"field_type" validate:"required"`
	Options      interface{}  `json:"options"`
	IsRequired   bool         `json:"is_required"`
	ShowInList   bool         `json:"show_in_list"`
	ShowInCard   bool         `json:"show_in_card"`
	Position     int          `json:"position"`
}

type UpdateCustomFieldRequest struct {
	Label        *string      `json:"label"`
	FieldType    *FieldType   `json:"field_type"`
	Options      interface{}  `json:"options"`
	IsRequired   *bool        `json:"is_required"`
	ShowInList   *bool        `json:"show_in_list"`
	ShowInCard   *bool        `json:"show_in_card"`
	Position     *int         `json:"position"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateCustomFieldRequest) (*CustomFieldDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*CustomFieldDefinition, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType EntityType) ([]*CustomFieldDefinition, error)
	ListAll(ctx context.Context, tenantID uuid.UUID) ([]*CustomFieldDefinition, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateCustomFieldRequest) (*CustomFieldDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Reorder(ctx context.Context, tenantID uuid.UUID, fieldIDs []uuid.UUID) error
}
