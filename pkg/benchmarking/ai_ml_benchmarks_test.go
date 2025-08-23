package benchmarking

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBenchmarkSuite(t *testing.T) {
	t.Run("NewBenchmarkSuite", func(t *testing.T) {
		suite := NewBenchmarkSuite()
		require.NotNil(t, suite)
		assert.NotNil(t, suite.Results)
		assert.Len(t, suite.Results, 0)
	})

	t.Run("AddResult", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		result := &BenchmarkResult{
			PackageName:     "test_package",
			TestName:        "test_benchmark",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      10000.0,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"key": "value"},
		}

		suite.AddResult(result)
		assert.Len(t, suite.Results, 1)
		assert.Equal(t, result, suite.Results[0])

		// Add another result
		result2 := &BenchmarkResult{
			PackageName: "test_package_2",
			TestName:    "test_benchmark_2",
			Duration:    200 * time.Millisecond,
		}
		suite.AddResult(result2)
		assert.Len(t, suite.Results, 2)
	})

	t.Run("GetResults", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		// Add some results
		result1 := &BenchmarkResult{PackageName: "pkg1", TestName: "test1"}
		result2 := &BenchmarkResult{PackageName: "pkg2", TestName: "test2"}

		suite.AddResult(result1)
		suite.AddResult(result2)

		// Get results
		results := suite.GetResults()
		assert.Len(t, results, 2)

		// Verify results are copied (not references)
		assert.NotSame(t, &suite.Results[0], &results[0])
		assert.NotSame(t, &suite.Results[1], &results[1])

		// Verify content
		assert.Equal(t, "pkg1", results[0].PackageName)
		assert.Equal(t, "pkg2", results[1].PackageName)
	})

	t.Run("RunAllBenchmarks", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		// Run all benchmarks
		err := suite.RunAllBenchmarks()
		assert.NoError(t, err)

		// Verify that results were added
		results := suite.GetResults()
		assert.Greater(t, len(results), 0)

		// Verify that all expected benchmark types are present
		benchmarkTypes := make(map[string]bool)
		for _, result := range results {
			benchmarkTypes[result.TestName] = true
		}

		// Debug: print all benchmark names
		t.Logf("Generated benchmark names: %v", benchmarkTypes)

		// Check for key benchmark types based on actual output
		assert.True(t, benchmarkTypes["Strategy Creation Performance"], "Strategy Creation Performance benchmark should be present")
		assert.True(t, benchmarkTypes["Prediction Performance"], "Prediction Performance benchmark should be present")
		assert.True(t, benchmarkTypes["Sentiment Analysis Performance"], "Sentiment Analysis Performance benchmark should be present")

		// Verify we have a good number of benchmarks
		assert.Greater(t, len(benchmarkTypes), 10, "Should have many different benchmark types")
	})
}

func TestBenchmarkResult(t *testing.T) {
	t.Run("BenchmarkResult_Complete", func(t *testing.T) {
		result := &BenchmarkResult{
			PackageName:     "test_package",
			TestName:        "test_benchmark",
			Duration:        150 * time.Millisecond,
			MemoryUsage:     2048,
			OperationsCount: 2000,
			Throughput:      13333.33,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata: map[string]interface{}{
				"cpu_cores": 4,
				"memory_gb": 8,
				"version":   "1.0.0",
			},
		}

		// Verify all fields are set correctly
		assert.Equal(t, "test_package", result.PackageName)
		assert.Equal(t, "test_benchmark", result.TestName)
		assert.Equal(t, 150*time.Millisecond, result.Duration)
		assert.Equal(t, uint64(2048), result.MemoryUsage)
		assert.Equal(t, int64(2000), result.OperationsCount)
		assert.Equal(t, 13333.33, result.Throughput)
		assert.Equal(t, 1.024, result.MemoryPerOp)
		assert.True(t, result.Timestamp.After(time.Time{}))
		assert.Len(t, result.Metadata, 3)
		assert.Equal(t, 4, result.Metadata["cpu_cores"])
		assert.Equal(t, 8, result.Metadata["memory_gb"])
		assert.Equal(t, "1.0.0", result.Metadata["version"])
	})

	t.Run("BenchmarkResult_Minimal", func(t *testing.T) {
		result := &BenchmarkResult{
			PackageName: "minimal_package",
			TestName:    "minimal_test",
		}

		// Verify minimal fields work
		assert.Equal(t, "minimal_package", result.PackageName)
		assert.Equal(t, "minimal_test", result.TestName)
		assert.Equal(t, time.Duration(0), result.Duration)
		assert.Equal(t, uint64(0), result.MemoryUsage)
		assert.Equal(t, int64(0), result.OperationsCount)
		assert.Equal(t, 0.0, result.Throughput)
		assert.Equal(t, 0.0, result.MemoryPerOp)
		assert.Equal(t, time.Time{}, result.Timestamp)
		assert.Nil(t, result.Metadata)
	})
}

func TestBenchmarkReport(t *testing.T) {
	t.Run("BenchmarkReport_Creation", func(t *testing.T) {
		report := &BenchmarkReport{
			Timestamp:       time.Now(),
			TotalBenchmarks: 5,
			Results: []*BenchmarkResult{
				{PackageName: "pkg1", TestName: "test1"},
				{PackageName: "pkg2", TestName: "test2"},
			},
			Summary: "All benchmarks completed successfully",
		}

		assert.True(t, report.Timestamp.After(time.Time{}))
		assert.Equal(t, 5, report.TotalBenchmarks)
		assert.Len(t, report.Results, 2)
		assert.Equal(t, "All benchmarks completed successfully", report.Summary)
	})
}

func TestConcurrentBenchmarkSuite(t *testing.T) {
	t.Run("ConcurrentAddResult", func(t *testing.T) {
		suite := NewBenchmarkSuite()
		const numGoroutines = 100
		const resultsPerGoroutine = 10

		var wg sync.WaitGroup

		// Start multiple goroutines adding results concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < resultsPerGoroutine; j++ {
					result := &BenchmarkResult{
						PackageName: fmt.Sprintf("pkg_%d", id),
						TestName:    fmt.Sprintf("test_%d_%d", id, j),
						Duration:    time.Duration(id+j) * time.Millisecond,
					}
					suite.AddResult(result)
				}
			}(i)
		}

		wg.Wait()

		// Verify all results were added
		results := suite.GetResults()
		expectedTotal := numGoroutines * resultsPerGoroutine
		assert.Equal(t, expectedTotal, len(results))
	})

	t.Run("ConcurrentGetResults", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		// Add some results first
		for i := 0; i < 50; i++ {
			result := &BenchmarkResult{
				PackageName: fmt.Sprintf("pkg_%d", i),
				TestName:    fmt.Sprintf("test_%d", i),
			}
			suite.AddResult(result)
		}

		const numReaders = 20
		var wg sync.WaitGroup

		// Start multiple goroutines reading results concurrently
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				results := suite.GetResults()
				assert.Len(t, results, 50)
				// Verify we can read the results safely
				for _, result := range results {
					assert.NotEmpty(t, result.PackageName)
					assert.NotEmpty(t, result.TestName)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestBenchmarkSuiteEdgeCases(t *testing.T) {
	t.Run("AddNilResult", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		// Adding nil result should not panic
		assert.NotPanics(t, func() {
			suite.AddResult(nil)
		})

		// Results should still be accessible
		results := suite.GetResults()
		assert.Len(t, results, 1)
		assert.Nil(t, results[0])
	})

	t.Run("EmptySuiteGetResults", func(t *testing.T) {
		suite := NewBenchmarkSuite()

		// Getting results from empty suite should return empty slice
		results := suite.GetResults()
		assert.NotNil(t, results)
		assert.Len(t, results, 0)
	})

	t.Run("LargeNumberOfResults", func(t *testing.T) {
		suite := NewBenchmarkSuite()
		const largeNumber = 10000

		// Add a large number of results
		for i := 0; i < largeNumber; i++ {
			result := &BenchmarkResult{
				PackageName: fmt.Sprintf("pkg_%d", i),
				TestName:    fmt.Sprintf("test_%d", i),
				Duration:    time.Duration(i) * time.Microsecond,
			}
			suite.AddResult(result)
		}

		// Verify all results were added
		results := suite.GetResults()
		assert.Len(t, results, largeNumber)

		// Verify some specific results
		assert.Equal(t, "pkg_0", results[0].PackageName)
		assert.Equal(t, "pkg_9999", results[9999].PackageName)
	})
}
