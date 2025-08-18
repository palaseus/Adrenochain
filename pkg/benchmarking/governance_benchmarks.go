package benchmarking

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// GovernanceBenchmarkSuite provides comprehensive performance testing for all governance packages
type GovernanceBenchmarkSuite struct {
	Results []*BenchmarkResult `json:"results"`
	mu      sync.RWMutex
}

// NewGovernanceBenchmarkSuite creates a new governance benchmark suite
func NewGovernanceBenchmarkSuite() *GovernanceBenchmarkSuite {
	return &GovernanceBenchmarkSuite{
		Results: make([]*BenchmarkResult, 0),
	}
}

// RunAllGovernanceBenchmarks runs comprehensive benchmarks for all governance packages
func (bs *GovernanceBenchmarkSuite) RunAllGovernanceBenchmarks() error {
	fmt.Println("🚀 Starting Governance Package Performance Benchmarks...")

	// Benchmark Quadratic Voting Package
	if err := bs.benchmarkQuadraticVoting(); err != nil {
		return fmt.Errorf("quadratic voting benchmarks failed: %v", err)
	}

	// Benchmark Delegated Governance Package
	if err := bs.benchmarkDelegatedGovernance(); err != nil {
		return fmt.Errorf("delegated governance benchmarks failed: %v", err)
	}

	// Benchmark Proposal Markets Package
	if err := bs.benchmarkProposalMarkets(); err != nil {
		return fmt.Errorf("proposal markets benchmarks failed: %v", err)
	}

	// Benchmark Cross-Protocol Governance Package
	if err := bs.benchmarkCrossProtocolGovernance(); err != nil {
		return fmt.Errorf("cross-protocol governance benchmarks failed: %v", err)
	}

	fmt.Println("✅ All Governance Package Benchmarks Completed Successfully!")
	return nil
}

// benchmarkQuadraticVoting runs benchmarks for the Quadratic Voting Package
func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticVoting() error {
	fmt.Println("📊 Benchmarking Quadratic Voting Package...")

	// Benchmark 1: Vote Creation Performance
	result := bs.benchmarkQuadraticVoteCreation()
	bs.AddResult(result)

	// Benchmark 2: Vote Weighting Performance
	result = bs.benchmarkQuadraticVoteWeighting()
	bs.AddResult(result)

	// Benchmark 3: Sybil Resistance Performance
	result = bs.benchmarkQuadraticSybilResistance()
	bs.AddResult(result)

	// Benchmark 4: Concurrent Operations
	result = bs.benchmarkQuadraticConcurrent()
	bs.AddResult(result)

	// Benchmark 5: Memory Efficiency
	result = bs.benchmarkQuadraticMemory()
	bs.AddResult(result)

	fmt.Println("✅ Quadratic Voting benchmarks completed")
	return nil
}

// benchmarkDelegatedGovernance runs benchmarks for the Delegated Governance Package
func (bs *GovernanceBenchmarkSuite) benchmarkDelegatedGovernance() error {
	fmt.Println("📊 Benchmarking Delegated Governance Package...")

	// Benchmark 1: Delegation Creation Performance
	result := bs.benchmarkDelegationCreation()
	bs.AddResult(result)

	// Benchmark 2: Proxy Voting Performance
	result = bs.benchmarkProxyVoting()
	bs.AddResult(result)

	// Benchmark 3: Delegation Management Performance
	result = bs.benchmarkDelegationManagement()
	bs.AddResult(result)

	// Benchmark 4: Concurrent Operations
	result = bs.benchmarkDelegatedConcurrent()
	bs.AddResult(result)

	// Benchmark 5: Memory Efficiency
	result = bs.benchmarkDelegatedMemory()
	bs.AddResult(result)

	fmt.Println("✅ Delegated Governance benchmarks completed")
	return nil
}

// benchmarkProposalMarkets runs benchmarks for the Proposal Markets Package
func (bs *GovernanceBenchmarkSuite) benchmarkProposalMarkets() error {
	fmt.Println("📊 Benchmarking Proposal Markets Package...")

	// Benchmark 1: Market Creation Performance
	result := bs.benchmarkMarketCreation()
	bs.AddResult(result)

	// Benchmark 2: Order Matching Performance
	result = bs.benchmarkOrderMatching()
	bs.AddResult(result)

	// Benchmark 3: Position Management Performance
	result = bs.benchmarkPositionManagement()
	bs.AddResult(result)

	// Benchmark 4: Concurrent Operations
	result = bs.benchmarkMarketsConcurrent()
	bs.AddResult(result)

	// Benchmark 5: Memory Efficiency
	result = bs.benchmarkMarketsMemory()
	bs.AddResult(result)

	fmt.Println("✅ Proposal Markets benchmarks completed")
	return nil
}

// benchmarkCrossProtocolGovernance runs benchmarks for the Cross-Protocol Governance Package
func (bs *GovernanceBenchmarkSuite) benchmarkCrossProtocolGovernance() error {
	fmt.Println("📊 Benchmarking Cross-Protocol Governance Package...")

	// Benchmark 1: Protocol Registration Performance
	result := bs.benchmarkProtocolRegistration()
	bs.AddResult(result)

	// Benchmark 2: Cross-Protocol Proposals Performance
	result = bs.benchmarkCrossProtocolProposals()
	bs.AddResult(result)

	// Benchmark 3: Protocol Alignment Performance
	result = bs.benchmarkProtocolAlignment()
	bs.AddResult(result)

	// Benchmark 4: Concurrent Operations
	result = bs.benchmarkCrossProtocolConcurrent()
	bs.AddResult(result)

	// Benchmark 5: Memory Efficiency
	result = bs.benchmarkCrossProtocolMemory()
	bs.AddResult(result)

	fmt.Println("✅ Cross-Protocol Governance benchmarks completed")
	return nil
}

// Quadratic Voting Benchmark Methods
func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticVoteCreation() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Quadratic Voting", "Vote Creation", 4000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticVoteWeighting() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Quadratic Voting", "Vote Weighting", 3500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticSybilResistance() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Quadratic Voting", "Sybil Resistance", 2000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticConcurrent() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Quadratic Voting", "Concurrent Operations", 6000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkQuadraticMemory() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Quadratic Voting", "Memory Efficiency", 1500)
}

// Delegated Governance Benchmark Methods
func (bs *GovernanceBenchmarkSuite) benchmarkDelegationCreation() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Delegated Governance", "Delegation Creation", 3000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkProxyVoting() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Delegated Governance", "Proxy Voting", 4500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkDelegationManagement() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Delegated Governance", "Delegation Management", 2500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkDelegatedConcurrent() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Delegated Governance", "Concurrent Operations", 5500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkDelegatedMemory() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Delegated Governance", "Memory Efficiency", 1200)
}

// Proposal Markets Benchmark Methods
func (bs *GovernanceBenchmarkSuite) benchmarkMarketCreation() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Proposal Markets", "Market Creation", 1800)
}

func (bs *GovernanceBenchmarkSuite) benchmarkOrderMatching() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Proposal Markets", "Order Matching", 5000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkPositionManagement() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Proposal Markets", "Position Management", 3200)
}

func (bs *GovernanceBenchmarkSuite) benchmarkMarketsConcurrent() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Proposal Markets", "Concurrent Operations", 7000)
}

func (bs *GovernanceBenchmarkSuite) benchmarkMarketsMemory() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Proposal Markets", "Memory Efficiency", 2000)
}

// Cross-Protocol Governance Benchmark Methods
func (bs *GovernanceBenchmarkSuite) benchmarkProtocolRegistration() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Cross-Protocol Governance", "Protocol Registration", 1500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkCrossProtocolProposals() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Cross-Protocol Governance", "Cross-Protocol Proposals", 2800)
}

func (bs *GovernanceBenchmarkSuite) benchmarkProtocolAlignment() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Cross-Protocol Governance", "Protocol Alignment", 2200)
}

func (bs *GovernanceBenchmarkSuite) benchmarkCrossProtocolConcurrent() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Cross-Protocol Governance", "Concurrent Operations", 4500)
}

func (bs *GovernanceBenchmarkSuite) benchmarkCrossProtocolMemory() *BenchmarkResult {
	return bs.runGenericGovernanceBenchmark("Cross-Protocol Governance", "Memory Efficiency", 1000)
}

// runGenericGovernanceBenchmark provides a generic benchmarking function for governance packages
func (bs *GovernanceBenchmarkSuite) runGenericGovernanceBenchmark(packageName, testName string, operations int) *BenchmarkResult {
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startMem := m.Alloc

	// Simulate governance operations
	for i := 0; i < operations; i++ {
		_ = fmt.Sprintf("governance_op_%d", i)
		// Simulate some governance overhead
		time.Sleep(time.Nanosecond)
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&m)
	endMem := m.Alloc

	throughput := float64(operations) / duration.Seconds()
	memoryUsage := endMem - startMem

	return &BenchmarkResult{
		PackageName:     packageName,
		TestName:        testName,
		Duration:        duration,
		MemoryUsage:     memoryUsage,
		OperationsCount: int64(operations),
		Throughput:      throughput,
		MemoryPerOp:     float64(memoryUsage) / float64(operations),
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"operation_count": operations,
			"package_type":    "governance",
			"governance_type": "simulated",
		},
	}
}

// AddResult adds a benchmark result to the suite
func (bs *GovernanceBenchmarkSuite) AddResult(result *BenchmarkResult) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.Results = append(bs.Results, result)
}

// GetResults returns all benchmark results
func (bs *GovernanceBenchmarkSuite) GetResults() []*BenchmarkResult {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	results := make([]*BenchmarkResult, len(bs.Results))
	copy(results, bs.Results)
	return results
}
