package utxo

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/gochain/gochain/pkg/block"
	"github.com/stretchr/testify/assert"
)

// calculateTxHash calculates the hash of a transaction for testing purposes.
func calculateTxHash(tx *block.Transaction) []byte {
	data := make([]byte, 0)

	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)

	for _, input := range tx.Inputs {
		data = append(data, input.PrevTxHash...)
		indexBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(indexBytes, input.PrevTxIndex)
		data = append(data, indexBytes...)
		data = append(data, input.ScriptSig...)
		seqBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(seqBytes, input.Sequence)
		data = append(data, seqBytes...)
	}

	for _, output := range tx.Outputs {
		valueBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(valueBytes, output.Value)
		data = append(data, valueBytes...)
		data = append(data, output.ScriptPubKey...)
	}

	lockTimeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)

	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	data = append(data, feeBytes...)

	hash := sha256.Sum256(data)
	return hash[:]
}

func TestUTXOSet(t *testing.T) {
	us := NewUTXOSet()

	utxo1 := &UTXO{
		TxHash:      calculateTxHash(&block.Transaction{Version: 1, Outputs: []*block.TxOutput{{Value: 100, ScriptPubKey: []byte("pubkey1")}}}),
		TxIndex:     0,
		Value:       100,
		ScriptPubKey: []byte("pubkey1"),
		Address:     "addr1",
		IsCoinbase:  false,
		Height:      1,
	}

	// Test AddUTXO
	us.AddUTXO(utxo1)
	assert.Equal(t, 1, us.GetUTXOCount())
	assert.Equal(t, uint64(100), us.GetBalance("addr1"))

	// Test GetUTXO
	retrievedUTXO := us.GetUTXO(utxo1.TxHash, 0)
	assert.Equal(t, utxo1, retrievedUTXO)

	// Test RemoveUTXO
	us.RemoveUTXO(utxo1.TxHash, 0)
	assert.Equal(t, 0, us.GetUTXOCount())
	assert.Equal(t, uint64(0), us.GetBalance("addr1"))
}

func TestProcessBlock(t *testing.T) {
	us := NewUTXOSet()

	// Create a coinbase transaction
	coinbaseTx := &block.Transaction{
		Version: 1,
		Outputs: []*block.TxOutput{
			{Value: 50, ScriptPubKey: []byte("miner_addr")},
		},
		LockTime: 0,
	}
	coinbaseTx.Hash = calculateTxHash(coinbaseTx)

	// Create a regular transaction
	tx1 := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{PrevTxHash: coinbaseTx.Hash, PrevTxIndex: 0, ScriptSig: []byte("sig")},
		},
		Outputs: []*block.TxOutput{
			{Value: 30, ScriptPubKey: []byte("addr2")},
			{Value: 15, ScriptPubKey: []byte("addr1")},
		},
		LockTime: 0,
	}
	tx1.Hash = calculateTxHash(tx1)

	// Create a block
	b := &block.Block{
		Header: &block.Header{
			Height: 1,
		},
		Transactions: []*block.Transaction{coinbaseTx, tx1},
	}

	// Process the block
	err := us.ProcessBlock(b)
	assert.NoError(t, err)

	// Verify UTXOs and balances
	assert.Equal(t, 2, us.GetUTXOCount())
	assert.Equal(t, uint64(30), us.GetBalance("addr2"))
	assert.Equal(t, uint64(15), us.GetBalance("addr1"))
	assert.Equal(t, uint64(0), us.GetBalance("miner_addr")) // Coinbase output should be spent
}