package errors

import (
	"fmt"
	"log/slog"
	"runtime"
)

// DomainError represents a domain-specific error with rich context for logging
type DomainError struct {
	Code       ErrorCode
	Message    string
	Cause      error
	HTTPStatus int
	Context    map[string]any
	StackTrace []StackFrame
}

// ErrorCode represents different types of application errors
type ErrorCode string

const (
	ErrorCodeValidation       ErrorCode = "validation_error"
	ErrorCodeAppleScript      ErrorCode = "applescript_error"
	ErrorCodeInternal         ErrorCode = "internal_error"
	ErrorCodeMethodNotAllowed ErrorCode = "method_not_allowed"
)

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error for error unwrapping
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// LogValue implements slog.LogValuer for rich structured logging
func (e *DomainError) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("code", string(e.Code)),
		slog.String("message", e.Message),
		slog.Int("http_status", e.HTTPStatus),
	}

	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}

	if len(e.Context) > 0 {
		contextAttrs := make([]slog.Attr, 0, len(e.Context))
		for k, v := range e.Context {
			contextAttrs = append(contextAttrs, slog.Any(k, v))
		}
		attrs = append(attrs, slog.Any("context", slog.GroupValue(contextAttrs...)))
	}

	// Add stack trace (first 5 frames for brevity)
	if len(e.StackTrace) > 0 {
		maxFrames := 5
		if len(e.StackTrace) < maxFrames {
			maxFrames = len(e.StackTrace)
		}

		stackAttrs := make([]slog.Attr, 0, maxFrames)
		for i := 0; i < maxFrames; i++ {
			frame := e.StackTrace[i]
			stackAttrs = append(stackAttrs, slog.Group(fmt.Sprintf("frame_%d", i),
				slog.String("function", frame.Function),
				slog.String("file", frame.File),
				slog.Int("line", frame.Line),
			))
		}
		attrs = append(attrs, slog.Any("stack_trace", slog.GroupValue(stackAttrs...)))
	}

	return slog.GroupValue(attrs...)
}

// NewDomainError creates a new DomainError with stack trace capture
func NewDomainError(code ErrorCode, message string, httpStatus int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Context:    make(map[string]any),
		StackTrace: captureStackTrace(2), // Skip NewDomainError and caller
	}
}

// WithCause adds a cause error
func (e *DomainError) WithCause(cause error) *DomainError {
	e.Cause = cause
	return e
}

// WithContext adds context key-value pairs
func (e *DomainError) WithContext(key string, value any) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []StackFrame {
	const maxFrames = 32
	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip+1, pcs)

	frames := make([]StackFrame, 0, n)
	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// NewInternalError creates an internal error
func NewInternalError(message string) *DomainError {
	return NewDomainError(ErrorCodeInternal, message, 500)
}
