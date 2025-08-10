package sync

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/storage"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockChain and MockStorage are defined in protocol_test.go to avoid duplication

func TestNewSyncManager(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := NewMockChain()
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	assert.NotNil(t, syncManager)
	assert.Equal(t, mockChain, syncManager.chain)
	assert.Equal(t, mockStorage, syncManager.storage)
	assert.Equal(t, config, syncManager.config)
	assert.NotNil(t, syncManager.peers)
	assert.False(t, syncManager.status.IsSyncing)
}

func TestSyncConfigDefaults(t *testing.T) {
	config := DefaultSyncConfig()

	assert.True(t, config.FastSyncEnabled)
	assert.False(t, config.LightClientEnabled)
	assert.Equal(t, 5, config.MaxSyncPeers)
	assert.Equal(t, 30*time.Second, config.SyncTimeout)
	assert.Equal(t, uint64(1000), config.BlockDownloadLimit)
	assert.True(t, config.StateSyncEnabled)
	assert.Equal(t, uint64(10000), config.CheckpointInterval)
}

func TestStartAndStopSync(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Test starting sync
	err := syncManager.StartSync()
	assert.NoError(t, err)
	assert.True(t, syncManager.status.IsSyncing)
	assert.Equal(t, uint64(100), syncManager.status.CurrentHeight)

	// Test stopping sync
	syncManager.StopSync()
	assert.False(t, syncManager.status.IsSyncing)
}

func TestAddAndRemovePeer(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := NewMockChain()
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Add peer
	peerID := "peer1"
	peerAddress := "127.0.0.1:8080"
	peerHeight := uint64(150)
	syncManager.AddPeer(peerID, peerAddress, peerHeight)

	// Verify peer was added
	peer, exists := syncManager.peers[peerID]
	assert.True(t, exists)
	assert.Equal(t, peerID, peer.ID)
	assert.Equal(t, peerAddress, peer.Address)
	assert.Equal(t, peerHeight, peer.Height)
	assert.Equal(t, "connected", peer.ConnectionState)

	// Remove peer
	syncManager.RemovePeer(peerID)
	_, exists = syncManager.peers[peerID]
	assert.False(t, exists)
}

func TestFindBestPeer(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Add multiple peers with different heights
	syncManager.AddPeer("peer1", "127.0.0.1:8080", 120)
	syncManager.AddPeer("peer2", "127.0.0.1:8081", 150)
	syncManager.AddPeer("peer3", "127.0.0.1:8082", 130)

	// Find best peer
	bestPeer := syncManager.findBestPeer()
	assert.NotNil(t, bestPeer)
	assert.Equal(t, "peer2", bestPeer.ID)
	assert.Equal(t, uint64(150), bestPeer.Height)
}

func TestSyncStatus(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Get initial status
	status := syncManager.GetStatus()
	assert.False(t, status.IsSyncing)
	assert.Equal(t, uint64(0), status.CurrentHeight)
	assert.Equal(t, uint64(0), status.TargetHeight)
	assert.Equal(t, 0, status.PeersConnected)

	// Start sync and check status
	err := syncManager.StartSync()
	require.NoError(t, err)

	status = syncManager.GetStatus()
	assert.True(t, status.IsSyncing)
	assert.Equal(t, uint64(100), status.CurrentHeight)
}

func TestValidateCheckpoint(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Test checkpoint validation (placeholder implementation)
	height := uint64(10000)
	hash := []byte("test_hash_12345")

	isValid := syncManager.ValidateCheckpoint(height, hash)
	assert.True(t, isValid) // Current implementation always returns true
}

func TestGetSyncProgress(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	peerID := peer.ID("test-peer")

	// The sync protocol should be initialized when host is provided
	assert.NotNil(t, syncManager.syncProtocol)

	// Test progress before starting sync (should return error because no peer state exists)
	_, err := syncManager.GetSyncProgress(peerID)
	assert.Error(t, err) // Should error because no peer state exists yet

	// Initialize sync protocol by starting sync
	err = syncManager.StartSync()
	assert.NoError(t, err)

	// Start sync with a specific peer to create peer state
	err = syncManager.StartSyncWithPeer(peerID)
	assert.NoError(t, err)

	// Wait a bit for sync to complete
	time.Sleep(200 * time.Millisecond)

	// Now test progress (should work and return progress since peer state exists)
	progress, err := syncManager.GetSyncProgress(peerID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, progress, 0.0) // Should have some progress
}

func TestSyncManagerClose(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Start sync
	err := syncManager.StartSync()
	require.NoError(t, err)
	assert.True(t, syncManager.status.IsSyncing)

	// Close sync manager
	err = syncManager.Close()
	assert.NoError(t, err)
	assert.False(t, syncManager.status.IsSyncing)
}

func TestSyncManagerConcurrency(t *testing.T) {
	config := DefaultSyncConfig()
	mockChain := &MockChain{height: 100}
	mockStorage := &MockStorage{}
	host := createTestHost(t)
	defer host.Close()

	syncManager := NewSyncManager(mockChain, mockChain, mockStorage, config, host)

	// Test concurrent peer operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			peerID := fmt.Sprintf("peer%d", id)
			syncManager.AddPeer(peerID, "127.0.0.1:8080", uint64(100+id))
			syncManager.RemovePeer(peerID)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no peers remain
	assert.Equal(t, 0, len(syncManager.peers))
}

func TestSyncManagerWithRealChain(t *testing.T) {
	// Create temporary storage
	dataDir := "./test_sync_data"
	defer func() {
		os.RemoveAll(dataDir)
	}()

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	require.NoError(t, err)
	defer storage.Close()

	// Create sync manager with mock chain
	syncConfig := DefaultSyncConfig()
	mockChain := NewMockChain()
	host := createTestHost(t)
	defer host.Close()
	syncManager := NewSyncManager(mockChain, mockChain, storage, syncConfig, host)

	// Test basic operations
	assert.NotNil(t, syncManager)
	assert.Equal(t, mockChain, syncManager.chain)
	assert.Equal(t, storage, syncManager.storage)

	// Test sync start
	err = syncManager.StartSync()
	assert.NoError(t, err)

	// Test sync stop
	syncManager.StopSync()
	assert.False(t, syncManager.status.IsSyncing)

	// Clean up
	syncManager.Close()
}
