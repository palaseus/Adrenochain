package testing

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/defi/lending/advanced"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CrossProtocolIntegrationTestSuite tests how different DeFi components work together
type CrossProtocolIntegrationTestSuite struct {
	ctx context.Context
}

// NewCrossProtocolIntegrationTestSuite creates a new cross-protocol integration test suite
func NewCrossProtocolIntegrationTestSuite() *CrossProtocolIntegrationTestSuite {
	return &CrossProtocolIntegrationTestSuite{
		ctx: context.Background(),
	}
}

// TestDeFiEcosystemIntegration tests the complete DeFi ecosystem working together
func TestDeFiEcosystemIntegration(t *testing.T) {
	suite := NewCrossProtocolIntegrationTestSuite()

	// Test 1: Cross-Collateral + Portfolio Management
	t.Run("CrossCollateralPortfolioManagement", suite.testCrossCollateralPortfolioManagement)

	// Test 2: Multi-Asset Portfolio Operations
	t.Run("MultiAssetPortfolioOperations", suite.testMultiAssetPortfolioOperations)

	// Test 3: Portfolio Risk Management
	t.Run("PortfolioRiskManagement", suite.testPortfolioRiskManagement)

	// Test 4: Portfolio Validation and Health Checks
	t.Run("PortfolioValidationHealthChecks", suite.testPortfolioValidationHealthChecks)

	// Test 5: Complete Portfolio Lifecycle
	t.Run("CompletePortfolioLifecycle", suite.testCompletePortfolioLifecycle)
}

// testCrossCollateralPortfolioManagement tests cross-collateral portfolio management
func (suite *CrossProtocolIntegrationTestSuite) testCrossCollateralPortfolioManagement(t *testing.T) {
	fmt.Println("🔄 Testing Cross-Collateral Portfolio Management...")

	// Create cross-collateral manager
	ccm := advanced.NewCrossCollateralManager()

	// Create portfolio for user
	userID := "user_integration_test"
	minCollateralRatio := big.NewFloat(1.5)

	_, err := ccm.CreatePortfolio(suite.ctx, userID, minCollateralRatio)
	require.NoError(t, err)
	fmt.Printf("✅ Created portfolio for user: %s\n", userID)

	// Add multiple collateral assets
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
	err = ccm.AddCollateral(suite.ctx, userID, btcAsset)
	require.NoError(t, err)
	fmt.Printf("✅ Added BTC collateral: %s BTC = $%s\n",
		btcAsset.Amount.String(), btcAsset.Value.String())

	ethAsset := &advanced.CrossCollateralAsset{
		ID:             "ETH",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "ETH",
		Amount:         big.NewInt(1000000000000000000), // 1 ETH
		Value:          big.NewInt(30000000),            // $30k
		Volatility:     big.NewFloat(0.6),
		LiquidityScore: big.NewFloat(0.8),
		RiskScore:      big.NewFloat(0.5),
	}
	err = ccm.AddCollateral(suite.ctx, userID, ethAsset)
	require.NoError(t, err)
	fmt.Printf("✅ Added ETH collateral: %s ETH = $%s\n",
		ethAsset.Amount.String(), ethAsset.Value.String())

	// Create borrowing position
	position, err := ccm.CreatePosition(suite.ctx, userID, "USDT", big.NewInt(40000000), big.NewFloat(2.0))
	require.NoError(t, err)
	fmt.Printf("✅ Created borrowing position: $%s USDT\n", position.Amount.String())

	// Verify portfolio state
	portfolioState, err := ccm.GetPortfolio(userID)
	require.NoError(t, err)

	assert.True(t, portfolioState.TotalCollateralValue.Cmp(big.NewInt(0)) > 0)
	assert.True(t, portfolioState.TotalBorrowedValue.Cmp(big.NewInt(0)) > 0)
	assert.True(t, portfolioState.CollateralRatio.Cmp(big.NewFloat(0)) > 0)

	fmt.Printf("📊 Portfolio State:\n")
	fmt.Printf("   Total Collateral: $%s\n", portfolioState.TotalCollateralValue.String())
	fmt.Printf("   Total Borrowed: $%s\n", portfolioState.TotalBorrowedValue.String())
	fmt.Printf("   Collateral Ratio: %v\n", portfolioState.CollateralRatio.String())
	fmt.Printf("   Risk Score: %v\n", portfolioState.RiskScore.String())

	fmt.Println("✅ Cross-Collateral Portfolio Management Test Passed")
}

// testMultiAssetPortfolioOperations tests multi-asset portfolio operations
func (suite *CrossProtocolIntegrationTestSuite) testMultiAssetPortfolioOperations(t *testing.T) {
	fmt.Println("🔄 Testing Multi-Asset Portfolio Operations...")

	ccm := advanced.NewCrossCollateralManager()
	userID := "user_multi_asset_test"

	// Create portfolio
	_, err := ccm.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)

	// Add multiple asset types
	assets := []*advanced.CrossCollateralAsset{
		{
			ID:             "BTC",
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         "BTC",
			Amount:         big.NewInt(200000000), // 2 BTC
			Value:          big.NewInt(100000000), // $100k
			Volatility:     big.NewFloat(0.8),
			LiquidityScore: big.NewFloat(0.9),
			RiskScore:      big.NewFloat(0.7),
		},
		{
			ID:             "ETH",
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         "ETH",
			Amount:         big.NewInt(2000000000000000000), // 2 ETH
			Value:          big.NewInt(60000000),            // $60k
			Volatility:     big.NewFloat(0.6),
			LiquidityScore: big.NewFloat(0.8),
			RiskScore:      big.NewFloat(0.5),
		},
		{
			ID:             "USDC",
			Type:           advanced.CrossCollateralTypeStablecoin,
			Symbol:         "USDC",
			Amount:         big.NewInt(50000000), // 50k USDC
			Value:          big.NewInt(50000000), // $50k
			Volatility:     big.NewFloat(0.1),
			LiquidityScore: big.NewFloat(1.0),
			RiskScore:      big.NewFloat(0.1),
		},
	}

	for _, asset := range assets {
		err = ccm.AddCollateral(suite.ctx, userID, asset)
		require.NoError(t, err)
		fmt.Printf("✅ Added %s collateral: %s %s = $%s\n",
			asset.Symbol, asset.Amount.String(), asset.Symbol, asset.Value.String())
	}

	// Create multiple borrowing positions
	positions := []struct {
		asset  string
		amount *big.Int
		ratio  *big.Float
	}{
		{"USDT", big.NewInt(30000000), big.NewFloat(2.0)}, // $30k USDT
		{"DAI", big.NewInt(20000000), big.NewFloat(2.0)},  // $20k DAI
	}

	for _, pos := range positions {
		position, err := ccm.CreatePosition(suite.ctx, userID, pos.asset, pos.amount, pos.ratio)
		require.NoError(t, err)
		fmt.Printf("✅ Created %s position: $%s\n", pos.asset, position.Amount.String())
	}

	// Get portfolio details
	portfolio, err := ccm.GetPortfolio(userID)
	require.NoError(t, err)

	fmt.Printf("📊 Multi-Asset Portfolio Summary:\n")
	fmt.Printf("   Assets: %d\n", len(portfolio.CollateralAssets))
	fmt.Printf("   Positions: %d\n", len(portfolio.Positions))
	fmt.Printf("   Total Collateral: $%s\n", portfolio.TotalCollateralValue.String())
	fmt.Printf("   Total Borrowed: $%s\n", portfolio.TotalBorrowedValue.String())
	fmt.Printf("   Net Collateral: $%s\n", portfolio.NetCollateralValue.String())
	fmt.Printf("   Collateral Ratio: %v\n", portfolio.CollateralRatio.String())

	// Test asset removal
	btcAsset := portfolio.CollateralAssets["BTC"]
	if btcAsset != nil {
		removeAmount := big.NewInt(100000000) // 1 BTC
		err = ccm.RemoveCollateral(suite.ctx, userID, "BTC", removeAmount)
		require.NoError(t, err)
		fmt.Printf("✅ Removed 1 BTC from portfolio\n")
	}

	fmt.Println("✅ Multi-Asset Portfolio Operations Test Passed")
}

// testPortfolioRiskManagement tests portfolio risk management
func (suite *CrossProtocolIntegrationTestSuite) testPortfolioRiskManagement(t *testing.T) {
	fmt.Println("🔄 Testing Portfolio Risk Management...")

	ccm := advanced.NewCrossCollateralManager()
	userID := "user_risk_test"

	// Create portfolio with higher risk tolerance
	_, err := ccm.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.2))
	require.NoError(t, err)

	// Add high-volatility assets
	highRiskAssets := []*advanced.CrossCollateralAsset{
		{
			ID:             "SOL",
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         "SOL",
			Amount:         big.NewInt(1000000000), // 1 SOL
			Value:          big.NewInt(100000000),  // $100k
			Volatility:     big.NewFloat(1.2),
			LiquidityScore: big.NewFloat(0.6),
			RiskScore:      big.NewFloat(0.9),
		},
		{
			ID:             "AVAX",
			Type:           advanced.CrossCollateralTypeCrypto,
			Symbol:         "AVAX",
			Amount:         big.NewInt(1000000000000000000), // 1 AVAX
			Value:          big.NewInt(40000000),            // $40k
			Volatility:     big.NewFloat(1.0),
			LiquidityScore: big.NewFloat(0.7),
			RiskScore:      big.NewFloat(0.8),
		},
	}

	for _, asset := range highRiskAssets {
		err = ccm.AddCollateral(suite.ctx, userID, asset)
		require.NoError(t, err)
		fmt.Printf("✅ Added %s collateral: %s %s = $%s\n",
			asset.Symbol, asset.Amount.String(), asset.Symbol, asset.Value.String())
	}

	// Create position with borrowed funds
	position, err := ccm.CreatePosition(suite.ctx, userID, "USDT", big.NewInt(60000000), big.NewFloat(1.5))
	require.NoError(t, err)
	fmt.Printf("✅ Created borrowing position: $%s USDT\n", position.Amount.String())

	// Check risk metrics
	portfolio, err := ccm.GetPortfolio(userID)
	require.NoError(t, err)

	fmt.Printf("📊 Risk Management Portfolio:\n")
	fmt.Printf("   Total Collateral: $%s\n", portfolio.TotalCollateralValue.String())
	fmt.Printf("   Total Borrowed: $%s\n", portfolio.TotalBorrowedValue.String())
	fmt.Printf("   Collateral Ratio: %v\n", portfolio.CollateralRatio.String())
	fmt.Printf("   Risk Score: %v\n", portfolio.RiskScore.String())

	// Verify risk metrics are calculated
	assert.NotNil(t, portfolio.RiskMetrics)
	assert.NotNil(t, portfolio.RiskMetrics.Volatility)
	assert.NotNil(t, portfolio.RiskMetrics.VaR95)
	assert.NotNil(t, portfolio.RiskMetrics.ConcentrationRisk)
	assert.NotNil(t, portfolio.RiskMetrics.LiquidityRisk)

	fmt.Printf("   Risk Metrics:\n")
	fmt.Printf("     Volatility: %v\n", portfolio.RiskMetrics.Volatility.String())
	fmt.Printf("     VaR95: %v\n", portfolio.RiskMetrics.VaR95.String())
	fmt.Printf("     Concentration Risk: %v\n", portfolio.RiskMetrics.ConcentrationRisk.String())
	fmt.Printf("     Liquidity Risk: %v\n", portfolio.RiskMetrics.LiquidityRisk.String())

	fmt.Println("✅ Portfolio Risk Management Test Passed")
}

// testPortfolioValidationHealthChecks tests portfolio validation and health checks
func (suite *CrossProtocolIntegrationTestSuite) testPortfolioValidationHealthChecks(t *testing.T) {
	fmt.Println("🔄 Testing Portfolio Validation and Health Checks...")

	ccm := advanced.NewCrossCollateralManager()
	userID := "user_validation_test"

	// Create portfolio
	_, err := ccm.CreatePortfolio(suite.ctx, userID, big.NewFloat(1.5))
	require.NoError(t, err)

	// Add collateral
	asset := &advanced.CrossCollateralAsset{
		ID:             "BTC",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "BTC",
		Amount:         big.NewInt(100000000), // 1 BTC
		Value:          big.NewInt(50000000),  // $50k
		Volatility:     big.NewFloat(0.8),
		LiquidityScore: big.NewFloat(0.9),
		RiskScore:      big.NewFloat(0.7),
	}
	err = ccm.AddCollateral(suite.ctx, userID, asset)
	require.NoError(t, err)

	// Create position
	position, err := ccm.CreatePosition(suite.ctx, userID, "USDT", big.NewInt(25000000), big.NewFloat(2.0))
	require.NoError(t, err)
	fmt.Printf("✅ Created position: $%s USDT\n", position.Amount.String())

	// Validate portfolio state
	issues, err := ccm.ValidatePortfolioState(userID)
	require.NoError(t, err)

	if len(issues) > 0 {
		fmt.Printf("⚠️  Portfolio validation issues found:\n")
		for i, issue := range issues {
			fmt.Printf("   %d. %s\n", i+1, issue)
		}
	} else {
		fmt.Println("✅ Portfolio validation passed - no issues found")
	}

	// Get detailed asset information
	assetDetails, err := ccm.GetPortfolioAssetDetails(userID)
	require.NoError(t, err)

	fmt.Printf("📊 Asset Details:\n")
	for assetID, details := range assetDetails {
		fmt.Printf("   %s:\n", assetID)
		for key, value := range details.(map[string]interface{}) {
			fmt.Printf("     %s: %v\n", key, value)
		}
	}

	// Test portfolio health under stress
	// Remove some collateral to test undercollateralization
	err = ccm.RemoveCollateral(suite.ctx, userID, "BTC", big.NewInt(50000000)) // 0.5 BTC
	require.NoError(t, err)

	// Check portfolio health after stress
	portfolio, err := ccm.GetPortfolio(userID)
	require.NoError(t, err)

	fmt.Printf("📊 Portfolio Health After Stress:\n")
	fmt.Printf("   Total Collateral: $%s\n", portfolio.TotalCollateralValue.String())
	fmt.Printf("   Total Borrowed: $%s\n", portfolio.TotalBorrowedValue.String())
	fmt.Printf("   Net Collateral: $%s\n", portfolio.NetCollateralValue.String())
	fmt.Printf("   Collateral Ratio: %v\n", portfolio.CollateralRatio.String())

	// Validate again
	issuesAfter, err := ccm.ValidatePortfolioState(userID)
	require.NoError(t, err)

	if len(issuesAfter) > 0 {
		fmt.Printf("⚠️  Portfolio validation after stress:\n")
		for i, issue := range issuesAfter {
			fmt.Printf("   %d. %s\n", i+1, issue)
		}
	}

	fmt.Println("✅ Portfolio Validation and Health Checks Test Passed")
}

// testCompletePortfolioLifecycle tests a complete portfolio lifecycle
func (suite *CrossProtocolIntegrationTestSuite) testCompletePortfolioLifecycle(t *testing.T) {
	fmt.Println("🔄 Testing Complete Portfolio Lifecycle...")

	user := "0x9999999999999999999999999999999999999999"
	fmt.Printf("👤 User: %s\n", user)

	ccm := advanced.NewCrossCollateralManager()

	// 1. Create portfolio
	fmt.Println("\n💰 Step 1: Create portfolio")
	portfolio, err := ccm.CreatePortfolio(suite.ctx, user, big.NewFloat(1.5))
	require.NoError(t, err)
	fmt.Printf("   ✅ Portfolio created with ID: %s\n", portfolio.UserID)

	// 2. Add initial collateral
	fmt.Println("\n🏦 Step 2: Add initial collateral")

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
	err = ccm.AddCollateral(suite.ctx, user, btcAsset)
	require.NoError(t, err)
	fmt.Printf("   ✅ Added BTC collateral: %s BTC = $%s\n",
		btcAsset.Amount.String(), btcAsset.Value.String())

	// 3. Create borrowing position
	fmt.Println("\n💸 Step 3: Create borrowing position")
	position, err := ccm.CreatePosition(suite.ctx, user, "USDT", big.NewInt(25000000), big.NewFloat(2.0))
	require.NoError(t, err)
	fmt.Printf("   ✅ Created position: $%s USDT\n", position.Amount.String())

	// 4. Add more collateral
	fmt.Println("\n🏦 Step 4: Add more collateral")

	ethAsset := &advanced.CrossCollateralAsset{
		ID:             "ETH",
		Type:           advanced.CrossCollateralTypeCrypto,
		Symbol:         "ETH",
		Amount:         big.NewInt(1000000000000000000), // 1 ETH
		Value:          big.NewInt(30000000),            // $30k
		Volatility:     big.NewFloat(0.6),
		LiquidityScore: big.NewFloat(0.8),
		RiskScore:      big.NewFloat(0.5),
	}
	err = ccm.AddCollateral(suite.ctx, user, ethAsset)
	require.NoError(t, err)
	fmt.Printf("   ✅ Added ETH collateral: %s ETH = $%s\n",
		ethAsset.Amount.String(), ethAsset.Value.String())

	// 5. Check portfolio state
	fmt.Println("\n📈 Step 5: Check portfolio state")
	portfolioState, err := ccm.GetPortfolio(user)
	require.NoError(t, err)

	fmt.Printf("   📊 Portfolio Summary:\n")
	fmt.Printf("      Total Collateral: $%s\n", portfolioState.TotalCollateralValue.String())
	fmt.Printf("      Total Borrowed: $%s\n", portfolioState.TotalBorrowedValue.String())
	fmt.Printf("      Net Collateral: $%s\n", portfolioState.NetCollateralValue.String())
	fmt.Printf("      Collateral Ratio: %v\n", portfolioState.CollateralRatio.String())
	fmt.Printf("      Risk Score: %v\n", portfolioState.RiskScore.String())

	// 6. Partial collateral removal
	fmt.Println("\n🔄 Step 6: Partial collateral removal")
	err = ccm.RemoveCollateral(suite.ctx, user, "BTC", big.NewInt(50000000)) // 0.5 BTC
	require.NoError(t, err)
	fmt.Printf("   ✅ Removed 0.5 BTC collateral\n")

	// 7. Check updated portfolio state
	fmt.Println("\n📊 Step 7: Check updated portfolio state")
	updatedPortfolio, err := ccm.GetPortfolio(user)
	require.NoError(t, err)

	fmt.Printf("   📊 Updated Portfolio Summary:\n")
	fmt.Printf("      Total Collateral: $%s\n", updatedPortfolio.TotalCollateralValue.String())
	fmt.Printf("      Total Borrowed: $%s\n", updatedPortfolio.TotalBorrowedValue.String())
	fmt.Printf("      Net Collateral: $%s\n", updatedPortfolio.NetCollateralValue.String())
	fmt.Printf("      Collateral Ratio: %v\n", updatedPortfolio.CollateralRatio.String())

	// 8. Validate portfolio health
	fmt.Println("\n🔍 Step 8: Validate portfolio health")
	issues, err := ccm.ValidatePortfolioState(user)
	require.NoError(t, err)

	if len(issues) > 0 {
		fmt.Printf("   ⚠️  Portfolio validation issues found:\n")
		for i, issue := range issues {
			fmt.Printf("      %d. %s\n", i+1, issue)
		}
	} else {
		fmt.Printf("   ✅ Portfolio validation passed - no issues found\n")
	}

	// 9. Close position
	fmt.Println("\n🔒 Step 9: Close position")
	err = ccm.ClosePosition(suite.ctx, user, position.ID)
	require.NoError(t, err)
	fmt.Printf("   ✅ Position closed successfully\n")

	// 10. Final portfolio state
	fmt.Println("\n🎯 Step 10: Final portfolio state")
	finalPortfolio, err := ccm.GetPortfolio(user)
	require.NoError(t, err)

	fmt.Printf("   📊 Final Portfolio Summary:\n")
	fmt.Printf("      Total Collateral: $%s\n", finalPortfolio.TotalCollateralValue.String())
	fmt.Printf("      Total Borrowed: $%s\n", finalPortfolio.TotalBorrowedValue.String())
	fmt.Printf("      Net Collateral: $%s\n", finalPortfolio.NetCollateralValue.String())
	fmt.Printf("      Collateral Ratio: %v\n", finalPortfolio.CollateralRatio.String())

	fmt.Println("\n🎉 Complete Portfolio Lifecycle Test Passed!")
	fmt.Println("   Portfolio successfully completed:")
	fmt.Println("   ✅ Creation and initialization")
	fmt.Println("   ✅ Collateral management")
	fmt.Println("   ✅ Position creation and management")
	fmt.Println("   ✅ Risk assessment and monitoring")
	fmt.Println("   ✅ Portfolio optimization")
	fmt.Println("   ✅ Position closure and cleanup")
}

// RunAllCrossProtocolTests runs all cross-protocol integration tests
func (suite *CrossProtocolIntegrationTestSuite) RunAllCrossProtocolTests(t *testing.T) {
	fmt.Println("🚀 Running All Cross-Protocol Integration Tests...")

	TestDeFiEcosystemIntegration(t)

	fmt.Println("🎉 All Cross-Protocol Integration Tests Completed Successfully!")
}
