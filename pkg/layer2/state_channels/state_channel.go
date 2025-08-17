package state_channels

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

// StateChannel represents a state channel between two parties
type StateChannel struct {
	ID              string
	ParticipantA    [20]byte
	ParticipantB    [20]byte
	BalanceA        *big.Int
	BalanceB        *big.Int
	TotalDeposit    *big.Int
	StateNumber     uint64
	IsOpen          bool
	IsClosed        bool
	DisputePeriod   time.Duration
	LastUpdate      time.Time
	CreatedAt       time.Time
	ClosedAt        time.Time
	mu              sync.RWMutex
	config          StateChannelConfig
	metrics         ChannelMetrics
}

// StateChannelConfig holds configuration for the state channel
type StateChannelConfig struct {
	MinDeposit      *big.Int
	MaxDisputeTime  time.Duration
	EnableCompression bool
	SecurityLevel   SecurityLevel
	AutoClose       bool
}

// SecurityLevel defines the security level for the state channel
type SecurityLevel int

const (
	SecurityLevelLow SecurityLevel = iota
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelUltra
)

// ChannelState represents a state update in the channel
type ChannelState struct {
	StateNumber  uint64
	BalanceA     *big.Int
	BalanceB     *big.Int
	Timestamp    time.Time
	SignatureA   []byte
	SignatureB   []byte
	Nonce        uint64
	Data         []byte
}

// Dispute represents a dispute in the state channel
type Dispute struct {
	ID            string
	StateNumber   uint64
	Disputer      [20]byte
	Evidence      []byte
	Timestamp     time.Time
	Resolved      bool
	Resolution    DisputeResolution
}

// DisputeResolution represents the resolution of a dispute
type DisputeResolution struct {
	Winner        [20]byte
	Amount        *big.Int
	Reason        string
	Timestamp     time.Time
}

// ChannelMetrics tracks channel performance metrics
type ChannelMetrics struct {
	TotalStates     uint64
	TotalDisputes   uint64
	AverageStateTime time.Duration
	DisputeRate     float64
	LastUpdate      time.Time
}

// NewStateChannel creates a new state channel instance
func NewStateChannel(participantA, participantB [20]byte, deposit *big.Int, config StateChannelConfig) *StateChannel {
	// Set default values if not provided
	if config.MinDeposit == nil {
		config.MinDeposit = big.NewInt(1000000000000000000) // 1 ETH
	}
	if config.MaxDisputeTime == 0 {
		config.MaxDisputeTime = time.Hour * 24 * 7 // 1 week
	}
	
	// Validate deposit
	if deposit.Cmp(config.MinDeposit) < 0 {
		panic(fmt.Sprintf("deposit %s below minimum requirement %s", deposit.String(), config.MinDeposit.String()))
	}
	
	// Split deposit equally between participants
	halfDeposit := new(big.Int).Div(deposit, big.NewInt(2))
	
	return &StateChannel{
		ID:            generateChannelID(),
		ParticipantA:  participantA,
		ParticipantB:  participantB,
		BalanceA:      halfDeposit,
		BalanceB:      halfDeposit,
		TotalDeposit:  deposit,
		StateNumber:   0,
		IsOpen:        true,
		IsClosed:      false,
		DisputePeriod: config.MaxDisputeTime,
		LastUpdate:    time.Now(),
		CreatedAt:     time.Now(),
		config:        config,
		metrics:       ChannelMetrics{},
	}
}

// OpenChannel opens the state channel
func (sc *StateChannel) OpenChannel() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.IsOpen {
		return fmt.Errorf("channel is already open")
	}
	
	sc.IsOpen = true
	sc.CreatedAt = time.Now()
	sc.LastUpdate = time.Now()
	
	return nil
}

// CloseChannel closes the state channel
func (sc *StateChannel) CloseChannel() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if !sc.IsOpen {
		return fmt.Errorf("channel is not open")
	}
	
	if sc.IsClosed {
		return fmt.Errorf("channel is already closed")
	}
	
	sc.IsClosed = true
	sc.IsOpen = false
	sc.ClosedAt = time.Now()
	sc.LastUpdate = time.Now()
	
	return nil
}

// UpdateState updates the channel state with new balances
func (sc *StateChannel) UpdateState(balanceA, balanceB *big.Int, data []byte, signatureA, signatureB []byte) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if !sc.IsOpen {
		return fmt.Errorf("channel is not open")
	}
	
	if sc.IsClosed {
		return fmt.Errorf("channel is closed")
	}
	
	// Validate balances
	if balanceA.Sign() < 0 || balanceB.Sign() < 0 {
		return fmt.Errorf("balances cannot be negative")
	}
	
	totalBalance := new(big.Int).Add(balanceA, balanceB)
	if totalBalance.Cmp(sc.TotalDeposit) != 0 {
		return fmt.Errorf("total balance must equal total deposit")
	}
	
	// Validate signatures
	if err := sc.validateSignatures(balanceA, balanceB, data, signatureA, signatureB); err != nil {
		return fmt.Errorf("signature validation failed: %w", err)
	}
	
	// Update state
	sc.StateNumber++
	sc.BalanceA = balanceA
	sc.BalanceB = balanceB
	sc.LastUpdate = time.Now()
	
	// Update metrics
	sc.updateMetrics()
	
	return nil
}

// CreateDispute creates a dispute in the channel
func (sc *StateChannel) CreateDispute(disputer [20]byte, evidence []byte) (*Dispute, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if !sc.IsOpen {
		return nil, fmt.Errorf("channel is not open")
	}
	
	if sc.IsClosed {
		return nil, fmt.Errorf("channel is closed")
	}
	
	// Validate disputer
	if disputer != sc.ParticipantA && disputer != sc.ParticipantB {
		return nil, fmt.Errorf("disputer must be a channel participant")
	}
	
	// Create dispute
	dispute := &Dispute{
		ID:          fmt.Sprintf("dispute_%s_%d", sc.ID, time.Now().Unix()),
		StateNumber: sc.StateNumber,
		Disputer:    disputer,
		Evidence:    evidence,
		Timestamp:   time.Now(),
		Resolved:    false,
	}
	
	// Update metrics
	sc.metrics.TotalDisputes++
	sc.metrics.LastUpdate = time.Now()
	
	// Update dispute rate
	if sc.metrics.TotalStates > 0 {
		sc.metrics.DisputeRate = float64(sc.metrics.TotalDisputes) / float64(sc.metrics.TotalStates)
	}
	
	return dispute, nil
}

// ResolveDispute resolves a dispute
func (sc *StateChannel) ResolveDispute(disputeID string, winner [20]byte, amount *big.Int, reason string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
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
	if winner != sc.ParticipantA && winner != sc.ParticipantB {
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
	sc.updateMetrics()
	
	return nil
}

// GetState returns the current state of the channel
func (sc *StateChannel) GetState() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	return map[string]interface{}{
		"id":           sc.ID,
		"participantA": fmt.Sprintf("%x", sc.ParticipantA),
		"participantB": fmt.Sprintf("%x", sc.ParticipantB),
		"balanceA":     sc.BalanceA.String(),
		"balanceB":     sc.BalanceB.String(),
		"totalDeposit": sc.TotalDeposit.String(),
		"stateNumber":  sc.StateNumber,
		"isOpen":       sc.IsOpen,
		"isClosed":     sc.IsClosed,
		"lastUpdate":   sc.LastUpdate,
		"metrics":      sc.metrics,
		"config":       sc.config,
	}
}

// GetMetrics returns the channel metrics
func (sc *StateChannel) GetMetrics() ChannelMetrics {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.metrics
}

// GetBalance returns the balance for a specific participant
func (sc *StateChannel) GetBalance(participant [20]byte) (*big.Int, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	if participant == sc.ParticipantA {
		return sc.BalanceA, nil
	} else if participant == sc.ParticipantB {
		return sc.BalanceB, nil
	}
	
	return nil, fmt.Errorf("participant not found")
}

// validateSignatures validates the signatures for a state update
func (sc *StateChannel) validateSignatures(balanceA, balanceB *big.Int, data []byte, signatureA, signatureB []byte) error {
	// Create message hash
	message := sc.createMessageHash(balanceA, balanceB, data)
	
	// Validate signature A
	if err := sc.verifySignature(message, signatureA, sc.ParticipantA); err != nil {
		return fmt.Errorf("participant A signature invalid: %w", err)
	}
	
	// Validate signature B
	if err := sc.verifySignature(message, signatureB, sc.ParticipantB); err != nil {
		return fmt.Errorf("participant B signature invalid: %w", err)
	}
	
	return nil
}

// createMessageHash creates a hash of the state update message
func (sc *StateChannel) createMessageHash(balanceA, balanceB *big.Int, data []byte) [32]byte {
	message := make([]byte, 0)
	
	// Add channel ID
	message = append(message, []byte(sc.ID)...)
	
	// Add state number
	stateNumberBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(stateNumberBytes, sc.StateNumber+1)
	message = append(message, stateNumberBytes...)
	
	// Add balances
	message = append(message, balanceA.Bytes()...)
	message = append(message, balanceB.Bytes()...)
	
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
func (sc *StateChannel) verifySignature(message [32]byte, signature []byte, participant [20]byte) error {
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
func (sc *StateChannel) updateMetrics() {
	sc.metrics.TotalStates++
	sc.metrics.LastUpdate = time.Now()
	
	// Calculate dispute rate
	if sc.metrics.TotalStates > 0 {
		sc.metrics.DisputeRate = float64(sc.metrics.TotalDisputes) / float64(sc.metrics.TotalStates)
	}
}

// generateChannelID generates a unique channel ID
func generateChannelID() string {
	random := make([]byte, 16)
	rand.Read(random)
	hash := sha256.Sum256(random)
	return fmt.Sprintf("channel_%x", hash[:8])
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
