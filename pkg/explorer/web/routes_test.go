package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetupWebRoutes tests route setup functionality
func TestSetupWebRoutes(t *testing.T) {
	mockService := &MockExplorerService{}
	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	router := SetupWebRoutes(handler)
	assert.NotNil(t, router)

	// Test that routes are properly configured
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test API route redirect
	req = httptest.NewRequest("GET", "/api/v1/blocks", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
}

// TestWebLoggingMiddleware tests the logging middleware
func TestWebLoggingMiddleware(t *testing.T) {
	// Create a test handler that sets a custom status
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("test response"))
	})

	// Apply the middleware
	middleware := webLoggingMiddleware(testHandler)

	// Test the middleware
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

// TestSecurityMiddleware tests the security middleware
func TestSecurityMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	// Apply the middleware
	middleware := securityMiddleware(testHandler)

	// Test the middleware
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))

	assert.Equal(t, "test response", w.Body.String())
}

// TestWebResponseWriter tests the response writer wrapper
func TestWebResponseWriter(t *testing.T) {
	// Create a mock response writer
	mockW := httptest.NewRecorder()
	wrapped := &webResponseWriter{ResponseWriter: mockW, statusCode: http.StatusOK}

	// Test WriteHeader
	wrapped.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, wrapped.statusCode)
	assert.Equal(t, http.StatusNotFound, mockW.Code)

	// Test Write
	testData := []byte("test data")
	n, err := wrapped.Write(testData)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, mockW.Body.Bytes())
}

// TestStaticFileServing tests static file serving
func TestStaticFileServing(t *testing.T) {
	mockService := &MockExplorerService{}
	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	router := SetupWebRoutes(handler)

	// Test static file route
	req := httptest.NewRequest("GET", "/static/css/style.css", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should either serve the file or return 404
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

// TestRouteMethods tests that routes only accept correct HTTP methods
func TestRouteMethods(t *testing.T) {
	mockService := &MockExplorerService{}
	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	router := SetupWebRoutes(handler)

	// Test that POST to GET-only routes returns 405 Method Not Allowed
	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// Test that PUT to GET-only routes returns 405 Method Not Allowed
	req = httptest.NewRequest("PUT", "/blocks", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestRouteNotFound tests 404 handling
func TestRouteNotFound(t *testing.T) {
	mockService := &MockExplorerService{}
	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	router := SetupWebRoutes(handler)

	// Test non-existent route
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
