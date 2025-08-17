package wasm

import (
	"fmt"
	"math/big"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// NewWASMEngine creates a new WASM execution engine
func NewWASMEngine(storage engine.ContractStorage, registry engine.ContractRegistry) *WASMEngine {
	return &WASMEngine{
		storage:    storage,
		registry:   registry,
		memory:     NewWASMMemory(65536, 1048576), // 1MB initial, 16MB max
		stack:      NewWASMStack(),
		globals:    make(map[uint32]*WASMGlobal),
		functions:  make(map[uint32]*WASMFunction),
		tables:     make(map[uint32]*WASMTable),
		instances:  make(map[string]*WASMInstance),
		gasPrice:   big.NewInt(0),
		blockNum:   0,
		timestamp:  0,
		coinbase:   engine.Address{},
		difficulty: big.NewInt(0),
		chainID:    big.NewInt(1),
	}
}

// Execute runs a WASM contract with given input and gas limit
func (w *WASMEngine) Execute(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Initialize execution context
	w.gasMeter = engine.NewGasMeter(gas)
	w.stack.Reset()
	w.memory.Reset(65536) // Reset to 1MB

	// Validate contract
	if contract == nil {
		return nil, engine.ErrInvalidContract
	}

	if len(contract.Code) == 0 {
		return nil, engine.ErrInvalidContract
	}

	// Parse WASM module
	module, err := w.parseWASMModule(contract.Code)
	if err != nil {
		return &engine.ExecutionResult{
			Success:      false,
			GasUsed:      w.gasMeter.GasConsumed(),
			GasRemaining: w.gasMeter.GasRemaining(),
			Error:        fmt.Errorf("failed to parse WASM module: %w", err),
		}, nil
	}

	// Create WASM instance
	instance := NewWASMInstance(module)
	w.instances[contract.Address.String()] = instance

	// Create execution context
	ctx := &ExecutionContext{
		Contract:   contract,
		Input:      input,
		Sender:     sender,
		Value:      value,
		GasPrice:   w.gasPrice,
		BlockNum:   w.blockNum,
		Timestamp:  w.timestamp,
		Coinbase:   w.coinbase,
		Difficulty: w.difficulty,
		ChainID:    w.chainID,
		Instance:   instance,
	}

	// Execute the contract
	result, err := w.executeWASMContract(ctx)
	if err != nil {
		return &engine.ExecutionResult{
			Success:      false,
			GasUsed:      w.gasMeter.GasConsumed(),
			GasRemaining: w.gasMeter.GasRemaining(),
			Error:        err,
		}, nil
	}

	return result, nil
}

// Deploy creates a new WASM contract with given code and constructor
func (w *WASMEngine) Deploy(code []byte, constructor []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.Contract, *engine.ExecutionResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Generate new contract address
	address := w.registry.GenerateAddress()
	
	// Create contract instance
	contract := &engine.Contract{
		Address: address,
		Code:    code,
		Creator: sender,
		Balance: big.NewInt(0),
		Nonce:   0,
	}
	
	// Register contract
	err := w.registry.Register(contract)
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
		result, err = w.executeWASMContractInternal(constructorContract, nil, gas, sender, value)
		if err != nil {
			// Rollback contract registration on failure
			w.registry.Remove(address)
			return nil, nil, fmt.Errorf("constructor execution failed: %w", err)
		}

		// Update contract in registry
		err = w.registry.UpdateContract(contract)
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

// EstimateGas estimates the gas cost for WASM contract execution
func (w *WASMEngine) EstimateGas(contract *engine.Contract, input []byte, sender engine.Address, value *big.Int) (uint64, error) {
	// Create a copy of the engine for estimation
	estimationEngine := w.Clone()
	
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

// Call executes a read-only WASM contract call
func (w *WASMEngine) Call(contract *engine.Contract, input []byte, sender engine.Address) ([]byte, error) {
	// Use a reasonable gas limit for calls
	gasLimit := uint64(1000000) // 1M gas
	
	result, err := w.Execute(contract, input, gasLimit, sender, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	
	if !result.Success {
		return nil, result.Error
	}
	
	return result.ReturnData, nil
}

// executeWASMContractInternal executes WASM code without requiring the mutex
// This is used internally to avoid deadlocks during constructor execution
func (w *WASMEngine) executeWASMContractInternal(contract *engine.Contract, input []byte, gas uint64, sender engine.Address, value *big.Int) (*engine.ExecutionResult, error) {
	// Initialize execution context
	ctx := &ExecutionContext{
		Contract:   contract,
		Input:      input,
		Sender:     sender,
		Value:      value,
		GasPrice:   w.gasPrice,
		BlockNum:   w.blockNum,
		Timestamp:  w.timestamp,
		Coinbase:   w.coinbase,
		Difficulty: w.difficulty,
		ChainID:    w.chainID,
	}

	// Initialize execution state
	w.gasMeter = engine.NewGasMeter(gas)
	w.stack.Reset()
	w.memory.Reset(65536)

	// Parse WASM module
	module, err := w.parseWASMModule(contract.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WASM module: %w", err)
	}

	// Create WASM instance
	instance := NewWASMInstance(module)
	ctx.Instance = instance

	// Execute the contract
	return w.executeWASMContract(ctx)
}

// executeWASMContract executes the actual WASM contract code
func (w *WASMEngine) executeWASMContract(ctx *ExecutionContext) (*engine.ExecutionResult, error) {
	// Load input data into memory if provided
	if len(ctx.Input) > 0 {
		err := w.memory.Write(0, ctx.Input)
		if err != nil {
			return &engine.ExecutionResult{
				Success:      false,
				GasUsed:      w.gasMeter.GasConsumed(),
				GasRemaining: w.gasMeter.GasRemaining(),
				Error:        fmt.Errorf("failed to write input to memory: %w", err),
			}, nil
		}
	}

	// Execute start function if present
	if ctx.Instance.Module.Start != nil {
		startIndex := *ctx.Instance.Module.Start
		if startIndex < uint32(len(ctx.Instance.Module.Functions)) {
			err := w.executeFunction(ctx, startIndex, nil)
			if err != nil {
				return &engine.ExecutionResult{
					Success:      false,
					GasUsed:      w.gasMeter.GasConsumed(),
					GasRemaining: w.gasMeter.GasRemaining(),
					Error:        fmt.Errorf("start function execution failed: %w", err),
				}, nil
			}
		}
	}

	// Success - return data from memory
	returnData, err := w.memory.Read(0, 32) // Return first 32 bytes
	if err != nil {
		returnData = []byte{} // Return empty if memory read fails
	}

	return &engine.ExecutionResult{
		Success:      true,
		ReturnData:   returnData,
		GasUsed:      w.gasMeter.GasConsumed(),
		GasRemaining: w.gasMeter.GasRemaining(),
		Logs:         []engine.Log{}, // TODO: Implement event logging
		StateChanges: []engine.StateChange{}, // TODO: Implement state change tracking
	}, nil
}

// parseWASMModule parses WASM bytecode into a module structure
func (w *WASMEngine) parseWASMModule(code []byte) (*WASMModule, error) {
	// Basic WASM validation - check magic number
	if len(code) < 8 {
		return nil, ErrInvalidWASM
	}
	
	// Check WASM magic number (0x00 0x61 0x73 0x6D)
	if code[0] != 0x00 || code[1] != 0x61 || code[2] != 0x73 || code[3] != 0x6D {
		return nil, ErrInvalidWASM
	}
	
	// Check version (should be 1)
	if code[4] != 0x01 || code[5] != 0x00 || code[6] != 0x00 || code[7] != 0x00 {
		return nil, ErrInvalidWASM
	}
	
	// For now, create a minimal module - full parsing would be implemented here
	module := NewWASMModule()
	
	// Create default memory (1MB)
	module.Memory = NewWASMMemory(65536, 1048576)
	
	// Create a simple function type (no params, no results)
	functionType := NewWASMFunctionType([]WASMValueType{}, []WASMValueType{})
	module.Types = append(module.Types, functionType)
	
	// Create a simple function
	function := NewWASMFunction(functionType, code, []WASMValueType{}, code)
	module.Functions = append(module.Functions, function)
	
	return module, nil
}

// executeFunction executes a WASM function
func (w *WASMEngine) executeFunction(ctx *ExecutionContext, functionIndex uint32, params []WASMValue) error {
	if functionIndex >= uint32(len(ctx.Instance.Module.Functions)) {
		return ErrFunctionNotFound
	}
	
	// Push parameters onto stack
	for _, param := range params {
		w.stack.Push(param)
	}
	
	// Execute function body (simplified - would implement full WASM execution here)
	// For now, just consume some gas and return
	err := w.gasMeter.ConsumeGas(100, "WASM function execution")
	if err != nil {
		return err
	}
	
	// TODO: Implement full WASM bytecode execution
	// This would include:
	// - Opcode decoding and execution
	// - Local variable management
	// - Control flow (if, loop, block)
	// - Function calls
	// - Memory operations
	
	return nil
}

// Clone creates a deep copy of the WASM engine for estimation
func (w *WASMEngine) Clone() *WASMEngine {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	clone := &WASMEngine{
		storage:    w.storage,
		registry:   w.registry,
		memory:     w.memory.Clone(),
		stack:      w.stack.Clone(),
		globals:    make(map[uint32]*WASMGlobal),
		functions:  make(map[uint32]*WASMFunction),
		tables:     make(map[uint32]*WASMTable),
		instances:  make(map[string]*WASMInstance),
		gasPrice:   new(big.Int).Set(w.gasPrice),
		blockNum:   w.blockNum,
		timestamp:  w.timestamp,
		coinbase:   w.coinbase,
		difficulty: new(big.Int).Set(w.difficulty),
		chainID:    new(big.Int).Set(w.chainID),
	}
	
	// Clone globals
	for k, v := range w.globals {
		clone.globals[k] = &WASMGlobal{
			Type:    v.Type,
			Mutable: v.Mutable,
			Value:   v.Value.Clone(),
		}
	}
	
	// Clone functions
	for k, v := range w.functions {
		clone.functions[k] = &WASMFunction{
			Type:       v.Type,
			Code:       append([]byte{}, v.Code...),
			LocalTypes: append([]WASMValueType{}, v.LocalTypes...),
			Body:       append([]byte{}, v.Body...),
		}
	}
	
	// Clone tables
	for k, v := range w.tables {
		clone.tables[k] = &WASMTable{
			ElementType: v.ElementType,
			Initial:     v.Initial,
			Maximum:     v.Maximum,
			Elements:    make([]WASMValue, len(v.Elements)),
		}
		for i, elem := range v.Elements {
			clone.tables[k].Elements[i] = elem.Clone()
		}
	}
	
	return clone
}

// SetBlockContext sets the block context for WASM execution
func (w *WASMEngine) SetBlockContext(blockNum uint64, timestamp uint64, coinbase engine.Address, difficulty *big.Int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.blockNum = blockNum
	w.timestamp = timestamp
	w.coinbase = coinbase
	w.difficulty = new(big.Int).Set(difficulty)
}

// SetGasPrice sets the gas price for WASM execution
func (w *WASMEngine) SetGasPrice(gasPrice *big.Int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.gasPrice = new(big.Int).Set(gasPrice)
}

// SetChainID sets the chain ID for WASM execution
func (w *WASMEngine) SetChainID(chainID *big.Int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.chainID = new(big.Int).Set(chainID)
}

// GetMemory returns the current memory instance
func (w *WASMEngine) GetMemory() *WASMMemory {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.memory
}

// GetStack returns the current stack instance
func (w *WASMEngine) GetStack() *WASMStack {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stack
}

// GetInstance returns a WASM instance by contract address
func (w *WASMEngine) GetInstance(address engine.Address) *WASMInstance {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.instances[address.String()]
}
