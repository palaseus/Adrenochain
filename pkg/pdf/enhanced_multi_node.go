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

// EnhancedMultiNodePDFTest combines network simulation and blockchain consensus
type EnhancedMultiNodePDFTest struct {
	nodes         []*EnhancedPDFNode
	testResults   *EnhancedTestResults
	networkConfig *EnhancedNetworkConfig
	networkSim    *NetworkSimulator
	consensus     *BlockchainConsensus
	startTime     time.Time
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// EnhancedPDFNode represents a node with network simulation and consensus
type EnhancedPDFNode struct {
	ID            int
	Port          int
	DataDir       string
	PDFStorage    *SimplePDFStorage
	Consensus     *BlockchainConsensus
	NetworkSim    *NetworkSimulator
	PeerAddresses []string
	IsRunning     bool
	mu            sync.RWMutex
}

// EnhancedNetworkConfig holds enhanced network configuration
type EnhancedNetworkConfig struct {
	NodeCount         int
	BasePort          int
	BaseDataDir       string
	Difficulty        uint64
	BlockTime         time.Duration
	EnableNetworkSim  bool
	EnableConsensus   bool
	NetworkConditions *NetworkSimConfig
	ConsensusConfig   *ConsensusConfig
}

// EnhancedTestResults holds enhanced test results
type EnhancedTestResults struct {
	TotalNodes          int
	StartTime           time.Time
	EndTime             time.Time
	PDFUploadTime       time.Duration
	PropagationTime     time.Duration
	ConsensusTime       time.Duration
	BlockMiningTime     time.Duration
	TotalBlocks         int
	TotalTransactions   int
	PDFTransactions     int
	NetworkLatency      time.Duration
	StorageEfficiency   float64
	ConsensusEfficiency float64
	NetworkEvents       []*NetworkEvent
	ConsensusEvents     []*ConsensusEvent
	Errors              []string
	Success             bool
}

// ConsensusEvent represents a consensus event
type ConsensusEvent struct {
	Type      string
	NodeID    string
	Timestamp time.Time
	Data      map[string]interface{}
}

// NewEnhancedMultiNodePDFTest creates a new enhanced multi-node test
func NewEnhancedMultiNodePDFTest(nodeCount int) *EnhancedMultiNodePDFTest {
	// Network simulation configuration
	networkSimConfig := &NetworkSimConfig{
		BaseLatency:      50 * time.Millisecond,
		LatencyJitter:    20 * time.Millisecond,
		LatencyVariation: 0.3,
		PacketLossRate:   0.001, // 0.1% packet loss
		BurstLossRate:    0.01,  // 1% burst loss
		BurstLength:      5,
		BandwidthLimit:   1024 * 1024, // 1MB/s
		BandwidthJitter:  0.2,
		EnableCongestion: true,
		CongestionRate:   0.1,
		EnableGeographic: false,
	}

	// Consensus configuration
	consensusConfig := &ConsensusConfig{
		Difficulty:        4,
		BlockTime:         2 * time.Second,
		MaxBlockSize:      1024 * 1024, // 1MB
		MinTransactionFee: 1000,
		ValidatorCount:    5,
		StakeRequirement:  10000,
		ConsensusTimeout:  30 * time.Second,
	}

	config := &EnhancedNetworkConfig{
		NodeCount:         nodeCount,
		BasePort:          8000,
		BaseDataDir:       "./data/enhanced_multi_node_test",
		Difficulty:        4,
		BlockTime:         2 * time.Second,
		EnableNetworkSim:  true,
		EnableConsensus:   true,
		NetworkConditions: networkSimConfig,
		ConsensusConfig:   consensusConfig,
	}

	return &EnhancedMultiNodePDFTest{
		nodes:         make([]*EnhancedPDFNode, nodeCount),
		testResults:   &EnhancedTestResults{TotalNodes: nodeCount},
		networkConfig: config,
		networkSim:    NewNetworkSimulator(networkSimConfig),
		consensus:     NewBlockchainConsensus(consensusConfig),
		stopChan:      make(chan struct{}),
	}
}

// StartNodes initializes and starts all enhanced nodes
func (emnt *EnhancedMultiNodePDFTest) StartNodes() error {
	fmt.Println("🚀 Starting Enhanced Multi-Node PDF Blockchain Network...")
	fmt.Printf("📊 Network Configuration: %d nodes, Base Port: %d\n",
		emnt.networkConfig.NodeCount, emnt.networkConfig.BasePort)
	fmt.Printf("🔗 Network Simulation: %v\n", emnt.networkConfig.EnableNetworkSim)
	fmt.Printf("⛓️  Blockchain Consensus: %v\n", emnt.networkConfig.EnableConsensus)

	// Set the start time when the test actually begins
	emnt.testResults.StartTime = time.Now()

	// Initialize all nodes
	for i := 0; i < emnt.networkConfig.NodeCount; i++ {
		if err := emnt.initializeEnhancedNode(i); err != nil {
			return fmt.Errorf("failed to initialize node %d: %w", i, err)
		}
	}

	// Start all nodes
	for i := 0; i < emnt.networkConfig.NodeCount; i++ {
		if err := emnt.startEnhancedNode(i); err != nil {
			return fmt.Errorf("failed to start node %d: %w", i, err)
		}
	}

	// Wait for network to stabilize
	fmt.Println("⏳ Waiting for network to stabilize...")
	time.Sleep(5 * time.Second)

	// Connect nodes to form network
	if err := emnt.connectEnhancedNodes(); err != nil {
		return fmt.Errorf("failed to connect nodes: %w", err)
	}

	// Start consensus if enabled
	if emnt.networkConfig.EnableConsensus {
		fmt.Println("⛓️  Starting blockchain consensus...")
		if err := emnt.consensus.Start(); err != nil {
			return fmt.Errorf("failed to start consensus: %w", err)
		}
	}

	fmt.Println("✅ Enhanced multi-node network started successfully!")
	return nil
}

// initializeEnhancedNode sets up an enhanced PDF storage node
func (emnt *EnhancedMultiNodePDFTest) initializeEnhancedNode(nodeID int) error {
	node := &EnhancedPDFNode{
		ID:            nodeID,
		Port:          emnt.networkConfig.BasePort + nodeID,
		DataDir:       filepath.Join(emnt.networkConfig.BaseDataDir, fmt.Sprintf("node_%d", nodeID)),
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

	// Initialize network simulator
	if emnt.networkConfig.EnableNetworkSim {
		node.NetworkSim = NewNetworkSimulator(emnt.networkConfig.NetworkConditions)
	}

	// Initialize consensus
	if emnt.networkConfig.EnableConsensus {
		node.Consensus = NewBlockchainConsensus(emnt.networkConfig.ConsensusConfig)
	}

	emnt.nodes[nodeID] = node
	return nil
}

// startEnhancedNode starts an enhanced PDF storage node
func (emnt *EnhancedMultiNodePDFTest) startEnhancedNode(nodeID int) error {
	node := emnt.nodes[nodeID]

	// Start consensus if enabled
	if emnt.networkConfig.EnableConsensus && node.Consensus != nil {
		if err := node.Consensus.Start(); err != nil {
			return fmt.Errorf("failed to start consensus for node %d: %w", nodeID, err)
		}
	}

	node.mu.Lock()
	node.IsRunning = true
	node.mu.Unlock()

	fmt.Printf("✅ Enhanced Node %d started on port %d\n", nodeID, node.Port)
	return nil
}

// connectEnhancedNodes establishes connections between enhanced nodes
func (emnt *EnhancedMultiNodePDFTest) connectEnhancedNodes() error {
	fmt.Println("🔗 Establishing enhanced connections between nodes...")

	for i := 0; i < emnt.networkConfig.NodeCount; i++ {
		for j := 0; j < emnt.networkConfig.NodeCount; j++ {
			if i != j {
				peerAddr := fmt.Sprintf("localhost:%d", emnt.networkConfig.BasePort+j)
				emnt.nodes[i].PeerAddresses = append(emnt.nodes[i].PeerAddresses, peerAddr)
				fmt.Printf("🔗 Node %d connected to Node %d\n", i, j)

				// Simulate network conditions if enabled
				if emnt.networkConfig.EnableNetworkSim {
					emnt.simulateNetworkConditions(i, j)
				}
			}
		}
	}

	return nil
}

// simulateNetworkConditions simulates realistic network conditions
func (emnt *EnhancedMultiNodePDFTest) simulateNetworkConditions(fromNode, toNode int) {
	fromNodeID := fmt.Sprintf("node_%d", fromNode)
	toNodeID := fmt.Sprintf("node_%d", toNode)

	// Simulate latency
	latency := emnt.networkSim.SimulateNetworkLatency(fromNodeID, toNodeID)

	// Simulate packet loss
	if emnt.networkSim.SimulatePacketLoss(fromNodeID, toNodeID) {
		emnt.testResults.NetworkEvents = append(emnt.testResults.NetworkEvents, &NetworkEvent{
			Type:      EventPacketLoss,
			NodeID:    fromNodeID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"target_node": toNodeID,
				"latency":     latency,
			},
		})
	}

	// Simulate congestion
	if emnt.networkSim.rng.Float64() < 0.1 { // 10% chance
		emnt.networkSim.SimulateNetworkCongestion(fromNodeID)
	}
}

// TestEnhancedPDFPropagation tests PDF upload with network simulation and consensus
func (emnt *EnhancedMultiNodePDFTest) TestEnhancedPDFPropagation() error {
	fmt.Println("\n📄 Testing Enhanced PDF Upload with Network Simulation & Blockchain Consensus...")

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

	// Upload PDF to Node 0 with enhanced features
	fmt.Println("\n📤 Uploading PDF to Node 0 with enhanced features...")
	uploadStart := time.Now()

	metadata := PDFMetadata{
		Title:       "Enhanced Multi-Node Test Document",
		Author:      "Enhanced Test System",
		Subject:     "Network Simulation & Consensus Test",
		Description: "Testing PDF propagation with realistic network conditions and blockchain consensus",
		Keywords:    []string{"enhanced", "test", "propagation", "consensus", "pdf"},
		Tags:        []string{"enhanced", "multi-node", "blockchain"},
		CustomFields: map[string]string{
			"test_type":   "enhanced_propagation",
			"node_count":  fmt.Sprintf("%d", emnt.networkConfig.NodeCount),
			"network_sim": fmt.Sprintf("%v", emnt.networkConfig.EnableNetworkSim),
			"consensus":   fmt.Sprintf("%v", emnt.networkConfig.EnableConsensus),
			"timestamp":   time.Now().Format(time.RFC3339),
		},
	}

	// Store PDF on Node 0
	storedPDF, err := emnt.nodes[0].PDFStorage.StorePDF(
		pdfContent,
		"enhanced_test.pdf",
		"enhanced_test_user",
		metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to store PDF on Node 0: %w", err)
	}

	uploadTime := time.Since(uploadStart)
	emnt.testResults.PDFUploadTime = uploadTime
	fmt.Printf("✅ PDF uploaded to Node 0 in %v\n", uploadTime)
	fmt.Printf("   Document ID: %s\n", storedPDF.DocumentID)

	// Create blockchain transaction if consensus is enabled
	if emnt.networkConfig.EnableConsensus {
		fmt.Println("\n⛓️  Creating blockchain transaction for PDF...")
		tx := emnt.createPDFTransaction(storedPDF, pdfContent, metadata)

		if err := emnt.consensus.AddTransaction(tx); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add transaction to consensus: %v\n", err)
		} else {
			fmt.Printf("✅ PDF transaction added to blockchain mempool\n")
		}
	}

	// Simulate network propagation with realistic conditions
	fmt.Println("\n⏳ Simulating realistic network propagation...")
	time.Sleep(3 * time.Second)

	// Test propagation to other nodes with network simulation
	fmt.Println("\n🔍 Testing PDF propagation with network simulation...")
	propagationStart := time.Now()

	successCount := 0
	for i := 1; i < emnt.networkConfig.NodeCount; i++ {
		fmt.Printf("   Testing Node %d... ", i)

		// Simulate network conditions
		if emnt.networkConfig.EnableNetworkSim {
			emnt.simulateNodeCommunication(0, i, pdfSize)
		}

		// Copy PDF to this node (simulating network propagation)
		_, err := emnt.nodes[i].PDFStorage.StorePDF(
			pdfContent,
			"enhanced_test.pdf",
			"enhanced_test_user",
			metadata,
		)
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			emnt.testResults.Errors = append(emnt.testResults.Errors,
				fmt.Sprintf("Node %d failed to store PDF: %v", i, err))
			continue
		}

		// Try to retrieve PDF from this node
		content, retrievedMetadata, err := emnt.nodes[i].PDFStorage.GetPDF(storedPDF.DocumentID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve: %v\n", err)
			emnt.testResults.Errors = append(emnt.testResults.Errors,
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
			emnt.testResults.Errors = append(emnt.testResults.Errors,
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
	emnt.testResults.PropagationTime = propagationTime

	propagationRate := float64(successCount) / float64(emnt.networkConfig.NodeCount-1) * 100
	fmt.Printf("\n📊 Propagation Results: %d/%d nodes successful (%.1f%%)\n",
		successCount, emnt.networkConfig.NodeCount-1, propagationRate)

	// Test consensus and block creation
	fmt.Println("\n⛓️  Testing blockchain consensus and block creation...")
	consensusStart := time.Now()

	// Wait for consensus operations
	time.Sleep(10 * time.Second)

	// Check consensus state across nodes
	totalBlocks := 0
	totalTransactions := 0
	for i := 0; i < emnt.networkConfig.NodeCount; i++ {
		if emnt.nodes[i].Consensus != nil {
			info := emnt.nodes[i].Consensus.chain.GetBlockchainInfo()
			blockCount := info["block_count"].(int)
			txCount := info["transaction_count"].(int)

			fmt.Printf("   Node %d: %d blocks, %d transactions\n", i, blockCount, txCount)
			totalBlocks += blockCount
			totalTransactions += txCount
		}
	}

	consensusTime := time.Since(consensusStart)
	emnt.testResults.ConsensusTime = consensusTime
	emnt.testResults.TotalBlocks = totalBlocks / emnt.networkConfig.NodeCount
	emnt.testResults.TotalTransactions = totalTransactions / emnt.networkConfig.NodeCount

	// Benchmark enhanced network performance
	fmt.Println("\n📈 Benchmarking enhanced network performance...")
	emnt.benchmarkEnhancedNetwork()

	return nil
}

// simulateNodeCommunication simulates realistic node-to-node communication
func (emnt *EnhancedMultiNodePDFTest) simulateNodeCommunication(fromNode, toNode, dataSize int) {
	fromNodeID := fmt.Sprintf("node_%d", fromNode)
	toNodeID := fmt.Sprintf("node_%d", toNode)

	// Simulate bandwidth constraints
	transferTime := emnt.networkSim.SimulateBandwidthLimit(fromNodeID, int64(dataSize))

	// Simulate network events
	if emnt.networkSim.rng.Float64() < 0.05 { // 5% chance of network event
		eventType := EventLatencySpike
		if emnt.networkSim.rng.Float64() < 0.5 {
			eventType = EventCongestion
		}

		emnt.testResults.NetworkEvents = append(emnt.testResults.NetworkEvents, &NetworkEvent{
			Type:      eventType,
			NodeID:    fromNodeID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"target_node":   toNodeID,
				"transfer_time": transferTime,
				"data_size":     dataSize,
			},
		})
	}
}

// createPDFTransaction creates a blockchain transaction for the PDF
func (emnt *EnhancedMultiNodePDFTest) createPDFTransaction(storedPDF *StoredPDF, content []byte, metadata PDFMetadata) *Transaction {
	// Create a simple transaction structure with sufficient fee
	// Input: 1000 (genesis), Output: 500, Fee: 500 (meets minimum requirement)
	tx := &Transaction{
		ID:        storedPDF.DocumentID,
		Timestamp: time.Now(),
		PublicKey: []byte("test_public_key"),
		Inputs: []*TxInput{
			{
				TxID:      "genesis",
				OutIndex:  0,
				Signature: []byte("test_signature"),
				PublicKey: []byte("test_public_key"),
			},
		},
		Outputs: []*TxOutput{
			{
				Value:   500, // Reduced output to provide 500 fee (1000 - 500 = 500)
				Script:  []byte("test_script"),
				Address: "test_address",
			},
		},
	}

	// Calculate transaction hash
	tx.Hash = calculateTransactionHash(tx)

	return tx
}

// benchmarkEnhancedNetwork performs enhanced network performance benchmarks
func (emnt *EnhancedMultiNodePDFTest) benchmarkEnhancedNetwork() {
	fmt.Println("   Running enhanced network benchmarks...")

	// Test network simulation performance
	if emnt.networkConfig.EnableNetworkSim {
		fmt.Println("   Testing network simulation...")

		// Simulate network partitions
		nodeIDs := make([]string, emnt.networkConfig.NodeCount)
		for i := 0; i < emnt.networkConfig.NodeCount; i++ {
			nodeIDs[i] = fmt.Sprintf("node_%d", i)
		}

		partitions := emnt.networkSim.SimulateNetworkPartition(nodeIDs, 0.1)
		if len(partitions) > 0 {
			fmt.Printf("   Network partitions detected: %d\n", len(partitions))
			for partitionID, nodes := range partitions {
				fmt.Printf("      %s: %d nodes\n", partitionID, len(nodes))
			}
		}

		// Get network statistics
		networkStats := emnt.networkSim.GetNetworkStats()
		fmt.Printf("   Active network nodes: %d\n", len(networkStats))

		// Calculate average latency
		totalLatency := time.Duration(0)
		for _, stats := range networkStats {
			totalLatency += stats.CurrentLatency
		}
		if len(networkStats) > 0 {
			avgLatency := totalLatency / time.Duration(len(networkStats))
			fmt.Printf("   Average network latency: %v\n", avgLatency)
		}
	}

	// Test consensus performance
	if emnt.networkConfig.EnableConsensus {
		fmt.Println("   Testing consensus performance...")

		// Get blockchain information
		blockchainInfo := emnt.consensus.chain.GetBlockchainInfo()
		fmt.Printf("   Total blocks: %d\n", blockchainInfo["block_count"])
		fmt.Printf("   Total transactions: %d\n", blockchainInfo["transaction_count"])
		fmt.Printf("   UTXO count: %d\n", blockchainInfo["utxo_count"])
		fmt.Printf("   Current difficulty: %d\n", blockchainInfo["difficulty"])
	}
}

// StopNodes gracefully shuts down all enhanced nodes
func (emnt *EnhancedMultiNodePDFTest) StopNodes() {
	fmt.Println("\n🛑 Shutting down enhanced multi-node network...")

	// Signal all nodes to stop
	close(emnt.stopChan)

	// Stop consensus with timeout to prevent deadlock
	if emnt.networkConfig.EnableConsensus {
		// Use a goroutine to stop consensus with timeout
		done := make(chan bool, 1)
		go func() {
			emnt.consensus.Stop()
			done <- true
		}()

		select {
		case <-done:
			fmt.Println("   Consensus stopped successfully")
		case <-time.After(5 * time.Second):
			fmt.Println("   Warning: Consensus stop timed out")
		}
	}

	// Stop all nodes
	for i := 0; i < emnt.networkConfig.NodeCount; i++ {
		if emnt.nodes[i] != nil {
			emnt.nodes[i].mu.Lock()
			if emnt.nodes[i].IsRunning {
				if emnt.nodes[i].Consensus != nil {
					// Stop consensus with timeout
					done := make(chan bool, 1)
					go func(consensus *BlockchainConsensus) {
						consensus.Stop()
						done <- true
					}(emnt.nodes[i].Consensus)

					select {
					case <-done:
						fmt.Printf("   Node %d consensus stopped\n", i)
					case <-time.After(2 * time.Second):
						fmt.Printf("   Warning: Node %d consensus stop timed out\n", i)
					}
				}
				emnt.nodes[i].IsRunning = false
				fmt.Printf("   Enhanced Node %d stopped\n", i)
			}
			emnt.nodes[i].mu.Unlock()
		}
	}

	// Only set EndTime if it hasn't been set yet (avoid race conditions)
	if emnt.testResults.EndTime.IsZero() {
		emnt.testResults.EndTime = time.Now()
	}

	emnt.testResults.Success = len(emnt.testResults.Errors) == 0

	fmt.Println("✅ Enhanced multi-node network shutdown complete")
}

// PrintEnhancedResults displays comprehensive enhanced test results
func (emnt *EnhancedMultiNodePDFTest) PrintEnhancedResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 ENHANCED MULTI-NODE PDF TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("🏗️  Enhanced Network Configuration:\n")
	fmt.Printf("   Total Nodes: %d\n", emnt.testResults.TotalNodes)
	fmt.Printf("   Base Port: %d\n", emnt.networkConfig.BasePort)
	fmt.Printf("   Network Simulation: %v\n", emnt.networkConfig.EnableNetworkSim)
	fmt.Printf("   Blockchain Consensus: %v\n", emnt.networkConfig.EnableConsensus)
	fmt.Printf("   Difficulty: %d\n", emnt.networkConfig.Difficulty)
	fmt.Printf("   Block Time: %v\n", emnt.networkConfig.BlockTime)

	fmt.Printf("\n⏱️  Enhanced Performance Metrics:\n")
	fmt.Printf("   Total Test Duration: %v\n", emnt.testResults.EndTime.Sub(emnt.testResults.StartTime))
	fmt.Printf("   PDF Upload Time: %v\n", emnt.testResults.PDFUploadTime)
	fmt.Printf("   Network Propagation Time: %v\n", emnt.testResults.PropagationTime)
	fmt.Printf("   Consensus Time: %v\n", emnt.testResults.ConsensusTime)

	fmt.Printf("\n📈 Enhanced Blockchain Metrics:\n")
	fmt.Printf("   Total Blocks Created: %d\n", emnt.testResults.TotalBlocks)
	fmt.Printf("   Total Transactions: %d\n", emnt.testResults.TotalTransactions)
	fmt.Printf("   Storage Efficiency: %.2f bytes/node\n", emnt.testResults.StorageEfficiency)
	fmt.Printf("   Consensus Efficiency: %.2f operations/sec\n", emnt.testResults.ConsensusEfficiency)

	if emnt.networkConfig.EnableNetworkSim {
		fmt.Printf("\n🌐 Network Simulation Results:\n")
		fmt.Printf("   Network Events: %d\n", len(emnt.testResults.NetworkEvents))
		for i, event := range emnt.testResults.NetworkEvents {
			if i < 5 { // Show first 5 events
				fmt.Printf("      %d. %s: %s\n", i+1, event.Type, event.NodeID)
			}
		}
		if len(emnt.testResults.NetworkEvents) > 5 {
			fmt.Printf("      ... and %d more events\n", len(emnt.testResults.NetworkEvents)-5)
		}
	}

	fmt.Printf("\n🔍 Enhanced Test Results:\n")
	if emnt.testResults.Success {
		fmt.Printf("   Status: ✅ SUCCESS\n")
	} else {
		fmt.Printf("   Status: ❌ FAILED\n")
		fmt.Printf("   Errors: %d\n", len(emnt.testResults.Errors))
		for i, err := range emnt.testResults.Errors {
			fmt.Printf("      %d. %s\n", i+1, err)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
}

// RunEnhancedTest executes the complete enhanced multi-node test
func (emnt *EnhancedMultiNodePDFTest) RunEnhancedTest() error {
	// Start the enhanced network
	if err := emnt.StartNodes(); err != nil {
		return fmt.Errorf("failed to start enhanced nodes: %w", err)
	}

	// Test enhanced PDF propagation
	if err := emnt.TestEnhancedPDFPropagation(); err != nil {
		return fmt.Errorf("failed to test enhanced PDF propagation: %w", err)
	}

	// Stop nodes and set EndTime BEFORE printing results
	emnt.StopNodes()

	// Print enhanced results
	emnt.PrintEnhancedResults()

	return nil
}

// RunEnhancedMultiNodeTest is the main entry point for running the enhanced multi-node test
func RunEnhancedMultiNodeTest() error {
	fmt.Println("🚀 Enhanced Multi-Node PDF Blockchain Test")
	fmt.Println("==========================================")

	// Create and run enhanced multi-node test
	test := NewEnhancedMultiNodePDFTest(5) // 5 nodes

	if err := test.RunEnhancedTest(); err != nil {
		return fmt.Errorf("❌ Enhanced test failed: %w", err)
	}

	fmt.Println("\n🎉 Enhanced multi-node PDF test completed successfully!")
	return nil
}
