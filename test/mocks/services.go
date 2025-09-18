package mocks

import (
	"context"

	"omnidrop/internal/services"
)

// MockOmniFocusService provides a mock implementation for testing
type MockOmniFocusService struct {
	CreateTaskFunc func(ctx context.Context, req services.TaskCreateRequest) services.TaskCreateResponse
}

func (m *MockOmniFocusService) CreateTask(ctx context.Context, req services.TaskCreateRequest) services.TaskCreateResponse {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, req)
	}
	return services.TaskCreateResponse{
		Status:  "ok",
		Created: true,
	}
}

// MockAppleScriptExecutor provides a mock implementation for testing
type MockAppleScriptExecutor struct {
	ExecuteFunc func(ctx context.Context, script string, args ...string) ([]byte, error)
}

func (m *MockAppleScriptExecutor) Execute(ctx context.Context, script string, args ...string) ([]byte, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, script, args...)
	}
	return []byte("mock_success"), nil
}

// MockHealthService provides a mock implementation for testing
type MockHealthService struct {
	CheckAppleScriptHealthFunc func() services.HealthResult
	CheckOmniFocusStatusFunc   func() bool
}

func (m *MockHealthService) CheckAppleScriptHealth() services.HealthResult {
	if m.CheckAppleScriptHealthFunc != nil {
		return m.CheckAppleScriptHealthFunc()
	}
	return services.HealthResult{
		AppleScriptAccessible: true,
		OmniFocusRunning:      true,
		ScriptPath:            "/mock/path/omnidrop.applescript",
		Details:               "Mock health check successful",
	}
}

func (m *MockHealthService) CheckOmniFocusStatus() bool {
	if m.CheckOmniFocusStatusFunc != nil {
		return m.CheckOmniFocusStatusFunc()
	}
	return true
}

// MockFilesService provides a mock implementation for testing
type MockFilesService struct {
	WriteFileFunc func(ctx context.Context, req services.FileWriteRequest) services.FileWriteResponse
}

func (m *MockFilesService) WriteFile(ctx context.Context, req services.FileWriteRequest) services.FileWriteResponse {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(ctx, req)
	}
	return services.FileWriteResponse{
		Status:  "ok",
		Created: true,
		Path:    req.Filename,
	}
}
