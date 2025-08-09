package mempool

import (
	"testing"
	

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

func TestMempool(t *testing.T) {
	config := DefaultMempoolConfig()
	mp := NewMempool(config)

	tx1 := &block.Transaction{
		Hash: []byte("tx1"),
		Fee:  100,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx"), PrevTxIndex: 0, ScriptSig: []byte("sig")},
		},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte("pubkey")},
		},
	}

	// Test AddTransaction
	err := mp.AddTransaction(tx1)
	assert.NoError(t, err)
	assert.Equal(t, 1, mp.GetTransactionCount())

	// Test adding a duplicate transaction
	err = mp.AddTransaction(tx1)
	assert.Error(t, err)

	// Test GetTransaction
	retrievedTx := mp.GetTransaction([]byte("tx1"))
	assert.Equal(t, tx1, retrievedTx)

	// Test RemoveTransaction
	mp.RemoveTransaction([]byte("tx1"))
	assert.Equal(t, 0, mp.GetTransactionCount())
}

func TestMempoolEviction(t *testing.T) {
	config := DefaultMempoolConfig()
	config.MaxSize = 200
	mp := NewMempool(config)

	tx1 := &block.Transaction{
		Hash: []byte("tx1"),
		Fee:  100,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx1"), PrevTxIndex: 0, ScriptSig: []byte("sig1")},
		},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte("pubkey1")},
		},
	}

	tx2 := &block.Transaction{
		Hash: []byte("tx2"),
		Fee:  200,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx2"), PrevTxIndex: 0, ScriptSig: []byte("sig2")},
		},
		Outputs: []*block.TxOutput{
			{Value: 200, ScriptPubKey: []byte("pubkey2")},
		},
	}

	tx3 := &block.Transaction{
		Hash: []byte("tx3"),
		Fee:  100,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx3"), PrevTxIndex: 0, ScriptSig: []byte("sig3")},
		},
		Outputs: []*block.TxOutput{
			{Value: 50, ScriptPubKey: []byte("pubkey3")},
		},
	}

	err := mp.AddTransaction(tx1)
	assert.NoError(t, err)

	err = mp.AddTransaction(tx2)
	assert.NoError(t, err)

	err = mp.AddTransaction(tx3)
	assert.NoError(t, err)

	// This should evict tx3
	tx4 := &block.Transaction{
		Hash: []byte("tx4"),
		Fee:  150,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx4"), PrevTxIndex: 0, ScriptSig: []byte("sig4")},
		},
		Outputs: []*block.TxOutput{
			{Value: 150, ScriptPubKey: []byte("pubkey4")},
		},
	}

	err = mp.AddTransaction(tx4)
	assert.NoError(t, err)

	assert.Nil(t, mp.GetTransaction([]byte("tx3")))
}
