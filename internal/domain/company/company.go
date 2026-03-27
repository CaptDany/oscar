package company

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CompanySize string

const (
	CompanySizeStartup   CompanySize = "startup"
	CompanySizeSmall     CompanySize = "small"
	CompanySizeMedium    CompanySize = "medium"
	CompanySizeLarge     CompanySize = "large"
	CompanySizeEnterprise CompanySize = "enterprise"
)

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type Company struct {
	ID             uuid.UUID    `json:"id"`
	TenantID       uuid.UUID    `json:"tenant_id"`
	Name           string       `json:"name"`
	Domain         *string      `json:"domain"`
	Industry       *string      `json:"industry"`
	Size           *CompanySize `json:"size"`
	AnnualRevenue  *float64     `json:"annual_revenue"`
	Website        *string      `json:"website"`
	Address        interface{}  `json:"address"`
	OwnerID        *uuid.UUID   `json:"owner_id"`
	ParentCompanyID *uuid.UUID  `json:"parent_company_id"`
	Tags           []string     `json:"tags"`
	CustomFields   interface{}  `json:"custom_fields"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedAt      *time.Time   `json:"-"`
}

type CreateCompanyRequest struct {
	Name           string        `json:"name" validate:"required,min=1,max=255"`
	Domain         *string       `json:"domain"`
	Industry       *string       `json:"industry"`
	Size           *CompanySize  `json:"size"`
	AnnualRevenue  *float64      `json:"annual_revenue"`
	Website        *string       `json:"website"`
	Address        interface{}   `json:"address"`
	OwnerID        *uuid.UUID    `json:"owner_id"`
	ParentCompanyID *uuid.UUID   `json:"parent_company_id"`
	Tags           []string      `json:"tags"`
	CustomFields   interface{}   `json:"custom_fields"`
}

type UpdateCompanyRequest struct {
	Name           *string       `json:"name"`
	Domain         *string       `json:"domain"`
	Industry       *string       `json:"industry"`
	Size           *CompanySize  `json:"size"`
	AnnualRevenue  *float64      `json:"annual_revenue"`
	Website        *string       `json:"website"`
	Address        interface{}   `json:"address"`
	OwnerID        *uuid.UUID    `json:"owner_id"`
	ParentCompanyID *uuid.UUID   `json:"parent_company_id"`
	Tags           []string      `json:"tags"`
	CustomFields   interface{}   `json:"custom_fields"`
}

type ListCompaniesFilter struct {
	OwnerID  *uuid.UUID
	Industry *string
	Search   string
	Cursor   string
	Limit    int
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateCompanyRequest) (*Company, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateCompanyRequest) (*Company, error)
	SoftDelete(ctx context.Context, id uuid.UUID) (*Company, error)
	List(ctx context.Context, tenantID uuid.UUID, filter *ListCompaniesFilter) ([]*Company, string, int, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*Company, error)
	Count(ctx context.Context, tenantID uuid.UUID, filter *ListCompaniesFilter) (int, error)
}
