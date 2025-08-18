package defi

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/security"
)

func TestNewPrivateDeFi(t *testing.T) {
	// Test with default config
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	if pdf == nil {
		t.Fatal("Expected non-nil PrivateDeFi")
	}

	if pdf.Config.MaxTransactions != 10000 {
		t.Errorf("Expected MaxTransactions to be 10000, got %d", pdf.Config.MaxTransactions)
	}

	if pdf.Config.MaxBalances != 100000 {
		t.Errorf("Expected MaxBalances to be 100000, got %d", pdf.Config.MaxBalances)
	}

	if pdf.Config.MaxOperations != 5000 {
		t.Errorf("Expected MaxOperations to be 5000, got %d", pdf.Config.MaxOperations)
	}

	if pdf.Config.EncryptionKeySize != 32 {
		t.Errorf("Expected EncryptionKeySize to be 32, got %d", pdf.Config.EncryptionKeySize)
	}

	if pdf.Config.ZKProofType != security.ProofTypeBulletproofs {
		t.Errorf("Expected ZKProofType to be Bulletproofs, got %v", pdf.Config.ZKProofType)
	}

	if pdf.Config.TransactionTimeout != time.Minute*10 {
		t.Errorf("Expected TransactionTimeout to be 10 minutes, got %v", pdf.Config.TransactionTimeout)
	}

	if pdf.Config.CleanupInterval != time.Hour {
		t.Errorf("Expected CleanupInterval to be 1 hour, got %v", pdf.Config.CleanupInterval)
	}

	if len(pdf.encryptionKey) != 32 {
		t.Errorf("Expected encryption key length to be 32, got %d", len(pdf.encryptionKey))
	}

	if pdf.running {
		t.Error("Expected PrivateDeFi to not be running initially")
	}

	if len(pdf.Transactions) != 0 {
		t.Errorf("Expected 0 transactions initially, got %d", len(pdf.Transactions))
	}

	if len(pdf.Balances) != 0 {
		t.Errorf("Expected 0 balances initially, got %d", len(pdf.Balances))
	}

	if len(pdf.Operations) != 0 {
		t.Errorf("Expected 0 operations initially, got %d", len(pdf.Operations))
	}
}

func TestNewPrivateDeFiCustomConfig(t *testing.T) {
	// Test with custom config
	customConfig := PrivateDeFiConfig{
		MaxTransactions:    5000,
		MaxBalances:        50000,
		MaxOperations:      2500,
		EncryptionKeySize:  64,
		ZKProofType:        security.ProofTypeZkSNARK,
		TransactionTimeout: time.Minute * 5,
		CleanupInterval:    time.Hour * 2,
	}

	pdf := NewPrivateDeFi(customConfig)

	if pdf.Config.MaxTransactions != 5000 {
		t.Errorf("Expected MaxTransactions to be 5000, got %d", pdf.Config.MaxTransactions)
	}

	if pdf.Config.MaxBalances != 50000 {
		t.Errorf("Expected MaxBalances to be 50000, got %d", pdf.Config.MaxBalances)
	}

	if pdf.Config.MaxOperations != 2500 {
		t.Errorf("Expected MaxOperations to be 2500, got %d", pdf.Config.MaxOperations)
	}

	if pdf.Config.EncryptionKeySize != 64 {
		t.Errorf("Expected EncryptionKeySize to be 64, got %d", pdf.Config.EncryptionKeySize)
	}

	if pdf.Config.ZKProofType != security.ProofTypeZkSNARK {
		t.Errorf("Expected ZKProofType to be ZkSNARK, got %v", pdf.Config.ZKProofType)
	}

	if pdf.Config.TransactionTimeout != time.Minute*5 {
		t.Errorf("Expected TransactionTimeout to be 5 minutes, got %v", pdf.Config.TransactionTimeout)
	}

	if pdf.Config.CleanupInterval != time.Hour*2 {
		t.Errorf("Expected CleanupInterval to be 2 hours, got %v", pdf.Config.CleanupInterval)
	}

	if len(pdf.encryptionKey) != 64 {
		t.Errorf("Expected encryption key length to be 64, got %d", len(pdf.encryptionKey))
	}
}

func TestStartStop(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test Start
	err := pdf.Start()
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}

	if !pdf.running {
		t.Error("Expected PrivateDeFi to be running after Start")
	}

	// Test Start when already running
	err = pdf.Start()
	if err == nil {
		t.Error("Expected error when starting already running PrivateDeFi")
	}

	// Test Stop
	err = pdf.Stop()
	if err != nil {
		t.Fatalf("Expected Stop to succeed, got error: %v", err)
	}

	if pdf.running {
		t.Error("Expected PrivateDeFi to not be running after Stop")
	}

	// Test Stop when not running
	err = pdf.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running PrivateDeFi")
	}
}

func TestCreateConfidentialTransaction(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test creating a valid transaction
	amount := big.NewInt(1000)
	data := map[string]interface{}{
		"fee":  0.001,
		"note": "Test transaction",
	}

	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		data,
	)

	if err != nil {
		t.Fatalf("Expected CreateConfidentialTransaction to succeed, got error: %v", err)
	}

	if tx == nil {
		t.Fatal("Expected non-nil transaction")
	}

	if tx.ID == "" {
		t.Error("Expected transaction ID to be generated")
	}

	if tx.Type != TransactionTypeSwap {
		t.Errorf("Expected Type to be Swap, got %v", tx.Type)
	}

	if tx.Asset != "BTC" {
		t.Errorf("Expected Asset to be 'BTC', got %s", tx.Asset)
	}

	if tx.Amount.Cmp(amount) != 0 {
		t.Errorf("Expected Amount to be %s, got %s", amount.String(), tx.Amount.String())
	}

	if tx.Sender != "alice" {
		t.Errorf("Expected Sender to be 'alice', got %s", tx.Sender)
	}

	if tx.Recipient != "bob" {
		t.Errorf("Expected Recipient to be 'bob', got %s", tx.Recipient)
	}

	if tx.Status != TransactionStatusPending {
		t.Errorf("Expected Status to be Pending, got %v", tx.Status)
	}

	if tx.ZKProof == nil {
		t.Error("Expected ZK proof to be generated")
	}

	if len(tx.EncryptedAmount) == 0 {
		t.Error("Expected amount to be encrypted")
	}

	if len(tx.EncryptedData) == 0 {
		t.Error("Expected data to be encrypted")
	}

	// Verify transaction was stored
	if len(pdf.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(pdf.Transactions))
	}

	storedTx, exists := pdf.Transactions[tx.ID]
	if !exists {
		t.Error("Expected transaction to be stored")
	}

	if storedTx.ID != tx.ID {
		t.Errorf("Expected stored transaction ID to match, got %s vs %s", storedTx.ID, tx.ID)
	}
}

func TestCreateConfidentialTransactionLimitReached(t *testing.T) {
	// Create PrivateDeFi with very low limits
	config := PrivateDeFiConfig{
		MaxTransactions: 1,
	}
	pdf := NewPrivateDeFi(config)

	// Create first transaction
	amount := big.NewInt(1000)
	_, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Expected first transaction to succeed, got error: %v", err)
	}

	// Try to create second transaction (should fail)
	_, err = pdf.CreateConfidentialTransaction(
		TransactionTypeLiquidity,
		"ETH",
		amount,
		"bob",
		"charlie",
		nil,
	)

	if err == nil {
		t.Error("Expected error when transaction limit reached")
	}

	if err.Error() != "transaction limit reached" {
		t.Errorf("Expected 'transaction limit reached' error, got: %v", err)
	}
}

func TestProcessTransaction(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create a transaction
	amount := big.NewInt(1000)
	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Create balances for sender and recipient
	_, err = pdf.CreatePrivateBalance("BTC", "alice", big.NewInt(2000))
	if err != nil {
		t.Fatalf("Failed to create sender balance: %v", err)
	}

	_, err = pdf.CreatePrivateBalance("BTC", "bob", big.NewInt(500))
	if err != nil {
		t.Fatalf("Failed to create recipient balance: %v", err)
	}

	// Process the transaction
	err = pdf.ProcessTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Expected ProcessTransaction to succeed, got error: %v", err)
	}

	// Verify transaction status was updated
	processedTx, err := pdf.GetTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Failed to get processed transaction: %v", err)
	}

	if processedTx.Status != TransactionStatusCompleted {
		t.Errorf("Expected Status to be Completed, got %v", processedTx.Status)
	}
}

func TestProcessTransactionNotFound(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Try to process non-existent transaction
	err := pdf.ProcessTransaction("non_existent_id")

	if err == nil {
		t.Error("Expected error when processing non-existent transaction")
	}

	if err.Error() != "transaction not found: non_existent_id" {
		t.Errorf("Expected 'transaction not found' error, got: %v", err)
	}
}

func TestProcessTransactionWrongStatus(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create a transaction
	amount := big.NewInt(1000)
	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Manually change status to completed
	tx.Status = TransactionStatusCompleted

	// Try to process completed transaction
	err = pdf.ProcessTransaction(tx.ID)

	if err == nil {
		t.Error("Expected error when processing non-pending transaction")
	}

	if err.Error() != "transaction is not in pending status: completed" {
		t.Errorf("Expected status error, got: %v", err)
	}
}

func TestCreatePrivateBalance(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test creating a valid balance
	amount := big.NewInt(5000)
	balance, err := pdf.CreatePrivateBalance("ETH", "alice", amount)

	if err != nil {
		t.Fatalf("Expected CreatePrivateBalance to succeed, got error: %v", err)
	}

	if balance == nil {
		t.Fatal("Expected non-nil balance")
	}

	if balance.Asset != "ETH" {
		t.Errorf("Expected Asset to be 'ETH', got %s", balance.Asset)
	}

	if balance.Owner != "alice" {
		t.Errorf("Expected Owner to be 'alice', got %s", balance.Owner)
	}

	if len(balance.EncryptedAmount) == 0 {
		t.Error("Expected amount to be encrypted")
	}

	if len(balance.Commitment) == 0 {
		t.Error("Expected commitment to be generated")
	}

	// Verify balance was stored
	balanceKey := "ETH:alice"
	if len(pdf.Balances) != 1 {
		t.Errorf("Expected 1 balance, got %d", len(pdf.Balances))
	}

	storedBalance, exists := pdf.Balances[balanceKey]
	if !exists {
		t.Error("Expected balance to be stored")
	}

	if storedBalance.Asset != balance.Asset {
		t.Errorf("Expected stored balance asset to match, got %s vs %s", storedBalance.Asset, balance.Asset)
	}
}

func TestCreatePrivateBalanceAlreadyExists(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create first balance
	amount := big.NewInt(5000)
	_, err := pdf.CreatePrivateBalance("BTC", "alice", amount)
	if err != nil {
		t.Fatalf("Failed to create first balance: %v", err)
	}

	// Try to create duplicate balance
	_, err = pdf.CreatePrivateBalance("BTC", "alice", big.NewInt(1000))

	if err == nil {
		t.Error("Expected error when creating duplicate balance")
	}

	if err.Error() != "balance already exists for asset BTC and owner alice" {
		t.Errorf("Expected duplicate balance error, got: %v", err)
	}
}

func TestCreatePrivateBalanceLimitReached(t *testing.T) {
	// Create PrivateDeFi with very low limits
	config := PrivateDeFiConfig{
		MaxBalances: 1,
	}
	pdf := NewPrivateDeFi(config)

	// Create first balance
	amount := big.NewInt(5000)
	_, err := pdf.CreatePrivateBalance("BTC", "alice", amount)
	if err != nil {
		t.Fatalf("Failed to create first balance: %v", err)
	}

	// Try to create second balance (should fail)
	_, err = pdf.CreatePrivateBalance("ETH", "bob", amount)

	if err == nil {
		t.Error("Expected error when balance limit reached")
	}

	if err.Error() != "balance limit reached" {
		t.Errorf("Expected 'balance limit reached' error, got: %v", err)
	}
}

func TestGetPrivateBalance(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create a balance
	amount := big.NewInt(5000)
	createdBalance, err := pdf.CreatePrivateBalance("BTC", "alice", amount)
	if err != nil {
		t.Fatalf("Failed to create balance: %v", err)
	}

	// Retrieve the balance
	retrievedBalance, err := pdf.GetPrivateBalance("BTC", "alice")
	if err != nil {
		t.Fatalf("Expected GetPrivateBalance to succeed, got error: %v", err)
	}

	if retrievedBalance == nil {
		t.Fatal("Expected non-nil retrieved balance")
	}

	if retrievedBalance.Asset != createdBalance.Asset {
		t.Errorf("Expected Asset to match, got %s vs %s", retrievedBalance.Asset, createdBalance.Asset)
	}

	if retrievedBalance.Owner != createdBalance.Owner {
		t.Errorf("Expected Owner to match, got %s vs %s", retrievedBalance.Owner, createdBalance.Owner)
	}

	// Verify it's a deep copy (modifying retrieved shouldn't affect stored)
	originalAsset := retrievedBalance.Asset
	retrievedBalance.Asset = "Modified"

	storedBalance, _ := pdf.GetPrivateBalance("BTC", "alice")
	if storedBalance.Asset == "Modified" {
		t.Error("Expected stored balance to not be affected by external modifications")
	}

	if storedBalance.Asset != originalAsset {
		t.Errorf("Expected stored balance asset to remain unchanged, got %s", storedBalance.Asset)
	}
}

func TestGetPrivateBalanceNotFound(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Try to get non-existent balance
	_, err := pdf.GetPrivateBalance("BTC", "alice")

	if err == nil {
		t.Error("Expected error when getting non-existent balance")
	}

	if err.Error() != "balance not found for asset BTC and owner alice" {
		t.Errorf("Expected 'balance not found' error, got: %v", err)
	}
}

func TestExecuteDeFiOperation(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test executing a valid operation
	params := map[string]interface{}{
		"amount":   1000,
		"duration": "30d",
		"rate":     0.05,
	}

	op, err := pdf.ExecuteDeFiOperation(
		TransactionTypeLending,
		"ETH",
		"create_lending_pool",
		params,
	)

	if err != nil {
		t.Fatalf("Expected ExecuteDeFiOperation to succeed, got error: %v", err)
	}

	if op == nil {
		t.Fatal("Expected non-nil operation")
	}

	if op.ID == "" {
		t.Error("Expected operation ID to be generated")
	}

	if op.Type != TransactionTypeLending {
		t.Errorf("Expected Type to be Lending, got %v", op.Type)
	}

	if op.Asset != "ETH" {
		t.Errorf("Expected Asset to be 'ETH', got %s", op.Asset)
	}

	if op.Operation != "create_lending_pool" {
		t.Errorf("Expected Operation to be 'create_lending_pool', got %s", op.Operation)
	}

	if op.Status != OperationStatusInitiated {
		t.Errorf("Expected Status to be Initiated, got %v", op.Status)
	}

	if op.ZKProof == nil {
		t.Error("Expected ZK proof to be generated")
	}

	if len(op.EncryptedParams) == 0 {
		t.Error("Expected parameters to be encrypted")
	}

	// Verify operation was stored
	if len(pdf.Operations) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(pdf.Operations))
	}

	storedOp, exists := pdf.Operations[op.ID]
	if !exists {
		t.Error("Expected operation to be stored")
	}

	if storedOp.ID != op.ID {
		t.Errorf("Expected stored operation ID to match, got %s vs %s", storedOp.ID, op.ID)
	}
}

func TestExecuteDeFiOperationLimitReached(t *testing.T) {
	// Create PrivateDeFi with very low limits
	pdf := NewPrivateDeFi(PrivateDeFiConfig{MaxOperations: 1})

	// Create first operation
	params := map[string]interface{}{"test": "value"}
	_, err := pdf.ExecuteDeFiOperation(
		TransactionTypeSwap,
		"BTC",
		"test_op",
		params,
	)

	if err != nil {
		t.Fatalf("Failed to create first operation: %v", err)
	}

	// Try to create second operation (should fail)
	_, err = pdf.ExecuteDeFiOperation(
		TransactionTypeLiquidity,
		"ETH",
		"test_op2",
		params,
	)

	if err == nil {
		t.Error("Expected error when operation limit reached")
	}

	if err.Error() != "operation limit reached" {
		t.Errorf("Expected 'operation limit reached' error, got: %v", err)
	}
}

func TestGetTransaction(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create a transaction
	amount := big.NewInt(1000)
	createdTx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Retrieve the transaction
	retrievedTx, err := pdf.GetTransaction(createdTx.ID)
	if err != nil {
		t.Fatalf("Expected GetTransaction to succeed, got error: %v", err)
	}

	if retrievedTx == nil {
		t.Fatal("Expected non-nil retrieved transaction")
	}

	if retrievedTx.ID != createdTx.ID {
		t.Errorf("Expected ID to match, got %s vs %s", retrievedTx.ID, createdTx.ID)
	}

	// Verify it's a deep copy
	originalType := retrievedTx.Type
	retrievedTx.Type = TransactionTypeCustom

	storedTx, _ := pdf.GetTransaction(createdTx.ID)
	if storedTx.Type == TransactionTypeCustom {
		t.Error("Expected stored transaction to not be affected by external modifications")
	}

	if storedTx.Type != originalType {
		t.Errorf("Expected stored transaction type to remain unchanged, got %v", storedTx.Type)
	}
}

func TestGetTransactionNotFound(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Try to get non-existent transaction
	_, err := pdf.GetTransaction("non_existent_id")

	if err == nil {
		t.Error("Expected error when getting non-existent transaction")
	}

	if err.Error() != "transaction not found: non_existent_id" {
		t.Errorf("Expected 'transaction not found' error, got: %v", err)
	}
}

func TestGetTransactions(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create multiple transactions
	amount := big.NewInt(1000)

	_, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create first transaction: %v", err)
	}

	_, err = pdf.CreateConfidentialTransaction(
		TransactionTypeLiquidity,
		"ETH",
		amount,
		"bob",
		"charlie",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create second transaction: %v", err)
	}

	// Get all transactions
	allTransactions := pdf.GetTransactions(nil)
	if len(allTransactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(allTransactions))
	}

	// Get transactions with filter
	btcTransactions := pdf.GetTransactions(func(tx *ConfidentialTransaction) bool {
		return tx.Asset == "BTC"
	})

	if len(btcTransactions) != 1 {
		t.Errorf("Expected 1 BTC transaction, got %d", len(btcTransactions))
	}

	if btcTransactions[0].Asset != "BTC" {
		t.Errorf("Expected filtered transaction to be BTC, got %s", btcTransactions[0].Asset)
	}
}

func TestGetOperations(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create multiple operations
	params := map[string]interface{}{"test": "value"}

	_, err := pdf.ExecuteDeFiOperation(
		TransactionTypeSwap,
		"BTC",
		"op1",
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create first operation: %v", err)
	}

	_, err = pdf.ExecuteDeFiOperation(
		TransactionTypeLiquidity,
		"ETH",
		"op2",
		params,
	)
	if err != nil {
		t.Fatalf("Failed to create second operation: %v", err)
	}

	// Get all operations
	allOperations := pdf.GetOperations(nil)
	if len(allOperations) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(allOperations))
	}

	// Get operations with filter
	btcOperations := pdf.GetOperations(func(op *DeFiOperation) bool {
		return op.Asset == "BTC"
	})

	if len(btcOperations) != 1 {
		t.Errorf("Expected 1 BTC operation, got %d", len(btcOperations))
	}

	if btcOperations[0].Asset != "BTC" {
		t.Errorf("Expected filtered operation to be BTC, got %s", btcOperations[0].Asset)
	}
}

func TestTransactionTypeString(t *testing.T) {
	testCases := []struct {
		txType   TransactionType
		expected string
	}{
		{TransactionTypeSwap, "swap"},
		{TransactionTypeLiquidity, "liquidity"},
		{TransactionTypeLending, "lending"},
		{TransactionTypeBorrowing, "borrowing"},
		{TransactionTypeYield, "yield"},
		{TransactionTypeStaking, "staking"},
		{TransactionTypeGovernance, "governance"},
		{TransactionTypeCustom, "custom"},
		{TransactionType(999), "unknown"},
	}

	for _, tc := range testCases {
		result := tc.txType.String()
		if result != tc.expected {
			t.Errorf("Expected %s for TransactionType(%d), got %s", tc.expected, tc.txType, result)
		}
	}
}

func TestTransactionStatusString(t *testing.T) {
	testCases := []struct {
		status   TransactionStatus
		expected string
	}{
		{TransactionStatusPending, "pending"},
		{TransactionStatusProcessing, "processing"},
		{TransactionStatusCompleted, "completed"},
		{TransactionStatusFailed, "failed"},
		{TransactionStatusCancelled, "cancelled"},
		{TransactionStatus(999), "unknown"},
	}

	for _, tc := range testCases {
		result := tc.status.String()
		if result != tc.expected {
			t.Errorf("Expected %s for TransactionStatus(%d), got %s", tc.expected, tc.status, result)
		}
	}
}

func TestOperationStatusString(t *testing.T) {
	testCases := []struct {
		status   OperationStatus
		expected string
	}{
		{OperationStatusInitiated, "initiated"},
		{OperationStatusExecuting, "executing"},
		{OperationStatusCompleted, "completed"},
		{OperationStatusFailed, "failed"},
		{OperationStatusRolledBack, "rolled_back"},
		{OperationStatus(999), "unknown"},
	}

	for _, tc := range testCases {
		result := tc.status.String()
		if result != tc.expected {
			t.Errorf("Expected %s for OperationStatus(%d), got %s", tc.expected, tc.status, result)
		}
	}
}

func TestConcurrency(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Start the system
	err := pdf.Start()
	if err != nil {
		t.Fatalf("Failed to start PrivateDeFi: %v", err)
	}
	defer pdf.Stop()

	// Test concurrent transaction creation
	const numGoroutines = 10
	const transactionsPerGoroutine = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < transactionsPerGoroutine; j++ {
				amount := big.NewInt(int64(j + 1))
				_, err := pdf.CreateConfidentialTransaction(
					TransactionTypeSwap,
					"BTC",
					amount,
					fmt.Sprintf("user_%d", id),
					fmt.Sprintf("recipient_%d", j),
					nil,
				)
				if err != nil {
					// Log error but continue (some might fail due to limits)
					continue
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify some transactions were created
	if len(pdf.Transactions) == 0 {
		t.Error("Expected some transactions to be created")
	}
}

func TestMemorySafety(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})
	
	// Test that modifications to returned data don't affect internal state
	amount := big.NewInt(1000)
	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}
	
	// Store the original ID
	originalID := tx.ID
	
	// Get the transaction again to test the copy mechanism
	retrievedTx, err := pdf.GetTransaction(originalID)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}
	
	// Modify the retrieved transaction
	retrievedTx.ID = "Modified ID"
	
	// Verify internal state wasn't affected
	storedTx, err := pdf.GetTransaction(originalID)
	if err != nil {
		t.Fatalf("Failed to get stored transaction: %v", err)
	}
	
	if storedTx.ID == "Modified ID" {
		t.Error("Expected internal state to not be affected by external modifications")
	}
	
	if storedTx.ID != originalID {
		t.Errorf("Expected stored transaction ID to remain unchanged, got %s", storedTx.ID)
	}
}

func TestEdgeCases(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test with zero amount
	zeroAmount := big.NewInt(0)
	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		zeroAmount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Expected zero amount transaction to succeed, got error: %v", err)
	}

	if tx.Amount.Cmp(zeroAmount) != 0 {
		t.Errorf("Expected amount to be zero, got %s", tx.Amount.String())
	}

	// Test with very large amount
	largeAmount := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	tx2, err := pdf.CreateConfidentialTransaction(
		TransactionTypeLiquidity,
		"ETH",
		largeAmount,
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Expected large amount transaction to succeed, got error: %v", err)
	}

	if tx2.Amount.Cmp(largeAmount) != 0 {
		t.Errorf("Expected amount to match large amount, got %s", tx2.Amount.String())
	}

	// Test with empty asset
	_, err = pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"",
		big.NewInt(1000),
		"alice",
		"bob",
		nil,
	)

	if err != nil {
		t.Fatalf("Expected empty asset transaction to succeed, got error: %v", err)
	}

	// Test with empty sender/recipient
	_, err = pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		big.NewInt(1000),
		"",
		"",
		nil,
	)

	if err != nil {
		t.Fatalf("Expected empty sender/recipient transaction to succeed, got error: %v", err)
	}
}

func TestEncryptionDecryption(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Test data encryption and decryption
	testData := []byte("Hello, Private DeFi!")

	encrypted, err := pdf.encryptData(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("Expected encrypted data to be non-empty")
	}

	if string(encrypted) == string(testData) {
		t.Error("Expected encrypted data to be different from plaintext")
	}

	// Test decryption
	decrypted, err := pdf.decryptData(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Errorf("Expected decrypted data to match original, got %s vs %s", string(decrypted), string(testData))
	}
}

func TestBalanceUpdateCommitments(t *testing.T) {
	pdf := NewPrivateDeFi(PrivateDeFiConfig{})

	// Create a balance
	amount := big.NewInt(5000)
	balance, err := pdf.CreatePrivateBalance("BTC", "alice", amount)
	if err != nil {
		t.Fatalf("Failed to create balance: %v", err)
	}

	// Store original commitment
	originalCommitment := make([]byte, len(balance.Commitment))
	copy(originalCommitment, balance.Commitment)

	// Manually update commitments
	pdf.updateBalanceCommitments()

	// Verify commitment was updated
	updatedBalance, err := pdf.GetPrivateBalance("BTC", "alice")
	if err != nil {
		t.Fatalf("Failed to get updated balance: %v", err)
	}

	// Commitments should be the same since the encrypted amount didn't change
	if !bytes.Equal(updatedBalance.Commitment, originalCommitment) {
		t.Error("Expected commitment to remain the same")
	}
}

func TestCleanupOldData(t *testing.T) {
	// Create PrivateDeFi with short timeout
	config := PrivateDeFiConfig{
		TransactionTimeout: time.Millisecond * 100,
		CleanupInterval:    time.Millisecond * 50,
	}
	pdf := NewPrivateDeFi(config)

	// Start the system
	err := pdf.Start()
	if err != nil {
		t.Fatalf("Failed to start PrivateDeFi: %v", err)
	}
	defer pdf.Stop()

	// Create a transaction
	amount := big.NewInt(1000)
	tx, err := pdf.CreateConfidentialTransaction(
		TransactionTypeSwap,
		"BTC",
		amount,
		"alice",
		"bob",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Process the transaction to complete it
	err = pdf.ProcessTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Failed to process transaction: %v", err)
	}

	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)

	// Verify transaction was cleaned up
	_, err = pdf.GetTransaction(tx.ID)
	if err == nil {
		t.Error("Expected completed transaction to be cleaned up")
	}
}
