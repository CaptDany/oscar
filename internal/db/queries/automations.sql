-- name: CreateAutomation :one
INSERT INTO automations (tenant_id, name, description, is_active, trigger_type, trigger_config, conditions, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAutomationByID :one
SELECT * FROM automations WHERE id = $1;

-- name: ListAutomations :many
SELECT * FROM automations 
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveAutomationsByTrigger :many
SELECT * FROM automations 
WHERE tenant_id = $1 AND is_active = true AND trigger_type = $2
ORDER BY created_at ASC;

-- name: UpdateAutomation :one
UPDATE automations
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    is_active = COALESCE($4, is_active),
    trigger_type = COALESCE($5, trigger_type),
    trigger_config = COALESCE($6, trigger_config),
    conditions = COALESCE($7, conditions)
WHERE id = $1
RETURNING *;

-- name: DeleteAutomation :one
DELETE FROM automations WHERE id = $1 RETURNING *;

-- name: CreateAutomationAction :one
INSERT INTO automation_actions (automation_id, position, action_type, action_config)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAutomationActions :many
SELECT * FROM automation_actions 
WHERE automation_id = $1
ORDER BY position ASC;

-- name: UpdateAutomationAction :one
UPDATE automation_actions
SET action_type = COALESCE($2, action_type), action_config = COALESCE($3, action_config)
WHERE id = $1
RETURNING *;

-- name: DeleteAutomationAction :one
DELETE FROM automation_actions WHERE id = $1 RETURNING *;

-- name: DeleteAutomationActions :exec
DELETE FROM automation_actions WHERE automation_id = $1;

-- name: CreateAutomationRun :one
INSERT INTO automation_runs (automation_id, tenant_id, trigger_entity_type, trigger_entity_id, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetAutomationRunByID :one
SELECT * FROM automation_runs WHERE id = $1;

-- name: ListAutomationRuns :many
SELECT ar.*, a.name as automation_name
FROM automation_runs ar
JOIN automations a ON ar.automation_id = a.id
WHERE ar.tenant_id = $1
ORDER BY ar.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAutomationRunsByAutomation :many
SELECT * FROM automation_runs 
WHERE automation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateAutomationRun :one
UPDATE automation_runs
SET status = $2, started_at = COALESCE($3, started_at), completed_at = COALESCE($4, completed_at), error = COALESCE($5, error)
WHERE id = $1
RETURNING *;

-- name: CreateAutomationRunAction :one
INSERT INTO automation_run_actions (run_id, action_id, status)
VALUES ($1, $2, 'pending')
RETURNING *;

-- name: UpdateAutomationRunAction :one
UPDATE automation_run_actions
SET status = $2, result = $3, executed_at = NOW(), error = $4
WHERE id = $1
RETURNING *;
