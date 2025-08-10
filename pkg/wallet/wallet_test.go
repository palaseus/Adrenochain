package wallet

import (
	"os"
	"testing"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"

	"github.com/gochain/gochain/pkg/storage" // Added import
	"github.com/gochain/gochain/pkg/utxo"
	"github.com/stretchr/testify/assert" // Added import for assert
)

// Helper function to create a temporary storage for tests
func newTestStorage(t *testing.T) *storage.Storage {
	tempDir, err := os.MkdirTemp("", "wallet_test_storage")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) }) // Clean up after test

	storageConfig := storage.DefaultStorageConfig().WithDataDir(tempDir)
	s, err := storage.NewStorage(storageConfig)
	assert.NoError(t, err)
	return s
}

func TestNewWallet(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotNil(t, wallet.defaultKey)
	assert.Equal(t, config.KeyType, wallet.keyType)
	assert.NotEmpty(t, wallet.accounts)
}

func TestDefaultWalletConfig(t *testing.T) {
	config := DefaultWalletConfig()

	assert.Equal(t, KeyTypeECDSA, config.KeyType)
	assert.Empty(t, config.Passphrase)
	assert.Equal(t, "wallet.dat", config.WalletFile) // Check default wallet file
}

func TestCreateDefaultAccount(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	account := wallet.GetDefaultAccount()
	assert.NotNil(t, account)
	assert.NotEmpty(t, account.Address)
	assert.NotEmpty(t, account.PublicKey)
	assert.NotEmpty(t, account.PrivateKey)
	assert.Equal(t, uint64(0), account.Balance)
	assert.Equal(t, uint64(0), account.Nonce)
}

func TestCreateAccount(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	initialAccountCount := len(wallet.GetAllAccounts())

	account, err := wallet.CreateAccount()
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEmpty(t, account.Address)
	assert.NotEmpty(t, account.PublicKey)
	assert.NotEmpty(t, account.PrivateKey)

	// Check that account was added to wallet
	newAccountCount := len(wallet.GetAllAccounts())
	assert.Equal(t, initialAccountCount+1, newAccountCount)

	// Verify account can be retrieved
	retrievedAccount := wallet.GetAccount(account.Address)
	assert.Equal(t, account, retrievedAccount)
}

func TestGetAllAccounts(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	accounts := wallet.GetAllAccounts()
	assert.NotEmpty(t, accounts)

	// Should have at least the default account
	assert.True(t, len(accounts) >= 1)

	// Check that all accounts have valid addresses
	for i, account := range accounts {
		assert.NotEmpty(t, account.Address, "Account %d has empty address", i)
	}
}

func TestCreateTransaction(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	fromAccount := wallet.GetDefaultAccount()

	// Create a test UTXO for the fromAccount so it has funds to spend
	testUTXO := &utxo.UTXO{
		TxHash:       []byte("test_tx_hash"),
		TxIndex:      0,
		Value:        5000, // Give it 5000 to spend
		ScriptPubKey: fromAccount.PublicKey,
		Address:      fromAccount.Address,
		IsCoinbase:   false,
		Height:       1,
	}
	us.AddUTXO(testUTXO)

	// Generate a valid recipient address
	toPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	toAddress := hex.EncodeToString(wallet.generateAddress(toPrivKey))
	amount := uint64(1000)
	fee := uint64(10)

	tx, err := wallet.CreateTransaction(fromAccount.Address, toAddress, amount, fee)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	assert.Equal(t, uint32(1), tx.Version)
	assert.Equal(t, fee, tx.Fee)
	assert.Equal(t, 2, len(tx.Outputs)) // 1 for recipient, 1 for change
	assert.Equal(t, amount, tx.Outputs[0].Value)
	assert.NotEmpty(t, tx.Hash)
}

func TestSignTransaction(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	fromAccount := wallet.GetDefaultAccount()

	// Create a test UTXO for the fromAccount so it has funds to spend
	testUTXO := &utxo.UTXO{
		TxHash:       []byte("test_tx_hash_sign"),
		TxIndex:      0,
		Value:        5000, // Give it 5000 to spend
		ScriptPubKey: fromAccount.PublicKey,
		Address:      fromAccount.Address,
		IsCoinbase:   false,
		Height:       1,
	}
	us.AddUTXO(testUTXO)

	// Generate a valid recipient address
	toPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	toAddress := hex.EncodeToString(wallet.generateAddress(toPrivKey))
	amount := uint64(1000)
	fee := uint64(10)

	tx, err := wallet.CreateTransaction(fromAccount.Address, toAddress, amount, fee)
	assert.NoError(t, err)

	// Sign the transaction
	err = wallet.SignTransaction(tx, fromAccount.Address)
	assert.NoError(t, err)

	// Verify the signature
	valid, err := wallet.VerifyTransaction(tx)
	assert.NoError(t, err)
	assert.True(t, valid, "Transaction signature verification failed")
}

func TestUpdateBalance(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	account := wallet.GetDefaultAccount()
	initialBalance := wallet.GetBalance(account.Address)

	// Update balance
	newBalance := uint64(5000)
	wallet.UpdateBalance(account.Address, newBalance)

	// Check that balance was updated
	updatedBalance := wallet.GetBalance(account.Address)
	assert.Equal(t, newBalance, updatedBalance)

	// Check that initial balance was different
	assert.NotEqual(t, initialBalance, updatedBalance)
}

func TestImportPrivateKey(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	// Export a private key from existing account
	account := wallet.GetDefaultAccount()
	privateKeyHex, err := wallet.ExportPrivateKey(account.Address)
	assert.NoError(t, err)

	// Import the private key
	importedAccount, err := wallet.ImportPrivateKey(privateKeyHex)
	assert.NoError(t, err)
	assert.NotNil(t, importedAccount)

	// Check that addresses match
	assert.Equal(t, account.Address, importedAccount.Address)

	// Check that public keys match
	assert.Equal(t, string(account.PublicKey), string(importedAccount.PublicKey))
}

func TestExportPrivateKey(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	account := wallet.GetDefaultAccount()
	privateKeyHex, err := wallet.ExportPrivateKey(account.Address)
	assert.NoError(t, err)

	assert.NotEmpty(t, privateKeyHex)

	// Check that it's a valid hex string (32 bytes = 64 hex chars for P-256 private key)
	assert.Equal(t, 64, len(privateKeyHex))
}

func TestWalletString(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	walletStr := wallet.String()
	assert.NotEmpty(t, walletStr)
}

func TestAccountString(t *testing.T) {
	s := newTestStorage(t)
	config := DefaultWalletConfig()
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s) // Pass storage
	assert.NoError(t, err)

	account := wallet.GetDefaultAccount()
	accountStr := account.String()
	assert.NotEmpty(t, accountStr)
}

func TestWalletPersistence(t *testing.T) {
	s := newTestStorage(t)
	passphrase := "test_passphrase"
	walletFile := "my_test_wallet.dat"

	// 1. Create a new wallet and save it
	config1 := DefaultWalletConfig()
	config1.Passphrase = passphrase
	config1.WalletFile = walletFile
	us1 := utxo.NewUTXOSet()
	wallet1, err := NewWallet(config1, us1, s)
	assert.NoError(t, err)
	assert.NotNil(t, wallet1)
	assert.NotEmpty(t, wallet1.GetAllAccounts())
	initialAccountAddress := wallet1.GetDefaultAccount().Address

	// 2. Load the wallet with the correct passphrase
	config2 := DefaultWalletConfig()
	config2.Passphrase = passphrase
	config2.WalletFile = walletFile
	us2 := utxo.NewUTXOSet()
	wallet2, err := NewWallet(config2, us2, s)
	assert.NoError(t, err)
	assert.NotNil(t, wallet2)
	assert.Equal(t, initialAccountAddress, wallet2.GetDefaultAccount().Address)
	assert.Equal(t, len(wallet1.GetAllAccounts()), len(wallet2.GetAllAccounts()))

	// 3. Attempt to load with an incorrect passphrase
	config3 := DefaultWalletConfig()
	config3.Passphrase = "wrong_passphrase"
	config3.WalletFile = walletFile
	us3 := utxo.NewUTXOSet()
	_, err = NewWallet(config3, us3, s)
	assert.Error(t, err) // Expect an error due to decryption failure
}

func TestWalletEncryptionDecryption(t *testing.T) {
	s := newTestStorage(t) // Need a storage instance for NewWallet, though not directly used here
	passphrase := "super_secret_key"
	config := DefaultWalletConfig()
	config.Passphrase = passphrase
	us := utxo.NewUTXOSet()
	wallet, err := NewWallet(config, us, s)
	assert.NoError(t, err)

	originalData := []byte("This is some sensitive data to be encrypted.")

	encryptedData, err := wallet.Encrypt(originalData)
	assert.NoError(t, err)
	assert.NotNil(t, encryptedData)
	assert.NotEqual(t, originalData, encryptedData) // Encrypted data should be different

	decryptedData, err := wallet.Decrypt(encryptedData)
	assert.NoError(t, err)
	assert.NotNil(t, decryptedData)
	assert.Equal(t, originalData, decryptedData) // Decrypted data should match original

	// Test with incorrect passphrase
	wallet.passphrase = "incorrect_passphrase"
	_, err = wallet.Decrypt(encryptedData)
	assert.Error(t, err) // Expect an error due to incorrect decryption
}
