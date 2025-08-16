package consensus

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
	"github.com/gochain/gochain/pkg/contracts/storage"
)

// StateManager interface is defined in state_transitions.go

// StateTransitionManagerInterface defines the interface for state transition operations
type StateTransitionManagerInterface interface {
	RollbackBlock(blockNumber uint64) error
	ExecuteTransaction(tx *ConsensusTransaction, consensusRound uint64) error
}

// ConsensusIntegration integrates contract execution with hybrid consensus
type ConsensusIntegration struct {
	mu sync.RWMutex

	// Core components
	contractEngine engine.ContractEngine
	stateManager   StateManager
	stateTransitions StateTransitionManagerInterface
	
	// Consensus integration
	consensusEngine interface{} // Will integrate with pkg/consensus/hybrid_consensus.go
	blockValidator  *BlockValidator
	gasAccounting   *GasAccounting
	
	// Contract execution state
	executionState map[string]*ContractExecutionState
	pendingBlocks  map[uint64]*PendingBlock
	
	// Configuration
	config ConsensusIntegrationConfig
	
	// Statistics
	TotalBlocksProcessed uint64
	TotalContractsExecuted uint64
	TotalGasUsed uint64
	LastBlockTime time.Time
}

// ConsensusIntegrationConfig holds configuration for consensus integration
type ConsensusIntegrationConfig struct {
	// Contract execution settings
	EnableContractExecution bool
	MaxContractsPerBlock   uint64
	MaxGasPerBlock         uint64
	MaxGasPerContract      uint64
	EnableGasAccounting    bool
	
	// State validation settings
	EnableStateValidation  bool
	EnableStateRollback    bool
	MaxRollbackDepth      int
	
	// Consensus settings
	ConsensusTimeout       time.Duration
	EnableBlockValidation  bool
	EnableTransactionOrdering bool
}

// BlockValidator validates blocks with contract execution
type BlockValidator struct {
	mu sync.RWMutex

	// Validation state
	validatedBlocks map[uint64]*ValidatedBlock
	validationRules []ValidationRule
	
	// Configuration
	config BlockValidationConfig
}

// BlockValidationConfig holds configuration for block validation
type BlockValidationConfig struct {
	EnableContractValidation bool
	EnableStateValidation    bool
	EnableGasValidation      bool
	MaxValidationTime        time.Duration
}

// ValidatedBlock represents a validated block
type ValidatedBlock struct {
	BlockNumber    uint64
	BlockHash      engine.Hash
	ValidatedAt    time.Time
	ValidationTime time.Duration
	Status         ValidationStatus
	Issues         []ValidationIssue
}

// ValidationStatus indicates the validation status
type ValidationStatus int

const (
	ValidationStatusPending ValidationStatus = iota
	ValidationStatusValidating
	ValidationStatusValid
	ValidationStatusInvalid
	ValidationStatusFailed
)

// ValidationIssue represents a validation issue
type ValidationIssue struct {
	Type        IssueType
	Severity    IssueSeverity
	Description string
	Location    string
	Timestamp   time.Time
}

// IssueType indicates the type of validation issue
type IssueType int

const (
	IssueTypeContractExecution IssueType = iota
	IssueTypeStateValidation
	IssueTypeGasAccounting
	IssueTypeConsensus
	IssueTypeOther
)

// IssueSeverity indicates the severity of a validation issue
type IssueSeverity int

const (
	IssueSeverityLow IssueSeverity = iota
	IssueSeverityMedium
	IssueSeverityHigh
	IssueSeverityCritical
)

// GasAccounting manages gas accounting within consensus
type GasAccounting struct {
	mu sync.RWMutex

	// Gas tracking
	blockGasUsed    map[uint64]uint64
	contractGasUsed map[string]uint64
	totalGasUsed    uint64
	
	// Gas limits
	maxGasPerBlock  uint64
	maxGasPerContract uint64
	
	// Statistics
	TotalBlocks uint64
	TotalContracts uint64
	AverageGasPerBlock float64
}

// ContractExecutionState tracks contract execution within consensus
type ContractExecutionState struct {
	BlockNumber    uint64
	Contract       engine.Address
	Method         string
	GasUsed        uint64
	Status         ExecutionStatus
	Result         *engine.ExecutionResult
	StateChanges   []storage.StateChange
	Timestamp      time.Time
}

// ExecutionStatus indicates the execution status
type ExecutionStatus int

const (
	ExecutionStatusPending ExecutionStatus = iota
	ExecutionStatusExecuting
	ExecutionStatusCompleted
	ExecutionStatusFailed
	ExecutionStatusRolledBack
)

// PendingBlock represents a block pending contract execution
type PendingBlock struct {
	BlockNumber    uint64
	BlockHash      engine.Hash
	Transactions   []*PendingTransaction
	Status         BlockStatus
	CreatedAt      time.Time
}

// BlockStatus indicates the block status
type BlockStatus int

const (
	BlockStatusPending BlockStatus = iota
	BlockStatusExecuting
	BlockStatusCompleted
	BlockStatusFailed
	BlockStatusRolledBack
)

// PendingTransaction represents a pending transaction
type PendingTransaction struct {
	Hash          engine.Hash
	Contract      engine.Address
	Method        string
	Args          []interface{}
	GasLimit      uint64
	GasPrice      *big.Int
	Value         *big.Int
	Status        TransactionStatus
	Timestamp     time.Time
}

// NewConsensusIntegration creates a new consensus integration
func NewConsensusIntegration(
	contractEngine engine.ContractEngine,
	stateManager StateManager,
	config ConsensusIntegrationConfig,
) *ConsensusIntegration {
	ci := &ConsensusIntegration{
		contractEngine:    contractEngine,
		stateManager:      stateManager,
		stateTransitions:  nil, // Will be set separately to avoid circular dependency
		consensusEngine:   nil, // Will be initialized separately
		blockValidator:    NewBlockValidator(BlockValidationConfig{}),
		gasAccounting:     NewGasAccounting(config.MaxGasPerBlock, config.MaxGasPerContract),
		executionState:    make(map[string]*ContractExecutionState),
		pendingBlocks:     make(map[uint64]*PendingBlock),
		config:            config,
		TotalBlocksProcessed: 0,
		TotalContractsExecuted: 0,
		TotalGasUsed: 0,
		LastBlockTime: time.Time{},
	}
	
	// Initialize state transition manager
	ci.stateTransitions = NewStateTransitionManager(contractEngine, stateManager, StateTransitionConfig{})
	
	return ci
}

// InitializeConsensusEngine initializes the consensus engine
func (ci *ConsensusIntegration) InitializeConsensusEngine(consensusEngine interface{}) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	ci.consensusEngine = consensusEngine
}

// SetStateTransitionManager method removed - no longer needed

// ProcessBlock processes a block with contract execution
func (ci *ConsensusIntegration) ProcessBlock(
	ctx context.Context,
	blockNumber uint64,
	blockHash engine.Hash,
	transactions []*PendingTransaction,
) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	if !ci.config.EnableContractExecution {
		return ErrContractExecutionNotEnabled
	}
	
	// Create pending block
	pendingBlock := &PendingBlock{
		BlockNumber:  blockNumber,
		BlockHash:    blockHash,
		Transactions: transactions,
		Status:       BlockStatusPending,
		CreatedAt:    time.Now(),
	}
	
	ci.pendingBlocks[blockNumber] = pendingBlock
	
	// Process block
	if err := ci.executeBlock(ctx, pendingBlock); err != nil {
		pendingBlock.Status = BlockStatusFailed
		return err
	}
	
	pendingBlock.Status = BlockStatusCompleted
	ci.TotalBlocksProcessed++
	ci.LastBlockTime = time.Now()
	
	return nil
}

// ValidateBlock validates a block with contract execution
func (ci *ConsensusIntegration) ValidateBlock(
	ctx context.Context,
	blockNumber uint64,
	blockHash engine.Hash,
) (*ValidatedBlock, error) {
	if !ci.config.EnableBlockValidation {
		return nil, ErrBlockValidationNotEnabled
	}
	
	return ci.blockValidator.ValidateBlock(ctx, blockNumber, blockHash)
}

// RollbackBlock rolls back contract execution for a block
func (ci *ConsensusIntegration) RollbackBlock(blockNumber uint64) error {
	if !ci.config.EnableStateRollback {
		return ErrStateRollbackNotEnabled
	}
	
	ci.mu.Lock()
	defer ci.mu.Unlock()
	
	pendingBlock, exists := ci.pendingBlocks[blockNumber]
	if !exists {
		return ErrBlockNotFound
	}
	
	// Rollback state transitions
	if err := ci.stateTransitions.RollbackBlock(blockNumber); err != nil {
		return err
	}
	
	// Update block status
	pendingBlock.Status = BlockStatusRolledBack
	
	return nil
}

// GetExecutionState returns contract execution state
func (ci *ConsensusIntegration) GetExecutionState(contract engine.Address) *ContractExecutionState {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	
	key := contract.String()
	if state, exists := ci.executionState[key]; exists {
		// Return a copy to avoid race conditions
		stateCopy := &ContractExecutionState{
			BlockNumber:  state.BlockNumber,
			Contract:     state.Contract,
			Method:       state.Method,
			GasUsed:      state.GasUsed,
			Status:       state.Status,
			Result:       state.Result,
			StateChanges: make([]storage.StateChange, len(state.StateChanges)),
			Timestamp:    state.Timestamp,
		}
		
		copy(stateCopy.StateChanges, state.StateChanges)
		
		return stateCopy
	}
	
	return nil
}

// GetGasAccounting returns gas accounting information
func (ci *ConsensusIntegration) GetGasAccounting() *GasAccountingInfo {
	return ci.gasAccounting.GetInfo()
}

// Helper functions
func (ci *ConsensusIntegration) executeBlock(
	ctx context.Context,
	pendingBlock *PendingBlock,
) error {
	pendingBlock.Status = BlockStatusExecuting
	
	// Execute transactions in order
	for _, tx := range pendingBlock.Transactions {
		if err := ci.executeTransaction(ctx, tx, pendingBlock.BlockNumber); err != nil {
			return err
		}
	}
	
	return nil
}

func (ci *ConsensusIntegration) executeTransaction(
	ctx context.Context,
	tx *PendingTransaction,
	blockNumber uint64,
) error {
	// Check for nil transaction
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	
	// Update transaction status
	tx.Status = TransactionStatusExecuting
	
	// Execute contract call
	// Create a mock contract for execution (in real implementation, this would be retrieved from storage)
	contract := &engine.Contract{
		Address: tx.Contract,
		Code:    []byte{}, // Will be populated from storage
	}
	
	// Convert method and args to input data (simplified)
	input := []byte(tx.Method) // In real implementation, this would be properly encoded
	
	result, err := ci.contractEngine.Execute(
		contract,
		input,
		tx.GasLimit,
		engine.Address{}, // sender (placeholder)
		tx.Value,
	)
	
	if err != nil {
		tx.Status = TransactionStatusFailed
		return err
	}
	
	// Update gas accounting
	ci.gasAccounting.RecordGasUsage(blockNumber, tx.Contract.String(), result.GasUsed)
	
	// Record execution state
	// Convert engine.StateChange to storage.StateChange
	stateChanges := make([]storage.StateChange, len(result.StateChanges))
	for i, change := range result.StateChanges {
		stateChanges[i] = storage.StateChange{
			Key:       change.Key,
			OldValue:  change.Value, // Simplified mapping
			NewValue:  change.Value,
			Type:      storage.StateChangeStorage, // Use the correct type from contract_state_manager
			Timestamp: time.Now(),
		}
	}
	
	executionState := &ContractExecutionState{
		BlockNumber:  blockNumber,
		Contract:     tx.Contract,
		Method:       tx.Method,
		GasUsed:      result.GasUsed,
		Status:       ExecutionStatusCompleted,
		Result:       result,
		StateChanges: stateChanges,
		Timestamp:    time.Now(),
	}
	
	ci.executionState[tx.Contract.String()] = executionState
	ci.TotalContractsExecuted++
	ci.TotalGasUsed += result.GasUsed
	
	tx.Status = TransactionStatusCommitted
	
	return nil
}

// NewBlockValidator creates a new block validator
func NewBlockValidator(config BlockValidationConfig) *BlockValidator {
	return &BlockValidator{
		validatedBlocks: make(map[uint64]*ValidatedBlock),
		validationRules: make([]ValidationRule, 0),
		config:          config,
	}
}

// ValidateBlock validates a block
func (bv *BlockValidator) ValidateBlock(
	ctx context.Context,
	blockNumber uint64,
	blockHash engine.Hash,
) (*ValidatedBlock, error) {
	bv.mu.Lock()
	defer bv.mu.Unlock()
	
	startTime := time.Now()
	
	validatedBlock := &ValidatedBlock{
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		ValidatedAt: time.Now(),
		Status:      ValidationStatusValidating,
		Issues:      make([]ValidationIssue, 0),
	}
	
	// Perform validation
	if err := bv.performValidation(ctx, validatedBlock); err != nil {
		validatedBlock.Status = ValidationStatusFailed
		validatedBlock.ValidationTime = time.Since(startTime)
		return validatedBlock, err
	}
	
	// Determine final status
	if len(validatedBlock.Issues) == 0 {
		validatedBlock.Status = ValidationStatusValid
	} else {
		validatedBlock.Status = ValidationStatusInvalid
	}
	
	validatedBlock.ValidationTime = time.Since(startTime)
	
	// Store result
	bv.validatedBlocks[blockNumber] = validatedBlock
	
	return validatedBlock, nil
}

func (bv *BlockValidator) performValidation(
	ctx context.Context,
	validatedBlock *ValidatedBlock,
) error {
	// In a real implementation, this would perform comprehensive validation
	// For now, simulate validation
	
	// Simulate validation time
	time.Sleep(10 * time.Millisecond)
	
	// Simulate finding some issues (for demonstration)
	if validatedBlock.BlockNumber%10 == 0 {
		validatedBlock.Issues = append(validatedBlock.Issues, ValidationIssue{
			Type:        IssueTypeContractExecution,
			Severity:    IssueSeverityLow,
			Description: "Minor gas optimization opportunity",
			Location:    "Block validation",
			Timestamp:   time.Now(),
		})
	}
	
	return nil
}

// NewGasAccounting creates new gas accounting
func NewGasAccounting(maxGasPerBlock, maxGasPerContract uint64) *GasAccounting {
	return &GasAccounting{
		blockGasUsed:      make(map[uint64]uint64),
		contractGasUsed:   make(map[string]uint64),
		totalGasUsed:      0,
		maxGasPerBlock:    maxGasPerBlock,
		maxGasPerContract: maxGasPerContract,
		TotalBlocks:       0,
		TotalContracts:    0,
		AverageGasPerBlock: 0,
	}
}

// RecordGasUsage records gas usage for a block and contract
func (ga *GasAccounting) RecordGasUsage(blockNumber uint64, contractKey string, gasUsed uint64) {
	ga.mu.Lock()
	defer ga.mu.Unlock()
	
	// Record block gas usage
	ga.blockGasUsed[blockNumber] += gasUsed
	
	// Record contract gas usage
	ga.contractGasUsed[contractKey] += gasUsed
	
	// Update totals
	ga.totalGasUsed += gasUsed
	
	// Update statistics
	if ga.blockGasUsed[blockNumber] == gasUsed {
		ga.TotalBlocks++
	}
	ga.TotalContracts++
	
	// Update average
	ga.AverageGasPerBlock = float64(ga.totalGasUsed) / float64(ga.TotalBlocks)
}

// GetInfo returns gas accounting information
func (ga *GasAccounting) GetInfo() *GasAccountingInfo {
	ga.mu.RLock()
	defer ga.mu.RUnlock()
	
	return &GasAccountingInfo{
		TotalGasUsed:       ga.totalGasUsed,
		TotalBlocks:        ga.TotalBlocks,
		TotalContracts:     ga.TotalContracts,
		AverageGasPerBlock: ga.AverageGasPerBlock,
		MaxGasPerBlock:     ga.maxGasPerBlock,
		MaxGasPerContract:  ga.maxGasPerContract,
	}
}

// GasAccountingInfo contains gas accounting information
type GasAccountingInfo struct {
	TotalGasUsed       uint64
	TotalBlocks        uint64
	TotalContracts     uint64
	AverageGasPerBlock float64
	MaxGasPerBlock     uint64
	MaxGasPerContract  uint64
}
