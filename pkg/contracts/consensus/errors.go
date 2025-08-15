package consensus

import "errors"

// Consensus-specific errors
var (
	ErrValidatorAlreadyExists  = errors.New("validator already exists")
	ErrRollbackNotEnabled      = errors.New("rollback not enabled")
	ErrInvalidContractAddress  = errors.New("invalid contract address")
	ErrInvalidSender           = errors.New("invalid sender")
	ErrInvalidGasLimit         = errors.New("invalid gas limit")
	ErrInvalidGasPrice         = errors.New("invalid gas price")
	ErrTransactionNotFound     = errors.New("transaction not found")
	ErrInvalidBlockNumber      = errors.New("invalid block number")
	ErrConsensusNotInitialized = errors.New("consensus not initialized")
	ErrStateValidationFailed   = errors.New("state validation failed")
	ErrExecutionTimeout        = errors.New("execution timeout")
	ErrValidationTimeout       = errors.New("validation timeout")
	ErrContractExecutionNotEnabled = errors.New("contract execution not enabled")
	ErrBlockValidationNotEnabled = errors.New("block validation not enabled")
	ErrStateRollbackNotEnabled = errors.New("state rollback not enabled")
	ErrBlockNotFound          = errors.New("block not found")
)
