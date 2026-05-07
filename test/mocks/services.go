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
