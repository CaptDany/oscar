package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/notification"
	"github.com/oscar/oscar/pkg/errs"
)

type NotificationHandler struct {
	repo notification.Repository
}

func NewNotificationHandler(repo notification.Repository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

type ListNotificationsQuery struct {
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
	UnreadOnly bool   `query:"unread_only"`
}

func (h *NotificationHandler) List(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var query ListNotificationsQuery
	if err := c.Bind(&query); err != nil {
		return errs.BadRequest("Invalid query parameters").HTTPError(c)
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	filter := &notification.ListNotificationsFilter{
		Cursor:     query.Cursor,
		Limit:      query.Limit,
		UnreadOnly: query.UnreadOnly,
	}

	notifications, nextCursor, total, err := h.repo.List(c.Request().Context(), tenantID, userID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        notifications,
		"next_cursor": nextCursor,
		"total":       total,
	})
}

func (h *NotificationHandler) Get(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid notification ID").HTTPError(c)
	}

	n, err := h.repo.GetByID(c.Request().Context(), notificationID)
	if err != nil {
		return errs.NotFound("Notification not found").HTTPError(c)
	}

	if n.UserID != userID {
		return errs.NotFound("Notification not found").HTTPError(c)
	}

	return c.JSON(http.StatusOK, n)
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid notification ID").HTTPError(c)
	}

	n, err := h.repo.MarkAsRead(c.Request().Context(), notificationID, userID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, n)
}

func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	count, err := h.repo.MarkAllAsRead(c.Request().Context(), tenantID, userID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"marked_count": count,
	})
}

func (h *NotificationHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.BadRequest("Invalid notification ID").HTTPError(c)
	}

	if err := h.repo.Delete(c.Request().Context(), notificationID, userID); err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Notification deleted",
	})
}

func (h *NotificationHandler) CountUnread(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	count, err := h.repo.CountUnread(c.Request().Context(), tenantID, userID)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"unread_count": count,
	})
}
