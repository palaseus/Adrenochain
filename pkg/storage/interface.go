package storage

import (
	"github.com/gochain/gochain/pkg/block"
)

// StorageInterface defines the common interface for all storage implementations
type StorageInterface interface {
	// Block operations
	StoreBlock(b *block.Block) error
	GetBlock(hash []byte) (*block.Block, error)
	
	// Chain state operations
	StoreChainState(state *ChainState) error
	GetChainState() (*ChainState, error)
	
	// Key-value operations
	Write(key []byte, value []byte) error
	Read(key []byte) ([]byte, error)
	Delete(key []byte) error
	Has(key []byte) (bool, error)
	
	// Utility operations
	Close() error
}

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeFile   StorageType = "file"
	StorageTypeLevelDB StorageType = "leveldb"
)

// StorageFactory creates storage instances based on configuration
type StorageFactory struct{}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{}
}

// CreateStorage creates a storage instance based on the specified type
func (f *StorageFactory) CreateStorage(storageType StorageType, dataDir string) (StorageInterface, error) {
	switch storageType {
	case StorageTypeLevelDB:
		config := DefaultLevelDBStorageConfig().WithDataDir(dataDir)
		return NewLevelDBStorage(config)
	case StorageTypeFile:
		fallthrough
	default:
		config := DefaultStorageConfig().WithDataDir(dataDir)
		return NewStorage(config)
	}
} 