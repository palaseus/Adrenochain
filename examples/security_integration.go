//go:build examples
// +build examples

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/palaseus/adrenochain/pkg/api"
)

// This example demonstrates how to integrate the enhanced security features
// into the existing Adrenochain API server

func main() {
	fmt.Println("🔒 Adrenochain Enhanced Security Integration Example")

	// Create enhanced security middleware
	securityMiddleware := api.NewEnhancedSecurityMiddleware()

	// Create router
	router := mux.NewRouter()

	// Apply security middleware to all routes
	router.Use(securityMiddleware.SecurityMiddleware)

	// Apply hash validation middleware to hash-based routes
	hashRouter := router.PathPrefix("/api/v1").Subrouter()
	hashRouter.Use(securityMiddleware.HashValidationMiddleware)

	// Example routes with enhanced security
	setupSecureRoutes(hashRouter)

	// Start server with security enhancements
	fmt.Println("✅ Starting secure API server on :8080")
	fmt.Println("🛡️  Security features enabled:")
	fmt.Println("   - Input validation and sanitization")
	fmt.Println("   - Rate limiting")
	fmt.Println("   - CORS protection")
	fmt.Println("   - Hash validation")
	fmt.Println("   - Security headers")
	fmt.Println("   - IP blocking")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupSecureRoutes(router *mux.Router) {
	// Secure block endpoints
	router.HandleFunc("/blocks/{hash}", getBlockHandlerSecure).Methods("GET")
	router.HandleFunc("/transactions/{hash}", getTransactionHandlerSecure).Methods("GET")

	// Other secure endpoints
	router.HandleFunc("/health", healthHandlerSecure).Methods("GET")
	router.HandleFunc("/status", statusHandlerSecure).Methods("GET")
}

func getBlockHandlerSecure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	// Hash is already validated by middleware at this point
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Block retrieved securely", "hash": "%s"}`, hash)
}

func getTransactionHandlerSecure(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	// Hash is already validated by middleware at this point
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Transaction retrieved securely", "hash": "%s"}`, hash)
}

func healthHandlerSecure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "healthy", "security": "enhanced"}`)
}

func statusHandlerSecure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "operational", "security_level": "high"}`)
}
