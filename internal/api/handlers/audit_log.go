package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/audit_log"
	"github.com/oscar/oscar/pkg/errs"
)

type AuditLogHandler struct {
	repo audit_log.Repository
}

func NewAuditLogHandler(repo audit_log.Repository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

type ListAuditLogsQuery struct {
	EntityType string `query:"entity_type"`
	EntityID   string `query:"entity_id"`
	UserID     string `query:"user_id"`
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
}

func (h *AuditLogHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var query ListAuditLogsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	filter := &audit_log.ListAuditLogsFilter{
		Cursor: query.Cursor,
		Limit:  limit,
	}

	if query.EntityType != "" {
		filter.EntityType = &query.EntityType
	}
	if query.EntityID != "" {
		id, err := uuid.Parse(query.EntityID)
		if err != nil {
			return errs.BadRequest("Invalid entity_id").HTTPError(c)
		}
		filter.EntityID = &id
	}
	if query.UserID != "" {
		id, err := uuid.Parse(query.UserID)
		if err != nil {
			return errs.BadRequest("Invalid user_id").HTTPError(c)
		}
		filter.UserID = &id
	}

	logs, nextCursor, total, err := h.repo.List(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if logs == nil {
		logs = []*audit_log.AuditLog{}
	}

	meta := map[string]interface{}{
		"total":       total,
		"next_cursor": nextCursor,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": logs,
		"meta": meta,
	})
}

func (h *AuditLogHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid audit log ID").HTTPError(c)
	}

	logEntry, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		return errs.NotFound("Audit log entry not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": logEntry,
	})
}
