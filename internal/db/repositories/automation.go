package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/automation"
)

type AutomationRepository struct {
	pool *pgxpool.Pool
}

func NewAutomationRepository(pool *pgxpool.Pool) *AutomationRepository {
	return &AutomationRepository{pool: pool}
}

func (r *AutomationRepository) Create(ctx context.Context, tenantID uuid.UUID, req *automation.CreateAutomationRequest, createdBy *uuid.UUID) (*automation.Automation, error) {
	query := `
		INSERT INTO automations (tenant_id, name, description, is_active, trigger_type, trigger_config, conditions, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	row := &generated.Automation{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Name, req.Description, req.IsActive,
		req.TriggerType, req.TriggerConfig, req.Conditions, createdBy,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsActive,
		&row.TriggerType, &row.TriggerConfig, &row.Conditions, &row.CreatedBy,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("automation.Create: %w", err)
	}

	return mapAutomationRowToDomain(row), nil
}

func (r *AutomationRepository) GetByID(ctx context.Context, id uuid.UUID) (*automation.Automation, error) {
	query := `SELECT * FROM automations WHERE id = $1`

	row := &generated.Automation{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsActive,
		&row.TriggerType, &row.TriggerConfig, &row.Conditions, &row.CreatedBy,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("automation.GetByID: automation not found")
		}
		return nil, fmt.Errorf("automation.GetByID: %w", err)
	}

	return mapAutomationRowToDomain(row), nil
}

func (r *AutomationRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*automation.Automation, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM automations WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("automation.List count: %w", err)
	}

	query := `
		SELECT * FROM automations 
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("automation.List: %w", err)
	}
	defer rows.Close()

	var automations []*automation.Automation
	for rows.Next() {
		row := &generated.Automation{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsActive,
			&row.TriggerType, &row.TriggerConfig, &row.Conditions, &row.CreatedBy,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("automation.List scan: %w", err)
		}
		automations = append(automations, mapAutomationRowToDomain(row))
	}

	return automations, total, nil
}

func (r *AutomationRepository) Update(ctx context.Context, id uuid.UUID, req *automation.UpdateAutomationRequest) (*automation.Automation, error) {
	query := `
		UPDATE automations SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			is_active = COALESCE($4, is_active),
			trigger_type = COALESCE($5, trigger_type),
			trigger_config = COALESCE($6, trigger_config),
			conditions = COALESCE($7, conditions)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.Automation{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Name, req.Description, req.IsActive,
		req.TriggerType, req.TriggerConfig, req.Conditions,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsActive,
		&row.TriggerType, &row.TriggerConfig, &row.Conditions, &row.CreatedBy,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("automation.Update: %w", err)
	}

	return mapAutomationRowToDomain(row), nil
}

func (r *AutomationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM automations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("automation.Delete: %w", err)
	}
	return nil
}

func (r *AutomationRepository) GetActiveByTrigger(ctx context.Context, tenantID uuid.UUID, trigger automation.TriggerType) ([]*automation.Automation, error) {
	query := `
		SELECT * FROM automations 
		WHERE tenant_id = $1 AND is_active = true AND trigger_type = $2
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, trigger)
	if err != nil {
		return nil, fmt.Errorf("automation.GetActiveByTrigger: %w", err)
	}
	defer rows.Close()

	var automations []*automation.Automation
	for rows.Next() {
		row := &generated.Automation{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsActive,
			&row.TriggerType, &row.TriggerConfig, &row.Conditions, &row.CreatedBy,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("automation.GetActiveByTrigger scan: %w", err)
		}
		automations = append(automations, mapAutomationRowToDomain(row))
	}

	return automations, nil
}

func mapAutomationRowToDomain(row *generated.Automation) *automation.Automation {
	return &automation.Automation{
		ID:            row.ID,
		TenantID:      row.TenantID,
		Name:          row.Name,
		Description:   row.Description,
		IsActive:      row.IsActive,
		TriggerType:   row.TriggerType,
		TriggerConfig: row.TriggerConfig,
		Conditions:    row.Conditions,
		CreatedBy:     row.CreatedBy,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

type AutomationActionRepository struct {
	pool *pgxpool.Pool
}

func NewAutomationActionRepository(pool *pgxpool.Pool) *AutomationActionRepository {
	return &AutomationActionRepository{pool: pool}
}

func (r *AutomationActionRepository) Create(ctx context.Context, automationID uuid.UUID, actions []automation.ActionConfig) ([]automation.AutomationAction, error) {
	var result []automation.AutomationAction

	for _, action := range actions {
		query := `INSERT INTO automation_actions (automation_id, position, action_type, action_config) VALUES ($1, $2, $3, $4) RETURNING *`

		row := &generated.AutomationAction{}
		err := r.pool.QueryRow(ctx, query, automationID, action.Position, action.ActionType, action.ActionConfig).Scan(
			&row.ID, &row.AutomationID, &row.Position, &row.ActionType, &row.ActionConfig,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("automationAction.Create: %w", err)
		}
		result = append(result, *mapAutomationActionRowToDomain(row))
	}

	return result, nil
}

func (r *AutomationActionRepository) ListByAutomation(ctx context.Context, automationID uuid.UUID) ([]automation.AutomationAction, error) {
	query := `SELECT * FROM automation_actions WHERE automation_id = $1 ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, automationID)
	if err != nil {
		return nil, fmt.Errorf("automationAction.ListByAutomation: %w", err)
	}
	defer rows.Close()

	var actions []automation.AutomationAction
	for rows.Next() {
		row := &generated.AutomationAction{}
		err := rows.Scan(
			&row.ID, &row.AutomationID, &row.Position, &row.ActionType, &row.ActionConfig,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("automationAction.ListByAutomation scan: %w", err)
		}
		actions = append(actions, *mapAutomationActionRowToDomain(row))
	}

	return actions, nil
}

func (r *AutomationActionRepository) Update(ctx context.Context, id uuid.UUID, actionType *automation.ActionType, config *[]byte) (*automation.AutomationAction, error) {
	query := `
		UPDATE automation_actions
		SET action_type = COALESCE($2, action_type), action_config = COALESCE($3, action_config)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.AutomationAction{}
	err := r.pool.QueryRow(ctx, query, id, actionType, config).Scan(
		&row.ID, &row.AutomationID, &row.Position, &row.ActionType, &row.ActionConfig,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("automationAction.Update: %w", err)
	}

	return mapAutomationActionRowToDomain(row), nil
}

func (r *AutomationActionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM automation_actions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("automationAction.Delete: %w", err)
	}
	return nil
}

func (r *AutomationActionRepository) DeleteByAutomation(ctx context.Context, automationID uuid.UUID) error {
	query := `DELETE FROM automation_actions WHERE automation_id = $1`
	_, err := r.pool.Exec(ctx, query, automationID)
	if err != nil {
		return fmt.Errorf("automationAction.DeleteByAutomation: %w", err)
	}
	return nil
}

func mapAutomationActionRowToDomain(row *generated.AutomationAction) *automation.AutomationAction {
	return &automation.AutomationAction{
		ID:           row.ID,
		AutomationID: row.AutomationID,
		Position:     int(row.Position),
		ActionType:   row.ActionType,
		ActionConfig: row.ActionConfig,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

type AutomationRunRepository struct {
	pool *pgxpool.Pool
}

func NewAutomationRunRepository(pool *pgxpool.Pool) *AutomationRunRepository {
	return &AutomationRunRepository{pool: pool}
}

func (r *AutomationRunRepository) Create(ctx context.Context, automationID, tenantID uuid.UUID, entityType *string, entityID *uuid.UUID) (*automation.AutomationRun, error) {
	query := `
		INSERT INTO automation_runs (automation_id, tenant_id, trigger_entity_type, trigger_entity_id, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING *
	`

	row := &generated.AutomationRun{}
	err := r.pool.QueryRow(ctx, query, automationID, tenantID, entityType, entityID).Scan(
		&row.ID, &row.AutomationID, &row.TenantID, &row.TriggerEntityType, &row.TriggerEntityID,
		&row.Status, &row.StartedAt, &row.CompletedAt, &row.Error,
		&row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("automationRun.Create: %w", err)
	}

	return mapAutomationRunRowToDomain(row), nil
}

func (r *AutomationRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*automation.AutomationRun, error) {
	query := `SELECT * FROM automation_runs WHERE id = $1`

	row := &generated.AutomationRun{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.AutomationID, &row.TenantID, &row.TriggerEntityType, &row.TriggerEntityID,
		&row.Status, &row.StartedAt, &row.CompletedAt, &row.Error,
		&row.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("automationRun.GetByID: run not found")
		}
		return nil, fmt.Errorf("automationRun.GetByID: %w", err)
	}

	return mapAutomationRunRowToDomain(row), nil
}

func (r *AutomationRunRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*automation.AutomationRun, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM automation_runs WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("automationRun.List count: %w", err)
	}

	query := `
		SELECT ar.*, a.name as automation_name
		FROM automation_runs ar
		JOIN automations a ON ar.automation_id = a.id
		WHERE ar.tenant_id = $1
		ORDER BY ar.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("automationRun.List: %w", err)
	}
	defer rows.Close()

	var runs []*automation.AutomationRun
	for rows.Next() {
		row := &generated.AutomationRun{}
		err := rows.Scan(
			&row.ID, &row.AutomationID, &row.TenantID, &row.TriggerEntityType, &row.TriggerEntityID,
			&row.Status, &row.StartedAt, &row.CompletedAt, &row.Error,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("automationRun.List scan: %w", err)
		}
		runs = append(runs, mapAutomationRunRowToDomain(row))
	}

	return runs, total, nil
}

func (r *AutomationRunRepository) ListByAutomation(ctx context.Context, automationID uuid.UUID, limit, offset int) ([]*automation.AutomationRun, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM automation_runs WHERE automation_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, automationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("automationRun.ListByAutomation count: %w", err)
	}

	query := `
		SELECT * FROM automation_runs 
		WHERE automation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, automationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("automationRun.ListByAutomation: %w", err)
	}
	defer rows.Close()

	var runs []*automation.AutomationRun
	for rows.Next() {
		row := &generated.AutomationRun{}
		err := rows.Scan(
			&row.ID, &row.AutomationID, &row.TenantID, &row.TriggerEntityType, &row.TriggerEntityID,
			&row.Status, &row.StartedAt, &row.CompletedAt, &row.Error,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("automationRun.ListByAutomation scan: %w", err)
		}
		runs = append(runs, mapAutomationRunRowToDomain(row))
	}

	return runs, total, nil
}

func (r *AutomationRunRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status automation.RunStatus, startedAt, completedAt *string, err *string) (*automation.AutomationRun, error) {
	query := `
		UPDATE automation_runs
		SET status = $2, started_at = COALESCE($3, started_at), completed_at = COALESCE($4, completed_at), error = COALESCE($5, error)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.AutomationRun{}
	errScan := r.pool.QueryRow(ctx, query, id, status, startedAt, completedAt, err).Scan(
		&row.ID, &row.AutomationID, &row.TenantID, &row.TriggerEntityType, &row.TriggerEntityID,
		&row.Status, &row.StartedAt, &row.CompletedAt, &row.Error,
		&row.CreatedAt,
	)
	if errScan != nil {
		return nil, fmt.Errorf("automationRun.UpdateStatus: %w", errScan)
	}

	return mapAutomationRunRowToDomain(row), nil
}

func (r *AutomationRunRepository) CreateRunAction(ctx context.Context, runID, actionID uuid.UUID) (*automation.AutomationRunAction, error) {
	query := `
		INSERT INTO automation_run_actions (run_id, action_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING *
	`

	row := &generated.AutomationRunAction{}
	err := r.pool.QueryRow(ctx, query, runID, actionID).Scan(
		&row.ID, &row.RunID, &row.ActionID, &row.Status, &row.Result, &row.ExecutedAt, &row.Error,
		&row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("automationRun.CreateRunAction: %w", err)
	}

	return mapAutomationRunActionRowToDomain(row), nil
}

func (r *AutomationRunRepository) UpdateRunAction(ctx context.Context, id uuid.UUID, status automation.RunStatus, result []byte, err *string) error {
	query := `
		UPDATE automation_run_actions
		SET status = $2, result = $3, executed_at = NOW(), error = $4
		WHERE id = $1
	`

	_, errExec := r.pool.Exec(ctx, query, id, status, result, err)
	if errExec != nil {
		return fmt.Errorf("automationRun.UpdateRunAction: %w", errExec)
	}
	return nil
}

func mapAutomationRunRowToDomain(row *generated.AutomationRun) *automation.AutomationRun {
	return &automation.AutomationRun{
		ID:                row.ID,
		AutomationID:      row.AutomationID,
		TenantID:          row.TenantID,
		TriggerEntityType: row.TriggerEntityType,
		TriggerEntityID:   row.TriggerEntityID,
		Status:             row.Status,
		StartedAt:          row.StartedAt,
		CompletedAt:        row.CompletedAt,
		Error:              row.Error,
		CreatedAt:          row.CreatedAt,
	}
}

func mapAutomationRunActionRowToDomain(row *generated.AutomationRunAction) *automation.AutomationRunAction {
	return &automation.AutomationRunAction{
		ID:         row.ID,
		RunID:      row.RunID,
		ActionID:   row.ActionID,
		Status:     row.Status,
		Result:     row.Result,
		ExecutedAt: row.ExecutedAt,
		Error:      row.Error,
		CreatedAt:  row.CreatedAt,
	}
}
