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
			Nonce:         0,
		},
		Transactions: make([]*Transaction, 0),
	}

	// Initialize the Merkle root for empty block
	block.Header.MerkleRoot = block.CalculateMerkleRoot()

	return block
}

// NewTransaction creates a new transaction with the given parameters.
// It initializes the transaction with default values and calculates the hash.
func NewTransaction(inputs []*TxInput, outputs []*TxOutput, fee uint64) *Transaction {
	tx := &Transaction{
		Version:  1,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: 0,
		Fee:      fee,
		Hash:     make([]byte, 32), // Initialize empty hash
	}

	// Calculate transaction hash
	tx.Hash = tx.CalculateHash()

	return tx
}

// AddTransaction adds a transaction to the block's list of transactions.
func (b *Block) AddTransaction(tx *Transaction) {
	// Calculate transaction hash if not already set
	if tx.Hash == nil {
		tx.Hash = tx.CalculateHash()
	}
	b.Transactions = append(b.Transactions, tx)

	// Update Merkle root after adding transaction
	b.Header.MerkleRoot = b.CalculateMerkleRoot()
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

// CalculateHash calculates the SHA256 hash of the transaction.
func (tx *Transaction) CalculateHash() []byte {
	data := make([]byte, 0)

	// Version
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)

	// Inputs
	for _, input := range tx.Inputs {
		if input != nil {
			data = append(data, input.PrevTxHash...)
			inputIndexBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(inputIndexBytes, input.PrevTxIndex)
			data = append(data, inputIndexBytes...)
			data = append(data, input.ScriptSig...)
			sequenceBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(sequenceBytes, input.Sequence)
			data = append(data, sequenceBytes...)
		}
	}

	// Outputs
	for _, output := range tx.Outputs {
		if output != nil {
			valueBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(valueBytes, output.Value)
			data = append(data, valueBytes...)
			data = append(data, output.ScriptPubKey...)
		}
	}

	// LockTime
	lockTimeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)

	// Fee
	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	data = append(data, feeBytes...)

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

// IsCoinbase checks if a transaction is a coinbase transaction.
// A coinbase transaction is one that has no inputs (creates new coins).
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 0
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

// Serialize converts the block to a byte array for network transmission
func (b *Block) Serialize() ([]byte, error) {
	// This is a simplified serialization implementation
	// In a real implementation, you'd use a more efficient format like Protocol Buffers

	// Check if header exists
	if b.Header == nil {
		return nil, fmt.Errorf("cannot serialize block with nil header")
	}

	data := make([]byte, 0)

	// Serialize header
	headerData, err := b.Header.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize header: %w", err)
	}

	// Add header length and data
	headerLen := uint32(len(headerData))
	headerLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(headerLenBytes, headerLen)
	data = append(data, headerLenBytes...)
	data = append(data, headerData...)

	// Add transaction count
	txCount := uint32(len(b.Transactions))
	txCountBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(txCountBytes, txCount)
	data = append(data, txCountBytes...)

	// Serialize transactions
	for _, tx := range b.Transactions {
		txData, err := tx.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize transaction: %w", err)
		}

		// Add transaction length and data
		txLen := uint32(len(txData))
		txLenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(txLenBytes, txLen)
		data = append(data, txLenBytes...)
		data = append(data, txData...)
	}

	return data, nil
}

// Deserialize reconstructs a block from a byte array
func (b *Block) Deserialize(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for block deserialization")
	}

	offset := 0

	// Read header length
	headerLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	if len(data) < offset+int(headerLen) {
		return fmt.Errorf("insufficient data for header")
	}

	// Deserialize header
	header := &Header{}
	if err := header.Deserialize(data[offset : offset+int(headerLen)]); err != nil {
		return fmt.Errorf("failed to deserialize header: %w", err)
	}
	b.Header = header
	offset += int(headerLen)

	// Read transaction count
	if len(data) < offset+4 {
		return fmt.Errorf("insufficient data for transaction count")
	}
	txCount := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Deserialize transactions
	b.Transactions = make([]*Transaction, 0, txCount)
	for i := uint32(0); i < txCount; i++ {
		if len(data) < offset+4 {
			return fmt.Errorf("insufficient data for transaction %d length", i)
		}

		txLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		if len(data) < offset+int(txLen) {
			return fmt.Errorf("insufficient data for transaction %d", i)
		}

		tx := &Transaction{}
		if err := tx.Deserialize(data[offset : offset+int(txLen)]); err != nil {
			return fmt.Errorf("failed to deserialize transaction %d: %w", i, err)
		}
		b.Transactions = append(b.Transactions, tx)
		offset += int(txLen)
	}

	// Recalculate Merkle root
	b.MerkleRoot = b.CalculateMerkleRoot()

	return nil
}

// GetHeader returns the block header
func (b *Block) GetHeader() interface{} {
	return b.Header
}

// GetVersion returns the header version
func (h *Header) GetVersion() uint32 {
	return h.Version
}

// GetPrevBlockHash returns the previous block hash
func (h *Header) GetPrevBlockHash() []byte {
	return h.PrevBlockHash
}

// GetMerkleRoot returns the merkle root
func (h *Header) GetMerkleRoot() []byte {
	return h.MerkleRoot
}

// GetTimestamp returns the timestamp
func (h *Header) GetTimestamp() time.Time {
	return h.Timestamp
}

// GetDifficulty returns the difficulty
func (h *Header) GetDifficulty() uint64 {
	return h.Difficulty
}

// GetNonce returns the nonce
func (h *Header) GetNonce() uint64 {
	return h.Nonce
}

// GetHeight returns the height
func (h *Header) GetHeight() uint64 {
	return h.Height
}

// Serialize converts the header to a byte array
func (h *Header) Serialize() ([]byte, error) {
	data := make([]byte, 0)

	// Version (4 bytes)
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, h.Version)
	data = append(data, versionBytes...)

	// Previous block hash (32 bytes)
	data = append(data, h.PrevBlockHash...)

	// Merkle root (32 bytes)
	data = append(data, h.MerkleRoot...)

	// Timestamp (8 bytes)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(h.Timestamp.Unix()))
	data = append(data, timestampBytes...)

	// Difficulty (8 bytes)
	difficultyBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(difficultyBytes, h.Difficulty)
	data = append(data, difficultyBytes...)

	// Nonce (8 bytes)
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, h.Nonce)
	data = append(data, nonceBytes...)

	// Height (8 bytes)
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, h.Height)
	data = append(data, heightBytes...)

	return data, nil
}

// Deserialize reconstructs a header from a byte array
func (h *Header) Deserialize(data []byte) error {
	if len(data) < 100 { // 4+32+32+8+8+8+8 = 100 bytes
		return fmt.Errorf("insufficient data for header deserialization")
	}

	offset := 0

	// Version
	h.Version = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Previous block hash
	h.PrevBlockHash = make([]byte, 32)
	copy(h.PrevBlockHash, data[offset:offset+32])
	offset += 32

	// Merkle root
	h.MerkleRoot = make([]byte, 32)
	copy(h.MerkleRoot, data[offset:offset+32])
	offset += 32

	// Timestamp
	timestamp := binary.BigEndian.Uint64(data[offset : offset+8])
	h.Timestamp = time.Unix(int64(timestamp), 0)
	offset += 8

	// Difficulty
	h.Difficulty = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Nonce
	h.Nonce = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Height
	h.Height = binary.BigEndian.Uint64(data[offset : offset+8])

	return nil
}

// Serialize converts the transaction to a byte array
func (tx *Transaction) Serialize() ([]byte, error) {
	// Check if hash exists
	if tx.Hash == nil {
		return nil, fmt.Errorf("cannot serialize transaction with nil hash")
	}

	data := make([]byte, 0)

	// Version (4 bytes)
	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, tx.Version)
	data = append(data, versionBytes...)

	// Input count (4 bytes)
	inputCount := uint32(len(tx.Inputs))
	inputCountBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(inputCountBytes, inputCount)
	data = append(data, inputCountBytes...)

	// Serialize inputs
	for _, input := range tx.Inputs {
		inputData, err := input.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize input: %w", err)
		}

		// Add input length and data
		inputLen := uint32(len(inputData))
		inputLenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(inputLenBytes, inputLen)
		data = append(data, inputLenBytes...)
		data = append(data, inputData...)
	}

	// Output count (4 bytes)
	outputCount := uint32(len(tx.Outputs))
	outputCountBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(outputCountBytes, outputCount)
	data = append(data, outputCountBytes...)

	// Serialize outputs
	for _, output := range tx.Outputs {
		outputData, err := output.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize output: %w", err)
		}

		// Add output length and data
		outputLen := uint32(len(outputData))
		outputLenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(outputLenBytes, outputLen)
		data = append(data, outputLenBytes...)
		data = append(data, outputData...)
	}

	// Lock time (8 bytes)
	lockTimeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lockTimeBytes, tx.LockTime)
	data = append(data, lockTimeBytes...)

	// Fee (8 bytes)
	feeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(feeBytes, tx.Fee)
	data = append(data, feeBytes...)

	// Hash (32 bytes)
	data = append(data, tx.Hash...)

	return data, nil
}

// Deserialize reconstructs a transaction from a byte array
func (tx *Transaction) Deserialize(data []byte) error {
	if len(data) < 60 { // Minimum size for a transaction
		return fmt.Errorf("insufficient data for transaction deserialization")
	}

	offset := 0

	// Version
	tx.Version = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Input count
	inputCount := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Deserialize inputs
	tx.Inputs = make([]*TxInput, 0, inputCount)
	for i := uint32(0); i < inputCount; i++ {
		if len(data) < offset+4 {
			return fmt.Errorf("insufficient data for input %d length", i)
		}

		inputLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		if len(data) < offset+int(inputLen) {
			return fmt.Errorf("insufficient data for input %d", i)
		}

		input := &TxInput{}
		if err := input.Deserialize(data[offset : offset+int(inputLen)]); err != nil {
			return fmt.Errorf("failed to deserialize input %d: %w", i, err)
		}
		tx.Inputs = append(tx.Inputs, input)
		offset += int(inputLen)
	}

	// Output count
	if len(data) < offset+4 {
		return fmt.Errorf("insufficient data for output count")
	}
	outputCount := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Deserialize outputs
	tx.Outputs = make([]*TxOutput, 0, outputCount)
	for i := uint32(0); i < outputCount; i++ {
		if len(data) < offset+4 {
			return fmt.Errorf("insufficient data for output %d length", i)
		}

		outputLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		if len(data) < offset+int(outputLen) {
			return fmt.Errorf("insufficient data for output %d", i)
		}

		output := &TxOutput{}
		if err := output.Deserialize(data[offset : offset+int(outputLen)]); err != nil {
			return fmt.Errorf("failed to deserialize output %d: %w", i, err)
		}
		tx.Outputs = append(tx.Outputs, output)
		offset += int(outputLen)
	}

	// Lock time
	if len(data) < offset+8 {
		return fmt.Errorf("insufficient data for lock time")
	}
	tx.LockTime = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Fee
	if len(data) < offset+8 {
		return fmt.Errorf("insufficient data for fee")
	}
	tx.Fee = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Hash
	if len(data) < offset+32 {
		return fmt.Errorf("insufficient data for hash")
	}
	tx.Hash = make([]byte, 32)
	copy(tx.Hash, data[offset:offset+32])

	return nil
}

// Serialize converts the transaction input to a byte array
func (in *TxInput) Serialize() ([]byte, error) {
	// Check if required fields exist
	if in.PrevTxHash == nil {
		return nil, fmt.Errorf("cannot serialize input with nil previous transaction hash")
	}

	data := make([]byte, 0)

	// Previous transaction hash (32 bytes)
	data = append(data, in.PrevTxHash...)

	// Previous transaction index (4 bytes)
	prevTxIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(prevTxIndexBytes, in.PrevTxIndex)
	data = append(data, prevTxIndexBytes...)

	// Script signature length (4 bytes)
	scriptSigLen := uint32(len(in.ScriptSig))
	scriptSigLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(scriptSigLenBytes, scriptSigLen)
	data = append(data, scriptSigLenBytes...)

	// Script signature
	data = append(data, in.ScriptSig...)

	// Sequence (4 bytes)
	sequenceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceBytes, in.Sequence)
	data = append(data, sequenceBytes...)

	return data, nil
}

// Deserialize reconstructs a transaction input from a byte array
func (in *TxInput) Deserialize(data []byte) error {
	if len(data) < 44 { // 32+4+4+4 = 44 bytes minimum
		return fmt.Errorf("insufficient data for transaction input deserialization")
	}

	offset := 0

	// Previous transaction hash
	in.PrevTxHash = make([]byte, 32)
	copy(in.PrevTxHash, data[offset:offset+32])
	offset += 32

	// Previous transaction index
	in.PrevTxIndex = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Script signature length
	scriptSigLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Script signature
	if len(data) < offset+int(scriptSigLen) {
		return fmt.Errorf("insufficient data for script signature")
	}
	in.ScriptSig = make([]byte, scriptSigLen)
	copy(in.ScriptSig, data[offset:offset+int(scriptSigLen)])
	offset += int(scriptSigLen)

	// Sequence
	if len(data) < offset+4 {
		return fmt.Errorf("insufficient data for sequence")
	}
	in.Sequence = binary.BigEndian.Uint32(data[offset : offset+4])

	return nil
}

// Serialize converts the transaction output to a byte array
func (out *TxOutput) Serialize() ([]byte, error) {
	// Check if required fields exist
	if out.ScriptPubKey == nil {
		return nil, fmt.Errorf("cannot serialize output with nil script public key")
	}

	data := make([]byte, 0)

	// Value (8 bytes)
	valueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBytes, out.Value)
	data = append(data, valueBytes...)

	// Script public key length (4 bytes)
	scriptPubKeyLen := uint32(len(out.ScriptPubKey))
	scriptPubKeyLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(scriptPubKeyLenBytes, scriptPubKeyLen)
	data = append(data, scriptPubKeyLenBytes...)

	// Script public key
	data = append(data, out.ScriptPubKey...)

	return data, nil
}

// Deserialize reconstructs a transaction output from a byte array
func (out *TxOutput) Deserialize(data []byte) error {
	if len(data) < 12 { // 8+4 = 12 bytes minimum
		return fmt.Errorf("insufficient data for transaction output deserialization")
	}

	offset := 0

	// Value
	out.Value = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Script public key length
	scriptPubKeyLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Script public key
	if len(data) < offset+int(scriptPubKeyLen) {
		return fmt.Errorf("insufficient data for script public key")
	}
	out.ScriptPubKey = make([]byte, scriptPubKeyLen)
	copy(out.ScriptPubKey, data[offset:offset+int(scriptPubKeyLen)])

	return nil
}
