package main

import (
	"fmt"
	"log"
	"os"

	"github.com/palaseus/adrenochain/pkg/benchmarking"
)

func main() {
	fmt.Println("🚀 ADRENOCHAIN BENCHMARKING SUITE")
	fmt.Println("==================================")

	// Create benchmark orchestrator
	orchestrator := benchmarking.NewMainBenchmarkOrchestrator()

	// Run all benchmarks
	if err := orchestrator.RunAllBenchmarks(); err != nil {
		log.Printf("Benchmarking failed: %v", err)
		os.Exit(1)
	}

	// Save benchmark report
	if err := orchestrator.SaveReportToFile(); err != nil {
		log.Printf("Failed to save benchmark report: %v", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ All benchmarks completed successfully!")
}
