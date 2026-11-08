package stealth

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

func TestStealthRoundTrip(t *testing.T) {
	// 1. Recipient generates meta-address + keys.
	meta, spendPriv, viewPriv, err := GenerateMetaAddress()
	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Len(t, meta.SpendPubKey, SpendPubKeyLen)
	assert.Len(t, meta.ViewPubKey, ViewPubKeyLen)

	// Serialization round-trip.
	rt, err := MetaAddressFromBytes(meta.ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, meta.SpendPubKey, rt.SpendPubKey)
	assert.Equal(t, meta.ViewPubKey, rt.ViewPubKey)

	// 2. Sender generates ephemeral key and stealth address.
	ephem, err := GenerateEphemeralKey()
	assert.NoError(t, err)
	assert.Len(t, ephem.PubKey, EphemeralPubKeyLen)

	stealth, ann, err := GenerateStealthAddress(meta, ephem)
	assert.NoError(t, err)
	assert.NotNil(t, stealth)
	assert.NotNil(t, ann)
	assert.Len(t, stealth.Address, StealthAddrLen)
	assert.Len(t, ann.EphemeralPubKey, EphemeralPubKeyLen)
	assert.Len(t, ann.StealthAddress, StealthAddrLen)
	assert.Equal(t, stealth.Address, ann.StealthAddress)

	// 3. Recipient scans and recognizes the announcement.
	assert.True(t, meta.Scan(ann, viewPriv), "recipient should recognize own announcement")

	// Wrong viewing key should not match.
	wrongMeta, _, wrongView, err := GenerateMetaAddress()
	assert.NoError(t, err)
	assert.False(t, wrongMeta.Scan(ann, wrongView), "unrelated recipient should not match")

	// 4. Recipient recovers the spend private key for the stealth address.
	recovered, err := RecoverSpendPrivateKey(meta, spendPriv, ann)
	assert.NoError(t, err)
	assert.NotNil(t, recovered)

	// Recovered key's public key should equal the stealth address.
	recoveredPub := recovered.PubKey().SerializeCompressed()
	assert.Equal(t, stealth.Address, recoveredPub)

	// The ephemeral public key in the announcement lets anyone send; only the
	// recipient with spendPriv can derive the matching private key.
	assert.Equal(t, ann.EphemeralPubKey, ephem.PubKey)
}

func TestStealthRejectsNil(t *testing.T) {
	meta, _, _, err := GenerateMetaAddress()
	assert.NoError(t, err)
	ephem, err := GenerateEphemeralKey()
	assert.NoError(t, err)

	_, _, err = GenerateStealthAddress(nil, ephem)
	assert.Error(t, err)
	_, _, err = GenerateStealthAddress(meta, nil)
	assert.Error(t, err)
	assert.False(t, meta.Scan(nil, ephem.PrivKey))
	_, err = RecoverSpendPrivateKey(nil, ephem.PrivKey, &Announcement{})
	assert.Error(t, err)
}

var _ = btcec.S256 // ensure btcec used in non-test helpers
