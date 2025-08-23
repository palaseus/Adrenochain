package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MarketRegime represents different market conditions
type MarketRegime int

const (
	MarketRegimeBull MarketRegime = iota
	MarketRegimeBear
	MarketRegimeSideways
	MarketRegimeVolatile
	MarketRegimeCrisis
)

// StressLevel represents the stress level of a strategy
type StressLevel int

const (
	StressLevelLow StressLevel = iota
	StressLevelMedium
	StressLevelHigh
	StressLevelExtreme
)

// StressResistantStrategy represents a strategy that can handle market stress
type StressResistantStrategy struct {
	ID                    string
	BaseStrategy          *Strategy
	RiskManager           *AdaptiveRiskManager
	StressDetector        *StressDetector
	EmergencyProtocol     *EmergencyProtocolManager
	RegimeAdaptation      *RegimeAdaptationEngine
	VolatilityAnalyzer    *VolatilityAnalyzer
	LiquidityManager      *LiquidityManager
	BlackSwanProtection   *BlackSwanProtectionSystem
	mu                    sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc
}

// AdaptiveRiskManager manages risk dynamically
type AdaptiveRiskManager struct {
	ID              string
	MaxRisk         float64
	CurrentRisk     float64
	RiskAdjustments map[string]float64
}

// StressDetector detects market stress conditions
type StressDetector struct {
	ID              string
	StressThreshold float64
	CurrentStress   float64
	StressHistory   []float64
}

// EmergencyProtocolManager handles emergency protocols
type EmergencyProtocolManager struct {
	ID              string
	Protocols       map[string]string
	IsActive        bool
	ActivationTime  time.Time
}

// RegimeAdaptationEngine adapts to market regime changes
type RegimeAdaptationEngine struct {
	ID              string
	CurrentRegime   MarketRegime
	RegimeHistory   []MarketRegime
	AdaptationRules map[MarketRegime]string
}

// VolatilityAnalyzer analyzes market volatility
type VolatilityAnalyzer struct {
	ID              string
	CurrentVolatility float64
	VolatilityHistory []float64
	Threshold        float64
}

// LiquidityManager manages liquidity during stress
type LiquidityManager struct {
	ID              string
	CurrentLiquidity float64
	MinLiquidity    float64
	LiquidityBuffer float64
}

// BlackSwanProtectionSystem provides protection against black swan events
type BlackSwanProtectionSystem struct {
	ID              string
	IsActive        bool
	ProtectionLevel float64
	LastActivation  time.Time
}

// NewStressResistantStrategy creates a new stress-resistant strategy
func NewStressResistantStrategy(baseStrategy *Strategy) *StressResistantStrategy {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StressResistantStrategy{
		ID:                    fmt.Sprintf("stress_strategy_%d", time.Now().UnixNano()),
		BaseStrategy:          baseStrategy,
		RiskManager:           &AdaptiveRiskManager{ID: "risk_mgr", MaxRisk: 0.3, CurrentRisk: 0.1},
		StressDetector:        &StressDetector{ID: "stress_detector", StressThreshold: 0.7},
		EmergencyProtocol:     &EmergencyProtocolManager{ID: "emergency", Protocols: make(map[string]string)},
		RegimeAdaptation:      &RegimeAdaptationEngine{ID: "regime_adapt", CurrentRegime: MarketRegimeSideways, AdaptationRules: make(map[MarketRegime]string)},
		VolatilityAnalyzer:    &VolatilityAnalyzer{ID: "vol_analyzer", Threshold: 0.5},
		LiquidityManager:      &LiquidityManager{ID: "liquidity_mgr", MinLiquidity: 0.2, LiquidityBuffer: 0.1},
		BlackSwanProtection:   &BlackSwanProtectionSystem{ID: "black_swan_prot", ProtectionLevel: 0.8},
		ctx:                   ctx,
		cancel:                cancel,
	}
}

// HandleMarketStress handles market stress conditions
func (srs *StressResistantStrategy) HandleMarketStress(stressLevel float64, regime MarketRegime) error {
	srs.mu.Lock()
	defer srs.mu.Unlock()
	
	// Update stress detection
	srs.StressDetector.CurrentStress = stressLevel
	srs.StressDetector.StressHistory = append(srs.StressDetector.StressHistory, stressLevel)
	
	// Update regime detection
	srs.RegimeAdaptation.CurrentRegime = regime
	srs.RegimeAdaptation.RegimeHistory = append(srs.RegimeAdaptation.RegimeHistory, regime)
	
	// Check if emergency protocols should be activated
	if stressLevel > srs.StressDetector.StressThreshold {
		srs.activateEmergencyProtocols()
	}
	
	// Adapt to new regime
	srs.adaptToRegime(regime)
	
	// Adjust risk management
	srs.adjustRiskManagement(stressLevel)
	
	// Update volatility analysis
	srs.updateVolatilityAnalysis(stressLevel)
	
	// Manage liquidity
	srs.manageLiquidity(stressLevel)
	
	// Check for black swan conditions
	if stressLevel > 0.9 {
		srs.activateBlackSwanProtection()
	}
	
	return nil
}

// activateEmergencyProtocols activates emergency protocols
func (srs *StressResistantStrategy) activateEmergencyProtocols() {
	srs.EmergencyProtocol.IsActive = true
	srs.EmergencyProtocol.ActivationTime = time.Now()
	
	// Add emergency protocols
	srs.EmergencyProtocol.Protocols["stop_loss"] = "activated"
	srs.EmergencyProtocol.Protocols["position_reduction"] = "activated"
	srs.EmergencyProtocol.Protocols["hedging"] = "activated"
}

// adaptToRegime adapts the strategy to the current market regime
func (srs *StressResistantStrategy) adaptToRegime(regime MarketRegime) {
	switch regime {
	case MarketRegimeBull:
		srs.RegimeAdaptation.AdaptationRules[regime] = "aggressive_growth"
	case MarketRegimeBear:
		srs.RegimeAdaptation.AdaptationRules[regime] = "defensive_conservation"
	case MarketRegimeSideways:
		srs.RegimeAdaptation.AdaptationRules[regime] = "range_trading"
	case MarketRegimeVolatile:
		srs.RegimeAdaptation.AdaptationRules[regime] = "volatility_harvesting"
	case MarketRegimeCrisis:
		srs.RegimeAdaptation.AdaptationRules[regime] = "crisis_management"
	}
}

// adjustRiskManagement adjusts risk management based on stress
func (srs *StressResistantStrategy) adjustRiskManagement(stressLevel float64) {
	// Reduce risk as stress increases
	if stressLevel > 0.7 {
		srs.RiskManager.CurrentRisk = srs.RiskManager.MaxRisk * 0.5
	} else if stressLevel > 0.5 {
		srs.RiskManager.CurrentRisk = srs.RiskManager.MaxRisk * 0.7
	} else {
		srs.RiskManager.CurrentRisk = srs.RiskManager.MaxRisk
	}
}

// updateVolatilityAnalysis updates volatility analysis
func (srs *StressResistantStrategy) updateVolatilityAnalysis(stressLevel float64) {
	srs.VolatilityAnalyzer.CurrentVolatility = stressLevel
	srs.VolatilityAnalyzer.VolatilityHistory = append(srs.VolatilityAnalyzer.VolatilityHistory, stressLevel)
	
	// Keep only last 100 volatility readings
	if len(srs.VolatilityAnalyzer.VolatilityHistory) > 100 {
		srs.VolatilityAnalyzer.VolatilityHistory = srs.VolatilityAnalyzer.VolatilityHistory[1:]
	}
}

// manageLiquidity manages liquidity during stress
func (srs *StressResistantStrategy) manageLiquidity(stressLevel float64) {
	// Increase liquidity buffer during high stress
	if stressLevel > 0.8 {
		srs.LiquidityManager.LiquidityBuffer = 0.3
	} else if stressLevel > 0.6 {
		srs.LiquidityManager.LiquidityBuffer = 0.2
	} else {
		srs.LiquidityManager.LiquidityBuffer = 0.1
	}
}

// activateBlackSwanProtection activates black swan protection
func (srs *StressResistantStrategy) activateBlackSwanProtection() {
	srs.BlackSwanProtection.IsActive = true
	srs.BlackSwanProtection.LastActivation = time.Now()
}

// GetStressLevel returns the current stress level
func (srs *StressResistantStrategy) GetStressLevel() StressLevel {
	srs.mu.RLock()
	defer srs.mu.RUnlock()
	
	stress := srs.StressDetector.CurrentStress
	
	if stress < 0.3 {
		return StressLevelLow
	} else if stress < 0.6 {
		return StressLevelMedium
	} else if stress < 0.8 {
		return StressLevelHigh
	} else {
		return StressLevelExtreme
	}
}

// GetCurrentRegime returns the current market regime
func (srs *StressResistantStrategy) GetCurrentRegime() MarketRegime {
	srs.mu.RLock()
	defer srs.mu.RUnlock()
	
	return srs.RegimeAdaptation.CurrentRegime
}

// IsEmergencyActive returns true if emergency protocols are active
func (srs *StressResistantStrategy) IsEmergencyActive() bool {
	srs.mu.RLock()
	defer srs.mu.RUnlock()
	
	return srs.EmergencyProtocol.IsActive
}

// GetRiskLevel returns the current risk level
func (srs *StressResistantStrategy) GetRiskLevel() float64 {
	srs.mu.RLock()
	defer srs.mu.RUnlock()
	
	return srs.RiskManager.CurrentRisk
}

// Shutdown gracefully shuts down the stress-resistant strategy
func (srs *StressResistantStrategy) Shutdown() {
	srs.cancel()
}
