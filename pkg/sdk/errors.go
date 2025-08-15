package sdk

import "errors"

// SDK-specific errors
var (
	ErrUnsupportedTokenType    = errors.New("unsupported token type")
	ErrAMMNotInitialized       = errors.New("AMM not initialized")
	ErrLendingNotInitialized   = errors.New("lending not initialized")
	ErrYieldFarmingNotInitialized = errors.New("yield farming not initialized")
	ErrGovernanceNotInitialized = errors.New("governance not initialized")
	ErrOracleNotInitialized    = errors.New("oracle not initialized")
	ErrInvalidConfiguration    = errors.New("invalid configuration")
	ErrOperationFailed        = errors.New("operation failed")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrInvalidAddress         = errors.New("invalid address")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrContractNotFound       = errors.New("contract not found")
)
