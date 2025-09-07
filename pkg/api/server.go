package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
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

// Server represents the HTTP API server
type Server struct {
	router  *mux.Router
	chain   ChainInterface
	wallet  WalletInterface
	network NetworkInterface
	port    int
}

// ServerConfig holds configuration for the API server
type ServerConfig struct {
	Port    int
	Chain   ChainInterface
	Wallet  WalletInterface
	Network NetworkInterface
}

// NewServer creates a new API server
func NewServer(config *ServerConfig) *Server {
	router := mux.NewRouter()
	server := &Server{
		router:  router,
		chain:   config.Chain,
		wallet:  config.Wallet,
		network: config.Network,
		port:    config.Port,
	}

	server.setupRoutes()
	return server
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

	// For now, we'll search through blocks to find the transaction
	// In a real implementation, you'd have a transaction index
	height := s.chain.GetHeight()
	var foundTx *block.Transaction

	for h := uint64(0); h <= height; h++ {
		block := s.chain.GetBlockByHeight(h)
		if block == nil {
			continue
		}

		for _, tx := range block.Transactions {
			if string(tx.Hash) == string(hash) {
				foundTx = tx
				break
			}
		}
		if foundTx != nil {
			break
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

	// For now, return empty list since we don't have mempool access in this context
	// In a real implementation, you'd access the mempool
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_transactions": []interface{}{},
		"count":                0,
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

	// For now, return basic network status
	// In a real implementation, you'd access the network layer
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "active",
		"peer_count": 0,
		"listening":  true,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
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
