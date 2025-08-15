package evm

import (
	"math/big"
	"sync"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// ExecutionContext holds the context for EVM execution
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
}

// EVMStack represents the EVM execution stack
type EVMStack struct {
	items []*big.Int
	mu    sync.RWMutex
}

// NewEVMStack creates a new EVM stack
func NewEVMStack() *EVMStack {
	return &EVMStack{
		items: make([]*big.Int, 0),
	}
}

// Push adds a value to the top of the stack
func (s *EVMStack) Push(value *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.items = append(s.items, new(big.Int).Set(value))
}

// Pop removes and returns the top value from the stack
func (s *EVMStack) Pop() *big.Int {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if len(s.items) == 0 {
		return big.NewInt(0)
	}
	
	value := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	
	return value
}

// Peek returns the top value without removing it
func (s *EVMStack) Peek() *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if len(s.items) == 0 {
		return big.NewInt(0)
	}
	
	return new(big.Int).Set(s.items[len(s.items)-1])
}

// Size returns the number of items on the stack
func (s *EVMStack) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return len(s.items)
}

// Reset clears the stack
func (s *EVMStack) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.items = make([]*big.Int, 0)
}

// Clone creates a deep copy of the stack
func (s *EVMStack) Clone() *EVMStack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	clone := &EVMStack{
		items: make([]*big.Int, len(s.items)),
	}
	
	for i, item := range s.items {
		clone.items[i] = new(big.Int).Set(item)
	}
	
	return clone
}

// EVMMemory represents the EVM execution memory
type EVMMemory struct {
	data map[uint64]byte
	mu   sync.RWMutex
}

// NewEVMMemory creates a new EVM memory
func NewEVMMemory() *EVMMemory {
	return &EVMMemory{
		data: make(map[uint64]byte),
	}
}

// Set stores a value in memory at the specified offset
func (m *EVMMemory) Set(offset uint64, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i, b := range data {
		m.data[offset+uint64(i)] = b
	}
}

// Get retrieves data from memory starting at the specified offset
func (m *EVMMemory) Get(offset uint64, size uint64) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]byte, size)
	for i := uint64(0); i < size; i++ {
		if b, exists := m.data[offset+i]; exists {
			result[i] = b
		}
	}
	
	return result
}

// Size returns the current memory size
func (m *EVMMemory) Size() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.data) == 0 {
		return 0
	}
	
	maxOffset := uint64(0)
	for offset := range m.data {
		if offset > maxOffset {
			maxOffset = offset
		}
	}
	
	return maxOffset + 1
}

// Reset clears the memory
func (m *EVMMemory) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.data = make(map[uint64]byte)
}

// Clone creates a deep copy of the memory
func (m *EVMMemory) Clone() *EVMMemory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clone := &EVMMemory{
		data: make(map[uint64]byte),
	}
	
	for offset, value := range m.data {
		clone.data[offset] = value
	}
	
	return clone
}

// Instruction represents an EVM instruction
type Instruction struct {
	Opcode   byte
	Name     string
	GasCost  uint64
	Size     uint64
	Halts    bool
	Pops     int
	Pushes   int
}

// Instructions maps opcodes to instruction definitions
var Instructions = map[byte]*Instruction{
	0x00: {Opcode: 0x00, Name: "STOP", GasCost: 0, Size: 1, Halts: true, Pops: 0, Pushes: 0},
	0x01: {Opcode: 0x01, Name: "ADD", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x02: {Opcode: 0x02, Name: "MUL", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x03: {Opcode: 0x03, Name: "SUB", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x04: {Opcode: 0x04, Name: "DIV", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x05: {Opcode: 0x05, Name: "SDIV", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x06: {Opcode: 0x06, Name: "MOD", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x07: {Opcode: 0x07, Name: "SMOD", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x08: {Opcode: 0x08, Name: "ADDMOD", GasCost: 8, Size: 1, Halts: false, Pops: 3, Pushes: 1},
	0x09: {Opcode: 0x09, Name: "MULMOD", GasCost: 8, Size: 1, Halts: false, Pops: 3, Pushes: 1},
	0x0A: {Opcode: 0x0A, Name: "SIGNEXTEND", GasCost: 5, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	
	// Comparison operations
	0x10: {Opcode: 0x10, Name: "LT", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x11: {Opcode: 0x11, Name: "GT", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x12: {Opcode: 0x12, Name: "SLT", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x13: {Opcode: 0x13, Name: "SGT", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x14: {Opcode: 0x14, Name: "EQ", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x15: {Opcode: 0x15, Name: "ISZERO", GasCost: 3, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x16: {Opcode: 0x16, Name: "AND", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x17: {Opcode: 0x17, Name: "OR", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x18: {Opcode: 0x18, Name: "XOR", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	0x19: {Opcode: 0x19, Name: "NOT", GasCost: 3, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x1A: {Opcode: 0x1A, Name: "BYTE", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	
	// SHA3
	0x20: {Opcode: 0x20, Name: "SHA3", GasCost: 30, Size: 1, Halts: false, Pops: 2, Pushes: 1},
	
	// Environment information
	0x30: {Opcode: 0x30, Name: "ADDRESS", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x31: {Opcode: 0x31, Name: "BALANCE", GasCost: 400, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x32: {Opcode: 0x32, Name: "ORIGIN", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x33: {Opcode: 0x33, Name: "CALLER", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x34: {Opcode: 0x34, Name: "CALLVALUE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x35: {Opcode: 0x35, Name: "CALLDATALOAD", GasCost: 3, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x36: {Opcode: 0x36, Name: "CALLDATASIZE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x37: {Opcode: 0x37, Name: "CALLDATACOPY", GasCost: 3, Size: 1, Halts: false, Pops: 3, Pushes: 0},
	0x38: {Opcode: 0x38, Name: "CODESIZE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x39: {Opcode: 0x39, Name: "CODECOPY", GasCost: 3, Size: 1, Halts: false, Pops: 3, Pushes: 0},
	0x3A: {Opcode: 0x3A, Name: "GASPRICE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x3B: {Opcode: 0x3B, Name: "EXTCODESIZE", GasCost: 700, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x3C: {Opcode: 0x3C, Name: "EXTCODECOPY", GasCost: 700, Size: 1, Halts: false, Pops: 4, Pushes: 0},
	0x3D: {Opcode: 0x3D, Name: "RETURNDATASIZE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x3E: {Opcode: 0x3E, Name: "RETURNDATACOPY", GasCost: 3, Size: 1, Halts: false, Pops: 3, Pushes: 0},
	0x3F: {Opcode: 0x3F, Name: "EXTCODEHASH", GasCost: 400, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	
	// Block information
	0x40: {Opcode: 0x40, Name: "BLOCKHASH", GasCost: 20, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x41: {Opcode: 0x41, Name: "COINBASE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x42: {Opcode: 0x42, Name: "TIMESTAMP", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x43: {Opcode: 0x43, Name: "NUMBER", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x44: {Opcode: 0x44, Name: "DIFFICULTY", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x45: {Opcode: 0x45, Name: "GASLIMIT", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x46: {Opcode: 0x46, Name: "CHAINID", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x47: {Opcode: 0x47, Name: "SELFBALANCE", GasCost: 5, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	
	// Stack operations
	0x50: {Opcode: 0x50, Name: "POP", GasCost: 2, Size: 1, Halts: false, Pops: 1, Pushes: 0},
	0x51: {Opcode: 0x51, Name: "MLOAD", GasCost: 3, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x52: {Opcode: 0x52, Name: "MSTORE", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 0},
	0x53: {Opcode: 0x53, Name: "MSTORE8", GasCost: 3, Size: 1, Halts: false, Pops: 2, Pushes: 0},
	0x54: {Opcode: 0x54, Name: "SLOAD", GasCost: 200, Size: 1, Halts: false, Pops: 1, Pushes: 1},
	0x55: {Opcode: 0x55, Name: "SSTORE", GasCost: 20000, Size: 1, Halts: false, Pops: 2, Pushes: 0},
	0x56: {Opcode: 0x56, Name: "JUMP", GasCost: 8, Size: 1, Halts: false, Pops: 1, Pushes: 0},
	0x57: {Opcode: 0x57, Name: "JUMPI", GasCost: 10, Size: 1, Halts: false, Pops: 2, Pushes: 0},
	0x58: {Opcode: 0x58, Name: "PC", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x59: {Opcode: 0x59, Name: "MSIZE", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x5A: {Opcode: 0x5A, Name: "GAS", GasCost: 2, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x5B: {Opcode: 0x5B, Name: "JUMPDEST", GasCost: 1, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	
	// Push operations
	0x60: {Opcode: 0x60, Name: "PUSH1", GasCost: 3, Size: 2, Halts: false, Pops: 0, Pushes: 1},
	0x61: {Opcode: 0x61, Name: "PUSH2", GasCost: 3, Size: 3, Halts: false, Pops: 0, Pushes: 1},
	0x62: {Opcode: 0x62, Name: "PUSH3", GasCost: 3, Size: 4, Halts: false, Pops: 0, Pushes: 1},
	0x63: {Opcode: 0x63, Name: "PUSH4", GasCost: 3, Size: 5, Halts: false, Pops: 0, Pushes: 1},
	0x64: {Opcode: 0x64, Name: "PUSH5", GasCost: 3, Size: 6, Halts: false, Pops: 0, Pushes: 1},
	0x65: {Opcode: 0x65, Name: "PUSH6", GasCost: 3, Size: 7, Halts: false, Pops: 0, Pushes: 1},
	0x66: {Opcode: 0x66, Name: "PUSH7", GasCost: 3, Size: 8, Halts: false, Pops: 0, Pushes: 1},
	0x67: {Opcode: 0x67, Name: "PUSH8", GasCost: 3, Size: 9, Halts: false, Pops: 0, Pushes: 1},
	0x68: {Opcode: 0x68, Name: "PUSH9", GasCost: 3, Size: 10, Halts: false, Pops: 0, Pushes: 1},
	0x69: {Opcode: 0x69, Name: "PUSH10", GasCost: 3, Size: 11, Halts: false, Pops: 0, Pushes: 1},
	0x6A: {Opcode: 0x6A, Name: "PUSH11", GasCost: 3, Size: 12, Halts: false, Pops: 0, Pushes: 1},
	0x6B: {Opcode: 0x6B, Name: "PUSH12", GasCost: 3, Size: 13, Halts: false, Pops: 0, Pushes: 1},
	0x6C: {Opcode: 0x6C, Name: "PUSH13", GasCost: 3, Size: 14, Halts: false, Pops: 0, Pushes: 1},
	0x6D: {Opcode: 0x6D, Name: "PUSH14", GasCost: 3, Size: 15, Halts: false, Pops: 0, Pushes: 1},
	0x6E: {Opcode: 0x6E, Name: "PUSH15", GasCost: 3, Size: 16, Halts: false, Pops: 0, Pushes: 1},
	0x6F: {Opcode: 0x6F, Name: "PUSH16", GasCost: 3, Size: 17, Halts: false, Pops: 0, Pushes: 1},
	0x70: {Opcode: 0x70, Name: "PUSH17", GasCost: 3, Size: 18, Halts: false, Pops: 0, Pushes: 1},
	0x71: {Opcode: 0x71, Name: "PUSH18", GasCost: 3, Size: 19, Halts: false, Pops: 0, Pushes: 1},
	0x72: {Opcode: 0x72, Name: "PUSH19", GasCost: 3, Size: 20, Halts: false, Pops: 0, Pushes: 1},
	0x73: {Opcode: 0x73, Name: "PUSH20", GasCost: 3, Size: 21, Halts: false, Pops: 0, Pushes: 1},
	0x74: {Opcode: 0x74, Name: "PUSH21", GasCost: 3, Size: 22, Halts: false, Pops: 0, Pushes: 1},
	0x75: {Opcode: 0x75, Name: "PUSH22", GasCost: 3, Size: 23, Halts: false, Pops: 0, Pushes: 1},
	0x76: {Opcode: 0x76, Name: "PUSH23", GasCost: 3, Size: 24, Halts: false, Pops: 0, Pushes: 1},
	0x77: {Opcode: 0x77, Name: "PUSH24", GasCost: 3, Size: 25, Halts: false, Pops: 0, Pushes: 1},
	0x78: {Opcode: 0x78, Name: "PUSH25", GasCost: 3, Size: 26, Halts: false, Pops: 0, Pushes: 1},
	0x79: {Opcode: 0x79, Name: "PUSH26", GasCost: 3, Size: 27, Halts: false, Pops: 0, Pushes: 1},
	0x7A: {Opcode: 0x7A, Name: "PUSH27", GasCost: 3, Size: 28, Halts: false, Pops: 0, Pushes: 1},
	0x7B: {Opcode: 0x7B, Name: "PUSH28", GasCost: 3, Size: 29, Halts: false, Pops: 0, Pushes: 1},
	0x7C: {Opcode: 0x7C, Name: "PUSH29", GasCost: 3, Size: 30, Halts: false, Pops: 0, Pushes: 1},
	0x7D: {Opcode: 0x7D, Name: "PUSH30", GasCost: 3, Size: 31, Halts: false, Pops: 0, Pushes: 1},
	0x7E: {Opcode: 0x7E, Name: "PUSH31", GasCost: 3, Size: 32, Halts: false, Pops: 0, Pushes: 1},
	0x7F: {Opcode: 0x7F, Name: "PUSH32", GasCost: 3, Size: 33, Halts: false, Pops: 0, Pushes: 1},
	
	// Duplication operations
	0x80: {Opcode: 0x80, Name: "DUP1", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x81: {Opcode: 0x81, Name: "DUP2", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x82: {Opcode: 0x82, Name: "DUP3", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x83: {Opcode: 0x83, Name: "DUP4", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x84: {Opcode: 0x84, Name: "DUP5", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x85: {Opcode: 0x85, Name: "DUP6", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x86: {Opcode: 0x86, Name: "DUP7", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x87: {Opcode: 0x87, Name: "DUP8", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x88: {Opcode: 0x88, Name: "DUP9", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x89: {Opcode: 0x89, Name: "DUP10", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8A: {Opcode: 0x8A, Name: "DUP11", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8B: {Opcode: 0x8B, Name: "DUP12", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8C: {Opcode: 0x8C, Name: "DUP13", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8D: {Opcode: 0x8D, Name: "DUP14", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8E: {Opcode: 0x8E, Name: "DUP15", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	0x8F: {Opcode: 0x8F, Name: "DUP16", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 1},
	
	// Exchange operations
	0x90: {Opcode: 0x90, Name: "SWAP1", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x91: {Opcode: 0x91, Name: "SWAP2", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x92: {Opcode: 0x92, Name: "SWAP3", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x93: {Opcode: 0x93, Name: "SWAP4", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x94: {Opcode: 0x94, Name: "SWAP5", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x95: {Opcode: 0x95, Name: "SWAP6", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x96: {Opcode: 0x96, Name: "SWAP7", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x97: {Opcode: 0x97, Name: "SWAP8", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x98: {Opcode: 0x98, Name: "SWAP9", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x99: {Opcode: 0x99, Name: "SWAP10", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9A: {Opcode: 0x9A, Name: "SWAP11", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9B: {Opcode: 0x9B, Name: "SWAP12", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9C: {Opcode: 0x9C, Name: "SWAP13", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9D: {Opcode: 0x9D, Name: "SWAP14", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9E: {Opcode: 0x9E, Name: "SWAP15", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	0x9F: {Opcode: 0x9F, Name: "SWAP16", GasCost: 3, Size: 1, Halts: false, Pops: 0, Pushes: 0},
	
	// Logging operations
	0xA0: {Opcode: 0xA0, Name: "LOG0", GasCost: 375, Size: 1, Halts: false, Pops: 2, Pushes: 0},
	0xA1: {Opcode: 0xA1, Name: "LOG1", GasCost: 750, Size: 1, Halts: false, Pops: 3, Pushes: 0},
	0xA2: {Opcode: 0xA2, Name: "LOG2", GasCost: 1125, Size: 1, Halts: false, Pops: 4, Pushes: 0},
	0xA3: {Opcode: 0xA3, Name: "LOG3", GasCost: 1500, Size: 1, Halts: false, Pops: 5, Pushes: 0},
	0xA4: {Opcode: 0xA4, Name: "LOG4", GasCost: 1875, Size: 1, Halts: false, Pops: 6, Pushes: 0},
	
	// System operations
	0xF0: {Opcode: 0xF0, Name: "CREATE", GasCost: 32000, Size: 1, Halts: false, Pops: 3, Pushes: 1},
	0xF1: {Opcode: 0xF1, Name: "CALL", GasCost: 0, Size: 1, Halts: false, Pops: 7, Pushes: 1},
	0xF2: {Opcode: 0xF2, Name: "CALLCODE", GasCost: 0, Size: 1, Halts: false, Pops: 7, Pushes: 1},
	0xF3: {Opcode: 0xF3, Name: "RETURN", GasCost: 0, Size: 1, Halts: true, Pops: 2, Pushes: 0},
	0xF4: {Opcode: 0xF4, Name: "DELEGATECALL", GasCost: 0, Size: 1, Halts: false, Pops: 6, Pushes: 1},
	0xF5: {Opcode: 0xF5, Name: "CREATE2", GasCost: 32000, Size: 1, Halts: false, Pops: 4, Pushes: 1},
	0xFA: {Opcode: 0xFA, Name: "STATICCALL", GasCost: 0, Size: 1, Halts: false, Pops: 6, Pushes: 1},
	0xFD: {Opcode: 0xFD, Name: "REVERT", GasCost: 0, Size: 1, Halts: true, Pops: 2, Pushes: 0},
	0xFE: {Opcode: 0xFE, Name: "INVALID", GasCost: 0, Size: 1, Halts: true, Pops: 0, Pushes: 0},
	0xFF: {Opcode: 0xFF, Name: "SUICIDE", GasCost: 5000, Size: 1, Halts: true, Pops: 1, Pushes: 0},
}
