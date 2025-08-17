package sync

import (
	"fmt"

	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/libp2p/go-libp2p"
)

// Example that shows how to use the sync package
func ExampleSyncManager() {
	// Create a test host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.DisableRelay(),
	)
	if err != nil {
		fmt.Printf("Failed to create host: %v\n", err)
		return
	}
	defer host.Close()

	// Create storage
	storageFactory := storage.NewStorageFactory()
	nodeStorage, err := storageFactory.CreateStorage(storage.StorageTypeFile, "./test_data")
	if err != nil {
		fmt.Printf("Failed to create storage: %v\n", err)
		return
	}

	// Create chain
	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()

	blockchain, err := chain.NewChain(chainConfig, consensusConfig, nodeStorage)
	if err != nil {
		fmt.Printf("Failed to create chain: %v\n", err)
		return
	}

	// Create chain adapter to implement the required interfaces
	chainAdapter := NewChainAdapter(blockchain)

	// Create sync manager
	config := DefaultSyncConfig()
	syncManager := NewSyncManager(chainAdapter, chainAdapter, nodeStorage, config, host)

	// Verify sync manager was created
	if syncManager == nil {
		fmt.Println("Failed to create sync manager")
		return
	}

	// Demonstrate basic operations
	err = syncManager.StartSync()
	if err != nil {
		fmt.Printf("Failed to start sync: %v\n", err)
		return
	}
	fmt.Println("Sync started successfully")

	syncManager.StopSync()
	fmt.Println("Sync stopped successfully")

	// Clean up
	syncManager.Close()
	blockchain.Close()
	nodeStorage.Close()
}
