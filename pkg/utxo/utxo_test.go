package utxo

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex" // Added import
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

	// Define dummy public key hashes
	pubkey1PubKeyHash := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	addr1PubKeyHash := []byte{0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	// Convert public key hashes to hex-encoded addresses for GetBalance calls
	addr1HexAddr := hex.EncodeToString(addr1PubKeyHash)

	utxo1 := &UTXO{
		TxHash:       calculateTxHash(&block.Transaction{Version: 1, Outputs: []*block.TxOutput{{Value: 100, ScriptPubKey: pubkey1PubKeyHash}}}),
		TxIndex:      0,
		Value:        100,
		ScriptPubKey: pubkey1PubKeyHash,
		Address:      addr1HexAddr, // Use hex-encoded address
		IsCoinbase:   false,
		Height:       1,
	}

	// Test AddUTXO
	us.AddUTXO(utxo1)
	assert.Equal(t, 1, us.GetUTXOCount())
	assert.Equal(t, uint64(100), us.GetBalance(addr1HexAddr)) // Use hex-encoded address

	// Test GetUTXO
	retrievedUTXO := us.GetUTXO(utxo1.TxHash, 0)
	assert.Equal(t, utxo1, retrievedUTXO)

	// Test RemoveUTXO
	us.RemoveUTXO(utxo1.TxHash, 0)
	assert.Equal(t, 0, us.GetUTXOCount())
	assert.Equal(t, uint64(0), us.GetBalance(addr1HexAddr)) // Use hex-encoded address
}

func TestProcessBlock(t *testing.T) {
	us := NewUTXOSet()

	// Define dummy public key hashes
	minerAddrPubKeyHash := []byte{0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c}
	addr2PubKeyHash := []byte{0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}
	addr1PubKeyHash := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}

	// Convert public key hashes to hex-encoded addresses for GetBalance calls
	minerAddrHex := hex.EncodeToString(minerAddrPubKeyHash)
	addr2Hex := hex.EncodeToString(addr2PubKeyHash)
	addr1Hex := hex.EncodeToString(addr1PubKeyHash)

	// Create a coinbase transaction
	coinbaseTx := &block.Transaction{
		Version: 1,
		Outputs: []*block.TxOutput{
			{Value: 50, ScriptPubKey: minerAddrPubKeyHash}, // Use raw public key hash bytes
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
			{Value: 30, ScriptPubKey: addr2PubKeyHash}, // Use raw public key hash bytes
			{Value: 15, ScriptPubKey: addr1PubKeyHash}, // Use raw public key hash bytes
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
	assert.Equal(t, uint64(30), us.GetBalance(addr2Hex))    // Use hex-encoded address
	assert.Equal(t, uint64(15), us.GetBalance(addr1Hex))    // Use hex-encoded address
	assert.Equal(t, uint64(0), us.GetBalance(minerAddrHex)) // Use hex-encoded address // Coinbase output should be spent
}
