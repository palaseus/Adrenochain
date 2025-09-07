package router

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ClusterManager manages cluster lifecycle and failover
type ClusterManager struct {
	mu               sync.RWMutex
	clusters         map[ClusterID]*ManagedCluster
	nodes            map[NodeID]*ManagedNode
	failoverPolicies map[ClusterID]*FailoverPolicy
	eventHandlers    []ClusterEventHandler
	config           *ClusterManagerConfig
	logger           *Logger
	ctx              context.Context
	cancel           context.CancelFunc
}

// ManagedCluster represents a managed cluster
type ManagedCluster struct {
	Cluster         *Cluster
	Status          ClusterStatus
	LastHealthCheck time.Time
	FailoverCount   int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ManagedNode represents a managed node
type ManagedNode struct {
	Node            *Node
	Status          NodeStatus
	LastHealthCheck time.Time
	FailoverCount   int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// FailoverPolicy defines failover behavior for a cluster
type FailoverPolicy struct {
	ClusterID           ClusterID              `json:"cluster_id"`
	Enabled             bool                   `json:"enabled"`
	MaxFailovers        int                    `json:"max_failovers"`
	FailoverTimeout     time.Duration          `json:"failover_timeout"`
	HealthCheckInterval time.Duration          `json:"health_check_interval"`
	RecoveryTimeout     time.Duration          `json:"recovery_timeout"`
	AutoRecovery        bool                   `json:"auto_recovery"`
	BackupClusters      []ClusterID            `json:"backup_clusters"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// ClusterEventHandler handles cluster events
type ClusterEventHandler interface {
	OnClusterEvent(event *ClusterEvent) error
}

// ClusterManagerConfig holds configuration for the cluster manager
type ClusterManagerConfig struct {
	EnableFailover      bool          `json:"enable_failover"`
	EnableAutoRecovery  bool          `json:"enable_auto_recovery"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	FailoverTimeout     time.Duration `json:"failover_timeout"`
	RecoveryTimeout     time.Duration `json:"recovery_timeout"`
	MaxFailovers        int           `json:"max_failovers"`
	EventBufferSize     int           `json:"event_buffer_size"`
}

// DefaultClusterManagerConfig returns the default cluster manager configuration
func DefaultClusterManagerConfig() *ClusterManagerConfig {
	return &ClusterManagerConfig{
		EnableFailover:      true,
		EnableAutoRecovery:  true,
		HealthCheckInterval: 30 * time.Second,
		FailoverTimeout:     60 * time.Second,
		RecoveryTimeout:     300 * time.Second,
		MaxFailovers:        3,
		EventBufferSize:     1000,
	}
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(config *ClusterManagerConfig) (*ClusterManager, error) {
	if config == nil {
		config = DefaultClusterManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &ClusterManager{
		clusters:         make(map[ClusterID]*ManagedCluster),
		nodes:            make(map[NodeID]*ManagedNode),
		failoverPolicies: make(map[ClusterID]*FailoverPolicy),
		eventHandlers:    make([]ClusterEventHandler, 0),
		config:           config,
		logger:           NewLogger("cluster-manager"),
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start background processes
	go manager.startHealthMonitoring()
	go manager.startFailoverMonitoring()
	go manager.startRecoveryMonitoring()

	return manager, nil
}

// RegisterCluster registers a cluster for management
func (cm *ClusterManager) RegisterCluster(cluster *Cluster) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	managedCluster := &ManagedCluster{
		Cluster:   cluster,
		Status:    cluster.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cm.clusters[cluster.ID] = managedCluster

	// Create default failover policy
	policy := &FailoverPolicy{
		ClusterID:           cluster.ID,
		Enabled:             cm.config.EnableFailover,
		MaxFailovers:        cm.config.MaxFailovers,
		FailoverTimeout:     cm.config.FailoverTimeout,
		HealthCheckInterval: cm.config.HealthCheckInterval,
		RecoveryTimeout:     cm.config.RecoveryTimeout,
		AutoRecovery:        cm.config.EnableAutoRecovery,
		BackupClusters:      []ClusterID{},
		Metadata:            make(map[string]interface{}),
	}

	cm.failoverPolicies[cluster.ID] = policy

	// Emit cluster created event
	event := &ClusterEvent{
		ID:        fmt.Sprintf("cluster-created-%d", time.Now().UnixNano()),
		Type:      ClusterEventTypeClusterCreated,
		ClusterID: cluster.ID,
		Timestamp: time.Now(),
		Data:      cluster,
		Metadata:  make(map[string]interface{}),
	}

	cm.emitEvent(event)

	cm.logger.Info("Registered cluster %s for management", cluster.ID)
	return nil
}

// UnregisterCluster unregisters a cluster from management
func (cm *ClusterManager) UnregisterCluster(clusterID ClusterID) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if managedCluster, exists := cm.clusters[clusterID]; exists {
		// Emit cluster deleted event
		event := &ClusterEvent{
			ID:        fmt.Sprintf("cluster-deleted-%d", time.Now().UnixNano()),
			Type:      ClusterEventTypeClusterDeleted,
			ClusterID: clusterID,
			Timestamp: time.Now(),
			Data:      managedCluster.Cluster,
			Metadata:  make(map[string]interface{}),
		}

		cm.emitEvent(event)

		delete(cm.clusters, clusterID)
		delete(cm.failoverPolicies, clusterID)

		cm.logger.Info("Unregistered cluster %s from management", clusterID)
	}

	return nil
}

// RegisterNode registers a node for management
func (cm *ClusterManager) RegisterNode(node *Node) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	managedNode := &ManagedNode{
		Node:      node,
		Status:    node.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cm.nodes[node.ID] = managedNode

	// Emit node joined event
	event := &ClusterEvent{
		ID:        fmt.Sprintf("node-joined-%d", time.Now().UnixNano()),
		Type:      ClusterEventTypeNodeJoined,
		ClusterID: node.ClusterID,
		NodeID:    node.ID,
		Timestamp: time.Now(),
		Data:      node,
		Metadata:  make(map[string]interface{}),
	}

	cm.emitEvent(event)

	cm.logger.Info("Registered node %s for management", node.ID)
	return nil
}

// UnregisterNode unregisters a node from management
func (cm *ClusterManager) UnregisterNode(nodeID NodeID) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if managedNode, exists := cm.nodes[nodeID]; exists {
		// Emit node left event
		event := &ClusterEvent{
			ID:        fmt.Sprintf("node-left-%d", time.Now().UnixNano()),
			Type:      ClusterEventTypeNodeLeft,
			ClusterID: managedNode.Node.ClusterID,
			NodeID:    nodeID,
			Timestamp: time.Now(),
			Data:      managedNode.Node,
			Metadata:  make(map[string]interface{}),
		}

		cm.emitEvent(event)

		delete(cm.nodes, nodeID)

		cm.logger.Info("Unregistered node %s from management", nodeID)
	}

	return nil
}

// SetFailoverPolicy sets the failover policy for a cluster
func (cm *ClusterManager) SetFailoverPolicy(policy *FailoverPolicy) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.failoverPolicies[policy.ClusterID] = policy
	cm.logger.Info("Set failover policy for cluster %s", policy.ClusterID)
	return nil
}

// GetFailoverPolicy returns the failover policy for a cluster
func (cm *ClusterManager) GetFailoverPolicy(clusterID ClusterID) (*FailoverPolicy, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	policy, exists := cm.failoverPolicies[clusterID]
	return policy, exists
}

// AddEventHandler adds a cluster event handler
func (cm *ClusterManager) AddEventHandler(handler ClusterEventHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.eventHandlers = append(cm.eventHandlers, handler)
}

// RemoveEventHandler removes a cluster event handler
func (cm *ClusterManager) RemoveEventHandler(handler ClusterEventHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, h := range cm.eventHandlers {
		if h == handler {
			cm.eventHandlers = append(cm.eventHandlers[:i], cm.eventHandlers[i+1:]...)
			break
		}
	}
}

// GetClusterStatus returns the status of all managed clusters
func (cm *ClusterManager) GetClusterStatus() map[ClusterID]*ManagedCluster {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[ClusterID]*ManagedCluster)
	for id, cluster := range cm.clusters {
		status[id] = cluster
	}
	return status
}

// GetNodeStatus returns the status of all managed nodes
func (cm *ClusterManager) GetNodeStatus() map[NodeID]*ManagedNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[NodeID]*ManagedNode)
	for id, node := range cm.nodes {
		status[id] = node
	}
	return status
}

// TriggerFailover triggers failover for a cluster
func (cm *ClusterManager) TriggerFailover(clusterID ClusterID) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	managedCluster, exists := cm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	policy, exists := cm.failoverPolicies[clusterID]
	if !exists || !policy.Enabled {
		return fmt.Errorf("failover not enabled for cluster %s", clusterID)
	}

	if managedCluster.FailoverCount >= policy.MaxFailovers {
		return fmt.Errorf("maximum failovers reached for cluster %s", clusterID)
	}

	// Perform failover
	cm.performFailover(managedCluster, policy)

	return nil
}

// performFailover performs the actual failover
func (cm *ClusterManager) performFailover(managedCluster *ManagedCluster, policy *FailoverPolicy) {
	cm.logger.Info("Performing failover for cluster %s", managedCluster.Cluster.ID)

	// Update cluster status
	managedCluster.Status = ClusterStatusFailed
	managedCluster.FailoverCount++
	managedCluster.UpdatedAt = time.Now()

	// Emit failover event
	event := &ClusterEvent{
		ID:        fmt.Sprintf("cluster-failed-%d", time.Now().UnixNano()),
		Type:      ClusterEventTypeNodeFailed,
		ClusterID: managedCluster.Cluster.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"failover_count": managedCluster.FailoverCount,
			"reason":         "manual_failover",
		},
		Metadata: make(map[string]interface{}),
	}

	cm.emitEvent(event)

	// If auto-recovery is enabled, schedule recovery
	if policy.AutoRecovery {
		go cm.scheduleRecovery(managedCluster.Cluster.ID, policy.RecoveryTimeout)
	}
}

// scheduleRecovery schedules cluster recovery
func (cm *ClusterManager) scheduleRecovery(clusterID ClusterID, timeout time.Duration) {
	cm.logger.Info("Scheduling recovery for cluster %s in %v", clusterID, timeout)

	time.Sleep(timeout)

	cm.mu.Lock()
	managedCluster, exists := cm.clusters[clusterID]
	if !exists {
		cm.mu.Unlock()
		return
	}
	cm.mu.Unlock()

	// Attempt recovery
	if err := cm.attemptRecovery(managedCluster); err != nil {
		cm.logger.Error("Failed to recover cluster %s: %v", clusterID, err)
	}
}

// attemptRecovery attempts to recover a cluster
func (cm *ClusterManager) attemptRecovery(managedCluster *ManagedCluster) error {
	cm.logger.Info("Attempting recovery for cluster %s", managedCluster.Cluster.ID)

	// Check if cluster is healthy
	if cm.isClusterHealthy(managedCluster.Cluster) {
		// Update status
		managedCluster.Status = ClusterStatusActive
		managedCluster.UpdatedAt = time.Now()

		// Emit recovery event
		event := &ClusterEvent{
			ID:        fmt.Sprintf("cluster-recovered-%d", time.Now().UnixNano()),
			Type:      ClusterEventTypeClusterCreated, // Reuse event type for recovery
			ClusterID: managedCluster.Cluster.ID,
			Timestamp: time.Now(),
			Data:      managedCluster.Cluster,
			Metadata:  make(map[string]interface{}),
		}

		cm.emitEvent(event)

		cm.logger.Info("Successfully recovered cluster %s", managedCluster.Cluster.ID)
		return nil
	}

	return fmt.Errorf("cluster %s is not healthy", managedCluster.Cluster.ID)
}

// isClusterHealthy checks if a cluster is healthy
func (cm *ClusterManager) isClusterHealthy(cluster *Cluster) bool {
	// Simple health check - in a real implementation, this would be more sophisticated
	return cluster.Status == ClusterStatusActive && len(cluster.Nodes) > 0
}

// startHealthMonitoring starts the health monitoring process
func (cm *ClusterManager) startHealthMonitoring() {
	ticker := time.NewTicker(cm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.performHealthChecks()
		}
	}
}

// startFailoverMonitoring starts the failover monitoring process
func (cm *ClusterManager) startFailoverMonitoring() {
	ticker := time.NewTicker(cm.config.FailoverTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.checkFailoverConditions()
		}
	}
}

// startRecoveryMonitoring starts the recovery monitoring process
func (cm *ClusterManager) startRecoveryMonitoring() {
	ticker := time.NewTicker(cm.config.RecoveryTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.checkRecoveryConditions()
		}
	}
}

// performHealthChecks performs health checks on all managed clusters and nodes
func (cm *ClusterManager) performHealthChecks() {
	cm.mu.RLock()
	clusters := make([]*ManagedCluster, 0, len(cm.clusters))
	nodes := make([]*ManagedNode, 0, len(cm.nodes))

	for _, cluster := range cm.clusters {
		clusters = append(clusters, cluster)
	}
	for _, node := range cm.nodes {
		nodes = append(nodes, node)
	}
	cm.mu.RUnlock()

	// Check cluster health
	for _, managedCluster := range clusters {
		cm.checkClusterHealth(managedCluster)
	}

	// Check node health
	for _, managedNode := range nodes {
		cm.checkNodeHealth(managedNode)
	}
}

// checkClusterHealth checks the health of a managed cluster
func (cm *ClusterManager) checkClusterHealth(managedCluster *ManagedCluster) {
	// Update last health check time
	managedCluster.LastHealthCheck = time.Now()

	// Check if cluster is healthy
	if !cm.isClusterHealthy(managedCluster.Cluster) {
		// Mark cluster as degraded or failed
		if managedCluster.Status == ClusterStatusActive {
			managedCluster.Status = ClusterStatusDegraded
			managedCluster.UpdatedAt = time.Now()

			// Emit health change event
			event := &ClusterEvent{
				ID:        fmt.Sprintf("cluster-health-changed-%d", time.Now().UnixNano()),
				Type:      ClusterEventTypeHealthChanged,
				ClusterID: managedCluster.Cluster.ID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"old_status": ClusterStatusActive,
					"new_status": ClusterStatusDegraded,
				},
				Metadata: make(map[string]interface{}),
			}

			cm.emitEvent(event)
		}
	}
}

// checkNodeHealth checks the health of a managed node
func (cm *ClusterManager) checkNodeHealth(managedNode *ManagedNode) {
	// Update last health check time
	managedNode.LastHealthCheck = time.Now()

	// Simple health check - in a real implementation, this would be more sophisticated
	if managedNode.Node.HealthScore < 0.5 {
		// Mark node as degraded or failed
		if managedNode.Status == NodeStatusActive {
			managedNode.Status = NodeStatusDegraded
			managedNode.UpdatedAt = time.Now()

			// Emit node failed event
			event := &ClusterEvent{
				ID:        fmt.Sprintf("node-failed-%d", time.Now().UnixNano()),
				Type:      ClusterEventTypeNodeFailed,
				ClusterID: managedNode.Node.ClusterID,
				NodeID:    managedNode.Node.ID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"old_status":   NodeStatusActive,
					"new_status":   NodeStatusDegraded,
					"health_score": managedNode.Node.HealthScore,
				},
				Metadata: make(map[string]interface{}),
			}

			cm.emitEvent(event)
		}
	}
}

// checkFailoverConditions checks if failover conditions are met
func (cm *ClusterManager) checkFailoverConditions() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for clusterID, managedCluster := range cm.clusters {
		policy, exists := cm.failoverPolicies[clusterID]
		if !exists || !policy.Enabled {
			continue
		}

		// Check if cluster has been unhealthy for too long
		if managedCluster.Status == ClusterStatusDegraded {
			timeSinceLastHealthCheck := time.Since(managedCluster.LastHealthCheck)
			if timeSinceLastHealthCheck > policy.FailoverTimeout {
				// Trigger automatic failover
				go cm.performFailover(managedCluster, policy)
			}
		}
	}
}

// checkRecoveryConditions checks if recovery conditions are met
func (cm *ClusterManager) checkRecoveryConditions() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for clusterID, managedCluster := range cm.clusters {
		policy, exists := cm.failoverPolicies[clusterID]
		if !exists || !policy.AutoRecovery {
			continue
		}

		// Check if cluster has been failed for too long
		if managedCluster.Status == ClusterStatusFailed {
			timeSinceLastUpdate := time.Since(managedCluster.UpdatedAt)
			if timeSinceLastUpdate > policy.RecoveryTimeout {
				// Attempt recovery
				go cm.attemptRecovery(managedCluster)
			}
		}
	}
}

// emitEvent emits a cluster event to all registered handlers
func (cm *ClusterManager) emitEvent(event *ClusterEvent) {
	for _, handler := range cm.eventHandlers {
		go func(h ClusterEventHandler) {
			if err := h.OnClusterEvent(event); err != nil {
				cm.logger.Error("Event handler failed: %v", err)
			}
		}(handler)
	}
}

// Close shuts down the cluster manager
func (cm *ClusterManager) Close() error {
	cm.cancel()
	cm.logger.Info("Cluster manager shut down")
	return nil
}
