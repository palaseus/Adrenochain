package validators

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// MultiChainValidator represents a validator that can participate in multiple chains
type MultiChainValidator struct {
	ID         string
	Address    [20]byte
	PublicKey  []byte
	Stake      *big.Int
	Chains     map[string]*ChainParticipation
	TotalStake *big.Int
	Reputation uint64
	Status     ValidatorStatus
	CreatedAt  time.Time
	LastActive time.Time
	mu         sync.RWMutex
	config     ValidatorConfig
	metrics    ValidatorMetrics
}

// ChainParticipation represents validator participation in a specific chain
type ChainParticipation struct {
	ChainID         string
	Stake           *big.Int
	VotingPower     uint64
	LastVote        time.Time
	VoteHistory     []VoteRecord
	SlashingHistory []SlashingRecord
	Status          ParticipationStatus
}

// ValidatorStatus represents the overall status of a validator
type ValidatorStatus int

const (
	ValidatorStatusActive ValidatorStatus = iota
	ValidatorStatusInactive
	ValidatorStatusSlashed
	ValidatorStatusJailed
	ValidatorStatusUnbonding
)

// ParticipationStatus represents the status of participation in a specific chain
type ParticipationStatus int

const (
	ParticipationStatusActive ParticipationStatus = iota
	ParticipationStatusInactive
	ParticipationStatusSlashed
	ParticipationStatusUnbonding
)

// ValidatorConfig holds configuration for validators
type ValidatorConfig struct {
	MinStake           *big.Int
	MaxStake           *big.Int
	UnbondingPeriod    time.Duration
	SlashingThreshold  uint64
	ReputationDecay    time.Duration
	EnableAutoRotation bool
	SecurityLevel      SecurityLevel
}

// SecurityLevel defines the security level for validator operations
type SecurityLevel int

const (
	SecurityLevelLow SecurityLevel = iota
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelUltra
)

// ValidatorMetrics tracks validator performance metrics
type ValidatorMetrics struct {
	TotalVotes        uint64
	SuccessfulVotes   uint64
	FailedVotes       uint64
	TotalStake        *big.Int
	AverageReputation float64
	LastUpdate        time.Time
}

// VoteRecord represents a voting record
type VoteRecord struct {
	BlockHash   [32]byte
	BlockHeight uint64
	Vote        bool
	Timestamp   time.Time
	ChainID     string
}

// SlashingRecord represents a slashing record
type SlashingRecord struct {
	Reason    string
	Amount    *big.Int
	Timestamp time.Time
	ChainID   string
	Evidence  []byte
}

// ValidatorNetwork represents a network of validators
type ValidatorNetwork struct {
	ID               string
	Validators       map[string]*MultiChainValidator
	TotalStake       *big.Int
	ActiveValidators uint64
	MinValidators    uint64
	ConsensusRules   ConsensusRules
	CreatedAt        time.Time
	mu               sync.RWMutex
	config           NetworkConfig
	metrics          NetworkMetrics
}

// ConsensusRules defines the consensus rules for the network
type ConsensusRules struct {
	QuorumSize        uint64
	FinalityThreshold uint64
	BlockTime         time.Duration
	MaxValidators     uint64
}

// NetworkConfig holds configuration for validator networks
type NetworkConfig struct {
	MinStake          *big.Int
	MaxValidators     uint64
	RotationInterval  time.Duration
	SlashingEnabled   bool
	ReputationEnabled bool
	SecurityLevel     SecurityLevel
}

// NetworkMetrics tracks network performance metrics
type NetworkMetrics struct {
	TotalValidators   uint64
	ActiveValidators  uint64
	TotalStake        *big.Int
	AverageReputation float64
	LastRotation      time.Time
	LastUpdate        time.Time
}

// NewMultiChainValidator creates a new multi-chain validator
func NewMultiChainValidator(address [20]byte, publicKey []byte, config ValidatorConfig) *MultiChainValidator {
	// Set default values if not provided
	if config.MinStake == nil {
		config.MinStake = big.NewInt(1000000) // 1M units
	}
	if config.MaxStake == nil {
		config.MaxStake = big.NewInt(1000000000) // 1B units
	}
	if config.UnbondingPeriod == 0 {
		config.UnbondingPeriod = time.Hour * 24 * 7 * 2 // 2 weeks
	}
	if config.SlashingThreshold == 0 {
		config.SlashingThreshold = 100
	}
	if config.ReputationDecay == 0 {
		config.ReputationDecay = time.Hour * 24 * 7 // 1 week
	}

	return &MultiChainValidator{
		ID:         generateValidatorID(),
		Address:    address,
		PublicKey:  publicKey,
		Stake:      big.NewInt(0),
		Chains:     make(map[string]*ChainParticipation),
		TotalStake: big.NewInt(0),
		Reputation: 1000, // Start with base reputation
		Status:     ValidatorStatusActive,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		config:     config,
		metrics: ValidatorMetrics{
			TotalStake: big.NewInt(0),
		},
	}
}

// NewValidatorNetwork creates a new validator network
func NewValidatorNetwork(config NetworkConfig) *ValidatorNetwork {
	// Set default values if not provided
	if config.MinStake == nil {
		config.MinStake = big.NewInt(1000000) // 1M units
	}
	if config.MaxValidators == 0 {
		config.MaxValidators = 100
	}
	if config.RotationInterval == 0 {
		config.RotationInterval = time.Hour * 24 * 7 // 1 week
	}

	return &ValidatorNetwork{
		ID:               generateNetworkID(),
		Validators:       make(map[string]*MultiChainValidator),
		TotalStake:       big.NewInt(0),
		ActiveValidators: 0,
		MinValidators:    config.MaxValidators / 3, // 1/3 of max validators
		ConsensusRules: ConsensusRules{
			QuorumSize:        (config.MaxValidators*2 + 2) / 3, // 2/3 quorum (round up)
			FinalityThreshold: (config.MaxValidators*2 + 2) / 3,
			BlockTime:         time.Second * 10,
			MaxValidators:     config.MaxValidators,
		},
		CreatedAt: time.Now(),
		config:    config,
		metrics: NetworkMetrics{
			TotalStake: big.NewInt(0),
		},
	}
}

// AddValidator adds a validator to the network
func (network *ValidatorNetwork) AddValidator(validator *MultiChainValidator) error {
	network.mu.Lock()
	defer network.mu.Unlock()

	if len(network.Validators) >= int(network.config.MaxValidators) {
		return fmt.Errorf("network is at maximum capacity")
	}

	// Ensure validator has a valid stake
	if validator.Stake == nil {
		validator.Stake = big.NewInt(0)
	}

	if validator.Stake.Cmp(network.config.MinStake) < 0 {
		return fmt.Errorf("validator stake %s is below minimum %s", validator.Stake.String(), network.config.MinStake.String())
	}

	network.Validators[validator.ID] = validator
	network.TotalStake.Add(network.TotalStake, validator.Stake)
	network.ActiveValidators++
	network.metrics.TotalValidators++
	network.metrics.TotalStake.Add(network.metrics.TotalStake, validator.Stake)
	network.metrics.LastUpdate = time.Now()

	return nil
}

// RemoveValidator removes a validator from the network
func (network *ValidatorNetwork) RemoveValidator(validatorID string) error {
	network.mu.Lock()
	defer network.mu.Unlock()

	validator, exists := network.Validators[validatorID]
	if !exists {
		return fmt.Errorf("validator %s not found", validatorID)
	}

	network.TotalStake.Sub(network.TotalStake, validator.Stake)
	network.ActiveValidators--
	network.metrics.TotalValidators--
	network.metrics.TotalStake.Sub(network.metrics.TotalStake, validator.Stake)
	network.metrics.LastUpdate = time.Now()

	delete(network.Validators, validatorID)

	return nil
}

// JoinChain allows a validator to join a specific chain
func (validator *MultiChainValidator) JoinChain(chainID string, stake *big.Int) error {
	validator.mu.Lock()
	defer validator.mu.Unlock()

	if validator.Status != ValidatorStatusActive {
		return fmt.Errorf("validator is not active, status: %d", validator.Status)
	}

	if stake.Cmp(validator.config.MinStake) < 0 {
		return fmt.Errorf("stake %s is below minimum %s", stake.String(), validator.config.MinStake.String())
	}

	// Check if adding this stake would exceed maximum
	if new(big.Int).Add(validator.TotalStake, stake).Cmp(validator.config.MaxStake) > 0 {
		return fmt.Errorf("total stake would exceed maximum %s", validator.config.MaxStake.String())
	}

	participation := &ChainParticipation{
		ChainID:     chainID,
		Stake:       stake,
		VotingPower: uint64(stake.Uint64() / 1000000), // 1M units = 1 voting power
		LastVote:    time.Now(),
		VoteHistory: []VoteRecord{},
		Status:      ParticipationStatusActive,
	}

	validator.Chains[chainID] = participation
	validator.TotalStake.Add(validator.TotalStake, stake)
	validator.metrics.TotalStake.Add(validator.metrics.TotalStake, stake)
	validator.metrics.LastUpdate = time.Now()

	return nil
}

// LeaveChain allows a validator to leave a specific chain
func (validator *MultiChainValidator) LeaveChain(chainID string) error {
	validator.mu.Lock()
	defer validator.mu.Unlock()

	participation, exists := validator.Chains[chainID]
	if !exists {
		return fmt.Errorf("not participating in chain %s", chainID)
	}

	if participation.Status == ParticipationStatusUnbonding {
		return fmt.Errorf("already unbonding from chain %s", chainID)
	}

	// Start unbonding period
	participation.Status = ParticipationStatusUnbonding

	// Schedule actual removal after unbonding period
	go func() {
		time.Sleep(validator.config.UnbondingPeriod)
		validator.mu.Lock()
		defer validator.mu.Unlock()

		if _, exists := validator.Chains[chainID]; exists {
			// Stake was already reduced, just clean up
			validator.metrics.LastUpdate = time.Now()
			delete(validator.Chains, chainID)
		}
	}()

	// Immediately reduce stake for unbonding
	validator.TotalStake.Sub(validator.TotalStake, participation.Stake)
	validator.metrics.TotalStake.Sub(validator.metrics.TotalStake, participation.Stake)

	// Ensure stake doesn't go negative
	if validator.TotalStake.Sign() < 0 {
		validator.TotalStake.SetInt64(0)
	}
	if validator.metrics.TotalStake.Sign() < 0 {
		validator.metrics.TotalStake.SetInt64(0)
	}

	return nil
}

// Vote allows a validator to vote on a block
func (validator *MultiChainValidator) Vote(chainID string, blockHash [32]byte, blockHeight uint64, vote bool) error {
	validator.mu.Lock()
	defer validator.mu.Unlock()

	participation, exists := validator.Chains[chainID]
	if !exists {
		return fmt.Errorf("not participating in chain %s", chainID)
	}

	if participation.Status != ParticipationStatusActive {
		return fmt.Errorf("participation status is not active: %d", participation.Status)
	}

	voteRecord := VoteRecord{
		BlockHash:   blockHash,
		BlockHeight: blockHeight,
		Vote:        vote,
		Timestamp:   time.Now(),
		ChainID:     chainID,
	}

	participation.VoteHistory = append(participation.VoteHistory, voteRecord)
	participation.LastVote = time.Now()

	// Update metrics
	validator.metrics.TotalVotes++
	if vote {
		validator.metrics.SuccessfulVotes++
	} else {
		validator.metrics.FailedVotes++
	}
	validator.metrics.LastUpdate = time.Now()

	// Update reputation based on voting consistency
	validator.updateReputation(vote)

	return nil
}

// SlashValidator slashes a validator for misbehavior
func (network *ValidatorNetwork) SlashValidator(validatorID string, reason string, amount *big.Int, evidence []byte) error {
	network.mu.Lock()
	defer network.mu.Unlock()

	validator, exists := network.Validators[validatorID]
	if !exists {
		return fmt.Errorf("validator %s not found", validatorID)
	}

	// Create slashing record
	slashingRecord := SlashingRecord{
		Reason:    reason,
		Amount:    amount,
		Timestamp: time.Now(),
		Evidence:  evidence,
	}

	// Apply slashing to all chain participations
	for _, participation := range validator.Chains {
		participation.SlashingHistory = append(participation.SlashingHistory, slashingRecord)
		participation.Status = ParticipationStatusSlashed

		// Reduce stake proportionally across chains
		slashAmount := new(big.Int).Div(amount, big.NewInt(int64(len(validator.Chains))))
		participation.Stake.Sub(participation.Stake, slashAmount)
	}

	// Reduce validator's total stake by the full slash amount
	validator.TotalStake.Sub(validator.TotalStake, amount)

	// Update validator status
	validator.Status = ValidatorStatusSlashed
	validator.Reputation = validator.Reputation / 2 // Halve reputation

	// Update network metrics
	network.metrics.TotalStake.Sub(network.metrics.TotalStake, amount)
	network.metrics.LastUpdate = time.Now()

	// Update network total stake (reduce by the slashed amount)
	network.TotalStake.Sub(network.TotalStake, amount)

	return nil
}

// RotateValidators performs validator rotation based on reputation and stake
func (network *ValidatorNetwork) RotateValidators() error {
	network.mu.Lock()
	defer network.mu.Unlock()

	if !network.config.ReputationEnabled {
		return fmt.Errorf("reputation-based rotation is disabled")
	}

	// Sort validators by reputation and stake
	type validatorScore struct {
		ID         string
		Score      float64
		Reputation uint64
		Stake      *big.Int
	}

	var scores []validatorScore
	for id, validator := range network.Validators {
		score := float64(validator.Reputation) * float64(validator.TotalStake.Uint64()) / 1e18
		scores = append(scores, validatorScore{
			ID:         id,
			Score:      score,
			Reputation: validator.Reputation,
			Stake:      validator.TotalStake,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Keep top validators and remove bottom ones
	_ = scores[:network.ConsensusRules.MaxValidators] // Top validators to keep
	removals := scores[network.ConsensusRules.MaxValidators:]

	for _, removal := range removals {
		validator := network.Validators[removal.ID]
		network.TotalStake.Sub(network.TotalStake, validator.TotalStake)
		network.ActiveValidators--
		network.metrics.TotalValidators--
		network.metrics.TotalStake.Sub(network.metrics.TotalStake, validator.TotalStake)
		delete(network.Validators, removal.ID)
	}

	network.metrics.LastRotation = time.Now()
	network.metrics.LastUpdate = time.Now()

	return nil
}

// GetConsensusQuorum returns the current consensus quorum
func (network *ValidatorNetwork) GetConsensusQuorum() ([]*MultiChainValidator, error) {
	network.mu.RLock()
	defer network.mu.RUnlock()

	if network.ActiveValidators < network.MinValidators {
		return nil, fmt.Errorf("insufficient validators: %d < %d", network.ActiveValidators, network.MinValidators)
	}

	var quorum []*MultiChainValidator
	totalStake := big.NewInt(0)

	// Sort validators by stake and reputation
	type validatorRank struct {
		validator *MultiChainValidator
		rankScore float64
	}

	var rankings []validatorRank
	for _, validator := range network.Validators {
		if validator.Status == ValidatorStatusActive {
			rankScore := float64(validator.Reputation) * float64(validator.TotalStake.Uint64()) / 1e18
			rankings = append(rankings, validatorRank{
				validator: validator,
				rankScore: rankScore,
			})
		}
	}

	// Sort by rank score (descending)
	for i := 0; i < len(rankings)-1; i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[i].rankScore < rankings[j].rankScore {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			}
		}
	}

	// Select top validators until quorum is reached
	for _, ranking := range rankings {
		if len(quorum) >= int(network.ConsensusRules.QuorumSize) {
			break
		}
		quorum = append(quorum, ranking.validator)
		totalStake.Add(totalStake, ranking.validator.TotalStake)
	}

	return quorum, nil
}

// updateReputation updates validator reputation based on voting behavior
func (validator *MultiChainValidator) updateReputation(vote bool) {
	// Simple reputation update: increase for consistent voting, decrease for inconsistent
	if vote {
		validator.Reputation++
	} else {
		if validator.Reputation > 0 {
			validator.Reputation--
		}
	}

	// Apply reputation decay over time
	timeSinceLastUpdate := time.Since(validator.metrics.LastUpdate)
	if timeSinceLastUpdate > validator.config.ReputationDecay {
		decayFactor := uint64(timeSinceLastUpdate.Hours() / 24) // Daily decay
		if validator.Reputation > decayFactor {
			validator.Reputation -= decayFactor
		} else {
			validator.Reputation = 0
		}
	}
}

// GetStatus returns the current status of the validator
func (validator *MultiChainValidator) GetStatus() ValidatorStatus {
	validator.mu.RLock()
	defer validator.mu.RUnlock()
	return validator.Status
}

// GetMetrics returns the validator metrics
func (validator *MultiChainValidator) GetMetrics() ValidatorMetrics {
	validator.mu.RLock()
	defer validator.mu.RUnlock()
	return validator.metrics
}

// GetNetworkMetrics returns the network metrics
func (network *ValidatorNetwork) GetNetworkMetrics() NetworkMetrics {
	network.mu.RLock()
	defer network.mu.RUnlock()
	return network.metrics
}

// generateValidatorID generates a unique validator ID
func generateValidatorID() string {
	random := make([]byte, 16)
	rand.Read(random)
	hash := sha256.Sum256(random)
	return fmt.Sprintf("validator_%x", hash[:8])
}

// generateNetworkID generates a unique network ID
func generateNetworkID() string {
	random := make([]byte, 16)
	rand.Read(random)
	hash := sha256.Sum256(random)
	return fmt.Sprintf("network_%x", hash[:8])
}
