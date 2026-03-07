package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMiddleware tests middleware creation
func TestNewMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		legacyEnabled     string
		legacyToken       string
		expectLegacy      bool
	}{
		{
			name:          "creates middleware without legacy auth",
			legacyEnabled: "",
			legacyToken:   "",
			expectLegacy:  false,
		},
		{
			name:          "creates middleware with legacy auth enabled",
			legacyEnabled: "true",
			legacyToken:   "test-legacy-token",
			expectLegacy:  true,
		},
		{
			name:          "legacy auth disabled when flag is false",
			legacyEnabled: "false",
			legacyToken:   "test-legacy-token",
			expectLegacy:  false,
		},
		{
			name:          "legacy auth disabled when token is empty",
			legacyEnabled: "true",
			legacyToken:   "",
			expectLegacy:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.legacyEnabled != "" {
				t.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", tt.legacyEnabled)
			}
			if tt.legacyToken != "" {
				t.Setenv("TOKEN", tt.legacyToken)
			}

			jm := newTestJWTManager()
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

			m := NewMiddleware(jm, logger)

			assert.NotNil(t, m)
			assert.NotNil(t, m.jwtManager)
			assert.NotNil(t, m.logger)

			if tt.expectLegacy {
				assert.True(t, m.legacyAuthEnabled, "expected legacy auth to be enabled")
				assert.Equal(t, tt.legacyToken, m.legacyToken)
			} else {
				// Either disabled flag or empty token results in no legacy auth
				if tt.legacyEnabled != "true" {
					assert.False(t, m.legacyAuthEnabled)
				}
			}
		})
	}
}

// TestMiddleware_Authenticate_PublicEndpoints tests that public endpoints are skipped
func TestMiddleware_Authenticate_PublicEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "skips /oauth/token endpoint",
			path:     "/oauth/token",
			expected: http.StatusOK,
		},
		{
			name:     "skips /health endpoint",
			path:     "/health",
			expected: http.StatusOK,
		},
		{
			name:     "skips /metrics endpoint",
			path:     "/metrics",
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := newTestJWTManager()
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
			m := NewMiddleware(jm, logger)

			// Create test handler that returns OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request without any auth header
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			m.Authenticate(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expected, rec.Code, "public endpoint should be accessible without auth")
		})
	}
}

// TestMiddleware_Authenticate_OAuth tests OAuth authentication
func TestMiddleware_Authenticate_OAuth(t *testing.T) {
	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	m := NewMiddleware(jm, logger)

	client := newTestOAuthClient("test-client", []string{"tasks:write", "files:read"})
	validToken := generateValidToken(t, jm, client)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		checkClaims    bool
	}{
		{
			name:           "valid OAuth token succeeds",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			checkClaims:    true,
		},
		{
			name:           "missing Authorization header returns 401",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "invalid Authorization format returns 401",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "Bearer without token returns 401",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "invalid token returns 401",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "malformed JWT returns 401",
			authHeader:     "Bearer not.a.valid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedClaims *Claims

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkClaims {
					claims, ok := GetClaims(r)
					require.True(t, ok, "claims should be in context")
					capturedClaims = claims
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			m.Authenticate(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkClaims {
				require.NotNil(t, capturedClaims)
				assert.Equal(t, client.ClientID, capturedClaims.ClientID)
				assert.Equal(t, client.Scopes, capturedClaims.Scopes)
			}

			// Check WWW-Authenticate header on 401
			if tt.expectedStatus == http.StatusUnauthorized {
				assert.Contains(t, rec.Header().Get("WWW-Authenticate"), "Bearer")
			}
		})
	}
}

// TestMiddleware_Authenticate_ExpiredToken tests expired token handling
func TestMiddleware_Authenticate_ExpiredToken(t *testing.T) {
	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	m := NewMiddleware(jm, logger)

	client := newTestOAuthClient("test-client", []string{"tasks:write"})
	expiredToken := generateExpiredToken(t, jm, client)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()

	m.Authenticate(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "expired token should return 401")
	assert.Contains(t, rec.Header().Get("WWW-Authenticate"), "Bearer")
}

// TestMiddleware_Authenticate_LegacyAuth tests legacy authentication fallback
func TestMiddleware_Authenticate_LegacyAuth(t *testing.T) {
	tests := []struct {
		name           string
		legacyEnabled  bool
		legacyToken    string
		authToken      string
		expectedStatus int
		expectedScopes []string
	}{
		{
			name:           "legacy auth succeeds when enabled",
			legacyEnabled:  true,
			legacyToken:    testLegacyToken,
			authToken:      testLegacyToken,
			expectedStatus: http.StatusOK,
			expectedScopes: []string{"tasks:write", "files:write"},
		},
		{
			name:           "legacy auth fails when disabled",
			legacyEnabled:  false,
			legacyToken:    testLegacyToken,
			authToken:      testLegacyToken,
			expectedStatus: http.StatusUnauthorized,
			expectedScopes: nil,
		},
		{
			name:           "wrong legacy token fails",
			legacyEnabled:  true,
			legacyToken:    testLegacyToken,
			authToken:      "wrong-token",
			expectedStatus: http.StatusUnauthorized,
			expectedScopes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment for legacy auth
			if tt.legacyEnabled {
				t.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
			}
			t.Setenv("TOKEN", tt.legacyToken)

			jm := newTestJWTManager()
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
			m := NewMiddleware(jm, logger)

			var capturedClaims *Claims

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := GetClaims(r)
				if ok {
					capturedClaims = claims
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			req.Header.Set("Authorization", "Bearer "+tt.authToken)
			rec := httptest.NewRecorder()

			m.Authenticate(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				require.NotNil(t, capturedClaims)
				assert.Equal(t, "legacy", capturedClaims.ClientID)
				assert.Equal(t, tt.expectedScopes, capturedClaims.Scopes)
			}
		})
	}
}

// TestMiddleware_Authenticate_LegacyIsolation tests that OAuth takes precedence over legacy
func TestMiddleware_Authenticate_LegacyIsolation(t *testing.T) {
	t.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	t.Setenv("TOKEN", testLegacyToken)

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	m := NewMiddleware(jm, logger)

	// Create a valid OAuth token
	client := newTestOAuthClient("oauth-client", []string{"admin:*"})
	validToken := generateValidToken(t, jm, client)

	var capturedClaims *Claims

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := GetClaims(r)
		capturedClaims = claims
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rec := httptest.NewRecorder()

	m.Authenticate(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	// Should use OAuth, not legacy
	assert.Equal(t, "oauth-client", capturedClaims.ClientID, "OAuth should take precedence over legacy")
	assert.NotEqual(t, "legacy", capturedClaims.ClientID)
}

// TestMiddleware_Authenticate_HeaderInjection tests header injection protection
func TestMiddleware_Authenticate_HeaderInjection(t *testing.T) {
	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	m := NewMiddleware(jm, logger)

	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "newline injection in header",
			authHeader: "Bearer token\r\nX-Injected: value",
		},
		{
			name:       "null byte injection",
			authHeader: "Bearer token\x00malicious",
		},
		{
			name:       "excessive whitespace",
			authHeader: "Bearer   token   with   spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			m.Authenticate(nextHandler).ServeHTTP(rec, req)

			// All these should fail authentication (401)
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "header injection should not succeed")
		})
	}
}

// TestMiddleware_Authenticate_ContextIntegrity tests that failed auth doesn't pollute context
func TestMiddleware_Authenticate_ContextIntegrity(t *testing.T) {
	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	m := NewMiddleware(jm, logger)

	// Pre-populate context with existing claims (simulating a previous request)
	existingClaims := &Claims{ClientID: "existing-client", Scopes: []string{"old:scope"}}
	ctx := context.WithValue(context.Background(), ContextKeyClaims, existingClaims)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This shouldn't be called for invalid auth
		t.Error("next handler should not be called for invalid auth")
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	m.Authenticate(nextHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	// The original context should be unchanged (request was rejected)
}

// TestRequireScopes tests scope validation middleware
func TestRequireScopes(t *testing.T) {
	tests := []struct {
		name           string
		clientScopes   []string
		requiredScopes []string
		expectedStatus int
	}{
		{
			name:           "passes with exact matching scope",
			clientScopes:   []string{"tasks:write"},
			requiredScopes: []string{"tasks:write"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "passes with wildcard scope",
			clientScopes:   []string{"tasks:*"},
			requiredScopes: []string{"tasks:write"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "passes with admin wildcard",
			clientScopes:   []string{"*"},
			requiredScopes: []string{"tasks:write", "files:read"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "passes with multiple matching scopes",
			clientScopes:   []string{"tasks:write", "files:read", "admin:*"},
			requiredScopes: []string{"tasks:write", "files:read"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "fails with missing scope",
			clientScopes:   []string{"tasks:write"},
			requiredScopes: []string{"tasks:write", "files:read"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "fails with no matching scopes",
			clientScopes:   []string{"tasks:read"},
			requiredScopes: []string{"tasks:write"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "fails with empty client scopes",
			clientScopes:   []string{},
			requiredScopes: []string{"tasks:write"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				ClientID: "test-client",
				Scopes:   tt.clientScopes,
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			ctx := context.WithValue(req.Context(), ContextKeyClaims, claims)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler := RequireScopes(tt.requiredScopes...)(nextHandler)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRequireScopes_NoClaims tests RequireScopes when no claims in context
func TestRequireScopes_NoClaims(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	// No claims in context
	rec := httptest.NewRecorder()

	handler := RequireScopes("tasks:write")(nextHandler)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "should return 401 when no claims in context")
}

// TestGetClaims tests extracting claims from request context
func TestGetClaims(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(*http.Request) *http.Request
		expectClaims  bool
		expectedID    string
	}{
		{
			name: "extracts claims from valid context",
			setupContext: func(r *http.Request) *http.Request {
				claims := &Claims{ClientID: "test-client", Scopes: []string{"tasks:write"}}
				ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
				return r.WithContext(ctx)
			},
			expectClaims: true,
			expectedID:   "test-client",
		},
		{
			name: "returns false when no claims in context",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			expectClaims: false,
			expectedID:   "",
		},
		{
			name: "returns false when wrong type in context",
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), ContextKeyClaims, "not-a-claims-struct")
				return r.WithContext(ctx)
			},
			expectClaims: false,
			expectedID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			req = tt.setupContext(req)

			claims, ok := GetClaims(req)

			assert.Equal(t, tt.expectClaims, ok)
			if tt.expectClaims {
				require.NotNil(t, claims)
				assert.Equal(t, tt.expectedID, claims.ClientID)
			} else {
				assert.Nil(t, claims)
			}
		})
	}
}

