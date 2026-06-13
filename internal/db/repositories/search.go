package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/domain/search"
)

type SearchRepository struct {
	pool *pgxpool.Pool
}

func NewSearchRepository(pool *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{pool: pool}
}

func (r *SearchRepository) Search(ctx context.Context, tenantID uuid.UUID, filter *search.SearchFilter) ([]*search.SearchResult, error) {
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 20
	}

	queries := make([]string, 0, 3)
	args := []interface{}{}
	argIdx := 1

	hasPerson := len(filter.Types) == 0
	hasCompany := len(filter.Types) == 0
	hasDeal := len(filter.Types) == 0
	for _, t := range filter.Types {
		switch t {
		case "person":
			hasPerson = true
		case "company":
			hasCompany = true
		case "deal":
			hasDeal = true
		}
	}

	if hasPerson {
		queries = append(queries, fmt.Sprintf(
			`SELECT id, $%d as tenant_id, 'person' as entity_type,
				first_name || ' ' || last_name as title,
				email::text as subtitle,
				ts_rank(search_person_text(first_name, last_name, email), plainto_tsquery('english', $%d)) as rank,
				created_at
			FROM persons
			WHERE tenant_id = $%d
			  AND deleted_at IS NULL
			  AND search_person_text(first_name, last_name, email) @@ plainto_tsquery('english', $%d)`,
			argIdx, argIdx+1, argIdx, argIdx+1,
		))
	}

	if hasCompany {
		if len(queries) > 0 {
			queries = append(queries, "UNION ALL")
		}
		queries = append(queries, fmt.Sprintf(
			`SELECT id, $%d as tenant_id, 'company' as entity_type,
				name as title,
				COALESCE(domain, '') as subtitle,
				ts_rank(search_company_text(name, domain), plainto_tsquery('english', $%d)) as rank,
				created_at
			FROM companies
			WHERE tenant_id = $%d
			  AND deleted_at IS NULL
			  AND search_company_text(name, domain) @@ plainto_tsquery('english', $%d)`,
			argIdx, argIdx+1, argIdx, argIdx+1,
		))
	}

	if hasDeal {
		if len(queries) > 0 {
			queries = append(queries, "UNION ALL")
		}
		queries = append(queries, fmt.Sprintf(
			`SELECT id, $%d as tenant_id, 'deal' as entity_type,
				title,
				'' as subtitle,
				ts_rank(to_tsvector('english', COALESCE(title, '')), plainto_tsquery('english', $%d)) as rank,
				created_at
			FROM deals
			WHERE tenant_id = $%d
			  AND deleted_at IS NULL
			  AND to_tsvector('english', COALESCE(title, '')) @@ plainto_tsquery('english', $%d)`,
			argIdx, argIdx+1, argIdx, argIdx+1,
		))
	}

	unionQuery := strings.Join(queries, " ")
	sql := unionQuery + fmt.Sprintf(`) sub
		WHERE rank > 0
		ORDER BY rank DESC, created_at DESC
		LIMIT $%d`, argIdx+2)

	args = append(args, tenantID, filter.Query, filter.Limit)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search.Search: %w", err)
	}
	defer rows.Close()

	var results []*search.SearchResult
	for rows.Next() {
		var r search.SearchResult
		err := rows.Scan(&r.ID, &r.TenantID, &r.EntityType, &r.Title, &r.Subtitle, &r.Rank, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("search.Search scan: %w", err)
		}
		results = append(results, &r)
	}

	return results, nil
}
