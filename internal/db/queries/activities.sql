-- name: CreateActivity :one
INSERT INTO activities (tenant_id, type, subject, body, outcome, direction, status, due_at, duration_seconds, owner_id, created_by, custom_fields)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetActivityByID :one
SELECT * FROM activities WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateActivity :one
UPDATE activities
SET 
    type = COALESCE($2, type),
    subject = COALESCE($3, subject),
    body = COALESCE($4, body),
    outcome = COALESCE($5, outcome),
    direction = COALESCE($6, direction),
    status = COALESCE($7, status),
    due_at = COALESCE($8, due_at),
    completed_at = COALESCE($9, completed_at),
    duration_seconds = COALESCE($10, duration_seconds),
    owner_id = COALESCE($11, owner_id),
    custom_fields = COALESCE($12, custom_fields)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CompleteActivity :one
UPDATE activities
SET status = 'completed', completed_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteActivity :one
UPDATE activities
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListActivities :many
SELECT * FROM activities 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActivitiesByOwner :many
SELECT * FROM activities 
WHERE tenant_id = $1 AND owner_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListActivitiesByType :many
SELECT * FROM activities 
WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListActivitiesByStatus :many
SELECT * FROM activities 
WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
ORDER BY due_at ASC NULLS LAST
LIMIT $3 OFFSET $4;

-- name: ListPendingReminders :many
SELECT * FROM activities 
WHERE tenant_id = $1 AND status = 'planned' AND due_at <= NOW() AND deleted_at IS NULL
ORDER BY due_at ASC;

-- name: CountActivities :one
SELECT COUNT(*) FROM activities WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: CountActivitiesByType :many
SELECT type, COUNT(*) as count
FROM activities
WHERE tenant_id = $1 AND deleted_at IS NULL
GROUP BY type;

-- name: CreateActivityAssociation :one
INSERT INTO activity_associations (activity_id, entity_type, entity_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListActivityAssociations :many
SELECT * FROM activity_associations WHERE activity_id = $1;

-- name: ListTimeline :many
SELECT a.*, array_agg(jsonb_build_object('entity_type', aa.entity_type, 'entity_id', aa.entity_id)) as associations
FROM activities a
JOIN activity_associations aa ON a.id = aa.activity_id
WHERE aa.entity_type = $1 AND aa.entity_id = $2 AND a.deleted_at IS NULL
GROUP BY a.id
ORDER BY a.created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteActivityAssociations :exec
DELETE FROM activity_associations WHERE activity_id = $1;
