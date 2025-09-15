package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
	log.Printf("📁 Working directory: %s", getWorkingDir())

	// Test AppleScript accessibility
	testAppleScriptAccess()

	// Initialize services
	omniFocusService := services.NewOmniFocusService(cfg)

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


func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

func testAppleScriptAccess() {
	log.Printf("🍎 Testing AppleScript access...")

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("❌ Cannot get home directory: %v", err)
		return
	}

	// Try different paths for the AppleScript
	// Development location is checked first for testing
	scriptPaths := []string{
		"omnidrop.applescript",
		fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir),
	}

	var foundScript string
	for _, path := range scriptPaths {
		if _, err := os.Stat(path); err == nil {
			foundScript = path
			break
		}
	}

	if foundScript == "" {
		log.Printf("⚠️ AppleScript file not found in expected locations")
		return
	}

	log.Printf("✅ AppleScript found: %s", foundScript)

	// Test basic AppleScript execution
	cmd := exec.Command("osascript", "-e", "tell application \"System Events\" to get name of processes")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ AppleScript test failed: %v", err)
		return
	}

	log.Printf("✅ AppleScript accessibility confirmed")

	// Check if OmniFocus is available
	if strings.Contains(string(output), "OmniFocus") {
		log.Printf("✅ OmniFocus detected in running processes")
	} else {
		log.Printf("⚠️ OmniFocus not currently running")
	}
}