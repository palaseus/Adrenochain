package utxo

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/gochain/gochain/pkg/block"
)

// UTXOSet represents the set of unspent transaction outputs
type UTXOSet struct {
	mu      sync.RWMutex
	utxos   map[string]*UTXO // key: "txHash:index"
	balances map[string]uint64 // address -> balance
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash      []byte `json:"tx_hash"`
	TxIndex     uint32 `json:"tx_index"`
	Value       uint64 `json:"value"`
	ScriptPubKey []byte `json:"script_pub_key"`
	Address     string `json:"address"`
	IsCoinbase  bool   `json:"is_coinbase"`
	Height      uint64 `json:"height"`
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
	us.mu.Lock()
	defer us.mu.Unlock()
	
	key := us.makeKey(utxo.TxHash, utxo.TxIndex)
	us.utxos[key] = utxo
	
	// Update balance
	us.balances[utxo.Address] += utxo.Value
}

// RemoveUTXO removes a UTXO from the set
func (us *UTXOSet) RemoveUTXO(txHash []byte, txIndex uint32) *UTXO {
	us.mu.Lock()
	defer us.mu.Unlock()
	
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
	us.mu.RLock()
	defer us.mu.RUnlock()
	
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
		
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return fmt.Errorf("input UTXO not found: %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}
		
		// Validate that the input can be spent
		if !us.canSpendUTXO(utxo, input.ScriptSig) {
			return fmt.Errorf("cannot spend UTXO: %x:%d", input.PrevTxHash, input.PrevTxIndex)
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
			TxHash:      tx.Hash,
			TxIndex:     uint32(i),
			Value:       output.Value,
			ScriptPubKey: output.ScriptPubKey,
			Address:     address,
			IsCoinbase:  isCoinbase,
			Height:      height,
		}
		
		us.AddUTXO(utxo)
	}
	
	return nil
}

// canSpendUTXO checks if a UTXO can be spent with the given script signature
func (us *UTXOSet) canSpendUTXO(utxo *UTXO, scriptSig []byte) bool {
	// In a real implementation, this would validate the script
	// For now, we'll use a simplified approach
	
	// Check if the script signature matches the expected pattern
	if len(scriptSig) == 0 {
		return false
	}
	
	// For coinbase transactions, we might have additional validation
	if utxo.IsCoinbase {
		// Coinbase outputs might have maturity requirements
		// For now, we'll allow them to be spent
		return true
	}
	
	// Basic validation: check if script signature is not empty
	return len(scriptSig) > 0
}

// extractAddress extracts an address from a script public key
func (us *UTXOSet) extractAddress(scriptPubKey []byte) string {
	// In a real implementation, this would parse the script and extract the address
	// For now, we'll use a simplified approach
	
	if len(scriptPubKey) == 0 {
		return "unknown"
	}
	
	// If it looks like a hex string, use it directly
	if len(scriptPubKey) >= 20 {
		return fmt.Sprintf("%x", scriptPubKey[:20])
	}
	
	// Otherwise, hash the script and use the first 20 bytes
	hash := sha256.Sum256(scriptPubKey)
	return fmt.Sprintf("%x", hash[:20])
}

// ValidateTransaction validates a transaction against the UTXO set
func (us *UTXOSet) ValidateTransaction(tx *block.Transaction) error {
	// Check if transaction has inputs and outputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}
	
	// Calculate total input value
	totalInput := uint64(0)
	for _, input := range tx.Inputs {
		// Skip coinbase transactions
		if len(input.PrevTxHash) == 0 {
			continue
		}
		
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return fmt.Errorf("input UTXO not found: %x:%d", input.PrevTxHash, input.PrevTxIndex)
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