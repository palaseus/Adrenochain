package sdk

import (
	"context"
	"math/big"
	"testing"
	"time"

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

// ============================================================================
// AMM OPERATIONS TESTS
// ============================================================================

func TestCreateAMM_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize AMM component
	ammInstance := &amm.AMM{}
	sdk.AMM = ammInstance

	ctx := context.Background()
	tokenA := createTestAddress(1)
	tokenB := createTestAddress(2)
	fee := big.NewInt(3000) // 0.3%
	owner := createTestAddress(3)

	result, err := sdk.CreateAMM(ctx, tokenA, tokenB, fee, owner)
	if err != nil {
		t.Fatalf("CreateAMM failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateAMM returned nil result")
	}

	if result.TokenA != tokenA {
		t.Error("TokenA not set correctly")
	}

	if result.TokenB != tokenB {
		t.Error("TokenB not set correctly")
	}

	if result.Fee.Cmp(fee) != 0 {
		t.Error("Fee not set correctly")
	}

	if result.Owner != owner {
		t.Error("Owner not set correctly")
	}

	if result.AMM == nil {
		t.Error("AMM instance should not be nil")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestCreateAMM_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	tokenA := createTestAddress(1)
	tokenB := createTestAddress(2)
	fee := big.NewInt(3000)
	owner := createTestAddress(3)

	_, err := sdk.CreateAMM(ctx, tokenA, tokenB, fee, owner)
	if err != ErrAMMNotInitialized {
		t.Errorf("Expected ErrAMMNotInitialized, got %v", err)
	}
}

func TestAddLiquidity_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize AMM component
	ammInstance := &amm.AMM{}
	sdk.AMM = ammInstance

	ctx := context.Background()
	ammAddress := createTestAddress(1)
	amountA := big.NewInt(1000000000000000000) // 1 token
	amountB := big.NewInt(2000000000000000000) // 2 tokens

	result, err := sdk.AddLiquidity(ctx, ammAddress, amountA, amountB)
	if err != nil {
		t.Fatalf("AddLiquidity failed: %v", err)
	}

	if result == nil {
		t.Fatal("AddLiquidity returned nil result")
	}

	if result.AmountA.Cmp(amountA) != 0 {
		t.Error("AmountA not set correctly")
	}

	if result.AmountB.Cmp(amountB) != 0 {
		t.Error("AmountB not set correctly")
	}

	// LP tokens should be amountA + amountB (simplified calculation)
	expectedLPTokens := new(big.Int).Add(amountA, amountB)
	if result.LPTokens.Cmp(expectedLPTokens) != 0 {
		t.Errorf("Expected LP tokens %v, got %v", expectedLPTokens, result.LPTokens)
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestAddLiquidity_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	ammAddress := createTestAddress(1)
	amountA := big.NewInt(1000000000000000000)
	amountB := big.NewInt(2000000000000000000)

	_, err := sdk.AddLiquidity(ctx, ammAddress, amountA, amountB)
	if err != ErrAMMNotInitialized {
		t.Errorf("Expected ErrAMMNotInitialized, got %v", err)
	}
}

func TestSwapTokens_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize AMM component
	ammInstance := &amm.AMM{}
	sdk.AMM = ammInstance

	ctx := context.Background()
	ammAddress := createTestAddress(1)
	tokenIn := createTestAddress(2)
	amountIn := big.NewInt(1000000000000000000) // 1 token
	minAmountOut := big.NewInt(900000000000000000) // 0.9 tokens

	result, err := sdk.SwapTokens(ctx, ammAddress, tokenIn, amountIn, minAmountOut)
	if err != nil {
		t.Fatalf("SwapTokens failed: %v", err)
	}

	if result == nil {
		t.Fatal("SwapTokens returned nil result")
	}

	if result.TokenIn != tokenIn {
		t.Error("TokenIn not set correctly")
	}

	if result.AmountIn.Cmp(amountIn) != 0 {
		t.Error("AmountIn not set correctly")
	}

	if result.MinAmountOut.Cmp(minAmountOut) != 0 {
		t.Error("MinAmountOut not set correctly")
	}

	// AmountOut should be amountIn / 1000 (simplified calculation with 0.1% fee)
	expectedAmountOut := new(big.Int).Div(amountIn, big.NewInt(1000))
	if result.AmountOut.Cmp(expectedAmountOut) != 0 {
		t.Errorf("Expected amount out %v, got %v", expectedAmountOut, result.AmountOut)
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestSwapTokens_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	ammAddress := createTestAddress(1)
	tokenIn := createTestAddress(2)
	amountIn := big.NewInt(1000000000000000000)
	minAmountOut := big.NewInt(900000000000000000)

	_, err := sdk.SwapTokens(ctx, ammAddress, tokenIn, amountIn, minAmountOut)
	if err != ErrAMMNotInitialized {
		t.Errorf("Expected ErrAMMNotInitialized, got %v", err)
	}
}

// ============================================================================
// LENDING OPERATIONS TESTS
// ============================================================================

func TestCreateLendingProtocol_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize Lending component
	lendingInstance := &lending.LendingProtocol{}
	sdk.Lending = lendingInstance

	ctx := context.Background()
	lendingConfig := LendingProtocolConfig{
		ProtocolID:           "lending_v1",
		Name:                 "Test Lending",
		Symbol:               "TLEND",
		Decimals:             18,
		Owner:                createTestAddress(1),
		LiquidationThreshold: big.NewInt(800000000000000000), // 80%
		LiquidationBonus:     big.NewInt(50000000000000000),  // 5%
	}

	result, err := sdk.CreateLendingProtocol(ctx, lendingConfig)
	if err != nil {
		t.Fatalf("CreateLendingProtocol failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateLendingProtocol returned nil result")
	}

	if result.ProtocolID != "lending_v1" {
		t.Error("ProtocolID not set correctly")
	}

	if result.Name != "Test Lending" {
		t.Error("Name not set correctly")
	}

	if result.Symbol != "TLEND" {
		t.Error("Symbol not set correctly")
	}

	if result.Owner != createTestAddress(1) {
		t.Error("Owner not set correctly")
	}

	if result.Protocol == nil {
		t.Error("Protocol instance not set")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestCreateLendingProtocol_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	lendingConfig := LendingProtocolConfig{
		ProtocolID: "lending_v1",
		Name:       "Test Lending",
		Symbol:     "TLEND",
		Decimals:   18,
		Owner:      createTestAddress(1),
	}

	_, err := sdk.CreateLendingProtocol(ctx, lendingConfig)
	if err != ErrLendingNotInitialized {
		t.Errorf("Expected ErrLendingNotInitialized, got %v", err)
	}
}

func TestSupplyAsset_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize Lending component
	lendingInstance := &lending.LendingProtocol{}
	sdk.Lending = lendingInstance

	ctx := context.Background()
	protocolAddress := createTestAddress(1)
	asset := createTestAddress(2)
	amount := big.NewInt(1000000000000000000) // 1 token
	user := createTestAddress(3)

	result, err := sdk.SupplyAsset(ctx, protocolAddress, asset, amount, user)
	if err != nil {
		t.Fatalf("SupplyAsset failed: %v", err)
	}

	if result == nil {
		t.Fatal("SupplyAsset returned nil result")
	}

	if result.Protocol != protocolAddress {
		t.Error("Protocol address not set correctly")
	}

	if result.Asset != asset {
		t.Error("Asset address not set correctly")
	}

	if result.Amount.Cmp(amount) != 0 {
		t.Error("Amount not set correctly")
	}

	if result.User != user {
		t.Error("User address not set correctly")
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestSupplyAsset_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	protocolAddress := createTestAddress(1)
	asset := createTestAddress(2)
	amount := big.NewInt(1000000000000000000)
	user := createTestAddress(3)

	_, err := sdk.SupplyAsset(ctx, protocolAddress, asset, amount, user)
	if err != ErrLendingNotInitialized {
		t.Errorf("Expected ErrLendingNotInitialized, got %v", err)
	}
}

// ============================================================================
// YIELD FARMING TESTS
// ============================================================================

func TestCreateYieldFarm_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize Yield Farming component
	yieldInstance := &yield.YieldFarm{}
	sdk.YieldFarming = yieldInstance

	ctx := context.Background()
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)
	
	yieldConfig := YieldFarmConfig{
		FarmID:          "farm_v1",
		Name:            "Test Farm",
		Symbol:          "TFARM",
		Decimals:        18,
		Owner:           createTestAddress(1),
		RewardToken:     createTestAddress(2),
		StakingToken:    createTestAddress(3),
		RewardPerSecond: big.NewInt(1000000000000000000), // 1 token per second
		StartTime:       startTime,
		EndTime:         endTime,
	}

	result, err := sdk.CreateYieldFarm(ctx, yieldConfig)
	if err != nil {
		t.Fatalf("CreateYieldFarm failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateYieldFarm returned nil result")
	}

	if result.FarmID != "farm_v1" {
		t.Error("FarmID not set correctly")
	}

	if result.Name != "Test Farm" {
		t.Error("Name not set correctly")
	}

	if result.Symbol != "TFARM" {
		t.Error("Symbol not set correctly")
	}

	if result.Owner != createTestAddress(1) {
		t.Error("Owner not set correctly")
	}

	if result.Farm == nil {
		t.Error("Farm instance not set")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestCreateYieldFarm_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)
	
	yieldConfig := YieldFarmConfig{
		FarmID:       "farm_v1",
		Name:         "Test Farm",
		Symbol:       "TFARM",
		Decimals:     18,
		Owner:        createTestAddress(1),
		RewardToken:  createTestAddress(2),
		StakingToken: createTestAddress(3),
		StartTime:    startTime,
		EndTime:      endTime,
	}

	_, err := sdk.CreateYieldFarm(ctx, yieldConfig)
	if err != ErrYieldFarmingNotInitialized {
		t.Errorf("Expected ErrYieldFarmingNotInitialized, got %v", err)
	}
}

// ============================================================================
// GOVERNANCE TESTS
// ============================================================================

func TestCreateGovernance_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize Governance component
	governanceInstance := &governance.Governance{}
	sdk.Governance = governanceInstance

	ctx := context.Background()
	governanceConfig := GovernanceConfig{
		GovernanceID:      "gov_v1",
		Name:              "Test Governance",
		Symbol:            "TGOV",
		Decimals:          18,
		Owner:             createTestAddress(1),
		GovernanceToken:   createTestAddress(2),
		MinQuorum:         big.NewInt(1000000000000000000), // 1 token
		ProposalThreshold: big.NewInt(100000000000000000),  // 0.1 token
		VotingPeriod:      7 * 24 * time.Hour,             // 7 days
		ExecutionDelay:    24 * time.Hour,                  // 1 day
	}

	result, err := sdk.CreateGovernance(ctx, governanceConfig)
	if err != nil {
		t.Fatalf("CreateGovernance failed: %v", err)
	}

	if result == nil {
		t.Fatal("CreateGovernance returned nil result")
	}

	if result.GovernanceID != "gov_v1" {
		t.Error("GovernanceID not set correctly")
	}

	if result.Name != "Test Governance" {
		t.Error("Name not set correctly")
	}

	if result.Symbol != "TGOV" {
		t.Error("Symbol not set correctly")
	}

	if result.Owner != createTestAddress(1) {
		t.Error("Owner not set correctly")
	}

	if result.Governance == nil {
		t.Error("Governance instance not set")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestCreateGovernance_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	governanceConfig := GovernanceConfig{
		GovernanceID: "gov_v1",
		Name:         "Test Governance",
		Symbol:       "TGOV",
		Decimals:     18,
		Owner:        createTestAddress(1),
	}

	_, err := sdk.CreateGovernance(ctx, governanceConfig)
	if err != ErrGovernanceNotInitialized {
		t.Errorf("Expected ErrGovernanceNotInitialized, got %v", err)
	}
}

// ============================================================================
// ORACLE TESTS
// ============================================================================

func TestGetPrice_Success(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)
	
	// Initialize Oracle component
	oracleInstance := &oracle.OracleAggregator{}
	sdk.OracleAggregator = oracleInstance

	ctx := context.Background()
	asset := "ETH"

	result, err := sdk.GetPrice(ctx, asset)
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}

	if result == nil {
		t.Fatal("GetPrice returned nil result")
	}

	if result.Asset != "ETH" {
		t.Error("Asset not set correctly")
	}

	// Price should be 1000000 (simplified implementation)
	expectedPrice := big.NewInt(1000000)
	if result.Price.Cmp(expectedPrice) != 0 {
		t.Errorf("Expected price %v, got %v", expectedPrice, result.Price)
	}

	if result.Time != 0 { // getCurrentTimestamp returns 0 in simplified implementation
		t.Error("Time should be 0 in simplified implementation")
	}

	if sdk.TotalOperations != 1 {
		t.Error("TotalOperations not incremented")
	}
}

func TestGetPrice_NotInitialized(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()
	asset := "ETH"

	_, err := sdk.GetPrice(ctx, asset)
	if err != ErrOracleNotInitialized {
		t.Errorf("Expected ErrOracleNotInitialized, got %v", err)
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestCalculateLPTokens(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	amountA := big.NewInt(1000000000000000000) // 1 token
	amountB := big.NewInt(2000000000000000000) // 2 tokens

	result := sdk.calculateLPTokens(amountA, amountB)
	
	// Should be amountA + amountB (simplified implementation)
	expected := new(big.Int).Add(amountA, amountB)
	if result.Cmp(expected) != 0 {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCalculateSwapOutput(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	amountIn := big.NewInt(1000000000000000000) // 1 token

	result := sdk.calculateSwapOutput(amountIn)
	
	// Should be amountIn / 1000 (simplified implementation with 0.1% fee)
	expected := new(big.Int).Div(amountIn, big.NewInt(1000))
	if result.Cmp(expected) != 0 {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	result := sdk.getCurrentTimestamp()
	
	// Should be 0 in simplified implementation
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestSDKFullInitialization(t *testing.T) {
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

	// Initialize all components
	mockEngine := &MockContractEngine{}
	ammInstance := &amm.AMM{}
	lendingInstance := &lending.LendingProtocol{}
	yieldInstance := &yield.YieldFarm{}
	governanceInstance := &governance.Governance{}
	oracleInstance := &oracle.OracleAggregator{}

	sdk.InitializeComponents(mockEngine, ammInstance, lendingInstance, yieldInstance, governanceInstance, oracleInstance)

	// Test that all components are accessible
	if sdk.ContractEngine != mockEngine {
		t.Error("ContractEngine not accessible after initialization")
	}

	if sdk.AMM != ammInstance {
		t.Error("AMM not accessible after initialization")
	}

	if sdk.Lending != lendingInstance {
		t.Error("Lending not accessible after initialization")
	}

	if sdk.YieldFarming != yieldInstance {
		t.Error("YieldFarming not accessible after initialization")
	}

	if sdk.Governance != governanceInstance {
		t.Error("Governance not accessible after initialization")
	}

	if sdk.OracleAggregator != oracleInstance {
		t.Error("OracleAggregator not accessible after initialization")
	}
}

func TestSDKOperationsCounter(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	// Initialize components
	mockEngine := &MockContractEngine{}
	ammInstance := &amm.AMM{}
	lendingInstance := &lending.LendingProtocol{}
	yieldInstance := &yield.YieldFarm{}
	governanceInstance := &governance.Governance{}
	oracleInstance := &oracle.OracleAggregator{}

	sdk.InitializeComponents(mockEngine, ammInstance, lendingInstance, yieldInstance, governanceInstance, oracleInstance)

	initialOps := sdk.TotalOperations

	// Perform multiple operations
	ctx := context.Background()

	// Create token
	tokenConfig := TokenCreationConfig{
		Name:        "Test Token",
		Symbol:      "TEST",
		Decimals:    18,
		TotalSupply: big.NewInt(1000000000000000000),
		Owner:       createTestAddress(1),
	}
	_, err := sdk.CreateToken(ctx, TokenTypeERC20, tokenConfig)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	// Create AMM
	tokenA := createTestAddress(1)
	tokenB := createTestAddress(2)
	fee := big.NewInt(3000)
	owner := createTestAddress(3)
	_, err = sdk.CreateAMM(ctx, tokenA, tokenB, fee, owner)
	if err != nil {
		t.Fatalf("CreateAMM failed: %v", err)
	}

	// Get price
	_, err = sdk.GetPrice(ctx, "ETH")
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}

	// Verify operations counter
	expectedOps := initialOps + 3
	if sdk.TotalOperations != expectedOps {
		t.Errorf("Expected %d operations, got %d", expectedOps, sdk.TotalOperations)
	}
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestSDKErrorHandling(t *testing.T) {
	config := SDKConfig{}
	sdk := NewGoChainSDK(config)

	ctx := context.Background()

	// Test all error conditions when components aren't initialized
	testCases := []struct {
		name        string
		testFunc    func() error
		expectedErr error
	}{
		{
			name: "CreateAMM without initialization",
			testFunc: func() error {
				_, err := sdk.CreateAMM(ctx, createTestAddress(1), createTestAddress(2), big.NewInt(3000), createTestAddress(3))
				return err
			},
			expectedErr: ErrAMMNotInitialized,
		},
		{
			name: "AddLiquidity without initialization",
			testFunc: func() error {
				_, err := sdk.AddLiquidity(ctx, createTestAddress(1), big.NewInt(1000), big.NewInt(2000))
				return err
			},
			expectedErr: ErrAMMNotInitialized,
		},
		{
			name: "SwapTokens without initialization",
			testFunc: func() error {
				_, err := sdk.SwapTokens(ctx, createTestAddress(1), createTestAddress(2), big.NewInt(1000), big.NewInt(900))
				return err
			},
			expectedErr: ErrAMMNotInitialized,
		},
		{
			name: "CreateLendingProtocol without initialization",
			testFunc: func() error {
				_, err := sdk.CreateLendingProtocol(ctx, LendingProtocolConfig{
					ProtocolID: "test",
					Name:       "Test",
					Symbol:     "TEST",
					Decimals:   18,
					Owner:      createTestAddress(1),
				})
				return err
			},
			expectedErr: ErrLendingNotInitialized,
		},
		{
			name: "SupplyAsset without initialization",
			testFunc: func() error {
				_, err := sdk.SupplyAsset(ctx, createTestAddress(1), createTestAddress(2), big.NewInt(1000), createTestAddress(3))
				return err
			},
			expectedErr: ErrLendingNotInitialized,
		},
		{
			name: "CreateYieldFarm without initialization",
			testFunc: func() error {
				_, err := sdk.CreateYieldFarm(ctx, YieldFarmConfig{
					FarmID:       "test",
					Name:         "Test",
					Symbol:       "TEST",
					Decimals:     18,
					Owner:        createTestAddress(1),
					RewardToken:  createTestAddress(2),
					StakingToken: createTestAddress(3),
					StartTime:    time.Now(),
					EndTime:      time.Now().Add(time.Hour),
				})
				return err
			},
			expectedErr: ErrYieldFarmingNotInitialized,
		},
		{
			name: "CreateGovernance without initialization",
			testFunc: func() error {
				_, err := sdk.CreateGovernance(ctx, GovernanceConfig{
					GovernanceID: "test",
					Name:         "Test",
					Symbol:       "TEST",
					Decimals:     18,
					Owner:        createTestAddress(1),
					GovernanceToken: createTestAddress(2),
					VotingPeriod: time.Hour,
				})
				return err
			},
			expectedErr: ErrGovernanceNotInitialized,
		},
		{
			name: "GetPrice without initialization",
			testFunc: func() error {
				_, err := sdk.GetPrice(ctx, "ETH")
				return err
			},
			expectedErr: ErrOracleNotInitialized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			if err != tc.expectedErr {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
