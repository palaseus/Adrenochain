package storage

import "errors"

// Contract storage specific errors
var (
	ErrContractAlreadyExists     = errors.New("contract already exists")
	ErrContractNotFound          = errors.New("contract not found")
	ErrInvalidStateChange        = errors.New("invalid state change")
	ErrStateUpdateFailed         = errors.New("state update failed")
	ErrStorageFull               = errors.New("storage is full")
	ErrInvalidAddress            = errors.New("invalid address")
	ErrInvalidCode               = errors.New("invalid code")
	ErrInvalidStorageKey         = errors.New("invalid storage key")
	ErrInvalidStorageValue       = errors.New("invalid storage value")
	ErrStateCorruption           = errors.New("state corruption detected")
	ErrPruningFailed             = errors.New("state pruning failed")
	ErrPruningAlreadyRunning     = errors.New("pruning already running")
	ErrAutoPruningNotEnabled     = errors.New("auto pruning not enabled")
	ErrInactiveCleanupNotEnabled = errors.New("inactive cleanup not enabled")
	ErrInvalidPruningInterval    = errors.New("invalid pruning interval")
	ErrInvalidMaxHistorySize     = errors.New("invalid max history size")
	ErrInvalidMaxStorageSize     = errors.New("invalid max storage size")
	ErrContractStorageNotEnabled = errors.New("contract storage not enabled")
	ErrCompressionNotEnabled     = errors.New("compression not enabled")
)
