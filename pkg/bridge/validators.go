package bridge

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// ValidatorManager handles validator operations and consensus
type ValidatorManager struct {
	bridge           *Bridge
	validators       map[string]*Validator
	activeValidators map[string]*Validator
	stakeThreshold   *big.Int
	mutex            sync.RWMutex
	consensusEngine  *ConsensusEngine
}

// ConsensusEngine handles validator consensus for bridge operations
type ConsensusEngine struct {
	validators            map[string]*Validator
	requiredConfirmations int
	confirmations         map[string]map[string]*Confirmation
	mutex                 sync.RWMutex
	timeout               time.Duration
}

// Confirmation represents a validator's confirmation of a transaction
type Confirmation struct {
	ValidatorID   string    `json:"validator_id"`
	TransactionID string    `json:"transaction_id"`
	Signature     []byte    `json:"signature"`
	Timestamp     time.Time `json:"timestamp"`
	IsValid       bool      `json:"is_valid"`
}

// NewValidatorManager creates a new validator manager
func NewValidatorManager(bridge *Bridge) *ValidatorManager {
	vm := &ValidatorManager{
		bridge:           bridge,
		validators:       make(map[string]*Validator),
		activeValidators: make(map[string]*Validator),
		stakeThreshold:   big.NewInt(1000000000000000000), // 1 ETH minimum stake
		consensusEngine:  NewConsensusEngine(bridge.config.RequiredConfirmations),
	}
	return vm
}

// NewConsensusEngine creates a new consensus engine
func NewConsensusEngine(requiredConfirmations int) *ConsensusEngine {
	return &ConsensusEngine{
		validators:            make(map[string]*Validator),
		requiredConfirmations: requiredConfirmations,
		confirmations:         make(map[string]map[string]*Confirmation),
		timeout:               30 * time.Minute,
	}
}

// AddValidator adds a new validator to the bridge
func (vm *ValidatorManager) AddValidator(
	address string,
	chainID ChainID,
	stakeAmount *big.Int,
	publicKey *ecdsa.PublicKey,
) (*Validator, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Check if validator already exists
	if _, exists := vm.validators[address]; exists {
		return nil, fmt.Errorf("validator already exists: %s", address)
	}

	// Validate stake amount
	if stakeAmount.Cmp(vm.stakeThreshold) < 0 {
		return nil, fmt.Errorf("stake amount %s is below threshold %s",
			stakeAmount.String(), vm.stakeThreshold.String())
	}

	// Encode public key
	publicKeyBytes, err := vm.encodePublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	// Create validator
	validator := &Validator{
		ID:             generateValidatorID(address),
		Address:        address,
		PublicKey:      publicKeyBytes,
		ChainID:        chainID,
		StakeAmount:    stakeAmount,
		IsActive:       true,
		LastHeartbeat:  time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TotalValidated: 0,
		SuccessRate:    100.0,
	}

	// Store validator
	vm.validators[address] = validator
	vm.activeValidators[address] = validator

	// Add to consensus engine
	vm.consensusEngine.AddValidator(validator)

	// Emit event
	vm.bridge.emitEvent("validator_added", validator)

	return validator, nil
}

// RemoveValidator removes a validator from the bridge
func (vm *ValidatorManager) RemoveValidator(address string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	validator, exists := vm.validators[address]
	if !exists {
		return fmt.Errorf("validator not found: %s", address)
	}

	// Check if validator has pending confirmations
	if vm.consensusEngine.HasPendingConfirmations(validator.ID) {
		return fmt.Errorf("validator has pending confirmations, cannot remove")
	}

	// Remove from active validators
	delete(vm.activeValidators, address)
	validator.IsActive = false
	validator.UpdatedAt = time.Now()

	// Remove from consensus engine
	vm.consensusEngine.RemoveValidator(validator.ID)

	// Emit event
	vm.bridge.emitEvent("validator_removed", validator)

	return nil
}

// UpdateValidatorStake updates a validator's stake amount
func (vm *ValidatorManager) UpdateValidatorStake(address string, newStakeAmount *big.Int) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	validator, exists := vm.validators[address]
	if !exists {
		return fmt.Errorf("validator not found: %s", address)
	}

	// Validate new stake amount
	if newStakeAmount.Cmp(vm.stakeThreshold) < 0 {
		return fmt.Errorf("stake amount %s is below threshold %s",
			newStakeAmount.String(), vm.stakeThreshold.String())
	}

	// Update stake
	validator.StakeAmount = newStakeAmount
	validator.UpdatedAt = time.Now()

	// Emit event
	vm.bridge.emitEvent("validator_stake_updated", validator)

	return nil
}

// Heartbeat updates validator heartbeat
func (vm *ValidatorManager) Heartbeat(address string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	validator, exists := vm.activeValidators[address]
	if !exists {
		return fmt.Errorf("validator not found or inactive: %s", address)
	}

	validator.LastHeartbeat = time.Now()
	validator.UpdatedAt = time.Now()

	return nil
}

// GetActiveValidators returns all active validators
func (vm *ValidatorManager) GetActiveValidators() []*Validator {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	var result []*Validator
	for _, validator := range vm.activeValidators {
		result = append(result, validator)
	}
	return result
}

// GetValidator returns a validator by address
func (vm *ValidatorManager) GetValidator(address string) (*Validator, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	validator, exists := vm.validators[address]
	if !exists {
		return nil, fmt.Errorf("validator not found: %s", address)
	}

	return validator, nil
}

// ValidateTransaction validates a cross-chain transaction
func (vm *ValidatorManager) ValidateTransaction(
	txID string,
	validatorAddress string,
	signature []byte,
) error {
	vm.mutex.RLock()
	validator, exists := vm.activeValidators[validatorAddress]
	vm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("validator not found or inactive: %s", validatorAddress)
	}

	// Get transaction
	transaction, err := vm.bridge.GetTransaction(txID)
	if err != nil {
		return err
	}

	// Verify signature
	if err := vm.verifyTransactionSignature(transaction, signature, validator); err != nil {
		return err
	}

	// Add confirmation to consensus engine
	confirmation := &Confirmation{
		ValidatorID:   validator.ID,
		TransactionID: txID,
		Signature:     signature,
		Timestamp:     time.Now(),
		IsValid:       true,
	}

	vm.consensusEngine.AddConfirmation(txID, confirmation)

	// Check if we have enough confirmations
	if vm.consensusEngine.HasEnoughConfirmations(txID) {
		// Execute transaction
		if err := vm.bridge.ConfirmTransaction(txID, validator.ID, signature); err != nil {
			return err
		}

		// Update validator stats
		vm.updateValidatorStats(validator, true)
	} else {
		// Update validator stats for partial confirmation
		vm.updateValidatorStats(validator, false)
	}

	return nil
}

// verifyTransactionSignature verifies a validator's signature on a transaction
func (vm *ValidatorManager) verifyTransactionSignature(
	transaction *CrossChainTransaction,
	signature []byte,
	validator *Validator,
) error {
	// Create transaction hash for signature verification
	txData := fmt.Sprintf("%s_%s_%s_%s_%s_%s",
		transaction.ID,
		transaction.SourceChain,
		transaction.DestinationChain,
		transaction.SourceAddress,
		transaction.DestinationAddress,
		transaction.Amount.String(),
	)

	// Hash the transaction data
	hash := sha256.Sum256([]byte(txData))

	// Verify the signature using ECDSA
	if err := vm.verifyECDSASignature(hash[:], signature, validator.PublicKey); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// verifyECDSASignature verifies an ECDSA signature
func (vm *ValidatorManager) verifyECDSASignature(hash []byte, signature []byte, publicKey []byte) error {
	// Parse the public key
	pubKey, err := vm.parsePublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Parse the signature
	r, s, err := vm.parseSignature(signature)
	if err != nil {
		return fmt.Errorf("failed to parse signature: %w", err)
	}

	// Verify the signature
	if !ecdsa.Verify(pubKey, hash, r, s) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// parsePublicKey parses a public key from bytes
func (vm *ValidatorManager) parsePublicKey(publicKeyBytes []byte) (*ecdsa.PublicKey, error) {
	// Try to parse as raw coordinates (64 bytes for P-256)
	if len(publicKeyBytes) == 64 {
		x := new(big.Int).SetBytes(publicKeyBytes[:32])
		y := new(big.Int).SetBytes(publicKeyBytes[32:])

		return &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}, nil
	}

	return nil, fmt.Errorf("unsupported public key format")
}

// ECDSASignature represents an ECDSA signature
type ECDSASignature struct {
	R, S *big.Int
}

// parseSignature parses a signature from bytes
func (vm *ValidatorManager) parseSignature(signatureBytes []byte) (*big.Int, *big.Int, error) {
	// Try to parse as DER-encoded signature first
	var sig ECDSASignature
	_, err := asn1.Unmarshal(signatureBytes, &sig)
	if err == nil {
		return sig.R, sig.S, nil
	}

	// If DER parsing fails, try to parse as raw r||s format (64 bytes)
	if len(signatureBytes) == 64 {
		r := new(big.Int).SetBytes(signatureBytes[:32])
		s := new(big.Int).SetBytes(signatureBytes[32:])
		return r, s, nil
	}

	return nil, nil, fmt.Errorf("unsupported signature format")
}

// encodePublicKey encodes an ECDSA public key to bytes
func (vm *ValidatorManager) encodePublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	// Encode as raw coordinates (64 bytes for P-256)
	if publicKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("unsupported curve: %s", publicKey.Curve.Params().Name)
	}

	// Convert coordinates to 32-byte arrays
	xBytes := publicKey.X.Bytes()
	yBytes := publicKey.Y.Bytes()

	// Pad to 32 bytes if necessary
	if len(xBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(xBytes):], xBytes)
		xBytes = padded
	}
	if len(yBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(yBytes):], yBytes)
		yBytes = padded
	}

	// Concatenate x and y coordinates
	result := make([]byte, 64)
	copy(result[:32], xBytes)
	copy(result[32:], yBytes)

	return result, nil
}

// updateValidatorStats updates validator statistics
func (vm *ValidatorManager) updateValidatorStats(validator *Validator, success bool) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	validator.TotalValidated++

	// Update success rate (simplified calculation)
	if success {
		// In a real implementation, you'd track success/failure over time
		validator.SuccessRate = 100.0
	}

	validator.UpdatedAt = time.Now()
}

// generateValidatorID generates a unique validator ID
func generateValidatorID(address string) string {
	data := fmt.Sprintf("%s_%d", address, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// Consensus Engine Methods

// AddValidator adds a validator to the consensus engine
func (ce *ConsensusEngine) AddValidator(validator *Validator) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.validators[validator.ID] = validator
}

// RemoveValidator removes a validator from the consensus engine
func (ce *ConsensusEngine) RemoveValidator(validatorID string) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	delete(ce.validators, validatorID)
}

// AddConfirmation adds a validator confirmation for a transaction
func (ce *ConsensusEngine) AddConfirmation(txID string, confirmation *Confirmation) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	if ce.confirmations[txID] == nil {
		ce.confirmations[txID] = make(map[string]*Confirmation)
	}

	ce.confirmations[txID][confirmation.ValidatorID] = confirmation
}

// HasEnoughConfirmations checks if a transaction has enough confirmations
func (ce *ConsensusEngine) HasEnoughConfirmations(txID string) bool {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	confirmations, exists := ce.confirmations[txID]
	if !exists {
		return false
	}

	// Count valid confirmations
	validCount := 0
	for _, conf := range confirmations {
		if conf.IsValid {
			validCount++
		}
	}

	return validCount >= ce.requiredConfirmations
}

// GetConfirmations returns all confirmations for a transaction
func (ce *ConsensusEngine) GetConfirmations(txID string) map[string]*Confirmation {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	confirmations, exists := ce.confirmations[txID]
	if !exists {
		return make(map[string]*Confirmation)
	}

	// Return a copy to avoid race conditions
	result := make(map[string]*Confirmation)
	for k, v := range confirmations {
		result[k] = v
	}

	return result
}

// GetConfirmationCount returns the number of confirmations for a transaction
func (ce *ConsensusEngine) GetConfirmationCount(txID string) int {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	confirmations, exists := ce.confirmations[txID]
	if !exists {
		return 0
	}

	validCount := 0
	for _, conf := range confirmations {
		if conf.IsValid {
			validCount++
		}
	}

	return validCount
}

// RemoveConfirmations removes all confirmations for a transaction
func (ce *ConsensusEngine) RemoveConfirmations(txID string) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	delete(ce.confirmations, txID)
}

// CleanupExpiredConfirmations removes confirmations older than the timeout
func (ce *ConsensusEngine) CleanupExpiredConfirmations() {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	now := time.Now()
	for txID, confirmations := range ce.confirmations {
		expired := true
		for _, conf := range confirmations {
			if now.Sub(conf.Timestamp) < ce.timeout {
				expired = false
				break
			}
		}
		if expired {
			delete(ce.confirmations, txID)
		}
	}
}

// GetValidatorConfirmation returns a specific validator's confirmation for a transaction
func (ce *ConsensusEngine) GetValidatorConfirmation(txID, validatorID string) (*Confirmation, bool) {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	confirmations, exists := ce.confirmations[txID]
	if !exists {
		return nil, false
	}

	conf, exists := confirmations[validatorID]
	return conf, exists
}

// IsValidatorConfirmed checks if a specific validator has confirmed a transaction
func (ce *ConsensusEngine) IsValidatorConfirmed(txID, validatorID string) bool {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	confirmations, exists := ce.confirmations[txID]
	if !exists {
		return false
	}

	conf, exists := confirmations[validatorID]
	return exists && conf.IsValid
}

// GetActiveValidators returns the list of active validators
func (ce *ConsensusEngine) GetActiveValidators() []*Validator {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	validators := make([]*Validator, 0, len(ce.validators))
	for _, validator := range ce.validators {
		if validator.IsActive {
			validators = append(validators, validator)
		}
	}

	return validators
}

// GetRequiredConfirmations returns the required number of confirmations
func (ce *ConsensusEngine) GetRequiredConfirmations() int {
	return ce.requiredConfirmations
}

// SetRequiredConfirmations sets the required number of confirmations
func (ce *ConsensusEngine) SetRequiredConfirmations(count int) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	ce.requiredConfirmations = count
}

// GetTimeout returns the confirmation timeout
func (ce *ConsensusEngine) GetTimeout() time.Duration {
	return ce.timeout
}

// SetTimeout sets the confirmation timeout
func (ce *ConsensusEngine) SetTimeout(timeout time.Duration) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	ce.timeout = timeout
}

// HasPendingConfirmations checks if a validator has pending confirmations
func (ce *ConsensusEngine) HasPendingConfirmations(validatorID string) bool {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	for _, txConfirmations := range ce.confirmations {
		if _, exists := txConfirmations[validatorID]; exists {
			return true
		}
	}

	return false
}
