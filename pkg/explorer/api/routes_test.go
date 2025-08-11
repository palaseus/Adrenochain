package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupRoutes(t *testing.T) {
	handler := &ExplorerHandler{}
	router := SetupRoutes(handler)
	
	assert.NotNil(t, router)
	
	// Test that the router is properly configured
	assert.IsType(t, &mux.Router{}, router)
}

func TestCorsMiddleware(t *testing.T) {
	handler := &ExplorerHandler{}
	router := SetupRoutes(handler)
	
	// Test CORS headers are set on a simple route that doesn't require service
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Check CORS headers - note that writeJSONResponse also sets some CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	// The middleware sets "Content-Type, Authorization" but writeJSONResponse overrides with just "Content-Type"
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCorsPreflightRequest(t *testing.T) {
	handler := &ExplorerHandler{}
	router := SetupRoutes(handler)
	
	// Test OPTIONS preflight request
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should return 200 OK for preflight
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check CORS headers are set
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestResponseWriterWrapper(t *testing.T) {
	// Test the responseWriter wrapper functionality
	original := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: original, statusCode: http.StatusOK}
	
	// Test WriteHeader
	wrapped.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, wrapped.statusCode)
	assert.Equal(t, http.StatusNotFound, original.Code)
	
	// Test Write
	testData := []byte("test data")
	n, err := wrapped.Write(testData)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, original.Body.Bytes())
}

func TestRouteRegistration(t *testing.T) {
	// Create a mock service to avoid panics
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)
	router := SetupRoutes(handler)
	
	// Test that all expected routes are registered
	expectedRoutes := []string{
		"/health",
		"/api/v1/dashboard",
		"/api/v1/blocks",
		"/api/v1/transactions",
		"/api/v1/search",
		"/api/v1/statistics",
	}
	
	for _, route := range expectedRoutes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()
		
		// Set up mock expectations for routes that require service
		if route != "/health" {
			switch route {
			case "/api/v1/dashboard":
				mockService.On("GetDashboard", mock.Anything).Return(nil, fmt.Errorf("mock error"))
			case "/api/v1/blocks":
				mockService.On("GetBlocks", mock.Anything, 20, 0).Return(nil, fmt.Errorf("mock error"))
			case "/api/v1/transactions":
				mockService.On("GetTransactions", mock.Anything, 20, 0).Return(nil, fmt.Errorf("mock error"))
			case "/api/v1/search":
				// Search requires a query parameter, so we'll test with a valid query
				req := httptest.NewRequest("GET", "/api/v1/search?q=test", nil)
				w := httptest.NewRecorder()
				mockService.On("Search", mock.Anything, "test").Return(nil, fmt.Errorf("mock error"))
				router.ServeHTTP(w, req)
				assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s not found", route)
				continue
			case "/api/v1/statistics":
				mockService.On("GetStatistics", mock.Anything).Return(nil, fmt.Errorf("mock error"))
			}
		}
		
		router.ServeHTTP(w, req)
		
		// Route should exist (even if handler returns error)
		// We're just checking that the route is registered, not that it works
		// For routes that require service, we expect 500 error, not 404
		if route == "/health" {
			assert.Equal(t, http.StatusOK, w.Code, "Route %s should work", route)
		} else {
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s not found", route)
		}
	}
	
	mockService.AssertExpectations(t)
}

func TestMiddlewareOrder(t *testing.T) {
	handler := &ExplorerHandler{}
	router := SetupRoutes(handler)
	
	// Test that middleware is applied in correct order
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Both CORS and logging middleware should be applied
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	// Note: logging middleware doesn't modify headers, so we can't easily test it
	// but the fact that the request processes means it's working
}

func TestRouterSubrouter(t *testing.T) {
	// Create a mock service to avoid panics
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)
	router := SetupRoutes(handler)
	
	// Test that API v1 subrouter is properly configured
	req := httptest.NewRequest("GET", "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()
	
	// Set up mock expectation
	mockService.On("GetDashboard", mock.Anything).Return(nil, fmt.Errorf("mock error"))
	
	router.ServeHTTP(w, req)
	
	// Should not return 404 (route exists)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	
	// Test that non-API routes still work
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Health route should work
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockService.AssertExpectations(t)
}
