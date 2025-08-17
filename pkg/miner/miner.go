package miner

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/mempool"
)

// Miner represents a blockchain miner
type Miner struct {
	mu           sync.RWMutex
	chain        *chain.Chain
	mempool      *mempool.Mempool
	config       *MinerConfig
	isMining     bool
	stopMining   chan struct{}
	currentBlock *block.Block
	ctx          context.Context
	cancel       context.CancelFunc
	consensus    *consensus.Consensus
	onBlockMined func(*block.Block) // Callback for when a block is successfully mined
}

// MinerConfig holds configuration for the miner
type MinerConfig struct {
	MiningEnabled   bool
	MiningThreads   int
	BlockTime       time.Duration
	MaxBlockSize    uint64
	CoinbaseAddress string
	CoinbaseReward  uint64
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
func NewMiner(chain *chain.Chain, mempool *mempool.Mempool, config *MinerConfig, consensusConfig *consensus.ConsensusConfig) *Miner {
	ctx, cancel := context.WithCancel(context.Background())

	return &Miner{
		chain:      chain,
		mempool:    mempool,
		config:     config,
		stopMining: make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
		consensus:  consensus.NewConsensus(consensusConfig, chain),
	}
}

// SetOnBlockMined sets the callback function for when a block is successfully mined
func (m *Miner) SetOnBlockMined(callback func(*block.Block)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onBlockMined = callback
}

// StartMining starts the mining process
func (m *Miner) StartMining() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isMining {
		// Stop current mining first
		m.isMining = false
		if m.stopMining != nil {
			select {
			case <-m.stopMining:
				// Channel already closed
			default:
				close(m.stopMining)
			}
		}
		// Wait for the goroutine to stop
		time.Sleep(50 * time.Millisecond)
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
	
	// Signal the mining goroutine to stop
	if m.stopMining != nil {
		select {
		case <-m.stopMining:
			// Channel already closed
		default:
			close(m.stopMining)
		}
	}
}

// Cleanup ensures the miner is properly stopped and cleaned up
func (m *Miner) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isMining {
		m.isMining = false
		if m.stopMining != nil {
			select {
			case <-m.stopMining:
				// Channel already closed
			default:
				close(m.stopMining)
			}
		}
		// Wait for goroutine to stop
		time.Sleep(100 * time.Millisecond)
	}
}

// IsMining returns whether the miner is currently mining
func (m *Miner) IsMining() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isMining
}

// shouldContinueMining checks if the miner should continue mining
func (m *Miner) shouldContinueMining() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isMining
}

// mineBlocks continuously mines new blocks
func (m *Miner) mineBlocks() {
	ticker := time.NewTicker(m.config.BlockTime)
	defer ticker.Stop()

	// Track if we're currently mining to prevent overlapping operations
	var isCurrentlyMining bool
	var miningMutex sync.Mutex

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.stopMining:
			return
		case <-ticker.C:
			// Check if we should still be mining
			if !m.shouldContinueMining() {
				return
			}

			// Check if we're already mining to prevent overlapping operations
			miningMutex.Lock()
			if isCurrentlyMining {
				miningMutex.Unlock()
				continue // Skip this tick if we're already mining
			}
			isCurrentlyMining = true
			miningMutex.Unlock()

			// Try to mine a new block
			if err := m.mineNextBlock(); err != nil {
				// Log error but continue mining
				fmt.Printf("Mining error: %v\n", err)
			}

			// Mark mining as complete
			miningMutex.Lock()
			isCurrentlyMining = false
			miningMutex.Unlock()
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

	// Check if the next block height already exists to prevent duplicate mining
	nextHeight := bestBlock.Header.Height + 1
	if existingBlock := m.chain.GetBlockByHeight(nextHeight); existingBlock != nil {
		return fmt.Errorf("block at height %d already exists", nextHeight)
	}

	// Create a new block
	newBlock := m.createNewBlock(bestBlock)
	if newBlock == nil {
		return fmt.Errorf("failed to create new block")
	}

	// Mine the block
	if err := m.mineBlock(newBlock); err != nil {
		if err.Error() == "mining stopped" {
			// This is expected when stopping mining, don't log as error
			return err
		}
		return fmt.Errorf("failed to mine block: %w", err)
	}

	// Add the block to the chain
	if err := m.chain.AddBlock(newBlock); err != nil {
		if err.Error() == "block already exists" {
			// This can happen due to race conditions, log but don't treat as critical error
			fmt.Printf("Block already exists (race condition): %v\n", err)
			return err
		}
		return fmt.Errorf("failed to add block to chain: %w", err)
	}

	// Call the callback if set
	if m.onBlockMined != nil {
		m.onBlockMined(newBlock)
	}

	fmt.Printf("Mined new block: %s\n", newBlock.String())

	return nil
}

// createNewBlock creates a new block for mining
func (m *Miner) createNewBlock(prevBlock *block.Block) *block.Block {
	// Get transactions from mempool
	transactions := m.mempool.GetTransactionsForBlock(m.config.MaxBlockSize)

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

	// Protect currentBlock access with mutex
	m.mu.Lock()
	m.currentBlock = newBlock
	m.mu.Unlock()

	// Create coinbase transaction
	coinbaseTx := m.createCoinbaseTransaction(prevBlock.Header.Height + 1)

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

	// Protect currentBlock access with mutex
	m.mu.RLock()
	for _, tx := range m.currentBlock.Transactions {
		if tx != nil {
			totalFees += tx.Fee
		}
	}
	m.mu.RUnlock()

	// Create coinbase output
	// Ensure we have a valid script public key (cannot be empty)
	scriptPubKey := m.config.CoinbaseAddress
	if scriptPubKey == "" {
		scriptPubKey = "coinbase" // Default fallback
	}

	// Ensure we have a valid value (cannot be zero)
	value := m.config.CoinbaseReward + totalFees
	if value == 0 {
		value = 1 // Minimum valid value
	}

	out := &block.TxOutput{
		Value:        value,
		ScriptPubKey: []byte(scriptPubKey),
	}

	// Create transaction
	tx := &block.Transaction{
		Version:  1,
		Inputs:   make([]*block.TxInput, 0), // Coinbase has no inputs
		Outputs:  []*block.TxOutput{out},
		LockTime: 0,
		Fee:      0,
	}

	// Calculate transaction hash
	tx.Hash = m.calculateTransactionHash(tx)

	return tx
}

// mineBlock performs proof-of-work mining on a block
func (m *Miner) mineBlock(block *block.Block) error {
	return m.consensus.MineBlock(block, m.stopMining)
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
	
	// Safely get chain information
	if bestBlock := m.chain.GetBestBlock(); bestBlock != nil {
		stats["difficulty"] = bestBlock.Header.Difficulty
		stats["height"] = bestBlock.Header.Height
		stats["bestBlockHash"] = fmt.Sprintf("%x", bestBlock.CalculateHash())
	} else {
		stats["difficulty"] = 0
		stats["height"] = 0
		stats["bestBlockHash"] = "none"
	}
	
	stats["config"] = map[string]interface{}{
		"miningEnabled":   m.config.MiningEnabled,
		"miningThreads":   m.config.MiningThreads,
		"blockTime":       m.config.BlockTime.String(),
		"maxBlockSize":    m.config.MaxBlockSize,
		"coinbaseReward":  m.config.CoinbaseReward,
	}

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
