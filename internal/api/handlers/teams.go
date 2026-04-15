package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/team"
	"github.com/oscar/oscar/pkg/errs"
)

type TeamHandler struct {
	repo team.Repository
}

func NewTeamHandler(repo team.Repository) *TeamHandler {
	return &TeamHandler{repo: repo}
}

type ListTeamsQuery struct {
	IncludeMembers bool `query:"include_members"`
}

func (h *TeamHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	var query ListTeamsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}

	teams, err := h.repo.List(c.Request().Context(), tenantID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	if query.IncludeMembers {
		for _, t := range teams {
			members, err := h.repo.ListMembers(c.Request().Context(), t.ID)
			if err != nil {
				return errs.Internal(err).HTTPError(c)
			}
			_ = members
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": teams,
	})
}

func (h *TeamHandler) Get(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	t, err := h.repo.GetByID(c.Request().Context(), teamID)
	if err != nil {
		return errs.NotFound("Team not found").HTTPError(c)
	}

	members, err := h.repo.ListMembers(c.Request().Context(), teamID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"team":    t,
		"members": members,
	})
}

type CreateTeamRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description"`
}

func (h *TeamHandler) Create(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	var req CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	t, err := h.repo.Create(c.Request().Context(), tenantID, &team.CreateTeamRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, t)
}

type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *TeamHandler) Update(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	var req UpdateTeamRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	t, err := h.repo.Update(c.Request().Context(), teamID, &team.UpdateTeamRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, t)
}

func (h *TeamHandler) Delete(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), teamID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Team deleted",
	})
}

type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	IsLead bool   `json:"is_lead"`
}

func (h *TeamHandler) AddMember(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return errs.BadRequest("Invalid request body").HTTPError(c)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	member, err := h.repo.AddMember(c.Request().Context(), teamID, userID, req.IsLead)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusCreated, member)
}

func (h *TeamHandler) RemoveMember(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	if err := h.repo.RemoveMember(c.Request().Context(), teamID, userID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Member removed from team",
	})
}

func (h *TeamHandler) SetLead(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return errs.BadRequest("Invalid user ID").HTTPError(c)
	}

	if err := h.repo.SetLead(c.Request().Context(), teamID, userID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Team lead set",
	})
}

func (h *TeamHandler) ListMembers(c echo.Context) error {
	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid team ID").HTTPError(c)
	}

	members, err := h.repo.ListMembers(c.Request().Context(), teamID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"members": members,
	})
}
