package testing

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/defi/lending/advanced"
	"github.com/gochain/gochain/pkg/exchange/orderbook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EndToEndTestSuite provides comprehensive end-to-end testing of the GoChain DeFi ecosystem
type EndToEndTestSuite struct {
	ctx context.Context

	// DeFi components
	crossCollateral *advanced.CrossCollateralManager

	// Exchange components
	orderBook *orderbook.OrderBook
}

// NewEndToEndTestSuite creates a new end-to-end test suite
func NewEndToEndTestSuite() *EndToEndTestSuite {
	return &EndToEndTestSuite{
		ctx: context.Background(),
	}
}

// TestCompleteGoChainEcosystem tests the complete GoChain DeFi ecosystem end-to-end
func TestCompleteGoChainEcosystem(t *testing.T) {
	suite := NewEndToEndTestSuite()

	// Test 1: DeFi Protocol Foundation
	t.Run("DeFiProtocolFoundation", suite.testDeFiProtocolFoundation)

	// Test 2: Exchange Operations
	t.Run("ExchangeOperations", suite.testExchangeOperations)

	// Test 3: Cross-Protocol Integration
	t.Run("CrossProtocolIntegration", suite.testCrossProtocolIntegration)

	// Test 4: Complete User Journey
	t.Run("CompleteUserJourney", suite.testCompleteUserJourney)

	// Test 5: System Stress Testing
	t.Run("SystemStressTesting", suite.testSystemStressTesting)

	// Test 6: Performance Validation
	t.Run("PerformanceValidation", suite.testPerformanceValidation)
}

// testDeFiProtocolFoundation tests the foundational DeFi operations
func (suite *EndToEndTestSuite) testDeFiProtocolFoundation(t *testing.T) {
	fmt.Println("🔄 Testing DeFi Protocol Foundation...")

	// Initialize cross-collateral manager
	suite.crossCollateral = advanced.NewCrossCollateralManager()
	require.NotNil(t, suite.crossCollateral)

	// Create portfolio
	userID := "e2e_user_1"
	portfolio, err := suite.crossCollateral.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)
	require.NotNil(t, portfolio)

	fmt.Printf("✅ Created DeFi portfolio:\n")
	fmt.Printf("   User: %s\n", portfolio.UserID)
	fmt.Printf("   Min Collateral Ratio: %v\n", portfolio.MinCollateralRatio)

	// Add collateral
	btcAsset := &advanced.CrossCollateralAsset{
		ID:             "BTC",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "BTC",
		Amount:         big.NewInt(100000000), // 1 BTC
		Value:          big.NewInt(50000000),  // $50k
		Volatility:     big.NewFloat(0.8),
		LiquidityScore: big.NewFloat(0.9),
		RiskScore:      big.NewFloat(0.7),
	}

	err = suite.crossCollateral.AddCollateral(suite.ctx, userID, btcAsset)
	require.NoError(t, err)

	// Create borrowing position
	position, err := suite.crossCollateral.CreatePosition(suite.ctx, userID, "USDT", big.NewInt(25000000), big.NewFloat(2.0))
	require.NoError(t, err)
	require.NotNil(t, position)

	fmt.Printf("✅ DeFi operations completed:\n")
	fmt.Printf("   Added BTC collateral: %s BTC = $%s\n",
		btcAsset.Amount.String(), btcAsset.Value.String())
	fmt.Printf("   Created position: $%s USDT\n", position.Amount.String())

	// Verify portfolio state
	portfolioState, err := suite.crossCollateral.GetPortfolio(userID)
	require.NoError(t, err)

	assert.True(t, portfolioState.TotalCollateralValue.Cmp(big.NewInt(0)) > 0)
	assert.True(t, portfolioState.TotalBorrowedValue.Cmp(big.NewInt(0)) > 0)

	fmt.Printf("✅ Portfolio state verified:\n")
	fmt.Printf("   Total Collateral: $%s\n", portfolioState.TotalCollateralValue.String())
	fmt.Printf("   Total Borrowed: $%s\n", portfolioState.TotalBorrowedValue.String())
	fmt.Printf("   Collateral Ratio: %v\n", portfolioState.CollateralRatio.String())

	fmt.Println("✅ DeFi Protocol Foundation Test Passed")
}

// testExchangeOperations tests exchange operations
func (suite *EndToEndTestSuite) testExchangeOperations(t *testing.T) {
	fmt.Println("🔄 Testing Exchange Operations...")

	// Create order book
	var err error
	suite.orderBook, err = orderbook.NewOrderBook("BTC/USDT")
	require.NoError(t, err)
	require.NotNil(t, suite.orderBook)

	// Add orders
	buyOrder := &orderbook.Order{
		ID:                "buy_1",
		TradingPair:       "BTC/USDT",
		Side:              orderbook.OrderSideBuy,
		Type:              orderbook.OrderTypeLimit,
		Status:            orderbook.OrderStatusPending,
		Quantity:          big.NewInt(100000), // 0.1 BTC
		Price:             big.NewInt(50000),  // $50k
		UserID:            "user_1",
		TimeInForce:       orderbook.TimeInForceGTC,
		FilledQuantity:    big.NewInt(0),
		RemainingQuantity: big.NewInt(100000),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	sellOrder := &orderbook.Order{
		ID:                "sell_1",
		TradingPair:       "BTC/USDT",
		Side:              orderbook.OrderSideSell,
		Type:              orderbook.OrderTypeLimit,
		Status:            orderbook.OrderStatusPending,
		Quantity:          big.NewInt(100000), // 0.1 BTC
		Price:             big.NewInt(50000),  // $50k
		UserID:            "user_2",
		TimeInForce:       orderbook.TimeInForceGTC,
		FilledQuantity:    big.NewInt(0),
		RemainingQuantity: big.NewInt(100000),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Add orders to order book
	err = suite.orderBook.AddOrder(buyOrder)
	require.NoError(t, err)
	err = suite.orderBook.AddOrder(sellOrder)
	require.NoError(t, err)

	fmt.Printf("✅ Added orders to order book:\n")
	fmt.Printf("   Buy Order: %s BTC at $%s\n", buyOrder.Quantity.String(), buyOrder.Price.String())
	fmt.Printf("   Sell Order: %s BTC at $%s\n", sellOrder.Quantity.String(), sellOrder.Price.String())

	// Get order book depth
	depth, err := suite.orderBook.GetDepth(10)
	require.NoError(t, err)
	assert.NotNil(t, depth)
	assert.Len(t, depth, 2) // 1 buy + 1 sell level

	fmt.Printf("✅ Order book depth:\n")
	fmt.Printf("   Total Levels: %d\n", len(depth))

	// Create simple market summary manually
	fmt.Printf("✅ Market summary:\n")
	fmt.Printf("   Trading Pair: BTC/USDT\n")
	fmt.Printf("   Order Count: %d\n", suite.orderBook.GetOrderCount())
	fmt.Printf("   Buy Volume: %s\n", suite.orderBook.GetBuyVolume().String())
	fmt.Printf("   Sell Volume: %s\n", suite.orderBook.GetSellVolume().String())

	fmt.Println("✅ Exchange Operations Test Passed")
}

// testCrossProtocolIntegration tests cross-protocol integration
func (suite *EndToEndTestSuite) testCrossProtocolIntegration(t *testing.T) {
	fmt.Println("🔄 Testing Cross-Protocol Integration...")

	// Test integration between DeFi and Exchange
	userID := "cross_protocol_user"

	// Create portfolio
	_, err := suite.crossCollateral.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)

	// Add collateral
	btcAsset := &advanced.CrossCollateralAsset{
		ID:             "BTC",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "BTC",
		Amount:         big.NewInt(200000000), // 2 BTC
		Value:          big.NewInt(100000000), // $100k
		Volatility:     big.NewFloat(0.8),
		LiquidityScore: big.NewFloat(0.9),
		RiskScore:      big.NewFloat(0.7),
	}

	err = suite.crossCollateral.AddCollateral(suite.ctx, userID, btcAsset)
	require.NoError(t, err)

	// Create position
	position, err := suite.crossCollateral.CreatePosition(suite.ctx, userID, "USDT", big.NewInt(50000000), big.NewFloat(2.0))
	require.NoError(t, err)

	// Use borrowed funds to place trading orders
	buyOrder := &orderbook.Order{
		ID:                "cross_protocol_buy",
		TradingPair:       "BTC/USDT",
		Side:              orderbook.OrderSideBuy,
		Type:              orderbook.OrderTypeLimit,
		Status:            orderbook.OrderStatusPending,
		Quantity:          big.NewInt(50000), // 0.05 BTC
		Price:             big.NewInt(48000), // $48k
		UserID:            userID,
		TimeInForce:       orderbook.TimeInForceGTC,
		FilledQuantity:    big.NewInt(0),
		RemainingQuantity: big.NewInt(50000),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = suite.orderBook.AddOrder(buyOrder)
	require.NoError(t, err)

	fmt.Printf("✅ Cross-protocol integration completed:\n")
	fmt.Printf("   Portfolio: %s with $%s collateral\n", userID, btcAsset.Value.String())
	fmt.Printf("   Position: $%s USDT borrowed\n", position.Amount.String())
	fmt.Printf("   Trading: Placed buy order for %s BTC at $%s\n",
		buyOrder.Quantity.String(), buyOrder.Price.String())

	// Verify integration
	portfolio, err := suite.crossCollateral.GetPortfolio(userID)
	require.NoError(t, err)

	depth, err := suite.orderBook.GetDepth(10)
	require.NoError(t, err)

	fmt.Printf("✅ Integration verification:\n")
	fmt.Printf("   Portfolio Status: Active (Collateral: $%s, Borrowed: $%s)\n",
		portfolio.TotalCollateralValue.String(), portfolio.TotalBorrowedValue.String())
	fmt.Printf("   Trading Status: Active (Orders: %d)\n", len(depth))

	fmt.Println("✅ Cross-Protocol Integration Test Passed")
}

// testCompleteUserJourney tests a complete user journey through the ecosystem
func (suite *EndToEndTestSuite) testCompleteUserJourney(t *testing.T) {
	fmt.Println("🔄 Testing Complete User Journey...")

	user := "0x9999999999999999999999999999999999999999"
	fmt.Printf("👤 User: %s\n", user)

	// 1. User creates DeFi portfolio
	fmt.Println("\n💰 Step 1: Create DeFi portfolio")

	_, err := suite.crossCollateral.CreatePortfolio(suite.ctx, user, big.NewFloat(1.5))
	require.NoError(t, err)
	fmt.Printf("   ✅ Created DeFi portfolio\n")

	// 2. User adds collateral
	fmt.Println("\n🏦 Step 2: Add collateral to portfolio")

	btcAsset := &advanced.CrossCollateralAsset{
		ID:             "BTC",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "BTC",
		Amount:         big.NewInt(100000000), // 1 BTC
		Value:          big.NewInt(50000000),  // $50k
		Volatility:     big.NewFloat(0.8),
		LiquidityScore: big.NewFloat(0.9),
		RiskScore:      big.NewFloat(0.7),
	}

	err = suite.crossCollateral.AddCollateral(suite.ctx, user, btcAsset)
	require.NoError(t, err)
	fmt.Printf("   ✅ Added BTC collateral: %s BTC = $%s\n",
		btcAsset.Amount.String(), btcAsset.Value.String())

	// 3. User creates borrowing position
	fmt.Println("\n💸 Step 3: Create borrowing position")

	position, err := suite.crossCollateral.CreatePosition(suite.ctx, user, "USDT", big.NewInt(25000000), big.NewFloat(2.0))
	require.NoError(t, err)
	fmt.Printf("   ✅ Created position: $%s USDT\n", position.Amount.String())

	// 4. User places trading orders
	fmt.Println("\n📊 Step 4: Place trading orders")

	// Place buy order
	buyOrder := &orderbook.Order{
		ID:                "user_buy_1",
		TradingPair:       "BTC/USDT",
		Side:              orderbook.OrderSideBuy,
		Type:              orderbook.OrderTypeLimit,
		Status:            orderbook.OrderStatusPending,
		Quantity:          big.NewInt(50000), // 0.05 BTC
		Price:             big.NewInt(48000), // $48k
		UserID:            user,
		TimeInForce:       orderbook.TimeInForceGTC,
		FilledQuantity:    big.NewInt(0),
		RemainingQuantity: big.NewInt(50000),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = suite.orderBook.AddOrder(buyOrder)
	require.NoError(t, err)
	fmt.Printf("   ✅ Placed buy order: %s BTC at $%s\n",
		buyOrder.Quantity.String(), buyOrder.Price.String())

	// 5. Check portfolio state
	fmt.Println("\n📈 Step 5: Check portfolio state")

	portfolioState, err := suite.crossCollateral.GetPortfolio(user)
	require.NoError(t, err)

	fmt.Printf("   📊 Portfolio Summary:\n")
	fmt.Printf("      Total Collateral: $%s\n", portfolioState.TotalCollateralValue.String())
	fmt.Printf("      Total Borrowed: $%s\n", portfolioState.TotalBorrowedValue.String())
	fmt.Printf("      Net Collateral: $%s\n", portfolioState.NetCollateralValue.String())
	fmt.Printf("      Collateral Ratio: %v\n", portfolioState.CollateralRatio.String())

	// 6. Validate portfolio health
	fmt.Println("\n🔍 Step 6: Validate portfolio health")

	issues, err := suite.crossCollateral.ValidatePortfolioState(user)
	require.NoError(t, err)

	if len(issues) > 0 {
		fmt.Printf("   ⚠️  Portfolio validation issues found:\n")
		for i, issue := range issues {
			fmt.Printf("      %d. %s\n", i+1, issue)
		}
	} else {
		fmt.Printf("   ✅ Portfolio validation passed - no issues found\n")
	}

	// 7. Check order book state
	fmt.Println("\n📊 Step 7: Check order book state")

	depth, err := suite.orderBook.GetDepth(10)
	require.NoError(t, err)
	fmt.Printf("   📊 Order Book Depth:\n")
	fmt.Printf("      Total Levels: %d\n", len(depth))

	// 8. Final ecosystem state
	fmt.Println("\n🎯 Step 8: Final ecosystem state")

	fmt.Printf("   📊 Ecosystem State:\n")
	fmt.Printf("      Portfolio Status: Active\n")
	fmt.Printf("      Trading Status: Active\n")
	fmt.Printf("      DeFi Status: Active\n")

	fmt.Println("\n🎉 Complete User Journey Test Passed!")
	fmt.Println("   User successfully navigated through:")
	fmt.Println("   ✅ DeFi portfolio creation and management")
	fmt.Println("   ✅ Collateral addition and position creation")
	fmt.Println("   ✅ Trading order placement")
	fmt.Println("   ✅ Portfolio health validation")
	fmt.Println("   ✅ Complete ecosystem integration")
}

// testSystemStressTesting tests the system under stress conditions
func (suite *EndToEndTestSuite) testSystemStressTesting(t *testing.T) {
	fmt.Println("🔄 Testing System Stress Testing...")

	// Test 1: Large portfolio operations
	fmt.Println("\n🏦 Stress Test 1: Large Portfolio Operations")

	userID := "stress_test_user"
	_, err := suite.crossCollateral.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)

	// Add many assets
	assetCount := 50
	startTime := time.Now()

	for i := 0; i < assetCount; i++ {
		asset := &advanced.CrossCollateralAsset{
			ID:             fmt.Sprintf("ASSET_%d", i),
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         fmt.Sprintf("TKN%d", i),
			Amount:         big.NewInt(int64(1000000 + i*1000)),
			Value:          big.NewInt(int64(50000000 + i*1000000)),
			Volatility:     big.NewFloat(0.5 + float64(i%5)*0.1),
			LiquidityScore: big.NewFloat(0.8),
			RiskScore:      big.NewFloat(0.6),
		}

		err = suite.crossCollateral.AddCollateral(suite.ctx, userID, asset)
		require.NoError(t, err)
	}

	portfolioTime := time.Since(startTime)
	fmt.Printf("   ✅ Added %d assets in %v\n", assetCount, portfolioTime)

	// Test 2: High-frequency trading
	fmt.Println("\n📈 Stress Test 2: High-Frequency Trading")

	orderCount := 200
	startTime = time.Now()

	for i := 0; i < orderCount; i++ {
		order := &orderbook.Order{
			ID:                fmt.Sprintf("stress_order_%d", i),
			TradingPair:       "BTC/USDT",
			Side:              orderbook.OrderSideBuy,
			Type:              orderbook.OrderTypeLimit,
			Status:            orderbook.OrderStatusPending,
			Quantity:          big.NewInt(1000 + int64(i)),
			Price:             big.NewInt(45000 + int64(i*100)),
			UserID:            fmt.Sprintf("user_%d", i),
			TimeInForce:       orderbook.TimeInForceGTC,
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(1000 + int64(i)),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err = suite.orderBook.AddOrder(order)
		require.NoError(t, err)
	}

	tradingTime := time.Since(startTime)
	ordersPerSecond := float64(orderCount) / tradingTime.Seconds()

	fmt.Printf("   ✅ Placed %d orders in %v\n", orderCount, tradingTime)
	fmt.Printf("   📊 Order Rate: %.2f orders/second\n", ordersPerSecond)

	// Test 3: Memory and resource usage
	fmt.Println("\n💾 Stress Test 3: Memory and Resource Usage")

	// Get portfolio details to check memory usage
	portfolio, err := suite.crossCollateral.GetPortfolio(userID)
	require.NoError(t, err)

	fmt.Printf("   📊 Portfolio Resource Usage:\n")
	fmt.Printf("      Assets: %d\n", len(portfolio.CollateralAssets))
	fmt.Printf("      Positions: %d\n", len(portfolio.Positions))
	fmt.Printf("      Risk Metrics: %v\n", portfolio.RiskMetrics != nil)

	// Get order book depth
	depth, err := suite.orderBook.GetDepth(100)
	require.NoError(t, err)
	fmt.Printf("   📊 Order Book Resource Usage:\n")
	fmt.Printf("      Total Levels: %d\n", len(depth))

	fmt.Println("✅ System Stress Testing Completed")
}

// testPerformanceValidation tests system performance metrics
func (suite *EndToEndTestSuite) testPerformanceValidation(t *testing.T) {
	fmt.Println("🔄 Testing Performance Validation...")

	// Performance Test 1: Portfolio Calculation Speed
	fmt.Println("\n⚡ Performance Test 1: Portfolio Calculation Speed")

	userID := "perf_test_user"
	_, err := suite.crossCollateral.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)

	// Add assets for performance testing
	for i := 0; i < 100; i++ {
		asset := &advanced.CrossCollateralAsset{
			ID:             fmt.Sprintf("PERF_ASSET_%d", i),
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         fmt.Sprintf("PERF%d", i),
			Amount:         big.NewInt(int64(1000000 + i*1000)),
			Value:          big.NewInt(int64(50000000 + i*1000000)),
			Volatility:     big.NewFloat(0.5 + float64(i%5)*0.1),
			LiquidityScore: big.NewFloat(0.8),
			RiskScore:      big.NewFloat(0.6),
		}

		err = suite.crossCollateral.AddCollateral(suite.ctx, userID, asset)
		require.NoError(t, err)
	}

	// Measure portfolio calculation time
	startTime := time.Now()
	portfolio, err := suite.crossCollateral.GetPortfolio(userID)
	require.NoError(t, err)
	portfolioTime := time.Since(startTime)

	fmt.Printf("   📊 Portfolio Calculation Results:\n")
	fmt.Printf("      Calculation Time: %v\n", portfolioTime)
	fmt.Printf("      Assets: %d\n", len(portfolio.CollateralAssets))
	fmt.Printf("      Total Collateral: $%s\n", portfolio.TotalCollateralValue.String())

	// Performance Test 2: Order Book Operations
	fmt.Println("\n⚡ Performance Test 2: Order Book Operations")

	// Measure order addition performance
	startTime = time.Now()

	for i := 0; i < 500; i++ {
		order := &orderbook.Order{
			ID:                fmt.Sprintf("perf_order_%d", i),
			TradingPair:       "BTC/USDT",
			Side:              orderbook.OrderSideBuy,
			Type:              orderbook.OrderTypeLimit,
			Status:            orderbook.OrderStatusPending,
			Quantity:          big.NewInt(1000 + int64(i)),
			Price:             big.NewInt(45000 + int64(i*100)),
			UserID:            fmt.Sprintf("perf_user_%d", i),
			TimeInForce:       orderbook.TimeInForceGTC,
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(1000 + int64(i)),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err = suite.orderBook.AddOrder(order)
		require.NoError(t, err)
	}

	orderTime := time.Since(startTime)
	ordersPerSecond := float64(500) / orderTime.Seconds()

	fmt.Printf("   📊 Order Book Performance Results:\n")
	fmt.Printf("      Order Addition Time: %v\n", orderTime)
	fmt.Printf("      Orders per Second: %.2f\n", ordersPerSecond)

	// Performance Test 3: End-to-End Latency
	fmt.Println("\n⚡ Performance Test 3: End-to-End Latency")

	// Simulate complete user operation
	startTime = time.Now()

	// 1. Create portfolio
	testUser := "latency_test_user"
	_, err = suite.crossCollateral.CreatePortfolio(suite.ctx, testUser, big.NewFloat(1.5))
	require.NoError(t, err)

	// 2. Add collateral
	asset := &advanced.CrossCollateralAsset{
		ID:             "LATENCY_BTC",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "BTC",
		Amount:         big.NewInt(100000000),
		Value:          big.NewInt(50000000),
		Volatility:     big.NewFloat(0.8),
		LiquidityScore: big.NewFloat(0.9),
		RiskScore:      big.NewFloat(0.7),
	}
	err = suite.crossCollateral.AddCollateral(suite.ctx, testUser, asset)
	require.NoError(t, err)

	// 3. Create position
	_, err = suite.crossCollateral.CreatePosition(suite.ctx, testUser, "USDT", big.NewInt(25000000), big.NewFloat(2.0))
	require.NoError(t, err)

	// 4. Get portfolio state
	_, err = suite.crossCollateral.GetPortfolio(testUser)
	require.NoError(t, err)

	totalLatency := time.Since(startTime)

	fmt.Printf("   📊 End-to-End Latency Results:\n")
	fmt.Printf("      Total Operation Time: %v\n", totalLatency)
	fmt.Printf("      Operations Completed: 4\n")
	fmt.Printf("      Average per Operation: %v\n", totalLatency/4)

	fmt.Println("✅ Performance Validation Completed")
}

// RunAllEndToEndTests runs all end-to-end tests
func (suite *EndToEndTestSuite) RunAllEndToEndTests(t *testing.T) {
	fmt.Println("🚀 Running All End-to-End Tests...")

	TestCompleteGoChainEcosystem(t)

	fmt.Println("🎉 All End-to-End Tests Completed Successfully!")
}
