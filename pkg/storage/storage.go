//go:build db
// +build db

package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gochain/gochain/pkg/block"
)

// Storage represents the blockchain storage layer
type Storage struct {
	mu    sync.RWMutex
	db    *badger.DB
	config *StorageConfig
}

// StorageConfig holds configuration for storage
type StorageConfig struct {
	DataDir string
	DBType  string
}

// DefaultStorageConfig returns the default storage configuration
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		DataDir: "./data",
		DBType:  "badger",
	}
}

// NewStorage creates a new storage instance
func NewStorage(config *StorageConfig) (*Storage, error) {
	// Open BadgerDB
	opts := badger.DefaultOptions(config.DataDir)
	opts.Logger = nil // Disable logging for now
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	storage := &Storage{
		db:     db,
		config: config,
	}
	
	return storage, nil
}

// StoreBlock stores a block in the database
func (s *Storage) StoreBlock(block *block.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Serialize block
	blockData, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}
	
	// Store block by hash
	blockHash := block.CalculateHash()
	blockKey := fmt.Sprintf("block:%x", blockHash)
	
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(blockKey), blockData)
	})
	if err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}
	
	// Store block by height
	heightKey := fmt.Sprintf("height:%d", block.Header.Height)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(heightKey), blockHash)
	})
	if err != nil {
		return fmt.Errorf("failed to store block height: %w", err)
	}
	
	// Store latest height
	latestHeightKey := "latest_height"
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(latestHeightKey), []byte(fmt.Sprintf("%d", block.Header.Height)))
	})
	if err != nil {
		return fmt.Errorf("failed to store latest height: %w", err)
	}
	
	return nil
}

// GetBlock retrieves a block by its hash
func (s *Storage) GetBlock(hash []byte) (*block.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	blockKey := fmt.Sprintf("block:%x", hash)
	
	var blockData []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(blockKey))
		if err != nil {
			return err
		}
		
		blockData, err = item.ValueCopy(nil)
		return err
	})
	
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("block not found")
		}
		return nil, fmt.Errorf("failed to retrieve block: %w", err)
	}
	
	var block block.Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	
	return &block, nil
}

// GetBlockByHeight retrieves a block by its height
func (s *Storage) GetBlockByHeight(height uint64) (*block.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	heightKey := fmt.Sprintf("height:%d", height)
	
	var blockHash []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(heightKey))
		if err != nil {
			return err
		}
		
		blockHash, err = item.ValueCopy(nil)
		return err
	})
	
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("block at height %d not found", height)
		}
		return nil, fmt.Errorf("failed to retrieve block height: %w", err)
	}
	
	return s.GetBlock(blockHash)
}

// GetLatestHeight returns the latest block height
func (s *Storage) GetLatestHeight() (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	latestHeightKey := "latest_height"
	
	var heightData []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(latestHeightKey))
		if err != nil {
			return err
		}
		
		heightData, err = item.ValueCopy(nil)
		return err
	})
	
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, nil // Genesis block
		}
		return 0, fmt.Errorf("failed to retrieve latest height: %w", err)
	}
	
	var height uint64
	if _, err := fmt.Sscanf(string(heightData), "%d", &height); err != nil {
		return 0, fmt.Errorf("failed to parse height: %w", err)
	}
	
	return height, nil
}

// StoreTransaction stores a transaction in the database
func (s *Storage) StoreTransaction(tx *block.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Serialize transaction
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}
	
	// Store transaction by hash
	txKey := fmt.Sprintf("tx:%x", tx.Hash)
	
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(txKey), txData)
	})
	if err != nil {
		return fmt.Errorf("failed to store transaction: %w", err)
	}
	
	return nil
}

// GetTransaction retrieves a transaction by its hash
func (s *Storage) GetTransaction(hash []byte) (*block.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	txKey := fmt.Sprintf("tx:%x", hash)
	
	var txData []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(txKey))
		if err != nil {
			return err
		}
		
		txData, err = item.ValueCopy(nil)
		return err
	})
	
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}
	
	var tx block.Transaction
	if err := json.Unmarshal(txData, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}
	
	return &tx, nil
}

// StoreChainState stores the current chain state
func (s *Storage) StoreChainState(state *ChainState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Serialize state
	stateData, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal chain state: %w", err)
	}
	
	// Store state
	stateKey := "chain_state"
	
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(stateKey), stateData)
	})
	if err != nil {
		return fmt.Errorf("failed to store chain state: %w", err)
	}
	
	return nil
}

// GetChainState retrieves the current chain state
func (s *Storage) GetChainState() (*ChainState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stateKey := "chain_state"
	
	var stateData []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(stateKey))
		if err != nil {
			return err
		}
		
		stateData, err = item.ValueCopy(nil)
		return err
	})
	
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return &ChainState{
				BestBlockHash: []byte{},
				Height:        0,
				Difficulty:    1,
				LastUpdate:    time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve chain state: %w", err)
	}
	
	var state ChainState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chain state: %w", err)
	}
	
	return &state, nil
}

// ChainState represents the current state of the blockchain
type ChainState struct {
	BestBlockHash []byte    `json:"best_block_hash"`
	Height        uint64    `json:"height"`
	Difficulty    uint64    `json:"difficulty"`
	LastUpdate    time.Time `json:"last_update"`
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

// Compact performs database compaction
func (s *Storage) Compact() error {
	return s.db.RunValueLogGC(0.7)
}

// GetStats returns database statistics
func (s *Storage) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Get database size
	var size int64
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			size++
		}
		
		return nil
	})
	
	stats["total_keys"] = size
	stats["data_dir"] = s.config.DataDir
	stats["db_type"] = s.config.DBType
	
	return stats
}

// String returns a string representation of the storage
func (s *Storage) String() string {
	stats := s.GetStats()
	return fmt.Sprintf("Storage{Type: %s, Keys: %v, DataDir: %s}", 
		stats["db_type"], stats["total_keys"], stats["data_dir"])
} 