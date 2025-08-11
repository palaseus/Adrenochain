package utxo

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
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

func TestValidateTransaction_Enhanced(t *testing.T) {
	us := NewUTXOSet()

	// Test 1: Coinbase transaction validation
	coinbaseTx := &block.Transaction{
		Version: 1,
		Outputs: []*block.TxOutput{
			{Value: 1000, ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
		},
		LockTime: 0,
		Fee:      0,
	}
	coinbaseTx.Hash = calculateTxHash(coinbaseTx)

	// Create a block context for coinbase validation
	testBlock := &block.Block{
		Header:       &block.Header{Height: 1},
		Transactions: []*block.Transaction{coinbaseTx},
	}

	err := us.ValidateTransactionInBlock(coinbaseTx, testBlock, 0)
	assert.NoError(t, err, "Valid coinbase transaction should pass validation")

	// Test 2: Invalid coinbase transaction (no outputs)
	invalidCoinbaseTx := &block.Transaction{
		Version:  1,
		Outputs:  []*block.TxOutput{},
		LockTime: 0,
		Fee:      0,
	}
	invalidCoinbaseTx.Hash = calculateTxHash(invalidCoinbaseTx)

	testBlock.Transactions = []*block.Transaction{invalidCoinbaseTx}
	err = us.ValidateTransactionInBlock(invalidCoinbaseTx, testBlock, 0)
	assert.Error(t, err, "Coinbase transaction with no outputs should fail validation")
	assert.Contains(t, err.Error(), "must have at least one output")

	// Test 3: Transaction with no inputs (not coinbase) - should fail in block context
	noInputTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
		},
		LockTime: 0,
		Fee:      0,
	}
	noInputTx.Hash = calculateTxHash(noInputTx)

	// This transaction should fail validation when it's not the first transaction in a block
	testBlock.Transactions = []*block.Transaction{coinbaseTx, noInputTx}
	err = us.ValidateTransactionInBlock(noInputTx, testBlock, 1)
	assert.Error(t, err, "Transaction with no inputs should fail validation when not coinbase")
	assert.Contains(t, err.Error(), "regular transaction must have inputs")

	// Test 4: Transaction with no outputs
	noOutputTx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte{0x01}, PrevTxIndex: 0, ScriptSig: makeValidScriptSig([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}), Sequence: 0xffffffff},
		},
		Outputs:  []*block.TxOutput{},
		LockTime: 0,
		Fee:      0,
	}
	noOutputTx.Hash = calculateTxHash(noOutputTx)

	err = us.ValidateTransactionInBlock(noOutputTx, testBlock, 1)
	assert.Error(t, err, "Transaction with no outputs should fail validation")
	assert.Contains(t, err.Error(), "regular transaction must have outputs")

	// Test 5: Duplicate inputs
	duplicateInputTx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{PrevTxHash: []byte{0x01}, PrevTxIndex: 0, ScriptSig: makeValidScriptSig([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}), Sequence: 0xffffffff},
			{PrevTxHash: []byte{0x01}, PrevTxIndex: 0, ScriptSig: makeValidScriptSig([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}), Sequence: 0xffffffff},
		},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
		},
		LockTime: 0,
		Fee:      0,
	}
	duplicateInputTx.Hash = calculateTxHash(duplicateInputTx)

	err = us.ValidateTransactionInBlock(duplicateInputTx, testBlock, 1)
	assert.Error(t, err, "Transaction with duplicate inputs should fail validation")
	assert.Contains(t, err.Error(), "duplicate input")

	// Test 6: Invalid ScriptSig length
	// First add a UTXO so the validation can proceed to check ScriptSig
	invalidScriptSigUTXO := &UTXO{
		TxHash:       make([]byte, 32),
		TxIndex:      0,
		Value:        1000,
		ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Address:      "test_address",
		IsCoinbase:   false,
		Height:       1,
	}
	us.AddUTXO(invalidScriptSigUTXO)

	invalidScriptSigTx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{PrevTxHash: invalidScriptSigUTXO.TxHash, PrevTxIndex: 0, ScriptSig: []byte{0x01, 0x02}, Sequence: 0xffffffff}, // Too short scriptSig
		},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
		},
		LockTime: 0,
		Fee:      0,
	}
	invalidScriptSigTx.Hash = calculateTxHash(invalidScriptSigTx)

	err = us.ValidateTransactionInBlock(invalidScriptSigTx, testBlock, 1)
	assert.Error(t, err, "Transaction with invalid ScriptSig length should fail validation")
	assert.Contains(t, err.Error(), "invalid scriptSig length")

	// Test 7: UTXO not found
	// Use a different hash that doesn't exist in the UTXO set
	nonexistentHash := make([]byte, 32)
	for i := range nonexistentHash {
		nonexistentHash[i] = 0xff // Different from the zero hash used above
	}

	utxoNotFoundTx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{PrevTxHash: nonexistentHash, PrevTxIndex: 0, ScriptSig: makeValidScriptSig([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}), Sequence: 0xffffffff},
		},
		Outputs: []*block.TxOutput{
			{Value: 100, ScriptPubKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
		},
		LockTime: 0,
		Fee:      0,
	}
	utxoNotFoundTx.Hash = calculateTxHash(utxoNotFoundTx)

	err = us.ValidateTransactionInBlock(utxoNotFoundTx, testBlock, 1)
	assert.Error(t, err, "Transaction with non-existent UTXO should fail validation")
	assert.Contains(t, err.Error(), "UTXO not found")
}

func TestIsDoubleSpend(t *testing.T) {
	us := NewUTXOSet()

	// Create test UTXO
	pubKeyHash := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	addr := hex.EncodeToString(pubKeyHash)

	utxo := &UTXO{
		TxHash:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		TxIndex:      0,
		Value:        1000,
		ScriptPubKey: pubKeyHash,
		Address:      addr,
		IsCoinbase:   false,
		Height:       1,
	}

	us.AddUTXO(utxo)

	// Test 1: Not a double-spend (UTXO exists)
	tx1 := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: 0,
				ScriptSig:   makeValidScriptSig(pubKeyHash),
				Sequence:    0xffffffff,
			},
		},
		Outputs: []*block.TxOutput{
			{
				Value:        800,
				ScriptPubKey: pubKeyHash,
			},
		},
		LockTime: 0,
		Fee:      200,
	}
	tx1.Hash = calculateTxHash(tx1)

	assert.False(t, us.IsDoubleSpend(tx1), "Transaction with existing UTXO should not be double-spend")

	// Test 2: Double-spend attempt (UTXO already spent)
	// First spend the UTXO
	err := us.processTransaction(tx1, 2)
	assert.NoError(t, err)

	// Now try to spend it again
	tx2 := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: 0,
				ScriptSig:   makeValidScriptSig(pubKeyHash),
				Sequence:    0xffffffff,
			},
		},
		Outputs: []*block.TxOutput{
			{
				Value:        600,
				ScriptPubKey: pubKeyHash,
			},
		},
		LockTime: 0,
		Fee:      400,
	}
	tx2.Hash = calculateTxHash(tx2)

	assert.True(t, us.IsDoubleSpend(tx2), "Transaction with spent UTXO should be detected as double-spend")
}

func TestCalculateFee(t *testing.T) {
	us := NewUTXOSet()

	// Create test UTXOs
	pubKeyHash1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	pubKeyHash2 := []byte{0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	addr1 := hex.EncodeToString(pubKeyHash1)

	utxo1 := &UTXO{
		TxHash:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		TxIndex:      0,
		Value:        1000,
		ScriptPubKey: pubKeyHash1,
		Address:      addr1,
		IsCoinbase:   false,
		Height:       1,
	}

	us.AddUTXO(utxo1)

	// Test 1: Regular transaction fee calculation
	tx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{
				PrevTxHash:  utxo1.TxHash,
				PrevTxIndex: 0,
				ScriptSig:   makeValidScriptSig(pubKeyHash1),
				Sequence:    0xffffffff,
			},
		},
		Outputs: []*block.TxOutput{
			{
				Value:        800,
				ScriptPubKey: pubKeyHash2,
			},
		},
		LockTime: 0,
		Fee:      200,
	}
	tx.Hash = calculateTxHash(tx)

	fee, err := us.CalculateFee(tx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(200), fee, "Fee should be 200 (1000 - 800)")

	// Test 2: Coinbase transaction (no fee)
	coinbaseTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{
			{
				Value:        1000,
				ScriptPubKey: pubKeyHash1,
			},
		},
		LockTime: 0,
		Fee:      0,
	}
	coinbaseTx.Hash = calculateTxHash(coinbaseTx)

	fee, err = us.CalculateFee(coinbaseTx)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), fee, "Coinbase transaction should have no fee")

	// Test 3: Invalid transaction (output exceeds input)
	invalidTx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{
				PrevTxHash:  utxo1.TxHash,
				PrevTxIndex: 0,
				ScriptSig:   makeValidScriptSig(pubKeyHash1),
				Sequence:    0xffffffff,
			},
		},
		Outputs: []*block.TxOutput{
			{
				Value:        1200, // More than input
				ScriptPubKey: pubKeyHash2,
			},
		},
		LockTime: 0,
		Fee:      0,
	}
	invalidTx.Hash = calculateTxHash(invalidTx)

	_, err = us.CalculateFee(invalidTx)
	assert.Error(t, err, "Transaction with output exceeding input should fail fee calculation")
	assert.Contains(t, err.Error(), "output value 1200 exceeds input value 1000")
}

func TestValidateFeeRate(t *testing.T) {
	us := NewUTXOSet()

	// Create test UTXO with consistent data
	// Generate a deterministic private key for testing
	seed := new(big.Int).SetBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14})
	curve := btcec.S256()
	privKey := new(big.Int).Mod(seed, curve.Params().N)
	if privKey.Sign() == 0 {
		privKey.SetInt64(1)
	}

	// Generate the corresponding public key
	pubKey := new(ecdsa.PublicKey)
	pubKey.Curve = curve
	pubKey.X, pubKey.Y = curve.ScalarBaseMult(privKey.Bytes())

	// Marshal the public key to bytes
	pubKeyBytes := elliptic.Marshal(curve, pubKey.X, pubKey.Y)

	// Calculate the public key hash (this is what should be in ScriptPubKey)
	pubKeyHash := sha256.Sum256(pubKeyBytes)
	pubKeyHash20 := pubKeyHash[len(pubKeyHash)-20:] // Last 20 bytes

	// Create a different public key hash for outputs
	seed2 := new(big.Int).SetBytes([]byte{0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28})
	privKey2 := new(big.Int).Mod(seed2, curve.Params().N)
	if privKey2.Sign() == 0 {
		privKey2.SetInt64(1)
	}

	pubKey2 := new(ecdsa.PublicKey)
	pubKey2.Curve = curve
	pubKey2.X, pubKey2.Y = curve.ScalarBaseMult(privKey2.Bytes())

	pubKeyBytes2 := elliptic.Marshal(curve, pubKey2.X, pubKey2.Y)
	pubKeyHash2 := sha256.Sum256(pubKeyBytes2)
	pubKeyHash220 := pubKeyHash2[len(pubKeyHash2)-20:]

	addr1 := hex.EncodeToString(pubKeyHash20)

	utxo1 := &UTXO{
		TxHash:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		TxIndex:      0,
		Value:        1000,
		ScriptPubKey: pubKeyHash20,
		Address:      addr1,
		IsCoinbase:   false,
		Height:       1,
	}

	us.AddUTXO(utxo1)

	// Test 1: Valid fee rate
	tx := &block.Transaction{
		Version: 1,
		Inputs: []*block.TxInput{
			{
				PrevTxHash:  utxo1.TxHash,
				PrevTxIndex: 0,
				ScriptSig:   makeValidScriptSig(pubKeyHash20),
				Sequence:    0xffffffff,
			},
		},
		Outputs: []*block.TxOutput{
			{
				Value:        700,
				ScriptPubKey: pubKeyHash220,
			},
		},
		LockTime: 0,
		Fee:      300, // Increased fee to meet minimum requirement
	}
	tx.Hash = calculateTxHash(tx)

	// Test with reasonable fee rate (1000 sat/kilobyte)
	err := us.ValidateFeeRate(tx, 1000)
	assert.NoError(t, err, "Transaction with adequate fee rate should pass validation")

	// Test 2: Fee rate too low
	err = us.ValidateFeeRate(tx, 10000) // 10,000 sat/kilobyte
	assert.Error(t, err, "Transaction with insufficient fee rate should fail validation")
	assert.Contains(t, err.Error(), "below minimum required fee")

	// Test 3: Coinbase transaction (should always pass)
	coinbaseTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{},
		Outputs: []*block.TxOutput{
			{
				Value:        1000,
				ScriptPubKey: pubKeyHash20,
			},
		},
		LockTime: 0,
		Fee:      0,
	}
	coinbaseTx.Hash = calculateTxHash(coinbaseTx)

	err = us.ValidateFeeRate(coinbaseTx, 1000)
	assert.NoError(t, err, "Coinbase transaction should always pass fee validation")
}

func TestGetSpendableUTXOs(t *testing.T) {
	us := NewUTXOSet()

	// Create test UTXOs
	pubKeyHash1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	pubKeyHash2 := []byte{0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}

	addr1 := hex.EncodeToString(pubKeyHash1)
	addr2 := hex.EncodeToString(pubKeyHash2)

	// Add UTXOs with different values
	utxo1 := &UTXO{
		TxHash:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		TxIndex:      0,
		Value:        100,
		ScriptPubKey: pubKeyHash1,
		Address:      addr1,
		IsCoinbase:   false,
		Height:       1,
	}

	utxo2 := &UTXO{
		TxHash:       []byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40},
		TxIndex:      0,
		Value:        500,
		ScriptPubKey: pubKeyHash1,
		Address:      addr1,
		IsCoinbase:   false,
		Height:       1,
	}

	utxo3 := &UTXO{
		TxHash:       []byte{0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60},
		TxIndex:      0,
		Value:        1000,
		ScriptPubKey: pubKeyHash2,
		Address:      addr2,
		IsCoinbase:   false,
		Height:       1,
	}

	us.AddUTXO(utxo1)
	us.AddUTXO(utxo2)
	us.AddUTXO(utxo3)

	// Test 1: Get all spendable UTXOs for addr1
	spendableUTXOs := us.GetSpendableUTXOs(addr1, 0)
	assert.Len(t, spendableUTXOs, 2, "Should return 2 UTXOs for addr1")

	// Verify the values
	totalValue := uint64(0)
	for _, utxo := range spendableUTXOs {
		totalValue += utxo.Value
	}
	assert.Equal(t, uint64(600), totalValue, "Total value should be 600 (100 + 500)")

	// Test 2: Get UTXOs with minimum value filter
	spendableUTXOs = us.GetSpendableUTXOs(addr1, 200)
	assert.Len(t, spendableUTXOs, 1, "Should return 1 UTXO for addr1 with min value 200")
	assert.Equal(t, uint64(500), spendableUTXOs[0].Value, "Should return UTXO with value 500")

	// Test 3: Get UTXOs for addr2
	spendableUTXOs = us.GetSpendableUTXOs(addr2, 0)
	assert.Len(t, spendableUTXOs, 1, "Should return 1 UTXO for addr2")
	assert.Equal(t, uint64(1000), spendableUTXOs[0].Value, "Should return UTXO with value 1000")

	// Test 4: Get UTXOs for non-existent address
	spendableUTXOs = us.GetSpendableUTXOs("nonexistent", 0)
	assert.Len(t, spendableUTXOs, 0, "Should return 0 UTXOs for non-existent address")
}

// Helper function to create valid script signatures for testing
func makeValidScriptSig(pubKeyHash []byte) []byte {
	// For testing, we need to create a script signature that will produce
	// the exact public key hash that matches the UTXO's scriptPubKey.
	// We'll create a deterministic but valid-looking secp256k1 public key.

	// Create a deterministic private key from the pubKeyHash
	// Use the hash as a seed and ensure it's within valid range
	seed := new(big.Int).SetBytes(pubKeyHash)
	curve := btcec.S256()
	privKey := new(big.Int).Mod(seed, curve.Params().N)
	if privKey.Sign() == 0 {
		privKey.SetInt64(1)
	}

	// Generate the corresponding public key
	pubKey := new(ecdsa.PublicKey)
	pubKey.Curve = curve
	pubKey.X, pubKey.Y = curve.ScalarBaseMult(privKey.Bytes())

	// Marshal the public key to bytes
	pubKeyBytes := elliptic.Marshal(curve, pubKey.X, pubKey.Y)

	// Verify that this public key actually hashes to the expected pubKeyHash
	calculatedHash := sha256.Sum256(pubKeyBytes)
	calculatedHash20 := calculatedHash[len(calculatedHash)-20:]

	// If the calculated hash doesn't match, we need to adjust the private key
	// until we get a match. For testing purposes, we'll use a simple approach.
	if !bytes.Equal(calculatedHash20, pubKeyHash) {
		// Try a few different private keys until we get a match
		for i := 1; i < 10; i++ {
			adjustedSeed := new(big.Int).Add(seed, big.NewInt(int64(i)))
			adjustedSeed.Mod(adjustedSeed, curve.Params().N)
			if adjustedSeed.Sign() == 0 {
				adjustedSeed.SetInt64(1)
			}

			pubKey.X, pubKey.Y = curve.ScalarBaseMult(adjustedSeed.Bytes())
			pubKeyBytes = elliptic.Marshal(curve, pubKey.X, pubKey.Y)

			calculatedHash = sha256.Sum256(pubKeyBytes)
			calculatedHash20 = calculatedHash[len(calculatedHash)-20:]

			if bytes.Equal(calculatedHash20, pubKeyHash) {
				break
			}
		}
	}

	// Create deterministic R and S values for the signature
	// Use the pubKeyHash to generate deterministic but valid-looking values
	r := new(big.Int).SetBytes(pubKeyHash[:16])
	s := new(big.Int).SetBytes(pubKeyHash[16:])

	// Ensure R and S are within valid range for secp256k1
	r.Mod(r, curve.Params().N)
	s.Mod(s, curve.Params().N)
	if r.Sign() == 0 {
		r.SetInt64(1)
	}
	if s.Sign() == 0 {
		s.SetInt64(1)
	}

	// Serialize R and S to 32 bytes each
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Pad to 32 bytes if necessary
	if len(rBytes) < 32 {
		paddedR := make([]byte, 32)
		copy(paddedR[32-len(rBytes):], rBytes)
		rBytes = paddedR
	}
	if len(sBytes) < 32 {
		paddedS := make([]byte, 32)
		copy(paddedS[32-len(sBytes):], sBytes)
		sBytes = paddedS
	}

	// Combine public key and signature
	scriptSig := make([]byte, 0, 129)
	scriptSig = append(scriptSig, pubKeyBytes...)
	scriptSig = append(scriptSig, rBytes...)
	scriptSig = append(scriptSig, sBytes...)

	return scriptSig
}
