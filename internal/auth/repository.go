package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

var (
	// ErrClientNotFound is returned when a client is not found
	ErrClientNotFound = errors.New("client not found")
	// ErrInvalidCredentials is returned when client credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrClientDisabled is returned when a client is disabled
	ErrClientDisabled = errors.New("client is disabled")
)

// Repository manages OAuth clients
type Repository struct {
	mu           sync.RWMutex
	configPath   string
	clients      map[string]*OAuthClient // clientID -> client
	lastModified int64
}

// NewRepository creates a new OAuth client repository
func NewRepository(configPath string) (*Repository, error) {
	if configPath == "" {
		configPath = defaultConfigPath()
	}

	repo := &Repository{
		configPath: configPath,
		clients:    make(map[string]*OAuthClient),
	}

	// Load initial configuration
	if err := repo.Load(); err != nil {
		// If file doesn't exist, create empty config
		if os.IsNotExist(err) {
			if err := repo.createEmptyConfig(); err != nil {
				return nil, fmt.Errorf("failed to create empty config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return repo, nil
}

// defaultConfigPath returns the default OAuth clients configuration path
func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".omnidrop/oauth-clients.yaml"
	}
	return filepath.Join(home, ".local/share/omnidrop/oauth-clients.yaml")
}

// Load loads OAuth clients from the configuration file
func (r *Repository) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check file modification time
	fileInfo, err := os.Stat(r.configPath)
	if err != nil {
		return err
	}

	modTime := fileInfo.ModTime().Unix()
	if modTime == r.lastModified {
		// File hasn't changed, no need to reload
		return nil
	}

	// Read configuration file
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config OAuthConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Build client map
	clients := make(map[string]*OAuthClient)
	for i := range config.Clients {
		client := &config.Clients[i]
		clients[client.ClientID] = client
	}

	r.clients = clients
	r.lastModified = modTime

	return nil
}

// GetByClientID retrieves a client by client ID
func (r *Repository) GetByClientID(clientID string) (*OAuthClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.clients[clientID]
	if !ok {
		return nil, ErrClientNotFound
	}

	if client.Disabled {
		return nil, ErrClientDisabled
	}

	return client, nil
}

// Authenticate validates client credentials
func (r *Repository) Authenticate(clientID, clientSecret string) (*OAuthClient, error) {
	client, err := r.GetByClientID(clientID)
	if err != nil {
		return nil, err
	}

	// Verify client secret
	err = bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return client, nil
}

// List returns all active clients
func (r *Repository) List() []*OAuthClient {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*OAuthClient, 0, len(r.clients))
	for _, client := range r.clients {
		if !client.Disabled {
			clients = append(clients, client)
		}
	}

	return clients
}

// createEmptyConfig creates an empty OAuth clients configuration file
func (r *Repository) createEmptyConfig() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(r.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create empty config
	config := OAuthConfig{
		Clients: []OAuthClient{},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	r.lastModified = time.Now().Unix()
	return nil
}

// Reload reloads the configuration from disk
func (r *Repository) Reload() error {
	// Reset last modified time to force reload
	r.mu.Lock()
	r.lastModified = 0
	r.mu.Unlock()

	return r.Load()
}
