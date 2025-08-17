package bridge

import (
	"errors"
	"math/big"
	"time"
)

// BridgeStatus represents the current status of the bridge
type BridgeStatus string

const (
	BridgeStatusActive    BridgeStatus = "active"
	BridgeStatusPaused    BridgeStatus = "paused"
	BridgeStatusEmergency BridgeStatus = "emergency"
	BridgeStatusUpgrading BridgeStatus = "upgrading"
)

// ChainID represents a blockchain network identifier
type ChainID string

const (
	ChainIDGoChain  ChainID = "gochain"
	ChainIDEthereum ChainID = "ethereum"
	ChainIDPolygon  ChainID = "polygon"
	ChainIDArbitrum ChainID = "arbitrum"
	ChainIDOptimism ChainID = "optimism"
)

// TransactionStatus represents the status of a cross-chain transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusExecuted  TransactionStatus = "executed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReverted  TransactionStatus = "reverted"
)

// AssetType represents the type of asset being bridged
type AssetType string

const (
	AssetTypeNative  AssetType = "native"  // Native token (e.g., ETH, MATIC)
	AssetTypeERC20   AssetType = "erc20"   // ERC-20 token
	AssetTypeERC721  AssetType = "erc721"  // NFT
	AssetTypeERC1155 AssetType = "erc1155" // Multi-token
)

// CrossChainTransaction represents a cross-chain transfer
type CrossChainTransaction struct {
	ID                 string            `json:"id"`
	SourceChain        ChainID           `json:"source_chain"`
	DestinationChain   ChainID           `json:"destination_chain"`
	SourceAddress      string            `json:"source_address"`
	DestinationAddress string            `json:"destination_address"`
	AssetType          AssetType         `json:"asset_type"`
	AssetAddress       string            `json:"asset_address"` // Contract address for tokens
	Amount             *big.Int          `json:"amount"`
	TokenID            *big.Int          `json:"token_id,omitempty"` // For NFTs
	Status             TransactionStatus `json:"status"`
	SourceTxHash       string            `json:"source_tx_hash"`
	DestinationTxHash  string            `json:"destination_tx_hash,omitempty"`
	ValidatorID        string            `json:"validator_id"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	ExecutedAt         *time.Time        `json:"executed_at,omitempty"`
	FailedAt           *time.Time        `json:"failed_at,omitempty"`
	FailureReason      string            `json:"failure_reason,omitempty"`
	GasUsed            *big.Int          `json:"gas_used,omitempty"`
	Fee                *big.Int          `json:"fee"`
}

// Validator represents a bridge validator
type Validator struct {
	ID             string    `json:"id"`
	Address        string    `json:"address"`
	ChainID        ChainID   `json:"chain_id"`
	StakeAmount    *big.Int  `json:"stake_amount"`
	IsActive       bool      `json:"is_active"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TotalValidated int64     `json:"total_validated"`
	SuccessRate    float64   `json:"success_rate"`
}

// AssetMapping represents a mapping between assets on different chains
type AssetMapping struct {
	ID               string    `json:"id"`
	SourceChain      ChainID   `json:"source_chain"`
	DestinationChain ChainID   `json:"destination_chain"`
	SourceAsset      string    `json:"source_asset"`
	DestinationAsset string    `json:"destination_asset"`
	AssetType        AssetType `json:"asset_type"`
	Decimals         uint8     `json:"decimals"`
	IsActive         bool      `json:"is_active"`
	MinAmount        *big.Int  `json:"min_amount"`
	MaxAmount        *big.Int  `json:"max_amount"`
	DailyLimit       *big.Int  `json:"daily_limit"`
	DailyUsed        *big.Int  `json:"daily_used"`
	FeePercentage    float64   `json:"fee_percentage"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BridgeConfig represents the bridge configuration
type BridgeConfig struct {
	ID                    string        `json:"id"`
	Status                BridgeStatus  `json:"status"`
	MinValidators         int           `json:"min_validators"`
	RequiredConfirmations int           `json:"required_confirmations"`
	MaxTransactionAmount  *big.Int      `json:"max_transaction_amount"`
	MinTransactionAmount  *big.Int      `json:"min_transaction_amount"`
	TransactionTimeout    time.Duration `json:"transaction_timeout"`
	GasLimit              *big.Int      `json:"gas_limit"`
	MaxDailyVolume        *big.Int      `json:"max_daily_volume"`
	DailyVolumeUsed       *big.Int      `json:"daily_volume_used"`
	FeeCollector          string        `json:"fee_collector"`
	EmergencyPaused       bool          `json:"emergency_paused"`
	PausedBy              string        `json:"paused_by,omitempty"`
	PausedAt              *time.Time    `json:"paused_at,omitempty"`
	PauseReason           string        `json:"pause_reason,omitempty"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

// TransferRequest represents a request for a cross-chain transfer
type TransferRequest struct {
	SourceChain        ChainID   `json:"source_chain"`
	DestinationChain   ChainID   `json:"destination_chain"`
	SourceAddress      string    `json:"source_address"`
	DestinationAddress string    `json:"destination_address"`
	AssetType          AssetType `json:"asset_type"`
	AssetAddress       string    `json:"asset_address"`
	Amount             *big.Int  `json:"amount"`
	TokenID            *big.Int  `json:"token_id,omitempty"`
}

// BridgeError represents bridge-specific errors
type BridgeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e BridgeError) Error() string {
	return e.Message
}

// Common bridge errors
var (
	ErrBridgeInactive         = errors.New("bridge is inactive")
	ErrBridgePaused           = errors.New("bridge is paused")
	ErrBridgeEmergency        = errors.New("bridge is in emergency mode")
	ErrInsufficientValidators = errors.New("insufficient validators")
	ErrInvalidChainID         = errors.New("invalid chain ID")
	ErrInvalidAssetMapping    = errors.New("invalid asset mapping")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrTransactionExpired     = errors.New("transaction expired")
	ErrInsufficientStake      = errors.New("insufficient validator stake")
	ErrValidatorInactive      = errors.New("validator is inactive")
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrAmountTooSmall         = errors.New("amount too small")
	ErrAmountTooLarge         = errors.New("amount too large")
	ErrDailyLimitExceeded     = errors.New("daily limit exceeded")
	ErrAssetNotSupported      = errors.New("asset not supported")
	ErrInvalidAddress         = errors.New("invalid address")
	ErrGasLimitExceeded       = errors.New("gas limit exceeded")
	ErrInsufficientFee        = errors.New("insufficient fee")
)
