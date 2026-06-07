package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oscar/oscar/internal/domain/audit_log"
)

type mockAuditLogRepo struct {
	listFn   func(ctx context.Context, tenantID uuid.UUID, filter *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*audit_log.AuditLog, error)
}

func (m *mockAuditLogRepo) List(ctx context.Context, tenantID uuid.UUID, filter *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error) {
	return m.listFn(ctx, tenantID, filter)
}
func (m *mockAuditLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*audit_log.AuditLog, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockAuditLogRepo) Create(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, entityType string, entityID uuid.UUID, diff json.RawMessage, ipAddress, userAgent *string) (*audit_log.AuditLog, error) {
	panic("unimplemented")
}
func (m *mockAuditLogRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit, offset int) ([]*audit_log.AuditLog, error) {
	panic("unimplemented")
}
func (m *mockAuditLogRepo) ListByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*audit_log.AuditLog, error) {
	panic("unimplemented")
}
func (m *mockAuditLogRepo) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	panic("unimplemented")
}

func TestAuditLogHandler_List(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		entries := []*audit_log.AuditLog{
			{ID: uuid.New(), EntityType: "person", Action: "created", CreatedAt: now},
		}

		mock := &mockAuditLogRepo{
			listFn: func(_ context.Context, _ uuid.UUID, filter *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error) {
				return entries, "", 1, nil
			},
		}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("filter by entity_type", func(t *testing.T) {
		mock := &mockAuditLogRepo{
			listFn: func(_ context.Context, _ uuid.UUID, filter *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error) {
				if filter.EntityType == nil || *filter.EntityType != "person" {
					t.Errorf("EntityType = %v", filter.EntityType)
				}
				return []*audit_log.AuditLog{}, "", 0, nil
			},
		}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs?entity_type=person", opts)

		if err := h.List(c); err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		mock := &mockAuditLogRepo{
			listFn: func(_ context.Context, _ uuid.UUID, _ *audit_log.ListAuditLogsFilter) ([]*audit_log.AuditLog, string, int, error) {
				return nil, "", 0, errors.New("db error")
			},
		}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs", opts)

		_ = h.List(c)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestAuditLogHandler_Get(t *testing.T) {
	e := setupEcho()
	opts := defaultContextOpts()

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		entry := &audit_log.AuditLog{
			ID: id, EntityType: "deal", Action: "updated",
		}

		mock := &mockAuditLogRepo{
			getByIDFn: func(_ context.Context, got uuid.UUID) (*audit_log.AuditLog, error) {
				if got != id {
					t.Errorf("got id %v, want %v", got, id)
				}
				return entry, nil
			},
		}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs/"+id.String(), opts)
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
		mock := &mockAuditLogRepo{
			getByIDFn: func(_ context.Context, _ uuid.UUID) (*audit_log.AuditLog, error) {
				return nil, errors.New("not found")
			},
		}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs/"+uuid.New().String(), opts)
		c.SetParamNames("id")
		c.SetParamValues(uuid.New().String())

		_ = h.Get(c)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mock := &mockAuditLogRepo{}
		h := NewAuditLogHandler(mock)
		c, rec := newGetContext(e, "/api/v1/audit-logs/invalid", opts)
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_ = h.Get(c)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
