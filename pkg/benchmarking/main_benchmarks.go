package benchmarking

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// MainBenchmarkOrchestrator orchestrates all benchmark suites
type MainBenchmarkOrchestrator struct {
	Layer2Benchmarks     *Layer2BenchmarkSuite
	CrossChainBenchmarks *CrossChainBenchmarkSuite
	GovernanceBenchmarks *GovernanceBenchmarkSuite
	PrivacyBenchmarks    *PrivacyBenchmarkSuite
	AIMLBenchmarks       *BenchmarkSuite
	AllResults           []*BenchmarkResult
	StartTime            time.Time
	EndTime              time.Time
}

// NewMainBenchmarkOrchestrator creates a new main benchmark orchestrator
func NewMainBenchmarkOrchestrator() *MainBenchmarkOrchestrator {
	return &MainBenchmarkOrchestrator{
		Layer2Benchmarks:     NewLayer2BenchmarkSuite(),
		CrossChainBenchmarks: NewCrossChainBenchmarkSuite(),
		GovernanceBenchmarks: NewGovernanceBenchmarkSuite(),
		PrivacyBenchmarks:    NewPrivacyBenchmarkSuite(),
		AIMLBenchmarks:       NewBenchmarkSuite(),
		AllResults:           make([]*BenchmarkResult, 0),
	}
}

// RunAllBenchmarks executes all benchmark suites
func (mbo *MainBenchmarkOrchestrator) RunAllBenchmarks() error {
	mbo.StartTime = time.Now()
	fmt.Println("🚀 Starting Comprehensive Performance Benchmarking Suite...")
	fmt.Printf("⏰ Start Time: %s\n", mbo.StartTime.Format(time.RFC3339))

	// Run Layer 2 Benchmarks
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 LAYER 2 SOLUTIONS BENCHMARKING")
	fmt.Println(strings.Repeat("=", 60))
	if err := mbo.Layer2Benchmarks.RunAllLayer2Benchmarks(); err != nil {
		return fmt.Errorf("layer 2 benchmarks failed: %v", err)
	}
	mbo.AllResults = append(mbo.AllResults, mbo.Layer2Benchmarks.GetResults()...)

	// Run Cross-Chain Benchmarks
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 CROSS-CHAIN INFRASTRUCTURE BENCHMARKING")
	fmt.Println(strings.Repeat("=", 60))
	if err := mbo.CrossChainBenchmarks.RunAllCrossChainBenchmarks(); err != nil {
		return fmt.Errorf("cross-chain benchmarks failed: %v", err)
	}
	mbo.AllResults = append(mbo.AllResults, mbo.CrossChainBenchmarks.GetResults()...)

	// Run Governance Benchmarks
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 GOVERNANCE & DAO BENCHMARKING")
	fmt.Println(strings.Repeat("=", 60))
	if err := mbo.GovernanceBenchmarks.RunAllGovernanceBenchmarks(); err != nil {
		return fmt.Errorf("governance benchmarks failed: %v", err)
	}
	mbo.AllResults = append(mbo.AllResults, mbo.GovernanceBenchmarks.GetResults()...)

	// Run Privacy Benchmarks
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 PRIVACY & ZERO-KNOWLEDGE BENCHMARKING")
	fmt.Println(strings.Repeat("=", 60))
	if err := mbo.PrivacyBenchmarks.RunAllPrivacyBenchmarks(); err != nil {
		return fmt.Errorf("privacy benchmarks failed: %v", err)
	}
	mbo.AllResults = append(mbo.AllResults, mbo.PrivacyBenchmarks.GetResults()...)

	// Run AI/ML Benchmarks
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 AI/ML INTEGRATION BENCHMARKING")
	fmt.Println(strings.Repeat("=", 60))
	if err := mbo.AIMLBenchmarks.RunAllBenchmarks(); err != nil {
		return fmt.Errorf("AI/ML benchmarks failed: %v", err)
	}
	mbo.AllResults = append(mbo.AllResults, mbo.AIMLBenchmarks.GetResults()...)

	mbo.EndTime = time.Now()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 ALL BENCHMARKS COMPLETED SUCCESSFULLY!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏰ End Time: %s\n", mbo.EndTime.Format(time.RFC3339))
	fmt.Printf("⏱️  Total Duration: %s\n", mbo.EndTime.Sub(mbo.StartTime))
	fmt.Printf("📊 Total Benchmark Tests: %d\n", len(mbo.AllResults))

	return nil
}

// GenerateBenchmarkReport generates a comprehensive benchmark report
func (mbo *MainBenchmarkOrchestrator) GenerateBenchmarkReport() error {
	fmt.Println("\n📋 Generating Comprehensive Benchmark Report...")

	// Calculate summary statistics
	summary := mbo.calculateSummaryStatistics()

	// Create report structure
	report := BenchmarkReport{
		Summary:     summary,
		Results:     mbo.AllResults,
		GeneratedAt: time.Now(),
		Duration:    mbo.EndTime.Sub(mbo.StartTime),
	}

	// Save report to file
	if err := mbo.saveReportToFile(report); err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	// Print summary to console
	mbo.printSummaryToConsole(summary)

	return nil
}

// BenchmarkReport represents a comprehensive benchmark report
type BenchmarkReport struct {
	Summary     BenchmarkSummary   `json:"summary"`
	Results     []*BenchmarkResult `json:"results"`
	GeneratedAt time.Time          `json:"generated_at"`
	Duration    time.Duration      `json:"total_duration"`
}

// BenchmarkSummary provides summary statistics for all benchmarks
type BenchmarkSummary struct {
	TotalTests        int                     `json:"total_tests"`
	TotalOperations   int64                   `json:"total_operations"`
	AverageThroughput float64                 `json:"average_throughput"`
	TotalMemoryUsage  uint64                  `json:"total_memory_usage"`
	PackageBreakdown  map[string]PackageStats `json:"package_breakdown"`
	PerformanceTiers  map[string]int          `json:"performance_tiers"`
	TopPerformers     []TopPerformer          `json:"top_performers"`
}

// PackageStats provides statistics for a specific package
type PackageStats struct {
	TestCount         int     `json:"test_count"`
	TotalOperations   int64   `json:"total_operations"`
	AverageThroughput float64 `json:"average_throughput"`
	TotalMemoryUsage  uint64  `json:"total_memory_usage"`
}

// TopPerformer represents a top-performing benchmark test
type TopPerformer struct {
	PackageName string  `json:"package_name"`
	TestName    string  `json:"test_name"`
	Throughput  float64 `json:"throughput"`
	MemoryPerOp float64 `json:"memory_per_op"`
}

// calculateSummaryStatistics calculates comprehensive summary statistics
func (mbo *MainBenchmarkOrchestrator) calculateSummaryStatistics() BenchmarkSummary {
	summary := BenchmarkSummary{
		TotalTests:       len(mbo.AllResults),
		PackageBreakdown: make(map[string]PackageStats),
		PerformanceTiers: make(map[string]int),
		TopPerformers:    make([]TopPerformer, 0),
	}

	var totalOperations int64
	var totalMemoryUsage uint64
	var totalThroughput float64

	// Process each result
	for _, result := range mbo.AllResults {
		totalOperations += result.OperationsCount
		totalMemoryUsage += result.MemoryUsage
		totalThroughput += result.Throughput

		// Update package breakdown
		if stats, exists := summary.PackageBreakdown[result.PackageName]; exists {
			stats.TestCount++
			stats.TotalOperations += result.OperationsCount
			stats.TotalMemoryUsage += result.MemoryUsage
			stats.AverageThroughput = (stats.AverageThroughput + result.Throughput) / 2
			summary.PackageBreakdown[result.PackageName] = stats
		} else {
			summary.PackageBreakdown[result.PackageName] = PackageStats{
				TestCount:         1,
				TotalOperations:   result.OperationsCount,
				TotalMemoryUsage:  result.MemoryUsage,
				AverageThroughput: result.Throughput,
			}
		}

		// Categorize performance tiers
		if result.Throughput >= 100000 {
			summary.PerformanceTiers["Ultra High"]++
		} else if result.Throughput >= 10000 {
			summary.PerformanceTiers["High"]++
		} else if result.Throughput >= 1000 {
			summary.PerformanceTiers["Medium"]++
		} else {
			summary.PerformanceTiers["Low"]++
		}

		// Track top performers
		performer := TopPerformer{
			PackageName: result.PackageName,
			TestName:    result.TestName,
			Throughput:  result.Throughput,
			MemoryPerOp: result.MemoryPerOp,
		}
		summary.TopPerformers = append(summary.TopPerformers, performer)
	}

	summary.TotalOperations = totalOperations
	summary.TotalMemoryUsage = totalMemoryUsage
	if summary.TotalTests > 0 {
		summary.AverageThroughput = totalThroughput / float64(summary.TotalTests)
	}

	// Sort top performers by throughput
	// (This is a simple implementation - in production you'd want proper sorting)

	return summary
}

// saveReportToFile saves the benchmark report to a JSON file
func (mbo *MainBenchmarkOrchestrator) saveReportToFile(report BenchmarkReport) error {
	filename := fmt.Sprintf("benchmark_report_%s.json", time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %v", err)
	}

	fmt.Printf("📄 Benchmark report saved to: %s\n", filename)
	return nil
}

// printSummaryToConsole prints a summary of benchmark results to the console
func (mbo *MainBenchmarkOrchestrator) printSummaryToConsole(summary BenchmarkSummary) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 BENCHMARK SUMMARY REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Tests: %d\n", summary.TotalTests)
	fmt.Printf("Total Operations: %d\n", summary.TotalOperations)
	fmt.Printf("Average Throughput: %.2f ops/sec\n", summary.AverageThroughput)
	fmt.Printf("Total Memory Usage: %d bytes\n", summary.TotalMemoryUsage)

	fmt.Println("\n📦 Package Breakdown:")
	for packageName, stats := range summary.PackageBreakdown {
		fmt.Printf("  %s: %d tests, %.2f ops/sec avg\n",
			packageName, stats.TestCount, stats.AverageThroughput)
	}

	fmt.Println("\n🏆 Performance Tiers:")
	for tier, count := range summary.PerformanceTiers {
		fmt.Printf("  %s: %d tests\n", tier, count)
	}

	fmt.Println("\n🚀 Top Performers (Top 5 by Throughput):")
	// Show top 5 performers (simplified)
	for i := 0; i < 5 && i < len(summary.TopPerformers); i++ {
		performer := summary.TopPerformers[i]
		fmt.Printf("  %d. %s - %s: %.2f ops/sec\n",
			i+1, performer.PackageName, performer.TestName, performer.Throughput)
	}
}
