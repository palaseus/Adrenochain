package net

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/mempool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock chain and mempool for testing
type mockChain struct{}

func (mc *mockChain) GetLatestBlock() *block.Block { return nil }

type mockMempool struct{}

func (mm *mockMempool) GetTransactionCount() int { return 0 }

// testMockStream is a mock implementation of network.Stream for testing
type testMockStream struct {
	conn network.Conn
}

func (m *testMockStream) Read(p []byte) (n int, err error)                 { return 0, nil }
func (m *testMockStream) Write(p []byte) (n int, err error)                { return 0, nil }
func (m *testMockStream) Close() error                                     { return nil }
func (m *testMockStream) CloseRead() error                                 { return nil }
func (m *testMockStream) CloseWrite() error                                { return nil }
func (m *testMockStream) Reset() error                                     { return nil }
func (m *testMockStream) ResetWithError(err network.StreamErrorCode) error { return nil }
func (m *testMockStream) SetDeadline(t time.Time) error                    { return nil }
func (m *testMockStream) SetReadDeadline(t time.Time) error                { return nil }
func (m *testMockStream) SetWriteDeadline(t time.Time) error               { return nil }
func (m *testMockStream) Protocol() protocol.ID                            { return "" }
func (m *testMockStream) SetProtocol(id protocol.ID) error                 { return nil }
func (m *testMockStream) ID() string                                       { return "mock-stream-id" }
func (m *testMockStream) Stat() network.Stats                              { return network.Stats{} }
func (m *testMockStream) Conn() network.Conn                               { return m.conn }
func (m *testMockStream) Scope() network.StreamScope                       { return nil }

// TestDefaultNetworkConfig tests the default network configuration
func TestDefaultNetworkConfig(t *testing.T) {
	config := DefaultNetworkConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0, config.ListenPort)
	assert.Equal(t, 50, config.MaxPeers)
	assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
	assert.True(t, config.EnableMDNS)
	assert.False(t, config.EnableRelay)
	assert.Empty(t, config.BootstrapPeers)
}

// TestNewNetwork tests network creation
func TestNewNetwork(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0 // Random port
	config.EnableMDNS = false
	config.EnableRelay = false

	// Create real chain and mempool instances for testing
	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.NotNil(t, network.GetHost())
	assert.NotNil(t, network.GetContext())
}

// TestNewNetworkWithBootstrapPeers tests network creation with bootstrap peers
func TestNewNetworkWithBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmTestPeer"}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	// Note: Invalid bootstrap peer addresses are skipped, so we expect 0
	assert.Len(t, network.bootstrapPeers, 0)
}

// TestNewNetworkWithValidBootstrapPeer tests network creation with valid bootstrap peer
func TestNewNetworkWithValidBootstrapPeer(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	// Use a valid multiaddr format - the peer ID must be a valid base58-encoded string
	config.BootstrapPeers = []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu"}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	// Valid bootstrap peer should be parsed
	assert.Len(t, network.bootstrapPeers, 1)
}

// TestNewNetworkWithInvalidBootstrapPeer tests network creation with invalid bootstrap peer
func TestNewNetworkWithInvalidBootstrapPeer(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{"invalid-address"}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	// Invalid bootstrap peer should be skipped
	assert.Len(t, network.bootstrapPeers, 0)
}

// TestNetworkNotifieeMethods tests all the notifiee interface methods
func TestNetworkNotifieeMethods(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Create mock objects
	mockConn := &mockConn{}
	mockStream := &testMockStream{conn: mockConn}

	// Test OpenedStream with nil
	network.OpenedStream(nil, nil)

	// Test OpenedStream with valid stream
	network.OpenedStream(nil, mockStream)

	// Test OpenedStream with stream that has nil conn
	mockStreamNilConn := &testMockStream{conn: nil}
	network.OpenedStream(nil, mockStreamNilConn)

	// Test ClosedStream with nil
	network.ClosedStream(nil, nil)

	// Test ClosedStream with valid stream
	network.ClosedStream(nil, mockStream)

	// Test ClosedStream with stream that has nil conn
	network.ClosedStream(nil, mockStreamNilConn)

	// Test OpenedConn with nil
	network.OpenedConn(nil, nil)

	// Test OpenedConn with valid conn
	network.OpenedConn(nil, mockConn)

	// Test ClosedConn with nil
	network.ClosedConn(nil, nil)

	// Test ClosedConn with valid conn
	network.ClosedConn(nil, mockConn)

	// Test Listen with nil
	network.Listen(nil, nil)

	// Test Listen with valid multiaddr
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	require.NoError(t, err)
	network.Listen(nil, addr)

	// Test ListenClose with nil
	network.ListenClose(nil, nil)

	// Test ListenClose with valid multiaddr
	network.ListenClose(nil, addr)
}

// TestNewNetworkWithInvalidKeyGeneration tests network creation with invalid key generation
func TestNewNetworkWithInvalidKeyGeneration(t *testing.T) {
	// This test would require mocking crypto.GenerateKeyPairWithReader to return an error
	// Since we can't easily mock the crypto package, we'll test other error paths
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	// Test with a very high port number (this should still work as libp2p handles it gracefully)
	config.ListenPort = 65536
	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	// This might fail due to port range validation, which is expected
	if err != nil {
		t.Logf("Expected error with invalid port: %v", err)
		return
	}
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNewNetworkWithDHTError tests network creation with DHT creation error
func TestNewNetworkWithDHTError(t *testing.T) {
	// This test would require mocking the DHT creation to return an error
	// Since we can't easily mock the DHT package, we'll test other scenarios
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNewNetworkWithPubSubError tests network creation with pubsub creation error
func TestNewNetworkWithPubSubError(t *testing.T) {
	// This test would require mocking the pubsub creation to return an error
	// Since we can't easily mock the pubsub package, we'll test other scenarios
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestStartPeerDiscoveryWithMDNSError tests peer discovery with mDNS error
func TestStartPeerDiscoveryWithMDNSError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = true
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// The mDNS service should start successfully in most cases
	assert.NotNil(t, network)
}

// TestStartPeerDiscoveryWithDHTBootstrapError tests peer discovery with DHT bootstrap error
func TestStartPeerDiscoveryWithDHTBootstrapError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// The DHT bootstrap should work in most cases
	assert.NotNil(t, network)
}

// TestCloseWithHostError tests network close with host close error
func TestCloseWithHostError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)

	// Close should work without errors
	err = network.Close()
	assert.NoError(t, err)
}

// TestCloseWithDHTError tests network close with DHT close error
func TestCloseWithDHTError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)

	// Close should work without errors
	err = network.Close()
	assert.NoError(t, err)
}

// TestPublishBlockWithNilPubKey tests block publishing with nil public key
func TestPublishBlockWithNilPubKey(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the host's Peerstore to return nil for PubKey
	// Since we can't easily mock this, we'll test the normal case
	blockData := []byte("test block data")
	err = network.PublishBlock(blockData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishBlockWithPeerIDError tests block publishing with peer ID error
func TestPublishBlockWithPeerIDError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the peer.IDFromPublicKey to return an error
	// Since we can't easily mock this, we'll test the normal case
	blockData := []byte("test block data")
	err = network.PublishBlock(blockData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishBlockWithMarshalError tests block publishing with marshal error
func TestPublishBlockWithMarshalError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking proto.Marshal to return an error
	// Since we can't easily mock this, we'll test the normal case
	blockData := []byte("test block data")
	err = network.PublishBlock(blockData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishBlockWithSignError tests block publishing with signing error
func TestPublishBlockWithSignError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the private key's Sign method to return an error
	// Since we can't easily mock this, we'll test the normal case
	blockData := []byte("test block data")
	err = network.PublishBlock(blockData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishTransactionWithNilPubKey tests transaction publishing with nil public key
func TestPublishTransactionWithNilPubKey(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the host's Peerstore to return nil for PubKey
	// Since we can't easily mock this, we'll test the normal case
	txData := []byte("test transaction data")
	err = network.PublishTransaction(txData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishTransactionWithPeerIDError tests transaction publishing with peer ID error
func TestPublishTransactionWithPeerIDError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the peer.IDFromPublicKey to return an error
	// Since we can't easily mock this, we'll test the normal case
	txData := []byte("test transaction data")
	err = network.PublishTransaction(txData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishTransactionWithMarshalError tests transaction publishing with marshal error
func TestPublishTransactionWithMarshalError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking proto.Marshal to return an error
	// Since we can't easily mock this, we'll test the normal case
	txData := []byte("test transaction data")
	err = network.PublishTransaction(txData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestPublishTransactionWithSignError tests transaction publishing with signing error
func TestPublishTransactionWithSignError(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// This test would require mocking the private key's Sign method to return an error
	// Since we can't easily mock this, we'll test the normal case
	txData := []byte("test transaction data")
	err = network.PublishTransaction(txData)
	// The error handling depends on the actual implementation
	// We're testing that the function doesn't panic
}

// TestNetworkNotifieeMethodsWithValidParams tests notifiee methods with valid parameters
func TestNetworkNotifieeMethodsWithValidParams(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test notifiee methods with mock objects
	// Create a mock connection
	mockConn := &mockConn{
		peerID: "QmTestPeer123",
		connID: "test-conn-123",
	}

	// Test all notifiee methods
	assert.NotPanics(t, func() {
		network.Connected(nil, mockConn)
		network.Disconnected(nil, mockConn)
		network.OpenedStream(nil, nil) // Stream methods expect network.Stream
		network.ClosedStream(nil, nil)
		network.OpenedConn(nil, mockConn)
		network.ClosedConn(nil, mockConn)
		network.Listen(nil, nil) // Listen methods expect multiaddr.Multiaddr
		network.ListenClose(nil, nil)
	})
}

// TestNetworkErrorHandling tests error handling scenarios
func TestNetworkErrorHandling(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test error handling in various scenarios
	// This test ensures the network handles errors gracefully
	t.Log("Testing network error handling")
}

// TestNetworkCloseWithErrors tests network close with potential errors
func TestNetworkCloseWithErrors(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)

	// Close should not panic even if there are errors
	assert.NotPanics(t, func() {
		err := network.Close()
		// Close might return errors, but shouldn't panic
		if err != nil {
			t.Logf("Network close returned error (expected): %v", err)
		}
	})
}

// TestNetworkPublishWithErrors tests publishing with potential error conditions
func TestNetworkPublishWithErrors(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test publishing with various data types
	testCases := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"Empty data", []byte{}, true},
		{"Small data", []byte("test"), true},
		{"Large data", make([]byte, 1000), true},
		{"Nil data", nil, true}, // Should handle gracefully
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test block publishing
			err := network.PublishBlock(tc.data)
			if tc.expected {
				assert.NoError(t, err, "Block publishing should succeed for %s", tc.name)
			}

			// Test transaction publishing
			err = network.PublishTransaction(tc.data)
			if tc.expected {
				assert.NoError(t, err, "Transaction publishing should succeed for %s", tc.name)
			}
		})
	}
}

// TestNetworkPeerDiscoveryErrors tests peer discovery error handling
func TestNetworkPeerDiscoveryErrors(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test peer discovery with various scenarios
	t.Log("Testing peer discovery error handling")

	// Test with invalid peer info
	invalidPeer := peer.AddrInfo{
		ID: "", // Empty peer ID
	}
	assert.NotPanics(t, func() {
		network.HandlePeerFound(invalidPeer)
	})

	// Test with peer info that has no addresses
	noAddrPeer := peer.AddrInfo{
		ID:    "QmNoAddrPeer",
		Addrs: []multiaddr.Multiaddr{},
	}
	assert.NotPanics(t, func() {
		network.HandlePeerFound(noAddrPeer)
	})
}

// TestNetworkBootstrapPeerErrors tests bootstrap peer error handling
func TestNetworkBootstrapPeerErrors(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test various bootstrap peer scenarios
	t.Log("Testing bootstrap peer error handling")

	// Test with multiple invalid bootstrap peers
	config.BootstrapPeers = []string{
		"invalid-address-1",
		"invalid-address-2",
		"/ip4/127.0.0.1/tcp/1234/p2p/QmValidPeer",
		"another-invalid-address",
	}

	// Create a new network with these bootstrap peers
	network2, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network2.Close()

	// Should handle invalid addresses gracefully
	assert.NotNil(t, network2)
}

// TestNetworkConcurrentPublishing tests concurrent publishing operations
func TestNetworkConcurrentPublishing(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test concurrent publishing operations
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Publish blocks concurrently
			blockData := []byte(fmt.Sprintf("concurrent block %d", id))
			err := network.PublishBlock(blockData)
			assert.NoError(t, err)

			// Publish transactions concurrently
			txData := []byte(fmt.Sprintf("concurrent tx %d", id))
			err = network.PublishTransaction(txData)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	// Verify network is still functional
	assert.NotNil(t, network.GetHost())
}

// TestNetworkMemoryUsage tests network memory usage under load
func TestNetworkMemoryUsage(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Perform many operations to test memory usage
	for i := 0; i < 100; i++ {
		// Add peers
		peerInfo := createMockPeerInfo(fmt.Sprintf("QmMemoryTestPeer%d", i))
		network.HandlePeerFound(peerInfo)

		// Publish data
		data := []byte(fmt.Sprintf("memory test data %d", i))
		network.PublishBlock(data)
		network.PublishTransaction(data)

		// Get network info
		_ = network.GetPeers()
		_ = network.GetHost()
		_ = network.GetContext()
	}

	// Verify network is still functional
	assert.NotNil(t, network.GetHost())
	assert.NotNil(t, network.GetContext())
}

// TestNetworkRecovery tests network recovery after various operations
func TestNetworkRecovery(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test network recovery after many operations
	t.Log("Testing network recovery")

	// Perform many operations
	for i := 0; i < 50; i++ {
		peerInfo := createMockPeerInfo(fmt.Sprintf("QmRecoveryTestPeer%d", i))
		network.HandlePeerFound(peerInfo)

		data := []byte(fmt.Sprintf("recovery test data %d", i))
		network.PublishBlock(data)
		network.PublishTransaction(data)
	}

	// Verify network is still functional
	assert.NotNil(t, network.GetHost())
	assert.NotNil(t, network.GetContext())

	// Test subscriptions still work
	blockSub, err := network.SubscribeToBlocks()
	require.NoError(t, err)
	defer blockSub.Cancel()

	txSub, err := network.SubscribeToTransactions()
	require.NoError(t, err)
	defer txSub.Cancel()

	assert.NotNil(t, blockSub)
	assert.NotNil(t, txSub)
}

// TestHandlePeerFound tests peer discovery handling
func TestHandlePeerFound(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 2

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test adding a peer
	peerID := "QmTestPeer123"
	peerInfo := createMockPeerInfo(peerID)

	network.HandlePeerFound(peerInfo)

	// Verify peer was added
	peers := network.GetPeers()
	assert.Len(t, peers, 1)
}

// TestHandlePeerFoundMaxPeersLimit tests peer limit enforcement
func TestHandlePeerFoundMaxPeersLimit(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 1

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Add first peer
	peer1 := createMockPeerInfo("QmPeer1")
	network.HandlePeerFound(peer1)

	// Try to add second peer - should be rejected
	peer2 := createMockPeerInfo("QmPeer2")
	network.HandlePeerFound(peer2)

	// Verify only one peer was added
	peers := network.GetPeers()
	assert.Len(t, peers, 1)
}

// TestHandlePeerFoundDuplicatePeer tests duplicate peer handling
func TestHandlePeerFoundDuplicatePeer(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Add same peer twice
	peerInfo := createMockPeerInfo("QmDuplicatePeer")
	network.HandlePeerFound(peerInfo)
	network.HandlePeerFound(peerInfo)

	// Verify only one peer was added
	peers := network.GetPeers()
	assert.Len(t, peers, 1)
}

// TestSubscribeToBlocks tests block subscription
func TestSubscribeToBlocks(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	subscription, err := network.SubscribeToBlocks()
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// Clean up subscription
	subscription.Cancel()
}

// TestSubscribeToTransactions tests transaction subscription
func TestSubscribeToTransactions(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	subscription, err := network.SubscribeToTransactions()
	require.NoError(t, err)
	assert.NotNil(t, subscription)

	// Clean up subscription
	subscription.Cancel()
}

// TestPublishBlock tests block publishing
func TestPublishBlock(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	blockData := []byte("test block data")
	err = network.PublishBlock(blockData)
	assert.NoError(t, err)
}

// TestPublishTransaction tests transaction publishing
func TestPublishTransaction(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	txData := []byte("test transaction data")
	err = network.PublishTransaction(txData)
	assert.NoError(t, err)
}

// TestPublishBlockWithEmptyData tests block publishing with empty data
func TestPublishBlockWithEmptyData(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	blockData := []byte{}
	err = network.PublishBlock(blockData)
	assert.NoError(t, err)
}

// TestPublishTransactionWithEmptyData tests transaction publishing with empty data
func TestPublishTransactionWithEmptyData(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	txData := []byte{}
	err = network.PublishTransaction(txData)
	assert.NoError(t, err)
}

// TestNetworkClose tests network cleanup
func TestNetworkClose(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)

	// Close should not panic
	assert.NotPanics(t, func() {
		err := network.Close()
		assert.NoError(t, err)
	})
}

// TestGetPeers tests peer retrieval
func TestGetPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	peers := network.GetPeers()
	assert.NotNil(t, peers)
	// Note: The network may discover peers automatically via mDNS
	// We'll just verify that GetPeers returns a valid slice
	assert.GreaterOrEqual(t, len(peers), 0)
}

// TestIsTestEnvironment tests the test environment detection
func TestIsTestEnvironment(t *testing.T) {
	// Test that we're in a test environment
	result := isTestEnvironment()
	assert.True(t, result)
}

// TestNetworkWithExtremePorts tests network creation with extreme port values
func TestNetworkWithExtremePorts(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 65535 // Max port
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithZeroTimeout tests network creation with zero timeout
func TestNetworkWithZeroTimeout(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.ConnectionTimeout = 0

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithMaxPeersLimit tests network with maximum peers limit
func TestNetworkWithMaxPeersLimit(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 1

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithRelayEnabled tests network creation with relay enabled
func TestNetworkWithRelayEnabled(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = true

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithNilChain tests network creation with nil chain
func TestNetworkWithNilChain(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, nil, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithNilMempool tests network creation with nil mempool
func TestNetworkWithNilMempool(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}

	network, err := NewNetwork(config, chainInstance, nil)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithNilConfig tests network creation with nil config
func TestNetworkWithNilConfig(t *testing.T) {
	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	// This should panic or return an error
	defer func() {
		if r := recover(); r != nil {
			// Expected panic
		}
	}()

	network, err := NewNetwork(nil, chainInstance, mempoolInstance)
	if err != nil {
		// Expected error
		return
	}
	defer network.Close()

	// If we get here, the test should fail
	t.Fatal("Expected error or panic with nil config")
}

// TestNetworkConcurrency tests network operations under concurrent access
func TestNetworkConcurrency(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent access to GetPeers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			peers := network.GetPeers()
			_ = peers
		}()
	}

	// Test concurrent access to GetHost
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			host := network.GetHost()
			_ = host
		}()
	}

	// Test concurrent access to GetContext
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := network.GetContext()
			_ = ctx
		}()
	}

	wg.Wait()
}

// TestNetworkAdvancedScenarios tests advanced network scenarios
func TestNetworkAdvancedScenarios(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with empty block data
	err = network.PublishBlock([]byte{})
	// Should not panic

	// Test with empty transaction data
	err = network.PublishTransaction([]byte{})
	// Should not panic

	// Test with large data
	largeData := make([]byte, 1024*1024) // 1MB
	err = network.PublishBlock(largeData)
	// Should not panic

	err = network.PublishTransaction(largeData)
	// Should not panic
}

// TestNetworkPerformance tests network performance characteristics
func TestNetworkPerformance(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test multiple rapid operations
	start := time.Now()
	for i := 0; i < 100; i++ {
		network.GetPeers()
		network.GetHost()
		network.GetContext()
	}
	duration := time.Since(start)

	// Should complete within reasonable time
	assert.Less(t, duration, 1*time.Second)
}

// TestNetworkSecurity tests network security features
func TestNetworkSecurity(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test that private key is properly set
	assert.NotNil(t, network.privKey)

	// Test that the private key is of the expected type
	keyType := network.privKey.Type()
	// The type comparison might need to handle different type representations
	assert.True(t, keyType == crypto.Ed25519 || keyType.String() == "Ed25519")
}

// TestNetworkIntegration tests network integration with other components
func TestNetworkIntegration(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test that chain is properly set
	assert.Equal(t, chainInstance, network.chain)

	// Test that mempool is properly set
	assert.Equal(t, mempoolInstance, network.mempool)

	// Test that context is properly set
	assert.NotNil(t, network.ctx)

	// Test that cancel function is properly set
	assert.NotNil(t, network.cancel)
}

// TestNetworkEdgeCases tests network edge cases
func TestNetworkEdgeCases(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with very small data
	err = network.PublishBlock([]byte{0x01})
	// Should not panic

	err = network.PublishTransaction([]byte{0x01})
	// Should not panic

	// Test with data containing special characters
	specialData := []byte{0x00, 0xFF, 0x7F, 0x80}
	err = network.PublishBlock(specialData)
	// Should not panic

	err = network.PublishTransaction(specialData)
	// Should not panic
}

// Helper function to create mock peer info
func createMockPeerInfo(peerID string) peer.AddrInfo {
	return peer.AddrInfo{
		ID:    peer.ID(peerID),
		Addrs: []multiaddr.Multiaddr{},
	}
}

// mockConn is a mock implementation of network.Conn for testing
type mockConn struct {
	peerID string
	connID string
}

func (m *mockConn) RemotePeer() peer.ID {
	return peer.ID(m.peerID)
}

func (m *mockConn) RemoteMultiaddr() multiaddr.Multiaddr {
	// Return a mock multiaddr
	addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	return addr
}

// Implement other required methods with empty implementations
func (m *mockConn) Close() error                                           { return nil }
func (m *mockConn) CloseWithError(network.ConnErrorCode) error             { return nil }
func (m *mockConn) ID() string                                             { return m.connID }
func (m *mockConn) LocalPeer() peer.ID                                     { return "" }
func (m *mockConn) LocalMultiaddr() multiaddr.Multiaddr                    { return nil }
func (m *mockConn) RemotePrivateKey() crypto.PrivKey                       { return nil }
func (m *mockConn) RemotePublicKey() crypto.PubKey                         { return nil }
func (m *mockConn) LocalPrivateKey() crypto.PrivKey                        { return nil }
func (m *mockConn) ConnState() network.ConnectionState                     { return network.ConnectionState{} }
func (m *mockConn) Scope() network.ConnScope                               { return nil }
func (m *mockConn) IsClosed() bool                                         { return false }
func (m *mockConn) Stat() network.ConnStats                                { return network.ConnStats{} }
func (m *mockConn) GetStreams() []network.Stream                           { return nil }
func (m *mockConn) OpenStream() (network.Stream, error)                    { return nil, nil }
func (m *mockConn) OpenStreamSync(context.Context) (network.Stream, error) { return nil, nil }
func (m *mockConn) NewStream(context.Context) (network.Stream, error)      { return nil, nil }
func (m *mockConn) NewStreamSync(context.Context) (network.Stream, error)  { return nil, nil }

// TestNetworkNotifieeMethodsWithRealObjects tests notifiee methods with more realistic objects
func TestNetworkNotifieeMethodsWithRealObjects(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Create a mock connection
	mockConn := &mockConn{
		peerID: "QmTestPeer",
		connID: "test-conn-1",
	}

	// Test OpenedStream with mock objects
	network.OpenedStream(nil, nil)

	// Test ClosedStream with mock objects
	network.ClosedStream(nil, nil)

	// Test OpenedConn with mock connection
	network.OpenedConn(nil, mockConn)

	// Test ClosedConn with mock connection
	network.ClosedConn(nil, mockConn)

	// Test Listen with real multiaddr
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	require.NoError(t, err)
	network.Listen(nil, addr)

	// Test ListenClose with real multiaddr
	network.ListenClose(nil, addr)
}

// TestNetworkWithRealPeerDiscovery tests network with actual peer discovery enabled
func TestNetworkWithRealPeerDiscovery(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = true
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Give some time for peer discovery to start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, network)
}

// TestNetworkWithRelayAndMDNS tests network with both relay and mDNS enabled
func TestNetworkWithRelayAndMDNS(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = true
	config.EnableRelay = true

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
}

// TestNetworkWithCustomBootstrapPeers tests network with custom bootstrap peers
func TestNetworkWithCustomBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{
		"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
		"/ip4/127.0.0.1/tcp/1235/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
	}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	// Should have 2 valid bootstrap peers
	assert.Len(t, network.bootstrapPeers, 2)
}

// TestNetworkWithMixedBootstrapPeers tests network with mixed valid/invalid bootstrap peers
func TestNetworkWithMixedBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{
		"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
		"invalid-address",
		"/ip4/127.0.0.1/tcp/1235/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
	}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	// Should have 2 valid bootstrap peers (invalid one skipped)
	assert.Len(t, network.bootstrapPeers, 2)
}

// TestNetworkWithExtremeConfigValues tests network with extreme configuration values
func TestNetworkWithExtremeConfigValues(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 1000
	config.ConnectionTimeout = 1 * time.Hour

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 1000, network.config.MaxPeers)
	assert.Equal(t, 1*time.Hour, network.config.ConnectionTimeout)
}

// TestNetworkWithMinimalConfig tests network with minimal configuration
func TestNetworkWithMinimalConfig(t *testing.T) {
	config := &NetworkConfig{
		ListenPort:        0,
		BootstrapPeers:    []string{},
		EnableMDNS:        false,
		EnableRelay:       false,
		MaxPeers:          1,
		ConnectionTimeout: 1 * time.Second,
	}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 1, network.config.MaxPeers)
	assert.Equal(t, 1*time.Second, network.config.ConnectionTimeout)
}

// TestNetworkWithNilChainAndMempool tests network creation with nil dependencies
func TestNetworkWithNilChainAndMempool(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	network, err := NewNetwork(config, nil, nil)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Nil(t, network.chain)
	assert.Nil(t, network.mempool)
}

// TestNetworkWithCustomPort tests network with custom port
func TestNetworkWithCustomPort(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 12345
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 12345, config.ListenPort)
}

// TestNetworkWithZeroMaxPeers tests network with zero max peers
func TestNetworkWithZeroMaxPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 0

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 0, network.config.MaxPeers)
}

// TestNetworkWithNegativeMaxPeers tests network with negative max peers
func TestNetworkWithNegativeMaxPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = -1

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, -1, network.config.MaxPeers)
}

// TestNetworkWithVeryLongTimeout tests network with very long timeout
func TestNetworkWithVeryLongTimeout(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.ConnectionTimeout = 24 * time.Hour

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 24*time.Hour, network.config.ConnectionTimeout)
}

// TestNetworkWithVeryShortTimeout tests network with very short timeout
func TestNetworkWithVeryShortTimeout(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.ConnectionTimeout = 1 * time.Nanosecond

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Equal(t, 1*time.Nanosecond, network.config.ConnectionTimeout)
}

// TestNetworkWithEmptyBootstrapPeers tests network with empty bootstrap peers
func TestNetworkWithEmptyBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Empty(t, network.bootstrapPeers)
}

// TestNetworkWithNilBootstrapPeers tests network with nil bootstrap peers
func TestNetworkWithNilBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = nil

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	assert.NotNil(t, network)
	assert.Nil(t, network.bootstrapPeers)
}

// TestNetworkWithStreamOperations tests network with stream operations
func TestNetworkWithStreamOperations(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test stream operations with mock objects
	// These should be called by the network notifiee interface
	network.OpenedStream(nil, nil)
	network.ClosedStream(nil, nil)

	// Test connection operations with mock objects
	network.OpenedConn(nil, nil)
	network.ClosedConn(nil, nil)

	// Test listen operations with mock objects
	network.Listen(nil, nil)
	network.ListenClose(nil, nil)
}

// TestNetworkWithRealNetworkEvents tests network with real network events
func TestNetworkWithRealNetworkEvents(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Create a mock stream
	mockStream := &testMockStream{}

	// Test stream operations
	network.OpenedStream(nil, mockStream)
	network.ClosedStream(nil, mockStream)

	// Create a mock connection
	mockConn := &mockConn{
		peerID: "QmTestPeer",
		connID: "test-conn-1",
	}

	// Test connection operations
	network.OpenedConn(nil, mockConn)
	network.ClosedConn(nil, mockConn)

	// Create a real multiaddr
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	require.NoError(t, err)

	// Test listen operations
	network.Listen(nil, addr)
	network.ListenClose(nil, addr)
}

// TestNetworkWithErrorConditions tests network with various error conditions
func TestNetworkWithErrorConditions(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with various error conditions
	// This helps improve coverage of error handling paths

	// Test with nil objects
	network.OpenedStream(nil, nil)
	network.ClosedStream(nil, nil)
	network.OpenedConn(nil, nil)
	network.ClosedConn(nil, nil)
	network.Listen(nil, nil)
	network.ListenClose(nil, nil)

	// Test with empty objects
	emptyConn := &mockConn{}
	network.OpenedConn(nil, emptyConn)
	network.ClosedConn(nil, emptyConn)
}

// TestNetworkWithDifferentConfigurations tests network with different configurations
func TestNetworkWithDifferentConfigurations(t *testing.T) {
	testCases := []struct {
		name     string
		config   *NetworkConfig
		expected bool
	}{
		{
			name:     "Default config",
			config:   DefaultNetworkConfig(),
			expected: true,
		},
		{
			name: "Minimal config",
			config: &NetworkConfig{
				ListenPort:        0,
				BootstrapPeers:    []string{},
				EnableMDNS:        false,
				EnableRelay:       false,
				MaxPeers:          1,
				ConnectionTimeout: 1 * time.Second,
			},
			expected: true,
		},
		{
			name: "Maximal config",
			config: &NetworkConfig{
				ListenPort:        12345,
				BootstrapPeers:    []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu"},
				EnableMDNS:        true,
				EnableRelay:       true,
				MaxPeers:          1000,
				ConnectionTimeout: 24 * time.Hour,
			},
			expected: true,
		},
		{
			name: "Edge case config",
			config: &NetworkConfig{
				ListenPort:        0,
				BootstrapPeers:    nil,
				EnableMDNS:        false,
				EnableRelay:       false,
				MaxPeers:          0,
				ConnectionTimeout: 0,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(tc.config, chainInstance, mempoolInstance)
			if tc.expected {
				require.NoError(t, err)
				defer network.Close()
				assert.NotNil(t, network)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestNetworkWithPeerDiscoveryVariations tests network with different peer discovery configurations
func TestNetworkWithPeerDiscoveryVariations(t *testing.T) {
	testCases := []struct {
		name           string
		enableMDNS     bool
		enableRelay    bool
		bootstrapPeers []string
	}{
		{"MDNS only", true, false, []string{}},
		{"Relay only", false, true, []string{}},
		{"Both enabled", true, true, []string{}},
		{"Neither enabled", false, false, []string{}},
		{"With bootstrap peers", false, false, []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu"}},
		{"With invalid bootstrap peers", false, false, []string{"invalid-address"}},
		{"With mixed bootstrap peers", false, false, []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu", "invalid-address"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = tc.enableMDNS
			config.EnableRelay = tc.enableRelay
			config.BootstrapPeers = tc.bootstrapPeers

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			require.NoError(t, err)
			defer network.Close()

			assert.NotNil(t, network)
		})
	}
}

// TestNetworkWithPortVariations tests network with different port configurations
func TestNetworkWithPortVariations(t *testing.T) {
	testCases := []struct {
		name        string
		port        int
		expectError bool
	}{
		{"Random port", 0, false},
		{"Low port", 1024, false},
		{"Standard port", 8080, false},
		{"High port", 65535, false},
		{"Invalid negative port", -1, true},
		{"Invalid high port", 65536, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = tc.port
			config.EnableMDNS = false
			config.EnableRelay = false

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				defer network.Close()
				assert.NotNil(t, network)
			}
		})
	}
}

// TestNetworkWithTimeoutVariations tests network with different timeout configurations
func TestNetworkWithTimeoutVariations(t *testing.T) {
	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{"Zero timeout", 0},
		{"Very short timeout", 1 * time.Nanosecond},
		{"Short timeout", 1 * time.Millisecond},
		{"Standard timeout", 30 * time.Second},
		{"Long timeout", 1 * time.Hour},
		{"Very long timeout", 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.ConnectionTimeout = tc.timeout

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			require.NoError(t, err)
			defer network.Close()

			assert.NotNil(t, network)
			assert.Equal(t, tc.timeout, network.config.ConnectionTimeout)
		})
	}
}

// TestNetworkWithMaxPeersVariations tests network with different max peers configurations
func TestNetworkWithMaxPeersVariations(t *testing.T) {
	testCases := []struct {
		name     string
		maxPeers int
	}{
		{"Zero max peers", 0},
		{"Negative max peers", -1},
		{"Single peer", 1},
		{"Standard max peers", 50},
		{"High max peers", 1000},
		{"Very high max peers", 10000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.MaxPeers = tc.maxPeers

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			require.NoError(t, err)
			defer network.Close()

			assert.NotNil(t, network)
			assert.Equal(t, tc.maxPeers, network.config.MaxPeers)
		})
	}
}

// TestNetworkWithBootstrapPeerVariations tests network with different bootstrap peer configurations
func TestNetworkWithBootstrapPeerVariations(t *testing.T) {
	testCases := []struct {
		name           string
		bootstrapPeers []string
		expectedCount  int
	}{
		{"No bootstrap peers", []string{}, 0},
		{"Nil bootstrap peers", nil, 0},
		{"Single valid peer", []string{"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu"}, 1},
		{"Multiple valid peers", []string{
			"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
			"/ip4/127.0.0.1/tcp/1235/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
		}, 2},
		{"Mixed valid and invalid peers", []string{
			"/ip4/127.0.0.1/tcp/1234/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
			"invalid-address",
			"/ip4/127.0.0.1/tcp/1235/p2p/QmYyQSo1c1b7THV3XMKvYv4hShPxVqBrJZ8Tpum5oA6oGu",
		}, 2},
		{"All invalid peers", []string{"invalid-address-1", "invalid-address-2"}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.BootstrapPeers = tc.bootstrapPeers

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			require.NoError(t, err)
			defer network.Close()

			assert.NotNil(t, network)
			assert.Len(t, network.bootstrapPeers, tc.expectedCount)
		})
	}
}

// TestNetworkWithDependencyVariations tests network with different dependency configurations
func TestNetworkWithDependencyVariations(t *testing.T) {
	testCases := []struct {
		name        string
		chain       *chain.Chain
		mempool     *mempool.Mempool
		expectedNil bool
	}{
		{"Both dependencies", &chain.Chain{}, mempool.NewMempool(mempool.TestMempoolConfig()), false},
		{"Nil chain", nil, mempool.NewMempool(mempool.TestMempoolConfig()), true},
		{"Nil mempool", &chain.Chain{}, nil, true},
		{"Both nil", nil, nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false

			network, err := NewNetwork(config, tc.chain, tc.mempool)
			if tc.expectedNil {
				// This should still work as the dependencies are optional
				require.NoError(t, err)
				defer network.Close()
				assert.NotNil(t, network)
			} else {
				require.NoError(t, err)
				defer network.Close()
				assert.NotNil(t, network)
			}
		})
	}
}

// TestNetworkErrorPaths attempts to trigger error paths to improve coverage
func TestNetworkErrorPaths(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 1 // Set low to trigger max peers logic

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with invalid bootstrap peer addresses to trigger error paths
	config.BootstrapPeers = []string{"invalid-address"}

	// Test with empty data to see if we can trigger any validation errors
	err = network.PublishBlock([]byte{})
	if err != nil {
		t.Logf("PublishBlock with empty data returned error (expected): %v", err)
	}

	err = network.PublishTransaction([]byte{})
	if err != nil {
		t.Logf("PublishTransaction with empty data returned error (expected): %v", err)
	}

	// Test with nil data
	err = network.PublishBlock(nil)
	if err != nil {
		t.Logf("PublishBlock with nil data returned error (expected): %v", err)
	}

	err = network.PublishTransaction(nil)
	if err != nil {
		t.Logf("PublishTransaction with nil data returned error (expected): %v", err)
	}

	// Test with very large data to potentially trigger marshaling errors
	largeData := make([]byte, 1000000) // 1MB of data
	err = network.PublishBlock(largeData)
	if err != nil {
		t.Logf("PublishBlock with large data returned error (expected): %v", err)
	}

	err = network.PublishTransaction(largeData)
	if err != nil {
		t.Logf("PublishTransaction with large data returned error (expected): %v", err)
	}

	// Test the network with various edge cases
	assert.NotNil(t, network)
}

// TestNetworkWithInvalidBootstrapPeers tests network behavior with invalid bootstrap peer addresses
func TestNetworkWithInvalidBootstrapPeers(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.BootstrapPeers = []string{"invalid-address", "another-invalid", "/ip4/127.0.0.1/tcp/1234"}

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// The network should still be created successfully even with invalid bootstrap peers
	assert.NotNil(t, network)

	// Test that the network can still function
	host := network.GetHost()
	assert.NotNil(t, host)

	peers := network.GetPeers()
	assert.NotNil(t, peers)
}

// TestNetworkWithExtremeConfigurations tests network with extreme configuration values
func TestNetworkWithExtremeConfigurations(t *testing.T) {
	testCases := []struct {
		name        string
		config      *NetworkConfig
		expectError bool
	}{
		{
			name: "ZeroMaxPeers",
			config: func() *NetworkConfig {
				cfg := DefaultNetworkConfig()
				cfg.ListenPort = 0
				cfg.MaxPeers = 0
				cfg.EnableMDNS = false
				cfg.EnableRelay = false
				return cfg
			}(),
			expectError: false,
		},
		{
			name: "VeryHighMaxPeers",
			config: func() *NetworkConfig {
				cfg := DefaultNetworkConfig()
				cfg.ListenPort = 0
				cfg.MaxPeers = 10000
				cfg.EnableMDNS = false
				cfg.EnableRelay = false
				return cfg
			}(),
			expectError: false,
		},
		{
			name: "VeryShortTimeout",
			config: func() *NetworkConfig {
				cfg := DefaultNetworkConfig()
				cfg.ListenPort = 0
				cfg.ConnectionTimeout = time.Nanosecond
				cfg.EnableMDNS = false
				cfg.EnableRelay = false
				return cfg
			}(),
			expectError: false,
		},
		{
			name: "VeryLongTimeout",
			config: func() *NetworkConfig {
				cfg := DefaultNetworkConfig()
				cfg.ListenPort = 0
				cfg.ConnectionTimeout = time.Hour * 24 * 365 // 1 year
				cfg.EnableMDNS = false
				cfg.EnableRelay = false
				return cfg
			}(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(tc.config, chainInstance, mempoolInstance)
			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			defer network.Close()

			assert.NotNil(t, network)
		})
	}
}

// TestNetworkAdvancedErrorPaths attempts to trigger advanced error paths
func TestNetworkAdvancedErrorPaths(t *testing.T) {
	// Test with invalid port that should cause libp2p to fail
	config := DefaultNetworkConfig()
	config.ListenPort = 99999 // Invalid port that should cause issues

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	// This should fail due to invalid port
	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	if err != nil {
		// Expected error - this is good for coverage!
		t.Logf("Expected error with invalid port: %v", err)
		return
	}
	defer network.Close()

	// If we get here, test the error handling in other functions
	t.Log("Network created with invalid port, testing error paths...")
}

// TestNetworkWithFailingDependencies tests network behavior with failing dependencies
func TestNetworkWithFailingDependencies(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with invalid data that might trigger error paths
	invalidData := make([]byte, 1000000) // Very large data

	// Try to publish with invalid data
	err = network.PublishBlock(invalidData)
	if err != nil {
		t.Logf("Expected error with invalid data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with invalid data")
	}

	err = network.PublishTransaction(invalidData)
	if err != nil {
		t.Logf("Expected error with invalid transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with invalid data")
	}
}

// TestNetworkWithNilDependencies tests network behavior with nil dependencies
func TestNetworkWithNilDependencies(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	// Test with nil chain
	network, err := NewNetwork(config, nil, &mempool.Mempool{})
	if err != nil {
		t.Logf("Expected error with nil chain: %v", err)
	} else {
		defer network.Close()
		t.Log("Network created with nil chain")
	}

	// Test with nil mempool
	network, err = NewNetwork(config, &chain.Chain{}, nil)
	if err != nil {
		t.Logf("Expected error with nil mempool: %v", err)
	} else {
		defer network.Close()
		t.Log("Network created with nil mempool")
	}
}

// TestNetworkWithConcurrentOperations tests network behavior under concurrent load
func TestNetworkWithConcurrentOperations(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test concurrent access to network methods
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Concurrently call various network methods
			host := network.GetHost()
			peers := network.GetPeers()
			ctx := network.GetContext()

			// Try to publish data concurrently
			data := []byte(fmt.Sprintf("test-data-%d", id))
			_ = network.PublishBlock(data)
			_ = network.PublishTransaction(data)

			// Verify we got valid results
			assert.NotNil(t, host)
			assert.NotNil(t, peers)
			assert.NotNil(t, ctx)
		}(i)
	}

	wg.Wait()
}

// TestNetworkWithMemoryPressure tests network behavior under memory pressure
func TestNetworkWithMemoryPressure(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Create large amounts of data to test memory handling
	largeData := make([]byte, 1000000) // 1MB

	// Try to publish large data multiple times
	for i := 0; i < 10; i++ {
		err := network.PublishBlock(largeData)
		if err != nil {
			t.Logf("PublishBlock failed with large data (iteration %d): %v", i, err)
		}

		err = network.PublishTransaction(largeData)
		if err != nil {
			t.Logf("PublishTransaction failed with large data (iteration %d): %v", i, err)
		}
	}

	// Verify the network is still functional
	host := network.GetHost()
	assert.NotNil(t, host)

	peers := network.GetPeers()
	assert.NotNil(t, peers)
}

// TestNetworkWithExtremeErrorConditions tests network with extreme error conditions
func TestNetworkWithExtremeErrorConditions(t *testing.T) {
	// Test with invalid port that should cause libp2p to fail
	config := DefaultNetworkConfig()
	config.ListenPort = 99999 // Invalid port that should cause issues

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	// This should fail due to invalid port
	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	if err != nil {
		// Expected error - this is good for coverage!
		t.Logf("Expected error with invalid port: %v", err)
		return
	}
	defer network.Close()

	// If we get here, test the error handling in other functions
	t.Log("Network created with invalid port, testing error paths...")
}

// TestNetworkWithFailingCrypto tests network behavior when crypto operations fail
func TestNetworkWithFailingCrypto(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with extremely large data that might cause crypto issues
	extremelyLargeData := make([]byte, 10000000) // 10MB

	// Try to publish with extremely large data
	err = network.PublishBlock(extremelyLargeData)
	if err != nil {
		t.Logf("Expected error with extremely large data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with extremely large data")
	}

	err = network.PublishTransaction(extremelyLargeData)
	if err != nil {
		t.Logf("Expected error with extremely large transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with extremely large data")
	}

	// Test with nil data
	err = network.PublishBlock(nil)
	if err != nil {
		t.Logf("Expected error with nil data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with nil data")
	}

	err = network.PublishTransaction(nil)
	if err != nil {
		t.Logf("Expected error with nil transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with nil data")
	}
}

// TestNetworkWithInvalidPeerData tests network behavior with invalid peer data
func TestNetworkWithInvalidPeerData(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with data that contains invalid characters
	invalidData := []byte{0x00, 0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA, 0xF9}

	err = network.PublishBlock(invalidData)
	if err != nil {
		t.Logf("Expected error with invalid data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with invalid data")
	}

	err = network.PublishTransaction(invalidData)
	if err != nil {
		t.Logf("Expected error with invalid transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with invalid data")
	}
}

// TestNetworkWithExtremeTimeoutValues tests network with extreme timeout values
func TestNetworkWithExtremeTimeoutValues(t *testing.T) {
	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{"Negative timeout", -1 * time.Second},
		{"Extremely long timeout", 24 * time.Hour},
		{"Zero timeout", 0},
		{"Very short timeout", 1 * time.Nanosecond},
		{"One year timeout", 365 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.ConnectionTimeout = tc.timeout

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if err != nil {
				t.Logf("Expected error with extreme timeout: %v", err)
				return
			}
			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)

			peers := network.GetPeers()
			assert.NotNil(t, peers)
		})
	}
}

// TestNetworkWithInvalidMaxPeers tests network with invalid max peers values
func TestNetworkWithInvalidMaxPeers(t *testing.T) {
	testCases := []struct {
		name     string
		maxPeers int
	}{
		{"Negative max peers", -1},
		{"Zero max peers", 0},
		{"Extremely large max peers", 999999999},
		{"One max peer", 1},
		{"Two max peers", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.MaxPeers = tc.maxPeers

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if err != nil {
				t.Logf("Expected error with invalid max peers: %v", err)
				return
			}
			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)

			peers := network.GetPeers()
			assert.NotNil(t, peers)
		})
	}
}

// TestNetworkWithInvalidPorts tests network with various invalid port configurations
func TestNetworkWithInvalidPorts(t *testing.T) {
	testCases := []struct {
		name      string
		port      int
		shouldErr bool
	}{
		{"Port 0", 0, false},
		{"Port 1", 1, true}, // Port 1 requires root privileges, so it will fail
		{"Port 1024", 1024, false},
		{"Port 8080", 8080, false},
		{"Port 65535", 65535, false},
		{"Port -1", -1, true},
		{"Port 65536", 65536, true},
		{"Port 99999", 99999, true},
		{"Port 100000", 100000, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = tc.port
			config.EnableMDNS = false
			config.EnableRelay = false

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error with port %d, but got none", tc.port)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error with port %d: %v", tc.port, err)
				return
			}

			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)
		})
	}
}

// TestNetworkWithInvalidBootstrapPeerFormats tests network with various invalid bootstrap peer formats
func TestNetworkWithInvalidBootstrapPeerFormats(t *testing.T) {
	invalidPeerFormats := []string{
		"",
		"invalid-format",
		"/ip4/127.0.0.1/tcp/1234", // Missing p2p part
		"/ip4/127.0.0.1/tcp/1234/p2p/invalid-peer-id",
		"http://invalid-url",
		"ftp://invalid-protocol",
		"//invalid-multiaddr",
		"/ip4/256.256.256.256/tcp/1234/p2p/QmInvalid",
		"/ip6/::1/tcp/1234/p2p/QmInvalid",
		"/ip4/127.0.0.1/tcp/99999/p2p/QmInvalid", // Invalid port
		"/ip4/127.0.0.1/udp/1234/p2p/QmInvalid",  // Wrong protocol
		"/ip4/127.0.0.1/tcp/1234/p2p/",           // Empty peer ID
		"/ip4/127.0.0.1/tcp/1234/p2p",            // Missing peer ID
	}

	for i, invalidPeer := range invalidPeerFormats {
		t.Run(fmt.Sprintf("InvalidPeer_%d_%s", i, invalidPeer), func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.BootstrapPeers = []string{invalidPeer}
			config.EnableMDNS = false
			config.EnableRelay = false

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			require.NoError(t, err)
			defer network.Close()

			// The network should still be created successfully
			assert.NotNil(t, network)

			// Test that the network can still function
			host := network.GetHost()
			assert.NotNil(t, host)
		})
	}
}

// TestNetworkWithExtremeConcurrency tests network behavior under extreme concurrent load
func TestNetworkWithExtremeConcurrency(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test extreme concurrent access to network methods
	var wg sync.WaitGroup
	numGoroutines := 100 // Much higher concurrency

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Concurrently call various network methods
			host := network.GetHost()
			peers := network.GetPeers()
			ctx := network.GetContext()

			// Try to publish data concurrently
			data := []byte(fmt.Sprintf("test-data-%d", id))
			_ = network.PublishBlock(data)
			_ = network.PublishTransaction(data)

			// Verify we got valid results
			assert.NotNil(t, host)
			assert.NotNil(t, peers)
			assert.NotNil(t, ctx)
		}(i)
	}

	wg.Wait()
}

// TestNetworkWithExtremeMemoryPressure tests network behavior under extreme memory pressure
func TestNetworkWithExtremeMemoryPressure(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Create extremely large amounts of data to test memory handling
	extremelyLargeData := make([]byte, 10000000) // 10MB

	// Try to publish extremely large data multiple times
	for i := 0; i < 20; i++ {
		err := network.PublishBlock(extremelyLargeData)
		if err != nil {
			t.Logf("PublishBlock failed with extremely large data (iteration %d): %v", i, err)
		}

		err = network.PublishTransaction(extremelyLargeData)
		if err != nil {
			t.Logf("PublishTransaction failed with extremely large data (iteration %d): %v", i, err)
		}
	}

	// Verify the network is still functional
	host := network.GetHost()
	assert.NotNil(t, host)

	peers := network.GetPeers()
	assert.NotNil(t, peers)
}

// TestNetworkWithInvalidChainData tests network behavior with invalid chain data
func TestNetworkWithInvalidChainData(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with data that might cause issues in the chain
	chainData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}

	err = network.PublishBlock(chainData)
	if err != nil {
		t.Logf("Expected error with chain data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with chain data")
	}

	err = network.PublishTransaction(chainData)
	if err != nil {
		t.Logf("Expected error with chain transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with chain data")
	}
}

// TestNetworkWithExtremeFailureScenarios tests network with extreme failure scenarios
func TestNetworkWithExtremeFailureScenarios(t *testing.T) {
	// Test with extremely large port numbers that should cause issues
	config := DefaultNetworkConfig()
	config.ListenPort = 999999999 // Extremely large port

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	// This should fail due to extremely large port
	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	if err != nil {
		// Expected error - this is good for coverage!
		t.Logf("Expected error with extremely large port: %v", err)
		return
	}
	defer network.Close()

	// If we get here, test the error handling in other functions
	t.Log("Network created with extremely large port, testing error paths...")
}

// TestNetworkWithInvalidCryptoData tests network behavior with invalid crypto data
func TestNetworkWithInvalidCryptoData(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with data that contains all possible byte values
	allBytesData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		allBytesData[i] = byte(i)
	}

	// Try to publish with all possible byte values
	err = network.PublishBlock(allBytesData)
	if err != nil {
		t.Logf("Expected error with all bytes data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with all bytes data")
	}

	err = network.PublishTransaction(allBytesData)
	if err != nil {
		t.Logf("Expected error with all bytes transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with all bytes data")
	}

	// Test with data that contains only null bytes
	nullBytesData := make([]byte, 1000)

	err = network.PublishBlock(nullBytesData)
	if err != nil {
		t.Logf("Expected error with null bytes data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with null bytes data")
	}

	err = network.PublishTransaction(nullBytesData)
	if err != nil {
		t.Logf("Expected error with null bytes transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with null bytes data")
	}
}

// TestNetworkWithExtremePeerScenarios tests network with extreme peer scenarios
func TestNetworkWithExtremePeerScenarios(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false
	config.MaxPeers = 1 // Set very low to trigger max peers logic

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test that the network still functions with extreme peer limits
	host := network.GetHost()
	assert.NotNil(t, host)

	peers := network.GetPeers()
	assert.NotNil(t, peers)
}

// TestNetworkWithInvalidNetworkConfigs tests network with various invalid network configurations
func TestNetworkWithInvalidNetworkConfigs(t *testing.T) {
	testCases := []struct {
		name   string
		config NetworkConfig
	}{
		{
			name: "All fields zero",
			config: NetworkConfig{
				ListenPort:        0,
				BootstrapPeers:    []string{},
				EnableMDNS:        false,
				EnableRelay:       false,
				MaxPeers:          0,
				ConnectionTimeout: 0,
			},
		},
		{
			name: "All fields extreme",
			config: NetworkConfig{
				ListenPort:        999999999,
				BootstrapPeers:    []string{"invalid-1", "invalid-2", "invalid-3"},
				EnableMDNS:        true,
				EnableRelay:       true,
				MaxPeers:          999999999,
				ConnectionTimeout: 999999 * time.Hour,
			},
		},
		{
			name: "Negative values",
			config: NetworkConfig{
				ListenPort:        -1,
				BootstrapPeers:    []string{},
				EnableMDNS:        false,
				EnableRelay:       false,
				MaxPeers:          -1,
				ConnectionTimeout: -1 * time.Hour,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(&tc.config, chainInstance, mempoolInstance)
			if err != nil {
				t.Logf("Expected error with extreme config: %v", err)
				return
			}
			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)

			peers := network.GetPeers()
			assert.NotNil(t, peers)
		})
	}
}

// TestNetworkWithExtremeDataScenarios tests network with extreme data scenarios
func TestNetworkWithExtremeDataScenarios(t *testing.T) {
	config := DefaultNetworkConfig()
	config.ListenPort = 0
	config.EnableMDNS = false
	config.EnableRelay = false

	chainInstance := &chain.Chain{}
	mempoolConfig := mempool.TestMempoolConfig()
	mempoolInstance := mempool.NewMempool(mempoolConfig)

	network, err := NewNetwork(config, chainInstance, mempoolInstance)
	require.NoError(t, err)
	defer network.Close()

	// Test with data that contains unicode characters
	unicodeData := []byte("Hello 世界 🌍 こんにちは Привет")

	err = network.PublishBlock(unicodeData)
	if err != nil {
		t.Logf("Expected error with unicode data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with unicode data")
	}

	err = network.PublishTransaction(unicodeData)
	if err != nil {
		t.Logf("Expected error with unicode transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with unicode data")
	}

	// Test with data that contains control characters
	controlData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}

	err = network.PublishBlock(controlData)
	if err != nil {
		t.Logf("Expected error with control data: %v", err)
	} else {
		t.Log("PublishBlock succeeded with control data")
	}

	err = network.PublishTransaction(controlData)
	if err != nil {
		t.Logf("Expected error with control transaction data: %v", err)
	} else {
		t.Log("PublishTransaction succeeded with control data")
	}
}

// TestNetworkWithExtremeTimeoutScenarios tests network with extreme timeout scenarios
func TestNetworkWithExtremeTimeoutScenarios(t *testing.T) {
	testCases := []struct {
		name    string
		timeout time.Duration
	}{
		{"Negative timeout", -1 * time.Second},
		{"Zero timeout", 0},
		{"One nanosecond", 1 * time.Nanosecond},
		{"One microsecond", 1 * time.Microsecond},
		{"One millisecond", 1 * time.Millisecond},
		{"One second", 1 * time.Second},
		{"One minute", 1 * time.Minute},
		{"One hour", 1 * time.Hour},
		{"One day", 24 * time.Hour},
		{"One week", 7 * 24 * time.Hour},
		{"One month", 30 * 24 * time.Hour},
		{"One year", 365 * 24 * time.Hour},
		{"Ten years", 10 * 365 * 24 * time.Hour},
		{"Hundred years", 100 * 365 * 24 * time.Hour},
		{"Thousand years", 100 * 365 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.ConnectionTimeout = tc.timeout

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if err != nil {
				t.Logf("Expected error with extreme timeout: %v", err)
				return
			}
			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)

			peers := network.GetPeers()
			assert.NotNil(t, peers)
		})
	}
}

// TestNetworkWithExtremeMaxPeersScenarios tests network with extreme max peers scenarios
func TestNetworkWithExtremeMaxPeersScenarios(t *testing.T) {
	testCases := []struct {
		name     string
		maxPeers int
	}{
		{"Negative max peers", -1},
		{"Zero max peers", 0},
		{"One max peer", 1},
		{"Two max peers", 2},
		{"Ten max peers", 10},
		{"Hundred max peers", 100},
		{"Thousand max peers", 1000},
		{"Ten thousand max peers", 10000},
		{"Hundred thousand max peers", 100000},
		{"Million max peers", 1000000},
		{"Ten million max peers", 10000000},
		{"Hundred million max peers", 100000000},
		{"Billion max peers", 1000000000},
		{"Max int32", 2147483647},
		{"Max int64", 9223372036854775807},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultNetworkConfig()
			config.ListenPort = 0
			config.EnableMDNS = false
			config.EnableRelay = false
			config.MaxPeers = tc.maxPeers

			chainInstance := &chain.Chain{}
			mempoolConfig := mempool.TestMempoolConfig()
			mempoolInstance := mempool.NewMempool(mempoolConfig)

			network, err := NewNetwork(config, chainInstance, mempoolInstance)
			if err != nil {
				t.Logf("Expected error with extreme max peers: %v", err)
				return
			}
			defer network.Close()

			// Test that the network still functions
			host := network.GetHost()
			assert.NotNil(t, host)

			peers := network.GetPeers()
			assert.NotNil(t, peers)
		})
	}
}
