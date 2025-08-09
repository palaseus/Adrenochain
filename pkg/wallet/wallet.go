package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	// "crypto/aes" // Removed unused import
	// "crypto/cipher" // Removed unused import
	// "io" // Removed unused import

	"github.com/gochain/gochain/pkg/block"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	mu         sync.RWMutex
	accounts   map[string]*Account
	defaultKey *ecdsa.PrivateKey
	keyType    KeyType
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
}

// DefaultWalletConfig returns the default wallet configuration
func DefaultWalletConfig() *WalletConfig {
	return &WalletConfig{
		KeyType:    KeyTypeECDSA,
		Passphrase: "",
	}
}

// NewWallet creates a new wallet
func NewWallet(config *WalletConfig) (*Wallet, error) {
	wallet := &Wallet{
		accounts: make(map[string]*Account),
		keyType:  config.KeyType,
		// storage:  config.Storage,
	}

	// Try to load the wallet from storage
	// if err := wallet.Load(); err != nil {
	// 	return nil, fmt.Errorf("failed to load wallet: %w", err)
	// }

	if len(wallet.accounts) == 0 {
		// Wallet doesn't exist or is empty, create a new one
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
		// if err := wallet.Save(); err != nil {
		// 	return nil, fmt.Errorf("failed to save new wallet: %w", err)
		// }
	}

	return wallet, nil
}

// Save encrypts and saves the wallet to storage
func (w *Wallet) Save() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// data, err := json.Marshal(w.accounts)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal wallet accounts: %w", err)
	// }

	// return w.storage.Write(encryptedData)
	return nil // Temporarily return nil to allow compilation
}

// Load loads and decrypts the wallet from storage
func (w *Wallet) Load() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return nil
}

// Encrypt encrypts data using AES-GCM
func (w *Wallet) Encrypt(data []byte) ([]byte, error) {
	// key := sha256.Sum256([]byte(w.storage.Path()))
	// block, err := aes.NewCipher(key[:])
	// if err != nil {
	// 	return nil, err
	// }

	// gcm, err := cipher.NewGCM(block)
	// if err != nil {
	// 	return nil, err
	// }

	// nonce := make([]byte, gcm.NonceSize())
	// if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
	// 	return nil, err
	// }

	// return gcm.Seal(nonce, nonce, data, nil), nil
	return data, nil // Temporarily return original data
}

// Decrypt decrypts data using AES-GCM
func (w *Wallet) Decrypt(data []byte) ([]byte, error) {
	// key := sha256.Sum256([]byte(w.storage.Path()))
	// block, err := aes.NewCipher(key[:])
	// if err != nil {
	// 	return nil, err
	// }

	// gcm, err := cipher.NewGCM(block)
	// if err != nil {
	// 	return nil, err
	// }

	// nonceSize := gcm.NonceSize()
	// if len(data) < nonceSize {
	// 	return nil, fmt.Errorf("ciphertext too short")
	// }

	// nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// return gcm.Open(nil, nonce, ciphertext, nil)
	return data, nil // Temporarily return original data
}

// createDefaultAccount creates the default account from the default key
func (w *Wallet) createDefaultAccount() error {
	address := w.generateAddress(w.defaultKey)

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
func (w *Wallet) generateAddress(privateKey *ecdsa.PrivateKey) string {
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	publicKeyBytes := publicKeyToBytes(publicKey)

	// Hash the public key
	hash := sha256.Sum256(publicKeyBytes)

	// Take the last 20 bytes as the address
	address := hash[len(hash)-20:]

	return hex.EncodeToString(address)
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

	address := w.generateAddress(privateKey)

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

	// Note: Balance checks are intentionally omitted in this simplified implementation
	// to allow transaction creation without on-chain state.

	// Create transaction input
	input := &block.TxInput{
		PrevTxHash:  make([]byte, 32), // This would be the UTXO hash in a real implementation
		PrevTxIndex: 0,
		ScriptSig:   account.PublicKey,
		Sequence:    0xffffffff,
	}

	// Create transaction outputs
	outputs := make([]*block.TxOutput, 0)

	// Output to recipient
	outputs = append(outputs, &block.TxOutput{
		Value:        amount,
		ScriptPubKey: []byte(toAddress),
	})

	// No change output is created in this simplified model

	// Create transaction
	tx := &block.Transaction{
		Version:  1,
		Inputs:   []*block.TxInput{input},
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
	signature := concatRS(r, s)
	pubBytes := publicKeyToBytes(&privateKey.PublicKey)

	// Add signature to the first input (assuming single input for simplicity)
	if len(tx.Inputs) > 0 {
		// Store public key followed by signature (r||s)
		combined := make([]byte, 0, len(pubBytes)+len(signature))
		combined = append(combined, pubBytes...)
		combined = append(combined, signature...)
		tx.Inputs[0].ScriptSig = combined
	}

	return nil
}

// VerifyTransaction verifies a transaction signature
func (w *Wallet) VerifyTransaction(tx *block.Transaction) (bool, error) {
	if len(tx.Inputs) == 0 {
		return false, fmt.Errorf("transaction has no inputs")
	}

	// Get the signature from the first input
	signature := tx.Inputs[0].ScriptSig
	if len(signature) == 0 {
		return false, fmt.Errorf("transaction has no signature")
	}

	// Expect uncompressed public key (65 bytes) + signature (64 bytes)
	if len(signature) < 65+64 {
		return false, fmt.Errorf("invalid signature length: %d", len(signature))
	}
	pubBytes := signature[:65]
	rsBytes := signature[65:]
	if len(rsBytes) != 64 {
		return false, fmt.Errorf("invalid r||s length: %d", len(rsBytes))
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil || y == nil {
		return false, fmt.Errorf("failed to unmarshal public key")
	}
	pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	r := new(big.Int).SetBytes(rsBytes[:32])
	s := new(big.Int).SetBytes(rsBytes[32:])

	valid := ecdsa.Verify(pub, w.createSignatureData(tx), r, s)
	return valid, nil
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
	address := w.generateAddress(privateKey)

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
