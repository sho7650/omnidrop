package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// ErrNoAuthConfigured is returned when no authentication method is configured
var ErrNoAuthConfigured = errors.New("no authentication method configured: either TOKEN (with OMNIDROP_LEGACY_AUTH_ENABLED=true) or OMNIDROP_JWT_SECRET must be set")

type Config struct {
	// Server configuration
	Port  string
	Token string

	// Environment configuration
	Environment string // production, development, test
	ScriptPath  string

	// AppleScript configuration
	AppleScriptFile string

	// Files configuration
	FilesDir string // Base directory for file operations

	// OAuth configuration
	JWTSecret         string
	TokenExpiry       time.Duration
	OAuthClientsFile  string
	LegacyAuthEnabled bool
}

func Load() (*Config, error) {
	// Load environment variables from .env file if it exists
	loadEnvFile()

	cfg := &Config{
		Port:              getEnvWithDefault("PORT", "8787"),
		Token:             os.Getenv("TOKEN"),
		Environment:       getEnvWithDefault("OMNIDROP_ENV", ""),
		ScriptPath:        os.Getenv("OMNIDROP_SCRIPT"),
		AppleScriptFile:   "omnidrop.applescript",
		FilesDir:          getFilesDir(),
		JWTSecret:         os.Getenv("OMNIDROP_JWT_SECRET"),
		TokenExpiry:       getTokenExpiry(),
		OAuthClientsFile:  getOAuthClientsFile(),
		LegacyAuthEnabled: getEnvWithDefault("OMNIDROP_LEGACY_AUTH_ENABLED", "false") == "true",
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// Validate authentication configuration based on mode
	if c.LegacyAuthEnabled {
		// Legacy authentication mode requires TOKEN
		if c.Token == "" {
			return fmt.Errorf("TOKEN environment variable is required when OMNIDROP_LEGACY_AUTH_ENABLED=true")
		}
	}

	// If OAuth is configured (JWT secret provided), validate it
	if c.JWTSecret != "" {
		// JWT secret should be at least 32 characters for security
		if len(c.JWTSecret) < 32 {
			return fmt.Errorf("OMNIDROP_JWT_SECRET must be at least 32 characters for security")
		}
	} else if !c.LegacyAuthEnabled {
		// If legacy auth is disabled and no JWT secret, we have no authentication method
		return fmt.Errorf("OMNIDROP_JWT_SECRET is required when OAuth is enabled (OMNIDROP_LEGACY_AUTH_ENABLED=false)")
	}

	// Validate environment-specific rules
	if err := c.validateEnvironment(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateEnvironment() error {
	// Protect production port
	if c.Port == "8787" && c.Environment != "production" {
		return fmt.Errorf("❌ FATAL: Port 8787 is reserved for production environment only")
	}

	// Validate test environment port range
	if c.Environment == "test" {
		portNum, err := strconv.Atoi(c.Port)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", c.Port)
		}
		if portNum < 8788 || portNum > 8799 {
			return fmt.Errorf("❌ FATAL: Test environment must use ports 8788-8799")
		}
	}

	// Protect production script in non-production environments
	if c.Environment != "production" {
		homeDir, _ := os.UserHomeDir()
		prodScriptPath := fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir)
		if c.ScriptPath == prodScriptPath {
			return fmt.Errorf("❌ FATAL: Cannot use production AppleScript in non-production environment")
		}
	}

	return nil
}

func (c *Config) GetAppleScriptPath() (string, error) {
	// Priority 1: Explicit path via OMNIDROP_SCRIPT environment variable
	if c.ScriptPath != "" {
		return validateScriptPath(c.ScriptPath)
	}

	// Priority 2: Environment-based selection
	switch c.Environment {
	case "production":
		return c.getProductionScriptPath()
	case "development":
		return c.getDevelopmentScriptPath()
	case "test":
		return c.getTestScriptPath()
	default:
		// Fallback to legacy behavior
		return c.getLegacyScriptPath()
	}
}

func (c *Config) getProductionScriptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}
	path := fmt.Sprintf("%s/.local/share/omnidrop/%s", homeDir, c.AppleScriptFile)
	return validateScriptPath(path)
}

func (c *Config) getDevelopmentScriptPath() (string, error) {
	return validateScriptPath(c.AppleScriptFile)
}

func (c *Config) getTestScriptPath() (string, error) {
	if c.ScriptPath != "" {
		return validateScriptPath(c.ScriptPath)
	}
	return "", fmt.Errorf("test environment requires OMNIDROP_SCRIPT to be set")
}

func (c *Config) getLegacyScriptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}

	scriptPaths := []string{
		c.AppleScriptFile,
		fmt.Sprintf("%s/.local/share/omnidrop/%s", homeDir, c.AppleScriptFile),
	}

	for _, path := range scriptPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("AppleScript file not found in any location")
}

func loadEnvFile() {
	envPaths := []string{".env", "cmd/omnidrop-server/.env"}
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			break
		}
	}
}

func validateScriptPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("AppleScript file not found at: %s", path)
	}
	return path, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getFilesDir() string {
	if dir := os.Getenv("OMNIDROP_FILES_DIR"); dir != "" {
		return dir
	}

	// Default location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./files" // Fallback to relative directory
	}

	return fmt.Sprintf("%s/.local/share/omnidrop/files", homeDir)
}

func getTokenExpiry() time.Duration {
	expiryStr := getEnvWithDefault("OMNIDROP_TOKEN_EXPIRY", "24h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return expiry
}

func getOAuthClientsFile() string {
	if file := os.Getenv("OMNIDROP_OAUTH_CLIENTS_FILE"); file != "" {
		return file
	}

	// Default location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./oauth-clients.yaml" // Fallback to relative path
	}

	return fmt.Sprintf("%s/.local/share/omnidrop/oauth-clients.yaml", homeDir)
}
