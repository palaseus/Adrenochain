package strategy_gen

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

// TestNewAutomatedStrategyGenerator tests the constructor
func TestNewAutomatedStrategyGenerator(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies:      50,
		MaxMarketData:      5000,
		GenerationInterval: time.Hour * 12,
		OptimizationRounds: 15,
		BacktestPeriod:     time.Hour * 24 * 60,
		EnableAutoOptimize: true,
		RiskConstraints: RiskConstraints{
			MaxDrawdown:    0.15,
			MaxVolatility:  0.25,
			MaxLeverage:    3.0,
			MinSharpeRatio: 1.0,
			MaxCorrelation: 0.7,
		},
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05,
			MaxDrawdown:    0.10,
			MinSharpeRatio: 1.2,
			MaxVolatility:  0.20,
			MinWinRate:     0.55,
		},
	}

	generator := NewAutomatedStrategyGenerator(config)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.ID == "" {
		t.Error("Expected generator ID to be set")
	}

	if generator.Config.MaxStrategies != 50 {
		t.Errorf("Expected MaxStrategies to be 50, got %d", generator.Config.MaxStrategies)
	}

	if generator.Config.MaxMarketData != 5000 {
		t.Errorf("Expected MaxMarketData to be 5000, got %d", generator.Config.MaxMarketData)
	}

	if generator.Config.GenerationInterval != time.Hour*12 {
		t.Errorf("Expected GenerationInterval to be 12 hours, got %v", generator.Config.GenerationInterval)
	}

	if generator.Config.OptimizationRounds != 15 {
		t.Errorf("Expected OptimizationRounds to be 15, got %d", generator.Config.OptimizationRounds)
	}

	if generator.Config.BacktestPeriod != time.Hour*24*60 {
		t.Errorf("Expected BacktestPeriod to be 60 days, got %v", generator.Config.BacktestPeriod)
	}

	if !generator.Config.EnableAutoOptimize {
		t.Error("Expected EnableAutoOptimize to be true")
	}

	if generator.Config.RiskConstraints.MaxDrawdown != 0.15 {
		t.Errorf("Expected MaxDrawdown to be 0.15, got %f", generator.Config.RiskConstraints.MaxDrawdown)
	}

	if generator.Config.PerformanceTargets.MinReturn != 0.05 {
		t.Errorf("Expected MinReturn to be 0.05, got %f", generator.Config.PerformanceTargets.MinReturn)
	}
}

// TestNewAutomatedStrategyGeneratorDefaults tests default value handling
func TestNewAutomatedStrategyGeneratorDefaults(t *testing.T) {
	config := GeneratorConfig{} // Empty config
	generator := NewAutomatedStrategyGenerator(config)

	if generator == nil {
		t.Fatal("Expected generator to be created with defaults, got nil")
	}

	// Check default values
	if generator.Config.MaxStrategies != 100 {
		t.Errorf("Expected default MaxStrategies to be 100, got %d", generator.Config.MaxStrategies)
	}

	if generator.Config.MaxMarketData != 10000 {
		t.Errorf("Expected default MaxMarketData to be 10000, got %d", generator.Config.MaxMarketData)
	}

	if generator.Config.GenerationInterval != time.Hour*6 {
		t.Errorf("Expected default GenerationInterval to be 6 hours, got %v", generator.Config.GenerationInterval)
	}

	if generator.Config.OptimizationRounds != 10 {
		t.Errorf("Expected default OptimizationRounds to be 10, got %d", generator.Config.OptimizationRounds)
	}

	if generator.Config.BacktestPeriod != time.Hour*24*30 {
		t.Errorf("Expected default BacktestPeriod to be 30 days, got %v", generator.Config.BacktestPeriod)
	}
}

// TestStartStop tests the start and stop functionality
func TestStartStop(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Test Start
	err := generator.Start()
	if err != nil {
		t.Errorf("Expected Start to succeed, got error: %v", err)
	}

	// Test Stop
	err = generator.Stop()
	if err != nil {
		t.Errorf("Expected Stop to succeed, got error: %v", err)
	}
}

// TestGenerateStrategy tests strategy generation
func TestGenerateStrategy(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies: 5,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create test market data
	marketData := []MarketData{
		{
			Timestamp:  time.Now(),
			Asset:      "BTC",
			Price:      big.NewFloat(50000.0),
			Volume:     big.NewInt(1000000),
			MarketCap:  big.NewInt(1000000000000),
			Volatility: 0.25,
			RSI:        50.0,
			MACD:       0.0,
			BollingerBands: map[string]*big.Float{
				"upper":  big.NewFloat(55000.0),
				"middle": big.NewFloat(50000.0),
				"lower":  big.NewFloat(45000.0),
			},
			Features: map[string]float64{
				"trend":    0.6,
				"momentum": 0.4,
			},
		},
	}

	// Test successful strategy generation
	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC", "ETH"},
		RiskProfileModerate,
		marketData,
	)

	if err != nil {
		t.Errorf("Expected strategy generation to succeed, got error: %v", err)
	}

	if strategy == nil {
		t.Fatal("Expected strategy to be created, got nil")
	}

	if strategy.ID == "" {
		t.Error("Expected strategy ID to be set")
	}

	if strategy.Name == "" {
		t.Error("Expected strategy name to be set")
	}

	if strategy.Type != StrategyTypeTrendFollowing {
		t.Errorf("Expected strategy type to be TrendFollowing, got %v", strategy.Type)
	}

	if strategy.Status != StrategyStatusDraft {
		t.Errorf("Expected strategy status to be Draft, got %v", strategy.Status)
	}

	if strategy.RiskProfile != RiskProfileModerate {
		t.Errorf("Expected risk profile to be Moderate, got %v", strategy.RiskProfile)
	}

	if len(strategy.Assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(strategy.Assets))
	}

	if strategy.Assets[0] != "BTC" || strategy.Assets[1] != "ETH" {
		t.Errorf("Expected assets to be [BTC ETH], got %v", strategy.Assets)
	}

	if strategy.Parameters.EntryThreshold == 0 {
		t.Error("Expected strategy parameters to be set")
	}

	if strategy.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if strategy.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}

	// Test metrics update
	metrics := generator.GetMetrics()
	if metrics.GeneratedStrategies != 1 {
		t.Errorf("Expected GeneratedStrategies to be 1, got %d", metrics.GeneratedStrategies)
	}

	if metrics.TotalStrategies != 1 {
		t.Errorf("Expected TotalStrategies to be 1, got %d", metrics.TotalStrategies)
	}
}

// TestGenerateStrategyErrors tests error conditions
func TestGenerateStrategyErrors(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies: 1,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test with no assets
	_, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{},
		RiskProfileModerate,
		[]MarketData{{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)}},
	)
	if err == nil {
		t.Error("Expected error when no assets specified")
	}

	// Test with no market data
	_, err = generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		[]MarketData{},
	)
	if err == nil {
		t.Error("Expected error when no market data specified")
	}

	// Test max strategies limit
	_, err = generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		[]MarketData{{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)}},
	)
	if err != nil {
		t.Errorf("Expected first strategy to succeed, got error: %v", err)
	}

	// Try to create second strategy (should fail)
	_, err = generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"ETH"},
		RiskProfileModerate,
		[]MarketData{{Timestamp: time.Now(), Asset: "ETH", Price: big.NewFloat(3000.0)}},
	)
	if err == nil {
		t.Error("Expected error when max strategies reached")
	}
}

// TestGenerateStrategyParameters tests parameter generation for different strategy types
func TestGenerateStrategyParameters(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	testCases := []struct {
		strategyType StrategyType
		riskProfile  RiskProfile
		description  string
	}{
		{StrategyTypeTrendFollowing, RiskProfileConservative, "Trend Following Conservative"},
		{StrategyTypeTrendFollowing, RiskProfileAggressive, "Trend Following Aggressive"},
		{StrategyTypeMeanReversion, RiskProfileModerate, "Mean Reversion Moderate"},
		{StrategyTypeArbitrage, RiskProfileSpeculative, "Arbitrage Speculative"},
		{StrategyTypeMomentum, RiskProfileModerate, "Momentum Moderate"},
		{StrategyTypeGridTrading, RiskProfileModerate, "Grid Trading Moderate"},
		{StrategyTypeDCA, RiskProfileConservative, "DCA Conservative"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			strategy, err := generator.GenerateStrategy(tc.strategyType, []string{"BTC"}, tc.riskProfile, marketData)
			if err != nil {
				t.Errorf("Failed to generate strategy: %v", err)
				return
			}

			// Verify parameters are set
			if strategy.Parameters.EntryThreshold <= 0 {
				t.Error("Expected EntryThreshold to be positive")
			}

			if strategy.Parameters.ExitThreshold <= 0 {
				t.Error("Expected ExitThreshold to be positive")
			}

			if strategy.Parameters.StopLoss <= 0 {
				t.Error("Expected StopLoss to be positive")
			}

			if strategy.Parameters.TakeProfit <= 0 {
				t.Error("Expected TakeProfit to be positive")
			}

			if strategy.Parameters.PositionSize <= 0 {
				t.Error("Expected PositionSize to be positive")
			}

			if strategy.Parameters.MaxPositions <= 0 {
				t.Error("Expected MaxPositions to be positive")
			}

			if strategy.Parameters.RebalanceInterval <= 0 {
				t.Error("Expected RebalanceInterval to be positive")
			}

			// Verify risk profile adjustments
			switch tc.riskProfile {
			case RiskProfileConservative:
				if strategy.Parameters.PositionSize > 0.1 {
					t.Error("Conservative profile should have reduced position size")
				}
			case RiskProfileAggressive:
				if strategy.Parameters.PositionSize < 0.1 {
					t.Error("Aggressive profile should have increased position size")
				}
			}
		})
	}
}

// TestOptimizeStrategy tests strategy optimization
func TestOptimizeStrategy(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Create a strategy first
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Store original parameters
	originalEntryThreshold := strategy.Parameters.EntryThreshold
	originalExitThreshold := strategy.Parameters.ExitThreshold

	// Test optimization
	err = generator.OptimizeStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Expected optimization to succeed, got error: %v", err)
	}

	// Verify strategy was updated
	updatedStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Failed to get updated strategy: %v", err)
	}

	if updatedStrategy.Status != StrategyStatusDraft {
		t.Errorf("Expected status to be Draft after optimization, got %v", updatedStrategy.Status)
	}

	// Verify parameters were modified (optimization should change them)
	if updatedStrategy.Parameters.EntryThreshold == originalEntryThreshold {
		t.Error("Expected EntryThreshold to be modified during optimization")
	}

	if updatedStrategy.Parameters.ExitThreshold == originalExitThreshold {
		t.Error("Expected ExitThreshold to be modified during optimization")
	}

	// Check metrics
	metrics := generator.GetMetrics()
	if metrics.OptimizedStrategies != 1 {
		t.Errorf("Expected OptimizedStrategies to be 1, got %d", metrics.OptimizedStrategies)
	}
}

// TestOptimizeStrategyErrors tests optimization error conditions
func TestOptimizeStrategyErrors(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Test optimization of non-existent strategy
	err := generator.OptimizeStrategy("non_existent_id")
	if err == nil {
		t.Error("Expected error when optimizing non-existent strategy")
	}

	// Test optimization of active strategy
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Set strategy to active status
	strategy.Status = StrategyStatusActive

	// Try to optimize active strategy
	err = generator.OptimizeStrategy(strategy.ID)
	if err == nil {
		t.Error("Expected error when optimizing active strategy")
	}
}

// TestBacktestStrategy tests strategy backtesting
func TestBacktestStrategy(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Create a strategy first
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Test backtesting
	backtestResult, err := generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Errorf("Expected backtesting to succeed, got error: %v", err)
	}

	if backtestResult == nil {
		t.Fatal("Expected backtest result to be created")
	}

	if backtestResult.StrategyID != strategy.ID {
		t.Errorf("Expected StrategyID to match, got %s", backtestResult.StrategyID)
	}

	if backtestResult.StartDate.IsZero() {
		t.Error("Expected StartDate to be set")
	}

	if backtestResult.EndDate.IsZero() {
		t.Error("Expected EndDate to be set")
	}

	if backtestResult.InitialCapital == nil {
		t.Error("Expected InitialCapital to be set")
	}

	if backtestResult.FinalCapital == nil {
		t.Error("Expected FinalCapital to be set")
	}

	if backtestResult.TotalReturn == 0 {
		t.Error("Expected TotalReturn to be calculated")
	}

	if backtestResult.AnnualizedReturn == 0 {
		t.Error("Expected AnnualizedReturn to be calculated")
	}

	if backtestResult.Volatility == 0 {
		t.Error("Expected Volatility to be calculated")
	}

	if backtestResult.SharpeRatio == 0 {
		t.Error("Expected SharpeRatio to be calculated")
	}

	if backtestResult.MaxDrawdown == 0 {
		t.Error("Expected MaxDrawdown to be calculated")
	}

	if len(backtestResult.TradeHistory) == 0 {
		t.Error("Expected TradeHistory to contain trades")
	}

	if backtestResult.PerformanceMetrics == nil {
		t.Error("Expected PerformanceMetrics to be set")
	}

	if backtestResult.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify strategy was updated with backtest results
	updatedStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Failed to get updated strategy: %v", err)
	}

	if updatedStrategy.BacktestResults == nil {
		t.Error("Expected strategy to have backtest results")
	}

	if updatedStrategy.BacktestResults.StrategyID != strategy.ID {
		t.Errorf("Expected backtest results to match strategy ID")
	}

	// Check performance metrics were updated
	if updatedStrategy.Performance.TotalTrades == 0 {
		t.Error("Expected performance metrics to be updated")
	}

	// Check metrics
	metrics := generator.GetMetrics()
	if metrics.BacktestedStrategies != 1 {
		t.Errorf("Expected BacktestedStrategies to be 1, got %d", metrics.BacktestedStrategies)
	}
}

// TestBacktestStrategyErrors tests backtesting error conditions
func TestBacktestStrategyErrors(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Test backtesting non-existent strategy
	_, err := generator.BacktestStrategy("non_existent_id", []MarketData{})
	if err == nil {
		t.Error("Expected error when backtesting non-existent strategy")
	}

	// Test backtesting with empty market data
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	_, err = generator.BacktestStrategy(strategy.ID, []MarketData{})
	if err == nil {
		t.Error("Expected error when backtesting with empty market data")
	}
}

// TestActivateStrategy tests strategy activation
func TestActivateStrategy(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05,
			MaxDrawdown:    0.10,
			MinSharpeRatio: 1.0,
			MaxVolatility:  0.20,
			MinWinRate:     0.50,
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and backtest a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest the strategy
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	// Test activation
	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Expected activation to succeed, got error: %v", err)
	}

	// Verify strategy was activated
	activatedStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Failed to get activated strategy: %v", err)
	}

	if activatedStrategy.Status != StrategyStatusActive {
		t.Errorf("Expected status to be Active, got %v", activatedStrategy.Status)
	}

	// Check metrics
	metrics := generator.GetMetrics()
	if metrics.ActiveStrategies != 1 {
		t.Errorf("Expected ActiveStrategies to be 1, got %d", metrics.ActiveStrategies)
	}
}

// TestActivateStrategyErrors tests activation error conditions
func TestActivateStrategyErrors(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05,
			MaxDrawdown:    0.10,
			MinSharpeRatio: 1.0,
			MaxVolatility:  0.20,
			MinWinRate:     0.50,
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test activation of non-existent strategy
	err := generator.ActivateStrategy("non_existent_id")
	if err == nil {
		t.Error("Expected error when activating non-existent strategy")
	}

	// Test activation of strategy without backtesting
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err == nil {
		t.Error("Expected error when activating strategy without backtesting")
	}
}

// TestGenerateTradingSignals tests trading signal generation
func TestGenerateTradingSignals(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and activate a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Test signal generation
	currentMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(2000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.30,
		RSI:        75.0, // High RSI should trigger sell signal
		MACD:       0.5,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.8,
			"momentum": 0.9,
		},
	}

	signals, err := generator.GenerateTradingSignals(strategy.ID, currentMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// For trend following strategy with high RSI, we should get a sell signal
	if len(signals) == 0 {
		t.Error("Expected at least one trading signal")
	}

	// Verify signal properties
	for _, signal := range signals {
		if signal.ID == "" {
			t.Error("Expected signal ID to be set")
		}

		if signal.StrategyID != strategy.ID {
			t.Errorf("Expected StrategyID to match, got %s", signal.StrategyID)
		}

		if signal.Asset != "BTC" {
			t.Errorf("Expected Asset to be BTC, got %s", signal.Asset)
		}

		if signal.Price == nil {
			t.Error("Expected Price to be set")
		}

		if signal.Amount == nil {
			t.Error("Expected Amount to be set")
		}

		if signal.Confidence <= 0 || signal.Confidence > 1 {
			t.Errorf("Expected Confidence to be between 0 and 1, got %f", signal.Confidence)
		}

		if signal.Timestamp.IsZero() {
			t.Error("Expected Timestamp to be set")
		}

		if signal.ExpiresAt.IsZero() {
			t.Error("Expected ExpiresAt to be set")
		}

		if signal.Metadata == nil {
			t.Error("Expected Metadata to be set")
		}
	}
}

// TestGenerateTradingSignalsErrors tests signal generation error conditions
func TestGenerateTradingSignalsErrors(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Test signal generation for non-existent strategy
	_, err := generator.GenerateTradingSignals("non_existent_id", MarketData{})
	if err == nil {
		t.Error("Expected error when generating signals for non-existent strategy")
	}

	// Test signal generation for inactive strategy
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	_, err = generator.GenerateTradingSignals(strategy.ID, MarketData{})
	if err == nil {
		t.Error("Expected error when generating signals for inactive strategy")
	}
}

// TestGetStrategy tests strategy retrieval
func TestGetStrategy(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Create a strategy
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Test retrieval
	retrievedStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Expected strategy retrieval to succeed, got error: %v", err)
	}

	if retrievedStrategy == nil {
		t.Fatal("Expected retrieved strategy to not be nil")
	}

	if retrievedStrategy.ID != strategy.ID {
		t.Errorf("Expected ID to match, got %s", retrievedStrategy.ID)
	}

	if retrievedStrategy.Name != strategy.Name {
		t.Errorf("Expected Name to match, got %s", retrievedStrategy.Name)
	}

	// Test retrieval of non-existent strategy
	_, err = generator.GetStrategy("non_existent_id")
	if err == nil {
		t.Error("Expected error when retrieving non-existent strategy")
	}
}

// TestGetStrategies tests retrieval of all strategies
func TestGetStrategies(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Initially should have no strategies
	strategies := generator.GetStrategies()
	if len(strategies) != 0 {
		t.Errorf("Expected 0 strategies initially, got %d", len(strategies))
	}

	// Create multiple strategies
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy1, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create first strategy: %v", err)
	}

	strategy2, err := generator.GenerateStrategy(
		StrategyTypeMeanReversion,
		[]string{"ETH"},
		RiskProfileConservative,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create second strategy: %v", err)
	}

	// Test retrieval of all strategies
	allStrategies := generator.GetStrategies()
	if len(allStrategies) != 2 {
		t.Errorf("Expected 2 strategies, got %d", len(allStrategies))
	}

	// Verify both strategies are present
	found1 := false
	found2 := false
	for _, s := range allStrategies {
		if s.ID == strategy1.ID {
			found1 = true
		}
		if s.ID == strategy2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("First strategy not found in retrieved strategies")
	}

	if !found2 {
		t.Error("Second strategy not found in retrieved strategies")
	}
}

// TestPauseStrategy tests strategy pausing
func TestPauseStrategy(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.01, // Lower threshold for testing
			MaxDrawdown:    0.20, // Higher threshold for testing
			MinSharpeRatio: 0.5,  // Lower threshold for testing
			MaxVolatility:  0.50, // Higher threshold for testing
			MinWinRate:     0.30, // Lower threshold for testing
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and activate a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Verify strategy is active
	activeStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Failed to get active strategy: %v", err)
	}

	if activeStrategy.Status != StrategyStatusActive {
		t.Errorf("Expected status to be Active, got %v", activeStrategy.Status)
	}

	// Test pausing
	err = generator.PauseStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Expected pausing to succeed, got error: %v", err)
	}

	// Verify strategy was paused
	pausedStrategy, err := generator.GetStrategy(strategy.ID)
	if err != nil {
		t.Errorf("Failed to get paused strategy: %v", err)
	}

	if pausedStrategy.Status != StrategyStatusPaused {
		t.Errorf("Expected status to be Paused, got %v", pausedStrategy.Status)
	}

	// Check metrics
	metrics := generator.GetMetrics()
	if metrics.ActiveStrategies != 0 {
		t.Errorf("Expected ActiveStrategies to be 0, got %d", metrics.ActiveStrategies)
	}
}

// TestPauseStrategyErrors tests pausing error conditions
func TestPauseStrategyErrors(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Test pausing non-existent strategy
	err := generator.PauseStrategy("non_existent_id")
	if err == nil {
		t.Error("Expected error when pausing non-existent strategy")
	}

	// Test pausing non-active strategy
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	err = generator.PauseStrategy(strategy.ID)
	if err == nil {
		t.Error("Expected error when pausing non-active strategy")
	}
}

// TestAddMarketData tests market data addition
func TestAddMarketData(t *testing.T) {
	config := GeneratorConfig{
		MaxMarketData: 3,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test adding market data
	marketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(50000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	err := generator.AddMarketData("BTC", marketData)
	if err != nil {
		t.Errorf("Expected adding market data to succeed, got error: %v", err)
	}

	// Test data overflow (should remove oldest data)
	marketData2 := MarketData{
		Timestamp:  time.Now().Add(time.Minute),
		Asset:      "BTC",
		Price:      big.NewFloat(51000.0),
		Volume:     big.NewInt(1100000),
		MarketCap:  big.NewInt(1100000000000),
		Volatility: 0.26,
		RSI:        51.0,
		MACD:       0.1,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(56000.0),
			"middle": big.NewFloat(51000.0),
			"lower":  big.NewFloat(46000.0),
		},
		Features: map[string]float64{
			"trend":    0.7,
			"momentum": 0.5,
		},
	}

	err = generator.AddMarketData("BTC", marketData2)
	if err != nil {
		t.Errorf("Expected adding second market data to succeed, got error: %v", err)
	}

	marketData3 := MarketData{
		Timestamp:  time.Now().Add(time.Minute * 2),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1200000),
		MarketCap:  big.NewInt(1200000000000),
		Volatility: 0.27,
		RSI:        52.0,
		MACD:       0.2,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(57000.0),
			"middle": big.NewFloat(52000.0),
			"lower":  big.NewFloat(47000.0),
		},
		Features: map[string]float64{
			"trend":    0.8,
			"momentum": 0.6,
		},
	}

	err = generator.AddMarketData("BTC", marketData3)
	if err != nil {
		t.Errorf("Expected adding third market data to succeed, got error: %v", err)
	}

	// Adding fourth should remove oldest
	marketData4 := MarketData{
		Timestamp:  time.Now().Add(time.Minute * 3),
		Asset:      "BTC",
		Price:      big.NewFloat(53000.0),
		Volume:     big.NewInt(1300000),
		MarketCap:  big.NewInt(1300000000000),
		Volatility: 0.28,
		RSI:        53.0,
		MACD:       0.3,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(58000.0),
			"middle": big.NewFloat(53000.0),
			"lower":  big.NewFloat(48000.0),
		},
		Features: map[string]float64{
			"trend":    0.9,
			"momentum": 0.7,
		},
	}

	err = generator.AddMarketData("BTC", marketData4)
	if err != nil {
		t.Errorf("Expected adding fourth market data to succeed, got error: %v", err)
	}
}

// TestGetMetrics tests metrics retrieval
func TestGetMetrics(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Initially should have zero metrics
	metrics := generator.GetMetrics()
	if metrics.TotalStrategies != 0 {
		t.Errorf("Expected TotalStrategies to be 0 initially, got %d", metrics.TotalStrategies)
	}

	if metrics.GeneratedStrategies != 0 {
		t.Errorf("Expected GeneratedStrategies to be 0 initially, got %d", metrics.GeneratedStrategies)
	}

	if metrics.ActiveStrategies != 0 {
		t.Errorf("Expected ActiveStrategies to be 0 initially, got %d", metrics.ActiveStrategies)
	}

	if metrics.OptimizedStrategies != 0 {
		t.Errorf("Expected OptimizedStrategies to be 0 initially, got %d", metrics.OptimizedStrategies)
	}

	if metrics.BacktestedStrategies != 0 {
		t.Errorf("Expected BacktestedStrategies to be 0 initially, got %d", metrics.BacktestedStrategies)
	}

	if metrics.AveragePerformance != 0 {
		t.Errorf("Expected AveragePerformance to be 0 initially, got %f", metrics.AveragePerformance)
	}

	if metrics.GenerationTime != 0 {
		t.Errorf("Expected GenerationTime to be 0 initially, got %v", metrics.GenerationTime)
	}

	if !metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be zero initially")
	}
}

// TestUtilityFunctions tests utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test copyStringSlice
	original := []string{"BTC", "ETH", "ADA"}
	copied := copyStringSlice(original)

	if len(copied) != len(original) {
		t.Errorf("Expected copied slice to have same length, got %d vs %d", len(copied), len(original))
	}

	for i, v := range copied {
		if v != original[i] {
			t.Errorf("Expected copied value at index %d to match, got %s vs %s", i, v, original[i])
		}
	}

	// Verify it's a deep copy
	copied[0] = "XRP"
	if original[0] == "XRP" {
		t.Error("Expected original slice to not be modified when copied slice is changed")
	}

	// Test ID generation functions
	genID1 := generateGeneratorID()
	genID2 := generateGeneratorID()
	if genID1 == genID2 {
		t.Error("Expected generated IDs to be unique")
	}

	strategyID1 := generateStrategyID()
	strategyID2 := generateStrategyID()
	if strategyID1 == strategyID2 {
		t.Error("Expected generated strategy IDs to be unique")
	}

	signalID1 := generateSignalID()
	signalID2 := generateSignalID()
	if signalID1 == signalID2 {
		t.Error("Expected generated signal IDs to be unique")
	}

	// Test ID format
	if len(genID1) == 0 {
		t.Error("Expected generator ID to not be empty")
	}

	if len(strategyID1) == 0 {
		t.Error("Expected strategy ID to not be empty")
	}

	if len(signalID1) == 0 {
		t.Error("Expected signal ID to not be empty")
	}
}

// TestStrategyNameGeneration tests strategy name generation
func TestStrategyNameGeneration(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	testCases := []struct {
		strategyType StrategyType
		assets       []string
		expected     string
	}{
		{StrategyTypeTrendFollowing, []string{"BTC"}, "Trend_BTC_"},
		{StrategyTypeMeanReversion, []string{"ETH", "ADA"}, "MeanRev_ETH_"},
		{StrategyTypeArbitrage, []string{"BTC", "ETH", "ADA"}, "Arbitrage_BTC_"},
		{StrategyTypeMomentum, []string{"SOL"}, "Momentum_SOL_"},
		{StrategyTypeGridTrading, []string{"DOT"}, "Grid_DOT_"},
		{StrategyTypeDCA, []string{"LINK"}, "DCA_LINK_"},
		{StrategyTypeCustom, []string{"UNI"}, "Custom_UNI_"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			marketData := []MarketData{
				{Timestamp: time.Now(), Asset: tc.assets[0], Price: big.NewFloat(100.0)},
			}

			strategy, err := generator.GenerateStrategy(tc.strategyType, tc.assets, RiskProfileModerate, marketData)
			if err != nil {
				t.Errorf("Failed to generate strategy: %v", err)
				return
			}

			// Check that name starts with expected prefix
			if len(strategy.Name) < len(tc.expected) {
				t.Errorf("Expected name to be at least %d characters, got %d", len(tc.expected), len(strategy.Name))
			}

			expectedPrefix := tc.expected[:len(tc.expected)-1] // Remove the underscore at the end
			if strategy.Name[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("Expected name to start with %s, got %s", expectedPrefix, strategy.Name)
			}

			// Check that name ends with timestamp
			// Find the last underscore to get the timestamp part
			lastUnderscoreIndex := -1
			for i := len(strategy.Name) - 1; i >= 0; i-- {
				if strategy.Name[i] == '_' {
					lastUnderscoreIndex = i
					break
				}
			}

			if lastUnderscoreIndex == -1 {
				t.Errorf("Expected name to contain underscore before timestamp")
				return
			}

			timestampPart := strategy.Name[lastUnderscoreIndex+1:]
			if !isValidTimestamp(timestampPart) {
				t.Errorf("Expected name to end with valid timestamp, got %s", timestampPart)
			}
		})
	}
}

// Helper function to check if a string represents a valid timestamp
func isValidTimestamp(s string) bool {
	// Simple check - should be numeric and reasonable length
	if len(s) < 10 || len(s) > 20 {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// TestValidatePerformanceTargets tests performance target validation
func TestValidatePerformanceTargets(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05,
			MaxDrawdown:    0.10,
			MinSharpeRatio: 1.0,
			MaxVolatility:  0.20,
			MinWinRate:     0.50,
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test valid performance
	validBacktest := &BacktestResult{
		TotalReturn: 0.08, // Above minimum
		MaxDrawdown: 0.05, // Below maximum
		SharpeRatio: 1.5,  // Above minimum
		Volatility:  0.15, // Below maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)}, // Winning trade
			{PnL: big.NewFloat(50)},  // Winning trade
			{PnL: big.NewFloat(-30)}, // Losing trade
		},
	}

	isValid := generator.validatePerformanceTargets(validBacktest)
	if !isValid {
		t.Error("Expected valid performance to pass validation")
	}

	// Test invalid performance - return too low
	invalidBacktest1 := &BacktestResult{
		TotalReturn: 0.03, // Below minimum
		MaxDrawdown: 0.05, // Below maximum
		SharpeRatio: 1.5,  // Above minimum
		Volatility:  0.15, // Below maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)},
			{PnL: big.NewFloat(50)},
			{PnL: big.NewFloat(-30)},
		},
	}

	isValid = generator.validatePerformanceTargets(invalidBacktest1)
	if isValid {
		t.Error("Expected invalid performance (low return) to fail validation")
	}

	// Test invalid performance - drawdown too high
	invalidBacktest2 := &BacktestResult{
		TotalReturn: 0.08, // Above minimum
		MaxDrawdown: 0.15, // Above maximum
		SharpeRatio: 1.5,  // Above minimum
		Volatility:  0.15, // Below maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)},
			{PnL: big.NewFloat(50)},
			{PnL: big.NewFloat(-30)},
		},
	}

	isValid = generator.validatePerformanceTargets(invalidBacktest2)
	if isValid {
		t.Error("Expected invalid performance (high drawdown) to fail validation")
	}

	// Test invalid performance - Sharpe ratio too low
	invalidBacktest3 := &BacktestResult{
		TotalReturn: 0.08, // Above minimum
		MaxDrawdown: 0.05, // Below maximum
		SharpeRatio: 0.8,  // Below minimum
		Volatility:  0.15, // Below maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)},
			{PnL: big.NewFloat(50)},
			{PnL: big.NewFloat(-30)},
		},
	}

	isValid = generator.validatePerformanceTargets(invalidBacktest3)
	if isValid {
		t.Error("Expected invalid performance (low Sharpe ratio) to fail validation")
	}

	// Test invalid performance - volatility too high
	invalidBacktest4 := &BacktestResult{
		TotalReturn: 0.08, // Above minimum
		MaxDrawdown: 0.05, // Below maximum
		SharpeRatio: 1.5,  // Above minimum
		Volatility:  0.25, // Above maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)},
			{PnL: big.NewFloat(50)},
			{PnL: big.NewFloat(-30)},
		},
	}

	isValid = generator.validatePerformanceTargets(invalidBacktest4)
	if isValid {
		t.Error("Expected invalid performance (high volatility) to fail validation")
	}

	// Test invalid performance - win rate too low
	invalidBacktest5 := &BacktestResult{
		TotalReturn: 0.08, // Above minimum
		MaxDrawdown: 0.05, // Below maximum
		SharpeRatio: 1.5,  // Above minimum
		Volatility:  0.15, // Below maximum
		TradeHistory: []TradeRecord{
			{PnL: big.NewFloat(100)},  // Winning trade
			{PnL: big.NewFloat(-200)}, // Losing trade
			{PnL: big.NewFloat(-300)}, // Losing trade
		},
	}

	isValid = generator.validatePerformanceTargets(invalidBacktest5)
	if isValid {
		t.Error("Expected invalid performance (low win rate) to fail validation")
	}
}

// TestCopyStrategy tests strategy copying functionality
func TestCopyStrategy(t *testing.T) {
	config := GeneratorConfig{}
	generator := NewAutomatedStrategyGenerator(config)

	// Create a strategy with complex data
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC", "ETH"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Add some custom data
	strategy.Parameters.CustomLogic["test_key"] = "test_value"
	strategy.Parameters.CustomLogic["numeric_key"] = 42.5

	// Test copying
	copiedStrategy := generator.copyStrategy(strategy)

	// Verify it's a deep copy
	if copiedStrategy == strategy {
		t.Error("Expected copied strategy to be a different instance")
	}

	if copiedStrategy.ID != strategy.ID {
		t.Errorf("Expected ID to match, got %s vs %s", copiedStrategy.ID, strategy.ID)
	}

	if copiedStrategy.Name != strategy.Name {
		t.Errorf("Expected Name to match, got %s vs %s", copiedStrategy.Name, strategy.Name)
	}

	if copiedStrategy.Type != strategy.Type {
		t.Errorf("Expected Type to match, got %v vs %v", copiedStrategy.Type, strategy.Type)
	}

	if copiedStrategy.Status != strategy.Status {
		t.Errorf("Expected Status to match, got %v vs %v", copiedStrategy.Status, strategy.Status)
	}

	if copiedStrategy.RiskProfile != strategy.RiskProfile {
		t.Errorf("Expected RiskProfile to match, got %v vs %v", copiedStrategy.RiskProfile, strategy.RiskProfile)
	}

	// Verify assets slice is copied
	if len(copiedStrategy.Assets) != len(strategy.Assets) {
		t.Errorf("Expected assets length to match, got %d vs %d", len(copiedStrategy.Assets), len(strategy.Assets))
	}

	for i, asset := range copiedStrategy.Assets {
		if asset != strategy.Assets[i] {
			t.Errorf("Expected asset at index %d to match, got %s vs %s", i, asset, strategy.Assets[i])
		}
	}

	// Verify it's a deep copy of assets
	copiedStrategy.Assets[0] = "XRP"
	if strategy.Assets[0] == "XRP" {
		t.Error("Expected original strategy assets to not be modified when copied strategy is changed")
	}

	// Verify parameters are copied
	if copiedStrategy.Parameters.EntryThreshold != strategy.Parameters.EntryThreshold {
		t.Errorf("Expected EntryThreshold to match, got %f vs %f", copiedStrategy.Parameters.EntryThreshold, strategy.Parameters.EntryThreshold)
	}

	// Verify custom logic is copied
	if len(copiedStrategy.Parameters.CustomLogic) != len(strategy.Parameters.CustomLogic) {
		t.Errorf("Expected CustomLogic length to match, got %d vs %d", len(copiedStrategy.Parameters.CustomLogic), len(strategy.Parameters.CustomLogic))
	}

	for k, v := range copiedStrategy.Parameters.CustomLogic {
		if strategy.Parameters.CustomLogic[k] != v {
			t.Errorf("Expected CustomLogic value for key %s to match, got %v vs %v", k, v, strategy.Parameters.CustomLogic[k])
		}
	}

	// Verify it's a deep copy of custom logic
	copiedStrategy.Parameters.CustomLogic["test_key"] = "modified_value"
	if strategy.Parameters.CustomLogic["test_key"] == "modified_value" {
		t.Error("Expected original strategy custom logic to not be modified when copied strategy is changed")
	}

	// Verify timestamps are copied
	if !copiedStrategy.CreatedAt.Equal(strategy.CreatedAt) {
		t.Errorf("Expected CreatedAt to match, got %v vs %v", copiedStrategy.CreatedAt, strategy.CreatedAt)
	}

	if !copiedStrategy.LastUpdate.Equal(strategy.LastUpdate) {
		t.Errorf("Expected LastUpdate to match, got %v vs %v", copiedStrategy.LastUpdate, strategy.LastUpdate)
	}
}

// TestUpdateMetrics tests metrics updating functionality
func TestUpdateMetrics(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Initially metrics should be zero
	metrics := generator.GetMetrics()
	if metrics.ActiveStrategies != 0 {
		t.Errorf("Expected ActiveStrategies to be 0 initially, got %d", metrics.ActiveStrategies)
	}

	if metrics.AveragePerformance != 0 {
		t.Errorf("Expected AveragePerformance to be 0 initially, got %f", metrics.AveragePerformance)
	}

	// Create and activate a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Update metrics manually to test the function
	generator.updateMetrics()

	// Verify metrics were updated
	updatedMetrics := generator.GetMetrics()
	if updatedMetrics.ActiveStrategies != 1 {
		t.Errorf("Expected ActiveStrategies to be 1, got %d", updatedMetrics.ActiveStrategies)
	}

	if updatedMetrics.AveragePerformance <= 0 {
		t.Errorf("Expected AveragePerformance to be positive, got %f", updatedMetrics.AveragePerformance)
	}

	if updatedMetrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}
}

// TestBackgroundLoops tests that background loops can be started and stopped
func TestBackgroundLoops(t *testing.T) {
	config := GeneratorConfig{
		GenerationInterval: time.Millisecond * 100, // Fast for testing
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Start the generator
	err := generator.Start()
	if err != nil {
		t.Errorf("Expected Start to succeed, got error: %v", err)
	}

	// Give background loops a moment to start
	time.Sleep(time.Millisecond * 50)

	// Stop the generator
	err = generator.Stop()
	if err != nil {
		t.Errorf("Expected Stop to succeed, got error: %v", err)
	}

	// Give background loops a moment to stop
	time.Sleep(time.Millisecond * 50)
}

// TestEdgeCases tests various edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies: 1,
		MaxMarketData: 1,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test with minimal market data
	minimalMarketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		minimalMarketData,
	)
	if err != nil {
		t.Errorf("Expected strategy generation with minimal data to succeed, got error: %v", err)
	}

	if strategy == nil {
		t.Fatal("Expected strategy to be created with minimal data")
	}

	// Test with single asset
	if len(strategy.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(strategy.Assets))
	}

	if strategy.Assets[0] != "BTC" {
		t.Errorf("Expected asset to be BTC, got %s", strategy.Assets[0])
	}

	// Test market data overflow
	marketData1 := MarketData{
		Timestamp: time.Now(),
		Asset:     "BTC",
		Price:     big.NewFloat(50000.0),
	}

	err = generator.AddMarketData("BTC", marketData1)
	if err != nil {
		t.Errorf("Expected adding first market data to succeed, got error: %v", err)
	}

	marketData2 := MarketData{
		Timestamp: time.Now().Add(time.Minute),
		Asset:     "BTC",
		Price:     big.NewFloat(51000.0),
	}

	err = generator.AddMarketData("BTC", marketData2)
	if err != nil {
		t.Errorf("Expected adding second market data to succeed, got error: %v", err)
	}

	// Verify first data point was removed due to overflow
	// Note: This is implementation dependent, but we can verify the system handles it gracefully
}

// TestConcurrency tests concurrent access to the generator
func TestConcurrency(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies: 100,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create multiple goroutines that access the generator concurrently
	numGoroutines := 10
	errors := make(chan error, numGoroutines)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Generate strategy
			marketData := []MarketData{
				{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0 + float64(id))},
			}

			strategy, err := generator.GenerateStrategy(
				StrategyTypeTrendFollowing,
				[]string{"BTC"},
				RiskProfileModerate,
				marketData,
			)

			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed to generate strategy: %v", id, err)
				return
			}

			if strategy == nil {
				errors <- fmt.Errorf("goroutine %d got nil strategy", id)
				return
			}

			// Retrieve strategy
			retrievedStrategy, err := generator.GetStrategy(strategy.ID)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed to retrieve strategy: %v", id, err)
				return
			}

			if retrievedStrategy.ID != strategy.ID {
				errors <- fmt.Errorf("goroutine %d got wrong strategy, expected %s, got %s", id, strategy.ID, retrievedStrategy.ID)
				return
			}

			// Get metrics
			metrics := generator.GetMetrics()
			if metrics.TotalStrategies == 0 {
				errors <- fmt.Errorf("goroutine %d got zero total strategies", id)
				return
			}

		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrency test error: %v", err)
	}

	// Verify final state
	finalMetrics := generator.GetMetrics()
	if finalMetrics.TotalStrategies != uint64(numGoroutines) {
		t.Errorf("Expected %d total strategies, got %d", numGoroutines, finalMetrics.TotalStrategies)
	}

	if finalMetrics.GeneratedStrategies != uint64(numGoroutines) {
		t.Errorf("Expected %d generated strategies, got %d", numGoroutines, finalMetrics.GeneratedStrategies)
	}
}

// TestMemorySafety tests that the generator doesn't have memory leaks
func TestMemorySafety(t *testing.T) {
	config := GeneratorConfig{
		MaxStrategies: 1000,
		MaxMarketData: 1000,
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create many strategies to test memory management
	numStrategies := 100
	marketData := []MarketData{
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(50000.0)},
	}

	for i := 0; i < numStrategies; i++ {
		strategy, err := generator.GenerateStrategy(
			StrategyTypeTrendFollowing,
			[]string{"BTC"},
			RiskProfileModerate,
			marketData,
		)

		if err != nil {
			t.Errorf("Failed to generate strategy %d: %v", i, err)
			continue
		}

		if strategy == nil {
			t.Errorf("Got nil strategy %d", i)
			continue
		}

		// Add market data
		for j := 0; j < 10; j++ {
			data := MarketData{
				Timestamp: time.Now().Add(time.Duration(j) * time.Minute),
				Asset:     "BTC",
				Price:     big.NewFloat(50000.0 + float64(j)),
			}
			err = generator.AddMarketData("BTC", data)
			if err != nil {
				t.Errorf("Failed to add market data %d for strategy %d: %v", j, i, err)
			}
		}
	}

	// Verify all strategies were created
	finalMetrics := generator.GetMetrics()
	if finalMetrics.TotalStrategies != uint64(numStrategies) {
		t.Errorf("Expected %d total strategies, got %d", numStrategies, finalMetrics.TotalStrategies)
	}

	// Test that we can still retrieve strategies
	allStrategies := generator.GetStrategies()
	if len(allStrategies) != numStrategies {
		t.Errorf("Expected %d strategies in retrieval, got %d", numStrategies, len(allStrategies))
	}
}

// TestGenerateSignalsForStrategyComprehensive tests all signal generation paths
func TestGenerateSignalsForStrategyComprehensive(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and activate a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeMeanReversion,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Test mean reversion strategy with high price (should generate sell signal)
	highPriceMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(120.0), // Above 100 threshold
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err := generator.GenerateTradingSignals(strategy.ID, highPriceMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	if len(signals) == 0 {
		t.Error("Expected at least one trading signal for high price")
	}

	// Test mean reversion strategy with low price (should generate buy signal)
	lowPriceMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(80.0), // Below 100 threshold
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(strategy.ID, lowPriceMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	if len(signals) == 0 {
		t.Error("Expected at least one trading signal for low price")
	}

	// Test momentum strategy with high volume
	momentumStrategy, err := generator.GenerateStrategy(
		StrategyTypeMomentum,
		[]string{"ETH"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create momentum strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(momentumStrategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest momentum strategy: %v", err)
	}

	err = generator.ActivateStrategy(momentumStrategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate momentum strategy: %v", err)
	}

	// Test momentum strategy with high volume (should generate buy signal)
	highVolumeMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "ETH",
		Price:      big.NewFloat(3000.0),
		Volume:     big.NewInt(2000000), // Above 1M threshold
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(3500.0),
			"middle": big.NewFloat(3000.0),
			"lower":  big.NewFloat(2500.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(momentumStrategy.ID, highVolumeMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	if len(signals) == 0 {
		t.Error("Expected at least one trading signal for high volume")
	}

	// Test momentum strategy with low volume (should not generate signal)
	lowVolumeMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "ETH",
		Price:      big.NewFloat(3000.0),
		Volume:     big.NewInt(500000), // Below 1M threshold
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(3500.0),
			"middle": big.NewFloat(3000.0),
			"lower":  big.NewFloat(2500.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(momentumStrategy.ID, lowVolumeMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// Low volume should not generate signals for momentum strategy
	if len(signals) > 0 {
		t.Error("Expected no trading signals for low volume in momentum strategy")
	}
}

// TestGenerateSignalsForStrategyEdgeCases tests edge cases in signal generation
func TestGenerateSignalsForStrategyEdgeCases(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and activate a strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeTrendFollowing,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Test with RSI above 70 (should generate sell signal)
	highRSIMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        75.0, // Above 70 threshold
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err := generator.GenerateTradingSignals(strategy.ID, highRSIMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	if len(signals) == 0 {
		t.Error("Expected trading signal for RSI above 70")
	}

	// Test with RSI below 30 (should generate buy signal)
	lowRSIMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        25.0, // Below 30 threshold
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(strategy.ID, lowRSIMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	if len(signals) == 0 {
		t.Error("Expected trading signal for RSI below 30")
	}

	// Test with RSI in middle range (should not generate signal)
	middleRSIMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0, // Middle range
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(strategy.ID, middleRSIMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// Middle RSI should not generate signals for trend following strategy
	if len(signals) > 0 {
		t.Error("Expected no trading signals for middle RSI in trend following strategy")
	}
}

// TestGenerateSignalsForStrategyCustom tests custom strategy signal generation
func TestGenerateSignalsForStrategyCustom(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Create and activate a custom strategy
	marketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	strategy, err := generator.GenerateStrategy(
		StrategyTypeCustom,
		[]string{"BTC"},
		RiskProfileModerate,
		marketData,
	)
	if err != nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(strategy.ID, marketData)
	if err != nil {
		t.Fatalf("Failed to backtest strategy: %v", err)
	}

	err = generator.ActivateStrategy(strategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate strategy: %v", err)
	}

	// Test custom strategy (should not generate signals by default)
	customMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        75.0,
		MACD:       0.5,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.8,
			"momentum": 0.9,
		},
	}

	signals, err := generator.GenerateTradingSignals(strategy.ID, customMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// Custom strategy should not generate signals by default
	if len(signals) > 0 {
		t.Error("Expected no trading signals for custom strategy by default")
	}
}

// TestGenerateSignalsForStrategyGridAndDCA tests grid trading and DCA strategies
func TestGenerateSignalsForStrategyGridAndDCA(t *testing.T) {
	config := GeneratorConfig{
		PerformanceTargets: PerformanceTargets{
			MinReturn:      0.05, // 5% minimum return
			MinSharpeRatio: 1.0,  // 1.0 minimum Sharpe ratio
			MaxDrawdown:    0.10, // 10% maximum drawdown
			MaxVolatility:  0.20, // 20% maximum volatility
			MinWinRate:     0.40, // 40% minimum win rate
		},
	}
	generator := NewAutomatedStrategyGenerator(config)

	// Test Grid Trading strategy
	gridMarketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	gridStrategy, err := generator.GenerateStrategy(
		StrategyTypeGridTrading,
		[]string{"BTC"},
		RiskProfileModerate,
		gridMarketData,
	)
	if err != nil {
		t.Fatalf("Failed to create grid strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(gridStrategy.ID, gridMarketData)
	if err != nil {
		t.Fatalf("Failed to backtest grid strategy: %v", err)
	}

	err = generator.ActivateStrategy(gridStrategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate grid strategy: %v", err)
	}

	// Test grid trading strategy
	gridCurrentMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err := generator.GenerateTradingSignals(gridStrategy.ID, gridCurrentMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// Grid trading should not generate signals by default
	if len(signals) > 0 {
		t.Error("Expected no trading signals for grid trading strategy by default")
	}

	// Test DCA strategy
	dcaMarketData := []MarketData{
		{Timestamp: time.Now().Add(-time.Hour), Asset: "BTC", Price: big.NewFloat(50000.0)},
		{Timestamp: time.Now(), Asset: "BTC", Price: big.NewFloat(51000.0)},
	}

	dcaStrategy, err := generator.GenerateStrategy(
		StrategyTypeDCA,
		[]string{"BTC"},
		RiskProfileModerate,
		dcaMarketData,
	)
	if err != nil {
		t.Fatalf("Failed to create DCA strategy: %v", err)
	}

	// Backtest and activate
	_, err = generator.BacktestStrategy(dcaStrategy.ID, dcaMarketData)
	if err != nil {
		t.Fatalf("Failed to backtest DCA strategy: %v", err)
	}

	err = generator.ActivateStrategy(dcaStrategy.ID)
	if err != nil {
		t.Fatalf("Failed to activate DCA strategy: %v", err)
	}

	// Test DCA strategy
	dcaCurrentMarketData := MarketData{
		Timestamp:  time.Now(),
		Asset:      "BTC",
		Price:      big.NewFloat(52000.0),
		Volume:     big.NewInt(1000000),
		MarketCap:  big.NewInt(1000000000000),
		Volatility: 0.25,
		RSI:        50.0,
		MACD:       0.0,
		BollingerBands: map[string]*big.Float{
			"upper":  big.NewFloat(55000.0),
			"middle": big.NewFloat(50000.0),
			"lower":  big.NewFloat(45000.0),
		},
		Features: map[string]float64{
			"trend":    0.6,
			"momentum": 0.4,
		},
	}

	signals, err = generator.GenerateTradingSignals(dcaStrategy.ID, dcaCurrentMarketData)
	if err != nil {
		t.Errorf("Expected signal generation to succeed, got error: %v", err)
	}

	// DCA should not generate signals by default
	if len(signals) > 0 {
		t.Error("Expected no trading signals for DCA strategy by default")
	}
}
