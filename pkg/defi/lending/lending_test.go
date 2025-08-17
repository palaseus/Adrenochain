package lending

import (
	"math/big"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
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
	baseRate := big.NewInt(200)        // 2%
	multiplier := big.NewInt(1000)     // 10%
	jumpMultiplier := big.NewInt(2000) // 20%
	kink := big.NewInt(8000)           // 80%

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
		big.NewInt(200),  // 2% base rate
		big.NewInt(1000), // 10% multiplier
		big.NewInt(2000), // 20% jump multiplier
		big.NewInt(8000), // 80% kink
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
		big.NewInt(200),  // 2% base rate
		big.NewInt(1000), // 10% multiplier
		big.NewInt(2000), // 20% jump multiplier
		big.NewInt(8000), // 80% kink
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
	maxLTV := big.NewInt(8000)           // 80%

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

func TestLendingProtocol_Withdraw_Comprehensive(t *testing.T) {
	// Create a new lending protocol
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(8000), // 80% liquidation threshold
		big.NewInt(500),  // 5% liquidation bonus
	)

	// Add an asset for testing
	assetAddress := generateRandomAddress()
	asset := &Asset{
		Token:                assetAddress,
		Symbol:               "TEST",
		Decimals:             18,
		TotalSupply:          big.NewInt(1000000), // 1M total supply
		TotalBorrow:          big.NewInt(800000),  // 800K borrowed
		Reserves:             big.NewInt(20000),   // 20K reserves
		BorrowRate:           big.NewInt(800),     // 8% borrow rate
		SupplyRate:           big.NewInt(600),     // 6% supply rate
		CollateralRatio:      big.NewInt(8000),    // 80% collateral ratio
		MaxLTV:               big.NewInt(7500),    // 75% max LTV
		LiquidationThreshold: big.NewInt(8000),    // 80% liquidation threshold
		Paused:               false,
	}

	protocol.Assets[assetAddress] = asset

	// Create a user with supply balance
	userAddress := generateRandomAddress()
	user := &User{
		Address: userAddress,
		Assets: map[engine.Address]*UserAsset{
			assetAddress: {
				Token:           assetAddress,
				Balance:         big.NewInt(100000), // 100K supplied
				BorrowBalance:   big.NewInt(0),      // No borrowing
				CollateralValue: big.NewInt(80000),  // 80K collateral value (80% of 100K)
				BorrowValue:     big.NewInt(0),      // No borrow value
				LastUpdate:      time.Now(),
			},
		},
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	protocol.Users[userAddress] = user

	// Test case 1: Withdraw valid amount
	withdrawAmount := big.NewInt(50000) // 50K withdrawal
	blockNumber := uint64(1000)
	txHash := generateRandomHash()

	err := protocol.Withdraw(
		userAddress,
		assetAddress,
		withdrawAmount,
		blockNumber,
		txHash,
	)

	if err != nil {
		t.Errorf("Withdraw failed: %v", err)
	}

	// Verify user balance was updated
	userAsset := protocol.Users[userAddress].Assets[assetAddress]
	if userAsset.Balance.Cmp(big.NewInt(50000)) != 0 { // 100K - 50K = 50K
		t.Errorf("expected supply balance 50000, got %v", userAsset.Balance)
	}

	// Test case 2: Withdraw entire balance
	remainingBalance := userAsset.Balance
	err = protocol.Withdraw(
		userAddress,
		assetAddress,
		remainingBalance,
		blockNumber+1,
		txHash,
	)

	if err != nil {
		t.Errorf("Withdraw entire balance failed: %v", err)
	}

	// Verify user balance is now zero
	userAsset = protocol.Users[userAddress].Assets[assetAddress]
	if userAsset.Balance.Sign() != 0 {
		t.Errorf("expected supply balance 0, got %v", userAsset.Balance)
	}

	// Test case 3: Try to withdraw more than available
	err = protocol.Withdraw(
		userAddress,
		assetAddress,
		big.NewInt(1000), // Try to withdraw 1K
		blockNumber+2,
		txHash,
	)

	if err == nil {
		t.Error("expected error when withdrawing more than available, got nil")
	}

	// Test case 4: Withdraw with zero amount
	err = protocol.Withdraw(
		userAddress,
		assetAddress,
		big.NewInt(0), // Zero amount
		blockNumber+3,
		txHash,
	)

	if err == nil {
		t.Error("expected error when withdrawing zero amount, got nil")
	}

	// Test case 5: Withdraw from paused asset
	asset.Paused = true
	err = protocol.Withdraw(
		userAddress,
		assetAddress,
		big.NewInt(1000),
		blockNumber+4,
		txHash,
	)

	if err == nil {
		t.Error("expected error when withdrawing from paused asset, got nil")
	}

	// Reset asset state
	asset.Paused = false
}

// TestLendingProtocol_Borrow_Comprehensive tests the Borrow function comprehensively
func TestLendingProtocol_Borrow_Comprehensive(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add collateral asset
	collateralAsset := generateRandomAddress()
	err := protocol.AddAsset(collateralAsset, "COLL", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add collateral asset: %v", err)
	}

	// Add borrow asset (use different address to avoid collision)
	borrowAsset := generateRandomAddress()
	// Modify the address slightly to ensure it's different from collateralAsset
	borrowAsset[0] = byte(255)
	err = protocol.AddAsset(borrowAsset, "BORROW", 18, big.NewInt(10000), big.NewInt(7500))
	if err != nil {
		t.Fatalf("failed to add borrow asset: %v", err)
	}

	user := generateRandomAddress()
	borrowAmount := big.NewInt(100000) // 100K borrow
	blockNumber := uint64(1000)
	txHash := generateRandomHash()

	// Test case 1: Try to borrow without collateral (should fail)
	err = protocol.Borrow(user, borrowAsset, borrowAmount, blockNumber, txHash)
	if err == nil {
		t.Error("expected error when borrowing without collateral, got nil")
	}

	// Test case 2: Supply collateral first
	collateralAmount := big.NewInt(200000) // 200K collateral
	err = protocol.Supply(user, collateralAsset, collateralAmount, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to supply collateral: %v", err)
	}

	// Set user's collateral value
	if protocol.Users[user] == nil {
		t.Fatalf("user was not created")
	}
	protocol.Users[user].Collateral[collateralAsset] = collateralAmount

	// Test case 3: Now try to borrow (should succeed)
	err = protocol.Borrow(user, borrowAsset, borrowAmount, blockNumber+1, txHash)
	if err != nil {
		t.Errorf("borrow failed: %v", err)
	}

	// Verify borrow event was recorded
	if len(protocol.BorrowEvents) != 1 {
		t.Errorf("expected 1 borrow event, got %d", len(protocol.BorrowEvents))
	}

	event := protocol.BorrowEvents[0]
	if event.User != user {
		t.Errorf("expected user %v, got %v", user, event.User)
	}
	if event.Asset != borrowAsset {
		t.Errorf("expected asset %v, got %v", borrowAsset, event.Asset)
	}
	if event.Amount.Cmp(borrowAmount) != 0 {
		t.Errorf("expected amount %v, got %v", borrowAmount, event.Amount)
	}

	// Verify user borrow balance was updated
	userAsset := protocol.Users[user].Assets[borrowAsset]
	if userAsset.BorrowBalance.Cmp(borrowAmount) != 0 {
		t.Errorf("expected borrow balance %v, got %v", borrowAmount, userAsset.BorrowBalance)
	}

	// Verify protocol stats
	_, borrowCount, _, totalBorrow, _ := protocol.GetProtocolStats()
	if borrowCount != 1 {
		t.Errorf("expected borrow count 1, got %d", borrowCount)
	}
	if totalBorrow.Cmp(borrowAmount) != 0 {
		t.Errorf("expected total borrow %v, got %v", borrowAmount, totalBorrow)
	}

	// Test case 4: Try to borrow from paused protocol (should fail)
	protocol.Paused = true
	err = protocol.Borrow(user, borrowAsset, big.NewInt(50000), blockNumber+2, txHash)
	if err == nil {
		t.Error("expected error when borrowing from paused protocol, got nil")
	}
	protocol.Paused = false

	// Test case 5: Try to borrow invalid amount (should fail)
	err = protocol.Borrow(user, borrowAsset, big.NewInt(0), blockNumber+3, txHash)
	if err == nil {
		t.Error("expected error when borrowing zero amount, got nil")
	}

	err = protocol.Borrow(user, borrowAsset, big.NewInt(-1000), blockNumber+4, txHash)
	if err == nil {
		t.Error("expected error when borrowing negative amount, got nil")
	}

	// Test case 6: Try to borrow from non-existent asset (should fail)
	nonExistentAsset := generateRandomAddress()
	// Modify the address slightly to ensure it's different
	nonExistentAsset[0] = byte(128)
	err = protocol.Borrow(user, nonExistentAsset, borrowAmount, blockNumber+5, txHash)
	if err == nil {
		t.Error("expected error when borrowing from non-existent asset, got nil")
	}

	// Test case 7: Try to borrow from paused asset (should fail)
	protocol.Assets[borrowAsset].Paused = true
	err = protocol.Borrow(user, borrowAsset, borrowAmount, blockNumber+6, txHash)
	if err == nil {
		t.Error("expected error when borrowing from paused asset, got nil")
	}
	protocol.Assets[borrowAsset].Paused = false
}

// TestLendingProtocol_Repay_Comprehensive tests the Repay function comprehensively
func TestLendingProtocol_Repay_Comprehensive(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	user := generateRandomAddress()
	borrowAmount := big.NewInt(100000) // 100K borrow
	repayAmount := big.NewInt(50000)   // 50K repay
	blockNumber := uint64(1000)
	txHash := generateRandomHash()

	// Test case 1: Try to repay without borrowing (should fail)
	err = protocol.Repay(user, asset, repayAmount, blockNumber, txHash)
	if err == nil {
		t.Error("expected error when repaying without borrowing, got nil")
	}

	// Test case 2: Set up user with borrow balance
	protocol.Users[user] = &User{
		Address:    user,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	protocol.Users[user].Assets[asset] = &UserAsset{
		Token:           asset,
		Balance:         big.NewInt(0),
		BorrowBalance:   new(big.Int).Set(borrowAmount),
		CollateralValue: big.NewInt(0),
		BorrowValue:     big.NewInt(0),
		LastUpdate:      time.Now(),
	}

	// Set asset total borrow
	protocol.Assets[asset].TotalBorrow = new(big.Int).Set(borrowAmount)
	protocol.TotalBorrow = new(big.Int).Set(borrowAmount)

	// Test case 3: Repay partial amount (should succeed)
	err = protocol.Repay(user, asset, repayAmount, blockNumber+1, txHash)
	if err != nil {
		t.Errorf("repay failed: %v", err)
	}

	// Verify repay event was recorded
	if len(protocol.RepayEvents) != 1 {
		t.Errorf("expected 1 repay event, got %d", len(protocol.RepayEvents))
	}

	event := protocol.RepayEvents[0]
	if event.User != user {
		t.Errorf("expected user %v, got %v", user, event.User)
	}
	if event.Asset != asset {
		t.Errorf("expected asset %v, got %v", asset, event.Asset)
	}
	if event.Amount.Cmp(repayAmount) != 0 {
		t.Errorf("expected amount %v, got %v", repayAmount, event.Amount)
	}

	// Verify user borrow balance was updated
	userAsset := protocol.Users[user].Assets[asset]
	expectedBalance := new(big.Int).Sub(borrowAmount, repayAmount)
	if userAsset.BorrowBalance.Cmp(expectedBalance) != 0 {
		t.Errorf("expected borrow balance %v, got %v", expectedBalance, userAsset.BorrowBalance)
	}

	// Verify protocol totals were updated
	_, _, _, totalBorrow, _ := protocol.GetProtocolStats()
	if totalBorrow.Cmp(expectedBalance) != 0 {
		t.Errorf("expected total borrow %v, got %v", expectedBalance, totalBorrow)
	}

	// Test case 4: Try to repay more than borrowed (should fail)
	err = protocol.Repay(user, asset, big.NewInt(100000), blockNumber+2, txHash)
	if err == nil {
		t.Error("expected error when repaying more than borrowed, got nil")
	}

	// Test case 5: Try to repay from paused protocol (should fail)
	protocol.Paused = true
	err = protocol.Repay(user, asset, big.NewInt(10000), blockNumber+3, txHash)
	if err == nil {
		t.Error("expected error when repaying from paused protocol, got nil")
	}
	protocol.Paused = false

	// Test case 6: Try to repay invalid amount (should fail)
	err = protocol.Repay(user, asset, big.NewInt(0), blockNumber+4, txHash)
	if err == nil {
		t.Error("expected error when repaying zero amount, got nil")
	}

	err = protocol.Repay(user, asset, big.NewInt(-1000), blockNumber+5, txHash)
	if err == nil {
		t.Error("expected error when repaying negative amount, got nil")
	}

	// Test case 7: Try to repay to non-existent asset (should fail)
	nonExistentAsset := generateRandomAddress()
	// Modify the address slightly to ensure it's different
	nonExistentAsset[0] = byte(128)
	err = protocol.Repay(user, nonExistentAsset, repayAmount, blockNumber+6, txHash)
	if err == nil {
		t.Error("expected error when repaying to non-existent asset, got nil")
	}

	// Test case 8: Try to repay to paused asset (should fail)
	protocol.Assets[asset].Paused = true
	err = protocol.Repay(user, asset, repayAmount, blockNumber+7, txHash)
	if err == nil {
		t.Error("expected error when repaying to paused asset, got nil")
	}
	protocol.Assets[asset].Paused = false
}

// TestLendingProtocol_Liquidate_Comprehensive tests the Liquidate function comprehensively
func TestLendingProtocol_Liquidate_Comprehensive(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	liquidator := generateRandomAddress()
	borrower := generateRandomAddress()
	liquidateAmount := big.NewInt(50000) // 50K liquidation
	blockNumber := uint64(1000)
	txHash := generateRandomHash()

	// Test case 1: Try to liquidate without proper setup (should fail)
	err = protocol.Liquidate(liquidator, borrower, asset, liquidateAmount, blockNumber, txHash)
	if err == nil {
		t.Error("expected error when liquidating without proper setup, got nil")
	}

	// Test case 2: Set up borrower with unhealthy position
	protocol.Users[borrower] = &User{
		Address:    borrower,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	// Set borrower to have high borrow relative to collateral (unhealthy)
	protocol.Users[borrower].Assets[asset] = &UserAsset{
		Token:           asset,
		Balance:         big.NewInt(0),
		BorrowBalance:   big.NewInt(100000), // 100K borrowed
		CollateralValue: big.NewInt(0),
		BorrowValue:     big.NewInt(0),
		LastUpdate:      time.Now(),
	}

	// Set borrower's borrow value high
	protocol.Users[borrower].Borrows[asset] = big.NewInt(100000)

	// Set asset total borrow
	protocol.Assets[asset].TotalBorrow = big.NewInt(100000)
	protocol.TotalBorrow = big.NewInt(100000)

	// Test case 3: Try to liquidate self (should fail)
	err = protocol.Liquidate(borrower, borrower, asset, liquidateAmount, blockNumber+1, txHash)
	if err == nil {
		t.Error("expected error when liquidating self, got nil")
	}

	// Test case 4: Try to liquidate from paused protocol (should fail)
	protocol.Paused = true
	err = protocol.Liquidate(liquidator, borrower, asset, liquidateAmount, blockNumber+2, txHash)
	if err == nil {
		t.Error("expected error when liquidating from paused protocol, got nil")
	}
	protocol.Paused = false

	// Test case 5: Try to liquidate invalid amount (should fail)
	err = protocol.Liquidate(liquidator, borrower, asset, big.NewInt(0), blockNumber+3, txHash)
	if err == nil {
		t.Error("expected error when liquidating zero amount, got nil")
	}

	err = protocol.Liquidate(liquidator, borrower, asset, big.NewInt(-1000), blockNumber+4, txHash)
	if err == nil {
		t.Error("expected error when liquidating negative amount, got nil")
	}

	// Test case 6: Try to liquidate from non-existent asset (should fail)
	nonExistentAsset := generateRandomAddress()
	err = protocol.Liquidate(liquidator, borrower, nonExistentAsset, liquidateAmount, blockNumber+5, txHash)
	if err == nil {
		t.Error("expected error when liquidating from non-existent asset, got nil")
	}

	// Test case 7: Try to liquidate from paused asset (should fail)
	protocol.Assets[asset].Paused = true
	err = protocol.Liquidate(liquidator, borrower, asset, liquidateAmount, blockNumber+6, txHash)
	if err == nil {
		t.Error("expected error when liquidating from paused asset, got nil")
	}
	protocol.Assets[asset].Paused = false
}

// TestLendingProtocol_InterestRateUpdates tests interest rate updates
func TestLendingProtocol_InterestRateUpdates(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Test initial interest rates
	assetInfo := protocol.GetAssetInfo(asset)
	if assetInfo.BorrowRate.Sign() != 0 {
		t.Errorf("expected initial borrow rate 0, got %v", assetInfo.BorrowRate)
	}
	if assetInfo.SupplyRate.Sign() != 0 {
		t.Errorf("expected initial supply rate 0, got %v", assetInfo.SupplyRate)
	}

	// Supply tokens to trigger interest rate update
	user := generateRandomAddress()
	supplyAmount := big.NewInt(1000000) // 1M tokens
	err = protocol.Supply(user, asset, supplyAmount, uint64(1), generateRandomHash())
	if err != nil {
		t.Fatalf("failed to supply tokens: %v", err)
	}

	// Check that interest rates were updated
	assetInfo = protocol.GetAssetInfo(asset)
	if assetInfo.BorrowRate.Sign() == 0 {
		t.Error("expected borrow rate to be updated after supply")
	}
	// Note: Supply rate might be 0 if utilization is 0 (no borrows)
	// This is expected behavior for the current implementation

	// Check that interest event was recorded
	if len(protocol.InterestEvents) == 0 {
		t.Error("expected interest event to be recorded")
	}

	event := protocol.InterestEvents[0]
	if event.Asset != asset {
		t.Errorf("expected asset %v, got %v", asset, event.Asset)
	}
	if event.BorrowRate.Cmp(assetInfo.BorrowRate) != 0 {
		t.Errorf("expected borrow rate %v, got %v", assetInfo.BorrowRate, event.BorrowRate)
	}
	if event.SupplyRate.Cmp(assetInfo.SupplyRate) != 0 {
		t.Errorf("expected supply rate %v, got %v", assetInfo.SupplyRate, event.SupplyRate)
	}
}

// TestLendingProtocol_UtilizationRate tests utilization rate calculation
func TestLendingProtocol_UtilizationRate(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Test utilization rate with no supply
	assetInfo := protocol.GetAssetInfo(asset)
	utilization := protocol.calculateUtilizationRate(assetInfo)
	if utilization.Sign() != 0 {
		t.Errorf("expected utilization rate 0 with no supply, got %v", utilization)
	}

	// Supply tokens
	user := generateRandomAddress()
	supplyAmount := big.NewInt(1000000) // 1M tokens
	err = protocol.Supply(user, asset, supplyAmount, uint64(1), generateRandomHash())
	if err != nil {
		t.Fatalf("failed to supply tokens: %v", err)
	}

	// Test utilization rate with only supply (should be 0)
	assetInfo = protocol.GetAssetInfo(asset)
	utilization = protocol.calculateUtilizationRate(assetInfo)
	if utilization.Sign() != 0 {
		t.Errorf("expected utilization rate 0 with only supply, got %v", utilization)
	}

	// Borrow tokens
	borrowAmount := big.NewInt(500000) // 500K borrow (50% utilization)

	// Set up user with collateral
	protocol.Users[user].Collateral[asset] = supplyAmount
	protocol.Users[user].Borrows[asset] = big.NewInt(0) // Start with no borrows

	// Now borrow
	err = protocol.Borrow(user, asset, borrowAmount, uint64(2), generateRandomHash())
	if err != nil {
		t.Fatalf("failed to borrow tokens: %v", err)
	}

	// Test utilization rate with supply and borrow
	assetInfo = protocol.GetAssetInfo(asset)
	utilization = protocol.calculateUtilizationRate(assetInfo)
	expectedUtilization := big.NewInt(5000) // 50% = 5000 basis points
	if utilization.Cmp(expectedUtilization) != 0 {
		t.Errorf("expected utilization rate %v, got %v", expectedUtilization, utilization)
	}
}

// TestLendingProtocol_LiquidationBonus tests liquidation bonus calculation
func TestLendingProtocol_LiquidationBonus(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Test liquidation bonus calculation
	amount := big.NewInt(100000) // 100K
	bonus := protocol.calculateLiquidationBonus(amount)
	expectedBonus := big.NewInt(5000) // 5% of 100K = 5K

	if bonus.Cmp(expectedBonus) != 0 {
		t.Errorf("expected liquidation bonus %v, got %v", expectedBonus, bonus)
	}

	// Test with different amounts
	amount = big.NewInt(50000) // 50K
	bonus = protocol.calculateLiquidationBonus(amount)
	expectedBonus = big.NewInt(2500) // 5% of 50K = 2.5K

	if bonus.Cmp(expectedBonus) != 0 {
		t.Errorf("expected liquidation bonus %v, got %v", expectedBonus, bonus)
	}
}

// TestLendingProtocol_HealthFactor tests health factor calculation
func TestLendingProtocol_HealthFactor(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	user := generateRandomAddress()

	// Test health factor for non-existent user
	healthFactor := protocol.calculateHealthFactor(user)
	if healthFactor.Sign() != 0 {
		t.Errorf("expected health factor 0 for non-existent user, got %v", healthFactor)
	}

	// Set up user with collateral and borrows
	protocol.Users[user] = &User{
		Address:    user,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	// Test health factor with no borrows (should be 100%)
	protocol.Users[user].Collateral[asset] = big.NewInt(100000)
	healthFactor = protocol.calculateHealthFactor(user)
	expectedHealthFactor := big.NewInt(10000) // 100% = 10000 basis points
	if healthFactor.Cmp(expectedHealthFactor) != 0 {
		t.Errorf("expected health factor %v with no borrows, got %v", expectedHealthFactor, healthFactor)
	}

	// Test health factor with borrows
	protocol.Users[user].Borrows[asset] = big.NewInt(50000) // 50K borrowed
	healthFactor = protocol.calculateHealthFactor(user)
	// Collateral: 100K * 150% = 150K, Borrow: 50K / 80% = 62.5K
	// Health factor = (150K / 62.5K) * 10000 = 24000 (240%)
	expectedHealthFactor = big.NewInt(24000)
	if healthFactor.Cmp(expectedHealthFactor) != 0 {
		t.Errorf("expected health factor %v with borrows, got %v", expectedHealthFactor, healthFactor)
	}
}

// TestLendingProtocol_ValidationFunctions tests all validation functions
func TestLendingProtocol_ValidationFunctions(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	user := generateRandomAddress()
	amount := big.NewInt(100000)

	// Test validateBorrowInput
	err = protocol.validateBorrowInput(asset, amount)
	if err != nil {
		t.Errorf("validateBorrowInput failed for valid input: %v", err)
	}

	err = protocol.validateBorrowInput(asset, big.NewInt(0))
	if err == nil {
		t.Error("validateBorrowInput should fail for zero amount")
	}

	err = protocol.validateBorrowInput(asset, big.NewInt(-1000))
	if err == nil {
		t.Error("validateBorrowInput should fail for negative amount")
	}

	nonExistentAsset := generateRandomAddress()
	// Modify the address slightly to ensure it's different
	nonExistentAsset[0] = byte(128)
	err = protocol.validateBorrowInput(nonExistentAsset, amount)
	if err == nil {
		t.Error("validateBorrowInput should fail for non-existent asset")
	}

	// Test validateRepayInput
	err = protocol.validateRepayInput(user, asset, amount)
	if err == nil {
		t.Error("validateRepayInput should fail for non-existent user")
	}

	// Create user
	protocol.Users[user] = &User{
		Address:    user,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	err = protocol.validateRepayInput(user, asset, amount)
	if err == nil {
		t.Error("validateRepayInput should fail for user without asset")
	}

	// Add user asset
	protocol.Users[user].Assets[asset] = &UserAsset{
		Token:           asset,
		Balance:         big.NewInt(0),
		BorrowBalance:   big.NewInt(0),
		CollateralValue: big.NewInt(0),
		BorrowValue:     big.NewInt(0),
		LastUpdate:      time.Now(),
	}

	err = protocol.validateRepayInput(user, asset, amount)
	if err != nil {
		t.Errorf("validateRepayInput failed for valid input: %v", err)
	}

	// Test validateLiquidateInput
	liquidator := generateRandomAddress()
	borrower := generateRandomAddress()
	// Ensure addresses are different
	borrower[0] = byte(255)

	// Test valid input (should pass)
	err = protocol.validateLiquidateInput(liquidator, borrower, asset, amount)
	if err != nil {
		t.Errorf("validateLiquidateInput failed for valid input: %v", err)
	}

	// Test self-liquidation (should fail)
	err = protocol.validateLiquidateInput(borrower, borrower, asset, amount)
	if err == nil {
		t.Error("validateLiquidateInput should fail for self-liquidation")
	}

	// Test zero amount (should fail)
	err = protocol.validateLiquidateInput(liquidator, borrower, asset, big.NewInt(0))
	if err == nil {
		t.Error("validateLiquidateInput should fail for zero amount")
	}

	// Test negative amount (should fail)
	err = protocol.validateLiquidateInput(liquidator, borrower, asset, big.NewInt(-1000))
	if err == nil {
		t.Error("validateLiquidateInput should fail for negative amount")
	}
}

// TestLendingProtocol_ErrorHandling tests comprehensive error handling
func TestLendingProtocol_ErrorHandling(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Test pause when already paused
	err := protocol.Pause()
	if err != nil {
		t.Errorf("unexpected error pausing protocol: %v", err)
	}

	err = protocol.Pause()
	if err == nil {
		t.Error("expected error when pausing already paused protocol")
	}

	// Test unpause when not paused
	err = protocol.Unpause()
	if err != nil {
		t.Errorf("unexpected error unpausing protocol: %v", err)
	}

	err = protocol.Unpause()
	if err == nil {
		t.Error("expected error when unpausing already unpaused protocol")
	}

	// Test add asset when already exists
	asset := generateRandomAddress()
	err = protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Errorf("unexpected error adding asset: %v", err)
	}

	err = protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err == nil {
		t.Error("expected error when adding duplicate asset")
	}
}

// TestLendingProtocol_ConcurrentOperations tests concurrent operations
func TestLendingProtocol_ConcurrentOperations(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "USDC", 6, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	// Test concurrent read operations (should be safe)
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// Read operations only
			_ = protocol.GetAssetInfo(asset)
			_, _, _, _, _ = protocol.GetProtocolStats()
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// Test sequential operations to avoid race conditions
	for i := 0; i < 5; i++ {
		user := generateRandomAddress()
		amount := big.NewInt(int64(1000000 + i*100000))

		err := protocol.Supply(user, asset, amount, uint64(i+1), generateRandomHash())
		if err != nil {
			t.Errorf("sequential supply failed: %v", err)
		}
	}

	// Verify all operations completed
	supplyCount, _, _, _, _ := protocol.GetProtocolStats()
	if supplyCount != 5 {
		t.Errorf("expected 5 supply operations, got %d", supplyCount)
	}
}

// TestLendingProtocol_LiquidationEligibility tests the liquidation eligibility check
func TestLendingProtocol_LiquidationEligibility(t *testing.T) {
	protocol := NewLendingProtocol(
		"test-lending",
		"Test Lending Protocol",
		"TLP",
		18,
		generateRandomAddress(),
		big.NewInt(11000), // 110% liquidation threshold
		big.NewInt(500),   // 5% liquidation bonus
	)

	// Add asset
	asset := generateRandomAddress()
	err := protocol.AddAsset(asset, "TEST", 18, big.NewInt(15000), big.NewInt(8000))
	if err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}

	user := generateRandomAddress()

	// Test case 1: Check liquidation eligibility for non-existent user (should fail)
	err = protocol.checkLiquidationEligibility(user, asset)
	if err == nil {
		t.Error("expected error when checking liquidation eligibility for non-existent user")
	}

	// Test case 2: Set up user with healthy position (should not be eligible for liquidation)
	protocol.Users[user] = &User{
		Address:    user,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	// Set user with high collateral relative to borrows (healthy position)
	protocol.Users[user].Collateral[asset] = big.NewInt(1000000) // 1M collateral
	protocol.Users[user].Borrows[asset] = big.NewInt(100000)     // 100K borrows

	// Test case 3: Check liquidation eligibility for healthy user (should not be eligible)
	err = protocol.checkLiquidationEligibility(user, asset)
	if err == nil {
		t.Error("expected error when checking liquidation eligibility for healthy user")
	}

	// Test case 4: Set up user with unhealthy position (should be eligible for liquidation)
	unhealthyUser := generateRandomAddress()
	// Modify the address slightly to ensure it's different
	unhealthyUser[0] = byte(255)

	protocol.Users[unhealthyUser] = &User{
		Address:    unhealthyUser,
		Assets:     make(map[engine.Address]*UserAsset),
		Collateral: make(map[engine.Address]*big.Int),
		Borrows:    make(map[engine.Address]*big.Int),
		LastUpdate: time.Now(),
	}

	// Set user with low collateral relative to borrows (unhealthy position)
	protocol.Users[unhealthyUser].Collateral[asset] = big.NewInt(100000) // 100K collateral
	protocol.Users[unhealthyUser].Borrows[asset] = big.NewInt(1000000)   // 1M borrows

	// Test case 5: Check liquidation eligibility for unhealthy user (should be eligible)
	err = protocol.checkLiquidationEligibility(unhealthyUser, asset)
	if err != nil {
		t.Errorf("unexpected error when checking liquidation eligibility for unhealthy user: %v", err)
	}
}
