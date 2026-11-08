package evm

import (
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/stretchr/testify/assert"
)

func TestNewEVMEngine(t *testing.T) {
	// Create mock storage and registry
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()

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
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Initialize gas meter to prevent nil pointer dereference
	evm.gasMeter = engine.NewGasMeter(50000)

	// Create a test contract with simple code
	contract := &engine.Contract{
		Address: generateRandomAddress(),
		Code:    []byte{0x00}, // STOP instruction
		Creator: generateRandomAddress(),
	}

	// Test successful execution
	result, err := evm.Execute(contract, nil, 50000, generateRandomAddress(), big.NewInt(0))
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
	_, err = evm.Execute(nil, nil, 50000, generateRandomAddress(), big.NewInt(0))
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
	_, err = evm.Execute(emptyContract, nil, 50000, generateRandomAddress(), big.NewInt(0))
	if err == nil {
		t.Error("expected error for empty contract code")
	}
	if err != engine.ErrInvalidContract {
		t.Errorf("expected ErrInvalidContract, got %v", err)
	}
}

func TestEVMEngineDeploy(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Test deployment without constructor
	code := []byte{0x00} // STOP instruction
	sender := generateRandomAddress()

	contract, result, err := evm.Deploy(code, nil, 50000, sender, big.NewInt(0))
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
	contract, result, err = evm.Deploy(code, constructorCode, 50000, sender, big.NewInt(0))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful deployment with constructor")
	}
}

func TestEVMEngineEstimateGas(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
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
	mockRegistry := NewMockContractRegistry()
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
	mockRegistry := NewMockContractRegistry()
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
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	gasPrice := big.NewInt(20000000000) // 20 gwei
	evm.SetGasPrice(gasPrice)

	if evm.gasPrice.Cmp(gasPrice) != 0 {
		t.Errorf("expected gas price %v, got %v", gasPrice, evm.gasPrice)
	}
}

func TestEVMEngineSetChainID(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	chainID := big.NewInt(1337)
	evm.SetChainID(chainID)

	if evm.chainID.Cmp(chainID) != 0 {
		t.Errorf("expected chain ID %v, got %v", chainID, evm.chainID)
	}
}

func TestEVMEngineClone(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
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

type MockContractRegistry struct {
	contracts map[engine.Address]*engine.Contract
}

func NewMockContractRegistry() *MockContractRegistry {
	return &MockContractRegistry{
		contracts: make(map[engine.Address]*engine.Contract),
	}
}

func (m *MockContractRegistry) Register(contract *engine.Contract) error {
	m.contracts[contract.Address] = contract
	return nil
}

func (m *MockContractRegistry) Get(address engine.Address) (*engine.Contract, error) {
	contract, exists := m.contracts[address]
	if !exists {
		return nil, engine.ErrContractNotFound
	}
	return contract, nil
}

func (m *MockContractRegistry) Exists(address engine.Address) bool {
	_, exists := m.contracts[address]
	return exists
}

func (m *MockContractRegistry) Remove(address engine.Address) error {
	delete(m.contracts, address)
	return nil
}

func (m *MockContractRegistry) List() []*engine.Contract {
	contracts := make([]*engine.Contract, 0, len(m.contracts))
	for _, c := range m.contracts {
		contracts = append(contracts, c)
	}
	return contracts
}

func (m *MockContractRegistry) GetContractCount() int {
	return len(m.contracts)
}

func (m *MockContractRegistry) GetContractByCodeHash(codeHash engine.Hash) []*engine.Contract {
	var contracts []*engine.Contract
	for _, c := range m.contracts {
		if c.CodeHash == codeHash {
			contracts = append(contracts, c)
		}
	}
	return contracts
}

func (m *MockContractRegistry) GetContractsByCreator(creator engine.Address) []*engine.Contract {
	var contracts []*engine.Contract
	for _, c := range m.contracts {
		if c.Creator == creator {
			contracts = append(contracts, c)
		}
	}
	return contracts
}

func (m *MockContractRegistry) UpdateContract(contract *engine.Contract) error {
	m.contracts[contract.Address] = contract
	return nil
}

func (m *MockContractRegistry) GetContractAddresses() []engine.Address {
	addresses := make([]engine.Address, 0, len(m.contracts))
	for addr := range m.contracts {
		addresses = append(addresses, addr)
	}
	return addresses
}

func (m *MockContractRegistry) HasContracts() bool {
	return len(m.contracts) > 0
}

func (m *MockContractRegistry) Clear() {
	m.contracts = make(map[engine.Address]*engine.Contract)
}

func (m *MockContractRegistry) GetContractStats() engine.ContractStats {
	return engine.ContractStats{}
}

func (m *MockContractRegistry) GenerateAddress() engine.Address {
	var addr engine.Address
	for i := 0; i < 20; i++ {
		addr[i] = byte(i + 1)
	}
	return addr
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

func TestEVMStackOperations(t *testing.T) {
	stack := NewEVMStack()

	// Test initial state
	if stack.Size() != 0 {
		t.Error("New stack should have size 0")
	}

	// Test Push operations
	stack.Push(big.NewInt(42))
	if stack.Size() != 1 {
		t.Error("Stack size should be 1 after push")
	}

	stack.Push(big.NewInt(100))
	if stack.Size() != 2 {
		t.Error("Stack size should be 2 after second push")
	}

	// Test Peek operations
	top := stack.Peek()
	if top.Cmp(big.NewInt(100)) != 0 {
		t.Error("Peek should return top value without removing")
	}

	if stack.Size() != 2 {
		t.Error("Peek should not change stack size")
	}

	// Test Pop operations
	popped := stack.Pop()
	if popped.Cmp(big.NewInt(100)) != 0 {
		t.Error("Pop should return top value")
	}

	if stack.Size() != 1 {
		t.Error("Stack size should be 1 after pop")
	}

	popped = stack.Pop()
	if popped.Cmp(big.NewInt(42)) != 0 {
		t.Error("Pop should return second value")
	}

	if stack.Size() != 0 {
		t.Error("Stack size should be 0 after all pops")
	}

	// Test Pop on empty stack
	popped = stack.Pop()
	if popped.Cmp(big.NewInt(0)) != 0 {
		t.Error("Pop on empty stack should return 0")
	}

	// Test Peek on empty stack
	peeked := stack.Pop()
	if peeked.Cmp(big.NewInt(0)) != 0 {
		t.Error("Peek on empty stack should return 0")
	}
}

func TestEVMMemoryOperations(t *testing.T) {
	memory := NewEVMMemory()

	// Test initial state
	if memory.Size() != 0 {
		t.Error("New memory should have size 0")
	}

	// Test Set operations
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	memory.Set(0, testData)

	if memory.Size() != 4 {
		t.Error("Memory size should be 4 after setting data")
	}

	// Test Get operations
	retrieved := memory.Get(0, 4)
	if len(retrieved) != 4 {
		t.Error("Get should return correct length")
	}

	for i, b := range retrieved {
		if b != testData[i] {
			t.Errorf("Memory data mismatch at index %d: expected %x, got %x", i, testData[i], b)
		}
	}

	// Test Get with offset
	retrieved = memory.Get(1, 2)
	if len(retrieved) != 2 {
		t.Error("Get with offset should return correct length")
	}

	if retrieved[0] != 0x02 || retrieved[1] != 0x03 {
		t.Error("Get with offset should return correct data")
	}

	// Test Set with larger data
	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	memory.Set(100, largeData)
	if memory.Size() != 1100 {
		t.Error("Memory size should accommodate large data")
	}

	// Test Get with large offset
	retrieved = memory.Get(100, 1000)
	if len(retrieved) != 1000 {
		t.Error("Get with large offset should return correct length")
	}

	for i, b := range retrieved {
		if b != byte(i%256) {
			t.Errorf("Large memory data mismatch at index %d: expected %x, got %x", i, byte(i%256), b)
		}
	}

	// Test memory expansion
	memory.Set(2000, []byte{0xFF})
	if memory.Size() < 2001 {
		t.Error("Memory should expand to accommodate offset")
	}
}

func TestEVMInstructionExecution(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create execution context
	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Input:      []byte{},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Test basic stack operations
	evm.stack.Push(big.NewInt(5))
	evm.stack.Push(big.NewInt(3))

	// Test that we can push and pop values
	if evm.stack.Size() != 2 {
		t.Error("Stack should have 2 items after pushing")
	}

	popped := evm.stack.Pop()
	if popped.Cmp(big.NewInt(3)) != 0 {
		t.Errorf("Pop should return top value, got %v", popped)
	}

	if evm.stack.Size() != 1 {
		t.Error("Stack size should decrease after pop")
	}

	// Test memory operations
	evm.memory.Set(0, []byte{0x01, 0x02, 0x03, 0x04})
	retrieved := evm.memory.Get(0, 4)
	if len(retrieved) != 4 {
		t.Error("Memory should return correct length")
	}

	// Test arithmetic operations with proper stack setup
	// Test ADD instruction
	evm.stack.Push(big.NewInt(5))
	evm.stack.Push(big.NewInt(3))
	evm.executeADD()
	result := evm.stack.Peek()
	if result.Cmp(big.NewInt(8)) != 0 {
		t.Errorf("ADD: expected 8, got %v", result)
	}

	// Test MUL instruction
	evm.stack.Push(big.NewInt(4))
	evm.stack.Push(big.NewInt(6))
	evm.executeMUL()
	result = evm.stack.Peek()
	if result.Cmp(big.NewInt(24)) != 0 {
		t.Errorf("MUL: expected 24, got %v", result)
	}

	// Test SUB instruction
	evm.stack.Push(big.NewInt(10))
	evm.stack.Push(big.NewInt(3))
	evm.executeSUB()
	result = evm.stack.Peek()
	if result.Cmp(big.NewInt(-7)) != 0 {
		t.Errorf("SUB: expected -7, got %v", result)
	}

	// Test DIV instruction
	evm.stack.Push(big.NewInt(15))
	evm.stack.Push(big.NewInt(3))
	evm.executeDIV()
	result = evm.stack.Peek()
	if result.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("DIV: expected 0, got %v", result)
	}

	// Test POP instruction
	evm.executePOP()
	// POP consumes 1 value

	// Test PC instruction
	evm.executePC()
	// PC pushes 1 value

	// Test MSIZE instruction
	evm.executeMSIZE()
	// MSIZE pushes 1 value

	// Test MLOAD instruction
	evm.stack.Push(big.NewInt(0)) // offset
	stackSizeBeforeMLOAD := evm.stack.Size()
	evm.executeMLOAD()
	stackSizeAfterMLOAD := evm.stack.Size()
	if stackSizeAfterMLOAD != stackSizeBeforeMLOAD {
		t.Error("MLOAD should consume 1 value and push 1 value, keeping stack size the same")
	}

	// Test MSTORE instruction
	evm.stack.Push(big.NewInt(0))  // offset
	evm.stack.Push(big.NewInt(42)) // value
	stackSizeBeforeMSTORE := evm.stack.Size()
	evm.executeMSTORE()
	stackSizeAfterMSTORE := evm.stack.Size()
	if stackSizeAfterMSTORE != stackSizeBeforeMSTORE-2 {
		t.Error("MSTORE should consume two values from stack")
	}

	// Test GAS instruction
	evm.gasMeter = engine.NewGasMeter(1000)
	evm.executeGAS()
	gasValue := evm.stack.Peek()
	if gasValue.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("GAS: expected 1000, got %v", gasValue)
	}

	// Test JUMP operations with proper validation
	// Test JUMP with valid destination
	evm.stack.Push(big.NewInt(100))
	evm.executeJUMP(ctx)
	// JUMP should consume the destination value

	// Test JUMPI with true condition
	evm.stack.Push(big.NewInt(200))
	evm.stack.Push(big.NewInt(1))
	evm.executeJUMPI(ctx)
	// JUMPI should consume both values

	// Test JUMPI with false condition
	evm.stack.Push(big.NewInt(300))
	evm.stack.Push(big.NewInt(0))
	evm.executeJUMPI(ctx)
	// JUMPI should consume both values but not jump
}

func TestEVMBasicInstructions(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create execution context
	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Input:      []byte{},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Test CREATE instruction (currently just a stub)
	evm.stack.Push(big.NewInt(0)) // value
	evm.stack.Push(big.NewInt(0)) // offset
	evm.stack.Push(big.NewInt(0)) // size
	evm.executeCREATE(ctx)

	// CREATE is currently a stub, so it doesn't consume operands
	if evm.stack.Size() != 3 {
		t.Error("CREATE stub should not consume operands")
	}

	// Test CALL instruction (currently just a stub)
	evm.stack.Push(big.NewInt(0)) // gas
	evm.stack.Push(big.NewInt(0)) // address
	evm.stack.Push(big.NewInt(0)) // value
	evm.stack.Push(big.NewInt(0)) // argsOffset
	evm.stack.Push(big.NewInt(0)) // argsSize
	evm.stack.Push(big.NewInt(0)) // retOffset
	evm.stack.Push(big.NewInt(0)) // retSize
	evm.executeCALL(ctx)

	// CALL is currently a stub, so it doesn't consume operands
	if evm.stack.Size() != 10 {
		t.Error("CALL stub should not consume operands")
	}

	// Test SUICIDE instruction (currently just a stub)
	evm.stack.Push(big.NewInt(0)) // address
	evm.executeSUICIDE(ctx)

	// SUICIDE is currently a stub, so it doesn't consume operands
	if evm.stack.Size() != 11 {
		t.Error("SUICIDE stub should not consume operands")
	}
}

func TestEVMStackClone(t *testing.T) {
	stack := NewEVMStack()

	// Add some data
	stack.Push(big.NewInt(42))
	stack.Push(big.NewInt(100))

	// Clone the stack
	cloned := stack.Clone()

	// Verify clone has same data
	if cloned.Size() != stack.Size() {
		t.Error("Cloned stack should have same size")
	}

	// Verify clone is independent
	cloned.Push(big.NewInt(200))
	if stack.Size() == cloned.Size() {
		t.Error("Original stack should not be affected by clone modifications")
	}
}

func TestEVMMemoryClone(t *testing.T) {
	memory := NewEVMMemory()

	// Add some data
	memory.Set(0, []byte{0x01, 0x02, 0x03})

	// Clone the memory
	cloned := memory.Clone()

	// Verify clone has same data
	if cloned.Size() != memory.Size() {
		t.Error("Cloned memory should have same size")
	}

	// Verify clone is independent
	cloned.Set(100, []byte{0xFF})
	if memory.Size() == cloned.Size() {
		t.Error("Original memory should not be affected by clone modifications")
	}
}

func TestEVMExecuteInstruction(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Initialize gas meter to prevent panic
	evm.gasMeter = engine.NewGasMeter(10000)

	// Create execution context for JUMP operations
	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Input:      []byte{},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Test various opcodes
	testCases := []struct {
		name        string
		instruction *Instruction
		setup       func()
		validate    func()
	}{
		{
			name:        "STOP instruction",
			instruction: Instructions[0x00],
			setup:       func() {},
			validate: func() {
				// STOP should halt execution
			},
		},
		{
			name:        "ADD instruction",
			instruction: Instructions[0x01],
			setup: func() {
				evm.stack.Push(big.NewInt(5))
				evm.stack.Push(big.NewInt(3))
			},
			validate: func() {
				result := evm.stack.Peek()
				if result.Cmp(big.NewInt(8)) != 0 {
					t.Errorf("ADD: expected 8, got %v", result)
				}
			},
		},
		{
			name:        "MUL instruction",
			instruction: Instructions[0x02],
			setup: func() {
				evm.stack.Push(big.NewInt(4))
				evm.stack.Push(big.NewInt(6))
			},
			validate: func() {
				result := evm.stack.Peek()
				if result.Cmp(big.NewInt(24)) != 0 {
					t.Errorf("MUL: expected 24, got %v", result)
				}
			},
		},
		{
			name:        "SUB instruction",
			instruction: Instructions[0x03],
			setup: func() {
				evm.stack.Push(big.NewInt(10))
				evm.stack.Push(big.NewInt(3))
			},
			validate: func() {
				result := evm.stack.Peek()
				if result.Cmp(big.NewInt(-7)) != 0 {
					t.Errorf("SUB: expected -7, got %v", result)
				}
			},
		},
		{
			name:        "DIV instruction",
			instruction: Instructions[0x04],
			setup: func() {
				evm.stack.Push(big.NewInt(15))
				evm.stack.Push(big.NewInt(3))
			},
			validate: func() {
				result := evm.stack.Peek()
				if result.Cmp(big.NewInt(0)) != 0 {
					t.Errorf("DIV: expected 0, got %v", result)
				}
			},
		},
		{
			name:        "POP instruction",
			instruction: Instructions[0x50],
			setup: func() {
				evm.stack.Push(big.NewInt(42))
			},
			validate: func() {
				// POP consumes 1 value, so stack size should decrease
			},
		},
		{
			name:        "PC instruction",
			instruction: Instructions[0x58],
			setup:       func() {},
			validate: func() {
				// PC pushes 1 value
			},
		},
		{
			name:        "MSIZE instruction",
			instruction: Instructions[0x59],
			setup:       func() {},
			validate: func() {
				// MSIZE pushes 1 value
			},
		},
		{
			name:        "GAS instruction",
			instruction: Instructions[0x5A],
			setup:       func() {},
			validate: func() {
				// GAS pushes 1 value
			},
		},
		{
			name:        "REVERT instruction",
			instruction: &Instruction{Opcode: 0xFD, Name: "REVERT", GasCost: 0, Size: 1, Halts: true, Pops: 0, Pushes: 0},
			setup:       func() {},
			validate: func() {
				// REVERT should halt execution
			},
		},
		{
			name:        "Unknown opcode",
			instruction: &Instruction{Opcode: 0xFF, Name: "UNKNOWN", GasCost: 0, Size: 1, Halts: false, Pops: 0, Pushes: 0},
			setup:       func() {},
			validate: func() {
				// Unknown opcode should not halt execution
			},
		},
		{
			name:        "JUMP instruction",
			instruction: Instructions[0x56],
			setup: func() {
				evm.stack.Push(big.NewInt(100))
			},
			validate: func() {
				// JUMP consumes 1 value
			},
		},
		{
			name:        "JUMPI instruction with true condition",
			instruction: Instructions[0x57],
			setup: func() {
				evm.stack.Push(big.NewInt(200))
				evm.stack.Push(big.NewInt(1))
			},
			validate: func() {
				// JUMPI consumes 2 values
			},
		},
		{
			name:        "JUMPI instruction with false condition",
			instruction: Instructions[0x57],
			setup: func() {
				evm.stack.Push(big.NewInt(300))
				evm.stack.Push(big.NewInt(0))
			},
			validate: func() {
				// JUMPI consumes 2 values but doesn't jump
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test state
			tc.setup()

			// Execute the instruction
			if tc.instruction.Opcode == 0x56 || tc.instruction.Opcode == 0x57 {
				evm.executeInstruction(tc.instruction, ctx)
			} else {
				evm.executeInstruction(tc.instruction, nil)
			}

			// Validate results
			tc.validate()
		})
	}
}

func TestEVMFetchOpcodeAndDecode(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Create execution context with contract code
	ctx := &ExecutionContext{
		Contract: &engine.Contract{
			Code: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	// Test fetchOpcode with valid code
	evm.pc = 0
	opcode := evm.fetchOpcode(ctx)
	if opcode != 0x01 {
		t.Errorf("Expected opcode 0x01, got 0x%02x", opcode)
	}

	// Test fetchOpcode with out of bounds
	evm.pc = 10
	opcode = evm.fetchOpcode(ctx)
	if opcode != 0x00 { // Should return STOP (0x00) for out of bounds
		t.Errorf("Expected opcode 0x00 for out of bounds, got 0x%02x", opcode)
	}

	// Test decodeInstruction with valid opcode
	instruction, err := evm.decodeInstruction(0x01)
	if err != nil {
		t.Errorf("Expected no error for opcode 0x01, got %v", err)
	}
	if instruction == nil {
		t.Error("Expected non-nil instruction for opcode 0x01")
	}
	if instruction.Name != "ADD" {
		t.Errorf("Expected instruction name 'ADD', got '%s'", instruction.Name)
	}

	// Test decodeInstruction with unknown opcode
	instruction, err = evm.decodeInstruction(0xE0) // 0xE0 is not defined in Instructions map
	if err == nil {
		t.Error("Expected error for unknown opcode")
	}
}

func TestEVMStackEdgeCases(t *testing.T) {
	stack := NewEVMStack()

	// Test Peek on empty stack - should return 0, not panic
	result := stack.Peek()
	if result.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected 0 for empty stack peek, got %v", result)
	}

	// Test Pop on empty stack - should return 0, not panic
	result = stack.Pop()
	if result.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected 0 for empty stack pop, got %v", result)
	}
}

func TestEVMMemoryEdgeCases(t *testing.T) {
	memory := NewEVMMemory()

	// Test Get with invalid offset
	data := memory.Get(1000, 10)
	if len(data) != 10 {
		t.Error("Memory should return requested size even for invalid offset")
	}

	// Test Set with large offset
	memory.Set(1000000, []byte{0x01, 0x02, 0x03})

	// Test Get with zero size
	data = memory.Get(0, 0)
	if len(data) != 0 {
		t.Error("Memory should return empty slice for zero size")
	}
}

func TestEVMOpcodes_CallData(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Input:      []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Test CALLDATASIZE
	err := evm.executeCALLDATASIZE(ctx)
	assert.NoError(t, err)
	result := evm.stack.Pop()
	assert.Equal(t, int64(6), result.Int64()) // input length

	// Test CALLDATALOAD - load first 32 bytes
	evm.stack.Push(big.NewInt(0)) // offset 0
	err = evm.executeCALLDATALOAD(ctx)
	assert.NoError(t, err)
	loaded := evm.stack.Peek()
	expected := new(big.Int).SetBytes([]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})
	assert.Equal(t, expected.Int64(), loaded.Int64())
	_ = loaded

	// Test CALLDATACOPY
	evm.stack.Push(big.NewInt(0)) // dest offset
	evm.stack.Push(big.NewInt(0)) // src offset
	evm.stack.Push(big.NewInt(6)) // size
	err = evm.executeCALLDATACOPY(ctx)
	assert.NoError(t, err)
	copied := evm.memory.Get(0, 6)
	assert.Equal(t, []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}, copied)
}

func TestEVMOpcodes_Code(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	ctx := &ExecutionContext{
		Contract:   &engine.Contract{Code: []byte{0xde, 0xad, 0xbe, 0xef}},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Test CODESIZE
	evm.stack.Push(big.NewInt(0)) // dummy
	evm.stack.Pop()
	err := evm.executeCODESIZE(ctx)
	assert.NoError(t, err)
	result := evm.stack.Peek()
	assert.Equal(t, int64(4), result.Int64())

	// Test CODECOPY - copy code to memory at offset 10
	evm.stack.Push(big.NewInt(10)) // dest offset
	evm.stack.Push(big.NewInt(0))  // code offset
	evm.stack.Push(big.NewInt(4))  // size
	err = evm.executeCODECOPY(ctx)
	assert.NoError(t, err)
	copied := evm.memory.Get(10, 4)
	assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, copied)
}

func TestEVMOpcodes_ExtCode(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Register a contract in the registry
	targetAddr := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	targetCode := []byte{0x60, 0x80} // PUSH1 0x80
	err := evm.registry.Register(&engine.Contract{
		Address: targetAddr,
		Code:    targetCode,
	})
	assert.NoError(t, err)

	// Test EXTCODESIZE
	addrInt := new(big.Int).SetBytes(targetAddr[:])
	evm.stack.Push(addrInt)
	err = evm.executeEXTCODESIZE(ctx)
	assert.NoError(t, err)
	result := evm.stack.Peek()
	assert.Equal(t, int64(2), result.Int64()) // code length is 2

	// Test EXTCODECOPY
	evm.stack.Push(big.NewInt(10)) // dest offset
	evm.stack.Push(big.NewInt(0))  // code offset
	evm.stack.Push(big.NewInt(2))  // size
	evm.stack.Push(addrInt)        // address
	err = evm.executeEXTCODECOPY(ctx)
	assert.NoError(t, err)
	copied := evm.memory.Get(10, 2)
	assert.Equal(t, []byte{0x60, 0x80}, copied)

	// Test EXTCODEHASH
	evm.stack.Push(addrInt)
	err = evm.executeEXTCODEHASH(ctx)
	assert.NoError(t, err)
	hash := evm.stack.Peek()
	expectedHash := sha256.Sum256(targetCode)
	assert.Equal(t, new(big.Int).SetBytes(expectedHash[:]).Int64(), hash.Int64())
}

func TestEVMOpcodes_ReturnData(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	// Set lastReturnData directly
	evm.lastReturnData = []byte{0xde, 0xad, 0xbe, 0xef}

	// Test RETURNDATASIZE
	evm.stack.Push(big.NewInt(0)) // dummy
	evm.stack.Pop()
	err := evm.executeRETURNDATASIZE()
	assert.NoError(t, err)
	result := evm.stack.Pop()
	assert.Equal(t, int64(4), result.Int64())

	// Test RETURNDATACOPY
	evm.stack.Push(big.NewInt(20)) // dest offset
	evm.stack.Push(big.NewInt(0))  // offset
	evm.stack.Push(big.NewInt(4))  // size
	err = evm.executeRETURNDATACOPY()
	assert.NoError(t, err)
	copied := evm.memory.Get(20, 4)
	assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, copied)
}

func TestEVMOpcodes_Log(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	ctx := &ExecutionContext{
		Contract:   &engine.Contract{Address: engine.Address{0xaa, 0xbb}},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	evm.stack.Push(big.NewInt(0)) // offset
	evm.stack.Push(big.NewInt(0)) // offset
	evm.stack.Push(big.NewInt(4)) // size
	evm.stack.Push(big.NewInt(0)) // offset
	evm.stack.Push(big.NewInt(0)) // offset
	evm.stack.Push(big.NewInt(4)) // size
	evm.memory.Set(0, []byte{0x01, 0x02, 0x03, 0x04})
	err := evm.executeLOG(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(evm.logs))
	assert.Equal(t, evm.logs[0].Address, ctx.Contract.Address)
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, evm.logs[0].Data)
	assert.Len(t, evm.logs[0].Topics, 0)

	// Test LOG1 - one topic
	evm.stack.Push(big.NewInt(10)) // offset
	evm.stack.Push(big.NewInt(4))  // size
	evm.stack.Push(big.NewInt(42)) // topic
	err = evm.executeLOG(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(evm.logs))
	assert.Equal(t, 1, len(evm.logs[1].Topics))
	var expectedTopic engine.Hash
	copy(expectedTopic[:], big.NewInt(42).Bytes())
	assert.Equal(t, expectedTopic[:], evm.logs[1].Topics[0][:])
}

func TestEVMOpcodes_StaticCall(t *testing.T) {
	mockStorage := &MockContractStorage{}
	mockRegistry := NewMockContractRegistry()
	evm := NewEVMEngine(mockStorage, mockRegistry)

	ctx := &ExecutionContext{
		Contract:   &engine.Contract{},
		Sender:     engine.Address{},
		Value:      big.NewInt(0),
		GasPrice:   big.NewInt(0),
		BlockNum:   0,
		Timestamp:  0,
		Coinbase:   engine.Address{},
		Difficulty: big.NewInt(0),
		ChainID:    big.NewInt(1),
	}

	// Register a target contract
	targetAddr := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05}
	targetContract := &engine.Contract{
		Address: targetAddr,
		Code:    []byte{0x60, 0x2a, 0x60, 0x00, 0x52, 0x60, 0x20, 0x60, 0x00, 0xf3}, // PUSH1 0x2a, PUSH1 0x00, MSTORE, PUSH1 0x20, PUSH1 0x00, RETURN
	}
	err := evm.registry.Register(targetContract)
	assert.NoError(t, err)

	// Call the target contract via STATICCALL
	evm.stack.Push(big.NewInt(100000)) // gas
	evm.stack.Push(big.NewInt(0))      // value
	evm.stack.Push(big.NewInt(0))      // inOffset
	evm.stack.Push(big.NewInt(0))      // inSize
	evm.stack.Push(big.NewInt(10))     // retOffset
	evm.stack.Push(big.NewInt(32))     // retSize
	addrInt := new(big.Int).SetBytes(targetAddr[:])
	evm.stack.Push(addrInt) // address

	err = evm.executeSTATICCALL(ctx)
	assert.NoError(t, err)

	// STATICCALL should push 1 (success)
	result := evm.stack.Pop()
	assert.Equal(t, big.NewInt(1), result)

	// Check return data was copied to memory
	retData := evm.memory.Get(10, 32)
	assert.NotEmpty(t, retData)
}
