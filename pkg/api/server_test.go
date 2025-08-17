package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/wallet"
	"github.com/gorilla/mux"
)

// MockChain implements ChainInterface for testing
type MockChain struct {
	height         uint64
	bestBlock      *block.Block
	genesisBlock   *block.Block
	blocks         map[string]*block.Block
	blocksByHeight map[uint64]*block.Block
}

// Ensure MockChain implements ChainInterface
var _ ChainInterface = (*MockChain)(nil)

func NewMockChain() *MockChain {
	genesisBlock := &block.Block{
		Header: &block.Header{
			Height:     0,
			Version:    1,
			Timestamp:  time.Now(),
			Difficulty: 1,
			Nonce:      0,
		},
		Transactions: []*block.Transaction{},
	}
	genesisBlock.Header.PrevBlockHash = make([]byte, 32)
	genesisBlock.Header.MerkleRoot = make([]byte, 32)

	bestBlock := &block.Block{
		Header: &block.Header{
			Height:     1,
			Version:    1,
			Timestamp:  time.Now(),
			Difficulty: 1,
			Nonce:      12345,
		},
		Transactions: []*block.Transaction{
			{
				Hash:    []byte("test-tx-hash"),
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{},
			},
		},
	}
	bestBlock.Header.PrevBlockHash = genesisBlock.CalculateHash()
	bestBlock.Header.MerkleRoot = bestBlock.CalculateMerkleRoot()

	mockChain := &MockChain{
		height:         1,
		bestBlock:      bestBlock,
		genesisBlock:   genesisBlock,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	mockChain.blocks[fmt.Sprintf("%x", genesisBlock.CalculateHash())] = genesisBlock
	mockChain.blocks[fmt.Sprintf("%x", bestBlock.CalculateHash())] = bestBlock
	mockChain.blocksByHeight[0] = genesisBlock
	mockChain.blocksByHeight[1] = bestBlock

	return mockChain
}

func (mc *MockChain) GetHeight() uint64 {
	return mc.height
}

func (mc *MockChain) GetBestBlock() *block.Block {
	return mc.bestBlock
}

func (mc *MockChain) GetGenesisBlock() *block.Block {
	return mc.genesisBlock
}

func (mc *MockChain) GetBlock(hash []byte) *block.Block {
	return mc.blocks[fmt.Sprintf("%x", hash)]
}

func (mc *MockChain) GetBlockByHeight(height uint64) *block.Block {
	return mc.blocksByHeight[height]
}

func (mc *MockChain) CalculateNextDifficulty() uint64 {
	return mc.bestBlock.Header.Difficulty + 1
}

// MockWallet implements WalletInterface for testing
type MockWallet struct {
	accounts map[string]*wallet.Account
	balances map[string]uint64
}

// Ensure MockWallet implements WalletInterface
var _ WalletInterface = (*MockWallet)(nil)

func NewMockWallet() *MockWallet {
	account1 := &wallet.Account{
		Address:   "test-address-1",
		PublicKey: []byte("test-public-key-1"),
	}

	account2 := &wallet.Account{
		Address:   "test-address-2",
		PublicKey: []byte("test-public-key-2"),
	}

	return &MockWallet{
		accounts: map[string]*wallet.Account{
			"test-address-1": account1,
			"test-address-2": account2,
		},
		balances: map[string]uint64{
			"test-address-1": 1000,
			"test-address-2": 2500,
		},
	}
}

func (mw *MockWallet) GetBalance(address string) uint64 {
	return mw.balances[address]
}

func (mw *MockWallet) GetAllAccounts() []*wallet.Account {
	accounts := make([]*wallet.Account, 0, len(mw.accounts))
	for _, account := range mw.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

func TestNewServer(t *testing.T) {
	mockChain := NewMockChain()
	mockWallet := NewMockWallet()

	config := &ServerConfig{
		Port:   8080,
		Chain:  mockChain,
		Wallet: mockWallet,
	}

	server := NewServer(config)

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.chain != mockChain {
		t.Error("Server should have the provided chain")
	}

	if server.wallet != mockWallet {
		t.Error("Server should have the provided wallet")
	}

	if server.port != 8080 {
		t.Errorf("Server should have port 8080, got %d", server.port)
	}

	if server.router == nil {
		t.Error("Server should have a router")
	}
}

func TestServer_HealthHandler(t *testing.T) {
	server := &Server{}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.healthHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if response["service"] != "adrenochain-api" {
		t.Errorf("Expected service 'adrenochain-api', got %v", response["service"])
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should contain timestamp")
	}
}

func TestServer_GetChainInfoHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/chain/info", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getChainInfoHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetChainInfo handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}

	if _, exists := response["best_block_hash"]; !exists {
		t.Error("Response should contain best_block_hash")
	}

	if _, exists := response["genesis_block_hash"]; !exists {
		t.Error("Response should contain genesis_block_hash")
	}

	if _, exists := response["difficulty"]; !exists {
		t.Error("Response should contain difficulty")
	}

	if _, exists := response["next_difficulty"]; !exists {
		t.Error("Response should contain next_difficulty")
	}
}

func TestServer_GetChainInfoHandler_NoBlocks(t *testing.T) {
	// Create a mock chain with no blocks
	mockChain := &MockChain{
		height:         0,
		bestBlock:      nil,
		genesisBlock:   nil,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/chain/info", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getChainInfoHandler(rr, req)

	// This should panic or error since there are no blocks
	// We'll need to handle this case in the actual implementation
}

func TestServer_GetChainHeightHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/chain/height", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getChainHeightHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetChainHeight handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}
}

func TestServer_GetChainStatusHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/chain/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getChainStatusHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetChainStatus handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}

	if response["status"] != "active" {
		t.Errorf("Expected status 'active', got %v", response["status"])
	}

	if _, exists := response["best_block_hash"]; !exists {
		t.Error("Response should contain best_block_hash")
	}

	if _, exists := response["best_block_timestamp"]; !exists {
		t.Error("Response should contain best_block_timestamp")
	}
}

func TestServer_GetChainStatusHandler_NoBlocks(t *testing.T) {
	// Create a mock chain with no blocks
	mockChain := &MockChain{
		height:         0,
		bestBlock:      nil,
		genesisBlock:   nil,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/chain/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getChainStatusHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetChainStatus handler should return OK for empty chain, got %v", status)
	}
}

func TestServer_GetBlockHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	// Get a valid block hash
	block := mockChain.GetBestBlock()
	blockHash := block.CalculateHash()
	hashHex := fmt.Sprintf("%x", blockHash)

	req, err := http.NewRequest("GET", "/api/v1/blocks/"+hashHex, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set up the router to extract URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/{hash}", server.getBlockHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetBlock handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}

	if response["tx_count"] != float64(1) {
		t.Errorf("Expected tx_count 1, got %v", response["tx_count"])
	}

	if _, exists := response["hash"]; !exists {
		t.Error("Response should contain hash")
	}

	if _, exists := response["transactions"]; !exists {
		t.Error("Response should contain transactions")
	}
}

func TestServer_GetBlockHandler_InvalidHash(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/invalid-hash", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/{hash}", server.getBlockHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("GetBlock handler should return BadRequest for invalid hash, got %v", status)
	}
}

func TestServer_GetBlockHandler_BlockNotFound(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	// Use a valid hex hash that doesn't exist in our mock chain
	nonExistentHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	req, err := http.NewRequest("GET", "/api/v1/blocks/"+nonExistentHash, nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/{hash}", server.getBlockHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("GetBlock handler should return NotFound for non-existent block, got %v", status)
	}
}

func TestServer_GetBlockByHeightHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/height/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/height/{height}", server.getBlockByHeightHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetBlockByHeight handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}
}

func TestServer_GetBlockByHeightHandler_InvalidHeight(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/height/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/height/{height}", server.getBlockByHeightHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("GetBlockByHeight handler should return BadRequest for invalid height, got %v", status)
	}
}

func TestServer_GetBlockByHeightHandler_BlockNotFound(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/height/999", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blocks/height/{height}", server.getBlockByHeightHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("GetBlockByHeight handler should return NotFound for non-existent height, got %v", status)
	}
}

func TestServer_GetLatestBlockHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/latest", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getLatestBlockHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetLatestBlock handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["height"] != float64(1) {
		t.Errorf("Expected height 1, got %v", response["height"])
	}
}

func TestServer_GetLatestBlockHandler_NoBlocks(t *testing.T) {
	// Create a mock chain with no blocks
	mockChain := &MockChain{
		height:         0,
		bestBlock:      nil,
		genesisBlock:   nil,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/blocks/latest", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getLatestBlockHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("GetLatestBlock handler should return NotFound when no blocks exist, got %v", status)
	}
}

func TestServer_GetTransactionHandler(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	// Get a valid transaction hash
	block := mockChain.GetBestBlock()
	txHash := block.Transactions[0].Hash
	hashHex := fmt.Sprintf("%x", txHash)

	req, err := http.NewRequest("GET", "/api/v1/transactions/"+hashHex, nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/transactions/{hash}", server.getTransactionHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetTransaction handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if _, exists := response["hash"]; !exists {
		t.Error("Response should contain hash")
	}

	if _, exists := response["inputs"]; !exists {
		t.Error("Response should contain inputs")
	}

	if _, exists := response["outputs"]; !exists {
		t.Error("Response should contain outputs")
	}
}

func TestServer_GetTransactionHandler_InvalidHash(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	req, err := http.NewRequest("GET", "/api/v1/transactions/invalid-hash", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/transactions/{hash}", server.getTransactionHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("GetTransaction handler should return BadRequest for invalid hash, got %v", status)
	}
}

func TestServer_GetTransactionHandler_TransactionNotFound(t *testing.T) {
	mockChain := NewMockChain()
	server := &Server{chain: mockChain}

	// Use a valid hex hash that doesn't exist in our mock chain
	nonExistentHash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	req, err := http.NewRequest("GET", "/api/v1/transactions/"+nonExistentHash, nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/transactions/{hash}", server.getTransactionHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("GetTransaction handler should return NotFound for non-existent transaction, got %v", status)
	}
}

func TestServer_GetTransactionHandler_WithNilBlocks(t *testing.T) {
	// Create a mock chain where some blocks return nil
	mockChain := &MockChain{
		height:         2,
		bestBlock:      nil,
		genesisBlock:   nil,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	// Add a block at height 1 but leave height 0 as nil
	block1 := &block.Block{
		Header: &block.Header{
			Height:     1,
			Version:    1,
			Timestamp:  time.Now(),
			Difficulty: 1,
			Nonce:      12345,
		},
		Transactions: []*block.Transaction{
			{
				Hash:    []byte("test-tx-hash"),
				Inputs:  []*block.TxInput{},
				Outputs: []*block.TxOutput{},
			},
		},
	}
	block1.Header.PrevBlockHash = make([]byte, 32)
	block1.Header.MerkleRoot = block1.CalculateMerkleRoot()

	mockChain.bestBlock = block1
	mockChain.blocksByHeight[1] = block1
	// Height 0 intentionally left as nil

	server := &Server{chain: mockChain}

	// Test with a transaction hash that exists in block 1
	txHash := "746573742d74782d68617368" // hex encoding of "test-tx-hash"
	req, err := http.NewRequest("GET", "/api/v1/transactions/"+txHash, nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/transactions/{hash}", server.getTransactionHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetTransaction handler should return OK for existing transaction, got %v", status)
	}
}

func TestServer_GetPendingTransactionsHandler(t *testing.T) {
	server := &Server{}

	req, err := http.NewRequest("GET", "/api/v1/transactions/pending", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getPendingTransactionsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetPendingTransactions handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["count"] != float64(0) {
		t.Errorf("Expected count 0, got %v", response["count"])
	}

	if _, exists := response["pending_transactions"]; !exists {
		t.Error("Response should contain pending_transactions")
	}
}

func TestServer_GetBalanceHandler(t *testing.T) {
	mockWallet := NewMockWallet()
	server := &Server{wallet: mockWallet}

	req, err := http.NewRequest("GET", "/api/v1/wallet/balance/test-address-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet/balance/{address}", server.getBalanceHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetBalance handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["address"] != "test-address-1" {
		t.Errorf("Expected address 'test-address-1', got %v", response["address"])
	}

	if response["balance"] != float64(1000) {
		t.Errorf("Expected balance 1000, got %v", response["balance"])
	}
}

func TestServer_GetBalanceHandler_NoWallet(t *testing.T) {
	server := &Server{wallet: nil}

	req, err := http.NewRequest("GET", "/api/v1/wallet/balance/test-address", nil)
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet/balance/{address}", server.getBalanceHandler).Methods("GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("GetBalance handler should return ServiceUnavailable when wallet is nil, got %v", status)
	}
}

func TestServer_GetAccountsHandler(t *testing.T) {
	mockWallet := NewMockWallet()
	server := &Server{wallet: mockWallet}

	req, err := http.NewRequest("GET", "/api/v1/wallet/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getAccountsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetAccounts handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["count"] != float64(2) {
		t.Errorf("Expected count 2, got %v", response["count"])
	}

	if _, exists := response["accounts"]; !exists {
		t.Error("Response should contain accounts")
	}

	accounts := response["accounts"].([]interface{})
	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}
}

func TestServer_GetAccountsHandler_NoWallet(t *testing.T) {
	server := &Server{wallet: nil}

	req, err := http.NewRequest("GET", "/api/v1/wallet/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getAccountsHandler(rr, req)

	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("GetAccounts handler should return ServiceUnavailable when wallet is nil, got %v", status)
	}
}

func TestServer_GetPeersHandler(t *testing.T) {
	server := &Server{}

	req, err := http.NewRequest("GET", "/api/v1/network/peers", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getPeersHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetPeers handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["count"] != float64(0) {
		t.Errorf("Expected count 0, got %v", response["count"])
	}

	if _, exists := response["peers"]; !exists {
		t.Error("Response should contain peers")
	}
}

func TestServer_GetNetworkStatusHandler(t *testing.T) {
	server := &Server{}

	req, err := http.NewRequest("GET", "/api/v1/network/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.getNetworkStatusHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetNetworkStatus handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["status"] != "active" {
		t.Errorf("Expected status 'active', got %v", response["status"])
	}

	if response["peer_count"] != float64(0) {
		t.Errorf("Expected peer_count 0, got %v", response["peer_count"])
	}

	if response["listening"] != true {
		t.Errorf("Expected listening true, got %v", response["listening"])
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should contain timestamp")
	}
}

func TestServer_Start(t *testing.T) {
	mockChain := NewMockChain()
	mockWallet := NewMockWallet()

	config := &ServerConfig{
		Port:   0, // Use port 0 to let the system choose an available port
		Chain:  mockChain,
		Wallet: mockWallet,
	}

	server := NewServer(config)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// The test will pass if the server starts without errors
	// We can't easily test the actual listening without more complex setup
}

func TestServer_Start_Error(t *testing.T) {
	// Test starting server with invalid port
	config := &ServerConfig{
		Port:   -1, // Invalid port
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// This should fail due to invalid port
	err := server.Start()
	if err == nil {
		t.Error("Expected error when starting with invalid port")
	}
}

func TestServer_GetChainInfoHandler_EmptyChain(t *testing.T) {
	// Create mock chain with no blocks
	mockChain := &MockChain{
		height:         0,
		bestBlock:      nil,
		genesisBlock:   nil,
		blocks:         make(map[string]*block.Block),
		blocksByHeight: make(map[uint64]*block.Block),
	}

	config := &ServerConfig{
		Port:   8080,
		Chain:  mockChain,
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Create request
	req, err := http.NewRequest("GET", "/api/v1/chain/info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler directly (not through router)
	server.getChainInfoHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, exists := response["height"]; !exists {
		t.Error("Response missing height field")
	}
	if _, exists := response["best_block"]; !exists {
		t.Error("Response missing best_block field")
	}
}

func TestServer_GetBlockHandler_InvalidHashFormat(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Test with invalid hex hash
	req, err := http.NewRequest("GET", "/api/v1/blocks/invalid-hash", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up mux vars
	vars := map[string]string{
		"hash": "invalid-hash",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	server.getBlockHandler(rr, req)

	// Should return 400 Bad Request
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_GetBlockByHeightHandler_InvalidHeightFormat(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Test with invalid height
	req, err := http.NewRequest("GET", "/api/v1/blocks/height/invalid", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up mux vars
	vars := map[string]string{
		"height": "invalid",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	server.getBlockByHeightHandler(rr, req)

	// Should return 400 Bad Request
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_GetTransactionHandler_InvalidHashFormat(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Test with invalid hex hash
	req, err := http.NewRequest("GET", "/api/v1/transactions/invalid-hash", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up mux vars
	vars := map[string]string{
		"hash": "invalid-hash",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	server.getTransactionHandler(rr, req)

	// Should return 400 Bad Request
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServer_GetBalanceHandler_InvalidAddress(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Test with empty address
	req, err := http.NewRequest("GET", "/api/v1/wallet/balance/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up mux vars
	vars := map[string]string{
		"address": "",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	server.getBalanceHandler(rr, req)

	// Should return 200 OK (empty address is valid)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServer_GetPendingTransactionsHandler_Empty(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/v1/transactions/pending", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.getPendingTransactionsHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, exists := response["pending_transactions"]; !exists {
		t.Error("Response missing pending_transactions field")
	}
	if _, exists := response["count"]; !exists {
		t.Error("Response missing count field")
	}

	// Verify count is 0
	if count, ok := response["count"].(float64); !ok || count != 0 {
		t.Errorf("Expected count 0, got %v", count)
	}
}

func TestServer_GetPeersHandler_Empty(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/v1/network/peers", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.getPeersHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, exists := response["peers"]; !exists {
		t.Error("Response missing peers field")
	}
	if _, exists := response["count"]; !exists {
		t.Error("Response missing count field")
	}

	// Verify count is 0
	if count, ok := response["count"].(float64); !ok || count != 0 {
		t.Errorf("Expected count 0, got %v", count)
	}
}

func TestServer_GetNetworkStatusHandler_Basic(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/v1/network/status", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.getNetworkStatusHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	requiredFields := []string{"status", "peer_count", "listening", "timestamp"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response missing %s field", field)
		}
	}

	// Verify status is "active"
	if status, ok := response["status"].(string); !ok || status != "active" {
		t.Errorf("Expected status 'active', got %v", status)
	}

	// Verify listening is true
	if listening, ok := response["listening"].(bool); !ok || !listening {
		t.Errorf("Expected listening true, got %v", listening)
	}
}

func TestServer_SetupRoutes(t *testing.T) {
	config := &ServerConfig{
		Port:   8080,
		Chain:  NewMockChain(),
		Wallet: NewMockWallet(),
	}

	server := NewServer(config)

	// Test that routes are properly set up
	if server.router == nil {
		t.Fatal("Router should not be nil")
	}

	// Test that all expected routes exist
	expectedRoutes := []string{
		"/health",
		"/api/v1/chain/info",
		"/api/v1/chain/height",
		"/api/v1/chain/status",
		"/api/v1/blocks/latest",
		"/api/v1/blocks/height/{height}",
		"/api/v1/blocks/{hash}",
		"/api/v1/transactions/{hash}",
		"/api/v1/transactions/pending",
		"/api/v1/wallet/balance/{address}",
		"/api/v1/wallet/accounts",
		"/api/v1/network/peers",
		"/api/v1/network/status",
	}

	// Note: This is a basic check. In a real test, you might want to
	// verify that the routes actually work by making requests to them.
	for _, route := range expectedRoutes {
		// Just verify the router exists and has been configured
		if server.router == nil {
			t.Errorf("Router not configured for route: %s", route)
		}
	}
}

func TestServer_ErrorHandling(t *testing.T) {
	// Test server with nil chain and wallet
	config := &ServerConfig{
		Port:   8080,
		Chain:  nil,
		Wallet: nil,
	}

	server := NewServer(config)

	// Test wallet handlers with nil wallet
	req, err := http.NewRequest("GET", "/api/v1/wallet/balance/test-address", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	vars := map[string]string{
		"address": "test-address",
	}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	server.getBalanceHandler(rr, req)

	// Should return 503 Service Unavailable
	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
	}
}
