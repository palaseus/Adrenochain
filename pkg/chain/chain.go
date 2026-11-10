package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/config"
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/logger"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/palaseus/adrenochain/pkg/utxo"
)

// Chain represents the blockchain, managing blocks, chain state, and interactions with storage, UTXO set, and consensus.
type Chain struct {
	mutationMu    sync.Mutex
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
	logger        *logger.Logger           // logger provides structured logging.

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
	systemConfig := config.DefaultSystemConfig()
	return &ChainConfig{
		GenesisBlockReward: systemConfig.Blockchain.GenesisBlockReward,
		MaxBlockSize:       systemConfig.Blockchain.MaxBlockSize,
		MaxReorgDepth:      systemConfig.Blockchain.MaxReorgDepth,
	}
}

// NewChain creates a new blockchain instance.
// It initializes the chain from storage or creates a new genesis block if no chain state is found.
func NewChain(config *ChainConfig, consensusConfig *consensus.ConsensusConfig, s storage.StorageInterface) (*Chain, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if consensusConfig == nil {
		return nil, fmt.Errorf("consensusConfig cannot be nil")
	}
	if s == nil {
		return nil, fmt.Errorf("storage cannot be nil")
	}

	// Initialize logger
	chainLogger := logger.NewLogger(&logger.Config{
		Prefix: "chain",
		Level:  logger.INFO,
	})

	chain := &Chain{
		blocks:                make(map[string]*block.Block, 1000), // Pre-allocate with capacity
		blockByHeight:         make(map[uint64]*block.Block, 1000), // Pre-allocate with capacity
		config:                config,
		storage:               s,
		UTXOSet:               utxo.NewUTXOSet(),               // Initialize UTXOSet
		accumulatedDifficulty: make(map[uint64]*big.Int, 1000), // Pre-allocate with capacity
		reorgDepth:            config.MaxReorgDepth,
		logger:                chainLogger,
	}

	chain.consensus = consensus.NewConsensus(consensusConfig, chain)

	// Load chain state from storage
	chainState, err := chain.storage.GetChainState()
	if err != nil {
		return nil, fmt.Errorf("failed to load chain state: %w", err)
	}

	if chainState == nil {
		return nil, fmt.Errorf("storage returned nil chain state")
	}
	if len(chainState.BestBlockHash) == 0 {
		if chainState.Height != 0 {
			return nil, fmt.Errorf("chain state has height without a tip")
		}
		chain.createGenesisBlock()
		if err := chain.storage.StoreBlock(chain.genesisBlock); err != nil {
			return nil, fmt.Errorf("failed to store genesis block: %w", err)
		}
		if err := chain.storage.StoreChainState(&storage.ChainState{BestBlockHash: chain.tipHash, Height: 0}); err != nil {
			return nil, fmt.Errorf("failed to store chain state: %w", err)
		}
	} else {
		// Follow parent hashes from the persisted tip. Never reset persisted funds on a read error.
		hash := chainState.BestBlockHash
		for height := chainState.Height; ; height-- {
			current, err := chain.storage.GetBlock(hash)
			if err != nil {
				return nil, fmt.Errorf("failed to restore block at height %d: %w", height, err)
			}
			if current == nil || current.Header == nil || current.Header.Height != height || !constantTimeEqual(current.CalculateHash(), hash) {
				return nil, fmt.Errorf("invalid stored block at height %d", height)
			}
			chain.blocks[string(hash)] = current
			chain.blockByHeight[height] = current
			if height == chainState.Height {
				chain.bestBlock = current
			}
			if height == 0 {
				chain.genesisBlock = current
				break
			}
			hash = current.Header.PrevBlockHash
		}
		chain.height = chainState.Height
		chain.tipHash = append([]byte(nil), chainState.BestBlockHash...)
	}
	// Rebuild balances and work from the complete canonical history, including genesis.
	chain.accumulatedDifficulty[0] = big.NewInt(0)
	for height := uint64(0); ; height++ {
		current := chain.blockByHeight[height]
		if err := chain.UTXOSet.ProcessBlock(current); err != nil {
			return nil, fmt.Errorf("failed to restore UTXOs at height %d: %w", height, err)
		}
		chain.updateAccumulatedDifficulty(current)
		if height == chain.height {
			break
		}
	}
	if err := chain.validateChainState(); err != nil {
		return nil, err
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
	if tx == nil {
		return nil
	}

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

	// Serialize writers while allowing consensus to use the public chain readers.
	c.mutationMu.Lock()
	defer c.mutationMu.Unlock()
	hash := block.CalculateHash()
	c.mu.RLock()
	_, exists := c.blocks[string(hash)]
	c.mu.RUnlock()
	if exists {
		return fmt.Errorf("block already exists")
	}
	prevBlock := c.GetBlock(block.Header.PrevBlockHash)
	if err := c.consensus.ValidateBlock(block, prevBlock); err != nil {
		return fmt.Errorf("consensus validation failed: %w", err)
	}
	if !constantTimeEqual(block.Header.PrevBlockHash, c.GetTipHash()) {
		return fmt.Errorf("block does not extend the canonical tip; fork reorganization is unsupported")
	}
	if err := c.validateBlock(block); err != nil {
		return fmt.Errorf("chain validation failed: %w", err)
	}
	c.mu.Lock()
	if _, exists := c.blocks[string(hash)]; exists {
		c.mu.Unlock()
		return fmt.Errorf("block already exists")
	}
	// Commit the persisted tip before publishing the new in-memory state.
	if err := c.storage.StoreBlock(block); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to store block: %w", err)
	}
	if err := c.storage.StoreChainState(&storage.ChainState{BestBlockHash: hash, Height: block.Header.Height}); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to store chain state: %w", err)
	}
	if err := c.UTXOSet.ProcessBlock(block); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to process block: %w", err)
	}
	c.bestBlock = block
	c.tipHash = hash
	c.height = block.Header.Height
	c.blocks[string(hash)] = block
	c.blockByHeight[block.Header.Height] = block
	c.updateAccumulatedDifficulty(block)
	c.mu.Unlock()
	if prevBlock != nil {
		c.consensus.UpdateDifficulty(block.Header.Timestamp.Sub(prevBlock.Header.Timestamp))
	}

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

	if err := c.UTXOSet.ValidateBlockTransactions(block); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	return nil
}

// GetBlockSize calculates the approximate size of a block
// GetBlockSize calculates the approximate size of a block in bytes.
func (c *Chain) GetBlockSize(b *block.Block) uint64 {
	if b == nil || b.Header == nil {
		return 0
	}
	// Header fields, header length, and transaction count.
	size := uint64(44 + len(b.Header.PrevBlockHash) + len(b.Header.MerkleRoot))
	for _, tx := range b.Transactions {
		size += 4 + c.getTransactionSize(tx)
	}
	return size
}

// getTransactionSize matches the length-prefixed binary transaction encoding.
func (c *Chain) getTransactionSize(tx *block.Transaction) uint64 {
	if tx == nil {
		return 0
	}
	size := uint64(28 + len(tx.Hash))
	for _, input := range tx.Inputs {
		if input != nil {
			size += uint64(16 + len(input.PrevTxHash) + len(input.ScriptSig))
		}
	}
	for _, output := range tx.Outputs {
		if output != nil {
			size += uint64(16 + len(output.ScriptPubKey))
		}
	}
	return size
}

// isBetterChain checks if the new block creates a better chain
// isBetterChain checks if the new block creates a better chain than the current best chain.
// Competing branches are not adopted without an explicit reorganization implementation.
func (c *Chain) isBetterChain(block *block.Block) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if block == nil || block.Header == nil {
		return false
	}

	// For simple cases, just check if this block extends the current best chain
	if c.bestBlock != nil {
		// Check if this block extends the current best chain
		if constantTimeEqual(block.Header.PrevBlockHash, c.bestBlock.CalculateHash()) {
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

	currentChainDiff, err := c.calculateAccumulatedDifficulty(c.height)
	if err != nil {
		return false // Can't calculate, assume not better
	}

	// Compare accumulated difficulties
	return newChainDiff.Cmp(currentChainDiff) > 0
}

// GetBlock returns a block by its hash.
// It first checks the in-memory cache, then loads from storage if not found.
func (c *Chain) GetBlock(hash []byte) *block.Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getBlockLocked(hash)
}

func (c *Chain) getBlockLocked(hash []byte) *block.Block {
	if hash == nil {
		return nil
	}

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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getBlockByHeightLocked(height)
}

func (c *Chain) getBlockByHeightLocked(height uint64) *block.Block {
	// Try to get from in-memory cache first
	if block, exists := c.blockByHeight[height]; exists {
		return block
	}

	if height > c.height {
		return nil
	}
	current := c.bestBlock
	for current != nil && current.Header.Height > height {
		current = c.getBlockLocked(current.Header.PrevBlockHash)
	}
	if current != nil && current.Header.Height == height {
		c.blockByHeight[height] = current
		return current
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// GetTipHash returns the hash of the current best block (tip) of the chain.
func (c *Chain) GetTipHash() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return append([]byte(nil), c.tipHash...)
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if diff, exists := c.accumulatedDifficulty[height]; exists {
		return new(big.Int).Set(diff), nil
	}

	// Calculate if not cached
	return c.calculateAccumulatedDifficulty(height)
}

// calculateAccumulatedDifficulty calculates the accumulated difficulty up to the given height.
func (c *Chain) calculateAccumulatedDifficulty(height uint64) (*big.Int, error) {
	accumulated := big.NewInt(0)

	for h := uint64(1); h <= height; h++ {
		block := c.getBlockByHeightLocked(h)
		if block == nil {
			return nil, fmt.Errorf("block not found at height %d", h)
		}

		blockDiff := new(big.Int).SetUint64(block.Header.Difficulty)
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
			// Check if we actually have a block at this height before calculating difficulty
			block := c.getBlockByHeightLocked(h)
			if block == nil {
				// Skip this height if no block exists
				// This can happen if the chain state is inconsistent
				continue
			}

			// Calculate accumulated difficulty up to this height
			accumulated := big.NewInt(0)
			for height := uint64(1); height <= h; height++ {
				blockAtHeight := c.getBlockByHeightLocked(height)
				if blockAtHeight == nil {
					// Skip this height if no block exists
					continue
				}
				blockDiff := big.NewInt(int64(blockAtHeight.Header.Difficulty))
				accumulated.Add(accumulated, blockDiff)
			}
			c.accumulatedDifficulty[h] = accumulated
		}
	}

	return nil
}

// loadBlocksFromStorage loads all blocks from storage into memory
func (c *Chain) loadBlocksFromStorage() error {
	// Load genesis block first (height 0)
	if c.genesisBlock != nil {
		c.blocks[string(c.genesisBlock.CalculateHash())] = c.genesisBlock
		c.blockByHeight[0] = c.genesisBlock
	}

	// For now, we'll just ensure the chain state is correct
	// The actual block loading can be done lazily when needed
	// This prevents the "block not found" errors during initialization

	// If we have a best block, make sure it's in our caches
	if c.bestBlock != nil {
		bestHash := c.bestBlock.CalculateHash()
		c.blocks[string(bestHash)] = c.bestBlock
		c.blockByHeight[c.bestBlock.Header.Height] = c.bestBlock
	}

	return nil
}

// updateAccumulatedDifficulty updates the accumulated difficulty cache when a new block is added.
// This method assumes the caller already holds the lock.
func (c *Chain) updateAccumulatedDifficulty(block *block.Block) {
	if block == nil || block.Header == nil {
		return
	}

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
	blockDiff := new(big.Int).SetUint64(block.Header.Difficulty)
	newDiff := new(big.Int).Add(prevDiff, blockDiff)
	c.accumulatedDifficulty[height] = newDiff
}

// ForkChoice implements the fork choice rules to determine the canonical chain.
// It uses accumulated difficulty to choose the best chain.
func (c *Chain) ForkChoice(newBlock *block.Block) error {
	if newBlock == nil {
		return fmt.Errorf("cannot perform fork choice on nil block")
	}

	return c.AddBlock(newBlock)
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

// validateChainState performs internal consistency checks on the chain's state.
func (c *Chain) validateChainState() error {
	// Check if genesis block exists and is correct
	if c.genesisBlock == nil {
		return fmt.Errorf("genesis block not found")
	}
	if c.genesisBlock.Header == nil {
		return fmt.Errorf("genesis block header is nil")
	}
	if c.genesisBlock.Header.Height != 0 {
		return fmt.Errorf("genesis block height is not 0")
	}

	// Check if best block is set and consistent
	if c.bestBlock == nil {
		return fmt.Errorf("best block not set")
	}
	if c.bestBlock.Header == nil {
		return fmt.Errorf("best block header is nil")
	}
	if c.bestBlock.Header.Height != c.height {
		return fmt.Errorf("best block height (%d) does not match chain height (%d)",
			c.bestBlock.Header.Height, c.height)
	}
	if !constantTimeEqual(c.bestBlock.CalculateHash(), c.tipHash) {
		return fmt.Errorf("best block hash (%x) does not match tip hash (%x)",
			c.bestBlock.CalculateHash(), c.tipHash)
	}

	// Check if UTXO set is initialized
	if c.UTXOSet == nil {
		return fmt.Errorf("UTXO set is not initialized")
	}

	// Check if accumulated difficulty is initialized and correct for genesis
	if c.accumulatedDifficulty == nil {
		return fmt.Errorf("accumulated difficulty cache is not initialized")
	}
	if c.accumulatedDifficulty[0] == nil {
		return fmt.Errorf("accumulated difficulty for height 0 is not initialized")
	}
	if c.accumulatedDifficulty[0].Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf("accumulated difficulty for height 0 is not 0")
	}

	// Check if consensus is initialized
	if c.consensus == nil {
		return fmt.Errorf("consensus module is not initialized")
	}

	// Check if storage is initialized
	if c.storage == nil {
		return fmt.Errorf("storage is not initialized")
	}

	return nil
}

// CleanupMemory performs memory cleanup to prevent memory leaks
func (c *Chain) CleanupMemory() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean up old blocks beyond reorg depth
	if len(c.blocks) > int(c.reorgDepth*2) {
		// Keep only the most recent blocks
		height := c.height
		if height <= c.reorgDepth {
			return
		}
		cutoffHeight := height - c.reorgDepth

		for hash, block := range c.blocks {
			if block.Header.Height < cutoffHeight {
				delete(c.blocks, hash)
				delete(c.blockByHeight, block.Header.Height)
				delete(c.accumulatedDifficulty, block.Header.Height)
			}
		}
	}

	// Force garbage collection
	runtime.GC()
}

// constantTimeEqual performs constant-time comparison to prevent timing attacks
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
