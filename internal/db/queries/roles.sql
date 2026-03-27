-- name: CreateRole :one
INSERT INTO roles (tenant_id, name, description, is_system, permissions)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE tenant_id = $1 AND name = $2;

-- name: GetSystemRoles :many
SELECT * FROM roles WHERE tenant_id = $1 AND is_system = true;

-- name: ListRoles :many
SELECT * FROM roles WHERE tenant_id = $1 ORDER BY is_system DESC, name ASC;

-- name: UpdateRole :one
UPDATE roles
SET name = COALESCE($2, name), description = COALESCE($3, description), permissions = COALESCE($4, permissions)
WHERE id = $1 AND is_system = false
RETURNING *;

-- name: DeleteRole :one
DELETE FROM roles WHERE id = $1 AND is_system = false RETURNING *;

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetUserRoleNames :exec
SELECT r.name FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: SetUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;
INSERT INTO user_roles (user_id, role_id)
SELECT $1, unnest($2::uuid[]);
