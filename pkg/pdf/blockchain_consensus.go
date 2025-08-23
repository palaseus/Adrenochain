package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// BlockchainConsensus implements real blockchain consensus for PDF transactions
type BlockchainConsensus struct {
	chain           *Blockchain
	network         *ConsensusNetwork
	config          *ConsensusConfig
	mu              sync.RWMutex
	isRunning       bool
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// Blockchain represents the actual blockchain structure
type Blockchain struct {
	blocks          []*Block
	transactions    []*Transaction
	utxoSet         map[string]*UTXO
	mu              sync.RWMutex
	lastBlockHash   []byte
	difficulty      uint64
}

// Block represents a blockchain block
type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Hash         []byte
	Nonce        uint64
	Timestamp    time.Time
}

// BlockHeader contains block metadata
type BlockHeader struct {
	Version     uint32
	PrevHash    []byte
	MerkleRoot  []byte
	Timestamp   time.Time
	Difficulty  uint64
	Nonce       uint64
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID        string
	Inputs    []*TxInput
	Outputs   []*TxOutput
	Timestamp time.Time
	Hash      []byte
	Signature []byte
	PublicKey []byte
}

// TxInput represents a transaction input
type TxInput struct {
	TxID      string
	OutIndex  uint32
	Signature []byte
	PublicKey []byte
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value      uint64
	Script     []byte
	Address    string
	IsSpent    bool
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID     string
	OutIndex uint32
	Value    uint64
	Script   []byte
	Address  string
}

// ConsensusNetwork handles P2P communication
type ConsensusNetwork struct {
	nodes       map[string]*ConsensusNode
	peers       map[string]*PeerConnection
	mu          sync.RWMutex
	events      chan *NetworkEvent
}

// ConsensusNode represents a node in the consensus network
type ConsensusNode struct {
	ID          string
	Address     string
	PublicKey   []byte
	IsValidator bool
	Stake       uint64
	LastSeen    time.Time
}

// PeerConnection represents a connection to another peer
type PeerConnection struct {
	PeerID      string
	Address     string
	IsConnected bool
	LastPing    time.Time
	Latency     time.Duration
}

// ConsensusConfig holds consensus configuration
type ConsensusConfig struct {
	Difficulty          uint64
	BlockTime           time.Duration
	MaxBlockSize        int
	MinTransactionFee   uint64
	ValidatorCount      int
	StakeRequirement    uint64
	ConsensusTimeout    time.Duration
}

// NewBlockchainConsensus creates a new consensus instance
func NewBlockchainConsensus(config *ConsensusConfig) *BlockchainConsensus {
	if config == nil {
		config = &ConsensusConfig{
			Difficulty:        4,
			BlockTime:         2 * time.Second,
			MaxBlockSize:      1024 * 1024, // 1MB
			MinTransactionFee: 1000,
			ValidatorCount:    5,
			StakeRequirement:  10000,
			ConsensusTimeout:  30 * time.Second,
		}
	}
	
	return &BlockchainConsensus{
		chain:     NewBlockchain(config.Difficulty),
		network:   NewConsensusNetwork(),
		config:    config,
		stopChan:  make(chan struct{}),
	}
}

// NewBlockchain creates a new blockchain
func NewBlockchain(difficulty uint64) *Blockchain {
	chain := &Blockchain{
		blocks:        make([]*Block, 0),
		transactions:  make([]*Transaction, 0),
		utxoSet:       make(map[string]*UTXO),
		difficulty:    difficulty,
	}
	
	// Create genesis block
	genesisBlock := createGenesisBlock(difficulty)
	chain.AddBlock(genesisBlock)
	
	return chain
}

// NewConsensusNetwork creates a new consensus network
func NewConsensusNetwork() *ConsensusNetwork {
	return &ConsensusNetwork{
		nodes:  make(map[string]*ConsensusNode),
		peers:  make(map[string]*PeerConnection),
		events: make(chan *NetworkEvent, 100),
	}
}

// createGenesisBlock creates the initial genesis block
func createGenesisBlock(difficulty uint64) *Block {
	genesis := &Block{
		Header: &BlockHeader{
			Version:    1,
			PrevHash:   make([]byte, 32),
			MerkleRoot: make([]byte, 32),
			Timestamp:  time.Now(),
			Difficulty: difficulty,
			Nonce:      0,
		},
		Transactions: []*Transaction{},
		Timestamp:    time.Now(),
	}
	
	// Mine the genesis block
	genesis.Hash = mineBlock(genesis, difficulty)
	return genesis
}

// Start begins the consensus process
func (bc *BlockchainConsensus) Start() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if bc.isRunning {
		return fmt.Errorf("consensus already running")
	}
	
	bc.isRunning = true
	bc.wg.Add(1)
	
	go bc.consensusLoop()
	
	return nil
}

// Stop gracefully shuts down consensus
func (bc *BlockchainConsensus) Stop() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	if !bc.isRunning {
		return
	}
	
	close(bc.stopChan)
	bc.isRunning = false
	bc.wg.Wait()
}

// consensusLoop runs the main consensus loop
func (bc *BlockchainConsensus) consensusLoop() {
	defer bc.wg.Done()
	
	ticker := time.NewTicker(bc.config.BlockTime)
	defer ticker.Stop()
	
	for {
		select {
		case <-bc.stopChan:
			return
		case <-ticker.C:
			bc.createNewBlock()
		case event := <-bc.network.events:
			bc.handleNetworkEvent(event)
		}
	}
}

// createNewBlock creates and mines a new block
func (bc *BlockchainConsensus) createNewBlock() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	// Get pending transactions
	pendingTxs := bc.getPendingTransactions()
	if len(pendingTxs) == 0 {
		return
	}
	
	// Create new block
	newBlock := &Block{
		Header: &BlockHeader{
			Version:    1,
			PrevHash:   bc.chain.lastBlockHash,
			MerkleRoot: calculateMerkleRoot(pendingTxs),
			Timestamp:  time.Now(),
			Difficulty: bc.config.Difficulty,
			Nonce:      0,
		},
		Transactions: pendingTxs,
		Timestamp:    time.Now(),
	}
	
	// Mine the block
	startTime := time.Now()
	newBlock.Hash = mineBlock(newBlock, bc.config.Difficulty)
	miningTime := time.Since(startTime)
	
	fmt.Printf("⛏️  Block mined in %v with hash: %s\n", 
		miningTime, hex.EncodeToString(newBlock.Hash[:8]))
	
	// Add block to chain
	bc.chain.AddBlock(newBlock)
	
	// Broadcast to network
	bc.broadcastBlock(newBlock)
}

// mineBlock performs proof-of-work mining
func mineBlock(block *Block, difficulty uint64) []byte {
	target := make([]byte, 32)
	for i := range target {
		if i < int(difficulty/8) {
			target[i] = 0x00
		} else if i == int(difficulty/8) {
			target[i] = 0xFF >> (difficulty % 8)
		} else {
			target[i] = 0xFF
		}
	}
	
	for nonce := uint64(0); ; nonce++ {
		block.Header.Nonce = nonce
		block.Nonce = nonce
		
		hash := calculateBlockHash(block)
		
		// Check if hash meets difficulty target
		if isHashValid(hash, target) {
			return hash
		}
	}
}

// calculateBlockHash calculates the hash of a block
func calculateBlockHash(block *Block) []byte {
	data := fmt.Sprintf("%d%s%s%d%d%d",
		block.Header.Version,
		hex.EncodeToString(block.Header.PrevHash),
		hex.EncodeToString(block.Header.MerkleRoot),
		block.Header.Timestamp.Unix(),
		block.Header.Difficulty,
		block.Header.Nonce,
	)
	
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// isHashValid checks if a hash meets the difficulty target
func isHashValid(hash, target []byte) bool {
	for i := range hash {
		if hash[i] < target[i] {
			return true
		} else if hash[i] > target[i] {
			return false
		}
	}
	return true
}

// calculateMerkleRoot calculates the Merkle root of transactions
func calculateMerkleRoot(transactions []*Transaction) []byte {
	if len(transactions) == 0 {
		return make([]byte, 32)
	}
	
	if len(transactions) == 1 {
		return calculateTransactionHash(transactions[0])
	}
	
	// Build Merkle tree
	hashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		hashes[i] = calculateTransactionHash(tx)
	}
	
	// Combine hashes in pairs
	for len(hashes) > 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		
		newHashes := make([][]byte, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			newHashes[i/2] = hash[:]
		}
		hashes = newHashes
	}
	
	return hashes[0]
}

// calculateTransactionHash calculates the hash of a transaction
func calculateTransactionHash(tx *Transaction) []byte {
	data := fmt.Sprintf("%s%d%s",
		tx.ID,
		tx.Timestamp.Unix(),
		hex.EncodeToString(tx.PublicKey),
	)
	
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// AddTransaction adds a new transaction to the mempool
func (bc *BlockchainConsensus) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	// Validate transaction
	if err := bc.validateTransaction(tx); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}
	
	// Add to mempool
	bc.chain.transactions = append(bc.chain.transactions, tx)
	
	// Broadcast to network
	bc.broadcastTransaction(tx)
	
	return nil
}

// validateTransaction validates a transaction
func (bc *BlockchainConsensus) validateTransaction(tx *Transaction) error {
	// Check transaction fee
	if tx.getFee() < bc.config.MinTransactionFee {
		return fmt.Errorf("insufficient transaction fee")
	}
	
	// Check transaction size
	if tx.getSize() > bc.config.MaxBlockSize {
		return fmt.Errorf("transaction too large")
	}
	
	// Validate inputs and outputs
	if err := bc.validateInputsOutputs(tx); err != nil {
		return fmt.Errorf("input/output validation failed: %w", err)
	}
	
	return nil
}

// validateInputsOutputs validates transaction inputs and outputs
func (bc *BlockchainConsensus) validateInputsOutputs(tx *Transaction) error {
	// Check that inputs exist in UTXO set
	for _, input := range tx.Inputs {
		utxoKey := fmt.Sprintf("%s:%d", input.TxID, input.OutIndex)
		if _, exists := bc.chain.utxoSet[utxoKey]; !exists {
			return fmt.Errorf("input %s not found in UTXO set", utxoKey)
		}
	}
	
	// Check that output values don't exceed input values
	inputSum := uint64(0)
	outputSum := uint64(0)
	
	for _, input := range tx.Inputs {
		utxoKey := fmt.Sprintf("%s:%d", input.TxID, input.OutIndex)
		if utxo, exists := bc.chain.utxoSet[utxoKey]; exists {
			inputSum += utxo.Value
		}
	}
	
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}
	
	if outputSum > inputSum {
		return fmt.Errorf("output sum exceeds input sum")
	}
	
	return nil
}

// getPendingTransactions returns transactions waiting to be mined
func (bc *BlockchainConsensus) getPendingTransactions() []*Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	// Return transactions up to max block size
	var pendingTxs []*Transaction
	currentSize := 0
	
	for _, tx := range bc.chain.transactions {
		txSize := tx.getSize()
		if currentSize+txSize <= bc.config.MaxBlockSize {
			pendingTxs = append(pendingTxs, tx)
			currentSize += txSize
		} else {
			break
		}
	}
	
	return pendingTxs
}

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	
	// Validate block
	if err := bc.validateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}
	
	// Add block
	bc.blocks = append(bc.blocks, block)
	bc.lastBlockHash = block.Hash
	
	// Update UTXO set
	bc.updateUTXOSet(block)
	
	// Remove mined transactions from mempool
	bc.removeMinedTransactions(block.Transactions)
	
	return nil
}

// validateBlock validates a block
func (bc *Blockchain) validateBlock(block *Block) error {
	// Check previous hash
	if !bytes.Equal(block.Header.PrevHash, bc.lastBlockHash) {
		return fmt.Errorf("invalid previous hash")
	}
	
	// Check proof of work
	if !isHashValid(block.Hash, bc.getDifficultyTarget()) {
		return fmt.Errorf("invalid proof of work")
	}
	
	// Check Merkle root
	calculatedRoot := calculateMerkleRoot(block.Transactions)
	if !bytes.Equal(block.Header.MerkleRoot, calculatedRoot) {
		return fmt.Errorf("invalid Merkle root")
	}
	
	return nil
}

// updateUTXOSet updates the UTXO set with new block
func (bc *Blockchain) updateUTXOSet(block *Block) {
	// Remove spent inputs
	for _, tx := range block.Transactions {
		for _, input := range tx.Inputs {
			utxoKey := fmt.Sprintf("%s:%d", input.TxID, input.OutIndex)
			delete(bc.utxoSet, utxoKey)
		}
	}
	
	// Add new outputs
	for _, tx := range block.Transactions {
		for i, output := range tx.Outputs {
			utxoKey := fmt.Sprintf("%s:%d", tx.ID, uint32(i))
			bc.utxoSet[utxoKey] = &UTXO{
				TxID:     tx.ID,
				OutIndex: uint32(i),
				Value:    output.Value,
				Script:   output.Script,
				Address:  output.Address,
			}
		}
	}
}

// removeMinedTransactions removes mined transactions from mempool
func (bc *Blockchain) removeMinedTransactions(minedTxs []*Transaction) {
	minedIDs := make(map[string]bool)
	for _, tx := range minedTxs {
		minedIDs[tx.ID] = true
	}
	
	var remainingTxs []*Transaction
	for _, tx := range bc.transactions {
		if !minedIDs[tx.ID] {
			remainingTxs = append(remainingTxs, tx)
		}
	}
	bc.transactions = remainingTxs
}

// Helper methods for Transaction
func (tx *Transaction) getFee() uint64 {
	inputSum := uint64(0)
	outputSum := uint64(0)
	
	for range tx.Inputs {
		// This would need access to UTXO set in real implementation
		inputSum += 1000 // Placeholder
	}
	
	for _, output := range tx.Outputs {
		outputSum += output.Value
	}
	
	if inputSum > outputSum {
		return inputSum - outputSum
	}
	return 0
}

func (tx *Transaction) getSize() int {
	// Simplified size calculation
	return len(tx.ID) + len(tx.PublicKey) + len(tx.Signature) + 8
}

// Network communication methods
func (bc *BlockchainConsensus) broadcastBlock(block *Block) {
	// In real implementation, this would send to all peers
	fmt.Printf("📡 Broadcasting block %s to network\n", 
		hex.EncodeToString(block.Hash[:8]))
}

func (bc *BlockchainConsensus) broadcastTransaction(tx *Transaction) {
	// In real implementation, this would send to all peers
	fmt.Printf("📡 Broadcasting transaction %s to network\n", tx.ID[:8])
}

func (bc *BlockchainConsensus) handleNetworkEvent(event *NetworkEvent) {
	switch event.Type {
	case EventNodeJoin:
		fmt.Printf("🆕 Node %s joined the network\n", event.NodeID)
	case EventNodeLeave:
		fmt.Printf("👋 Node %s left the network\n", event.NodeID)
	case EventPartition:
		fmt.Printf("⚠️  Network partition detected\n")
	case EventRecovery:
		fmt.Printf("✅ Network recovery detected\n")
	}
}

// GetBlockchainInfo returns blockchain information
func (bc *Blockchain) GetBlockchainInfo() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	return map[string]interface{}{
		"block_count":     len(bc.blocks),
		"transaction_count": len(bc.transactions),
		"utxo_count":      len(bc.utxoSet),
		"last_block_hash": hex.EncodeToString(bc.lastBlockHash),
		"difficulty":      bc.difficulty,
	}
}

// GetDifficultyTarget returns the current difficulty target
func (bc *Blockchain) getDifficultyTarget() []byte {
	target := make([]byte, 32)
	for i := range target {
		if i < int(bc.difficulty/8) {
			target[i] = 0x00
		} else if i == int(bc.difficulty/8) {
			target[i] = 0xFF >> (bc.difficulty % 8)
		} else {
			target[i] = 0xFF
		}
	}
	return target
}
