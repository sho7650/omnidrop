package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	sloghttp "github.com/samber/slog-http"
)

// HTTPLogging returns a middleware that logs HTTP requests using structured logging.
// /health and /metrics are skipped to keep poller traffic out of the log.
func HTTPLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return sloghttp.NewWithConfig(logger, sloghttp.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
		WithRequestID:    true,
		Filters: []sloghttp.Filter{
			sloghttp.IgnorePath("/health", "/metrics"),
		},
	})
}

// RequestIDMiddleware adds a unique X-Request-ID response header for debugging.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", generateRequestID())
		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a unique request ID using UUID
func generateRequestID() string {
	return uuid.New().String()
}
