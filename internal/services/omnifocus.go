package services

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"omnidrop/internal/config"
)

type OmniFocusService struct {
	cfg *config.Config
}

// Ensure OmniFocusService implements OmniFocusServiceInterface
var _ OmniFocusServiceInterface = (*OmniFocusService)(nil)

func NewOmniFocusService(cfg *config.Config) *OmniFocusService {
	return &OmniFocusService{
		cfg: cfg,
	}
}

func (s *OmniFocusService) CreateTask(ctx context.Context, req TaskCreateRequest) TaskCreateResponse {
	// Get AppleScript path with environment-based resolution
	scriptPath, err := s.cfg.GetAppleScriptPath()
	if err != nil {
		return TaskCreateResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript path error: %v", errors.Wrap(err, "failed to resolve AppleScript path")),
		}
	}

	// Prepare arguments for direct passing to AppleScript
	tagsString := strings.Join(req.Tags, ",")

	// Execute AppleScript with direct arguments
	slog.Info("📝 Creating OmniFocus task",
		slog.String("title", req.Title),
		slog.String("script_path", scriptPath),
		slog.String("note", req.Note),
		slog.String("project", req.Project),
		slog.String("tags", tagsString))

	// Create command with context for timeout/cancellation
	cmd := exec.CommandContext(ctx, "osascript", scriptPath, req.Title, req.Note, req.Project, tagsString)
	output, err := cmd.CombinedOutput()

	if err != nil {
		wrappedErr := errors.Wrapf(err, "AppleScript execution failed for task '%s'", req.Title)
		slog.Error("❌ AppleScript execution failed",
			slog.String("task_title", req.Title),
			slog.String("error", wrappedErr.Error()),
			slog.String("output", string(output)))
		return TaskCreateResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript execution failed: %v - Output: %s", wrappedErr, string(output)),
		}
	}

	// Parse AppleScript output
	result := strings.TrimSpace(string(output))
	slog.Info("📋 AppleScript result", slog.String("result", result))

	if s.isSuccessResult(result) {
		slog.Info("✅ Task created successfully", slog.String("task_title", req.Title))
		return TaskCreateResponse{
			Status:  "ok",
			Created: true,
		}
	}

	slog.Error("❌ Task creation failed", 
		slog.String("task_title", req.Title),
		slog.String("applescript_result", result))
	return TaskCreateResponse{
		Status:  "error",
		Created: false,
		Reason:  fmt.Sprintf("AppleScript returned: %s", result),
	}
}

func (s *OmniFocusService) isSuccessResult(result string) bool {
	// Define success patterns (case-insensitive)
	successPatterns := []string{"true", "ok", "success", "created", "done"}

	result = strings.TrimSpace(result)
	slog.Debug("🔍 Checking result against success patterns", slog.String("result", result))

	// Check if the last line matches any success pattern
	lines := strings.Split(result, "\n")
	lastLine := strings.ToLower(strings.TrimSpace(lines[len(lines)-1]))

	for _, pattern := range successPatterns {
		if strings.ToLower(pattern) == lastLine {
			slog.Debug("✅ Match found: last line matches pattern",
				slog.String("last_line", lastLine),
				slog.String("pattern", pattern))
			return true
		}
	}

	// Fallback: check if any line contains a success pattern
	resultLower := strings.ToLower(result)
	for _, pattern := range successPatterns {
		if strings.Contains(resultLower, strings.ToLower(pattern)) {
			slog.Debug("✅ Match found: result contains pattern", slog.String("pattern", pattern))
			return true
		}
	}

	slog.Debug("❌ No success pattern matched", slog.String("result", result))
	return false
}
