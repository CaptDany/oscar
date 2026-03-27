-- name: CreateDeal :one
INSERT INTO deals (tenant_id, title, value, currency, stage_id, pipeline_id, person_id, company_id, owner_id, expected_close_date, probability, tags, custom_fields)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetDealByID :one
SELECT * FROM deals WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateDeal :one
UPDATE deals
SET 
    title = COALESCE($2, title),
    value = COALESCE($3, value),
    currency = COALESCE($4, currency),
    stage_id = COALESCE($5, stage_id),
    pipeline_id = COALESCE($6, pipeline_id),
    person_id = COALESCE($7, person_id),
    company_id = COALESCE($8, company_id),
    owner_id = COALESCE($9, owner_id),
    expected_close_date = COALESCE($10, expected_close_date),
    probability = COALESCE($11, probability),
    tags = COALESCE($12, tags),
    custom_fields = COALESCE($13, custom_fields),
    closed_at = COALESCE($14, closed_at),
    won_reason = COALESCE($15, won_reason),
    lost_reason = COALESCE($16, lost_reason)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteDeal :one
UPDATE deals
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: MoveDealToStage :one
UPDATE deals
SET stage_id = $2, pipeline_id = $3, probability = $4, closed_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CloseDealAsWon :one
UPDATE deals
SET stage_id = $2, closed_at = NOW(), probability = 100
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CloseDealAsLost :one
UPDATE deals
SET stage_id = $2, closed_at = NOW(), probability = 0
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListDeals :many
SELECT * FROM deals 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDealsByOwner :many
SELECT * FROM deals 
WHERE tenant_id = $1 AND owner_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListDealsByPipeline :many
SELECT * FROM deals 
WHERE tenant_id = $1 AND pipeline_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListDealsByStage :many
SELECT * FROM deals 
WHERE tenant_id = $1 AND stage_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetDealsKanban :many
SELECT d.*, ps.name as stage_name, ps.position as stage_position, ps.stage_type
FROM deals d
JOIN pipeline_stages ps ON d.stage_id = ps.id
WHERE d.tenant_id = $1 AND d.pipeline_id = $2 AND d.deleted_at IS NULL
ORDER BY ps.position ASC, d.created_at DESC;

-- name: SearchDeals :many
SELECT * FROM deals 
WHERE tenant_id = $1 
  AND deleted_at IS NULL
  AND title ILIKE '%' || $2 || '%'
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountDeals :one
SELECT COUNT(*) FROM deals WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: GetDealsByCloseDate :many
SELECT * FROM deals 
WHERE tenant_id = $1 
  AND expected_close_date <= $2 
  AND closed_at IS NULL
  AND deleted_at IS NULL
ORDER BY expected_close_date ASC;

-- name: GetPipelineStats :many
SELECT 
    ps.id as stage_id,
    ps.name as stage_name,
    ps.probability,
    COUNT(d.id) as deal_count,
    COALESCE(SUM(d.value), 0) as total_value
FROM pipeline_stages ps
LEFT JOIN deals d ON d.stage_id = ps.id AND d.deleted_at IS NULL
WHERE ps.pipeline_id = $1
GROUP BY ps.id, ps.name, ps.probability, ps.position
ORDER BY ps.position ASC;
