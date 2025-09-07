package router

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// LoadBalancer implements various load balancing strategies
type LoadBalancer struct {
	strategy        RoutingStrategy
	roundRobinIndex int
	mu              sync.RWMutex
	weights         map[NodeID]float64
	connections     map[NodeID]int64
	lastUsed        map[NodeID]time.Time
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy RoutingStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy:        strategy,
		roundRobinIndex: 0,
		weights:         make(map[NodeID]float64),
		connections:     make(map[NodeID]int64),
		lastUsed:        make(map[NodeID]time.Time),
	}
}

// SelectNode selects the best node based on the configured strategy
func (lb *LoadBalancer) SelectNode(nodes []*Node, req *Request) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Filter nodes based on request requirements
	candidates := lb.filterNodes(nodes, req)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable nodes found")
	}

	// Apply load balancing strategy
	switch lb.strategy {
	case RoutingStrategyRoundRobin:
		return lb.selectRoundRobin(candidates)
	case RoutingStrategyLeastConn:
		return lb.selectLeastConnections(candidates)
	case RoutingStrategyLeastLatency:
		return lb.selectLeastLatency(candidates)
	case RoutingStrategyLeastLoad:
		return lb.selectLeastLoad(candidates)
	case RoutingStrategyWeighted:
		return lb.selectWeighted(candidates)
	case RoutingStrategyAdaptive:
		return lb.selectAdaptive(candidates, req)
	default:
		return lb.selectRoundRobin(candidates)
	}
}

// filterNodes filters nodes based on request requirements
func (lb *LoadBalancer) filterNodes(nodes []*Node, req *Request) []*Node {
	var candidates []*Node

	for _, node := range nodes {
		// Check if node is active
		if node.Status != NodeStatusActive {
			continue
		}

		// Check health score
		if node.HealthScore < 0.5 {
			continue
		}

		// Check load threshold
		if node.Load > 0.9 {
			continue
		}

		// Check latency threshold
		if node.Latency > 5*time.Second {
			continue
		}

		// Check region preference
		if req.Region != "" {
			if region, ok := node.Metadata["region"].(string); ok && region != req.Region {
				continue
			}
		}

		candidates = append(candidates, node)
	}

	return candidates
}

// selectRoundRobin selects nodes in a round-robin fashion
func (lb *LoadBalancer) selectRoundRobin(nodes []*Node) (*Node, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	selected := nodes[lb.roundRobinIndex%len(nodes)]
	lb.roundRobinIndex++

	// Update last used time
	lb.lastUsed[selected.ID] = time.Now()

	return selected, nil
}

// selectLeastConnections selects the node with the least active connections
func (lb *LoadBalancer) selectLeastConnections(nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Sort by connection count
	sort.Slice(nodes, func(i, j int) bool {
		connI := lb.connections[nodes[i].ID]
		connJ := lb.connections[nodes[j].ID]
		return connI < connJ
	})

	selected := nodes[0]
	lb.lastUsed[selected.ID] = time.Now()

	return selected, nil
}

// selectLeastLatency selects the node with the lowest latency
func (lb *LoadBalancer) selectLeastLatency(nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Sort by latency
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Latency < nodes[j].Latency
	})

	selected := nodes[0]
	lb.mu.Lock()
	lb.lastUsed[selected.ID] = time.Now()
	lb.mu.Unlock()

	return selected, nil
}

// selectLeastLoad selects the node with the lowest load
func (lb *LoadBalancer) selectLeastLoad(nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Sort by load
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Load < nodes[j].Load
	})

	selected := nodes[0]
	lb.mu.Lock()
	lb.lastUsed[selected.ID] = time.Now()
	lb.mu.Unlock()

	return selected, nil
}

// selectWeighted selects a node based on weighted distribution
func (lb *LoadBalancer) selectWeighted(nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Calculate total weight
	totalWeight := 0.0
	for _, node := range nodes {
		weight := lb.getNodeWeight(node)
		totalWeight += weight
	}

	if totalWeight == 0 {
		// Fallback to round-robin if no weights
		return lb.selectRoundRobin(nodes)
	}

	// Select based on weighted random
	random := rand.Float64() * totalWeight
	currentWeight := 0.0

	for _, node := range nodes {
		weight := lb.getNodeWeight(node)
		currentWeight += weight
		if random <= currentWeight {
			lb.lastUsed[node.ID] = time.Now()
			return node, nil
		}
	}

	// Fallback to last node
	selected := nodes[len(nodes)-1]
	lb.lastUsed[selected.ID] = time.Now()
	return selected, nil
}

// selectAdaptive selects a node using adaptive algorithms
func (lb *LoadBalancer) selectAdaptive(nodes []*Node, req *Request) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Calculate adaptive scores for each node
	scores := make([]float64, len(nodes))
	for i, node := range nodes {
		scores[i] = lb.calculateAdaptiveScore(node, req)
	}

	// Find node with highest score
	bestIndex := 0
	bestScore := scores[0]

	for i, score := range scores {
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}

	selected := nodes[bestIndex]
	lb.mu.Lock()
	lb.lastUsed[selected.ID] = time.Now()
	lb.mu.Unlock()

	return selected, nil
}

// getNodeWeight calculates the weight for a node
func (lb *LoadBalancer) getNodeWeight(node *Node) float64 {
	// Base weight from configuration
	weight := lb.weights[node.ID]
	if weight == 0 {
		weight = 1.0 // Default weight
	}

	// Adjust based on health score
	weight *= node.HealthScore

	// Adjust based on load (inverse relationship)
	weight *= (1.0 - node.Load)

	// Adjust based on latency (inverse relationship)
	if node.Latency > 0 {
		latencyFactor := 1.0 / (1.0 + float64(node.Latency.Milliseconds())/1000.0)
		weight *= latencyFactor
	}

	return weight
}

// calculateAdaptiveScore calculates an adaptive score for node selection
func (lb *LoadBalancer) calculateAdaptiveScore(node *Node, req *Request) float64 {
	score := 1.0

	// Health score factor
	score *= node.HealthScore

	// Load factor (lower load is better)
	score *= (1.0 - node.Load)

	// Latency factor (lower latency is better)
	if node.Latency > 0 {
		latencyFactor := 1.0 / (1.0 + float64(node.Latency.Milliseconds())/1000.0)
		score *= latencyFactor
	}

	// Capacity factor
	if node.Capacity > 0 {
		capacityUtilization := float64(node.RequestCount) / float64(node.Capacity)
		score *= (1.0 - capacityUtilization*0.5)
	}

	// Recent usage factor (prefer less recently used nodes)
	lb.mu.RLock()
	lastUsed := lb.lastUsed[node.ID]
	lb.mu.RUnlock()

	if !lastUsed.IsZero() {
		timeSinceLastUse := time.Since(lastUsed)
		if timeSinceLastUse > 0 {
			// Boost score for nodes not used recently
			score *= (1.0 + float64(timeSinceLastUse.Seconds())/60.0)
		}
	}

	// Request type factor
	if req.Type != "" {
		if nodeType, ok := node.Metadata["type"].(string); ok {
			if nodeType == req.Type {
				score *= 1.2 // Boost for matching type
			}
		}
	}

	// Region factor
	if req.Region != "" {
		if region, ok := node.Metadata["region"].(string); ok {
			if region == req.Region {
				score *= 1.1 // Boost for matching region
			}
		}
	}

	return score
}

// SetNodeWeight sets the weight for a specific node
func (lb *LoadBalancer) SetNodeWeight(nodeID NodeID, weight float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.weights[nodeID] = weight
}

// RecordConnection records a connection to a node
func (lb *LoadBalancer) RecordConnection(nodeID NodeID) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.connections[nodeID]++
}

// RecordDisconnection records a disconnection from a node
func (lb *LoadBalancer) RecordDisconnection(nodeID NodeID) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.connections[nodeID] > 0 {
		lb.connections[nodeID]--
	}
}

// GetConnectionCount returns the connection count for a node
func (lb *LoadBalancer) GetConnectionCount(nodeID NodeID) int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.connections[nodeID]
}

// SetStrategy changes the load balancing strategy
func (lb *LoadBalancer) SetStrategy(strategy RoutingStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.strategy = strategy
}

// GetStrategy returns the current load balancing strategy
func (lb *LoadBalancer) GetStrategy() RoutingStrategy {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.strategy
}

// GetStats returns load balancer statistics
func (lb *LoadBalancer) GetStats() *LoadBalancerStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := &LoadBalancerStats{
		Strategy:         lb.strategy,
		TotalWeights:     len(lb.weights),
		TotalConnections: 0,
		NodeStats:        make(map[NodeID]*NodeLoadStats),
	}

	// Calculate total connections and node stats
	for nodeID, connections := range lb.connections {
		stats.TotalConnections += connections

		nodeStats := &NodeLoadStats{
			Connections: connections,
			Weight:      lb.weights[nodeID],
			LastUsed:    lb.lastUsed[nodeID],
		}

		stats.NodeStats[nodeID] = nodeStats
	}

	return stats
}

// LoadBalancerStats contains statistics about the load balancer
type LoadBalancerStats struct {
	Strategy         RoutingStrategy
	TotalWeights     int
	TotalConnections int64
	NodeStats        map[NodeID]*NodeLoadStats
}

// NodeLoadStats contains load statistics for a specific node
type NodeLoadStats struct {
	Connections int64
	Weight      float64
	LastUsed    time.Time
}

// Reset resets the load balancer state
func (lb *LoadBalancer) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.roundRobinIndex = 0
	lb.weights = make(map[NodeID]float64)
	lb.connections = make(map[NodeID]int64)
	lb.lastUsed = make(map[NodeID]time.Time)
}
