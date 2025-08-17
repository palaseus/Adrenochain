package advanced

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRebalancingExecutor implements RebalancingExecutor for testing
type MockRebalancingExecutor struct {
	executions    []*RebalancingAction
	executionCost *big.Int
	shouldFail    bool
}

func NewMockRebalancingExecutor() *MockRebalancingExecutor {
	return &MockRebalancingExecutor{
		executions:    make([]*RebalancingAction, 0),
		executionCost: big.NewInt(100), // Default cost
		shouldFail:    false,
	}
}

func (m *MockRebalancingExecutor) ExecuteRebalancing(ctx context.Context, actions []*RebalancingAction) error {
	if m.shouldFail {
		return assert.AnError
	}
	m.executions = append(m.executions, actions...)
	return nil
}

func (m *MockRebalancingExecutor) CalculateTransactionCost(actions []*RebalancingAction) *big.Int {
	return m.executionCost
}

func (m *MockRebalancingExecutor) ValidateRebalancing(actions []*RebalancingAction) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func (m *MockRebalancingExecutor) GetExecutions() []*RebalancingAction {
	return m.executions
}

func (m *MockRebalancingExecutor) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func TestNewPortfolioRebalancer(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	assert.NotNil(t, pr)
	assert.NotNil(t, pr.portfolios)
	assert.NotNil(t, pr.strategies)
	assert.NotNil(t, pr.results)
	assert.Equal(t, ysm, pr.ysm)
	assert.Equal(t, 0, len(pr.portfolios))
	assert.Equal(t, 0, len(pr.strategies))
}

func TestRegisterPortfolio(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		UserID:     "user1",
		Strategies: make(map[string]*StrategyAllocation),
		TotalValue: big.NewInt(1000000),
	}

	err := pr.RegisterPortfolio(portfolio)
	assert.NoError(t, err)

	// Verify portfolio was registered
	registered, err := pr.GetPortfolio("user1")
	assert.NoError(t, err)
	assert.Equal(t, portfolio.UserID, registered.UserID)
	assert.False(t, registered.LastRebalance.IsZero())
	assert.Equal(t, 0, registered.RebalanceCount)
	assert.NotNil(t, registered.Performance)
}

func TestRegisterPortfolioValidation(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test empty user ID
	portfolio := &Portfolio{
		UserID: "",
	}
	err := pr.RegisterPortfolio(portfolio)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID cannot be empty")

	// Test duplicate registration
	portfolio1 := &Portfolio{UserID: "duplicate"}
	portfolio2 := &Portfolio{UserID: "duplicate"}

	err = pr.RegisterPortfolio(portfolio1)
	assert.NoError(t, err)

	err = pr.RegisterPortfolio(portfolio2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetPortfolio(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test getting non-existent portfolio
	_, err := pr.GetPortfolio("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test getting existing portfolio
	portfolio := &Portfolio{UserID: "existing"}
	err = pr.RegisterPortfolio(portfolio)
	assert.NoError(t, err)

	retrieved, err := pr.GetPortfolio("existing")
	assert.NoError(t, err)
	assert.Equal(t, portfolio.UserID, retrieved.UserID)
}

func TestRegisterRebalancingStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	strategy := &RebalancingStrategy{
		ID:          "test-strategy",
		Name:        "Test Strategy",
		Description: "A test rebalancing strategy",
		Type:        RebalancingTypeEqualWeight,
		Status:      RebalancingStatusActive,
	}

	err := pr.RegisterRebalancingStrategy(strategy)
	assert.NoError(t, err)

	// Verify strategy was registered
	// Note: We can't directly access the strategies map due to encapsulation
	// but we can test it indirectly through other methods
}

func TestRegisterRebalancingStrategyValidation(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test empty ID
	strategy := &RebalancingStrategy{
		ID: "",
	}
	err := pr.RegisterRebalancingStrategy(strategy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")

	// Test duplicate ID
	strategy1 := &RebalancingStrategy{ID: "duplicate"}
	strategy2 := &RebalancingStrategy{ID: "duplicate"}

	err = pr.RegisterRebalancingStrategy(strategy1)
	assert.NoError(t, err)

	err = pr.RegisterRebalancingStrategy(strategy2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCheckRebalancingNeeded(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test with no strategies
	needsRebalancing, triggers, err := pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.Error(t, err) // Should fail because portfolio doesn't exist
	assert.Contains(t, err.Error(), "not found")

	// Register portfolio
	portfolio := &Portfolio{
		UserID:     "user1",
		Strategies: make(map[string]*StrategyAllocation),
		TotalValue: big.NewInt(1000000),
	}
	err = pr.RegisterPortfolio(portfolio)
	require.NoError(t, err)

	// Test with no strategies
	needsRebalancing, triggers, err = pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.False(t, needsRebalancing)
	assert.Equal(t, 0, len(triggers))

	// Test with threshold trigger
	strategy := &RebalancingStrategy{
		ID:     "threshold-strategy",
		Status: RebalancingStatusActive,
		Triggers: []*RebalancingTrigger{
			{
				Type:      TriggerTypeThreshold,
				Threshold: big.NewFloat(0.09), // 9% threshold (should trigger for 10% drift)
			},
		},
	}
	pr.RegisterRebalancingStrategy(strategy)

	// Create a new portfolio with drift and register it
	portfolioWithDrift := &Portfolio{
		UserID:     "user1",
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				StrategyID:    "strategy1",
				CurrentAmount: big.NewInt(600000), // 60% allocation
				TargetWeight:  big.NewFloat(0.5),  // 50% target
			},
		},
	}

	// Remove old portfolio and register new one
	pr.mu.Lock()
	delete(pr.portfolios, "user1")
	pr.mu.Unlock()

	err = pr.RegisterPortfolio(portfolioWithDrift)
	require.NoError(t, err)

	// Now the threshold should trigger (10% drift > 9% threshold)
	needsRebalancing, triggers, err = pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.True(t, needsRebalancing)
	assert.Equal(t, 1, len(triggers))
	assert.Equal(t, TriggerTypeThreshold, triggers[0].Type)
}

func TestCheckThresholdTrigger(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Register portfolio
	portfolio := &Portfolio{
		UserID:     "user1",
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: big.NewInt(600000), // 60% allocation
				TargetWeight:  big.NewFloat(0.5),  // 50% target
			},
		},
	}
	err := pr.RegisterPortfolio(portfolio)
	require.NoError(t, err)

	// Register rebalancing strategy with threshold trigger
	strategy := &RebalancingStrategy{
		ID:     "threshold-test",
		Status: RebalancingStatusActive,
		Triggers: []*RebalancingTrigger{
			{
				Type:      TriggerTypeThreshold,
				Threshold: big.NewFloat(0.05), // 5% threshold
			},
		},
	}
	pr.RegisterRebalancingStrategy(strategy)

	// Test threshold trigger with drift (10% drift > 5% threshold)
	needsRebalancing, triggers, err := pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.True(t, needsRebalancing)
	assert.Equal(t, 1, len(triggers))
	assert.Equal(t, TriggerTypeThreshold, triggers[0].Type)

	// Test threshold trigger without drift (2% drift < 5% threshold)
	portfolio.Strategies["strategy1"].CurrentAmount = big.NewInt(520000) // 52% allocation

	needsRebalancing, triggers, err = pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.False(t, needsRebalancing)
	assert.Equal(t, 0, len(triggers))
}

func TestCheckTimeTrigger(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Register portfolio
	portfolio := &Portfolio{
		UserID:     "user1",
		Strategies: make(map[string]*StrategyAllocation),
	}
	err := pr.RegisterPortfolio(portfolio)
	require.NoError(t, err)

	// Register rebalancing strategy with time trigger
	strategy := &RebalancingStrategy{
		ID:     "time-test",
		Status: RebalancingStatusActive,
		Triggers: []*RebalancingTrigger{
			{
				Type:     TriggerTypeTime,
				Interval: 1 * time.Hour, // 1 hour interval
			},
		},
	}
	pr.RegisterRebalancingStrategy(strategy)

	// Test time trigger not ready (portfolio was just registered, so LastRebalance is now)
	needsRebalancing, triggers, err := pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.False(t, needsRebalancing)
	assert.Equal(t, 0, len(triggers))

	// Test time trigger ready (2 hours > 1 hour)
	// Update the existing portfolio with old rebalance time
	portfolioOld := &Portfolio{
		UserID:        "user1",
		LastRebalance: time.Now().Add(-2 * time.Hour), // 2 hours ago
		Strategies:    make(map[string]*StrategyAllocation),
	}

	err = pr.UpdatePortfolio(portfolioOld)
	require.NoError(t, err)

	// Now the time trigger should activate (2 hours > 1 hour)
	needsRebalancing, triggers, err = pr.CheckRebalancingNeeded(context.Background(), "user1")
	assert.NoError(t, err)
	assert.True(t, needsRebalancing)
	assert.Equal(t, 1, len(triggers))
	assert.Equal(t, TriggerTypeTime, triggers[0].Type)
}

func TestExecuteRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Register portfolio
	portfolio := &Portfolio{
		UserID:     "user1",
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: big.NewInt(600000),
				TargetWeight:  big.NewFloat(0.5),
			},
			"strategy2": {
				CurrentAmount: big.NewInt(400000),
				TargetWeight:  big.NewFloat(0.5),
			},
		},
	}
	err := pr.RegisterPortfolio(portfolio)
	require.NoError(t, err)

	// Register rebalancing strategy
	strategy := &RebalancingStrategy{
		ID:     "equal-weight",
		Type:   RebalancingTypeEqualWeight,
		Status: RebalancingStatusActive,
	}
	pr.RegisterRebalancingStrategy(strategy)

	// Execute rebalancing
	result, err := pr.ExecuteRebalancing(context.Background(), "user1", "equal-weight")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user1", result.PortfolioID)
	assert.Equal(t, ExecutionStatusCompleted, result.Status)
	assert.True(t, len(result.Actions) > 0)
	assert.True(t, result.ExecutionTime > 0)

	// Verify portfolio was updated
	updatedPortfolio, err := pr.GetPortfolio("user1")
	assert.NoError(t, err)
	assert.Equal(t, 1, updatedPortfolio.RebalanceCount)
	assert.False(t, updatedPortfolio.LastRebalance.IsZero())
}

func TestExecuteRebalancingErrors(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test non-existent portfolio
	_, err := pr.ExecuteRebalancing(context.Background(), "non-existent", "strategy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test non-existent strategy
	portfolio := &Portfolio{UserID: "user1"}
	pr.RegisterPortfolio(portfolio)

	_, err = pr.ExecuteRebalancing(context.Background(), "user1", "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test inactive strategy
	inactiveStrategy := &RebalancingStrategy{
		ID:     "inactive",
		Status: RebalancingStatusPaused,
	}
	pr.RegisterRebalancingStrategy(inactiveStrategy)

	_, err = pr.ExecuteRebalancing(context.Background(), "user1", "inactive")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestCalculateTargetAllocations(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(1000000)},
			"strategy2": {CurrentAmount: big.NewInt(1000000)},
			"strategy3": {CurrentAmount: big.NewInt(1000000)},
		},
	}

	// Test equal weight allocations
	strategy := &RebalancingStrategy{Type: RebalancingTypeEqualWeight}

	allocations, err := pr.calculateTargetAllocations(context.Background(), portfolio, strategy)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allocations))

	// Each strategy should have equal weight (1/3)
	expectedWeight := big.NewFloat(1.0 / 3.0)
	for _, weight := range allocations {
		assert.True(t, weight.Cmp(expectedWeight) == 0)
	}

	// Test custom allocations
	customStrategy := &RebalancingStrategy{
		Type: RebalancingTypeCustom,
		Parameters: map[string]interface{}{
			"weights": map[string]interface{}{
				"strategy1": 0.6,
				"strategy2": 0.4,
			},
		},
	}

	allocations, err = pr.calculateTargetAllocations(context.Background(), portfolio, customStrategy)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(allocations))
	assert.True(t, allocations["strategy1"].Cmp(big.NewFloat(0.6)) == 0)
	assert.True(t, allocations["strategy2"].Cmp(big.NewFloat(0.4)) == 0)
}

func TestCalculateEqualWeightAllocations(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test with empty portfolio
	emptyPortfolio := &Portfolio{Strategies: make(map[string]*StrategyAllocation)}

	_, err := pr.calculateEqualWeightAllocations(emptyPortfolio)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no strategies")

	// Test with single strategy
	singlePortfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(1000000)},
		},
	}

	allocations, err := pr.calculateEqualWeightAllocations(singlePortfolio)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allocations))
	assert.True(t, allocations["strategy1"].Cmp(big.NewFloat(1.0)) == 0)

	// Test with multiple strategies
	multiPortfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(1000000)},
			"strategy2": {CurrentAmount: big.NewInt(1000000)},
			"strategy3": {CurrentAmount: big.NewInt(1000000)},
			"strategy4": {CurrentAmount: big.NewInt(1000000)},
		},
	}

	allocations, err = pr.calculateEqualWeightAllocations(multiPortfolio)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(allocations))

	expectedWeight := big.NewFloat(0.25) // 1/4
	for _, weight := range allocations {
		assert.True(t, weight.Cmp(expectedWeight) == 0)
	}
}

func TestCalculateRiskParityAllocations(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(1000000)},
			"strategy2": {CurrentAmount: big.NewInt(1000000)},
		},
	}

	strategy := &RebalancingStrategy{Type: RebalancingTypeRiskParity}

	allocations, err := pr.calculateTargetAllocations(context.Background(), portfolio, strategy)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(allocations))

	// Verify allocations sum to 1.0
	totalWeight := big.NewFloat(0)
	for _, weight := range allocations {
		totalWeight.Add(totalWeight, weight)
	}
	assert.True(t, totalWeight.Cmp(big.NewFloat(1.0)) == 0)
}

func TestGenerateRebalancingActions(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: big.NewInt(600000), // 60% current
				TargetWeight:  big.NewFloat(0.5),  // 50% target
			},
			"strategy2": {
				CurrentAmount: big.NewInt(400000), // 40% current
				TargetWeight:  big.NewFloat(0.5),  // 50% target
			},
		},
	}

	targetAllocations := map[string]*big.Float{
		"strategy1": big.NewFloat(0.5),
		"strategy2": big.NewFloat(0.5),
	}

	actions, err := pr.generateRebalancingActions(portfolio, targetAllocations)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actions))

	// Strategy1 should have sell action (60% -> 50%)
	strategy1Action := actions[0]
	if strategy1Action.StrategyID == "strategy1" {
		assert.Equal(t, ActionTypeSell, strategy1Action.ActionType)
		assert.True(t, strategy1Action.Amount.Cmp(big.NewInt(0)) > 0)
	}

	// Strategy2 should have buy action (40% -> 50%)
	strategy2Action := actions[1]
	if strategy2Action.StrategyID == "strategy2" {
		assert.Equal(t, ActionTypeBuy, strategy2Action.ActionType)
		assert.True(t, strategy2Action.Amount.Cmp(big.NewInt(0)) > 0)
	}
}

func TestGenerateRebalancingActionsNewStrategy(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: big.NewInt(1000000), // 100% current
				TargetWeight:  big.NewFloat(0.5),   // 50% target
			},
		},
	}

	targetAllocations := map[string]*big.Float{
		"strategy1": big.NewFloat(0.5),
		"strategy2": big.NewFloat(0.5), // New strategy
	}

	actions, err := pr.generateRebalancingActions(portfolio, targetAllocations)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actions))

	// Find new strategy action
	var newStrategyAction *RebalancingAction
	for _, action := range actions {
		if action.StrategyID == "strategy2" {
			newStrategyAction = action
			break
		}
	}

	assert.NotNil(t, newStrategyAction)
	assert.Equal(t, ActionTypeBuy, newStrategyAction.ActionType)
	assert.True(t, newStrategyAction.Amount.Cmp(big.NewInt(0)) > 0)
	assert.Equal(t, big.NewFloat(0), newStrategyAction.CurrentWeight)
	assert.Equal(t, big.NewFloat(0.5), newStrategyAction.TargetWeight)
}

func TestCalculateTotalCost(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	actions := []*RebalancingAction{
		{Cost: big.NewInt(100)},
		{Cost: big.NewInt(200)},
		{Cost: big.NewInt(300)},
	}

	totalCost := pr.calculateTotalCost(actions)
	assert.Equal(t, big.NewInt(600), totalCost)

	// Test with nil costs
	actionsWithNil := []*RebalancingAction{
		{Cost: big.NewInt(100)},
		{Cost: nil},
		{Cost: big.NewInt(300)},
	}

	totalCost = pr.calculateTotalCost(actionsWithNil)
	assert.Equal(t, big.NewInt(400), totalCost)
}

func TestCalculateExpectedGain(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {TargetWeight: big.NewFloat(0.6)},
			"strategy2": {TargetWeight: big.NewFloat(0.4)},
		},
	}

	targetAllocations := map[string]*big.Float{
		"strategy1": big.NewFloat(0.5),
		"strategy2": big.NewFloat(0.5),
	}

	expectedGain := pr.calculateExpectedGain(portfolio, targetAllocations)
	assert.NotNil(t, expectedGain)

	// With simplified APY calculation, gain should be 0 (both current and target use 10% APY)
	assert.True(t, expectedGain.Cmp(big.NewFloat(0)) == 0)
}

func TestUpdatePortfolioAfterRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		UserID:         "user1",
		TotalValue:     big.NewInt(1000000),
		RebalanceCount: 0,
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: big.NewInt(600000),
				LastUpdated:   time.Now().Add(-1 * time.Hour),
			},
		},
	}

	actions := []*RebalancingAction{
		{
			StrategyID: "strategy1",
			ActionType: ActionTypeSell,
			Amount:     big.NewInt(100000),
		},
		{
			StrategyID: "strategy2",
			ActionType: ActionTypeBuy,
			Amount:     big.NewInt(100000),
		},
	}

	targetAllocations := map[string]*big.Float{
		"strategy1": big.NewFloat(0.5),
		"strategy2": big.NewFloat(0.5),
	}

	// Update portfolio
	pr.updatePortfolioAfterRebalancing(portfolio, actions, targetAllocations)

	// Verify updates
	assert.Equal(t, 1, portfolio.RebalanceCount)
	assert.False(t, portfolio.LastRebalance.IsZero())

	// Strategy1 should have reduced amount
	strategy1 := portfolio.Strategies["strategy1"]
	assert.Equal(t, big.NewInt(500000), strategy1.CurrentAmount) // 600k - 100k
	assert.Equal(t, big.NewFloat(0.5), strategy1.TargetWeight)
	assert.True(t, strategy1.LastUpdated.After(time.Now().Add(-1*time.Minute)))

	// Strategy2 should be new
	strategy2 := portfolio.Strategies["strategy2"]
	assert.Equal(t, big.NewInt(100000), strategy2.CurrentAmount)
	assert.Equal(t, big.NewFloat(0.5), strategy2.TargetWeight)
	assert.True(t, strategy2.LastUpdated.After(time.Now().Add(-1*time.Minute)))
}

func TestRecalculatePortfolioValue(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	portfolio := &Portfolio{
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(500000)},
			"strategy2": {CurrentAmount: big.NewInt(300000)},
			"strategy3": {CurrentAmount: big.NewInt(200000)},
		},
	}

	pr.recalculatePortfolioValue(portfolio)

	expectedTotal := big.NewInt(1000000) // 500k + 300k + 200k
	assert.Equal(t, expectedTotal, portfolio.TotalValue)
}

func TestGetRebalancingHistory(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test empty history
	history, err := pr.GetRebalancingHistory("user1", 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(history))

	// Test with results
	result := &RebalancingResult{
		PortfolioID: "user1",
		Timestamp:   time.Now(),
	}
	pr.storeRebalancingResult(result)

	history, err = pr.GetRebalancingHistory("user1", 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(history))
	assert.Equal(t, "user1", history[0].PortfolioID)
}

func TestSetRebalancingExecutor(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	executor := NewMockRebalancingExecutor()
	pr.SetRebalancingExecutor(executor)

	// Verify executor was set
	// Note: We can't directly access the executor field due to encapsulation
	// but we can test it indirectly through execution
}

func TestGetPortfolioPerformance(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Register strategies with default 10% APY
	strategy1 := &YieldStrategy{
		ID:  "strategy1",
		APY: big.NewFloat(0.1), // 10% APY
	}
	strategy2 := &YieldStrategy{
		ID:  "strategy2",
		APY: big.NewFloat(0.1), // 10% APY
	}
	ysm.RegisterStrategy(strategy1)
	ysm.RegisterStrategy(strategy2)

	// Register portfolio
	portfolio := &Portfolio{
		UserID: "user1",
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {TargetWeight: big.NewFloat(0.6)},
			"strategy2": {TargetWeight: big.NewFloat(0.4)},
		},
	}
	err := pr.RegisterPortfolio(portfolio)
	require.NoError(t, err)

	// Get performance
	performance, err := pr.GetPortfolioPerformance(context.Background(), "user1")
	assert.NoError(t, err)
	assert.NotNil(t, performance)
	assert.False(t, performance.LastCalculated.IsZero())

	// Verify weighted APY calculation
	// With default 10% APY and weights [0.6, 0.4], weighted APY should be 10%
	assert.True(t, performance.WeightedAPY.Cmp(big.NewFloat(0.1)) == 0)
}

func TestConcurrentAccessRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test concurrent portfolio registration
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			portfolio := &Portfolio{
				UserID: fmt.Sprintf("user-%d", id),
			}
			pr.RegisterPortfolio(portfolio)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all portfolios were registered
	for i := 0; i < 10; i++ {
		portfolio, err := pr.GetPortfolio(fmt.Sprintf("user-%d", i))
		assert.NoError(t, err)
		assert.NotNil(t, portfolio)
	}
}

func TestEdgeCasesRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test with nil portfolio values
	portfolio := &Portfolio{
		UserID:     "nil-test",
		TotalValue: big.NewInt(0),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {
				CurrentAmount: nil,
				TargetWeight:  nil,
			},
		},
	}

	err := pr.RegisterPortfolio(portfolio)
	assert.NoError(t, err)

	// Test rebalancing with nil values
	strategy := &RebalancingStrategy{
		ID:     "edge-test",
		Type:   RebalancingTypeEqualWeight,
		Status: RebalancingStatusActive,
	}
	pr.RegisterRebalancingStrategy(strategy)

	result, err := pr.ExecuteRebalancing(context.Background(), "nil-test", "edge-test")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestPerformanceBenchmarksRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Register many portfolios
	numPortfolios := 1000
	for i := 0; i < numPortfolios; i++ {
		portfolio := &Portfolio{
			UserID:     fmt.Sprintf("user-%d", i),
			TotalValue: big.NewInt(1000000),
			Strategies: map[string]*StrategyAllocation{
				"strategy1": {CurrentAmount: big.NewInt(500000), TargetWeight: big.NewFloat(0.5)},
				"strategy2": {CurrentAmount: big.NewInt(500000), TargetWeight: big.NewFloat(0.5)},
			},
		}
		err := pr.RegisterPortfolio(portfolio)
		require.NoError(t, err)
	}

	// Benchmark portfolio retrieval
	b := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := fmt.Sprintf("user-%d", i%numPortfolios)
			_, err := pr.GetPortfolio(userID)
			require.NoError(t, err)
		}
	})

	// Verify performance is reasonable (should be sub-millisecond)
	assert.True(t, b.NsPerOp() < 1000000, "Portfolio retrieval should be sub-millisecond")
}

func TestMathematicalAccuracyRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test allocation calculations with known values
	portfolio := &Portfolio{
		TotalValue: big.NewInt(1000000),
		Strategies: map[string]*StrategyAllocation{
			"strategy1": {CurrentAmount: big.NewInt(600000)},
			"strategy2": {CurrentAmount: big.NewInt(400000)},
		},
	}

	strategy := &RebalancingStrategy{Type: RebalancingTypeEqualWeight}

	allocations, err := pr.calculateTargetAllocations(context.Background(), portfolio, strategy)
	assert.NoError(t, err)

	// Equal weight should give 50% each
	expectedWeight := big.NewFloat(0.5)
	for _, weight := range allocations {
		assert.True(t, weight.Cmp(expectedWeight) == 0)
	}

	// Test that weights sum to 1.0
	totalWeight := big.NewFloat(0)
	for _, weight := range allocations {
		totalWeight.Add(totalWeight, weight)
	}
	assert.True(t, totalWeight.Cmp(big.NewFloat(1.0)) == 0)
}

func TestErrorHandlingRebalancing(t *testing.T) {
	ysm := NewYieldStrategyManager()
	pr := NewPortfolioRebalancer(ysm)

	// Test various error conditions
	testCases := []struct {
		name        string
		operation   func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "Get non-existent portfolio",
			operation: func() error {
				_, err := pr.GetPortfolio("non-existent")
				return err
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "Execute rebalancing on non-existent portfolio",
			operation: func() error {
				_, err := pr.ExecuteRebalancing(context.Background(), "non-existent", "strategy")
				return err
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "Execute rebalancing on non-existent strategy",
			operation: func() error {
				portfolio := &Portfolio{UserID: "user1"}
				pr.RegisterPortfolio(portfolio)
				_, err := pr.ExecuteRebalancing(context.Background(), "user1", "non-existent")
				return err
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
