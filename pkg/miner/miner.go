package miner

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/mempool"
)

// Miner represents a blockchain miner
type Miner struct {
	mu            sync.RWMutex
	chain         *chain.Chain
	mempool       *mempool.Mempool
	config        *MinerConfig
	isMining      bool
	stopMining    chan struct{}
	currentBlock  *block.Block
	ctx           context.Context
	cancel        context.CancelFunc
}

// MinerConfig holds configuration for the miner
type MinerConfig struct {
	MiningEnabled     bool
	MiningThreads     int
	BlockTime         time.Duration
	MaxBlockSize      uint64
	CoinbaseAddress   string
	CoinbaseReward    uint64
}

// DefaultMinerConfig returns the default miner configuration
func DefaultMinerConfig() *MinerConfig {
	return &MinerConfig{
		MiningEnabled:   true,
		MiningThreads:   1,
		BlockTime:       10 * time.Second,
		MaxBlockSize:    1000000, // 1MB
		CoinbaseAddress: "",
		CoinbaseReward:  1000000000, // 1 billion units
	}
}

// NewMiner creates a new miner
func NewMiner(chain *chain.Chain, mempool *mempool.Mempool, config *MinerConfig) *Miner {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Miner{
		chain:      chain,
		mempool:    mempool,
		config:     config,
		stopMining: make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// StartMining starts the mining process
func (m *Miner) StartMining() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.isMining {
		return fmt.Errorf("mining already in progress")
	}
	
	m.isMining = true
	m.stopMining = make(chan struct{})
	
	// Start mining in a goroutine
	go m.mineBlocks()
	
	return nil
}

// StopMining stops the mining process
func (m *Miner) StopMining() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.isMining {
		return
	}
	
	m.isMining = false
	close(m.stopMining)
}

// IsMining returns whether the miner is currently mining
func (m *Miner) IsMining() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.isMining
}

// mineBlocks continuously mines new blocks
func (m *Miner) mineBlocks() {
	ticker := time.NewTicker(m.config.BlockTime)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.stopMining:
			return
		case <-ticker.C:
			// Try to mine a new block
			if err := m.mineNextBlock(); err != nil {
				// Log error but continue mining
				fmt.Printf("Mining error: %v\n", err)
			}
		}
	}
}

// mineNextBlock mines the next block
func (m *Miner) mineNextBlock() error {
	// Get the current best block
	bestBlock := m.chain.GetBestBlock()
	if bestBlock == nil {
		return fmt.Errorf("no best block available")
	}
	
	// Create a new block
	newBlock := m.createNewBlock(bestBlock)
	if newBlock == nil {
		return fmt.Errorf("failed to create new block")
	}
	
	// Mine the block
	if err := m.mineBlock(newBlock); err != nil {
		return fmt.Errorf("failed to mine block: %w", err)
	}
	
	// Add the block to the chain
	if err := m.chain.AddBlock(newBlock); err != nil {
		return fmt.Errorf("failed to add block to chain: %w", err)
	}
	
	fmt.Printf("Mined new block: %s\n", newBlock.String())
	
	return nil
}

// createNewBlock creates a new block for mining
func (m *Miner) createNewBlock(prevBlock *block.Block) *block.Block {
	// Get transactions from mempool
	transactions := m.mempool.GetTransactionsForBlock(m.config.MaxBlockSize)
	
	// Create coinbase transaction
	coinbaseTx := m.createCoinbaseTransaction(prevBlock.Header.Height + 1)
	
	// Create new block
	newBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevBlock.CalculateHash(),
			MerkleRoot:    nil, // Will be calculated after adding transactions
			Timestamp:     time.Now(),
			Difficulty:    m.chain.CalculateNextDifficulty(),
			Nonce:         0,
			Height:        prevBlock.Header.Height + 1,
		},
		Transactions: make([]*block.Transaction, 0),
	}
	
	// Add coinbase transaction first
	newBlock.AddTransaction(coinbaseTx)
	
	// Add other transactions
	for _, tx := range transactions {
		newBlock.AddTransaction(tx)
	}
	
	// Calculate Merkle root
	newBlock.Header.MerkleRoot = newBlock.CalculateMerkleRoot()
	
	return newBlock
}

// createCoinbaseTransaction creates a coinbase transaction
func (m *Miner) createCoinbaseTransaction(height uint64) *block.Transaction {
	// Calculate total fees from transactions
	totalFees := uint64(0)
	for _, tx := range m.currentBlock.Transactions {
		if tx != nil {
			totalFees += tx.Fee
		}
	}
	
	// Create coinbase output
	output := &block.TxOutput{
		Value:        m.config.CoinbaseReward + totalFees,
		ScriptPubKey: []byte(m.config.CoinbaseAddress),
	}
	
	// Create transaction
	tx := &block.Transaction{
		Version:  1,
		Inputs:   make([]*block.TxInput, 0), // Coinbase has no inputs
		Outputs:  []*block.TxOutput{output},
		LockTime: 0,
		Fee:      0,
	}
	
	// Calculate transaction hash
	tx.Hash = m.calculateTransactionHash(tx)
	
	return tx
}

// mineBlock performs proof-of-work mining on a block
func (m *Miner) mineBlock(block *block.Block) error {
	target := m.calculateTarget(block.Header.Difficulty)
	
	// Try different nonces
	for nonce := uint64(0); nonce < ^uint64(0); nonce++ {
		select {
		case <-m.ctx.Done():
			return fmt.Errorf("mining cancelled")
		case <-m.stopMining:
			return fmt.Errorf("mining stopped")
		default:
			// Continue mining
		}
		
		// Set nonce
		block.Header.Nonce = nonce
		
		// Calculate hash
		hash := block.CalculateHash()
		
		// Check if hash meets target
		if m.hashLessThan(hash, target) {
			return nil // Block mined successfully
		}
	}
	
	return fmt.Errorf("failed to find valid nonce")
}

// calculateTarget calculates the target hash for a given difficulty
func (m *Miner) calculateTarget(difficulty uint64) []byte {
	// Simple difficulty calculation
	// Target = 2^(256-difficulty)
	target := make([]byte, 32)
	
	if difficulty >= 256 {
		// Maximum difficulty: all zeros
		return target
	}
	
	// Set the target based on difficulty
	byteIndex := difficulty / 8
	bitIndex := difficulty % 8
	
	if byteIndex < 32 {
		target[byteIndex] = 1 << bitIndex
	}
	
	return target
}

// hashLessThan checks if hash1 is less than hash2
func (m *Miner) hashLessThan(hash1, hash2 []byte) bool {
	for i := 0; i < len(hash1); i++ {
		if hash1[i] < hash2[i] {
			return true
		}
		if hash1[i] > hash2[i] {
			return false
		}
	}
	return false
}

// calculateTransactionHash calculates the hash of a transaction
func (m *Miner) calculateTransactionHash(tx *block.Transaction) []byte {
	data := make([]byte, 0)
	
	// Version
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)
	
	// Inputs
	for _, input := range tx.Inputs {
		data = append(data, input.PrevTxHash...)
		indexBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(indexBytes, input.PrevTxIndex)
		data = append(data, indexBytes...)
		data = append(data, input.ScriptSig...)
		seqBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(seqBytes, input.Sequence)
		data = append(data, seqBytes...)
	}
	
	// Outputs
	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(valueBytes, output.Value)
		data = append(data, valueBytes...)
		data = append(data, output.ScriptPubKey...)
	}
	
	// Lock time
	lockTimeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)
	
	// Fee
	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	data = append(data, feeBytes...)
	
	hash := sha256.Sum256(data)
	return hash[:]
}

// GetCurrentBlock returns the current block being mined
func (m *Miner) GetCurrentBlock() *block.Block {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.currentBlock
}

// GetMiningStats returns mining statistics
func (m *Miner) GetMiningStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["isMining"] = m.isMining
	stats["currentBlock"] = m.currentBlock
	stats["difficulty"] = m.chain.GetBestBlock().Header.Difficulty
	stats["height"] = m.chain.GetHeight()
	
	return stats
}

// Close closes the miner
func (m *Miner) Close() error {
	m.StopMining()
	m.cancel()
	return nil
}

// String returns a string representation of the miner
func (m *Miner) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return fmt.Sprintf("Miner{Mining: %t, Threads: %d, BlockTime: %v}", 
		m.isMining, m.config.MiningThreads, m.config.BlockTime)
} 