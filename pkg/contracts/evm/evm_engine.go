package evm

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/palaseus/adrenochain/pkg/contracts/storage"
)

// EVMEngine implements the ContractEngine interface for EVM-compatible smart contracts
type EVMEngine struct {
	storage  storage.ContractStorage
	registry engine.ContractRegistry
	gasMeter engine.GasMeter
	mu       sync.RWMutex

	// EVM-specific state
	stack      *EVMStack
	memory     *EVMMemory
	pc         uint64 // Program counter
	gasPrice   *big.Int
	blockNum   uint64
	timestamp  uint64
	coinbase   engine.Address
	difficulty *big.Int
	gasLimit   uint64
	chainID    *big.Int
}

// NewEVMEngine creates a new EVM execution engine
func NewEVMEngine(storage storage.ContractStorage, registry engine.ContractRegistry) *EVMEngine {
	return &EVMEngine{
		storage:    storage,
		registry:   registry,
		stack:      NewEVMStack(),
		memory:     NewEVMMemory(),
		gasPrice:   big.NewInt(0),
		blockNum:   0,
		timestamp:  0,
		coinbase:   engine.Address{},
		difficulty: big.NewInt(0),
		gasLimit:   0,
		chainID:    big.NewInt(1),
	}
}

// Execute runs a contract with given input and gas limit
func (evm *EVMEngine) Execute(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	evm.mu.RLock()
	defer evm.mu.RUnlock()

	// Initialize execution context
	evm.gasMeter = engine.NewGasMeter(gas)
	evm.stack.Reset()
	evm.memory.Reset()
	evm.pc = 0

	// Validate contract
	if contract == nil {
		return nil, engine.ErrInvalidContract
	}

	if len(contract.Code) == 0 {
		return nil, engine.ErrInvalidContract
	}

	// Create execution context
	ctx := &ExecutionContext{
		Contract:   contract,
		Input:      input,
		Sender:     sender,
		Value:      value,
		GasPrice:   evm.gasPrice,
		BlockNum:   evm.blockNum,
		Timestamp:  evm.timestamp,
		Coinbase:   evm.coinbase,
		Difficulty: evm.difficulty,
		ChainID:    evm.chainID,
	}

	// Execute the contract
	result, err := evm.executeContract(ctx)
	if err != nil {
		return &engine.ExecutionResult{
			Success:      false,
			GasUsed:      evm.gasMeter.GasConsumed(),
			GasRemaining: evm.gasMeter.GasRemaining(),
			Error:        err,
		}, nil
	}

	return result, nil
}

// Deploy creates a new contract with given code and constructor
func (evm *EVMEngine) Deploy(code []byte, constructor []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.Contract, *engine.ExecutionResult, error) {
	evm.mu.Lock()
	defer evm.mu.Unlock()

	// Generate new contract address
	address := evm.registry.GenerateAddress()

	// Create contract instance
	contract := &engine.Contract{
		Address: address,
		Code:    code,
		Creator: sender,
		Balance: big.NewInt(0),
		Nonce:   0,
	}

	// Register contract
	err := evm.registry.Register(contract)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register contract: %w", err)
	}

	// Execute constructor if provided
	var result *engine.ExecutionResult
	if len(constructor) > 0 {
		// Create a copy of the contract for constructor execution
		constructorContract := &engine.Contract{
			Address: contract.Address,
			Code:    constructor,
			Creator: contract.Creator,
			Balance: contract.Balance,
			Nonce:   contract.Nonce,
		}

		// Execute constructor using internal method to avoid mutex deadlock
		result, err = evm.executeContractInternal(constructorContract, nil, gas, sender, value)
		if err != nil {
			// Rollback contract registration on failure
			evm.registry.Remove(address)
			return nil, nil, fmt.Errorf("constructor execution failed: %w", err)
		}

		// Update contract in registry
		err = evm.registry.UpdateContract(contract)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update contract after constructor: %w", err)
		}
	} else {
		// No constructor, create empty result
		result = &engine.ExecutionResult{
			Success:      true,
			GasUsed:      0,
			GasRemaining: gas,
		}
	}

	return contract, result, nil
}

// EstimateGas estimates the gas cost for contract execution
func (evm *EVMEngine) EstimateGas(contract *engine.Contract, input []byte, sender engine.Address, value *big.Int) (uint64, error) {
	// Create a copy of the engine for estimation
	estimationEngine := evm.Clone()

	// Use a high gas limit for estimation
	estimationGas := uint64(10000000) // 10M gas

	// Execute with estimation gas limit
	result, err := estimationEngine.Execute(contract, input, estimationGas, sender, value)
	if err != nil {
		return 0, err
	}

	// Return gas used plus some buffer
	estimatedGas := result.GasUsed
	if estimatedGas < 21000 { // Minimum gas for any transaction
		estimatedGas = 21000
	}

	// Add 20% buffer for safety
	estimatedGas = uint64(float64(estimatedGas) * 1.2)

	return estimatedGas, nil
}

// Call executes a read-only contract call
func (evm *EVMEngine) Call(contract *engine.Contract, input []byte, sender engine.Address) ([]byte, error) {
	// Use a reasonable gas limit for calls
	gasLimit := uint64(1000000) // 1M gas

	result, err := evm.Execute(contract, input, gasLimit, sender, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, result.Error
	}

	return result.ReturnData, nil
}

// executeContractInternal executes contract code without requiring the mutex
// This is used internally to avoid deadlocks during constructor execution
func (evm *EVMEngine) executeContractInternal(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	// Initialize execution context
	ctx := &ExecutionContext{
		Contract:   contract,
		Input:      input,
		Sender:     sender,
		Value:      value,
		GasPrice:   evm.gasPrice,
		BlockNum:   evm.blockNum,
		Timestamp:  evm.timestamp,
		Coinbase:   evm.coinbase,
		Difficulty: evm.difficulty,
		ChainID:    evm.chainID,
	}

	// Initialize execution state
	evm.gasMeter = engine.NewGasMeter(gas)
	evm.stack.Reset()
	evm.memory.Reset()
	evm.pc = 0

	// Load input data into memory if provided
	if len(ctx.Input) > 0 {
		evm.memory.Set(0, ctx.Input)
	}

	// Main execution loop
	for evm.pc < uint64(len(ctx.Contract.Code)) {
		// Check gas
		if evm.gasMeter.IsOutOfGas() {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: 0,
				Error:        engine.ErrOutOfGas,
			}, nil
		}

		// Fetch and decode instruction
		opcode := evm.fetchOpcode(ctx)
		instruction, err := evm.decodeInstruction(opcode)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: evm.gasMeter.GasRemaining(),
				Error:        err,
			}, nil
		}

		// Execute instruction
		halts, err := evm.executeInstruction(instruction, ctx)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: evm.gasMeter.GasRemaining(),
				Error:        err,
			}, nil
		}

		// Check for halt conditions
		if halts {
			break
		}

		// Advance program counter
		evm.pc += instruction.Size
	}

	// Success
	return &engine.ExecutionResult{
		Success:      true,
		ReturnData:   evm.memory.Get(0, 32), // Return first 32 bytes from memory
		GasUsed:      evm.gasMeter.GasConsumed(),
		GasRemaining: evm.gasMeter.GasRemaining(),
		Logs:         []engine.Log{},         // TODO: Implement event logging
		StateChanges: []engine.StateChange{}, // TODO: Implement state change tracking
	}, nil
}

// fetchOpcode fetches the next opcode from contract code
func (evm *EVMEngine) fetchOpcode(ctx *ExecutionContext) byte {
	if evm.pc >= uint64(len(ctx.Contract.Code)) {
		return 0x00 // STOP opcode
	}
	return ctx.Contract.Code[evm.pc]
}

// decodeInstruction decodes an opcode into an instruction
func (evm *EVMEngine) decodeInstruction(opcode byte) (*Instruction, error) {
	instruction, exists := Instructions[opcode]
	if !exists {
		return nil, fmt.Errorf("%w: 0x%02x", engine.ErrInvalidOpcode, opcode)
	}
	return instruction, nil
}

// executeInstruction executes a single EVM instruction
func (evm *EVMEngine) executeInstruction(instruction *Instruction, ctx *ExecutionContext) (bool, error) {
	// Consume gas for instruction
	err := evm.gasMeter.ConsumeGas(instruction.GasCost, instruction.Name)
	if err != nil {
		return false, err
	}

	// Execute instruction logic
	switch instruction.Opcode {
	case 0x00: // STOP
		return true, nil

	case 0x01: // ADD
		err := evm.executeADD()
		return false, err

	case 0x02: // MUL
		err := evm.executeMUL()
		return false, err

	case 0x03: // SUB
		err := evm.executeSUB()
		return false, err

	case 0x04: // DIV
		err := evm.executeDIV()
		return false, err

	case 0x50: // POP
		err := evm.executePOP()
		return false, err

	case 0x51: // MLOAD
		err := evm.executeMLOAD()
		return false, err

	case 0x52: // MSTORE
		err := evm.executeMSTORE()
		return false, err

	case 0x56: // JUMP
		err := evm.executeJUMP(ctx)
		return false, err

	case 0x57: // JUMPI
		err := evm.executeJUMPI(ctx)
		return false, err

	case 0x58: // PC
		err := evm.executePC()
		return false, err

	case 0x59: // MSIZE
		err := evm.executeMSIZE()
		return false, err

	case 0x5A: // GAS
		err := evm.executeGAS()
		return false, err

	case 0xF0: // CREATE
		err := evm.executeCREATE(ctx)
		return false, err

	case 0xF1: // CALL
		err := evm.executeCALL(ctx)
		return false, err

	case 0xFD: // REVERT
		return true, engine.ErrContractReverted

	case 0xFF: // SUICIDE
		err := evm.executeSUICIDE(ctx)
		return true, err

	default:
		// For unimplemented opcodes, just consume gas and continue
		return false, nil
	}
}

// Helper methods for instruction execution
func (evm *EVMEngine) executeADD() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	a := evm.stack.Pop()
	b := evm.stack.Pop()

	result := new(big.Int).Add(a, b)
	evm.stack.Push(result)

	return nil
}

func (evm *EVMEngine) executeMUL() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	a := evm.stack.Pop()
	b := evm.stack.Pop()

	result := new(big.Int).Mul(a, b)
	evm.stack.Push(result)

	return nil
}

func (evm *EVMEngine) executeSUB() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	a := evm.stack.Pop()
	b := evm.stack.Pop()

	result := new(big.Int).Sub(a, b)
	evm.stack.Push(result)

	return nil
}

func (evm *EVMEngine) executeDIV() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	a := evm.stack.Pop()
	b := evm.stack.Pop()

	if b.Sign() == 0 {
		return engine.ErrInvalidInstruction
	}

	result := new(big.Int).Div(a, b)
	evm.stack.Push(result)

	return nil
}

func (evm *EVMEngine) executePOP() error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	evm.stack.Pop()
	return nil
}

func (evm *EVMEngine) executeMLOAD() error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	offset := evm.stack.Pop()
	if offset.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}

	data := evm.memory.Get(offset.Uint64(), 32)
	value := new(big.Int).SetBytes(data)
	evm.stack.Push(value)

	return nil
}

func (evm *EVMEngine) executeMSTORE() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	offset := evm.stack.Pop()
	value := evm.stack.Pop()

	if offset.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}

	data := value.Bytes()
	evm.memory.Set(offset.Uint64(), data)

	return nil
}

func (evm *EVMEngine) executeJUMP(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	dest := evm.stack.Pop()
	if dest.Cmp(big.NewInt(0)) < 0 || dest.Uint64() >= uint64(len(ctx.Contract.Code)) {
		return engine.ErrInvalidJump
	}

	evm.pc = dest.Uint64()
	return nil
}

func (evm *EVMEngine) executeJUMPI(ctx *ExecutionContext) error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	dest := evm.stack.Pop()
	condition := evm.stack.Pop()

	if condition.Sign() != 0 { // Non-zero condition
		if dest.Cmp(big.NewInt(0)) < 0 || dest.Uint64() >= uint64(len(ctx.Contract.Code)) {
			return engine.ErrInvalidJump
		}
		evm.pc = dest.Uint64()
	}

	return nil
}

func (evm *EVMEngine) executePC() error {
	evm.stack.Push(big.NewInt(int64(evm.pc)))
	return nil
}

func (evm *EVMEngine) executeMSIZE() error {
	evm.stack.Push(big.NewInt(int64(evm.memory.Size())))
	return nil
}

func (evm *EVMEngine) executeGAS() error {
	evm.stack.Push(big.NewInt(int64(evm.gasMeter.GasRemaining())))
	return nil
}

func (evm *EVMEngine) executeCREATE(ctx *ExecutionContext) error {
	// TODO: Implement contract creation
	return nil
}

func (evm *EVMEngine) executeCALL(ctx *ExecutionContext) error {
	// TODO: Implement contract calls
	return nil
}

func (evm *EVMEngine) executeSUICIDE(ctx *ExecutionContext) error {
	// TODO: Implement contract self-destruct
	return nil
}

// Clone creates a deep copy of the EVM engine for estimation
func (evm *EVMEngine) Clone() *EVMEngine {
	evm.mu.RLock()
	defer evm.mu.RUnlock()

	clone := &EVMEngine{
		storage:    evm.storage,
		registry:   evm.registry,
		stack:      evm.stack.Clone(),
		memory:     evm.memory.Clone(),
		pc:         evm.pc,
		gasPrice:   new(big.Int).Set(evm.gasPrice),
		blockNum:   evm.blockNum,
		timestamp:  evm.timestamp,
		coinbase:   evm.coinbase,
		difficulty: new(big.Int).Set(evm.difficulty),
		gasLimit:   evm.gasLimit,
		chainID:    new(big.Int).Set(evm.chainID),
	}

	return clone
}

// SetBlockContext sets the block context for EVM execution
func (evm *EVMEngine) SetBlockContext(blockNum uint64, timestamp uint64, coinbase engine.Address, difficulty *big.Int) {
	evm.mu.Lock()
	defer evm.mu.Unlock()

	evm.blockNum = blockNum
	evm.timestamp = timestamp
	evm.coinbase = coinbase
	evm.difficulty = new(big.Int).Set(difficulty)
}

// SetGasPrice sets the gas price for EVM execution
func (evm *EVMEngine) SetGasPrice(gasPrice *big.Int) {
	evm.mu.Lock()
	defer evm.mu.Unlock()

	evm.gasPrice = new(big.Int).Set(gasPrice)
}

// SetChainID sets the chain ID for EVM execution
func (evm *EVMEngine) SetChainID(chainID *big.Int) {
	evm.mu.Lock()
	defer evm.mu.Unlock()
	
	evm.chainID = new(big.Int).Set(chainID)
}

// executeContract executes the actual contract code
func (evm *EVMEngine) executeContract(ctx *ExecutionContext) (*engine.ExecutionResult, error) {
	// Initialize execution state
	evm.gasMeter = engine.NewGasMeter(1000000) // Use default gas limit
	evm.stack.Reset()
	evm.memory.Reset()
	evm.pc = 0

	// Load input data into memory if provided
	if len(ctx.Input) > 0 {
		evm.memory.Set(0, ctx.Input)
	}

	// Main execution loop
	for evm.pc < uint64(len(ctx.Contract.Code)) {
		// Check gas
		if evm.gasMeter.IsOutOfGas() {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: 0,
				Error:        engine.ErrOutOfGas,
			}, nil
		}

		// Fetch and decode instruction
		opcode := evm.fetchOpcode(ctx)
		instruction, err := evm.decodeInstruction(opcode)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: evm.gasMeter.GasRemaining(),
				Error:        err,
			}, nil
		}

		// Execute instruction
		halts, err := evm.executeInstruction(instruction, ctx)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      evm.gasMeter.GasConsumed(),
				GasRemaining: evm.gasMeter.GasRemaining(),
				Error:        err,
			}, nil
		}

		// Check for halt conditions
		if halts {
			break
		}

		// Advance program counter
		evm.pc += instruction.Size
	}

	// Success
	return &engine.ExecutionResult{
		Success:      true,
		ReturnData:   evm.memory.Get(0, 32), // Return first 32 bytes from memory
		GasUsed:      evm.gasMeter.GasConsumed(),
		GasRemaining: evm.gasMeter.GasRemaining(),
		Logs:         []engine.Log{},         // TODO: Implement event logging
		StateChanges: []engine.StateChange{}, // TODO: Implement state change tracking
	}, nil
}
