package pqc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKEMRoundTrip(t *testing.T) {
	kp, err := GenerateKEMKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, kp)
	assert.Len(t, kp.Classical, 32)
	assert.Len(t, kp.PQ, 1184)

	ct, shared, err := kp.Encapsulate()
	assert.NoError(t, err)
	assert.NotNil(t, ct)
	assert.NotEmpty(t, shared)
	assert.Len(t, shared, 32)
	assert.Len(t, ct.Classical, 32)
	assert.Len(t, ct.PQ, 1088)

	// Decapsulate with the same keypair recovers the shared key.
	recovered, err := kp.Decapsulate(ct)
	assert.NoError(t, err)
	assert.Equal(t, shared, recovered)
}

func TestKEMDifferentKeysDiffer(t *testing.T) {
	kp1, _ := GenerateKEMKeyPair()
	kp2, _ := GenerateKEMKeyPair()
	assert.NotEqual(t, kp1.PQ, kp2.PQ)
}

func TestSigRoundTrip(t *testing.T) {
	sk, err := GenerateSigKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, sk)
	assert.Len(t, sk.Classical, 33)
	assert.Len(t, sk.PQ, WOTSLen*WOTSBytes)

	msg := []byte("post-quantum adrenochain")
	sig, err := sk.Sign(msg)
	assert.NoError(t, err)
	assert.NotNil(t, sig)
	assert.Len(t, sig.Classical, 64)
	assert.Len(t, sig.PQ, WOTSLen*WOTSBytes)

	assert.True(t, sk.Verify(msg, sig.Classical, sig.PQ))
	// Wrong message should fail.
	assert.False(t, sk.Verify([]byte("tampered"), sig.Classical, sig.PQ))
}

func TestHybridEncryptDecrypt(t *testing.T) {
	kp, err := GenerateKEMKeyPair()
	assert.NoError(t, err)
	msg := []byte("confidential payload")
	ct, err := kp.HybridEncrypt(msg)
	assert.NoError(t, err)
	out, err := kp.HybridDecrypt(ct)
	assert.NoError(t, err)
	assert.Equal(t, msg, out)
}
