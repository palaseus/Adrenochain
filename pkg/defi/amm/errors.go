package amm

import "errors"

// AMM-specific errors
var (
	ErrPoolAlreadyPaused      = errors.New("pool is already paused")
	ErrPoolNotPaused          = errors.New("pool is not paused")
	ErrPoolPaused             = errors.New("pool is paused")
	ErrInvalidToken           = errors.New("invalid token")
	ErrInsufficientOutputAmount = errors.New("insufficient output amount")
	ErrInsufficientLPTokens   = errors.New("insufficient LP tokens")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrInsufficientAmountA    = errors.New("insufficient amount A")
	ErrInsufficientAmountB    = errors.New("insufficient amount B")
	ErrInsufficientLiquidity  = errors.New("insufficient liquidity")
	ErrInvalidReserves        = errors.New("invalid reserves")
	ErrInvalidFee             = errors.New("invalid fee")
)
