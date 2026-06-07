package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/deal"
	"github.com/oscar/oscar/internal/domain/product"
	"github.com/oscar/oscar/pkg/errs"
)

type LineItemHandler struct {
	lineItemRepo product.LineItemRepository
	dealRepo     deal.Repository
}

func NewLineItemHandler(lineItemRepo product.LineItemRepository, dealRepo deal.Repository) *LineItemHandler {
	return &LineItemHandler{
		lineItemRepo: lineItemRepo,
		dealRepo:     dealRepo,
	}
}

func (h *LineItemHandler) ListByDeal(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	dealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	d, err := h.dealRepo.GetByID(c.Request().Context(), dealID)
	if err != nil {
		return errs.NotFound("Deal not found").HTTPError(c)
	}
	if d.TenantID != tenantID {
		return errs.NotFound("Deal not found").HTTPError(c)
	}

	items, err := h.lineItemRepo.ListByDeal(c.Request().Context(), dealID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

type createLineItemRequest struct {
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    float64    `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	DiscountPct float64    `json:"discount_pct"`
}

func (h *LineItemHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	dealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	d, err := h.dealRepo.GetByID(c.Request().Context(), dealID)
	if err != nil {
		return errs.NotFound("Deal not found").HTTPError(c)
	}
	if d.TenantID != tenantID {
		return errs.NotFound("Deal not found").HTTPError(c)
	}

	var req createLineItemRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if req.Quantity <= 0 {
		return errs.BadRequest("Quantity must be greater than 0").HTTPError(c)
	}
	if req.UnitPrice < 0 {
		return errs.BadRequest("Unit price must be non-negative").HTTPError(c)
	}
	if req.DiscountPct < 0 || req.DiscountPct > 100 {
		return errs.BadRequest("Discount must be between 0 and 100").HTTPError(c)
	}

	item, err := h.lineItemRepo.Create(c.Request().Context(), dealID, &product.CreateLineItemRequest{
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, item)
}

type updateLineItemRequest struct {
	ProductID   *uuid.UUID `json:"product_id"`
	Quantity    *float64   `json:"quantity"`
	UnitPrice   *float64   `json:"unit_price"`
	DiscountPct *float64   `json:"discount_pct"`
}

func (h *LineItemHandler) Update(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	dealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	lineItemID, err := uuid.Parse(c.Param("line_item_id"))
	if err != nil {
		return errs.BadRequest("Invalid line item ID").HTTPError(c)
	}

	d, err := h.dealRepo.GetByID(c.Request().Context(), dealID)
	if err != nil {
		return errs.NotFound("Deal not found").HTTPError(c)
	}
	if d.TenantID != tenantID {
		return errs.NotFound("Deal not found").HTTPError(c)
	}

	var req updateLineItemRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	if req.Quantity != nil && *req.Quantity <= 0 {
		return errs.BadRequest("Quantity must be greater than 0").HTTPError(c)
	}
	if req.UnitPrice != nil && *req.UnitPrice < 0 {
		return errs.BadRequest("Unit price must be non-negative").HTTPError(c)
	}
	if req.DiscountPct != nil && (*req.DiscountPct < 0 || *req.DiscountPct > 100) {
		return errs.BadRequest("Discount must be between 0 and 100").HTTPError(c)
	}

	item, err := h.lineItemRepo.Update(c.Request().Context(), lineItemID, &product.UpdateLineItemRequest{
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, item)
}

func (h *LineItemHandler) Delete(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	dealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	lineItemID, err := uuid.Parse(c.Param("line_item_id"))
	if err != nil {
		return errs.BadRequest("Invalid line item ID").HTTPError(c)
	}

	d, err := h.dealRepo.GetByID(c.Request().Context(), dealID)
	if err != nil {
		return errs.NotFound("Deal not found").HTTPError(c)
	}
	if d.TenantID != tenantID {
		return errs.NotFound("Deal not found").HTTPError(c)
	}

	if err := h.lineItemRepo.Delete(c.Request().Context(), lineItemID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Line item deleted",
	})
}
