package chain

import (
	"bytes"
	"math/big"
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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
	// Calculate the actual transaction hash
	tx.Hash = tx.CalculateHash()
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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
	// Calculate the actual transaction hash
	dummyTx.Hash = dummyTx.CalculateHash()
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
	// Calculate the actual transaction hash
	dummyTx2.Hash = dummyTx2.CalculateHash()
	wrongHeightBlock.AddTransaction(dummyTx2)
	wrongHeightBlock.Header.MerkleRoot = wrongHeightBlock.CalculateMerkleRoot()

	if err := chain.AddBlock(wrongHeightBlock); err == nil {
		t.Error("Should fail to add block with wrong height")
	}
}

func TestGetBlock(t *testing.T) {
	dataDir := "./test_chain_data_get_block"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
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

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test with block within size limit
	validBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: chain.GetTipHash(),
			MerkleRoot:    nil, // Will be calculated
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         0,
			Height:        1,
		},
		Transactions: []*block.Transaction{
			{
				Version: 1,
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{
					{
						Value:        1000000,
						ScriptPubKey: []byte("script"),
					},
				},
				LockTime: 0,
			},
		},
	}

	// Calculate transaction hashes first
	for _, tx := range validBlock.Transactions {
		tx.Hash = tx.CalculateHash()
	}

	// Calculate merkle root
	validBlock.Header.MerkleRoot = validBlock.CalculateMerkleRoot()

	// Mine the block to satisfy proof of work
	err = chain.GetConsensus().MineBlock(validBlock, nil)
	if err != nil {
		t.Fatalf("Failed to mine block: %v", err)
	}

	err = chain.AddBlock(validBlock)
	if err != nil {
		t.Errorf("Expected valid block to be added, got error: %v", err)
	}

	// Test with block exceeding size limit
	largeBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: chain.GetTipHash(),
			MerkleRoot:    nil, // Will be calculated
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         0,
			Height:        2,
		},
		Transactions: []*block.Transaction{
			{
				Version: 1,
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{
					{
						Value:        1000000,
						ScriptPubKey: make([]byte, config.MaxBlockSize+1), // Exceed max size
					},
				},
				LockTime: 0,
			},
		},
	}

	// Calculate transaction hashes first
	for _, tx := range largeBlock.Transactions {
		tx.Hash = tx.CalculateHash()
	}

	// Calculate merkle root
	largeBlock.Header.MerkleRoot = largeBlock.CalculateMerkleRoot()

	// Mine the large block to satisfy proof of work (this should succeed)
	err = chain.GetConsensus().MineBlock(largeBlock, nil)
	if err != nil {
		t.Fatalf("Failed to mine large block: %v", err)
	}

	// Debug: Print the actual block size vs max block size
	actualSize := chain.GetBlockSize(largeBlock)
	maxSize := chain.config.MaxBlockSize
	t.Logf("Large block size: %d, Max block size: %d", actualSize, maxSize)

	err = chain.AddBlock(largeBlock)
	if err == nil {
		t.Error("Expected error for block exceeding size limit")
	}
}

func TestGetBlockByHeight(t *testing.T) {
	dataDir := "./test_chain_data_get_block_by_height"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test getting genesis block by height
	genesisBlock := chain.GetBlockByHeight(0)
	if genesisBlock == nil {
		t.Fatal("Expected genesis block at height 0")
	}
	if genesisBlock.Header.Height != 0 {
		t.Errorf("Expected height 0, got %d", genesisBlock.Header.Height)
	}

	// Test getting non-existent height
	nonExistentBlock := chain.GetBlockByHeight(999)
	if nonExistentBlock != nil {
		t.Error("Expected nil for non-existent height")
	}
}

func TestGetBestBlock(t *testing.T) {
	dataDir := "./test_chain_data_get_best_block"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	bestBlock := chain.GetBestBlock()
	if bestBlock == nil {
		t.Fatal("Expected best block to be non-nil")
	}
	if bestBlock.Header.Height != 0 {
		t.Errorf("Expected height 0, got %d", bestBlock.Header.Height)
	}
}

func TestGetHeight(t *testing.T) {
	dataDir := "./test_chain_data_get_height"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	height := chain.GetHeight()
	if height != 0 {
		t.Errorf("Expected height 0, got %d", height)
	}
}

func TestGetTipHash(t *testing.T) {
	dataDir := "./test_chain_data_get_tip_hash"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	tipHash := chain.GetTipHash()
	if tipHash == nil {
		t.Fatal("Expected tip hash to be non-nil")
	}
	if len(tipHash) != 32 {
		t.Errorf("Expected tip hash length 32, got %d", len(tipHash))
	}
}

func TestGetGenesisBlock(t *testing.T) {
	dataDir := "./test_chain_data_get_genesis_block"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	genesisBlock := chain.GetGenesisBlock()
	if genesisBlock == nil {
		t.Fatal("Expected genesis block to be non-nil")
	}
	if genesisBlock.Header.Height != 0 {
		t.Errorf("Expected genesis height 0, got %d", genesisBlock.Header.Height)
	}
}

func TestCalculateNextDifficulty(t *testing.T) {
	dataDir := "./test_chain_data_calculate_next_difficulty"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	difficulty := chain.CalculateNextDifficulty()
	if difficulty == 0 {
		t.Error("Expected non-zero difficulty")
	}
}

func TestGetAccumulatedDifficulty(t *testing.T) {
	dataDir := "./test_chain_data_get_accumulated_difficulty"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test getting accumulated difficulty for genesis
	accDiff, err := chain.GetAccumulatedDifficulty(0)
	if err != nil {
		t.Errorf("Expected no error for genesis height, got: %v", err)
	}
	if accDiff == nil {
		t.Error("Expected non-nil accumulated difficulty")
	}
	if accDiff.Cmp(big.NewInt(0)) != 0 {
		t.Error("Expected genesis accumulated difficulty to be 0")
	}

	// Test getting accumulated difficulty for non-existent height
	_, err = chain.GetAccumulatedDifficulty(999)
	if err == nil {
		t.Error("Expected error for non-existent height")
	}
}

func TestForkChoice(t *testing.T) {
	dataDir := "./test_chain_data_fork_choice"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test fork choice with nil block
	err = chain.ForkChoice(nil)
	if err == nil {
		t.Error("Expected error for nil block")
	}

	// Test fork choice with valid block
	validBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: chain.GetTipHash(),
			MerkleRoot:    nil, // Will be calculated
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         0,
			Height:        1,
		},
		Transactions: []*block.Transaction{
			{
				Version: 1,
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{
					{
						Value:        1000000,
						ScriptPubKey: []byte("script"),
					},
				},
				LockTime: 0,
			},
		},
	}

	// Calculate transaction hashes first
	for _, tx := range validBlock.Transactions {
		tx.Hash = tx.CalculateHash()
	}

	// Calculate merkle root
	validBlock.Header.MerkleRoot = validBlock.CalculateMerkleRoot()

	// Mine the block to find a valid nonce
	err = chain.GetConsensus().MineBlock(validBlock, nil)
	if err != nil {
		t.Fatalf("Failed to mine block: %v", err)
	}

	err = chain.ForkChoice(validBlock)
	if err != nil {
		t.Errorf("Expected no error for valid fork choice, got: %v", err)
	}
}

func TestClose(t *testing.T) {
	dataDir := "./test_chain_data_close"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	err = chain.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}
}

func TestString(t *testing.T) {
	dataDir := "./test_chain_data_string"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	str := chain.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func TestChainConcurrency(t *testing.T) {
	dataDir := "./test_chain_data_concurrency"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test concurrent access to chain methods
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Concurrent reads
			chain.GetHeight()
			chain.GetTipHash()
			chain.GetBestBlock()
			chain.GetGenesisBlock()

			// Concurrent writes (should be safe due to mutex)
			chain.CalculateNextDifficulty()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestChainErrorHandling(t *testing.T) {
	dataDir := "./test_chain_data_error_handling"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Test with invalid block (nil header)
	invalidBlock := &block.Block{
		Header:       nil,
		Transactions: []*block.Transaction{},
	}

	err = chain.AddBlock(invalidBlock)
	if err == nil {
		t.Error("Expected error for block with nil header")
	}

	// Test with block having invalid height
	invalidHeightBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: chain.GetTipHash(),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         0,
			Height:        999, // Invalid height
		},
		Transactions: []*block.Transaction{},
	}

	err = chain.AddBlock(invalidHeightBlock)
	if err == nil {
		t.Error("Expected error for block with invalid height")
	}
}

func TestChainStatePersistence(t *testing.T) {
	dataDir := "./test_chain_data_persistence"
	defer os.RemoveAll(dataDir)

	storageInstance, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	config := DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, consensusConfig, storageInstance)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Get initial state
	initialHeight := chain.GetHeight()
	initialTipHash := chain.GetTipHash()

	// Close the chain
	err = chain.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}

	// Reopen storage and create new chain
	storage2, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage2.Close()

	chain2, err := NewChain(config, consensusConfig, storage2)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}

	// Verify state was persisted
	if chain2.GetHeight() != initialHeight {
		t.Errorf("Expected height %d, got %d", initialHeight, chain2.GetHeight())
	}

	if !bytes.Equal(chain2.GetTipHash(), initialTipHash) {
		t.Error("Expected tip hash to be persisted")
	}
}
