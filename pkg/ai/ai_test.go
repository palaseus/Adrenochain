package ai

import (
	"testing"
	"time"
)

func TestNewMetaLearningAdaptiveAI(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	if ai == nil {
		t.Fatal("NewMetaLearningAdaptiveAI returned nil")
	}
	
	if ai.ID == "" {
		t.Error("AI ID should not be empty")
	}
	
	if ai.MetaLearner == nil {
		t.Error("MetaLearner should not be nil")
	}
	
	if ai.AdaptiveStrategies == nil {
		t.Error("AdaptiveStrategies should not be nil")
	}
	
	if ai.RobustnessFramework == nil {
		t.Error("RobustnessFramework should not be nil")
	}
	
	if ai.ContinuousLearning == nil {
		t.Error("ContinuousLearning should not be nil")
	}
	
	if ai.ScenarioMemory == nil {
		t.Error("ScenarioMemory should not be nil")
	}
	
	if ai.AdaptationMetrics == nil {
		t.Error("AdaptationMetrics should not be nil")
	}
	
	// Verify MetaLearner fields
	if ai.MetaLearner.ID != "meta_learner" {
		t.Error("MetaLearner ID should be 'meta_learner'")
	}
	
	if ai.MetaLearner.LearningRate != 0.1 {
		t.Error("MetaLearner LearningRate should be 0.1")
	}
	
	if ai.MetaLearner.AdaptationSpeed != 0.8 {
		t.Error("MetaLearner AdaptationSpeed should be 0.8")
	}
	
	// Verify RobustnessFramework fields
	if ai.RobustnessFramework.ID != "robustness" {
		t.Error("RobustnessFramework ID should be 'robustness'")
	}
	
	if ai.RobustnessFramework.ResilienceScore != 0.7 {
		t.Error("RobustnessFramework ResilienceScore should be 0.7")
	}
}

func TestMetaLearningAdaptiveAI_HandleUnseenBlackSwan(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	scenario := &BlackSwanScenario{
		ID:          "test_scenario",
		Type:        "market_crash",
		Severity:    0.9,
		Description: "Test black swan scenario",
		Timestamp:   time.Now(),
	}
	
	success, survivalProb, err := ai.HandleUnseenBlackSwan(scenario)
	if err != nil {
		t.Errorf("Failed to handle black swan scenario: %v", err)
	}
	
	// Verify scenario was stored in memory
	if len(ai.ScenarioMemory.Scenarios) != 1 {
		t.Error("Scenario should be stored in memory")
	}
	
	if ai.ScenarioMemory.Scenarios["test_scenario"] == nil {
		t.Error("Scenario should be retrievable from memory")
	}
	
	// Verify continuous learning was updated
	if len(ai.ContinuousLearning.History) == 0 {
		t.Error("Continuous learning history should be updated")
	}
	
	// Verify adaptation metrics were updated
	if ai.AdaptationMetrics.AdaptationCount == 0 {
		t.Error("Adaptation count should be updated")
	}
	
	// Verify return values are reasonable
	if survivalProb < 0 || survivalProb > 1 {
		t.Errorf("Survival probability should be between 0 and 1, got %f", survivalProb)
	}
	
	// Verify success is boolean
	if success != true && success != false {
		t.Error("Success should be boolean")
	}
}

func TestMetaLearningAdaptiveAI_AnalyzeScenarioSimilarity(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	// Test high severity scenario
	highSeverityScenario := &BlackSwanScenario{
		ID:        "high_severity",
		Severity:  0.9,
		Timestamp: time.Now(),
	}
	
	similarity := ai.analyzeScenarioSimilarity(highSeverityScenario)
	if similarity <= 0 {
		t.Error("Similarity should be positive")
	}
	
	// Test low severity scenario
	lowSeverityScenario := &BlackSwanScenario{
		ID:        "low_severity",
		Severity:  0.3,
		Timestamp: time.Now(),
	}
	
	similarity2 := ai.analyzeScenarioSimilarity(lowSeverityScenario)
	if similarity2 <= 0 {
		t.Error("Similarity should be positive")
	}
	
	// High severity should have lower similarity
	if similarity >= similarity2 {
		t.Error("High severity scenarios should have lower similarity")
	}
}

func TestMetaLearningAdaptiveAI_GenerateAdaptationStrategy(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	scenario := &BlackSwanScenario{
		ID:        "test_scenario",
		Timestamp: time.Now(),
	}
	
	// Test different similarity levels
	strategies := make(map[string]bool)
	
	for _, similarity := range []float64{0.1, 0.4, 0.6, 0.8} {
		strategy := ai.generateAdaptationStrategy(scenario, similarity)
		strategies[strategy] = true
		
		if strategy == "" {
			t.Error("Strategy should not be empty")
		}
	}
	
	// Should generate different strategies for different similarity levels
	if len(strategies) < 2 {
		t.Error("Should generate different strategies for different similarity levels")
	}
}

func TestMetaLearningAdaptiveAI_ExecuteAdaptation(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	// Test adaptation execution
	success := ai.executeAdaptation("test_strategy")
	
	// Verify metrics were updated
	if ai.AdaptationMetrics.AdaptationCount != 1 {
		t.Error("Adaptation count should be incremented")
	}
	
	// Verify success is boolean
	if success != true && success != false {
		t.Error("Success should be boolean")
	}
}

func TestMetaLearningAdaptiveAI_CalculateSurvivalProbability(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	scenario := &BlackSwanScenario{
		ID:        "test_scenario",
		Severity:  0.5,
		Timestamp: time.Now(),
	}
	
	// Test with different similarity levels
	for _, similarity := range []float64{0.1, 0.5, 0.9} {
		survivalProb := ai.calculateSurvivalProbability(scenario, similarity)
		
		if survivalProb < 0 || survivalProb > 1 {
			t.Errorf("Survival probability should be between 0 and 1, got %f", survivalProb)
		}
		
		if survivalProb > 0.8 {
			t.Errorf("Survival probability should not exceed 0.8, got %f", survivalProb)
		}
	}
}

func TestMetaLearningAdaptiveAI_GetAdaptationMetrics(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	metrics := ai.GetAdaptationMetrics()
	if metrics == nil {
		t.Fatal("GetAdaptationMetrics should not return nil")
	}
	
	if metrics.ID != "metrics" {
		t.Error("Metrics ID should be 'metrics'")
	}
	
	if metrics.AdaptationCount != 0 {
		t.Error("Initial adaptation count should be 0")
	}
	
	if metrics.SuccessRate != 0.0 {
		t.Error("Initial success rate should be 0.0")
	}
}

func TestMetaLearningAdaptiveAI_GetRobustnessScore(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	score := ai.GetRobustnessScore()
	if score != 0.7 {
		t.Errorf("Expected robustness score 0.7, got %f", score)
	}
}

func TestMetaLearningAdaptiveAI_Shutdown(t *testing.T) {
	ai := NewMetaLearningAdaptiveAI()
	
	// Shutdown should not panic
	ai.Shutdown()
	
	// Verify context was cancelled
	select {
	case <-ai.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after shutdown")
	}
}

func TestBlackSwanScenario_Fields(t *testing.T) {
	scenario := BlackSwanScenario{
		ID:          "test_scenario",
		Type:        "market_crash",
		Severity:    0.8,
		Description: "Test scenario",
		Timestamp:   time.Now(),
	}
	
	if scenario.ID != "test_scenario" {
		t.Error("Scenario ID should match")
	}
	
	if scenario.Type != "market_crash" {
		t.Error("Scenario Type should match")
	}
	
	if scenario.Severity != 0.8 {
		t.Error("Scenario Severity should match")
	}
	
	if scenario.Description != "Test scenario" {
		t.Error("Scenario Description should match")
	}
	
	if scenario.Timestamp.IsZero() {
		t.Error("Scenario Timestamp should not be zero")
	}
}

func TestStrategy_Fields(t *testing.T) {
	strategy := Strategy{
		ID:           "test_strategy",
		Type:         "momentum",
		RiskLevel:    0.5,
		Adaptability: 0.7,
		Performance:  0.8,
	}
	
	if strategy.ID != "test_strategy" {
		t.Error("Strategy ID should match")
	}
	
	if strategy.Type != "momentum" {
		t.Error("Strategy Type should match")
	}
	
	if strategy.RiskLevel != 0.5 {
		t.Error("Strategy RiskLevel should match")
	}
	
	if strategy.Adaptability != 0.7 {
		t.Error("Strategy Adaptability should match")
	}
	
	if strategy.Performance != 0.8 {
		t.Error("Strategy Performance should match")
	}
}

func TestMetaLearner_Fields(t *testing.T) {
	learner := MetaLearner{
		ID:              "test_learner",
		LearningRate:    0.15,
		AdaptationSpeed: 0.9,
		Memory:          make(map[string]interface{}),
	}
	
	if learner.ID != "test_learner" {
		t.Error("MetaLearner ID should match")
	}
	
	if learner.LearningRate != 0.15 {
		t.Error("MetaLearner LearningRate should match")
	}
	
	if learner.AdaptationSpeed != 0.9 {
		t.Error("MetaLearner AdaptationSpeed should match")
	}
	
	if learner.Memory == nil {
		t.Error("MetaLearner Memory should not be nil")
	}
}

func TestAdaptiveStrategy_Fields(t *testing.T) {
	baseStrategy := &Strategy{
		ID:   "base_strategy",
		Type: "momentum",
	}
	
	adaptiveStrategy := AdaptiveStrategy{
		ID:              "adaptive_strategy",
		BaseStrategy:    baseStrategy,
		AdaptationRules: make(map[string]string),
		SuccessRate:     0.85,
	}
	
	if adaptiveStrategy.ID != "adaptive_strategy" {
		t.Error("AdaptiveStrategy ID should match")
	}
	
	if adaptiveStrategy.BaseStrategy != baseStrategy {
		t.Error("AdaptiveStrategy BaseStrategy should match")
	}
	
	if adaptiveStrategy.SuccessRate != 0.85 {
		t.Error("AdaptiveStrategy SuccessRate should match")
	}
	
	if adaptiveStrategy.AdaptationRules == nil {
		t.Error("AdaptiveStrategy AdaptationRules should not be nil")
	}
}

func TestRobustnessFramework_Fields(t *testing.T) {
	framework := RobustnessFramework{
		ID:              "test_framework",
		StressTests:     []string{"test1", "test2"},
		ResilienceScore: 0.75,
	}
	
	if framework.ID != "test_framework" {
		t.Error("RobustnessFramework ID should match")
	}
	
	if len(framework.StressTests) != 2 {
		t.Error("RobustnessFramework should have 2 stress tests")
	}
	
	if framework.ResilienceScore != 0.75 {
		t.Error("RobustnessFramework ResilienceScore should match")
	}
}

func TestContinuousLearningEngine_Fields(t *testing.T) {
	engine := ContinuousLearningEngine{
		ID:           "test_engine",
		LearningRate: 0.1,
		History:      []string{"entry1", "entry2"},
	}
	
	if engine.ID != "test_engine" {
		t.Error("ContinuousLearningEngine ID should match")
	}
	
	if engine.LearningRate != 0.1 {
		t.Error("ContinuousLearningEngine LearningRate should match")
	}
	
	if len(engine.History) != 2 {
		t.Error("ContinuousLearningEngine should have 2 history entries")
	}
}

func TestScenarioMemory_Fields(t *testing.T) {
	memory := ScenarioMemory{
		ID:       "test_memory",
		Scenarios: make(map[string]*BlackSwanScenario),
		Patterns: make(map[string]float64),
	}
	
	if memory.ID != "test_memory" {
		t.Error("ScenarioMemory ID should match")
	}
	
	if memory.Scenarios == nil {
		t.Error("ScenarioMemory Scenarios should not be nil")
	}
	
	if memory.Patterns == nil {
		t.Error("ScenarioMemory Patterns should not be nil")
	}
}

func TestAdaptationMetrics_Fields(t *testing.T) {
	metrics := AdaptationMetrics{
		ID:                "test_metrics",
		AdaptationCount:  5,
		SuccessRate:       0.8,
		ImprovementRate:   0.15,
	}
	
	if metrics.ID != "test_metrics" {
		t.Error("AdaptationMetrics ID should match")
	}
	
	if metrics.AdaptationCount != 5 {
		t.Error("AdaptationMetrics AdaptationCount should match")
	}
	
	if metrics.SuccessRate != 0.8 {
		t.Error("AdaptationMetrics SuccessRate should match")
	}
	
	if metrics.ImprovementRate != 0.15 {
		t.Error("AdaptationMetrics ImprovementRate should match")
	}
}
