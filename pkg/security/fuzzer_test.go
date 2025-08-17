package security

import (
	"fmt"
	"math/rand"
	"sync"
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

// Create a mock chain that can be used with the fuzzer
func createMockChain() *chain.Chain {
	// For testing purposes, we'll create a minimal chain
	// This is a simplified approach - in a real scenario, you'd use the actual chain package
	return &chain.Chain{}
}

func TestNewFuzzer(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	assert.NotNil(t, fuzzer)
	assert.Equal(t, mockChain, fuzzer.chain)
	assert.Equal(t, mockStorage, fuzzer.storage)
	assert.NotNil(t, fuzzer.results)
	assert.NotNil(t, fuzzer.config)
	assert.False(t, fuzzer.isRunning)
}

func TestDefaultFuzzConfig(t *testing.T) {
	config := DefaultFuzzConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 60*time.Second, config.Duration)
	assert.Equal(t, int64(10000), config.MaxIterations)
	assert.Equal(t, 100*time.Millisecond, config.Timeout)
	assert.Equal(t, 1024, config.MaxInputSize)
	assert.Equal(t, 1, config.MinInputSize)
	assert.True(t, config.EnableMutation)
	assert.True(t, config.EnableCoverage)
	assert.True(t, config.EnableCrashReport)
	assert.True(t, config.Seed > 0)
	assert.Equal(t, 4, config.Concurrency)
	assert.Empty(t, config.TargetFunctions)
	assert.Empty(t, config.ExcludeFunctions)
}

func TestFuzzer_StartStop(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Initially not running
	assert.False(t, fuzzer.IsRunning())

	// Start fuzzing
	config := DefaultFuzzConfig()
	config.Duration = 100 * time.Millisecond // Short duration for testing
	config.MaxIterations = 100               // Low iteration count for testing

	err := fuzzer.StartFuzzing(config)
	assert.NoError(t, err)
	assert.True(t, fuzzer.IsRunning())

	// Wait a bit for fuzzing to start
	time.Sleep(50 * time.Millisecond)

	// Stop fuzzing
	fuzzer.StopFuzzing()
	assert.False(t, fuzzer.IsRunning())
}

func TestFuzzer_GenerateFuzzInput(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test input generation
	input := fuzzer.generateFuzzInput(rand.New(rand.NewSource(1)))

	assert.NotNil(t, input)
	assert.GreaterOrEqual(t, len(input), fuzzer.config.MinInputSize)
	assert.LessOrEqual(t, len(input), fuzzer.config.MaxInputSize)
}

func TestFuzzer_MutateInput(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test input mutation
	original := []byte{1, 2, 3, 4, 5}
	mutated := fuzzer.mutateInput(original, rand.New(rand.NewSource(1)))

	assert.NotNil(t, mutated)
	// Mutation should change the input in some way
	assert.NotEqual(t, original, mutated)
}

func TestFuzzer_ExecuteFuzzTest(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test fuzz test execution
	input := []byte{1, 2, 3, 4, 5}
	result := fuzzer.executeFuzzTest(input, 1, 1)

	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Status)
	assert.Equal(t, input, result.Input)
	assert.True(t, result.Duration > 0)
}

func TestFuzzer_TestBlockParsing(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test block parsing with valid input
	validInput := make([]byte, 64)
	for i := range validInput {
		validInput[i] = byte(i % 256)
	}

	// Should not panic
	assert.NotPanics(t, func() {
		fuzzer.testBlockParsing(validInput)
	})
}

func TestFuzzer_TestTransactionParsing(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test transaction parsing with valid input
	validInput := make([]byte, 64)
	for i := range validInput {
		validInput[i] = byte(i % 256)
	}

	// Should not panic
	assert.NotPanics(t, func() {
		fuzzer.testTransactionParsing(validInput)
	})
}

func TestFuzzer_TestStorageOperations(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test storage operations with valid input
	validInput := make([]byte, 64)
	for i := range validInput {
		validInput[i] = byte(i % 256)
	}

	// Should not panic
	assert.NotPanics(t, func() {
		fuzzer.testStorageOperations(validInput)
	})
}

func TestFuzzer_TestChainOperations(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test chain operations with valid input
	validInput := make([]byte, 64)
	for i := range validInput {
		validInput[i] = byte(i % 256)
	}

	// Should not panic
	assert.NotPanics(t, func() {
		fuzzer.testChainOperations(validInput)
	})
}

func TestFuzzer_GetResults(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Initially no results
	results := fuzzer.GetResults()
	assert.Len(t, results, 0)

	// Add some mock results
	fuzzer.mu.Lock()
	fuzzer.results["test"] = &FuzzResult{
		Name:       "test",
		Iterations: 100,
		Timestamp:  time.Now(),
	}
	fuzzer.mu.Unlock()

	// Now should have results
	results = fuzzer.GetResults()
	assert.Len(t, results, 1)
	assert.Contains(t, results, "test")
}

func TestFuzzer_GenerateReport(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Add some mock results
	fuzzer.mu.Lock()
	fuzzer.results["test"] = &FuzzResult{
		Name:         "test",
		Iterations:   100,
		CrashCount:   5,
		TimeoutCount: 2,
		ErrorCount:   3,
		SuccessCount: 90,
		Timestamp:    time.Now(),
	}
	fuzzer.mu.Unlock()

	// Generate report
	report := fuzzer.GenerateReport()

	// Verify report content
	assert.Contains(t, report, "# 🔒 GoChain Fuzz Test Report")
	assert.Contains(t, report, "## 📊 Summary")
	assert.Contains(t, report, "## 🎯 Overall Results")
	assert.Contains(t, report, "test")
	assert.Contains(t, report, "100")
	assert.Contains(t, report, "5")
}

func TestFuzzer_EdgeCases(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	t.Run("EmptyInput", func(t *testing.T) {
		// Test with empty input
		assert.NotPanics(t, func() {
			fuzzer.testBlockParsing([]byte{})
			fuzzer.testTransactionParsing([]byte{})
			fuzzer.testStorageOperations([]byte{})
			fuzzer.testChainOperations([]byte{})
		})
	})

	t.Run("VeryLargeInput", func(t *testing.T) {
		// Test with very large input
		largeInput := make([]byte, 10000)
		assert.NotPanics(t, func() {
			fuzzer.testBlockParsing(largeInput)
			fuzzer.testTransactionParsing(largeInput)
			fuzzer.testStorageOperations(largeInput)
			fuzzer.testChainOperations(largeInput)
		})
	})

	t.Run("NilInput", func(t *testing.T) {
		// Test with nil input
		assert.NotPanics(t, func() {
			fuzzer.testBlockParsing(nil)
			fuzzer.testTransactionParsing(nil)
			fuzzer.testStorageOperations(nil)
			fuzzer.testChainOperations(nil)
		})
	})
}

func TestFuzzer_Concurrency(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test concurrent access to results
	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Read results
			_ = fuzzer.GetResults()

			// Write results
			fuzzer.mu.Lock()
			fuzzer.results[fmt.Sprintf("goroutine_%d", id)] = &FuzzResult{
				Name:       fmt.Sprintf("goroutine_%d", id),
				Iterations: int64(id),
				Timestamp:  time.Now(),
			}
			fuzzer.mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all results were written
	results := fuzzer.GetResults()
	assert.Len(t, results, numGoroutines)
}

func TestFuzzer_Configuration(t *testing.T) {
	mockChain := createMockChain()
	mockStorage := &MockStorage{}

	fuzzer := NewFuzzer(mockChain, mockStorage)

	// Test custom configuration
	customConfig := &FuzzConfig{
		Duration:       30 * time.Second,
		MaxIterations:  5000,
		Timeout:        50 * time.Millisecond,
		MaxInputSize:   512,
		MinInputSize:   10,
		EnableMutation: false,
		Concurrency:    2,
	}

	err := fuzzer.StartFuzzing(customConfig)
	assert.NoError(t, err)

	// Verify configuration was applied
	assert.Equal(t, customConfig, fuzzer.config)

	// Stop fuzzing
	fuzzer.StopFuzzing()
}
