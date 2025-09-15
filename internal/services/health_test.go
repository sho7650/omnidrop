package services

import (
	"context"
	"fmt"
	"testing"

	"omnidrop/internal/config"
)

// MockExecutorForTesting provides a simple mock for health service tests
type MockExecutorForTesting struct {
	executeFunc       func(ctx context.Context, script string, args ...string) ([]byte, error)
	executeSimpleFunc func(ctx context.Context, script string) ([]byte, error)
}

func (m *MockExecutorForTesting) Execute(ctx context.Context, script string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, script, args...)
	}
	return []byte("mock_success"), nil
}

func (m *MockExecutorForTesting) ExecuteSimple(ctx context.Context, script string) ([]byte, error) {
	if m.executeSimpleFunc != nil {
		return m.executeSimpleFunc(ctx, script)
	}
	return []byte("mock_processes"), nil
}

func TestNewHealthService(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
		Port:  "8788",
	}

	service := NewHealthService(cfg)
	if service == nil {
		t.Error("NewHealthService returned nil")
	}

	// Verify it implements the interface
	var _ HealthService = service
}

func TestNewHealthServiceWithExecutor(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
		Port:  "8788",
	}

	mockExecutor := &MockExecutorForTesting{}
	service := NewHealthServiceWithExecutor(cfg, mockExecutor)

	if service == nil {
		t.Error("NewHealthServiceWithExecutor returned nil")
	}

	// Verify it implements the interface
	var _ HealthService = service
}

func TestHealthServiceImpl_CheckAppleScriptHealth_Success(t *testing.T) {
	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      "./test_script.applescript",
		AppleScriptFile: "test_script.applescript",
	}

	// Create a mock executor that simulates success
	mockExecutor := &MockExecutorForTesting{
		executeSimpleFunc: func(ctx context.Context, script string) ([]byte, error) {
			return []byte("System Events"), nil
		},
	}

	service := NewHealthServiceWithExecutor(cfg, mockExecutor)
	result := service.CheckAppleScriptHealth()

	if !result.AppleScriptAccessible {
		t.Error("Expected AppleScript to be accessible with mock executor")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, but got: %v", result.Errors)
	}

	if result.Details == "" {
		t.Error("Expected details to be set")
	}
}

func TestHealthServiceImpl_CheckAppleScriptHealth_ScriptNotFound(t *testing.T) {
	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      "/nonexistent/path/script.applescript",
		AppleScriptFile: "script.applescript",
	}

	mockExecutor := &MockExecutorForTesting{}
	service := NewHealthServiceWithExecutor(cfg, mockExecutor)
	result := service.CheckAppleScriptHealth()

	if result.AppleScriptAccessible {
		t.Error("Expected AppleScript to be inaccessible when script file doesn't exist")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors when script file doesn't exist")
	}

	if result.ScriptPath != "" {
		t.Error("Expected script path to be empty when file doesn't exist")
	}
}

func TestHealthServiceImpl_CheckAppleScriptHealth_ExecutionFailure(t *testing.T) {
	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      "./test_script.applescript",
		AppleScriptFile: "test_script.applescript",
	}

	// Create a mock executor that simulates execution failure
	mockExecutor := &MockExecutorForTesting{
		executeSimpleFunc: func(ctx context.Context, script string) ([]byte, error) {
			return []byte("execution error"), fmt.Errorf("mock execution failure")
		},
	}

	service := NewHealthServiceWithExecutor(cfg, mockExecutor)
	result := service.CheckAppleScriptHealth()

	if result.AppleScriptAccessible {
		t.Error("Expected AppleScript to be inaccessible when execution fails")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors when AppleScript execution fails")
	}

	if result.Details == "" {
		t.Error("Expected details to be set when execution fails")
	}
}

func TestHealthServiceImpl_CheckOmniFocusStatus_WithMockExecutor(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
		Port:  "8788",
	}

	// Test with OmniFocus present
	mockExecutor := &MockExecutorForTesting{
		executeSimpleFunc: func(ctx context.Context, script string) ([]byte, error) {
			return []byte("Finder, OmniFocus, Safari"), nil
		},
	}

	service := NewHealthServiceWithExecutor(cfg, mockExecutor)
	impl := service.(*HealthServiceImpl)

	// For mock executors, CheckOmniFocusStatus should return false
	// since it can't check real processes
	result := impl.CheckOmniFocusStatus()
	if result {
		t.Error("Expected CheckOmniFocusStatus to return false for mock executor")
	}
}

func TestGetWorkingDirectory(t *testing.T) {
	dir := GetWorkingDirectory()
	if dir == "" {
		t.Error("GetWorkingDirectory returned empty string")
	}
	if dir == "unknown" {
		t.Log("Working directory is unknown - this might be expected in some test environments")
	}
}

func TestHealthResult_Structure(t *testing.T) {
	result := HealthResult{
		AppleScriptAccessible: true,
		OmniFocusRunning:     false,
		ScriptPath:           "/test/path",
		Errors:              []string{"test error"},
		Details:             "test details",
	}

	if !result.AppleScriptAccessible {
		t.Error("Expected AppleScriptAccessible to be true")
	}

	if result.OmniFocusRunning {
		t.Error("Expected OmniFocusRunning to be false")
	}

	if result.ScriptPath != "/test/path" {
		t.Errorf("Expected ScriptPath to be '/test/path', got %s", result.ScriptPath)
	}

	if len(result.Errors) != 1 || result.Errors[0] != "test error" {
		t.Errorf("Expected one error 'test error', got %v", result.Errors)
	}

	if result.Details != "test details" {
		t.Errorf("Expected Details to be 'test details', got %s", result.Details)
	}
}