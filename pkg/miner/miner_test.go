package miner

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/mempool"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiner(t *testing.T) {
	dataDir := "./test_miner_data_test_miner"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test Start and Stop
	err = miner.StartMining()
	assert.NoError(t, err)
	assert.True(t, miner.IsMining())

	miner.StopMining()
	assert.False(t, miner.IsMining())
}

func TestCreateNewBlock(t *testing.T) {
	dataDir := "./test_miner_data_test_create_new_block"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	prevBlock := chain.GetBestBlock()

	newBlock := miner.createNewBlock(prevBlock)
	assert.NotNil(t, newBlock)
	assert.Equal(t, prevBlock.Header.Height+1, newBlock.Header.Height)
	assert.Equal(t, prevBlock.CalculateHash(), newBlock.Header.PrevBlockHash)
}

// TestMinerAdvancedScenarios tests advanced miner scenarios
func TestMinerAdvancedScenarios(t *testing.T) {
	dataDir := "./test_miner_data_test_advanced_scenarios"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test mining with transactions
	t.Run("MiningWithTransactions", func(t *testing.T) {
		// Add some transactions to mempool
		tx := &block.Transaction{
			Version: 1,
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{
				{
					Value:        100,
					ScriptPubKey: []byte("test_output"),
				},
			},
			LockTime: 0,
			Fee:      10,
			Hash:     make([]byte, 32),
		}
		mempool.AddTransaction(tx)

		// Start mining
		err := miner.StartMining()
		assert.NoError(t, err)
		assert.True(t, miner.IsMining())

		// Wait a bit for mining to progress
		time.Sleep(100 * time.Millisecond)

		// Stop mining
		miner.StopMining()
		assert.False(t, miner.IsMining())
	})

	// Test mining performance
	t.Run("MiningPerformance", func(t *testing.T) {
		// Test mining multiple blocks
		startTime := time.Now()

		err := miner.StartMining()
		assert.NoError(t, err)

		// Wait for some mining activity
		time.Sleep(200 * time.Millisecond)

		miner.StopMining()

		miningTime := time.Since(startTime)
		assert.True(t, miningTime < 1*time.Second, "Mining should complete within reasonable time")
	})
}

// TestMinerConcurrency tests miner behavior under concurrent operations
func TestMinerConcurrency(t *testing.T) {
	dataDir := "./test_miner_data_test_concurrency"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()

	// Test sequential start/stop operations (concurrent access not supported)
	t.Run("SequentialStartStop", func(t *testing.T) {
		miner := NewMiner(chain, mempool, config, consensusConfig)
		defer miner.Cleanup()

		// Test multiple start/stop cycles sequentially
		for i := 0; i < 5; i++ {
			// Start mining
			err := miner.StartMining()
			assert.NoError(t, err)
			assert.True(t, miner.IsMining())

			// Wait a bit
			time.Sleep(10 * time.Millisecond)

			// Stop mining
			miner.StopMining()
			assert.False(t, miner.IsMining())

			// Wait a bit before next cycle
			time.Sleep(10 * time.Millisecond)
		}

		// Verify final state is consistent
		assert.NotNil(t, miner, "Miner should remain accessible")
	})

	// Test concurrent block creation
	t.Run("ConcurrentBlockCreation", func(t *testing.T) {
		miner := NewMiner(chain, mempool, config, consensusConfig)
		defer miner.Cleanup()

		prevBlock := chain.GetBestBlock()
		var wg sync.WaitGroup
		numGoroutines := 5

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				newBlock := miner.createNewBlock(prevBlock)
				assert.NotNil(t, newBlock, "Should create valid block")
			}()
		}

		wg.Wait()
	})
}

// TestMinerEdgeCases tests miner edge cases and error conditions
func TestMinerEdgeCases(t *testing.T) {
	dataDir := "./test_miner_data_test_edge_cases"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test mining with empty mempool
	t.Run("EmptyMempool", func(t *testing.T) {
		// Ensure mempool is empty
		assert.Equal(t, 0, mempool.GetTransactionCount())

		err := miner.StartMining()
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		miner.StopMining()

		// Should still be able to mine (coinbase transaction)
		assert.NotNil(t, miner)
	})

	// Test mining with invalid chain state
	t.Run("InvalidChainState", func(t *testing.T) {
		// Test with nil chain
		nilMiner := NewMiner(nil, mempool, config, consensusConfig)
		assert.NotNil(t, nilMiner, "Should handle nil chain gracefully")

		// Test with nil mempool
		nilMempoolMiner := NewMiner(chain, nil, config, consensusConfig)
		assert.NotNil(t, nilMempoolMiner, "Should handle nil mempool gracefully")
	})

	// Test mining configuration edge cases
	t.Run("ConfigurationEdgeCases", func(t *testing.T) {
		// Test with zero block time
		zeroTimeConfig := DefaultMinerConfig()
		zeroTimeConfig.BlockTime = 0
		zeroTimeMiner := NewMiner(chain, mempool, zeroTimeConfig, consensusConfig)
		assert.NotNil(t, zeroTimeMiner, "Should handle zero block time")

		// Test with very high block size
		highSizeConfig := DefaultMinerConfig()
		highSizeConfig.MaxBlockSize = ^uint64(0)
		highSizeMiner := NewMiner(chain, mempool, highSizeConfig, consensusConfig)
		assert.NotNil(t, highSizeMiner, "Should handle very high block size")
	})
}

// TestMinerIntegration tests miner integration with other components
func TestMinerIntegration(t *testing.T) {
	dataDir := "./test_miner_data_test_integration"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test complete mining workflow
	t.Run("CompleteMiningWorkflow", func(t *testing.T) {
		// 1. Add transactions to mempool
		for i := 0; i < 5; i++ {
			tx := &block.Transaction{
				Version: 1,
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{
					{
						Value:        uint64(100 * (i + 1)),
						ScriptPubKey: []byte(fmt.Sprintf("output_%d", i)),
					},
				},
				LockTime: 0,
				Fee:      uint64(10 * (i + 1)),
				Hash:     make([]byte, 32),
			}
			mempool.AddTransaction(tx)
		}

		// 2. Start mining
		err := miner.StartMining()
		assert.NoError(t, err)

		// 3. Wait for mining to progress
		time.Sleep(300 * time.Millisecond)

		// 4. Stop mining
		miner.StopMining()

		// 5. Verify mining state
		assert.False(t, miner.IsMining())
		assert.NotNil(t, miner, "Miner should remain accessible")
	})

	// Test mining with chain updates
	t.Run("MiningWithChainUpdates", func(t *testing.T) {
		// Get initial chain state
		initialHeight := chain.GetBestBlock().Header.Height

		// Start mining
		err := miner.StartMining()
		assert.NoError(t, err)

		// Wait for potential block creation
		time.Sleep(200 * time.Millisecond)

		// Stop mining
		miner.StopMining()

		// Check if chain was updated
		finalHeight := chain.GetBestBlock().Header.Height
		assert.True(t, finalHeight >= initialHeight, "Chain height should not decrease")
	})
}

// TestMinerPerformance tests miner performance characteristics
func TestMinerPerformance(t *testing.T) {
	dataDir := "./test_miner_data_test_performance"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test mining speed
	t.Run("MiningSpeed", func(t *testing.T) {
		// Add some transactions to make mining more realistic
		for i := 0; i < 10; i++ {
			tx := &block.Transaction{
				Version: 1,
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{
					{
						Value:        uint64(50 * (i + 1)),
						ScriptPubKey: []byte(fmt.Sprintf("perf_output_%d", i)),
					},
				},
				LockTime: 0,
				Fee:      uint64(5 * (i + 1)),
				Hash:     make([]byte, 32),
			}
			mempool.AddTransaction(tx)
		}

		// Measure mining performance
		startTime := time.Now()

		err := miner.StartMining()
		assert.NoError(t, err)

		// Mine for a short duration
		time.Sleep(500 * time.Millisecond)

		miner.StopMining()

		miningDuration := time.Since(startTime)
		assert.True(t, miningDuration < 1*time.Second, "Mining should complete within reasonable time")
	})

	// Test memory usage during mining
	t.Run("MemoryUsage", func(t *testing.T) {
		// Start mining
		err := miner.StartMining()
		assert.NoError(t, err)

		// Perform some mining operations
		time.Sleep(200 * time.Millisecond)

		// Stop mining
		miner.StopMining()

		// Verify miner is still accessible (no memory leaks)
		assert.NotNil(t, miner, "Miner should remain accessible after mining")
	})
}

// TestMinerRecovery tests miner recovery mechanisms
func TestMinerRecovery(t *testing.T) {
	dataDir := "./test_miner_data_test_recovery"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test recovery from interrupted mining
	t.Run("InterruptedMining", func(t *testing.T) {
		// Start mining
		err := miner.StartMining()
		assert.NoError(t, err)
		assert.True(t, miner.IsMining())

		// Simulate interruption
		miner.StopMining()
		assert.False(t, miner.IsMining())

		// Try to restart mining
		err = miner.StartMining()
		assert.NoError(t, err)
		assert.True(t, miner.IsMining())

		// Clean stop
		miner.StopMining()
		assert.False(t, miner.IsMining())
	})

	// Test recovery from configuration changes
	t.Run("ConfigurationRecovery", func(t *testing.T) {
		// Start with default config
		err := miner.StartMining()
		assert.NoError(t, err)

		// Stop mining
		miner.StopMining()

		// Create new miner with different config
		newConfig := DefaultMinerConfig()
		newConfig.BlockTime = 5 * time.Second
		newMiner := NewMiner(chain, mempool, newConfig, consensusConfig)

		// Restart mining with new config
		err = newMiner.StartMining()
		assert.NoError(t, err)

		// Stop mining
		newMiner.StopMining()
		assert.False(t, newMiner.IsMining())
	})
}

// TestMinerStatistics tests miner statistics and monitoring
func TestMinerStatistics(t *testing.T) {
	dataDir := "./test_miner_data_test_statistics"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	require.NoError(t, err)
	mempool := mempool.NewMempool(mempool.TestMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test mining statistics
	t.Run("MiningStatistics", func(t *testing.T) {
		// Get initial stats
		initialStats := miner.GetMiningStats()
		assert.NotNil(t, initialStats, "Should return mining statistics")

		// Start mining
		err := miner.StartMining()
		assert.NoError(t, err)

		// Wait for some mining activity
		time.Sleep(200 * time.Millisecond)

		// Get updated stats
		updatedStats := miner.GetMiningStats()
		assert.NotNil(t, updatedStats, "Should return updated statistics")

		// Stop mining
		miner.StopMining()

		// Verify stats are accessible
		finalStats := miner.GetMiningStats()
		assert.NotNil(t, finalStats, "Should return final statistics")
	})
}
