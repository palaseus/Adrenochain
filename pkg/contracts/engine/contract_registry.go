package engine

import (
	"crypto/rand"
	"fmt"
	"sync"
)

// ContractRegistryImpl implements the ContractRegistry interface
type ContractRegistryImpl struct {
	contracts map[string]*Contract
	addresses map[string]Address
	mu        sync.RWMutex
}

// NewContractRegistry creates a new contract registry
func NewContractRegistry() *ContractRegistryImpl {
	return &ContractRegistryImpl{
		contracts: make(map[string]*Contract),
		addresses: make(map[string]Address),
	}
}

// Register adds a contract to the registry
func (cr *ContractRegistryImpl) Register(contract *Contract) error {
	if contract == nil {
		return fmt.Errorf("%w: contract cannot be nil", ErrInvalidContract)
	}

	if contract.Address == (Address{}) {
		return fmt.Errorf("%w: contract address cannot be zero", ErrInvalidContract)
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	addressStr := contract.Address.String()

	// Check if contract already exists
	if _, exists := cr.contracts[addressStr]; exists {
		return fmt.Errorf("%w: contract already exists at address %s", ErrInvalidContract, addressStr)
	}

	// Register the contract
	cr.contracts[addressStr] = contract
	cr.addresses[addressStr] = contract.Address

	return nil
}

// Get retrieves a contract by address
func (cr *ContractRegistryImpl) Get(address Address) (*Contract, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	addressStr := address.String()
	contract, exists := cr.contracts[addressStr]
	if !exists {
		return nil, fmt.Errorf("%w: contract not found at address %s", ErrContractNotFound, addressStr)
	}

	return contract, nil
}

// Exists checks if a contract exists at the given address
func (cr *ContractRegistryImpl) Exists(address Address) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	addressStr := address.String()
	_, exists := cr.contracts[addressStr]
	return exists
}

// Remove removes a contract from the registry
func (cr *ContractRegistryImpl) Remove(address Address) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	addressStr := address.String()

	if _, exists := cr.contracts[addressStr]; !exists {
		return fmt.Errorf("%w: contract not found at address %s", ErrContractNotFound, addressStr)
	}

	delete(cr.contracts, addressStr)
	delete(cr.addresses, addressStr)

	return nil
}

// List returns all registered contracts
func (cr *ContractRegistryImpl) List() []*Contract {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	contracts := make([]*Contract, 0, len(cr.contracts))
	for _, contract := range cr.contracts {
		contracts = append(contracts, contract)
	}

	return contracts
}

// GetContractCount returns the total number of registered contracts
func (cr *ContractRegistryImpl) GetContractCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return len(cr.contracts)
}

// GetContractByCodeHash returns contracts with the specified code hash
func (cr *ContractRegistryImpl) GetContractByCodeHash(codeHash Hash) []*Contract {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var contracts []*Contract
	for _, contract := range cr.contracts {
		if contract.CodeHash == codeHash {
			contracts = append(contracts, contract)
		}
	}

	return contracts
}

// GetContractsByCreator returns contracts created by the specified address
func (cr *ContractRegistryImpl) GetContractsByCreator(creator Address) []*Contract {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var contracts []*Contract
	for _, contract := range cr.contracts {
		if contract.Creator == creator {
			contracts = append(contracts, contract)
		}
	}

	return contracts
}

// UpdateContract updates an existing contract
func (cr *ContractRegistryImpl) UpdateContract(contract *Contract) error {
	if contract == nil {
		return fmt.Errorf("%w: contract cannot be nil", ErrInvalidContract)
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	addressStr := contract.Address.String()

	if _, exists := cr.contracts[addressStr]; !exists {
		return fmt.Errorf("%w: contract not found at address %s", ErrContractNotFound, addressStr)
	}

	cr.contracts[addressStr] = contract
	return nil
}

// GetContractAddresses returns all contract addresses
func (cr *ContractRegistryImpl) GetContractAddresses() []Address {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	addresses := make([]Address, 0, len(cr.addresses))
	for _, address := range cr.addresses {
		addresses = append(addresses, address)
	}

	return addresses
}

// HasContracts checks if the registry has any contracts
func (cr *ContractRegistryImpl) HasContracts() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return len(cr.contracts) > 0
}

// Clear removes all contracts from the registry
func (cr *ContractRegistryImpl) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.contracts = make(map[string]*Contract)
	cr.addresses = make(map[string]Address)
}

// GenerateAddress generates a new unique contract address
func (cr *ContractRegistryImpl) GenerateAddress() Address {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	var address Address
	for {
		// Generate random bytes for address
		rand.Read(address[:])

		// Check if address is already used
		addressStr := address.String()
		if _, exists := cr.contracts[addressStr]; !exists {
			break
		}
	}

	return address
}

// GetContractStats returns statistics about the registry
func (cr *ContractRegistryImpl) GetContractStats() ContractStats {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	stats := ContractStats{
		TotalContracts: len(cr.contracts),
		TotalCodeSize:  0,
		UniqueCreators: make(map[string]bool),
	}

	for _, contract := range cr.contracts {
		stats.TotalCodeSize += len(contract.Code)
		stats.UniqueCreators[contract.Creator.String()] = true
	}

	stats.UniqueCreatorCount = len(stats.UniqueCreators)
	return stats
}

// ContractStats provides statistics about the contract registry
type ContractStats struct {
	TotalContracts     int
	TotalCodeSize      int
	UniqueCreators     map[string]bool
	UniqueCreatorCount int
}

// String returns a string representation of the contract registry
func (cr *ContractRegistryImpl) String() string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return fmt.Sprintf("ContractRegistry{contracts: %d}", len(cr.contracts))
}
