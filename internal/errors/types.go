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
	// Validation errors
	ErrorCodeValidation ErrorCode = "validation_error"

	// Authentication and authorization errors
	ErrorCodeAuthentication ErrorCode = "authentication_error"
	ErrorCodeAuthorization  ErrorCode = "authorization_error"

	// Resource errors
	ErrorCodeNotFound      ErrorCode = "not_found"
	ErrorCodeAlreadyExists ErrorCode = "already_exists"
	ErrorCodeConflict      ErrorCode = "conflict"

	// Integration errors
	ErrorCodeAppleScript    ErrorCode = "applescript_error"
	ErrorCodeFileSystem     ErrorCode = "filesystem_error"
	ErrorCodeExternalSystem ErrorCode = "external_system_error"

	// System errors
	ErrorCodeInternal         ErrorCode = "internal_error"
	ErrorCodeMethodNotAllowed ErrorCode = "method_not_allowed"
	ErrorCodeTimeout          ErrorCode = "timeout_error"
	ErrorCodeRateLimit        ErrorCode = "rate_limit_exceeded"
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

	// Add cause if present
	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}

	// Add context fields
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

// WithContextMap adds multiple context key-value pairs
func (e *DomainError) WithContextMap(ctx map[string]any) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	for k, v := range ctx {
		e.Context[k] = v
	}
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

// Common error constructors for convenience

// NewValidationError creates a validation error
func NewValidationError(message string) *DomainError {
	return NewDomainError(ErrorCodeValidation, message, 400)
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *DomainError {
	return NewDomainError(ErrorCodeAuthentication, message, 401)
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *DomainError {
	return NewDomainError(ErrorCodeAuthorization, message, 403)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *DomainError {
	return NewDomainError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), 404)
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource string) *DomainError {
	return NewDomainError(ErrorCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), 409)
}

// NewAppleScriptError creates an AppleScript error
func NewAppleScriptError(message string) *DomainError {
	return NewDomainError(ErrorCodeAppleScript, message, 500)
}

// NewFileSystemError creates a filesystem error
func NewFileSystemError(message string) *DomainError {
	return NewDomainError(ErrorCodeFileSystem, message, 500)
}

// NewInternalError creates an internal error
func NewInternalError(message string) *DomainError {
	return NewDomainError(ErrorCodeInternal, message, 500)
}

// NewMethodNotAllowedError creates a method not allowed error
func NewMethodNotAllowedError(method string) *DomainError {
	return NewDomainError(ErrorCodeMethodNotAllowed, fmt.Sprintf("method %s not allowed", method), 405)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *DomainError {
	return NewDomainError(ErrorCodeTimeout, fmt.Sprintf("%s timed out", operation), 504)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *DomainError {
	return NewDomainError(ErrorCodeRateLimit, message, 429)
}
