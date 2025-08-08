package chain

import (
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

func TestNewChain(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	if chain == nil {
		t.Fatal("NewChain returned nil")
	}

	if chain.genesisBlock == nil {
		t.Fatal("Genesis block was not created")
	}

	if chain.bestBlock == nil {
		t.Fatal("Best block was not set")
	}

	if chain.height != 0 {
		t.Errorf("Expected height 0, got %d", chain.height)
	}

	if chain.genesisBlock.Header.Height != 0 {
		t.Errorf("Expected genesis height 0, got %d", chain.genesisBlock.Header.Height)
	}
}

func TestGenesisBlock(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	genesis := chain.GetGenesisBlock()
	if genesis == nil {
		t.Fatal("Genesis block is nil")
	}

	if genesis.Header.Height != 0 {
		t.Errorf("Expected genesis height 0, got %d", genesis.Header.Height)
	}

	if len(genesis.Header.PrevBlockHash) != 32 {
		t.Errorf("Expected genesis prev hash length 32, got %d", len(genesis.Header.PrevBlockHash))
	}

	// Check if all bytes are zero
	for _, b := range genesis.Header.PrevBlockHash {
		if b != 0 {
			t.Error("Genesis prev hash should be all zeros")
		}
	}

	if len(genesis.Transactions) != 1 {
		t.Errorf("Expected 1 genesis transaction, got %d", len(genesis.Transactions))
	}

	// Check coinbase transaction
	coinbaseTx := genesis.Transactions[0]
	if coinbaseTx == nil {
		t.Fatal("Coinbase transaction is nil")
	}

	if len(coinbaseTx.Inputs) != 0 {
		t.Error("Coinbase transaction should have no inputs")
	}

	if len(coinbaseTx.Outputs) != 1 {
		t.Error("Coinbase transaction should have one output")
	}

	if coinbaseTx.Outputs[0].Value != config.GenesisBlockReward {
		t.Errorf("Expected coinbase reward %d, got %d", config.GenesisBlockReward, coinbaseTx.Outputs[0].Value)
	}
}

func TestAddBlock(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Create a valid block
	prevBlock := chain.GetBestBlock()
	prevHash := prevBlock.CalculateHash()
	
	newBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    0, // Use difficulty 0 for testing (any hash is valid)
			Nonce:         42,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}

	// Add transaction to calculate merkle root
	tx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{}, // Empty inputs for coinbase
		Outputs: []*block.TxOutput{
			{
				Value:        1000, // Coinbase reward
				ScriptPubKey: []byte("coinbase_output"),
			},
		},
		Fee: 0, // Coinbase has no fee
	}
	newBlock.AddTransaction(tx)
	newBlock.Header.MerkleRoot = newBlock.CalculateMerkleRoot()

	// Add block should succeed
	if err := chain.AddBlock(newBlock); err != nil {
		t.Errorf("Failed to add valid block: %v", err)
	}

	if chain.GetHeight() != 1 {
		t.Errorf("Expected height 1, got %d", chain.GetHeight())
	}

	if chain.GetBestBlock() != newBlock {
		t.Error("Best block was not updated")
	}
}

func TestBlockValidation(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Test invalid block (wrong prev hash)
	invalidBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte("wrong_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    0, // Use difficulty 0 for testing (any hash is valid)
			Nonce:         42,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}

	if err := chain.AddBlock(invalidBlock); err == nil {
		t.Error("Should fail to add block with wrong prev hash")
	}

	// Test invalid block (wrong height)
	prevBlock := chain.GetBestBlock()
	prevHash := prevBlock.CalculateHash()
	
	wrongHeightBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    0, // Use difficulty 0 for testing (any hash is valid)
			Nonce:         42,
			Height:        2, // Wrong height
		},
		Transactions: []*block.Transaction{},
	}

	if err := chain.AddBlock(wrongHeightBlock); err == nil {
		t.Error("Should fail to add block with wrong height")
	}
}

func TestGetBlock(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Get genesis block by hash
	genesis := chain.GetGenesisBlock()
	genesisHash := genesis.CalculateHash()
	
	retrievedBlock := chain.GetBlock(genesisHash)
	if retrievedBlock == nil {
		t.Fatal("Failed to retrieve genesis block by hash")
	}

	if retrievedBlock != genesis {
		t.Error("Retrieved block is not the same as genesis block")
	}

	// Get genesis block by height
	retrievedBlockByHeight := chain.GetBlockByHeight(0)
	if retrievedBlockByHeight == nil {
		t.Fatal("Failed to retrieve genesis block by height")
	}

	if retrievedBlockByHeight != genesis {
		t.Error("Retrieved block by height is not the same as genesis block")
	}
}

func TestChainState(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Check initial state
	if chain.GetHeight() != 0 {
		t.Errorf("Expected initial height 0, got %d", chain.GetHeight())
	}

	genesis := chain.GetGenesisBlock()
	if chain.GetBestBlock() != genesis {
		t.Error("Best block should be genesis block initially")
	}

	if chain.GetTipHash() == nil {
		t.Error("Tip hash should not be nil")
	}

	// Check tip hash matches genesis
	genesisHash := genesis.CalculateHash()
	if string(chain.GetTipHash()) != string(genesisHash) {
		t.Error("Tip hash should match genesis block hash")
	}
}

func TestDifficultyCalculation(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Initial difficulty should be 1
	if chain.CalculateNextDifficulty() != 1 {
		t.Errorf("Expected initial difficulty 1, got %d", chain.CalculateNextDifficulty())
	}

	// Add some blocks to test difficulty adjustment
	// This would require implementing a more sophisticated difficulty adjustment algorithm
	// For now, we just test that the function doesn't panic
	_ = chain.CalculateNextDifficulty()
}

func TestChainConfig(t *testing.T) {
	config := DefaultChainConfig()

	if config.GenesisBlockReward == 0 {
		t.Error("Genesis block reward should not be zero")
	}

	if config.BlockTime <= 0 {
		t.Error("Block time should be positive")
	}

	if config.DifficultyAdjustmentInterval == 0 {
		t.Error("Difficulty adjustment interval should not be zero")
	}

	if config.MaxBlockSize == 0 {
		t.Error("Max block size should not be zero")
	}
}

func TestBlockSizeValidation(t *testing.T) {
	config := DefaultChainConfig()
	chain := NewChain(config)

	// Create a block that exceeds max size
	prevBlock := chain.GetBestBlock()
	prevHash := prevBlock.CalculateHash()
	
	largeBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    1000,
			Nonce:         42,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}

	// Add many large transactions to exceed block size limit
	for i := 0; i < 1000; i++ {
		tx := &block.Transaction{
			Version: 1,
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{},
			Fee:     10,
		}
		largeBlock.AddTransaction(tx)
	}

	largeBlock.Header.MerkleRoot = largeBlock.CalculateMerkleRoot()

	// This should fail due to block size validation
	if err := chain.AddBlock(largeBlock); err == nil {
		t.Error("Should fail to add block that exceeds max size")
	}
} 