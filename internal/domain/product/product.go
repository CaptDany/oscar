package product

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	SKU         *string   `json:"sku"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Unit        string    `json:"unit"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DealLineItem struct {
	ID          uuid.UUID  `json:"id"`
	DealID      uuid.UUID  `json:"deal_id"`
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    float64    `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	DiscountPct float64    `json:"discount_pct"`
	Total       float64    `json:"total"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type DealLineItemWithProduct struct {
	DealLineItem
	ProductName *string `json:"product_name"`
	ProductSKU  *string `json:"product_sku"`
}

type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=255"`
	Description *string  `json:"description"`
	SKU         *string  `json:"sku"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Unit        string   `json:"unit"`
	IsActive    bool     `json:"is_active"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	SKU         *string  `json:"sku"`
	Price       *float64 `json:"price"`
	Currency    *string  `json:"currency"`
	Unit        *string  `json:"unit"`
	IsActive    *bool    `json:"is_active"`
}

type CreateLineItemRequest struct {
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    float64    `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64    `json:"unit_price" validate:"required,gte=0"`
	DiscountPct float64    `json:"discount_pct" validate:"gte=0,lte=100"`
}

type UpdateLineItemRequest struct {
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    *float64   `json:"quantity"`
	UnitPrice   *float64   `json:"unit_price"`
	DiscountPct *float64   `json:"discount_pct"`
}

type Repository interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *CreateProductRequest) (*Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Product, int, error)
	ListActive(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Product, int, error)
}

type LineItemRepository interface {
	Create(ctx context.Context, dealID uuid.UUID, req *CreateLineItemRequest) (*DealLineItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*DealLineItem, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateLineItemRequest) (*DealLineItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByDeal(ctx context.Context, dealID uuid.UUID) ([]*DealLineItemWithProduct, error)
	GetDealTotal(ctx context.Context, dealID uuid.UUID) (float64, error)
}
