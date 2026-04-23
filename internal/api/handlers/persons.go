package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/person"
	"github.com/oscar/oscar/pkg/errs"
)

type PersonHandler struct {
	repo person.Repository
}

func NewPersonHandler(repo person.Repository) *PersonHandler {
	return &PersonHandler{repo: repo}
}

type CreatePersonRequest struct {
	Type      string   `json:"type" validate:"required,oneof=lead contact customer"`
	Status    string   `json:"status"`
	FirstName string   `json:"first_name" validate:"required,titlecase"`
	LastName  string   `json:"last_name" validate:"required,titlecase"`
	Email     []string `json:"email" validate:"omitempty,dive,email"`
	Phone     []string `json:"phone" validate:"omitempty,dive,phone"`
	AvatarURL *string  `json:"avatar_url"`
	CompanyID *string  `json:"company_id"`
	OwnerID   *string  `json:"owner_id"`
	Source    *string  `json:"source"`
	Tags      []string `json:"tags"`
}

type UpdatePersonRequest struct {
	Type      *string  `json:"type"`
	Status    *string  `json:"status"`
	FirstName *string  `json:"first_name" validate:"omitempty,titlecase"`
	LastName  *string  `json:"last_name" validate:"omitempty,titlecase"`
	Email     []string `json:"email" validate:"omitempty,dive,email"`
	Phone     []string `json:"phone" validate:"omitempty,dive,phone"`
	AvatarURL *string  `json:"avatar_url"`
	CompanyID *string  `json:"company_id"`
	OwnerID   *string  `json:"owner_id"`
	Source    *string  `json:"source"`
	Score     *int     `json:"score"`
	Tags      []string `json:"tags"`
}

type ConvertPersonRequest struct {
	Type   string `json:"type" validate:"required,oneof=contact customer"`
	Status string `json:"status" validate:"required"`
}

type ListPersonsQuery struct {
	Type    string `query:"type"`
	Status  string `query:"status"`
	OwnerID string `query:"owner_id"`
	Search  string `query:"search"`
	Cursor  string `query:"cursor"`
	Limit   int    `query:"limit"`
}

func (h *PersonHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListPersonsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	filter := &person.ListPersonsFilter{
		Type:   person.PersonType(query.Type),
		Status: person.PersonStatus(query.Status),
		Search: query.Search,
		Cursor: query.Cursor,
		Limit:  query.Limit,
	}

	if query.OwnerID != "" {
		ownerID, err := uuid.Parse(query.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		filter.OwnerID = &ownerID
	}

	persons, nextCursor, total, err := h.repo.List(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": persons,
		"meta": map[string]interface{}{
			"next_cursor": nextCursor,
			"total":       total,
		},
	})
}

func (h *PersonHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreatePersonRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return parseValidationError(err, c)
	}

	createReq := &person.CreatePersonRequest{
		Type:      person.PersonType(req.Type),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Tags:      req.Tags,
	}

	if req.CompanyID != nil {
		id, err := uuid.Parse(*req.CompanyID)
		if err != nil {
			return errs.BadRequest("Invalid company_id").HTTPError(c)
		}
		createReq.CompanyID = &id
	}
	if req.OwnerID != nil {
		id, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		createReq.OwnerID = &id
	}

	p, err := h.repo.Create(c.Request().Context(), tenantID, createReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	p, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Person not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	var req UpdatePersonRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	updateReq := &person.UpdatePersonRequest{}

	if req.Type != nil {
		t := person.PersonType(*req.Type)
		updateReq.Type = &t
	}
	if req.Status != nil {
		s := person.PersonStatus(*req.Status)
		updateReq.Status = &s
	}
	if req.FirstName != nil {
		updateReq.FirstName = req.FirstName
	}
	if req.LastName != nil {
		updateReq.LastName = req.LastName
	}
	if req.Email != nil {
		updateReq.Email = req.Email
	}
	if req.Phone != nil {
		updateReq.Phone = req.Phone
	}
	if req.Score != nil {
		updateReq.Score = req.Score
	}
	if req.Tags != nil {
		updateReq.Tags = req.Tags
	}
	if req.Source != nil {
		s := person.PersonSource(*req.Source)
		updateReq.Source = &s
	}
	if req.CompanyID != nil {
		companyID, err := uuid.Parse(*req.CompanyID)
		if err != nil {
			return errs.BadRequest("Invalid company_id").HTTPError(c)
		}
		updateReq.CompanyID = &companyID
	}
	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		updateReq.OwnerID = &ownerID
	}

	p, err := h.repo.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	p, err := h.repo.SoftDelete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) Convert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	var req ConvertPersonRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	p, err := h.repo.Convert(c.Request().Context(), id, person.PersonType(req.Type), person.PersonStatus(req.Status))
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) AddTag(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	var req struct {
		Tag string `json:"tag" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	p, err := h.repo.AddTag(c.Request().Context(), id, req.Tag)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) RemoveTag(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid person ID").HTTPError(c)
	}

	tag := c.QueryParam("tag")
	if tag == "" {
		return errs.BadRequest("Tag is required").HTTPError(c)
	}

	p, err := h.repo.RemoveTag(c.Request().Context(), id, tag)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PersonHandler) Search(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	query := c.QueryParam("q")
	if query == "" {
		return errs.BadRequest("Search query is required").HTTPError(c)
	}

	limit := 20
	offset := 0

	persons, err := h.repo.Search(c.Request().Context(), tenantID, query, limit, offset)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": persons,
		"meta": map[string]interface{}{
			"total": len(persons),
		},
	})
}

func parseUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func parseValidationError(err error, c echo.Context) error {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return errs.ValidationFailed(errs.Detail{Message: err.Error()}).HTTPError(c)
	}

	details := make([]errs.Detail, 0, len(ve))
	for _, fe := range ve {
		field := toSnakeCase(fe.Field())
		var msg string
		switch fe.Tag() {
		case "required":
			msg = "This field is required"
		case "email":
			msg = "Invalid email format"
		case "phone":
			msg = "Phone must have at least 10 digits"
		case "titlecase":
			msg = "Must be in Title Case (e.g., John Doe)"
		case "oneof":
			msg = "Must be one of: " + fe.Param()
		default:
			msg = "Invalid value for " + fe.Field()
		}
		details = append(details, errs.Detail{Field: field, Message: msg})
	}

	return errs.ValidationFailed(details...).HTTPError(c)
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
