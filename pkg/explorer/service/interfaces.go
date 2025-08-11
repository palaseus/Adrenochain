package service

import (
	"context"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// BlockchainDataProvider defines the interface for accessing blockchain data
type BlockchainDataProvider interface {
	// Block operations
	GetBlock(hash []byte) (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetLatestBlock() (*block.Block, error)
	GetBlockHeight() uint64

	// Transaction operations
	GetTransaction(hash []byte) (*block.Transaction, error)
	GetTransactionsByBlock(blockHash []byte) ([]*block.Transaction, error)
	GetPendingTransactions() ([]*block.Transaction, error)

	// Address operations
	GetAddressBalance(address string) (uint64, error)
	GetAddressTransactions(address string, limit, offset int) ([]*block.Transaction, error)
	GetAddressUTXOs(address string) ([]*UTXO, error)

	// Statistics and metrics
	GetBlockchainStats() (*BlockchainStats, error)
	GetNetworkInfo() (*NetworkInfo, error)
}

// CacheProvider defines the interface for caching frequently accessed data
type CacheProvider interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	GetStats() CacheStats
}

// SearchProvider defines the interface for search functionality
type SearchProvider interface {
	Search(query string) (*SearchResult, error)
	SearchBlocks(query string, limit, offset int) ([]*BlockSummary, error)
	SearchTransactions(query string, limit, offset int) ([]*TransactionSummary, error)
	SearchAddresses(query string, limit, offset int) ([]*AddressSummary, error)
}

// ExplorerService defines the main explorer service interface
type ExplorerService interface {
	// Core explorer functionality
	GetDashboard(ctx context.Context) (*Dashboard, error)
	GetBlockDetails(ctx context.Context, hash []byte) (*BlockDetails, error)
	GetTransactionDetails(ctx context.Context, hash []byte) (*TransactionDetails, error)
	GetAddressDetails(ctx context.Context, address string) (*AddressDetails, error)

	// List operations with pagination
	GetBlocks(ctx context.Context, limit, offset int) ([]*BlockSummary, error)
	GetTransactions(ctx context.Context, limit, offset int) ([]*TransactionSummary, error)

	// Search functionality
	Search(ctx context.Context, query string) (*SearchResult, error)

	// Statistics and metrics
	GetStatistics(ctx context.Context) (*Statistics, error)
}

// Data Models

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash    []byte `json:"tx_hash"`
	TxIndex   uint32 `json:"tx_index"`
	Value     uint64 `json:"value"`
	Script    []byte `json:"script"`
	Address   string `json:"address"`
	BlockHash []byte `json:"block_hash"`
	Height    uint64 `json:"height"`
}

// BlockchainStats represents overall blockchain statistics
type BlockchainStats struct {
	TotalBlocks       uint64    `json:"total_blocks"`
	TotalTransactions uint64    `json:"total_transactions"`
	TotalAddresses    uint64    `json:"total_addresses"`
	TotalSupply       uint64    `json:"total_supply"`
	LastBlockTime     time.Time `json:"last_block_time"`
	AverageBlockTime  float64   `json:"average_block_time"`
	Difficulty        uint64    `json:"difficulty"`
}

// NetworkInfo represents current network information
type NetworkInfo struct {
	Status          string    `json:"status"`
	PeerCount       int       `json:"peer_count"`
	IsListening     bool      `json:"is_listening"`
	LastUpdate      time.Time `json:"last_update"`
	NetworkVersion  string    `json:"network_version"`
	ProtocolVersion string    `json:"protocol_version"`
}

// Dashboard represents the main dashboard data
type Dashboard struct {
	Stats        *BlockchainStats      `json:"stats"`
	RecentBlocks []*BlockSummary       `json:"recent_blocks"`
	RecentTxs    []*TransactionSummary `json:"recent_transactions"`
	NetworkInfo  *NetworkInfo          `json:"network_info"`
	LastUpdate   time.Time             `json:"last_update"`
}

// BlockSummary represents a summary of block information
type BlockSummary struct {
	Hash          []byte    `json:"hash"`
	Height        uint64    `json:"height"`
	Timestamp     time.Time `json:"timestamp"`
	TxCount       int       `json:"tx_count"`
	Size          uint64    `json:"size"`
	Difficulty    uint64    `json:"difficulty"`
	Confirmations uint64    `json:"confirmations"`
}

// BlockDetails represents detailed block information
type BlockDetails struct {
	*BlockSummary
	PrevHash     []byte                `json:"prev_hash"`
	NextHash     []byte                `json:"next_hash,omitempty"`
	MerkleRoot   []byte                `json:"merkle_root"`
	Nonce        uint64                `json:"nonce"`
	Version      uint32                `json:"version"`
	Transactions []*TransactionSummary `json:"transactions"`
	Validation   *BlockValidation      `json:"validation"`
}

// BlockValidation represents block validation status
type BlockValidation struct {
	IsValid       bool   `json:"is_valid"`
	Error         string `json:"error,omitempty"`
	Confirmations uint64 `json:"confirmations"`
	Finality      string `json:"finality"`
}

// TransactionSummary represents a summary of transaction information
type TransactionSummary struct {
	Hash      []byte    `json:"hash"`
	BlockHash []byte    `json:"block_hash"`
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	Inputs    int       `json:"input_count"`
	Outputs   int       `json:"output_count"`
	Amount    uint64    `json:"amount"`
	Fee       uint64    `json:"fee"`
	Status    string    `json:"status"`
}

// TransactionDetails represents detailed transaction information
type TransactionDetails struct {
	*TransactionSummary
	RawTx         *block.Transaction `json:"raw_transaction"`
	InputDetails  []*InputDetail     `json:"input_details"`
	OutputDetails []*OutputDetail    `json:"output_details"`
	BlockInfo     *BlockSummary      `json:"block_info"`
}

// InputDetail represents detailed input information
type InputDetail struct {
	TxHash  []byte `json:"tx_hash"`
	TxIndex uint32 `json:"tx_index"`
	Script  []byte `json:"script"`
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

// OutputDetail represents detailed output information
type OutputDetail struct {
	Index   uint32 `json:"index"`
	Script  []byte `json:"script"`
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Spent   bool   `json:"spent"`
	SpentBy []byte `json:"spent_by,omitempty"`
}

// AddressSummary represents a summary of address information
type AddressSummary struct {
	Address   string    `json:"address"`
	Balance   uint64    `json:"balance"`
	TxCount   int       `json:"transaction_count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// AddressDetails represents detailed address information
type AddressDetails struct {
	*AddressSummary
	UTXOs         []*UTXO               `json:"utxos"`
	Transactions  []*TransactionSummary `json:"transactions"`
	TotalReceived uint64                `json:"total_received"`
	TotalSent     uint64                `json:"total_sent"`
}

// SearchResult represents a search result
type SearchResult struct {
	Query       string              `json:"query"`
	Type        string              `json:"type"`
	Block       *BlockSummary       `json:"block,omitempty"`
	Transaction *TransactionSummary `json:"transaction,omitempty"`
	Address     *AddressSummary     `json:"address,omitempty"`
	Suggestions []string            `json:"suggestions,omitempty"`
	Error       string              `json:"error,omitempty"`
}

// Statistics represents comprehensive blockchain statistics
type Statistics struct {
	Blockchain  *BlockchainStats  `json:"blockchain"`
	Network     *NetworkInfo      `json:"network"`
	Performance *PerformanceStats `json:"performance"`
	LastUpdate  time.Time         `json:"last_update"`
}

// PerformanceStats represents performance metrics
type PerformanceStats struct {
	AverageResponseTime float64 `json:"average_response_time"`
	RequestsPerSecond   float64 `json:"requests_per_second"`
	CacheHitRate        float64 `json:"cache_hit_rate"`
	DatabaseQueries     int     `json:"database_queries"`
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
	Size    int     `json:"size"`
	MaxSize int     `json:"max_size"`
}
