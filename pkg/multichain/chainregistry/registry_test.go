package chainregistry

import (
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/multichain/types"
	"github.com/stretchr/testify/assert"
)

func TestChainTypeStringAndFamily(t *testing.T) {
	assert.Equal(t, "Ethereum Mainnet", types.ChainTypeEthereumMainnet.String())
	assert.Equal(t, types.ChainFamilyEVM, types.ChainTypeEthereumMainnet.Family())
	assert.Equal(t, types.ChainFamilyBitcoin, types.ChainTypeBitcoinMainnet.Family())
	assert.Equal(t, types.ChainFamilySolana, types.ChainTypeSolanaMainnetBeta.Family())
	assert.Equal(t, types.ChainFamilyCosmos, types.ChainTypeCosmosHubMainnet.Family())
	assert.Equal(t, types.ChainFamilyAdrenochain, types.ChainTypeAdrenochainMainnet.Family())
	assert.Equal(t, "Unknown", types.ChainTypeUnknown.String())
}

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()
	assert.NotNil(t, r)

	eth := r.GetChain(types.ChainTypeEthereumMainnet)
	assert.NotNil(t, eth)
	assert.Equal(t, big.NewInt(1), eth.ChainID)
	assert.Equal(t, "ETH", eth.NativeAsset.Symbol)
	assert.Equal(t, "EU", eth.Region)
	assert.True(t, eth.SupportsSmartContracts)

	btc := r.GetChain(types.ChainTypeBitcoinMainnet)
	assert.NotNil(t, btc)
	assert.False(t, btc.SupportsSmartContracts)
	assert.Equal(t, uint64(6), btc.ConfirmationBlocks)

	// Testnet detection.
	assert.True(t, r.GetChain(types.ChainTypeEthereumSepolia).IsTestnet)
	assert.False(t, r.GetChain(types.ChainTypeEthereumMainnet).IsTestnet)

	// Region filtering.
	euChains := r.ChainsByRegion("EU")
	assert.Contains(t, euChains, r.GetChain(types.ChainTypeEthereumMainnet))
	assert.Contains(t, euChains, r.GetChain(types.ChainTypeAdrenochainMainnet))
	assert.NotContains(t, euChains, r.GetChain(types.ChainTypePolygonPOS))

	// List all.
	all := r.ListChains()
	assert.GreaterOrEqual(t, len(all), 15)
}

func TestRegisterConnector(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterChain(&types.ChainSpec{Type: types.ChainTypeEthereumMainnet, ChainID: big.NewInt(1)})
	assert.Nil(t, r.GetConnector(types.ChainTypeEthereumMainnet))

	err := r.RegisterConnector(types.ChainTypeEthereumMainnet, fakeConn{chainType: types.ChainTypeEthereumMainnet})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetConnector(types.ChainTypeEthereumMainnet))
	assert.Equal(t, types.ChainTypeEthereumMainnet, r.GetConnector(types.ChainTypeEthereumMainnet).ChainType())
}

// fakeConn is a test double implementing the Connector interface.
type fakeConn struct {
	chainType types.ChainType
}

func (f fakeConn) ChainType() types.ChainType                                      { return f.chainType }
func (f fakeConn) HealthCheck(types.ChainContext) error                            { return nil }
func (f fakeConn) GetBlockHeight(types.ChainContext) (uint64, error)               { return 0, nil }
func (f fakeConn) SubmitTransaction(types.ChainContext, []byte) (string, error)    { return "", nil }
func (f fakeConn) GetTransactionStatus(types.ChainContext, string) (string, error) { return "", nil }
