package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	dataDir := "./test_data"
	defer os.RemoveAll(dataDir)

	config := &StorageConfig{DataDir: dataDir}
	s, err := NewStorage(config)
	assert.NoError(t, err)
	defer s.Close()

	// Test StoreBlock and GetBlock
	b := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now(),
			Difficulty:    1,
			Height:        1,
		},
	}
	b.Header.MerkleRoot = b.CalculateMerkleRoot()

	err = s.StoreBlock(b)
	assert.NoError(t, err)

	retrievedBlock, err := s.GetBlock(b.CalculateHash())
	assert.NoError(t, err)
	assert.Equal(t, b.HexHash(), retrievedBlock.HexHash())

	// Test StoreChainState and GetChainState
	state := &ChainState{
		BestBlockHash: b.CalculateHash(),
		Height:        1,
	}

	err = s.StoreChainState(state)
	assert.NoError(t, err)

	retrievedState, err := s.GetChainState()
	assert.NoError(t, err)
	assert.Equal(t, state.BestBlockHash, retrievedState.BestBlockHash)
	assert.Equal(t, state.Height, retrievedState.Height)
}

// TestStorageErrorHandling tests comprehensive error scenarios
func TestStorageErrorHandling(t *testing.T) {
	// Test with invalid data directory (no write permissions)
	t.Run("InvalidDataDirectory", func(t *testing.T) {
		// Try to create storage in a directory that doesn't exist and can't be created
		invalidConfig := &StorageConfig{DataDir: "/root/invalid/path/that/cannot/be/created"}
		_, err := NewStorage(invalidConfig)
		assert.Error(t, err)
		// The error message can vary depending on the system, so just check that it's an error
		assert.True(t, strings.Contains(err.Error(), "failed to create") ||
			strings.Contains(err.Error(), "permission denied") ||
			strings.Contains(err.Error(), "no such file"))
	})

	// Test with nil block
	t.Run("StoreNilBlock", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		err = storage.StoreBlock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot store nil block")
	})

	// Test with nil chain state
	t.Run("StoreNilChainState", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		err = storage.StoreChainState(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot store nil chain state")
	})

	// Test with invalid hash
	t.Run("GetBlockWithInvalidHash", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Test nil hash
		_, err = storage.GetBlock(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hash: cannot be nil or empty")

		// Test empty hash
		_, err = storage.GetBlock([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hash: cannot be nil or empty")
	})

	// Test with invalid key-value operations
	t.Run("InvalidKeyValueOperations", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Test nil key
		err = storage.Write(nil, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test empty key
		err = storage.Write([]byte{}, []byte("value"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test nil value
		err = storage.Write([]byte("key"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value: cannot be nil")

		// Test nil key for read
		_, err = storage.Read(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test nil key for delete
		err = storage.Delete(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")

		// Test nil key for has
		_, err = storage.Has(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key: cannot be nil or empty")
	})
}

// TestStorageEdgeCases tests edge cases and boundary conditions
func TestStorageEdgeCases(t *testing.T) {
	t.Run("EmptyDataDirectory", func(t *testing.T) {
		// Test with empty data directory - this should fail as it's invalid
		_, err := NewStorage(&StorageConfig{DataDir: ""})
		assert.Error(t, err)
		// Empty data directory should not be allowed - check for various error messages
		assert.True(t, strings.Contains(err.Error(), "failed to create") ||
			strings.Contains(err.Error(), "no such file") ||
			strings.Contains(err.Error(), "mkdir"))
	})

	t.Run("VeryLongKeys", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Test with long key (but not too long for filesystem)
		// Using 100 bytes = 200 hex characters, which is well under filesystem limits
		longKey := make([]byte, 100)
		for i := range longKey {
			longKey[i] = byte(i % 256)
		}
		value := []byte("test_value")

		err = storage.Write(longKey, value)
		assert.NoError(t, err)

		// Read it back
		readValue, err := storage.Read(longKey)
		assert.NoError(t, err)
		assert.Equal(t, value, readValue)
	})

	t.Run("VeryLargeValues", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Test with very large value (1MB)
		largeValue := make([]byte, 1024*1024)
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}
		key := []byte("large_key")

		err = storage.Write(key, largeValue)
		assert.NoError(t, err)

		// Read it back
		readValue, err := storage.Read(key)
		assert.NoError(t, err)
		assert.Equal(t, largeValue, readValue)
	})

	t.Run("SpecialCharactersInKeys", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Test with special characters in key
		specialKey := []byte("key_with_special_chars_!@#$%^&*()_+-=[]{}|;':\",./<>?")
		value := []byte("special_value")

		err = storage.Write(specialKey, value)
		assert.NoError(t, err)

		// Read it back
		readValue, err := storage.Read(specialKey)
		assert.NoError(t, err)
		assert.Equal(t, value, readValue)
	})
}

// TestStorageConcurrencyAdvanced tests advanced concurrency scenarios
func TestStorageConcurrencyAdvanced(t *testing.T) {
	t.Run("ConcurrentBlockOperations", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		const numGoroutines = 10
		const blocksPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < blocksPerGoroutine; j++ {
					block := &block.Block{
						Header: &block.Header{
							Version:       1,
							PrevBlockHash: make([]byte, 32),
							MerkleRoot:    make([]byte, 32),
							Timestamp:     time.Now(),
							Difficulty:    uint64(id*1000 + j),
							Nonce:         uint64(id*100 + j),
							Height:        uint64(id*1000 + j),
						},
						Transactions: []*block.Transaction{},
					}

					err := storage.StoreBlock(block)
					assert.NoError(t, err)

					// Read it back
					hash := block.CalculateHash()
					retrievedBlock, err := storage.GetBlock(hash)
					assert.NoError(t, err)
					assert.Equal(t, block.Header.Height, retrievedBlock.Header.Height)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("ConcurrentKeyValueOperations", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		const numGoroutines = 20
		const operationsPerGoroutine = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := []byte(fmt.Sprintf("key_%d_%d", id, j))
					value := []byte(fmt.Sprintf("value_%d_%d", id, j))

					// Write
					err := storage.Write(key, value)
					assert.NoError(t, err)

					// Read
					readValue, err := storage.Read(key)
					assert.NoError(t, err)
					assert.Equal(t, value, readValue)

					// Check exists
					exists, err := storage.Has(key)
					assert.NoError(t, err)
					assert.True(t, exists)

					// Delete
					err = storage.Delete(key)
					assert.NoError(t, err)

					// Check no longer exists
					exists, err = storage.Has(key)
					assert.NoError(t, err)
					assert.False(t, exists)
				}
			}(i)
		}

		wg.Wait()
	})
}

// TestStoragePerformance tests performance and stress scenarios
func TestStoragePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("BulkBlockStorage", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		const numBlocks = 100 // Reduced from 1000 for faster testing
		start := time.Now()

		// Store blocks and keep track of their hashes
		blockHashes := make([][]byte, numBlocks)
		
		// Store many blocks
		for i := 0; i < numBlocks; i++ {
			// Create unique block data to ensure different hashes
			prevHash := make([]byte, 32)
			copy(prevHash, []byte(fmt.Sprintf("prev_%d", i)))
			
			merkleRoot := make([]byte, 32)
			copy(merkleRoot, []byte(fmt.Sprintf("merkle_%d", i)))
			
			block := &block.Block{
				Header: &block.Header{
					Version:       1,
					PrevBlockHash: prevHash,
					MerkleRoot:    merkleRoot,
					Timestamp:     time.Now().Add(time.Duration(i) * time.Millisecond), // Unique timestamp
					Difficulty:    uint64(i),
					Nonce:         uint64(i),
					Height:        uint64(i),
				},
				Transactions: []*block.Transaction{},
			}

			// Store the block and save its hash
			err := storage.StoreBlock(block)
			assert.NoError(t, err)
			blockHashes[i] = block.CalculateHash()
		}

		storageTime := time.Since(start)
		t.Logf("Stored %d blocks in %v", numBlocks, storageTime)

		// Verify all blocks can be retrieved using the stored hashes
		start = time.Now()
		for i := 0; i < numBlocks; i++ {
			hash := blockHashes[i]
			require.NotNil(t, hash, "Block hash should not be nil")
			
			retrievedBlock, err := storage.GetBlock(hash)
			assert.NoError(t, err)
			assert.Equal(t, uint64(i), retrievedBlock.Header.Height)
		}

		retrievalTime := time.Since(start)
		t.Logf("Retrieved %d blocks in %v", numBlocks, retrievalTime)
	})

	t.Run("BulkKeyValueOperations", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		const numOperations = 5000
		start := time.Now()

		// Write many key-value pairs
		for i := 0; i < numOperations; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			value := []byte(fmt.Sprintf("value_%d", i))

			err := storage.Write(key, value)
			assert.NoError(t, err)
		}

		writeTime := time.Since(start)
		t.Logf("Wrote %d key-value pairs in %v", numOperations, writeTime)

		// Read all key-value pairs
		start = time.Now()
		for i := 0; i < numOperations; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			expectedValue := []byte(fmt.Sprintf("value_%d", i))

			value, err := storage.Read(key)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, value)
		}

		readTime := time.Since(start)
		t.Logf("Read %d key-value pairs in %v", numOperations, readTime)
	})
}

// TestStorageRecovery tests recovery from corrupted or invalid data
func TestStorageRecovery(t *testing.T) {
	t.Run("CorruptedBlockFile", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Create a valid block
		block := &block.Block{
			Header: &block.Header{
				Version:       1,
				PrevBlockHash: make([]byte, 32),
				MerkleRoot:    make([]byte, 32),
				Timestamp:     time.Now(),
				Difficulty:    1000,
				Nonce:         0,
				Height:        100,
			},
			Transactions: []*block.Transaction{},
		}

		// Store the block
		err = storage.StoreBlock(block)
		assert.NoError(t, err)

		// Corrupt the block file by writing invalid JSON
		hash := block.CalculateHash()
		blockPath := filepath.Join(storage.dataDir, fmt.Sprintf("%x", hash))
		err = os.WriteFile(blockPath, []byte("invalid json content"), 0644)
		assert.NoError(t, err)

		// Try to read the corrupted block
		_, err = storage.GetBlock(hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode block")
	})

	t.Run("CorruptedChainStateFile", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Create a valid chain state
		state := &ChainState{
			BestBlockHash: make([]byte, 32),
			Height:        1000,
		}

		// Store the chain state
		err = storage.StoreChainState(state)
		assert.NoError(t, err)

		// Corrupt the chain state file
		chainStatePath := filepath.Join(storage.dataDir, "chainstate")
		err = os.WriteFile(chainStatePath, []byte("invalid json content"), 0644)
		assert.NoError(t, err)

		// Try to read the corrupted chain state
		_, err = storage.GetChainState()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode chain state")
	})

	t.Run("MissingDataDirectory", func(t *testing.T) {
		// Create storage in a temporary directory
		tempDir := t.TempDir()
		storage, err := NewStorage(&StorageConfig{DataDir: tempDir})
		require.NoError(t, err)

		// Store some data
		key := []byte("test_key")
		value := []byte("test_value")
		err = storage.Write(key, value)
		assert.NoError(t, err)

		// Close storage
		storage.Close()

		// Remove the data directory
		err = os.RemoveAll(tempDir)
		assert.NoError(t, err)

		// Try to create storage in the removed directory
		_, err = NewStorage(&StorageConfig{DataDir: tempDir})
		assert.NoError(t, err) // Should recreate the directory
	})
}

// TestStorageInterfaceCompliance tests that Storage implements the expected interface
func TestStorageInterfaceCompliance(t *testing.T) {
	// Test that Storage implements StorageInterface
	var _ StorageInterface = (*Storage)(nil)

	// Test that StorageConfig implements configuration pattern
	config := DefaultStorageConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "./data", config.DataDir)

	// Test configuration chaining
	config2 := config.WithDataDir("/custom/path")
	assert.Equal(t, "/custom/path", config2.DataDir)
	assert.NotEqual(t, config.DataDir, config2.DataDir) // Should not modify original
}

// TestStorageMetrics tests storage metrics and statistics
func TestStorageMetrics(t *testing.T) {
	t.Run("StorageSizeCalculation", func(t *testing.T) {
		storage, err := NewStorage(&StorageConfig{DataDir: t.TempDir()})
		require.NoError(t, err)
		defer storage.Close()

		// Store some data and calculate approximate size
		const numItems = 100
		totalSize := 0

		for i := 0; i < numItems; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			value := []byte(fmt.Sprintf("value_%d", i))
			totalSize += len(key) + len(value)

			err := storage.Write(key, value)
			assert.NoError(t, err)
		}

		// Verify all items were stored
		for i := 0; i < numItems; i++ {
			key := []byte(fmt.Sprintf("key_%d", i))
			exists, err := storage.Has(key)
			assert.NoError(t, err)
			assert.True(t, exists)
		}

		t.Logf("Stored %d items with total approximate size: %d bytes", numItems, totalSize)
	})
}
