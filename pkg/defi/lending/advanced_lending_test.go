package lending

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLendingService tests the LendingService functionality
func TestLendingService(t *testing.T) {
	t.Run("NewLendingService", func(t *testing.T) {
		service := NewLendingService()
		require.NotNil(t, service)
		assert.Empty(t, service.pools)
		assert.Empty(t, service.loans)
		assert.Empty(t, service.users)
		assert.NotNil(t, service.logger)
	})

	t.Run("CreatePool", func(t *testing.T) {
		service := NewLendingService()

		assets := []string{"BTC", "ETH", "USDT"}
		pool, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)
		require.NotNil(t, pool)

		assert.Equal(t, "pool_1", pool.ID)
		assert.Equal(t, "Test Pool", pool.Name)
		assert.Len(t, pool.Assets, 3)

		// Check that assets were initialized correctly
		for _, asset := range assets {
			assetData, exists := pool.Assets[asset]
			assert.True(t, exists)
			assert.Equal(t, asset, assetData.Symbol)
			assert.Equal(t, 0.8, assetData.CollateralFactor)
			assert.Equal(t, 0.85, assetData.LiquidationThreshold)
			assert.Equal(t, 0.1, assetData.ReserveFactor)
		}

		// Test duplicate pool creation
		_, err = service.CreatePool("pool_1", "Duplicate Pool", assets)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("SupplyAsset", func(t *testing.T) {
		service := NewLendingService()

		// Create pool first
		assets := []string{"BTC", "ETH"}
		pool, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		// Supply assets
		err = service.SupplyAsset("pool_1", "BTC", "user_1", 100.0)
		require.NoError(t, err)

		// Check pool state
		pool, _ = service.GetPool("pool_1")
		assert.Equal(t, 100.0, pool.TotalLiquidity)
		assert.Equal(t, 0.0, pool.TotalBorrowed)
		assert.Equal(t, 0.0, pool.UtilizationRate)

		// Check asset state
		asset := pool.Assets["BTC"]
		assert.Equal(t, 100.0, asset.TotalSupply)
		assert.Equal(t, 100.0, asset.AvailableLiquidity)

		// Check user state
		user, exists := service.GetUser("user_1")
		assert.True(t, exists)
		assert.Equal(t, 100.0, user.TotalSupplied)
		assert.Equal(t, 100.0, user.Collateral["BTC"])

		// Test invalid pool
		err = service.SupplyAsset("invalid_pool", "BTC", "user_1", 50.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test invalid asset
		err = service.SupplyAsset("pool_1", "INVALID", "user_1", 50.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test invalid amount
		err = service.SupplyAsset("pool_1", "BTC", "user_1", -50.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("BorrowAsset", func(t *testing.T) {
		service := NewLendingService()

		// Create pool and supply assets
		assets := []string{"BTC", "ETH"}
		_, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		err = service.SupplyAsset("pool_1", "BTC", "user_1", 1000.0)
		require.NoError(t, err)

		// Create collateral
		collateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      10.0,
				Value:       2000.0, // 10 ETH * $200
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		// Borrow assets
		loan, err := service.BorrowAsset("pool_1", "BTC", "user_2", 500.0, collateral)
		require.NoError(t, err)
		require.NotNil(t, loan)

		assert.Equal(t, "user_2", loan.UserID)
		assert.Equal(t, "BTC", loan.AssetSymbol)
		assert.Equal(t, 500.0, loan.Amount)
		assert.Equal(t, 500.0, loan.BorrowedAmount)
		assert.Equal(t, LoanStatusActive, loan.Status)
		assert.Len(t, loan.Collateral, 1)

		// Check pool state
		pool, _ := service.GetPool("pool_1")
		assert.Equal(t, 1000.0, pool.TotalLiquidity)
		assert.Equal(t, 500.0, pool.TotalBorrowed)
		assert.Equal(t, 0.5, pool.UtilizationRate)

		// Check asset state
		asset := pool.Assets["BTC"]
		assert.Equal(t, 1000.0, asset.TotalSupply)
		assert.Equal(t, 500.0, asset.TotalBorrowed)
		assert.Equal(t, 500.0, asset.AvailableLiquidity)

		// Check user state
		user, exists := service.GetUser("user_2")
		assert.True(t, exists)
		assert.Equal(t, 500.0, user.TotalBorrowed)
		assert.Len(t, user.Loans, 1)
		assert.Equal(t, loan.ID, user.Loans[0])

		// Test insufficient liquidity
		_, err = service.BorrowAsset("pool_1", "BTC", "user_3", 1000.0, collateral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient liquidity")

		// Test invalid amount
		_, err = service.BorrowAsset("pool_1", "BTC", "user_3", -100.0, collateral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")

		// Test insufficient collateral
		insufficientCollateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      1.0,
				Value:       200.0, // 1 ETH * $200
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}
		_, err = service.BorrowAsset("pool_1", "BTC", "user_3", 500.0, insufficientCollateral)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collateral validation failed")
	})

	t.Run("RepayLoan", func(t *testing.T) {
		service := NewLendingService()

		// Create pool, supply assets, and borrow
		assets := []string{"BTC", "ETH"}
		_, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		err = service.SupplyAsset("pool_1", "BTC", "user_1", 1000.0)
		require.NoError(t, err)

		collateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      10.0,
				Value:       2000.0,
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		loan, err := service.BorrowAsset("pool_1", "BTC", "user_2", 500.0, collateral)
		require.NoError(t, err)

		// Repay loan
		err = service.RepayLoan(loan.ID, 200.0)
		require.NoError(t, err)

		// Check loan state
		updatedLoan, _ := service.GetLoan(loan.ID)
		assert.Equal(t, 200.0, updatedLoan.RepaidAmount)
		assert.Equal(t, LoanStatusActive, updatedLoan.Status) // Not fully repaid yet

		// Check pool state
		pool, _ := service.GetPool("pool_1")
		assert.Equal(t, 300.0, pool.TotalBorrowed)
		assert.Equal(t, 700.0, pool.Assets["BTC"].AvailableLiquidity)

		// Check user state
		user, _ := service.GetUser("user_2")
		assert.Equal(t, 300.0, user.TotalBorrowed)

		// Repay remaining amount
		err = service.RepayLoan(loan.ID, 300.0)
		require.NoError(t, err)

		// Check loan is fully repaid
		updatedLoan, _ = service.GetLoan(loan.ID)
		assert.Equal(t, LoanStatusRepaid, updatedLoan.Status)

		// Test invalid loan ID
		err = service.RepayLoan("invalid_loan", 100.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test invalid amount on a new loan
		loan2, err := service.BorrowAsset("pool_1", "BTC", "user_3", 300.0, collateral)
		require.NoError(t, err)
		
		err = service.RepayLoan(loan2.ID, -100.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("LiquidateLoan", func(t *testing.T) {
		service := NewLendingService()

		// Create pool and loan
		assets := []string{"BTC", "ETH"}
		_, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		err = service.SupplyAsset("pool_1", "BTC", "user_1", 1000.0)
		require.NoError(t, err)

		// Create loan with collateral that will become underwater
		collateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      10.0,
				Value:       1000.0, // 10 ETH * $100 (low value)
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		loan, err := service.BorrowAsset("pool_1", "BTC", "user_2", 800.0, collateral)
		require.NoError(t, err)

		// Manually make loan underwater by reducing collateral value
		loan.Collateral[0].Value = 500.0 // Collateral value drops

		// Liquidate loan
		err = service.LiquidateLoan(loan.ID)
		require.NoError(t, err)

		// Check loan status
		updatedLoan, _ := service.GetLoan(loan.ID)
		assert.Equal(t, LoanStatusLiquidated, updatedLoan.Status)

		// Check pool state
		pool, _ := service.GetPool("pool_1")
		assert.Equal(t, 0.0, pool.TotalBorrowed)
		assert.Equal(t, 1000.0, pool.Assets["BTC"].AvailableLiquidity)

		// Check user state
		user, _ := service.GetUser("user_2")
		assert.Equal(t, 0.0, user.TotalBorrowed)

		// Test liquidating non-underwater loan
		// Create new loan with sufficient collateral
		goodCollateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      20.0,
				Value:       4000.0, // 20 ETH * $200
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		goodLoan, err := service.BorrowAsset("pool_1", "BTC", "user_3", 500.0, goodCollateral)
		require.NoError(t, err)

		// Try to liquidate (should fail)
		err = service.LiquidateLoan(goodLoan.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not underwater")
	})

	t.Run("GetMethods", func(t *testing.T) {
		service := NewLendingService()

		// Create test data
		assets := []string{"BTC", "ETH"}
		pool, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		err = service.SupplyAsset("pool_1", "BTC", "user_1", 1000.0)
		require.NoError(t, err)

		collateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      10.0,
				Value:       2000.0,
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		loan, err := service.BorrowAsset("pool_1", "BTC", "user_2", 500.0, collateral)
		require.NoError(t, err)

		// Test GetPool
		retrievedPool, exists := service.GetPool("pool_1")
		assert.True(t, exists)
		assert.Equal(t, pool.ID, retrievedPool.ID)

		// Test GetUser
		user, exists := service.GetUser("user_1")
		assert.True(t, exists)
		assert.Equal(t, "user_1", user.ID)

		// Test GetLoan
		retrievedLoan, exists := service.GetLoan(loan.ID)
		assert.True(t, exists)
		assert.Equal(t, loan.ID, retrievedLoan.ID)

		// Test GetPools
		pools := service.GetPools()
		assert.Len(t, pools, 1)
		assert.Equal(t, "pool_1", pools[0].ID)

		// Test GetUserLoans
		userLoans := service.GetUserLoans("user_2")
		assert.Len(t, userLoans, 1)
		assert.Equal(t, loan.ID, userLoans[0].ID)

		// Test non-existent items
		_, exists = service.GetPool("invalid_pool")
		assert.False(t, exists)

		_, exists = service.GetUser("invalid_user")
		assert.False(t, exists)

		_, exists = service.GetLoan("invalid_loan")
		assert.False(t, exists)
	})
}

// TestLendingPool tests the LendingPool functionality
func TestLendingPool(t *testing.T) {
	t.Run("PoolInitialization", func(t *testing.T) {
		service := NewLendingService()

		assets := []string{"BTC", "ETH", "USDT"}
		pool, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		assert.Equal(t, "pool_1", pool.ID)
		assert.Equal(t, "Test Pool", pool.Name)
		assert.Len(t, pool.Assets, 3)
		assert.Equal(t, 0.0, pool.TotalLiquidity)
		assert.Equal(t, 0.0, pool.TotalBorrowed)
		assert.Equal(t, 0.0, pool.UtilizationRate)
		assert.Equal(t, 0.0, pool.APY)
		assert.NotNil(t, pool.CreatedAt)
		assert.NotNil(t, pool.UpdatedAt)
		assert.NotNil(t, pool.logger)
	})
}

// TestLendingAsset tests the LendingAsset functionality
func TestLendingAsset(t *testing.T) {
	t.Run("AssetInitialization", func(t *testing.T) {
		service := NewLendingService()

		assets := []string{"BTC"}
		pool, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		asset := pool.Assets["BTC"]
		assert.Equal(t, "BTC", asset.Symbol)
		assert.Equal(t, 0.0, asset.TotalSupply)
		assert.Equal(t, 0.0, asset.TotalBorrowed)
		assert.Equal(t, 0.0, asset.AvailableLiquidity)
		assert.Equal(t, 0.8, asset.CollateralFactor)
		assert.Equal(t, 0.85, asset.LiquidationThreshold)
		assert.Equal(t, 0.1, asset.ReserveFactor)
		assert.NotNil(t, asset.LastUpdate)
	})
}

// TestLoan tests the Loan functionality
func TestLoan(t *testing.T) {
	t.Run("LoanCreation", func(t *testing.T) {
		service := NewLendingService()

		// Setup
		assets := []string{"BTC", "ETH"}
		_, err := service.CreatePool("pool_1", "Test Pool", assets)
		require.NoError(t, err)

		err = service.SupplyAsset("pool_1", "BTC", "user_1", 1000.0)
		require.NoError(t, err)

		collateral := []Collateral{
			{
				AssetSymbol: "ETH",
				Amount:      10.0,
				Value:       2000.0,
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		// Create loan
		loan, err := service.BorrowAsset("pool_1", "BTC", "user_2", 500.0, collateral)
		require.NoError(t, err)

		assert.NotEmpty(t, loan.ID)
		assert.Equal(t, "user_2", loan.UserID)
		assert.Equal(t, "BTC", loan.AssetSymbol)
		assert.Equal(t, LoanTypeCollateralized, loan.Type)
		assert.Equal(t, 500.0, loan.Amount)
		assert.Equal(t, 500.0, loan.BorrowedAmount)
		assert.Equal(t, 0.0, loan.RepaidAmount)
		assert.Equal(t, LoanStatusActive, loan.Status)
		assert.Len(t, loan.Collateral, 1)
		assert.NotNil(t, loan.StartDate)
		assert.NotNil(t, loan.DueDate)
		assert.NotNil(t, loan.LastPayment)
	})
}

// TestCollateral tests the Collateral functionality
func TestCollateral(t *testing.T) {
	t.Run("CollateralValidation", func(t *testing.T) {
		collateral := Collateral{
			AssetSymbol: "ETH",
			Amount:      10.0,
			Value:       2000.0,
			LTV:         0.8,
			PledgedAt:   time.Now(),
		}

		assert.Equal(t, "ETH", collateral.AssetSymbol)
		assert.Equal(t, 10.0, collateral.Amount)
		assert.Equal(t, 2000.0, collateral.Value)
		assert.Equal(t, 0.8, collateral.LTV)
		assert.NotNil(t, collateral.PledgedAt)
	})
}

// TestIntegration tests the complete lending workflow
func TestIntegration(t *testing.T) {
	t.Run("CompleteLendingWorkflow", func(t *testing.T) {
		service := NewLendingService()

		// 1. Create lending pool
		assets := []string{"BTC", "ETH", "USDT"}
		pool, err := service.CreatePool("main_pool", "Main Lending Pool", assets)
		require.NoError(t, err)
		assert.Equal(t, "main_pool", pool.ID)

		// 2. User supplies assets
		err = service.SupplyAsset("main_pool", "BTC", "alice", 1000.0)
		require.NoError(t, err)

		err = service.SupplyAsset("main_pool", "ETH", "bob", 500.0)
		require.NoError(t, err)

		// 3. Check pool state after supply
		pool, _ = service.GetPool("main_pool")
		assert.Equal(t, 1500.0, pool.TotalLiquidity)
		assert.Equal(t, 0.0, pool.TotalBorrowed)
		assert.Equal(t, 0.0, pool.UtilizationRate)

		// 4. User borrows assets
		collateral := []Collateral{
			{
				AssetSymbol: "USDT",
				Amount:      1000.0,
				Value:       1000.0,
				LTV:         0.8,
				PledgedAt:   time.Now(),
			},
		}

		loan, err := service.BorrowAsset("main_pool", "BTC", "charlie", 200.0, collateral)
		require.NoError(t, err)
		assert.Equal(t, LoanStatusActive, loan.Status)

		// 5. Check pool state after borrow
		pool, _ = service.GetPool("main_pool")
		assert.Equal(t, 1500.0, pool.TotalLiquidity)
		assert.Equal(t, 200.0, pool.TotalBorrowed)
		assert.InDelta(t, 0.133, pool.UtilizationRate, 0.001) // 200/1500 with tolerance

		// 6. User repays loan
		err = service.RepayLoan(loan.ID, 100.0)
		require.NoError(t, err)

		// 7. Check final state
		updatedLoan, _ := service.GetLoan(loan.ID)
		assert.Equal(t, 100.0, updatedLoan.RepaidAmount)
		assert.Equal(t, LoanStatusActive, updatedLoan.Status) // Not fully repaid

		pool, _ = service.GetPool("main_pool")
		assert.Equal(t, 100.0, pool.TotalBorrowed)
		assert.InDelta(t, 0.067, pool.UtilizationRate, 0.001) // 100/1500 with tolerance

		// 8. Check user states
		alice, _ := service.GetUser("alice")
		assert.Equal(t, 1000.0, alice.TotalSupplied)

		bob, _ := service.GetUser("bob")
		assert.Equal(t, 500.0, bob.TotalSupplied)

		charlie, _ := service.GetUser("charlie")
		assert.Equal(t, 100.0, charlie.TotalBorrowed)
		assert.Len(t, charlie.Loans, 1)
	})
}
