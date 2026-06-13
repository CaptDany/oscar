package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/search"
	"github.com/oscar/oscar/pkg/errs"
)

type SearchHandler struct {
	repo search.Repository
}

func NewSearchHandler(repo search.Repository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

func (h *SearchHandler) GlobalSearch(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	query := c.QueryParam("q")
	if query == "" {
		return errs.BadRequest("Search query is required").HTTPError(c)
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	filter := &search.SearchFilter{
		Query: query,
		Limit: limit,
	}

	if typeStr := c.QueryParam("type"); typeStr != "" {
		filter.Types = strings.Split(typeStr, ",")
		for _, t := range filter.Types {
			if t != "person" && t != "company" && t != "deal" {
				return errs.BadRequest("Invalid entity type: %s", t).HTTPError(c)
			}
		}
	}

	results, err := h.repo.Search(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": results,
		"meta": map[string]interface{}{
			"total": len(results),
			"query": query,
		},
	})
}


