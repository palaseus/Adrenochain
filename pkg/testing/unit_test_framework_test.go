package testing

import (
	"fmt"
	"testing"
	"time"
)

func TestNewUnitTestFramework(t *testing.T) {
	config := UnitTestConfig{
		MaxConcurrentTests:       10,
		TestTimeout:              30 * time.Second,
		EnableParallel:           true,
		EnableRaceDetection:      false,
		MinCoverageThreshold:     80.0,
		EnableCoverageReport:     true,
		CoverageOutputFormat:     "html",
		EnableAutoGeneration:     false,
		MaxGeneratedTests:        100,
		TestDataSeed:             12345,
		EnableDetailedReports:    true,
		EnablePerformanceProfiling: true,
		ReportOutputPath:         "./reports",
	}

	framework := NewUnitTestFramework(config)

	if framework == nil {
		t.Fatal("NewUnitTestFramework returned nil")
	}

	if framework.config.MaxConcurrentTests != 10 {
		t.Error("MaxConcurrentTests not set correctly")
	}

	if framework.config.TestTimeout != 30*time.Second {
		t.Error("TestTimeout not set correctly")
	}

	if !framework.config.EnableParallel {
		t.Error("EnableParallel not set correctly")
	}

	if framework.config.EnableRaceDetection {
		t.Error("EnableRaceDetection not set correctly")
	}

	if framework.config.MinCoverageThreshold != 80.0 {
		t.Error("MinCoverageThreshold not set correctly")
	}

	if !framework.config.EnableCoverageReport {
		t.Error("EnableCoverageReport not set correctly")
	}

	if framework.config.CoverageOutputFormat != "html" {
		t.Error("CoverageOutputFormat not set correctly")
	}

	if framework.config.EnableAutoGeneration {
		t.Error("EnableAutoGeneration not set correctly")
	}

	if framework.config.MaxGeneratedTests != 100 {
		t.Error("MaxGeneratedTests not set correctly")
	}

	if framework.config.TestDataSeed != 12345 {
		t.Error("TestDataSeed not set correctly")
	}

	if !framework.config.EnableDetailedReports {
		t.Error("EnableDetailedReports not set correctly")
	}

	if !framework.config.EnablePerformanceProfiling {
		t.Error("EnablePerformanceProfiling not set correctly")
	}

	if framework.config.ReportOutputPath != "./reports" {
		t.Error("ReportOutputPath not set correctly")
	}

	if len(framework.testSuites) != 0 {
		t.Error("TestSuites should be empty initially")
	}

	if len(framework.testCases) != 0 {
		t.Error("TestCases should be empty initially")
	}

	if framework.TotalTests != 0 {
		t.Error("TotalTests should be 0 initially")
	}

	if framework.PassedTests != 0 {
		t.Error("PassedTests should be 0 initially")
	}

	if framework.FailedTests != 0 {
		t.Error("FailedTests should be 0 initially")
	}

	if framework.SkippedTests != 0 {
		t.Error("SkippedTests should be 0 initially")
	}
}

func TestUnitTestFramework_RegisterTestSuite(t *testing.T) {
	framework := NewUnitTestFramework(UnitTestConfig{})

	testSuite := &TestSuite{
		ID:          "suite1",
		Name:        "Test Suite 1",
		Description: "A test suite for testing",
		TestCases:   []*TestCase{},
		Setup:       func() error { return nil },
		Teardown:    func() error { return nil },
		Metadata:    map[string]interface{}{"priority": "high"},
	}

	err := framework.RegisterTestSuite(testSuite)
	if err != nil {
		t.Fatalf("RegisterTestSuite failed: %v", err)
	}

	if len(framework.testSuites) != 1 {
		t.Error("TestSuite not added")
	}

	if framework.testSuites["suite1"] != testSuite {
		t.Error("TestSuite not stored correctly")
	}

	// Try to add the same suite again
	err = framework.RegisterTestSuite(testSuite)
	if err != ErrTestSuiteAlreadyExists {
		t.Errorf("Expected ErrTestSuiteAlreadyExists, got %v", err)
	}
}

func TestUnitTestFramework_AddTestCase(t *testing.T) {
	framework := NewUnitTestFramework(UnitTestConfig{})

	testCase := &TestCase{
		ID:          "test1",
		Name:        "Test Case 1",
		Description: "A test case for testing",
		SuiteID:     "suite1",
		Function:    func(t interface{}) error { return nil },
		Status:      TestStatusPending,
		Priority:    TestPriorityNormal,
	}

	err := framework.AddTestCase(testCase)
	if err != nil {
		t.Fatalf("AddTestCase failed: %v", err)
	}

	if len(framework.testCases) != 1 {
		t.Error("TestCase not added")
	}

	if framework.testCases["test1"] != testCase {
		t.Error("TestCase not stored correctly")
	}

	if framework.TotalTests != 1 {
		t.Error("TotalTests not incremented")
	}

	// Try to add the same test case again
	err = framework.AddTestCase(testCase)
	if err == nil {
		t.Error("Expected error when adding duplicate test case")
	}
}

func TestUnitTestFramework_GetCoverageReport(t *testing.T) {
	framework := NewUnitTestFramework(UnitTestConfig{})

	// Get coverage report
	report := framework.GetCoverageReport()

	if report == nil {
		t.Error("Coverage report should not be nil")
	}
}

func TestUnitTestFramework_GetTestStatistics(t *testing.T) {
	framework := NewUnitTestFramework(UnitTestConfig{})

	// Initially should have zero stats
	stats := framework.GetTestStatistics()
	if stats.TotalTests != 0 {
		t.Error("Should have 0 total tests initially")
	}

	// Add a test case
	testCase := &TestCase{
		ID:          "test1",
		Name:        "Test Case 1",
		Description: "A test case for testing",
		Function:    func(t interface{}) error { return nil },
	}

	err := framework.AddTestCase(testCase)
	if err != nil {
		t.Fatalf("AddTestCase failed: %v", err)
	}

	// Update framework stats
	framework.TotalTests = 1
	framework.PassedTests = 1

	// Now should have stats
	stats = framework.GetTestStatistics()
	if stats.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", stats.TotalTests)
	}

	if stats.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", stats.PassedTests)
	}
}

func TestUnitTestFramework_TestStatusValues(t *testing.T) {
	// Test all test status values
	statuses := []TestStatus{
		TestStatusPending,
		TestStatusRunning,
		TestStatusPassed,
		TestStatusFailed,
		TestStatusSkipped,
		TestStatusTimeout,
	}

	for i, status := range statuses {
		if int(status) != i {
			t.Errorf("TestStatus %d has unexpected value: %d", i, int(status))
		}
	}
}

func TestUnitTestFramework_TestPriorityValues(t *testing.T) {
	// Test all test priority values
	priorities := []TestPriority{
		TestPriorityLow,
		TestPriorityNormal,
		TestPriorityHigh,
		TestPriorityCritical,
	}

	for i, priority := range priorities {
		if int(priority) != i {
			t.Errorf("TestPriority %d has unexpected value: %d", i, int(priority))
		}
	}
}

func TestUnitTestFramework_Concurrency(t *testing.T) {
	framework := NewUnitTestFramework(UnitTestConfig{
		MaxConcurrentTests: 10,
		EnableParallel:     true,
	})

	// Test concurrent additions
	const numGoroutines = 20
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			testSuite := &TestSuite{
				ID:          fmt.Sprintf("suite%d", id),
				Name:        fmt.Sprintf("Test Suite %d", id),
				Description: fmt.Sprintf("A test suite %d for testing", id),
			}

			err := framework.RegisterTestSuite(testSuite)
			if err != nil && err != ErrTestSuiteAlreadyExists {
				t.Errorf("Concurrent RegisterTestSuite %d failed: %v", id, err)
			}

			testCase := &TestCase{
				ID:          fmt.Sprintf("test%d", id),
				Name:        fmt.Sprintf("Test Case %d", id),
				Description: fmt.Sprintf("A test case %d for testing", id),
				Function:    func(t interface{}) error { return nil },
			}

			err = framework.AddTestCase(testCase)
			if err != nil {
				t.Errorf("Concurrent AddTestCase %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify that some test suites and test cases were added
	if len(framework.testSuites) == 0 {
		t.Error("No test suites were added concurrently")
	}

	if len(framework.testCases) == 0 {
		t.Error("No test cases were added concurrently")
	}
}
