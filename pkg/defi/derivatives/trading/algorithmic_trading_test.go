package trading

import (
	"math/big"
	"testing"
	"time"
)

// Helper function to create big.Float from string to avoid overflow issues
func bigFloatFromString(s string) *big.Float {
	v, _ := new(big.Float).SetString(s)
	return v
}

// Helper function to compare big.Float values with tolerance
func compareBigFloat(t *testing.T, expected, actual *big.Float, tolerance float64, message string) {
	expectedVal, _ := expected.Float64()
	actualVal, _ := actual.Float64()
	
	if abs(expectedVal-actualVal) > tolerance {
		t.Errorf("%s: expected %.6f, got %.6f", message, expectedVal, actualVal)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestNewTradingEngine(t *testing.T) {
	engine := NewTradingEngine()
	
	if engine == nil {
		t.Fatal("NewTradingEngine returned nil")
	}
	
	if engine.Strategies == nil {
		t.Error("Strategies map not initialized")
	}
	if engine.Orders == nil {
		t.Error("Orders map not initialized")
	}
	if engine.Trades == nil {
		t.Error("Trades map not initialized")
	}
	if engine.MarketData == nil {
		t.Error("MarketData map not initialized")
	}
	
	// Check timestamps are set
	if engine.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if engine.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewStrategy(t *testing.T) {
	params := map[string]*big.Float{
		"param1": big.NewFloat(1.5),
		"param2": big.NewFloat(2.0),
	}
	
	strategy, err := NewStrategy("test_id", "Test Strategy", "Test Description", MarketMaking, params)
	if err != nil {
		t.Fatalf("NewStrategy failed: %v", err)
	}
	
	if strategy.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got '%s'", strategy.ID)
	}
	if strategy.Name != "Test Strategy" {
		t.Errorf("Expected name 'Test Strategy', got '%s'", strategy.Name)
	}
	if strategy.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", strategy.Description)
	}
	if strategy.Type != MarketMaking {
		t.Errorf("Expected type MarketMaking, got %d", strategy.Type)
	}
	if strategy.Status != Inactive {
		t.Errorf("Expected status Inactive, got %d", strategy.Status)
	}
	if strategy.Parameters == nil {
		t.Error("Parameters not set")
	}
	if strategy.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if strategy.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewStrategyValidation(t *testing.T) {
	// Test empty ID
	_, err := NewStrategy("", "Test", "Description", MarketMaking, nil)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty name
	_, err = NewStrategy("test_id", "", "Description", MarketMaking, nil)
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test nil parameters (should work, creates empty map)
	strategy, err := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	if err != nil {
		t.Errorf("Unexpected error for nil parameters: %v", err)
	}
	if strategy.Parameters == nil {
		t.Error("Parameters should be initialized as empty map")
	}
}

func TestNewMarketMaker(t *testing.T) {
	spread := big.NewFloat(0.02)      // 2%
	inventoryTarget := big.NewFloat(100)
	maxInventory := big.NewFloat(200)
	minOrderSize := big.NewFloat(1)
	maxOrderSize := big.NewFloat(10)
	rebalanceThreshold := big.NewFloat(0.1)
	
	mm, err := NewMarketMaker("mm_id", "Market Maker", "Test MM", spread, inventoryTarget, maxInventory, minOrderSize, maxOrderSize, rebalanceThreshold)
	if err != nil {
		t.Fatalf("NewMarketMaker failed: %v", err)
	}
	
	if mm.ID != "mm_id" {
		t.Errorf("Expected ID 'mm_id', got '%s'", mm.ID)
	}
	if mm.Type != MarketMaking {
		t.Errorf("Expected type MarketMaking, got %d", mm.Type)
	}
	if mm.SpreadPercentage.Cmp(spread) != 0 {
		t.Error("SpreadPercentage not set correctly")
	}
	if mm.InventoryTarget.Cmp(inventoryTarget) != 0 {
		t.Error("InventoryTarget not set correctly")
	}
	if mm.MaxInventory.Cmp(maxInventory) != 0 {
		t.Error("MaxInventory not set correctly")
	}
	if mm.MinOrderSize.Cmp(minOrderSize) != 0 {
		t.Error("MinOrderSize not set correctly")
	}
	if mm.MaxOrderSize.Cmp(maxOrderSize) != 0 {
		t.Error("MaxOrderSize not set correctly")
	}
	if mm.RebalanceThreshold.Cmp(rebalanceThreshold) != 0 {
		t.Error("RebalanceThreshold not set correctly")
	}
}

func TestNewMarketMakerValidation(t *testing.T) {
	validParams := map[string]*big.Float{
		"spread":           big.NewFloat(0.02),
		"inventoryTarget":  big.NewFloat(100),
		"maxInventory":     big.NewFloat(200),
		"minOrderSize":     big.NewFloat(1),
		"maxOrderSize":     big.NewFloat(10),
		"rebalanceThreshold": big.NewFloat(0.1),
	}
	
	// Test nil spread
	_, err := NewMarketMaker("id", "name", "desc", nil, validParams["inventoryTarget"], validParams["maxInventory"], validParams["minOrderSize"], validParams["maxOrderSize"], validParams["rebalanceThreshold"])
	if err == nil {
		t.Error("Expected error for nil spread")
	}
	
	// Test negative spread
	_, err = NewMarketMaker("id", "name", "desc", big.NewFloat(-0.02), validParams["inventoryTarget"], validParams["maxInventory"], validParams["minOrderSize"], validParams["maxOrderSize"], validParams["rebalanceThreshold"])
	if err == nil {
		t.Error("Expected error for negative spread")
	}
	
	// Test zero spread
	_, err = NewMarketMaker("id", "name", "desc", big.NewFloat(0), validParams["inventoryTarget"], validParams["maxInventory"], validParams["minOrderSize"], validParams["maxOrderSize"], validParams["rebalanceThreshold"])
	if err == nil {
		t.Error("Expected error for zero spread")
	}
}

func TestNewArbitrageStrategy(t *testing.T) {
	minProfit := big.NewFloat(0.01)  // 1%
	maxSlippage := big.NewFloat(0.005) // 0.5%
	executionDelay := 100 * time.Millisecond
	maxConcurrent := 5
	
	as, err := NewArbitrageStrategy("arb_id", "Arbitrage", "Test Arb", minProfit, maxSlippage, executionDelay, maxConcurrent)
	if err != nil {
		t.Fatalf("NewArbitrageStrategy failed: %v", err)
	}
	
	if as.ID != "arb_id" {
		t.Errorf("Expected ID 'arb_id', got '%s'", as.ID)
	}
	if as.Type != Arbitrage {
		t.Errorf("Expected type Arbitrage, got %d", as.Type)
	}
	if as.MinProfitThreshold.Cmp(minProfit) != 0 {
		t.Error("MinProfitThreshold not set correctly")
	}
	if as.MaxSlippage.Cmp(maxSlippage) != 0 {
		t.Error("MaxSlippage not set correctly")
	}
	if as.ExecutionDelay != executionDelay {
		t.Errorf("Expected execution delay %v, got %v", executionDelay, as.ExecutionDelay)
	}
	if as.MaxConcurrentTrades != maxConcurrent {
		t.Errorf("Expected max concurrent trades %d, got %d", maxConcurrent, as.MaxConcurrentTrades)
	}
}

func TestNewArbitrageStrategyValidation(t *testing.T) {
	validParams := map[string]*big.Float{
		"minProfit": big.NewFloat(0.01),
		"maxSlippage": big.NewFloat(0.005),
	}
	
	// Test nil minProfit
	_, err := NewArbitrageStrategy("id", "name", "desc", nil, validParams["maxSlippage"], 100*time.Millisecond, 5)
	if err == nil {
		t.Error("Expected error for nil minProfit")
	}
	
	// Test negative execution delay
	_, err = NewArbitrageStrategy("id", "name", "desc", validParams["minProfit"], validParams["maxSlippage"], -100*time.Millisecond, 5)
	if err == nil {
		t.Error("Expected error for negative execution delay")
	}
	
	// Test zero max concurrent trades
	_, err = NewArbitrageStrategy("id", "name", "desc", validParams["minProfit"], validParams["maxSlippage"], 100*time.Millisecond, 0)
	if err == nil {
		t.Error("Expected error for zero max concurrent trades")
	}
}

func TestNewOrder(t *testing.T) {
	size := big.NewFloat(10)
	price := big.NewFloat(100)
	
	order, err := NewOrder("order_id", "strategy_id", "asset_id", Buy, size, price, Limit)
	if err != nil {
		t.Fatalf("NewOrder failed: %v", err)
	}
	
	if order.ID != "order_id" {
		t.Errorf("Expected ID 'order_id', got '%s'", order.ID)
	}
	if order.StrategyID != "strategy_id" {
		t.Errorf("Expected StrategyID 'strategy_id', got '%s'", order.StrategyID)
	}
	if order.AssetID != "asset_id" {
		t.Errorf("Expected AssetID 'asset_id', got '%s'", order.AssetID)
	}
	if order.Side != Buy {
		t.Errorf("Expected side Buy, got %d", order.Side)
	}
	if order.Size.Cmp(size) != 0 {
		t.Error("Size not set correctly")
	}
	if order.Price.Cmp(price) != 0 {
		t.Error("Price not set correctly")
	}
	if order.OrderType != Limit {
		t.Errorf("Expected order type Limit, got %d", order.OrderType)
	}
	if order.Status != New {
		t.Errorf("Expected status New, got %d", order.Status)
	}
	if order.Timestamp.IsZero() {
		t.Error("Timestamp not set")
	}
	if order.Expiry.IsZero() {
		t.Error("Expiry not set")
	}
	if order.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
	
	// Check expiry is 24 hours from timestamp
	expectedExpiry := order.Timestamp.Add(24 * time.Hour)
	if !order.Expiry.Equal(expectedExpiry) {
		t.Errorf("Expected expiry %v, got %v", expectedExpiry, order.Expiry)
	}
}

func TestNewOrderValidation(t *testing.T) {
	validSize := big.NewFloat(10)
	validPrice := big.NewFloat(100)
	
	// Test empty ID
	_, err := NewOrder("", "strategy_id", "asset_id", Buy, validSize, validPrice, Limit)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty strategy ID
	_, err = NewOrder("order_id", "", "asset_id", Buy, validSize, validPrice, Limit)
	if err == nil {
		t.Error("Expected error for empty strategy ID")
	}
	
	// Test empty asset ID
	_, err = NewOrder("order_id", "strategy_id", "", Buy, validSize, validPrice, Limit)
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
	
	// Test nil size
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, nil, validPrice, Limit)
	if err == nil {
		t.Error("Expected error for nil size")
	}
	
	// Test negative size
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, big.NewFloat(-10), validPrice, Limit)
	if err == nil {
		t.Error("Expected error for negative size")
	}
	
	// Test zero size
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, big.NewFloat(0), validPrice, Limit)
	if err == nil {
		t.Error("Expected error for zero size")
	}
	
	// Test nil price
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, validSize, nil, Limit)
	if err == nil {
		t.Error("Expected error for nil price")
	}
	
	// Test negative price
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, validSize, big.NewFloat(-100), Limit)
	if err == nil {
		t.Error("Expected error for negative price")
	}
	
	// Test zero price
	_, err = NewOrder("order_id", "strategy_id", "asset_id", Buy, validSize, big.NewFloat(0), Limit)
	if err == nil {
		t.Error("Expected error for zero price")
	}
}

func TestTradingEngineAddStrategy(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	
	err := engine.AddStrategy(strategy)
	if err != nil {
		t.Fatalf("AddStrategy failed: %v", err)
	}
	
	if _, exists := engine.Strategies["test_id"]; !exists {
		t.Error("Strategy not added to engine")
	}
	
	// Test adding duplicate strategy
	err = engine.AddStrategy(strategy)
	if err == nil {
		t.Error("Expected error for duplicate strategy")
	}
	
	// Test adding nil strategy
	err = engine.AddStrategy(nil)
	if err == nil {
		t.Error("Expected error for nil strategy")
	}
}

func TestTradingEngineRemoveStrategy(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	engine.AddStrategy(strategy)
	
	err := engine.RemoveStrategy("test_id")
	if err != nil {
		t.Fatalf("RemoveStrategy failed: %v", err)
	}
	
	if _, exists := engine.Strategies["test_id"]; exists {
		t.Error("Strategy not removed from engine")
	}
	
	// Test removing non-existent strategy
	err = engine.RemoveStrategy("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
	
	// Test removing with empty ID
	err = engine.RemoveStrategy("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestTradingEngineStartStopStrategy(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	engine.AddStrategy(strategy)
	
	// Test starting strategy
	err := engine.StartStrategy("test_id")
	if err != nil {
		t.Fatalf("StartStrategy failed: %v", err)
	}
	
	if engine.Strategies["test_id"].Status != Active {
		t.Error("Strategy status not set to Active")
	}
	
	// Test stopping strategy
	err = engine.StopStrategy("test_id")
	if err != nil {
		t.Fatalf("StopStrategy failed: %v", err)
	}
	
	if engine.Strategies["test_id"].Status != Stopped {
		t.Error("Strategy status not set to Stopped")
	}
	
	// Test starting non-existent strategy
	err = engine.StartStrategy("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
}

func TestTradingEnginePlaceOrder(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	engine.AddStrategy(strategy)
	engine.StartStrategy("test_id")
	
	order, _ := NewOrder("order_id", "test_id", "asset_id", Buy, big.NewFloat(10), big.NewFloat(100), Limit)
	
	err := engine.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	
	if _, exists := engine.Orders["order_id"]; !exists {
		t.Error("Order not added to engine")
	}
	
	// Test placing duplicate order
	err = engine.PlaceOrder(order)
	if err == nil {
		t.Error("Expected error for duplicate order")
	}
	
	// Test placing order with non-existent strategy
	order2, _ := NewOrder("order_id2", "non_existent", "asset_id", Buy, big.NewFloat(10), big.NewFloat(100), Limit)
	err = engine.PlaceOrder(order2)
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
	
	// Test placing order with inactive strategy
	engine.StopStrategy("test_id")
	order3, _ := NewOrder("order_id3", "test_id", "asset_id", Buy, big.NewFloat(10), big.NewFloat(100), Limit)
	err = engine.PlaceOrder(order3)
	if err == nil {
		t.Error("Expected error for inactive strategy")
	}
	
	// Test placing nil order
	err = engine.PlaceOrder(nil)
	if err == nil {
		t.Error("Expected error for nil order")
	}
}

func TestTradingEngineCancelOrder(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	engine.AddStrategy(strategy)
	engine.StartStrategy("test_id")
	
	order, _ := NewOrder("order_id", "test_id", "asset_id", Buy, big.NewFloat(10), big.NewFloat(100), Limit)
	engine.PlaceOrder(order)
	
	err := engine.CancelOrder("order_id")
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	
	if engine.Orders["order_id"].Status != Cancelled {
		t.Error("Order status not set to Cancelled")
	}
	
	// Test cancelling non-existent order
	err = engine.CancelOrder("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent order")
	}
	
	// Test cancelling with empty ID
	err = engine.CancelOrder("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestTradingEngineUpdateMarketData(t *testing.T) {
	engine := NewTradingEngine()
	
	marketData := &MarketData{
		AssetID:   "asset_id",
		BidPrice:  big.NewFloat(99),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	err := engine.UpdateMarketData(marketData)
	if err != nil {
		t.Fatalf("UpdateMarketData failed: %v", err)
	}
	
	if _, exists := engine.MarketData["asset_id"]; !exists {
		t.Error("Market data not added to engine")
	}
	
	// Test updating with nil market data
	err = engine.UpdateMarketData(nil)
	if err == nil {
		t.Error("Expected error for nil market data")
	}
	
	// Test updating with empty asset ID
	marketData2 := &MarketData{
		AssetID:   "",
		BidPrice:  big.NewFloat(99),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	err = engine.UpdateMarketData(marketData2)
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
}

func TestMarketMakerExecuteMarketMaking(t *testing.T) {
	mm, _ := NewMarketMaker("mm_id", "Market Maker", "Test MM", 
		big.NewFloat(0.02), big.NewFloat(100), big.NewFloat(200), 
		big.NewFloat(1), big.NewFloat(10), big.NewFloat(0.1))
	
	marketData := &MarketData{
		AssetID:   "asset_id",
		BidPrice:  big.NewFloat(99),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	currentInventory := big.NewFloat(150) // Above target
	
	orders, err := mm.ExecuteMarketMaking(marketData, currentInventory)
	if err != nil {
		t.Fatalf("ExecuteMarketMaking failed: %v", err)
	}
	
	if len(orders) == 0 {
		t.Error("Expected orders to be generated")
	}
	
	// Test with nil market data
	_, err = mm.ExecuteMarketMaking(nil, currentInventory)
	if err == nil {
		t.Error("Expected error for nil market data")
	}
	
	// Test with nil inventory
	_, err = mm.ExecuteMarketMaking(marketData, nil)
	if err == nil {
		t.Error("Expected error for nil inventory")
	}
}

func TestArbitrageStrategyDetectArbitrage(t *testing.T) {
	as, _ := NewArbitrageStrategy("arb_id", "Arbitrage", "Test Arb", 
		big.NewFloat(0.01), big.NewFloat(0.005), 100*time.Millisecond, 5)
	
	marketData1 := &MarketData{
		AssetID:   "asset1",
		BidPrice:  big.NewFloat(100),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100.5),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	marketData2 := &MarketData{
		AssetID:   "asset2",
		BidPrice:  big.NewFloat(102), // Higher bid than ask1
		AskPrice:  big.NewFloat(103),
		LastPrice: big.NewFloat(102.5),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	opportunity, err := as.DetectArbitrage(marketData1, marketData2)
	if err != nil {
		t.Fatalf("DetectArbitrage failed: %v", err)
	}
	
	if opportunity == nil {
		t.Error("Expected arbitrage opportunity to be detected")
	}
	
	// Test with nil market data
	_, err = as.DetectArbitrage(nil, marketData2)
	if err == nil {
		t.Error("Expected error for nil market data")
	}
	
	_, err = as.DetectArbitrage(marketData1, nil)
	if err == nil {
		t.Error("Expected error for nil market data")
	}
}

func TestArbitrageStrategyExecuteArbitrage(t *testing.T) {
	as, _ := NewArbitrageStrategy("arb_id", "Arbitrage", "Test Arb", 
		big.NewFloat(0.01), big.NewFloat(0.005), 100*time.Millisecond, 5)
	
	opportunity := &ArbitrageOpportunity{
		Asset1ID:  "asset1",
		Asset2ID:  "asset2",
		BuyAsset:  "asset1",
		SellAsset: "asset2",
		BuyPrice:  big.NewFloat(101),
		SellPrice: big.NewFloat(102),
		Profit:    big.NewFloat(1),
		Timestamp: time.Now(),
	}
	
	orders, err := as.ExecuteArbitrage(opportunity)
	if err != nil {
		t.Fatalf("ExecuteArbitrage failed: %v", err)
	}
	
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
	
	// Test with nil opportunity
	_, err = as.ExecuteArbitrage(nil)
	if err == nil {
		t.Error("Expected error for nil opportunity")
	}
	
	// Test with expired opportunity
	expiredOpportunity := &ArbitrageOpportunity{
		Asset1ID:  "asset1",
		Asset2ID:  "asset2",
		BuyAsset:  "asset1",
		SellAsset: "asset2",
		BuyPrice:  big.NewFloat(101),
		SellPrice: big.NewFloat(102),
		Profit:    big.NewFloat(1),
		Timestamp: time.Now().Add(-200 * time.Millisecond), // Expired
	}
	_, err = as.ExecuteArbitrage(expiredOpportunity)
	if err == nil {
		t.Error("Expected error for expired opportunity")
	}
}

func TestTradingEngineGetters(t *testing.T) {
	engine := NewTradingEngine()
	strategy, _ := NewStrategy("test_id", "Test", "Description", MarketMaking, nil)
	engine.AddStrategy(strategy)
	
	// Test GetStrategy
	retrievedStrategy, err := engine.GetStrategy("test_id")
	if err != nil {
		t.Fatalf("GetStrategy failed: %v", err)
	}
	if retrievedStrategy != strategy {
		t.Error("GetStrategy returned different strategy")
	}
	
	// Test GetStrategy with non-existent ID
	_, err = engine.GetStrategy("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent strategy")
	}
	
	// Test GetStrategy with empty ID
	_, err = engine.GetStrategy("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test GetActiveStrategies
	engine.StartStrategy("test_id")
	activeStrategies := engine.GetActiveStrategies()
	if len(activeStrategies) != 1 {
		t.Errorf("Expected 1 active strategy, got %d", len(activeStrategies))
	}
	
	// Test GetMarketData
	marketData := &MarketData{
		AssetID:   "asset_id",
		BidPrice:  big.NewFloat(99),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	engine.UpdateMarketData(marketData)
	
	retrievedMarketData, err := engine.GetMarketData("asset_id")
	if err != nil {
		t.Fatalf("GetMarketData failed: %v", err)
	}
	if retrievedMarketData != marketData {
		t.Error("GetMarketData returned different market data")
	}
	
	// Test GetMarketData with non-existent asset
	_, err = engine.GetMarketData("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
}

func TestGenerateOrderID(t *testing.T) {
	strategyID := "test_strategy"
	
	orderID1 := generateOrderID("BID", strategyID)
	orderID2 := generateOrderID("ASK", strategyID)
	
	if orderID1 == orderID2 {
		t.Error("Order IDs should be unique")
	}
	
	if len(orderID1) == 0 {
		t.Error("Order ID should not be empty")
	}
	
	// Check format: prefix_strategyID_timestamp
	if orderID1[:4] != "BID_" {
		t.Errorf("Order ID should start with 'BID_', got '%s'", orderID1[:4])
	}
	
	if orderID2[:4] != "ASK_" {
		t.Errorf("Order ID should start with 'ASK_', got '%s'", orderID2[:4])
	}
}

// Benchmark tests for performance
func BenchmarkMarketMakerExecuteMarketMaking(b *testing.B) {
	mm, _ := NewMarketMaker("mm_id", "Market Maker", "Test MM", 
		big.NewFloat(0.02), big.NewFloat(100), big.NewFloat(200), 
		big.NewFloat(1), big.NewFloat(10), big.NewFloat(0.1))
	
	marketData := &MarketData{
		AssetID:   "asset_id",
		BidPrice:  big.NewFloat(99),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	currentInventory := big.NewFloat(150)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mm.ExecuteMarketMaking(marketData, currentInventory)
	}
}

func BenchmarkArbitrageStrategyDetectArbitrage(b *testing.B) {
	as, _ := NewArbitrageStrategy("arb_id", "Arbitrage", "Test Arb", 
		big.NewFloat(0.01), big.NewFloat(0.005), 100*time.Millisecond, 5)
	
	marketData1 := &MarketData{
		AssetID:   "asset1",
		BidPrice:  big.NewFloat(100),
		AskPrice:  big.NewFloat(101),
		LastPrice: big.NewFloat(100.5),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	marketData2 := &MarketData{
		AssetID:   "asset2",
		BidPrice:  big.NewFloat(102),
		AskPrice:  big.NewFloat(103),
		LastPrice: big.NewFloat(102.5),
		Volume:    big.NewFloat(1000),
		Timestamp: time.Now(),
		BidSize:   big.NewFloat(50),
		AskSize:   big.NewFloat(50),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = as.DetectArbitrage(marketData1, marketData2)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
