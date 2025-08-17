package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// Bridge represents the main bridge instance
type Bridge struct {
	config                    *BridgeConfig
	validators               map[string]*Validator
	transactions             map[string]*CrossChainTransaction
	assetMappings            map[string]*AssetMapping
	validatorSet             map[string]bool // Active validator addresses
	mutex                    sync.RWMutex
	lastBlockNumber          map[ChainID]*big.Int
	eventHandlers            map[string][]func(interface{})
	validatorManager         *ValidatorManager
	crossChainTxManager      *CrossChainTransactionManager
	securityManager          *SecurityManager
}

// NewBridge creates a new bridge instance
func NewBridge(config *BridgeConfig) *Bridge {
	if config == nil {
		config = getDefaultConfig()
	}

	bridge := &Bridge{
		config:          config,
		validators:      make(map[string]*Validator),
		transactions:    make(map[string]*CrossChainTransaction),
		assetMappings:   make(map[string]*AssetMapping),
		validatorSet:    make(map[string]bool),
		lastBlockNumber: make(map[ChainID]*big.Int),
		eventHandlers:   make(map[string][]func(interface{})),
	}

	// Initialize managers
	bridge.validatorManager = NewValidatorManager(bridge)
	bridge.crossChainTxManager = NewCrossChainTransactionManager(bridge)
	bridge.securityManager = NewSecurityManager(bridge)

	// Initialize default asset mappings
	bridge.initializeDefaultAssetMappings()

	return bridge
}

// mustParseBigInt parses a string to big.Int, panics on error
func mustParseBigInt(s string) *big.Int {
	result, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to parse big.Int: " + s)
	}
	return result
}

// getDefaultConfig returns default bridge configuration
func getDefaultConfig() *BridgeConfig {
	return &BridgeConfig{
		ID:                    "gochain_bridge",
		Status:                BridgeStatusActive,
		MinValidators:         3,
		RequiredConfirmations: 2,
		MaxTransactionAmount:  big.NewInt(1000000000000000000), // 1 ETH
		MinTransactionAmount:  big.NewInt(1000000000000000),    // 0.001 ETH
		TransactionTimeout:    24 * time.Hour,
		GasLimit:              big.NewInt(500000),
		MaxDailyVolume:        mustParseBigInt("10000000000000000000"), // 10 ETH
		DailyVolumeUsed:       big.NewInt(0),
		FeeCollector:          "",
		EmergencyPaused:       false,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}

// initializeDefaultAssetMappings sets up default asset mappings
func (b *Bridge) initializeDefaultAssetMappings() {
	// GoChain <-> Ethereum native token mapping
	b.assetMappings["gochain_ethereum_native"] = &AssetMapping{
		ID:               "gochain_ethereum_native",
		SourceChain:      ChainIDGoChain,
		DestinationChain: ChainIDEthereum,
		SourceAsset:      "0x0000000000000000000000000000000000000000", // Native token
		DestinationAsset: "0x0000000000000000000000000000000000000000", // Native token
		AssetType:        AssetTypeNative,
		Decimals:         18,
		IsActive:         true,
		MinAmount:        big.NewInt(1000000000000000),            // 0.001
		MaxAmount:        big.NewInt(1000000000000000000),         // 1
		DailyLimit:       mustParseBigInt("10000000000000000000"), // 10
		DailyUsed:        big.NewInt(0),
		FeePercentage:    0.1, // 0.1%
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Ethereum <-> GoChain native token mapping
	b.assetMappings["ethereum_gochain_native"] = &AssetMapping{
		ID:               "ethereum_gochain_native",
		SourceChain:      ChainIDEthereum,
		DestinationChain: ChainIDGoChain,
		SourceAsset:      "0x0000000000000000000000000000000000000000", // Native token
		DestinationAsset: "0x0000000000000000000000000000000000000000", // Native token
		AssetType:        AssetTypeNative,
		Decimals:         18,
		IsActive:         true,
		MinAmount:        big.NewInt(1000000000000000),            // 0.001
		MaxAmount:        big.NewInt(1000000000000000000),         // 1
		DailyLimit:       mustParseBigInt("10000000000000000000"), // 10
		DailyUsed:        big.NewInt(0),
		FeePercentage:    0.1, // 0.1%
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// InitiateTransfer initiates a cross-chain transfer
func (b *Bridge) InitiateTransfer(
	sourceChain ChainID,
	destinationChain ChainID,
	sourceAddress string,
	destinationAddress string,
	assetType AssetType,
	assetAddress string,
	amount *big.Int,
	tokenID *big.Int,
) (*CrossChainTransaction, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Validate bridge status
	if err := b.validateBridgeStatus(); err != nil {
		return nil, err
	}

	// Validate chains
	if err := b.validateChains(sourceChain, destinationChain); err != nil {
		return nil, err
	}

	// Validate addresses
	if err := b.validateAddresses(sourceAddress, destinationAddress); err != nil {
		return nil, err
	}

	// Validate asset mapping
	assetMapping, err := b.getAssetMapping(sourceChain, destinationChain, assetAddress, assetType)
	if err != nil {
		return nil, err
	}

	// Validate amount
	if err := b.validateAmount(amount, assetMapping); err != nil {
		return nil, err
	}

	// Check daily limits
	if err := b.checkDailyLimits(amount, assetMapping); err != nil {
		return nil, err
	}

	// Generate transaction ID
	txID := b.generateTransactionID(sourceChain, destinationChain, sourceAddress, amount)

	// Calculate fee
	fee := b.calculateFee(amount, assetMapping.FeePercentage)

	// Create transaction
	transaction := &CrossChainTransaction{
		ID:                 txID,
		SourceChain:        sourceChain,
		DestinationChain:   destinationChain,
		SourceAddress:      sourceAddress,
		DestinationAddress: destinationAddress,
		AssetType:          assetType,
		AssetAddress:       assetAddress,
		Amount:             amount,
		TokenID:            tokenID,
		Status:             TransactionStatusPending,
		Fee:                fee,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Store transaction
	b.transactions[txID] = transaction

	// Update daily volume
	b.updateDailyVolume(amount, assetMapping)

	// Emit event
	b.emitEvent("transfer_initiated", transaction)

	return transaction, nil
}

// ConfirmTransaction confirms a cross-chain transaction
func (b *Bridge) ConfirmTransaction(txID string, validatorID string, signature []byte) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Get transaction
	transaction, exists := b.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	// Validate transaction status
	if transaction.Status != TransactionStatusPending {
		return fmt.Errorf("transaction status is %s, expected pending", transaction.Status)
	}

	// Validate validator
	validator, exists := b.validators[validatorID]
	if !exists || !validator.IsActive {
		return ErrValidatorInactive
	}

	// Verify signature
	if err := b.verifySignature(transaction, signature, validator.Address); err != nil {
		return ErrInvalidSignature
	}

	// Update transaction status
	transaction.Status = TransactionStatusConfirmed
	transaction.ValidatorID = validatorID
	transaction.UpdatedAt = time.Now()

	// Check if we have enough confirmations
	if b.hasEnoughConfirmations(txID) {
		transaction.Status = TransactionStatusExecuted
		transaction.ExecutedAt = &time.Time{}
		*transaction.ExecutedAt = time.Now()
		transaction.UpdatedAt = time.Now()

		// Update validator stats
		validator.TotalValidated++
		validator.UpdatedAt = time.Now()

		// Emit event
		b.emitEvent("transaction_executed", transaction)
	}

	return nil
}

// ExecuteTransaction executes a confirmed transaction on the destination chain
func (b *Bridge) ExecuteTransaction(txID string, destinationTxHash string, gasUsed *big.Int) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Get transaction
	transaction, exists := b.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	// Validate transaction status
	if transaction.Status != TransactionStatusConfirmed {
		return fmt.Errorf("transaction status is %s, expected confirmed", transaction.Status)
	}

	// Update transaction
	transaction.Status = TransactionStatusExecuted
	transaction.DestinationTxHash = destinationTxHash
	transaction.GasUsed = gasUsed
	transaction.ExecutedAt = &time.Time{}
	*transaction.ExecutedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	// Emit event
	b.emitEvent("transaction_executed", transaction)

	return nil
}

// FailTransaction marks a transaction as failed
func (b *Bridge) FailTransaction(txID string, reason string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Get transaction
	transaction, exists := b.transactions[txID]
	if !exists {
		return ErrTransactionNotFound
	}

	// Update transaction
	transaction.Status = TransactionStatusFailed
	transaction.FailureReason = reason
	transaction.FailedAt = &time.Time{}
	*transaction.FailedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	// Emit event
	b.emitEvent("transaction_failed", transaction)

	return nil
}

// GetTransaction returns a transaction by ID
func (b *Bridge) GetTransaction(txID string) (*CrossChainTransaction, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	transaction, exists := b.transactions[txID]
	if !exists {
		return nil, ErrTransactionNotFound
	}

	return transaction, nil
}

// GetTransactionsByStatus returns transactions by status
func (b *Bridge) GetTransactionsByStatus(status TransactionStatus) []*CrossChainTransaction {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var result []*CrossChainTransaction
	for _, tx := range b.transactions {
		if tx.Status == status {
			result = append(result, tx)
		}
	}

	return result
}

// GetTransactionsByAddress returns transactions for a specific address
func (b *Bridge) GetTransactionsByAddress(address string) []*CrossChainTransaction {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var result []*CrossChainTransaction
	for _, tx := range b.transactions {
		if tx.SourceAddress == address || tx.DestinationAddress == address {
			result = append(result, tx)
		}
	}

	return result
}

// GetBridgeStats returns bridge statistics
func (b *Bridge) GetBridgeStats() map[string]interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_transactions"] = len(b.transactions)
	stats["pending_transactions"] = len(b.GetTransactionsByStatus(TransactionStatusPending))
	stats["confirmed_transactions"] = len(b.GetTransactionsByStatus(TransactionStatusConfirmed))
	stats["executed_transactions"] = len(b.GetTransactionsByStatus(TransactionStatusExecuted))
	stats["failed_transactions"] = len(b.GetTransactionsByStatus(TransactionStatusFailed))
	stats["total_validators"] = len(b.validators)
	stats["active_validators"] = len(b.validatorSet)
	stats["total_asset_mappings"] = len(b.assetMappings)
	stats["bridge_status"] = b.config.Status
	stats["daily_volume_used"] = b.config.DailyVolumeUsed.String()
	stats["max_daily_volume"] = b.config.MaxDailyVolume.String()

	// Add manager stats
	if b.validatorManager != nil {
		validatorStats := b.validatorManager.GetActiveValidators()
		stats["active_validators_count"] = len(validatorStats)
	}

	if b.crossChainTxManager != nil {
		batchStats := b.crossChainTxManager.GetBatchStats()
		for k, v := range batchStats {
			stats["batch_"+k] = v
		}
	}

	if b.securityManager != nil {
		securityStats := b.securityManager.GetSecurityStats()
		for k, v := range securityStats {
			stats["security_"+k] = v
		}
	}

	return stats
}

// validateBridgeStatus validates that the bridge is active
func (b *Bridge) validateBridgeStatus() error {
	switch b.config.Status {
	case BridgeStatusActive:
		return nil
	case BridgeStatusPaused:
		return ErrBridgePaused
	case BridgeStatusEmergency:
		return ErrBridgeEmergency
	case BridgeStatusUpgrading:
		return fmt.Errorf("bridge is upgrading")
	default:
		return fmt.Errorf("unknown bridge status: %s", b.config.Status)
	}
}

// validateChains validates source and destination chains
func (b *Bridge) validateChains(sourceChain, destinationChain ChainID) error {
	if sourceChain == destinationChain {
		return fmt.Errorf("source and destination chains cannot be the same")
	}

	// Validate chain IDs
	validChains := map[ChainID]bool{
		ChainIDGoChain:  true,
		ChainIDEthereum: true,
		ChainIDPolygon:  true,
		ChainIDArbitrum: true,
		ChainIDOptimism: true,
	}

	if !validChains[sourceChain] {
		return fmt.Errorf("invalid source chain: %s", sourceChain)
	}

	if !validChains[destinationChain] {
		return fmt.Errorf("invalid destination chain: %s", destinationChain)
	}

	return nil
}

// validateAddresses validates source and destination addresses
func (b *Bridge) validateAddresses(sourceAddress, destinationAddress string) error {
	if sourceAddress == "" || destinationAddress == "" {
		return ErrInvalidAddress
	}

	// Basic address validation (could be enhanced with chain-specific validation)
	if len(sourceAddress) < 20 || len(destinationAddress) < 20 {
		return ErrInvalidAddress
	}

	return nil
}

// getAssetMapping gets the asset mapping for the specified chains and asset
func (b *Bridge) getAssetMapping(sourceChain, destinationChain ChainID, assetAddress string, assetType AssetType) (*AssetMapping, error) {
	mappingKey := fmt.Sprintf("%s_%s_%s", sourceChain, destinationChain, assetType)

	mapping, exists := b.assetMappings[mappingKey]
	if !exists || !mapping.IsActive {
		return nil, ErrAssetNotSupported
	}

	return mapping, nil
}

// validateAmount validates the transfer amount
func (b *Bridge) validateAmount(amount *big.Int, assetMapping *AssetMapping) error {
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return ErrAmountTooSmall
	}

	if amount.Cmp(assetMapping.MinAmount) < 0 {
		return fmt.Errorf("amount %s is below minimum %s", amount.String(), assetMapping.MinAmount.String())
	}

	if amount.Cmp(assetMapping.MaxAmount) > 0 {
		return fmt.Errorf("amount %s exceeds maximum %s", amount.String(), assetMapping.MaxAmount.String())
	}

	if amount.Cmp(b.config.MaxTransactionAmount) > 0 {
		return ErrAmountTooLarge
	}

	if amount.Cmp(b.config.MinTransactionAmount) < 0 {
		return ErrAmountTooSmall
	}

	return nil
}

// checkDailyLimits checks if the transfer would exceed daily limits
func (b *Bridge) checkDailyLimits(amount *big.Int, assetMapping *AssetMapping) error {
	// Check asset-specific daily limit
	dailyRemaining := new(big.Int).Sub(assetMapping.DailyLimit, assetMapping.DailyUsed)
	if amount.Cmp(dailyRemaining) > 0 {
		return ErrDailyLimitExceeded
	}

	// Check bridge-wide daily limit
	bridgeDailyRemaining := new(big.Int).Sub(b.config.MaxDailyVolume, b.config.DailyVolumeUsed)
	if amount.Cmp(bridgeDailyRemaining) > 0 {
		return ErrDailyLimitExceeded
	}

	return nil
}

// updateDailyVolume updates the daily volume counters
func (b *Bridge) updateDailyVolume(amount *big.Int, assetMapping *AssetMapping) {
	// Update asset-specific daily volume
	assetMapping.DailyUsed.Add(assetMapping.DailyUsed, amount)
	assetMapping.UpdatedAt = time.Now()

	// Update bridge-wide daily volume
	b.config.DailyVolumeUsed.Add(b.config.DailyVolumeUsed, amount)
	b.config.UpdatedAt = time.Now()
}

// calculateFee calculates the bridge fee for a transfer
func (b *Bridge) calculateFee(amount *big.Int, feePercentage float64) *big.Int {
	fee := new(big.Int).Mul(amount, big.NewInt(int64(feePercentage*1000)))
	fee.Div(fee, big.NewInt(100000)) // Divide by 100000 to get percentage
	return fee
}

// generateTransactionID generates a unique transaction ID
func (b *Bridge) generateTransactionID(sourceChain, destinationChain ChainID, sourceAddress string, amount *big.Int) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%d", sourceChain, destinationChain, sourceAddress, amount.String(), time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

// hasEnoughConfirmations checks if a transaction has enough validator confirmations
func (b *Bridge) hasEnoughConfirmations(txID string) bool {
	// This is a simplified implementation
	// In a real bridge, you would track individual validator confirmations
	// For now, always return false to prevent immediate execution
	return false
}

// verifySignature verifies a validator signature
func (b *Bridge) verifySignature(transaction *CrossChainTransaction, signature []byte, validatorAddress string) error {
	// This is a simplified implementation
	// In a real bridge, you would verify the signature against the transaction data
	return nil
}

// emitEvent emits an event to registered handlers
func (b *Bridge) emitEvent(eventType string, data interface{}) {
	if handlers, exists := b.eventHandlers[eventType]; exists {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}

// On registers an event handler
func (b *Bridge) On(eventType string, handler func(interface{})) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.eventHandlers[eventType] == nil {
		b.eventHandlers[eventType] = make([]func(interface{}), 0)
	}

	b.eventHandlers[eventType] = append(b.eventHandlers[eventType], handler)
}

// GetValidatorManager returns the validator manager
func (b *Bridge) GetValidatorManager() *ValidatorManager {
	return b.validatorManager
}

// GetCrossChainTransactionManager returns the cross-chain transaction manager
func (b *Bridge) GetCrossChainTransactionManager() *CrossChainTransactionManager {
	return b.crossChainTxManager
}

// GetSecurityManager returns the security manager
func (b *Bridge) GetSecurityManager() *SecurityManager {
	return b.securityManager
}
