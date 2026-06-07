package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgUUIDToUUID(v pgtype.UUID) uuid.UUID {
	if v.Valid {
		return uuid.UUID(v.Bytes)
	}
	return uuid.Nil
}

func pgUUIDToPtr(v pgtype.UUID) *uuid.UUID {
	if v.Valid {
		id := uuid.UUID(v.Bytes)
		return &id
	}
	return nil
}

func pgTextToStr(v pgtype.Text) *string {
	if v.Valid {
		return &v.String
	}
	return nil
}

func pgTextToStrStr(v pgtype.Text) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func pgTimestamptzToTime(v pgtype.Timestamptz) *time.Time {
	if v.Valid {
		t := v.Time
		return &t
	}
	return nil
}

func pgInt4ToInt(v pgtype.Int4) int {
	if v.Valid {
		return int(v.Int32)
	}
	return 0
}

func pgInt4ToPtr(v pgtype.Int4) *int {
	if v.Valid {
		i := int(v.Int32)
		return &i
	}
	return nil
}

func pgNumericToFloat(v pgtype.Numeric) float64 {
	if v.Valid {
		f, _ := v.Float64Value()
		return f.Float64
	}
	return 0
}

func pgNumericToPtrFloat(v pgtype.Numeric) *float64 {
	if v.Valid {
		f, _ := v.Float64Value()
		return &f.Float64
	}
	return nil
}

func pgDateToTime(v pgtype.Date) *time.Time {
	if v.Valid {
		t := v.Time
		return &t
	}
	return nil
}
