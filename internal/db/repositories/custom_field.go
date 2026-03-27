package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/custom_field"
)

type CustomFieldRepository struct {
	pool *pgxpool.Pool
}

func NewCustomFieldRepository(pool *pgxpool.Pool) *CustomFieldRepository {
	return &CustomFieldRepository{pool: pool}
}

func (r *CustomFieldRepository) Create(ctx context.Context, tenantID uuid.UUID, req *custom_field.CreateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error) {
	query := `
		INSERT INTO custom_field_definitions (tenant_id, entity_type, field_key, label, field_type, options, is_required, show_in_list, show_in_card, position)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`

	row := &generated.CustomFieldDefinition{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.EntityType, req.FieldKey, req.Label, req.FieldType,
		req.Options, req.IsRequired, req.ShowInList, req.ShowInCard, req.Position,
	).Scan(
		&row.ID, &row.TenantID, &row.EntityType, &row.FieldKey, &row.Label, &row.FieldType,
		&row.Options, &row.IsRequired, &row.ShowInList, &row.ShowInCard, &row.Position,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("customField.Create: %w", err)
	}

	return mapCustomFieldRowToDomain(row), nil
}

func (r *CustomFieldRepository) GetByID(ctx context.Context, id uuid.UUID) (*custom_field.CustomFieldDefinition, error) {
	query := `SELECT * FROM custom_field_definitions WHERE id = $1`

	row := &generated.CustomFieldDefinition{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.EntityType, &row.FieldKey, &row.Label, &row.FieldType,
		&row.Options, &row.IsRequired, &row.ShowInList, &row.ShowInCard, &row.Position,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customField.GetByID: custom field not found")
		}
		return nil, fmt.Errorf("customField.GetByID: %w", err)
	}

	return mapCustomFieldRowToDomain(row), nil
}

func (r *CustomFieldRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType custom_field.EntityType) ([]*custom_field.CustomFieldDefinition, error) {
	query := `
		SELECT * FROM custom_field_definitions 
		WHERE tenant_id = $1 AND entity_type = $2
		ORDER BY position ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, entityType)
	if err != nil {
		return nil, fmt.Errorf("customField.ListByEntity: %w", err)
	}
	defer rows.Close()

	var fields []*custom_field.CustomFieldDefinition
	for rows.Next() {
		row := &generated.CustomFieldDefinition{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.EntityType, &row.FieldKey, &row.Label, &row.FieldType,
			&row.Options, &row.IsRequired, &row.ShowInList, &row.ShowInCard, &row.Position,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("customField.ListByEntity scan: %w", err)
		}
		fields = append(fields, mapCustomFieldRowToDomain(row))
	}

	return fields, nil
}

func (r *CustomFieldRepository) ListAll(ctx context.Context, tenantID uuid.UUID) ([]*custom_field.CustomFieldDefinition, error) {
	query := `
		SELECT * FROM custom_field_definitions 
		WHERE tenant_id = $1
		ORDER BY entity_type, position ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("customField.ListAll: %w", err)
	}
	defer rows.Close()

	var fields []*custom_field.CustomFieldDefinition
	for rows.Next() {
		row := &generated.CustomFieldDefinition{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.EntityType, &row.FieldKey, &row.Label, &row.FieldType,
			&row.Options, &row.IsRequired, &row.ShowInList, &row.ShowInCard, &row.Position,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("customField.ListAll scan: %w", err)
		}
		fields = append(fields, mapCustomFieldRowToDomain(row))
	}

	return fields, nil
}

func (r *CustomFieldRepository) Update(ctx context.Context, id uuid.UUID, req *custom_field.UpdateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error) {
	query := `
		UPDATE custom_field_definitions SET
			label = COALESCE($2, label),
			field_type = COALESCE($3, field_type),
			options = COALESCE($4, options),
			is_required = COALESCE($5, is_required),
			show_in_list = COALESCE($6, show_in_list),
			show_in_card = COALESCE($7, show_in_card),
			position = COALESCE($8, position)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.CustomFieldDefinition{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Label, req.FieldType, req.Options,
		req.IsRequired, req.ShowInList, req.ShowInCard, req.Position,
	).Scan(
		&row.ID, &row.TenantID, &row.EntityType, &row.FieldKey, &row.Label, &row.FieldType,
		&row.Options, &row.IsRequired, &row.ShowInList, &row.ShowInCard, &row.Position,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("customField.Update: %w", err)
	}

	return mapCustomFieldRowToDomain(row), nil
}

func (r *CustomFieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM custom_field_definitions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("customField.Delete: %w", err)
	}
	return nil
}

func (r *CustomFieldRepository) Reorder(ctx context.Context, tenantID uuid.UUID, fieldIDs []uuid.UUID) error {
	query := `
		UPDATE custom_field_definitions
		SET position = (SELECT position FROM UNNEST($2::uuid[]) WITH ORDINALITY AS t(id, ord) WHERE t.id = custom_field_definitions.id)
		WHERE id = ANY($2::uuid[]) AND tenant_id = $1
	`
	_, err := r.pool.Exec(ctx, query, tenantID, fieldIDs)
	if err != nil {
		return fmt.Errorf("customField.Reorder: %w", err)
	}
	return nil
}

func mapCustomFieldRowToDomain(row *generated.CustomFieldDefinition) *custom_field.CustomFieldDefinition {
	createdAt := pgTimestamptzToTime(row.CreatedAt)
	updatedAt := pgTimestamptzToTime(row.UpdatedAt)
	if createdAt == nil {
		t := time.Time{}
		createdAt = &t
	}
	if updatedAt == nil {
		t := time.Time{}
		updatedAt = &t
	}
	return &custom_field.CustomFieldDefinition{
		ID:           pgUUIDToUUID(row.ID),
		TenantID:     pgUUIDToUUID(row.TenantID),
		EntityType:   custom_field.EntityType(row.EntityType),
		FieldKey:     row.FieldKey,
		Label:        row.Label,
		FieldType:    custom_field.FieldType(row.FieldType),
		Options:      row.Options,
		IsRequired:   row.IsRequired,
		ShowInList:   row.ShowInList,
		ShowInCard:   row.ShowInCard,
		Position:     int(row.Position),
		CreatedAt:    *createdAt,
		UpdatedAt:    *updatedAt,
	}
}
