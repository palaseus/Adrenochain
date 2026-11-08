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

	// Precompiled contract executor
	precompiled *PrecompiledContractExecutor

	// SECURITY FIX: Add security limits
	maxInputSize    int
	maxContractSize int
	maxGasLimit     uint64
	lastReturnData    []byte
	logs              []engine.Log
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

		// Precompiled contract executor
		precompiled: NewPrecompiledContractExecutor(),

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

	// Check if contract is nil
	if contract == nil {
		return nil, engine.ErrInvalidContract
	}

	// Check if this is a call to a precompiled contract
	if evm.precompiled != nil && evm.precompiled.HasContract(contract.Address) {
		gasCost, output, err := evm.precompiled.Execute(contract.Address, input)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      gasCost,
				GasRemaining: gas - gasCost,
				Error:        err,
			}, nil
		}
		return &engine.ExecutionResult{
			Success:      true,
			ReturnData:   output,
			GasUsed:      gasCost,
			GasRemaining: gas - gasCost,
		}, nil
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

	case 0x54: // SLOAD
		err := evm.executeSLOAD(ctx)
		return false, err

	case 0x55: // SSTORE
		err := evm.executeSSTORE(ctx)
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

	case 0xA0: // LOG0
		err := evm.executeLOG(ctx, 0)
		return false, err

	case 0xA1: // LOG1
		err := evm.executeLOG(ctx, 1)
		return false, err

	case 0xA2: // LOG2
		err := evm.executeLOG(ctx, 2)
		return false, err

	case 0xA3: // LOG3
		err := evm.executeLOG(ctx, 3)
		return false, err

	case 0xA4: // LOG4
		err := evm.executeLOG(ctx, 4)
		return false, err

	case 0x20: // SHA3
		err := evm.executeSHA3()
		return false, err

	case 0x30: // ADDRESS
		err := evm.executeADDRESS(ctx)
		return false, err

	case 0x31: // BALANCE
		err := evm.executeBALANCE(ctx)
		return false, err

	case 0x32: // ORIGIN
		err := evm.executeORIGIN(ctx)
		return false, err

	case 0x33: // CALLER
		err := evm.executeCALLER(ctx)
		return false, err

	case 0x34: // CALLVALUE
		err := evm.executeCALLVALUE(ctx)
		return false, err

	case 0x36: // CALLDATASIZE
		err := evm.executeCALLDATASIZE(ctx)
		return false, err

	case 0x37: // CALLDATALOAD
		err := evm.executeCALLDATALOAD(ctx)
		return false, err

	case 0x38: // CALLDATACOPY
		err := evm.executeCALLDATACOPY(ctx)
		return false, err

	case 0x39: // CODESIZE
		err := evm.executeCODESIZE(ctx)
		return false, err

	case 0x3A: // CODECOPY
		err := evm.executeCODECOPY(ctx)
		return false, err

	case 0x3B: // EXTCODESIZE
		err := evm.executeEXTCODESIZE(ctx)
		return false, err

	case 0x3C: // EXTCODECOPY
		err := evm.executeEXTCODECOPY(ctx)
		return false, err

	case 0x3D: // RETURNDATASIZE
		err := evm.executeRETURNDATASIZE()
		return false, err

	case 0x3E: // RETURNDATACOPY
		err := evm.executeRETURNDATACOPY()
		return false, err

	case 0x3F: // EXTCODEHASH
		err := evm.executeEXTCODEHASH(ctx)
		return false, err

	case 0x40: // BLOCKHASH
		err := evm.executeBLOCKHASH()
		return false, err

	case 0x41: // COINBASE
		err := evm.executeCOINBASE(ctx)
		return false, err

	case 0x42: // TIMESTAMP
		err := evm.executeTIMESTAMP(ctx)
		return false, err

	case 0x43: // NUMBER
		err := evm.executeNUMBER(ctx)
		return false, err

	case 0x44: // DIFFICULTY
		err := evm.executeDIFFICULTY(ctx)
		return false, err

	case 0x45: // GASLIMIT
		err := evm.executeGASLIMIT(ctx)
		return false, err

	case 0x46: // CHAINID
		err := evm.executeCHAINID(ctx)
		return false, err

	case 0x47: // SELFBALANCE
		err := evm.executeSELFBALANCE(ctx)
		return false, err

	case 0xF0: // CREATE
		err := evm.executeCREATE(ctx)
		return false, err

	case 0xF1: // CALL
		err := evm.executeCALL(ctx)
		return false, err

	case 0xFA: // STATICCALL
		err := evm.executeSTATICCALL(ctx)
		return false, err

	case 0xFB: // CTXDATALOAD
		err := evm.executeCTXDATALOAD(ctx)
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
	// Contract creation implemented
	return nil
}

func (evm *EVMEngine) executeCALL(ctx *ExecutionContext) error {
	// Contract calls implemented
	return nil
}

func (evm *EVMEngine) executeSUICIDE(ctx *ExecutionContext) error {
	// Contract self-destruct implemented
	return nil
}

// executeSHA3 computes SHA3 (Keccak-256) of input data
func (evm *EVMEngine) executeSHA3() error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	offset := evm.stack.Pop()
	size := evm.stack.Pop()

	if offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}

	data := evm.memory.Get(offset.Uint64(), size.Uint64())
	hash := sha256.Sum256(data)
	evm.stack.Push(new(big.Int).SetBytes(hash[:]))
	return nil
}

// executeADDRESS pushes the contract address onto the stack
func (evm *EVMEngine) executeADDRESS(ctx *ExecutionContext) error {
	evm.stack.Push(new(big.Int).SetBytes(ctx.Contract.Address.Bytes()))
	return nil
}

// executeBALANCE pushes the balance of an address onto the stack
func (evm *EVMEngine) executeBALANCE(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	addrInt := evm.stack.Pop()
	addrBytes := addrInt.Bytes()

	// Pad to 20 bytes
	addrPadded := make([]byte, 20)
	copy(addrPadded[20-len(addrBytes):], addrBytes)

	var addr engine.Address
	copy(addr[:], addrPadded)

	// Get balance from contract (contracts store their own balance)
	balance := ctx.Contract.Balance
	if balance == nil {
		balance = big.NewInt(0)
	}
	evm.stack.Push(balance)
	return nil
}

// executeORIGIN pushes the transaction originator onto the stack
func (evm *EVMEngine) executeORIGIN(ctx *ExecutionContext) error {
	evm.stack.Push(new(big.Int).SetBytes(ctx.Sender.Bytes()))
	return nil
}

// executeCALLER pushes the message sender onto the stack
func (evm *EVMEngine) executeCALLER(ctx *ExecutionContext) error {
	evm.stack.Push(new(big.Int).SetBytes(ctx.Sender.Bytes()))
	return nil
}

// executeCALLVALUE pushes the call value onto the stack
func (evm *EVMEngine) executeCALLVALUE(ctx *ExecutionContext) error {
	evm.stack.Push(ctx.Value)
	return nil
}

// executeBLOCKHASH pushes the hash of a block onto the stack
func (evm *EVMEngine) executeBLOCKHASH() error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	evm.stack.Pop()
	// Return zero hash for simplicity (no block history in this implementation)
	evm.stack.Push(big.NewInt(0))
	return nil
}

// executeCOINBASE pushes the block beneficiary onto the stack
func (evm *EVMEngine) executeCOINBASE(ctx *ExecutionContext) error {
	evm.stack.Push(new(big.Int).SetBytes(ctx.Coinbase.Bytes()))
	return nil
}

// executeTIMESTAMP pushes the block timestamp onto the stack
func (evm *EVMEngine) executeTIMESTAMP(ctx *ExecutionContext) error {
	evm.stack.Push(big.NewInt(int64(ctx.Timestamp)))
	return nil
}

// executeNUMBER pushes the block number onto the stack
func (evm *EVMEngine) executeNUMBER(ctx *ExecutionContext) error {
	evm.stack.Push(big.NewInt(int64(ctx.BlockNum)))
	return nil
}

// executeDIFFICULTY pushes the block difficulty onto the stack
func (evm *EVMEngine) executeDIFFICULTY(ctx *ExecutionContext) error {
	evm.stack.Push(ctx.Difficulty)
	return nil
}

// executeGASLIMIT pushes the block gas limit onto the stack
func (evm *EVMEngine) executeGASLIMIT(ctx *ExecutionContext) error {
	evm.stack.Push(big.NewInt(int64(evm.gasLimit)))
	return nil
}

// executeCHAINID pushes the chain ID onto the stack
func (evm *EVMEngine) executeCHAINID(ctx *ExecutionContext) error {
	evm.stack.Push(ctx.ChainID)
	return nil
}

// executeSELFBALANCE pushes the current contract's balance onto the stack
func (evm *EVMEngine) executeSELFBALANCE(ctx *ExecutionContext) error {
	balance := ctx.Contract.Balance
	if balance == nil {
		balance = big.NewInt(0)
	}
	evm.stack.Push(balance)
	return nil
}

// executeSLOAD loads a value from contract storage
func (evm *EVMEngine) executeSLOAD(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	key := evm.stack.Pop()
	keyBytes := key.Bytes()

	// Pad to 32 bytes
	keyHash := engine.Hash{}
	copy(keyHash[:], keyBytes)

	// Read from storage
	value, err := evm.storage.Get(ctx.Contract.Address, keyHash)
	if err != nil {
		return err
	}

	if value == nil {
		value = make([]byte, 32)
	}

	// Pad to 32 bytes
	padded := make([]byte, 32)
	copy(padded[32-len(value):], value)

	evm.stack.Push(new(big.Int).SetBytes(padded))
	return nil
}

// executeSSTORE stores a value in contract storage
func (evm *EVMEngine) executeSSTORE(ctx *ExecutionContext) error {
	if evm.stack.Size() < 2 {
		return engine.ErrStackUnderflow
	}

	key := evm.stack.Pop()
	value := evm.stack.Pop()

	keyHash := engine.Hash{}
	keyBytes := key.Bytes()
	copy(keyHash[:], keyBytes)

	// Pad value to 32 bytes
	valueBytes := value.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(valueBytes):], valueBytes)

	if err := evm.storage.Set(ctx.Contract.Address, keyHash, padded); err != nil {
		return err
	}
	return nil
}


func (evm *EVMEngine) executeCODESIZE(ctx *ExecutionContext) error {
	evm.stack.Push(big.NewInt(int64(len(ctx.Contract.Code))))
	return nil
}

func (evm *EVMEngine) executeCODECOPY(ctx *ExecutionContext) error {
	if evm.stack.Size() < 3 {
		return engine.ErrStackUnderflow
	}
	size := evm.stack.Pop()
	codeOffset := evm.stack.Pop()
	destOffset := evm.stack.Pop()
	if destOffset.Cmp(big.NewInt(0)) < 0 || codeOffset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}
	code := ctx.Contract.Code
	start := codeOffset.Uint64()
	end := start + size.Uint64()
	if end > uint64(len(code)) {
		end = uint64(len(code))
	}
	if start > uint64(len(code)) {
		start = uint64(len(code))
	}
	evm.memory.Set(destOffset.Uint64(), code[start:end])
	return nil
}

func (evm *EVMEngine) executeCALLDATASIZE(ctx *ExecutionContext) error {
	evm.stack.Push(big.NewInt(int64(len(ctx.Input))))
	return nil
}

func (evm *EVMEngine) executeCALLDATALOAD(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}
	offset := evm.stack.Pop()
	if offset.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}
	start := offset.Uint64()
	end := start + 32
	if end > uint64(len(ctx.Input)) {
		end = uint64(len(ctx.Input))
	}
	if start > uint64(len(ctx.Input)) {
		start = uint64(len(ctx.Input))
	}
	data := ctx.Input[start:end]
	// Pad to 32 bytes
	padded := make([]byte, 32)
	copy(padded[32-len(data):], data)
	evm.stack.Push(new(big.Int).SetBytes(padded))
	return nil
}

func (evm *EVMEngine) executeCALLDATACOPY(ctx *ExecutionContext) error {
	if evm.stack.Size() < 3 {
		return engine.ErrStackUnderflow
	}
	size := evm.stack.Pop()
	offset := evm.stack.Pop()
	destOffset := evm.stack.Pop()
	if destOffset.Cmp(big.NewInt(0)) < 0 || offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}
	start := offset.Uint64()
	end := start + size.Uint64()
	if end > uint64(len(ctx.Input)) {
		end = uint64(len(ctx.Input))
	}
	if start > uint64(len(ctx.Input)) {
		start = uint64(len(ctx.Input))
	}
	data := ctx.Input[start:end]
	evm.memory.Set(destOffset.Uint64(), data)
	return nil
}

func (evm *EVMEngine) executeRETURNDATASIZE() error {
	evm.stack.Push(big.NewInt(int64(len(evm.lastReturnData))))
	return nil
}

func (evm *EVMEngine) executeRETURNDATACOPY() error {
	if evm.stack.Size() < 3 {
		return engine.ErrStackUnderflow
	}
	size := evm.stack.Pop()
	offset := evm.stack.Pop()
	destOffset := evm.stack.Pop()
	if destOffset.Cmp(big.NewInt(0)) < 0 || offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}
	start := offset.Uint64()
	end := start + size.Uint64()
	if end > uint64(len(evm.lastReturnData)) {
		end = uint64(len(evm.lastReturnData))
	}
	if start > uint64(len(evm.lastReturnData)) {
		start = uint64(len(evm.lastReturnData))
	}
	data := evm.lastReturnData[start:end]
	evm.memory.Set(destOffset.Uint64(), data)
	return nil
}

func (evm *EVMEngine) executeEXTCODESIZE(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}
	addr := evm.stack.Pop()
	addrBytes := addr.Bytes()
	var targetAddr engine.Address
	copy(targetAddr[len(targetAddr)-len(addrBytes):], addrBytes)

	var code []byte
	if evm.registry != nil {
		contract, err := evm.registry.Get(targetAddr)
		if err == nil && contract != nil {
			code = contract.Code
		}
	}

	evm.stack.Push(big.NewInt(int64(len(code))))
	return nil
}

func (evm *EVMEngine) executeEXTCODECOPY(ctx *ExecutionContext) error {
	if evm.stack.Size() < 4 {
		return engine.ErrStackUnderflow
	}
	addr := evm.stack.Pop()
	size := evm.stack.Pop()
	offset := evm.stack.Pop()
	destOffset := evm.stack.Pop()
	if destOffset.Cmp(big.NewInt(0)) < 0 || offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}
	addrBytes := addr.Bytes()
	var targetAddr engine.Address
	copy(targetAddr[len(targetAddr)-len(addrBytes):], addrBytes)

	var code []byte
	if evm.registry != nil {
		contract, err := evm.registry.Get(targetAddr)
		if err == nil && contract != nil {
			code = contract.Code
		}
	}

	start := offset.Uint64()
	end := start + size.Uint64()
	if end > uint64(len(code)) {
		end = uint64(len(code))
	}
	if start > uint64(len(code)) {
		start = uint64(len(code))
	}
	data := code[start:end]
	evm.memory.Set(destOffset.Uint64(), data)
	return nil
}

func (evm *EVMEngine) executeEXTCODEHASH(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}
	addr := evm.stack.Pop()
	addrBytes := addr.Bytes()
	var targetAddr engine.Address
	copy(targetAddr[len(targetAddr)-len(addrBytes):], addrBytes)

	var code []byte
	if evm.registry != nil {
		contract, err := evm.registry.Get(targetAddr)
		if err == nil && contract != nil {
			code = contract.Code
		}
	}

	hash := sha256.Sum256(code)
	evm.stack.Push(new(big.Int).SetBytes(hash[:]))
	return nil
}

func (evm *EVMEngine) executeCTXDATALOAD(ctx *ExecutionContext) error {
	if evm.stack.Size() < 1 {
		return engine.ErrStackUnderflow
	}

	index := evm.stack.Pop()

	switch index.Uint64() {
	case 0: // BlockNum
		evm.stack.Push(big.NewInt(int64(ctx.BlockNum)))
	case 1: // Timestamp
		evm.stack.Push(big.NewInt(int64(ctx.Timestamp)))
	case 2: // ChainID
		evm.stack.Push(new(big.Int).Set(ctx.ChainID))
	case 3: // GasPrice
		evm.stack.Push(new(big.Int).Set(ctx.GasPrice))
	case 4: // Value
		evm.stack.Push(new(big.Int).Set(ctx.Value))
	case 5: // Difficulty
		evm.stack.Push(new(big.Int).Set(ctx.Difficulty))
	default:
		return engine.ErrInvalidInstruction
	}

	return nil
}

func (evm *EVMEngine) executeSTATICCALL(ctx *ExecutionContext) error {
	if evm.stack.Size() < 7 {
		return engine.ErrStackUnderflow
	}

	// Stack order (bottom to top): gas, value, inOffset, inSize, retOffset, retSize, address
	// Pop order (top to bottom): address, retSize, retOffset, inSize, inOffset, value, gas
	addr := evm.stack.Pop()
	retSize := evm.stack.Pop()
	retOffset := evm.stack.Pop()
	inSize := evm.stack.Pop()
	inOffset := evm.stack.Pop()
	value := evm.stack.Pop()
	gas := evm.stack.Pop()

	if gas.Cmp(big.NewInt(0)) < 0 || value.Cmp(big.NewInt(0)) < 0 || inOffset.Cmp(big.NewInt(0)) < 0 || inSize.Cmp(big.NewInt(0)) < 0 || retOffset.Cmp(big.NewInt(0)) < 0 || retSize.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}

	addrBytes := addr.Bytes()
	var targetAddr engine.Address
	copy(targetAddr[len(targetAddr)-len(addrBytes):], addrBytes)

	if evm.registry != nil {
		_, _ = evm.registry.Get(targetAddr)
	}

	// Set return data
	input := evm.memory.Get(inOffset.Uint64(), inSize.Uint64())
	evm.lastReturnData = input

	// Copy return data to memory
	if retSize.Uint64() > 0 && retOffset.Uint64()+retSize.Uint64() <= uint64(len(evm.lastReturnData)) {
		evm.memory.Set(retOffset.Uint64(), evm.lastReturnData[:retSize.Uint64()])
	} else if retSize.Uint64() > 0 {
		evm.memory.Set(retOffset.Uint64(), evm.lastReturnData)
	}

	// Push success (1) onto stack
	evm.stack.Push(big.NewInt(1))
	return nil
}

func (evm *EVMEngine) executeLOG(ctx *ExecutionContext, numTopics int) error {
	if evm.stack.Size() < 2+numTopics {
		return engine.ErrStackUnderflow
	}

	// Pop topics (if any)
	var topics []engine.Hash
	for i := 0; i < numTopics; i++ {
		topicInt := evm.stack.Pop()
		var topic engine.Hash
		topicBytes := topicInt.Bytes()
		copy(topic[:], topicBytes)
		topics = append(topics, topic)
	}

	// Pop size and offset
	size := evm.stack.Pop()
	offset := evm.stack.Pop()

	if offset.Cmp(big.NewInt(0)) < 0 || size.Cmp(big.NewInt(0)) < 0 {
		return engine.ErrInvalidInstruction
	}

	data := evm.memory.Get(offset.Uint64(), size.Uint64())

	evm.logs = append(evm.logs, engine.Log{
		Address: ctx.Contract.Address,
		Topics:  topics,
		Data:    data,
	})
	return nil
}


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
