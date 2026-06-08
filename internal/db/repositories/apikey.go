package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/apikey"
)

type APIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

func GenerateAPIKey() (fullKey, hash, prefix string, err error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", "", fmt.Errorf("apikey.Generate: %w", err)
	}

	rawKey := hex.EncodeToString(keyBytes)
	prefix = rawKey[:8]
	fullKey = "osc_" + rawKey

	hashBytes := sha256.Sum256([]byte(fullKey))
	hash = hex.EncodeToString(hashBytes[:])

	return fullKey, hash, prefix, nil
}

func (r *APIKeyRepository) Create(ctx context.Context, tenantID, userID uuid.UUID, req *apikey.CreateAPIKeyRequest, keyHash, keyPrefix string) (*apikey.APIKey, error) {
	query := `
		INSERT INTO api_keys (tenant_id, user_id, key_hash, name, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, user_id, key_hash, name, last_used_at, expires_at, created_at, updated_at
	`

	row := &generated.ApiKey{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, userID, keyHash, req.Name, req.ExpiresAt,
	).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.KeyHash, &row.Name,
		&row.LastUsedAt, &row.ExpiresAt, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("apikey.Create: %w", err)
	}

	return &apikey.APIKey{
		ID:         pgUUIDToUUID(row.ID),
		TenantID:   pgUUIDToUUID(row.TenantID),
		UserID:     pgUUIDToUUID(row.UserID),
		KeyPrefix:  keyPrefix,
		Name:       row.Name,
		LastUsedAt: pgTimestamptzToTime(row.LastUsedAt),
		ExpiresAt:  pgTimestamptzToTime(row.ExpiresAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}

func (r *APIKeyRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*apikey.APIKey, error) {
	query := `
		SELECT id, tenant_id, user_id, name, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("apikey.List: %w", err)
	}
	defer rows.Close()

	var keys []*apikey.APIKey
	for rows.Next() {
		row := &generated.ListAPIKeysRow{}
		if err := rows.Scan(
			&row.ID, &row.TenantID, &row.UserID, &row.Name,
			&row.LastUsedAt, &row.ExpiresAt, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("apikey.List scan: %w", err)
		}
		keys = append(keys, &apikey.APIKey{
			ID:         pgUUIDToUUID(row.ID),
			TenantID:   pgUUIDToUUID(row.TenantID),
			UserID:     pgUUIDToUUID(row.UserID),
			Name:       row.Name,
			LastUsedAt: pgTimestamptzToTime(row.LastUsedAt),
			ExpiresAt:  pgTimestamptzToTime(row.ExpiresAt),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		})
	}

	return keys, nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*apikey.APIKey, error) {
	query := `SELECT id, tenant_id, user_id, key_hash, name, last_used_at, expires_at, created_at, updated_at FROM api_keys WHERE id = $1`

	row := &generated.ApiKey{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.KeyHash, &row.Name,
		&row.LastUsedAt, &row.ExpiresAt, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("apikey.GetByID: not found")
		}
		return nil, fmt.Errorf("apikey.GetByID: %w", err)
	}

	return &apikey.APIKey{
		ID:         pgUUIDToUUID(row.ID),
		TenantID:   pgUUIDToUUID(row.TenantID),
		UserID:     pgUUIDToUUID(row.UserID),
		Name:       row.Name,
		LastUsedAt: pgTimestamptzToTime(row.LastUsedAt),
		ExpiresAt:  pgTimestamptzToTime(row.ExpiresAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}

func (r *APIKeyRepository) GetByHash(ctx context.Context, hash string) (*apikey.APIKey, error) {
	query := `SELECT id, tenant_id, user_id, key_hash, name, last_used_at, expires_at, created_at, updated_at FROM api_keys WHERE key_hash = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	row := &generated.ApiKey{}
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&row.ID, &row.TenantID, &row.UserID, &row.KeyHash, &row.Name,
		&row.LastUsedAt, &row.ExpiresAt, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("apikey.GetByHash: not found or expired")
		}
		return nil, fmt.Errorf("apikey.GetByHash: %w", err)
	}

	return &apikey.APIKey{
		ID:         pgUUIDToUUID(row.ID),
		TenantID:   pgUUIDToUUID(row.TenantID),
		UserID:     pgUUIDToUUID(row.UserID),
		Name:       row.Name,
		LastUsedAt: pgTimestamptzToTime(row.LastUsedAt),
		ExpiresAt:  pgTimestamptzToTime(row.ExpiresAt),
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("apikey.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("apikey.Delete: not found")
	}
	return nil
}
