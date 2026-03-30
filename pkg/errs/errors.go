package errs

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Code string

const (
	CodeNotFound         Code = "NOT_FOUND"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeConflict         Code = "CONFLICT"
	CodeUnprocessable    Code = "UNPROCESSABLE_ENTITY"
	CodeInternalError    Code = "INTERNAL_ERROR"
	CodeTooManyRequests  Code = "TOO_MANY_REQUESTS"
	CodeValidationFailed Code = "VALIDATION_FAILED"
)

type Error struct {
	Code    Code     `json:"code"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
	Err     error    `json:"-"`
}

type Detail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type PermissionDeniedDetail struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) WithDetails(details ...Detail) *Error {
	e.Details = append(e.Details, details...)
	return e
}

func (e *Error) WithError(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeValidationFailed:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (e *Error) HTTPError(c echo.Context) error {
	return c.JSON(e.HTTPStatus(), e)
}

var (
	ErrNotFound        = &Error{Code: CodeNotFound, Message: "Resource not found"}
	ErrUnauthorized    = &Error{Code: CodeUnauthorized, Message: "Authentication required"}
	ErrForbidden       = &Error{Code: CodeForbidden, Message: "You don't have permission to perform this action"}
	ErrConflict        = &Error{Code: CodeConflict, Message: "Resource already exists"}
	ErrInternal        = &Error{Code: CodeInternalError, Message: "An internal error occurred"}
	ErrTooManyRequests = &Error{Code: CodeTooManyRequests, Message: "Rate limit exceeded"}
)

func NotFound(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

func Unauthorized(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeUnauthorized,
		Message: fmt.Sprintf(format, args...),
	}
}

func Forbidden(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeForbidden,
		Message: fmt.Sprintf(format, args...),
	}
}

func BadRequest(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeBadRequest,
		Message: fmt.Sprintf(format, args...),
	}
}

func Conflict(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeConflict,
		Message: fmt.Sprintf(format, args...),
	}
}

func Unprocessable(format string, args ...interface{}) *Error {
	return &Error{
		Code:    CodeUnprocessable,
		Message: fmt.Sprintf(format, args...),
	}
}

func Internal(err error) *Error {
	return &Error{
		Code:    CodeInternalError,
		Message: "An internal error occurred",
		Err:     err,
	}
}

func ValidationFailed(details ...Detail) *Error {
	return &Error{
		Code:    CodeValidationFailed,
		Message: "Validation failed",
		Details: details,
	}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func PermissionDenied(resource, action string) *Error {
	return &Error{
		Code:    CodeForbidden,
		Message: fmt.Sprintf("You don't have permission to %s %s", action, resource),
		Details: []Detail{
			{Field: "resource", Message: resource},
			{Field: "action", Message: action},
		},
	}
}

func ResourceNotFound(resource string, id string) *Error {
	return &Error{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: []Detail{
			{Field: "resource", Message: resource},
			{Field: "id", Message: id},
		},
	}
}
