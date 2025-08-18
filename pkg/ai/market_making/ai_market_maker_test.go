package market_making

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

// TestNewAIMarketMaker tests creation of new AI market maker
func TestNewAIMarketMaker(t *testing.T) {
	config := MarketMakerConfig{
		MaxPositions:        50,
		MinLiquidity:        big.NewInt(1000000000000000000),
		MaxSlippage:         0.03,
		RebalanceThreshold:  0.15,
		LearningRate:        0.02,
		UpdateInterval:      time.Minute * 10,
		EnableAutoRebalance: true,
		RiskManagement: RiskManagementConfig{
			MaxPositionSize:  0.1,
			MaxDrawdown:      0.2,
			CorrelationLimit: 0.7,
			VolatilityLimit:  0.3,
			LiquidityBuffer:  0.05,
		},
		MLConfig: MLModelConfig{
			ModelType:            "LSTM",
			TrainingInterval:     time.Hour,
			PredictionHorizon:    time.Hour * 24,
			FeatureWindow:        time.Hour * 168, // 1 week
			ConfidenceThreshold:  0.8,
			EnableOnlineLearning: true,
		},
	}

	mm := NewAIMarketMaker(config)

	if mm == nil {
		t.Fatal("Expected market maker to be created")
	}

	if mm.ID == "" {
		t.Error("Expected market maker ID to be set")
	}

	if mm.Strategy != StrategyAdaptive {
		t.Errorf("Expected default strategy %d, got %d", StrategyAdaptive, mm.Strategy)
	}

	if mm.RiskLevel != RiskLevelModerate {
		t.Errorf("Expected default risk level %d, got %d", RiskLevelModerate, mm.RiskLevel)
	}

	if len(mm.Positions) != 0 {
		t.Error("Expected empty positions map")
	}

	if len(mm.MarketStates) != 0 {
		t.Error("Expected empty market states map")
	}

	if mm.Config.MaxPositions != 50 {
		t.Errorf("Expected max positions 50, got %d", mm.Config.MaxPositions)
	}

	if mm.Config.MinLiquidity.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Error("Expected min liquidity to match config")
	}

	if mm.Config.MaxSlippage != 0.03 {
		t.Errorf("Expected max slippage 0.03, got %f", mm.Config.MaxSlippage)
	}
}

// TestNewAIMarketMakerDefaults tests default value initialization
func TestNewAIMarketMakerDefaults(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	if mm.Config.MaxPositions != 100 {
		t.Errorf("Expected default max positions 100, got %d", mm.Config.MaxPositions)
	}

	if mm.Config.MinLiquidity.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Error("Expected default min liquidity")
	}

	if mm.Config.MaxSlippage != 0.05 {
		t.Errorf("Expected default max slippage 0.05, got %f", mm.Config.MaxSlippage)
	}

	if mm.Config.RebalanceThreshold != 0.1 {
		t.Errorf("Expected default rebalance threshold 0.1, got %f", mm.Config.RebalanceThreshold)
	}

	if mm.Config.LearningRate != 0.01 {
		t.Errorf("Expected default learning rate 0.01, got %f", mm.Config.LearningRate)
	}

	if mm.Config.UpdateInterval != time.Minute*5 {
		t.Errorf("Expected default update interval 5 minutes, got %v", mm.Config.UpdateInterval)
	}
}

// TestCreatePosition tests position creation
func TestCreatePosition(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MaxPositions: 5,
		MinLiquidity: big.NewInt(1000000000000000000), // 1 ETH
	})

	amountA := big.NewInt(1000000000000000000) // 1 ETH
	amountB := big.NewInt(2000000000)          // 2000 USDC

	position, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Errorf("Failed to create position: %v", err)
	}

	if position == nil {
		t.Fatal("Expected position to be created")
	}

	if position.AssetA != "ETH" {
		t.Errorf("Expected asset A ETH, got %s", position.AssetA)
	}

	if position.AssetB != "USDC" {
		t.Errorf("Expected asset B USDC, got %s", position.AssetB)
	}

	if position.AmountA.Cmp(amountA) != 0 {
		t.Errorf("Expected amount A %s, got %s", amountA.String(), position.AmountA.String())
	}

	if position.AmountB.Cmp(amountB) != 0 {
		t.Errorf("Expected amount B %s, got %s", amountB.String(), position.AmountB.String())
	}

	if position.EntryPrice == nil {
		t.Error("Expected entry price to be calculated")
	}

	if position.RiskScore < 0 || position.RiskScore > 1 {
		t.Errorf("Expected risk score between 0 and 1, got %f", position.RiskScore)
	}

	if len(mm.Positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(mm.Positions))
	}
}

// TestCreatePositionValidation tests position creation validation
func TestCreatePositionValidation(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MaxPositions: 1,
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Test zero amounts
	_, err := mm.CreatePosition("ETH", "USDC", big.NewInt(0), big.NewInt(1000000000))
	if err == nil {
		t.Error("Expected error for zero amount A")
	}

	_, err = mm.CreatePosition("ETH", "USDC", big.NewInt(1000000000000000000), big.NewInt(0))
	if err == nil {
		t.Error("Expected error for zero amount B")
	}

	// Test negative amounts
	_, err = mm.CreatePosition("ETH", "USDC", big.NewInt(-1000000000000000000), big.NewInt(1000000000))
	if err == nil {
		t.Error("Expected error for negative amount A")
	}

	// Test insufficient liquidity
	_, err = mm.CreatePosition("ETH", "USDC", big.NewInt(100000000000000000), big.NewInt(100000000))
	if err == nil {
		t.Error("Expected error for insufficient liquidity")
	}

	// Test maximum positions reached
	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	// Create first position
	_, err = mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Errorf("Failed to create first position: %v", err)
	}

	// Try to create second position
	_, err = mm.CreatePosition("BTC", "USDC", amountA, amountB)
	if err == nil {
		t.Error("Expected error when maximum positions reached")
	}
}

// TestUpdatePosition tests position updates
func TestUpdatePosition(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	position, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Update position
	newAmountA := big.NewInt(1500000000000000000) // 1.5 ETH
	newAmountB := big.NewInt(3000000000)          // 3000 USDC

	err = mm.UpdatePosition(position.ID, newAmountA, newAmountB)
	if err != nil {
		t.Errorf("Failed to update position: %v", err)
	}

	// Verify update
	updatedPosition := mm.Positions[position.ID]
	if updatedPosition.AmountA.Cmp(newAmountA) != 0 {
		t.Errorf("Expected updated amount A %s, got %s", newAmountA.String(), updatedPosition.AmountA.String())
	}

	if updatedPosition.AmountB.Cmp(newAmountB) != 0 {
		t.Errorf("Expected updated amount B %s, got %s", newAmountB.String(), updatedPosition.AmountB.String())
	}

	// Test updating non-existent position
	err = mm.UpdatePosition("non-existent", newAmountA, newAmountB)
	if err == nil {
		t.Error("Expected error for non-existent position")
	}

	// Test updating with invalid amounts
	err = mm.UpdatePosition(position.ID, big.NewInt(0), newAmountB)
	if err == nil {
		t.Error("Expected error for zero amount A")
	}
}

// TestClosePosition tests position closure
func TestClosePosition(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	position, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Verify position exists
	if len(mm.Positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(mm.Positions))
	}

	// Close position
	pnl, err := mm.ClosePosition(position.ID)
	if err != nil {
		t.Errorf("Failed to close position: %v", err)
	}

	if pnl == nil {
		t.Error("Expected PnL to be returned")
	}

	// Verify position is removed
	if len(mm.Positions) != 0 {
		t.Errorf("Expected 0 positions after closure, got %d", len(mm.Positions))
	}

	// Test closing non-existent position
	_, err = mm.ClosePosition("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}
}

// TestOptimizeLiquidity tests liquidity optimization
func TestOptimizeLiquidity(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Add some market states
	mm.MarketStates["ETH-USDC"] = &MarketState{
		Price:           big.NewFloat(2000.0),
		Volume24h:       new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 ETH
		Volatility:      0.15,
		BidAskSpread:    0.002,
		LiquidityDepth:  new(big.Int).Mul(big.NewInt(500), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 500 ETH
		MarketSentiment: 0.7,
		LastUpdate:      time.Now(),
	}

	mm.MarketStates["BTC-USDC"] = &MarketState{
		Price:           big.NewFloat(50000.0),
		Volume24h:       new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil)), // 1000 BTC
		Volatility:      0.25,
		BidAskSpread:    0.003,
		LiquidityDepth:  new(big.Int).Mul(big.NewInt(20), new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil)), // 20 BTC
		MarketSentiment: 0.5,
		LastUpdate:      time.Now(),
	}

	// Test optimization
	err := mm.OptimizeLiquidity()
	if err != nil {
		t.Errorf("Failed to optimize liquidity: %v", err)
	}
}

// TestAdjustSpread tests spread adjustment
func TestAdjustSpread(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Add market state
	marketID := "ETH-USDC"
	mm.MarketStates[marketID] = &MarketState{
		Price:           big.NewFloat(2000.0),
		Volume24h:       new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		Volatility:      0.15,
		BidAskSpread:    0.002,
		LiquidityDepth:  new(big.Int).Mul(big.NewInt(500), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		MarketSentiment: 0.7,
		LastUpdate:      time.Now(),
	}

	// Test valid spread adjustment
	newSpread := 0.003
	err := mm.AdjustSpread(marketID, newSpread)
	if err != nil {
		t.Errorf("Failed to adjust spread: %v", err)
	}

	if mm.MarketStates[marketID].BidAskSpread != newSpread {
		t.Errorf("Expected spread %f, got %f", newSpread, mm.MarketStates[marketID].BidAskSpread)
	}

	// Test invalid spread values
	err = mm.AdjustSpread(marketID, -0.1)
	if err == nil {
		t.Error("Expected error for negative spread")
	}

	err = mm.AdjustSpread(marketID, 1.5)
	if err == nil {
		t.Error("Expected error for spread > 1")
	}

	// Test non-existent market
	err = mm.AdjustSpread("non-existent", 0.002)
	if err != nil {
		t.Errorf("Expected no error for non-existent market, got %v", err)
	}
}

// TestGetMetrics tests metrics retrieval
func TestGetMetrics(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Create and close a position to generate metrics
	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	position, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Close position to generate PnL
	_, err = mm.ClosePosition(position.ID)
	if err != nil {
		t.Fatalf("Failed to close position: %v", err)
	}

	// Get metrics
	metrics := mm.GetMetrics()

	if metrics.TotalTrades != 1 {
		t.Errorf("Expected 1 total trade, got %d", metrics.TotalTrades)
	}

	if metrics.TotalPnL == nil {
		t.Error("Expected total PnL to be set")
	}

	if metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestGetPositions tests position retrieval
func TestGetPositions(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Create multiple positions
	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	position1, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
	if err != nil {
		t.Fatalf("Failed to create position 1: %v", err)
	}

	position2, err := mm.CreatePosition("BTC", "USDC", amountA, amountB)
	if err != nil {
		t.Fatalf("Failed to create position 2: %v", err)
	}

	// Get positions
	positions := mm.GetPositions()

	if len(positions) != 2 {
		t.Errorf("Expected 2 positions, got %d", len(positions))
	}

	// Verify position 1
	pos1, exists := positions[position1.ID]
	if !exists {
		t.Error("Expected position 1 to exist")
	}

	if pos1.AssetA != "ETH" {
		t.Errorf("Expected position 1 asset A ETH, got %s", pos1.AssetA)
	}

	// Verify position 2
	pos2, exists := positions[position2.ID]
	if !exists {
		t.Error("Expected position 2 to exist")
	}

	if pos2.AssetA != "BTC" {
		t.Errorf("Expected position 2 asset A BTC, got %s", pos2.AssetA)
	}
}

// TestStartStop tests market maker start and stop functionality
func TestStartStop(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		UpdateInterval: time.Millisecond * 100,
		MLConfig: MLModelConfig{
			TrainingInterval: time.Millisecond * 100,
		},
	})

	// Start market maker
	err := mm.Start()
	if err != nil {
		t.Errorf("Failed to start market maker: %v", err)
	}

	// Wait a bit for goroutines to start
	time.Sleep(time.Millisecond * 50)

	// Stop market maker
	err = mm.Stop()
	if err != nil {
		t.Errorf("Failed to stop market maker: %v", err)
	}

	// Wait a bit for goroutines to stop
	time.Sleep(time.Millisecond * 50)
}

// TestConcurrency tests concurrent operations
func TestConcurrency(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MaxPositions: 100,
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Test concurrent position creation
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			amountA := big.NewInt(1000000000000000000)
			amountB := big.NewInt(2000000000)

			_, err := mm.CreatePosition("ETH", "USDC", amountA, amountB)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for any errors
	select {
	case err := <-errors:
		t.Errorf("Error in concurrent operation: %v", err)
	default:
		// No errors
	}

	// Verify positions were created
	positions := mm.GetPositions()
	if len(positions) != 10 {
		t.Errorf("Expected 10 positions, got %d", len(positions))
	}
}

// TestRiskCalculation tests risk score calculations
func TestRiskCalculation(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Test different position sizes
	testCases := []struct {
		amountA      *big.Int
		amountB      *big.Int
		price        *big.Float
		expectedRisk float64
	}{
		{
			amountA:      big.NewInt(1000000000000000000), // 1 ETH
			amountB:      big.NewInt(2000000000),          // 2000 USDC
			price:        big.NewFloat(2000.0),
			expectedRisk: 1.0, // Maximum risk for large ETH position
		},
		{
			amountA:      new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),   // 100 ETH
			amountB:      new(big.Int).Mul(big.NewInt(200000), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)), // 200,000 USDC
			price:        big.NewFloat(2000.0),
			expectedRisk: 1.0, // Maximum risk for very large ETH position
		},
	}

	for i, tc := range testCases {
		risk := mm.calculatePositionRisk(tc.amountA, tc.amountB, tc.price)

		if risk < 0 || risk > 1 {
			t.Errorf("Test case %d: Expected risk between 0 and 1, got %f", i, risk)
		}

		if risk < tc.expectedRisk*0.5 || risk > tc.expectedRisk*2.0 {
			t.Errorf("Test case %d: Expected risk around %f, got %f", i, tc.expectedRisk, risk)
		}
	}
}

// TestPnLCalculation tests PnL calculations
func TestPnLCalculation(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Create a position
	position := &Position{
		ID:           "test",
		AssetA:       "ETH",
		AssetB:       "USDC",
		AmountA:      big.NewInt(1000000000000000000), // 1 ETH
		AmountB:      big.NewInt(2000000000),          // 2000 USDC
		EntryPrice:   big.NewFloat(2000.0),
		CurrentPrice: big.NewFloat(2100.0), // Price increased by 5%
		CreatedAt:    time.Now(),
		LastUpdate:   time.Now(),
	}

	// Calculate PnL
	pnl := mm.calculatePositionPnL(position)

	if pnl == nil {
		t.Fatal("Expected PnL to be calculated")
	}

	// PnL should be positive (price increased)
	if pnl.Sign() <= 0 {
		t.Error("Expected positive PnL for price increase")
	}

	// Test negative PnL
	position.CurrentPrice = big.NewFloat(1900.0) // Price decreased by 5%
	pnl = mm.calculatePositionPnL(position)

	if pnl.Sign() >= 0 {
		t.Error("Expected negative PnL for price decrease")
	}
}

// TestSpreadCalculation tests average spread calculations
func TestSpreadCalculation(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Test with no market states
	avgSpread := mm.calculateAverageSpread()
	if avgSpread != 0.0 {
		t.Errorf("Expected 0.0 spread for no markets, got %f", avgSpread)
	}

	// Add market states
	mm.MarketStates["ETH-USDC"] = &MarketState{
		BidAskSpread: 0.002,
	}
	mm.MarketStates["BTC-USDC"] = &MarketState{
		BidAskSpread: 0.003,
	}

	avgSpread = mm.calculateAverageSpread()
	expectedSpread := (0.002 + 0.003) / 2.0

	if avgSpread != expectedSpread {
		t.Errorf("Expected average spread %f, got %f", expectedSpread, avgSpread)
	}
}

// TestMetricsUpdate tests metrics updates
func TestMetricsUpdate(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Set initial metrics
	mm.Metrics.TotalTrades = 5
	mm.Metrics.SuccessfulTrades = 4

	// Update metrics
	mm.updateMetrics()

	// Verify win rate calculation
	expectedWinRate := 4.0 / 5.0
	if mm.Metrics.WinRate != expectedWinRate {
		t.Errorf("Expected win rate %f, got %f", expectedWinRate, mm.Metrics.WinRate)
	}

	// Verify last update
	if mm.Metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestSharpeRatioCalculation tests Sharpe ratio calculations
func TestSharpeRatioCalculation(t *testing.T) {
	mm := NewAIMarketMaker(MarketMakerConfig{})

	// Test with zero PnL
	sharpe := mm.calculateSharpeRatio()
	if sharpe != 0.0 {
		t.Errorf("Expected Sharpe ratio 0.0 for zero PnL, got %f", sharpe)
	}

	// Test with positive PnL
	mm.Metrics.TotalPnL = big.NewFloat(1000.0)
	sharpe = mm.calculateSharpeRatio()

	if sharpe <= 0 {
		t.Error("Expected positive Sharpe ratio for positive PnL")
	}
}

// Benchmark tests for performance
func BenchmarkCreatePosition(b *testing.B) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MaxPositions: 1000,
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.CreatePosition("ETH", "USDC", amountA, amountB)
	}
}

func BenchmarkUpdatePosition(b *testing.B) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MaxPositions: 1000,
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000)

	position, _ := mm.CreatePosition("ETH", "USDC", amountA, amountB)

	newAmountA := big.NewInt(1500000000000000000)
	newAmountB := big.NewInt(3000000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.UpdatePosition(position.ID, newAmountA, newAmountB)
	}
}

func BenchmarkOptimizeLiquidity(b *testing.B) {
	mm := NewAIMarketMaker(MarketMakerConfig{
		MinLiquidity: big.NewInt(1000000000000000000),
	})

	// Add market states
	for i := 0; i < 10; i++ {
		marketID := fmt.Sprintf("TOKEN%d-USDC", i)
		mm.MarketStates[marketID] = &MarketState{
			Price:           big.NewFloat(100.0 + float64(i)),
			Volume24h:       new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
			Volatility:      0.15,
			BidAskSpread:    0.002,
			LiquidityDepth:  new(big.Int).Mul(big.NewInt(500), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
			MarketSentiment: 0.7,
			LastUpdate:      time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.OptimizeLiquidity()
	}
}
