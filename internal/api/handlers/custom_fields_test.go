package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/oscar/oscar/internal/domain/custom_field"
)

type mockCustomFieldRepo struct {
	listAllFn      func(ctx context.Context, tenantID uuid.UUID) ([]*custom_field.CustomFieldDefinition, error)
	listByEntityFn func(ctx context.Context, tenantID uuid.UUID, entityType custom_field.EntityType) ([]*custom_field.CustomFieldDefinition, error)
	createFn       func(ctx context.Context, tenantID uuid.UUID, req *custom_field.CreateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*custom_field.CustomFieldDefinition, error)
	updateFn       func(ctx context.Context, id uuid.UUID, req *custom_field.UpdateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
	reorderFn      func(ctx context.Context, tenantID uuid.UUID, fieldIDs []uuid.UUID) error
}

func (m *mockCustomFieldRepo) ListAll(ctx context.Context, tenantID uuid.UUID) ([]*custom_field.CustomFieldDefinition, error) {
	return m.listAllFn(ctx, tenantID)
}
func (m *mockCustomFieldRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType custom_field.EntityType) ([]*custom_field.CustomFieldDefinition, error) {
	return m.listByEntityFn(ctx, tenantID, entityType)
}
func (m *mockCustomFieldRepo) Create(ctx context.Context, tenantID uuid.UUID, req *custom_field.CreateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error) {
	return m.createFn(ctx, tenantID, req)
}
func (m *mockCustomFieldRepo) GetByID(ctx context.Context, id uuid.UUID) (*custom_field.CustomFieldDefinition, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCustomFieldRepo) Update(ctx context.Context, id uuid.UUID, req *custom_field.UpdateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockCustomFieldRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}
func (m *mockCustomFieldRepo) Reorder(ctx context.Context, tenantID uuid.UUID, fieldIDs []uuid.UUID) error {
	return m.reorderFn(ctx, tenantID, fieldIDs)
}

func TestCustomFieldHandler_List(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("list all", func(t *testing.T) {
		fields := []*custom_field.CustomFieldDefinition{
			{ID: uuid.New(), Label: "Field 1"},
		}
		mock := &mockCustomFieldRepo{
			listAllFn: func(_ context.Context, _ uuid.UUID) ([]*custom_field.CustomFieldDefinition, error) {
				return fields, nil
			},
		}
		h := NewCustomFieldHandler(mock)
		c, rec := newGetContext(e, "/api/v1/custom-fields", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("filter by entity", func(t *testing.T) {
		fields := []*custom_field.CustomFieldDefinition{
			{ID: uuid.New(), Label: "Person Field"},
		}
		mock := &mockCustomFieldRepo{
			listByEntityFn: func(_ context.Context, _ uuid.UUID, entityType custom_field.EntityType) ([]*custom_field.CustomFieldDefinition, error) {
				if entityType != custom_field.EntityTypePerson {
					t.Errorf("entityType = %q", entityType)
				}
				return fields, nil
			},
		}
		h := NewCustomFieldHandler(mock)
		c, rec := newGetContext(e, "/api/v1/custom-fields?entity_type=person", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid entity type", func(t *testing.T) {
		mock := &mockCustomFieldRepo{}
		h := NewCustomFieldHandler(mock)
		c, rec := newGetContext(e, "/api/v1/custom-fields?entity_type=invalid", opts)

		_ = h.List(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestCustomFieldHandler_Create(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		field := &custom_field.CustomFieldDefinition{ID: uuid.New(), Label: "Test Field"}

		mock := &mockCustomFieldRepo{
			createFn: func(_ context.Context, _ uuid.UUID, req *custom_field.CreateCustomFieldRequest) (*custom_field.CustomFieldDefinition, error) {
				if req.Label != "Test Field" {
					t.Errorf("Label = %q", req.Label)
				}
				return field, nil
			},
		}
		h := NewCustomFieldHandler(mock)
		body := `{
			"entity_type": "person",
			"field_key": "test_field",
			"label": "Test Field",
			"field_type": "text",
			"is_required": false,
			"show_in_list": true,
			"show_in_card": true,
			"position": 0
		}`
		c, rec := newPostContext(e, "/api/v1/custom-fields", body, opts)

		if err := h.Create(c); err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockCustomFieldRepo{}
		h := NewCustomFieldHandler(mock)
		c, rec := newPostContext(e, "/api/v1/custom-fields", "invalid json", opts)

		_ = h.Create(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestCustomFieldHandler_Get(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		field := &custom_field.CustomFieldDefinition{ID: id, Label: "Test"}

		mock := &mockCustomFieldRepo{
			getByIDFn: func(_ context.Context, got uuid.UUID) (*custom_field.CustomFieldDefinition, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return field, nil
			},
		}
		h := NewCustomFieldHandler(mock)
		c, rec := newGetContext(e, "/api/v1/custom-fields/"+id.String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		if err := h.Get(c); err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockCustomFieldRepo{}
		h := NewCustomFieldHandler(mock)
		c, rec := newGetContext(e, "/api/v1/custom-fields/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_ = h.Get(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestCustomFieldHandler_Delete(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		mock := &mockCustomFieldRepo{
			deleteFn: func(_ context.Context, got uuid.UUID) error {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return nil
			},
		}
		h := NewCustomFieldHandler(mock)
		c, rec := newDeleteContext(e, "/api/v1/custom-fields/"+id.String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		if err := h.Delete(c); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &mockCustomFieldRepo{
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return errors.New("delete failed")
			},
		}
		h := NewCustomFieldHandler(mock)
		c, rec := newDeleteContext(e, "/api/v1/custom-fields/"+uuid.New().String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(uuid.New().String())

		_ = h.Delete(c)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestCustomFieldHandler_Reorder(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()

		mock := &mockCustomFieldRepo{
			reorderFn: func(_ context.Context, _ uuid.UUID, fieldIDs []uuid.UUID) error {
				if len(fieldIDs) != 2 {
					t.Errorf("got %d field IDs", len(fieldIDs))
				}
				return nil
			},
		}
		h := NewCustomFieldHandler(mock)
		body := `{"field_ids":["` + id1.String() + `","` + id2.String() + `"]}`
		c, rec := newPostContext(e, "/api/v1/custom-fields/reorder", body, opts)

		if err := h.Reorder(c); err != nil {
			t.Fatalf("Reorder() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid field ids", func(t *testing.T) {
		mock := &mockCustomFieldRepo{}
		h := NewCustomFieldHandler(mock)
		body := `{"field_ids":["invalid-uuid"]}`
		c, rec := newPostContext(e, "/api/v1/custom-fields/reorder", body, opts)

		_ = h.Reorder(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
