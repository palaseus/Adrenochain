package amm

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkAMMOperations benchmarks basic AMM operations
func BenchmarkAMMOperations(b *testing.B) {
	// Create test addresses
	tokenA, _ := engine.ParseAddress("0x1234567890123456789012345678901234567890")
	tokenB, _ := engine.ParseAddress("0x0987654321098765432109876543210987654321")
	provider, _ := engine.ParseAddress("0x1111111111111111111111111111111111111111")
	user, _ := engine.ParseAddress("0x2222222222222222222222222222222222222222")

	// Create AMM instance
	amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

	// Add initial liquidity
	_, err := amm.AddLiquidity(provider, big.NewInt(1000000000000000000), big.NewInt(1000000000000000000), big.NewInt(0), 1, engine.Hash{})
	require.NoError(b, err)

	b.ResetTimer()
	b.Run("AddLiquidity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := amm.AddLiquidity(
				provider,
				big.NewInt(1000000000000000000), // 1 token
				big.NewInt(1000000000000000000), // 1 token
				big.NewInt(0),                   // No minimum LP tokens
				uint64(i+2),                     // Block number
				engine.Hash{},                   // Transaction hash
			)
			require.NoError(b, err)
		}
	})

	b.Run("Swap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := amm.Swap(
				user,
				tokenA,
				big.NewInt(100000000000000000), // 0.1 token
				big.NewInt(0),                  // No minimum output
				uint64(i+1000),                 // Block number
				engine.Hash{},                  // Transaction hash
			)
			require.NoError(b, err)
		}
	})

	b.Run("RemoveLiquidity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create a fresh AMM instance for each iteration to avoid LP token depletion
			freshAMM := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

			// Add liquidity to get LP tokens
			initialAmount := new(big.Int)
			initialAmount.SetString("1000000000000000000000", 10)
			lpTokens, err := freshAMM.AddLiquidity(provider, initialAmount, initialAmount, big.NewInt(0), 1, engine.Hash{})
			require.NoError(b, err)
			require.NotNil(b, lpTokens)
			require.True(b, lpTokens.Cmp(big.NewInt(0)) > 0, "LP tokens should be greater than 0")

			// Remove a small amount of liquidity
			removeAmount := new(big.Int).Div(lpTokens, big.NewInt(10)) // Remove 10% of LP tokens
			if removeAmount.Cmp(big.NewInt(0)) == 0 {
				removeAmount = big.NewInt(1) // Ensure we have at least 1 token to remove
			}

			_, _, err = freshAMM.RemoveLiquidity(
				provider,
				removeAmount,
				big.NewInt(0),  // No minimum amount A
				big.NewInt(0),  // No minimum amount B
				uint64(i+2000), // Block number
				engine.Hash{},  // Transaction hash
			)
			require.NoError(b, err)
		}
	})
}

// BenchmarkConcurrentOperations tests AMM performance under concurrent load
func BenchmarkConcurrentOperations(b *testing.B) {
	// Create test addresses
	tokenA, _ := engine.ParseAddress("0x1234567890123456789012345678901234567890")
	tokenB, _ := engine.ParseAddress("0x0987654321098765432109876543210987654321")
	provider, _ := engine.ParseAddress("0x1111111111111111111111111111111111111111")
	user, _ := engine.ParseAddress("0x2222222222222222222222222222222222222222")

	// Create AMM instance
	amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

	// Add initial liquidity
	largeAmount := new(big.Int)
	largeAmount.SetString("10000000000000000000000", 10)
	_, err := amm.AddLiquidity(provider, largeAmount, largeAmount, big.NewInt(0), 1, engine.Hash{})
	require.NoError(b, err)

	b.ResetTimer()
	b.Run("ConcurrentSwaps", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := amm.Swap(
					user,
					tokenA,
					big.NewInt(100000000000000000), // 0.1 token
					big.NewInt(0),                  // No minimum output
					1000,                           // Block number
					engine.Hash{},                  // Transaction hash
				)
				require.NoError(b, err)
			}
		})
	})

	b.Run("ConcurrentLiquidityOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Create a fresh AMM instance for each operation to avoid state conflicts
				freshAMM := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

				if pb.Next() {
					// Add liquidity
					_, err := freshAMM.AddLiquidity(
						provider,
						big.NewInt(1000000000000000000), // 1 token
						big.NewInt(1000000000000000000), // 1 token
						big.NewInt(0),                   // No minimum LP tokens
						1000,                            // Block number
						engine.Hash{},                   // Transaction hash
					)
					require.NoError(b, err)
				} else {
					// Add liquidity first, then remove a small amount
					initialAmount := new(big.Int)
					initialAmount.SetString("1000000000000000000000", 10)
					lpTokens, err := freshAMM.AddLiquidity(provider, initialAmount, initialAmount, big.NewInt(0), 1, engine.Hash{})
					require.NoError(b, err)

					// Remove a small fraction
					removeAmount := new(big.Int).Div(lpTokens, big.NewInt(10)) // Remove 10% of LP tokens
					if removeAmount.Cmp(big.NewInt(0)) == 0 {
						removeAmount = big.NewInt(1)
					}

					_, _, err = freshAMM.RemoveLiquidity(
						provider,
						removeAmount,
						big.NewInt(0), // No minimum amount A
						big.NewInt(0), // No minimum amount B
						1000,          // Block number
						engine.Hash{}, // Transaction hash
					)
					require.NoError(b, err)
				}
			}
		})
	})
}

// TestAMMLoadTesting tests AMM performance under high load
func TestAMMLoadTesting(t *testing.T) {
	// Create test addresses
	tokenA, _ := engine.ParseAddress("0x1234567890123456789012345678901234567890")
	tokenB, _ := engine.ParseAddress("0x0987654321098765432109876543210987654321")
	provider, _ := engine.ParseAddress("0x1111111111111111111111111111111111111111")
	user, _ := engine.ParseAddress("0x2222222222222222222222222222222222222222")

	// Create AMM instance
	amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

	// Add substantial initial liquidity
	largeAmount := new(big.Int)
	largeAmount.SetString("1000000000000000000000000", 10)
	_, err := amm.AddLiquidity(provider, largeAmount, largeAmount, big.NewInt(0), 1, engine.Hash{})
	require.NoError(t, err)

	// Test concurrent operations
	const numGoroutines = 100
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Perform swap
				_, err := amm.Swap(
					user,
					tokenA,
					big.NewInt(100000000000000000), // 0.1 token
					big.NewInt(0),                  // No minimum output
					uint64(id*1000+j),              // Block number
					engine.Hash{},                  // Transaction hash
				)
				require.NoError(t, err)

				// Add small liquidity
				_, err = amm.AddLiquidity(
					provider,
					big.NewInt(100000000000000000), // 0.1 token
					big.NewInt(100000000000000000), // 0.1 token
					big.NewInt(0),                  // No minimum LP tokens
					uint64(id*1000+j+10000),        // Block number
					engine.Hash{},                  // Transaction hash
				)
				require.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOperations := numGoroutines * operationsPerGoroutine
	operationsPerSecond := float64(totalOperations) / duration.Seconds()

	t.Logf("Load test completed:")
	t.Logf("  Total operations: %d", totalOperations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Operations per second: %.2f", operationsPerSecond)

	// Verify AMM state is consistent
	poolA, poolB := amm.GetReserves()
	assert.True(t, poolA.Cmp(big.NewInt(0)) > 0, "Pool A should have positive reserves")
	assert.True(t, poolB.Cmp(big.NewInt(0)) > 0, "Pool B should have positive reserves")
}

// TestAMMGasOptimization tests gas efficiency of AMM operations
func TestAMMGasOptimization(t *testing.T) {
	// Create test addresses
	tokenA, _ := engine.ParseAddress("0x1234567890123456789012345678901234567890")
	tokenB, _ := engine.ParseAddress("0x0987654321098765432109876543210987654321")
	provider, _ := engine.ParseAddress("0x1111111111111111111111111111111111111111")

	// Create AMM instance
	amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

	// Test gas efficiency of different operation sizes
	testCases := []struct {
		name     string
		amountA  *big.Int
		amountB  *big.Int
		expected bool
	}{
		{"Small amounts", big.NewInt(100000000000000000), big.NewInt(100000000000000000), true},
		{"Medium amounts", big.NewInt(1000000000000000000), big.NewInt(1000000000000000000), true},
		{"Large amounts", func() *big.Int { v, _ := new(big.Int).SetString("10000000000000000000", 10); return v }(), func() *big.Int { v, _ := new(big.Int).SetString("10000000000000000000", 10); return v }(), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startTime := time.Now()

			_, err := amm.AddLiquidity(provider, tc.amountA, tc.amountB, big.NewInt(0), 1, engine.Hash{})
			require.NoError(t, err)

			duration := time.Since(startTime)
			t.Logf("Operation completed in %v", duration)

			// Verify operation was successful
			poolA, poolB := amm.GetReserves()
			assert.True(t, poolA.Cmp(big.NewInt(0)) > 0, "Pool A should have positive reserves")
			assert.True(t, poolB.Cmp(big.NewInt(0)) > 0, "Pool B should have positive reserves")
		})
	}
}

// TestAMMScalability tests AMM performance with increasing pool sizes
func TestAMMScalability(t *testing.T) {
	// Test different pool sizes
	poolSizes := []int64{1000, 10000, 100000, 1000000} // 1K to 1M tokens

	for _, size := range poolSizes {
		t.Run(fmt.Sprintf("PoolSize_%d", size), func(t *testing.T) {
			// Create test addresses
			tokenA, _ := engine.ParseAddress("0x1234567890123456789012345678901234567890")
			tokenB, _ := engine.ParseAddress("0x0987654321098765432109876543210987654321")
			provider, _ := engine.ParseAddress("0x1111111111111111111111111111111111111111")
			user, _ := engine.ParseAddress("0x2222222222222222222222222222222222222222")

			// Create AMM instance
			amm := NewAMM("test-pool", tokenA, tokenB, "Test Pool", "TEST", 18, provider, big.NewInt(30))

			// Add initial liquidity
			_, err := amm.AddLiquidity(provider, big.NewInt(size*1000000000000000000), big.NewInt(size*1000000000000000000), big.NewInt(0), 1, engine.Hash{})
			require.NoError(t, err)

			// Measure time for multiple operations
			const numOperations = 100
			startTime := time.Now()

			for i := 0; i < numOperations; i++ {
				// Perform swap
				_, err := amm.Swap(
					user,
					tokenA,
					big.NewInt(100000000000000000), // 0.1 token
					big.NewInt(0),                  // No minimum output
					uint64(i+1000),                 // Block number
					engine.Hash{},                  // Transaction hash
				)
				require.NoError(t, err)
			}

			duration := time.Since(startTime)
			operationsPerSecond := float64(numOperations) / duration.Seconds()

			t.Logf("Pool size %d: %d operations in %v (%.2f ops/sec)",
				size, numOperations, duration, operationsPerSecond)

			// Verify final state
			poolA, poolB := amm.GetReserves()
			assert.True(t, poolA.Cmp(big.NewInt(0)) > 0, "Pool A should have positive reserves")
			assert.True(t, poolB.Cmp(big.NewInt(0)) > 0, "Pool B should have positive reserves")
		})
	}
}
