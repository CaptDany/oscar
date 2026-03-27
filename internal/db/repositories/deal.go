package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/deal"
)

type DealRepository struct {
	pool *pgxpool.Pool
}

func NewDealRepository(pool *pgxpool.Pool) *DealRepository {
	return &DealRepository{pool: pool}
}

func (r *DealRepository) Create(ctx context.Context, tenantID uuid.UUID, req *deal.CreateDealRequest) (*deal.Deal, error) {
	query := `
		INSERT INTO deals (tenant_id, title, value, currency, stage_id, pipeline_id, person_id, company_id, owner_id, expected_close_date, tags, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *
	`

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Title, req.Value, currency, req.StageID, req.PipelineID,
		req.PersonID, req.CompanyID, req.OwnerID, req.ExpectedCloseDate,
		req.Tags, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.Create: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) GetByID(ctx context.Context, id uuid.UUID) (*deal.Deal, error) {
	query := `SELECT * FROM deals WHERE id = $1 AND deleted_at IS NULL`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("deal.GetByID: deal not found")
		}
		return nil, fmt.Errorf("deal.GetByID: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) Update(ctx context.Context, id uuid.UUID, req *deal.UpdateDealRequest) (*deal.Deal, error) {
	query := `
		UPDATE deals SET
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
		RETURNING *
	`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Title, req.Value, req.Currency, req.StageID, req.PipelineID,
		req.PersonID, req.CompanyID, req.OwnerID, req.ExpectedCloseDate,
		req.Probability, req.Tags, req.CustomFields,
		req.ClosedAt, req.WonReason, req.LostReason,
	).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.Update: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) SoftDelete(ctx context.Context, id uuid.UUID) (*deal.Deal, error) {
	query := `UPDATE deals SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING *`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.SoftDelete: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) MoveToStage(ctx context.Context, id uuid.UUID, stageID, pipelineID uuid.UUID, probability int, closedAt *time.Time) (*deal.Deal, error) {
	query := `
		UPDATE deals
		SET stage_id = $2, pipeline_id = $3, probability = $4, closed_at = $5
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query, id, stageID, pipelineID, probability, closedAt).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.MoveToStage: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) CloseAsWon(ctx context.Context, id uuid.UUID, stageID uuid.UUID, reason string) (*deal.Deal, error) {
	query := `
		UPDATE deals
		SET stage_id = $2, closed_at = NOW(), probability = 100, won_reason = $3
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query, id, stageID, reason).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.CloseAsWon: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) CloseAsLost(ctx context.Context, id uuid.UUID, stageID uuid.UUID, reason string) (*deal.Deal, error) {
	query := `
		UPDATE deals
		SET stage_id = $2, closed_at = NOW(), probability = 0, lost_reason = $3
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Deal{}
	err := r.pool.QueryRow(ctx, query, id, stageID, reason).Scan(
		&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
		&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
		&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
		&row.Probability, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("deal.CloseAsLost: %w", err)
	}

	return mapDealRowToDomain(row), nil
}

func (r *DealRepository) List(ctx context.Context, tenantID uuid.UUID, filter *deal.ListDealsFilter) ([]*deal.Deal, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.PipelineID != nil {
		baseQuery += fmt.Sprintf(" AND pipeline_id = $%d", argIdx)
		args = append(args, *filter.PipelineID)
		argIdx++
	}
	if filter.StageID != nil {
		baseQuery += fmt.Sprintf(" AND stage_id = $%d", argIdx)
		args = append(args, *filter.StageID)
		argIdx++
	}
	if filter.OwnerID != nil {
		baseQuery += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *filter.OwnerID)
		argIdx++
	}
	if filter.PersonID != nil {
		baseQuery += fmt.Sprintf(" AND person_id = $%d", argIdx)
		args = append(args, *filter.PersonID)
		argIdx++
	}
	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM deals ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("deal.List count: %w", err)
	}

	listQuery := `SELECT * FROM deals ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, 0)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("deal.List: %w", err)
	}
	defer rows.Close()

	var deals []*deal.Deal
	for rows.Next() {
		row := &generated.Deal{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
			&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
			&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
			&row.Probability, &row.Tags, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("deal.List scan: %w", err)
		}
		deals = append(deals, mapDealRowToDomain(row))
	}

	nextCursor := ""
	if len(deals) > limit {
		deals = deals[:limit]
		nextCursor = deals[len(deals)-1].ID.String()
	}

	return deals, nextCursor, total, nil
}

func (r *DealRepository) ListByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) ([]*deal.DealWithStage, error) {
	query := `
		SELECT d.*, ps.name as stage_name, ps.position as stage_position, ps.stage_type
		FROM deals d
		JOIN pipeline_stages ps ON d.stage_id = ps.id
		WHERE d.tenant_id = $1 AND d.pipeline_id = $2 AND d.deleted_at IS NULL
		ORDER BY ps.position ASC, d.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("deal.ListByStage: %w", err)
	}
	defer rows.Close()

	var deals []*deal.DealWithStage
	for rows.Next() {
		var d deal.DealWithStage
		err := rows.Scan(
			&d.ID, &d.TenantID, &d.Title, &d.Value, &d.Currency, &d.StageID,
			&d.PipelineID, &d.PersonID, &d.CompanyID, &d.OwnerID,
			&d.ExpectedCloseDate, &d.ClosedAt, &d.WonReason, &d.LostReason,
			&d.Probability, &d.Tags, &d.CustomFields,
			&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
			&d.StageName, &d.StagePosition, &d.StageType,
		)
		if err != nil {
			return nil, fmt.Errorf("deal.ListByStage scan: %w", err)
		}
		deals = append(deals, &d)
	}

	return deals, nil
}

func (r *DealRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*deal.Deal, error) {
	if limit <= 0 {
		limit = 20
	}

	sql := `
		SELECT * FROM deals 
		WHERE tenant_id = $1 
		  AND deleted_at IS NULL
		  AND title ILIKE $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, sql, tenantID, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("deal.Search: %w", err)
	}
	defer rows.Close()

	var deals []*deal.Deal
	for rows.Next() {
		row := &generated.Deal{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
			&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
			&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
			&row.Probability, &row.Tags, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("deal.Search scan: %w", err)
		}
		deals = append(deals, mapDealRowToDomain(row))
	}

	return deals, nil
}

func (r *DealRepository) Count(ctx context.Context, tenantID uuid.UUID, filter *deal.ListDealsFilter) (int, error) {
	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}

	if filter.PipelineID != nil {
		baseQuery += " AND pipeline_id = $2"
		args = append(args, *filter.PipelineID)
	}

	countQuery := `SELECT COUNT(*) FROM deals ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("deal.Count: %w", err)
	}

	return total, nil
}

func (r *DealRepository) GetByCloseDate(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*deal.Deal, error) {
	query := `
		SELECT * FROM deals 
		WHERE tenant_id = $1 
		  AND expected_close_date <= $2 
		  AND closed_at IS NULL
		  AND deleted_at IS NULL
		ORDER BY expected_close_date ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, before)
	if err != nil {
		return nil, fmt.Errorf("deal.GetByCloseDate: %w", err)
	}
	defer rows.Close()

	var deals []*deal.Deal
	for rows.Next() {
		row := &generated.Deal{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Title, &row.Value, &row.Currency, &row.StageID,
			&row.PipelineID, &row.PersonID, &row.CompanyID, &row.OwnerID,
			&row.ExpectedCloseDate, &row.ClosedAt, &row.WonReason, &row.LostReason,
			&row.Probability, &row.Tags, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("deal.GetByCloseDate scan: %w", err)
		}
		deals = append(deals, mapDealRowToDomain(row))
	}

	return deals, nil
}

func (r *DealRepository) GetPipelineStats(ctx context.Context, pipelineID uuid.UUID) ([]deal.PipelineStats, error) {
	query := `
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
		ORDER BY ps.position ASC
	`

	rows, err := r.pool.Query(ctx, query, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("deal.GetPipelineStats: %w", err)
	}
	defer rows.Close()

	var stats []deal.PipelineStats
	for rows.Next() {
		var s deal.PipelineStats
		if err := rows.Scan(&s.StageID, &s.StageName, &s.Probability, &s.DealCount, &s.TotalValue); err != nil {
			return nil, fmt.Errorf("deal.GetPipelineStats scan: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func mapDealRowToDomain(row *generated.Deal) *deal.Deal {
	return &deal.Deal{
		ID:                 pgUUIDToUUID(row.ID),
		TenantID:           pgUUIDToUUID(row.TenantID),
		Title:              row.Title,
		Value:              pgNumericToFloat(row.Value),
		Currency:           row.Currency,
		StageID:            pgUUIDToPtr(row.StageID),
		PipelineID:         pgUUIDToPtr(row.PipelineID),
		PersonID:           pgUUIDToPtr(row.PersonID),
		CompanyID:          pgUUIDToPtr(row.CompanyID),
		OwnerID:            pgUUIDToPtr(row.OwnerID),
		ExpectedCloseDate:   pgDateToTime(row.ExpectedCloseDate),
		ClosedAt:           pgTimestamptzToTime(row.ClosedAt),
		WonReason:          pgTextToStr(row.WonReason),
		LostReason:         pgTextToStr(row.LostReason),
		Probability:        pgInt4ToInt(row.Probability),
		Tags:               row.Tags,
		CustomFields:       row.CustomFields,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
		DeletedAt:          pgTimestamptzToTime(row.DeletedAt),
	}
}

type PipelineRepository struct {
	pool *pgxpool.Pool
}

func NewPipelineRepository(pool *pgxpool.Pool) *PipelineRepository {
	return &PipelineRepository{pool: pool}
}

func (r *PipelineRepository) Create(ctx context.Context, tenantID uuid.UUID, name string, isDefault bool, currency string) (*deal.Pipeline, error) {
	query := `INSERT INTO pipelines (tenant_id, name, is_default, currency) VALUES ($1, $2, $3, $4) RETURNING *`

	row := &generated.Pipeline{}
	err := r.pool.QueryRow(ctx, query, tenantID, name, isDefault, currency).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.IsDefault, &row.Currency,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pipeline.Create: %w", err)
	}

	return mapPipelineRowToDomain(row), nil
}

func (r *PipelineRepository) GetByID(ctx context.Context, id uuid.UUID) (*deal.Pipeline, error) {
	query := `SELECT * FROM pipelines WHERE id = $1`

	row := &generated.Pipeline{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.IsDefault, &row.Currency,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pipeline.GetByID: pipeline not found")
		}
		return nil, fmt.Errorf("pipeline.GetByID: %w", err)
	}

	return mapPipelineRowToDomain(row), nil
}

func (r *PipelineRepository) GetDefault(ctx context.Context, tenantID uuid.UUID) (*deal.Pipeline, error) {
	query := `SELECT * FROM pipelines WHERE tenant_id = $1 AND is_default = true`

	row := &generated.Pipeline{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.IsDefault, &row.Currency,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pipeline.GetDefault: pipeline not found")
		}
		return nil, fmt.Errorf("pipeline.GetDefault: %w", err)
	}

	return mapPipelineRowToDomain(row), nil
}

func (r *PipelineRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*deal.Pipeline, error) {
	query := `SELECT * FROM pipelines WHERE tenant_id = $1 ORDER BY is_default DESC, name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("pipeline.List: %w", err)
	}
	defer rows.Close()

	var pipelines []*deal.Pipeline
	for rows.Next() {
		row := &generated.Pipeline{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.IsDefault, &row.Currency,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("pipeline.List scan: %w", err)
		}
		pipelines = append(pipelines, mapPipelineRowToDomain(row))
	}

	return pipelines, nil
}

func (r *PipelineRepository) Update(ctx context.Context, id uuid.UUID, name *string, isDefault *bool, currency *string) (*deal.Pipeline, error) {
	query := `
		UPDATE pipelines
		SET name = COALESCE($2, name), is_default = COALESCE($3, is_default), currency = COALESCE($4, currency)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.Pipeline{}
	err := r.pool.QueryRow(ctx, query, id, name, isDefault, currency).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.IsDefault, &row.Currency,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pipeline.Update: %w", err)
	}

	return mapPipelineRowToDomain(row), nil
}

func (r *PipelineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pipelines WHERE id = $1 AND is_default = false`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("pipeline.Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("pipeline.Delete: cannot delete default pipeline")
	}
	return nil
}

func (r *PipelineRepository) SetDefault(ctx context.Context, tenantID, pipelineID uuid.UUID) error {
	query := `
		UPDATE pipelines SET is_default = false WHERE tenant_id = $1 AND is_default = true;
		UPDATE pipelines SET is_default = true WHERE id = $2;
	`
	_, err := r.pool.Exec(ctx, query, tenantID, pipelineID)
	if err != nil {
		return fmt.Errorf("pipeline.SetDefault: %w", err)
	}
	return nil
}

func (r *PipelineRepository) CreateStage(ctx context.Context, pipelineID uuid.UUID, name string, position, probability int, stageType string) (*deal.PipelineStage, error) {
	query := `INSERT INTO pipeline_stages (pipeline_id, name, position, probability, stage_type) VALUES ($1, $2, $3, $4, $5) RETURNING *`

	row := &generated.PipelineStage{}
	err := r.pool.QueryRow(ctx, query, pipelineID, name, position, probability, stageType).Scan(
		&row.ID, &row.PipelineID, &row.Name, &row.Position, &row.Probability, &row.StageType,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pipeline.CreateStage: %w", err)
	}

	return mapPipelineStageRowToDomain(row), nil
}

func (r *PipelineRepository) GetStageByID(ctx context.Context, id uuid.UUID) (*deal.PipelineStage, error) {
	query := `SELECT * FROM pipeline_stages WHERE id = $1`

	row := &generated.PipelineStage{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.PipelineID, &row.Name, &row.Position, &row.Probability, &row.StageType,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pipeline.GetStageByID: stage not found")
		}
		return nil, fmt.Errorf("pipeline.GetStageByID: %w", err)
	}

	return mapPipelineStageRowToDomain(row), nil
}

func (r *PipelineRepository) ListStages(ctx context.Context, pipelineID uuid.UUID) ([]deal.PipelineStage, error) {
	query := `SELECT * FROM pipeline_stages WHERE pipeline_id = $1 ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline.ListStages: %w", err)
	}
	defer rows.Close()

	var stages []deal.PipelineStage
	for rows.Next() {
		row := &generated.PipelineStage{}
		err := rows.Scan(
			&row.ID, &row.PipelineID, &row.Name, &row.Position, &row.Probability, &row.StageType,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("pipeline.ListStages scan: %w", err)
		}
		stages = append(stages, *mapPipelineStageRowToDomain(row))
	}

	return stages, nil
}

func (r *PipelineRepository) UpdateStage(ctx context.Context, id uuid.UUID, name *string, probability *int, stageType *string) (*deal.PipelineStage, error) {
	query := `
		UPDATE pipeline_stages
		SET name = COALESCE($2, name), probability = COALESCE($3, probability), stage_type = COALESCE($4, stage_type)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.PipelineStage{}
	err := r.pool.QueryRow(ctx, query, id, name, probability, stageType).Scan(
		&row.ID, &row.PipelineID, &row.Name, &row.Position, &row.Probability, &row.StageType,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("pipeline.UpdateStage: %w", err)
	}

	return mapPipelineStageRowToDomain(row), nil
}

func (r *PipelineRepository) ReorderStages(ctx context.Context, pipelineID uuid.UUID, stageIDs []uuid.UUID) error {
	query := `
		UPDATE pipeline_stages
		SET position = (SELECT position FROM UNNEST($2::uuid[]) WITH ORDINALITY AS t(id, ord) WHERE t.id = pipeline_stages.id)
		WHERE id = ANY($2::uuid[])
	`
	_, err := r.pool.Exec(ctx, query, pipelineID, stageIDs)
	if err != nil {
		return fmt.Errorf("pipeline.ReorderStages: %w", err)
	}
	return nil
}

func (r *PipelineRepository) DeleteStage(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pipeline_stages WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("pipeline.DeleteStage: %w", err)
	}
	return nil
}

func mapPipelineRowToDomain(row *generated.Pipeline) *deal.Pipeline {
	return &deal.Pipeline{
		ID:        pgUUIDToUUID(row.ID),
		TenantID:  pgUUIDToUUID(row.TenantID),
		Name:      row.Name,
		IsDefault: row.IsDefault,
		Currency:  row.Currency,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func mapPipelineStageRowToDomain(row *generated.PipelineStage) *deal.PipelineStage {
	return &deal.PipelineStage{
		ID:          pgUUIDToUUID(row.ID),
		PipelineID:  pgUUIDToUUID(row.PipelineID),
		Name:        row.Name,
		Position:    int(row.Position),
		Probability: int(row.Probability),
		StageType:   string(row.StageType),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
