package pagination

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

type Cursor struct {
	ID        uuid.UUID
	CreatedAt string
}

func EncodeCursor(id uuid.UUID, createdAt string) string {
	data := fmt.Sprintf("%s:%s", id.String(), createdAt)
	return base64.URLEncoding.EncodeToString([]byte(data))
}

func DecodeCursor(encoded string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding")
	}

	var id, createdAt string
	n, _ := fmt.Sscanf(string(data), "%s:%s", &id, &createdAt)
	if n != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor UUID")
	}

	return &Cursor{
		ID:        parsedID,
		CreatedAt: createdAt,
	}, nil
}

type Meta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	Total      int    `json:"total"`
}

func NewMeta(items []interface{}, total int, limit int) *Meta {
	meta := &Meta{
		Total: total,
	}

	if len(items) > limit {
		meta.NextCursor = "more"
	}

	return meta
}
