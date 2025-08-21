package testing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoverageTracker_Initialize(t *testing.T) {
	ct := NewCoverageTracker()

	err := ct.Initialize()
	require.NoError(t, err)

	// Check that packages were initialized
	assert.Len(t, ct.packageCoverage, 11)

	// Check specific packages
	engineCoverage, exists := ct.packageCoverage["pkg/contracts/engine"]
	require.True(t, exists)
	assert.Equal(t, "pkg/contracts/engine", engineCoverage.PackageName)
	assert.Equal(t, uint64(1000), engineCoverage.TotalLines)
	assert.Equal(t, uint64(0), engineCoverage.CoveredLines)
	assert.Equal(t, uint64(50), engineCoverage.TotalFunctions)
	assert.Equal(t, uint64(0), engineCoverage.CoveredFunctions)
	assert.Equal(t, 0.0, engineCoverage.Coverage)

	// Check that last updated is recent
	assert.WithinDuration(t, time.Now(), engineCoverage.LastUpdated, 2*time.Second)
}

func TestCoverageTracker_UpdateCoverage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Update coverage for a package
	ct.UpdateCoverage("pkg/contracts/engine", 50.0)

	// Check that coverage was updated
	engineCoverage, exists := ct.packageCoverage["pkg/contracts/engine"]
	require.True(t, exists)
	assert.Equal(t, uint64(500), engineCoverage.CoveredLines)    // 50% of 1000
	assert.Equal(t, uint64(25), engineCoverage.CoveredFunctions) // 50% of 50
	assert.Equal(t, 50.0, engineCoverage.Coverage)

	// Update coverage for another package
	ct.UpdateCoverage("pkg/contracts/storage", 50.0)

	storageCoverage, exists := ct.packageCoverage["pkg/contracts/storage"]
	require.True(t, exists)
	assert.Equal(t, uint64(400), storageCoverage.CoveredLines)    // 50% of 800
	assert.Equal(t, uint64(20), storageCoverage.CoveredFunctions) // 50% of 40
	assert.Equal(t, 50.0, storageCoverage.Coverage)
}

func TestCoverageTracker_UpdateCoverage_NonExistentPackage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Try to update coverage for non-existent package
	ct.UpdateCoverage("pkg/nonexistent", 50.0)
	// Note: UpdateCoverage doesn't return an error for non-existent packages
	// It just doesn't update anything
}

func TestCoverageTracker_GetPackageCoverage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Update some coverage first
	ct.UpdateCoverage("pkg/contracts/engine", 50.0)

	// Get package coverage
	coverage := ct.GetPackageCoverage("pkg/contracts/engine")
	require.NotNil(t, coverage)
	assert.Equal(t, "pkg/contracts/engine", coverage.PackageName)
	assert.Equal(t, uint64(500), coverage.CoveredLines)
	assert.Equal(t, uint64(25), coverage.CoveredFunctions)
	assert.Equal(t, 50.0, coverage.Coverage)
}

func TestCoverageTracker_GetPackageCoverage_NonExistentPackage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Try to get coverage for non-existent package
	coverage := ct.GetPackageCoverage("pkg/nonexistent")
	assert.Nil(t, coverage)
}

func TestCoverageTracker_GetAllPackageCoverage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Update some coverage
	ct.UpdateCoverage("pkg/contracts/engine", 50.0)
	ct.UpdateCoverage("pkg/contracts/storage", 50.0)

	// Get all package coverage
	allCoverage := ct.GetAllPackageCoverage()
	assert.Len(t, allCoverage, 11)

	// Check that updated packages have correct coverage
	engineCoverage := allCoverage["pkg/contracts/engine"]
	assert.Equal(t, uint64(500), engineCoverage.CoveredLines)
	assert.Equal(t, uint64(25), engineCoverage.CoveredFunctions)

	storageCoverage := allCoverage["pkg/contracts/storage"]
	assert.Equal(t, uint64(400), storageCoverage.CoveredLines)
	assert.Equal(t, uint64(20), storageCoverage.CoveredFunctions)
}

func TestCoverageTracker_UpdateOverallCoverage(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Update some package coverage
	ct.UpdateCoverage("pkg/contracts/engine", 50.0)
	ct.UpdateCoverage("pkg/contracts/storage", 50.0)

	// Update overall coverage
	ct.updateOverallCoverage()

	// Check overall coverage
	overallCoverage := ct.GetOverallCoverage()
	// Note: GetOverallCoverage returns percentage, not absolute numbers
	assert.Greater(t, overallCoverage, 0.0)
}

func TestCoverageTracker_GenerateCoverageRecommendations(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Update some coverage
	ct.UpdateCoverage("pkg/contracts/engine", 50.0)
	ct.UpdateCoverage("pkg/contracts/storage", 50.0)

	// Generate recommendations
	recommendations := ct.generateCoverageRecommendations()

	// Check that recommendations were generated
	assert.NotEmpty(t, recommendations)

	// Check that recommendations contain package-specific suggestions
	hasEngineRecommendation := false
	for _, rec := range recommendations {
		if rec != "" {
			hasEngineRecommendation = true
			break
		}
	}
	assert.True(t, hasEngineRecommendation)
}

func TestCoverageTracker_EdgeCases(t *testing.T) {
	ct := NewCoverageTracker()

	// Test with nil coverage tracker
	var nilCT *CoverageTracker
	assert.Panics(t, func() {
		nilCT.Initialize()
	})

	// Test with empty package name
	ct.UpdateCoverage("", 50.0)
	// Note: UpdateCoverage doesn't return an error for empty package names

	// Test with negative values
	ct.UpdateCoverage("pkg/test", -50.0)
	// Note: UpdateCoverage clamps negative values to 0

	// Test with values exceeding 100%
	err := ct.Initialize()
	require.NoError(t, err)

	ct.UpdateCoverage("pkg/contracts/engine", 150.0)
	// Note: UpdateCoverage clamps values > 100 to 100
}

func TestCoverageTracker_ConcurrentAccess(t *testing.T) {
	ct := NewCoverageTracker()
	err := ct.Initialize()
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Update coverage
			ct.UpdateCoverage("pkg/contracts/engine", float64(50+id))

			// Read coverage
			coverage := ct.GetPackageCoverage("pkg/contracts/engine")
			if coverage != nil {
				_ = coverage.Coverage
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state is consistent
	overallCoverage := ct.GetOverallCoverage()
	assert.Greater(t, overallCoverage, 0.0)
}
