package storage

import (
	"crypto/sha256"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// StorageIntegration extends existing storage for smart contract data
type StorageIntegration struct {
	mu sync.RWMutex

	// Core components
	contractStateManager *ContractStateManager
	statePruningManager  *StatePruningManager

	// Storage integration
	storageManager interface{} // Will integrate with pkg/storage/
	trieManager    interface{} // Will integrate with pkg/storage/trie.go
	pruningManager interface{} // Will integrate with pkg/storage/pruning.go

	// Contract storage state
	contractStorage map[engine.Address]*ContractStorageData
	storageHistory  map[engine.Address][]StorageSnapshot

	// Configuration
	config StorageIntegrationConfig

	// Statistics
	TotalContracts   uint64
	TotalStorageSize uint64
	TotalSnapshots   uint64
	LastUpdate       time.Time
}

// StorageIntegrationConfig holds configuration for storage integration
type StorageIntegrationConfig struct {
	// Contract storage settings
	EnableContractStorage bool
	MaxContractStorage    uint64
	EnableStorageHistory  bool
	MaxHistorySnapshots   int

	// Storage optimization settings
	EnableCompression    bool
	EnableDeduplication  bool
	CompressionThreshold uint64

	// Pruning settings
	EnableAutoPruning bool
	PruningInterval   time.Duration
	MaxStorageAge     time.Duration
}

// ContractStorageData represents contract storage within the storage system
type ContractStorageData struct {
	Address     engine.Address
	Code        []byte
	CodeHash    engine.Hash
	Balance     *big.Int
	Nonce       uint64
	StorageRoot engine.Hash
	StorageTrie interface{} // Will integrate with existing trie
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Size        uint64
	Compressed  bool
}

// StorageSnapshot represents a storage snapshot
type StorageSnapshot struct {
	BlockNumber uint64
	StorageHash engine.Hash
	Timestamp   time.Time
	Size        uint64
	Changes     []StorageChange
}

// StorageChange represents a change to contract storage
type StorageChange struct {
	Key       engine.Hash
	OldValue  []byte
	NewValue  []byte
	Type      StorageChangeType
	Timestamp time.Time
}

// StorageChangeType indicates the type of storage change
type StorageChangeType int

const (
	StorageChangeTypeCode StorageChangeType = iota
	StorageChangeTypeBalance
	StorageChangeTypeNonce
	StorageChangeTypeStorage
	StorageChangeTypeMetadata
)

// NewStorageIntegration creates a new storage integration
func NewStorageIntegration(
	contractStateManager *ContractStateManager,
	statePruningManager *StatePruningManager,
	config StorageIntegrationConfig,
) *StorageIntegration {
	return &StorageIntegration{
		contractStateManager: contractStateManager,
		statePruningManager:  statePruningManager,
		storageManager:       nil, // Will be initialized separately
		trieManager:          nil, // Will be initialized separately
		pruningManager:       nil, // Will be initialized separately
		contractStorage:      make(map[engine.Address]*ContractStorageData),
		storageHistory:       make(map[engine.Address][]StorageSnapshot),
		config:               config,
		TotalContracts:       0,
		TotalStorageSize:     0,
		TotalSnapshots:       0,
		LastUpdate:           time.Now(),
	}
}

// InitializeStorageManager initializes the storage manager
func (si *StorageIntegration) InitializeStorageManager(storageManager interface{}) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.storageManager = storageManager
}

// InitializeTrieManager initializes the trie manager
func (si *StorageIntegration) InitializeTrieManager(trieManager interface{}) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.trieManager = trieManager
}

// InitializePruningManager initializes the pruning manager
func (si *StorageIntegration) InitializePruningManager(pruningManager interface{}) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.pruningManager = pruningManager
}

// CreateContractStorage creates storage for a new contract
func (si *StorageIntegration) CreateContractStorage(
	address engine.Address,
	code []byte,
	balance *big.Int,
) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.config.EnableContractStorage {
		return ErrContractStorageNotEnabled
	}

	// Check if contract already exists
	if _, exists := si.contractStorage[address]; exists {
		return ErrContractAlreadyExists
	}

	// Calculate code hash
	codeHash := sha256.Sum256(code)

	// Create contract storage
	contractStorage := &ContractStorageData{
		Address:     address,
		Code:        make([]byte, len(code)),
		CodeHash:    engine.Hash(codeHash),
		Balance:     new(big.Int).Set(balance),
		Nonce:       0,
		StorageRoot: engine.Hash{},
		StorageTrie: nil, // Will be initialized with trie manager
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Size:        uint64(len(code)),
		Compressed:  false,
	}

	// Copy code to avoid external modifications
	copy(contractStorage.Code, code)

	// Store contract storage
	si.contractStorage[address] = contractStorage

	// Update statistics
	si.TotalContracts++
	si.TotalStorageSize += contractStorage.Size
	si.LastUpdate = time.Now()

	// Create initial snapshot
	if si.config.EnableStorageHistory {
		si.createStorageSnapshot(address, 0, []StorageChange{})
	}

	return nil
}

// UpdateContractStorage updates contract storage
func (si *StorageIntegration) UpdateContractStorage(
	address engine.Address,
	changes []StorageChange,
	blockNumber uint64,
) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	contractStorage, exists := si.contractStorage[address]
	if !exists {
		return ErrContractNotFound
	}

	// Apply changes
	for _, change := range changes {
		if err := si.applyStorageChange(contractStorage, change); err != nil {
			return err
		}
	}

	// Update metadata
	contractStorage.UpdatedAt = time.Now()

	// Update size
	oldSize := contractStorage.Size
	contractStorage.Size = si.calculateStorageSize(contractStorage)
	sizeDiff := contractStorage.Size - oldSize
	si.TotalStorageSize += sizeDiff

	// Create snapshot
	if si.config.EnableStorageHistory {
		si.createStorageSnapshot(address, blockNumber, changes)
	}

	si.LastUpdate = time.Now()

	return nil
}

// GetContractStorage returns contract storage
func (si *StorageIntegration) GetContractStorage(address engine.Address) *ContractStorageData {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if storage, exists := si.contractStorage[address]; exists {
		// Return a copy to avoid race conditions
		storageCopy := &ContractStorageData{
			Address:     storage.Address,
			Code:        make([]byte, len(storage.Code)),
			CodeHash:    storage.CodeHash,
			Balance:     new(big.Int).Set(storage.Balance),
			Nonce:       storage.Nonce,
			StorageRoot: storage.StorageRoot,
			StorageTrie: storage.StorageTrie,
			CreatedAt:   storage.CreatedAt,
			UpdatedAt:   storage.UpdatedAt,
			Size:        storage.Size,
			Compressed:  storage.Compressed,
		}

		copy(storageCopy.Code, storage.Code)

		return storageCopy
	}

	return nil
}

// GetStorageHistory returns storage history for a contract
func (si *StorageIntegration) GetStorageHistory(
	address engine.Address,
	limit int,
) []StorageSnapshot {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if history, exists := si.storageHistory[address]; exists {
		// Return recent snapshots
		if limit > 0 && len(history) > limit {
			history = history[len(history)-limit:]
		}

		// Return copies to avoid race conditions
		snapshots := make([]StorageSnapshot, len(history))
		for i, snapshot := range history {
			snapshots[i] = StorageSnapshot{
				BlockNumber: snapshot.BlockNumber,
				StorageHash: snapshot.StorageHash,
				Timestamp:   snapshot.Timestamp,
				Size:        snapshot.Size,
				Changes:     make([]StorageChange, len(snapshot.Changes)),
			}

			for j, change := range snapshot.Changes {
				snapshots[i].Changes[j] = StorageChange{
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

// OptimizeStorage optimizes contract storage
func (si *StorageIntegration) OptimizeStorage() error {
	if !si.config.EnableCompression {
		return ErrCompressionNotEnabled
	}

	si.mu.Lock()
	defer si.mu.Unlock()

	// Optimize each contract
	for _, storage := range si.contractStorage {
		if storage.Size > si.config.CompressionThreshold && !storage.Compressed {
			if err := si.compressContractStorage(storage); err != nil {
				return err
			}
		}
	}

	// Update statistics
	si.updateStorageStatistics()

	return nil
}

// PruneStorage prunes old storage data
func (si *StorageIntegration) PruneStorage() error {
	if !si.config.EnableAutoPruning {
		return ErrAutoPruningNotEnabled
	}

	si.mu.Lock()
	defer si.mu.Unlock()

	// Prune old snapshots
	for address, history := range si.storageHistory {
		if len(history) > si.config.MaxHistorySnapshots {
			// Keep only recent snapshots
			si.storageHistory[address] = history[len(history)-si.config.MaxHistorySnapshots:]
		}
	}

	// Prune old contracts
	cutoffTime := time.Now().Add(-si.config.MaxStorageAge)
	for address, storage := range si.contractStorage {
		if storage.UpdatedAt.Before(cutoffTime) {
			// Remove old contract
			delete(si.contractStorage, address)
			delete(si.storageHistory, address)

			// Update statistics
			si.TotalContracts--
			si.TotalStorageSize -= storage.Size
		}
	}

	si.LastUpdate = time.Now()

	return nil
}

// GetStorageStatistics returns storage statistics
func (si *StorageIntegration) GetStorageStatistics() *StorageStatistics {
	si.mu.RLock()
	defer si.mu.RUnlock()

	return &StorageStatistics{
		TotalContracts:   si.TotalContracts,
		TotalStorageSize: si.TotalStorageSize,
		TotalSnapshots:   si.TotalSnapshots,
		LastUpdate:       si.LastUpdate,
		Config:           si.config,
	}
}

// Helper functions
func (si *StorageIntegration) applyStorageChange(
	storage *ContractStorageData,
	change StorageChange,
) error {
	switch change.Type {
	case StorageChangeTypeCode:
		if change.NewValue != nil {
			storage.Code = make([]byte, len(change.NewValue))
			copy(storage.Code, change.NewValue)
			storage.CodeHash = sha256.Sum256(change.NewValue)
		}
	case StorageChangeTypeBalance:
		if change.NewValue != nil {
			storage.Balance = new(big.Int).SetBytes(change.NewValue)
		}
	case StorageChangeTypeNonce:
		if change.NewValue != nil {
			storage.Nonce = new(big.Int).SetBytes(change.NewValue).Uint64()
		}
	case StorageChangeTypeStorage:
		// Storage changes are handled by the trie manager
		return nil
	case StorageChangeTypeMetadata:
		// Handle metadata changes
		return nil
	}

	return nil
}

func (si *StorageIntegration) calculateStorageSize(storage *ContractStorageData) uint64 {
	size := uint64(len(storage.Code))

	// Add balance size
	if storage.Balance != nil {
		size += uint64(len(storage.Balance.Bytes()))
	}

	// Add nonce size
	size += 8 // uint64

	// Add storage root size
	size += 32 // hash size

	return size
}

func (si *StorageIntegration) createStorageSnapshot(
	address engine.Address,
	blockNumber uint64,
	changes []StorageChange,
) {
	snapshot := StorageSnapshot{
		BlockNumber: blockNumber,
		StorageHash: engine.Hash{}, // Will be calculated
		Timestamp:   time.Now(),
		Size:        0, // Will be calculated
		Changes:     make([]StorageChange, len(changes)),
	}

	// Copy changes
	copy(snapshot.Changes, changes)

	// Calculate storage hash and size
	if storage, exists := si.contractStorage[address]; exists {
		snapshot.Size = storage.Size
		// In a real implementation, calculate proper storage hash
	}

	// Add to history
	if si.storageHistory[address] == nil {
		si.storageHistory[address] = make([]StorageSnapshot, 0)
	}
	si.storageHistory[address] = append(si.storageHistory[address], snapshot)

	// Update statistics
	si.TotalSnapshots++
}

func (si *StorageIntegration) compressContractStorage(storage *ContractStorageData) error {
	// In a real implementation, this would compress the contract storage
	// For now, just mark as compressed

	oldSize := storage.Size
	storage.Compressed = true
	storage.Size = oldSize * 80 / 100 // Assume 20% compression

	// Update total storage size
	sizeDiff := storage.Size - oldSize
	si.TotalStorageSize += sizeDiff

	return nil
}

func (si *StorageIntegration) updateStorageStatistics() {
	// Recalculate total storage size
	totalSize := uint64(0)
	for _, storage := range si.contractStorage {
		totalSize += storage.Size
	}
	si.TotalStorageSize = totalSize
}

// StorageStatistics contains storage statistics
type StorageStatistics struct {
	TotalContracts   uint64
	TotalStorageSize uint64
	TotalSnapshots   uint64
	LastUpdate       time.Time
	Config           StorageIntegrationConfig
}
