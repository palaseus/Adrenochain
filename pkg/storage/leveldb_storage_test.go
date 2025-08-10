package storage

import (
	"fmt"
	"os"
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
