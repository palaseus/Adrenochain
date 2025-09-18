package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/palaseus/adrenochain/pkg/security"
)

// EnhancedSecurityMiddleware provides comprehensive API security
type EnhancedSecurityMiddleware struct {
	validator      *security.EnhancedSecurityValidator
	rateLimiter    *security.SecurityRateLimitTracker
	blockedIPs     map[string]time.Time
	allowedOrigins []string
}

// NewEnhancedSecurityMiddleware creates new enhanced security middleware
func NewEnhancedSecurityMiddleware() *EnhancedSecurityMiddleware {
	return &EnhancedSecurityMiddleware{
		validator:      security.NewEnhancedSecurityValidator(nil),
		rateLimiter:    security.NewSecurityRateLimitTracker(1*time.Minute, 100),
		blockedIPs:     make(map[string]time.Time),
		allowedOrigins: []string{"http://localhost:3000", "https://adrenochain.io"},
	}
}

// SecurityMiddleware provides comprehensive security for API endpoints
func (esm *EnhancedSecurityMiddleware) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := esm.getClientIP(r)

		// Check if IP is blocked
		if esm.isIPBlocked(clientIP) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Rate limiting
		if !esm.rateLimiter.IsAllowed(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// CORS validation
		if !esm.validateCORS(r) {
			http.Error(w, "CORS validation failed", http.StatusForbidden)
			return
		}

		// Input validation for parameters
		if err := esm.validateRequestInputs(r); err != nil {
			http.Error(w, fmt.Sprintf("Input validation failed: %v", err), http.StatusBadRequest)
			return
		}

		// Add security headers
		esm.addSecurityHeaders(w)

		next.ServeHTTP(w, r)
	})
}

// HashValidationMiddleware validates hash parameters
func (esm *EnhancedSecurityMiddleware) HashValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		// Validate hash parameter
		if hashHex, exists := vars["hash"]; exists {
			_, err := esm.validator.ValidateHash(hashHex)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid hash: %v", err), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Helper methods
func (esm *EnhancedSecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (esm *EnhancedSecurityMiddleware) isIPBlocked(ip string) bool {
	if blockTime, exists := esm.blockedIPs[ip]; exists {
		if time.Since(blockTime) < 24*time.Hour {
			return true
		}
		delete(esm.blockedIPs, ip)
	}
	return false
}

func (esm *EnhancedSecurityMiddleware) validateCORS(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // No origin header, allow
	}

	for _, allowed := range esm.allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (esm *EnhancedSecurityMiddleware) validateRequestInputs(r *http.Request) error {
	// Validate query parameters
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if err := esm.validator.ValidateInput(value); err != nil {
				return fmt.Errorf("invalid query parameter %s: %w", key, err)
			}
		}
	}

	// Validate path parameters
	vars := mux.Vars(r)
	for key, value := range vars {
		if err := esm.validator.ValidateInput(value); err != nil {
			return fmt.Errorf("invalid path parameter %s: %w", key, err)
		}
	}

	return nil
}

func (esm *EnhancedSecurityMiddleware) addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}
