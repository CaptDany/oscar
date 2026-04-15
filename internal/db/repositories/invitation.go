package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/domain/invitation"
)

type InvitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{pool: pool}
}

func (r *InvitationRepository) Create(ctx context.Context, tenantID uuid.UUID, invitedBy uuid.UUID, req *invitation.CreateInvitationRequest, token string, expiresAt time.Time) (*invitation.Invitation, error) {
	query := `
		INSERT INTO invitations (tenant_id, email, token, first_name, last_name, role_name, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	var row struct {
		ID         uuid.UUID
		TenantID   uuid.UUID
		Email      string
		Token      string
		FirstName  string
		LastName   string
		RoleName   string
		InvitedBy  *uuid.UUID
		ExpiresAt  time.Time
		AcceptedAt *time.Time
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Email, token, req.FirstName, req.LastName,
		req.RoleName, invitedBy, expiresAt,
	).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.Token, &row.FirstName, &row.LastName,
		&row.RoleName, &row.InvitedBy, &row.ExpiresAt, &row.AcceptedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitation.Create: %w", err)
	}

	return &invitation.Invitation{
		ID:         row.ID,
		TenantID:   row.TenantID,
		Email:      row.Email,
		Token:      row.Token,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		RoleName:   row.RoleName,
		InvitedBy:  row.InvitedBy,
		ExpiresAt:  row.ExpiresAt,
		AcceptedAt: row.AcceptedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *InvitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error) {
	query := `SELECT * FROM invitations WHERE id = $1`

	var row struct {
		ID         uuid.UUID
		TenantID   uuid.UUID
		Email      string
		Token      string
		FirstName  string
		LastName   string
		RoleName   string
		InvitedBy  *uuid.UUID
		ExpiresAt  time.Time
		AcceptedAt *time.Time
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.Token, &row.FirstName, &row.LastName,
		&row.RoleName, &row.InvitedBy, &row.ExpiresAt, &row.AcceptedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitation.GetByID: %w", err)
	}

	return &invitation.Invitation{
		ID:         row.ID,
		TenantID:   row.TenantID,
		Email:      row.Email,
		Token:      row.Token,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		RoleName:   row.RoleName,
		InvitedBy:  row.InvitedBy,
		ExpiresAt:  row.ExpiresAt,
		AcceptedAt: row.AcceptedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *InvitationRepository) GetByToken(ctx context.Context, token string) (*invitation.Invitation, error) {
	query := `SELECT * FROM invitations WHERE token = $1`

	var row struct {
		ID         uuid.UUID
		TenantID   uuid.UUID
		Email      string
		Token      string
		FirstName  string
		LastName   string
		RoleName   string
		InvitedBy  *uuid.UUID
		ExpiresAt  time.Time
		AcceptedAt *time.Time
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.Token, &row.FirstName, &row.LastName,
		&row.RoleName, &row.InvitedBy, &row.ExpiresAt, &row.AcceptedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitation.GetByToken: %w", err)
	}

	return &invitation.Invitation{
		ID:         row.ID,
		TenantID:   row.TenantID,
		Email:      row.Email,
		Token:      row.Token,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		RoleName:   row.RoleName,
		InvitedBy:  row.InvitedBy,
		ExpiresAt:  row.ExpiresAt,
		AcceptedAt: row.AcceptedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *InvitationRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*invitation.Invitation, error) {
	query := `SELECT * FROM invitations WHERE tenant_id = $1 AND email = $2 AND accepted_at IS NULL`

	var row struct {
		ID         uuid.UUID
		TenantID   uuid.UUID
		Email      string
		Token      string
		FirstName  string
		LastName   string
		RoleName   string
		InvitedBy  *uuid.UUID
		ExpiresAt  time.Time
		AcceptedAt *time.Time
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	err := r.pool.QueryRow(ctx, query, tenantID, email).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.Token, &row.FirstName, &row.LastName,
		&row.RoleName, &row.InvitedBy, &row.ExpiresAt, &row.AcceptedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invitation.GetByEmail: %w", err)
	}

	return &invitation.Invitation{
		ID:         row.ID,
		TenantID:   row.TenantID,
		Email:      row.Email,
		Token:      row.Token,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		RoleName:   row.RoleName,
		InvitedBy:  row.InvitedBy,
		ExpiresAt:  row.ExpiresAt,
		AcceptedAt: row.AcceptedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *InvitationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*invitation.InvitationWithInviter, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	countQuery := `SELECT COUNT(*) FROM invitations WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("invitation.ListByTenant count: %w", err)
	}

	query := `
		SELECT 
			i.id, i.tenant_id, i.email, i.token, i.first_name, i.last_name, 
			i.role_name, i.invited_by, i.expires_at, i.accepted_at, i.created_at, i.updated_at,
			NULLIF(CONCAT(u.first_name, ' ', u.last_name), ' ') as inviter_name,
			u.email as inviter_email
		FROM invitations i
		LEFT JOIN users u ON i.invited_by = u.id
		WHERE i.tenant_id = $1
		ORDER BY i.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invitation.ListByTenant: %w", err)
	}
	defer rows.Close()

	var invitations []*invitation.InvitationWithInviter
	for rows.Next() {
		var inv invitation.InvitationWithInviter
		var inviterName, inviterEmail *string
		var invitedBy *uuid.UUID
		var expiresAt, createdAt, updatedAt time.Time
		var acceptedAt *time.Time

		err := rows.Scan(
			&inv.ID, &inv.TenantID, &inv.Email, &inv.Token, &inv.FirstName, &inv.LastName,
			&inv.RoleName, &invitedBy, &expiresAt, &acceptedAt, &createdAt, &updatedAt,
			&inviterName, &inviterEmail,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("invitation.ListByTenant scan: %w", err)
		}

		inv.InvitedBy = invitedBy
		inv.ExpiresAt = expiresAt
		inv.AcceptedAt = acceptedAt
		inv.CreatedAt = createdAt
		inv.UpdatedAt = updatedAt

		if inviterName != nil && inviterEmail != nil {
			inv.Inviter = &invitation.InvitedByUser{
				Name:  *inviterName,
				Email: *inviterEmail,
			}
		}

		invitations = append(invitations, &inv)
	}

	return invitations, total, nil
}

func (r *InvitationRepository) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE invitations SET accepted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("invitation.MarkAccepted: %w", err)
	}
	return nil
}

func (r *InvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("invitation.Delete: %w", err)
	}
	return nil
}

func (r *InvitationRepository) DeleteByEmail(ctx context.Context, tenantID uuid.UUID, email string) error {
	query := `DELETE FROM invitations WHERE tenant_id = $1 AND email = $2`
	_, err := r.pool.Exec(ctx, query, tenantID, email)
	if err != nil {
		return fmt.Errorf("invitation.DeleteByEmail: %w", err)
	}
	return nil
}
