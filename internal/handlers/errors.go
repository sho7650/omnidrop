package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/pkg/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ErrorCode represents different types of application errors
type ErrorCode string

const (
	ErrorCodeValidation       ErrorCode = "validation_error"
	ErrorCodeAuthentication   ErrorCode = "authentication_error"
	ErrorCodeInternal         ErrorCode = "internal_error"
	ErrorCodeNotFound         ErrorCode = "not_found"
	ErrorCodeMethodNotAllowed ErrorCode = "method_not_allowed"
	ErrorCodeAppleScript      ErrorCode = "applescript_error"
)

// writeErrorResponse writes a standardized error response with proper logging
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode ErrorCode, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    string(errorCode),
	}

	// Log the error with stack trace if available
	if err != nil {
		if stackErr, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
			slog.Error("❌ Error occurred",
				slog.String("code", string(errorCode)),
				slog.String("message", message),
				slog.Any("stack_trace", stackErr.StackTrace()))
		} else {
			slog.Error("❌ Error occurred",
				slog.String("code", string(errorCode)),
				slog.String("message", message),
				slog.String("error", err.Error()))
		}
	} else {
		slog.Error("❌ Error occurred",
			slog.String("code", string(errorCode)),
			slog.String("message", message))
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("❌ Failed to encode error response", slog.String("error", err.Error()))
	}
}

// writeValidationError writes a validation error response
func writeValidationError(w http.ResponseWriter, message string) {
	writeErrorResponse(w, http.StatusBadRequest, ErrorCodeValidation, message, nil)
}

// writeAuthenticationError writes an authentication error response
func writeAuthenticationError(w http.ResponseWriter, message string) {
	writeErrorResponse(w, http.StatusUnauthorized, ErrorCodeAuthentication, message, nil)
}

// writeInternalError writes an internal server error response
func writeInternalError(w http.ResponseWriter, message string, err error) {
	writeErrorResponse(w, http.StatusInternalServerError, ErrorCodeInternal, message, err)
}

// writeMethodNotAllowedError writes a method not allowed error response
func writeMethodNotAllowedError(w http.ResponseWriter, message string) {
	writeErrorResponse(w, http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed, message, nil)
}

// writeAppleScriptError writes an AppleScript-specific error response
func writeAppleScriptError(w http.ResponseWriter, message string, err error) {
	writeErrorResponse(w, http.StatusInternalServerError, ErrorCodeAppleScript, message, err)
}
