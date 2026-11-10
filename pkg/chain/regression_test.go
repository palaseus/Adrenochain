package chain

import (
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestRestartRestoresCanonicalHistoryAndBalances(t *testing.T) {
	store, err := storage.NewStorage(&storage.StorageConfig{DataDir: t.TempDir()})
	require.NoError(t, err)
	defer store.Close()
	config := DefaultChainConfig()
	cc := consensus.DefaultConsensusConfig()
	chain, err := NewChain(config, cc, store)
	require.NoError(t, err)
	for height := uint64(1); height <= 4; height++ {
		require.NoError(t, chain.AddBlock(createEmptyTestBlock(chain.GetBestBlock(), height, 1)))
	}
	stats := chain.UTXOSet.GetStats()
	restarted, err := NewChain(config, cc, store)
	require.NoError(t, err)
	require.Equal(t, chain.GetHeight(), restarted.GetHeight())
	require.Equal(t, chain.GetTipHash(), restarted.GetTipHash())
	require.Equal(t, stats, restarted.UTXOSet.GetStats())
	require.Equal(t, chain.GetGenesisBlock().CalculateHash(), restarted.GetGenesisBlock().CalculateHash())
	for height := uint64(0); height <= 4; height++ {
		require.Equal(t, chain.GetBlockByHeight(height).CalculateHash(), restarted.GetBlockByHeight(height).CalculateHash())
	}
	// Exercise lazy parent lookup after caches are evicted.
	restarted.mu.Lock()
	delete(restarted.blockByHeight, 2)
	restarted.mu.Unlock()
	require.Equal(t, chain.GetBlockByHeight(2).CalculateHash(), restarted.GetBlockByHeight(2).CalculateHash())
	require.NoError(t, restarted.AddBlock(createEmptyTestBlock(restarted.GetBestBlock(), 5, 1)))
}

func TestFailedTipPersistenceDoesNotPublishBlock(t *testing.T) {
	store, err := storage.NewStorage(&storage.StorageConfig{DataDir: t.TempDir()})
	require.NoError(t, err)
	defer store.Close()
	failing := &MockFailingStorage{StorageInterface: store}
	chain, err := NewChain(DefaultChainConfig(), consensus.DefaultConsensusConfig(), failing)
	require.NoError(t, err)
	tip := chain.GetTipHash()
	stats := chain.UTXOSet.GetStats()
	next := createEmptyTestBlock(chain.GetBestBlock(), 1, 1)
	failing.failOnStoreChainState = true
	require.Error(t, chain.AddBlock(next))
	require.Zero(t, chain.GetHeight())
	require.Equal(t, tip, chain.GetTipHash())
	require.Equal(t, stats, chain.UTXOSet.GetStats())
	failing.failOnStoreChainState = false
	require.NoError(t, chain.AddBlock(next))
}

func TestConcurrentChainReadersAndWriter(t *testing.T) {
	store, err := storage.NewStorage(&storage.StorageConfig{DataDir: t.TempDir()})
	require.NoError(t, err)
	defer store.Close()
	chain, err := NewChain(DefaultChainConfig(), consensus.DefaultConsensusConfig(), store)
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				height := chain.GetHeight()
				chain.GetBlockByHeight(height)
				chain.GetBlock(chain.GetTipHash())
				_, _ = chain.GetAccumulatedDifficulty(height)
			}
		}()
	}
	for height := uint64(1); height <= 5; height++ {
		require.NoError(t, chain.AddBlock(createEmptyTestBlock(chain.GetBestBlock(), height, 1)))
	}
	wg.Wait()
	tip := chain.GetTipHash()
	tip[0] ^= 0xff
	require.NotEqual(t, tip, chain.GetTipHash())
}

func TestBlockSizeMatchesWireEncoding(t *testing.T) {
	store, err := storage.NewStorage(&storage.StorageConfig{DataDir: t.TempDir()})
	require.NoError(t, err)
	defer store.Close()
	chain, err := NewChain(DefaultChainConfig(), consensus.DefaultConsensusConfig(), store)
	require.NoError(t, err)
	b := createEmptyTestBlock(chain.GetBestBlock(), 1, 1)
	encoded, err := b.Serialize()
	require.NoError(t, err)
	require.EqualValues(t, len(encoded), chain.GetBlockSize(b))
	txEncoded, err := b.Transactions[0].Serialize()
	require.NoError(t, err)
	require.EqualValues(t, len(txEncoded), chain.getTransactionSize(b.Transactions[0]))
}
