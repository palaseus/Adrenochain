package payment_channels

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// PaymentChannel represents a payment channel between two parties
type PaymentChannel struct {
	ID            string
	ParticipantA  [20]byte
	ParticipantB  [20]byte
	BalanceA      *big.Int
	BalanceB      *big.Int
	TotalDeposit  *big.Int
	PaymentNumber uint64
	IsOpen        bool
	IsClosed      bool
	DisputePeriod time.Duration
	LastUpdate    time.Time
	CreatedAt     time.Time
	ClosedAt      time.Time
	mu            sync.RWMutex
	config        PaymentChannelConfig
	metrics       ChannelMetrics
}

// PaymentChannelConfig holds configuration for the payment channel
type PaymentChannelConfig struct {
	MinDeposit        *big.Int
	MaxDisputeTime    time.Duration
	EnableCompression bool
	SecurityLevel     SecurityLevel
	AutoClose         bool
	MaxPaymentSize    *big.Int
}

// SecurityLevel defines the security level for the payment channel
type SecurityLevel int

const (
	SecurityLevelLow SecurityLevel = iota
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelUltra
)

// Payment represents a payment in the channel
type Payment struct {
	PaymentNumber uint64
	Amount        *big.Int
	Direction     PaymentDirection
	Timestamp     time.Time
	SignatureA    []byte
	SignatureB    []byte
	Nonce         uint64
	Data          []byte
}

// PaymentDirection defines the direction of payment
type PaymentDirection int

const (
	PaymentDirectionAToB PaymentDirection = iota
	PaymentDirectionBToA
)

// Dispute represents a dispute in the payment channel
type Dispute struct {
	ID            string
	PaymentNumber uint64
	Disputer      [20]byte
	Evidence      []byte
	Timestamp     time.Time
	Resolved      bool
	Resolution    DisputeResolution
}

// DisputeResolution represents the resolution of a dispute
type DisputeResolution struct {
	Winner    [20]byte
	Amount    *big.Int
	Reason    string
	Timestamp time.Time
}

// ChannelMetrics tracks channel performance metrics
type ChannelMetrics struct {
	TotalPayments      uint64
	TotalDisputes      uint64
	AveragePaymentTime time.Duration
	DisputeRate        float64
	LastUpdate         time.Time
}

// NewPaymentChannel creates a new payment channel instance
func NewPaymentChannel(participantA, participantB [20]byte, deposit *big.Int, config PaymentChannelConfig) *PaymentChannel {
	// Set default values if not provided
	if config.MinDeposit == nil {
		config.MinDeposit = big.NewInt(1000000000000000000) // 1 ETH
	}
	if config.MaxDisputeTime == 0 {
		config.MaxDisputeTime = time.Hour * 24 * 7 // 1 week
	}
	if config.MaxPaymentSize == nil {
		config.MaxPaymentSize = big.NewInt(100000000000000000) // 0.1 ETH
	}

	// Validate deposit
	if deposit.Cmp(config.MinDeposit) < 0 {
		panic(fmt.Sprintf("deposit %s below minimum requirement %s", deposit.String(), config.MinDeposit.String()))
	}

	// Split deposit equally between participants
	halfDeposit := new(big.Int).Div(deposit, big.NewInt(2))

	return &PaymentChannel{
		ID:            generateChannelID(),
		ParticipantA:  participantA,
		ParticipantB:  participantB,
		BalanceA:      new(big.Int).Set(halfDeposit),
		BalanceB:      new(big.Int).Set(halfDeposit),
		TotalDeposit:  deposit,
		PaymentNumber: 0,
		IsOpen:        true,
		IsClosed:      false,
		DisputePeriod: config.MaxDisputeTime,
		LastUpdate:    time.Now(),
		CreatedAt:     time.Now(),
		config:        config,
		metrics:       ChannelMetrics{},
	}
}

// OpenChannel opens the payment channel
func (pc *PaymentChannel) OpenChannel() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.IsOpen {
		return fmt.Errorf("channel is already open")
	}

	pc.IsOpen = true
	pc.CreatedAt = time.Now()
	pc.LastUpdate = time.Now()

	return nil
}

// CloseChannel closes the payment channel
func (pc *PaymentChannel) CloseChannel() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if !pc.IsOpen {
		return fmt.Errorf("channel is not open")
	}

	if pc.IsClosed {
		return fmt.Errorf("channel is already closed")
	}

	pc.IsClosed = true
	pc.IsOpen = false
	pc.ClosedAt = time.Now()
	pc.LastUpdate = time.Now()

	return nil
}

// MakePayment makes a payment in the channel
func (pc *PaymentChannel) MakePayment(amount *big.Int, direction PaymentDirection, data []byte, signatureA, signatureB []byte) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if !pc.IsOpen {
		return fmt.Errorf("channel is not open")
	}

	if pc.IsClosed {
		return fmt.Errorf("channel is closed")
	}

	// Validate amount
	if amount.Sign() <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	if amount.Cmp(pc.config.MaxPaymentSize) > 0 {
		return fmt.Errorf("payment amount %s exceeds maximum %s", amount.String(), pc.config.MaxPaymentSize.String())
	}

	// Check if payment is possible
	if direction == PaymentDirectionAToB {
		if pc.BalanceA.Cmp(amount) < 0 {
			return fmt.Errorf("insufficient balance for payment: %s, available: %s", amount.String(), pc.BalanceA.String())
		}
	} else {
		if pc.BalanceB.Cmp(amount) < 0 {
			return fmt.Errorf("insufficient balance for payment: %s, available: %s", amount.String(), pc.BalanceB.String())
		}
	}

	// Validate signatures
	if err := pc.validatePaymentSignatures(amount, direction, data, signatureA, signatureB); err != nil {
		return fmt.Errorf("signature validation failed: %w", err)
	}

	// Execute payment
	if direction == PaymentDirectionAToB {
		pc.BalanceA.Sub(pc.BalanceA, amount)
		pc.BalanceB.Add(pc.BalanceB, amount)
	} else {
		pc.BalanceB.Sub(pc.BalanceB, amount)
		pc.BalanceA.Add(pc.BalanceA, amount)
	}

	// Update payment number and timestamp
	pc.PaymentNumber++
	pc.LastUpdate = time.Now()

	// Update metrics
	pc.updateMetrics()

	return nil
}

// CreateDispute creates a dispute in the channel
func (pc *PaymentChannel) CreateDispute(disputer [20]byte, evidence []byte) (*Dispute, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if !pc.IsOpen {
		return nil, fmt.Errorf("channel is not open")
	}

	if pc.IsClosed {
		return nil, fmt.Errorf("channel is closed")
	}

	// Validate disputer
	if disputer != pc.ParticipantA && disputer != pc.ParticipantB {
		return nil, fmt.Errorf("disputer must be a channel participant")
	}

	// Create dispute
	dispute := &Dispute{
		ID:            fmt.Sprintf("dispute_%s_%d", pc.ID, time.Now().Unix()),
		PaymentNumber: pc.PaymentNumber,
		Disputer:      disputer,
		Evidence:      evidence,
		Timestamp:     time.Now(),
		Resolved:      false,
	}

	// Update metrics
	pc.metrics.TotalDisputes++
	pc.metrics.LastUpdate = time.Now()

	// Update dispute rate
	if pc.metrics.TotalPayments > 0 {
		pc.metrics.DisputeRate = float64(pc.metrics.TotalDisputes) / float64(pc.metrics.TotalPayments)
	}

	return dispute, nil
}

// ResolveDispute resolves a dispute
func (pc *PaymentChannel) ResolveDispute(disputeID string, winner [20]byte, amount *big.Int, reason string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Find the dispute
	var targetDispute *Dispute
	// This is a simplified implementation - in practice, disputes would be stored separately
	// For now, we'll just return an error as disputes are not fully implemented

	if targetDispute == nil {
		return fmt.Errorf("dispute %s not found", disputeID)
	}

	if targetDispute.Resolved {
		return fmt.Errorf("dispute %s already resolved", disputeID)
	}

	// Validate winner
	if winner != pc.ParticipantA && winner != pc.ParticipantB {
		return fmt.Errorf("winner must be a channel participant")
	}

	// Resolve dispute
	targetDispute.Resolved = true
	targetDispute.Resolution = DisputeResolution{
		Winner:    winner,
		Amount:    amount,
		Reason:    reason,
		Timestamp: time.Now(),
	}

	// Update metrics
	pc.updateMetrics()

	return nil
}

// GetState returns the current state of the channel
func (pc *PaymentChannel) GetState() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return map[string]interface{}{
		"id":            pc.ID,
		"participantA":  fmt.Sprintf("%x", pc.ParticipantA),
		"participantB":  fmt.Sprintf("%x", pc.ParticipantB),
		"balanceA":      pc.BalanceA.String(),
		"balanceB":      pc.BalanceB.String(),
		"totalDeposit":  pc.TotalDeposit.String(),
		"paymentNumber": pc.PaymentNumber,
		"isOpen":        pc.IsOpen,
		"isClosed":      pc.IsClosed,
		"lastUpdate":    pc.LastUpdate,
		"metrics":       pc.metrics,
		"config":        pc.config,
	}
}

// GetMetrics returns the channel metrics
func (pc *PaymentChannel) GetMetrics() ChannelMetrics {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.metrics
}

// GetBalance returns the balance for a specific participant
func (pc *PaymentChannel) GetBalance(participant [20]byte) (*big.Int, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if participant == pc.ParticipantA {
		return pc.BalanceA, nil
	} else if participant == pc.ParticipantB {
		return pc.BalanceB, nil
	}

	return nil, fmt.Errorf("participant not found")
}

// validatePaymentSignatures validates the signatures for a payment
func (pc *PaymentChannel) validatePaymentSignatures(amount *big.Int, direction PaymentDirection, data []byte, signatureA, signatureB []byte) error {
	// Create message hash
	message := pc.createPaymentMessageHash(amount, direction, data)

	// Validate signature A
	if err := pc.verifySignature(message, signatureA, pc.ParticipantA); err != nil {
		return fmt.Errorf("participant A signature invalid: %w", err)
	}

	// Validate signature B
	if err := pc.verifySignature(message, signatureB, pc.ParticipantB); err != nil {
		return fmt.Errorf("participant B signature invalid: %w", err)
	}

	return nil
}

// createPaymentMessageHash creates a hash of the payment message
func (pc *PaymentChannel) createPaymentMessageHash(amount *big.Int, direction PaymentDirection, data []byte) [32]byte {
	message := make([]byte, 0)

	// Add channel ID
	message = append(message, []byte(pc.ID)...)

	// Add payment number
	paymentNumberBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(paymentNumberBytes, pc.PaymentNumber+1)
	message = append(message, paymentNumberBytes...)

	// Add amount
	message = append(message, amount.Bytes()...)

	// Add direction
	directionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(directionBytes, uint64(direction))
	message = append(message, directionBytes...)

	// Add data
	message = append(message, data...)

	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(time.Now().Unix()))
	message = append(message, timestampBytes...)

	hash := sha256.Sum256(message)
	return hash
}

// verifySignature verifies a signature
func (pc *PaymentChannel) verifySignature(message [32]byte, signature []byte, participant [20]byte) error {
	// This is a simplified signature verification
	// In practice, you would use proper cryptographic verification

	if len(signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	// For now, just check that signature is not empty
	// In a real implementation, you would verify the actual signature

	return nil
}

// updateMetrics updates channel metrics
func (pc *PaymentChannel) updateMetrics() {
	pc.metrics.TotalPayments++
	pc.metrics.LastUpdate = time.Now()

	// Calculate dispute rate
	if pc.metrics.TotalPayments > 0 {
		pc.metrics.DisputeRate = float64(pc.metrics.TotalDisputes) / float64(pc.metrics.TotalPayments)
	}
}

// generateChannelID generates a unique channel ID
func generateChannelID() string {
	random := make([]byte, 16)
	rand.Read(random)
	hash := sha256.Sum256(random)
	return fmt.Sprintf("payment_channel_%x", hash[:8])
}

// Mock implementations for testing
type MockSignatureVerifier struct{}

func NewMockSignatureVerifier() *MockSignatureVerifier {
	return &MockSignatureVerifier{}
}

func (m *MockSignatureVerifier) VerifySignature(message [32]byte, signature []byte, publicKey []byte) bool {
	return true
}

func (m *MockSignatureVerifier) GenerateSignature(message [32]byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	return []byte("mock_signature"), nil
}
