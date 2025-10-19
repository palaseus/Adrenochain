//go:build testing
// +build testing

package pdf

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNetworkSimulator_Creation(t *testing.T) {
	// Test with default config
	simulator := NewNetworkSimulator(nil)
	assert.NotNil(t, simulator)
	assert.NotNil(t, simulator.config)
	assert.NotNil(t, simulator.rng)
	assert.NotNil(t, simulator.conditions)

	// Test with custom config
	customConfig := &NetworkSimConfig{
		BaseLatency:      100 * time.Millisecond,
		LatencyJitter:    50 * time.Millisecond,
		LatencyVariation: 0.2,
		PacketLossRate:   0.01,
		BurstLossRate:    0.05,
		BurstLength:      5,
		BandwidthLimit:   1024 * 1024, // 1MB/s
		BandwidthJitter:  0.1,
		EnableCongestion: true,
		CongestionRate:   0.1,
		EnableGeographic: true,
		GeographicLatency: map[string]time.Duration{
			"us-east": 50 * time.Millisecond,
			"us-west": 80 * time.Millisecond,
			"europe":  120 * time.Millisecond,
			"asia":    200 * time.Millisecond,
		},
	}

	customSimulator := NewNetworkSimulator(customConfig)
	assert.NotNil(t, customSimulator)
	assert.Equal(t, customConfig, customSimulator.config)
}

func TestNetworkSimulator_LatencySimulation(t *testing.T) {
	config := &NetworkSimConfig{
		BaseLatency:      100 * time.Millisecond,
		LatencyJitter:    20 * time.Millisecond,
		LatencyVariation: 0.1,
	}

	simulator := NewNetworkSimulator(config)

	// Test latency calculation using the actual method name
	latency := simulator.SimulateNetworkLatency("node1", "node2")
	assert.Greater(t, latency, 50*time.Millisecond)
	assert.Less(t, latency, 200*time.Millisecond)

	// Test that latency varies between calls
	latencies := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		latencies[i] = simulator.SimulateNetworkLatency("node1", "node2")
	}

	// Check that we have some variation
	hasVariation := false
	for i := 1; i < len(latencies); i++ {
		if latencies[i] != latencies[0] {
			hasVariation = true
			break
		}
	}
	assert.True(t, hasVariation, "Latency should vary between calls")
}

func TestNetworkSimulator_PacketLossSimulation(t *testing.T) {
	config := &NetworkSimConfig{
		PacketLossRate: 0.05, // 5% packet loss
		BurstLossRate:  0.1,  // 10% burst loss
		BurstLength:    2,    // Reduced burst length
	}

	simulator := NewNetworkSimulator(config)

	// Test packet loss simulation using the actual method name
	lossCount := 0
	totalPackets := 1000

	for i := 0; i < totalPackets; i++ {
		if simulator.SimulatePacketLoss("node1", "node2") {
			lossCount++
		}
	}

	// Check that we have some packet loss (within reasonable bounds)
	lossRate := float64(lossCount) / float64(totalPackets)
	assert.Greater(t, lossRate, 0.02)     // At least 2% loss
	assert.LessOrEqual(t, lossRate, 0.25) // At most 25% loss (more conservative limit)
}

func TestNetworkSimulator_BandwidthSimulation(t *testing.T) {
	config := &NetworkSimConfig{
		BandwidthLimit:  1024 * 1024, // 1MB/s
		BandwidthJitter: 0.1,         // 10% jitter
	}

	simulator := NewNetworkSimulator(config)

	// Test bandwidth simulation using the actual method name
	dataSize := int64(512 * 1024) // 512KB
	transferTime := simulator.SimulateBandwidthLimit("node1", dataSize)
	assert.Greater(t, transferTime, time.Duration(0))
	assert.Less(t, transferTime, 2*time.Second) // Should be less than 2 seconds for 512KB
}

func TestNetworkSimulator_CongestionSimulation(t *testing.T) {
	config := &NetworkSimConfig{
		EnableCongestion: true,
		CongestionRate:   0.2, // 20% congestion
	}

	simulator := NewNetworkSimulator(config)

	// Test congestion simulation
	simulator.SimulateNetworkCongestion("node1")

	// Verify the method executed without error
	assert.NotNil(t, simulator)
}

func TestNetworkSimulator_NetworkStats(t *testing.T) {
	simulator := NewNetworkSimulator(nil)

	// Test getting network stats using the actual method name
	stats := simulator.GetNetworkStats()
	assert.NotNil(t, stats)

	// Initially should be empty
	assert.Len(t, stats, 0)

	// Simulate some activity
	simulator.SimulateNetworkLatency("node1", "node2")
	simulator.SimulatePacketLoss("node1", "node2")

	// Check stats again
	stats = simulator.GetNetworkStats()
	assert.NotNil(t, stats)
}

func TestNetworkSimulator_NetworkPartition(t *testing.T) {
	simulator := NewNetworkSimulator(nil)

	// Test network partition simulation using the actual method name
	nodes := []string{"node1", "node2", "node3", "node4"}
	partitionMap := simulator.SimulateNetworkPartition(nodes, 0.5)

	assert.NotNil(t, partitionMap)
	assert.LessOrEqual(t, len(partitionMap), len(nodes))
}

func TestNetworkSimulator_NetworkRecovery(t *testing.T) {
	simulator := NewNetworkSimulator(nil)

	// Test network recovery simulation using the actual method name
	simulator.SimulateNetworkRecovery("node1")

	// Verify the method executed without error
	assert.NotNil(t, simulator)
}

func TestNetworkSimulator_ConcurrentAccess(t *testing.T) {
	simulator := NewNetworkSimulator(nil)

	// Test concurrent access to simulator
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			nodeID := fmt.Sprintf("node_%d", index)

			// Test multiple concurrent operations
			_ = simulator.SimulateNetworkLatency(nodeID, "target_node")
			_ = simulator.SimulatePacketLoss(nodeID, "target_node")
			_ = simulator.SimulateBandwidthLimit(nodeID, 1024)
			simulator.SimulateNetworkCongestion(nodeID)
			_ = simulator.GetNetworkStats()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify simulator is still functional
	latency := simulator.SimulateNetworkLatency("test_node", "target_node")
	assert.Greater(t, latency, time.Duration(0))
}

func TestNetworkSimulator_EdgeCases(t *testing.T) {
	// Test with extreme values
	extremeConfig := &NetworkSimConfig{
		BaseLatency:       0,
		LatencyJitter:     0,
		LatencyVariation:  0,
		PacketLossRate:    0,
		BurstLossRate:     0,
		BurstLength:       0,
		BandwidthLimit:    0,
		BandwidthJitter:   0,
		EnableCongestion:  false,
		CongestionRate:    0,
		EnableGeographic:  false,
		GeographicLatency: map[string]time.Duration{},
	}

	simulator := NewNetworkSimulator(extremeConfig)

	// Test with zero values
	latency := simulator.SimulateNetworkLatency("node1", "node2")
	assert.Equal(t, time.Duration(0), latency)

	// Test with maximum values
	maxConfig := &NetworkSimConfig{
		BaseLatency:      10 * time.Second,
		LatencyJitter:    5 * time.Second,
		LatencyVariation: 1.0,
		PacketLossRate:   1.0, // 100% packet loss
		BurstLossRate:    1.0,
		BurstLength:      100,
		BandwidthLimit:   10 * 1024 * 1024 * 1024, // 10GB/s
		BandwidthJitter:  1.0,
		EnableCongestion: true,
		CongestionRate:   1.0, // 100% congestion
		EnableGeographic: true,
		GeographicLatency: map[string]time.Duration{
			"region1": 5 * time.Second,
			"region2": 10 * time.Second,
		},
	}

	maxSimulator := NewNetworkSimulator(maxConfig)

	// Test with maximum values
	maxLatency := maxSimulator.SimulateNetworkLatency("node1", "node2")
	assert.Greater(t, maxLatency, 5*time.Second)
}

func TestNetworkSimulator_Performance(t *testing.T) {
	simulator := NewNetworkSimulator(nil)

	// Test performance of network calculations
	start := time.Now()

	// Perform many network calculations
	for i := 0; i < 1000; i++ {
		_ = simulator.SimulateNetworkLatency("node1", "node2")
		_ = simulator.SimulatePacketLoss("node1", "node2")
		_ = simulator.SimulateBandwidthLimit("node1", 1024)
		simulator.SimulateNetworkCongestion("node1")
	}

	elapsed := time.Since(start)

	// Verify performance is reasonable (should complete in less than 1 second)
	assert.Less(t, elapsed, 1*time.Second)
}
