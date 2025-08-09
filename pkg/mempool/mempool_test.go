package mempool

import (
	"testing"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/utxo"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a dummy UTXO
func createDummyUTXO(txHash []byte, txIndex uint32, value uint64, address string) *utxo.UTXO {
	return &utxo.UTXO{
		TxHash:       txHash,
		TxIndex:      txIndex,
		Value:        value,
		ScriptPubKey: []byte(address),
		Address:      address,
		IsCoinbase:   false,
		Height:       1, // Dummy height
	}
}

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
	config.MaxSize = 300
	config.MinFeeRate = 0 // Set MinFeeRate to 0 for this test
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
		Fee:  100,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx2"), PrevTxIndex: 0, ScriptSig: []byte("sig2")},
		},
		Outputs: []*block.TxOutput{
			{Value: 200, ScriptPubKey: []byte("pubkey2")},
		},
	}

	tx3 := &block.Transaction{
		Hash: []byte("tx3"),
		Fee:  1, // Changed from 100 to 1
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte("prev_tx3"), PrevTxIndex: 0, ScriptSig: []byte("sig3")},
		},
		Outputs: []*block.TxOutput{
			{Value: 149, ScriptPubKey: []byte("pubkey3")}, // Changed from 50 to 149
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
		Fee:  100,
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
