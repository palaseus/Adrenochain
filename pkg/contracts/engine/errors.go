package engine

import "errors"

// Common errors for the contract engine
var (
	ErrInvalidAddress     = errors.New("invalid address format")
	ErrInvalidHash        = errors.New("invalid hash format")
	ErrInsufficientGas    = errors.New("insufficient gas for operation")
	ErrContractNotFound   = errors.New("contract not found")
	ErrInvalidContract    = errors.New("invalid contract code")
	ErrExecutionFailed    = errors.New("contract execution failed")
	ErrStorageError       = errors.New("storage operation failed")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrGasLimitExceeded   = errors.New("gas limit exceeded")
	ErrInvalidOperation   = errors.New("invalid operation")
	ErrStateCorruption    = errors.New("state corruption detected")
	ErrConsensusMismatch  = errors.New("consensus state mismatch")
	ErrInvalidNonce       = errors.New("invalid nonce")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrContractReverted   = errors.New("contract execution reverted")
	ErrOutOfGas           = errors.New("out of gas")
	ErrStackOverflow      = errors.New("stack overflow")
	ErrStackUnderflow     = errors.New("stack underflow")
	ErrInvalidJump        = errors.New("invalid jump destination")
	ErrInvalidInstruction = errors.New("invalid instruction")
	ErrMemoryOverflow     = errors.New("memory overflow")
	ErrInvalidOpcode      = errors.New("invalid opcode")
)
