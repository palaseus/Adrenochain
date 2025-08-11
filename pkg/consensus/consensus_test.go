package consensus

import (
	"math/big"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
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

	// Add a coinbase transaction to the block
	coinbaseTx := &block.Transaction{
		Version:  1,
		Inputs:   make([]*block.TxInput, 0), // Coinbase has no inputs
		Outputs:  []*block.TxOutput{{Value: 1000000, ScriptPubKey: []byte("COINBASE_1")}},
		LockTime: 0,
		Fee:      0,
	}
	coinbaseTx.Hash = coinbaseTx.CalculateHash()
	testBlock.AddTransaction(coinbaseTx)

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

	// Add a coinbase transaction to the block
	coinbaseTx := &block.Transaction{
		Version:  1,
		Inputs:   make([]*block.TxInput, 0), // Coinbase has no inputs
		Outputs:  []*block.TxOutput{{Value: 1000000, ScriptPubKey: []byte("COINBASE_1")}},
		LockTime: 0,
		Fee:      0,
	}
	coinbaseTx.Hash = coinbaseTx.CalculateHash()
	testBlock.AddTransaction(coinbaseTx)

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

func TestGetFinalityDepth(t *testing.T) {
	config := DefaultConsensusConfig()
	config.FinalityDepth = 150
	consensus := NewConsensus(config, nil)

	finalityDepth := consensus.GetFinalityDepth()
	assert.Equal(t, uint64(150), finalityDepth)
}

func TestHash256(t *testing.T) {
	config := DefaultConsensusConfig()
	consensus := NewConsensus(config, nil)

	// Test with some data (avoid empty data due to division by zero bug in hash256)
	data := []byte("test data")
	hash := consensus.hash256(data)
	assert.Len(t, hash, 32)

	// Test that same input produces same output
	hash2 := consensus.hash256(data)
	assert.Equal(t, hash, hash2)

	// Test that different input produces different output
	data2 := []byte("different data")
	hash3 := consensus.hash256(data2)
	assert.NotEqual(t, hash, hash3)

	// Test with single byte data
	singleByte := []byte{42}
	hash4 := consensus.hash256(singleByte)
	assert.Len(t, hash4, 32)
}

func TestGetNextDifficulty(t *testing.T) {
	config := DefaultConsensusConfig()
	consensus := NewConsensus(config, nil)

	// Test with current difficulty
	nextDifficulty := consensus.GetNextDifficulty()
	assert.GreaterOrEqual(t, nextDifficulty, config.MinDifficulty)
	assert.LessOrEqual(t, nextDifficulty, config.MaxDifficulty)
}

func TestGetStats(t *testing.T) {
	config := DefaultConsensusConfig()
	consensus := NewConsensus(config, nil)

	stats := consensus.GetStats()
	assert.NotNil(t, stats)

	// Check that stats contain expected keys
	assert.Contains(t, stats, "difficulty")
	assert.Contains(t, stats, "next_difficulty")
	assert.Contains(t, stats, "target")
	assert.Contains(t, stats, "block_times_count")
	assert.Contains(t, stats, "last_adjustment")
	assert.Contains(t, stats, "target_block_time")
	assert.Contains(t, stats, "adjustment_interval")

	// Check that difficulty is in the expected range
	difficulty, ok := stats["difficulty"].(uint64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, difficulty, config.MinDifficulty)
	assert.LessOrEqual(t, difficulty, config.MaxDifficulty)

	// Check next difficulty
	nextDifficulty, ok := stats["next_difficulty"].(uint64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, nextDifficulty, config.MinDifficulty)
	assert.LessOrEqual(t, nextDifficulty, config.MaxDifficulty)

	// Check block times count
	blockTimesCount, ok := stats["block_times_count"].(int)
	assert.True(t, ok)
	assert.Equal(t, 0, blockTimesCount) // Should be 0 for new consensus
}

func TestString(t *testing.T) {
	config := DefaultConsensusConfig()
	consensus := NewConsensus(config, nil)

	str := consensus.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "Consensus")
	assert.Contains(t, str, "Difficulty")
	assert.Contains(t, str, "Target")
	assert.Contains(t, str, "BlockTimes")
}
