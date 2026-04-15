package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/product"
	"github.com/oscar/oscar/pkg/errs"
)

type ProductHandler struct {
	repo product.Repository
}

func NewProductHandler(repo product.Repository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

type ListProductsQuery struct {
	Offset     int  `query:"offset"`
	Limit      int  `query:"limit"`
	ActiveOnly bool `query:"active_only"`
}

func (h *ProductHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListProductsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	var products []*product.Product
	var total int
	var err error

	if query.ActiveOnly {
		products, total, err = h.repo.ListActive(c.Request().Context(), tenantID, query.Limit, query.Offset)
	} else {
		products, total, err = h.repo.List(c.Request().Context(), tenantID, query.Limit, query.Offset)
	}

	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  products,
		"total": total,
	})
}

func (h *ProductHandler) Get(c echo.Context) error {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid product ID").HTTPError(c)
	}

	p, err := h.repo.GetByID(c.Request().Context(), productID)
	if err != nil {
		return errs.NotFound("Product not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, p)
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description"`
	SKU         *string `json:"sku"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Unit        string  `json:"unit"`
	IsActive    bool    `json:"is_active"`
}

func (h *ProductHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	p, err := h.repo.Create(c.Request().Context(), tenantID, &product.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		Currency:    req.Currency,
		Unit:        req.Unit,
		IsActive:    req.IsActive,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, p)
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

func (h *ProductHandler) Update(c echo.Context) error {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid product ID").HTTPError(c)
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	p, err := h.repo.Update(c.Request().Context(), productID, &product.UpdateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		Currency:    req.Currency,
		Unit:        req.Unit,
		IsActive:    req.IsActive,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, p)
}

func (h *ProductHandler) Delete(c echo.Context) error {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid product ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), productID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Product deleted",
	})
}
