-- name: CreateAuditLog :one
INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, diff, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAuditLogs :many
SELECT al.*, u.email as user_email, u.first_name, u.last_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.tenant_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByEntity :many
SELECT al.*, u.email as user_email, u.first_name, u.last_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.tenant_id = $1 AND al.entity_type = $2 AND al.entity_id = $3
ORDER BY al.created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListAuditLogsByUser :many
SELECT al.*, u.email as user_email, u.first_name, u.last_name
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.tenant_id = $1 AND al.user_id = $2
ORDER BY al.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1;
