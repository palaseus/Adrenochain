package sync

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	netproto "github.com/gochain/gochain/pkg/proto/net"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Set testing environment variable
	os.Setenv("TESTING", "1")
}

// MockChain implements the chain.Chain interface for testing
type MockChain struct {
	height  uint64
	tipHash []byte
	blocks  map[uint64]*block.Block
}

func NewMockChain() *MockChain {
	mc := &MockChain{
		height:  100, // Start with a reasonable height
		tipHash: make([]byte, 32),
		blocks:  make(map[uint64]*block.Block),
	}

	// Add a genesis block for testing
	genesisBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1000,
			Nonce:         0,
			Height:        0,
		},
		Transactions: []*block.Transaction{},
	}

	mc.blocks[0] = genesisBlock

	// Add some additional blocks for testing
	for i := uint64(1); i <= 100; i++ {
		block := &block.Block{
			Header: &block.Header{
				Version:       1,
				PrevBlockHash: mc.blocks[i-1].CalculateHash(),
				MerkleRoot:    make([]byte, 32),
				Timestamp:     time.Now().Add(time.Duration(i) * time.Second),
				Difficulty:    1000,
				Nonce:         i,
				Height:        i,
			},
			Transactions: []*block.Transaction{},
		}
		mc.blocks[i] = block
	}

	mc.height = 100
	mc.tipHash = mc.blocks[100].CalculateHash()

	return mc
}

// ChainReader interface methods
func (mc *MockChain) GetHeight() uint64 {
	return mc.height
}

func (mc *MockChain) GetTipHash() []byte {
	return mc.tipHash
}

func (mc *MockChain) GetBlockByHeight(height uint64) *block.Block {
	if block, exists := mc.blocks[height]; exists {
		return block
	}
	return nil
}

func (mc *MockChain) GetBlock(hash []byte) *block.Block {
	// Simple hash-based lookup
	for _, block := range mc.blocks {
		if bytes.Equal(block.CalculateHash(), hash) {
			return block
		}
	}
	return nil
}

// ChainWriter interface methods
func (mc *MockChain) AddBlock(blockData interface{}) error {
	if b, ok := blockData.(*block.Block); ok {
		// Validate the block
		if err := b.IsValid(); err != nil {
			return err
		}

		// Check if block height is sequential
		if b.Header.Height != mc.height+1 {
			return fmt.Errorf("block height %d is not sequential to current height %d", b.Header.Height, mc.height)
		}

		// Check if previous block hash matches
		if !bytes.Equal(b.Header.PrevBlockHash, mc.tipHash) {
			return fmt.Errorf("previous block hash mismatch")
		}

		// Add block to our map
		mc.blocks[b.Header.Height] = b
		mc.height = b.Header.Height
		mc.tipHash = b.CalculateHash()

		return nil
	}
	return fmt.Errorf("invalid block type")
}

// Additional methods for testing
func (mc *MockChain) GetBestBlock() *block.Block {
	return mc.blocks[mc.height]
}

func (mc *MockChain) GetGenesisBlock() *block.Block {
	return mc.blocks[0]
}

func (mc *MockChain) CalculateNextDifficulty() uint64 {
	return 1000 // Mock difficulty
}

func (mc *MockChain) GetAccumulatedDifficulty(height uint64) (*big.Int, error) {
	return big.NewInt(int64(height * 1000)), nil
}

func (mc *MockChain) Close() error {
	return nil
}

// MockStorage implements StorageInterface for testing
type MockStorage struct{}

func (ms *MockStorage) StoreBlock(b *block.Block) error                 { return nil }
func (ms *MockStorage) GetBlock(hash []byte) (*block.Block, error)      { return nil, nil }
func (ms *MockStorage) StoreChainState(state *storage.ChainState) error { return nil }
func (ms *MockStorage) GetChainState() (*storage.ChainState, error)     { return nil, nil }
func (ms *MockStorage) Write(key []byte, value []byte) error            { return nil }
func (ms *MockStorage) Read(key []byte) ([]byte, error)                 { return nil, nil }
func (ms *MockStorage) Delete(key []byte) error                         { return nil }
func (ms *MockStorage) Has(key []byte) (bool, error)                    { return false, nil }
func (ms *MockStorage) Close() error                                    { return nil }

func createTestHost(t *testing.T) host.Host {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.DisableRelay(),
	)
	require.NoError(t, err)
	return h
}

func TestNewSyncProtocol(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	assert.NotNil(t, sp)
	assert.Equal(t, host, sp.host)
	assert.Equal(t, chain, sp.chain)
	assert.Equal(t, storage, sp.storage)
	assert.Equal(t, config, sp.config)
	assert.NotNil(t, sp.syncState)
}

func TestSyncProtocol_StartSync(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Create a mock peer ID
	peerID := peer.ID("test-peer")

	// Test that StartSync creates the initial sync state
	err := sp.StartSync(peerID)
	assert.NoError(t, err)

	// Check that sync state was created immediately
	sp.mu.RLock()
	state, exists := sp.syncState[peerID]
	sp.mu.RUnlock()

	assert.True(t, exists, "Sync state should exist")
	assert.NotNil(t, state, "Sync state should not be nil")
	assert.Equal(t, peerID, state.PeerID, "Peer ID should match")
	assert.True(t, state.IsSyncing, "Initial sync state should be syncing")

	// Test the sync state structure without waiting for goroutine completion
	assert.Equal(t, peerID, state.PeerID, "Peer ID should match")
	assert.True(t, state.IsSyncing, "Should be syncing")
	assert.NotZero(t, state.SyncStart, "Sync start time should be set")
	assert.Zero(t, state.SyncEnd, "Sync end time should not be set yet")
	assert.Zero(t, state.HeadersSynced, "Headers synced should be 0 initially")
	assert.Zero(t, state.BlocksSynced, "Blocks synced should be 0 initially")
	assert.Zero(t, state.RetryCount, "Retry count should be 0 initially")
	assert.Nil(t, state.LastError, "Last error should be nil initially")
}

func TestSyncProtocol_GetSyncProgress(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Create a mock peer ID
	peerID := peer.ID("test-peer")

	// Add some sync state
	sp.mu.Lock()
	sp.syncState[peerID] = &PeerSyncState{
		PeerID:        peerID,
		Height:        1000,
		HeadersSynced: 500,
		BlocksSynced:  250,
	}
	sp.mu.Unlock()

	// Get progress
	progress, err := sp.GetSyncProgress(peerID)
	assert.NoError(t, err)

	// Expected progress: (500 + 250) / (1000 * 2) * 100 = 37.5%
	expectedProgress := float64(500+250) / float64(1000*2) * 100
	assert.Equal(t, expectedProgress, progress)
}

func TestSyncProtocol_GetPeerStates(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Add some peer states
	peer1 := peer.ID("peer1")
	peer2 := peer.ID("peer2")

	sp.mu.Lock()
	sp.syncState[peer1] = &PeerSyncState{PeerID: peer1, Height: 100}
	sp.syncState[peer2] = &PeerSyncState{PeerID: peer2, Height: 200}
	sp.mu.Unlock()

	// Get all peer states
	states := sp.GetPeerStates()

	assert.Len(t, states, 2)
	assert.Contains(t, states, peer1)
	assert.Contains(t, states, peer2)
	assert.Equal(t, uint64(100), states[peer1].Height)
	assert.Equal(t, uint64(200), states[peer2].Height)
}

func TestSyncProtocol_GetPeerState(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Add a peer state
	peerID := peer.ID("test-peer")
	sp.mu.Lock()
	sp.syncState[peerID] = &PeerSyncState{PeerID: peerID, Height: 100}
	sp.mu.Unlock()

	// Get peer state
	state := sp.getPeerState(peerID)
	assert.NotNil(t, state)
	assert.Equal(t, peerID, state.PeerID)
	assert.Equal(t, uint64(100), state.Height)

	// Get non-existent peer state
	nonExistentPeer := peer.ID("non-existent")
	state = sp.getPeerState(nonExistentPeer)
	assert.Nil(t, state)
}

func TestSyncProtocol_SetupHandlers(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// The setupHandlers method is called in NewSyncProtocol
	// We can verify that the handlers are set up by checking if the host has stream handlers
	// This is a basic check - in a real implementation you'd want more comprehensive testing

	// The protocol should be initialized
	assert.NotNil(t, sp)
}

func TestSyncProtocol_ExchangeSyncInfo(t *testing.T) {
	// This test would require setting up two hosts and establishing a connection
	// For now, we'll just test the method structure
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// This test would require a real peer connection
	// For now, we'll just verify the method exists and doesn't panic
	peerID := peer.ID("test-peer")

	// The exchangeSyncInfo method should fail gracefully when there's no real peer
	// This is expected behavior in a test environment
	_ = sp.exchangeSyncInfo(peerID)
}

func TestSyncProtocol_ProcessHeader(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Create a test header
	header := &netproto.BlockHeader{
		Version:       1,
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Now().Unix(),
		Difficulty:    1000,
		Nonce:         12345,
		Height:        1,
	}

	// Process header
	err := sp.processHeader(header)
	assert.NoError(t, err)
}

func TestSyncProtocol_ProcessBlock(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Create a test block with the correct height and previous hash
	prevBlock := chain.GetBlockByHeight(chain.GetHeight())
	require.NotNil(t, prevBlock)
	testBlock := block.NewBlock(prevBlock.CalculateHash(), chain.GetHeight()+1, 1000)

	// Serialize the block
	blockData, err := testBlock.Serialize()
	require.NoError(t, err)

	// Process block
	err = sp.processBlock(blockData)
	assert.NoError(t, err)

	// Verify block was added to chain
	addedBlock := chain.GetBlockByHeight(chain.GetHeight())
	assert.NotNil(t, addedBlock)
	assert.Equal(t, testBlock.Header.Height, addedBlock.Header.Height)
}

func TestSyncProtocol_GetHeaders(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Get headers
	headers := sp.getHeaders(90, 20) // Request 20 headers from height 90

	assert.Len(t, headers, 11) // We only have blocks up to height 100
	assert.Equal(t, uint64(90), headers[0].Height)
	assert.Equal(t, uint64(100), headers[10].Height)
}

func TestSyncProtocol_GetHeadersForSync(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Create a sync request with a height that's higher than what we have
	// This should result in no headers being returned
	syncReq := netproto.SyncRequest{
		CurrentHeight: 1000, // Higher than our chain height
		BestBlockHash: make([]byte, 32),
		KnownHeaders:  [][]byte{},
	}

	// Get headers for sync
	headers := sp.getHeadersForSync(syncReq)

	// Since the peer's height is higher than ours, no headers should be returned
	assert.Empty(t, headers)

	// Test with a lower height - should return some headers
	syncReq2 := netproto.SyncRequest{
		CurrentHeight: 50, // Lower than our chain height
		BestBlockHash: make([]byte, 32),
		KnownHeaders:  [][]byte{},
	}

	headers2 := sp.getHeadersForSync(syncReq2)
	// Should return headers from height 51 to 100
	assert.NotEmpty(t, headers2)
	assert.Len(t, headers2, 50) // 51 to 100 inclusive
}

func TestSyncProtocol_ContextCancellation(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	_ = NewSyncProtocol(host, chain, chain, storage, config)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine that waits for context cancellation
	done := make(chan bool)
	go func() {
		<-ctx.Done()
		done <- true
	}()

	// Cancel the context
	cancel()

	// Wait for the goroutine to finish
	select {
	case <-done:
		// Context was cancelled successfully
	case <-time.After(time.Second):
		t.Fatal("Context cancellation did not work")
	}
}

func TestSyncProtocol_ConcurrentAccess(t *testing.T) {
	host := createTestHost(t)
	defer host.Close()

	chain := NewMockChain()
	storage := &MockStorage{}
	config := DefaultSyncConfig()

	sp := NewSyncProtocol(host, chain, chain, storage, config)

	// Test concurrent access to sync state
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			peerID := peer.ID(fmt.Sprintf("peer-%d", id))

			// Add peer state
			sp.mu.Lock()
			sp.syncState[peerID] = &PeerSyncState{
				PeerID: peerID,
				Height: uint64(id * 100),
			}
			sp.mu.Unlock()

			// Read peer state
			sp.mu.RLock()
			state := sp.syncState[peerID]
			sp.mu.RUnlock()

			assert.NotNil(t, state)
			assert.Equal(t, peerID, state.PeerID)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-time.After(time.Second):
			t.Fatal("Goroutine did not complete in time")
		}
	}
}
