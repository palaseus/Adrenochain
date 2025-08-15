package oracle

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

// MockOracleProvider is a mock implementation of OracleProvider for testing
type MockOracleProvider struct {
	name        string
	description string
	url         string
	active      bool
	reliability float64

	// Mock data
	prices map[string]*PriceData
	delay  time.Duration // Simulate network delay

	// Statistics
	requests  uint64
	successes uint64
	failures  uint64
}

// NewMockOracleProvider creates a new mock oracle provider
func NewMockOracleProvider(name, description, url string, reliability float64) *MockOracleProvider {
	return &MockOracleProvider{
		name:        name,
		description: description,
		url:         url,
		active:      true,
		reliability: reliability,
		prices:      make(map[string]*PriceData),
		delay:       100 * time.Millisecond,
	}
}

// GetPrice retrieves the current price for a given asset
func (m *MockOracleProvider) GetPrice(ctx context.Context, asset string) (*PriceData, error) {
	m.requests++

	// Simulate network delay
	time.Sleep(m.delay)

	// Check if we should fail based on reliability
	if m.shouldFail() {
		m.failures++
		return nil, ErrProviderFailure
	}

	// Check if a specific price was set for this asset
	if price, exists := m.prices[asset]; exists {
		m.successes++
		return price, nil
	}

	// Generate mock price data if no specific price was set
	price := m.generateMockPrice(asset)
	m.prices[asset] = price
	m.successes++

	return price, nil
}

// ValidateProof validates cryptographic proof for oracle data
func (m *MockOracleProvider) ValidateProof(ctx context.Context, proof *OracleProof) error {
	if proof == nil {
		return ErrInvalidProof
	}

	// Mock validation - always succeeds
	return nil
}

// UpdatePrice updates the price for a given asset
func (m *MockOracleProvider) UpdatePrice(ctx context.Context, asset string, price *big.Int, proof *OracleProof) error {
	if price == nil || price.Sign() <= 0 {
		return ErrInvalidPrice
	}

	// Validate proof first
	if err := m.ValidateProof(ctx, proof); err != nil {
		return err
	}

	// Update the price
	m.prices[asset] = &PriceData{
		Asset:       asset,
		Price:       new(big.Int).Set(price),
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
		Provider:    m.name,
		Confidence:  95, // High confidence for direct updates
		Source:      "mock_update",
	}

	return nil
}

// GetProviderInfo returns information about the oracle provider
func (m *MockOracleProvider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:        m.name,
		Description: m.description,
		URL:         m.url,
		PublicKey:   []byte("mock_public_key"),
		Active:      m.active,
		LastUpdate:  time.Now(),
		Reliability: m.reliability,
	}
}

// SetDelay sets the simulated network delay
func (m *MockOracleProvider) SetDelay(delay time.Duration) {
	m.delay = delay
}

// SetActive sets whether the provider is active
func (m *MockOracleProvider) SetActive(active bool) {
	m.active = active
}

// GetStats returns provider statistics
func (m *MockOracleProvider) GetStats() (uint64, uint64, uint64) {
	return m.requests, m.successes, m.failures
}

// shouldFail determines if the provider should fail based on reliability
func (m *MockOracleProvider) shouldFail() bool {
	// Generate random number between 0 and 1
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomValue := float64(randomBytes[0]) / 255.0

	return randomValue > m.reliability
}

// generateMockPrice generates mock price data for an asset
func (m *MockOracleProvider) generateMockPrice(asset string) *PriceData {
	// Generate random price between 1 and 100000
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	priceValue := new(big.Int).SetBytes(randomBytes[:4])
	priceValue.Mod(priceValue, big.NewInt(100000))
	priceValue.Add(priceValue, big.NewInt(1))

	// Generate confidence between 80-100
	confidence := uint8(80 + (randomBytes[4] % 21))

	return &PriceData{
		Asset:       asset,
		Price:       priceValue,
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
		Provider:    m.name,
		Confidence:  confidence,
		Source:      "mock_generated",
	}
}

// SetMockPrice sets a specific mock price for testing
func (m *MockOracleProvider) SetMockPrice(asset string, price *big.Int, confidence uint8) {
	m.prices[asset] = &PriceData{
		Asset:       asset,
		Price:       new(big.Int).Set(price),
		Timestamp:   time.Now(),
		BlockNumber: 0,
		Provider:    m.name,
		Confidence:  confidence,
		Source:      "mock_set",
	}
}

// GetMockPrice retrieves a previously set mock price
func (m *MockOracleProvider) GetMockPrice(asset string) *PriceData {
	return m.prices[asset]
}
