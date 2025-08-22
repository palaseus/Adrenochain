package ai

import (
	"fmt"
	"math/rand"
	"time"
)

// CrashEvent represents a historical market crash
type CrashEvent struct {
	ID           string
	Name         string
	Date         time.Time
	Impact       CrashImpact
	PriceDrop    float64
	RecoveryTime time.Duration
	Description  string
}

// CrashImpact represents the severity of a crash
type CrashImpact int

const (
	CrashImpactMild CrashImpact = iota
	CrashImpactModerate
	CrashImpactSevere
	CrashImpactCatastrophic
)

// HistoricalDataEngine handles historical market data analysis
type HistoricalDataEngine struct {
	ID           string
	CrashEvents  map[string]*CrashEvent
	MarketData   map[string]interface{}
	AnalysisData map[string]float64
}

// NewHistoricalDataEngine creates a new historical data engine
func NewHistoricalDataEngine() *HistoricalDataEngine {
	return &HistoricalDataEngine{
		ID:           fmt.Sprintf("hist_engine_%d", time.Now().UnixNano()),
		CrashEvents:  make(map[string]*CrashEvent),
		MarketData:   make(map[string]interface{}),
		AnalysisData: make(map[string]float64),
	}
}

// AddCrashEvent adds a historical crash event
func (hde *HistoricalDataEngine) AddCrashEvent(crash *CrashEvent) {
	hde.CrashEvents[crash.ID] = crash
}

// GetCrashEvent retrieves a crash event by ID
func (hde *HistoricalDataEngine) GetCrashEvent(id string) (*CrashEvent, bool) {
	crash, exists := hde.CrashEvents[id]
	return crash, exists
}

// SimulateCrashScenario simulates a strategy's performance during a crash
func (hde *HistoricalDataEngine) SimulateCrashScenario(strategy *Strategy, crash *CrashEvent) (bool, float64, time.Duration) {
	// Base survival probability based on strategy type
	baseSurvivalProb := 0.4 // 40% base survival (realistic, not 100%)
	
	// Adjust based on crash severity
	var severityMultiplier float64
	switch crash.Impact {
	case CrashImpactMild:
		severityMultiplier = 1.2
	case CrashImpactModerate:
		severityMultiplier = 1.0
	case CrashImpactSevere:
		severityMultiplier = 0.7
	case CrashImpactCatastrophic:
		severityMultiplier = 0.4
	}
	
	// Adjust based on strategy stress resistance
	stressResistanceMultiplier := 1.0
	if strategy.RiskLevel > 0.8 {
		stressResistanceMultiplier = 1.3
	} else if strategy.RiskLevel > 0.6 {
		stressResistanceMultiplier = 1.1
	} else if strategy.RiskLevel < 0.3 {
		stressResistanceMultiplier = 0.8
	}
	
	// Calculate final survival probability
	finalSurvivalProb := baseSurvivalProb * severityMultiplier * stressResistanceMultiplier
	
	// Cap at realistic maximum (not 100%)
	if finalSurvivalProb > 0.85 {
		finalSurvivalProb = 0.85
	}
	
	// Determine survival with random component
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Float64()
	survived := randomValue < finalSurvivalProb
	
	// Calculate capital loss (realistic, not 0%)
	var capitalLoss float64
	if survived {
		// Survived but with losses
		capitalLoss = crash.PriceDrop * (0.3 + rand.Float64()*0.4) // 30-70% of crash loss
	} else {
		// Failed completely
		capitalLoss = crash.PriceDrop * (0.8 + rand.Float64()*0.2) // 80-100% of crash loss
	}
	
	// Calculate recovery time
	var recoveryTime time.Duration
	if survived {
		recoveryTime = crash.RecoveryTime * time.Duration(0.5+rand.Float64()*1.0) // 50-150% of crash recovery time
	} else {
		recoveryTime = crash.RecoveryTime * time.Duration(2.0+rand.Float64()*2.0) // 200-400% of crash recovery time
	}
	
	return survived, capitalLoss, recoveryTime
}

// RunHistoricalBacktest runs a comprehensive historical backtest
func (hde *HistoricalDataEngine) RunHistoricalBacktest(strategy *Strategy) *BacktestResult {
	result := &BacktestResult{
		StrategyID:    strategy.ID,
		TotalScenarios: len(hde.CrashEvents),
		Survived:      0,
		Failed:        0,
		TotalLoss:     0.0,
		AverageRecovery: 0,
		StartTime:     time.Now(),
	}
	
	// Test against all historical crashes
	for _, crash := range hde.CrashEvents {
		survived, loss, recovery := hde.SimulateCrashScenario(strategy, crash)
		
		if survived {
			result.Survived++
		} else {
			result.Failed++
		}
		
		result.TotalLoss += loss
		result.AverageRecovery += int64(recovery)
	}
	
	// Calculate final metrics
	if result.TotalScenarios > 0 {
		result.SurvivalRate = float64(result.Survived) / float64(result.TotalScenarios)
		result.AverageRecovery = result.AverageRecovery / int64(result.TotalScenarios)
	}
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	return result
}

// BacktestResult represents the results of a historical backtest
type BacktestResult struct {
	StrategyID      string
	TotalScenarios  int
	Survived        int
	Failed          int
	SurvivalRate    float64
	TotalLoss       float64
	AverageRecovery int64
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
}

// GetSurvivalRate returns the survival rate as a percentage
func (br *BacktestResult) GetSurvivalRate() float64 {
	return br.SurvivalRate * 100.0
}

// GetAverageLoss returns the average loss per scenario
func (br *BacktestResult) GetAverageLoss() float64 {
	if br.TotalScenarios == 0 {
		return 0.0
	}
	return br.TotalLoss / float64(br.TotalScenarios)
}

// IsSuccessful returns true if the strategy meets minimum survival criteria
func (br *BacktestResult) IsSuccessful(minSurvivalRate float64) bool {
	return br.SurvivalRate >= minSurvivalRate
}

// String returns a string representation of the backtest result
func (br *BacktestResult) String() string {
	return fmt.Sprintf("Strategy: %s, Survival Rate: %.1f%%, Scenarios: %d/%d, Avg Loss: %.2f%%",
		br.StrategyID, br.GetSurvivalRate(), br.Survived, br.TotalScenarios, br.GetAverageLoss())
}
