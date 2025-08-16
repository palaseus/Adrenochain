package block

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAdvancedTransactionScenarios tests complex transaction scenarios
func TestAdvancedTransactionScenarios(t *testing.T) {
	t.Run("TestMultiInputMultiOutputTransaction", func(t *testing.T) {
		// Create a transaction with multiple inputs and outputs
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("script1"),
				Sequence:    0xffffffff,
			},
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 1,
				ScriptSig:   []byte("script2"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("prev_tx_1"))
		copy(inputs[1].PrevTxHash, []byte("prev_tx_2"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("output1")},
			{Value: 500, ScriptPubKey: []byte("output2")},
			{Value: 250, ScriptPubKey: []byte("output3")},
		}

		tx := NewTransaction(inputs, outputs, 50)
		assert.NotNil(t, tx)
		assert.Equal(t, 2, len(tx.Inputs))
		assert.Equal(t, 3, len(tx.Outputs))
		assert.Equal(t, uint64(50), tx.Fee)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err)

		// Verify hash calculation
		hash1 := tx.CalculateHash()
		hash2 := tx.CalculateHash()
		assert.Equal(t, hash1, hash2, "Hash should be deterministic")
		assert.Equal(t, tx.Hash, hash1, "Transaction hash should match calculated hash")
	})

	t.Run("TestTransactionWithHighValues", func(t *testing.T) {
		// Test transaction with very large values
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("high_value_script"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("high_value_input"))

		// Use maximum uint64 values
		maxValue := ^uint64(0) // Maximum uint64 value
		outputs := []*TxOutput{
			{Value: maxValue, ScriptPubKey: []byte("max_output")},
		}

		tx := NewTransaction(inputs, outputs, maxValue)
		assert.NotNil(t, tx)
		assert.Equal(t, maxValue, tx.Outputs[0].Value)
		assert.Equal(t, maxValue, tx.Fee)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err)

		// Verify hash calculation works with large values
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")
	})

	t.Run("TestTransactionWithEmptyScripts", func(t *testing.T) {
		// Test transaction with empty script signatures (should still be valid)
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte{}, // Empty script
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("empty_script_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("output_script")},
		}

		tx := NewTransaction(inputs, outputs, 10)
		assert.NotNil(t, tx)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with empty script should be valid")
	})

	t.Run("TestTransactionWithSpecialSequences", func(t *testing.T) {
		// Test different sequence values
		testCases := []uint32{
			0x00000000, // Final transaction
			0xffffffff, // Replace-by-fee enabled
			0x00000001, // Locktime enabled
			0x00000002, // Locktime enabled
		}

		for _, sequence := range testCases {
			t.Run(fmt.Sprintf("Sequence_%08x", sequence), func(t *testing.T) {
				inputs := []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte("sequence_test"),
						Sequence:    sequence,
					},
				}
				copy(inputs[0].PrevTxHash, []byte(fmt.Sprintf("seq_%08x", sequence)))

				outputs := []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("output")},
				}

				tx := NewTransaction(inputs, outputs, 10)
				assert.NotNil(t, tx)
				assert.Equal(t, sequence, tx.Inputs[0].Sequence)

				// Validate transaction
				err := tx.IsValid()
				assert.NoError(t, err, "Transaction with sequence %08x should be valid", sequence)
			})
		}
	})
}

// TestTransactionSerializationEdgeCasesAdvanced tests edge cases in transaction serialization
func TestTransactionSerializationEdgeCasesAdvanced(t *testing.T) {
	t.Run("TestTransactionWithNilFields", func(t *testing.T) {
		// Test transaction with nil inputs and outputs (should handle gracefully)
		tx := &Transaction{
			Version:  1,
			Inputs:   nil,
			Outputs:  nil,
			LockTime: 0,
			Fee:      0,
			Hash:     make([]byte, 32),
		}

		// This should fail validation but not crash
		err := tx.IsValid()
		assert.Error(t, err, "Transaction with nil outputs should fail validation")

		// Hash calculation should not panic
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash calculation should work with nil fields")
	})

	t.Run("TestTransactionWithZeroValues", func(t *testing.T) {
		// Test transaction with zero values in various fields
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("zero_test"),
				Sequence:    0,
			},
		}

		outputs := []*TxOutput{
			{Value: 1, ScriptPubKey: []byte("min_value")}, // Minimum valid value
		}

		tx := NewTransaction(inputs, outputs, 0) // Zero fee
		assert.NotNil(t, tx)
		assert.Equal(t, uint64(0), tx.Fee)
		assert.Equal(t, uint32(0), tx.Inputs[0].Sequence)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with zero fee and sequence should be valid")
	})

	t.Run("TestTransactionWithMaxLockTime", func(t *testing.T) {
		// Test transaction with maximum locktime
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("max_locktime"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("max_locktime_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("output")},
		}

		maxLockTime := ^uint64(0) // Maximum uint64 value
		tx := NewTransaction(inputs, outputs, 10)
		tx.LockTime = maxLockTime

		assert.NotNil(t, tx)
		assert.Equal(t, maxLockTime, tx.LockTime)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with maximum locktime should be valid")

		// Verify hash calculation works with max locktime
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")
	})
}

// TestTransactionPerformance tests transaction performance characteristics
func TestTransactionPerformance(t *testing.T) {
	t.Run("TestLargeTransactionHashCalculation", func(t *testing.T) {
		// Create a transaction with many inputs and outputs to test performance
		numInputs := 100
		numOutputs := 50

		inputs := make([]*TxInput, numInputs)
		for i := 0; i < numInputs; i++ {
			inputs[i] = &TxInput{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: uint32(i),
				ScriptSig:   []byte(fmt.Sprintf("script_%d", i)),
				Sequence:    0xffffffff,
			}
			copy(inputs[i].PrevTxHash, []byte(fmt.Sprintf("input_%d", i)))
		}

		outputs := make([]*TxOutput, numOutputs)
		for i := 0; i < numOutputs; i++ {
			outputs[i] = &TxOutput{
				Value:        uint64(1000 + i),
				ScriptPubKey: []byte(fmt.Sprintf("output_%d", i)),
			}
		}

		tx := NewTransaction(inputs, outputs, 100)
		assert.NotNil(t, tx)
		assert.Equal(t, numInputs, len(tx.Inputs))
		assert.Equal(t, numOutputs, len(tx.Outputs))

		// Time the hash calculation
		start := time.Now()
		hash := tx.CalculateHash()
		duration := time.Since(start)

		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")
		assert.Less(t, duration, 100*time.Millisecond, "Hash calculation should be fast")

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Large transaction should be valid")
	})

	t.Run("TestTransactionHashUniqueness", func(t *testing.T) {
		// Test that different transactions produce different hashes
		hashes := make(map[string]bool)
		numTransactions := 1000

		for i := 0; i < numTransactions; i++ {
			inputs := []*TxInput{
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: uint32(i),
					ScriptSig:   []byte(fmt.Sprintf("unique_%d", i)),
					Sequence:    0xffffffff,
				},
			}
			copy(inputs[0].PrevTxHash, []byte(fmt.Sprintf("unique_input_%d", i)))

			outputs := []*TxOutput{
				{Value: uint64(1000 + i), ScriptPubKey: []byte(fmt.Sprintf("unique_output_%d", i))},
			}

			tx := NewTransaction(inputs, outputs, uint64(i))
			hash := tx.CalculateHash()
			hashHex := hex.EncodeToString(hash)

			// Check for hash collisions
			if hashes[hashHex] {
				t.Errorf("Hash collision detected for transaction %d: %s", i, hashHex)
			}
			hashes[hashHex] = true
		}

		assert.Equal(t, numTransactions, len(hashes), "All transactions should have unique hashes")
	})
}

// TestTransactionSecurity tests security aspects of transactions
func TestTransactionSecurity(t *testing.T) {
	t.Run("TestTransactionHashTampering", func(t *testing.T) {
		// Test that changing transaction data invalidates the hash
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("security_test"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("security_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("output")},
		}

		tx := NewTransaction(inputs, outputs, 10)
		originalHash := tx.CalculateHash()
		assert.Equal(t, originalHash, tx.Hash, "Original hash should match calculated hash")

		// Tamper with the transaction
		tx.Outputs[0].Value = 999
		newHash := tx.CalculateHash()

		// Hash should change
		assert.NotEqual(t, originalHash, newHash, "Hash should change when transaction is tampered with")
		assert.NotEqual(t, tx.Hash, newHash, "Stored hash should not match new calculated hash")
	})

	t.Run("TestTransactionInputValidation", func(t *testing.T) {
		// Test various invalid input scenarios
		testCases := []struct {
			name        string
			input       *TxInput
			shouldError bool
		}{
			{
				name: "Valid input",
				input: &TxInput{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: 0,
					ScriptSig:   []byte("valid"),
					Sequence:    0xffffffff,
				},
				shouldError: false,
			},
			{
				name: "Invalid hash length",
				input: &TxInput{
					PrevTxHash:  make([]byte, 16), // Wrong length
					PrevTxIndex: 0,
					ScriptSig:   []byte("invalid"),
					Sequence:    0xffffffff,
				},
				shouldError: true,
			},
			{
				name: "Nil hash",
				input: &TxInput{
					PrevTxHash:  nil,
					PrevTxIndex: 0,
					ScriptSig:   []byte("nil_hash"),
					Sequence:    0xffffffff,
				},
				shouldError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.input.IsValid()
				if tc.shouldError {
					assert.Error(t, err, "Input should fail validation: %s", tc.name)
				} else {
					assert.NoError(t, err, "Input should pass validation: %s", tc.name)
				}
			})
		}
	})

	t.Run("TestTransactionOutputValidation", func(t *testing.T) {
		// Test various invalid output scenarios
		testCases := []struct {
			name        string
			output      *TxOutput
			shouldError bool
		}{
			{
				name: "Valid output",
				output: &TxOutput{
					Value:        1000,
					ScriptPubKey: []byte("valid"),
				},
				shouldError: false,
			},
			{
				name: "Zero value",
				output: &TxOutput{
					Value:        0,
					ScriptPubKey: []byte("zero_value"),
				},
				shouldError: true,
			},
			{
				name: "Empty script",
				output: &TxOutput{
					Value:        1000,
					ScriptPubKey: []byte{},
				},
				shouldError: true,
			},
			{
				name: "Nil script",
				output: &TxOutput{
					Value:        1000,
					ScriptPubKey: nil,
				},
				shouldError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.output.IsValid()
				if tc.shouldError {
					assert.Error(t, err, "Output should fail validation: %s", tc.name)
				} else {
					assert.NoError(t, err, "Output should pass validation: %s", tc.name)
				}
			})
		}
	})
}

// TestTransactionIntegration tests integration between different transaction components
func TestTransactionIntegration(t *testing.T) {
	t.Run("TestTransactionInBlock", func(t *testing.T) {
		// Test adding a transaction to a block
		block := NewBlock(make([]byte, 32), 1, 1000)
		assert.Equal(t, 0, len(block.Transactions), "New block should have no transactions")

		// Create and add a transaction
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("block_test"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("block_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("block_output")},
		}

		tx := NewTransaction(inputs, outputs, 10)
		block.AddTransaction(tx)

		assert.Equal(t, 1, len(block.Transactions), "Block should have one transaction")
		assert.Equal(t, tx, block.Transactions[0], "Transaction should be added to block")

		// Verify Merkle root is updated
		merkleRoot := block.CalculateMerkleRoot()
		assert.NotNil(t, merkleRoot, "Merkle root should be calculated")
		assert.Equal(t, merkleRoot, block.Header.MerkleRoot, "Block header should have updated Merkle root")
	})

	t.Run("TestTransactionSerializationRoundTrip", func(t *testing.T) {
		// Test that transaction serialization and deserialization work correctly
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("serialization_test"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("serialization_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("serialization_output")},
		}

		originalTx := NewTransaction(inputs, outputs, 10)
		assert.NotNil(t, originalTx)

		// Serialize
		data, err := originalTx.Serialize()
		assert.NoError(t, err, "Transaction should serialize successfully")
		assert.NotNil(t, data, "Serialized data should not be nil")

		// Deserialize
		deserializedTx := &Transaction{}
		err = deserializedTx.Deserialize(data)
		assert.NoError(t, err, "Transaction should deserialize successfully")

		// Verify all fields match
		assert.Equal(t, originalTx.Version, deserializedTx.Version)
		assert.Equal(t, originalTx.LockTime, deserializedTx.LockTime)
		assert.Equal(t, originalTx.Fee, deserializedTx.Fee)
		assert.Equal(t, len(originalTx.Inputs), len(deserializedTx.Inputs))
		assert.Equal(t, len(originalTx.Outputs), len(deserializedTx.Outputs))

		// Verify inputs
		for i, input := range originalTx.Inputs {
			assert.Equal(t, input.PrevTxIndex, deserializedTx.Inputs[i].PrevTxIndex)
			assert.Equal(t, input.Sequence, deserializedTx.Inputs[i].Sequence)
			assert.Equal(t, input.PrevTxHash, deserializedTx.Inputs[i].PrevTxHash)
			assert.Equal(t, input.ScriptSig, deserializedTx.Inputs[i].ScriptSig)
		}

		// Verify outputs
		for i, output := range originalTx.Outputs {
			assert.Equal(t, output.Value, deserializedTx.Outputs[i].Value)
			assert.Equal(t, output.ScriptPubKey, deserializedTx.Outputs[i].ScriptPubKey)
		}

		// Verify hash
		assert.Equal(t, originalTx.Hash, deserializedTx.Hash)
	})
}
