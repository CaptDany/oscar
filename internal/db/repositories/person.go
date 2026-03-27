package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oscar/oscar/internal/db/generated"
	"github.com/oscar/oscar/internal/domain/person"
)

type PersonRepository struct {
	pool *pgxpool.Pool
}

func NewPersonRepository(pool *pgxpool.Pool) *PersonRepository {
	return &PersonRepository{pool: pool}
}

func (r *PersonRepository) Create(ctx context.Context, tenantID uuid.UUID, req *person.CreatePersonRequest) (*person.Person, error) {
	query := `
		INSERT INTO persons (tenant_id, type, status, first_name, last_name, email, phone, avatar_url, company_id, owner_id, source, score, tags, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING *
	`

	status := req.Status
	if status == "" {
		status = person.PersonStatusNew
	}

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query,
		tenantID, req.Type, status, req.FirstName, req.LastName,
		req.Email, req.Phone, req.AvatarURL, req.CompanyID, req.OwnerID,
		req.Source, req.Score, req.Tags, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.Create: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	query := `SELECT * FROM persons WHERE id = $1 AND deleted_at IS NULL`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("person.GetByID: person not found")
		}
		return nil, fmt.Errorf("person.GetByID: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) Update(ctx context.Context, id uuid.UUID, req *person.UpdatePersonRequest) (*person.Person, error) {
	query := `
		UPDATE persons SET
			type = COALESCE($2, type),
			status = COALESCE($3, status),
			first_name = COALESCE($4, first_name),
			last_name = COALESCE($5, last_name),
			email = COALESCE($6, email),
			phone = COALESCE($7, phone),
			avatar_url = COALESCE($8, avatar_url),
			company_id = COALESCE($9, company_id),
			owner_id = COALESCE($10, owner_id),
			source = COALESCE($11, source),
			score = COALESCE($12, score),
			tags = COALESCE($13, tags),
			custom_fields = COALESCE($14, custom_fields)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query,
		id, req.Type, req.Status, req.FirstName, req.LastName,
		req.Email, req.Phone, req.AvatarURL, req.CompanyID, req.OwnerID,
		req.Source, req.Score, req.Tags, req.CustomFields,
	).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.Update: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) SoftDelete(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	query := `UPDATE persons SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING *`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.SoftDelete: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) Convert(ctx context.Context, id uuid.UUID, toType person.PersonType, status person.PersonStatus) (*person.Person, error) {
	query := `
		UPDATE persons
		SET type = $2, status = $3, converted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id, toType, status).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.Convert: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) List(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) ([]*person.Person, string, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.OwnerID != nil {
		baseQuery += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *filter.OwnerID)
		argIdx++
	}
	if filter.CompanyID != nil {
		baseQuery += fmt.Sprintf(" AND company_id = $%d", argIdx)
		args = append(args, *filter.CompanyID)
		argIdx++
	}
	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email::text ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM persons ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("person.List count: %w", err)
	}

	cursor := filter.Cursor
	offset := 0
	if cursor != "" {
		if cursorID, err := uuid.Parse(cursor); err == nil {
			cursorQuery := `SELECT created_at FROM persons WHERE id = $1`
			var cursorTime time.Time
			if err := r.pool.QueryRow(ctx, cursorQuery, cursorID).Scan(&cursorTime); err == nil {
				args = append(args, cursorTime)
				baseQuery += fmt.Sprintf(" AND created_at < $%d", argIdx)
			}
		}
	}

	listQuery := `SELECT * FROM persons ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx+1) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+2)
	args = append(args, limit+1, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("person.List: %w", err)
	}
	defer rows.Close()

	var persons []*person.Person
	for rows.Next() {
		row := &generated.Person{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
			&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
			&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, "", 0, fmt.Errorf("person.List scan: %w", err)
		}
		persons = append(persons, mapPersonRowToDomain(row))
	}

	nextCursor := ""
	if len(persons) > limit {
		persons = persons[:limit]
		nextCursor = persons[len(persons)-1].ID.String()
	}

	return persons, nextCursor, total, nil
}

func (r *PersonRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*person.Person, error) {
	if limit <= 0 {
		limit = 20
	}

	sql := `
		SELECT * FROM persons 
		WHERE tenant_id = $1 
		  AND deleted_at IS NULL
		  AND (
		    first_name ILIKE $2 OR
		    last_name ILIKE $2 OR
		    email::text ILIKE $2
		  )
		ORDER BY first_name ASC, last_name ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, sql, tenantID, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("person.Search: %w", err)
	}
	defer rows.Close()

	var persons []*person.Person
	for rows.Next() {
		row := &generated.Person{}
		err := rows.Scan(
			&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
			&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
			&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("person.Search scan: %w", err)
		}
		persons = append(persons, mapPersonRowToDomain(row))
	}

	return persons, nil
}

func (r *PersonRepository) Count(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) (int, error) {
	baseQuery := `WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
	}

	countQuery := `SELECT COUNT(*) FROM persons ` + baseQuery
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("person.Count: %w", err)
	}

	return total, nil
}

func (r *PersonRepository) AddTag(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error) {
	query := `
		UPDATE persons
		SET tags = array_distinct(array_append(tags, $2))
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id, tag).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.AddTag: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) RemoveTag(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error) {
	query := `
		UPDATE persons
		SET tags = array_remove(tags, $2)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id, tag).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.RemoveTag: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func (r *PersonRepository) UpdateScore(ctx context.Context, id uuid.UUID, score int) (*person.Person, error) {
	query := `
		UPDATE persons
		SET score = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING *
	`

	row := &generated.Person{}
	err := r.pool.QueryRow(ctx, query, id, score).Scan(
		&row.ID, &row.TenantID, &row.Type, &row.Status, &row.FirstName, &row.LastName,
		&row.Email, &row.Phone, &row.AvatarURL, &row.CompanyID, &row.OwnerID,
		&row.Source, &row.Score, &row.Tags, &row.CustomFields, &row.ConvertedAt,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("person.UpdateScore: %w", err)
	}

	return mapPersonRowToDomain(row), nil
}

func mapPersonRowToDomain(row *generated.Person) *person.Person {
	return &person.Person{
		ID:          row.ID,
		TenantID:    row.TenantID,
		Type:        row.Type,
		Status:      row.Status,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		Email:       row.Email,
		Phone:       row.Phone,
		AvatarURL:   row.AvatarURL,
		CompanyID:   row.CompanyID,
		OwnerID:     row.OwnerID,
		Source:      row.Source,
		Score:       int(row.Score),
		Tags:        row.Tags,
		CustomFields: row.CustomFields,
		ConvertedAt: row.ConvertedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		DeletedAt:   row.DeletedAt,
	}
}
