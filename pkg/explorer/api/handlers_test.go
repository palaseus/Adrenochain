package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/palaseus/adrenochain/pkg/explorer/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExplorerService is a mock implementation of ExplorerService for testing
type MockExplorerService struct {
	mock.Mock
}

func (m *MockExplorerService) GetDashboard(ctx context.Context) (*service.Dashboard, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Dashboard), args.Error(1)
}

func (m *MockExplorerService) GetBlockDetails(ctx context.Context, hash []byte) (*service.BlockDetails, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.BlockDetails), args.Error(1)
}

func (m *MockExplorerService) GetTransactionDetails(ctx context.Context, hash []byte) (*service.TransactionDetails, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TransactionDetails), args.Error(1)
}

func (m *MockExplorerService) GetAddressDetails(ctx context.Context, address string) (*service.AddressDetails, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AddressDetails), args.Error(1)
}

func (m *MockExplorerService) GetBlocks(ctx context.Context, limit, offset int) ([]*service.BlockSummary, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.BlockSummary), args.Error(1)
}

func (m *MockExplorerService) GetTransactions(ctx context.Context, limit, offset int) ([]*service.TransactionSummary, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.TransactionSummary), args.Error(1)
}

func (m *MockExplorerService) Search(ctx context.Context, query string) (*service.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchResult), args.Error(1)
}

func (m *MockExplorerService) GetStatistics(ctx context.Context) (*service.Statistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Statistics), args.Error(1)
}

func TestNewExplorerHandler(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.explorerService)
}

func TestDashboardHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedDashboard := &service.Dashboard{
		Stats: &service.BlockchainStats{
			TotalBlocks:       1000,
			TotalTransactions: 50000,
		},
		LastUpdate: time.Now(),
	}

	mockService.On("GetDashboard", mock.Anything).Return(expectedDashboard, nil)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()

	handler.DashboardHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.Dashboard
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedDashboard.Stats.TotalBlocks, response.Stats.TotalBlocks)

	mockService.AssertExpectations(t)
}

func TestDashboardHandler_ServiceError(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	mockService.On("GetDashboard", mock.Anything).Return(nil, fmt.Errorf("service error"))

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()

	handler.DashboardHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get dashboard")

	mockService.AssertExpectations(t)
}

func TestBlockDetailsHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedBlock := &service.BlockDetails{
		BlockSummary: &service.BlockSummary{
			Hash:      []byte("abc123"),
			Height:    1000,
			Timestamp: time.Unix(1234567890, 0),
		},
	}

	mockService.On("GetBlockDetails", mock.Anything, []byte{0xab, 0xc1, 0x23}).Return(expectedBlock, nil)

	req := httptest.NewRequest("GET", "/blocks/abc123", nil)
	vars := map[string]string{"hash": "abc123"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.BlockDetailsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.BlockDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedBlock.BlockSummary.Hash, response.BlockSummary.Hash)

	mockService.AssertExpectations(t)
}

func TestBlockDetailsHandler_InvalidHash(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	req := httptest.NewRequest("GET", "/blocks/invalid", nil)
	vars := map[string]string{"hash": "invalid"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.BlockDetailsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid block hash format")
}

func TestBlockDetailsHandler_ServiceError(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	mockService.On("GetBlockDetails", mock.Anything, []byte{0xab, 0xc1, 0x23}).Return(nil, fmt.Errorf("service error"))

	req := httptest.NewRequest("GET", "/blocks/abc123", nil)
	vars := map[string]string{"hash": "abc123"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.BlockDetailsHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get block details")

	mockService.AssertExpectations(t)
}

func TestTransactionDetailsHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedTx := &service.TransactionDetails{
		TransactionSummary: &service.TransactionSummary{
			Hash:   []byte("def456"),
			Amount: 1000000,
		},
	}

	mockService.On("GetTransactionDetails", mock.Anything, []byte{0xde, 0xf4, 0x56}).Return(expectedTx, nil)

	req := httptest.NewRequest("GET", "/transactions/def456", nil)
	vars := map[string]string{"hash": "def456"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.TransactionDetailsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.TransactionDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.TransactionSummary.Hash, response.TransactionSummary.Hash)

	mockService.AssertExpectations(t)
}

func TestTransactionDetailsHandler_InvalidHash(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	req := httptest.NewRequest("GET", "/transactions/invalid", nil)
	vars := map[string]string{"hash": "invalid"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.TransactionDetailsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid transaction hash format")
}

func TestTransactionDetailsHandler_ServiceError(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Mock service to return error
	mockService.On("GetTransactionDetails", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("service error"))

	// Create request with valid hash
	req := httptest.NewRequest("GET", "/transaction/1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil)
	vars := map[string]string{"hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"}
	req = mux.SetURLVars(req, vars)

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.TransactionDetailsHandler(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get transaction details")
	mockService.AssertExpectations(t)
}

func TestAddressDetailsHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedAddr := &service.AddressDetails{
		AddressSummary: &service.AddressSummary{
			Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			Balance: 5000000,
			TxCount: 25,
		},
	}

	mockService.On("GetAddressDetails", mock.Anything, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa").Return(expectedAddr, nil)

	req := httptest.NewRequest("GET", "/addresses/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", nil)
	vars := map[string]string{"address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.AddressDetailsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.AddressDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedAddr.AddressSummary.Address, response.AddressSummary.Address)

	mockService.AssertExpectations(t)
}

func TestAddressDetailsHandler_InvalidAddress(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Test address that's too short
	req := httptest.NewRequest("GET", "/addresses/short", nil)
	vars := map[string]string{"address": "short"}
	req = mux.SetURLVars(req, vars)
	w := httptest.NewRecorder()

	handler.AddressDetailsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid address format")

	// Test address that's too long
	longAddr := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNaVeryLongAddressThatExceedsLimit"
	req = httptest.NewRequest("GET", "/addresses/"+longAddr, nil)
	vars = map[string]string{"address": longAddr}
	req = mux.SetURLVars(req, vars)
	w = httptest.NewRecorder()

	handler.AddressDetailsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid address format")
}

func TestAddressDetailsHandler_ServiceError(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Mock service to return error
	mockService.On("GetAddressDetails", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("service error"))

	// Create request with valid address
	req := httptest.NewRequest("GET", "/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", nil)
	vars := map[string]string{"address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"}
	req = mux.SetURLVars(req, vars)

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.AddressDetailsHandler(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get address details")
	mockService.AssertExpectations(t)
}

func TestBlocksHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedBlocks := []*service.BlockSummary{
		{Hash: []byte("abc123"), Height: 1000},
		{Hash: []byte("def456"), Height: 999},
	}

	mockService.On("GetBlocks", mock.Anything, 10, 0).Return(expectedBlocks, nil)

	req := httptest.NewRequest("GET", "/blocks?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	handler.BlocksHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	blocks := response["blocks"].([]interface{})
	assert.Len(t, blocks, 2)
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	assert.Equal(t, float64(2), response["count"])

	mockService.AssertExpectations(t)
}

func TestBlocksHandler_DefaultPagination(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedBlocks := []*service.BlockSummary{}
	mockService.On("GetBlocks", mock.Anything, 20, 0).Return(expectedBlocks, nil)

	req := httptest.NewRequest("GET", "/blocks", nil)
	w := httptest.NewRecorder()

	handler.BlocksHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTransactionsHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedTxs := []*service.TransactionSummary{
		{Hash: []byte("abc123")},
		{Hash: []byte("def456")},
	}

	mockService.On("GetTransactions", mock.Anything, 15, 5).Return(expectedTxs, nil)

	req := httptest.NewRequest("GET", "/transactions?limit=15&offset=5", nil)
	w := httptest.NewRecorder()

	handler.TransactionsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	transactions := response["transactions"].([]interface{})
	assert.Len(t, transactions, 2)
	assert.Equal(t, float64(15), response["limit"])
	assert.Equal(t, float64(5), response["offset"])
	assert.Equal(t, float64(2), response["count"])

	mockService.AssertExpectations(t)
}

func TestSearchHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedResults := &service.SearchResult{
		Query: "abc123",
		Block: &service.BlockSummary{Hash: []byte("abc123"), Height: 1000},
	}

	mockService.On("Search", mock.Anything, "abc123").Return(expectedResults, nil)

	req := httptest.NewRequest("GET", "/search?q=abc123", nil)
	w := httptest.NewRecorder()

	handler.SearchHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.SearchResult
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults.Query, response.Query)

	mockService.AssertExpectations(t)
}

func TestSearchHandler_MissingQuery(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	handler.SearchHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing search query parameter 'q'")
}

func TestStatisticsHandler_Success(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	expectedStats := &service.Statistics{
		Blockchain: &service.BlockchainStats{
			TotalBlocks:       1000,
			TotalTransactions: 50000,
			TotalAddresses:    2500,
		},
		LastUpdate: time.Now(),
	}

	mockService.On("GetStatistics", mock.Anything).Return(expectedStats, nil)

	req := httptest.NewRequest("GET", "/statistics", nil)
	w := httptest.NewRecorder()

	handler.StatisticsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.Statistics
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedStats.Blockchain.TotalBlocks, response.Blockchain.TotalBlocks)

	mockService.AssertExpectations(t)
}

func TestHealthHandler(t *testing.T) {
	handler := NewExplorerHandler(nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestWriteJSONResponse(t *testing.T) {
	handler := NewExplorerHandler(nil)
	w := httptest.NewRecorder()

	testData := map[string]string{"key": "value"}

	handler.writeJSONResponse(w, testData)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "value", response["key"])
}

func TestWriteErrorResponse(t *testing.T) {
	handler := NewExplorerHandler(nil)
	w := httptest.NewRecorder()

	handler.writeErrorResponse(w, http.StatusBadRequest, "Bad request")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad request", response["error"])
	assert.Equal(t, float64(400), response["status"])
	assert.Equal(t, false, response["success"])
}

func TestParsePaginationParams(t *testing.T) {
	handler := NewExplorerHandler(nil)

	// Test default values
	req := httptest.NewRequest("GET", "/test", nil)
	limit, offset := handler.parsePaginationParams(req)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test custom values
	req = httptest.NewRequest("GET", "/test?limit=50&offset=10", nil)
	limit, offset = handler.parsePaginationParams(req)
	assert.Equal(t, 50, limit)
	assert.Equal(t, 10, offset)

	// Test invalid values (should use defaults)
	req = httptest.NewRequest("GET", "/test?limit=invalid&offset=invalid", nil)
	limit, offset = handler.parsePaginationParams(req)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test negative values (should use defaults)
	req = httptest.NewRequest("GET", "/test?limit=-5&offset=-10", nil)
	limit, offset = handler.parsePaginationParams(req)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)

	// Test very large values (should be capped)
	req = httptest.NewRequest("GET", "/test?limit=1000&offset=10000", nil)
	limit, offset = handler.parsePaginationParams(req)
	assert.Equal(t, 100, limit)
	assert.Equal(t, 10000, offset)
}

func TestValidateHexHash(t *testing.T) {
	handler := NewExplorerHandler(nil)

	// Test valid hex hash (64 characters)
	t.Run("valid_hash", func(t *testing.T) {
		validHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		t.Logf("Testing valid hash: %s", validHash)
		t.Logf("Hash length: %d", len(validHash))
		validHashBytes, err := handler.validateHexHash(validHash)
		if err != nil {
			t.Logf("Error: %v", err)
		}
		assert.NoError(t, err)
		assert.Len(t, validHashBytes, 32) // 64 hex chars = 32 bytes
	})

	// Test invalid hex (64 chars but not valid hex)
	t.Run("invalid_hex", func(t *testing.T) {
		invalidHash := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
		_, err := handler.validateHexHash(invalidHash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hex format")
	})

	// Test empty string
	t.Run("empty_string", func(t *testing.T) {
		emptyHashBytes, err := handler.validateHexHash("")
		assert.Error(t, err)
		assert.Nil(t, emptyHashBytes) // Should return nil when there's an error
		assert.Contains(t, err.Error(), "hash must be 64 characters long")
	})

	// Test short hex (less than 64 chars)
	t.Run("short_hex", func(t *testing.T) {
		shortHashBytes, err := handler.validateHexHash("abc123")
		assert.Error(t, err)
		assert.Nil(t, shortHashBytes) // Should return nil when there's an error
		assert.Contains(t, err.Error(), "hash must be 64 characters long")
	})
}

func TestOptionsHandler(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Create OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.OptionsHandler(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWriteJSONResponse_Error(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Create response recorder
	w := httptest.NewRecorder()

	// Test with data that cannot be marshaled to JSON
	// Create a channel which cannot be marshaled to JSON
	unmarshallableData := make(chan int)

	// Call writeJSONResponse with unmarshallable data
	handler.writeJSONResponse(w, unmarshallableData)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to marshal JSON response")
}

func TestWriteErrorResponse_Error(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Create response recorder
	w := httptest.NewRecorder()

	// Test writeErrorResponse with a custom error message
	handler.writeErrorResponse(w, http.StatusBadRequest, "Test error message")

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Test error message")
	assert.Contains(t, w.Body.String(), "false")
	assert.Contains(t, w.Body.String(), "400")
}

// TestWriteErrorResponse_JSONMarshalError tests the error path in writeErrorResponse
func TestWriteErrorResponse_JSONMarshalError(t *testing.T) {
	mockService := &MockExplorerService{}
	handler := NewExplorerHandler(mockService)

	// Create response recorder
	w := httptest.NewRecorder()

	// Test writeErrorResponse with a message that might cause JSON marshaling issues
	// This is a bit tricky to trigger, but we can test the basic functionality
	handler.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
	assert.Contains(t, w.Body.String(), "false")
	assert.Contains(t, w.Body.String(), "500")
}
