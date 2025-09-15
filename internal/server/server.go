package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
)

// Server manages the HTTP server lifecycle and configuration
type Server struct {
	config   *config.Config
	handlers *handlers.Handlers
	httpSrv  *http.Server
	router   chi.Router
}

// NewServer creates a new server instance with the given configuration and handlers
func NewServer(cfg *config.Config, handlers *handlers.Handlers) *Server {
	s := &Server{
		config:   cfg,
		handlers: handlers,
	}
	s.setupRouter()
	s.setupHTTPServer()
	return s
}

// setupRouter configures the chi router with middleware and routes
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Post("/tasks", s.handlers.CreateTask)
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
	log.Printf("🚀 Server starting on port %s", s.config.Port)
	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server without interrupting active connections
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down server...")

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("✅ Server gracefully stopped")
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