package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex" // Added import
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/palaseus/adrenochain/pkg/block"
)

// Storage implements a file-based storage for blocks and chain state.
type Storage struct {
	dataDir string

	// SECURITY FIX: Encryption support
	encryptionKey    []byte
	enableEncryption bool
}

// StorageConfig holds configuration for storage.
type StorageConfig struct {
	DataDir string

	// SECURITY FIX: Encryption configuration
	EnableEncryption bool
	EncryptionKey    []byte
}

// DefaultStorageConfig returns the default storage configuration.
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{DataDir: "./data"}
}

// WithDataDir sets the data directory for the storage config.
func (c *StorageConfig) WithDataDir(dataDir string) *StorageConfig {
	newConfig := &StorageConfig{
		DataDir: dataDir,
	}
	return newConfig
}

// NewStorage creates a new file-based storage.
func NewStorage(storageConfig *StorageConfig) (*Storage, error) {
	// SECURITY FIX: Use secure file permissions (owner read/write only)
	filePerms := os.FileMode(0700) // Owner read/write/execute only
	if err := os.MkdirAll(storageConfig.DataDir, filePerms); err != nil {
		return nil, err
	}

	// SECURITY FIX: Generate encryption key if not provided
	encryptionKey := storageConfig.EncryptionKey
	if storageConfig.EnableEncryption && len(encryptionKey) == 0 {
		encryptionKey = make([]byte, 32) // 256-bit key
		if _, err := rand.Read(encryptionKey); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
	}

	return &Storage{
		dataDir:          storageConfig.DataDir,
		encryptionKey:    encryptionKey,
		enableEncryption: storageConfig.EnableEncryption,
	}, nil
}

// StoreBlock stores a block to a file.
func (s *Storage) StoreBlock(b *block.Block) error {
	if b == nil {
		return fmt.Errorf("cannot store nil block")
	}

	file, err := os.Create(filepath.Join(s.dataDir, b.HexHash()))
	if err != nil {
		return fmt.Errorf("failed to create block file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(b); err != nil {
		return fmt.Errorf("failed to encode block: %w", err)
	}
	return nil
}

// GetBlock retrieves a block from a file.
func (s *Storage) GetBlock(hash []byte) (*block.Block, error) {
	if hash == nil || len(hash) == 0 {
		return nil, fmt.Errorf("invalid hash: cannot be nil or empty")
	}

	file, err := os.Open(filepath.Join(s.dataDir, fmt.Sprintf("%x", hash)))
	if err != nil {
		return nil, fmt.Errorf("failed to open block file: %w", err)
	}
	defer file.Close()

	var b block.Block
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&b); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}
	return &b, nil
}

// ChainState represents the state of the blockchain.
type ChainState struct {
	BestBlockHash []byte `json:"best_block_hash"`
	Height        uint64 `json:"height"`
}

// StoreChainState stores the chain state to a file.
func (s *Storage) StoreChainState(state *ChainState) error {
	if state == nil {
		return fmt.Errorf("cannot store nil chain state")
	}

	file, err := os.Create(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		return fmt.Errorf("failed to create chain state file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(state); err != nil {
		return fmt.Errorf("failed to encode chain state: %w", err)
	}
	return nil
}

// GetChainState retrieves the chain state from a file.
func (s *Storage) GetChainState() (*ChainState, error) {
	file, err := os.Open(filepath.Join(s.dataDir, "chainstate"))
	if err != nil {
		if os.IsNotExist(err) {
			return &ChainState{}, nil
		}
		return nil, fmt.Errorf("failed to open chain state file: %w", err)
	}
	defer file.Close()

	var state ChainState
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode chain state: %w", err)
	}
	return &state, nil
}

// Write writes a key-value pair to storage.
func (s *Storage) Write(key []byte, value []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}
	if value == nil {
		return fmt.Errorf("invalid value: cannot be nil")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))

	// SECURITY FIX: Encrypt data if encryption is enabled
	dataToWrite := value
	if s.enableEncryption && len(s.encryptionKey) > 0 {
		encryptedData, err := s.encryptData(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		dataToWrite = encryptedData
	}

	// SECURITY FIX: Use secure file permissions (owner read/write only)
	if err := os.WriteFile(filename, dataToWrite, 0600); err != nil {
		return fmt.Errorf("failed to write key-value pair: %w", err)
	}
	return nil
}

// Read reads a value from storage given a key.
func (s *Storage) Read(key []byte) ([]byte, error) {
	if key == nil || len(key) == 0 {
		return nil, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err // Return the original os.IsNotExist error
		}
		return nil, fmt.Errorf("failed to read key-value pair: %w", err)
	}

	// SECURITY FIX: Decrypt data if encryption is enabled
	if s.enableEncryption && len(s.encryptionKey) > 0 {
		decryptedData, err := s.decryptData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data: %w", err)
		}
		return decryptedData, nil
	}

	return data, nil
}

// SECURITY FIX: encryptData encrypts data using AES-256-GCM
func (s *Storage) encryptData(data []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// SECURITY FIX: decryptData decrypts data using AES-256-GCM
func (s *Storage) decryptData(encryptedData []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Delete deletes a key-value pair from storage.
func (s *Storage) Delete(key []byte) error {
	if key == nil || len(key) == 0 {
		return fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete key-value pair: %w", err)
	}
	return nil
}

// Has checks if a key exists in storage.
func (s *Storage) Has(key []byte) (bool, error) {
	if key == nil || len(key) == 0 {
		return false, fmt.Errorf("invalid key: cannot be nil or empty")
	}

	filename := filepath.Join(s.dataDir, hex.EncodeToString(key))
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if key exists: %w", err)
	}
	return true, nil
}

// Close is a no-op for file-based storage.
func (s *Storage) Close() error {
	return nil
}
