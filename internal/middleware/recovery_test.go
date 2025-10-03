package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		expectPanic    bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
			expectPanic:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "panic with string",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			expectPanic:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":"error","message":"internal server error","code":"internal_error"}`,
		},
		{
			name: "panic with error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrAbortHandler)
			}),
			expectPanic:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":"error","message":"internal server error","code":"internal_error"}`,
		},
		{
			name: "panic with nil",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(nil)
			}),
			expectPanic:    true, // panic(nil) does actually panic in Go
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":"error","message":"internal server error","code":"internal_error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Wrap handler with recovery middleware
			handler := Recovery(tt.handler)

			// Execute request
			handler.ServeHTTP(rec, req)

			// Check status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", rec.Code, tt.expectedStatus)
			}

			// Check body if expected
			if tt.expectedBody != "" {
				body := rec.Body.String()
				if body != tt.expectedBody {
					t.Errorf("Body = %v, want %v", body, tt.expectedBody)
				}
			}

			// Check Content-Type for error responses
			if tt.expectPanic {
				contentType := rec.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Content-Type = %v, want application/json", contentType)
				}
			}
		})
	}
}

func TestRecovery_PreservesHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	Recovery(handler).ServeHTTP(rec, req)

	// Recovery should set Content-Type header
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type not set correctly after recovery")
	}

	// Custom header set before panic should still be present
	if rec.Header().Get("X-Custom-Header") != "test-value" {
		t.Errorf("Custom header lost after recovery")
	}
}

func TestRecovery_WithChainedMiddleware(t *testing.T) {
	// Simulate a middleware chain
	requestIDMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Request-ID", "test-123")
			next.ServeHTTP(w, r)
		})
	}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify middleware ran before panic
		if r.Header.Get("X-Request-ID") != "test-123" {
			t.Error("Middleware chain broken")
		}
		panic("test panic in chain")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Chain: Recovery -> RequestID -> Handler
	handler := Recovery(requestIDMiddleware(panicHandler))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusInternalServerError)
	}
}
