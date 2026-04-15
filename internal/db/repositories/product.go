package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/product"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, tenantID uuid.UUID, req *product.CreateProductRequest) (*product.Product, error) {
	query := `
		INSERT INTO products (tenant_id, name, description, sku, price, currency, unit, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	row := &generated.Product{}
	var description, sku, unit *string
	if req.Description != nil {
		description = req.Description
	}
	if req.SKU != nil {
		sku = req.SKU
	}
	unitStr := req.Unit
	if req.Unit == "" {
		unitStr = "unit"
	}
	unit = &unitStr

	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Name, description, sku, req.Price, req.Currency, unit, req.IsActive,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.Sku,
		&row.Price, &row.Currency, &row.Unit, &row.IsActive,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("product.Create: %w", err)
	}

	return mapProductRowToDomain(row), nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	query := `SELECT * FROM products WHERE id = $1`

	row := &generated.Product{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.Sku,
		&row.Price, &row.Currency, &row.Unit, &row.IsActive,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product.GetByID: product not found")
		}
		return nil, fmt.Errorf("product.GetByID: %w", err)
	}

	return mapProductRowToDomain(row), nil
}

func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, req *product.UpdateProductRequest) (*product.Product, error) {
	query := `
		UPDATE products
		SET 
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			sku = COALESCE($4, sku),
			price = COALESCE($5, price),
			currency = COALESCE($6, currency),
			unit = COALESCE($7, unit),
			is_active = COALESCE($8, is_active),
			updated_at = NOW()
		WHERE id = $1
		RETURNING *
	`

	row := &generated.Product{}
	var description, sku, unit *string
	if req.Description != nil {
		description = req.Description
	}
	if req.SKU != nil {
		sku = req.SKU
	}
	if req.Unit != nil {
		unit = req.Unit
	}

	err := r.pool.QueryRow(ctx, query,
		id, req.Name, description, sku, req.Price, req.Currency, unit, req.IsActive,
	).Scan(
		&row.ID, &row.TenantID, &row.Name, &row.Description, &row.Sku,
		&row.Price, &row.Currency, &row.Unit, &row.IsActive,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("product.Update: %w", err)
	}

	return mapProductRowToDomain(row), nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("product.Delete: %w", err)
	}
	return nil
}

func (r *ProductRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*product.Product, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM products WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("product.List count: %w", err)
	}

	query := `
		SELECT * FROM products
		WHERE tenant_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("product.List: %w", err)
	}
	defer rows.Close()

	var products []*product.Product
	for rows.Next() {
		row := &generated.Product{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.Sku,
			&row.Price, &row.Currency, &row.Unit, &row.IsActive,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("product.List scan: %w", err)
		}
		products = append(products, mapProductRowToDomain(row))
	}

	return products, total, nil
}

func (r *ProductRepository) ListActive(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*product.Product, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM products WHERE tenant_id = $1 AND is_active = true`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("product.ListActive count: %w", err)
	}

	query := `
		SELECT * FROM products
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("product.ListActive: %w", err)
	}
	defer rows.Close()

	var products []*product.Product
	for rows.Next() {
		row := &generated.Product{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Name, &row.Description, &row.Sku,
			&row.Price, &row.Currency, &row.Unit, &row.IsActive,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("product.ListActive scan: %w", err)
		}
		products = append(products, mapProductRowToDomain(row))
	}

	return products, total, nil
}

func mapProductRowToDomain(row *generated.Product) *product.Product {
	var desc, sku, unit *string
	if row.Description.Valid {
		desc = &row.Description.String
	}
	if row.Sku.Valid {
		sku = &row.Sku.String
	}
	if row.Unit.Valid {
		unit = &row.Unit.String
	}

	var price float64
	if row.Price.Valid {
		f, _ := row.Price.Float64Value()
		price = f.Float64
	}

	return &product.Product{
		ID:          pgUUIDToUUID(row.ID),
		TenantID:    pgUUIDToUUID(row.TenantID),
		Name:        row.Name,
		Description: desc,
		SKU:         sku,
		Price:       price,
		Currency:    row.Currency,
		Unit:        derefString(unit, "unit"),
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func derefString(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

type LineItemRepository struct {
	pool *pgxpool.Pool
}

func NewLineItemRepository(pool *pgxpool.Pool) *LineItemRepository {
	return &LineItemRepository{pool: pool}
}

func (r *LineItemRepository) Create(ctx context.Context, dealID uuid.UUID, req *product.CreateLineItemRequest) (*product.DealLineItem, error) {
	total := req.Quantity * req.UnitPrice * (1 - req.DiscountPct/100)

	query := `
		INSERT INTO deal_line_items (deal_id, product_id, quantity, unit_price, discount_pct, total)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	`

	row := &generated.DealLineItem{}
	err := r.pool.QueryRow(ctx, query,
		dealID, req.ProductID, req.Quantity, req.UnitPrice, req.DiscountPct, total,
	).Scan(
		&row.ID, &row.DealID, &row.ProductID, &row.Quantity, &row.UnitPrice,
		&row.DiscountPct, &row.Total, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("lineItem.Create: %w", err)
	}

	return mapLineItemRowToDomain(row), nil
}

func (r *LineItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*product.DealLineItem, error) {
	query := `SELECT * FROM deal_line_items WHERE id = $1`

	row := &generated.DealLineItem{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.DealID, &row.ProductID, &row.Quantity, &row.UnitPrice,
		&row.DiscountPct, &row.Total, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("lineItem.GetByID: line item not found")
		}
		return nil, fmt.Errorf("lineItem.GetByID: %w", err)
	}

	return mapLineItemRowToDomain(row), nil
}

func (r *LineItemRepository) Update(ctx context.Context, id uuid.UUID, req *product.UpdateLineItemRequest) (*product.DealLineItem, error) {
	query := `
		UPDATE deal_line_items
		SET 
			product_id = COALESCE($2, product_id),
			quantity = COALESCE($3, quantity),
			unit_price = COALESCE($4, unit_price),
			discount_pct = COALESCE($5, discount_pct),
			updated_at = NOW()
		WHERE id = $1
		RETURNING *
	`

	row := &generated.DealLineItem{}
	err := r.pool.QueryRow(ctx, query, id, req.ProductID, req.Quantity, req.UnitPrice, req.DiscountPct).Scan(
		&row.ID, &row.DealID, &row.ProductID, &row.Quantity, &row.UnitPrice,
		&row.DiscountPct, &row.Total, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("lineItem.Update: %w", err)
	}

	return mapLineItemRowToDomain(row), nil
}

func (r *LineItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM deal_line_items WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("lineItem.Delete: %w", err)
	}
	return nil
}

func (r *LineItemRepository) ListByDeal(ctx context.Context, dealID uuid.UUID) ([]*product.DealLineItemWithProduct, error) {
	query := `
		SELECT dli.*, p.name as product_name, p.sku as product_sku
		FROM deal_line_items dli
		LEFT JOIN products p ON dli.product_id = p.id
		WHERE dli.deal_id = $1
		ORDER BY dli.created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, dealID)
	if err != nil {
		return nil, fmt.Errorf("lineItem.ListByDeal: %w", err)
	}
	defer rows.Close()

	var items []*product.DealLineItemWithProduct
	for rows.Next() {
		row := &generated.ListDealLineItemsRow{}
		err := rows.Scan(
			&row.ID, &row.DealID, &row.ProductID, &row.Quantity, &row.UnitPrice,
			&row.DiscountPct, &row.Total, &row.CreatedAt, &row.UpdatedAt,
			&row.ProductName, &row.ProductSku,
		)
		if err != nil {
			return nil, fmt.Errorf("lineItem.ListByDeal scan: %w", err)
		}
		items = append(items, mapLineItemWithProductRowToDomain(row))
	}

	return items, nil
}

func (r *LineItemRepository) GetDealTotal(ctx context.Context, dealID uuid.UUID) (float64, error) {
	query := `SELECT COALESCE(SUM(total), 0) FROM deal_line_items WHERE deal_id = $1`

	var total float64
	err := r.pool.QueryRow(ctx, query, dealID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("lineItem.GetDealTotal: %w", err)
	}

	return total, nil
}

func mapLineItemRowToDomain(row *generated.DealLineItem) *product.DealLineItem {
	var quantity, unitPrice, discountPct, total float64
	if row.Quantity.Valid {
		f, _ := row.Quantity.Float64Value()
		quantity = f.Float64
	}
	if row.UnitPrice.Valid {
		f, _ := row.UnitPrice.Float64Value()
		unitPrice = f.Float64
	}
	if row.DiscountPct.Valid {
		f, _ := row.DiscountPct.Float64Value()
		discountPct = f.Float64
	}
	if row.Total.Valid {
		f, _ := row.Total.Float64Value()
		total = f.Float64
	}

	return &product.DealLineItem{
		ID:          pgUUIDToUUID(row.ID),
		DealID:      pgUUIDToUUID(row.DealID),
		ProductID:   pgUUIDToPtr(row.ProductID),
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		DiscountPct: discountPct,
		Total:       total,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapLineItemWithProductRowToDomain(row *generated.ListDealLineItemsRow) *product.DealLineItemWithProduct {
	item := &product.DealLineItemWithProduct{
		DealLineItem: *mapLineItemRowToDomain(&generated.DealLineItem{
			ID:          row.ID,
			DealID:      row.DealID,
			ProductID:   row.ProductID,
			Quantity:    row.Quantity,
			UnitPrice:   row.UnitPrice,
			DiscountPct: row.DiscountPct,
			Total:       row.Total,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}),
	}

	if row.ProductName.Valid {
		item.ProductName = &row.ProductName.String
	}
	if row.ProductSku.Valid {
		item.ProductSKU = &row.ProductSku.String
	}

	return item
}
