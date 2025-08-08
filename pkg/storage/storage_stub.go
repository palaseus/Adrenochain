//go:build !db
// +build !db

package storage

import (
	"fmt"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// Storage is a no-op stub when built without the 'db' tag.
type Storage struct{}

type StorageConfig struct {
	DataDir string
	DBType  string
}

func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{DataDir: "./data", DBType: "badger"}
}

func NewStorage(config *StorageConfig) (*Storage, error) { return &Storage{}, nil }

func (s *Storage) StoreBlock(b *block.Block) error                    { return nil }
func (s *Storage) GetBlock(hash []byte) (*block.Block, error)         { return nil, fmt.Errorf("not implemented without db tag") }
func (s *Storage) GetBlockByHeight(height uint64) (*block.Block, error) {
	return nil, fmt.Errorf("not implemented without db tag")
}
func (s *Storage) GetLatestHeight() (uint64, error) { return 0, nil }
func (s *Storage) StoreTransaction(tx *block.Transaction) error { return nil }
func (s *Storage) GetTransaction(hash []byte) (*block.Transaction, error) {
	return nil, fmt.Errorf("not implemented without db tag")
}

type ChainState struct {
	BestBlockHash []byte    `json:"best_block_hash"`
	Height        uint64    `json:"height"`
	Difficulty    uint64    `json:"difficulty"`
	LastUpdate    time.Time `json:"last_update"`
}

func (s *Storage) StoreChainState(state *ChainState) error { return nil }
func (s *Storage) GetChainState() (*ChainState, error) {
	return &ChainState{BestBlockHash: []byte{}, Height: 0, Difficulty: 1, LastUpdate: time.Now()}, nil
}
func (s *Storage) Close() error                    { return nil }
func (s *Storage) Compact() error                  { return nil }
func (s *Storage) GetStats() map[string]interface{} { return map[string]interface{}{"db_type": "stub", "total_keys": 0, "data_dir": ""} }
func (s *Storage) String() string                  { return "Storage{Type: stub}" }
