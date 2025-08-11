package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBlockchainDataProvider is a mock implementation of BlockchainDataProvider
type MockBlockchainDataProvider struct {
	mock.Mock
}

func (m *MockBlockchainDataProvider) GetBlock(hash []byte) (*block.Block, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetBlockByHeight(height uint64) (*block.Block, error) {
	args := m.Called(height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetLatestBlock() (*block.Block, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*block.Block), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetBlockHeight() uint64 {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *MockBlockchainDataProvider) GetTransaction(hash []byte) (*block.Transaction, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*block.Transaction), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetTransactionsByBlock(blockHash []byte) ([]*block.Transaction, error) {
	args := m.Called(blockHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*block.Transaction), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetPendingTransactions() ([]*block.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*block.Transaction), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetAddressBalance(address string) (uint64, error) {
	args := m.Called(address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetAddressTransactions(address string, limit, offset int) ([]*block.Transaction, error) {
	args := m.Called(address, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*block.Transaction), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetAddressUTXOs(address string) ([]*UTXO, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*UTXO), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetBlockchainStats() (*BlockchainStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BlockchainStats), args.Error(1)
}

func (m *MockBlockchainDataProvider) GetNetworkInfo() (*NetworkInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*NetworkInfo), args.Error(1)
}

// MockCacheProvider is a mock implementation of CacheProvider
type MockCacheProvider struct {
	mock.Mock
}

func (m *MockCacheProvider) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheProvider) Set(key string, value interface{}, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockCacheProvider) Delete(key string) {
	m.Called(key)
}

func (m *MockCacheProvider) Clear() {
	m.Called()
}

func (m *MockCacheProvider) GetStats() CacheStats {
	args := m.Called()
	return args.Get(0).(CacheStats)
}

// MockSearchProvider is a mock implementation of SearchProvider
type MockSearchProvider struct {
	mock.Mock
}

func (m *MockSearchProvider) Search(query string) (*SearchResult, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SearchResult), args.Error(1)
}

func (m *MockSearchProvider) SearchBlocks(query string, limit, offset int) ([]*BlockSummary, error) {
	args := m.Called(query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*BlockSummary), args.Error(1)
}

func (m *MockSearchProvider) SearchTransactions(query string, limit, offset int) ([]*TransactionSummary, error) {
	args := m.Called(query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TransactionSummary), args.Error(1)
}

func (m *MockSearchProvider) SearchAddresses(query string, limit, offset int) ([]*AddressSummary, error) {
	args := m.Called(query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AddressSummary), args.Error(1)
}

// Helper function to create test blocks
func createTestBlock(height uint64, hash []byte, prevHash []byte, txCount int) *block.Block {
	testBlock := &block.Block{
		Header: &block.Header{
			Height:        height,
			Version:       1,
			Timestamp:     time.Now(),
			Difficulty:    1000,
			Nonce:         12345,
			PrevBlockHash: prevHash,
			MerkleRoot:    make([]byte, 32),
		},
		Transactions: make([]*block.Transaction, txCount),
	}

	// Set the hash
	copy(testBlock.Header.MerkleRoot, hash)

	// Create test transactions
	for i := 0; i < txCount; i++ {
		testBlock.Transactions[i] = &block.Transaction{
			Hash:    []byte(fmt.Sprintf("tx-%d", i)),
			Inputs:  []*block.TxInput{},
			Outputs: []*block.TxOutput{},
		}
	}

	return testBlock
}

// Helper function to create test transactions
func createTestTransaction(hash []byte, inputs, outputs int) *block.Transaction {
	testTx := &block.Transaction{
		Hash:    hash,
		Inputs:  make([]*block.TxInput, inputs),
		Outputs: make([]*block.TxOutput, outputs),
	}

	// Create test inputs
	for i := 0; i < inputs; i++ {
		testTx.Inputs[i] = &block.TxInput{
			PrevTxHash:  []byte(fmt.Sprintf("input-tx-%d", i)),
			PrevTxIndex: uint32(i),
			ScriptSig:   []byte(fmt.Sprintf("script-%d", i)),
		}
	}

	// Create test outputs
	for i := 0; i < outputs; i++ {
		testTx.Outputs[i] = &block.TxOutput{
			Value:        uint64(1000 + i*100),
			ScriptPubKey: []byte(fmt.Sprintf("script-%d", i)),
		}
	}

	return testTx
}

func TestNewExplorerService(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	assert.NotNil(t, service)
	// Note: We can't test private fields directly, but we can test that the service works
}

func TestExplorerService_GetDashboard(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	stats := &BlockchainStats{
		TotalBlocks:       1000,
		TotalTransactions: 5000,
		TotalAddresses:    2000,
		TotalSupply:       1000000000,
		LastBlockTime:     time.Now(),
		AverageBlockTime:  10.5,
		Difficulty:        1000,
	}

	networkInfo := &NetworkInfo{
		Status:          "active",
		PeerCount:       10,
		IsListening:     true,
		LastUpdate:      time.Now(),
		NetworkVersion:  "1.0.0",
		ProtocolVersion: "1.0.0",
	}

	// Setup expectations
	mockDataProvider.On("GetBlockchainStats").Return(stats, nil)
	mockDataProvider.On("GetNetworkInfo").Return(networkInfo, nil)
	mockDataProvider.On("GetLatestBlock").Return(createTestBlock(1000, []byte("latest"), []byte("prev"), 5), nil)

	// Execute
	ctx := context.Background()
	dashboard, err := service.GetDashboard(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
	assert.Equal(t, stats, dashboard.Stats)
	assert.Equal(t, networkInfo, dashboard.NetworkInfo)
	assert.Len(t, dashboard.RecentBlocks, 1)
	assert.Len(t, dashboard.RecentTxs, 5)

	mockDataProvider.AssertExpectations(t)
}

func TestExplorerService_GetBlockDetails(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	blockHash := []byte("test-block-hash")
	prevHash := []byte("prev-block-hash")
	testBlock := createTestBlock(100, blockHash, prevHash, 3)

	// Create the previous block that will be returned
	prevBlock := createTestBlock(99, prevHash, []byte("prev-prev"), 2)

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "block:"+string(blockHash)).Return(nil, false)
	mockCacheProvider.On("Set", "block:"+string(blockHash), mock.Anything, mock.Anything).Return()

	// Setup blockchain expectations
	mockDataProvider.On("GetBlock", blockHash).Return(testBlock, nil)
	mockDataProvider.On("GetBlockByHeight", uint64(99)).Return(prevBlock, nil)
	mockDataProvider.On("GetBlockByHeight", uint64(101)).Return(nil, assert.AnError)
	mockDataProvider.On("GetBlockHeight").Return(uint64(1000))

	// Execute
	ctx := context.Background()
	blockDetails, err := service.GetBlockDetails(ctx, blockHash)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, blockDetails)
	assert.Equal(t, testBlock.CalculateHash(), blockDetails.Hash)
	assert.Equal(t, uint64(100), blockDetails.Height)
	assert.Equal(t, prevBlock.CalculateHash(), blockDetails.PrevHash)
	assert.Equal(t, 3, blockDetails.TxCount)
	assert.Len(t, blockDetails.Transactions, 3)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetBlockDetails_BlockNotFound(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	blockHash := []byte("non-existent-block")

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "block:"+string(blockHash)).Return(nil, false)

	// Setup blockchain expectations
	mockDataProvider.On("GetBlock", blockHash).Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	blockDetails, err := service.GetBlockDetails(ctx, blockHash)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, blockDetails)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetTransactionDetails(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	txHash := []byte("test-tx-hash")
	testTx := createTestTransaction(txHash, 2, 3)
	testBlock := createTestBlock(100, []byte("block-hash"), []byte("prev-hash"), 1)
	testBlock.Transactions[0] = testTx

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "tx:"+string(txHash)).Return(nil, false)
	mockCacheProvider.On("Set", "tx:"+string(txHash), mock.Anything, mock.Anything).Return()

	// Setup blockchain expectations
	mockDataProvider.On("GetTransaction", txHash).Return(testTx, nil)
	mockDataProvider.On("GetBlockHeight").Return(uint64(1000))

	// Setup expectations for findTransactionBlock search
	for h := uint64(0); h <= 100; h++ {
		if h == 100 {
			mockDataProvider.On("GetBlockByHeight", h).Return(testBlock, nil)
		} else {
			mockDataProvider.On("GetBlockByHeight", h).Return(nil, assert.AnError)
		}
	}

	// Execute
	ctx := context.Background()
	txDetails, err := service.GetTransactionDetails(ctx, txHash)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, txDetails)
	assert.Equal(t, txHash, txDetails.Hash)
	assert.Equal(t, 2, txDetails.Inputs)
	assert.Equal(t, 3, txDetails.Outputs)
	assert.Equal(t, testTx, txDetails.RawTx)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetTransactionDetails_TransactionNotFound(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	txHash := []byte("non-existent-tx")

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "tx:"+string(txHash)).Return(nil, false)

	// Setup blockchain expectations
	mockDataProvider.On("GetTransaction", txHash).Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	txDetails, err := service.GetTransactionDetails(ctx, txHash)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, txDetails)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetAddressDetails(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	address := "test-address"
	balance := uint64(5000)
	utxos := []*UTXO{
		{
			TxHash:  []byte("utxo-tx-1"),
			TxIndex: 0,
			Value:   3000,
			Address: address,
		},
		{
			TxHash:  []byte("utxo-tx-2"),
			TxIndex: 1,
			Value:   2000,
			Address: address,
		},
	}
	transactions := []*block.Transaction{
		createTestTransaction([]byte("tx-1"), 1, 2),
		createTestTransaction([]byte("tx-2"), 1, 1),
	}

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "address:"+address).Return(nil, false)
	mockCacheProvider.On("Set", "address:"+address, mock.Anything, mock.Anything).Return()

	// Setup blockchain expectations
	mockDataProvider.On("GetAddressBalance", address).Return(balance, nil)
	mockDataProvider.On("GetAddressUTXOs", address).Return(utxos, nil)
	mockDataProvider.On("GetAddressTransactions", address, 50, 0).Return(transactions, nil)
	mockDataProvider.On("GetBlockHeight").Return(uint64(1000))

	// Setup expectations for findTransactionBlock search (for each transaction)
	for h := uint64(0); h <= 100; h++ {
		if h == 100 {
			// Return a block for the transaction
			mockDataProvider.On("GetBlockByHeight", h).Return(createTestBlock(h, []byte(fmt.Sprintf("block-%d", h)), []byte("prev"), 5), nil)
		} else {
			mockDataProvider.On("GetBlockByHeight", h).Return(nil, assert.AnError)
		}
	}

	// Execute
	ctx := context.Background()
	addressDetails, err := service.GetAddressDetails(ctx, address)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, addressDetails)
	assert.Equal(t, address, addressDetails.Address)
	assert.Equal(t, balance, addressDetails.Balance)
	assert.Len(t, addressDetails.UTXOs, 2)
	assert.Len(t, addressDetails.Transactions, 2)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetBlocks(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	limit := 10
	offset := 0
	height := uint64(1000)

	// Create test blocks
	var blocks []*block.Block
	for i := 0; i < limit; i++ {
		blockHeight := height - uint64(i)
		block := createTestBlock(blockHeight, []byte(fmt.Sprintf("block-%d", blockHeight)), []byte("prev"), 5)
		blocks = append(blocks, block)
	}

	// Setup expectations
	mockDataProvider.On("GetBlockHeight").Return(height)
	for i := 0; i < limit; i++ {
		blockHeight := height - uint64(i)
		mockDataProvider.On("GetBlockByHeight", blockHeight).Return(blocks[i], nil)
	}

	// Execute
	ctx := context.Background()
	blockSummaries, err := service.GetBlocks(ctx, limit, offset)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, blockSummaries)
	assert.Len(t, blockSummaries, limit)
	assert.Equal(t, height, blockSummaries[0].Height)
	assert.Equal(t, height-uint64(limit-1), blockSummaries[len(blockSummaries)-1].Height)

	mockDataProvider.AssertExpectations(t)
}

func TestExplorerService_GetTransactions(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	limit := 10
	offset := 0
	height := uint64(1000)

	// Create test transactions
	var transactions []*block.Transaction
	for i := 0; i < limit; i++ {
		tx := createTestTransaction([]byte(fmt.Sprintf("tx-%d", i)), 1, 2)
		transactions = append(transactions, tx)
	}

	// Setup expectations
	mockDataProvider.On("GetBlockHeight").Return(height)
	for i := 0; i < limit; i++ {
		blockHeight := height - uint64(i/5) // 5 transactions per block
		mockDataProvider.On("GetBlockByHeight", blockHeight).Return(createTestBlock(blockHeight, []byte(fmt.Sprintf("block-%d", blockHeight)), []byte("prev"), 5), nil)
	}

	// Execute
	ctx := context.Background()
	txSummaries, err := service.GetTransactions(ctx, limit, offset)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, txSummaries)
	assert.Len(t, txSummaries, limit)

	mockDataProvider.AssertExpectations(t)
}

func TestExplorerService_Search(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	query := "test-query"
	searchResult := &SearchResult{
		Query: query,
		Type:  "block",
		Block: &BlockSummary{
			Hash:   []byte("test-block"),
			Height: 100,
		},
	}

	// Setup cache expectations (cache miss)
	mockCacheProvider.On("Get", "search:"+query).Return(nil, false)
	mockCacheProvider.On("Set", "search:"+query, mock.Anything, mock.Anything).Return()

	// Setup search expectations
	mockSearchProvider.On("Search", query).Return(searchResult, nil)

	// Execute
	ctx := context.Background()
	result, err := service.Search(ctx, query)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, query, result.Query)
	assert.Equal(t, "block", result.Type)

	mockSearchProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_GetStatistics(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	stats := &BlockchainStats{
		TotalBlocks:       1000,
		TotalTransactions: 5000,
		TotalAddresses:    2000,
		TotalSupply:       1000000000,
		LastBlockTime:     time.Now(),
		AverageBlockTime:  10.5,
		Difficulty:        1000,
	}

	networkInfo := &NetworkInfo{
		Status:          "active",
		PeerCount:       10,
		IsListening:     true,
		LastUpdate:      time.Now(),
		NetworkVersion:  "1.0.0",
		ProtocolVersion: "1.0.0",
	}

	cacheStats := CacheStats{
		Hits:    100,
		Misses:  50,
		HitRate: 0.67,
		Size:    1000,
		MaxSize: 10000,
	}

	// Setup expectations
	mockDataProvider.On("GetBlockchainStats").Return(stats, nil)
	mockDataProvider.On("GetNetworkInfo").Return(networkInfo, nil)
	mockCacheProvider.On("GetStats").Return(cacheStats)

	// Execute
	ctx := context.Background()
	statistics, err := service.GetStatistics(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, statistics)
	assert.Equal(t, stats, statistics.Blockchain)
	assert.Equal(t, networkInfo, statistics.Network)
	assert.Equal(t, cacheStats.HitRate, statistics.Performance.CacheHitRate)

	mockDataProvider.AssertExpectations(t)
	mockCacheProvider.AssertExpectations(t)
}

func TestExplorerService_WithCache(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test data
	blockHash := []byte("cached-block-hash")
	cachedBlock := createTestBlock(100, blockHash, []byte("prev"), 3)

	// Create a BlockDetails object to return from cache
	blockDetails := &BlockDetails{
		BlockSummary: &BlockSummary{
			Hash:          cachedBlock.CalculateHash(),
			Height:        cachedBlock.Header.Height,
			Timestamp:     cachedBlock.Header.Timestamp,
			TxCount:       len(cachedBlock.Transactions),
			Size:          uint64(len(cachedBlock.Transactions) * 100),
			Difficulty:    cachedBlock.Header.Difficulty,
			Confirmations: 0,
		},
		PrevHash:   cachedBlock.Header.PrevBlockHash,
		MerkleRoot: cachedBlock.Header.MerkleRoot,
		Nonce:      cachedBlock.Header.Nonce,
		Version:    cachedBlock.Header.Version,
		Validation: &BlockValidation{
			IsValid:       true,
			Confirmations: 0,
			Finality:      "unconfirmed",
		},
	}

	// Setup cache expectations
	mockCacheProvider.On("Get", "block:"+string(blockHash)).Return(blockDetails, true)

	// Execute
	ctx := context.Background()
	blockDetails, err := service.GetBlockDetails(ctx, blockHash)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, blockDetails)
	assert.Equal(t, cachedBlock.CalculateHash(), blockDetails.Hash)

	// Verify cache was used and blockchain provider was not called
	mockCacheProvider.AssertExpectations(t)
	mockDataProvider.AssertNotCalled(t, "GetBlock", mock.Anything)
}

func TestExplorerService_ErrorHandling(t *testing.T) {
	mockDataProvider := &MockBlockchainDataProvider{}
	mockCacheProvider := &MockCacheProvider{}
	mockSearchProvider := &MockSearchProvider{}

	service := NewExplorerService(mockDataProvider, mockCacheProvider, mockSearchProvider)

	// Test various error scenarios
	t.Run("GetBlockchainStats_Error", func(t *testing.T) {
		mockDataProvider.On("GetBlockchainStats").Return(nil, assert.AnError)

		ctx := context.Background()
		dashboard, err := service.GetDashboard(ctx)

		assert.Error(t, err)
		assert.Nil(t, dashboard)
	})

	t.Run("GetNetworkInfo_Error", func(t *testing.T) {
		stats := &BlockchainStats{TotalBlocks: 100}
		mockDataProvider.On("GetBlockchainStats").Return(stats, nil)
		mockDataProvider.On("GetNetworkInfo").Return(nil, assert.AnError)

		ctx := context.Background()
		dashboard, err := service.GetDashboard(ctx)

		assert.Error(t, err)
		assert.Nil(t, dashboard)
	})

	t.Run("GetLatestBlock_Error", func(t *testing.T) {
		stats := &BlockchainStats{TotalBlocks: 100}
		networkInfo := &NetworkInfo{Status: "active"}
		mockDataProvider.On("GetBlockchainStats").Return(stats, nil)
		mockDataProvider.On("GetNetworkInfo").Return(networkInfo, nil)
		mockDataProvider.On("GetLatestBlock").Return(nil, assert.AnError)

		ctx := context.Background()
		dashboard, err := service.GetDashboard(ctx)

		assert.Error(t, err)
		assert.Nil(t, dashboard)
	})
}
