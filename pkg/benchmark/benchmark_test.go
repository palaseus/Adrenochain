package benchmark

import (
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/stretchr/testify/assert"
)

// MockStorage implements storage.StorageInterface for testing
type MockStorage struct{}

func (m *MockStorage) StoreBlock(b *block.Block) error                 { return nil }
func (m *MockStorage) GetBlock(hash []byte) (*block.Block, error)      { return nil, nil }
func (m *MockStorage) StoreChainState(state *storage.ChainState) error { return nil }
func (m *MockStorage) GetChainState() (*storage.ChainState, error)     { return nil, nil }
func (m *MockStorage) Write(key []byte, value []byte) error            { return nil }
func (m *MockStorage) Read(key []byte) ([]byte, error)                 { return nil, nil }
func (m *MockStorage) Delete(key []byte) error                         { return nil }
func (m *MockStorage) Has(key []byte) (bool, error)                    { return false, nil }
func (m *MockStorage) Close() error                                    { return nil }

// MockChain implements chain.Chain for testing
type MockChain struct{}

func (m *MockChain) GetHeight() uint64                           { return 100 }
func (m *MockChain) GetTipHash() []byte                          { return make([]byte, 32) }
func (m *MockChain) GetBlockByHeight(height uint64) *block.Block { return nil }
func (m *MockChain) GetBlock(hash []byte) *block.Block           { return nil }
func (m *MockChain) AddBlock(block interface{}) error            { return nil }

// Create a mock chain that can be used with the benchmark suite
func createMockChain() *chain.Chain {
	// For testing purposes, we'll create a minimal chain
	// This is a simplified approach - in a real scenario, you'd use the actual chain package
	return &chain.Chain{}
}

func TestNewBenchmarkSuite(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	suite := NewBenchmarkSuite(mockChain, mockStorage)

	assert.NotNil(t, suite)
	assert.Equal(t, mockChain, suite.chain)
	assert.Equal(t, mockStorage, suite.storage)
	assert.NotNil(t, suite.results)
}

func TestDefaultBenchmarkConfig(t *testing.T) {
	config := DefaultBenchmarkConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.Duration)
	assert.Equal(t, 4, config.Concurrency)
	assert.Equal(t, 1000, config.TransactionCount)
	assert.Equal(t, 1024*1024, config.BlockSize)
	assert.Equal(t, 100*time.Millisecond, config.NetworkLatency)
	assert.True(t, config.EnableProfiling)
	assert.True(t, config.EnableMemoryStats)
	assert.True(t, config.EnableCPUStats)
}

func TestBenchmarkSuite_RunAllBenchmarks(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	suite := NewBenchmarkSuite(mockChain, mockStorage)
	config := DefaultBenchmarkConfig()

	// Run a shorter benchmark for testing
	config.Duration = 100 * time.Millisecond
	config.TransactionCount = 50

	results := suite.RunAllBenchmarks(config)

	// Verify all benchmarks ran
	expectedBenchmarks := []string{
		"TransactionThroughput",
		"BlockPropagation",
		"StoragePerformance",
		"ChainValidation",
		"ConcurrentOperations",
		"MemoryEfficiency",
		"NetworkLatency",
	}

	for _, name := range expectedBenchmarks {
		result, exists := results[name]
		assert.True(t, exists, "Benchmark %s should exist", name)
		assert.NotNil(t, result, "Benchmark result %s should not be nil", name)
		assert.Equal(t, name, result.Name)
		assert.True(t, result.Duration > 0, "Duration should be positive")
		assert.True(t, result.Operations > 0, "Operations should be positive")
		assert.True(t, result.Throughput > 0, "Throughput should be positive")
		assert.True(t, result.SuccessRate >= 0, "Success rate should be non-negative")
	}
}

func TestBenchmarkSuite_TransactionThroughput(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	suite := NewBenchmarkSuite(mockChain, mockStorage)
	config := DefaultBenchmarkConfig()
	config.TransactionCount = 50
	config.Concurrency = 2

	result := suite.BenchmarkTransactionThroughput(config)

	assert.NotNil(t, result)
	assert.Equal(t, "TransactionThroughput", result.Name)
	assert.True(t, result.Duration > 0)
	assert.Equal(t, int64(50), result.Operations)
	assert.True(t, result.Throughput > 0)
	assert.Equal(t, int64(0), result.ErrorCount)
	assert.Equal(t, 100.0, result.SuccessRate)
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, 2, result.Metadata["concurrency"])
	assert.Equal(t, 50, result.Metadata["transaction_count"])
}

func TestBenchmarkSuite_ReportGeneration(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	suite := NewBenchmarkSuite(mockChain, mockStorage)
	config := DefaultBenchmarkConfig()
	config.TransactionCount = 100

	// Run benchmarks to generate results
	suite.RunAllBenchmarks(config)

	// Generate report
	report := suite.GenerateReport()

	// Verify report content
	assert.Contains(t, report, "# 🚀 GoChain Benchmark Report")
	assert.Contains(t, report, "## 📊 Summary")
	assert.Contains(t, report, "## 🎯 Overall Performance")
	assert.Contains(t, report, "TransactionThroughput")
	assert.Contains(t, report, "BlockPropagation")
	assert.Contains(t, report, "StoragePerformance")
	assert.Contains(t, report, "ChainValidation")
	assert.Contains(t, report, "ConcurrentOperations")
	assert.Contains(t, report, "MemoryEfficiency")
	assert.Contains(t, report, "NetworkLatency")
}

func TestBenchmarkSuite_ResultsManagement(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	suite := NewBenchmarkSuite(mockChain, mockStorage)

	// Initially no results
	results := suite.GetResults()
	assert.Len(t, results, 0)

	// Run benchmarks
	config := DefaultBenchmarkConfig()
	config.TransactionCount = 25
	config.Duration = 50 * time.Millisecond

	suite.RunAllBenchmarks(config)

	// Now should have results
	results = suite.GetResults()
	assert.Len(t, results, 7) // All 7 benchmark types
}
