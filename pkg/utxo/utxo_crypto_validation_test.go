package utxo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/crypto_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHash creates a 32-byte hash from a string for testing
func makeHash(s string) []byte {
	hash := sha256.Sum256([]byte(s))
	return hash[:]
}

// createTestUTXO creates a UTXO with proper scriptPubKey format for testing
func createTestUTXO(hashStr string, index uint32, value uint64, keyPair *crypto_utils.TestKeyPair, isCoinbase bool, height uint64) *UTXO {
	scriptPubKey, _ := hex.DecodeString(keyPair.Address)
	return &UTXO{
		TxHash:       makeHash(hashStr),
		TxIndex:      index,
		Value:        value,
		ScriptPubKey: scriptPubKey, // Store the raw bytes, not hex string
		Address:      keyPair.Address,
		IsCoinbase:   isCoinbase,
		Height:       height,
	}
}

// TestValidateTransactionCompleteCoverage replaces the skipped test with real cryptographic validation
func TestValidateTransactionCompleteCoverage(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	// Test 1: Fee validation - actual fee less than specified fee
	t.Run("FeeValidation_ActualLessThanSpecified", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("fee_test_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create transaction where actual fee < specified fee
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
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 100) // Specified fee > actual fee (50)

		// This should fail at fee validation
		err := us.ValidateTransaction(tx)
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

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 0)

		// This should fail because output exceeds input
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Should fail when output exceeds input")
		assert.Contains(t, err.Error(), "output value 1200 exceeds input value 1000")
	})
}

// TestValidateTransactionInBlockCompleteCoverage replaces the skipped block validation test
func TestValidateTransactionInBlockCompleteCoverage(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	t.Run("BasicBlockValidation", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("block_test_hash", 0, 1000, alice, false, 1)
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

		outputs := []*block.TxOutput{
			{
				Value:        900,
				ScriptPubKey: []byte("output1"),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 100)

		// Create a coinbase transaction for the block
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

		// Create a mock block with coinbase + regular transaction
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
			Transactions: []*block.Transaction{coinbaseTx, tx}, // Coinbase at index 0, regular tx at index 1
		}

		// Validate transaction in block context (index 1, not 0, since 0 should be coinbase)
		err := us.ValidateTransactionInBlock(tx, mockBlock, 1)
		assert.NoError(t, err, "Valid transaction should pass block validation")
	})

	t.Run("CoinbaseValidation", func(t *testing.T) {
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

		// Validate coinbase transaction in block context
		err := us.ValidateTransactionInBlock(coinbaseTx, mockBlock, 0)
		assert.NoError(t, err, "Valid coinbase transaction should pass")
	})

	t.Run("InvalidTransactionIndex", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a valid transaction
		tx, err := ctu.CreateTestTransaction(alice, alice, 500, 100)
		require.NoError(t, err)

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
			Transactions: []*block.Transaction{tx},
		}

		// Try to validate with invalid transaction index
		err = us.ValidateTransactionInBlock(tx, mockBlock, 5) // Index 5 doesn't exist
		assert.Error(t, err, "Should fail with invalid transaction index")
		assert.Contains(t, err.Error(), "transaction index 5 out of bounds")
	})
}

// TestValidateTransactionBusinessLogic replaces the skipped business logic test
func TestValidateTransactionBusinessLogic(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	t.Run("CompleteBusinessLogicValidation", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()
		bob := ctu.GenerateTestKeyPair()

		// Create UTXOs for Alice
		utxo1 := createTestUTXO("business_test_1", 0, 5000, alice, false, 1)
		utxo2 := createTestUTXO("business_test_2", 0, 3000, alice, false, 1)
		us.AddUTXOSafe(utxo1)
		us.AddUTXOSafe(utxo2)

		// Create a complex multi-input, multi-output transaction
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

		outputs := []*block.TxOutput{
			{
				Value:        3000, // To Bob
				ScriptPubKey: []byte(bob.Address),
			},
			{
				Value:        4500, // Change back to Alice
				ScriptPubKey: []byte(alice.Address),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 500) // 8000 - 7500 = 500 fee

		// Validate complete business logic
		err := us.ValidateTransaction(tx)
		assert.NoError(t, err, "Complex transaction should pass all business logic validation")

		// Verify that the transaction follows all business rules
		assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs")
		assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs")
		assert.Equal(t, uint64(500), tx.Fee, "Should have correct fee")

		// Verify each input is properly signed
		for i, input := range tx.Inputs {
			assert.Len(t, input.ScriptSig, 65+64, "Input %d should have proper signature length", i)
			assert.NotEmpty(t, input.ScriptSig, "Input %d should have signature", i)
		}
	})

	t.Run("DoubleSpendPrevention", func(t *testing.T) {
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

		outputs1 := []*block.TxOutput{
			{
				Value:        900,
				ScriptPubKey: []byte(bob.Address),
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

		outputs2 := []*block.TxOutput{
			{
				Value:        800,
				ScriptPubKey: []byte(charlie.Address),
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
	})

	t.Run("SignatureVerificationFailure", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()
		bob := ctu.GenerateTestKeyPair()
		mallory := ctu.GenerateTestKeyPair() // Attacker

		// Create a UTXO for Alice
		utxo := createTestUTXO("sig_test_hash", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create transaction inputs
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
				Value:        900,
				ScriptPubKey: []byte(bob.Address),
			},
		}

		// Mallory tries to sign Alice's transaction (should fail)
		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: mallory, // Wrong key pair!
		}
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 100)

		// This should fail signature verification
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Transaction with invalid signature should fail")
		assert.Contains(t, err.Error(), "public key hash")
	})
}

// TestCryptographicEdgeCases tests various cryptographic edge cases
func TestCryptographicEdgeCases(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	us := NewUTXOSet()

	t.Run("MalformedSignature", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("malformed_sig_test", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create a transaction with malformed signature
		tx := &block.Transaction{
			Version: 1,
			Inputs: []*block.TxInput{
				{
					PrevTxHash:  utxo.TxHash,
					PrevTxIndex: utxo.TxIndex,
					ScriptSig:   []byte("malformed_signature"), // Invalid signature format
					Sequence:    0xffffffff,
				},
			},
			Outputs: []*block.TxOutput{
				{
					Value:        900,
					ScriptPubKey: []byte("output"),
				},
			},
			LockTime: 0,
			Fee:      100,
		}

		// This should fail due to malformed signature
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Transaction with malformed signature should fail")
		assert.Contains(t, err.Error(), "invalid scriptSig length")
	})

	t.Run("EmptyScriptSig", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()

		// Create a UTXO for Alice
		utxo := createTestUTXO("empty_sig_test", 0, 1000, alice, false, 1)
		us.AddUTXOSafe(utxo)

		// Create a transaction with empty signature
		tx := &block.Transaction{
			Version: 1,
			Inputs: []*block.TxInput{
				{
					PrevTxHash:  utxo.TxHash,
					PrevTxIndex: utxo.TxIndex,
					ScriptSig:   []byte{}, // Empty signature
					Sequence:    0xffffffff,
				},
			},
			Outputs: []*block.TxOutput{
				{
					Value:        900,
					ScriptPubKey: []byte("output"),
				},
			},
			LockTime: 0,
			Fee:      100,
		}

		// This should fail due to empty signature
		err := us.ValidateTransaction(tx)
		assert.Error(t, err, "Transaction with empty signature should fail")
		assert.Contains(t, err.Error(), "invalid scriptSig length")
	})

	t.Run("MaximumValidTransaction", func(t *testing.T) {
		alice := ctu.GenerateTestKeyPair()
		bob := ctu.GenerateTestKeyPair()

		// Create multiple UTXOs for testing maximum inputs
		var utxos []*UTXO
		for i := 0; i < 10; i++ {
			utxo := createTestUTXO(fmt.Sprintf("max_test_%d", i), 0, 1000, alice, false, 1)
			us.AddUTXOSafe(utxo)
			utxos = append(utxos, utxo)
		}

		// Create transaction with many inputs
		var inputs []*block.TxInput
		for _, utxo := range utxos {
			inputs = append(inputs, &block.TxInput{
				PrevTxHash:  utxo.TxHash,
				PrevTxIndex: utxo.TxIndex,
				ScriptSig:   []byte{},
				Sequence:    0xffffffff,
			})
		}

		outputs := []*block.TxOutput{
			{
				Value:        9000, // 10000 - 1000 fee
				ScriptPubKey: []byte(bob.Address),
			},
		}

		keyPairs := map[string]*crypto_utils.TestKeyPair{
			alice.Address: alice,
		}
		tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, 1000)

		// This should pass with many valid inputs
		err := us.ValidateTransaction(tx)
		assert.NoError(t, err, "Transaction with many valid inputs should pass")
		assert.Equal(t, 10, len(tx.Inputs), "Should have 10 inputs")

		// Verify all inputs are properly signed
		for i, input := range tx.Inputs {
			assert.Len(t, input.ScriptSig, 65+64, "Input %d should have proper signature length", i)
		}
	})
}
