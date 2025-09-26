package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"omnidrop/internal/config"
)

// FilesService handles file operations with security validation
type FilesService struct {
	cfg *config.Config
}

// Ensure FilesService implements FilesServiceInterface
var _ FilesServiceInterface = (*FilesService)(nil)

// NewFilesService creates a new FilesService instance
func NewFilesService(cfg *config.Config) *FilesService {
	return &FilesService{
		cfg: cfg,
	}
}

// WriteFile writes content to a file with security validation
func (s *FilesService) WriteFile(ctx context.Context, req FileWriteRequest) FileWriteResponse {
	// Validate required fields
	if req.Filename == "" {
		return FileWriteResponse{
			Status: "error",
			Reason: "filename is required and cannot be empty",
		}
	}

	// Build and validate file path
	safePath, relativePath, err := s.validateAndBuildPath(req.Filename, req.Directory)
	if err != nil {
		return FileWriteResponse{
			Status: "error",
			Reason: err.Error(),
		}
	}

	// Check if file already exists
	if _, err := os.Stat(safePath); err == nil {
		return FileWriteResponse{
			Status: "error",
			Reason: "file already exists",
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FileWriteResponse{
			Status: "error",
			Reason: fmt.Sprintf("failed to create directory: %v", err),
		}
	}

	// Write file with appropriate permissions
	if err := os.WriteFile(safePath, []byte(req.Content), 0644); err != nil {
		return FileWriteResponse{
			Status: "error",
			Reason: fmt.Sprintf("failed to write file: %v", err),
		}
	}

	return FileWriteResponse{
		Status:  "ok",
		Created: true,
		Path:    relativePath,
	}
}

// validateAndBuildPath validates the file path and prevents path traversal attacks
func (s *FilesService) validateAndBuildPath(filename, directory string) (string, string, error) {
	// Validate filename
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return "", "", fmt.Errorf("invalid path: filename cannot contain path separators or '..'")
	}

	// Build the target path
	var targetPath string
	var relativePath string

	if directory != "" {
		// Clean the directory path to prevent path traversal
		cleanDir := filepath.Clean(directory)

		// Check for path traversal attempts in directory
		if strings.Contains(cleanDir, "..") || strings.HasPrefix(cleanDir, "/") {
			return "", "", fmt.Errorf("invalid path: directory contains invalid characters")
		}

		targetPath = filepath.Join(s.cfg.FilesDir, cleanDir, filename)
		relativePath = filepath.Join(cleanDir, filename)
	} else {
		targetPath = filepath.Join(s.cfg.FilesDir, filename)
		relativePath = filename
	}

	// Resolve to absolute path and check it's within base directory
	absBasePath, err := filepath.Abs(s.cfg.FilesDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve base directory: %v", err)
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve target path: %v", err)
	}

	// Ensure the target path is within the base directory
	if !strings.HasPrefix(absTargetPath, absBasePath+string(filepath.Separator)) && absTargetPath != absBasePath {
		return "", "", fmt.Errorf("invalid path: outside base directory")
	}

	return absTargetPath, relativePath, nil
}