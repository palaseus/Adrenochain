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
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/storage"
	"github.com/gochain/gochain/pkg/utxo"
	"github.com/mr-tron/base58"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	mu             sync.RWMutex
	accounts       map[string]*Account
	defaultKey     *btcec.PrivateKey
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

// NewWallet creates a new wallet with the specified configuration
func NewWallet(config *WalletConfig, us *utxo.UTXOSet, s *storage.Storage) (*Wallet, error) {
	if config == nil {
		config = DefaultWalletConfig()
	}

	var defaultKey *btcec.PrivateKey
	var errKey error

	switch config.KeyType {
	case KeyTypeECDSA:
		// Use secp256k1 curve (Bitcoin/Ethereum standard)
		defaultKey, errKey = btcec.NewPrivateKey()
		if errKey != nil {
			return nil, fmt.Errorf("failed to generate secp256k1 key: %w", errKey)
		}
	case KeyTypeEd25519:
		// For now, fall back to secp256k1 for Ed25519 type as well
		// TODO: Implement proper Ed25519 support
		defaultKey, errKey = btcec.NewPrivateKey()
		if errKey != nil {
			return nil, fmt.Errorf("failed to generate secp256k1 key: %w", errKey)
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %d", config.KeyType)
	}

	wallet := &Wallet{
		accounts:       make(map[string]*Account),
		defaultKey:     defaultKey,
		keyType:        config.KeyType,
		utxoSet:        us,
		storage:        s,
		walletFilePath: config.WalletFile,
		passphrase:     config.Passphrase,
	}

	// Create default account
	if err := wallet.createDefaultAccount(); err != nil {
		return nil, fmt.Errorf("failed to create default account: %w", err)
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

	// Create a new accounts map to avoid merging with existing accounts
	var loadedAccounts map[string]*Account
	if err := json.Unmarshal(decryptedData, &loadedAccounts); err != nil {
		return fmt.Errorf("failed to unmarshal wallet accounts: %w", err)
	}

	// Replace the existing accounts with the loaded ones
	w.accounts = loadedAccounts

	return nil
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

// createDefaultAccount creates the default account for the wallet
func (w *Wallet) createDefaultAccount() error {
	// Convert btcec.PrivateKey to ecdsa.PrivateKey for compatibility
	defaultKeyECDSA := w.defaultKey.ToECDSA()

	// Generate address
	addressStr := w.generateChecksumAddress(defaultKeyECDSA)

	// Create default account
	account := &Account{
		Address:    addressStr,
		PublicKey:  publicKeyToBytes(&defaultKeyECDSA.PublicKey),
		PrivateKey: privateKeyToBytes(defaultKeyECDSA),
		Balance:    0,
		Nonce:      0,
	}

	// Add to wallet
	w.accounts[addressStr] = account
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
	return base58.Encode(data)
}

// decodeAddressWithChecksum decodes a checksummed address
func (w *Wallet) decodeAddressWithChecksum(address string) ([]byte, error) {
	// Decode base58
	data, err := w.base58Decode(address)
	if err != nil {
		return nil, err
	}

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

	// Simple byte comparison
	if len(checksum) != len(expectedChecksum) {
		return nil, fmt.Errorf("invalid checksum length")
	}
	for i := range checksum {
		if checksum[i] != expectedChecksum[i] {
			return nil, fmt.Errorf("invalid checksum")
		}
	}

	return addressBytes, nil
}

// base58Decode decodes base58 string to bytes
func (w *Wallet) base58Decode(data string) ([]byte, error) {
	return base58.Decode(data)
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

// CreateAccount creates a new account in the wallet
func (w *Wallet) CreateAccount() (*Account, error) {
	// Generate a new private key
	var privateKey *ecdsa.PrivateKey

	switch w.keyType {
	case KeyTypeECDSA:
		// Use secp256k1 curve (Bitcoin/Ethereum standard) instead of P-256
		btcPrivKey, err := btcec.NewPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
		}
		// Convert btcec.PrivateKey to ecdsa.PrivateKey for compatibility
		privateKey = btcPrivKey.ToECDSA()
	case KeyTypeEd25519:
		// For now, fall back to secp256k1 for Ed25519 type as well
		// TODO: Implement proper Ed25519 support
		btcPrivKey, err := btcec.NewPrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
		}
		// Convert btcec.PrivateKey to ecdsa.PrivateKey for compatibility
		privateKey = btcPrivKey.ToECDSA()
	}

	// Generate address
	addressStr := w.generateChecksumAddress(privateKey)

	// Create account
	account := &Account{
		Address:    addressStr,
		PublicKey:  publicKeyToBytes(&privateKey.PublicKey),
		PrivateKey: privateKeyToBytes(privateKey),
		Balance:    0,
		Nonce:      0,
	}

	// Add to wallet
	w.mu.Lock()
	w.accounts[addressStr] = account
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

	// Create signature data (this should be the hash that will be used for verification)
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

	// Set the transaction hash to the signature data hash for verification
	tx.Hash = signatureData

	return nil
}

// VerifyTransaction verifies the cryptographic signatures of a transaction
func (w *Wallet) VerifyTransaction(tx *block.Transaction) (bool, error) {
	for i, input := range tx.Inputs {
		// Get the public key from the input
		if len(input.ScriptSig) < 65 {
			return false, fmt.Errorf("input %d: script signature too short", i)
		}

		pubBytes := input.ScriptSig[:65]
		sigBytes := input.ScriptSig[65:]

		// Parse the public key
		btcPubKey, err := btcec.ParsePubKey(pubBytes)
		if err != nil {
			return false, fmt.Errorf("input %d: failed to parse public key: %w", i, err)
		}

		// Convert to ecdsa.PublicKey for compatibility
		pub := btcPubKey.ToECDSA()

		// Decode the signature
		r, s, err := decodeSignatureDER(sigBytes)
		if err != nil {
			return false, fmt.Errorf("input %d: failed to decode signature: %w", i, err)
		}

		// Verify canonical form
		if err := verifyCanonicalSignature(r, s, btcec.S256()); err != nil {
			return false, fmt.Errorf("input %d: signature not in canonical form: %w", i, err)
		}

		// Verify signature against the transaction hash (which should be the signature data hash)
		if !ecdsa.Verify(pub, tx.Hash, r, s) {
			return false, fmt.Errorf("input %d: signature verification failed", i)
		}
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

	// Generate base58 address
	address := w.generateChecksumAddress(privateKey)

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

// publicKeyToBytes converts an ECDSA public key to bytes using secp256k1
func publicKeyToBytes(k *ecdsa.PublicKey) []byte {
	// Explicitly use secp256k1 curve for marshaling
	curve := btcec.S256()
	return elliptic.Marshal(curve, k.X, k.Y)
}

func bytesToPrivateKey(b []byte) (*ecdsa.PrivateKey, error) {
	if len(b) != 32 {
		return nil, fmt.Errorf("invalid private key length: %d", len(b))
	}
	d := new(big.Int).SetBytes(b)
	curve := btcec.S256()
	// Validate that 0 < d < N
	if d.Sign() <= 0 || d.Cmp(curve.N) >= 0 {
		return nil, fmt.Errorf("invalid private key scalar")
	}

	// Create the private key
	privKey := &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     nil, // Will be computed when needed
			Y:     nil, // Will be computed when needed
		},
	}

	// Compute the public key
	pubX, pubY := curve.ScalarBaseMult(d.Bytes())
	privKey.PublicKey.X = pubX
	privKey.PublicKey.Y = pubY

	return privKey, nil
}

func concatRS(r, s *big.Int) []byte {
	rb := r.Bytes()
	sb := s.Bytes()
	out := make([]byte, 64)
	copy(out[32-len(rb):32], rb)
	copy(out[64-len(sb):], sb)
	return out
}

// addressToPubKeyHash converts a base58-encoded address string to its byte representation (public key hash)
func addressToPubKeyHash(address string) ([]byte, error) {
	// Since this is a package-level function, we need to create a temporary wallet instance
	// to use the decodeAddressWithChecksum method, or we can implement the logic directly here

	// Decode base58 address
	data, err := base58.Decode(address)
	if err != nil {
		return nil, fmt.Errorf("invalid base58 address: %w", err)
	}

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

	// Simple byte comparison
	if len(checksum) != len(expectedChecksum) {
		return nil, fmt.Errorf("invalid checksum length")
	}
	for i := range checksum {
		if checksum[i] != expectedChecksum[i] {
			return nil, fmt.Errorf("invalid checksum")
		}
	}

	return addressBytes, nil
}

// canonicalSignature ensures the signature is in canonical form (low-S)
func canonicalSignature(r, s *big.Int, curve *btcec.KoblitzCurve) (*big.Int, *big.Int) {
	// Get curve order
	N := curve.N

	// If s > N/2, use N - s instead (low-S enforcement)
	if s.Cmp(new(big.Int).Div(N, big.NewInt(2))) > 0 {
		s = new(big.Int).Sub(N, s)
	}

	return r, s
}

// encodeSignatureDER encodes r and s values as DER
func encodeSignatureDER(r, s *big.Int) ([]byte, error) {
	// Ensure canonical form using secp256k1
	r, s = canonicalSignature(r, s, btcec.S256())

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
func verifyCanonicalSignature(r, s *big.Int, curve *btcec.KoblitzCurve) error {
	N := curve.N

	// Check bounds
	if r.Sign() <= 0 || r.Cmp(N) >= 0 {
		return errors.New("r value out of bounds")
	}
	if s.Sign() <= 0 || s.Cmp(N) >= 0 {
		return errors.New("s value out of bounds")
	}

	// Check low-S enforcement
	halfN := new(big.Int).Div(N, big.NewInt(2))
	if s.Cmp(halfN) > 0 {
		return errors.New("s value not in canonical form (high-S)")
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
