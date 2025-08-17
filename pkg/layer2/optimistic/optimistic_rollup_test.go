package optimistic

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
)

// TestNewOptimisticRollup tests the creation of a new optimistic rollup
func TestNewOptimisticRollup(t *testing.T) {
	config := OptimisticRollupConfig{
		MaxBatchSize:      1000,
		ChallengePeriod:   time.Hour * 7,
		MaxProofTime:      time.Second * 30,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelHigh,
		MinStake:          big.NewInt(2000000000000000000), // 2 ETH
	}

	rollup := NewOptimisticRollup(config)

	// Test basic initialization
	if rollup == nil {
		t.Fatal("rollup should not be nil")
	}

	if rollup.ID == "" {
		t.Error("rollup ID should not be empty")
	}

	if rollup.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", rollup.BatchNumber)
	}

	if len(rollup.Transactions) != 0 {
		t.Errorf("expected empty transactions, got %d", len(rollup.Transactions))
	}

	if len(rollup.Batches) != 0 {
		t.Errorf("expected empty batches, got %d", len(rollup.Batches))
	}

	if len(rollup.Challenges) != 0 {
		t.Errorf("expected empty challenges, got %d", len(rollup.Challenges))
	}

	if rollup.config.MaxBatchSize != 1000 {
		t.Errorf("expected max batch size 1000, got %d", rollup.config.MaxBatchSize)
	}

	if rollup.config.ChallengePeriod != time.Hour*7 {
		t.Errorf("expected challenge period 7h, got %v", rollup.config.ChallengePeriod)
	}

	if rollup.config.SecurityLevel != SecurityLevelHigh {
		t.Errorf("expected security level %d, got %d", SecurityLevelHigh, rollup.config.SecurityLevel)
	}

	if rollup.config.MinStake.Cmp(big.NewInt(2000000000000000000)) != 0 {
		t.Errorf("expected min stake 2 ETH, got %s", rollup.config.MinStake.String())
	}
}

// TestNewOptimisticRollupDefaults tests default value assignment
func TestNewOptimisticRollupDefaults(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	if rollup.config.MaxBatchSize != 1000 {
		t.Errorf("expected default max batch size 1000, got %d", rollup.config.MaxBatchSize)
	}

	if rollup.config.ChallengePeriod != time.Hour*7 {
		t.Errorf("expected default challenge period 7h, got %v", rollup.config.ChallengePeriod)
	}

	if rollup.config.MaxProofTime != time.Second*30 {
		t.Errorf("expected default max proof time 30s, got %v", rollup.config.MaxProofTime)
	}

	if rollup.config.MinStake.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Errorf("expected default min stake 1 ETH, got %s", rollup.config.MinStake.String())
	}
}

// TestAddTransaction tests adding transactions to the rollup
func TestAddTransaction(t *testing.T) {
	config := OptimisticRollupConfig{
		MaxBatchSize:      2,
		ChallengePeriod:   time.Hour * 7,
		MaxProofTime:      time.Second * 30,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelMedium,
		MinStake:          big.NewInt(1000000000000000000),
	}

	rollup := NewOptimisticRollup(config)

	// Test valid transaction
	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Errorf("failed to add valid transaction: %v", err)
	}

	if len(rollup.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(rollup.Transactions))
	}

	// Test second transaction
	tx2 := createValidTransaction("tx2")
	err = rollup.AddTransaction(tx2)
	if err != nil {
		t.Errorf("failed to add second transaction: %v", err)
	}

	if len(rollup.Transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(rollup.Transactions))
	}

	// Test batch size limit
	tx3 := createValidTransaction("tx3")
	err = rollup.AddTransaction(tx3)
	if err == nil {
		t.Error("expected error when batch size limit reached")
	}

	if len(rollup.Transactions) != 2 {
		t.Errorf("expected 2 transactions after limit, got %d", len(rollup.Transactions))
	}
}

// TestAddTransactionValidation tests transaction validation
func TestAddTransactionValidation(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test empty transaction ID
	tx := createValidTransaction("")
	err := rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for empty transaction ID")
	}

	// Test negative value
	tx = createValidTransaction("tx1")
	tx.Value = big.NewInt(-100)
	err = rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for negative value")
	}

	// Test zero gas limit
	tx = createValidTransaction("tx1")
	tx.GasLimit = 0
	err = rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for zero gas limit")
	}

	// Test invalid gas price
	tx = createValidTransaction("tx1")
	tx.GasPrice = big.NewInt(0)
	err = rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for invalid gas price")
	}

	// Test missing signature
	tx = createValidTransaction("tx1")
	tx.Signature = nil
	err = rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for missing signature")
	}

	// Test zero timestamp
	tx = createValidTransaction("tx1")
	tx.Timestamp = time.Time{}
	err = rollup.AddTransaction(tx)
	if err == nil {
		t.Error("expected error for zero timestamp")
	}
}

// TestProcessBatch tests batch processing
func TestProcessBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test empty batch
	_, err := rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error for empty batch")
	}

	// Test batch with transactions
	tx1 := createValidTransaction("tx1")
	tx2 := createValidTransaction("tx2")

	if err := rollup.AddTransaction(tx1); err != nil {
		t.Fatalf("failed to add first transaction: %v", err)
	}
	if err := rollup.AddTransaction(tx2); err != nil {
		t.Fatalf("failed to add second transaction: %v", err)
	}

	result, err := rollup.ProcessBatch()
	if err != nil {
		t.Errorf("failed to process batch: %v", err)
	}

	if result == nil {
		t.Fatal("batch result should not be nil")
	}

	if !result.Success {
		t.Error("batch should be successful")
	}

	if result.Transactions != 2 {
		t.Errorf("expected 2 transactions, got %d", result.Transactions)
	}

	// Verify rollup state updated
	if rollup.BatchNumber != 1 {
		t.Errorf("expected batch number 1, got %d", rollup.BatchNumber)
	}

	if len(rollup.Transactions) != 0 {
		t.Errorf("expected empty transactions after processing, got %d", len(rollup.Transactions))
	}

	if len(rollup.Batches) != 1 {
		t.Errorf("expected 1 batch, got %d", len(rollup.Batches))
	}

	// Verify batch details
	batch := rollup.Batches[0]
	if batch.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", batch.BatchNumber)
	}

	if batch.Finalized {
		t.Error("batch should not be finalized immediately")
	}

	if time.Now().After(batch.ChallengePeriod) {
		t.Error("challenge period should be in the future")
	}
}

// TestProcessBatchValidationFailure tests batch validation failure
func TestProcessBatchValidationFailure(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Create a rollup with a custom batch processor that always fails validation
	rollup.BatchProcessor = &FailingBatchProcessor{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail validation
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when batch validation fails")
	}

	if !contains(err.Error(), "batch validation failed") {
		t.Errorf("expected validation error, got: %s", err.Error())
	}
}

// TestProcessBatchProcessingFailure tests batch processing failure
func TestProcessBatchProcessingFailure(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Create a rollup with a custom batch processor that always fails processing
	rollup.BatchProcessor = &ProcessingFailingBatchProcessor{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail processing
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when batch processing fails")
	}

	if !contains(err.Error(), "batch processing failed") {
		t.Errorf("expected processing error, got: %s", err.Error())
	}
}





// TestResolveChallengeInvalidChallenge tests resolving an invalid challenge
func TestResolveChallengeInvalidChallenge(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Try to resolve a non-existent challenge
	err := rollup.ResolveChallenge("invalid_challenge_id")
	if err == nil {
		t.Error("expected error when resolving invalid challenge")
	}

	if !contains(err.Error(), "challenge invalid_challenge_id not found") {
		t.Errorf("expected challenge not found error, got: %s", err.Error())
	}
}

// TestFinalizeBatchInvalidBatch tests finalizing an invalid batch
func TestFinalizeBatchInvalidBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Try to finalize a non-existent batch
	err := rollup.FinalizeBatch(999)
	if err == nil {
		t.Error("expected error when finalizing invalid batch")
	}

	if !contains(err.Error(), "batch 999 not found") {
		t.Errorf("expected batch not found error, got: %s", err.Error())
	}
}

// TestUpdateMetricsEdgeCases tests edge cases in updateMetrics
func TestUpdateMetricsEdgeCases(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test first batch (should set average time directly)
	batchResult := &BatchResult{
		BatchNumber:    1,
		StateRoot:      [32]byte{1, 2, 3, 4},
		GasUsed:        21000,
		Transactions:   1,
		ProcessingTime: time.Millisecond * 100,
		Success:        true,
	}

	rollup.updateMetrics(batchResult, time.Millisecond*100)

	if rollup.metrics.TotalBatches != 1 {
		t.Errorf("expected 1 total batch, got %d", rollup.metrics.TotalBatches)
	}

	if rollup.metrics.AverageBatchTime != time.Millisecond*100 {
		t.Errorf("expected 100ms average time, got %v", rollup.metrics.AverageBatchTime)
	}

	// Test second batch (should calculate average)
	batchResult2 := &BatchResult{
		BatchNumber:    2,
		StateRoot:      [32]byte{5, 6, 7, 8},
		GasUsed:        42000,
		Transactions:   2,
		ProcessingTime: time.Millisecond * 200,
		Success:        true,
	}

	rollup.updateMetrics(batchResult2, time.Millisecond*200)

	if rollup.metrics.TotalBatches != 2 {
		t.Errorf("expected 2 total batches, got %d", rollup.metrics.TotalBatches)
	}

	// Average should be (100 + 200) / 2 = 150ms
	expectedAvg := time.Millisecond * 150
	if rollup.metrics.AverageBatchTime != expectedAvg {
		t.Errorf("expected %v average time, got %v", expectedAvg, rollup.metrics.AverageBatchTime)
	}

	if rollup.metrics.TotalGasUsed != 63000 {
		t.Errorf("expected 63000 total gas, got %d", rollup.metrics.TotalGasUsed)
	}
}

// TestChallengeBatch tests batch challenging
func TestChallengeBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{
		MinStake: big.NewInt(1000000000000000000), // 1 ETH
	})

	// Add transactions and process batch
	tx1 := createValidTransaction("tx1")
	tx2 := createValidTransaction("tx2")

	rollup.AddTransaction(tx1)
	rollup.AddTransaction(tx2)
	rollup.ProcessBatch()

	// Test valid challenge
	challenger := [20]byte{1, 2, 3, 4, 5}
	evidence := []byte("fraud evidence")
	stake := big.NewInt(2000000000000000000) // 2 ETH

	challenge, err := rollup.ChallengeBatch(0, challenger, evidence, stake)
	if err != nil {
		t.Errorf("failed to challenge batch: %v", err)
	}

	if challenge == nil {
		t.Fatal("challenge should not be nil")
	}

	if challenge.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", challenge.BatchNumber)
	}

	if challenge.Challenger != challenger {
		t.Error("challenger should match")
	}

	if challenge.Resolved {
		t.Error("challenge should not be resolved initially")
	}

	// Test challenge with insufficient stake
	lowStake := big.NewInt(500000000000000000) // 0.5 ETH
	_, err = rollup.ChallengeBatch(0, challenger, evidence, lowStake)
	if err == nil {
		t.Error("expected error for insufficient stake")
	}

	// Test challenge for non-existent batch
	_, err = rollup.ChallengeBatch(999, challenger, evidence, stake)
	if err == nil {
		t.Error("expected error for non-existent batch")
	}
}

// TestResolveChallenge tests challenge resolution
func TestResolveChallenge(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{
		MinStake: big.NewInt(1000000000000000000), // 1 ETH
	})

	// Add transactions and process batch
	tx1 := createValidTransaction("tx1")
	rollup.AddTransaction(tx1)
	rollup.ProcessBatch()

	// Create challenge
	challenger := [20]byte{1, 2, 3, 4, 5}
	evidence := []byte("fraud evidence")
	stake := big.NewInt(2000000000000000000) // 2 ETH

	challenge, err := rollup.ChallengeBatch(0, challenger, evidence, stake)
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	// Resolve challenge
	err = rollup.ResolveChallenge(challenge.ID)
	if err != nil {
		t.Errorf("failed to resolve challenge: %v", err)
	}

	// Verify challenge is resolved
	resolvedChallenge, err := rollup.GetChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("failed to get resolved challenge: %v", err)
	}

	if !resolvedChallenge.Resolved {
		t.Error("challenge should be resolved")
	}

	// Test resolving already resolved challenge
	err = rollup.ResolveChallenge(challenge.ID)
	if err == nil {
		t.Error("expected error for already resolved challenge")
	}

	// Test resolving non-existent challenge
	err = rollup.ResolveChallenge("non_existent")
	if err == nil {
		t.Error("expected error for non-existent challenge")
	}
}

// TestFinalizeBatch tests batch finalization
func TestFinalizeBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{
		ChallengePeriod: time.Millisecond * 100, // Short challenge period for testing
	})

	// Add transactions and process batch
	tx1 := createValidTransaction("tx1")
	rollup.AddTransaction(tx1)
	rollup.ProcessBatch()

	// Try to finalize before challenge period expires
	err := rollup.FinalizeBatch(0)
	if err == nil {
		t.Error("expected error for premature finalization")
	}

	// Wait for challenge period to expire
	time.Sleep(time.Millisecond * 150)

	// Finalize batch
	err = rollup.FinalizeBatch(0)
	if err != nil {
		t.Errorf("failed to finalize batch: %v", err)
	}

	// Verify batch is finalized
	batch, err := rollup.GetBatch(0)
	if err != nil {
		t.Fatalf("failed to get finalized batch: %v", err)
	}

	if !batch.Finalized {
		t.Error("batch should be finalized")
	}

	// Test finalizing already finalized batch
	err = rollup.FinalizeBatch(0)
	if err == nil {
		t.Error("expected error for already finalized batch")
	}

	// Test finalizing non-existent batch
	err = rollup.FinalizeBatch(999)
	if err == nil {
		t.Error("expected error for non-existent batch")
	}
}

// TestGetState tests getting rollup state
func TestGetState(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	state := rollup.GetState()

	if state == nil {
		t.Fatal("state should not be nil")
	}

	if state["id"] != rollup.ID {
		t.Errorf("expected ID %s, got %v", rollup.ID, state["id"])
	}

	if state["batchNumber"] != rollup.BatchNumber {
		t.Errorf("expected batch number %d, got %v", rollup.BatchNumber, state["batchNumber"])
	}

	if state["transactions"] != len(rollup.Transactions) {
		t.Errorf("expected transactions %d, got %v", len(rollup.Transactions), state["transactions"])
	}

	if state["batches"] != len(rollup.Batches) {
		t.Errorf("expected batches %d, got %v", len(rollup.Batches), state["batches"])
	}

	if state["challenges"] != len(rollup.Challenges) {
		t.Errorf("expected challenges %d, got %v", len(rollup.Challenges), state["challenges"])
	}
}

// TestGetMetrics tests getting rollup metrics
func TestGetMetrics(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	metrics := rollup.GetMetrics()

	if metrics.TotalBatches != 0 {
		t.Errorf("expected 0 total batches, got %d", metrics.TotalBatches)
	}

	if metrics.TotalTransactions != 0 {
		t.Errorf("expected 0 total transactions, got %d", metrics.TotalTransactions)
	}

	if metrics.TotalGasUsed != 0 {
		t.Errorf("expected 0 total gas used, got %d", metrics.TotalGasUsed)
	}

	if metrics.TotalChallenges != 0 {
		t.Errorf("expected 0 total challenges, got %d", metrics.TotalChallenges)
	}

	if metrics.ChallengeRate != 0 {
		t.Errorf("expected 0 challenge rate, got %f", metrics.ChallengeRate)
	}
}

// TestGetBatch tests getting batches by batch number
func TestGetBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test getting non-existent batch
	_, err := rollup.GetBatch(1)
	if err == nil {
		t.Error("expected error for non-existent batch")
	}

	// Add transactions and process batch
	tx := createValidTransaction("tx1")
	rollup.AddTransaction(tx)
	rollup.ProcessBatch()

	// Test getting existing batch
	batch, err := rollup.GetBatch(0)
	if err != nil {
		t.Errorf("failed to get batch: %v", err)
	}

	if batch == nil {
		t.Fatal("batch should not be nil")
	}

	if batch.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", batch.BatchNumber)
	}
}

// TestGetChallenge tests getting challenges by ID
func TestGetChallenge(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{
		MinStake: big.NewInt(1000000000000000000), // 1 ETH
	})

	// Test getting non-existent challenge
	_, err := rollup.GetChallenge("non_existent")
	if err == nil {
		t.Error("expected error for non-existent challenge")
	}

	// Add transactions, process batch, and create challenge
	tx := createValidTransaction("tx1")
	rollup.AddTransaction(tx)
	rollup.ProcessBatch()

	challenger := [20]byte{1, 2, 3, 4, 5}
	evidence := []byte("fraud evidence")
	stake := big.NewInt(2000000000000000000) // 2 ETH

	challenge, err := rollup.ChallengeBatch(0, challenger, evidence, stake)
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	// Test getting existing challenge
	retrievedChallenge, err := rollup.GetChallenge(challenge.ID)
	if err != nil {
		t.Errorf("failed to get challenge: %v", err)
	}

	if retrievedChallenge == nil {
		t.Fatal("challenge should not be nil")
	}

	if retrievedChallenge.ID != challenge.ID {
		t.Errorf("expected challenge ID %s, got %s", challenge.ID, retrievedChallenge.ID)
	}
}

// TestValidateTransaction tests transaction validation
func TestValidateTransaction(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test valid transaction
	tx := createValidTransaction("tx1")
	err := rollup.ValidateTransaction(tx)
	if err != nil {
		t.Errorf("valid transaction should not fail validation: %v", err)
	}

	// Test invalid transaction
	tx.ID = ""
	err = rollup.ValidateTransaction(tx)
	if err == nil {
		t.Error("invalid transaction should fail validation")
	}
}

// TestGenerateTransactionHash tests transaction hash generation
func TestGenerateTransactionHash(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	tx1 := createValidTransaction("tx1")
	tx2 := createValidTransaction("tx2")

	hash1 := rollup.generateTransactionHash(tx1)
	hash2 := rollup.generateTransactionHash(tx2)

	// Hashes should be different for different transactions
	if hash1 == hash2 {
		t.Error("different transactions should have different hashes")
	}

	// Hash should be consistent for same transaction
	hash1Again := rollup.generateTransactionHash(tx1)
	if hash1 != hash1Again {
		t.Error("same transaction should have consistent hash")
	}
}

// TestUpdateState tests state updates
func TestUpdateState(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	batch := &Batch{
		ID:              "test_batch",
		BatchNumber:     0,
		StateRoot:       [32]byte{1, 2, 3, 4},
		Transactions:    []Transaction{},
		Timestamp:       time.Now(),
		GasUsed:         21000,
		Success:         true,
		ChallengePeriod: time.Now().Add(time.Hour),
		Finalized:       false,
	}

	err := rollup.updateState(batch)
	if err != nil {
		t.Errorf("failed to update state: %v", err)
	}
}

// TestRollbackBatch tests batch rollback
func TestRollbackBatch(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Add transactions and process batch
	tx := createValidTransaction("tx1")
	rollup.AddTransaction(tx)
	rollup.ProcessBatch()

	// Rollback batch
	err := rollup.rollbackBatch(0)
	if err != nil {
		t.Errorf("failed to rollback batch: %v", err)
	}

	// Verify batch is marked as unsuccessful
	batch, err := rollup.GetBatch(0)
	if err != nil {
		t.Fatalf("failed to get rolled back batch: %v", err)
	}

	if batch.Success {
		t.Error("rolled back batch should be marked as unsuccessful")
	}
}

// TestUpdateMetrics tests metrics updates
func TestUpdateMetrics(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	batchResult := &BatchResult{
		BatchNumber:    0,
		StateRoot:      [32]byte{1, 2, 3, 4},
		GasUsed:        21000,
		Transactions:   1,
		ProcessingTime: time.Millisecond * 50,
		Success:        true,
		Error:          nil,
	}

	processingTime := time.Millisecond * 100
	rollup.updateMetrics(batchResult, processingTime)

	metrics := rollup.GetMetrics()
	if metrics.TotalBatches != 1 {
		t.Errorf("expected 1 total batch, got %d", metrics.TotalBatches)
	}

	if metrics.TotalGasUsed != 21000 {
		t.Errorf("expected 21000 total gas used, got %d", metrics.TotalGasUsed)
	}

	if metrics.AverageBatchTime != processingTime {
		t.Errorf("expected average batch time %v, got %v", processingTime, metrics.AverageBatchTime)
	}

	if metrics.ChallengeRate != 0 {
		t.Errorf("expected challenge rate 0, got %f", metrics.ChallengeRate)
	}
}

// TestGenerateRollupID tests rollup ID generation
func TestGenerateRollupID(t *testing.T) {
	id1 := generateRollupID()
	id2 := generateRollupID()

	if id1 == "" {
		t.Error("rollup ID should not be empty")
	}

	if id2 == "" {
		t.Error("rollup ID should not be empty")
	}

	// IDs should be different (generated at different times)
	if id1 == id2 {
		t.Error("rollup IDs should be different")
	}

	// ID should start with "optimistic_rollup_"
	if len(id1) < 18 || id1[:18] != "optimistic_rollup_" {
		t.Errorf("rollup ID should start with 'optimistic_rollup_', got %s", id1)
	}

	// ID should be exactly 34 characters (optimistic_rollup_ + 16 hex chars)
	if len(id1) != 34 {
		t.Errorf("rollup ID should be exactly 34 characters, got %d: %s", len(id1), id1)
	}
}

// TestMockImplementations tests mock implementations
func TestMockImplementations(t *testing.T) {
	// Test MockStateManager
	stateManager := NewMockStateManager()

	key := [32]byte{1, 2, 3, 4}
	value, err := stateManager.GetState(key)
	if err != nil {
		t.Errorf("mock state manager GetState failed: %v", err)
	}
	if string(value) != "mock_state" {
		t.Errorf("expected 'mock_state', got %s", value)
	}

	err = stateManager.SetState(key, []byte("test"))
	if err != nil {
		t.Errorf("mock state manager SetState failed: %v", err)
	}

	stateRoot, err := stateManager.CommitState()
	if err != nil {
		t.Errorf("mock state manager CommitState failed: %v", err)
	}
	if stateRoot != [32]byte{1, 2, 3, 4} {
		t.Errorf("expected [1,2,3,4], got %v", stateRoot)
	}

	// Test RollbackState
	err = stateManager.RollbackState()
	if err != nil {
		t.Errorf("mock state manager RollbackState failed: %v", err)
	}

	// Test GetStateRoot
	root := stateManager.GetStateRoot()
	if root != [32]byte{1, 2, 3, 4} {
		t.Errorf("expected [1,2,3,4], got %v", root)
	}

	// Test MockFraudProofVerifier
	verifier := NewMockFraudProofVerifier()

	challenge := &Challenge{
		ID:          "test_challenge",
		BatchNumber: 0,
		Challenger:  [20]byte{1, 2, 3, 4, 5},
		Evidence:    []byte("test evidence"),
		Timestamp:   time.Now(),
		Resolved:    false,
		Valid:       false,
		Stake:       big.NewInt(1000000000000000000),
	}

	valid, err := verifier.VerifyFraudProof(challenge)
	if err != nil {
		t.Errorf("mock verifier VerifyFraudProof failed: %v", err)
	}
	if valid {
		t.Error("mock verifier should always return false")
	}

	proof, err := verifier.GenerateFraudProof(&Batch{})
	if err != nil {
		t.Errorf("mock verifier GenerateFraudProof failed: %v", err)
	}
	if string(proof) != "mock_fraud_proof" {
		t.Errorf("expected 'mock_fraud_proof', got %s", proof)
	}

	validChallenge := verifier.ValidateChallenge(challenge)
	if !validChallenge {
		t.Error("mock verifier should validate non-empty evidence")
	}

	// Test MockBatchProcessor
	processor := NewMockBatchProcessor()

	transactions := []Transaction{
		createValidTransaction("tx1"),
		createValidTransaction("tx2"),
	}

	err = processor.ValidateBatch(transactions)
	if err != nil {
		t.Errorf("mock processor ValidateBatch failed: %v", err)
	}

	err = processor.ValidateBatch([]Transaction{})
	if err == nil {
		t.Error("mock processor should reject empty batch")
	}

	result, err := processor.ProcessBatch(transactions)
	if err != nil {
		t.Errorf("mock processor ProcessBatch failed: %v", err)
	}
	if result == nil {
		t.Fatal("batch result should not be nil")
	}
	if result.Transactions != 2 {
		t.Errorf("expected 2 transactions, got %d", result.Transactions)
	}

	optimized, err := processor.OptimizeBatch(transactions)
	if err != nil {
		t.Errorf("mock processor OptimizeBatch failed: %v", err)
	}
	if len(optimized) != 2 {
		t.Errorf("expected 2 optimized transactions, got %d", len(optimized))
	}
}

// TestConcurrency tests concurrent access to rollup
func TestConcurrency(t *testing.T) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})

	// Test concurrent transaction addition
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			tx := createValidTransaction(fmt.Sprintf("tx%d", id))
			err := rollup.AddTransaction(tx)
			if err != nil {
				t.Errorf("failed to add transaction %d: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if len(rollup.Transactions) != 10 {
		t.Errorf("expected 10 transactions, got %d", len(rollup.Transactions))
	}
}

// TestSecurityLevels tests security level constants
func TestSecurityLevels(t *testing.T) {
	if SecurityLevelLow != 0 {
		t.Errorf("expected SecurityLevelLow to be 0, got %d", SecurityLevelLow)
	}

	if SecurityLevelMedium != 1 {
		t.Errorf("expected SecurityLevelMedium to be 1, got %d", SecurityLevelMedium)
	}

	if SecurityLevelHigh != 2 {
		t.Errorf("expected SecurityLevelHigh to be 2, got %d", SecurityLevelHigh)
	}

	if SecurityLevelUltra != 3 {
		t.Errorf("expected SecurityLevelUltra to be 3, got %d", SecurityLevelUltra)
	}
}

// Helper function to create valid transactions for testing
func createValidTransaction(id string) Transaction {
	return Transaction{
		ID:         id,
		From:       [20]byte{1, 2, 3, 4, 5},
		To:         [20]byte{6, 7, 8, 9, 10},
		Value:      big.NewInt(1000000),
		Data:       []byte("test data"),
		Nonce:      1,
		Signature:  []byte("test signature"),
		Timestamp:  time.Now(),
		GasLimit:   21000,
		GasPrice:   big.NewInt(20000000000), // 20 gwei
		RollupHash: [32]byte{},
	}
}

// Failing implementations for testing error paths
type FailingBatchProcessor struct{}

func (f *FailingBatchProcessor) ProcessBatch(transactions []Transaction) (*BatchResult, error) {
	return nil, fmt.Errorf("mock processing failure")
}

func (f *FailingBatchProcessor) ValidateBatch(transactions []Transaction) error {
	return fmt.Errorf("mock validation failure")
}

func (f *FailingBatchProcessor) OptimizeBatch(transactions []Transaction) ([]Transaction, error) {
	return nil, fmt.Errorf("mock optimization failure")
}

type ProcessingFailingBatchProcessor struct{}

func (f *ProcessingFailingBatchProcessor) ProcessBatch(transactions []Transaction) (*BatchResult, error) {
	return nil, fmt.Errorf("mock processing failure")
}

func (f *ProcessingFailingBatchProcessor) ValidateBatch(transactions []Transaction) error {
	return nil
}

func (f *ProcessingFailingBatchProcessor) OptimizeBatch(transactions []Transaction) ([]Transaction, error) {
	return transactions, nil
}

type FailingFraudProofVerifier struct{}

func (f *FailingFraudProofVerifier) VerifyFraudProof(challenge *Challenge) (bool, error) {
	return false, fmt.Errorf("mock verification failure")
}

func (f *FailingFraudProofVerifier) GenerateFraudProof(batch *Batch) ([]byte, error) {
	return nil, fmt.Errorf("mock proof generation failure")
}

func (f *FailingFraudProofVerifier) ValidateChallenge(challenge *Challenge) bool {
	return false
}

type FailingStateManager struct{}

func (f *FailingStateManager) GetState(key [32]byte) ([]byte, error) {
	return nil, fmt.Errorf("mock state retrieval failure")
}

func (f *FailingStateManager) SetState(key [32]byte, value []byte) error {
	return fmt.Errorf("mock state setting failure")
}

func (f *FailingStateManager) CommitState() ([32]byte, error) {
	return [32]byte{}, fmt.Errorf("mock commit failure")
}

func (f *FailingStateManager) RollbackState() error {
	return fmt.Errorf("mock rollback failure")
}

func (f *FailingStateManager) GetStateRoot() [32]byte {
	return [32]byte{}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Benchmark tests for performance
func BenchmarkAddTransaction(b *testing.B) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{MaxBatchSize: 10000})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := createValidTransaction(fmt.Sprintf("tx%d", i))
		rollup.AddTransaction(tx)
	}
}

func BenchmarkProcessBatch(b *testing.B) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{MaxBatchSize: 1000})

	// Pre-populate with transactions
	for i := 0; i < 1000; i++ {
		tx := createValidTransaction(fmt.Sprintf("tx%d", i))
		rollup.AddTransaction(tx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rollup.ProcessBatch()
	}
}

func BenchmarkChallengeBatch(b *testing.B) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{
		MinStake: big.NewInt(1000000000000000000),
	})

	// Pre-populate with a batch
	tx := createValidTransaction("benchmark_tx")
	rollup.AddTransaction(tx)
	rollup.ProcessBatch()

	challenger := [20]byte{1, 2, 3, 4, 5}
	evidence := []byte("benchmark evidence")
	stake := big.NewInt(2000000000000000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rollup.ChallengeBatch(0, challenger, evidence, stake)
	}
}

func BenchmarkGenerateTransactionHash(b *testing.B) {
	rollup := NewOptimisticRollup(OptimisticRollupConfig{})
	tx := createValidTransaction("benchmark_tx")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rollup.generateTransactionHash(tx)
	}
}
