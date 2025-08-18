package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/palaseus/adrenochain/pkg/benchmarking"
)

func main() {
	fmt.Println("🚀 Adrenochain Comprehensive Performance Benchmarking Suite")
	fmt.Println(strings.Repeat("=", 60))

	// Create the main benchmark orchestrator
	orchestrator := benchmarking.NewMainBenchmarkOrchestrator()

	// Run all benchmarks
	fmt.Println("Starting comprehensive performance benchmarking...")
	if err := orchestrator.RunAllBenchmarks(); err != nil {
		log.Fatalf("Benchmarking failed: %v", err)
	}

	// Generate comprehensive report
	fmt.Println("\nGenerating comprehensive benchmark report...")
	if err := orchestrator.GenerateBenchmarkReport(); err != nil {
		log.Fatalf("Report generation failed: %v", err)
	}

	fmt.Println("\n🎉 Benchmarking completed successfully!")
	fmt.Println("Check the generated JSON report for detailed results.")
	os.Exit(0)
}
