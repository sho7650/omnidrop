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
			err:      NewInternalError("boom"),
			expected: "internal_error: boom",
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
	err := NewInternalError("invalid field").
		WithContext("field", "email").
		WithContext("value", "invalid@")

	if err.Context["field"] != "email" {
		t.Errorf("Context[field] = %v, want email", err.Context["field"])
	}
	if err.Context["value"] != "invalid@" {
		t.Errorf("Context[value] = %v, want invalid@", err.Context["value"])
	}
}

func TestDomainError_LogValue(t *testing.T) {
	err := NewInternalError("failed to write file").
		WithCause(errors.New("disk full")).
		WithContext("filename", "test.txt").
		WithContext("size", 1024)

	logValue := err.LogValue()

	if logValue.Kind() != slog.KindGroup {
		t.Errorf("LogValue().Kind() = %v, want %v", logValue.Kind(), slog.KindGroup)
	}

	// Ensure LogValue plugs into slog without panicking
	_ = slog.Any("error", err)
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("boom")

	if err.Code != ErrorCodeInternal {
		t.Errorf("Code = %v, want %v", err.Code, ErrorCodeInternal)
	}
	if err.HTTPStatus != 500 {
		t.Errorf("HTTPStatus = %d, want 500", err.HTTPStatus)
	}
	if len(err.StackTrace) == 0 {
		t.Error("StackTrace is empty, want non-empty")
	}
}

func TestCaptureStackTrace(t *testing.T) {
	err := NewInternalError("test error")

	if len(err.StackTrace) == 0 {
		t.Fatal("StackTrace is empty, want non-empty")
	}

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
