package api

import "errors"

// Contract API specific errors
var (
	ErrContractNotFound      = errors.New("contract not found")
	ErrInvalidBytecode       = errors.New("invalid bytecode")
	ErrContractTooLarge      = errors.New("contract too large")
	ErrInvalidGasPrice       = errors.New("invalid gas price")
	ErrInvalidContractAddress = errors.New("invalid contract address")
	ErrInvalidMethod         = errors.New("invalid method")
	ErrInvalidArgs           = errors.New("invalid arguments")
	ErrGasLimitExceeded      = errors.New("gas limit exceeded")
	ErrContractDeploymentFailed = errors.New("contract deployment failed")
	ErrContractCallFailed    = errors.New("contract call failed")
)
