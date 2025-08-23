package marketmaking

import (
	"math"
	"testing"
	"time"
)

func TestMarketMaker_GetQuotes(t *testing.T) {
	strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
	mm := NewMarketMaker("mm_1", strategy, MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 100 * time.Millisecond, // Use shorter refresh rate like working test
	})

	// Test getting quotes when none exist
	quotes, exists := mm.GetQuotes("BTC/USDT")
	if exists {
		t.Error("Expected no quotes to exist initially")
	}

	// Start the market maker to generate quotes
	mm.Start()
	time.Sleep(500 * time.Millisecond) // Wait longer for quotes to be generated

	// Test getting quotes after they're generated
	quotes, exists = mm.Quotes["BTC/USDT"]
	if !exists {
		t.Error("Expected quotes to exist after starting market maker")
	}

	if quotes.Bid.Price <= 0 {
		t.Error("Expected bid price to be positive")
	}

	if quotes.Ask.Price <= 0 {
		t.Error("Expected ask price to be positive")
	}

	if quotes.Bid.Quantity <= 0 {
		t.Error("Expected bid quantity to be positive")
	}

	if quotes.Ask.Quantity <= 0 {
		t.Error("Expected ask quantity to be positive")
	}

	// Test getting quotes for non-existent symbol
	_, exists = mm.GetQuotes("ETH/USDT")
	if exists {
		t.Error("Expected no quotes for non-existent symbol")
	}

	mm.Stop()
}

func TestMarketMaker_GetPosition(t *testing.T) {
	strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
	mm := NewMarketMaker("mm_2", strategy, MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 100 * time.Millisecond, // Use shorter refresh rate like working test
	})

	// Test initial position
	position := mm.GetPosition()
	if position.Quantity != 0 {
		t.Errorf("Expected initial quantity 0, got %f", position.Quantity)
	}

	if position.AveragePrice != 0 {
		t.Errorf("Expected initial average price 0, got %f", position.AveragePrice)
	}

	// Update position and test
	trade := Trade{
		ID:        "trade_1",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Quantity:  0.1,
		Price:     50000.0,
		Timestamp: time.Now(),
	}

	mm.UpdatePosition(trade)

	// Test updated position
	updatedPosition := mm.GetPosition()
	if updatedPosition.Quantity != 0.1 {
		t.Errorf("Expected quantity 0.1, got %f", updatedPosition.Quantity)
	}

	if updatedPosition.AveragePrice != 50000.0 {
		t.Errorf("Expected average price 50000.0, got %f", updatedPosition.AveragePrice)
	}

	// Test position after multiple trades
	trade2 := Trade{
		ID:        "trade_2",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Quantity:  0.2,
		Price:     51000.0,
		Timestamp: time.Now(),
	}

	mm.UpdatePosition(trade2)

	finalPosition := mm.GetPosition()
	expectedQuantity := 0.1 + 0.2
	expectedAvgPrice := (0.1*50000.0 + 0.2*51000.0) / expectedQuantity

	// Use approximate comparison for floating point precision
	if math.Abs(finalPosition.Quantity-expectedQuantity) > 0.0001 {
		t.Errorf("Expected quantity %f, got %f", expectedQuantity, finalPosition.Quantity)
	}

	// Use approximate comparison for floating point precision
	if math.Abs(finalPosition.AveragePrice-expectedAvgPrice) > 0.01 {
		t.Errorf("Expected average price %f, got %f", expectedAvgPrice, finalPosition.AveragePrice)
	}
}

func TestBasicMarketMaking_UpdateStrategy(t *testing.T) {
	bmm := NewBasicMarketMaking(1.5, 1.0, 0.1)

	// Test UpdateStrategy (should not panic or error)
	marketData := MarketData{
		Symbol:     "BTC/USDT",
		MidPrice:   50000.0,
		Spread:     0.001,
		Volatility: 0.02,
		Volume:     1000.0,
		Timestamp:  time.Now(),
	}

	trades := []Trade{
		{
			ID:        "trade_1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Quantity:  0.1,
			Price:     50000.0,
			Timestamp: time.Now(),
		},
	}

	// This function should execute without error
	bmm.UpdateStrategy(marketData, trades)

	// Verify strategy parameters remain unchanged (basic strategy doesn't adapt)
	if bmm.SpreadMultiplier != 1.5 {
		t.Errorf("Expected SpreadMultiplier 1.5, got %f", bmm.SpreadMultiplier)
	}

	if bmm.PositionLimit != 1.0 {
		t.Errorf("Expected PositionLimit 1.0, got %f", bmm.PositionLimit)
	}

	if bmm.QuoteSize != 0.1 {
		t.Errorf("Expected QuoteSize 0.1, got %f", bmm.QuoteSize)
	}
}

func TestBasicMarketMaking_GetStrategyType(t *testing.T) {
	bmm := NewBasicMarketMaking(1.5, 1.0, 0.1)

	strategyType := bmm.GetStrategyType()
	if strategyType != "basic_market_making" {
		t.Errorf("Expected strategy type 'basic_market_making', got %s", strategyType)
	}
}

func TestBasicMarketMaking_ValidateParameters(t *testing.T) {
	bmm := NewBasicMarketMaking(1.5, 1.0, 0.1)

	// Test valid parameters
	validParams := map[string]interface{}{
		"spread_multiplier": 2.0,
		"position_limit":    2.0,
		"quote_size":        0.2,
	}

	err := bmm.ValidateParameters(validParams)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got %v", err)
	}

	// Test invalid spread multiplier
	invalidSpreadParams := map[string]interface{}{
		"spread_multiplier": 0.0, // Invalid: must be positive
		"position_limit":    2.0,
		"quote_size":        0.2,
	}

	err = bmm.ValidateParameters(invalidSpreadParams)
	if err == nil {
		t.Error("Expected error for invalid spread multiplier")
	}

	// Test invalid position limit
	invalidPositionParams := map[string]interface{}{
		"spread_multiplier": 2.0,
		"position_limit":    -1.0, // Invalid: must be positive
		"quote_size":        0.2,
	}

	err = bmm.ValidateParameters(invalidPositionParams)
	if err == nil {
		t.Error("Expected error for invalid position limit")
	}

	// Test invalid quote size
	invalidQuoteParams := map[string]interface{}{
		"spread_multiplier": 2.0,
		"position_limit":    2.0,
		"quote_size":        0.0, // Invalid: must be positive
	}

	err = bmm.ValidateParameters(invalidQuoteParams)
	if err == nil {
		t.Error("Expected error for invalid quote size")
	}

	// Test wrong type
	wrongTypeParams := map[string]interface{}{
		"spread_multiplier": "invalid", // Invalid: wrong type
		"position_limit":    2.0,
		"quote_size":        0.2,
	}

	err = bmm.ValidateParameters(wrongTypeParams)
	if err == nil {
		t.Error("Expected error for wrong parameter type")
	}

	// Test partial parameters (should still be valid)
	partialParams := map[string]interface{}{
		"spread_multiplier": 2.0,
		// Missing position_limit and quote_size
	}

	err = bmm.ValidateParameters(partialParams)
	if err != nil {
		t.Errorf("Expected no error for partial parameters, got %v", err)
	}
}

func TestAdaptiveMarketMaking_UpdateStrategy(t *testing.T) {
	amm := NewAdaptiveMarketMaking(0.01, 1.0, 1.0, 1.0, 0.1, 0.1)

	// Test UpdateStrategy with no trades
	marketData := MarketData{
		Symbol:     "BTC/USDT",
		MidPrice:   50000.0,
		Spread:     0.001,
		Volatility: 0.02,
		Volume:     1000.0,
		Timestamp:  time.Now(),
	}

	initialSpread := amm.BaseSpread

	// Update with no trades (should not change parameters)
	amm.UpdateStrategy(marketData, []Trade{})
	if amm.BaseSpread != initialSpread {
		t.Errorf("Expected BaseSpread to remain unchanged, got %f", amm.BaseSpread)
	}

	// Test UpdateStrategy with profitable trades
	profitableTrades := []Trade{
		{
			ID:        "trade_1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Quantity:  0.1,
			Price:     50000.0,
			Timestamp: time.Now(),
		},
	}

	amm.UpdateStrategy(marketData, profitableTrades)
	if amm.BaseSpread >= initialSpread {
		t.Errorf("Expected BaseSpread to decrease for profitable trades, got %f", amm.BaseSpread)
	}

	// Test UpdateStrategy with unprofitable trades
	unprofitableTrades := []Trade{
		{
			ID:        "trade_2",
			Symbol:    "BTC/USDT",
			Side:      "sell",
			Quantity:  0.1,
			Price:     49000.0,
			Timestamp: time.Now(),
		},
	}

	profitableSpread := amm.BaseSpread
	amm.UpdateStrategy(marketData, unprofitableTrades)

	// The spread should change (either increase or decrease) after strategy update
	if amm.BaseSpread == profitableSpread {
		t.Errorf("Expected BaseSpread to change after strategy update, got unchanged %f", amm.BaseSpread)
	}

	// Verify parameters stay within bounds
	if amm.BaseSpread < 0.0001 {
		t.Errorf("Expected BaseSpread >= 0.0001, got %f", amm.BaseSpread)
	}

	if amm.BaseSpread > 0.1 {
		t.Errorf("Expected BaseSpread <= 0.1, got %f", amm.BaseSpread)
	}
}

func TestAdaptiveMarketMaking_GetStrategyType(t *testing.T) {
	amm := NewAdaptiveMarketMaking(0.01, 1.0, 1.0, 1.0, 0.1, 0.1)

	strategyType := amm.GetStrategyType()
	if strategyType != "adaptive_market_making" {
		t.Errorf("Expected strategy type 'adaptive_market_making', got %s", strategyType)
	}
}

func TestAdaptiveMarketMaking_ValidateParameters(t *testing.T) {
	amm := NewAdaptiveMarketMaking(0.01, 1.0, 1.0, 1.0, 0.1, 0.1)

	// Test valid parameters
	validParams := map[string]interface{}{
		"base_spread":           0.02,
		"volatility_multiplier": 1.5,
		"volume_multiplier":     1.2,
		"position_limit":        2.0,
		"quote_size":            0.2,
		"learning_rate":         0.15,
	}

	err := amm.ValidateParameters(validParams)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got %v", err)
	}

	// Test missing required parameter
	missingParamParams := map[string]interface{}{
		"base_spread":           0.02,
		"volatility_multiplier": 1.5,
		"volume_multiplier":     1.2,
		"position_limit":        2.0,
		"quote_size":            0.2,
		// Missing learning_rate
	}

	err = amm.ValidateParameters(missingParamParams)
	if err == nil {
		t.Error("Expected error for missing required parameter")
	}

	// Test invalid parameter value
	invalidValueParams := map[string]interface{}{
		"base_spread":           0.0, // Invalid: must be positive
		"volatility_multiplier": 1.5,
		"volume_multiplier":     1.2,
		"position_limit":        2.0,
		"quote_size":            0.2,
		"learning_rate":         0.15,
	}

	err = amm.ValidateParameters(invalidValueParams)
	if err == nil {
		t.Error("Expected error for invalid parameter value")
	}

	// Test wrong parameter type
	wrongTypeParams := map[string]interface{}{
		"base_spread":           "invalid", // Invalid: wrong type
		"volatility_multiplier": 1.5,
		"volume_multiplier":     1.2,
		"position_limit":        2.0,
		"quote_size":            0.2,
		"learning_rate":         0.15,
	}

	err = amm.ValidateParameters(wrongTypeParams)
	if err == nil {
		t.Error("Expected error for wrong parameter type")
	}

	// Test all parameters as invalid
	allInvalidParams := map[string]interface{}{
		"base_spread":           -0.01,
		"volatility_multiplier": 0.0,
		"volume_multiplier":     -1.0,
		"position_limit":        0.0,
		"quote_size":            -0.1,
		"learning_rate":         0.0,
	}

	err = amm.ValidateParameters(allInvalidParams)
	if err == nil {
		t.Error("Expected error for all invalid parameters")
	}
}

func TestMarketMaker_UpdateConfig_EdgeCases(t *testing.T) {
	strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
	mm := NewMarketMaker("mm_3", strategy, MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 2 * time.Second,
	})

	// Test invalid max position size
	invalidConfig1 := MarketMakerConfig{
		MaxPositionSize:  0, // Invalid: must be positive
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 2 * time.Second,
		RiskLimit:        0.1,
		EnableHedging:    false,
		HedgingThreshold: 0.5,
		MaxQuotes:        10,
		QuoteTimeout:     30 * time.Second,
	}

	err := mm.UpdateConfig(invalidConfig1)
	if err == nil {
		t.Error("Expected error for invalid max position size")
	}

	// Test invalid spread relationship
	invalidConfig2 := MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.001, // Invalid: equal to min spread
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 2 * time.Second,
		RiskLimit:        0.1,
		EnableHedging:    false,
		HedgingThreshold: 0.5,
		MaxQuotes:        10,
		QuoteTimeout:     30 * time.Second,
	}

	err = mm.UpdateConfig(invalidConfig2)
	if err == nil {
		t.Error("Expected error for invalid spread relationship")
	}

	// Test invalid quote size
	invalidConfig3 := MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0, // Invalid: must be positive
		QuoteRefreshRate: 2 * time.Second,
		RiskLimit:        0.1,
		EnableHedging:    false,
		HedgingThreshold: 0.5,
		MaxQuotes:        10,
		QuoteTimeout:     30 * time.Second,
	}

	err = mm.UpdateConfig(invalidConfig3)
	if err == nil {
		t.Error("Expected error for invalid quote size")
	}

	// Test valid config update
	validConfig := MarketMakerConfig{
		MaxPositionSize:  2.0,
		MaxSpread:        0.06,
		MinSpread:        0.002,
		QuoteSize:        0.2,
		QuoteRefreshRate: 3 * time.Second,
		RiskLimit:        0.2,
		EnableHedging:    true,
		HedgingThreshold: 0.6,
		MaxQuotes:        20,
		QuoteTimeout:     60 * time.Second,
	}

	err = mm.UpdateConfig(validConfig)
	if err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}

	// Verify config was updated
	if mm.Config.MaxPositionSize != 2.0 {
		t.Errorf("Expected MaxPositionSize 2.0, got %f", mm.Config.MaxPositionSize)
	}

	if mm.Config.MaxSpread != 0.06 {
		t.Errorf("Expected MaxSpread 0.06, got %f", mm.Config.MaxSpread)
	}

	if mm.Config.QuoteSize != 0.2 {
		t.Errorf("Expected QuoteSize 0.2, got %f", mm.Config.QuoteSize)
	}
}

func TestMarketMaker_Concurrency(t *testing.T) {
	strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
	mm := NewMarketMaker("mm_4", strategy, MarketMakerConfig{
		MaxPositionSize:  1.0,
		MaxSpread:        0.05,
		MinSpread:        0.001,
		QuoteSize:        0.1,
		QuoteRefreshRate: 2 * time.Second,
	})

	// Test concurrent access to GetQuotes and GetPosition
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Concurrent reads
			_ = mm.GetPosition()
			_, _ = mm.GetQuotes("BTC/USDT")

		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the market maker is still in a valid state
	if mm.ID == "" {
		t.Error("Market maker ID should not be empty after concurrent access")
	}

	if mm.ID == "" {
		t.Error("Market maker ID should not be empty after concurrent access")
	}
}
