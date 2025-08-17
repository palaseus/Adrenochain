package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CrossChainTransactionManager handles cross-chain transaction operations
type CrossChainTransactionManager struct {
	bridge        *Bridge
	transactions  map[string]*CrossChainTransaction
	batches       map[string]*TransactionBatch
	mutex         sync.RWMutex
	batchTimeout  time.Duration
	maxBatchSize  int
}

// TransactionBatch represents a batch of cross-chain transactions
type TransactionBatch struct {
	ID           string                    `json:"id"`
	Transactions []*CrossChainTransaction  `json:"transactions"`
	Status       BatchStatus               `json:"status"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
	ExecutedAt   *time.Time                `json:"executed_at,omitempty"`
	GasUsed      *big.Int                  `json:"gas_used,omitempty"`
	TotalFee     *big.Int                  `json:"total_fee,omitempty"`
}

// BatchStatus represents the status of a transaction batch
type BatchStatus string

const (
	BatchStatusPending   BatchStatus = "pending"
	BatchStatusConfirmed BatchStatus = "confirmed"
	BatchStatusExecuted  BatchStatus = "executed"
	BatchStatusFailed    BatchStatus = "failed"
)

// NewCrossChainTransactionManager creates a new cross-chain transaction manager
func NewCrossChainTransactionManager(bridge *Bridge) *CrossChainTransactionManager {
	return &CrossChainTransactionManager{
		bridge:       bridge,
		transactions: make(map[string]*CrossChainTransaction),
		batches:      make(map[string]*TransactionBatch),
		batchTimeout: 10 * time.Minute,
		maxBatchSize: 100,
	}
}

// InitiateBatchTransfer initiates a batch cross-chain transfer
func (ctm *CrossChainTransactionManager) InitiateBatchTransfer(
	transfers []*TransferRequest,
) (*TransactionBatch, error) {
	ctm.mutex.Lock()
	defer ctm.mutex.Unlock()

	if len(transfers) == 0 {
		return nil, fmt.Errorf("no transfers provided")
	}

	if len(transfers) > ctm.maxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(transfers), ctm.maxBatchSize)
	}

	// Validate all transfers
	for _, transfer := range transfers {
		if err := ctm.validateTransferRequest(transfer); err != nil {
			return nil, fmt.Errorf("invalid transfer: %w", err)
		}
	}

	// Create batch
	batch := &TransactionBatch{
		ID:           ctm.generateBatchID(),
		Transactions: make([]*CrossChainTransaction, 0, len(transfers)),
		Status:       BatchStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Process each transfer
	for _, transfer := range transfers {
		transaction, err := ctm.processTransferRequest(transfer)
		if err != nil {
			return nil, fmt.Errorf("failed to process transfer: %w", err)
		}

		batch.Transactions = append(batch.Transactions, transaction)
	}

	// Store batch
	ctm.batches[batch.ID] = batch

	// Emit event
	ctm.bridge.emitEvent("batch_transfer_initiated", batch)

	return batch, nil
}

// ProcessBatch processes a batch of transactions
func (ctm *CrossChainTransactionManager) ProcessBatch(batchID string) error {
	ctm.mutex.Lock()
	batch, exists := ctm.batches[batchID]
	ctm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	if batch.Status != BatchStatusPending {
		return fmt.Errorf("batch status is %s, expected pending", batch.Status)
	}

	// Process each transaction in the batch
	for _, transaction := range batch.Transactions {
		if err := ctm.processTransaction(transaction); err != nil {
			// Mark batch as failed
			ctm.markBatchFailed(batchID, fmt.Sprintf("transaction %s failed: %v", transaction.ID, err))
			return err
		}
	}

	// Mark batch as confirmed
	ctm.markBatchConfirmed(batchID)

	return nil
}

// ExecuteBatch executes a confirmed batch on the destination chain
func (ctm *CrossChainTransactionManager) ExecuteBatch(
	batchID string,
	destinationTxHash string,
	gasUsed *big.Int,
) error {
	ctm.mutex.Lock()
	batch, exists := ctm.batches[batchID]
	ctm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	if batch.Status != BatchStatusConfirmed {
		return fmt.Errorf("batch status is %s, expected confirmed", batch.Status)
	}

	// Execute each transaction in the batch
	for _, transaction := range batch.Transactions {
		if err := ctm.bridge.ExecuteTransaction(transaction.ID, destinationTxHash, gasUsed); err != nil {
			// Mark batch as failed
			ctm.markBatchFailed(batchID, fmt.Sprintf("failed to execute transaction %s: %v", transaction.ID, err))
			return err
		}
	}

	// Mark batch as executed
	ctm.markBatchExecuted(batchID, destinationTxHash, gasUsed)

	return nil
}

// GetBatch returns a batch by ID
func (ctm *CrossChainTransactionManager) GetBatch(batchID string) (*TransactionBatch, error) {
	ctm.mutex.RLock()
	defer ctm.mutex.RUnlock()

	batch, exists := ctm.batches[batchID]
	if !exists {
		return nil, fmt.Errorf("batch not found: %s", batchID)
	}

	return batch, nil
}

// GetBatchesByStatus returns batches by status
func (ctm *CrossChainTransactionManager) GetBatchesByStatus(status BatchStatus) []*TransactionBatch {
	ctm.mutex.RLock()
	defer ctm.mutex.RUnlock()

	var result []*TransactionBatch
	for _, batch := range ctm.batches {
		if batch.Status == status {
			result = append(result, batch)
		}
	}

	return result
}

// RetryFailedBatch retries a failed batch
func (ctm *CrossChainTransactionManager) RetryFailedBatch(batchID string) error {
	ctm.mutex.Lock()
	batch, exists := ctm.batches[batchID]
	ctm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	if batch.Status != BatchStatusFailed {
		return fmt.Errorf("batch status is %s, expected failed", batch.Status)
	}

	// Reset batch status
	batch.Status = BatchStatusPending
	batch.UpdatedAt = time.Now()

	// Reset transaction statuses
	for _, transaction := range batch.Transactions {
		transaction.Status = TransactionStatusPending
		transaction.UpdatedAt = time.Now()
	}

	// Emit event
	ctm.bridge.emitEvent("batch_retry_initiated", batch)

	return nil
}

// validateTransferRequest validates a transfer request
func (ctm *CrossChainTransactionManager) validateTransferRequest(transfer *TransferRequest) error {
	if transfer == nil {
		return fmt.Errorf("transfer request is nil")
	}

	if transfer.SourceChain == transfer.DestinationChain {
		return fmt.Errorf("source and destination chains cannot be the same")
	}

	if transfer.SourceAddress == "" || transfer.DestinationAddress == "" {
		return fmt.Errorf("invalid addresses")
	}

	if transfer.Amount == nil || transfer.Amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("invalid amount")
	}

	return nil
}

// processTransferRequest processes a single transfer request
func (ctm *CrossChainTransactionManager) processTransferRequest(
	transfer *TransferRequest,
) (*CrossChainTransaction, error) {
	// Use the existing bridge InitiateTransfer method
	transaction, err := ctm.bridge.InitiateTransfer(
		transfer.SourceChain,
		transfer.DestinationChain,
		transfer.SourceAddress,
		transfer.DestinationAddress,
		transfer.AssetType,
		transfer.AssetAddress,
		transfer.Amount,
		transfer.TokenID,
	)

	if err != nil {
		return nil, err
	}

	// Store transaction
	ctm.transactions[transaction.ID] = transaction

	return transaction, nil
}

// processTransaction processes a single transaction
func (ctm *CrossChainTransactionManager) processTransaction(transaction *CrossChainTransaction) error {
	// This would typically involve:
	// 1. Validating the transaction
	// 2. Checking if it meets batch requirements
	// 3. Preparing it for execution
	// 4. Updating status

	// For now, we'll just mark it as confirmed
	transaction.Status = TransactionStatusConfirmed
	transaction.UpdatedAt = time.Now()

	return nil
}

// markBatchConfirmed marks a batch as confirmed
func (ctm *CrossChainTransactionManager) markBatchConfirmed(batchID string) {
	ctm.mutex.Lock()
	defer ctm.mutex.Unlock()

	if batch, exists := ctm.batches[batchID]; exists {
		batch.Status = BatchStatusConfirmed
		batch.UpdatedAt = time.Now()

		// Emit event
		ctm.bridge.emitEvent("batch_confirmed", batch)
	}
}

// markBatchExecuted marks a batch as executed
func (ctm *CrossChainTransactionManager) markBatchExecuted(
	batchID string,
	destinationTxHash string,
	gasUsed *big.Int,
) {
	ctm.mutex.Lock()
	defer ctm.mutex.Unlock()

	if batch, exists := ctm.batches[batchID]; exists {
		batch.Status = BatchStatusExecuted
		batch.ExecutedAt = &time.Time{}
		*batch.ExecutedAt = time.Now()
		batch.UpdatedAt = time.Now()

		// Calculate total fee
		totalFee := big.NewInt(0)
		for _, transaction := range batch.Transactions {
			if transaction.Fee != nil {
				totalFee.Add(totalFee, transaction.Fee)
			}
		}
		batch.TotalFee = totalFee

		// Emit event
		ctm.bridge.emitEvent("batch_executed", batch)
	}
}

// markBatchFailed marks a batch as failed
func (ctm *CrossChainTransactionManager) markBatchFailed(batchID string, reason string) {
	ctm.mutex.Lock()
	defer ctm.mutex.Unlock()

	if batch, exists := ctm.batches[batchID]; exists {
		batch.Status = BatchStatusFailed
		batch.UpdatedAt = time.Now()

		// Emit event
		ctm.bridge.emitEvent("batch_failed", map[string]interface{}{
			"batch_id": batchID,
			"reason":   reason,
		})
	}
}

// generateBatchID generates a unique batch ID
func (ctm *CrossChainTransactionManager) generateBatchID() string {
	data := fmt.Sprintf("batch_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// CleanupExpiredBatches removes expired batches
func (ctm *CrossChainTransactionManager) CleanupExpiredBatches() {
	ctm.mutex.Lock()
	defer ctm.mutex.Unlock()

	now := time.Now()
	for batchID, batch := range ctm.batches {
		if now.Sub(batch.CreatedAt) > ctm.batchTimeout {
			delete(ctm.batches, batchID)
		}
	}
}

// GetBatchStats returns batch statistics
func (ctm *CrossChainTransactionManager) GetBatchStats() map[string]interface{} {
	ctm.mutex.RLock()
	defer ctm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_batches"] = len(ctm.batches)
	stats["pending_batches"] = len(ctm.GetBatchesByStatus(BatchStatusPending))
	stats["confirmed_batches"] = len(ctm.GetBatchesByStatus(BatchStatusConfirmed))
	stats["executed_batches"] = len(ctm.GetBatchesByStatus(BatchStatusExecuted))
	stats["failed_batches"] = len(ctm.GetBatchesByStatus(BatchStatusFailed))

	return stats
}
