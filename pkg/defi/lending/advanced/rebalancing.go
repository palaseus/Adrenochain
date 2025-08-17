package advanced

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// RebalancingTrigger represents when to trigger portfolio rebalancing
type RebalancingTrigger struct {
	Type      TriggerType   `json:"type"`
	Threshold *big.Float    `json:"threshold"`
	Interval  time.Duration `json:"interval"`
	LastCheck time.Time     `json:"last_check"`
}

// TriggerType represents the type of rebalancing trigger
type TriggerType string

const (
	TriggerTypeThreshold TriggerType = "threshold" // Rebalance when allocation drifts beyond threshold
	TriggerTypeTime      TriggerType = "time"      // Rebalance at regular time intervals
	TriggerTypeManual    TriggerType = "manual"    // Manual rebalancing trigger
	TriggerTypeEvent     TriggerType = "event"     // Event-based rebalancing
)

// RebalancingStrategy represents a portfolio rebalancing strategy
type RebalancingStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RebalancingType        `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Triggers    []*RebalancingTrigger  `json:"triggers"`
	Status      RebalancingStatus      `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RebalancingType represents the type of rebalancing strategy
type RebalancingType string

const (
	RebalancingTypeEqualWeight    RebalancingType = "equal_weight"
	RebalancingTypeRiskParity     RebalancingType = "risk_parity"
	RebalancingTypeBlackLitterman RebalancingType = "black_litterman"
	RebalancingTypeMeanVariance   RebalancingType = "mean_variance"
	RebalancingTypeCustom         RebalancingType = "custom"
)

// RebalancingStatus represents the status of a rebalancing strategy
type RebalancingStatus string

const (
	RebalancingStatusActive  RebalancingStatus = "active"
	RebalancingStatusPaused  RebalancingStatus = "paused"
	RebalancingStatusStopped RebalancingStatus = "stopped"
	RebalancingStatusRunning RebalancingStatus = "running"
)

// Portfolio represents a user's portfolio of yield strategies
type Portfolio struct {
	UserID         string                         `json:"user_id"`
	Strategies     map[string]*StrategyAllocation `json:"strategies"`
	TotalValue     *big.Int                       `json:"total_value"`
	LastRebalance  time.Time                      `json:"last_rebalance"`
	RebalanceCount int                            `json:"rebalance_count"`
	Performance    *PortfolioPerformance          `json:"performance"`
}

// StrategyAllocation represents allocation to a specific strategy
type StrategyAllocation struct {
	StrategyID    string     `json:"strategy_id"`
	CurrentAmount *big.Int   `json:"current_amount"`
	TargetAmount  *big.Int   `json:"target_amount"`
	TargetWeight  *big.Float `json:"target_weight"`
	LastUpdated   time.Time  `json:"last_updated"`
}

// PortfolioPerformance represents portfolio performance metrics
type PortfolioPerformance struct {
	TotalReturn    *big.Float `json:"total_return"`
	WeightedAPY    *big.Float `json:"weighted_apy"`
	Volatility     *big.Float `json:"volatility"`
	SharpeRatio    *big.Float `json:"sharpe_ratio"`
	MaxDrawdown    *big.Float `json:"max_drawdown"`
	RebalanceGain  *big.Float `json:"rebalance_gain"`
	LastCalculated time.Time  `json:"last_calculated"`
}

// RebalancingResult represents the result of a rebalancing operation
type RebalancingResult struct {
	PortfolioID   string                     `json:"portfolio_id"`
	Timestamp     time.Time                  `json:"timestamp"`
	Actions       []*RebalancingAction       `json:"actions"`
	TotalCost     *big.Int                   `json:"total_cost"`
	ExpectedGain  *big.Float                 `json:"expected_gain"`
	ExecutionTime time.Duration              `json:"execution_time"`
	Status        RebalancingExecutionStatus `json:"status"`
	Error         string                     `json:"error,omitempty"`
}

// RebalancingAction represents a specific action to rebalance the portfolio
type RebalancingAction struct {
	StrategyID    string     `json:"strategy_id"`
	ActionType    ActionType `json:"action_type"`
	Amount        *big.Int   `json:"amount"`
	CurrentWeight *big.Float `json:"current_weight"`
	TargetWeight  *big.Float `json:"target_weight"`
	Cost          *big.Int   `json:"cost"`
}

// ActionType represents the type of rebalancing action
type ActionType string

const (
	ActionTypeBuy       ActionType = "buy"
	ActionTypeSell      ActionType = "sell"
	ActionTypeHold      ActionType = "hold"
	ActionTypeRebalance ActionType = "rebalance"
)

// RebalancingExecutionStatus represents the status of rebalancing execution
type RebalancingExecutionStatus string

const (
	ExecutionStatusPending   RebalancingExecutionStatus = "pending"
	ExecutionStatusRunning   RebalancingExecutionStatus = "running"
	ExecutionStatusCompleted RebalancingExecutionStatus = "completed"
	ExecutionStatusFailed    RebalancingExecutionStatus = "failed"
)

// PortfolioRebalancer manages portfolio rebalancing operations
type PortfolioRebalancer struct {
	portfolios map[string]*Portfolio
	strategies map[string]*RebalancingStrategy
	results    map[string]*RebalancingResult
	mu         sync.RWMutex
	ysm        *YieldStrategyManager
	executor   RebalancingExecutor
}

// RebalancingExecutor executes rebalancing actions
type RebalancingExecutor interface {
	ExecuteRebalancing(ctx context.Context, actions []*RebalancingAction) error
	CalculateTransactionCost(actions []*RebalancingAction) *big.Int
	ValidateRebalancing(actions []*RebalancingAction) error
}

// NewPortfolioRebalancer creates a new portfolio rebalancer
func NewPortfolioRebalancer(ysm *YieldStrategyManager) *PortfolioRebalancer {
	return &PortfolioRebalancer{
		portfolios: make(map[string]*Portfolio),
		strategies: make(map[string]*RebalancingStrategy),
		results:    make(map[string]*RebalancingResult),
		ysm:        ysm,
	}
}

// RegisterPortfolio registers a new portfolio
func (pr *PortfolioRebalancer) RegisterPortfolio(portfolio *Portfolio) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if portfolio.UserID == "" {
		return errors.New("portfolio user ID cannot be empty")
	}

	if _, exists := pr.portfolios[portfolio.UserID]; exists {
		return fmt.Errorf("portfolio for user %s already exists", portfolio.UserID)
	}

	portfolio.LastRebalance = time.Now()
	portfolio.RebalanceCount = 0
	portfolio.Performance = &PortfolioPerformance{
		LastCalculated: time.Now(),
	}

	pr.portfolios[portfolio.UserID] = portfolio
	return nil
}

// GetPortfolio retrieves a portfolio by user ID
func (pr *PortfolioRebalancer) GetPortfolio(userID string) (*Portfolio, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	portfolio, exists := pr.portfolios[userID]
	if !exists {
		return nil, fmt.Errorf("portfolio for user %s not found", userID)
	}

	return portfolio, nil
}

// UpdatePortfolio updates an existing portfolio
func (pr *PortfolioRebalancer) UpdatePortfolio(portfolio *Portfolio) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if portfolio.UserID == "" {
		return errors.New("portfolio user ID cannot be empty")
	}

	if _, exists := pr.portfolios[portfolio.UserID]; !exists {
		return fmt.Errorf("portfolio for user %s not found", portfolio.UserID)
	}

	pr.portfolios[portfolio.UserID] = portfolio

	return nil
}

// RegisterRebalancingStrategy registers a new rebalancing strategy
func (pr *PortfolioRebalancer) RegisterRebalancingStrategy(strategy *RebalancingStrategy) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if strategy.ID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	if _, exists := pr.strategies[strategy.ID]; exists {
		return fmt.Errorf("rebalancing strategy with ID %s already exists", strategy.ID)
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	pr.strategies[strategy.ID] = strategy

	return nil
}

// CheckRebalancingNeeded checks if a portfolio needs rebalancing
func (pr *PortfolioRebalancer) CheckRebalancingNeeded(ctx context.Context, userID string) (bool, []*RebalancingTrigger, error) {
	portfolio, err := pr.GetPortfolio(userID)
	if err != nil {
		return false, nil, err
	}

	// Check all active rebalancing strategies
	var triggeredTriggers []*RebalancingTrigger
	needsRebalancing := false

	for _, strategy := range pr.strategies {
		if strategy.Status != RebalancingStatusActive {
			continue
		}

		for _, trigger := range strategy.Triggers {
			if pr.isTriggerActivated(ctx, portfolio, trigger) {
				triggeredTriggers = append(triggeredTriggers, trigger)
				needsRebalancing = true
			}
		}
	}

	return needsRebalancing, triggeredTriggers, nil
}

// isTriggerActivated checks if a specific trigger is activated
func (pr *PortfolioRebalancer) isTriggerActivated(ctx context.Context, portfolio *Portfolio, trigger *RebalancingTrigger) bool {
	switch trigger.Type {
	case TriggerTypeThreshold:
		return pr.checkThresholdTrigger(portfolio, trigger)
	case TriggerTypeTime:
		return pr.checkTimeTrigger(portfolio, trigger)
	case TriggerTypeManual:
		return false // Manual triggers are activated externally
	case TriggerTypeEvent:
		return false // Event triggers are handled separately
	default:
		return false
	}
}

// checkThresholdTrigger checks if threshold-based trigger is activated
func (pr *PortfolioRebalancer) checkThresholdTrigger(portfolio *Portfolio, trigger *RebalancingTrigger) bool {
	if trigger.Threshold == nil {
		return false
	}

	// Calculate current allocation drift
	for _, allocation := range portfolio.Strategies {
		if allocation.TargetWeight == nil {
			continue
		}

		// Calculate current weight
		currentWeight := big.NewFloat(0)
		if portfolio.TotalValue.Cmp(big.NewInt(0)) > 0 {
			currentWeight.Quo(new(big.Float).SetInt(allocation.CurrentAmount), new(big.Float).SetInt(portfolio.TotalValue))
		}

		// Check if drift exceeds threshold
		drift := new(big.Float).Sub(currentWeight, allocation.TargetWeight)
		drift.Abs(drift)

		if drift.Cmp(trigger.Threshold) > 0 {
			return true
		}
	}

	return false
}

// checkTimeTrigger checks if time-based trigger is activated
func (pr *PortfolioRebalancer) checkTimeTrigger(portfolio *Portfolio, trigger *RebalancingTrigger) bool {
	if trigger.Interval <= 0 {
		return false
	}

	timeSinceLastRebalance := time.Since(portfolio.LastRebalance)
	return timeSinceLastRebalance >= trigger.Interval
}

// ExecuteRebalancing executes portfolio rebalancing
func (pr *PortfolioRebalancer) ExecuteRebalancing(ctx context.Context, userID string, strategyID string) (*RebalancingResult, error) {
	portfolio, err := pr.GetPortfolio(userID)
	if err != nil {
		return nil, err
	}

	strategy, exists := pr.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("rebalancing strategy %s not found", strategyID)
	}

	if strategy.Status != RebalancingStatusActive {
		return nil, fmt.Errorf("rebalancing strategy %s is not active", strategyID)
	}

	// Create rebalancing result
	result := &RebalancingResult{
		PortfolioID: userID,
		Timestamp:   time.Now(),
		Status:      ExecutionStatusRunning,
	}

	startTime := time.Now()

	// Calculate target allocations
	targetAllocations, err := pr.calculateTargetAllocations(ctx, portfolio, strategy)
	if err != nil {
		result.Status = ExecutionStatusFailed
		result.Error = err.Error()
		result.ExecutionTime = time.Since(startTime)
		pr.storeRebalancingResult(result)
		return result, err
	}

	// Generate rebalancing actions
	actions, err := pr.generateRebalancingActions(portfolio, targetAllocations)
	if err != nil {
		result.Status = ExecutionStatusFailed
		result.Error = err.Error()
		result.ExecutionTime = time.Since(startTime)
		pr.storeRebalancingResult(result)
		return result, err
	}

	result.Actions = actions

	// Calculate costs and expected gains
	result.TotalCost = pr.calculateTotalCost(actions)
	result.ExpectedGain = pr.calculateExpectedGain(portfolio, targetAllocations)

	// Execute rebalancing if executor is available
	if pr.executor != nil {
		if err := pr.executor.ExecuteRebalancing(ctx, actions); err != nil {
			result.Status = ExecutionStatusFailed
			result.Error = err.Error()
		} else {
			result.Status = ExecutionStatusCompleted

			// Update portfolio
			pr.updatePortfolioAfterRebalancing(portfolio, actions, targetAllocations)
		}
	} else {
		// Simulate execution
		result.Status = ExecutionStatusCompleted
		pr.updatePortfolioAfterRebalancing(portfolio, actions, targetAllocations)
	}

	result.ExecutionTime = time.Since(startTime)
	pr.storeRebalancingResult(result)

	return result, nil
}

// calculateTargetAllocations calculates target allocations based on strategy type
func (pr *PortfolioRebalancer) calculateTargetAllocations(ctx context.Context, portfolio *Portfolio, strategy *RebalancingStrategy) (map[string]*big.Float, error) {
	switch strategy.Type {
	case RebalancingTypeEqualWeight:
		return pr.calculateEqualWeightAllocations(portfolio)
	case RebalancingTypeRiskParity:
		return pr.calculateRiskParityAllocations(ctx, portfolio)
	case RebalancingTypeMeanVariance:
		return pr.calculateMeanVarianceAllocations(ctx, portfolio)
	case RebalancingTypeCustom:
		return pr.calculateCustomAllocations(portfolio, strategy)
	default:
		return nil, fmt.Errorf("unsupported rebalancing type: %s", strategy.Type)
	}
}

// calculateEqualWeightAllocations calculates equal weight allocations
func (pr *PortfolioRebalancer) calculateEqualWeightAllocations(portfolio *Portfolio) (map[string]*big.Float, error) {
	if len(portfolio.Strategies) == 0 {
		return nil, errors.New("portfolio has no strategies")
	}

	allocations := make(map[string]*big.Float)
	equalWeight := new(big.Float).Quo(big.NewFloat(1.0), big.NewFloat(float64(len(portfolio.Strategies))))

	for strategyID := range portfolio.Strategies {
		allocations[strategyID] = new(big.Float).Copy(equalWeight)
	}

	return allocations, nil
}

// calculateRiskParityAllocations calculates risk parity allocations
func (pr *PortfolioRebalancer) calculateRiskParityAllocations(ctx context.Context, portfolio *Portfolio) (map[string]*big.Float, error) {
	allocations := make(map[string]*big.Float)
	totalRisk := big.NewFloat(0)

	// Calculate risk for each strategy
	risks := make(map[string]*big.Float)
	for strategyID := range portfolio.Strategies {
		risk, err := pr.calculateStrategyRisk(ctx, strategyID)
		if err != nil {
			// Use default risk if calculation fails
			risk = big.NewFloat(0.1)
		}
		risks[strategyID] = risk
		totalRisk.Add(totalRisk, risk)
	}

	if totalRisk.Cmp(big.NewFloat(0)) == 0 {
		return pr.calculateEqualWeightAllocations(portfolio)
	}

	// Allocate based on inverse risk
	totalInverseRisk := big.NewFloat(0)
	for strategyID, risk := range risks {
		inverseRisk := new(big.Float).Quo(big.NewFloat(1.0), risk)
		allocations[strategyID] = inverseRisk
		totalInverseRisk.Add(totalInverseRisk, inverseRisk)
	}

	// Normalize allocations to sum to 1.0
	for strategyID := range allocations {
		allocations[strategyID] = new(big.Float).Quo(allocations[strategyID], totalInverseRisk)
	}

	return allocations, nil
}

// calculateMeanVarianceAllocations calculates mean-variance optimal allocations
func (pr *PortfolioRebalancer) calculateMeanVarianceAllocations(ctx context.Context, portfolio *Portfolio) (map[string]*big.Float, error) {
	// This is a simplified mean-variance implementation
	// In practice, this would involve solving a quadratic programming problem

	// For now, use equal weight as a fallback
	return pr.calculateEqualWeightAllocations(portfolio)
}

// calculateCustomAllocations calculates custom allocations based on strategy parameters
func (pr *PortfolioRebalancer) calculateCustomAllocations(portfolio *Portfolio, strategy *RebalancingStrategy) (map[string]*big.Float, error) {
	allocations := make(map[string]*big.Float)

	// Extract custom weights from parameters
	if weights, ok := strategy.Parameters["weights"].(map[string]interface{}); ok {
		for strategyID, weight := range weights {
			if weightFloat, ok := weight.(float64); ok {
				allocations[strategyID] = big.NewFloat(weightFloat)
			}
		}
	}

	// If no custom weights found, fall back to equal weight
	if len(allocations) == 0 {
		return pr.calculateEqualWeightAllocations(portfolio)
	}

	return allocations, nil
}

// calculateStrategyRisk calculates the risk of a strategy
func (pr *PortfolioRebalancer) calculateStrategyRisk(ctx context.Context, strategyID string) (*big.Float, error) {
	// Get risk metrics from yield strategy manager
	metrics, err := pr.ysm.GetRiskMetrics(strategyID)
	if err != nil {
		return nil, err
	}

	// Use volatility as risk measure
	if metrics.Volatility != nil {
		return metrics.Volatility, nil
	}

	// Fallback to default risk
	return big.NewFloat(0.1), nil
}

// generateRebalancingActions generates actions needed to rebalance the portfolio
func (pr *PortfolioRebalancer) generateRebalancingActions(portfolio *Portfolio, targetAllocations map[string]*big.Float) ([]*RebalancingAction, error) {
	var actions []*RebalancingAction

	for strategyID, targetWeight := range targetAllocations {
		allocation, exists := portfolio.Strategies[strategyID]
		if !exists {
			// New strategy - buy action
			targetAmountFloat := new(big.Float).Mul(new(big.Float).SetInt(portfolio.TotalValue), targetWeight)
			targetAmount, _ := targetAmountFloat.Int(nil)
			action := &RebalancingAction{
				StrategyID:    strategyID,
				ActionType:    ActionTypeBuy,
				Amount:        targetAmount,
				CurrentWeight: big.NewFloat(0),
				TargetWeight:  targetWeight,
				Cost:          big.NewInt(0), // Will be calculated later
			}
			actions = append(actions, action)
			continue
		}

		// Ensure CurrentAmount is initialized
		if allocation.CurrentAmount == nil {
			allocation.CurrentAmount = big.NewInt(0)
		}

		// Calculate current weight
		currentWeight := big.NewFloat(0)
		if portfolio.TotalValue.Cmp(big.NewInt(0)) > 0 {
			currentWeight.Quo(new(big.Float).SetInt(allocation.CurrentAmount), new(big.Float).SetInt(portfolio.TotalValue))
		}

		// Calculate target amount
		targetAmountFloat := new(big.Float).Mul(new(big.Float).SetInt(portfolio.TotalValue), targetWeight)
		targetAmount, _ := targetAmountFloat.Int(nil)

		// Determine action type and amount
		if targetAmount.Cmp(allocation.CurrentAmount) > 0 {
			// Need to buy more
			buyAmount := new(big.Int).Sub(targetAmount, allocation.CurrentAmount)
			action := &RebalancingAction{
				StrategyID:    strategyID,
				ActionType:    ActionTypeBuy,
				Amount:        buyAmount,
				CurrentWeight: currentWeight,
				TargetWeight:  targetWeight,
				Cost:          big.NewInt(0),
			}
			actions = append(actions, action)
		} else if targetAmount.Cmp(allocation.CurrentAmount) < 0 {
			// Need to sell some
			sellAmount := new(big.Int).Sub(allocation.CurrentAmount, targetAmount)
			action := &RebalancingAction{
				StrategyID:    strategyID,
				ActionType:    ActionTypeSell,
				Amount:        sellAmount,
				CurrentWeight: currentWeight,
				TargetWeight:  targetWeight,
				Cost:          big.NewInt(0),
			}
			actions = append(actions, action)
		} else {
			// No action needed
			action := &RebalancingAction{
				StrategyID:    strategyID,
				ActionType:    ActionTypeHold,
				Amount:        big.NewInt(0),
				CurrentWeight: currentWeight,
				TargetWeight:  targetWeight,
				Cost:          big.NewInt(0),
			}
			actions = append(actions, action)
		}
	}

	return actions, nil
}

// calculateTotalCost calculates the total cost of rebalancing actions
func (pr *PortfolioRebalancer) calculateTotalCost(actions []*RebalancingAction) *big.Int {
	totalCost := big.NewInt(0)

	for _, action := range actions {
		if action.Cost != nil {
			totalCost.Add(totalCost, action.Cost)
		}
	}

	return totalCost
}

// calculateExpectedGain calculates the expected gain from rebalancing
func (pr *PortfolioRebalancer) calculateExpectedGain(portfolio *Portfolio, targetAllocations map[string]*big.Float) *big.Float {
	// This is a simplified calculation
	// In practice, this would involve more sophisticated modeling

	currentAPY := big.NewFloat(0)
	targetAPY := big.NewFloat(0)

	// Calculate current weighted APY
	for _, allocation := range portfolio.Strategies {
		if allocation.TargetWeight != nil {
			// Get strategy APY (simplified)
			apy := big.NewFloat(0.1) // Default 10% APY
			weightedAPY := new(big.Float).Mul(apy, allocation.TargetWeight)
			currentAPY.Add(currentAPY, weightedAPY)
		}
	}

	// Calculate target weighted APY
	for _, targetWeight := range targetAllocations {
		// Get strategy APY (simplified)
		apy := big.NewFloat(0.1) // Default 10% APY
		weightedAPY := new(big.Float).Mul(apy, targetWeight)
		targetAPY.Add(targetAPY, weightedAPY)
	}

	// Expected gain is the difference
	gain := new(big.Float).Sub(targetAPY, currentAPY)
	return gain
}

// updatePortfolioAfterRebalancing updates portfolio after rebalancing
func (pr *PortfolioRebalancer) updatePortfolioAfterRebalancing(portfolio *Portfolio, actions []*RebalancingAction, targetAllocations map[string]*big.Float) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	// Update allocations based on actions
	for _, action := range actions {
		allocation, exists := portfolio.Strategies[action.StrategyID]
		if !exists {
			// Create new allocation
			allocation = &StrategyAllocation{
				StrategyID:    action.StrategyID,
				CurrentAmount: big.NewInt(0),
				LastUpdated:   time.Now(),
			}
			portfolio.Strategies[action.StrategyID] = allocation
		}

		// Ensure CurrentAmount is initialized
		if allocation.CurrentAmount == nil {
			allocation.CurrentAmount = big.NewInt(0)
		}

		switch action.ActionType {
		case ActionTypeBuy:
			allocation.CurrentAmount.Add(allocation.CurrentAmount, action.Amount)
		case ActionTypeSell:
			allocation.CurrentAmount.Sub(allocation.CurrentAmount, action.Amount)
		}

		// Update target weight
		if targetWeight, exists := targetAllocations[action.StrategyID]; exists {
			allocation.TargetWeight = targetWeight
		}

		allocation.LastUpdated = time.Now()
	}

	// Update portfolio metadata
	portfolio.LastRebalance = time.Now()
	portfolio.RebalanceCount++

	// Recalculate total value
	pr.recalculatePortfolioValue(portfolio)
}

// recalculatePortfolioValue recalculates the total portfolio value
func (pr *PortfolioRebalancer) recalculatePortfolioValue(portfolio *Portfolio) {
	totalValue := big.NewInt(0)

	for _, allocation := range portfolio.Strategies {
		if allocation.CurrentAmount != nil {
			totalValue.Add(totalValue, allocation.CurrentAmount)
		}
	}

	portfolio.TotalValue = totalValue
}

// storeRebalancingResult stores a rebalancing result
func (pr *PortfolioRebalancer) storeRebalancingResult(result *RebalancingResult) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	resultID := fmt.Sprintf("%s-%d", result.PortfolioID, result.Timestamp.Unix())
	pr.results[resultID] = result
}

// GetRebalancingHistory returns rebalancing history for a portfolio
func (pr *PortfolioRebalancer) GetRebalancingHistory(userID string, limit int) ([]*RebalancingResult, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var userResults []*RebalancingResult

	for _, result := range pr.results {
		if result.PortfolioID == userID {
			userResults = append(userResults, result)
		}
	}

	// Sort by timestamp (newest first)
	// In a real implementation, you'd want proper sorting
	// For now, just return the results

	if limit > 0 && limit < len(userResults) {
		return userResults[:limit], nil
	}

	return userResults, nil
}

// SetRebalancingExecutor sets the rebalancing executor
func (pr *PortfolioRebalancer) SetRebalancingExecutor(executor RebalancingExecutor) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.executor = executor
}

// GetPortfolioPerformance calculates and returns portfolio performance
func (pr *PortfolioRebalancer) GetPortfolioPerformance(ctx context.Context, userID string) (*PortfolioPerformance, error) {
	portfolio, err := pr.GetPortfolio(userID)
	if err != nil {
		return nil, err
	}

	performance := &PortfolioPerformance{
		LastCalculated: time.Now(),
	}

	// Calculate weighted APY
	weightedAPY := big.NewFloat(0)
	totalWeight := big.NewFloat(0)

	for strategyID, allocation := range portfolio.Strategies {
		if allocation.TargetWeight != nil && allocation.TargetWeight.Cmp(big.NewFloat(0)) > 0 {
			apy, err := pr.ysm.GetStrategyAPY(ctx, strategyID)
			if err == nil && apy != nil {
				weightedAPY.Add(weightedAPY, new(big.Float).Mul(apy, allocation.TargetWeight))
				totalWeight.Add(totalWeight, allocation.TargetWeight)
			}
		}
	}

	if totalWeight.Cmp(big.NewFloat(0)) > 0 {
		performance.WeightedAPY = new(big.Float).Quo(weightedAPY, totalWeight)
	} else {
		performance.WeightedAPY = big.NewFloat(0)
	}

	// Calculate other metrics (simplified)
	performance.TotalReturn = big.NewFloat(0)    // Would calculate from historical data
	performance.Volatility = big.NewFloat(0.1)   // Would calculate from historical data
	performance.SharpeRatio = big.NewFloat(0.5)  // Would calculate from historical data
	performance.MaxDrawdown = big.NewFloat(0.05) // Would calculate from historical data

	// Update portfolio performance
	portfolio.Performance = performance

	return performance, nil
}
