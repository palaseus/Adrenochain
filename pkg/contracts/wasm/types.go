package wasm

import (
	"math/big"
	"sync"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// WASMEngine implements the ContractEngine interface for WebAssembly smart contracts
type WASMEngine struct {
	storage  engine.ContractStorage
	registry engine.ContractRegistry
	gasMeter engine.GasMeter
	mu       sync.RWMutex

	// WASM-specific state
	memory    *WASMMemory
	stack     *WASMStack
	globals   map[uint32]*WASMGlobal
	functions map[uint32]*WASMFunction
	tables    map[uint32]*WASMTable
	instances map[string]*WASMInstance

	// Execution context
	gasPrice   *big.Int
	blockNum   uint64
	timestamp  uint64
	coinbase   engine.Address
	difficulty *big.Int
	chainID    *big.Int
}

// WASMMemory represents the linear memory for WASM execution
type WASMMemory struct {
	data    []byte
	size    uint32
	maxSize uint32
	mu      sync.RWMutex
}

// NewWASMMemory creates a new WASM memory instance
func NewWASMMemory(initialSize, maxSize uint32) *WASMMemory {
	return &WASMMemory{
		data:    make([]byte, initialSize),
		size:    initialSize,
		maxSize: maxSize,
	}
}

// Grow increases memory size by the specified number of pages (64KB each)
func (m *WASMMemory) Grow(pages uint32) (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newSize := m.size + (pages * 65536) // 64KB per page
	if newSize > m.maxSize {
		return 0, ErrMemoryGrowExceedsMaximum
	}

	// Extend the memory slice
	newData := make([]byte, newSize)
	copy(newData, m.data)
	m.data = newData
	m.size = newSize

	return m.size / 65536, nil // Return current page count
}

// Read reads data from memory at the specified offset
func (m *WASMMemory) Read(offset, size uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset+size > m.size {
		return nil, ErrMemoryAccessOutOfBounds
	}

	result := make([]byte, size)
	copy(result, m.data[offset:offset+size])
	return result, nil
}

// Write writes data to memory at the specified offset
func (m *WASMMemory) Write(offset uint32, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offset+uint32(len(data)) > m.size {
		return ErrMemoryAccessOutOfBounds
	}

	copy(m.data[offset:], data)
	return nil
}

// Size returns the current memory size in bytes
func (m *WASMMemory) Size() uint32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

// PageCount returns the current memory size in pages (64KB each)
func (m *WASMMemory) PageCount() uint32 {
	return m.Size() / 65536
}

// Reset clears the memory and resets to initial size
func (m *WASMMemory) Reset(initialSize uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make([]byte, initialSize)
	m.size = initialSize
}

// Clone creates a deep copy of the memory
func (m *WASMMemory) Clone() *WASMMemory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &WASMMemory{
		data:    make([]byte, m.size),
		size:    m.size,
		maxSize: m.maxSize,
	}
	copy(clone.data, m.data)
	return clone
}

// WASMStack represents the value stack for WASM execution
type WASMStack struct {
	values []WASMValue
	mu     sync.RWMutex
}

// NewWASMStack creates a new WASM stack
func NewWASMStack() *WASMStack {
	return &WASMStack{
		values: make([]WASMValue, 0),
	}
}

// Push adds a value to the top of the stack
func (s *WASMStack) Push(value WASMValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, value)
}

// Pop removes and returns the top value from the stack
func (s *WASMStack) Pop() (WASMValue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.values) == 0 {
		return nil, ErrStackUnderflow
	}

	value := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return value, nil
}

// Peek returns the top value without removing it
func (s *WASMStack) Peek() (WASMValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.values) == 0 {
		return nil, ErrStackUnderflow
	}

	return s.values[len(s.values)-1], nil
}

// Size returns the number of values on the stack
func (s *WASMStack) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.values)
}

// Reset clears the stack
func (s *WASMStack) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make([]WASMValue, 0)
}

// Clone creates a deep copy of the stack
func (s *WASMStack) Clone() *WASMStack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := &WASMStack{
		values: make([]WASMValue, len(s.values)),
	}

	for i, value := range s.values {
		clone.values[i] = value.Clone()
	}

	return clone
}

// WASMValue represents a value that can be stored on the WASM stack
type WASMValue interface {
	Type() WASMValueType
	Value() interface{}
	Clone() WASMValue
}

// WASMValueType represents the type of a WASM value
type WASMValueType int

const (
	WASMValueTypeI32 WASMValueType = iota
	WASMValueTypeI64
	WASMValueTypeF32
	WASMValueTypeF64
)

// WASMGlobal represents a global variable in WASM
type WASMGlobal struct {
	Type    WASMValueType
	Mutable bool
	Value   WASMValue
	mu      sync.RWMutex
}

// NewWASMGlobal creates a new WASM global variable
func NewWASMGlobal(valueType WASMValueType, mutable bool, initialValue WASMValue) *WASMGlobal {
	return &WASMGlobal{
		Type:    valueType,
		Mutable: mutable,
		Value:   initialValue,
	}
}

// Get returns the current value of the global
func (g *WASMGlobal) Get() WASMValue {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Value
}

// Set sets the value of the global (only if mutable)
func (g *WASMGlobal) Set(value WASMValue) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.Mutable {
		return ErrGlobalImmutable
	}

	if value.Type() != g.Type {
		return ErrGlobalTypeMismatch
	}

	g.Value = value
	return nil
}

// WASMFunction represents a function in WASM
type WASMFunction struct {
	Type       *WASMFunctionType
	Code       []byte
	LocalTypes []WASMValueType
	Body       []byte
}

// NewWASMFunction creates a new WASM function
func NewWASMFunction(functionType *WASMFunctionType, code []byte, localTypes []WASMValueType, body []byte) *WASMFunction {
	return &WASMFunction{
		Type:       functionType,
		Code:       code,
		LocalTypes: localTypes,
		Body:       body,
	}
}

// WASMFunctionType represents the signature of a WASM function
type WASMFunctionType struct {
	Params  []WASMValueType
	Results []WASMValueType
}

// NewWASMFunctionType creates a new WASM function type
func NewWASMFunctionType(params, results []WASMValueType) *WASMFunctionType {
	return &WASMFunctionType{
		Params:  params,
		Results: results,
	}
}

// WASMTable represents a table in WASM (for function references)
type WASMTable struct {
	ElementType WASMValueType
	Initial     uint32
	Maximum     uint32
	Elements    []WASMValue
	mu          sync.RWMutex
}

// NewWASMTable creates a new WASM table
func NewWASMTable(elementType WASMValueType, initial, maximum uint32) *WASMTable {
	return &WASMTable{
		ElementType: elementType,
		Initial:     initial,
		Maximum:     maximum,
		Elements:    make([]WASMValue, initial),
	}
}

// Get returns the element at the specified index
func (t *WASMTable) Get(index uint32) (WASMValue, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if index >= uint32(len(t.Elements)) {
		return nil, ErrTableIndexOutOfBounds
	}

	return t.Elements[index], nil
}

// Set sets the element at the specified index
func (t *WASMTable) Set(index uint32, value WASMValue) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if index >= uint32(len(t.Elements)) {
		return ErrTableIndexOutOfBounds
	}

	if value.Type() != t.ElementType {
		return ErrTableTypeMismatch
	}

	t.Elements[index] = value
	return nil
}

// Size returns the current size of the table
func (t *WASMTable) Size() uint32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return uint32(len(t.Elements))
}

// Grow increases the table size by the specified number of elements
func (t *WASMTable) Grow(count uint32, value WASMValue) (uint32, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	newSize := uint32(len(t.Elements)) + count
	if newSize > t.Maximum {
		return 0, ErrTableGrowExceedsMaximum
	}

	// Extend the table
	for i := 0; i < int(count); i++ {
		t.Elements = append(t.Elements, value.Clone())
	}

	return uint32(len(t.Elements)), nil
}

// WASMInstance represents a running WASM module instance
type WASMInstance struct {
	Module    *WASMModule
	Memory    *WASMMemory
	Globals   map[uint32]*WASMGlobal
	Functions map[uint32]*WASMFunction
	Tables    map[uint32]*WASMTable
	Exports   map[string]WASMExport
	mu        sync.RWMutex
}

// NewWASMInstance creates a new WASM instance
func NewWASMInstance(module *WASMModule) *WASMInstance {
	return &WASMInstance{
		Module:    module,
		Memory:    module.Memory,
		Globals:   make(map[uint32]*WASMGlobal),
		Functions: make(map[uint32]*WASMFunction),
		Tables:    make(map[uint32]*WASMTable),
		Exports:   make(map[string]WASMExport),
	}
}

// WASMModule represents a compiled WASM module
type WASMModule struct {
	Types     []*WASMFunctionType
	Functions []*WASMFunction
	Tables    []*WASMTable
	Memory    *WASMMemory
	Globals   []*WASMGlobal
	Exports   []WASMExport
	Imports   []WASMImport
	Start     *uint32 // Start function index
}

// NewWASMModule creates a new WASM module
func NewWASMModule() *WASMModule {
	return &WASMModule{
		Types:     make([]*WASMFunctionType, 0),
		Functions: make([]*WASMFunction, 0),
		Tables:    make([]*WASMTable, 0),
		Globals:   make([]*WASMGlobal, 0),
		Exports:   make([]WASMExport, 0),
		Imports:   make([]WASMImport, 0),
	}
}

// WASMExport represents an exported item from a WASM module
type WASMExport struct {
	Name  string
	Kind  WASMExportKind
	Index uint32
}

// WASMExportKind represents the type of export
type WASMExportKind int

const (
	WASMExportKindFunction WASMExportKind = iota
	WASMExportKindTable
	WASMExportKindMemory
	WASMExportKindGlobal
)

// WASMImport represents an imported item in a WASM module
type WASMImport struct {
	Module string
	Name   string
	Kind   WASMExportKind
	Index  uint32
}

// ExecutionContext holds the context for WASM execution
type ExecutionContext struct {
	Contract   *engine.Contract
	Input      []byte
	Sender     engine.Address
	Value      *big.Int
	GasPrice   *big.Int
	BlockNum   uint64
	Timestamp  uint64
	Coinbase   engine.Address
	Difficulty *big.Int
	ChainID    *big.Int
	Instance   *WASMInstance
}
