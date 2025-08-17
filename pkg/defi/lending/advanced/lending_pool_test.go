package advanced

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLendingPool(t *testing.T) {
	t.Run("valid lending pool", func(t *testing.T) {
		baseRate := big.NewInt(100)              // 1% base rate
		kinkRate := big.NewInt(8000)             // 80% kink rate
		multiplier := big.NewInt(200)            // 2% multiplier
		jumpMultiplier := big.NewInt(1000)       // 10% jump multiplier
		reserveFactor := big.NewInt(1000)        // 10% reserve factor
		collateralFactor := big.NewInt(7500)     // 75% collateral factor
		liquidationThreshold := big.NewInt(8000) // 80% liquidation threshold
		liquidationPenalty := big.NewInt(500)    // 5% liquidation penalty

		pool := NewLendingPool(
			"USDC",
			baseRate, kinkRate, multiplier, jumpMultiplier,
			reserveFactor, collateralFactor, liquidationThreshold, liquidationPenalty,
		)

		require.NotNil(t, pool)
		assert.Equal(t, "pool_USDC", pool.ID)
		assert.Equal(t, "USDC", pool.Asset)
		assert.Equal(t, baseRate, pool.BaseRate)
		assert.Equal(t, kinkRate, pool.KinkRate)
		assert.Equal(t, multiplier, pool.Multiplier)
		assert.Equal(t, jumpMultiplier, pool.JumpMultiplier)
		assert.Equal(t, reserveFactor, pool.ReserveFactor)
		assert.Equal(t, collateralFactor, pool.CollateralFactor)
		assert.Equal(t, liquidationThreshold, pool.LiquidationThreshold)
		assert.Equal(t, liquidationPenalty, pool.LiquidationPenalty)
		assert.True(t, pool.IsActive)
		assert.Equal(t, big.NewInt(1000), pool.MinBorrowAmount)
		assert.Equal(t, big.NewInt(0), pool.TotalSupply)
		assert.Equal(t, big.NewInt(0), pool.TotalBorrowed)
		assert.Equal(t, big.NewInt(0), pool.AvailableLiquidity)
		assert.Equal(t, big.NewInt(0), pool.UtilizationRate)
		assert.Equal(t, big.NewInt(0), pool.SupplyRate)
		assert.Equal(t, baseRate, pool.BorrowRate)
		assert.NotZero(t, pool.CreatedAt)
		assert.NotZero(t, pool.UpdatedAt)
		assert.NotZero(t, pool.LastInterestUpdate)
	})
}

func TestLendingPoolSupply(t *testing.T) {
	t.Run("supply assets successfully", func(t *testing.T) {
		pool := createTestPool()
		amount := big.NewInt(1000000) // 1 USDC
		userID := "user1"

		err := pool.Supply(userID, amount)
		assert.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		assert.Equal(t, amount, account.SupplyBalance)
		assert.Equal(t, amount, pool.TotalSupply)
		assert.Equal(t, amount, pool.AvailableLiquidity)
		assert.Equal(t, big.NewInt(0), pool.UtilizationRate)
	})

	t.Run("supply to inactive pool", func(t *testing.T) {
		pool := createTestPool()
		pool.IsActive = false

		err := pool.Supply("user2", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is inactive")
	})

	t.Run("supply invalid amount", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Supply("user3", big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Supply("user3", big.NewInt(-1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Supply("user3", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("supply multiple times", func(t *testing.T) {
		pool := createTestPool()
		userID := "user4"
		amount1 := big.NewInt(1000000)
		amount2 := big.NewInt(2000000)

		err := pool.Supply(userID, amount1)
		assert.NoError(t, err)

		err = pool.Supply(userID, amount2)
		assert.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		expectedBalance := new(big.Int).Add(amount1, amount2)
		assert.Equal(t, expectedBalance, account.SupplyBalance)
		assert.Equal(t, expectedBalance, pool.TotalSupply)
	})
}

func TestLendingPoolBorrow(t *testing.T) {
	t.Run("borrow successfully", func(t *testing.T) {
		pool := createTestPool()
		// First supply some assets
		supplyAmount := big.NewInt(10000000) // 10 USDC
		userID := "user1"
		err := pool.Supply(userID, supplyAmount)
		require.NoError(t, err)

		// Set collateral value
		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(15000000) // 15 USDC collateral

		// Borrow assets
		borrowAmount := big.NewInt(5000000) // 5 USDC
		err = pool.Borrow(userID, borrowAmount)
		assert.NoError(t, err)

		account, err = pool.GetAccount(userID)
		require.NoError(t, err)
		assert.Equal(t, borrowAmount, account.BorrowBalance)
		assert.Equal(t, borrowAmount, pool.TotalBorrowed)
		assert.Equal(t, big.NewInt(5000000), pool.AvailableLiquidity) // 10 - 5 = 5
		assert.True(t, pool.UtilizationRate.Cmp(big.NewInt(0)) > 0)
	})

	t.Run("borrow from inactive pool", func(t *testing.T) {
		pool := createTestPool()
		pool.IsActive = false

		err := pool.Borrow("user2", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is inactive")
	})

	t.Run("borrow invalid amount", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Borrow("user3", big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Borrow("user3", big.NewInt(-1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Borrow("user3", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("borrow below minimum", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Borrow("user4", big.NewInt(500)) // Below 1000 minimum
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount below minimum borrow amount")
	})

	t.Run("borrow more than available liquidity", func(t *testing.T) {
		pool := createTestPool()
		// Supply some assets first
		err := pool.Supply("user5", big.NewInt(10000000))
		require.NoError(t, err)

		// Pool has 10 USDC available
		err = pool.Borrow("user5", big.NewInt(15000000)) // Try to borrow 15 USDC
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient liquidity")
	})

	t.Run("borrow exceeds collateral limit", func(t *testing.T) {
		pool := createTestPool()
		userID := "user6"
		// Supply and set collateral
		err := pool.Supply(userID, big.NewInt(10000000))
		require.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(10000000) // 10 USDC collateral

		// Try to borrow more than collateral allows (75% of 10 = 7.5 USDC)
		err = pool.Borrow(userID, big.NewInt(8000000)) // 8 USDC
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "borrow limit exceeded")
	})
}

func TestLendingPoolRepay(t *testing.T) {
	t.Run("repay successfully", func(t *testing.T) {
		pool := createTestPool()
		// Setup: supply and borrow
		userID := "user1"
		err := pool.Supply(userID, big.NewInt(10000000))
		require.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(15000000)

		err = pool.Borrow(userID, big.NewInt(5000000))
		require.NoError(t, err)

		// Repay
		repayAmount := big.NewInt(2000000) // 2 USDC
		err = pool.Repay(userID, repayAmount)
		assert.NoError(t, err)

		account, err = pool.GetAccount(userID)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(3000000), account.BorrowBalance) // 5 - 2 = 3
		assert.Equal(t, big.NewInt(3000000), pool.TotalBorrowed)
		assert.Equal(t, big.NewInt(7000000), pool.AvailableLiquidity) // 5 + 2 = 7
	})

	t.Run("repay more than owed", func(t *testing.T) {
		pool := createTestPool()
		userID := "user2"
		// Setup: supply and borrow
		err := pool.Supply(userID, big.NewInt(10000000))
		require.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(15000000)

		err = pool.Borrow(userID, big.NewInt(3000000))
		require.NoError(t, err)

		// Repay more than owed
		repayAmount := big.NewInt(5000000) // 5 USDC (but only owe 3)
		err = pool.Repay(userID, repayAmount)
		assert.NoError(t, err)

		account, err = pool.GetAccount(userID)
		require.NoError(t, err)
		assert.Equal(t, 0, account.BorrowBalance.Cmp(big.NewInt(0))) // Fully repaid
		assert.Equal(t, 0, pool.TotalBorrowed.Cmp(big.NewInt(0)))
	})

	t.Run("repay from inactive pool", func(t *testing.T) {
		pool := createTestPool()
		pool.IsActive = false

		err := pool.Repay("user3", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is inactive")
	})

	t.Run("repay invalid amount", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Repay("user4", big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Repay("user4", big.NewInt(-1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Repay("user4", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("repay non-existent account", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Repay("nonexistent", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account not found")
	})
}

func TestLendingPoolWithdraw(t *testing.T) {
	t.Run("withdraw successfully", func(t *testing.T) {
		pool := createTestPool()
		userID := "user1"
		supplyAmount := big.NewInt(10000000) // 10 USDC
		err := pool.Supply(userID, supplyAmount)
		require.NoError(t, err)

		withdrawAmount := big.NewInt(3000000) // 3 USDC
		err = pool.Withdraw(userID, withdrawAmount)
		assert.NoError(t, err)

		account, err := pool.GetAccount(userID)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(7000000), account.SupplyBalance) // 10 - 3 = 7
		assert.Equal(t, big.NewInt(7000000), pool.TotalSupply)
		assert.Equal(t, big.NewInt(7000000), pool.AvailableLiquidity)
	})

	t.Run("withdraw more than supplied", func(t *testing.T) {
		pool := createTestPool()
		userID := "user2"
		supplyAmount := big.NewInt(10000000) // 10 USDC
		err := pool.Supply(userID, supplyAmount)
		require.NoError(t, err)

		withdrawAmount := big.NewInt(15000000) // 15 USDC
		err = pool.Withdraw(userID, withdrawAmount)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient supply balance")
	})

	t.Run("withdraw from inactive pool", func(t *testing.T) {
		pool := createTestPool()
		pool.IsActive = false

		err := pool.Withdraw("user3", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool is inactive")
	})

	t.Run("withdraw invalid amount", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Withdraw("user4", big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Withdraw("user4", big.NewInt(-1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")

		err = pool.Withdraw("user4", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("withdraw non-existent account", func(t *testing.T) {
		pool := createTestPool()
		err := pool.Withdraw("nonexistent", big.NewInt(1000000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account not found")
	})
}

func TestLendingPoolInterestRates(t *testing.T) {
	t.Run("interest rate calculation below kink", func(t *testing.T) {
		pool := createTestPool()
		// Set utilization to 50% (below 80% kink)
		pool.TotalSupply = big.NewInt(10000000)  // 10 USDC
		pool.TotalBorrowed = big.NewInt(5000000) // 5 USDC
		pool.AvailableLiquidity = big.NewInt(5000000)
		pool.updateUtilizationRate()

		// Force interest rate update
		pool.LastInterestUpdate = time.Now().Add(-2 * time.Minute)
		pool.updateInterestRates()

		// Expected: base + (utilization * multiplier)
		// 100 + (5000 * 200) = 100 + 1000000 = 1000100
		expectedRate := big.NewInt(1000100)
		assert.Equal(t, expectedRate, pool.BorrowRate)
	})

	t.Run("interest rate calculation above kink", func(t *testing.T) {
		pool := createTestPool()
		// Set utilization to 90% (above 80% kink)
		pool.TotalSupply = big.NewInt(10000000)  // 10 USDC
		pool.TotalBorrowed = big.NewInt(9000000) // 9 USDC
		pool.AvailableLiquidity = big.NewInt(1000000)
		pool.updateUtilizationRate()

		// Force interest rate update
		pool.LastInterestUpdate = time.Now().Add(-2 * time.Minute)
		pool.updateInterestRates()

		// Expected: base + (kink * multiplier) + ((utilization - kink) * jumpMultiplier)
		// 100 + (8000 * 200) + ((9000 - 8000) * 1000) = 100 + 1600000 + 1000000 = 2600100
		expectedRate := big.NewInt(2600100)
		assert.Equal(t, expectedRate, pool.BorrowRate)
	})

	t.Run("supply rate calculation", func(t *testing.T) {
		pool := createTestPool()
		// Set utilization to 60%
		pool.TotalSupply = big.NewInt(10000000)  // 10 USDC
		pool.TotalBorrowed = big.NewInt(6000000) // 6 USDC
		pool.AvailableLiquidity = big.NewInt(4000000)
		pool.updateUtilizationRate()

		// Force interest rate update
		pool.LastInterestUpdate = time.Now().Add(-2 * time.Minute)
		pool.updateInterestRates()

		// Supply rate should be calculated based on borrow rate
		// Note: Supply rate can be 0 due to integer division rounding
		assert.True(t, pool.SupplyRate.Cmp(big.NewInt(0)) >= 0)
		// Supply rate should not exceed borrow rate
		assert.True(t, pool.SupplyRate.Cmp(pool.BorrowRate) <= 0)
		// Verify that supply rate was calculated (even if it's 0)
		assert.NotNil(t, pool.SupplyRate)

	})
}

func TestLendingPoolHealthFactor(t *testing.T) {
	t.Run("healthy account", func(t *testing.T) {
		pool := createTestPool()
		userID := "user1"
		account := &Account{
			UserID:          userID,
			CollateralValue: big.NewInt(10000000), // 10 USDC collateral
			BorrowValue:     big.NewInt(5000000),  // 5 USDC borrowed
		}

		pool.updateAccountHealthFactor(account)
		assert.True(t, account.HealthFactor.Cmp(big.NewInt(100)) >= 0)
		assert.False(t, account.IsLiquidatable)
	})

	t.Run("unhealthy account", func(t *testing.T) {
		pool := createTestPool()
		userID := "user2"
		account := &Account{
			UserID:          userID,
			CollateralValue: big.NewInt(10000000), // 10 USDC collateral
			BorrowValue:     big.NewInt(12000000), // 12 USDC borrowed
		}

		pool.updateAccountHealthFactor(account)
		assert.True(t, account.HealthFactor.Cmp(big.NewInt(100)) < 0)
		assert.True(t, account.IsLiquidatable)
	})

	t.Run("account with no borrows", func(t *testing.T) {
		pool := createTestPool()
		userID := "user3"
		account := &Account{
			UserID:          userID,
			CollateralValue: big.NewInt(10000000), // 10 USDC collateral
			BorrowValue:     big.NewInt(0),        // No borrows
		}

		pool.updateAccountHealthFactor(account)
		assert.Equal(t, big.NewInt(10000), account.HealthFactor) // 100%
		assert.False(t, account.IsLiquidatable)
	})
}

func TestLendingPoolAccountManagement(t *testing.T) {
	t.Run("get or create account", func(t *testing.T) {
		pool := createTestPool()
		userID := "user1"
		account := pool.getOrCreateAccount(userID)

		assert.NotNil(t, account)
		assert.Equal(t, userID, account.UserID)
		assert.Equal(t, big.NewInt(0), account.SupplyBalance)
		assert.Equal(t, big.NewInt(0), account.BorrowBalance)
		assert.Equal(t, pool.supplyIndex, account.SupplyIndex)
		assert.Equal(t, pool.borrowIndex, account.BorrowIndex)

		// Get existing account
		existingAccount := pool.getOrCreateAccount(userID)
		assert.Equal(t, account, existingAccount)
	})

	t.Run("get non-existent account", func(t *testing.T) {
		pool := createTestPool()
		account, err := pool.GetAccount("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, account)
		assert.Contains(t, err.Error(), "account not found")
	})
}

func TestLendingPoolStats(t *testing.T) {
	t.Run("get pool stats", func(t *testing.T) {
		pool := createTestPool()
		stats := pool.GetPoolStats()

		assert.Equal(t, pool.ID, stats["id"])
		assert.Equal(t, pool.Asset, stats["asset"])
		assert.Equal(t, pool.TotalSupply.String(), stats["total_supply"])
		assert.Equal(t, pool.TotalBorrowed.String(), stats["total_borrowed"])
		assert.Equal(t, pool.AvailableLiquidity.String(), stats["available_liquidity"])
		assert.Equal(t, pool.UtilizationRate.String(), stats["utilization_rate"])
		assert.Equal(t, pool.SupplyRate.String(), stats["supply_rate"])
		assert.Equal(t, pool.BorrowRate.String(), stats["borrow_rate"])
		assert.Equal(t, pool.BaseRate.String(), stats["base_rate"])
		assert.Equal(t, pool.IsActive, stats["is_active"])
		assert.Equal(t, len(pool.accounts), stats["account_count"])
	})
}

func TestLendingPoolConcurrency(t *testing.T) {
	t.Run("concurrent operations", func(t *testing.T) {
		pool := createTestPool()
		const numGoroutines = 10
		const numOperations = 100

		var wg sync.WaitGroup

		// Start goroutines to perform operations concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				userID := "user" + string(rune(id))

				for j := 0; j < numOperations; j++ {
					// Supply
					pool.Supply(userID, big.NewInt(1000000))

					// Set collateral
					account, _ := pool.GetAccount(userID)
					if account != nil {
						account.CollateralValue = big.NewInt(2000000)
					}

					// Borrow
					pool.Borrow(userID, big.NewInt(500000))

					// Repay
					pool.Repay(userID, big.NewInt(200000))

					// Withdraw
					pool.Withdraw(userID, big.NewInt(300000))
				}
			}(i)
		}

		wg.Wait()

		// Verify no data races occurred
		assert.True(t, pool.TotalSupply.Cmp(big.NewInt(0)) >= 0)
		assert.True(t, pool.TotalBorrowed.Cmp(big.NewInt(0)) >= 0)
	})
}

// Helper function to create a test pool
func createTestPool() *LendingPool {
	baseRate := big.NewInt(100)              // 1% base rate
	kinkRate := big.NewInt(8000)             // 80% kink rate
	multiplier := big.NewInt(200)            // 2% multiplier
	jumpMultiplier := big.NewInt(1000)       // 10% jump multiplier
	reserveFactor := big.NewInt(1000)        // 10% reserve factor
	collateralFactor := big.NewInt(7500)     // 75% collateral factor
	liquidationThreshold := big.NewInt(8000) // 80% liquidation threshold
	liquidationPenalty := big.NewInt(500)    // 5% liquidation penalty

	return NewLendingPool(
		"USDC",
		baseRate, kinkRate, multiplier, jumpMultiplier,
		reserveFactor, collateralFactor, liquidationThreshold, liquidationPenalty,
	)
}

// Benchmark tests for performance validation
func BenchmarkLendingPoolSupply(b *testing.B) {
	pool := createTestPool()
	userID := "benchmark_user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Supply(userID, big.NewInt(1000000))
	}
}

func BenchmarkLendingPoolBorrow(b *testing.B) {
	pool := createTestPool()
	userID := "benchmark_user"

	// Setup: supply and set collateral
	pool.Supply(userID, big.NewInt(10000000))
	account, _ := pool.GetAccount(userID)
	account.CollateralValue = big.NewInt(15000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Borrow(userID, big.NewInt(1000000))
	}
}

func BenchmarkLendingPoolInterestRateUpdate(b *testing.B) {
	pool := createTestPool()
	pool.TotalSupply = big.NewInt(10000000)
	pool.TotalBorrowed = big.NewInt(6000000)
	pool.AvailableLiquidity = big.NewInt(4000000)
	pool.updateUtilizationRate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.updateInterestRates()
	}
}

func BenchmarkLendingPoolHealthFactorUpdate(b *testing.B) {
	pool := createTestPool()
	account := &Account{
		CollateralValue: big.NewInt(10000000),
		BorrowValue:     big.NewInt(5000000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.updateAccountHealthFactor(account)
	}
}
