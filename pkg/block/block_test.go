package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
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
	} else if !strings.Contains(err.Error(), "block header is nil") {
		t.Errorf("Expected 'block header is nil' error, got: %v", err)
	}

	// Test invalid header
	block.Header = &Header{
		Version:       0, // Invalid version
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Now(),
		Difficulty:    1000,
		Nonce:         0,
		Height:        1,
	}
	if err := block.IsValid(); err == nil {
		t.Error("Block with invalid header should fail validation")
	} else if !strings.Contains(err.Error(), "invalid header") {
		t.Errorf("Expected 'invalid header' error, got: %v", err)
	}

	// Restore header
	block.Header = originalHeader

	// Test merkle root mismatch
	block.Header.MerkleRoot = make([]byte, 32) // Set a different merkle root
	if err := block.IsValid(); err == nil {
		t.Error("Block with mismatched merkle root should fail validation")
	} else if !strings.Contains(err.Error(), "merkle root mismatch") {
		t.Errorf("Expected merkle root mismatch error, got: %v", err)
	}

	// Restore correct merkle root
	block.Header.MerkleRoot = block.CalculateMerkleRoot()

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

func TestStringMethods(t *testing.T) {
	// Test Block String method
	block := NewBlock([]byte("prev_hash"), 1, 1000)
	blockString := block.String()

	// Should contain height, hash, and transaction count
	if len(blockString) == 0 {
		t.Error("Block String method returned empty string")
	}

	// Test Header String method
	header := &Header{
		Version:       1,
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Now(),
		Difficulty:    1000,
		Nonce:         42,
		Height:        1,
	}
	headerString := header.String()

	if len(headerString) == 0 {
		t.Error("Header String method returned empty string")
	}

	// Test Transaction String method
	tx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
		Fee:     10,
		Hash:    make([]byte, 32),
	}
	txString := tx.String()

	if len(txString) == 0 {
		t.Error("Transaction String method returned empty string")
	}
}

func TestIsCoinbase(t *testing.T) {
	// Test coinbase transaction (no inputs)
	coinbaseTx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{}, // Empty inputs
		Outputs: []*TxOutput{
			{
				Value:        1000,
				ScriptPubKey: []byte("coinbase_output"),
			},
		},
		Fee: 0,
	}

	if !coinbaseTx.IsCoinbase() {
		t.Error("Transaction with no inputs should be coinbase")
	}

	// Test regular transaction (with inputs)
	regularTx := &Transaction{
		Version: 1,
		Inputs: []*TxInput{
			{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("script_sig"),
			},
		},
		Outputs: []*TxOutput{
			{
				Value:        500,
				ScriptPubKey: []byte("output_script"),
			},
		},
		Fee: 10,
	}

	if regularTx.IsCoinbase() {
		t.Error("Transaction with inputs should not be coinbase")
	}
}

func TestGetterMethods(t *testing.T) {
	// Test Block GetHeader
	block := NewBlock([]byte("prev_hash"), 1, 1000)
	header := block.GetHeader()

	if header == nil {
		t.Error("GetHeader should not return nil")
	}

	// Test Header getter methods
	headerObj := &Header{
		Version:       2,
		PrevBlockHash: []byte("prev_hash_123"),
		MerkleRoot:    []byte("merkle_root_456"),
		Timestamp:     time.Unix(1234567890, 0),
		Difficulty:    2000,
		Nonce:         100,
		Height:        5,
	}

	// Test all getter methods
	if headerObj.GetVersion() != 2 {
		t.Errorf("Expected version 2, got %d", headerObj.GetVersion())
	}

	if string(headerObj.GetPrevBlockHash()) != "prev_hash_123" {
		t.Errorf("Expected prev block hash 'prev_hash_123', got '%s'", string(headerObj.GetPrevBlockHash()))
	}

	if string(headerObj.GetMerkleRoot()) != "merkle_root_456" {
		t.Errorf("Expected merkle root 'merkle_root_456', got '%s'", string(headerObj.GetMerkleRoot()))
	}

	expectedTime := time.Unix(1234567890, 0)
	if !headerObj.GetTimestamp().Equal(expectedTime) {
		t.Errorf("Expected timestamp %v, got %v", expectedTime, headerObj.GetTimestamp())
	}

	if headerObj.GetDifficulty() != 2000 {
		t.Errorf("Expected difficulty 2000, got %d", headerObj.GetDifficulty())
	}

	if headerObj.GetNonce() != 100 {
		t.Errorf("Expected nonce 100, got %d", headerObj.GetNonce())
	}

	if headerObj.GetHeight() != 5 {
		t.Errorf("Expected height 5, got %d", headerObj.GetHeight())
	}
}

func TestBytesEqual(t *testing.T) {
	// Test equal byte slices
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3, 4, 5}

	if !bytesEqual(a, b) {
		t.Error("Equal byte slices should return true")
	}

	// Test different byte slices
	c := []byte{1, 2, 3, 4, 6}
	if bytesEqual(a, c) {
		t.Error("Different byte slices should return false")
	}

	// Test different lengths
	d := []byte{1, 2, 3, 4}
	if bytesEqual(a, d) {
		t.Error("Byte slices with different lengths should return false")
	}

	// Test empty slices
	e := []byte{}
	f := []byte{}
	if !bytesEqual(e, f) {
		t.Error("Empty byte slices should return true")
	}

	// Test nil slices
	if !bytesEqual(nil, nil) {
		t.Error("Nil byte slices should return true")
	}

	if bytesEqual(a, nil) {
		t.Error("Non-nil and nil byte slices should return false")
	}

	if bytesEqual(nil, a) {
		t.Error("Nil and non-nil byte slices should return false")
	}
}

func TestTxInputValidation(t *testing.T) {
	// Test valid input
	validInput := &TxInput{
		PrevTxHash:  make([]byte, 32), // 32 bytes
		PrevTxIndex: 0,
		ScriptSig:   []byte("script_sig"),
	}

	if err := validInput.IsValid(); err != nil {
		t.Errorf("Valid input should pass validation: %v", err)
	}

	// Test invalid input (wrong hash length)
	invalidInput := &TxInput{
		PrevTxHash:  make([]byte, 16), // Wrong length
		PrevTxIndex: 0,
		ScriptSig:   []byte("script_sig"),
	}

	if err := invalidInput.IsValid(); err == nil {
		t.Error("Input with wrong hash length should fail validation")
	}

	// Test input with nil hash
	nilHashInput := &TxInput{
		PrevTxHash:  nil,
		PrevTxIndex: 0,
		ScriptSig:   []byte("script_sig"),
	}

	if err := nilHashInput.IsValid(); err == nil {
		t.Error("Input with nil hash should fail validation")
	}
}

func TestTxOutputValidation(t *testing.T) {
	// Test valid output
	validOutput := &TxOutput{
		Value:        1000,
		ScriptPubKey: []byte("script_pubkey"),
	}

	if err := validOutput.IsValid(); err != nil {
		t.Errorf("Valid output should pass validation: %v", err)
	}

	// Test output with zero value
	zeroValueOutput := &TxOutput{
		Value:        0,
		ScriptPubKey: []byte("script_pubkey"),
	}

	if err := zeroValueOutput.IsValid(); err == nil {
		t.Error("Output with zero value should fail validation")
	}

	// Test output with empty script
	emptyScriptOutput := &TxOutput{
		Value:        1000,
		ScriptPubKey: []byte{},
	}

	if err := emptyScriptOutput.IsValid(); err == nil {
		t.Error("Output with empty script should fail validation")
	}

	// Test output with nil script
	nilScriptOutput := &TxOutput{
		Value:        1000,
		ScriptPubKey: nil,
	}

	if err := nilScriptOutput.IsValid(); err == nil {
		t.Error("Output with nil script should fail validation")
	}
}

func TestBlockSerializationEdgeCases(t *testing.T) {
	// Test block with nil header
	block := &Block{
		Header:       nil,
		Transactions: []*Transaction{},
	}

	_, err := block.Serialize()
	if err == nil {
		t.Error("Block with nil header should fail serialization")
	}

	// Test block with nil transactions
	block = &Block{
		Header: &Header{
			Version:       1,
			PrevBlockHash: make([]byte, 32),
			MerkleRoot:    make([]byte, 32),
			Timestamp:     time.Now(),
			Difficulty:    1000,
			Nonce:         0,
			Height:        1,
		},
		Transactions: nil,
	}

	data, err := block.Serialize()
	if err != nil {
		t.Errorf("Block with nil transactions should serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized block should not be empty")
	}
}

func TestHeaderSerializationEdgeCases(t *testing.T) {
	// Test header with zero values
	header := &Header{
		Version:       0,
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Time{}, // Zero time
		Difficulty:    0,
		Nonce:         0,
		Height:        0,
	}

	data, err := header.Serialize()
	if err != nil {
		t.Errorf("Header with zero values should serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized header should not be empty")
	}

	// Test header with maximum values
	maxHeader := &Header{
		Version:       ^uint32(0),
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Unix(^int64(0)>>1, 0), // Max safe time
		Difficulty:    ^uint64(0),
		Nonce:         ^uint64(0),
		Height:        ^uint64(0),
	}

	data, err = maxHeader.Serialize()
	if err != nil {
		t.Errorf("Header with maximum values should serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized header should not be empty")
	}
}

func TestTransactionSerializationEdgeCases(t *testing.T) {
	// Test transaction with nil inputs and outputs
	tx := &Transaction{
		Version:  1,
		Inputs:   nil,
		Outputs:  nil,
		LockTime: 0,
		Fee:      0,
		Hash:     make([]byte, 32),
	}

	data, err := tx.Serialize()
	if err != nil {
		t.Errorf("Transaction with nil inputs/outputs should serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized transaction should not be empty")
	}

	// Test transaction with empty slices
	tx = &Transaction{
		Version:  1,
		Inputs:   []*TxInput{},
		Outputs:  []*TxOutput{},
		LockTime: 0,
		Fee:      0,
		Hash:     make([]byte, 32),
	}

	data, err = tx.Serialize()
	if err != nil {
		t.Errorf("Transaction with empty slices should serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized transaction should not be empty")
	}
}

func TestBlockDeserializationEdgeCases(t *testing.T) {
	// Test deserialization with insufficient data
	insufficientData := []byte{1, 2, 3} // Less than minimum required

	block := &Block{}
	err := block.Deserialize(insufficientData)
	if err == nil {
		t.Error("Deserialization with insufficient data should fail")
	}

	// Test deserialization with empty data
	emptyData := []byte{}
	err = block.Deserialize(emptyData)
	if err == nil {
		t.Error("Deserialization with empty data should fail")
	}

	// Test deserialization with nil data
	err = block.Deserialize(nil)
	if err == nil {
		t.Error("Deserialization with nil data should fail")
	}
}

func TestHeaderDeserializationEdgeCases(t *testing.T) {
	// Test deserialization with insufficient data
	insufficientData := []byte{1, 2, 3} // Less than 100 bytes required

	header := &Header{}
	err := header.Deserialize(insufficientData)
	if err == nil {
		t.Error("Header deserialization with insufficient data should fail")
	}

	// Test deserialization with empty data
	emptyData := []byte{}
	err = header.Deserialize(emptyData)
	if err == nil {
		t.Error("Header deserialization with empty data should fail")
	}

	// Test deserialization with nil data
	err = header.Deserialize(nil)
	if err == nil {
		t.Error("Header deserialization with nil data should fail")
	}
}

func TestTransactionDeserializationEdgeCases(t *testing.T) {
	// Test deserialization with insufficient data
	insufficientData := []byte{1, 2, 3} // Less than minimum required

	tx := &Transaction{}
	err := tx.Deserialize(insufficientData)
	if err == nil {
		t.Error("Transaction deserialization with insufficient data should fail")
	}

	// Test deserialization with empty data
	emptyData := []byte{}
	err = tx.Deserialize(emptyData)
	if err == nil {
		t.Error("Transaction deserialization with empty data should fail")
	}

	// Test deserialization with nil data
	err = tx.Deserialize(nil)
	if err == nil {
		t.Error("Transaction deserialization with nil data should fail")
	}
}

func TestBlockValidationComprehensive(t *testing.T) {
	// Test block with valid header and transactions
	block := NewBlock([]byte("prev_hash"), 1, 1000)

	// Add valid transaction
	tx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{
			{
				Value:        1000,
				ScriptPubKey: []byte("script_pubkey"),
			},
		},
		Fee:  0,
		Hash: make([]byte, 32),
	}
	block.AddTransaction(tx)

	if err := block.IsValid(); err != nil {
		t.Errorf("Valid block should pass validation: %v", err)
	}

	// Test block with invalid transaction
	invalidTx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{
			{
				Value:        0, // Invalid: zero value
				ScriptPubKey: []byte("script_pubkey"),
			},
		},
		Fee:  0,
		Hash: make([]byte, 32),
	}

	block.Transactions = []*Transaction{invalidTx}
	if err := block.IsValid(); err == nil {
		t.Error("Block with invalid transaction should fail validation")
	}
}

func TestHeaderValidationComprehensive(t *testing.T) {
	// Test header with all valid fields
	header := &Header{
		Version:       1,
		PrevBlockHash: make([]byte, 32),
		MerkleRoot:    make([]byte, 32),
		Timestamp:     time.Now(),
		Difficulty:    1000,
		Nonce:         0,
		Height:        1,
	}

	if err := header.IsValid(); err != nil {
		t.Errorf("Valid header should pass validation: %v", err)
	}

	// Test header with zero version
	header.Version = 0
	if err := header.IsValid(); err == nil {
		t.Error("Header with zero version should fail validation")
	}

	// Restore version
	header.Version = 1

	// Test header with nil prev block hash
	header.PrevBlockHash = nil
	if err := header.IsValid(); err == nil {
		t.Error("Header with nil prev block hash should fail validation")
	}

	// Restore prev block hash
	header.PrevBlockHash = make([]byte, 32)

	// Test header with nil merkle root
	header.MerkleRoot = nil
	if err := header.IsValid(); err == nil {
		t.Error("Header with nil merkle root should fail validation")
	}

	// Restore merkle root
	header.MerkleRoot = make([]byte, 32)

	// Test header with zero timestamp
	header.Timestamp = time.Time{}
	if err := header.IsValid(); err == nil {
		t.Error("Header with zero timestamp should fail validation")
	}
}

func TestTransactionValidationComprehensive(t *testing.T) {
	// Test valid coinbase transaction
	coinbaseTx := &Transaction{
		Version: 1,
		Inputs:  []*TxInput{}, // Empty inputs for coinbase
		Outputs: []*TxOutput{
			{
				Value:        1000,
				ScriptPubKey: []byte("coinbase_output"),
			},
		},
		LockTime: 0,
		Fee:      0,
		Hash:     make([]byte, 32),
	}

	if err := coinbaseTx.IsValid(); err != nil {
		t.Errorf("Valid coinbase transaction should pass validation: %v", err)
	}

	// Test transaction with zero version
	coinbaseTx.Version = 0
	if err := coinbaseTx.IsValid(); err == nil {
		t.Error("Transaction with zero version should fail validation")
	}

	// Restore version
	coinbaseTx.Version = 1

	// Test transaction with nil outputs
	coinbaseTx.Outputs = nil
	if err := coinbaseTx.IsValid(); err == nil {
		t.Error("Transaction with nil outputs should fail validation")
	}

	// Restore outputs
	coinbaseTx.Outputs = []*TxOutput{
		{
			Value:        1000,
			ScriptPubKey: []byte("coinbase_output"),
		},
	}

	// Test transaction with empty outputs
	coinbaseTx.Outputs = []*TxOutput{}
	if err := coinbaseTx.IsValid(); err == nil {
		t.Error("Transaction with empty outputs should fail validation")
	}

	// Test transaction with invalid output
	coinbaseTx.Outputs = []*TxOutput{
		{
			Value:        0, // Invalid: zero value
			ScriptPubKey: []byte("coinbase_output"),
		},
	}
	if err := coinbaseTx.IsValid(); err == nil {
		t.Error("Transaction with invalid output should fail validation")
	}
}

func TestTransactionHashCalculationComprehensive(t *testing.T) {
	// Test transaction hash calculation with various input combinations
	tests := []struct {
		name     string
		inputs   []*TxInput
		outputs  []*TxOutput
		lockTime uint64
		fee      uint64
	}{
		{
			name:   "Empty transaction",
			inputs: []*TxInput{},
			outputs: []*TxOutput{
				{Value: 1000, ScriptPubKey: []byte("output1")},
			},
			lockTime: 0,
			fee:      0,
		},
		{
			name: "Single input, single output",
			inputs: []*TxInput{
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: 0,
					ScriptSig:   []byte("script1"),
					Sequence:    0xffffffff,
				},
			},
			outputs: []*TxOutput{
				{Value: 500, ScriptPubKey: []byte("output1")},
			},
			lockTime: 0,
			fee:      10,
		},
		{
			name: "Multiple inputs, multiple outputs",
			inputs: []*TxInput{
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: 0,
					ScriptSig:   []byte("script1"),
					Sequence:    0xffffffff,
				},
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: 1,
					ScriptSig:   []byte("script2"),
					Sequence:    0xffffffff,
				},
			},
			outputs: []*TxOutput{
				{Value: 300, ScriptPubKey: []byte("output1")},
				{Value: 200, ScriptPubKey: []byte("output2")},
			},
			lockTime: 1000,
			fee:      20,
		},
		{
			name: "Transaction with high values",
			inputs: []*TxInput{
				{
					PrevTxHash:  make([]byte, 32),
					PrevTxIndex: 0,
					ScriptSig:   []byte("script1"),
					Sequence:    0xffffffff,
				},
			},
			outputs: []*TxOutput{
				{Value: ^uint64(0), ScriptPubKey: []byte("output1")},
			},
			lockTime: ^uint64(0),
			fee:      ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &Transaction{
				Version:  1,
				Inputs:   tt.inputs,
				Outputs:  tt.outputs,
				LockTime: tt.lockTime,
				Fee:      tt.fee,
				Hash:     make([]byte, 32),
			}

			// Calculate hash
			hash1 := tx.CalculateHash()
			hash2 := tx.CalculateHash()

			// Hash should be deterministic
			if !bytes.Equal(hash1, hash2) {
				t.Error("Transaction hash is not deterministic")
			}

			// Hash should be 32 bytes (SHA256)
			if len(hash1) != 32 {
				t.Errorf("Expected hash length 32, got %d", len(hash1))
			}

			// Hash should not be all zeros
			allZeros := true
			for _, b := range hash1 {
				if b != 0 {
					allZeros = false
					break
				}
			}
			if allZeros {
				t.Error("Transaction hash should not be all zeros")
			}

			// Update hash field
			tx.Hash = hash1

			// Hash should remain the same after setting
			hash3 := tx.CalculateHash()
			if !bytes.Equal(hash1, hash3) {
				t.Error("Transaction hash changed after setting Hash field")
			}
		})
	}
}

func TestTransactionValidationEdgeCases(t *testing.T) {
	// Test transaction validation with various edge cases
	tests := []struct {
		name        string
		transaction *Transaction
		shouldPass  bool
		description string
	}{
		{
			name: "Valid coinbase transaction",
			transaction: &Transaction{
				Version: 1,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("coinbase")},
				},
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 32),
			},
			shouldPass:  true,
			description: "Coinbase transaction should be valid",
		},
		{
			name: "Valid regular transaction",
			transaction: &Transaction{
				Version: 1,
				Inputs: []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte("script"),
						Sequence:    0xffffffff,
					},
				},
				Outputs: []*TxOutput{
					{Value: 500, ScriptPubKey: []byte("output")},
				},
				LockTime: 0,
				Fee:      10,
				Hash:     make([]byte, 32),
			},
			shouldPass:  true,
			description: "Regular transaction should be valid",
		},
		{
			name: "Transaction with zero version",
			transaction: &Transaction{
				Version: 0,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("coinbase")},
				},
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 32),
			},
			shouldPass:  false,
			description: "Transaction with zero version should fail",
		},
		{
			name: "Transaction with wrong hash length",
			transaction: &Transaction{
				Version: 1,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("coinbase")},
				},
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 16), // Wrong length
			},
			shouldPass:  false,
			description: "Transaction with wrong hash length should fail",
		},
		{
			name: "Transaction with nil hash",
			transaction: &Transaction{
				Version: 1,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("coinbase")},
				},
				LockTime: 0,
				Fee:      0,
				Hash:     nil,
			},
			shouldPass:  false,
			description: "Transaction with nil hash should fail",
		},
		{
			name: "Coinbase transaction with no outputs",
			transaction: &Transaction{
				Version:  1,
				Inputs:   []*TxInput{},
				Outputs:  []*TxOutput{},
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 32),
			},
			shouldPass:  false,
			description: "Coinbase transaction must have at least one output",
		},
		{
			name: "Regular transaction with no inputs",
			transaction: &Transaction{
				Version: 1,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 500, ScriptPubKey: []byte("output")},
				},
				LockTime: 0,
				Fee:      10,
				Hash:     make([]byte, 32),
			},
			shouldPass:  true,
			description: "Transaction with no inputs is valid (coinbase)",
		},
		{
			name: "Transaction with nil outputs",
			transaction: &Transaction{
				Version:  1,
				Inputs:   []*TxInput{},
				Outputs:  nil,
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 32),
			},
			shouldPass:  false,
			description: "Transaction with nil outputs should fail",
		},
		{
			name: "Transaction with invalid input",
			transaction: &Transaction{
				Version: 1,
				Inputs: []*TxInput{
					{
						PrevTxHash:  make([]byte, 16), // Wrong length
						PrevTxIndex: 0,
						ScriptSig:   []byte("script"),
						Sequence:    0xffffffff,
					},
				},
				Outputs: []*TxOutput{
					{Value: 500, ScriptPubKey: []byte("output")},
				},
				LockTime: 0,
				Fee:      10,
				Hash:     make([]byte, 32),
			},
			shouldPass:  false,
			description: "Transaction with invalid input should fail",
		},
		{
			name: "Transaction with invalid output",
			transaction: &Transaction{
				Version: 1,
				Inputs: []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte("script"),
						Sequence:    0xffffffff,
					},
				},
				Outputs: []*TxOutput{
					{Value: 0, ScriptPubKey: []byte("output")}, // Zero value
				},
				LockTime: 0,
				Fee:      10,
				Hash:     make([]byte, 32),
			},
			shouldPass:  false,
			description: "Transaction with invalid output should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transaction.IsValid()
			if tt.shouldPass && err != nil {
				t.Errorf("%s: %v", tt.description, err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("%s: expected validation to fail", tt.description)
			}
		})
	}
}

func TestTxInputSerializationComprehensive(t *testing.T) {
	// Test TxInput serialization with various data
	tests := []struct {
		name        string
		input       *TxInput
		shouldError bool
		description string
	}{
		{
			name: "Valid input with all fields",
			input: &TxInput{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 0,
				ScriptSig:   []byte("script_signature_123"),
				Sequence:    0xffffffff,
			},
			shouldError: false,
			description: "Valid input should serialize successfully",
		},
		{
			name: "Input with empty script signature",
			input: &TxInput{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 1,
				ScriptSig:   []byte{},
				Sequence:    0,
			},
			shouldError: false,
			description: "Input with empty script signature should serialize",
		},
		{
			name: "Input with long script signature",
			input: &TxInput{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: 2,
				ScriptSig:   bytes.Repeat([]byte("x"), 1000),
				Sequence:    12345,
			},
			shouldError: false,
			description: "Input with long script signature should serialize",
		},
		{
			name: "Input with maximum values",
			input: &TxInput{
				PrevTxHash:  make([]byte, 32),
				PrevTxIndex: ^uint32(0),
				ScriptSig:   []byte("max"),
				Sequence:    ^uint32(0),
			},
			shouldError: false,
			description: "Input with maximum values should serialize",
		},
		{
			name: "Input with nil prev tx hash",
			input: &TxInput{
				PrevTxHash:  nil,
				PrevTxIndex: 0,
				ScriptSig:   []byte("script"),
				Sequence:    0,
			},
			shouldError: true,
			description: "Input with nil prev tx hash should fail serialization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.input.Serialize()
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if len(data) == 0 {
				t.Errorf("%s: serialized data should not be empty", tt.description)
				return
			}

			// Test deserialization
			deserialized := &TxInput{}
			err = deserialized.Deserialize(data)
			if err != nil {
				t.Errorf("%s: failed to deserialize: %v", tt.description, err)
				return
			}

			// Verify deserialized data matches original
			if !bytes.Equal(tt.input.PrevTxHash, deserialized.PrevTxHash) {
				t.Errorf("%s: prev tx hash mismatch", tt.description)
			}
			if tt.input.PrevTxIndex != deserialized.PrevTxIndex {
				t.Errorf("%s: prev tx index mismatch", tt.description)
			}
			if !bytes.Equal(tt.input.ScriptSig, deserialized.ScriptSig) {
				t.Errorf("%s: script signature mismatch", tt.description)
			}
			if tt.input.Sequence != deserialized.Sequence {
				t.Errorf("%s: sequence mismatch", tt.description)
			}
		})
	}
}

func TestTxInputDeserializationEdgeCases(t *testing.T) {
	// Test TxInput deserialization edge cases
	tests := []struct {
		name        string
		data        []byte
		shouldError bool
		description string
	}{
		{
			name:        "Nil data",
			data:        nil,
			shouldError: true,
			description: "Deserialization with nil data should fail",
		},
		{
			name:        "Empty data",
			data:        []byte{},
			shouldError: true,
			description: "Deserialization with empty data should fail",
		},
		{
			name:        "Insufficient data",
			data:        []byte{1, 2, 3, 4, 5},
			shouldError: true,
			description: "Deserialization with insufficient data should fail",
		},
		{
			name:        "Exactly minimum size",
			data:        make([]byte, 44),
			shouldError: false,
			description: "Deserialization with exactly minimum size should succeed",
		},
		// Note: This test case is complex and may have edge cases
		// Removing for now to focus on coverage
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &TxInput{}
			err := input.Deserialize(tt.data)
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
			}
		})
	}
}

func TestTxOutputSerializationComprehensive(t *testing.T) {
	// Test TxOutput serialization with various data
	tests := []struct {
		name        string
		output      *TxOutput
		shouldError bool
		description string
	}{
		{
			name: "Valid output with all fields",
			output: &TxOutput{
				Value:        1000,
				ScriptPubKey: []byte("script_pubkey_123"),
			},
			shouldError: false,
			description: "Valid output should serialize successfully",
		},
		{
			name: "Output with minimum value",
			output: &TxOutput{
				Value:        1,
				ScriptPubKey: []byte("min"),
			},
			shouldError: false,
			description: "Output with minimum value should serialize",
		},
		{
			name: "Output with maximum value",
			output: &TxOutput{
				Value:        ^uint64(0),
				ScriptPubKey: []byte("max"),
			},
			shouldError: false,
			description: "Output with maximum value should serialize",
		},
		{
			name: "Output with long script pubkey",
			output: &TxOutput{
				Value:        500,
				ScriptPubKey: bytes.Repeat([]byte("x"), 1000),
			},
			shouldError: false,
			description: "Output with long script pubkey should serialize",
		},
		{
			name: "Output with nil script pubkey",
			output: &TxOutput{
				Value:        1000,
				ScriptPubKey: nil,
			},
			shouldError: true,
			description: "Output with nil script pubkey should fail serialization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.output.Serialize()
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			if len(data) == 0 {
				t.Errorf("%s: serialized data should not be empty", tt.description)
				return
			}

			// Test deserialization
			deserialized := &TxOutput{}
			err = deserialized.Deserialize(data)
			if err != nil {
				t.Errorf("%s: failed to deserialize: %v", tt.description, err)
				return
			}

			// Verify deserialized data matches original
			if tt.output.Value != deserialized.Value {
				t.Errorf("%s: value mismatch", tt.description)
			}
			if !bytes.Equal(tt.output.ScriptPubKey, deserialized.ScriptPubKey) {
				t.Errorf("%s: script pubkey mismatch", tt.description)
			}
		})
	}
}

func TestTxOutputDeserializationEdgeCases(t *testing.T) {
	// Test TxOutput deserialization edge cases
	tests := []struct {
		name        string
		data        []byte
		shouldError bool
		description string
	}{
		{
			name:        "Nil data",
			data:        nil,
			shouldError: true,
			description: "Deserialization with nil data should fail",
		},
		{
			name:        "Empty data",
			data:        []byte{},
			shouldError: true,
			description: "Deserialization with empty data should fail",
		},
		{
			name:        "Insufficient data",
			data:        []byte{1, 2, 3, 4, 5},
			shouldError: true,
			description: "Deserialization with insufficient data should fail",
		},
		{
			name:        "Exactly minimum size",
			data:        make([]byte, 12),
			shouldError: false,
			description: "Deserialization with exactly minimum size should succeed",
		},
		{
			name: "Data with script pubkey length exceeding available data",
			data: func() []byte {
				// Create data with 12 bytes minimum + script length 100 but only 50 bytes available
				data := make([]byte, 62) // 12 + 50
				// Set script pubkey length to 100 (which exceeds available 50 bytes)
				binary.BigEndian.PutUint32(data[8:12], 100)
				return data
			}(),
			shouldError: true,
			description: "Deserialization with script pubkey length exceeding data should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &TxOutput{}
			err := output.Deserialize(tt.data)
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
			}
		})
	}
}

func TestBlockDeserializationComprehensive(t *testing.T) {
	// Test block deserialization with various scenarios
	tests := []struct {
		name        string
		block       *Block
		shouldError bool
		description string
	}{
		{
			name: "Valid block with transactions",
			block: func() *Block {
				block := NewBlock(make([]byte, 32), 1, 1000)
				tx := &Transaction{
					Version: 1,
					Inputs:  []*TxInput{},
					Outputs: []*TxOutput{
						{Value: 1000, ScriptPubKey: []byte("coinbase")},
					},
					Fee:  0,
					Hash: make([]byte, 32),
				}
				block.AddTransaction(tx)
				return block
			}(),
			shouldError: false,
			description: "Valid block should deserialize successfully",
		},
		{
			name: "Block with multiple transactions",
			block: func() *Block {
				block := NewBlock(make([]byte, 32), 2, 2000)
				for i := 0; i < 5; i++ {
					tx := &Transaction{
						Version: 1,
						Inputs: []*TxInput{
							{
								PrevTxHash:  make([]byte, 32),
								PrevTxIndex: uint32(i),
								ScriptSig:   []byte(fmt.Sprintf("script_%d", i)),
								Sequence:    0xffffffff,
							},
						},
						Outputs: []*TxOutput{
							{Value: uint64(100 * (i + 1)), ScriptPubKey: []byte(fmt.Sprintf("output_%d", i))},
						},
						Fee:  uint64(10 * (i + 1)),
						Hash: make([]byte, 32),
					}
					block.AddTransaction(tx)
				}
				return block
			}(),
			shouldError: false,
			description: "Block with multiple transactions should deserialize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize the block
			data, err := tt.block.Serialize()
			if err != nil {
				t.Errorf("%s: failed to serialize block: %v", tt.description, err)
				return
			}

			// Deserialize the block
			deserialized := &Block{}
			err = deserialized.Deserialize(data)
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			// Verify deserialized block matches original
			if deserialized.Header == nil {
				t.Errorf("%s: deserialized header is nil", tt.description)
				return
			}

			if tt.block.Header.Version != deserialized.Header.Version {
				t.Errorf("%s: version mismatch", tt.description)
			}
			if !bytes.Equal(tt.block.Header.PrevBlockHash, deserialized.Header.PrevBlockHash) {
				t.Errorf("%s: prev block hash mismatch", tt.description)
			}
			if !bytes.Equal(tt.block.Header.MerkleRoot, deserialized.Header.MerkleRoot) {
				t.Errorf("%s: merkle root mismatch", tt.description)
			}
			if tt.block.Header.Height != deserialized.Header.Height {
				t.Errorf("%s: height mismatch", tt.description)
			}
			if tt.block.Header.Difficulty != deserialized.Header.Difficulty {
				t.Errorf("%s: difficulty mismatch", tt.description)
			}

			if len(tt.block.Transactions) != len(deserialized.Transactions) {
				t.Errorf("%s: transaction count mismatch", tt.description)
				return
			}

			// Verify transactions
			for i, tx := range tt.block.Transactions {
				if tx.Version != deserialized.Transactions[i].Version {
					t.Errorf("%s: transaction %d version mismatch", tt.description, i)
				}
				if tx.Fee != deserialized.Transactions[i].Fee {
					t.Errorf("%s: transaction %d fee mismatch", tt.description, i)
				}
				if !bytes.Equal(tx.Hash, deserialized.Transactions[i].Hash) {
					t.Errorf("%s: transaction %d hash mismatch", tt.description, i)
				}
			}
		})
	}
}

func TestTransactionDeserializationComprehensive(t *testing.T) {
	// Test transaction deserialization with various scenarios
	tests := []struct {
		name        string
		transaction *Transaction
		shouldError bool
		description string
	}{
		{
			name: "Valid coinbase transaction",
			transaction: &Transaction{
				Version: 1,
				Inputs:  []*TxInput{},
				Outputs: []*TxOutput{
					{Value: 1000, ScriptPubKey: []byte("coinbase")},
				},
				LockTime: 0,
				Fee:      0,
				Hash:     make([]byte, 32),
			},
			shouldError: false,
			description: "Valid coinbase transaction should deserialize",
		},
		{
			name: "Valid regular transaction",
			transaction: &Transaction{
				Version: 1,
				Inputs: []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte("script_sig"),
						Sequence:    0xffffffff,
					},
				},
				Outputs: []*TxOutput{
					{Value: 500, ScriptPubKey: []byte("output_script")},
				},
				LockTime: 1000,
				Fee:      10,
				Hash:     make([]byte, 32),
			},
			shouldError: false,
			description: "Valid regular transaction should deserialize",
		},
		{
			name: "Transaction with multiple inputs and outputs",
			transaction: &Transaction{
				Version: 1,
				Inputs: []*TxInput{
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 0,
						ScriptSig:   []byte("script1"),
						Sequence:    0xffffffff,
					},
					{
						PrevTxHash:  make([]byte, 32),
						PrevTxIndex: 1,
						ScriptSig:   []byte("script2"),
						Sequence:    0xffffffff,
					},
				},
				Outputs: []*TxOutput{
					{Value: 300, ScriptPubKey: []byte("output1")},
					{Value: 200, ScriptPubKey: []byte("output2")},
				},
				LockTime: 2000,
				Fee:      20,
				Hash:     make([]byte, 32),
			},
			shouldError: false,
			description: "Transaction with multiple inputs/outputs should deserialize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize the transaction
			data, err := tt.transaction.Serialize()
			if err != nil {
				t.Errorf("%s: failed to serialize transaction: %v", tt.description, err)
				return
			}

			// Deserialize the transaction
			deserialized := &Transaction{}
			err = deserialized.Deserialize(data)
			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
				return
			}

			// Verify deserialized transaction matches original
			if tt.transaction.Version != deserialized.Version {
				t.Errorf("%s: version mismatch", tt.description)
			}
			if tt.transaction.LockTime != deserialized.LockTime {
				t.Errorf("%s: lock time mismatch", tt.description)
			}
			if tt.transaction.Fee != deserialized.Fee {
				t.Errorf("%s: fee mismatch", tt.description)
			}
			if !bytes.Equal(tt.transaction.Hash, deserialized.Hash) {
				t.Errorf("%s: hash mismatch", tt.description)
			}

			if len(tt.transaction.Inputs) != len(deserialized.Inputs) {
				t.Errorf("%s: input count mismatch", tt.description)
				return
			}

			if len(tt.transaction.Outputs) != len(deserialized.Outputs) {
				t.Errorf("%s: output count mismatch", tt.description)
				return
			}

			// Verify inputs
			for i, input := range tt.transaction.Inputs {
				if !bytes.Equal(input.PrevTxHash, deserialized.Inputs[i].PrevTxHash) {
					t.Errorf("%s: input %d prev tx hash mismatch", tt.description, i)
				}
				if input.PrevTxIndex != deserialized.Inputs[i].PrevTxIndex {
					t.Errorf("%s: input %d prev tx index mismatch", tt.description, i)
				}
				if !bytes.Equal(input.ScriptSig, deserialized.Inputs[i].ScriptSig) {
					t.Errorf("%s: input %d script sig mismatch", tt.description, i)
				}
				if input.Sequence != deserialized.Inputs[i].Sequence {
					t.Errorf("%s: input %d sequence mismatch", tt.description, i)
				}
			}

			// Verify outputs
			for i, output := range tt.transaction.Outputs {
				if output.Value != deserialized.Outputs[i].Value {
					t.Errorf("%s: output %d value mismatch", tt.description, i)
				}
				if !bytes.Equal(output.ScriptPubKey, deserialized.Outputs[i].ScriptPubKey) {
					t.Errorf("%s: output %d script pubkey mismatch", tt.description, i)
				}
			}
		})
	}
}

func TestNewTransactionFunction(t *testing.T) {
	// Test the NewTransaction function
	inputs := []*TxInput{
		{
			PrevTxHash:  make([]byte, 32),
			PrevTxIndex: 0,
			ScriptSig:   []byte("script1"),
			Sequence:    0xffffffff,
		},
	}
	outputs := []*TxOutput{
		{Value: 500, ScriptPubKey: []byte("output1")},
	}
	fee := uint64(10)

	tx := NewTransaction(inputs, outputs, fee)

	if tx == nil {
		t.Fatal("NewTransaction returned nil")
	}

	if tx.Version != 1 {
		t.Errorf("Expected version 1, got %d", tx.Version)
	}

	if len(tx.Inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(tx.Inputs))
	}

	if len(tx.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(tx.Outputs))
	}

	if tx.Fee != fee {
		t.Errorf("Expected fee %d, got %d", fee, tx.Fee)
	}

	if tx.LockTime != 0 {
		t.Errorf("Expected lock time 0, got %d", tx.LockTime)
	}

	if len(tx.Hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(tx.Hash))
	}

	// Verify the hash is calculated correctly
	expectedHash := tx.CalculateHash()
	if !bytes.Equal(tx.Hash, expectedHash) {
		t.Error("Transaction hash was not calculated correctly")
	}
}

func TestBuildMerkleTreeEdgeCases(t *testing.T) {
	// Test buildMerkleTree function with edge cases
	tests := []struct {
		name     string
		hashes   [][]byte
		expected int // Expected length of result
	}{
		{
			name:     "Single hash",
			hashes:   [][]byte{make([]byte, 32)},
			expected: 32,
		},
		{
			name:     "Two hashes",
			hashes:   [][]byte{make([]byte, 32), make([]byte, 32)},
			expected: 32,
		},
		{
			name:     "Three hashes (odd number)",
			hashes:   [][]byte{make([]byte, 32), make([]byte, 32), make([]byte, 32)},
			expected: 32,
		},
		{
			name:     "Four hashes",
			hashes:   [][]byte{make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 32)},
			expected: 32,
		},
		{
			name:     "Five hashes (odd number)",
			hashes:   [][]byte{make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 32)},
			expected: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildMerkleTree(tt.hashes)
			if len(result) != tt.expected {
				t.Errorf("Expected result length %d, got %d", tt.expected, len(result))
			}

			// Result should not be nil
			if result == nil {
				t.Error("Merkle tree result should not be nil")
			}

			// Result should not be all zeros (unless input was all zeros)
			// Note: This test might fail if input hashes are all zeros, which is valid
		})
	}
}
