package storage

import (
	"testing"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/contracts/engine"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// Mock storage implementation for testing
type MockStorage struct {
	data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

func (m *MockStorage) StoreBlock(b *block.Block) error {
	return nil
}

func (m *MockStorage) GetBlock(hash []byte) (*block.Block, error) {
	return nil, nil
}

func (m *MockStorage) StoreChainState(state *storage.ChainState) error {
	return nil
}

func (m *MockStorage) GetChainState() (*storage.ChainState, error) {
	return nil, nil
}

func (m *MockStorage) Read(key []byte) ([]byte, error) {
	if value, exists := m.data[string(key)]; exists {
		return value, nil
	}
	return nil, leveldb.ErrNotFound
}

func (m *MockStorage) Write(key []byte, value []byte) error {
	m.data[string(key)] = value
	return nil
}

func (m *MockStorage) Delete(key []byte) error {
	delete(m.data, string(key))
	return nil
}

func (m *MockStorage) Has(key []byte) (bool, error) {
	_, exists := m.data[string(key)]
	return exists, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestNewContractStorage(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	if cs == nil {
		t.Fatal("NewContractStorage returned nil")
	}

	if cs.storage != mockStorage {
		t.Error("Storage not set correctly")
	}

	if len(cs.cache) != 0 {
		t.Error("Cache should be empty initially")
	}

	if len(cs.pending) != 0 {
		t.Error("Pending should be empty initially")
	}

	if len(cs.deleted) != 0 {
		t.Error("Deleted should be empty initially")
	}

	if cs.committed {
		t.Error("Should not be committed initially")
	}
}

func TestContractStorage_Set(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Check if value is in pending
	storageKey := cs.makeStorageKey(address, key)
	if cs.pending[storageKey] == nil {
		t.Error("Value not added to pending")
	}

	// Check if value is in cache
	if cs.cache[storageKey] == nil {
		t.Error("Value not added to cache")
	}

	// Check if value is not marked as deleted
	if cs.deleted[storageKey] {
		t.Error("Value should not be marked as deleted")
	}
}

func TestContractStorage_Get(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Set a value first
	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the value
	retrievedValue, err := cs.Get(address, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrievedValue) != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", string(retrievedValue))
	}
}

func TestContractStorage_Get_FromPersistentStorage(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("persistent_value")

	// Write directly to persistent storage
	storageKey := cs.makeStorageKey(address, key)
	mockStorage.data[storageKey] = value

	// Get the value (should come from persistent storage)
	retrievedValue, err := cs.Get(address, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrievedValue) != "persistent_value" {
		t.Errorf("Expected 'persistent_value', got '%s'", string(retrievedValue))
	}

	// Check if value is now cached
	if cs.cache[storageKey] == nil {
		t.Error("Value should be cached after first read")
	}
}

func TestContractStorage_Delete(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Set a value first
	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete the value
	err = cs.Delete(address, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	storageKey := cs.makeStorageKey(address, key)

	// Check if value is marked as deleted
	if !cs.deleted[storageKey] {
		t.Error("Value should be marked as deleted")
	}

	// Check if value is removed from pending
	if cs.pending[storageKey] != nil {
		t.Error("Value should be removed from pending")
	}

	// Check if value is removed from cache
	if cs.cache[storageKey] != nil {
		t.Error("Value should be removed from cache")
	}
}

func TestContractStorage_Get_DeletedValue(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Set a value first
	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete the value
	err = cs.Delete(address, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to get the deleted value
	retrievedValue, err := cs.Get(address, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrievedValue != nil {
		t.Error("Deleted value should return nil")
	}
}

func TestContractStorage_Commit(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key1 := engine.Hash{10, 20, 30, 40, 50}
	key2 := engine.Hash{11, 21, 31, 41, 51}
	value1 := []byte("value1")
	value2 := []byte("value2")

	// Set values
	err := cs.Set(address, key1, value1)
	if err != nil {
		t.Fatalf("Set 1 failed: %v", err)
	}

	err = cs.Set(address, key2, value2)
	if err != nil {
		t.Fatalf("Set 2 failed: %v", err)
	}

	// Delete one value
	err = cs.Delete(address, key1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Commit changes
	err = cs.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Check if committed
	if !cs.committed {
		t.Error("Should be marked as committed")
	}

	// Check if pending is cleared
	if len(cs.pending) != 0 {
		t.Error("Pending should be cleared after commit")
	}

	// Check if deleted is cleared
	if len(cs.deleted) != 0 {
		t.Error("Deleted should be cleared after commit")
	}

	// Check if value2 was written to persistent storage
	storageKey2 := cs.makeStorageKey(address, key2)
	persistentValue, err := mockStorage.Read([]byte(storageKey2))
	if err != nil {
		t.Fatalf("Failed to read from persistent storage: %v", err)
	}

	if string(persistentValue) != "value2" {
		t.Errorf("Expected 'value2' in persistent storage, got '%s'", string(persistentValue))
	}
}

func TestContractStorage_Commit_AlreadyCommitted(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Commit once
	err := cs.Commit()
	if err != nil {
		t.Fatalf("First commit failed: %v", err)
	}

	// Try to commit again
	err = cs.Commit()
	if err == nil {
		t.Error("Second commit should fail")
	}
}

func TestContractStorage_Rollback(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Set a value
	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Rollback changes
	err = cs.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Check if pending is cleared
	if len(cs.pending) != 0 {
		t.Error("Pending should be cleared after rollback")
	}

	// Check if deleted is cleared
	if len(cs.deleted) != 0 {
		t.Error("Deleted should be cleared after rollback")
	}

	// Check if cache is cleared
	if len(cs.cache) != 0 {
		t.Error("Cache should be cleared after rollback")
	}
}

func TestContractStorage_Rollback_AlreadyCommitted(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Commit first
	err := cs.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Try to rollback
	err = cs.Rollback()
	if err == nil {
		t.Error("Rollback should fail after commit")
	}
}

func TestContractStorage_HasKey(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Initially key should not exist
	if cs.HasKey(address, key) {
		t.Error("Key should not exist initially")
	}

	// Set a value
	err := cs.Set(address, key, value)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Now key should exist
	if !cs.HasKey(address, key) {
		t.Error("Key should exist after setting")
	}

	// Delete the key
	err = cs.Delete(address, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Key should not exist after deletion
	if cs.HasKey(address, key) {
		t.Error("Key should not exist after deletion")
	}
}

func TestContractStorage_GetStorageRoot(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}

	// Get storage root
	root, err := cs.GetStorageRoot(address)
	if err != nil {
		t.Fatalf("GetStorageRoot failed: %v", err)
	}

	// Should return a hash (even if empty)
	if root == (engine.Hash{}) {
		t.Error("Storage root should not be empty hash")
	}
}

func TestContractStorage_GetContractStorage(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}

	// Get contract storage
	contractStorage, err := cs.GetContractStorage(address)
	if err != nil {
		t.Fatalf("GetContractStorage failed: %v", err)
	}

	// Should return a map (even if empty)
	if contractStorage == nil {
		t.Error("Contract storage should not be nil")
	}

	if len(contractStorage) != 0 {
		t.Error("Contract storage should be empty initially")
	}
}

func TestContractStorage_GetStorageSize(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}

	// Get storage size
	size, err := cs.GetStorageSize(address)
	if err != nil {
		t.Fatalf("GetStorageSize failed: %v", err)
	}

	// Should return 0 initially
	if size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}
}

func TestContractStorage_ClearContractStorage(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address1 := engine.Address{1, 2, 3, 4, 5}
	address2 := engine.Address{5, 4, 3, 2, 1}
	key1 := engine.Hash{10, 20, 30, 40, 50}
	key2 := engine.Hash{11, 21, 31, 41, 51}
	value1 := []byte("value1")
	value2 := []byte("value2")

	// Set values for both contracts
	err := cs.Set(address1, key1, value1)
	if err != nil {
		t.Fatalf("Set 1 failed: %v", err)
	}

	err = cs.Set(address2, key2, value2)
	if err != nil {
		t.Fatalf("Set 2 failed: %v", err)
	}

	// Clear storage for address1
	err = cs.ClearContractStorage(address1)
	if err != nil {
		t.Fatalf("ClearContractStorage failed: %v", err)
	}

	// Check if address1 storage is cleared
	if cs.HasKey(address1, key1) {
		t.Error("Address1 storage should be cleared")
	}

	// Check if address2 storage is still there
	if !cs.HasKey(address2, key2) {
		t.Error("Address2 storage should still exist")
	}
}

func TestContractStorage_ClearContractStorage_AlreadyCommitted(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Commit first
	err := cs.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	address := engine.Address{1, 2, 3, 4, 5}

	// Try to clear storage
	err = cs.ClearContractStorage(address)
	if err == nil {
		t.Error("ClearContractStorage should fail after commit")
	}
}

func TestContractStorage_GetStorageProof(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}

	// Get storage proof
	proof, err := cs.GetStorageProof(address, key)
	if err != nil {
		t.Fatalf("GetStorageProof failed: %v", err)
	}

	// Should return a placeholder proof
	if string(proof) != "storage_proof_placeholder" {
		t.Errorf("Expected 'storage_proof_placeholder', got '%s'", string(proof))
	}
}

func TestContractStorage_VerifyStorageProof(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	root := engine.Hash{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")
	proof := []byte("test_proof")

	// Verify storage proof
	result := cs.VerifyStorageProof(root, key, value, proof)
	if !result {
		t.Error("VerifyStorageProof should return true (placeholder)")
	}
}

func TestContractStorage_Set_AfterCommit(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Commit first
	err := cs.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}
	value := []byte("test_value")

	// Try to set after commit
	err = cs.Set(address, key, value)
	if err == nil {
		t.Error("Set should fail after commit")
	}
}

func TestContractStorage_Delete_AfterCommit(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Commit first
	err := cs.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}

	// Try to delete after commit
	err = cs.Delete(address, key)
	if err == nil {
		t.Error("Delete should fail after commit")
	}
}

func TestContractStorage_MakeStorageKey(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}
	key := engine.Hash{10, 20, 30, 40, 50}

	storageKey := cs.makeStorageKey(address, key)
	
	// The address and hash are 20 and 32 bytes respectively, so they'll have many zeros
	// Just check that the key contains the expected format
	if len(storageKey) == 0 {
		t.Error("Storage key should not be empty")
	}
	
	// Check that it contains a colon separator somewhere in the middle
	if len(storageKey) < 2 {
		t.Error("Storage key should have reasonable length")
	}
	
	// The format should be "address:key", so there should be a colon
	hasColon := false
	for i := 0; i < len(storageKey); i++ {
		if storageKey[i] == ':' {
			hasColon = true
			break
		}
	}
	
	if !hasColon {
		t.Error("Storage key should contain a colon separator")
	}
}

func TestContractStorage_MakeAddressPrefix(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	address := engine.Address{1, 2, 3, 4, 5}

	prefix := cs.makeAddressPrefix(address)
	
	// The address is 20 bytes, so it'll have many zeros
	// Just check that the prefix contains the expected format
	if len(prefix) == 0 {
		t.Error("Address prefix should not be empty")
	}
	
	// Check that it ends with a colon separator
	if prefix[len(prefix)-1] != ':' {
		t.Error("Address prefix should end with colon")
	}
}

func TestContractStorage_HasAddressPrefix(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	key := "0102030405:0a141e2832"
	prefix := "0102030405:"

	// Should have prefix
	if !cs.hasAddressPrefix(key, prefix) {
		t.Error("Key should have the address prefix")
	}

	// Should not have different prefix
	differentPrefix := "9999999999:"
	if cs.hasAddressPrefix(key, differentPrefix) {
		t.Error("Key should not have the different prefix")
	}

	// Should not have longer prefix
	longerPrefix := "0102030405:0a141e2832:extra"
	if cs.hasAddressPrefix(key, longerPrefix) {
		t.Error("Key should not have the longer prefix")
	}
}

func TestContractStorage_CalculateStorageRoot(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	storage := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	root := cs.calculateStorageRoot(storage)

	// Should return a hash
	if root == (engine.Hash{}) {
		t.Error("Storage root should not be empty hash")
	}
}

func TestContractStorage_Concurrency(t *testing.T) {
	mockStorage := NewMockStorage()
	cs := NewContractStorage(mockStorage)

	// Test concurrent operations
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			address := engine.Address{byte(id), 2, 3, 4, 5}
			key := engine.Hash{byte(id), 20, 30, 40, 50}
			value := []byte("value")

			// Set value
			err := cs.Set(address, key, value)
			if err != nil {
				t.Errorf("Concurrent Set %d failed: %v", id, err)
			}

			// Get value
			_, err = cs.Get(address, key)
			if err != nil {
				t.Errorf("Concurrent Get %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all values were set
	for i := 0; i < numGoroutines; i++ {
		address := engine.Address{byte(i), 2, 3, 4, 5}
		key := engine.Hash{byte(i), 20, 30, 40, 50}

		if !cs.HasKey(address, key) {
			t.Errorf("Key for goroutine %d should exist", i)
		}
	}
}
