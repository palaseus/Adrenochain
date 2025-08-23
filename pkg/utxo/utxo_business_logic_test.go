package utxo

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/crypto_utils"
	"github.com/stretchr/testify/assert"
)

// TestBusinessLogicValidationWithRealCrypto tests business logic validation with properly signed transactions
func TestBusinessLogicValidationWithRealCrypto(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	// Test 1: Fee validation - actual fee less than specified fee
	t.Run("FeeValidation_ActualLessThanSpecified", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("fee_test_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create properly signed transaction where actual fee < specified fee
		inputs := []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			},
		}

		outputs := []*block.TxOutput{
			{
				Value:        950, // 1000 - 950 = 50 actual fee
				ScriptPubKey: []byte("output1"),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}

		// Create the transaction with the correct fee (50) first to get a valid signature
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 50)

		// Now modify the specified fee to be higher (this should cause validation failure)
		tx.Fee = 100 // Specified fee > actual fee (50)

		// This should fail at fee validation using business logic validation (skips signature checks)
		err := us.ValidateTransactionBusinessLogic(tx)
		assert.Error(t, err, "Should fail fee validation")
		assert.Contains(t, err.Error(), "actual fee 50 is less than specified fee 100")
	})

	// Test 2: High fee protection - fee > 50% of input value
	t.Run("FeeValidation_UnreasonablyHighFee", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("high_fee_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create transaction with very high fee (>50% of input)
		inputs := []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			},
		}

		outputs := []*block.TxOutput{
			{
				Value:        300, // 1000 - 300 = 700 fee (70% > 50%)
				ScriptPubKey: []byte("output1"),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}

		// Create properly signed transaction with high fee
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 700)

		// This should fail at high fee protection
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Should fail high fee protection")
		assert.Contains(t, err.Error(), "fee 700 is unreasonably high (more than 50% of input value 1000)")
	})

	// Test 3: Dust threshold validation
	t.Run("DustThresholdValidation", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("dust_test_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create transaction with dust output (<546 satoshis)
		inputs := []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			},
		}

		outputs := []*block.TxOutput{
			{
				Value:        500, // 500 < 546 (dust threshold)
				ScriptPubKey: []byte("output1"),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}

		// Create properly signed transaction with dust output
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 500)

		// This should fail at dust threshold validation
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Should fail dust threshold validation")
		assert.Contains(t, err.Error(), "dust threshold")
	})

	// Test 4: Output value exceeds input value
	t.Run("OutputExceedsInput", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("exceed_test_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create transaction where output value > input value
		inputs := []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			},
		}

		outputs := []*block.TxOutput{
			{
				Value:        1200, // 1200 > 1000 input
				ScriptPubKey: []byte("output1"),
			},
		}

		// Note: This will fail during signing because output > input,
		// but let's create it anyway to test validation
		tx := &block.Transaction{
			Version:  1,
			Inputs:   inputs,
			Outputs:  outputs,
			LockTime: 0,
			Fee:      0,
		}

		// Create a fake signature for testing (this will fail crypto validation, but that's expected)
		fakeSignature := make([]byte, 65+64)
		copy(fakeSignature[:65], alice.PublicKey.SerializeUncompressed())
		// Fill signature part with zeros (invalid but proper length)
		for i := 65; i < len(fakeSignature); i++ {
			fakeSignature[i] = 0
		}
		tx.Inputs[0].ScriptSig = fakeSignature
		tx.Hash = []byte("fake_hash_for_testing_business_logic")

		// This should fail because output exceeds input using business logic validation (skips signature checks)
		err := us.ValidateTransactionBusinessLogic(tx)
		assert.Error(t, err, "Should fail when output exceeds input")
		assert.Contains(t, err.Error(), "output value 1200 exceeds input value 1000")
	})

	// Test 5: Valid transaction (should pass all validations)
	t.Run("ValidTransaction", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()
		bob := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("valid_test_hash", 0, 10000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create a valid transaction: Alice sends 8000 to Bob with 1000 fee
		inputs := []*block.TxInput{
			{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			},
		}

		// Decode Bob's address to get proper ScriptPubKey
		bobScriptPubKey, _ := hex.DecodeString(bob.Address)
		aliceScriptPubKey, _ := hex.DecodeString(alice.Address)

		outputs := []*block.TxOutput{
			{
				Value:        8000,
				ScriptPubKey: bobScriptPubKey,
			},
			{
				Value:        1000, // Change back to Alice
				ScriptPubKey: aliceScriptPubKey,
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}

		// Create properly signed transaction
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 1000)

		// This should pass all validations
		err := us.ValidateTransaction(tx)
		assert.NoError(t, err, "Valid transaction should pass all validations")
	})
}

// TestDoubleSpendPrevention tests double spend detection with real cryptography
func TestDoubleSpendPrevention(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	alice := ctu.GenerateTestKeyPair()
	bob := ctu.GenerateTestKeyPair()
	charlie := ctu.GenerateTestKeyPair()

	// Create a UTXO for Alice
	utxo := createTestUTXO("double_spend_test", 0, 1000, alice, false, 1)
	us.AddUTXOSafe(utxo)

	// Create first transaction: Alice sends to Bob
	inputs1 := []*block.TxInput{
		{
			PrevTxHash:  utxo.TxHash,
			PrevTxIndex: utxo.TxIndex,
			ScriptSig:   []byte{},
			Sequence:    0xffffffff,
		},
	}

	bobScriptPubKey, _ := hex.DecodeString(bob.Address)
	outputs1 := []*block.TxOutput{
		{
			Value:        900,
			ScriptPubKey: bobScriptPubKey,
		},
	}

	keyPairs := map[string]*crypto_utils.TestKeyPair{
		alice.Address: alice,
	}
	tx1 := ctu.CreateSignedTransaction(inputs1, outputs1, keyPairs, 100)

	// Create second transaction: Alice tries to send same UTXO to Charlie (double spend)
	inputs2 := []*block.TxInput{
		{
			PrevTxHash:  utxo.TxHash,
			PrevTxIndex: utxo.TxIndex,
			ScriptSig:   []byte{},
			Sequence:    0xffffffff,
		},
	}

	charlieScriptPubKey, _ := hex.DecodeString(charlie.Address)
	outputs2 := []*block.TxOutput{
		{
			Value:        800,
			ScriptPubKey: charlieScriptPubKey,
		},
	}

	tx2 := ctu.CreateSignedTransaction(inputs2, outputs2, keyPairs, 200)

	// First transaction should be valid
	err1 := us.ValidateTransaction(tx1)
	assert.NoError(t, err1, "First transaction should be valid")

	// Process first transaction
	us.processTransaction(tx1, 1)

	// Second transaction should fail (double spend)
	err2 := us.ValidateTransaction(tx2)
	assert.Error(t, err2, "Second transaction should fail due to double spend")
	assert.Contains(t, err2.Error(), "UTXO not found")
}

// TestSignatureTampering tests that tampered signatures are detected
func TestSignatureTampering(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	alice := ctu.GenerateTestKeyPair()
	bob := ctu.GenerateTestKeyPair()

	// Create a UTXO for Alice
	utxo := createTestUTXO("tamper_test_hash", 0, 10000, alice, false, 1)
	us.AddUTXOSafe(utxo)

	// Create a valid transaction
	inputs := []*block.TxInput{
		{
			PrevTxHash:  utxo.TxHash,
			PrevTxIndex: utxo.TxIndex,
			ScriptSig:   []byte{},
			Sequence:    0xffffffff,
		},
	}

	bobScriptPubKey, _ := hex.DecodeString(bob.Address)
	outputs := []*block.TxOutput{
		{
			Value:        9000,
			ScriptPubKey: bobScriptPubKey,
		},
	}

	keyPairs := map[string]*crypto_utils.TestKeyPair{
		alice.Address: alice,
	}
	tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 1000)

	// Verify original transaction is valid
	err := us.ValidateTransaction(tx)
	assert.NoError(t, err, "Original transaction should be valid")

	// Tamper with the signature (change one byte)
	if len(tx.Inputs[0].ScriptSig) > 65 {
		tx.Inputs[0].ScriptSig[65] ^= 0x01 // Flip one bit in the signature
	}

	// Verify tampered transaction is invalid
	err = us.ValidateTransaction(tx)
	assert.Error(t, err, "Tampered transaction should fail validation")
	assert.Contains(t, err.Error(), "invalid signature")
}

// TestMultiInputTransaction tests validation with multiple inputs
func TestMultiInputTransaction(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	alice := ctu.GenerateTestKeyPair()
	bob := ctu.GenerateTestKeyPair()

	// Create multiple UTXOs for Alice
	utxo1 := createTestUTXO("multi_input_1", 0, 5000, alice, false, 1)
	utxo2 := createTestUTXO("multi_input_2", 0, 3000, alice, false, 1)
	us.AddUTXOSafe(utxo1)
	us.AddUTXOSafe(utxo2)

	// Create transaction with multiple inputs
	inputs := []*block.TxInput{
		{
			PrevTxHash:  utxo1.TxHash,
			PrevTxIndex: utxo1.TxIndex,
			ScriptSig:   []byte{},
			Sequence:    0xffffffff,
		},
		{
			PrevTxHash:  utxo2.TxHash,
			PrevTxIndex: utxo2.TxIndex,
			ScriptSig:   []byte{},
			Sequence:    0xffffffff,
		},
	}

	bobScriptPubKey, _ := hex.DecodeString(bob.Address)
	outputs := []*block.TxOutput{
		{
			Value:        7000, // 5000 + 3000 - 1000 = 7000
			ScriptPubKey: bobScriptPubKey,
		},
	}

	keyPairs := map[string]*crypto_utils.TestKeyPair{
		alice.Address: alice,
	}
	tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 1000)

	// Validate the multi-input transaction
	err := us.ValidateTransaction(tx)
	assert.NoError(t, err, "Multi-input transaction should be valid")

	// Verify both inputs were properly signed
	for i, input := range tx.Inputs {
		assert.Len(t, input.ScriptSig, 65+64, "Input %d should have proper signature length", i)
	}
}

// TestCoinbaseTransactionValidation tests coinbase transaction validation
func TestCoinbaseTransactionValidation(t *testing.T) {
	us := NewUTXOSet()

	// Create a coinbase transaction (no inputs)
	coinbaseTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{}, // No inputs for coinbase
		Outputs: []*block.TxOutput{
			{
				Value:        5000000000, // 50 BTC reward
				ScriptPubKey: []byte("miner_address"),
			},
		},
		LockTime: 0,
		Fee:      0,
	}

	// Create a mock block
	mockBlock := &block.Block{
		Header: &block.Header{
			Version:       1,
			Height:        2,
			Timestamp:     time.Unix(1234567890, 0),
			Difficulty:    1000,
			Nonce:         0,
			MerkleRoot:    make([]byte, 32),
			PrevBlockHash: make([]byte, 32),
		},
		Transactions: []*block.Transaction{coinbaseTx},
	}

	// Validate coinbase transaction in block context (should pass)
	err := us.ValidateTransactionInBlock(coinbaseTx, mockBlock, 0)
	assert.NoError(t, err, "Valid coinbase transaction should pass")

	// Test coinbase validation outside block context (should also pass)
	err = us.ValidateTransaction(coinbaseTx)
	assert.NoError(t, err, "Coinbase transaction should be valid outside block context too")
}
