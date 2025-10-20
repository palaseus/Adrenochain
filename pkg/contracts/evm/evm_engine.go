package evm

import (
	"bytes"
	"crypto/sha256"
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

	// SECURITY FIX: Add security limits
	maxInputSize    int
	maxContractSize int
	maxGasLimit     uint64
	minGasLimit     uint64
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

		// SECURITY FIX: Initialize security limits
		maxInputSize:    1024 * 1024, // 1MB max input
		maxContractSize: 24 * 1024,   // 24KB max contract (Ethereum limit)
		maxGasLimit:     10000000,    // 10M gas max
		minGasLimit:     21000,       // 21K gas min (basic transaction)
	}
}

// SECURITY FIX: validateGasLimit performs comprehensive gas validation
func (evm *EVMEngine) validateGasLimit(gas uint64, contract *engine.Contract) error {
	// Check minimum gas limit
	if gas < evm.minGasLimit {
		return fmt.Errorf("gas limit %d below minimum %d", gas, evm.minGasLimit)
	}

	// Check maximum gas limit
	if gas > evm.maxGasLimit {
		return fmt.Errorf("gas limit %d exceeds maximum %d", gas, evm.maxGasLimit)
	}

	// Estimate gas based on contract complexity
	estimatedGas := evm.estimateGasForContract(contract)
	if gas < estimatedGas {
		return fmt.Errorf("gas limit %d insufficient for contract complexity (estimated: %d)", gas, estimatedGas)
	}

	// Check for gas limit manipulation (prevent extremely high gas limits)
	if gas > evm.maxGasLimit/2 {
		// For high gas limits, require additional validation
		if contract == nil || len(contract.Code) == 0 {
			return fmt.Errorf("high gas limit %d requires valid contract", gas)
		}
	}

	return nil
}

// SECURITY FIX: estimateGasForContract estimates gas based on contract complexity
func (evm *EVMEngine) estimateGasForContract(contract *engine.Contract) uint64 {
	if contract == nil || len(contract.Code) == 0 {
		return evm.minGasLimit
	}

	// Base gas for contract execution
	baseGas := uint64(21000)

	// Add gas based on contract size
	sizeGas := uint64(len(contract.Code)) * 200 // 200 gas per byte

	// Add gas for complexity (count of different opcodes)
	complexityGas := evm.calculateComplexityGas(contract.Code)

	totalGas := baseGas + sizeGas + complexityGas

	// Ensure minimum gas
	if totalGas < evm.minGasLimit {
		totalGas = evm.minGasLimit
	}

	return totalGas
}

// SECURITY FIX: calculateComplexityGas calculates gas based on opcode complexity
func (evm *EVMEngine) calculateComplexityGas(code []byte) uint64 {
	complexity := uint64(0)

	// Count expensive operations
	for _, opcode := range code {
		switch opcode {
		case 0xF0: // CREATE
			complexity += 32000
		case 0xF1: // CALL
			complexity += 700
		case 0xF2: // CALLCODE
			complexity += 700
		case 0xF4: // DELEGATECALL
			complexity += 700
		case 0xF5: // CREATE2
			complexity += 32000
		case 0x56: // JUMP
			complexity += 8
		case 0x57: // JUMPI
			complexity += 10
		case 0x52: // MSTORE
			complexity += 3
		case 0x53: // MSTORE8
			complexity += 3
		case 0x54: // SLOAD
			complexity += 200
		case 0x55: // SSTORE
			complexity += 20000
		default:
			complexity += 1
		}
	}

	return complexity
}

// Execute runs a contract with given input and gas limit
func (evm *EVMEngine) Execute(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	evm.mu.RLock()
	defer evm.mu.RUnlock()

	// SECURITY FIX: Comprehensive gas validation
	if err := evm.validateGasLimit(gas, contract); err != nil {
		return nil, fmt.Errorf("gas validation failed: %w", err)
	}

	// SECURITY FIX: Validate input size to prevent DoS
	if len(input) > evm.maxInputSize {
		return nil, fmt.Errorf("input size %d exceeds maximum %d", len(input), evm.maxInputSize)
	}

	// Initialize execution context
	evm.gasMeter = engine.NewGasMeter(gas)
	evm.gasLimit = gas // SECURITY FIX: Store validated gas limit
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

	// SECURITY FIX: Validate contract size
	if len(contract.Code) > evm.maxContractSize {
		return nil, fmt.Errorf("contract size %d exceeds maximum %d", len(contract.Code), evm.maxContractSize)
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
			if err := evm.registry.Remove(address); err != nil {
				// Log error but continue with cleanup
				fmt.Printf("Failed to remove contract from registry: %v\n", err)
			}
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
		Logs:         []engine.Log{},
		StateChanges: []engine.StateChange{},
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
		return fmt.Errorf("invalid jump destination: destination %d is out of bounds (code length: %d)", dest.Uint64(), len(ctx.Contract.Code))
	}

	// Check if destination is a valid JUMPDEST
	if ctx.Contract.Code[dest.Uint64()] != 0x5B {
		return fmt.Errorf("invalid jump destination: expected JUMPDEST at position %d, got 0x%02x", dest.Uint64(), ctx.Contract.Code[dest.Uint64()])
	}

	evm.pc = dest.Uint64()
	return nil
}

func (evm *EVMEngine) executeJUMPI(ctx *ExecutionContext) error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	condition := evm.stack.Pop()
	dest := evm.stack.Pop()

	if condition.Sign() != 0 { // Non-zero condition
		if dest.Cmp(big.NewInt(0)) < 0 || dest.Uint64() >= uint64(len(ctx.Contract.Code)) {
			return fmt.Errorf("invalid jump destination: destination %d is out of bounds (code length: %d)", dest.Uint64(), len(ctx.Contract.Code))
		}

		// Check if destination is a valid JUMPDEST
		if ctx.Contract.Code[dest.Uint64()] != 0x5B {
			return fmt.Errorf("invalid jump destination: expected JUMPDEST at position %d, got 0x%02x", dest.Uint64(), ctx.Contract.Code[dest.Uint64()])
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
	// CREATE opcode: CREATE(value, offset, size) -> address
	if evm.stack.Size() < 3 {
		return engine.ErrStackUnderflow
	}

	// Pop parameters from stack
	value := evm.stack.Pop()
	offset := evm.stack.Pop()
	size := evm.stack.Pop()

	// Validate parameters
	if offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Check if offset + size exceeds memory bounds
	if new(big.Int).Add(offset, size).Cmp(big.NewInt(int64(evm.memory.Size()))) > 0 {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Read init code from memory
	initCode := evm.memory.Get(offset.Uint64(), size.Uint64())
	if len(initCode) == 0 {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Generate contract address
	contractAddr := evm.generateContractAddress(ctx.Sender)

	// Deploy contract
	err := evm.deployContract(contractAddr, initCode, value)
	if err != nil {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Push contract address on success
	evm.stack.Push(big.NewInt(0).SetBytes(contractAddr[:]))
	return nil
}

func (evm *EVMEngine) executeCALL(ctx *ExecutionContext) error {
	// CALL opcode: CALL(gas, address, value, argsOffset, argsSize, retOffset, retSize) -> success
	if evm.stack.Size() < 7 {
		return engine.ErrStackUnderflow
	}

	// Pop parameters from stack
	gas := evm.stack.Pop()
	address := evm.stack.Pop()
	value := evm.stack.Pop()
	argsOffset := evm.stack.Pop()
	argsSize := evm.stack.Pop()
	retOffset := evm.stack.Pop()
	retSize := evm.stack.Pop()

	// Validate parameters
	if gas.Cmp(big.NewInt(0)) < 0 || argsOffset.Cmp(big.NewInt(0)) < 0 ||
		argsSize.Cmp(big.NewInt(0)) < 0 || retOffset.Cmp(big.NewInt(0)) < 0 ||
		retSize.Cmp(big.NewInt(0)) < 0 {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Check if args offset + size exceeds memory bounds
	if new(big.Int).Add(argsOffset, argsSize).Cmp(big.NewInt(int64(evm.memory.Size()))) > 0 {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Read call data from memory
	callData := evm.memory.Get(argsOffset.Uint64(), argsSize.Uint64())

	// Convert address from big.Int to Address
	var targetAddr engine.Address
	addrBytes := address.Bytes()
	if len(addrBytes) > 20 {
		addrBytes = addrBytes[len(addrBytes)-20:] // Take last 20 bytes
	}
	copy(targetAddr[:], addrBytes)

	// Check if target address exists (has code or is EOA)
	hasCode, err := evm.hasContractCode(targetAddr)
	if err != nil {
		evm.stack.Push(big.NewInt(0)) // Push 0 on failure
		return nil
	}

	// Execute call
	var result []byte
	if hasCode {
		// Contract call - execute the contract
		callCtx := &ExecutionContext{
			Contract:   &engine.Contract{Address: targetAddr, Code: []byte{}},
			Input:      callData,
			Sender:     ctx.Sender,
			Value:      value,
			GasPrice:   ctx.GasPrice,
			BlockNum:   ctx.BlockNum,
			Timestamp:  ctx.Timestamp,
			Coinbase:   ctx.Coinbase,
			Difficulty: ctx.Difficulty,
			ChainID:    ctx.ChainID,
		}
		result, err = evm.executeCall(callCtx)
		if err != nil {
			evm.stack.Push(big.NewInt(0)) // Push 0 on failure
			return nil
		}
	} else {
		// EOA call - just transfer value
		err = evm.transferBalance(ctx.Sender, targetAddr, value)
		if err != nil {
			evm.stack.Push(big.NewInt(0)) // Push 0 on failure
			return nil
		}
		result = []byte{} // EOA calls return empty result
	}

	// Write return data to memory
	if len(result) > 0 && retSize.Cmp(big.NewInt(0)) > 0 {
		// Ensure we don't write beyond memory bounds
		writeSize := int(retSize.Uint64())
		if writeSize > len(result) {
			writeSize = len(result)
		}
		evm.memory.Set(retOffset.Uint64(), result[:writeSize])
	}

	// Push 1 for success
	evm.stack.Push(big.NewInt(1))
	return nil
}

func (evm *EVMEngine) executeSUICIDE(ctx *ExecutionContext) error {
	// SUICIDE opcode: SUICIDE(address) - transfers balance and destroys contract
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	// Pop beneficiary address from stack
	beneficiary := evm.stack.Pop()

	// Convert address from big.Int to Address
	var beneficiaryAddr engine.Address
	addrBytes := beneficiary.Bytes()
	if len(addrBytes) > 20 {
		addrBytes = addrBytes[len(addrBytes)-20:] // Take last 20 bytes
	}
	copy(beneficiaryAddr[:], addrBytes)

	// Get contract balance
	var contractBalance *big.Int
	if ctx.Contract != nil {
		contractBalance = evm.getContractBalance(ctx.Contract.Address)
	} else {
		contractBalance = big.NewInt(0)
	}

	// Transfer balance to beneficiary if there's any
	if contractBalance.Cmp(big.NewInt(0)) > 0 && ctx.Contract != nil {
		err := evm.transferBalance(ctx.Contract.Address, beneficiaryAddr, contractBalance)
		if err != nil {
			// Even if transfer fails, we still mark for destruction
			// This matches EVM behavior
		}
	}

	// Mark contract for destruction
	if ctx.Contract != nil {
		evm.markContractForDestruction(ctx.Contract.Address)
		// Clear contract storage
		evm.clearContractStorage(ctx.Contract.Address)
	}

	// SUICIDE halts execution (handled by caller)
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

		// Copy security limits
		maxInputSize:    evm.maxInputSize,
		maxContractSize: evm.maxContractSize,
		maxGasLimit:     evm.maxGasLimit,
		minGasLimit:     evm.minGasLimit,
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

// Helper methods for instruction implementations

// generateContractAddress generates a new contract address using proper CREATE address generation
func (evm *EVMEngine) generateContractAddress(sender engine.Address) engine.Address {
	// Get sender nonce from storage
	nonce, err := evm.getSenderNonce(sender)
	if err != nil {
		// If we can't get nonce, use 0
		nonce = 0
	}

	// CREATE address = keccak256(rlp([sender, nonce]))
	// For simplicity, we'll use a deterministic approach based on sender and nonce
	// In a full implementation, this would use RLP encoding and keccak256
	data := append(sender[:], make([]byte, 8)...)
	// Encode nonce as big-endian 8 bytes
	for i := 0; i < 8; i++ {
		data[20+i] = byte(nonce >> (8 * (7 - i)))
	}

	hash := sha256.Sum256(data)
	addr := engine.Address{}
	copy(addr[:], hash[:20])
	return addr
}

// deployContract deploys a new contract
func (evm *EVMEngine) deployContract(address engine.Address, code []byte, value *big.Int) error {
	// Store contract code in storage using a special key
	codeKey := engine.Hash{} // Use zero hash for contract code
	err := evm.storage.Set(address, codeKey, code)
	if err != nil {
		return fmt.Errorf("failed to store contract code: %w", err)
	}

	// Transfer initial value if any
	if value.Cmp(big.NewInt(0)) > 0 {
		err = evm.transferBalance(engine.Address{}, address, value)
		if err != nil {
			return fmt.Errorf("failed to transfer initial value: %w", err)
		}
	}

	return nil
}

// executeCall executes a contract call
func (evm *EVMEngine) executeCall(ctx *ExecutionContext) ([]byte, error) {
	// Get contract code from storage using special key
	codeKey := engine.Hash{} // Use zero hash for contract code
	code, err := evm.storage.Get(ctx.Contract.Address, codeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract code: %w", err)
	}

	if len(code) == 0 {
		// EOA call - just return empty result
		return []byte{}, nil
	}

	// Create new execution context for the call
	callCtx := &ExecutionContext{
		Contract:   &engine.Contract{Address: ctx.Contract.Address, Code: code},
		Input:      ctx.Input,
		Sender:     ctx.Sender,
		Value:      ctx.Value,
		GasPrice:   ctx.GasPrice,
		BlockNum:   ctx.BlockNum,
		Timestamp:  ctx.Timestamp,
		Coinbase:   ctx.Coinbase,
		Difficulty: ctx.Difficulty,
		ChainID:    ctx.ChainID,
	}

	// Execute the contract
	result, err := evm.executeContract(callCtx)
	if err != nil {
		return nil, fmt.Errorf("contract execution failed: %w", err)
	}

	return result.ReturnData, nil
}

// getContractBalance gets the balance of a contract
func (evm *EVMEngine) getContractBalance(address engine.Address) *big.Int {
	// Query balance from storage using special key
	balanceKey := engine.Hash{0x01} // Use special key for balance
	balanceBytes, err := evm.storage.Get(address, balanceKey)
	if err != nil {
		// If we can't get balance, return 0
		return big.NewInt(0)
	}

	// Convert bytes to big.Int
	balance := new(big.Int)
	balance.SetBytes(balanceBytes)
	return balance
}

// transferBalance transfers balance between addresses
func (evm *EVMEngine) transferBalance(from, to engine.Address, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil // No transfer needed
	}

	// Get sender balance
	fromBalance := evm.getContractBalance(from)
	if fromBalance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance: have %s, need %s", fromBalance.String(), amount.String())
	}

	// Get recipient balance
	toBalance := evm.getContractBalance(to)

	// Update balances
	newFromBalance := new(big.Int).Sub(fromBalance, amount)
	newToBalance := new(big.Int).Add(toBalance, amount)

	// Store updated balances using special keys
	fromBalanceKey := engine.Hash{0x01} // Use special key for balance
	err := evm.storage.Set(from, fromBalanceKey, newFromBalance.Bytes())
	if err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}

	toBalanceKey := engine.Hash{0x01} // Use special key for balance
	err = evm.storage.Set(to, toBalanceKey, newToBalance.Bytes())
	if err != nil {
		return fmt.Errorf("failed to update recipient balance: %w", err)
	}

	return nil
}

// markContractForDestruction marks a contract for destruction
func (evm *EVMEngine) markContractForDestruction(address engine.Address) {
	// Mark contract for destruction using special key
	destructionKey := engine.Hash{0x02} // Use special key for destruction flag
	if err := evm.storage.Set(address, destructionKey, []byte{0x01}); err != nil {
		// Log error but continue with destruction
		fmt.Printf("Failed to mark contract as destroyed: %v\n", err)
	}
}

// clearContractStorage clears the storage of a contract
func (evm *EVMEngine) clearContractStorage(address engine.Address) {
	// Clear all storage slots for the contract
	err := evm.storage.ClearContractStorage(address)
	if err != nil {
		// Log error but don't fail - this is best effort
		// In a real implementation, you'd have proper logging
	}
}

// getSenderNonce gets the nonce for a sender address
func (evm *EVMEngine) getSenderNonce(sender engine.Address) (uint64, error) {
	// Get nonce from storage using special key
	nonceKey := engine.Hash{0x03} // Use special key for nonce
	nonceBytes, err := evm.storage.Get(sender, nonceKey)
	if err != nil {
		return 0, err
	}

	// Convert bytes to uint64
	nonce := uint64(0)
	for i, b := range nonceBytes {
		nonce |= uint64(b) << (8 * (len(nonceBytes) - 1 - i))
	}
	return nonce, nil
}

// hasContractCode checks if an address has contract code
func (evm *EVMEngine) hasContractCode(address engine.Address) (bool, error) {
	codeKey := engine.Hash{} // Use zero hash for contract code
	code, err := evm.storage.Get(address, codeKey)
	if err != nil {
		return false, err
	}
	return len(code) > 0, nil
}

// executeContract executes the actual contract code
func (evm *EVMEngine) executeContract(ctx *ExecutionContext) (*engine.ExecutionResult, error) {
	// Initialize execution state
	// SECURITY FIX: Use validated gas limit instead of hardcoded value
	// The gas limit should be passed from the Execute method
	evm.gasMeter = engine.NewGasMeter(evm.gasLimit) // Use validated gas limit
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
		Logs:         []engine.Log{},
		StateChanges: []engine.StateChange{},
	}, nil
}

// generateEventLogs creates event logs for contract execution
func (ce *EVMEngine) generateEventLogs(contract *engine.Contract, input []byte, result *engine.ExecutionResult) []engine.Log {
	logs := []engine.Log{}

	// Generate logs based on function calls
	if len(input) >= 4 {
		// Extract function selector
		selector := input[:4]

		// Generate appropriate logs based on function
		switch {
		case bytes.Equal(selector, []byte{0x00, 0x00, 0x00, 0x01}): // transfer
			transferHash := engine.Hash{}
			copy(transferHash[:], []byte("Transfer"))
			logs = append(logs, engine.Log{
				Address: contract.Address,
				Topics:  []engine.Hash{transferHash},
				Data:    input[4:],
			})
		case bytes.Equal(selector, []byte{0x00, 0x00, 0x00, 0x02}): // approval
			approvalHash := engine.Hash{}
			copy(approvalHash[:], []byte("Approval"))
			logs = append(logs, engine.Log{
				Address: contract.Address,
				Topics:  []engine.Hash{approvalHash},
				Data:    input[4:],
			})
		default:
			// Generic function call log
			callHash := engine.Hash{}
			copy(callHash[:], []byte("FunctionCall"))
			logs = append(logs, engine.Log{
				Address: contract.Address,
				Topics:  []engine.Hash{callHash},
				Data:    input,
			})
		}
	}

	return logs
}

// generateStateChanges creates state changes for contract execution
func (ce *EVMEngine) generateStateChanges(contract *engine.Contract, input []byte, result *engine.ExecutionResult) []engine.StateChange {
	changes := []engine.StateChange{}

	// Generate state changes based on execution
	if result.Success {
		// Contract balance change
		changes = append(changes, engine.StateChange{
			Address: contract.Address,
			Key:     engine.Hash{0x01}, // Balance key
			Value:   contract.Balance.Bytes(),
		})

		// Nonce increment
		changes = append(changes, engine.StateChange{
			Address: contract.Address,
			Key:     engine.Hash{0x02}, // Nonce key
			Value:   []byte{byte(contract.Nonce + 1)},
		})
	}

	return changes
}
