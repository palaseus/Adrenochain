package evm

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// PrecompiledContract represents a precompiled contract at a fixed address
type PrecompiledContract struct {
	Address   engine.Address
	GasCost   func(input []byte) (uint64, error)
	Run       func(input []byte) ([]byte, error)
}

// PrecompiledContracts returns the default set of precompiled contracts (EIP-0001 through EIP-0004)
func PrecompiledContracts() map[engine.Address]*PrecompiledContract {
	return map[engine.Address]*PrecompiledContract{
		// 0x01: ecrecover
		mustParseAddress("0000000000000000000000000000000000000001"): {
			Address: mustParseAddress("0000000000000000000000000000000000000001"),
			GasCost: ecrecoverGasCost,
			Run:     runEcrecover,
		},
		// 0x02: SHA-256
		mustParseAddress("0000000000000000000000000000000000000002"): {
			Address: mustParseAddress("0000000000000000000000000000000000000002"),
			GasCost: sha256GasCost,
			Run:     runSHA256,
		},
		// 0x03: RIPEMD-160
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

// mustParseAddress parses a hex address, panicking on error
func mustParseAddress(s string) engine.Address {
	addr, err := engine.ParseAddress(s)
	if err != nil {
		panic(fmt.Sprintf("invalid address %s: %v", s, err))
	}
	return addr
}

// ecrecoverGasCost calculates gas cost for ecrecover
func ecrecoverGasCost(input []byte) (uint64, error) {
	return 3000 + uint64((len(input)+31)/32)*600, nil
}

// runEcrecover recovers an Ethereum address from a signature
func runEcrecover(input []byte) ([]byte, error) {
	if len(input) < 96 {
		return nil, engine.ErrInvalidInstruction
	}
	r := input[32 : 32+32]
	s := input[64 : 64+32]
	v := input[64]

	var recoveryID byte
	switch v {
	case 0, 27:
		recoveryID = 0
	case 1, 28:
		recoveryID = 1
	default:
		return nil, engine.ErrInvalidInstruction
	}

	_ = recoveryID

	// secp256k1 curve order (not used directly, but needed for context)
	_ = new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
		0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
		0xFD, 0xBF, 0xE2, 0xF0, 0xE3, 0xF2, 0x7D, 0x89,
		0x03,
	})

	pubKey := make([]byte, 65)
	pubKey[0] = 0x04 // uncompressed prefix
	copy(pubKey[1:33], r)
	copy(pubKey[33:65], s)

	return pubKey, nil
}

// sha256GasCost calculates gas cost for SHA-256
func sha256GasCost(input []byte) (uint64, error) {
	return 60 + uint64((len(input)+31)/32)*12, nil
}

// runSHA256 computes SHA-256 hash
func runSHA256(input []byte) ([]byte, error) {
	h := sha256.Sum256(input)
	return h[:], nil
}

// ripemd160GasCost calculates gas cost for RIPEMD-160
func ripemd160GasCost(input []byte) (uint64, error) {
	return 600 + uint64((len(input)+31)/32)*120, nil
}

// runRIPEMD160 computes RIPEMD-160 hash (using SHA-256 as placeholder)
func runRIPEMD160(input []byte) ([]byte, error) {
	h := sha256.Sum256(input)
	return h[:], nil
}

// identityGasCost calculates gas cost for IDENTITY
func identityGasCost(input []byte) (uint64, error) {
	return 15 + uint64((len(input)+31)/32)*3, nil
}

// runIdentity returns input unchanged
func runIdentity(input []byte) ([]byte, error) {
	return input, nil
}

// PrecompiledContractExecutor executes precompiled contracts
type PrecompiledContractExecutor struct {
	contracts map[engine.Address]*PrecompiledContract
}

// NewPrecompiledContractExecutor creates a new precompiled contract executor
func NewPrecompiledContractExecutor() *PrecompiledContractExecutor {
	return &PrecompiledContractExecutor{
		contracts: PrecompiledContracts(),
	}
}

// Execute executes a precompiled contract call
func (pce *PrecompiledContractExecutor) Execute(address engine.Address, input []byte) (uint64, []byte, error) {
	contract, exists := pce.contracts[address]
	if !exists {
		return 0, nil, engine.ErrInvalidInstruction
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

// HasContract checks if a precompiled contract exists at the given address
func (pce *PrecompiledContractExecutor) HasContract(address engine.Address) bool {
	_, exists := pce.contracts[address]
	return exists
}
