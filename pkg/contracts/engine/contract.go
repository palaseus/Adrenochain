package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

// Address represents a contract or account address
type Address [20]byte

// Hash represents a 32-byte hash
type Hash [32]byte

// Contract represents a deployed smart contract
type Contract struct {
	Address     Address
	Code        []byte
	CodeHash    Hash
	Balance     *big.Int
	Nonce       uint64
	StorageRoot Hash
	CreatedAt   time.Time
	Creator     Address
}

// NewContract creates a new contract instance
func NewContract(address Address, code []byte, creator Address) *Contract {
	codeHash := sha256.Sum256(code)
	return &Contract{
		Address:     address,
		Code:        code,
		CodeHash:    codeHash,
		Balance:     big.NewInt(0),
		Nonce:       0,
		StorageRoot: Hash{},
		CreatedAt:   time.Now(),
		Creator:     creator,
	}
}

// ExecutionResult represents the result of contract execution
type ExecutionResult struct {
	Success      bool
	ReturnData   []byte
	GasUsed      uint64
	GasRemaining uint64
	Logs         []Log
	Error        error
	StateChanges []StateChange
}

// Log represents a contract event log
type Log struct {
	Address     Address
	Topics      []Hash
	Data        []byte
	BlockNumber uint64
	TxHash      Hash
	TxIndex     uint
	Index       uint
	Removed     bool
}

// StateChange represents a change to contract state
type StateChange struct {
	Address Address
	Key     Hash
	Value   []byte
	Type    StateChangeType
}

// StateChangeType indicates the type of state change
type StateChangeType int

const (
	StateChangeStorage StateChangeType = iota
	StateChangeBalance
	StateChangeCode
	StateChangeNonce
)

// GasMeter tracks gas consumption during contract execution
type GasMeter interface {
	ConsumeGas(amount uint64, operation string) error
	RefundGas(amount uint64, operation string)
	GasConsumed() uint64
	GasRemaining() uint64
	Reset()
	IsOutOfGas() bool
}

// ContractEngine defines the interface for executing smart contracts
type ContractEngine interface {
	// Execute runs a contract with given input and gas limit
	Execute(contract *Contract, input []byte, gas uint64, sender Address, value *big.Int) (*ExecutionResult, error)

	// Deploy creates a new contract with given code and constructor
	Deploy(code []byte, constructor []byte, gas uint64, sender Address, value *big.Int) (*Contract, *ExecutionResult, error)

	// EstimateGas estimates the gas cost for contract execution
	EstimateGas(contract *Contract, input []byte, sender Address, value *big.Int) (uint64, error)

	// Call executes a read-only contract call
	Call(contract *Contract, input []byte, sender Address) ([]byte, error)
}

// ContractStorage defines the interface for contract state storage
type ContractStorage interface {
	// Get retrieves a value from contract storage
	Get(address Address, key Hash) ([]byte, error)

	// Set stores a value in contract storage
	Set(address Address, key Hash, value []byte) error

	// Delete removes a value from contract storage
	Delete(address Address, key Hash) error

	// GetStorageRoot returns the storage root hash for a contract
	GetStorageRoot(address Address) (Hash, error)

	// Commit commits all pending storage changes
	Commit() error

	// Rollback rolls back all pending storage changes
	Rollback() error
}

// ContractRegistry manages deployed contracts
type ContractRegistry interface {
	// Register adds a contract to the registry
	Register(contract *Contract) error

	// Get retrieves a contract by address
	Get(address Address) (*Contract, error)

	// Exists checks if a contract exists at the given address
	Exists(address Address) bool

	// Remove removes a contract from the registry
	Remove(address Address) error

	// List returns all registered contracts
	List() []*Contract

	// GetContractCount returns the total number of registered contracts
	GetContractCount() int

	// GetContractByCodeHash returns contracts with the specified code hash
	GetContractByCodeHash(codeHash Hash) []*Contract

	// GetContractsByCreator returns contracts created by the specified address
	GetContractsByCreator(creator Address) []*Contract

	// UpdateContract updates an existing contract
	UpdateContract(contract *Contract) error

	// GetContractAddresses returns all contract addresses
	GetContractAddresses() []Address

	// HasContracts checks if the registry has any contracts
	HasContracts() bool

	// Clear removes all contracts from the registry
	Clear()

	// GenerateAddress generates a new unique contract address
	GenerateAddress() Address

	// GetContractStats returns statistics about the registry
	GetContractStats() ContractStats
}

// Utility functions for address and hash operations
func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (a Address) Bytes() []byte {
	return a[:]
}

func (h Hash) Bytes() []byte {
	return h[:]
}

// ParseAddress converts a hex string to an Address
func ParseAddress(hexStr string) (Address, error) {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	if len(hexStr) != 40 {
		return Address{}, ErrInvalidAddress
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return Address{}, err
	}

	var addr Address
	copy(addr[:], bytes)
	return addr, nil
}

// ParseHash converts a hex string to a Hash
func ParseHash(hexStr string) (Hash, error) {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	if len(hexStr) != 64 {
		return Hash{}, ErrInvalidHash
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return Hash{}, err
	}

	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}
