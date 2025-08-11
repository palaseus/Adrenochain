package data

import (
	"testing"

	"github.com/gochain/gochain/pkg/explorer/service"
)

// MockSearchProvider implements SearchProvider interface for testing
type MockSearchProvider struct {
	searchResults map[string]*service.SearchResult
}

func NewMockSearchProvider() *MockSearchProvider {
	return &MockSearchProvider{
		searchResults: make(map[string]*service.SearchResult),
	}
}

func (msp *MockSearchProvider) Search(query string) (*service.SearchResult, error) {
	if result, exists := msp.searchResults[query]; exists {
		return result, nil
	}
	return &service.SearchResult{}, nil
}

func (msp *MockSearchProvider) SearchBlocks(query string, limit, offset int) ([]*service.BlockSummary, error) {
	return []*service.BlockSummary{}, nil
}

func (msp *MockSearchProvider) SearchTransactions(query string, limit, offset int) ([]*service.TransactionSummary, error) {
	return []*service.TransactionSummary{}, nil
}

func (msp *MockSearchProvider) SearchAddresses(query string, limit, offset int) ([]*service.AddressSummary, error) {
	return []*service.AddressSummary{}, nil
}

func (msp *MockSearchProvider) AddSearchResult(query string, result *service.SearchResult) {
	msp.searchResults[query] = result
}

func (msp *MockSearchProvider) ClearSearchResults() {
	msp.searchResults = make(map[string]*service.SearchResult)
}

func TestSearchProvider_Search(t *testing.T) {
	provider := NewMockSearchProvider()
	
	// Test searching with no results
	result, err := provider.Search("non-existent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}
	
	// Test searching with results
	query := "test-query"
	expectedResult := &service.SearchResult{
		Query: query,
		Type:  "block",
		Block: &service.BlockSummary{
			Hash:   []byte("abc123"),
			Height: uint64(123),
		},
	}
	
	provider.AddSearchResult(query, expectedResult)
	
	result, err = provider.Search(query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result to be returned")
	}
	
	if result.Type != expectedResult.Type {
		t.Errorf("Expected type %s, got %s", expectedResult.Type, result.Type)
	}
	if result.Block == nil {
		t.Error("Expected block to be present")
	}
}

func TestSearchProvider_AddSearchResult(t *testing.T) {
	provider := NewMockSearchProvider()
	
	// Test adding single result
	query := "single-result"
	result := &service.SearchResult{
		Query: query,
		Type:  "address",
		Address: &service.AddressSummary{
			Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			Balance: uint64(1000),
		},
	}
	
	provider.AddSearchResult(query, result)
	
	// Verify result was added
	retrievedResult, err := provider.Search(query)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrievedResult == nil {
		t.Fatal("Expected result to be returned")
	}
	
	if retrievedResult.Type != result.Type {
		t.Errorf("Expected type %s, got %s", result.Type, retrievedResult.Type)
	}
}

func TestSearchProvider_ClearSearchResults(t *testing.T) {
	provider := NewMockSearchProvider()
	
	// Add some results
	provider.AddSearchResult("query1", &service.SearchResult{Type: "block"})
	provider.AddSearchResult("query2", &service.SearchResult{Type: "transaction"})
	
	// Verify results exist
	result1, _ := provider.Search("query1")
	result2, _ := provider.Search("query2")
	
	if result1.Type != "block" || result2.Type != "transaction" {
		t.Error("Expected results to exist before clearing")
	}
	
	// Clear results
	provider.ClearSearchResults()
	
	// Verify results are gone
	result1, _ = provider.Search("query1")
	result2, _ = provider.Search("query2")
	
	if result1.Type != "" || result2.Type != "" {
		t.Error("Expected results to be cleared")
	}
}

func TestSearchProvider_EdgeCases(t *testing.T) {
	provider := NewMockSearchProvider()
	
	// Test empty query
	result, err := provider.Search("")
	if err != nil {
		t.Fatalf("Expected no error for empty query, got %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result for empty query")
	}
	
	// Test very long query
	longQuery := string(make([]byte, 1000))
	result, err = provider.Search(longQuery)
	if err != nil {
		t.Fatalf("Expected no error for long query, got %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result for long query")
	}
	
	// Test adding nil result
	provider.AddSearchResult("nil-result", nil)
	result, err = provider.Search("nil-result")
	if err != nil {
		t.Fatalf("Expected no error for nil result, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result to be returned as nil")
	}
}
