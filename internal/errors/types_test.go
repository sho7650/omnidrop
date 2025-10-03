package errors

import (
	"errors"
	"log/slog"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		expected string
	}{
		{
			name:     "error without cause",
			err:      NewValidationError("invalid input"),
			expected: "validation_error: invalid input",
		},
		{
			name:     "error with cause",
			err:      NewInternalError("database error").WithCause(errors.New("connection failed")),
			expected: "internal_error: database error (caused by: connection failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewInternalError("wrapper").WithCause(cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestDomainError_WithContext(t *testing.T) {
	err := NewValidationError("invalid field").
		WithContext("field", "email").
		WithContext("value", "invalid@")

	if err.Context["field"] != "email" {
		t.Errorf("Context[field] = %v, want email", err.Context["field"])
	}
	if err.Context["value"] != "invalid@" {
		t.Errorf("Context[value] = %v, want invalid@", err.Context["value"])
	}
}

func TestDomainError_WithContextMap(t *testing.T) {
	ctx := map[string]any{
		"user_id":    123,
		"request_id": "req-456",
	}

	err := NewAuthenticationError("invalid token").WithContextMap(ctx)

	if err.Context["user_id"] != 123 {
		t.Errorf("Context[user_id] = %v, want 123", err.Context["user_id"])
	}
	if err.Context["request_id"] != "req-456" {
		t.Errorf("Context[request_id] = %v, want req-456", err.Context["request_id"])
	}
}

func TestDomainError_LogValue(t *testing.T) {
	err := NewFileSystemError("failed to write file").
		WithCause(errors.New("disk full")).
		WithContext("filename", "test.txt").
		WithContext("size", 1024)

	logValue := err.LogValue()

	// Verify it returns a slog.Value
	if logValue.Kind() != slog.KindGroup {
		t.Errorf("LogValue().Kind() = %v, want %v", logValue.Kind(), slog.KindGroup)
	}

	// Test that LogValue can be used with slog
	// This ensures the implementation is correct
	_ = slog.Any("error", err)
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		constructor func() *DomainError
		wantCode   ErrorCode
		wantStatus int
	}{
		{
			name:       "NewValidationError",
			constructor: func() *DomainError { return NewValidationError("test") },
			wantCode:   ErrorCodeValidation,
			wantStatus: 400,
		},
		{
			name:       "NewAuthenticationError",
			constructor: func() *DomainError { return NewAuthenticationError("test") },
			wantCode:   ErrorCodeAuthentication,
			wantStatus: 401,
		},
		{
			name:       "NewAuthorizationError",
			constructor: func() *DomainError { return NewAuthorizationError("test") },
			wantCode:   ErrorCodeAuthorization,
			wantStatus: 403,
		},
		{
			name:       "NewNotFoundError",
			constructor: func() *DomainError { return NewNotFoundError("resource") },
			wantCode:   ErrorCodeNotFound,
			wantStatus: 404,
		},
		{
			name:       "NewAlreadyExistsError",
			constructor: func() *DomainError { return NewAlreadyExistsError("resource") },
			wantCode:   ErrorCodeAlreadyExists,
			wantStatus: 409,
		},
		{
			name:       "NewAppleScriptError",
			constructor: func() *DomainError { return NewAppleScriptError("test") },
			wantCode:   ErrorCodeAppleScript,
			wantStatus: 500,
		},
		{
			name:       "NewFileSystemError",
			constructor: func() *DomainError { return NewFileSystemError("test") },
			wantCode:   ErrorCodeFileSystem,
			wantStatus: 500,
		},
		{
			name:       "NewInternalError",
			constructor: func() *DomainError { return NewInternalError("test") },
			wantCode:   ErrorCodeInternal,
			wantStatus: 500,
		},
		{
			name:       "NewMethodNotAllowedError",
			constructor: func() *DomainError { return NewMethodNotAllowedError("GET") },
			wantCode:   ErrorCodeMethodNotAllowed,
			wantStatus: 405,
		},
		{
			name:       "NewTimeoutError",
			constructor: func() *DomainError { return NewTimeoutError("operation") },
			wantCode:   ErrorCodeTimeout,
			wantStatus: 504,
		},
		{
			name:       "NewRateLimitError",
			constructor: func() *DomainError { return NewRateLimitError("test") },
			wantCode:   ErrorCodeRateLimit,
			wantStatus: 429,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, tt.wantStatus)
			}
			if len(err.StackTrace) == 0 {
				t.Error("StackTrace is empty, want non-empty")
			}
		})
	}
}

func TestCaptureStackTrace(t *testing.T) {
	err := NewValidationError("test error")

	if len(err.StackTrace) == 0 {
		t.Fatal("StackTrace is empty, want non-empty")
	}

	// Verify stack trace contains function names
	for _, frame := range err.StackTrace {
		if frame.Function == "" {
			t.Error("StackFrame.Function is empty")
		}
		if frame.File == "" {
			t.Error("StackFrame.File is empty")
		}
		if frame.Line == 0 {
			t.Error("StackFrame.Line is 0")
		}
	}
}
