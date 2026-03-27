package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo *echo.Echo
}

func New() *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = customErrorHandler

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Request-ID"},
	}))
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.Recover())

	e.GET("/health", healthCheck)

	return &Server{echo: e}
}

func (s *Server) Group(prefix string, m ...echo.MiddlewareFunc) *echo.Group {
	return s.echo.Group(prefix, m...)
}

func (s *Server) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.GET(path, h, m...)
}

func (s *Server) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.POST(path, h, m...)
}

func (s *Server) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.PUT(path, h, m...)
}

func (s *Server) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.PATCH(path, h, m...)
}

func (s *Server) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.DELETE(path, h, m...)
}

func (s *Server) WS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	s.echo.GET(path, h, m...)
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

func (s *Server) Shutdown() error {
	return s.echo.Close()
}

func (s *Server) Echo() *echo.Echo {
	return s.echo
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"version": "1.0.0",
	})
}

func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	tenantID := c.Get("tenant_id")
	requestID := c.Get("request_id")

	c.JSON(code, map[string]interface{}{
		"error": map[string]interface{}{
			"code":       code,
			"message":    message,
			"request_id": requestID,
			"tenant_id":  tenantID,
		},
	})
}

type TenantContext interface {
	GetTenantID() uuid.UUID
	GetUserID() uuid.UUID
	GetRoles() []string
}

type tenantContext struct {
	tenantID uuid.UUID
	userID   uuid.UUID
	roles    []string
}

func (tc *tenantContext) GetTenantID() uuid.UUID {
	return tc.tenantID
}

func (tc *tenantContext) GetUserID() uuid.UUID {
	return tc.userID
}

func (tc *tenantContext) GetRoles() []string {
	return tc.roles
}
