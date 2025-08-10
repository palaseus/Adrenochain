package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gochain/gochain/pkg/block"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBStorage implements persistent storage using LevelDB
type LevelDBStorage struct {
	db      *leveldb.DB
	dataDir string
}

// LevelDBStorageConfig holds configuration for LevelDB storage
type LevelDBStorageConfig struct {
	DataDir string
	// LevelDB specific options
	WriteBufferSize        int
	OpenFilesCacheCapacity int
	Compression            bool
}

// DefaultLevelDBStorageConfig returns the default LevelDB storage configuration
func DefaultLevelDBStorageConfig() *LevelDBStorageConfig {
	return &LevelDBStorageConfig{
		DataDir:                "./data/leveldb",
		WriteBufferSize:        64 * 1024 * 1024, // 64MB
		OpenFilesCacheCapacity: 1000,
		Compression:            true,
	}
}

// WithDataDir sets the data directory for the LevelDB storage config
func (c *LevelDBStorageConfig) WithDataDir(dataDir string) *LevelDBStorageConfig {
	c.DataDir = dataDir
	return c
}

// WithWriteBufferSize sets the write buffer size for LevelDB
func (c *LevelDBStorageConfig) WithWriteBufferSize(size int) *LevelDBStorageConfig {
	c.WriteBufferSize = size
	return c
}

// WithOpenFilesCacheCapacity sets the open files cache capacity for LevelDB
func (c *LevelDBStorageConfig) WithOpenFilesCacheCapacity(capacity int) *LevelDBStorageConfig {
	c.OpenFilesCacheCapacity = capacity
	return c
}

// WithCompression enables or disables compression
func (c *LevelDBStorageConfig) WithCompression(enable bool) *LevelDBStorageConfig {
	c.Compression = enable
	return c
}

// NewLevelDBStorage creates a new LevelDB-based storage
func NewLevelDBStorage(config *LevelDBStorageConfig) (*LevelDBStorage, error) {
	// Create data directory if it doesn't exist
	if err := ensureDir(config.DataDir); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Configure LevelDB options
	options := &opt.Options{
		WriteBuffer:            config.WriteBufferSize,
		OpenFilesCacheCapacity: config.OpenFilesCacheCapacity,
		Compression:            opt.SnappyCompression,
		WriteL0PauseTrigger:    12,
		WriteL0SlowdownTrigger: 8,
	}

	if !config.Compression {
		options.Compression = opt.NoCompression
	}

	// Open LevelDB database
	db, err := leveldb.OpenFile(config.DataDir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}

	return &LevelDBStorage{
		db:      db,
		dataDir: config.DataDir,
	}, nil
}

// StoreBlock stores a block in LevelDB
func (s *LevelDBStorage) StoreBlock(b *block.Block) error {
	if b == nil {
		return fmt.Errorf("cannot store nil block")
	}

	// Serialize block to JSON
	data, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Store with key prefix for blocks
	key := makeBlockKey(b.CalculateHash())
	return s.db.Put(key, data, nil)
}

// GetBlock retrieves a block from LevelDB
func (s *LevelDBStorage) GetBlock(hash []byte) (*block.Block, error) {
	if hash == nil || len(hash) == 0 {
		return nil, fmt.Errorf("invalid hash: cannot be nil or empty")
	}

	key := makeBlockKey(hash)
	data, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("block not found: %x", hash)
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	var b block.Block
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &b, nil
}

// StoreChainState stores the chain state in LevelDB
func (s *LevelDBStorage) StoreChainState(state *ChainState) error {
	if state == nil {
		return fmt.Errorf("cannot store nil chain state")
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal chain state: %w", err)
	}

	key := []byte("chainstate")
	return s.db.Put(key, data, nil)
}

// GetChainState retrieves the chain state from LevelDB
func (s *LevelDBStorage) GetChainState() (*ChainState, error) {
	key := []byte("chainstate")
	data, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return &ChainState{}, nil
		}
		return nil, fmt.Errorf("failed to get chain state: %w", err)
	}

	var state ChainState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chain state: %w", err)
	}

	return &state, nil
}

// Write writes a key-value pair to LevelDB
func (s *LevelDBStorage) Write(key []byte, value []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}
	if value == nil {
		return fmt.Errorf("invalid value: cannot be nil")
	}

	return s.db.Put(key, value, nil)
}

// Read reads a value from LevelDB given a key
func (s *LevelDBStorage) Read(key []byte) ([]byte, error) {
	if key == nil || len(key) == 0 {
		return nil, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	data, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read key-value pair: %w", err)
	}

	return data, nil
}

// Delete deletes a key-value pair from LevelDB
func (s *LevelDBStorage) Delete(key []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}

	return s.db.Delete(key, nil)
}

// Has checks if a key exists in LevelDB
func (s *LevelDBStorage) Has(key []byte) (bool, error) {
	if key == nil || len(key) == 0 {
		return false, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	return s.db.Has(key, nil)
}

// Close closes the LevelDB connection
func (s *LevelDBStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Compact compacts the LevelDB database to reclaim space
func (s *LevelDBStorage) Compact() error {
	if s.db != nil {
		// Compact the entire database
		return s.db.CompactRange(util.Range{Start: nil, Limit: nil})
	}
	return nil
}

// GetStats returns LevelDB statistics
func (s *LevelDBStorage) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if s.db != nil {
		// Get basic database info
		stats["data_dir"] = s.dataDir

		// Note: LevelDB doesn't expose many metrics by default
		// In a production system, you might want to use prometheus or similar
		stats["db_open"] = true
	}

	return stats
}

// makeBlockKey creates a key for storing blocks with a prefix
func makeBlockKey(hash []byte) []byte {
	prefix := []byte("block:")
	key := make([]byte, len(prefix)+len(hash))
	copy(key, prefix)
	copy(key[len(prefix):], hash)
	return key
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
