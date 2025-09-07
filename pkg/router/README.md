# Intelligent Cluster-Based Router Protocol

A sophisticated, production-ready cluster-based routing system for the Adrenochain blockchain platform. This implementation provides intelligent load balancing, health monitoring, failover mechanisms, and comprehensive metrics collection.

## 🚀 Features

### Core Functionality
- **Intelligent Routing**: Multiple routing strategies including adaptive, load-based, latency-based, and geographic routing
- **Cluster Management**: Full lifecycle management of clusters and nodes with automatic failover
- **Health Monitoring**: Comprehensive health checks with configurable intervals and recovery mechanisms
- **Load Balancing**: Advanced load balancing with multiple algorithms and real-time metrics
- **Service Discovery**: Multi-method cluster and peer discovery (mDNS, DNS, Bootstrap, Broadcast)
- **API Gateway**: RESTful API for cluster management and monitoring
- **Metrics Collection**: Detailed performance metrics and analytics

### Routing Strategies
- **Round Robin**: Simple round-robin distribution
- **Least Connections**: Route to node with fewest active connections
- **Least Latency**: Route to node with lowest latency
- **Least Load**: Route to node with lowest current load
- **Weighted**: Route based on configurable node weights
- **Adaptive**: Intelligent routing based on multiple factors (load, latency, health, capacity)

### Health Monitoring
- **TCP Health Checks**: Basic connectivity testing
- **HTTP Health Checks**: Application-level health verification
- **Custom Health Checks**: Pluggable health check implementations
- **Automatic Recovery**: Self-healing with configurable recovery policies
- **Health History**: Historical health data for trend analysis

### Cluster Discovery
- **mDNS Discovery**: Local network service discovery
- **DNS Discovery**: SRV record-based cluster discovery
- **Bootstrap Discovery**: Peer-to-peer cluster discovery
- **Broadcast Discovery**: Network broadcast-based discovery

## 📋 Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │ Cluster Router  │    │ Health Monitor  │
│                 │    │                 │    │                 │
│ • REST API      │◄──►│ • Load Balancer │◄──►│ • Health Checks │
│ • Authentication│    │ • Routing Table │    │ • Recovery      │
│ • Rate Limiting │    │ • Metrics       │    │ • History       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Cluster Manager │    │ Cluster Discovery│    │ Metrics Collector│
│                 │    │                 │    │                 │
│ • Lifecycle     │    │ • mDNS          │    │ • Performance   │
│ • Failover      │    │ • DNS           │    │ • Analytics     │
│ • Events        │    │ • Bootstrap     │    │ • Reporting     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Installation

```bash
# Add to your Go module
go get github.com/palaseus/adrenochain/pkg/router
```

## 📖 Usage

### Basic Setup

```go
package main

import (
    "log"
    "time"
    "github.com/palaseus/adrenochain/pkg/router"
)

func main() {
    // Create cluster router configuration
    config := &router.ClusterRouterConfig{
        MaxClusters:        10,
        MaxNodesPerCluster: 50,
        HealthCheckInterval: 30 * time.Second,
        LoadUpdateInterval:  10 * time.Second,
        RoutingStrategy:     router.RoutingStrategyAdaptive,
        EnableFailover:      true,
        EnableLoadBalancing: true,
        EnableMetrics:       true,
        MaxRetries:          3,
        Timeout:             30 * time.Second,
    }

    // Create cluster router
    router, err := router.NewClusterRouter(config)
    if err != nil {
        log.Fatalf("Failed to create cluster router: %v", err)
    }
    defer router.Close()

    // Register clusters and nodes...
}
```

### Registering Clusters and Nodes

```go
// Create a cluster
cluster := &router.Cluster{
    ID:     "api-cluster-1",
    Name:   "API Cluster 1",
    Type:   router.ClusterTypeAPI,
    Region: "us-east-1",
    Nodes:  make(map[router.NodeID]*router.Node),
    Status: router.ClusterStatusActive,
}

// Register the cluster
err := router.RegisterCluster(cluster)
if err != nil {
    log.Fatalf("Failed to register cluster: %v", err)
}

// Create and register nodes
node := &router.Node{
    ID:          "api-node-1",
    Address:     "192.168.1.10",
    Port:        8080,
    ClusterID:   "api-cluster-1",
    Status:      router.NodeStatusActive,
    HealthScore: 1.0,
    Load:        0.2,
    Metadata: map[string]interface{}{
        "region": "us-east-1",
        "zone":   "us-east-1a",
    },
}

err = router.RegisterNode(node)
if err != nil {
    log.Fatalf("Failed to register node: %v", err)
}
```

### Routing Requests

```go
// Create a request
req := &router.Request{
    ID:          "request-1",
    Type:        "api",
    ClusterType: router.ClusterTypeAPI,
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
    log.Printf("Request routed to node %s in cluster %s", 
        response.NodeID, response.ClusterID)
}
```

### API Gateway

```go
// Create API gateway
gatewayConfig := &router.APIGatewayConfig{
    ListenAddr:    "0.0.0.0",
    Port:          8080,
    EnableCORS:    true,
    EnableAuth:    false,
    RateLimit:     1000,
    Timeout:       30 * time.Second,
    EnableMetrics: true,
    EnableHealth:  true,
}

gateway, err := router.NewAPIGateway(router, gatewayConfig)
if err != nil {
    log.Fatalf("Failed to create API gateway: %v", err)
}

// Start the API gateway
go func() {
    err := gateway.Start()
    if err != nil {
        log.Printf("API gateway error: %v", err)
    }
}()
```

### Cluster Discovery

```go
// Create discovery configuration
discoveryConfig := &router.DiscoveryConfig{
    EnableMDNS:        true,
    EnableDNS:         true,
    EnableBootstrap:   true,
    EnableBroadcast:   false,
    DiscoveryInterval: 30 * time.Second,
    BootstrapPeers:    []string{"192.168.1.100:8080"},
    DNSSeeds:          []string{"seed1.adrenochain.com"},
    ServiceName:       "adrenochain-cluster",
    Timeout:           10 * time.Second,
}

// Create cluster discovery
discovery, err := router.NewClusterDiscovery(discoveryConfig)
if err != nil {
    log.Fatalf("Failed to create cluster discovery: %v", err)
}
defer discovery.Close()

// Get discovered clusters
clusters := discovery.GetDiscoveredClusters()
log.Printf("Discovered %d clusters", len(clusters))
```

### Cluster Management

```go
// Create cluster manager
managerConfig := &router.ClusterManagerConfig{
    EnableFailover:      true,
    EnableAutoRecovery:  true,
    HealthCheckInterval: 30 * time.Second,
    FailoverTimeout:     60 * time.Second,
    RecoveryTimeout:     300 * time.Second,
    MaxFailovers:        3,
    EventBufferSize:     1000,
}

manager, err := router.NewClusterManager(managerConfig)
if err != nil {
    log.Fatalf("Failed to create cluster manager: %v", err)
}
defer manager.Close()

// Register cluster for management
err = manager.RegisterCluster(cluster)
if err != nil {
    log.Fatalf("Failed to register cluster: %v", err)
}

// Set failover policy
policy := &router.FailoverPolicy{
    ClusterID:           "api-cluster-1",
    Enabled:             true,
    MaxFailovers:        3,
    FailoverTimeout:     60 * time.Second,
    HealthCheckInterval: 30 * time.Second,
    RecoveryTimeout:     300 * time.Second,
    AutoRecovery:        true,
    BackupClusters:      []router.ClusterID{"backup-cluster-1"},
}

err = manager.SetFailoverPolicy(policy)
if err != nil {
    log.Fatalf("Failed to set failover policy: %v", err)
}
```

## 🔧 Configuration

### Cluster Router Configuration

```go
type ClusterRouterConfig struct {
    MaxClusters        int           // Maximum number of clusters
    MaxNodesPerCluster int           // Maximum nodes per cluster
    HealthCheckInterval time.Duration // Health check interval
    LoadUpdateInterval  time.Duration // Load update interval
    RoutingStrategy     RoutingStrategy // Routing strategy
    EnableFailover      bool         // Enable failover
    EnableLoadBalancing bool         // Enable load balancing
    EnableMetrics       bool         // Enable metrics
    MaxRetries          int          // Maximum retries
    Timeout             time.Duration // Request timeout
}
```

### Discovery Configuration

```go
type DiscoveryConfig struct {
    EnableMDNS        bool          // Enable mDNS discovery
    EnableDNS         bool          // Enable DNS discovery
    EnableBootstrap   bool          // Enable bootstrap discovery
    EnableBroadcast   bool          // Enable broadcast discovery
    DiscoveryInterval time.Duration // Discovery interval
    BootstrapPeers    []string      // Bootstrap peer addresses
    DNSSeeds          []string      // DNS seed addresses
    BroadcastPort     int           // Broadcast port
    ServiceName       string        // Service name
    Timeout           time.Duration // Discovery timeout
}
```

### API Gateway Configuration

```go
type APIGatewayConfig struct {
    ListenAddr     string        // Listen address
    Port           int           // Listen port
    EnableCORS     bool          // Enable CORS
    EnableAuth     bool          // Enable authentication
    RateLimit      int           // Rate limit (requests per second)
    Timeout        time.Duration // Request timeout
    EnableMetrics  bool          // Enable metrics endpoint
    EnableHealth   bool          // Enable health endpoint
}
```

## 📊 API Endpoints

### Health and Metrics
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus-compatible metrics

### Cluster Management
- `GET /api/v1/clusters` - List all clusters
- `POST /api/v1/clusters` - Create a new cluster
- `GET /api/v1/clusters/{id}` - Get cluster details
- `PUT /api/v1/clusters/{id}` - Update cluster
- `DELETE /api/v1/clusters/{id}` - Delete cluster

### Node Management
- `GET /api/v1/nodes` - List all nodes
- `POST /api/v1/nodes` - Create a new node
- `GET /api/v1/nodes/{id}` - Get node details
- `PUT /api/v1/nodes/{id}` - Update node
- `DELETE /api/v1/nodes/{id}` - Delete node

### Routing Operations
- `POST /api/v1/route` - Route a request
- `GET /api/v1/route/status` - Get routing status

### Discovery Operations
- `GET /api/v1/discovery/clusters` - Get discovered clusters
- `GET /api/v1/discovery/nodes` - Get discovered nodes
- `POST /api/v1/discovery/refresh` - Trigger discovery refresh

### Health Monitoring
- `GET /api/v1/health/clusters` - Get cluster health status
- `GET /api/v1/health/nodes` - Get node health status
- `POST /api/v1/health/check/{id}` - Perform health check

### Load Balancing
- `GET /api/v1/loadbalancer/strategy` - Get load balancer strategy
- `PUT /api/v1/loadbalancer/strategy` - Set load balancer strategy
- `GET /api/v1/loadbalancer/stats` - Get load balancer statistics

## 📈 Metrics

The cluster router provides comprehensive metrics including:

- **Request Metrics**: Total requests, success rate, error rate, throughput
- **Performance Metrics**: Response time, latency percentiles, connection counts
- **Cluster Metrics**: Active clusters, cluster health, cluster load
- **Node Metrics**: Active nodes, node health, node load, node utilization
- **Load Balancer Metrics**: Strategy, connection distribution, selection times
- **Health Monitor Metrics**: Check counts, success rates, failure rates

## 🔍 Monitoring

### Health Checks
- **TCP Health Checks**: Basic connectivity testing
- **HTTP Health Checks**: Application-level health verification
- **Custom Health Checks**: Pluggable health check implementations

### Alerting
- Node failure alerts
- Cluster degradation alerts
- High load alerts
- Latency threshold alerts
- Health check failure alerts

### Dashboards
- Real-time cluster status
- Node performance metrics
- Request routing statistics
- Health monitoring data
- Load balancing metrics

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./pkg/router/...

# Run tests with coverage
go test -cover ./pkg/router/...

# Run benchmarks
go test -bench=. ./pkg/router/...

# Run specific test
go test -run TestClusterRouter ./pkg/router/
```

## 🚀 Performance

The cluster router is designed for high performance:

- **Concurrent Operations**: All operations are thread-safe and optimized for concurrency
- **Efficient Data Structures**: Optimized routing tables and indexes for fast lookups
- **Minimal Overhead**: Low-latency request routing with minimal processing overhead
- **Scalable Architecture**: Supports thousands of clusters and nodes
- **Memory Efficient**: Optimized memory usage with configurable limits

## 🔒 Security

- **Authentication**: Configurable authentication for API endpoints
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **CORS Support**: Configurable CORS policies
- **Input Validation**: Comprehensive input validation and sanitization
- **Secure Communication**: Support for TLS/SSL encryption

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:

- Create an issue on GitHub
- Check the documentation
- Review the examples
- Join the community discussions

## 🔮 Roadmap

- [ ] WebSocket support for real-time updates
- [ ] GraphQL API endpoint
- [ ] Advanced authentication (JWT, OAuth2)
- [ ] Distributed tracing integration
- [ ] Machine learning-based routing optimization
- [ ] Multi-region cluster support
- [ ] Advanced monitoring dashboards
- [ ] Kubernetes integration
- [ ] Docker Compose examples
- [ ] Helm charts for deployment
