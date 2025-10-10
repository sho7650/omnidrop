package app

import (
	"os"
	"testing"
	"time"

	"omnidrop/internal/observability"
)

func TestNew(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.version != "dev" {
		t.Errorf("Expected version 'dev', got %s", app.version)
	}

	if app.buildTime != "unknown" {
		t.Errorf("Expected buildTime 'unknown', got %s", app.buildTime)
	}
}

func TestNewWithVersion(t *testing.T) {
	version := "1.0.0"
	buildTime := "2025-09-15T19:00:00Z"

	app := NewWithVersion(version, buildTime)
	if app == nil {
		t.Fatal("NewWithVersion() returned nil")
	}

	if app.version != version {
		t.Errorf("Expected version %s, got %s", version, app.version)
	}

	if app.buildTime != buildTime {
		t.Errorf("Expected buildTime %s, got %s", buildTime, app.buildTime)
	}
}

func TestApplication_Initialize(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize
	err := app.initialize()

	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	if app.config == nil {
		t.Error("Config not initialized")
	}

	if app.healthService == nil {
		t.Error("Health service not initialized")
	}

	if app.omniFocusService == nil {
		t.Error("OmniFocus service not initialized")
	}

	if app.server == nil {
		t.Error("Server not initialized")
	}
}

func TestApplication_Initialize_ConfigError(t *testing.T) {
	// Clear TOKEN environment variable to trigger config error
	// Enable legacy auth to make TOKEN required
	originalToken := os.Getenv("TOKEN")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	os.Unsetenv("TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("TOKEN", originalToken)
		}
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize
	err := app.initialize()

	if err == nil {
		t.Error("Expected initialize() to fail when TOKEN is not set and legacy auth is enabled")
	}
}

func TestApplication_GetMethods(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize
	err := app.initialize()
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	// Test getter methods
	if app.GetConfig() == nil {
		t.Error("GetConfig() returned nil")
	}

	if app.GetServer() == nil {
		t.Error("GetServer() returned nil")
	}

	if app.GetHealthService() == nil {
		t.Error("GetHealthService() returned nil")
	}

	if app.GetOmniFocusService() == nil {
		t.Error("GetOmniFocusService() returned nil")
	}
}

func TestApplication_DisplayStartupInfo(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := NewWithVersion("1.0.0", "2025-09-15")
	app.logger = observability.SetupLogger() // Setup logger before initialize

	err := app.initialize()
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	// This method just logs output, so we test that it doesn't panic
	app.displayStartupInfo()
}

func TestApplication_PerformHealthChecks(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize

	err := app.initialize()
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	// This method just logs output, so we test that it doesn't panic
	app.performHealthChecks()
}

func TestApplication_Shutdown(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize

	err := app.initialize()
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	// Test shutdown without starting server (should not error)
	err = app.shutdown()
	if err != nil {
		t.Errorf("shutdown() failed: %v", err)
	}
}

func TestApplication_Lifecycle(t *testing.T) {
	// This test verifies the basic lifecycle without actually running the server
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8788")
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()
	app.logger = observability.SetupLogger() // Setup logger before initialize

	// Test initialization
	err := app.initialize()
	if err != nil {
		t.Fatalf("initialize() failed: %v", err)
	}

	// Verify all components are initialized
	if app.config == nil || app.healthService == nil ||
		app.omniFocusService == nil || app.server == nil {
		t.Error("Not all components were initialized")
	}

	// Test that we can call startup methods without panicking
	app.displayStartupInfo()
	app.performHealthChecks()

	// Test shutdown
	err = app.shutdown()
	if err != nil {
		t.Errorf("shutdown() failed: %v", err)
	}
}

// TestApplication_RunWithMock tests the Run method behavior with mocked components
// Note: This is a basic test since Run() normally blocks
func TestApplication_RunBasicFlow(t *testing.T) {
	// Set required environment variables for testing
	os.Setenv("TOKEN", "test-token")
	os.Setenv("PORT", "8789") // Use different port to avoid conflicts
	os.Setenv("OMNIDROP_ENV", "test")
	os.Setenv("OMNIDROP_LEGACY_AUTH_ENABLED", "true")
	defer func() {
		os.Unsetenv("TOKEN")
		os.Unsetenv("PORT")
		os.Unsetenv("OMNIDROP_ENV")
		os.Unsetenv("OMNIDROP_LEGACY_AUTH_ENABLED")
	}()

	app := New()

	// Test that initialization works as part of Run()
	go func() {
		// Send interrupt signal after a short delay to stop the application
		time.Sleep(100 * time.Millisecond)
		if p, err := os.FindProcess(os.Getpid()); err == nil {
			p.Signal(os.Interrupt)
		}
	}()

	// This should initialize, start, and then shutdown gracefully
	// The test will timeout if the graceful shutdown doesn't work
	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Run() did not complete within timeout")
	}
}
