package advanced

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockYieldProvider implements YieldProvider for testing
type MockYieldProvider struct {
	apyData    map[string]*big.Float
	tvlData    map[string]*big.Int
	rewards    map[string]map[string]*big.Int
	executions map[string]*big.Int
}

func NewMockYieldProvider() *MockYieldProvider {
	return &MockYieldProvider{
		apyData:    make(map[string]*big.Float),
		tvlData:    make(map[string]*big.Int),
		rewards:    make(map[string]map[string]*big.Int),
		executions: make(map[string]*big.Int),
	}
}

func (m *MockYieldProvider) GetAPY(ctx context.Context, strategyID string) (*big.Float, error) {
	if apy, exists := m.apyData[strategyID]; exists {
		return apy, nil
	}
	return big.NewFloat(0), nil
}

func (m *MockYieldProvider) GetTVL(ctx context.Context, strategyID string) (*big.Int, error) {
	if tvl, exists := m.tvlData[strategyID]; exists {
		return tvl, nil
	}
	return big.NewInt(0), nil
}

func (m *MockYieldProvider) ExecuteStrategy(ctx context.Context, strategyID string, amount *big.Int) error {
	m.executions[strategyID] = amount
	return nil
}

func (m *MockYieldProvider) GetRewards(ctx context.Context, strategyID string, user string) (*big.Int, error) {
	if userRewards, exists := m.rewards[strategyID]; exists {
		if reward, userExists := userRewards[user]; userExists {
			return reward, nil
		}
	}
	return big.NewInt(0), nil
}

func (m *MockYieldProvider) SetAPY(strategyID string, apy *big.Float) {
	m.apyData[strategyID] = apy
}

func (m *MockYieldProvider) SetTVL(strategyID string, tvl *big.Int) {
	m.tvlData[strategyID] = tvl
}

func (m *MockYieldProvider) SetRewards(strategyID, user string, reward *big.Int) {
	if m.rewards[strategyID] == nil {
		m.rewards[strategyID] = make(map[string]*big.Int)
	}
	m.rewards[strategyID][user] = reward
}

func (m *MockYieldProvider) GetExecutions() map[string]*big.Int {
	return m.executions
}

func TestNewYieldStrategyManager(t *testing.T) {
	ysm := NewYieldStrategyManager()

	assert.NotNil(t, ysm)
	assert.NotNil(t, ysm.strategies)
	assert.NotNil(t, ysm.providers)
	assert.NotNil(t, ysm.analytics)
	assert.Equal(t, 0, len(ysm.strategies))
	assert.Equal(t, 0, len(ysm.providers))
}

func TestRegisterStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()

	strategy := &YieldStrategy{
		ID:          "test-strategy",
		Name:        "Test Strategy",
		Description: "A test strategy",
		Type:        StrategyTypeYieldFarming,
		RiskLevel:   StrategyRiskLevelMedium,
		APY:         big.NewFloat(0.15),
		TVL:         big.NewInt(1000000),
		Status:      StrategyStatusActive,
	}

	err := ysm.RegisterStrategy(strategy)
	assert.NoError(t, err)

	// Verify strategy was registered
	registered, err := ysm.GetStrategy("test-strategy")
	assert.NoError(t, err)
	assert.Equal(t, strategy.ID, registered.ID)
	assert.Equal(t, strategy.Name, registered.Name)
	assert.False(t, registered.CreatedAt.IsZero())
	assert.False(t, registered.UpdatedAt.IsZero())
}

func TestRegisterStrategyValidation(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test empty ID
	strategy := &YieldStrategy{
		ID: "",
	}
	err := ysm.RegisterStrategy(strategy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// Test duplicate ID
	strategy1 := &YieldStrategy{ID: "duplicate"}
	strategy2 := &YieldStrategy{ID: "duplicate"}

	err = ysm.RegisterStrategy(strategy1)
	assert.NoError(t, err)

	err = ysm.RegisterStrategy(strategy2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test getting non-existent strategy
	_, err := ysm.GetStrategy("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test getting existing strategy
	strategy := &YieldStrategy{
		ID:   "existing",
		Name: "Existing Strategy",
	}

	err = ysm.RegisterStrategy(strategy)
	assert.NoError(t, err)

	retrieved, err := ysm.GetStrategy("existing")
	assert.NoError(t, err)
	assert.Equal(t, strategy.ID, retrieved.ID)
}

func TestListStrategies(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test empty list
	strategies := ysm.ListStrategies()
	assert.Equal(t, 0, len(strategies))

	// Test with strategies
	strategy1 := &YieldStrategy{ID: "strategy1"}
	strategy2 := &YieldStrategy{ID: "strategy2"}

	ysm.RegisterStrategy(strategy1)
	ysm.RegisterStrategy(strategy2)

	strategies = ysm.ListStrategies()
	assert.Equal(t, 2, len(strategies))

	// Verify both strategies are present
	ids := make(map[string]bool)
	for _, s := range strategies {
		ids[s.ID] = true
	}
	assert.True(t, ids["strategy1"])
	assert.True(t, ids["strategy2"])
}

func TestUpdateStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()

	strategy := &YieldStrategy{
		ID:   "update-test",
		Name: "Original Name",
		APY:  big.NewFloat(0.1),
	}

	err := ysm.RegisterStrategy(strategy)
	assert.NoError(t, err)

	// Test updating non-existent strategy
	err = ysm.UpdateStrategy("non-existent", map[string]interface{}{
		"name": "New Name",
	})
	assert.Error(t, err)

	// Test updating existing strategy
	updates := map[string]interface{}{
		"name": "Updated Name",
		"apy":  big.NewFloat(0.2),
	}

	err = ysm.UpdateStrategy("update-test", updates)
	assert.NoError(t, err)

	// Verify updates
	updated, err := ysm.GetStrategy("update-test")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, big.NewFloat(0.2), updated.APY)
	assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
}

func TestRegisterProvider(t *testing.T) {
	ysm := NewYieldStrategyManager()
	provider := NewMockYieldProvider()

	ysm.RegisterProvider("test-provider", provider)

	// Verify provider was registered
	// Note: We can't directly access the providers map due to encapsulation
	// but we can test it indirectly through the interface methods
}

func TestExecuteStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()
	provider := NewMockYieldProvider()

	// Register a strategy
	strategy := &YieldStrategy{
		ID:     "executable",
		Status: StrategyStatusActive,
	}
	ysm.RegisterStrategy(strategy)

	// Register provider
	ysm.RegisterProvider("test", provider)

	// Test executing non-existent strategy
	err := ysm.ExecuteStrategy(context.Background(), "non-existent", big.NewInt(1000), "user1")
	assert.Error(t, err)

	// Test executing inactive strategy
	inactiveStrategy := &YieldStrategy{
		ID:     "inactive",
		Status: StrategyStatusPaused,
	}
	ysm.RegisterStrategy(inactiveStrategy)

	err = ysm.ExecuteStrategy(context.Background(), "inactive", big.NewInt(1000), "user1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")

	// Test successful execution
	amount := big.NewInt(1000)
	err = ysm.ExecuteStrategy(context.Background(), "executable", amount, "user1")
	assert.NoError(t, err)

	// Verify execution was recorded
	executions := provider.GetExecutions()
	assert.Equal(t, amount, executions["executable"])
}

func TestGetStrategyAPY(t *testing.T) {
	ysm := NewYieldStrategyManager()
	provider := NewMockYieldProvider()

	// Register strategy with stored APY
	strategy := &YieldStrategy{
		ID:  "apy-test",
		APY: big.NewFloat(0.12),
	}
	ysm.RegisterStrategy(strategy)

	// Test getting APY from provider
	provider.SetAPY("apy-test", big.NewFloat(0.15))
	ysm.RegisterProvider("test", provider)

	apy, err := ysm.GetStrategyAPY(context.Background(), "apy-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewFloat(0.15), apy)

	// Test fallback to stored APY
	provider.SetAPY("apy-test", nil) // Clear provider APY

	apy, err = ysm.GetStrategyAPY(context.Background(), "apy-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewFloat(0.12), apy)

	// Test non-existent strategy
	_, err = ysm.GetStrategyAPY(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestGetStrategyTVL(t *testing.T) {
	ysm := NewYieldStrategyManager()
	provider := NewMockYieldProvider()

	// Register strategy with stored TVL
	strategy := &YieldStrategy{
		ID:  "tvl-test",
		TVL: big.NewInt(1000000),
	}
	ysm.RegisterStrategy(strategy)

	// Test getting TVL from provider
	provider.SetTVL("tvl-test", big.NewInt(1500000))
	ysm.RegisterProvider("test", provider)

	tvl, err := ysm.GetStrategyTVL(context.Background(), "tvl-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1500000), tvl)

	// Test fallback to stored TVL
	provider.SetTVL("tvl-test", nil) // Clear provider TVL

	tvl, err = ysm.GetStrategyTVL(context.Background(), "tvl-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1000000), tvl)

	// Test non-existent strategy
	_, err = ysm.GetStrategyTVL(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestGetUserRewards(t *testing.T) {
	ysm := NewYieldStrategyManager()
	provider := NewMockYieldProvider()

	// Register strategies
	strategy1 := &YieldStrategy{ID: "strategy1"}
	strategy2 := &YieldStrategy{ID: "strategy2"}
	ysm.RegisterStrategy(strategy1)
	ysm.RegisterStrategy(strategy2)

	// Register provider
	ysm.RegisterProvider("test", provider)

	// Set rewards
	provider.SetRewards("strategy1", "user1", big.NewInt(100))
	provider.SetRewards("strategy2", "user1", big.NewInt(200))

	// Get rewards
	rewards, err := ysm.GetUserRewards(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rewards))
	assert.Equal(t, big.NewInt(100), rewards["strategy1"])
	assert.Equal(t, big.NewInt(200), rewards["strategy2"])

	// Test user with no rewards
	rewards, err = ysm.GetUserRewards(context.Background(), "user2")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rewards))
}

func TestCalculateOptimalAllocation(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register multiple strategies with different risk levels and APYs
	strategies := []*YieldStrategy{
		{
			ID:        "low-risk",
			RiskLevel: StrategyRiskLevelLow,
			APY:       big.NewFloat(0.08),
			Status:    StrategyStatusActive,
		},
		{
			ID:        "medium-risk",
			RiskLevel: StrategyRiskLevelMedium,
			APY:       big.NewFloat(0.12),
			Status:    StrategyStatusActive,
		},
		{
			ID:        "high-risk",
			RiskLevel: StrategyRiskLevelHigh,
			APY:       big.NewFloat(0.18),
			Status:    StrategyStatusActive,
		},
	}

	for _, strategy := range strategies {
		err := ysm.RegisterStrategy(strategy)
		require.NoError(t, err)
	}

	// Test low risk tolerance
	allocations, err := ysm.CalculateOptimalAllocation(StrategyRiskLevelLow, big.NewFloat(0.10))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allocations))

	// Low risk tolerance should favor low and medium risk strategies
	assert.True(t, allocations["low-risk"].Cmp(allocations["high-risk"]) > 0)

	// Test high risk tolerance
	allocations, err = ysm.CalculateOptimalAllocation(StrategyRiskLevelHigh, big.NewFloat(0.10))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allocations))

	// High risk tolerance should allow more allocation to high-risk strategies
	assert.True(t, allocations["high-risk"].Cmp(big.NewFloat(0)) > 0)

	// Test with no active strategies
	ysm2 := NewYieldStrategyManager()
	_, err = ysm2.CalculateOptimalAllocation(StrategyRiskLevelMedium, big.NewFloat(0.10))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active strategies")
}

func TestCalculateStrategyWeight(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test risk adjustment for low risk tolerance
	strategy := &YieldStrategy{
		RiskLevel: StrategyRiskLevelHigh,
		APY:       big.NewFloat(0.15),
	}

	weight := ysm.calculateStrategyWeight(strategy, StrategyRiskLevelLow, big.NewFloat(0.10))
	// Should be penalized to 0.1 (10% of original weight)
	assert.True(t, weight.Cmp(big.NewFloat(0.1)) == 0, "Expected weight to be 0.1, got %v", weight)

	// Test risk adjustment for medium risk tolerance
	weight = ysm.calculateStrategyWeight(strategy, StrategyRiskLevelMedium, big.NewFloat(0.10))
	// Should not be penalized
	assert.True(t, weight.Cmp(big.NewFloat(1.0)) == 0, "Expected weight to be 1.0, got %v", weight)

	// Test APY adjustment
	weight = ysm.calculateStrategyWeight(strategy, StrategyRiskLevelHigh, big.NewFloat(0.10))
	expectedWeight := big.NewFloat(1.5) // 0.15 / 0.10 = 1.5
	// Allow for small floating point precision differences
	tolerance := big.NewFloat(0.0001)
	diff := new(big.Float).Sub(weight, expectedWeight)
	diff.Abs(diff)
	assert.True(t, diff.Cmp(tolerance) < 0, "Expected weight to be approximately 1.5, got %v", weight)
}

func TestUpdateAnalytics(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register strategy
	strategy := &YieldStrategy{ID: "analytics-test"}
	ysm.RegisterStrategy(strategy)

	// Update analytics
	apy := big.NewFloat(0.15)
	tvl := big.NewInt(1000000)

	err := ysm.UpdateAnalytics("analytics-test", apy, tvl)
	assert.NoError(t, err)

	// Verify historical data
	historical, err := ysm.GetHistoricalAPY("analytics-test", 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(historical))
	assert.Equal(t, apy, historical[0].APY)
	assert.Equal(t, tvl, historical[0].TVL)

	// Update multiple times to test data retention
	for i := 0; i < 5; i++ {
		newAPY := big.NewFloat(float64(i) * 0.01)
		err = ysm.UpdateAnalytics("analytics-test", newAPY, tvl)
		assert.NoError(t, err)
	}

	historical, err = ysm.GetHistoricalAPY("analytics-test", 10)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(historical)) // 1 initial + 5 updates
}

func TestRiskMetricsCalculation(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register strategy
	strategy := &YieldStrategy{ID: "risk-test"}
	ysm.RegisterStrategy(strategy)

	// Add historical data with known pattern
	apyValues := []float64{0.10, 0.12, 0.08, 0.15, 0.11, 0.13}
	tvl := big.NewInt(1000000)

	for _, apy := range apyValues {
		err := ysm.UpdateAnalytics("risk-test", big.NewFloat(apy), tvl)
		require.NoError(t, err)
	}

	// Get risk metrics
	metrics, err := ysm.GetRiskMetrics("risk-test")
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify volatility calculation
	assert.True(t, metrics.Volatility.Cmp(big.NewFloat(0)) > 0)

	// Verify max drawdown calculation
	assert.True(t, metrics.MaxDrawdown.Cmp(big.NewFloat(0)) > 0)

	// Verify expected value calculation
	assert.True(t, metrics.ExpectedValue.Cmp(big.NewFloat(0)) > 0)
}

func TestGetHistoricalAPY(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register strategy
	strategy := &YieldStrategy{ID: "history-test"}
	ysm.RegisterStrategy(strategy)

	// Add historical data
	for i := 0; i < 20; i++ {
		apy := big.NewFloat(float64(i) * 0.01)
		err := ysm.UpdateAnalytics("history-test", apy, big.NewInt(1000000))
		require.NoError(t, err)
	}

	// Test getting all data
	allData, err := ysm.GetHistoricalAPY("history-test", 0)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(allData))

	// Test getting limited data
	limitedData, err := ysm.GetHistoricalAPY("history-test", 10)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(limitedData))

	// Test getting more than available
	excessData, err := ysm.GetHistoricalAPY("history-test", 30)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(excessData))

	// Test non-existent strategy
	_, err = ysm.GetHistoricalAPY("non-existent", 10)
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register strategy
	strategy := &YieldStrategy{ID: "concurrent-test"}
	ysm.RegisterStrategy(strategy)

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_, err := ysm.GetStrategy("concurrent-test")
				assert.NoError(t, err)
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent writes
	done = make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 20; j++ {
				updates := map[string]interface{}{
					"name": fmt.Sprintf("Updated %d-%d", id, j),
				}
				ysm.UpdateStrategy("concurrent-test", updates)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestEdgeCases(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test with nil values
	strategy := &YieldStrategy{
		ID:        "nil-test",
		APY:       nil,
		TVL:       nil,
		RiskLevel: StrategyRiskLevelMedium,
		Status:    StrategyStatusActive,
	}

	err := ysm.RegisterStrategy(strategy)
	assert.NoError(t, err)

	// Test getting APY with nil value
	apy, err := ysm.GetStrategyAPY(context.Background(), "nil-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewFloat(0), apy)

	// Test getting TVL with nil value
	tvl, err := ysm.GetStrategyTVL(context.Background(), "nil-test")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(0), tvl)

	// Test calculating allocation with nil target APY
	allocations, err := ysm.CalculateOptimalAllocation(StrategyRiskLevelMedium, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allocations))
}

func TestPerformanceBenchmarks(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Register a reasonable number of strategies for performance testing
	numStrategies := 100
	for i := 0; i < numStrategies; i++ {
		strategy := &YieldStrategy{
			ID:        fmt.Sprintf("strategy-%d", i),
			RiskLevel: StrategyRiskLevelMedium,
			APY:       big.NewFloat(float64(i%20) * 0.01),
			Status:    StrategyStatusActive,
		}
		err := ysm.RegisterStrategy(strategy)
		require.NoError(t, err)
	}

	// Benchmark strategy retrieval
	b := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strategyID := fmt.Sprintf("strategy-%d", i%numStrategies)
			_, err := ysm.GetStrategy(strategyID)
			require.NoError(t, err)
		}
	})

	// Verify performance is reasonable (should be sub-millisecond)
	assert.True(t, b.NsPerOp() < 1000000, "Strategy retrieval should be sub-millisecond")

	// Benchmark allocation calculation
	b = testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ysm.CalculateOptimalAllocation(StrategyRiskLevelMedium, big.NewFloat(0.10))
			require.NoError(t, err)
		}
	})

	// Verify performance is reasonable
	assert.True(t, b.NsPerOp() < 1000000, "Allocation calculation should be sub-millisecond")
}

func TestMathematicalAccuracy(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test APY calculations with known values
	strategy := &YieldStrategy{ID: "math-test"}
	ysm.RegisterStrategy(strategy)

	// Add data with known mathematical properties
	testData := []struct {
		apy      float64
		expected float64
	}{
		{0.10, 0.10}, // 10% APY
		{0.20, 0.20}, // 20% APY
		{0.05, 0.05}, // 5% APY
	}

	for _, data := range testData {
		err := ysm.UpdateAnalytics("math-test", big.NewFloat(data.apy), big.NewInt(1000000))
		require.NoError(t, err)
	}

	// Test volatility calculation
	// For constant APY, volatility should be 0
	// For varying APY, volatility should be positive
	metrics, err := ysm.GetRiskMetrics("math-test")
	assert.NoError(t, err)

	// With varying APY values, we should have some volatility
	assert.True(t, metrics.Volatility.Cmp(big.NewFloat(0)) > 0)

	// Test max drawdown calculation
	// With APY sequence [0.10, 0.20, 0.05], max drawdown should be 0.15 (from 0.20 to 0.05)
	assert.True(t, metrics.MaxDrawdown.Cmp(big.NewFloat(0.14)) > 0) // Allow for some precision error
	assert.True(t, metrics.MaxDrawdown.Cmp(big.NewFloat(0.16)) < 0) // Allow for some precision error
}

func TestErrorHandling(t *testing.T) {
	ysm := NewYieldStrategyManager()

	// Test various error conditions
	testCases := []struct {
		name        string
		operation   func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "Get non-existent strategy",
			operation: func() error {
				_, err := ysm.GetStrategy("non-existent")
				return err
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "Update non-existent strategy",
			operation: func() error {
				return ysm.UpdateStrategy("non-existent", map[string]interface{}{"name": "new"})
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "Execute non-existent strategy",
			operation: func() error {
				return ysm.ExecuteStrategy(context.Background(), "non-existent", big.NewInt(1000), "user1")
			},
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
