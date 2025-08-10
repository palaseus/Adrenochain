package storage

import (
	"encoding/hex" // Added import
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gochain/gochain/pkg/block"
)

// Storage implements a file-based storage for blocks and chain state.
type Storage struct {
	dataDir string
}

// StorageConfig holds configuration for storage.
type StorageConfig struct {
	DataDir string
}

// DefaultStorageConfig returns the default storage configuration.
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{DataDir: "./data"}
}

// WithDataDir sets the data directory for the storage config.
func (c *StorageConfig) WithDataDir(dataDir string) *StorageConfig {
	c.DataDir = dataDir
	return c
}

// NewStorage creates a new file-based storage.
func NewStorage(config *StorageConfig) (*Storage, error) {
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: config.DataDir}, nil
}

// StoreBlock stores a block to a file.
func (s *Storage) StoreBlock(b *block.Block) error {
	if b == nil {
		return fmt.Errorf("cannot store nil block")
	}

	file, err := os.Create(filepath.Join(s.dataDir, b.HexHash()))
	if err != nil {
		return fmt.Errorf("failed to create block file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(b); err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}
	return nil
}

// GetBlock retrieves a block from a file.
func (s *Storage) GetBlock(hash []byte) (*block.Block, error) {
	if hash == nil || len(hash) == 0 {
		return nil, fmt.Errorf("invalid hash: cannot be nil or empty")
	}

	file, err := os.Open(filepath.Join(s.dataDir, fmt.Sprintf("%x", hash)))
	if err != nil {
		return nil, fmt.Errorf("failed to open block file: %w", err)
	}
	defer file.Close()

	var b block.Block
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&b); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}
	return &b, nil
}

// ChainState represents the state of the blockchain.
type ChainState struct {
	BestBlockHash []byte `json:"best_block_hash"`
	Height        uint64 `json:"height"`
}

// StoreChainState stores the chain state to a file.
func (s *Storage) StoreChainState(state *ChainState) error {
	if state == nil {
		return fmt.Errorf("cannot store nil chain state")
	}

	file, err := os.Create(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		return fmt.Errorf("failed to create chain state file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(state); err != nil {
		return fmt.Errorf("failed to encode chain state: %w", err)
	}
	return nil
}

// GetChainState retrieves the chain state from a file.
func (s *Storage) GetChainState() (*ChainState, error) {
	file, err := os.Open(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		if os.IsNotExist(err) {
			return &ChainState{}, nil
		}
		return nil, fmt.Errorf("failed to open chain state file: %w", err)
	}
	defer file.Close()

	var state ChainState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode chain state: %w", err)
	}
	return &state, nil
}

// Write writes a key-value pair to storage.
func (s *Storage) Write(key []byte, value []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}
	if value == nil {
		return fmt.Errorf("invalid value: cannot be nil")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	if err := os.WriteFile(filename, value, 0644); err != nil {
		return fmt.Errorf("failed to write key-value pair: %w", err)
	}
	return nil
}

// Read reads a value from storage given a key.
func (s *Storage) Read(key []byte) ([]byte, error) {
	if key == nil || len(key) == 0 {
		return nil, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err // Return the original os.IsNotExist error
		}
		return nil, fmt.Errorf("failed to read key-value pair: %w", err)
	}
	return data, nil
}

// Delete deletes a key-value pair from storage.
func (s *Storage) Delete(key []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete key-value pair: %w", err)
	}
	return nil
}

// Has checks if a key exists in storage.
func (s *Storage) Has(key []byte) (bool, error) {
	if key == nil || len(key) == 0 {
		return false, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if key exists: %w", err)
	}
	return true, nil
}

// Close is a no-op for file-based storage.
func (s *Storage) Close() error {
	return nil
}
