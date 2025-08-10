package consensus

import (
	"math/big"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// MockChainReader implements ChainReader for testing
type MockChainReader struct {
	blocks map[uint64]*block.Block
	height uint64
}

func (m *MockChainReader) GetHeight() uint64 {
	return m.height
}

func (m *MockChainReader) GetBlockByHeight(height uint64) *block.Block {
	return m.blocks[height]
}

func (m *MockChainReader) GetBlock(hash []byte) *block.Block {
	// Not used in these tests
	return nil
}

func (m *MockChainReader) GetAccumulatedDifficulty(height uint64) (*big.Int, error) {
	// Mock implementation for testing
	return nil, nil
}

func TestNewConsensus(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &MockChainReader{height: 0}

	consensus := NewConsensus(config, mockChain)

	if consensus == nil {
		t.Fatal("NewConsensus returned nil")
	}

	if consensus.config != config {
		t.Error("consensus config not set correctly")
	}

	if consensus.chain != mockChain {
		t.Error("consensus chain not set correctly")
	}

	if consensus.finalityDepth != config.FinalityDepth {
		t.Errorf("expected finality depth %d, got %d", config.FinalityDepth, consensus.finalityDepth)
	}
}

func TestIsBlockFinal(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &MockChainReader{height: 200}
	consensus := NewConsensus(config, mockChain)

	// Block at height 100 should be final (200 - 100 >= 100)
	if !consensus.IsBlockFinal(100) {
		t.Error("block at height 100 should be final")
	}

	// Block at height 150 should not be final (200 - 150 < 100)
	if consensus.IsBlockFinal(150) {
		t.Error("block at height 150 should not be final")
	}
}

func TestAddAndValidateCheckpoint(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &MockChainReader{height: 0}
	consensus := NewConsensus(config, mockChain)

	// Add a checkpoint
	height := uint64(10000)
	hash := []byte("test_hash_12345")
	consensus.AddCheckpoint(height, hash)

	// Validate the checkpoint
	if !consensus.ValidateCheckpoint(height, hash) {
		t.Error("checkpoint validation failed for correct hash")
	}

	// Validate with wrong hash
	if consensus.ValidateCheckpoint(height, []byte("wrong_hash")) {
		t.Error("checkpoint validation should fail for wrong hash")
	}

	// Validate non-existent checkpoint
	if consensus.ValidateCheckpoint(9999, []byte("any_hash")) {
		t.Error("non-existent checkpoint should fail validation")
	}
}

func TestGetAccumulatedDifficulty(t *testing.T) {
	config := DefaultConsensusConfig()

	// Create mock chain with some blocks
	mockChain := &MockChainReader{
		blocks: make(map[uint64]*block.Block),
		height: 3,
	}

	// Add blocks with different difficulties
	for i := uint64(1); i <= 3; i++ {
		mockChain.blocks[i] = &block.Block{
			Header: &block.Header{
				Height:     i,
				Difficulty: i * 10, // 10, 20, 30
			},
		}
	}

	consensus := NewConsensus(config, mockChain)

	// Test accumulated difficulty calculation
	diff, err := consensus.GetAccumulatedDifficulty(3)
	if err != nil {
		t.Fatalf("failed to calculate accumulated difficulty: %v", err)
	}

	expected := int64(10 + 20 + 30) // 60
	if diff.Int64() != expected {
		t.Errorf("expected accumulated difficulty %d, got %d", expected, diff.Int64())
	}
}

func TestValidateBlockWithCheckpoint(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &MockChainReader{height: 0}
	// Add a genesis block to the mock chain
	genesisBlock := block.NewBlock(make([]byte, 32), 0, 1)
	genesisBlock.Header.Timestamp = time.Now().Add(-time.Hour) // Genesis block in the past
	mockChain.blocks = map[uint64]*block.Block{0: genesisBlock}

	consensus := NewConsensus(config, mockChain)

	// Create a block with all required fields
	testBlock := block.NewBlock(make([]byte, 32), 1, 1) // Use difficulty 1 to match genesis
	testBlock.Header.Timestamp = time.Now()
	testBlock.Header.Nonce = 12345

	// Mock the difficulty calculation
	consensus.difficulty = 10

	// Mine the block to get valid proof of work
	stopChan := make(chan struct{})
	err := consensus.MineBlock(testBlock, stopChan)
	if err != nil {
		t.Fatalf("failed to mine block: %v", err)
	}

	// Test validation without checkpoint (should pass)
	if err := consensus.ValidateBlock(testBlock, nil); err != nil {
		t.Errorf("block validation should pass without checkpoint: %v", err)
	}

	// Add a checkpoint for this height
	hash := testBlock.CalculateHash()
	consensus.AddCheckpoint(1, hash)

	// Test validation with correct checkpoint
	if err := consensus.ValidateBlock(testBlock, nil); err != nil {
		t.Errorf("block validation should pass with correct checkpoint: %v", err)
	}

	// Test validation with wrong checkpoint
	consensus.AddCheckpoint(1, []byte("wrong_hash"))
	if err := consensus.ValidateBlock(testBlock, nil); err == nil {
		t.Error("block validation should fail with wrong checkpoint")
	}
}

func TestMineAndValidateBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	config.MinDifficulty = 1
	config.MaxDifficulty = 10

	mockChain := &MockChainReader{height: 0}
	consensus := NewConsensus(config, mockChain)

	// Create a block to mine with all required fields
	testBlock := block.NewBlock(make([]byte, 32), 1, 1)
	testBlock.Header.Timestamp = time.Now()
	testBlock.Header.Nonce = 0

	// Mine the block
	stopChan := make(chan struct{})
	err := consensus.MineBlock(testBlock, stopChan)
	if err != nil {
		t.Fatalf("failed to mine block: %v", err)
	}

	// Validate the mined block
	if !consensus.ValidateProofOfWork(testBlock) {
		t.Error("mined block should have valid proof of work")
	}
}

func TestValidateInvalidBlock(t *testing.T) {
	config := DefaultConsensusConfig()
	mockChain := &MockChainReader{height: 0}
	consensus := NewConsensus(config, mockChain)

	// Test nil block
	if err := consensus.ValidateBlock(nil, nil); err == nil {
		t.Error("validation should fail for nil block")
	}

	// Test block with nil header
	testBlock := block.NewBlock(make([]byte, 32), 1, 1)
	testBlock.Header = nil
	if err := consensus.ValidateBlock(testBlock, nil); err == nil {
		t.Error("validation should fail for block with nil header")
	}

	// Test block with invalid header (this will cause IsValid to fail)
	invalidBlock := block.NewBlock(make([]byte, 32), 1, 1)
	invalidBlock.Header.Version = 0             // Invalid version
	invalidBlock.Header.Timestamp = time.Time{} // Zero time
	if err := consensus.ValidateBlock(invalidBlock, nil); err == nil {
		t.Error("validation should fail for block with invalid header")
	}
}

func TestConsensusConfigDefaults(t *testing.T) {
	config := DefaultConsensusConfig()

	if config.FinalityDepth != 100 {
		t.Errorf("expected finality depth 100, got %d", config.FinalityDepth)
	}

	if config.CheckpointInterval != 10000 {
		t.Errorf("expected checkpoint interval 10000, got %d", config.CheckpointInterval)
	}

	if config.TargetBlockTime != 10*time.Second {
		t.Errorf("expected target block time 10s, got %v", config.TargetBlockTime)
	}
}
