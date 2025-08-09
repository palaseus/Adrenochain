package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/storage"
)

// Chain represents the blockchain
type Chain struct {
	mu            sync.RWMutex
	blocks        map[string]*block.Block // hash -> block (in-memory cache)
	blockByHeight map[uint64]*block.Block // height -> block (in-memory cache)
	bestBlock     *block.Block
	genesisBlock  *block.Block
	tipHash       []byte
	height        uint64
	config        *ChainConfig
	storage       *storage.Storage
}

// ChainConfig holds configuration for the blockchain
type ChainConfig struct {
	GenesisBlockReward uint64
	BlockTime          time.Duration
	DifficultyAdjustmentInterval uint64
	TargetBlockTime    time.Duration
	MaxBlockSize       uint64
}

// DefaultChainConfig returns the default chain configuration
func DefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		GenesisBlockReward: 1000000000, // 1 billion units
		BlockTime:          10 * time.Second,
		DifficultyAdjustmentInterval: 2016,
		TargetBlockTime:    10 * time.Second,
		MaxBlockSize:       1000000, // 1MB
	}
}

// NewChain creates a new blockchain
func NewChain(config *ChainConfig, s *storage.Storage) (*Chain, error) {
	chain := &Chain{
		blocks:        make(map[string]*block.Block),
		blockByHeight: make(map[uint64]*block.Block),
		config:        config,
		storage:       s,
	}

	// Load chain state from storage
	chainState, err := chain.storage.GetChainState()
	if err != nil {
		return nil, fmt.Errorf("failed to load chain state: %w", err)
	}

	if chainState.Height == 0 {
		// No chain state found, create genesis block
		chain.createGenesisBlock()
		// Store genesis block in storage
		if err := chain.storage.StoreBlock(chain.genesisBlock); err != nil {
			return nil, fmt.Errorf("failed to store genesis block: %w", err)
		}
		if err := chain.storage.StoreChainState(&storage.ChainState{
			BestBlockHash: chain.genesisBlock.CalculateHash(),
			Height:        chain.genesisBlock.Header.Height,
		}); err != nil {
			return nil, fmt.Errorf("failed to store chain state: %w", err)
		}
	} else {
		// Load best block from storage
		bestBlock, err := chain.storage.GetBlock(chainState.BestBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to load best block: %w", err)
		}
		chain.bestBlock = bestBlock
		chain.tipHash = chainState.BestBlockHash
		chain.height = chainState.Height
	}

	return chain, nil
}

// createGenesisBlock creates the genesis block
func (c *Chain) createGenesisBlock() {
	genesis := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32), // 32 bytes of zeros
			MerkleRoot:    make([]byte, 32), // Will be calculated
			Timestamp:     time.Unix(1231006505, 0), // Bitcoin genesis timestamp
			Difficulty:    1,
			Nonce:         0,
			Height:        0,
		},
		Transactions: make([]*block.Transaction, 0),
	}
	
	// Create coinbase transaction
	coinbaseTx := c.createCoinbaseTransaction(genesis.Header.Height, c.config.GenesisBlockReward)
	genesis.AddTransaction(coinbaseTx)
	
	// Calculate Merkle root
	genesis.Header.MerkleRoot = genesis.CalculateMerkleRoot()
	
	// Calculate hash
	hash := genesis.CalculateHash()
	
	// Store genesis block
	c.blocks[string(hash)] = genesis
	c.blockByHeight[0] = genesis
	c.genesisBlock = genesis
	c.bestBlock = genesis
	c.tipHash = hash
	c.height = 0
}

// createCoinbaseTransaction creates a coinbase transaction
func (c *Chain) createCoinbaseTransaction(height uint64, reward uint64) *block.Transaction {
	// Create a simple coinbase transaction
	output := &block.TxOutput{
		Value:        reward,
		ScriptPubKey: []byte(fmt.Sprintf("COINBASE_%d", height)),
	}
	
	tx := &block.Transaction{
		Version:  1,
		Inputs:   make([]*block.TxInput, 0), // Coinbase has no inputs
		Outputs:  []*block.TxOutput{output},
		LockTime: 0,
		Fee:      0,
	}
	
	// Calculate transaction hash
	tx.Hash = c.calculateTransactionHash(tx)
	
	return tx
}

// calculateTransactionHash calculates the hash of a transaction
func (c *Chain) calculateTransactionHash(tx *block.Transaction) []byte {
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

// AddBlock adds a new block to the chain
func (c *Chain) AddBlock(block *block.Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Validate the block
	if err := c.validateBlock(block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}
	
	// Check if block already exists
	hash := block.CalculateHash()
	if _, exists := c.blocks[string(hash)]; exists {
		return fmt.Errorf("block already exists")
	}
	
	// Add block to storage
	if err := c.storage.StoreBlock(block); err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}
	
	// Update chain tip if this block extends the current best chain
	if c.isBetterChain(block) {
		c.bestBlock = block
		c.tipHash = hash
		c.height = block.Header.Height
		
		// Store updated chain state
		if err := c.storage.StoreChainState(&storage.ChainState{
			BestBlockHash: c.tipHash,
			Height:        c.height,
		}); err != nil {
			return fmt.Errorf("failed to store chain state: %w", err)
		}
	}
	
	return nil
}

// validateBlock validates a block before adding it to the chain
func (c *Chain) validateBlock(block *block.Block) error {
	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}
	
	// Check block size
	if c.getBlockSize(block) > c.config.MaxBlockSize {
		return fmt.Errorf("block size %d exceeds maximum %d", 
			c.getBlockSize(block), c.config.MaxBlockSize)
	}
	
	// Check if previous block exists (except for genesis)
	if block.Header.Height > 0 {
		prevBlock, err := c.storage.GetBlock(block.Header.PrevBlockHash)
		if err != nil || prevBlock == nil {
			return fmt.Errorf("previous block not found")
		}
		
		// Check height continuity
		if prevBlock.Header.Height+1 != block.Header.Height {
			return fmt.Errorf("height discontinuity: expected %d, got %d", 
				prevBlock.Header.Height+1, block.Header.Height)
		}
		
		// Check timestamp
		if block.Header.Timestamp.Before(prevBlock.Header.Timestamp) {
			return fmt.Errorf("block timestamp %v is before previous block %v", 
				block.Header.Timestamp, prevBlock.Header.Timestamp)
		}
	}
	
	// Validate proof of work
	if !c.validateProofOfWork(block) {
		return fmt.Errorf("invalid proof of work")
	}
	
	return nil
}

// getBlockSize calculates the approximate size of a block
func (c *Chain) getBlockSize(block *block.Block) uint64 {
	size := uint64(0)
	
	// Header size (fixed)
	size += 80 // 32 + 32 + 8 + 8 + 8 + 4 = 92, rounded to 80 for simplicity
	
	// Transaction count
	size += 4
	
	// Transaction sizes
	for _, tx := range block.Transactions {
		size += c.getTransactionSize(tx)
	}
	
	return size
}

// getTransactionSize calculates the approximate size of a transaction
func (c *Chain) getTransactionSize(tx *block.Transaction) uint64 {
	size := uint64(0)
	
	// Version + LockTime + Fee
	size += 4 + 8 + 8
	
	// Input count + Output count
	size += 4 + 4
	
	// Inputs
	for _, input := range tx.Inputs {
		size += 32 + 4 + uint64(len(input.ScriptSig)) + 4
	}
	
	// Outputs
	for _, output := range tx.Outputs {
		size += 8 + uint64(len(output.ScriptPubKey))
	}
	
	return size
}

// validateProofOfWork validates the proof of work for a block
func (c *Chain) validateProofOfWork(block *block.Block) bool {
	hash := block.CalculateHash()
	target := c.calculateTarget(block.Header.Difficulty)
	
	// Check if hash is less than target
	return c.hashLessThan(hash, target)
}

// calculateTarget calculates the target hash for a given difficulty
func (c *Chain) calculateTarget(difficulty uint64) []byte {
	// For difficulty 0, any hash is valid (return all zeros)
	if difficulty == 0 {
		return make([]byte, 32)
	}
	
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
func (c *Chain) hashLessThan(hash1, hash2 []byte) bool {
	// If target is all zeros (difficulty 0), any hash is valid
	allZeros := true
	for _, b := range hash2 {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return true
	}
	
	// Normal comparison
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

// isBetterChain checks if the new block creates a better chain
func (c *Chain) isBetterChain(block *block.Block) bool {
	// For now, use longest chain rule
	// In a real implementation, this would consider accumulated difficulty
	return block.Header.Height > c.height
}

// GetBlock returns a block by its hash
func (c *Chain) GetBlock(hash []byte) *block.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Try to get from in-memory cache first
	if block, exists := c.blocks[string(hash)]; exists {
		return block
	}
	
	// Otherwise, load from storage
	block, err := c.storage.GetBlock(hash)
	if err != nil {
		return nil
	}
	
	// Add to in-memory cache
	c.blocks[string(hash)] = block
	
	return block
}

// GetBlockByHeight returns a block by its height
func (c *Chain) GetBlockByHeight(height uint64) *block.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Try to get from in-memory cache first
	if block, exists := c.blockByHeight[height]; exists {
		return block
	}
	
	// Otherwise, iterate through blocks to find by height (less efficient)
	// In a real implementation, storage would provide this directly
	for _, block := range c.blocks {
		if block.Header.Height == height {
			c.blockByHeight[height] = block // Cache it
			return block
		}
	}
	
	return nil
}

// GetBestBlock returns the current best block
func (c *Chain) GetBestBlock() *block.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.bestBlock
}

// GetHeight returns the current chain height
func (c *Chain) GetHeight() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.height
}

// GetTipHash returns the hash of the current tip
func (c *Chain) GetTipHash() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.tipHash
}

// GetGenesisBlock returns the genesis block
func (c *Chain) GetGenesisBlock() *block.Block {
	return c.genesisBlock
}

// CalculateNextDifficulty calculates the difficulty for the next block
func (c *Chain) CalculateNextDifficulty() uint64 {
	if c.height < c.config.DifficultyAdjustmentInterval {
		return c.bestBlock.Header.Difficulty
	}
	
	// Get the block from difficulty adjustment interval ago
	oldBlock := c.GetBlockByHeight(c.height - c.config.DifficultyAdjustmentInterval + 1)
	if oldBlock == nil {
		return c.bestBlock.Header.Difficulty
	}
	
	// Calculate time difference
	timeDiff := c.bestBlock.Header.Timestamp.Sub(oldBlock.Header.Timestamp)
	expectedTime := c.config.TargetBlockTime * time.Duration(c.config.DifficultyAdjustmentInterval)
	
	// Adjust difficulty
	newDifficulty := c.bestBlock.Header.Difficulty
	if timeDiff < expectedTime/4 {
		newDifficulty = newDifficulty * 4
	} else if timeDiff < expectedTime/2 {
		newDifficulty = newDifficulty * 2
	} else if timeDiff > expectedTime*4 {
		newDifficulty = newDifficulty / 4
	} else if timeDiff > expectedTime*2 {
		newDifficulty = newDifficulty / 2
	}
	
	// Ensure minimum difficulty
	if newDifficulty < 1 {
		newDifficulty = 1
	}
	
	return newDifficulty
}

// ForkChoice implements fork choice rules
func (c *Chain) ForkChoice(newBlock *block.Block) error {
	// For now, implement longest chain rule
	// In a real implementation, this would consider accumulated difficulty
	if newBlock.Header.Height > c.height {
		return c.AddBlock(newBlock)
	}
	
	return fmt.Errorf("block does not extend the best chain")
}

// Close closes the chain's storage
func (c *Chain) Close() error {
	return c.storage.Close()
}

// String returns a string representation of the chain
func (c *Chain) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return fmt.Sprintf("Chain{Height: %d, BestBlock: %s, TipHash: %x}", 
		c.height, c.bestBlock, c.tipHash)
} 