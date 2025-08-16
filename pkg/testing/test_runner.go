package testing

import (
	"context"
	"fmt"
	"time"
)

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		isRunning:          false,
		activeTests:        make(map[string]*TestCase),
		completedTests:     make(map[string]*TestCase),
		performanceMonitor: NewPerformanceMonitor(),
		memoryMonitor:      NewMemoryMonitor(),
		cpuMonitor:         NewCPUMonitor(),
		testResults:        make(map[string]*TestResult),
		resultQueue:        make(chan *TestResult, 1000),
	}
}

// Start begins test execution
func (tr *TestRunner) Start() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.isRunning {
		return ErrTestExecutionFailed
	}

	tr.isRunning = true

	// Start monitoring
	tr.performanceMonitor.Start()
	tr.memoryMonitor.Start()
	tr.cpuMonitor.Start()

	// Start background monitoring
	go tr.monitoringLoop()

	return nil
}

// Stop ends test execution
func (tr *TestRunner) Stop() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if !tr.isRunning {
		return
	}

	tr.isRunning = false

	// Stop monitoring
	tr.performanceMonitor.Stop()

	// Close result queue safely
	select {
	case <-tr.resultQueue:
		// Channel is already closed or empty
	default:
		// Channel is open, close it
		close(tr.resultQueue)
	}
}

// ExecuteTest executes a single test case
func (tr *TestRunner) ExecuteTest(ctx context.Context, testCase *TestCase) (*TestResult, error) {
	// Handle nil test case
	if testCase == nil {
		return &TestResult{
			TestCaseID:  "",
			Status:      TestStatusFailed,
			Duration:    0,
			Error:       fmt.Errorf("test case is nil"),
			MemoryUsage: 0,
			CPUUsage:    0,
			Logs:        []string{},
			Timestamp:   time.Now(),
		}, nil
	}

	tr.mu.Lock()
	tr.activeTests[testCase.ID] = testCase
	tr.mu.Unlock()

	// Execute test
	result := tr.runTestCase(ctx, testCase)

	// Record result
	tr.mu.Lock()
	tr.completedTests[testCase.ID] = testCase
	delete(tr.activeTests, testCase.ID)
	tr.testResults[testCase.ID] = result
	tr.mu.Unlock()

	// Send to result queue
	select {
	case tr.resultQueue <- result:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return result, nil
}

// GetActiveTests returns currently active tests
func (tr *TestRunner) GetActiveTests() []*TestCase {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	activeTests := make([]*TestCase, 0, len(tr.activeTests))
	for _, testCase := range tr.activeTests {
		activeTests = append(activeTests, testCase)
	}

	return activeTests
}

// GetCompletedTests returns completed tests
func (tr *TestRunner) GetCompletedTests() []*TestCase {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	completedTests := make([]*TestCase, 0, len(tr.completedTests))
	for _, testCase := range tr.completedTests {
		completedTests = append(completedTests, testCase)
	}

	return completedTests
}

// GetTestResult returns the result of a specific test
func (tr *TestRunner) GetTestResult(testCaseID string) *TestResult {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if result, exists := tr.testResults[testCaseID]; exists {
		// Return a copy to avoid race conditions
		resultCopy := &TestResult{
			TestCaseID: result.TestCaseID,
			Status:     result.Status,
			Duration:   result.Duration,
			Error:      result.Error,
			// Coverage is tracked separately
			MemoryUsage: result.MemoryUsage,
			CPUUsage:    result.CPUUsage,
			Logs:        make([]string, len(result.Logs)),
			Timestamp:   result.Timestamp,
		}

		copy(resultCopy.Logs, result.Logs)

		return resultCopy
	}

	return nil
}

// GetPerformanceMetrics returns current performance metrics
func (tr *TestRunner) GetPerformanceMetrics() *PerformanceMetrics {
	return tr.performanceMonitor.GetPerformanceMetrics()
}

// GetMemoryMetrics returns current memory metrics
func (tr *TestRunner) GetMemoryMetrics() *MemoryMetrics {
	return tr.memoryMonitor.GetMemoryMetrics()
}

// GetCPUMetrics returns current CPU metrics
func (tr *TestRunner) GetCPUMetrics() *CPUMetrics {
	return tr.cpuMonitor.GetCPUMetrics()
}

// Helper functions
func (tr *TestRunner) runTestCase(ctx context.Context, testCase *TestCase) *TestResult {
	startTime := time.Now()

	// Sample memory and CPU before test
	tr.memoryMonitor.Sample()
	tr.cpuMonitor.Sample()
	startMemory := tr.memoryMonitor.GetMemoryMetrics().CurrentMemory
	startCPU := tr.cpuMonitor.GetCPUMetrics().CurrentCPU

	// Execute test function
	var err error
	if testCase.Function == nil {
		err = fmt.Errorf("test function is nil")
	} else {
		// Create test context with timeout
		testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Run test in goroutine to handle timeout
		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- ErrTestExecutionFailed
				}
			}()

			// Execute test
			if testCase.Setup != nil {
				if setupErr := testCase.Setup(); setupErr != nil {
					done <- setupErr
					return
				}
			}

			// Run test function
			testErr := testCase.Function(nil) // Placeholder for testing.T

			// Run teardown
			if testCase.Teardown != nil {
				if teardownErr := testCase.Teardown(); teardownErr != nil {
					done <- teardownErr
					return
				}
			}

			done <- testErr
		}()

		// Wait for completion or timeout
		select {
		case testErr := <-done:
			err = testErr
		case <-testCtx.Done():
			err = ErrTestTimeout
		}
	}

	duration := time.Since(startTime)

	// Sample memory and CPU after test
	tr.memoryMonitor.Sample()
	tr.cpuMonitor.Sample()
	endMemory := tr.memoryMonitor.GetMemoryMetrics().CurrentMemory
	endCPU := tr.cpuMonitor.GetCPUMetrics().CurrentCPU

	// Calculate memory usage
	memoryUsage := uint64(0)
	if endMemory > startMemory {
		memoryUsage = endMemory - startMemory
	}

	// Calculate CPU usage
	cpuUsage := endCPU - startCPU
	if cpuUsage < 0 {
		cpuUsage = 0
	}

	// Determine test status
	status := TestStatusPassed
	if err != nil {
		if err == ErrTestTimeout {
			status = TestStatusTimeout
		} else {
			status = TestStatusFailed
		}
	}

	// Create test result
	result := &TestResult{
		TestCaseID: testCase.ID,
		Status:     status,
		Duration:   duration,
		Error:      err,
		// Coverage is tracked separately
		MemoryUsage: memoryUsage,
		CPUUsage:    cpuUsage,
		Logs:        make([]string, 0),
		Timestamp:   time.Now(),
	}

	// Record performance metrics
	tr.performanceMonitor.RecordTestExecution(duration, memoryUsage, cpuUsage)

	return result
}

func (tr *TestRunner) monitoringLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if tr.isRunning {
				tr.memoryMonitor.Sample()
				tr.cpuMonitor.Sample()
			}
		}

		tr.mu.RLock()
		if !tr.isRunning {
			tr.mu.RUnlock()
			break
		}
		tr.mu.RUnlock()
	}
}
