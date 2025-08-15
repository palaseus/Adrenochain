package evm

import (
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

func TestNewEVMEngine(t *testing.T) {
	// Create mock storage and registry
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}

	evm := NewEVMEngine(mockStorage, mockRegistry)

	if evm == nil {
		t.Fatal("expected non-nil EVM engine")
	}

	if evm.storage != mockStorage {
		t.Error("storage not set correctly")
	}

	if evm.registry != mockRegistry {
		t.Error("registry not set correctly")
	}

	if evm.stack == nil {
		t.Error("stack not initialized")
	}

	if evm.memory == nil {
		t.Error("memory not initialized")
	}

	if evm.pc != 0 {
		t.Error("program counter not initialized to 0")
	}
}

func TestEVMEngineExecute(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create a test contract with simple code
	contract := &engine.Contract{
		Address: generateRandomAddress(),
		Code:    []byte{0x00}, // STOP instruction
		Creator: generateRandomAddress(),
	}

	// Test successful execution
	result, err := evm.Execute(contract, nil, 1000, generateRandomAddress(), big.NewInt(0))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful execution")
	}

	if result.GasUsed != 0 { // STOP instruction costs 0 gas
		t.Errorf("expected 0 gas used, got %d", result.GasUsed)
	}

	// Test nil contract
	_, err = evm.Execute(nil, nil, 1000, generateRandomAddress(), big.NewInt(0))
	if err == nil {
		t.Error("expected error for nil contract")
	}
	if err != engine.ErrInvalidContract {
		t.Errorf("expected ErrInvalidContract, got %v", err)
	}

	// Test empty contract code
	emptyContract := &engine.Contract{
		Address: generateRandomAddress(),
		Code:    []byte{},
		Creator: generateRandomAddress(),
	}
	_, err = evm.Execute(emptyContract, nil, 1000, generateRandomAddress(), big.NewInt(0))
	if err == nil {
		t.Error("expected error for empty contract code")
	}
	if err != engine.ErrInvalidContract {
		t.Errorf("expected ErrInvalidContract, got %v", err)
	}
}

func TestEVMEngineDeploy(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Test deployment without constructor
	code := []byte{0x00} // STOP instruction
	sender := generateRandomAddress()

	contract, result, err := evm.Deploy(code, nil, 1000, sender, big.NewInt(0))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if contract == nil {
		t.Fatal("expected non-nil contract")
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if !result.Success {
		t.Error("expected successful deployment")
	}

	if contract.Creator != sender {
		t.Errorf("expected creator %v, got %v", sender, contract.Creator)
	}

	// Test deployment with constructor
	constructorCode := []byte{0x00} // STOP instruction
	contract, result, err = evm.Deploy(code, constructorCode, 1000, sender, big.NewInt(0))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful deployment with constructor")
	}
}

func TestEVMEngineEstimateGas(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create a test contract
	contract := &engine.Contract{
		Address: generateRandomAddress(),
		Code:    []byte{0x00}, // STOP instruction
		Creator: generateRandomAddress(),
	}

	// Test gas estimation
	estimatedGas, err := evm.EstimateGas(contract, nil, generateRandomAddress(), big.NewInt(0))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// STOP instruction costs 0 gas, but we add buffer
	if estimatedGas < 21000 {
		t.Errorf("expected estimated gas >= 21000, got %d", estimatedGas)
	}
}

func TestEVMEngineCall(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create a test contract
	contract := &engine.Contract{
		Address: generateRandomAddress(),
		Code:    []byte{0x00}, // STOP instruction
		Creator: generateRandomAddress(),
	}

	// Test read-only call
	result, err := evm.Call(contract, nil, generateRandomAddress())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestEVMEngineSetBlockContext(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	blockNum := uint64(12345)
	timestamp := uint64(1609459200)
	coinbase := generateRandomAddress()
	difficulty := big.NewInt(1000000)

	evm.SetBlockContext(blockNum, timestamp, coinbase, difficulty)

	if evm.blockNum != blockNum {
		t.Errorf("expected block number %d, got %d", blockNum, evm.blockNum)
	}

	if evm.timestamp != timestamp {
		t.Errorf("expected timestamp %d, got %d", timestamp, evm.timestamp)
	}

	if evm.coinbase != coinbase {
		t.Errorf("expected coinbase %v, got %v", coinbase, evm.coinbase)
	}

	if evm.difficulty.Cmp(difficulty) != 0 {
		t.Errorf("expected difficulty %v, got %v", difficulty, evm.difficulty)
	}
}

func TestEVMEngineSetGasPrice(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	gasPrice := big.NewInt(20000000000) // 20 gwei
	evm.SetGasPrice(gasPrice)

	if evm.gasPrice.Cmp(gasPrice) != 0 {
		t.Errorf("expected gas price %v, got %v", gasPrice, evm.gasPrice)
	}
}

func TestEVMEngineSetChainID(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	chainID := big.NewInt(1337)
	evm.SetChainID(chainID)

	if evm.chainID.Cmp(chainID) != 0 {
		t.Errorf("expected chain ID %v, got %v", chainID, evm.chainID)
	}
}

func TestEVMEngineClone(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := &MockContractRegistry{}
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Set some context
	evm.SetBlockContext(100, 1609459200, generateRandomAddress(), big.NewInt(1000000))
	evm.SetGasPrice(big.NewInt(20000000000))
	evm.SetChainID(big.NewInt(1337))

	// Clone the engine
	clone := evm.Clone()

	if clone == nil {
		t.Fatal("expected non-nil clone")
	}

	// Verify clone has same values
	if clone.blockNum != evm.blockNum {
		t.Errorf("clone block number mismatch: expected %d, got %d", evm.blockNum, clone.blockNum)
	}

	if clone.timestamp != evm.timestamp {
		t.Errorf("clone timestamp mismatch: expected %d, got %d", evm.timestamp, clone.timestamp)
	}

	if clone.gasPrice.Cmp(evm.gasPrice) != 0 {
		t.Errorf("clone gas price mismatch: expected %v, got %v", evm.gasPrice, clone.gasPrice)
	}

	if clone.chainID.Cmp(evm.chainID) != 0 {
		t.Errorf("clone chain ID mismatch: expected %v, got %v", evm.chainID, clone.chainID)
	}

	// Verify clone is independent
	evm.SetBlockContext(200, 1609545600, generateRandomAddress(), big.NewInt(2000000))

	if clone.blockNum == evm.blockNum {
		t.Error("clone should be independent of original")
	}
}

// Mock implementations for testing
type MockContractStorage struct{}

func (m *MockContractStorage) Get(address engine.Address, key engine.Hash) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockContractStorage) Set(address engine.Address, key engine.Hash, value []byte) error {
	return nil
}

func (m *MockContractStorage) Delete(address engine.Address, key engine.Hash) error {
	return nil
}

func (m *MockContractStorage) GetStorageRoot(address engine.Address) (engine.Hash, error) {
	return engine.Hash{}, nil
}

func (m *MockContractStorage) Commit() error {
	return nil
}

func (m *MockContractStorage) Rollback() error {
	return nil
}

func (m *MockContractStorage) HasKey(address engine.Address, key engine.Hash) bool {
	return false
}

func (m *MockContractStorage) GetContractStorage(address engine.Address) (map[engine.Hash][]byte, error) {
	return make(map[engine.Hash][]byte), nil
}

func (m *MockContractStorage) GetStorageSize(address engine.Address) (int, error) {
	return 0, nil
}

func (m *MockContractStorage) ClearContractStorage(address engine.Address) error {
	return nil
}

func (m *MockContractStorage) GetStorageProof(address engine.Address, key engine.Hash) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockContractStorage) VerifyStorageProof(root engine.Hash, key engine.Hash, value []byte, proof []byte) bool {
	return true
}

type MockContractRegistry struct{}

func (m *MockContractRegistry) Register(contract *engine.Contract) error {
	return nil
}

func (m *MockContractRegistry) Get(address engine.Address) (*engine.Contract, error) {
	return nil, engine.ErrContractNotFound
}

func (m *MockContractRegistry) Exists(address engine.Address) bool {
	return false
}

func (m *MockContractRegistry) Remove(address engine.Address) error {
	return nil
}

func (m *MockContractRegistry) List() []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockContractRegistry) GetContractCount() int {
	return 0
}

func (m *MockContractRegistry) GetContractByCodeHash(codeHash engine.Hash) []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockContractRegistry) GetContractsByCreator(creator engine.Address) []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockContractRegistry) UpdateContract(contract *engine.Contract) error {
	return nil
}

func (m *MockContractRegistry) GetContractAddresses() []engine.Address {
	return []engine.Address{}
}

func (m *MockContractRegistry) HasContracts() bool {
	return false
}

func (m *MockContractRegistry) Clear() {}

func (m *MockContractRegistry) GenerateAddress() engine.Address {
	return generateRandomAddress()
}

func (m *MockContractRegistry) GetContractStats() engine.ContractStats {
	return engine.ContractStats{}
}

// Helper function to generate random addresses for testing
func generateRandomAddress() engine.Address {
	var address engine.Address
	// Fill with some test data
	for i := 0; i < 20; i++ {
		address[i] = byte(i + 1)
	}
	return address
}
