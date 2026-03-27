package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/opencrm/opencrm/internal/domain/deal"
	"github.com/opencrm/opencrm/pkg/errs"
)

type DealHandler struct {
	dealRepo     deal.Repository
	pipelineRepo deal.PipelineRepository
}

func NewDealHandler(dealRepo deal.Repository, pipelineRepo deal.PipelineRepository) *DealHandler {
	return &DealHandler{
		dealRepo:     dealRepo,
		pipelineRepo: pipelineRepo,
	}
}

type CreateDealRequest struct {
	Title             string   `json:"title" validate:"required"`
	Value             float64  `json:"value"`
	Currency          string   `json:"currency"`
	StageID           *string  `json:"stage_id"`
	PipelineID        *string  `json:"pipeline_id"`
	PersonID          *string  `json:"person_id"`
	CompanyID         *string  `json:"company_id"`
	OwnerID           *string  `json:"owner_id"`
	ExpectedCloseDate *string  `json:"expected_close_date"`
	Tags              []string `json:"tags"`
}

type UpdateDealRequest struct {
	Title             *string  `json:"title"`
	Value             *float64 `json:"value"`
	Currency          *string  `json:"currency"`
	StageID           *string  `json:"stage_id"`
	PipelineID        *string  `json:"pipeline_id"`
	PersonID          *string  `json:"person_id"`
	CompanyID         *string  `json:"company_id"`
	OwnerID           *string  `json:"owner_id"`
	ExpectedCloseDate *string  `json:"expected_close_date"`
	Probability       *int     `json:"probability"`
	Tags              []string `json:"tags"`
}

type MoveDealRequest struct {
	StageID     string `json:"stage_id" validate:"required"`
	Probability int    `json:"probability"`
}

type ListDealsQuery struct {
	StageID    string `query:"stage_id"`
	PipelineID string `query:"pipeline_id"`
	OwnerID    string `query:"owner_id"`
	Search     string `query:"search"`
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
}

func (h *DealHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListDealsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	filter := &deal.ListDealsFilter{
		Search: query.Search,
		Cursor: query.Cursor,
		Limit:  query.Limit,
	}

	if query.StageID != "" {
		stageID, err := uuid.Parse(query.StageID)
		if err != nil {
			return errs.BadRequest("Invalid stage_id").HTTPError(c)
		}
		filter.StageID = &stageID
	}
	if query.PipelineID != "" {
		pipelineID, err := uuid.Parse(query.PipelineID)
		if err != nil {
			return errs.BadRequest("Invalid pipeline_id").HTTPError(c)
		}
		filter.PipelineID = &pipelineID
	}
	if query.OwnerID != "" {
		ownerID, err := uuid.Parse(query.OwnerID)
		if err != nil {
			return errs.BadRequest("Invalid owner_id").HTTPError(c)
		}
		filter.OwnerID = &ownerID
	}

	deals, nextCursor, total, err := h.dealRepo.List(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": deals,
		"meta": map[string]interface{}{
			"next_cursor": nextCursor,
			"total":       total,
		},
	})
}

func (h *DealHandler) Kanban(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	pipelineIDStr := c.QueryParam("pipeline_id")
	if pipelineIDStr == "" {
		defaultPipeline, err := h.pipelineRepo.GetDefault(c.Request().Context(), tenantID)
		if err != nil {
			return errs.BadRequest("pipeline_id is required or set a default pipeline").HTTPError(c)
		}
		pipelineIDStr = defaultPipeline.ID.String()
	}

	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		return errs.BadRequest("Invalid pipeline_id").HTTPError(c)
	}

	stages, err := h.pipelineRepo.ListStages(c.Request().Context(), pipelineID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	deals, err := h.dealRepo.ListByStage(c.Request().Context(), tenantID, pipelineID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	stagesMap := make(map[uuid.UUID][]*deal.DealWithStage)
	for _, d := range deals {
		if d.StageID != nil {
			stagesMap[*d.StageID] = append(stagesMap[*d.StageID], d)
		}
	}

	result := make([]map[string]interface{}, len(stages))
	for i, stage := range stages {
		result[i] = map[string]interface{}{
			"id":          stage.ID,
			"name":        stage.Name,
			"position":    stage.Position,
			"probability": stage.Probability,
			"stage_type":  stage.StageType,
			"deals":       stagesMap[stage.ID],
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": result,
	})
}

func (h *DealHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreateDealRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	createReq := &deal.CreateDealRequest{
		Title:    req.Title,
		Value:    req.Value,
		Currency: req.Currency,
		Tags:     req.Tags,
	}

	if req.StageID != nil {
		id, err := uuid.Parse(*req.StageID)
		if err != nil {
			return errs.BadRequest("Invalid stage_id").HTTPError(c)
		}
		createReq.StageID = &id
	}
	if req.PipelineID != nil {
		id, err := uuid.Parse(*req.PipelineID)
		if err != nil {
			return errs.BadRequest("Invalid pipeline_id").HTTPError(c)
		}
		createReq.PipelineID = &id
	}
	if req.PersonID != nil {
		id, err := uuid.Parse(*req.PersonID)
		if err != nil {
			return errs.BadRequest("Invalid person_id").HTTPError(c)
		}
		createReq.PersonID = &id
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
	if req.ExpectedCloseDate != nil {
		t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate)
		if err != nil {
			return errs.BadRequest("Invalid expected_close_date format. Use YYYY-MM-DD").HTTPError(c)
		}
		createReq.ExpectedCloseDate = &t
	}

	d, err := h.dealRepo.Create(c.Request().Context(), tenantID, createReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	d, err := h.dealRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Deal not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	var req UpdateDealRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	updateReq := &deal.UpdateDealRequest{
		Title:       req.Title,
		Value:       req.Value,
		Currency:    req.Currency,
		Probability: req.Probability,
		Tags:        req.Tags,
	}

	if req.StageID != nil {
		stageID, err := uuid.Parse(*req.StageID)
		if err != nil {
			return errs.BadRequest("Invalid stage_id").HTTPError(c)
		}
		updateReq.StageID = &stageID
	}
	if req.PipelineID != nil {
		pipelineID, err := uuid.Parse(*req.PipelineID)
		if err != nil {
			return errs.BadRequest("Invalid pipeline_id").HTTPError(c)
		}
		updateReq.PipelineID = &pipelineID
	}
	if req.PersonID != nil {
		personID, err := uuid.Parse(*req.PersonID)
		if err != nil {
			return errs.BadRequest("Invalid person_id").HTTPError(c)
		}
		updateReq.PersonID = &personID
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
	if req.ExpectedCloseDate != nil {
		t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate)
		if err != nil {
			return errs.BadRequest("Invalid expected_close_date format. Use YYYY-MM-DD").HTTPError(c)
		}
		updateReq.ExpectedCloseDate = &t
	}

	d, err := h.dealRepo.Update(c.Request().Context(), id, updateReq)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	d, err := h.dealRepo.SoftDelete(c.Request().Context(), id)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) MoveStage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	var req MoveDealRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}
	if err := c.Validate(&req); err != nil {
		return errs.ValidationFailed().HTTPError(c)
	}

	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		return errs.BadRequest("Invalid stage_id").HTTPError(c)
	}

	stage, err := h.pipelineRepo.GetStageByID(c.Request().Context(), stageID)
	if err != nil {
		return errs.NotFound("Stage not found").HTTPError(c)
	}

	probability := req.Probability
	if probability == 0 {
		probability = stage.Probability
	}

	var closedAt *time.Time
	if stage.StageType == "won" || stage.StageType == "lost" {
		now := time.Now()
		closedAt = &now
	}

	d, err := h.dealRepo.MoveToStage(c.Request().Context(), id, stageID, stage.PipelineID, probability, closedAt)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) Win(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&req)

	stageID, err := uuid.Parse(c.QueryParam("stage_id"))
	if err != nil {
		return errs.BadRequest("stage_id is required").HTTPError(c)
	}

	d, err := h.dealRepo.CloseAsWon(c.Request().Context(), id, stageID, req.Reason)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}

func (h *DealHandler) Lose(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid deal ID").HTTPError(c)
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&req)

	stageID, err := uuid.Parse(c.QueryParam("stage_id"))
	if err != nil {
		return errs.BadRequest("stage_id is required").HTTPError(c)
	}

	d, err := h.dealRepo.CloseAsLost(c.Request().Context(), id, stageID, req.Reason)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": d,
	})
}
