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
