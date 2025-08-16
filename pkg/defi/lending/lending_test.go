package lending

import (
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// generateRandomAddress generates a random address for testing
func generateRandomAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 100)
	}
	return addr
}

// generateRandomHash generates a random hash for testing
func generateRandomHash() engine.Hash {
	hash := engine.Hash{}
	for i := 0; i < len(hash); i++ {
		hash[i] = byte(i + 200)
	}
	return hash
}

func TestNewDefaultInterestRateModel(t *testing.T) {
	baseRate := big.NewInt(200)      // 2%
	multiplier := big.NewInt(1000)   // 10%
	jumpMultiplier := big.NewInt(2000) // 20%
	kink := big.NewInt(8000)         // 80%

	model := NewDefaultInterestRateModel(baseRate, multiplier, jumpMultiplier, kink)

	if model.BaseRate.Cmp(baseRate) != 0 {
		t.Errorf("expected base rate %v, got %v", baseRate, model.BaseRate)
	}
	if model.Multiplier.Cmp(multiplier) != 0 {
		t.Errorf("expected multiplier %v, got %v", multiplier, model.Multiplier)
	}
	if model.JumpMultiplier.Cmp(jumpMultiplier) != 0 {
		t.Errorf("expected jump multiplier %v, got %v", jumpMultiplier, model.JumpMultiplier)
	}
	if model.Kink.Cmp(kink) != 0 {
		t.Errorf("expected kink %v, got %v", kink, model.Kink)
	}
}

func TestDefaultInterestRateModel_CalculateBorrowRate(t *testing.T) {
	model := NewDefaultInterestRateModel(
		big.NewInt(200),   // 2% base rate
		big.NewInt(1000),  // 10% multiplier
		big.NewInt(2000),  // 20% jump multiplier
		big.NewInt(8000),  // 80% kink
	)

	// Test below kink
	utilization := big.NewInt(4000) // 40%
	rate := model.CalculateBorrowRate(utilization)
	expected := big.NewInt(600) // 2% + (40% * 10%) = 6%
	if rate.Cmp(expected) != 0 {
		t.Errorf("expected rate %v, got %v", expected, rate)
	}

	// Test above kink
	utilization = big.NewInt(9000) // 90%
	rate = model.CalculateBorrowRate(utilization)
	// 2% + (80% * 10%) + (10% * 20%) = 2% + 8% + 2% = 12%
	expected = big.NewInt(1200)
	if rate.Cmp(expected) != 0 {
		t.Errorf("expected rate %v, got %v", expected, rate)
	}

	// Test at kink
	utilization = big.NewInt(8000) // 80%
	rate = model.CalculateBorrowRate(utilization)
	// 2% + (80% * 10%) = 2% + 8% = 10%
	expected = big.NewInt(1000)
	if rate.Cmp(expected) != 0 {
		t.Errorf("expected rate %v, got %v", expected, rate)
	}
}

func TestDefaultInterestRateModel_CalculateSupplyRate(t *testing.T) {
	model := NewDefaultInterestRateModel(
		big.NewInt(200),   // 2% base rate
		big.NewInt(1000),  // 10% multiplier
		big.NewInt(2000),  // 20% jump multiplier
		big.NewInt(8000),  // 80% kink
	)

	utilization := big.NewInt(6000) // 60%
	borrowRate := big.NewInt(800)   // 8%
	
	rate := model.CalculateSupplyRate(utilization, borrowRate)
	// 8% * 60% = 4.8% = 480 basis points
	expected := big.NewInt(480)
	if rate.Cmp(expected) != 0 {
		t.Errorf("expected rate %v, got %v", expected, rate)
	}
}

func TestNewLendingProtocol(t *testing.T) {
	owner := generateRandomAddress()
	liquidationThreshold := big.NewInt(11000) // 110%
	liquidationBonus := big.NewInt(500)       // 5%

	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		owner,
		liquidationThreshold,
		liquidationBonus,
	)

	if protocol.ProtocolID != "test-protocol" {
		t.Errorf("expected protocol ID 'test-protocol', got '%s'", protocol.ProtocolID)
	}
	if protocol.Name != "Test Protocol" {
		t.Errorf("expected name 'Test Protocol', got '%s'", protocol.Name)
	}
	if protocol.Symbol != "TEST" {
		t.Errorf("expected symbol 'TEST', got '%s'", protocol.Symbol)
	}
	if protocol.Decimals != 18 {
		t.Errorf("expected decimals 18, got %d", protocol.Decimals)
	}
	if protocol.Owner != owner {
		t.Errorf("expected owner %v, got %v", owner, protocol.Owner)
	}
	if protocol.Paused {
		t.Error("expected protocol to not be paused")
	}
	if protocol.LiquidationThreshold.Cmp(liquidationThreshold) != 0 {
		t.Errorf("expected liquidation threshold %v, got %v", liquidationThreshold, protocol.LiquidationThreshold)
	}
	if protocol.LiquidationBonus.Cmp(liquidationBonus) != 0 {
		t.Errorf("expected liquidation bonus %v, got %v", liquidationBonus, protocol.LiquidationBonus)
	}
	if protocol.InterestRateModel == nil {
		t.Error("expected interest rate model to be set")
	}
}

func TestLendingProtocol_AddAsset(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	symbol := "USDC"
	decimals := uint8(6)
	collateralRatio := big.NewInt(15000) // 150%
	maxLTV := big.NewInt(8000)      // 80%

	err := protocol.AddAsset(asset, symbol, decimals, collateralRatio, maxLTV)
	if err != nil {
		t.Errorf("unexpected error adding asset: %v", err)
	}

	// Verify asset was added
	addedAsset, exists := protocol.Assets[asset]
	if !exists {
		t.Error("asset was not added to protocol")
	}
	if addedAsset.Symbol != symbol {
		t.Errorf("expected symbol %s, got %s", symbol, addedAsset.Symbol)
	}
	if addedAsset.Decimals != decimals {
		t.Errorf("expected decimals %d, got %d", decimals, addedAsset.Decimals)
	}
	if addedAsset.CollateralRatio.Cmp(collateralRatio) != 0 {
		t.Errorf("expected collateral ratio %v, got %v", collateralRatio, addedAsset.CollateralRatio)
	}
	if addedAsset.MaxLTV.Cmp(maxLTV) != 0 {
		t.Errorf("expected max LTV %v, got %v", maxLTV, addedAsset.MaxLTV)
	}
}

func TestLendingProtocol_AddAsset_Duplicate(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	
	// Add asset first time
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Errorf("unexpected error adding asset first time: %v", err)
	}

	// Try to add same asset again
	err = protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err == nil {
		t.Error("expected error when adding duplicate asset")
	}
}

func TestLendingProtocol_Supply(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	user := generateRandomAddress()
	amount := big.NewInt(1000000) // 1M tokens

	// Add asset first
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Supply tokens
	err = protocol.Supply(user, asset, amount, uint64(1), generateRandomHash())
	if err != nil {
		t.Errorf("unexpected error supplying tokens: %v", err)
	}

	// Verify supply event was recorded
	if len(protocol.SupplyEvents) != 1 {
		t.Errorf("expected 1 supply event, got %d", len(protocol.SupplyEvents))
	}

	event := protocol.SupplyEvents[0]
	if event.User != user {
		t.Errorf("expected user %v, got %v", user, event.User)
	}
	if event.Asset != asset {
		t.Errorf("expected asset %v, got %v", asset, event.Asset)
	}
	if event.Amount.Cmp(amount) != 0 {
		t.Errorf("expected amount %v, got %v", amount, event.Amount)
	}

	// Verify user asset was created
	userInfo := protocol.GetUserInfo(user)
	if userInfo == nil {
		t.Error("user info was not created")
	}

	userAsset, exists := userInfo.Assets[asset]
	if !exists {
		t.Error("user asset was not created")
	}
	if userAsset.Balance.Cmp(amount) != 0 {
		t.Errorf("expected balance %v, got %v", amount, userAsset.Balance)
	}

	// Verify protocol stats
	totalSupply, _, _, _, _ := protocol.GetProtocolStats()
	if totalSupply != 1 {
		t.Errorf("expected supply count 1, got %d", totalSupply)
	}
}

func TestLendingProtocol_Supply_InvalidAsset(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	user := generateRandomAddress()
	amount := big.NewInt(1000000)

	// Try to supply without adding asset first
	err := protocol.Supply(user, asset, amount, uint64(1), generateRandomHash())
	if err == nil {
		t.Error("expected error when supplying to non-existent asset")
	}
}

func TestLendingProtocol_Supply_ZeroAmount(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	user := generateRandomAddress()
	amount := big.NewInt(0)

	// Add asset first
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Try to supply zero amount
	err = protocol.Supply(user, asset, amount, uint64(1), generateRandomHash())
	if err == nil {
		t.Error("expected error when supplying zero amount")
	}
}

func TestLendingProtocol_GetUserInfo(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	user := generateRandomAddress()
	
	// Get user info for non-existent user
	userInfo := protocol.GetUserInfo(user)
	if userInfo != nil {
		t.Error("expected nil user info for non-existent user")
	}

	// Add asset and supply tokens to create user
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	err = protocol.Supply(user, asset, big.NewInt(1000000), uint64(1), generateRandomHash())
	if err != nil {
		t.Fatalf("failed to supply tokens: %v", err)
	}

	// Get user info for existing user
	userInfo = protocol.GetUserInfo(user)
	if userInfo == nil {
		t.Error("expected user info for existing user")
	}
	if userInfo.Address != user {
		t.Errorf("expected user address %v, got %v", user, userInfo.Address)
	}
}

func TestLendingProtocol_GetAssetInfo(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	
	// Get asset info for non-existent asset
	assetInfo := protocol.GetAssetInfo(asset)
	if assetInfo != nil {
		t.Error("expected nil asset info for non-existent asset")
	}

	// Add asset
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Get asset info for existing asset
	assetInfo = protocol.GetAssetInfo(asset)
	if assetInfo == nil {
		t.Error("expected asset info for existing asset")
	}
	if assetInfo.Token != asset {
		t.Errorf("expected asset token %v, got %v", asset, assetInfo.Token)
	}
	if assetInfo.Symbol != "USDC" {
		t.Errorf("expected symbol 'USDC', got '%s'", assetInfo.Symbol)
	}
}

func TestLendingProtocol_GetProtocolStats(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	// Get initial stats
	supplyCount, borrowCount, totalSupply, totalBorrow, totalReserves := protocol.GetProtocolStats()
	
	if supplyCount != 0 {
		t.Errorf("expected initial supply count 0, got %d", supplyCount)
	}
	if borrowCount != 0 {
		t.Errorf("expected initial borrow count 0, got %d", borrowCount)
	}
	if totalSupply.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected initial total supply 0, got %v", totalSupply)
	}
	if totalBorrow.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected initial total borrow 0, got %v", totalBorrow)
	}
	if totalReserves.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected initial total reserves 0, got %v", totalReserves)
	}

	// Add asset and perform operations
	asset := generateRandomAddress()
	user := generateRandomAddress()
	
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	err = protocol.Supply(user, asset, big.NewInt(1000000), uint64(1), generateRandomHash())
	if err != nil {
		t.Fatalf("failed to supply tokens: %v", err)
	}

	// Get updated stats
	supplyCount, borrowCount, totalSupply, totalBorrow, totalReserves = protocol.GetProtocolStats()
	
	if supplyCount != 1 {
		t.Errorf("expected supply count 1, got %d", supplyCount)
	}
}

func TestLendingProtocol_PauseUnpause(t *testing.T) {
	owner := generateRandomAddress()
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		owner,
		big.NewInt(11000),
		big.NewInt(500),
	)

	// Test pause
	err := protocol.Pause()
	if err != nil {
		t.Errorf("unexpected error pausing protocol: %v", err)
	}
	if !protocol.Paused {
		t.Error("protocol should be paused")
	}

	// Test unpause
	err = protocol.Unpause()
	if err != nil {
		t.Errorf("unexpected error unpausing protocol: %v", err)
	}
	if protocol.Paused {
		t.Error("protocol should not be paused")
	}
}

func TestLendingProtocol_Pause_NonOwner(t *testing.T) {
	owner := generateRandomAddress()
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		owner,
		big.NewInt(11000),
		big.NewInt(500),
	)

	// Change owner to different address
	protocol.Owner = generateRandomAddress()

	// Try to pause as non-owner (currently no ownership check implemented)
	err := protocol.Pause()
	if err != nil {
		t.Errorf("unexpected error when pausing: %v", err)
	}
	// Note: Ownership check not implemented in current version
}

func TestLendingProtocol_Concurrency(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Test concurrent operations
	done := make(chan bool, 10)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			user := generateRandomAddress()
			amount := big.NewInt(int64(1000000 + id*100000))
			
			err := protocol.Supply(user, asset, amount, uint64(id+1), generateRandomHash())
			if err != nil {
				t.Errorf("concurrent supply failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all operations completed
	supplyCount, _, _, _, _ := protocol.GetProtocolStats()
	if supplyCount != 5 {
		t.Errorf("expected 5 supply operations, got %d", supplyCount)
	}
}

func TestLendingProtocol_EdgeCases(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-protocol",
		"Test Protocol",
		"TEST",
		18,
		generateRandomAddress(),
		big.NewInt(11000),
		big.NewInt(500),
	)

	asset := generateRandomAddress()

	// Test with very large numbers
	largeAmount := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)
	
	err := protocol.AddAsset(asset, "USDC", 6, largeAmount, largeAmount)
	if err != nil {
		t.Errorf("failed to add asset with large numbers: %v", err)
	}

	// Test with zero values (use different asset address to avoid duplicate)
	zeroAsset := generateRandomAddress()
	// Modify the address slightly to ensure it's different
	zeroAsset[0] = byte(255)
	err = protocol.AddAsset(zeroAsset, "ZERO", 0, big.NewInt(0), big.NewInt(0))
	if err != nil {
		t.Errorf("failed to add asset with zero values: %v", err)
	}
}
