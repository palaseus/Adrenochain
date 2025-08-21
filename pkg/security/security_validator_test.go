package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityValidator(t *testing.T) {
	sv := NewSecurityValidator()

	assert.NotNil(t, sv)
	assert.NotNil(t, sv.Results)
	assert.Equal(t, 0, len(sv.Results))
}

func TestAddResult(t *testing.T) {
	sv := NewSecurityValidator()

	// Create a test result
	result := &SecurityValidationResult{
		PackageName:    "test_package",
		TestType:       "test_type",
		Status:         "PASS",
		Duration:       time.Millisecond * 100,
		IssuesFound:    0,
		CriticalIssues: 0,
		Warnings:       0,
		Timestamp:      time.Now(),
		Details:        map[string]interface{}{"test": "value"},
	}

	// Add the result
	sv.AddResult(result)

	// Verify the result was added
	assert.Equal(t, 1, len(sv.Results))
	assert.Equal(t, result, sv.Results[0])
}

func TestGetResults(t *testing.T) {
	sv := NewSecurityValidator()

	// Create multiple test results
	result1 := &SecurityValidationResult{
		PackageName: "package1",
		TestType:    "test1",
		Status:      "PASS",
		Timestamp:   time.Now(),
	}

	result2 := &SecurityValidationResult{
		PackageName: "package2",
		TestType:    "test2",
		Status:      "FAIL",
		Timestamp:   time.Now(),
	}

	// Add results
	sv.AddResult(result1)
	sv.AddResult(result2)

	// Get results
	results := sv.GetResults()

	// Verify results
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "package1", results[0].PackageName)
	assert.Equal(t, "package2", results[1].PackageName)

	// Verify that GetResults returns a copy, not the original slice
	results[0].PackageName = "modified"
	assert.Equal(t, "package1", sv.Results[0].PackageName)
}

func TestRunFuzzTest(t *testing.T) {
	sv := NewSecurityValidator()

	// Test fuzz test with no issues
	result := sv.runFuzzTest("test_package", "test_type", 10)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Fuzz Test - test_type", result.TestType)
	assert.Equal(t, "PASS", result.Status)
	assert.Equal(t, 10, result.Details["iterations"])
	assert.Equal(t, "random, boundary, malformed", result.Details["fuzz_patterns"])
	assert.Equal(t, "comprehensive", result.Details["test_coverage"])

	// Test fuzz test with issues
	result = sv.runFuzzTest("test_package", "test_type", 200)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Fuzz Test - test_type", result.TestType)
	assert.Equal(t, "FAIL", result.Status)
	assert.True(t, result.IssuesFound > 0)

	// Test fuzz test with critical issues
	result = sv.runFuzzTest("test_package", "test_type", 300)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Fuzz Test - test_type", result.TestType)
	assert.Equal(t, "FAIL", result.Status)
	assert.True(t, result.CriticalIssues > 0)
}

func TestRunRaceDetection(t *testing.T) {
	sv := NewSecurityValidator()

	// Test race detection with no issues
	result := sv.runRaceDetection("test_package", "test_type", 10)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Race Detection - test_type", result.TestType)
	assert.Equal(t, "PASS", result.Status)
	assert.Equal(t, 10, result.Details["iterations"])
	assert.Equal(t, "high", result.Details["concurrency_level"])
	assert.Equal(t, "static_analysis", result.Details["detection_method"])

	// Test race detection with issues
	result = sv.runRaceDetection("test_package", "test_type", 200)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Race Detection - test_type", result.TestType)
	assert.Equal(t, "WARNING", result.Status)
	assert.True(t, result.IssuesFound > 0)

	// Test race detection with critical issues
	result = sv.runRaceDetection("test_package", "test_type", 400)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Race Detection - test_type", result.TestType)
	assert.Equal(t, "FAIL", result.Status)
	assert.True(t, result.CriticalIssues > 0)
}

func TestRunMemoryLeakTest(t *testing.T) {
	sv := NewSecurityValidator()

	// Test memory leak test with no issues
	result := sv.runMemoryLeakTest("test_package", "test_type", 10)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Memory Leak Test - test_type", result.TestType)
	assert.Equal(t, "PASS", result.Status)
	assert.Equal(t, 10, result.Details["iterations"])
	assert.Equal(t, "allocation, deallocation, gc", result.Details["memory_patterns"])
	assert.Equal(t, "extended", result.Details["test_duration"])

	// Test memory leak test with issues
	result = sv.runMemoryLeakTest("test_package", "test_type", 100)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Memory Leak Test - test_type", result.TestType)
	assert.Equal(t, "WARNING", result.Status)
	assert.True(t, result.IssuesFound > 0)

	// Test memory leak test with critical issues
	result = sv.runMemoryLeakTest("test_package", "test_type", 200)

	assert.NotNil(t, result)
	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "Memory Leak Test - test_type", result.TestType)
	assert.Equal(t, "FAIL", result.Status)
	assert.True(t, result.CriticalIssues > 0)
}

func TestValidateLayer2Security(t *testing.T) {
	sv := NewSecurityValidator()

	// Test Layer 2 security validation
	err := sv.validateLayer2Security()

	assert.NoError(t, err)
	assert.Equal(t, 13, len(sv.Results)) // 6 packages with varying test types

	// Verify specific results
	zkRollupsFound := false
	optimisticRollupsFound := false
	stateChannelsFound := false
	paymentChannelsFound := false
	sidechainsFound := false
	shardingFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "ZK Rollups":
			zkRollupsFound = true
		case "Optimistic Rollups":
			optimisticRollupsFound = true
		case "State Channels":
			stateChannelsFound = true
		case "Payment Channels":
			paymentChannelsFound = true
		case "Sidechains":
			sidechainsFound = true
		case "Sharding":
			shardingFound = true
		}
	}

	assert.True(t, zkRollupsFound)
	assert.True(t, optimisticRollupsFound)
	assert.True(t, stateChannelsFound)
	assert.True(t, paymentChannelsFound)
	assert.True(t, sidechainsFound)
	assert.True(t, shardingFound)
}

func TestValidateCrossChainSecurity(t *testing.T) {
	sv := NewSecurityValidator()

	// Test cross-chain security validation
	err := sv.validateCrossChainSecurity()

	assert.NoError(t, err)
	assert.Equal(t, 8, len(sv.Results)) // 4 packages * 2 test types each

	// Verify specific results
	ibcFound := false
	atomicSwapsFound := false
	multiChainValidatorsFound := false
	crossChainDeFiFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "IBC Protocol":
			ibcFound = true
		case "Atomic Swaps":
			atomicSwapsFound = true
		case "Multi-Chain Validators":
			multiChainValidatorsFound = true
		case "Cross-Chain DeFi":
			crossChainDeFiFound = true
		}
	}

	assert.True(t, ibcFound)
	assert.True(t, atomicSwapsFound)
	assert.True(t, multiChainValidatorsFound)
	assert.True(t, crossChainDeFiFound)
}

func TestValidateGovernanceSecurity(t *testing.T) {
	sv := NewSecurityValidator()

	// Test governance security validation
	err := sv.validateGovernanceSecurity()

	assert.NoError(t, err)
	assert.Equal(t, 8, len(sv.Results)) // 4 packages * 2 test types each

	// Verify specific results
	quadraticVotingFound := false
	delegatedGovernanceFound := false
	proposalMarketsFound := false
	crossProtocolGovernanceFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "Quadratic Voting":
			quadraticVotingFound = true
		case "Delegated Governance":
			delegatedGovernanceFound = true
		case "Proposal Markets":
			proposalMarketsFound = true
		case "Cross-Protocol Governance":
			crossProtocolGovernanceFound = true
		}
	}

	assert.True(t, quadraticVotingFound)
	assert.True(t, delegatedGovernanceFound)
	assert.True(t, proposalMarketsFound)
	assert.True(t, crossProtocolGovernanceFound)
}

func TestValidatePrivacySecurity(t *testing.T) {
	sv := NewSecurityValidator()

	// Test privacy security validation
	err := sv.validatePrivacySecurity()

	assert.NoError(t, err)
	assert.Equal(t, 6, len(sv.Results)) // 3 packages * 2 test types each

	// Verify specific results
	privateDeFiFound := false
	privacyPoolsFound := false
	privacyZKRollupsFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "Private DeFi":
			privateDeFiFound = true
		case "Privacy Pools":
			privacyPoolsFound = true
		case "Privacy ZK-Rollups":
			privacyZKRollupsFound = true
		}
	}

	assert.True(t, privateDeFiFound)
	assert.True(t, privacyPoolsFound)
	assert.True(t, privacyZKRollupsFound)
}

func TestValidateAIMLSecurity(t *testing.T) {
	sv := NewSecurityValidator()

	// Test AI/ML security validation
	err := sv.validateAIMLSecurity()

	assert.NoError(t, err)
	assert.Equal(t, 6, len(sv.Results)) // 3 packages * 2 test types each

	// Verify specific results
	strategyGenerationFound := false
	predictiveAnalyticsFound := false
	sentimentAnalysisFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "Strategy Generation":
			strategyGenerationFound = true
		case "Predictive Analytics":
			predictiveAnalyticsFound = true
		case "Sentiment Analysis":
			sentimentAnalysisFound = true
		}
	}

	assert.True(t, strategyGenerationFound)
	assert.True(t, predictiveAnalyticsFound)
	assert.True(t, sentimentAnalysisFound)
}

func TestRunAllSecurityValidations(t *testing.T) {
	sv := NewSecurityValidator()

	// Test running all security validations
	err := sv.RunAllSecurityValidations()

	assert.NoError(t, err)
	assert.Equal(t, 41, len(sv.Results)) // Total of all validation results

	// Verify that all validation types were run
	layer2Found := false
	crossChainFound := false
	governanceFound := false
	privacyFound := false
	aimlFound := false

	for _, result := range sv.Results {
		switch result.PackageName {
		case "ZK Rollups", "Optimistic Rollups", "State Channels", "Payment Channels", "Sidechains", "Sharding":
			layer2Found = true
		case "IBC Protocol", "Atomic Swaps", "Multi-Chain Validators", "Cross-Chain DeFi":
			crossChainFound = true
		case "Quadratic Voting", "Delegated Governance", "Proposal Markets", "Cross-Protocol Governance":
			governanceFound = true
		case "Private DeFi", "Privacy Pools", "Privacy ZK-Rollups":
			privacyFound = true
		case "Strategy Generation", "Predictive Analytics", "Sentiment Analysis":
			aimlFound = true
		}
	}

	assert.True(t, layer2Found)
	assert.True(t, crossChainFound)
	assert.True(t, governanceFound)
	assert.True(t, privacyFound)
	assert.True(t, aimlFound)
}

func TestSecurityValidationResultFields(t *testing.T) {
	// Test that SecurityValidationResult has all required fields
	result := &SecurityValidationResult{
		PackageName:    "test_package",
		TestType:       "test_type",
		Status:         "PASS",
		Duration:       time.Millisecond * 100,
		IssuesFound:    0,
		CriticalIssues: 0,
		Warnings:       0,
		Timestamp:      time.Now(),
		Details:        map[string]interface{}{"test": "value"},
	}

	assert.Equal(t, "test_package", result.PackageName)
	assert.Equal(t, "test_type", result.TestType)
	assert.Equal(t, "PASS", result.Status)
	assert.Equal(t, time.Millisecond*100, result.Duration)
	assert.Equal(t, 0, result.IssuesFound)
	assert.Equal(t, 0, result.CriticalIssues)
	assert.Equal(t, 0, result.Warnings)
	assert.NotNil(t, result.Timestamp)
	assert.Equal(t, "value", result.Details["test"])
}

func TestConcurrentAccess(t *testing.T) {
	sv := NewSecurityValidator()

	// Test concurrent access to AddResult and GetResults
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			result := &SecurityValidationResult{
				PackageName: "concurrent_test",
				TestType:    "concurrent",
				Status:      "PASS",
				Timestamp:   time.Now(),
			}
			sv.AddResult(result)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(id int) {
			_ = sv.GetResults()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify results
	results := sv.GetResults()
	assert.Equal(t, 5, len(results)) // Only AddResult calls should add results
}
