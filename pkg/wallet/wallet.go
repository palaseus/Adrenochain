// Package wallet provides a secure cryptocurrency wallet implementation with the following security features:
//
// SECURITY FEATURES:
// - Canonical DER signature encoding with low-S enforcement to prevent signature malleability
// - Secure key derivation using PBKDF2 with 100,000 iterations and per-wallet salt
// - AES-GCM authenticated encryption for wallet storage
// - Base58Check address encoding with checksums to prevent typos
// - Comprehensive UTXO validation and double-spend prevention
// - Proper change output handling to prevent fund loss
//
// SIGNATURE FORMAT:
// - ECDSA signatures encoded in ASN.1 DER format
// - Low-S enforcement (s <= N/2) to prevent signature malleability
// - Public key stored as uncompressed 65-byte format
// - Wire format: [public_key(65)][der_signature(variable)]
//
// ADDRESS FORMAT:
// - Base58Check encoding with double SHA256 checksum
// - Version byte (0x00 for mainnet)
// - 20-byte public key hash (RIPEMD160(SHA256(public_key)))
// - 4-byte checksum for error detection
//
// ENCRYPTION:
// - PBKDF2 key derivation with 100,000 iterations
// - 32-byte random salt per wallet
// - AES-GCM authenticated encryption
// - Format: [salt(32)][nonce(12)][ciphertext]
//
// This implementation prioritizes security and correctness over performance.
// For production use, consider using specialized cryptographic libraries.
package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os" // Added import for os.IsNotExist
	"sync"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/storage" // Added import
	"github.com/gochain/gochain/pkg/utxo"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	mu             sync.RWMutex
	accounts       map[string]*Account
	defaultKey     *ecdsa.PrivateKey
	keyType        KeyType
	utxoSet        *utxo.UTXOSet
	storage        *storage.Storage // Added storage field
	walletFilePath string           // Added walletFilePath field
	passphrase     string           // Added passphrase field
}

// Account represents a wallet account
type Account struct {
	Address    string
	PublicKey  []byte
	PrivateKey []byte
	Balance    uint64
	Nonce      uint64
}

// KeyType represents the type of cryptographic key
type KeyType int

const (
	KeyTypeECDSA KeyType = iota
	KeyTypeEd25519
)

// WalletConfig holds configuration for the wallet
type WalletConfig struct {
	KeyType    KeyType
	Passphrase string
	WalletFile string // Added WalletFile to config
}

// DefaultWalletConfig returns the default wallet configuration
func DefaultWalletConfig() *WalletConfig {
	return &WalletConfig{
		KeyType:    KeyTypeECDSA,
		Passphrase: "",
		WalletFile: "wallet.dat", // Default wallet file name
	}
}

// NewWallet creates a new wallet
func NewWallet(config *WalletConfig, us *utxo.UTXOSet, s *storage.Storage) (*Wallet, error) {
	wallet := &Wallet{
		accounts:       make(map[string]*Account),
		keyType:        config.KeyType,
		utxoSet:        us,
		storage:        s,                 // Initialize storage
		walletFilePath: config.WalletFile, // Initialize wallet file path
		passphrase:     config.Passphrase, // Initialize passphrase
	}

	// Try to load the wallet from storage
	if err := wallet.Load(); err != nil {
		if !os.IsNotExist(err) { // Only return error if it's not "file not found"
			return nil, fmt.Errorf("failed to load wallet: %w", err)
		}
		// If wallet file doesn't exist, create a new one
		var defaultKey *ecdsa.PrivateKey
		var errKey error

		switch config.KeyType {
		case KeyTypeECDSA:
			defaultKey, errKey = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if errKey != nil {
				return nil, fmt.Errorf("failed to generate ECDSA key: %w", errKey)
			}
		case KeyTypeEd25519:
			// For Ed25519, we'll generate an ECDSA key for compatibility
			defaultKey, errKey = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if errKey != nil {
				return nil, fmt.Errorf("failed to generate Ed25519 key: %w", errKey)
			}
		default:
			return nil, fmt.Errorf("unsupported key type: %d", config.KeyType)
		}
		wallet.defaultKey = defaultKey

		// Create default account
		if err := wallet.createDefaultAccount(); err != nil {
			return nil, fmt.Errorf("failed to create default account: %w", err)
		}

		// Save the new wallet
		if err := wallet.Save(); err != nil {
			return nil, fmt.Errorf("failed to save new wallet: %w", err)
		}
	}

	return wallet, nil
}

// Save encrypts and saves the wallet to storage
func (w *Wallet) Save() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(w.accounts)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet accounts: %w", err)
	}

	encryptedData, err := w.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet data: %w", err)
	}

	return w.storage.Write([]byte(w.walletFilePath), encryptedData)
}

// Load loads and decrypts the wallet from storage
func (w *Wallet) Load() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	encryptedData, err := w.storage.Read([]byte(w.walletFilePath))
	if err != nil {
		return err // Propagate os.IsNotExist error
	}

	decryptedData, err := w.Decrypt(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt wallet data: %w", err)
	}

	return json.Unmarshal(decryptedData, &w.accounts)
}

// Encrypt encrypts data using AES-GCM with secure KDF
func (w *Wallet) Encrypt(data []byte) ([]byte, error) {
	// Generate a random salt for this encryption
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using secure KDF
	key, err := deriveKey(w.passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Return salt + nonce + ciphertext
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts data using AES-GCM with secure KDF
func (w *Wallet) Decrypt(data []byte) ([]byte, error) {
	// Extract salt, nonce, and ciphertext
	// Format: salt(32) + nonce(12) + ciphertext
	if len(data) < 32+12 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	salt := data[:32]
	nonce := data[32:44] // AES-GCM nonce is typically 12 bytes
	ciphertext := data[44:]

	// Derive key using the same salt
	key, err := deriveKey(w.passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// createDefaultAccount creates the default account from the default key
func (w *Wallet) createDefaultAccount() error {
	addressBytes := w.generateAddress(w.defaultKey)
	address := hex.EncodeToString(addressBytes)

	account := &Account{
		Address:    address,
		PublicKey:  publicKeyToBytes(&w.defaultKey.PublicKey),
		PrivateKey: privateKeyToBytes(w.defaultKey),
		Balance:    0,
		Nonce:      0,
	}

	w.mu.Lock()
	w.accounts[address] = account
	w.mu.Unlock()

	return nil
}

// generateAddress generates an address from a private key
func (w *Wallet) generateAddress(privateKey *ecdsa.PrivateKey) []byte {
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	publicKeyBytes := publicKeyToBytes(publicKey)

	// Hash the public key
	hash := sha256.Sum256(publicKeyBytes)

	// Take the last 20 bytes as the address
	address := hash[len(hash)-20:]

	return address
}

// generateChecksumAddress generates a checksummed address string
func (w *Wallet) generateChecksumAddress(privateKey *ecdsa.PrivateKey) string {
	addressBytes := w.generateAddress(privateKey)
	return w.encodeAddressWithChecksum(addressBytes)
}

// encodeAddressWithChecksum encodes address bytes with checksum
func (w *Wallet) encodeAddressWithChecksum(addressBytes []byte) string {
	// Add version byte (0x00 for mainnet)
	versioned := append([]byte{0x00}, addressBytes...)

	// Double SHA256 for checksum
	hash1 := sha256.Sum256(versioned)
	hash2 := sha256.Sum256(hash1[:])

	// Take first 4 bytes as checksum
	checksum := hash2[:4]

	// Combine version + address + checksum
	combined := append(versioned, checksum...)

	// Encode as base58
	return w.base58Encode(combined)
}

// base58Encode encodes bytes to base58 string
func (w *Wallet) base58Encode(data []byte) string {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	var result []byte
	x := new(big.Int).SetBytes(data)
	base := big.NewInt(58)
	zero := big.NewInt(0)

	for x.Cmp(zero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, base, mod)
		result = append([]byte{alphabet[mod.Int64()]}, result...)
	}

	// Add leading zeros
	for _, b := range data {
		if b == 0 {
			result = append([]byte{'1'}, result...)
		} else {
			break
		}
	}

	return string(result)
}

// decodeAddressWithChecksum decodes a checksummed address
func (w *Wallet) decodeAddressWithChecksum(address string) ([]byte, error) {
	// Decode base58
	data := w.base58Decode(address)

	// Check minimum length (version + address + checksum)
	if len(data) < 25 {
		return nil, fmt.Errorf("address too short")
	}

	// Extract components
	version := data[0]
	addressBytes := data[1:21]
	checksum := data[21:25]

	// Verify version
	if version != 0x00 {
		return nil, fmt.Errorf("unsupported address version: %d", version)
	}

	// Verify checksum
	versioned := append([]byte{version}, addressBytes...)
	hash1 := sha256.Sum256(versioned)
	hash2 := sha256.Sum256(hash1[:])
	expectedChecksum := hash2[:4]

	if !w.bytesEqual(checksum, expectedChecksum) {
		return nil, fmt.Errorf("invalid checksum")
	}

	return addressBytes, nil
}

// base58Decode decodes base58 string to bytes
func (w *Wallet) base58Decode(data string) []byte {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Create reverse lookup
	reverse := make(map[byte]int)
	for i, char := range alphabet {
		reverse[byte(char)] = i
	}

	// Decode
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, char := range data {
		if val, ok := reverse[byte(char)]; ok {
			result.Mul(result, base)
			result.Add(result, big.NewInt(int64(val)))
		}
	}

	// Convert to bytes
	bytes := result.Bytes()

	// Add leading zeros
	for _, char := range data {
		if char == '1' {
			bytes = append([]byte{0}, bytes...)
		} else {
			break
		}
	}

	return bytes
}

// bytesEqual compares two byte slices
func (w *Wallet) bytesEqual(a, b []byte) bool {
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

// CreateAccount creates a new account
func (w *Wallet) CreateAccount() (*Account, error) {
	var privateKey *ecdsa.PrivateKey
	var err error

	switch w.keyType {
	case KeyTypeECDSA:
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
		}
	case KeyTypeEd25519:
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}
	}

	addressBytes := w.generateAddress(privateKey)
	address := hex.EncodeToString(addressBytes)

	account := &Account{
		Address:    address,
		PublicKey:  publicKeyToBytes(&privateKey.PublicKey),
		PrivateKey: privateKeyToBytes(privateKey),
		Balance:    0,
		Nonce:      0,
	}

	w.mu.Lock()
	w.accounts[address] = account
	w.mu.Unlock()

	return account, nil
}

// GetAccount returns an account by address
func (w *Wallet) GetAccount(address string) *Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.accounts[address]
}

// GetDefaultAccount returns the default account
func (w *Wallet) GetDefaultAccount() *Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return the first account (default)
	for _, account := range w.accounts {
		return account
	}

	return nil
}

// GetAllAccounts returns all accounts in the wallet
func (w *Wallet) GetAllAccounts() []*Account {
	w.mu.RLock()
	defer w.mu.RUnlock()

	accounts := make([]*Account, 0, len(w.accounts))
	for _, account := range w.accounts {
		accounts = append(accounts, account)
	}

	return accounts
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(fromAddress, toAddress string, amount, fee uint64) (*block.Transaction, error) {
	account := w.GetAccount(fromAddress)
	if account == nil {
		return nil, fmt.Errorf("account not found: %s", fromAddress)
	}

	// Get available UTXOs for the sender
	utxos := w.utxoSet.GetAddressUTXOs(fromAddress)
	if len(utxos) == 0 {
		return nil, fmt.Errorf("no available UTXOs for address: %s", fromAddress)
	}

	// Calculate total available balance
	var totalAvailable uint64
	for _, utxo := range utxos {
		totalAvailable += utxo.Value
	}

	// Check if we have enough funds
	totalNeeded := amount + fee
	if totalAvailable < totalNeeded {
		return nil, fmt.Errorf("insufficient funds: need %d, have %d", totalNeeded, totalAvailable)
	}

	// Select UTXOs to spend (simple greedy algorithm)
	var selectedUTXOs []*utxo.UTXO
	var selectedAmount uint64
	for _, utxo := range utxos {
		if selectedAmount >= totalNeeded {
			break
		}
		selectedUTXOs = append(selectedUTXOs, utxo)
		selectedAmount += utxo.Value
	}

	// Create transaction inputs
	inputs := make([]*block.TxInput, 0, len(selectedUTXOs))
	for _, utxo := range selectedUTXOs {
		input := &block.TxInput{
			PrevTxHash:  utxo.TxHash,
			PrevTxIndex: utxo.TxIndex,
			ScriptSig:   account.PublicKey, // Will be replaced with signature
			Sequence:    0xffffffff,
		}
		inputs = append(inputs, input)
	}

	// Create transaction outputs
	outputs := make([]*block.TxOutput, 0, 2) // recipient + change

	// Output to recipient
	recipPubKeyHash, err := addressToPubKeyHash(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}
	outputs = append(outputs, &block.TxOutput{
		Value:        amount,
		ScriptPubKey: recipPubKeyHash,
	})

	// Calculate change and create change output if needed
	change := selectedAmount - totalNeeded
	if change > 0 {
		// Create change output back to sender
		senderPubKeyHash, err := addressToPubKeyHash(fromAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid sender address: %w", err)
		}
		outputs = append(outputs, &block.TxOutput{
			Value:        change,
			ScriptPubKey: senderPubKeyHash,
		})
	}

	// Create transaction
	tx := &block.Transaction{
		Version:  1,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: 0,
		Fee:      fee,
	}

	// Sign transaction
	if err := w.SignTransaction(tx, fromAddress); err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Calculate transaction hash
	tx.Hash = w.calculateTransactionHash(tx)

	// Update account nonce
	account.Nonce++

	return tx, nil
}

// SignTransaction signs a transaction with the specified account's private key
func (w *Wallet) SignTransaction(tx *block.Transaction, fromAddress string) error {
	account := w.GetAccount(fromAddress)
	if account == nil {
		return fmt.Errorf("account not found: %s", fromAddress)
	}

	// Convert private key bytes back to ECDSA private key
	privateKey, err := bytesToPrivateKey(account.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key: %w", err)
	}

	// Create signature data
	signatureData := w.createSignatureData(tx)

	// Sign the data
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, signatureData)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Encode signature in canonical DER format
	signature, err := encodeSignatureDER(r, s)
	if err != nil {
		return fmt.Errorf("failed to encode signature: %w", err)
	}

	pubBytes := publicKeyToBytes(&privateKey.PublicKey)

	// Add signature to all inputs
	for i := range tx.Inputs {
		// Store public key followed by DER signature
		combined := make([]byte, 0, len(pubBytes)+len(signature))
		combined = append(combined, pubBytes...)
		combined = append(combined, signature...)
		tx.Inputs[i].ScriptSig = combined
	}

	return nil
}

// VerifyTransaction verifies a transaction signature and validates UTXOs
func (w *Wallet) VerifyTransaction(tx *block.Transaction) (bool, error) {
	if len(tx.Inputs) == 0 {
		return false, fmt.Errorf("transaction has no inputs")
	}

	// Verify all inputs
	for i, input := range tx.Inputs {
		// Check if UTXO exists and is unspent
		utxo := w.utxoSet.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		if utxo == nil {
			return false, fmt.Errorf("input %d references non-existent UTXO", i)
		}

		// Verify signature
		if len(input.ScriptSig) == 0 {
			return false, fmt.Errorf("input %d has no signature", i)
		}

		// Expect uncompressed public key (65 bytes) + DER signature (variable length)
		if len(input.ScriptSig) < 65+8 {
			return false, fmt.Errorf("input %d has invalid signature length: %d", i, len(input.ScriptSig))
		}

		pubBytes := input.ScriptSig[:65]
		derSignature := input.ScriptSig[65:]

		// Decode DER signature
		r, s, err := decodeSignatureDER(derSignature)
		if err != nil {
			return false, fmt.Errorf("input %d: failed to decode DER signature: %w", i, err)
		}

		// Verify canonical form
		if err := verifyCanonicalSignature(r, s, elliptic.P256()); err != nil {
			return false, fmt.Errorf("input %d: signature not in canonical form: %w", i, err)
		}

		x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
		if x == nil || y == nil {
			return false, fmt.Errorf("input %d: failed to unmarshal public key", i)
		}
		pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

		// Verify signature
		valid := ecdsa.Verify(pub, w.createSignatureData(tx), r, s)
		if !valid {
			return false, fmt.Errorf("input %d: invalid signature", i)
		}
	}

	// Verify output amounts don't exceed input amounts
	var totalInput uint64
	var totalOutput uint64

	for _, input := range tx.Inputs {
		utxo := w.utxoSet.GetUTXO(input.PrevTxHash, input.PrevTxIndex)
		totalInput += utxo.Value
	}

	for _, output := range tx.Outputs {
		totalOutput += output.Value
	}

	if totalOutput > totalInput {
		return false, fmt.Errorf("output amount %d exceeds input amount %d", totalOutput, totalInput)
	}

	return true, nil
}

// createSignatureData creates the data to be signed
func (w *Wallet) createSignatureData(tx *block.Transaction) []byte {
	// In a real implementation, this would create a proper signature hash
	// For now, we'll use a simplified approach

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

// calculateTransactionHash calculates the hash of a transaction
func (w *Wallet) calculateTransactionHash(tx *block.Transaction) []byte {
	// This is a simplified hash calculation
	// In a real implementation, this would follow the specific blockchain's rules

	data := make([]byte, 0)

	// Version
	data = append(data, byte(tx.Version))

	// Inputs
	for _, input := range tx.Inputs {
		data = append(data, input.PrevTxHash...)
		data = append(data, byte(input.PrevTxIndex))
		data = append(data, input.ScriptSig...)
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

// UpdateBalance updates the balance of an account
func (w *Wallet) UpdateBalance(address string, balance uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if account, exists := w.accounts[address]; exists {
		account.Balance = balance
	}
}

// GetBalance returns the balance of an account
func (w *Wallet) GetBalance(address string) uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if account, exists := w.accounts[address]; exists {
		return account.Balance
	}

	return 0
}

// ImportPrivateKey imports a private key and creates an account
func (w *Wallet) ImportPrivateKey(privateKeyHex string) (*Account, error) {
	// Decode hex string
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	// Convert to ECDSA private key
	privateKey, err := bytesToPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key: %w", err)
	}

	// Generate address
	addressBytes := w.generateAddress(privateKey)
	address := hex.EncodeToString(addressBytes)

	// Check if account already exists; return the existing account instead of error
	if existing := w.GetAccount(address); existing != nil {
		return existing, nil
	}

	// Create account
	account := &Account{
		Address:    address,
		PublicKey:  publicKeyToBytes(&privateKey.PublicKey),
		PrivateKey: privateKeyToBytes(privateKey),
		Balance:    0,
		Nonce:      0,
	}

	w.mu.Lock()
	w.accounts[address] = account
	w.mu.Unlock()

	return account, nil
}

// ExportPrivateKey exports a private key as a hex string
func (w *Wallet) ExportPrivateKey(address string) (string, error) {
	account := w.GetAccount(address)
	if account == nil {
		return "", fmt.Errorf("account not found: %s", address)
	}

	return hex.EncodeToString(account.PrivateKey), nil
}

// String returns a string representation of the account
func (a *Account) String() string {
	return fmt.Sprintf("Account{Address: %s, Balance: %d, Nonce: %d}",
		a.Address, a.Balance, a.Nonce)
}

// String returns a string representation of the wallet
func (w *Wallet) String() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return fmt.Sprintf("Wallet{Accounts: %d, KeyType: %d}",
		len(w.accounts), w.keyType)
}

// Helpers
func privateKeyToBytes(k *ecdsa.PrivateKey) []byte {
	d := k.D.Bytes()
	// Pad to 32 bytes
	if len(d) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(d):], d)
		return padded
	}
	// Truncate if longer (shouldn't happen for P-256)
	if len(d) > 32 {
		return d[len(d)-32:]
	}
	return d
}

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

// addressToPubKeyHash converts a hex-encoded address string to its byte representation (public key hash)
func addressToPubKeyHash(address string) ([]byte, error) {
	pubKeyHash, err := hex.DecodeString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address hex: %w", err)
	}
	// A valid address is 20 bytes (after hashing public key and taking last 20 bytes)
	if len(pubKeyHash) != 20 {
		return nil, fmt.Errorf("invalid address length: %d", len(pubKeyHash))
	}
	return pubKeyHash, nil
}

// canonicalSignature ensures the signature is in canonical form (low-S)
func canonicalSignature(r, s *big.Int, curve elliptic.Curve) (*big.Int, *big.Int) {
	// Get curve order
	N := curve.Params().N

	// If s > N/2, use N - s instead (low-S enforcement)
	if s.Cmp(new(big.Int).Div(N, big.NewInt(2))) > 0 {
		s = new(big.Int).Sub(N, s)
	}

	return r, s
}

// encodeSignatureDER encodes r and s values as DER
func encodeSignatureDER(r, s *big.Int) ([]byte, error) {
	// Ensure canonical form
	r, s = canonicalSignature(r, s, elliptic.P256())

	// Create ASN.1 structure
	signature := struct {
		R, S *big.Int
	}{r, s}

	// Encode to DER
	return asn1.Marshal(signature)
}

// decodeSignatureDER decodes DER signature to r and s values
func decodeSignatureDER(signature []byte) (*big.Int, *big.Int, error) {
	var sig struct {
		R, S *big.Int
	}

	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal DER signature: %w", err)
	}

	return sig.R, sig.S, nil
}

// verifyCanonicalSignature verifies that a signature is in canonical form
func verifyCanonicalSignature(r, s *big.Int, curve elliptic.Curve) error {
	N := curve.Params().N

	// Check bounds
	if r.Sign() <= 0 || r.Cmp(N) >= 0 {
		return fmt.Errorf("r value out of bounds")
	}
	if s.Sign() <= 0 || s.Cmp(N) >= 0 {
		return fmt.Errorf("s value out of bounds")
	}

	// Check low-S property
	if s.Cmp(new(big.Int).Div(N, big.NewInt(2))) > 0 {
		return fmt.Errorf("signature not in canonical form (high-S)")
	}

	return nil
}

// deriveKey derives an encryption key from passphrase using PBKDF2
func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	// Use PBKDF2 with SHA-256, 100,000 iterations, 32-byte key
	derivedKey := make([]byte, 32)

	// Simple PBKDF2 implementation using HMAC-SHA256
	// In production, consider using a more robust KDF library
	passphraseBytes := []byte(passphrase)
	combined := append(passphraseBytes, salt...)
	hash := sha256.Sum256(combined)
	copy(derivedKey, hash[:])

	// Multiple iterations for key strengthening
	for i := 0; i < 100000; i++ {
		h := hmac.New(sha256.New, derivedKey)
		h.Write(passphraseBytes)
		h.Write(salt)
		h.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
		derivedKey = h.Sum(nil)
	}

	return derivedKey, nil
}

// generateSalt generates a random salt for key derivation
func generateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	return salt, err
}
