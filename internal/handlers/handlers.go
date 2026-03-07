package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"omnidrop/internal/config"
	"omnidrop/internal/services"
)

const (
	// MaxTaskRequestSize is the maximum allowed request body size for task creation (1MB)
	MaxTaskRequestSize = 1 << 20
	// MaxFileRequestSize is the maximum allowed request body size for file creation (10MB)
	MaxFileRequestSize = 10 << 20
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

type Handlers struct {
	cfg              *config.Config
	omniFocusService services.OmniFocusServiceInterface
	filesService     services.FilesServiceInterface
}

func New(cfg *config.Config, omniFocusService services.OmniFocusServiceInterface, filesService services.FilesServiceInterface) *Handlers {
	return &Handlers{
		cfg:              cfg,
		omniFocusService: omniFocusService,
		filesService:     filesService,
	}
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if r.Method != http.MethodPost {
		writeMethodNotAllowedError(w, "Only POST method is allowed for task creation")
		return
	}

	// Authentication is handled by middleware - no need to re-authenticate here

	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, MaxTaskRequestSize)

	// Parse request body
	var taskReq TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		writeValidationError(w, "Invalid JSON format in request body")
		return
	}

	// Validate required fields
	if taskReq.Title == "" {
		writeValidationError(w, "Title field is required and cannot be empty")
		return
	}

	// Create task via OmniFocus service
	response := h.omniFocusService.CreateTask(ctx, services.TaskCreateRequest{
		Title:   taskReq.Title,
		Note:    taskReq.Note,
		Project: taskReq.Project,
		Tags:    taskReq.Tags,
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if response.Status == "error" {
		writeAppleScriptError(w, response.Reason, nil)
		return
	}

	taskResponse := TaskResponse{
		Status:  response.Status,
		Created: response.Created,
		Reason:  response.Reason,
	}

	if err := json.NewEncoder(w).Encode(taskResponse); err != nil {
		// Headers already sent by Encode's first Write call; cannot change response status
		slog.Error("Failed to encode task response", slog.String("error", err.Error()))
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": h.getVersion(),
	}); err != nil {
		// Headers already sent by Encode's first Write call; cannot change response status
		slog.Error("Failed to encode health response", slog.String("error", err.Error()))
	}
}

// getVersion returns the version - will be set via ldflags during build
func (h *Handlers) getVersion() string {
	return "dev" // This will be replaced by build-time variables
}
