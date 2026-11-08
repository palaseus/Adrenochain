// Package eip7930 implements ERC-7930: Interoperable Address Format.
//
// ERC-7930 defines a standard format for chain-aware addresses:
//   eip7930:<chainId>:<address>
//
// where chainId is a CAIP-2 blockchain identifier and address is the
// chain-specific account address (hex-encoded for EVM chains, base58 for
// Bitcoin, etc.).
//
// This enables cross-chain applications to unambiguously identify accounts
// on any chain without relying on external context.
//
// Specification: https://eips.ethereum.org/EIPS/eip-7930
//
// See pkg/eip7930/README.md for usage examples.

package eip7930

import (
	"errors"
	"strings"
)

// Prefix is the EIP-7930 address scheme prefix.
const Prefix = "eip7930"

// Address is a parsed EIP-7930 interoperable address.
type Address struct {
	// ChainID is the CAIP-2 blockchain identifier (e.g. "eip155:1" for Ethereum
	// mainnet, "solana:5eykt8UsFv8PN8N8..." for Solana).
	ChainID string

	// Address is the chain-specific account identifier (hex-encoded for EVM,
	// base58 for Bitcoin, etc.).
	Address string
}

// String returns the canonical eip7930:<chainId>:<address> representation.
func (a Address) String() string {
	return Prefix + ":" + a.ChainID + ":" + a.Address
}

// Parse parses an eip7930:<chainId>:<address> string into an Address.
// The chainId is a CAIP-2 identifier (e.g. "eip155:1" for Ethereum mainnet,
// "solana:5eykt8UsFv8PN8N8" for Solana mainnet).
func Parse(s string) (Address, error) {
	if !strings.HasPrefix(s, Prefix+":") {
		return Address{}, errors.New("invalid eip7930 address: missing " + Prefix + " prefix")
	}

	rest := strings.TrimPrefix(s, Prefix+":")
	// Format: <namespace>:<reference>:<address>  (chainId = <namespace>:<reference>)
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) < 3 {
		return Address{}, errors.New("invalid eip7930 address: expected <chainId>:<address>")
	}
	chainID := parts[0] + ":" + parts[1]
	addr := parts[2]
	if chainID == "" || addr == "" {
		return Address{}, errors.New("invalid eip7930 address: chainId and address must not be empty")
	}

	return Address{ChainID: chainID, Address: addr}, nil
}

// MustParse is like Parse but panics on error.
func MustParse(s string) Address {
	a, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return a
}

// IsEVM reports whether the address is on an EVM chain (chainId starts with "eip155:").
func (a Address) IsEVM() bool {
	return strings.HasPrefix(a.ChainID, "eip155:")
}

// EVMChainID returns the numeric EVM chain ID if the address is on an EVM chain,
// or 0 otherwise.
func (a Address) EVMChainID() uint64 {
	if !strings.HasPrefix(a.ChainID, "eip155:") {
		return 0
	}
	var id uint64
	for _, c := range a.ChainID[len("eip155:"):] {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + uint64(c-'0')
	}
	return id
}

// Common chain IDs.
const (
	ChainIDEthereumMainnet = "eip155:1"
	ChainIDEthereumSepolia = "eip155:11155111"
	ChainIDEthereumHolesky = "eip155:17000"
	ChainIDArbitrumOne     = "eip155:42161"
	ChainIDOptimism        = "eip155:10"
	ChainIDBase            = "eip155:8453"
	ChainIDPolygon         = "eip155:137"
	ChainIDBSC             = "eip155:56"
	ChainIDAvalancheC      = "eip155:43114"
	ChainIDSolanaMainnet   = "solana:5eykt8UsFv8PN8N8"
)
