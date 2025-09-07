package router

import (
	"sync"
	"time"
)

// RoutingTable manages the routing information for clusters and nodes
type RoutingTable struct {
	mu       sync.RWMutex
	clusters map[ClusterID]*ClusterEntry
	nodes    map[NodeID]*NodeEntry
	indexes  map[string][]NodeID // Index by various criteria
}

// ClusterEntry represents a cluster entry in the routing table
type ClusterEntry struct {
	Cluster     *Cluster
	LastUpdate  time.Time
	RouteCount  int64
	SuccessRate float64
}

// NodeEntry represents a node entry in the routing table
type NodeEntry struct {
	Node        *Node
	LastUpdate  time.Time
	RouteCount  int64
	SuccessRate float64
	Latency     time.Duration
}

// NewRoutingTable creates a new routing table
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		clusters: make(map[ClusterID]*ClusterEntry),
		nodes:    make(map[NodeID]*NodeEntry),
		indexes:  make(map[string][]NodeID),
	}
}

// UpdateCluster updates cluster information in the routing table
func (rt *RoutingTable) UpdateCluster(cluster *Cluster) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	entry, exists := rt.clusters[cluster.ID]
	if !exists {
		entry = &ClusterEntry{
			Cluster:    cluster,
			LastUpdate: time.Now(),
		}
		rt.clusters[cluster.ID] = entry
	} else {
		entry.Cluster = cluster
		entry.LastUpdate = time.Now()
	}

	// Update node entries for this cluster
	for _, node := range cluster.Nodes {
		rt.updateNode(node)
	}
}

// UpdateNode updates node information in the routing table
func (rt *RoutingTable) UpdateNode(node *Node) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.updateNode(node)
}

// updateNode is the internal method to update node information
func (rt *RoutingTable) updateNode(node *Node) {
	entry, exists := rt.nodes[node.ID]
	if !exists {
		entry = &NodeEntry{
			Node:       node,
			LastUpdate: time.Now(),
		}
		rt.nodes[node.ID] = entry
	} else {
		entry.Node = node
		entry.LastUpdate = time.Now()
	}

	// Update indexes
	rt.updateIndexes(node)
}

// updateIndexes updates the various indexes for efficient lookups
func (rt *RoutingTable) updateIndexes(node *Node) {
	// Index by cluster
	clusterKey := "cluster:" + string(node.ClusterID)
	rt.updateIndex(clusterKey, node.ID)

	// Index by status
	statusKey := "status:" + string(node.Status)
	rt.updateIndex(statusKey, node.ID)

	// Index by region (if available)
	if region, ok := node.Metadata["region"].(string); ok {
		regionKey := "region:" + region
		rt.updateIndex(regionKey, node.ID)
	}

	// Index by type (if available)
	if nodeType, ok := node.Metadata["type"].(string); ok {
		typeKey := "type:" + nodeType
		rt.updateIndex(typeKey, node.ID)
	}
}

// updateIndex updates a specific index
func (rt *RoutingTable) updateIndex(key string, nodeID NodeID) {
	nodes := rt.indexes[key]

	// Remove existing entry if present
	for i, id := range nodes {
		if id == nodeID {
			nodes = append(nodes[:i], nodes[i+1:]...)
			break
		}
	}

	// Add new entry
	nodes = append(nodes, nodeID)
	rt.indexes[key] = nodes
}

// GetCluster returns a cluster by ID
func (rt *RoutingTable) GetCluster(id ClusterID) (*Cluster, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	entry, exists := rt.clusters[id]
	if !exists {
		return nil, false
	}
	return entry.Cluster, true
}

// GetNode returns a node by ID
func (rt *RoutingTable) GetNode(id NodeID) (*Node, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	entry, exists := rt.nodes[id]
	if !exists {
		return nil, false
	}
	return entry.Node, true
}

// GetNodesByCluster returns all nodes in a specific cluster
func (rt *RoutingTable) GetNodesByCluster(clusterID ClusterID) []*Node {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	clusterKey := "cluster:" + string(clusterID)
	nodeIDs := rt.indexes[clusterKey]

	var nodes []*Node
	for _, nodeID := range nodeIDs {
		if entry, exists := rt.nodes[nodeID]; exists {
			nodes = append(nodes, entry.Node)
		}
	}

	return nodes
}

// GetNodesByStatus returns all nodes with a specific status
func (rt *RoutingTable) GetNodesByStatus(status NodeStatus) []*Node {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	statusKey := "status:" + string(status)
	nodeIDs := rt.indexes[statusKey]

	var nodes []*Node
	for _, nodeID := range nodeIDs {
		if entry, exists := rt.nodes[nodeID]; exists {
			nodes = append(nodes, entry.Node)
		}
	}

	return nodes
}

// GetNodesByRegion returns all nodes in a specific region
func (rt *RoutingTable) GetNodesByRegion(region string) []*Node {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	regionKey := "region:" + region
	nodeIDs := rt.indexes[regionKey]

	var nodes []*Node
	for _, nodeID := range nodeIDs {
		if entry, exists := rt.nodes[nodeID]; exists {
			nodes = append(nodes, entry.Node)
		}
	}

	return nodes
}

// GetActiveNodes returns all active nodes
func (rt *RoutingTable) GetActiveNodes() []*Node {
	return rt.GetNodesByStatus(NodeStatusActive)
}

// RecordRouteSuccess records a successful route
func (rt *RoutingTable) RecordRouteSuccess(nodeID NodeID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if entry, exists := rt.nodes[nodeID]; exists {
		entry.RouteCount++
		// Simple success rate calculation
		entry.SuccessRate = (entry.SuccessRate*float64(entry.RouteCount-1) + 1.0) / float64(entry.RouteCount)
	}
}

// RecordRouteFailure records a failed route
func (rt *RoutingTable) RecordRouteFailure(nodeID NodeID) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if entry, exists := rt.nodes[nodeID]; exists {
		entry.RouteCount++
		// Simple success rate calculation
		entry.SuccessRate = (entry.SuccessRate*float64(entry.RouteCount-1) + 0.0) / float64(entry.RouteCount)
	}
}

// UpdateLatency updates the latency for a node
func (rt *RoutingTable) UpdateLatency(nodeID NodeID, latency time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if entry, exists := rt.nodes[nodeID]; exists {
		entry.Latency = latency
	}
}

// GetBestNodes returns the best nodes based on various criteria
func (rt *RoutingTable) GetBestNodes(criteria NodeSelectionCriteria) []*Node {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var candidates []*Node

	// Get initial candidates based on criteria
	if criteria.ClusterID != "" {
		candidates = rt.GetNodesByCluster(criteria.ClusterID)
	} else if criteria.Status != "" {
		candidates = rt.GetNodesByStatus(criteria.Status)
	} else {
		// Get all active nodes
		candidates = rt.GetActiveNodes()
	}

	// Filter by additional criteria
	var filtered []*Node
	for _, node := range candidates {
		if rt.matchesCriteria(node, criteria) {
			filtered = append(filtered, node)
		}
	}

	// Sort by performance metrics
	rt.sortNodesByPerformance(filtered, criteria)

	// Return top N nodes
	if criteria.Limit > 0 && len(filtered) > criteria.Limit {
		return filtered[:criteria.Limit]
	}

	return filtered
}

// NodeSelectionCriteria defines criteria for node selection
type NodeSelectionCriteria struct {
	ClusterID      ClusterID
	Status         NodeStatus
	MinHealthScore float64
	MaxLoad        float64
	MaxLatency     time.Duration
	Region         string
	Limit          int
}

// matchesCriteria checks if a node matches the selection criteria
func (rt *RoutingTable) matchesCriteria(node *Node, criteria NodeSelectionCriteria) bool {
	// Check health score
	if criteria.MinHealthScore > 0 && node.HealthScore < criteria.MinHealthScore {
		return false
	}

	// Check load
	if criteria.MaxLoad > 0 && node.Load > criteria.MaxLoad {
		return false
	}

	// Check latency
	if criteria.MaxLatency > 0 && node.Latency > criteria.MaxLatency {
		return false
	}

	// Check region
	if criteria.Region != "" {
		if region, ok := node.Metadata["region"].(string); !ok || region != criteria.Region {
			return false
		}
	}

	return true
}

// sortNodesByPerformance sorts nodes by performance metrics
func (rt *RoutingTable) sortNodesByPerformance(nodes []*Node, criteria NodeSelectionCriteria) {
	// Simple sorting by health score and load
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			scoreI := rt.calculateNodeScore(nodes[i])
			scoreJ := rt.calculateNodeScore(nodes[j])

			if scoreI < scoreJ {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// calculateNodeScore calculates a performance score for a node
func (rt *RoutingTable) calculateNodeScore(node *Node) float64 {
	score := node.HealthScore

	// Factor in load (lower is better)
	score *= (1.0 - node.Load)

	// Factor in latency (lower is better)
	if node.Latency > 0 {
		score *= (1.0 / (1.0 + float64(node.Latency.Milliseconds())/1000.0))
	}

	// Factor in success rate
	if entry, exists := rt.nodes[node.ID]; exists {
		score *= entry.SuccessRate
	}

	return score
}

// Cleanup removes stale entries from the routing table
func (rt *RoutingTable) Cleanup(maxAge time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	now := time.Now()

	// Clean up stale nodes
	for nodeID, entry := range rt.nodes {
		if now.Sub(entry.LastUpdate) > maxAge {
			delete(rt.nodes, nodeID)
			rt.removeFromIndexes(nodeID)
		}
	}

	// Clean up stale clusters
	for clusterID, entry := range rt.clusters {
		if now.Sub(entry.LastUpdate) > maxAge {
			delete(rt.clusters, clusterID)
		}
	}
}

// removeFromIndexes removes a node from all indexes
func (rt *RoutingTable) removeFromIndexes(nodeID NodeID) {
	for key, nodeIDs := range rt.indexes {
		for i, id := range nodeIDs {
			if id == nodeID {
				rt.indexes[key] = append(nodeIDs[:i], nodeIDs[i+1:]...)
				break
			}
		}
	}
}

// GetStats returns statistics about the routing table
func (rt *RoutingTable) GetStats() *RoutingTableStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	stats := &RoutingTableStats{
		ClusterCount: len(rt.clusters),
		NodeCount:    len(rt.nodes),
		IndexCount:   len(rt.indexes),
	}

	// Count nodes by status
	statusCounts := make(map[NodeStatus]int)
	for _, entry := range rt.nodes {
		statusCounts[entry.Node.Status]++
	}
	stats.NodeStatusCounts = statusCounts

	// Count clusters by status
	clusterStatusCounts := make(map[ClusterStatus]int)
	for _, entry := range rt.clusters {
		clusterStatusCounts[entry.Cluster.Status]++
	}
	stats.ClusterStatusCounts = clusterStatusCounts

	return stats
}

// RoutingTableStats contains statistics about the routing table
type RoutingTableStats struct {
	ClusterCount        int
	NodeCount           int
	IndexCount          int
	NodeStatusCounts    map[NodeStatus]int
	ClusterStatusCounts map[ClusterStatus]int
}
