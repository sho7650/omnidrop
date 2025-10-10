package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"omnidrop/internal/auth"
	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	omnimiddleware "omnidrop/internal/middleware"
)

// Server manages the HTTP server lifecycle and configuration
type Server struct {
	config               *config.Config
	handlers             *handlers.Handlers
	authMiddleware       *auth.Middleware
	legacyAuthMiddleware *omnimiddleware.LegacyAuthMiddleware
	tokenHandler         *auth.TokenHandler
	logger               *slog.Logger
	httpSrv              *http.Server
	router               chi.Router
}

// NewServer creates a new server instance with the given configuration and handlers
func NewServer(cfg *config.Config, handlers *handlers.Handlers, authMiddleware *auth.Middleware, legacyAuthMiddleware *omnimiddleware.LegacyAuthMiddleware, tokenHandler *auth.TokenHandler, logger *slog.Logger) *Server {
	s := &Server{
		config:               cfg,
		handlers:             handlers,
		authMiddleware:       authMiddleware,
		legacyAuthMiddleware: legacyAuthMiddleware,
		tokenHandler:         tokenHandler,
		logger:               logger,
	}
	s.setupRouter()
	s.setupHTTPServer()
	return s
}

// setupRouter configures the chi router with middleware and routes
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Create logging configuration
	loggingCfg := omnimiddleware.DefaultLoggingConfig(s.logger)

	// Middleware stack (order matters!)
	r.Use(omnimiddleware.Recovery)              // Panic recovery (first for safety)
	r.Use(omnimiddleware.RequestIDMiddleware)   // Request ID generation
	r.Use(middleware.RealIP)                    // Real IP detection
	r.Use(omnimiddleware.HTTPLogging(loggingCfg)) // Structured logging
	r.Use(omnimiddleware.Metrics)               // Prometheus metrics collection
	r.Use(middleware.Timeout(60 * time.Second)) // Request timeout

	// Public routes (no authentication required)
	r.Get("/health", s.handlers.Health)
	r.Handle("/metrics", promhttp.Handler()) // Prometheus metrics endpoint

	// OAuth token endpoint
	if s.tokenHandler != nil {
		r.Post("/oauth/token", s.tokenHandler.HandleToken)
	}

	// Protected routes (authentication required)
	if s.authMiddleware != nil {
		// OAuth authentication mode
		s.logger.Info("Authentication: OAuth 2.0 with JWT tokens")
		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware.Authenticate)

			// Task creation requires tasks:write scope
			r.With(auth.RequireScopes("tasks:write")).Post("/tasks", s.handlers.CreateTask)

			// File creation requires files:write scope
			r.With(auth.RequireScopes("files:write")).Post("/files", s.handlers.CreateFile)
		})
	} else if s.legacyAuthMiddleware != nil {
		// Legacy authentication mode (TOKEN-based)
		s.logger.Warn("⚠️ Authentication: Legacy token-based (migration mode)")
		r.Group(func(r chi.Router) {
			r.Use(s.legacyAuthMiddleware.Authenticate)
			r.Post("/tasks", s.handlers.CreateTask)
			r.Post("/files", s.handlers.CreateFile)
		})
	} else {
		// No authentication configured - this should never happen in production
		panic("FATAL: No authentication middleware configured - server cannot start safely")
	}

	s.router = r
}

// setupHTTPServer configures the HTTP server with appropriate timeouts
func (s *Server) setupHTTPServer() {
	s.httpSrv = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Start begins serving HTTP requests
// This method blocks until the server shuts down or encounters an error
func (s *Server) Start() error {
	s.logger.Info("🚀 Server starting", slog.String("port", s.config.Port))
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server without interrupting active connections
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down server...")

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("✅ Server gracefully stopped")
	return nil
}

// GetAddress returns the server's listening address
func (s *Server) GetAddress() string {
	if s.httpSrv != nil {
		return s.httpSrv.Addr
	}
	return ":" + s.config.Port
}

// GetRouter returns the configured router (useful for testing)
func (s *Server) GetRouter() chi.Router {
	return s.router
}
