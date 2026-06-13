package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/internal/domain/activity"
	"github.com/oscar/oscar/pkg/errs"
)

type ReportsHandler struct {
	activityRepo activity.Repository
}

func NewReportsHandler(activityRepo activity.Repository) *ReportsHandler {
	return &ReportsHandler{activityRepo: activityRepo}
}

func (h *ReportsHandler) ActivityReport(c echo.Context) error {
	tenantID := c.Get("tenant_id").(uuid.UUID)

	now := time.Now()
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	var startAt, endAt time.Time
	if startStr != "" {
		var err error
		startAt, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return errs.BadRequest("Invalid start date, use YYYY-MM-DD").HTTPError(c)
		}
	} else {
		startAt = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endStr != "" {
		var err error
		endAt, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return errs.BadRequest("Invalid end date, use YYYY-MM-DD").HTTPError(c)
		}
		endAt = endAt.Add(24 * time.Hour)
	} else {
		endAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
	}

	filter := &activity.ActivityReportFilter{
		StartAt: startAt,
		EndAt:   endAt,
	}

	if userIDsStr := c.QueryParam("user_ids"); userIDsStr != "" {
		parts := strings.Split(userIDsStr, ",")
		for _, p := range parts {
			id, err := uuid.Parse(strings.TrimSpace(p))
			if err != nil {
				return errs.BadRequest("Invalid user_id: %s", p).HTTPError(c)
			}
			filter.UserIDs = append(filter.UserIDs, id)
		}
	}

	if typesStr := c.QueryParam("types"); typesStr != "" {
		parts := strings.Split(typesStr, ",")
		for _, p := range parts {
			t := activity.ActivityType(strings.TrimSpace(p))
			switch t {
			case activity.ActivityTypeCall, activity.ActivityTypeEmail, activity.ActivityTypeMeeting,
				activity.ActivityTypeNote, activity.ActivityTypeTask, activity.ActivityTypeWhatsapp,
				activity.ActivityTypeSMS:
				filter.Types = append(filter.Types, t)
			default:
				return errs.BadRequest("Invalid activity type: %s", p).HTTPError(c)
			}
		}
	}

	report, err := h.activityRepo.GetActivityReport(c.Request().Context(), tenantID, filter)
	if err != nil {
		return errs.Internal(err).HTTPError(c)
	}

	format := c.QueryParam("format")
	if format == "csv" {
		return h.exportActivityCSV(c, report)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": report,
	})
}

func (h *ReportsHandler) exportActivityCSV(c echo.Context, report *activity.ActivityReport) error {
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=activity-report.csv")

	var sb strings.Builder
	sb.WriteString("section,label,count\n")

	sb.WriteString("overall,total,")
	sb.WriteString(itoa(report.Total))
	sb.WriteString("\n\n")

	sb.WriteString("by_type,type,count\n")
	for _, item := range report.ByType {
		sb.WriteString("by_type,")
		sb.WriteString(string(item.Type))
		sb.WriteString(",")
		sb.WriteString(itoa(item.Count))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("by_user,user_first_name,user_last_name,count\n")
	for _, item := range report.ByUser {
		sb.WriteString("by_user,")
		sb.WriteString(item.FirstName)
		sb.WriteString(" ")
		sb.WriteString(item.LastName)
		sb.WriteString(",")
		sb.WriteString(itoa(item.Count))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("by_day,date,count\n")
	for _, item := range report.ByDay {
		sb.WriteString("by_day,")
		sb.WriteString(item.Date)
		sb.WriteString(",")
		sb.WriteString(itoa(item.Count))
		sb.WriteString("\n")
	}

	return c.String(http.StatusOK, sb.String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
