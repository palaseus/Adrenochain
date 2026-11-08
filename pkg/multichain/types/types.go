// Package types provides shared types for the multi-chain framework, used by
// both the chain registry and individual chain connectors (e.g. Solana) without
// creating circular dependencies.

package types

import (
	"math/big"
)

// ChainFamily groups chains with similar transaction and address formats.
type ChainFamily int

const (
	ChainFamilyEVM ChainFamily = iota
	ChainFamilyBitcoin
	ChainFamilySolana
	ChainFamilyCosmos
	ChainFamilySubstrate
	ChainFamilyAdrenochain
)

// ChainType is a specific blockchain network.
type ChainType int

const (
	ChainTypeUnknown ChainType = iota
	// EVM
	ChainTypeEthereumMainnet
	ChainTypeEthereumSepolia
	ChainTypeEthereumHolesky
	ChainTypeArbitrumOne
	ChainTypeArbitrumSepolia
	ChainTypeOptimismMainnet
	ChainTypeOptimismSepolia
	ChainTypeBaseMainnet
	ChainTypeBaseSepolia
	ChainTypePolygonPOS
	ChainTypePolygonAmoy
	ChainTypeBSCMainnet
	ChainTypeBSCTestnet
	ChainTypeAvalancheC
	ChainTypeAvalancheFuji
	// Bitcoin
	ChainTypeBitcoinMainnet
	ChainTypeBitcoinTestnet
	ChainTypeBitcoinSignet
	ChainTypeBitcoinRegtest
	// Solana
	ChainTypeSolanaMainnetBeta
	ChainTypeSolanaTestnet
	ChainTypeSolanaDevnet
	// Cosmos
	ChainTypeCosmosHubMainnet
	ChainTypeOsmosisMainnet
	ChainTypeInjectiveMainnet
	// Substrate
	ChainTypePolkadot
	ChainTypeKusama
	// Adrenochain
	ChainTypeAdrenochainMainnet
	ChainTypeAdrenochainTestnet
)

// String returns the human-readable name of a chain type.
func (c ChainType) String() string {
	names := map[ChainType]string{
		ChainTypeEthereumMainnet:    "Ethereum Mainnet",
		ChainTypeEthereumSepolia:    "Ethereum Sepolia",
		ChainTypeEthereumHolesky:    "Ethereum Holesky",
		ChainTypeArbitrumOne:        "Arbitrum One",
		ChainTypeArbitrumSepolia:    "Arbitrum Sepolia",
		ChainTypeOptimismMainnet:    "Optimism Mainnet",
		ChainTypeOptimismSepolia:    "Optimism Sepolia",
		ChainTypeBaseMainnet:        "Base Mainnet",
		ChainTypeBaseSepolia:        "Base Sepolia",
		ChainTypePolygonPOS:         "Polygon PoS",
		ChainTypePolygonAmoy:        "Polygon Amoy",
		ChainTypeBSCMainnet:         "BSC Mainnet",
		ChainTypeBSCTestnet:         "BSC Testnet",
		ChainTypeAvalancheC:         "Avalanche C-Chain",
		ChainTypeAvalancheFuji:      "Avalanche Fuji",
		ChainTypeBitcoinMainnet:     "Bitcoin Mainnet",
		ChainTypeBitcoinTestnet:     "Bitcoin Testnet",
		ChainTypeBitcoinSignet:      "Bitcoin Signet",
		ChainTypeBitcoinRegtest:     "Bitcoin Regtest",
		ChainTypeSolanaMainnetBeta:  "Solana Mainnet Beta",
		ChainTypeSolanaTestnet:      "Solana Testnet",
		ChainTypeSolanaDevnet:       "Solana Devnet",
		ChainTypeCosmosHubMainnet:   "Cosmos Hub Mainnet",
		ChainTypeOsmosisMainnet:     "Osmosis Mainnet",
		ChainTypeInjectiveMainnet:   "Injective Mainnet",
		ChainTypePolkadot:           "Polkadot",
		ChainTypeKusama:             "Kusama",
		ChainTypeAdrenochainMainnet: "Adrenochain Mainnet",
		ChainTypeAdrenochainTestnet: "Adrenochain Testnet",
	}
	if n, ok := names[c]; ok {
		return n
	}
	return "Unknown"
}

// Family returns the chain family for a chain type.
func (c ChainType) Family() ChainFamily {
	switch c {
	case ChainTypeEthereumMainnet, ChainTypeEthereumSepolia, ChainTypeEthereumHolesky,
		ChainTypeArbitrumOne, ChainTypeArbitrumSepolia, ChainTypeOptimismMainnet,
		ChainTypeOptimismSepolia, ChainTypeBaseMainnet, ChainTypeBaseSepolia,
		ChainTypePolygonPOS, ChainTypePolygonAmoy, ChainTypeBSCMainnet,
		ChainTypeBSCTestnet, ChainTypeAvalancheC, ChainTypeAvalancheFuji:
		return ChainFamilyEVM
	case ChainTypeBitcoinMainnet, ChainTypeBitcoinTestnet, ChainTypeBitcoinSignet, ChainTypeBitcoinRegtest:
		return ChainFamilyBitcoin
	case ChainTypeSolanaMainnetBeta, ChainTypeSolanaTestnet, ChainTypeSolanaDevnet:
		return ChainFamilySolana
	case ChainTypeCosmosHubMainnet, ChainTypeOsmosisMainnet, ChainTypeInjectiveMainnet:
		return ChainFamilyCosmos
	case ChainTypePolkadot, ChainTypeKusama:
		return ChainFamilySubstrate
	case ChainTypeAdrenochainMainnet, ChainTypeAdrenochainTestnet:
		return ChainFamilyAdrenochain
	default:
		return ChainFamilyAdrenochain
	}
}

// ChainSpec describes a blockchain network for cross-chain operation.
type ChainSpec struct {
	Type                   ChainType   `json:"type"`
	ChainID                *big.Int    `json:"chain_id"`
	NativeAsset            NativeAsset `json:"native_asset"`
	Endpoints              []string    `json:"endpoints"`
	ExplorerURL            string      `json:"explorer_url"`
	Region                 string      `json:"region"`
	Jurisdictions          []string    `json:"jurisdictions"`
	ConfirmationBlocks     uint64      `json:"confirmation_blocks"`
	MaxMemoSize            int         `json:"max_memo_size"`
	SupportsSmartContracts bool        `json:"supports_smart_contracts"`
	IsTestnet              bool        `json:"is_testnet"`
}

// NativeAsset describes the native token of a chain.
type NativeAsset struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Address  string `json:"address"`
}

// Connector is the interface every chain connector must implement.
type Connector interface {
	ChainType() ChainType
	HealthCheck(ctx ChainContext) error
	GetBlockHeight(ctx ChainContext) (uint64, error)
	SubmitTransaction(ctx ChainContext, tx []byte) (string, error)
	GetTransactionStatus(ctx ChainContext, hash string) (string, error)
}

// ChainContext is a minimal context-like interface for chain operations.
type ChainContext interface {
	Done() <-chan struct{}
}
