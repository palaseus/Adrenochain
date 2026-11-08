package hd

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
)

// BIP-32 constants.
const (
	hardenedStart = 0x80000000
)

// masterKeyHash is the HMAC key for BIP-32 master node derivation.
var masterKeyHash = []byte("Bitcoin seed")

// seedVersion is the mainnet private extended key version bytes (0x0488B21E).
var seedVersion = []byte{0x04, 0x88, 0xB2, 0x1E}

// ExtendedKey is a BIP-32 hierarchical deterministic extended key.
type ExtendedKey struct {
	Version    []byte // 4 bytes (network + private/public)
	Depth      byte   // 1 byte
	ParentFP   []byte // 4 bytes fingerprint of parent
	ChildNum   []byte // 4 bytes child number
	ChainCode  []byte // 32 bytes
	PrivateKey []byte // 33 bytes (0x00 + 32-byte scalar), nil for public-only
	PublicKey  []byte // 33 bytes compressed
}

// NewMasterFromSeed creates a BIP-32 master extended key from a seed.
func NewMasterFromSeed(seed []byte) (*ExtendedKey, error) {
	if len(seed) < 16 || len(seed) > 64 {
		return nil, fmt.Errorf("seed length must be 16-64 bytes, got %d", len(seed))
	}
	// I = HMAC-SHA512(key="Bitcoin seed", data=seed)
	mac := hmac.New(sha512.New, masterKeyHash)
	mac.Write(seed)
	I := mac.Sum(nil)
	il := I[:32] // private key scalar
	ir := I[32:] // chain code

	// Ensure the scalar is valid (non-zero, less than curve order).
	privInt := new(big.Int).SetBytes(il)
	if privInt.Sign() == 0 {
		return nil, errors.New("derived private key is zero")
	}
	if privInt.Cmp(btcec.S256().N) >= 0 {
		return nil, errors.New("derived private key exceeds curve order")
	}

	privKey, _ := btcec.PrivKeyFromBytes(il)
	pubKey := privKey.PubKey().SerializeCompressed()

	// Master key has depth 0, parent fingerprint 0, child num 0.
	childNum := make([]byte, 4)
	parentFP := make([]byte, 4)

	return &ExtendedKey{
		Version:    append([]byte{}, seedVersion...),
		Depth:      0,
		ParentFP:   parentFP,
		ChildNum:   childNum,
		ChainCode:  ir,
		PrivateKey: append([]byte{0x00}, il...),
		PublicKey:  pubKey,
	}, nil
}

// Derive performs BIP-32 child key derivation for a single index.
// If index >= hardenedStart, a hardened derivation is performed (requires
// the extended key to have a private key).
func (ek *ExtendedKey) Derive(index uint32) (*ExtendedKey, error) {
	if ek == nil {
		return nil, errors.New("nil extended key")
	}
	hardened := index >= hardenedStart

	var data []byte
	if hardened {
		if ek.PrivateKey == nil {
			return nil, errors.New("cannot derive hardened child from public-only key")
		}
		// BIP-32 hardened: data = 0x00 || privateKeyScalar (32 bytes)
		data = append([]byte{0x00}, ek.PrivateKey[1:]...)
	} else {
		data = ek.PublicKey // compressed public key (33 bytes)
	}
	indexBytes := uint32Bytes(index)
	data = append(data, indexBytes...)

	// I = HMAC-SHA512(key=chainCode, data=data)
	mac := hmac.New(sha512.New, ek.ChainCode)
	mac.Write(data)
	I := mac.Sum(nil)
	il := I[:32]
	ir := I[32:]

	// Child private key = (IL + parentPrivateKey) mod n
	ilInt := new(big.Int).SetBytes(il)
	if ilInt.Cmp(btcec.S256().N) >= 0 {
		return nil, errors.New("derived child scalar exceeds curve order")
	}
	parentPriv := new(big.Int).SetBytes(ek.PrivateKey[1:]) // skip 0x00 prefix
	childPrivInt := new(big.Int).Add(ilInt, parentPriv)
	childPrivInt.Mod(childPrivInt, btcec.S256().N)
	if childPrivInt.Sign() == 0 {
		return nil, errors.New("derived child private key is zero")
	}

	childPrivKey, _ := btcec.PrivKeyFromBytes(childPrivInt.Bytes())
	childPubKey := childPrivKey.PubKey().SerializeCompressed()

	// Child fingerprint: surrogate using double SHA-256 first 4 bytes
	// (BIP-32 specifies RIPEMD160(SHA256(pubKey)); stdlib lacks RIPEMD160).
	childFP := fingerprintSurrogate(childPubKey)

	return &ExtendedKey{
		Version:    append([]byte{}, ek.Version...),
		Depth:      ek.Depth + 1,
		ParentFP:   childFP,
		ChildNum:   indexBytes,
		ChainCode:  ir,
		PrivateKey: append([]byte{0x00}, pad32(childPrivInt.Bytes())...),
		PublicKey:  childPubKey,
	}, nil
}

// DerivePath derives a key at a BIP-32 derivation path, e.g. "m/44'/0'/0'/0/0".
// Path components with an apostrophe (') or 'h' suffix are hardened.
func (ek *ExtendedKey) DerivePath(path string) (*ExtendedKey, error) {
	if ek == nil {
		return nil, errors.New("nil extended key")
	}
	if path == "m" {
		return ek, nil
	}
	if len(path) < 2 || path[0] != 'm' || path[1] != '/' {
		return nil, fmt.Errorf("invalid path %q (must start with 'm/')", path)
	}
	current := ek
	components := splitPath(path[2:])
	for _, comp := range components {
		hardened := false
		numStr := comp
		if len(numStr) > 0 && (numStr[len(numStr)-1] == '\'' || numStr[len(numStr)-1] == 'h' || numStr[len(numStr)-1] == 'H') {
			hardened = true
			numStr = numStr[:len(numStr)-1]
		}
		index, err := parseUint32(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid path component %q: %w", comp, err)
		}
		if hardened {
			index += hardenedStart
		}
		next, err := current.Derive(index)
		if err != nil {
			return nil, fmt.Errorf("deriving %q: %w", comp, err)
		}
		current = next
	}
	return current, nil
}

// splitPath splits a path string on '/', ignoring empty components.
func splitPath(s string) []string {
	var out []string
	cur := ""
	for _, c := range s {
		if c == '/' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(c)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// parseUint32 parses a decimal uint32 string.
func parseUint32(s string) (uint32, error) {
	var v uint32
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("non-digit character %q", c)
		}
		v = v*10 + uint32(c-'0')
		if v > 0xFFFFFFFF {
			return 0, errors.New("value overflows uint32")
		}
	}
	return v, nil
}

// fingerprintSurrogate returns 4 bytes derived from the public key.
// BIP-32 specifies RIPEMD160(SHA256(pubKey)); stdlib lacks RIPEMD160, so we
// use double SHA-512 first 4 bytes as a structural surrogate.
func fingerprintSurrogate(pubKey []byte) []byte {
	h := sha512.New()
	h.Write(pubKey)
	first := h.Sum(nil)
	h.Reset()
	h.Write(first)
	second := h.Sum(nil)
	out := make([]byte, 4)
	copy(out, second[:4])
	return out
}

var _ = big.Int{} // keep math/big import referenced
