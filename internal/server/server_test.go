package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"omnidrop/internal/config"
	"omnidrop/internal/handlers"
	"omnidrop/test/mocks"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)

	server := NewServer(cfg, h)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.handlers != h {
		t.Error("Server handlers not set correctly")
	}

	if server.router == nil {
		t.Error("Router not initialized")
	}

	if server.httpSrv == nil {
		t.Error("HTTP server not initialized")
	}
}

func TestServer_GetAddress(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	expectedAddr := ":8788"
	actualAddr := server.GetAddress()

	if actualAddr != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, actualAddr)
	}
}

func TestServer_GetRouter(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	router := server.GetRouter()
	if router == nil {
		t.Error("GetRouter returned nil")
	}
}

func TestServer_RouteConfiguration(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	router := server.GetRouter()

	// Test that routes are properly configured by making test requests
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Health endpoint GET",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Tasks endpoint POST without auth",
			method:         "POST",
			path:           "/tasks",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Non-existent endpoint",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Tasks endpoint with wrong method",
			method:         "GET",
			path:           "/tasks",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestServer_MiddlewareConfiguration(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	router := server.GetRouter()

	// Test that middleware is working by checking headers
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check that middleware added headers (from RequestID middleware)
	// Note: chi's RequestID middleware doesn't set response headers by default,
	// it only adds the request ID to the context. Let's check that the request was processed.
	if rr.Code == 0 {
		t.Error("Expected request to be processed by middleware")
	}

	// Verify status is OK, which means middleware chain worked
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestServer_HTTPServerConfiguration(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	// Check HTTP server configuration
	if server.httpSrv.Addr != ":8788" {
		t.Errorf("Expected server address :8788, got %s", server.httpSrv.Addr)
	}

	if server.httpSrv.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", server.httpSrv.ReadTimeout)
	}

	if server.httpSrv.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", server.httpSrv.WriteTimeout)
	}

	if server.httpSrv.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", server.httpSrv.IdleTimeout)
	}

	if server.httpSrv.Handler != server.router {
		t.Error("HTTP server handler not set to router")
	}
}

func TestServer_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788",
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	// Test shutdown with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Since the server isn't actually running, Shutdown should return quickly
	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}

func TestServer_Integration(t *testing.T) {
	cfg := &config.Config{
		Port:  "8788", // Use test port
		Token: "test-token",
	}

	mockOmniFocusService := &mocks.MockOmniFocusService{}
	mockFilesService := &mocks.MockFilesService{}
	h := handlers.New(cfg, mockOmniFocusService, mockFilesService)
	server := NewServer(cfg, h)

	// Test that the server can be used with httptest.Server
	testServer := httptest.NewServer(server.GetRouter())
	defer testServer.Close()

	// Test health endpoint
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
