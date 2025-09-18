package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
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

	log.Printf("🍎 Testing AppleScript access...")

	// Check if AppleScript file exists and get its path
	scriptPath, err := h.config.GetAppleScriptPath()
	if err != nil {
		wrappedErr := errors.Wrap(err, "AppleScript path resolution failed")
		result.Errors = append(result.Errors, fmt.Sprintf("AppleScript file not found: %v", wrappedErr))
		result.Details = "AppleScript file not found in expected locations"
		log.Printf("❌ AppleScript file not found: %v", wrappedErr)
		return result
	}

	result.ScriptPath = scriptPath
	log.Printf("✅ AppleScript found: %s", scriptPath)

	// Test basic AppleScript execution
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := h.testBasicAppleScriptAccess(ctx)
	if err != nil {
		wrappedErr := errors.Wrap(err, "AppleScript accessibility test failed")
		result.Errors = append(result.Errors, fmt.Sprintf("AppleScript test failed: %v", wrappedErr))
		result.Details = fmt.Sprintf("AppleScript execution failed: %s", string(output))
		log.Printf("❌ AppleScript test failed: %v", wrappedErr)
		return result
	}

	result.AppleScriptAccessible = true
	log.Printf("✅ AppleScript accessibility confirmed")

	// Check if OmniFocus is available
	result.OmniFocusRunning = h.CheckOmniFocusStatus()
	if result.OmniFocusRunning {
		log.Printf("✅ OmniFocus detected in running processes")
		result.Details = "AppleScript and OmniFocus are both accessible"
	} else {
		log.Printf("⚠️ OmniFocus not currently running")
		result.Details = "AppleScript accessible but OmniFocus not running"
	}

	return result
}

// CheckOmniFocusStatus checks if OmniFocus is currently running
func (h *HealthServiceImpl) CheckOmniFocusStatus() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if we have a default executor with ExecuteSimple method
	if executor, ok := h.executor.(*DefaultAppleScriptExecutor); ok {
		output, err := executor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
		if err != nil {
			wrappedErr := errors.Wrap(err, "failed to query system processes")
			log.Printf("❌ Failed to check running processes: %v", wrappedErr)
			return false
		}
		return strings.Contains(string(output), "OmniFocus")
	}

	// For TestExecutor and other mock executors, use ExecuteSimple if available
	if testExecutor, ok := h.executor.(*TestAppleScriptExecutor); ok {
		output, err := testExecutor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
		if err != nil {
			wrappedErr := errors.Wrap(err, "test executor failed to query processes")
			log.Printf("❌ Failed to check running processes: %v", wrappedErr)
			return false
		}
		return strings.Contains(string(output), "OmniFocus")
	}

	// For other mock executors, return false (backward compatibility)
	return false
}

// testBasicAppleScriptAccess tests basic AppleScript functionality
func (h *HealthServiceImpl) testBasicAppleScriptAccess(ctx context.Context) ([]byte, error) {
	// Check if we have a default executor with ExecuteSimple method
	if executor, ok := h.executor.(*DefaultAppleScriptExecutor); ok {
		return executor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
	}

	// For TestExecutor, use ExecuteSimple method
	if testExecutor, ok := h.executor.(*TestAppleScriptExecutor); ok {
		return testExecutor.ExecuteSimple(ctx, "tell application \"System Events\" to get name of processes")
	}

	// For other executor types (mocks), use the standard Execute method
	return h.executor.Execute(ctx, "-e", "tell application \"System Events\" to get name of processes")
}

// GetWorkingDirectory returns the current working directory (moved from main.go)
func GetWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
