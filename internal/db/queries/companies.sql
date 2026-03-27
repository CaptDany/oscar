-- name: CreateCompany :one
INSERT INTO companies (tenant_id, name, domain, industry, size, annual_revenue, website, address, owner_id, parent_company_id, tags, custom_fields)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetCompanyByID :one
SELECT * FROM companies WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateCompany :one
UPDATE companies
SET 
    name = COALESCE($2, name),
    domain = COALESCE($3, domain),
    industry = COALESCE($4, industry),
    size = COALESCE($5, size),
    annual_revenue = COALESCE($6, annual_revenue),
    website = COALESCE($7, website),
    address = COALESCE($8, address),
    owner_id = COALESCE($9, owner_id),
    parent_company_id = COALESCE($10, parent_company_id),
    tags = COALESCE($11, tags),
    custom_fields = COALESCE($12, custom_fields)
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCompany :one
UPDATE companies
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListCompanies :many
SELECT * FROM companies 
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCompaniesByOwner :many
SELECT * FROM companies 
WHERE tenant_id = $1 AND owner_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: SearchCompanies :many
SELECT * FROM companies 
WHERE tenant_id = $1 
  AND deleted_at IS NULL
  AND (
    name ILIKE '%' || $2 || '%' OR
    COALESCE(domain, '') ILIKE '%' || $2 || '%' OR
    COALESCE(industry, '') ILIKE '%' || $2 || '%'
  )
ORDER BY name ASC
LIMIT $3 OFFSET $4;

-- name: CountCompanies :one
SELECT COUNT(*) FROM companies WHERE tenant_id = $1 AND deleted_at IS NULL;
