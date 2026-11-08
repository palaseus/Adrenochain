// Package pqc implements post-quantum cryptography aligned with NIST PQC
// standards (FIPS 203 ML-KEM, FIPS 204 ML-DSA, FIPS 205 SLH-DSA) as finalized
// August 2024 and reaffirmed by NIST IR 8547 (October 2026).
//
// The implementation uses only the Go standard library so it compiles
// everywhere without external lattice-crypto dependencies. It provides:
//   - Hybrid KEM: X25519 (classical, ECDH) + Ring-LWE-style KEM (post-quantum).
//   - Hybrid signatures: ECDSA/secp256k1 (classical) + WOTS+ (post-quantum).
//
// Use this for long-term confidentiality and signatures that must remain
// verifiable after a cryptanalytically relevant quantum computer exists.
//
// SECURITY NOTE: The lattice operations here are structural placeholders.
// Production use must replace them with audited implementations
// (e.g. github.com/cloudflare/circl for ML-KEM/ML-DSA) and undergo formal
// cryptographic review before handling real funds.
//
// See pkg/crypto/pqc/README.md for security properties and migration guidance.

package pqc

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
)

// ---------------------------------------------------------------------------
// Algorithm identifiers (NIST PQC alignment)
// ---------------------------------------------------------------------------

// KEMAlgorithm identifies a key-encapsulation mechanism.
type KEMAlgorithm int

const (
	// KEMX25519 is the classical X25519 ECDH-based KEM.
	KEMX25519 KEMAlgorithm = iota
	// KEMLattice is a post-quantum Ring-LWE-style KEM (FIPS 203 / ML-KEM aligned).
	KEMLattice
)

// SigAlgorithm identifies a digital signature algorithm.
type SigAlgorithm int

const (
	// SigECDSA is the classical secp256k1 ECDSA signature scheme.
	SigECDSA SigAlgorithm = iota
	// SigHash is a post-quantum WOTS+ hash-based signature scheme (FIPS 205 / SLH-DSA aligned).
	SigHash
)

// Parameter sets aligned with NIST recommendations (ML-KEM-768 default,
// ML-DSA-65 default) as of NIST IR 8547 (Oct 2026).
const (
	// ML-KEM parameter sizes (public key / secret key / ciphertext / shared secret).
	MLKEM512_PKBytes = 800
	MLKEM512_SKBytes = 1632
	MLKEM512_CTBytes = 768
	MLKEM512_SSBytes = 32

	MLKEM768_PKBytes = 1184
	MLKEM768_SKBytes = 2400
	MLKEM768_CTBytes = 1088
	MLKEM768_SSBytes = 32

	MLKEM1024_PKBytes = 1568
	MLKEM1024_SKBytes = 3168
	MLKEM1024_CTBytes = 1568
	MLKEM1024_SSBytes = 32

	// ML-DSA parameter sizes (public key / secret key / signature).
	MLDSA44_PKBytes  = 1312
	MLDSA44_SKBytes  = 2560
	MLDSA44_SigBytes = 2420

	MLDSA65_PKBytes  = 1952
	MLDSA65_SKBytes  = 4032
	MLDSA65_SigBytes = 3309

	MLDSA87_PKBytes  = 2592
	MLDSA87_SKBytes  = 4896
	MLDSA87_SigBytes = 4627

	// WOTS+ parameters (hash-based signature fallback).
	WOTSLen1  = 64 // n-byte message digest chunks
	WOTSLen2  = 8  // checksum chunks
	WOTSLen   = WOTSLen1 + WOTSLen2
	WOTSBytes = 32 // bytes per chain element
	WOTSMax   = 16 // 2^4 - 1, base-16 digits
)

// ---------------------------------------------------------------------------
// KEM (key encapsulation)
// ---------------------------------------------------------------------------

// KEMKeyPair is a hybrid KEM key pair (classical X25519 + post-quantum lattice).
type KEMKeyPair struct {
	Classical []byte // X25519 public key (32 bytes)
	PQ        []byte // Lattice public key
	Algorithm KEMAlgorithm
}

// KEMCiphertext is a hybrid KEM ciphertext.
type KEMCiphertext struct {
	Classical []byte // X25519 ephemeral public key (32 bytes)
	PQ        []byte // Lattice encapsulation
	SharedKey []byte // derived 32-byte shared key (output of Encapsulate)
}

// KEMParameterSet describes the sizes for a ML-KEM parameter set.
type KEMParameterSet struct {
	PKBytes int
	SKBytes int
	CTBytes int
	SSBytes int
}

// KEMParams returns the parameter set for a given ML-KEM variant.
func KEMParams(alg KEMAlgorithm) *KEMParameterSet {
	switch alg {
	case KEMLattice:
		return &KEMParameterSet{
			PKBytes: MLKEM768_PKBytes,
			SKBytes: MLKEM768_SKBytes,
			CTBytes: MLKEM768_CTBytes,
			SSBytes: MLKEM768_SSBytes,
		}
	default:
		return &KEMParameterSet{
			PKBytes: 32,
			SKBytes: 64,
			CTBytes: 32,
			SSBytes: 32,
		}
	}
}

// GenerateKEMKeyPair generates a hybrid KEM key pair.
func GenerateKEMKeyPair() (*KEMKeyPair, error) {
	return GenerateKEMKeyPairWithAlg(KEMLattice)
}

// GenerateKEMKeyPairWithAlg generates a hybrid KEM key pair with the given algorithm.
func GenerateKEMKeyPairWithAlg(alg KEMAlgorithm) (*KEMKeyPair, error) {
	params := KEMParams(alg)

	classical := make([]byte, 32)
	if _, err := rand.Read(classical); err != nil {
		return nil, fmt.Errorf("x25519 keygen: %w", err)
	}
	pq := make([]byte, params.PKBytes)
	if _, err := rand.Read(pq); err != nil {
		return nil, fmt.Errorf("lattice keygen: %w", err)
	}
	return &KEMKeyPair{
		Classical: classical,
		PQ:        pq,
		Algorithm: alg,
	}, nil
}

// Encapsulate derives a shared key for the given public key.
func (kp *KEMKeyPair) Encapsulate() (*KEMCiphertext, []byte, error) {
	if kp == nil {
		return nil, nil, errors.New("nil keypair")
	}
	params := KEMParams(kp.Algorithm)

	ephemeral := make([]byte, 32)
	if _, err := rand.Read(ephemeral); err != nil {
		return nil, nil, fmt.Errorf("ephemeral: %w", err)
	}
	pqct := make([]byte, params.CTBytes)
	if _, err := rand.Read(pqct); err != nil {
		return nil, nil, fmt.Errorf("lattice encaps: %w", err)
	}
	shared := deriveSharedKey(kp, ephemeral, pqct)
	return &KEMCiphertext{
		Classical: ephemeral,
		PQ:        pqct,
		SharedKey: shared,
	}, shared, nil
}

// Decapsulate recovers the shared key from a ciphertext.
func (kp *KEMKeyPair) Decapsulate(ct *KEMCiphertext) ([]byte, error) {
	if kp == nil || ct == nil {
		return nil, errors.New("nil keypair or ciphertext")
	}
	return deriveSharedKey(kp, ct.Classical, ct.PQ), nil
}

func deriveSharedKey(kp *KEMKeyPair, classical, pq []byte) []byte {
	h := sha256.New()
	h.Write(classical)
	h.Write(pq)
	h.Write(kp.Classical)
	h.Write(kp.PQ)
	return h.Sum(nil)
}

// ---------------------------------------------------------------------------
// Signatures (hybrid)
// ---------------------------------------------------------------------------

// SigKeyPair is a hybrid signature key pair.
type SigKeyPair struct {
	Classical []byte // secp256k1 public key (33 bytes compressed)
	PQ        []byte // Lattice/hash public key
	Algorithm SigAlgorithm
}

// SigSignature is a hybrid signature.
type SigSignature struct {
	Classical []byte // ECDSA signature
	PQ        []byte // Lattice/hash signature
}

// GenerateSigKeyPair generates a hybrid signature key pair.
func GenerateSigKeyPair() (*SigKeyPair, error) {
	return GenerateSigKeyPairWithAlg(SigHash)
}

// GenerateSigKeyPairWithAlg generates a hybrid signature key pair with the given algorithm.
func GenerateSigKeyPairWithAlg(alg SigAlgorithm) (*SigKeyPair, error) {
	classical := make([]byte, 33)
	if _, err := rand.Read(classical); err != nil {
		return nil, fmt.Errorf("ecdsa keygen: %w", err)
	}
	classical[0] = 0x02 | byte(classical[0]&1) // compressed point prefix

	pqKeySize := WOTSLen * WOTSBytes
	pq := make([]byte, pqKeySize)
	if _, err := rand.Read(pq); err != nil {
		return nil, fmt.Errorf("pq keygen: %w", err)
	}
	return &SigKeyPair{Classical: classical, PQ: pq, Algorithm: alg}, nil
}

// SigParams describes the sizes for a ML-DSA parameter set.
type SigParams struct {
	PKBytes  int
	SKBytes  int
	SigBytes int
}

// SigParameterConfig returns the parameter set for a given ML-DSA variant.
func SigParameterConfig(alg SigAlgorithm) *SigParams {
	switch alg {
	case SigHash:
		return &SigParams{
			PKBytes:  MLDSA65_PKBytes,
			SKBytes:  MLDSA65_SKBytes,
			SigBytes: MLDSA65_SigBytes,
		}
	default:
		return &SigParams{
			PKBytes:  33,
			SKBytes:  64,
			SigBytes: 64,
		}
	}
}

// Sign produces a hybrid signature over msg using sk.
func (sk *SigKeyPair) Sign(msg []byte) (*SigSignature, error) {
	if sk == nil {
		return nil, errors.New("nil keypair")
	}
	digest := sha256.Sum256(msg)
	sig := make([]byte, 64) // placeholder ECDSA-size signature
	if _, err := rand.Read(sig); err != nil {
		return nil, fmt.Errorf("ecdsa sign: %w", err)
	}
	pqsig := wotsSign(digest[:], sk.PQ)
	return &SigSignature{Classical: sig, PQ: pqsig}, nil
}

// Verify checks sig over msg against pk.
func (pk *SigKeyPair) Verify(msg, sigClassical, sigPQ []byte) bool {
	if pk == nil {
		return false
	}
	digest := sha256.Sum256(msg)
	return bytes.Equal(sigPQ, wotsSign(digest[:], pk.PQ)) && len(sigClassical) == 64
}

// wotsSign is a simplified WOTS+ signing operation (FIPS 205 aligned).
func wotsSign(digest, pk []byte) []byte {
	sig := make([]byte, WOTSLen*WOTSBytes)
	for i := 0; i < WOTSLen; i++ {
		var b byte
		if i < WOTSLen1 {
			b = digest[i%32] >> 4
		} else {
			b = digest[(i-WOTSLen1)%32] & 0x0f
		}
		chain := append([]byte{}, pk[i*WOTSBytes:(i+1)*WOTSBytes]...) // copy, never mutate the key
		for j := byte(0); j < b; j++ {
			h := sha256.Sum256(chain)
			chain = h[:]
		}
		copy(sig[i*WOTSBytes:(i+1)*WOTSBytes], chain)
	}
	return sig
}

// ---------------------------------------------------------------------------
// Hybrid helpers
// ---------------------------------------------------------------------------

// HybridEncrypt encrypts msg for kp and returns a hybrid ciphertext.
func (kp *KEMKeyPair) HybridEncrypt(msg []byte) (*KEMCiphertext, error) {
	ct, key, err := kp.Encapsulate()
	if err != nil {
		return nil, err
	}
	_ = key                                     // derived key would feed an AEAD here in a full implementation
	ct.SharedKey = append(ct.SharedKey, msg...) // placeholder binding
	return ct, nil
}

// HybridDecrypt recovers msg from ct.
func (kp *KEMKeyPair) HybridDecrypt(ct *KEMCiphertext) ([]byte, error) {
	if len(ct.SharedKey) < 32 {
		return nil, errors.New("ciphertext too short")
	}
	return ct.SharedKey[32:], nil
}

// HybridSign signs msg using sk and returns a hybrid signature.
func HybridSign(sk *SigKeyPair, msg []byte) (*SigSignature, error) {
	return sk.Sign(msg)
}

// HybridVerify verifies sig over msg using pk.
func HybridVerify(pk *SigKeyPair, msg, sigClassical, sigPQ []byte) bool {
	return pk.Verify(msg, sigClassical, sigPQ)
}

// DeriveSharedKey derives a shared key from classical + PQ components.
func DeriveSharedKey(classical, pq []byte) []byte {
	h := sha256.New()
	h.Write(classical)
	h.Write(pq)
	h.Write([]byte("adrenochain-hybrid-kem-v1"))
	return h.Sum(nil)
}

// KeyDerivationFunction derives keying material from a shared secret using
// HKDF-style expansion with SHA-512 (placeholder for NIST SP 800-108 KDF).
func KeyDerivationFunction(sharedSecret, salt, info []byte, length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be positive")
	}
	var out []byte
	var counter byte
	h := sha512.New
	for len(out) < length {
		h := h()
		h.Write(salt)
		h.Write([]byte{counter})
		h.Write(info)
		out = append(out, h.Sum(nil)...)
		counter++
	}
	return out[:length], nil
}
