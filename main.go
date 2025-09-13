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
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8787"
	}
	
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost"
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}
	
	// Setup HTTP handler
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Check authorization
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + token
		if authHeader != expectedAuth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Parse JSON body
		var taskReq TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		
		// Validate required fields
		if taskReq.Title == "" {
			sendErrorResponse(w, http.StatusBadRequest, "Title is required")
			return
		}
		
		// Convert request to JSON string for AppleScript
		jsonData, err := json.Marshal(taskReq)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to marshal JSON")
			return
		}
		
		// Execute AppleScript
		cmd := exec.Command("osascript", "omnidrop.applescript", string(jsonData))
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			log.Printf("AppleScript error: %v, output: %s", err, string(output))
			sendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("AppleScript error: %s", strings.TrimSpace(string(output))))
			return
		}
		
		// Check if successful
		if strings.TrimSpace(string(output)) == "OK" {
			sendSuccessResponse(w)
			log.Printf("Task created successfully: %s", taskReq.Title)
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unexpected response: %s", string(output)))
		}
	})
	
	// Start server
	addr := fmt.Sprintf("%s:%s", hostname, port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Token is configured")
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func sendSuccessResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TaskResponse{
		Status:  "ok",
		Created: true,
	})
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(TaskResponse{
		Status:  "error",
		Created: false,
		Reason:  reason,
	})
}