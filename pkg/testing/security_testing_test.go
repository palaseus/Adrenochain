package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityTestFramework_GetResults(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Initially should be empty
	results := stf.GetResults()
	require.NotNil(t, results)
	assert.Len(t, results, 0)

	// Run security tests to populate results
	testResults := stf.RunAllSecurityTests()
	require.NotNil(t, testResults)

	// Should now have results
	results = stf.GetResults()
	require.NotNil(t, results)
	assert.Greater(t, len(results), 0)

	// Check that results contain expected components
	foundComponents := make(map[string]bool)
	for _, result := range results {
		assert.NotEmpty(t, result.TestType)
		assert.NotEmpty(t, result.Component)
		assert.NotEmpty(t, result.Status)
		foundComponents[result.Component] = true
	}

	// Should have multiple different components
	assert.True(t, len(foundComponents) > 1)
}

func TestSecurityTestFramework_PrintResults(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Run security tests first
	testResults := stf.RunAllSecurityTests()
	require.NotNil(t, testResults)

	// This should not panic and should execute successfully
	// Since PrintResults outputs to stdout, we can't easily capture the output
	// but we can ensure it doesn't crash
	assert.NotPanics(t, func() {
		stf.PrintResults()
	})
}

func TestSecurityTestFramework_RunAllSecurityTests(t *testing.T) {
	// Test the standalone function
	assert.NotPanics(t, func() {
		RunAllSecurityTests()
	})
}

func TestSecurityTestFramework_EdgeCases(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Test printing results with no security tests run
	assert.NotPanics(t, func() {
		stf.PrintResults()
	})

	// Test getting results with no security tests run
	results := stf.GetResults()
	require.NotNil(t, results)
	assert.Len(t, results, 0)
}

func TestSecurityTestFramework_ResultsStructure(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Run security tests
	testResults := stf.RunAllSecurityTests()
	require.NotNil(t, testResults)

	results := stf.GetResults()
	require.NotNil(t, results)
	require.Greater(t, len(results), 0)

	// Verify each result has the required fields
	for name, result := range results {
		assert.NotEmpty(t, name)
		assert.NotEmpty(t, result.TestType)
		assert.NotEmpty(t, result.Component)
		assert.NotEmpty(t, result.Status)
		assert.NotEmpty(t, result.Severity)
		assert.NotEmpty(t, result.Description)
		assert.False(t, result.Timestamp.IsZero())
	}
}

func TestSecurityTestFramework_ConcurrentAccess(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Run security tests to populate results
	testResults := stf.RunAllSecurityTests()
	require.NotNil(t, testResults)

	// Test concurrent access to GetResults
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			results := stf.GetResults()
			assert.NotNil(t, results)
			assert.Greater(t, len(results), 0)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSecurityTestFramework_MultipleRuns(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Run security tests multiple times
	for i := 0; i < 3; i++ {
		testResults := stf.RunAllSecurityTests()
		require.NotNil(t, testResults)

		results := stf.GetResults()
		require.NotNil(t, results)
		assert.Greater(t, len(results), 0)
	}

	// Final results should be valid
	results := stf.GetResults()
	require.NotNil(t, results)
	assert.Greater(t, len(results), 0)
}

func TestSecurityTestFramework_SpecificComponents(t *testing.T) {
	stf := NewSecurityTestFramework()

	// Run security tests
	testResults := stf.RunAllSecurityTests()
	require.NotNil(t, testResults)

	results := stf.GetResults()
	require.NotNil(t, results)
	require.Greater(t, len(results), 0)

	// Check for expected security test categories
	expectedComponents := []string{
		"OrderBook", "MatchingEngine", "TradingPair", "Bridge",
		"Validator", "VotingSystem", "Treasury", "Lending", "AMM",
	}

	foundComponents := make(map[string]bool)
	for _, result := range results {
		foundComponents[result.Component] = true
	}

	// Should find at least some of the expected components
	foundCount := 0
	for _, expected := range expectedComponents {
		if foundComponents[expected] {
			foundCount++
		}
	}

	assert.Greater(t, foundCount, 0)
}
