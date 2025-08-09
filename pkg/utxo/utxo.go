package utxo

import (
	"fmt"
	"sync"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"github.com/gochain/gochain/pkg/block"
)

// UTXOSet represents the set of unspent transaction outputs
type UTXOSet struct {
	mu       sync.RWMutex
	utxos    map[string]*UTXO  // key: "txHash:index"
	balances map[string]uint64 // address -> balance
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash       []byte `json:"tx_hash"`
	TxIndex      uint32 `json:"tx_index"`
	Value        uint64 `json:"value"`
	ScriptPubKey []byte `json:"script_pub_key"`
	Address      string `json:"address"`
	IsCoinbase   bool   `json:"is_coinbase"`
	Height       uint64 `json:"height"`
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos:    make(map[string]*UTXO),
		balances: make(map[string]uint64),
	}
}

// AddUTXO adds a UTXO to the set
func (us *UTXOSet) AddUTXO(utxo *UTXO) {
	key := us.makeKey(utxo.TxHash, utxo.TxIndex)
	us.utxos[key] = utxo

	// Update balance
	us.balances[utxo.Address] += utxo.Value
}

// RemoveUTXO removes a UTXO from the set
func (us *UTXOSet) RemoveUTXO(txHash []byte, txIndex uint32) *UTXO {
	key := us.makeKey(txHash, txIndex)
	utxo, exists := us.utxos[key]
	if !exists {
		return nil
	}

	// Update balance
	us.balances[utxo.Address] -= utxo.Value
	if us.balances[utxo.Address] == 0 {
		delete(us.balances, utxo.Address)
	}

	delete(us.utxos, key)
	return utxo
}

// GetUTXO retrieves a UTXO by transaction hash and index
func (us *UTXOSet) GetUTXO(txHash []byte, txIndex uint32) *UTXO {
	key := us.makeKey(txHash, txIndex)
	return us.utxos[key]
}

// GetBalance returns the balance of an address
func (us *UTXOSet) GetBalance(address string) uint64 {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return us.balances[address]
}

// GetAddressUTXOs returns all UTXOs for a given address
func (us *UTXOSet) GetAddressUTXOs(address string) []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var addressUTXOs []*UTXO
	for _, utxo := range us.utxos {
		if utxo.Address == address {
			addressUTXOs = append(addressUTXOs, utxo)
		}
	}

	return addressUTXOs
}

// makeKey creates a key for the UTXO map
func (us *UTXOSet) makeKey(txHash []byte, txIndex uint32) string {
	return fmt.Sprintf("%x:%d", txHash, txIndex)
}

// extractAddress extracts an address from a script public key (which is now a public key hash)
func (us *UTXOSet) extractAddress(scriptPubKey []byte) string {
	return hex.EncodeToString(scriptPubKey)
}

// ProcessBlock processes a block and updates the UTXO set
func (us *UTXOSet) ProcessBlock(block *block.Block) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		if err := us.processTransaction(tx, block.Header.Height); err != nil {
			return fmt.Errorf("failed to process transaction: %w", err)
		}
	}

	return nil
}

// processTransaction processes a single transaction
func (us *UTXOSet) processTransaction(tx *block.Transaction, height uint64) error {
	// Remove spent inputs
	for _, input := range tx.Inputs {
		// Skip coinbase transactions (they have no inputs)
		if len(input.PrevTxHash) == 0 {
			continue
		}

		// Remove the spent UTXO
		us.RemoveUTXO(input.PrevTxHash, input.PrevTxIndex)
	}

	// Add new outputs
	for i, output := range tx.Outputs {
		// Determine if this is a coinbase transaction
		isCoinbase := len(tx.Inputs) == 0

		// Extract address from script public key (simplified)
		address := us.extractAddress(output.ScriptPubKey)

		utxo := &UTXO{
			TxHash:       tx.Hash,
			TxIndex:      uint32(i),
			Value:        output.Value,
			ScriptPubKey: output.ScriptPubKey,
			Address:      address,
			IsCoinbase:   isCoinbase,
			Height:       height,
		}

		us.AddUTXO(utxo)
	}

	return nil
}

// ValidateTransaction validates a transaction against the UTXO set
func (us *UTXOSet) ValidateTransaction(tx *block.Transaction) error {
	// Coinbase transactions have no inputs, so skip input validation
	if len(tx.Inputs) == 0 {
		if len(tx.Outputs) == 0 {
			return fmt.Errorf("coinbase transaction must have at least one output")
		}
		return nil // Coinbase transactions are valid if they have outputs
	}

	// Regular transactions must have inputs and outputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Calculate total input value and verify signatures
	totalInput := uint64(0)
	for _, input := range tx.Inputs {
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return fmt.Errorf("input UTXO not found: %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}

		// Verify signature
		// The ScriptSig contains the public key (65 bytes) followed by the signature (64 bytes)
		if len(input.ScriptSig) < 65+64 {
			return fmt.Errorf("invalid scriptSig length: %d", len(input.ScriptSig))
		}
		pubBytes := input.ScriptSig[:65]
		rsBytes := input.ScriptSig[65:]

		x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
		if x == nil || y == nil {
			return fmt.Errorf("failed to unmarshal public key from scriptSig")
		}
		pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

		// Check if the public key hash matches the ScriptPubKey of the UTXO
		pubKeyHash := sha256.Sum256(pubBytes)
		if hex.EncodeToString(pubKeyHash[len(pubKeyHash)-20:]) != hex.EncodeToString(utxo.ScriptPubKey) {
			return fmt.Errorf("public key hash in scriptSig does not match UTXO scriptPubKey")
		}

		r := new(big.Int).SetBytes(rsBytes[:32])
		s := new(big.Int).SetBytes(rsBytes[32:])

		signatureData := us.getTxSignatureData(tx)
		verified := ecdsa.Verify(pub, signatureData, r, s)
		if !verified {
			return fmt.Errorf("invalid signature for input: %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}

		totalInput += utxo.Value
	}

	// Calculate total output value
	totalOutput := uint64(0)
	for _, output := range tx.Outputs {
		totalOutput += output.Value
	}

	// Check if outputs exceed inputs (including fees)
	if totalOutput > totalInput {
		return fmt.Errorf("output value %d exceeds input value %d", totalOutput, totalInput)
	}

	// Validate that the fee is reasonable
	fee := totalInput - totalOutput
	if fee < tx.Fee {
		return fmt.Errorf("actual fee %d is less than specified fee %d", fee, tx.Fee)
	}

	return nil
}

// GetStats returns UTXO set statistics
func (us *UTXOSet) GetStats() map[string]interface{} {
	us.mu.RLock()
	defer us.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_utxos"] = len(us.utxos)
	stats["total_addresses"] = len(us.balances)

	// Calculate total value
	totalValue := uint64(0)
	for _, balance := range us.balances {
		totalValue += balance
	}
	stats["total_value"] = totalValue

	return stats
}

// GetUTXOCount returns the total number of UTXOs
func (us *UTXOSet) GetUTXOCount() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.utxos)
}

// GetAddressCount returns the total number of addresses
func (us *UTXOSet) GetAddressCount() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.balances)
}

// String returns a string representation of the UTXO set
func (us *UTXOSet) String() string {
	stats := us.GetStats()
	return fmt.Sprintf("UTXOSet{UTXOs: %v, Addresses: %v, TotalValue: %v}",
		stats["total_utxos"], stats["total_addresses"], stats["total_value"])
}

// getTxSignatureData creates the data to be signed for a transaction
func (us *UTXOSet) getTxSignatureData(tx *block.Transaction) []byte {
	data := make([]byte, 0)

	// Version
	data = append(data, byte(tx.Version))

	// Inputs (excluding signatures)
	for _, input := range tx.Inputs {
		data = append(data, input.PrevTxHash...)
		data = append(data, byte(input.PrevTxIndex))
		data = append(data, byte(input.Sequence))
	}

	// Outputs
	for _, output := range tx.Outputs {
		data = append(data, byte(output.Value))
		data = append(data, output.ScriptPubKey...)
	}

	// Lock time and fee
	data = append(data, byte(tx.LockTime))
	data = append(data, byte(tx.Fee))

	// Hash the data
	hash := sha256.Sum256(data)
	return hash[:]
}

// Helpers (copied from wallet for now)
func publicKeyToBytes(k *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(elliptic.P256(), k.X, k.Y)
}

func bytesToPrivateKey(b []byte) (*ecdsa.PrivateKey, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty private key bytes")
	}
	d := new(big.Int).SetBytes(b)
	curve := elliptic.P256()
	// Validate that 0 < d < N
	if d.Sign() <= 0 || d.Cmp(curve.Params().N) >= 0 {
		return nil, fmt.Errorf("invalid private key scalar")
	}
	x, y := curve.ScalarBaseMult(b)
	return &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: curve, X: x, Y: y}, D: d}, nil
}

func concatRS(r, s *big.Int) []byte {
	rb := r.Bytes()
	sb := s.Bytes()
	out := make([]byte, 64)
	copy(out[32-len(rb):32], rb)
	copy(out[64-len(sb):], sb)
	return out
}
