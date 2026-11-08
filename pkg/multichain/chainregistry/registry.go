package chainregistry

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/palaseus/adrenochain/pkg/multichain/solana"
	"github.com/palaseus/adrenochain/pkg/multichain/types"
	"github.com/palaseus/adrenochain/pkg/resilience/circuitbreaker"
)

// Registry holds all registered chain specs and connectors.
type Registry struct {
	mu     sync.RWMutex
	chains map[types.ChainType]*types.ChainSpec
	conns  map[types.ChainType]types.Connector
}

// NewRegistry creates an empty multi-chain registry.
func NewRegistry() *Registry {
	return &Registry{
		chains: make(map[types.ChainType]*types.ChainSpec),
		conns:  make(map[types.ChainType]types.Connector),
	}
}

// RegisterChain adds a chain spec to the registry.
func (r *Registry) RegisterChain(spec *types.ChainSpec) error {
	if spec == nil {
		return fmt.Errorf("nil chain spec")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chains[spec.Type] = spec
	return nil
}

// RegisterConnector attaches a live connector for a chain type.
func (r *Registry) RegisterConnector(chainType types.ChainType, conn types.Connector) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[chainType] = conn
	return nil
}

// GetChain returns the chain spec for a chain type, or nil if not registered.
func (r *Registry) GetChain(chainType types.ChainType) *types.ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chains[chainType]
}

// GetConnector returns the connector for a chain type, or nil if not registered.
func (r *Registry) GetConnector(chainType types.ChainType) types.Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conns[chainType]
}

// ListChains returns all registered chain specs.
func (r *Registry) ListChains() []*types.ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*types.ChainSpec, 0, len(r.chains))
	for _, spec := range r.chains {
		out = append(out, spec)
	}
	return out
}

// ChainsByRegion returns chain specs filtered by region.
func (r *Registry) ChainsByRegion(region string) []*types.ChainSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*types.ChainSpec
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
	chains := []*types.ChainSpec{
		// EVM
		{Type: types.ChainTypeEthereumMainnet, ChainID: big.NewInt(1), NativeAsset: types.NativeAsset{Symbol: "ETH", Name: "Ethereum", Decimals: 18}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 12, SupportsSmartContracts: true},
		{Type: types.ChainTypeEthereumSepolia, ChainID: big.NewInt(11155111), NativeAsset: types.NativeAsset{Symbol: "ETH", Name: "Ethereum", Decimals: 18}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 12, SupportsSmartContracts: true, IsTestnet: true},
		{Type: types.ChainTypeArbitrumOne, ChainID: big.NewInt(42161), NativeAsset: types.NativeAsset{Symbol: "ETH", Name: "Ethereum", Decimals: 18}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: types.ChainTypeOptimismMainnet, ChainID: big.NewInt(10), NativeAsset: types.NativeAsset{Symbol: "ETH", Name: "Ethereum", Decimals: 18}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: types.ChainTypeBaseMainnet, ChainID: big.NewInt(8453), NativeAsset: types.NativeAsset{Symbol: "ETH", Name: "Ethereum", Decimals: 18}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: types.ChainTypePolygonPOS, ChainID: big.NewInt(137), NativeAsset: types.NativeAsset{Symbol: "MATIC", Name: "Polygon", Decimals: 18}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 128, SupportsSmartContracts: true},
		{Type: types.ChainTypeBSCMainnet, ChainID: big.NewInt(56), NativeAsset: types.NativeAsset{Symbol: "BNB", Name: "BNB", Decimals: 18}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 15, SupportsSmartContracts: true},
		{Type: types.ChainTypeAvalancheC, ChainID: big.NewInt(43114), NativeAsset: types.NativeAsset{Symbol: "AVAX", Name: "Avalanche", Decimals: 18}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		// Bitcoin
		{Type: types.ChainTypeBitcoinMainnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "BTC", Name: "Bitcoin", Decimals: 8}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 6, SupportsSmartContracts: false},
		{Type: types.ChainTypeBitcoinTestnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "tBTC", Name: "Bitcoin Testnet", Decimals: 8}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 6, SupportsSmartContracts: false, IsTestnet: true},
		// Solana
		{Type: types.ChainTypeSolanaMainnetBeta, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "SOL", Name: "Solana", Decimals: 9}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 32, SupportsSmartContracts: true},
		{Type: types.ChainTypeSolanaTestnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "SOL", Name: "Solana Testnet", Decimals: 9}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 32, SupportsSmartContracts: true, IsTestnet: true},
		{Type: types.ChainTypeSolanaDevnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "SOL", Name: "Solana Devnet", Decimals: 9}, Region: "US", Jurisdictions: []string{"US"}, ConfirmationBlocks: 32, SupportsSmartContracts: true, IsTestnet: true},
		// Cosmos
		{Type: types.ChainTypeCosmosHubMainnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "ATOM", Name: "Cosmos", Decimals: 6}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		{Type: types.ChainTypeOsmosisMainnet, ChainID: big.NewInt(0), NativeAsset: types.NativeAsset{Symbol: "OSMO", Name: "Osmosis", Decimals: 6}, Region: "SG", Jurisdictions: []string{"SG"}, ConfirmationBlocks: 1, SupportsSmartContracts: true},
		// Adrenochain
		{Type: types.ChainTypeAdrenochainMainnet, ChainID: big.NewInt(1), NativeAsset: types.NativeAsset{Symbol: "ADN", Name: "Adrenochain", Decimals: 18}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 10, SupportsSmartContracts: true},
		{Type: types.ChainTypeAdrenochainTestnet, ChainID: big.NewInt(2), NativeAsset: types.NativeAsset{Symbol: "tADN", Name: "Adrenochain Testnet", Decimals: 18}, Region: "EU", Jurisdictions: []string{"EU"}, ConfirmationBlocks: 10, SupportsSmartContracts: true, IsTestnet: true},
	}
	for _, c := range chains {
		_ = r.RegisterChain(c)
	}

	// Register Solana connectors (placeholder RPC URLs; connect() must be called
	// with a reachable endpoint before use). Wrap each with a circuit breaker.
	solanaChains := []types.ChainType{
		types.ChainTypeSolanaMainnetBeta,
		types.ChainTypeSolanaTestnet,
		types.ChainTypeSolanaDevnet,
	}
	for _, ct := range solanaChains {
		var rpcURL string
		switch ct {
		case types.ChainTypeSolanaMainnetBeta:
			rpcURL = "https://api.mainnet-beta.solana.com"
		case types.ChainTypeSolanaTestnet:
			rpcURL = "https://api.testnet.solana.com"
		case types.ChainTypeSolanaDevnet:
			rpcURL = "https://api.devnet.solana.com"
		}
		conn := solana.NewConnector(ct, rpcURL)
		protected := NewProtectedConnector(conn, circuitbreaker.DefaultConfig())
		_ = r.RegisterConnector(ct, protected)
	}

	return r
}

// ProtectedConnector wraps a types.Connector with a circuit breaker so that
// calls to a failing chain fail fast instead of cascading.
type ProtectedConnector struct {
	inner   types.Connector
	breaker *circuitbreaker.Breaker
}

// NewProtectedConnector wraps conn with a circuit breaker using cfg.
func NewProtectedConnector(conn types.Connector, cfg circuitbreaker.Config) *ProtectedConnector {
	return &ProtectedConnector{
		inner:   conn,
		breaker: circuitbreaker.New(cfg),
	}
}

func (pc *ProtectedConnector) ChainType() types.ChainType { return pc.inner.ChainType() }

func (pc *ProtectedConnector) HealthCheck(ctx types.ChainContext) error {
	return pc.breaker.Execute(func() error { return pc.inner.HealthCheck(ctx) })
}

func (pc *ProtectedConnector) GetBlockHeight(ctx types.ChainContext) (uint64, error) {
	var h uint64
	err := pc.breaker.Execute(func() error {
		var e error
		h, e = pc.inner.GetBlockHeight(ctx)
		return e
	})
	return h, err
}

func (pc *ProtectedConnector) SubmitTransaction(ctx types.ChainContext, tx []byte) (string, error) {
	var hash string
	err := pc.breaker.Execute(func() error {
		var e error
		hash, e = pc.inner.SubmitTransaction(ctx, tx)
		return e
	})
	return hash, err
}

func (pc *ProtectedConnector) GetTransactionStatus(ctx types.ChainContext, hash string) (string, error) {
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
