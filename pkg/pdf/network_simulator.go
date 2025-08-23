package pdf

import (
	"math/rand"
	"sync"
	"time"
)

// NetworkSimulator provides realistic network simulation
type NetworkSimulator struct {
	config     *NetworkSimConfig
	rng        *rand.Rand
	mu         sync.RWMutex
	conditions map[string]*NetworkConditions
}

// NetworkSimConfig holds network simulation configuration
type NetworkSimConfig struct {
	// Latency simulation
	BaseLatency      time.Duration
	LatencyJitter    time.Duration
	LatencyVariation float64 // 0.0 = no variation, 1.0 = 100% variation

	// Packet loss simulation
	PacketLossRate float64 // 0.0 = no loss, 1.0 = 100% loss
	BurstLossRate  float64 // Burst packet loss probability
	BurstLength    int     // Average burst length

	// Bandwidth simulation
	BandwidthLimit  int64   // bytes per second
	BandwidthJitter float64 // 0.0 = no jitter, 1.0 = 100% jitter

	// Network conditions
	EnableCongestion bool
	CongestionRate   float64 // 0.0 = no congestion, 1.0 = severe congestion

	// Geographic simulation
	EnableGeographic  bool
	GeographicLatency map[string]time.Duration // region -> latency
}

// NetworkConditions represents current network conditions for a node
type NetworkConditions struct {
	NodeID          string
	CurrentLatency  time.Duration
	PacketLossRate  float64
	BandwidthLimit  int64
	CongestionLevel float64
	LastUpdate      time.Time
}

// NewNetworkSimulator creates a new network simulator
func NewNetworkSimulator(config *NetworkSimConfig) *NetworkSimulator {
	if config == nil {
		config = &NetworkSimConfig{
			BaseLatency:      50 * time.Millisecond,
			LatencyJitter:    20 * time.Millisecond,
			LatencyVariation: 0.3,
			PacketLossRate:   0.001, // 0.1% packet loss
			BurstLossRate:    0.01,  // 1% burst loss
			BurstLength:      5,
			BandwidthLimit:   1024 * 1024, // 1MB/s
			BandwidthJitter:  0.2,
			EnableCongestion: true,
			CongestionRate:   0.1,
			EnableGeographic: false,
		}
	}

	return &NetworkSimulator{
		config:     config,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		conditions: make(map[string]*NetworkConditions),
	}
}

// SimulateNetworkLatency simulates realistic network latency
func (ns *NetworkSimulator) SimulateNetworkLatency(fromNode, toNode string) time.Duration {
	ns.mu.RLock()

	// Base latency
	latency := ns.config.BaseLatency

	// Add geographic latency if enabled
	if ns.config.EnableGeographic {
		if geoLatency, exists := ns.config.GeographicLatency[toNode]; exists {
			latency += geoLatency
		}
	}

	// Add random jitter
	jitter := time.Duration(ns.rng.Float64() * float64(ns.config.LatencyJitter))
	latency += jitter

	// Add variation based on current conditions
	fromConditions := ns.getNodeConditions(fromNode)
	toConditions := ns.getNodeConditions(toNode)

	var congestionFactor float64
	if fromConditions != nil && toConditions != nil {
		// Congestion affects latency
		congestionFactor = (fromConditions.CongestionLevel + toConditions.CongestionLevel) / 2
		latency += time.Duration(congestionFactor * float64(ns.config.BaseLatency))
	}

	ns.mu.RUnlock()

	// Update conditions outside of the read lock to avoid deadlock
	if congestionFactor > 0 {
		ns.updateNodeConditions(fromNode, toNode)
	}

	return latency
}

// SimulatePacketLoss simulates realistic packet loss
func (ns *NetworkSimulator) SimulatePacketLoss(fromNode, toNode string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Base packet loss
	if ns.rng.Float64() < ns.config.PacketLossRate {
		return true
	}

	// Burst packet loss
	if ns.rng.Float64() < ns.config.BurstLossRate {
		burstLength := ns.rng.Intn(ns.config.BurstLength*2) + 1
		ns.simulateBurstLoss(fromNode, toNode, burstLength)
		return true
	}

	// Congestion-based packet loss
	fromConditions := ns.getNodeConditions(fromNode)
	if fromConditions != nil && fromConditions.CongestionLevel > 0.8 {
		if ns.rng.Float64() < fromConditions.CongestionLevel {
			return true
		}
	}

	return false
}

// SimulateBandwidthLimit simulates bandwidth constraints
func (ns *NetworkSimulator) SimulateBandwidthLimit(nodeID string, dataSize int64) time.Duration {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	conditions := ns.getNodeConditions(nodeID)
	if conditions == nil {
		conditions = &NetworkConditions{
			NodeID:         nodeID,
			BandwidthLimit: ns.config.BandwidthLimit,
			LastUpdate:     time.Now(),
		}
		ns.conditions[nodeID] = conditions
	}

	// Calculate effective bandwidth
	effectiveBandwidth := conditions.BandwidthLimit

	// Add jitter
	jitter := 1.0 + (ns.rng.Float64()-0.5)*ns.config.BandwidthJitter
	effectiveBandwidth = int64(float64(effectiveBandwidth) * jitter)

	// Congestion reduces bandwidth
	if ns.config.EnableCongestion {
		congestionFactor := 1.0 - conditions.CongestionLevel
		effectiveBandwidth = int64(float64(effectiveBandwidth) * congestionFactor)
	}

	// Calculate transfer time
	transferTime := time.Duration(float64(dataSize) / float64(effectiveBandwidth) * float64(time.Second))

	// Add minimum latency
	if transferTime < ns.config.BaseLatency {
		transferTime = ns.config.BaseLatency
	}

	return transferTime
}

// SimulateNetworkCongestion simulates network congestion
func (ns *NetworkSimulator) SimulateNetworkCongestion(nodeID string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	conditions := ns.getNodeConditions(nodeID)
	if conditions == nil {
		conditions = &NetworkConditions{
			NodeID:     nodeID,
			LastUpdate: time.Now(),
		}
		ns.conditions[nodeID] = conditions
	}

	// Random congestion spikes
	if ns.rng.Float64() < ns.config.CongestionRate {
		conditions.CongestionLevel = ns.rng.Float64() * 0.8 // 0-80% congestion
	} else {
		// Gradual congestion decay
		conditions.CongestionLevel *= 0.95
		if conditions.CongestionLevel < 0.01 {
			conditions.CongestionLevel = 0.0
		}
	}

	conditions.LastUpdate = time.Now()
}

// SimulateNetworkPartition simulates network partitions
func (ns *NetworkSimulator) SimulateNetworkPartition(nodes []string, partitionProbability float64) map[string][]string {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	partitions := make(map[string][]string)

	for _, node := range nodes {
		if ns.rng.Float64() < partitionProbability {
			// Create a new partition
			partitionID := "partition_" + node
			partitions[partitionID] = []string{node}

			// Add nearby nodes to the same partition
			for _, otherNode := range nodes {
				if otherNode != node && ns.rng.Float64() < 0.3 {
					partitions[partitionID] = append(partitions[partitionID], otherNode)
				}
			}
		}
	}

	return partitions
}

// SimulateNetworkRecovery simulates network recovery after issues
func (ns *NetworkSimulator) SimulateNetworkRecovery(nodeID string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	conditions := ns.getNodeConditions(nodeID)
	if conditions != nil {
		// Reset congestion
		conditions.CongestionLevel *= 0.5

		// Restore bandwidth
		conditions.BandwidthLimit = ns.config.BandwidthLimit

		// Reduce packet loss temporarily
		conditions.PacketLossRate = ns.config.PacketLossRate * 0.1

		conditions.LastUpdate = time.Now()
	}
}

// GetNetworkStats returns current network statistics
func (ns *NetworkSimulator) GetNetworkStats() map[string]*NetworkConditions {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	stats := make(map[string]*NetworkConditions)
	for nodeID, conditions := range ns.conditions {
		stats[nodeID] = &NetworkConditions{
			NodeID:          conditions.NodeID,
			CurrentLatency:  conditions.CurrentLatency,
			PacketLossRate:  conditions.PacketLossRate,
			BandwidthLimit:  conditions.BandwidthLimit,
			CongestionLevel: conditions.CongestionLevel,
			LastUpdate:      conditions.LastUpdate,
		}
	}

	return stats
}

// Helper methods
func (ns *NetworkSimulator) getNodeConditions(nodeID string) *NetworkConditions {
	return ns.conditions[nodeID]
}

func (ns *NetworkSimulator) updateCongestionDirectly(nodeID string) {
	// Directly update congestion levels to avoid recursive locking
	ns.mu.Lock()
	defer ns.mu.Unlock()

	conditions := ns.getNodeConditions(nodeID)
	if conditions == nil {
		conditions = &NetworkConditions{
			NodeID:     nodeID,
			LastUpdate: time.Now(),
		}
		ns.conditions[nodeID] = conditions
	}

	// Random congestion spikes
	if ns.rng.Float64() < ns.config.CongestionRate {
		conditions.CongestionLevel = ns.rng.Float64() * 0.8 // 0-80% congestion
	} else {
		// Gradual congestion decay
		conditions.CongestionLevel *= 0.95
		if conditions.CongestionLevel < 0.01 {
			conditions.CongestionLevel = 0.0
		}
	}

	conditions.LastUpdate = time.Now()
}

func (ns *NetworkSimulator) updateNodeConditions(fromNode, toNode string) {
	// Update congestion based on traffic (without calling SimulateNetworkCongestion to avoid deadlock)
	if ns.rng.Float64() < 0.1 { // 10% chance of congestion increase
		// Directly update congestion levels to avoid recursive locking
		ns.updateCongestionDirectly(fromNode)
		ns.updateCongestionDirectly(toNode)
	}
}

func (ns *NetworkSimulator) simulateBurstLoss(fromNode, toNode string, burstLength int) {
	// Simulate burst packet loss by temporarily increasing loss rate
	fromConditions := ns.getNodeConditions(fromNode)
	if fromConditions != nil {
		fromConditions.PacketLossRate = ns.config.PacketLossRate * float64(burstLength)
		fromConditions.LastUpdate = time.Now()
	}
}

// NetworkEvent represents a network event
type NetworkEvent struct {
	Type      string
	NodeID    string
	Timestamp time.Time
	Data      map[string]interface{}
}

// NetworkEventType constants
const (
	EventNodeJoin     = "node_join"
	EventNodeLeave    = "node_leave"
	EventPartition    = "partition"
	EventRecovery     = "recovery"
	EventCongestion   = "congestion"
	EventPacketLoss   = "packet_loss"
	EventLatencySpike = "latency_spike"
)
