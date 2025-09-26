package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"omnidrop/internal/config"
	"omnidrop/internal/services"
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

	// Check authentication
	if !h.authenticateRequest(r) {
		writeAuthenticationError(w, "Invalid or missing authentication token")
		return
	}

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
		writeInternalError(w, "Failed to encode response", err)
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": h.getVersion(),
	}); err != nil {
		writeInternalError(w, "Failed to encode health response", err)
	}
}

func (h *Handlers) authenticateRequest(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	providedToken := strings.TrimPrefix(authHeader, "Bearer ")
	return providedToken == h.cfg.Token
}

// getVersion returns the version - will be set via ldflags during build
func (h *Handlers) getVersion() string {
	return "dev" // This will be replaced by build-time variables
}
