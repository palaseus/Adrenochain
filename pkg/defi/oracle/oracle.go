package oracle

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"
)

// OracleProvider represents an oracle data provider
type OracleProvider interface {
	// GetPrice retrieves the current price for a given asset
	GetPrice(ctx context.Context, asset string) (*PriceData, error)

	// ValidateProof validates cryptographic proof for oracle data
	ValidateProof(ctx context.Context, proof *OracleProof) error

	// UpdatePrice updates the price for a given asset
	UpdatePrice(ctx context.Context, asset string, price *big.Int, proof *OracleProof) error

	// GetProviderInfo returns information about the oracle provider
	GetProviderInfo() *ProviderInfo
}

// PriceData represents oracle price data
type PriceData struct {
	Asset       string    `json:"asset"`
	Price       *big.Int  `json:"price"`
	Timestamp   time.Time `json:"timestamp"`
	BlockNumber uint64    `json:"block_number"`
	Provider    string    `json:"provider"`
	Confidence  uint8     `json:"confidence"` // 0-100 confidence level
	Source      string    `json:"source"`     // Data source identifier
}

// OracleProof represents cryptographic proof for oracle data
type OracleProof struct {
	Data      []byte    `json:"data"`
	Signature []byte    `json:"signature"`
	PublicKey []byte    `json:"public_key"`
	Timestamp time.Time `json:"timestamp"`
	Nonce     uint64    `json:"nonce"`
}

// ProviderInfo contains information about an oracle provider
type ProviderInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	PublicKey   []byte    `json:"public_key"`
	Active      bool      `json:"active"`
	LastUpdate  time.Time `json:"last_update"`
	Reliability float64   `json:"reliability"` // 0.0-1.0 reliability score
}

// OracleAggregator aggregates data from multiple oracle providers
type OracleAggregator struct {
	mu sync.RWMutex

	// Configuration
	config OracleConfig

	// Providers
	providers map[string]OracleProvider
	weights   map[string]float64

	// State
	prices     map[string]*AggregatedPrice
	lastUpdate map[string]time.Time
	events     []OracleEvent

	// Statistics
	stats *OracleStats
}

// OracleConfig holds configuration for the oracle aggregator
type OracleConfig struct {
	MinProviders        int           `json:"min_providers"`        // Minimum providers required
	MaxPriceAge         time.Duration `json:"max_price_age"`        // Maximum age of price data
	UpdateInterval      time.Duration `json:"update_interval"`      // How often to update prices
	ConfidenceThreshold uint8         `json:"confidence_threshold"` // Minimum confidence level
	FallbackEnabled     bool          `json:"fallback_enabled"`     // Enable fallback mechanisms
}

// DefaultOracleConfig returns default oracle configuration
func DefaultOracleConfig() OracleConfig {
	return OracleConfig{
		MinProviders:        3,
		MaxPriceAge:         5 * time.Minute,
		UpdateInterval:      1 * time.Minute,
		ConfidenceThreshold: 80,
		FallbackEnabled:     true,
	}
}

// AggregatedPrice represents aggregated price data from multiple providers
type AggregatedPrice struct {
	Asset       string    `json:"asset"`
	Price       *big.Int  `json:"price"`
	Timestamp   time.Time `json:"timestamp"`
	BlockNumber uint64    `json:"block_number"`
	Confidence  uint8     `json:"confidence"`
	Providers   int       `json:"providers"`
	Variance    float64   `json:"variance"` // Price variance across providers
	Outliers    int       `json:"outliers"` // Number of outlier prices removed
}

// OracleEvent represents an oracle-related event
type OracleEvent struct {
	Type      string                 `json:"type"`
	Asset     string                 `json:"asset"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Block     uint64                 `json:"block"`
}

// OracleStats contains oracle performance statistics
type OracleStats struct {
	TotalRequests      uint64                    `json:"total_requests"`
	SuccessfulRequests uint64                    `json:"successful_requests"`
	FailedRequests     uint64                    `json:"failed_requests"`
	AverageLatency     time.Duration             `json:"average_latency"`
	ProviderStats      map[string]*ProviderStats `json:"provider_stats"`
	LastUpdate         time.Time                 `json:"last_update"`
}

// ProviderStats contains statistics for a specific provider
type ProviderStats struct {
	Requests       uint64        `json:"requests"`
	Successes      uint64        `json:"successes"`
	Failures       uint64        `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	LastSuccess    time.Time     `json:"last_success"`
	LastFailure    time.Time     `json:"last_failure"`
	Reliability    float64       `json:"reliability"`
}

// NewOracleAggregator creates a new oracle aggregator
func NewOracleAggregator(config OracleConfig) *OracleAggregator {
	return &OracleAggregator{
		config:     config,
		providers:  make(map[string]OracleProvider),
		weights:    make(map[string]float64),
		prices:     make(map[string]*AggregatedPrice),
		lastUpdate: make(map[string]time.Time),
		events:     make([]OracleEvent, 0),
		stats: &OracleStats{
			ProviderStats: make(map[string]*ProviderStats),
		},
	}
}

// AddProvider adds an oracle provider to the aggregator
func (oa *OracleAggregator) AddProvider(name string, provider OracleProvider, weight float64) error {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	if weight <= 0 {
		return fmt.Errorf("weight must be positive")
	}

	oa.providers[name] = provider
	oa.weights[name] = weight

	// Initialize provider stats
	oa.stats.ProviderStats[name] = &ProviderStats{
		Reliability: 1.0,
	}

	// Record event
	oa.recordEvent("provider_added", "", map[string]interface{}{
		"provider": name,
		"weight":   weight,
	})

	return nil
}

// RemoveProvider removes an oracle provider from the aggregator
func (oa *OracleAggregator) RemoveProvider(name string) error {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	if _, exists := oa.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	delete(oa.providers, name)
	delete(oa.weights, name)
	delete(oa.stats.ProviderStats, name)

	// Record event
	oa.recordEvent("provider_removed", "", map[string]interface{}{
		"provider": name,
	})

	return nil
}

// GetAggregatedPrice retrieves aggregated price for an asset
func (oa *OracleAggregator) GetAggregatedPrice(ctx context.Context, asset string) (*AggregatedPrice, error) {
	oa.mu.RLock()

	// Check if we have a recent price
	if price, exists := oa.prices[asset]; exists {
		if time.Since(price.Timestamp) < oa.config.MaxPriceAge {
			oa.mu.RUnlock()
			// Still update statistics even for cached prices
			oa.updateRequestStats()
			return price, nil
		}
	}

	oa.mu.RUnlock()

	// Update price if needed
	return oa.UpdateAggregatedPrice(ctx, asset)
}

// UpdateAggregatedPrice updates the aggregated price for an asset
func (oa *OracleAggregator) UpdateAggregatedPrice(ctx context.Context, asset string) (*AggregatedPrice, error) {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	if len(oa.providers) < oa.config.MinProviders {
		return nil, fmt.Errorf("insufficient providers: need %d, have %d", oa.config.MinProviders, len(oa.providers))
	}

	// Collect prices from all providers
	prices := make([]*PriceData, 0)
	startTime := time.Now()

	for name, provider := range oa.providers {
		price, err := oa.getProviderPrice(ctx, name, provider, asset)
		if err != nil {
			oa.recordProviderFailure(name, err)
			continue
		}

		// Validate price data
		if err := oa.validatePriceData(price); err != nil {
			oa.recordProviderFailure(name, err)
			continue
		}

		prices = append(prices, price)
		oa.recordProviderSuccess(name, time.Since(startTime))
	}

	if len(prices) < oa.config.MinProviders {
		return nil, fmt.Errorf("insufficient valid prices: need %d, have %d", oa.config.MinProviders, len(prices))
	}

	// Aggregate prices
	aggregatedPrice := oa.aggregatePrices(asset, prices)
	oa.prices[asset] = aggregatedPrice
	oa.lastUpdate[asset] = time.Now()

	// Record event
	oa.recordEvent("price_updated", asset, map[string]interface{}{
		"price":      aggregatedPrice.Price.String(),
		"providers":  aggregatedPrice.Providers,
		"confidence": aggregatedPrice.Confidence,
		"variance":   aggregatedPrice.Variance,
	})

	return aggregatedPrice, nil
}

// getProviderPrice retrieves price from a specific provider
func (oa *OracleAggregator) getProviderPrice(ctx context.Context, name string, provider OracleProvider, asset string) (*PriceData, error) {
	startTime := time.Now()

	price, err := provider.GetPrice(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("provider %s error: %w", name, err)
	}

	// Update provider stats
	if stats, exists := oa.stats.ProviderStats[name]; exists {
		stats.AverageLatency = oa.updateAverageLatency(stats.AverageLatency, time.Since(startTime))
	}

	return price, nil
}

// validatePriceData validates price data from a provider
func (oa *OracleAggregator) validatePriceData(price *PriceData) error {
	if price == nil {
		return fmt.Errorf("price data is nil")
	}

	if price.Price == nil || price.Price.Sign() <= 0 {
		return fmt.Errorf("invalid price value")
	}

	if price.Confidence < oa.config.ConfidenceThreshold {
		return fmt.Errorf("insufficient confidence: %d < %d", price.Confidence, oa.config.ConfidenceThreshold)
	}

	if time.Since(price.Timestamp) > oa.config.MaxPriceAge {
		return fmt.Errorf("price data too old: %v", time.Since(price.Timestamp))
	}

	return nil
}

// aggregatePrices aggregates prices from multiple providers
func (oa *OracleAggregator) aggregatePrices(asset string, prices []*PriceData) *AggregatedPrice {
	if len(prices) == 0 {
		return nil
	}

	// Remove outliers using IQR method
	filteredPrices := oa.removeOutliers(prices)
	outliersRemoved := len(prices) - len(filteredPrices)

	// Calculate weighted average price
	totalWeight := big.NewFloat(0)
	weightedSum := big.NewFloat(0)

	for _, price := range filteredPrices {
		weight := oa.weights[price.Provider]
		priceFloat := new(big.Float).SetInt(price.Price)

		weightedPrice := new(big.Float).Mul(priceFloat, big.NewFloat(weight))
		weightedSum.Add(weightedSum, weightedPrice)
		totalWeight.Add(totalWeight, big.NewFloat(weight))
	}

	// Calculate final price
	var finalPrice *big.Int
	if totalWeight.Sign() > 0 {
		avgPrice := new(big.Float).Quo(weightedSum, totalWeight)
		finalPrice, _ = avgPrice.Int(nil)
	} else {
		// Fallback to simple average if no weights
		sum := big.NewInt(0)
		for _, price := range filteredPrices {
			sum.Add(sum, price.Price)
		}
		finalPrice = new(big.Int).Div(sum, big.NewInt(int64(len(filteredPrices))))
	}

	// Calculate confidence and variance
	confidence := oa.calculateConfidence(filteredPrices)
	variance := oa.calculateVariance(filteredPrices, finalPrice)

	// Find latest timestamp and block number
	var latestTimestamp time.Time
	var latestBlock uint64
	for _, price := range filteredPrices {
		if price.Timestamp.After(latestTimestamp) {
			latestTimestamp = price.Timestamp
		}
		if price.BlockNumber > latestBlock {
			latestBlock = price.BlockNumber
		}
	}

	return &AggregatedPrice{
		Asset:       asset,
		Price:       finalPrice,
		Timestamp:   latestTimestamp,
		BlockNumber: latestBlock,
		Confidence:  confidence,
		Providers:   len(filteredPrices),
		Variance:    variance,
		Outliers:    outliersRemoved,
	}
}

// removeOutliers removes outlier prices using IQR method
func (oa *OracleAggregator) removeOutliers(prices []*PriceData) []*PriceData {
	if len(prices) < 4 {
		return prices
	}

	// Convert prices to float64 for statistical analysis
	priceValues := make([]float64, len(prices))
	for i, price := range prices {
		priceValues[i] = float64(price.Price.Int64())
	}

	// For small datasets, use standard deviation method as it's more effective
	if len(prices) <= 5 {
		return oa.removeOutliersByStdDev(prices, priceValues)
	}

	// Calculate Q1, Q3, and IQR
	q1 := oa.percentile(priceValues, 25)
	q3 := oa.percentile(priceValues, 75)
	iqr := q3 - q1

	// Define outlier bounds
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	// Filter out outliers
	filtered := make([]*PriceData, 0)
	for _, price := range prices {
		priceValue := float64(price.Price.Int64())
		if priceValue >= lowerBound && priceValue <= upperBound {
			filtered = append(filtered, price)
		}
	}

	return filtered
}

// removeOutliersByStdDev removes outliers using standard deviation method for small datasets
func (oa *OracleAggregator) removeOutliersByStdDev(prices []*PriceData, priceValues []float64) []*PriceData {
	if len(priceValues) < 2 {
		return prices
	}

	// Calculate mean
	sum := 0.0
	for _, price := range priceValues {
		sum += price
	}
	mean := sum / float64(len(priceValues))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, price := range priceValues {
		diff := price - mean
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(priceValues)))

	// Define outlier bounds (2 standard deviations)
	lowerBound := mean - 2*stdDev
	upperBound := mean + 2*stdDev

	// For small datasets, use a simple approach: remove prices that are more than 3x the mean
	// This is more effective for small datasets with extreme outliers
	meanPrice := 0.0
	for _, price := range priceValues {
		meanPrice += price
	}
	meanPrice = meanPrice / float64(len(priceValues))

	simpleLowerBound := meanPrice / 3
	simpleUpperBound := meanPrice * 3

	// Filter out outliers
	filtered := make([]*PriceData, 0)
	for i, price := range prices {
		priceValue := priceValues[i]
		// Use the more restrictive of the two bounds
		actualLowerBound := math.Max(lowerBound, simpleLowerBound)
		actualUpperBound := math.Min(upperBound, simpleUpperBound)

		if priceValue >= actualLowerBound && priceValue <= actualUpperBound {
			filtered = append(filtered, price)
		}
	}

	return filtered
}

// percentile calculates the nth percentile of a slice of values
func (oa *OracleAggregator) percentile(values []float64, n int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple implementation - for production use a proper sorting library
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (n * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// calculateConfidence calculates overall confidence from provider data
func (oa *OracleAggregator) calculateConfidence(prices []*PriceData) uint8 {
	if len(prices) == 0 {
		return 0
	}

	totalConfidence := 0
	for _, price := range prices {
		totalConfidence += int(price.Confidence)
	}

	avgConfidence := totalConfidence / len(prices)

	// Boost confidence based on number of providers
	providerBonus := len(prices) * 2
	finalConfidence := avgConfidence + providerBonus

	if finalConfidence > 100 {
		finalConfidence = 100
	}

	return uint8(finalConfidence)
}

// calculateVariance calculates price variance across providers
func (oa *OracleAggregator) calculateVariance(prices []*PriceData, meanPrice *big.Int) float64 {
	if len(prices) < 2 {
		return 0
	}

	mean := float64(meanPrice.Int64())
	sumSquaredDiff := 0.0

	for _, price := range prices {
		diff := float64(price.Price.Int64()) - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(prices))
}

// recordProviderSuccess records a successful provider request
func (oa *OracleAggregator) recordProviderSuccess(providerName string, latency time.Duration) {
	oa.stats.TotalRequests++
	oa.stats.SuccessfulRequests++

	if stats, exists := oa.stats.ProviderStats[providerName]; exists {
		stats.Requests++
		stats.Successes++
		stats.LastSuccess = time.Now()
		stats.Reliability = float64(stats.Successes) / float64(stats.Requests)
	}
}

// recordProviderFailure records a failed provider request
func (oa *OracleAggregator) recordProviderFailure(providerName string, err error) {
	oa.stats.TotalRequests++
	oa.stats.FailedRequests++

	if stats, exists := oa.stats.ProviderStats[providerName]; exists {
		stats.Requests++
		stats.Failures++
		stats.LastFailure = time.Now()
		stats.Reliability = float64(stats.Successes) / float64(stats.Requests)
	}
}

// updateRequestStats updates request statistics for cached responses
func (oa *OracleAggregator) updateRequestStats() {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	oa.stats.TotalRequests++
	oa.stats.SuccessfulRequests++

	// Update provider statistics for cached responses
	// Since we don't know which specific provider provided the cached price,
	// we'll update all active providers' statistics
	for name := range oa.providers {
		if stats, exists := oa.stats.ProviderStats[name]; exists {
			stats.Requests++
			stats.Successes++
			stats.LastSuccess = time.Now()
			stats.Reliability = float64(stats.Successes) / float64(stats.Requests)
		}
	}
}

// recordEvent records an oracle event
func (oa *OracleAggregator) recordEvent(eventType, asset string, data map[string]interface{}) {
	event := OracleEvent{
		Type:      eventType,
		Asset:     asset,
		Data:      data,
		Timestamp: time.Now(),
		Block:     0, // Would be set from blockchain context
	}
	oa.events = append(oa.events, event)
}

// updateAverageLatency updates running average latency
func (oa *OracleAggregator) updateAverageLatency(current, new time.Duration) time.Duration {
	if current == 0 {
		return new
	}
	// Simple exponential moving average
	alpha := 0.1
	return time.Duration(float64(current)*(1-alpha) + float64(new)*alpha)
}

// GetStats returns oracle statistics
func (oa *OracleAggregator) GetStats() *OracleStats {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := &OracleStats{
		TotalRequests:      oa.stats.TotalRequests,
		SuccessfulRequests: oa.stats.SuccessfulRequests,
		FailedRequests:     oa.stats.FailedRequests,
		AverageLatency:     oa.stats.AverageLatency,
		LastUpdate:         oa.stats.LastUpdate,
		ProviderStats:      make(map[string]*ProviderStats),
	}

	for name, stats := range oa.stats.ProviderStats {
		statsCopy.ProviderStats[name] = &ProviderStats{
			Requests:       stats.Requests,
			Successes:      stats.Successes,
			Failures:       stats.Failures,
			AverageLatency: stats.AverageLatency,
			LastSuccess:    stats.LastSuccess,
			LastFailure:    stats.LastFailure,
			Reliability:    stats.Reliability,
		}
	}

	return statsCopy
}

// GetEvents returns oracle events
func (oa *OracleAggregator) GetEvents() []OracleEvent {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	events := make([]OracleEvent, len(oa.events))
	copy(events, oa.events)
	return events
}

// GetProviders returns information about all providers
func (oa *OracleAggregator) GetProviders() map[string]*ProviderInfo {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	providers := make(map[string]*ProviderInfo)
	for name, provider := range oa.providers {
		providers[name] = provider.GetProviderInfo()
	}
	return providers
}

// ValidateProof validates an oracle proof
func (oa *OracleAggregator) ValidateProof(ctx context.Context, proof *OracleProof) error {
	if proof == nil {
		return fmt.Errorf("proof cannot be nil")
	}

	// Validate proof structure
	if len(proof.Data) == 0 {
		return fmt.Errorf("proof data cannot be empty")
	}

	if len(proof.Signature) == 0 {
		return fmt.Errorf("proof signature cannot be empty")
	}

	if len(proof.PublicKey) == 0 {
		return fmt.Errorf("proof public key cannot be empty")
	}

	// Validate timestamp
	if time.Since(proof.Timestamp) > oa.config.MaxPriceAge {
		return fmt.Errorf("proof timestamp too old: %v", time.Since(proof.Timestamp))
	}

	// Validate nonce
	if proof.Nonce == 0 {
		return fmt.Errorf("proof nonce cannot be zero")
	}

	// Calculate data hash for future signature verification
	_ = sha256.Sum256(proof.Data)

	// TODO: Implement signature verification
	// For now, just validate the structure

	return nil
}
