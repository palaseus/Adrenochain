package solana

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

// ed25519 curve parameters.
var (
	ed25519P = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19)) // 2^255 - 19
	// d = -121665 / 121666 mod p
	ed25519D = func() *big.Int {
		d := new(big.Int).ModInverse(big.NewInt(121666), ed25519P)
		d.Mul(d, big.NewInt(-121665))
		d.Mod(d, ed25519P)
		return d
	}()
)

// DeriveProgramAddress derives a Program Derived Address (PDA) from seeds and a
// program ID, following the Solana PDA algorithm:
//
//	hashed_seeds = SHA256(seed1 || seed2 || ... || seedN)
//	for bump from 255 down to 0:
//	    candidate = SHA256(program_id || hashed_seeds || [bump])
//	    if candidate is NOT on the ed25519 curve:
//	        return (candidate, bump)
//
// The resulting PDA is off the ed25519 curve, so no private key exists for it.
// Only the owning program can authorize operations via invoke_signed.
func DeriveProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, uint8, error) {
	if len(seeds) == 0 {
		return PublicKey{}, 0, errors.New("at least one seed is required")
	}
	if len(seeds) > MaxSeeds {
		return PublicKey{}, 0, fmt.Errorf("too many seeds: %d (max %d)", len(seeds), MaxSeeds)
	}
	for i, seed := range seeds {
		if len(seed) > SeedMaxLen {
			return PublicKey{}, 0, fmt.Errorf("seed %d too long: %d (max %d bytes)", i, len(seed), SeedMaxLen)
		}
	}

	// hashed_seeds = SHA256(seed1 || seed2 || ...)
	hashedSeeds := sha256.Sum256(concatenateSeeds(seeds))

	// Try bumps from 255 down to 0 (canonical bump is the highest valid one).
	for bump := byte(255); ; bump-- {
		candidate := derivePDACandidate(programID, hashedSeeds[:], bump)
		onCurve, err := isEd25519OnCurve(candidate)
		if err != nil {
			return PublicKey{}, 0, err
		}
		if !onCurve {
			return NewPublicKeyFromEd25519(candidate), bump, nil
		}
		if bump == 0 {
			return PublicKey{}, 0, errors.New("exhausted all bump seeds; no valid PDA found")
		}
	}
}

// concatenateSeeds joins all seed bytes into a single buffer.
func concatenateSeeds(seeds [][]byte) []byte {
	total := 0
	for _, s := range seeds {
		total += len(s)
	}
	out := make([]byte, 0, total)
	for _, s := range seeds {
		out = append(out, s...)
	}
	return out
}

// derivePDACandidate computes SHA256(program_id || hashed_seeds || bump).
func derivePDACandidate(programID PublicKey, hashedSeeds []byte, bump byte) []byte {
	h := sha256.New()
	h.Write(programID[:])
	h.Write(hashedSeeds)
	h.Write([]byte{bump})
	return h.Sum(nil)
}

// isEd25519OnCurve reports whether the given 32-byte value encodes a point on
// the ed25519 curve. Uses the curve equation:
//
//	-x^2 + y^2 = 1 + d*x^2*y^2   (mod p)
//
// where the compressed point encodes y (with the top bit as the sign of x).
func isEd25519OnCurve(encoded []byte) (bool, error) {
	if len(encoded) != 32 {
		return false, errors.New("encoded point must be 32 bytes")
	}

	// Decode y: low 255 bits; top bit is the sign of x (ignored for curve check).
	yBytes := make([]byte, 32)
	copy(yBytes, encoded)
	yBytes[0] &^= 0x80 // clear the sign bit
	y := new(big.Int).SetBytes(yBytes)

	// y must be < p.
	if y.Cmp(ed25519P) >= 0 {
		return false, nil
	}

	// x^2 = (y^2 - 1) / (d*y^2 + 1) mod p
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, ed25519P)

	numerator := new(big.Int).Sub(y2, big.NewInt(1))
	numerator.Mod(numerator, ed25519P)

	denominator := new(big.Int).Mul(ed25519D, y2)
	denominator.Add(denominator, big.NewInt(1))
	denominator.Mod(denominator, ed25519P)

	denominatorInv := new(big.Int).ModInverse(denominator, ed25519P)
	if denominatorInv == nil {
		return false, nil
	}

	x2 := new(big.Int).Mul(numerator, denominatorInv)
	x2.Mod(x2, ed25519P)

	// x^2 is a quadratic residue iff x^2^((p-1)/2) == 1 mod p (or 0).
	// p = 2^255 - 19, so (p-1)/2 = 2^254 - 10.
	euler := new(big.Int).Exp(x2, new(big.Int).Sub(new(big.Int).Rsh(ed25519P, 1), big.NewInt(1)), ed25519P)

	return euler.Cmp(big.NewInt(1)) == 0 || euler.Cmp(big.NewInt(0)) == 0, nil
}

var _ = errors.New
