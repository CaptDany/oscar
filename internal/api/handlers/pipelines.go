package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/opencrm/opencrm/internal/domain/deal"
	"github.com/opencrm/opencrm/pkg/errs"
)

type PipelineHandler struct {
	repo deal.PipelineRepository
}

func NewPipelineHandler(repo deal.PipelineRepository) *PipelineHandler {
	return &PipelineHandler{repo: repo}
}

type CreatePipelineRequest struct {
	Name     string `json:"name" validate:"required"`
	IsDefault bool   `json:"is_default"`
	Currency  string `json:"currency"`
}

type UpdatePipelineRequest struct {
	Name      *string `json:"name"`
	IsDefault *bool   `json:"is_default"`
	Currency  *string `json:"currency"`
}

type CreateStageRequest struct {
	Name        string `json:"name" validate:"required"`
	Position    int    `json:"position"`
	Probability int    `json:"probability"`
	StageType   string `json:"stage_type"`
}

type UpdateStageRequest struct {
	Name        *string `json:"name"`
	Probability *int    `json:"probability"`
	StageType   *string `json:"stage_type"`
}

type ReorderStagesRequest struct {
	StageIDs []string `json:"stage_ids" validate:"required"`
}

func (h *PipelineHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	pipelines, err := h.repo.List(c.Request().Context(), tenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	for _, p := range pipelines {
		stages, err := h.repo.ListStages(c.Request().Context(), p.ID)
		if err != nil {
			return errs.Internal(err).HTTPError(c)
		}
		p.Stages = stages
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": pipelines,
	})
}

func (h *PipelineHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreatePipelineRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	p, err := h.repo.Create(c.Request().Context(), tenantID, req.Name, req.IsDefault, currency)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": p,
	})
}

func (h *PipelineHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	p, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Pipeline not found").HTTPError(c)
	}

	stages, err := h.repo.ListStages(c.Request().Context(), p.ID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}
	p.Stages = stages

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PipelineHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	var req UpdatePipelineRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	p, err := h.repo.Update(c.Request().Context(), id, req.Name, req.IsDefault, req.Currency)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *PipelineHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *PipelineHandler) CreateStage(c echo.Context) error {
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	var req CreateStageRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	stageType := req.StageType
	if stageType == "" {
		stageType = "open"
	}

	s, err := h.repo.CreateStage(c.Request().Context(), pipelineID, req.Name, req.Position, req.Probability, stageType)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": s,
	})
}

func (h *PipelineHandler) ListStages(c echo.Context) error {
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	stages, err := h.repo.ListStages(c.Request().Context(), pipelineID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": stages,
	})
}

func (h *PipelineHandler) UpdateStage(c echo.Context) error {
	stageID, err := uuid.Parse(c.Param("stage_id"))
	if err != nil {
		return errs.BadRequest("Invalid stage ID").HTTPError(c)
	}

	var req UpdateStageRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	s, err := h.repo.UpdateStage(c.Request().Context(), stageID, req.Name, req.Probability, req.StageType)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": s,
	})
}

func (h *PipelineHandler) ReorderStages(c echo.Context) error {
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid pipeline ID").HTTPError(c)
	}

	var req ReorderStagesRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	stageIDs := make([]uuid.UUID, len(req.StageIDs))
	for i, s := range req.StageIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return errs.BadRequest("Invalid stage ID at index %d", i).HTTPError(c)
		}
		stageIDs[i] = id
	}

	if err := h.repo.ReorderStages(c.Request().Context(), pipelineID, stageIDs); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	stages, err := h.repo.ListStages(c.Request().Context(), pipelineID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": stages,
	})
}

func (h *PipelineHandler) DeleteStage(c echo.Context) error {
	stageID, err := uuid.Parse(c.Param("stage_id"))
	if err != nil {
		return errs.BadRequest("Invalid stage ID").HTTPError(c)
	}

	if err := h.repo.DeleteStage(c.Request().Context(), stageID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.NoContent(http.StatusNoContent)
}
