package atomic_swaps

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// AtomicSwap represents a cross-chain atomic swap
type AtomicSwap struct {
	ID           string
	ChainA       string
	ChainB       string
	AssetA       string
	AssetB       string
	AmountA      *big.Int
	AmountB      *big.Int
	ParticipantA [20]byte
	ParticipantB [20]byte
	SecretHash   [32]byte
	Secret       []byte
	TimelockA    time.Time
	TimelockB    time.Time
	Status       SwapStatus
	CreatedAt    time.Time
	CompletedAt  time.Time
	mu           sync.RWMutex
	config       SwapConfig
	metrics      SwapMetrics
}

// SwapStatus represents the current status of an atomic swap
type SwapStatus int

const (
	SwapStatusInitiated SwapStatus = iota
	SwapStatusFundedA
	SwapStatusFundedB
	SwapStatusCompleted
	SwapStatusExpired
	SwapStatusRefunded
)

// SwapConfig holds configuration for atomic swaps
type SwapConfig struct {
	MinAmount         *big.Int
	MaxAmount         *big.Int
	DefaultTimelock   time.Duration
	EnableCompression bool
	SecurityLevel     SecurityLevel
	AutoRefund        bool
}

// SecurityLevel defines the security level for atomic swaps
type SecurityLevel int

const (
	SecurityLevelLow SecurityLevel = iota
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelUltra
)

// HTLC represents a Hash Time-Locked Contract
type HTLC struct {
	ID         string
	Chain      string
	Asset      string
	Amount     *big.Int
	Recipient  [20]byte
	SecretHash [32]byte
	Timelock   time.Time
	IsFunded   bool
	IsRedeemed bool
	IsRefunded bool
	CreatedAt  time.Time
	FundedAt   time.Time
	RedeemedAt time.Time
	RefundedAt time.Time
}

// SwapMetrics tracks swap performance metrics
type SwapMetrics struct {
	TotalSwaps      uint64
	SuccessfulSwaps uint64
	FailedSwaps     uint64
	ExpiredSwaps    uint64
	AverageSwapTime time.Duration
	TotalVolume     *big.Int
	LastUpdate      time.Time
}

// NewAtomicSwap creates a new atomic swap instance
func NewAtomicSwap(chainA, chainB, assetA, assetB string, amountA, amountB *big.Int, participantA, participantB [20]byte, config SwapConfig) *AtomicSwap {
	// Set default values if not provided
	if config.MinAmount == nil {
		config.MinAmount = big.NewInt(1000000000000000000) // 1 ETH
	}
	if config.MaxAmount == nil {
		config.MaxAmount = big.NewInt(0).Mul(big.NewInt(1000000000000000000), big.NewInt(100)) // 100 ETH
	}
	if config.DefaultTimelock == 0 {
		config.DefaultTimelock = time.Hour * 24 * 7 // 1 week
	}

	// Validate amounts
	if amountA.Cmp(config.MinAmount) < 0 || amountA.Cmp(config.MaxAmount) > 0 {
		panic(fmt.Sprintf("amount A %s outside valid range [%s, %s]", amountA.String(), config.MinAmount.String(), config.MaxAmount.String()))
	}
	if amountB.Cmp(config.MinAmount) < 0 || amountB.Cmp(config.MaxAmount) > 0 {
		panic(fmt.Sprintf("amount B %s outside valid range [%s, %s]", amountB.String(), config.MinAmount.String(), config.MaxAmount.String()))
	}

	// Generate secret and hash
	secret := generateSecret()
	secretHash := sha256.Sum256(secret)

	// Calculate timelocks
	now := time.Now()
	timelockA := now.Add(config.DefaultTimelock)
	timelockB := now.Add(config.DefaultTimelock + time.Hour) // B has slightly longer timelock

	return &AtomicSwap{
		ID:           generateSwapID(),
		ChainA:       chainA,
		ChainB:       chainB,
		AssetA:       assetA,
		AssetB:       assetB,
		AmountA:      amountA,
		AmountB:      amountB,
		ParticipantA: participantA,
		ParticipantB: participantB,
		SecretHash:   secretHash,
		Secret:       secret,
		TimelockA:    timelockA,
		TimelockB:    timelockB,
		Status:       SwapStatusInitiated,
		CreatedAt:    now,
		config:       config,
		metrics: SwapMetrics{
			TotalVolume: big.NewInt(0),
		},
	}
}

// InitiateSwap initiates the atomic swap process
func (as *AtomicSwap) InitiateSwap() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Debug: Check if swap object is valid
	if as == nil {
		return fmt.Errorf("swap object is nil")
	}

	if as.Status != SwapStatusInitiated {
		return fmt.Errorf("swap %s cannot be initiated, current status: %d", as.ID, as.Status)
	}

	// Create HTLC for chain A (in a real implementation, this would be deployed on chain A)
	_ = &HTLC{
		ID:         fmt.Sprintf("%s_htlc_a", as.ID),
		Chain:      as.ChainA,
		Asset:      as.AssetA,
		Amount:     as.AmountA,
		Recipient:  as.ParticipantB,
		SecretHash: as.SecretHash,
		Timelock:   as.TimelockA,
		CreatedAt:  time.Now(),
	}

	// In a real implementation, you would deploy this HTLC on chain A
	// For now, we'll just mark it as ready
	as.Status = SwapStatusInitiated
	as.updateMetrics()

	return nil
}

// FundSwapA funds the swap on chain A
func (as *AtomicSwap) FundSwapA() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.Status != SwapStatusInitiated {
		return fmt.Errorf("swap %s cannot be funded on chain A, current status: %d", as.ID, as.Status)
	}

	// In a real implementation, you would fund the HTLC on chain A
	// For now, we'll just update the status
	as.Status = SwapStatusFundedA
	as.updateMetrics()

	return nil
}

// FundSwapB funds the swap on chain B
func (as *AtomicSwap) FundSwapB() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.Status != SwapStatusFundedA {
		return fmt.Errorf("swap %s cannot be funded on chain B, current status: %d", as.ID, as.Status)
	}

	// In a real implementation, you would fund the HTLC on chain B
	// For now, we'll just update the status
	as.Status = SwapStatusFundedB
	as.updateMetrics()

	return nil
}

// CompleteSwap completes the atomic swap by revealing the secret
func (as *AtomicSwap) CompleteSwap(secret []byte) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.Status != SwapStatusFundedB {
		return fmt.Errorf("swap %s cannot be completed, current status: %d", as.ID, as.Status)
	}

	// Verify the secret
	secretHash := sha256.Sum256(secret)
	if secretHash != as.SecretHash {
		return fmt.Errorf("invalid secret for swap %s", as.ID)
	}

	// In a real implementation, you would redeem both HTLCs using the secret
	// For now, we'll just update the status
	as.Status = SwapStatusCompleted
	as.CompletedAt = time.Now()
	as.updateMetrics()

	return nil
}

// RefundSwap refunds the swap if it expires
func (as *AtomicSwap) RefundSwap() error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.Status != SwapStatusFundedA && as.Status != SwapStatusFundedB {
		return fmt.Errorf("swap %s cannot be refunded, current status: %d", as.ID, as.Status)
	}

	// Check if timelock has expired
	now := time.Now()
	if now.Before(as.TimelockA) && now.Before(as.TimelockB) {
		return fmt.Errorf("swap %s timelock has not expired yet", as.ID)
	}

	// In a real implementation, you would refund the HTLCs
	// For now, we'll just update the status
	as.Status = SwapStatusRefunded
	as.updateMetrics()

	return nil
}

// GetStatus returns the current status of the swap
func (as *AtomicSwap) GetStatus() SwapStatus {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.Status
}

// GetMetrics returns the swap metrics
func (as *AtomicSwap) GetMetrics() SwapMetrics {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.metrics
}

// GetSecret returns the secret for the swap
func (as *AtomicSwap) GetSecret() []byte {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.Secret
}

// GetSecretHash returns the secret hash for the swap
func (as *AtomicSwap) GetSecretHash() [32]byte {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.SecretHash
}

// IsExpired checks if the swap has expired
func (as *AtomicSwap) IsExpired() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()

	now := time.Now()
	return now.After(as.TimelockA) && now.After(as.TimelockB)
}

// updateMetrics updates swap metrics
func (as *AtomicSwap) updateMetrics() {
	as.metrics.TotalSwaps++
	as.metrics.LastUpdate = time.Now()

	switch as.Status {
	case SwapStatusCompleted:
		as.metrics.SuccessfulSwaps++
		// Calculate total volume for this swap
		swapVolume := new(big.Int).Add(as.AmountA, as.AmountB)
		as.metrics.TotalVolume.Add(as.metrics.TotalVolume, swapVolume)
		if !as.CreatedAt.IsZero() && !as.CompletedAt.IsZero() {
			as.metrics.AverageSwapTime = as.CompletedAt.Sub(as.CreatedAt)
		}
	case SwapStatusExpired, SwapStatusRefunded:
		as.metrics.FailedSwaps++
		// Note: IsExpired check removed to avoid deadlock
		// In a real implementation, this would be handled differently
	}
}

// generateSecret generates a random secret for the swap
func generateSecret() []byte {
	secret := make([]byte, 32)
	rand.Read(secret)
	return secret
}

// generateSwapID generates a unique swap ID
func generateSwapID() string {
	random := make([]byte, 16)
	rand.Read(random)
	hash := sha256.Sum256(random)
	return fmt.Sprintf("atomic_swap_%x", hash[:8])
}

// Mock implementations for testing
type MockHTLCDeployer struct{}

func NewMockHTLCDeployer() *MockHTLCDeployer {
	return &MockHTLCDeployer{}
}

func (m *MockHTLCDeployer) DeployHTLC(chain string, asset string, amount *big.Int, recipient [20]byte, secretHash [32]byte, timelock time.Time) (string, error) {
	return fmt.Sprintf("htlc_%s_%s", chain, hex.EncodeToString(secretHash[:8])), nil
}

func (m *MockHTLCDeployer) FundHTLC(htlcID string, amount *big.Int) error {
	return nil
}

func (m *MockHTLCDeployer) RedeemHTLC(htlcID string, secret []byte) error {
	return nil
}

func (m *MockHTLCDeployer) RefundHTLC(htlcID string) error {
	return nil
}
