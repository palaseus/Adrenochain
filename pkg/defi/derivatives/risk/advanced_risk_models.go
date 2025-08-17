package risk

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

// VaRMethodology represents the methodology used for VaR calculation
type VaRMethodology int

const (
	ParametricVaR VaRMethodology = iota
	HistoricalVaR
	MonteCarloVaR
	FilteredHistoricalVaR
)

// StressTestScenario represents a stress test scenario
type StressTestScenario struct {
	ID          string
	Name        string
	Description string
	Parameters  map[string]*big.Float
	Severity    StressSeverity
	CreatedAt   time.Time
}

// StressSeverity represents the severity of a stress test
type StressSeverity int

const (
	MildStress StressSeverity = iota
	ModerateStress
	SevereStress
	ExtremeStress
)

// ScenarioAnalysis represents a scenario analysis result
type ScenarioAnalysis struct {
	ID          string
	Scenario    *StressTestScenario
	Portfolio   *Portfolio
	Results     map[MetricType]*big.Float
	CreatedAt   time.Time
}

// MonteCarloSimulation represents a Monte Carlo simulation configuration
type MonteCarloSimulation struct {
	ID              string
	NumSimulations  int
	TimeHorizon     *big.Float
	ConfidenceLevel *big.Float
	Seed            int64
	CreatedAt       time.Time
}

// MonteCarloResult represents the result of a Monte Carlo simulation
type MonteCarloResult struct {
	ID           string
	Simulation   *MonteCarloSimulation
	Portfolio    *Portfolio
	VaR          *big.Float
	CVaR         *big.Float
	Percentiles  map[int]*big.Float
	Simulations  []*big.Float
	CreatedAt    time.Time
}

// AdvancedRiskManager extends the basic risk manager with advanced models
type AdvancedRiskManager struct {
	*RiskManager
	scenarios        map[string]*StressTestScenario
	simulations      map[string]*MonteCarloSimulation
	historicalData   map[string][]*big.Float
	correlationMatrix map[string]map[string]*big.Float
}

// NewAdvancedRiskManager creates a new advanced risk manager
func NewAdvancedRiskManager(riskFreeRate, confidenceLevel *big.Float) (*AdvancedRiskManager, error) {
	baseManager, err := NewRiskManager(riskFreeRate, confidenceLevel)
	if err != nil {
		return nil, err
	}

	return &AdvancedRiskManager{
		RiskManager:      baseManager,
		scenarios:        make(map[string]*StressTestScenario),
		simulations:      make(map[string]*MonteCarloSimulation),
		historicalData:   make(map[string][]*big.Float),
		correlationMatrix: make(map[string]map[string]*big.Float),
	}, nil
}

// NewStressTestScenario creates a new stress test scenario
func NewStressTestScenario(id, name, description string, parameters map[string]*big.Float, severity StressSeverity) (*StressTestScenario, error) {
	if id == "" {
		return nil, errors.New("scenario ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("scenario name cannot be empty")
	}
	if parameters == nil {
		return nil, errors.New("parameters cannot be nil")
	}

	return &StressTestScenario{
		ID:          id,
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Severity:    severity,
		CreatedAt:   time.Now(),
	}, nil
}

// NewMonteCarloSimulation creates a new Monte Carlo simulation configuration
func NewMonteCarloSimulation(id string, numSimulations int, timeHorizon, confidenceLevel *big.Float) (*MonteCarloSimulation, error) {
	if id == "" {
		return nil, errors.New("simulation ID cannot be empty")
	}
	if numSimulations <= 0 {
		return nil, errors.New("number of simulations must be positive")
	}
	if timeHorizon == nil || timeHorizon.Sign() <= 0 {
		return nil, errors.New("time horizon must be positive")
	}
	if confidenceLevel == nil || confidenceLevel.Cmp(big.NewFloat(0)) <= 0 || confidenceLevel.Cmp(big.NewFloat(1)) >= 0 {
		return nil, errors.New("confidence level must be between 0 and 1")
	}

	return &MonteCarloSimulation{
		ID:              id,
		NumSimulations:  numSimulations,
		TimeHorizon:     new(big.Float).Copy(timeHorizon),
		ConfidenceLevel: new(big.Float).Copy(confidenceLevel),
		Seed:            time.Now().UnixNano(),
		CreatedAt:       time.Now(),
	}, nil
}

// CalculateVaRAdvanced calculates VaR using multiple methodologies
func (arm *AdvancedRiskManager) CalculateVaRAdvanced(
	portfolio *Portfolio,
	timeHorizon *big.Float,
	methodology VaRMethodology,
) (*big.Float, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if timeHorizon == nil || timeHorizon.Sign() <= 0 {
		return nil, errors.New("time horizon must be positive")
	}

	switch methodology {
	case ParametricVaR:
		return arm.calculateParametricVaR(portfolio, timeHorizon)
	case HistoricalVaR:
		return arm.calculateHistoricalVaR(portfolio, timeHorizon)
	case MonteCarloVaR:
		return arm.calculateMonteCarloVaR(portfolio, timeHorizon)
	case FilteredHistoricalVaR:
		return arm.calculateFilteredHistoricalVaR(portfolio, timeHorizon)
	default:
		return nil, errors.New("unsupported VaR methodology")
	}
}

// calculateParametricVaR calculates VaR using parametric (normal distribution) approach
func (arm *AdvancedRiskManager) calculateParametricVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	// Use the base VaR calculation which is parametric
	return arm.CalculateVaR(portfolio, timeHorizon)
}

// calculateHistoricalVaR calculates VaR using historical simulation
func (arm *AdvancedRiskManager) calculateHistoricalVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	portfolioReturns := arm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for historical VaR calculation")
	}

	// Sort returns in ascending order
	sortedReturns := make([]float64, len(portfolioReturns))
	for i, ret := range portfolioReturns {
		sortedReturns[i], _ = ret.Float64()
	}
	sort.Float64s(sortedReturns)

	// Calculate the percentile based on confidence level
	confidenceFloat, _ := arm.ConfidenceLevel.Float64()
	percentile := (1 - confidenceFloat) * float64(len(sortedReturns))
	index := int(math.Floor(percentile))

	if index < 0 {
		index = 0
	} else if index >= len(sortedReturns) {
		index = len(sortedReturns) - 1
	}

	// Apply time horizon scaling
	timeHorizonFloat, _ := timeHorizon.Float64()
	scaledVaR := sortedReturns[index] * math.Sqrt(timeHorizonFloat)

	// Convert to positive value (VaR is typically reported as positive)
	if scaledVaR < 0 {
		scaledVaR = -scaledVaR
	}

	varResult := big.NewFloat(scaledVaR)

	// Store the risk metric
	portfolio.RiskMetrics[VaR] = &RiskMetric{
		Value:      varResult,
		Type:       VaR,
		Timestamp:  time.Now(),
		Confidence: new(big.Float).Copy(arm.ConfidenceLevel),
	}

	return varResult, nil
}

// calculateMonteCarloVaR calculates VaR using Monte Carlo simulation
func (arm *AdvancedRiskManager) calculateMonteCarloVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	// Create a default Monte Carlo simulation
	simulation, err := NewMonteCarloSimulation(
		"default_mc",
		10000, // 10,000 simulations
		timeHorizon,
		arm.ConfidenceLevel,
	)
	if err != nil {
		return nil, err
	}

	// Run the simulation
	result, err := arm.RunMonteCarloSimulation(portfolio, simulation)
	if err != nil {
		return nil, err
	}

	// Store the risk metric
	portfolio.RiskMetrics[VaR] = &RiskMetric{
		Value:      result.VaR,
		Type:       VaR,
		Timestamp:  time.Now(),
		Confidence: new(big.Float).Copy(arm.ConfidenceLevel),
	}

	return result.VaR, nil
}

// calculateFilteredHistoricalVaR calculates VaR using filtered historical simulation
func (arm *AdvancedRiskManager) calculateFilteredHistoricalVaR(portfolio *Portfolio, timeHorizon *big.Float) (*big.Float, error) {
	// This is a simplified version of filtered historical VaR
	// In practice, this would use GARCH or similar models to filter volatility clustering
	
	// For now, use historical VaR as a base
	baseVaR, err := arm.calculateHistoricalVaR(portfolio, timeHorizon)
	if err != nil {
		return nil, err
	}

	// Apply a simple volatility adjustment (this is a placeholder)
	// In practice, this would use sophisticated volatility modeling
	volatilityAdjustment := big.NewFloat(1.1) // 10% increase for volatility clustering
	adjustedVaR := new(big.Float).Mul(baseVaR, volatilityAdjustment)

	// Store the risk metric
	portfolio.RiskMetrics[VaR] = &RiskMetric{
		Value:      adjustedVaR,
		Type:       VaR,
		Timestamp:  time.Now(),
		Confidence: new(big.Float).Copy(arm.ConfidenceLevel),
	}

	return adjustedVaR, nil
}

// RunMonteCarloSimulation runs a Monte Carlo simulation for risk assessment
func (arm *AdvancedRiskManager) RunMonteCarloSimulation(
	portfolio *Portfolio,
	simulation *MonteCarloSimulation,
) (*MonteCarloResult, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if simulation == nil {
		return nil, errors.New("simulation cannot be nil")
	}

	// Set random seed
	rand.Seed(simulation.Seed)

	// Get portfolio returns for parameter estimation
	portfolioReturns := arm.calculatePortfolioReturns(portfolio)
	if len(portfolioReturns) == 0 {
		return nil, errors.New("insufficient data for Monte Carlo simulation")
	}

	// Calculate parameters from historical data
	mean := arm.calculateMean(portfolioReturns)
	stdDev := arm.calculateStandardDeviation(portfolioReturns, mean)
	timeHorizonFloat, _ := simulation.TimeHorizon.Float64()
	meanFloat := mean
	stdDevFloat, _ := stdDev.Float64()

	// Generate simulations
	simulations := make([]*big.Float, simulation.NumSimulations)
	for i := 0; i < simulation.NumSimulations; i++ {
		// Generate random return using normal distribution
		randomReturn := rand.NormFloat64()*stdDevFloat + meanFloat
		
		// Scale by time horizon
		scaledReturn := randomReturn * math.Sqrt(timeHorizonFloat)
		
		simulations[i] = big.NewFloat(scaledReturn)
	}

	// Sort simulations for percentile calculations
	sortedSimulations := make([]float64, len(simulations))
	for i, sim := range simulations {
		sortedSimulations[i], _ = sim.Float64()
	}
	sort.Float64s(sortedSimulations)

	// Calculate VaR and CVaR
	confidenceFloat, _ := simulation.ConfidenceLevel.Float64()
	percentile := (1 - confidenceFloat) * float64(len(sortedSimulations))
	index := int(math.Floor(percentile))

	if index < 0 {
		index = 0
	} else if index >= len(sortedSimulations) {
		index = len(sortedSimulations) - 1
	}

	varValue := big.NewFloat(sortedSimulations[index])
	if sortedSimulations[index] < 0 {
		varValue = big.NewFloat(-sortedSimulations[index])
	}

	// Calculate CVaR (Expected Shortfall)
	var sumLosses float64
	var countLosses int
	varFloat, _ := varValue.Float64()
	
	for _, sim := range sortedSimulations {
		if sim < varFloat {
			sumLosses += sim
			countLosses++
		}
	}

	var cvarValue float64
	if countLosses > 0 {
		cvarValue = sumLosses / float64(countLosses)
	}
	if cvarValue < 0 {
		cvarValue = -cvarValue
	}

	cvarValueBig := big.NewFloat(cvarValue)

	// Calculate key percentiles
	percentiles := make(map[int]*big.Float)
	keyPercentiles := []int{1, 5, 10, 25, 50, 75, 90, 95, 99}
	
	for _, p := range keyPercentiles {
		index := int(float64(len(sortedSimulations)) * float64(p) / 100.0)
		if index >= 0 && index < len(sortedSimulations) {
			percentiles[p] = big.NewFloat(sortedSimulations[index])
		}
	}

	// Convert simulations back to big.Float
	simulationsBig := make([]*big.Float, len(simulations))
	for i, sim := range simulations {
		simFloat, _ := sim.Float64()
		simulationsBig[i] = big.NewFloat(simFloat)
	}

	result := &MonteCarloResult{
		ID:          generateMonteCarloResultID(portfolio.ID, simulation.ID),
		Simulation:  simulation,
		Portfolio:   portfolio,
		VaR:         varValue,
		CVaR:        cvarValueBig,
		Percentiles: percentiles,
		Simulations: simulationsBig,
		CreatedAt:   time.Now(),
	}

	return result, nil
}

// RunStressTest runs a stress test on a portfolio
func (arm *AdvancedRiskManager) RunStressTest(
	portfolio *Portfolio,
	scenario *StressTestScenario,
) (*ScenarioAnalysis, error) {
	if portfolio == nil {
		return nil, errors.New("portfolio cannot be nil")
	}
	if scenario == nil {
		return nil, errors.New("scenario cannot be nil")
	}

	// Create a copy of the portfolio for stress testing
	stressedPortfolio := arm.createStressedPortfolio(portfolio, scenario)

	// Calculate risk metrics under stress
	results := make(map[MetricType]*big.Float)

	// Calculate VaR under stress
	timeHorizon := big.NewFloat(1.0) // 1 day horizon for stress testing
	stressVaR, err := arm.CalculateVaR(stressedPortfolio, timeHorizon)
	if err == nil {
		results[VaR] = stressVaR
	}

	// Calculate CVaR under stress
	stressCVaR, err := arm.CalculateCVaR(stressedPortfolio, timeHorizon)
	if err == nil {
		results[CVaR] = stressCVaR
	}

	// Calculate volatility under stress
	stressVolatility, err := arm.CalculateVolatility(stressedPortfolio)
	if err == nil {
		results[Volatility] = stressVolatility
	}

	// Calculate other metrics as needed
	stressSharpe, err := arm.CalculateSharpeRatio(stressedPortfolio)
	if err == nil {
		results[SharpeRatio] = stressSharpe
	}

	analysis := &ScenarioAnalysis{
		ID:        generateScenarioAnalysisID(portfolio.ID, scenario.ID),
		Scenario:  scenario,
		Portfolio: portfolio,
		Results:   results,
		CreatedAt: time.Now(),
	}

	return analysis, nil
}

// createStressedPortfolio creates a portfolio with stressed parameters
func (arm *AdvancedRiskManager) createStressedPortfolio(
	portfolio *Portfolio,
	scenario *StressTestScenario,
) *Portfolio {
	// Create a copy of the portfolio
	stressedPortfolio := &Portfolio{
		ID:          portfolio.ID + "_stressed",
		Positions:   make(map[string]*Position),
		RiskMetrics: make(map[MetricType]*RiskMetric),
		CreatedAt:   portfolio.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	// Apply stress parameters to positions
	for id, position := range portfolio.Positions {
		stressedPosition := &Position{
			ID:        position.ID,
			AssetID:   position.AssetID,
			Size:      new(big.Float).Copy(position.Size),
			Price:     new(big.Float).Copy(position.Price),
			Value:     new(big.Float).Copy(position.Value),
			Weight:    new(big.Float).Copy(position.Weight),
			Returns:   make([]*big.Float, len(position.Returns)),
			Volatility: new(big.Float).Copy(position.Volatility),
			Beta:      new(big.Float).Copy(position.Beta),
			UpdatedAt: time.Now(),
		}

		// Copy returns
		for i, ret := range position.Returns {
			stressedPosition.Returns[i] = new(big.Float).Copy(ret)
		}

		// Apply stress parameters if they exist
		if stressFactor, exists := scenario.Parameters["price_shock_"+position.AssetID]; exists {
			// Apply price shock
			stressedPosition.Price = new(big.Float).Mul(position.Price, stressFactor)
			stressedPosition.Value = new(big.Float).Mul(stressedPosition.Size, stressedPosition.Price)
		}

		if volatilityShock, exists := scenario.Parameters["volatility_shock_"+position.AssetID]; exists {
			// Apply volatility shock
			stressedPosition.Volatility = new(big.Float).Mul(position.Volatility, volatilityShock)
		}

		stressedPortfolio.Positions[id] = stressedPosition
	}

	// Recalculate weights
	stressedPortfolio.recalculateWeights()

	return stressedPortfolio
}

// AddStressTestScenario adds a stress test scenario
func (arm *AdvancedRiskManager) AddStressTestScenario(scenario *StressTestScenario) error {
	if scenario == nil {
		return errors.New("scenario cannot be nil")
	}

	arm.scenarios[scenario.ID] = scenario
	return nil
}

// GetStressTestScenario retrieves a stress test scenario
func (arm *AdvancedRiskManager) GetStressTestScenario(id string) (*StressTestScenario, error) {
	if id == "" {
		return nil, errors.New("scenario ID cannot be empty")
	}

	scenario, exists := arm.scenarios[id]
	if !exists {
		return nil, errors.New("scenario not found")
	}

	return scenario, nil
}

// AddMonteCarloSimulation adds a Monte Carlo simulation configuration
func (arm *AdvancedRiskManager) AddMonteCarloSimulation(sim *MonteCarloSimulation) error {
	if sim == nil {
		return errors.New("simulation cannot be nil")
	}

	arm.simulations[sim.ID] = sim
	return nil
}

// GetMonteCarloSimulation retrieves a Monte Carlo simulation configuration
func (arm *AdvancedRiskManager) GetMonteCarloSimulation(id string) (*MonteCarloSimulation, error) {
	if id == "" {
		return nil, errors.New("simulation ID cannot be empty")
	}

	sim, exists := arm.simulations[id]
	if !exists {
		return nil, errors.New("simulation not found")
	}

	return sim, nil
}

// generateMonteCarloResultID generates a unique ID for Monte Carlo results
func generateMonteCarloResultID(portfolioID, simulationID string) string {
	return portfolioID + "_" + simulationID + "_" + time.Now().Format("20060102150405")
}

// generateScenarioAnalysisID generates a unique ID for scenario analysis
func generateScenarioAnalysisID(portfolioID, scenarioID string) string {
	return portfolioID + "_" + scenarioID + "_" + time.Now().Format("20060102150405")
}
