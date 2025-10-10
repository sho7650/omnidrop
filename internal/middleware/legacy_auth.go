package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
)

// LegacyAuthMiddleware provides token-based authentication for legacy deployments
type LegacyAuthMiddleware struct {
	token  string
	logger *slog.Logger
}

// NewLegacyAuthMiddleware creates a new legacy authentication middleware
func NewLegacyAuthMiddleware(token string, logger *slog.Logger) *LegacyAuthMiddleware {
	return &LegacyAuthMiddleware{
		token:  token,
		logger: logger,
	}
}

// Authenticate validates the bearer token from the Authorization header
func (m *LegacyAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondUnauthorized(w, r, "Missing authorization header")
			return
		}

		// Validate Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.respondUnauthorized(w, r, "Invalid authorization header format")
			return
		}

		// Extract token
		providedToken := strings.TrimPrefix(authHeader, "Bearer ")

		// Constant-time comparison to prevent timing attacks
		if !tokensEqual(providedToken, m.token) {
			m.respondUnauthorized(w, r, "Invalid authentication token")
			return
		}

		// Authentication successful
		m.logger.Debug("Legacy authentication successful",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method))

		next.ServeHTTP(w, r)
	})
}

// respondUnauthorized sends a 401 response with appropriate headers
func (m *LegacyAuthMiddleware) respondUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	m.logger.Warn("Legacy authentication failed",
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("reason", message))

	w.Header().Set("WWW-Authenticate", `Bearer realm="omnidrop"`)
	http.Error(w, message, http.StatusUnauthorized)
}

// tokensEqual performs constant-time comparison of two tokens to prevent timing attacks
func tokensEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
