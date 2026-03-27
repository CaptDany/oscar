-- name: CreateProduct :one
INSERT INTO products (tenant_id, name, description, sku, price, currency, unit, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: UpdateProduct :one
UPDATE products
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    sku = COALESCE($4, sku),
    price = COALESCE($5, price),
    currency = COALESCE($6, currency),
    unit = COALESCE($7, unit),
    is_active = COALESCE($8, is_active)
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = $1 RETURNING *;

-- name: ListProducts :many
SELECT * FROM products 
WHERE tenant_id = $1
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: ListActiveProducts :many
SELECT * FROM products 
WHERE tenant_id = $1 AND is_active = true
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: CountProducts :one
SELECT COUNT(*) FROM products WHERE tenant_id = $1;

-- name: CreateDealLineItem :one
INSERT INTO deal_line_items (deal_id, product_id, quantity, unit_price, discount_pct, total)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDealLineItemByID :one
SELECT * FROM deal_line_items WHERE id = $1;

-- name: ListDealLineItems :many
SELECT dli.*, p.name as product_name, p.sku as product_sku
FROM deal_line_items dli
LEFT JOIN products p ON dli.product_id = p.id
WHERE dli.deal_id = $1
ORDER BY dli.created_at ASC;

-- name: UpdateDealLineItem :one
UPDATE deal_line_items
SET 
    product_id = COALESCE($2, product_id),
    quantity = COALESCE($3, quantity),
    unit_price = COALESCE($4, unit_price),
    discount_pct = COALESCE($5, discount_pct),
    total = COALESCE($6, total)
WHERE id = $1
RETURNING *;

-- name: DeleteDealLineItem :one
DELETE FROM deal_line_items WHERE id = $1 RETURNING *;

-- name: GetDealTotal :one
SELECT COALESCE(SUM(total), 0) as deal_total
FROM deal_line_items
WHERE deal_id = $1;
