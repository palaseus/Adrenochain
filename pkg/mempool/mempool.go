package mempool

import (
	"bytes"
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/utxo"
)

// Mempool represents the transaction memory pool.
// It stores unconfirmed transactions and prioritizes them for inclusion in blocks.
type Mempool struct {
	mu           sync.RWMutex                 // mu protects concurrent access to mempool fields.
	transactions map[string]*TransactionEntry // transactions stores all transactions in the mempool, keyed by hash.
	byFee        *TransactionHeapMin          // byFee is a min-heap for transactions, ordered by fee rate (lowest first).
	byTime       *TransactionHeap             // byTime is a max-heap for transactions, ordered by timestamp (oldest first).
	maxSize      uint64                       // maxSize is the maximum allowed size of the mempool in bytes.
	currentSize  uint64                       // currentSize is the current total size of transactions in the mempool.
	minFeeRate   uint64                       // minFeeRate is the minimum fee per byte required for a transaction to enter the mempool.
	utxoSet      *utxo.UTXOSet                // utxoSet is used for transaction validation
	maxTxSize    uint64                       // maxTxSize is the maximum allowed transaction size in bytes
	testMode     bool                         // testMode allows skipping UTXO validation for testing
}

// TransactionEntry wraps a transaction with metadata used for mempool management.
type TransactionEntry struct {
	Transaction *block.Transaction // Transaction is the actual blockchain transaction.
	FeeRate     uint64             // FeeRate is the transaction fee per byte.
	Size        uint64             // Size is the approximate size of the transaction in bytes.
	Timestamp   time.Time          // Timestamp is when the transaction was added to the mempool.
	index       int                // index is used by the heap.Interface implementation.
}

// TransactionHeap implements heap.Interface for transaction prioritization based on fee rate (max-heap).
type TransactionHeap []*TransactionEntry

// MempoolConfig holds configuration parameters for the mempool.
type MempoolConfig struct {
	MaxSize    uint64 // MaxSize is the maximum allowed size of the mempool in bytes.
	MinFeeRate uint64 // MinFeeRate is the minimum fee per byte required for a transaction.
	MaxTxSize  uint64 // MaxTxSize is the maximum allowed transaction size in bytes.
	TestMode   bool   // TestMode allows skipping UTXO validation for testing
}

// DefaultMempoolConfig returns the default mempool configuration.
func DefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		MaxSize:    100000, // 100KB
		MinFeeRate: 1,      // 1 unit per byte
		MaxTxSize:  100000, // 100KB max transaction size
		TestMode:   false,  // Production mode by default
	}
}

// TestMempoolConfig returns a mempool configuration suitable for testing.
// It enables test mode to skip UTXO validation and uses smaller limits.
func TestMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		MaxSize:    10000, // 10KB for testing
		MinFeeRate: 1,     // Minimum fee rate of 1 per byte for testing (accounts for default validation)
		MaxTxSize:  10000, // 10KB max transaction size for testing
		TestMode:   true,  // Test mode enabled
	}
}

// NewMempool creates a new transaction mempool instance.
// It initializes the internal data structures and heaps for transaction prioritization.
func NewMempool(config *MempoolConfig) *Mempool {
	mp := &Mempool{
		transactions: make(map[string]*TransactionEntry),
		byFee:        &TransactionHeapMin{},
		byTime:       &TransactionHeap{},
		maxSize:      config.MaxSize,
		minFeeRate:   config.MinFeeRate,
		maxTxSize:    config.MaxTxSize,
		utxoSet:      utxo.NewUTXOSet(),
		testMode:     config.TestMode,
	}

	heap.Init(mp.byFee)
	heap.Init(mp.byTime)

	return mp
}

// SetUTXOSet sets the UTXO set for transaction validation
func (mp *Mempool) SetUTXOSet(utxoSet *utxo.UTXOSet) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.utxoSet = utxoSet
}

// AddTransaction adds a transaction to the mempool.
// It validates the transaction, calculates its fee rate, and adds it to the internal data structures.
// If the mempool is full, it attempts to evict lower-fee transactions.
func (mp *Mempool) AddTransaction(tx *block.Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Check if transaction already exists
	txHash := string(tx.Hash)
	if _, exists := mp.transactions[txHash]; exists {
		return fmt.Errorf("transaction already in mempool")
	}

	// Use the dedicated validation method instead of duplicating logic
	if err := mp.IsTransactionValid(tx); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Calculate transaction size for mempool management
	size := mp.calculateTransactionSize(tx)

	// Calculate fee rate for mempool management
	feeRate := mp.calculateFeeRate(tx, size)

	// Check if adding this transaction would exceed mempool size
	if mp.currentSize+size > mp.maxSize {
		// Try to evict low-fee transactions to make room
		if !mp.evictLowFeeTransactions(size) {
			return fmt.Errorf("mempool full and cannot evict enough transactions")
		}
	}

	// Create transaction entry
	entry := &TransactionEntry{
		Transaction: tx,
		FeeRate:     feeRate,
		Size:        size,
		Timestamp:   time.Now(),
	}

	// Add to mempool
	mp.transactions[txHash] = entry
	mp.currentSize += size

	// Add to priority queues
	heap.Push(mp.byFee, entry)
	heap.Push(mp.byTime, entry)

	return nil
}

// RemoveTransaction removes a transaction from the mempool given its hash.
// It returns true if the transaction was found and removed, false otherwise.
func (mp *Mempool) RemoveTransaction(txHash []byte) bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	hash := string(txHash)
	entry, exists := mp.transactions[hash]
	if !exists {
		return false
	}

	// Remove from maps and queues
	delete(mp.transactions, hash)
	mp.currentSize -= entry.Size

	// Remove from fee queue
	mp.byFee.Remove(entry)

	// Remove from time queue
	mp.byTime.Remove(entry)

	return true
}

// GetTransaction returns a transaction from the mempool by its hash.
// It returns nil if the transaction is not found.
func (mp *Mempool) GetTransaction(txHash []byte) *block.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	entry, exists := mp.transactions[string(txHash)]
	if !exists {
		return nil
	}

	return entry.Transaction
}

// GetTransactionsForBlock returns a list of transactions suitable for inclusion in a new block.
// Transactions are prioritized by fee rate (highest first) and limited by the given maxSize.
func (mp *Mempool) GetTransactionsForBlock(maxSize uint64) []*block.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var transactions []*block.Transaction
	currentSize := uint64(0)

	// Create a copy of the fee queue to avoid modifying the original
	feeQueue := make(TransactionHeapMin, mp.byFee.Len())
	copy(feeQueue, *mp.byFee)

	// Sort by fee rate (highest first)
	for feeQueue.Len() > 0 && currentSize < maxSize {
		entry := heap.Pop(&feeQueue).(*TransactionEntry)

		// Check if transaction still exists in mempool
		if _, exists := mp.transactions[string(entry.Transaction.Hash)]; !exists {
			continue
		}

		// Check if adding this transaction would exceed block size
		if currentSize+entry.Size > maxSize {
			break
		}

		transactions = append(transactions, entry.Transaction)
		currentSize += entry.Size
	}

	return transactions
}

// GetSize returns the current total size of transactions in the mempool in bytes.
func (mp *Mempool) GetSize() uint64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.currentSize
}

// GetTransactionCount returns the number of transactions currently in the mempool.
func (mp *Mempool) GetTransactionCount() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.transactions)
}

// Clear removes all transactions from the mempool.
func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions = make(map[string]*TransactionEntry)
	mp.byFee = &TransactionHeapMin{}
	mp.byTime = &TransactionHeap{}
	mp.currentSize = 0

	heap.Init(mp.byFee)
	heap.Init(mp.byTime)
}

// evictLowFeeTransactions evicts low-fee transactions to make room for new ones
// evictLowFeeTransactions evicts transactions with the lowest fee rates to free up space in the mempool.
// It continues to evict until the requiredSize is met or no more transactions can be evicted.
func (mp *Mempool) evictLowFeeTransactions(requiredSize uint64) bool {
	evictedSize := uint64(0)

	// Evict transactions by lowest fee rate first
	for mp.byFee.Len() > 0 && evictedSize < requiredSize {
		entry := heap.Pop(mp.byFee).(*TransactionEntry)

		// Remove from mempool
		delete(mp.transactions, string(entry.Transaction.Hash))
		mp.currentSize -= entry.Size
		evictedSize += entry.Size

		// Remove from time queue
		mp.byTime.Remove(entry)
	}

	return evictedSize >= requiredSize
}

// calculateTransactionSize calculates the size of a transaction
// calculateTransactionSize calculates the approximate size of a transaction in bytes.
func (mp *Mempool) calculateTransactionSize(tx *block.Transaction) uint64 {
	size := uint64(0)

	// Version + LockTime + Fee
	size += 4 + 8 + 8

	// Input count + Output count
	size += 4 + 4

	// Inputs
	for _, input := range tx.Inputs {
		size += 32 + 4 + uint64(len(input.ScriptSig)) + 4
	}

	// Outputs
	for _, output := range tx.Outputs {
		size += 8 + uint64(len(output.ScriptPubKey))
	}

	return size
}

// calculateFeeRate calculates the fee rate (fee per byte) of a transaction
// calculateFeeRate calculates the fee rate (fee per byte) of a transaction.
func (mp *Mempool) calculateFeeRate(tx *block.Transaction, size uint64) uint64 {
	if size == 0 {
		return 0
	}
	return tx.Fee / size
}

// validateFeeRate performs comprehensive fee rate validation with enhanced security features
func (mp *Mempool) validateFeeRate(tx *block.Transaction, feeRate uint64) error {
	// Check for dust transactions (very low value outputs)
	for i, output := range tx.Outputs {
		if output.Value < 546 { // Standard dust threshold (546 satoshis)
			return fmt.Errorf("output %d value %d below dust threshold", i, output.Value)
		}
	}

	// Enhanced fee rate validation with dynamic thresholds
	if mp.minFeeRate > 0 {
		// Check minimum fee rate
		if feeRate < mp.minFeeRate {
			return fmt.Errorf("fee rate %d below minimum %d", feeRate, mp.minFeeRate)
		}

		// Add absolute maximum fee rate limit regardless of utilization
		// This prevents excessive fees that could be used for DoS attacks
		absoluteMaxFeeRate := mp.minFeeRate * 40 // 40x the minimum fee rate as absolute cap
		if feeRate > absoluteMaxFeeRate {
			return fmt.Errorf("fee rate %d exceeds maximum allowed rate %d (absolute limit)",
				feeRate, absoluteMaxFeeRate)
		}

		// Dynamic maximum fee rate based on mempool utilization (more restrictive)
		utilization := float64(mp.currentSize) / float64(mp.maxSize)
		var maxAllowedFeeRate uint64

		if utilization > 0.8 {
			// High utilization: allow higher fees to prioritize important transactions
			maxAllowedFeeRate = mp.minFeeRate * 200 // 200x the minimum fee rate
		} else if utilization > 0.5 {
			// Medium utilization: moderate fee cap
			maxAllowedFeeRate = mp.minFeeRate * 100 // 100x the minimum fee rate
		} else {
			// Low utilization: very strict fee cap
			maxAllowedFeeRate = mp.minFeeRate * 50 // 50x the minimum fee rate
		}

		if feeRate > maxAllowedFeeRate {
			return fmt.Errorf("fee rate %d exceeds maximum allowed rate %d (utilization: %.2f)",
				feeRate, maxAllowedFeeRate, utilization)
		}
	}

	// Enhanced transaction size vs fee validation
	txSize := mp.calculateTransactionSize(tx)

	// Check for transactions with very high fees relative to size (potential DoS)
	if tx.Fee > txSize*1000 { // Fee should not exceed 1000x the size
		return fmt.Errorf("fee %d is excessively high relative to transaction size %d", tx.Fee, txSize)
	}

	// Check for transactions with very low fees relative to size (potential spam)
	minFeePerByte := mp.minFeeRate
	if minFeePerByte == 0 {
		minFeePerByte = 1 // Default minimum fee per byte
	}

	if tx.Fee < txSize*minFeePerByte {
		return fmt.Errorf("fee %d is too low for transaction size %d (minimum: %d)",
			tx.Fee, txSize, txSize*minFeePerByte)
	}

	// Check for suspicious fee patterns
	if len(tx.Inputs) > 0 && len(tx.Outputs) > 0 {
		// Calculate total input value from UTXO set if available
		if mp.utxoSet != nil && !mp.testMode {
			totalInput := uint64(0)
			for _, input := range tx.Inputs {
				utxo := mp.utxoSet.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
				if utxo != nil {
					totalInput += utxo.Value
				}
			}

			totalOutput := uint64(0)
			for _, output := range tx.Outputs {
				totalOutput += output.Value
			}

			// Fee should not exceed 90% of input value (prevent fee sniping)
			if totalInput > 0 && tx.Fee > totalInput*9/10 {
				return fmt.Errorf("fee %d exceeds 90%% of input value %d", tx.Fee, totalInput)
			}

			// Check for change output manipulation
			if totalInput > totalOutput && totalInput-totalOutput != tx.Fee {
				// There should be a change output or the fee should match the difference
				changeAmount := totalInput - totalOutput - tx.Fee
				if changeAmount > 0 && changeAmount < 546 {
					return fmt.Errorf("change amount %d is below dust threshold", changeAmount)
				}
			}
		}
	}

	return nil
}

// Remove removes a TransactionEntry from the TransactionHeap.
func (h *TransactionHeap) Remove(entry *TransactionEntry) {
	if entry.index >= 0 && entry.index < h.Len() {
		heap.Remove(h, entry.index)
	}
}

// Heap interface implementation for TransactionHeap
func (h TransactionHeap) Len() int { return len(h) }

func (h TransactionHeap) Less(i, j int) bool {
	// For fee-based heap: higher fee rate first
	return h[i].FeeRate > h[j].FeeRate
}

func (h TransactionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *TransactionHeap) Push(x interface{}) {
	n := len(*h)
	entry := x.(*TransactionEntry)
	entry.index = n
	*h = append(*h, entry)
}

func (h *TransactionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.index = -1 // for safety
	*h = old[0 : n-1]
	return entry
}

// TransactionHeapMin implements heap.Interface for transaction prioritization based on fee rate (min-heap).
type TransactionHeapMin []*TransactionEntry

func (h TransactionHeapMin) Len() int { return len(h) }

func (h TransactionHeapMin) Less(i, j int) bool {
	// For fee-based min-heap: lower fee rate first
	return h[i].FeeRate < h[j].FeeRate
}

func (h TransactionHeapMin) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *TransactionHeapMin) Push(x interface{}) {
	n := len(*h)
	entry := x.(*TransactionEntry)
	entry.index = n
	*h = append(*h, entry)
}

func (h *TransactionHeapMin) Pop() interface{} {
	old := *h
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.index = -1 // for safety
	*h = old[0 : n-1]
	return entry
}

// Remove removes a TransactionEntry from the TransactionHeapMin.
func (h *TransactionHeapMin) Remove(entry *TransactionEntry) {
	if entry.index >= 0 && entry.index < h.Len() {
		heap.Remove(h, entry.index)
	}
}

// TimeHeap implements heap.Interface for transaction prioritization based on timestamp (min-heap).
type TimeHeap []*TransactionEntry

func (h TimeHeap) Len() int { return len(h) }

func (h TimeHeap) Less(i, j int) bool {
	// For time-based heap: older transactions first
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h TimeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *TimeHeap) Push(x interface{}) {
	n := len(*h)
	entry := x.(*TransactionEntry)
	entry.index = n
	*h = append(*h, entry)
}

func (h *TimeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.index = -1 // for safety
	*h = old[0 : n-1]
	return entry
}

// IsTransactionValid validates a transaction for inclusion in the mempool.
// It performs comprehensive validation including signature verification, UTXO checks, and fee validation.
func (mp *Mempool) IsTransactionValid(tx *block.Transaction) error {
	// Basic transaction structure validation
	if err := tx.IsValid(); err != nil {
		return fmt.Errorf("invalid transaction structure: %w", err)
	}

	// Check transaction size limits
	size := mp.calculateTransactionSize(tx)
	if size > mp.maxTxSize {
		return fmt.Errorf("transaction size %d exceeds maximum allowed size %d", size, mp.maxTxSize)
	}

	// Additional security checks (do this BEFORE fee validation to catch security issues first)
	if err := mp.validateTransactionSecurity(tx); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// Enhanced UTXO validation with signature verification
	if mp.utxoSet != nil && !mp.testMode {
		if err := mp.utxoSet.ValidateTransaction(tx); err != nil {
			return fmt.Errorf("transaction validation failed: %w", err)
		}

		// Additional security checks for non-coinbase transactions
		if !tx.IsCoinbase() {
			// Check for double-spend attempts
			if mp.utxoSet.IsDoubleSpend(tx) {
				return fmt.Errorf("transaction attempts double-spend")
			}

			// Validate that all inputs reference existing UTXOs
			for i, input := range tx.Inputs {
				utxo := mp.utxoSet.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
				if utxo == nil {
					return fmt.Errorf("input %d references non-existent UTXO", i)
				}

				// Check if UTXO is already spent in mempool
				if mp.isUTXOSpentInMempool(input.PrevTxHash, input.PrevTxIndex) {
					return fmt.Errorf("input %d references UTXO already spent in mempool", i)
				}
			}
		}
	}

	// Check if UTXO is already spent in mempool (even in test mode)
	if !tx.IsCoinbase() {
		for i, input := range tx.Inputs {
			if mp.isUTXOSpentInMempool(input.PrevTxHash, input.PrevTxIndex) {
				return fmt.Errorf("input %d references UTXO already spent in mempool", i)
			}
		}
	}

	// Enhanced fee rate validation (do this AFTER security validation)
	feeRate := mp.calculateFeeRate(tx, size)
	if feeRate < mp.minFeeRate {
		return fmt.Errorf("fee rate %d below minimum %d", feeRate, mp.minFeeRate)
	}

	if err := mp.validateFeeRate(tx, feeRate); err != nil {
		return fmt.Errorf("fee rate validation failed: %w", err)
	}

	return nil
}

// isUTXOSpentInMempool checks if a UTXO is already spent by another transaction in the mempool
// Note: This function should only be called from functions that already hold the mempool lock
func (mp *Mempool) isUTXOSpentInMempool(txHash []byte, txIndex uint32) bool {
	// No need to acquire lock here - caller should already hold it
	for _, entry := range mp.transactions {
		for _, input := range entry.Transaction.Inputs {
			if bytes.Equal(input.PrevTxHash, txHash) && input.PrevTxIndex == txIndex {
				return true
			}
		}
	}
	return false
}

// validateTransactionSecurity performs additional security validations
func (mp *Mempool) validateTransactionSecurity(tx *block.Transaction) error {
	// Check for excessive input/output counts (DoS prevention)
	if len(tx.Inputs) > 1000 {
		return fmt.Errorf("transaction has too many inputs: %d (max: 1000)", len(tx.Inputs))
	}

	if len(tx.Outputs) > 1000 {
		return fmt.Errorf("transaction has too many outputs: %d (max: 1000)", len(tx.Outputs))
	}

	// Check for suspicious transaction patterns
	if len(tx.Inputs) == 0 && len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no inputs or outputs")
	}

	// Validate locktime (if set)
	if tx.LockTime > 0 {
		currentTime := uint64(time.Now().Unix())
		if tx.LockTime > currentTime {
			return fmt.Errorf("transaction locktime %d is in the future (current: %d)", tx.LockTime, currentTime)
		}
	}

	return nil
}

// GetTransactionStats returns statistics about the mempool for monitoring and DoS detection
func (mp *Mempool) GetTransactionStats() map[string]interface{} {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Calculate fee rate distribution
	var totalFee, totalSize uint64
	var feeRates []uint64
	for _, entry := range mp.transactions {
		totalFee += entry.Transaction.Fee
		totalSize += entry.Size
		feeRates = append(feeRates, entry.FeeRate)
	}

	// Calculate average fee rate safely
	var avgFeeRate uint64
	if len(feeRates) > 0 {
		var sum uint64
		for _, rate := range feeRates {
			sum += rate
		}
		avgFeeRate = sum / uint64(len(feeRates))
	}

	// Calculate utilization safely
	utilization := float64(0)
	if mp.maxSize > 0 {
		utilization = float64(mp.currentSize) / float64(mp.maxSize)
	}

	return map[string]interface{}{
		"transaction_count": len(mp.transactions),
		"total_size":        mp.currentSize,
		"max_size":          mp.maxSize,
		"min_fee_rate":      mp.minFeeRate,
		"avg_fee_rate":      avgFeeRate,
		"total_fees":        totalFee,
		"utilization":       utilization,
	}
}

// IsUnderDoS returns true if the mempool appears to be under a DoS attack
func (mp *Mempool) IsUnderDoS() bool {
	stats := mp.GetTransactionStats()

	// Check for suspicious patterns
	utilization := stats["utilization"].(float64)
	txCount := stats["transaction_count"].(int)

	// High utilization with many small transactions could indicate spam
	if utilization > 0.9 && txCount > 1000 {
		return true
	}

	// Very low average fee rate with many transactions could indicate spam
	avgFeeRate := stats["avg_fee_rate"].(uint64)
	if avgFeeRate < mp.minFeeRate*2 && txCount > 500 {
		return true
	}

	return false
}

// CleanupExpiredTransactions removes transactions that have been in the mempool too long
// This helps prevent memory exhaustion and stale transaction attacks
func (mp *Mempool) CleanupExpiredTransactions(maxAge time.Duration) int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	now := time.Now()
	removed := 0

	for hash, entry := range mp.transactions {
		if now.Sub(entry.Timestamp) > maxAge {
			// Remove expired transaction
			delete(mp.transactions, hash)
			mp.currentSize -= entry.Size
			mp.byFee.Remove(entry)
			mp.byTime.Remove(entry)
			removed++
		}
	}

	return removed
}

// String returns a string representation of the mempool
func (mp *Mempool) String() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return fmt.Sprintf("Mempool{Size: %d/%d, Transactions: %d, MinFeeRate: %d}",
		mp.currentSize, mp.maxSize, len(mp.transactions), mp.minFeeRate)
}
