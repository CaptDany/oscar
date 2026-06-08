package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/domain/attachment"
)

type AttachmentRepository struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepository(pool *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{pool: pool}
}

func (r *AttachmentRepository) Create(ctx context.Context, tenantID uuid.UUID, req *attachment.CreateAttachmentRequest, storagePath string, uploadedBy uuid.UUID) (*attachment.Attachment, error) {
	query := `
		INSERT INTO attachments (tenant_id, entity_type, entity_id, file_name, file_size, mime_type, storage_path, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, entity_type, entity_id, file_name, file_size, mime_type, storage_path, uploaded_by, created_at
	`

	var (
		id, tid, eid, uid [16]byte
		entityType, fn, mt, sp string
		fs                    int64
		created               time.Time
	)
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.EntityType, req.EntityID, req.FileName, req.FileSize, req.MimeType,
		storagePath, uploadedBy,
	).Scan(&id, &tid, &entityType, &eid, &fn, &fs, &mt, &sp, &uid, &created)
	if err != nil {
		return nil, fmt.Errorf("attachment.Create: %w", err)
	}

	return &attachment.Attachment{
		ID:          uuid.UUID(id),
		TenantID:    uuid.UUID(tid),
		EntityType:  entityType,
		EntityID:    uuid.UUID(eid),
		FileName:    fn,
		FileSize:    fs,
		MimeType:    mt,
		StoragePath: sp,
		UploadedBy:  uuid.UUID(uid),
		CreatedAt:   created,
	}, nil
}

func (r *AttachmentRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*attachment.Attachment, error) {
	query := `
		SELECT id, tenant_id, entity_type, entity_id, file_name, file_size, mime_type, storage_path, uploaded_by, created_at
		FROM attachments
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("attachment.ListByEntity: %w", err)
	}
	defer rows.Close()

	var atts []*attachment.Attachment
	for rows.Next() {
		var (
			id, tid, eid, uid [16]byte
			entityType, fn, mt, sp string
			fs                    int64
			created               time.Time
		)
		if err := rows.Scan(&id, &tid, &entityType, &eid, &fn, &fs, &mt, &sp, &uid, &created); err != nil {
			return nil, fmt.Errorf("attachment.ListByEntity scan: %w", err)
		}
		atts = append(atts, &attachment.Attachment{
			ID:          uuid.UUID(id),
			TenantID:    uuid.UUID(tid),
			EntityType:  entityType,
			EntityID:    uuid.UUID(eid),
			FileName:    fn,
			FileSize:    fs,
			MimeType:    mt,
			StoragePath: sp,
			UploadedBy:  uuid.UUID(uid),
			CreatedAt:   created,
		})
	}

	return atts, nil
}

func (r *AttachmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*attachment.Attachment, error) {
	query := `SELECT id, tenant_id, entity_type, entity_id, file_name, file_size, mime_type, storage_path, uploaded_by, created_at FROM attachments WHERE id = $1`

	var (
		rowID, rowTID, rowEID, rowUID [16]byte
		entityType, fn, mt, sp        string
		fs                            int64
		created                       time.Time
	)
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rowID, &rowTID, &entityType, &rowEID, &fn, &fs, &mt, &sp, &rowUID, &created,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attachment.GetByID: not found")
		}
		return nil, fmt.Errorf("attachment.GetByID: %w", err)
	}

	return &attachment.Attachment{
		ID:          uuid.UUID(rowID),
		TenantID:    uuid.UUID(rowTID),
		EntityType:  entityType,
		EntityID:    uuid.UUID(rowEID),
		FileName:    fn,
		FileSize:    fs,
		MimeType:    mt,
		StoragePath: sp,
		UploadedBy:  uuid.UUID(rowUID),
		CreatedAt:   created,
	}, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM attachments WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("attachment.Delete: %w", err)
	}
	return nil
}
