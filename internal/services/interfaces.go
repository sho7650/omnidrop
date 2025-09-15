package services

import (
	"context"
)

// TaskCreateRequest represents a request to create a task in OmniFocus
type TaskCreateRequest struct {
	Title   string
	Note    string
	Project string
	Tags    []string
}

// TaskCreateResponse represents the response from creating a task
type TaskCreateResponse struct {
	Status  string
	Created bool
	Reason  string
}

// OmniFocusServiceInterface defines the interface for OmniFocus operations
type OmniFocusServiceInterface interface {
	CreateTask(ctx context.Context, req TaskCreateRequest) TaskCreateResponse
}

// HealthService defines the interface for system health checks
type HealthService interface {
	CheckAppleScriptHealth() HealthResult
	CheckOmniFocusStatus() bool
}

// AppleScriptExecutor defines the interface for AppleScript execution
// This allows for easy mocking in tests
type AppleScriptExecutor interface {
	Execute(ctx context.Context, script string, args ...string) ([]byte, error)
}

// HealthResult contains structured health check information
type HealthResult struct {
	AppleScriptAccessible bool     `json:"applescript_accessible"`
	OmniFocusRunning     bool     `json:"omnifocus_running"`
	ScriptPath           string   `json:"script_path"`
	Errors              []string `json:"errors,omitempty"`
	Details             string   `json:"details,omitempty"`
}