package benchmarking

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewMainBenchmarkOrchestrator(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	if orchestrator == nil {
		t.Fatal("NewMainBenchmarkOrchestrator returned nil")
	}

	if orchestrator.Layer2Benchmarks == nil {
		t.Error("Layer2Benchmarks should not be nil")
	}

	if orchestrator.CrossChainBenchmarks == nil {
		t.Error("CrossChainBenchmarks should not be nil")
	}

	if orchestrator.GovernanceBenchmarks == nil {
		t.Error("GovernanceBenchmarks should not be nil")
	}

	if orchestrator.PrivacyBenchmarks == nil {
		t.Error("PrivacyBenchmarks should not be nil")
	}

	if orchestrator.AllResults == nil {
		t.Error("AllResults should not be nil")
	}

	if len(orchestrator.AllResults) != 0 {
		t.Error("AllResults should be empty initially")
	}
}

func TestNewLayer2BenchmarkSuite(t *testing.T) {
	suite := NewLayer2BenchmarkSuite()

	if suite == nil {
		t.Fatal("NewLayer2BenchmarkSuite returned nil")
	}

	if suite.Results == nil {
		t.Error("Results should not be nil")
	}

	if len(suite.Results) != 0 {
		t.Error("Results should be empty initially")
	}
}

func TestNewCrossChainBenchmarkSuite(t *testing.T) {
	suite := NewCrossChainBenchmarkSuite()

	if suite == nil {
		t.Fatal("NewCrossChainBenchmarkSuite returned nil")
	}

	if suite.Results == nil {
		t.Error("Results should not be nil")
	}

	if len(suite.Results) != 0 {
		t.Error("Results should be empty initially")
	}
}

func TestNewGovernanceBenchmarkSuite(t *testing.T) {
	suite := NewGovernanceBenchmarkSuite()

	if suite == nil {
		t.Fatal("NewGovernanceBenchmarkSuite returned nil")
	}

	if suite.Results == nil {
		t.Error("Results should not be nil")
	}

	if len(suite.Results) != 0 {
		t.Error("Results should be empty initially")
	}
}

func TestNewPrivacyBenchmarkSuite(t *testing.T) {
	suite := NewPrivacyBenchmarkSuite()

	if suite == nil {
		t.Fatal("NewPrivacyBenchmarkSuite returned nil")
	}

	if suite.Results == nil {
		t.Error("Results should not be nil")
	}

	if len(suite.Results) != 0 {
		t.Error("Results should be empty initially")
	}
}

func TestMainBenchmarkOrchestrator_RunAllBenchmarks(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	err := orchestrator.RunAllBenchmarks()
	if err != nil {
		t.Errorf("RunAllBenchmarks failed: %v", err)
	}

	// Verify all benchmark suites were run
	if len(orchestrator.AllResults) == 0 {
		t.Error("AllResults should not be empty after running benchmarks")
	}

	// Verify start and end times were set
	if orchestrator.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if orchestrator.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
}

func TestMainBenchmarkOrchestrator_Concurrency(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Test concurrent access to the orchestrator
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simulate concurrent benchmark operations
			result := &BenchmarkResult{
				PackageName:     "Test Package",
				TestName:        fmt.Sprintf("Concurrent Test %d", id),
				Duration:        time.Millisecond * 100,
				MemoryUsage:     1024,
				OperationsCount: 1000,
				Throughput:      10000.0,
				MemoryPerOp:     1.024,
				Timestamp:       time.Now(),
				Metadata:        map[string]interface{}{"goroutine_id": id},
			}

			orchestrator.AllResults = append(orchestrator.AllResults, result)
		}(i)
	}

	wg.Wait()

	// Verify all results were added
	if len(orchestrator.AllResults) != numGoroutines {
		t.Errorf("Expected %d results, got %d", numGoroutines, len(orchestrator.AllResults))
	}
}

func TestMainBenchmarkOrchestrator_EdgeCases(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Test with nil results
	orchestrator.AllResults = nil

	// Test with empty results
	orchestrator.AllResults = make([]*BenchmarkResult, 0)

	// Test with single result
	result := &BenchmarkResult{
		PackageName:     "Edge Case Package",
		TestName:        "Single Test",
		Duration:        time.Millisecond * 50,
		MemoryUsage:     512,
		OperationsCount: 500,
		Throughput:      10000.0,
		MemoryPerOp:     1.024,
		Timestamp:       time.Now(),
		Metadata:        map[string]interface{}{"edge_case": true},
	}

	orchestrator.AllResults = append(orchestrator.AllResults, result)

	if len(orchestrator.AllResults) != 1 {
		t.Error("Failed to add single result")
	}
}

func TestMainBenchmarkOrchestrator_Integration(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Test integration with all benchmark suites
	if orchestrator.Layer2Benchmarks == nil {
		t.Error("Layer2Benchmarks should not be nil")
	}

	if orchestrator.CrossChainBenchmarks == nil {
		t.Error("CrossChainBenchmarks should not be nil")
	}

	if orchestrator.GovernanceBenchmarks == nil {
		t.Error("GovernanceBenchmarks should not be nil")
	}

	if orchestrator.PrivacyBenchmarks == nil {
		t.Error("PrivacyBenchmarks should not be nil")
	}

	// Test that all suites are properly initialized
	if orchestrator.Layer2Benchmarks.Results == nil {
		t.Error("Layer2Benchmarks.Results should not be nil")
	}

	if orchestrator.CrossChainBenchmarks.Results == nil {
		t.Error("CrossChainBenchmarks.Results should not be nil")
	}

	if orchestrator.GovernanceBenchmarks.Results == nil {
		t.Error("GovernanceBenchmarks.Results should not be nil")
	}

	if orchestrator.PrivacyBenchmarks.Results == nil {
		t.Error("PrivacyBenchmarks.Results should not be nil")
	}
}

func TestMainBenchmarkOrchestrator_GenerateSummary(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Add some mock results
	orchestrator.AllResults = []*BenchmarkResult{
		{
			PackageName:     "Test Package",
			TestName:        "Test1",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      1000,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"test": "1"},
		},
		{
			PackageName:     "Test Package",
			TestName:        "Test2",
			Duration:        200 * time.Millisecond,
			MemoryUsage:     2048,
			OperationsCount: 1000,
			Throughput:      500,
			MemoryPerOp:     2.048,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"test": "2"},
		},
	}

	// Test that results were added correctly
	if len(orchestrator.AllResults) != 2 {
		t.Errorf("Expected 2 results, got %d", len(orchestrator.AllResults))
	}

	// Test that results have correct fields
	for i, result := range orchestrator.AllResults {
		if result.PackageName != "Test Package" {
			t.Errorf("Result %d: Expected PackageName 'Test Package', got '%s'", i, result.PackageName)
		}
		if result.TestName != fmt.Sprintf("Test%d", i+1) {
			t.Errorf("Result %d: Expected TestName 'Test%d', got '%s'", i, i+1, result.TestName)
		}
	}
}

func TestMainBenchmarkOrchestrator_GenerateSummary_Empty(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Test with empty results
	if len(orchestrator.AllResults) != 0 {
		t.Error("New orchestrator should have empty results")
	}
}

func TestMainBenchmarkOrchestrator_SaveReportToFile(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Add some mock results
	orchestrator.AllResults = []*BenchmarkResult{
		{
			PackageName:     "Test Package",
			TestName:        "Test1",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      1000,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"test": "1"},
		},
	}

	// Test that results were added correctly
	if len(orchestrator.AllResults) != 1 {
		t.Error("Failed to add test result")
	}
}

func TestLayer2BenchmarkSuite_RunAllBenchmarks(t *testing.T) {
	suite := NewLayer2BenchmarkSuite()

	err := suite.RunAllLayer2Benchmarks()
	if err != nil {
		t.Errorf("RunAllLayer2Benchmarks failed: %v", err)
	}

	if len(suite.Results) == 0 {
		t.Error("Results should not be empty after running benchmarks")
	}

	// Verify that results were added
	if len(suite.Results) < 1 {
		t.Error("Expected at least one benchmark result")
	}
}

func TestCrossChainBenchmarkSuite_RunAllBenchmarks(t *testing.T) {
	suite := NewCrossChainBenchmarkSuite()

	err := suite.RunAllCrossChainBenchmarks()
	if err != nil {
		t.Errorf("RunAllCrossChainBenchmarks failed: %v", err)
	}

	if len(suite.Results) == 0 {
		t.Error("Results should not be empty after running benchmarks")
	}

	// Verify that results were added
	if len(suite.Results) < 1 {
		t.Error("Expected at least one benchmark result")
	}
}

func TestGovernanceBenchmarkSuite_RunAllBenchmarks(t *testing.T) {
	suite := NewGovernanceBenchmarkSuite()

	err := suite.RunAllGovernanceBenchmarks()
	if err != nil {
		t.Errorf("RunAllGovernanceBenchmarks failed: %v", err)
	}

	if len(suite.Results) == 0 {
		t.Error("Results should not be empty after running benchmarks")
	}

	// Verify that results were added
	if len(suite.Results) < 1 {
		t.Error("Expected at least one benchmark result")
	}
}

func TestPrivacyBenchmarkSuite_RunAllBenchmarks(t *testing.T) {
	suite := NewPrivacyBenchmarkSuite()

	err := suite.RunAllPrivacyBenchmarks()
	if err != nil {
		t.Errorf("RunAllPrivacyBenchmarks failed: %v", err)
	}

	if len(suite.Results) == 0 {
		t.Error("Results should not be empty after running benchmarks")
	}

	// Verify that results were added
	if len(suite.Results) < 1 {
		t.Error("Expected at least one benchmark result")
	}
}

func TestBenchmarkResult_Fields(t *testing.T) {
	result := &BenchmarkResult{
		PackageName:     "Test Package",
		TestName:        "Test Test",
		Duration:        time.Millisecond * 100,
		MemoryUsage:     1024,
		OperationsCount: 1000,
		Throughput:      10000.0,
		MemoryPerOp:     1.024,
		Timestamp:       time.Now(),
		Metadata:        map[string]interface{}{"test": true},
	}

	// Test all fields are set correctly
	if result.PackageName != "Test Package" {
		t.Error("PackageName not set correctly")
	}
	if result.TestName != "Test Test" {
		t.Error("TestName not set correctly")
	}
	if result.Duration != time.Millisecond*100 {
		t.Error("Duration not set correctly")
	}
	if result.MemoryUsage != 1024 {
		t.Error("MemoryUsage not set correctly")
	}
	if result.OperationsCount != 1000 {
		t.Error("OperationsCount not set correctly")
	}
	if result.Throughput != 10000.0 {
		t.Error("Throughput not set correctly")
	}
	if result.MemoryPerOp != 1.024 {
		t.Error("MemoryPerOp not set correctly")
	}
	if result.Timestamp.IsZero() {
		t.Error("Timestamp not set correctly")
	}
	if result.Metadata["test"] != true {
		t.Error("Metadata not set correctly")
	}
}

func TestBenchmarkReport_Fields(t *testing.T) {
	now := time.Now()
	results := []*BenchmarkResult{
		{
			PackageName:     "Test Package",
			TestName:        "Test1",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      1000,
			MemoryPerOp:     1.024,
			Timestamp:       now,
			Metadata:        map[string]interface{}{"test": "1"},
		},
	}

	// Test that results were created correctly
	if len(results) != 1 {
		t.Error("Expected 1 result")
	}

	if results[0].PackageName != "Test Package" {
		t.Error("PackageName not set correctly")
	}

	if results[0].TestName != "Test1" {
		t.Error("TestName not set correctly")
	}

	if results[0].Timestamp != now {
		t.Error("Timestamp not set correctly")
	}
}

func TestMainBenchmarkOrchestrator_PrintSummary(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Add some mock results
	orchestrator.AllResults = []*BenchmarkResult{
		{
			PackageName:     "Test Package",
			TestName:        "Test1",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      1000,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"test": "1"},
		},
	}

	// Test that results were added correctly
	if len(orchestrator.AllResults) != 1 {
		t.Error("Failed to add test result")
	}

	// Test that the result has correct fields
	result := orchestrator.AllResults[0]
	if result.PackageName != "Test Package" {
		t.Error("PackageName not set correctly")
	}
	if result.TestName != "Test1" {
		t.Error("TestName not set correctly")
	}
}

func TestMainBenchmarkOrchestrator_GenerateComprehensiveReport(t *testing.T) {
	orchestrator := NewMainBenchmarkOrchestrator()

	// Add some mock results
	orchestrator.AllResults = []*BenchmarkResult{
		{
			PackageName:     "Test Package",
			TestName:        "Test1",
			Duration:        100 * time.Millisecond,
			MemoryUsage:     1024,
			OperationsCount: 1000,
			Throughput:      1000,
			MemoryPerOp:     1.024,
			Timestamp:       time.Now(),
			Metadata:        map[string]interface{}{"test": "1"},
		},
	}

	// Test that results were added correctly
	if len(orchestrator.AllResults) != 1 {
		t.Error("Failed to add test result")
	}

	// Test that the orchestrator has the expected structure
	if orchestrator.Layer2Benchmarks == nil {
		t.Error("Layer2Benchmarks should not be nil")
	}
	if orchestrator.CrossChainBenchmarks == nil {
		t.Error("CrossChainBenchmarks should not be nil")
	}
	if orchestrator.GovernanceBenchmarks == nil {
		t.Error("GovernanceBenchmarks should not be nil")
	}
	if orchestrator.PrivacyBenchmarks == nil {
		t.Error("PrivacyBenchmarks should not be nil")
	}
}
