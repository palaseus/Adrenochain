package governance

import (
	"fmt"
	"math/big"
	"sync"
	"time"
)

// GovernanceCoordinator coordinates all governance components
type GovernanceCoordinator struct {
	votingSystem    *VotingSystem
	treasuryManager *TreasuryManager
	config          *GovernanceConfig
	mutex           sync.RWMutex
	eventHandlers   map[string][]func(interface{})
}

// GovernanceConfig represents governance configuration
type GovernanceConfig struct {
	Quorum                 *big.Int        `json:"quorum"`
	VotingPeriod           time.Duration   `json:"voting_period"`
	MaxTransactionAmount   *big.Int        `json:"max_transaction_amount"`
	DailyLimit             *big.Int        `json:"daily_limit"`
	MinProposalPower       *big.Int        `json:"min_proposal_power"`
	MultisigAddresses      []string        `json:"multisig_addresses"`
	RequiredSignatures     int             `json:"required_signatures"`
	EmergencyThreshold     *big.Int        `json:"emergency_threshold"`
	SnapshotInterval       time.Duration   `json:"snapshot_interval"`
}

// NewGovernanceCoordinator creates a new governance coordinator
func NewGovernanceCoordinator(config *GovernanceConfig) *GovernanceCoordinator {
	if config == nil {
		config = getDefaultGovernanceConfig()
	}

	coordinator := &GovernanceCoordinator{
		config:          config,
		eventHandlers:   make(map[string][]func(interface{})),
	}

	// Initialize voting system
	coordinator.votingSystem = NewVotingSystem(config.Quorum, config.VotingPeriod)

	// Initialize treasury manager
	coordinator.treasuryManager = NewTreasuryManager(
		config.MaxTransactionAmount,
		config.DailyLimit,
		config.MultisigAddresses,
		config.RequiredSignatures,
	)

	return coordinator
}

// getDefaultGovernanceConfig returns default governance configuration
func getDefaultGovernanceConfig() *GovernanceConfig {
	return &GovernanceConfig{
		Quorum:                 mustParseBigInt("1000000000000000000000"), // 1000 tokens
		VotingPeriod:           7 * 24 * time.Hour,                       // 7 days
		MaxTransactionAmount:   mustParseBigInt("100000000000000000000"),  // 100 tokens
		DailyLimit:             mustParseBigInt("1000000000000000000000"), // 1000 tokens
		MinProposalPower:       mustParseBigInt("100000000000000000000"),  // 100 tokens
		MultisigAddresses:      []string{},
		RequiredSignatures:     3,
		EmergencyThreshold:     mustParseBigInt("10000000000000000000000"), // 10000 tokens
		SnapshotInterval:       24 * time.Hour,                            // 24 hours
	}
}

// mustParseBigInt parses a string to big.Int, panics on error
func mustParseBigInt(s string) *big.Int {
	result, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to parse big.Int: " + s)
	}
	return result
}

// Governance Proposal Management

// CreateGovernanceProposal creates a new governance proposal
func (gc *GovernanceCoordinator) CreateGovernanceProposal(
	title string,
	description string,
	proposer string,
	proposalType ProposalType,
	quorumRequired *big.Int,
) (*Proposal, error) {
	// Check if proposer has minimum voting power
	minPower := gc.config.MinProposalPower
	if gc.votingSystem.GetVotingPower(proposer).Cmp(minPower) < 0 {
		return nil, fmt.Errorf("proposer does not have minimum voting power required")
	}

	return gc.votingSystem.CreateProposal(title, description, proposer, proposalType, quorumRequired, minPower)
}

// ActivateGovernanceProposal activates a governance proposal
func (gc *GovernanceCoordinator) ActivateGovernanceProposal(proposalID string) error {
	return gc.votingSystem.ActivateProposal(proposalID)
}

// VoteOnGovernanceProposal votes on a governance proposal
func (gc *GovernanceCoordinator) VoteOnGovernanceProposal(
	proposalID string,
	voter string,
	voteChoice VoteChoice,
	reason string,
) error {
	return gc.votingSystem.CastVote(proposalID, voter, voteChoice, reason)
}

// DelegateVotingPower delegates voting power
func (gc *GovernanceCoordinator) DelegateVotingPower(
	delegator string,
	delegate string,
	amount *big.Int,
) error {
	return gc.votingSystem.DelegateVote(delegator, delegate, amount)
}

// FinalizeGovernanceProposal finalizes a governance proposal
func (gc *GovernanceCoordinator) FinalizeGovernanceProposal(proposalID string) error {
	return gc.votingSystem.FinalizeProposal(proposalID)
}

// ExecuteGovernanceProposal executes a passed governance proposal
func (gc *GovernanceCoordinator) ExecuteGovernanceProposal(proposalID string, executor string) error {
	return gc.votingSystem.ExecuteProposal(proposalID, executor)
}

// Treasury Management

// CreateTreasuryProposal creates a new treasury proposal
func (gc *GovernanceCoordinator) CreateTreasuryProposal(
	title string,
	description string,
	proposer string,
	amount *big.Int,
	asset string,
	recipient string,
	purpose string,
) (*TreasuryProposal, error) {
	// Check if proposer has minimum voting power
	minPower := gc.config.MinProposalPower
	if gc.votingSystem.GetVotingPower(proposer).Cmp(minPower) < 0 {
		return nil, fmt.Errorf("proposer does not have minimum voting power required")
	}

	return gc.treasuryManager.CreateTreasuryProposal(title, description, proposer, amount, asset, recipient, purpose)
}

// ActivateTreasuryProposal activates a treasury proposal
func (gc *GovernanceCoordinator) ActivateTreasuryProposal(proposalID string) error {
	return gc.treasuryManager.ActivateTreasuryProposal(proposalID)
}

// VoteOnTreasuryProposal votes on a treasury proposal
func (gc *GovernanceCoordinator) VoteOnTreasuryProposal(
	proposalID string,
	voter string,
	voteChoice VoteChoice,
	votingPower *big.Int,
) error {
	return gc.treasuryManager.VoteOnTreasuryProposal(proposalID, voter, voteChoice, votingPower)
}

// FinalizeTreasuryProposal finalizes a treasury proposal
func (gc *GovernanceCoordinator) FinalizeTreasuryProposal(proposalID string) error {
	return gc.treasuryManager.FinalizeTreasuryProposal(proposalID)
}

// ExecuteTreasuryProposal executes a passed treasury proposal
func (gc *GovernanceCoordinator) ExecuteTreasuryProposal(proposalID string, executor string) (*TreasuryTransaction, error) {
	return gc.treasuryManager.ExecuteTreasuryProposal(proposalID, executor)
}

// CreateDirectTreasuryTransaction creates a direct treasury transaction
func (gc *GovernanceCoordinator) CreateDirectTreasuryTransaction(
	transactionType TreasuryTransactionType,
	amount *big.Int,
	asset string,
	to string,
	description string,
	executor string,
) (*TreasuryTransaction, error) {
	return gc.treasuryManager.CreateDirectTransaction(transactionType, amount, asset, to, description, executor)
}

// Treasury Operations

// GetTreasuryBalance returns the treasury balance for an asset
func (gc *GovernanceCoordinator) GetTreasuryBalance(asset string) *big.Int {
	return gc.treasuryManager.GetBalance(asset)
}

// SetTreasuryBalance sets the treasury balance for an asset
func (gc *GovernanceCoordinator) SetTreasuryBalance(asset string, amount *big.Int) {
	gc.treasuryManager.SetBalance(asset, amount)
}

// Voting Power Management

// SetVotingPower sets the voting power of an address
func (gc *GovernanceCoordinator) SetVotingPower(address string, power *big.Int) {
	gc.votingSystem.SetVotingPower(address, power)
}

// GetVotingPower returns the voting power of an address
func (gc *GovernanceCoordinator) GetVotingPower(address string) *big.Int {
	return gc.votingSystem.GetVotingPower(address)
}

// Query Methods

// GetGovernanceProposal returns a governance proposal by ID
func (gc *GovernanceCoordinator) GetGovernanceProposal(proposalID string) (*Proposal, error) {
	return gc.votingSystem.GetProposal(proposalID)
}

// GetGovernanceProposalsByStatus returns governance proposals by status
func (gc *GovernanceCoordinator) GetGovernanceProposalsByStatus(status ProposalStatus) []*Proposal {
	return gc.votingSystem.GetProposalsByStatus(status)
}

// GetTreasuryProposal returns a treasury proposal by ID
func (gc *GovernanceCoordinator) GetTreasuryProposal(proposalID string) (*TreasuryProposal, error) {
	return gc.treasuryManager.GetTreasuryProposal(proposalID)
}

// GetTreasuryProposalsByStatus returns treasury proposals by status
func (gc *GovernanceCoordinator) GetTreasuryProposalsByStatus(status TreasuryProposalStatus) []*TreasuryProposal {
	return gc.treasuryManager.GetTreasuryProposalsByStatus(status)
}

// GetVotesForProposal returns all votes for a governance proposal
func (gc *GovernanceCoordinator) GetVotesForProposal(proposalID string) []*Vote {
	return gc.votingSystem.GetVotesForProposal(proposalID)
}

// GetTreasuryTransactions returns treasury transactions
func (gc *GovernanceCoordinator) GetTreasuryTransactions(limit int) []*TreasuryTransaction {
	return gc.treasuryManager.GetTreasuryTransactions(limit)
}

// Statistics and Analytics

// GetGovernanceStats returns governance statistics
func (gc *GovernanceCoordinator) GetGovernanceStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Voting system stats
	votingStats := gc.votingSystem.GetVotingStats()
	for k, v := range votingStats {
		stats["voting_"+k] = v
	}

	// Treasury stats
	treasuryStats := gc.treasuryManager.GetTreasuryStats()
	for k, v := range treasuryStats {
		stats["treasury_"+k] = v
	}

	// Configuration stats
	stats["quorum"] = gc.config.Quorum.String()
	stats["voting_period_hours"] = int(gc.config.VotingPeriod.Hours())
	stats["max_transaction_amount"] = gc.config.MaxTransactionAmount.String()
	stats["daily_limit"] = gc.config.DailyLimit.String()
	stats["min_proposal_power"] = gc.config.MinProposalPower.String()
	stats["multisig_signers"] = len(gc.config.MultisigAddresses)
	stats["required_signatures"] = gc.config.RequiredSignatures

	return stats
}

// Emergency Operations

// EmergencyPause pauses governance operations
func (gc *GovernanceCoordinator) EmergencyPause(pausedBy string, reason string) error {
	// This would implement emergency pause logic
	// For now, we'll just emit an event
	gc.emitEvent("emergency_pause", map[string]interface{}{
		"paused_by": pausedBy,
		"reason":    reason,
		"timestamp": time.Now(),
	})

	return nil
}

// EmergencyResume resumes governance operations
func (gc *GovernanceCoordinator) EmergencyResume(resumedBy string) error {
	// This would implement emergency resume logic
	gc.emitEvent("emergency_resume", map[string]interface{}{
		"resumed_by": resumedBy,
		"timestamp":  time.Now(),
	})

	return nil
}

// Event System

// On registers an event handler
func (gc *GovernanceCoordinator) On(eventType string, handler func(interface{})) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	if gc.eventHandlers[eventType] == nil {
		gc.eventHandlers[eventType] = make([]func(interface{}), 0)
	}

	gc.eventHandlers[eventType] = append(gc.eventHandlers[eventType], handler)
}

// emitEvent emits an event to registered handlers
func (gc *GovernanceCoordinator) emitEvent(eventType string, data interface{}) {
	gc.mutex.RLock()
	handlers, exists := gc.eventHandlers[eventType]
	gc.mutex.RUnlock()

	if exists {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}

// Getter Methods

// GetVotingSystem returns the voting system
func (gc *GovernanceCoordinator) GetVotingSystem() *VotingSystem {
	return gc.votingSystem
}

// GetTreasuryManager returns the treasury manager
func (gc *GovernanceCoordinator) GetTreasuryManager() *TreasuryManager {
	return gc.treasuryManager
}

// GetConfig returns the governance configuration
func (gc *GovernanceCoordinator) GetConfig() *GovernanceConfig {
	return gc.config
}
