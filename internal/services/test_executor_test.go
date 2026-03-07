package services

import (
	"context"
	"strings"
)

// TestAppleScriptExecutor provides a simple mock implementation for testing
type TestAppleScriptExecutor struct {
	// Configurable responses for different test scenarios
	ExecuteResponse       []byte
	ExecuteError          error
	ExecuteSimpleResponse []byte
	ExecuteSimpleError    error
}

// Execute simulates AppleScript execution for file-based scripts
func (t *TestAppleScriptExecutor) Execute(ctx context.Context, script string, args ...string) ([]byte, error) {
	if t.ExecuteError != nil {
		return t.ExecuteResponse, t.ExecuteError
	}

	// Default success response for task creation
	if strings.Contains(script, "omnidrop") || strings.Contains(script, ".applescript") {
		if t.ExecuteResponse != nil {
			return t.ExecuteResponse, nil
		}
		return []byte("success"), nil
	}

	// Default response
	return []byte("test_output"), nil
}

// ExecuteSimple simulates AppleScript execution for inline scripts
func (t *TestAppleScriptExecutor) ExecuteSimple(ctx context.Context, script string) ([]byte, error) {
	if t.ExecuteSimpleError != nil {
		return t.ExecuteSimpleResponse, t.ExecuteSimpleError
	}

	// Simulate different script types
	if strings.Contains(script, "System Events") && strings.Contains(script, "processes") {
		if t.ExecuteSimpleResponse != nil {
			return t.ExecuteSimpleResponse, nil
		}
		// Default: simulate OmniFocus is running
		return []byte("Finder, OmniFocus, Safari"), nil
	}

	// Default response
	return []byte("test_simple_output"), nil
}

// NewTestExecutor creates a new test executor with default success responses
func NewTestExecutor() *TestAppleScriptExecutor {
	return &TestAppleScriptExecutor{
		ExecuteResponse:       []byte("success"),
		ExecuteSimpleResponse: []byte("Finder, OmniFocus, Safari"),
	}
}

// NewTestExecutorWithFailure creates a test executor that simulates AppleScript failures
func NewTestExecutorWithFailure(executeErr error, executeSimpleErr error) *TestAppleScriptExecutor {
	return &TestAppleScriptExecutor{
		ExecuteError:          executeErr,
		ExecuteSimpleError:    executeSimpleErr,
		ExecuteResponse:       []byte("execution failed"),
		ExecuteSimpleResponse: []byte("script failed"),
	}
}

// NewTestExecutorWithOmniFocusNotRunning creates a test executor that simulates OmniFocus not running
func NewTestExecutorWithOmniFocusNotRunning() *TestAppleScriptExecutor {
	return &TestAppleScriptExecutor{
		ExecuteResponse:       []byte("success"),
		ExecuteSimpleResponse: []byte("Finder, Safari, Chrome"), // No OmniFocus
	}
}
