package person

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PersonType string

const (
	PersonTypeLead    PersonType = "lead"
	PersonTypeContact PersonType = "contact"
	PersonTypeCustomer PersonType = "customer"
)

type PersonStatus string

const (
	PersonStatusNew        PersonStatus = "new"
	PersonStatusContacted  PersonStatus = "contacted"
	PersonStatusQualified  PersonStatus = "qualified"
	PersonStatusUnqualified PersonStatus = "unqualified"
	PersonStatusActive     PersonStatus = "active"
	PersonStatusInactive   PersonStatus = "inactive"
)

type PersonSource string

const (
	PersonSourceWebsite  PersonSource = "website"
	PersonSourceReferral PersonSource = "referral"
	PersonSourceSocial   PersonSource = "social"
	PersonSourceEmail    PersonSource = "email"
	PersonSourcePhone    PersonSource = "phone"
	PersonSourceEvent    PersonSource = "event"
	PersonSourceOther    PersonSource = "other"
)

type Person struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Type        PersonType   `json:"type"`
	Status      PersonStatus `json:"status"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	Email       []string     `json:"email"`
	Phone       []string     `json:"phone"`
	AvatarURL   *string      `json:"avatar_url"`
	CompanyID   *uuid.UUID   `json:"company_id"`
	OwnerID     *uuid.UUID   `json:"owner_id"`
	Source      *PersonSource `json:"source"`
	Score       int          `json:"score"`
	Tags        []string     `json:"tags"`
	CustomFields interface{} `json:"custom_fields"`
	ConvertedAt *time.Time   `json:"converted_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"-"`
}

func (p *Person) FullName() string {
	return p.FirstName + " " + p.LastName
}

func (p *Person) PrimaryEmail() string {
	if len(p.Email) > 0 {
		return p.Email[0]
	}
	return ""
}

func (p *Person) PrimaryPhone() string {
	if len(p.Phone) > 0 {
		return p.Phone[0]
	}
	return ""
}

type CreatePersonRequest struct {
	Type         PersonType    `json:"type" validate:"required,oneof=lead contact customer"`
	Status       PersonStatus  `json:"status"`
	FirstName    string        `json:"first_name" validate:"required,min=1,max=100"`
	LastName     string        `json:"last_name" validate:"required,min=1,max=100"`
	Email        []string      `json:"email"`
	Phone        []string      `json:"phone"`
	AvatarURL    *string       `json:"avatar_url"`
	CompanyID    *uuid.UUID    `json:"company_id"`
	OwnerID      *uuid.UUID    `json:"owner_id"`
	Source       *PersonSource `json:"source"`
	Tags         []string      `json:"tags"`
	CustomFields interface{}   `json:"custom_fields"`
}

type UpdatePersonRequest struct {
	Type         *PersonType   `json:"type"`
	Status       *PersonStatus `json:"status"`
	FirstName    *string       `json:"first_name"`
	LastName     *string       `json:"last_name"`
	Email        []string      `json:"email"`
	Phone        []string      `json:"phone"`
	AvatarURL    *string       `json:"avatar_url"`
	CompanyID    *uuid.UUID    `json:"company_id"`
	OwnerID      *uuid.UUID    `json:"owner_id"`
	Source       *PersonSource `json:"source"`
	Score        *int          `json:"score"`
	Tags         []string      `json:"tags"`
	CustomFields interface{}   `json:"custom_fields"`
}

type ConvertPersonRequest struct {
	Type   PersonType   `json:"type" validate:"required,oneof=contact customer"`
	Status PersonStatus `json:"status" validate:"required"`
}

type ListPersonsFilter struct {
	Type     PersonType
	Status   PersonStatus
	OwnerID  *uuid.UUID
	CompanyID *uuid.UUID
	Search   string
	Tags     []string
	Cursor   string
	Limit    int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreatePersonRequest) (*Person, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Person, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdatePersonRequest) (*Person, error)
	SoftDelete(ctx context.Context, id uuid.UUID) (*Person, error)
	Convert(ctx context.Context, id uuid.UUID, toType PersonType, status PersonStatus) (*Person, error)
	List(ctx context.Context, tenantID uuid.UUID, filter *ListPersonsFilter) ([]*Person, string, int, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*Person, error)
	Count(ctx context.Context, tenantID uuid.UUID, filter *ListPersonsFilter) (int, error)
	AddTag(ctx context.Context, id uuid.UUID, tag string) (*Person, error)
	RemoveTag(ctx context.Context, id uuid.UUID, tag string) (*Person, error)
	UpdateScore(ctx context.Context, id uuid.UUID, score int) (*Person, error)
}
