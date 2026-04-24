package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/tenant"
)

type TenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

func (r *TenantRepository) Create(ctx context.Context, req *tenant.CreateTenantRequest) (*tenant.Tenant, error) {
	query := `
		INSERT INTO tenants (slug, name, status, subscription_tier)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slug, name, status, subscription_tier, settings, created_at, updated_at, invite_only
	`

	status := "active"
	tier := string(req.SubscriptionTier)
	if tier == "" {
		tier = "free"
	}

	var row generated.Tenant
	err := r.pool.QueryRow(ctx, query,
		req.Slug, req.Name, status, tier,
	).Scan(
		&row.ID, &row.Slug, &row.Name, &row.Status, &row.SubscriptionTier, &row.Settings,
		&row.CreatedAt, &row.UpdatedAt, &row.InviteOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant.Create: %w", err)
	}

	return mapTenantRowToDomain(&row), nil
}

func (r *TenantRepository) CreateTx(ctx context.Context, tx pgx.Tx, req *tenant.CreateTenantRequest) (*tenant.Tenant, error) {
	query := `
		INSERT INTO tenants (slug, name, status, subscription_tier)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slug, name, status, subscription_tier, settings, created_at, updated_at, invite_only
	`

	status := "active"
	tier := string(req.SubscriptionTier)
	if tier == "" {
		tier = "free"
	}

	var row generated.Tenant
	err := tx.QueryRow(ctx, query,
		req.Slug, req.Name, status, tier,
	).Scan(
		&row.ID, &row.Slug, &row.Name, &row.Status, &row.SubscriptionTier, &row.Settings,
		&row.CreatedAt, &row.UpdatedAt, &row.InviteOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant.Create: %w", err)
	}

	return mapTenantRowToDomain(&row), nil
}

func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	query := `SELECT id, slug, name, status, subscription_tier, settings, created_at, updated_at, invite_only FROM tenants WHERE id = $1`

	var row generated.Tenant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.Slug, &row.Name, &row.Status, &row.SubscriptionTier, &row.Settings,
		&row.CreatedAt, &row.UpdatedAt, &row.InviteOnly,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant.GetByID: tenant not found")
		}
		return nil, fmt.Errorf("tenant.GetByID: %w", err)
	}

	return mapTenantRowToDomain(&row), nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*tenant.Tenant, error) {
	query := `SELECT id, slug, name, status, subscription_tier, settings, created_at, updated_at, invite_only FROM tenants WHERE slug = $1`

	var row generated.Tenant
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&row.ID, &row.Slug, &row.Name, &row.Status, &row.SubscriptionTier, &row.Settings,
		&row.CreatedAt, &row.UpdatedAt, &row.InviteOnly,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant.GetBySlug: tenant not found")
		}
		return nil, fmt.Errorf("tenant.GetBySlug: %w", err)
	}

	return mapTenantRowToDomain(&row), nil
}

func (r *TenantRepository) Update(ctx context.Context, id uuid.UUID, req *tenant.UpdateTenantRequest) (*tenant.Tenant, error) {
	query := `
		UPDATE tenants SET
			name = COALESCE($2, name),
			status = COALESCE($3, status),
			subscription_tier = COALESCE($4, subscription_tier),
			settings = COALESCE($5, settings)
		WHERE id = $1
		RETURNING id, slug, name, status, subscription_tier, settings, created_at, updated_at, invite_only
	`

	var settings []byte
	if req.Settings != nil {
		settings = req.Settings
	}

	var row generated.Tenant
	err := r.pool.QueryRow(ctx, query,
		id, req.Name, req.Status, req.SubscriptionTier, settings,
	).Scan(
		&row.ID, &row.Slug, &row.Name, &row.Status, &row.SubscriptionTier, &row.Settings,
		&row.CreatedAt, &row.UpdatedAt, &row.InviteOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant.Update: %w", err)
	}

	return mapTenantRowToDomain(&row), nil
}

func (r *TenantRepository) SeedRoles(ctx context.Context, tenantID uuid.UUID) error {
	query := `SELECT seed_tenant_roles($1)`
	_, err := r.pool.Exec(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("tenant.SeedRoles: %w", err)
	}
	return nil
}

func (r *TenantRepository) SeedPipeline(ctx context.Context, tenantID uuid.UUID) error {
	query := `SELECT seed_tenant_pipeline($1)`
	_, err := r.pool.Exec(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("tenant.SeedPipeline: %w", err)
	}
	return nil
}

func mapTenantRowToDomain(row *generated.Tenant) *tenant.Tenant {
	settings := json.RawMessage{}
	if row.Settings != nil {
		settings = row.Settings
	}
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

	return &tenant.Tenant{
		ID:               pgUUIDToUUID(row.ID),
		Slug:             row.Slug,
		Name:             row.Name,
		Status:           tenant.Status(row.Status),
		SubscriptionTier: tenant.SubscriptionTier(row.SubscriptionTier),
		Settings:         settings,
		CreatedAt:        *createdAt,
		UpdatedAt:        *updatedAt,
	}
}

type BrandingRepository struct {
	pool *pgxpool.Pool
}

func NewBrandingRepository(pool *pgxpool.Pool) *BrandingRepository {
	return &BrandingRepository{pool: pool}
}

func (r *BrandingRepository) Get(ctx context.Context, tenantID uuid.UUID) (*tenant.TenantBranding, error) {
	query := `SELECT * FROM tenant_branding WHERE tenant_id = $1`

var row generated.TenantBranding
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&row.TenantID, &row.LogoLightUrl, &row.LogoDarkUrl, &row.FaviconUrl,
		&row.PrimaryColor, &row.SecondaryColor, &row.AccentColor,
		&row.FontFamily, &row.AppName, &row.CustomCss, &row.EmailHeaderHtml,
		&row.CreatedAt, &row.UpdatedAt, &row.MonoFont,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("branding.Get: branding not found")
		}
		return nil, fmt.Errorf("branding.Get: %w", err)
	}

	return mapBrandingRowToDomain(&row), nil
}

func (r *BrandingRepository) Create(ctx context.Context, tenantID uuid.UUID) (*tenant.TenantBranding, error) {
	query := `
		INSERT INTO tenant_branding (tenant_id)
		VALUES ($1)
		RETURNING *
	`

	var row generated.TenantBranding
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&row.TenantID, &row.LogoLightUrl, &row.LogoDarkUrl, &row.FaviconUrl,
		&row.PrimaryColor, &row.SecondaryColor, &row.AccentColor,
		&row.FontFamily, &row.AppName, &row.CustomCss, &row.EmailHeaderHtml,
		&row.CreatedAt, &row.UpdatedAt, &row.MonoFont,
	)
	if err != nil {
		return nil, fmt.Errorf("branding.Create: %w", err)
	}

	return mapBrandingRowToDomain(&row), nil
}

func (r *BrandingRepository) CreateTx(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (*tenant.TenantBranding, error) {
	query := `
		INSERT INTO tenant_branding (tenant_id)
		VALUES ($1)
		RETURNING *
	`

	var row generated.TenantBranding
	err := tx.QueryRow(ctx, query, tenantID).Scan(
		&row.TenantID, &row.LogoLightUrl, &row.LogoDarkUrl, &row.FaviconUrl,
		&row.PrimaryColor, &row.SecondaryColor, &row.AccentColor,
		&row.FontFamily, &row.AppName, &row.CustomCss, &row.EmailHeaderHtml,
		&row.CreatedAt, &row.UpdatedAt, &row.MonoFont,
	)
	if err != nil {
		return nil, fmt.Errorf("branding.Create: %w", err)
	}

	return mapBrandingRowToDomain(&row), nil
}

func (r *BrandingRepository) Update(ctx context.Context, tenantID uuid.UUID, req *tenant.UpdateBrandingRequest) (*tenant.TenantBranding, error) {
	setClauses := []string{}
	args := []interface{}{}

	if req.PrimaryColor != nil {
		setClauses = append(setClauses, "primary_color")
		args = append(args, *req.PrimaryColor)
	}
	if req.SecondaryColor != nil {
		setClauses = append(setClauses, "secondary_color")
		args = append(args, *req.SecondaryColor)
	}
	if req.AccentColor != nil {
		setClauses = append(setClauses, "accent_color")
		args = append(args, *req.AccentColor)
	}
	if req.FontFamily != nil {
		setClauses = append(setClauses, "font_family")
		args = append(args, *req.FontFamily)
	}
	if req.MonoFont != nil {
		setClauses = append(setClauses, "mono_font")
		args = append(args, *req.MonoFont)
	}
	if req.AppName != nil {
		setClauses = append(setClauses, "app_name")
		args = append(args, *req.AppName)
	}

	if len(setClauses) == 0 {
		return r.Get(ctx, tenantID)
	}

	placeholders := make([]string, len(setClauses))
	updateSet := make([]string, len(setClauses))
	for i := range setClauses {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		updateSet[i] = setClauses[i] + " = excluded." + setClauses[i]
	}

	query := fmt.Sprintf(`
		INSERT INTO tenant_branding (tenant_id, %s)
		VALUES ($1, %s)
		ON CONFLICT (tenant_id) DO UPDATE SET %s
	`, strings.Join(setClauses, ", "), strings.Join(placeholders, ", "), strings.Join(updateSet, ", "))

	args = append([]interface{}{tenantID}, args...)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("branding.Update: %w", err)
	}

	return r.Get(ctx, tenantID)
}

func mapBrandingRowToDomain(row *generated.TenantBranding) *tenant.TenantBranding {
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
	return &tenant.TenantBranding{
		TenantID:        pgUUIDToUUID(row.TenantID),
		LogoLightURL:    pgTextToStr(row.LogoLightUrl),
		LogoDarkURL:     pgTextToStr(row.LogoDarkUrl),
		FaviconURL:      pgTextToStr(row.FaviconUrl),
		PrimaryColor:    pgTextToStrStr(row.PrimaryColor),
		SecondaryColor:  pgTextToStrStr(row.SecondaryColor),
		AccentColor:     pgTextToStrStr(row.AccentColor),
		FontFamily:      pgTextToStrStr(row.FontFamily),
		AppName:         pgTextToStrStr(row.AppName),
		CustomCSS:       pgTextToStr(row.CustomCss),
		EmailHeaderHTML: pgTextToStr(row.EmailHeaderHtml),
		CreatedAt:       *createdAt,
		UpdatedAt:       *updatedAt,
		MonoFont:        pgTextToStrStr(row.MonoFont),
	}
}

type TenantPool struct {
	*pgxpool.Pool
}

func (p *TenantPool) SetTenantContext(ctx context.Context, tenantID uuid.UUID) (context.Context, pgx.Tx, error) {
	tx, err := p.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ctx, nil, err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", tenantID.String()))
	if err != nil {
		tx.Rollback(ctx)
		return ctx, nil, err
	}
	return context.WithValue(ctx, "tx", tx), tx, nil
}

func NewTenantPool(pool *pgxpool.Pool) *TenantPool {
	return &TenantPool{Pool: pool}
}

func (r *BrandingRepository) UpdateBrandAssets(ctx context.Context, tenantID uuid.UUID, logoLightURL, logoDarkURL, faviconURL *string) error {
	query := `
		UPDATE tenant_branding SET
			logo_light_url = COALESCE($2, logo_light_url),
			logo_dark_url = COALESCE($3, logo_dark_url),
			favicon_url = COALESCE($4, favicon_url)
		WHERE tenant_id = $1
	`

	_, err := r.pool.Exec(ctx, query, tenantID, logoLightURL, logoDarkURL, faviconURL)
	if err != nil {
		return fmt.Errorf("branding.UpdateBrandAssets: %w", err)
	}
	return nil
}

type BrandAssetsUpdater interface {
	UpdateBrandAssets(ctx context.Context, tenantID uuid.UUID, logoLightURL, logoDarkURL, faviconURL *string) error
}
