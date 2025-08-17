package advanced

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFlashLoanManager(t *testing.T) {
	t.Run("valid flash loan manager", func(t *testing.T) {
		lendingPool := createTestPool()
		maxAmount := big.NewInt(1000000000) // 1000 USDC
		minAmount := big.NewInt(1000000)    // 1 USDC
		feeRate := big.NewInt(50)           // 0.5%
		maxDuration := 30 * time.Second

		manager := NewFlashLoanManager(lendingPool, maxAmount, minAmount, feeRate, maxDuration)

		require.NotNil(t, manager)
		assert.Equal(t, lendingPool, manager.lendingPool)
		assert.Equal(t, maxAmount, manager.maxFlashLoanAmount)
		assert.Equal(t, minAmount, manager.minFlashLoanAmount)
		assert.Equal(t, feeRate, manager.flashLoanFeeRate)
		assert.Equal(t, maxDuration, manager.maxFlashLoanDuration)
		assert.NotNil(t, manager.activeLoans)
		assert.Equal(t, uint64(0), manager.loanCounter)
	})
}

func TestFlashLoanManagerExecuteFlashLoan(t *testing.T) {
	t.Run("execute flash loan successfully", func(t *testing.T) {
		lendingPool := createTestPool()
		// Supply some assets to the pool
		err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
		require.NoError(t, err)

		manager := NewFlashLoanManager(
			lendingPool,
			big.NewInt(1000000000), // 1000 USDC max
			big.NewInt(1000000),    // 1 USDC min
			big.NewInt(50),         // 0.5% fee
			30*time.Second,
		)

		// Create a successful callback
		callback := func(flashLoan *FlashLoan) error {
			// Simulate some operations with the borrowed funds
			// In a real scenario, this would be arbitrage, liquidation, etc.
			return nil
		}

		flashLoan, err := manager.ExecuteFlashLoan(
			"user2",
			"USDC",
			big.NewInt(5000000), // 5 USDC
			callback,
			[]byte("arbitrage_data"),
		)

		require.NoError(t, err)
		assert.NotNil(t, flashLoan)
		assert.Equal(t, "user2", flashLoan.UserID)
		assert.Equal(t, "USDC", flashLoan.Asset)
		assert.Equal(t, big.NewInt(5000000), flashLoan.Amount)
		assert.Equal(t, FlashLoanStatusCompleted, flashLoan.Status)
		assert.True(t, flashLoan.IsSuccessful)
		assert.NotZero(t, flashLoan.BorrowTime)
		assert.NotZero(t, flashLoan.RepayTime)
		assert.Equal(t, []byte("arbitrage_data"), flashLoan.CallbackData)

		// Verify fee calculation
		expectedFee := new(big.Int).Mul(big.NewInt(5000000), big.NewInt(50))
		expectedFee.Div(expectedFee, big.NewInt(10000)) // 0.5% = 2500 (0.0025 USDC)
		assert.Equal(t, expectedFee, flashLoan.Fee)
	})

	t.Run("execute flash loan with callback failure", func(t *testing.T) {
		lendingPool := createTestPool()
		// Supply some assets to the pool
		err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
		require.NoError(t, err)

		manager := NewFlashLoanManager(
			lendingPool,
			big.NewInt(1000000000), // 1000 USDC max
			big.NewInt(1000000),    // 1 USDC min
			big.NewInt(50),         // 0.5% fee
			30*time.Second,
		)

		// Create a failing callback
		callback := func(flashLoan *FlashLoan) error {
			return assert.AnError
		}

		flashLoan, err := manager.ExecuteFlashLoan(
			"user3",
			"USDC",
			big.NewInt(3000000), // 3 USDC
			callback,
			nil,
		)

		require.Error(t, err)
		assert.NotNil(t, flashLoan)
		assert.Equal(t, FlashLoanStatusFailed, flashLoan.Status)
		assert.False(t, flashLoan.IsSuccessful)
		assert.Contains(t, flashLoan.ErrorMessage, "flash loan callback execution failed")
		assert.Contains(t, err.Error(), "flash loan callback execution failed")
	})

	t.Run("execute flash loan with insufficient liquidity", func(t *testing.T) {
		lendingPool := createTestPool()
		// Supply some assets to the pool
		err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
		require.NoError(t, err)

		manager := NewFlashLoanManager(
			lendingPool,
			big.NewInt(1000000000), // 1000 USDC max
			big.NewInt(1000000),    // 1 USDC min
			big.NewInt(50),         // 0.5% fee
			30*time.Second,
		)

		callback := func(flashLoan *FlashLoan) error {
			return nil
		}

		flashLoan, err := manager.ExecuteFlashLoan(
			"user4",
			"USDC",
			big.NewInt(15000000), // 15 USDC (more than available)
			callback,
			nil,
		)

		require.Error(t, err)
		assert.Nil(t, flashLoan)
		assert.Contains(t, err.Error(), "insufficient liquidity for flash loan")
	})

	t.Run("execute flash loan with inactive pool", func(t *testing.T) {
		lendingPool := createTestPool()
		lendingPool.IsActive = false

		manager := NewFlashLoanManager(
			lendingPool,
			big.NewInt(1000000000), // 1000 USDC max
			big.NewInt(1000000),    // 1 USDC min
			big.NewInt(50),         // 0.5% fee
			30*time.Second,
		)

		callback := func(flashLoan *FlashLoan) error {
			return nil
		}

		flashLoan, err := manager.ExecuteFlashLoan(
			"user5",
			"USDC",
			big.NewInt(5000000), // 5 USDC
			callback,
			nil,
		)

		require.Error(t, err)
		assert.Nil(t, flashLoan)
		assert.Contains(t, err.Error(), "flash loan not supported for this asset")
	})
}

func TestFlashLoanManagerValidation(t *testing.T) {
	manager := NewFlashLoanManager(
		createTestPool(),
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("validate empty user ID", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("", "USDC", big.NewInt(5000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("validate empty asset", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "", big.NewInt(5000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "asset cannot be empty")
	})

	t.Run("validate nil amount", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("validate zero amount", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("validate negative amount", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(-1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("validate amount below minimum", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(500000)) // 0.5 USDC
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flash loan amount below minimum required")
	})

	t.Run("validate amount above maximum", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(2000000000)) // 2000 USDC
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flash loan amount exceeds maximum allowed")
	})

	t.Run("validate valid amount", func(t *testing.T) {
		err := manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(5000000)) // 5 USDC
		assert.NoError(t, err)
	})
}

func TestFlashLoanManagerFeeCalculation(t *testing.T) {
	manager := NewFlashLoanManager(
		createTestPool(),
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(100),        // 1% fee
		30*time.Second,
	)

	t.Run("calculate fee for 1% rate", func(t *testing.T) {
		amount := big.NewInt(10000000) // 10 USDC
		fee := manager.calculateFlashLoanFee(amount)

		// Expected: 10 USDC * 1% = 0.1 USDC = 100000
		expectedFee := big.NewInt(100000)
		assert.Equal(t, expectedFee, fee)
	})

	t.Run("calculate fee for 0.5% rate", func(t *testing.T) {
		manager.flashLoanFeeRate = big.NewInt(50) // 0.5%
		amount := big.NewInt(20000000)            // 20 USDC
		fee := manager.calculateFlashLoanFee(amount)

		// Expected: 20 USDC * 0.5% = 0.1 USDC = 100000
		expectedFee := big.NewInt(100000)
		assert.Equal(t, expectedFee, fee)
	})

	t.Run("calculate fee for zero amount", func(t *testing.T) {
		amount := big.NewInt(0)
		fee := manager.calculateFlashLoanFee(amount)

		expectedFee := big.NewInt(0)
		assert.Equal(t, expectedFee, fee)
	})
}

func TestFlashLoanManagerLoanManagement(t *testing.T) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
	require.NoError(t, err)

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("get flash loan by ID", func(t *testing.T) {
		// Execute a flash loan first
		callback := func(flashLoan *FlashLoan) error {
			return nil
		}

		flashLoan, err := manager.ExecuteFlashLoan(
			"user2",
			"USDC",
			big.NewInt(3000000), // 3 USDC
			callback,
			nil,
		)
		require.NoError(t, err)

		// Get the loan by ID
		retrievedLoan, err := manager.GetFlashLoan(flashLoan.ID)
		require.NoError(t, err)
		assert.Equal(t, flashLoan.ID, retrievedLoan.ID)
		assert.Equal(t, flashLoan.UserID, retrievedLoan.UserID)
		assert.Equal(t, flashLoan.Asset, retrievedLoan.Asset)
	})

	t.Run("get non-existent flash loan", func(t *testing.T) {
		loan, err := manager.GetFlashLoan("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, loan)
		assert.Contains(t, err.Error(), "flash loan not found")
	})

	t.Run("get user flash loans", func(t *testing.T) {
		// Execute another flash loan for the same user
		callback := func(flashLoan *FlashLoan) error {
			return nil
		}

		_, err := manager.ExecuteFlashLoan(
			"user3",
			"USDC",
			big.NewInt(2000000), // 2 USDC
			callback,
			nil,
		)
		require.NoError(t, err)

		userLoans := manager.GetUserFlashLoans("user3")
		assert.Len(t, userLoans, 1)
		assert.Equal(t, "user3", userLoans[0].UserID)
	})

	t.Run("get active flash loans", func(t *testing.T) {
		activeLoans := manager.GetActiveFlashLoans()
		// All loans should be completed, so no active ones
		assert.Len(t, activeLoans, 0)
	})
}

func TestFlashLoanManagerCancellation(t *testing.T) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
	require.NoError(t, err)

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("cancel pending flash loan", func(t *testing.T) {
		// Create a flash loan manually to test cancellation
		flashLoan := manager.createFlashLoan("user4", "USDC", big.NewInt(1000000), nil)
		flashLoan.Status = FlashLoanStatusPending

		err := manager.CancelFlashLoan(flashLoan.ID, "user4")
		assert.NoError(t, err)

		// Verify the loan was cancelled
		retrievedLoan, err := manager.GetFlashLoan(flashLoan.ID)
		require.NoError(t, err)
		assert.Equal(t, FlashLoanStatusFailed, retrievedLoan.Status)
		assert.Contains(t, retrievedLoan.ErrorMessage, "cancelled by user")
		assert.False(t, retrievedLoan.IsSuccessful)
	})

	t.Run("cancel non-existent flash loan", func(t *testing.T) {
		err := manager.CancelFlashLoan("nonexistent", "user5")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flash loan not found")
	})

	t.Run("cancel flash loan with wrong user", func(t *testing.T) {
		// Create a flash loan for user6
		flashLoan := manager.createFlashLoan("user6", "USDC", big.NewInt(1000000), nil)
		flashLoan.Status = FlashLoanStatusPending

		// Try to cancel with wrong user
		err := manager.CancelFlashLoan(flashLoan.ID, "user7")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not authorized to cancel this flash loan")
	})

	t.Run("cancel completed flash loan", func(t *testing.T) {
		// Create a flash loan and mark it as completed
		flashLoan := manager.createFlashLoan("user8", "USDC", big.NewInt(1000000), nil)
		flashLoan.Status = FlashLoanStatusCompleted

		err := manager.CancelFlashLoan(flashLoan.ID, "user8")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid flash loan status for this operation")
	})
}

func TestFlashLoanManagerExpiration(t *testing.T) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
	require.NoError(t, err)

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		1*time.Millisecond,     // Very short duration for testing
	)

	t.Run("cleanup expired loans", func(t *testing.T) {
		// Create a flash loan and mark it as active
		flashLoan := manager.createFlashLoan("user9", "USDC", big.NewInt(1000000), nil)
		flashLoan.Status = FlashLoanStatusActive
		flashLoan.BorrowTime = time.Now().Add(-2 * time.Millisecond) // Expired

		// Cleanup expired loans
		manager.CleanupExpiredLoans()

		// Verify the loan was marked as expired
		retrievedLoan, err := manager.GetFlashLoan(flashLoan.ID)
		require.NoError(t, err)
		assert.Equal(t, FlashLoanStatusExpired, retrievedLoan.Status)
		assert.Contains(t, retrievedLoan.ErrorMessage, "flash loan expired")
		assert.False(t, retrievedLoan.IsSuccessful)
	})
}

func TestFlashLoanManagerStatistics(t *testing.T) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
	require.NoError(t, err)

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("get flash loan stats", func(t *testing.T) {
		// Execute a successful flash loan
		callback := func(flashLoan *FlashLoan) error {
			return nil
		}

		_, err := manager.ExecuteFlashLoan(
			"user10",
			"USDC",
			big.NewInt(4000000), // 4 USDC
			callback,
			nil,
		)
		require.NoError(t, err)

		// Create a failed flash loan
		failedCallback := func(flashLoan *FlashLoan) error {
			return assert.AnError
		}

		_, err = manager.ExecuteFlashLoan(
			"user11",
			"USDC",
			big.NewInt(2000000), // 2 USDC
			failedCallback,
			nil,
		)
		require.Error(t, err)

		// Get statistics
		stats := manager.GetFlashLoanStats()

		assert.Equal(t, 2, stats["total_loans"])
		assert.Equal(t, 0, stats["active_loans"])
		assert.Equal(t, 1, stats["completed_loans"])
		assert.Equal(t, 1, stats["failed_loans"])
		assert.Equal(t, "4000000", stats["total_volume"])
		assert.Equal(t, "20000", stats["total_fees"]) // 4 USDC * 0.5% = 20000 (0.02 USDC)
		assert.Equal(t, "50", stats["fee_rate"])
		assert.Equal(t, "1000000000", stats["max_amount"])
		assert.Equal(t, "1000000", stats["min_amount"])
		assert.Equal(t, "30s", stats["max_duration"])
	})
}

func TestFlashLoanManagerSettingsUpdate(t *testing.T) {
	manager := NewFlashLoanManager(
		createTestPool(),
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("update flash loan settings", func(t *testing.T) {
		newMaxAmount := big.NewInt(2000000000) // 2000 USDC
		newMinAmount := big.NewInt(2000000)    // 2 USDC
		newFeeRate := big.NewInt(100)          // 1%
		newDuration := 60 * time.Second

		err := manager.UpdateFlashLoanSettings(newMaxAmount, newMinAmount, newFeeRate, newDuration)
		assert.NoError(t, err)

		assert.Equal(t, newMaxAmount, manager.maxFlashLoanAmount)
		assert.Equal(t, newMinAmount, manager.minFlashLoanAmount)
		assert.Equal(t, newFeeRate, manager.flashLoanFeeRate)
		assert.Equal(t, newDuration, manager.maxFlashLoanDuration)
	})

	t.Run("update with nil values", func(t *testing.T) {
		originalMaxAmount := manager.maxFlashLoanAmount
		originalMinAmount := manager.minFlashLoanAmount
		originalFeeRate := manager.flashLoanFeeRate
		originalDuration := manager.maxFlashLoanDuration

		err := manager.UpdateFlashLoanSettings(nil, nil, nil, 0)
		assert.NoError(t, err)

		// Values should remain unchanged
		assert.Equal(t, originalMaxAmount, manager.maxFlashLoanAmount)
		assert.Equal(t, originalMinAmount, manager.minFlashLoanAmount)
		assert.Equal(t, originalFeeRate, manager.flashLoanFeeRate)
		assert.Equal(t, originalDuration, manager.maxFlashLoanDuration)
	})
}

func TestFlashLoanManagerConcurrency(t *testing.T) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
	require.NoError(t, err)

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	t.Run("concurrent flash loan execution", func(t *testing.T) {
		const numGoroutines = 5
		const numLoans = 10

		var wg sync.WaitGroup

		// Start goroutines to execute flash loans concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				userID := "user" + string(rune(id))

				for j := 0; j < numLoans; j++ {
					callback := func(flashLoan *FlashLoan) error {
						return nil
					}

					_, err := manager.ExecuteFlashLoan(
						userID,
						"USDC",
						big.NewInt(1000000), // 1 USDC
						callback,
						nil,
					)
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify no data races occurred
		stats := manager.GetFlashLoanStats()
		totalLoans := stats["total_loans"].(int)
		assert.Equal(t, 50, totalLoans) // 5 * 10 = 50
	})
}

// Benchmark tests for performance validation
func BenchmarkFlashLoanExecution(b *testing.B) {
	lendingPool := createTestPool()
	// Supply some assets to the pool
	lendingPool.Supply("benchmark_user", big.NewInt(10000000)) // 10 USDC

	manager := NewFlashLoanManager(
		lendingPool,
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	callback := func(flashLoan *FlashLoan) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ExecuteFlashLoan(
			"benchmark_user",
			"USDC",
			big.NewInt(1000000), // 1 USDC
			callback,
			nil,
		)
	}
}

func BenchmarkFlashLoanFeeCalculation(b *testing.B) {
	manager := NewFlashLoanManager(
		createTestPool(),
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	amount := big.NewInt(10000000) // 10 USDC

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.calculateFlashLoanFee(amount)
	}
}

func BenchmarkFlashLoanValidation(b *testing.B) {
	manager := NewFlashLoanManager(
		createTestPool(),
		big.NewInt(1000000000), // 1000 USDC max
		big.NewInt(1000000),    // 1 USDC min
		big.NewInt(50),         // 0.5% fee
		30*time.Second,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.validateFlashLoanRequest("user1", "USDC", big.NewInt(5000000))
	}
}
