package bridge

import (
	"math/big"
	"testing"
	"time"
)

func TestValidatorManager(t *testing.T) {
	// Create bridge and validator manager
	bridge := NewBridge(nil)
	vm := bridge.GetValidatorManager()

	// Test adding validators
	t.Run("AddValidator", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		stakeAmount := big.NewInt(2000000000000000000) // 2 ETH
		
		validator, err := vm.AddValidator(address, ChainIDGoChain, stakeAmount, nil)
		if err != nil {
			t.Fatalf("Failed to add validator: %v", err)
		}

		if validator.Address != address {
			t.Errorf("Expected address %s, got %s", address, validator.Address)
		}

		if validator.StakeAmount.Cmp(stakeAmount) != 0 {
			t.Errorf("Expected stake amount %s, got %s", stakeAmount.String(), validator.StakeAmount.String())
		}

		if !validator.IsActive {
			t.Error("Expected validator to be active")
		}
	})

	t.Run("AddValidatorInsufficientStake", func(t *testing.T) {
		address := "0x2345678901234567890123456789012345678901"
		stakeAmount := big.NewInt(500000000000000000) // 0.5 ETH (below threshold)
		
		_, err := vm.AddValidator(address, ChainIDGoChain, stakeAmount, nil)
		if err == nil {
			t.Error("Expected error for insufficient stake")
		}
	})

	t.Run("AddDuplicateValidator", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		stakeAmount := big.NewInt(2000000000000000000) // 2 ETH
		
		_, err := vm.AddValidator(address, ChainIDGoChain, stakeAmount, nil)
		if err == nil {
			t.Error("Expected error for duplicate validator")
		}
	})

	t.Run("GetActiveValidators", func(t *testing.T) {
		validators := vm.GetActiveValidators()
		if len(validators) != 1 {
			t.Errorf("Expected 1 active validator, got %d", len(validators))
		}
	})

	t.Run("Heartbeat", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		
		err := vm.Heartbeat(address)
		if err != nil {
			t.Fatalf("Failed to update heartbeat: %v", err)
		}

		validator, err := vm.GetValidator(address)
		if err != nil {
			t.Fatalf("Failed to get validator: %v", err)
		}

		if time.Since(validator.LastHeartbeat) > time.Second {
			t.Error("Heartbeat not updated")
		}
	})

	t.Run("UpdateValidatorStake", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		newStake := big.NewInt(3000000000000000000) // 3 ETH
		
		err := vm.UpdateValidatorStake(address, newStake)
		if err != nil {
			t.Fatalf("Failed to update stake: %v", err)
		}

		validator, err := vm.GetValidator(address)
		if err != nil {
			t.Fatalf("Failed to get validator: %v", err)
		}

		if validator.StakeAmount.Cmp(newStake) != 0 {
			t.Errorf("Expected stake amount %s, got %s", newStake.String(), validator.StakeAmount.String())
		}
	})

	t.Run("RemoveValidator", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		
		err := vm.RemoveValidator(address)
		if err != nil {
			t.Fatalf("Failed to remove validator: %v", err)
		}

		validators := vm.GetActiveValidators()
		if len(validators) != 0 {
			t.Errorf("Expected 0 active validators, got %d", len(validators))
		}
	})
}

func TestConsensusEngine(t *testing.T) {
	// Create consensus engine
	ce := NewConsensusEngine(2) // Require 2 confirmations

	t.Run("AddValidator", func(t *testing.T) {
		validator := &Validator{
			ID:       "validator1",
			Address:  "0x1234567890123456789012345678901234567890",
			IsActive: true,
		}

		ce.AddValidator(validator)
		if len(ce.validators) != 1 {
			t.Errorf("Expected 1 validator, got %d", len(ce.validators))
		}
	})

	t.Run("AddConfirmation", func(t *testing.T) {
		txID := "tx123"
		confirmation := &Confirmation{
			ValidatorID:   "validator1",
			TransactionID: txID,
			IsValid:       true,
			Timestamp:     time.Now(),
		}

		ce.AddConfirmation(txID, confirmation)
		if len(ce.confirmations[txID]) != 1 {
			t.Errorf("Expected 1 confirmation, got %d", len(ce.confirmations[txID]))
		}
	})

	t.Run("HasEnoughConfirmations", func(t *testing.T) {
		txID := "tx123"
		
		// Add second confirmation
		confirmation2 := &Confirmation{
			ValidatorID:   "validator2",
			TransactionID: txID,
			IsValid:       true,
			Timestamp:     time.Now(),
		}
		ce.AddConfirmation(txID, confirmation2)

		if !ce.HasEnoughConfirmations(txID) {
			t.Error("Expected enough confirmations")
		}
	})

	t.Run("HasPendingConfirmations", func(t *testing.T) {
		if !ce.HasPendingConfirmations("validator1") {
			t.Error("Expected validator1 to have pending confirmations")
		}
	})

	t.Run("GetConfirmations", func(t *testing.T) {
		txID := "tx123"
		confirmations := ce.GetConfirmations(txID)
		if len(confirmations) != 2 {
			t.Errorf("Expected 2 confirmations, got %d", len(confirmations))
		}
	})
}

func TestCrossChainTransactionManager(t *testing.T) {
	// Create bridge and transaction manager
	bridge := NewBridge(nil)
	ctm := bridge.GetCrossChainTransactionManager()

	t.Run("InitiateBatchTransfer", func(t *testing.T) {
		transfers := []*TransferRequest{
			{
				SourceChain:       ChainIDGoChain,
				DestinationChain:  ChainIDEthereum,
				SourceAddress:     "0x1234567890123456789012345678901234567890",
				DestinationAddress: "0x0987654321098765432109876543210987654321",
				AssetType:         AssetTypeNative,
				AssetAddress:      "0x0000000000000000000000000000000000000000",
				Amount:            big.NewInt(1000000000000000000), // 1 ETH
			},
		}

		batch, err := ctm.InitiateBatchTransfer(transfers)
		if err != nil {
			t.Fatalf("Failed to initiate batch transfer: %v", err)
		}

		if batch.Status != BatchStatusPending {
			t.Errorf("Expected status %s, got %s", BatchStatusPending, batch.Status)
		}

		if len(batch.Transactions) != 1 {
			t.Errorf("Expected 1 transaction, got %d", len(batch.Transactions))
		}
	})

	t.Run("ProcessBatch", func(t *testing.T) {
		// Get the batch we just created
		batches := ctm.GetBatchesByStatus(BatchStatusPending)
		if len(batches) == 0 {
			t.Fatal("No pending batches found")
		}

		batchID := batches[0].ID
		err := ctm.ProcessBatch(batchID)
		if err != nil {
			t.Fatalf("Failed to process batch: %v", err)
		}

		batch, err := ctm.GetBatch(batchID)
		if err != nil {
			t.Fatalf("Failed to get batch: %v", err)
		}

		if batch.Status != BatchStatusConfirmed {
			t.Errorf("Expected status %s, got %s", BatchStatusConfirmed, batch.Status)
		}
	})

	t.Run("ExecuteBatch", func(t *testing.T) {
		// Get the confirmed batch
		batches := ctm.GetBatchesByStatus(BatchStatusConfirmed)
		if len(batches) == 0 {
			t.Fatal("No confirmed batches found")
		}

		batchID := batches[0].ID
		destinationTxHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		gasUsed := big.NewInt(21000)

		err := ctm.ExecuteBatch(batchID, destinationTxHash, gasUsed)
		if err != nil {
			t.Fatalf("Failed to execute batch: %v", err)
		}

		batch, err := ctm.GetBatch(batchID)
		if err != nil {
			t.Fatalf("Failed to get batch: %v", err)
		}

		if batch.Status != BatchStatusExecuted {
			t.Errorf("Expected status %s, got %s", BatchStatusExecuted, batch.Status)
		}
	})

	t.Run("GetBatchStats", func(t *testing.T) {
		stats := ctm.GetBatchStats()
		
		expectedKeys := []string{"total_batches", "pending_batches", "confirmed_batches", "executed_batches", "failed_batches"}
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})
}

func TestSecurityManager(t *testing.T) {
	t.Run("CheckRateLimit", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		address := "0x1234567890123456789012345678901234567890"
		
		// First 10 requests should succeed
		for i := 0; i < 10; i++ {
			err := sm.CheckRateLimit(address)
			if err != nil {
				t.Fatalf("Rate limit check %d failed: %v", i+1, err)
			}
		}

		// 11th request should fail
		err := sm.CheckRateLimit(address)
		if err == nil {
			t.Error("Expected rate limit to be exceeded")
		}
	})

	t.Run("CheckTransferSecurity", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		sourceAddress := "0x1234567890123456789012345678901234567890"
		destinationAddress := "0x0987654321098765432109876543210987654321"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		assetType := AssetTypeNative

		err := sm.CheckTransferSecurity(sourceAddress, destinationAddress, amount, assetType)
		if err != nil {
			t.Fatalf("Transfer security check failed: %v", err)
		}
	})

	t.Run("AddSuspiciousPattern", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		pattern := "high_frequency"
		riskScore := 0.8
		threshold := 0.7

		err := sm.AddSuspiciousPattern(pattern, riskScore, threshold)
		if err != nil {
			t.Fatalf("Failed to add suspicious pattern: %v", err)
		}
	})

	t.Run("BlacklistAddress", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		address := "0x1111111111111111111111111111111111111111"
		reason := "Suspicious activity detected"

		err := sm.BlacklistAddress(address, reason)
		if err != nil {
			t.Fatalf("Failed to blacklist address: %v", err)
		}
	})

	t.Run("PauseBridge", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		pausedBy := "admin"
		reason := "Emergency maintenance"

		err := sm.PauseBridge(pausedBy, reason)
		if err != nil {
			t.Fatalf("Failed to pause bridge: %v", err)
		}
	})

	t.Run("ResumeBridge", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		// First pause the bridge
		pausedBy := "admin"
		reason := "Emergency maintenance"
		err := sm.PauseBridge(pausedBy, reason)
		if err != nil {
			t.Fatalf("Failed to pause bridge: %v", err)
		}
		
		// Then resume it
		resumedBy := "admin"
		err = sm.ResumeBridge(resumedBy)
		if err != nil {
			t.Fatalf("Failed to resume bridge: %v", err)
		}
	})

	t.Run("GetSecurityStats", func(t *testing.T) {
		// Create fresh bridge and security manager for this test
		bridge := NewBridge(nil)
		sm := bridge.GetSecurityManager()
		
		stats := sm.GetSecurityStats()
		
		expectedKeys := []string{"total_security_events", "emergency_paused", "blacklisted_addresses", "suspicious_patterns", "rate_limiters"}
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})
}

func TestBridgeIntegration(t *testing.T) {
	// Test full bridge integration
	bridge := NewBridge(nil)

	t.Run("BridgeStats", func(t *testing.T) {
		stats := bridge.GetBridgeStats()
		
		// Check that all expected stats are present
		expectedKeys := []string{
			"total_transactions", "pending_transactions", "confirmed_transactions",
			"executed_transactions", "failed_transactions", "total_validators",
			"active_validators", "total_asset_mappings", "bridge_status",
			"daily_volume_used", "max_daily_volume", "active_validators_count",
		}
		
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})

	t.Run("ManagerAccess", func(t *testing.T) {
		// Test that all managers are accessible
		if bridge.GetValidatorManager() == nil {
			t.Error("Validator manager is nil")
		}

		if bridge.GetCrossChainTransactionManager() == nil {
			t.Error("Cross-chain transaction manager is nil")
		}

		if bridge.GetSecurityManager() == nil {
			t.Error("Security manager is nil")
		}
	})
}
