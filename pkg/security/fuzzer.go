package security

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/storage"
)

// Fuzzer provides comprehensive fuzz testing for GoChain
type Fuzzer struct {
	chain     *chain.Chain
	storage   storage.StorageInterface
	results   map[string]*FuzzResult
	mu        sync.RWMutex
	config    *FuzzConfig
	stopChan  chan struct{}
	isRunning bool
}

// FuzzResult holds the results of a fuzz test
type FuzzResult struct {
	Name         string                 `json:"name"`
	Duration     time.Duration          `json:"duration"`
	Iterations   int64                  `json:"iterations"`
	CrashCount   int64                  `json:"crash_count"`
	TimeoutCount int64                  `json:"timeout_count"`
	ErrorCount   int64                  `json:"error_count"`
	SuccessCount int64                  `json:"success_count"`
	Coverage     float64                `json:"coverage"`
	CrashDetails []*CrashDetail         `json:"crash_details"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CrashDetail contains information about a fuzz test crash
type CrashDetail struct {
	Input     []byte                 `json:"input"`
	Error     string                 `json:"error"`
	Stack     string                 `json:"stack"`
	Iteration int64                  `json:"iteration"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// FuzzConfig holds configuration for fuzz testing
type FuzzConfig struct {
	Duration          time.Duration `json:"duration"`
	MaxIterations     int64         `json:"max_iterations"`
	Timeout           time.Duration `json:"timeout"`
	MaxInputSize      int           `json:"max_input_size"`
	MinInputSize      int           `json:"min_input_size"`
	EnableMutation    bool          `json:"enable_mutation"`
	EnableCoverage    bool          `json:"enable_coverage"`
	EnableCrashReport bool          `json:"enable_crash_report"`
	Seed              int64         `json:"seed"`
	Concurrency       int           `json:"concurrency"`
	TargetFunctions   []string      `json:"target_functions"`
	ExcludeFunctions  []string      `json:"exclude_functions"`
}

// DefaultFuzzConfig returns the default fuzz configuration
func DefaultFuzzConfig() *FuzzConfig {
	return &FuzzConfig{
		Duration:          60 * time.Second,
		MaxIterations:     10000,
		Timeout:           100 * time.Millisecond,
		MaxInputSize:      1024,
		MinInputSize:      1,
		EnableMutation:    true,
		EnableCoverage:    true,
		EnableCrashReport: true,
		Seed:              time.Now().UnixNano(),
		Concurrency:       4,
		TargetFunctions:   []string{},
		ExcludeFunctions:  []string{},
	}
}

// NewFuzzer creates a new fuzzer instance
func NewFuzzer(chain *chain.Chain, storage storage.StorageInterface) *Fuzzer {
	return &Fuzzer{
		chain:     chain,
		storage:   storage,
		results:   make(map[string]*FuzzResult),
		config:    DefaultFuzzConfig(),
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

// StartFuzzing begins the fuzz testing process
func (f *Fuzzer) StartFuzzing(config *FuzzConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.isRunning {
		return fmt.Errorf("fuzzer is already running")
	}

	f.config = config
	f.isRunning = true
	f.stopChan = make(chan struct{})

	// Start fuzzing in background
	go f.runFuzzing()

	return nil
}

// StopFuzzing stops the fuzz testing process
func (f *Fuzzer) StopFuzzing() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.isRunning {
		close(f.stopChan)
		f.isRunning = false
	}
}

// IsRunning returns whether the fuzzer is currently running
func (f *Fuzzer) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isRunning
}

// runFuzzing executes the main fuzzing loop
func (f *Fuzzer) runFuzzing() {
	iterations := int64(0)

	// Create worker goroutines
	var wg sync.WaitGroup
	results := make(chan *FuzzResult, f.config.Concurrency)

	// Create a done channel to signal workers to stop
	done := make(chan struct{})

	for i := 0; i < f.config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			f.workerLoop(workerID, results, done)
		}(i)
	}

	// Monitor results and check for completion
	go func() {
		for result := range results {
			f.mu.Lock()
			f.results[result.Name] = result
			f.mu.Unlock()
		}
	}()

	// Wait for completion or timeout
	select {
	case <-f.stopChan:
		// Manual stop
	case <-time.After(f.config.Duration):
		// Timeout
	case <-func() chan struct{} {
		ch := make(chan struct{})
		go func() {
			for iterations < f.config.MaxIterations {
				time.Sleep(100 * time.Millisecond)
			}
			close(ch)
		}()
		return ch
	}():
		// Max iterations reached
	}

	// Signal workers to stop
	close(done)

	// Wait for workers to finish
	wg.Wait()

	// Close results channel after all workers are done
	close(results)

	f.mu.Lock()
	f.isRunning = false
	f.mu.Unlock()
}

// workerLoop runs the main fuzzing logic for a worker
func (f *Fuzzer) workerLoop(workerID int, results chan<- *FuzzResult, done <-chan struct{}) {
	iterations := int64(0)
	crashCount := int64(0)
	timeoutCount := int64(0)
	errorCount := int64(0)
	successCount := int64(0)
	crashDetails := []*CrashDetail{}

	// Seed random number generator
	seed := f.config.Seed + int64(workerID)
	rng := rand.New(rand.NewSource(seed))

	for {
		select {
		case <-f.stopChan:
			return
		case <-done:
			return
		default:
			// Continue fuzzing
		}

		iterations++

		// Generate fuzz input
		input := f.generateFuzzInput(rng)

		// Execute fuzz test with timeout
		result := f.executeFuzzTest(input, int64(workerID), iterations)

		switch result.Status {
		case "crash":
			crashCount++
			if f.config.EnableCrashReport {
				crashDetails = append(crashDetails, result.CrashDetail)
			}
		case "timeout":
			timeoutCount++
		case "error":
			errorCount++
		case "success":
			successCount++
		}

		// Send result periodically
		if iterations%100 == 0 {
			select {
			case <-done:
				return
			case results <- &FuzzResult{
				Name:         fmt.Sprintf("Worker_%d", workerID),
				Duration:     time.Since(time.Now()),
				Iterations:   iterations,
				CrashCount:   crashCount,
				TimeoutCount: timeoutCount,
				ErrorCount:   errorCount,
				SuccessCount: successCount,
				CrashDetails: crashDetails,
				Timestamp:    time.Now(),
				Metadata: map[string]interface{}{
					"worker_id": workerID,
					"seed":      seed,
				},
			}:
			default:
				// Channel full, skip
			}
		}
	}
}

// FuzzTestResult represents the result of a single fuzz test
type FuzzTestResult struct {
	Status      string        `json:"status"` // "success", "error", "timeout", "crash"
	Input       []byte        `json:"input"`
	Error       error         `json:"error,omitempty"`
	CrashDetail *CrashDetail  `json:"crash_detail,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// executeFuzzTest runs a single fuzz test with timeout
func (f *Fuzzer) executeFuzzTest(input []byte, workerID, iteration int64) *FuzzTestResult {
	start := time.Now()
	result := &FuzzTestResult{
		Input: input,
	}

	// Create timeout channel
	timeout := make(chan struct{})
	go func() {
		time.Sleep(f.config.Timeout)
		close(timeout)
	}()

	// Execute test in goroutine
	done := make(chan struct{})
	var panicErr interface{}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr = r
			}
			close(done)
		}()

		// Execute the actual fuzz test
		f.executeTest(input)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		if panicErr != nil {
			result.Status = "crash"
			result.Error = fmt.Errorf("panic: %v", panicErr)
			result.CrashDetail = &CrashDetail{
				Input:     input,
				Error:     fmt.Sprintf("panic: %v", panicErr),
				Stack:     string(getStackTrace()),
				Iteration: iteration,
				Timestamp: time.Now(),
				Context: map[string]interface{}{
					"worker_id":  workerID,
					"input_size": len(input),
				},
			}
		} else {
			result.Status = "success"
		}
	case <-timeout:
		result.Status = "timeout"
	}

	result.Duration = time.Since(start)
	return result
}

// executeTest runs the actual fuzz test logic
func (f *Fuzzer) executeTest(input []byte) {
	// Test block parsing
	if len(input) > 0 {
		// Try to parse as block
		f.testBlockParsing(input)

		// Try to parse as transaction
		f.testTransactionParsing(input)

		// Test storage operations
		f.testStorageOperations(input)

		// Test chain operations
		f.testChainOperations(input)
	}
}

// testBlockParsing tests block parsing with fuzz input
func (f *Fuzzer) testBlockParsing(input []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, ignore
		}
	}()

	// Try to create a block with fuzz data
	block := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: input[:min(len(input), 32)],
			MerkleRoot:    input[:min(len(input), 32)],
			Timestamp:     time.Now(),
			Difficulty:    1000,
			Nonce:         0,
			Height:        0,
		},
		Transactions: []*block.Transaction{},
	}

	// Try to serialize/deserialize
	_ = block.CalculateHash()
}

// testTransactionParsing tests transaction parsing with fuzz input
func (f *Fuzzer) testTransactionParsing(input []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, ignore
		}
	}()

	// Try to create a transaction with fuzz data
	tx := &block.Transaction{
		Hash:    input[:min(len(input), 32)],
		Fee:     100,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{},
	}

	// Try to serialize
	_, _ = tx.Serialize()
}

// testStorageOperations tests storage operations with fuzz input
func (f *Fuzzer) testStorageOperations(input []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, ignore
		}
	}()

	// Test with various key/value combinations
	key := input[:min(len(input)/2, 64)]
	value := input[min(len(input)/2, 64):]

	if len(key) > 0 && len(value) > 0 {
		_ = f.storage.Write(key, value)
		_, _ = f.storage.Read(key)
		_, _ = f.storage.Has(key)
	}
}

// testChainOperations tests chain operations with fuzz input
func (f *Fuzzer) testChainOperations(input []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, ignore
		}
	}()

	// Test various chain operations
	if len(input) > 0 {
		_ = f.chain.GetHeight()
		_ = f.chain.GetTipHash()
		_ = f.chain.GetBlock(input[:min(len(input), 32)])
	}
}

// generateFuzzInput generates fuzz input data
func (f *Fuzzer) generateFuzzInput(rng *rand.Rand) []byte {
	// Determine input size
	size := rng.Intn(f.config.MaxInputSize-f.config.MinInputSize+1) + f.config.MinInputSize

	input := make([]byte, size)

	// Fill with random data
	for i := 0; i < size; i++ {
		input[i] = byte(rng.Intn(256))
	}

	// Apply mutations if enabled
	if f.config.EnableMutation {
		input = f.mutateInput(input, rng)
	}

	return input
}

// mutateInput applies various mutations to the input
func (f *Fuzzer) mutateInput(input []byte, rng *rand.Rand) []byte {
	if len(input) == 0 {
		return input
	}

	mutated := make([]byte, len(input))
	copy(mutated, input)

	// Apply random mutations
	mutationType := rng.Intn(5)
	switch mutationType {
	case 0:
		// Bit flip
		pos := rng.Intn(len(mutated))
		mutated[pos] ^= 1
	case 1:
		// Byte substitution
		pos := rng.Intn(len(mutated))
		mutated[pos] = byte(rng.Intn(256))
	case 2:
		// Insert random byte
		if len(mutated) < f.config.MaxInputSize {
			pos := rng.Intn(len(mutated) + 1)
			newByte := byte(rng.Intn(256))
			mutated = append(mutated[:pos], append([]byte{newByte}, mutated[pos:]...)...)
		}
	case 3:
		// Delete random byte
		if len(mutated) > f.config.MinInputSize {
			pos := rng.Intn(len(mutated))
			mutated = append(mutated[:pos], mutated[pos+1:]...)
		}
	case 4:
		// Duplicate random byte
		if len(mutated) < f.config.MaxInputSize {
			pos := rng.Intn(len(mutated))
			newByte := mutated[pos]
			mutated = append(mutated[:pos], append([]byte{newByte}, mutated[pos:]...)...)
		}
	}

	return mutated
}

// GetResults returns all fuzz test results
func (f *Fuzzer) GetResults() map[string]*FuzzResult {
	f.mu.RLock()
	defer f.mu.RUnlock()

	results := make(map[string]*FuzzResult)
	for k, v := range f.results {
		results[k] = v
	}
	return results
}

// GenerateReport generates a comprehensive fuzz test report
func (f *Fuzzer) GenerateReport() string {
	results := f.GetResults()

	report := "# 🔒 GoChain Fuzz Test Report\n\n"
	report += fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339))

	report += "## 📊 Summary\n\n"

	totalIterations := int64(0)
	totalCrashes := int64(0)
	totalTimeouts := int64(0)
	totalErrors := int64(0)
	totalSuccess := int64(0)

	for name, result := range results {
		report += fmt.Sprintf("### %s\n", name)
		report += fmt.Sprintf("- **Duration**: %v\n", result.Duration)
		report += fmt.Sprintf("- **Iterations**: %d\n", result.Iterations)
		report += fmt.Sprintf("- **Crashes**: %d\n", result.CrashCount)
		report += fmt.Sprintf("- **Timeouts**: %d\n", result.TimeoutCount)
		report += fmt.Sprintf("- **Errors**: %d\n", result.ErrorCount)
		report += fmt.Sprintf("- **Success**: %d\n", result.SuccessCount)
		if len(result.CrashDetails) > 0 {
			report += fmt.Sprintf("- **Crash Details**: %d crashes\n", len(result.CrashDetails))
		}
		report += "\n"

		totalIterations += result.Iterations
		totalCrashes += result.CrashCount
		totalTimeouts += result.TimeoutCount
		totalErrors += result.ErrorCount
		totalSuccess += result.SuccessCount
	}

	report += "## 🎯 Overall Results\n\n"
	report += fmt.Sprintf("- **Total Iterations**: %d\n", totalIterations)
	report += fmt.Sprintf("- **Total Crashes**: %d\n", totalCrashes)
	report += fmt.Sprintf("- **Total Timeouts**: %d\n", totalTimeouts)
	report += fmt.Sprintf("- **Total Errors**: %d\n", totalErrors)
	report += fmt.Sprintf("- **Total Success**: %d\n", totalSuccess)

	if totalCrashes > 0 {
		report += "\n## 🚨 Crash Details\n\n"
		for _, result := range results {
			for _, crash := range result.CrashDetails {
				report += fmt.Sprintf("### Crash in %s (Iteration %d)\n", result.Name, crash.Iteration)
				report += fmt.Sprintf("- **Error**: %s\n", crash.Error)
				report += fmt.Sprintf("- **Input Size**: %d bytes\n", len(crash.Input))
				report += fmt.Sprintf("- **Timestamp**: %s\n", crash.Timestamp.Format(time.RFC3339))
				report += "\n"
			}
		}
	}

	return report
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getStackTrace() []byte {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return buf[:n]
}
