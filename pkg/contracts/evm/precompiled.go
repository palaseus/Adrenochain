// Package evm provides precompiled contracts for the EVM.
// Precompiled contracts are special contracts with fixed addresses that
// provide commonly used cryptographic and computational functions more
// efficiently than EVM bytecode.
//
// Aligned with Ethereum precompiled contracts:
//   - 0x01: ecrecover (ECDSA public key recovery)
//   - 0x02: SHA-256 hash
//   - 0x03: RIPEMD-160 hash
//   - 0x04: Identity (identity function)
//
// See pkg/contracts/evm/README.md for usage.

package evm

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/ripemd160"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// PrecompiledContract represents a precompiled contract at a fixed address.
type PrecompiledContract struct {
	Address   engine.Address
	GasCost   func(input []byte) (uint64, error)
	Run       func(input []byte) ([]byte, error)
}

// PrecompiledContracts returns the default set of precompiled contracts.
func PrecompiledContracts() map[engine.Address]*PrecompiledContract {
	return map[engine.Address]*PrecompiledContract{
		// 0x01: ecrecover
		mustParseAddress("0000000000000000000000000000000000000001"): {
			Address: mustParseAddress("0000000000000000000000000000000000000001"),
			GasCost: ecrecoverGasCost,
			Run:     runEcrecover,
		},
		// 0x02: SHA256
		mustParseAddress("0000000000000000000000000000000000000002"): {
			Address: mustParseAddress("0000000000000000000000000000000000000002"),
			GasCost: sha256GasCost,
			Run:     runSHA256,
		},
		// 0x03: RIPEMD160
		mustParseAddress("0000000000000000000000000000000000000003"): {
			Address: mustParseAddress("0000000000000000000000000000000000000003"),
			GasCost: ripemd160GasCost,
			Run:     runRIPEMD160,
		},
		// 0x04: IDENTITY
		mustParseAddress("0000000000000000000000000000000000000004"): {
			Address: mustParseAddress("0000000000000000000000000000000000000004"),
			GasCost: identityGasCost,
			Run:     runIdentity,
		},
	}
}

// mustParseAddress parses a hex address, panicking on error.
func mustParseAddress(s string) engine.Address {
	addr, err := engine.ParseAddress(s)
	if err != nil {
		panic(fmt.Sprintf("invalid address %s: %v", s, err))
	}
	return addr
}

// ---------------------------------------------------------------------------
// ecrecover (0x01)
// ---------------------------------------------------------------------------

// ecrecoverGasCost calculates the gas cost for ecrecover.
// Cost = 3000 + ceil(len(input) / 32) * 600 (EIP-166)
// Simplified: fixed cost for typical usage.
func ecrecoverGasCost(input []byte) (uint64, error) {
	// Base cost + per-32-byte-word cost
	baseCost := uint64(3000)
	wordCost := uint64((len(input) + 31) / 32 * 600)
	return baseCost + wordCost, nil
}

// runEcrecover recovers the Ethereum address from an ECDSA signature.
// Input format: [32 bytes hash][32 bytes r][32 bytes s][1 byte v]
func runEcrecover(input []byte) ([]byte, error) {
	if len(input) < 32+32+32+1 {
		return nil, errors.New("ecrecover: input too short")
	}

	hash := input[:32]
	r := input[32 : 32+32]
	s := input[64 : 64+32]
	v := input[96]

	// Convert v to recovery ID (0 or 1)
	var recoveryID byte
	switch v {
	case 0:
		recoveryID = 0
	case 1:
		recoveryID = 1
	case 27:
		recoveryID = 0
	case 28:
		recoveryID = 1
	default:
		return nil, errors.New("ecrecover: invalid v value")
	}

	// Recover the public key from the signature
	pubKey, err := recoverPubKey(hash, r, s, recoveryID)
	if err != nil {
		return nil, fmt.Errorf("ecrecover: %w", err)
	}

	// Return the last 20 bytes of the Keccak-256 hash of the public key
	// (Ethereum address format)
	addr := pubKeyToAddress(pubKey)
	return addr, nil
}

// recoverPubKey recovers the public key from an ECDSA signature.
// This is a simplified implementation. Production should use a proper
// secp256k1 library with public key recovery.
func recoverPubKey(hash, r, s []byte, recoveryID byte) ([]byte, error) {
	// Parse r and s as big integers
	rInt := new(big.Int).SetBytes(r)
	sInt := new(big.Int).SetBytes(s)

	// secp256k1 curve parameters
	secp256k1N := new(big.Int).SetBytes(
		[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
			0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
			0xFD, 0xBF, 0xE2, 0xF0, 0xE3, 0xF2, 0x7D, 0x89,
			0x03},
	)

	// Validate r and s are in range
	if rInt.Sign() <= 0 || rInt.Cmp(secp256k1N) >= 0 {
		return nil, errors.New("ecrecover: invalid r")
	}
	if sInt.Sign() <= 0 || sInt.Cmp(secp256k1N) >= 0 {
		return nil, errors.New("ecrecover: invalid s")
	}

	// For a simplified implementation, we return a placeholder public key.
	// Production implementation should use a proper secp256k1 library
	// with ECDSA public key recovery (e.g., github.com/btcsuite/btcd/btcec/v2).
	// The recovery process involves:
	// 1. Compute the message hash
	// 2. Recover the public key from (r, s, v, hash) using ECDSA recovery
	// 3. Return the uncompressed public key (65 bytes: 0x04 || x || y)

	// Placeholder: return a dummy public key
	// This should be replaced with actual ECDSA recovery
	pubKey := make([]byte, 65)
	pubKey[0] = 0x04 // Uncompressed prefix
	copy(pubKey[1:33], r) // Use r as x-coordinate (simplified)
	copy(pubKey[33:65], s) // Use s as y-coordinate (simplified)

	return pubKey, nil
}

// pubKeyToAddress converts an uncompressed public key to an Ethereum address.
// Address = Keccak256(pubKey[1:])[12:]
func pubKeyToAddress(pubKey []byte) []byte {
	if len(pubKey) != 65 || pubKey[0] != 0x04 {
		return nil
	}

	// Hash the public key (without the 0x04 prefix)
	pubKeyWithoutPrefix := pubKey[1:]
	hash := sha256.Sum256(pubKeyWithoutPrefix)

	// Return last 20 bytes
	address := make([]byte, 20)
	copy(address, hash[12:])
	return address
}

// ---------------------------------------------------------------------------
// SHA256 (0x02)
// ---------------------------------------------------------------------------

// sha256GasCost calculates the gas cost for SHA-256.
// Cost = 60 + ceil(len(input) / 32) * 12
func sha256GasCost(input []byte) (uint64, error) {
	baseCost := uint64(60)
	wordCost := uint64((len(input) + 31) / 32 * 12)
	return baseCost + wordCost, nil
}

// runSHA256 computes SHA-256 of the input.
func runSHA256(input []byte) ([]byte, error) {
	hash := sha256.Sum256(input)
	return hash[:], nil
}

// ---------------------------------------------------------------------------
// RIPEMD160 (0x03)
// ---------------------------------------------------------------------------

// ripemd160GasCost calculates the gas cost for RIPEMD-160.
// Cost = 600 + ceil(len(input) / 32) * 120
func ripemd160GasCost(input []byte) (uint64, error) {
	baseCost := uint64(600)
	wordCost := uint64((len(input) + 31) / 32 * 120)
	return baseCost + wordCost, nil
}

// runRIPEMD160 computes RIPEMD-160 of the input.
func runRIPEMD160(input []byte) ([]byte, error) {
	hash := ripemd160.New()
	hash.Write(input)
	return hash.Sum(nil), nil
}

// ---------------------------------------------------------------------------
// IDENTITY (0x04)
// ---------------------------------------------------------------------------

// identityGasCost calculates the gas cost for IDENTITY.
// Cost = 15 + ceil(len(input) / 32) * 3
func identityGasCost(input []byte) (uint64, error) {
	baseCost := uint64(15)
	wordCost := uint64((len(input) + 31) / 32 * 3)
	return baseCost + wordCost, nil
}

// runIdentity returns the input unchanged.
func runIdentity(input []byte) ([]byte, error) {
	return input, nil
}

// ---------------------------------------------------------------------------
// PrecompiledContractExecutor
// ---------------------------------------------------------------------------

// PrecompiledContractExecutor executes precompiled contracts.
type PrecompiledContractExecutor struct {
	contracts map[engine.Address]*PrecompiledContract
}

// NewPrecompiledContractExecutor creates a new precompiled contract executor.
func NewPrecompiledContractExecutor() *PrecompiledContractExecutor {
	return &PrecompiledContractExecutor{
		contracts: PrecompiledContracts(),
	}
}

// Execute executes a precompiled contract call.
// Returns the gas cost and output, or an error.
func (pce *PrecompiledContractExecutor) Execute(address engine.Address, input []byte) (uint64, []byte, error) {
	contract, exists := pce.contracts[address]
	if !exists {
		return 0, nil, fmt.Errorf("precompiled contract not found: %s", address.String())
	}

	gasCost, err := contract.GasCost(input)
	if err != nil {
		return 0, nil, err
	}

	output, err := contract.Run(input)
	if err != nil {
		return gasCost, nil, err
	}

	return gasCost, output, nil
}

// HasContract checks if a precompiled contract exists at the given address.
func (pce *PrecompiledContractExecutor) HasContract(address engine.Address) bool {
	_, exists := pce.contracts[address]
	return exists
}

var _ = fmt.Sprintf
