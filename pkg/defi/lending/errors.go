package lending

import "errors"

// Lending-specific errors
var (
	ErrAssetAlreadyExists              = errors.New("asset already exists")
	ErrProtocolPaused                  = errors.New("protocol is paused")
	ErrInsufficientBalance             = errors.New("insufficient balance")
	ErrInsufficientBorrowBalance       = errors.New("insufficient borrow balance")
	ErrInsufficientCollateral          = errors.New("insufficient collateral")
	ErrInsufficientLiquidationCollateral = errors.New("insufficient liquidation collateral")
	ErrProtocolAlreadyPaused           = errors.New("protocol is already paused")
	ErrProtocolNotPaused               = errors.New("protocol is not paused")
	ErrAssetNotSupported               = errors.New("asset not supported")
	ErrUserNotFound                    = errors.New("user not found")
	ErrNotEligibleForLiquidation      = errors.New("not eligible for liquidation")
	ErrCannotLiquidateSelf             = errors.New("cannot liquidate self")
	ErrInvalidAmount                   = errors.New("invalid amount")
	ErrInvalidCollateralRatio          = errors.New("invalid collateral ratio")
	ErrInvalidMaxLTV                   = errors.New("invalid max LTV")
	ErrInvalidInterestRate             = errors.New("invalid interest rate")
)
