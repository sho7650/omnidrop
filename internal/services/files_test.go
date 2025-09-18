package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"omnidrop/internal/config"
)

func TestFilesService_WriteFile_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cfg := &config.Config{
		FilesDir: tempDir,
	}
	service := NewFilesService(cfg)

	// Execute
	response := service.WriteFile(context.Background(), FileWriteRequest{
		Filename: "test.txt",
		Content:  "Hello, World!",
	})

	// Assert
	assert.Equal(t, "ok", response.Status)
	assert.True(t, response.Created)
	assert.Equal(t, "test.txt", response.Path)

	// Verify file exists and has correct content
	filePath := filepath.Join(tempDir, "test.txt")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))
}

func TestFilesService_WriteFile_WithDirectory(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cfg := &config.Config{
		FilesDir: tempDir,
	}
	service := NewFilesService(cfg)

	// Execute
	response := service.WriteFile(context.Background(), FileWriteRequest{
		Filename:  "report.txt",
		Content:   "Monthly report",
		Directory: "reports/2025",
	})

	// Assert
	assert.Equal(t, "ok", response.Status)
	assert.True(t, response.Created)
	assert.Equal(t, "reports/2025/report.txt", response.Path)

	// Verify file exists and directory was created
	filePath := filepath.Join(tempDir, "reports", "2025", "report.txt")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "Monthly report", string(content))
}

func TestFilesService_WriteFile_PathTraversal(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cfg := &config.Config{
		FilesDir: tempDir,
	}
	service := NewFilesService(cfg)

	// Test cases for path traversal attacks
	testCases := []struct {
		name      string
		filename  string
		directory string
	}{
		{
			name:     "filename with path traversal",
			filename: "../../../etc/passwd",
		},
		{
			name:      "directory with path traversal",
			filename:  "malicious.txt",
			directory: "../../../etc",
		},
		{
			name:     "filename with current directory",
			filename: "./../../etc/passwd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := service.WriteFile(context.Background(), FileWriteRequest{
				Filename:  tc.filename,
				Content:   "malicious content",
				Directory: tc.directory,
			})

			assert.Equal(t, "error", response.Status)
			assert.False(t, response.Created)
			assert.Contains(t, response.Reason, "invalid path")
		})
	}
}

func TestFilesService_WriteFile_FileExists(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("existing content"), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		FilesDir: tempDir,
	}
	service := NewFilesService(cfg)

	// Execute
	response := service.WriteFile(context.Background(), FileWriteRequest{
		Filename: "existing.txt",
		Content:  "new content",
	})

	// Assert
	assert.Equal(t, "error", response.Status)
	assert.False(t, response.Created)
	assert.Contains(t, response.Reason, "file already exists")

	// Verify original content is unchanged
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.Equal(t, "existing content", string(content))
}

func TestFilesService_WriteFile_EmptyFilename(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cfg := &config.Config{
		FilesDir: tempDir,
	}
	service := NewFilesService(cfg)

	// Execute
	response := service.WriteFile(context.Background(), FileWriteRequest{
		Filename: "",
		Content:  "test content",
	})

	// Assert
	assert.Equal(t, "error", response.Status)
	assert.False(t, response.Created)
	assert.Contains(t, response.Reason, "filename is required")
}

func TestFilesService_WriteFile_InvalidBaseDirectory(t *testing.T) {
	// Setup with non-existent base directory
	cfg := &config.Config{
		FilesDir: "/non/existent/directory",
	}
	service := NewFilesService(cfg)

	// Execute
	response := service.WriteFile(context.Background(), FileWriteRequest{
		Filename: "test.txt",
		Content:  "test content",
	})

	// Assert
	assert.Equal(t, "error", response.Status)
	assert.False(t, response.Created)
	assert.Contains(t, response.Reason, "failed to create directory")
}