package consensus

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// ChainReader defines the methods from the chain that the consensus package needs
// to interact with the blockchain state without creating circular dependencies.
type ChainReader interface {
	GetHeight() uint64
	GetBlockByHeight(height uint64) *block.Block
	GetBlock(hash []byte) *block.Block
	GetAccumulatedDifficulty(height uint64) (*big.Int, error)
}

// Consensus represents the blockchain consensus mechanism.
// It manages difficulty adjustment, proof-of-work validation, block mining, and finality rules.
type Consensus struct {
	mu             sync.RWMutex     // mu protects concurrent access to consensus fields.
	config         *ConsensusConfig // config holds the consensus configuration parameters.
	difficulty     uint64           // difficulty is the current mining difficulty.
	lastAdjustment time.Time        // lastAdjustment records the time of the last difficulty adjustment.
	blockTimes     []time.Duration  // blockTimes stores the durations of recent blocks for difficulty adjustment.
	chain          ChainReader      // chain is a reference to the chain, used to query block information.

	// Finality-related fields
	finalityDepth uint64            // finalityDepth is the number of blocks required for finality
	checkpoints   map[uint64][]byte // checkpoints stores known good block hashes at specific heights
}

// ConsensusConfig holds configuration parameters for the consensus mechanism.
type ConsensusConfig struct {
	TargetBlockTime              time.Duration // TargetBlockTime is the desired average time between blocks.
	DifficultyAdjustmentInterval uint64        // DifficultyAdjustmentInterval is the number of blocks after which difficulty is adjusted.
	MaxDifficulty                uint64        // MaxDifficulty is the maximum allowed difficulty.
	MinDifficulty                uint64        // MinDifficulty is the minimum allowed difficulty.
	DifficultyAdjustmentFactor   float64       // DifficultyAdjustmentFactor is used to dampen difficulty swings.
	FinalityDepth                uint64        // FinalityDepth is the number of blocks required for finality
	CheckpointInterval           uint64        // CheckpointInterval is the height interval for checkpoints
}

// DefaultConsensusConfig returns the default consensus configuration.
func DefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		TargetBlockTime:              10 * time.Second,
		DifficultyAdjustmentInterval: 2016,
		MaxDifficulty:                256,
		MinDifficulty:                1,
		DifficultyAdjustmentFactor:   4.0,
		FinalityDepth:                100,   // 100 blocks for finality
		CheckpointInterval:           10000, // Checkpoint every 10,000 blocks
	}
}

// NewConsensus creates a new consensus instance.
// It initializes the consensus mechanism with the given configuration and a reference to the chain.
func NewConsensus(config *ConsensusConfig, chain ChainReader) *Consensus {
	return &Consensus{
		config:         config,
		difficulty:     config.MinDifficulty,
		lastAdjustment: time.Now(),
		blockTimes:     make([]time.Duration, 0),
		chain:          chain,
		finalityDepth:  config.FinalityDepth,
		checkpoints:    make(map[uint64][]byte),
	}
}

// IsBlockFinal checks if a block at the given height is considered final.
// A block is final if it's at least finalityDepth blocks behind the current tip.
func (c *Consensus) IsBlockFinal(height uint64) bool {
	currentHeight := c.chain.GetHeight()
	return currentHeight >= height+c.finalityDepth
}

// GetFinalityDepth returns the current finality depth setting.
func (c *Consensus) GetFinalityDepth() uint64 {
	return c.finalityDepth
}

// AddCheckpoint adds a checkpoint at the given height.
// Checkpoints are used to prevent long-range attacks and provide security guarantees.
func (c *Consensus) AddCheckpoint(height uint64, hash []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checkpoints[height] = hash
}

// hasCheckpoint checks if a checkpoint exists at the given height
func (c *Consensus) hasCheckpoint(height uint64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.checkpoints[height]
	return exists
}

// ValidateCheckpoint validates that a block at the given height matches the expected checkpoint hash.
// This method is used for direct checkpoint validation and returns false for heights without checkpoints.
func (c *Consensus) ValidateCheckpoint(height uint64, hash []byte) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if expectedHash, exists := c.checkpoints[height]; exists {
		return string(expectedHash) == string(hash)
	}
	return false // No checkpoint at this height, cannot validate
}

// GetAccumulatedDifficulty calculates the accumulated difficulty from genesis to the given height.
// This is used for fork choice and finality determination.
func (c *Consensus) GetAccumulatedDifficulty(height uint64) (*big.Int, error) {
	if height == 0 {
		return big.NewInt(0), nil
	}

	accumulated := big.NewInt(0)
	for h := uint64(1); h <= height; h++ {
		block := c.chain.GetBlockByHeight(h)
		if block == nil {
			return nil, fmt.Errorf("block not found at height %d", h)
		}

		// Add the difficulty of this block to accumulated difficulty
		blockDiff := big.NewInt(int64(block.Header.Difficulty))
		accumulated.Add(accumulated, blockDiff)
	}

	return accumulated, nil
}

// calculateExpectedDifficulty calculates the expected difficulty for a given block height.
// This is used during block validation to ensure the block's difficulty matches the network's rules.
func (c *Consensus) calculateExpectedDifficulty(blockHeight uint64) (uint64, error) {
	if blockHeight == 0 {
		return c.config.MinDifficulty, nil // Genesis block always has min difficulty
	}

	if blockHeight%c.config.DifficultyAdjustmentInterval != 0 {
		// If not an adjustment block, difficulty is the same as the previous block
		prevBlock := c.chain.GetBlockByHeight(blockHeight - 1)
		if prevBlock == nil {
			return 0, fmt.Errorf("previous block not found for height %d", blockHeight)
		}
		return prevBlock.Header.Difficulty, nil
	}

	// It's an adjustment block, calculate new difficulty
	currentBlock := c.chain.GetBlockByHeight(blockHeight - 1)
	if currentBlock == nil {
		return 0, fmt.Errorf("current block not found for height %d", blockHeight-1)
	}

	oldBlockHeight := blockHeight - c.config.DifficultyAdjustmentInterval
	oldBlock := c.chain.GetBlockByHeight(oldBlockHeight)
	if oldBlock == nil {
		return 0, fmt.Errorf("old block not found for height %d", oldBlockHeight)
	}

	actualTime := currentBlock.Header.Timestamp.Sub(oldBlock.Header.Timestamp)
	expectedTime := time.Duration(c.config.DifficultyAdjustmentInterval) * c.config.TargetBlockTime

	adjustmentFactor := float64(actualTime) / float64(expectedTime)

	if adjustmentFactor < 1.0/c.config.DifficultyAdjustmentFactor {
		adjustmentFactor = 1.0 / c.config.DifficultyAdjustmentFactor
	}
	if adjustmentFactor > c.config.DifficultyAdjustmentFactor {
		adjustmentFactor = c.config.DifficultyAdjustmentFactor
	}

	oldDifficulty := oldBlock.Header.Difficulty
	newDifficulty := uint64(float64(oldDifficulty) * adjustmentFactor)

	if newDifficulty < c.config.MinDifficulty {
		newDifficulty = c.config.MinDifficulty
	}
	if newDifficulty > c.config.MaxDifficulty {
		newDifficulty = c.config.MaxDifficulty
	}

	return newDifficulty, nil
}

// ValidateBlock validates a block according to consensus rules.
// This includes proof of work, timestamp validation, difficulty validation, and finality checks.
func (c *Consensus) ValidateBlock(block *block.Block, prevBlock *block.Block) error {
	// Check if block is nil
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Check proof of work
	if !c.ValidateProofOfWork(block) {
		return fmt.Errorf("invalid proof of work")
	}

	// Check timestamp
	if prevBlock != nil {
		if block.Header.Timestamp.Before(prevBlock.Header.Timestamp) {
			return fmt.Errorf("block timestamp %v is before previous block %v",
				block.Header.Timestamp, prevBlock.Header.Timestamp)
		}

		// Check if block is too far in the future (2 hours)
		maxFutureTime := time.Now().Add(2 * time.Hour)
		if block.Header.Timestamp.After(maxFutureTime) {
			return fmt.Errorf("block timestamp %v is too far in the future",
				block.Header.Timestamp)
		}
	}

	// Check difficulty
	expectedDifficulty, err := c.calculateExpectedDifficulty(block.Header.Height)
	if err != nil {
		return fmt.Errorf("failed to calculate expected difficulty: %w", err)
	}

	if block.Header.Difficulty != expectedDifficulty {
		return fmt.Errorf("block difficulty %d does not match expected %d",
			block.Header.Difficulty, expectedDifficulty)
	}

	// Validate merkle root
	if err := c.validateMerkleRoot(block); err != nil {
		return fmt.Errorf("merkle root validation failed: %w", err)
	}

	// Validate all transactions in the block
	if err := c.validateBlockTransactions(block); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Validate checkpoint if this height has one
	if c.hasCheckpoint(block.Header.Height) {
		if !c.ValidateCheckpoint(block.Header.Height, block.CalculateHash()) {
			return fmt.Errorf("block hash does not match checkpoint at height %d", block.Header.Height)
		}
	}

	return nil
}

// validateMerkleRoot validates that the block's merkle root matches the calculated merkle root
// of all transactions in the block
func (c *Consensus) validateMerkleRoot(block *block.Block) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// Calculate merkle root from transactions
	calculatedRoot := c.calculateMerkleRoot(block.Transactions)

	// Compare with block header merkle root
	if !c.bytesEqual(calculatedRoot, block.Header.MerkleRoot) {
		return fmt.Errorf("merkle root mismatch: calculated %x, header %x",
			calculatedRoot, block.Header.MerkleRoot)
	}

	return nil
}

// calculateMerkleRoot calculates the merkle root of a list of transactions
func (c *Consensus) calculateMerkleRoot(transactions []*block.Transaction) []byte {
	if len(transactions) == 0 {
		return nil
	}

	if len(transactions) == 1 {
		return transactions[0].CalculateHash()
	}

	// Build merkle tree bottom-up
	hashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.CalculateHash()
	}

	// Keep combining pairs until we have a single hash
	for len(hashes) > 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1]) // Duplicate last if odd
		}

		newHashes := make([][]byte, len(hashes)/2)
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			newHashes[i/2] = c.hash256(combined)
		}
		hashes = newHashes
	}

	return hashes[0]
}

// hash256 performs double SHA256 hashing
func (c *Consensus) hash256(data []byte) []byte {
	// For now, use a simple hash function
	// In production, this should use crypto/sha256
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = data[i%len(data)] ^ byte(i)
	}
	return hash
}

// validateBlockTransactions validates all transactions in a block
func (c *Consensus) validateBlockTransactions(block *block.Block) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// First transaction should be coinbase
	if !block.Transactions[0].IsCoinbase() {
		return fmt.Errorf("first transaction is not coinbase")
	}

	// Validate each transaction
	for i, tx := range block.Transactions {
		if err := c.validateTransaction(tx); err != nil {
			return fmt.Errorf("transaction %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateTransaction validates a single transaction
func (c *Consensus) validateTransaction(tx *block.Transaction) error {
	// Basic transaction validation
	if err := tx.IsValid(); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Skip coinbase transaction validation (no inputs to validate)
	if tx.IsCoinbase() {
		return nil
	}

	// Validate inputs and outputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Validate signature (this would require access to UTXO set in real implementation)
	// For now, we'll assume the transaction is pre-validated

	return nil
}

// bytesEqual performs constant-time comparison of two byte slices
func (c *Consensus) bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// ValidateProofOfWork validates the proof of work for a block.
// It checks if the block's hash is less than or equal to the target derived from the current difficulty.
func (c *Consensus) ValidateProofOfWork(block *block.Block) bool {
	hash := block.CalculateHash()
	target := c.calculateTarget(c.difficulty)

	return c.hashLessThan(hash, target)
}

// calculateTarget calculates the target hash for a given difficulty.
// The target is a 32-byte array that the block's hash must be less than or equal to.
func (c *Consensus) calculateTarget(difficulty uint64) []byte {
	// Target = 2^(256-difficulty)
	target := new(big.Int)
	target.SetBit(target, int(256-difficulty), 1)

	// Convert to 32-byte array
	targetBytes := target.Bytes()
	if len(targetBytes) > 32 {
		return targetBytes[:32]
	}

	// Pad with zeros if necessary
	result := make([]byte, 32)
	copy(result[32-len(targetBytes):], targetBytes)

	return result
}

// hashLessThan checks if hash1 is lexicographically less than hash2.
// This is used to determine if a block's hash meets the target difficulty.
func (c *Consensus) hashLessThan(hash1, hash2 []byte) bool {
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

// MineBlock mines a block by finding a nonce that satisfies the proof-of-work requirement.
// It continuously increments the nonce and calculates the block hash until the target is met or mining is stopped.
func (c *Consensus) MineBlock(block *block.Block, stopChan <-chan struct{}) error {
	target := c.calculateTarget(c.difficulty)

	// Try different nonces
	for nonce := uint64(0); nonce < ^uint64(0); nonce++ {
		select {
		case <-stopChan:
			return fmt.Errorf("mining stopped")
		default:
			// Continue mining
		}

		// Set nonce
		block.Header.Nonce = nonce

		// Calculate hash
		hash := block.CalculateHash()

		// Check if hash meets target
		if c.hashLessThan(hash, target) {
			return nil // Block mined successfully
		}
	}

	return fmt.Errorf("failed to find valid nonce")
}

// UpdateDifficulty updates the difficulty based on recent block times.
// It collects block times and triggers a difficulty adjustment when enough blocks have been mined.
func (c *Consensus) UpdateDifficulty(blockTime time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add block time to history
	c.blockTimes = append(c.blockTimes, blockTime)

	// Keep only recent block times for adjustment
	if len(c.blockTimes) > int(c.config.DifficultyAdjustmentInterval) {
		c.blockTimes = c.blockTimes[1:]
	}

	// Check if it's time for difficulty adjustment
	if len(c.blockTimes) == int(c.config.DifficultyAdjustmentInterval) {
		c.adjustDifficulty()
	}
}

// adjustDifficulty adjusts the difficulty based on recent block times.
// It aims to keep the average block time close to the TargetBlockTime.
func (c *Consensus) adjustDifficulty() {
	// Get the current height from the chain
	currentHeight := c.chain.GetHeight()
	if currentHeight < c.config.DifficultyAdjustmentInterval {
		// Not enough blocks to adjust difficulty yet
		return
	}

	// Get the current block (tip of the chain)
	currentBlock := c.chain.GetBlockByHeight(currentHeight)
	if currentBlock == nil {
		return
	}

	// Get the block from DifficultyAdjustmentInterval ago
	oldBlockHeight := currentHeight - c.config.DifficultyAdjustmentInterval
	oldBlock := c.chain.GetBlockByHeight(oldBlockHeight)
	if oldBlock == nil {
		return
	}

	// Calculate actual time taken for the last DifficultyAdjustmentInterval blocks
	actualTime := currentBlock.Header.Timestamp.Sub(oldBlock.Header.Timestamp)

	// Calculate expected time for the last DifficultyAdjustmentInterval blocks
	expectedTime := time.Duration(c.config.DifficultyAdjustmentInterval) * c.config.TargetBlockTime

	// Calculate adjustment factor
	adjustmentFactor := float64(actualTime) / float64(expectedTime)

	// Apply damping to prevent large swings
	if adjustmentFactor < 1.0/c.config.DifficultyAdjustmentFactor {
		adjustmentFactor = 1.0 / c.config.DifficultyAdjustmentFactor
	}
	if adjustmentFactor > c.config.DifficultyAdjustmentFactor {
		adjustmentFactor = c.config.DifficultyAdjustmentFactor
	}

	// Adjust difficulty
	oldDifficulty := c.difficulty
	newDifficulty := uint64(float64(oldDifficulty) * adjustmentFactor)

	// Ensure difficulty is within bounds
	if newDifficulty < c.config.MinDifficulty {
		newDifficulty = c.config.MinDifficulty
	}
	if newDifficulty > c.config.MaxDifficulty {
		newDifficulty = c.config.MaxDifficulty
	}

	c.difficulty = newDifficulty
	c.lastAdjustment = time.Now()

	fmt.Printf("Difficulty adjusted from %d to %d (actual time: %v, expected time: %v)\n",
		oldDifficulty, c.difficulty, actualTime, expectedTime)
}

// GetDifficulty returns the current mining difficulty.
func (c *Consensus) GetDifficulty() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.difficulty
}

// GetTarget returns the current target hash for the mining difficulty.
func (c *Consensus) GetTarget() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.calculateTarget(c.difficulty)
}

// GetNextDifficulty calculates and returns what the next difficulty would be
// based on the collected block times, without actually adjusting the current difficulty.
func (c *Consensus) GetNextDifficulty() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If we have enough block times, calculate what the next difficulty would be
	if len(c.blockTimes) == int(c.config.DifficultyAdjustmentInterval) {
		// Calculate what the difficulty would be after adjustment
		totalTime := time.Duration(0)
		for _, blockTime := range c.blockTimes {
			totalTime += blockTime
		}
		averageTime := totalTime / time.Duration(len(c.blockTimes))

		targetTime := c.config.TargetBlockTime * time.Duration(c.config.DifficultyAdjustmentInterval)
		nextDifficulty := c.difficulty

		if averageTime < targetTime/4 {
			nextDifficulty = uint64(float64(nextDifficulty) * c.config.DifficultyAdjustmentFactor)
		} else if averageTime < targetTime/2 {
			nextDifficulty = uint64(float64(nextDifficulty) * 2)
		} else if averageTime > targetTime*4 {
			nextDifficulty = uint64(float64(nextDifficulty) / c.config.DifficultyAdjustmentFactor)
		} else if averageTime > targetTime*2 {
			nextDifficulty = uint64(float64(nextDifficulty) / 2)
		}

		// Ensure difficulty is within bounds
		if nextDifficulty < c.config.MinDifficulty {
			nextDifficulty = c.config.MinDifficulty
		}
		if nextDifficulty > c.config.MaxDifficulty {
			nextDifficulty = c.config.MaxDifficulty
		}

		return nextDifficulty
	}

	return c.difficulty
}

// GetStats returns a map of current consensus statistics.
func (c *Consensus) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["difficulty"] = c.difficulty
	stats["next_difficulty"] = c.GetNextDifficulty()
	stats["target"] = fmt.Sprintf("%x", c.calculateTarget(c.difficulty))
	stats["block_times_count"] = len(c.blockTimes)
	stats["last_adjustment"] = c.lastAdjustment
	stats["target_block_time"] = c.config.TargetBlockTime
	stats["adjustment_interval"] = c.config.DifficultyAdjustmentInterval

	return stats
}

// String returns a human-readable string representation of the consensus state.
func (c *Consensus) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("Consensus{Difficulty: %d, Target: %x, BlockTimes: %d}",
		c.difficulty, c.calculateTarget(c.difficulty), len(c.blockTimes))
}
