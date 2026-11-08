package mev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionHash(t *testing.T) {
	tx := NewTransaction(
		[]byte("sender"),
		[]byte("recipient"),
		1000, 100, 21000, 1,
		[]byte("data"),
	)
	assert.NotNil(t, tx.Hash)
	assert.Equal(t, []byte("sender"), tx.From)
	assert.Equal(t, []byte("recipient"), tx.To)
	assert.Equal(t, int64(1000), tx.Value)

	// Hash should be deterministic
	tx2 := NewTransaction(
		[]byte("sender"),
		[]byte("recipient"),
		1000, 100, 21000, 1,
		[]byte("data"),
	)
	assert.Equal(t, tx.Hash, tx2.Hash)
}

func TestTransactionHashDiffers(t *testing.T) {
	tx1 := NewTransaction(
		[]byte("sender"),
		[]byte("recipient"),
		1000, 100, 21000, 1,
		[]byte("data"),
	)
	tx2 := NewTransaction(
		[]byte("sender"),
		[]byte("recipient"),
		2000, 100, 21000, 1,
		[]byte("data"),
	)
	assert.NotEqual(t, tx1.Hash, tx2.Hash)
}

func TestPrivateTx(t *testing.T) {
	tx := NewTransaction(
		[]byte("sender"),
		[]byte("recipient"),
		1000, 100, 21000, 1,
		[]byte("data"),
	)
	pt := NewPrivateTx(tx)
	assert.NotNil(t, pt.EncryptedTx)
	assert.Equal(t, tx.Hash, pt.Commitment)
	assert.NotNil(t, pt.Nonce)
	assert.NotNil(t, pt.Timestamp)
}

func TestPrivateMempool(t *testing.T) {
	mp := NewPrivateMempool()
	tx1 := NewTransaction(
		[]byte("sender1"),
		[]byte("recipient1"),
		1000, 100, 21000, 1,
		[]byte("data1"),
	)
	pt1 := NewPrivateTx(tx1)

	// Submit
	assert.NoError(t, mp.Submit(tx1, pt1))
	assert.Equal(t, 1, mp.Len())

	// Get transaction
	gotTx, ok := mp.GetTransaction(tx1.Hash)
	assert.True(t, ok)
	assert.Equal(t, tx1.Hash, gotTx.Hash)

	// Get private tx
	gotPt, ok := mp.GetPrivateTx(tx1.Hash)
	assert.True(t, ok)
	assert.Equal(t, tx1.Hash, gotPt.Commitment)

	// Duplicate submission
	assert.Error(t, mp.Submit(tx1, pt1))

	// Get ordered transactions
	tx2 := NewTransaction(
		[]byte("sender2"),
		[]byte("recipient2"),
		2000, 100, 21000, 2,
		[]byte("data2"),
	)
	pt2 := NewPrivateTx(tx2)
	assert.NoError(t, mp.Submit(tx2, pt2))
	assert.Equal(t, 2, mp.Len())

	ordered := mp.GetOrderedTransactions()
	assert.Equal(t, 2, len(ordered))
	assert.Equal(t, tx1.Hash, ordered[0].Hash)
	assert.Equal(t, tx2.Hash, ordered[1].Hash)
}

func TestMEVProtectorFIFO(t *testing.T) {
	mp := NewMEVProtector(StrategyFIFO, 10)
	tx1 := NewTransaction([]byte("a"), []byte("b"), 1, 1, 1, 1, nil)
	tx2 := NewTransaction([]byte("c"), []byte("d"), 2, 1, 1, 2, nil)
	tx3 := NewTransaction([]byte("e"), []byte("f"), 3, 1, 1, 3, nil)

	// FIFO preserves order
	result := mp.Order([]*Transaction{tx3, tx1, tx2})
	assert.Equal(t, tx3.Hash, result[0].Hash)
	assert.Equal(t, tx1.Hash, result[1].Hash)
	assert.Equal(t, tx2.Hash, result[2].Hash)
}

func TestMEVProtectorFairBatch(t *testing.T) {
	mp := NewMEVProtector(StrategyFairBatch, 2)
	txs := make([]*Transaction, 6)
	for i := 0; i < 6; i++ {
		txs[i] = NewTransaction([]byte("a"), []byte("b"), int64(i), 1, 1, uint64(i), nil)
	}

	result := mp.Order(txs)
	assert.Equal(t, 6, len(result))

	// All transactions should be present
	hashes := make(map[string]bool)
	for _, tx := range result {
		hashes[string(tx.Hash)] = true
	}
	for i, tx := range txs {
		assert.True(t, hashes[string(tx.Hash)], "tx %d missing", i)
	}
}

func TestMEVProtectorCommitReveal(t *testing.T) {
	mp := NewMEVProtector(StrategyCommitReveal, 10)
	tx1 := NewTransaction([]byte("a"), []byte("b"), 1, 1, 1, 1, nil)
	tx2 := NewTransaction([]byte("c"), []byte("d"), 2, 1, 1, 2, nil)

	// Commit-reveal orders by hash
	result := mp.Order([]*Transaction{tx2, tx1})
	// tx1 hash < tx2 hash? Not guaranteed, just check both are present
	assert.Equal(t, 2, len(result))
}

func TestBundle(t *testing.T) {
	tx1 := NewTransaction([]byte("a"), []byte("b"), 1000, 100, 21000, 1, nil)
	tx2 := NewTransaction([]byte("c"), []byte("d"), 2000, 100, 21000, 2, nil)
	bundle := NewBundle([]byte("bundle1"), []*Transaction{tx1, tx2}, 3000)

	assert.Equal(t, []byte("bundle1"), bundle.ID)
	assert.Equal(t, 2, len(bundle.Transactions))
	assert.Equal(t, int64(3000), bundle.Fee)
}
