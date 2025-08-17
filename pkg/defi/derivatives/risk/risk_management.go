package risk

import (
	"errors"
	"math"
	"math/big"
	"time"
)

// RiskMetric represents a risk measurement
type RiskMetric struct {
	Value     *big.Float
	Type      MetricType
	Timestamp time.Time
	Confidence *big.Float
}

// MetricType represents the type of risk metric
type MetricType int

const (
	VaR MetricType = iota
	CVaR
	Volatility
	SharpeRatio
	MaxDrawdown
	Beta
	Correlation
)

// Portfolio represents a collection of positions for risk analysis
type Portfolio struct {
	ID          string
	Positions   map[string]*Position
	RiskMetrics map[MetricType]*RiskMetric
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Position represents a position in the portfolio
type Position struct {
	ID           string
	AssetID      string
	Size         *big.Float
	Price        *big.Float
	Value        *big.Float
	Weight       *big.Float
	Returns      []*big.Float
	Volatility   *big.Float
	Beta         *big.Float
	UpdatedAt    time.Time
}

// MarketData represents historical market data for risk calculations
type MarketData struct {
	AssetID    string
	Prices     []*big.Float
	Returns    []*big.Float
	Timestamps []time.Time
}

// RiskManager handles portfolio risk calculations and management
type RiskManager struct {
	Portfolios map[string]*Portfolio
	MarketData map[string]*MarketData
	RiskFreeRate *big.Float
	ConfidenceLevel *big.Float
}

// NewRiskManager creates a new risk manager
func NewRiskManager(riskFreeRate, confidenceLevel *big.Float) (*RiskManager, error) {
	if riskFreeRate == nil || confidenceLevel == nil {
		return nil, errors.New("risk-free rate and confidence level cannot be nil")
	}
	
	if confidenceLevel.Cmp(big.NewFloat(0)) <= 0 || confidenceLevel.Cmp(big.NewFloat(1)) >= 0 {
		return nil, errors.New("confidence level must be between 0 and 1")
	}
	
	return &RiskManager{
		Portfolios:     make(map[string]*Portfolio),
		MarketData:     make(map[string]*MarketData),
		RiskFreeRate:   new(big.Float).Copy(riskFreeRate),
		ConfidenceLevel: new(big.Float).Copy(confidenceLevel),
	}, nil
}

// NewPortfolio creates a new portfolio
func NewPortfolio(id string) (*Portfolio, error) {
	if id == "" {
		return nil, errors.New("portfolio ID cannot be empty")
	}
	
	now := time.Now()
	
	return &Portfolio{
		ID:          id,
		Positions:   make(map[string]*Position),
		RiskMetrics: make(map[MetricType]*RiskMetric),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewPosition creates a new position
func NewPosition(id, assetID string, size, price *big.Float) (*Position, error) {
	if id == "" {
		return nil, errors.New("position ID cannot be empty")
	}
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if size == nil || size.Sign() <= 0 {
		return nil, errors.New("position size must be positive")
	}
	if price == nil || price.Sign() <= 0 {
		return nil, errors.New("position price must be positive")
	}
	
	now := time.Now()
	
	return &Position{
		ID:        id,
		AssetID:   assetID,
		Size:      new(big.Float).Copy(size),
		Price:     new(big.Float).Copy(price),
		Value:     new(big.Float).Mul(size, price),
		Weight:    big.NewFloat(0),
		Returns:   make([]*big.Float, 0),
		Volatility: big.NewFloat(0),
		Beta:      big.NewFloat(0),
		UpdatedAt: now,
	}, nil
}

// AddPosition adds a position to a portfolio
func (p *Portfolio) AddPosition(position *Position) error {
	if position == nil {
		return errors.New("position cannot be nil")
	}
	
	p.Positions[position.ID] = position
	p.UpdatedAt = time.Now()
	
	// Recalculate weights
	p.recalculateWeights()
	
	return nil
}

// RemovePosition removes a position from a portfolio
func (p *Portfolio) RemovePosition(positionID string) error {
	if positionID == "" {
		return errors.New("position ID cannot be empty")
	}
	
	if _, exists := p.Positions[positionID]; !exists {
		return errors.New("position not found")
	}
	
	delete(p.Positions, positionID)
	p.UpdatedAt = time.Now()
	
	// Recalculate weights
	p.recalculateWeights()
	
	return nil
}

// recalculateWeights recalculates position weights based on current values
func (p *Portfolio) recalculateWeights() {
	totalValue := p.GetTotalValue()
	
	if totalValue.Sign() <= 0 {
		return
	}
	
	for _, position := range p.Positions {
		position.Weight = new(big.Float).Quo(position.Value, totalValue)
	}
}

// GetTotalValue calculates the total portfolio value
func (p *Portfolio) GetTotalValue() *big.Float {
	totalValue := big.NewFloat(0)
	
	for _, position := range p.Positions {
		totalValue.Add(totalValue, position.Value)
	}
	
	return totalValue
}

// CalculateVaR calculates the Value at Risk for the portfolio
func (rm *RiskManager) CalculateVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if timeHorizon == nil || timeHorizon.Sign() <= 0 {
		return nil, errors.New("time horizon must be positive")
	}
	
	// Calculate portfolio returns
	portfolioReturns := rm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for VaR calculation")
	}
	
	// Calculate mean and standard deviation
	mean := rm.calculateMean(portfolioReturns)
	stdDev := rm.calculateStandardDeviation(portfolioReturns, mean)
	
	// Calculate VaR using normal distribution assumption
	// VaR = mean - (z-score * stdDev * sqrt(timeHorizon))
	zScore := rm.getZScore(rm.ConfidenceLevel)
	
	timeHorizonFloat, _ := timeHorizon.Float64()
	sqrtTimeHorizon := math.Sqrt(timeHorizonFloat)
	
	zScoreFloat, _ := zScore.Float64()
	stdDevFloat, _ := stdDev.Float64()
	
	varValue := zScoreFloat * stdDevFloat * sqrtTimeHorizon
	varValue = mean - varValue
	
	// Convert back to big.Float
	varResult := big.NewFloat(varValue)
	
	// Store the risk metric
	portfolio.RiskMetrics[VaR] = &RiskMetric{
		Value:      varResult,
		Type:       VaR,
		Timestamp:  time.Now(),
		Confidence: new(big.Float).Copy(rm.ConfidenceLevel),
	}
	
	return varResult, nil
}

// CalculateCVaR calculates the Conditional Value at Risk (Expected Shortfall)
func (rm *RiskManager) CalculateCVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if timeHorizon == nil || timeHorizon.Sign() <= 0 {
		return nil, errors.New("time horizon must be positive")
	}
	
	// Calculate VaR first
	varValue, err := rm.CalculateVaR(portfolio, timeHorizon)
	if err != nil {
		return nil, err
	}
	
	// Calculate CVaR as the expected loss beyond VaR
	portfolioReturns := rm.calculatePortfolioReturns(portfolio)
	varFloat, _ := varValue.Float64()
	
	var sumLosses float64
	var countLosses int
	
	for _, ret := range portfolioReturns {
		retFloat, _ := ret.Float64()
		if retFloat < varFloat {
			sumLosses += retFloat
			countLosses++
		}
	}
	
	var cvarValue float64
	if countLosses > 0 {
		cvarValue = sumLosses / float64(countLosses)
	}
	
	cvarResult := big.NewFloat(cvarValue)
	
	// Store the risk metric
	portfolio.RiskMetrics[CVaR] = &RiskMetric{
		Value:      cvarResult,
		Type:       CVaR,
		Timestamp:  time.Now(),
		Confidence: new(big.Float).Copy(rm.ConfidenceLevel),
	}
	
	return cvarResult, nil
}

// CalculateVolatility calculates the portfolio volatility
func (rm *RiskManager) CalculateVolatility(portfolio *Portfolio) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	
	portfolioReturns := rm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for volatility calculation")
	}
	
	mean := rm.calculateMean(portfolioReturns)
	volatility := rm.calculateStandardDeviation(portfolioReturns, mean)
	
	// Store the risk metric
	portfolio.RiskMetrics[Volatility] = &RiskMetric{
		Value:      volatility,
		Type:       Volatility,
		Timestamp:  time.Now(),
		Confidence: big.NewFloat(0),
	}
	
	return volatility, nil
}

// CalculateSharpeRatio calculates the Sharpe ratio for the portfolio
func (rm *RiskManager) CalculateSharpeRatio(portfolio *Portfolio) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	
	// Calculate excess return (portfolio return - risk-free rate)
	portfolioReturns := rm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for Sharpe ratio calculation")
	}
	
	meanReturn := rm.calculateMean(portfolioReturns)
	meanReturnBig := big.NewFloat(meanReturn)
	excessReturn := new(big.Float).Sub(meanReturnBig, rm.RiskFreeRate)
	
	// Calculate volatility
	volatility, err := rm.CalculateVolatility(portfolio)
	if err != nil {
		return nil, err
	}
	
	// Sharpe ratio = excess return / volatility
	var sharpeRatio *big.Float
	if volatility.Sign() > 0 {
		sharpeRatio = new(big.Float).Quo(excessReturn, volatility)
	} else {
		sharpeRatio = big.NewFloat(0)
	}
	
	// Store the risk metric
	portfolio.RiskMetrics[SharpeRatio] = &RiskMetric{
		Value:      sharpeRatio,
		Type:       SharpeRatio,
		Timestamp:  time.Now(),
		Confidence: big.NewFloat(0),
	}
	
	return sharpeRatio, nil
}

// CalculateMaxDrawdown calculates the maximum drawdown of the portfolio
func (rm *RiskManager) CalculateMaxDrawdown(portfolio *Portfolio) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	
	portfolioValues := rm.calculatePortfolioValues(portfolio)
	if len(portfolioValues) == 0 {
		return nil, errors.New("insufficient data for drawdown calculation")
	}
	
	var maxDrawdown float64
	var peak float64
	
	for i, value := range portfolioValues {
		valueFloat, _ := value.Float64()
		
		if i == 0 || valueFloat > peak {
			peak = valueFloat
		}
		
		drawdown := (peak - valueFloat) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	
	maxDrawdownResult := big.NewFloat(maxDrawdown)
	
	// Store the risk metric
	portfolio.RiskMetrics[MaxDrawdown] = &RiskMetric{
		Value:      maxDrawdownResult,
		Type:       MaxDrawdown,
		Timestamp:  time.Now(),
		Confidence: big.NewFloat(0),
	}
	
	return maxDrawdownResult, nil
}

// CalculateBeta calculates the beta of the portfolio relative to a benchmark
func (rm *RiskManager) CalculateBeta(portfolio *Portfolio, benchmarkReturns []*big.Float) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if len(benchmarkReturns) == 0 {
		return nil, errors.New("benchmark returns cannot be empty")
	}
	
	portfolioReturns := rm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for beta calculation")
	}
	
	// Ensure both arrays have the same length
	minLength := len(portfolioReturns)
	if len(benchmarkReturns) < minLength {
		minLength = len(benchmarkReturns)
	}
	
	// Calculate covariance and variance
	var covariance float64
	var benchmarkVariance float64
	
	portfolioMean := rm.calculateMean(portfolioReturns[:minLength])
	benchmarkMean := rm.calculateMean(benchmarkReturns[:minLength])
	
	for i := 0; i < minLength; i++ {
		portfolioRet, _ := portfolioReturns[i].Float64()
		benchmarkRet, _ := benchmarkReturns[i].Float64()
		
		portfolioDiff := portfolioRet - portfolioMean
		benchmarkDiff := benchmarkRet - benchmarkMean
		
		covariance += portfolioDiff * benchmarkDiff
		benchmarkVariance += benchmarkDiff * benchmarkDiff
	}
	
	covariance /= float64(minLength)
	benchmarkVariance /= float64(minLength)
	
	// Beta = covariance / benchmark variance
	var beta float64
	if benchmarkVariance > 0 {
		beta = covariance / benchmarkVariance
	}
	
	betaResult := big.NewFloat(beta)
	
	// Store the risk metric
	portfolio.RiskMetrics[Beta] = &RiskMetric{
		Value:      betaResult,
		Type:       Beta,
		Timestamp:  time.Now(),
		Confidence: big.NewFloat(0),
	}
	
	return betaResult, nil
}

// StressTest performs stress testing on the portfolio
func (rm *RiskManager) StressTest(portfolio *Portfolio, scenarios []StressScenario) (map[string]*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if len(scenarios) == 0 {
		return nil, errors.New("stress scenarios cannot be empty")
	}
	
	results := make(map[string]*big.Float)
	
	for _, scenario := range scenarios {
		// Apply scenario to portfolio
		scenarioValue := rm.applyStressScenario(portfolio, scenario)
		results[scenario.Name] = scenarioValue
	}
	
	return results, nil
}

// StressScenario represents a stress test scenario
type StressScenario struct {
	Name        string
	AssetShocks map[string]*big.Float // Percentage shocks to apply
	CorrelationShock *big.Float       // Correlation breakdown shock
	VolatilityShock *big.Float        // Volatility increase shock
}

// applyStressScenario applies a stress scenario to the portfolio
func (rm *RiskManager) applyStressScenario(portfolio *Portfolio, scenario StressScenario) *big.Float {
	// Calculate base portfolio value
	baseValue := portfolio.GetTotalValue()
	
	// Apply asset-specific shocks
	var stressedValue *big.Float
	for assetID, shock := range scenario.AssetShocks {
		for _, position := range portfolio.Positions {
			if position.AssetID == assetID {
				shockMultiplier := new(big.Float).Add(big.NewFloat(1), shock)
				positionValue := new(big.Float).Mul(position.Value, shockMultiplier)
				
				if stressedValue == nil {
					stressedValue = positionValue
				} else {
					stressedValue.Add(stressedValue, positionValue)
				}
			}
		}
	}
	
	// If no specific shocks applied, use base value
	if stressedValue == nil {
		stressedValue = new(big.Float).Copy(baseValue)
	}
	
	// Calculate percentage change
	change := new(big.Float).Sub(stressedValue, baseValue)
	percentageChange := new(big.Float).Quo(change, baseValue)
	
	return percentageChange
}

// Helper functions for calculations

// calculatePortfolioReturns calculates the portfolio returns over time
func (rm *RiskManager) calculatePortfolioReturns(portfolio *Portfolio) []*big.Float {
	// This is a simplified implementation
	// In practice, you'd use actual historical portfolio values
	portfolioValues := rm.calculatePortfolioValues(portfolio)
	
	if len(portfolioValues) < 2 {
		return []*big.Float{}
	}
	
	returns := make([]*big.Float, len(portfolioValues)-1)
	
	for i := 1; i < len(portfolioValues); i++ {
		currentValue := portfolioValues[i]
		previousValue := portfolioValues[i-1]
		
		returnValue := new(big.Float).Sub(currentValue, previousValue)
		returnValue.Quo(returnValue, previousValue)
		
		returns[i-1] = returnValue
	}
	
	return returns
}

// calculatePortfolioValues calculates the portfolio values over time
func (rm *RiskManager) calculatePortfolioValues(portfolio *Portfolio) []*big.Float {
	// This is a simplified implementation
	// In practice, you'd use actual historical portfolio values
	values := make([]*big.Float, 0)
	
	// Simulate some historical values
	baseValue := portfolio.GetTotalValue()
	if baseValue.Sign() <= 0 {
		return values
	}
	
	// Generate some sample values (in practice, use real data)
	for i := 0; i < 30; i++ { // 30 days of data
		// Add some random variation
		variation := big.NewFloat(0.95 + 0.1*float64(i%10)/10.0)
		value := new(big.Float).Mul(baseValue, variation)
		values = append(values, value)
	}
	
	return values
}

// calculateMean calculates the mean of a slice of big.Float values
func (rm *RiskManager) calculateMean(values []*big.Float) float64 {
	if len(values) == 0 {
		return 0
	}
	
	var sum float64
	for _, value := range values {
		val, _ := value.Float64()
		sum += val
	}
	
	return sum / float64(len(values))
}

// calculateStandardDeviation calculates the standard deviation
func (rm *RiskManager) calculateStandardDeviation(values []*big.Float, mean float64) *big.Float {
	if len(values) == 0 {
		return big.NewFloat(0)
	}
	
	var sumSquaredDiff float64
	for _, value := range values {
		val, _ := value.Float64()
		diff := val - mean
		sumSquaredDiff += diff * diff
	}
	
	variance := sumSquaredDiff / float64(len(values))
	stdDev := math.Sqrt(variance)
	
	return big.NewFloat(stdDev)
}

// getZScore returns the z-score for a given confidence level
func (rm *RiskManager) getZScore(confidenceLevel *big.Float) *big.Float {
	// Simplified z-score calculation
	// In practice, use proper statistical tables or libraries
	confidence, _ := confidenceLevel.Float64()
	
	var zScore float64
	switch {
	case confidence >= 0.99:
		zScore = 2.326
	case confidence >= 0.95:
		zScore = 1.645
	case confidence >= 0.90:
		zScore = 1.282
	default:
		zScore = 1.0
	}
	
	return big.NewFloat(zScore)
}

// UpdatePosition updates a position's price and recalculates value
func (p *Position) UpdatePosition(newPrice *big.Float) error {
	if newPrice == nil || newPrice.Sign() < 0 {
		return errors.New("new price must be non-negative")
	}
	
	p.Price = new(big.Float).Copy(newPrice)
	p.Value = new(big.Float).Mul(p.Size, p.Price)
	p.UpdatedAt = time.Now()
	
	return nil
}

// AddReturn adds a return value to the position's return history
func (p *Position) AddReturn(returnValue *big.Float) {
	if returnValue == nil {
		return
	}
	
	p.Returns = append(p.Returns, new(big.Float).Copy(returnValue))
	
	// Keep only the last 100 returns to avoid memory issues
	if len(p.Returns) > 100 {
		p.Returns = p.Returns[len(p.Returns)-100:]
	}
}
