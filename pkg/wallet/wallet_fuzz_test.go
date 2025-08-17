//go:build go1.18

package wallet

import (
	"strings"
	"testing"

	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/palaseus/adrenochain/pkg/utxo"
)

// FuzzAddressValidation tests address validation with fuzzed data
func FuzzAddressValidation(f *testing.F) {
	// Seed corpus with valid addresses
	f.Add("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")         // Example Bitcoin address
	f.Add("bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh") // Example Bech32 address

	f.Fuzz(func(t *testing.T, address string) {
		// Skip very long addresses
		if len(address) > 1000 {
			t.Skip("Address too long")
		}

		// Create a minimal wallet for testing
		config := DefaultWalletConfig()
		utxoSet := &utxo.UTXOSet{}
		storage := &storage.Storage{}
		wallet, err := NewWallet(config, utxoSet, storage)
		if err != nil {
			t.Skip("Failed to create wallet for testing")
		}

		// Test address decoding
		_, err = wallet.decodeAddressWithChecksum(address)
		if err != nil {
			// Invalid addresses should fail decoding
			return
		}

		// If decoding succeeds, verify the address format
		if len(address) < 26 || len(address) > 35 {
			t.Errorf("Valid address has unexpected length: %d", len(address))
		}

		// Test address encoding round-trip
		decoded, err := wallet.decodeAddressWithChecksum(address)
		if err != nil {
			return
		}

		encoded := wallet.encodeAddressWithChecksum(decoded)
		if encoded != address {
			t.Errorf("Address encoding round-trip failed: %s != %s", encoded, address)
		}
	})
}

// FuzzPrivateKeyImport tests private key import with fuzzed data
func FuzzPrivateKeyImport(f *testing.F) {
	// Seed corpus with valid private keys
	f.Add("5KJvsngHeMpm884wtkJNzQGaCErckhHJBGFsvd3VyK5qMZXj3hS")
	f.Add("L5KjX2aNqjDFtLQ4TdVfncNrw3aHRiCP45HiTiDCc5Evwek2FhN")

	f.Fuzz(func(t *testing.T, privateKeyHex string) {
		// Skip very long inputs
		if len(privateKeyHex) > 1000 {
			t.Skip("Private key too long")
		}

		// Create a minimal wallet for testing
		config := DefaultWalletConfig()
		utxoSet := &utxo.UTXOSet{}
		storage := &storage.Storage{}
		wallet, err := NewWallet(config, utxoSet, storage)
		if err != nil {
			t.Skip("Failed to create wallet for testing")
		}

		// Test private key import
		account, err := wallet.ImportPrivateKey(privateKeyHex)
		if err != nil {
			// Invalid private keys should fail import
			return
		}

		// If import succeeds, verify the account
		if account == nil {
			t.Errorf("Imported account is nil")
			return
		}

		if account.Address == "" {
			t.Errorf("Imported account has empty address")
		}

		if len(account.PrivateKey) == 0 {
			t.Errorf("Imported account has empty private key")
		}

		// Test private key export
		exportedKey, err := wallet.ExportPrivateKey(account.Address)
		if err != nil {
			t.Errorf("Failed to export private key: %v", err)
			return
		}

		// Compare case-insensitively since hex encoding might produce different case
		if strings.ToLower(exportedKey) != strings.ToLower(privateKeyHex) {
			t.Errorf("Private key export mismatch: %s != %s", exportedKey, privateKeyHex)
		}
	})
}

// FuzzEncryptionDecryption tests wallet encryption/decryption with fuzzed data
func FuzzEncryptionDecryption(f *testing.F) {
	// Seed corpus with valid data
	f.Add([]byte("test data"))
	f.Add([]byte("another test"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Data too large")
		}

		// Create a minimal wallet for testing
		config := DefaultWalletConfig()
		config.Passphrase = "test_passphrase"
		utxoSet := &utxo.UTXOSet{}
		storage := &storage.Storage{}
		wallet, err := NewWallet(config, utxoSet, storage)
		if err != nil {
			t.Skip("Failed to create wallet for testing")
		}

		// Test encryption
		encrypted, err := wallet.Encrypt(data)
		if err != nil {
			t.Errorf("Failed to encrypt data: %v", err)
			return
		}

		// Verify encrypted data is different from original
		if len(encrypted) == len(data) && string(encrypted) == string(data) {
			t.Errorf("Encrypted data should be different from original")
		}

		// Test decryption
		decrypted, err := wallet.Decrypt(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt data: %v", err)
			return
		}

		// Verify decrypted data matches original
		if string(decrypted) != string(data) {
			t.Errorf("Decrypted data mismatch: %s != %s", string(decrypted), string(data))
		}
	})
}

// FuzzSignatureValidation tests signature validation with fuzzed data
func FuzzSignatureValidation(f *testing.F) {
	// Seed corpus with valid signature data
	f.Add([]byte{0x30, 0x06, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01}) // Example DER signature
	f.Add([]byte{0x30, 0x08, 0x02, 0x02, 0x00, 0x01, 0x02, 0x02, 0x00, 0x01})

	f.Fuzz(func(t *testing.T, signatureData []byte) {
		// Skip very large inputs
		if len(signatureData) > 1000000 {
			t.Skip("Signature data too large")
		}

		// Test DER signature decoding
		r, s, err := decodeSignatureDER(signatureData)
		if err != nil {
			// Invalid signatures should fail decoding
			return
		}

		// If decoding succeeds, verify the signature components
		if r == nil {
			t.Errorf("Decoded signature has nil r component")
			return
		}

		if s == nil {
			t.Errorf("Decoded signature has nil s component")
			return
		}

		// Test signature encoding round-trip
		encoded, err := encodeSignatureDER(r, s)
		if err != nil {
			t.Errorf("Failed to encode signature: %v", err)
			return
		}

		// Note: DER encoding might not be exactly the same due to different representations
		// So we just verify that re-decoding gives the same r and s values
		r2, s2, err := decodeSignatureDER(encoded)
		if err != nil {
			t.Errorf("Failed to decode re-encoded signature: %v", err)
			return
		}

		if r.Cmp(r2) != 0 || s.Cmp(s2) != 0 {
			t.Errorf("Signature encoding round-trip failed")
		}
	})
}

// FuzzBase58Encoding tests base58 encoding/decoding with fuzzed data
func FuzzBase58Encoding(f *testing.F) {
	// Seed corpus with valid data
	f.Add([]byte("test"))
	f.Add([]byte{0x00, 0x01, 0x02, 0x03})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Data too large")
		}

		// Create a minimal wallet for testing
		config := DefaultWalletConfig()
		utxoSet := &utxo.UTXOSet{}
		storage := &storage.Storage{}
		wallet, err := NewWallet(config, utxoSet, storage)
		if err != nil {
			t.Skip("Failed to create wallet for testing")
		}

		// Test base58 encoding
		encoded := wallet.base58Encode(data)

		// Empty input should produce empty output (this is correct behavior)
		if len(data) == 0 && encoded != "" {
			t.Errorf("Base58 encoding of empty data should return empty string, got: '%s'", encoded)
			return
		}

		// Non-empty input should produce non-empty output
		if len(data) > 0 && encoded == "" {
			t.Errorf("Base58 encoding of non-empty data returned empty string")
			return
		}

		// Test base58 decoding (skip for empty data)
		if len(data) > 0 {
			decoded, err := wallet.base58Decode(encoded)
			if err != nil {
				t.Errorf("Failed to decode base58: %v", err)
				return
			}

			// Verify decoded data matches original
			if len(decoded) != len(data) {
				t.Errorf("Decoded data length mismatch: %d != %d", len(decoded), len(data))
				return
			}

			for i, b := range data {
				if decoded[i] != b {
					t.Errorf("Decoded data mismatch at index %d: %d != %d", i, decoded[i], b)
					break
				}
			}
		}
	})
}
