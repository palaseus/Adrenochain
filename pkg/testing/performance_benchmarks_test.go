package testing

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceBenchmarks_GetResults(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Initially should be empty
	results := pb.GetResults()
	require.NotNil(t, results)
	assert.Len(t, results, 0)

	// Run some benchmarks to populate results
	benchmarkResults := pb.RunAllBenchmarks()
	require.NotNil(t, benchmarkResults)

	// Should now have results
	results = pb.GetResults()
	require.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check that results contain expected components
	foundComponents := make(map[string]bool)
	for _, result := range results {
		assert.NotEmpty(t, result.Component)
		assert.NotEmpty(t, result.Operation)
		assert.True(t, result.Duration > 0)
		foundComponents[result.Component] = true
	}

	// Should have multiple different components
	assert.True(t, len(foundComponents) > 1)
}

func TestPerformanceBenchmarks_PrintResults(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Run benchmarks first
	benchmarkResults := pb.RunAllBenchmarks()
	require.NotNil(t, benchmarkResults)

	// This should not panic and should execute successfully
	// Since PrintResults outputs to stdout, we can't easily capture the output
	// but we can ensure it doesn't crash
	assert.NotPanics(t, func() {
		pb.PrintResults()
	})
}

func TestPerformanceBenchmarks_RunPerformanceBenchmarks(t *testing.T) {
	// Test the standalone function
	assert.NotPanics(t, func() {
		RunPerformanceBenchmarks()
	})
}

func TestPerformanceBenchmarks_EdgeCases(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Test printing results with no benchmarks run
	assert.NotPanics(t, func() {
		pb.PrintResults()
	})

	// Test getting results with no benchmarks run
	results := pb.GetResults()
	require.NotNil(t, results)
	assert.Len(t, results, 0)
}

func TestPerformanceBenchmarks_ResultsStructure(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Run benchmarks
	benchmarkResults := pb.RunAllBenchmarks()
	require.NotNil(t, benchmarkResults)

	results := pb.GetResults()
	require.NotNil(t, results)
	require.Greater(t, len(results), 0)

	// Verify each result has the required fields
	for name, result := range results {
		assert.NotEmpty(t, name)
		assert.NotEmpty(t, result.Component)
		assert.NotEmpty(t, result.Operation)
		assert.True(t, result.Duration > 0)
		assert.True(t, result.OperationsPerSecond >= 0)
		assert.True(t, result.Throughput >= 0)
		assert.True(t, result.Latency >= 0)
		assert.True(t, result.ErrorRate >= 0)
		assert.True(t, result.ErrorRate <= 100) // Error rate should be a percentage
	}
}

func TestPerformanceBenchmarks_ConcurrentAccess(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Run benchmarks to populate results
	benchmarkResults := pb.RunAllBenchmarks()
	require.NotNil(t, benchmarkResults)

	// Test concurrent access to GetResults
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			results := pb.GetResults()
			assert.NotNil(t, results)
			assert.Greater(t, len(results), 0)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestPerformanceBenchmarks_MultipleRuns(t *testing.T) {
	pb := NewPerformanceBenchmarks()

	// Run benchmarks multiple times
	for i := 0; i < 3; i++ {
		benchmarkResults := pb.RunAllBenchmarks()
		require.NotNil(t, benchmarkResults)

		results := pb.GetResults()
		require.NotNil(t, results)
		assert.Greater(t, len(results), 0)
	}

	// Final results should be valid
	results := pb.GetResults()
	require.NotNil(t, results)
	assert.Greater(t, len(results), 0)
}

func TestCreateTestOrder(t *testing.T) {
	// Test the helper function used by benchmarks
	quantity := big.NewInt(1000000)
	price := big.NewInt(50000)

	order := createTestOrder("test-id", "buy", "limit", quantity, price)
	require.NotNil(t, order)

	assert.Equal(t, "test-id", order.ID)
	assert.Equal(t, "BTC/USDT", order.TradingPair)
	assert.Equal(t, quantity, order.Quantity)
	assert.Equal(t, price, order.Price)
}
