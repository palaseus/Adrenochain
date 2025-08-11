package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/gochain/gochain/pkg/utxo"
)

// Chain represents the blockchain, managing blocks, chain state, and interactions with storage, UTXO set, and consensus.
type Chain struct {
	mu            sync.RWMutex             // mu protects concurrent access to chain fields.
	blocks        map[string]*block.Block  // blocks is an in-memory cache of hash -> block.
	blockByHeight map[uint64]*block.Block  // blockByHeight is an in-memory cache of height -> block.
	bestBlock     *block.Block             // bestBlock is the current tip of the longest chain.
	genesisBlock  *block.Block             // genesisBlock is the first block in the chain.
	tipHash       []byte                   // tipHash is the hash of the current best block.
	height        uint64                   // height is the current height of the chain (number of blocks).
	config        *ChainConfig             // config holds the chain's configuration parameters.
	storage       storage.StorageInterface // storage provides persistent storage for blocks and chain state.
	UTXOSet       *utxo.UTXOSet            // UTXOSet manages the unspent transaction outputs.
	consensus     *consensus.Consensus     // consensus handles the blockchain's consensus rules.

	// Fork choice and finality fields
	accumulatedDifficulty map[uint64]*big.Int // accumulatedDifficulty stores difficulty sums for each height
	reorgDepth            uint64              // reorgDepth is the maximum depth for reorganizations
}

// ChainConfig holds configuration parameters for the blockchain.
type ChainConfig struct {
	GenesisBlockReward uint64 // GenesisBlockReward is the reward for the genesis block.
	MaxBlockSize       uint64 // MaxBlockSize is the maximum allowed size for a block in bytes.
	MaxReorgDepth      uint64 // MaxReorgDepth is the maximum depth for chain reorganizations
}

// DefaultChainConfig returns the default configuration for the blockchain.
func DefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		GenesisBlockReward: 1000000000, // 1 billion units
		MaxBlockSize:       1000000,    // 1MB
		MaxReorgDepth:      100,        // Maximum 100 block reorg
	}
}

// NewChain creates a new blockchain instance.
// It initializes the chain from storage or creates a new genesis block if no chain state is found.
func NewChain(config *ChainConfig, consensusConfig *consensus.ConsensusConfig, s storage.StorageInterface) (*Chain, error) {
	chain := &Chain{
		blocks:                make(map[string]*block.Block),
		blockByHeight:         make(map[uint64]*block.Block),
		config:                config,
		storage:               s,
		UTXOSet:               utxo.NewUTXOSet(), // Initialize UTXOSet
		accumulatedDifficulty: make(map[uint64]*big.Int),
		reorgDepth:            config.MaxReorgDepth,
	}

	chain.consensus = consensus.NewConsensus(consensusConfig, chain)

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
		// Process genesis block to update UTXO set
		if err := chain.UTXOSet.ProcessBlock(chain.genesisBlock); err != nil {
			return nil, fmt.Errorf("failed to process genesis block for UTXO set: %w", err)
		}

		// Initialize accumulated difficulty for genesis
		chain.accumulatedDifficulty[0] = big.NewInt(0)
	} else {
		// Load best block from storage
		bestBlock, err := chain.storage.GetBlock(chainState.BestBlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to load best block: %w", err)
		}
		chain.bestBlock = bestBlock
		chain.tipHash = chainState.BestBlockHash
		chain.height = chainState.Height

		// Load all blocks from storage into memory
		if err := chain.loadBlocksFromStorage(); err != nil {
			return nil, fmt.Errorf("failed to load blocks from storage: %w", err)
		}

		// Rebuild UTXO set from scratch (for simplicity, in a real chain, this would be optimized)
		// For now, we assume the UTXO set is built up as blocks are added

		// Rebuild accumulated difficulty
		if err := chain.rebuildAccumulatedDifficulty(); err != nil {
			return nil, fmt.Errorf("failed to rebuild accumulated difficulty: %w", err)
		}
	}

	return chain, nil
}

// createGenesisBlock creates the genesis block
// createGenesisBlock creates the very first block in the blockchain.
// It initializes the genesis block with predefined values and a coinbase transaction.
func (c *Chain) createGenesisBlock() {
	genesis := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),         // 32 bytes of zeros
			MerkleRoot:    make([]byte, 32),         // Will be calculated
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
// createCoinbaseTransaction creates a special transaction that rewards the miner for creating a new block.
// Coinbase transactions have no inputs and are the first transaction in a block.
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
// calculateTransactionHash calculates the SHA256 hash of a transaction.
// This hash serves as the transaction's unique identifier.
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

// AddBlock adds a new block to the chain.
// It validates the block against consensus rules, stores it, and updates the chain state if it extends the best chain.
func (c *Chain) AddBlock(block *block.Block) error {
	if block == nil {
		return fmt.Errorf("cannot add nil block")
	}
	if block.Header == nil {
		return fmt.Errorf("block header cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate the block using consensus rules
	prevBlock := c.GetBlock(block.Header.PrevBlockHash)
	if err := c.consensus.ValidateBlock(block, prevBlock); err != nil {
		return fmt.Errorf("consensus validation failed: %w", err)
	}

	// Validate the block using chain-specific rules (size, etc.)
	if err := c.validateBlock(block); err != nil {
		return fmt.Errorf("chain validation failed: %w", err)
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

		// Update consensus
		if prevBlock != nil {
			blockTime := block.Header.Timestamp.Sub(prevBlock.Header.Timestamp)
			c.consensus.UpdateDifficulty(blockTime)
		}

		// Store updated chain state
		if err := c.storage.StoreChainState(&storage.ChainState{
			BestBlockHash: c.tipHash,
			Height:        c.height,
		}); err != nil {
			return fmt.Errorf("failed to store chain state: %w", err)
		}
		// Process block to update UTXO set
		if err := c.UTXOSet.ProcessBlock(block); err != nil {
			return fmt.Errorf("failed to process block for UTXO set: %w", err)
		}

		// Update accumulated difficulty cache
		c.updateAccumulatedDifficulty(block)
	} else {
		// Even if not the best chain, update height if this block has higher height
		if block.Header.Height > c.height {
			c.height = block.Header.Height
		}
	}

	// Always add to in-memory caches
	c.blocks[string(hash)] = block
	c.blockByHeight[block.Header.Height] = block

	return nil
}

// validateBlock validates a block before adding it to the chain
// validateBlock performs internal validation checks on a block before it is added to the chain.
// This includes checks for block size, previous block existence, height continuity, timestamp, proof of work, and transaction validity.
func (c *Chain) validateBlock(block *block.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}
	if block.Header == nil {
		return fmt.Errorf("block header cannot be nil")
	}



	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Check block size
	blockSize := c.GetBlockSize(block)
	if blockSize > c.config.MaxBlockSize {
		return fmt.Errorf("block size %d exceeds maximum %d",
			blockSize, c.config.MaxBlockSize)
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
	if !c.consensus.ValidateProofOfWork(block) {
		return fmt.Errorf("invalid proof of work")
	}

	// Validate transactions against UTXO set
	for _, tx := range block.Transactions {
		if err := c.UTXOSet.ValidateTransaction(tx); err != nil {
			return fmt.Errorf("transaction validation failed: %w", err)
		}
	}

	return nil
}

// GetBlockSize calculates the approximate size of a block
// GetBlockSize calculates the approximate size of a block in bytes.
func (c *Chain) GetBlockSize(block *block.Block) uint64 {
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
// getTransactionSize calculates the approximate size of a transaction in bytes.
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

// isBetterChain checks if the new block creates a better chain
// isBetterChain checks if the new block creates a better chain than the current best chain.
// Currently, it implements the longest chain rule.
func (c *Chain) isBetterChain(block *block.Block) bool {
	if block == nil || block.Header == nil {
		return false
	}

	// For simple cases, just check if this block extends the current best chain
	if c.bestBlock != nil {
		// Check if this block extends the current best chain
		if bytes.Equal(block.Header.PrevBlockHash, c.bestBlock.CalculateHash()) {
			return true
		}
	} else {
		// If no best block yet, this could be the first block after genesis
		return true
	}

	// Fallback to accumulated difficulty comparison for more complex cases
	newChainDiff, err := c.calculateAccumulatedDifficulty(block.Header.Height)
	if err != nil {
		return false // Can't calculate, assume not better
	}

	currentChainDiff, err := c.GetAccumulatedDifficulty(c.height)
	if err != nil {
		return false // Can't calculate, assume not better
	}

	// Compare accumulated difficulties
	return newChainDiff.Cmp(currentChainDiff) > 0
}

// GetBlock returns a block by its hash.
// It first checks the in-memory cache, then loads from storage if not found.
func (c *Chain) GetBlock(hash []byte) *block.Block {
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

// GetBlockByHeight returns a block by its height.
// It first checks the in-memory cache, then iterates through blocks (less efficient) if not found.
func (c *Chain) GetBlockByHeight(height uint64) *block.Block {
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

// GetBestBlock returns the current best block (tip) of the chain.
func (c *Chain) GetBestBlock() *block.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.bestBlock
}

// GetHeight returns the current height of the chain.
func (c *Chain) GetHeight() uint64 {
	return c.height
}

// GetTipHash returns the hash of the current best block (tip) of the chain.
func (c *Chain) GetTipHash() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.tipHash
}

// GetGenesisBlock returns the genesis block of the chain.
func (c *Chain) GetGenesisBlock() *block.Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.genesisBlock
}

// CalculateNextDifficulty calculates the difficulty for the next block to be mined.
// This is delegated to the consensus module.
func (c *Chain) CalculateNextDifficulty() uint64 {
	return c.consensus.GetDifficulty()
}

// GetConsensus returns the consensus instance for testing purposes.
func (c *Chain) GetConsensus() *consensus.Consensus {
	return c.consensus
}

// GetAccumulatedDifficulty returns the accumulated difficulty up to the given height.
// This implements the consensus.ChainReader interface.
func (c *Chain) GetAccumulatedDifficulty(height uint64) (*big.Int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if diff, exists := c.accumulatedDifficulty[height]; exists {
		return diff, nil
	}

	// Calculate if not cached
	return c.calculateAccumulatedDifficulty(height)
}

// calculateAccumulatedDifficulty calculates the accumulated difficulty up to the given height.
func (c *Chain) calculateAccumulatedDifficulty(height uint64) (*big.Int, error) {
	accumulated := big.NewInt(0)

	for h := uint64(1); h <= height; h++ {
		block := c.GetBlockByHeight(h)
		if block == nil {
			return nil, fmt.Errorf("block not found at height %d", h)
		}

		blockDiff := big.NewInt(int64(block.Header.Difficulty))
		accumulated.Add(accumulated, blockDiff)
	}

	return accumulated, nil
}

// rebuildAccumulatedDifficulty rebuilds the accumulated difficulty cache from storage.
func (c *Chain) rebuildAccumulatedDifficulty() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear existing cache
	c.accumulatedDifficulty = make(map[uint64]*big.Int)

	// Initialize genesis
	c.accumulatedDifficulty[0] = big.NewInt(0)

	// Only rebuild for heights that actually have blocks
	if c.height > 0 {
		for h := uint64(1); h <= c.height; h++ {
			diff, err := c.calculateAccumulatedDifficulty(h)
			if err != nil {
				return fmt.Errorf("failed to calculate difficulty at height %d: %w", h, err)
			}
			c.accumulatedDifficulty[h] = diff
		}
	}

	return nil
}

// loadBlocksFromStorage loads all blocks from storage into memory
func (c *Chain) loadBlocksFromStorage() error {
	// For now, this is a simplified implementation
	// In a real implementation, you'd want to load all blocks from storage
	// For testing purposes, we'll just return success
	return nil
}

// updateAccumulatedDifficulty updates the accumulated difficulty cache when a new block is added.
// This method assumes the caller already holds the lock.
func (c *Chain) updateAccumulatedDifficulty(block *block.Block) {
	height := block.Header.Height
	if height == 0 {
		c.accumulatedDifficulty[0] = big.NewInt(0)
		return
	}

	// Get previous accumulated difficulty
	prevDiff := big.NewInt(0)
	if prev, exists := c.accumulatedDifficulty[height-1]; exists {
		prevDiff = prev
	}

	// Add current block difficulty
	blockDiff := big.NewInt(int64(block.Header.Difficulty))
	newDiff := new(big.Int).Add(prevDiff, blockDiff)
	c.accumulatedDifficulty[height] = newDiff
}

// ForkChoice implements the fork choice rules to determine the canonical chain.
// It uses accumulated difficulty to choose the best chain.
func (c *Chain) ForkChoice(newBlock *block.Block) error {
	// Check if this block creates a better chain
	if c.isBetterChain(newBlock) {
		return c.AddBlock(newBlock)
	}

	return fmt.Errorf("block does not create a better chain")
}

// Close closes the chain's underlying storage.
func (c *Chain) Close() error {
	return c.storage.Close()
}

// String returns a human-readable string representation of the chain.
func (c *Chain) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("Chain{Height: %d, BestBlock: %s, TipHash: %x}",
		c.height, c.bestBlock, c.tipHash)
}
