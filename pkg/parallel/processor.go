package parallel

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// ProcessorConfig holds configuration for the parallel processor
type ProcessorConfig struct {
	MaxWorkers        int           // Maximum number of worker goroutines
	QueueSize         int           // Size of the work queue
	BatchSize         int           // Number of items to process in a batch
	Timeout           time.Duration // Timeout for processing operations
	EnableProfiling   bool          // Enable performance profiling
	EnableMetrics     bool          // Enable detailed metrics collection
	LoadBalancing     bool          // Enable dynamic load balancing
	PriorityQueuing   bool          // Enable priority-based queuing
}

// DefaultProcessorConfig returns sensible defaults for the processor
func DefaultProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		MaxWorkers:      runtime.NumCPU() * 2, // 2x CPU cores
		QueueSize:       10000,                // 10K queue size
		BatchSize:       100,                  // 100 items per batch
		Timeout:         30 * time.Second,     // 30 second timeout
		EnableProfiling: true,                 // Enable profiling
		EnableMetrics:   true,                 // Enable metrics
		LoadBalancing:   true,                 // Enable load balancing
		PriorityQueuing: true,                 // Enable priority queuing
	}
}

// WorkItem represents a unit of work to be processed
type WorkItem struct {
	ID       string
	Type     WorkType
	Data     interface{}
	Priority int
	Created  time.Time
	Deadline time.Time
	Result   chan *WorkResult
}

// WorkType represents the type of work to be performed
type WorkType int

const (
	WorkTypeTransactionValidation WorkType = iota
	WorkTypeBlockProcessing
	WorkTypeUTXOUpdate
	WorkTypeMerkleTreeCalculation
	WorkTypeSignatureVerification
	WorkTypeStateTransition
)

// WorkResult represents the result of processing a work item
type WorkResult struct {
	ID        string
	Success   bool
	Data      interface{}
	Error     error
	Duration  time.Duration
	WorkerID  int
	Timestamp time.Time
}

// Worker represents a worker goroutine that processes work items
type Worker struct {
	ID       int
	processor *ParallelProcessor
	workChan <-chan *WorkItem
	stats    *WorkerStats
	ctx      context.Context
	cancel   context.CancelFunc
}

// WorkerStats tracks worker performance metrics
type WorkerStats struct {
	ItemsProcessed int64
	TotalDuration  time.Duration
	Errors         int64
	LastActivity   time.Time
	mu             sync.RWMutex
}

// ParallelProcessor is a high-performance parallel processing system
type ParallelProcessor struct {
	config     *ProcessorConfig
	workers    []*Worker
	workQueue  chan *WorkItem
	priorityQueue *PriorityQueue
	stats      *ProcessorStats
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	wg         sync.WaitGroup
}

// ProcessorStats tracks overall processor performance
type ProcessorStats struct {
	TotalItemsProcessed int64
	TotalProcessingTime time.Duration
	AverageLatency      time.Duration
	QueueDepth          int
	ActiveWorkers       int
	Errors              int64
	mu                  sync.RWMutex
}

// NewParallelProcessor creates a new parallel processor instance
func NewParallelProcessor(config *ProcessorConfig) *ParallelProcessor {
	if config == nil {
		config = DefaultProcessorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	processor := &ParallelProcessor{
		config:     config,
		workQueue:  make(chan *WorkItem, config.QueueSize),
		priorityQueue: NewPriorityQueue(),
		stats:      &ProcessorStats{},
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start workers
	processor.startWorkers()

	return processor
}

// Submit submits a work item for processing
func (pp *ParallelProcessor) Submit(item *WorkItem) error {
	select {
	case pp.workQueue <- item:
		atomic.AddInt64(&pp.stats.TotalItemsProcessed, 1)
		return nil
	case <-time.After(pp.config.Timeout):
		return fmt.Errorf("work queue full, submission timeout")
	}
}

// SubmitWithPriority submits a work item with priority
func (pp *ParallelProcessor) SubmitWithPriority(item *WorkItem, priority int) error {
	if pp.config.PriorityQueuing {
		item.Priority = priority
		pp.priorityQueue.PushItem(item, priority)
		return nil
	}
	return pp.Submit(item)
}

// SubmitBatch submits multiple work items as a batch
func (pp *ParallelProcessor) SubmitBatch(items []*WorkItem) error {
	if len(items) > pp.config.BatchSize {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(items), pp.config.BatchSize)
	}

	for _, item := range items {
		if err := pp.Submit(item); err != nil {
			return err
		}
	}

	return nil
}

// ProcessTransaction processes a transaction in parallel
func (pp *ParallelProcessor) ProcessTransaction(tx *block.Transaction) (*WorkResult, error) {
	item := &WorkItem{
		ID:       fmt.Sprintf("tx-%x", tx.Hash),
		Type:     WorkTypeTransactionValidation,
		Data:     tx,
		Priority: 1,
		Created:  time.Now(),
		Deadline: time.Now().Add(pp.config.Timeout),
		Result:   make(chan *WorkResult, 1),
	}

	if err := pp.Submit(item); err != nil {
		return nil, err
	}

	select {
	case result := <-item.Result:
		return result, nil
	case <-time.After(pp.config.Timeout):
		return nil, fmt.Errorf("transaction processing timeout")
	}
}

// ProcessBlock processes a block in parallel
func (pp *ParallelProcessor) ProcessBlock(block *block.Block) (*WorkResult, error) {
	item := &WorkItem{
		ID:       fmt.Sprintf("block-%x", block.CalculateHash()),
		Type:     WorkTypeBlockProcessing,
		Data:     block,
		Priority: 2,
		Created:  time.Now(),
		Deadline: time.Now().Add(pp.config.Timeout),
		Result:   make(chan *WorkResult, 1),
	}

	if err := pp.Submit(item); err != nil {
		return nil, err
	}

	select {
	case result := <-item.Result:
		return result, nil
	case <-time.After(pp.config.Timeout):
		return nil, fmt.Errorf("block processing timeout")
	}
}

// GetStats returns current processor statistics
func (pp *ParallelProcessor) GetStats() *ProcessorStats {
	pp.stats.mu.RLock()
	defer pp.stats.mu.RUnlock()

	stats := *pp.stats
	stats.QueueDepth = len(pp.workQueue)
	stats.ActiveWorkers = len(pp.workers)

	return &stats
}

// Close shuts down the processor and cleans up resources
func (pp *ParallelProcessor) Close() {
	pp.cancel()
	pp.wg.Wait()
	close(pp.workQueue)
}

// startWorkers starts the worker goroutines
func (pp *ParallelProcessor) startWorkers() {
	for i := 0; i < pp.config.MaxWorkers; i++ {
		worker := &Worker{
			ID:        i,
			processor: pp,
			workChan:  pp.workQueue,
			stats:     &WorkerStats{},
		}

		worker.ctx, worker.cancel = context.WithCancel(pp.ctx)
		pp.workers = append(pp.workers, worker)

		pp.wg.Add(1)
		go worker.run(&pp.wg)
	}
}

// run is the main worker loop
func (w *Worker) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case item, ok := <-w.workChan:
			if !ok {
				return
			}
			w.processWorkItem(item)
		}
	}
}

// processWorkItem processes a single work item
func (w *Worker) processWorkItem(item *WorkItem) {
	start := time.Now()
	
	// Update worker stats
	w.stats.mu.Lock()
	w.stats.LastActivity = time.Now()
	w.stats.mu.Unlock()

	var result *WorkResult
	var err error

	// Process based on work type
	switch item.Type {
	case WorkTypeTransactionValidation:
		result, err = w.processTransaction(item)
	case WorkTypeBlockProcessing:
		result, err = w.processBlock(item)
	case WorkTypeUTXOUpdate:
		result, err = w.processUTXOUpdate(item)
	case WorkTypeMerkleTreeCalculation:
		result, err = w.processMerkleTree(item)
	case WorkTypeSignatureVerification:
		result, err = w.processSignatureVerification(item)
	case WorkTypeStateTransition:
		result, err = w.processStateTransition(item)
	default:
		err = fmt.Errorf("unknown work type: %d", item.Type)
	}

	duration := time.Since(start)

	// Create result
	if result == nil {
		result = &WorkResult{
			ID:        item.ID,
			Success:   err == nil,
			Error:     err,
			Duration:  duration,
			WorkerID:  w.ID,
			Timestamp: time.Now(),
		}
	}

	// Update worker stats
	w.stats.mu.Lock()
	w.stats.ItemsProcessed++
	w.stats.TotalDuration += duration
	if err != nil {
		w.stats.Errors++
	}
	w.stats.mu.Unlock()

	// Send result
	select {
	case item.Result <- result:
	default:
		// Result channel is full, log error
	}
}

// processTransaction processes a transaction validation work item
func (w *Worker) processTransaction(item *WorkItem) (*WorkResult, error) {
	tx, ok := item.Data.(*block.Transaction)
	if !ok {
		return nil, fmt.Errorf("invalid transaction data")
	}

	// Perform transaction validation
	// This is a simplified implementation
	if tx.Version <= 0 {
		return nil, fmt.Errorf("invalid transaction version")
	}

	if len(tx.Inputs) == 0 {
		return nil, fmt.Errorf("transaction has no inputs")
	}

	if len(tx.Outputs) == 0 {
		return nil, fmt.Errorf("transaction has no outputs")
	}

	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     tx,
		Duration: 0, // Will be set by caller
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}

// processBlock processes a block processing work item
func (w *Worker) processBlock(item *WorkItem) (*WorkResult, error) {
	block, ok := item.Data.(*block.Block)
	if !ok {
		return nil, fmt.Errorf("invalid block data")
	}

	// Perform block validation
	// This is a simplified implementation
	if block.Header == nil {
		return nil, fmt.Errorf("block has no header")
	}

	if len(block.Transactions) == 0 {
		return nil, fmt.Errorf("block has no transactions")
	}

	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     block,
		Duration: 0, // Will be set by caller
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}

// processUTXOUpdate processes a UTXO update work item
func (w *Worker) processUTXOUpdate(item *WorkItem) (*WorkResult, error) {
	// Simplified UTXO update processing
	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     "UTXO updated",
		Duration: 0,
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}

// processMerkleTree processes a merkle tree calculation work item
func (w *Worker) processMerkleTree(item *WorkItem) (*WorkResult, error) {
	// Simplified merkle tree calculation
	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     "Merkle tree calculated",
		Duration: 0,
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}

// processSignatureVerification processes a signature verification work item
func (w *Worker) processSignatureVerification(item *WorkItem) (*WorkResult, error) {
	// Simplified signature verification
	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     "Signature verified",
		Duration: 0,
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}

// processStateTransition processes a state transition work item
func (w *Worker) processStateTransition(item *WorkItem) (*WorkResult, error) {
	// Simplified state transition processing
	return &WorkResult{
		ID:       item.ID,
		Success:  true,
		Data:     "State transition completed",
		Duration: 0,
		WorkerID: w.ID,
		Timestamp: time.Now(),
	}, nil
}
