package net

import (
	"context"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/mempool"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/stretchr/testify/assert"
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
