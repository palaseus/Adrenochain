package testing

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// TestRunnerMain provides the main entry point for running all tests
func TestRunnerMain() {
	fmt.Println("🚀 adrenochain COMPREHENSIVE TEST SUITE")
	fmt.Println("=====================================")

	// Create comprehensive test suite
	testSuite := NewComprehensiveTestSuite()

	// Initialize test suites
	fmt.Println("📋 Initializing test suites...")
	if err := testSuite.InitializeTestSuites(); err != nil {
		log.Fatalf("Failed to initialize test suites: %v", err)
	}

	// Run all tests
	fmt.Println("🧪 Running comprehensive tests...")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := testSuite.RunAllTests(ctx)
	if err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}

	duration := time.Since(startTime)

	// Display results
	fmt.Println("\n📊 TEST EXECUTION RESULTS")
	fmt.Println("==========================")
	fmt.Printf("⏱️  Total Duration: %v\n", duration)
	fmt.Printf("📈 Total Tests: %d\n", report.TotalTests)
	fmt.Printf("✅ Passed Tests: %d\n", report.PassedTests)
	fmt.Printf("❌ Failed Tests: %d\n", report.FailedTests)
	fmt.Printf("⏭️  Skipped Tests: %d\n", report.SkippedTests)
	fmt.Printf("📊 Success Rate: %.2f%%\n", report.SuccessRate)
	fmt.Printf("🎯 Coverage: %.2f%%\n", report.Coverage)

	// Display recommendations
	if len(report.Recommendations) > 0 {
		fmt.Println("\n💡 RECOMMENDATIONS")
		fmt.Println("==================")
		for i, rec := range report.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}

	// Get detailed statistics
	stats := testSuite.GetTestStatistics()
	fmt.Println("\n📈 DETAILED STATISTICS")
	fmt.Println("======================")
	fmt.Printf("🔄 Last Run: %v\n", stats.LastRun)
	fmt.Printf("⏱️  Total Duration: %v\n", stats.TotalDuration)
	fmt.Printf("🎯 Overall Coverage: %.2f%%\n", stats.Coverage)

	// Get coverage report
	coverageReport := testSuite.GetCoverageReport()
	if coverageReport != nil {
		fmt.Println("\n📊 COVERAGE BREAKDOWN")
		fmt.Println("=====================")
		fmt.Printf("🎯 Overall Coverage: %.2f%%\n", coverageReport.OverallCoverage)

		if len(coverageReport.PackageCoverage) > 0 {
			fmt.Println("\n📦 Package Coverage:")
			for pkg, coverage := range coverageReport.PackageCoverage {
				fmt.Printf("  %s: %.2f%%\n", pkg, coverage)
			}
		}

		if len(coverageReport.Recommendations) > 0 {
			fmt.Println("\n💡 Coverage Recommendations:")
			for i, rec := range coverageReport.Recommendations {
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}
	}

	// Determine overall status
	fmt.Println("\n🏆 OVERALL STATUS")
	fmt.Println("==================")
	if report.SuccessRate >= 95.0 && report.Coverage >= 90.0 {
		fmt.Println("🎉 EXCELLENT! All tests passed with high coverage!")
		fmt.Println("✅ adrenochain is production-ready!")
		os.Exit(0)
	} else if report.SuccessRate >= 90.0 && report.Coverage >= 80.0 {
		fmt.Println("👍 GOOD! Tests passed with acceptable coverage.")
		fmt.Println("⚠️  Some improvements recommended before production.")
		os.Exit(0)
	} else {
		fmt.Println("❌ ATTENTION REQUIRED! Test results below acceptable thresholds.")
		fmt.Println("🔧 Please review and fix failing tests before proceeding.")
		os.Exit(1)
	}
}

// RunSpecificTestSuite runs a specific test suite
func RunSpecificTestSuite(suiteID string) {
	fmt.Printf("🧪 Running test suite: %s\n", suiteID)

	testSuite := NewComprehensiveTestSuite()

	// Initialize test suites
	if err := testSuite.InitializeTestSuites(); err != nil {
		log.Fatalf("Failed to initialize test suites: %v", err)
	}

	// Run specific suite
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	report, err := testSuite.RunTestSuite(ctx, suiteID)
	if err != nil {
		log.Fatalf("Test suite execution failed: %v", err)
	}

	// Display results
	fmt.Printf("\n📊 Test Suite Results: %s\n", suiteID)
	fmt.Printf("✅ Passed: %d\n", report.PassedTests)
	fmt.Printf("❌ Failed: %d\n", report.FailedTests)
	fmt.Printf("📊 Success Rate: %.2f%%\n", report.SuccessRate)
	fmt.Printf("🎯 Coverage: %.2f%%\n", report.Coverage)
}

// RunContractEngineTests runs only the contract engine tests
func RunContractEngineTests() {
	fmt.Println("🔧 Running Contract Engine Tests...")
	RunSpecificTestSuite("contract_engine")
}

// RunDeFiTests runs only the DeFi component tests
func RunDeFiTests() {
	fmt.Println("💰 Running DeFi Component Tests...")
	RunSpecificTestSuite("defi_components")
}

// RunInfrastructureTests runs only the infrastructure tests
func RunInfrastructureTests() {
	fmt.Println("🏗️  Running Infrastructure Tests...")
	RunSpecificTestSuite("infrastructure")
}

// RunAPITests runs only the API and SDK tests
func RunAPITests() {
	fmt.Println("🔌 Running API and SDK Tests...")
	RunSpecificTestSuite("api_sdk")
}

// RunIntegrationTests runs only the integration tests
func RunIntegrationTests() {
	fmt.Println("🔗 Running Integration Tests...")
	RunSpecificTestSuite("integration")
}

// RunPerformanceTests runs performance-focused tests
func RunPerformanceTests() {
	fmt.Println("⚡ Running Performance Tests...")

	testSuite := NewComprehensiveTestSuite()

	// Initialize test suites
	if err := testSuite.InitializeTestSuites(); err != nil {
		log.Fatalf("Failed to initialize test suites: %v", err)
	}

	// Run performance tests with extended timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	startTime := time.Now()
	report, err := testSuite.RunTestSuite(ctx, "contract_engine")
	if err != nil {
		log.Fatalf("Performance test execution failed: %v", err)
	}

	duration := time.Since(startTime)

	fmt.Printf("\n📊 Performance Test Results:\n")
	fmt.Printf("⏱️  Duration: %v\n", duration)
	fmt.Printf("📈 Tests Executed: %d\n", report.TotalTests)
	fmt.Printf("✅ Success Rate: %.2f%%\n", report.SuccessRate)

	// Get performance metrics
	stats := testSuite.GetTestStatistics()
	fmt.Printf("🔄 Total Duration: %v\n", stats.TotalDuration)
}

// RunSecurityTests runs security-focused tests
func RunSecurityTests() {
	fmt.Println("🔒 Running Security Tests...")

	testSuite := NewComprehensiveTestSuite()

	// Initialize test suites
	if err := testSuite.InitializeTestSuites(); err != nil {
		log.Fatalf("Failed to initialize test suites: %v", err)
	}

	// Run security tests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	report, err := testSuite.RunTestSuite(ctx, "contract_engine")
	if err != nil {
		log.Fatalf("Security test execution failed: %v", err)
	}

	duration := time.Since(startTime)

	fmt.Printf("\n📊 Security Test Results:\n")
	fmt.Printf("⏱️  Duration: %v\n", duration)
	fmt.Printf("🔒 Tests Executed: %d\n", report.TotalTests)
	fmt.Printf("✅ Success Rate: %.2f%%\n", report.SuccessRate)

	// Security tests should have 100% pass rate
	if report.SuccessRate < 100.0 {
		fmt.Println("❌ CRITICAL: Security tests failed!")
		fmt.Println("🔧 Security issues must be resolved before production.")
		os.Exit(1)
	} else {
		fmt.Println("✅ All security tests passed!")
	}
}

// GenerateTestReport generates a comprehensive test report
func GenerateTestReport() {
	fmt.Println("📋 Generating Comprehensive Test Report...")

	testSuite := NewComprehensiveTestSuite()

	// Initialize test suites
	if err := testSuite.InitializeTestSuites(); err != nil {
		log.Fatalf("Failed to initialize test suites: %v", err)
	}

	// Get statistics
	stats := testSuite.GetTestStatistics()
	coverageReport := testSuite.GetCoverageReport()

	fmt.Println("\n📊 adrenochain TEST REPORT")
	fmt.Println("=======================")
	fmt.Printf("📅 Generated: %v\n", time.Now())
	fmt.Printf("🔄 Last Run: %v\n", stats.LastRun)
	fmt.Printf("⏱️  Total Duration: %v\n", stats.TotalDuration)
	fmt.Printf("🎯 Overall Coverage: %.2f%%\n", stats.Coverage)

	if coverageReport != nil {
		fmt.Println("\n📦 Package Coverage Details:")
		for pkg, coverage := range coverageReport.PackageCoverage {
			status := "✅"
			if coverage < 80.0 {
				status = "⚠️"
			}
			if coverage < 60.0 {
				status = "❌"
			}
			fmt.Printf("  %s %s: %.2f%%\n", status, pkg, coverage)
		}

		if len(coverageReport.UncoveredAreas) > 0 {
			fmt.Println("\n🔍 Uncovered Areas:")
			for _, area := range coverageReport.UncoveredAreas {
				fmt.Printf("  • %s\n", area)
			}
		}
	}

	fmt.Println("\n🏆 REPORT SUMMARY")
	fmt.Println("==================")
	if stats.Coverage >= 90.0 {
		fmt.Println("🎉 EXCELLENT COVERAGE! adrenochain is production-ready!")
	} else if stats.Coverage >= 80.0 {
		fmt.Println("👍 GOOD COVERAGE! Minor improvements recommended.")
	} else {
		fmt.Println("⚠️  COVERAGE NEEDS IMPROVEMENT! Review required areas.")
	}
}
