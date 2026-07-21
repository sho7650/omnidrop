package services

import (
	"fmt"
	"os"
	"testing"

	"omnidrop/internal/config"
)

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
	var _ = service
}

func TestNewHealthServiceWithExecutor(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
		Port:  "8788",
	}

	service := NewHealthServiceWithExecutor(cfg, NewTestExecutor())

	if service == nil {
		t.Error("NewHealthServiceWithExecutor returned nil")
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

	service := NewHealthServiceWithExecutor(cfg, NewTestExecutor())
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

func TestHealthServiceImpl_CheckOmniFocusStatus_WithTestExecutor(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
		Port:  "8788",
	}

	// NewTestExecutor's default ExecuteSimpleResponse includes "OmniFocus" in the process list.
	service := NewHealthServiceWithExecutor(cfg, NewTestExecutor())
	impl := service.(*HealthServiceImpl)

	if !impl.CheckOmniFocusStatus() {
		t.Error("Expected CheckOmniFocusStatus to return true when executor reports OmniFocus running")
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
		OmniFocusRunning:      false,
		ScriptPath:            "/test/path",
		Errors:                []string{"test error"},
		Details:               "test details",
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

// Helper functions for test script management
func createTempTestScript(path string) error {
	content := `#!/usr/bin/osascript
# Temporary test script
return "test"`
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(content)
	return err
}

func removeTempTestScript(path string) {
	_ = os.Remove(path)
}

// Tests using the new TestAppleScriptExecutor

func TestHealthServiceImpl_CheckAppleScriptHealth_Success(t *testing.T) {
	// Create a temporary test script
	testScript := "./temp_test_script.applescript"
	err := createTempTestScript(testScript)
	if err != nil {
		t.Fatalf("Failed to create temp test script: %v", err)
	}
	defer removeTempTestScript(testScript)

	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      testScript,
		AppleScriptFile: "temp_test_script.applescript",
	}

	// Use the new TestExecutor
	testExecutor := NewTestExecutor()
	service := NewHealthServiceWithExecutor(cfg, testExecutor)
	result := service.CheckAppleScriptHealth()

	if !result.AppleScriptAccessible {
		t.Error("Expected AppleScript to be accessible with TestExecutor")
	}

	if !result.OmniFocusRunning {
		t.Error("Expected OmniFocus to be running with default TestExecutor")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, but got: %v", result.Errors)
	}

	if result.ScriptPath != testScript {
		t.Errorf("Expected script path %s, got %s", testScript, result.ScriptPath)
	}
}

func TestHealthServiceImpl_CheckAppleScriptHealth_OmniFocusNotRunning(t *testing.T) {
	// Create a temporary test script
	testScript := "./temp_test_script.applescript"
	err := createTempTestScript(testScript)
	if err != nil {
		t.Fatalf("Failed to create temp test script: %v", err)
	}
	defer removeTempTestScript(testScript)

	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      testScript,
		AppleScriptFile: "temp_test_script.applescript",
	}

	// Use TestExecutor that simulates OmniFocus not running
	testExecutor := NewTestExecutorWithOmniFocusNotRunning()
	service := NewHealthServiceWithExecutor(cfg, testExecutor)
	result := service.CheckAppleScriptHealth()

	if !result.AppleScriptAccessible {
		t.Error("Expected AppleScript to be accessible")
	}

	if result.OmniFocusRunning {
		t.Error("Expected OmniFocus to NOT be running with this TestExecutor")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, but got: %v", result.Errors)
	}

	expectedDetails := "AppleScript accessible but OmniFocus not running"
	if result.Details != expectedDetails {
		t.Errorf("Expected details '%s', got '%s'", expectedDetails, result.Details)
	}
}

func TestHealthServiceImpl_CheckAppleScriptHealth_ExecutionFailure(t *testing.T) {
	// Create a temporary test script
	testScript := "./temp_test_script.applescript"
	err := createTempTestScript(testScript)
	if err != nil {
		t.Fatalf("Failed to create temp test script: %v", err)
	}
	defer removeTempTestScript(testScript)

	cfg := &config.Config{
		Token:           "test-token",
		Port:            "8788",
		Environment:     "test",
		ScriptPath:      testScript,
		AppleScriptFile: "temp_test_script.applescript",
	}

	// Use TestExecutor that simulates execution failure
	testExecutor := NewTestExecutorWithFailure(nil, fmt.Errorf("AppleScript execution failed"))
	service := NewHealthServiceWithExecutor(cfg, testExecutor)
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
