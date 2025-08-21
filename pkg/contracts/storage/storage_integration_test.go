package storage

import (
	"math/big"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/stretchr/testify/assert"
)

// TestCreateContractStorage tests the CreateContractStorage function
func TestCreateContractStorage(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	// Create required managers
	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	address := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	code := []byte{0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xF3}
	balance := big.NewInt(1000)

	// Test successful contract creation
	err := si.CreateContractStorage(address, code, balance)
	assert.NoError(t, err)

	// Verify contract was created
	contractStorage := si.GetContractStorage(address)
	assert.NotNil(t, contractStorage)
	assert.Equal(t, address, contractStorage.Address)
	assert.Equal(t, code, contractStorage.Code)
	assert.Equal(t, balance, contractStorage.Balance)
	assert.Equal(t, uint64(0), contractStorage.Nonce)
	assert.False(t, contractStorage.Compressed)

	// Test creating duplicate contract (should fail)
	err = si.CreateContractStorage(address, code, balance)
	assert.Error(t, err)
	assert.Equal(t, ErrContractAlreadyExists, err)

	// Test with contract storage disabled
	si.config.EnableContractStorage = false
	err = si.CreateContractStorage(engine.Address{0x02, 0x03, 0x04, 0x05, 0x06}, code, balance)
	assert.Error(t, err)
	assert.Equal(t, ErrContractStorageNotEnabled, err)
}

// TestUpdateContractStorage tests the UpdateContractStorage function
func TestUpdateContractStorage(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	address := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	code := []byte{0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xF3}
	balance := big.NewInt(1000)

	// Create contract first
	err := si.CreateContractStorage(address, code, balance)
	assert.NoError(t, err)

	// Test updating contract storage
	changes := []StorageChange{
		{
			Type:      StorageChangeTypeCode,
			Key:       engine.Hash{0x10, 0x20, 0x30, 0x40, 0x50},
			OldValue:  code,
			NewValue:  []byte{0x60, 0x01, 0x52, 0x60, 0x21, 0x60, 0x01, 0xF3},
			Timestamp: time.Now(),
		},
	}

	err = si.UpdateContractStorage(address, changes, 1)
	assert.NoError(t, err)

	// Verify contract was updated
	contractStorage := si.GetContractStorage(address)
	assert.NotNil(t, contractStorage)
	assert.NotEqual(t, code, contractStorage.Code)

	// Test updating non-existent contract (should fail)
	err = si.UpdateContractStorage(engine.Address{0x02, 0x03, 0x04, 0x05, 0x06}, changes, 1)
	assert.Error(t, err)
	assert.Equal(t, ErrContractNotFound, err)
}

// TestGetContractStorage tests the GetContractStorage function
func TestGetContractStorage(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	address := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	code := []byte{0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xF3}
	balance := big.NewInt(1000)

	// Test getting non-existent contract (should return nil)
	contractStorage := si.GetContractStorage(address)
	assert.Nil(t, contractStorage)

	// Create contract
	err := si.CreateContractStorage(address, code, balance)
	assert.NoError(t, err)

	// Test getting existing contract
	contractStorage = si.GetContractStorage(address)
	assert.NotNil(t, contractStorage)
	assert.Equal(t, address, contractStorage.Address)
	assert.Equal(t, code, contractStorage.Code)
	assert.Equal(t, balance, contractStorage.Balance)
}

// TestGetStorageHistory tests the GetStorageHistory function
func TestGetStorageHistory(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	address := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	code := []byte{0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xF3}
	balance := big.NewInt(1000)

	// Create contract
	err := si.CreateContractStorage(address, code, balance)
	assert.NoError(t, err)

	// Test getting storage history
	history := si.GetStorageHistory(address, 10)
	assert.NotNil(t, history)
	// History might be empty initially, but the function should not panic
}

// TestOptimizeStorage tests the OptimizeStorage function
func TestOptimizeStorage(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	// Test storage optimization
	err := si.OptimizeStorage()
	assert.NoError(t, err)

	// Verify optimization was performed
	stats := si.GetStorageStatistics()
	assert.NotNil(t, stats)
}

// TestPruneStorage tests the PruneStorage function
func TestPruneStorage(t *testing.T) {
	config := StorageIntegrationConfig{
		EnableContractStorage: true,
		MaxContractStorage:    1000,
		EnableStorageHistory:  true,
		MaxHistorySnapshots:   100,
		EnableCompression:     true,
		EnableDeduplication:   true,
		CompressionThreshold:  1024 * 1024,
		EnableAutoPruning:     true,
		PruningInterval:       time.Hour,
		MaxStorageAge:         time.Hour * 24 * 7,
	}

	csmConfig := ContractStateConfig{
		MaxHistorySize:     100,
		EnableStatePruning: true,
		PruningInterval:    time.Hour,
		MaxStorageSize:     1000,
		EnableCompression:  true,
		SnapshotInterval:   time.Hour,
	}

	csm := NewContractStateManager(&MockStorageManager{}, &MockTrieManager{}, csmConfig)
	spm := NewStatePruningManager(csm, PruningConfig{})
	si := NewStorageIntegration(csm, spm, config)

	// Test storage pruning
	err := si.PruneStorage()
	assert.NoError(t, err)

	// Verify pruning was performed
	stats := si.GetStorageStatistics()
	assert.NotNil(t, stats)
}
