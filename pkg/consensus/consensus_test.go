package consensus

import (
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

type mockChainReader struct {
	blocks map[uint64]*block.Block
}

func (m *mockChainReader) GetHeight() uint64 {
	return 0
}

func (m *mockChainReader) GetBlockByHeight(height uint64) *block.Block {
	return m.blocks[height]
}

func (m *mockChainReader) GetBlock(hash []byte) *block.Block {
	for _, b := range m.blocks {
		if assert.ObjectsAreEqual(b.CalculateHash(), hash) {
			return b
		}
	}
	return nil
}

func TestNewConsensus(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &mockChainReader{blocks: make(map[uint64]*block.Block)}
	c := NewConsensus(config, mockChain)

	assert.NotNil(t, c)
	assert.Equal(t, config, c.config)
	assert.Equal(t, config.MinDifficulty, c.difficulty)
}

func TestMineAndValidateBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	config.MinDifficulty = 1
	config.MaxDifficulty = 1
	mockChain := &mockChainReader{blocks: make(map[uint64]*block.Block)}
	c := NewConsensus(config, mockChain)

	b := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now(),
			Difficulty:    c.GetDifficulty(),
		},
	}
	b.Header.MerkleRoot = b.CalculateMerkleRoot()

	err := c.MineBlock(b, nil)
	assert.NoError(t, err)

	assert.True(t, c.ValidateProofOfWork(b))

	// For ValidateBlock, we need a prevBlock. Let's create a dummy one.
	prevBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now().Add(-1 * time.Minute),
			Difficulty:    c.GetDifficulty(),
			Height:        0,
		},
	}
	prevBlock.Header.MerkleRoot = prevBlock.CalculateMerkleRoot()

	err = c.ValidateBlock(b, prevBlock)
	assert.NoError(t, err)
}

func TestValidateInvalidBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &mockChainReader{blocks: make(map[uint64]*block.Block)}
	c := NewConsensus(config, mockChain)

	// Block with invalid version
	b := &block.Block{
		Header: &block.Header{
			Version:       0,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now(),
			Difficulty:    c.GetDifficulty(),
		},
	}
	b.Header.MerkleRoot = b.CalculateMerkleRoot()

	err := c.ValidateBlock(b, nil)
	assert.Error(t, err)

	// Test with a block that has a difficulty mismatch
	prevBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now().Add(-1 * time.Minute),
			Difficulty:    c.GetDifficulty(),
			Height:        0,
		},
	}
	prevBlock.Header.MerkleRoot = prevBlock.CalculateMerkleRoot()
	mockChain.blocks[0] = prevBlock // Add prevBlock to mock chain

	b2 := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte{},
			Timestamp:     time.Now(),
			Difficulty:    c.GetDifficulty() + 1, // Incorrect difficulty
			Height:        1,
		},
	}
	b2.Header.MerkleRoot = b2.CalculateMerkleRoot()

	err = c.ValidateBlock(b2, prevBlock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match expected")
}
