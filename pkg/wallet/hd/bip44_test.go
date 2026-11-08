package hd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountPath(t *testing.T) {
	// Ethereum account 0, external, index 5: m/44'/60'/0'/0/5
	path := AccountPath(CoinEthereum, 0, 0, 5)
	assert.Equal(t, "m/44'/60'/0'/0/5", path)

	// Bitcoin account 1, change (internal), index 0: m/44'/0'/1'/1/0
	path = AccountPath(CoinBitcoin, 1, 1, 0)
	assert.Equal(t, "m/44'/0'/1'/1/0", path)

	// Solana account 0: m/44'/501'/0'/0/0
	path = AccountPath(CoinSolana, 0, 0, 0)
	assert.Equal(t, "m/44'/501'/0'/0/0", path)
}

func TestAccountPathDerives(t *testing.T) {
	seed := make([]byte, 32)
	master, err := NewMasterFromSeed(seed)
	assert.NoError(t, err)

	// Derive an Ethereum account 0 external address key.
	path := AccountPath(CoinEthereum, 0, 0, 7)
	ak, err := master.DerivePath(path)
	assert.NoError(t, err)
	assert.NotNil(t, ak)
	assert.Equal(t, byte(5), ak.Depth) // m/44'/60'/0'/0/7 = depth 5
}

func TestCoinTypeString(t *testing.T) {
	assert.Equal(t, "Ethereum", CoinEthereum.String())
	assert.Equal(t, "Bitcoin", CoinBitcoin.String())
	assert.Equal(t, "Solana", CoinSolana.String())
	assert.Contains(t, CoinType(9999).String(), "coin-")
}

func TestParsePathSplitAndIsHardened(t *testing.T) {
	parts, err := ParsePathSplit("m/44'/60'/0'/0/5")
	assert.NoError(t, err)
	assert.Equal(t, []string{"44'", "60'", "0'", "0", "5"}, parts)
	assert.True(t, IsHardened("44'"))
	assert.True(t, IsHardened("60h"))
	assert.True(t, IsHardened("0'H"))
	assert.False(t, IsHardened("0"))
	assert.False(t, IsHardened("5"))
}
