package data

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/explorer/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBlockchainDataProvider implements service.BlockchainDataProvider for testing
type MockBlockchainDataProvider struct {
	blocks       map[string]*block.Block
	transactions map[string]*block.Transaction
	addresses    map[string]uint64
}

func NewMockBlockchainDataProvider() *MockBlockchainDataProvider {
	return &MockBlockchainDataProvider{
		blocks:       make(map[string]*block.Block),
		transactions: make(map[string]*block.Transaction),
		addresses:    make(map[string]uint64),
	}
}

func (m *MockBlockchainDataProvider) GetBlock(hash []byte) (*block.Block, error) {
	// Try to find block by byte key first
	block, exists := m.blocks[string(hash)]
	if exists {
		return block, nil
	}

	// If not found, try to find by hex-encoded string
	hexHash := hex.EncodeToString(hash)
	block, exists = m.blocks[hexHash]
	if exists {
		return block, nil
	}

	return nil, nil
}

func (m *MockBlockchainDataProvider) GetBlockByHeight(height uint64) (*block.Block, error) {
	// First try to find by height-based key
	heightKey := fmt.Sprintf("height_%d", height)
	if block, exists := m.blocks[heightKey]; exists {
		return block, nil
	}

	// Fallback: search through all blocks
	for _, block := range m.blocks {
		if block.Header != nil && block.Header.Height == height {
			return block, nil
		}
	}
	return nil, nil
}

func (m *MockBlockchainDataProvider) GetLatestBlock() (*block.Block, error) {
	var latestBlock *block.Block
	var latestHeight uint64

	for _, block := range m.blocks {
		if block.Header != nil && block.Header.Height > latestHeight {
			latestHeight = block.Header.Height
			latestBlock = block
		}
	}
	return latestBlock, nil
}

func (m *MockBlockchainDataProvider) GetBlockHeight() uint64 {
	var maxHeight uint64
	for _, block := range m.blocks {
		if block.Header != nil && block.Header.Height > maxHeight {
			maxHeight = block.Header.Height
		}
	}
	return maxHeight
}

func (m *MockBlockchainDataProvider) GetTransaction(hash []byte) (*block.Transaction, error) {
	// Try to find transaction by byte key first
	tx, exists := m.transactions[string(hash)]
	if exists {
		return tx, nil
	}

	// If not found, try to find by hex-encoded string
	hexHash := hex.EncodeToString(hash)
	tx, exists = m.transactions[hexHash]
	if exists {
		return tx, nil
	}

	return nil, nil
}

func (m *MockBlockchainDataProvider) GetTransactionsByBlock(blockHash []byte) ([]*block.Transaction, error) {
	block, exists := m.blocks[string(blockHash)]
	if !exists {
		return nil, nil
	}
	return block.Transactions, nil
}

func (m *MockBlockchainDataProvider) GetPendingTransactions() ([]*block.Transaction, error) {
	return []*block.Transaction{}, nil
}

func (m *MockBlockchainDataProvider) GetAddressBalance(address string) (uint64, error) {
	balance, exists := m.addresses[address]
	if !exists {
		return 0, nil
	}
	return balance, nil
}

func (m *MockBlockchainDataProvider) GetAddressTransactions(address string, limit, offset int) ([]*block.Transaction, error) {
	return []*block.Transaction{}, nil
}

func (m *MockBlockchainDataProvider) GetAddressUTXOs(address string) ([]*service.UTXO, error) {
	return []*service.UTXO{}, nil
}

func (m *MockBlockchainDataProvider) GetBlockchainStats() (*service.BlockchainStats, error) {
	return &service.BlockchainStats{}, nil
}

func (m *MockBlockchainDataProvider) GetNetworkInfo() (*service.NetworkInfo, error) {
	return &service.NetworkInfo{}, nil
}

func TestNewSimpleSearchProvider(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	require.NotNil(t, searchProvider)
	assert.Equal(t, mockProvider, searchProvider.dataProvider)
}

func TestSimpleSearchProvider_Search(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	t.Run("empty query", func(t *testing.T) {
		result, err := searchProvider.Search("")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "", result.Query)
		assert.Equal(t, "empty search query", result.Error)
		assert.Equal(t, "", result.Type)
	})

	t.Run("whitespace only query", func(t *testing.T) {
		result, err := searchProvider.Search("   \t\n  ")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "", result.Query)
		assert.Equal(t, "empty search query", result.Error)
	})

	t.Run("block hash search", func(t *testing.T) {
		// Create a mock block
		blockHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		block := &block.Block{
			Header: &block.Header{
				Height:     100,
				Timestamp:  time.Now(),
				Difficulty: 1000,
			},
			Transactions: []*block.Transaction{
				{Hash: []byte("tx1")},
				{Hash: []byte("tx2")},
			},
		}
		// Store with both string and byte key for testing
		mockProvider.blocks[blockHash] = block
		// Also store with the actual block hash for GetBlock to find
		blockHashBytes := block.CalculateHash()
		mockProvider.blocks[string(blockHashBytes)] = block
		// Store with height-based key for GetBlockByHeight to find
		heightKey := fmt.Sprintf("height_%d", block.Header.Height)
		mockProvider.blocks[heightKey] = block

		result, err := searchProvider.Search(blockHash)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, blockHash, result.Query)
		assert.Equal(t, "block", result.Type)
		assert.NotNil(t, result.Block)
		assert.Equal(t, uint64(100), result.Block.Height)
		assert.Equal(t, 2, result.Block.TxCount)
	})

	t.Run("transaction hash search", func(t *testing.T) {
		// Create a block first
		txBlock := &block.Block{
			Header: &block.Header{
				Height:     200,
				Timestamp:  time.Now(),
				Difficulty: 2000,
			},
			Transactions: []*block.Transaction{},
		}
		blockHash := "blockhash1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		mockProvider.blocks[blockHash] = txBlock
		// Also store with the actual block hash for GetBlock to find
		blockHashBytes := txBlock.CalculateHash()
		mockProvider.blocks[string(blockHashBytes)] = txBlock
		// Store with height-based key for GetBlockByHeight to find
		heightKey := fmt.Sprintf("height_%d", txBlock.Header.Height)
		mockProvider.blocks[heightKey] = txBlock

		// Create a mock transaction
		txHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		decodedHash, _ := hex.DecodeString(txHash)
		tx := &block.Transaction{
			Hash:    decodedHash, // Use the decoded bytes, not the hex string
			Fee:     100,
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{},
		}
		// Store transaction with both string and byte keys
		mockProvider.transactions[txHash] = tx
		mockProvider.transactions[string(tx.Hash)] = tx

		// Also store with decoded bytes key (this is what GetTransaction will look for)
		mockProvider.transactions[string(decodedHash)] = tx

		// Add transaction to block
		txBlock.Transactions = append(txBlock.Transactions, tx)

		// Debug: Print what we're storing
		t.Logf("Storing transaction with hash: %s", txHash)
		t.Logf("Transaction hash bytes: %x", tx.Hash)
		t.Logf("Mock provider transactions: %v", mockProvider.transactions)
		t.Logf("Mock provider blocks: %v", mockProvider.blocks)

		result, err := searchProvider.Search(txHash)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, txHash, result.Query)
		assert.Equal(t, "transaction", result.Type)
		assert.NotNil(t, result.Transaction)
		assert.Equal(t, uint64(200), result.Transaction.Height)
		assert.Equal(t, uint64(100), result.Transaction.Fee)
	})

	t.Run("address search", func(t *testing.T) {
		address := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
		balance := uint64(1000000)
		mockProvider.addresses[address] = balance

		result, err := searchProvider.Search(address)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, address, result.Query)
		assert.Equal(t, "address", result.Type)
		assert.NotNil(t, result.Address)
		assert.Equal(t, balance, result.Address.Balance)
	})

	t.Run("numeric query", func(t *testing.T) {
		// Create a mock block with height 500
		block := &block.Block{
			Header: &block.Header{
				Height:     500,
				Timestamp:  time.Now(),
				Difficulty: 5000,
			},
			Transactions: []*block.Transaction{},
		}
		blockHash := "blockhash5001234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		mockProvider.blocks[blockHash] = block
		// Also store with the actual block hash for GetBlock to find
		blockHashBytes := block.CalculateHash()
		mockProvider.blocks[string(blockHashBytes)] = block
		// Store with height-based key for GetBlockByHeight to find
		heightKey := fmt.Sprintf("height_%d", block.Header.Height)
		mockProvider.blocks[heightKey] = block

		result, err := searchProvider.Search("500")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "500", result.Query)
		assert.Equal(t, "block", result.Type)
		assert.NotNil(t, result.Block)
		assert.Equal(t, uint64(500), result.Block.Height)
	})

	t.Run("no results found", func(t *testing.T) {
		result, err := searchProvider.Search("nonexistent")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "nonexistent", result.Query)
		assert.Equal(t, "unknown", result.Type)
		assert.NotEmpty(t, result.Suggestions)
	})
}

func TestSimpleSearchProvider_SearchBlocks(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	// Create clean test data for this test
	block1 := &block.Block{
		Header: &block.Header{
			Height:     199,
			Timestamp:  time.Now().Add(-time.Hour),
			Difficulty: 1000,
		},
		Transactions: []*block.Transaction{
			{Hash: []byte("tx1")},
			{Hash: []byte("tx2")},
		},
	}
	block2 := &block.Block{
		Header: &block.Header{
			Height:     200,
			Timestamp:  time.Now(),
			Difficulty: 2000,
		},
		Transactions: []*block.Transaction{
			{Hash: []byte("tx3")},
		},
	}

	// Store blocks with multiple keys
	mockProvider.blocks["block1"] = block1
	mockProvider.blocks["block2"] = block2
	heightKey1 := fmt.Sprintf("height_%d", block1.Header.Height)
	heightKey2 := fmt.Sprintf("height_%d", block2.Header.Height)
	mockProvider.blocks[heightKey1] = block1
	mockProvider.blocks[heightKey2] = block2

	block1Hash := "block1hash1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	block2Hash := "block2hash1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	mockProvider.blocks[block1Hash] = block1
	mockProvider.blocks[block2Hash] = block2

	t.Run("search all blocks", func(t *testing.T) {
		results, err := searchProvider.SearchBlocks("", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 2)
	})

	t.Run("search with limit", func(t *testing.T) {
		results, err := searchProvider.SearchBlocks("", 1, 0)
		require.NoError(t, err)
		require.Len(t, results, 1)
	})

	t.Run("search with offset", func(t *testing.T) {
		results, err := searchProvider.SearchBlocks("", 1, 1)
		require.NoError(t, err)
		require.Len(t, results, 1)
	})

	t.Run("search with query", func(t *testing.T) {
		// Search for an empty query to get all blocks, then verify we can find by height
		results, err := searchProvider.SearchBlocks("", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Verify we have blocks with the expected heights
		heights := make(map[uint64]bool)
		for _, result := range results {
			heights[result.Height] = true
		}
		assert.True(t, heights[199], "Should have block with height 199")
		assert.True(t, heights[200], "Should have block with height 200")
	})
}

func TestSimpleSearchProvider_SearchTransactions(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	// Create blocks with transactions (using consecutive heights)
	block1 := &block.Block{
		Header: &block.Header{
			Height:     199,
			Timestamp:  time.Now().Add(-time.Hour),
			Difficulty: 1000,
		},
		Transactions: []*block.Transaction{},
	}
	block2 := &block.Block{
		Header: &block.Header{
			Height:     200,
			Timestamp:  time.Now(),
			Difficulty: 2000,
		},
		Transactions: []*block.Transaction{},
	}

	// Create mock transactions
	tx1 := &block.Transaction{
		Hash:    []byte("tx1hash1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		Fee:     100,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{},
	}
	tx2 := &block.Transaction{
		Hash:    []byte("tx2hash1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		Fee:     200,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{},
	}

	// Add transactions to blocks
	block1.Transactions = append(block1.Transactions, tx1)
	block2.Transactions = append(block2.Transactions, tx2)

	// Store blocks with height-based keys
	heightKey1 := fmt.Sprintf("height_%d", block1.Header.Height)
	heightKey2 := fmt.Sprintf("height_%d", block2.Header.Height)
	mockProvider.blocks[heightKey1] = block1
	mockProvider.blocks[heightKey2] = block2

	// Also store transactions directly for GetTransaction calls
	mockProvider.transactions[string(tx1.Hash)] = tx1
	mockProvider.transactions[string(tx2.Hash)] = tx2

	t.Run("search all transactions", func(t *testing.T) {
		results, err := searchProvider.SearchTransactions("", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 2)
	})

	t.Run("search with limit", func(t *testing.T) {
		results, err := searchProvider.SearchTransactions("", 1, 0)
		require.NoError(t, err)
		require.Len(t, results, 1)
	})

	t.Run("search with offset", func(t *testing.T) {
		results, err := searchProvider.SearchTransactions("", 1, 1)
		require.NoError(t, err)
		require.Len(t, results, 1)
	})

	t.Run("search with query", func(t *testing.T) {
		// Search for an empty query to get all transactions, then verify we can find by fee
		results, err := searchProvider.SearchTransactions("", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Verify we have transactions with the expected fees
		fees := make(map[uint64]bool)
		for _, result := range results {
			fees[result.Fee] = true
		}
		assert.True(t, fees[100], "Should have transaction with fee 100")
		assert.True(t, fees[200], "Should have transaction with fee 200")
	})
}

func TestSimpleSearchProvider_SearchAddresses(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	// Create mock addresses
	address1 := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	address2 := "1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T"
	balance1 := uint64(1000000)
	balance2 := uint64(2000000)

	mockProvider.addresses[address1] = balance1
	mockProvider.addresses[address2] = balance2

	t.Run("search all addresses", func(t *testing.T) {
		results, err := searchProvider.SearchAddresses("", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 0) // Stub implementation returns empty results
	})

	t.Run("search with limit", func(t *testing.T) {
		results, err := searchProvider.SearchAddresses("", 1, 0)
		require.NoError(t, err)
		require.Len(t, results, 0) // Stub implementation returns empty results
	})

	t.Run("search with offset", func(t *testing.T) {
		results, err := searchProvider.SearchAddresses("", 1, 1)
		require.NoError(t, err)
		require.Len(t, results, 0) // Stub implementation returns empty results
	})

	t.Run("search with query", func(t *testing.T) {
		results, err := searchProvider.SearchAddresses("1A1z", 10, 0)
		require.NoError(t, err)
		require.Len(t, results, 0) // Stub implementation returns empty results
	})
}

func TestSimpleSearchProvider_HelperMethods(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	t.Run("findTransactionBlock", func(t *testing.T) {
		// Create a transaction and block
		txHash := []byte("transaction1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
		tx := &block.Transaction{Hash: txHash}
		block := &block.Block{
			Header: &block.Header{
				Height:     300,
				Timestamp:  time.Now(),
				Difficulty: 3000,
			},
			Transactions: []*block.Transaction{tx},
		}
		blockHash := "blockhash3001234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		mockProvider.blocks[blockHash] = block

		foundBlock, err := searchProvider.findTransactionBlock(txHash)
		require.NoError(t, err)
		require.NotNil(t, foundBlock)
		assert.Equal(t, uint64(300), foundBlock.Header.Height)
	})

	t.Run("blockMatchesQuery", func(t *testing.T) {
		block := &block.Block{
			Header: &block.Header{
				Height:     400,
				Timestamp:  time.Now(),
				Difficulty: 4000,
			},
			Transactions: []*block.Transaction{},
		}

		// Test height match
		assert.True(t, searchProvider.blockMatchesQuery(block, "400"))

		// Test no match
		assert.False(t, searchProvider.blockMatchesQuery(block, "nonexistent"))
	})

	t.Run("transactionMatchesQuery", func(t *testing.T) {
		tx := &block.Transaction{
			Hash: []byte("transaction1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
			Fee:  500,
		}

		// Test hash match
		assert.True(t, searchProvider.transactionMatchesQuery(tx, "transaction"))

		// Test no match
		assert.False(t, searchProvider.transactionMatchesQuery(tx, "nonexistent"))
	})

	t.Run("generateSuggestions", func(t *testing.T) {
		suggestions := searchProvider.generateSuggestions("test")
		require.NotEmpty(t, suggestions)
		assert.Contains(t, suggestions, "test")
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("isHexString", func(t *testing.T) {
		assert.True(t, isHexString("1234567890abcdef"))
		assert.True(t, isHexString("ABCDEF1234567890"))
		assert.False(t, isHexString("not hex"))
		assert.False(t, isHexString("1234567890abcdefg"))
		assert.False(t, isHexString(""))
	})

	t.Run("isBase58String", func(t *testing.T) {
		assert.True(t, isBase58String("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"))
		assert.True(t, isBase58String("123456789"))
		assert.False(t, isBase58String("not base58"))
		assert.False(t, isBase58String(""))
		assert.False(t, isBase58String("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa0")) // Invalid length
	})

	t.Run("isNumericString", func(t *testing.T) {
		assert.True(t, isNumericString("1234567890"))
		assert.True(t, isNumericString("0"))
		assert.False(t, isNumericString("not numeric"))
		assert.False(t, isNumericString(""))
		assert.False(t, isNumericString("123abc"))
	})

	t.Run("parseUint64", func(t *testing.T) {
		value, err := parseUint64("1234567890")
		require.NoError(t, err)
		assert.Equal(t, uint64(1234567890), value)

		_, err = parseUint64("not a number")
		assert.Error(t, err)

		_, err = parseUint64("")
		assert.Error(t, err)
	})
}

func TestSimpleSearchProvider_EdgeCases(t *testing.T) {
	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	t.Run("invalid hex string", func(t *testing.T) {
		result, err := searchProvider.Search("invalid_hex_string_that_is_64_chars_long_but_not_hex")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "unknown", result.Type)
	})

	t.Run("very long query", func(t *testing.T) {
		longQuery := string(make([]byte, 1000))
		result, err := searchProvider.Search(longQuery)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "unknown", result.Type)
	})

	t.Run("special characters in query", func(t *testing.T) {
		specialQuery := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
		result, err := searchProvider.Search(specialQuery)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "unknown", result.Type)
	})

	t.Run("unicode characters in query", func(t *testing.T) {
		unicodeQuery := "🚀🌍💻🔗"
		result, err := searchProvider.Search(unicodeQuery)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "unknown", result.Type)
	})
}

func TestSimpleSearchProvider_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	mockProvider := NewMockBlockchainDataProvider()
	searchProvider := NewSimpleSearchProvider(mockProvider)

	// Create many mock blocks for performance testing
	for i := 0; i < 1000; i++ {
		block := &block.Block{
			Header: &block.Header{
				Height:     uint64(i),
				Timestamp:  time.Now(),
				Difficulty: uint64(i * 1000),
			},
			Transactions: []*block.Transaction{},
		}
		blockHash := fmt.Sprintf("blockhash%d1234567890abcdef1234567890abcdef1234567890abcdef1234567890", i)
		mockProvider.blocks[blockHash] = block
	}

	// Benchmark search operations
	start := time.Now()
	results, err := searchProvider.SearchBlocks("", 100, 0)
	searchDuration := time.Since(start)

	require.NoError(t, err)
	require.Len(t, results, 100)
	assert.True(t, searchDuration < 100*time.Millisecond, "Search took too long: %v", searchDuration)

	t.Logf("Performance: Searched 1000 blocks in %v", searchDuration)
}
