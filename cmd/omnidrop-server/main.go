package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

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

	// Start server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createOmniFocusTask(task TaskRequest) TaskResponse {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return TaskResponse{
			Status: "error",
			Reason: fmt.Sprintf("Could not get home directory: %v", err),
		}
	}

	// Try different paths for the AppleScript
	// Development location is checked first for testing
	scriptPaths := []string{
		"omnidrop.applescript", // Development location (priority)
		fmt.Sprintf("%s/.local/share/omnidrop/omnidrop.applescript", homeDir), // Installed location
	}

	var scriptPath string
	for _, path := range scriptPaths {
		if _, err := os.Stat(path); err == nil {
			scriptPath = path
			break
		}
	}

	if scriptPath == "" {
		return TaskResponse{
			Status: "error",
			Reason: "AppleScript file not found",
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

	result = strings.ToLower(strings.TrimSpace(result))
	log.Printf("🔍 Checking result: '%s' against success patterns", result)

	for _, pattern := range successPatterns {
		if strings.ToLower(pattern) == result {
			log.Printf("✅ Match found: '%s' matches pattern '%s'", result, pattern)
			return true
		}
	}

	log.Printf("❌ No success pattern matched for result: '%s'", result)
	return false
}