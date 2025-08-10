package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStateTrie(t *testing.T) {
	trie := NewStateTrie()
	assert.NotNil(t, trie)
	assert.NotNil(t, trie.root)
	assert.Equal(t, NodeTypeBranch, trie.root.Type)
	assert.Equal(t, 0, len(trie.root.Children))
}

func TestStateTriePutAndGet(t *testing.T) {
	trie := NewStateTrie()

	// Test basic put and get
	key := []byte("test_key")
	value := []byte("test_value")

	err := trie.Put(key, value)
	assert.NoError(t, err)

	retrieved, err := trie.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// Test updating existing key
	newValue := []byte("new_value")
	err = trie.Put(key, newValue)
	assert.NoError(t, err)

	retrieved, err = trie.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, newValue, retrieved)
}

func TestStateTrieDelete(t *testing.T) {
	trie := NewStateTrie()

	// Add a key-value pair
	key := []byte("test_key")
	value := []byte("test_value")
	err := trie.Put(key, value)
	assert.NoError(t, err)

	// Verify it exists
	retrieved, err := trie.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// Delete it
	err = trie.Delete(key)
	assert.NoError(t, err)

	// Verify it's gone
	retrieved, err = trie.Get(key)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestStateTrieMultipleKeys(t *testing.T) {
	trie := NewStateTrie()

	// Add multiple keys
	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}
	values := [][]byte{
		[]byte("value1"),
		[]byte("value2"),
		[]byte("value3"),
	}

	for i, key := range keys {
		err := trie.Put(key, values[i])
		assert.NoError(t, err)
	}

	// Verify all keys exist
	for i, key := range keys {
		retrieved, err := trie.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, values[i], retrieved)
	}

	// Test non-existent key
	nonExistent, err := trie.Get([]byte("non_existent"))
	assert.NoError(t, err)
	assert.Nil(t, nonExistent)
}

func TestStateTrieRootHash(t *testing.T) {
	trie := NewStateTrie()

	// Initially, root hash should be nil (empty trie)
	initialHash := trie.RootHash()
	assert.Nil(t, initialHash)

	// Add a key-value pair
	key := []byte("test_key")
	value := []byte("test_value")
	err := trie.Put(key, value)
	assert.NoError(t, err)

	// Root hash should now exist
	hash1 := trie.RootHash()
	assert.NotNil(t, hash1)

	// Update the value
	newValue := []byte("new_value")
	err = trie.Put(key, newValue)
	assert.NoError(t, err)

	// Root hash should change
	hash2 := trie.RootHash()
	assert.NotNil(t, hash2)
	assert.NotEqual(t, hash1, hash2)
}

func TestStateTrieCommit(t *testing.T) {
	trie := NewStateTrie()

	// Add some data
	key1 := []byte("key1")
	value1 := []byte("value1")
	key2 := []byte("key2")
	value2 := []byte("value2")

	err := trie.Put(key1, value1)
	assert.NoError(t, err)
	err = trie.Put(key2, value2)
	assert.NoError(t, err)

	// Get root hash before commit
	beforeHash := trie.RootHash()

	// Commit changes
	commitHash := trie.Commit()

	// Hash should remain the same after commit
	assert.Equal(t, beforeHash, commitHash)

	// Verify data still exists
	retrieved1, err := trie.Get(key1)
	assert.NoError(t, err)
	assert.Equal(t, value1, retrieved1)

	retrieved2, err := trie.Get(key2)
	assert.NoError(t, err)
	assert.Equal(t, value2, retrieved2)
}

func TestStateTrieGetProof(t *testing.T) {
	trie := NewStateTrie()

	// Add a key-value pair
	key := []byte("test_key")
	value := []byte("test_value")
	err := trie.Put(key, value)
	assert.NoError(t, err)

	// Get proof
	proof, err := trie.GetProof(key)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Greater(t, len(proof), 0)

	// Verify proof
	rootHash := trie.RootHash()
	isValid := trie.VerifyProof(rootHash, key, value, proof)
	assert.True(t, isValid)
}

func TestStateTrieStats(t *testing.T) {
	trie := NewStateTrie()

	// Initially empty
	stats := trie.GetStats()
	assert.Equal(t, "0", stats["root_hash"])
	assert.Equal(t, 0, stats["dirty_nodes"])
	assert.Equal(t, 1, stats["total_nodes"]) // Root node
	assert.Equal(t, 0, stats["leaf_nodes"])

	// Add some data
	key1 := []byte("key1")
	value1 := []byte("value1")
	key2 := []byte("key2")
	value2 := []byte("value2")

	err := trie.Put(key1, value1)
	assert.NoError(t, err)
	err = trie.Put(key2, value2)
	assert.NoError(t, err)

	// Check stats after adding data
	stats = trie.GetStats()
	assert.NotEqual(t, "0", stats["root_hash"])
	assert.Greater(t, stats["dirty_nodes"], 0)
	assert.Greater(t, stats["total_nodes"], 1)
	assert.Greater(t, stats["leaf_nodes"], 0)
}

func TestStateTrieConcurrentAccess(t *testing.T) {
	trie := NewStateTrie()
	done := make(chan bool)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := []byte("key_" + string(rune(id+'0')))
			value := []byte("value_" + string(rune(id+'0')))

			err := trie.Put(key, value)
			assert.NoError(t, err)

			retrieved, err := trie.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, value, retrieved)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all data was stored
	for i := 0; i < 10; i++ {
		key := []byte("key_" + string(rune(i+'0')))
		expectedValue := []byte("value_" + string(rune(i+'0')))

		retrieved, err := trie.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, retrieved)
	}
}

func TestStateTrieNilKey(t *testing.T) {
	trie := NewStateTrie()

	// Test nil key
	err := trie.Put(nil, []byte("value"))
	assert.Error(t, err)

	_, err = trie.Get(nil)
	assert.Error(t, err)

	err = trie.Delete(nil)
	assert.Error(t, err)
}

func TestStateTrieEmptyKey(t *testing.T) {
	trie := NewStateTrie()

	// Test empty key
	err := trie.Put([]byte{}, []byte("value"))
	assert.Error(t, err)

	_, err = trie.Get([]byte{})
	assert.Error(t, err)

	err = trie.Delete([]byte{})
	assert.Error(t, err)
}

func TestStateTrieNilValue(t *testing.T) {
	trie := NewStateTrie()

	// Test nil value
	key := []byte("test_key")
	err := trie.Put(key, nil)
	assert.Error(t, err)
}

func TestStateTrieLargeData(t *testing.T) {
	trie := NewStateTrie()

	// Test with large data
	largeKey := make([]byte, 1000)
	largeValue := make([]byte, 10000)

	for i := range largeKey {
		largeKey[i] = byte(i % 256)
	}
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	err := trie.Put(largeKey, largeValue)
	assert.NoError(t, err)

	retrieved, err := trie.Get(largeKey)
	assert.NoError(t, err)
	assert.Equal(t, largeValue, retrieved)
}
