package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oscar/oscar/internal/domain/company"
)

func TestCompanyHandler_List(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		companies := []*company.Company{
			{ID: uuid.New(), Name: "Acme Corp", CreatedAt: now},
		}

		mock := &mockCompanyRepo{
			listFn: func(_ context.Context, _ uuid.UUID, _ *company.ListCompaniesFilter) ([]*company.Company, string, int, error) {
				return companies, "", 1, nil
			},
		}
		h := NewCompanyHandler(mock)
		c, rec := newGetContext(e, "/api/v1/companies", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &mockCompanyRepo{
			listFn: func(_ context.Context, _ uuid.UUID, _ *company.ListCompaniesFilter) ([]*company.Company, string, int, error) {
				return nil, "", 0, errors.New("db error")
			},
		}
		h := NewCompanyHandler(mock)
		c, rec := newGetContext(e, "/api/v1/companies?include_total=true", opts)

		if err := h.List(c); err == nil {
			t.Fatal("expected error")
		}
		_ = rec.Code
	})
}

func TestCompanyHandler_Get(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		comp := &company.Company{ID: id, Name: "Acme Corp"}

		mock := &mockCompanyRepo{
			getByIDFn: func(_ context.Context, got uuid.UUID) (*company.Company, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return comp, nil
			},
		}
		h := NewCompanyHandler(mock)
		c, rec := newGetContext(e, "/api/v1/companies/"+id.String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		if err := h.Get(c); err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockCompanyRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*company.Company, error) {
				return nil, errors.New("not found")
			},
		}
		h := NewCompanyHandler(mock)
		c, _ := newGetContext(e, "/api/v1/companies/"+uuid.New().String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(uuid.New().String())

		if err := h.Get(c); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockCompanyRepo{}
		h := NewCompanyHandler(mock)
		c, _ := newGetContext(e, "/api/v1/companies/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		if err := h.Get(c); err == nil {
			t.Fatal("expected error for invalid id")
		}
	})
}

func TestCompanyHandler_Create(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		comp := &company.Company{ID: uuid.New(), Name: "New Corp"}

		mock := &mockCompanyRepo{
			createFn: func(_ context.Context, _ uuid.UUID, req *company.CreateCompanyRequest) (*company.Company, error) {
				if req.Name != "New Corp" {
					t.Errorf("Name = %q", req.Name)
				}
				return comp, nil
			},
		}
		h := NewCompanyHandler(mock)
		body := `{"name":"New Corp"}`
		c, rec := newPostContext(e, "/api/v1/companies", body, opts)

		if err := h.Create(c); err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockCompanyRepo{}
		h := NewCompanyHandler(mock)
		c, _ := newPostContext(e, "/api/v1/companies", "invalid json", opts)

		if err := h.Create(c); err == nil {
			t.Fatal("expected error for invalid body")
		}
	})
}

func TestCompanyHandler_Update(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		updated := &company.Company{ID: id, Name: "Updated Corp"}

		mock := &mockCompanyRepo{
			updateFn: func(_ context.Context, got uuid.UUID, _ *company.UpdateCompanyRequest) (*company.Company, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return updated, nil
			},
		}
		h := NewCompanyHandler(mock)
		body := `{"name":"Updated Corp"}`
		c, rec := newPutContext(e, "/api/v1/companies/"+id.String(), body, opts)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		if err := h.Update(c); err != nil {
			t.Fatalf("Update() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockCompanyRepo{}
		h := NewCompanyHandler(mock)
		body := `{"name":"Test"}`
		c, _ := newPutContext(e, "/api/v1/companies/invalid", body, opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		if err := h.Update(c); err == nil {
			t.Fatal("expected error for invalid id")
		}
	})
}

func TestCompanyHandler_Delete(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		deleted := &company.Company{ID: id, Name: "Deleted Corp"}

		mock := &mockCompanyRepo{
			softDeleteFn: func(_ context.Context, got uuid.UUID) (*company.Company, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return deleted, nil
			},
		}
		h := NewCompanyHandler(mock)
		c, rec := newDeleteContext(e, "/api/v1/companies/"+id.String(), opts)
		c.SetParamNames("id")
	c.SetParamValues(id.String())

		if err := h.Delete(c); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockCompanyRepo{}
		h := NewCompanyHandler(mock)
		c, _ := newDeleteContext(e, "/api/v1/companies/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		if err := h.Delete(c); err == nil {
			t.Fatal("expected error for invalid id")
		}
	})
}
