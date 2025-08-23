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

// TestTemplateHelperFunctions tests the template helper functions
func TestTemplateHelperFunctions(t *testing.T) {
	t.Run("formatHash", func(t *testing.T) {
		// Test empty hash
		result := formatHash([]byte{})
		assert.Equal(t, "N/A", result)

		// Test short hash
		result = formatHash([]byte{1, 2, 3, 4})
		assert.Equal(t, "01020304", result)

		// Test long hash
		longHash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		result = formatHash(longHash)
		assert.Equal(t, "01020304...090a0b0c", result)
	})

	t.Run("formatAddress", func(t *testing.T) {
		// Test short address
		result := formatAddress("123456789")
		assert.Equal(t, "123456789", result)

		// Test long address
		longAddr := "1234567890123456789012345678901234567890"
		result = formatAddress(longAddr)
		assert.Equal(t, "12345678...7890", result)
	})

	t.Run("formatAmount", func(t *testing.T) {
		// Test zero amount
		result := formatAmount(0)
		assert.Equal(t, "0", result)

		// Test small amount
		result = formatAmount(100000000) // 1.00000000
		assert.Equal(t, "1.00000000", result)

		// Test larger amount
		result = formatAmount(123456789000000000) // 1234567890.00000000
		assert.Equal(t, "1234567890.00000000", result)
	})

	t.Run("formatTime", func(t *testing.T) {
		// Test int64 timestamp
		result := formatTime(int64(1640995200))
		assert.Equal(t, "1640995200", result)

		// Test string time
		result = formatTime("2022-01-01")
		assert.Equal(t, "2022-01-01", result)

		// Test other type
		result = formatTime(123.45)
		assert.Equal(t, "123.45", result)
	})

	t.Run("formatDifficulty", func(t *testing.T) {
		// Test zero difficulty
		result := formatDifficulty(0)
		assert.Equal(t, "0", result)

		// Test small difficulty
		result = formatDifficulty(1500)
		assert.Equal(t, "1.50 K", result)

		// Test medium difficulty
		result = formatDifficulty(2500000)
		assert.Equal(t, "2.50 M", result)

		// Test large difficulty
		result = formatDifficulty(1500000000)
		assert.Equal(t, "1.50 G", result)

		// Test very small difficulty
		result = formatDifficulty(500)
		assert.Equal(t, "500", result)
	})

	t.Run("math_operations", func(t *testing.T) {
		// Test addition
		assert.Equal(t, 5, add(2, 3))
		assert.Equal(t, -1, add(-2, 1))

		// Test subtraction
		assert.Equal(t, 1, sub(3, 2))
		assert.Equal(t, -5, sub(-2, 3))

		// Test multiplication
		assert.Equal(t, 6, mul(2, 3))
		assert.Equal(t, -6, mul(2, -3))

		// Test division
		assert.Equal(t, 2, div(6, 3))
		assert.Equal(t, 0, div(6, 0)) // Division by zero returns 0

		// Test modulo
		assert.Equal(t, 1, mod(7, 3))
		assert.Equal(t, 0, mod(6, 0)) // Modulo by zero returns 0
	})
}
