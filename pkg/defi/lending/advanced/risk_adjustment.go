package advanced

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// RiskAdjustmentStrategy represents a risk adjustment strategy
type RiskAdjustmentStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RiskAdjustmentType     `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      RiskAdjustmentStatus   `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RiskAdjustmentType represents the type of risk adjustment
type RiskAdjustmentType string

const (
	RiskAdjustmentTypeSharpeRatio    RiskAdjustmentType = "sharpe_ratio"
	RiskAdjustmentTypeSortinoRatio   RiskAdjustmentType = "sortino_ratio"
	RiskAdjustmentTypeCalmarRatio    RiskAdjustmentType = "calmar_ratio"
	RiskAdjustmentTypeInformationRatio RiskAdjustmentType = "information_ratio"
	RiskAdjustmentTypeCustom         RiskAdjustmentType = "custom"
)

// RiskAdjustmentStatus represents the status of a risk adjustment strategy
type RiskAdjustmentStatus string

const (
	RiskAdjustmentStatusActive   RiskAdjustmentStatus = "active"
	RiskAdjustmentStatusPaused   RiskAdjustmentStatus = "paused"
	RiskAdjustmentStatusStopped  RiskAdjustmentStatus = "stopped"
)

// RiskAdjustedMetrics represents risk-adjusted performance metrics
type RiskAdjustedMetrics struct {
	StrategyID        string     `json:"strategy_id"`
	SharpeRatio       *big.Float `json:"sharpe_ratio"`
	SortinoRatio      *big.Float `json:"sortino_ratio"`
	CalmarRatio       *big.Float `json:"calmar_ratio"`
	InformationRatio  *big.Float `json:"information_ratio"`
	TreynorRatio      *big.Float `json:"treynor_ratio"`
	JensenAlpha      *big.Float `json:"jensen_alpha"`
	RiskAdjustedReturn *big.Float `json:"risk_adjusted_return"`
	RiskScore         *big.Float `json:"risk_score"`
	LastCalculated    time.Time  `json:"last_calculated"`
}

// RiskProfile represents a comprehensive risk profile
type RiskProfile struct {
	UserID            string                 `json:"user_id"`
	RiskTolerance     RiskToleranceLevel     `json:"risk_tolerance"`
	InvestmentHorizon time.Duration          `json:"investment_horizon"`
	LiquidityNeeds    LiquidityRequirement   `json:"liquidity_needs"`
	RiskConstraints   map[string]*big.Float  `json:"risk_constraints"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// RiskToleranceLevel represents user risk tolerance
type RiskToleranceLevel string

const (
	RiskToleranceConservative RiskToleranceLevel = "conservative"
	RiskToleranceModerate    RiskToleranceLevel = "moderate"
	RiskToleranceAggressive  RiskToleranceLevel = "aggressive"
)

// LiquidityRequirement represents liquidity requirements
type LiquidityRequirement string

const (
	LiquidityRequirementHigh   LiquidityRequirement = "high"
	LiquidityRequirementMedium LiquidityRequirement = "medium"
	LiquidityRequirementLow    LiquidityRequirement = "low"
)

// RiskAdjustmentEngine manages risk adjustment operations
type RiskAdjustmentEngine struct {
	strategies map[string]*RiskAdjustmentStrategy
	profiles   map[string]*RiskProfile
	metrics    map[string]*RiskAdjustedMetrics
	mu         sync.RWMutex
	ysm        *YieldStrategyManager
}

// NewRiskAdjustmentEngine creates a new risk adjustment engine
func NewRiskAdjustmentEngine(ysm *YieldStrategyManager) *RiskAdjustmentEngine {
	return &RiskAdjustmentEngine{
		strategies: make(map[string]*RiskAdjustmentStrategy),
		profiles:   make(map[string]*RiskProfile),
		metrics:    make(map[string]*RiskAdjustedMetrics),
		ysm:        ysm,
	}
}

// RegisterRiskAdjustmentStrategy registers a new risk adjustment strategy
func (rae *RiskAdjustmentEngine) RegisterRiskAdjustmentStrategy(strategy *RiskAdjustmentStrategy) error {
	rae.mu.Lock()
	defer rae.mu.Unlock()

	if strategy.ID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	if _, exists := rae.strategies[strategy.ID]; exists {
		return fmt.Errorf("risk adjustment strategy with ID %s already exists", strategy.ID)
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	rae.strategies[strategy.ID] = strategy

	return nil
}

// GetRiskAdjustmentStrategy retrieves a risk adjustment strategy by ID
func (rae *RiskAdjustmentEngine) GetRiskAdjustmentStrategy(strategyID string) (*RiskAdjustmentStrategy, error) {
	rae.mu.RLock()
	defer rae.mu.RUnlock()

	strategy, exists := rae.strategies[strategyID]
	if !exists {
		return nil, fmt.Errorf("risk adjustment strategy with ID %s not found", strategyID)
	}

	return strategy, nil
}

// RegisterRiskProfile registers a new risk profile
func (rae *RiskAdjustmentEngine) RegisterRiskProfile(profile *RiskProfile) error {
	rae.mu.Lock()
	defer rae.mu.Unlock()

	if profile.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	if _, exists := rae.profiles[profile.UserID]; exists {
		return fmt.Errorf("risk profile for user %s already exists", profile.UserID)
	}

	profile.LastUpdated = time.Now()
	rae.profiles[profile.UserID] = profile

	return nil
}

// GetRiskProfile retrieves a risk profile by user ID
func (rae *RiskAdjustmentEngine) GetRiskProfile(userID string) (*RiskProfile, error) {
	rae.mu.RLock()
	defer rae.mu.RUnlock()

	profile, exists := rae.profiles[userID]
	if !exists {
		return nil, fmt.Errorf("risk profile for user %s not found", userID)
	}

	return profile, nil
}

// CalculateRiskAdjustedMetrics calculates risk-adjusted metrics for a strategy
func (rae *RiskAdjustmentEngine) CalculateRiskAdjustedMetrics(ctx context.Context, strategyID string) (*RiskAdjustedMetrics, error) {
	// Get strategy risk metrics
	riskMetrics, err := rae.ysm.GetRiskMetrics(strategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk metrics: %w", err)
	}

	// Get strategy APY
	apy, err := rae.ysm.GetStrategyAPY(ctx, strategyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy APY: %w", err)
	}

	metrics := &RiskAdjustedMetrics{
		StrategyID:     strategyID,
		LastCalculated: time.Now(),
	}

	// Calculate Sharpe Ratio
	metrics.SharpeRatio = rae.calculateSharpeRatio(apy, riskMetrics.Volatility)

	// Calculate Sortino Ratio
	metrics.SortinoRatio = rae.calculateSortinoRatio(apy, riskMetrics.MaxDrawdown)

	// Calculate Calmar Ratio
	metrics.CalmarRatio = rae.calculateCalmarRatio(apy, riskMetrics.MaxDrawdown)

	// Calculate Information Ratio (simplified)
	metrics.InformationRatio = rae.calculateInformationRatio(apy, riskMetrics.Volatility)

	// Calculate Treynor Ratio (simplified)
	metrics.TreynorRatio = rae.calculateTreynorRatio(apy, riskMetrics.Volatility)

	// Calculate Jensen Alpha (simplified)
	metrics.JensenAlpha = rae.calculateJensenAlpha(apy, riskMetrics.Volatility)

	// Calculate Risk-Adjusted Return
	metrics.RiskAdjustedReturn = rae.calculateRiskAdjustedReturn(apy, riskMetrics.Volatility)

	// Calculate overall risk score
	metrics.RiskScore = rae.calculateRiskScore(riskMetrics)

	// Store metrics
	rae.mu.Lock()
	rae.metrics[strategyID] = metrics
	rae.mu.Unlock()

	return metrics, nil
}

// calculateSharpeRatio calculates the Sharpe ratio
func (rae *RiskAdjustmentEngine) calculateSharpeRatio(apy, volatility *big.Float) *big.Float {
	if volatility == nil || volatility.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Risk-free rate (simplified)
	riskFreeRate := big.NewFloat(0.02) // 2%

	// Excess return
	excessReturn := new(big.Float).Sub(apy, riskFreeRate)

	// Sharpe ratio = excess return / volatility
	sharpeRatio := new(big.Float).Quo(excessReturn, volatility)
	return sharpeRatio
}

// calculateSortinoRatio calculates the Sortino ratio
func (rae *RiskAdjustmentEngine) calculateSortinoRatio(apy, maxDrawdown *big.Float) *big.Float {
	if maxDrawdown == nil || maxDrawdown.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Risk-free rate
	riskFreeRate := big.NewFloat(0.02) // 2%

	// Excess return
	excessReturn := new(big.Float).Sub(apy, riskFreeRate)

	// Sortino ratio = excess return / downside deviation (using max drawdown as proxy)
	sortinoRatio := new(big.Float).Quo(excessReturn, maxDrawdown)
	return sortinoRatio
}

// calculateCalmarRatio calculates the Calmar ratio
func (rae *RiskAdjustmentEngine) calculateCalmarRatio(apy, maxDrawdown *big.Float) *big.Float {
	if maxDrawdown == nil || maxDrawdown.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Calmar ratio = annual return / max drawdown
	calmarRatio := new(big.Float).Quo(apy, maxDrawdown)
	return calmarRatio
}

// calculateInformationRatio calculates the Information ratio
func (rae *RiskAdjustmentEngine) calculateInformationRatio(apy, volatility *big.Float) *big.Float {
	if volatility == nil || volatility.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Benchmark return (simplified)
	benchmarkReturn := big.NewFloat(0.08) // 8%

	// Active return
	activeReturn := new(big.Float).Sub(apy, benchmarkReturn)

	// Information ratio = active return / tracking error (using volatility as proxy)
	informationRatio := new(big.Float).Quo(activeReturn, volatility)
	return informationRatio
}

// calculateTreynorRatio calculates the Treynor ratio
func (rae *RiskAdjustmentEngine) calculateTreynorRatio(apy, volatility *big.Float) *big.Float {
	if volatility == nil || volatility.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Risk-free rate
	riskFreeRate := big.NewFloat(0.02) // 2%

	// Excess return
	excessReturn := new(big.Float).Sub(apy, riskFreeRate)

	// Treynor ratio = excess return / beta (using volatility as proxy)
	treynorRatio := new(big.Float).Quo(excessReturn, volatility)
	return treynorRatio
}

// calculateJensenAlpha calculates the Jensen Alpha
func (rae *RiskAdjustmentEngine) calculateJensenAlpha(apy, volatility *big.Float) *big.Float {
	if volatility == nil || volatility.Cmp(big.NewFloat(0)) == 0 {
		return big.NewFloat(0)
	}

	// Risk-free rate
	riskFreeRate := big.NewFloat(0.02) // 2%

	// Market return (simplified)
	marketReturn := big.NewFloat(0.10) // 10%

	// Beta (simplified, using volatility ratio)
	beta := new(big.Float).Quo(volatility, big.NewFloat(0.15)) // Assuming market volatility of 15%

	// Expected return based on CAPM
	expectedReturn := new(big.Float).Add(riskFreeRate, new(big.Float).Mul(beta, new(big.Float).Sub(marketReturn, riskFreeRate)))

	// Jensen Alpha = actual return - expected return
	jensenAlpha := new(big.Float).Sub(apy, expectedReturn)
	return jensenAlpha
}

// calculateRiskAdjustedReturn calculates the risk-adjusted return
func (rae *RiskAdjustmentEngine) calculateRiskAdjustedReturn(apy, volatility *big.Float) *big.Float {
	if volatility == nil || volatility.Cmp(big.NewFloat(0)) == 0 {
		return apy
	}

	// Risk-adjusted return = return / (1 + risk)
	// Using volatility as risk measure
	riskAdjustedReturn := new(big.Float).Quo(apy, new(big.Float).Add(big.NewFloat(1), volatility))
	return riskAdjustedReturn
}

// calculateRiskScore calculates an overall risk score
func (rae *RiskAdjustmentEngine) calculateRiskScore(riskMetrics *RiskMetrics) *big.Float {
	// Simple weighted average of risk metrics
	score := big.NewFloat(0)
	weight := big.NewFloat(0.25) // Equal weights

	// Volatility component
	if riskMetrics.Volatility != nil {
		volatilityScore := new(big.Float).Mul(riskMetrics.Volatility, weight)
		score.Add(score, volatilityScore)
	}

	// Max drawdown component
	if riskMetrics.MaxDrawdown != nil {
		drawdownScore := new(big.Float).Mul(riskMetrics.MaxDrawdown, weight)
		score.Add(score, drawdownScore)
	}

	// Sharpe ratio component (inverted, so lower is riskier)
	if riskMetrics.SharpeRatio != nil {
		sharpeScore := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(1), riskMetrics.SharpeRatio), weight)
		score.Add(score, sharpeScore)
	}

	// Expected value component (inverted, so lower is riskier)
	if riskMetrics.ExpectedValue != nil {
		expectedValueScore := new(big.Float).Mul(new(big.Float).Quo(big.NewFloat(1), riskMetrics.ExpectedValue), weight)
		score.Add(score, expectedValueScore)
	}

	return score
}

// GetRiskAdjustedMetrics retrieves risk-adjusted metrics for a strategy
func (rae *RiskAdjustmentEngine) GetRiskAdjustedMetrics(strategyID string) (*RiskAdjustedMetrics, error) {
	rae.mu.RLock()
	defer rae.mu.RUnlock()

	metrics, exists := rae.metrics[strategyID]
	if !exists {
		return nil, fmt.Errorf("no risk-adjusted metrics available for strategy %s", strategyID)
	}

	return metrics, nil
}

// CalculatePortfolioRiskAdjustedMetrics calculates risk-adjusted metrics for a portfolio
func (rae *RiskAdjustmentEngine) CalculatePortfolioRiskAdjustedMetrics(ctx context.Context, portfolio *Portfolio) (*RiskAdjustedMetrics, error) {
	if len(portfolio.Strategies) == 0 {
		return nil, errors.New("portfolio has no strategies")
	}

	portfolioMetrics := &RiskAdjustedMetrics{
		StrategyID:     "portfolio",
		LastCalculated: time.Now(),
	}

	// Calculate weighted averages across all strategies
	var totalWeightedSharpe, totalWeightedSortino, totalWeightedCalmar, totalWeightedReturn *big.Float
	var totalWeight *big.Float

	totalWeightedSharpe = big.NewFloat(0)
	totalWeightedSortino = big.NewFloat(0)
	totalWeightedCalmar = big.NewFloat(0)
	totalWeightedReturn = big.NewFloat(0)
	totalWeight = big.NewFloat(0)

	for strategyID, allocation := range portfolio.Strategies {
		if allocation.TargetWeight == nil || allocation.TargetWeight.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		// Get strategy metrics
		strategyMetrics, err := rae.CalculateRiskAdjustedMetrics(ctx, strategyID)
		if err != nil {
			continue // Skip strategies with errors
		}

		// Weighted metrics
		if strategyMetrics.SharpeRatio != nil {
			weightedSharpe := new(big.Float).Mul(strategyMetrics.SharpeRatio, allocation.TargetWeight)
			totalWeightedSharpe.Add(totalWeightedSharpe, weightedSharpe)
		}

		if strategyMetrics.SortinoRatio != nil {
			weightedSortino := new(big.Float).Mul(strategyMetrics.SortinoRatio, allocation.TargetWeight)
			totalWeightedSortino.Add(totalWeightedSortino, weightedSortino)
		}

		if strategyMetrics.CalmarRatio != nil {
			weightedCalmar := new(big.Float).Mul(strategyMetrics.CalmarRatio, allocation.TargetWeight)
			totalWeightedCalmar.Add(totalWeightedCalmar, weightedCalmar)
		}

		if strategyMetrics.RiskAdjustedReturn != nil {
			weightedReturn := new(big.Float).Mul(strategyMetrics.RiskAdjustedReturn, allocation.TargetWeight)
			totalWeightedReturn.Add(totalWeightedReturn, weightedReturn)
		}

		totalWeight.Add(totalWeight, allocation.TargetWeight)
	}

	// Calculate portfolio-level metrics
	if totalWeight.Cmp(big.NewFloat(0)) > 0 {
		portfolioMetrics.SharpeRatio = new(big.Float).Quo(totalWeightedSharpe, totalWeight)
		portfolioMetrics.SortinoRatio = new(big.Float).Quo(totalWeightedSortino, totalWeight)
		portfolioMetrics.CalmarRatio = new(big.Float).Quo(totalWeightedCalmar, totalWeight)
		portfolioMetrics.RiskAdjustedReturn = new(big.Float).Quo(totalWeightedReturn, totalWeight)
	} else {
		portfolioMetrics.SharpeRatio = big.NewFloat(0)
		portfolioMetrics.SortinoRatio = big.NewFloat(0)
		portfolioMetrics.CalmarRatio = big.NewFloat(0)
		portfolioMetrics.RiskAdjustedReturn = big.NewFloat(0)
	}

	// Calculate portfolio risk score
	portfolioMetrics.RiskScore = rae.calculatePortfolioRiskScore(portfolio)

	return portfolioMetrics, nil
}

// calculatePortfolioRiskScore calculates risk score for a portfolio
func (rae *RiskAdjustmentEngine) calculatePortfolioRiskScore(portfolio *Portfolio) *big.Float {
	// Simple portfolio risk score based on strategy weights and individual risk scores
	portfolioRiskScore := big.NewFloat(0)
	totalWeight := big.NewFloat(0)

	for strategyID, allocation := range portfolio.Strategies {
		if allocation.TargetWeight == nil || allocation.TargetWeight.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		// Get strategy risk metrics
		riskMetrics, err := rae.ysm.GetRiskMetrics(strategyID)
		if err != nil {
			continue
		}

		// Weighted risk score
		if riskMetrics.Volatility != nil {
			weightedRisk := new(big.Float).Mul(riskMetrics.Volatility, allocation.TargetWeight)
			portfolioRiskScore.Add(portfolioRiskScore, weightedRisk)
		}

		totalWeight.Add(totalWeight, allocation.TargetWeight)
	}

	if totalWeight.Cmp(big.NewFloat(0)) > 0 {
		return new(big.Float).Quo(portfolioRiskScore, totalWeight)
	}

	return big.NewFloat(0)
}

// OptimizePortfolioRisk optimizes portfolio allocation based on risk constraints
func (rae *RiskAdjustmentEngine) OptimizePortfolioRisk(ctx context.Context, userID string, targetReturn *big.Float) (map[string]*big.Float, error) {
	// Get user risk profile
	riskProfile, err := rae.GetRiskProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk profile: %w", err)
	}

	// Get available strategies
	strategies := rae.ysm.ListStrategies()
	if len(strategies) == 0 {
		return nil, errors.New("no strategies available for optimization")
	}

	// Calculate optimal allocation based on risk tolerance
	var allocations map[string]*big.Float

	switch riskProfile.RiskTolerance {
	case RiskToleranceConservative:
		allocations, err = rae.optimizeConservativePortfolio(ctx, strategies, targetReturn)
	case RiskToleranceModerate:
		allocations, err = rae.optimizeModeratePortfolio(ctx, strategies, targetReturn)
	case RiskToleranceAggressive:
		allocations, err = rae.optimizeAggressivePortfolio(ctx, strategies, targetReturn)
	default:
		allocations, err = rae.optimizeModeratePortfolio(ctx, strategies, targetReturn)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to optimize portfolio: %w", err)
	}

	return allocations, nil
}

// optimizeConservativePortfolio optimizes portfolio for conservative risk tolerance
func (rae *RiskAdjustmentEngine) optimizeConservativePortfolio(ctx context.Context, strategies []*YieldStrategy, targetReturn *big.Float) (map[string]*big.Float, error) {
	allocations := make(map[string]*big.Float)
	
	// Conservative approach: favor low-risk strategies
	var lowRiskStrategies []*YieldStrategy
	var mediumRiskStrategies []*YieldStrategy
	
	for _, strategy := range strategies {
		if strategy.Status != StrategyStatusActive {
			continue
		}
		
		switch strategy.RiskLevel {
		case StrategyRiskLevelLow:
			lowRiskStrategies = append(lowRiskStrategies, strategy)
		case StrategyRiskLevelMedium:
			mediumRiskStrategies = append(mediumRiskStrategies, strategy)
		}
	}
	
	// Allocate 70% to low-risk, 30% to medium-risk
	totalLowRiskWeight := big.NewFloat(0.7)
	totalMediumRiskWeight := big.NewFloat(0.3)
	
	// Distribute among low-risk strategies
	if len(lowRiskStrategies) > 0 {
		weightPerStrategy := new(big.Float).Quo(totalLowRiskWeight, big.NewFloat(float64(len(lowRiskStrategies))))
		for _, strategy := range lowRiskStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	// Distribute among medium-risk strategies
	if len(mediumRiskStrategies) > 0 {
		weightPerStrategy := new(big.Float).Quo(totalMediumRiskWeight, big.NewFloat(float64(len(mediumRiskStrategies))))
		for _, strategy := range mediumRiskStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	return allocations, nil
}

// optimizeModeratePortfolio optimizes portfolio for moderate risk tolerance
func (rae *RiskAdjustmentEngine) optimizeModeratePortfolio(ctx context.Context, strategies []*YieldStrategy, targetReturn *big.Float) (map[string]*big.Float, error) {
	allocations := make(map[string]*big.Float)
	
	// Moderate approach: balanced allocation
	var lowRiskStrategies []*YieldStrategy
	var mediumRiskStrategies []*YieldStrategy
	var highRiskStrategies []*YieldStrategy
	
	for _, strategy := range strategies {
		if strategy.Status != StrategyStatusActive {
			continue
		}
		
		switch strategy.RiskLevel {
		case StrategyRiskLevelLow:
			lowRiskStrategies = append(lowRiskStrategies, strategy)
		case StrategyRiskLevelMedium:
			mediumRiskStrategies = append(mediumRiskStrategies, strategy)
		case StrategyRiskLevelHigh:
			highRiskStrategies = append(highRiskStrategies, strategy)
		}
	}
	
	// Allocate 40% to low-risk, 40% to medium-risk, 20% to high-risk
	totalLowRiskWeight := big.NewFloat(0.4)
	totalMediumRiskWeight := big.NewFloat(0.4)
	totalHighRiskWeight := big.NewFloat(0.2)
	
	// Distribute weights
	if len(lowRiskStrategies) > 0 {
		weightPerStrategy := new(big.Float).Quo(totalLowRiskWeight, big.NewFloat(float64(len(lowRiskStrategies))))
		for _, strategy := range lowRiskStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	if len(mediumRiskStrategies) > 0 {
		weightPerStrategy := new(big.Float).Quo(totalMediumRiskWeight, big.NewFloat(float64(len(mediumRiskStrategies))))
		for _, strategy := range mediumRiskStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	if len(highRiskStrategies) > 0 {
		weightPerStrategy := new(big.Float).Quo(totalHighRiskWeight, big.NewFloat(float64(len(highRiskStrategies))))
		for _, strategy := range highRiskStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	return allocations, nil
}

// optimizeAggressivePortfolio optimizes portfolio for aggressive risk tolerance
func (rae *RiskAdjustmentEngine) optimizeAggressivePortfolio(ctx context.Context, strategies []*YieldStrategy, targetReturn *big.Float) (map[string]*big.Float, error) {
	allocations := make(map[string]*big.Float)
	
	// Aggressive approach: favor high-return strategies
	var allStrategies []*YieldStrategy
	
	for _, strategy := range strategies {
		if strategy.Status != StrategyStatusActive {
			continue
		}
		allStrategies = append(allStrategies, strategy)
	}
	
	// Sort strategies by APY (descending)
	// In a real implementation, you'd want proper sorting
	// For now, use equal weights but favor strategies with higher APY
	
	if len(allStrategies) > 0 {
		// Equal weight distribution
		weightPerStrategy := new(big.Float).Quo(big.NewFloat(1.0), big.NewFloat(float64(len(allStrategies))))
		for _, strategy := range allStrategies {
			allocations[strategy.ID] = weightPerStrategy
		}
	}
	
	return allocations, nil
}

// GetRiskAdjustedRecommendations provides risk-adjusted investment recommendations
func (rae *RiskAdjustmentEngine) GetRiskAdjustedRecommendations(ctx context.Context, userID string) ([]*InvestmentRecommendation, error) {
	// Get user risk profile
	riskProfile, err := rae.GetRiskProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk profile: %w", err)
	}

	// Get available strategies
	strategies := rae.ysm.ListStrategies()
	if len(strategies) == 0 {
		return nil, errors.New("no strategies available")
	}

	var recommendations []*InvestmentRecommendation

	for _, strategy := range strategies {
		if strategy.Status != StrategyStatusActive {
			continue
		}

		// Calculate risk-adjusted metrics
		metrics, err := rae.CalculateRiskAdjustedMetrics(ctx, strategy.ID)
		if err != nil {
			continue // Skip strategies with errors
		}

		// Generate recommendation based on risk profile and metrics
		recommendation := rae.generateRecommendation(strategy, metrics, riskProfile)
		if recommendation != nil {
			recommendations = append(recommendations, recommendation)
		}
	}

	// Sort recommendations by score (descending)
	// In a real implementation, you'd want proper sorting
	// For now, just return the recommendations

	return recommendations, nil
}

// InvestmentRecommendation represents an investment recommendation
type InvestmentRecommendation struct {
	StrategyID        string     `json:"strategy_id"`
	StrategyName      string     `json:"strategy_name"`
	Recommendation    string     `json:"recommendation"`
	Score             *big.Float `json:"score"`
	RiskLevel         string     `json:"risk_level"`
	ExpectedReturn    *big.Float `json:"expected_return"`
	RiskAdjustedReturn *big.Float `json:"risk_adjusted_return"`
	Reasoning         string     `json:"reasoning"`
}

// generateRecommendation generates an investment recommendation
func (rae *RiskAdjustmentEngine) generateRecommendation(strategy *YieldStrategy, metrics *RiskAdjustedMetrics, profile *RiskProfile) *InvestmentRecommendation {
	recommendation := &InvestmentRecommendation{
		StrategyID:        strategy.ID,
		StrategyName:      strategy.Name,
		RiskLevel:         string(strategy.RiskLevel),
		ExpectedReturn:    strategy.APY,
		RiskAdjustedReturn: metrics.RiskAdjustedReturn,
	}

	// Calculate recommendation score
	score := rae.calculateRecommendationScore(strategy, metrics, profile)
	recommendation.Score = score

	// Generate recommendation text
	recommendation.Recommendation = rae.generateRecommendationText(score, strategy.RiskLevel, profile.RiskTolerance)
	recommendation.Reasoning = rae.generateRecommendationReasoning(strategy, metrics, profile)

	return recommendation
}

// calculateRecommendationScore calculates a recommendation score
func (rae *RiskAdjustmentEngine) calculateRecommendationScore(strategy *YieldStrategy, metrics *RiskAdjustedMetrics, profile *RiskProfile) *big.Float {
	score := big.NewFloat(0)

	// APY component (40% weight)
	if strategy.APY != nil {
		apyScore := new(big.Float).Mul(strategy.APY, big.NewFloat(0.4))
		score.Add(score, apyScore)
	}

	// Risk-adjusted return component (30% weight)
	if metrics.RiskAdjustedReturn != nil {
		rarScore := new(big.Float).Mul(metrics.RiskAdjustedReturn, big.NewFloat(0.3))
		score.Add(score, rarScore)
	}

	// Sharpe ratio component (20% weight)
	if metrics.SharpeRatio != nil {
		sharpeScore := new(big.Float).Mul(metrics.SharpeRatio, big.NewFloat(0.2))
		score.Add(score, sharpeScore)
	}

	// Risk compatibility component (10% weight)
	riskCompatibility := rae.calculateRiskCompatibility(strategy.RiskLevel, profile.RiskTolerance)
	score.Add(score, new(big.Float).Mul(riskCompatibility, big.NewFloat(0.1)))

	return score
}

// calculateRiskCompatibility calculates risk compatibility score
func (rae *RiskAdjustmentEngine) calculateRiskCompatibility(strategyRisk StrategyRiskLevel, userTolerance RiskToleranceLevel) *big.Float {
	// Risk compatibility matrix
	compatibilityMatrix := map[StrategyRiskLevel]map[RiskToleranceLevel]*big.Float{
		StrategyRiskLevelLow: {
			RiskToleranceConservative: big.NewFloat(1.0),
			RiskToleranceModerate:     big.NewFloat(0.8),
			RiskToleranceAggressive:   big.NewFloat(0.6),
		},
		StrategyRiskLevelMedium: {
			RiskToleranceConservative: big.NewFloat(0.7),
			RiskToleranceModerate:     big.NewFloat(1.0),
			RiskToleranceAggressive:   big.NewFloat(0.8),
		},
		StrategyRiskLevelHigh: {
			RiskToleranceConservative: big.NewFloat(0.4),
			RiskToleranceModerate:     big.NewFloat(0.7),
			RiskToleranceAggressive:   big.NewFloat(1.0),
		},
		StrategyRiskLevelVeryHigh: {
			RiskToleranceConservative: big.NewFloat(0.2),
			RiskToleranceModerate:     big.NewFloat(0.5),
			RiskToleranceAggressive:   big.NewFloat(0.8),
		},
	}

	if compatibility, exists := compatibilityMatrix[strategyRisk]; exists {
		if score, exists := compatibility[userTolerance]; exists {
			return score
		}
	}

	return big.NewFloat(0.5) // Default compatibility
}

// generateRecommendationText generates recommendation text
func (rae *RiskAdjustmentEngine) generateRecommendationText(score *big.Float, strategyRisk StrategyRiskLevel, userTolerance RiskToleranceLevel) string {
	if score.Cmp(big.NewFloat(0.8)) > 0 {
		return "Strong Buy"
	} else if score.Cmp(big.NewFloat(0.6)) > 0 {
		return "Buy"
	} else if score.Cmp(big.NewFloat(0.4)) > 0 {
		return "Hold"
	} else if score.Cmp(big.NewFloat(0.2)) > 0 {
		return "Sell"
	} else {
		return "Strong Sell"
	}
}

// generateRecommendationReasoning generates reasoning for recommendation
func (rae *RiskAdjustmentEngine) generateRecommendationReasoning(strategy *YieldStrategy, metrics *RiskAdjustedMetrics, profile *RiskProfile) string {
	var reasons []string

	if strategy.APY != nil && strategy.APY.Cmp(big.NewFloat(0.1)) > 0 {
		reasons = append(reasons, "High APY")
	}

	if metrics.SharpeRatio != nil && metrics.SharpeRatio.Cmp(big.NewFloat(1.0)) > 0 {
		reasons = append(reasons, "Good risk-adjusted returns")
	}

	if strategy.RiskLevel == StrategyRiskLevelLow && profile.RiskTolerance == RiskToleranceConservative {
		reasons = append(reasons, "Risk level matches profile")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Moderate performance")
	}

	return fmt.Sprintf("Recommendation based on: %s", reasons[0])
}
