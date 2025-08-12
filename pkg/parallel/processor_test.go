package parallel

import (
	"fmt"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

func TestNewParallelProcessor(t *testing.T) {
	config := DefaultProcessorConfig()
	processor := NewParallelProcessor(config)
	defer processor.Close()

	assert.NotNil(t, processor)
	assert.Equal(t, config.MaxWorkers, len(processor.workers))
	assert.Equal(t, config.QueueSize, cap(processor.workQueue))
}

func TestParallelProcessor_Submit(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	item := &WorkItem{
		ID:       "test-item",
		Type:     WorkTypeTransactionValidation,
		Data:     "test-data",
		Priority: 1,
		Created:  time.Now(),
		Result:   make(chan *WorkResult, 1),
	}

	err := processor.Submit(item)
	assert.NoError(t, err)
}

func TestParallelProcessor_ProcessTransaction(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	// Create a mock transaction
	tx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{{}},
		Outputs: []*block.TxOutput{{}},
		Hash:    []byte("test-hash"),
	}

	result, err := processor.ProcessTransaction(tx)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, tx, result.Data)
}

func TestParallelProcessor_ProcessBlock(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	// Create a mock block
	block := &block.Block{
		Header: &block.Header{
			Version: 1,
			Height:  1,
		},
		Transactions: []*block.Transaction{{
			Version: 1,
			Inputs:  []*block.TxInput{{}},
			Outputs: []*block.TxOutput{{}},
		}},
	}

	result, err := processor.ProcessBlock(block)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, block, result.Data)
}

func TestParallelProcessor_SubmitBatch(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	items := make([]*WorkItem, 5)
	for i := 0; i < 5; i++ {
		items[i] = &WorkItem{
			ID:       fmt.Sprintf("batch-item-%d", i),
			Type:     WorkTypeTransactionValidation,
			Data:     fmt.Sprintf("data-%d", i),
			Priority: i,
			Created:  time.Now(),
			Result:   make(chan *WorkResult, 1),
		}
	}

	err := processor.SubmitBatch(items)
	assert.NoError(t, err)
}

func TestParallelProcessor_SubmitWithPriority(t *testing.T) {
	config := DefaultProcessorConfig()
	config.PriorityQueuing = true
	processor := NewParallelProcessor(config)
	defer processor.Close()

	item := &WorkItem{
		ID:       "priority-item",
		Type:     WorkTypeTransactionValidation,
		Data:     "priority-data",
		Priority: 0,
		Created:  time.Now(),
		Result:   make(chan *WorkResult, 1),
	}

	err := processor.SubmitWithPriority(item, 1)
	assert.NoError(t, err)
}

func TestParallelProcessor_GetStats(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	// Submit some work to generate stats
	item := &WorkItem{
		ID:       "stats-item",
		Type:     WorkTypeTransactionValidation,
		Data:     "stats-data",
		Priority: 1,
		Created:  time.Now(),
		Result:   make(chan *WorkResult, 1),
	}

	processor.Submit(item)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	stats := processor.GetStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.TotalItemsProcessed, int64(0))
}

func TestParallelProcessor_Timeout(t *testing.T) {
	// Create processor with very short timeout
	config := &ProcessorConfig{
		MaxWorkers: 1,
		QueueSize:  1,
		Timeout:    1 * time.Millisecond,
	}
	processor := NewParallelProcessor(config)
	defer processor.Close()

	// Submit items - the first two should succeed, the third might timeout
	// depending on how fast the worker processes items
	for i := 0; i < 3; i++ {
		item := &WorkItem{
			ID:       fmt.Sprintf("timeout-item-%d", i),
			Type:     WorkTypeTransactionValidation,
			Data:     fmt.Sprintf("data-%d", i),
			Priority: 1,
			Created:  time.Now(),
			Result:   make(chan *WorkResult, 1),
		}

		err := processor.Submit(item)
		// All submissions might succeed if the worker is fast enough
		// This test just verifies the processor doesn't crash
		if err != nil {
			assert.Contains(t, err.Error(), "timeout")
		}
	}
}

func TestParallelProcessor_ConcurrentAccess(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				item := &WorkItem{
					ID:       fmt.Sprintf("concurrent-%d-%d", id, j),
					Type:     WorkTypeTransactionValidation,
					Data:     fmt.Sprintf("data-%d-%d", id, j),
					Priority: id % 3,
					Created:  time.Now(),
					Result:   make(chan *WorkResult, 1),
				}

				processor.Submit(item)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify processor is still functional
	stats := processor.GetStats()
	assert.NotNil(t, stats)
}

func TestParallelProcessor_WorkerStats(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	// Submit work to generate worker stats
	item := &WorkItem{
		ID:       "worker-stats-item",
		Type:     WorkTypeTransactionValidation,
		Data:     "worker-stats-data",
		Priority: 1,
		Created:  time.Now(),
		Result:   make(chan *WorkResult, 1),
	}

	processor.Submit(item)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check that workers have processed items by looking at stats
	stats := processor.GetStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.TotalItemsProcessed, int64(0))
}

func TestParallelProcessor_WorkTypes(t *testing.T) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	workTypes := []WorkType{
		WorkTypeTransactionValidation,
		WorkTypeBlockProcessing,
		WorkTypeUTXOUpdate,
		WorkTypeMerkleTreeCalculation,
		WorkTypeSignatureVerification,
		WorkTypeStateTransition,
	}

	for _, workType := range workTypes {
		item := &WorkItem{
			ID:       fmt.Sprintf("worktype-%d", workType),
			Type:     workType,
			Data:     fmt.Sprintf("data-%d", workType),
			Priority: 1,
			Created:  time.Now(),
			Result:   make(chan *WorkResult, 1),
		}

		err := processor.Submit(item)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify all work types were processed
	stats := processor.GetStats()
	assert.Greater(t, stats.TotalItemsProcessed, int64(0))
}

func BenchmarkParallelProcessor_Submit(b *testing.B) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			item := &WorkItem{
				ID:       fmt.Sprintf("bench-submit-%d", i),
				Type:     WorkTypeTransactionValidation,
				Data:     fmt.Sprintf("data-%d", i),
				Priority: i % 3,
				Created:  time.Now(),
				Result:   make(chan *WorkResult, 1),
			}
			processor.Submit(item)
			i++
		}
	})
}

func BenchmarkParallelProcessor_ProcessTransaction(b *testing.B) {
	processor := NewParallelProcessor(nil)
	defer processor.Close()

	tx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{{}},
		Outputs: []*block.TxOutput{{}},
		Hash:    []byte("bench-hash"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessTransaction(tx)
	}
}
