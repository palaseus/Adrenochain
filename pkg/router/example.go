package router

import (
	"fmt"
	"log"
	"time"
)

// ExampleClusterRouter demonstrates how to use the cluster router
func ExampleClusterRouter() {
	// Create cluster router configuration
	config := &ClusterRouterConfig{
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

	// Create cluster router
	router, err := NewClusterRouter(config)
	if err != nil {
		log.Fatalf("Failed to create cluster router: %v", err)
	}
	defer router.Close()

	// Create API cluster
	apiCluster := &Cluster{
		ID:     "api-cluster-1",
		Name:   "API Cluster 1",
		Type:   ClusterTypeAPI,
		Region: "us-east-1",
		Nodes:  make(map[NodeID]*Node),
		Status: ClusterStatusActive,
	}

	// Register cluster
	err = router.RegisterCluster(apiCluster)
	if err != nil {
		log.Fatalf("Failed to register cluster: %v", err)
	}

	// Create and register nodes
	nodes := []*Node{
		{
			ID:          "api-node-1",
			Address:     "192.168.1.10",
			Port:        8080,
			ClusterID:   "api-cluster-1",
			Status:      NodeStatusActive,
			HealthScore: 1.0,
			Load:        0.2,
			Metadata: map[string]interface{}{
				"region": "us-east-1",
				"zone":   "us-east-1a",
			},
		},
		{
			ID:          "api-node-2",
			Address:     "192.168.1.11",
			Port:        8080,
			ClusterID:   "api-cluster-1",
			Status:      NodeStatusActive,
			HealthScore: 0.9,
			Load:        0.4,
			Metadata: map[string]interface{}{
				"region": "us-east-1",
				"zone":   "us-east-1b",
			},
		},
		{
			ID:          "api-node-3",
			Address:     "192.168.1.12",
			Port:        8080,
			ClusterID:   "api-cluster-1",
			Status:      NodeStatusActive,
			HealthScore: 0.8,
			Load:        0.6,
			Metadata: map[string]interface{}{
				"region": "us-east-1",
				"zone":   "us-east-1c",
			},
		},
	}

	for _, node := range nodes {
		err = router.RegisterNode(node)
		if err != nil {
			log.Fatalf("Failed to register node %s: %v", node.ID, err)
		}
	}

	// Create a request
	req := &Request{
		ID:          "request-1",
		Type:        "api",
		ClusterType: ClusterTypeAPI,
		Region:      "us-east-1",
		Priority:    1,
		Timeout:     30 * time.Second,
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"user_id": "user-123",
			"service": "web-api",
		},
	}

	// Route the request
	response, err := router.RouteRequest(req)
	if err != nil {
		log.Printf("Failed to route request: %v", err)
	} else {
		fmt.Printf("Request routed successfully to node %s in cluster %s\n",
			response.NodeID, response.ClusterID)
	}

	// Get cluster status
	clusters := router.GetClusterStatus()
	fmt.Printf("Registered clusters: %d\n", len(clusters))

	// Get node status
	nodes_status := router.GetNodeStatus()
	fmt.Printf("Registered nodes: %d\n", len(nodes_status))

	// Get metrics
	metrics := router.GetMetrics()
	if metrics != nil {
		fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
		fmt.Printf("Success rate: %.2f%%\n",
			float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	}
}

// ExampleClusterDiscovery demonstrates cluster discovery
func ExampleClusterDiscovery() {
	// Create discovery configuration
	config := &DiscoveryConfig{
		EnableMDNS:        true,
		EnableDNS:         true,
		EnableBootstrap:   true,
		EnableBroadcast:   false,
		DiscoveryInterval: 30 * time.Second,
		BootstrapPeers:    []string{"192.168.1.100:8080", "192.168.1.101:8080"},
		DNSSeeds:          []string{"seed1.adrenochain.com", "seed2.adrenochain.com"},
		ServiceName:       "adrenochain-cluster",
		Timeout:           10 * time.Second,
	}

	// Create cluster discovery
	discovery, err := NewClusterDiscovery(config)
	if err != nil {
		log.Fatalf("Failed to create cluster discovery: %v", err)
	}
	defer discovery.Close()

	// Wait for discovery to find clusters
	time.Sleep(5 * time.Second)

	// Get discovered clusters
	clusters := discovery.GetDiscoveredClusters()
	fmt.Printf("Discovered clusters: %d\n", len(clusters))

	// Get discovered nodes
	nodes := discovery.GetDiscoveredNodes()
	fmt.Printf("Discovered nodes: %d\n", len(nodes))

	// Get clusters by type
	apiClusters := discovery.GetClustersByType(ClusterTypeAPI)
	fmt.Printf("API clusters discovered: %d\n", len(apiClusters))
}

// ExampleClusterManager demonstrates cluster management
func ExampleClusterManager() {
	// Create cluster manager configuration
	config := &ClusterManagerConfig{
		EnableFailover:      true,
		EnableAutoRecovery:  true,
		HealthCheckInterval: 30 * time.Second,
		FailoverTimeout:     60 * time.Second,
		RecoveryTimeout:     300 * time.Second,
		MaxFailovers:        3,
		EventBufferSize:     1000,
	}

	// Create cluster manager
	manager, err := NewClusterManager(config)
	if err != nil {
		log.Fatalf("Failed to create cluster manager: %v", err)
	}
	defer manager.Close()

	// Create test cluster
	cluster := &Cluster{
		ID:     "managed-cluster-1",
		Name:   "Managed Cluster 1",
		Type:   ClusterTypeAPI,
		Region: "us-west-2",
		Nodes:  make(map[NodeID]*Node),
		Status: ClusterStatusActive,
	}

	// Register cluster for management
	err = manager.RegisterCluster(cluster)
	if err != nil {
		log.Fatalf("Failed to register cluster: %v", err)
	}

	// Create failover policy
	policy := &FailoverPolicy{
		ClusterID:           "managed-cluster-1",
		Enabled:             true,
		MaxFailovers:        3,
		FailoverTimeout:     60 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		RecoveryTimeout:     300 * time.Second,
		AutoRecovery:        true,
		BackupClusters:      []ClusterID{"backup-cluster-1", "backup-cluster-2"},
		Metadata: map[string]interface{}{
			"priority": "high",
			"owner":    "team-alpha",
		},
	}

	// Set failover policy
	err = manager.SetFailoverPolicy(policy)
	if err != nil {
		log.Fatalf("Failed to set failover policy: %v", err)
	}

	// Add event handler
	manager.AddEventHandler(&ExampleEventHandler{})

	// Get cluster status
	clusters := manager.GetClusterStatus()
	fmt.Printf("Managed clusters: %d\n", len(clusters))

	// Get failover policy
	retrievedPolicy, exists := manager.GetFailoverPolicy("managed-cluster-1")
	if exists {
		fmt.Printf("Failover policy enabled: %t\n", retrievedPolicy.Enabled)
		fmt.Printf("Max failovers: %d\n", retrievedPolicy.MaxFailovers)
	}
}

// ExampleEventHandler implements ClusterEventHandler for demonstration
type ExampleEventHandler struct{}

func (h *ExampleEventHandler) OnClusterEvent(event *ClusterEvent) error {
	fmt.Printf("Cluster event: %s - %s\n", event.Type, event.ClusterID)
	return nil
}

// ExampleAPIGateway demonstrates the API gateway
func ExampleAPIGateway() {
	// Create cluster router
	router, err := NewClusterRouter(nil)
	if err != nil {
		log.Fatalf("Failed to create cluster router: %v", err)
	}
	defer router.Close()

	// Create API gateway configuration
	gatewayConfig := &APIGatewayConfig{
		ListenAddr:    "0.0.0.0",
		Port:          8080,
		EnableCORS:    true,
		EnableAuth:    false,
		RateLimit:     1000,
		Timeout:       30 * time.Second,
		EnableMetrics: true,
		EnableHealth:  true,
	}

	// Create API gateway
	gateway, err := NewAPIGateway(router, gatewayConfig)
	if err != nil {
		log.Fatalf("Failed to create API gateway: %v", err)
	}

	// Start API gateway in a goroutine
	go func() {
		err := gateway.Start()
		if err != nil {
			log.Printf("API gateway error: %v", err)
		}
	}()

	fmt.Println("API gateway started on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health - Health check")
	fmt.Println("  GET  /metrics - Metrics")
	fmt.Println("  GET  /api/v1/clusters - List clusters")
	fmt.Println("  POST /api/v1/clusters - Create cluster")
	fmt.Println("  GET  /api/v1/nodes - List nodes")
	fmt.Println("  POST /api/v1/nodes - Create node")
	fmt.Println("  POST /api/v1/route - Route request")
	fmt.Println("  GET  /api/v1/route/status - Routing status")

	// Keep the example running
	time.Sleep(10 * time.Second)
}

// ExampleLoadBalancing demonstrates different load balancing strategies
func ExampleLoadBalancing() {
	// Create load balancer with different strategies
	strategies := []RoutingStrategy{
		RoutingStrategyRoundRobin,
		RoutingStrategyLeastConn,
		RoutingStrategyLeastLatency,
		RoutingStrategyLeastLoad,
		RoutingStrategyWeighted,
		RoutingStrategyAdaptive,
	}

	for _, strategy := range strategies {
		fmt.Printf("\nTesting %s strategy:\n", strategy)

		// Create load balancer
		lb := NewLoadBalancer(strategy)

		// Create test nodes
		nodes := []*Node{
			{
				ID:          "node-1",
				Address:     "192.168.1.10",
				Port:        8080,
				Status:      NodeStatusActive,
				Load:        0.2,
				HealthScore: 1.0,
				Latency:     10 * time.Millisecond,
			},
			{
				ID:          "node-2",
				Address:     "192.168.1.11",
				Port:        8080,
				Status:      NodeStatusActive,
				Load:        0.5,
				HealthScore: 0.9,
				Latency:     20 * time.Millisecond,
			},
			{
				ID:          "node-3",
				Address:     "192.168.1.12",
				Port:        8080,
				Status:      NodeStatusActive,
				Load:        0.8,
				HealthScore: 0.8,
				Latency:     30 * time.Millisecond,
			},
		}

		// Create test request
		req := &Request{
			ID:          "test-request",
			Type:        "api",
			ClusterType: ClusterTypeAPI,
			CreatedAt:   time.Now(),
		}

		// Test node selection
		for i := 0; i < 5; i++ {
			selectedNode, err := lb.SelectNode(nodes, req)
			if err != nil {
				fmt.Printf("  Error selecting node: %v\n", err)
				continue
			}
			fmt.Printf("  Selected node: %s (load: %.2f, latency: %v)\n",
				selectedNode.ID, selectedNode.Load, selectedNode.Latency)
		}
	}
}

// ExampleHealthMonitoring demonstrates health monitoring
func ExampleHealthMonitoring() {
	// Create health monitor
	monitor := NewHealthMonitor(30 * time.Second)

	// Create test nodes
	nodes := []*Node{
		{
			ID:      "healthy-node",
			Address: "127.0.0.1",
			Port:    8080,
			Status:  NodeStatusActive,
		},
		{
			ID:      "unhealthy-node",
			Address: "192.168.1.999", // Invalid address
			Port:    8080,
			Status:  NodeStatusActive,
		},
	}

	// Register nodes for health monitoring
	for _, node := range nodes {
		err := monitor.RegisterNode(node, HealthCheckTypeTCP)
		if err != nil {
			log.Printf("Failed to register node %s: %v", node.ID, err)
		}
	}

	// Wait for health checks
	time.Sleep(5 * time.Second)

	// Get health status
	healthyNodes := monitor.GetHealthyNodes()
	unhealthyNodes := monitor.GetUnhealthyNodes()

	fmt.Printf("Healthy nodes: %d\n", len(healthyNodes))
	fmt.Printf("Unhealthy nodes: %d\n", len(unhealthyNodes))

	// Get health history for a node
	history := monitor.GetHealthHistory("healthy-node", 10)
	fmt.Printf("Health history for healthy-node: %d records\n", len(history))

	// Get monitor statistics
	stats := monitor.GetStats()
	fmt.Printf("Total nodes monitored: %d\n", stats.TotalNodes)
	fmt.Printf("Enabled nodes: %d\n", stats.EnabledNodes)

	// Clean up
	monitor.Close()
}

// ExampleMetricsCollection demonstrates metrics collection
func ExampleMetricsCollection() {
	// Create metrics collector
	collector := NewMetricsCollector()

	// Simulate some metrics
	for i := 0; i < 10; i++ {
		metrics := &ClusterMetrics{
			Timestamp:          time.Now(),
			TotalRequests:      int64(100 + i*10),
			SuccessfulRequests: int64(95 + i*9),
			FailedRequests:     int64(5 + i),
			AvgResponseTime:    time.Duration(50+i*5) * time.Millisecond,
			ClusterCount:       3,
			NodeCount:          15,
			ActiveClusters:     3,
			ActiveNodes:        14,
			UnhealthyNodes:     1,
		}

		collector.RecordMetrics(metrics)
		time.Sleep(100 * time.Millisecond)
	}

	// Get latest metrics
	latest := collector.GetLatestMetrics()
	if latest != nil {
		fmt.Printf("Latest metrics:\n")
		fmt.Printf("  Total requests: %d\n", latest.TotalRequests)
		fmt.Printf("  Success rate: %.2f%%\n",
			float64(latest.SuccessfulRequests)/float64(latest.TotalRequests)*100)
		fmt.Printf("  Avg response time: %v\n", latest.AvgResponseTime)
	}

	// Get average metrics over last minute
	avgMetrics := collector.GetAverageMetrics(1 * time.Minute)
	if avgMetrics != nil {
		fmt.Printf("\nAverage metrics (last minute):\n")
		fmt.Printf("  Avg requests: %d\n", avgMetrics.TotalRequests)
		fmt.Printf("  Avg success rate: %.2f%%\n",
			float64(avgMetrics.SuccessfulRequests)/float64(avgMetrics.TotalRequests)*100)
	}

	// Get success rate
	successRate := collector.GetSuccessRate(1 * time.Minute)
	fmt.Printf("\nSuccess rate: %.2f%%\n", successRate*100)

	// Get throughput
	throughput := collector.GetThroughput(1 * time.Minute)
	fmt.Printf("Throughput: %.2f requests/second\n", throughput)

	// Get latency percentile
	p95Latency := collector.GetLatencyPercentile(1*time.Minute, 95)
	fmt.Printf("95th percentile latency: %v\n", p95Latency)

	// Get collector statistics
	stats := collector.GetStats()
	fmt.Printf("\nCollector stats:\n")
	fmt.Printf("  Total metrics: %d\n", stats.TotalMetrics)
	fmt.Printf("  Max history: %d\n", stats.MaxHistory)
}
