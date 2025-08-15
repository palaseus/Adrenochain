package amm

import (
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// generateRandomAddress generates a random address for testing
func generateRandomAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 1)
	}
	return addr
}

// generateDifferentAddress generates a different address for testing
func generateDifferentAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 200) // Use much different range
	}
	return addr
}

// generateThirdAddress generates a third different address for testing
func generateThirdAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 300) // Use third different range
	}
	return addr
}

// generateRandomHash generates a random hash for testing
func generateRandomHash() engine.Hash {
	hash := engine.Hash{}
	for i := 0; i < len(hash); i++ {
		hash[i] = byte(i + 1)
	}
	return hash
}

// TestNewAMM tests AMM creation
func TestNewAMM(t *testing.T) {
	poolID := "BTC-ETH-POOL"
	tokenA := generateRandomAddress()
	tokenB := generateRandomAddress()
	name := "BTC-ETH Liquidity Pool"
	symbol := "BTC-ETH-LP"
	decimals := uint8(18)
	owner := generateRandomAddress()
	fee := big.NewInt(30) // 0.3%

	amm := NewAMM(poolID, tokenA, tokenB, name, symbol, decimals, owner, fee)

	if amm == nil {
		t.Fatal("expected non-nil AMM")
	}

	if amm.PoolID != poolID {
		t.Errorf("expected PoolID %s, got %s", poolID, amm.PoolID)
	}

	if amm.TokenA != tokenA {
		t.Errorf("expected TokenA %v, got %v", tokenA, amm.TokenA)
	}

	if amm.TokenB != tokenB {
		t.Errorf("expected TokenB %v, got %v", tokenB, amm.TokenB)
	}

	if amm.Name != name {
		t.Errorf("expected Name %s, got %s", name, amm.Name)
	}

	if amm.Symbol != symbol {
		t.Errorf("expected Symbol %s, got %s", symbol, amm.Symbol)
	}

	if amm.Decimals != decimals {
		t.Errorf("expected Decimals %d, got %d", decimals, amm.Decimals)
	}

	if amm.Owner != owner {
		t.Errorf("expected Owner %v, got %v", owner, amm.Owner)
	}

	if amm.Fee.Cmp(fee) != 0 {
		t.Errorf("expected Fee %s, got %s", fee.String(), amm.Fee.String())
	}

	if amm.ReserveA.Sign() != 0 {
		t.Errorf("expected initial ReserveA 0, got %s", amm.ReserveA.String())
	}

	if amm.ReserveB.Sign() != 0 {
		t.Errorf("expected initial ReserveB 0, got %s", amm.ReserveB.String())
	}

	if amm.TotalSupply.Sign() != 0 {
		t.Errorf("expected initial TotalSupply 0, got %s", amm.TotalSupply.String())
	}
}

// TestAMMAddLiquidity tests adding liquidity to the pool
func TestAMMAddLiquidity(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	provider := generateRandomAddress()
	amountA := big.NewInt(1000)
	amountB := big.NewInt(2000)
	minLPTokens := big.NewInt(100)
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Test adding liquidity
	lpTokens, err := amm.AddLiquidity(provider, amountA, amountB, minLPTokens, blockNumber, txHash)
	if err != nil {
		t.Errorf("unexpected error adding liquidity: %v", err)
	}

	if lpTokens == nil {
		t.Fatal("expected non-nil LP tokens")
	}

	if lpTokens.Sign() <= 0 {
		t.Errorf("expected positive LP tokens, got %s", lpTokens.String())
	}

	// Check reserves were updated
	reserveA, reserveB := amm.GetReserves()
	if reserveA.Cmp(amountA) != 0 {
		t.Errorf("expected ReserveA %s, got %s", amountA.String(), reserveA.String())
	}

	if reserveB.Cmp(amountB) != 0 {
		t.Errorf("expected ReserveB %s, got %s", amountB.String(), reserveB.String())
	}

	// Check total supply was updated
	_, _, _, _, _, totalSupply := amm.GetPoolInfo()
	if totalSupply.Cmp(lpTokens) != 0 {
		t.Errorf("expected TotalSupply %s, got %s", lpTokens.String(), totalSupply.String())
	}

	// Check provider balance was updated
	providerBalance := amm.GetLiquidityProviderBalance(provider)
	if providerBalance.Cmp(lpTokens) != 0 {
		t.Errorf("expected provider balance %s, got %s", lpTokens.String(), providerBalance.String())
	}

	// Check mint events
	mintEvents := amm.GetMintEvents()
	if len(mintEvents) != 1 {
		t.Errorf("expected 1 mint event, got %d", len(mintEvents))
	}

	event := mintEvents[0]
	if event.Provider != provider {
		t.Errorf("expected event provider %v, got %v", provider, event.Provider)
	}

	if event.AmountA.Cmp(amountA) != 0 {
		t.Errorf("expected event AmountA %s, got %s", amountA.String(), event.AmountA.String())
	}

	if event.AmountB.Cmp(amountB) != 0 {
		t.Errorf("expected event AmountB %s, got %s", amountB.String(), event.AmountB.String())
	}

	if event.LPTokens.Cmp(lpTokens) != 0 {
		t.Errorf("expected event LPTokens %s, got %s", lpTokens.String(), event.LPTokens.String())
	}
}

// TestAMMRemoveLiquidity tests removing liquidity from the pool
func TestAMMRemoveLiquidity(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	provider := generateRandomAddress()
	amountA := big.NewInt(1000)
	amountB := big.NewInt(2000)
	minLPTokens := big.NewInt(100)
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Add liquidity first
	lpTokens, err := amm.AddLiquidity(provider, amountA, amountB, minLPTokens, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to add liquidity: %v", err)
	}

	// Test removing liquidity
	removeAmount := big.NewInt(50) // Remove half
	minAmountA := big.NewInt(25)
	minAmountB := big.NewInt(50)

	returnedA, returnedB, err := amm.RemoveLiquidity(provider, removeAmount, minAmountA, minAmountB, blockNumber, txHash)
	if err != nil {
		t.Errorf("unexpected error removing liquidity: %v", err)
	}

	if returnedA == nil || returnedA.Sign() <= 0 {
		t.Errorf("expected positive returnedA, got %s", returnedA.String())
	}

	if returnedB == nil || returnedB.Sign() <= 0 {
		t.Errorf("expected positive returnedB, got %s", returnedB.String())
	}

	// Check reserves were updated
	reserveA, reserveB := amm.GetReserves()
	expectedReserveA := new(big.Int).Sub(amountA, returnedA)
	expectedReserveB := new(big.Int).Sub(amountB, returnedB)

	if reserveA.Cmp(expectedReserveA) != 0 {
		t.Errorf("expected ReserveA %s, got %s", expectedReserveA.String(), reserveA.String())
	}

	if reserveB.Cmp(expectedReserveB) != 0 {
		t.Errorf("expected ReserveB %s, got %s", expectedReserveB.String(), reserveB.String())
	}

	// Check total supply was updated
	_, _, _, _, _, totalSupply := amm.GetPoolInfo()
	expectedTotalSupply := new(big.Int).Sub(lpTokens, removeAmount)

	if totalSupply.Cmp(expectedTotalSupply) != 0 {
		t.Errorf("expected TotalSupply %s, got %s", expectedTotalSupply.String(), totalSupply.String())
	}

	// Check provider balance was updated
	providerBalance := amm.GetLiquidityProviderBalance(provider)
	expectedBalance := new(big.Int).Sub(lpTokens, removeAmount)

	if providerBalance.Cmp(expectedBalance) != 0 {
		t.Errorf("expected provider balance %s, got %s", expectedBalance.String(), providerBalance.String())
	}

	// Check burn events
	burnEvents := amm.GetBurnEvents()
	if len(burnEvents) != 1 {
		t.Errorf("expected 1 burn event, got %d", len(burnEvents))
	}

	event := burnEvents[0]
	if event.Provider != provider {
		t.Errorf("expected event provider %v, got %v", provider, event.Provider)
	}

	if event.LPTokens.Cmp(removeAmount) != 0 {
		t.Errorf("expected event LPTokens %s, got %s", removeAmount.String(), event.LPTokens.String())
	}
}

// TestAMMSwap tests token swapping
func TestAMMSwap(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	provider := generateRandomAddress()
	user := generateRandomAddress()
	amountA := big.NewInt(1000)
	amountB := big.NewInt(2000)
	minLPTokens := big.NewInt(100)
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Add liquidity first
	_, err := amm.AddLiquidity(provider, amountA, amountB, minLPTokens, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to add liquidity: %v", err)
	}

	// Test swap
	swapAmount := big.NewInt(100)
	minAmountOut := big.NewInt(180) // Expect at least 180 tokens out

	amountOut, err := amm.Swap(user, amm.TokenA, swapAmount, minAmountOut, blockNumber, txHash)
	if err != nil {
		t.Errorf("unexpected error swapping: %v", err)
	}

	if amountOut == nil || amountOut.Sign() <= 0 {
		t.Errorf("expected positive amountOut, got %s", amountOut.String())
	}

	if amountOut.Cmp(minAmountOut) < 0 {
		t.Errorf("expected amountOut >= %s, got %s", minAmountOut.String(), amountOut.String())
	}

	// Check reserves were updated
	reserveA, reserveB := amm.GetReserves()
	
	// ReserveA should increase by swap amount (minus fee)
	fee := amm.calculateFee(swapAmount)
	amountInAfterFee := new(big.Int).Sub(swapAmount, fee)
	expectedReserveA := new(big.Int).Add(amountA, amountInAfterFee)
	
	if reserveA.Cmp(expectedReserveA) != 0 {
		t.Errorf("expected ReserveA %s, got %s", expectedReserveA.String(), reserveA.String())
	}

	// ReserveB should decrease by amount out
	expectedReserveB := new(big.Int).Sub(amountB, amountOut)
	
	if reserveB.Cmp(expectedReserveB) != 0 {
		t.Errorf("expected ReserveB %s, got %s", expectedReserveB.String(), reserveB.String())
	}

	// Check swap events
	swapEvents := amm.GetSwapEvents()
	if len(swapEvents) != 1 {
		t.Errorf("expected 1 swap event, got %d", len(swapEvents))
	}

	event := swapEvents[0]
	if event.User != user {
		t.Errorf("expected event user %v, got %v", user, event.User)
	}

	if event.TokenIn != amm.TokenA {
		t.Errorf("expected event TokenIn %v, got %v", amm.TokenA, event.TokenIn)
	}

	if event.TokenOut != amm.TokenB {
		t.Errorf("expected event TokenOut %v, got %v", amm.TokenB, event.TokenOut)
	}

	if event.AmountIn.Cmp(swapAmount) != 0 {
		t.Errorf("expected event AmountIn %s, got %s", swapAmount.String(), event.AmountIn.String())
	}

	if event.AmountOut.Cmp(amountOut) != 0 {
		t.Errorf("expected event AmountOut %s, got %s", amountOut.String(), event.AmountOut.String())
	}

	if event.Fee.Cmp(fee) != 0 {
		t.Errorf("expected event Fee %s, got %s", fee.String(), event.Fee.String())
	}
}

// TestAMMPause tests pool pausing functionality
func TestAMMPause(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))

	// Test pausing
	err := amm.Pause()
	if err != nil {
		t.Errorf("unexpected error pausing: %v", err)
	}

	if !amm.IsPaused() {
		t.Error("expected pool to be paused")
	}

	// Test pausing again (should fail)
	err = amm.Pause()
	if err == nil {
		t.Error("expected error when pausing already paused pool")
	}

	// Test unpausing
	err = amm.Unpause()
	if err != nil {
		t.Errorf("unexpected error unpausing: %v", err)
	}

	if amm.IsPaused() {
		t.Error("expected pool to not be paused")
	}

	// Test unpausing again (should fail)
	err = amm.Unpause()
	if err == nil {
		t.Error("expected error when unpausing already unpaused pool")
	}
}

// TestAMMStatistics tests statistics collection
func TestAMMStatistics(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	provider := generateRandomAddress()
	user := generateRandomAddress()
	amountA := big.NewInt(1000)
	amountB := big.NewInt(2000)
	minLPTokens := big.NewInt(100)
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Add liquidity
	_, err := amm.AddLiquidity(provider, amountA, amountB, minLPTokens, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to add liquidity: %v", err)
	}

	// Perform swap
	swapAmount := big.NewInt(10000) // Use larger amount to ensure fee > 0
	minAmountOut := big.NewInt(1000) // Set reasonable minimum output

	_, err = amm.Swap(user, amm.TokenA, swapAmount, minAmountOut, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to swap: %v", err)
	}

	// Check statistics
	swapCount, volume24h, fees24h, totalFees := amm.GetStatistics()

	if swapCount != 1 {
		t.Errorf("expected SwapCount 1, got %d", swapCount)
	}

	if volume24h.Cmp(swapAmount) != 0 {
		t.Errorf("expected Volume24h %s, got %s", swapAmount.String(), volume24h.String())
	}

	if fees24h.Sign() <= 0 {
		t.Errorf("expected positive Fees24h, got %s", fees24h.String())
	}

	if totalFees.Sign() <= 0 {
		t.Errorf("expected positive TotalFees, got %s", totalFees.String())
	}
}

// TestAMMClone tests AMM cloning functionality
func TestAMMClone(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	provider := generateRandomAddress()
	amountA := big.NewInt(1000)
	amountB := big.NewInt(2000)
	minLPTokens := big.NewInt(100)
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Add some liquidity
	_, err := amm.AddLiquidity(provider, amountA, amountB, minLPTokens, blockNumber, txHash)
	if err != nil {
		t.Fatalf("failed to add liquidity: %v", err)
	}

	// Clone the AMM
	clone := amm.Clone()

	if clone == amm {
		t.Error("expected clone to be different from original")
	}

	// Check that values are copied
	if clone.PoolID != amm.PoolID {
		t.Error("expected clone PoolID to match original")
	}

	if clone.TokenA != amm.TokenA {
		t.Error("expected clone TokenA to match original")
	}

	if clone.TokenB != amm.TokenB {
		t.Error("expected clone TokenB to match original")
	}

	// Get reserves for comparison
	originalReserveA, originalReserveB := amm.GetReserves()
	cloneReserveA, cloneReserveB := clone.GetReserves()

	if cloneReserveA.Cmp(originalReserveA) != 0 {
		t.Error("expected clone ReserveA to match original")
	}

	if cloneReserveB.Cmp(originalReserveB) != 0 {
		t.Error("expected clone ReserveB to match original")
	}

	// Get total supply for comparison
	_, _, _, _, _, originalTotalSupply := amm.GetPoolInfo()
	_, _, _, _, _, cloneTotalSupply := clone.GetPoolInfo()

	if cloneTotalSupply.Cmp(originalTotalSupply) != 0 {
		t.Error("expected clone TotalSupply to match original")
	}

	// Check that provider balance is copied
	originalBalance := amm.GetLiquidityProviderBalance(provider)
	cloneBalance := clone.GetLiquidityProviderBalance(provider)

	if originalBalance.Cmp(cloneBalance) != 0 {
		t.Error("expected clone provider balance to match original")
	}
}

// TestAMMConcurrency tests concurrent access to AMM
func TestAMMConcurrency(t *testing.T) {
	amm := NewAMM("test-pool", generateRandomAddress(), generateRandomAddress(), "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	
	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _, _, _, _, _ = amm.GetPoolInfo()
				_, _ = amm.GetReserves()
				_ = amm.GetFee()
				_ = amm.GetOwner()
				_ = amm.IsPaused()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestAMMInvalidInputs tests invalid input handling
func TestAMMInvalidInputs(t *testing.T) {
	tokenA := generateRandomAddress()
	tokenB := generateDifferentAddress()
	amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST-LP", 18, generateRandomAddress(), big.NewInt(30))
	user := generateRandomAddress()
	blockNumber := uint64(12345)
	txHash := generateRandomHash()

	// Test swap with invalid token
	invalidToken := generateThirdAddress()
	_, err := amm.Swap(user, invalidToken, big.NewInt(100), big.NewInt(0), blockNumber, txHash)
	if err == nil {
		t.Error("expected error for invalid token")
	}

	// Test swap with zero amount
	_, err = amm.Swap(user, amm.TokenA, big.NewInt(0), big.NewInt(0), blockNumber, txHash)
	if err == nil {
		t.Error("expected error for zero amount")
	}

	// Test swap with negative amount
	_, err = amm.Swap(user, amm.TokenA, big.NewInt(-100), big.NewInt(0), blockNumber, txHash)
	if err == nil {
		t.Error("expected error for negative amount")
	}

	// Test adding liquidity with zero amounts
	_, err = amm.AddLiquidity(user, big.NewInt(0), big.NewInt(100), big.NewInt(0), blockNumber, txHash)
	if err == nil {
		t.Error("expected error for zero amountA")
	}

	_, err = amm.AddLiquidity(user, big.NewInt(100), big.NewInt(0), big.NewInt(0), blockNumber, txHash)
	if err == nil {
		t.Error("expected error for zero amountB")
	}
}
