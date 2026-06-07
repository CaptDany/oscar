package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantCode Code
		wantMsg  string
	}{
		{"NotFound", NotFound("user %s not found", "123"), CodeNotFound, "user 123 not found"},
		{"Unauthorized", Unauthorized("invalid token"), CodeUnauthorized, "invalid token"},
		{"Forbidden", Forbidden("access denied"), CodeForbidden, "access denied"},
		{"BadRequest", BadRequest("invalid input"), CodeBadRequest, "invalid input"},
		{"Conflict", Conflict("duplicate key"), CodeConflict, "duplicate key"},
		{"Unprocessable", Unprocessable("cannot process"), CodeUnprocessable, "cannot process"},
		{"Internal", Internal(errors.New("db error")), CodeInternalError, "An internal error occurred"},
		{"ValidationFailed", ValidationFailed(Detail{Field: "email", Message: "required"}), CodeValidationFailed, "Validation failed"},
		{"PermissionDenied", PermissionDenied("persons", "edit"), CodeForbidden, "You don't have permission to edit persons"},
		{"ResourceNotFound", ResourceNotFound("user", "abc"), CodeNotFound, "user not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("got code %q, want %q", tt.err.Code, tt.wantCode)
			}
			if tt.err.Message != tt.wantMsg {
				t.Errorf("got message %q, want %q", tt.err.Message, tt.wantMsg)
			}
		})
	}
}

func TestErrorHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantCode int
	}{
		{"NotFound", NotFound("x"), http.StatusNotFound},
		{"Unauthorized", Unauthorized("x"), http.StatusUnauthorized},
		{"Forbidden", Forbidden("x"), http.StatusForbidden},
		{"BadRequest", BadRequest("x"), http.StatusBadRequest},
		{"Conflict", Conflict("x"), http.StatusConflict},
		{"Unprocessable", Unprocessable("x"), http.StatusUnprocessableEntity},
		{"TooManyRequests", &Error{Code: CodeTooManyRequests}, http.StatusTooManyRequests},
		{"Internal", Internal(nil), http.StatusInternalServerError},
		{"Default", &Error{Code: "UNKNOWN"}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.HTTPStatus(); got != tt.wantCode {
				t.Errorf("%s.HTTPStatus() = %d, want %d", tt.name, got, tt.wantCode)
			}
		})
	}
}

func TestErrorWithDetails(t *testing.T) {
	err := BadRequest("validation error").WithDetails(
		Detail{Field: "email", Message: "required"},
		Detail{Field: "name", Message: "min=1"},
	)

	if len(err.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(err.Details))
	}
	if err.Details[0].Field != "email" {
		t.Errorf("expected field 'email', got %q", err.Details[0].Field)
	}
	if err.Details[1].Field != "name" {
		t.Errorf("expected field 'name', got %q", err.Details[1].Field)
	}
}

func TestErrorWithError(t *testing.T) {
	original := errors.New("original error")
	err := Internal(original)

	if !errors.Is(err.Err, original) {
		t.Error("wrapped error should contain original")
	}
}

func TestErrorUnwrap(t *testing.T) {
	original := errors.New("underlying error")
	err := Internal(original)

	if !errors.Is(err, original) {
		t.Error("errors.Is should detect wrapped error")
	}

	var target *Error
	if !errors.As(err, &target) {
		t.Error("errors.As should match *Error")
	}
}

func TestSentinelError(t *testing.T) {
	if ErrNotFound.Code != CodeNotFound {
		t.Errorf("ErrNotFound.Code = %q, want %q", ErrNotFound.Code, CodeNotFound)
	}
	if ErrUnauthorized.Code != CodeUnauthorized {
		t.Errorf("ErrUnauthorized.Code = %q, want %q", ErrUnauthorized.Code, CodeUnauthorized)
	}
	if ErrForbidden.Code != CodeForbidden {
		t.Errorf("ErrForbidden.Code = %q, want %q", ErrForbidden.Code, CodeForbidden)
	}
	if ErrConflict.Code != CodeConflict {
		t.Errorf("ErrConflict.Code = %q, want %q", ErrConflict.Code, CodeConflict)
	}
	if ErrInternal.Code != CodeInternalError {
		t.Errorf("ErrInternal.Code = %q, want %q", ErrInternal.Code, CodeInternalError)
	}
	if ErrTooManyRequests.Code != CodeTooManyRequests {
		t.Errorf("ErrTooManyRequests.Code = %q, want %q", ErrTooManyRequests.Code, CodeTooManyRequests)
	}
}

func TestErrorString(t *testing.T) {
	err := BadRequest("invalid input")
	expected := "BAD_REQUEST: invalid input"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	errWithCause := Internal(errors.New("db failure"))
	if errWithCause.Error() != "INTERNAL_ERROR: An internal error occurred: db failure" {
		t.Errorf("unexpected error string: %q", errWithCause.Error())
	}
}

func TestPermissionDenied(t *testing.T) {
	err := PermissionDenied("deals", "delete")
	if err.Code != CodeForbidden {
		t.Errorf("code = %q, want %q", err.Code, CodeForbidden)
	}
	if len(err.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(err.Details))
	}
	if err.Details[0].Field != "resource" || err.Details[0].Message != "deals" {
		t.Errorf("unexpected first detail: %+v", err.Details[0])
	}
	if err.Details[1].Field != "action" || err.Details[1].Message != "delete" {
		t.Errorf("unexpected second detail: %+v", err.Details[1])
	}
}
