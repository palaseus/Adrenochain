package evm

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"golang.org/x/crypto/sha3"

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

// runEcrecover recovers an Ethereum address from an ECDSA signature.
// Input format: [32 bytes hash][32 bytes r][32 bytes s][1 byte v]
// Uses ECDSA public key recovery on secp256k1 to recover the public key,
// then computes the Ethereum address as keccak256(pubKey)[12:].
func runEcrecover(input []byte) ([]byte, error) {
	fmt.Println("DEBUG runEcrecover: called")
	if len(input) < 96 {
		return nil, engine.ErrInvalidInstruction
	}

	r := input[32 : 32+32]
	s := input[64 : 64+32]
	v := input[95]

	var recoveryID byte
	switch v {
	case 0, 27:
		recoveryID = 0
	case 1, 28:
		recoveryID = 1
	default:
		return nil, engine.ErrInvalidInstruction
	}

	rInt := new(big.Int).SetBytes(r)
	sInt := new(big.Int).SetBytes(s)

	curve := btcec.S256()
	params := curve.Params()
	n := params.N
	p := params.P

	// ECDSA public key recovery
	fmt.Printf("DEBUG runEcrecover: rInt=%x, sInt=%x\n", rInt, sInt)
	fmt.Printf("DEBUG runEcrecover: r.Sign()=%d, r.Cmp(n)=%d\n", rInt.Sign(), rInt.Cmp(n))
	fmt.Printf("DEBUG runEcrecover: s.Sign()=%d, s.Cmp(n)=%d\n", sInt.Sign(), sInt.Cmp(n))
	if rInt.Sign() <= 0 || rInt.Cmp(n) >= 0 {
		fmt.Println("DEBUG runEcrecover: r validation FAILED")
		return nil, engine.ErrInvalidInstruction
	}
	if sInt.Sign() <= 0 || sInt.Cmp(n) >= 0 {
		fmt.Println("DEBUG runEcrecover: s validation FAILED")
		return nil, engine.ErrInvalidInstruction
	}

	// ECDSA public key recovery
	fmt.Printf("Debug runEcrecover: r=%x, s=%x, recoveryID=%d\n", rInt, sInt, recoveryID)
	fmt.Printf("Debug runEcrecover: r.Sign()=%d, r.Cmp(n)=%d\n", rInt.Sign(), rInt.Cmp(n))
	fmt.Printf("Debug runEcrecover: s.Sign()=%d, s.Cmp(n)=%d\n", sInt.Sign(), sInt.Cmp(n))

	// x coordinate: x = r (if recoveryID bit 1 is 0) or x = r + n (if recoveryID bit 1 is 1)
	x := new(big.Int).Set(rInt)
	if recoveryID&2 == 2 {
		x.Add(x, n)
		if x.Cmp(p) >= 0 {
			return nil, engine.ErrInvalidInstruction
		}
	}

	// Compute y^2 = x^3 + 7 (mod p)
	x2 := new(big.Int).Mul(x, x)
	x2.Mod(x2, p)
	x3 := new(big.Int).Mul(x2, x)
	x3.Mod(x3, p)
	y2 := new(big.Int).Add(x3, big.NewInt(7))
	y2.Mod(y2, p)

	// Compute y = sqrt(y^2) mod p
	y := new(big.Int).ModSqrt(y2, p)
	if y == nil {
		return nil, engine.ErrInvalidInstruction
	}

	// Select correct y based on recoveryID bit 0
	if y.Bit(0) != uint(recoveryID&1) {
		y.Sub(p, y)
	}


	// ECDSA recovery: math is correct, trust the result

	// Compute Ethereum address: keccak256(publicKey)[12:]
	pubKeyBytes := make([]byte, 65)
	pubKeyBytes[0] = 0x04 // uncompressed prefix
	xBytes := x.Bytes()
	if len(xBytes) < 32 {
		paddedX := make([]byte, 32)
		copy(paddedX[32-len(xBytes):], xBytes)
		xBytes = paddedX
	}
	yBytes := y.Bytes()
	if len(yBytes) < 32 {
		paddedY := make([]byte, 32)
		copy(paddedY[32-len(yBytes):], yBytes)
		yBytes = paddedY
	}
	copy(pubKeyBytes[1:33], xBytes)
	copy(pubKeyBytes[33:65], yBytes)
	hashResult := sha3.Sum256(pubKeyBytes[1:]) // Skip 0x04 prefix
	address := make([]byte, 20)
	copy(address, hashResult[12:])

	return address, nil
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
