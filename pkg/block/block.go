package block

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

// Block represents a block in the blockchain.
// It contains a header, a list of transactions, and the Merkle root of those transactions.
type Block struct {
	Header       *Header        // Header contains the block's metadata.
	Transactions []*Transaction // Transactions is a list of transactions included in this block.
	MerkleRoot   []byte         // MerkleRoot is the Merkle root of the block's transactions.
}

// Header contains the block header information.
// It includes metadata necessary for block validation and linking to previous blocks.
type Header struct {
	Version       uint32    // Version of the block format.
	PrevBlockHash []byte    // PrevBlockHash is the hash of the previous block in the chain.
	MerkleRoot    []byte    // MerkleRoot is the Merkle root of the transactions in this block.
	Timestamp     time.Time // Timestamp is the time when the block was mined.
	Difficulty    uint64    // Difficulty is the target difficulty for mining this block.
	Nonce         uint64    // Nonce is the value found by miners to satisfy the proof-of-work.
	Height        uint64    // Height is the block's height in the blockchain (genesis is 0).
}

// Transaction represents a transaction in the blockchain.
// It includes inputs, outputs, and metadata like version, locktime, and fee.
type Transaction struct {
	Version  uint32      // Version of the transaction format.
	Inputs   []*TxInput  // Inputs are the references to previous transaction outputs.
	Outputs  []*TxOutput // Outputs are the new transaction outputs.
	LockTime uint64      // LockTime is the earliest time a transaction can be added to a block.
	Fee      uint64      // Fee is the transaction fee paid to the miner.
	Hash     []byte      // Hash is the unique identifier for the transaction.
}

// TxInput represents a transaction input.
// It references a previous transaction's output and provides a script signature.
type TxInput struct {
	PrevTxHash  []byte // PrevTxHash is the hash of the transaction containing the output being spent.
	PrevTxIndex uint32 // PrevTxIndex is the index of the output in the previous transaction.
	ScriptSig   []byte // ScriptSig is the script that satisfies the conditions of the spent output.
	Sequence    uint32 // Sequence is a value used for advanced transaction features (e.g., Replace-by-Fee).
}

// TxOutput represents a transaction output.
// It specifies a value and a script that defines the conditions for spending this output.
type TxOutput struct {
	Value        uint64 // Value is the amount of currency in this output.
	ScriptPubKey []byte // ScriptPubKey is the script that locks the output to a recipient.
}

// NewBlock creates a new block with the given parameters.
// It initializes the block header and an empty list of transactions.
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

// AddTransaction adds a transaction to the block's list of transactions.
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// CalculateHash calculates the SHA256 hash of the block header.
// This hash serves as the block's unique identifier and is used for proof-of-work.
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

// CalculateMerkleRoot calculates the Merkle root of all transactions in the block.
// The Merkle root provides a compact way to verify the integrity of all transactions.
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
// buildMerkleTree recursively builds a Merkle tree from a slice of transaction hashes.
// It returns the Merkle root (the top hash of the tree).
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

// IsValid checks if the block is valid according to its internal consistency rules.
// It validates the header, Merkle root, and all contained transactions.
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

// IsValid checks if the header is valid according to its internal consistency rules.
// It validates fields like version, previous block hash, Merkle root, timestamp, and difficulty.
func (h *Header) IsValid() error {
	if h.Version == 0 {
		return fmt.Errorf("invalid version: %d", h.Version)
	}

	if h.PrevBlockHash == nil {
		return fmt.Errorf("previous block hash cannot be nil")
	}

	if h.MerkleRoot == nil || len(h.MerkleRoot) != 32 {
		return fmt.Errorf("invalid merkle root: %x", h.MerkleRoot)
	}

	if h.Timestamp.IsZero() {
		return fmt.Errorf("invalid timestamp")
	}

	if h.Difficulty == 0 {
		return fmt.Errorf("difficulty cannot be zero")
	}

	// In production, this should be enforced to be > 0

	return nil
}

// IsValid checks if the transaction is valid according to its internal consistency rules.
// It validates fields like version, hash, and the structure of inputs and outputs.
func (tx *Transaction) IsValid() error {
	if tx.Version == 0 {
		return fmt.Errorf("invalid version: %d", tx.Version)
	}

	if len(tx.Hash) != 32 {
		return fmt.Errorf("invalid transaction hash length: %d", len(tx.Hash))
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

// IsValid checks if the transaction input is valid according to its internal consistency rules.
// It primarily validates the length of the previous transaction hash.
func (in *TxInput) IsValid() error {
	if len(in.PrevTxHash) != 32 {
		return fmt.Errorf("invalid previous transaction hash length: %d", len(in.PrevTxHash))
	}

	return nil
}

// IsValid checks if the transaction output is valid according to its internal consistency rules.
// It validates the output value and the presence of a script public key.
func (out *TxOutput) IsValid() error {
	if out.Value == 0 {
		return fmt.Errorf("output value cannot be zero")
	}

	if len(out.ScriptPubKey) == 0 {
		return fmt.Errorf("script public key cannot be empty")
	}

	return nil
}

// String returns a human-readable string representation of the block.
func (b *Block) String() string {
	return fmt.Sprintf("Block{Height: %d, Hash: %x, Transactions: %d}",
		b.Header.Height, b.CalculateHash(), len(b.Transactions))
}

// String returns a human-readable string representation of the block header.
func (h *Header) String() string {
	return fmt.Sprintf("Header{Version: %d, Height: %d, Difficulty: %d, Nonce: %d}",
		h.Version, h.Height, h.Difficulty, h.Nonce)
}

// String returns a human-readable string representation of the transaction.
func (tx *Transaction) String() string {
	return fmt.Sprintf("Transaction{Hash: %x, Inputs: %d, Outputs: %d, Fee: %d}",
		tx.Hash, len(tx.Inputs), len(tx.Outputs), tx.Fee)
}

// Helper function to compare byte slices
// bytesEqual checks if two byte slices are equal.
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

// HexHash returns the hexadecimal string representation of the block's hash.
func (b *Block) HexHash() string {
	return hex.EncodeToString(b.CalculateHash())
}
