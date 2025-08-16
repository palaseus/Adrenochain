package engine

import (
	"testing"
)

func TestAddressString(t *testing.T) {
	// Test with zero address
	zeroAddr := Address{}
	str := zeroAddr.String()
	expected := "0000000000000000000000000000000000000000"
	if str != expected {
		t.Errorf("Expected zero address string to be %s, got %s", expected, str)
	}

	// Test with specific address
	addr := Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14}
	str = addr.String()
	expected = "0102030405060708090a0b0c0d0e0f1011121314"
	if str != expected {
		t.Errorf("Expected address string to be %s, got %s", expected, str)
	}

	// Test with address that has all bytes set
	allBytesAddr := Address{}
	for i := range allBytesAddr {
		allBytesAddr[i] = 0xFF
	}
	str = allBytesAddr.String()
	expected = "ffffffffffffffffffffffffffffffffffffffff"
	if str != expected {
		t.Errorf("Expected all-bytes address string to be %s, got %s", expected, str)
	}
}

func TestHashString(t *testing.T) {
	// Test with zero hash
	zeroHash := Hash{}
	str := zeroHash.String()
	expected := "0000000000000000000000000000000000000000000000000000000000000000"
	if str != expected {
		t.Errorf("Expected zero hash string to be %s, got %s", expected, str)
	}

	// Test with specific hash
	hash := Hash{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}
	str = hash.String()
	expected = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	if str != expected {
		t.Errorf("Expected hash string to be %s, got %s", expected, str)
	}

	// Test with hash that has all bytes set
	allBytesHash := Hash{}
	for i := range allBytesHash {
		allBytesHash[i] = 0xFF
	}
	str = allBytesHash.String()
	expected = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if str != expected {
		t.Errorf("Expected all-bytes hash string to be %s, got %s", expected, str)
	}
}

func TestAddressBytes(t *testing.T) {
	// Test with zero address
	zeroAddr := Address{}
	bytes := zeroAddr.Bytes()
	if len(bytes) != 20 {
		t.Errorf("Expected address bytes length to be 20, got %d", len(bytes))
	}
	for _, b := range bytes {
		if b != 0 {
			t.Error("Expected all bytes to be 0")
		}
	}

	// Test with specific address
	addr := Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14}
	bytes = addr.Bytes()
	if len(bytes) != 20 {
		t.Errorf("Expected address bytes length to be 20, got %d", len(bytes))
	}
	if bytes[0] != 0x01 || bytes[19] != 0x14 {
		t.Error("Expected bytes to match address values")
	}

	// Test that bytes are a copy, not a reference
	bytes[0] = 0xFF
	if addr[0] != 0x01 {
		t.Error("Modifying returned bytes should not affect original address")
	}
}

func TestHashBytes(t *testing.T) {
	// Test with zero hash
	zeroHash := Hash{}
	bytes := zeroHash.Bytes()
	if len(bytes) != 32 {
		t.Errorf("Expected hash bytes length to be 32, got %d", len(bytes))
	}
	for _, b := range bytes {
		if b != 0 {
			t.Error("Expected all bytes to be 0")
		}
	}

	// Test with specific hash
	hash := Hash{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}
	bytes = hash.Bytes()
	if len(bytes) != 32 {
		t.Errorf("Expected hash bytes length to be 32, got %d", len(bytes))
	}
	if bytes[0] != 0x01 || bytes[31] != 0x20 {
		t.Error("Expected bytes to match hash values")
	}

	// Test that bytes are a copy, not a reference
	bytes[0] = 0xFF
	if hash[0] != 0x01 {
		t.Error("Modifying returned bytes should not affect original hash")
	}
}

func TestParseAddress(t *testing.T) {
	// Test with valid address without 0x prefix
	addr, err := ParseAddress("0000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("Expected no error parsing valid address, got: %v", err)
	}
	if addr != (Address{}) {
		t.Error("Expected zero address")
	}

	// Test with valid address with 0x prefix
	addr, err = ParseAddress("0x0000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("Expected no error parsing valid address with 0x prefix, got: %v", err)
	}
	if addr != (Address{}) {
		t.Error("Expected zero address")
	}

	// Test with specific address
	addr, err = ParseAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
	if err != nil {
		t.Errorf("Expected no error parsing specific address, got: %v", err)
	}
	expected := Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14}
	if addr != expected {
		t.Error("Expected parsed address to match expected")
	}

	// Test with invalid length (too short)
	_, err = ParseAddress("1234567890abcdef")
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for short address, got: %v", err)
	}

	// Test with invalid length (too long)
	_, err = ParseAddress("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for long address, got: %v", err)
	}

	// Test with invalid hex characters
	_, err = ParseAddress("000000000000000000000000000000000000000g")
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}

	// Test with empty string
	_, err = ParseAddress("")
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for empty string, got: %v", err)
	}
}

func TestParseHash(t *testing.T) {
	// Test with valid hash without 0x prefix
	hash, err := ParseHash("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("Expected no error parsing valid hash, got: %v", err)
	}
	if hash != (Hash{}) {
		t.Error("Expected zero hash")
	}

	// Test with valid hash with 0x prefix
	hash, err = ParseHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Errorf("Expected no error parsing valid hash with 0x prefix, got: %v", err)
	}
	if hash != (Hash{}) {
		t.Error("Expected zero hash")
	}

	// Test with specific hash
	hash, err = ParseHash("0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	if err != nil {
		t.Errorf("Expected no error parsing specific hash, got: %v", err)
	}
	expected := Hash{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}
	if hash != expected {
		t.Error("Expected parsed hash to match expected")
	}

	// Test with invalid length (too short)
	_, err = ParseHash("1234567890abcdef1234567890abcdef")
	if err != ErrInvalidHash {
		t.Errorf("Expected ErrInvalidHash for short hash, got: %v", err)
	}

	// Test with invalid length (too long)
	_, err = ParseHash("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != ErrInvalidHash {
		t.Errorf("Expected ErrInvalidHash for long hash, got: %v", err)
	}

	// Test with invalid hex characters
	_, err = ParseHash("000000000000000000000000000000000000000000000000000000000000000g")
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}

	// Test with empty string
	_, err = ParseHash("")
	if err != ErrInvalidHash {
		t.Errorf("Expected ErrInvalidHash for empty string, got: %v", err)
	}
}

func TestAddressAndHashEdgeCases(t *testing.T) {
	// Test address with lowercase hex (40 characters = 20 bytes)
	addr, err := ParseAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	if err != nil {
		t.Errorf("Expected no error parsing lowercase address, got: %v", err)
	}
	if addr.String() != "abcdef1234567890abcdef1234567890abcdef12" {
		t.Error("Expected address string to match input")
	}

	// Test hash with lowercase hex (64 characters = 32 bytes)
	hash, err := ParseHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	if err != nil {
		t.Errorf("Expected no error parsing lowercase hash, got: %v", err)
	}
	if hash.String() != "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890" {
		t.Error("Expected hash string to match input")
	}

	// Test address with leading zeros
	addr, err = ParseAddress("0x0000000000000000000000000000000000000001")
	if err != nil {
		t.Errorf("Expected no error parsing address with leading zeros, got: %v", err)
	}
	expected := Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	if addr != expected {
		t.Error("Expected parsed address to match expected")
	}

	// Test hash with leading zeros
	hash, err = ParseHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Errorf("Expected no error parsing hash with leading zeros, got: %v", err)
	}
	expectedHash := Hash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	if hash != expectedHash {
		t.Error("Expected parsed hash to match expected")
	}
}
