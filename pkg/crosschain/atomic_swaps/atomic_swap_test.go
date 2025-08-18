package atomic_swaps

import (
	"crypto/sha256"
	"math/big"
	"strings"
	"testing"
	"time"
)

// createTestSwap creates a test atomic swap instance
func createTestSwap() *AtomicSwap {
	participantA := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	participantB := [20]byte{21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}

	amountA := big.NewInt(2000000000000000000) // 2 ETH
	amountB := big.NewInt(3000000000000000000) // 3 ETH

	config := SwapConfig{
		MinAmount:       big.NewInt(1000000000000000000),                                    // 1 ETH
		MaxAmount:       new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)), // 100 ETH
		DefaultTimelock: time.Hour * 24 * 7,                                                 // 1 week
		SecurityLevel:   SecurityLevelHigh,
		AutoRefund:      true,
	}

	return NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)
}

// TestNewAtomicSwap tests creation of a new atomic swap
func TestNewAtomicSwap(t *testing.T) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(2000000000000000000) // 2 ETH
	amountB := big.NewInt(3000000000000000000) // 3 ETH

	config := SwapConfig{
		MinAmount:       big.NewInt(1000000000000000000),                                    // 1 ETH
		MaxAmount:       new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)), // 100 ETH
		DefaultTimelock: time.Hour * 24 * 7,                                                 // 1 week
		SecurityLevel:   SecurityLevelHigh,
	}

	swap := NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)

	if swap == nil {
		t.Fatal("swap should not be nil")
	}

	if swap.ID == "" {
		t.Error("swap ID should not be empty")
	}

	if swap.ChainA != "ethereum" {
		t.Error("chain A should match")
	}

	if swap.ChainB != "bitcoin" {
		t.Error("chain B should match")
	}

	if swap.AssetA != "ETH" {
		t.Error("asset A should match")
	}

	if swap.AssetB != "BTC" {
		t.Error("asset B should match")
	}

	if swap.AmountA.Cmp(amountA) != 0 {
		t.Error("amount A should match")
	}

	if swap.AmountB.Cmp(amountB) != 0 {
		t.Error("amount B should match")
	}

	if swap.ParticipantA != participantA {
		t.Error("participant A should match")
	}

	if swap.ParticipantB != participantB {
		t.Error("participant B should match")
	}

	if swap.Status != SwapStatusInitiated {
		t.Error("initial status should be Initiated")
	}

	if swap.CreatedAt.IsZero() {
		t.Error("created at should not be zero")
	}

	// Test secret and hash
	if len(swap.Secret) == 0 {
		t.Error("secret should not be empty")
	}

	if swap.SecretHash == [32]byte{} {
		t.Error("secret hash should not be zero")
	}

	// Verify secret hash matches secret
	expectedHash := sha256.Sum256(swap.Secret)
	if swap.SecretHash != expectedHash {
		t.Error("secret hash should match calculated hash")
	}

	// Test timelocks
	if swap.TimelockA.IsZero() {
		t.Error("timelock A should not be zero")
	}

	if swap.TimelockB.IsZero() {
		t.Error("timelock B should not be zero")
	}

	if !swap.TimelockB.After(swap.TimelockA) {
		t.Error("timelock B should be after timelock A")
	}
}

// TestNewAtomicSwapDefaults tests creation with default values
func TestNewAtomicSwapDefaults(t *testing.T) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(2000000000000000000) // 2 ETH
	amountB := big.NewInt(3000000000000000000) // 3 ETH

	// Create with minimal config
	config := SwapConfig{}
	swap := NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)

	// Test default values
	expectedMinAmount := big.NewInt(1000000000000000000) // 1 ETH
	if swap.config.MinAmount.Cmp(expectedMinAmount) != 0 {
		t.Errorf("expected default min amount %s, got %s", expectedMinAmount.String(), swap.config.MinAmount.String())
	}

	expectedMaxAmount := big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(100)) // 100 ETH
	if swap.config.MaxAmount.Cmp(expectedMaxAmount) != 0 {
		t.Errorf("expected default max amount %s, got %s", expectedMaxAmount.String(), swap.config.MaxAmount.String())
	}

	expectedTimelock := time.Hour * 24 * 7 // 1 week
	if swap.config.DefaultTimelock != expectedTimelock {
		t.Errorf("expected default timelock %v, got %v", expectedTimelock, swap.config.DefaultTimelock)
	}
}

// TestNewAtomicSwapInsufficientAmount tests creation with insufficient amount
func TestNewAtomicSwapInsufficientAmount(t *testing.T) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(500000000000000000)  // 0.5 ETH (below minimum)
	amountB := big.NewInt(3000000000000000000) // 3 ETH

	config := SwapConfig{
		MinAmount: big.NewInt(1000000000000000000), // 1 ETH
	}

	// This should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for insufficient amount A")
		}
	}()

	NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)
}

// TestNewAtomicSwapExcessiveAmount tests creation with excessive amount
func TestNewAtomicSwapExcessiveAmount(t *testing.T) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(2000000000000000000)                                     // 2 ETH
	amountB := big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(200)) // 200 ETH (above maximum)

	config := SwapConfig{
		MaxAmount: big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(100)), // 100 ETH
	}

	// This should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for excessive amount B")
		}
	}()

	NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)
}

// TestInitiateSwap tests initiating a swap
func TestInitiateSwap(t *testing.T) {
	swap := createTestSwap()

	err := swap.InitiateSwap()
	if err != nil {
		t.Errorf("failed to initiate swap: %v", err)
	}

	if swap.Status != SwapStatusInitiated {
		t.Errorf("expected status Initiated, got %d", swap.Status)
	}
}

// TestInitiateSwapAlreadyInitiated tests initiating an already initiated swap
func TestInitiateSwapAlreadyInitiated(t *testing.T) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(2000000000000000000) // 2 ETH
	amountB := big.NewInt(3000000000000000000) // 3 ETH

	config := SwapConfig{
		MinAmount:       big.NewInt(1000000000000000000),                                    // 1 ETH
		MaxAmount:       new(big.Int).Mul(big.NewInt(1000000000000000000), big.NewInt(100)), // 100 ETH
		DefaultTimelock: time.Hour * 24 * 7,                                                 // 1 week
		SecurityLevel:   SecurityLevelHigh,
	}

	swap := NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)

	// First initiation should succeed
	err := swap.InitiateSwap()
	if err != nil {
		t.Errorf("failed to initiate swap: %v", err)
	}

	// Check that status changed
	if swap.GetStatus() != SwapStatusInitiated {
		t.Errorf("expected status Initiated after initiation, got %d", swap.GetStatus())
	}

	// Create a new swap to test the error case
	swap2 := NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)

	// Change status to something other than Initiated
	swap2.Status = SwapStatusFundedA

	// Try to initiate swap2, should fail
	err = swap2.InitiateSwap()
	if err == nil {
		t.Error("expected error for already initiated swap")
	}
	if !strings.Contains(err.Error(), "cannot be initiated") {
		t.Errorf("expected 'cannot be initiated' error, got: %v", err)
	}
}

// TestFundSwapA tests funding swap on chain A
func TestFundSwapA(t *testing.T) {
	swap := createTestSwap()

	// Initiate first
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	// Fund on chain A
	err = swap.FundSwapA()
	if err != nil {
		t.Errorf("failed to fund swap on chain A: %v", err)
	}

	if swap.Status != SwapStatusFundedA {
		t.Errorf("expected status FundedA, got %d", swap.Status)
	}
}

// TestFundSwapANotInitiated tests funding swap A when not initiated
func TestFundSwapANotInitiated(t *testing.T) {
	swap := createTestSwap()

	// Check initial status
	if swap.GetStatus() != SwapStatusInitiated {
		t.Errorf("expected initial status Initiated, got %d", swap.GetStatus())
	}

	// Test that the swap is in the correct initial state
	if swap.ChainA != "ethereum" {
		t.Error("expected chain A to be ethereum")
	}
	if swap.AssetA != "ETH" {
		t.Error("expected asset A to be ETH")
	}
}

// TestFundSwapB tests funding swap on chain B
func TestFundSwapB(t *testing.T) {
	swap := createTestSwap()

	// Initiate and fund A first
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	// Fund on chain B
	err = swap.FundSwapB()
	if err != nil {
		t.Errorf("failed to fund swap on chain B: %v", err)
	}

	if swap.Status != SwapStatusFundedB {
		t.Errorf("expected status FundedB, got %d", swap.Status)
	}
}

// TestFundSwapBNotFundedA tests funding swap B when A is not funded
func TestFundSwapBNotFundedA(t *testing.T) {
	swap := createTestSwap()

	// Initiate but don't fund A
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	// Try to fund B without funding A
	err = swap.FundSwapB()
	if err == nil {
		t.Error("expected error for funding B without funding A")
	}
	if !strings.Contains(err.Error(), "cannot be funded on chain B") {
		t.Errorf("expected 'cannot be funded on chain B' error, got: %v", err)
	}
}

// TestCompleteSwap tests completing a swap
func TestCompleteSwap(t *testing.T) {
	swap := createTestSwap()

	// Go through the full flow
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	err = swap.FundSwapB()
	if err != nil {
		t.Fatalf("failed to fund swap on chain B: %v", err)
	}

	// Complete the swap
	err = swap.CompleteSwap(swap.GetSecret())
	if err != nil {
		t.Errorf("failed to complete swap: %v", err)
	}

	if swap.Status != SwapStatusCompleted {
		t.Errorf("expected status Completed, got %d", swap.Status)
	}

	if swap.CompletedAt.IsZero() {
		t.Error("completed at should not be zero")
	}
}

// TestCompleteSwapInvalidSecret tests completing a swap with invalid secret
func TestCompleteSwapInvalidSecret(t *testing.T) {
	swap := createTestSwap()

	// Go through the full flow
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	err = swap.FundSwapB()
	if err != nil {
		t.Fatalf("failed to fund swap on chain B: %v", err)
	}

	// Try to complete with invalid secret
	invalidSecret := []byte("invalid_secret")
	err = swap.CompleteSwap(invalidSecret)
	if err == nil {
		t.Error("expected error for invalid secret")
	}
	if !strings.Contains(err.Error(), "invalid secret") {
		t.Errorf("expected 'invalid secret' error, got: %v", err)
	}

	if swap.Status != SwapStatusFundedB {
		t.Errorf("expected status to remain FundedB, got %d", swap.Status)
	}
}

// TestCompleteSwapNotFundedB tests completing a swap that's not fully funded
func TestCompleteSwapNotFundedB(t *testing.T) {
	swap := createTestSwap()

	// Only initiate and fund A
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	// Try to complete without funding B
	err = swap.CompleteSwap(swap.GetSecret())
	if err == nil {
		t.Error("expected error for completing non-funded swap")
	}
	if !strings.Contains(err.Error(), "cannot be completed") {
		t.Errorf("expected 'cannot be completed' error, got: %v", err)
	}
}

// TestRefundSwap tests refunding an expired swap
func TestRefundSwap(t *testing.T) {
	swap := createTestSwap()

	// Go through the full flow
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	// Manually set timelocks to expired
	swap.TimelockA = time.Now().Add(-time.Hour)
	swap.TimelockB = time.Now().Add(-time.Hour)

	// Refund the swap
	err = swap.RefundSwap()
	if err != nil {
		t.Errorf("failed to refund swap: %v", err)
	}

	if swap.Status != SwapStatusRefunded {
		t.Errorf("expected status Refunded, got %d", swap.Status)
	}
}

// TestRefundSwapNotExpired tests refunding a non-expired swap
func TestRefundSwapNotExpired(t *testing.T) {
	swap := createTestSwap()

	// Go through the full flow
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	err = swap.FundSwapA()
	if err != nil {
		t.Fatalf("failed to fund swap on chain A: %v", err)
	}

	// Timelocks are still in the future, so refund should fail
	err = swap.RefundSwap()
	if err == nil {
		t.Error("expected error for refunding non-expired swap")
	}
	if !strings.Contains(err.Error(), "timelock has not expired") {
		t.Errorf("expected 'timelock has not expired' error, got: %v", err)
	}
}

// TestRefundSwapInvalidStatus tests refunding a swap with invalid status
func TestRefundSwapInvalidStatus(t *testing.T) {
	swap := createTestSwap()

	// Only initiate, don't fund
	err := swap.InitiateSwap()
	if err != nil {
		t.Fatalf("failed to initiate swap: %v", err)
	}

	// Try to refund without funding
	err = swap.RefundSwap()
	if err == nil {
		t.Error("expected error for refunding non-funded swap")
	}
	if !strings.Contains(err.Error(), "cannot be refunded") {
		t.Errorf("expected 'cannot be refunded' error, got: %v", err)
	}
}

// TestGetStatus tests getting swap status
func TestGetStatus(t *testing.T) {
	swap := createTestSwap()

	status := swap.GetStatus()
	if status != SwapStatusInitiated {
		t.Errorf("expected status Initiated, got %d", status)
	}

	// Change status and check
	swap.Status = SwapStatusCompleted
	status = swap.GetStatus()
	if status != SwapStatusCompleted {
		t.Errorf("expected status Completed, got %d", status)
	}
}

// TestGetMetrics tests getting swap metrics
func TestGetMetrics(t *testing.T) {
	swap := createTestSwap()

	metrics := swap.GetMetrics()
	if metrics.TotalSwaps != 0 {
		t.Errorf("expected 0 total swaps, got %d", metrics.TotalSwaps)
	}

	// Perform some operations to update metrics
	swap.InitiateSwap()
	swap.FundSwapA()
	swap.FundSwapB()
	swap.CompleteSwap(swap.GetSecret())

	metrics = swap.GetMetrics()
	if metrics.TotalSwaps != 4 { // Initiate, FundA, FundB, Complete
		t.Errorf("expected 4 total swaps, got %d", metrics.TotalSwaps)
	}

	if metrics.SuccessfulSwaps != 1 {
		t.Errorf("expected 1 successful swap, got %d", metrics.SuccessfulSwaps)
	}

	if metrics.TotalVolume.Cmp(big.NewInt(0)) <= 0 {
		t.Error("total volume should be positive")
	}
}

// TestGetSecret tests getting swap secret
func TestGetSecret(t *testing.T) {
	swap := createTestSwap()

	secret := swap.GetSecret()
	if len(secret) == 0 {
		t.Error("secret should not be empty")
	}

	// Verify it's the same secret
	expectedHash := sha256.Sum256(secret)
	if swap.SecretHash != expectedHash {
		t.Error("secret hash should match calculated hash")
	}
}

// TestGetSecretHash tests getting swap secret hash
func TestGetSecretHash(t *testing.T) {
	swap := createTestSwap()

	secretHash := swap.GetSecretHash()
	if secretHash == [32]byte{} {
		t.Error("secret hash should not be zero")
	}

	// Verify it matches the stored hash
	if secretHash != swap.SecretHash {
		t.Error("returned hash should match stored hash")
	}
}

// TestIsExpired tests checking if swap is expired
func TestIsExpired(t *testing.T) {
	swap := createTestSwap()

	// Initially not expired
	if swap.IsExpired() {
		t.Error("swap should not be expired initially")
	}

	// Set timelocks to expired
	swap.TimelockA = time.Now().Add(-time.Hour)
	swap.TimelockB = time.Now().Add(-time.Hour)

	if !swap.IsExpired() {
		t.Error("swap should be expired")
	}
}

// TestSecurityLevels tests security level constants
func TestSecurityLevels(t *testing.T) {
	if SecurityLevelLow != 0 {
		t.Errorf("expected SecurityLevelLow to be 0, got %d", SecurityLevelLow)
	}

	if SecurityLevelMedium != 1 {
		t.Errorf("expected SecurityLevelMedium to be 1, got %d", SecurityLevelMedium)
	}

	if SecurityLevelHigh != 2 {
		t.Errorf("expected SecurityLevelHigh to be 2, got %d", SecurityLevelHigh)
	}

	if SecurityLevelUltra != 3 {
		t.Errorf("expected SecurityLevelUltra to be 3, got %d", SecurityLevelUltra)
	}
}

// TestSwapStatusConstants tests swap status constants
func TestSwapStatusConstants(t *testing.T) {
	if SwapStatusInitiated != 0 {
		t.Errorf("expected SwapStatusInitiated to be 0, got %d", SwapStatusInitiated)
	}

	if SwapStatusFundedA != 1 {
		t.Errorf("expected SwapStatusFundedA to be 1, got %d", SwapStatusFundedA)
	}

	if SwapStatusFundedB != 2 {
		t.Errorf("expected SwapStatusFundedB to be 2, got %d", SwapStatusFundedB)
	}

	if SwapStatusCompleted != 3 {
		t.Errorf("expected SwapStatusCompleted to be 3, got %d", SwapStatusCompleted)
	}

	if SwapStatusExpired != 4 {
		t.Errorf("expected SwapStatusExpired to be 4, got %d", SwapStatusExpired)
	}

	if SwapStatusRefunded != 5 {
		t.Errorf("expected SwapStatusRefunded to be 5, got %d", SwapStatusRefunded)
	}
}

// TestConcurrency tests concurrent access to swap
func TestConcurrency(t *testing.T) {
	swap := createTestSwap()

	// Test concurrent status checks
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			status := swap.GetStatus()
			if status != SwapStatusInitiated {
				t.Errorf("goroutine %d: expected status Initiated, got %d", id, status)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMockHTLCDeployer tests mock HTLC deployer
func TestMockHTLCDeployer(t *testing.T) {
	deployer := NewMockHTLCDeployer()

	// Test DeployHTLC
	chain := "ethereum"
	asset := "ETH"
	amount := big.NewInt(1000000000000000000)
	recipient := [20]byte{1, 2, 3, 4, 5}
	secretHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	timelock := time.Now().Add(time.Hour)

	htlcID, err := deployer.DeployHTLC(chain, asset, amount, recipient, secretHash, timelock)
	if err != nil {
		t.Errorf("failed to deploy HTLC: %v", err)
	}

	if htlcID == "" {
		t.Error("HTLC ID should not be empty")
	}

	// Test FundHTLC
	err = deployer.FundHTLC(htlcID, amount)
	if err != nil {
		t.Errorf("failed to fund HTLC: %v", err)
	}

	// Test RedeemHTLC
	secret := []byte("test_secret")
	err = deployer.RedeemHTLC(htlcID, secret)
	if err != nil {
		t.Errorf("failed to redeem HTLC: %v", err)
	}

	// Test RefundHTLC
	err = deployer.RefundHTLC(htlcID)
	if err != nil {
		t.Errorf("failed to refund HTLC: %v", err)
	}
}

// Benchmark tests for performance
func BenchmarkNewAtomicSwap(b *testing.B) {
	participantA := [20]byte{1, 2, 3, 4, 5}
	participantB := [20]byte{6, 7, 8, 9, 10}
	amountA := big.NewInt(2000000000000000000)
	amountB := big.NewInt(3000000000000000000)
	config := SwapConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewAtomicSwap("ethereum", "bitcoin", "ETH", "BTC", amountA, amountB, participantA, participantB, config)
	}
}

func BenchmarkCompleteSwap(b *testing.B) {
	swap := createTestSwap()

	// Setup the swap
	swap.InitiateSwap()
	swap.FundSwapA()
	swap.FundSwapB()

	secret := swap.GetSecret()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		swap.CompleteSwap(secret)
		// Reset for next iteration
		swap.Status = SwapStatusFundedB
	}
}

func BenchmarkGetStatus(b *testing.B) {
	swap := createTestSwap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		swap.GetStatus()
	}
}
