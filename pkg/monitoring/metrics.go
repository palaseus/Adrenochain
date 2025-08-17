package monitoring

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics represents a collection of blockchain metrics
type Metrics struct {
	mu sync.RWMutex

	// Blockchain metrics
	blockHeight     int64
	totalBlocks     int64
	totalTxns       int64
	pendingTxns     int64
	chainDifficulty float64

	// Network metrics
	connectedPeers int64
	totalPeers     int64
	networkLatency int64 // in milliseconds

	// Mining metrics
	hashRate      int64 // hashes per second
	blocksMined   int64
	miningEnabled bool

	// Performance metrics
	blockProcessingTime int64 // in milliseconds
	txnProcessingTime   int64 // in milliseconds
	memoryUsage         int64 // in bytes

	// Error metrics
	totalErrors      int64
	validationErrors int64
	networkErrors    int64

	// Timestamps
	lastBlockTime time.Time
	lastSyncTime  time.Time
	startTime     time.Time

	// Additional blockchain metrics
	utxoCount      int64
	chainSize      int64 // in bytes
	orphanedBlocks int64
	rejectedBlocks int64
	rejectedTxns   int64
	avgBlockTime   int64 // in seconds
	avgTxnPerBlock float64
	avgBlockSize   int64 // in bytes
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// UpdateBlockHeight updates the current block height
func (m *Metrics) UpdateBlockHeight(height int64) {
	atomic.StoreInt64(&m.blockHeight, height)
}

// UpdateTotalBlocks updates the total number of blocks
func (m *Metrics) UpdateTotalBlocks(count int64) {
	atomic.StoreInt64(&m.totalBlocks, count)
}

// UpdateTotalTxns updates the total number of transactions
func (m *Metrics) UpdateTotalTxns(count int64) {
	atomic.StoreInt64(&m.totalTxns, count)
}

// UpdatePendingTxns updates the number of pending transactions
func (m *Metrics) UpdatePendingTxns(count int64) {
	atomic.StoreInt64(&m.pendingTxns, count)
}

// UpdateChainDifficulty updates the current chain difficulty
func (m *Metrics) UpdateChainDifficulty(difficulty float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chainDifficulty = difficulty
}

// UpdateConnectedPeers updates the number of connected peers
func (m *Metrics) UpdateConnectedPeers(count int64) {
	atomic.StoreInt64(&m.connectedPeers, count)
}

// UpdateTotalPeers updates the total number of known peers
func (m *Metrics) UpdateTotalPeers(count int64) {
	atomic.StoreInt64(&m.totalPeers, count)
}

// UpdateNetworkLatency updates the average network latency
func (m *Metrics) UpdateNetworkLatency(latency int64) {
	atomic.StoreInt64(&m.networkLatency, latency)
}

// UpdateHashRate updates the current hash rate
func (m *Metrics) UpdateHashRate(rate int64) {
	atomic.StoreInt64(&m.hashRate, rate)
}

// UpdateBlocksMined updates the number of blocks mined
func (m *Metrics) UpdateBlocksMined(count int64) {
	atomic.StoreInt64(&m.blocksMined, count)
}

// SetMiningEnabled sets whether mining is enabled
func (m *Metrics) SetMiningEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.miningEnabled = enabled
}

// UpdateBlockProcessingTime updates the average block processing time
func (m *Metrics) UpdateBlockProcessingTime(duration time.Duration) {
	atomic.StoreInt64(&m.blockProcessingTime, int64(duration.Milliseconds()))
}

// UpdateTxnProcessingTime updates the average transaction processing time
func (m *Metrics) UpdateTxnProcessingTime(duration time.Duration) {
	atomic.StoreInt64(&m.txnProcessingTime, int64(duration.Milliseconds()))
}

// UpdateMemoryUsage updates the current memory usage
func (m *Metrics) UpdateMemoryUsage(bytes int64) {
	atomic.StoreInt64(&m.memoryUsage, bytes)
}

// IncrementErrors increments the total error count
func (m *Metrics) IncrementErrors() {
	atomic.AddInt64(&m.totalErrors, 1)
}

// IncrementValidationErrors increments the validation error count
func (m *Metrics) IncrementValidationErrors() {
	atomic.AddInt64(&m.validationErrors, 1)
}

// IncrementNetworkErrors increments the network error count
func (m *Metrics) IncrementNetworkErrors() {
	atomic.AddInt64(&m.networkErrors, 1)
}

// UpdateLastBlockTime updates the timestamp of the last block
func (m *Metrics) UpdateLastBlockTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastBlockTime = t
}

// UpdateLastSyncTime updates the timestamp of the last sync
func (m *Metrics) UpdateLastSyncTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastSyncTime = t
}

// UpdateUTXOCount updates the UTXO count
func (m *Metrics) UpdateUTXOCount(count int64) {
	atomic.StoreInt64(&m.utxoCount, count)
}

// UpdateChainSize updates the chain size in bytes
func (m *Metrics) UpdateChainSize(size int64) {
	atomic.StoreInt64(&m.chainSize, size)
}

// IncrementOrphanedBlocks increments the orphaned blocks count
func (m *Metrics) IncrementOrphanedBlocks() {
	atomic.AddInt64(&m.orphanedBlocks, 1)
}

// IncrementRejectedBlocks increments the rejected blocks count
func (m *Metrics) IncrementRejectedBlocks() {
	atomic.AddInt64(&m.rejectedBlocks, 1)
}

// IncrementRejectedTxns increments the rejected transactions count
func (m *Metrics) IncrementRejectedTxns() {
	atomic.AddInt64(&m.rejectedTxns, 1)
}

// UpdateAvgBlockTime updates the average block time
func (m *Metrics) UpdateAvgBlockTime(seconds int64) {
	atomic.StoreInt64(&m.avgBlockTime, seconds)
}

// UpdateAvgTxnPerBlock updates the average transactions per block
func (m *Metrics) UpdateAvgTxnPerBlock(avg float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.avgTxnPerBlock = avg
}

// UpdateAvgBlockSize updates the average block size
func (m *Metrics) UpdateAvgBlockSize(size int64) {
	atomic.StoreInt64(&m.avgBlockSize, size)
}

// GetMetrics returns a copy of all current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	return map[string]interface{}{
		"blockchain": map[string]interface{}{
			"block_height":           atomic.LoadInt64(&m.blockHeight),
			"total_blocks":           atomic.LoadInt64(&m.totalBlocks),
			"total_transactions":     atomic.LoadInt64(&m.totalTxns),
			"pending_transactions":   atomic.LoadInt64(&m.pendingTxns),
			"chain_difficulty":       m.chainDifficulty,
			"last_block_time":        m.lastBlockTime,
			"utxo_count":             atomic.LoadInt64(&m.utxoCount),
			"chain_size_bytes":       atomic.LoadInt64(&m.chainSize),
			"orphaned_blocks":        atomic.LoadInt64(&m.orphanedBlocks),
			"rejected_blocks":        atomic.LoadInt64(&m.rejectedBlocks),
			"rejected_transactions":  atomic.LoadInt64(&m.rejectedTxns),
			"avg_block_time_seconds": atomic.LoadInt64(&m.avgBlockTime),
			"avg_txn_per_block":      m.avgTxnPerBlock,
			"avg_block_size_bytes":   atomic.LoadInt64(&m.avgBlockSize),
		},
		"network": map[string]interface{}{
			"connected_peers": atomic.LoadInt64(&m.connectedPeers),
			"total_peers":     atomic.LoadInt64(&m.totalPeers),
			"network_latency": atomic.LoadInt64(&m.networkLatency),
			"last_sync_time":  m.lastSyncTime,
		},
		"mining": map[string]interface{}{
			"hash_rate":      atomic.LoadInt64(&m.hashRate),
			"blocks_mined":   atomic.LoadInt64(&m.blocksMined),
			"mining_enabled": m.miningEnabled,
		},
		"performance": map[string]interface{}{
			"block_processing_time": atomic.LoadInt64(&m.blockProcessingTime),
			"txn_processing_time":   atomic.LoadInt64(&m.txnProcessingTime),
			"memory_usage":          atomic.LoadInt64(&m.memoryUsage),
		},
		"errors": map[string]interface{}{
			"total_errors":      atomic.LoadInt64(&m.totalErrors),
			"validation_errors": atomic.LoadInt64(&m.validationErrors),
			"network_errors":    atomic.LoadInt64(&m.networkErrors),
		},
		"system": map[string]interface{}{
			"uptime":     uptime.String(),
			"start_time": m.startTime,
		},
	}
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (m *Metrics) GetPrometheusMetrics() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime).Seconds()

	var prometheus string

	// Blockchain metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_block_height Current blockchain height\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_block_height gauge\n")
	prometheus += fmt.Sprintf("adrenochain_block_height %d\n", atomic.LoadInt64(&m.blockHeight))

	prometheus += fmt.Sprintf("# HELP adrenochain_total_blocks Total number of blocks\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_total_blocks counter\n")
	prometheus += fmt.Sprintf("adrenochain_total_blocks %d\n", atomic.LoadInt64(&m.totalBlocks))

	prometheus += fmt.Sprintf("# HELP adrenochain_total_transactions Total number of transactions\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_total_transactions counter\n")
	prometheus += fmt.Sprintf("adrenochain_total_transactions %d\n", atomic.LoadInt64(&m.totalTxns))

	prometheus += fmt.Sprintf("# HELP adrenochain_pending_transactions Number of pending transactions\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_pending_transactions gauge\n")
	prometheus += fmt.Sprintf("adrenochain_pending_transactions %d\n", atomic.LoadInt64(&m.pendingTxns))

	prometheus += fmt.Sprintf("# HELP adrenochain_chain_difficulty Current chain difficulty\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_chain_difficulty gauge\n")
	prometheus += fmt.Sprintf("adrenochain_chain_difficulty %f\n", m.chainDifficulty)

	// Network metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_connected_peers Number of connected peers\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_connected_peers gauge\n")
	prometheus += fmt.Sprintf("adrenochain_connected_peers %d\n", atomic.LoadInt64(&m.connectedPeers))

	prometheus += fmt.Sprintf("# HELP adrenochain_total_peers Total number of known peers\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_total_peers gauge\n")
	prometheus += fmt.Sprintf("adrenochain_total_peers %d\n", atomic.LoadInt64(&m.totalPeers))

	// Mining metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_hash_rate Current hash rate\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_hash_rate gauge\n")
	prometheus += fmt.Sprintf("adrenochain_hash_rate %d\n", atomic.LoadInt64(&m.hashRate))

	prometheus += fmt.Sprintf("# HELP adrenochain_blocks_mined Total blocks mined\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_blocks_mined counter\n")
	prometheus += fmt.Sprintf("adrenochain_blocks_mined %d\n", atomic.LoadInt64(&m.blocksMined))

	// Performance metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_memory_usage_bytes Current memory usage in bytes\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_memory_usage_bytes gauge\n")
	prometheus += fmt.Sprintf("adrenochain_memory_usage_bytes %d\n", atomic.LoadInt64(&m.memoryUsage))

	// Error metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_total_errors Total number of errors\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_total_errors counter\n")
	prometheus += fmt.Sprintf("adrenochain_total_errors %d\n", atomic.LoadInt64(&m.totalErrors))

	// System metrics
	prometheus += fmt.Sprintf("# HELP adrenochain_uptime_seconds Node uptime in seconds\n")
	prometheus += fmt.Sprintf("# TYPE adrenochain_uptime_seconds gauge\n")
	prometheus += fmt.Sprintf("adrenochain_uptime_seconds %f\n", uptime)

	return prometheus
}

// Reset resets all metrics to zero
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.blockHeight, 0)
	atomic.StoreInt64(&m.totalBlocks, 0)
	atomic.StoreInt64(&m.totalTxns, 0)
	atomic.StoreInt64(&m.pendingTxns, 0)
	atomic.StoreInt64(&m.connectedPeers, 0)
	atomic.StoreInt64(&m.totalPeers, 0)
	atomic.StoreInt64(&m.networkLatency, 0)
	atomic.StoreInt64(&m.hashRate, 0)
	atomic.StoreInt64(&m.blocksMined, 0)
	atomic.StoreInt64(&m.blockProcessingTime, 0)
	atomic.StoreInt64(&m.txnProcessingTime, 0)
	atomic.StoreInt64(&m.memoryUsage, 0)
	atomic.StoreInt64(&m.totalErrors, 0)
	atomic.StoreInt64(&m.validationErrors, 0)
	atomic.StoreInt64(&m.networkErrors, 0)
	atomic.StoreInt64(&m.utxoCount, 0)
	atomic.StoreInt64(&m.chainSize, 0)
	atomic.StoreInt64(&m.orphanedBlocks, 0)
	atomic.StoreInt64(&m.rejectedBlocks, 0)
	atomic.StoreInt64(&m.rejectedTxns, 0)
	atomic.StoreInt64(&m.avgBlockTime, 0)
	atomic.StoreInt64(&m.avgBlockSize, 0)

	m.chainDifficulty = 0
	m.miningEnabled = false
	m.lastBlockTime = time.Time{}
	m.lastSyncTime = time.Time{}
	m.avgTxnPerBlock = 0
	m.startTime = time.Now()
}
