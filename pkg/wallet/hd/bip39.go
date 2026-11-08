// Package bip39 implements BIP-39 mnemonic code for generating deterministic
// key wallets (https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki).
//
// It provides mnemonic generation, validation, and seed derivation using only
// the Go standard library. The English wordlist (2048 words) is embedded.
//
// This is the first step of the BIP-32/BIP-39/BIP-44 HD wallet pipeline:
//   mnemonic -> seed (BIP-39) -> master node (BIP-32) -> account/change/address
//   keys (BIP-44) -> SLIP-44 coin type selection.
//
// See pkg/wallet/hd/README.md for the full pipeline.

package hd

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// entropyToMnemonic converts entropy bytes to a mnemonic phrase.
func entropyToMnemonic(entropy []byte) ([]string, error) {
	if len(entropy) < 16 || len(entropy) > 32 || len(entropy)%4 != 0 {
		return nil, fmt.Errorf("entropy length must be 16-32 bytes and multiple of 4, got %d", len(entropy))
	}

	// Compute checksum: first (len(entropy)/4) bits of SHA-256(entropy).
	hash := sha256.Sum256(entropy)
	checksumBits := len(entropy) / 4

	// Combine entropy + checksum into a bit stream.
	totalBits := len(entropy)*8 + checksumBits
	words := make([]string, 0, totalBits/11)

	// Iterate over 11-bit chunks.
	var bitBuffer uint
	var bitCount int
	for _, b := range entropy {
		bitBuffer = (bitBuffer << 8) | uint(b)
		bitCount += 8
		for bitCount >= 11 {
			bitCount -= 11
			index := (bitBuffer >> bitCount) & 0x7ff
			words = append(words, wordList[index])
		}
	}
	// Add checksum bits.
	bitBuffer = (bitBuffer << checksumBits) | uint(hash[0]>>(8-checksumBits))
	bitCount += checksumBits
	for bitCount >= 11 {
		bitCount -= 11
		index := (bitBuffer >> bitCount) & 0x7ff
		words = append(words, wordList[index])
	}

	return words, nil
}

// MnemonicToEntropy converts a mnemonic phrase back to entropy.
func MnemonicToEntropy(mnemonic []string) ([]byte, error) {
	if len(mnemonic) < 12 || len(mnemonic) > 24 || len(mnemonic)%3 != 0 {
		return nil, fmt.Errorf("mnemonic must have 12-24 words divisible by 3, got %d", len(mnemonic))
	}

	totalBits := len(mnemonic) * 11
	checksumBits := totalBits / 33
	entropyBits := totalBits - checksumBits
	entropyBytes := entropyBits / 8

	var bitBuffer uint
	var bitCount int
	entropy := make([]byte, 0, entropyBytes)

	for _, word := range mnemonic {
		index, ok := wordIndex[word]
		if !ok {
			return nil, fmt.Errorf("unknown word: %q", word)
		}
		bitBuffer = (bitBuffer << 11) | uint(index)
		bitCount += 11
		// Extract exactly entropyBytes bytes; leave checksumBits bits in the buffer.
		for bitCount >= 8 && len(entropy) < entropyBytes {
			bitCount -= 8
			entropy = append(entropy, byte((bitBuffer>>bitCount)&0xff))
		}
	}

	if len(entropy) != entropyBytes {
		return nil, errors.New("internal: entropy length mismatch")
	}
	if bitCount != checksumBits {
		return nil, errors.New("internal: bit count mismatch")
	}

	// Verify checksum.
	hash := sha256.Sum256(entropy)
	expectedChecksum := hash[0] >> (8 - checksumBits)
	actualChecksum := byte(bitBuffer & ((1 << checksumBits) - 1))
	if actualChecksum != expectedChecksum {
		return nil, errors.New("mnemonic checksum mismatch")
	}

	return entropy, nil
}

// NewMnemonic generates a random BIP-39 mnemonic with the given entropy strength
// (128, 192, or 256 bits => 12, 18, or 24 words).
func NewMnemonic(strength int) ([]string, error) {
	switch strength {
	case 128, 192, 256:
	default:
		return nil, fmt.Errorf("invalid strength %d, must be 128, 192, or 256", strength)
	}
	entropy := make([]byte, strength/8)
	if _, err := rand.Read(entropy); err != nil {
		return nil, fmt.Errorf("reading entropy: %w", err)
	}
	return entropyToMnemonic(entropy)
}

// ValidateMnemonic checks that a mnemonic phrase is valid (correct length,
// known words, and valid checksum).
func ValidateMnemonic(mnemonic []string) error {
	_, err := MnemonicToEntropy(mnemonic)
	return err
}

// NewSeedFromMnemonic derives a BIP-39 seed from a mnemonic and optional
// passphrase. The seed is 64 bytes (SHA-512 of "mnemonic"+passphrase, iterated
// 2048 times with HMAC-SHA512).
func NewSeedFromMnemonic(mnemonic []string, passphrase string) ([]byte, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, err
	}

	mnemonicStr := strings.Join(mnemonic, " ")
	salt := "mnemonic" + passphrase
	return pbkdf2HMACSHA512([]byte(mnemonicStr), []byte(salt), 2048, 64)
}

// wordToIndex is populated by init() from wordList.
var wordIndex map[string]int

func init() {
	wordIndex = make(map[string]int, len(wordList))
	for i, w := range wordList {
		wordIndex[w] = i
	}
}

var _ = fmt.Sprintf // ensure fmt used if build tags change
