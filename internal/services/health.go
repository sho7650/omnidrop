package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"omnidrop/internal/config"
)

// DefaultAppleScriptExecutor provides the default implementation for AppleScript execution
type DefaultAppleScriptExecutor struct{}

func (e *DefaultAppleScriptExecutor) Execute(ctx context.Context, script string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "osascript", append([]string{script}, args...)...)
	return cmd.CombinedOutput()
}

// ExecuteSimple executes a simple AppleScript command
func (e *DefaultAppleScriptExecutor) ExecuteSimple(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	return cmd.CombinedOutput()
}

// HealthServiceImpl implements the HealthService interface
type HealthServiceImpl struct {
	config   *config.Config
	executor AppleScriptExecutor
}

// NewHealthService creates a new health service with the default AppleScript executor
func NewHealthService(cfg *config.Config) HealthService {
	return &HealthServiceImpl{
		config:   cfg,
		executor: &DefaultAppleScriptExecutor{},
	}
}

// NewHealthServiceWithExecutor creates a new health service with a custom executor (for testing)
func NewHealthServiceWithExecutor(cfg *config.Config, executor AppleScriptExecutor) HealthService {
	return &HealthServiceImpl{
		config:   cfg,
		executor: executor,
	}
}

// CheckAppleScriptHealth performs comprehensive AppleScript health checks
func (h *HealthServiceImpl) CheckAppleScriptHealth() HealthResult {
	result := HealthResult{
		AppleScriptAccessible: false,
		OmniFocusRunning:      false,
		Errors:                []string{},
	}

	slog.Info("🍎 Testing AppleScript access...")

	// Check if AppleScript file exists and get its path
	scriptPath, err := h.config.GetAppleScriptPath()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("AppleScript file not found: AppleScript path resolution failed: %v", err))
		result.Details = "AppleScript file not found in expected locations"
		slog.Error("AppleScript file not found", slog.String("error", err.Error()))
		return result
	}

	result.ScriptPath = scriptPath
	slog.Info("✅ AppleScript found", slog.String("script_path", scriptPath))

	// Test basic AppleScript execution
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := h.testBasicAppleScriptAccess(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("AppleScript test failed: AppleScript accessibility test failed: %v", err))
		result.Details = fmt.Sprintf("AppleScript execution failed: %s", string(output))
		slog.Error("AppleScript test failed", slog.String("error", err.Error()))
		return result
	}

	result.AppleScriptAccessible = true
	slog.Info("✅ AppleScript accessibility confirmed")

	// Check if OmniFocus is available
	result.OmniFocusRunning = h.CheckOmniFocusStatus()
	if result.OmniFocusRunning {
		slog.Info("✅ OmniFocus detected in running processes")
		result.Details = "AppleScript and OmniFocus are both accessible"
	} else {
		slog.Warn("⚠️ OmniFocus not currently running")
		result.Details = "AppleScript accessible but OmniFocus not running"
	}

	return result
}

// CheckOmniFocusStatus checks if OmniFocus is currently running
func (h *HealthServiceImpl) CheckOmniFocusStatus() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := h.executor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
	if err != nil {
		slog.Error("Failed to check running processes", slog.String("error", err.Error()))
		return false
	}
	return strings.Contains(string(output), "OmniFocus")
}

// testBasicAppleScriptAccess tests basic AppleScript functionality
func (h *HealthServiceImpl) testBasicAppleScriptAccess(ctx context.Context) ([]byte, error) {
	return h.executor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
}

// GetWorkingDirectory returns the current working directory (moved from main.go)
func GetWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
