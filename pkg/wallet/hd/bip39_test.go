package hd

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func TestNewMnemonicLengths(t *testing.T) {
	for _, strength := range []int{128, 192, 256} {
		words, err := NewMnemonic(strength)
		assert.NoError(t, err)
		assert.Len(t, words, strength/32*3)
		assert.NoError(t, ValidateMnemonic(words))
	}
}

func TestMnemonicRoundTrip(t *testing.T) {
	words, err := NewMnemonic(128)
	assert.NoError(t, err)

	entropy, err := MnemonicToEntropy(words)
	assert.NoError(t, err)
	assert.Len(t, entropy, 16)

	regenerated, err := entropyToMnemonic(entropy)
	assert.NoError(t, err)
	assert.Equal(t, words, regenerated)
}

func TestSeedFromMnemonic(t *testing.T) {
	words := []string{
		"abandon", "abandon", "abandon", "abandon", "abandon", "abandon",
		"abandon", "abandon", "abandon", "abandon", "abandon", "about",
	}
	seed, err := NewSeedFromMnemonic(words, "TREZOR")
	assert.NoError(t, err)
	assert.Len(t, seed, 64)
	// BIP-39 test vector: mnemonic+passphrase "TREZOR" -> known seed (verified against hashlib.pbkdf2_hmac).
	assert.Equal(t,
		"c55257c360c07c72029aebc1b53c05ed0362ada38ead3e3e9efa3708e53495531f09a6987599d18264c1e1c92f2cf141630c7a3c4ab7c81b2f001698e7463b04",
		bytesToHex(seed))
}

func TestInvalidMnemonic(t *testing.T) {
	assert.Error(t, ValidateMnemonic(nil))
	assert.Error(t, ValidateMnemonic([]string{"notaword"}))
	assert.Error(t, ValidateMnemonic([]string{"abandon"})) // too short
}

func TestNewMnemonicInvalidStrength(t *testing.T) {
	_, err := NewMnemonic(100)
	assert.Error(t, err)
}
