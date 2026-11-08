package chainregistry

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainTypeStringAndFamily(t *testing.T) {
	assert.Equal(t, "Ethereum Mainnet", ChainTypeEthereumMainnet.String())
	assert.Equal(t, ChainFamilyEVM, ChainTypeEthereumMainnet.Family())
	assert.Equal(t, ChainFamilyBitcoin, ChainTypeBitcoinMainnet.Family())
	assert.Equal(t, ChainFamilySolana, ChainTypeSolanaMainnetBeta.Family())
	assert.Equal(t, ChainFamilyCosmos, ChainTypeCosmosHubMainnet.Family())
	assert.Equal(t, ChainFamilyAdrenochain, ChainTypeAdrenochainMainnet.Family())
	assert.Equal(t, "Unknown", ChainTypeUnknown.String())
}

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()
	assert.NotNil(t, r)

	eth := r.GetChain(ChainTypeEthereumMainnet)
	assert.NotNil(t, eth)
	assert.Equal(t, big.NewInt(1), eth.ChainID)
	assert.Equal(t, "ETH", eth.NativeAsset.Symbol)
	assert.Equal(t, "EU", eth.Region)
	assert.True(t, eth.SupportsSmartContracts)

	btc := r.GetChain(ChainTypeBitcoinMainnet)
	assert.NotNil(t, btc)
	assert.False(t, btc.SupportsSmartContracts)
	assert.Equal(t, uint64(6), btc.ConfirmationBlocks)

	// Testnet detection.
	assert.True(t, r.GetChain(ChainTypeEthereumSepolia).IsTestnet)
	assert.False(t, r.GetChain(ChainTypeEthereumMainnet).IsTestnet)

	// Region filtering.
	euChains := r.ChainsByRegion("EU")
	assert.Contains(t, euChains, r.GetChain(ChainTypeEthereumMainnet))
	assert.Contains(t, euChains, r.GetChain(ChainTypeAdrenochainMainnet))
	assert.NotContains(t, euChains, r.GetChain(ChainTypePolygonPOS))

	// List all.
	all := r.ListChains()
	assert.GreaterOrEqual(t, len(all), 15)
}

func TestRegisterConnector(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterChain(&ChainSpec{Type: ChainTypeEthereumMainnet, ChainID: big.NewInt(1)})
	assert.Nil(t, r.GetConnector(ChainTypeEthereumMainnet))

	err := r.RegisterConnector(ChainTypeEthereumMainnet, fakeConn{chainType: ChainTypeEthereumMainnet})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetConnector(ChainTypeEthereumMainnet))
	assert.Equal(t, ChainTypeEthereumMainnet, r.GetConnector(ChainTypeEthereumMainnet).ChainType())
}

// fakeConn is a test double implementing the Connector interface.
type fakeConn struct {
	chainType ChainType
}

func (f fakeConn) ChainType() ChainType                                      { return f.chainType }
func (f fakeConn) HealthCheck(ChainContext) error                            { return nil }
func (f fakeConn) GetBlockHeight(ChainContext) (uint64, error)               { return 0, nil }
func (f fakeConn) SubmitTransaction(ChainContext, []byte) (string, error)    { return "", nil }
func (f fakeConn) GetTransactionStatus(ChainContext, string) (string, error) { return "", nil }
