package services

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"omnidrop/internal/config"
	"omnidrop/internal/observability"
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
	// Start timing for metrics
	start := time.Now()
	
	// Get AppleScript path with environment-based resolution
	scriptPath, err := s.cfg.GetAppleScriptPath()
	if err != nil {
		observability.TaskCreationsTotal.WithLabelValues("failure").Inc()
		return TaskCreateResponse{
			Status: "error",
			Reason: fmt.Sprintf("failed to resolve AppleScript path: %v", err),
		}
	}

	// Prepare arguments for direct passing to AppleScript
	// Sanitize inputs to prevent AppleScript injection via string terminators
	title := sanitizeAppleScriptArg(req.Title)
	note := sanitizeAppleScriptArg(req.Note)
	project := sanitizeAppleScriptArg(req.Project)
	tagsString := sanitizeAppleScriptArg(strings.Join(req.Tags, ","))

	// Collect business metrics
	if req.Project != "" {
		observability.TasksWithProjectTotal.Inc()
	}
	if len(req.Tags) > 0 {
		observability.TasksWithTagsTotal.Inc()
	}

	// Execute AppleScript with direct arguments
	slog.Info("📝 Creating OmniFocus task",
		slog.String("title", req.Title),
		slog.String("script_path", scriptPath),
		slog.String("note", req.Note),
		slog.String("project", req.Project),
		slog.String("tags", tagsString))

	// Start timing AppleScript execution
	scriptStart := time.Now()
	
	// Create command with context for timeout/cancellation
	cmd := exec.CommandContext(ctx, "osascript", scriptPath, title, note, project, tagsString)
	output, err := cmd.CombinedOutput()
	
	// Record AppleScript execution duration
	observability.AppleScriptExecutionDuration.Observe(time.Since(scriptStart).Seconds())

	if err != nil {
		// Categorize error type for metrics
		errorType := "runtime"
		if strings.Contains(err.Error(), "compile") {
			errorType = "compilation"
		}
		observability.AppleScriptErrorsTotal.WithLabelValues(errorType).Inc()
		observability.AppleScriptExecutionsTotal.WithLabelValues("failure").Inc()
		
		slog.Error("❌ AppleScript execution failed",
			slog.String("task_title", req.Title),
			slog.String("error", err.Error()),
			slog.String("output", string(output)))

		// Record overall task creation duration
		observability.TaskCreationDuration.Observe(time.Since(start).Seconds())
		observability.TaskCreationsTotal.WithLabelValues("failure").Inc()

		return TaskCreateResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript execution failed for task '%s': %v - Output: %s", req.Title, err, string(output)),
		}
	}

	// Parse AppleScript output
	result := strings.TrimSpace(string(output))
	slog.Info("📋 AppleScript result", slog.String("result", result))

	if s.isSuccessResult(result) {
		observability.AppleScriptExecutionsTotal.WithLabelValues("success").Inc()
		observability.TaskCreationDuration.Observe(time.Since(start).Seconds())
		observability.TaskCreationsTotal.WithLabelValues("success").Inc()
		
		slog.Info("✅ Task created successfully", slog.String("task_title", req.Title))
		return TaskCreateResponse{
			Status:  "ok",
			Created: true,
		}
	}

	// AppleScript ran but returned failure
	observability.AppleScriptExecutionsTotal.WithLabelValues("failure").Inc()
	observability.AppleScriptErrorsTotal.WithLabelValues("unknown").Inc()
	observability.TaskCreationDuration.Observe(time.Since(start).Seconds())
	observability.TaskCreationsTotal.WithLabelValues("failure").Inc()
	
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

// sanitizeAppleScriptArg removes characters that could be used for AppleScript injection.
// Since arguments are passed via osascript argv (not shell), the main risk is AppleScript
// string terminators (double quotes and backslashes) that could break out of string context.
func sanitizeAppleScriptArg(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
