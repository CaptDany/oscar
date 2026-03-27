package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/notification"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, tenantID uuid.UUID, req *notification.CreateNotificationRequest) (*notification.Notification, error) {
	query := `
		INSERT INTO notifications (tenant_id, user_id, type, title, body, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`

	row := &generated.Notification{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.UserID, req.Type, req.Title, req.Body, req.EntityType, req.EntityID,
	).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.Type, &row.Title, &row.Body,
		&row.EntityType, &row.EntityID, &row.IsRead,
		&row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("notification.Create: %w", err)
	}

	return mapNotificationRowToDomain(row), nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	query := `SELECT * FROM notifications WHERE id = $1`

	row := &generated.Notification{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.Type, &row.Title, &row.Body,
		&row.EntityType, &row.EntityID, &row.IsRead,
		&row.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("notification.GetByID: notification not found")
		}
		return nil, fmt.Errorf("notification.GetByID: %w", err)
	}

	return mapNotificationRowToDomain(row), nil
}

func (r *NotificationRepository) List(ctx context.Context, tenantID, userID uuid.UUID, filter *notification.ListNotificationsFilter) ([]*notification.Notification, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE tenant_id = $1 AND user_id = $2`
	args := []interface{}{tenantID, userID}
	argIdx := 3

	if filter.UnreadOnly {
		baseQuery += ` AND is_read = false`
	}

	countQuery := `SELECT COUNT(*) FROM notifications ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("notification.List count: %w", err)
	}

	listQuery := `SELECT * FROM notifications ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, 0)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("notification.List: %w", err)
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		row := &generated.Notification{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.UserID, &row.Type, &row.Title, &row.Body,
			&row.EntityType, &row.EntityID, &row.IsRead,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("notification.List scan: %w", err)
		}
		notifications = append(notifications, mapNotificationRowToDomain(row))
	}

	nextCursor := ""
	if len(notifications) > limit {
		notifications = notifications[:limit]
		nextCursor = notifications[len(notifications)-1].ID.String()
	}

	return notifications, nextCursor, total, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) (*notification.Notification, error) {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND user_id = $2
		RETURNING *
	`

	row := &generated.Notification{}
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.Type, &row.Title, &row.Body,
		&row.EntityType, &row.EntityID, &row.IsRead,
		&row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("notification.MarkAsRead: %w", err)
	}

	return mapNotificationRowToDomain(row), nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE tenant_id = $1 AND user_id = $2 AND is_read = false
	`

	result, err := r.pool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return 0, fmt.Errorf("notification.MarkAllAsRead: %w", err)
	}

	return int(result.RowsAffected()), nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND user_id = $2 AND is_read = false`

	var count int
	err := r.pool.QueryRow(ctx, query, tenantID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("notification.CountUnread: %w", err)
	}

	return count, nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("notification.Delete: %w", err)
	}
	return nil
}

func mapNotificationRowToDomain(row *generated.Notification) *notification.Notification {
	return &notification.Notification{
		ID:         row.ID,
		TenantID:   row.TenantID,
		UserID:     row.UserID,
		Type:       row.Type,
		Title:      row.Title,
		Body:       row.Body,
		EntityType: row.EntityType,
		EntityID:   row.EntityID,
		IsRead:     row.IsRead,
		CreatedAt:  row.CreatedAt,
	}
}
