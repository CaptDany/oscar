-- name: CreateUser :one
INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, avatar_url, timezone, locale)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET 
    email = COALESCE($2, email),
    first_name = COALESCE($3, first_name),
    last_name = COALESCE($4, last_name),
    avatar_url = COALESCE($5, avatar_url),
    timezone = COALESCE($6, timezone),
    locale = COALESCE($7, locale),
    is_active = COALESCE($8, is_active),
    last_login_at = COALESCE($9, last_login_at)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUsersByIDs :many
SELECT * FROM users 
WHERE tenant_id = $1 AND id = ANY($2) AND deleted_at IS NULL;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: UpdateLastLogin :one
UPDATE users
SET last_login_at = NOW()
WHERE id = $1
RETURNING *;
