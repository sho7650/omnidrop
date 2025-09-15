package main

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

	// Initialize handlers and server
	h := handlers.New(cfg, omniFocusService)
	srv := server.NewServer(cfg, h)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
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