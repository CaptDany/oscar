package tenant

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

type SubscriptionTier string

const (
	TierFree         SubscriptionTier = "free"
	TierStarter      SubscriptionTier = "starter"
	TierProfessional SubscriptionTier = "professional"
	TierEnterprise   SubscriptionTier = "enterprise"
)

type Tenant struct {
	ID                uuid.UUID        `json:"id"`
	Slug              string           `json:"slug"`
	Name              string           `json:"name"`
	Status            Status           `json:"status"`
	SubscriptionTier  SubscriptionTier `json:"subscription_tier"`
	Settings          json.RawMessage  `json:"settings"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type TenantBranding struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	LogoLightURL    *string   `json:"logo_light_url"`
	LogoDarkURL     *string   `json:"logo_dark_url"`
	FaviconURL      *string   `json:"favicon_url"`
	PrimaryColor    string    `json:"primary_color"`
	SecondaryColor  string    `json:"secondary_color"`
	AccentColor     string    `json:"accent_color"`
	FontFamily      string    `json:"font_family"`
	AppName         string    `json:"app_name"`
	CustomCSS       *string   `json:"custom_css"`
	EmailHeaderHTML *string   `json:"email_header_html"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	Slug             string           `json:"slug" validate:"required,min=2,max=63,alphanumdash"`
	Name             string           `json:"name" validate:"required,min=2,max=255"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier"`
}

type UpdateTenantRequest struct {
	Name             *string           `json:"name"`
	Status           *Status           `json:"status"`
	SubscriptionTier *SubscriptionTier `json:"subscription_tier"`
	Settings         json.RawMessage   `json:"settings"`
}

type UpdateBrandingRequest struct {
	LogoLightURL    *string `json:"logo_light_url"`
	LogoDarkURL     *string `json:"logo_dark_url"`
	FaviconURL      *string `json:"favicon_url"`
	PrimaryColor    *string `json:"primary_color"`
	SecondaryColor  *string `json:"secondary_color"`
	AccentColor     *string `json:"accent_color"`
	FontFamily      *string `json:"font_family"`
	AppName         *string `json:"app_name"`
	CustomCSS       *string `json:"custom_css"`
	EmailHeaderHTML *string `json:"email_header_html"`
}

type Repository interface {
	Create(ctx context.Context, req *CreateTenantRequest) (*Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*Tenant, error)
	SeedRoles(ctx context.Context, tenantID uuid.UUID) error
	SeedPipeline(ctx context.Context, tenantID uuid.UUID) error
}

type BrandingRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID) (*TenantBranding, error)
	Get(ctx context.Context, tenantID uuid.UUID) (*TenantBranding, error)
	Update(ctx context.Context, tenantID uuid.UUID, req *UpdateBrandingRequest) (*TenantBranding, error)
}
