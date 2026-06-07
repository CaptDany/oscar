package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/oscar/oscar/internal/domain/company"
	"github.com/oscar/oscar/internal/domain/person"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (v *testValidator) Validate(i interface{}) error {
	return nil
}

type contextOptions struct {
	tenantID uuid.UUID
	userID   uuid.UUID
	roles    []string
}

func defaultContextOpts() contextOptions {
	return contextOptions{
		tenantID: uuid.New(),
		userID:   uuid.New(),
		roles:    []string{"Owner"},
	}
}

func newContext(e *echo.Echo, method, path, body string, opts contextOptions) (echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("tenant_id", opts.tenantID)
	c.Set("user_id", opts.userID)
	c.Set("roles", opts.roles)
	return c, rec
}

func newGetContext(e *echo.Echo, path string, opts contextOptions) (echo.Context, *httptest.ResponseRecorder) {
	return newContext(e, http.MethodGet, path, "", opts)
}

func newPostContext(e *echo.Echo, path, body string, opts contextOptions) (echo.Context, *httptest.ResponseRecorder) {
	return newContext(e, http.MethodPost, path, body, opts)
}

func newPutContext(e *echo.Echo, path, body string, opts contextOptions) (echo.Context, *httptest.ResponseRecorder) {
	return newContext(e, http.MethodPut, path, body, opts)
}

func newDeleteContext(e *echo.Echo, path string, opts contextOptions) (echo.Context, *httptest.ResponseRecorder) {
	return newContext(e, http.MethodDelete, path, "", opts)
}

type mockPersonRepo struct {
	createFn     func(ctx context.Context, tenantID uuid.UUID, req *person.CreatePersonRequest) (*person.Person, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*person.Person, error)
	updateFn     func(ctx context.Context, id uuid.UUID, req *person.UpdatePersonRequest) (*person.Person, error)
	softDeleteFn func(ctx context.Context, id uuid.UUID) (*person.Person, error)
	listFn       func(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) ([]*person.Person, string, int, error)
	searchFn     func(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*person.Person, error)
	convertFn    func(ctx context.Context, id uuid.UUID, toType person.PersonType, status person.PersonStatus) (*person.Person, error)
	addTagFn     func(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error)
	removeTagFn  func(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error)
	countFn      func(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) (int, error)
	updateScoreFn func(ctx context.Context, id uuid.UUID, score int) (*person.Person, error)
}

func (m *mockPersonRepo) Create(ctx context.Context, tenantID uuid.UUID, req *person.CreatePersonRequest) (*person.Person, error) {
	return m.createFn(ctx, tenantID, req)
}
func (m *mockPersonRepo) GetByID(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockPersonRepo) Update(ctx context.Context, id uuid.UUID, req *person.UpdatePersonRequest) (*person.Person, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockPersonRepo) SoftDelete(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	return m.softDeleteFn(ctx, id)
}
func (m *mockPersonRepo) List(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) ([]*person.Person, string, int, error) {
	return m.listFn(ctx, tenantID, filter)
}
func (m *mockPersonRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*person.Person, error) {
	return m.searchFn(ctx, tenantID, query, limit, offset)
}
func (m *mockPersonRepo) Convert(ctx context.Context, id uuid.UUID, toType person.PersonType, status person.PersonStatus) (*person.Person, error) {
	return m.convertFn(ctx, id, toType, status)
}
func (m *mockPersonRepo) AddTag(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error) {
	return m.addTagFn(ctx, id, tag)
}
func (m *mockPersonRepo) RemoveTag(ctx context.Context, id uuid.UUID, tag string) (*person.Person, error) {
	return m.removeTagFn(ctx, id, tag)
}
func (m *mockPersonRepo) Count(ctx context.Context, tenantID uuid.UUID, filter *person.ListPersonsFilter) (int, error) {
	return m.countFn(ctx, tenantID, filter)
}
func (m *mockPersonRepo) UpdateScore(ctx context.Context, id uuid.UUID, score int) (*person.Person, error) {
	return m.updateScoreFn(ctx, id, score)
}

type mockCompanyRepo struct {
	createFn     func(ctx context.Context, tenantID uuid.UUID, req *company.CreateCompanyRequest) (*company.Company, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*company.Company, error)
	updateFn     func(ctx context.Context, id uuid.UUID, req *company.UpdateCompanyRequest) (*company.Company, error)
	softDeleteFn func(ctx context.Context, id uuid.UUID) (*company.Company, error)
	listFn       func(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) ([]*company.Company, string, int, error)
	searchFn     func(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*company.Company, error)
	countFn      func(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) (int, error)
}

func (m *mockCompanyRepo) Create(ctx context.Context, tenantID uuid.UUID, req *company.CreateCompanyRequest) (*company.Company, error) {
	return m.createFn(ctx, tenantID, req)
}
func (m *mockCompanyRepo) GetByID(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCompanyRepo) Update(ctx context.Context, id uuid.UUID, req *company.UpdateCompanyRequest) (*company.Company, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockCompanyRepo) SoftDelete(ctx context.Context, id uuid.UUID) (*company.Company, error) {
	return m.softDeleteFn(ctx, id)
}
func (m *mockCompanyRepo) List(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) ([]*company.Company, string, int, error) {
	return m.listFn(ctx, tenantID, filter)
}
func (m *mockCompanyRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*company.Company, error) {
	return m.searchFn(ctx, tenantID, query, limit, offset)
}
func (m *mockCompanyRepo) Count(ctx context.Context, tenantID uuid.UUID, filter *company.ListCompaniesFilter) (int, error) {
	return m.countFn(ctx, tenantID, filter)
}

func uuidPtr(s string) *uuid.UUID {
	id := uuid.MustParse(s)
	return &id
}
