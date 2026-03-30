package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/oscar/oscar/pkg/errs"
)

type Server struct {
	echo *echo.Echo
}

func New() *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	v := validator.New()

	v.RegisterValidation("titlecase", func(fl validator.FieldLevel) bool {
		return isTitleCase(fl.Field().String())
	})
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		return isValidPhone(fl.Field().String())
	})

	e.Validator = &CustomValidator{validator: v}

	e.HTTPErrorHandler = customErrorHandler

	e.Use(minimalLogger())
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

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func isTitleCase(s string) bool {
	if s == "" {
		return true
	}
	words := strings.Fields(s)
	for _, word := range words {
		if len(word) == 0 {
			return false
		}
		if !unicode.IsUpper(rune(word[0])) || !unicode.IsLower(rune(word[1])) {
			return false
		}
		for i := 1; i < len(word); i++ {
			if unicode.IsUpper(rune(word[i])) {
				return false
			}
		}
	}
	return true
}

func isValidPhone(phone string) bool {
	if phone == "" {
		return true
	}
	cleaned := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")
	return len(cleaned) >= 10 && len(cleaned) <= 15 && regexp.MustCompile(`^\+?[0-9\s\-\+\(\)]+$`).MatchString(phone)
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
		"status":  "healthy",
		"version": "1.0.0",
	})
}

func minimalLogger() echo.MiddlewareFunc {
	reset := "\033[0m"
	green := "\033[32m"
	yellow := "\033[33m"
	red := "\033[31m"
	cyan := "\033[36m"
	purple := "\033[35m"
	dim := "\033[2m"

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			status := c.Response().Status
			method := c.Request().Method
			path := c.Request().URL.Path
			timestamp := start.Format("15:04:05")

			statusIcon := green + "✓" + reset
			statusColor := green
			if status >= 500 {
				statusIcon = red + "✗" + reset
				statusColor = red
			} else if status >= 400 {
				statusIcon = yellow + "!" + reset
				statusColor = yellow
			}

			methodColor := cyan
			if method == "POST" {
				methodColor = green
			} else if method == "PUT" || method == "PATCH" {
				methodColor = yellow
			} else if method == "DELETE" {
				methodColor = red
			}

			statusStr := fmt.Sprintf("%s%d%s", statusColor, status, reset)
			fmt.Printf("%s[%s]%s %s%-7s%s %s  %s%s  %s\n",
				dim, timestamp,
				reset, methodColor, method,
				reset, statusIcon,
				purple, path,
				statusStr,
			)

			return err
		}
	}
}

func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	codeName := "INTERNAL_ERROR"
	message := "An internal error occurred"
	var details []interface{}
	requestID := c.Response().Header().Get("X-Request-ID")
	tenantID := c.Get("tenant_id")

	if appErr, ok := err.(*errs.Error); ok {
		code = appErr.HTTPStatus()
		codeName = string(appErr.Code)
		message = appErr.Message
		if appErr.Details != nil {
			details = make([]interface{}, len(appErr.Details))
			for i, d := range appErr.Details {
				details[i] = map[string]string{
					"field":   d.Field,
					"message": d.Message,
				}
			}
		}
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		codeName = httpStatusToCodeName(code)
		if he.Message != nil {
			if m, ok := he.Message.(map[string]interface{}); ok {
				message = getStringFromMap(m, "message", message)
				if d, ok := m["details"].([]interface{}); ok {
					details = d
				}
			} else if s, ok := he.Message.(string); ok {
				message = s
			}
		}
	}

	c.JSON(code, map[string]interface{}{
		"error": map[string]interface{}{
			"code":       code,
			"code_name":  codeName,
			"message":    message,
			"details":    details,
			"request_id": requestID,
			"tenant_id":  tenantID,
		},
	})
}

func httpStatusToCodeName(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	default:
		return "INTERNAL_ERROR"
	}
}

func getStringFromMap(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
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
