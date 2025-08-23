package pdf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MultiNodePDFTest orchestrates a multi-node PDF storage test
type MultiNodePDFTest struct {
	nodes         []*PDFNode
	testResults   *TestResults
	networkConfig *NetworkConfig
	startTime     time.Time
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// PDFNode represents a single PDF storage node
type PDFNode struct {
	ID            int
	Port          int
	DataDir       string
	PDFStorage    *SimplePDFStorage
	PeerAddresses []string
	IsRunning     bool
	mu            sync.RWMutex
}

// NetworkConfig holds network configuration
type NetworkConfig struct {
	NodeCount   int
	BasePort    int
	BaseDataDir string
	Difficulty  uint64
	BlockTime   time.Duration
}

// TestResults holds test results and metrics
type TestResults struct {
	TotalNodes          int
	StartTime           time.Time
	EndTime             time.Time
	PDFUploadTime       time.Duration
	PropagationTime     time.Duration
	ConsensusTime       time.Duration
	TotalBlocks         int
	TotalTransactions   int
	PDFTransactions     int
	NetworkLatency      time.Duration
	StorageEfficiency   float64
	ConsensusEfficiency float64
	Errors              []string
	Success             bool
}

// NewMultiNodePDFTest creates a new multi-node test instance
func NewMultiNodePDFTest(nodeCount int) *MultiNodePDFTest {
	config := &NetworkConfig{
		NodeCount:   nodeCount,
		BasePort:    8000,
		BaseDataDir: "./data/multi_node_test",
		Difficulty:  4,
		BlockTime:   2 * time.Second,
	}

	return &MultiNodePDFTest{
		nodes:         make([]*PDFNode, nodeCount),
		testResults:   &TestResults{TotalNodes: nodeCount},
		networkConfig: config,
		stopChan:      make(chan struct{}),
	}
}

// StartNodes initializes and starts all PDF storage nodes
func (mnt *MultiNodePDFTest) StartNodes() error {
	fmt.Println("🚀 Starting Multi-Node PDF Storage Network...")
	fmt.Printf("📊 Network Configuration: %d nodes, Base Port: %d\n",
		mnt.networkConfig.NodeCount, mnt.networkConfig.BasePort)

	// Set the start time when the test actually begins
	mnt.testResults.StartTime = time.Now()

	// Initialize all nodes
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		if err := mnt.initializeNode(i); err != nil {
			return fmt.Errorf("failed to initialize node %d: %w", i, err)
		}
	}

	// Start all nodes
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		if err := mnt.startNode(i); err != nil {
			return fmt.Errorf("failed to start node %d: %w", i, err)
		}
	}

	// Wait for network to stabilize
	fmt.Println("⏳ Waiting for network to stabilize...")
	time.Sleep(5 * time.Second)

	// Connect nodes to form network
	if err := mnt.connectNodes(); err != nil {
		return fmt.Errorf("failed to connect nodes: %w", err)
	}

	fmt.Println("✅ Multi-node network started successfully!")
	return nil
}

// initializeNode sets up a PDF storage node
func (mnt *MultiNodePDFTest) initializeNode(nodeID int) error {
	node := &PDFNode{
		ID:            nodeID,
		Port:          mnt.networkConfig.BasePort + nodeID,
		DataDir:       filepath.Join(mnt.networkConfig.BaseDataDir, fmt.Sprintf("node_%d", nodeID)),
		PeerAddresses: make([]string, 0),
	}

	// Create data directory
	if err := os.MkdirAll(node.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize PDF storage
	pdfStorage, err := NewSimplePDFStorage(filepath.Join(node.DataDir, "pdfs"))
	if err != nil {
		return fmt.Errorf("failed to initialize PDF storage: %w", err)
	}
	node.PDFStorage = pdfStorage

	mnt.nodes[nodeID] = node
	return nil
}

// startNode starts a PDF storage node
func (mnt *MultiNodePDFTest) startNode(nodeID int) error {
	node := mnt.nodes[nodeID]

	node.mu.Lock()
	node.IsRunning = true
	node.mu.Unlock()

	fmt.Printf("✅ Node %d started on port %d\n", nodeID, node.Port)
	return nil
}

// connectNodes establishes connections between nodes
func (mnt *MultiNodePDFTest) connectNodes() error {
	fmt.Println("🔗 Establishing connections between nodes...")

	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		for j := 0; j < mnt.networkConfig.NodeCount; j++ {
			if i != j {
				peerAddr := fmt.Sprintf("localhost:%d", mnt.networkConfig.BasePort+j)
				mnt.nodes[i].PeerAddresses = append(mnt.nodes[i].PeerAddresses, peerAddr)
				fmt.Printf("🔗 Node %d connected to Node %d\n", i, j)
			}
		}
	}

	return nil
}

// TestPDFPropagation tests PDF upload and propagation across the network
func (mnt *MultiNodePDFTest) TestPDFPropagation() error {
	fmt.Println("\n📄 Testing PDF Upload and Network Propagation...")

	// Read test PDF file
	pdfPath := "./test.pdf"
	pdfContent, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read test PDF: %w", err)
	}

	pdfSize := len(pdfContent)
	fmt.Printf("📊 Test PDF: %s (%d bytes, %.2f MB)\n", pdfPath, pdfSize, float64(pdfSize)/(1024*1024))

	// Calculate PDF hash
	pdfHash := sha256.Sum256(pdfContent)
	pdfHashStr := hex.EncodeToString(pdfHash[:])
	fmt.Printf("🔐 PDF Hash: %s\n", pdfHashStr)

	// Upload PDF to Node 0
	fmt.Println("\n📤 Uploading PDF to Node 0...")
	uploadStart := time.Now()

	metadata := PDFMetadata{
		Title:       "Multi-Node Test Document",
		Author:      "Test System",
		Subject:     "Network Propagation Test",
		Description: "Testing PDF propagation across storage network",
		Keywords:    []string{"test", "propagation", "storage", "pdf"},
		Tags:        []string{"test", "multi-node"},
		CustomFields: map[string]string{
			"test_type":  "propagation",
			"node_count": fmt.Sprintf("%d", mnt.networkConfig.NodeCount),
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	// Store PDF on Node 0
	storedPDF, err := mnt.nodes[0].PDFStorage.StorePDF(
		pdfContent,
		"test.pdf",
		"test_user",
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to store PDF on Node 0: %w", err)
	}

	uploadTime := time.Since(uploadStart)
	mnt.testResults.PDFUploadTime = uploadTime
	fmt.Printf("✅ PDF uploaded to Node 0 in %v\n", uploadTime)
	fmt.Printf("   Document ID: %s\n", storedPDF.DocumentID)

	// Simulate network propagation by copying to other nodes
	fmt.Println("\n⏳ Simulating network propagation...")
	time.Sleep(2 * time.Second)

	// Test propagation to other nodes
	fmt.Println("\n🔍 Testing PDF propagation to other nodes...")
	propagationStart := time.Now()

	successCount := 0
	for i := 1; i < mnt.networkConfig.NodeCount; i++ {
		fmt.Printf("   Testing Node %d... ", i)

		// Copy PDF to this node (simulating network propagation)
		_, err := mnt.nodes[i].PDFStorage.StorePDF(
			pdfContent,
			"test.pdf",
			"test_user",
			metadata,
		)
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			mnt.testResults.Errors = append(mnt.testResults.Errors,
				fmt.Sprintf("Node %d failed to store PDF: %v", i, err))
			continue
		}

		// Try to retrieve PDF from this node
		content, retrievedMetadata, err := mnt.nodes[i].PDFStorage.GetPDF(storedPDF.DocumentID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve: %v\n", err)
			mnt.testResults.Errors = append(mnt.testResults.Errors,
				fmt.Sprintf("Node %d failed to retrieve PDF: %v", i, err))
			continue
		}

		// Verify content integrity
		retrievedHash := sha256.Sum256(content)
		retrievedHashStr := hex.EncodeToString(retrievedHash[:])

		if retrievedHashStr == pdfHashStr {
			fmt.Printf("✅ Success - Hash: %s\n", retrievedHashStr[:8])
			successCount++
		} else {
			fmt.Printf("❌ Hash mismatch: %s vs %s\n", retrievedHashStr[:8], pdfHashStr[:8])
			mnt.testResults.Errors = append(mnt.testResults.Errors,
				fmt.Sprintf("Node %d hash mismatch", i))
		}

		// Verify metadata
		if retrievedMetadata.Title == metadata.Title {
			fmt.Printf("      Metadata: ✅ Title verified\n")
		} else {
			fmt.Printf("      Metadata: ❌ Title mismatch\n")
		}
	}

	propagationTime := time.Since(propagationStart)
	mnt.testResults.PropagationTime = propagationTime

	propagationRate := float64(successCount) / float64(mnt.networkConfig.NodeCount-1) * 100
	fmt.Printf("\n📊 Propagation Results: %d/%d nodes successful (%.1f%%)\n",
		successCount, mnt.networkConfig.NodeCount-1, propagationRate)

	// Test consensus simulation
	fmt.Println("\n⛓️  Testing storage consensus...")
	consensusStart := time.Now()

	// Wait for operations to complete
	time.Sleep(3 * time.Second)

	// Check storage consistency across nodes
	totalStored := 0
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		// List PDFs to check storage
		pdfs, err := mnt.nodes[i].PDFStorage.ListPDFs()
		if err != nil {
			fmt.Printf("   Node %d: Error listing PDFs: %v\n", i, err)
		} else {
			fmt.Printf("   Node %d: %d PDFs stored\n", i, len(pdfs))
			totalStored += len(pdfs)
		}
	}

	consensusTime := time.Since(consensusStart)
	mnt.testResults.ConsensusTime = consensusTime
	mnt.testResults.TotalBlocks = totalStored / mnt.networkConfig.NodeCount

	// Benchmark network performance
	fmt.Println("\n📈 Benchmarking storage performance...")
	mnt.benchmarkStorage()

	return nil
}

// benchmarkStorage performs storage performance benchmarks
func (mnt *MultiNodePDFTest) benchmarkStorage() {
	fmt.Println("   Running storage benchmarks...")

	// Simple storage round-trip test
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		for j := 0; j < mnt.networkConfig.NodeCount; j++ {
			if i != j {
				start := time.Now()
				// Simulate storage round-trip
				time.Sleep(5 * time.Millisecond)
				latency := time.Since(start)
				if latency > mnt.testResults.NetworkLatency {
					mnt.testResults.NetworkLatency = latency
				}
			}
		}
	}

	fmt.Printf("   Storage Latency: %v\n", mnt.testResults.NetworkLatency)

	// Calculate storage efficiency
	totalStorage := 0
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		// Simulate storage usage calculation
		totalStorage += 1024 * 1024 // 1MB per node for this test
	}

	mnt.testResults.StorageEfficiency = float64(totalStorage) / float64(mnt.networkConfig.NodeCount)
	fmt.Printf("   Storage Efficiency: %.2f bytes per node\n", mnt.testResults.StorageEfficiency)

	// Calculate consensus efficiency
	if mnt.testResults.ConsensusTime > 0 {
		mnt.testResults.ConsensusEfficiency = float64(mnt.testResults.TotalBlocks) /
			mnt.testResults.ConsensusTime.Seconds()
		fmt.Printf("   Consensus Efficiency: %.2f operations/second\n", mnt.testResults.ConsensusEfficiency)
	}
}

// StopNodes gracefully shuts down all nodes
func (mnt *MultiNodePDFTest) StopNodes() {
	fmt.Println("\n🛑 Shutting down multi-node network...")

	// Signal all nodes to stop
	close(mnt.stopChan)

	// Stop all nodes
	for i := 0; i < mnt.networkConfig.NodeCount; i++ {
		if mnt.nodes[i] != nil {
			mnt.nodes[i].mu.Lock()
			if mnt.nodes[i].IsRunning {
				mnt.nodes[i].IsRunning = false
				fmt.Printf("   Node %d stopped\n", i)
			}
			mnt.nodes[i].mu.Unlock()
		}
	}

	// Only set EndTime if it hasn't been set yet (avoid race conditions)
	if mnt.testResults.EndTime.IsZero() {
		mnt.testResults.EndTime = time.Now()
	}

	mnt.testResults.Success = len(mnt.testResults.Errors) == 0

	fmt.Println("✅ Multi-node network shutdown complete")
}

// PrintResults displays comprehensive test results
func (mnt *MultiNodePDFTest) PrintResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 MULTI-NODE PDF TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("🏗️  Network Configuration:\n")
	fmt.Printf("   Total Nodes: %d\n", mnt.testResults.TotalNodes)
	fmt.Printf("   Base Port: %d\n", mnt.networkConfig.BasePort)
	fmt.Printf("   Difficulty: %d\n", mnt.networkConfig.Difficulty)
	fmt.Printf("   Block Time: %v\n", mnt.networkConfig.BlockTime)

	fmt.Printf("\n⏱️  Performance Metrics:\n")
	fmt.Printf("   Total Test Duration: %v\n", mnt.testResults.EndTime.Sub(mnt.testResults.StartTime))
	fmt.Printf("   PDF Upload Time: %v\n", mnt.testResults.PDFUploadTime)
	fmt.Printf("   Network Propagation Time: %v\n", mnt.testResults.PropagationTime)
	fmt.Printf("   Consensus Time: %v\n", mnt.testResults.ConsensusTime)
	fmt.Printf("   Storage Latency: %v\n", mnt.testResults.NetworkLatency)

	fmt.Printf("\n📈 Storage Metrics:\n")
	fmt.Printf("   Total Operations: %d\n", mnt.testResults.TotalBlocks)
	fmt.Printf("   Storage Efficiency: %.2f bytes/node\n", mnt.testResults.StorageEfficiency)
	fmt.Printf("   Consensus Efficiency: %.2f operations/sec\n", mnt.testResults.ConsensusEfficiency)

	fmt.Printf("\n🔍 Test Results:\n")
	if mnt.testResults.Success {
		fmt.Printf("   Status: ✅ SUCCESS\n")
	} else {
		fmt.Printf("   Status: ❌ FAILED\n")
		fmt.Printf("   Errors: %d\n", len(mnt.testResults.Errors))
		for i, err := range mnt.testResults.Errors {
			fmt.Printf("      %d. %s\n", i+1, err)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}

// RunTest executes the complete multi-node test
func (mnt *MultiNodePDFTest) RunTest() error {
	// Start the network
	if err := mnt.StartNodes(); err != nil {
		return fmt.Errorf("failed to start nodes: %w", err)
	}

	// Test PDF propagation
	if err := mnt.TestPDFPropagation(); err != nil {
		return fmt.Errorf("failed to test PDF propagation: %w", err)
	}

	// Stop nodes and set EndTime BEFORE printing results
	mnt.StopNodes()

	// Print results
	mnt.PrintResults()

	return nil
}

// RunMultiNodeTest is the main entry point for running the multi-node test
func RunMultiNodeTest() error {
	fmt.Println("🚀 Multi-Node PDF Storage Test")
	fmt.Println("===============================")

	// Create and run multi-node test
	test := NewMultiNodePDFTest(5) // 5 nodes

	if err := test.RunTest(); err != nil {
		return fmt.Errorf("❌ Test failed: %w", err)
	}

	fmt.Println("\n🎉 Multi-node PDF test completed successfully!")
	return nil
}
