package invitation

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Invitation struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Email      string     `json:"email"`
	Token      string     `json:"-"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	RoleName   string     `json:"role_name"`
	InvitedBy  *uuid.UUID `json:"invited_by,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type InvitedByUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type InvitationWithInviter struct {
	Invitation
	Inviter *InvitedByUser `json:"inviter,omitempty"`
}

func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

func (i *Invitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

func (i *Invitation) IsValid() bool {
	return !i.IsExpired() && !i.IsAccepted()
}

type CreateInvitationRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	RoleName  string `json:"role_name" validate:"required,oneof=Admin Member Read Only Sales Manager"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, invitedBy uuid.UUID, req *CreateInvitationRequest, token string, expiresAt time.Time) (*Invitation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*Invitation, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*InvitationWithInviter, int, error)
	MarkAccepted(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByEmail(ctx context.Context, tenantID uuid.UUID, email string) error
}
