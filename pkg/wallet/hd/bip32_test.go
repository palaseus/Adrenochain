package hd

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMasterFromSeedBIP32TestVector(t *testing.T) {
	// BIP-32 test vector 1: all-zero seed.
	seed, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	master, err := NewMasterFromSeed(seed)
	assert.NoError(t, err)
	assert.NotNil(t, master)

	// Expected chain m values (verified against Python hmac/hashlib stdlib,
	// which matches BIP-32: I = HMAC-SHA512(key="Bitcoin seed", data=seed)).
	assert.Equal(t,
		"82a34e3867ea7ea4e67e27865d500ae84e98d07ab1bab06526f0a5a5fdcc3eba",
		hex.EncodeToString(master.PrivateKey[1:]))
	assert.Equal(t,
		"73a206dcfea6640cc65575148d8d18744893bf7fe6d1190ad6acf2af193de4e9",
		hex.EncodeToString(master.ChainCode))
	assert.Equal(t, byte(0), master.Depth)
}

func TestDeriveHardenedPath(t *testing.T) {
	seed, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	master, err := NewMasterFromSeed(seed)
	assert.NoError(t, err)

	// m/0' values verified against Python hmac/hashlib stdlib.
	child, err := master.DerivePath("m/0'")
	assert.NoError(t, err)
	assert.NotNil(t, child)
	assert.Equal(t, uint32(0x80000000), bytesToUint32(child.ChildNum))
	assert.Equal(t, byte(1), child.Depth)
	assert.Equal(t,
		"6866b0f3edb7023cc8feaa186611760544efc9e6946b5bb452a094af1730c3b7",
		hex.EncodeToString(child.PrivateKey[1:]))
	assert.Equal(t,
		"995e1809382c96fb466dbb97949b6065fcfe3ba97ab5c56c69c7eb3a8778c992",
		hex.EncodeToString(child.ChainCode))

	// m/0'/1
	child2, err := master.DerivePath("m/0'/1")
	assert.NoError(t, err)
	assert.NotNil(t, child2)
	assert.Equal(t, byte(2), child2.Depth)
}

func TestDerivePathRejectsPublicOnly(t *testing.T) {
	seed, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	master, err := NewMasterFromSeed(seed)
	assert.NoError(t, err)

	// Create a public-only key (strip private key).
	pubOnly := &ExtendedKey{
		Version:   append([]byte{}, master.Version...),
		Depth:     master.Depth,
		ParentFP:  append([]byte{}, master.ParentFP...),
		ChildNum:  append([]byte{}, master.ChildNum...),
		ChainCode: append([]byte{}, master.ChainCode...),
		PublicKey: append([]byte{}, master.PublicKey...),
	}

	_, err = pubOnly.DerivePath("m/0'")
	assert.Error(t, err)
}

func TestDerivePathInvalid(t *testing.T) {
	seed, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	master, err := NewMasterFromSeed(seed)
	assert.NoError(t, err)

	_, err = master.DerivePath("invalid")
	assert.Error(t, err)

	_, err = master.DerivePath("m/foo")
	assert.Error(t, err)
}

func TestSeedValidation(t *testing.T) {
	_, err := NewMasterFromSeed([]byte{1, 2, 3}) // too short
	assert.Error(t, err)

	_, err = NewMasterFromSeed(make([]byte, 65)) // too long
	assert.Error(t, err)
}

func bytesToUint32(b []byte) uint32 {
	var v uint32
	for _, x := range b {
		v = v<<8 + uint32(x)
	}
	return v
}
