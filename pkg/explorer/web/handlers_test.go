package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/explorer/service"
	"github.com/gorilla/mux"
)

// MockExplorerService implements service.ExplorerService for testing
type MockExplorerService struct {
	dashboard *service.Dashboard
	blocks    []*service.BlockSummary
	txs       []*service.TransactionSummary
	block     *service.BlockDetails
	tx        *service.TransactionDetails
	address   *service.AddressDetails
	stats     *service.Statistics
	search    *service.SearchResult
	err       error
}

func (m *MockExplorerService) GetDashboard(ctx context.Context) (*service.Dashboard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.dashboard, nil
}

func (m *MockExplorerService) GetBlockDetails(ctx context.Context, hash []byte) (*service.BlockDetails, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.block, nil
}

func (m *MockExplorerService) GetTransactionDetails(ctx context.Context, hash []byte) (*service.TransactionDetails, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tx, nil
}

func (m *MockExplorerService) GetAddressDetails(ctx context.Context, address string) (*service.AddressDetails, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.address, nil
}

func (m *MockExplorerService) GetBlocks(ctx context.Context, limit, offset int) ([]*service.BlockSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.blocks, nil
}

func (m *MockExplorerService) GetTransactions(ctx context.Context, limit, offset int) ([]*service.TransactionSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.txs, nil
}

func (m *MockExplorerService) Search(ctx context.Context, query string) (*service.SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.search, nil
}

func (m *MockExplorerService) GetStatistics(ctx context.Context) (*service.Statistics, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func TestNewWebHandler(t *testing.T) {
	mockService := &MockExplorerService{}
	templates := NewTemplates()

	handler := NewWebHandler(mockService, templates)

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}

	if handler.explorerService != mockService {
		t.Error("Expected explorer service to be set")
	}

	if handler.templates != templates {
		t.Error("Expected templates to be set")
	}
}

func TestHomeHandler(t *testing.T) {
	mockService := &MockExplorerService{
		dashboard: &service.Dashboard{
			Stats: &service.BlockchainStats{
				TotalBlocks:       100,
				TotalTransactions: 1000,
				TotalAddresses:    500,
				Difficulty:        12345,
			},
			RecentBlocks: []*service.BlockSummary{
				{
					Hash:          []byte("test-hash"),
					Height:        100,
					Timestamp:     time.Unix(1234567890, 0),
					TxCount:       5,
					Size:          1000,
					Difficulty:    12345,
					Confirmations: 0,
				},
			},
			RecentTxs: []*service.TransactionSummary{
				{
					Hash:      []byte("test-tx-hash"),
					BlockHash: []byte("test-hash"),
					Height:    100,
					Timestamp: time.Unix(1234567890, 0),
					Inputs:    2,
					Outputs:   1,
					Amount:    100000000,
					Fee:       1000,
					Status:    "confirmed",
				},
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HomeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "GoChain Blockchain Explorer") {
		t.Error("Expected page title in response")
	}

	if !strings.Contains(body, "100") {
		t.Error("Expected total blocks in response")
	}
}

func TestBlockListHandler(t *testing.T) {
	mockService := &MockExplorerService{
		blocks: []*service.BlockSummary{
			{
				Hash:          []byte("test-hash"),
				Height:        100,
				Timestamp:     time.Unix(1234567890, 0),
				TxCount:       5,
				Size:          1000,
				Difficulty:    12345,
				Confirmations: 0,
			},
		},
		stats: &service.Statistics{
			Blockchain: &service.BlockchainStats{
				TotalBlocks: 100,
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/blocks", nil)
	w := httptest.NewRecorder()

	handler.BlockListHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Blocks") {
		t.Error("Expected page title in response")
	}

	if !strings.Contains(body, "Block #100") {
		t.Error("Expected block height in response")
	}
}

func TestBlockDetailHandler(t *testing.T) {
	mockService := &MockExplorerService{
		block: &service.BlockDetails{
			BlockSummary: &service.BlockSummary{
				Hash:          []byte("test-hash"),
				Height:        100,
				Timestamp:     time.Unix(1234567890, 0),
				TxCount:       5,
				Size:          1000,
				Difficulty:    12345,
				Confirmations: 0,
			},
			PrevHash:   []byte("prev-hash"),
			NextHash:   []byte("next-hash"),
			MerkleRoot: []byte("merkle-root"),
			Nonce:      12345,
			Version:    1,
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/blocks/746573742d68617368", nil)
	w := httptest.NewRecorder()

	// Set up router to extract URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/blocks/{hash}", handler.BlockDetailHandler)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Block #100") {
		t.Error("Expected block height in response")
	}
}

func TestTransactionListHandler(t *testing.T) {
	mockService := &MockExplorerService{
		txs: []*service.TransactionSummary{
			{
				Hash:      []byte("test-tx-hash"),
				BlockHash: []byte("test-hash"),
				Height:    100,
				Timestamp: time.Unix(1234567890, 0),
				Inputs:    2,
				Outputs:   1,
				Amount:    100000000,
				Fee:       1000,
				Status:    "confirmed",
			},
		},
		stats: &service.Statistics{
			Blockchain: &service.BlockchainStats{
				TotalTransactions: 1000,
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	handler.TransactionListHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Transactions") {
		t.Error("Expected page title in response")
	}
}

func TestTransactionDetailHandler(t *testing.T) {
	mockService := &MockExplorerService{
		tx: &service.TransactionDetails{
			TransactionSummary: &service.TransactionSummary{
				Hash:      []byte("test-tx-hash"),
				BlockHash: []byte("test-hash"),
				Height:    100,
				Timestamp: time.Unix(1234567890, 0),
				Inputs:    2,
				Outputs:   1,
				Amount:    100000000,
				Fee:       1000,
				Status:    "confirmed",
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/transactions/746573742d74782d68617368", nil)
	w := httptest.NewRecorder()

	// Set up router to extract URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/transactions/{hash}", handler.TransactionDetailHandler)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Transaction") {
		t.Error("Expected page title in response")
	}
}

func TestAddressDetailHandler(t *testing.T) {
	mockService := &MockExplorerService{
		address: &service.AddressDetails{
			AddressSummary: &service.AddressSummary{
				Address:   "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
				Balance:   100000000,
				TxCount:   5,
				FirstSeen: time.Unix(1234567890, 0),
				LastSeen:  time.Unix(1234567890, 0),
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/addresses/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", nil)
	w := httptest.NewRecorder()

	// Set up router to extract URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/addresses/{address}", handler.AddressDetailHandler)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Address") {
		t.Error("Expected page title in response")
	}
}

func TestSearchHandler(t *testing.T) {
	mockService := &MockExplorerService{
		search: &service.SearchResult{
			Query: "test-query",
			Type:  "block",
			Block: &service.BlockSummary{
				Hash:   []byte("test-hash"),
				Height: 100,
			},
		},
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	// Test search form display
	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	handler.SearchHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Search") {
		t.Error("Expected search page title in response")
	}

	// Test search with query
	req = httptest.NewRequest("GET", "/search?q=test-query", nil)
	w = httptest.NewRecorder()

	handler.SearchHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body = w.Body.String()
	if !strings.Contains(body, "test-query") {
		t.Error("Expected search query in response")
	}
}

func TestErrorHandling(t *testing.T) {
	mockService := &MockExplorerService{
		err: errors.New("block not found"),
	}

	templates := NewTemplates()
	handler := NewWebHandler(mockService, templates)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HomeHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Error") {
		t.Error("Expected error page title in response")
	}
}

func TestPaginationHelpers(t *testing.T) {
	handler := &WebHandler{}

	// Test pagination creation
	pagination := handler.createPagination(20, 40, 100)

	if pagination["CurrentPage"] != 3 {
		t.Errorf("Expected current page 3, got %v", pagination["CurrentPage"])
	}

	if pagination["TotalPages"] != 5 {
		t.Errorf("Expected total pages 5, got %v", pagination["TotalPages"])
	}

	if !pagination["HasPrev"].(bool) {
		t.Error("Expected has previous to be true")
	}

	if !pagination["HasNext"].(bool) {
		t.Error("Expected has next to be true")
	}
}

func TestParsePaginationParams(t *testing.T) {
	handler := &WebHandler{}

	// Test default values
	req := httptest.NewRequest("GET", "/", nil)
	limit, offset := handler.parsePaginationParams(req)

	if limit != 20 {
		t.Errorf("Expected default limit 20, got %d", limit)
	}

	if offset != 0 {
		t.Errorf("Expected default offset 0, got %d", offset)
	}

	// Test custom values
	req = httptest.NewRequest("GET", "/?limit=50&offset=100", nil)
	limit, offset = handler.parsePaginationParams(req)

	if limit != 50 {
		t.Errorf("Expected limit 50, got %d", limit)
	}

	if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}

	// Test invalid values
	req = httptest.NewRequest("GET", "/?limit=invalid&offset=-10", nil)
	limit, offset = handler.parsePaginationParams(req)

	if limit != 20 {
		t.Errorf("Expected default limit 20 for invalid input, got %d", limit)
	}

	if offset != 0 {
		t.Errorf("Expected default offset 0 for invalid input, got %d", offset)
	}
}

func TestWebServer(t *testing.T) {
	mockService := &MockExplorerService{}
	webServer := NewWebServer(mockService)

	if webServer == nil {
		t.Fatal("Expected web server to be created")
	}

	// Test health check
	err := webServer.HealthCheck()
	if err != nil {
		t.Errorf("Expected health check to pass, got error: %v", err)
	}

	// Test server info
	info := webServer.GetServerInfo()
	if info["type"] != "web" {
		t.Error("Expected server type to be 'web'")
	}

	if info["status"] != "running" {
		t.Error("Expected server status to be 'running'")
	}

	// Test handler and templates access
	if webServer.GetHandler() == nil {
		t.Error("Expected handler to be accessible")
	}

	if webServer.GetTemplates() == nil {
		t.Error("Expected templates to be accessible")
	}
}
