package wasm

import (
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// Simple test to verify WASM engine creation works
func TestWASMEngineCreation(t *testing.T) {
	mockStorage := &MockWASMStorage{}
	mockRegistry := &MockWASMRegistry{}

	wasmEngine := NewWASMEngine(mockStorage, mockRegistry)

	if wasmEngine == nil {
		t.Fatal("expected non-nil WASM engine")
	}

	if wasmEngine.storage != mockStorage {
		t.Error("expected storage to be set correctly")
	}

	if wasmEngine.registry != mockRegistry {
		t.Error("expected registry to be set correctly")
	}
}

// Test WASM memory functionality
func TestWASMMemory(t *testing.T) {
	memory := NewWASMMemory(65536, 1048576) // 1MB initial, 16MB max

	if memory.Size() != 65536 {
		t.Errorf("expected memory size 65536, got %d", memory.Size())
	}

	// Test memory write and read
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	err := memory.Write(0, testData)
	if err != nil {
		t.Errorf("unexpected error writing to memory: %v", err)
	}

	readData, err := memory.Read(0, 4)
	if err != nil {
		t.Errorf("unexpected error reading from memory: %v", err)
	}

	if len(readData) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(readData))
	}

	for i, b := range testData {
		if readData[i] != b {
			t.Errorf("expected byte %d to be %x, got %x", i, b, readData[i])
		}
	}
}

// Test WASM stack functionality
func TestWASMStack(t *testing.T) {
	stack := NewWASMStack()

	if stack.Size() != 0 {
		t.Errorf("expected empty stack, got size %d", stack.Size())
	}

	// Test push and pop
	value1 := NewI32(42)
	value2 := NewI64(123456789)

	stack.Push(value1)
	stack.Push(value2)

	if stack.Size() != 2 {
		t.Errorf("expected stack size 2, got %d", stack.Size())
	}

	// Test pop (should get values in reverse order)
	popped2, err := stack.Pop()
	if err != nil {
		t.Errorf("unexpected error popping from stack: %v", err)
	}

	if popped2 != value2 {
		t.Error("expected popped value to match pushed value")
	}

	popped1, err := stack.Pop()
	if err != nil {
		t.Errorf("unexpected error popping from stack: %v", err)
	}

	if popped1 != value1 {
		t.Error("expected popped value to match pushed value")
	}

	if stack.Size() != 0 {
		t.Errorf("expected empty stack after popping all values, got size %d", stack.Size())
	}
}

// Test WASM value creation and type conversion
func TestWASMValues(t *testing.T) {
	// Test I32 values
	i32Val := NewI32(42)
	if i32Val.Type() != WASMValueTypeI32 {
		t.Error("expected I32 value type")
	}

	i32, err := AsI32(i32Val)
	if err != nil {
		t.Errorf("unexpected error converting I32: %v", err)
	}
	if i32 != 42 {
		t.Errorf("expected I32 value 42, got %d", i32)
	}

	// Test I64 values
	i64Val := NewI64(123456789)
	if i64Val.Type() != WASMValueTypeI64 {
		t.Error("expected I64 value type")
	}

	i64, err := AsI64(i64Val)
	if err != nil {
		t.Errorf("unexpected error converting I64: %v", err)
	}
	if i64 != 123456789 {
		t.Errorf("expected I64 value 123456789, got %d", i64)
	}

	// Test F32 values
	f32Val := NewF32(3.14)
	if f32Val.Type() != WASMValueTypeF32 {
		t.Error("expected F32 value type")
	}

	f32, err := AsF32(f32Val)
	if err != nil {
		t.Errorf("unexpected error converting F32: %v", err)
	}
	if f32 != 3.14 {
		t.Errorf("expected F32 value 3.14, got %f", f32)
	}

	// Test F64 values
	f64Val := NewF64(2.718281828)
	if f64Val.Type() != WASMValueTypeF64 {
		t.Error("expected F64 value type")
	}

	f64, err := AsF64(f64Val)
	if err != nil {
		t.Errorf("unexpected error converting F64: %v", err)
	}
	if f64 != 2.718281828 {
		t.Errorf("expected F64 value 2.718281828, got %f", f64)
	}
}

// Test WASM memory advanced functionality
func TestWASMMemoryAdvanced(t *testing.T) {
	memory := NewWASMMemory(65536, 1048576) // 1MB initial, 16MB max

	// Test memory growth
	initialSize := memory.Size()
	_, err := memory.Grow(1) // Add 1 page (64KB)
	if err != nil {
		t.Errorf("unexpected error growing memory: %v", err)
	}

	expectedSize := initialSize + 65536
	if memory.Size() != expectedSize {
		t.Errorf("expected memory size %d after growth, got %d", expectedSize, memory.Size())
	}

	// Test page count
	pageCount := memory.PageCount()
	expectedPages := expectedSize / 65536
	if pageCount != expectedPages {
		t.Errorf("expected %d pages, got %d", expectedPages, pageCount)
	}

	// Test memory reset
	memory.Reset(65536)
	if memory.Size() != 65536 {
		t.Errorf("expected memory size 65536 after reset, got %d", memory.Size())
	}

	// Test memory clone
	clonedMemory := memory.Clone()
	if clonedMemory.Size() != memory.Size() {
		t.Errorf("expected cloned memory size %d, got %d", memory.Size(), clonedMemory.Size())
	}

	// Test writing to cloned memory doesn't affect original
	testData := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	err = clonedMemory.Write(0, testData)
	if err != nil {
		t.Errorf("unexpected error writing to cloned memory: %v", err)
	}

	// Original memory should still be empty at that location
	originalData, err := memory.Read(0, 4)
	if err != nil {
		t.Errorf("unexpected error reading from original memory: %v", err)
	}

	// Check that original memory is still zeroed
	for _, b := range originalData {
		if b != 0 {
			t.Error("expected original memory to be unaffected by clone writes")
			break
		}
	}
}

// Test WASM stack advanced functionality
func TestWASMStackAdvanced(t *testing.T) {
	stack := NewWASMStack()

	// Test peek functionality
	value1 := NewI32(42)
	value2 := NewI64(123456789)

	stack.Push(value1)
	stack.Push(value2)

	// Peek should return top value without removing it
	peekedValue, err := stack.Peek()
	if err != nil {
		t.Errorf("unexpected error peeking from stack: %v", err)
	}

	if peekedValue != value2 {
		t.Error("expected peeked value to match top value")
	}

	// Stack size should remain the same after peek
	if stack.Size() != 2 {
		t.Errorf("expected stack size 2 after peek, got %d", stack.Size())
	}

	// Test stack reset
	stack.Reset()
	if stack.Size() != 0 {
		t.Errorf("expected empty stack after reset, got size %d", stack.Size())
	}

	// Test stack clone
	stack.Push(value1)
	stack.Push(value2)

	clonedStack := stack.Clone()
	if clonedStack.Size() != stack.Size() {
		t.Errorf("expected cloned stack size %d, got %d", stack.Size(), clonedStack.Size())
	}

	// Test that cloned stack has same values
	clonedValue2, err := clonedStack.Pop()
	if err != nil {
		t.Errorf("unexpected error popping from cloned stack: %v", err)
	}

	if clonedValue2 == nil {
		t.Error("expected non-nil value from cloned stack")
	}

	// Test that cloned stack has same values by checking type and value
	if clonedValue2.Type() != value2.Type() {
		t.Error("expected cloned stack value to have same type")
	}
}

// Test WASM global variables
func TestWASMGlobal(t *testing.T) {
	// Test I32 global
	i32Global := NewWASMGlobal(WASMValueTypeI32, true, NewI32(100))
	if i32Global.Get().Type() != WASMValueTypeI32 {
		t.Error("expected I32 global type")
	}

	// Test setting global value
	newValue := NewI32(200)
	i32Global.Set(newValue)

	currentValue := i32Global.Get()
	if currentValue != newValue {
		t.Error("expected global value to be updated")
	}

	// Test I64 global
	i64Global := NewWASMGlobal(WASMValueTypeI64, false, NewI64(999999))
	if i64Global.Get().Type() != WASMValueTypeI64 {
		t.Error("expected I64 global type")
	}

	// Test F32 global
	f32Global := NewWASMGlobal(WASMValueTypeF32, true, NewF32(3.14159))
	if f32Global.Get().Type() != WASMValueTypeF32 {
		t.Error("expected F32 global type")
	}

	// Test F64 global
	f64Global := NewWASMGlobal(WASMValueTypeF64, false, NewF64(2.718281828))
	if f64Global.Get().Type() != WASMValueTypeF64 {
		t.Error("expected F64 global type")
	}
}

// Test WASM function types
func TestWASMFunctionTypes(t *testing.T) {
	// Test function type creation
	paramTypes := []WASMValueType{WASMValueTypeI32, WASMValueTypeI64}
	resultTypes := []WASMValueType{WASMValueTypeF32}

	functionType := NewWASMFunctionType(paramTypes, resultTypes)

	if len(functionType.Params) != 2 {
		t.Errorf("expected 2 parameter types, got %d", len(functionType.Params))
	}

	if len(functionType.Results) != 1 {
		t.Errorf("expected 1 result type, got %d", len(functionType.Results))
	}

	if functionType.Params[0] != WASMValueTypeI32 {
		t.Error("expected first parameter type to be I32")
	}

	if functionType.Params[1] != WASMValueTypeI64 {
		t.Error("expected second parameter type to be I64")
	}

	if functionType.Results[0] != WASMValueTypeF32 {
		t.Error("expected result type to be F32")
	}
}

// Test WASM functions
func TestWASMFunction(t *testing.T) {
	// Test function creation
	paramTypes := []WASMValueType{WASMValueTypeI32}
	resultTypes := []WASMValueType{WASMValueTypeI32}

	functionType := NewWASMFunctionType(paramTypes, resultTypes)

	// Create a simple function with code and body
	code := []byte{0x01, 0x02, 0x03, 0x04}
	localTypes := []WASMValueType{WASMValueTypeI32}
	body := []byte{0x05, 0x06, 0x07, 0x08}

	function := NewWASMFunction(functionType, code, localTypes, body)

	if function.Type != functionType {
		t.Error("expected function type to match")
	}

	if len(function.Code) != 4 {
		t.Errorf("expected code length 4, got %d", len(function.Code))
	}

	if len(function.LocalTypes) != 1 {
		t.Errorf("expected 1 local type, got %d", len(function.LocalTypes))
	}

	if len(function.Body) != 4 {
		t.Errorf("expected body length 4, got %d", len(function.Body))
	}
}

// Test WASM tables
func TestWASMTable(t *testing.T) {
	// Test table creation
	table := NewWASMTable(WASMValueTypeI32, 10, 100)

	if table.Size() != 10 {
		t.Errorf("expected table size 10, got %d", table.Size())
	}

	// Test setting and getting values
	testValue := NewI32(12345)
	err := table.Set(5, testValue)
	if err != nil {
		t.Errorf("unexpected error setting table value: %v", err)
	}

	retrievedValue, err := table.Get(5)
	if err != nil {
		t.Errorf("unexpected error getting table value: %v", err)
	}

	if retrievedValue != testValue {
		t.Error("expected retrieved value to match set value")
	}

	// Test table growth
	_, err = table.Grow(5, NewI32(0))
	if err != nil {
		t.Errorf("unexpected error growing table: %v", err)
	}

	if table.Size() != 15 {
		t.Errorf("expected table size 15 after growth, got %d", table.Size())
	}

	// Test setting value in grown area
	newValue := NewI32(54321)
	err = table.Set(12, newValue)
	if err != nil {
		t.Errorf("unexpected error setting value in grown area: %v", err)
	}

	retrievedNewValue, err := table.Get(12)
	if err != nil {
		t.Errorf("unexpected error getting value from grown area: %v", err)
	}

	if retrievedNewValue != newValue {
		t.Error("expected retrieved value from grown area to match set value")
	}
}

// Test WASM value type conversions and edge cases
func TestWASMValueConversions(t *testing.T) {
	// Test type conversion errors
	i32Val := NewI32(42)

	// Try to convert I32 to I64 (should fail - different types)
	_, err := AsI64(i32Val)
	if err == nil {
		t.Error("expected error when converting I32 to I64")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Try to convert I32 to F32 (should fail - different types)
	_, err = AsF32(i32Val)
	if err == nil {
		t.Error("expected error when converting I32 to F32")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Try to convert I32 to F64 (should fail - different types)
	_, err = AsF64(i32Val)
	if err == nil {
		t.Error("expected error when converting I32 to F64")
	}
	if err != ErrGlobalTypeMismatch {
		t.Errorf("expected ErrGlobalTypeMismatch, got %v", err)
	}

	// Test value cloning
	clonedValue := i32Val.Clone()
	if clonedValue == nil {
		t.Error("expected non-nil cloned value")
	}
	
	// Test that cloned value has same type and value
	if clonedValue.Type() != i32Val.Type() {
		t.Error("expected cloned value to have same type")
	}
	
	clonedIntValue := clonedValue.(*WASMI32Value).Int32()
	originalIntValue := i32Val.(*WASMI32Value).Int32()
	if clonedIntValue != originalIntValue {
		t.Errorf("expected cloned value %d to equal original value %d", clonedIntValue, originalIntValue)
	}

	// Test that cloned value is independent
	originalValue, _ := AsI32(i32Val)
	clonedValueInt, _ := AsI32(clonedValue)

	if originalValue != clonedValueInt {
		t.Error("expected cloned value to have same integer value")
	}

	// Test that cloned value is independent
	clonedValue.(*WASMI32Value).value = 100
	if i32Val.(*WASMI32Value).Int32() == 100 {
		t.Error("expected original value to be unaffected by cloned value changes")
	}
}

// Test WASM memory edge cases
func TestWASMMemoryEdgeCases(t *testing.T) {
	memory := NewWASMMemory(65536, 1048576) // 1MB initial, 16MB max

	// Test writing at memory boundary
	testData := []byte{0xFF, 0xFE, 0xFD, 0xFC}
	err := memory.Write(65532, testData) // Write 4 bytes starting at 65532
	if err != nil {
		t.Errorf("unexpected error writing at memory boundary: %v", err)
	}

	// Test reading at memory boundary
	readData, err := memory.Read(65532, 4)
	if err != nil {
		t.Errorf("unexpected error reading at memory boundary: %v", err)
	}

	if len(readData) != 4 {
		t.Errorf("expected 4 bytes, got %d", len(readData))
	}

	for i, b := range testData {
		if readData[i] != b {
			t.Errorf("expected byte %d to be %x, got %x", i, b, readData[i])
		}
	}

	// Test writing beyond memory boundary (should fail)
	err = memory.Write(65536, testData)
	if err == nil {
		t.Error("expected error when writing beyond memory boundary")
	}

	// Test reading beyond memory boundary (should fail)
	_, err = memory.Read(65536, 4)
	if err == nil {
		t.Error("expected error when reading beyond memory boundary")
	}

	// Test writing zero bytes (should succeed)
	err = memory.Write(0, []byte{})
	if err != nil {
		t.Errorf("unexpected error writing zero bytes: %v", err)
	}

	// Test reading zero bytes (should succeed)
	emptyData, err := memory.Read(0, 0)
	if err != nil {
		t.Errorf("unexpected error reading zero bytes: %v", err)
	}

	if len(emptyData) != 0 {
		t.Errorf("expected 0 bytes, got %d", len(emptyData))
	}
}

// Mock implementations for testing
type MockWASMStorage struct{}

func (m *MockWASMStorage) Get(address engine.Address, key engine.Hash) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockWASMStorage) Set(address engine.Address, key engine.Hash, value []byte) error {
	return nil
}

func (m *MockWASMStorage) Delete(address engine.Address, key engine.Hash) error {
	return nil
}

func (m *MockWASMStorage) GetStorageRoot(address engine.Address) (engine.Hash, error) {
	return engine.Hash{}, nil
}

func (m *MockWASMStorage) Commit() error {
	return nil
}

func (m *MockWASMStorage) Rollback() error {
	return nil
}

func (m *MockWASMStorage) HasKey(address engine.Address, key engine.Hash) bool {
	return false
}

func (m *MockWASMStorage) GetContractStorage(address engine.Address) (map[engine.Hash][]byte, error) {
	return make(map[engine.Hash][]byte), nil
}

func (m *MockWASMStorage) GetStorageSize(address engine.Address) (int, error) {
	return 0, nil
}

func (m *MockWASMStorage) ClearContractStorage(address engine.Address) error {
	return nil
}

func (m *MockWASMStorage) GetStorageProof(address engine.Address, key engine.Hash) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockWASMStorage) VerifyStorageProof(root engine.Hash, key engine.Hash, value []byte, proof []byte) bool {
	return true
}

type MockWASMRegistry struct{}

func (m *MockWASMRegistry) Register(contract *engine.Contract) error {
	return nil
}

func (m *MockWASMRegistry) Get(address engine.Address) (*engine.Contract, error) {
	return nil, engine.ErrContractNotFound
}

func (m *MockWASMRegistry) Exists(address engine.Address) bool {
	return false
}

func (m *MockWASMRegistry) Remove(address engine.Address) error {
	return nil
}

func (m *MockWASMRegistry) List() []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockWASMRegistry) GetContractCount() int {
	return 0
}

func (m *MockWASMRegistry) GetContractByCodeHash(codeHash engine.Hash) []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockWASMRegistry) GetContractsByCreator(creator engine.Address) []*engine.Contract {
	return []*engine.Contract{}
}

func (m *MockWASMRegistry) UpdateContract(contract *engine.Contract) error {
	return nil
}

func (m *MockWASMRegistry) GetContractAddresses() []engine.Address {
	return []engine.Address{}
}

func (m *MockWASMRegistry) HasContracts() bool {
	return false
}

func (m *MockWASMRegistry) Clear() {}

func (m *MockWASMRegistry) GenerateAddress() engine.Address {
	return engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
}

func (m *MockWASMRegistry) GetContractStats() engine.ContractStats {
	return engine.ContractStats{}
}
