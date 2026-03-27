package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/company"
)

type CompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

func (r *CompanyRepository) Create(ctx context.Context, tenantID uuid.UUID, req *company.CreateCompanyRequest) (*company.Company, error) {
	query := `
		INSERT INTO companies (tenant_id, name, domain, industry, size, annual_revenue, website, address, owner_id, parent_company_id, tags, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *
	`

	row := &generated.Company{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Name, req.Domain, req.Industry, req.Size,
		req.AnnualRevenue, req.Website, req.Address, req.OwnerID,
		req.ParentCompanyID, req.Tags, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
		&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
		&row.ParentCompanyID, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("company.Create: %w", err)
	}

	return mapCompanyRowToDomain(row), nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	query := `SELECT * FROM companies WHERE id = $1 AND deleted_at IS NULL`

	row := &generated.Company{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
		&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
		&row.ParentCompanyID, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("company.GetByID: company not found")
		}
		return nil, fmt.Errorf("company.GetByID: %w", err)
	}

	return mapCompanyRowToDomain(row), nil
}

func (r *CompanyRepository) Update(ctx context.Context, id uuid.UUID, req *company.UpdateCompanyRequest) (*company.Company, error) {
	query := `
		UPDATE companies SET
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
		RETURNING *
	`

	row := &generated.Company{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Name, req.Domain, req.Industry, req.Size,
		req.AnnualRevenue, req.Website, req.Address, req.OwnerID,
		req.ParentCompanyID, req.Tags, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
		&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
		&row.ParentCompanyID, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("company.Update: %w", err)
	}

	return mapCompanyRowToDomain(row), nil
}

func (r *CompanyRepository) SoftDelete(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	query := `UPDATE companies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING *`

	row := &generated.Company{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
		&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
		&row.ParentCompanyID, &row.Tags, &row.CustomFields,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("company.SoftDelete: %w", err)
	}

	return mapCompanyRowToDomain(row), nil
}

func (r *CompanyRepository) List(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) ([]*company.Company, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.OwnerID != nil {
		baseQuery += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *filter.OwnerID)
		argIdx++
	}
	if filter.Industry != nil {
		baseQuery += fmt.Sprintf(" AND industry = $%d", argIdx)
		args = append(args, *filter.Industry)
		argIdx++
	}
	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR COALESCE(domain, '') ILIKE $%d OR COALESCE(industry, '') ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM companies ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("company.List count: %w", err)
	}

	listQuery := `SELECT * FROM companies ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, 0)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("company.List: %w", err)
	}
	defer rows.Close()

	var companies []*company.Company
	for rows.Next() {
		row := &generated.Company{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
			&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
			&row.ParentCompanyID, &row.Tags, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("company.List scan: %w", err)
		}
		companies = append(companies, mapCompanyRowToDomain(row))
	}

	nextCursor := ""
	if len(companies) > limit {
		companies = companies[:limit]
		nextCursor = companies[len(companies)-1].ID.String()
	}

	return companies, nextCursor, total, nil
}

func (r *CompanyRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*company.Company, error) {
	if limit <= 0 {
		limit = 20
	}

	sql := `
		SELECT * FROM companies 
		WHERE tenant_id = $1 
		  AND deleted_at IS NULL
		  AND (
		    name ILIKE $2 OR
		    COALESCE(domain, '') ILIKE $2 OR
		    COALESCE(industry, '') ILIKE $2
		  )
		ORDER BY name ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, sql, tenantID, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("company.Search: %w", err)
	}
	defer rows.Close()

	var companies []*company.Company
	for rows.Next() {
		row := &generated.Company{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Domain, &row.Industry, &row.Size,
			&row.AnnualRevenue, &row.Website, &row.Address, &row.OwnerID,
			&row.ParentCompanyID, &row.Tags, &row.CustomFields,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("company.Search scan: %w", err)
		}
		companies = append(companies, mapCompanyRowToDomain(row))
	}

	return companies, nil
}

func (r *CompanyRepository) Count(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) (int, error) {
	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}

	if filter.Industry != nil {
		baseQuery += " AND industry = $2"
		args = append(args, *filter.Industry)
	}

	countQuery := `SELECT COUNT(*) FROM companies ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("company.Count: %w", err)
	}

	return total, nil
}

func mapCompanyRowToDomain(row *generated.Company) *company.Company {
	var size *company.CompanySize
	if row.Size.Valid {
		s := company.CompanySize(row.Size.CompanySize)
		size = &s
	}
	return &company.Company{
		ID:              pgUUIDToUUID(row.ID),
		TenantID:        pgUUIDToUUID(row.TenantID),
		Name:            row.Name,
		Domain:          pgTextToStr(row.Domain),
		Industry:        pgTextToStr(row.Industry),
		Size:            size,
		AnnualRevenue:   pgNumericToPtrFloat(row.AnnualRevenue),
		Website:         pgTextToStr(row.Website),
		Address:         row.Address,
		OwnerID:         pgUUIDToPtr(row.OwnerID),
		ParentCompanyID: pgUUIDToPtr(row.ParentCompanyID),
		Tags:            row.Tags,
		CustomFields:    row.CustomFields,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       pgTimestamptzToTime(row.DeletedAt),
	}
}
