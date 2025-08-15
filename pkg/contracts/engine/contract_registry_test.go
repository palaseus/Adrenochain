package engine

import (
	"crypto/rand"
	"testing"
)

func TestNewContractRegistry(t *testing.T) {
	cr := NewContractRegistry()

	if cr == nil {
		t.Fatal("expected non-nil contract registry")
	}

	if cr.GetContractCount() != 0 {
		t.Errorf("expected initial contract count 0, got %d", cr.GetContractCount())
	}

	if cr.HasContracts() {
		t.Error("expected no contracts initially")
	}
}

func TestContractRegistryRegister(t *testing.T) {
	cr := NewContractRegistry()

	// Create a test contract
	address := generateRandomAddress()
	code := []byte("test contract code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	// Test successful registration
	err := cr.Register(contract)
	if err != nil {
		t.Errorf("unexpected error registering contract: %v", err)
	}

	if !cr.Exists(address) {
		t.Error("contract should exist after registration")
	}

	if cr.GetContractCount() != 1 {
		t.Errorf("expected contract count 1, got %d", cr.GetContractCount())
	}

	// Test duplicate registration
	err = cr.Register(contract)
	if err == nil {
		t.Error("expected error when registering duplicate contract")
	}
	if err == nil || !contains(err.Error(), ErrInvalidContract.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrInvalidContract, err)
	}

	// Test nil contract
	err = cr.Register(nil)
	if err == nil {
		t.Error("expected error when registering nil contract")
	}
	if err == nil || !contains(err.Error(), ErrInvalidContract.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrInvalidContract, err)
	}

	// Test zero address
	zeroContract := NewContract(Address{}, code, creator)
	err = cr.Register(zeroContract)
	if err == nil {
		t.Error("expected error when registering contract with zero address")
	}
	if err == nil || !contains(err.Error(), ErrInvalidContract.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrInvalidContract, err)
	}
}

func TestContractRegistryGet(t *testing.T) {
	cr := NewContractRegistry()

	// Create and register a contract
	address := generateRandomAddress()
	code := []byte("test contract code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	// Test successful retrieval
	retrieved, err := cr.Get(address)
	if err != nil {
		t.Errorf("unexpected error getting contract: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected non-nil contract")
	}

	if retrieved.Address != address {
		t.Errorf("expected address %v, got %v", address, retrieved.Address)
	}

	// Test getting non-existent contract
	nonExistentAddr := generateRandomAddress()
	_, err = cr.Get(nonExistentAddr)
	if err == nil {
		t.Error("expected error when getting non-existent contract")
	}
	if err == nil || !contains(err.Error(), ErrContractNotFound.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrContractNotFound, err)
	}
}

func TestContractRegistryExists(t *testing.T) {
	cr := NewContractRegistry()

	// Test non-existent contract
	address := generateRandomAddress()
	if cr.Exists(address) {
		t.Error("expected contract to not exist")
	}

	// Register contract and test existence
	code := []byte("test contract code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	if !cr.Exists(address) {
		t.Error("expected contract to exist after registration")
	}
}

func TestContractRegistryRemove(t *testing.T) {
	cr := NewContractRegistry()

	// Create and register a contract
	address := generateRandomAddress()
	code := []byte("test contract code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	// Test successful removal
	err = cr.Remove(address)
	if err != nil {
		t.Errorf("unexpected error removing contract: %v", err)
	}

	if cr.Exists(address) {
		t.Error("contract should not exist after removal")
	}

	if cr.GetContractCount() != 0 {
		t.Errorf("expected contract count 0 after removal, got %d", cr.GetContractCount())
	}

	// Test removing non-existent contract
	err = cr.Remove(address)
	if err == nil {
		t.Error("expected error when removing non-existent contract")
	}
	if err == nil || !contains(err.Error(), ErrContractNotFound.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrContractNotFound, err)
	}
}

func TestContractRegistryList(t *testing.T) {
	cr := NewContractRegistry()

	// Test empty list
	contracts := cr.List()
	if len(contracts) != 0 {
		t.Errorf("expected empty list, got %d contracts", len(contracts))
	}

	// Register multiple contracts
	numContracts := 5
	registeredContracts := make([]*Contract, numContracts)

	for i := 0; i < numContracts; i++ {
		address := generateRandomAddress()
		code := []byte{byte(i)}
		creator := generateRandomAddress()
		contract := NewContract(address, code, creator)

		err := cr.Register(contract)
		if err != nil {
			t.Fatalf("failed to register contract %d: %v", i, err)
		}

		registeredContracts[i] = contract
	}

	// Test list retrieval
	contracts = cr.List()
	if len(contracts) != numContracts {
		t.Errorf("expected %d contracts, got %d", numContracts, len(contracts))
	}

	// Verify all contracts are in the list
	for _, expected := range registeredContracts {
		found := false
		for _, actual := range contracts {
			if actual.Address == expected.Address {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("contract with address %v not found in list", expected.Address)
		}
	}
}

func TestContractRegistryGetContractCount(t *testing.T) {
	cr := NewContractRegistry()

	// Test initial count
	if cr.GetContractCount() != 0 {
		t.Errorf("expected initial count 0, got %d", cr.GetContractCount())
	}

	// Register contracts and test count
	numContracts := 3
	contracts := make([]*Contract, numContracts)

	for i := 0; i < numContracts; i++ {
		address := generateRandomAddress()
		code := []byte{byte(i)}
		creator := generateRandomAddress()
		contract := NewContract(address, code, creator)
		contracts[i] = contract

		err := cr.Register(contract)
		if err != nil {
			t.Fatalf("failed to register contract %d: %v", i, err)
		}

		if cr.GetContractCount() != i+1 {
			t.Errorf("expected count %d after registering contract %d, got %d", i+1, i, cr.GetContractCount())
		}
	}

	// Remove contracts and test count
	for i := numContracts - 1; i >= 0; i-- {
		address := contracts[i].Address
		err := cr.Remove(address)
		if err != nil {
			t.Fatalf("failed to remove contract %d: %v", i, err)
		}

		if cr.GetContractCount() != i {
			t.Errorf("expected count %d after removing contract %d, got %d", i, i, cr.GetContractCount())
		}
	}
}

func TestContractRegistryGetContractByCodeHash(t *testing.T) {
	cr := NewContractRegistry()

	// Register contracts with different code hashes
	code1 := []byte("contract 1")
	code2 := []byte("contract 2")

	contract1 := NewContract(generateRandomAddress(), code1, generateRandomAddress())
	contract2 := NewContract(generateRandomAddress(), code2, generateRandomAddress())
	contract3 := NewContract(generateRandomAddress(), code1, generateRandomAddress()) // Same code as contract1

	err := cr.Register(contract1)
	if err != nil {
		t.Fatalf("failed to register contract1: %v", err)
	}

	err = cr.Register(contract2)
	if err != nil {
		t.Fatalf("failed to register contract2: %v", err)
	}

	err = cr.Register(contract3)
	if err != nil {
		t.Fatalf("failed to register contract3: %v", err)
	}

	// Test finding contracts by code hash
	contractsWithCode1 := cr.GetContractByCodeHash(contract1.CodeHash)
	if len(contractsWithCode1) != 2 {
		t.Errorf("expected 2 contracts with code hash %v, got %d", contract1.CodeHash, len(contractsWithCode1))
	}

	contractsWithCode2 := cr.GetContractByCodeHash(contract2.CodeHash)
	if len(contractsWithCode2) != 1 {
		t.Errorf("expected 1 contract with code hash %v, got %d", contract2.CodeHash, len(contractsWithCode2))
	}

	// Test finding non-existent code hash
	nonExistentHash := Hash{}
	contractsWithNonExistentHash := cr.GetContractByCodeHash(nonExistentHash)
	if len(contractsWithNonExistentHash) != 0 {
		t.Errorf("expected 0 contracts with non-existent code hash, got %d", len(contractsWithNonExistentHash))
	}
}

func TestContractRegistryGetContractsByCreator(t *testing.T) {
	cr := NewContractRegistry()

	// Register contracts with different creators
	creator1 := generateRandomAddress()
	creator2 := generateRandomAddress()

	contract1 := NewContract(generateRandomAddress(), []byte("code1"), creator1)
	contract2 := NewContract(generateRandomAddress(), []byte("code2"), creator2)
	contract3 := NewContract(generateRandomAddress(), []byte("code3"), creator1) // Same creator as contract1

	err := cr.Register(contract1)
	if err != nil {
		t.Fatalf("failed to register contract1: %v", err)
	}

	err = cr.Register(contract2)
	if err != nil {
		t.Fatalf("failed to register contract2: %v", err)
	}

	err = cr.Register(contract3)
	if err != nil {
		t.Fatalf("failed to register contract3: %v", err)
	}

	// Test finding contracts by creator
	contractsByCreator1 := cr.GetContractsByCreator(creator1)
	if len(contractsByCreator1) != 2 {
		t.Errorf("expected 2 contracts by creator %v, got %d", creator1, len(contractsByCreator1))
	}

	contractsByCreator2 := cr.GetContractsByCreator(creator2)
	if len(contractsByCreator2) != 1 {
		t.Errorf("expected 1 contract by creator %v, got %d", creator2, len(contractsByCreator2))
	}

	// Test finding contracts by non-existent creator
	nonExistentCreator := generateRandomAddress()
	contractsByNonExistentCreator := cr.GetContractsByCreator(nonExistentCreator)
	if len(contractsByNonExistentCreator) != 0 {
		t.Errorf("expected 0 contracts by non-existent creator, got %d", len(contractsByNonExistentCreator))
	}
}

func TestContractRegistryUpdateContract(t *testing.T) {
	cr := NewContractRegistry()

	// Create and register a contract
	address := generateRandomAddress()
	code := []byte("original code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	// Update the contract
	updatedCode := []byte("updated code")
	contract.Code = updatedCode
	contract.CodeHash = Hash{} // Reset hash to simulate update

	err = cr.UpdateContract(contract)
	if err != nil {
		t.Errorf("unexpected error updating contract: %v", err)
	}

	// Verify the update
	retrieved, err := cr.Get(address)
	if err != nil {
		t.Fatalf("failed to get updated contract: %v", err)
	}

	if string(retrieved.Code) != string(updatedCode) {
		t.Errorf("expected updated code %s, got %s", string(updatedCode), string(retrieved.Code))
	}

	// Test updating non-existent contract
	nonExistentContract := NewContract(generateRandomAddress(), []byte("code"), generateRandomAddress())
	err = cr.UpdateContract(nonExistentContract)
	if err == nil {
		t.Error("expected error when updating non-existent contract")
	}
	if err == nil || !contains(err.Error(), ErrContractNotFound.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrContractNotFound, err)
	}

	// Test updating with nil contract
	err = cr.UpdateContract(nil)
	if err == nil {
		t.Error("expected error when updating nil contract")
	}
	if err == nil || !contains(err.Error(), ErrInvalidContract.Error()) {
		t.Errorf("expected error containing '%v', got %v", ErrInvalidContract, err)
	}
}

func TestContractRegistryGetContractAddresses(t *testing.T) {
	cr := NewContractRegistry()

	// Test empty addresses list
	addresses := cr.GetContractAddresses()
	if len(addresses) != 0 {
		t.Errorf("expected empty addresses list, got %d addresses", len(addresses))
	}

	// Register contracts and test addresses list
	numContracts := 3
	expectedAddresses := make([]Address, numContracts)

	for i := 0; i < numContracts; i++ {
		address := generateRandomAddress()
		code := []byte{byte(i)}
		creator := generateRandomAddress()
		contract := NewContract(address, code, creator)

		err := cr.Register(contract)
		if err != nil {
			t.Fatalf("failed to register contract %d: %v", i, err)
		}

		expectedAddresses[i] = address
	}

	// Test addresses retrieval
	addresses = cr.GetContractAddresses()
	if len(addresses) != numContracts {
		t.Errorf("expected %d addresses, got %d", numContracts, len(addresses))
	}

	// Verify all addresses are in the list
	for _, expected := range expectedAddresses {
		found := false
		for _, actual := range addresses {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("address %v not found in addresses list", expected)
		}
	}
}

func TestContractRegistryHasContracts(t *testing.T) {
	cr := NewContractRegistry()

	// Test initial state
	if cr.HasContracts() {
		t.Error("expected no contracts initially")
	}

	// Register a contract and test
	address := generateRandomAddress()
	code := []byte("test code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	if !cr.HasContracts() {
		t.Error("expected contracts to exist after registration")
	}

	// Remove contract and test
	err = cr.Remove(address)
	if err != nil {
		t.Fatalf("failed to remove contract: %v", err)
	}

	if cr.HasContracts() {
		t.Error("expected no contracts after removal")
	}
}

func TestContractRegistryClear(t *testing.T) {
	cr := NewContractRegistry()

	// Register some contracts
	numContracts := 3
	for i := 0; i < numContracts; i++ {
		address := generateRandomAddress()
		code := []byte{byte(i)}
		creator := generateRandomAddress()
		contract := NewContract(address, code, creator)

		err := cr.Register(contract)
		if err != nil {
			t.Fatalf("failed to register contract %d: %v", i, err)
		}
	}

	// Verify contracts exist
	if cr.GetContractCount() != numContracts {
		t.Errorf("expected %d contracts before clear, got %d", numContracts, cr.GetContractCount())
	}

	// Clear registry
	cr.Clear()

	// Verify registry is empty
	if cr.GetContractCount() != 0 {
		t.Errorf("expected 0 contracts after clear, got %d", cr.GetContractCount())
	}

	if cr.HasContracts() {
		t.Error("expected no contracts after clear")
	}

	// Verify individual contracts don't exist
	for i := 0; i < numContracts; i++ {
		address := generateRandomAddress()
		if cr.Exists(address) {
			t.Errorf("contract %d should not exist after clear", i)
		}
	}
}

func TestContractRegistryGenerateAddress(t *testing.T) {
	cr := NewContractRegistry()

	// Generate multiple addresses
	numAddresses := 10
	addresses := make([]Address, numAddresses)

	for i := 0; i < numAddresses; i++ {
		address := cr.GenerateAddress()

		// Verify address is not zero
		if address == (Address{}) {
			t.Errorf("generated address %d is zero", i)
		}

		// Verify address is unique
		for j := 0; j < i; j++ {
			if addresses[j] == address {
				t.Errorf("duplicate address generated: %v", address)
			}
		}

		addresses[i] = address
	}
}

func TestContractRegistryGetContractStats(t *testing.T) {
	cr := NewContractRegistry()

	// Test empty stats
	stats := cr.GetContractStats()
	if stats.TotalContracts != 0 {
		t.Errorf("expected 0 total contracts, got %d", stats.TotalContracts)
	}
	if stats.TotalCodeSize != 0 {
		t.Errorf("expected 0 total code size, got %d", stats.TotalCodeSize)
	}
	if stats.UniqueCreatorCount != 0 {
		t.Errorf("expected 0 unique creators, got %d", stats.UniqueCreatorCount)
	}

	// Register contracts and test stats
	creator1 := generateRandomAddress()
	creator2 := generateRandomAddress()

	contract1 := NewContract(generateRandomAddress(), []byte("code1"), creator1)
	contract2 := NewContract(generateRandomAddress(), []byte("code2"), creator2)
	contract3 := NewContract(generateRandomAddress(), []byte("code3"), creator1) // Same creator

	err := cr.Register(contract1)
	if err != nil {
		t.Fatalf("failed to register contract1: %v", err)
	}

	err = cr.Register(contract2)
	if err != nil {
		t.Fatalf("failed to register contract2: %v", err)
	}

	err = cr.Register(contract3)
	if err != nil {
		t.Fatalf("failed to register contract3: %v", err)
	}

	// Test populated stats
	stats = cr.GetContractStats()
	if stats.TotalContracts != 3 {
		t.Errorf("expected 3 total contracts, got %d", stats.TotalContracts)
	}

	expectedCodeSize := len(contract1.Code) + len(contract2.Code) + len(contract3.Code)
	if stats.TotalCodeSize != expectedCodeSize {
		t.Errorf("expected total code size %d, got %d", expectedCodeSize, stats.TotalCodeSize)
	}

	if stats.UniqueCreatorCount != 2 {
		t.Errorf("expected 2 unique creators, got %d", stats.UniqueCreatorCount)
	}

	// Verify unique creators
	if !stats.UniqueCreators[creator1.String()] {
		t.Error("creator1 not found in unique creators")
	}
	if !stats.UniqueCreators[creator2.String()] {
		t.Error("creator2 not found in unique creators")
	}
}

func TestContractRegistryString(t *testing.T) {
	cr := NewContractRegistry()

	// Test empty registry string
	str := cr.String()
	expectedEmpty := "ContractRegistry{contracts: 0}"
	if str != expectedEmpty {
		t.Errorf("expected string '%s', got '%s'", expectedEmpty, str)
	}

	// Register a contract and test string
	address := generateRandomAddress()
	code := []byte("test code")
	creator := generateRandomAddress()
	contract := NewContract(address, code, creator)

	err := cr.Register(contract)
	if err != nil {
		t.Fatalf("failed to register contract: %v", err)
	}

	str = cr.String()
	expectedWithContract := "ContractRegistry{contracts: 1}"
	if str != expectedWithContract {
		t.Errorf("expected string '%s', got '%s'", expectedWithContract, str)
	}
}

// Helper function to generate random addresses for testing
func generateRandomAddress() Address {
	var address Address
	rand.Read(address[:])
	return address
}

// Helper function to generate random hashes for testing
func generateRandomHash() Hash {
	var hash Hash
	rand.Read(hash[:])
	return hash
}


