package evm

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// PrecompiledContract represents a precompiled contract at a fixed address
type PrecompiledContract struct {
	Address engine.Address
	GasCost func(input []byte) (uint64, error)
	Run     func(input []byte) ([]byte, error)
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

// runEcrecover recovers an Ethereum address from an ECDSA signature.
// Input format: [32 bytes hash][32 bytes r][32 bytes s][1 byte v]
// Uses ECDSA public key recovery on secp256k1 to recover the public key,
// then computes the Ethereum address as keccak256(pubKey)[12:].
func runEcrecover(input []byte) ([]byte, error) {
	// Preserve this engine's existing 97-byte hash/r/s/v interface.
	if len(input) != 97 {
		return nil, engine.ErrInvalidInstruction
	}
	v := input[96]
	if v == 0 || v == 1 {
		v += 27
	}
	if v != 27 && v != 28 {
		return nil, engine.ErrInvalidInstruction
	}
	compact := make([]byte, 65)
	compact[0] = v
	copy(compact[1:33], input[32:64])
	copy(compact[33:], input[64:96])
	pub, _, err := ecdsa.RecoverCompact(compact, input[:32])
	if err != nil {
		return nil, engine.ErrInvalidInstruction
	}
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pub.SerializeUncompressed()[1:])
	digest := hasher.Sum(nil)
	return append([]byte(nil), digest[12:]...), nil
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

// runRIPEMD160 returns the digest padded to a 32-byte EVM word.
func runRIPEMD160(input []byte) ([]byte, error) {
	hasher := ripemd160.New()
	hasher.Write(input)
	output := make([]byte, 32)
	copy(output[12:], hasher.Sum(nil))
	return output, nil
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
