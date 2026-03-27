-- name: CreateTenantBranding :one
INSERT INTO tenant_branding (tenant_id, logo_light_url, logo_dark_url, favicon_url, primary_color, secondary_color, accent_color, font_family, app_name, custom_css, email_header_html)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetTenantBranding :one
SELECT * FROM tenant_branding WHERE tenant_id = $1;

-- name: UpdateTenantBranding :one
UPDATE tenant_branding
SET 
    logo_light_url = COALESCE($2, logo_light_url),
    logo_dark_url = COALESCE($3, logo_dark_url),
    favicon_url = COALESCE($4, favicon_url),
    primary_color = COALESCE($5, primary_color),
    secondary_color = COALESCE($6, secondary_color),
    accent_color = COALESCE($7, accent_color),
    font_family = COALESCE($8, font_family),
    app_name = COALESCE($9, app_name),
    custom_css = COALESCE($10, custom_css),
    email_header_html = COALESCE($11, email_header_html)
WHERE tenant_id = $1
RETURNING *;
