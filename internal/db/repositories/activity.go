package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/activity"
)

type ActivityRepository struct {
	pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{pool: pool}
}

func (r *ActivityRepository) Create(ctx context.Context, tenantID uuid.UUID, req *activity.CreateActivityRequest) (*activity.Activity, error) {
	query := `
		INSERT INTO activities (tenant_id, type, subject, body, outcome, direction, status, due_at, duration_seconds, owner_id, created_by, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *
	`

	status := req.Status
	if status == "" {
		status = activity.ActivityStatusPlanned
	}

	row := &generated.Activity{}
	createdBy := req.CreatedBy
	if createdBy == nil {
		createdBy = req.OwnerID
	}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Type, req.Subject, req.Body, req.Outcome, req.Direction,
		status, req.DueAt, req.DurationSeconds, req.OwnerID, createdBy, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activity.Create: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*activity.Activity, error) {
	query := `SELECT * FROM activities WHERE id = $1 AND deleted_at IS NULL`

	row := &generated.Activity{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("activity.GetByID: activity not found")
		}
		return nil, fmt.Errorf("activity.GetByID: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) Update(ctx context.Context, id uuid.UUID, req *activity.UpdateActivityRequest) (*activity.Activity, error) {
	query := `
		UPDATE activities SET
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
		RETURNING *
	`

	row := &generated.Activity{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Type, req.Subject, req.Body, req.Outcome, req.Direction,
		req.Status, req.DueAt, req.CompletedAt, req.DurationSeconds, req.OwnerID, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activity.Update: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) Uncomplete(ctx context.Context, id uuid.UUID) (*activity.Activity, error) {
	query := `
		UPDATE activities
		SET status = 'planned', completed_at = NULL
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Activity{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activity.Uncomplete: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) Complete(ctx context.Context, id uuid.UUID) (*activity.Activity, error) {
	query := `
		UPDATE activities
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Activity{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activity.Complete: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) SoftDelete(ctx context.Context, id uuid.UUID) (*activity.Activity, error) {
	query := `UPDATE activities SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING *`

	row := &generated.Activity{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
		&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
		&row.OwnerID, &row.CreatedBy, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activity.SoftDelete: %w", err)
	}

	return mapActivityRowToDomain(row), nil
}

func (r *ActivityRepository) List(ctx context.Context, tenantID uuid.UUID, filter *activity.ListActivitiesFilter) ([]*activity.Activity, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.OwnerID != nil {
		baseQuery += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *filter.OwnerID)
		argIdx++
	}
	if filter.DueBefore != nil {
		baseQuery += fmt.Sprintf(" AND due_at <= $%d", argIdx)
		args = append(args, *filter.DueBefore)
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM activities ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("activity.List count: %w", err)
	}

	listQuery := `SELECT * FROM activities ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, 0)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("activity.List: %w", err)
	}
	defer rows.Close()

	var activities []*activity.Activity
	for rows.Next() {
		row := &generated.Activity{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
			&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
			&row.OwnerID, &row.CreatedBy, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("activity.List scan: %w", err)
		}
		activities = append(activities, mapActivityRowToDomain(row))
	}

	nextCursor := ""
	if len(activities) > limit {
		activities = activities[:limit]
		nextCursor = activities[len(activities)-1].ID.String()
	}

	return activities, nextCursor, total, nil
}

func (r *ActivityRepository) GetPendingReminders(ctx context.Context, tenantID uuid.UUID) ([]*activity.Activity, error) {
	query := `
		SELECT * FROM activities 
		WHERE tenant_id = $1 AND status = 'planned' AND due_at <= NOW() AND deleted_at IS NULL
		ORDER BY due_at ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("activity.GetPendingReminders: %w", err)
	}
	defer rows.Close()

	var activities []*activity.Activity
	for rows.Next() {
		row := &generated.Activity{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Type, &row.Subject, &row.Body, &row.Outcome,
			&row.Direction, &row.Status, &row.DueAt, &row.CompletedAt, &row.DurationSeconds,
			&row.OwnerID, &row.CreatedBy, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("activity.GetPendingReminders scan: %w", err)
		}
		activities = append(activities, mapActivityRowToDomain(row))
	}

	return activities, nil
}

func (r *ActivityRepository) Count(ctx context.Context, tenantID uuid.UUID, filter *activity.ListActivitiesFilter) (int, error) {
	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}

	if filter.Type != "" {
		baseQuery += " AND type = $2"
		args = append(args, filter.Type)
	}

	countQuery := `SELECT COUNT(*) FROM activities ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("activity.Count: %w", err)
	}

	return total, nil
}

func (r *ActivityRepository) CountByType(ctx context.Context, tenantID uuid.UUID) (map[activity.ActivityType]int, error) {
	query := `
		SELECT type, COUNT(*) as count
		FROM activities
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY type
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("activity.CountByType: %w", err)
	}
	defer rows.Close()

	result := make(map[activity.ActivityType]int)
	for rows.Next() {
		var activityType activity.ActivityType
		var count int
		if err := rows.Scan(&activityType, &count); err != nil {
			return nil, fmt.Errorf("activity.CountByType scan: %w", err)
		}
		result[activityType] = count
	}

	return result, nil
}

func (r *ActivityRepository) GetActivityReport(ctx context.Context, tenantID uuid.UUID, filter *activity.ActivityReportFilter) (*activity.ActivityReport, error) {
	report := &activity.ActivityReport{
		PeriodStart: filter.StartAt,
		PeriodEnd:   filter.EndAt,
	}

	baseWhere := `WHERE a.tenant_id = $1 AND a.deleted_at IS NULL AND a.created_at >= $2 AND a.created_at < $3`
	args := []interface{}{tenantID, filter.StartAt, filter.EndAt}
	argIdx := 4

	if len(filter.Types) > 0 {
		baseWhere += fmt.Sprintf(" AND a.type = ANY($%d)", argIdx)
		args = append(args, filter.Types)
		argIdx++
	}

	if len(filter.UserIDs) > 0 {
		baseWhere += fmt.Sprintf(" AND (a.owner_id = ANY($%d) OR a.created_by = ANY($%d))", argIdx, argIdx)
		args = append(args, filter.UserIDs)
		argIdx++
	}

	baseWhere += fmt.Sprintf(" AND a.status = $%d", argIdx)
	args = append(args, activity.ActivityStatusCompleted)
	argIdx++

	totalQuery := `SELECT COUNT(*) FROM activities a ` + baseWhere
	if err := r.pool.QueryRow(ctx, totalQuery, args...).Scan(&report.Total); err != nil {
		return nil, fmt.Errorf("activity.GetActivityReport total: %w", err)
	}

	byTypeQuery := `SELECT a.type, COUNT(*) as count FROM activities a ` + baseWhere + ` GROUP BY a.type ORDER BY count DESC`
	rows, err := r.pool.Query(ctx, byTypeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("activity.GetActivityReport by_type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item activity.ActivityCountByType
		if err := rows.Scan(&item.Type, &item.Count); err != nil {
			return nil, fmt.Errorf("activity.GetActivityReport by_type scan: %w", err)
		}
		report.ByType = append(report.ByType, item)
	}

	byUserQuery := `SELECT COALESCE(u.id, '00000000-0000-0000-0000-000000000000'), COALESCE(u.first_name, 'Deleted'), COALESCE(u.last_name, 'User'), COUNT(*) as count
		FROM activities a
		LEFT JOIN users u ON u.id = COALESCE(a.owner_id, a.created_by)
		` + baseWhere + ` GROUP BY u.id, u.first_name, u.last_name ORDER BY count DESC`
	rows2, err := r.pool.Query(ctx, byUserQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("activity.GetActivityReport by_user: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var item activity.ActivityCountByUser
		if err := rows2.Scan(&item.UserID, &item.FirstName, &item.LastName, &item.Count); err != nil {
			return nil, fmt.Errorf("activity.GetActivityReport by_user scan: %w", err)
		}
		report.ByUser = append(report.ByUser, item)
	}

	byDayQuery := `SELECT DATE(a.created_at) as date, COUNT(*) as count FROM activities a ` + baseWhere + ` GROUP BY DATE(a.created_at) ORDER BY date ASC`
	rows3, err := r.pool.Query(ctx, byDayQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("activity.GetActivityReport by_day: %w", err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var date pgtype.Date
		var item activity.ActivityCountByDay
		if err := rows3.Scan(&date, &item.Count); err != nil {
			return nil, fmt.Errorf("activity.GetActivityReport by_day scan: %w", err)
		}
		if date.Valid {
			item.Date = date.Time.Format("2006-01-02")
		}
		report.ByDay = append(report.ByDay, item)
	}

	return report, nil
}

func mapActivityRowToDomain(row *generated.Activity) *activity.Activity {
	dir := activity.ActivityDirection(row.Direction.ActivityDirection)
	return &activity.Activity{
		ID:              pgUUIDToUUID(row.ID),
		TenantID:        pgUUIDToUUID(row.TenantID),
		Type:            activity.ActivityType(row.Type),
		Subject:         row.Subject,
		Body:            pgTextToStr(row.Body),
		Outcome:         pgTextToStr(row.Outcome),
		Direction:       &dir,
		Status:          activity.ActivityStatus(row.Status),
		DueAt:           pgTimestamptzToTime(row.DueAt),
		CompletedAt:     pgTimestamptzToTime(row.CompletedAt),
		DurationSeconds: pgInt4ToPtr(row.DurationSeconds),
		OwnerID:         pgUUIDToPtr(row.OwnerID),
		CreatedBy:       pgUUIDToPtr(row.CreatedBy),
		CustomFields:    row.CustomFields,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       pgTimestamptzToTime(row.DeletedAt),
	}
}

type ActivityAssociationRepository struct {
	pool *pgxpool.Pool
}

func NewActivityAssociationRepository(pool *pgxpool.Pool) *ActivityAssociationRepository {
	return &ActivityAssociationRepository{pool: pool}
}

func (r *ActivityAssociationRepository) Create(ctx context.Context, activityID uuid.UUID, entityType activity.EntityType, entityID uuid.UUID) (*activity.ActivityAssociation, error) {
	query := `
		INSERT INTO activity_associations (activity_id, entity_type, entity_id)
		VALUES ($1, $2, $3)
		RETURNING *
	`

	row := &generated.ActivityAssociation{}
	err := r.pool.QueryRow(ctx, query, activityID, entityType, entityID).Scan(
		&row.ID, &row.ActivityID, &row.EntityType, &row.EntityID, &row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("activityAssociation.Create: %w", err)
	}

	return mapActivityAssociationRowToDomain(row), nil
}

func (r *ActivityAssociationRepository) ListByActivity(ctx context.Context, activityID uuid.UUID) ([]activity.ActivityAssociation, error) {
	query := `SELECT * FROM activity_associations WHERE activity_id = $1`

	rows, err := r.pool.Query(ctx, query, activityID)
	if err != nil {
		return nil, fmt.Errorf("activityAssociation.ListByActivity: %w", err)
	}
	defer rows.Close()

	var associations []activity.ActivityAssociation
	for rows.Next() {
		row := &generated.ActivityAssociation{}
		err := rows.Scan(
			&row.ID, &row.ActivityID, &row.EntityType, &row.EntityID, &row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("activityAssociation.ListByActivity scan: %w", err)
		}
		associations = append(associations, *mapActivityAssociationRowToDomain(row))
	}

	return associations, nil
}

func (r *ActivityAssociationRepository) ListTimeline(ctx context.Context, entityType activity.EntityType, entityID uuid.UUID, limit, offset int) ([]*activity.TimelineEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT a.id, a.tenant_id, a.type, a.subject, a.body, a.outcome, a.direction, a.status, 
		       a.due_at, a.completed_at, a.duration_seconds, a.owner_id, a.created_by, a.custom_fields,
		       a.created_at, a.updated_at, a.deleted_at,
		       COALESCE(
				(SELECT jsonb_agg(jsonb_build_object('entity_type', aa.entity_type, 'entity_id', aa.entity_id))
				FROM activity_associations aa WHERE aa.activity_id = a.id),
				'[]'::jsonb
			) as associations
		FROM activities a
		JOIN activity_associations aa ON a.id = aa.activity_id
		WHERE aa.entity_type = $1 AND aa.entity_id = $2 AND a.deleted_at IS NULL
		GROUP BY a.id
		ORDER BY a.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("activityAssociation.ListTimeline: %w", err)
	}
	defer rows.Close()

	var entries []*activity.TimelineEntry
	for rows.Next() {
		var entry activity.TimelineEntry
		var direction generated.NullActivityDirection
		var body, outcome pgtype.Text
		var dueAt, completedAt, createdAt, updatedAt, deletedAt pgtype.Timestamptz
		var durationSeconds pgtype.Int4
		var ownerID, createdBy pgtype.UUID
		var customFields []byte

		err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.Type, &entry.Subject, &body, &outcome,
			&direction, &entry.Status, &dueAt, &completedAt, &durationSeconds,
			&ownerID, &createdBy, &customFields,
			&createdAt, &updatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("activityAssociation.ListTimeline scan: %w", err)
		}

		entry.Body = pgTextToStr(body)
		entry.Outcome = pgTextToStr(outcome)
		entry.DueAt = pgTimestamptzToTime(dueAt)
		entry.CompletedAt = pgTimestamptzToTime(completedAt)
		entry.DurationSeconds = pgInt4ToPtr(durationSeconds)
		entry.OwnerID = pgUUIDToPtr(ownerID)
		entry.CreatedBy = pgUUIDToPtr(createdBy)
		entry.CustomFields = customFields
		entry.CreatedAt = createdAt.Time
		entry.UpdatedAt = updatedAt.Time
		entry.DeletedAt = pgTimestamptzToTime(deletedAt)

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *ActivityAssociationRepository) DeleteByActivity(ctx context.Context, activityID uuid.UUID) error {
	query := `DELETE FROM activity_associations WHERE activity_id = $1`
	_, err := r.pool.Exec(ctx, query, activityID)
	if err != nil {
		return fmt.Errorf("activityAssociation.DeleteByActivity: %w", err)
	}
	return nil
}

func mapActivityAssociationRowToDomain(row *generated.ActivityAssociation) *activity.ActivityAssociation {
	return &activity.ActivityAssociation{
		ID:         pgUUIDToUUID(row.ID),
		ActivityID: pgUUIDToUUID(row.ActivityID),
		EntityType: activity.EntityType(row.EntityType),
		EntityID:   pgUUIDToUUID(row.EntityID),
		CreatedAt:  row.CreatedAt.Time,
	}
}
