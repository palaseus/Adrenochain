package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// VotingSystem handles governance voting mechanisms
type VotingSystem struct {
	proposals    map[string]*Proposal
	votes        map[string]map[string]*Vote
	votingPower  map[string]*big.Int
	quorum       *big.Int
	votingPeriod time.Duration
	mutex        sync.RWMutex
}

// Proposal represents a governance proposal
type Proposal struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Proposer       string         `json:"proposer"`
	ProposalType   ProposalType   `json:"proposal_type"`
	Status         ProposalStatus `json:"status"`
	VotingStart    time.Time      `json:"voting_start"`
	VotingEnd      time.Time      `json:"voting_end"`
	ExecutedAt     *time.Time     `json:"executed_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	QuorumRequired *big.Int       `json:"quorum_required"`
	MinVotingPower *big.Int       `json:"min_voting_power"`
}

// ProposalType represents the type of proposal
type ProposalType string

const (
	ProposalTypeParameterChange ProposalType = "parameter_change"
	ProposalTypeTreasury        ProposalType = "treasury"
	ProposalTypeUpgrade         ProposalType = "upgrade"
	ProposalTypeEmergency       ProposalType = "emergency"
	ProposalTypeGeneral         ProposalType = "general"
)

// ProposalStatus represents the status of a proposal
type ProposalStatus string

const (
	ProposalStatusDraft     ProposalStatus = "draft"
	ProposalStatusActive    ProposalStatus = "active"
	ProposalStatusPassed    ProposalStatus = "passed"
	ProposalStatusRejected  ProposalStatus = "rejected"
	ProposalStatusExecuted  ProposalStatus = "executed"
	ProposalStatusCancelled ProposalStatus = "cancelled"
)

// Vote represents a vote on a proposal
type Vote struct {
	ProposalID  string     `json:"proposal_id"`
	Voter       string     `json:"voter"`
	VoteChoice  VoteChoice `json:"vote_choice"`
	VotingPower *big.Int   `json:"voting_power"`
	Reason      string     `json:"reason,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
	IsDelegated bool       `json:"is_delegated"`
	Delegator   string     `json:"delegator,omitempty"`
}

// VoteChoice represents the choice in a vote
type VoteChoice string

const (
	VoteChoiceFor     VoteChoice = "for"
	VoteChoiceAgainst VoteChoice = "against"
	VoteChoiceAbstain VoteChoice = "abstain"
)

// NewVotingSystem creates a new voting system
func NewVotingSystem(quorum *big.Int, votingPeriod time.Duration) *VotingSystem {
	return &VotingSystem{
		proposals:    make(map[string]*Proposal),
		votes:        make(map[string]map[string]*Vote),
		votingPower:  make(map[string]*big.Int),
		quorum:       quorum,
		votingPeriod: votingPeriod,
	}
}

// CreateProposal creates a new governance proposal
func (vs *VotingSystem) CreateProposal(
	title string,
	description string,
	proposer string,
	proposalType ProposalType,
	quorumRequired *big.Int,
	minVotingPower *big.Int,
) (*Proposal, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Validate proposer has minimum voting power
	if vs.votingPower[proposer] == nil || vs.votingPower[proposer].Cmp(minVotingPower) < 0 {
		return nil, fmt.Errorf("proposer does not have minimum voting power required")
	}

	proposal := &Proposal{
		ID:             vs.generateProposalID(),
		Title:          title,
		Description:    description,
		Proposer:       proposer,
		ProposalType:   proposalType,
		Status:         ProposalStatusDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		QuorumRequired: quorumRequired,
		MinVotingPower: minVotingPower,
	}

	vs.proposals[proposal.ID] = proposal
	vs.votes[proposal.ID] = make(map[string]*Vote)

	return proposal, nil
}

// ActivateProposal activates a proposal for voting
func (vs *VotingSystem) ActivateProposal(proposalID string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusDraft {
		return fmt.Errorf("proposal status is %s, expected draft", proposal.Status)
	}

	// Set voting period
	proposal.VotingStart = time.Now()
	proposal.VotingEnd = proposal.VotingStart.Add(vs.votingPeriod)
	proposal.Status = ProposalStatusActive
	proposal.UpdatedAt = time.Now()

	return nil
}

// CastVote casts a vote on a proposal
func (vs *VotingSystem) CastVote(
	proposalID string,
	voter string,
	voteChoice VoteChoice,
	reason string,
) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusActive {
		return fmt.Errorf("proposal is not active for voting")
	}

	if time.Now().After(proposal.VotingEnd) {
		return fmt.Errorf("voting period has ended")
	}

	// Check if voter has already voted
	if _, alreadyVoted := vs.votes[proposalID][voter]; alreadyVoted {
		return fmt.Errorf("voter has already voted on this proposal")
	}

	// Get voter's voting power
	votingPower, exists := vs.votingPower[voter]
	if !exists || votingPower.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("voter has no voting power")
	}

	// Create vote
	vote := &Vote{
		ProposalID:  proposalID,
		Voter:       voter,
		VoteChoice:  voteChoice,
		VotingPower: votingPower,
		Reason:      reason,
		Timestamp:   time.Now(),
		IsDelegated: false,
	}

	vs.votes[proposalID][voter] = vote

	return nil
}

// DelegateVote delegates voting power to another address
func (vs *VotingSystem) DelegateVote(
	delegator string,
	delegate string,
	amount *big.Int,
) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Check if delegator has enough voting power
	currentPower, exists := vs.votingPower[delegator]
	if !exists || currentPower.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient voting power to delegate")
	}

	// Transfer voting power
	vs.votingPower[delegator] = new(big.Int).Sub(currentPower, amount)

	if vs.votingPower[delegate] == nil {
		vs.votingPower[delegate] = big.NewInt(0)
	}
	vs.votingPower[delegate] = new(big.Int).Add(vs.votingPower[delegate], amount)

	return nil
}

// CastDelegatedVote casts a vote using delegated voting power
func (vs *VotingSystem) CastDelegatedVote(
	proposalID string,
	delegate string,
	delegator string,
	voteChoice VoteChoice,
	reason string,
) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusActive {
		return fmt.Errorf("proposal is not active for voting")
	}

	if time.Now().After(proposal.VotingEnd) {
		return fmt.Errorf("voting period has ended")
	}

	// Check if delegate has already voted
	if _, alreadyVoted := vs.votes[proposalID][delegate]; alreadyVoted {
		return fmt.Errorf("delegate has already voted on this proposal")
	}

	// Get delegated voting power
	delegatedPower, exists := vs.votingPower[delegate]
	if !exists || delegatedPower.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("delegate has no delegated voting power")
	}

	// Create delegated vote
	vote := &Vote{
		ProposalID:  proposalID,
		Voter:       delegate,
		VoteChoice:  voteChoice,
		VotingPower: delegatedPower,
		Reason:      reason,
		Timestamp:   time.Now(),
		IsDelegated: true,
		Delegator:   delegator,
	}

	vs.votes[proposalID][delegate] = vote

	return nil
}

// FinalizeProposal finalizes a proposal after voting period ends
func (vs *VotingSystem) FinalizeProposal(proposalID string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusActive {
		return fmt.Errorf("proposal is not active")
	}

	if time.Now().Before(proposal.VotingEnd) {
		return fmt.Errorf("voting period has not ended yet")
	}

	// Calculate voting results
	totalVotes, forVotes, againstVotes, _ := vs.calculateVotingResults(proposalID)

	// Check quorum
	if totalVotes.Cmp(proposal.QuorumRequired) < 0 {
		proposal.Status = ProposalStatusRejected
		proposal.UpdatedAt = time.Now()
		return nil
	}

	// Determine if proposal passed
	if forVotes.Cmp(againstVotes) > 0 {
		proposal.Status = ProposalStatusPassed
	} else {
		proposal.Status = ProposalStatusRejected
	}

	proposal.UpdatedAt = time.Now()

	return nil
}

// ExecuteProposal executes a passed proposal
func (vs *VotingSystem) ExecuteProposal(proposalID string, executor string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusPassed {
		return fmt.Errorf("proposal status is %s, expected passed", proposal.Status)
	}

	// Execute the proposal (implementation would depend on proposal type)
	proposal.Status = ProposalStatusExecuted
	proposal.ExecutedAt = &time.Time{}
	*proposal.ExecutedAt = time.Now()
	proposal.UpdatedAt = time.Now()

	return nil
}

// GetProposal returns a proposal by ID
func (vs *VotingSystem) GetProposal(proposalID string) (*Proposal, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	proposal, exists := vs.proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	return proposal, nil
}

// GetProposalsByStatus returns proposals by status
func (vs *VotingSystem) GetProposalsByStatus(status ProposalStatus) []*Proposal {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	var result []*Proposal
	for _, proposal := range vs.proposals {
		if proposal.Status == status {
			result = append(result, proposal)
		}
	}

	return result
}

// GetVotesForProposal returns all votes for a proposal
func (vs *VotingSystem) GetVotesForProposal(proposalID string) []*Vote {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	votes, exists := vs.votes[proposalID]
	if !exists {
		return nil
	}

	var result []*Vote
	for _, vote := range votes {
		result = append(result, vote)
	}

	return result
}

// GetVotingPower returns the voting power of an address
func (vs *VotingSystem) GetVotingPower(address string) *big.Int {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	power, exists := vs.votingPower[address]
	if !exists {
		return big.NewInt(0)
	}

	return power
}

// SetVotingPower sets the voting power of an address
func (vs *VotingSystem) SetVotingPower(address string, power *big.Int) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	vs.votingPower[address] = power
}

// GetVotingStats returns voting statistics
func (vs *VotingSystem) GetVotingStats() map[string]interface{} {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_proposals"] = len(vs.proposals)
	stats["draft_proposals"] = len(vs.GetProposalsByStatus(ProposalStatusDraft))
	stats["active_proposals"] = len(vs.GetProposalsByStatus(ProposalStatusActive))
	stats["passed_proposals"] = len(vs.GetProposalsByStatus(ProposalStatusPassed))
	stats["rejected_proposals"] = len(vs.GetProposalsByStatus(ProposalStatusRejected))
	stats["executed_proposals"] = len(vs.GetProposalsByStatus(ProposalStatusExecuted))
	stats["total_voters"] = len(vs.votingPower)

	return stats
}

// calculateVotingResults calculates the voting results for a proposal
func (vs *VotingSystem) calculateVotingResults(proposalID string) (*big.Int, *big.Int, *big.Int, *big.Int) {
	votes, exists := vs.votes[proposalID]
	if !exists {
		return big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)
	}

	totalVotes := big.NewInt(0)
	forVotes := big.NewInt(0)
	againstVotes := big.NewInt(0)
	abstainVotes := big.NewInt(0)

	for _, vote := range votes {
		totalVotes.Add(totalVotes, vote.VotingPower)

		switch vote.VoteChoice {
		case VoteChoiceFor:
			forVotes.Add(forVotes, vote.VotingPower)
		case VoteChoiceAgainst:
			againstVotes.Add(againstVotes, vote.VotingPower)
		case VoteChoiceAbstain:
			abstainVotes.Add(abstainVotes, vote.VotingPower)
		}
	}

	return totalVotes, forVotes, againstVotes, abstainVotes
}

// generateProposalID generates a unique proposal ID
func (vs *VotingSystem) generateProposalID() string {
	data := fmt.Sprintf("proposal_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}
