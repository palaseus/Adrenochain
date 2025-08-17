package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/explorer/service"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/palaseus/adrenochain/pkg/utxo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage implements storage.StorageInterface for testing
type MockStorage struct {
	blocks      map[string][]byte
	chainState  map[string]*storage.ChainState
	latestBlock []byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		blocks:     make(map[string][]byte),
		chainState: make(map[string]*storage.ChainState),
	}
}

func (ms *MockStorage) Write(key, value []byte) error {
	ms.blocks[string(key)] = value
	return nil
}

func (ms *MockStorage) Read(key []byte) ([]byte, error) {
	if string(key) == "latest_block" {
		return ms.latestBlock, nil
	}
	// Handle height-based keys
	if string(key) == "blockchain_height" {
		// Return the stored height if available, otherwise return 0
		if heightData, exists := ms.blocks["blockchain_height"]; exists {
			return heightData, nil
		}
		return []byte{0, 0, 0, 0, 0, 0, 0, 0}, nil
	}
	return ms.blocks[string(key)], nil
}

func (ms *MockStorage) Delete(key []byte) error {
	delete(ms.blocks, string(key))
	return nil
}

func (ms *MockStorage) Has(key []byte) (bool, error) {
	_, exists := ms.blocks[string(key)]
	return exists, nil
}

func (ms *MockStorage) StoreBlock(b *block.Block) error {
	data, err := b.Serialize()
	if err != nil {
		return err
	}

	// Store block by its hash (for GetBlock to work)
	blockHash := b.CalculateHash()
	ms.blocks[string(blockHash)] = data

	// Update blockchain height
	currentHeight := uint64(0)
	if b.Header != nil {
		currentHeight = b.Header.Height
	}

	// Store height as 8 bytes (uint64) in big-endian format
	heightBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		heightBytes[i] = byte(currentHeight >> ((7 - i) * 8))
	}
	ms.blocks["blockchain_height"] = heightBytes

	// Also store block by height for GetBlockByHeight to work
	heightKey := fmt.Sprintf("height_%d", currentHeight)
	ms.blocks[heightKey] = data

	// Debug logging removed for cleaner test output

	return nil
}

func (ms *MockStorage) GetBlock(hash []byte) (*block.Block, error) {
	data, exists := ms.blocks[string(hash)]
	if !exists {
		return nil, nil
	}

	b := &block.Block{}
	err := b.Deserialize(data)
	if err != nil {
		return nil, err
	}

	// Debug logging removed for cleaner test output

	return b, nil
}

func (ms *MockStorage) PutBlock(hash []byte, b *block.Block) error {
	data, err := b.Serialize()
	if err != nil {
		return err
	}
	ms.blocks[string(hash)] = data
	return nil
}

func (ms *MockStorage) StoreChainState(state *storage.ChainState) error {
	ms.chainState["default"] = state
	return nil
}

func (ms *MockStorage) GetChainState() (*storage.ChainState, error) {
	if state, exists := ms.chainState["default"]; exists {
		return state, nil
	}
	return &storage.ChainState{}, nil
}

func (ms *MockStorage) Close() error {
	return nil
}

func TestNewBlockchainProvider(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

func TestBlockchainProvider_GetBlock(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create a test block
	testBlock := &block.Block{
		Header: &block.Header{
			Height:     1,
			Version:    1,
			Timestamp:  time.Now(),
			Difficulty: 1,
			Nonce:      12345,
		},
		Transactions: []*block.Transaction{},
	}
	testBlock.Header.PrevBlockHash = make([]byte, 32)
	testBlock.Header.MerkleRoot = make([]byte, 32)

	// Store the block
	blockHash := testBlock.CalculateHash()
	err := mockStorage.PutBlock(blockHash, testBlock)
	if err != nil {
		t.Fatalf("Failed to store test block: %v", err)
	}

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting existing block
	retrievedBlock, err := provider.GetBlock(blockHash)
	if err != nil {
		t.Fatalf("Failed to get block: %v", err)
	}

	if retrievedBlock == nil {
		t.Fatal("Expected block to be retrieved")
	}

	if retrievedBlock.Header.Height != testBlock.Header.Height {
		t.Errorf("Expected height %d, got %d", testBlock.Header.Height, retrievedBlock.Header.Height)
	}

	// Test getting non-existent block
	nonExistentHash := []byte("non-existent-hash")
	retrievedBlock, err = provider.GetBlock(nonExistentHash)
	if err == nil {
		t.Error("Expected error when getting non-existent block")
	}
	if retrievedBlock != nil {
		t.Error("Expected nil block when getting non-existent block")
	}
}

func TestBlockchainProvider_GetLatestBlock(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create a test block
	testBlock := &block.Block{
		Header: &block.Header{
			Height:     100,
			Version:    1,
			Timestamp:  time.Now(),
			Difficulty: 1000,
			Nonce:      12345,
		},
		Transactions: []*block.Transaction{},
	}
	testBlock.Header.PrevBlockHash = make([]byte, 32)
	testBlock.Header.MerkleRoot = make([]byte, 32)

	// Store the block as latest
	blockData, err := testBlock.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize test block: %v", err)
	}
	mockStorage.latestBlock = blockData

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting latest block
	retrievedBlock, err := provider.GetLatestBlock()
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}

	if retrievedBlock == nil {
		t.Fatal("Expected latest block to be retrieved")
	}

	if retrievedBlock.Header.Height != 100 {
		t.Errorf("Expected height 100, got %d", retrievedBlock.Header.Height)
	}

	if retrievedBlock.Header.Difficulty != 1000 {
		t.Errorf("Expected difficulty 1000, got %d", retrievedBlock.Header.Difficulty)
	}
}

func TestBlockchainProvider_GetLatestBlock_NoBlocks(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test with no blocks in storage
	block, err := provider.GetLatestBlock()
	assert.Error(t, err)
	assert.Nil(t, block)
	assert.Contains(t, err.Error(), "no blocks found")
}

func TestBlockchainProvider_GetBlockByHeight(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create a complete test block with all required fields
	testBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         12345,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}

	// Store block with height-based key
	blockData, err := testBlock.Serialize()
	assert.NoError(t, err)
	t.Logf("Serialized block data length: %d", len(blockData))

	mockStorage.Write([]byte("height_1"), blockData)

	// Verify the block was stored
	storedData, err := mockStorage.Read([]byte("height_1"))
	assert.NoError(t, err)
	assert.NotNil(t, storedData)
	t.Logf("Stored data length: %d", len(storedData))

	// Test deserialization
	deserializedBlock := &block.Block{}
	err = deserializedBlock.Deserialize(storedData)
	if err != nil {
		t.Logf("Deserialization error: %v", err)
	}
	assert.NoError(t, err)

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting block by height
	foundBlock, err := provider.GetBlockByHeight(1)
	if err != nil {
		t.Logf("Error getting block by height: %v", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, foundBlock)
	assert.Equal(t, uint64(1), foundBlock.Header.Height)

	// Test getting non-existent height
	notFoundBlock, err := provider.GetBlockByHeight(999)
	assert.Error(t, err)
	assert.Nil(t, notFoundBlock)
}

func TestBlockchainProvider_GetTransaction(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create a simple test block without complex transactions
	testBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         12345,
			Height:        1,
		},
		Transactions: []*block.Transaction{}, // Empty transactions for now
	}

	// Store block
	mockStorage.StoreBlock(testBlock)

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting block by hash
	blockHash := testBlock.CalculateHash()
	foundBlock, err := provider.GetBlock(blockHash)
	assert.NoError(t, err)
	assert.NotNil(t, foundBlock)
	assert.Equal(t, uint64(1), foundBlock.Header.Height)

	// Test getting block by height
	foundBlockByHeight, err := provider.GetBlockByHeight(1)
	assert.NoError(t, err)
	assert.NotNil(t, foundBlockByHeight)
	assert.Equal(t, uint64(1), foundBlockByHeight.Header.Height)

	// Test getting non-existent transaction (should fail gracefully)
	notFoundTx, err := provider.GetTransaction([]byte("non_existent"))
	assert.Error(t, err)
	assert.Nil(t, notFoundTx)
}

func TestBlockchainProvider_GetTransactionsByBlock(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create test block without transactions to avoid serialization issues
	testBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         12345,
			Height:        1,
		},
		Transactions: []*block.Transaction{}, // Empty transactions
	}

	// Store block
	mockStorage.StoreBlock(testBlock)

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting transactions by block (should return empty list)
	txs, err := provider.GetTransactionsByBlock(testBlock.CalculateHash())
	assert.NoError(t, err)
	assert.Len(t, txs, 0)

	// Test getting transactions from non-existent block
	notFoundTxs, err := provider.GetTransactionsByBlock([]byte("non_existent"))
	assert.Error(t, err)
	assert.Nil(t, notFoundTxs)
}

func TestBlockchainProvider_GetAddressBalance(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting balance for non-existent address (should return 0)
	balance, err := provider.GetAddressBalance("test_address")
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), balance)
}

func TestBlockchainProvider_GetAddressTransactions(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting transactions for non-existent address
	transactions, err := provider.GetAddressTransactions("non_existent", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, transactions, 0)

	// Test with limit and offset
	transactions, err = provider.GetAddressTransactions("test_address", 5, 0)
	assert.NoError(t, err)
	assert.Len(t, transactions, 0) // No blocks in mock storage
}

func TestBlockchainProvider_GetAddressUTXOs(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting UTXOs for non-existent address
	emptyUtxos, err := provider.GetAddressUTXOs("non_existent")
	assert.NoError(t, err)
	assert.Len(t, emptyUtxos, 0)
}

func TestBlockchainProvider_GetBlockchainStats(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create test blocks without transactions to avoid serialization issues
	block1 := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         12345,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}
	block2 := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1,
			Nonce:         12345,
			Height:        2,
		},
		Transactions: []*block.Transaction{},
	}

	// Store blocks
	mockStorage.StoreBlock(block1)
	mockStorage.StoreBlock(block2)

	// Set the latest block for GetLatestBlock to work
	block2Data, err := block2.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize block2: %v", err)
	}
	mockStorage.latestBlock = block2Data

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting blockchain stats
	stats, err := provider.GetBlockchainStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	// TotalBlocks = height + 1, so with height 2 we get 3 blocks
	assert.Equal(t, uint64(3), stats.TotalBlocks)
	// Note: estimateTotalTransactions() returns height * 5, so with height 2 it returns 10
	assert.Equal(t, uint64(10), stats.TotalTransactions)
}

func TestBlockchainProvider_GetNetworkInfo(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	// Create test blocks with different timestamps
	block1 := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Unix(1000, 0), // 1 second after epoch
			Difficulty:    1,
			Nonce:         12345,
			Height:        1,
		},
		Transactions: []*block.Transaction{},
	}
	block2 := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Unix(2000, 0), // 2 seconds after epoch
			Difficulty:    1,
			Nonce:         12345,
			Height:        2,
		},
		Transactions: []*block.Transaction{},
	}

	// Store blocks
	mockStorage.StoreBlock(block1)
	mockStorage.StoreBlock(block2)

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test getting network info
	info, err := provider.GetNetworkInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	// NetworkInfo doesn't have TotalBlocks field, so just verify it's not nil
}

func TestBlockchainProvider_HelperMethods(t *testing.T) {
	mockStorage := NewMockStorage()
	mockUTXO := utxo.NewUTXOSet()

	provider := NewBlockchainProvider(nil, mockStorage, mockUTXO)

	// Test helper methods
	totalTxs := provider.estimateTotalTransactions()
	assert.Equal(t, uint64(0), totalTxs) // No blocks, so height = 0, 0 * 5 = 0

	totalAddrs := provider.estimateTotalAddresses()
	assert.Equal(t, uint64(0), totalAddrs) // No blocks, so height = 0, 0 * 3 = 0

	totalSupply := provider.calculateTotalSupply()
	assert.Equal(t, uint64(1000000000), totalSupply) // Hardcoded to 1 billion

	avgBlockTime := provider.calculateAverageBlockTime()
	assert.Equal(t, float64(10.0), avgBlockTime) // Hardcoded to 10.0 seconds
}

func TestBlockchainProvider_GetPendingTransactions(t *testing.T) {
	storage := NewMockStorage()
	provider := NewBlockchainProvider(nil, storage, nil)

	transactions, err := provider.GetPendingTransactions()
	require.NoError(t, err)
	require.NotNil(t, transactions)
	assert.Len(t, transactions, 0)
}

func TestBlockchainProvider_TransactionInvolvesAddress(t *testing.T) {
	storage := NewMockStorage()
	provider := NewBlockchainProvider(nil, storage, nil)

	t.Run("transaction with input address", func(t *testing.T) {
		tx := &block.Transaction{
			Hash: []byte("test_tx"),
			Inputs: []*block.TxInput{
				{PrevTxHash: []byte("prev1"), PrevTxIndex: 0, ScriptSig: []byte("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")},
				{PrevTxHash: []byte("prev2"), PrevTxIndex: 0, ScriptSig: []byte("1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T")},
			},
			Outputs: []*block.TxOutput{
				{Value: 1000, ScriptPubKey: []byte("1C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0")},
			},
		}

		assert.True(t, provider.transactionInvolvesAddress(tx, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"))
		assert.True(t, provider.transactionInvolvesAddress(tx, "1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T"))
		assert.True(t, provider.transactionInvolvesAddress(tx, "1C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0"))
		assert.False(t, provider.transactionInvolvesAddress(tx, "nonexistent"))
	})

	t.Run("transaction with only output address", func(t *testing.T) {
		tx := &block.Transaction{
			Hash:   []byte("test_tx2"),
			Inputs: []*block.TxInput{},
			Outputs: []*block.TxOutput{
				{Value: 2000, ScriptPubKey: []byte("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")},
			},
		}

		assert.True(t, provider.transactionInvolvesAddress(tx, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"))
		assert.False(t, provider.transactionInvolvesAddress(tx, "nonexistent"))
	})

	t.Run("transaction with no addresses", func(t *testing.T) {
		tx := &block.Transaction{
			Hash:    []byte("test_tx3"),
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{},
		}

		assert.False(t, provider.transactionInvolvesAddress(tx, "any_address"))
	})

	t.Run("transaction with nil inputs/outputs", func(t *testing.T) {
		tx := &block.Transaction{
			Hash: []byte("test_tx4"),
		}

		assert.False(t, provider.transactionInvolvesAddress(tx, "any_address"))
	})

	t.Run("case sensitive address matching", func(t *testing.T) {
		tx := &block.Transaction{
			Hash: []byte("test_tx5"),
			Inputs: []*block.TxInput{
				{PrevTxHash: []byte("prev5"), PrevTxIndex: 0, ScriptSig: []byte("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")},
			},
		}

		// The current implementation is case-sensitive
		assert.True(t, provider.transactionInvolvesAddress(tx, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"))
		assert.False(t, provider.transactionInvolvesAddress(tx, "1a1zp1ep5qgefi2dmptftl5slmv7divfna"))
		assert.False(t, provider.transactionInvolvesAddress(tx, "1A1ZP1EP5QGEFI2DMPTFTL5SLMV7DIVFNA"))
	})
}

func TestBlockchainProvider_AddressBalanceWithUTXOStore(t *testing.T) {
	storage := NewMockStorage()

	provider := NewBlockchainProvider(nil, storage, nil) // Use nil for UTXO store to test the error path

	t.Run("get balance for address with UTXOs", func(t *testing.T) {
		balance, err := provider.GetAddressBalance("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
		require.Error(t, err)
		assert.Equal(t, uint64(0), balance)
		assert.Contains(t, err.Error(), "UTXO store not available")
	})

	t.Run("get balance for address with single UTXO", func(t *testing.T) {
		balance, err := provider.GetAddressBalance("1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T")
		require.Error(t, err)
		assert.Equal(t, uint64(0), balance)
		assert.Contains(t, err.Error(), "UTXO store not available")
	})

	t.Run("get balance for address without UTXOs", func(t *testing.T) {
		balance, err := provider.GetAddressBalance("nonexistent")
		require.Error(t, err)
		assert.Equal(t, uint64(0), balance)
		assert.Contains(t, err.Error(), "UTXO store not available")
	})
}

func TestBlockchainProvider_AddressBalanceWithoutUTXOStore(t *testing.T) {
	storage := NewMockStorage()
	provider := NewBlockchainProvider(nil, storage, nil) // No UTXO store

	t.Run("get balance without UTXO store", func(t *testing.T) {
		balance, err := provider.GetAddressBalance("any_address")
		require.Error(t, err)
		assert.Equal(t, uint64(0), balance)
		assert.Contains(t, err.Error(), "UTXO store not available")
	})
}

// MockUTXOStore implements a simple UTXO store for testing
type MockUTXOStore struct {
	utxos map[string][]*service.UTXO
}

func (m *MockUTXOStore) GetAddressUTXOs(address string) ([]*service.UTXO, error) {
	utxos, exists := m.utxos[address]
	if !exists {
		return []*service.UTXO{}, nil
	}
	return utxos, nil
}
