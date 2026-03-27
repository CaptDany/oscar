-- name: CreateCustomFieldDefinition :one
INSERT INTO custom_field_definitions (tenant_id, entity_type, field_key, label, field_type, options, is_required, show_in_list, show_in_card, position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCustomFieldDefinitionByID :one
SELECT * FROM custom_field_definitions WHERE id = $1;

-- name: ListCustomFieldDefinitions :many
SELECT * FROM custom_field_definitions 
WHERE tenant_id = $1 AND entity_type = $2
ORDER BY position ASC;

-- name: ListAllCustomFieldDefinitions :many
SELECT * FROM custom_field_definitions 
WHERE tenant_id = $1
ORDER BY entity_type, position ASC;

-- name: UpdateCustomFieldDefinition :one
UPDATE custom_field_definitions
SET 
    label = COALESCE($2, label),
    field_type = COALESCE($3, field_type),
    options = COALESCE($4, options),
    is_required = COALESCE($5, is_required),
    show_in_list = COALESCE($6, show_in_list),
    show_in_card = COALESCE($7, show_in_card),
    position = COALESCE($8, position)
WHERE id = $1
RETURNING *;

-- name: DeleteCustomFieldDefinition :one
DELETE FROM custom_field_definitions WHERE id = $1 RETURNING *;

-- name: ReorderCustomFieldDefinitions :exec
-- Note: This is handled via a transaction in the repository
