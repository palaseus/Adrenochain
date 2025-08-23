package api

import (
	"context"
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/palaseus/adrenochain/pkg/defi/amm"
	"github.com/palaseus/adrenochain/pkg/defi/governance"
	"github.com/palaseus/adrenochain/pkg/defi/lending"
	"github.com/palaseus/adrenochain/pkg/defi/oracle"
	"github.com/palaseus/adrenochain/pkg/defi/tokens"
	"github.com/palaseus/adrenochain/pkg/defi/yield"
)

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

func TestNewContractAPI(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:        1000000,
		MaxContractSize:    1000,
		DefaultGasPrice:    big.NewInt(20000000000),
		EnableDebugMode:    true,
		EnableMetrics:      true,
		RateLimitPerSecond: 1000,
	}

	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	if api == nil {
		t.Fatal("NewContractAPI returned nil")
	}

	if api.ContractEngine != mockEngine {
		t.Error("ContractEngine not set correctly")
	}

	if api.Config.MaxGasLimit != 1000000 {
		t.Error("Config not set correctly")
	}

	if len(api.ERC20Tokens) != 0 {
		t.Error("ERC20Tokens should be empty initially")
	}

	if len(api.ERC721Tokens) != 0 {
		t.Error("ERC721Tokens should be empty initially")
	}

	if len(api.ERC1155Tokens) != 0 {
		t.Error("ERC1155Tokens should be empty initially")
	}
}

func TestInitializeDeFiComponents(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Create mock DeFi instances
	ammInstance := &amm.AMM{}
	lendingInstance := &lending.LendingProtocol{}
	yieldInstance := &yield.YieldFarm{}
	governanceInstance := &governance.Governance{}
	oracleInstance := &oracle.OracleAggregator{}

	api.InitializeDeFiComponents(ammInstance, lendingInstance, yieldInstance, governanceInstance, oracleInstance)

	if api.AMM != ammInstance {
		t.Error("AMM not initialized correctly")
	}

	if api.Lending != lendingInstance {
		t.Error("Lending not initialized correctly")
	}

	if api.YieldFarming != yieldInstance {
		t.Error("YieldFarming not initialized correctly")
	}

	if api.Governance != governanceInstance {
		t.Error("Governance not initialized correctly")
	}

	if api.OracleAggregator != oracleInstance {
		t.Error("OracleAggregator not initialized correctly")
	}
}

func TestDeployContract(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	bytecode := []byte("0x608060405234")
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	result, err := api.DeployContract(ctx, ContractTypeCustom, bytecode, []interface{}{}, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployContract failed: %v", err)
	}

	if result == nil {
		t.Fatal("DeployContract returned nil result")
	}

	if result.ContractType != ContractTypeCustom {
		t.Error("ContractType not set correctly")
	}

	if result.GasUsed != gasLimit {
		t.Error("GasUsed not set correctly")
	}

	if result.GasPrice.Cmp(gasPrice) != 0 {
		t.Error("GasPrice not set correctly")
	}

	if len(result.Bytecode) != len(bytecode) {
		t.Error("Bytecode not set correctly")
	}

	if api.TotalContracts != 1 {
		t.Error("TotalContracts not incremented")
	}

	if api.TotalCalls != 1 {
		t.Error("TotalCalls not incremented")
	}
}

func TestDeployContractWithDefaultGas(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	bytecode := []byte("0x608060405234")

	result, err := api.DeployContract(ctx, ContractTypeCustom, bytecode, []interface{}{}, 0, nil)
	if err != nil {
		t.Fatalf("DeployContract failed: %v", err)
	}

	if result.GasUsed != 1000000 { // Default gas limit
		t.Error("Default gas limit not used")
	}
}

func TestDeployContractValidationErrors(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()

	// Test invalid bytecode
	_, err := api.DeployContract(ctx, ContractTypeCustom, nil, []interface{}{}, 1000000, big.NewInt(20000000000))
	if err != ErrInvalidBytecode {
		t.Errorf("Expected ErrInvalidBytecode, got %v", err)
	}

	// Test invalid gas price
	_, err = api.DeployContract(ctx, ContractTypeCustom, []byte("valid"), []interface{}{}, 1000000, big.NewInt(0))
	if err != ErrInvalidGasPrice {
		t.Errorf("Expected ErrInvalidGasPrice, got %v", err)
	}

	// Test gas limit exceeded
	_, err = api.DeployContract(ctx, ContractTypeCustom, []byte("valid"), []interface{}{}, 2000000, big.NewInt(20000000000))
	if err != ErrGasLimitExceeded {
		t.Errorf("Expected ErrGasLimitExceeded, got %v", err)
	}
}

func TestCallContract(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Deploy a contract first
	ctx := context.Background()
	bytecode := []byte("0x608060405234")
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	deployResult, err := api.DeployContract(ctx, ContractTypeERC20, bytecode, []interface{}{}, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployContract failed: %v", err)
	}

	// Now call the contract
	callResult, err := api.CallContract(ctx, deployResult.ContractAddress, "transfer", []interface{}{}, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("CallContract failed: %v", err)
	}

	if callResult == nil {
		t.Fatal("CallContract returned nil result")
	}

	if callResult.ContractAddress != deployResult.ContractAddress {
		t.Error("ContractAddress not set correctly")
	}

	if callResult.Method != "transfer" {
		t.Error("Method not set correctly")
	}

	if callResult.GasUsed != gasLimit {
		t.Error("GasUsed not set correctly")
	}

	if callResult.GasPrice.Cmp(gasPrice) != 0 {
		t.Error("GasPrice not set correctly")
	}

	if api.TotalCalls != 2 { // 1 deploy + 1 call
		t.Error("TotalCalls not incremented correctly")
	}
}

func TestCallContractValidationErrors(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()

	// Test invalid contract address
	_, err := api.CallContract(ctx, engine.Address{}, "method", []interface{}{}, 1000000, big.NewInt(20000000000))
	if err != ErrInvalidContractAddress {
		t.Errorf("Expected ErrInvalidContractAddress, got %v", err)
	}

	// Test invalid method
	_, err = api.CallContract(ctx, engine.Address{1}, "", []interface{}{}, 1000000, big.NewInt(20000000000))
	if err != ErrInvalidMethod {
		t.Errorf("Expected ErrInvalidMethod, got %v", err)
	}

	// Test invalid gas price
	_, err = api.CallContract(ctx, engine.Address{1}, "method", []interface{}{}, 1000000, big.NewInt(0))
	if err != ErrInvalidGasPrice {
		t.Errorf("Expected ErrInvalidGasPrice, got %v", err)
	}
}

func TestCallContractNotFound(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	nonExistentAddress := engine.Address{1, 2, 3, 4, 5}

	_, err := api.CallContract(ctx, nonExistentAddress, "method", []interface{}{}, 1000000, big.NewInt(20000000000))
	if err != ErrContractNotFound {
		t.Errorf("Expected ErrContractNotFound, got %v", err)
	}
}

func TestDeployERC20(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	name := "Test Token"
	symbol := "TEST"
	decimals := uint8(18)
	totalSupply := big.NewInt(1000000000000000000) // 1000 tokens
	owner := engine.Address{1, 2, 3, 4, 5}
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	result, err := api.DeployERC20(ctx, name, symbol, decimals, totalSupply, owner, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployERC20 failed: %v", err)
	}

	if result == nil {
		t.Fatal("DeployERC20 returned nil result")
	}

	if result.ContractType != ContractTypeERC20 {
		t.Error("ContractType not set correctly")
	}

	// Check if token was registered
	if len(api.ERC20Tokens) != 1 {
		t.Error("ERC20 token not registered")
	}

	// Verify token details
	for _, token := range api.ERC20Tokens {
		if token.Name != name {
			t.Error("Token name not set correctly")
		}
		if token.Symbol != symbol {
			t.Error("Token symbol not set correctly")
		}
		if token.Decimals != decimals {
			t.Error("Token decimals not set correctly")
		}
		if token.TotalSupply.Cmp(totalSupply) != 0 {
			t.Error("Token total supply not set correctly")
		}
		if token.Owner != owner {
			t.Error("Token owner not set correctly")
		}
		break
	}
}

func TestDeployERC721(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	name := "Test NFT"
	symbol := "TNFT"
	baseURI := "https://example.com/metadata/"
	owner := engine.Address{1, 2, 3, 4, 5}
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	result, err := api.DeployERC721(ctx, name, symbol, baseURI, owner, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployERC721 failed: %v", err)
	}

	if result == nil {
		t.Fatal("DeployERC721 returned nil result")
	}

	if result.ContractType != ContractTypeERC721 {
		t.Error("ContractType not set correctly")
	}

	// Check if token was registered
	if len(api.ERC721Tokens) != 1 {
		t.Error("ERC721 token not registered")
	}

	// Verify token details
	for _, token := range api.ERC721Tokens {
		if token.Name != name {
			t.Error("Token name not set correctly")
		}
		if token.Symbol != symbol {
			t.Error("Token symbol not set correctly")
		}
		if token.BaseURI != baseURI {
			t.Error("Token base URI not set correctly")
		}
		if token.Owner != owner {
			t.Error("Token owner not set correctly")
		}
		break
	}
}

func TestDeployERC1155(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()
	uri := "https://example.com/metadata/"
	owner := engine.Address{1, 2, 3, 4, 5}
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	result, err := api.DeployERC1155(ctx, uri, owner, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployERC1155 failed: %v", err)
	}

	if result == nil {
		t.Fatal("DeployERC1155 returned nil result")
	}

	if result.ContractType != ContractTypeERC1155 {
		t.Error("ContractType not set correctly")
	}

	// Check if token was registered
	if len(api.ERC1155Tokens) != 1 {
		t.Error("ERC1155 token not registered")
	}

	// Verify token details
	for _, token := range api.ERC1155Tokens {
		if token.URI != uri {
			t.Error("Token URI not set correctly")
		}
		if token.Owner != owner {
			t.Error("Token owner not set correctly")
		}
		break
	}
}

func TestGetContractInfo(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()

	// Deploy an ERC20 token first
	name := "Test Token"
	symbol := "TEST"
	decimals := uint8(18)
	totalSupply := big.NewInt(1000000000000000000)
	owner := engine.Address{1, 2, 3, 4, 5}
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	deployResult, err := api.DeployERC20(ctx, name, symbol, decimals, totalSupply, owner, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployERC20 failed: %v", err)
	}

	// Get contract info
	info := api.GetContractInfo(deployResult.ContractAddress)

	if info == nil {
		t.Fatal("GetContractInfo returned nil")
	}

	if info.Address != deployResult.ContractAddress {
		t.Error("Address not set correctly")
	}

	if info.Type != ContractTypeERC20 {
		t.Error("Type not set correctly")
	}

	if info.Name != name {
		t.Error("Name not set correctly")
	}

	if info.Symbol != symbol {
		t.Error("Symbol not set correctly")
	}

	if info.Decimals != decimals {
		t.Error("Decimals not set correctly")
	}

	if info.TotalSupply.Cmp(totalSupply) != 0 {
		t.Error("TotalSupply not set correctly")
	}

	if info.Owner != owner {
		t.Error("Owner not set correctly")
	}
}

func TestGetContractInfoNotFound(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	nonExistentAddress := engine.Address{1, 2, 3, 4, 5}

	info := api.GetContractInfo(nonExistentAddress)
	if info != nil {
		t.Error("Expected nil for non-existent contract")
	}
}

func TestGetAPIStats(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Deploy a contract to increase stats
	ctx := context.Background()
	bytecode := []byte("0x608060405234")
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	_, err := api.DeployContract(ctx, ContractTypeCustom, bytecode, []interface{}{}, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployContract failed: %v", err)
	}

	// Get stats
	stats := api.GetAPIStats()

	if stats.TotalContracts != 1 {
		t.Error("TotalContracts not set correctly")
	}

	if stats.TotalCalls != 1 {
		t.Error("TotalCalls not set correctly")
	}

	if stats.Config.MaxGasLimit != config.MaxGasLimit {
		t.Error("Config not set correctly")
	}
}

func TestContractExists(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()

	// Initially no contracts exist
	if api.contractExists(engine.Address{1, 2, 3, 4, 5}) {
		t.Error("Contract should not exist initially")
	}

	// Deploy a contract
	bytecode := []byte("0x608060405234")
	gasLimit := uint64(500000)
	gasPrice := big.NewInt(20000000000)

	deployResult, err := api.DeployContract(ctx, ContractTypeERC20, bytecode, []interface{}{}, gasLimit, gasPrice)
	if err != nil {
		t.Fatalf("DeployContract failed: %v", err)
	}

	// Now contract should exist
	if !api.contractExists(deployResult.ContractAddress) {
		t.Error("Contract should exist after deployment")
	}
}

func TestConcurrency(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test concurrent deployments
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			ctx := context.Background()
			bytecode := []byte("0x608060405234")
			gasLimit := uint64(500000)
			gasPrice := big.NewInt(20000000000)

			_, err := api.DeployContract(ctx, ContractTypeCustom, bytecode, []interface{}{}, gasLimit, gasPrice)
			if err != nil {
				t.Errorf("Concurrent deployment %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all contracts were deployed
	if api.TotalContracts != numGoroutines {
		t.Errorf("Expected %d contracts, got %d", numGoroutines, api.TotalContracts)
	}
}

func TestEdgeCases(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
		DefaultGasPrice: big.NewInt(20000000000),
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	ctx := context.Background()

	// Test with empty bytecode (should fail validation)
	_, err := api.DeployContract(ctx, ContractTypeCustom, []byte{}, []interface{}{}, 1000000, big.NewInt(20000000000))
	if err != ErrInvalidBytecode {
		t.Errorf("Expected ErrInvalidBytecode for empty bytecode, got %v", err)
	}

	// Test with very large gas limit (should fail validation)
	_, err = api.DeployContract(ctx, ContractTypeCustom, []byte("valid"), []interface{}{}, 999999999, big.NewInt(20000000000))
	if err != ErrGasLimitExceeded {
		t.Errorf("Expected ErrGasLimitExceeded for large gas limit, got %v", err)
	}

	// Test with negative gas price (should fail validation)
	_, err = api.DeployContract(ctx, ContractTypeCustom, []byte("valid"), []interface{}{}, 1000000, big.NewInt(-1))
	if err != ErrInvalidGasPrice {
		t.Errorf("Expected ErrInvalidGasPrice for negative gas price, got %v", err)
	}
}

func TestGetContractInfoEdgeCases(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test with non-existent contract address
	nonExistentAddr := engine.Address{0xFF}
	info := api.GetContractInfo(nonExistentAddr)
	if info != nil {
		t.Error("GetContractInfo should return nil for non-existent contract")
	}

	// Test with ERC20 token that has all fields populated
	erc20Addr := engine.Address{0x01}
	token := tokens.NewERC20Token("TestToken", "TEST", 18, big.NewInt(1000000), engine.Address{0x02}, tokens.DefaultTokenConfig())
	api.ERC20Tokens[erc20Addr] = token

	info = api.GetContractInfo(erc20Addr)
	if info == nil {
		t.Fatal("GetContractInfo should return info for existing ERC20 token")
	}
	if info.Type != ContractTypeERC20 {
		t.Error("Contract type should be ERC20")
	}
	if info.Name != "TestToken" {
		t.Error("Token name should match")
	}
	if info.Symbol != "TEST" {
		t.Error("Token symbol should match")
	}
	if info.Decimals != 18 {
		t.Error("Token decimals should match")
	}
	if info.TotalSupply.Cmp(big.NewInt(1000000)) != 0 {
		t.Error("Total supply should match")
	}
	if info.Owner != (engine.Address{0x02}) {
		t.Error("Owner should match")
	}

	// Test with ERC721 token
	erc721Addr := engine.Address{0x03}
	erc721Token := tokens.NewERC721Token("TestNFT", "TNFT", "https://example.com/", engine.Address{0x04}, tokens.DefaultERC721TokenConfig())
	api.ERC721Tokens[erc721Addr] = erc721Token

	info = api.GetContractInfo(erc721Addr)
	if info == nil {
		t.Fatal("GetContractInfo should return info for existing ERC721 token")
	}
	if info.Type != ContractTypeERC721 {
		t.Error("Contract type should be ERC721")
	}
	if info.Name != "TestNFT" {
		t.Error("Token name should match")
	}
	if info.Symbol != "TNFT" {
		t.Error("Token symbol should match")
	}
	if info.BaseURI != "https://example.com/" {
		t.Error("Base URI should match")
	}
	if info.Owner != (engine.Address{0x04}) {
		t.Error("Owner should match")
	}

	// Test with ERC1155 token
	erc1155Addr := engine.Address{0x05}
	erc1155Token := tokens.NewERC1155Token("Test1155", engine.Address{0x06}, tokens.DefaultERC1155TokenConfig())
	api.ERC1155Tokens[erc1155Addr] = erc1155Token

	info = api.GetContractInfo(erc1155Addr)
	if info == nil {
		t.Fatal("GetContractInfo should return info for existing ERC1155 token")
	}
	if info.Type != ContractTypeERC1155 {
		t.Error("Contract type should be ERC1155")
	}
	if info.Owner != (engine.Address{0x06}) {
		t.Error("Owner should match")
	}
}

func TestValidateDeployInputEdgeCases(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test with empty bytecode
	err := api.validateDeployInput(ContractTypeCustom, []byte{}, 100000, big.NewInt(20000000000))
	if err != ErrInvalidBytecode {
		t.Errorf("Expected ErrInvalidBytecode, got: %v", err)
	}

	// Test with bytecode too large
	largeBytecode := make([]byte, 1001)
	err = api.validateDeployInput(ContractTypeCustom, largeBytecode, 100000, big.NewInt(20000000000))
	if err != ErrContractTooLarge {
		t.Errorf("Expected ErrContractTooLarge, got: %v", err)
	}

	// Test with gas limit exceeded
	err = api.validateDeployInput(ContractTypeCustom, []byte{0x01, 0x02, 0x03}, 1000001, big.NewInt(20000000000))
	if err != ErrGasLimitExceeded {
		t.Errorf("Expected ErrGasLimitExceeded, got: %v", err)
	}

	// Test with nil gas price
	err = api.validateDeployInput(ContractTypeCustom, []byte{0x01, 0x02, 0x03}, 100000, nil)
	if err != nil {
		t.Errorf("Expected no error with nil gas price, got: %v", err)
	}

	// Test with zero gas price
	err = api.validateDeployInput(ContractTypeCustom, []byte{0x01, 0x02, 0x03}, 100000, big.NewInt(0))
	if err != ErrInvalidGasPrice {
		t.Errorf("Expected ErrInvalidGasPrice, got: %v", err)
	}

	// Test with negative gas price
	err = api.validateDeployInput(ContractTypeCustom, []byte{0x01, 0x02, 0x03}, 100000, big.NewInt(-1))
	if err != ErrInvalidGasPrice {
		t.Errorf("Expected ErrInvalidGasPrice, got: %v", err)
	}

	// Test with valid input
	err = api.validateDeployInput(ContractTypeCustom, []byte{0x01, 0x02, 0x03}, 100000, big.NewInt(20000000000))
	if err != nil {
		t.Errorf("Expected no error with valid input, got: %v", err)
	}
}

func TestContractExistsEdgeCases(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test with empty address
	exists := api.contractExists(engine.Address{})
	if exists {
		t.Error("Empty address should not exist")
	}

	// Test with non-existent address
	nonExistentAddr := engine.Address{0xFF}
	exists = api.contractExists(nonExistentAddr)
	if exists {
		t.Error("Non-existent address should not exist")
	}

	// Test with existing ERC20 address
	erc20Addr := engine.Address{0x01}
	token := tokens.NewERC20Token("Test", "TST", 18, big.NewInt(1000), engine.Address{0x02}, tokens.DefaultTokenConfig())
	api.ERC20Tokens[erc20Addr] = token
	exists = api.contractExists(erc20Addr)
	if !exists {
		t.Error("Existing ERC20 address should exist")
	}

	// Test with existing ERC721 address
	erc721Addr := engine.Address{0x03}
	erc721Token := tokens.NewERC721Token("TestNFT", "TNFT", "https://example.com/", engine.Address{0x04}, tokens.DefaultERC721TokenConfig())
	api.ERC721Tokens[erc721Addr] = erc721Token
	exists = api.contractExists(erc721Addr)
	if !exists {
		t.Error("Existing ERC721 address should exist")
	}

	// Test with existing ERC1155 address
	erc1155Addr := engine.Address{0x05}
	erc1155Token := tokens.NewERC1155Token("Test1155", engine.Address{0x06}, tokens.DefaultERC1155TokenConfig())
	api.ERC1155Tokens[erc1155Addr] = erc1155Token
	exists = api.contractExists(erc1155Addr)
	if !exists {
		t.Error("Existing ERC1155 address should exist")
	}
}

func TestRegisterContractEdgeCases(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test registering ERC20 contract
	erc20Addr := engine.Address{0x01}
	err := api.registerContract(ContractTypeERC20, erc20Addr, []interface{}{"Test", "TST", uint8(18)})
	if err != nil {
		t.Errorf("Failed to register ERC20 contract: %v", err)
	}
	if len(api.ERC20Tokens) != 1 {
		t.Error("ERC20 token should be registered")
	}
	if api.ERC20Tokens[erc20Addr] == nil {
		t.Error("ERC20 token should be stored")
	}

	// Test registering ERC721 contract
	erc721Addr := engine.Address{0x02}
	err = api.registerContract(ContractTypeERC721, erc721Addr, []interface{}{"TestNFT", "TNFT"})
	if err != nil {
		t.Errorf("Failed to register ERC721 contract: %v", err)
	}
	if len(api.ERC721Tokens) != 1 {
		t.Error("ERC721 token should be registered")
	}
	if api.ERC721Tokens[erc721Addr] == nil {
		t.Error("ERC721 token should be stored")
	}

	// Test registering ERC1155 contract
	erc1155Addr := engine.Address{0x03}
	err = api.registerContract(ContractTypeERC1155, erc1155Addr, []interface{}{})
	if err != nil {
		t.Errorf("Failed to register ERC1155 contract: %v", err)
	}
	if len(api.ERC1155Tokens) != 1 {
		t.Error("ERC1155 token should be registered")
	}
	if api.ERC1155Tokens[erc1155Addr] == nil {
		t.Error("ERC1155 token should be stored")
	}

	// Test registering custom contract
	customAddr := engine.Address{0x04}
	err = api.registerContract(ContractTypeCustom, customAddr, []interface{}{})
	if err != nil {
		t.Errorf("Failed to register custom contract: %v", err)
	}
	// Custom contracts don't get stored in token maps, so no need to check
}

func TestGenerateContractAddress(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test initial address generation
	initialAddr := api.generateContractAddress()
	if initialAddr == (engine.Address{}) {
		t.Error("Generated address should not be empty")
	}

	// Test that addresses are unique (but TotalContracts doesn't increment automatically)
	// The address generation depends on TotalContracts, which only increments during deployment
	addr1 := api.generateContractAddress()
	addr2 := api.generateContractAddress()
	// Since TotalContracts is 0, both addresses will have byte[0] = 1
	if addr1[0] != 1 || addr2[0] != 1 {
		t.Error("Address first byte should be 1 when TotalContracts is 0")
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test that timestamp is returned
	timestamp := api.getCurrentTimestamp()
	if timestamp != 0 {
		t.Error("Mock timestamp should be 0")
	}
}

func TestDeployContractEdgeCases(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test with valid bytecode
	bytecode := []byte{0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xF3} // Simple contract
	args := []interface{}{"Test", "TST", uint8(18)}

	result, err := api.DeployContract(context.Background(), ContractTypeERC20, bytecode, args, 100000, big.NewInt(20000000000))
	if err != nil {
		t.Errorf("DeployContract should succeed with valid input: %v", err)
	}
	if result == nil {
		t.Fatal("DeployContract should return result")
	}
	if result.ContractType != ContractTypeERC20 {
		t.Error("Contract type should match")
	}
	if result.Bytecode == nil {
		t.Error("Bytecode should be returned")
	}
	if len(result.ConstructorArgs) != 3 {
		t.Error("Constructor args should be returned")
	}

	// Test with empty bytecode (should fail validation)
	_, err = api.DeployContract(context.Background(), ContractTypeCustom, []byte{}, args, 100000, big.NewInt(20000000000))
	if err != ErrInvalidBytecode {
		t.Errorf("Expected ErrInvalidBytecode, got: %v", err)
	}
}

func TestCallContractEdgeCases(t *testing.T) {
	config := APIConfig{}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test with valid contract address (first register the contract)
	contractAddr := engine.Address{0x01}
	api.registerContract(ContractTypeERC20, contractAddr, []interface{}{})

	method := "transfer"
	args := []interface{}{engine.Address{0x02}, big.NewInt(100)}

	result, err := api.CallContract(context.Background(), contractAddr, method, args, 100000, big.NewInt(20000000000))
	if err != nil {
		t.Errorf("CallContract should succeed with valid input: %v", err)
	}
	if result == nil {
		t.Fatal("CallContract should return result")
	}
	if result.ContractAddress != contractAddr {
		t.Error("Contract address should match")
	}
	if result.Method != method {
		t.Error("Method should match")
	}
	if len(result.Args) != 2 {
		t.Error("Args should be returned")
	}

	// Test with empty contract address (should fail validation)
	_, err = api.CallContract(context.Background(), engine.Address{}, method, args, 100000, big.NewInt(20000000000))
	if err != ErrInvalidContractAddress {
		t.Errorf("Expected ErrInvalidContractAddress, got: %v", err)
	}

	// Test with empty method (should fail validation)
	_, err = api.CallContract(context.Background(), contractAddr, "", args, 100000, big.NewInt(20000000000))
	if err != ErrInvalidMethod {
		t.Errorf("Expected ErrInvalidMethod, got: %v", err)
	}

	// Test with zero gas limit to trigger default gas limit path
	result, err = api.CallContract(context.Background(), contractAddr, method, args, 0, big.NewInt(20000000000))
	if err != nil {
		t.Errorf("CallContract should succeed with zero gas limit: %v", err)
	}
	if result.GasUsed != 1000000 { // Default gas limit
		t.Errorf("Expected default gas limit 1000000, got %d", result.GasUsed)
	}
}

// TestDeployContractRegisterFailure tests registerContract failure path
func TestDeployContractRegisterFailure(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Test registerContract directly with empty address to trigger the error path
	err := api.registerContract(ContractTypeERC20, engine.Address{}, []interface{}{})
	if err != ErrInvalidContractAddress {
		t.Errorf("Expected ErrInvalidContractAddress, got %v", err)
	}

	// Test registerContract with valid address (should succeed)
	validAddr := engine.Address{0x01}
	err = api.registerContract(ContractTypeERC20, validAddr, []interface{}{})
	if err != nil {
		t.Errorf("Expected no error with valid address, got %v", err)
	}
}

// TestDeployContractAddressGenerationFailure tests the error path when generateContractAddress returns empty address
func TestDeployContractAddressGenerationFailure(t *testing.T) {
	config := APIConfig{
		MaxGasLimit:     1000000,
		MaxContractSize: 1000,
	}
	mockEngine := &MockContractEngine{}
	api := NewContractAPI(mockEngine, config)

	// Set TotalContracts to 255 to trigger the overflow condition in generateContractAddress
	api.TotalContracts = 255

	bytecode := []byte{0x60, 0x00, 0x52}
	args := []interface{}{}

	// This should fail because generateContractAddress will return empty address
	_, err := api.DeployContract(context.Background(), ContractTypeERC20, bytecode, args, 100000, big.NewInt(20000000000))
	if err != ErrInvalidContractAddress {
		t.Errorf("Expected ErrInvalidContractAddress when TotalContracts >= 255, got %v", err)
	}
}
