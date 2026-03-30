package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/company"
	"github.com/oscar/oscar/pkg/errs"
)

type CompanyHandler struct {
	repo company.Repository
}

func NewCompanyHandler(repo company.Repository) *CompanyHandler {
	return &CompanyHandler{repo: repo}
}

type CreateCompanyRequest struct {
	Name          string   `json:"name" validate:"required"`
	Domain        *string  `json:"domain"`
	Industry      *string  `json:"industry"`
	Size          *string  `json:"size"`
	AnnualRevenue *float64 `json:"annual_revenue"`
	Website       *string  `json:"website"`
	OwnerID       *string  `json:"owner_id"`
	Tags          []string `json:"tags"`
}

type UpdateCompanyRequest struct {
	Name          *string  `json:"name"`
	Domain        *string  `json:"domain"`
	Industry      *string  `json:"industry"`
	Size          *string  `json:"size"`
	AnnualRevenue *float64 `json:"annual_revenue"`
	Website       *string  `json:"website"`
	OwnerID       *string  `json:"owner_id"`
	Tags          []string `json:"tags"`
}

type ListCompaniesQuery struct {
	OwnerID      string `query:"owner_id"`
	Search       string `query:"search"`
	Cursor       string `query:"cursor"`
	Limit        int    `query:"limit"`
	IncludeTotal bool   `query:"include_total"`
}

func (h *CompanyHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListCompaniesQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	filter := &company.ListCompaniesFilter{
		Search:       query.Search,
		Cursor:       query.Cursor,
		Limit:        query.Limit,
		IncludeTotal: query.IncludeTotal,
	}

	if query.OwnerID != "" {
		ownerID, err := uuid.Parse(query.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		filter.OwnerID = &ownerID
	}

	if query.Cursor != "" {
		parts := strings.SplitN(query.Cursor, ":", 2)
		if len(parts) != 2 {
			return errs.BadRequest("Invalid cursor format").HTTPError(c)
		}
		cursorAfter, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			return errs.BadRequest("Invalid cursor timestamp").HTTPError(c)
		}
		cursorID, err := uuid.Parse(parts[1])
		if err != nil {
			return errs.BadRequest("Invalid cursor ID").HTTPError(c)
		}
		filter.CursorAfter = &cursorAfter
		filter.CursorID = &cursorID
	}

	companies, nextCursor, total, err := h.repo.List(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	meta := map[string]interface{}{
		"next_cursor": nextCursor,
	}
	if query.IncludeTotal {
		meta["total"] = total
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": companies,
		"meta": meta,
	})
}

func (h *CompanyHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreateCompanyRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	createReq := &company.CreateCompanyRequest{
		Name:          req.Name,
		Domain:        req.Domain,
		Industry:      req.Industry,
		AnnualRevenue: req.AnnualRevenue,
		Website:       req.Website,
		Tags:          req.Tags,
	}

	if req.OwnerID != nil {
		id, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		createReq.OwnerID = &id
	}

	comp, err := h.repo.Create(c.Request().Context(), tenantID, createReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": comp,
	})
}

func (h *CompanyHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid company ID").HTTPError(c)
	}

	comp, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Company not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": comp,
	})
}

func (h *CompanyHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid company ID").HTTPError(c)
	}

	var req UpdateCompanyRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	updateReq := &company.UpdateCompanyRequest{
		Name:          req.Name,
		Domain:        req.Domain,
		Industry:      req.Industry,
		AnnualRevenue: req.AnnualRevenue,
		Website:       req.Website,
		Tags:          req.Tags,
	}

	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		updateReq.OwnerID = &ownerID
	}

	comp, err := h.repo.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": comp,
	})
}

func (h *CompanyHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid company ID").HTTPError(c)
	}

	comp, err := h.repo.SoftDelete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": comp,
	})
}
