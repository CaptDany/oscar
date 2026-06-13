package search

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SearchResult struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	EntityType  string    `json:"entity_type"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle,omitempty"`
	Rank        float64   `json:"rank"`
	CreatedAt   time.Time `json:"created_at"`
}

type SearchFilter struct {
	Query string   `json:"q"`
	Types []string `json:"types,omitempty"`
	Limit int      `json:"limit"`
	Page  int      `json:"page"`
}

type Repository interface {
	Search(ctx context.Context, tenantID uuid.UUID, filter *SearchFilter) ([]*SearchResult, error)
}
