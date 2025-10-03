package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"omnidrop/internal/errors"
)

// Recovery returns a middleware that recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Capture panic details
				panicErr := fmt.Errorf("panic: %v", rvr)
				stack := string(debug.Stack())

				// Create domain error for panic
				domainErr := errors.NewInternalError("internal server error").
					WithCause(panicErr).
					WithContext("panic_value", rvr).
					WithContext("request_method", r.Method).
					WithContext("request_path", r.URL.Path).
					WithContext("request_remote_addr", r.RemoteAddr)

				// Log panic with full stack trace
				slog.Error("🚨 Panic recovered",
					slog.Any("error", domainErr),
					slog.String("stack_trace", stack),
				)

				// Return error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status":"error","message":"internal server error","code":"internal_error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
