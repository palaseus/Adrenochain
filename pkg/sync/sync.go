package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ChainReader defines the interface that the sync package needs from the chain
type ChainReader interface {
	GetHeight() uint64
	GetTipHash() []byte
	GetBlockByHeight(height uint64) *block.Block
	// Additional methods needed by the sync protocol
	GetBlock(hash []byte) *block.Block
}

// ChainWriter defines the interface for adding blocks to the chain
type ChainWriter interface {
	AddBlock(block interface{}) error
}

// BlockInterface defines the interface that blocks must implement for sync operations
type BlockInterface interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	IsValid() error
	CalculateHash() []byte
	GetHeader() interface{}
}

// HeaderInterface defines the interface that block headers must implement
type HeaderInterface interface {
	GetVersion() uint32
	GetPrevBlockHash() []byte
	GetMerkleRoot() []byte
	GetTimestamp() time.Time
	GetDifficulty() uint64
	GetNonce() uint64
	GetHeight() uint64
	IsValid() error
}

// ChainAdapter adapts the concrete *chain.Chain type to our interfaces
type ChainAdapter struct {
	chain *chain.Chain
}

// NewChainAdapter creates a new chain adapter
func NewChainAdapter(chain *chain.Chain) *ChainAdapter {
	return &ChainAdapter{chain: chain}
}

// GetHeight returns the current height of the chain
func (ca *ChainAdapter) GetHeight() uint64 {
	return ca.chain.GetHeight()
}

// GetTipHash returns the hash of the current best block
func (ca *ChainAdapter) GetTipHash() []byte {
	return ca.chain.GetTipHash()
}

// GetBlockByHeight returns a block by height
func (ca *ChainAdapter) GetBlockByHeight(height uint64) *block.Block {
	return ca.chain.GetBlockByHeight(height)
}

// GetBlock returns a block by hash
func (ca *ChainAdapter) GetBlock(hash []byte) *block.Block {
	return ca.chain.GetBlock(hash)
}

// AddBlock adds a block to the chain
func (ca *ChainAdapter) AddBlock(blockData interface{}) error {
	if b, ok := blockData.(*block.Block); ok {
		return ca.chain.AddBlock(b)
	}
	return fmt.Errorf("invalid block type")
}

// SyncManager manages blockchain synchronization between nodes.
// It implements fast sync, light client sync, and state synchronization protocols.
type SyncManager struct {
	mu          sync.RWMutex
	chain       ChainReader
	chainWriter ChainWriter
	storage     storage.StorageInterface
	config      *SyncConfig
	status      SyncStatus
	peers       map[string]*PeerInfo

	// New sync protocol
	syncProtocol *SyncProtocol
	host         host.Host

	ctx    context.Context
	cancel context.CancelFunc
}

// SyncConfig holds configuration parameters for synchronization.
type SyncConfig struct {
	FastSyncEnabled    bool          // FastSyncEnabled enables fast synchronization mode
	LightClientEnabled bool          // LightClientEnabled enables light client mode
	MaxSyncPeers       int           // MaxSyncPeers is the maximum number of peers to sync with
	SyncTimeout        time.Duration // SyncTimeout is the timeout for sync operations
	BlockDownloadLimit uint64        // BlockDownloadLimit is the maximum blocks to download per request
	StateSyncEnabled   bool          // StateSyncEnabled enables state synchronization
	CheckpointInterval uint64        // CheckpointInterval is the height interval for checkpoints
}

// DefaultSyncConfig returns the default synchronization configuration.
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		FastSyncEnabled:    true,
		LightClientEnabled: false,
		MaxSyncPeers:       5,
		SyncTimeout:        30 * time.Second,
		BlockDownloadLimit: 1000,
		StateSyncEnabled:   true,
		CheckpointInterval: 10000,
	}
}

// SyncStatus represents the current synchronization status.
type SyncStatus struct {
	IsSyncing        bool      // IsSyncing indicates if synchronization is in progress
	StartTime        time.Time // StartTime is when synchronization started
	CurrentHeight    uint64    // CurrentHeight is the current blockchain height
	TargetHeight     uint64    // TargetHeight is the target height to sync to
	PeersConnected   int       // PeersConnected is the number of connected peers
	BlocksDownloaded uint64    // BlocksDownloaded is the number of blocks downloaded
	LastBlockTime    time.Time // LastBlockTime is the timestamp of the last block
}

// PeerInfo represents information about a peer during synchronization.
type PeerInfo struct {
	ID              string    // ID is the peer identifier
	Address         string    // Address is the peer's network address
	Height          uint64    // Height is the peer's blockchain height
	LastSeen        time.Time // LastSeen is when the peer was last seen
	IsTrusted       bool      // IsTrusted indicates if this peer is trusted
	ConnectionState string    // ConnectionState is the current connection state
}

// NewSyncManager creates a new synchronization manager.
func NewSyncManager(chain ChainReader, chainWriter ChainWriter, storage storage.StorageInterface, config *SyncConfig, host host.Host) *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())

	sm := &SyncManager{
		chain:       chain,
		chainWriter: chainWriter,
		storage:     storage,
		config:      config,
		peers:       make(map[string]*PeerInfo),
		host:        host,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize sync protocol if host is provided
	if host != nil {
		// Create a simplified sync protocol that works with interfaces
		sm.syncProtocol = NewSyncProtocol(host, chain, chainWriter, storage, config)
	}

	return sm
}

// StartSync begins the synchronization process with connected peers.
func (sm *SyncManager) StartSync() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.status.IsSyncing {
		return fmt.Errorf("sync already in progress")
	}

	// Ensure sync protocol is initialized
	if sm.syncProtocol == nil {
		return fmt.Errorf("sync protocol not initialized")
	}

	sm.status.IsSyncing = true
	sm.status.StartTime = time.Now()
	sm.status.CurrentHeight = sm.chain.GetHeight()

	// Start sync in background
	go sm.syncLoop()

	return nil
}

// StartSyncWithPeer initiates synchronization with a specific peer using the new protocol
func (sm *SyncManager) StartSyncWithPeer(peerID peer.ID) error {
	if sm.syncProtocol == nil {
		return fmt.Errorf("sync protocol not initialized")
	}

	return sm.syncProtocol.StartSync(peerID)
}

// GetSyncProgress returns the sync progress for a specific peer
func (sm *SyncManager) GetSyncProgress(peerID peer.ID) (float64, error) {
	if sm.syncProtocol == nil {
		return 0, fmt.Errorf("sync protocol not initialized")
	}

	return sm.syncProtocol.GetSyncProgress(peerID)
}

// GetPeerStates returns all peer sync states
func (sm *SyncManager) GetPeerStates() map[peer.ID]*PeerSyncState {
	if sm.syncProtocol == nil {
		return make(map[peer.ID]*PeerSyncState)
	}

	return sm.syncProtocol.GetPeerStates()
}

// StopSync stops the synchronization process.
func (sm *SyncManager) StopSync() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.status.IsSyncing = false
	sm.cancel()
}

// GetStatus returns the current synchronization status.
func (sm *SyncManager) GetStatus() SyncStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.status
}

// AddPeer adds a peer for synchronization.
func (sm *SyncManager) AddPeer(id, address string, height uint64) {
	// Don't add peers with empty IDs
	if id == "" {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.peers[id] = &PeerInfo{
		ID:              id,
		Address:         address,
		Height:          height,
		LastSeen:        time.Now(),
		IsTrusted:       false,
		ConnectionState: "connected",
	}
}

// RemovePeer removes a peer from synchronization.
func (sm *SyncManager) RemovePeer(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.peers, id)
}

// syncLoop is the main synchronization loop.
func (sm *SyncManager) syncLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.performSync()
		}
	}
}

// performSync performs one synchronization cycle.
func (sm *SyncManager) performSync() {
	sm.mu.Lock()
	if !sm.status.IsSyncing {
		sm.mu.Unlock()
		return
	}
	sm.mu.Unlock()

	// Find best peer
	bestPeer := sm.findBestPeer()
	if bestPeer == nil {
		return
	}

	// Check if we need to sync
	if bestPeer.Height <= sm.chain.GetHeight() {
		return
	}

	// Perform fast sync if enabled
	if sm.config.FastSyncEnabled {
		sm.performFastSync(bestPeer)
	}
}

// findBestPeer finds the peer with the highest blockchain height.
func (sm *SyncManager) findBestPeer() *PeerInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return nil if no peers exist
	if len(sm.peers) == 0 {
		return nil
	}

	var bestPeer *PeerInfo
	var bestHeight uint64

	for _, peer := range sm.peers {
		if peer.Height > bestHeight {
			bestHeight = peer.Height
			bestPeer = peer
		}
	}

	return bestPeer
}

// performFastSync performs fast synchronization with a peer.
func (sm *SyncManager) performFastSync(peer *PeerInfo) {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Download block headers in batches
	// 2. Validate proof of work
	// 3. Download blocks in parallel
	// 4. Validate and apply blocks

	sm.mu.Lock()
	sm.status.TargetHeight = peer.Height
	sm.mu.Unlock()

	// Simulate block download
	sm.simulateBlockDownload(peer)
}

// simulateBlockDownload simulates downloading blocks from a peer.
func (sm *SyncManager) simulateBlockDownload(peer *PeerInfo) {
	// This is a placeholder for actual block download logic
	// In a real implementation, this would download actual blocks

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Simulate progress
	sm.status.BlocksDownloaded += 100
	if sm.status.BlocksDownloaded > peer.Height-sm.status.CurrentHeight {
		sm.status.BlocksDownloaded = peer.Height - sm.status.CurrentHeight
	}

	sm.status.LastBlockTime = time.Now()
}

// ValidateCheckpoint validates a checkpoint at the given height.
func (sm *SyncManager) ValidateCheckpoint(height uint64, hash []byte) bool {
	// This would validate against known checkpoints
	// For now, return true as a placeholder
	return true
}

// Close closes the synchronization manager.
func (sm *SyncManager) Close() error {
	sm.StopSync()
	return nil
}
