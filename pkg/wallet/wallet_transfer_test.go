package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gochain/gochain/pkg/utxo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWalletTransfer tests the complete flow of sending value from one wallet to another
func TestWalletTransfer(t *testing.T) {
	t.Run("TestBasicTransfer", func(t *testing.T) {
		// Create a test storage
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		// Get the default account (Alice)
		aliceAccount := wallet.GetDefaultAccount()
		require.NotNil(t, aliceAccount, "Alice account should exist")
		t.Logf("Alice's address: %s", aliceAccount.Address)

		// Create a second account (Bob) by generating a new key pair
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())
		t.Logf("Bob's address: %s", bobAddress)

		// Give Alice some initial funds by creating a coinbase UTXO
		aliceInitialBalance := uint64(10000) // 10,000 coins
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32), // Proper 32-byte hash
			TxIndex:      0,
			Value:        aliceInitialBalance,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   true,
			Height:       1,
		}
		// Fill with some data to make it unique
		copy(aliceUTXO.TxHash, []byte("alice_coinbase_tx_hash_32bytes"))
		us.AddUTXO(aliceUTXO)

		// Verify Alice's initial balance in UTXO set
		aliceBalance := us.GetBalance(aliceAccount.Address)
		assert.Equal(t, aliceInitialBalance, aliceBalance, "Alice should have initial balance in UTXO set")
		t.Logf("Alice's initial balance: %d", aliceBalance)

		// Alice wants to send 3000 coins to Bob
		transferAmount := uint64(3000)
		fee := uint64(546) // Minimum fee

		// Create the transaction
		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, fee)
		require.NoError(t, err, "Transaction creation should succeed")
		require.NotNil(t, tx, "Transaction should not be nil")

		// Verify transaction structure
		assert.Equal(t, uint32(1), tx.Version)
		assert.Equal(t, 1, len(tx.Inputs), "Should have 1 input (Alice's UTXO)")
		assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs (Bob + change)")
		assert.Equal(t, fee, tx.Fee)
		assert.Equal(t, uint64(0), tx.LockTime)

		// Verify the input references Alice's UTXO
		input := tx.Inputs[0]
		assert.Equal(t, aliceUTXO.TxHash, input.PrevTxHash, "Input should reference Alice's UTXO")
		assert.Equal(t, aliceUTXO.TxIndex, input.PrevTxIndex, "Input should reference correct index")

		// Verify outputs
		// First output should be to Bob
		bobOutput := tx.Outputs[0]
		assert.Equal(t, transferAmount, bobOutput.Value, "Bob should receive the transfer amount")

		// Second output should be change back to Alice
		changeOutput := tx.Outputs[1]
		expectedChange := aliceInitialBalance - transferAmount - fee
		assert.Equal(t, expectedChange, changeOutput.Value, "Change should be correct")

		// Verify transaction hash
		assert.Equal(t, 32, len(tx.Hash), "Hash should be 32 bytes")
		t.Logf("Transaction hash: %x", tx.Hash)

		// Verify transaction is valid
		err = tx.IsValid()
		assert.NoError(t, err, "Transaction should be valid")

		// Verify transaction signature
		valid, err := wallet.VerifyTransaction(tx)
		require.NoError(t, err, "Transaction verification should not error")
		assert.True(t, valid, "Transaction signature should be valid")

		t.Logf("Transaction created successfully:")
		t.Logf("  Hash: %x", tx.Hash)
		t.Logf("  Inputs: %d", len(tx.Inputs))
		t.Logf("  Outputs: %d", len(tx.Outputs))
		t.Logf("  Transfer Amount: %d", transferAmount)
		t.Logf("  Fee: %d", fee)
		t.Logf("  Change: %d", expectedChange)
	})

	t.Run("TestTransferWithMultipleUTXOs", func(t *testing.T) {
		// Test sending value when the sender has multiple UTXOs
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		// Get Alice's account
		aliceAccount := wallet.GetDefaultAccount()
		require.NotNil(t, aliceAccount)

		// Create Bob's address
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		// Give Alice multiple UTXOs
		utxo1 := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        5000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       2,
		}
		copy(utxo1.TxHash, []byte("utxo_1_hash_32bytes_long_hash"))

		utxo2 := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        3000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       3,
		}
		copy(utxo2.TxHash, []byte("utxo_2_hash_32bytes_long_hash"))

		utxo3 := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        2000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       4,
		}
		copy(utxo3.TxHash, []byte("utxo_3_hash_32bytes_long_hash"))

		us.AddUTXO(utxo1)
		us.AddUTXO(utxo2)
		us.AddUTXO(utxo3)

		// Verify total balance in UTXO set
		totalBalance := us.GetBalance(aliceAccount.Address)
		expectedTotal := utxo1.Value + utxo2.Value + utxo3.Value
		assert.Equal(t, expectedTotal, totalBalance, "Total balance should match sum of UTXOs")
		t.Logf("Alice's total balance: %d", totalBalance)

		// Try to send an amount that requires multiple UTXOs
		transferAmount := uint64(6000) // Requires utxo1 (5000) + utxo2 (3000) = 8000
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, fee)
		require.NoError(t, err, "Transaction creation should succeed")

		// Should have multiple inputs
		assert.GreaterOrEqual(t, len(tx.Inputs), 2, "Should have at least 2 inputs")
		t.Logf("Transaction uses %d inputs", len(tx.Inputs))

		// Verify total input value covers transfer + fee
		totalInputValue := uint64(0)
		for _, input := range tx.Inputs {
			// Find the corresponding UTXO
			for _, utxo := range []*utxo.UTXO{utxo1, utxo2, utxo3} {
				if hex.EncodeToString(utxo.TxHash) == hex.EncodeToString(input.PrevTxHash) && utxo.TxIndex == input.PrevTxIndex {
					totalInputValue += utxo.Value
					break
				}
			}
		}

		assert.GreaterOrEqual(t, totalInputValue, transferAmount+fee, "Total input should cover transfer + fee")
		t.Logf("Total input value: %d", totalInputValue)

		// Verify outputs
		assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs (recipient + change)")
		assert.Equal(t, transferAmount, tx.Outputs[0].Value, "First output should be transfer amount")

		// Verify transaction is valid and signed
		err = tx.IsValid()
		assert.NoError(t, err, "Transaction should be valid")

		valid, err := wallet.VerifyTransaction(tx)
		require.NoError(t, err)
		assert.True(t, valid, "Transaction signature should be valid")
	})

	t.Run("TestTransferWithExactAmount", func(t *testing.T) {
		// Test sending exactly the amount available (no change output)
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		aliceAccount := wallet.GetDefaultAccount()
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		// Give Alice exactly 1000 coins
		exactAmount := uint64(1000)
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        exactAmount,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       5,
		}
		copy(aliceUTXO.TxHash, []byte("exact_amount_utxo_32bytes_hash"))
		us.AddUTXO(aliceUTXO)

		// Try to send the exact amount (minus fee)
		transferAmount := exactAmount - 546 // 1000 - 546 = 454
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, fee)
		require.NoError(t, err, "Transaction creation should succeed")

		// Should have only 1 output (to Bob) since no change
		assert.Equal(t, 1, len(tx.Outputs), "Should have only 1 output (no change)")
		assert.Equal(t, transferAmount, tx.Outputs[0].Value, "Bob should receive the transfer amount")

		// Verify transaction
		err = tx.IsValid()
		assert.NoError(t, err, "Transaction should be valid")

		valid, err := wallet.VerifyTransaction(tx)
		require.NoError(t, err)
		assert.True(t, valid, "Transaction signature should be valid")

		t.Logf("Exact amount transfer successful:")
		t.Logf("  Sent: %d", transferAmount)
		t.Logf("  Fee: %d", fee)
		t.Logf("  Outputs: %d", len(tx.Outputs))
	})

	t.Run("TestTransferInsufficientFunds", func(t *testing.T) {
		// Test that transfer fails when insufficient funds
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		aliceAccount := wallet.GetDefaultAccount()
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		// Give Alice only 1000 coins
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        1000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       6,
		}
		copy(aliceUTXO.TxHash, []byte("small_utxo_32bytes_hash"))
		us.AddUTXO(aliceUTXO)

		// Try to send more than available
		transferAmount := uint64(2000) // More than available
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, fee)
		assert.Error(t, err, "Transaction creation should fail with insufficient funds")
		assert.Nil(t, tx, "Transaction should be nil")
		assert.Contains(t, err.Error(), "insufficient funds", "Error should mention insufficient funds")

		t.Logf("Insufficient funds test passed: %v", err)
	})

	t.Run("TestTransferWithLowFee", func(t *testing.T) {
		// Test that transfer fails with fee below dust threshold
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		aliceAccount := wallet.GetDefaultAccount()
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		// Give Alice some funds
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        10000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       7,
		}
		copy(aliceUTXO.TxHash, []byte("fee_test_utxo_32bytes_hash"))
		us.AddUTXO(aliceUTXO)

		// Try to send with fee below dust threshold
		transferAmount := uint64(1000)
		lowFee := uint64(100) // Below 546 dust threshold

		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, lowFee)
		assert.Error(t, err, "Transaction creation should fail with low fee")
		assert.Nil(t, tx, "Transaction should be nil")
		assert.Contains(t, err.Error(), "fee too low", "Error should mention fee too low")

		t.Logf("Low fee test passed: %v", err)
	})
}

// TestWalletTransferEdgeCases tests edge cases in wallet transfers
func TestWalletTransferEdgeCases(t *testing.T) {
	t.Run("TestTransferToInvalidAddress", func(t *testing.T) {
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		aliceAccount := wallet.GetDefaultAccount()

		// Give Alice some funds
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        10000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       8,
		}
		copy(aliceUTXO.TxHash, []byte("invalid_addr_test_32bytes_hash"))
		us.AddUTXO(aliceUTXO)

		// Try to send to invalid address
		invalidAddress := "invalid_address_format"
		transferAmount := uint64(1000)
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(aliceAccount.Address, invalidAddress, transferAmount, fee)
		assert.Error(t, err, "Transaction creation should fail with invalid address")
		assert.Nil(t, tx, "Transaction should be nil")
		assert.Contains(t, err.Error(), "invalid recipient address", "Error should mention invalid address")

		t.Logf("Invalid address test passed: %v", err)
	})

	t.Run("TestTransferFromNonexistentAccount", func(t *testing.T) {
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		// Try to send from an account that doesn't exist
		nonexistentAddress := "nonexistent_address"
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		transferAmount := uint64(1000)
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(nonexistentAddress, bobAddress, transferAmount, fee)
		assert.Error(t, err, "Transaction creation should fail with nonexistent account")
		assert.Nil(t, tx, "Transaction should be nil")
		assert.Contains(t, err.Error(), "account not found", "Error should mention account not found")

		t.Logf("Nonexistent account test passed: %v", err)
	})

	t.Run("TestTransferWithZeroAmount", func(t *testing.T) {
		s := newTestStorage(t)
		config := DefaultWalletConfig()
		us := utxo.NewUTXOSet()
		wallet, err := NewWallet(config, us, s)
		require.NoError(t, err)

		aliceAccount := wallet.GetDefaultAccount()
		bobPrivKey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		bobAddress := wallet.generateChecksumAddress(bobPrivKey.ToECDSA())

		// Give Alice some funds
		aliceUTXO := &utxo.UTXO{
			TxHash:       make([]byte, 32),
			TxIndex:      0,
			Value:        10000,
			ScriptPubKey: aliceAccount.PublicKey,
			Address:      aliceAccount.Address,
			IsCoinbase:   false,
			Height:       9,
		}
		copy(aliceUTXO.TxHash, []byte("zero_amount_test_32bytes_hash"))
		us.AddUTXO(aliceUTXO)

		// Try to send zero amount
		transferAmount := uint64(0)
		fee := uint64(546)

		tx, err := wallet.CreateTransaction(aliceAccount.Address, bobAddress, transferAmount, fee)
		// Zero amount transfers might be allowed by the wallet but fail validation
		if err != nil {
			t.Logf("Zero amount transfer failed at creation: %v", err)
		} else {
			// If it succeeds at creation, it should fail validation
			err = tx.IsValid()
			assert.Error(t, err, "Zero amount transaction should fail validation")
			t.Logf("Zero amount transfer created but failed validation: %v", err)
		}
	})
}
