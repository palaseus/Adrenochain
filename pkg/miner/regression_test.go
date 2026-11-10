package miner

import (
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/mempool"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConsecutiveMinedBlocksHaveDistinctCoinbases(t *testing.T) {
	store, err := storage.NewStorage(&storage.StorageConfig{DataDir: t.TempDir()})
	require.NoError(t, err)
	defer store.Close()
	cc := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chain.DefaultChainConfig(), cc, store)
	require.NoError(t, err)
	mp := mempool.NewMempool(mempool.DefaultMempoolConfig())
	miner := NewMiner(chain, mp, DefaultMinerConfig(), cc)
	defer miner.Cleanup()
	require.NoError(t, miner.mineNextBlock())
	first := chain.GetBestBlock().Transactions[0]
	require.NoError(t, miner.mineNextBlock())
	second := chain.GetBestBlock().Transactions[0]
	require.NotEqual(t, first.Hash, second.Hash)
	require.NotNil(t, chain.UTXOSet.GetUTXO(first.Hash, 0))
	require.NotNil(t, chain.UTXOSet.GetUTXO(second.Hash, 0))
	require.EqualValues(t, 2, chain.GetHeight())
}
