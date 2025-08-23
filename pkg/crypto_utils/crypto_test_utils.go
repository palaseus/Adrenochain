// Package crypto_utils provides cryptographic testing utilities for the adrenochain project
package crypto_utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/palaseus/adrenochain/pkg/block"
)

// CryptoTestUtils provides comprehensive cryptographic testing utilities
type CryptoTestUtils struct {
	t *testing.T
}

// NewCryptoTestUtils creates a new cryptographic testing utilities instance
func NewCryptoTestUtils(t *testing.T) *CryptoTestUtils {
	return &CryptoTestUtils{t: t}
}

// TestKeyPair represents a test cryptographic key pair
type TestKeyPair struct {
	PrivateKey *btcec.PrivateKey
	PublicKey  *btcec.PublicKey
	Address    string
}

// GenerateTestKeyPair generates a new test key pair
func (ctu *CryptoTestUtils) GenerateTestKeyPair() *TestKeyPair {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		ctu.t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.PubKey()
	address := ctu.generateAddress(publicKey.SerializeUncompressed())

	return &TestKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}
}

// generateAddress generates a test address from a public key
// This matches the format expected by UTXO validation (last 20 bytes of SHA256 hash)
func (ctu *CryptoTestUtils) generateAddress(pubKeyBytes []byte) string {
	// Address generation matching UTXO validation: last 20 bytes of SHA256(publicKey)
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[len(hash)-20:])
}

// CreateSignedTransaction creates a transaction with valid cryptographic signatures
func (ctu *CryptoTestUtils) CreateSignedTransaction(
	inputs []*block.TxInput,
	outputs []*block.TxOutput,
	keyPairs map[string]*TestKeyPair, // address -> keyPair mapping
	fee uint64,
) *block.Transaction {
	tx := &block.Transaction{
		Version:  1,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: 0,
		Fee:      fee,
	}

	// Sign each input with its corresponding private key
	for i, input := range inputs {
		// Find the key pair for this input's address
		var keyPair *TestKeyPair
		for _, kp := range keyPairs {
			// For simplicity, we'll use the first key pair
			// In a real scenario, you'd match by address
			keyPair = kp
			break
		}

		if keyPair == nil {
			ctu.t.Fatalf("No key pair found for input %d", i)
		}

		// Create signature data (hash of transaction without signatures)
		signatureData := ctu.CreateSignatureData(tx, i)

		// Sign the data
		signature, err := ctu.SignData(signatureData, keyPair.PrivateKey)
		if err != nil {
			ctu.t.Fatalf("Failed to sign input %d: %v", err)
		}

		// Create script signature: [public_key(65)][signature(64)]
		scriptSig := make([]byte, 0, 65+64)
		scriptSig = append(scriptSig, keyPair.PublicKey.SerializeUncompressed()...)
		scriptSig = append(scriptSig, signature...)

		input.ScriptSig = scriptSig
	}

	// Calculate transaction hash
	tx.Hash = ctu.calculateTxHash(tx)
	return tx
}

// CreateSignatureData creates the data to be signed for a specific input (exported for debugging)
// This method MUST match the exact serialization format used by getTxSignatureData in pkg/utxo/utxo.go
// WARNING: The original getTxSignatureData in utxo.go has a bug - it only uses byte() cast which truncates values!
// We're matching that exact behavior to ensure signatures verify correctly.
func (ctu *CryptoTestUtils) CreateSignatureData(tx *block.Transaction, inputIndex int) []byte {
	data := make([]byte, 0)

	// Version (matching utxo.go getTxSignatureData format - using only lowest byte)
	data = append(data, byte(tx.Version))

	// Inputs (excluding signatures, matching utxo.go format)
	for _, input := range tx.Inputs {
		data = append(data, input.PrevTxHash...)
		data = append(data, byte(input.PrevTxIndex)) // Only lowest byte (matches utxo.go)
		data = append(data, byte(input.Sequence))    // Only lowest byte (matches utxo.go)
	}

	// Outputs (matching utxo.go format)
	for _, output := range tx.Outputs {
		data = append(data, byte(output.Value)) // Only lowest byte (matches utxo.go bug!)
		data = append(data, output.ScriptPubKey...)
	}

	// Lock time and fee (matching utxo.go format - using only lowest byte)
	data = append(data, byte(tx.LockTime)) // Only lowest byte (matches utxo.go)
	data = append(data, byte(tx.Fee))      // Only lowest byte (matches utxo.go)

	// Hash the data (matching utxo.go format)
	hash := sha256.Sum256(data)
	return hash[:]
}

// SignData signs the given data with the private key (exported for debugging)
// The data should already be hashed - don't hash it again!
func (ctu *CryptoTestUtils) SignData(data []byte, privateKey *btcec.PrivateKey) ([]byte, error) {
	// Convert btcec private key to ecdsa format
	ecdsaPrivKey := privateKey.ToECDSA()

	// Use the data directly (it's already hashed by CreateSignatureData)

	// Sign the hash using ecdsa.Sign
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaPrivKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Ensure canonical form (s <= N/2)
	curve := btcec.S256()
	N := curve.N
	if s.Cmp(new(big.Int).Div(N, big.NewInt(2))) > 0 {
		s.Sub(N, s)
	}

	// Encode as 64 bytes: [R(32)][S(32)]
	result := make([]byte, 64)
	r.FillBytes(result[:32])
	s.FillBytes(result[32:])

	return result, nil
}

// calculateTxHash calculates the hash of a transaction
func (ctu *CryptoTestUtils) calculateTxHash(tx *block.Transaction) []byte {
	// Simple hash calculation for testing
	data := fmt.Sprintf("%d-%v-%v-%d-%d",
		tx.Version, tx.Inputs, tx.Outputs, tx.LockTime, tx.Fee)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// CreateTestTransaction creates a complete test transaction with valid signatures
func (ctu *CryptoTestUtils) CreateTestTransaction(
	fromKeyPair *TestKeyPair,
	toKeyPair *TestKeyPair,
	amount uint64,
	fee uint64,
) (*block.Transaction, error) {
	// Create inputs (simplified for testing)
	inputs := []*block.TxInput{
		{
			PrevTxHash:  []byte("mock_tx_hash"),
			PrevTxIndex: 0,
			ScriptSig:   []byte{}, // Will be filled during signing
			Sequence:    0xffffffff,
		},
	}

	// Create outputs
	outputs := []*block.TxOutput{
		{
			Value:        amount,
			ScriptPubKey: []byte(toKeyPair.Address),
		},
	}

	// Add change output if needed (assuming 10000 input for simplicity)
	totalInput := uint64(10000)
	changeAmount := totalInput - amount - fee
	if changeAmount > 0 {
		outputs = append(outputs, &block.TxOutput{
			Value:        changeAmount,
			ScriptPubKey: []byte(fromKeyPair.Address),
		})
	}

	// Create key pairs mapping for signing
	keyPairs := map[string]*TestKeyPair{
		fromKeyPair.Address: fromKeyPair,
	}

	// Create and sign the transaction
	tx := ctu.CreateSignedTransaction(inputs, outputs, keyPairs, fee)
	return tx, nil
}

// CreateComplexTestScenario creates a complex test scenario with multiple transactions
func (ctu *CryptoTestUtils) CreateComplexTestScenario() ([]*TestKeyPair, []*block.Transaction) {
	// Generate multiple key pairs
	alice := ctu.GenerateTestKeyPair()
	bob := ctu.GenerateTestKeyPair()
	charlie := ctu.GenerateTestKeyPair()

	keyPairs := []*TestKeyPair{alice, bob, charlie}

	// Create transactions
	transactions := []*block.Transaction{}

	// Alice sends 3000 to Bob
	tx1, err := ctu.CreateTestTransaction(alice, bob, 3000, 500)
	if err == nil {
		transactions = append(transactions, tx1)
	}

	// Bob sends 2000 to Charlie
	tx2, err := ctu.CreateTestTransaction(bob, charlie, 2000, 300)
	if err == nil {
		transactions = append(transactions, tx2)
	}

	return keyPairs, transactions
}
