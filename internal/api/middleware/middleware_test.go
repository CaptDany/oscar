package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func echoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestRequestID(t *testing.T) {
	c, rec := echoContext(http.MethodGet, "/")
	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should not be empty")
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	c, rec := echoContext(http.MethodGet, "/")
	c.Request().Header.Set("X-Request-ID", "existing-id")

	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != "existing-id" {
		t.Errorf("X-Request-ID = %q, want %q", requestID, "existing-id")
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		userRoles  []string
		required   []string
		wantAccess bool
	}{
		{"single match", []string{"Owner"}, []string{"Owner"}, true},
		{"multiple match", []string{"Admin", "Editor"}, []string{"Editor"}, true},
		{"no match", []string{"Read Only"}, []string{"Owner"}, false},
		{"empty user roles", []string{}, []string{"Owner"}, false},
		{"any role works", []string{"Viewer"}, []string{"Owner", "Viewer", "Admin"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := echoContext(http.MethodGet, "/")
			c.Set("roles", tt.userRoles)

			var called bool
			handler := RequireRole(tt.required...)(func(c echo.Context) error {
				called = true
				return c.String(http.StatusOK, "ok")
			})

			err := handler(c)
			if tt.wantAccess && err != nil {
				t.Errorf("expected access, got error: %v", err)
			}
			if !tt.wantAccess && err == nil {
				t.Error("expected forbidden")
			}
			if tt.wantAccess != called {
				t.Errorf("handler called = %v, want %v", called, tt.wantAccess)
			}
		})
	}
}

func TestRequireRole_MissingRoles(t *testing.T) {
	c, _ := echoContext(http.MethodGet, "/")

	handler := RequireRole("Owner")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error when roles not set")
	}
}

func TestRecover_Panics(t *testing.T) {
	c, _ := echoContext(http.MethodGet, "/")

	handler := Recover()(func(c echo.Context) error {
		panic("test panic")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if err != echo.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestRecover_NoPanic(t *testing.T) {
	c, rec := echoContext(http.MethodGet, "/")

	var called bool
	handler := Recover()(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestInMemoryRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(nil, 3, time.Minute)

	for i := 0; i < 3; i++ {
		allowed, count, err := rl.Allow("127.0.0.1")
		if err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}
		if !allowed {
			t.Errorf("iteration %d: not allowed (count=%d)", i, count)
		}
	}

	allowed, _, _ := rl.Allow("127.0.0.1")
	if allowed {
		t.Error("expected rate limited after 3 requests")
	}
}

func TestInMemoryRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(nil, 1, time.Minute)

	allowed1, _, _ := rl.Allow("192.168.1.1")
	if !allowed1 {
		t.Error("first request from 192.168.1.1 should be allowed")
	}

	allowed2, _, _ := rl.Allow("192.168.1.2")
	if !allowed2 {
		t.Error("first request from 192.168.1.2 should be allowed")
	}

	blocked, _, _ := rl.Allow("192.168.1.1")
	if blocked {
		t.Error("second request from 192.168.1.1 should be blocked")
	}
}

func TestInMemoryRateLimiter_Expiry(t *testing.T) {
	rl := NewRateLimiter(nil, 1, 50*time.Millisecond)

	allowed, _, _ := rl.Allow("127.0.0.1")
	if !allowed {
		t.Error("first request should be allowed")
	}

	time.Sleep(60 * time.Millisecond)

	allowed, _, _ = rl.Allow("127.0.0.1")
	if !allowed {
		t.Error("request after expiry should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	c, _ := echoContext(http.MethodGet, "/")
	rl := NewRateLimiter(nil, 2, time.Minute)
	mw := RateLimitMiddleware(rl)

	called := false
	handler := mw(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRateLimitMiddleware_Blocked(t *testing.T) {
	c, _ := echoContext(http.MethodGet, "/")
	rl := NewRateLimiter(nil, 1, time.Minute)
	mw := RateLimitMiddleware(rl)

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	_ = handler(c)

	c2, _ := echoContext(http.MethodGet, "/")
	err := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c2)

	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if err != echo.ErrTooManyRequests {
		t.Errorf("expected ErrTooManyRequests, got %v", err)
	}
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(nil, 10, time.Minute)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.maxReqs != 10 {
		t.Errorf("maxReqs = %d", rl.maxReqs)
	}
}

func TestInMemoryRateLimiter_ConcurrentSafe(t *testing.T) {
	rl := NewRateLimiter(nil, 100, time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := rl.Allow("10.0.0.1")
			if err != nil {
				t.Errorf("Allow() returned error: %v", err)
			}
		}()
	}
	wg.Wait()
}
