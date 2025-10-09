package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"omnidrop/internal/services"
)

// FileRequest represents the JSON request for file creation
type FileRequest struct {
	Filename  string `json:"filename"`
	Content   string `json:"content"`
	Directory string `json:"directory,omitempty"`
}

// FileResponse represents the JSON response for file creation
type FileResponse struct {
	Status  string `json:"status"`
	Created bool   `json:"created"`
	Path    string `json:"path,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// CreateFile handles POST requests to create files
func (h *Handlers) CreateFile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Method validation
	if r.Method != http.MethodPost {
		writeMethodNotAllowedError(w, "Only POST method is allowed for file creation")
		return
	}

	// Authentication is handled by middleware - no need to re-authenticate here

	// Parse request body
	var fileReq FileRequest
	if err := json.NewDecoder(r.Body).Decode(&fileReq); err != nil {
		writeValidationError(w, "Invalid JSON format in request body")
		return
	}

	// Validate required fields
	if fileReq.Filename == "" {
		writeValidationError(w, "Filename field is required and cannot be empty")
		return
	}

	if fileReq.Content == "" {
		writeValidationError(w, "Content field is required and cannot be empty")
		return
	}

	// Create file via FilesService
	response := h.filesService.WriteFile(ctx, services.FileWriteRequest{
		Filename:  fileReq.Filename,
		Content:   fileReq.Content,
		Directory: fileReq.Directory,
	})

	// Prepare response
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate HTTP status code
	if response.Status == "error" {
		w.WriteHeader(http.StatusBadRequest)
	}

	fileResponse := FileResponse{
		Status:  response.Status,
		Created: response.Created,
		Path:    response.Path,
		Reason:  response.Reason,
	}

	if err := json.NewEncoder(w).Encode(fileResponse); err != nil {
		writeInternalError(w, "Failed to encode response", err)
	}
}