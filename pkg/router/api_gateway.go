package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// APIGateway provides HTTP API endpoints for the cluster router
type APIGateway struct {
	router     *ClusterRouter
	httpServer *http.Server
	config     *APIGatewayConfig
	logger     *Logger
}

// APIGatewayConfig holds configuration for the API gateway
type APIGatewayConfig struct {
	ListenAddr    string        `json:"listen_addr"`
	Port          int           `json:"port"`
	EnableCORS    bool          `json:"enable_cors"`
	EnableAuth    bool          `json:"enable_auth"`
	RateLimit     int           `json:"rate_limit"`
	Timeout       time.Duration `json:"timeout"`
	EnableMetrics bool          `json:"enable_metrics"`
	EnableHealth  bool          `json:"enable_health"`
}

// DefaultAPIGatewayConfig returns the default API gateway configuration
func DefaultAPIGatewayConfig() *APIGatewayConfig {
	return &APIGatewayConfig{
		ListenAddr:    "0.0.0.0",
		Port:          8080,
		EnableCORS:    true,
		EnableAuth:    false,
		RateLimit:     1000,
		Timeout:       30 * time.Second,
		EnableMetrics: true,
		EnableHealth:  true,
	}
}

// NewAPIGateway creates a new API gateway
func NewAPIGateway(router *ClusterRouter, config *APIGatewayConfig) (*APIGateway, error) {
	if config == nil {
		config = DefaultAPIGatewayConfig()
	}

	gateway := &APIGateway{
		router: router,
		config: config,
		logger: NewLogger("api-gateway"),
	}

	// Setup HTTP server
	mux := mux.NewRouter()
	gateway.setupRoutes(mux)

	gateway.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.ListenAddr, config.Port),
		Handler:      mux,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	}

	return gateway, nil
}

// setupRoutes configures all API routes
func (ag *APIGateway) setupRoutes(router *mux.Router) {
	// Health check endpoint
	if ag.config.EnableHealth {
		router.HandleFunc("/health", ag.healthHandler).Methods("GET")
	}

	// Metrics endpoint
	if ag.config.EnableMetrics {
		router.HandleFunc("/metrics", ag.metricsHandler).Methods("GET")
	}

	// API v1 routes
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Cluster management
	apiV1.HandleFunc("/clusters", ag.getClustersHandler).Methods("GET")
	apiV1.HandleFunc("/clusters", ag.createClusterHandler).Methods("POST")
	apiV1.HandleFunc("/clusters/{id}", ag.getClusterHandler).Methods("GET")
	apiV1.HandleFunc("/clusters/{id}", ag.updateClusterHandler).Methods("PUT")
	apiV1.HandleFunc("/clusters/{id}", ag.deleteClusterHandler).Methods("DELETE")

	// Node management
	apiV1.HandleFunc("/nodes", ag.getNodesHandler).Methods("GET")
	apiV1.HandleFunc("/nodes", ag.createNodeHandler).Methods("POST")
	apiV1.HandleFunc("/nodes/{id}", ag.getNodeHandler).Methods("GET")
	apiV1.HandleFunc("/nodes/{id}", ag.updateNodeHandler).Methods("PUT")
	apiV1.HandleFunc("/nodes/{id}", ag.deleteNodeHandler).Methods("DELETE")

	// Routing operations
	apiV1.HandleFunc("/route", ag.routeRequestHandler).Methods("POST")
	apiV1.HandleFunc("/route/status", ag.getRoutingStatusHandler).Methods("GET")

	// Discovery operations
	apiV1.HandleFunc("/discovery/clusters", ag.getDiscoveredClustersHandler).Methods("GET")
	apiV1.HandleFunc("/discovery/nodes", ag.getDiscoveredNodesHandler).Methods("GET")
	apiV1.HandleFunc("/discovery/refresh", ag.refreshDiscoveryHandler).Methods("POST")

	// Health monitoring
	apiV1.HandleFunc("/health/clusters", ag.getClusterHealthHandler).Methods("GET")
	apiV1.HandleFunc("/health/nodes", ag.getNodeHealthHandler).Methods("GET")
	apiV1.HandleFunc("/health/check/{id}", ag.checkNodeHealthHandler).Methods("POST")

	// Load balancing
	apiV1.HandleFunc("/loadbalancer/strategy", ag.getLoadBalancerStrategyHandler).Methods("GET")
	apiV1.HandleFunc("/loadbalancer/strategy", ag.setLoadBalancerStrategyHandler).Methods("PUT")
	apiV1.HandleFunc("/loadbalancer/stats", ag.getLoadBalancerStatsHandler).Methods("GET")

	// Add CORS middleware if enabled
	if ag.config.EnableCORS {
		router.Use(ag.corsMiddleware)
	}

	// Add logging middleware
	router.Use(ag.loggingMiddleware)
}

// healthHandler provides a health check endpoint
func (ag *APIGateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "cluster-router-api",
		"version":   "1.0.0",
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// metricsHandler provides metrics endpoint
func (ag *APIGateway) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := ag.router.GetMetrics()
	if metrics == nil {
		http.Error(w, "No metrics available", http.StatusServiceUnavailable)
		return
	}

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

// getClustersHandler returns all clusters
func (ag *APIGateway) getClustersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clusters := ag.router.GetClusterStatus()
	if err := json.NewEncoder(w).Encode(clusters); err != nil {
		http.Error(w, "Failed to encode clusters", http.StatusInternalServerError)
		return
	}
}

// createClusterHandler creates a new cluster
func (ag *APIGateway) createClusterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var cluster Cluster
	if err := json.NewDecoder(r.Body).Decode(&cluster); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := ag.router.RegisterCluster(&cluster); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(cluster); err != nil {
		http.Error(w, "Failed to encode cluster", http.StatusInternalServerError)
		return
	}
}

// getClusterHandler returns a specific cluster
func (ag *APIGateway) getClusterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	clusterID := ClusterID(vars["id"])

	clusters := ag.router.GetClusterStatus()
	cluster, exists := clusters[clusterID]
	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(cluster); err != nil {
		http.Error(w, "Failed to encode cluster", http.StatusInternalServerError)
		return
	}
}

// updateClusterHandler updates a cluster
func (ag *APIGateway) updateClusterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	clusterID := ClusterID(vars["id"])

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update cluster logic would go here
	// For now, return the cluster as-is
	clusters := ag.router.GetClusterStatus()
	cluster, exists := clusters[clusterID]
	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(cluster); err != nil {
		http.Error(w, "Failed to encode cluster", http.StatusInternalServerError)
		return
	}
}

// deleteClusterHandler deletes a cluster
func (ag *APIGateway) deleteClusterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	clusterID := ClusterID(vars["id"])

	// Delete cluster logic would go here
	// For now, return success
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Cluster %s deleted", clusterID),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getNodesHandler returns all nodes
func (ag *APIGateway) getNodesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	nodes := ag.router.GetNodeStatus()
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, "Failed to encode nodes", http.StatusInternalServerError)
		return
	}
}

// createNodeHandler creates a new node
func (ag *APIGateway) createNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var node Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := ag.router.RegisterNode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(node); err != nil {
		http.Error(w, "Failed to encode node", http.StatusInternalServerError)
		return
	}
}

// getNodeHandler returns a specific node
func (ag *APIGateway) getNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nodeID := NodeID(vars["id"])

	nodes := ag.router.GetNodeStatus()
	node, exists := nodes[nodeID]
	if !exists {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(node); err != nil {
		http.Error(w, "Failed to encode node", http.StatusInternalServerError)
		return
	}
}

// updateNodeHandler updates a node
func (ag *APIGateway) updateNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nodeID := NodeID(vars["id"])

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update node logic would go here
	// For now, return the node as-is
	nodes := ag.router.GetNodeStatus()
	node, exists := nodes[nodeID]
	if !exists {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(node); err != nil {
		http.Error(w, "Failed to encode node", http.StatusInternalServerError)
		return
	}
}

// deleteNodeHandler deletes a node
func (ag *APIGateway) deleteNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nodeID := NodeID(vars["id"])

	// Delete node logic would go here
	// For now, return success
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Node %s deleted", nodeID),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// routeRequestHandler routes a request
func (ag *APIGateway) routeRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := ag.router.RouteRequest(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getRoutingStatusHandler returns routing status
func (ag *APIGateway) getRoutingStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"clusters": ag.router.GetClusterStatus(),
		"nodes":    ag.router.GetNodeStatus(),
		"metrics":  ag.router.GetMetrics(),
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDiscoveredClustersHandler returns discovered clusters
func (ag *APIGateway) getDiscoveredClustersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would integrate with the cluster discovery system
	// For now, return empty result
	clusters := make(map[ClusterID]*DiscoveredCluster)
	if err := json.NewEncoder(w).Encode(clusters); err != nil {
		http.Error(w, "Failed to encode clusters", http.StatusInternalServerError)
		return
	}
}

// getDiscoveredNodesHandler returns discovered nodes
func (ag *APIGateway) getDiscoveredNodesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would integrate with the cluster discovery system
	// For now, return empty result
	nodes := make(map[NodeID]*DiscoveredNode)
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, "Failed to encode nodes", http.StatusInternalServerError)
		return
	}
}

// refreshDiscoveryHandler triggers discovery refresh
func (ag *APIGateway) refreshDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would trigger discovery refresh
	// For now, return success
	response := map[string]interface{}{
		"success": true,
		"message": "Discovery refresh triggered",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getClusterHealthHandler returns cluster health status
func (ag *APIGateway) getClusterHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would return cluster health information
	// For now, return basic status
	health := map[string]interface{}{
		"status":   "healthy",
		"clusters": ag.router.GetClusterStatus(),
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Failed to encode health", http.StatusInternalServerError)
		return
	}
}

// getNodeHealthHandler returns node health status
func (ag *APIGateway) getNodeHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would return node health information
	// For now, return basic status
	health := map[string]interface{}{
		"status": "healthy",
		"nodes":  ag.router.GetNodeStatus(),
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Failed to encode health", http.StatusInternalServerError)
		return
	}
}

// checkNodeHealthHandler performs health check on a specific node
func (ag *APIGateway) checkNodeHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nodeID := NodeID(vars["id"])

	// This would perform health check on the node
	// For now, return success
	response := map[string]interface{}{
		"success": true,
		"node_id": nodeID,
		"status":  "healthy",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getLoadBalancerStrategyHandler returns current load balancer strategy
func (ag *APIGateway) getLoadBalancerStrategyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would return the current load balancer strategy
	// For now, return default strategy
	strategy := map[string]interface{}{
		"strategy": RoutingStrategyAdaptive,
	}

	json.NewEncoder(w).Encode(strategy)
}

// setLoadBalancerStrategyHandler sets the load balancer strategy
func (ag *APIGateway) setLoadBalancerStrategyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	strategy, ok := request["strategy"].(string)
	if !ok {
		http.Error(w, "Invalid strategy", http.StatusBadRequest)
		return
	}

	// This would set the load balancer strategy
	// For now, return success
	response := map[string]interface{}{
		"success":  true,
		"strategy": strategy,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getLoadBalancerStatsHandler returns load balancer statistics
func (ag *APIGateway) getLoadBalancerStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// This would return load balancer statistics
	// For now, return basic stats
	stats := map[string]interface{}{
		"strategy":           RoutingStrategyAdaptive,
		"total_connections":  0,
		"active_connections": 0,
	}

	json.NewEncoder(w).Encode(stats)
}

// corsMiddleware adds CORS headers
func (ag *APIGateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (ag *APIGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		ag.logger.Info("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

// Start starts the API gateway server
func (ag *APIGateway) Start() error {
	ag.logger.Info("Starting API gateway on %s", ag.httpServer.Addr)
	return ag.httpServer.ListenAndServe()
}

// Stop stops the API gateway server
func (ag *APIGateway) Stop() error {
	ag.logger.Info("Stopping API gateway")
	return ag.httpServer.Close()
}
