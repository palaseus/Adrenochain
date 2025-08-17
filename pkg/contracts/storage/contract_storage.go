package storage

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/palaseus/adrenochain/pkg/storage"
)

// ContractStorageImpl implements the ContractStorage interface
type ContractStorageImpl struct {
	storage    storage.StorageInterface
	cache     map[string][]byte
	pending   map[string][]byte
	deleted   map[string]bool
	mu        sync.RWMutex
	committed bool
}

// NewContractStorage creates a new contract storage instance
func NewContractStorage(storage storage.StorageInterface) *ContractStorageImpl {
	return &ContractStorageImpl{
		storage:  storage,
		cache:   make(map[string][]byte),
		pending: make(map[string][]byte),
		deleted: make(map[string]bool),
	}
}

// Get retrieves a value from contract storage
func (cs *ContractStorageImpl) Get(address engine.Address, key engine.Hash) ([]byte, error) {
	storageKey := cs.makeStorageKey(address, key)
	
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Check pending changes first
	if value, exists := cs.pending[storageKey]; exists {
		if cs.deleted[storageKey] {
			return nil, nil
		}
		return value, nil
	}

	// Check cache
	if value, exists := cs.cache[storageKey]; exists {
		if cs.deleted[storageKey] {
			return nil, nil
		}
		return value, nil
	}

	// Get from persistent storage
	value, err := cs.storage.Read([]byte(storageKey))
	if err != nil {
		// Check if key doesn't exist
		exists, _ := cs.storage.Has([]byte(storageKey))
		if !exists {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: failed to get key %s: %v", engine.ErrStorageError, storageKey, err)
	}

	// Cache the value
	cs.cache[storageKey] = value
	return value, nil
}

// Set stores a value in contract storage
func (cs *ContractStorageImpl) Set(address engine.Address, key engine.Hash, value []byte) error {
	if cs.committed {
		return fmt.Errorf("%w: cannot modify committed storage", engine.ErrStorageError)
	}

	storageKey := cs.makeStorageKey(address, key)
	
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Remove from deleted set if it was marked for deletion
	delete(cs.deleted, storageKey)
	
	// Store in pending changes
	cs.pending[storageKey] = value
	
	// Update cache
	cs.cache[storageKey] = value

	return nil
}

// Delete removes a value from contract storage
func (cs *ContractStorageImpl) Delete(address engine.Address, key engine.Hash) error {
	if cs.committed {
		return fmt.Errorf("%w: cannot modify committed storage", engine.ErrStorageError)
	}

	storageKey := cs.makeStorageKey(address, key)
	
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Mark as deleted
	cs.deleted[storageKey] = true
	
	// Remove from pending and cache
	delete(cs.pending, storageKey)
	delete(cs.cache, storageKey)

	return nil
}

// GetStorageRoot returns the storage root hash for a contract
func (cs *ContractStorageImpl) GetStorageRoot(address engine.Address) (engine.Hash, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Create a temporary storage for this contract
	tempStorage := make(map[string][]byte)
	
	// This would typically iterate over all keys with the prefix
	// For now, we'll use a simplified approach
	// In production, you'd want to implement proper key iteration
	
	// Calculate root hash from all storage values
	rootHash := cs.calculateStorageRoot(tempStorage)
	return rootHash, nil
}

// Commit commits all pending storage changes
func (cs *ContractStorageImpl) Commit() error {
	if cs.committed {
		return fmt.Errorf("%w: storage already committed", engine.ErrStorageError)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Apply all pending changes to persistent storage
	for key, value := range cs.pending {
		if cs.deleted[key] {
			// Delete the key
			err := cs.storage.Delete([]byte(key))
			if err != nil {
				return fmt.Errorf("%w: failed to delete key %s: %v", engine.ErrStorageError, key, err)
			}
		} else {
			// Set the value
			err := cs.storage.Write([]byte(key), value)
			if err != nil {
				return fmt.Errorf("%w: failed to set key %s: %v", engine.ErrStorageError, key, err)
			}
		}
	}

	// Mark as committed
	cs.committed = true
	
	// Clear pending changes
	cs.pending = make(map[string][]byte)
	cs.deleted = make(map[string]bool)

	return nil
}

// Rollback rolls back all pending storage changes
func (cs *ContractStorageImpl) Rollback() error {
	if cs.committed {
		return fmt.Errorf("%w: cannot rollback committed storage", engine.ErrStorageError)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Clear all pending changes
	cs.pending = make(map[string][]byte)
	cs.deleted = make(map[string]bool)
	
	// Clear cache to ensure fresh reads
	cs.cache = make(map[string][]byte)

	return nil
}

// HasKey checks if a storage key exists
func (cs *ContractStorageImpl) HasKey(address engine.Address, key engine.Hash) bool {
	value, err := cs.Get(address, key)
	return err == nil && value != nil
}

// GetContractStorage returns all storage for a specific contract
func (cs *ContractStorageImpl) GetContractStorage(address engine.Address) (map[engine.Hash][]byte, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	contractStorage := make(map[engine.Hash][]byte)
	
	// This would typically iterate over all keys with the prefix
	// For now, we'll return an empty map
	// In production, implement proper key iteration and filtering
	
	return contractStorage, nil
}

// GetStorageSize returns the number of storage keys for a contract
func (cs *ContractStorageImpl) GetStorageSize(address engine.Address) (int, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// This would count all keys with the prefix
	// For now, return 0
	// In production, implement proper key counting
	
	return 0, nil
}

// ClearContractStorage removes all storage for a specific contract
func (cs *ContractStorageImpl) ClearContractStorage(address engine.Address) error {
	if cs.committed {
		return fmt.Errorf("%w: cannot modify committed storage", engine.ErrStorageError)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// This would iterate over all keys with the prefix and mark them for deletion
	// For now, we'll just clear the cache and pending changes
	// In production, implement proper key iteration and deletion
	
	// Clear cache entries for this contract
	for key := range cs.cache {
		if cs.hasAddressPrefix(key, cs.makeAddressPrefix(address)) {
			delete(cs.cache, key)
		}
	}
	
	// Clear pending changes for this contract
	for key := range cs.pending {
		if cs.hasAddressPrefix(key, cs.makeAddressPrefix(address)) {
			delete(cs.pending, key)
		}
	}

	return nil
}

// makeStorageKey creates a storage key from address and hash
func (cs *ContractStorageImpl) makeStorageKey(address engine.Address, key engine.Hash) string {
	// Combine address and key with a separator
	return fmt.Sprintf("%s:%s", address.String(), key.String())
}

// makeAddressPrefix creates a prefix for all storage keys of a contract
func (cs *ContractStorageImpl) makeAddressPrefix(address engine.Address) string {
	return fmt.Sprintf("%s:", address.String())
}

// hasAddressPrefix checks if a key has the given address prefix
func (cs *ContractStorageImpl) hasAddressPrefix(key, prefix string) bool {
	return len(key) >= len(prefix) && key[:len(prefix)] == prefix
}

// calculateStorageRoot calculates the root hash of contract storage
func (cs *ContractStorageImpl) calculateStorageRoot(storage map[string][]byte) engine.Hash {
	// This is a simplified implementation
	// In production, you'd want to implement a proper Merkle tree
	
	// For now, just hash all key-value pairs
	var data []byte
	for key, value := range storage {
		data = append(data, []byte(key)...)
		data = append(data, value...)
	}
	
	hash := sha256.Sum256(data)
	var result engine.Hash
	copy(result[:], hash[:])
	return result
}

// GetStorageProof generates a Merkle proof for a storage value
func (cs *ContractStorageImpl) GetStorageProof(address engine.Address, key engine.Hash) ([]byte, error) {
	// This would generate a Merkle proof for the storage value
	// For now, return a placeholder
	return []byte("storage_proof_placeholder"), nil
}

// VerifyStorageProof verifies a storage proof
func (cs *ContractStorageImpl) VerifyStorageProof(root engine.Hash, key engine.Hash, value []byte, proof []byte) bool {
	// This would verify the Merkle proof
	// For now, return true as placeholder
	return true
}
