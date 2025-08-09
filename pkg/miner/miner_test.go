package miner

import (
	"os"
	"testing"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/consensus"
	"github.com/gochain/gochain/pkg/mempool"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestMiner(t *testing.T) {
	dataDir := "./test_miner_data_test_miner"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}
	mempool := mempool.NewMempool(mempool.DefaultMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	// Test Start and Stop
	err = miner.StartMining()
	assert.NoError(t, err)
	assert.True(t, miner.IsMining())

	miner.StopMining()
	assert.False(t, miner.IsMining())
}

func TestCreateNewBlock(t *testing.T) {
	dataDir := "./test_miner_data_test_create_new_block"
	defer os.RemoveAll(dataDir)

	storage, err := storage.NewStorage(&storage.StorageConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, storage)
	if err != nil {
		t.Fatalf("NewChain returned error: %v", err)
	}
	mempool := mempool.NewMempool(mempool.DefaultMempoolConfig())
	config := DefaultMinerConfig()
	miner := NewMiner(chain, mempool, config, consensusConfig)

	prevBlock := chain.GetBestBlock()

	newBlock := miner.createNewBlock(prevBlock)
	assert.NotNil(t, newBlock)
	assert.Equal(t, prevBlock.Header.Height+1, newBlock.Header.Height)
	assert.Equal(t, prevBlock.CalculateHash(), newBlock.Header.PrevBlockHash)
}
