package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	omnimiddleware "omnidrop/internal/middleware"
)

// Server manages the HTTP server lifecycle and configuration
type Server struct {
	config   *config.Config
	handlers *handlers.Handlers
	logger   *slog.Logger
	httpSrv  *http.Server
	router   chi.Router
}

// NewServer creates a new server instance with the given configuration and handlers
func NewServer(cfg *config.Config, handlers *handlers.Handlers, logger *slog.Logger) *Server {
	s := &Server{
		config:   cfg,
		handlers: handlers,
		logger:   logger,
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

	// Middleware stack
	r.Use(omnimiddleware.RequestIDMiddleware)
	r.Use(middleware.RealIP)
	r.Use(omnimiddleware.HTTPLogging(loggingCfg))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Post("/tasks", s.handlers.CreateTask)
	r.Post("/files", s.handlers.CreateFile)
	r.Get("/health", s.handlers.Health)

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
