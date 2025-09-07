package router

import (
	"time"
)

// Request represents a routing request
type Request struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	ClusterType ClusterType            `json:"cluster_type"`
	Region      string                 `json:"region"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Response represents a routing response
type Response struct {
	ID        string                 `json:"id"`
	RequestID string                 `json:"request_id"`
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data"`
	Error     string                 `json:"error,omitempty"`
	NodeID    NodeID                 `json:"node_id"`
	ClusterID ClusterID              `json:"cluster_id"`
	Latency   time.Duration          `json:"latency"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// ClusterDiscoveryRequest represents a cluster discovery request
type ClusterDiscoveryRequest struct {
	RequesterID  string                 `json:"requester_id"`
	ClusterType  ClusterType            `json:"cluster_type"`
	Region       string                 `json:"region"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ClusterDiscoveryResponse represents a cluster discovery response
type ClusterDiscoveryResponse struct {
	Clusters   []*Cluster             `json:"clusters"`
	Nodes      []*Node                `json:"nodes"`
	TotalFound int                    `json:"total_found"`
	Latency    time.Duration          `json:"latency"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ClusterJoinRequest represents a request to join a cluster
type ClusterJoinRequest struct {
	NodeID       NodeID                 `json:"node_id"`
	ClusterID    ClusterID              `json:"cluster_id"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ClusterJoinResponse represents a response to a cluster join request
type ClusterJoinResponse struct {
	Success   bool                   `json:"success"`
	ClusterID ClusterID              `json:"cluster_id"`
	NodeID    NodeID                 `json:"node_id"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ClusterLeaveRequest represents a request to leave a cluster
type ClusterLeaveRequest struct {
	NodeID    NodeID                 `json:"node_id"`
	ClusterID ClusterID              `json:"cluster_id"`
	Reason    string                 `json:"reason"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ClusterLeaveResponse represents a response to a cluster leave request
type ClusterLeaveResponse struct {
	Success  bool                   `json:"success"`
	NodeID   NodeID                 `json:"node_id"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	NodeID      NodeID                 `json:"node_id"`
	ClusterID   ClusterID              `json:"cluster_id"`
	Status      NodeStatus             `json:"status"`
	Load        float64                `json:"load"`
	HealthScore float64                `json:"health_score"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// HeartbeatResponse represents a heartbeat response
type HeartbeatResponse struct {
	Success   bool                   `json:"success"`
	NodeID    NodeID                 `json:"node_id"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	ID              ClusterID              `json:"id"`
	Name            string                 `json:"name"`
	Type            ClusterType            `json:"type"`
	Region          string                 `json:"region"`
	MaxNodes        int                    `json:"max_nodes"`
	MinNodes        int                    `json:"min_nodes"`
	ConsensusType   string                 `json:"consensus_type"`
	LoadThreshold   float64                `json:"load_threshold"`
	HealthThreshold float64                `json:"health_threshold"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NodeConfig represents node configuration
type NodeConfig struct {
	ID             NodeID                 `json:"id"`
	Address        string                 `json:"address"`
	Port           int                    `json:"port"`
	ClusterID      ClusterID              `json:"cluster_id"`
	Capabilities   []string               `json:"capabilities"`
	MaxConnections int                    `json:"max_connections"`
	MaxLoad        float64                `json:"max_load"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RoutingConfig represents routing configuration
type RoutingConfig struct {
	Strategy            RoutingStrategy        `json:"strategy"`
	MaxRetries          int                    `json:"max_retries"`
	Timeout             time.Duration          `json:"timeout"`
	EnableFailover      bool                   `json:"enable_failover"`
	EnableLoadBalancing bool                   `json:"enable_load_balancing"`
	HealthCheckInterval time.Duration          `json:"health_check_interval"`
	LoadUpdateInterval  time.Duration          `json:"load_update_interval"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// ClusterEvent represents a cluster event
type ClusterEvent struct {
	ID        string                 `json:"id"`
	Type      ClusterEventType       `json:"type"`
	ClusterID ClusterID              `json:"cluster_id"`
	NodeID    NodeID                 `json:"node_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ClusterEventType represents the type of cluster event
type ClusterEventType string

const (
	ClusterEventTypeNodeJoined     ClusterEventType = "node_joined"
	ClusterEventTypeNodeLeft       ClusterEventType = "node_left"
	ClusterEventTypeNodeFailed     ClusterEventType = "node_failed"
	ClusterEventTypeNodeRecovered  ClusterEventType = "node_recovered"
	ClusterEventTypeClusterCreated ClusterEventType = "cluster_created"
	ClusterEventTypeClusterDeleted ClusterEventType = "cluster_deleted"
	ClusterEventTypeLoadChanged    ClusterEventType = "load_changed"
	ClusterEventTypeHealthChanged  ClusterEventType = "health_changed"
)

// ClusterStats represents cluster statistics
type ClusterStats struct {
	ClusterID          ClusterID     `json:"cluster_id"`
	TotalNodes         int           `json:"total_nodes"`
	ActiveNodes        int           `json:"active_nodes"`
	InactiveNodes      int           `json:"inactive_nodes"`
	FailedNodes        int           `json:"failed_nodes"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	AvgLoad            float64       `json:"avg_load"`
	AvgHealthScore     float64       `json:"avg_health_score"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// NodeStats represents node statistics
type NodeStats struct {
	NodeID             NodeID        `json:"node_id"`
	ClusterID          ClusterID     `json:"cluster_id"`
	Status             NodeStatus    `json:"status"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	CurrentLoad        float64       `json:"current_load"`
	HealthScore        float64       `json:"health_score"`
	Uptime             time.Duration `json:"uptime"`
	LastHeartbeat      time.Time     `json:"last_heartbeat"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// RoutingStats represents routing statistics
type RoutingStats struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	SuccessRate        float64       `json:"success_rate"`
	ErrorRate          float64       `json:"error_rate"`
	Throughput         float64       `json:"throughput"`
	ActiveClusters     int           `json:"active_clusters"`
	ActiveNodes        int           `json:"active_nodes"`
	UnhealthyNodes     int           `json:"unhealthy_nodes"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// ClusterTopology represents the cluster topology
type ClusterTopology struct {
	Clusters    map[ClusterID]*ClusterTopologyInfo `json:"clusters"`
	Connections []*ClusterConnection               `json:"connections"`
	LastUpdated time.Time                          `json:"last_updated"`
}

// ClusterTopologyInfo represents topology information for a cluster
type ClusterTopologyInfo struct {
	ClusterID ClusterID              `json:"cluster_id"`
	Nodes     []NodeID               `json:"nodes"`
	Leader    NodeID                 `json:"leader"`
	Status    ClusterStatus          `json:"status"`
	Load      float64                `json:"load"`
	Latency   time.Duration          `json:"latency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ClusterConnection represents a connection between clusters
type ClusterConnection struct {
	FromCluster ClusterID              `json:"from_cluster"`
	ToCluster   ClusterID              `json:"to_cluster"`
	Type        ConnectionType         `json:"type"`
	Latency     time.Duration          `json:"latency"`
	Bandwidth   int64                  `json:"bandwidth"`
	Status      ConnectionStatus       `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ConnectionType represents the type of cluster connection
type ConnectionType string

const (
	ConnectionTypeDirect ConnectionType = "direct"
	ConnectionTypeRelay  ConnectionType = "relay"
	ConnectionTypeBridge ConnectionType = "bridge"
	ConnectionTypeTunnel ConnectionType = "tunnel"
)

// ConnectionStatus represents the status of a cluster connection
type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "active"
	ConnectionStatusInactive ConnectionStatus = "inactive"
	ConnectionStatusDegraded ConnectionStatus = "degraded"
	ConnectionStatusFailed   ConnectionStatus = "failed"
)

// ClusterPolicy represents cluster policies
type ClusterPolicy struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      PolicyType             `json:"type"`
	Rules     []*PolicyRule          `json:"rules"`
	Priority  int                    `json:"priority"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// PolicyType represents the type of policy
type PolicyType string

const (
	PolicyTypeRouting       PolicyType = "routing"
	PolicyTypeLoadBalancing PolicyType = "load_balancing"
	PolicyTypeHealth        PolicyType = "health"
	PolicyTypeSecurity      PolicyType = "security"
	PolicyTypeAccess        PolicyType = "access"
)

// PolicyRule represents a policy rule
type PolicyRule struct {
	ID         string                 `json:"id"`
	Condition  string                 `json:"condition"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
	Enabled    bool                   `json:"enabled"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ClusterAlert represents a cluster alert
type ClusterAlert struct {
	ID           string                 `json:"id"`
	Type         AlertType              `json:"type"`
	Severity     AlertSeverity          `json:"severity"`
	ClusterID    ClusterID              `json:"cluster_id"`
	NodeID       NodeID                 `json:"node_id"`
	Message      string                 `json:"message"`
	Timestamp    time.Time              `json:"timestamp"`
	Acknowledged bool                   `json:"acknowledged"`
	Resolved     bool                   `json:"resolved"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeNodeDown        AlertType = "node_down"
	AlertTypeNodeDegraded    AlertType = "node_degraded"
	AlertTypeHighLoad        AlertType = "high_load"
	AlertTypeHighLatency     AlertType = "high_latency"
	AlertTypeClusterDown     AlertType = "cluster_down"
	AlertTypeClusterDegraded AlertType = "cluster_degraded"
	AlertTypeSecurity        AlertType = "security"
	AlertTypePerformance     AlertType = "performance"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo      AlertSeverity = "info"
	AlertSeverityWarning   AlertSeverity = "warning"
	AlertSeverityCritical  AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)
