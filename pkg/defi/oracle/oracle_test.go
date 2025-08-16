package oracle

import (
	"context"
	"math/big"
	"testing"
	"time"
)

// TestNewOracleAggregator tests oracle aggregator creation
func TestNewOracleAggregator(t *testing.T) {
	config := DefaultOracleConfig()
	aggregator := NewOracleAggregator(config)

	if aggregator == nil {
		t.Fatal("expected non-nil aggregator")
	}

	if aggregator.config.MinProviders != config.MinProviders {
		t.Errorf("expected MinProviders %d, got %d", config.MinProviders, aggregator.config.MinProviders)
	}

	if aggregator.config.MaxPriceAge != config.MaxPriceAge {
		t.Errorf("expected MaxPriceAge %v, got %v", config.MaxPriceAge, aggregator.config.MaxPriceAge)
	}
}

// TestOracleAggregatorAddProvider tests adding oracle providers
func TestOracleAggregatorAddProvider(t *testing.T) {
	aggregator := NewOracleAggregator(DefaultOracleConfig())

	provider1 := NewMockOracleProvider("provider1", "Test Provider 1", "http://test1.com", 0.95)
	provider2 := NewMockOracleProvider("provider2", "Test Provider 2", "http://test2.com", 0.90)

	// Test adding providers
	err := aggregator.AddProvider("provider1", provider1, 1.0)
	if err != nil {
		t.Errorf("unexpected error adding provider1: %v", err)
	}

	err = aggregator.AddProvider("provider2", provider2, 0.8)
	if err != nil {
		t.Errorf("unexpected error adding provider2: %v", err)
	}

	// Test adding provider with invalid weight
	err = aggregator.AddProvider("invalid", provider1, 0)
	if err == nil {
		t.Error("expected error for zero weight")
	}

	// Test adding nil provider
	err = aggregator.AddProvider("nil", nil, 1.0)
	if err == nil {
		t.Error("expected error for nil provider")
	}

	// Check providers were added
	providers := aggregator.GetProviders()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}

// TestOracleAggregatorRemoveProvider tests removing oracle providers
func TestOracleAggregatorRemoveProvider(t *testing.T) {
	aggregator := NewOracleAggregator(DefaultOracleConfig())

	provider := NewMockOracleProvider("test", "Test Provider", "http://test.com", 0.95)

	// Add provider first
	err := aggregator.AddProvider("test", provider, 1.0)
	if err != nil {
		t.Fatalf("failed to add provider: %v", err)
	}

	// Test removing provider
	err = aggregator.RemoveProvider("test")
	if err != nil {
		t.Errorf("unexpected error removing provider: %v", err)
	}

	// Test removing non-existent provider
	err = aggregator.RemoveProvider("nonexistent")
	if err == nil {
		t.Error("expected error removing non-existent provider")
	}

	// Check provider was removed
	providers := aggregator.GetProviders()
	if len(providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(providers))
	}
}

// TestOracleAggregatorGetAggregatedPrice tests price aggregation
func TestOracleAggregatorGetAggregatedPrice(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 2
	aggregator := NewOracleAggregator(config)

	// Add providers
	provider1 := NewMockOracleProvider("provider1", "Test Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Test Provider 2", "http://test2.com", 1.0)

	err := aggregator.AddProvider("provider1", provider1, 1.0)
	if err != nil {
		t.Fatalf("failed to add provider1: %v", err)
	}

	err = aggregator.AddProvider("provider2", provider2, 1.0)
	if err != nil {
		t.Fatalf("failed to add provider2: %v", err)
	}

	// Test getting aggregated price
	ctx := context.Background()
	price, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err != nil {
		t.Errorf("unexpected error getting price: %v", err)
	}

	if price == nil {
		t.Fatal("expected non-nil price")
	}

	if price.Asset != "BTC" {
		t.Errorf("expected asset BTC, got %s", price.Asset)
	}

	if price.Providers != 2 {
		t.Errorf("expected 2 providers, got %d", price.Providers)
	}

	if price.Confidence < 80 {
		t.Errorf("expected confidence >= 80, got %d", price.Confidence)
	}
}

// TestOracleAggregatorInsufficientProviders tests behavior with insufficient providers
func TestOracleAggregatorInsufficientProviders(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 3
	aggregator := NewOracleAggregator(config)

	// Add only 2 providers when 3 are required
	provider1 := NewMockOracleProvider("provider1", "Test Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Test Provider 2", "http://test2.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 1.0)
	aggregator.AddProvider("provider2", provider2, 1.0)

	// Test getting price with insufficient providers
	ctx := context.Background()
	_, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err == nil {
		t.Error("expected error for insufficient providers")
	}
}

// TestOracleAggregatorPriceValidation tests price data validation
func TestOracleAggregatorPriceValidation(t *testing.T) {
	config := DefaultOracleConfig()
	config.ConfidenceThreshold = 90
	aggregator := NewOracleAggregator(config)

	// Add provider with low confidence
	provider := NewMockOracleProvider("low_confidence", "Low Confidence Provider", "http://test.com", 0.5)
	aggregator.AddProvider("low_confidence", provider, 1.0)

	// Test getting price with low confidence
	ctx := context.Background()
	_, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err == nil {
		t.Error("expected error for low confidence")
	}
}

// TestOracleAggregatorOutlierRemoval tests outlier price removal
func TestOracleAggregatorOutlierRemoval(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 4
	aggregator := NewOracleAggregator(config)

	// Add providers with different reliability
	provider1 := NewMockOracleProvider("provider1", "Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Provider 2", "http://test2.com", 1.0)
	provider3 := NewMockOracleProvider("provider3", "Provider 3", "http://test3.com", 1.0)
	provider4 := NewMockOracleProvider("provider4", "Provider 4", "http://test4.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 1.0)
	aggregator.AddProvider("provider2", provider2, 1.0)
	aggregator.AddProvider("provider3", provider3, 1.0)
	aggregator.AddProvider("provider4", provider4, 1.0)

	// Set specific prices to test outlier removal
	provider1.SetMockPrice("BTC", big.NewInt(100), 95)
	provider2.SetMockPrice("BTC", big.NewInt(105), 95)
	provider3.SetMockPrice("BTC", big.NewInt(110), 95)
	provider4.SetMockPrice("BTC", big.NewInt(1000), 95) // Outlier

	// Test getting aggregated price
	ctx := context.Background()
	price, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err != nil {
		t.Errorf("unexpected error getting price: %v", err)
	}

	if price.Outliers == 0 {
		t.Error("expected outliers to be removed")
	}

	// Check that the outlier was removed
	if price.Price.Cmp(big.NewInt(500)) > 0 {
		t.Errorf("expected price to be reasonable after outlier removal, got %s", price.Price.String())
	}
}

// TestOracleAggregatorWeightedAverage tests weighted average calculation
func TestOracleAggregatorWeightedAverage(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 2
	aggregator := NewOracleAggregator(config)

	// Add providers with different weights
	provider1 := NewMockOracleProvider("provider1", "Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Provider 2", "http://test2.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 2.0) // Higher weight
	aggregator.AddProvider("provider2", provider2, 1.0) // Lower weight

	// Set specific prices
	provider1.SetMockPrice("BTC", big.NewInt(100), 95)
	provider2.SetMockPrice("BTC", big.NewInt(200), 95)

	// Test getting aggregated price
	ctx := context.Background()
	price, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err != nil {
		t.Errorf("unexpected error getting price: %v", err)
	}

	// Weighted average: (2*100 + 1*200) / (2+1) = 400/3 ≈ 133
	expectedMin := big.NewInt(130)
	expectedMax := big.NewInt(140)

	if price.Price.Cmp(expectedMin) < 0 || price.Price.Cmp(expectedMax) > 0 {
		t.Errorf("expected price around 133, got %s", price.Price.String())
	}
}

// TestOracleAggregatorStatistics tests statistics collection
func TestOracleAggregatorStatistics(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 2
	aggregator := NewOracleAggregator(config)

	// Add providers
	provider1 := NewMockOracleProvider("provider1", "Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Provider 2", "http://test2.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 1.0)
	aggregator.AddProvider("provider2", provider2, 1.0)

	// Get prices multiple times
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := aggregator.GetAggregatedPrice(ctx, "BTC")
		if err != nil {
			t.Errorf("unexpected error getting price: %v", err)
		}
	}

	// Check statistics
	stats := aggregator.GetStats()
	// First call: 2 provider requests, subsequent calls: cached responses (count as 1 request each)
	// Total: 2 + 1 + 1 = 4 requests
	if stats.TotalRequests < 4 {
		t.Errorf("expected at least 4 total requests, got %d", stats.TotalRequests)
	}

	providerStats := stats.ProviderStats
	if len(providerStats) != 2 {
		t.Errorf("expected 2 provider stats, got %d", len(providerStats))
	}

	// Check provider1 stats
	if stats1, exists := providerStats["provider1"]; exists {
		// Provider1 gets 1 request from first call + 2 requests from cached responses
		if stats1.Requests < 3 {
			t.Errorf("expected provider1 to have at least 3 requests, got %d", stats1.Requests)
		}
		if stats1.Reliability < 0.8 {
			t.Errorf("expected provider1 reliability >= 0.8, got %f", stats1.Reliability)
		}
	}
}

// TestOracleAggregatorEvents tests event recording
func TestOracleAggregatorEvents(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 2
	aggregator := NewOracleAggregator(config)

	// Add providers
	provider1 := NewMockOracleProvider("provider1", "Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Provider 2", "http://test2.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 1.0)
	aggregator.AddProvider("provider2", provider2, 1.0)

	// Get price to trigger events
	ctx := context.Background()
	_, err := aggregator.GetAggregatedPrice(ctx, "BTC")
	if err != nil {
		t.Errorf("unexpected error getting price: %v", err)
	}

	// Check events
	events := aggregator.GetEvents()
	if len(events) < 3 { // provider_added (2) + price_updated (1)
		t.Errorf("expected at least 3 events, got %d", len(events))
	}

	// Check for specific event types
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	expectedTypes := []string{"provider_added", "price_updated"}
	for _, expectedType := range expectedTypes {
		if !eventTypes[expectedType] {
			t.Errorf("expected event type %s", expectedType)
		}
	}
}

// TestOracleAggregatorConcurrency tests concurrent access
func TestOracleAggregatorConcurrency(t *testing.T) {
	config := DefaultOracleConfig()
	config.MinProviders = 2
	aggregator := NewOracleAggregator(config)

	// Add providers
	provider1 := NewMockOracleProvider("provider1", "Provider 1", "http://test1.com", 1.0)
	provider2 := NewMockOracleProvider("provider2", "Provider 2", "http://test2.com", 1.0)

	aggregator.AddProvider("provider1", provider1, 1.0)
	aggregator.AddProvider("provider2", provider2, 1.0)

	// Test concurrent price requests
	ctx := context.Background()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := aggregator.GetAggregatedPrice(ctx, "BTC")
			if err != nil {
				t.Errorf("concurrent price request failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that all requests were processed
	stats := aggregator.GetStats()
	// First call: 2 provider requests, subsequent calls: cached responses (count as 1 request each)
	// Total: 2 + 9 = 11 requests
	if stats.TotalRequests < 11 {
		t.Errorf("expected at least 11 total requests, got %d", stats.TotalRequests)
	}
}

// TestMockOracleProvider tests mock provider functionality
func TestMockOracleProvider(t *testing.T) {
	provider := NewMockOracleProvider("test", "Test Provider", "http://test.com", 1.0) // 100% reliability for consistent testing

	// Test GetPrice
	ctx := context.Background()
	price, err := provider.GetPrice(ctx, "BTC")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if price == nil {
		t.Fatal("expected non-nil price")
	}

	if price.Asset != "BTC" {
		t.Errorf("expected asset BTC, got %s", price.Asset)
	}

	if price.Provider != "test" {
		t.Errorf("expected provider 'test', got %s", price.Provider)
	}

	// Test SetMockPrice
	expectedPrice := big.NewInt(50000)
	provider.SetMockPrice("ETH", expectedPrice, 95)

	mockPrice := provider.GetMockPrice("ETH")
	if mockPrice == nil {
		t.Fatal("expected non-nil mock price")
	}

	if mockPrice.Price.Cmp(expectedPrice) != 0 {
		t.Errorf("expected price %s, got %s", expectedPrice.String(), mockPrice.Price.String())
	}

	// Test statistics
	requests, successes, failures := provider.GetStats()
	if requests != 1 {
		t.Errorf("expected 1 request, got %d", requests)
	}
	if successes != 1 {
		t.Errorf("expected 1 success, got %d", successes)
	}
	if failures != 0 {
		t.Errorf("expected 0 failures, got %d", failures)
	}
}

// TestOracleProofValidation tests oracle proof validation
func TestOracleProofValidation(t *testing.T) {
	aggregator := NewOracleAggregator(DefaultOracleConfig())

	// Test nil proof
	err := aggregator.ValidateProof(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil proof")
	}

	// Test empty proof data
	proof := &OracleProof{
		Data:      []byte{},
		Signature: []byte("signature"),
		PublicKey: []byte("public_key"),
		Timestamp: time.Now(),
		Nonce:     1,
	}

	err = aggregator.ValidateProof(context.Background(), proof)
	if err == nil {
		t.Error("expected error for empty proof data")
	}

	// Test valid proof
	validProof := &OracleProof{
		Data:      []byte("valid_data"),
		Signature: []byte("signature"),
		PublicKey: []byte("public_key"),
		Timestamp: time.Now(),
		Nonce:     1,
	}

	err = aggregator.ValidateProof(context.Background(), validProof)
	if err != nil {
		t.Errorf("unexpected error for valid proof: %v", err)
	}
}
