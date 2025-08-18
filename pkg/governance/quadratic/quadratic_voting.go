package quadratic

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/security"
)

// VoteType represents the type of vote
type VoteType int

const (
	VoteTypeYes VoteType = iota
	VoteTypeNo
	VoteTypeAbstain
	VoteTypeCustom
)

// VoteType.String() returns the string representation
func (vt VoteType) String() string {
	switch vt {
	case VoteTypeYes:
		return "yes"
	case VoteTypeNo:
		return "no"
	case VoteTypeAbstain:
		return "abstain"
	case VoteTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ProposalStatus represents the status of a proposal
type ProposalStatus int

const (
	ProposalStatusDraft ProposalStatus = iota
	ProposalStatusActive
	ProposalStatusVoting
	ProposalStatusClosed
	ProposalStatusExecuted
	ProposalStatusRejected
	ProposalStatusExpired
)

// ProposalStatus.String() returns the string representation
func (ps ProposalStatus) String() string {
	switch ps {
	case ProposalStatusDraft:
		return "draft"
	case ProposalStatusActive:
		return "active"
	case ProposalStatusVoting:
		return "voting"
	case ProposalStatusClosed:
		return "closed"
	case ProposalStatusExecuted:
		return "executed"
	case ProposalStatusRejected:
		return "rejected"
	case ProposalStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// Proposal represents a governance proposal
type Proposal struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Creator         string                 `json:"creator"`
	Status          ProposalStatus         `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	VotingStart     time.Time              `json:"voting_start"`
	VotingEnd       time.Time              `json:"voting_end"`
	ExecutionDelay  time.Duration          `json:"execution_delay"`
	Quorum          *big.Int               `json:"quorum"`
	Threshold       *big.Int               `json:"threshold"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Vote represents a quadratic vote
type Vote struct {
	ID           string                 `json:"id"`
	ProposalID   string                 `json:"proposal_id"`
	Voter        string                 `json:"voter"`
	VoteType     VoteType               `json:"vote_type"`
	VotePower    *big.Int               `json:"vote_power"`
	VoteCost     *big.Int               `json:"vote_cost"`
	Timestamp    time.Time              `json:"timestamp"`
	ZKProof      *security.ZKProof      `json:"zk_proof"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Voter represents a voter in the quadratic voting system
type Voter struct {
	ID              string                 `json:"id"`
	Address         string                 `json:"address"`
	VotingPower    *big.Int               `json:"voting_power"`
	UsedPower      *big.Int               `json:"used_power"`
	Reputation     *big.Int               `json:"reputation"`
	SybilResistance *big.Int              `json:"sybil_resistance"`
	JoinedAt       time.Time              `json:"joined_at"`
	LastVote       time.Time              `json:"last_vote"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// VotingResult represents the result of a proposal vote
type VotingResult struct {
	ProposalID     string                 `json:"proposal_id"`
	TotalVotes     uint64                 `json:"total_votes"`
	YesVotes       *big.Int               `json:"yes_votes"`
	NoVotes        *big.Int               `json:"no_votes"`
	AbstainVotes   *big.Int               `json:"abstain_votes"`
	TotalPower     *big.Int               `json:"total_power"`
	QuorumReached  bool                   `json:"quorum_reached"`
	ThresholdMet   bool                   `json:"threshold_met"`
	Passed         bool                   `json:"passed"`
	FinalizedAt    time.Time              `json:"finalized_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// SybilResistance represents sybil resistance mechanisms
type SybilResistance struct {
	ID              string                 `json:"id"`
	VoterID         string                 `json:"voter_id"`
	ProofOfWork     []byte                 `json:"proof_of_work"`
	SocialGraph     []string               `json:"social_graph"`
	ReputationScore *big.Int               `json:"reputation_score"`
	VerifiedAt      time.Time              `json:"verified_at"`
	Status          SybilStatus            `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SybilStatus represents the status of sybil resistance verification
type SybilStatus int

const (
	SybilStatusPending SybilStatus = iota
	SybilStatusVerified
	SybilStatusRejected
	SybilStatusSuspicious
)

// SybilStatus.String() returns the string representation
func (ss SybilStatus) String() string {
	switch ss {
	case SybilStatusPending:
		return "pending"
	case SybilStatusVerified:
		return "verified"
	case SybilStatusRejected:
		return "rejected"
	case SybilStatusSuspicious:
		return "suspicious"
	default:
		return "unknown"
	}
}

// QuadraticVotingConfig represents configuration for the Quadratic Voting system
type QuadraticVotingConfig struct {
	MaxProposals       uint64        `json:"max_proposals"`
	MaxVoters          uint64        `json:"max_voters"`
	MaxVotes           uint64        `json:"max_votes"`
	EncryptionKeySize  int           `json:"encryption_key_size"`
	ZKProofType        security.ProofType `json:"zk_proof_type"`
	VotingTimeout      time.Duration `json:"voting_timeout"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	MinQuorum          *big.Int      `json:"min_quorum"`
	DefaultThreshold   *big.Int      `json:"default_threshold"`
	SybilResistanceEnabled bool      `json:"sybil_resistance_enabled"`
}

// QuadraticVoting represents the main Quadratic Voting system
type QuadraticVoting struct {
	mu              sync.RWMutex
	Proposals       map[string]*Proposal           `json:"proposals"`
	Votes           map[string]*Vote               `json:"votes"`
	Voters          map[string]*Voter              `json:"voters"`
	Results         map[string]*VotingResult       `json:"results"`
	SybilResistance map[string]*SybilResistance    `json:"sybil_resistance"`
	Config          QuadraticVotingConfig          `json:"config"`
	encryptionKey   []byte
	running         bool
	stopChan        chan struct{}
}

// NewQuadraticVoting creates a new Quadratic Voting system
func NewQuadraticVoting(config QuadraticVotingConfig) *QuadraticVoting {
	// Set default values if not provided
	if config.MaxProposals == 0 {
		config.MaxProposals = 1000
	}
	if config.MaxVoters == 0 {
		config.MaxVoters = 10000
	}
	if config.MaxVotes == 0 {
		config.MaxVotes = 50000
	}
	if config.EncryptionKeySize == 0 {
		config.EncryptionKeySize = 32 // 256 bits
	}
	if config.ZKProofType == 0 {
		config.ZKProofType = security.ProofTypeBulletproofs
	}
	if config.VotingTimeout == 0 {
		config.VotingTimeout = time.Hour * 24 * 7 // 1 week
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Hour
	}
	if config.MinQuorum == nil {
		config.MinQuorum = big.NewInt(1000)
	}
	if config.DefaultThreshold == nil {
		config.DefaultThreshold = big.NewInt(5000)
	}

	// Generate encryption key
	encryptionKey := make([]byte, config.EncryptionKeySize)
	if _, err := rand.Read(encryptionKey); err != nil {
		panic(fmt.Sprintf("Failed to generate encryption key: %v", err))
	}

	return &QuadraticVoting{
		Proposals:       make(map[string]*Proposal),
		Votes:           make(map[string]*Vote),
		Voters:          make(map[string]*Voter),
		Results:         make(map[string]*VotingResult),
		SybilResistance: make(map[string]*SybilResistance),
		Config:          config,
		encryptionKey:   encryptionKey,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the Quadratic Voting system operations
func (qv *QuadraticVoting) Start() error {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	if qv.running {
		return fmt.Errorf("Quadratic Voting system is already running")
	}

	qv.running = true

	// Start background goroutines
	go qv.proposalManagementLoop()
	go qv.votingLoop()
	go qv.sybilResistanceLoop()
	go qv.cleanupLoop()

	return nil
}

// Stop halts all Quadratic Voting system operations
func (qv *QuadraticVoting) Stop() error {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	if !qv.running {
		return fmt.Errorf("Quadratic Voting system is not running")
	}

	close(qv.stopChan)
	qv.running = false

	return nil
}

// CreateProposal creates a new governance proposal
func (qv *QuadraticVoting) CreateProposal(
	title string,
	description string,
	creator string,
	votingStart time.Time,
	votingEnd time.Time,
	executionDelay time.Duration,
	quorum *big.Int,
	threshold *big.Int,
	metadata map[string]interface{},
) (*Proposal, error) {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	// Check limits
	if uint64(len(qv.Proposals)) >= qv.Config.MaxProposals {
		return nil, fmt.Errorf("proposal limit reached")
	}

	// Validate parameters
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if creator == "" {
		return nil, fmt.Errorf("creator cannot be empty")
	}
	if votingStart.After(votingEnd) {
		return nil, fmt.Errorf("voting start must be before voting end")
	}
	if quorum.Cmp(qv.Config.MinQuorum) < 0 {
		return nil, fmt.Errorf("quorum must be at least %s", qv.Config.MinQuorum.String())
	}
	if threshold == nil {
		threshold = qv.Config.DefaultThreshold
	}

	proposal := &Proposal{
		ID:             generateProposalID(),
		Title:          title,
		Description:    description,
		Creator:        creator,
		Status:         ProposalStatusDraft,
		CreatedAt:      time.Now(),
		VotingStart:    votingStart,
		VotingEnd:      votingEnd,
		ExecutionDelay: executionDelay,
		Quorum:         quorum,
		Threshold:      threshold,
		Metadata:       metadata,
	}

	qv.Proposals[proposal.ID] = proposal
	return proposal, nil
}

// ActivateProposal activates a proposal for voting
func (qv *QuadraticVoting) ActivateProposal(proposalID string) error {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusDraft {
		return fmt.Errorf("proposal is not in draft status: %s", proposal.Status.String())
	}

	if time.Now().Before(proposal.VotingStart) {
		return fmt.Errorf("voting has not started yet")
	}

	proposal.Status = ProposalStatusActive
	return nil
}

// StartVoting starts the voting period for a proposal
func (qv *QuadraticVoting) StartVoting(proposalID string) error {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusActive {
		return fmt.Errorf("proposal is not active: %s", proposal.Status.String())
	}

	if time.Now().Before(proposal.VotingStart) {
		return fmt.Errorf("voting has not started yet")
	}

	proposal.Status = ProposalStatusVoting
	return nil
}

// RegisterVoter registers a new voter in the system
func (qv *QuadraticVoting) RegisterVoter(
	address string,
	votingPower *big.Int,
	metadata map[string]interface{},
) (*Voter, error) {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	// Check limits
	if uint64(len(qv.Voters)) >= qv.Config.MaxVoters {
		return nil, fmt.Errorf("voter limit reached")
	}

	// Check if voter already exists
	for _, voter := range qv.Voters {
		if voter.Address == address {
			return nil, fmt.Errorf("voter already exists: %s", address)
		}
	}

	// Validate parameters
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if votingPower.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("voting power must be positive")
	}

	// Generate sybil resistance score
	sybilResistance := qv.generateSybilResistanceScore(address, votingPower)

	voter := &Voter{
		ID:              generateVoterID(),
		Address:         address,
		VotingPower:    votingPower,
		UsedPower:      big.NewInt(0),
		Reputation:     big.NewInt(100), // Start with base reputation
		SybilResistance: sybilResistance,
		JoinedAt:       time.Now(),
		Metadata:       metadata,
	}

	qv.Voters[voter.ID] = voter

	// Create sybil resistance record if enabled
	if qv.Config.SybilResistanceEnabled {
		sybilRecord := &SybilResistance{
			ID:              generateSybilResistanceID(),
			VoterID:         voter.ID,
			ProofOfWork:     qv.generateProofOfWork(address),
			SocialGraph:     qv.generateSocialGraph(address),
			ReputationScore: sybilResistance,
			VerifiedAt:      time.Now(),
			Status:          SybilStatusVerified,
			Metadata:        make(map[string]interface{}),
		}
		qv.SybilResistance[sybilRecord.ID] = sybilRecord
	}

	return voter, nil
}

// CastVote casts a quadratic vote on a proposal
func (qv *QuadraticVoting) CastVote(
	proposalID string,
	voterAddress string,
	voteType VoteType,
	votePower *big.Int,
	metadata map[string]interface{},
) (*Vote, error) {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	// Check limits
	if uint64(len(qv.Votes)) >= qv.Config.MaxVotes {
		return nil, fmt.Errorf("vote limit reached")
	}

	// Validate proposal
	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusVoting {
		return nil, fmt.Errorf("proposal is not in voting status: %s", proposal.Status.String())
	}

	if time.Now().After(proposal.VotingEnd) {
		return nil, fmt.Errorf("voting has ended")
	}

	// Find voter
	var voter *Voter
	for _, v := range qv.Voters {
		if v.Address == voterAddress {
			voter = v
			break
		}
	}

	if voter == nil {
		return nil, fmt.Errorf("voter not found: %s", voterAddress)
	}

	// Check if voter has already voted on this proposal
	for _, vote := range qv.Votes {
		if vote.ProposalID == proposalID && vote.Voter == voterAddress {
			return nil, fmt.Errorf("voter has already voted on this proposal")
		}
	}

	// Validate vote power
	if votePower.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("vote power must be positive")
	}

	// Check if voter has enough voting power
	availablePower := new(big.Int).Sub(voter.VotingPower, voter.UsedPower)
	if votePower.Cmp(availablePower) > 0 {
		return nil, fmt.Errorf("insufficient voting power: %s < %s", availablePower.String(), votePower.String())
	}

	// Calculate vote cost (quadratic formula: cost = power^2)
	voteCost := new(big.Int).Mul(votePower, votePower)

	// Generate ZK proof for the vote
	zkProver := security.NewZKProver(qv.Config.ZKProofType)
	statement := []byte(fmt.Sprintf("%s:%s:%s:%s", proposalID, voterAddress, voteType.String(), votePower.String()))
	witness := []byte(fmt.Sprintf("%s", voterAddress))
	zkProof, err := zkProver.GenerateProof(statement, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ZK proof: %v", err)
	}

	vote := &Vote{
		ID:         generateVoteID(),
		ProposalID: proposalID,
		Voter:      voterAddress,
		VoteType:   voteType,
		VotePower:  votePower,
		VoteCost:   voteCost,
		Timestamp:  time.Now(),
		ZKProof:    zkProof,
		Metadata:   metadata,
	}

	qv.Votes[vote.ID] = vote

	// Update voter's used power
	voter.UsedPower.Add(voter.UsedPower, votePower)
	voter.LastVote = time.Now()

	return vote, nil
}

// CloseVoting closes voting for a proposal and calculates results
func (qv *QuadraticVoting) CloseVoting(proposalID string) (*VotingResult, error) {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusVoting {
		return nil, fmt.Errorf("proposal is not in voting status: %s", proposal.Status.String())
	}

	// Calculate voting results
	result := qv.calculateVotingResult(proposalID)
	if result == nil {
		return nil, fmt.Errorf("failed to calculate voting result")
	}

	// Update proposal status
	if result.Passed {
		proposal.Status = ProposalStatusClosed
	} else {
		proposal.Status = ProposalStatusRejected
	}

	// Store result
	qv.Results[proposalID] = result

	return result, nil
}

// ExecuteProposal executes a passed proposal
func (qv *QuadraticVoting) ExecuteProposal(proposalID string) error {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != ProposalStatusClosed {
		return fmt.Errorf("proposal is not closed: %s", proposal.Status.String())
	}

	result, exists := qv.Results[proposalID]
	if !exists {
		return fmt.Errorf("voting result not found for proposal: %s", proposalID)
	}

	if !result.Passed {
		return fmt.Errorf("proposal did not pass: %s", proposalID)
	}

	// Check execution delay
	if time.Since(result.FinalizedAt) < proposal.ExecutionDelay {
		return fmt.Errorf("execution delay not met: %s", proposalID)
	}

	proposal.Status = ProposalStatusExecuted
	return nil
}

// GetProposal retrieves a proposal by ID
func (qv *QuadraticVoting) GetProposal(proposalID string) (*Proposal, error) {
	qv.mu.RLock()
	defer qv.mu.RUnlock()

	proposal, exists := qv.Proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	return qv.copyProposal(proposal), nil
}

// GetVotingResult retrieves the voting result for a proposal
func (qv *QuadraticVoting) GetVotingResult(proposalID string) (*VotingResult, error) {
	qv.mu.RLock()
	defer qv.mu.RUnlock()

	result, exists := qv.Results[proposalID]
	if !exists {
		return nil, fmt.Errorf("voting result not found for proposal: %s", proposalID)
	}

	return qv.copyVotingResult(result), nil
}

// GetVoter retrieves a voter by address
func (qv *QuadraticVoting) GetVoter(address string) (*Voter, error) {
	qv.mu.RLock()
	defer qv.mu.RUnlock()

	for _, voter := range qv.Voters {
		if voter.Address == address {
			return qv.copyVoter(voter), nil
		}
	}

	return nil, fmt.Errorf("voter not found: %s", address)
}

// Background loops
func (qv *QuadraticVoting) proposalManagementLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-qv.stopChan:
			return
		case <-ticker.C:
			qv.manageProposals()
		}
	}
}

func (qv *QuadraticVoting) votingLoop() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-qv.stopChan:
			return
		case <-ticker.C:
			qv.processVoting()
		}
	}
}

func (qv *QuadraticVoting) sybilResistanceLoop() {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for {
		select {
		case <-qv.stopChan:
			return
		case <-ticker.C:
			qv.updateSybilResistance()
		}
	}
}

func (qv *QuadraticVoting) cleanupLoop() {
	ticker := time.NewTicker(qv.Config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-qv.stopChan:
			return
		case <-ticker.C:
			qv.cleanupOldData()
		}
	}
}

// Helper functions for background loops
func (qv *QuadraticVoting) manageProposals() {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	now := time.Now()

	for _, proposal := range qv.Proposals {
		switch proposal.Status {
		case ProposalStatusDraft:
			if now.After(proposal.VotingStart) {
				proposal.Status = ProposalStatusActive
			}
		case ProposalStatusActive:
			if now.After(proposal.VotingStart) {
				proposal.Status = ProposalStatusVoting
			}
		case ProposalStatusVoting:
			if now.After(proposal.VotingEnd) {
				proposal.Status = ProposalStatusClosed
			}
		}
	}
}

func (qv *QuadraticVoting) processVoting() {
	qv.mu.RLock()
	var activeProposals []string
	for id, proposal := range qv.Proposals {
		if proposal.Status == ProposalStatusVoting && time.Now().After(proposal.VotingEnd) {
			activeProposals = append(activeProposals, id)
		}
	}
	qv.mu.RUnlock()

	for _, proposalID := range activeProposals {
		if _, err := qv.CloseVoting(proposalID); err != nil {
			// Log error but continue processing other proposals
			continue
		}
	}
}

func (qv *QuadraticVoting) updateSybilResistance() {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	for _, voter := range qv.Voters {
		// Update reputation based on voting behavior
		if voter.LastVote.After(time.Now().Add(-time.Hour*24*30)) { // Last 30 days
			voter.Reputation.Add(voter.Reputation, big.NewInt(1))
		}

		// Update sybil resistance score
		voter.SybilResistance = qv.generateSybilResistanceScore(voter.Address, voter.VotingPower)
	}
}

func (qv *QuadraticVoting) cleanupOldData() {
	qv.mu.Lock()
	defer qv.mu.Unlock()

	cutoffTime := time.Now().Add(-qv.Config.VotingTimeout)

	// Clean up old proposals
	for id, proposal := range qv.Proposals {
		if proposal.Status == ProposalStatusRejected && proposal.CreatedAt.Before(cutoffTime) {
			delete(qv.Proposals, id)
		}
	}

	// Clean up old votes
	for id, vote := range qv.Votes {
		if vote.Timestamp.Before(cutoffTime) {
			delete(qv.Votes, id)
		}
	}
}

// Helper functions
func (qv *QuadraticVoting) calculateVotingResult(proposalID string) *VotingResult {
	proposal := qv.Proposals[proposalID]
	if proposal == nil {
		return nil
	}

	var totalVotes uint64
	var yesVotes, noVotes, abstainVotes, totalPower *big.Int = big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)

	for _, vote := range qv.Votes {
		if vote.ProposalID == proposalID {
			totalVotes++
			totalPower.Add(totalPower, vote.VotePower)

			switch vote.VoteType {
			case VoteTypeYes:
				yesVotes.Add(yesVotes, vote.VotePower)
			case VoteTypeNo:
				noVotes.Add(noVotes, vote.VotePower)
			case VoteTypeAbstain:
				abstainVotes.Add(abstainVotes, vote.VotePower)
			}
		}
	}

	quorumReached := totalPower.Cmp(proposal.Quorum) >= 0
	thresholdMet := yesVotes.Cmp(proposal.Threshold) >= 0
	passed := quorumReached && thresholdMet

	return &VotingResult{
		ProposalID:    proposalID,
		TotalVotes:    totalVotes,
		YesVotes:      yesVotes,
		NoVotes:       noVotes,
		AbstainVotes:  abstainVotes,
		TotalPower:    totalPower,
		QuorumReached: quorumReached,
		ThresholdMet:  thresholdMet,
		Passed:        passed,
		FinalizedAt:   time.Now(),
		Metadata:      make(map[string]interface{}),
	}
}

func (qv *QuadraticVoting) generateSybilResistanceScore(address string, votingPower *big.Int) *big.Int {
	// Simple sybil resistance based on address hash and voting power
	hash := sha256.Sum256([]byte(address))
	hashInt := new(big.Int).SetBytes(hash[:])
	
	// Combine with voting power for final score
	score := new(big.Int).Add(hashInt, votingPower)
	return score.Mod(score, big.NewInt(1000)) // Normalize to 0-999
}

func (qv *QuadraticVoting) generateProofOfWork(address string) []byte {
	// Simple proof of work simulation
	hash := sha256.Sum256([]byte(address + "proof"))
	return hash[:]
}

func (qv *QuadraticVoting) generateSocialGraph(address string) []string {
	// Simple social graph simulation
	return []string{fmt.Sprintf("connection_%s_1", address), fmt.Sprintf("connection_%s_2", address)}
}

// Deep copy functions
func (qv *QuadraticVoting) copyProposal(proposal *Proposal) *Proposal {
	if proposal == nil {
		return nil
	}

	copied := *proposal
	copied.Metadata = qv.copyMap(proposal.Metadata)
	return &copied
}

func (qv *QuadraticVoting) copyVote(vote *Vote) *Vote {
	if vote == nil {
		return nil
	}

	copied := *vote
	if vote.ZKProof != nil {
		copied.ZKProof = &security.ZKProof{
			Type:            vote.ZKProof.Type,
			Proof:           make([]byte, len(vote.ZKProof.Proof)),
			PublicInputs:    make([]byte, len(vote.ZKProof.PublicInputs)),
			VerificationKey: make([]byte, len(vote.ZKProof.VerificationKey)),
			Timestamp:       vote.ZKProof.Timestamp,
		}
		copy(copied.ZKProof.Proof, vote.ZKProof.Proof)
		copy(copied.ZKProof.PublicInputs, vote.ZKProof.PublicInputs)
		copy(copied.ZKProof.VerificationKey, vote.ZKProof.VerificationKey)
	}

	copied.Metadata = qv.copyMap(vote.Metadata)
	return &copied
}

func (qv *QuadraticVoting) copyVoter(voter *Voter) *Voter {
	if voter == nil {
		return nil
	}

	copied := *voter
	copied.Metadata = qv.copyMap(voter.Metadata)
	return &copied
}

func (qv *QuadraticVoting) copyVotingResult(result *VotingResult) *VotingResult {
	if result == nil {
		return nil
	}

	copied := *result
	copied.Metadata = qv.copyMap(result.Metadata)
	return &copied
}

func (qv *QuadraticVoting) copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	copied := make(map[string]interface{})
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

// ID generation functions
func generateProposalID() string {
	return fmt.Sprintf("proposal_%d", time.Now().UnixNano())
}

func generateVoteID() string {
	return fmt.Sprintf("vote_%d", time.Now().UnixNano())
}

func generateVoterID() string {
	return fmt.Sprintf("voter_%d", time.Now().UnixNano())
}

func generateSybilResistanceID() string {
	return fmt.Sprintf("sybil_%d", time.Now().UnixNano())
}
