package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"encoding/hex" // Added import

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
	file, err := os.Create(filepath.Join(s.dataDir, b.HexHash()))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(b)
}

// GetBlock retrieves a block from a file.
func (s *Storage) GetBlock(hash []byte) (*block.Block, error) {
	file, err := os.Open(filepath.Join(s.dataDir, fmt.Sprintf("%x", hash)))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var b block.Block
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&b); err != nil {
		return nil, err
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
	file, err := os.Create(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(state)
}

// GetChainState retrieves the chain state from a file.
func (s *Storage) GetChainState() (*ChainState, error) {
	file, err := os.Open(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		if os.IsNotExist(err) {
			return &ChainState{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var state ChainState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

// Write writes a key-value pair to storage.
func (s *Storage) Write(key []byte, value []byte) error {
	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	return os.WriteFile(filename, value, 0644)
}

// Read reads a value from storage given a key.
func (s *Storage) Read(key []byte) ([]byte, error) {
	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Delete deletes a key-value pair from storage.
func (s *Storage) Delete(key []byte) error {
	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	return os.Remove(filename)
}

// Close is a no-op for file-based storage.
func (s *Storage) Close() error {
	return nil
}