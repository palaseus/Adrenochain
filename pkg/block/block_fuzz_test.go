//go:build go1.18

package block

import (
	"bytes"
	"testing"
)

// FuzzBlockSerialization tests block serialization/deserialization with fuzzed data
func FuzzBlockSerialization(f *testing.F) {
	// Seed corpus with valid blocks
	validBlock := NewBlock([]byte("prevhash"), 1, 1000)
	validTx := NewTransaction([]*TxInput{}, []*TxOutput{
		{Value: 100, ScriptPubKey: []byte("script")},
	}, 0)
	validBlock.AddTransaction(validTx)
	validData, _ := validBlock.Serialize()
	f.Add(validData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs to avoid memory issues
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		// Try to deserialize the fuzzed data
		block := &Block{}
		err := block.Deserialize(data)
		if err != nil {
			// If deserialization fails, that's expected for invalid data
			return
		}

		// If deserialization succeeds, try to serialize it back
		serialized, err := block.Serialize()
		if err != nil {
			t.Errorf("Failed to serialize deserialized block: %v", err)
			return
		}

		// Check that serialization is deterministic
		serialized2, err := block.Serialize()
		if err != nil {
			t.Errorf("Failed to serialize block second time: %v", err)
			return
		}

		if !bytes.Equal(serialized, serialized2) {
			t.Errorf("Serialization is not deterministic")
		}

		// Test that we can deserialize the serialized data back to the same block
		block2 := &Block{}
		err = block2.Deserialize(serialized)
		if err != nil {
			t.Errorf("Failed to deserialize serialized block: %v", err)
			return
		}

		// Compare the blocks (excluding timestamp which may change)
		if block.Header.Version != block2.Header.Version {
			t.Errorf("Version mismatch after round-trip: %d != %d", block.Header.Version, block2.Header.Version)
		}
		if !bytes.Equal(block.Header.PrevBlockHash, block2.Header.PrevBlockHash) {
			t.Errorf("PrevBlockHash mismatch after round-trip")
		}
		if !bytes.Equal(block.Header.MerkleRoot, block2.Header.MerkleRoot) {
			t.Errorf("MerkleRoot mismatch after round-trip")
		}
		if block.Header.Difficulty != block2.Header.Difficulty {
			t.Errorf("Difficulty mismatch after round-trip: %d != %d", block.Header.Difficulty, block2.Header.Difficulty)
		}
		if block.Header.Height != block2.Header.Height {
			t.Errorf("Height mismatch after round-trip: %d != %d", block.Header.Height, block2.Header.Height)
		}
		if block.Header.Nonce != block2.Header.Nonce {
			t.Errorf("Nonce mismatch after round-trip: %d != %d", block.Header.Nonce, block2.Header.Nonce)
		}
		if len(block.Transactions) != len(block2.Transactions) {
			t.Errorf("Transaction count mismatch after round-trip: %d != %d", len(block.Transactions), len(block2.Transactions))
		}
	})
}

// FuzzBlockValidation tests block validation with fuzzed data
func FuzzBlockValidation(f *testing.F) {
	// Seed corpus with valid blocks
	validBlock := NewBlock([]byte("prevhash"), 1, 1000)
	validTx := NewTransaction([]*TxInput{}, []*TxOutput{
		{Value: 100, ScriptPubKey: []byte("script")},
	}, 0)
	validBlock.AddTransaction(validTx)
	validData, _ := validBlock.Serialize()
	f.Add(validData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		block := &Block{}
		err := block.Deserialize(data)
		if err != nil {
			return
		}

		// Test validation methods
		if err := block.IsValid(); err != nil {
			// Invalid blocks should fail validation
			return
		}

		// Test header validation
		if err := block.Header.IsValid(); err != nil {
			t.Errorf("Valid block has invalid header: %v", err)
		}

		// Test transaction validation
		for _, tx := range block.Transactions {
			if err := tx.IsValid(); err != nil {
				t.Errorf("Valid block contains invalid transaction: %v", err)
			}
		}
	})
}

// FuzzTransactionSerialization tests transaction serialization/deserialization with fuzzed data
func FuzzTransactionSerialization(f *testing.F) {
	// Seed corpus with valid transactions
	validTx := NewTransaction([]*TxInput{}, []*TxOutput{
		{Value: 100, ScriptPubKey: []byte("script")},
	}, 0)
	validData, _ := validTx.Serialize()
	f.Add(validData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		tx := &Transaction{}
		err := tx.Deserialize(data)
		if err != nil {
			return
		}

		// Try to serialize it back
		serialized, err := tx.Serialize()
		if err != nil {
			t.Errorf("Failed to serialize deserialized transaction: %v", err)
			return
		}

		// Check determinism
		serialized2, err := tx.Serialize()
		if err != nil {
			t.Errorf("Failed to serialize transaction second time: %v", err)
			return
		}

		if !bytes.Equal(serialized, serialized2) {
			t.Errorf("Transaction serialization is not deterministic")
		}

		// Test that we can deserialize the serialized data back to the same transaction
		tx2 := &Transaction{}
		err = tx2.Deserialize(serialized)
		if err != nil {
			t.Errorf("Failed to deserialize serialized transaction: %v", err)
			return
		}

		// Compare the transactions (excluding hash which may change)
		if tx.Version != tx2.Version {
			t.Errorf("Version mismatch after round-trip: %d != %d", tx.Version, tx2.Version)
		}
		if len(tx.Inputs) != len(tx2.Inputs) {
			t.Errorf("Input count mismatch after round-trip: %d != %d", len(tx.Inputs), len(tx2.Inputs))
		}
		if len(tx.Outputs) != len(tx2.Outputs) {
			t.Errorf("Output count mismatch after round-trip: %d != %d", len(tx.Outputs), len(tx2.Outputs))
		}
		if tx.LockTime != tx2.LockTime {
			t.Errorf("LockTime mismatch after round-trip: %d != %d", tx.LockTime, tx2.LockTime)
		}
		if tx.Fee != tx2.Fee {
			t.Errorf("Fee mismatch after round-trip: %d != %d", tx.Fee, tx2.Fee)
		}
	})
}

// FuzzMerkleRootCalculation tests merkle root calculation with fuzzed transaction data
func FuzzMerkleRootCalculation(f *testing.F) {
	// Seed corpus with valid transaction hashes
	validTx := NewTransaction([]*TxInput{}, []*TxOutput{
		{Value: 100, ScriptPubKey: []byte("script")},
	}, 0)
	validHash := validTx.CalculateHash()
	f.Add(validHash)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		// Create a block with fuzzed transaction data
		block := NewBlock([]byte("prevhash"), 1, 1000)

		// Create a valid transaction with fuzzed data
		tx := &Transaction{
			Version: 1,
			Inputs:  []*TxInput{},
			Outputs: []*TxOutput{
				{Value: 100, ScriptPubKey: []byte("script")},
			},
			LockTime: 0,
			Fee:      0,
		}

		// Use fuzzed data as transaction hash
		if len(data) >= 32 {
			tx.Hash = data[:32]
		} else {
			tx.Hash = make([]byte, 32)
			copy(tx.Hash, data)
		}

		block.AddTransaction(tx)

		// Calculate merkle root
		merkleRoot := block.CalculateMerkleRoot()

		// Verify merkle root is not nil and has correct length
		if merkleRoot == nil {
			t.Errorf("Merkle root is nil")
			return
		}

		if len(merkleRoot) != 32 {
			t.Errorf("Merkle root has incorrect length: %d", len(merkleRoot))
		}

		// Verify merkle root is deterministic
		merkleRoot2 := block.CalculateMerkleRoot()
		if !bytes.Equal(merkleRoot, merkleRoot2) {
			t.Errorf("Merkle root calculation is not deterministic")
		}
	})
}

// FuzzBlockHashCalculation tests block hash calculation with fuzzed data
func FuzzBlockHashCalculation(f *testing.F) {
	// Seed corpus with valid block data
	validBlock := NewBlock([]byte("prevhash"), 1, 1000)
	validTx := NewTransaction([]*TxInput{}, []*TxOutput{
		{Value: 100, ScriptPubKey: []byte("script")},
	}, 0)
	validBlock.AddTransaction(validTx)
	validData, _ := validBlock.Serialize()
	f.Add(validData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		block := &Block{}
		err := block.Deserialize(data)
		if err != nil {
			return
		}

		// Calculate hash
		hash := block.CalculateHash()

		// Verify hash is not nil and has correct length
		if hash == nil {
			t.Errorf("Block hash is nil")
			return
		}

		if len(hash) != 32 {
			t.Errorf("Block hash has incorrect length: %d", len(hash))
		}

		// Verify hash is deterministic
		hash2 := block.CalculateHash()
		if !bytes.Equal(hash, hash2) {
			t.Errorf("Block hash calculation is not deterministic")
		}

		// Verify hex representation
		hexHash := block.HexHash()
		if len(hexHash) != 64 {
			t.Errorf("Hex hash has incorrect length: %d", len(hexHash))
		}
	})
}

// FuzzHeaderValidation tests header validation with fuzzed data
func FuzzHeaderValidation(f *testing.F) {
	// Seed corpus with valid headers
	validHeader := &Header{
		Version:       1,
		PrevBlockHash: []byte("prevhash"),
		MerkleRoot:    []byte("merkleroot"),
		Difficulty:    1000,
		Height:        1,
		Nonce:         0,
	}
	validData, _ := validHeader.Serialize()
	f.Add(validData)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Skip very large inputs
		if len(data) > 1000000 {
			t.Skip("Input too large")
		}

		header := &Header{}
		err := header.Deserialize(data)
		if err != nil {
			return
		}

		// Test header validation
		if err := header.IsValid(); err != nil {
			// Invalid headers should fail validation
			return
		}

		// Verify header fields are reasonable
		if header.Version == 0 {
			t.Errorf("Header version should not be 0")
		}

		if header.Height < 0 {
			t.Errorf("Header height should not be negative")
		}

		if header.Difficulty == 0 {
			t.Errorf("Header difficulty should not be 0")
		}
	})
}
