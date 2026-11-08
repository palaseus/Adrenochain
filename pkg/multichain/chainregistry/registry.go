// Package chainregistry provides a multi-chain registry and connector framework
// for operating Adrenochain across all major blockchain networks and regions.
//
// Supported chain families:
//   - EVM-compatible (Ethereum, Arbitrum, Optimism, Base, Polygon, BSC, Avalanche C-Chain)
//   - Bitcoin (mainnet, testnet, signet, regtest)
//   - Solana (mainnet-beta, testnet, devnet)
//   - Cosmos SDK / IBC (Cosmos Hub, Osmosis, Injective, etc.)
//   - Polkadot / Substrate (Polkadot, Kusama, parachains)
//   - Adrenochain native
//
// Each chain is described by ChainSpec, which encodes network parameters,
// native asset metadata, and regional/jurisdictional constraints.
//
// See pkg/multichain/README.md for architecture and region mapping.

package chainregistry

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/palaseus/adrenochain/pkg/resilience/circuitbreaker"
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
	// ChainType uniquely identifies the chain.
	Type ChainType `json:"type"`
	// ChainID is the network chain ID (EVM: network id; Bitcoin: network magic; etc.).
	ChainID *big.Int `json:"chain_id"`
	// Native asset metadata.
	NativeAsset NativeAsset `json:"native_asset"`
	// RPC / API endpoints for this chain.
	Endpoints []string `json:"endpoints"`
	// Explorer base URL for transaction links.
	ExplorerURL string `json:"explorer_url"`
	// Region is the primary regulatory region (see compliance package).
	Region string `json:"region"`
	// Jurisdictions where this chain's validators/operators are subject to law.
	Jurisdictions []string `json:"jurisdictions"`
	// ConfirmationBlocks is the number of blocks to wait for finality.
	ConfirmationBlocks uint64 `json:"confirmation_blocks"`
	// MaxMemoSize is the max size of a cross-chain memo in bytes.
	MaxMemoSize int `json:"max_memo_size"`
	// SupportsSmartContracts indicates whether the chain supports EVM/WASM contracts.
	SupportsSmartContracts bool `json:"supports_smart_contracts"`
	// IsTestnet marks non-production networks.
	IsTestnet bool `json:"is_testnet"`
}

// NativeAsset describes the native token of a chain.
type NativeAsset struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
	Address  string `json:"address"` // empty for native, set for wrapped/ERC-20
}

// ChainContext is a minimal context-like interface for chain operations.
type ChainContext interface {
	Done() <-chan struct{}
}

// Connector is the interface every chain connector must implement.
type Connector interface {
	// ChainType returns the chain type this connector serves.
	ChainType() ChainType
	// HealthCheck verifies connectivity to the chain.
	HealthCheck(ctx ChainContext) error
	// GetBlockHeight returns the latest block height.
	GetBlockHeight(ctx ChainContext) (uint64, error)
	// SubmitTransaction broadcasts a signed transaction.
	SubmitTransaction(ctx ChainContext, tx []byte) (string, error)
	// GetTransactionStatus checks the status of a submitted transaction.
	GetTransactionStatus(ctx ChainContext, hash string) (string, error)
}

// ProtectedConnector wraps a Connector with a circuit breaker so that calls to
// a failing chain fail fast instead of cascading.
type ProtectedConnector struct {
	inner   Connector
	breaker *circuitbreaker.Breaker
}

// NewProtectedConnector wraps conn with a circuit breaker using cfg.
func NewProtectedConnector(conn Connector, cfg circuitbreaker.Config) *ProtectedConnector {
	return &ProtectedConnector{
		inner:   conn,
		breaker: circuitbreaker.New(cfg),
	}
}

func (pc *ProtectedConnector) ChainType() ChainType { return pc.inner.ChainType() }

func (pc *ProtectedConnector) HealthCheck(ctx ChainContext) error {
	return pc.breaker.Execute(func() error { return pc.inner.HealthCheck(ctx) })
}

func (pc *ProtectedConnector) GetBlockHeight(ctx ChainContext) (uint64, error) {
	var h uint64
	err := pc.breaker.Execute(func() error {
		var e error
		h, e = pc.inner.GetBlockHeight(ctx)
		return e
	})
	return h, err
}

func (pc *ProtectedConnector) SubmitTransaction(ctx ChainContext, tx []byte) (string, error) {
	var hash string
	err := pc.breaker.Execute(func() error {
		var e error
		hash, e = pc.inner.SubmitTransaction(ctx, tx)
		return e
	})
	return hash, err
}

func (pc *ProtectedConnector) GetTransactionStatus(ctx ChainContext, hash string) (string, error) {
	var status string
	err := pc.breaker.Execute(func() error {
		var e error
		status, e = pc.inner.GetTransactionStatus(ctx, hash)
		return e
	})
	return status, err
}

// BreakerState returns the current circuit breaker state.
func (pc *ProtectedConnector) BreakerState() circuitbreaker.State {
	return pc.breaker.State()
}

// BreakerStats returns a snapshot of circuit breaker statistics.
func (pc *ProtectedConnector) BreakerStats() circuitbreaker.Stats {
	return pc.breaker.Stats()
}

// Registry holds all registered chain specs and connectors.
type Registry struct {
	mu     sync.RWMutex
	chains map[ChainType]*ChainSpec
	conns  map[ChainType]Connector
}

// NewRegistry creates an empty multi-chain registry.
func NewRegistry() *Registry {
	return &Registry{
		chains: make(map[ChainType]*ChainSpec),
		conns:  make(map[ChainType]Connector),
	}
}

// RegisterChain adds a chain spec to the registry.
func (r *Registry) RegisterChain(spec *ChainSpec) error {
	if spec == nil {
		return fmt.Errorf("nil chain spec")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chains[spec.Type] = spec
	return nil
}

// RegisterConnector attaches a live connector for a chain type.
func (r *Registry) RegisterConnector(chainType ChainType, conn Connector) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[chainType] = conn
	return nil
}

// GetChain returns the chain spec for a chain type, or nil if not registered.
func (r *Registry) GetChain(chainType ChainType) *ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chains[chainType]
}

// GetConnector returns the connector for a chain type, or nil if not registered.
func (r *Registry) GetConnector(chainType ChainType) Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conns[chainType]
}

// ListChains returns all registered chain specs.
func (r *Registry) ListChains() []*ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*ChainSpec, 0, len(r.chains))
	for _, spec := range r.chains {
		out = append(out, spec)
	}
	return out
}

// ChainsByRegion returns chain specs filtered by region.
func (r *Registry) ChainsByRegion(region string) []*ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*ChainSpec
	for _, spec := range r.chains {
		if spec.Region == region {
			out = append(out, spec)
		}
	}
	return out
}

// DefaultRegistry returns a registry pre-populated with all major chains.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	chains := []*ChainSpec{
		// EVM
		{Type: ChainTypeEthereumMainnet, ChainID: big.NewInt(1), NativeAsset: NativeAsset{"ETH", "Ethereum", 18, ""}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 12, SupportsSmartContracts: true},
		{Type: ChainTypeEthereumSepolia, ChainID: big.NewInt(11155111), NativeAsset: NativeAsset{"ETH", "Ethereum", 18, ""}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 12, SupportsSmartContracts: true, IsTestnet: true},
		{Type: ChainTypeArbitrumOne, ChainID: big.NewInt(42161), NativeAsset: NativeAsset{"ETH", "Ethereum", 18, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: ChainTypeOptimismMainnet, ChainID: big.NewInt(10), NativeAsset: NativeAsset{"ETH", "Ethereum", 18, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: ChainTypeBaseMainnet, ChainID: big.NewInt(8453), NativeAsset: NativeAsset{"ETH", "Ethereum", 18, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: ChainTypePolygonPOS, ChainID: big.NewInt(137), NativeAsset: NativeAsset{"MATIC", "Polygon", 18, ""}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 128, SupportsSmartContracts: true},
		{Type: ChainTypeBSCMainnet, ChainID: big.NewInt(56), NativeAsset: NativeAsset{"BNB", "BNB", 18, ""}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 15, SupportsSmartContracts: true},
		{Type: ChainTypeAvalancheC, ChainID: big.NewInt(43114), NativeAsset: NativeAsset{"AVAX", "Avalanche", 18, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		// Bitcoin
		{Type: ChainTypeBitcoinMainnet, ChainID: big.NewInt(0), NativeAsset: NativeAsset{"BTC", "Bitcoin", 8, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 6, SupportsSmartContracts: false},
		{Type: ChainTypeBitcoinTestnet, ChainID: big.NewInt(0), NativeAsset: NativeAsset{"tBTC", "Bitcoin Testnet", 8, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 6, SupportsSmartContracts: false, IsTestnet: true},
		// Solana
		{Type: ChainTypeSolanaMainnetBeta, ChainID: big.NewInt(0), NativeAsset: NativeAsset{"SOL", "Solana", 9, ""}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 32, SupportsSmartContracts: true},
		// Cosmos
		{Type: ChainTypeCosmosHubMainnet, ChainID: big.NewInt(0), NativeAsset: NativeAsset{"ATOM", "Cosmos", 6, ""}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: ChainTypeOsmosisMainnet, ChainID: big.NewInt(0), NativeAsset: NativeAsset{"OSMO", "Osmosis", 6, ""}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		// Adrenochain
		{Type: ChainTypeAdrenochainMainnet, ChainID: big.NewInt(1), NativeAsset: NativeAsset{"ADN", "Adrenochain", 18, ""}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 10, SupportsSmartContracts: true},
		{Type: ChainTypeAdrenochainTestnet, ChainID: big.NewInt(2), NativeAsset: NativeAsset{"tADN", "Adrenochain Testnet", 18, ""}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 10, SupportsSmartContracts: true, IsTestnet: true},
	}
	for _, c := range chains {
		_ = r.RegisterChain(c)
	}
	return r
}
