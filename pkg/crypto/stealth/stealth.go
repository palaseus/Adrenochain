// Package stealth implements ERC-5564 stealth addresses: private, non-interactive
// transfers where a sender generates a one-time destination address for a recipient
// without prior interaction, and only the recipient can recognize/redeem funds.
//
// Aligned with ERC-5564 (final, Standards Track, scheme 1 = secp256k1) and
// ERC-6538 (stealth meta-address registry). EC math via
// github.com/decred/dcrd/dcrec/secp256k1/v4 (re-exported as btcec/v2).
//
// Protocol (sender):
//   ephemeral key (e_priv, e_pub)
//   spend_shared_x   = ECDH(e_priv, spend_pub)
//   spend_shared     = SHA512(e_pub || spend_pub || spend_shared_x)
//   scalar           = spend_shared[0:32] mod n
//   stealth_spend    = spend_pub + scalar * G
//   view_shared_x    = ECDH(e_priv, view_pub)
//   view_shared      = SHA512(e_pub || view_pub || view_shared_x)
//   view_tag         = view_shared[32]
//
// Recipient scans with the viewing key; recovers spend authority with the
// spending key.
//
// See pkg/crypto/stealth/README.md for the protocol flow and privacy properties.

package stealth

import (
	"crypto/sha512"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Constants aligned with ERC-5564 scheme 1 (secp256k1).
const (
	SchemeID           = 1
	SpendPubKeyLen     = 33 // compressed secp256k1 public key
	ViewPubKeyLen      = 33
	StealthAddrLen     = 33
	ViewTagLen         = 1
	MetaAddressLen     = SpendPubKeyLen + ViewPubKeyLen // 66
	EphemeralPubKeyLen = 33
)

// MetaAddress is a recipient's stealth meta-address (spending + viewing pub keys).
type MetaAddress struct {
	SpendPubKey []byte // compressed
	ViewPubKey  []byte // compressed
}

// EphemeralKey is the sender's one-time ECDH key for a stealth transfer.
type EphemeralKey struct {
	PrivKey *btcec.PrivateKey
	PubKey  []byte // compressed
}

// StealthAddress is the one-time destination derived for a recipient.
type StealthAddress struct {
	Address []byte // compressed stealth spending pub key
	ViewTag byte
}

// Announcement is on-chain data attached to a stealth transfer so the recipient
// can locate it without interaction.
type Announcement struct {
	EphemeralPubKey []byte // compressed
	StealthAddress  []byte // compressed stealth address
	ViewTag         byte
}

// GenerateEphemeralKey creates a fresh sender ephemeral key pair.
func GenerateEphemeralKey() (*EphemeralKey, error) {
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("ephemeral keygen: %w", err)
	}
	return &EphemeralKey{
		PrivKey: priv,
		PubKey:  priv.PubKey().SerializeCompressed(),
	}, nil
}

// MetaAddressFromBytes parses spend||view (66 bytes).
func MetaAddressFromBytes(data []byte) (*MetaAddress, error) {
	if len(data) != MetaAddressLen {
		return nil, fmt.Errorf("expected %d bytes, got %d", MetaAddressLen, len(data))
	}
	return &MetaAddress{
		SpendPubKey: append([]byte{}, data[:SpendPubKeyLen]...),
		ViewPubKey:  append([]byte{}, data[SpendPubKeyLen:]...),
	}, nil
}

// ToBytes serializes the meta-address as spend||view.
func (m *MetaAddress) ToBytes() []byte {
	out := make([]byte, 0, MetaAddressLen)
	out = append(out, m.SpendPubKey...)
	out = append(out, m.ViewPubKey...)
	return out
}

// GenerateMetaAddress creates a fresh recipient meta-address.
func GenerateMetaAddress() (*MetaAddress, *btcec.PrivateKey, *btcec.PrivateKey, error) {
	spend, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, nil, err
	}
	view, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, nil, err
	}
	return &MetaAddress{
		SpendPubKey: spend.PubKey().SerializeCompressed(),
		ViewPubKey:  view.PubKey().SerializeCompressed(),
	}, spend, view, nil
}

// ecdhSharedX returns the x-coordinate of priv * peerPub (EC Diffie-Hellman).
func ecdhSharedX(priv *btcec.PrivateKey, peerPub []byte) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("nil private key")
	}
	pub, err := btcec.ParsePubKey(peerPub)
	if err != nil {
		return nil, fmt.Errorf("parse peer pubkey: %w", err)
	}
	sharedX := secp256k1.GenerateSharedSecret(priv, pub)
	if len(sharedX) == 0 {
		return nil, errors.New("ecdh produced empty shared secret")
	}
	return sharedX, nil
}

// deriveShared hashes the transcript into 64 bytes:
//
//	[0:32] scalar material, [32] view-tag material, rest reserved.
func deriveShared(ephemeralPub, recipientPub, sharedX []byte) []byte {
	h := sha512.New()
	h.Write(ephemeralPub)
	h.Write(recipientPub)
	h.Write(sharedX)
	return h.Sum(nil)
}

// GenerateStealthAddress derives the one-time stealth address for a recipient.
func GenerateStealthAddress(meta *MetaAddress, ephem *EphemeralKey) (*StealthAddress, *Announcement, error) {
	if meta == nil || ephem == nil || ephem.PrivKey == nil {
		return nil, nil, errors.New("nil meta or ephemeral")
	}

	spendSharedX, err := ecdhSharedX(ephem.PrivKey, meta.SpendPubKey)
	if err != nil {
		return nil, nil, err
	}
	spendShared := deriveShared(ephem.PubKey, meta.SpendPubKey, spendSharedX)

	// scalar = spend_shared[0:32] reduced mod n (little-endian per ERC-5564)
	scalar := secp256k1.PrivKeyFromBytes(spendShared[:32])

	spend, err := btcec.ParsePubKey(meta.SpendPubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("parse spend: %w", err)
	}

	// offset = scalar * G
	offset := scalar.PubKey()

	// stealth = spend + offset via Jacobian point addition
	var spendJ, offsetJ, resultJ secp256k1.JacobianPoint
	spend.AsJacobian(&spendJ)
	offset.AsJacobian(&offsetJ)
	secp256k1.AddNonConst(&spendJ, &offsetJ, &resultJ)
	resultJ.ToAffine()
	stealth := secp256k1.NewPublicKey(&resultJ.X, &resultJ.Y)
	stealthAddr := stealth.SerializeCompressed()

	// view_tag from the viewing-key shared secret
	viewSharedX, err := ecdhSharedX(ephem.PrivKey, meta.ViewPubKey)
	if err != nil {
		return nil, nil, err
	}
	viewShared := deriveShared(ephem.PubKey, meta.ViewPubKey, viewSharedX)
	viewTag := viewShared[32]

	ann := &Announcement{
		EphemeralPubKey: ephem.PubKey,
		StealthAddress:  stealthAddr,
		ViewTag:         viewTag,
	}
	return &StealthAddress{Address: stealthAddr, ViewTag: viewTag}, ann, nil
}

// Scan checks whether an announcement is addressed to this recipient by
// recomputing the view tag from the viewing key (off-chain scanning, ERC-5564).
func (m *MetaAddress) Scan(ann *Announcement, viewPriv *btcec.PrivateKey) bool {
	if m == nil || ann == nil || viewPriv == nil {
		return false
	}
	viewSharedX, err := ecdhSharedX(viewPriv, ann.EphemeralPubKey)
	if err != nil {
		return false
	}
	viewShared := deriveShared(ann.EphemeralPubKey, m.ViewPubKey, viewSharedX)
	return viewShared[32] == ann.ViewTag
}

// RecoverSpendPrivateKey derives the spending private key for a stealth address
// so the recipient can spend funds sent to it.
func RecoverSpendPrivateKey(meta *MetaAddress, spendPriv *btcec.PrivateKey, ann *Announcement) (*btcec.PrivateKey, error) {
	if meta == nil || spendPriv == nil || ann == nil {
		return nil, errors.New("nil arguments")
	}
	spendSharedX, err := ecdhSharedX(spendPriv, ann.EphemeralPubKey)
	if err != nil {
		return nil, err
	}
	spendShared := deriveShared(ann.EphemeralPubKey, meta.SpendPubKey, spendSharedX)

	// stealth_priv = spend_priv + scalar(spend_shared[0:32]) (mod n)
	scalar := secp256k1.PrivKeyFromBytes(spendShared[:32])
	priv := new(secp256k1.ModNScalar)
	priv.Set(&spendPriv.Key)
	priv.Add(&scalar.Key)
	privKey := secp256k1.NewPrivateKey(priv)
	return (*btcec.PrivateKey)(privKey), nil
}
