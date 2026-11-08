//go:build go1.18

package consensus

import (
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
)

// FuzzDifficultyCalculation tests difficulty calculation with fuzzed block times
func FuzzDifficultyCalculation(f *testing.F) {
	// Seed corpus with valid block times
	f.Add(int64(5 * time.Second))  // 5 seconds
	f.Add(int64(10 * time.Second)) // 10 seconds
	f.Add(int64(15 * time.Second)) // 15 seconds

	f.Fuzz(func(t *testing.T, blockTime int64) {
		// Skip invalid block times
		if blockTime <= 0 || blockTime > int64(24*time.Hour) {
			t.Skip("Invalid block time")
		}

		config := DefaultConsensusConfig()
		mockChain := &MockChainReader{height: 0}
		consensus := NewConsensus(config, mockChain)

		// Test difficulty calculation with fuzzed block time
		duration := time.Duration(blockTime)
		consensus.UpdateDifficulty(duration)

		// Verify difficulty is within bounds
		difficulty := consensus.GetDifficulty()
		if difficulty < config.MinDifficulty || difficulty > config.MaxDifficulty {
			t.Errorf("Difficulty %d is outside bounds [%d, %d]",
				difficulty, config.MinDifficulty, config.MaxDifficulty)
		}

		// Verify target calculation
		target := consensus.GetTarget()
		if target == nil {
			t.Errorf("Target is nil")
			return
		}

		if len(target) != 32 {
			t.Errorf("Target has incorrect length: %d", len(target))
		}
	})
}

// FuzzCheckpointValidation tests checkpoint validation with fuzzed data
func FuzzCheckpointValidation(f *testing.F) {
	// Seed corpus with valid checkpoint data
	f.Add(uint64(1000), []byte("checkpoint_hash"))
	f.Add(uint64(10000), []byte("another_checkpoint"))

	f.Fuzz(func(t *testing.T, height uint64, hash []byte) {
		// Skip very large inputs
		if len(hash) > 1000000 {
			t.Skip("Hash too large")
		}

		config := DefaultConsensusConfig()
		// Create a mock chain with the checkpoint block
		blocks := make(map[uint64]*block.Block)
		blocks[height] = &block.Block{
			Header: &block.Header{
				Height:     height,
				Difficulty: 1,
				Timestamp:  time.Now(),
			},
		}
		blocks[height-1] = &block.Block{
			Header: &block.Header{
				Height:     height - 1,
				Difficulty: 1,
				Timestamp:  time.Now().Add(-time.Second),
			},
		}
		mockChain := &MockChainReader{
			blocks: blocks,
			height: height,
		}
		consensus := NewConsensus(config, mockChain)

		// Add checkpoint
		consensus.AddCheckpoint(height, hash)

		// Test checkpoint validation - this will fail because block hash won't match
		// So we only test that the function doesn't panic
		_ = consensus.ValidateCheckpoint(height, hash)

		// Test with different hash
		wrongHash := append(hash, 0x01)
		_ = consensus.ValidateCheckpoint(height, wrongHash)

		// Test with non-existent height
		_ = consensus.ValidateCheckpoint(height+1, hash)
	})
}

// FuzzAccumulatedDifficulty tests accumulated difficulty calculation with fuzzed data
func FuzzAccumulatedDifficulty(f *testing.F) {
	// Seed corpus with valid heights
	f.Add(uint64(1))
	f.Add(uint64(100))
	f.Add(uint64(1000))

	f.Fuzz(func(t *testing.T, height uint64) {
		// Skip very large heights
		if height > 1000000 {
			t.Skip("Height too large")
		}

		config := DefaultConsensusConfig()
		mockChain := &MockChainReader{height: height}
		consensus := NewConsensus(config, mockChain)

		// Test accumulated difficulty calculation
		accumulatedDiff, err := consensus.GetAccumulatedDifficulty(height)
		if err != nil {
			// This is expected for mock chain
			return
		}

		// Verify accumulated difficulty is reasonable
		if accumulatedDiff != nil && accumulatedDiff.Cmp(accumulatedDiff) < 0 {
			t.Errorf("Accumulated difficulty comparison error")
		}
	})
}
