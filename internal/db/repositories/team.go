package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/team"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) Create(ctx context.Context, tenantID uuid.UUID, req *team.CreateTeamRequest) (*team.Team, error) {
	query := `INSERT INTO teams (tenant_id, name, description) VALUES ($1, $2, $3) RETURNING *`

	row := &generated.Team{}
	err := r.pool.QueryRow(ctx, query, tenantID, req.Name, req.Description).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("team.Create: %w", err)
	}

	return mapTeamRowToDomain(row), nil
}

func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	query := `SELECT * FROM teams WHERE id = $1`

	row := &generated.Team{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("team.GetByID: team not found")
		}
		return nil, fmt.Errorf("team.GetByID: %w", err)
	}

	return mapTeamRowToDomain(row), nil
}

func (r *TeamRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*team.Team, error) {
	query := `SELECT * FROM teams WHERE tenant_id = $1 ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("team.List: %w", err)
	}
	defer rows.Close()

	var teams []*team.Team
	for rows.Next() {
		row := &generated.Team{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("team.List scan: %w", err)
		}
		teams = append(teams, mapTeamRowToDomain(row))
	}

	return teams, nil
}

func (r *TeamRepository) Update(ctx context.Context, id uuid.UUID, req *team.UpdateTeamRequest) (*team.Team, error) {
	query := `
		UPDATE teams
		SET name = COALESCE($2, name), description = COALESCE($3, description)
		WHERE id = $1
		RETURNING *
	`

	row := &generated.Team{}
	err := r.pool.QueryRow(ctx, query, id, req.Name, req.Description).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("team.Update: %w", err)
	}

	return mapTeamRowToDomain(row), nil
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("team.Delete: %w", err)
	}
	return nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID uuid.UUID, isLead bool) (*team.TeamMember, error) {
	query := `
		INSERT INTO team_members (team_id, user_id, is_lead)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET is_lead = $3
		RETURNING *
	`

	row := &generated.TeamMember{}
	err := r.pool.QueryRow(ctx, query, teamID, userID, isLead).Scan(
		&row.TeamID, &row.UserID, &row.IsLead, &row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("team.AddMember: %w", err)
	}

	return mapTeamMemberRowToDomain(row), nil
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("team.RemoveMember: %w", err)
	}
	return nil
}

func (r *TeamRepository) ListMembers(ctx context.Context, teamID uuid.UUID) ([]team.TeamMember, error) {
	query := `
		SELECT tm.*, u.email, u.first_name, u.last_name, u.avatar_url
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
	`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("team.ListMembers: %w", err)
	}
	defer rows.Close()

	var members []team.TeamMember
	for rows.Next() {
		row := &generated.TeamMember{}
		err := rows.Scan(
			&row.TeamID, &row.UserID, &row.IsLead, &row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("team.ListMembers scan: %w", err)
		}
		members = append(members, *mapTeamMemberRowToDomain(row))
	}

	return members, nil
}

func (r *TeamRepository) ListUserTeams(ctx context.Context, userID uuid.UUID) ([]*team.Team, error) {
	query := `
		SELECT t.*, tm.is_lead
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("team.ListUserTeams: %w", err)
	}
	defer rows.Close()

	var teams []*team.Team
	for rows.Next() {
		row := &generated.Team{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("team.ListUserTeams scan: %w", err)
		}
		teams = append(teams, mapTeamRowToDomain(row))
	}

	return teams, nil
}

func (r *TeamRepository) SetLead(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `
		UPDATE team_members SET is_lead = false WHERE team_id = $1;
		UPDATE team_members SET is_lead = true WHERE team_id = $1 AND user_id = $2;
	`
	_, err := r.pool.Exec(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("team.SetLead: %w", err)
	}
	return nil
}

func mapTeamRowToDomain(row *generated.Team) *team.Team {
	desc := pgTextToStr(row.Description)
	createdAt := pgTimestamptzToTime(row.CreatedAt)
	updatedAt := pgTimestamptzToTime(row.UpdatedAt)
	if createdAt == nil {
		t := time.Time{}
		createdAt = &t
	}
	if updatedAt == nil {
		t := time.Time{}
		updatedAt = &t
	}
	return &team.Team{
		ID:          pgUUIDToUUID(row.ID),
		TenantID:    pgUUIDToUUID(row.TenantID),
		Name:        row.Name,
		Description: desc,
		CreatedAt:   *createdAt,
		UpdatedAt:   *updatedAt,
	}
}

func mapTeamMemberRowToDomain(row *generated.TeamMember) *team.TeamMember {
	createdAt := pgTimestamptzToTime(row.CreatedAt)
	if createdAt == nil {
		t := time.Time{}
		createdAt = &t
	}
	return &team.TeamMember{
		ID:        pgUUIDToUUID(row.TeamID),
		TeamID:   pgUUIDToUUID(row.TeamID),
		UserID:   pgUUIDToUUID(row.UserID),
		IsLead:   row.IsLead,
		CreatedAt: *createdAt,
	}
}
