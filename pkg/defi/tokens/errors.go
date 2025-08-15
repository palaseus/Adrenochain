package tokens

import "errors"

// Token-specific errors
var (
	ErrTokenPaused              = errors.New("token is paused")
	ErrAddressBlacklisted       = errors.New("address is blacklisted")
	ErrInvalidAmount            = errors.New("invalid amount")
	ErrSelfTransfer             = errors.New("cannot transfer to self")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrInsufficientAllowance    = errors.New("insufficient allowance")
	ErrMintingNotAllowed        = errors.New("minting not allowed")
	ErrBurningNotAllowed        = errors.New("burning not allowed")
	ErrPausingNotAllowed        = errors.New("pausing not allowed")
	ErrBlacklistingNotAllowed   = errors.New("blacklisting not allowed")
	ErrExceedsMaxSupply         = errors.New("exceeds maximum supply")
	ErrTokenAlreadyPaused       = errors.New("token is already paused")
	ErrTokenNotPaused           = errors.New("token is not paused")
	ErrAddressAlreadyBlacklisted = errors.New("address is already blacklisted")
	ErrAddressNotBlacklisted    = errors.New("address is not blacklisted")
	ErrInvalidTokenConfig       = errors.New("invalid token configuration")
	ErrTokenNotFound            = errors.New("token not found")
	ErrUnauthorizedOperation    = errors.New("unauthorized operation")
	ErrInvalidTokenAddress      = errors.New("invalid token address")
	ErrTokenTransferFailed      = errors.New("token transfer failed")
	ErrTokenApprovalFailed      = errors.New("token approval failed")
	ErrTokenMintingFailed       = errors.New("token minting failed")
	ErrTokenBurningFailed       = errors.New("token burning failed")
	ErrTokenAlreadyExists       = errors.New("token already exists")
	ErrInvalidBatchTransfer     = errors.New("invalid batch transfer")
)
