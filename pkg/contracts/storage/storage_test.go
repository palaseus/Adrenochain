package storage

import (
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

func TestStorageIntegrationCreation(t *testing.T) {
	// Test creating storage integration
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   10,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1000,
		EnableAutoPruning:     true,
		PruningInterval:       1 * time.Hour,
		MaxStorageAge:         24 * time.Hour,
	}

	// Create mock managers (nil for now)
	var contractStateManager *ContractStateManager
	var statePruningManager *StatePruningManager

	storageIntegration := NewStorageIntegration(
		contractStateManager,
		statePruningManager,
		config,
	)

	if storageIntegration == nil {
		t.Fatal("StorageIntegration should not be nil")
	}

	if storageIntegration.config != config {
		t.Error("Configuration should match")
	}

	if storageIntegration.TotalContracts != 0 {
		t.Error("Initial total contracts should be 0")
	}

	if storageIntegration.TotalStorageSize != 0 {
		t.Error("Initial total storage size should be 0")
	}
}

func TestContractStorageDataCreation(t *testing.T) {
	// Test creating contract storage data
	address := engine.Address{}
	code := []byte("test code")
	// balance := uint64(1000) // Will be used when implementing balance functionality

	storageData := &ContractStorageData{
		Address:     address,
		Code:        code,
		CodeHash:    engine.Hash{},
		Balance:     nil, // Will be set later
		Nonce:       0,
		StorageRoot: engine.Hash{},
		StorageTrie: nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Size:        uint64(len(code)),
		Compressed:  false,
	}

	if storageData == nil {
		t.Fatal("ContractStorageData should not be nil")
	}

	if storageData.Size != uint64(len(code)) {
		t.Error("Size should match code length")
	}

	if storageData.Compressed {
		t.Error("Should not be compressed initially")
	}
}

func TestStorageChangeTypes(t *testing.T) {
	// Test storage change types
	if StorageChangeTypeCode != 0 {
		t.Error("StorageChangeTypeCode should be 0")
	}

	if StorageChangeTypeBalance != 1 {
		t.Error("StorageChangeTypeBalance should be 1")
	}

	if StorageChangeTypeNonce != 2 {
		t.Error("StorageChangeTypeNonce should be 2")
	}

	if StorageChangeTypeStorage != 3 {
		t.Error("StorageChangeTypeStorage should be 3")
	}

	if StorageChangeTypeMetadata != 4 {
		t.Error("StorageChangeTypeMetadata should be 4")
	}
}

func TestStorageSnapshotCreation(t *testing.T) {
	// Test creating storage snapshot
	snapshot := StorageSnapshot{
		BlockNumber: 12345,
		StorageHash: engine.Hash{},
		Timestamp:   time.Now(),
		Size:        1024,
		Changes:     []StorageChange{},
	}

	if snapshot.BlockNumber != 12345 {
		t.Error("Block number should match")
	}

	if snapshot.Size != 1024 {
		t.Error("Size should match")
	}

	if len(snapshot.Changes) != 0 {
		t.Error("Changes should be empty initially")
	}
}
