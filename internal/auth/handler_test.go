package auth

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTokenHandler tests handler creation
func TestNewTokenHandler(t *testing.T) {
	repo, cleanup := createTestRepository(t, []OAuthClient{})
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.repository)
	assert.NotNil(t, handler.jwtManager)
	assert.Equal(t, time.Hour, handler.tokenExpiry)
	assert.NotNil(t, handler.logger)
}

// TestTokenHandler_HandleToken_JSON tests token request with JSON body
func TestTokenHandler_HandleToken_JSON(t *testing.T) {
	// Create test client with hashed password
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "test-client",
			ClientSecretHash: hashedSecret,
			Name:             "Test Client",
			Scopes:           []string{"tasks:write", "files:read"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedError  string
		checkToken     bool
	}{
		{
			name: "valid request returns token",
			body: map[string]string{
				"grant_type":    "client_credentials",
				"client_id":     "test-client",
				"client_secret": "test-secret",
			},
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name: "unsupported grant_type returns error",
			body: map[string]string{
				"grant_type":    "password",
				"client_id":     "test-client",
				"client_secret": "test-secret",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrorUnsupportedGrantType,
			checkToken:     false,
		},
		{
			name: "missing client_id returns error",
			body: map[string]string{
				"grant_type":    "client_credentials",
				"client_secret": "test-secret",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrorInvalidRequest,
			checkToken:     false,
		},
		{
			name: "missing client_secret returns error",
			body: map[string]string{
				"grant_type": "client_credentials",
				"client_id":  "test-client",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrorInvalidRequest,
			checkToken:     false,
		},
		{
			name: "invalid client_id returns error",
			body: map[string]string{
				"grant_type":    "client_credentials",
				"client_id":     "non-existent-client",
				"client_secret": "test-secret",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrorInvalidClient,
			checkToken:     false,
		},
		{
			name: "wrong client_secret returns error",
			body: map[string]string{
				"grant_type":    "client_credentials",
				"client_id":     "test-client",
				"client_secret": "wrong-secret",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrorInvalidClient,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleToken(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))

			if tt.checkToken {
				var response TokenResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.AccessToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, int64(3600), response.ExpiresIn) // 1 hour
				assert.Contains(t, response.Scope, "tasks:write")
				assert.Contains(t, response.Scope, "files:read")
			}

			if tt.expectedError != "" {
				var errResponse ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errResponse.Error)
			}
		})
	}
}

// TestTokenHandler_HandleToken_FormURLEncoded tests token request with form data
func TestTokenHandler_HandleToken_FormURLEncoded(t *testing.T) {
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "test-client",
			ClientSecretHash: hashedSecret,
			Name:             "Test Client",
			Scopes:           []string{"tasks:write"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", "test-client")
	form.Set("client_secret", "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.HandleToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response TokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(3600), response.ExpiresIn)
}

// TestTokenHandler_HandleToken_InvalidJSON tests handling of malformed JSON
func TestTokenHandler_HandleToken_InvalidJSON(t *testing.T) {
	repo, cleanup := createTestRepository(t, []OAuthClient{})
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "malformed JSON",
			body: `{"grant_type": "client_credentials", "client_id": `,
		},
		{
			name: "invalid JSON structure",
			body: `[1, 2, 3]`,
		},
		{
			name: "empty body",
			body: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleToken(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResponse ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
			require.NoError(t, err)
			assert.Equal(t, ErrorInvalidRequest, errResponse.Error)
		})
	}
}

// TestTokenHandler_HandleToken_DisabledClient tests that disabled clients cannot get tokens
func TestTokenHandler_HandleToken_DisabledClient(t *testing.T) {
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "disabled-client",
			ClientSecretHash: hashedSecret,
			Name:             "Disabled Client",
			Scopes:           []string{"tasks:write"},
			Disabled:         true,
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	body := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "disabled-client",
		"client_secret": "test-secret",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResponse ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errResponse)
	require.NoError(t, err)
	assert.Equal(t, ErrorInvalidClient, errResponse.Error)
}

// TestTokenHandler_HandleToken_TokenValidity tests that returned token is valid
func TestTokenHandler_HandleToken_TokenValidity(t *testing.T) {
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "test-client",
			ClientSecretHash: hashedSecret,
			Name:             "Test Client",
			Scopes:           []string{"tasks:write", "files:read"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	body := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     "test-client",
		"client_secret": "test-secret",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleToken(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response TokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate the returned token
	claims, err := jm.ValidateToken(response.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, "test-client", claims.ClientID)
	assert.Contains(t, claims.Scopes, "tasks:write")
	assert.Contains(t, claims.Scopes, "files:read")
}

// TestTokenHandler_HandleToken_MultipleClients tests with multiple clients
func TestTokenHandler_HandleToken_MultipleClients(t *testing.T) {
	clients := []OAuthClient{
		{
			ClientID:         "client-a",
			ClientSecretHash: hashPassword(t, "secret-a"),
			Name:             "Client A",
			Scopes:           []string{"scope-a"},
		},
		{
			ClientID:         "client-b",
			ClientSecretHash: hashPassword(t, "secret-b"),
			Name:             "Client B",
			Scopes:           []string{"scope-b"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	jm := newTestJWTManager()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTokenHandler(repo, jm, time.Hour, logger)

	tests := []struct {
		clientID     string
		clientSecret string
		expectedOK   bool
	}{
		{"client-a", "secret-a", true},
		{"client-b", "secret-b", true},
		{"client-a", "secret-b", false}, // wrong secret
		{"client-b", "secret-a", false}, // wrong secret
	}

	for _, tt := range tests {
		t.Run(tt.clientID+"_with_"+tt.clientSecret, func(t *testing.T) {
			body := map[string]string{
				"grant_type":    "client_credentials",
				"client_id":     tt.clientID,
				"client_secret": tt.clientSecret,
			}
			bodyJSON, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleToken(rec, req)

			if tt.expectedOK {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

// createTestRepository creates a temporary repository for testing
func createTestRepository(t *testing.T, clients []OAuthClient) (*Repository, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "oauth-clients.yaml")

	// Create config file with test clients
	config := OAuthConfig{
		Clients: clients,
	}

	configData, err := yamlMarshal(config)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err)

	// Create repository
	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}

	return repo, cleanup
}

// yamlMarshal is a helper to marshal YAML
func yamlMarshal(v interface{}) ([]byte, error) {
	// Simple YAML marshaling for test config
	switch c := v.(type) {
	case OAuthConfig:
		return marshalOAuthConfig(c), nil
	}
	return nil, nil
}

// marshalOAuthConfig manually marshals OAuthConfig to avoid import cycle in tests
func marshalOAuthConfig(config OAuthConfig) []byte {
	var b strings.Builder
	b.WriteString("clients:\n")
	for _, client := range config.Clients {
		b.WriteString("  - client_id: " + client.ClientID + "\n")
		b.WriteString("    client_secret_hash: " + client.ClientSecretHash + "\n")
		b.WriteString("    name: " + client.Name + "\n")
		if len(client.Scopes) > 0 {
			b.WriteString("    scopes:\n")
			for _, scope := range client.Scopes {
				b.WriteString("      - " + scope + "\n")
			}
		}
		if client.Disabled {
			b.WriteString("    disabled: true\n")
		}
	}
	return []byte(b.String())
}
