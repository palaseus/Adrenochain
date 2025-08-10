package chain

import (
	"os"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewChain(t *testing.T) {
	dataDir := "./test_chain_data_new_chain"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

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
	dataDir := "./test_chain_data_genesis_block"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

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
	dataDir := "./test_chain_data_add_block"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Create a valid block
	prevBlock := chain.GetBestBlock()
	prevHash := prevBlock.CalculateHash()

	newBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    nil, // Will be calculated after adding transactions
			Timestamp:     time.Now(),
			Difficulty:    chain.CalculateNextDifficulty(),
			Nonce:         0,
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
	tx.Hash = make([]byte, 32) // Set a dummy hash for validation
	newBlock.AddTransaction(tx)
	newBlock.Header.MerkleRoot = newBlock.CalculateMerkleRoot()

	// Mine the block to satisfy proof of work
	stopChan := make(chan struct{})
	defer close(stopChan)
	err = chain.consensus.MineBlock(newBlock, stopChan)
	assert.NoError(t, err, "Failed to mine block")

	// Add block should succeed
	if err := chain.AddBlock(newBlock); err != nil {
		t.Errorf("Failed to add valid block: %v", err)
	}

	if chain.GetHeight() != 1 {
		t.Errorf("Expected height 1, got %d", chain.GetHeight())
	}

	if chain.GetBestBlock().HexHash() != newBlock.HexHash() {
		t.Error("Best block was not updated")
	}
}

func TestBlockValidation(t *testing.T) {
	dataDir := "./test_chain_data_block_validation"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test invalid block (wrong prev hash)
	invalidBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: []byte("wrong_hash"),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    0, // Use difficulty 0 for testing (any hash is valid)
			Nonce:         42,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}

	// Add a dummy transaction to ensure MerkleRoot is calculated
	dummyTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{{Value: 1, ScriptPubKey: []byte("dummy")}},
	}
	dummyTx.Hash = make([]byte, 32)
	invalidBlock.AddTransaction(dummyTx)
	invalidBlock.Header.MerkleRoot = invalidBlock.CalculateMerkleRoot()

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
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    0, // Use difficulty 0 for testing (any hash is valid)
			Nonce:         42,
			Height:        2, // Wrong height
		},
		Transactions: []*block.Transaction{},
	}

	// Add a dummy transaction to ensure MerkleRoot is calculated
	dummyTx2 := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{{Value: 1, ScriptPubKey: []byte("dummy")}},
	}
	dummyTx2.Hash = make([]byte, 32)
	wrongHeightBlock.AddTransaction(dummyTx2)
	wrongHeightBlock.Header.MerkleRoot = wrongHeightBlock.CalculateMerkleRoot()

	if err := chain.AddBlock(wrongHeightBlock); err == nil {
		t.Error("Should fail to add block with wrong height")
	}
}

func TestGetBlock(t *testing.T) {
	dataDir := "./test_chain_data_get_block"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Get genesis block by hash
	genesis := chain.GetGenesisBlock()
	genesisHash := genesis.CalculateHash()

	retrievedBlock := chain.GetBlock(genesisHash)
	if retrievedBlock == nil {
		t.Fatal("Failed to retrieve genesis block by hash")
	}

	if retrievedBlock.HexHash() != genesis.HexHash() {
		t.Error("Retrieved block is not the same as genesis block")
	}

	// Get genesis block by height
	retrievedBlockByHeight := chain.GetBlockByHeight(0)
	if retrievedBlockByHeight == nil {
		t.Fatal("Failed to retrieve genesis block by height")
	}

	if retrievedBlockByHeight.HexHash() != genesis.HexHash() {
		t.Error("Retrieved block by height is not the same as genesis block")
	}
}

func TestChainState(t *testing.T) {
	dataDir := "./test_chain_data_chain_state"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Check initial state
	if chain.GetHeight() != 0 {
		t.Errorf("Expected initial height 0, got %d", chain.GetHeight())
	}

	genesis := chain.GetGenesisBlock()
	if chain.GetBestBlock().HexHash() != genesis.HexHash() {
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
	dataDir := "./test_chain_data_difficulty_calculation"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

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

	if config.MaxBlockSize == 0 {
		t.Error("Max block size should not be zero")
	}
}

func TestBlockSizeValidation(t *testing.T) {
	dataDir := "./test_chain_data_block_size_validation"
	defer os.RemoveAll(dataDir)

	storage, storageErr := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if storageErr != nil {
		t.Fatalf("Failed to create storage: %v", storageErr)
	}
	defer storage.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Create a block that exceeds max size
	prevBlock := chain.GetBestBlock()
	prevHash := prevBlock.CalculateHash()

	largeBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: prevHash,
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    chain.CalculateNextDifficulty(),
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
