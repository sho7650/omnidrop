package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	"omnidrop/internal/server"
	"omnidrop/internal/services"
)

// Application manages the complete application lifecycle
type Application struct {
	config        *config.Config
	healthService services.HealthService
	omniFocusService services.OmniFocusServiceInterface
	server        *server.Server
	version       string
	buildTime     string
}

// New creates a new application instance
func New() *Application {
	return &Application{
		version:   "dev",
		buildTime: "unknown",
	}
}

// NewWithVersion creates a new application instance with version information
func NewWithVersion(version, buildTime string) *Application {
	return &Application{
		version:   version,
		buildTime: buildTime,
	}
}

// Run starts the application and handles the complete lifecycle
func (a *Application) Run() error {
	// Initialize the application
	if err := a.initialize(); err != nil {
		return err
	}

	// Display startup information
	a.displayStartupInfo()

	// Perform health checks
	a.performHealthChecks()

	// Start the server
	return a.startAndWait()
}

// initialize sets up all application components
func (a *Application) initialize() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.config = cfg

	// Initialize services
	a.healthService = services.NewHealthService(cfg)
	a.omniFocusService = services.NewOmniFocusService(cfg)

	// Initialize handlers and server
	h := handlers.New(cfg, a.omniFocusService)
	a.server = server.NewServer(cfg, h)

	return nil
}

// displayStartupInfo shows application startup information
func (a *Application) displayStartupInfo() {
	log.Printf("🚀 OmniDrop Server %s (built at %s)", a.version, a.buildTime)
	log.Printf("📡 Starting server on port %s", a.config.Port)
	log.Printf("🔐 Authentication token configured: %t", a.config.Token != "")
	log.Printf("📁 Working directory: %s", services.GetWorkingDirectory())
}

// performHealthChecks runs system health checks
func (a *Application) performHealthChecks() {
	healthResult := a.healthService.CheckAppleScriptHealth()
	if !healthResult.AppleScriptAccessible {
		log.Printf("⚠️ AppleScript health check failed: %v", healthResult.Errors)
	}
}

// startAndWait starts the server and waits for shutdown signals
func (a *Application) startAndWait() error {
	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.Start(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-sigChan:
		return a.shutdown()
	}
}

// shutdown gracefully shuts down the application
func (a *Application) shutdown() error {
	log.Println("🛑 Shutting down application...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
		return err
	}

	log.Println("✅ Application gracefully stopped")
	return nil
}

// GetConfig returns the application configuration (useful for testing)
func (a *Application) GetConfig() *config.Config {
	return a.config
}

// GetServer returns the server instance (useful for testing)
func (a *Application) GetServer() *server.Server {
	return a.server
}

// GetHealthService returns the health service (useful for testing)
func (a *Application) GetHealthService() services.HealthService {
	return a.healthService
}

// GetOmniFocusService returns the OmniFocus service (useful for testing)
func (a *Application) GetOmniFocusService() services.OmniFocusServiceInterface {
	return a.omniFocusService
}