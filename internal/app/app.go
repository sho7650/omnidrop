package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnidrop/internal/auth"
	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	"omnidrop/internal/middleware"
	"omnidrop/internal/observability"
	"omnidrop/internal/server"
	"omnidrop/internal/services"
)

// Application manages the complete application lifecycle
type Application struct {
	config           *config.Config
	healthService    services.HealthService
	omniFocusService services.OmniFocusServiceInterface
	server           *server.Server
	logger           *slog.Logger
	version          string
	buildTime        string
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
	// Setup structured logging
	a.logger = observability.SetupLogger()

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

	// Initialize OAuth components
	var oauthRepo *auth.Repository
	var jwtManager *auth.JWTManager
	var authMiddleware *auth.Middleware
	var legacyAuthMiddleware *middleware.LegacyAuthMiddleware
	var tokenHandler *auth.TokenHandler

	if cfg.JWTSecret != "" {
		// Initialize OAuth client repository
		oauthRepo, err = auth.NewRepository(cfg.OAuthClientsFile)
		if err != nil {
			a.logger.Warn("Failed to initialize OAuth repository",
				slog.String("error", err.Error()),
				slog.String("clients_file", cfg.OAuthClientsFile))
		} else {
			// Initialize JWT manager
			jwtManager = auth.NewJWTManager(cfg.JWTSecret)

			// Initialize OAuth middleware
			authMiddleware = auth.NewMiddleware(jwtManager, a.logger)

			// Initialize token handler
			tokenHandler = auth.NewTokenHandler(oauthRepo, jwtManager, cfg.TokenExpiry, a.logger)

			a.logger.Info("✅ OAuth authentication initialized",
				slog.String("clients_file", cfg.OAuthClientsFile),
				slog.Duration("token_expiry", cfg.TokenExpiry),
				slog.Bool("legacy_auth_enabled", cfg.LegacyAuthEnabled))
		}
	}

	// Initialize legacy authentication if enabled
	if cfg.LegacyAuthEnabled && cfg.Token != "" {
		legacyAuthMiddleware = middleware.NewLegacyAuthMiddleware(cfg.Token, a.logger)
		a.logger.Info("✅ Legacy authentication initialized (migration mode)")
	}

	// Ensure at least one authentication method is configured
	if authMiddleware == nil && legacyAuthMiddleware == nil {
		a.logger.Error("FATAL: No authentication configured")
		return config.ErrNoAuthConfigured
	}

	// Initialize services
	a.healthService = services.NewHealthService(cfg)
	a.omniFocusService = services.NewOmniFocusService(cfg)
	filesService := services.NewFilesService(cfg)

	// Initialize handlers and server
	h := handlers.New(cfg, a.omniFocusService, filesService)
	a.server = server.NewServer(cfg, h, authMiddleware, legacyAuthMiddleware, tokenHandler, a.logger)

	return nil
}

// displayStartupInfo shows application startup information
func (a *Application) displayStartupInfo() {
	a.logger.Info("🚀 OmniDrop Server starting",
		slog.String("version", a.version),
		slog.String("build_time", a.buildTime),
		slog.String("port", a.config.Port),
		slog.Bool("auth_configured", a.config.Token != ""),
		slog.String("working_dir", services.GetWorkingDirectory()),
	)
}

// performHealthChecks runs system health checks
func (a *Application) performHealthChecks() {
	healthResult := a.healthService.CheckAppleScriptHealth()
	if !healthResult.AppleScriptAccessible {
		a.logger.Warn("⚠️ AppleScript health check failed",
			slog.Any("errors", healthResult.Errors))
	} else {
		a.logger.Info("✅ AppleScript health check passed")
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
	a.logger.Info("🛑 Shutting down application...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Error during server shutdown", slog.String("error", err.Error()))
		return err
	}

	a.logger.Info("✅ Application gracefully stopped")
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
