package utxo

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/palaseus/adrenochain/pkg/block"
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

// NewUTXO creates a new UTXO with the given parameters
func NewUTXO(txHash []byte, txIndex uint32, value uint64, scriptPubKey []byte, address string, isCoinbase bool, height uint64) *UTXO {
	return &UTXO{
		TxHash:       txHash,
		TxIndex:      txIndex,
		Value:        value,
		ScriptPubKey: scriptPubKey,
		Address:      address,
		IsCoinbase:   isCoinbase,
		Height:       height,
	}
}

// AddUTXO adds a UTXO to the set
func (us *UTXOSet) AddUTXO(utxo *UTXO) {
	if utxo == nil {
		return
	}
	key := us.makeKey(utxo.TxHash, utxo.TxIndex)
	us.utxos[key] = utxo

	// Update balance
	us.balances[utxo.Address] += utxo.Value
}

// AddUTXOSafe adds a UTXO to the set with proper locking (for external use)
func (us *UTXOSet) AddUTXOSafe(utxo *UTXO) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.AddUTXO(utxo)
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

// RemoveUTXOSafe removes a UTXO from the set with proper locking (for external use)
func (us *UTXOSet) RemoveUTXOSafe(txHash []byte, txIndex uint32) *UTXO {
	us.mu.Lock()
	defer us.mu.Unlock()
	return us.RemoveUTXO(txHash, txIndex)
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

// extractAddress extracts an address from a script public key (which is now a public key hash)
func (us *UTXOSet) extractAddress(scriptPubKey []byte) string {
	return hex.EncodeToString(scriptPubKey)
}

// ProcessBlock processes a block and updates the UTXO set
func (us *UTXOSet) ProcessBlock(block *block.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if block.Header == nil {
		return fmt.Errorf("block header is nil")
	}

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

// ValidateTransaction validates a transaction against the current UTXO set.
// It performs comprehensive validation including signature verification, UTXO existence,
// and proper fee calculation.
// Note: This method treats transactions with no inputs as potentially valid (coinbase-like),
// but for strict validation in block context, use ValidateTransactionInBlock.
func (us *UTXOSet) ValidateTransaction(tx *block.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// Transactions with no inputs are potentially coinbase transactions
	if len(tx.Inputs) == 0 {
		if len(tx.Outputs) == 0 {
			return fmt.Errorf("transaction with no inputs must have at least one output")
		}
		// Validate outputs
		for i, output := range tx.Outputs {
			if output.Value == 0 {
				return fmt.Errorf("output %d has zero value", i)
			}
			if len(output.ScriptPubKey) == 0 {
				return fmt.Errorf("output %d has empty script public key", i)
			}
		}
		return nil // Transactions with no inputs are valid if they have valid outputs
	}

	// Regular transactions must have outputs
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Check for duplicate inputs (double-spend prevention)
	inputSet := make(map[string]bool)
	for _, input := range tx.Inputs {
		inputKey := fmt.Sprintf("%x:%d", input.PrevTxHash, input.PrevTxIndex)
		if inputSet[inputKey] {
			return fmt.Errorf("duplicate input: %s", inputKey)
		}
		inputSet[inputKey] = true
	}

	// Calculate total input value and verify signatures
	totalInput := uint64(0)
	for i, input := range tx.Inputs {
		// Validate input structure
		if err := input.IsValid(); err != nil {
			return fmt.Errorf("invalid input %d: %w", i, err)
		}

		// Check if UTXO exists and is not already spent
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return fmt.Errorf("input UTXO not found: %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}

		// Check if UTXO is coinbase and has matured (if applicable)
		if utxo.IsCoinbase {
			// For now, we'll allow coinbase UTXOs to be spent immediately
			// In a real implementation, you might want to enforce maturity requirements
		}

		// Verify signature length and structure
		if len(input.ScriptSig) < 65+64 {
			return fmt.Errorf("input %d: invalid scriptSig length: %d (expected >= 129)", i, len(input.ScriptSig))
		}

		// Extract public key and signature from ScriptSig
		pubBytes := input.ScriptSig[:65]
		rsBytes := input.ScriptSig[65:]

		// Validate public key format
		pubKey, err := btcec.ParsePubKey(pubBytes)
		if err != nil {
			return fmt.Errorf("input %d: failed to unmarshal public key from scriptSig: %v", i, err)
		}
		pub := pubKey.ToECDSA()

		// Verify public key hash matches the UTXO's ScriptPubKey
		pubKeyHash := sha256.Sum256(pubBytes)
		expectedAddress := hex.EncodeToString(pubKeyHash[len(pubKeyHash)-20:])
		utxoAddress := hex.EncodeToString(utxo.ScriptPubKey)

		if expectedAddress != utxoAddress {
			return fmt.Errorf("input %d: public key hash %s does not match UTXO scriptPubKey %s",
				i, expectedAddress, utxoAddress)
		}

		// Extract R and S components from signature
		if len(rsBytes) < 64 {
			return fmt.Errorf("input %d: insufficient signature data", i)
		}
		r := new(big.Int).SetBytes(rsBytes[:32])
		s := new(big.Int).SetBytes(rsBytes[32:64])

		// Validate signature components
		if r.Sign() <= 0 || s.Sign() <= 0 {
			return fmt.Errorf("input %d: invalid signature components (R or S <= 0)", i)
		}

		// Verify signature
		signatureData := us.getTxSignatureData(tx)
		verified := ecdsa.Verify(pub, signatureData, r, s)
		if !verified {
			return fmt.Errorf("input %d: invalid signature for UTXO %x:%d", i, input.PrevTxHash, input.PrevTxIndex)
		}

		totalInput += utxo.Value
	}

	// Calculate total output value and validate outputs
	totalOutput := uint64(0)
	for i, output := range tx.Outputs {
		if err := output.IsValid(); err != nil {
			return fmt.Errorf("invalid output %d: %w", i, err)
		}
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

	// Additional security checks
	if fee > totalInput/2 {
		return fmt.Errorf("fee %d is unreasonably high (more than 50%% of input value %d)", fee, totalInput)
	}

	// Check for dust outputs (very small outputs that are uneconomical)
	const dustThreshold = 546 // Satoshis, equivalent to Bitcoin's dust threshold
	for i, output := range tx.Outputs {
		if output.Value < dustThreshold {
			return fmt.Errorf("output %d value %d is below dust threshold %d", i, output.Value, dustThreshold)
		}
	}

	return nil
}

// ValidateTransactionInBlock validates a transaction in the context of a block.
// This method properly distinguishes between coinbase transactions (first transaction in block)
// and regular transactions.
func (us *UTXOSet) ValidateTransactionInBlock(tx *block.Transaction, block *block.Block, txIndex int) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if txIndex < 0 || txIndex >= len(block.Transactions) {
		return fmt.Errorf("transaction index %d out of bounds for block with %d transactions", txIndex, len(block.Transactions))
	}

	// Check if this is a coinbase transaction (first transaction in block)
	isCoinbase := txIndex == 0 && len(block.Transactions) > 0 && tx == block.Transactions[0]

	if isCoinbase {
		// Coinbase transactions have no inputs
		if len(tx.Inputs) != 0 {
			return fmt.Errorf("coinbase transaction should have no inputs")
		}
		if len(tx.Outputs) == 0 {
			return fmt.Errorf("coinbase transaction must have at least one output")
		}
		// Validate coinbase transaction outputs
		for i, output := range tx.Outputs {
			if output.Value == 0 {
				return fmt.Errorf("coinbase output %d has zero value", i)
			}
			if len(output.ScriptPubKey) == 0 {
				return fmt.Errorf("coinbase output %d has empty script public key", i)
			}
		}
		return nil // Coinbase transactions are valid if they have valid outputs
	}

	// Regular transactions must have inputs and outputs
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("regular transaction must have inputs")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("regular transaction must have outputs")
	}

	// Check for duplicate inputs (double-spend prevention)
	inputSet := make(map[string]bool)
	for _, input := range tx.Inputs {
		inputKey := fmt.Sprintf("%x:%d", input.PrevTxHash, input.PrevTxIndex)
		if inputSet[inputKey] {
			return fmt.Errorf("duplicate input: %s", inputKey)
		}
		inputSet[inputKey] = true
	}

	// Calculate total input value and verify signatures
	totalInput := uint64(0)
	for i, input := range tx.Inputs {
		// Validate input structure
		if err := input.IsValid(); err != nil {
			return fmt.Errorf("invalid input %d: %w", i, err)
		}

		// Check if UTXO exists and is not already spent
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return fmt.Errorf("input UTXO not found: %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}

		// Check if UTXO is coinbase and has matured (if applicable)
		if utxo.IsCoinbase {
			// For now, we'll allow coinbase UTXOs to be spent immediately
			// In a real implementation, you might want to enforce maturity requirements
		}

		// Verify signature length and structure
		if len(input.ScriptSig) < 65+64 {
			return fmt.Errorf("input %d: invalid scriptSig length: %d (expected >= 129)", i, len(input.ScriptSig))
		}

		// Extract public key and signature from ScriptSig
		pubBytes := input.ScriptSig[:65]
		rsBytes := input.ScriptSig[65:]

		// Validate public key format
		pubKey, err := btcec.ParsePubKey(pubBytes)
		if err != nil {
			return fmt.Errorf("input %d: failed to unmarshal public key from scriptSig: %v", i, err)
		}
		pub := pubKey.ToECDSA()

		// Verify public key hash matches the UTXO's ScriptPubKey
		pubKeyHash := sha256.Sum256(pubBytes)
		expectedAddress := hex.EncodeToString(pubKeyHash[len(pubKeyHash)-20:])
		utxoAddress := hex.EncodeToString(utxo.ScriptPubKey)

		if expectedAddress != utxoAddress {
			return fmt.Errorf("input %d: public key hash %s does not match UTXO scriptPubKey %s",
				i, expectedAddress, utxoAddress)
		}

		// Extract R and S components from signature
		if len(rsBytes) < 64 {
			return fmt.Errorf("input %d: insufficient signature data", i)
		}
		r := new(big.Int).SetBytes(rsBytes[:32])
		s := new(big.Int).SetBytes(rsBytes[32:64])

		// Validate signature components
		if r.Sign() <= 0 || s.Sign() <= 0 {
			return fmt.Errorf("input %d: invalid signature components (R or S <= 0)", i)
		}

		// Verify signature
		signatureData := us.getTxSignatureData(tx)
		verified := ecdsa.Verify(pub, signatureData, r, s)
		if !verified {
			return fmt.Errorf("input %d: invalid signature for UTXO %x:%d", i, input.PrevTxHash, input.PrevTxIndex)
		}

		totalInput += utxo.Value
	}

	// Calculate total output value and validate outputs
	totalOutput := uint64(0)
	for i, output := range tx.Outputs {
		if err := output.IsValid(); err != nil {
			return fmt.Errorf("invalid output %d: %w", i, err)
		}
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

	// Additional security checks
	if fee > totalInput/2 {
		return fmt.Errorf("fee %d is unreasonably high (more than 50%% of input value %d)", fee, totalInput)
	}

	// Check for dust outputs (very small outputs that are uneconomical)
	const dustThreshold = 546 // Satoshis, equivalent to Bitcoin's dust threshold
	for i, output := range tx.Outputs {
		if output.Value < dustThreshold {
			return fmt.Errorf("output %d value %d is below dust threshold %d", i, output.Value, dustThreshold)
		}
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

// IsDoubleSpend checks if a transaction attempts to spend UTXOs that are already spent
func (us *UTXOSet) IsDoubleSpend(tx *block.Transaction) bool {
	for _, input := range tx.Inputs {
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			// UTXO doesn't exist, which means it's already spent or never existed
			return true
		}
	}
	return false
}

// CalculateFee calculates the transaction fee based on input and output values
func (us *UTXOSet) CalculateFee(tx *block.Transaction) (uint64, error) {
	if len(tx.Inputs) == 0 {
		// Coinbase transaction has no fee
		return 0, nil
	}

	totalInput := uint64(0)
	for _, input := range tx.Inputs {
		utxo := us.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return 0, fmt.Errorf("UTXO not found for input %x:%d", input.PrevTxHash, input.PrevTxIndex)
		}
		totalInput += utxo.Value
	}

	totalOutput := uint64(0)
	for _, output := range tx.Outputs {
		totalOutput += output.Value
	}

	if totalOutput > totalInput {
		return 0, fmt.Errorf("output value %d exceeds input value %d", totalOutput, totalInput)
	}

	return totalInput - totalOutput, nil
}

// ValidateFeeRate validates that the transaction fee meets minimum requirements
func (us *UTXOSet) ValidateFeeRate(tx *block.Transaction, minFeeRate uint64) error {
	if len(tx.Inputs) == 0 {
		// Coinbase transactions don't need fee validation
		return nil
	}

	// Calculate actual transaction size by serializing the transaction
	txSize := uint64(0)

	// Version (4 bytes)
	txSize += 4

	// Input count (varint, but we'll use 1 byte for simplicity in tests)
	txSize += 1

	// Inputs
	for _, input := range tx.Inputs {
		txSize += 32                           // PrevTxHash
		txSize += 4                            // PrevTxIndex
		txSize += uint64(len(input.ScriptSig)) // ScriptSig
		txSize += 4                            // Sequence
	}

	// Output count (varint, but we'll use 1 byte for simplicity in tests)
	txSize += 1

	// Outputs
	for _, output := range tx.Outputs {
		txSize += 8                                // Value
		txSize += uint64(len(output.ScriptPubKey)) // ScriptPubKey
	}

	// LockTime (8 bytes)
	txSize += 8

	// Fee (8 bytes)
	txSize += 8

	// Calculate minimum required fee
	minFee := txSize * minFeeRate / 1000 // Fee rate is in satoshis per kilobyte

	actualFee, err := us.CalculateFee(tx)
	if err != nil {
		return fmt.Errorf("failed to calculate fee: %w", err)
	}

	if actualFee < minFee {
		return fmt.Errorf("fee %d is below minimum required fee %d (size: %d bytes, rate: %d sat/kilobyte)",
			actualFee, minFee, txSize, minFeeRate)
	}

	return nil
}

// GetSpendableUTXOs returns all spendable UTXOs for a given address
// This is useful for wallet implementations to find available funds
func (us *UTXOSet) GetSpendableUTXOs(address string, minValue uint64) []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var spendableUTXOs []*UTXO
	for _, utxo := range us.utxos {
		if utxo.Address == address && utxo.Value >= minValue {
			spendableUTXOs = append(spendableUTXOs, utxo)
		}
	}
	return spendableUTXOs
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

func concatRS(r, s *big.Int) []byte {
	rb := r.Bytes()
	sb := s.Bytes()
	out := make([]byte, 64)
	copy(out[32-len(rb):32], rb)
	copy(out[64-len(sb):], sb)
	return out
}
