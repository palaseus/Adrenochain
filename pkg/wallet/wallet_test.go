package wallet

import (
	"testing"
)

func TestNewWallet(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)

	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	if wallet == nil {
		t.Fatal("Wallet is nil")
	}

	if wallet.defaultKey == nil {
		t.Fatal("Default key is nil")
	}

	if wallet.keyType != config.KeyType {
		t.Errorf("Expected key type %d, got %d", config.KeyType, wallet.keyType)
	}

	if len(wallet.accounts) == 0 {
		t.Fatal("No accounts were created")
	}
}

func TestDefaultWalletConfig(t *testing.T) {
	config := DefaultWalletConfig()

	if config.KeyType != KeyTypeECDSA {
		t.Errorf("Expected default key type ECDSA, got %d", config.KeyType)
	}

	if config.Passphrase != "" {
		t.Error("Expected empty passphrase by default")
	}
}

func TestCreateDefaultAccount(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetDefaultAccount()
	if account == nil {
		t.Fatal("Default account is nil")
	}

	if account.Address == "" {
		t.Error("Account address is empty")
	}

	if len(account.PublicKey) == 0 {
		t.Error("Account public key is empty")
	}

	if len(account.PrivateKey) == 0 {
		t.Error("Account private key is empty")
	}

	if account.Balance != 0 {
		t.Errorf("Expected initial balance 0, got %d", account.Balance)
	}

	if account.Nonce != 0 {
		t.Errorf("Expected initial nonce 0, got %d", account.Nonce)
	}
}

func TestCreateAccount(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	initialAccountCount := len(wallet.GetAllAccounts())

	account, err := wallet.CreateAccount()
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	if account == nil {
		t.Fatal("Created account is nil")
	}

	if account.Address == "" {
		t.Error("Account address is empty")
	}

	if len(account.PublicKey) == 0 {
		t.Error("Account public key is empty")
	}

	if len(account.PrivateKey) == 0 {
		t.Error("Account private key is empty")
	}

	// Check that account was added to wallet
	newAccountCount := len(wallet.GetAllAccounts())
	if newAccountCount != initialAccountCount+1 {
		t.Errorf("Expected %d accounts, got %d", initialAccountCount+1, newAccountCount)
	}

	// Verify account can be retrieved
	retrievedAccount := wallet.GetAccount(account.Address)
	if retrievedAccount == nil {
		t.Fatal("Failed to retrieve created account")
	}

	if retrievedAccount != account {
		t.Error("Retrieved account is not the same as created account")
	}
}

func TestGetAllAccounts(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	accounts := wallet.GetAllAccounts()
	if len(accounts) == 0 {
		t.Fatal("No accounts found")
	}

	// Should have at least the default account
	if len(accounts) < 1 {
		t.Errorf("Expected at least 1 account, got %d", len(accounts))
	}

	// Check that all accounts have valid addresses
	for i, account := range accounts {
		if account.Address == "" {
			t.Errorf("Account %d has empty address", i)
		}
	}
}

func TestCreateTransaction(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	fromAccount := wallet.GetDefaultAccount()
	toAddress := "recipient_address_123"
	amount := uint64(1000)
	fee := uint64(10)

	tx, err := wallet.CreateTransaction(fromAccount.Address, toAddress, amount, fee)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	if tx == nil {
		t.Fatal("Created transaction is nil")
	}

	if tx.Version != 1 {
		t.Errorf("Expected transaction version 1, got %d", tx.Version)
	}

	if tx.Fee != fee {
		t.Errorf("Expected fee %d, got %d", fee, tx.Fee)
	}

	if len(tx.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(tx.Outputs))
	}

	if tx.Outputs[0].Value != amount {
		t.Errorf("Expected output value %d, got %d", amount, tx.Outputs[0].Value)
	}

	// Check that transaction hash is calculated
	if len(tx.Hash) == 0 {
		t.Error("Transaction hash is empty")
	}
}

func TestSignTransaction(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	fromAccount := wallet.GetDefaultAccount()
	toAddress := "recipient_address_123"
	amount := uint64(1000)
	fee := uint64(10)

	tx, err := wallet.CreateTransaction(fromAccount.Address, toAddress, amount, fee)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Sign the transaction
	err = wallet.SignTransaction(tx, fromAccount.Address)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify the signature
	valid, err := wallet.VerifyTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to verify transaction: %v", err)
	}

	if !valid {
		t.Error("Transaction signature verification failed")
	}
}

func TestUpdateBalance(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetDefaultAccount()
	initialBalance := wallet.GetBalance(account.Address)

	// Update balance
	newBalance := uint64(5000)
	wallet.UpdateBalance(account.Address, newBalance)

	// Check that balance was updated
	updatedBalance := wallet.GetBalance(account.Address)
	if updatedBalance != newBalance {
		t.Errorf("Expected balance %d, got %d", newBalance, updatedBalance)
	}

	// Check that initial balance was different
	if initialBalance == updatedBalance {
		t.Error("Balance should have changed")
	}
}

func TestImportPrivateKey(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Export a private key from existing account
	account := wallet.GetDefaultAccount()
	privateKeyHex, err := wallet.ExportPrivateKey(account.Address)
	if err != nil {
		t.Fatalf("Failed to export private key: %v", err)
	}

	// Import the private key
	importedAccount, err := wallet.ImportPrivateKey(privateKeyHex)
	if err != nil {
		t.Fatalf("Failed to import private key: %v", err)
	}

	if importedAccount == nil {
		t.Fatal("Imported account is nil")
	}

	// Check that addresses match
	if importedAccount.Address != account.Address {
		t.Errorf("Expected address %s, got %s", account.Address, importedAccount.Address)
	}

	// Check that public keys match
	if string(importedAccount.PublicKey) != string(account.PublicKey) {
		t.Error("Public keys don't match")
	}
}

func TestExportPrivateKey(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetDefaultAccount()
	privateKeyHex, err := wallet.ExportPrivateKey(account.Address)
	if err != nil {
		t.Fatalf("Failed to export private key: %v", err)
	}

	if privateKeyHex == "" {
		t.Error("Exported private key is empty")
	}

	// Check that it's a valid hex string
	if len(privateKeyHex) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("Expected private key hex length 64, got %d", len(privateKeyHex))
	}
}

func TestWalletString(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	walletStr := wallet.String()
	if walletStr == "" {
		t.Error("Wallet string representation is empty")
	}
}

func TestAccountString(t *testing.T) {
	config := DefaultWalletConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetDefaultAccount()
	accountStr := account.String()
	if accountStr == "" {
		t.Error("Account string representation is empty")
	}
}
