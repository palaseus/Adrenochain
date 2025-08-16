package storage

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelDBStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create LevelDB storage
	config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
	storage, err := NewLevelDBStorage(config)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("Store and retrieve block", func(t *testing.T) {
		// Create a test block
		prevHash := make([]byte, 32)
		testBlock := block.NewBlock(prevHash, 1, 1000)

		// Add a transaction
		tx := &block.Transaction{
			Version: 1,
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{
				{Value: 100, ScriptPubKey: []byte{0x01, 0x02, 0x03}},
			},
			LockTime: 0,
			Fee:      0,
		}
		testBlock.AddTransaction(tx)

		// Store the block
		err := storage.StoreBlock(testBlock)
		assert.NoError(t, err)

		// Retrieve the block
		blockHash := testBlock.CalculateHash()
		retrievedBlock, err := storage.GetBlock(blockHash)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedBlock)
		assert.Equal(t, testBlock.Header.Height, retrievedBlock.Header.Height)
		assert.Equal(t, len(testBlock.Transactions), len(retrievedBlock.Transactions))
	})

	t.Run("Store and retrieve chain state", func(t *testing.T) {
		state := &ChainState{
			BestBlockHash: make([]byte, 32),
			Height:        42,
		}

		// Store the chain state
		err := storage.StoreChainState(state)
		assert.NoError(t, err)

		// Retrieve the chain state
		retrievedState, err := storage.GetChainState()
		assert.NoError(t, err)
		assert.NotNil(t, retrievedState)
		assert.Equal(t, state.Height, retrievedState.Height)
		assert.Equal(t, state.BestBlockHash, retrievedState.BestBlockHash)
	})

	t.Run("Key-value operations", func(t *testing.T) {
		key := []byte("test_key")
		value := []byte("test_value")

		// Write key-value pair
		err := storage.Write(key, value)
		assert.NoError(t, err)

		// Read key-value pair
		retrievedValue, err := storage.Read(key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		// Check if key exists
		exists, err := storage.Has(key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Delete key-value pair
		err = storage.Delete(key)
		assert.NoError(t, err)

		// Check if key still exists
		exists, err = storage.Has(key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Block not found", func(t *testing.T) {
		nonexistentHash := make([]byte, 32)
		for i := range nonexistentHash {
			nonexistentHash[i] = 0xff
		}

		_, err := storage.GetBlock(nonexistentHash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "block not found")
	})

	t.Run("Invalid operations", func(t *testing.T) {
		// Try to store nil block
		err := storage.StoreBlock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot store nil block")

		// Try to store nil chain state
		err = storage.StoreChainState(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot store nil chain state")

		// Try to write with nil key
		err = storage.Write(nil, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key")

		// Try to write with nil value
		err = storage.Write([]byte("key"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
	})

	t.Run("Database compaction", func(t *testing.T) {
		// Add some data to trigger compaction
		for i := 0; i < 100; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			value := []byte(fmt.Sprintf("value_%d", i))
			err := storage.Write(key, value)
			assert.NoError(t, err)
		}

		// Compact the database
		err := storage.Compact()
		assert.NoError(t, err)
	})

	t.Run("Get stats", func(t *testing.T) {
		stats := storage.GetStats()
		assert.NotNil(t, stats)
		assert.Equal(t, tempDir, stats["data_dir"])
		assert.True(t, stats["db_open"].(bool))
	})
}

func TestLevelDBStorageConfig(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		config := DefaultLevelDBStorageConfig()
		assert.Equal(t, "./data/leveldb", config.DataDir)
		assert.Equal(t, 64*1024*1024, config.WriteBufferSize)
		assert.Equal(t, 1000, config.OpenFilesCacheCapacity)
		assert.True(t, config.Compression)
	})

	t.Run("Configuration chaining", func(t *testing.T) {
		config := DefaultLevelDBStorageConfig().
			WithDataDir("/custom/path").
			WithWriteBufferSize(128 * 1024 * 1024).
			WithOpenFilesCacheCapacity(2000).
			WithCompression(false)

		assert.Equal(t, "/custom/path", config.DataDir)
		assert.Equal(t, 128*1024*1024, config.WriteBufferSize)
		assert.Equal(t, 2000, config.OpenFilesCacheCapacity)
		assert.False(t, config.Compression)
	})
}

func TestLevelDBStoragePersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "leveldb_persistence_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create LevelDB storage and add some data
	config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
	storage1, err := NewLevelDBStorage(config)
	require.NoError(t, err)

	// Add a block
	prevHash := make([]byte, 32)
	testBlock := block.NewBlock(prevHash, 1, 1000)
	err = storage1.StoreBlock(testBlock)
	require.NoError(t, err)

	// Add chain state
	state := &ChainState{
		BestBlockHash: testBlock.CalculateHash(),
		Height:        1,
	}
	err = storage1.StoreChainState(state)
	require.NoError(t, err)

	// Close the first storage
	storage1.Close()

	// Create a new storage instance pointing to the same directory
	storage2, err := NewLevelDBStorage(config)
	require.NoError(t, err)
	defer storage2.Close()

	// Verify data persistence
	retrievedBlock, err := storage2.GetBlock(testBlock.CalculateHash())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedBlock)
	assert.Equal(t, testBlock.Header.Height, retrievedBlock.Header.Height)

	retrievedState, err := storage2.GetChainState()
	assert.NoError(t, err)
	assert.NotNil(t, retrievedState)
	assert.Equal(t, state.Height, retrievedState.Height)
	assert.Equal(t, state.BestBlockHash, retrievedState.BestBlockHash)
}

func TestLevelDBStorageConcurrency(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "leveldb_concurrency_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create LevelDB storage
	config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
	storage, err := NewLevelDBStorage(config)
	require.NoError(t, err)
	defer storage.Close()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := []byte(fmt.Sprintf("concurrent_key_%d", id))
			value := []byte(fmt.Sprintf("concurrent_value_%d", id))

			err := storage.Write(key, value)
			assert.NoError(t, err)

			// Read back the value
			retrievedValue, err := storage.Read(key)
			assert.NoError(t, err)
			assert.Equal(t, value, retrievedValue)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestLevelDBStorageEdgeCases tests edge cases and error conditions
func TestLevelDBStorageEdgeCases(t *testing.T) {
	t.Run("GetBlockEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)

		// Test getting block with nil hash
		block, err := storage.GetBlock(nil)
		assert.Error(t, err)
		assert.Nil(t, block)
		assert.Contains(t, err.Error(), "invalid hash: cannot be nil or empty")

		// Test getting block with empty hash
		block, err = storage.GetBlock([]byte{})
		assert.Error(t, err)
		assert.Nil(t, block)
		assert.Contains(t, err.Error(), "invalid hash: cannot be nil or empty")

		// Test getting block with zero hash
		zeroHash := make([]byte, 32)
		block, err = storage.GetBlock(zeroHash)
		assert.Error(t, err)
		assert.Nil(t, block)
		assert.Contains(t, err.Error(), "block not found")

		// Test getting block with invalid hash length
		invalidHash := []byte{1, 2, 3} // Too short
		block, err = storage.GetBlock(invalidHash)
		assert.Error(t, err)
		assert.Nil(t, block)
	})

	t.Run("GetChainStateEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)

		// Test getting chain state when none exists
		state, err := storage.GetChainState()
		assert.NoError(t, err)
		assert.NotNil(t, state)
		// Returns empty ChainState when none exists

		// Test getting chain state after storing invalid state
		invalidState := &ChainState{
			BestBlockHash: []byte{1, 2, 3}, // Invalid hash
			Height:        0,               // Invalid height
		}
		err = storage.StoreChainState(invalidState)
		assert.NoError(t, err)

		// Should still be able to retrieve it
		retrievedState, err := storage.GetChainState()
		assert.NoError(t, err)
		assert.NotNil(t, retrievedState)
		assert.Equal(t, invalidState.Height, retrievedState.Height)
	})

	t.Run("ReadWriteEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)

		// Test writing nil key
		err = storage.Write(nil, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test writing nil value
		err = storage.Write([]byte("key"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value: cannot be nil")

		// Test writing empty key
		err = storage.Write([]byte{}, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test reading nil key
		value, err := storage.Read(nil)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test reading empty key
		value, err = storage.Read([]byte{})
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test reading non-existent key
		value, err = storage.Read([]byte("nonexistent"))
		assert.Error(t, err)
		assert.Nil(t, value)
		// Returns leveldb.ErrNotFound

		// Test Has with nil key
		exists, err := storage.Has(nil)
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test Has with empty key
		exists, err = storage.Has([]byte{})
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")
	})

	t.Run("DeleteEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)

		// Test deleting nil key
		err = storage.Delete(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test deleting empty key
		err = storage.Delete([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test deleting non-existent key
		err = storage.Delete([]byte("nonexistent"))
		assert.NoError(t, err) // Deleting non-existent key should not error

		// Test deleting after writing
		err = storage.Write([]byte("test"), []byte("value"))
		assert.NoError(t, err)

		// Verify it exists
		exists, err := storage.Has([]byte("test"))
		assert.NoError(t, err)
		assert.True(t, exists)

		// Delete it
		err = storage.Delete([]byte("test"))
		assert.NoError(t, err)

		// Verify it's gone
		exists, err = storage.Has([]byte("test"))
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CompactEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)

		// Test compacting empty database
		err = storage.Compact()
		assert.NoError(t, err)

		// Test compacting after some operations
		for i := 0; i < 100; i++ {
			key := []byte(fmt.Sprintf("key%d", i))
			value := []byte(fmt.Sprintf("value%d", i))
			err = storage.Write(key, value)
			assert.NoError(t, err)
		}

		// Delete some keys to create fragmentation
		for i := 0; i < 50; i++ {
			key := []byte(fmt.Sprintf("key%d", i))
			err = storage.Delete(key)
			assert.NoError(t, err)
		}

		// Compact the database
		err = storage.Compact()
		assert.NoError(t, err)

		// Verify remaining keys are still accessible
		for i := 50; i < 100; i++ {
			key := []byte(fmt.Sprintf("key%d", i))
			value, err := storage.Read(key)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("value%d", i), string(value))
		}
	})



	t.Run("ConcurrentAccessEdgeCases", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Test concurrent writes
		var wg sync.WaitGroup
		concurrency := 10
		operations := 100

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					key := []byte(fmt.Sprintf("key_%d_%d", id, j))
					value := []byte(fmt.Sprintf("value_%d_%d", id, j))
					err := storage.Write(key, value)
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify all operations were successful
		for i := 0; i < concurrency; i++ {
			for j := 0; j < operations; j++ {
				key := []byte(fmt.Sprintf("key_%d_%d", i, j))
				value, err := storage.Read(key)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("value_%d_%d", i, j), string(value))
			}
		}
	})

	t.Run("LargeDataHandling", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Test with large values
		largeValue := make([]byte, 1024*1024) // 1MB
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}

		err = storage.Write([]byte("large_key"), largeValue)
		assert.NoError(t, err)

		// Read it back
		retrievedValue, err := storage.Read([]byte("large_key"))
		assert.NoError(t, err)
		assert.Equal(t, largeValue, retrievedValue)

		// Test with many small keys
		for i := 0; i < 1000; i++ {
			key := []byte(fmt.Sprintf("small_key_%d", i))
			value := []byte(fmt.Sprintf("small_value_%d", i))
			err = storage.Write(key, value)
			assert.NoError(t, err)
		}

		// Verify all small keys
		for i := 0; i < 1000; i++ {
			key := []byte(fmt.Sprintf("small_key_%d", i))
			value, err := storage.Read(key)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("small_value_%d", i), string(value))
		}
	})

	t.Run("RecoveryAfterErrors", func(t *testing.T) {
		tempDir := t.TempDir()
		config := DefaultLevelDBStorageConfig().WithDataDir(tempDir)
		storage, err := NewLevelDBStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Write some data
		err = storage.Write([]byte("recovery_key"), []byte("recovery_value"))
		assert.NoError(t, err)

		// Simulate some operations that might fail
		// Test with invalid operations that should be handled gracefully
		err = storage.Write([]byte(""), []byte("value")) // Empty key
		assert.Error(t, err)

		// Verify the database is still functional
		value, err := storage.Read([]byte("recovery_key"))
		assert.NoError(t, err)
		assert.Equal(t, "recovery_value", string(value))

		// Test that we can still write new data
		err = storage.Write([]byte("new_key"), []byte("new_value"))
		assert.NoError(t, err)

		// Verify new data
		value, err = storage.Read([]byte("new_key"))
		assert.NoError(t, err)
		assert.Equal(t, "new_value", string(value))
	})
}
