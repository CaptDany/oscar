package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oscar/oscar/internal/domain/person"
)

func TestPersonHandler_List(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		persons := []*person.Person{
			{ID: uuid.New(), FirstName: "Alice", LastName: "Smith", CreatedAt: now},
			{ID: uuid.New(), FirstName: "Bob", LastName: "Jones", CreatedAt: now},
		}

		mock := &mockPersonRepo{
			listFn: func(ctx context.Context, _ uuid.UUID, _ *person.ListPersonsFilter) ([]*person.Person, string, int, error) {
				return persons, "next_cursor", 2, nil
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		mock := &mockPersonRepo{
			listFn: func(ctx context.Context, _ uuid.UUID, _ *person.ListPersonsFilter) ([]*person.Person, string, int, error) {
				return []*person.Person{}, "", 0, nil
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &mockPersonRepo{
			listFn: func(ctx context.Context, _ uuid.UUID, _ *person.ListPersonsFilter) ([]*person.Person, string, int, error) {
				return nil, "", 0, errors.New("db error")
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons", opts)

		_ = h.List(c)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestPersonHandler_Get(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		p := &person.Person{ID: id, FirstName: "Alice", LastName: "Smith"}

		mock := &mockPersonRepo{
			getByIDFn: func(_ context.Context, got uuid.UUID) (*person.Person, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return p, nil
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons/"+id.String(), opts)
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
		mock := &mockPersonRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*person.Person, error) {
				return nil, errors.New("not found")
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons/"+uuid.New().String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(uuid.New().String())

		_ = h.Get(c)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockPersonRepo{}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_ = h.Get(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPersonHandler_Create(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		newPerson := &person.Person{
			ID:        uuid.New(),
			FirstName: "Alice",
			LastName:  "Smith",
		}

		mock := &mockPersonRepo{
			createFn: func(_ context.Context, _ uuid.UUID, req *person.CreatePersonRequest) (*person.Person, error) {
				if req.FirstName != "Alice" {
					t.Errorf("FirstName = %q", req.FirstName)
				}
				return newPerson, nil
			},
		}
		h := NewPersonHandler(mock)
		body := `{"first_name":"Alice","last_name":"Smith","type":"lead"}`
		c, rec := newPostContext(e, "/api/v1/persons", body, opts)

		if err := h.Create(c); err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		mock := &mockPersonRepo{}
		h := NewPersonHandler(mock)
		c, rec := newPostContext(e, "/api/v1/persons", "invalid json", opts)

		_ = h.Create(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPersonHandler_Update(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		updated := &person.Person{ID: id, FirstName: "Updated", LastName: "Name"}

		mock := &mockPersonRepo{
			updateFn: func(_ context.Context, got uuid.UUID, _ *person.UpdatePersonRequest) (*person.Person, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return updated, nil
			},
		}
		h := NewPersonHandler(mock)
		body := `{"first_name":"Updated","last_name":"Name"}`
		c, rec := newPutContext(e, "/api/v1/persons/"+id.String(), body, opts)
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
		mock := &mockPersonRepo{}
		h := NewPersonHandler(mock)
		body := `{"first_name":"Test"}`
		c, rec := newPutContext(e, "/api/v1/persons/invalid", body, opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_ = h.Update(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPersonHandler_Delete(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		deleted := &person.Person{ID: id, FirstName: "Deleted"}

		mock := &mockPersonRepo{
			softDeleteFn: func(_ context.Context, got uuid.UUID) (*person.Person, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return deleted, nil
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newDeleteContext(e, "/api/v1/persons/"+id.String(), opts)
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
		mock := &mockPersonRepo{}
		h := NewPersonHandler(mock)
		c, rec := newDeleteContext(e, "/api/v1/persons/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_ = h.Delete(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPersonHandler_Search(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		results := []*person.Person{
			{ID: uuid.New(), FirstName: "Found", LastName: "Person"},
		}

		mock := &mockPersonRepo{
			searchFn: func(_ context.Context, _ uuid.UUID, query string, limit, offset int) ([]*person.Person, error) {
				if query != "alice" {
					t.Errorf("query = %q", query)
				}
				return results, nil
			},
		}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons/search?q=alice", opts)

		if err := h.Search(c); err != nil {
			t.Fatalf("Search() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("missing query", func(t *testing.T) {
		mock := &mockPersonRepo{}
		h := NewPersonHandler(mock)
		c, rec := newGetContext(e, "/api/v1/persons/search", opts)

		_ = h.Search(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPersonHandler_Convert(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		converted := &person.Person{ID: id, Type: person.PersonTypeContact}

		mock := &mockPersonRepo{
			convertFn: func(_ context.Context, got uuid.UUID, toType person.PersonType, status person.PersonStatus) (*person.Person, error) {
				if toType != person.PersonTypeContact {
					t.Errorf("toType = %q", toType)
				}
				return converted, nil
			},
		}
		h := NewPersonHandler(mock)
		body := `{"type":"contact","status":"active"}`
		c, rec := newPostContext(e, "/api/v1/persons/"+id.String()+"/convert", body, opts)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		if err := h.Convert(c); err != nil {
			t.Fatalf("Convert() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}


