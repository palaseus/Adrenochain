package consensus

import (
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

func TestNewConsensus(t *testing.T) {
	config := DefaultConsensusConfig()
	c := NewConsensus(config)

	assert.NotNil(t, c)
	assert.Equal(t, config, c.config)
	assert.Equal(t, config.MinDifficulty, c.difficulty)
}

func TestMineAndValidateBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	config.MinDifficulty = 1
	config.MaxDifficulty = 1
	c := NewConsensus(config)

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

	err = c.ValidateBlock(b, nil)
	assert.NoError(t, err)
}

func TestValidateInvalidBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	c := NewConsensus(config)

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
}
