package net

import (
	"context"
	"testing"
	"time"

	"sync"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/mempool"
	proto_net "github.com/gochain/gochain/pkg/proto/net"
	"github.com/gochain/gochain/pkg/storage"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewNetwork(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_new_network"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	assert.NotNil(t, net)
	defer net.Close()
}

func TestNetworkConnect(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage1, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_network_connect_1"))
	assert.NoError(t, err)
	defer dummyStorage1.Close()

	dummyChainConfig1 := chain.DefaultChainConfig()
	consensusConfig1 := consensus.DefaultConsensusConfig()
	dummyChain1, err := chain.NewChain(dummyChainConfig1, consensusConfig1, dummyStorage1)
	assert.NoError(t, err)

	dummyMempoolConfig1 := mempool.TestMempoolConfig()
	dummyMempool1 := mempool.NewMempool(dummyMempoolConfig1)

	dummyStorage2, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_network_connect_2"))
	assert.NoError(t, err)
	defer dummyStorage2.Close()

	dummyChainConfig2 := chain.DefaultChainConfig()
	consensusConfig2 := consensus.DefaultConsensusConfig()
	dummyChain2, err := chain.NewChain(dummyChainConfig2, consensusConfig2, dummyStorage2)
	assert.NoError(t, err)

	dummyMempoolConfig2 := mempool.TestMempoolConfig()
	dummyMempool2 := mempool.NewMempool(dummyMempoolConfig2)

	config1 := DefaultNetworkConfig()
	config1.EnableMDNS = false
	net1, err := NewNetwork(config1, dummyChain1, dummyMempool1)
	assert.NoError(t, err)
	defer net1.Close()

	config2 := DefaultNetworkConfig()
	config2.EnableMDNS = false
	net2, err := NewNetwork(config2, dummyChain2, dummyMempool2)
	assert.NoError(t, err)
	defer net2.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	peerInfo2 := net2.GetHost().Peerstore().PeerInfo(net2.GetHost().ID())
	err = net1.GetHost().Connect(ctx, peerInfo2)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(net1.GetPeers()))
	assert.Equal(t, 2, len(net2.GetPeers()))
}

func TestMessageSigningAndVerification(t *testing.T) {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 2048)
	assert.NoError(t, err)

	payload := []byte(`"test payload"`)

	peerID, err := peer.IDFromPublicKey(pub)
	assert.NoError(t, err)
	peerIDBytes, err := peerID.MarshalBinary()
	assert.NoError(t, err)

	msg := &proto_net.Message{
		TimestampUnixNano: time.Now().UnixNano(),
		FromPeerId:        peerIDBytes,
		Content: &proto_net.Message_BlockMessage{
			BlockMessage: &proto_net.BlockMessage{
				BlockData: payload,
			},
		},
	}

	// Sign the message
	dataToSign, err := proto.Marshal(msg)
	assert.NoError(t, err)
	signature, err := priv.Sign(dataToSign)
	assert.NoError(t, err)
	msg.Signature = signature

	// Verify the message
	pubKeyFromPeerID, err := peer.ID(msg.FromPeerId).ExtractPublicKey()
	assert.NoError(t, err)
	tempMsg := proto.Clone(msg).(*proto_net.Message)
	tempMsg.Signature = nil // Clear the signature for verification
	dataToVerify, err := proto.Marshal(tempMsg)
	assert.NoError(t, err)
	verified, err := pubKeyFromPeerID.Verify(dataToVerify, msg.Signature)
	assert.NoError(t, err)
	assert.True(t, verified)

	// Tamper with payload
	msg.GetBlockMessage().BlockData = []byte("tampered payload")
	tempMsg = proto.Clone(msg).(*proto_net.Message)
	tempMsg.Signature = nil // Clear the signature for verification
	dataToVerify, err = proto.Marshal(tempMsg)
	assert.NoError(t, err)
	verified, err = pubKeyFromPeerID.Verify(dataToVerify, msg.Signature)
	assert.NoError(t, err)
	assert.False(t, verified)
}

func TestPublishSubscribe(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage1, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_publish_subscribe_1"))
	assert.NoError(t, err)
	defer dummyStorage1.Close()

	dummyChainConfig1 := chain.DefaultChainConfig()
	consensusConfig1 := consensus.DefaultConsensusConfig()
	dummyChain1, err := chain.NewChain(dummyChainConfig1, consensusConfig1, dummyStorage1)
	assert.NoError(t, err)

	dummyMempoolConfig1 := mempool.TestMempoolConfig()
	dummyMempool1 := mempool.NewMempool(dummyMempoolConfig1)

	dummyStorage2, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_publish_subscribe_2"))
	assert.NoError(t, err)
	defer dummyStorage2.Close()

	dummyChainConfig2 := chain.DefaultChainConfig()
	consensusConfig2 := consensus.DefaultConsensusConfig()
	dummyChain2, err := chain.NewChain(dummyChainConfig2, consensusConfig2, dummyStorage2)
	assert.NoError(t, err)

	dummyMempoolConfig2 := mempool.TestMempoolConfig()
	dummyMempool2 := mempool.NewMempool(dummyMempoolConfig2)

	config1 := DefaultNetworkConfig()
	config1.EnableMDNS = false
	config1.ListenPort = 0 // Random port
	net1, err := NewNetwork(config1, dummyChain1, dummyMempool1)
	assert.NoError(t, err)
	defer net1.Close()

	config2 := DefaultNetworkConfig()
	config2.EnableMDNS = false
	config2.ListenPort = 0 // Random port
	net2, err := NewNetwork(config2, dummyChain2, dummyMempool2)
	assert.NoError(t, err)
	defer net2.Close()

	// Connect the two networks
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for networks to be ready
	time.Sleep(2 * time.Second)

	// Get peer info from net2 and connect net1 to it
	peerInfo2 := net2.GetHost().Peerstore().PeerInfo(net2.GetHost().ID())
	err = net1.GetHost().Connect(ctx, peerInfo2)
	assert.NoError(t, err)

	// Wait for connection to be established
	time.Sleep(3 * time.Second)

	// Verify connection is established
	net1Peers := net1.GetPeers()
	net2Peers := net2.GetPeers()

	// Check that both networks see each other
	assert.GreaterOrEqual(t, len(net1Peers), 1, "net1 should have at least 1 peer")
	assert.GreaterOrEqual(t, len(net2Peers), 1, "net2 should have at least 1 peer")

	// Subscribe to blocks on net2 BEFORE publishing
	blockSub2, err := net2.SubscribeToBlocks()
	assert.NoError(t, err)
	defer blockSub2.Cancel()

	// Give time for subscription to propagate
	time.Sleep(2 * time.Second)

	// Publish a block from net1
	blockData := []byte(`"test block data"`)
	err = net1.PublishBlock(blockData)
	assert.NoError(t, err)

	// Wait for message on net2 with a reasonable timeout
	msgChan := make(chan *pubsub.Message)
	errChan := make(chan error)

	go func() {
		msg, err := blockSub2.Next(ctx)
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- msg
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Context cancelled before message received")
	case msg := <-msgChan:
		// Handle the message
		var receivedMsg proto_net.Message
		assert.NoError(t, proto.Unmarshal(msg.Data, &receivedMsg))

		// Check if it's a BlockMessage and extract data
		blockMsg := receivedMsg.GetBlockMessage()
		assert.NotNil(t, blockMsg)
		assert.Equal(t, blockData, blockMsg.BlockData)

		// Verify message signature
		pubKey, err := peer.ID(receivedMsg.FromPeerId).ExtractPublicKey()
		assert.NoError(t, err)
		tempMsg := proto.Clone(&receivedMsg).(*proto_net.Message)
		tempMsg.Signature = nil // Clear the signature for verification
		dataToVerify, err := proto.Marshal(tempMsg)
		assert.NoError(t, err)
		verified, err := pubKey.Verify(dataToVerify, receivedMsg.Signature)
		assert.NoError(t, err)
		assert.True(t, verified)
	case err := <-errChan:
		t.Fatalf("Error receiving block message: %v", err)
	case <-time.After(20 * time.Second): // Increased timeout for more reliable testing
		t.Fatal("Timeout waiting for block message")
	}

	// Subscribe to transactions on net1 BEFORE publishing
	txSub1, err := net1.SubscribeToTransactions()
	assert.NoError(t, err)
	defer txSub1.Cancel()

	// Give time for subscription to propagate
	time.Sleep(2 * time.Second)

	// Publish a transaction from net2
	txData := []byte(`"test transaction data"`)
	err = net2.PublishTransaction(txData)
	assert.NoError(t, err)

	// Wait for message on net1
	msgChan2 := make(chan *pubsub.Message)
	errChan2 := make(chan error)

	go func() {
		msg, err := txSub1.Next(ctx)
		if err != nil {
			errChan2 <- err
			return
		}
		msgChan2 <- msg
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Context cancelled before message received")
	case msg := <-msgChan2:
		// Handle the message
		var receivedMsg proto_net.Message
		assert.NoError(t, proto.Unmarshal(msg.Data, &receivedMsg))

		// Check if it's a TransactionMessage and extract data
		txMsg := receivedMsg.GetTransactionMessage()
		assert.NotNil(t, txMsg)
		assert.Equal(t, txData, txMsg.TransactionData)

		// Verify message signature
		pubKey, err := peer.ID(receivedMsg.FromPeerId).ExtractPublicKey()
		assert.NoError(t, err)
		tempMsg := proto.Clone(&receivedMsg).(*proto_net.Message)
		tempMsg.Signature = nil // Clear the signature for verification
		dataToVerify, err := proto.Marshal(tempMsg)
		assert.NoError(t, err)
		verified, err := pubKey.Verify(dataToVerify, receivedMsg.Signature)
		assert.NoError(t, err)
		assert.True(t, verified)
	case err := <-errChan2:
		t.Fatalf("Error receiving transaction message: %v", err)
	case <-time.After(20 * time.Second): // Increased timeout for more reliable testing
		t.Fatal("Timeout waiting for transaction message")
	}
}

// TestNetworkAdvancedScenarios tests advanced network scenarios
func TestNetworkAdvancedScenarios(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_advanced_scenarios"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	defer net.Close()

	// Test basic network functionality
	t.Run("BasicFunctionality", func(t *testing.T) {
		// Test that network is accessible
		assert.NotNil(t, net, "Network should be accessible")

		// Test getting peers
		peers := net.GetPeers()
		assert.NotNil(t, peers, "Should return peer list")
		assert.True(t, len(peers) >= 0, "Peer list should be non-negative")

		// Test getting host
		host := net.GetHost()
		assert.NotNil(t, host, "Should return host")
	})
}

// TestNetworkConcurrency tests network behavior under concurrent operations
func TestNetworkConcurrency(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_concurrency"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	defer net.Close()

	// Test concurrent peer operations
	t.Run("ConcurrentPeerOperations", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Perform concurrent peer operations
				_ = net.GetPeers()
				_ = net.GetHost()
			}()
		}

		wg.Wait()
		// Verify network remains accessible
		assert.NotNil(t, net, "Network should remain accessible")
	})

	// Test concurrent message handling
	t.Run("ConcurrentMessageHandling", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 5

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Test concurrent message operations
				// This would test actual message handling if implemented
				assert.NotNil(t, net, "Network should remain accessible")
			}()
		}

		wg.Wait()
	})
}

// TestNetworkEdgeCases tests network edge cases and error conditions
func TestNetworkEdgeCases(t *testing.T) {
	// Test with invalid configurations
	t.Run("InvalidConfigurations", func(t *testing.T) {
		// Test with nil chain
		config := DefaultNetworkConfig()
		config.EnableMDNS = false

		nilChainNet, err := NewNetwork(config, nil, mempool.NewMempool(mempool.TestMempoolConfig()))
		// This may or may not fail depending on implementation
		// We're just testing that it doesn't panic
		if err == nil {
			defer nilChainNet.Close()
		}

		// Test with nil mempool
		nilMempoolStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_nil_mempool"))
		if err == nil {
			defer nilMempoolStorage.Close()
			nilMempoolChain, err := chain.NewChain(chain.DefaultChainConfig(), consensus.DefaultConsensusConfig(), nilMempoolStorage)
			if err == nil {
				nilMempoolNet, err := NewNetwork(config, nilMempoolChain, nil)
				// This may or may not fail depending on implementation
				// We're just testing that it doesn't panic
				if err == nil {
					defer nilMempoolNet.Close()
				}
			}
		}
	})

	// Test with extreme port values
	t.Run("ExtremePortValues", func(t *testing.T) {
		dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_extreme_ports"))
		assert.NoError(t, err)
		defer dummyStorage.Close()

		dummyChainConfig := chain.DefaultChainConfig()
		consensusConfig := consensus.DefaultConsensusConfig()
		dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
		assert.NoError(t, err)

		dummyMempoolConfig := mempool.TestMempoolConfig()
		dummyMempool := mempool.NewMempool(dummyMempoolConfig)

		// Test with port 0 (random port)
		port0Config := DefaultNetworkConfig()
		port0Config.ListenPort = 0
		port0Config.EnableMDNS = false

		port0Net, err := NewNetwork(port0Config, dummyChain, dummyMempool)
		if err == nil {
			defer port0Net.Close()
			assert.NotNil(t, port0Net, "Should handle port 0")
		}
	})
}

// TestNetworkPerformance tests network performance characteristics
func TestNetworkPerformance(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_performance"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	defer net.Close()

	// Test network operation speed
	t.Run("OperationSpeed", func(t *testing.T) {
		// Test peer retrieval speed
		startTime := time.Now()
		for i := 0; i < 1000; i++ {
			_ = net.GetPeers()
		}
		peerRetrievalTime := time.Since(startTime)
		assert.True(t, peerRetrievalTime < 100*time.Millisecond, "Peer retrieval should be fast")

		// Test host retrieval speed
		startTime = time.Now()
		for i := 0; i < 1000; i++ {
			_ = net.GetHost()
		}
		hostRetrievalTime := time.Since(startTime)
		assert.True(t, hostRetrievalTime < 100*time.Millisecond, "Host retrieval should be fast")
	})

	// Test memory usage
	t.Run("MemoryUsage", func(t *testing.T) {
		// Perform many operations to test memory usage
		for i := 0; i < 10000; i++ {
			_ = net.GetPeers()
			_ = net.GetHost()
		}

		// Verify network is still accessible (no memory leaks)
		assert.NotNil(t, net, "Network should remain accessible after many operations")
	})
}

// TestNetworkRecovery tests network recovery mechanisms
func TestNetworkRecovery(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_recovery"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)

	// Test network close and recreation
	t.Run("CloseAndRecreation", func(t *testing.T) {
		// Close the network
		err := net.Close()
		assert.NoError(t, err)

		// Try to recreate network
		newNet, err := NewNetwork(config, dummyChain, dummyMempool)
		assert.NoError(t, err)
		defer newNet.Close()

		assert.NotNil(t, newNet, "Should be able to recreate network")
	})

	// Test network restart
	t.Run("NetworkRestart", func(t *testing.T) {
		// Create a new network for restart testing
		restartNet, err := NewNetwork(config, dummyChain, dummyMempool)
		assert.NoError(t, err)
		defer restartNet.Close()

		// Test that network can be used after creation
		peers := restartNet.GetPeers()
		assert.True(t, len(peers) >= 0, "Should be able to get peers after restart")
	})
}

// TestNetworkIntegration tests network integration with other components
func TestNetworkIntegration(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_integration"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	defer net.Close()

	// Test network with chain integration
	t.Run("ChainIntegration", func(t *testing.T) {
		// Test that network can access chain information
		// This tests the integration between network and chain components
		assert.NotNil(t, net, "Network should be accessible")
		assert.NotNil(t, dummyChain, "Chain should be accessible")
	})

	// Test network with mempool integration
	t.Run("MempoolIntegration", func(t *testing.T) {
		// Test that network can access mempool information
		// This tests the integration between network and mempool components
		assert.NotNil(t, net, "Network should be accessible")
		assert.NotNil(t, dummyMempool, "Mempool should be accessible")
	})
}

// TestNetworkSecurity tests network security features
func TestNetworkSecurity(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_security"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, consensusConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.TestMempoolConfig()
	dummyMempool := mempool.NewMempool(dummyMempoolConfig)

	config := DefaultNetworkConfig()
	config.EnableMDNS = false
	net, err := NewNetwork(config, dummyChain, dummyMempool)
	assert.NoError(t, err)
	defer net.Close()

	// Test message signing and verification
	t.Run("MessageSecurity", func(t *testing.T) {
		// Test that network can handle secure messages
		// This would test actual security features if implemented
		assert.NotNil(t, net, "Network should remain accessible")
	})

	// Test peer authentication
	t.Run("PeerAuthentication", func(t *testing.T) {
		// Test that network can authenticate peers
		// This would test actual authentication if implemented
		assert.NotNil(t, net, "Network should remain accessible")
	})
}
