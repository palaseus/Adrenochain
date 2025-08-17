package wasm

import (
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// TestWASMEngineDeploy tests contract deployment functionality
func TestWASMEngineDeploy(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test deployment with constructor
	code := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00} // Valid WASM magic
	constructor := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	contract, result, err := wasmEngine.Deploy(code, constructor, 1000000, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during deployment: %v", err)
	}

	if contract == nil {
		t.Fatal("expected non-nil contract")
	}

	if result == nil {
		t.Fatal("expected non-nil execution result")
	}

	if !result.Success {
		t.Errorf("expected successful deployment, got error: %v", result.Error)
	}

	// Test deployment without constructor
	contract2, result2, err := wasmEngine.Deploy(code, nil, 1000000, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during deployment without constructor: %v", err)
	}

	if contract2 == nil {
		t.Fatal("expected non-nil contract")
	}

	if result2 == nil {
		t.Fatal("expected non-nil execution result")
	}

	if !result2.Success {
		t.Errorf("expected successful deployment without constructor, got error: %v", result2.Error)
	}
}

// TestWASMEngineDeployInvalidCode tests deployment with invalid WASM code
func TestWASMEngineDeployInvalidCode(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test deployment with invalid WASM code
	// Note: The current Deploy method doesn't validate WASM code before registration
	// It only validates when executing the constructor
	invalidCode := []byte{0x01, 0x02, 0x03, 0x04} // Invalid WASM magic
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	// Deploy should succeed since code validation happens during execution, not deployment
	contract, result, err := wasmEngine.Deploy(invalidCode, nil, 1000000, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during deployment: %v", err)
	}

	if contract == nil {
		t.Fatal("expected non-nil contract")
	}

	if result == nil {
		t.Fatal("expected non-nil execution result")
	}

	if !result.Success {
		t.Errorf("expected successful deployment, got error: %v", result.Error)
	}

	// Now test that execution with invalid code fails
	executionResult, err := wasmEngine.Execute(contract, []byte{}, 1000000, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during execution: %v", err)
	}

	if executionResult.Success {
		t.Error("expected failed execution with invalid WASM code")
	}

	if executionResult.Error == nil {
		t.Error("expected error in execution result")
	}
}

// TestWASMEngineEstimateGas tests gas estimation functionality
func TestWASMEngineEstimateGas(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	code := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00} // Valid WASM magic
	contract := &engine.Contract{
		Address: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Code:    code,
		Creator: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	input := []byte{0x01, 0x02, 0x03, 0x04}
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	estimatedGas, err := wasmEngine.EstimateGas(contract, input, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during gas estimation: %v", err)
	}

	if estimatedGas == 0 {
		t.Error("expected non-zero estimated gas")
	}

	if estimatedGas < 21000 {
		t.Errorf("expected estimated gas >= 21000, got %d", estimatedGas)
	}
}

// TestWASMEngineCall tests read-only contract calls
func TestWASMEngineCall(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	code := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00} // Valid WASM magic
	contract := &engine.Contract{
		Address: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Code:    code,
		Creator: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	input := []byte{0x01, 0x02, 0x03, 0x04}
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}

	result, err := wasmEngine.Call(contract, input, sender)
	if err != nil {
		t.Fatalf("unexpected error during contract call: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result from contract call")
	}
}

// TestWASMEngineExecute tests contract execution
func TestWASMEngineExecute(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	code := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00} // Valid WASM magic
	contract := &engine.Contract{
		Address: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Code:    code,
		Creator: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	input := []byte{0x01, 0x02, 0x03, 0x04}
	gas := uint64(1000000)
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	result, err := wasmEngine.Execute(contract, input, gas, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during contract execution: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil execution result")
	}

	if !result.Success {
		t.Errorf("expected successful execution, got error: %v", result.Error)
	}
}

// TestWASMEngineExecuteNilContract tests execution with nil contract
func TestWASMEngineExecuteNilContract(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	input := []byte{0x01, 0x02, 0x03, 0x04}
	gas := uint64(1000000)
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	_, err := wasmEngine.Execute(nil, input, gas, sender, value)
	if err == nil {
		t.Fatal("expected error when executing nil contract")
	}
}

// TestWASMEngineExecuteEmptyCode tests execution with empty code
func TestWASMEngineExecuteEmptyCode(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	contract := &engine.Contract{
		Address: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Code:    []byte{}, // Empty code
		Creator: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	input := []byte{0x01, 0x02, 0x03, 0x04}
	gas := uint64(1000000)
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	_, err := wasmEngine.Execute(contract, input, gas, sender, value)
	if err == nil {
		t.Fatal("expected error when executing contract with empty code")
	}
}

// TestWASMEngineExecuteInvalidWASM tests execution with invalid WASM code
func TestWASMEngineExecuteInvalidWASM(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	contract := &engine.Contract{
		Address: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Code:    []byte{0x01, 0x02, 0x03, 0x04}, // Invalid WASM magic
		Creator: engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14},
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	input := []byte{0x01, 0x02, 0x03, 0x04}
	gas := uint64(1000000)
	sender := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	value := big.NewInt(0)

	result, err := wasmEngine.Execute(contract, input, gas, sender, value)
	if err != nil {
		t.Fatalf("unexpected error during contract execution: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil execution result")
	}

	if result.Success {
		t.Error("expected failed execution with invalid WASM code")
	}

	if result.Error == nil {
		t.Error("expected error in execution result")
	}
}

// TestWASMEngineParseWASMModule tests WASM module parsing
func TestWASMEngineParseWASMModule(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test valid WASM module
	validCode := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}

	module, err := wasmEngine.parseWASMModule(validCode)
	if err != nil {
		t.Fatalf("unexpected error parsing valid WASM module: %v", err)
	}

	if module == nil {
		t.Fatal("expected non-nil WASM module")
	}

	if module.Memory == nil {
		t.Error("expected non-nil memory in module")
	}

	if len(module.Types) == 0 {
		t.Error("expected non-empty types in module")
	}

	if len(module.Functions) == 0 {
		t.Error("expected non-empty functions in module")
	}
}

// TestWASMEngineParseWASMModuleInvalid tests WASM module parsing with invalid code
func TestWASMEngineParseWASMModuleInvalid(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test invalid WASM module (too short)
	invalidCode := []byte{0x00, 0x61, 0x73}

	_, err := wasmEngine.parseWASMModule(invalidCode)
	if err == nil {
		t.Fatal("expected error when parsing invalid WASM module")
	}

	if err != ErrInvalidWASM {
		t.Errorf("expected ErrInvalidWASM, got %v", err)
	}

	// Test invalid WASM module (wrong magic number)
	invalidCode2 := []byte{0x01, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}

	_, err = wasmEngine.parseWASMModule(invalidCode2)
	if err == nil {
		t.Fatal("expected error when parsing WASM module with wrong magic number")
	}

	if err != ErrInvalidWASM {
		t.Errorf("expected ErrInvalidWASM, got %v", err)
	}

	// Test invalid WASM module (wrong version)
	invalidCode3 := []byte{0x00, 0x61, 0x73, 0x6D, 0x02, 0x00, 0x00, 0x00}

	_, err = wasmEngine.parseWASMModule(invalidCode3)
	if err == nil {
		t.Fatal("expected error when parsing WASM module with wrong version")
	}

	if err != ErrInvalidWASM {
		t.Errorf("expected ErrInvalidWASM, got %v", err)
	}
}

// TestWASMEngineExecuteFunction tests function execution
func TestWASMEngineExecuteFunction(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Initialize gas meter to avoid panic
	wasmEngine.gasMeter = engine.NewGasMeter(1000000)

	// Create a test module with functions
	module := NewWASMModule()
	functionType := NewWASMFunctionType([]WASMValueType{}, []WASMValueType{})
	module.Types = append(module.Types, functionType)

	function := NewWASMFunction(functionType, []byte{0x01, 0x02, 0x03}, []WASMValueType{}, []byte{0x01, 0x02, 0x03})
	module.Functions = append(module.Functions, function)

	instance := NewWASMInstance(module)

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
		Instance:   instance,
	}

	// Test executing valid function
	err := wasmEngine.executeFunction(ctx, 0, []WASMValue{})
	if err != nil {
		t.Fatalf("unexpected error executing function: %v", err)
	}

	// Test executing non-existent function
	err = wasmEngine.executeFunction(ctx, 999, []WASMValue{})
	if err == nil {
		t.Fatal("expected error when executing non-existent function")
	}

	if err != ErrFunctionNotFound {
		t.Errorf("expected ErrFunctionNotFound, got %v", err)
	}
}

// TestWASMEngineClone tests engine cloning functionality
func TestWASMEngineClone(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Set some context values
	wasmEngine.SetBlockContext(12345, 67890, engine.Address{0x01}, big.NewInt(100))
	wasmEngine.SetGasPrice(big.NewInt(200))
	wasmEngine.SetChainID(big.NewInt(3))

	// Add some data to memory and stack
	wasmEngine.memory.Write(0, []byte{0xAA, 0xBB, 0xCC, 0xDD})
	wasmEngine.stack.Push(NewI32(42))

	// Clone the engine
	clonedEngine := wasmEngine.Clone()

	if clonedEngine == nil {
		t.Fatal("expected non-nil cloned engine")
	}

	// Verify that cloned engine has same values but is independent
	if clonedEngine.blockNum != wasmEngine.blockNum {
		t.Errorf("expected cloned blockNum %d, got %d", wasmEngine.blockNum, clonedEngine.blockNum)
	}

	if clonedEngine.timestamp != wasmEngine.timestamp {
		t.Errorf("expected cloned timestamp %d, got %d", wasmEngine.timestamp, clonedEngine.timestamp)
	}

	if clonedEngine.gasPrice.Cmp(wasmEngine.gasPrice) != 0 {
		t.Errorf("expected cloned gasPrice %v, got %v", wasmEngine.gasPrice, clonedEngine.gasPrice)
	}

	if clonedEngine.chainID.Cmp(wasmEngine.chainID) != 0 {
		t.Errorf("expected cloned chainID %v, got %v", wasmEngine.chainID, clonedEngine.chainID)
	}

	// Verify memory is cloned but independent
	clonedMemory := clonedEngine.GetMemory()
	originalMemory := wasmEngine.GetMemory()

	if clonedMemory.Size() != originalMemory.Size() {
		t.Errorf("expected cloned memory size %d, got %d", originalMemory.Size(), clonedMemory.Size())
	}

	// Modify cloned memory and verify original is unaffected
	clonedMemory.Write(0, []byte{0xFF, 0xFF, 0xFF, 0xFF})

	originalData, _ := originalMemory.Read(0, 4)
	clonedData, _ := clonedMemory.Read(0, 4)

	if originalData[0] == clonedData[0] {
		t.Error("expected original memory to be unaffected by cloned memory changes")
	}
}

// TestWASMEngineContextSetters tests context setter methods
func TestWASMEngineContextSetters(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test SetBlockContext
	blockNum := uint64(12345)
	timestamp := uint64(67890)
	coinbase := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	difficulty := big.NewInt(100)

	wasmEngine.SetBlockContext(blockNum, timestamp, coinbase, difficulty)

	if wasmEngine.blockNum != blockNum {
		t.Errorf("expected blockNum %d, got %d", blockNum, wasmEngine.blockNum)
	}

	if wasmEngine.timestamp != timestamp {
		t.Errorf("expected timestamp %d, got %d", timestamp, wasmEngine.timestamp)
	}

	if wasmEngine.coinbase != coinbase {
		t.Errorf("expected coinbase %v, got %v", coinbase, wasmEngine.coinbase)
	}

	if wasmEngine.difficulty.Cmp(difficulty) != 0 {
		t.Errorf("expected difficulty %v, got %v", difficulty, wasmEngine.difficulty)
	}

	// Test SetGasPrice
	gasPrice := big.NewInt(200)
	wasmEngine.SetGasPrice(gasPrice)

	if wasmEngine.gasPrice.Cmp(gasPrice) != 0 {
		t.Errorf("expected gasPrice %v, got %v", gasPrice, wasmEngine.gasPrice)
	}

	// Test SetChainID
	chainID := big.NewInt(3)
	wasmEngine.SetChainID(chainID)

	if wasmEngine.chainID.Cmp(chainID) != 0 {
		t.Errorf("expected chainID %v, got %v", chainID, wasmEngine.chainID)
	}
}

// TestWASMEngineGetters tests getter methods
func TestWASMEngineGetters(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	// Test GetMemory
	memory := wasmEngine.GetMemory()
	if memory == nil {
		t.Fatal("expected non-nil memory")
	}

	if memory != wasmEngine.memory {
		t.Error("expected GetMemory to return the same memory instance")
	}

	// Test GetStack
	stack := wasmEngine.GetStack()
	if stack == nil {
		t.Fatal("expected non-nil stack")
	}

	if stack != wasmEngine.stack {
		t.Error("expected GetStack to return the same stack instance")
	}

	// Test GetInstance with non-existent address
	address := engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	instance := wasmEngine.GetInstance(address)
	if instance != nil {
		t.Error("expected nil instance for non-existent address")
	}

	// Test GetInstance with existing address
	module := NewWASMModule()
	existingInstance := NewWASMInstance(module)
	wasmEngine.instances[address.String()] = existingInstance

	retrievedInstance := wasmEngine.GetInstance(address)
	if retrievedInstance == nil {
		t.Fatal("expected non-nil instance for existing address")
	}

	if retrievedInstance != existingInstance {
		t.Error("expected GetInstance to return the same instance")
	}
}
