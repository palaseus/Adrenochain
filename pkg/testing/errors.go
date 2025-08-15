package testing

import "errors"

// Testing-specific errors
var (
	ErrTestSuiteAlreadyExists    = errors.New("test suite already exists")
	ErrTestSuiteNotFound         = errors.New("test suite not found")
	ErrTestCaseNotFound          = errors.New("test case not found")
	ErrInvalidTestCaseID         = errors.New("invalid test case ID")
	ErrInvalidTestCaseName       = errors.New("invalid test case name")
	ErrInvalidTestCaseFunction   = errors.New("invalid test case function")
	ErrAutoGenerationNotEnabled  = errors.New("auto generation not enabled")
	ErrTestExecutionFailed       = errors.New("test execution failed")
	ErrCoverageTrackingFailed    = errors.New("coverage tracking failed")
	ErrPerformanceMonitoringFailed = errors.New("performance monitoring failed")
	ErrTestTimeout               = errors.New("test timeout")
	ErrTestSetupFailed           = errors.New("test setup failed")
	ErrTestTeardownFailed        = errors.New("test teardown failed")
)
