package wallet

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/palaseus/adrenochain/pkg/utxo"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestWalletSigningInteroperatesWithUTXOValidation(t *testing.T) {
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(DefaultWalletConfig(), us, newTestStorage(t))
	require.NoError(t, err)
	alice := wallet.GetDefaultAccount()
	bobKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	bob := wallet.generateChecksumAddress(bobKey.ToECDSA())
	hash := sha256.Sum256(alice.PublicKey)
	script := hash[12:]
	funding := sha256.Sum256([]byte("wallet-funding"))
	us.AddUTXOSafe(utxo.NewUTXO(funding[:], 0, 10000, script, hex.EncodeToString(script), false, 0))
	tx, err := wallet.CreateTransaction(alice.Address, bob, 3000, 546)
	require.NoError(t, err)
	require.Len(t, tx.Inputs[0].ScriptSig, 129)
	require.Equal(t, tx.CalculateHash(), tx.Hash)
	require.NoError(t, us.ValidateTransaction(tx))
	valid, err := wallet.VerifyTransaction(tx)
	require.NoError(t, err)
	require.True(t, valid)
	tx.Outputs[0].Value += 256
	valid, err = wallet.VerifyTransaction(tx)
	require.Error(t, err)
	require.False(t, valid)
	// Even recalculating the public transaction hash cannot repair a tampered signature.
	tx.Hash = tx.CalculateHash()
	valid, err = wallet.VerifyTransaction(tx)
	require.Error(t, err)
	require.False(t, valid)
	require.Error(t, us.ValidateTransaction(tx))
	tx.Outputs[0].Value -= 256
	tx.Fee += 256
	require.Error(t, us.ValidateTransaction(tx))
	_, err = wallet.CreateTransaction(alice.Address, bob, math.MaxUint64, 546)
	require.Error(t, err)
}
