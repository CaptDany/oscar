package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/audit_log"
)

type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) Create(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, entityType string, entityID uuid.UUID, diff json.RawMessage, ipAddress, userAgent *string) (*audit_log.AuditLog, error) {
	query := `
		INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, diff, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	row := &generated.AuditLog{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, userID, action, entityType, entityID, diff, ipAddress, userAgent,
	).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.Action, &row.EntityType, &row.EntityID,
		&row.Diff, &row.IPAddress, &row.UserAgent,
		&row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("auditLog.Create: %w", err)
	}

	return mapAuditLogRowToDomain(row), nil
}

func (r *AuditLogRepository) List(ctx context.Context, tenantID uuid.UUID, filter *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE al.tenant_id = $1`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EntityType != nil {
		baseQuery += fmt.Sprintf(" AND al.entity_type = $%d", argIdx)
		args = append(args, *filter.EntityType)
		argIdx++
	}
	if filter.EntityID != nil {
		baseQuery += fmt.Sprintf(" AND al.entity_id = $%d", argIdx)
		args = append(args, *filter.EntityID)
		argIdx++
	}
	if filter.UserID != nil {
		baseQuery += fmt.Sprintf(" AND al.user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs al ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("auditLog.List count: %w", err)
	}

	listQuery := `
		SELECT al.*, u.email as user_email, u.first_name, u.last_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		` + baseQuery + ` ORDER BY al.created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, 0)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("auditLog.List: %w", err)
	}
	defer rows.Close()

	var logs []*audit_log.AuditLog
	for rows.Next() {
		row := &generated.AuditLog{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.UserID, &row.Action, &row.EntityType, &row.EntityID,
			&row.Diff, &row.IPAddress, &row.UserAgent,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("auditLog.List scan: %w", err)
		}
		log := mapAuditLogRowToDomain(row)

		if userEmail, ok := rows.ColumnTypes()[9]; ok {
			var email *string
			if err := rows.Scan(&email); err == nil {
				log.UserEmail = email
			}
		}

		logs = append(logs, log)
	}

	nextCursor := ""
	if len(logs) > limit {
		logs = logs[:limit]
		nextCursor = logs[len(logs)-1].ID.String()
	}

	return logs, nextCursor, total, nil
}

func (r *AuditLogRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit, offset int) ([]*audit_log.AuditLog, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT al.*, u.email as user_email, u.first_name, u.last_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.tenant_id = $1 AND al.entity_type = $2 AND al.entity_id = $3
		ORDER BY al.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, tenantID, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("auditLog.ListByEntity: %w", err)
	}
	defer rows.Close()

	var logs []*audit_log.AuditLog
	for rows.Next() {
		row := &generated.AuditLog{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.UserID, &row.Action, &row.EntityType, &row.EntityID,
			&row.Diff, &row.IPAddress, &row.UserAgent,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("auditLog.ListByEntity scan: %w", err)
		}
		logs = append(logs, mapAuditLogRowToDomain(row))
	}

	return logs, nil
}

func (r *AuditLogRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*audit_log.AuditLog, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT al.*, u.email as user_email, u.first_name, u.last_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.tenant_id = $1 AND al.user_id = $2
		ORDER BY al.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, tenantID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("auditLog.ListByUser: %w", err)
	}
	defer rows.Close()

	var logs []*audit_log.AuditLog
	for rows.Next() {
		row := &generated.AuditLog{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.UserID, &row.Action, &row.EntityType, &row.EntityID,
			&row.Diff, &row.IPAddress, &row.UserAgent,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("auditLog.ListByUser scan: %w", err)
		}
		logs = append(logs, mapAuditLogRowToDomain(row))
	}

	return logs, nil
}

func (r *AuditLogRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("auditLog.Count: %w", err)
	}

	return count, nil
}

func mapAuditLogRowToDomain(row *generated.AuditLog) *audit_log.AuditLog {
	return &audit_log.AuditLog{
		ID:         row.ID,
		TenantID:   row.TenantID,
		UserID:     row.UserID,
		Action:     row.Action,
		EntityType: row.EntityType,
		EntityID:   row.EntityID,
		Diff:       row.Diff,
		IPAddress:  row.IPAddress,
		UserAgent:  row.UserAgent,
		CreatedAt:  row.CreatedAt,
	}
}
