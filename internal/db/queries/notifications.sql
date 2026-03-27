-- name: CreateNotification :one
INSERT INTO notifications (tenant_id, user_id, type, title, body, entity_type, entity_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: ListNotifications :many
SELECT * FROM notifications 
WHERE tenant_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListUnreadNotifications :many
SELECT * FROM notifications 
WHERE tenant_id = $1 AND user_id = $2 AND is_read = false
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: MarkNotificationRead :one
UPDATE notifications
SET is_read = true
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: MarkAllNotificationsRead :one
UPDATE notifications
SET is_read = true
WHERE tenant_id = $1 AND user_id = $2 AND is_read = false
RETURNING *;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND is_read = false;

-- name: DeleteNotification :one
DELETE FROM notifications WHERE id = $1 AND user_id = $2 RETURNING *;
