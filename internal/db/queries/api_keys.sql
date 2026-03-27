-- name: CreateAPIKey :one
INSERT INTO api_keys (tenant_id, user_id, key_hash, name, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys WHERE id = $1;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND (expires_at IS NULL OR expires_at > NOW());

-- name: ListAPIKeys :many
SELECT id, tenant_id, user_id, name, last_used_at, expires_at, created_at, updated_at
FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :one
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAPIKey :one
DELETE FROM api_keys WHERE id = $1 RETURNING *;

-- name: DeleteAPIKeysByUser :exec
DELETE FROM api_keys WHERE user_id = $1;

-- name: CountAPIKeys :one
SELECT COUNT(*) FROM api_keys WHERE tenant_id = $1;
