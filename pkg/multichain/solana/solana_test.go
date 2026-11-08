package solana

import (
	"crypto/ed25519"
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/multichain/types"
	"github.com/stretchr/testify/assert"
)

func TestKeypairGenerateAndSign(t *testing.T) {
	kp, err := NewKeypair()
	assert.NoError(t, err)
	assert.NotNil(t, kp)
	assert.NotEqual(t, PublicKey{}, kp.PublicKey)

	msg := []byte("solana test message")
	sig := kp.Sign(msg)
	assert.Len(t, sig, SignatureSize)

	// Verify with ed25519.
	assert.True(t, ed25519Verify(kp.PublicKey[:], msg, sig))
}

func TestFromPrivateKey(t *testing.T) {
	kp1, err := NewKeypair()
	assert.NoError(t, err)

	kp2, err := FromPrivateKey(kp1.PrivateKey[:ed25519.SeedSize])
	assert.NoError(t, err)
	assert.Equal(t, kp1.PublicKey, kp2.PublicKey)

	// Wrong seed length.
	_, err = FromPrivateKey([]byte{1, 2, 3})
	assert.Error(t, err)
}

func TestTransferTransaction(t *testing.T) {
	from, err := NewKeypair()
	assert.NoError(t, err)
	to, err := NewKeypair()
	assert.NoError(t, err)

	var bh [HashSize]byte
	copy(bh[:], []byte("recent-blockhash-32-bytes-long!!"))

	tx, err := NewTransferTransaction(from.PublicKey, to.PublicKey, 1_000_000, bh)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Len(t, tx.Signatures, 1)

	// Sign.
	err = tx.Sign([]*Keypair{from})
	assert.NoError(t, err)
	assert.NotNil(t, tx.Signatures[0])
	assert.Len(t, tx.Signatures[0], SignatureSize)

	// Serialize.
	serialized, err := tx.Serialize()
	assert.NoError(t, err)
	assert.Greater(t, len(serialized), 0)
}

func TestTransferTransactionSameAccountFails(t *testing.T) {
	kp, err := NewKeypair()
	assert.NoError(t, err)

	var bh [HashSize]byte
	_, err = NewTransferTransaction(kp.PublicKey, kp.PublicKey, 100, bh)
	assert.Error(t, err)
}

func TestPDADerivation(t *testing.T) {
	programID := mustPublicKey("11111111111111111111111111111111")
	seeds := [][]byte{[]byte("escrow"), []byte{1, 2, 3, 4}}

	pda, bump, err := DeriveProgramAddress(seeds, programID)
	assert.NoError(t, err)
	assert.NotEqual(t, PublicKey{}, pda)
	assert.Greater(t, bump, byte(0))

	// Deterministic: same inputs -> same PDA.
	pda2, bump2, err := DeriveProgramAddress(seeds, programID)
	assert.NoError(t, err)
	assert.Equal(t, pda, pda2)
	assert.Equal(t, bump, bump2)

	// PDA must be off-curve (not a valid ed25519 public key).
	onCurve, err := isEd25519OnCurve(pda[:])
	assert.NoError(t, err)
	assert.False(t, onCurve, "PDA must be off-curve")

	// Different seeds -> different PDA.
	pda3, _, err := DeriveProgramAddress([][]byte{[]byte("other")}, programID)
	assert.NoError(t, err)
	assert.NotEqual(t, pda, pda3)
}

func TestPDADerivationErrors(t *testing.T) {
	programID := mustPublicKey("11111111111111111111111111111111")

	// No seeds.
	_, _, err := DeriveProgramAddress(nil, programID)
	assert.Error(t, err)

	// Too many seeds.
	manySeeds := make([][]byte, MaxSeeds+1)
	for i := range manySeeds {
		manySeeds[i] = []byte{byte(i)}
	}
	_, _, err = DeriveProgramAddress(manySeeds, programID)
	assert.Error(t, err)

	// Seed too long.
	longSeed := make([]byte, SeedMaxLen+1)
	_, _, err = DeriveProgramAddress([][]byte{longSeed}, programID)
	assert.Error(t, err)
}

func TestSPLTokenTransfer(t *testing.T) {
	token := NewSPLToken()
	source := mustPublicKey("source1111111111111111111111111111111111")
	dest := mustPublicKey("dest11111111111111111111111111111111111")
	auth := mustPublicKey("auth11111111111111111111111111111111111")

	instr, err := token.BuildTransfer(source, dest, auth, 5000000)
	assert.NoError(t, err)
	assert.Equal(t, mustPublicKey(TokenProgram), instr.ProgramID)
	assert.Len(t, instr.Accounts, 3)
	assert.True(t, instr.Accounts[0].IsWritable)
	assert.True(t, instr.Accounts[1].IsWritable)
	assert.True(t, instr.Accounts[2].IsSigner)
	assert.Equal(t, byte(3), instr.Data[0])

	// Data: [0]=3, [1:9]=amount LE.
	assert.Equal(t, uint64(5000000), DecodeUint64LE(instr.Data[1:9]))

	// Same source/dest fails.
	_, err = token.BuildTransfer(source, source, auth, 100)
	assert.Error(t, err)
}

func TestSPLTokenMintAndBurn(t *testing.T) {
	token := NewSPLToken()
	mint := mustPublicKey("mint1111111111111111111111111111111111")
	dest := mustPublicKey("dest1111111111111111111111111111111111")
	auth := mustPublicKey("auth11111111111111111111111111111111111")

	mi, err := token.BuildMintTo(mint, dest, auth, 1000)
	assert.NoError(t, err)
	assert.Equal(t, byte(8), mi.Data[0])
	assert.Equal(t, uint64(1000), DecodeUint64LE(mi.Data[1:9]))

	bi, err := token.BuildBurn(mint, dest, auth, 500)
	assert.NoError(t, err)
	assert.Equal(t, byte(9), bi.Data[0])
	assert.Len(t, bi.Accounts, 3)
}

func TestATADerivation(t *testing.T) {
	token := NewSPLToken()
	owner := mustPublicKey("owner1111111111111111111111111111111111")
	mint := mustPublicKey("mint1111111111111111111111111111111111")

	ata, bump, err := token.DeriveAssociatedTokenAccount(owner, mint)
	assert.NoError(t, err)
	assert.NotEqual(t, PublicKey{}, ata)
	assert.Greater(t, bump, byte(0))

	// Deterministic.
	ata2, bump2, err := token.DeriveAssociatedTokenAccount(owner, mint)
	assert.NoError(t, err)
	assert.Equal(t, ata, ata2)
	assert.Equal(t, bump, bump2)
}

func TestWormholeVAARoundTrip(t *testing.T) {
	emitterAddr := [32]byte{1}
	payload := []byte("cross-chain payload")

	vaa := NewVAA(1700000000, 42, ChainIDSolana, emitterAddr, 12345, 1, payload)

	serialized, err := vaa.Serialize()
	assert.NoError(t, err)
	assert.Greater(t, len(serialized), 0)

	deserialized, err := Deserialize(serialized)
	assert.NoError(t, err)
	assert.Equal(t, vaa.Version, deserialized.Version)
	assert.Equal(t, vaa.Timestamp, deserialized.Timestamp)
	assert.Equal(t, vaa.Nonce, deserialized.Nonce)
	assert.Equal(t, vaa.EmitterChain, deserialized.EmitterChain)
	assert.Equal(t, vaa.EmitterAddress, deserialized.EmitterAddress)
	assert.Equal(t, vaa.Sequence, deserialized.Sequence)
	assert.Equal(t, vaa.ConsistencyLevel, deserialized.ConsistencyLevel)
	assert.Equal(t, vaa.Payload, deserialized.Payload)
	assert.Equal(t, vaa.Hash, deserialized.Hash)

	// Hash verification.
	assert.True(t, deserialized.VerifyHash())

	// Threshold: 10 guardians, 7 sigs -> pass (7 > 6).
	assert.True(t, VerifyThreshold(7, 10))
	// 6 sigs -> fail (6 not > 6).
	assert.False(t, VerifyThreshold(6, 10))
}

func TestWormholeVAATamperDetected(t *testing.T) {
	vaa := NewVAA(1700000000, 1, ChainIDSolana, [32]byte{9}, 1, 1, []byte("data"))
	serialized, _ := vaa.Serialize()

	// Tamper with the payload.
	serialized[len(serialized)-10] ^= 0xFF

	deserialized, err := Deserialize(serialized)
	assert.NoError(t, err)
	assert.False(t, deserialized.VerifyHash())
}

func TestConnector(t *testing.T) {
	c := NewConnector(types.ChainTypeSolanaDevnet, "http://localhost:8899")
	assert.Equal(t, ConnectorStatusDisconnected, c.Status())

	err := c.Connect()
	assert.NoError(t, err)
	assert.Equal(t, ConnectorStatusConnected, c.Status())

	slot, err := c.GetSlot()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), slot)

	bh, err := c.GetBlockhash(CommitmentConfirmed)
	assert.NoError(t, err)
	assert.NotEqual(t, Blockhash{}, bh)

	bal, err := c.GetBalance(mustPublicKey("11111111111111111111111111111111"))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, bal, uint64(0))

	sig, err := c.SendTransaction(&Transaction{
		Message: Message{
			AccountKeys: []PublicKey{mustPublicKey("11111111111111111111111111111111")},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, sig, HashSize)

	err = c.Close()
	assert.NoError(t, err)
	assert.Equal(t, ConnectorStatusDisconnected, c.Status())
}

func TestBigIntToUint64(t *testing.T) {
	amt, err := BigIntToUint64(bigFromInt(5000))
	assert.NoError(t, err)
	assert.Equal(t, uint64(5000), amt)

	_, err = BigIntToUint64(bigFromInt(-1))
	assert.Error(t, err)

	_, err = BigIntToUint64(new(big.Int).Lsh(big.NewInt(1), 64))
	assert.Error(t, err)
}

func bigFromInt(v int64) *big.Int {
	return new(big.Int).SetInt64(v)
}

// ed25519Verify verifies an ed25519 signature (re-implemented to avoid importing
// the ed25519 package's Verify which needs crypto/ed25519).
func ed25519Verify(pub []byte, msg, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pub), msg, sig)
}
