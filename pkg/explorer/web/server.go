package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/palaseus/adrenochain/pkg/explorer/service"
)

// WebServer manages the explorer web interface
type WebServer struct {
	explorerService service.ExplorerService
	templates       *Templates
	handler         *WebHandler
	server          *http.Server
}

// NewWebServer creates a new web server instance
func NewWebServer(explorerService service.ExplorerService) *WebServer {
	templates := NewTemplates()
	handler := NewWebHandler(explorerService, templates)

	return &WebServer{
		explorerService: explorerService,
		templates:       templates,
		handler:         handler,
	}
}

// Start starts the web server on the specified port
func (ws *WebServer) Start(port int) error {
	// Setup web routes
	webRouter := SetupWebRoutes(ws.handler)

	// Create HTTP server
	ws.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      webRouter,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting GoChain Explorer web interface on port %d", port)
	log.Printf("Web interface available at: http://localhost:%d", port)
	log.Printf("API endpoints available at: http://localhost:%d/api/v1", port)

	return ws.server.ListenAndServe()
}

// Stop gracefully stops the web server
func (ws *WebServer) Stop() error {
	if ws.server != nil {
		log.Println("Stopping GoChain Explorer web interface...")
		return ws.server.Close()
	}
	return nil
}

// GetHandler returns the web handler for integration with other servers
func (ws *WebServer) GetHandler() *WebHandler {
	return ws.handler
}

// GetTemplates returns the templates manager
func (ws *WebServer) GetTemplates() *Templates {
	return ws.templates
}

// HealthCheck performs a health check on the web server
func (ws *WebServer) HealthCheck() error {
	// Check if templates are loaded
	if ws.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	// Check if handler is initialized
	if ws.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	// Check if explorer service is available
	if ws.explorerService == nil {
		return fmt.Errorf("explorer service not available")
	}

	return nil
}

// GetServerInfo returns information about the web server
func (ws *WebServer) GetServerInfo() map[string]interface{} {
	info := map[string]interface{}{
		"type":       "web",
		"status":     "running",
		"templates":  len(ws.templates.templates),
		"started_at": time.Now(),
	}

	if ws.server != nil {
		info["address"] = ws.server.Addr
	}

	return info
}
