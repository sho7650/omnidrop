package middleware

import (
	"net/http"
	"strconv"
	"time"

	"omnidrop/internal/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Metrics returns a middleware that collects HTTP metrics for Prometheus
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default status
			bytesWritten:   0,
		}

		// Get request size
		requestSize := r.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get route pattern from chi context
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		method := r.Method
		status := strconv.Itoa(wrapped.statusCode)

		// Record metrics
		observability.HTTPRequestsTotal.WithLabelValues(method, routePattern, status).Inc()
		observability.HTTPRequestDuration.WithLabelValues(method, routePattern, status).Observe(duration)
		observability.HTTPRequestSize.WithLabelValues(method, routePattern).Observe(float64(requestSize))
		observability.HTTPResponseSize.WithLabelValues(method, routePattern).Observe(float64(wrapped.bytesWritten))
	})
}

// WrapHandler wraps chi middleware.WrapResponseWriter with our responseWriter
func WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapped, r)
	})
}