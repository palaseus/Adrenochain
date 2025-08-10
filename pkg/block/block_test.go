package block

import (
	"fmt"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {
	prevHash := []byte("previous_block_hash_123456789")
	height := uint64(1)
	difficulty := uint64(1000)

	block := NewBlock(prevHash, height, difficulty)

	if block == nil {
		t.Fatal("NewBlock returned nil")
	}

	if block.Header == nil {
		t.Fatal("Block header is nil")
	}

	if block.Header.Version != 1 {
		t.Errorf("Expected version 1, got %d", block.Header.Version)
	}

	if string(block.Header.PrevBlockHash) != string(prevHash) {
		t.Errorf("Expected prev hash %s, got %s", string(prevHash), string(block.Header.PrevBlockHash))
	}

	if block.Header.Height != height {
		t.Errorf("Expected height %d, got %d", height, block.Header.Height)
	}

	if block.Header.Difficulty != difficulty {
		t.Errorf("Expected difficulty %d, got %d", difficulty, block.Header.Difficulty)
	}

	if len(block.Transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(block.Transactions))
	}
}

func TestAddTransaction(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1, 1000)

	tx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
		Fee:     10,
	}

	block.AddTransaction(tx)

	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}

	if block.Transactions[0] != tx {
		t.Error("Transaction was not added correctly")
	}
}

func TestCalculateHash(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1, 1000)
	block.Header.Timestamp = time.Unix(1234567890, 0)
	block.Header.Nonce = 42

	hash1 := block.CalculateHash()
	hash2 := block.CalculateHash()

	// Hash should be deterministic
	if string(hash1) != string(hash2) {
		t.Error("Block hash is not deterministic")
	}

	// Hash should be 32 bytes (SHA256)
	if len(hash1) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash1))
	}

	// Different nonce should produce different hash
	block.Header.Nonce = 43
	hash3 := block.CalculateHash()
	if string(hash1) == string(hash3) {
		t.Error("Different nonce should produce different hash")
	}
}

func TestCalculateMerkleRoot(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1, 1000)

	// Empty block should have zero merkle root
	merkleRoot := block.CalculateMerkleRoot()
	if len(merkleRoot) != 32 {
		t.Errorf("Expected merkle root length 32, got %d", len(merkleRoot))
	}

	// Add a transaction
	tx1 := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
		Fee:     10,
	}
	block.AddTransaction(tx1)

	merkleRoot1 := block.CalculateMerkleRoot()
	if string(merkleRoot) == string(merkleRoot1) {
		t.Error("Merkle root should change when adding transaction")
	}

	// Add another transaction
	tx2 := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
		Fee:     20,
	}
	block.AddTransaction(tx2)

	merkleRoot2 := block.CalculateMerkleRoot()
	if string(merkleRoot1) == string(merkleRoot2) {
		t.Error("Merkle root should change when adding second transaction")
	}
}

func TestBlockValidation(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1, 1000)

	// Valid block should pass validation
	if err := block.IsValid(); err != nil {
		t.Errorf("Valid block failed validation: %v", err)
	}

	// Block with nil header should fail
	originalHeader := block.Header
	block.Header = nil
	if err := block.IsValid(); err == nil {
		t.Error("Block with nil header should fail validation")
	}

	// Restore header
	block.Header = originalHeader

	// Block with empty transactions should pass validation
	if err := block.IsValid(); err != nil {
		t.Errorf("Block with empty transactions should pass validation: %v", err)
	}
}

func TestHeaderValidation(t *testing.T) {
	header := &Header{
		Version:       1,
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Now(),
		Difficulty:    1000,
		Nonce:         0,
		Height:        1,
	}

	// Valid header should pass validation
	if err := header.IsValid(); err != nil {
		t.Errorf("Valid header failed validation: %v", err)
	}

	// Header with zero version should fail
	header.Version = 0
	if err := header.IsValid(); err == nil {
		t.Error("Header with zero version should fail validation")
	}

	// Restore version
	header.Version = 1

	// Header with nil prev block hash should fail
	header.PrevBlockHash = nil
	if err := header.IsValid(); err == nil {
		t.Error("Header with nil prev block hash should fail validation")
	}
}

func TestTransactionValidation(t *testing.T) {
	tx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{}, // Empty inputs for coinbase
		Outputs: []*TxOutput{
			{
				Value:        1000, // Coinbase reward
				ScriptPubKey: []byte("coinbase_output"),
			},
		},
		LockTime: 0,
		Fee:      0, // Coinbase has no fee
	}
	tx.Hash = make([]byte, 32) // Set a dummy hash for validation

	// Valid transaction should pass validation
	if err := tx.IsValid(); err != nil {
		t.Errorf("Valid transaction failed validation: %v", err)
	}

	// Transaction with zero version should fail
	tx.Version = 0
	if err := tx.IsValid(); err == nil {
		t.Error("Transaction with zero version should fail validation")
	}

	// Restore version
	tx.Version = 1

	// Transaction with nil inputs should pass (coinbase transaction)
	if err := tx.IsValid(); err != nil {
		t.Errorf("Transaction with nil inputs should pass validation: %v", err)
	}

	// Transaction with nil outputs should fail
	tx.Outputs = nil
	if err := tx.IsValid(); err == nil {
		t.Error("Transaction with nil outputs should fail validation")
	}
}

func TestHexHash(t *testing.T) {
	block := NewBlock([]byte("prev_hash"), 1, 1000)
	hash := block.CalculateHash()
	hexHash := block.HexHash()

	// Hex hash should be a valid hex string
	if len(hexHash) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("Expected hex hash length 64, got %d", len(hexHash))
	}

	// Hex hash should match the calculated hash
	expectedHex := ""
	for _, b := range hash {
		expectedHex += fmt.Sprintf("%02x", b)
	}
	if hexHash != expectedHex {
		t.Errorf("Expected hex hash %s, got %s", expectedHex, hexHash)
	}
}
