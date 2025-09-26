package services

import "context"

// FilesServiceInterface defines the interface for file operations
type FilesServiceInterface interface {
	WriteFile(ctx context.Context, req FileWriteRequest) FileWriteResponse
}

// FileWriteRequest represents a request to write a file
type FileWriteRequest struct {
	Filename  string // Required: name of the file to create
	Content   string // Required: content to write to the file
	Directory string // Optional: subdirectory path within the base directory
}

// FileWriteResponse represents the response from a file write operation
type FileWriteResponse struct {
	Status  string // "ok" or "error"
	Created bool   // true if file was successfully created
	Path    string // relative path of the created file (on success)
	Reason  string // error message (on failure)
}