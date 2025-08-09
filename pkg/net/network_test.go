package net

import (
	"context"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/mempool"
	proto_net "github.com/gochain/gochain/pkg/proto/net"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewNetwork(t *testing.T) {
	// Create dummy chain and mempool for testing
	dummyStorage, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_new_network"))
	assert.NoError(t, err)
	defer dummyStorage.Close()

	dummyChainConfig := chain.DefaultChainConfig()
	dummyChain, err := chain.NewChain(dummyChainConfig, dummyStorage)
	assert.NoError(t, err)

	dummyMempoolConfig := mempool.DefaultMempoolConfig()
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
	dummyChain1, err := chain.NewChain(dummyChainConfig1, dummyStorage1)
	assert.NoError(t, err)

	dummyMempoolConfig1 := mempool.DefaultMempoolConfig()
	dummyMempool1 := mempool.NewMempool(dummyMempoolConfig1)

	dummyStorage2, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_network_connect_2"))
	assert.NoError(t, err)
	defer dummyStorage2.Close()

	dummyChainConfig2 := chain.DefaultChainConfig()
	dummyChain2, err := chain.NewChain(dummyChainConfig2, dummyStorage2)
	assert.NoError(t, err)

	dummyMempoolConfig2 := mempool.DefaultMempoolConfig()
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
		FromPeerId: peerIDBytes,
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
	dummyChain1, err := chain.NewChain(dummyChainConfig1, dummyStorage1)
	assert.NoError(t, err)

	dummyMempoolConfig1 := mempool.DefaultMempoolConfig()
	dummyMempool1 := mempool.NewMempool(dummyMempoolConfig1)

	dummyStorage2, err := storage.NewStorage(storage.DefaultStorageConfig().WithDataDir("./test_data_net_test_publish_subscribe_2"))
	assert.NoError(t, err)
	defer dummyStorage2.Close()

	dummyChainConfig2 := chain.DefaultChainConfig()
	dummyChain2, err := chain.NewChain(dummyChainConfig2, dummyStorage2)
	assert.NoError(t, err)

	dummyMempoolConfig2 := mempool.DefaultMempoolConfig()
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

	// Connect the two networks
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	peerInfo2 := net2.GetHost().Peerstore().PeerInfo(net2.GetHost().ID())
	err = net1.GetHost().Connect(ctx, peerInfo2)
	assert.NoError(t, err)

	// Give time for connection to establish and pubsub to propagate
	time.Sleep(2 * time.Second)

	// Subscribe to blocks on net2
	blockSub2, err := net2.SubscribeToBlocks()
	assert.NoError(t, err)
	defer blockSub2.Cancel()

	// Publish a block from net1
	blockData := []byte(`"test block data"`)
	err = net1.PublishBlock(blockData)
	assert.NoError(t, err)

	// Wait for message on net2
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
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for block message")
	}

	// Subscribe to transactions on net1
	txSub1, err := net1.SubscribeToTransactions()
	assert.NoError(t, err)
	defer txSub1.Cancel()

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
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for transaction message")
	}
}