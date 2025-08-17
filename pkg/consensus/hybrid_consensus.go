package consensus

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/storage"
)

// ConsensusType represents the type of consensus mechanism
type ConsensusType int

const (
	ConsensusTypePoW ConsensusType = iota
	ConsensusTypePoS
	ConsensusTypeHybrid
)

// HybridConsensusConfig holds configuration for hybrid consensus
type HybridConsensusConfig struct {
	// PoW Configuration
	TargetBlockTime        time.Duration // Target time between blocks
	DifficultyAdjustment   uint64        // Blocks between difficulty adjustments
	MaxDifficulty          uint64        // Maximum allowed difficulty
	MinDifficulty          uint64        // Minimum allowed difficulty
	
	// PoS Configuration
	StakeRequirement       uint64        // Minimum stake required for validation
	ValidatorReward        uint64        // Reward for validators
	SlashingPenalty        uint64        // Penalty for malicious behavior
	EpochLength            uint64        // Length of staking epochs
	
	// Hybrid Configuration
	PoWWeight              float64       // Weight of PoW in hybrid consensus (0.0-1.0)
	PoSWeight              float64       // Weight of PoS in hybrid consensus (0.0-1.0)
	HybridThreshold        float64       // Threshold for hybrid consensus validation
	TransitionHeight       uint64        // Height at which hybrid consensus activates
}

// DefaultHybridConsensusConfig returns sensible defaults
func DefaultHybridConsensusConfig() *HybridConsensusConfig {
	return &HybridConsensusConfig{
		TargetBlockTime:      10 * time.Second,
		DifficultyAdjustment: 2016,
		MaxDifficulty:        256,
		MinDifficulty:        1,
		StakeRequirement:     1000, // 1000 tokens minimum stake
		ValidatorReward:      50,   // 50 tokens per block
		SlashingPenalty:      100,  // 100 tokens penalty
		EpochLength:          10080, // ~1 week (10080 blocks)
		PoWWeight:            0.6,  // 60% PoW
		PoSWeight:            0.4,  // 40% PoS
		HybridThreshold:      0.7,  // 70% consensus required
		TransitionHeight:     100000, // Activate at block 100,000
	}
}

// Validator represents a PoS validator
type Validator struct {
	Address     []byte
	Stake       uint64
	PublicKey   []byte
	IsActive    bool
	LastStake   time.Time
	Rewards     uint64
	Penalties   uint64
	Votes       uint64
}

// HybridConsensus implements hybrid PoW/PoS consensus
type HybridConsensus struct {
	config           *HybridConsensusConfig
	consensusType    ConsensusType
	currentHeight    uint64
	difficulty       uint64
	validators       map[string]*Validator
	stakePool        uint64
	epochStart       uint64
	lastAdjustment   time.Time
	blockTimes       []time.Duration
	mu               sync.RWMutex
	chain            ChainReader
	storage          storage.StorageInterface
}

// NewHybridConsensus creates a new hybrid consensus instance
func NewHybridConsensus(config *HybridConsensusConfig, chain ChainReader, storage storage.StorageInterface) *HybridConsensus {
	if config == nil {
		config = DefaultHybridConsensusConfig()
	}

	consensus := &HybridConsensus{
		config:        config,
		consensusType: ConsensusTypePoW, // Start with PoW
		difficulty:    config.MinDifficulty,
		validators:    make(map[string]*Validator),
		stakePool:     0,
		epochStart:    0,
		lastAdjustment: time.Now(),
		blockTimes:    make([]time.Duration, 0),
		chain:         chain,
		storage:       storage,
	}

	return consensus
}

// GetConsensusType returns the current consensus type
func (hc *HybridConsensus) GetConsensusType() ConsensusType {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.consensusType
}

// UpdateConsensusType updates the consensus type based on current height
func (hc *HybridConsensus) UpdateConsensusType(height uint64) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.currentHeight = height

	if height >= hc.config.TransitionHeight {
		hc.consensusType = ConsensusTypeHybrid
	} else if height >= hc.config.TransitionHeight/2 {
		hc.consensusType = ConsensusTypePoS
	} else {
		hc.consensusType = ConsensusTypePoW
	}
}

// ValidateBlock validates a block according to the current consensus type
func (hc *HybridConsensus) ValidateBlock(block *block.Block, prevBlock *block.Block) error {
	hc.mu.RLock()
	consensusType := hc.consensusType
	hc.mu.RUnlock()

	switch consensusType {
	case ConsensusTypePoW:
		return hc.validatePoWBlock(block, prevBlock)
	case ConsensusTypePoS:
		return hc.validatePoSBlock(block, prevBlock)
	case ConsensusTypeHybrid:
		return hc.validateHybridBlock(block, prevBlock)
	default:
		return fmt.Errorf("unknown consensus type: %d", consensusType)
	}
}

// validatePoWBlock validates a PoW block
func (hc *HybridConsensus) validatePoWBlock(block *block.Block, prevBlock *block.Block) error {
	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Check proof of work
	if !hc.validateProofOfWork(block) {
		return fmt.Errorf("invalid proof of work")
	}

	// Check difficulty
	expectedDifficulty, err := hc.calculateExpectedDifficulty(block.Header.Height)
	if err != nil {
		return fmt.Errorf("failed to calculate expected difficulty: %w", err)
	}

	if block.Header.Difficulty != expectedDifficulty {
		return fmt.Errorf("block difficulty %d does not match expected %d",
			block.Header.Difficulty, expectedDifficulty)
	}

	return nil
}

// validatePoSBlock validates a PoS block
func (hc *HybridConsensus) validatePoSBlock(block *block.Block, prevBlock *block.Block) error {
	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Check validator signature
	if err := hc.validateValidatorSignature(block); err != nil {
		return fmt.Errorf("invalid validator signature: %w", err)
	}

	// Check stake requirements
	if err := hc.validateStakeRequirements(block); err != nil {
		return fmt.Errorf("stake requirements not met: %w", err)
	}

	return nil
}

// validateHybridBlock validates a hybrid PoW/PoS block
func (hc *HybridConsensus) validateHybridBlock(block *block.Block, prevBlock *block.Block) error {
	// Basic block validation
	if err := block.IsValid(); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Validate both PoW and PoS components
	powValid := hc.validateProofOfWork(block)
	posValid := hc.validateValidatorSignature(block) == nil

	// Calculate consensus score
	powScore := 0.0
	posScore := 0.0

	if powValid {
		powScore = hc.config.PoWWeight
	}
	if posValid {
		posScore = hc.config.PoSWeight
	}

	totalScore := powScore + posScore

	// Check if consensus threshold is met
	if totalScore < hc.config.HybridThreshold {
		return fmt.Errorf("hybrid consensus threshold not met: PoW=%.2f, PoS=%.2f, Total=%.2f",
			powScore, posScore, totalScore)
	}

	return nil
}

// validateProofOfWork validates the proof of work for a block
func (hc *HybridConsensus) validateProofOfWork(block *block.Block) bool {
	hash := block.CalculateHash()
	target := hc.calculateTarget(hc.difficulty)
	return hc.hashLessThan(hash, target)
}

// validateValidatorSignature validates the validator signature for a PoS block
func (hc *HybridConsensus) validateValidatorSignature(block *block.Block) error {
	// This is a simplified implementation
	// In a real system, you would verify cryptographic signatures
	// For now, we'll check if the block has a non-zero nonce as a proxy for PoS validation
	if block.Header.Nonce == 0 {
		return fmt.Errorf("PoS block must have validator signature (non-zero nonce)")
	}

	return nil
}

// validateStakeRequirements validates that the validator meets stake requirements
func (hc *HybridConsensus) validateStakeRequirements(block *block.Block) error {
	// For now, we'll use the block hash as a proxy for validator address
	// In a real implementation, this would be extracted from the block header
	validatorAddr := block.CalculateHash()[:8] // Use first 8 bytes as validator address
	
	validator, exists := hc.validators[string(validatorAddr)]
	if !exists {
		return fmt.Errorf("validator not found: %x", validatorAddr)
	}

	if !validator.IsActive {
		return fmt.Errorf("validator is not active: %x", validatorAddr)
	}

	if validator.Stake < hc.config.StakeRequirement {
		return fmt.Errorf("validator stake %d below requirement %d: %x",
			validator.Stake, hc.config.StakeRequirement, validatorAddr)
	}

	return nil
}

// AddValidator adds a new validator to the consensus
func (hc *HybridConsensus) AddValidator(address []byte, stake uint64, publicKey []byte) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if stake < hc.config.StakeRequirement {
		return fmt.Errorf("stake %d below minimum requirement %d", stake, hc.config.StakeRequirement)
	}

	validator := &Validator{
		Address:   address,
		Stake:     stake,
		PublicKey: publicKey,
		IsActive:  true,
		LastStake: time.Now(),
		Rewards:   0,
		Penalties: 0,
		Votes:     0,
	}

	hc.validators[string(address)] = validator
	hc.stakePool += stake

	return nil
}

// RemoveValidator removes a validator from the consensus
func (hc *HybridConsensus) RemoveValidator(address []byte) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	validator, exists := hc.validators[string(address)]
	if !exists {
		return fmt.Errorf("validator not found: %x", address)
	}

	hc.stakePool -= validator.Stake
	delete(hc.validators, string(address))

	return nil
}

// UpdateStake updates a validator's stake
func (hc *HybridConsensus) UpdateStake(address []byte, newStake uint64) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	validator, exists := hc.validators[string(address)]
	if !exists {
		return fmt.Errorf("validator not found: %x", address)
	}

	oldStake := validator.Stake
	validator.Stake = newStake
	validator.LastStake = time.Now()

	hc.stakePool = hc.stakePool - oldStake + newStake

	return nil
}

// GetValidators returns all active validators
func (hc *HybridConsensus) GetValidators() []*Validator {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	validators := make([]*Validator, 0, len(hc.validators))
	for _, validator := range hc.validators {
		if validator.IsActive {
			validators = append(validators, validator)
		}
	}

	return validators
}

// GetStakePool returns the total stake in the pool
func (hc *HybridConsensus) GetStakePool() uint64 {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.stakePool
}

// SelectValidator selects a validator for the next block based on stake
func (hc *HybridConsensus) SelectValidator() (*Validator, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if hc.stakePool == 0 {
		return nil, fmt.Errorf("no stake in pool")
	}

	// Simple stake-weighted random selection
	// In a real implementation, you might use more sophisticated algorithms
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomValue := binary.BigEndian.Uint64(randomBytes)
	selectedStake := randomValue % hc.stakePool

	currentStake := uint64(0)
	for _, validator := range hc.validators {
		if !validator.IsActive {
			continue
		}
		currentStake += validator.Stake
		if currentStake > selectedStake {
			return validator, nil
		}
	}

	// Fallback to first active validator
	for _, validator := range hc.validators {
		if validator.IsActive {
			return validator, nil
		}
	}

	return nil, fmt.Errorf("no active validators found")
}

// RewardValidator rewards a validator for producing a valid block
func (hc *HybridConsensus) RewardValidator(address []byte) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	validator, exists := hc.validators[string(address)]
	if !exists {
		return fmt.Errorf("validator not found: %x", address)
	}

	validator.Rewards += hc.config.ValidatorReward
	validator.Votes++

	return nil
}

// SlashValidator penalizes a validator for malicious behavior
func (hc *HybridConsensus) SlashValidator(address []byte) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	validator, exists := hc.validators[string(address)]
	if !exists {
		return fmt.Errorf("validator not found: %x", address)
	}

	penalty := hc.config.SlashingPenalty
	if penalty > validator.Stake {
		penalty = validator.Stake
	}

	validator.Penalties += penalty
	validator.Stake -= penalty
	hc.stakePool -= penalty

	if validator.Stake < hc.config.StakeRequirement {
		validator.IsActive = false
	}

	return nil
}

// calculateExpectedDifficulty calculates the expected difficulty for a block
func (hc *HybridConsensus) calculateExpectedDifficulty(height uint64) (uint64, error) {
	if height == 0 {
		return hc.config.MinDifficulty, nil
	}

	if height%hc.config.DifficultyAdjustment != 0 {
		// If not an adjustment block, difficulty is the same as the previous block
		prevBlock := hc.chain.GetBlockByHeight(height - 1)
		if prevBlock == nil {
			return 0, fmt.Errorf("previous block not found for height %d", height)
		}
		return prevBlock.Header.Difficulty, nil
	}

	// It's an adjustment block, calculate new difficulty
	currentBlock := hc.chain.GetBlockByHeight(height - 1)
	if currentBlock == nil {
		return 0, fmt.Errorf("current block not found for height %d", height-1)
	}

	oldBlockHeight := height - hc.config.DifficultyAdjustment
	oldBlock := hc.chain.GetBlockByHeight(oldBlockHeight)
	if oldBlock == nil {
		return 0, fmt.Errorf("old block not found for height %d", oldBlockHeight)
	}

	actualTime := currentBlock.Header.Timestamp.Sub(oldBlock.Header.Timestamp)
	expectedTime := time.Duration(hc.config.DifficultyAdjustment) * hc.config.TargetBlockTime

	adjustmentFactor := float64(actualTime) / float64(expectedTime)

	// Limit adjustment factor to prevent extreme swings
	if adjustmentFactor < 0.25 {
		adjustmentFactor = 0.25
	}
	if adjustmentFactor > 4.0 {
		adjustmentFactor = 4.0
	}

	oldDifficulty := oldBlock.Header.Difficulty
	newDifficulty := uint64(float64(oldDifficulty) * adjustmentFactor)

	if newDifficulty < hc.config.MinDifficulty {
		newDifficulty = hc.config.MinDifficulty
	}
	if newDifficulty > hc.config.MaxDifficulty {
		newDifficulty = hc.config.MaxDifficulty
	}

	return newDifficulty, nil
}

// calculateTarget calculates the target hash for a given difficulty
func (hc *HybridConsensus) calculateTarget(difficulty uint64) []byte {
	target := new(big.Int)
	target.SetBit(target, int(256-difficulty), 1)

	targetBytes := target.Bytes()
	if len(targetBytes) > 32 {
		return targetBytes[:32]
	}

	result := make([]byte, 32)
	copy(result[32-len(targetBytes):], targetBytes)

	return result
}

// hashLessThan checks if hash1 is lexicographically less than hash2
func (hc *HybridConsensus) hashLessThan(hash1, hash2 []byte) bool {
	for i := 0; i < len(hash1); i++ {
		if hash1[i] < hash2[i] {
			return true
		}
		if hash1[i] > hash2[i] {
			return false
		}
	}
	return false
}

// GetConsensusStats returns consensus statistics
func (hc *HybridConsensus) GetConsensusStats() map[string]interface{} {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	activeValidators := 0
	totalStake := uint64(0)
	for _, validator := range hc.validators {
		if validator.IsActive {
			activeValidators++
			totalStake += validator.Stake
		}
	}

	return map[string]interface{}{
		"consensus_type":     hc.consensusType,
		"current_height":     hc.currentHeight,
		"difficulty":         hc.difficulty,
		"active_validators":  activeValidators,
		"total_stake":        totalStake,
		"stake_pool":         hc.stakePool,
		"epoch_start":        hc.epochStart,
		"last_adjustment":    hc.lastAdjustment,
		"block_times_count":  len(hc.blockTimes),
	}
}
