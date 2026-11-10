package mempool

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestSelectionAndRemovalKeepHeapsConsistent(t *testing.T) {
	mp := NewMempool(TestMempoolConfig())
	for i := 0; i < 12; i++ {
		tx := createBasicValidTransaction(fmt.Sprintf("heap-%d", i), uint64(300+i*100))
		require.NoError(t, mp.AddTransaction(tx))
	}
	// Repeated selection used to overwrite the live heap's indices.
	for i := 0; i < 4; i++ {
		require.Len(t, mp.GetTransactionsForBlock(10000), 12)
	}
	entries := mp.GetTransactionsForBlock(10000)
	for i := 0; i < len(entries); i += 2 {
		require.True(t, mp.RemoveTransaction(entries[i].Hash))
	}
	require.Equal(t, mp.GetTransactionCount(), mp.byFee.Len())
	require.Equal(t, mp.GetTransactionCount(), mp.byTime.Len())
	for i, entry := range *mp.byFee {
		require.Equal(t, i, entry.feeIndex)
		require.Same(t, entry, mp.transactions[string(entry.Transaction.Hash)])
	}
	for i, entry := range *mp.byTime {
		require.Equal(t, i, entry.timeIndex)
		require.Same(t, entry, mp.transactions[string(entry.Transaction.Hash)])
	}
	require.Equal(t, 6, mp.CleanupExpiredTransactions(-time.Second))
	require.Zero(t, mp.byFee.Len())
	require.Zero(t, mp.byTime.Len())
	require.Zero(t, mp.GetSize())
}

func TestConcurrentSelectionAndValidation(t *testing.T) {
	mp := NewMempool(TestMempoolConfig())
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tx := createBasicValidTransaction(fmt.Sprintf("worker-%d-%d", worker, j), 1000)
				_ = mp.IsTransactionValid(tx)
				_ = mp.AddTransaction(tx)
				mp.GetTransactionsForBlock(10000)
				mp.RemoveTransaction(tx.Hash)
			}
		}(i)
	}
	wg.Wait()
	require.Zero(t, mp.GetTransactionCount())
	require.Zero(t, mp.byFee.Len())
	require.Zero(t, mp.byTime.Len())
}
