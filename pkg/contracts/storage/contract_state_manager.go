package storage

import (
	"crypto/sha256"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// ContractStateManager manages contract state within the existing storage layer
type ContractStateManager struct {
	mu sync.RWMutex

	// Core storage integration (will be integrated with existing storage layer)
	storageManager interface{} // Placeholder for storage.StateManager
	trieManager    interface{} // Placeholder for storage.TrieManager

	// Contract state tracking
	contractStates map[engine.Address]*ContractState
	stateHistory   map[engine.Address][]StateSnapshot

	// Transaction management
	pendingChanges map[string]*PendingChange
	transactions   map[string]*Transaction

	// Configuration
	config ContractStateConfig

	// Statistics
	TotalContracts uint64
	TotalStates    uint64
	LastUpdate     time.Time
}

// ContractState represents the current state of a contract
type ContractState struct {
	Address     engine.Address
	Code        []byte
	CodeHash    engine.Hash
	Balance     *big.Int
	Nonce       uint64
	StorageRoot engine.Hash
	CreatedAt   time.Time
	Creator     engine.Address
	UpdatedAt   time.Time

	// Contract-specific storage
	Storage map[engine.Hash][]byte

	// Metadata
	Type     string
	Version  uint64
	IsActive bool
}

// StateSnapshot represents a historical state snapshot
type StateSnapshot struct {
	BlockNumber uint64
	StateHash   engine.Hash
	Timestamp   time.Time
	Changes     []StateChange
}

// StateChange represents a change to contract state
type StateChange struct {
	Key       engine.Hash
	OldValue  []byte
	NewValue  []byte
	Type      StateChangeType
	Timestamp time.Time
}

// StateChangeType indicates the type of state change
type StateChangeType int

const (
	StateChangeStorage StateChangeType = iota
	StateChangeBalance
	StateChangeCode
	StateChangeNonce
	StateChangeMetadata
)

// PendingChange represents a pending state change
type PendingChange struct {
	ID        string
	Contract  engine.Address
	Changes   []StateChange
	Timestamp time.Time
	Status    ChangeStatus
}

// ChangeStatus indicates the status of a pending change
type ChangeStatus int

const (
	ChangeStatusPending ChangeStatus = iota
	ChangeStatusCommitted
	ChangeStatusRolledBack
)

// Transaction represents a contract execution transaction
type Transaction struct {
	ID          string
	Contract    engine.Address
	Method      string
	Args        []interface{}
	GasLimit    uint64
	GasPrice    *big.Int
	Sender      engine.Address
	Value       *big.Int
	Status      TransactionStatus
	Result      *engine.ExecutionResult
	Timestamp   time.Time
	BlockNumber uint64
}

// TransactionStatus indicates the status of a transaction
type TransactionStatus int

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusExecuting
	TransactionStatusCommitted
	TransactionStatusFailed
	TransactionStatusRolledBack
)

// ContractStateConfig holds configuration for contract state management
type ContractStateConfig struct {
	MaxHistorySize     int
	EnableStatePruning bool
	PruningInterval    time.Duration
	MaxStorageSize     uint64
	EnableCompression  bool
	SnapshotInterval   time.Duration
}

// NewContractStateManager creates a new contract state manager
func NewContractStateManager(
	storageManager interface{}, // Placeholder for storage.StateManager
	trieManager interface{},    // Placeholder for storage.TrieManager
	config ContractStateConfig,
) *ContractStateManager {
	return &ContractStateManager{
		storageManager: storageManager,
		trieManager:    trieManager,
		contractStates: make(map[engine.Address]*ContractState),
		stateHistory:   make(map[engine.Address][]StateSnapshot),
		pendingChanges: make(map[string]*PendingChange),
		transactions:   make(map[string]*Transaction),
		config:         config,
		TotalContracts: 0,
		TotalStates:    0,
		LastUpdate:     time.Now(),
	}
}

// CreateContract creates a new contract state
func (csm *ContractStateManager) CreateContract(
	address engine.Address,
	code []byte,
	creator engine.Address,
	contractType string,
) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Check if contract already exists
	if _, exists := csm.contractStates[address]; exists {
		return ErrContractAlreadyExists
	}

	// Calculate code hash
	codeHash := sha256.Sum256(code)

	// Create contract state
	contractState := &ContractState{
		Address:     address,
		Code:        make([]byte, len(code)),
		CodeHash:    engine.Hash(codeHash),
		Balance:     big.NewInt(0),
		Nonce:       0,
		StorageRoot: engine.Hash{},
		CreatedAt:   time.Now(),
		Creator:     creator,
		UpdatedAt:   time.Now(),
		Storage:     make(map[engine.Hash][]byte),
		Type:        contractType,
		Version:     1,
		IsActive:    true,
	}

	// Copy code to avoid external modifications
	copy(contractState.Code, code)

	// Store contract state
	csm.contractStates[address] = contractState

	// Create initial snapshot
	initialSnapshot := StateSnapshot{
		BlockNumber: 0,
		StateHash:   csm.calculateStateHash(contractState),
		Timestamp:   time.Now(),
		Changes:     []StateChange{},
	}

	csm.stateHistory[address] = []StateSnapshot{initialSnapshot}

	csm.TotalContracts++
	csm.TotalStates++
	csm.LastUpdate = time.Now()

	return nil
}

// GetContractState returns the current state of a contract
func (csm *ContractStateManager) GetContractState(address engine.Address) *ContractState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	if state, exists := csm.contractStates[address]; exists {
		// Return a copy to avoid race conditions
		stateCopy := &ContractState{
			Address:     state.Address,
			Code:        make([]byte, len(state.Code)),
			CodeHash:    state.CodeHash,
			Balance:     new(big.Int).Set(state.Balance),
			Nonce:       state.Nonce,
			StorageRoot: state.StorageRoot,
			CreatedAt:   state.CreatedAt,
			Creator:     state.Creator,
			UpdatedAt:   state.UpdatedAt,
			Storage:     make(map[engine.Hash][]byte),
			Type:        state.Type,
			Version:     state.Version,
			IsActive:    state.IsActive,
		}

		// Copy code
		copy(stateCopy.Code, state.Code)

		// Copy storage
		for key, value := range state.Storage {
			stateCopy.Storage[key] = make([]byte, len(value))
			copy(stateCopy.Storage[key], value)
		}

		return stateCopy
	}

	return nil
}

// UpdateContractState updates contract state atomically
func (csm *ContractStateManager) UpdateContractState(
	address engine.Address,
	changes []StateChange,
	blockNumber uint64,
) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Get current contract state
	contractState, exists := csm.contractStates[address]
	if !exists {
		return ErrContractNotFound
	}

	// Create state backup for rollback capability
	stateBackup := csm.backupContractState(contractState)

	// Apply changes
	for _, change := range changes {
		if err := csm.applyStateChange(contractState, change); err != nil {
			// Rollback on error
			csm.rollbackContractState(contractState, stateBackup)
			return err
		}
	}

	// Update metadata
	contractState.UpdatedAt = time.Now()
	contractState.Version++

	// Create new snapshot
	newSnapshot := StateSnapshot{
		BlockNumber: blockNumber,
		StateHash:   csm.calculateStateHash(contractState),
		Timestamp:   time.Now(),
		Changes:     changes,
	}

	csm.stateHistory[address] = append(csm.stateHistory[address], newSnapshot)

	// Prune old snapshots if needed
	csm.pruneOldSnapshots(address)

	csm.TotalStates++
	csm.LastUpdate = time.Now()

	return nil
}

// GetStorageValue retrieves a value from contract storage
func (csm *ContractStateManager) GetStorageValue(
	contractAddress engine.Address,
	key engine.Hash,
) ([]byte, error) {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	contractState, exists := csm.contractStates[contractAddress]
	if !exists {
		return nil, ErrContractNotFound
	}

	if value, exists := contractState.Storage[key]; exists {
		// Return a copy to avoid race conditions
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		return valueCopy, nil
	}

	return nil, nil
}

// SetStorageValue sets a value in contract storage
func (csm *ContractStateManager) SetStorageValue(
	contractAddress engine.Address,
	key engine.Hash,
	value []byte,
) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	contractState, exists := csm.contractStates[contractAddress]
	if !exists {
		return ErrContractNotFound
	}

	// Create change record
	oldValue := contractState.Storage[key]
	change := StateChange{
		Key:       key,
		OldValue:  oldValue,
		NewValue:  make([]byte, len(value)),
		Type:      StateChangeStorage,
		Timestamp: time.Now(),
	}

	// Copy new value
	copy(change.NewValue, value)

	// Update storage
	if value == nil {
		delete(contractState.Storage, key)
	} else {
		contractState.Storage[key] = make([]byte, len(value))
		copy(contractState.Storage[key], value)
	}

	// Update metadata
	contractState.UpdatedAt = time.Now()

	return nil
}

// GetStateHistory returns the state history for a contract
func (csm *ContractStateManager) GetStateHistory(
	address engine.Address,
	limit int,
) []StateSnapshot {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	if history, exists := csm.stateHistory[address]; exists {
		// Return recent snapshots
		if limit > 0 && len(history) > limit {
			history = history[len(history)-limit:]
		}

		// Return copies to avoid race conditions
		snapshots := make([]StateSnapshot, len(history))
		for i, snapshot := range history {
			snapshots[i] = StateSnapshot{
				BlockNumber: snapshot.BlockNumber,
				StateHash:   snapshot.StateHash,
				Timestamp:   snapshot.Timestamp,
				Changes:     make([]StateChange, len(snapshot.Changes)),
			}

			for j, change := range snapshot.Changes {
				snapshots[i].Changes[j] = StateChange{
					Key:       change.Key,
					OldValue:  change.OldValue,
					NewValue:  change.NewValue,
					Type:      change.Type,
					Timestamp: change.Timestamp,
				}
			}
		}

		return snapshots
	}

	return nil
}

// PruneStateHistory removes old state snapshots
func (csm *ContractStateManager) PruneStateHistory(
	address engine.Address,
	keepCount int,
) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if history, exists := csm.stateHistory[address]; exists {
		if len(history) > keepCount {
			// Keep only the most recent snapshots
			csm.stateHistory[address] = history[len(history)-keepCount:]
		}
	}

	return nil
}

// GetStatistics returns contract state management statistics
func (csm *ContractStateManager) GetStatistics() *ContractStateStats {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	return &ContractStateStats{
		TotalContracts: csm.TotalContracts,
		TotalStates:    csm.TotalStates,
		LastUpdate:     csm.LastUpdate,
		Config:         csm.config,
	}
}

// Helper functions
func (csm *ContractStateManager) calculateStateHash(state *ContractState) engine.Hash {
	// In a real implementation, this would calculate a proper state hash
	// For now, return a placeholder
	return engine.Hash{}
}

func (csm *ContractStateManager) backupContractState(state *ContractState) *ContractState {
	// Create a deep copy of the contract state
	backup := &ContractState{
		Address:     state.Address,
		Code:        make([]byte, len(state.Code)),
		CodeHash:    state.CodeHash,
		Balance:     new(big.Int).Set(state.Balance),
		Nonce:       state.Nonce,
		StorageRoot: state.StorageRoot,
		CreatedAt:   state.CreatedAt,
		Creator:     state.Creator,
		UpdatedAt:   state.UpdatedAt,
		Storage:     make(map[engine.Hash][]byte),
		Type:        state.Type,
		Version:     state.Version,
		IsActive:    state.IsActive,
	}

	// Copy code
	copy(backup.Code, state.Code)

	// Copy storage
	for key, value := range state.Storage {
		backup.Storage[key] = make([]byte, len(value))
		copy(backup.Storage[key], value)
	}

	return backup
}

func (csm *ContractStateManager) rollbackContractState(
	current *ContractState,
	backup *ContractState,
) {
	// Restore from backup
	current.Code = make([]byte, len(backup.Code))
	copy(current.Code, backup.Code)
	current.CodeHash = backup.CodeHash
	current.Balance = new(big.Int).Set(backup.Balance)
	current.Nonce = backup.Nonce
	current.StorageRoot = backup.StorageRoot
	current.UpdatedAt = backup.UpdatedAt
	current.Version = backup.Version
	current.IsActive = backup.IsActive

	// Restore storage
	current.Storage = make(map[engine.Hash][]byte)
	for key, value := range backup.Storage {
		current.Storage[key] = make([]byte, len(value))
		copy(current.Storage[key], value)
	}
}

func (csm *ContractStateManager) applyStateChange(
	state *ContractState,
	change StateChange,
) error {
	switch change.Type {
	case StateChangeStorage:
		// Storage changes are handled separately
		return nil
	case StateChangeBalance:
		if change.NewValue != nil {
			state.Balance = new(big.Int).SetBytes(change.NewValue)
		}
	case StateChangeCode:
		if change.NewValue != nil {
			state.Code = make([]byte, len(change.NewValue))
			copy(state.Code, change.NewValue)
			state.CodeHash = sha256.Sum256(change.NewValue)
		}
	case StateChangeNonce:
		if change.NewValue != nil {
			state.Nonce = new(big.Int).SetBytes(change.NewValue).Uint64()
		}
	case StateChangeMetadata:
		// Handle metadata changes
		return nil
	}

	return nil
}

func (csm *ContractStateManager) pruneOldSnapshots(address engine.Address) {
	if history, exists := csm.stateHistory[address]; exists {
		if len(history) > csm.config.MaxHistorySize {
			// Keep only the most recent snapshots
			csm.stateHistory[address] = history[len(history)-csm.config.MaxHistorySize:]
		}
	}
}

// ContractStateStats contains statistics about contract state management
type ContractStateStats struct {
	TotalContracts uint64
	TotalStates    uint64
	LastUpdate     time.Time
	Config         ContractStateConfig
}
