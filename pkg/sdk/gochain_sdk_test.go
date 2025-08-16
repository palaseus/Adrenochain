package sdk

import (
	"context"
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
	"github.com/gochain/gochain/pkg/defi/amm"
	"github.com/gochain/gochain/pkg/defi/governance"
	"github.com/gochain/gochain/pkg/defi/lending"
	"github.com/gochain/gochain/pkg/defi/oracle"
	"github.com/gochain/gochain/pkg/defi/yield"
)

// Helper function to create test addresses
func createTestAddress(id byte) engine.Address {
	var addr engine.Address
	addr[0] = id
	return addr
}

// Mock contract engine for testing
type MockContractEngine struct{}

func (m *MockContractEngine) Execute(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	return &engine.ExecutionResult{
		Success:      true,
		ReturnData:   []byte("success"),
		GasUsed:      gas,
		GasRemaining: 0,
		Logs:         []engine.Log{},
		Error:        nil,
		StateChanges: []engine.StateChange{},
	}, nil
}

func (m *MockContractEngine) Deploy(code []byte, constructor []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.Contract, *engine.ExecutionResult, error) {
	contract := engine.NewContract(engine.Address{}, code, sender)
	result := &engine.ExecutionResult{
		Success:      true,
		ReturnData:   []byte{},
		GasUsed:      gas,
		GasRemaining: 0,
		Logs:         []engine.Log{},
		Error:        nil,
		StateChanges: []engine.StateChange{},
	}
	return contract, result, nil
}

func (m *MockContractEngine) EstimateGas(contract *engine.Contract, input []byte, sender engine.Address, value *big.Int) (uint64, error) {
	return 100000, nil
}

func (m *MockContractEngine) Call(contract *engine.Contract, input []byte, sender engine.Address) ([]byte, error) {
	return []byte("success"), nil
}

func TestNewGoChainSDK(t *testing.T) {
	config := SDKConfig{
		NetworkID:       1,
		RPCEndpoint:     "http://localhost:8545",
		ChainID:         1,
		DefaultGasPrice: big.NewInt(20000000000),
		MaxGasLimit:     1000000,
		EnableDebugMode: true,
		EnableMetrics:   true,
	}

	sdk := NewGoChainSDK(config)

	if sdk == nil {
		t.Fatal("NewGoChainSDK returned nil")
	}

	if sdk.Config.NetworkID != 1 {
		t.Error("NetworkID not set correctly")
	}

	if sdk.Config.RPCEndpoint != "http://localhost:8545" {
		t.Error("RPCEndpoint not set correctly")
	}

	if sdk.Config.ChainID != 1 {
		t.Error("ChainID not set correctly")
	}

	if sdk.Config.DefaultGasPrice.Cmp(big.NewInt(20000000000)) != 0 {
		t.Error("DefaultGasPrice not set correctly")
	}

	if sdk.Config.MaxGasLimit != 1000000 {
		t.Error("MaxGasLimit not set correctly")
	}

	if !sdk.Config.EnableDebugMode {
		t.Error("EnableDebugMode not set correctly")
	}

	if !sdk.Config.EnableMetrics {
		t.Error("EnableMetrics not set correctly")
	}

	if sdk.TotalOperations != 0 {
		t.Error("TotalOperations should be 0 initially")
	}
}

func TestInitializeComponents(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	// Create mock instances
	mockEngine := &MockContractEngine{}
	ammInstance := &amm.AMM{}
	lendingInstance := &lending.LendingProtocol{}
	yieldInstance := &yield.YieldFarm{}
	governanceInstance := &governance.Governance{}
	oracleInstance := &oracle.OracleAggregator{}

	sdk.InitializeComponents(mockEngine, ammInstance, lendingInstance, yieldInstance, governanceInstance, oracleInstance)

	if sdk.ContractEngine != mockEngine {
		t.Error("ContractEngine not initialized correctly")
	}

	if sdk.AMM != ammInstance {
		t.Error("AMM not initialized correctly")
	}

	if sdk.Lending != lendingInstance {
		t.Error("Lending not initialized correctly")
	}

	if sdk.YieldFarming != yieldInstance {
		t.Error("YieldFarming not initialized correctly")
	}

	if sdk.Governance != governanceInstance {
		t.Error("Governance not initialized correctly")
	}

	if sdk.OracleAggregator != oracleInstance {
		t.Error("OracleAggregator not initialized correctly")
	}
}

func TestCreateToken_ERC20(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	tokenConfig := TokenCreationConfig{
		Name:        "Test Token",
		Symbol:      "TEST",
		Decimals:    18,
		TotalSupply: big.NewInt(1000000000000000000),
		Owner:       createTestAddress(1),
	}

	result, err := sdk.CreateToken(ctx, TokenTypeERC20, tokenConfig)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateToken returned nil result")
	}

	if result.Type != TokenTypeERC20 {
		t.Error("Token type not set correctly")
	}

	if result.Name != "Test Token" {
		t.Error("Token name not set correctly")
	}

	if result.Symbol != "TEST" {
		t.Error("Token symbol not set correctly")
	}

	if result.Decimals != 18 {
		t.Error("Token decimals not set correctly")
	}

	if result.TotalSupply.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Error("Token total supply not set correctly")
	}

	if result.Owner != createTestAddress(1) {
		t.Error("Token owner not set correctly")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestCreateToken_ERC721(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	tokenConfig := TokenCreationConfig{
		Name:    "Test NFT",
		Symbol:  "TNFT",
		BaseURI: "https://example.com/metadata/",
		Owner:   createTestAddress(2),
	}

	result, err := sdk.CreateToken(ctx, TokenTypeERC721, tokenConfig)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateToken returned nil result")
	}

	if result.Type != TokenTypeERC721 {
		t.Error("Token type not set correctly")
	}

	if result.Name != "Test NFT" {
		t.Error("Token name not set correctly")
	}

	if result.Symbol != "TNFT" {
		t.Error("Token symbol not set correctly")
	}

	if result.BaseURI != "https://example.com/metadata/" {
		t.Error("Token base URI not set correctly")
	}

	if result.Owner != createTestAddress(2) {
		t.Error("Token owner not set correctly")
	}
}

func TestCreateToken_ERC1155(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	tokenConfig := TokenCreationConfig{
		URI:   "https://example.com/metadata/",
		Owner: createTestAddress(3),
	}

	result, err := sdk.CreateToken(ctx, TokenTypeERC1155, tokenConfig)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateToken returned nil result")
	}

	if result.Type != TokenTypeERC1155 {
		t.Error("Token type not set correctly")
	}

	if result.URI != "https://example.com/metadata/" {
		t.Error("Token URI not set correctly")
	}

	if result.Owner != createTestAddress(3) {
		t.Error("Token owner not set correctly")
	}
}

func TestCreateToken_UnsupportedType(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	tokenConfig := TokenCreationConfig{
		Name:   "Test Token",
		Symbol: "TEST",
		Owner:  createTestAddress(4),
	}

	_, err := sdk.CreateToken(ctx, "unsupported", tokenConfig)
	if err != ErrUnsupportedTokenType {
		t.Errorf("Expected ErrUnsupportedTokenType, got %v", err)
	}
}

func TestGenerateAddress(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	// Generate first address
	addr1 := sdk.generateAddress()
	
	// Increment operations to get different addresses
	sdk.TotalOperations++
	addr2 := sdk.generateAddress()
	
	sdk.TotalOperations++
	addr3 := sdk.generateAddress()

	if addr1 == addr2 {
		t.Error("Generated addresses should be unique")
	}

	if addr2 == addr3 {
		t.Error("Generated addresses should be unique")
	}

	if addr1 == addr3 {
		t.Error("Generated addresses should be unique")
	}

	// Check that addresses are not empty
	if addr1 == (engine.Address{}) {
		t.Error("Generated address should not be empty")
	}

	if addr2 == (engine.Address{}) {
		t.Error("Generated address should not be empty")
	}

	if addr3 == (engine.Address{}) {
		t.Error("Generated address should not be empty")
	}
}

func TestConcurrency(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	// Test concurrent token creation
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			ctx := context.Background()
			tokenConfig := TokenCreationConfig{
				Name:        "Test Token",
				Symbol:      "TEST",
				Decimals:    18,
				TotalSupply: big.NewInt(1000000000000000000),
				Owner:       createTestAddress(byte(id)),
			}

			_, err := sdk.CreateToken(ctx, TokenTypeERC20, tokenConfig)
			if err != nil {
				t.Errorf("Concurrent token creation %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all operations were recorded
	if sdk.TotalOperations != numGoroutines {
		t.Errorf("Expected %d operations, got %d", numGoroutines, sdk.TotalOperations)
	}
}
