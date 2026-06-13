package activity

import (
	"time"

	"github.com/google/uuid"
)

type ActivityReportFilter struct {
	UserIDs []uuid.UUID
	TeamIDs []uuid.UUID
	Types   []ActivityType
	StartAt time.Time
	EndAt   time.Time
}

type ActivityCountByType struct {
	Type  ActivityType `json:"type"`
	Count int          `json:"count"`
}

type ActivityCountByUser struct {
	UserID    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Count     int       `json:"count"`
}

type ActivityCountByDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type ActivityReport struct {
	Total       int                  `json:"total"`
	ByType      []ActivityCountByType `json:"by_type"`
	ByUser      []ActivityCountByUser `json:"by_user"`
	ByDay       []ActivityCountByDay  `json:"by_day"`
	PeriodStart time.Time            `json:"period_start"`
	PeriodEnd   time.Time            `json:"period_end"`
}
