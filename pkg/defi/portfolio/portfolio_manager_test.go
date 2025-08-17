package portfolio

import (
	"math/big"
	"testing"
)

func TestNewPortfolioManager(t *testing.T) {
	pm := NewPortfolioManager()
	
	if pm == nil {
		t.Fatal("NewPortfolioManager returned nil")
	}
	
	if pm.Portfolios == nil {
		t.Error("Portfolios map not initialized")
	}
	if pm.Assets == nil {
		t.Error("Assets map not initialized")
	}
	
	// Check timestamps are set
	if pm.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if pm.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewPortfolio(t *testing.T) {
	portfolio, err := NewPortfolio("test_id", "Test Portfolio", "Test Description", "owner123", Moderate, BuyAndHold)
	if err != nil {
		t.Fatalf("NewPortfolio failed: %v", err)
	}
	
	if portfolio.ID != "test_id" {
		t.Errorf("Expected ID 'test_id', got '%s'", portfolio.ID)
	}
	if portfolio.Name != "Test Portfolio" {
		t.Errorf("Expected name 'Test Portfolio', got '%s'", portfolio.Name)
	}
	if portfolio.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", portfolio.Description)
	}
	if portfolio.Owner != "owner123" {
		t.Errorf("Expected owner 'owner123', got '%s'", portfolio.Owner)
	}
	if portfolio.RiskProfile != Moderate {
		t.Errorf("Expected risk profile Moderate, got %d", portfolio.RiskProfile)
	}
	if portfolio.Strategy != BuyAndHold {
		t.Errorf("Expected strategy BuyAndHold, got %d", portfolio.Strategy)
	}
	if portfolio.TotalValue.Sign() != 0 {
		t.Error("Expected initial total value to be 0")
	}
	if portfolio.Positions == nil {
		t.Error("Positions map not initialized")
	}
	if portfolio.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if portfolio.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewPortfolioValidation(t *testing.T) {
	// Test empty ID
	_, err := NewPortfolio("", "Test", "Description", "owner", Moderate, BuyAndHold)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty name
	_, err = NewPortfolio("test_id", "", "Description", "owner", Moderate, BuyAndHold)
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test empty owner
	_, err = NewPortfolio("test_id", "Test", "Description", "", Moderate, BuyAndHold)
	if err == nil {
		t.Error("Expected error for empty owner")
	}
}

func TestNewAsset(t *testing.T) {
	price := big.NewFloat(50000)
	marketCap := big.NewFloat(1000000000)
	volume24h := big.NewFloat(50000000)
	riskScore := big.NewFloat(0.7)
	
	asset, err := NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, price, marketCap, volume24h, riskScore)
	if err != nil {
		t.Fatalf("NewAsset failed: %v", err)
	}
	
	if asset.ID != "btc" {
		t.Errorf("Expected ID 'btc', got '%s'", asset.ID)
	}
	if asset.Symbol != "BTC" {
		t.Errorf("Expected symbol 'BTC', got '%s'", asset.Symbol)
	}
	if asset.Name != "Bitcoin" {
		t.Errorf("Expected name 'Bitcoin', got '%s'", asset.Name)
	}
	if asset.Type != Cryptocurrency {
		t.Errorf("Expected type Cryptocurrency, got %d", asset.Type)
	}
	if asset.Price.Cmp(price) != 0 {
		t.Error("Price not set correctly")
	}
	if asset.MarketCap.Cmp(marketCap) != 0 {
		t.Error("MarketCap not set correctly")
	}
	if asset.Volume24h.Cmp(volume24h) != 0 {
		t.Error("Volume24h not set correctly")
	}
	if asset.RiskScore.Cmp(riskScore) != 0 {
		t.Error("RiskScore not set correctly")
	}
	if asset.LastUpdated.IsZero() {
		t.Error("LastUpdated not set")
	}
}

func TestNewAssetValidation(t *testing.T) {
	validPrice := big.NewFloat(50000)
	validMarketCap := big.NewFloat(1000000000)
	validVolume24h := big.NewFloat(50000000)
	validRiskScore := big.NewFloat(0.7)
	
	// Test empty ID
	_, err := NewAsset("", "BTC", "Bitcoin", Cryptocurrency, validPrice, validMarketCap, validVolume24h, validRiskScore)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty symbol
	_, err = NewAsset("btc", "", "Bitcoin", Cryptocurrency, validPrice, validMarketCap, validVolume24h, validRiskScore)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// Test empty name
	_, err = NewAsset("btc", "BTC", "", Cryptocurrency, validPrice, validMarketCap, validVolume24h, validRiskScore)
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test nil price
	_, err = NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, nil, validMarketCap, validVolume24h, validRiskScore)
	if err == nil {
		t.Error("Expected error for nil price")
	}
	
	// Test negative price
	_, err = NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, big.NewFloat(-50000), validMarketCap, validVolume24h, validRiskScore)
	if err == nil {
		t.Error("Expected error for negative price")
	}
}

func TestNewPosition(t *testing.T) {
	quantity := big.NewFloat(1.5)
	entryPrice := big.NewFloat(50000)
	
	position, err := NewPosition("btc", quantity, entryPrice)
	if err != nil {
		t.Fatalf("NewPosition failed: %v", err)
	}
	
	if position.AssetID != "btc" {
		t.Errorf("Expected AssetID 'btc', got '%s'", position.AssetID)
	}
	if position.Quantity.Cmp(quantity) != 0 {
		t.Error("Quantity not set correctly")
	}
	if position.EntryPrice.Cmp(entryPrice) != 0 {
		t.Error("EntryPrice not set correctly")
	}
	if position.CurrentPrice.Cmp(entryPrice) != 0 {
		t.Error("CurrentPrice not set correctly")
	}
	
	// Check value calculation
	expectedValue := new(big.Float).Mul(quantity, entryPrice)
	if position.Value.Cmp(expectedValue) != 0 {
		t.Error("Value not calculated correctly")
	}
	
	if position.Pnl.Sign() != 0 {
		t.Error("Expected initial PnL to be 0")
	}
	if position.PnlPercent.Sign() != 0 {
		t.Error("Expected initial PnL percentage to be 0")
	}
	if position.Weight.Sign() != 0 {
		t.Error("Expected initial weight to be 0")
	}
	if position.LastUpdated.IsZero() {
		t.Error("LastUpdated not set")
	}
}

func TestNewPositionValidation(t *testing.T) {
	validQuantity := big.NewFloat(1.5)
	validPrice := big.NewFloat(50000)
	
	// Test empty asset ID
	_, err := NewPosition("", validQuantity, validPrice)
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
	
	// Test nil quantity
	_, err = NewPosition("btc", nil, validPrice)
	if err == nil {
		t.Error("Expected error for nil quantity")
	}
	
	// Test negative quantity
	_, err = NewPosition("btc", big.NewFloat(-1.5), validPrice)
	if err == nil {
		t.Error("Expected error for negative quantity")
	}
	
	// Test zero quantity
	_, err = NewPosition("btc", big.NewFloat(0), validPrice)
	if err == nil {
		t.Error("Expected error for zero quantity")
	}
	
	// Test nil price
	_, err = NewPosition("btc", validQuantity, nil)
	if err == nil {
		t.Error("Expected error for nil price")
	}
	
	// Test negative price
	_, err = NewPosition("btc", validQuantity, big.NewFloat(-50000))
	if err == nil {
		t.Error("Expected error for negative price")
	}
	
	// Test zero price
	_, err = NewPosition("btc", validQuantity, big.NewFloat(0))
	if err == nil {
		t.Error("Expected error for zero price")
	}
}

func TestPortfolioManagerAddPortfolio(t *testing.T) {
	pm := NewPortfolioManager()
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	
	err := pm.AddPortfolio(portfolio)
	if err != nil {
		t.Fatalf("AddPortfolio failed: %v", err)
	}
	
	if _, exists := pm.Portfolios["test_id"]; !exists {
		t.Error("Portfolio not added to manager")
	}
	
	// Test adding duplicate portfolio
	err = pm.AddPortfolio(portfolio)
	if err == nil {
		t.Error("Expected error for duplicate portfolio")
	}
	
	// Test adding nil portfolio
	err = pm.AddPortfolio(nil)
	if err == nil {
		t.Error("Expected error for nil portfolio")
	}
}

func TestPortfolioManagerRemovePortfolio(t *testing.T) {
	pm := NewPortfolioManager()
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	pm.AddPortfolio(portfolio)
	
	err := pm.RemovePortfolio("test_id")
	if err != nil {
		t.Fatalf("RemovePortfolio failed: %v", err)
	}
	
	if _, exists := pm.Portfolios["test_id"]; exists {
		t.Error("Portfolio not removed from manager")
	}
	
	// Test removing non-existent portfolio
	err = pm.RemovePortfolio("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent portfolio")
	}
	
	// Test removing with empty ID
	err = pm.RemovePortfolio("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestPortfolioManagerAddAsset(t *testing.T) {
	pm := NewPortfolioManager()
	asset, _ := NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, big.NewFloat(50000), nil, nil, nil)
	
	err := pm.AddAsset(asset)
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}
	
	if _, exists := pm.Assets["btc"]; !exists {
		t.Error("Asset not added to manager")
	}
	
	// Test adding duplicate asset
	err = pm.AddAsset(asset)
	if err == nil {
		t.Error("Expected error for duplicate asset")
	}
	
	// Test adding nil asset
	err = pm.AddAsset(nil)
	if err == nil {
		t.Error("Expected error for nil asset")
	}
}

func TestPortfolioManagerUpdateAssetPrice(t *testing.T) {
	pm := NewPortfolioManager()
	asset, _ := NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, big.NewFloat(50000), nil, nil, nil)
	pm.AddAsset(asset)
	
	// Add portfolio with position in this asset
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	pm.AddPortfolio(portfolio)
	
	newPrice := big.NewFloat(55000)
	err := pm.UpdateAssetPrice("btc", newPrice)
	if err != nil {
		t.Fatalf("UpdateAssetPrice failed: %v", err)
	}
	
	// Check asset price updated
	if pm.Assets["btc"].Price.Cmp(newPrice) != 0 {
		t.Error("Asset price not updated")
	}
	
	// Check portfolio position updated
	if portfolio.Positions["btc"].CurrentPrice.Cmp(newPrice) != 0 {
		t.Error("Portfolio position price not updated")
	}
	
	// Test updating non-existent asset
	err = pm.UpdateAssetPrice("non_existent", newPrice)
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
	
	// Test updating with nil price
	err = pm.UpdateAssetPrice("btc", nil)
	if err == nil {
		t.Error("Expected error for nil price")
	}
	
	// Test updating with negative price
	err = pm.UpdateAssetPrice("btc", big.NewFloat(-55000))
	if err == nil {
		t.Error("Expected error for negative price")
	}
}

func TestPortfolioAddPosition(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	
	err := portfolio.AddPosition(position)
	if err != nil {
		t.Fatalf("AddPosition failed: %v", err)
	}
	
	if _, exists := portfolio.Positions["btc"]; !exists {
		t.Error("Position not added to portfolio")
	}
	
	// Check portfolio value updated
	expectedValue := big.NewFloat(50000)
	if portfolio.TotalValue.Cmp(expectedValue) != 0 {
		t.Errorf("Expected portfolio value %v, got %v", expectedValue, portfolio.TotalValue)
	}
	
	// Test adding duplicate position
	err = portfolio.AddPosition(position)
	if err == nil {
		t.Error("Expected error for duplicate position")
	}
	
	// Test adding nil position
	err = portfolio.AddPosition(nil)
	if err == nil {
		t.Error("Expected error for nil position")
	}
}

func TestPortfolioUpdatePosition(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	
	// Update position
	newQuantity := big.NewFloat(2)
	newPrice := big.NewFloat(55000)
	err := portfolio.UpdatePosition("btc", newQuantity, newPrice)
	if err != nil {
		t.Fatalf("UpdatePosition failed: %v", err)
	}
	
	// Check position updated
	if portfolio.Positions["btc"].Quantity.Cmp(newQuantity) != 0 {
		t.Error("Position quantity not updated")
	}
	if portfolio.Positions["btc"].EntryPrice.Cmp(newPrice) != 0 {
		t.Error("Position entry price not updated")
	}
	if portfolio.Positions["btc"].CurrentPrice.Cmp(newPrice) != 0 {
		t.Error("Position current price not updated")
	}
	
	// Check portfolio value updated
	expectedValue := new(big.Float).Mul(newQuantity, newPrice)
	if portfolio.TotalValue.Cmp(expectedValue) != 0 {
		t.Errorf("Expected portfolio value %v, got %v", expectedValue, portfolio.TotalValue)
	}
	
	// Test updating non-existent position
	err = portfolio.UpdatePosition("non_existent", newQuantity, newPrice)
	if err == nil {
		t.Error("Expected error for non-existent position")
	}
	
	// Test updating with invalid parameters
	err = portfolio.UpdatePosition("btc", big.NewFloat(-2), newPrice)
	if err == nil {
		t.Error("Expected error for negative quantity")
	}
	
	err = portfolio.UpdatePosition("btc", newQuantity, big.NewFloat(-55000))
	if err == nil {
		t.Error("Expected error for negative price")
	}
}

func TestPortfolioRemovePosition(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	
	err := portfolio.RemovePosition("btc")
	if err != nil {
		t.Fatalf("RemovePosition failed: %v", err)
	}
	
	if _, exists := portfolio.Positions["btc"]; exists {
		t.Error("Position not removed from portfolio")
	}
	
	// Check portfolio value updated
	if portfolio.TotalValue.Sign() != 0 {
		t.Error("Portfolio value not updated after position removal")
	}
	
	// Test removing non-existent position
	err = portfolio.RemovePosition("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}
	
	// Test removing with empty asset ID
	err = portfolio.RemovePosition("")
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
}

func TestPortfolioCalculatePnL(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	
	// Update asset price to simulate price change
	portfolio.updatePositionPrice("btc", big.NewFloat(55000))
	
	// Check PnL calculation
	expectedPnL := big.NewFloat(5000) // (55000 - 50000) * 1
	if position.Pnl.Cmp(expectedPnL) != 0 {
		t.Errorf("Expected PnL %v, got %v", expectedPnL, position.Pnl)
	}
	
	// Check PnL percentage
	expectedPnLPercent := big.NewFloat(10) // (5000 / 50000) * 100
	if position.PnlPercent.Cmp(expectedPnLPercent) != 0 {
		t.Errorf("Expected PnL percentage %v, got %v", expectedPnLPercent, position.PnlPercent)
	}
}

func TestPortfolioUpdatePositionWeights(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	
	// Add multiple positions
	btcPosition, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	ethPosition, _ := NewPosition("eth", big.NewFloat(10), big.NewFloat(3000))
	
	portfolio.AddPosition(btcPosition)
	portfolio.AddPosition(ethPosition)
	
	// Update weights
	portfolio.updatePositionWeights()
	
	// Check weights (btc: 50000, eth: 30000, total: 80000)
	// btc weight should be 50000/80000 * 100 = 62.5%
	// eth weight should be 30000/80000 * 100 = 37.5%
	
	btcWeight := portfolio.Positions["btc"].Weight
	ethWeight := portfolio.Positions["eth"].Weight
	
	if btcWeight.Cmp(big.NewFloat(62.5)) != 0 {
		t.Errorf("Expected BTC weight 62.5, got %v", btcWeight)
	}
	
	if ethWeight.Cmp(big.NewFloat(37.5)) != 0 {
		t.Errorf("Expected ETH weight 37.5, got %v", ethWeight)
	}
}

func TestPortfolioRebalancePortfolio(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	
	// Add positions
	btcPosition, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	ethPosition, _ := NewPosition("eth", big.NewFloat(10), big.NewFloat(3000))
	
	portfolio.AddPosition(btcPosition)
	portfolio.AddPosition(ethPosition)
	
	// Set target weights (60% BTC, 40% ETH)
	targetWeights := map[string]*big.Float{
		"btc": big.NewFloat(60),
		"eth": big.NewFloat(40),
	}
	
	err := portfolio.RebalancePortfolio(targetWeights)
	if err != nil {
		t.Fatalf("RebalancePortfolio failed: %v", err)
	}
	
	// Check rebalancing trades
	trades := portfolio.GetRebalancingTrades()
	if len(trades) == 0 {
		t.Error("Expected rebalancing trades to be generated")
	}
	
	// Clear trades
	portfolio.ClearRebalancingTrades()
	if len(portfolio.GetRebalancingTrades()) != 0 {
		t.Error("Rebalancing trades not cleared")
	}
	
	// Test invalid target weights (not summing to 100%)
	invalidWeights := map[string]*big.Float{
		"btc": big.NewFloat(60),
		"eth": big.NewFloat(50), // Sum = 110%
	}
	
	err = portfolio.RebalancePortfolio(invalidWeights)
	if err == nil {
		t.Error("Expected error for invalid target weights")
	}
	
	// Test nil target weights
	err = portfolio.RebalancePortfolio(nil)
	if err == nil {
		t.Error("Expected error for nil target weights")
	}
}

func TestPortfolioGetters(t *testing.T) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	
	// Test GetPortfolioValue
	value := portfolio.GetPortfolioValue()
	if value.Sign() != 0 {
		t.Error("Expected initial portfolio value to be 0")
	}
	
	// Test GetTotalPnL
	totalPnL := portfolio.GetTotalPnL()
	if totalPnL.Sign() != 0 {
		t.Error("Expected initial total PnL to be 0")
	}
	
	// Test GetTotalPnLPercent
	totalPnLPercent := portfolio.GetTotalPnLPercent()
	if totalPnLPercent.Sign() != 0 {
		t.Error("Expected initial total PnL percentage to be 0")
	}
	
	// Test GetAssetAllocation
	allocation := portfolio.GetAssetAllocation()
	if len(allocation) != 0 {
		t.Error("Expected initial asset allocation to be empty")
	}
	
	// Add a position and test again
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	
	value = portfolio.GetPortfolioValue()
	if value.Cmp(big.NewFloat(50000)) != 0 {
		t.Errorf("Expected portfolio value 50000, got %v", value)
	}
	
	allocation = portfolio.GetAssetAllocation()
	if len(allocation) != 1 {
		t.Error("Expected asset allocation to have 1 asset")
	}
	
	// Test GetPosition
	retrievedPosition, err := portfolio.GetPosition("btc")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if retrievedPosition != position {
		t.Error("GetPosition returned different position")
	}
	
	// Test GetPosition with non-existent asset
	_, err = portfolio.GetPosition("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}
}

func TestPortfolioManagerGetters(t *testing.T) {
	pm := NewPortfolioManager()
	
	// Test GetPortfolio
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	pm.AddPortfolio(portfolio)
	
	retrievedPortfolio, err := pm.GetPortfolio("test_id")
	if err != nil {
		t.Fatalf("GetPortfolio failed: %v", err)
	}
	if retrievedPortfolio != portfolio {
		t.Error("GetPortfolio returned different portfolio")
	}
	
	// Test GetPortfolio with non-existent ID
	_, err = pm.GetPortfolio("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent portfolio")
	}
	
	// Test GetPortfolio with empty ID
	_, err = pm.GetPortfolio("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test GetAsset
	asset, _ := NewAsset("btc", "BTC", "Bitcoin", Cryptocurrency, big.NewFloat(50000), nil, nil, nil)
	pm.AddAsset(asset)
	
	retrievedAsset, err := pm.GetAsset("btc")
	if err != nil {
		t.Fatalf("GetAsset failed: %v", err)
	}
	if retrievedAsset != asset {
		t.Error("GetAsset returned different asset")
	}
	
	// Test GetAsset with non-existent ID
	_, err = pm.GetAsset("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
	
	// Test GetPortfoliosByOwner
	portfolio2, _ := NewPortfolio("test_id2", "Test2", "Description2", "owner", Moderate, BuyAndHold)
	pm.AddPortfolio(portfolio2)
	
	portfolios := pm.GetPortfoliosByOwner("owner")
	if len(portfolios) != 2 {
		t.Errorf("Expected 2 portfolios for owner, got %d", len(portfolios))
	}
	
	// Test with non-existent owner
	portfolios = pm.GetPortfoliosByOwner("non_existent")
	if len(portfolios) != 0 {
		t.Error("Expected 0 portfolios for non-existent owner")
	}
	
	// Test with empty owner
	portfolios = pm.GetPortfoliosByOwner("")
	if portfolios != nil {
		t.Error("Expected nil for empty owner")
	}
}

// Benchmark tests for performance
func BenchmarkPortfolioAddPosition(b *testing.B) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		portfolio.AddPosition(position)
		portfolio.RemovePosition("btc")
	}
}

func BenchmarkPortfolioUpdatePosition(b *testing.B) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	position, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	portfolio.AddPosition(position)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		portfolio.UpdatePosition("btc", big.NewFloat(1), big.NewFloat(50000+float64(i)))
	}
}

func BenchmarkPortfolioRebalance(b *testing.B) {
	portfolio, _ := NewPortfolio("test_id", "Test", "Description", "owner", Moderate, BuyAndHold)
	
	// Add multiple positions
	btcPosition, _ := NewPosition("btc", big.NewFloat(1), big.NewFloat(50000))
	ethPosition, _ := NewPosition("eth", big.NewFloat(10), big.NewFloat(3000))
	portfolio.AddPosition(btcPosition)
	portfolio.AddPosition(ethPosition)
	
	targetWeights := map[string]*big.Float{
		"btc": big.NewFloat(60),
		"eth": big.NewFloat(40),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		portfolio.RebalancePortfolio(targetWeights)
		portfolio.ClearRebalancingTrades()
	}
}
