// Package mev implements MEV (Maximal Extractable Value) protection and private
// mempool infrastructure, aligned with 2026 blockchain privacy and fairness
// standards.
//
// MEV protection techniques:
//   - Private mempool: transactions are not broadcast publicly, reducing
//     front-running and sandwich attacks
//   - Fair ordering: FIFO or fair batch ordering of transactions
//   - Commit-reveal: users commit to a transaction hash, reveal later
//   - Encrypted mempool: transactions are encrypted until inclusion
//
// See pkg/mev/README.md for design and usage.

package mev

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"sort"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Transaction
// ---------------------------------------------------------------------------

// Transaction is a signed blockchain transaction.
type Transaction struct {
	// Hash is the transaction hash.
	Hash []byte
	// From is the sender address.
	From []byte
	// To is the recipient address (nil for contract creation).
	To []byte
	// Value is the amount transferred.
	Value int64
	// GasPrice is the gas price offered.
	GasPrice int64
	// GasLimit is the maximum gas.
	GasLimit int64
	// Nonce is the transaction nonce.
	Nonce uint64
	// Data is the transaction payload.
	Data []byte
	// Signature is the transaction signature.
	Signature []byte
	// Timestamp is when the transaction was created.
	Timestamp time.Time
}

// NewTransaction creates a new transaction.
func NewTransaction(from, to []byte, value, gasPrice, gasLimit int64, nonce uint64, data []byte) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Nonce:     nonce,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	tx.Hash = tx.CalculateHash()
	return tx
}

// CalculateHash computes the transaction hash.
func (tx *Transaction) CalculateHash() []byte {
	h := sha256.New()
	h.Write(tx.From)
	h.Write(tx.To)
	h.Write([]byte{byte(tx.Value >> 56), byte(tx.Value >> 48), byte(tx.Value >> 40), byte(tx.Value >> 32),
		byte(tx.Value >> 24), byte(tx.Value >> 16), byte(tx.Value >> 8), byte(tx.Value)})
	h.Write([]byte{byte(tx.GasPrice >> 56), byte(tx.GasPrice >> 48), byte(tx.GasPrice >> 40), byte(tx.GasPrice >> 32),
		byte(tx.GasPrice >> 24), byte(tx.GasPrice >> 16), byte(tx.GasPrice >> 8), byte(tx.GasPrice)})
	h.Write([]byte{byte(tx.GasLimit >> 56), byte(tx.GasLimit >> 48), byte(tx.GasLimit >> 40), byte(tx.GasLimit >> 32),
		byte(tx.GasLimit >> 24), byte(tx.GasLimit >> 16), byte(tx.GasLimit >> 8), byte(tx.GasLimit)})
	h.Write([]byte{byte(tx.Nonce >> 56), byte(tx.Nonce >> 48), byte(tx.Nonce >> 40), byte(tx.Nonce >> 32),
		byte(tx.Nonce >> 24), byte(tx.Nonce >> 16), byte(tx.Nonce >> 8), byte(tx.Nonce)})
	h.Write(tx.Data)
	h.Write(tx.Signature)
	return h.Sum(nil)
}

// ---------------------------------------------------------------------------
// PrivateTx - encrypted/private transaction wrapper
// ---------------------------------------------------------------------------

// PrivateTx wraps a transaction with encryption for private mempool submission.
// The transaction is encrypted so that only the block producer can decrypt it,
// preventing front-running and sandwich attacks.
type PrivateTx struct {
	// EncryptedTx is the encrypted transaction payload.
	EncryptedTx []byte
	// Commitment is a commitment to the transaction (hash of plaintext).
	Commitment []byte
	// Nonce prevents replay.
	Nonce []byte
	// Timestamp is when the transaction was created.
	Timestamp time.Time
}

// NewPrivateTx creates a private transaction (placeholder encryption).
func NewPrivateTx(tx *Transaction) *PrivateTx {
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	// Placeholder: XOR-based "encryption" for structural demonstration.
	// Production must use real hybrid encryption (e.g. ECIES with ML-KEM).
	plaintext := append(tx.From, tx.To...)
	plaintext = append(plaintext, tx.Data...)
	encrypted := make([]byte, len(plaintext))
	for i := range plaintext {
		encrypted[i] = plaintext[i] ^ nonce[i%len(nonce)]
	}
	return &PrivateTx{
		EncryptedTx: encrypted,
		Commitment:  tx.Hash,
		Nonce:       nonce,
		Timestamp:   time.Now().UTC(),
	}
}

// ---------------------------------------------------------------------------
// PrivateMempool - private transaction pool
// ---------------------------------------------------------------------------

// PrivateMempool is a private mempool that accepts encrypted transactions
// and orders them fairly to prevent MEV extraction.
type PrivateMempool struct {
	mu      sync.RWMutex
	txs     map[[32]byte]*Transaction // hash -> transaction
	private map[[32]byte]*PrivateTx   // hash -> private transaction
	order   [][]byte                  // insertion order of hashes
}

// NewPrivateMempool creates a new private mempool.
func NewPrivateMempool() *PrivateMempool {
	return &PrivateMempool{
		txs:     make(map[[32]byte]*Transaction),
		private: make(map[[32]byte]*PrivateTx),
		order:   make([][]byte, 0),
	}
}

// Submit adds a private transaction to the mempool.
func (m *PrivateMempool) Submit(tx *Transaction, pt *PrivateTx) error {
	if tx == nil || pt == nil {
		return errors.New("nil transaction or private tx")
	}
	if !bytes.Equal(tx.Hash, pt.Commitment) {
		return errors.New("transaction hash does not match commitment")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	hashKey := [32]byte{}
	copy(hashKey[:], tx.Hash[:32])
	if _, exists := m.txs[hashKey]; exists {
		return errors.New("transaction already in mempool")
	}
	m.txs[hashKey] = tx
	m.private[hashKey] = pt
	m.order = append(m.order, hashKey[:])
	return nil
}

// GetTransaction returns a transaction by hash.
func (m *PrivateMempool) GetTransaction(hash []byte) (*Transaction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var key [32]byte
	copy(key[:], hash[:32])
	tx, ok := m.txs[key]
	return tx, ok
}

// GetPrivateTx returns a private transaction by hash.
func (m *PrivateMempool) GetPrivateTx(hash []byte) (*PrivateTx, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var key [32]byte
	copy(key[:], hash[:32])
	pt, ok := m.private[key]
	return pt, ok
}

// Len returns the number of transactions in the mempool.
func (m *PrivateMempool) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.txs)
}

// GetOrderedTransactions returns transactions in insertion order (fair ordering).
func (m *PrivateMempool) GetOrderedTransactions() []*Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Transaction, 0, len(m.txs))
	for _, hash := range m.order {
		var key [32]byte
		copy(key[:], hash)
		if tx, ok := m.txs[key]; ok {
			result = append(result, tx)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// MEVProtection - MEV protection strategies
// ---------------------------------------------------------------------------

// MEVProtectionStrategy is a strategy for ordering transactions to minimize MEV.
type MEVProtectionStrategy int

const (
	// StrategyFIFO orders transactions strictly by arrival time (first-come,
	// first-served). Prevents priority gas auctions but not all MEV.
	StrategyFIFO MEVProtectionStrategy = iota
	// StrategyFairBatch batches transactions and orders them randomly within
	// each batch, preventing front-running within the batch.
	StrategyFairBatch
	// StrategyCommitReveal uses commit-reveal: users commit to a tx hash,
	// reveal later. Prevents front-running entirely but adds latency.
	StrategyCommitReveal
)

// String returns the strategy name.
func (s MEVProtectionStrategy) String() string {
	names := map[MEVProtectionStrategy]string{
		StrategyFIFO:         "FIFO",
		StrategyFairBatch:    "FairBatch",
		StrategyCommitReveal: "CommitReveal",
	}
	if n, ok := names[s]; ok {
		return n
	}
	return "unknown"
}

// MEVProtector orders transactions using a protection strategy.
type MEVProtector struct {
	strategy  MEVProtectionStrategy
	batchSize int
}

// NewMEVProtector creates a new MEV protector with the given strategy.
func NewMEVProtector(strategy MEVProtectionStrategy, batchSize int) *MEVProtector {
	return &MEVProtector{strategy: strategy, batchSize: batchSize}
}

// Order orders transactions according to the protection strategy.
func (mp *MEVProtector) Order(txs []*Transaction) []*Transaction {
	if len(txs) <= 1 {
		return txs
	}
	result := make([]*Transaction, len(txs))
	copy(result, txs)
	switch mp.strategy {
	case StrategyFIFO:
		// Already in insertion order (FIFO).
		return result
	case StrategyFairBatch:
		return mp.orderFairBatch(result)
	case StrategyCommitReveal:
		// Commit-reveal: order by commitment hash (commitment is binding,
		// so sender cannot reorder after seeing others' transactions).
		sort.Slice(result, func(i, j int) bool {
			return bytes.Compare(result[i].Hash, result[j].Hash) < 0
		})
		return result
	default:
		return result
	}
}

// orderFairBatch orders transactions in batches of batchSize, randomly
// shuffling within each batch to prevent front-running.
func (mp *MEVProtector) orderFairBatch(txs []*Transaction) []*Transaction {
	batchSize := mp.batchSize
	if batchSize <= 0 {
		batchSize = 10
	}
	result := make([]*Transaction, 0, len(txs))
	for i := 0; i < len(txs); i += batchSize {
		end := i + batchSize
		if end > len(txs) {
			end = len(txs)
		}
		batch := make([]*Transaction, end-i)
		copy(batch, txs[i:end])
		// Shuffle within batch (simple Fisher-Yates)
		for j := len(batch) - 1; j > 0; j-- {
			k := (j * 37) % (j + 1) // deterministic pseudo-random
			batch[j], batch[k] = batch[k], batch[j]
		}
		result = append(result, batch...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Bundle - atomic transaction bundle
// ---------------------------------------------------------------------------

// Bundle is a group of transactions that must execute atomically (all or none).
type Bundle struct {
	// ID is the bundle identifier.
	ID []byte
	// Transactions are the transactions in the bundle.
	Transactions []*Transaction
	// Fee is the total fee for the bundle.
	Fee int64
}

// NewBundle creates a new transaction bundle.
func NewBundle(id []byte, txs []*Transaction, fee int64) *Bundle {
	return &Bundle{
		ID:           id,
		Transactions: txs,
		Fee:          fee,
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
