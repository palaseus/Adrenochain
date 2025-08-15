package yield

import "errors"

// Yield farming specific errors
var (
	ErrFarmPaused              = errors.New("farm is paused")
	ErrFarmAlreadyPaused       = errors.New("farm is already paused")
	ErrFarmNotPaused           = errors.New("farm is not paused")
	ErrPoolNotFound            = errors.New("pool not found")
	ErrUserNotInPool           = errors.New("user not in pool")
	ErrInsufficientStaked      = errors.New("insufficient staked amount")
	ErrInvalidStakingToken     = errors.New("invalid staking token")
	ErrInvalidAllocPoint       = errors.New("invalid allocation point")
	ErrInvalidUser             = errors.New("invalid user")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrInvalidRewardToken      = errors.New("invalid reward token")
	ErrInvalidRewardPerSecond  = errors.New("invalid reward per second")
	ErrInvalidStartTime        = errors.New("invalid start time")
	ErrInvalidEndTime          = errors.New("invalid end time")
)
