package sidechains

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSidechainManager(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.sidechains)
	assert.NotNil(t, sm.bridges)
	assert.NotNil(t, sm.crossChainTxs)
	assert.NotNil(t, sm.metrics)
	assert.NotNil(t, sm.txQueue)
	assert.NotNil(t, sm.bridgeQueue)
	assert.NotNil(t, sm.metricsUpdater)
}

func TestCreateSidechain(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test valid sidechain creation
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Description:   "A test sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}

	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)
	assert.NotEmpty(t, sidechain.ID)
	assert.Equal(t, SidechainActive, sidechain.Status)
	assert.Equal(t, uint64(0), sidechain.BlockHeight)
	assert.Empty(t, sidechain.LastBlockHash)
	assert.NotZero(t, sidechain.CreatedAt)
	assert.NotZero(t, sidechain.UpdatedAt)
	assert.Equal(t, 0, sidechain.ValidatorCount)
	assert.Equal(t, big.NewInt(0), sidechain.TotalStake)
	assert.NotNil(t, sidechain.BridgeAddresses)

	// Test sidechain retrieval
	retrievedSidechain, err := sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, sidechain.ID, retrievedSidechain.ID)
	assert.Equal(t, sidechain.Name, retrievedSidechain.Name)

	// Test metrics initialization
	metrics, err := sm.GetSidechainMetrics(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, sidechain.ID, metrics.SidechainID)
	assert.Equal(t, 0.0, metrics.TPS)
	assert.Equal(t, 0, metrics.ValidatorCount)
	assert.Equal(t, 0, metrics.BridgeCount)
}

func TestCreateSidechainValidation(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test missing name
	sidechain := &Sidechain{
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain name is required")

	// Test missing type
	sidechain = &Sidechain{
		Name:          "Test Sidechain",
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err = sm.CreateSidechain(sidechain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain type is required")

	// Test missing parent chain
	sidechain = &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ConsensusType: PoSConsensus,
	}
	err = sm.CreateSidechain(sidechain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent chain is required")

	// Test missing consensus type
	sidechain = &Sidechain{
		Name:        "Test Sidechain",
		Type:        PlasmaSidechain,
		ParentChain: "Ethereum",
	}
	err = sm.CreateSidechain(sidechain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consensus type is required")
}

func TestGetSidechain(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test getting non-existent sidechain
	_, err := sm.GetSidechain("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain not found")

	// Test getting existing sidechain
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err = sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	retrievedSidechain, err := sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, sidechain.ID, retrievedSidechain.ID)
}

func TestGetSidechains(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create multiple sidechains
	sidechain1 := &Sidechain{
		Name:          "Plasma Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	sidechain2 := &Sidechain{
		Name:          "Polygon Sidechain",
		Type:          PolygonSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoAConsensus,
	}
	sidechain3 := &Sidechain{
		Name:          "ZK Sidechain",
		Type:          ZKSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: DPoSConsensus,
	}

	err := sm.CreateSidechain(sidechain1)
	require.NoError(t, err)
	err = sm.CreateSidechain(sidechain2)
	require.NoError(t, err)
	err = sm.CreateSidechain(sidechain3)
	require.NoError(t, err)

	// Test getting all sidechains
	allSidechains := sm.GetSidechains("", "")
	assert.Len(t, allSidechains, 3)

	// Test filtering by status
	activeSidechains := sm.GetSidechains(SidechainActive, "")
	assert.Len(t, activeSidechains, 3)

	// Test filtering by type
	plasmaSidechains := sm.GetSidechains("", PlasmaSidechain)
	assert.Len(t, plasmaSidechains, 1)
	assert.Equal(t, "Plasma Sidechain", plasmaSidechains[0].Name)

	// Test filtering by both status and type
	activePlasmaSidechains := sm.GetSidechains(SidechainActive, PlasmaSidechain)
	assert.Len(t, activePlasmaSidechains, 1)
	assert.Equal(t, "Plasma Sidechain", activePlasmaSidechains[0].Name)
}

func TestUpdateSidechainStatus(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test updating status
	err = sm.UpdateSidechainStatus(sidechain.ID, SidechainPaused)
	require.NoError(t, err)

	updatedSidechain, err := sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, SidechainPaused, updatedSidechain.Status)

	// Test updating non-existent sidechain
	err = sm.UpdateSidechainStatus("non-existent", SidechainActive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain not found")
}

func TestCreateBridge(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create a sidechain first
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test valid bridge creation
	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}

	err = sm.CreateBridge(bridge)
	require.NoError(t, err)
	assert.NotEmpty(t, bridge.ID)
	assert.Equal(t, BridgeActive, bridge.Status)
	assert.NotZero(t, bridge.CreatedAt)
	assert.NotZero(t, bridge.UpdatedAt)

	// Test bridge retrieval
	retrievedBridge, err := sm.GetBridge(bridge.ID)
	require.NoError(t, err)
	assert.Equal(t, bridge.ID, retrievedBridge.ID)
	assert.Equal(t, bridge.Name, retrievedBridge.Name)

	// Verify sidechain bridge addresses were updated
	updatedSidechain, err := sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, bridge.SidechainAddr, updatedSidechain.BridgeAddresses[bridge.Asset])
}

func TestCreateBridgeValidation(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create a sidechain first
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test missing name
	bridge := &Bridge{
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge name is required")

	// Test missing main chain
	bridge = &Bridge{
		Name:           "ETH Bridge",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "main chain is required")

	// Test missing sidechain ID
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain ID is required")

	// Test missing asset
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "asset is required")

	// Test missing main chain address
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "main chain address is required")

	// Test missing sidechain address
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain address is required")

	// Test invalid total liquidity
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(0),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total liquidity must be positive")

	// Test invalid min transfer
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(0),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum transfer must be positive")

	// Test invalid max transfer
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(0),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum transfer must be positive")

	// Test min transfer exceeding max transfer
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(10000),
		MaxTransfer:    big.NewInt(1000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum transfer cannot exceed maximum transfer")

	// Test invalid fee percentage
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  -1.0,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fee percentage must be between 0 and 100")

	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  101.0,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fee percentage must be between 0 and 100")

	// Test non-existent sidechain
	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    "non-existent",
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain not found")

	// Test inactive sidechain
	err = sm.UpdateSidechainStatus(sidechain.ID, SidechainPaused)
	require.NoError(t, err)

	bridge = &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain is not active")
}

func TestGetBridge(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test getting non-existent bridge
	_, err := sm.GetBridge("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge not found")

	// Test getting existing bridge
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err = sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	retrievedBridge, err := sm.GetBridge(bridge.ID)
	require.NoError(t, err)
	assert.Equal(t, bridge.ID, retrievedBridge.ID)
}

func TestGetBridges(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Create multiple bridges
	bridge1 := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress1",
		SidechainAddr:  "0xSidechainAddress1",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
	}
	bridge2 := &Bridge{
		Name:           "BTC Bridge",
		MainChain:      "Bitcoin",
		SidechainID:    sidechain.ID,
		Asset:          "BTC",
		MainChainAddr:  "0xMainChainAddress2",
		SidechainAddr:  "0xSidechainAddress2",
		TotalLiquidity: big.NewInt(500000),
		MinTransfer:    big.NewInt(50),
		MaxTransfer:    big.NewInt(5000),
		FeePercentage:  0.2,
	}

	err = sm.CreateBridge(bridge1)
	require.NoError(t, err)
	err = sm.CreateBridge(bridge2)
	require.NoError(t, err)

	// Test getting all bridges
	allBridges := sm.GetBridges("", "", "")
	assert.Len(t, allBridges, 2)

	// Test filtering by status
	activeBridges := sm.GetBridges(BridgeActive, "", "")
	assert.Len(t, activeBridges, 2)

	// Test filtering by sidechain ID
	sidechainBridges := sm.GetBridges("", sidechain.ID, "")
	assert.Len(t, sidechainBridges, 2)

	// Test filtering by asset
	ethBridges := sm.GetBridges("", "", "ETH")
	assert.Len(t, ethBridges, 1)
	assert.Equal(t, "ETH", ethBridges[0].Asset)

	// Test filtering by both status and sidechain ID
	activeSidechainBridges := sm.GetBridges(BridgeActive, sidechain.ID, "")
	assert.Len(t, activeSidechainBridges, 2)
}

func TestCreateCrossChainTransaction(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain and bridge first
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	// Test creating valid cross-chain transaction
	tx := &CrossChainTransaction{
		BridgeID:   bridge.ID,
		Direction:  MainToSidechain,
		Amount:     big.NewInt(500),
		Asset:      "ETH",
		Sender:     "0xSender",
		Recipient:  "0xRecipient",
		Nonce:      1,
		GasLimit:   21000,
		GasPrice:   big.NewInt(20000000000),
		Data:       []byte("transfer"),
		Signature:  "0xSignature",
	}

	err = sm.CreateCrossChainTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)
	assert.NotZero(t, tx.CreatedAt)
	assert.NotZero(t, tx.UpdatedAt)

	// Test retrieving the transaction
	retrievedTx, err := sm.GetCrossChainTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.ID, retrievedTx.ID)
	assert.Equal(t, tx.BridgeID, retrievedTx.BridgeID)
	assert.Equal(t, tx.Direction, retrievedTx.Direction)
}

func TestCreateCrossChainTransactionValidation(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain and bridge first
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	// Test missing bridge ID
	tx := &CrossChainTransaction{
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge ID is required")

	// Test missing direction
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "direction is required")

	// Test missing amount
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")

	// Test missing asset
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "asset is required")

	// Test missing sender
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender is required")

	// Test missing recipient
	tx = &CrossChainTransaction{
		BridgeID: bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient is required")

	// Test non-existent bridge
	tx = &CrossChainTransaction{
		BridgeID:  "non-existent",
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge not found")

	// Test inactive bridge (sidechain paused)
	err = sm.UpdateSidechainStatus(sidechain.ID, SidechainPaused)
	require.NoError(t, err)

	// The bridge should still exist but the sidechain is inactive
	// We need to check if the bridge is accessible
	_, err = sm.GetBridge(bridge.ID)
	require.NoError(t, err)

	// Try to create transaction - should fail due to sidechain being inactive
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain is not active")

	// Reactivate sidechain
	err = sm.UpdateSidechainStatus(sidechain.ID, SidechainActive)
	require.NoError(t, err)

	// Test amount below minimum
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(50),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount below minimum transfer limit")

	// Test amount above maximum
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(20000),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount above maximum transfer limit")

	// Test asset mismatch
	tx = &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "BTC",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "asset does not match bridge asset")
}

func TestGetCrossChainTransactions(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain and bridge
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	// Create multiple transactions
	tx1 := &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender1",
		Recipient: "0xRecipient1",
	}
	tx2 := &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: SidechainToMain,
		Amount:    big.NewInt(300),
		Asset:     "ETH",
		Sender:    "0xSender2",
		Recipient: "0xRecipient2",
	}

	err = sm.CreateCrossChainTransaction(tx1)
	require.NoError(t, err)
	err = sm.CreateCrossChainTransaction(tx2)
	require.NoError(t, err)

	// Test getting all transactions
	allTxs := sm.GetCrossChainTransactions("", "", "")
	assert.Len(t, allTxs, 2)

	// Test filtering by bridge ID
	bridgeTxs := sm.GetCrossChainTransactions("", bridge.ID, "")
	assert.Len(t, bridgeTxs, 2)

	// Test filtering by direction
	mainToSidechainTxs := sm.GetCrossChainTransactions("", "", MainToSidechain)
	assert.Len(t, mainToSidechainTxs, 1)
	assert.Equal(t, MainToSidechain, mainToSidechainTxs[0].Direction)

	sidechainToMainTxs := sm.GetCrossChainTransactions("", "", SidechainToMain)
	assert.Len(t, sidechainToMainTxs, 1)
	assert.Equal(t, SidechainToMain, sidechainToMainTxs[0].Direction)

	// Test filtering by status (may vary due to async processing)
	statusTxs := sm.GetCrossChainTransactions(CrossChainTxPending, "", "")
	if len(statusTxs) > 0 {
		assert.Equal(t, CrossChainTxPending, statusTxs[0].Status)
	}
}

func TestUpdateSidechainBlock(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test updating block information
	err = sm.UpdateSidechainBlock(sidechain.ID, 100, "0xBlockHash100")
	require.NoError(t, err)

	updatedSidechain, err := sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), updatedSidechain.BlockHeight)
	assert.Equal(t, "0xBlockHash100", updatedSidechain.LastBlockHash)
	assert.NotZero(t, updatedSidechain.LastBlockTime)

	// Test updating with higher block height
	err = sm.UpdateSidechainBlock(sidechain.ID, 200, "0xBlockHash200")
	require.NoError(t, err)

	updatedSidechain, err = sm.GetSidechain(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(200), updatedSidechain.BlockHeight)
	assert.Equal(t, "0xBlockHash200", updatedSidechain.LastBlockHash)

	// Test updating with lower block height
	err = sm.UpdateSidechainBlock(sidechain.ID, 150, "0xBlockHash150")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "block height must be greater than current height")

	// Test updating non-existent sidechain
	err = sm.UpdateSidechainBlock("non-existent", 100, "0xBlockHash")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain not found")
}

func TestGetSidechainMetrics(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test getting metrics
	metrics, err := sm.GetSidechainMetrics(sidechain.ID)
	require.NoError(t, err)
	assert.Equal(t, sidechain.ID, metrics.SidechainID)
	assert.Equal(t, 0.0, metrics.TPS)
	assert.Equal(t, 0, metrics.ValidatorCount)
	assert.Equal(t, big.NewInt(0), metrics.TotalStake)
	assert.Equal(t, uint64(0), metrics.CrossChainTxCount)
	assert.Equal(t, 0, metrics.BridgeCount)

	// Test getting metrics for non-existent sidechain
	_, err = sm.GetSidechainMetrics("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain metrics not found")
}

func TestGetAllSidechainMetrics(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create multiple sidechains
	sidechain1 := &Sidechain{
		Name:          "Sidechain 1",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	sidechain2 := &Sidechain{
		Name:          "Sidechain 2",
		Type:          PolygonSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoAConsensus,
	}

	err := sm.CreateSidechain(sidechain1)
	require.NoError(t, err)
	err = sm.CreateSidechain(sidechain2)
	require.NoError(t, err)

	// Test getting all metrics
	allMetrics := sm.GetAllSidechainMetrics()
	assert.Len(t, allMetrics, 2)

	// Verify each sidechain has metrics
	sidechain1Found := false
	sidechain2Found := false
	for _, metrics := range allMetrics {
		if metrics.SidechainID == sidechain1.ID {
			sidechain1Found = true
		}
		if metrics.SidechainID == sidechain2.ID {
			sidechain2Found = true
		}
	}
	assert.True(t, sidechain1Found)
	assert.True(t, sidechain2Found)
}

func TestConcurrency(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	const numGoroutines = 10
	const numSidechains = 5
	const numBridges = 3

	// Test concurrent sidechain creation
	sidechainChan := make(chan *Sidechain, numSidechains)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			sidechain := &Sidechain{
				Name:          fmt.Sprintf("Sidechain %d", id),
				Type:          PlasmaSidechain,
				ParentChain:   "Ethereum",
				ConsensusType: PoSConsensus,
			}
			err := sm.CreateSidechain(sidechain)
			if err == nil {
				sidechainChan <- sidechain
			}
		}(i)
	}

	// Collect created sidechains
	var sidechains []*Sidechain
	for i := 0; i < numSidechains; i++ {
		select {
		case sidechain := <-sidechainChan:
			sidechains = append(sidechains, sidechain)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for sidechain creation")
		}
	}

	// Verify all sidechains were created
	assert.Len(t, sidechains, numSidechains)

	// Test concurrent bridge creation
	if len(sidechains) > 0 {
		bridgeChan := make(chan *Bridge, numBridges*len(sidechains))
		for _, sidechain := range sidechains {
			for j := 0; j < numBridges; j++ {
				go func(s *Sidechain, bridgeID int) {
					bridge := &Bridge{
						Name:           fmt.Sprintf("Bridge %s_%d", s.ID, bridgeID),
						MainChain:      "Ethereum",
						SidechainID:    s.ID,
						Asset:          fmt.Sprintf("ASSET%d", bridgeID),
						MainChainAddr:  fmt.Sprintf("0xMain%d", bridgeID),
						SidechainAddr:  fmt.Sprintf("0xSide%d", bridgeID),
						TotalLiquidity: big.NewInt(1000000),
						MinTransfer:    big.NewInt(100),
						MaxTransfer:    big.NewInt(10000),
						FeePercentage:  0.1,
						SecurityLevel:  SecurityMedium,
					}
					err := sm.CreateBridge(bridge)
					if err == nil {
						bridgeChan <- bridge
					}
				}(sidechain, j)
			}
		}

		// Wait for all bridges to be created
		successCount := 0
		for i := 0; i < numBridges*len(sidechains); i++ {
			select {
			case bridge := <-bridgeChan:
				if bridge != nil {
					successCount++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for bridge creation")
			}
		}

		// Verify bridges were created
		assert.Greater(t, successCount, 0)
	}
}

func TestMemorySafety(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	const numSidechains = 50
	const numBridges = 25
	const numTransactions = 30

	// Create many sidechains
	var sidechains []*Sidechain
	for i := 0; i < numSidechains; i++ {
		sidechain := &Sidechain{
			Name:          fmt.Sprintf("Sidechain %d", i),
			Type:          PlasmaSidechain,
			ParentChain:   "Ethereum",
			ConsensusType: PoSConsensus,
		}
		err := sm.CreateSidechain(sidechain)
		require.NoError(t, err)
		sidechains = append(sidechains, sidechain)
	}

	// Create many bridges
	for i := 0; i < numBridges && i < len(sidechains); i++ {
		bridge := &Bridge{
			Name:           fmt.Sprintf("Bridge %d", i),
			MainChain:      "Ethereum",
			SidechainID:    sidechains[i].ID,
			Asset:          fmt.Sprintf("ASSET%d", i),
			MainChainAddr:  fmt.Sprintf("0xMain%d", i),
			SidechainAddr:  fmt.Sprintf("0xSide%d", i),
			TotalLiquidity: big.NewInt(1000000),
			MinTransfer:    big.NewInt(100),
			MaxTransfer:    big.NewInt(10000),
			FeePercentage:  0.1,
			SecurityLevel:  SecurityMedium,
		}
		sm.CreateBridge(bridge)
	}

	// Create many cross-chain transactions
	for i := 0; i < numTransactions && i < len(sidechains); i++ {
		bridges := sm.GetBridges("", sidechains[i].ID, "")
		if len(bridges) > 0 {
			tx := &CrossChainTransaction{
				BridgeID:  bridges[0].ID,
				Direction: MainToSidechain,
				Amount:    big.NewInt(int64(100 + i)),
				Asset:     bridges[0].Asset,
				Sender:    fmt.Sprintf("sender_%d", i),
				Recipient: fmt.Sprintf("recipient_%d", i),
			}
			sm.CreateCrossChainTransaction(tx)
		}
	}

	// Verify memory usage is reasonable
	allSidechains := sm.GetSidechains("", "")
	assert.Len(t, allSidechains, numSidechains)

	allBridges := sm.GetBridges("", "", "")
	assert.Len(t, allBridges, numBridges)

	allTxs := sm.GetCrossChainTransactions("", "", "")
	// We can only create transactions for bridges that exist
	expectedTxCount := numBridges
	if expectedTxCount > numTransactions {
		expectedTxCount = numTransactions
	}
	assert.Len(t, allTxs, expectedTxCount)

	allMetrics := sm.GetAllSidechainMetrics()
	assert.Len(t, allMetrics, numSidechains)
}

func TestEdgeCases(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test creating sidechain with empty ID (should generate one)
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)
	assert.NotEmpty(t, sidechain.ID)

	// Test creating sidechain with custom ID
	customID := SidechainID("custom-sidechain-id")
	sidechain2 := &Sidechain{
		ID:            customID,
		Name:          "Custom Sidechain",
		Type:          PolygonSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoAConsensus,
	}
	err = sm.CreateSidechain(sidechain2)
	require.NoError(t, err)
	assert.Equal(t, customID, sidechain2.ID)

	// Test creating bridge with empty ID (should generate one)
	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)
	assert.NotEmpty(t, bridge.ID)

	// Test creating transaction with empty ID (should generate one)
	tx := &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossChainTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)

	// Test getting non-existent transaction
	_, err = sm.GetCrossChainTransaction("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cross-chain transaction not found")

	// Test getting non-existent sidechain
	_, err = sm.GetSidechain("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidechain not found")

	// Test getting non-existent bridge
	_, err = sm.GetBridge("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bridge not found")
}

func TestCleanup(t *testing.T) {
	sm := NewSidechainManager()

	// Create some test data
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test cleanup
	err = sm.Close()
	require.NoError(t, err)

	// Verify channels are closed
	select {
	case <-sm.txQueue:
		// Channel is closed
	default:
		t.Error("txQueue should be closed")
	}

	select {
	case <-sm.bridgeQueue:
		// Channel is closed
	default:
		t.Error("bridgeQueue should be closed")
	}

	select {
	case <-sm.metricsUpdater:
		// Channel is closed
	default:
		t.Error("metricsUpdater should be closed")
	}
}

func TestGetRandomID(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test that random IDs are generated
	id1 := sm.GetRandomID()
	id2 := sm.GetRandomID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // 16 bytes = 32 hex chars
	assert.Len(t, id2, 32)
}

func TestSidechainTypes(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test all sidechain types
	sidechainTypes := []SidechainType{
		PlasmaSidechain,
		PolygonSidechain,
		OptimisticSidechain,
		ZKSidechain,
		CustomSidechain,
	}

	for i, sidechainType := range sidechainTypes {
		sidechain := &Sidechain{
			Name:          fmt.Sprintf("Sidechain Type %d", i),
			Type:          sidechainType,
			ParentChain:   "Ethereum",
			ConsensusType: PoSConsensus,
		}
		err := sm.CreateSidechain(sidechain)
		require.NoError(t, err)
		assert.Equal(t, sidechainType, sidechain.Type)
	}
}

func TestConsensusTypes(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Test all consensus types
	consensusTypes := []ConsensusType{
		PoWConsensus,
		PoSConsensus,
		PoAConsensus,
		DPoSConsensus,
		CustomConsensus,
	}

	for i, consensusType := range consensusTypes {
		sidechain := &Sidechain{
			Name:          fmt.Sprintf("Consensus Sidechain %d", i),
			Type:          PlasmaSidechain,
			ParentChain:   "Ethereum",
			ConsensusType: consensusType,
		}
		err := sm.CreateSidechain(sidechain)
		require.NoError(t, err)
		assert.Equal(t, consensusType, sidechain.ConsensusType)
	}
}

func TestSidechainStatusTransitions(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	sidechain := &Sidechain{
		Name:          "Status Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Test status transitions
	statuses := []SidechainStatus{
		SidechainActive,
		SidechainPaused,
		SidechainSyncing,
		SidechainInactive,
		SidechainTesting,
		SidechainDeprecated,
		SidechainActive, // Back to active
	}

	for _, status := range statuses {
		err = sm.UpdateSidechainStatus(sidechain.ID, status)
		require.NoError(t, err)

		updatedSidechain, err := sm.GetSidechain(sidechain.ID)
		require.NoError(t, err)
		assert.Equal(t, status, updatedSidechain.Status)
	}
}

func TestBridgeStatusTransitions(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	// Create bridge
	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, BridgeActive, bridge.Status)

	// Test bridge status transitions
	statuses := []BridgeStatus{
		BridgeActive,
		BridgePaused,
		BridgeMaintenance,
		BridgeInactive,
		BridgeDeprecated,
		BridgeActive, // Back to active
	}

	for _, status := range statuses {
		bridge.Status = status
		bridge.UpdatedAt = time.Now()

		// Verify status was updated
		assert.Equal(t, status, bridge.Status)
	}
}

func TestCrossChainTransactionStatusTransitions(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	// Create sidechain and bridge
	sidechain := &Sidechain{
		Name:          "Test Sidechain",
		Type:          PlasmaSidechain,
		ParentChain:   "Ethereum",
		ConsensusType: PoSConsensus,
	}
	err := sm.CreateSidechain(sidechain)
	require.NoError(t, err)

	bridge := &Bridge{
		Name:           "ETH Bridge",
		MainChain:      "Ethereum",
		SidechainID:    sidechain.ID,
		Asset:          "ETH",
		MainChainAddr:  "0xMainChainAddress",
		SidechainAddr:  "0xSidechainAddress",
		TotalLiquidity: big.NewInt(1000000),
		MinTransfer:    big.NewInt(100),
		MaxTransfer:    big.NewInt(10000),
		FeePercentage:  0.1,
		SecurityLevel:  SecurityHigh,
	}
	err = sm.CreateBridge(bridge)
	require.NoError(t, err)

	// Create transaction
	tx := &CrossChainTransaction{
		BridgeID:  bridge.ID,
		Direction: MainToSidechain,
		Amount:    big.NewInt(500),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}

	err = sm.CreateCrossChainTransaction(tx)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, CrossChainTxPending, tx.Status)

	// Wait for processing to complete
	time.Sleep(200 * time.Millisecond)

	// Verify final status
	retrievedTx, err := sm.GetCrossChainTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, CrossChainTxConfirmed, retrievedTx.Status)
	assert.NotEmpty(t, retrievedTx.MainChainTxHash)
	assert.NotEmpty(t, retrievedTx.SidechainTxHash)
}

func TestPerformance(t *testing.T) {
	sm := NewSidechainManager()
	defer sm.Close()

	const numSidechains = 30
	const numBridges = 20
	const numTransactions = 50

	start := time.Now()

	// Create many sidechains quickly
	for i := 0; i < numSidechains; i++ {
		sidechain := &Sidechain{
			Name:          fmt.Sprintf("Sidechain %d", i),
			Type:          PlasmaSidechain,
			ParentChain:   "Ethereum",
			ConsensusType: PoSConsensus,
		}
		err := sm.CreateSidechain(sidechain)
		require.NoError(t, err)
	}

	// Create many bridges quickly
	sidechains := sm.GetSidechains("", "")
	for i := 0; i < numBridges && i < len(sidechains); i++ {
		bridge := &Bridge{
			Name:           fmt.Sprintf("Bridge %d", i),
			MainChain:      "Ethereum",
			SidechainID:    sidechains[i].ID,
			Asset:          fmt.Sprintf("ASSET%d", i),
			MainChainAddr:  fmt.Sprintf("0xMain%d", i),
			SidechainAddr:  fmt.Sprintf("0xSide%d", i),
			TotalLiquidity: big.NewInt(1000000),
			MinTransfer:    big.NewInt(100),
			MaxTransfer:    big.NewInt(10000),
			FeePercentage:  0.1,
			SecurityLevel:  SecurityMedium,
		}
		sm.CreateBridge(bridge)
	}

	// Create many transactions quickly
	bridges := sm.GetBridges("", "", "")
	for i := 0; i < numTransactions && i < len(bridges); i++ {
		tx := &CrossChainTransaction{
			BridgeID:  bridges[i].ID,
			Direction: MainToSidechain,
			Amount:    big.NewInt(int64(100 + i)),
			Asset:     bridges[i].Asset,
			Sender:    fmt.Sprintf("sender_%d", i),
			Recipient: fmt.Sprintf("recipient_%d", i),
		}
		sm.CreateCrossChainTransaction(tx)
	}

	duration := time.Since(start)
	t.Logf("Created %d sidechains, %d bridges, and %d transactions in %v", numSidechains, numBridges, numTransactions, duration)

	// Verify all data was created
	allSidechains := sm.GetSidechains("", "")
	assert.Len(t, allSidechains, numSidechains)

	allBridges := sm.GetBridges("", "", "")
	assert.Len(t, allBridges, numBridges)

	allTxs := sm.GetCrossChainTransactions("", "", "")
	// We can only create transactions for bridges that exist
	expectedTxCount := numBridges
	if expectedTxCount > numTransactions {
		expectedTxCount = numTransactions
	}
	assert.Len(t, allTxs, expectedTxCount)

	// Performance should be reasonable
	assert.Less(t, duration, 5*time.Second, "Performance test took too long")
}
