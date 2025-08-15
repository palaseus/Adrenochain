package wasm

import "errors"

// WASM-specific errors
var (
	ErrMemoryGrowExceedsMaximum = errors.New("memory grow exceeds maximum size")
	ErrMemoryAccessOutOfBounds  = errors.New("memory access out of bounds")
	ErrStackUnderflow           = errors.New("stack underflow")
	ErrGlobalImmutable          = errors.New("global variable is immutable")
	ErrGlobalTypeMismatch       = errors.New("global variable type mismatch")
	ErrTableIndexOutOfBounds    = errors.New("table index out of bounds")
	ErrTableTypeMismatch        = errors.New("table element type mismatch")
	ErrTableGrowExceedsMaximum  = errors.New("table grow exceeds maximum size")
	ErrInvalidWASM              = errors.New("invalid WASM module")
	ErrUnsupportedOpcode        = errors.New("unsupported WASM opcode")
	ErrFunctionNotFound         = errors.New("function not found")
	ErrInvalidFunctionCall      = errors.New("invalid function call")
	ErrMemoryAllocationFailed   = errors.New("memory allocation failed")
	ErrExecutionFailed          = errors.New("WASM execution failed")
	ErrGasLimitExceeded         = errors.New("gas limit exceeded")
	ErrInvalidImport            = errors.New("invalid import")
	ErrInvalidExport            = errors.New("invalid export")
	ErrModuleValidationFailed   = errors.New("WASM module validation failed")
)
