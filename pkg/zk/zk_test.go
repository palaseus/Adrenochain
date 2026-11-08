package zk

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitment(t *testing.T) {
	value := big.NewInt(100)
	blinding := big.NewInt(42)
	c := NewCommitment(value, blinding)
	assert.NotEqual(t, Commitment{}, c)

	// Same inputs -> same commitment.
	assert.Equal(t, c, NewCommitment(value, blinding))
	// Different value -> different commitment.
	assert.NotEqual(t, c, NewCommitment(big.NewInt(101), blinding))
	// Open succeeds with correct opening.
	assert.True(t, c.Open(value, blinding))
	assert.False(t, c.Open(big.NewInt(101), blinding))
}

func TestRangeProof(t *testing.T) {
	value := big.NewInt(100)
	blinding := big.NewInt(42)
	commitment := NewCommitment(value, blinding)
	proof := NewRangeProof(value, blinding)
	assert.NotEqual(t, RangeProof{}, proof)
	assert.True(t, proof.Verify(commitment))

	// Wrong commitment -> proof fails.
	assert.False(t, proof.Verify(NewCommitment(big.NewInt(200), blinding)))
	// Empty proof -> invalid.
	assert.False(t, RangeProof{}.Verify(commitment))
}

func TestCiphertext(t *testing.T) {
	amount := big.NewInt(5000)
	pubKey := []byte("recipient-view-key")
	ct := EncryptAmount(amount, pubKey)
	assert.NotEqual(t, Ciphertext{}, ct)
	// Same inputs -> same ciphertext (deterministic placeholder).
	assert.Equal(t, ct, EncryptAmount(amount, pubKey))
	// Different amount -> different ciphertext.
	assert.NotEqual(t, ct, EncryptAmount(big.NewInt(6000), pubKey))
}

func TestConfidentialBalance(t *testing.T) {
	amount := big.NewInt(1000)
	blinding := big.NewInt(7)
	viewKey := []byte("view-key-1")
	cb := NewConfidentialBalance(amount, blinding, viewKey)
	assert.NotNil(t, cb.ViewKey)
	assert.True(t, cb.Verify())

	// Nil balance -> invalid.
	assert.False(t, (&ConfidentialBalance{}).Verify())
}

func TestConfidentialTransfer(t *testing.T) {
	sender := []byte("sender-1")
	recipient := []byte("recipient-1")
	amount := big.NewInt(500)
	senderBlinding := big.NewInt(1)
	recipientBlinding := big.NewInt(2)
	viewKey := []byte("view-key-1")

	ct := NewConfidentialTransfer(sender, recipient, amount, senderBlinding, recipientBlinding, viewKey)
	assert.Equal(t, sender, ct.Sender)
	assert.Equal(t, recipient, ct.Recipient)
	assert.NotEqual(t, []byte{}, ct.Nonce)
	assert.NotEqual(t, []byte{}, ct.Signature)

	// Placeholder Verify: non-empty commitments/proofs -> valid.
	assert.True(t, ct.Verify())

	// Missing sender -> invalid.
	bad := &ConfidentialTransfer{
		Recipient:           recipient,
		EncryptedAmount:     ct.EncryptedAmount,
		SenderCommitment:    ct.SenderCommitment,
		RecipientCommitment: ct.RecipientCommitment,
		SenderRangeProof:    ct.SenderRangeProof,
		RecipientRangeProof: ct.RecipientRangeProof,
		Nonce:               ct.Nonce,
		Signature:           ct.Signature,
	}
	assert.False(t, bad.Verify())
}

func TestViewKeyDerivation(t *testing.T) {
	master := []byte("master-key")
	asset := []byte("asset-1")
	k1 := ViewKeyDerivation(master, asset)
	k2 := ViewKeyDerivation(master, asset)
	assert.Equal(t, k1, k2)
	assert.NotEqual(t, k1, ViewKeyDerivation(master, []byte("asset-2")))
	assert.NotEqual(t, k1, ViewKeyDerivation([]byte("master-2"), asset))
}
