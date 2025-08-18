package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/palaseus/adrenochain/pkg/security"
)

func main() {
	fmt.Println("🔒 Adrenochain Comprehensive Security Validation Suite")
	fmt.Println(strings.Repeat("=", 60))

	// Create the real security validator
	validator := security.NewRealSecurityValidator()

		// Run all real security validations
	fmt.Println("Starting comprehensive real security validation...")
	if err := validator.RunAllRealSecurityValidations(); err != nil {
		log.Fatalf("Security validation failed: %v", err)
	}
	
	// Generate security report
	fmt.Println("\nGenerating comprehensive security report...")
	if err := generateSecurityReport(validator); err != nil {
		log.Fatalf("Report generation failed: %v", err)
	}

	fmt.Println("\n🎉 Security validation completed successfully!")
	fmt.Println("Check the generated JSON report for detailed results.")
	os.Exit(0)
}

// generateSecurityReport creates a comprehensive security validation report
func generateSecurityReport(validator *security.RealSecurityValidator) error {
	results := validator.GetResults()

	// Calculate summary statistics
	summary := calculateSecuritySummary(results)

	// Create report structure
	report := SecurityReport{
		Summary:     summary,
		Results:     results,
		GeneratedAt: time.Now(),
	}

	// Save report to file
	if err := saveSecurityReportToFile(report); err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	// Print summary to console
	printSecuritySummaryToConsole(summary)

	return nil
}

// SecurityReport represents a comprehensive security validation report
type SecurityReport struct {
	Summary     SecuritySummary                      `json:"summary"`
	Results     []*security.SecurityValidationResult `json:"results"`
	GeneratedAt time.Time                            `json:"generated_at"`
}

// SecuritySummary provides summary statistics for all security validations
type SecuritySummary struct {
	TotalTests        int                             `json:"total_tests"`
	PassedTests       int                             `json:"passed_tests"`
	FailedTests       int                             `json:"failed_tests"`
	WarningTests      int                             `json:"warning_tests"`
	TotalIssues       int                             `json:"total_issues"`
	CriticalIssues    int                             `json:"critical_issues"`
	TotalWarnings     int                             `json:"total_warnings"`
	PackageBreakdown  map[string]PackageSecurityStats `json:"package_breakdown"`
	TestTypeBreakdown map[string]TestTypeStats        `json:"test_type_breakdown"`
}

// PackageSecurityStats provides security statistics for a specific package
type PackageSecurityStats struct {
	TestCount      int `json:"test_count"`
	PassedTests    int `json:"passed_tests"`
	FailedTests    int `json:"failed_tests"`
	WarningTests   int `json:"warning_tests"`
	TotalIssues    int `json:"total_issues"`
	CriticalIssues int `json:"critical_issues"`
	TotalWarnings  int `json:"total_warnings"`
}

// TestTypeStats provides statistics for a specific test type
type TestTypeStats struct {
	TestCount      int `json:"test_count"`
	PassedTests    int `json:"passed_tests"`
	FailedTests    int `json:"failed_tests"`
	WarningTests   int `json:"warning_tests"`
	TotalIssues    int `json:"total_issues"`
	CriticalIssues int `json:"critical_issues"`
	TotalWarnings  int `json:"total_warnings"`
}

// calculateSecuritySummary calculates comprehensive security summary statistics
func calculateSecuritySummary(results []*security.SecurityValidationResult) SecuritySummary {
	summary := SecuritySummary{
		PackageBreakdown:  make(map[string]PackageSecurityStats),
		TestTypeBreakdown: make(map[string]TestTypeStats),
	}

	for _, result := range results {
		summary.TotalTests++

		// Update overall counts
		if result.Status == "PASS" {
			summary.PassedTests++
		} else if result.Status == "FAIL" {
			summary.FailedTests++
		} else if result.Status == "WARNING" {
			summary.WarningTests++
		}

		summary.TotalIssues += result.IssuesFound
		summary.CriticalIssues += result.CriticalIssues
		summary.TotalWarnings += result.Warnings

		// Update package breakdown
		if stats, exists := summary.PackageBreakdown[result.PackageName]; exists {
			stats.TestCount++
			if result.Status == "PASS" {
				stats.PassedTests++
			} else if result.Status == "FAIL" {
				stats.FailedTests++
			} else if result.Status == "WARNING" {
				stats.WarningTests++
			}
			stats.TotalIssues += result.IssuesFound
			stats.CriticalIssues += result.CriticalIssues
			stats.TotalWarnings += result.Warnings
			summary.PackageBreakdown[result.PackageName] = stats
		} else {
			stats := PackageSecurityStats{TestCount: 1}
			if result.Status == "PASS" {
				stats.PassedTests = 1
			} else if result.Status == "FAIL" {
				stats.FailedTests = 1
			} else if result.Status == "WARNING" {
				stats.WarningTests = 1
			}
			stats.TotalIssues = result.IssuesFound
			stats.CriticalIssues = result.CriticalIssues
			stats.TotalWarnings = result.Warnings
			summary.PackageBreakdown[result.PackageName] = stats
		}

		// Update test type breakdown
		if stats, exists := summary.TestTypeBreakdown[result.TestType]; exists {
			stats.TestCount++
			if result.Status == "PASS" {
				stats.PassedTests++
			} else if result.Status == "FAIL" {
				stats.FailedTests++
			} else if result.Status == "WARNING" {
				stats.WarningTests++
			}
			stats.TotalIssues += result.IssuesFound
			stats.CriticalIssues += result.CriticalIssues
			stats.TotalWarnings += result.Warnings
			summary.TestTypeBreakdown[result.TestType] = stats
		} else {
			stats := TestTypeStats{TestCount: 1}
			if result.Status == "PASS" {
				stats.PassedTests = 1
			} else if result.Status == "FAIL" {
				stats.FailedTests = 1
			} else if result.Status == "WARNING" {
				stats.WarningTests = 1
			}
			stats.TotalIssues = result.IssuesFound
			stats.CriticalIssues = result.CriticalIssues
			stats.TotalWarnings = result.Warnings
			summary.TestTypeBreakdown[result.TestType] = stats
		}
	}

	return summary
}

// saveSecurityReportToFile saves the security report to a JSON file
func saveSecurityReportToFile(report SecurityReport) error {
	filename := fmt.Sprintf("security_report_%s.json", time.Now().Format("20060102_150405"))

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

	fmt.Printf("📄 Security report saved to: %s\n", filename)
	return nil
}

// printSecuritySummaryToConsole prints a summary of security validation results to the console
func printSecuritySummaryToConsole(summary SecuritySummary) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔒 SECURITY VALIDATION SUMMARY REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Tests: %d\n", summary.TotalTests)
	fmt.Printf("Passed Tests: %d\n", summary.PassedTests)
	fmt.Printf("Failed Tests: %d\n", summary.FailedTests)
	fmt.Printf("Warning Tests: %d\n", summary.WarningTests)
	fmt.Printf("Total Issues: %d\n", summary.TotalIssues)
	fmt.Printf("Critical Issues: %d\n", summary.CriticalIssues)
	fmt.Printf("Total Warnings: %d\n", summary.TotalWarnings)

	fmt.Println("\n📦 Package Security Breakdown:")
	for packageName, stats := range summary.PackageBreakdown {
		status := "✅"
		if stats.FailedTests > 0 {
			status = "❌"
		} else if stats.WarningTests > 0 {
			status = "⚠️"
		}
		fmt.Printf("  %s %s: %d tests, %d issues, %d critical\n",
			status, packageName, stats.TestCount, stats.TotalIssues, stats.CriticalIssues)
	}

	fmt.Println("\n🧪 Test Type Breakdown:")
	for testType, stats := range summary.TestTypeBreakdown {
		status := "✅"
		if stats.FailedTests > 0 {
			status = "❌"
		} else if stats.WarningTests > 0 {
			status = "⚠️"
		}
		fmt.Printf("  %s %s: %d tests, %d issues, %d critical\n",
			status, testType, stats.TestCount, stats.TotalIssues, stats.CriticalIssues)
	}
}
