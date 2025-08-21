package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestDataset_GetSummary(t *testing.T) {
	tdg := NewTestDataGenerators()

	// Generate some test data first
	dataset := tdg.GenerateTestDataset()
	require.NotNil(t, dataset)
	assert.Greater(t, len(dataset.Orders), 0)

	// Get summary
	summary := dataset.GetSummary()
	require.NotNil(t, summary)

	// Check that summary contains the expected data
	assert.Contains(t, summary, "Test Dataset Summary:")
	assert.Contains(t, summary, "Trading Pairs:")
	assert.Contains(t, summary, "Orders:")
	assert.Contains(t, summary, "Order Books:")
	assert.Contains(t, summary, "Bridge Transactions:")
	assert.Contains(t, summary, "Proposals:")
	assert.Contains(t, summary, "Votes:")
	assert.Contains(t, summary, "Treasury Transactions:")
	assert.Contains(t, summary, "Validators:")
	assert.Contains(t, summary, "Asset Mappings:")
}

func TestTestDataset_GetSummary_EmptyDataset(t *testing.T) {
	// Create an empty dataset
	dataset := &TestDataset{}

	// Get summary of empty dataset
	summary := dataset.GetSummary()
	require.NotEmpty(t, summary)

	// Should contain zero counts
	assert.Contains(t, summary, "Trading Pairs: 0")
	assert.Contains(t, summary, "Orders: 0")
	assert.Contains(t, summary, "Order Books: 0")
}

func TestTestDataset_GetSummary_WithDifferentData(t *testing.T) {
	tdg := NewTestDataGenerators()

	// Generate multiple datasets and test their summaries
	for i := 0; i < 3; i++ {
		dataset := tdg.GenerateTestDataset()
		require.NotNil(t, dataset)

		summary := dataset.GetSummary()
		require.NotEmpty(t, summary)

		// Each summary should be valid
		assert.Contains(t, summary, "Test Dataset Summary:")
	}
}

func TestTestDataset_GetSummary_Consistency(t *testing.T) {
	tdg := NewTestDataGenerators()

	// Generate a dataset
	dataset := tdg.GenerateTestDataset()
	require.NotNil(t, dataset)

	// Get summary multiple times
	summary1 := dataset.GetSummary()
	summary2 := dataset.GetSummary()

	require.NotEmpty(t, summary1)
	require.NotEmpty(t, summary2)

	// Should be identical since dataset didn't change
	assert.Equal(t, summary1, summary2)
}
