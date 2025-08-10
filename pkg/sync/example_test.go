package sync

import (
	"testing"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/assert"
)

// Example test that shows how to use the sync package
func ExampleSyncManager_BasicUsage(t *testing.T) {
	// Create a test host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.DisableRelay(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()

	// Create storage
	storageFactory := storage.NewStorageFactory()
	nodeStorage, err := storageFactory.CreateStorage(storage.StorageTypeFile, "./test_data")
	if err != nil {
		t.Fatal(err)
	}

	// Create chain
	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()

	blockchain, err := chain.NewChain(chainConfig, consensusConfig, nodeStorage)
	if err != nil {
		t.Fatal(err)
	}

	// Create chain adapter to implement the required interfaces
	chainAdapter := NewChainAdapter(blockchain)

	// Create sync manager
	config := DefaultSyncConfig()
	syncManager := NewSyncManager(chainAdapter, chainAdapter, nodeStorage, config, host)

	// Verify sync manager was created
	assert.NotNil(t, syncManager)
	assert.Equal(t, chainAdapter, syncManager.chain)
	assert.Equal(t, nodeStorage, syncManager.storage)
	assert.Equal(t, config, syncManager.config)
	assert.NotNil(t, syncManager.peers)
	assert.False(t, syncManager.status.IsSyncing)

	// Test basic operations
	err = syncManager.StartSync()
	assert.NoError(t, err)
	assert.True(t, syncManager.status.IsSyncing)

	syncManager.StopSync()
	assert.False(t, syncManager.status.IsSyncing)

	// Clean up
	syncManager.Close()
	blockchain.Close()
	nodeStorage.Close()
}
