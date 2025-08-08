package consensus

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// Consensus represents the blockchain consensus mechanism
type Consensus struct {
	mu            sync.RWMutex
	config        *ConsensusConfig
	difficulty    uint64
	lastAdjustment time.Time
	blockTimes    []time.Duration
}

// ConsensusConfig holds configuration for consensus
type ConsensusConfig struct {
	TargetBlockTime        time.Duration
	DifficultyAdjustmentInterval uint64
	MaxDifficulty          uint64
	MinDifficulty          uint64
	DifficultyAdjustmentFactor float64
}

// DefaultConsensusConfig returns the default consensus configuration
func DefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		TargetBlockTime:            10 * time.Second,
		DifficultyAdjustmentInterval: 2016,
		MaxDifficulty:              256,
		MinDifficulty:              1,
		DifficultyAdjustmentFactor: 4.0,
	}
}

// NewConsensus creates a new consensus instance
func NewConsensus(config *ConsensusConfig) *Consensus {
	return &Consensus{
		config:        config,
		difficulty:    config.MinDifficulty,
		lastAdjustment: time.Now(),
		blockTimes:    make([]time.Duration, 0),
	}
}

// ValidateBlock validates a block according to consensus rules
func (c *Consensus) ValidateBlock(block *block.Block, prevBlock *block.Block) error {
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
	if block.Header.Difficulty != c.difficulty {
		return fmt.Errorf("block difficulty %d does not match expected %d", 
			block.Header.Difficulty, c.difficulty)
	}
	
	return nil
}

// ValidateProofOfWork validates the proof of work for a block
func (c *Consensus) ValidateProofOfWork(block *block.Block) bool {
	hash := block.CalculateHash()
	target := c.calculateTarget(c.difficulty)
	
	return c.hashLessThan(hash, target)
}

// calculateTarget calculates the target hash for a given difficulty
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

// hashLessThan checks if hash1 is less than hash2
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

// MineBlock mines a block with the given difficulty
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

// UpdateDifficulty updates the difficulty based on recent block times
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

// adjustDifficulty adjusts the difficulty based on recent block times
func (c *Consensus) adjustDifficulty() {
	if len(c.blockTimes) < int(c.config.DifficultyAdjustmentInterval) {
		return
	}
	
	// Calculate average block time
	totalTime := time.Duration(0)
	for _, blockTime := range c.blockTimes {
		totalTime += blockTime
	}
	averageTime := totalTime / time.Duration(len(c.blockTimes))
	
	// Calculate target time
	targetTime := c.config.TargetBlockTime * time.Duration(c.config.DifficultyAdjustmentInterval)
	
	// Adjust difficulty
	oldDifficulty := c.difficulty
	if averageTime < targetTime/4 {
		// Blocks are coming too fast, increase difficulty significantly
		c.difficulty = uint64(float64(c.difficulty) * c.config.DifficultyAdjustmentFactor)
	} else if averageTime < targetTime/2 {
		// Blocks are coming fast, increase difficulty moderately
		c.difficulty = uint64(float64(c.difficulty) * 2)
	} else if averageTime > targetTime*4 {
		// Blocks are coming too slow, decrease difficulty significantly
		c.difficulty = uint64(float64(c.difficulty) / c.config.DifficultyAdjustmentFactor)
	} else if averageTime > targetTime*2 {
		// Blocks are coming slow, decrease difficulty moderately
		c.difficulty = uint64(float64(c.difficulty) / 2)
	}
	
	// Ensure difficulty is within bounds
	if c.difficulty < c.config.MinDifficulty {
		c.difficulty = c.config.MinDifficulty
	}
	if c.difficulty > c.config.MaxDifficulty {
		c.difficulty = c.config.MaxDifficulty
	}
	
	// Reset block times for next adjustment
	c.blockTimes = make([]time.Duration, 0)
	c.lastAdjustment = time.Now()
	
	fmt.Printf("Difficulty adjusted from %d to %d (avg block time: %v, target: %v)\n",
		oldDifficulty, c.difficulty, averageTime, c.config.TargetBlockTime)
}

// GetDifficulty returns the current difficulty
func (c *Consensus) GetDifficulty() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.difficulty
}

// GetTarget returns the current target hash
func (c *Consensus) GetTarget() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.calculateTarget(c.difficulty)
}

// GetNextDifficulty returns the next difficulty that will be used
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

// GetStats returns consensus statistics
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

// String returns a string representation of the consensus
func (c *Consensus) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return fmt.Sprintf("Consensus{Difficulty: %d, Target: %x, BlockTimes: %d}", 
		c.difficulty, c.calculateTarget(c.difficulty), len(c.blockTimes))
} 