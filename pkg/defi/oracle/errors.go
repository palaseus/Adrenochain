package oracle

import "errors"

// Oracle-specific errors
var (
	ErrProviderFailure = errors.New("oracle provider failure")
	ErrInvalidProof    = errors.New("invalid oracle proof")
	ErrInvalidPrice    = errors.New("invalid price data")
	ErrInsufficientProviders = errors.New("insufficient oracle providers")
	ErrPriceTooOld     = errors.New("price data too old")
	ErrLowConfidence   = errors.New("insufficient confidence level")
	ErrProviderNotFound = errors.New("oracle provider not found")
	ErrInvalidWeight   = errors.New("invalid provider weight")
)
