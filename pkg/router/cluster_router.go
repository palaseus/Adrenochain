package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ClusterRouter implements an intelligent cluster-based routing protocol
type ClusterRouter struct {
	mu               sync.RWMutex
	clusters         map[ClusterID]*Cluster
	nodes            map[NodeID]*Node
	routingTable     *RoutingTable
	loadBalancer     *LoadBalancer
	healthMonitor    *HealthMonitor
	metricsCollector *MetricsCollector
	config           *ClusterRouterConfig
	logger           *Logger
	ctx              context.Context
	cancel           context.CancelFunc

	// Performance tracking
	requestCount    int64
	successCount    int64
	errorCount      int64
	avgResponseTime time.Duration
	lastHealthCheck time.Time
}

// ClusterID represents a unique identifier for a cluster
type ClusterID string

// NodeID represents a unique identifier for a node
type NodeID string

// Cluster represents a group of nodes working together
type Cluster struct {
	ID            ClusterID              `json:"id"`
	Name          string                 `json:"name"`
	Type          ClusterType            `json:"type"`
	Region        string                 `json:"region"`
	Nodes         map[NodeID]*Node       `json:"nodes"`
	Leader        NodeID                 `json:"leader"`
	Status        ClusterStatus          `json:"status"`
	Load          float64                `json:"load"`
	Capacity      int64                  `json:"capacity"`
	Latency       time.Duration          `json:"latency"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Node represents a single node in the cluster
type Node struct {
	ID            NodeID                 `json:"id"`
	Address       string                 `json:"address"`
	Port          int                    `json:"port"`
	ClusterID     ClusterID              `json:"cluster_id"`
	Status        NodeStatus             `json:"status"`
	Load          float64                `json:"load"`
	Capacity      int64                  `json:"capacity"`
	Latency       time.Duration          `json:"latency"`
	HealthScore   float64                `json:"health_score"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	RequestCount  int64                  `json:"request_count"`
	ErrorCount    int64                  `json:"error_count"`
	ResponseTime  time.Duration          `json:"response_time"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ClusterType represents the type of cluster
type ClusterType string

const (
	ClusterTypeAPI       ClusterType = "api"
	ClusterTypeConsensus ClusterType = "consensus"
	ClusterTypeStorage   ClusterType = "storage"
	ClusterTypeGateway   ClusterType = "gateway"
	ClusterTypeValidator ClusterType = "validator"
	ClusterTypeMixed     ClusterType = "mixed"
)

// ClusterStatus represents the status of a cluster
type ClusterStatus string

const (
	ClusterStatusActive      ClusterStatus = "active"
	ClusterStatusInactive    ClusterStatus = "inactive"
	ClusterStatusDegraded    ClusterStatus = "degraded"
	ClusterStatusMaintenance ClusterStatus = "maintenance"
	ClusterStatusFailed      ClusterStatus = "failed"
)

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusActive      NodeStatus = "active"
	NodeStatusInactive    NodeStatus = "inactive"
	NodeStatusDegraded    NodeStatus = "degraded"
	NodeStatusMaintenance NodeStatus = "maintenance"
	NodeStatusFailed      NodeStatus = "failed"
)

// RoutingStrategy represents different routing strategies
type RoutingStrategy string

const (
	RoutingStrategyRoundRobin   RoutingStrategy = "round_robin"
	RoutingStrategyLeastConn    RoutingStrategy = "least_connections"
	RoutingStrategyLeastLatency RoutingStrategy = "least_latency"
	RoutingStrategyLeastLoad    RoutingStrategy = "least_load"
	RoutingStrategyGeographic   RoutingStrategy = "geographic"
	RoutingStrategyWeighted     RoutingStrategy = "weighted"
	RoutingStrategyAdaptive     RoutingStrategy = "adaptive"
)

// ClusterRouterConfig holds configuration for the cluster router
type ClusterRouterConfig struct {
	MaxClusters         int             `json:"max_clusters"`
	MaxNodesPerCluster  int             `json:"max_nodes_per_cluster"`
	HealthCheckInterval time.Duration   `json:"health_check_interval"`
	LoadUpdateInterval  time.Duration   `json:"load_update_interval"`
	RoutingStrategy     RoutingStrategy `json:"routing_strategy"`
	EnableFailover      bool            `json:"enable_failover"`
	EnableLoadBalancing bool            `json:"enable_load_balancing"`
	EnableMetrics       bool            `json:"enable_metrics"`
	MaxRetries          int             `json:"max_retries"`
	Timeout             time.Duration   `json:"timeout"`
	TestMode            bool            `json:"test_mode"` // Disable background processes for testing
}

// DefaultClusterRouterConfig returns the default configuration
func DefaultClusterRouterConfig() *ClusterRouterConfig {
	return &ClusterRouterConfig{
		MaxClusters:         10,
		MaxNodesPerCluster:  50,
		HealthCheckInterval: 30 * time.Second,
		LoadUpdateInterval:  10 * time.Second,
		RoutingStrategy:     RoutingStrategyAdaptive,
		EnableFailover:      true,
		EnableLoadBalancing: true,
		EnableMetrics:       true,
		MaxRetries:          3,
		Timeout:             30 * time.Second,
	}
}

// NewClusterRouter creates a new cluster router
func NewClusterRouter(config *ClusterRouterConfig) (*ClusterRouter, error) {
	if config == nil {
		config = DefaultClusterRouterConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logger
	var routerLogger *Logger
	if config.TestMode {
		routerLogger = NewTestLogger("cluster-router")
	} else {
		routerLogger = NewLogger("cluster-router")
	}

	// Initialize components
	routingTable := NewRoutingTable()
	loadBalancer := NewLoadBalancer(config.RoutingStrategy)
	healthMonitor := NewHealthMonitorWithTestMode(config.HealthCheckInterval, config.TestMode)
	metricsCollector := NewMetricsCollector()

	router := &ClusterRouter{
		clusters:         make(map[ClusterID]*Cluster),
		nodes:            make(map[NodeID]*Node),
		routingTable:     routingTable,
		loadBalancer:     loadBalancer,
		healthMonitor:    healthMonitor,
		metricsCollector: metricsCollector,
		config:           config,
		logger:           routerLogger,
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start background processes only if not in test mode
	if !config.TestMode {
		go router.startHealthMonitoring()
		go router.startLoadBalancing()
		go router.startMetricsCollection()
	}

	return router, nil
}

// RegisterCluster registers a new cluster
func (cr *ClusterRouter) RegisterCluster(cluster *Cluster) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if len(cr.clusters) >= cr.config.MaxClusters {
		return fmt.Errorf("maximum number of clusters reached: %d", cr.config.MaxClusters)
	}

	// Validate cluster
	if err := cr.validateCluster(cluster); err != nil {
		return fmt.Errorf("invalid cluster: %w", err)
	}

	// Set timestamps
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	// Register cluster
	cr.clusters[cluster.ID] = cluster

	// Register all nodes in the cluster
	for _, node := range cluster.Nodes {
		if err := cr.registerNode(node); err != nil {
			cr.logger.Error("Failed to register node %s: %v", node.ID, err)
		}
	}

	// Update routing table
	cr.routingTable.UpdateCluster(cluster)

	cr.logger.Info("Registered cluster %s with %d nodes", cluster.ID, len(cluster.Nodes))
	return nil
}

// RegisterNode registers a new node
func (cr *ClusterRouter) RegisterNode(node *Node) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	return cr.registerNode(node)
}

// registerNode is the internal method to register a node
func (cr *ClusterRouter) registerNode(node *Node) error {
	// Validate node
	if err := cr.validateNode(node); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}

	// Check if cluster exists
	cluster, exists := cr.clusters[node.ClusterID]
	if !exists {
		return fmt.Errorf("cluster %s does not exist", node.ClusterID)
	}

	// Check cluster capacity
	if len(cluster.Nodes) >= cr.config.MaxNodesPerCluster {
		return fmt.Errorf("cluster %s has reached maximum capacity", node.ClusterID)
	}

	// Set timestamps
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now

	// Register node
	cr.nodes[node.ID] = node
	cluster.Nodes[node.ID] = node

	// Update routing table
	cr.routingTable.UpdateNode(node)

	cr.logger.Info("Registered node %s in cluster %s", node.ID, node.ClusterID)
	return nil
}

// RouteRequest routes a request to the best available node
func (cr *ClusterRouter) RouteRequest(req *Request) (*Response, error) {
	start := time.Now()
	atomic.AddInt64(&cr.requestCount, 1)

	// Find best cluster for the request
	cluster, err := cr.selectBestCluster(req)
	if err != nil {
		atomic.AddInt64(&cr.errorCount, 1)
		return nil, fmt.Errorf("failed to select cluster: %w", err)
	}

	// Find best node in the cluster
	node, err := cr.selectBestNode(cluster, req)
	if err != nil {
		atomic.AddInt64(&cr.errorCount, 1)
		return nil, fmt.Errorf("failed to select node: %w", err)
	}

	// Execute request with retries
	var response *Response
	var lastErr error

	for attempt := 0; attempt <= cr.config.MaxRetries; attempt++ {
		response, lastErr = cr.executeRequest(node, req)
		if lastErr == nil {
			break
		}

		// If this is not the last attempt, try to find an alternative node
		if attempt < cr.config.MaxRetries {
			cr.logger.Warn("Request failed on node %s (attempt %d): %v", node.ID, attempt+1, lastErr)

			// Mark node as degraded temporarily
			cr.markNodeDegraded(node.ID)

			// Try to find alternative node
			altNode, altErr := cr.selectBestNode(cluster, req)
			if altErr == nil && altNode.ID != node.ID {
				node = altNode
				cr.logger.Info("Switched to alternative node %s", node.ID)
			}
		}
	}

	// Update metrics
	duration := time.Since(start)
	cr.updateResponseTime(duration)

	if lastErr != nil {
		atomic.AddInt64(&cr.errorCount, 1)
		atomic.AddInt64(&node.ErrorCount, 1)
		return nil, lastErr
	}

	atomic.AddInt64(&cr.successCount, 1)
	atomic.AddInt64(&node.RequestCount, 1)
	node.ResponseTime = duration

	// Update node load
	cr.updateNodeLoad(node.ID, 1)

	return response, nil
}

// selectBestCluster selects the best cluster for a request
func (cr *ClusterRouter) selectBestCluster(req *Request) (*Cluster, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var candidates []*Cluster

	// Filter clusters by type and status
	for _, cluster := range cr.clusters {
		if cluster.Status == ClusterStatusActive &&
			(cluster.Type == req.ClusterType || cluster.Type == ClusterTypeMixed) {
			candidates = append(candidates, cluster)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available clusters for request type %s", req.ClusterType)
	}

	// Apply routing strategy
	switch cr.config.RoutingStrategy {
	case RoutingStrategyLeastLoad:
		return cr.selectClusterByLoad(candidates)
	case RoutingStrategyLeastLatency:
		return cr.selectClusterByLatency(candidates, req)
	case RoutingStrategyGeographic:
		return cr.selectClusterByGeography(candidates, req)
	case RoutingStrategyAdaptive:
		return cr.selectClusterAdaptive(candidates, req)
	default:
		return cr.selectClusterRoundRobin(candidates)
	}
}

// selectBestNode selects the best node in a cluster
func (cr *ClusterRouter) selectBestNode(cluster *Cluster, req *Request) (*Node, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var candidates []*Node

	// Filter nodes by status
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusActive {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available nodes in cluster %s", cluster.ID)
	}

	// Apply load balancing strategy
	return cr.loadBalancer.SelectNode(candidates, req)
}

// executeRequest executes a request on a specific node
func (cr *ClusterRouter) executeRequest(node *Node, req *Request) (*Response, error) {
	// Create connection to node
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Address, node.Port), cr.config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node %s: %w", node.ID, err)
	}
	defer conn.Close()

	// Set timeout
	if err := conn.SetDeadline(time.Now().Add(cr.config.Timeout)); err != nil {
		// Log error but continue with connection
		fmt.Printf("Failed to set deadline: %v\n", err)
	}

	// Send request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := conn.Write(reqData); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	response := &Response{}
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// validateCluster validates a cluster configuration
func (cr *ClusterRouter) validateCluster(cluster *Cluster) error {
	if cluster.ID == "" {
		return fmt.Errorf("cluster ID cannot be empty")
	}
	if cluster.Name == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}
	if cluster.Type == "" {
		return fmt.Errorf("cluster type cannot be empty")
	}
	if len(cluster.Nodes) == 0 {
		return fmt.Errorf("cluster must have at least one node")
	}
	return nil
}

// validateNode validates a node configuration
func (cr *ClusterRouter) validateNode(node *Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if node.Address == "" {
		return fmt.Errorf("node address cannot be empty")
	}
	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("invalid node port: %d", node.Port)
	}
	if node.ClusterID == "" {
		return fmt.Errorf("node cluster ID cannot be empty")
	}
	return nil
}

// markNodeDegraded marks a node as degraded
func (cr *ClusterRouter) markNodeDegraded(nodeID NodeID) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if node, exists := cr.nodes[nodeID]; exists {
		node.Status = NodeStatusDegraded
		node.UpdatedAt = time.Now()
		cr.logger.Warn("Marked node %s as degraded", nodeID)
	}
}

// updateResponseTime updates the average response time
func (cr *ClusterRouter) updateResponseTime(duration time.Duration) {
	// Simple moving average
	cr.avgResponseTime = (cr.avgResponseTime + duration) / 2
}

// updateNodeLoad updates the load of a node
func (cr *ClusterRouter) updateNodeLoad(nodeID NodeID, loadDelta float64) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if node, exists := cr.nodes[nodeID]; exists {
		node.Load += loadDelta
		node.UpdatedAt = time.Now()
	}
}

// startHealthMonitoring starts the health monitoring process
func (cr *ClusterRouter) startHealthMonitoring() {
	ticker := time.NewTicker(cr.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.performHealthCheck()
		}
	}
}

// startLoadBalancing starts the load balancing process
func (cr *ClusterRouter) startLoadBalancing() {
	ticker := time.NewTicker(cr.config.LoadUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.updateLoadBalancing()
		}
	}
}

// startMetricsCollection starts the metrics collection process
func (cr *ClusterRouter) startMetricsCollection() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-ticker.C:
			cr.collectMetrics()
		}
	}
}

// performHealthCheck performs health checks on all nodes
func (cr *ClusterRouter) performHealthCheck() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for _, node := range cr.nodes {
		go cr.checkNodeHealth(node)
	}

	cr.lastHealthCheck = time.Now()
}

// checkNodeHealth checks the health of a specific node
func (cr *ClusterRouter) checkNodeHealth(node *Node) {
	// Simple TCP connection check
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Address, node.Port), 5*time.Second)
	if err != nil {
		cr.logger.Warn("Health check failed for node %s: %v", node.ID, err)
		cr.markNodeUnhealthy(node.ID)
		return
	}
	conn.Close()

	// Update health score
	cr.mu.Lock()
	node.HealthScore = 1.0
	node.LastHeartbeat = time.Now()
	node.UpdatedAt = time.Now()
	cr.mu.Unlock()
}

// markNodeUnhealthy marks a node as unhealthy
func (cr *ClusterRouter) markNodeUnhealthy(nodeID NodeID) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if node, exists := cr.nodes[nodeID]; exists {
		node.Status = NodeStatusFailed
		node.HealthScore = 0.0
		node.UpdatedAt = time.Now()
		cr.logger.Error("Marked node %s as unhealthy", nodeID)
	}
}

// updateLoadBalancing updates load balancing information
func (cr *ClusterRouter) updateLoadBalancing() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Update cluster loads
	for _, cluster := range cr.clusters {
		totalLoad := 0.0
		activeNodes := 0

		for _, node := range cluster.Nodes {
			if node.Status == NodeStatusActive {
				totalLoad += node.Load
				activeNodes++
			}
		}

		if activeNodes > 0 {
			cluster.Load = totalLoad / float64(activeNodes)
		} else {
			cluster.Load = 0.0
		}

		cluster.UpdatedAt = time.Now()
	}
}

// collectMetrics collects performance metrics
func (cr *ClusterRouter) collectMetrics() {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	metrics := &ClusterMetrics{
		Timestamp:          time.Now(),
		TotalRequests:      atomic.LoadInt64(&cr.requestCount),
		SuccessfulRequests: atomic.LoadInt64(&cr.successCount),
		FailedRequests:     atomic.LoadInt64(&cr.errorCount),
		AvgResponseTime:    cr.avgResponseTime,
		ClusterCount:       len(cr.clusters),
		NodeCount:          len(cr.nodes),
		ActiveClusters:     0,
		ActiveNodes:        0,
	}

	// Count active clusters and nodes
	for _, cluster := range cr.clusters {
		if cluster.Status == ClusterStatusActive {
			metrics.ActiveClusters++
		}
	}

	for _, node := range cr.nodes {
		if node.Status == NodeStatusActive {
			metrics.ActiveNodes++
		}
	}

	// Store metrics
	cr.metricsCollector.RecordMetrics(metrics)
}

// GetClusterStatus returns the status of all clusters
func (cr *ClusterRouter) GetClusterStatus() map[ClusterID]*Cluster {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	status := make(map[ClusterID]*Cluster)
	for id, cluster := range cr.clusters {
		status[id] = cluster
	}
	return status
}

// GetNodeStatus returns the status of all nodes
func (cr *ClusterRouter) GetNodeStatus() map[NodeID]*Node {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	status := make(map[NodeID]*Node)
	for id, node := range cr.nodes {
		status[id] = node
	}
	return status
}

// GetMetrics returns current performance metrics
func (cr *ClusterRouter) GetMetrics() *ClusterMetrics {
	return cr.metricsCollector.GetLatestMetrics()
}

// Close shuts down the cluster router
func (cr *ClusterRouter) Close() error {
	cr.cancel()

	// Close health monitor
	if cr.healthMonitor != nil {
		cr.healthMonitor.Close()
	}

	cr.logger.Info("Cluster router shut down")
	return nil
}

// Cluster selection algorithms

func (cr *ClusterRouter) selectClusterByLoad(candidates []*Cluster) (*Cluster, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Load < candidates[j].Load
	})

	return candidates[0], nil
}

func (cr *ClusterRouter) selectClusterByLatency(candidates []*Cluster, req *Request) (*Cluster, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Latency < candidates[j].Latency
	})

	return candidates[0], nil
}

func (cr *ClusterRouter) selectClusterByGeography(candidates []*Cluster, req *Request) (*Cluster, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}

	// Simple geographic selection based on region
	if req.Region != "" {
		for _, cluster := range candidates {
			if cluster.Region == req.Region {
				return cluster, nil
			}
		}
	}

	// Fallback to first available cluster
	return candidates[0], nil
}

func (cr *ClusterRouter) selectClusterAdaptive(candidates []*Cluster, req *Request) (*Cluster, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}

	// Adaptive selection based on multiple factors
	bestCluster := candidates[0]
	bestScore := 0.0

	for _, cluster := range candidates {
		score := cr.calculateClusterScore(cluster, req)
		if score > bestScore {
			bestScore = score
			bestCluster = cluster
		}
	}

	return bestCluster, nil
}

func (cr *ClusterRouter) selectClusterRoundRobin(candidates []*Cluster) (*Cluster, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates available")
	}

	// Simple round-robin selection
	index := int(time.Now().UnixNano()) % len(candidates)
	return candidates[index], nil
}

// calculateClusterScore calculates a score for cluster selection
func (cr *ClusterRouter) calculateClusterScore(cluster *Cluster, req *Request) float64 {
	score := 1.0

	// Factor in load (lower is better)
	score *= (1.0 - cluster.Load)

	// Factor in latency (lower is better)
	if cluster.Latency > 0 {
		score *= (1.0 / (1.0 + float64(cluster.Latency.Milliseconds())/1000.0))
	}

	// Factor in health (higher is better)
	activeNodes := 0
	totalHealth := 0.0
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusActive {
			activeNodes++
			totalHealth += node.HealthScore
		}
	}

	if activeNodes > 0 {
		avgHealth := totalHealth / float64(activeNodes)
		score *= avgHealth
	}

	// Factor in capacity utilization
	if cluster.Capacity > 0 {
		utilization := float64(activeNodes) / float64(cluster.Capacity)
		score *= (1.0 - utilization*0.5) // Penalize high utilization
	}

	return score
}
