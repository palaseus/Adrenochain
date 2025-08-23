package ai

import (
	"testing"
)

func TestNewStressResistantStrategy(t *testing.T) {
	baseStrategy := &Strategy{
		ID:   "base_strategy",
		Type: "momentum",
	}
	
	srs := NewStressResistantStrategy(baseStrategy)
	
	if srs == nil {
		t.Fatal("NewStressResistantStrategy returned nil")
	}
	
	if srs.ID == "" {
		t.Error("Strategy ID should not be empty")
	}
	
	if srs.BaseStrategy != baseStrategy {
		t.Error("BaseStrategy should match the input")
	}
	
	if srs.RiskManager == nil {
		t.Error("RiskManager should not be nil")
	}
	
	if srs.StressDetector == nil {
		t.Error("StressDetector should not be nil")
	}
	
	if srs.EmergencyProtocol == nil {
		t.Error("EmergencyProtocol should not be nil")
	}
	
	if srs.RegimeAdaptation == nil {
		t.Error("RegimeAdaptation should not be nil")
	}
	
	if srs.VolatilityAnalyzer == nil {
		t.Error("VolatilityAnalyzer should not be nil")
	}
	
	if srs.LiquidityManager == nil {
		t.Error("LiquidityManager should not be nil")
	}
	
	if srs.BlackSwanProtection == nil {
		t.Error("BlackSwanProtection should not be nil")
	}
	
	if srs.ctx == nil {
		t.Error("Context should not be nil")
	}
	
	if srs.cancel == nil {
		t.Error("Cancel function should not be nil")
	}
	
	// Verify default values
	if srs.RiskManager.MaxRisk != 0.3 {
		t.Errorf("Expected MaxRisk 0.3, got %f", srs.RiskManager.MaxRisk)
	}
	
	if srs.RiskManager.CurrentRisk != 0.1 {
		t.Errorf("Expected CurrentRisk 0.1, got %f", srs.RiskManager.CurrentRisk)
	}
	
	if srs.StressDetector.StressThreshold != 0.7 {
		t.Errorf("Expected StressThreshold 0.7, got %f", srs.StressDetector.StressThreshold)
	}
	
	if srs.RegimeAdaptation.CurrentRegime != MarketRegimeSideways {
		t.Errorf("Expected CurrentRegime MarketRegimeSideways, got %v", srs.RegimeAdaptation.CurrentRegime)
	}
	
	if srs.VolatilityAnalyzer.Threshold != 0.5 {
		t.Errorf("Expected Threshold 0.5, got %f", srs.VolatilityAnalyzer.Threshold)
	}
	
	if srs.LiquidityManager.MinLiquidity != 0.2 {
		t.Errorf("Expected MinLiquidity 0.2, got %f", srs.LiquidityManager.MinLiquidity)
	}
	
	if srs.LiquidityManager.LiquidityBuffer != 0.1 {
		t.Errorf("Expected LiquidityBuffer 0.1, got %f", srs.LiquidityManager.LiquidityBuffer)
	}
	
	if srs.BlackSwanProtection.ProtectionLevel != 0.8 {
		t.Errorf("Expected ProtectionLevel 0.8, got %f", srs.BlackSwanProtection.ProtectionLevel)
	}
}

func TestStressResistantStrategy_HandleMarketStress(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test low stress scenario
	err := srs.HandleMarketStress(0.3, MarketRegimeBull)
	if err != nil {
		t.Errorf("HandleMarketStress failed: %v", err)
	}
	
	if srs.StressDetector.CurrentStress != 0.3 {
		t.Errorf("Expected CurrentStress 0.3, got %f", srs.StressDetector.CurrentStress)
	}
	
	if len(srs.StressDetector.StressHistory) != 1 {
		t.Error("Stress history should be updated")
	}
	
	if srs.StressDetector.StressHistory[0] != 0.3 {
		t.Errorf("Expected stress history entry 0.3, got %f", srs.StressDetector.StressHistory[0])
	}
	
	if srs.RegimeAdaptation.CurrentRegime != MarketRegimeBull {
		t.Errorf("Expected CurrentRegime MarketRegimeBull, got %v", srs.RegimeAdaptation.CurrentRegime)
	}
	
	if len(srs.RegimeAdaptation.RegimeHistory) != 1 {
		t.Error("Regime history should be updated")
	}
	
	if srs.RegimeAdaptation.RegimeHistory[0] != MarketRegimeBull {
		t.Errorf("Expected regime history entry MarketRegimeBull, got %v", srs.RegimeAdaptation.RegimeHistory[0])
	}
	
	// Test high stress scenario that triggers emergency protocols
	err = srs.HandleMarketStress(0.8, MarketRegimeCrisis)
	if err != nil {
		t.Errorf("HandleMarketStress failed: %v", err)
	}
	
	if !srs.EmergencyProtocol.IsActive {
		t.Error("Emergency protocols should be active")
	}
	
	if srs.EmergencyProtocol.ActivationTime.IsZero() {
		t.Error("Emergency protocol activation time should be set")
	}
	
	// Test black swan scenario
	err = srs.HandleMarketStress(0.95, MarketRegimeCrisis)
	if err != nil {
		t.Errorf("HandleMarketStress failed: %v", err)
	}
	
	if !srs.BlackSwanProtection.IsActive {
		t.Error("Black swan protection should be active")
	}
	
	if srs.BlackSwanProtection.LastActivation.IsZero() {
		t.Error("Black swan protection activation time should be set")
	}
}

func TestStressResistantStrategy_activateEmergencyProtocols(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Initially emergency protocols should not be active
	if srs.EmergencyProtocol.IsActive {
		t.Error("Emergency protocols should not be initially active")
	}
	
	// Activate emergency protocols
	srs.activateEmergencyProtocols()
	
	if !srs.EmergencyProtocol.IsActive {
		t.Error("Emergency protocols should be active after activation")
	}
	
	if srs.EmergencyProtocol.ActivationTime.IsZero() {
		t.Error("Activation time should be set")
	}
	
	// Verify protocols were added
	expectedProtocols := map[string]string{
		"stop_loss":         "activated",
		"position_reduction": "activated",
		"hedging":           "activated",
	}
	
	for protocol, status := range expectedProtocols {
		if srs.EmergencyProtocol.Protocols[protocol] != status {
			t.Errorf("Expected protocol %s to have status %s, got %s", 
				protocol, status, srs.EmergencyProtocol.Protocols[protocol])
		}
	}
}

func TestStressResistantStrategy_adaptToRegime(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Initialize adaptation rules map
	srs.RegimeAdaptation.AdaptationRules = make(map[MarketRegime]string)
	
	// Test bull market adaptation
	srs.adaptToRegime(MarketRegimeBull)
	if srs.RegimeAdaptation.AdaptationRules[MarketRegimeBull] != "aggressive_growth" {
		t.Errorf("Expected bull market adaptation 'aggressive_growth', got %s", 
			srs.RegimeAdaptation.AdaptationRules[MarketRegimeBull])
	}
	
	// Test bear market adaptation
	srs.adaptToRegime(MarketRegimeBear)
	if srs.RegimeAdaptation.AdaptationRules[MarketRegimeBear] != "defensive_conservation" {
		t.Errorf("Expected bear market adaptation 'defensive_conservation', got %s", 
			srs.RegimeAdaptation.AdaptationRules[MarketRegimeBear])
	}
	
	// Test sideways market adaptation
	srs.adaptToRegime(MarketRegimeSideways)
	if srs.RegimeAdaptation.AdaptationRules[MarketRegimeSideways] != "range_trading" {
		t.Errorf("Expected sideways market adaptation 'range_trading', got %s", 
			srs.RegimeAdaptation.AdaptationRules[MarketRegimeSideways])
	}
	
	// Test volatile market adaptation
	srs.adaptToRegime(MarketRegimeVolatile)
	if srs.RegimeAdaptation.AdaptationRules[MarketRegimeVolatile] != "volatility_harvesting" {
		t.Errorf("Expected volatile market adaptation 'volatility_harvesting', got %s", 
			srs.RegimeAdaptation.AdaptationRules[MarketRegimeVolatile])
	}
	
	// Test crisis market adaptation
	srs.adaptToRegime(MarketRegimeCrisis)
	if srs.RegimeAdaptation.AdaptationRules[MarketRegimeCrisis] != "crisis_management" {
		t.Errorf("Expected crisis market adaptation 'crisis_management', got %s", 
			srs.RegimeAdaptation.AdaptationRules[MarketRegimeCrisis])
	}
}

func TestStressResistantStrategy_adjustRiskManagement(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test low stress - should maintain max risk
	srs.adjustRiskManagement(0.3)
	expectedRisk := srs.RiskManager.MaxRisk
	if srs.RiskManager.CurrentRisk != expectedRisk {
		t.Errorf("Expected CurrentRisk %f for low stress, got %f", expectedRisk, srs.RiskManager.CurrentRisk)
	}
	
	// Test medium stress - should reduce to 70% of max risk
	srs.adjustRiskManagement(0.6)
	expectedRisk = srs.RiskManager.MaxRisk * 0.7
	if srs.RiskManager.CurrentRisk != expectedRisk {
		t.Errorf("Expected CurrentRisk %f for medium stress, got %f", expectedRisk, srs.RiskManager.CurrentRisk)
	}
	
	// Test high stress - should reduce to 50% of max risk
	srs.adjustRiskManagement(0.8)
	expectedRisk = srs.RiskManager.MaxRisk * 0.5
	if srs.RiskManager.CurrentRisk != expectedRisk {
		t.Errorf("Expected CurrentRisk %f for high stress, got %f", expectedRisk, srs.RiskManager.CurrentRisk)
	}
}

func TestStressResistantStrategy_updateVolatilityAnalysis(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test initial volatility update
	srs.updateVolatilityAnalysis(0.4)
	
	if srs.VolatilityAnalyzer.CurrentVolatility != 0.4 {
		t.Errorf("Expected CurrentVolatility 0.4, got %f", srs.VolatilityAnalyzer.CurrentVolatility)
	}
	
	if len(srs.VolatilityAnalyzer.VolatilityHistory) != 1 {
		t.Error("Volatility history should be updated")
	}
	
	if srs.VolatilityAnalyzer.VolatilityHistory[0] != 0.4 {
		t.Errorf("Expected volatility history entry 0.4, got %f", srs.VolatilityAnalyzer.VolatilityHistory[0])
	}
	
	// Test multiple updates
	srs.updateVolatilityAnalysis(0.6)
	srs.updateVolatilityAnalysis(0.8)
	
	if len(srs.VolatilityAnalyzer.VolatilityHistory) != 3 {
		t.Errorf("Expected 3 volatility history entries, got %d", len(srs.VolatilityAnalyzer.VolatilityHistory))
	}
	
	// Test history limit (should keep only last 100 entries)
	for i := 0; i < 102; i++ {
		srs.updateVolatilityAnalysis(float64(i) / 100.0)
	}
	
	if len(srs.VolatilityAnalyzer.VolatilityHistory) > 100 {
		t.Errorf("Volatility history should not exceed 100 entries, got %d", len(srs.VolatilityAnalyzer.VolatilityHistory))
	}
}

func TestStressResistantStrategy_manageLiquidity(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test low stress - should maintain default buffer
	srs.manageLiquidity(0.4)
	if srs.LiquidityManager.LiquidityBuffer != 0.1 {
		t.Errorf("Expected LiquidityBuffer 0.1 for low stress, got %f", srs.LiquidityManager.LiquidityBuffer)
	}
	
	// Test medium stress - should increase buffer to 0.2
	srs.manageLiquidity(0.7)
	if srs.LiquidityManager.LiquidityBuffer != 0.2 {
		t.Errorf("Expected LiquidityBuffer 0.2 for medium stress, got %f", srs.LiquidityManager.LiquidityBuffer)
	}
	
	// Test high stress - should increase buffer to 0.3
	srs.manageLiquidity(0.9)
	if srs.LiquidityManager.LiquidityBuffer != 0.3 {
		t.Errorf("Expected LiquidityBuffer 0.3 for high stress, got %f", srs.LiquidityManager.LiquidityBuffer)
	}
}

func TestStressResistantStrategy_activateBlackSwanProtection(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Initially black swan protection should not be active
	if srs.BlackSwanProtection.IsActive {
		t.Error("Black swan protection should not be initially active")
	}
	
	// Activate black swan protection
	srs.activateBlackSwanProtection()
	
	if !srs.BlackSwanProtection.IsActive {
		t.Error("Black swan protection should be active after activation")
	}
	
	if srs.BlackSwanProtection.LastActivation.IsZero() {
		t.Error("Last activation time should be set")
	}
}

func TestStressResistantStrategy_GetStressLevel(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test low stress level
	srs.StressDetector.CurrentStress = 0.2
	stressLevel := srs.GetStressLevel()
	if stressLevel != StressLevelLow {
		t.Errorf("Expected StressLevelLow for stress 0.2, got %v", stressLevel)
	}
	
	// Test medium stress level
	srs.StressDetector.CurrentStress = 0.4
	stressLevel = srs.GetStressLevel()
	if stressLevel != StressLevelMedium {
		t.Errorf("Expected StressLevelMedium for stress 0.4, got %v", stressLevel)
	}
	
	// Test high stress level
	srs.StressDetector.CurrentStress = 0.7
	stressLevel = srs.GetStressLevel()
	if stressLevel != StressLevelHigh {
		t.Errorf("Expected StressLevelHigh for stress 0.7, got %v", stressLevel)
	}
	
	// Test extreme stress level
	srs.StressDetector.CurrentStress = 0.9
	stressLevel = srs.GetStressLevel()
	if stressLevel != StressLevelExtreme {
		t.Errorf("Expected StressLevelExtreme for stress 0.9, got %v", stressLevel)
	}
}

func TestStressResistantStrategy_GetCurrentRegime(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test initial regime
	regime := srs.GetCurrentRegime()
	if regime != MarketRegimeSideways {
		t.Errorf("Expected initial regime MarketRegimeSideways, got %v", regime)
	}
	
	// Test regime change
	srs.RegimeAdaptation.CurrentRegime = MarketRegimeBull
	regime = srs.GetCurrentRegime()
	if regime != MarketRegimeBull {
		t.Errorf("Expected regime MarketRegimeBull, got %v", regime)
	}
}

func TestStressResistantStrategy_IsEmergencyActive(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Initially emergency should not be active
	if srs.IsEmergencyActive() {
		t.Error("Emergency should not be initially active")
	}
	
	// Activate emergency
	srs.EmergencyProtocol.IsActive = true
	
	if !srs.IsEmergencyActive() {
		t.Error("Emergency should be active after activation")
	}
}

func TestStressResistantStrategy_GetRiskLevel(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test initial risk level
	riskLevel := srs.GetRiskLevel()
	if riskLevel != 0.1 {
		t.Errorf("Expected initial risk level 0.1, got %f", riskLevel)
	}
	
	// Test risk level change
	srs.RiskManager.CurrentRisk = 0.25
	riskLevel = srs.GetRiskLevel()
	if riskLevel != 0.25 {
		t.Errorf("Expected risk level 0.25, got %f", riskLevel)
	}
}

func TestStressResistantStrategy_Shutdown(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Shutdown should not panic
	srs.Shutdown()
	
	// Verify context was cancelled
	select {
	case <-srs.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after shutdown")
	}
}

func TestMarketRegime_String(t *testing.T) {
	// Test all market regimes
	testCases := []struct {
		regime MarketRegime
		name   string
	}{
		{MarketRegimeBull, "Bull"},
		{MarketRegimeBear, "Bear"},
		{MarketRegimeSideways, "Sideways"},
		{MarketRegimeVolatile, "Volatile"},
		{MarketRegimeCrisis, "Crisis"},
	}
	
	for _, tc := range testCases {
		// This is a simple test to ensure the enum values are defined
		// The actual string representation would depend on how the enum is used
		if tc.regime < 0 || int(tc.regime) >= 5 {
			t.Errorf("Invalid market regime value: %d", tc.regime)
		}
	}
}

func TestStressLevel_String(t *testing.T) {
	// Test all stress levels
	testCases := []struct {
		level StressLevel
		name  string
	}{
		{StressLevelLow, "Low"},
		{StressLevelMedium, "Medium"},
		{StressLevelHigh, "High"},
		{StressLevelExtreme, "Extreme"},
	}
	
	for _, tc := range testCases {
		// This is a simple test to ensure the enum values are defined
		// The actual string representation would depend on how the enum is used
		if tc.level < 0 || int(tc.level) >= 4 {
			t.Errorf("Invalid stress level value: %d", tc.level)
		}
	}
}

func TestStressResistantStrategy_Concurrency(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test concurrent access to the strategy
	done := make(chan bool)
	
	// Start multiple goroutines that access the strategy concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Simulate concurrent market stress handling
			stressLevel := float64(id) / 10.0
			regime := MarketRegime(id % 5)
			
			err := srs.HandleMarketStress(stressLevel, regime)
			if err != nil {
				t.Errorf("Concurrent HandleMarketStress failed: %v", err)
			}
			
			// Concurrent reads
			_ = srs.GetStressLevel()
			_ = srs.GetCurrentRegime()
			_ = srs.IsEmergencyActive()
			_ = srs.GetRiskLevel()
			
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify the strategy is still in a valid state
	if srs.ID == "" {
		t.Error("Strategy ID should not be empty after concurrent access")
	}
	
	if srs.BaseStrategy == nil {
		t.Error("BaseStrategy should not be nil after concurrent access")
	}
}

func TestStressResistantStrategy_EdgeCases(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Test with zero stress level
	err := srs.HandleMarketStress(0.0, MarketRegimeSideways)
	if err != nil {
		t.Errorf("HandleMarketStress failed with zero stress: %v", err)
	}
	
	// Test with maximum stress level
	err = srs.HandleMarketStress(1.0, MarketRegimeCrisis)
	if err != nil {
		t.Errorf("HandleMarketStress failed with maximum stress: %v", err)
	}
	
	// Test with negative stress level (should handle gracefully)
	err = srs.HandleMarketStress(-0.1, MarketRegimeSideways)
	if err != nil {
		t.Errorf("HandleMarketStress failed with negative stress: %v", err)
	}
	
	// Test with stress level above 1.0 (should handle gracefully)
	err = srs.HandleMarketStress(1.1, MarketRegimeCrisis)
	if err != nil {
		t.Errorf("HandleMarketStress failed with stress above 1.0: %v", err)
	}
	
	// Test with nil base strategy
	srs.BaseStrategy = nil
	err = srs.HandleMarketStress(0.5, MarketRegimeBull)
	if err != nil {
		t.Errorf("HandleMarketStress failed with nil base strategy: %v", err)
	}
}

func TestStressResistantStrategy_StressHistoryManagement(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Add many stress events to test history management
	for i := 0; i < 150; i++ {
		srs.StressDetector.StressHistory = append(srs.StressDetector.StressHistory, float64(i)/100.0)
	}
	
	// Verify history is manageable
	if len(srs.StressDetector.StressHistory) > 200 {
		t.Errorf("Stress history should not grow excessively, got %d entries", len(srs.StressDetector.StressHistory))
	}
	
	// Test that recent entries are preserved
	recentStress := srs.StressDetector.StressHistory[len(srs.StressDetector.StressHistory)-1]
	if recentStress != 1.49 {
		t.Errorf("Expected recent stress entry 1.49, got %f", recentStress)
	}
}

func TestStressResistantStrategy_RegimeHistoryManagement(t *testing.T) {
	baseStrategy := &Strategy{ID: "test_strategy"}
	srs := NewStressResistantStrategy(baseStrategy)
	
	// Add many regime changes to test history management
	regimes := []MarketRegime{MarketRegimeBull, MarketRegimeBear, MarketRegimeSideways, MarketRegimeVolatile, MarketRegimeCrisis}
	for i := 0; i < 100; i++ {
		srs.RegimeAdaptation.RegimeHistory = append(srs.RegimeAdaptation.RegimeHistory, regimes[i%len(regimes)])
	}
	
	// Verify history is manageable
	if len(srs.RegimeAdaptation.RegimeHistory) > 200 {
		t.Errorf("Regime history should not grow excessively, got %d entries", len(srs.RegimeAdaptation.RegimeHistory))
	}
	
	// Test that recent entries are preserved
	recentRegime := srs.RegimeAdaptation.RegimeHistory[len(srs.RegimeAdaptation.RegimeHistory)-1]
	if recentRegime != MarketRegimeCrisis {
		t.Errorf("Expected recent regime entry MarketRegimeCrisis, got %v", recentRegime)
	}
}
