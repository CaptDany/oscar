-- name: CreatePerson :one
INSERT INTO persons (tenant_id, type, status, first_name, last_name, email, phone, avatar_url, company_id, owner_id, source, score, tags, custom_fields)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetPersonByID :one
SELECT * FROM persons WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdatePerson :one
UPDATE persons
SET 
    type = COALESCE($2, type),
    status = COALESCE($3, status),
    first_name = COALESCE($4, first_name),
    last_name = COALESCE($5, last_name),
    email = COALESCE($6, email),
    phone = COALESCE($7, phone),
    avatar_url = COALESCE($8, avatar_url),
    company_id = COALESCE($9, company_id),
    owner_id = COALESCE($10, owner_id),
    source = COALESCE($11, source),
    score = COALESCE($12, score),
    tags = COALESCE($13, tags),
    custom_fields = COALESCE($14, custom_fields),
    converted_at = COALESCE($15, converted_at)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePerson :one
UPDATE persons
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ConvertPerson :one
UPDATE persons
SET type = $2, status = $3, converted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListPersons :many
SELECT * FROM persons 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPersonsByType :many
SELECT * FROM persons 
WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPersonsByStatus :many
SELECT * FROM persons 
WHERE tenant_id = $1 AND type = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListPersonsByOwner :many
SELECT * FROM persons 
WHERE tenant_id = $1 AND owner_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: SearchPersons :many
SELECT * FROM persons 
WHERE tenant_id = $1 
  AND deleted_at IS NULL
  AND (
    first_name ILIKE '%' || $2 || '%' OR
    last_name ILIKE '%' || $2 || '%' OR
    email::text ILIKE '%' || $2 || '%'
  )
ORDER BY first_name ASC, last_name ASC
LIMIT $3 OFFSET $4;

-- name: CountPersons :one
SELECT COUNT(*) FROM persons WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountPersonsByType :one
SELECT COUNT(*) FROM persons WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL;

-- name: CountPersonsByStatus :one
SELECT COUNT(*) FROM persons WHERE tenant_id = $1 AND type = $2 AND status = $3 AND deleted_at IS NULL;

-- name: UpdatePersonScore :one
UPDATE persons
SET score = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: AddTagToPerson :one
UPDATE persons
SET tags = array_distinct(array_append(tags, $2))
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: RemoveTagFromPerson :one
UPDATE persons
SET tags = array_remove(tags, $2)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
