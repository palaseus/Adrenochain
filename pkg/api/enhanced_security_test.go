package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestEnhancedSecurityMiddleware_SecurityMiddleware(t *testing.T) {
	middleware := NewEnhancedSecurityMiddleware()

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with security middleware
	secureHandler := middleware.SecurityMiddleware(testHandler)

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "valid request",
			method:         "GET",
			path:           "/api/test",
			headers:        map[string]string{"Origin": "http://localhost:3000"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "blocked origin",
			method:         "GET",
			path:           "/api/test",
			headers:        map[string]string{"Origin": "http://evil.com"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "malicious query parameter",
			method:         "GET",
			path:           "/api/test?param=<script>alert('xss')</script>",
			headers:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			secureHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("SecurityMiddleware() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestEnhancedSecurityMiddleware_HashValidationMiddleware(t *testing.T) {
	middleware := NewEnhancedSecurityMiddleware()

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with hash validation middleware
	hashHandler := middleware.HashValidationMiddleware(testHandler)

	tests := []struct {
		name           string
		hash           string
		expectedStatus int
	}{
		{
			name:           "valid hash",
			hash:           "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid hash",
			hash:           "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malicious hash",
			hash:           "<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			req = mux.SetURLVars(req, map[string]string{"hash": tt.hash})

			rr := httptest.NewRecorder()
			hashHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("HashValidationMiddleware() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestEnhancedSecurityMiddleware_getClientIP(t *testing.T) {
	middleware := NewEnhancedSecurityMiddleware()

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			remoteAddr: "192.168.1.1:8080",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1, 192.168.1.1"},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "192.168.1.1:8080",
			headers:    map[string]string{"X-Real-IP": "203.0.113.2"},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "203.0.113.3:8080",
			headers:    map[string]string{},
			expectedIP: "203.0.113.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			ip := middleware.getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestEnhancedSecurityMiddleware_validateCORS(t *testing.T) {
	middleware := NewEnhancedSecurityMiddleware()

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{
			name:     "allowed origin",
			origin:   "http://localhost:3000",
			expected: true,
		},
		{
			name:     "blocked origin",
			origin:   "http://evil.com",
			expected: false,
		},
		{
			name:     "no origin header",
			origin:   "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			result := middleware.validateCORS(req)
			if result != tt.expected {
				t.Errorf("validateCORS() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnhancedSecurityMiddleware_addSecurityHeaders(t *testing.T) {
	middleware := NewEnhancedSecurityMiddleware()

	rr := httptest.NewRecorder()
	middleware.addSecurityHeaders(rr)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		if rr.Header().Get(header) != expectedValue {
			t.Errorf("addSecurityHeaders() header %s = %v, want %v",
				header, rr.Header().Get(header), expectedValue)
		}
	}
}

func BenchmarkSecurityMiddleware(b *testing.B) {
	middleware := NewEnhancedSecurityMiddleware()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	secureHandler := middleware.SecurityMiddleware(testHandler)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		secureHandler.ServeHTTP(rr, req)
	}
}
