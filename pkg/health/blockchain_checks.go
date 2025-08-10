package health

import (
	"fmt"
	"time"

	"github.com/gochain/gochain/pkg/chain"
)

// ChainHealthChecker checks the health of the blockchain
type ChainHealthChecker struct {
	chain *chain.Chain
	name  string
}

// NewChainHealthChecker creates a new chain health checker
func NewChainHealthChecker(chain *chain.Chain) *ChainHealthChecker {
	return &ChainHealthChecker{
		chain: chain,
		name:  "blockchain",
	}
}

// Name returns the name of this health checker
func (c *ChainHealthChecker) Name() string {
	return c.name
}

// Check performs a health check on the blockchain
func (c *ChainHealthChecker) Check() (*Component, error) {
	start := time.Now()

	// Get current chain state
	height := c.chain.GetHeight()
	bestBlock := c.chain.GetBestBlock()

	if bestBlock == nil {
		return &Component{
			Name:      c.Name(),
			Status:    StatusUnhealthy,
			Message:   "No best block available",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height": height,
				"error":  "best block is nil",
			},
		}, nil
	}

	// Check if the best block hash matches the expected hash
	expectedHash := bestBlock.CalculateHash()
	if expectedHash == nil {
		return &Component{
			Name:      c.Name(),
			Status:    StatusUnhealthy,
			Message:   "Failed to calculate best block hash",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height": height,
				"error":  "hash calculation failed",
			},
		}, nil
	}

	// Check if the chain has reasonable height (not stuck at 0)
	if height == 0 && bestBlock.Header.Height == 0 {
		// This might be normal for a new chain, but let's check if it's the genesis block
		genesisBlock := c.chain.GetGenesisBlock()
		if genesisBlock == nil {
			return &Component{
				Name:      c.Name(),
				Status:    StatusUnhealthy,
				Message:   "No genesis block available",
				LastCheck: time.Now(),
				CheckTime: time.Since(start),
				Details: map[string]interface{}{
					"height": height,
					"error":  "genesis block missing",
				},
			}, nil
		}
	}

	// Check if the last block is recent (within reasonable time)
	now := time.Now()
	blockAge := now.Sub(bestBlock.Header.Timestamp)

	// Consider block unhealthy if it's older than 1 hour (for a 10-second block time chain)
	maxBlockAge := time.Hour
	if blockAge > maxBlockAge {
		return &Component{
			Name:      c.Name(),
			Status:    StatusDegraded,
			Message:   fmt.Sprintf("Last block is %v old", blockAge),
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height":          height,
				"last_block_time": bestBlock.Header.Timestamp,
				"block_age":       blockAge.String(),
				"max_block_age":   maxBlockAge.String(),
				"best_block_hash": fmt.Sprintf("%x", expectedHash),
				"difficulty":      bestBlock.Header.Difficulty,
			},
		}, nil
	}

	// Check if difficulty is reasonable (not 0 or extremely high)
	if bestBlock.Header.Difficulty <= 0 {
		return &Component{
			Name:      c.Name(),
			Status:    StatusDegraded,
			Message:   "Block difficulty is zero or negative",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height":          height,
				"last_block_time": bestBlock.Header.Timestamp,
				"block_age":       blockAge.String(),
				"best_block_hash": fmt.Sprintf("%x", expectedHash),
				"difficulty":      bestBlock.Header.Difficulty,
				"warning":         "difficulty should be positive",
			},
		}, nil
	}

	// Chain appears healthy
	return &Component{
		Name:      c.Name(),
		Status:    StatusHealthy,
		Message:   "Blockchain is healthy",
		LastCheck: time.Now(),
		CheckTime: time.Since(start),
		Details: map[string]interface{}{
			"height":          height,
			"last_block_time": bestBlock.Header.Timestamp,
			"block_age":       blockAge.String(),
			"best_block_hash": fmt.Sprintf("%x", expectedHash),
			"difficulty":      bestBlock.Header.Difficulty,
			"transactions":    len(bestBlock.Transactions),
		},
	}, nil
}
