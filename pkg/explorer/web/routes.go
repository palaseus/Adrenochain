package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupWebRoutes configures all the routes for the explorer web interface
func SetupWebRoutes(handler *WebHandler) *mux.Router {
	router := mux.NewRouter()

	// Web page routes
	router.HandleFunc("/", handler.HomeHandler).Methods("GET")
	router.HandleFunc("/blocks", handler.BlockListHandler).Methods("GET")
	router.HandleFunc("/blocks/{hash}", handler.BlockDetailHandler).Methods("GET")
	router.HandleFunc("/transactions", handler.TransactionListHandler).Methods("GET")
	router.HandleFunc("/transactions/{hash}", handler.TransactionDetailHandler).Methods("GET")
	router.HandleFunc("/addresses/{address}", handler.AddressDetailHandler).Methods("GET")
	router.HandleFunc("/search", handler.SearchHandler).Methods("GET")

	// Static file serving
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("pkg/explorer/web/static"))))

	// API routes (redirect to API endpoints)
	router.PathPrefix("/api/").HandlerFunc(handler.APIHandler)

	// Add middleware for logging and security
	router.Use(webLoggingMiddleware)
	router.Use(securityMiddleware)

	return router
}

// webLoggingMiddleware logs web page requests
func webLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		// In production, you'd want proper structured logging
		// For now, we'll just use a simple format

		// Create a response writer wrapper to capture status code
		wrapped := &webResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Log the response
		// log.Printf("WEB: %s %s %d", r.Method, r.URL.Path, wrapped.statusCode)
	})
}

// securityMiddleware adds security headers to web responses
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// webResponseWriter wraps http.ResponseWriter to capture status code
type webResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *webResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *webResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
