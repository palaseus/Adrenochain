package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the explorer API
func SetupRoutes(handler *ExplorerHandler) *mux.Router {
	router := mux.NewRouter()

	// API version prefix
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	router.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Dashboard
	apiV1.HandleFunc("/dashboard", handler.DashboardHandler).Methods("GET")

	// Block operations
	apiV1.HandleFunc("/blocks", handler.BlocksHandler).Methods("GET")
	apiV1.HandleFunc("/blocks/{hash}", handler.BlockDetailsHandler).Methods("GET")

	// Transaction operations
	apiV1.HandleFunc("/transactions", handler.TransactionsHandler).Methods("GET")
	apiV1.HandleFunc("/transactions/{hash}", handler.TransactionDetailsHandler).Methods("GET")

	// Address operations
	apiV1.HandleFunc("/addresses/{address}", handler.AddressDetailsHandler).Methods("GET")

	// Search
	apiV1.HandleFunc("/search", handler.SearchHandler).Methods("GET")

	// Statistics
	apiV1.HandleFunc("/statistics", handler.StatisticsHandler).Methods("GET")

	// Add catch-all OPTIONS handler for CORS preflight
	router.HandleFunc("/{path:.*}", handler.OptionsHandler).Methods("OPTIONS")

	// Add middleware for CORS and logging
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	return router
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		// In production, you'd want proper structured logging
		// For now, we'll just use a simple format

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Log the response
		// log.Printf("%s %s %d", r.Method, r.URL.Path, wrapped.statusCode)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
