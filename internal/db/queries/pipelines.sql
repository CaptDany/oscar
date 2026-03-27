-- name: CreatePipeline :one
INSERT INTO pipelines (tenant_id, name, is_default, currency)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPipelineByID :one
SELECT * FROM pipelines WHERE id = $1;

-- name: GetDefaultPipeline :one
SELECT * FROM pipelines WHERE tenant_id = $1 AND is_default = true;

-- name: ListPipelines :many
SELECT * FROM pipelines WHERE tenant_id = $1 ORDER BY is_default DESC, name ASC;

-- name: UpdatePipeline :one
UPDATE pipelines
SET name = COALESCE($2, name), is_default = COALESCE($3, is_default), currency = COALESCE($4, currency)
WHERE id = $1
RETURNING *;

-- name: SetDefaultPipeline :exec
UPDATE pipelines
SET is_default = false
WHERE tenant_id = $1 AND is_default = true;
UPDATE pipelines
SET is_default = true
WHERE id = $2;

-- name: DeletePipeline :one
DELETE FROM pipelines WHERE id = $1 AND is_default = false RETURNING *;

-- name: CreatePipelineStage :one
INSERT INTO pipeline_stages (pipeline_id, name, position, probability, stage_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPipelineStageByID :one
SELECT * FROM pipeline_stages WHERE id = $1;

-- name: ListPipelineStages :many
SELECT * FROM pipeline_stages WHERE pipeline_id = $1 ORDER BY position ASC;

-- name: UpdatePipelineStage :one
UPDATE pipeline_stages
SET name = COALESCE($2, name), probability = COALESCE($3, probability), stage_type = COALESCE($4, stage_type)
WHERE id = $1
RETURNING *;

-- name: ReorderPipelineStages :exec
-- Note: This is handled via a transaction in the repository

-- name: DeletePipelineStage :one
DELETE FROM pipeline_stages WHERE id = $1 RETURNING *;
