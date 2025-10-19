package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/wallet"
)

// ChainInterface defines the interface for blockchain operations
type ChainInterface interface {
	GetHeight() uint64
	GetBestBlock() *block.Block
	GetGenesisBlock() *block.Block
	GetBlock(hash []byte) *block.Block
	GetBlockByHeight(height uint64) *block.Block
	CalculateNextDifficulty() uint64
}

// WalletInterface defines the interface for wallet operations
type WalletInterface interface {
	GetBalance(address string) uint64
	GetAllAccounts() []*wallet.Account
}

// NetworkInterface defines the interface for network operations
type NetworkInterface interface {
	GetPeers() []string
	GetPeerCount() int
}

// MempoolInterface defines the interface for mempool operations
type MempoolInterface interface {
	GetPendingTransactions() []*block.Transaction
	GetPendingTransactionCount() int
	AddTransaction(tx *block.Transaction) error
	RemoveTransaction(txHash []byte) error
}

// Server represents the HTTP API server
type Server struct {
	router  *mux.Router
	chain   ChainInterface
	wallet  WalletInterface
	network NetworkInterface
	mempool MempoolInterface
	port    int

	// SECURITY FIX: Rate limiting
	rateLimiter *RateLimiter

	// Transaction index for O(1) lookup
	txIndex map[string]*TxIndexEntry
	txMutex sync.RWMutex
}

// TxIndexEntry represents a transaction index entry
type TxIndexEntry struct {
	BlockHeight uint64
	TxIndex     int
	BlockHash   []byte
}

// SECURITY FIX: RateLimiter implements rate limiting for API endpoints
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	window   time.Duration
	limit    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

// IsAllowed checks if a request is allowed based on rate limiting
func (rl *RateLimiter) IsAllowed(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if requests, exists := rl.requests[identifier]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[identifier] = validRequests
	}

	// Check if under limit
	if len(rl.requests[identifier]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[identifier] = append(rl.requests[identifier], now)
	return true
}

// ServerConfig holds configuration for the API server
type ServerConfig struct {
	Port    int
	Chain   ChainInterface
	Wallet  WalletInterface
	Network NetworkInterface
	Mempool MempoolInterface
}

// NewServer creates a new API server
func NewServer(config *ServerConfig) *Server {
	router := mux.NewRouter()
	server := &Server{
		router:  router,
		chain:   config.Chain,
		wallet:  config.Wallet,
		network: config.Network,
		mempool: config.Mempool,
		port:    config.Port,

		// SECURITY FIX: Initialize rate limiter
		rateLimiter: NewRateLimiter(1*time.Minute, 100), // 100 requests per minute

		// Initialize transaction index
		txIndex: make(map[string]*TxIndexEntry),
	}

	// SECURITY FIX: Add rate limiting middleware
	router.Use(server.rateLimitMiddleware)

	server.setupRoutes()

	// Build initial transaction index
	server.buildTransactionIndex()

	return server
}

// buildTransactionIndex builds the transaction index from the blockchain
func (s *Server) buildTransactionIndex() {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()

	if s.chain == nil {
		return // No chain available, skip indexing
	}

	height := s.chain.GetHeight()
	for h := uint64(0); h <= height; h++ {
		block := s.chain.GetBlockByHeight(h)
		if block == nil {
			continue
		}

		for i, tx := range block.Transactions {
			txHash := string(tx.Hash)
			s.txIndex[txHash] = &TxIndexEntry{
				BlockHeight: h,
				TxIndex:     i,
				BlockHash:   block.CalculateHash(),
			}
		}
	}
}

// addTransactionToIndex adds a transaction to the index
func (s *Server) addTransactionToIndex(tx *block.Transaction, blockHeight uint64, txIndex int, blockHash []byte) {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()

	txHash := string(tx.Hash)
	s.txIndex[txHash] = &TxIndexEntry{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
		BlockHash:   blockHash,
	}
}

// BuildTransactionIndex manually builds the transaction index (for testing)
func (s *Server) BuildTransactionIndex() {
	s.buildTransactionIndex()
}

// SECURITY FIX: rateLimitMiddleware implements rate limiting for all endpoints
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := s.getClientIP(r)

		// Check rate limit
		if !s.rateLimiter.IsAllowed(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SECURITY FIX: getClientIP extracts client IP from request
func (s *Server) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Metrics endpoint for Prometheus
	s.router.HandleFunc("/metrics", s.metricsHandler).Methods("GET")
	s.router.HandleFunc("/prometheus", s.prometheusHandler).Methods("GET")

	// Blockchain information
	s.router.HandleFunc("/api/v1/chain/info", s.getChainInfoHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/chain/height", s.getChainHeightHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/chain/status", s.getChainStatusHandler).Methods("GET")

	// Block operations
	s.router.HandleFunc("/api/v1/blocks/latest", s.getLatestBlockHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/blocks/height/{height}", s.getBlockByHeightHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/blocks/{hash}", s.getBlockHandler).Methods("GET")

	// Transaction operations
	s.router.HandleFunc("/api/v1/transactions/{hash}", s.getTransactionHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/transactions/pending", s.getPendingTransactionsHandler).Methods("GET")

	// Wallet operations
	s.router.HandleFunc("/api/v1/wallet/balance/{address}", s.getBalanceHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/wallet/accounts", s.getAccountsHandler).Methods("GET")

	// Network operations
	s.router.HandleFunc("/api/v1/network/peers", s.getPeersHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/network/status", s.getNetworkStatusHandler).Methods("GET")
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)

	return http.ListenAndServe(addr, s.router)
}

// healthHandler provides a simple health check endpoint
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "adrenochain-api",
	})
}

// getChainInfoHandler returns general blockchain information
func (s *Server) getChainInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bestBlock := s.chain.GetBestBlock()
	genesisBlock := s.chain.GetGenesisBlock()

	info := map[string]interface{}{
		"height":    s.chain.GetHeight(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add block-specific information if available
	if bestBlock != nil {
		info["best_block"] = fmt.Sprintf("%x", bestBlock.CalculateHash())
		info["best_block_hash"] = fmt.Sprintf("%x", bestBlock.CalculateHash())
		info["difficulty"] = bestBlock.Header.Difficulty
		info["next_difficulty"] = s.chain.CalculateNextDifficulty()
	} else {
		info["best_block"] = ""
		info["best_block_hash"] = ""
		info["difficulty"] = uint64(0)
		info["next_difficulty"] = uint64(0)
	}

	if genesisBlock != nil {
		info["genesis_block_hash"] = fmt.Sprintf("%x", genesisBlock.CalculateHash())
	} else {
		info["genesis_block_hash"] = ""
	}

	json.NewEncoder(w).Encode(info)
}

// getChainHeightHandler returns the current blockchain height
func (s *Server) getChainHeightHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"height": s.chain.GetHeight(),
	})
}

// getChainStatusHandler returns detailed chain status
func (s *Server) getChainStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bestBlock := s.chain.GetBestBlock()
	genesisBlock := s.chain.GetGenesisBlock()

	status := map[string]interface{}{
		"height":    s.chain.GetHeight(),
		"status":    "active",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add block-specific information if available
	if bestBlock != nil {
		status["best_block_hash"] = fmt.Sprintf("%x", bestBlock.CalculateHash())
		status["best_block_timestamp"] = bestBlock.Header.Timestamp.Format(time.RFC3339)
		status["difficulty"] = bestBlock.Header.Difficulty
		status["next_difficulty"] = s.chain.CalculateNextDifficulty()
	} else {
		status["best_block_hash"] = ""
		status["best_block_timestamp"] = ""
		status["difficulty"] = uint64(0)
		status["next_difficulty"] = uint64(0)
	}

	if genesisBlock != nil {
		status["genesis_block_hash"] = fmt.Sprintf("%x", genesisBlock.CalculateHash())
	} else {
		status["genesis_block_hash"] = ""
	}

	json.NewEncoder(w).Encode(status)
}

// getBlockHandler returns a specific block by hash
func (s *Server) getBlockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	hashHex := vars["hash"]

	// SECURITY FIX: Comprehensive input validation
	if err := s.validateHashInput(hashHex); err != nil {
		// SECURITY FIX: Sanitize error message to prevent information disclosure
		http.Error(w, "Invalid input provided", http.StatusBadRequest)
		return
	}

	// Convert hex string to bytes
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	block := s.chain.GetBlock(hash)
	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	// Convert block to JSON-friendly format
	blockInfo := map[string]interface{}{
		"hash":         fmt.Sprintf("%x", block.CalculateHash()),
		"height":       block.Header.Height,
		"version":      block.Header.Version,
		"prev_hash":    fmt.Sprintf("%x", block.Header.PrevBlockHash),
		"merkle_root":  fmt.Sprintf("%x", block.Header.MerkleRoot),
		"timestamp":    block.Header.Timestamp.Format(time.RFC3339),
		"difficulty":   block.Header.Difficulty,
		"nonce":        block.Header.Nonce,
		"tx_count":     len(block.Transactions),
		"transactions": make([]map[string]interface{}, 0),
	}

	// Add transaction hashes
	for _, tx := range block.Transactions {
		txInfo := map[string]interface{}{
			"hash": fmt.Sprintf("%x", tx.Hash),
			"type": "transaction",
		}
		blockInfo["transactions"] = append(blockInfo["transactions"].([]map[string]interface{}), txInfo)
	}

	json.NewEncoder(w).Encode(blockInfo)
}

// SECURITY FIX: validateHashInput performs comprehensive hash input validation
func (s *Server) validateHashInput(hashHex string) error {
	// Check length
	if len(hashHex) == 0 {
		return fmt.Errorf("hash cannot be empty")
	}
	if len(hashHex) < 32 {
		return fmt.Errorf("hash too short: minimum 32 characters required")
	}
	if len(hashHex) > 128 {
		return fmt.Errorf("hash too long: maximum 128 characters allowed")
	}

	// Check for malicious patterns
	maliciousPatterns := []string{
		"<script", "javascript:", "onload=", "onerror=", "eval(", "expression(",
		"../", "..\\", "union select", "drop table", "delete from", "insert into",
		"rm -rf", "chmod 777", "cat /etc/passwd", "wget", "curl",
	}

	hashLower := strings.ToLower(hashHex)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(hashLower, pattern) {
			return fmt.Errorf("hash contains blocked pattern: %s", pattern)
		}
	}

	// Check for null bytes
	if strings.Contains(hashHex, "\x00") {
		return fmt.Errorf("hash contains null bytes")
	}

	// Validate hex format
	if len(hashHex)%2 != 0 {
		return fmt.Errorf("invalid hex format: odd length")
	}

	// Check if all characters are valid hex
	for _, char := range hashHex {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return fmt.Errorf("invalid hex character: %c", char)
		}
	}

	return nil
}

// getBlockByHeightHandler returns a block by its height
func (s *Server) getBlockByHeightHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	heightStr := vars["height"]

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid height format", http.StatusBadRequest)
		return
	}

	block := s.chain.GetBlockByHeight(height)
	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	// Return block data directly instead of redirecting
	blockInfo := map[string]interface{}{
		"hash":         fmt.Sprintf("%x", block.CalculateHash()),
		"height":       block.Header.Height,
		"version":      block.Header.Version,
		"prev_hash":    fmt.Sprintf("%x", block.Header.PrevBlockHash),
		"merkle_root":  fmt.Sprintf("%x", block.Header.MerkleRoot),
		"timestamp":    block.Header.Timestamp.Format(time.RFC3339),
		"difficulty":   block.Header.Difficulty,
		"nonce":        block.Header.Nonce,
		"tx_count":     len(block.Transactions),
		"transactions": make([]map[string]interface{}, 0),
	}

	// Add transaction hashes
	for _, tx := range block.Transactions {
		txInfo := map[string]interface{}{
			"hash": fmt.Sprintf("%x", tx.Hash),
			"type": "transaction",
		}
		blockInfo["transactions"] = append(blockInfo["transactions"].([]map[string]interface{}), txInfo)
	}

	json.NewEncoder(w).Encode(blockInfo)
}

// getLatestBlockHandler returns the latest block
func (s *Server) getLatestBlockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bestBlock := s.chain.GetBestBlock()
	if bestBlock == nil {
		http.Error(w, "No blocks found", http.StatusNotFound)
		return
	}

	// Return block data directly instead of redirecting
	blockInfo := map[string]interface{}{
		"hash":         fmt.Sprintf("%x", bestBlock.CalculateHash()),
		"height":       bestBlock.Header.Height,
		"version":      bestBlock.Header.Version,
		"prev_hash":    fmt.Sprintf("%x", bestBlock.Header.PrevBlockHash),
		"merkle_root":  fmt.Sprintf("%x", bestBlock.Header.MerkleRoot),
		"timestamp":    bestBlock.Header.Timestamp.Format(time.RFC3339),
		"difficulty":   bestBlock.Header.Difficulty,
		"nonce":        bestBlock.Header.Nonce,
		"tx_count":     len(bestBlock.Transactions),
		"transactions": make([]map[string]interface{}, 0),
	}

	// Add transaction hashes
	for _, tx := range bestBlock.Transactions {
		txInfo := map[string]interface{}{
			"hash": fmt.Sprintf("%x", tx.Hash),
			"type": "transaction",
		}
		blockInfo["transactions"] = append(blockInfo["transactions"].([]map[string]interface{}), txInfo)
	}

	json.NewEncoder(w).Encode(blockInfo)
}

// getTransactionHandler returns a specific transaction by hash
func (s *Server) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	hashHex := vars["hash"]

	// Convert hex string to bytes
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	// Use transaction index for O(1) lookup
	s.txMutex.RLock()
	indexEntry, exists := s.txIndex[string(hash)]
	s.txMutex.RUnlock()

	var foundTx *block.Transaction
	if exists {
		block := s.chain.GetBlockByHeight(indexEntry.BlockHeight)
		if block != nil && indexEntry.TxIndex < len(block.Transactions) {
			foundTx = block.Transactions[indexEntry.TxIndex]
		}
	}

	if foundTx == nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Convert transaction to JSON-friendly format
	txInfo := map[string]interface{}{
		"hash":      fmt.Sprintf("%x", foundTx.Hash),
		"inputs":    len(foundTx.Inputs),
		"outputs":   len(foundTx.Outputs),
		"timestamp": time.Now().UTC().Format(time.RFC3339), // This would be the block timestamp in a real implementation
	}

	json.NewEncoder(w).Encode(txInfo)
}

// getPendingTransactionsHandler returns pending transactions from mempool
func (s *Server) getPendingTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.mempool == nil {
		http.Error(w, "Mempool not available", http.StatusServiceUnavailable)
		return
	}

	pendingTxs := s.mempool.GetPendingTransactions()
	count := s.mempool.GetPendingTransactionCount()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_transactions": pendingTxs,
		"count":                count,
	})
}

// getBalanceHandler returns the balance for a specific address
func (s *Server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	address := vars["address"]

	if s.wallet == nil {
		http.Error(w, "Wallet not available", http.StatusServiceUnavailable)
		return
	}

	balance := s.wallet.GetBalance(address)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"balance": balance,
	})
}

// getAccountsHandler returns all wallet accounts
func (s *Server) getAccountsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		http.Error(w, "Wallet not available", http.StatusServiceUnavailable)
		return
	}

	accounts := s.wallet.GetAllAccounts()
	accountList := make([]map[string]interface{}, 0)

	for _, account := range accounts {
		accountInfo := map[string]interface{}{
			"address":    account.Address,
			"public_key": fmt.Sprintf("%x", account.PublicKey),
		}
		accountList = append(accountList, accountInfo)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accountList,
		"count":    len(accountList),
	})
}

// getPeersHandler returns connected peers
func (s *Server) getPeersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.network == nil {
		// Fallback if network is not available
		json.NewEncoder(w).Encode(map[string]interface{}{
			"peers": []interface{}{},
			"count": 0,
		})
		return
	}

	// Get actual peer information from network
	peers := s.network.GetPeers()
	peerCount := s.network.GetPeerCount()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"peers": peers,
		"count": peerCount,
	})
}

// getNetworkStatusHandler returns network status information
func (s *Server) getNetworkStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.network == nil {
		http.Error(w, "Network not available", http.StatusServiceUnavailable)
		return
	}

	peers := s.network.GetPeers()
	peerCount := s.network.GetPeerCount()

	// Get additional network metrics
	chainHeight := s.chain.GetHeight()

	// Calculate network health metrics
	networkHealth := "healthy"
	if peerCount == 0 {
		networkHealth = "disconnected"
	} else if peerCount < 3 {
		networkHealth = "degraded"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "active",
		"peer_count":     peerCount,
		"peers":          peers,
		"listening":      true,
		"chain_height":   chainHeight,
		"network_health": networkHealth,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

// metricsHandler provides basic metrics in JSON format
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"blockchain": map[string]interface{}{
			"height": s.chain.GetHeight(),
			"difficulty": func() uint64 {
				if bestBlock := s.chain.GetBestBlock(); bestBlock != nil {
					return bestBlock.Header.Difficulty
				}
				return 0
			}(),
		},
		"network": map[string]interface{}{
			"peer_count": func() int {
				if s.network != nil {
					return s.network.GetPeerCount()
				}
				return 0
			}(),
		},
		"system": map[string]interface{}{
			"memory_usage_bytes": m.Alloc,
			"goroutines":         runtime.NumGoroutine(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(metrics)
}

// prometheusHandler provides metrics in Prometheus format
func (s *Server) prometheusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	bestBlock := s.chain.GetBestBlock()
	var difficulty uint64
	if bestBlock != nil {
		difficulty = bestBlock.Header.Difficulty
	}

	peerCount := 0
	if s.network != nil {
		peerCount = s.network.GetPeerCount()
	}

	prometheusMetrics := fmt.Sprintf(`# HELP adrenochain_block_height Current blockchain height
# TYPE adrenochain_block_height gauge
adrenochain_block_height %d

# HELP adrenochain_chain_difficulty Current chain difficulty
# TYPE adrenochain_chain_difficulty gauge
adrenochain_chain_difficulty %d

# HELP adrenochain_connected_peers Number of connected peers
# TYPE adrenochain_connected_peers gauge
adrenochain_connected_peers %d

# HELP adrenochain_memory_usage_bytes Current memory usage in bytes
# TYPE adrenochain_memory_usage_bytes gauge
adrenochain_memory_usage_bytes %d

# HELP adrenochain_goroutines Number of goroutines
# TYPE adrenochain_goroutines gauge
adrenochain_goroutines %d

# HELP adrenochain_uptime_seconds Node uptime in seconds
# TYPE adrenochain_uptime_seconds gauge
adrenochain_uptime_seconds %d
`,
		s.chain.GetHeight(),
		difficulty,
		peerCount,
		m.Alloc,
		runtime.NumGoroutine(),
		int64(time.Since(time.Now().Add(-time.Hour)).Seconds()), // Mock uptime
	)

	w.Write([]byte(prometheusMetrics))
}
