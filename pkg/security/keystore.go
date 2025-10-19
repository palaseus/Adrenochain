package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KeyStore provides secure storage and management of cryptographic keys
type KeyStore struct {
	mu       sync.RWMutex
	keys     map[string]*StoredKey
	filePath string
	password []byte
}

// StoredKey represents a stored cryptographic key
type StoredKey struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	PublicKey   []byte                 `json:"public_key"`
	PrivateKey  []byte                 `json:"private_key,omitempty"` // Encrypted
	CreatedAt   time.Time              `json:"created_at"`
	LastUsed    time.Time              `json:"last_used"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// KeyType represents the type of cryptographic key
type KeyType string

const (
	KeyTypeECDSA   KeyType = "ecdsa"
	KeyTypeRSA     KeyType = "rsa"
	KeyTypeEd25519 KeyType = "ed25519"
)

// NewKeyStore creates a new secure key store
func NewKeyStore(filePath string, password []byte) (*KeyStore, error) {
	ks := &KeyStore{
		keys:     make(map[string]*StoredKey),
		filePath: filePath,
		password: password,
	}

	// Load existing keys if file exists
	if err := ks.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}

	return ks, nil
}

// GenerateKeyPair generates a new ECDSA key pair
func (ks *KeyStore) GenerateKeyPair(keyID, description string) (*ecdsa.PrivateKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	// Check if key already exists
	if _, exists := ks.keys[keyID]; exists {
		return nil, fmt.Errorf("key with ID %s already exists", keyID)
	}

	// Generate new ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Encode private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encrypt private key
	encryptedPrivateKey, err := ks.encryptData(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Store the key
	now := time.Now()
	storedKey := &StoredKey{
		ID:          keyID,
		Type:        string(KeyTypeECDSA),
		PublicKey:   publicKeyBytes,
		PrivateKey:  encryptedPrivateKey,
		CreatedAt:   now,
		LastUsed:    now,
		Description: description,
		Metadata:    make(map[string]interface{}),
	}

	ks.keys[keyID] = storedKey

	// Save to file
	if err := ks.saveKeys(); err != nil {
		return nil, fmt.Errorf("failed to save keys: %w", err)
	}

	return privateKey, nil
}

// GetPrivateKey retrieves and decrypts a private key
func (ks *KeyStore) GetPrivateKey(keyID string) (*ecdsa.PrivateKey, error) {
	ks.mu.RLock()
	storedKey, exists := ks.keys[keyID]
	ks.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key with ID %s not found", keyID)
	}

	// Decrypt private key
	privateKeyBytes, err := ks.decryptData(storedKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA")
	}

	// Update last used time
	ks.mu.Lock()
	storedKey.LastUsed = time.Now()
	ks.mu.Unlock()

	return ecdsaKey, nil
}

// GetPublicKey retrieves a public key
func (ks *KeyStore) GetPublicKey(keyID string) (*ecdsa.PublicKey, error) {
	ks.mu.RLock()
	storedKey, exists := ks.keys[keyID]
	ks.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key with ID %s not found", keyID)
	}

	// Parse public key
	publicKey, err := x509.ParsePKIXPublicKey(storedKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA")
	}

	return ecdsaKey, nil
}

// ListKeys returns a list of all stored key IDs
func (ks *KeyStore) ListKeys() []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keyIDs := make([]string, 0, len(ks.keys))
	for keyID := range ks.keys {
		keyIDs = append(keyIDs, keyID)
	}

	return keyIDs
}

// DeleteKey removes a key from the store
func (ks *KeyStore) DeleteKey(keyID string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if _, exists := ks.keys[keyID]; !exists {
		return fmt.Errorf("key with ID %s not found", keyID)
	}

	delete(ks.keys, keyID)

	// Save to file
	if err := ks.saveKeys(); err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	return nil
}

// SignData signs data with the specified private key
func (ks *KeyStore) SignData(keyID string, data []byte) ([]byte, error) {
	privateKey, err := ks.GetPrivateKey(keyID)
	if err != nil {
		return nil, err
	}

	// Hash the data
	hash := sha256.Sum256(data)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode signature as DER
	signature, err := ks.encodeSignature(r, s)
	if err != nil {
		return nil, fmt.Errorf("failed to encode signature: %w", err)
	}

	return signature, nil
}

// VerifySignature verifies a signature
func (ks *KeyStore) VerifySignature(keyID string, data, signature []byte) (bool, error) {
	publicKey, err := ks.GetPublicKey(keyID)
	if err != nil {
		return false, err
	}

	// Hash the data
	hash := sha256.Sum256(data)

	// Decode signature
	r, s, err := ks.decodeSignature(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// encryptData encrypts data using AES-GCM
func (ks *KeyStore) encryptData(data []byte) ([]byte, error) {
	// Derive key from password
	key := sha256.Sum256(ks.password)

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// decryptData decrypts data using AES-GCM
func (ks *KeyStore) decryptData(data []byte) ([]byte, error) {
	// Derive key from password
	key := sha256.Sum256(ks.password)

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// encodeSignature encodes an ECDSA signature as DER
func (ks *KeyStore) encodeSignature(r, s *big.Int) ([]byte, error) {
	// Create ASN.1 structure
	sig := struct {
		R *big.Int
		S *big.Int
	}{R: r, S: s}

	return asn1.Marshal(sig)
}

// decodeSignature decodes a DER-encoded ECDSA signature
func (ks *KeyStore) decodeSignature(signature []byte) (*big.Int, *big.Int, error) {
	// Parse ASN.1 structure
	var sig struct {
		R *big.Int
		S *big.Int
	}

	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		return nil, nil, err
	}

	return sig.R, sig.S, nil
}

// saveKeys saves keys to file
func (ks *KeyStore) saveKeys() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(ks.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Marshal keys to JSON
	data, err := json.MarshalIndent(ks.keys, "", "  ")
	if err != nil {
		return err
	}

	// Write to file with restricted permissions
	return os.WriteFile(ks.filePath, data, 0600)
}

// loadKeys loads keys from file
func (ks *KeyStore) loadKeys() error {
	// Check if file exists
	if _, err := os.Stat(ks.filePath); os.IsNotExist(err) {
		return nil // No keys to load
	}

	// Read file
	data, err := os.ReadFile(ks.filePath)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	return json.Unmarshal(data, &ks.keys)
}
