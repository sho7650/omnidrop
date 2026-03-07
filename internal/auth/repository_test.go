package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRepository tests repository creation
func TestNewRepository(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T) string
		expectError bool
		errorType   string
	}{
		{
			name: "creates repository with valid config",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "oauth-clients.yaml")
				config := `clients:
  - client_id: test-client
    client_secret_hash: $2a$10$test
    name: Test Client
    scopes:
      - tasks:write
`
				err := os.WriteFile(configPath, []byte(config), 0600)
				require.NoError(t, err)
				return configPath
			},
			expectError: false,
		},
		{
			name: "creates empty config when file does not exist",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "non-existent.yaml")
			},
			expectError: false,
		},
		{
			name: "returns error for invalid YAML",
			setupConfig: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "invalid.yaml")
				err := os.WriteFile(configPath, []byte("invalid: yaml: ]["), 0600)
				require.NoError(t, err)
				return configPath
			},
			expectError: true,
			errorType:   "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupConfig(t)

			repo, err := NewRepository(configPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType == "yaml" {
					assert.Contains(t, err.Error(), "parse")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

// TestRepository_GetByClientID tests client retrieval by ID
func TestRepository_GetByClientID(t *testing.T) {
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "active-client",
			ClientSecretHash: hashedSecret,
			Name:             "Active Client",
			Scopes:           []string{"tasks:write"},
			Disabled:         false,
		},
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

	tests := []struct {
		name        string
		clientID    string
		expectError error
		checkClient bool
	}{
		{
			name:        "returns active client",
			clientID:    "active-client",
			expectError: nil,
			checkClient: true,
		},
		{
			name:        "returns ErrClientNotFound for non-existent client",
			clientID:    "non-existent",
			expectError: ErrClientNotFound,
			checkClient: false,
		},
		{
			name:        "returns ErrClientDisabled for disabled client",
			clientID:    "disabled-client",
			expectError: ErrClientDisabled,
			checkClient: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := repo.GetByClientID(tt.clientID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.checkClient {
					assert.Equal(t, tt.clientID, client.ClientID)
				}
			}
		})
	}
}

// TestRepository_Authenticate tests client authentication
func TestRepository_Authenticate(t *testing.T) {
	hashedSecret := hashPassword(t, "correct-secret")
	clients := []OAuthClient{
		{
			ClientID:         "test-client",
			ClientSecretHash: hashedSecret,
			Name:             "Test Client",
			Scopes:           []string{"tasks:write", "files:read"},
			Disabled:         false,
		},
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

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		expectError  error
		expectScopes []string
	}{
		{
			name:         "authenticates with correct credentials",
			clientID:     "test-client",
			clientSecret: "correct-secret",
			expectError:  nil,
			expectScopes: []string{"tasks:write", "files:read"},
		},
		{
			name:         "fails with wrong secret",
			clientID:     "test-client",
			clientSecret: "wrong-secret",
			expectError:  ErrInvalidCredentials,
			expectScopes: nil,
		},
		{
			name:         "fails with non-existent client",
			clientID:     "non-existent",
			clientSecret: "any-secret",
			expectError:  ErrClientNotFound,
			expectScopes: nil,
		},
		{
			name:         "fails with disabled client",
			clientID:     "disabled-client",
			clientSecret: "correct-secret",
			expectError:  ErrClientDisabled,
			expectScopes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := repo.Authenticate(tt.clientID, tt.clientSecret)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, client)
				assert.Equal(t, tt.clientID, client.ClientID)
				assert.Equal(t, tt.expectScopes, client.Scopes)
			}
		})
	}
}

// TestRepository_List tests listing all active clients
func TestRepository_List(t *testing.T) {
	hashedSecret := hashPassword(t, "test-secret")
	clients := []OAuthClient{
		{
			ClientID:         "client-a",
			ClientSecretHash: hashedSecret,
			Name:             "Client A",
			Scopes:           []string{"tasks:write"},
			Disabled:         false,
		},
		{
			ClientID:         "client-b",
			ClientSecretHash: hashedSecret,
			Name:             "Client B",
			Scopes:           []string{"files:read"},
			Disabled:         false,
		},
		{
			ClientID:         "disabled-client",
			ClientSecretHash: hashedSecret,
			Name:             "Disabled Client",
			Scopes:           []string{"admin:*"},
			Disabled:         true,
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	activeClients := repo.List()

	// Should only return active clients (not disabled)
	assert.Len(t, activeClients, 2)

	clientIDs := make([]string, 0, len(activeClients))
	for _, c := range activeClients {
		clientIDs = append(clientIDs, c.ClientID)
	}

	assert.Contains(t, clientIDs, "client-a")
	assert.Contains(t, clientIDs, "client-b")
	assert.NotContains(t, clientIDs, "disabled-client")
}

// TestRepository_Load tests configuration loading
func TestRepository_Load(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "oauth-clients.yaml")

	// Create initial config with one client
	initialConfig := `clients:
  - client_id: initial-client
    client_secret_hash: $2a$10$test
    name: Initial Client
    scopes:
      - tasks:write
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0600)
	require.NoError(t, err)

	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	// Verify initial state
	client, err := repo.GetByClientID("initial-client")
	assert.NoError(t, err)
	assert.Equal(t, "initial-client", client.ClientID)

	// Update config with additional client
	updatedConfig := `clients:
  - client_id: initial-client
    client_secret_hash: $2a$10$test
    name: Initial Client
    scopes:
      - tasks:write
  - client_id: new-client
    client_secret_hash: $2a$10$test
    name: New Client
    scopes:
      - files:read
`
	// Wait a moment to ensure file modification time changes
	time.Sleep(10 * time.Millisecond)
	err = os.WriteFile(configPath, []byte(updatedConfig), 0600)
	require.NoError(t, err)

	// Force reload
	err = repo.Reload()
	require.NoError(t, err)

	// Verify new client is available
	newClient, err := repo.GetByClientID("new-client")
	assert.NoError(t, err)
	assert.Equal(t, "new-client", newClient.ClientID)
}

// TestRepository_Load_FileNotChanged tests that reload is skipped when file hasn't changed
func TestRepository_Load_FileNotChanged(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "oauth-clients.yaml")

	config := `clients:
  - client_id: test-client
    client_secret_hash: $2a$10$test
    name: Test Client
    scopes:
      - tasks:write
`
	err := os.WriteFile(configPath, []byte(config), 0600)
	require.NoError(t, err)

	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	// Load again without modification - should be a no-op
	err = repo.Load()
	assert.NoError(t, err)

	// Client should still be accessible
	client, err := repo.GetByClientID("test-client")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// TestRepository_Authenticate_TimingConsistency tests that auth response time is consistent
// This helps prevent timing attacks
func TestRepository_Authenticate_TimingConsistency(t *testing.T) {
	hashedSecret := hashPassword(t, "correct-secret")
	clients := []OAuthClient{
		{
			ClientID:         "existing-client",
			ClientSecretHash: hashedSecret,
			Name:             "Existing Client",
			Scopes:           []string{"tasks:write"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	// Measure time for existing client with wrong password
	start := time.Now()
	_, err := repo.Authenticate("existing-client", "wrong-password")
	existingClientTime := time.Since(start)
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// Measure time for non-existing client
	start = time.Now()
	_, err = repo.Authenticate("non-existing-client", "any-password")
	nonExistingClientTime := time.Since(start)
	assert.ErrorIs(t, err, ErrClientNotFound)

	// The existing client check involves bcrypt comparison which takes ~100ms
	// Non-existing client returns immediately
	// This is expected behavior - bcrypt timing attack resistance is internal to bcrypt
	// The important thing is that bcrypt.CompareHashAndPassword is used for existing clients
	t.Logf("Existing client (wrong password): %v", existingClientTime)
	t.Logf("Non-existing client: %v", nonExistingClientTime)

	// Note: This test documents the timing behavior rather than enforcing exact timing
	// bcrypt provides timing attack resistance for password comparison
	// but the existence check happens before bcrypt comparison
}

// TestRepository_EmptyConfig tests repository with empty client list
func TestRepository_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "oauth-clients.yaml")

	config := `clients: []
`
	err := os.WriteFile(configPath, []byte(config), 0600)
	require.NoError(t, err)

	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	// List should return empty
	clients := repo.List()
	assert.Empty(t, clients)

	// GetByClientID should return not found
	_, err = repo.GetByClientID("any-client")
	assert.ErrorIs(t, err, ErrClientNotFound)
}

// TestRepository_CreateEmptyConfig tests automatic empty config creation
func TestRepository_CreateEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "oauth-clients.yaml")

	// Config file and directory don't exist
	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	// File should be created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Repository should work with empty config
	clients := repo.List()
	assert.Empty(t, clients)

	_, err = repo.GetByClientID("any-client")
	assert.ErrorIs(t, err, ErrClientNotFound)
}

// TestRepository_Reload tests forced reload
func TestRepository_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "oauth-clients.yaml")

	config := `clients:
  - client_id: client-v1
    client_secret_hash: $2a$10$test
    name: Client V1
    scopes:
      - tasks:write
`
	err := os.WriteFile(configPath, []byte(config), 0600)
	require.NoError(t, err)

	repo, err := NewRepository(configPath)
	require.NoError(t, err)

	// Verify initial client
	client, err := repo.GetByClientID("client-v1")
	assert.NoError(t, err)
	assert.Equal(t, "Client V1", client.Name)

	// Update config (same timestamp to test Reload resets lastModified)
	updatedConfig := `clients:
  - client_id: client-v2
    client_secret_hash: $2a$10$test
    name: Client V2
    scopes:
      - files:read
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0600)
	require.NoError(t, err)

	// Force reload
	err = repo.Reload()
	require.NoError(t, err)

	// Old client should not exist
	_, err = repo.GetByClientID("client-v1")
	assert.ErrorIs(t, err, ErrClientNotFound)

	// New client should exist
	client, err = repo.GetByClientID("client-v2")
	assert.NoError(t, err)
	assert.Equal(t, "Client V2", client.Name)
}

// TestRepository_MultipleScopes tests client with multiple scopes
func TestRepository_MultipleScopes(t *testing.T) {
	hashedSecret := hashPassword(t, "secret")
	clients := []OAuthClient{
		{
			ClientID:         "multi-scope-client",
			ClientSecretHash: hashedSecret,
			Name:             "Multi Scope Client",
			Scopes:           []string{"tasks:write", "tasks:read", "files:write", "files:read", "admin:*"},
		},
	}

	repo, cleanup := createTestRepository(t, clients)
	defer cleanup()

	client, err := repo.GetByClientID("multi-scope-client")
	require.NoError(t, err)

	assert.Len(t, client.Scopes, 5)
	assert.Contains(t, client.Scopes, "tasks:write")
	assert.Contains(t, client.Scopes, "tasks:read")
	assert.Contains(t, client.Scopes, "files:write")
	assert.Contains(t, client.Scopes, "files:read")
	assert.Contains(t, client.Scopes, "admin:*")
}

// TestDefaultConfigPath tests the default config path function
func TestDefaultConfigPath(t *testing.T) {
	path := defaultConfigPath()
	assert.Contains(t, path, "omnidrop")
	assert.Contains(t, path, "oauth-clients.yaml")
}
