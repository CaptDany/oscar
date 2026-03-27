package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, req *user.CreateUserRequest, passwordHash string) (*user.User, error) {
	query := `
		INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, avatar_url, timezone, locale)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	var row generated.User
	err := r.pool.QueryRow(ctx, query,
		req.TenantID, req.Email, passwordHash, req.FirstName, req.LastName,
		req.AvatarURL, req.Timezone, req.Locale,
	).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
		&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user.Create: %w", err)
	}

	return mapUserRowToDomain(&row), nil
}

func (r *UserRepository) CreateTx(ctx context.Context, tx pgx.Tx, req *user.CreateUserRequest, passwordHash string) (*user.User, error) {
	query := `
		INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, avatar_url, timezone, locale)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	var row generated.User
	err := tx.QueryRow(ctx, query,
		req.TenantID, req.Email, passwordHash, req.FirstName, req.LastName,
		req.AvatarURL, req.Timezone, req.Locale,
	).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
		&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user.Create: %w", err)
	}

	return mapUserRowToDomain(&row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`

	var row generated.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
		&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user.GetByID: user not found")
		}
		return nil, fmt.Errorf("user.GetByID: %w", err)
	}

	return mapUserRowToDomain(&row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*user.User, error) {
	query := `SELECT * FROM users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL`

	var row generated.User
	err := r.pool.QueryRow(ctx, query, tenantID, email).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
		&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user.GetByEmail: user not found")
		}
		return nil, fmt.Errorf("user.GetByEmail: %w", err)
	}

	return mapUserRowToDomain(&row), nil
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, req *user.UpdateUserRequest) (*user.User, error) {
	query := `
		UPDATE users SET
			email = COALESCE($2, email),
			first_name = COALESCE($3, first_name),
			last_name = COALESCE($4, last_name),
			avatar_url = COALESCE($5, avatar_url),
			timezone = COALESCE($6, timezone),
			locale = COALESCE($7, locale),
			is_active = COALESCE($8, is_active)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	var row generated.User
	err := r.pool.QueryRow(ctx, query,
		id, req.Email, req.FirstName, req.LastName, req.AvatarURL, req.Timezone, req.Locale, req.IsActive,
	).Scan(
		&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
		&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user.Update: %w", err)
	}

	return mapUserRowToDomain(&row), nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id, passwordHash)
	if err != nil {
		return fmt.Errorf("user.UpdatePassword: %w", err)
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("user.SoftDelete: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*user.User, int, error) {
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user.List count: %w", err)
	}

	query := `
		SELECT * FROM users 
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user.List: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var row generated.User
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Email, &row.PasswordHash, &row.FirstName, &row.LastName,
			&row.AvatarURL, &row.Timezone, &row.Locale, &row.IsActive, &row.LastLoginAt,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("user.List scan: %w", err)
		}
		users = append(users, mapUserRowToDomain(&row))
	}

	return users, total, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("user.UpdateLastLogin: %w", err)
	}
	return nil
}

func mapUserRowToDomain(row *generated.User) *user.User {
	return &user.User{
		ID:           row.ID,
		TenantID:     row.TenantID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FirstName:    row.FirstName,
		LastName:      row.LastName,
		AvatarURL:    row.AvatarURL,
		Timezone:     row.Timezone,
		Locale:       row.Locale,
		IsActive:     row.IsActive,
		LastLoginAt:   row.LastLoginAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    row.DeletedAt,
	}
}

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.Role, error) {
	query := `SELECT * FROM roles WHERE id = $1`

	var row generated.Role
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsSystem, &row.Permissions,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role.GetByID: role not found")
		}
		return nil, fmt.Errorf("role.GetByID: %w", err)
	}

	return mapRoleRowToDomain(&row), nil
}

func (r *RoleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*user.Role, error) {
	query := `SELECT * FROM roles WHERE tenant_id = $1 AND name = $2`

	var row generated.Role
	err := r.pool.QueryRow(ctx, query, tenantID, name).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsSystem, &row.Permissions,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role.GetByName: role not found")
		}
		return nil, fmt.Errorf("role.GetByName: %w", err)
	}

	return mapRoleRowToDomain(&row), nil
}

func (r *RoleRepository) GetSystemRoles(ctx context.Context, tenantID uuid.UUID) ([]user.Role, error) {
	query := `SELECT * FROM roles WHERE tenant_id = $1 AND is_system = true`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("role.GetSystemRoles: %w", err)
	}
	defer rows.Close()

	var roles []user.Role
	for rows.Next() {
		var row generated.Role
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsSystem, &row.Permissions,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("role.GetSystemRoles scan: %w", err)
		}
		roles = append(roles, *mapRoleRowToDomain(&row))
	}

	return roles, nil
}

func (r *RoleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]user.Role, error) {
	query := `SELECT * FROM roles WHERE tenant_id = $1 ORDER BY is_system DESC, name ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("role.List: %w", err)
	}
	defer rows.Close()

	var roles []user.Role
	for rows.Next() {
		var row generated.Role
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsSystem, &row.Permissions,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("role.List scan: %w", err)
		}
		roles = append(roles, *mapRoleRowToDomain(&row))
	}

	return roles, nil
}

func (r *RoleRepository) AssignToUser(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	for _, roleID := range roleIDs {
		query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err := r.pool.Exec(ctx, query, userID, roleID)
		if err != nil {
			return fmt.Errorf("role.AssignToUser: %w", err)
		}
	}
	return nil
}

func (r *RoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]user.Role, error) {
	query := `
		SELECT r.* FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("role.GetUserRoles: %w", err)
	}
	defer rows.Close()

	var roles []user.Role
	for rows.Next() {
		var row generated.Role
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.IsSystem, &row.Permissions,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("role.GetUserRoles scan: %w", err)
		}
		roles = append(roles, *mapRoleRowToDomain(&row))
	}

	return roles, nil
}

func (r *RoleRepository) GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT r.name FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("role.GetUserRoleNames: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("role.GetUserRoleNames scan: %w", err)
		}
		names = append(names, name)
	}

	return names, nil
}

func mapRoleRowToDomain(row *generated.Role) *user.Role {
	permissions := make(map[string]user.Permission)
	if row.Permissions != nil {
		json.Unmarshal(row.Permissions, &permissions)
	}

	return &user.Role{
		ID:          row.ID,
		TenantID:    row.TenantID,
		Name:        row.Name,
		Description: row.Description,
		IsSystem:    row.IsSystem,
		Permissions: permissions,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
