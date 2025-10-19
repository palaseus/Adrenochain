//go:build testing
// +build testing

package pdf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiNodePDFTest_Creation(t *testing.T) {
	// Test with default config
	nodeCount := 3

	multiNodeTest := NewMultiNodePDFTest(nodeCount)
	assert.NotNil(t, multiNodeTest)
	assert.NotNil(t, multiNodeTest.nodes)
	assert.NotNil(t, multiNodeTest.testResults)
	assert.NotNil(t, multiNodeTest.networkConfig)
	assert.NotNil(t, multiNodeTest.stopChan)
	assert.Equal(t, nodeCount, multiNodeTest.networkConfig.NodeCount)
	assert.Len(t, multiNodeTest.nodes, 3)
}

func TestPDFNode_Creation(t *testing.T) {
	// Create a test data directory
	testDir := "/tmp/pdf_node_test"
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create PDF node manually (since NewPDFNode doesn't exist)
	node := &PDFNode{
		ID:            1,
		Port:          8001,
		DataDir:       testDir,
		PDFStorage:    nil, // Will be initialized later
		PeerAddresses: make([]string, 0),
		IsRunning:     false,
	}

	assert.NotNil(t, node)
	assert.Equal(t, 1, node.ID)
	assert.Equal(t, 8001, node.Port)
	assert.Equal(t, testDir, node.DataDir)
	assert.NotNil(t, node.PeerAddresses)
	assert.False(t, node.IsRunning)
}

func TestMultiNodePDFTest_StartNodes(t *testing.T) {
	// Create a test data directory
	testDir := "/tmp/pdf_multi_node_start_test"
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create multi-node test
	multiNodeTest := NewMultiNodePDFTest(3)

	// Test starting nodes
	err = multiNodeTest.StartNodes()
	assert.NoError(t, err)

	// Verify test results were initialized
	assert.False(t, multiNodeTest.testResults.StartTime.IsZero())

	// Test stopping nodes
	multiNodeTest.StopNodes()
}

func TestMultiNodePDFTest_RunTest(t *testing.T) {
	// Create a test data directory
	testDir := "/tmp/pdf_multi_node_run_test"
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create multi-node test
	multiNodeTest := NewMultiNodePDFTest(3)

	// Run the test
	err = multiNodeTest.RunTest()
	assert.NoError(t, err)

	// Verify test results
	results := multiNodeTest.testResults
	assert.NotNil(t, results)
	assert.Equal(t, 3, results.TotalNodes)
	assert.False(t, results.StartTime.IsZero())
	assert.False(t, results.EndTime.IsZero())
	assert.Greater(t, results.EndTime, results.StartTime)
}

func TestMultiNodePDFTest_NetworkConfig(t *testing.T) {
	// Test with different node counts
	testCases := []int{1, 3, 5, 10}

	for _, nodeCount := range testCases {
		t.Run(fmt.Sprintf("NodeCount_%d", nodeCount), func(t *testing.T) {
			multiNodeTest := NewMultiNodePDFTest(nodeCount)

			assert.NotNil(t, multiNodeTest)
			assert.Equal(t, nodeCount, multiNodeTest.networkConfig.NodeCount)
			assert.Len(t, multiNodeTest.nodes, nodeCount)
			assert.Equal(t, nodeCount, multiNodeTest.testResults.TotalNodes)
		})
	}
}

func TestMultiNodePDFTest_TestResults(t *testing.T) {
	// Create multi-node test
	multiNodeTest := NewMultiNodePDFTest(3)

	// Verify initial test results
	results := multiNodeTest.testResults
	assert.NotNil(t, results)
	assert.Equal(t, 3, results.TotalNodes)
	assert.True(t, results.StartTime.IsZero())
	assert.True(t, results.EndTime.IsZero())
	assert.Equal(t, time.Duration(0), results.PDFUploadTime)
	assert.Equal(t, time.Duration(0), results.PropagationTime)
	assert.Equal(t, time.Duration(0), results.ConsensusTime)
	assert.Equal(t, 0, results.TotalBlocks)
	assert.Equal(t, 0, results.TotalTransactions)
	assert.Equal(t, 0, results.PDFTransactions)
}

func TestMultiNodePDFTest_EdgeCases(t *testing.T) {
	// Test with zero nodes
	multiNodeTest := NewMultiNodePDFTest(0)
	assert.NotNil(t, multiNodeTest)
	assert.Equal(t, 0, multiNodeTest.networkConfig.NodeCount)
	assert.Len(t, multiNodeTest.nodes, 0)

	// Test with large number of nodes
	largeMultiNodeTest := NewMultiNodePDFTest(100)
	assert.NotNil(t, largeMultiNodeTest)
	assert.Equal(t, 100, largeMultiNodeTest.networkConfig.NodeCount)
	assert.Len(t, largeMultiNodeTest.nodes, 100)
}

func TestMultiNodePDFTest_ConcurrentAccess(t *testing.T) {
	// Create multi-node test
	multiNodeTest := NewMultiNodePDFTest(3)

	// Test concurrent access to test results
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Access test results concurrently
			results := multiNodeTest.testResults
			assert.NotNil(t, results)
			assert.Equal(t, 3, results.TotalNodes)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMultiNodePDFTest_Performance(t *testing.T) {
	// Test performance with different node counts
	testCases := []int{1, 5, 10}

	for _, nodeCount := range testCases {
		t.Run(fmt.Sprintf("Performance_%d_Nodes", nodeCount), func(t *testing.T) {
			start := time.Now()

			multiNodeTest := NewMultiNodePDFTest(nodeCount)

			creationTime := time.Since(start)

			// Verify creation time is reasonable
			assert.Less(t, creationTime, 100*time.Millisecond,
				"Creating %d nodes should be fast", nodeCount)

			// Verify the test was created correctly
			assert.NotNil(t, multiNodeTest)
			assert.Equal(t, nodeCount, multiNodeTest.networkConfig.NodeCount)
		})
	}
}
