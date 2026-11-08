package eip7930

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValid(t *testing.T) {
	tests := []struct {
		input    string
		chainID  string
		address  string
		isEVM    bool
		evmChain uint64
	}{
		{"eip7930:eip155:1:0xabc123", "eip155:1", "0xabc123", true, 1},
		{"eip7930:eip155:11155111:0xdeadbeef", "eip155:11155111", "0xdeadbeef", true, 11155111},
		{"eip7930:solana:5eykt8UsFv8PN8N8:someSolanaAddr", "solana:5eykt8UsFv8PN8N8", "someSolanaAddr", false, 0},
		{"eip7930:eip155:137:0xpolygon", "eip155:137", "0xpolygon", true, 137},
	}

	for _, tt := range tests {
		addr, err := Parse(tt.input)
		assert.NoError(t, err, "input: %s", tt.input)
		assert.Equal(t, tt.chainID, addr.ChainID)
		assert.Equal(t, tt.address, addr.Address)
		assert.Equal(t, tt.isEVM, addr.IsEVM())
		assert.Equal(t, tt.evmChain, addr.EVMChainID())
		assert.Equal(t, tt.input, addr.String())
	}
}

func TestParseInvalid(t *testing.T) {
	invalid := []string{
		"",
		"eip7930:",
		"eip7930:eip155:1",
		"eip7930::0xabc",
		"http://eip155:1:0xabc",
		"eip155:1:0xabc",
		"eip7930:eip155:1:",
	}

	for _, input := range invalid {
		_, err := Parse(input)
		assert.Error(t, err, "input: %s", input)
	}
}

func TestMustParse(t *testing.T) {
	addr := MustParse("eip7930:eip155:1:0xabc")
	assert.Equal(t, "eip155:1", addr.ChainID)
	assert.Equal(t, "0xabc", addr.Address)

	assert.Panics(t, func() {
		MustParse("invalid")
	})
}

func TestCommonChainIDs(t *testing.T) {
	assert.Equal(t, "eip155:1", ChainIDEthereumMainnet)
	assert.Equal(t, "eip155:11155111", ChainIDEthereumSepolia)
	assert.Equal(t, "eip155:42161", ChainIDArbitrumOne)
	assert.Equal(t, "eip155:10", ChainIDOptimism)
	assert.Equal(t, "eip155:137", ChainIDPolygon)
	assert.Equal(t, "eip155:56", ChainIDBSC)
	assert.Equal(t, "eip155:43114", ChainIDAvalancheC)
	assert.Equal(t, "solana:5eykt8UsFv8PN8N8", ChainIDSolanaMainnet)
}
