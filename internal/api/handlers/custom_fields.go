package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/custom_field"
	"github.com/oscar/oscar/pkg/errs"
)

type CustomFieldHandler struct {
	repo custom_field.Repository
}

func NewCustomFieldHandler(repo custom_field.Repository) *CustomFieldHandler {
	return &CustomFieldHandler{repo: repo}
}

type CreateCustomFieldRequest struct {
	EntityType     string      `json:"entity_type" validate:"required,oneof=person company deal"`
	FieldKey       string      `json:"field_key" validate:"required,min=1,max=50"`
	Label          string      `json:"label" validate:"required,min=1,max=100"`
	FieldType      string      `json:"field_type" validate:"required,oneof=text number date select multi_select boolean url currency"`
	Options        interface{} `json:"options,omitempty"`
	IsRequired     bool        `json:"is_required"`
	ShowInList     bool        `json:"show_in_list"`
	ShowInCard     bool        `json:"show_in_card"`
	Position       int         `json:"position"`
	RoleVisibility []string    `json:"role_visibility,omitempty"`
}

type UpdateCustomFieldRequest struct {
	Label          *string      `json:"label"`
	FieldType      *string      `json:"field_type"`
	Options        interface{}  `json:"options,omitempty"`
	IsRequired     *bool        `json:"is_required"`
	ShowInList     *bool        `json:"show_in_list"`
	ShowInCard     *bool        `json:"show_in_card"`
	Position       *int         `json:"position"`
	RoleVisibility []string     `json:"role_visibility,omitempty"`
}

type ListCustomFieldsQuery struct {
	EntityType string `query:"entity_type"`
}

type ReorderCustomFieldsRequest struct {
	FieldIDs []string `json:"field_ids" validate:"required,min=1"`
}

func (h *CustomFieldHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListCustomFieldsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}

	var fields []*custom_field.CustomFieldDefinition
	var err error

	if query.EntityType != "" {
		entityType := custom_field.EntityType(query.EntityType)
		if entityType != custom_field.EntityTypePerson &&
			entityType != custom_field.EntityTypeCompany &&
			entityType != custom_field.EntityTypeDeal {
			return errs.BadRequest("Invalid entity_type: must be person, company, or deal").HTTPError(c)
		}
		fields, err = h.repo.ListByEntity(c.Request().Context(), tenantID, entityType)
	} else {
		fields, err = h.repo.ListAll(c.Request().Context(), tenantID)
	}

	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if fields == nil {
		fields = []*custom_field.CustomFieldDefinition{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": fields,
	})
}

func (h *CustomFieldHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreateCustomFieldRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return parseValidationError(err, c)
	}

	createReq := &custom_field.CreateCustomFieldRequest{
		EntityType:     custom_field.EntityType(req.EntityType),
		FieldKey:       req.FieldKey,
		Label:          req.Label,
		FieldType:      custom_field.FieldType(req.FieldType),
		Options:        req.Options,
		IsRequired:     req.IsRequired,
		ShowInList:     req.ShowInList,
		ShowInCard:     req.ShowInCard,
		Position:       req.Position,
		RoleVisibility: req.RoleVisibility,
	}

	field, err := h.repo.Create(c.Request().Context(), tenantID, createReq)
	if err != nil {
		if isDuplicateKeyError(err) {
			return errs.Conflict("A custom field with key '" + req.FieldKey + "' already exists for this entity type").HTTPError(c)
		}
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": field,
	})
}

func (h *CustomFieldHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid custom field ID").HTTPError(c)
	}

	field, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Custom field not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": field,
	})
}

func (h *CustomFieldHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid custom field ID").HTTPError(c)
	}

	var req UpdateCustomFieldRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	updateReq := &custom_field.UpdateCustomFieldRequest{}

	if req.Label != nil {
		updateReq.Label = req.Label
	}
	if req.FieldType != nil {
		ft := custom_field.FieldType(*req.FieldType)
		updateReq.FieldType = &ft
	}
	if req.Options != nil {
		updateReq.Options = req.Options
	}
	if req.IsRequired != nil {
		updateReq.IsRequired = req.IsRequired
	}
	if req.ShowInList != nil {
		updateReq.ShowInList = req.ShowInList
	}
	if req.ShowInCard != nil {
		updateReq.ShowInCard = req.ShowInCard
	}
	if req.Position != nil {
		updateReq.Position = req.Position
	}
	if req.RoleVisibility != nil {
		updateReq.RoleVisibility = req.RoleVisibility
	}

	field, err := h.repo.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		if isDuplicateKeyError(err) {
			return errs.Conflict("A custom field with this key already exists").HTTPError(c)
		}
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": field,
	})
}

func (h *CustomFieldHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid custom field ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "Custom field deleted"},
	})
}

func (h *CustomFieldHandler) Reorder(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req ReorderCustomFieldsRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	fieldIDs := make([]uuid.UUID, len(req.FieldIDs))
	for i, idStr := range req.FieldIDs {
		fieldID, err := uuid.Parse(idStr)
		if err != nil {
			return errs.BadRequest("Invalid field ID: " + idStr).HTTPError(c)
		}
		fieldIDs[i] = fieldID
	}

	if err := h.repo.Reorder(c.Request().Context(), tenantID, fieldIDs); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "Fields reordered"},
	})
}

func isDuplicateKeyError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return true
	}
	return false
}
