package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"omnidrop/internal/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// writeErrorResponse writes a standardized error response with proper logging
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode errors.ErrorCode, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    string(errorCode),
	}

	// Log the error using slog.LogValuer if it's a DomainError
	if domainErr, ok := err.(*errors.DomainError); ok {
		slog.Error("❌ Error occurred", slog.Any("error", domainErr))
	} else if err != nil {
		slog.Error("❌ Error occurred",
			slog.String("code", string(errorCode)),
			slog.String("message", message),
			slog.String("error", err.Error()))
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
	writeErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeValidation, message, nil)
}

// writeMethodNotAllowedError writes a method not allowed error response
func writeMethodNotAllowedError(w http.ResponseWriter, message string) {
	writeErrorResponse(w, http.StatusMethodNotAllowed, errors.ErrorCodeMethodNotAllowed, message, nil)
}

// writeAppleScriptError writes an AppleScript-specific error response
func writeAppleScriptError(w http.ResponseWriter, message string, err error) {
	writeErrorResponse(w, http.StatusInternalServerError, errors.ErrorCodeAppleScript, message, err)
}
