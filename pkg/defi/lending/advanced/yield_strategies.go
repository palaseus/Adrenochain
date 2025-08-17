package advanced

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// YieldStrategy represents a yield optimization strategy
type YieldStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        StrategyType           `json:"type"`
	RiskLevel   StrategyRiskLevel      `json:"risk_level"`
	APY         *big.Float             `json:"apy"`
	TVL         *big.Int               `json:"tvl"`
	Status      StrategyStatus         `json:"status"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// StrategyType represents the type of yield strategy
type StrategyType string

const (
	StrategyTypeYieldFarming     StrategyType = "yield_farming"
	StrategyTypeLiquidityMining  StrategyType = "liquidity_mining"
	StrategyTypeYieldAggregation StrategyType = "yield_aggregation"
	StrategyTypeStaking          StrategyType = "staking"
	StrategyTypeLending          StrategyType = "lending"
	StrategyTypeArbitrage        StrategyType = "arbitrage"
)

// StrategyRiskLevel represents the risk level of a yield strategy
type StrategyRiskLevel string

const (
	StrategyRiskLevelLow      StrategyRiskLevel = "low"
	StrategyRiskLevelMedium   StrategyRiskLevel = "medium"
	StrategyRiskLevelHigh     StrategyRiskLevel = "high"
	StrategyRiskLevelVeryHigh StrategyRiskLevel = "very_high"
)

// StrategyStatus represents the status of a strategy
type StrategyStatus string

const (
	StrategyStatusActive    StrategyStatus = "active"
	StrategyStatusPaused    StrategyStatus = "paused"
	StrategyStatusStopped   StrategyStatus = "stopped"
	StrategyStatusMigrating StrategyStatus = "migrating"
)

// YieldStrategyManager manages yield optimization strategies
type YieldStrategyManager struct {
	strategies map[string]*YieldStrategy
	mu         sync.RWMutex
	providers  map[string]YieldProvider
	analytics  *YieldAnalytics
}

// YieldProvider interface for different yield sources
type YieldProvider interface {
	GetAPY(ctx context.Context, strategyID string) (*big.Float, error)
	GetTVL(ctx context.Context, strategyID string) (*big.Int, error)
	ExecuteStrategy(ctx context.Context, strategyID string, amount *big.Int) error
	GetRewards(ctx context.Context, strategyID string, user string) (*big.Int, error)
}

// YieldAnalytics provides analytics for yield strategies
type YieldAnalytics struct {
	historicalAPY map[string][]APYDataPoint
	riskMetrics   map[string]*RiskMetrics
	mu            sync.RWMutex
}

// APYDataPoint represents a historical APY data point
type APYDataPoint struct {
	Timestamp time.Time  `json:"timestamp"`
	APY       *big.Float `json:"apy"`
	TVL       *big.Int   `json:"tvl"`
}

// RiskMetrics represents risk metrics for a strategy
type RiskMetrics struct {
	Volatility    *big.Float `json:"volatility"`
	MaxDrawdown   *big.Float `json:"max_drawdown"`
	SharpeRatio   *big.Float `json:"sharpe_ratio"`
	SortinoRatio  *big.Float `json:"sortino_ratio"`
	VaR95         *big.Float `json:"var_95"`
	ExpectedValue *big.Float `json:"expected_value"`
}

// NewYieldStrategyManager creates a new yield strategy manager
func NewYieldStrategyManager() *YieldStrategyManager {
	return &YieldStrategyManager{
		strategies: make(map[string]*YieldStrategy),
		providers:  make(map[string]YieldProvider),
		analytics: &YieldAnalytics{
			historicalAPY: make(map[string][]APYDataPoint),
			riskMetrics:   make(map[string]*RiskMetrics),
		},
	}
}

// RegisterStrategy registers a new yield strategy
func (ysm *YieldStrategyManager) RegisterStrategy(strategy *YieldStrategy) error {
	ysm.mu.Lock()
	defer ysm.mu.Unlock()

	if strategy.ID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	if _, exists := ysm.strategies[strategy.ID]; exists {
		return fmt.Errorf("strategy with ID %s already exists", strategy.ID)
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	ysm.strategies[strategy.ID] = strategy

	return nil
}

// GetStrategy retrieves a strategy by ID
func (ysm *YieldStrategyManager) GetStrategy(strategyID string) (*YieldStrategy, error) {
	ysm.mu.RLock()
	defer ysm.mu.RUnlock()

	strategy, exists := ysm.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("strategy with ID %s not found", strategyID)
	}

	return strategy, nil
}

// ListStrategies returns all registered strategies
func (ysm *YieldStrategyManager) ListStrategies() []*YieldStrategy {
	ysm.mu.RLock()
	defer ysm.mu.RUnlock()

	strategies := make([]*YieldStrategy, 0, len(ysm.strategies))
	for _, strategy := range ysm.strategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// UpdateStrategy updates an existing strategy
func (ysm *YieldStrategyManager) UpdateStrategy(strategyID string, updates map[string]interface{}) error {
	ysm.mu.Lock()
	defer ysm.mu.Unlock()

	strategy, exists := ysm.strategies[strategyID]
	if !exists {
		return fmt.Errorf("strategy with ID %s not found", strategyID)
	}

	// Update fields based on updates map
	if name, ok := updates["name"].(string); ok {
		strategy.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		strategy.Description = description
	}
	if apy, ok := updates["apy"].(*big.Float); ok {
		strategy.APY = apy
	}
	if tvl, ok := updates["tvl"].(*big.Int); ok {
		strategy.TVL = tvl
	}
	if status, ok := updates["status"].(StrategyStatus); ok {
		strategy.Status = status
	}

	strategy.UpdatedAt = time.Now()
	return nil
}

// RegisterProvider registers a yield provider
func (ysm *YieldStrategyManager) RegisterProvider(name string, provider YieldProvider) {
	ysm.mu.Lock()
	defer ysm.mu.Unlock()
	ysm.providers[name] = provider
}

// ExecuteStrategy executes a yield strategy
func (ysm *YieldStrategyManager) ExecuteStrategy(ctx context.Context, strategyID string, amount *big.Int, user string) error {
	ysm.mu.RLock()
	strategy, exists := ysm.strategies[strategyID]
	ysm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("strategy with ID %s not found", strategyID)
	}

	if strategy.Status != StrategyStatusActive {
		return fmt.Errorf("strategy %s is not active (status: %s)", strategyID, strategy.Status)
	}

	// Find appropriate provider and execute
	for _, provider := range ysm.providers {
		if err := provider.ExecuteStrategy(ctx, strategyID, amount); err == nil {
			return nil
		}
	}

	return errors.New("no suitable provider found for strategy execution")
}

// GetStrategyAPY gets the current APY for a strategy
func (ysm *YieldStrategyManager) GetStrategyAPY(ctx context.Context, strategyID string) (*big.Float, error) {
	ysm.mu.RLock()
	strategy, exists := ysm.strategies[strategyID]
	ysm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy with ID %s not found", strategyID)
	}

	// Try to get real-time APY from providers
	for _, provider := range ysm.providers {
		if apy, err := provider.GetAPY(ctx, strategyID); err == nil && apy != nil {
			return apy, nil
		}
	}

	// Fall back to stored APY
	if strategy.APY != nil {
		return strategy.APY, nil
	}

	return big.NewFloat(0), nil
}

// GetStrategyTVL gets the current TVL for a strategy
func (ysm *YieldStrategyManager) GetStrategyTVL(ctx context.Context, strategyID string) (*big.Int, error) {
	ysm.mu.RLock()
	strategy, exists := ysm.strategies[strategyID]
	ysm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("strategy with ID %s not found", strategyID)
	}

	// Try to get real-time TVL from providers
	for _, provider := range ysm.providers {
		if tvl, err := provider.GetTVL(ctx, strategyID); err == nil && tvl != nil {
			return tvl, nil
		}
	}

	// Fall back to stored TVL
	if strategy.TVL != nil {
		return strategy.TVL, nil
	}

	return big.NewInt(0), nil
}

// GetUserRewards gets rewards for a user across all strategies
func (ysm *YieldStrategyManager) GetUserRewards(ctx context.Context, user string) (map[string]*big.Int, error) {
	rewards := make(map[string]*big.Int)

	for _, provider := range ysm.providers {
		for _, strategy := range ysm.strategies {
			if reward, err := provider.GetRewards(ctx, strategy.ID, user); err == nil && reward.Cmp(big.NewInt(0)) > 0 {
				rewards[strategy.ID] = reward
			}
		}
	}

	return rewards, nil
}

// CalculateOptimalAllocation calculates optimal allocation across strategies
func (ysm *YieldStrategyManager) CalculateOptimalAllocation(riskTolerance StrategyRiskLevel, targetAPY *big.Float) (map[string]*big.Float, error) {
	ysm.mu.RLock()
	defer ysm.mu.RUnlock()

	strategies := make([]*YieldStrategy, 0)
	for _, strategy := range ysm.strategies {
		if strategy.Status == StrategyStatusActive {
			strategies = append(strategies, strategy)
		}
	}

	if len(strategies) == 0 {
		return nil, errors.New("no active strategies available")
	}

	// Simple allocation based on risk tolerance and APY
	allocations := make(map[string]*big.Float)
	totalWeight := big.NewFloat(0)

	for _, strategy := range strategies {
		weight := ysm.calculateStrategyWeight(strategy, riskTolerance, targetAPY)
		allocations[strategy.ID] = weight
		totalWeight.Add(totalWeight, weight)
	}

	// Normalize allocations to sum to 1.0
	if totalWeight.Cmp(big.NewFloat(0)) > 0 {
		for id, weight := range allocations {
			normalized := new(big.Float).Quo(weight, totalWeight)
			allocations[id] = normalized
		}
	}

	return allocations, nil
}

// calculateStrategyWeight calculates the weight for a strategy based on risk and APY
func (ysm *YieldStrategyManager) calculateStrategyWeight(strategy *YieldStrategy, riskTolerance StrategyRiskLevel, targetAPY *big.Float) *big.Float {
	weight := big.NewFloat(1.0)

	// Risk adjustment
	switch riskTolerance {
	case StrategyRiskLevelLow:
		if strategy.RiskLevel == StrategyRiskLevelHigh || strategy.RiskLevel == StrategyRiskLevelVeryHigh {
			weight = new(big.Float).Mul(weight, big.NewFloat(0.1))
		}
	case StrategyRiskLevelMedium:
		if strategy.RiskLevel == StrategyRiskLevelVeryHigh {
			weight = new(big.Float).Mul(weight, big.NewFloat(0.3))
		}
	case StrategyRiskLevelHigh:
		// No risk penalty for high risk tolerance
	case StrategyRiskLevelVeryHigh:
		// No risk penalty for very high risk tolerance
	}

	// APY adjustment - only apply if risk tolerance allows it
	if strategy.APY != nil && targetAPY != nil {
		// Check if risk tolerance allows APY bonuses
		allowAPYBonus := false
		switch riskTolerance {
		case StrategyRiskLevelLow:
			// Low risk tolerance only allows bonuses for low/medium risk strategies
			allowAPYBonus = strategy.RiskLevel == StrategyRiskLevelLow || strategy.RiskLevel == StrategyRiskLevelMedium
		case StrategyRiskLevelMedium:
			// Medium risk tolerance allows bonuses for low/medium risk strategies only
			allowAPYBonus = strategy.RiskLevel == StrategyRiskLevelLow || strategy.RiskLevel == StrategyRiskLevelMedium
		case StrategyRiskLevelHigh, StrategyRiskLevelVeryHigh:
			// High risk tolerance allows bonuses for all strategies
			allowAPYBonus = true
		}

		// Only apply APY adjustment if risk tolerance allows it
		if allowAPYBonus {
			apyRatio := new(big.Float).Quo(strategy.APY, targetAPY)
			if apyRatio.Cmp(big.NewFloat(1.0)) > 0 {
				weight = new(big.Float).Mul(weight, apyRatio)
			}
		}
	}

	return weight
}

// UpdateAnalytics updates analytics for a strategy
func (ysm *YieldStrategyManager) UpdateAnalytics(strategyID string, apy *big.Float, tvl *big.Int) error {
	ysm.analytics.mu.Lock()
	defer ysm.analytics.mu.Unlock()

	dataPoint := APYDataPoint{
		Timestamp: time.Now(),
		APY:       apy,
		TVL:       tvl,
	}

	ysm.analytics.historicalAPY[strategyID] = append(ysm.analytics.historicalAPY[strategyID], dataPoint)

	// Keep only last 1000 data points
	if len(ysm.analytics.historicalAPY[strategyID]) > 1000 {
		ysm.analytics.historicalAPY[strategyID] = ysm.analytics.historicalAPY[strategyID][1:]
	}

	// Update risk metrics
	ysm.updateRiskMetrics(strategyID)

	return nil
}

// updateRiskMetrics updates risk metrics for a strategy
func (ysm *YieldStrategyManager) updateRiskMetrics(strategyID string) {
	data := ysm.analytics.historicalAPY[strategyID]
	if len(data) < 2 {
		return
	}

	// Calculate volatility
	var returns []*big.Float
	for i := 1; i < len(data); i++ {
		if data[i-1].APY.Cmp(big.NewFloat(0)) > 0 {
			returnRate := new(big.Float).Quo(data[i].APY, data[i-1].APY)
			returns = append(returns, returnRate)
		}
	}

	if len(returns) == 0 {
		return
	}

	// Calculate mean return
	meanReturn := big.NewFloat(0)
	for _, ret := range returns {
		meanReturn.Add(meanReturn, ret)
	}
	meanReturn.Quo(meanReturn, big.NewFloat(float64(len(returns))))

	// Calculate variance
	variance := big.NewFloat(0)
	for _, ret := range returns {
		diff := new(big.Float).Sub(ret, meanReturn)
		diffSquared := new(big.Float).Mul(diff, diff)
		variance.Add(variance, diffSquared)
	}
	variance.Quo(variance, big.NewFloat(float64(len(returns))))

	// Calculate volatility (square root of variance)
	volatility := new(big.Float).Sqrt(variance)

	// Calculate max drawdown
	maxDrawdown := ysm.calculateMaxDrawdown(data)

	// Calculate Sharpe ratio (simplified)
	riskFreeRate := big.NewFloat(0.02) // 2% risk-free rate
	sharpeRatio := big.NewFloat(0)
	if volatility.Cmp(big.NewFloat(0)) > 0 {
		excessReturn := new(big.Float).Sub(meanReturn, riskFreeRate)
		sharpeRatio.Quo(excessReturn, volatility)
	}

	ysm.analytics.riskMetrics[strategyID] = &RiskMetrics{
		Volatility:    volatility,
		MaxDrawdown:   maxDrawdown,
		SharpeRatio:   sharpeRatio,
		ExpectedValue: meanReturn,
	}
}

// calculateMaxDrawdown calculates the maximum drawdown from peak
func (ysm *YieldStrategyManager) calculateMaxDrawdown(data []APYDataPoint) *big.Float {
	if len(data) == 0 {
		return big.NewFloat(0)
	}

	peak := data[0].APY
	maxDrawdown := big.NewFloat(0)

	for _, point := range data {
		if point.APY.Cmp(peak) > 0 {
			peak = point.APY
		} else {
			drawdown := new(big.Float).Sub(peak, point.APY)
			if drawdown.Cmp(maxDrawdown) > 0 {
				maxDrawdown = drawdown
			}
		}
	}

	return maxDrawdown
}

// GetRiskMetrics returns risk metrics for a strategy
func (ysm *YieldStrategyManager) GetRiskMetrics(strategyID string) (*RiskMetrics, error) {
	ysm.analytics.mu.RLock()
	defer ysm.analytics.mu.RUnlock()

	metrics, exists := ysm.analytics.riskMetrics[strategyID]
	if !exists {
		return nil, fmt.Errorf("no risk metrics available for strategy %s", strategyID)
	}

	return metrics, nil
}

// GetHistoricalAPY returns historical APY data for a strategy
func (ysm *YieldStrategyManager) GetHistoricalAPY(strategyID string, limit int) ([]APYDataPoint, error) {
	ysm.analytics.mu.RLock()
	defer ysm.analytics.mu.RUnlock()

	data, exists := ysm.analytics.historicalAPY[strategyID]
	if !exists {
		return nil, fmt.Errorf("no historical data available for strategy %s", strategyID)
	}

	if limit > 0 && limit < len(data) {
		start := len(data) - limit
		return data[start:], nil
	}

	return data, nil
}
