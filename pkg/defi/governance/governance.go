package governance

import (
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// Governance represents a DAO governance system
type Governance struct {
	mu sync.RWMutex

	// Governance information
	GovernanceID string
	Name         string
	Symbol       string
	Decimals     uint8
	Owner        engine.Address
	Paused       bool
	
	// Governance token
	GovernanceToken engine.Address
	
	// Voting settings
	MinQuorum        *big.Int // Minimum votes required for proposal to pass
	VotingPeriod     time.Duration // How long voting is open
	ExecutionDelay   time.Duration // Delay before proposal can be executed
	ProposalThreshold *big.Int // Minimum tokens required to create proposal
	
	// Proposals
	Proposals map[uint64]*Proposal
	
	// Users
	Users map[engine.Address]*User
	
	// Events
	ProposalCreatedEvents []ProposalCreatedEvent
	VoteCastEvents        []VoteCastEvent
	ProposalExecutedEvents []ProposalExecutedEvent
	ProposalCanceledEvents []ProposalCanceledEvent
	
	// Statistics
	TotalProposals uint64
	TotalVotes     uint64
	LastUpdate     time.Time
}

// Proposal represents a governance proposal
type Proposal struct {
	ProposalID      uint64
	Proposer        engine.Address
	Targets         []engine.Address
	Values          []*big.Int
	Signatures      []string
	Calldatas      []string
	Description     string
	StartTime      time.Time
	EndTime        time.Time
	Executed       bool
	Canceled       bool
	ForVotes       *big.Int
	AgainstVotes   *big.Int
	AbstainVotes   *big.Int
	TotalVotes     *big.Int
	QuorumReached  bool
	Votes          map[engine.Address]*Vote
	State          ProposalState
}

// Vote represents a user's vote on a proposal
type Vote struct {
	ProposalID uint64
	Voter      engine.Address
	Support    VoteSupport
	Votes      *big.Int
	Reason     string
	Timestamp  time.Time
}

// User represents a governance user
type User struct {
	Address        engine.Address
	VotingPower   *big.Int
	DelegatedTo   engine.Address
	Delegators    map[engine.Address]*big.Int
	LastVote      time.Time
	ProposalsCreated uint64
}

// ProposalState represents the state of a proposal
type ProposalState uint8

const (
	ProposalStatePending ProposalState = iota
	ProposalStateActive
	ProposalStateCanceled
	ProposalStateDefeated
	ProposalStateSucceeded
	ProposalStateQueued
	ProposalStateExpired
	ProposalStateExecuted
)

// VoteSupport represents the support level for a proposal
type VoteSupport uint8

const (
	VoteSupportAgainst VoteSupport = iota
	VoteSupportFor
	VoteSupportAbstain
)

// NewGovernance creates a new governance system
func NewGovernance(
	governanceID, name, symbol string,
	decimals uint8,
	owner engine.Address,
	governanceToken engine.Address,
	minQuorum, proposalThreshold *big.Int,
	votingPeriod, executionDelay time.Duration,
) *Governance {
	return &Governance{
		GovernanceID:     governanceID,
		Name:             name,
		Symbol:           symbol,
		Decimals:         decimals,
		Owner:            owner,
		Paused:           false,
		GovernanceToken:  governanceToken,
		MinQuorum:        new(big.Int).Set(minQuorum),
		VotingPeriod:     votingPeriod,
		ExecutionDelay:   executionDelay,
		ProposalThreshold: new(big.Int).Set(proposalThreshold),
		Proposals:        make(map[uint64]*Proposal),
		Users:            make(map[engine.Address]*User),
		ProposalCreatedEvents: make([]ProposalCreatedEvent, 0),
		VoteCastEvents:        make([]VoteCastEvent, 0),
		ProposalExecutedEvents: make([]ProposalExecutedEvent, 0),
		ProposalCanceledEvents: make([]ProposalCanceledEvent, 0),
		TotalProposals:   0,
		TotalVotes:       0,
		LastUpdate:       time.Now(),
	}
}

// CreateProposal creates a new governance proposal
func (g *Governance) CreateProposal(
	proposer engine.Address,
	targets []engine.Address,
	values []*big.Int,
	signatures []string,
	calldatas []string,
	description string,
	blockNumber uint64,
	txHash engine.Hash,
) (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Check if governance is paused
	if g.Paused {
		return 0, ErrGovernancePaused
	}
	
	// Validate input
	if err := g.validateCreateProposalInput(proposer, targets, values, signatures, calldatas, description); err != nil {
		return 0, err
	}
	
	// Check if proposer has sufficient voting power
	if err := g.checkProposalThreshold(proposer); err != nil {
		return 0, err
	}
	
	// Create new proposal
	proposalID := g.TotalProposals
	proposal := &Proposal{
		ProposalID:      proposalID,
		Proposer:        proposer,
		Targets:         make([]engine.Address, len(targets)),
		Values:          make([]*big.Int, len(values)),
		Signatures:      make([]string, len(signatures)),
		Calldatas:      make([]string, len(calldatas)),
		Description:     description,
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(g.VotingPeriod),
		Executed:       false,
		Canceled:       false,
		ForVotes:       big.NewInt(0),
		AgainstVotes:   big.NewInt(0),
		AbstainVotes:   big.NewInt(0),
		TotalVotes:     big.NewInt(0),
		QuorumReached:  false,
		Votes:          make(map[engine.Address]*Vote),
		State:          ProposalStatePending,
	}
	
	// Copy slices to avoid external modifications
	copy(proposal.Targets, targets)
	for i, value := range values {
		proposal.Values[i] = new(big.Int).Set(value)
	}
	copy(proposal.Signatures, signatures)
	copy(proposal.Calldatas, calldatas)
	
	g.Proposals[proposalID] = proposal
	g.TotalProposals++
	
	// Get or create user
	if g.Users[proposer] == nil {
		g.Users[proposer] = &User{
			Address:         proposer,
			VotingPower:     big.NewInt(0),
			DelegatedTo:     engine.Address{},
			Delegators:      make(map[engine.Address]*big.Int),
			LastVote:        time.Now(),
			ProposalsCreated: 0,
		}
	}
	g.Users[proposer].ProposalsCreated++
	
	// Record event
	event := ProposalCreatedEvent{
		ProposalID:  proposalID,
		Proposer:    proposer,
		Targets:     make([]engine.Address, len(targets)),
		Values:      make([]*big.Int, len(values)),
		Signatures:  make([]string, len(signatures)),
		Calldatas:  make([]string, len(calldatas)),
		Description: description,
		StartTime:  proposal.StartTime,
		EndTime:    proposal.EndTime,
		Timestamp:  time.Now(),
		BlockNumber: blockNumber,
		TxHash:     txHash,
	}
	copy(event.Targets, targets)
	for i, value := range values {
		event.Values[i] = new(big.Int).Set(value)
	}
	copy(event.Signatures, signatures)
	copy(event.Calldatas, calldatas)
	
	g.ProposalCreatedEvents = append(g.ProposalCreatedEvents, event)
	
	return proposalID, nil
}

// CastVote casts a vote on a proposal
func (g *Governance) CastVote(
	voter engine.Address,
	proposalID uint64,
	support VoteSupport,
	reason string,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Check if governance is paused
	if g.Paused {
		return ErrGovernancePaused
	}
	
	// Validate input
	if err := g.validateCastVoteInput(voter, proposalID, support); err != nil {
		return err
	}
	
	// Get proposal
	proposal := g.Proposals[proposalID]
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	// Check if proposal is active
	if proposal.State != ProposalStateActive {
		return ErrProposalNotActive
	}
	
	// Check if voting period is still open
	if time.Now().After(proposal.EndTime) {
		return ErrVotingPeriodClosed
	}
	
	// Check if user has already voted
	if _, exists := proposal.Votes[voter]; exists {
		return ErrAlreadyVoted
	}
	
	// Get user voting power
	votingPower := g.getVotingPower(voter)
	if votingPower.Sign() == 0 {
		return ErrInsufficientVotingPower
	}
	
	// Create vote
	vote := &Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Support:    support,
		Votes:      new(big.Int).Set(votingPower),
		Reason:     reason,
		Timestamp:  time.Now(),
	}
	
	proposal.Votes[voter] = vote
	
	// Update proposal totals
	switch support {
	case VoteSupportFor:
		proposal.ForVotes = new(big.Int).Add(proposal.ForVotes, votingPower)
	case VoteSupportAgainst:
		proposal.AgainstVotes = new(big.Int).Add(proposal.AgainstVotes, votingPower)
	case VoteSupportAbstain:
		proposal.AbstainVotes = new(big.Int).Add(proposal.AbstainVotes, votingPower)
	}
	
	proposal.TotalVotes = new(big.Int).Add(proposal.TotalVotes, votingPower)
	
	// Check if quorum is reached
	if proposal.TotalVotes.Cmp(g.MinQuorum) >= 0 {
		proposal.QuorumReached = true
	}
	
	// Update proposal state
	g.updateProposalState(proposal)
	
	g.TotalVotes++
	
	// Update user last vote time
	if g.Users[voter] == nil {
		g.Users[voter] = &User{
			Address:         voter,
			VotingPower:     big.NewInt(0),
			DelegatedTo:     engine.Address{},
			Delegators:      make(map[engine.Address]*big.Int),
			LastVote:        time.Now(),
			ProposalsCreated: 0,
		}
	}
	g.Users[voter].LastVote = time.Now()
	
	// Record event
	event := VoteCastEvent{
		Voter:       voter,
		ProposalID:  proposalID,
		Support:     support,
		Votes:       new(big.Int).Set(votingPower),
		Reason:      reason,
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	g.VoteCastEvents = append(g.VoteCastEvents, event)
	
	return nil
}

// ExecuteProposal executes a successful proposal
func (g *Governance) ExecuteProposal(
	proposalID uint64,
	executor engine.Address,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Check if governance is paused
	if g.Paused {
		return ErrGovernancePaused
	}
	
	// Validate input
	if err := g.validateExecuteProposalInput(proposalID, executor); err != nil {
		return err
	}
	
	// Get proposal
	proposal := g.Proposals[proposalID]
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	// Check if proposal can be executed
	if err := g.checkProposalExecution(proposal); err != nil {
		return err
	}
	
	// Execute proposal
	proposal.Executed = true
	proposal.State = ProposalStateExecuted
	
	// Record event
	event := ProposalExecutedEvent{
		ProposalID:  proposalID,
		Executor:    executor,
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	g.ProposalExecutedEvents = append(g.ProposalExecutedEvents, event)
	
	return nil
}

// CancelProposal cancels a proposal (only proposer can cancel)
func (g *Governance) CancelProposal(
	proposalID uint64,
	canceler engine.Address,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Check if governance is paused
	if g.Paused {
		return ErrGovernancePaused
	}
	
	// Validate input
	if err := g.validateCancelProposalInput(proposalID, canceler); err != nil {
		return err
	}
	
	// Get proposal
	proposal := g.Proposals[proposalID]
	if proposal == nil {
		return ErrProposalNotFound
	}
	
	// Check if canceler is the proposer
	if proposal.Proposer != canceler {
		return ErrNotProposer
	}
	
	// Check if proposal can be canceled
	if proposal.State != ProposalStatePending && proposal.State != ProposalStateActive {
		return ErrProposalCannotBeCanceled
	}
	
	// Cancel proposal
	proposal.Canceled = true
	proposal.State = ProposalStateCanceled
	
	// Record event
	event := ProposalCanceledEvent{
		ProposalID:  proposalID,
		Canceler:    canceler,
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	g.ProposalCanceledEvents = append(g.ProposalCanceledEvents, event)
	
	return nil
}

// GetProposalInfo returns proposal information
func (g *Governance) GetProposalInfo(proposalID uint64) *Proposal {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if proposal, exists := g.Proposals[proposalID]; exists {
		// Return a copy to avoid race conditions
		proposalCopy := &Proposal{
			ProposalID:      proposal.ProposalID,
			Proposer:        proposal.Proposer,
			Targets:         make([]engine.Address, len(proposal.Targets)),
			Values:          make([]*big.Int, len(proposal.Values)),
			Signatures:      make([]string, len(proposal.Signatures)),
			Calldatas:      make([]string, len(proposal.Calldatas)),
			Description:     proposal.Description,
			StartTime:      proposal.StartTime,
			EndTime:        proposal.EndTime,
			Executed:       proposal.Executed,
			Canceled:       proposal.Canceled,
			ForVotes:       new(big.Int).Set(proposal.ForVotes),
			AgainstVotes:   new(big.Int).Set(proposal.AgainstVotes),
			AbstainVotes:   new(big.Int).Set(proposal.AbstainVotes),
			TotalVotes:     new(big.Int).Set(proposal.TotalVotes),
			QuorumReached:  proposal.QuorumReached,
			Votes:          make(map[engine.Address]*Vote),
			State:          proposal.State,
		}
		
		copy(proposalCopy.Targets, proposal.Targets)
		for i, value := range proposal.Values {
			proposalCopy.Values[i] = new(big.Int).Set(value)
		}
		copy(proposalCopy.Signatures, proposal.Signatures)
		copy(proposalCopy.Calldatas, proposal.Calldatas)
		
		for voter, vote := range proposal.Votes {
			proposalCopy.Votes[voter] = &Vote{
				ProposalID: vote.ProposalID,
				Voter:      vote.Voter,
				Support:    vote.Support,
				Votes:      new(big.Int).Set(vote.Votes),
				Reason:     vote.Reason,
				Timestamp:  vote.Timestamp,
			}
		}
		
		return proposalCopy
	}
	
	return nil
}

// GetUserInfo returns user information
func (g *Governance) GetUserInfo(user engine.Address) *User {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if userInfo, exists := g.Users[user]; exists {
		// Return a copy to avoid race conditions
		userCopy := &User{
			Address:         userInfo.Address,
			VotingPower:     new(big.Int).Set(userInfo.VotingPower),
			DelegatedTo:     userInfo.DelegatedTo,
			Delegators:      make(map[engine.Address]*big.Int),
			LastVote:        userInfo.LastVote,
			ProposalsCreated: userInfo.ProposalsCreated,
		}
		
		for delegator, power := range userInfo.Delegators {
			userCopy.Delegators[delegator] = new(big.Int).Set(power)
		}
		
		return userCopy
	}
	
	return nil
}

// GetGovernanceStats returns governance statistics
func (g *Governance) GetGovernanceStats() (uint64, uint64, *big.Int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return g.TotalProposals,
		   g.TotalVotes,
		   new(big.Int).Set(g.MinQuorum)
}

// Pause pauses the governance system
func (g *Governance) Pause() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if g.Paused {
		return ErrGovernanceAlreadyPaused
	}
	
	g.Paused = true
	return nil
}

// Unpause resumes the governance system
func (g *Governance) Unpause() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if !g.Paused {
		return ErrGovernanceNotPaused
	}
	
	g.Paused = false
	return nil
}

// updateProposalState updates the state of a proposal
func (g *Governance) updateProposalState(proposal *Proposal) {
	now := time.Now()
	
	if proposal.Canceled {
		proposal.State = ProposalStateCanceled
		return
	}
	
	if proposal.Executed {
		proposal.State = ProposalStateExecuted
		return
	}
	
	if now.Before(proposal.EndTime) {
		proposal.State = ProposalStateActive
		return
	}
	
	// Voting period has ended
	if proposal.QuorumReached && proposal.ForVotes.Cmp(proposal.AgainstVotes) > 0 {
		proposal.State = ProposalStateSucceeded
	} else {
		proposal.State = ProposalStateDefeated
	}
}

// getVotingPower returns the voting power of a user
func (g *Governance) getVotingPower(user engine.Address) *big.Int {
	userInfo := g.Users[user]
	if userInfo == nil {
		return big.NewInt(0)
	}
	
	// For now, return a fixed voting power
	// In a real implementation, this would query the governance token balance
	return big.NewInt(1000) // 1000 voting power
}

// checkProposalThreshold checks if proposer has sufficient voting power
func (g *Governance) checkProposalThreshold(proposer engine.Address) error {
	votingPower := g.getVotingPower(proposer)
	if votingPower.Cmp(g.ProposalThreshold) < 0 {
		return ErrInsufficientProposalPower
	}
	return nil
}

// checkProposalExecution checks if a proposal can be executed
func (g *Governance) checkProposalExecution(proposal *Proposal) error {
	if proposal.State != ProposalStateSucceeded {
		return ErrProposalNotSucceeded
	}
	
	// Check execution delay
	if time.Now().Before(proposal.EndTime.Add(g.ExecutionDelay)) {
		return ErrExecutionDelayNotMet
	}
	
	return nil
}

// Event types
type ProposalCreatedEvent struct {
	ProposalID  uint64
	Proposer    engine.Address
	Targets     []engine.Address
	Values      []*big.Int
	Signatures  []string
	Calldatas  []string
	Description string
	StartTime  time.Time
	EndTime    time.Time
	Timestamp  time.Time
	BlockNumber uint64
	TxHash     engine.Hash
}

type VoteCastEvent struct {
	Voter       engine.Address
	ProposalID  uint64
	Support     VoteSupport
	Votes       *big.Int
	Reason      string
	Timestamp  time.Time
	BlockNumber uint64
	TxHash     engine.Hash
}

type ProposalExecutedEvent struct {
	ProposalID  uint64
	Executor    engine.Address
	Timestamp  time.Time
	BlockNumber uint64
	TxHash     engine.Hash
}

type ProposalCanceledEvent struct {
	ProposalID  uint64
	Canceler    engine.Address
	Timestamp  time.Time
	BlockNumber uint64
	TxHash     engine.Hash
}

// Validation functions
func (g *Governance) validateCreateProposalInput(
	proposer engine.Address,
	targets []engine.Address,
	values []*big.Int,
	signatures []string,
	calldatas []string,
	description string,
) error {
	if proposer == (engine.Address{}) {
		return ErrInvalidProposer
	}
	
	if len(targets) == 0 {
		return ErrInvalidTargets
	}
	
	if len(values) != len(targets) {
		return ErrInvalidValues
	}
	
	if len(signatures) != len(targets) {
		return ErrInvalidSignatures
	}
	
	if len(calldatas) != len(targets) {
		return ErrInvalidCalldatas
	}
	
	if description == "" {
		return ErrInvalidDescription
	}
	
	return nil
}

func (g *Governance) validateCastVoteInput(
	voter engine.Address,
	proposalID uint64,
	support VoteSupport,
) error {
	if voter == (engine.Address{}) {
		return ErrInvalidVoter
	}
	
	if support > VoteSupportAbstain {
		return ErrInvalidVoteSupport
	}
	
	return nil
}

func (g *Governance) validateExecuteProposalInput(
	proposalID uint64,
	executor engine.Address,
) error {
	if executor == (engine.Address{}) {
		return ErrInvalidExecutor
	}
	
	return nil
}

func (g *Governance) validateCancelProposalInput(
	proposalID uint64,
	canceler engine.Address,
) error {
	if canceler == (engine.Address{}) {
		return ErrInvalidCanceler
	}
	
	return nil
}
