package testing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestRunner_GetActiveTests(t *testing.T) {
	tr := NewTestRunner()

	// Initially should be empty
	activeTests := tr.GetActiveTests()
	require.NotNil(t, activeTests)
	assert.Len(t, activeTests, 0)

	// Start some tests to populate active tests
	// Note: We'll need to use the Start method and let it run briefly
	go tr.Start()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Check active tests
	activeTests = tr.GetActiveTests()
	require.NotNil(t, activeTests)
	// Active tests may or may not be present depending on timing
	// The important thing is that the method doesn't panic and returns a valid slice
}

func TestTestRunner_GetCompletedTests(t *testing.T) {
	tr := NewTestRunner()

	// Initially should be empty
	completedTests := tr.GetCompletedTests()
	require.NotNil(t, completedTests)
	assert.Len(t, completedTests, 0)

	// Start and let some tests complete
	go tr.Start()

	// Give it time to complete some tests
	time.Sleep(100 * time.Millisecond)

	// Check completed tests
	completedTests = tr.GetCompletedTests()
	require.NotNil(t, completedTests)
	// Should be a valid slice (may be empty depending on timing)
}

func TestTestRunner_GetTestResult(t *testing.T) {
	tr := NewTestRunner()

	// Test with non-existent test ID
	result := tr.GetTestResult("non-existent-test")
	assert.Nil(t, result) // Should return nil for non-existent test

	// Start some tests
	go tr.Start()

	// Give it time to run
	time.Sleep(50 * time.Millisecond)

	// Test with potentially valid test ID
	// Since we don't know the exact test IDs, we'll test the method doesn't panic
	assert.NotPanics(t, func() {
		result := tr.GetTestResult("test-1")
		// Result may be nil or valid, both are acceptable
		_ = result
	})
}

func TestTestRunner_GetPerformanceMetrics(t *testing.T) {
	tr := NewTestRunner()

	// Should not panic even without running tests
	assert.NotPanics(t, func() {
		metrics := tr.GetPerformanceMetrics()
		// May return nil or valid metrics
		_ = metrics
	})

	// Start tests to generate metrics
	go tr.Start()

	// Give it time to generate metrics
	time.Sleep(50 * time.Millisecond)

	// Get performance metrics
	metrics := tr.GetPerformanceMetrics()
	// Should be accessible without panic
	_ = metrics
}

func TestTestRunner_GetMemoryMetrics(t *testing.T) {
	tr := NewTestRunner()

	// Should not panic even without running tests
	assert.NotPanics(t, func() {
		metrics := tr.GetMemoryMetrics()
		// May return nil or valid metrics
		_ = metrics
	})

	// Start tests to generate metrics
	go tr.Start()

	// Give it time to generate metrics
	time.Sleep(50 * time.Millisecond)

	// Get memory metrics
	metrics := tr.GetMemoryMetrics()
	// Should be accessible without panic
	_ = metrics
}

func TestTestRunner_GetCPUMetrics(t *testing.T) {
	tr := NewTestRunner()

	// Should not panic even without running tests
	assert.NotPanics(t, func() {
		metrics := tr.GetCPUMetrics()
		// May return nil or valid metrics
		_ = metrics
	})

	// Start tests to generate metrics
	go tr.Start()

	// Give it time to generate metrics
	time.Sleep(50 * time.Millisecond)

	// Get CPU metrics
	metrics := tr.GetCPUMetrics()
	// Should be accessible without panic
	_ = metrics
}

func TestTestRunner_GettersConcurrentAccess(t *testing.T) {
	tr := NewTestRunner()

	// Start test runner
	go tr.Start()

	// Give it time to initialize
	time.Sleep(25 * time.Millisecond)

	// Test concurrent access to all getter methods
	done := make(chan bool, 30)

	// Test GetActiveTests concurrently
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			activeTests := tr.GetActiveTests()
			assert.NotNil(t, activeTests)
		}()
	}

	// Test GetCompletedTests concurrently
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			completedTests := tr.GetCompletedTests()
			assert.NotNil(t, completedTests)
		}()
	}

	// Test GetTestResult concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			result := tr.GetTestResult("test-" + string(rune(id)))
			// Result may be nil, which is fine
			_ = result
		}(i)
	}

	// Test metrics getters concurrently
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			perfMetrics := tr.GetPerformanceMetrics()
			memMetrics := tr.GetMemoryMetrics()
			cpuMetrics := tr.GetCPUMetrics()
			// These may be nil or valid, both are acceptable
			_, _, _ = perfMetrics, memMetrics, cpuMetrics
			done <- true
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestTestRunner_GettersEdgeCases(t *testing.T) {
	tr := NewTestRunner()

	// Test all getters on uninitialized runner
	assert.NotPanics(t, func() {
		activeTests := tr.GetActiveTests()
		completedTests := tr.GetCompletedTests()
		result := tr.GetTestResult("")
		perfMetrics := tr.GetPerformanceMetrics()
		memMetrics := tr.GetMemoryMetrics()
		cpuMetrics := tr.GetCPUMetrics()

		// All should be accessible
		_, _, _, _, _, _ = activeTests, completedTests, result, perfMetrics, memMetrics, cpuMetrics
	})
}
