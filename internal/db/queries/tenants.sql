-- name: CreateTenant :one
INSERT INTO tenants (slug, name, status, subscription_tier, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, status = $3, subscription_tier = $4, settings = $5
WHERE id = $1
RETURNING *;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE status != 'deleted'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants WHERE status != 'deleted';
