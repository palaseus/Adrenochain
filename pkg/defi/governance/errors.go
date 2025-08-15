package governance

import "errors"

// Governance-specific errors
var (
	ErrGovernancePaused              = errors.New("governance is paused")
	ErrGovernanceAlreadyPaused       = errors.New("governance is already paused")
	ErrGovernanceNotPaused           = errors.New("governance is not paused")
	ErrProposalNotFound              = errors.New("proposal not found")
	ErrProposalNotActive             = errors.New("proposal is not active")
	ErrProposalNotSucceeded          = errors.New("proposal has not succeeded")
	ErrProposalCannotBeCanceled      = errors.New("proposal cannot be canceled")
	ErrVotingPeriodClosed            = errors.New("voting period is closed")
	ErrAlreadyVoted                  = errors.New("user has already voted")
	ErrInsufficientVotingPower       = errors.New("insufficient voting power")
	ErrInsufficientProposalPower     = errors.New("insufficient proposal power")
	ErrExecutionDelayNotMet          = errors.New("execution delay not met")
	ErrNotProposer                   = errors.New("not the proposer")
	ErrInvalidProposer               = errors.New("invalid proposer")
	ErrInvalidTargets                = errors.New("invalid targets")
	ErrInvalidValues                 = errors.New("invalid values")
	ErrInvalidSignatures             = errors.New("invalid signatures")
	ErrInvalidCalldatas              = errors.New("invalid calldatas")
	ErrInvalidDescription            = errors.New("invalid description")
	ErrInvalidVoter                  = errors.New("invalid voter")
	ErrInvalidVoteSupport            = errors.New("invalid vote support")
	ErrInvalidExecutor               = errors.New("invalid executor")
	ErrInvalidCanceler               = errors.New("invalid canceler")
)
