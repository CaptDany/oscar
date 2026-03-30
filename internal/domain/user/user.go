package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	AvatarURL    *string    `json:"avatar_url"`
	Timezone     string     `json:"timezone"`
	Locale       string     `json:"locale"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type Role struct {
	ID          uuid.UUID             `json:"id"`
	TenantID    uuid.UUID             `json:"tenant_id"`
	Name        string                `json:"name"`
	Description *string               `json:"description"`
	IsSystem    bool                  `json:"is_system"`
	Permissions map[string]Permission `json:"permissions"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type Permission struct {
	View   string `json:"view"`
	Create string `json:"create"`
	Edit   string `json:"edit"`
	Delete string `json:"delete"`
	Export string `json:"export,omitempty"`
}

const (
	ScopeNone = "none"
	ScopeOwn  = "own"
	ScopeTeam = "team"
	ScopeAll  = "all"
)

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) HasPermission(resource, action string, roles []Role) bool {
	for _, role := range roles {
		perm, ok := role.Permissions[resource]
		if !ok {
			continue
		}
		switch action {
		case "view":
			return perm.View == ScopeAll || perm.View == ScopeTeam || perm.View == ScopeOwn
		case "create":
			return perm.Create == ScopeAll || perm.Create == ScopeTeam || perm.Create == ScopeOwn
		case "edit":
			return perm.Edit == ScopeAll || perm.Edit == ScopeTeam || perm.Edit == ScopeOwn
		case "delete":
			return perm.Delete == ScopeAll || perm.Delete == ScopeTeam || perm.Delete == ScopeOwn
		case "export":
			return perm.Export == ScopeAll || perm.Export == ScopeTeam || perm.Export == ScopeOwn
		}
	}
	return false
}

type CreateUserRequest struct {
	Email     string      `json:"email" validate:"required,email"`
	Password  string      `json:"password" validate:"required,min=8"`
	FirstName string      `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string      `json:"last_name" validate:"required,min=1,max=100"`
	Timezone  string      `json:"timezone"`
	Locale    string      `json:"locale"`
	RoleIDs   []uuid.UUID `json:"role_ids"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url"`
	Timezone  *string `json:"timezone"`
	Locale    *string `json:"locale"`
	IsActive  *bool   `json:"is_active"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url"`
	Timezone  string    `json:"timezone"`
	Locale    string    `json:"locale"`
	IsActive  bool      `json:"is_active"`
	Roles     []Role    `json:"roles"`
}

type InviteUserRequest struct {
	Email     string      `json:"email" validate:"required,email"`
	FirstName string      `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string      `json:"last_name" validate:"required,min=1,max=100"`
	RoleIDs   []uuid.UUID `json:"role_ids" validate:"required,min=1"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateUserRequest, passwordHash string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarKey string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*User, int, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*Role, error)
	GetSystemRoles(ctx context.Context, tenantID uuid.UUID) ([]Role, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]Role, error)
	AssignToUser(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	RemoveFromUser(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error)
	Create(ctx context.Context, tenantID uuid.UUID, name string, permissions map[string]Permission) (*Role, error)
}
