package services

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"omnidrop/internal/config"
)

type OmniFocusService struct {
	cfg *config.Config
}

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
			Reason: fmt.Sprintf("AppleScript path error: %v", err),
		}
	}

	// Prepare arguments for direct passing to AppleScript
	tagsString := strings.Join(req.Tags, ",")

	// Execute AppleScript with direct arguments
	log.Printf("📝 Creating OmniFocus task: %s", req.Title)
	log.Printf("🍎 Executing AppleScript: %s", scriptPath)
	log.Printf("📋 Arguments: title='%s', note='%s', project='%s', tags='%s'",
		req.Title, req.Note, req.Project, tagsString)

	// Create command with context for timeout/cancellation
	cmd := exec.CommandContext(ctx, "osascript", scriptPath, req.Title, req.Note, req.Project, tagsString)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ AppleScript execution failed: %v", err)
		log.Printf("📄 AppleScript output: %s", string(output))
		return TaskCreateResponse{
			Status: "error",
			Reason: fmt.Sprintf("AppleScript execution failed: %v - Output: %s", err, string(output)),
		}
	}

	// Parse AppleScript output
	result := strings.TrimSpace(string(output))
	log.Printf("📋 AppleScript result: %s", result)

	if s.isSuccessResult(result) {
		log.Printf("✅ Task created successfully: %s", req.Title)
		return TaskCreateResponse{
			Status:  "ok",
			Created: true,
		}
	}

	log.Printf("❌ Task creation failed - AppleScript returned: %s", result)
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
	log.Printf("🔍 Checking result: '%s' against success patterns", result)

	// Check if the last line matches any success pattern
	lines := strings.Split(result, "\n")
	lastLine := strings.ToLower(strings.TrimSpace(lines[len(lines)-1]))

	for _, pattern := range successPatterns {
		if strings.ToLower(pattern) == lastLine {
			log.Printf("✅ Match found: last line '%s' matches pattern '%s'", lastLine, pattern)
			return true
		}
	}

	// Fallback: check if any line contains a success pattern
	resultLower := strings.ToLower(result)
	for _, pattern := range successPatterns {
		if strings.Contains(resultLower, strings.ToLower(pattern)) {
			log.Printf("✅ Match found: result contains pattern '%s'", pattern)
			return true
		}
	}

	log.Printf("❌ No success pattern matched for result: '%s'", result)
	return false
}