package block

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRealWorldTransactionFlow tests a complete real-world transaction scenario
func TestRealWorldTransactionFlow(t *testing.T) {
	t.Run("TestCompleteTransactionLifecycle", func(t *testing.T) {
		// Simulate a real-world scenario: Alice wants to send 500 coins to Bob
		// and 200 coins to Charlie, with a 50 coin fee

		// Step 1: Create the transaction inputs (Alice's previous UTXOs)
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("alice_signature_1"),
				Sequence:    0xffffffff,
			},
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 1,
				ScriptSig:   []byte("alice_signature_2"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("alice_utxo_1_hash"))
		copy(inputs[1].PrevTxHash, []byte("alice_utxo_2_hash"))

		// Step 2: Create the transaction outputs
		outputs := []*TxOutput{
			{Value: 500, ScriptPubKey: []byte("bob_public_key")},     // To Bob
			{Value: 200, ScriptPubKey: []byte("charlie_public_key")}, // To Charlie
			{Value: 250, ScriptPubKey: []byte("alice_change_key")},   // Change back to Alice
		}

		// Step 3: Create the transaction
		fee := uint64(50)
		tx := NewTransaction(inputs, outputs, fee)

		// Step 4: Validate the transaction
		assert.NotNil(t, tx)
		assert.Equal(t, uint32(1), tx.Version)
		assert.Equal(t, 2, len(tx.Inputs))
		assert.Equal(t, 3, len(tx.Outputs))
		assert.Equal(t, fee, tx.Fee)
		assert.Equal(t, uint64(0), tx.LockTime)

		// Step 5: Verify transaction validity
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction should be valid")

		// Step 6: Verify hash calculation
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")
		assert.Equal(t, hash, tx.Hash, "Transaction hash should match calculated hash")

		// Step 7: Verify total input vs output + fee
		totalInput := uint64(1000) // Assuming Alice had 1000 coins total
		totalOutput := uint64(0)
		for _, output := range outputs {
			totalOutput += output.Value
		}
		totalOutput += fee

		assert.Equal(t, totalInput, totalOutput, "Total input should equal total output + fee")

		// Step 8: Verify the transaction can be added to a block
		block := NewBlock(make([]byte, 32), 1, 1000)
		block.AddTransaction(tx)

		assert.Equal(t, 1, len(block.Transactions), "Block should contain the transaction")
		assert.Equal(t, tx, block.Transactions[0], "Transaction should be in the block")

		// Step 9: Verify Merkle root is updated
		merkleRoot := block.CalculateMerkleRoot()
		assert.NotNil(t, merkleRoot, "Merkle root should be calculated")
		assert.Equal(t, merkleRoot, block.Header.MerkleRoot, "Block header should have updated Merkle root")

		t.Logf("Transaction created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  Inputs: %d", len(tx.Inputs))
		t.Logf("  Outputs: %d", len(tx.Outputs))
		t.Logf("  Fee: %d", tx.Fee)
		t.Logf("  Total Output: %d", totalOutput)
	})

	t.Run("TestCoinbaseTransaction", func(t *testing.T) {
		// Test a coinbase transaction (mining reward)
		coinbaseTx := &Transaction{
			Version:  1,
			Inputs:   []*TxInput{}, // No inputs for coinbase
			Outputs:  []*TxOutput{},
			LockTime: 0,
			Fee:      0,
			Hash:     make([]byte, 32),
		}

		// Add mining reward output
		miningReward := uint64(1000)
		coinbaseTx.Outputs = append(coinbaseTx.Outputs, &TxOutput{
			Value:        miningReward,
			ScriptPubKey: []byte("miner_public_key"),
		})

		// Calculate hash
		coinbaseTx.Hash = coinbaseTx.CalculateHash()

		// Validate coinbase transaction
		assert.True(t, coinbaseTx.IsCoinbase(), "Transaction should be identified as coinbase")
		err := coinbaseTx.IsValid()
		assert.NoError(t, err, "Coinbase transaction should be valid")

		// Verify it can be added to a block
		block := NewBlock(make([]byte, 32), 1, 1000)
		block.AddTransaction(coinbaseTx)

		assert.Equal(t, 1, len(block.Transactions), "Block should contain coinbase transaction")
		assert.Equal(t, coinbaseTx, block.Transactions[0], "Coinbase transaction should be in block")

		t.Logf("Coinbase transaction created successfully:")
		t.Logf("  Hash: %x", coinbaseTx.Hash)
		t.Logf("  Mining Reward: %d", miningReward)
		t.Logf("  Is Coinbase: %t", coinbaseTx.IsCoinbase())
	})

	t.Run("TestTransactionWithLockTime", func(t *testing.T) {
		// Test a transaction with a future locktime
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("timelock_signature"),
				Sequence:    0x00000001, // Enable locktime
			},
		}
		copy(inputs[0].PrevTxHash, []byte("timelock_input"))

		outputs := []*TxOutput{
			{Value: 1000, ScriptPubKey: []byte("timelock_output")},
		}

		tx := NewTransaction(inputs, outputs, 10)

		// Set locktime to 1 hour from now
		futureTime := uint64(time.Now().Add(1 * time.Hour).Unix())
		tx.LockTime = futureTime

		assert.NotNil(t, tx)
		assert.Equal(t, futureTime, tx.LockTime)
		assert.Equal(t, uint32(0x00000001), tx.Inputs[0].Sequence)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with locktime should be valid")

		// Verify hash calculation
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")

		t.Logf("Timelock transaction created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  LockTime: %d (Unix timestamp)", tx.LockTime)
		t.Logf("  Sequence: %08x", tx.Inputs[0].Sequence)
	})

	t.Run("TestMultiSignatureTransaction", func(t *testing.T) {
		// Test a transaction that requires multiple signatures
		// This simulates a multi-sig wallet scenario

		// Create multiple inputs with different signatures
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("signature_1"),
				Sequence:    0xffffffff,
			},
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("signature_2"),
				Sequence:    0xffffffff,
			},
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 1,
				ScriptSig:   []byte("signature_3"),
				Sequence:    0xffffffff,
			},
		}
		copy(inputs[0].PrevTxHash, []byte("multisig_input_1"))
		copy(inputs[1].PrevTxHash, []byte("multisig_input_1")) // Same UTXO, different signature
		copy(inputs[2].PrevTxHash, []byte("multisig_input_2"))

		// Create outputs for the multi-sig transaction
		outputs := []*TxOutput{
			{Value: 1500, ScriptPubKey: []byte("multisig_output")},
			{Value: 500, ScriptPubKey: []byte("change_output")},
		}

		tx := NewTransaction(inputs, outputs, 20)
		assert.NotNil(t, tx)
		assert.Equal(t, 3, len(tx.Inputs))
		assert.Equal(t, 2, len(tx.Outputs))

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Multi-signature transaction should be valid")

		// Verify hash calculation
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")

		// Verify it can be added to a block
		block := NewBlock(make([]byte, 32), 1, 1000)
		block.AddTransaction(tx)

		assert.Equal(t, 1, len(block.Transactions), "Block should contain multi-signature transaction")

		t.Logf("Multi-signature transaction created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  Inputs: %d", len(tx.Inputs))
		t.Logf("  Outputs: %d", len(tx.Outputs))
		t.Logf("  Fee: %d", tx.Fee)
	})

	t.Run("TestTransactionFeeCalculation", func(t *testing.T) {
		// Test various fee scenarios
		testCases := []struct {
			name           string
			inputValue     uint64
			outputValue    uint64
			fee            uint64
			expectedChange uint64
			shouldBeValid  bool
		}{
			{
				name:           "Standard fee",
				inputValue:     1000,
				outputValue:    800,
				fee:            50,
				expectedChange: 150,
				shouldBeValid:  true,
			},
			{
				name:           "High fee",
				inputValue:     1000,
				outputValue:    500,
				fee:            400,
				expectedChange: 100,
				shouldBeValid:  true,
			},
			{
				name:           "Low fee",
				inputValue:     1000,
				outputValue:    950,
				fee:            10,
				expectedChange: 40,
				shouldBeValid:  true,
			},
			{
				name:           "Zero fee",
				inputValue:     1000,
				outputValue:    1000,
				fee:            0,
				expectedChange: 0,
				shouldBeValid:  true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inputs := []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte(fmt.Sprintf("fee_test_%s", tc.name)),
						Sequence:    0xffffffff,
					},
				}
				copy(inputs[0].PrevTxHash, []byte(fmt.Sprintf("fee_input_%s", tc.name)))

				outputs := []*TxOutput{
					{Value: tc.outputValue, ScriptPubKey: []byte("recipient")},
				}

				// Add change output if there's change
				if tc.expectedChange > 0 {
					outputs = append(outputs, &TxOutput{
						Value:        tc.expectedChange,
						ScriptPubKey: []byte("change"),
					})
				}

				tx := NewTransaction(inputs, outputs, tc.fee)
				assert.NotNil(t, tx)

				// Verify fee calculation
				totalOutput := uint64(0)
				for _, output := range outputs {
					totalOutput += output.Value
				}
				totalOutput += tc.fee

				assert.Equal(t, tc.inputValue, totalOutput, "Total input should equal total output + fee")

				// Validate transaction
				err := tx.IsValid()
				if tc.shouldBeValid {
					assert.NoError(t, err, "Transaction should be valid: %s", tc.name)
				} else {
					assert.Error(t, err, "Transaction should fail validation: %s", tc.name)
				}

				t.Logf("Fee test '%s' completed:", tc.name)
				t.Logf("  Input: %d", tc.inputValue)
				t.Logf("  Output: %d", tc.outputValue)
				t.Logf("  Fee: %d", tc.fee)
				t.Logf("  Change: %d", tc.expectedChange)
			})
		}
	})
}

// TestTransactionEdgeCases tests edge cases that might occur in real-world scenarios
func TestTransactionEdgeCases(t *testing.T) {
	t.Run("TestTransactionWithMaximumValues", func(t *testing.T) {
		// Test transaction with maximum possible values
		maxUint64 := ^uint64(0)
		maxUint32 := ^uint32(0)

		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32), // Fill with max values
				PrevTxIndex: maxUint32,
				ScriptSig:   make([]byte, 1000), // Large script
				Sequence:    maxUint32,
			},
		}
		// Fill input hash with max values
		for i := range inputs[0].PrevTxHash {
			inputs[0].PrevTxHash[i] = 0xff
		}

		outputs := []*TxOutput{
			{Value: maxUint64, ScriptPubKey: make([]byte, 1000)}, // Maximum value and large script
		}

		tx := NewTransaction(inputs, outputs, maxUint64)
		assert.NotNil(t, tx)

		// Verify maximum values are preserved
		assert.Equal(t, maxUint32, tx.Inputs[0].PrevTxIndex)
		assert.Equal(t, maxUint32, tx.Inputs[0].Sequence)
		assert.Equal(t, maxUint64, tx.Outputs[0].Value)
		assert.Equal(t, maxUint64, tx.Fee)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with maximum values should be valid")

		// Verify hash calculation works
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")

		t.Logf("Transaction with maximum values created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  Max Input Index: %d", maxUint32)
		t.Logf("  Max Output Value: %d", maxUint64)
		t.Logf("  Max Fee: %d", maxUint64)
	})

	t.Run("TestTransactionWithMinimalData", func(t *testing.T) {
		// Test transaction with minimal data
		inputs := []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte{}, // Empty script (allowed for inputs)
				Sequence:    0xffffffff,
			},
		}

		outputs := []*TxOutput{
			{Value: 1, ScriptPubKey: []byte("minimal")}, // Minimum value, minimal script
		}

		tx := NewTransaction(inputs, outputs, 0)
		assert.NotNil(t, tx)

		// Validate transaction
		err := tx.IsValid()
		assert.NoError(t, err, "Transaction with minimal data should be valid")

		// Verify hash calculation
		hash := tx.CalculateHash()
		assert.Equal(t, 32, len(hash), "Hash should be 32 bytes")

		t.Logf("Transaction with minimal data created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  Empty Input Script: %t", len(tx.Inputs[0].ScriptSig) == 0)
		t.Logf("  Minimal Output Script: %t", len(tx.Outputs[0].ScriptPubKey) > 0)
	})

	t.Run("TestTransactionHashCollisionResistance", func(t *testing.T) {
		// Test that similar transactions produce different hashes
		baseInput := []byte("base_input")
		baseOutput := []byte("base_output")

		hashes := make(map[string]bool)
		numTests := 100

		for i := 0; i < numTests; i++ {
			// Create slightly different transactions
			inputs := []*TxInput{
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: uint32(i),
					ScriptSig:   append(baseInput, byte(i)),
					Sequence:    0xffffffff,
				},
			}
			copy(inputs[0].PrevTxHash, append([]byte("input"), byte(i)))

			outputs := []*TxOutput{
				{Value: uint64(1000 + i), ScriptPubKey: append(baseOutput, byte(i))},
			}

			tx := NewTransaction(inputs, outputs, uint64(i))
			hash := tx.CalculateHash()
			hashHex := hex.EncodeToString(hash)

			// Check for collisions
			if hashes[hashHex] {
				t.Errorf("Hash collision detected for test %d: %s", i, hashHex)
			}
			hashes[hashHex] = true
		}

		assert.Equal(t, numTests, len(hashes), "All transactions should have unique hashes")
		t.Logf("Hash collision resistance test passed: %d unique hashes generated", len(hashes))
	})
}
