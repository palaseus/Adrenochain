package consensus

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/palaseus/adrenochain/pkg/contracts/storage"
)

// StateManager defines the interface for state management operations
type StateManager interface {
	RollbackBlock(blockNumber uint64) error
	CreateContract(address engine.Address, code []byte, creator engine.Address) error
	GetContractState(address engine.Address) *storage.ContractState
	UpdateContractState(address engine.Address, changes []storage.StateChange) error
}

// StateTransitionManager manages atomic contract execution within consensus
type StateTransitionManager struct {
	mu sync.RWMutex

	// Core components
	contractEngine engine.ContractEngine
	stateManager   StateManager

	// Consensus integration
	consensusEngine interface{} // Placeholder for consensus.ConsensusEngine

	// Transaction management
	pendingTransactions  map[string]*ConsensusTransaction
	executedTransactions map[string]*ConsensusTransaction
	blockTransactions    map[uint64][]string

	// State validation
	stateValidators map[string]StateValidator
	validationRules []ValidationRule

	// Configuration
	config StateTransitionConfig

	// Statistics
	TotalTransactions      uint64
	SuccessfulTransactions uint64
	FailedTransactions     uint64
	LastUpdate             time.Time
}

// ConsensusTransaction represents a transaction within consensus
type ConsensusTransaction struct {
	ID             string
	BlockNumber    uint64
	BlockHash      engine.Hash
	Contract       engine.Address
	Method         string
	Args           []interface{}
	GasLimit       uint64
	GasPrice       *big.Int
	Sender         engine.Address
	Value          *big.Int
	Nonce          uint64
	Status         TransactionStatus
	Result         *engine.ExecutionResult
	StateChanges   []storage.StateChange
	Timestamp      time.Time
	ConsensusRound uint64
}

// TransactionStatus indicates the status of a consensus transaction
type TransactionStatus int

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusValidating
	TransactionStatusExecuting
	TransactionStatusCommitted
	TransactionStatusFailed
	TransactionStatusRolledBack
)

// StateValidator validates contract state transitions
type StateValidator interface {
	ValidateStateTransition(
		contract engine.Address,
		changes []storage.StateChange,
		context ValidationContext,
	) error
}

// ValidationContext provides context for state validation
type ValidationContext struct {
	BlockNumber    uint64
	BlockHash      engine.Hash
	ConsensusRound uint64
	Timestamp      time.Time
	GasUsed        uint64
	Sender         engine.Address
}

// ValidationRule defines a rule for state validation
type ValidationRule struct {
	ID          string
	Name        string
	Description string
	Priority    int
	Enabled     bool
	Validator   StateValidator
}

// StateTransitionConfig holds configuration for state transitions
type StateTransitionConfig struct {
	MaxTransactionsPerBlock uint64
	EnableStateValidation   bool
	EnableRollback          bool
	MaxRollbackDepth        int
	ValidationTimeout       time.Duration
	ExecutionTimeout        time.Duration
}

// NewStateTransitionManager creates a new state transition manager
func NewStateTransitionManager(
	contractEngine engine.ContractEngine,
	stateManager StateManager,
	config StateTransitionConfig,
) *StateTransitionManager {
	return &StateTransitionManager{
		contractEngine:         contractEngine,
		stateManager:           stateManager,
		consensusEngine:        nil, // Will be initialized separately
		pendingTransactions:    make(map[string]*ConsensusTransaction),
		executedTransactions:   make(map[string]*ConsensusTransaction),
		blockTransactions:      make(map[uint64][]string),
		stateValidators:        make(map[string]StateValidator),
		validationRules:        make([]ValidationRule, 0),
		config:                 config,
		TotalTransactions:      0,
		SuccessfulTransactions: 0,
		FailedTransactions:     0,
		LastUpdate:             time.Now(),
	}
}

// InitializeConsensusEngine initializes the consensus engine
func (stm *StateTransitionManager) InitializeConsensusEngine(consensusEngine interface{}) {
	stm.mu.Lock()
	defer stm.mu.Unlock()

	stm.consensusEngine = consensusEngine
}

// AddStateValidator adds a new state validator
func (stm *StateTransitionManager) AddStateValidator(
	id string,
	validator StateValidator,
	priority int,
) error {
	stm.mu.Lock()
	defer stm.mu.Unlock()

	if _, exists := stm.stateValidators[id]; exists {
		return ErrValidatorAlreadyExists
	}

	stm.stateValidators[id] = validator

	rule := ValidationRule{
		ID:          id,
		Name:        id,
		Description: "State validator for " + id,
		Priority:    priority,
		Enabled:     true,
		Validator:   validator,
	}

	stm.validationRules = append(stm.validationRules, rule)

	// Sort by priority
	stm.sortValidationRules()

	return nil
}

// ExecuteTransaction executes a transaction within consensus
func (stm *StateTransitionManager) ExecuteTransaction(
	tx *ConsensusTransaction,
	consensusRound uint64,
) error {
	stm.mu.Lock()
	defer stm.mu.Unlock()

	// Validate transaction
	if err := stm.validateTransaction(tx); err != nil {
		tx.Status = TransactionStatusFailed
		stm.FailedTransactions++
		return err
	}

	// Set consensus round
	tx.ConsensusRound = consensusRound
	tx.Status = TransactionStatusValidating

	// Validate state transition
	if stm.config.EnableStateValidation {
		if err := stm.validateStateTransition(tx); err != nil {
			tx.Status = TransactionStatusFailed
			stm.FailedTransactions++
			return err
		}
	}

	// Execute transaction
	tx.Status = TransactionStatusExecuting
	if err := stm.executeTransaction(tx); err != nil {
		tx.Status = TransactionStatusFailed
		stm.FailedTransactions++
		return err
	}

	// Commit transaction
	tx.Status = TransactionStatusCommitted
	stm.SuccessfulTransactions++

	// Record transaction
	stm.executedTransactions[tx.ID] = tx
	if stm.blockTransactions[tx.BlockNumber] == nil {
		stm.blockTransactions[tx.BlockNumber] = make([]string, 0)
	}
	stm.blockTransactions[tx.BlockNumber] = append(stm.blockTransactions[tx.BlockNumber], tx.ID)

	stm.TotalTransactions++
	stm.LastUpdate = time.Now()

	return nil
}

// RollbackBlock rolls back all transactions in a block
func (stm *StateTransitionManager) RollbackBlock(blockNumber uint64) error {
	stm.mu.Lock()
	defer stm.mu.Unlock()

	if !stm.config.EnableRollback {
		return ErrRollbackNotEnabled
	}

	transactionIDs, exists := stm.blockTransactions[blockNumber]
	if !exists {
		return nil // No transactions to rollback
	}

	// Rollback transactions in reverse order
	for i := len(transactionIDs) - 1; i >= 0; i-- {
		txID := transactionIDs[i]
		if tx, exists := stm.executedTransactions[txID]; exists {
			if err := stm.rollbackTransaction(tx); err != nil {
				return err
			}
			tx.Status = TransactionStatusRolledBack
		}
	}

	// Remove block transactions
	delete(stm.blockTransactions, blockNumber)

	return nil
}

// GetTransaction returns a transaction by ID
func (stm *StateTransitionManager) GetTransaction(txID string) *ConsensusTransaction {
	stm.mu.RLock()
	defer stm.mu.RUnlock()

	if tx, exists := stm.executedTransactions[txID]; exists {
		// Return a copy to avoid race conditions
		txCopy := &ConsensusTransaction{
			ID:             tx.ID,
			BlockNumber:    tx.BlockNumber,
			BlockHash:      tx.BlockHash,
			Contract:       tx.Contract,
			Method:         tx.Method,
			Args:           make([]interface{}, len(tx.Args)),
			GasLimit:       tx.GasLimit,
			GasPrice:       tx.GasPrice,
			Sender:         tx.Sender,
			Value:          tx.Value,
			Nonce:          tx.Nonce,
			Status:         tx.Status,
			Result:         tx.Result,
			StateChanges:   make([]storage.StateChange, len(tx.StateChanges)),
			Timestamp:      tx.Timestamp,
			ConsensusRound: tx.ConsensusRound,
		}

		// Copy args
		copy(txCopy.Args, tx.Args)

		// Copy state changes
		for i, change := range tx.StateChanges {
			txCopy.StateChanges[i] = storage.StateChange{
				Key:       change.Key,
				OldValue:  change.OldValue,
				NewValue:  change.NewValue,
				Type:      change.Type,
				Timestamp: change.Timestamp,
			}
		}

		return txCopy
	}

	return nil
}

// GetBlockTransactions returns all transactions in a block
func (stm *StateTransitionManager) GetBlockTransactions(blockNumber uint64) []*ConsensusTransaction {
	stm.mu.RLock()
	defer stm.mu.RUnlock()

	transactionIDs, exists := stm.blockTransactions[blockNumber]
	if !exists {
		return nil
	}

	transactions := make([]*ConsensusTransaction, 0, len(transactionIDs))
	for _, txID := range transactionIDs {
		if tx, exists := stm.executedTransactions[txID]; exists {
			transactions = append(transactions, tx)
		}
	}

	return transactions
}

// GetStatistics returns state transition statistics
func (stm *StateTransitionManager) GetStatistics() *StateTransitionStats {
	stm.mu.RLock()
	defer stm.mu.RUnlock()

	return &StateTransitionStats{
		TotalTransactions:      stm.TotalTransactions,
		SuccessfulTransactions: stm.SuccessfulTransactions,
		FailedTransactions:     stm.FailedTransactions,
		LastUpdate:             stm.LastUpdate,
		Config:                 stm.config,
	}
}

// Helper functions
func (stm *StateTransitionManager) validateTransaction(tx *ConsensusTransaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	
	if tx.Contract == (engine.Address{}) {
		return ErrInvalidContractAddress
	}

	if tx.Sender == (engine.Address{}) {
		return ErrInvalidSender
	}

	if tx.GasLimit == 0 {
		return ErrInvalidGasLimit
	}

	if tx.GasPrice != nil && tx.GasPrice.Sign() <= 0 {
		return ErrInvalidGasPrice
	}

	return nil
}

func (stm *StateTransitionManager) validateStateTransition(tx *ConsensusTransaction) error {
	// Create validation context
	context := ValidationContext{
		BlockNumber:    tx.BlockNumber,
		BlockHash:      tx.BlockHash,
		ConsensusRound: tx.ConsensusRound,
		Timestamp:      tx.Timestamp,
		GasUsed:        0, // Will be updated after execution
		Sender:         tx.Sender,
	}

	// Apply validation rules in priority order
	for _, rule := range stm.validationRules {
		if !rule.Enabled {
			continue
		}

		if err := rule.Validator.ValidateStateTransition(tx.Contract, tx.StateChanges, context); err != nil {
			return err
		}
	}

	return nil
}

func (stm *StateTransitionManager) executeTransaction(tx *ConsensusTransaction) error {
	// In a real implementation, this would execute the contract
	// For now, create a placeholder result

	// Simulate execution
	time.Sleep(1 * time.Millisecond)

	// Create execution result
	tx.Result = &engine.ExecutionResult{
		Success:      true,
		ReturnData:   []byte("success"),
		GasUsed:      tx.GasLimit / 2, // Simulate gas usage
		GasRemaining: tx.GasLimit / 2,
		Logs:         []engine.Log{},
		Error:        nil,
		StateChanges: []engine.StateChange{},
	}

		// Validate final state
	if stm.config.EnableStateValidation {
		if err := stm.validateStateTransition(tx); err != nil {
			return err
		}
	}

	return nil
}

func (stm *StateTransitionManager) rollbackTransaction(tx *ConsensusTransaction) error {
	// In a real implementation, this would rollback the transaction
	// For now, just mark it as rolled back

	// Update statistics
	if tx.Status == TransactionStatusCommitted {
		stm.SuccessfulTransactions--
	} else if tx.Status == TransactionStatusFailed {
		stm.FailedTransactions--
	}

	return nil
}

func (stm *StateTransitionManager) sortValidationRules() {
	// Simple bubble sort by priority (higher priority first)
	for i := 0; i < len(stm.validationRules)-1; i++ {
		for j := 0; j < len(stm.validationRules)-i-1; j++ {
			if stm.validationRules[j].Priority < stm.validationRules[j+1].Priority {
				stm.validationRules[j], stm.validationRules[j+1] = stm.validationRules[j+1], stm.validationRules[j]
			}
		}
	}
}

// StateTransitionStats contains statistics about state transitions
type StateTransitionStats struct {
	TotalTransactions      uint64
	SuccessfulTransactions uint64
	FailedTransactions     uint64
	LastUpdate             time.Time
	Config                 StateTransitionConfig
}
