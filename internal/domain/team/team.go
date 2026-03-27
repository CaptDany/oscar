package team

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Name        string       `json:"name"`
	Description *string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type TeamMember struct {
	ID       uuid.UUID `json:"id"`
	TeamID   uuid.UUID `json:"team_id"`
	UserID   uuid.UUID `json:"user_id"`
	IsLead   bool      `json:"is_lead"`
	Email    string    `json:"email,omitempty"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type CreateTeamRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	IsLead bool      `json:"is_lead"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateTeamRequest) (*Team, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*Team, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateTeamRequest) (*Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, teamID, userID uuid.UUID, isLead bool) (*TeamMember, error)
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	ListMembers(ctx context.Context, teamID uuid.UUID) ([]TeamMember, error)
	ListUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	SetLead(ctx context.Context, teamID, userID uuid.UUID) error
}
