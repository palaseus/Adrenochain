package storage

import (
	"os"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	dataDir := "./test_data"
	defer os.RemoveAll(dataDir)

	config := &StorageConfig{DataDir: dataDir}
	s, err := NewStorage(config)
	assert.NoError(t, err)
	defer s.Close()

	// Test StoreBlock and GetBlock
	b := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now(),
			Difficulty:    1,
			Height:        1,
		},
	}
	b.Header.MerkleRoot = b.CalculateMerkleRoot()

	err = s.StoreBlock(b)
	assert.NoError(t, err)

	retrievedBlock, err := s.GetBlock(b.CalculateHash())
	assert.NoError(t, err)
	assert.Equal(t, b.HexHash(), retrievedBlock.HexHash())

	// Test StoreChainState and GetChainState
	state := &ChainState{
		BestBlockHash: b.CalculateHash(),
		Height:        1,
	}

	err = s.StoreChainState(state)
	assert.NoError(t, err)

	retrievedState, err := s.GetChainState()
	assert.NoError(t, err)
	assert.Equal(t, state.BestBlockHash, retrievedState.BestBlockHash)
	assert.Equal(t, state.Height, retrievedState.Height)
}
