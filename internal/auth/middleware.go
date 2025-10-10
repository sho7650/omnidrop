package auth

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"omnidrop/internal/observability"
)

// Middleware provides OAuth authentication middleware
type Middleware struct {
	jwtManager       *JWTManager
	logger           *slog.Logger
	legacyAuthEnabled bool
	legacyToken      string
}

// NewMiddleware creates a new OAuth middleware
func NewMiddleware(jwtManager *JWTManager, logger *slog.Logger) *Middleware {
	// Check for legacy auth support
	legacyEnabled := os.Getenv("OMNIDROP_LEGACY_AUTH_ENABLED") == "true"
	legacyToken := os.Getenv("TOKEN")

	if legacyEnabled && legacyToken != "" {
		logger.Warn("Legacy authentication is enabled - this should only be used during migration")
	}

	return &Middleware{
		jwtManager:        jwtManager,
		logger:            logger,
		legacyAuthEnabled: legacyEnabled,
		legacyToken:       legacyToken,
	}
}

// Authenticate is the main authentication middleware
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip OAuth endpoint itself
		if r.URL.Path == "/oauth/token" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip health and metrics endpoints (public)
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondUnauthorized(w, "Missing authorization header")
			return
		}

		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.respondUnauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Try OAuth JWT authentication first
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err == nil {
			// Valid OAuth token
			observability.TokenValidationTotal.WithLabelValues("success").Inc()

			// Store claims in context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)

			m.logger.Debug("OAuth authentication successful",
				slog.String("client_id", claims.ClientID),
				slog.Int("scopes_count", len(claims.Scopes)))

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// OAuth failed - try legacy authentication if enabled
		if m.legacyAuthEnabled && m.legacyToken != "" && tokenString == m.legacyToken {
			m.logger.Debug("Legacy authentication successful (migration mode)")

			// Create pseudo-claims for legacy token
			legacyClaims := &Claims{
				ClientID: "legacy",
				Scopes:   []string{"tasks:write", "files:write"}, // Default legacy scopes
			}

			ctx := context.WithValue(r.Context(), ContextKeyClaims, legacyClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Both OAuth and legacy auth failed
		observability.TokenValidationTotal.WithLabelValues("invalid").Inc()

		m.logger.Warn("Authentication failed",
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()))

		m.respondUnauthorized(w, "Invalid or expired token")
	})
}

// RequireScopes creates a middleware that requires specific scopes
func RequireScopes(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims from context
			claims, ok := r.Context().Value(ContextKeyClaims).(*Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if client has required scopes
			if !HasRequiredScopes(claims.Scopes, scopes) {
				observability.ScopeValidationFailures.WithLabelValues(claims.ClientID, strings.Join(scopes, ",")).Inc()

				slog.Warn("Insufficient permissions",
					slog.String("client_id", claims.ClientID),
					slog.String("required_scopes", strings.Join(scopes, ",")),
					slog.String("client_scopes", strings.Join(claims.Scopes, ",")))

				respondForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HasRequiredScopes checks if client scopes satisfy required scopes
func HasRequiredScopes(clientScopes []string, requiredScopes []string) bool {
	for _, required := range requiredScopes {
		matched := false
		for _, clientScope := range clientScopes {
			if MatchScope(clientScope, required) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// MatchScope matches a client scope against a required scope
// Supports exact match and wildcard matching (e.g., "automation:*" matches "automation:video-convert")
func MatchScope(clientScope, required string) bool {
	// Exact match
	if clientScope == required {
		return true
	}

	// Wildcard match: "automation:*" matches "automation:video-convert"
	if strings.HasSuffix(clientScope, ":*") {
		prefix := strings.TrimSuffix(clientScope, "*")
		return strings.HasPrefix(required, prefix)
	}

	// Wildcard match: "*" matches everything (admin scope)
	if clientScope == "*" {
		return true
	}

	return false
}

// respondUnauthorized sends an unauthorized response
func (m *Middleware) respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="omnidrop"`)
	http.Error(w, message, http.StatusUnauthorized)
}

// respondForbidden sends a forbidden response
func respondForbidden(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusForbidden)
}

// GetClaims extracts OAuth claims from request context
func GetClaims(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(ContextKeyClaims).(*Claims)
	return claims, ok
}
