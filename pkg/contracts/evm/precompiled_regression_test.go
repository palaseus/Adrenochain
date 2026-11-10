package evm

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRIPEMD160PrecompileKnownVector(t *testing.T) {
	result, err := runRIPEMD160(nil)
	require.NoError(t, err)
	require.Equal(t, "0000000000000000000000009c1185a5c5e9fc54612808977ee8f548b2258d31", hex.EncodeToString(result))
}
