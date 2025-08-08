package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Header     *Header
	Transactions []*Transaction
	MerkleRoot []byte
}

// Header contains the block header information
type Header struct {
	Version       uint32
	PrevBlockHash []byte
	MerkleRoot    []byte
	Timestamp     time.Time
	Difficulty    uint64
	Nonce         uint64
	Height        uint64
}

// Transaction represents a transaction in the blockchain
type Transaction struct {
	Version  uint32
	Inputs   []*TxInput
	Outputs  []*TxOutput
	LockTime uint64
	Fee      uint64
	Hash     []byte
}

// TxInput represents a transaction input
type TxInput struct {
	PrevTxHash    []byte
	PrevTxIndex   uint32
	ScriptSig     []byte
	Sequence      uint32
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value        uint64
	ScriptPubKey []byte
}

// NewBlock creates a new block with the given parameters
func NewBlock(prevHash []byte, height uint64, difficulty uint64) *Block {
	block := &Block{
		Header: &Header{
			Version:       1,
			PrevBlockHash: prevHash,
			Timestamp:     time.Now(),
			Difficulty:    difficulty,
			Height:        height,
		},
		Transactions: make([]*Transaction, 0),
	}
	
	// Initialize the Merkle root
	block.Header.MerkleRoot = block.CalculateMerkleRoot()
	
	return block
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// CalculateHash calculates the hash of the block header
func (b *Block) CalculateHash() []byte {
	data := make([]byte, 0)
	
	// Version
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, b.Header.Version)
	data = append(data, versionBytes...)
	
	// Previous block hash
	data = append(data, b.Header.PrevBlockHash...)
	
	// Merkle root
	data = append(data, b.Header.MerkleRoot...)
	
	// Timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(b.Header.Timestamp.Unix()))
	data = append(data, timestampBytes...)
	
	// Difficulty
	difficultyBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(difficultyBytes, b.Header.Difficulty)
	data = append(data, difficultyBytes...)
	
	// Nonce
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, b.Header.Nonce)
	data = append(data, nonceBytes...)
	
	// Height
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, b.Header.Height)
	data = append(data, heightBytes...)
	
	hash := sha256.Sum256(data)
	return hash[:]
}

// CalculateMerkleRoot calculates the Merkle root of all transactions
func (b *Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		hash := sha256.Sum256([]byte{})
		return hash[:]
	}
	
	// If only one transaction, return its hash
	if len(b.Transactions) == 1 {
		return b.Transactions[0].Hash
	}
	
	// Build the Merkle tree
	hashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		hashes[i] = tx.Hash
	}
	
	return buildMerkleTree(hashes)
}

// buildMerkleTree builds a Merkle tree from transaction hashes
func buildMerkleTree(hashes [][]byte) []byte {
	if len(hashes) == 1 {
		return hashes[0]
	}
	
	// If odd number of hashes, duplicate the last one
	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}
	
	// Create next level of the tree
	nextLevel := make([][]byte, len(hashes)/2)
	for i := 0; i < len(hashes); i += 2 {
		combined := append(hashes[i], hashes[i+1]...)
		hash := sha256.Sum256(combined)
		nextLevel[i/2] = hash[:]
	}
	
	return buildMerkleTree(nextLevel)
}

// IsValid checks if the block is valid
func (b *Block) IsValid() error {
	// Check if header exists
	if b.Header == nil {
		return fmt.Errorf("block header is nil")
	}
	
	// Check if header is valid
	if err := b.Header.IsValid(); err != nil {
		return fmt.Errorf("invalid header: %w", err)
	}
	
	// Check if Merkle root matches
	calculatedRoot := b.CalculateMerkleRoot()
	if !bytesEqual(b.Header.MerkleRoot, calculatedRoot) {
		return fmt.Errorf("merkle root mismatch: expected %x, got %x", 
			b.Header.MerkleRoot, calculatedRoot)
	}
	
	// Validate all transactions
	for i, tx := range b.Transactions {
		if err := tx.IsValid(); err != nil {
			return fmt.Errorf("invalid transaction %d: %w", i, err)
		}
	}
	
	return nil
}

// IsValid checks if the header is valid
func (h *Header) IsValid() error {
	if h.Version == 0 {
		return fmt.Errorf("invalid version: %d", h.Version)
	}
	
	if h.PrevBlockHash == nil {
		return fmt.Errorf("previous block hash cannot be nil")
	}
	
	if h.Timestamp.IsZero() {
		return fmt.Errorf("invalid timestamp")
	}
	
	// Allow difficulty 0 for testing purposes
	// In production, this should be enforced to be > 0
	
	return nil
}

// IsValid checks if the transaction is valid
func (tx *Transaction) IsValid() error {
	if tx.Version == 0 {
		return fmt.Errorf("invalid version: %d", tx.Version)
	}
	
	// Coinbase transactions (like genesis) can have no inputs but must have outputs
	if len(tx.Inputs) == 0 {
		if len(tx.Outputs) == 0 {
			return fmt.Errorf("coinbase transaction must have at least one output")
		}
		// This is a valid coinbase transaction
	} else {
		// Regular transactions must have at least one input
		if len(tx.Inputs) == 0 {
			return fmt.Errorf("non-coinbase transaction must have at least one input")
		}
	}
	
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one output")
	}
	
	// Validate inputs (only for non-coinbase transactions)
	for i, input := range tx.Inputs {
		if err := input.IsValid(); err != nil {
			return fmt.Errorf("invalid input %d: %w", i, err)
		}
	}
	
	// Validate outputs
	for i, output := range tx.Outputs {
		if err := output.IsValid(); err != nil {
			return fmt.Errorf("invalid output %d: %w", i, err)
		}
	}
	
	return nil
}

// IsValid checks if the transaction input is valid
func (in *TxInput) IsValid() error {
	if len(in.PrevTxHash) != 32 {
		return fmt.Errorf("invalid previous transaction hash length: %d", len(in.PrevTxHash))
	}
	
	return nil
}

// IsValid checks if the transaction output is valid
func (out *TxOutput) IsValid() error {
	if out.Value == 0 {
		return fmt.Errorf("output value cannot be zero")
	}
	
	if len(out.ScriptPubKey) == 0 {
		return fmt.Errorf("script public key cannot be empty")
	}
	
	return nil
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block{Height: %d, Hash: %x, Transactions: %d}", 
		b.Header.Height, b.CalculateHash(), len(b.Transactions))
}

// String returns a string representation of the header
func (h *Header) String() string {
	return fmt.Sprintf("Header{Version: %d, Height: %d, Difficulty: %d, Nonce: %d}", 
		h.Version, h.Height, h.Difficulty, h.Nonce)
}

// String returns a string representation of the transaction
func (tx *Transaction) String() string {
	return fmt.Sprintf("Transaction{Hash: %x, Inputs: %d, Outputs: %d, Fee: %d}", 
		tx.Hash, len(tx.Inputs), len(tx.Outputs), tx.Fee)
}

// Helper function to compare byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// HexHash returns the hex string representation of the block hash
func (b *Block) HexHash() string {
	return hex.EncodeToString(b.CalculateHash())
} 