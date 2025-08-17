package bridge

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBridge(t *testing.T) {
	t.Run("create bridge with default config", func(t *testing.T) {
		bridge := NewBridge(nil)
		require.NotNil(t, bridge)
		assert.Equal(t, BridgeStatusActive, bridge.config.Status)
		assert.Equal(t, 3, bridge.config.MinValidators)
		assert.Equal(t, 2, bridge.config.RequiredConfirmations)
		assert.Equal(t, "adrenochain_bridge", bridge.config.ID)
	})

	t.Run("create bridge with custom config", func(t *testing.T) {
		config := &BridgeConfig{
			ID:                    "custom_bridge",
			Status:                BridgeStatusPaused,
			MinValidators:         5,
			RequiredConfirmations: 3,
		}
		bridge := NewBridge(config)
		require.NotNil(t, bridge)
		assert.Equal(t, config.ID, bridge.config.ID)
		assert.Equal(t, config.Status, bridge.config.Status)
		assert.Equal(t, config.MinValidators, bridge.config.MinValidators)
		assert.Equal(t, config.RequiredConfirmations, bridge.config.RequiredConfirmations)
	})
}

func TestBridgeInitiateTransfer(t *testing.T) {
	bridge := NewBridge(nil)

	t.Run("initiate valid transfer", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000) // 1 ETH
		transaction, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		require.NoError(t, err)
		require.NotNil(t, transaction)
		assert.Equal(t, ChainIDadrenochain, transaction.SourceChain)
		assert.Equal(t, ChainIDEthereum, transaction.DestinationChain)
		assert.Equal(t, amount, transaction.Amount)
		assert.Equal(t, TransactionStatusPending, transaction.Status)
		assert.NotEmpty(t, transaction.ID)
		assert.NotZero(t, transaction.Fee)
	})

	t.Run("initiate transfer with same chains", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDadrenochain, // Same as source
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source and destination chains cannot be the same")
	})

	t.Run("initiate transfer with invalid source chain", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000)
		_, err := bridge.InitiateTransfer(
			"invalid_chain",
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid source chain")
	})

	t.Run("initiate transfer with invalid destination chain", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			"invalid_chain",
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination chain")
	})

	t.Run("initiate transfer with empty addresses", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"", // Empty source address
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAddress, err)
	})

	t.Run("initiate transfer with amount too small", func(t *testing.T) {
		amount := big.NewInt(100000000000000) // 0.0001 ETH (below minimum)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "below minimum")
	})

	t.Run("initiate transfer with amount too large", func(t *testing.T) {
		amount := big.NewInt(2000000000000000000) // 2 ETH (above maximum)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum")
	})

	t.Run("initiate transfer with unsupported asset", func(t *testing.T) {
		amount := big.NewInt(1000000000000000000)
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeERC20,
			"0x1234567890123456789012345678901234567890", // Unsupported ERC20
			amount,
			nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrAssetNotSupported, err)
	})
}

func TestBridgeConfirmTransaction(t *testing.T) {
	bridge := NewBridge(nil)

	t.Run("confirm valid transaction", func(t *testing.T) {
		// Create a fresh test transaction for this test
		amount := big.NewInt(1000000000000000000)
		transaction, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)
		require.NoError(t, err)

		// Add a validator first
		validator := &Validator{
			ID:            "validator1",
			Address:       "0x1111111111111111111111111111111111111111",
			ChainID:       ChainIDadrenochain,
			StakeAmount:   big.NewInt(1000000000000000000),
			IsActive:      true,
			LastHeartbeat: time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		bridge.validators["validator1"] = validator
		bridge.validatorSet[validator.Address] = true

		err = bridge.ConfirmTransaction(transaction.ID, "validator1", []byte("signature"))
		require.NoError(t, err)

		// Check that transaction was confirmed
		updatedTx, err := bridge.GetTransaction(transaction.ID)
		require.NoError(t, err)
		assert.Equal(t, TransactionStatusConfirmed, updatedTx.Status)
		assert.Equal(t, "validator1", updatedTx.ValidatorID)
	})

	t.Run("confirm non-existent transaction", func(t *testing.T) {
		err := bridge.ConfirmTransaction("non_existent_id", "validator1", []byte("signature"))
		assert.Error(t, err)
		assert.Equal(t, ErrTransactionNotFound, err)
	})

	t.Run("confirm with non-existent validator", func(t *testing.T) {
		// Create a fresh test transaction for this test
		amount := big.NewInt(1000000000000000000)
		transaction, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)
		require.NoError(t, err)

		err = bridge.ConfirmTransaction(transaction.ID, "non_existent_validator", []byte("signature"))
		assert.Error(t, err)
		assert.Equal(t, ErrValidatorInactive, err)
	})
}

func TestBridgeExecuteTransaction(t *testing.T) {
	bridge := NewBridge(nil)

	// Create and confirm a test transaction
	amount := big.NewInt(1000000000000000000)
	transaction, err := bridge.InitiateTransfer(
		ChainIDadrenochain,
		ChainIDEthereum,
		"0x1234567890123456789012345678901234567890",
		"0x0987654321098765432109876543210987654321",
		AssetTypeNative,
		"0x0000000000000000000000000000000000000000",
		amount,
		nil,
	)
	require.NoError(t, err)

	// Add validator and confirm
	validator := &Validator{
		ID:            "validator1",
		Address:       "0x1111111111111111111111111111111111111111",
		ChainID:       ChainIDadrenochain,
		StakeAmount:   big.NewInt(1000000000000000000),
		IsActive:      true,
		LastHeartbeat: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	bridge.validators["validator1"] = validator
	bridge.validatorSet[validator.Address] = true

	err = bridge.ConfirmTransaction(transaction.ID, "validator1", []byte("signature"))
	require.NoError(t, err)

	t.Run("execute confirmed transaction", func(t *testing.T) {
		destinationTxHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		gasUsed := big.NewInt(21000)

		err := bridge.ExecuteTransaction(transaction.ID, destinationTxHash, gasUsed)
		require.NoError(t, err)

		// Check that transaction was executed
		updatedTx, err := bridge.GetTransaction(transaction.ID)
		require.NoError(t, err)
		assert.Equal(t, TransactionStatusExecuted, updatedTx.Status)
		assert.Equal(t, destinationTxHash, updatedTx.DestinationTxHash)
		assert.Equal(t, gasUsed, updatedTx.GasUsed)
		assert.NotNil(t, updatedTx.ExecutedAt)
	})

	t.Run("execute non-existent transaction", func(t *testing.T) {
		err := bridge.ExecuteTransaction("non_existent_id", "txhash", big.NewInt(21000))
		assert.Error(t, err)
		assert.Equal(t, ErrTransactionNotFound, err)
	})

	t.Run("execute pending transaction", func(t *testing.T) {
		// Create a new pending transaction
		amount2 := big.NewInt(500000000000000000) // 0.5 ETH
		transaction2, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount2,
			nil,
		)
		require.NoError(t, err)

		err = bridge.ExecuteTransaction(transaction2.ID, "txhash", big.NewInt(21000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected confirmed")
	})
}

func TestBridgeFailTransaction(t *testing.T) {
	bridge := NewBridge(nil)

	// Create a test transaction
	amount := big.NewInt(1000000000000000000)
	transaction, err := bridge.InitiateTransfer(
		ChainIDadrenochain,
		ChainIDEthereum,
		"0x1234567890123456789012345678901234567890",
		"0x0987654321098765432109876543210987654321",
		AssetTypeNative,
		"0x0000000000000000000000000000000000000000",
		amount,
		nil,
	)
	require.NoError(t, err)

	t.Run("fail valid transaction", func(t *testing.T) {
		reason := "insufficient gas"
		err := bridge.FailTransaction(transaction.ID, reason)
		require.NoError(t, err)

		// Check that transaction was failed
		updatedTx, err := bridge.GetTransaction(transaction.ID)
		require.NoError(t, err)
		assert.Equal(t, TransactionStatusFailed, updatedTx.Status)
		assert.Equal(t, reason, updatedTx.FailureReason)
		assert.NotNil(t, updatedTx.FailedAt)
	})

	t.Run("fail non-existent transaction", func(t *testing.T) {
		err := bridge.FailTransaction("non_existent_id", "reason")
		assert.Error(t, err)
		assert.Equal(t, ErrTransactionNotFound, err)
	})
}

func TestBridgeGetTransactions(t *testing.T) {
	bridge := NewBridge(nil)

	// Create multiple test transactions
	amount := big.NewInt(1000000000000000000)
	address1 := "0x1234567890123456789012345678901234567890"
	address2 := "0x0987654321098765432109876543210987654321"

	// Transaction 1
	tx1, err := bridge.InitiateTransfer(
		ChainIDadrenochain,
		ChainIDEthereum,
		address1,
		address2,
		AssetTypeNative,
		"0x0000000000000000000000000000000000000000",
		amount,
		nil,
	)
	require.NoError(t, err)

	// Transaction 2
	_, err = bridge.InitiateTransfer(
		ChainIDEthereum,
		ChainIDadrenochain,
		address2,
		address1,
		AssetTypeNative,
		"0x0000000000000000000000000000000000000000",
		amount,
		nil,
	)
	require.NoError(t, err)

	t.Run("get transaction by ID", func(t *testing.T) {
		retrievedTx, err := bridge.GetTransaction(tx1.ID)
		require.NoError(t, err)
		assert.Equal(t, tx1.ID, retrievedTx.ID)
		assert.Equal(t, tx1.SourceChain, retrievedTx.SourceChain)
		assert.Equal(t, tx1.DestinationChain, retrievedTx.DestinationChain)
	})

	t.Run("get transactions by status", func(t *testing.T) {
		pendingTxs := bridge.GetTransactionsByStatus(TransactionStatusPending)
		assert.Len(t, pendingTxs, 2)

		// Fail one transaction
		err := bridge.FailTransaction(tx1.ID, "test failure")
		require.NoError(t, err)

		pendingTxs = bridge.GetTransactionsByStatus(TransactionStatusPending)
		assert.Len(t, pendingTxs, 1)

		failedTxs := bridge.GetTransactionsByStatus(TransactionStatusFailed)
		assert.Len(t, failedTxs, 1)
		assert.Equal(t, tx1.ID, failedTxs[0].ID)
	})

	t.Run("get transactions by address", func(t *testing.T) {
		address1Txs := bridge.GetTransactionsByAddress(address1)
		assert.Len(t, address1Txs, 2) // Both as source and destination

		address2Txs := bridge.GetTransactionsByAddress(address2)
		assert.Len(t, address2Txs, 2) // Both as source and destination
	})
}

func TestBridgeGetBridgeStats(t *testing.T) {
	bridge := NewBridge(nil)

	// Create some test transactions
	amount := big.NewInt(1000000000000000000)
	for i := 0; i < 3; i++ {
		_, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)
		require.NoError(t, err)
	}

	// Fail one transaction
	transactions := bridge.GetTransactionsByStatus(TransactionStatusPending)
	require.Len(t, transactions, 3)
	err := bridge.FailTransaction(transactions[0].ID, "test failure")
	require.NoError(t, err)

	t.Run("get bridge statistics", func(t *testing.T) {
		stats := bridge.GetBridgeStats()

		assert.Equal(t, 3, stats["total_transactions"])
		assert.Equal(t, 2, stats["pending_transactions"])
		assert.Equal(t, 0, stats["confirmed_transactions"])
		assert.Equal(t, 0, stats["executed_transactions"])
		assert.Equal(t, 1, stats["failed_transactions"])
		assert.Equal(t, 0, stats["total_validators"])
		assert.Equal(t, 0, stats["active_validators"])
		assert.Equal(t, 2, stats["total_asset_mappings"]) // Default mappings
		assert.Equal(t, BridgeStatusActive, stats["bridge_status"])
		assert.NotEmpty(t, stats["daily_volume_used"])
		assert.NotEmpty(t, stats["max_daily_volume"])
	})
}

func TestBridgeEventHandling(t *testing.T) {
	bridge := NewBridge(nil)

	t.Run("register and emit events", func(t *testing.T) {
		eventReceived := false
		var receivedData interface{}

		// Register event handler
		bridge.On("transfer_initiated", func(data interface{}) {
			eventReceived = true
			receivedData = data
		})

		// Trigger event
		amount := big.NewInt(1000000000000000000)
		transaction, err := bridge.InitiateTransfer(
			ChainIDadrenochain,
			ChainIDEthereum,
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
			AssetTypeNative,
			"0x0000000000000000000000000000000000000000",
			amount,
			nil,
		)
		require.NoError(t, err)

		// Wait a bit for the event to be processed
		time.Sleep(100 * time.Millisecond)

		assert.True(t, eventReceived)
		assert.Equal(t, transaction, receivedData)
	})
}

func TestBridgeValidation(t *testing.T) {
	bridge := NewBridge(nil)

	t.Run("validate bridge status", func(t *testing.T) {
		// Bridge should be active by default
		err := bridge.validateBridgeStatus()
		assert.NoError(t, err)

		// Test paused status
		bridge.config.Status = BridgeStatusPaused
		err = bridge.validateBridgeStatus()
		assert.Error(t, err)
		assert.Equal(t, ErrBridgePaused, err)

		// Test emergency status
		bridge.config.Status = BridgeStatusEmergency
		err = bridge.validateBridgeStatus()
		assert.Error(t, err)
		assert.Equal(t, ErrBridgeEmergency, err)

		// Reset to active
		bridge.config.Status = BridgeStatusActive
	})

	t.Run("validate chains", func(t *testing.T) {
		// Valid chains
		err := bridge.validateChains(ChainIDadrenochain, ChainIDEthereum)
		assert.NoError(t, err)

		// Same chains
		err = bridge.validateChains(ChainIDadrenochain, ChainIDadrenochain)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be the same")

		// Invalid source chain
		err = bridge.validateChains("invalid_chain", ChainIDEthereum)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid source chain")

		// Invalid destination chain
		err = bridge.validateChains(ChainIDadrenochain, "invalid_chain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination chain")
	})

	t.Run("validate addresses", func(t *testing.T) {
		// Valid addresses
		err := bridge.validateAddresses(
			"0x1234567890123456789012345678901234567890",
			"0x0987654321098765432109876543210987654321",
		)
		assert.NoError(t, err)

		// Empty addresses
		err = bridge.validateAddresses("", "0x0987654321098765432109876543210987654321")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAddress, err)

		err = bridge.validateAddresses("0x1234567890123456789012345678901234567890", "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAddress, err)

		// Short addresses
		err = bridge.validateAddresses("0x123", "0x0987654321098765432109876543210987654321")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAddress, err)
	})
}
