package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	sloghttp "github.com/samber/slog-http"
)

// LoggingConfig holds configuration for HTTP logging middleware
type LoggingConfig struct {
	Logger         *slog.Logger
	WithRequestID  bool
	WithBody       bool
	WithHeaders    bool
	SkipPaths      []string
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level
}

// DefaultLoggingConfig returns sensible defaults for HTTP logging
func DefaultLoggingConfig(logger *slog.Logger) LoggingConfig {
	return LoggingConfig{
		Logger:           logger,
		WithRequestID:    true,
		WithBody:         false, // Disable by default for security
		WithHeaders:      false, // Disable by default for security
		SkipPaths:        []string{"/health", "/metrics"},
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
	}
}

// HTTPLogging returns a middleware that logs HTTP requests using structured logging
func HTTPLogging(cfg LoggingConfig) func(http.Handler) http.Handler {
	config := sloghttp.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: cfg.ClientErrorLevel,
		ServerErrorLevel: cfg.ServerErrorLevel,

		WithRequestID:      cfg.WithRequestID,
		WithRequestBody:    cfg.WithBody,
		WithRequestHeader:  cfg.WithHeaders,
		WithResponseBody:   cfg.WithBody,
		WithResponseHeader: cfg.WithHeaders,

		Filters: []sloghttp.Filter{
			// Skip health check and metrics endpoints
			sloghttp.IgnorePath(cfg.SkipPaths...),
		},
	}

	return sloghttp.NewWithConfig(cfg.Logger, config)
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		
		// Add request ID to response header for debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "request_id", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a unique request ID using UUID
func generateRequestID() string {
	return uuid.New().String()
}