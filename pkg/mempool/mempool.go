package mempool

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// Mempool represents the transaction memory pool
type Mempool struct {
	mu            sync.RWMutex
	transactions  map[string]*TransactionEntry // tx hash -> entry
	byFee         *TransactionHeap            // Priority queue by fee
	byTime        *TransactionHeap            // Priority queue by time
	maxSize       uint64
	currentSize   uint64
	minFeeRate    uint64 // minimum fee per byte
}

// TransactionEntry wraps a transaction with metadata
type TransactionEntry struct {
	Transaction *block.Transaction
	FeeRate     uint64 // fee per byte
	Size        uint64
	Timestamp   time.Time
	index       int // for heap operations
}

// TransactionHeap implements heap.Interface for transaction prioritization
type TransactionHeap []*TransactionEntry

// MempoolConfig holds configuration for the mempool
type MempoolConfig struct {
	MaxSize    uint64
	MinFeeRate uint64
}

// DefaultMempoolConfig returns the default mempool configuration
func DefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		MaxSize:    100000, // 100KB
		MinFeeRate: 1,      // 1 unit per byte
	}
}

// NewMempool creates a new transaction mempool
func NewMempool(config *MempoolConfig) *Mempool {
	mp := &Mempool{
		transactions: make(map[string]*TransactionEntry),
		byFee:        &TransactionHeap{},
		byTime:       &TransactionHeap{},
		maxSize:      config.MaxSize,
		minFeeRate:   config.MinFeeRate,
	}
	
	heap.Init(mp.byFee)
	heap.Init(mp.byTime)
	
	return mp
}

// AddTransaction adds a transaction to the mempool
func (mp *Mempool) AddTransaction(tx *block.Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	// Check if transaction already exists
	txHash := string(tx.Hash)
	if _, exists := mp.transactions[txHash]; exists {
		return fmt.Errorf("transaction already in mempool")
	}
	
	// Calculate transaction size and fee rate
	size := mp.calculateTransactionSize(tx)
	feeRate := mp.calculateFeeRate(tx, size)
	
	// Check minimum fee rate
	if feeRate < mp.minFeeRate {
		return fmt.Errorf("fee rate %d below minimum %d", feeRate, mp.minFeeRate)
	}
	
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

// RemoveTransaction removes a transaction from the mempool
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

// GetTransaction returns a transaction by its hash
func (mp *Mempool) GetTransaction(txHash []byte) *block.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	entry, exists := mp.transactions[string(txHash)]
	if !exists {
		return nil
	}
	
	return entry.Transaction
}

// GetTransactionsForBlock returns transactions for block assembly
func (mp *Mempool) GetTransactionsForBlock(maxSize uint64) []*block.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	var transactions []*block.Transaction
	currentSize := uint64(0)
	
	// Create a copy of the fee queue to avoid modifying the original
	feeQueue := make(TransactionHeap, mp.byFee.Len())
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

// GetSize returns the current mempool size
func (mp *Mempool) GetSize() uint64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	return mp.currentSize
}

// GetTransactionCount returns the number of transactions in the mempool
func (mp *Mempool) GetTransactionCount() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	return len(mp.transactions)
}

// Clear removes all transactions from the mempool
func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	mp.transactions = make(map[string]*TransactionEntry)
	mp.byFee = &TransactionHeap{}
	mp.byTime = &TransactionHeap{}
	mp.currentSize = 0
	
	heap.Init(mp.byFee)
	heap.Init(mp.byTime)
}

// evictLowFeeTransactions evicts low-fee transactions to make room for new ones
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
func (mp *Mempool) calculateFeeRate(tx *block.Transaction, size uint64) uint64 {
	if size == 0 {
		return 0
	}
	return tx.Fee / size
}

// Remove removes a transaction from the heap
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
	old[n-1] = nil  // avoid memory leak
	entry.index = -1 // for safety
	*h = old[0 : n-1]
	return entry
}

// TimeHeap implements heap.Interface for time-based prioritization
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
	old[n-1] = nil  // avoid memory leak
	entry.index = -1 // for safety
	*h = old[0 : n-1]
	return entry
}

// String returns a string representation of the mempool
func (mp *Mempool) String() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	return fmt.Sprintf("Mempool{Size: %d/%d, Transactions: %d, MinFeeRate: %d}", 
		mp.currentSize, mp.maxSize, len(mp.transactions), mp.minFeeRate)
} 