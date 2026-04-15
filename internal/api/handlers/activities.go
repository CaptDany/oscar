package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/activity"
	"github.com/oscar/oscar/pkg/errs"
	"github.com/oscar/oscar/pkg/validator"
)

type ActivityHandler struct {
	repo      activity.Repository
	assocRepo activity.AssociationRepository
}

func NewActivityHandler(repo activity.Repository, assocRepo activity.AssociationRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo, assocRepo: assocRepo}
}

type CreateActivityRequest struct {
	Type            string  `json:"type" validate:"required,oneof=note call email meeting task whatsapp sms"`
	Subject         string  `json:"subject"`
	Title           string  `json:"title" validate:"required"` // Alias for subject, made required
	Body            *string `json:"body"`
	Outcome         *string `json:"outcome"`
	Direction       *string `json:"direction"`
	Status          string  `json:"status"`
	DueAt           *string `json:"due_at"`
	DurationSeconds *int    `json:"duration_seconds"`
	OwnerID         *string `json:"owner_id"`
	EntityType      *string `json:"entity_type"`
	EntityID        *string `json:"entity_id"`
}

type UpdateActivityRequest struct {
	Type            *string `json:"type"`
	Subject         *string `json:"subject"`
	Title           *string `json:"title"` // Alias for subject
	Body            *string `json:"body"`
	Outcome         *string `json:"outcome"`
	Direction       *string `json:"direction"`
	Status          *string `json:"status"`
	DueAt           *string `json:"due_at"`
	CompletedAt     *string `json:"completed_at"`
	DurationSeconds *int    `json:"duration_seconds"`
	OwnerID         *string `json:"owner_id"`
}

type ListActivitiesQuery struct {
	Type       string `query:"type"`
	Status     string `query:"status"`
	OwnerID    string `query:"owner_id"`
	EntityType string `query:"entity_type"`
	EntityID   string `query:"entity_id"`
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
}

func (h *ActivityHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListActivitiesQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	filter := &activity.ListActivitiesFilter{
		Type:   activity.ActivityType(query.Type),
		Status: activity.ActivityStatus(query.Status),
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
	if query.EntityType != "" {
		filter.EntityType = activity.EntityType(query.EntityType)
	}
	if query.EntityID != "" {
		entityID, err := uuid.Parse(query.EntityID)
		if err != nil {
			return errs.BadRequest("Invalid entity_id").HTTPError(c)
		}
		filter.EntityID = &entityID
	}

	activities, nextCursor, total, err := h.repo.List(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": activities,
		"meta": map[string]interface{}{
			"next_cursor": nextCursor,
			"total":       total,
		},
	})
}

func (h *ActivityHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var req CreateActivityRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_FAILED",
				"message": "Validation failed",
				"details": validationErrors,
			},
		})
	}

	subject := req.Subject
	if subject == "" {
		subject = req.Title
	}

	createReq := &activity.CreateActivityRequest{
		Type:            activity.ActivityType(req.Type),
		Subject:         subject,
		Body:            req.Body,
		Outcome:         req.Outcome,
		Status:          activity.ActivityStatus(req.Status),
		DurationSeconds: req.DurationSeconds,
		CreatedBy:       &userID,
	}

	if req.Direction != nil {
		dir := activity.ActivityDirection(*req.Direction)
		createReq.Direction = &dir
	}
	if req.DueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			return errs.BadRequest("Invalid due_at format. Use RFC3339").HTTPError(c)
		}
		createReq.DueAt = &t
	}
	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		createReq.OwnerID = &ownerID
	}
	if req.EntityType != nil && req.EntityID != nil {
		entityType := activity.EntityType(*req.EntityType)
		entityID, err := uuid.Parse(*req.EntityID)
		if err != nil {
			return errs.BadRequest("Invalid entity_id").HTTPError(c)
		}
		createReq.EntityType = &entityType
		createReq.EntityID = &entityID
	}

	a, err := h.repo.Create(c.Request().Context(), tenantID, createReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if createReq.EntityType != nil && createReq.EntityID != nil {
		_, err = h.assocRepo.Create(c.Request().Context(), a.ID, *createReq.EntityType, *createReq.EntityID)
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": a,
	})
}

func (h *ActivityHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid activity ID").HTTPError(c)
	}

	a, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Activity not found").HTTPError(c)
	}

	associations, _ := h.assocRepo.ListByActivity(c.Request().Context(), id)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"activity":     a,
			"associations": associations,
		},
	})
}

func (h *ActivityHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid activity ID").HTTPError(c)
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Activity not found").HTTPError(c)
	}

	if err := h.checkActivityPermission(c, existing); err != nil {
		return err
	}

	var req UpdateActivityRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	updateReq := &activity.UpdateActivityRequest{}

	if req.Type != nil {
		t := activity.ActivityType(*req.Type)
		updateReq.Type = &t
	}
	if req.Subject != nil {
		updateReq.Subject = req.Subject
	} else if req.Title != nil && *req.Title != "" {
		updateReq.Subject = req.Title
	}
	if req.Body != nil {
		updateReq.Body = req.Body
	}
	if req.Outcome != nil {
		updateReq.Outcome = req.Outcome
	}
	if req.Direction != nil {
		dir := activity.ActivityDirection(*req.Direction)
		updateReq.Direction = &dir
	}
	if req.Status != nil {
		s := activity.ActivityStatus(*req.Status)
		updateReq.Status = &s
	}
	if req.DueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			return errs.BadRequest("Invalid due_at format").HTTPError(c)
		}
		updateReq.DueAt = &t
	}
	if req.CompletedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.CompletedAt)
		if err != nil {
			return errs.BadRequest("Invalid completed_at format").HTTPError(c)
		}
		updateReq.CompletedAt = &t
	}
	if req.DurationSeconds != nil {
		updateReq.DurationSeconds = req.DurationSeconds
	}
	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		updateReq.OwnerID = &ownerID
	}

	a, err := h.repo.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": a,
	})
}

func (h *ActivityHandler) Complete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid activity ID").HTTPError(c)
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Activity not found").HTTPError(c)
	}
	if err := h.checkActivityPermission(c, existing); err != nil {
		return err
	}

	a, err := h.repo.Complete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": a,
	})
}

func (h *ActivityHandler) Uncomplete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid activity ID").HTTPError(c)
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Activity not found").HTTPError(c)
	}
	if err := h.checkActivityPermission(c, existing); err != nil {
		return err
	}

	a, err := h.repo.Uncomplete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": a,
	})
}

func (h *ActivityHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid activity ID").HTTPError(c)
	}

	existing, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Activity not found").HTTPError(c)
	}
	if err := h.checkActivityPermission(c, existing); err != nil {
		return err
	}

	a, err := h.repo.SoftDelete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": a,
	})
}

func (h *ActivityHandler) Timeline(c echo.Context) error {
	entityType := c.QueryParam("entity_type")
	entityIDStr := c.QueryParam("entity_id")

	if entityType == "" || entityIDStr == "" {
		return errs.BadRequest("entity_type and entity_id are required").HTTPError(c)
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return errs.BadRequest("Invalid entity_id").HTTPError(c)
	}

	limit := 20
	offset := 0

	entries, err := h.assocRepo.ListTimeline(c.Request().Context(), activity.EntityType(entityType), entityID, limit, offset)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": entries,
	})
}

func (h *ActivityHandler) checkActivityPermission(c echo.Context, activity *activity.Activity) error {
	userID := c.Get("user_id").(uuid.UUID)
	roles := c.Get("roles").([]string)

	if activity.CreatedBy != nil && *activity.CreatedBy == userID {
		return nil
	}

	for _, role := range roles {
		if role == "Owner" || role == "Admin" {
			return nil
		}
	}

	return errs.Forbidden("You do not have permission to modify this activity").HTTPError(c)
}
