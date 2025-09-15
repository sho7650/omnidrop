package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	"omnidrop/internal/services"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("🚀 OmniDrop Server %s (built at %s)", Version, BuildTime)
	log.Printf("📡 Starting server on port %s", cfg.Port)
	log.Printf("🔐 Authentication token configured: %t", cfg.Token != "")
	log.Printf("📁 Working directory: %s", services.GetWorkingDirectory())

	// Initialize services
	healthService := services.NewHealthService(cfg)
	omniFocusService := services.NewOmniFocusService(cfg)

	// Test AppleScript accessibility
	healthResult := healthService.CheckAppleScriptHealth()
	if !healthResult.AppleScriptAccessible {
		log.Printf("⚠️ AppleScript health check failed: %v", healthResult.Errors)
	}

	// Initialize handlers
	h := handlers.New(cfg, omniFocusService)

	// Set up router with middleware
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Post("/tasks", h.CreateTask)
	r.Get("/health", h.Health)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server gracefully stopped")
	}
}