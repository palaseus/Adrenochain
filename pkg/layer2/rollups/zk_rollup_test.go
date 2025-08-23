package rollups

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
)

// TestNewZKRollup tests the creation of a new ZK rollup
func TestNewZKRollup(t *testing.T) {
	config := ZKRollupConfig{
		MaxBatchSize:      1000,
		MaxProofTime:      time.Second * 30,
		VerificationDelay: time.Second * 5,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelHigh,
	}

	rollup := NewZKRollup(config)

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

	if len(rollup.Proofs) != 0 {
		t.Errorf("expected empty proofs, got %d", len(rollup.Proofs))
	}

	if rollup.config.MaxBatchSize != 1000 {
		t.Errorf("expected max batch size 1000, got %d", rollup.config.MaxBatchSize)
	}

	if rollup.config.SecurityLevel != SecurityLevelHigh {
		t.Errorf("expected security level %d, got %d", SecurityLevelHigh, rollup.config.SecurityLevel)
	}
}

// TestAddTransaction tests adding transactions to the rollup
func TestAddTransaction(t *testing.T) {
	config := ZKRollupConfig{
		MaxBatchSize:      2,
		MaxProofTime:      time.Second * 30,
		VerificationDelay: time.Second * 5,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelMedium,
	}

	rollup := NewZKRollup(config)

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
	rollup := NewZKRollup(ZKRollupConfig{})

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
	rollup := NewZKRollup(ZKRollupConfig{})

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

	if len(rollup.Proofs) != 1 {
		t.Errorf("expected 1 proof, got %d", len(rollup.Proofs))
	}
}

// TestProcessBatchValidationFailure tests batch validation failure
func TestProcessBatchValidationFailure(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

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
	rollup := NewZKRollup(ZKRollupConfig{})

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

// TestProcessBatchStateUpdateFailure tests state update failure
func TestProcessBatchStateUpdateFailure(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	// Create a rollup with a custom state manager that always fails
	rollup.StateManager = &FailingStateManager{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail state update
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when state update fails")
	}

	if !contains(err.Error(), "state update failed") {
		t.Errorf("expected state update error, got: %s", err.Error())
	}
}

// TestProcessBatchProofVerificationFailure tests proof verification failure
func TestProcessBatchProofVerificationFailure(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	// Create a rollup with a custom proof verifier that always fails verification
	rollup.Verifier = &VerificationFailingProofVerifier{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail proof verification
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when proof verification fails")
	}

	if !contains(err.Error(), "proof verification failed") {
		t.Errorf("expected proof verification error, got: %s", err.Error())
	}
}

// TestUpdateMetricsEdgeCases tests edge cases in updateMetrics
func TestUpdateMetricsEdgeCases(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

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

// TestGetState tests getting rollup state
func TestGetState(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

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

	if state["proofs"] != len(rollup.Proofs) {
		t.Errorf("expected proofs %d, got %v", len(rollup.Proofs), state["proofs"])
	}
}

// TestGetMetrics tests getting rollup metrics
func TestGetMetrics(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

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

	if metrics.ProofSuccessRate != 0 {
		t.Errorf("expected 0 proof success rate, got %f", metrics.ProofSuccessRate)
	}
}

// TestGetProof tests getting proofs by batch number
func TestGetProof(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	// Test getting proof for non-existent batch
	_, err := rollup.GetProof(1)
	if err == nil {
		t.Error("expected error for non-existent proof")
	}

	// Add transactions and process batch
	tx := createValidTransaction("tx1")
	if err := rollup.AddTransaction(tx); err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}
	if _, err := rollup.ProcessBatch(); err != nil {
		t.Fatalf("failed to process batch: %v", err)
	}

	// Test getting proof for existing batch
	proof, err := rollup.GetProof(0)
	if err != nil {
		t.Errorf("failed to get proof: %v", err)
	}

	if proof == nil {
		t.Fatal("proof should not be nil")
	}

	if proof.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", proof.BatchNumber)
	}
}

// TestValidateTransaction tests transaction validation
func TestValidateTransaction(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

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
	rollup := NewZKRollup(ZKRollupConfig{})

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

// TestGenerateProof tests proof generation
func TestGenerateProof(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	batchResult := &BatchResult{
		BatchNumber:    0,
		StateRoot:      [32]byte{1, 2, 3, 4},
		GasUsed:        42000,
		Transactions:   2,
		ProcessingTime: time.Millisecond * 100,
		Success:        true,
	}

	proof, err := rollup.generateProof(batchResult)
	if err != nil {
		t.Errorf("failed to generate proof: %v", err)
	}

	if proof == nil {
		t.Fatal("proof should not be nil")
	}

	if proof.BatchNumber != 0 {
		t.Errorf("expected batch number 0, got %d", proof.BatchNumber)
	}

	if proof.GasUsed != 42000 {
		t.Errorf("expected gas used 42000, got %d", proof.GasUsed)
	}

	if len(proof.ProofData) == 0 {
		t.Error("proof data should not be empty")
	}

	if len(proof.PublicInputs) == 0 {
		t.Error("public inputs should not be empty")
	}
}

// TestUpdateState tests state updates
func TestUpdateState(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	batchResult := &BatchResult{
		BatchNumber:    0,
		StateRoot:      [32]byte{1, 2, 3, 4},
		GasUsed:        21000,
		Transactions:   1,
		ProcessingTime: time.Millisecond * 50,
		Success:        true,
	}

	err := rollup.updateState(batchResult)
	if err != nil {
		t.Errorf("failed to update state: %v", err)
	}
}

// TestUpdateMetrics tests metrics updates
func TestUpdateMetrics(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	batchResult := &BatchResult{
		BatchNumber:    0,
		StateRoot:      [32]byte{1, 2, 3, 4},
		GasUsed:        21000,
		Transactions:   1,
		ProcessingTime: time.Millisecond * 50,
		Success:        true,
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

	if metrics.ProofSuccessRate != 1.0 {
		t.Errorf("expected proof success rate 1.0, got %f", metrics.ProofSuccessRate)
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

	// ID should start with "rollup_"
	if len(id1) < 7 || id1[:7] != "rollup_" {
		t.Errorf("rollup ID should start with 'rollup_', got %s", id1)
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

	// Test MockProofVerifier
	verifier := NewMockProofVerifier()

	proof := &ZKProof{
		ID:           "test_proof",
		BatchNumber:  0,
		StateRoot:    [32]byte{1, 2, 3, 4},
		ProofData:    []byte("test"),
		PublicInputs: []*big.Int{big.NewInt(1)},
		Verification: false,
		Timestamp:    time.Now(),
		GasUsed:      21000,
	}

	verified, err := verifier.VerifyProof(proof)
	if err != nil {
		t.Errorf("mock verifier VerifyProof failed: %v", err)
	}
	if !verified {
		t.Error("mock verifier should always return true")
	}

	key2, err := verifier.GenerateVerificationKey()
	if err != nil {
		t.Errorf("mock verifier GenerateVerificationKey failed: %v", err)
	}
	if string(key2) != "mock_verification_key" {
		t.Errorf("expected 'mock_verification_key', got %s", key2)
	}

	valid := verifier.ValidatePublicInputs([]*big.Int{big.NewInt(1)})
	if !valid {
		t.Error("mock verifier should validate non-empty inputs")
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
	rollup := NewZKRollup(ZKRollupConfig{})

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

type FailingProofVerifier struct{}

func (f *FailingProofVerifier) VerifyProof(proof *ZKProof) (bool, error) {
	return false, fmt.Errorf("mock verification failure")
}

func (f *FailingProofVerifier) GenerateVerificationKey() ([]byte, error) {
	return nil, fmt.Errorf("mock key generation failure")
}

func (f *FailingProofVerifier) ValidatePublicInputs(inputs []*big.Int) bool {
	return false
}

type VerificationFailingProofVerifier struct{}

func (f *VerificationFailingProofVerifier) VerifyProof(proof *ZKProof) (bool, error) {
	return false, nil // Verification fails but no error
}

func (f *VerificationFailingProofVerifier) GenerateVerificationKey() ([]byte, error) {
	return []byte("mock_key"), nil
}

func (f *VerificationFailingProofVerifier) ValidatePublicInputs(inputs []*big.Int) bool {
	return true
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

type ErroringProofVerifier struct{}

func (e *ErroringProofVerifier) VerifyProof(proof *ZKProof) (bool, error) {
	return false, fmt.Errorf("mock proof verification error")
}

func (e *ErroringProofVerifier) GenerateVerificationKey() ([]byte, error) {
	return nil, fmt.Errorf("mock key generation error")
}

func (e *ErroringProofVerifier) ValidatePublicInputs(inputs []*big.Int) bool {
	return false
}

type NilBatchResultProcessor struct{}

func (n *NilBatchResultProcessor) ProcessBatch(transactions []Transaction) (*BatchResult, error) {
	// Return nil batch result to trigger proof generation failure
	return nil, nil
}

func (n *NilBatchResultProcessor) ValidateBatch(transactions []Transaction) error {
	return nil
}

func (n *NilBatchResultProcessor) OptimizeBatch(transactions []Transaction) ([]Transaction, error) {
	return transactions, nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Benchmark tests for performance
func BenchmarkAddTransaction(b *testing.B) {
	rollup := NewZKRollup(ZKRollupConfig{MaxBatchSize: 10000})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := createValidTransaction(fmt.Sprintf("tx%d", i))
		rollup.AddTransaction(tx)
	}
}

func BenchmarkProcessBatch(b *testing.B) {
	rollup := NewZKRollup(ZKRollupConfig{MaxBatchSize: 1000})

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

func BenchmarkGenerateTransactionHash(b *testing.B) {
	rollup := NewZKRollup(ZKRollupConfig{})
	tx := createValidTransaction("benchmark_tx")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rollup.generateTransactionHash(tx)
	}
}

// TestProcessBatchProofGenerationFailure tests proof generation failure
func TestProcessBatchProofGenerationFailure(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	// Create a rollup with a custom batch processor that returns nil batch result
	rollup.BatchProcessor = &NilBatchResultProcessor{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail proof generation
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when proof generation fails")
	}

	if !contains(err.Error(), "proof generation failed") {
		t.Errorf("expected proof generation error, got: %s", err.Error())
	}
}

// TestProcessBatchProofVerificationError tests proof verification error (not failure)
func TestProcessBatchProofVerificationError(t *testing.T) {
	rollup := NewZKRollup(ZKRollupConfig{})

	// Create a rollup with a custom proof verifier that returns an error during verification
	rollup.Verifier = &ErroringProofVerifier{}

	tx := createValidTransaction("tx1")
	err := rollup.AddTransaction(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to process batch - should fail proof verification with error
	_, err = rollup.ProcessBatch()
	if err == nil {
		t.Error("expected error when proof verification returns error")
	}

	if !contains(err.Error(), "proof verification failed") {
		t.Errorf("expected proof verification error, got: %s", err.Error())
	}
}
