package storage

import (
	"github.com/gochain/gochain/pkg/contracts/engine"
)

// ContractStorage defines the interface for contract state storage
type ContractStorage interface {
	// Get retrieves a value from contract storage
	Get(address engine.Address, key engine.Hash) ([]byte, error)
	
	// Set stores a value in contract storage
	Set(address engine.Address, key engine.Hash, value []byte) error
	
	// Delete removes a value from contract storage
	Delete(address engine.Address, key engine.Hash) error
	
	// GetStorageRoot returns the storage root hash for a contract
	GetStorageRoot(address engine.Address) (engine.Hash, error)
	
	// Commit commits all pending storage changes
	Commit() error
	
	// Rollback rolls back all pending storage changes
	Rollback() error
	
	// HasKey checks if a storage key exists
	HasKey(address engine.Address, key engine.Hash) bool
	
	// GetContractStorage returns all storage for a specific contract
	GetContractStorage(address engine.Address) (map[engine.Hash][]byte, error)
	
	// GetStorageSize returns the number of storage keys for a contract
	GetStorageSize(address engine.Address) (int, error)
	
	// ClearContractStorage removes all storage for a specific contract
	ClearContractStorage(address engine.Address) error
	
	// GetStorageProof generates a Merkle proof for a storage value
	GetStorageProof(address engine.Address, key engine.Hash) ([]byte, error)
	
	// VerifyStorageProof verifies a storage proof
	VerifyStorageProof(root engine.Hash, key engine.Hash, value []byte, proof []byte) bool
}
