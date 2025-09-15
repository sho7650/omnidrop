package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

type TaskRequest struct {
	Title   string   `json:"title"`
	Note    string   `json:"note,omitempty"`
	Project string   `json:"project,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type TaskResponse struct {
	Status  string `json:"status"`
	Created bool   `json:"created"`
	Reason  string `json:"reason,omitempty"`
}

func main() {
	// Load environment variables from .env file if it exists
	// Try project root first (for development and make commands)
	envPaths := []string{".env", "cmd/omnidrop-server/.env"}
	var envLoaded bool
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("✅ Loaded environment from: %s", envPath)
			envLoaded = true
			break
		}
	}
	if !envLoaded {
		log.Printf("⚠️ No .env file found (checked: %v)", envPaths)
	}

	// Get configuration from environment
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8787"
	}

	log.Printf("🚀 OmniDrop Server %s (built at %s)", Version, BuildTime)
	
	// Validate environment before starting server
	if err := validateEnvironment(port); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}
	
	log.Printf("📡 Starting server on port %s", port)
	log.Printf("🔐 Authentication token configured: %t", token != "")
	log.Printf("📁 Working directory: %s", getWorkingDir())

	// Test AppleScript accessibility
	testAppleScriptAccess()

	// Set up routes
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check authentication
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		providedToken := strings.TrimPrefix(authHeader, "Bearer ")
		if providedToken != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var taskReq TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if taskReq.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		// Create task via AppleScript
		response := createOmniFocusTask(taskReq)

		// Return response
		w.Header().Set("Content-Type", "application/json")
		if response.Status == "error" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(response)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"version": Version,
		})
	})

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      nil, // Use default mux
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
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

func createOmniFocusTask(task TaskRequest) TaskResponse {
	// Get AppleScript path with environment-based resolution
	scriptPath, err := getAppleScriptPath()
	if err != nil {
		return TaskResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript path error: %v", err),
		}
	}

	// Prepare arguments for direct passing to AppleScript
	tagsString := strings.Join(task.Tags, ",")

	// Execute AppleScript with direct arguments
	log.Printf("📝 Creating OmniFocus task: %s", task.Title)
	log.Printf("🍎 Executing AppleScript: %s", scriptPath)
	log.Printf("📋 Arguments: title='%s', note='%s', project='%s', tags='%s'",
		task.Title, task.Note, task.Project, tagsString)

	cmd := exec.Command("osascript", scriptPath, task.Title, task.Note, task.Project, tagsString)
	output, err := cmd.CombinedOutput() // Get both stdout and stderr

	if err != nil {
		log.Printf("❌ AppleScript execution failed: %v", err)
		log.Printf("📄 AppleScript output: %s", string(output))
		return TaskResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript execution failed: %v - Output: %s", err, string(output)),
		}
	}

	// Parse AppleScript output
	result := strings.TrimSpace(string(output))
	log.Printf("📋 AppleScript result: %s", result)

	if isSuccessResult(result) {
		log.Printf("✅ Task created successfully: %s", task.Title)
		return TaskResponse{
			Status:  "ok",
			Created: true,
		}
	}

	log.Printf("❌ Task creation failed - AppleScript returned: %s", result)
	return TaskResponse{
		Status:  "error",
		Created: false,
		Reason:  fmt.Sprintf("AppleScript returned: %s", result),
	}
}

// getAppleScriptPath returns the AppleScript path based on environment configuration
func getAppleScriptPath() (string, error) {
	// Priority 1: Explicit path via OMNIDROP_SCRIPT environment variable
	if scriptPath := os.Getenv("OMNIDROP_SCRIPT"); scriptPath != "" {
		log.Printf("🎯 Using explicit script path: %s", scriptPath)
		return validateScriptPath(scriptPath)
	}

	// Priority 2: Environment-based selection
	env := os.Getenv("OMNIDROP_ENV")
	log.Printf("🌍 Environment: %s", env)

	switch env {
	case "production":
		return getProductionScriptPath()
	case "development":
		return getDevelopmentScriptPath()
	case "test":
		return getTestScriptPath()
	default:
		// Fallback to legacy behavior with warning
		log.Printf("⚠️ OMNIDROP_ENV not set, using legacy path resolution")
		return getLegacyScriptPath()
	}
}

// validateScriptPath ensures the script file exists and is accessible
func validateScriptPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("AppleScript file not found at: %s", path)
	}
	log.Printf("✅ AppleScript found: %s", path)
	return path, nil
}

// getProductionScriptPath returns the production AppleScript path
func getProductionScriptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}
	path := fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir)
	return validateScriptPath(path)
}

// getDevelopmentScriptPath returns the development AppleScript path
func getDevelopmentScriptPath() (string, error) {
	path := "omnidrop.applescript"
	return validateScriptPath(path)
}

// getTestScriptPath returns the test AppleScript path
func getTestScriptPath() (string, error) {
	// For test environment, OMNIDROP_SCRIPT should be set
	if scriptPath := os.Getenv("OMNIDROP_SCRIPT"); scriptPath != "" {
		return validateScriptPath(scriptPath)
	}
	return "", fmt.Errorf("test environment requires OMNIDROP_SCRIPT to be set")
}

// getLegacyScriptPath provides fallback behavior
func getLegacyScriptPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}

	scriptPaths := []string{
		"omnidrop.applescript",
		fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir),
	}

	for _, path := range scriptPaths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("✅ AppleScript found (legacy): %s", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("AppleScript file not found in any location")
}

// validateEnvironment performs safety checks before server startup
func validateEnvironment(port string) error {
	env := os.Getenv("OMNIDROP_ENV")
	
	// Protect production port
	if port == "8787" && env != "production" {
		return fmt.Errorf("❌ FATAL: Port 8787 is reserved for production environment only")
	}
	
	// Validate test environment port range
	if env == "test" {
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", port)
		}
		if portNum < 8788 || portNum > 8799 {
			return fmt.Errorf("❌ FATAL: Test environment must use ports 8788-8799")
		}
	}
	
	// Protect production script in non-production environments
	homeDir, _ := os.UserHomeDir()
	prodScriptPath := fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir)
	if scriptPath := os.Getenv("OMNIDROP_SCRIPT"); scriptPath == prodScriptPath && env != "production" {
		return fmt.Errorf("❌ FATAL: Cannot use production AppleScript in non-production environment")
	}
	
	log.Printf("✅ Environment validation passed: %s on port %s", env, port)
	return nil
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

func isSuccessResult(result string) bool {
	// Define success patterns (case-insensitive)
	successPatterns := []string{"true", "ok", "success", "created", "done"}

	result = strings.TrimSpace(result)
	log.Printf("🔍 Checking result: '%s' against success patterns", result)

	// Check if the last line matches any success pattern
	lines := strings.Split(result, "\n")
	lastLine := strings.ToLower(strings.TrimSpace(lines[len(lines)-1]))

	for _, pattern := range successPatterns {
		if strings.ToLower(pattern) == lastLine {
			log.Printf("✅ Match found: last line '%s' matches pattern '%s'", lastLine, pattern)
			return true
		}
	}

	// Fallback: check if any line contains a success pattern
	resultLower := strings.ToLower(result)
	for _, pattern := range successPatterns {
		if strings.Contains(resultLower, strings.ToLower(pattern)) {
			log.Printf("✅ Match found: result contains pattern '%s'", pattern)
			return true
		}
	}

	log.Printf("❌ No success pattern matched for result: '%s'", result)
	return false
}