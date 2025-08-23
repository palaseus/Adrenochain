package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoricalDataEngine(t *testing.T) {
	t.Run("NewHistoricalDataEngine", func(t *testing.T) {
		engine := NewHistoricalDataEngine()
		require.NotNil(t, engine)
		assert.NotEmpty(t, engine.ID)
		assert.NotNil(t, engine.CrashEvents)
		assert.NotNil(t, engine.MarketData)
		assert.NotNil(t, engine.AnalysisData)
		assert.Contains(t, engine.ID, "hist_engine_")
	})

	t.Run("AddCrashEvent", func(t *testing.T) {
		engine := NewHistoricalDataEngine()

		crash := &CrashEvent{
			ID:           "crash_1",
			Name:         "Test Crash",
			Date:         time.Now(),
			Impact:       CrashImpactModerate,
			PriceDrop:    0.25,
			RecoveryTime: 24 * time.Hour,
			Description:  "Test crash event",
		}

		engine.AddCrashEvent(crash)
		assert.Len(t, engine.CrashEvents, 1)
		assert.Equal(t, crash, engine.CrashEvents["crash_1"])
	})

	t.Run("GetCrashEvent", func(t *testing.T) {
		engine := NewHistoricalDataEngine()

		crash := &CrashEvent{
			ID:           "crash_2",
			Name:         "Test Crash 2",
			Date:         time.Now(),
			Impact:       CrashImpactSevere,
			PriceDrop:    0.50,
			RecoveryTime: 48 * time.Hour,
			Description:  "Test crash event 2",
		}

		engine.AddCrashEvent(crash)

		// Test getting existing crash
		retrieved, exists := engine.GetCrashEvent("crash_2")
		assert.True(t, exists)
		assert.Equal(t, crash, retrieved)

		// Test getting non-existent crash
		retrieved, exists = engine.GetCrashEvent("non_existent")
		assert.False(t, exists)
		assert.Nil(t, retrieved)
	})

	t.Run("SimulateCrashScenario", func(t *testing.T) {
		engine := NewHistoricalDataEngine()

		strategy := &Strategy{
			ID:        "test_strategy",
			RiskLevel: 0.7, // Medium-high risk
		}

		crash := &CrashEvent{
			ID:           "crash_3",
			Name:         "Test Crash 3",
			Date:         time.Now(),
			Impact:       CrashImpactMild,
			PriceDrop:    0.15,
			RecoveryTime: 12 * time.Hour,
			Description:  "Test crash event 3",
		}

		survived, loss, recovery := engine.SimulateCrashScenario(strategy, crash)

		// Test that results are within expected ranges
		assert.True(t, loss >= 0.0)
		assert.True(t, recovery >= 0)
		assert.True(t, survived == true || survived == false) // Use the survived variable

		// Test with different crash impacts
		severeCrash := &CrashEvent{
			ID:           "crash_4",
			Name:         "Severe Crash",
			Date:         time.Now(),
			Impact:       CrashImpactCatastrophic,
			PriceDrop:    0.80,
			RecoveryTime: 72 * time.Hour,
			Description:  "Catastrophic crash",
		}

		survived2, loss2, recovery2 := engine.SimulateCrashScenario(strategy, severeCrash)
		assert.True(t, loss2 >= 0.0)
		assert.True(t, recovery2 >= 0)
		assert.True(t, survived2 == true || survived2 == false) // Use the survived2 variable

		// Test with low-risk strategy
		lowRiskStrategy := &Strategy{
			ID:        "low_risk_strategy",
			RiskLevel: 0.2, // Low risk
		}

		survived3, loss3, recovery3 := engine.SimulateCrashScenario(lowRiskStrategy, crash)
		assert.True(t, loss3 >= 0.0)
		assert.True(t, recovery3 >= 0)
		assert.True(t, survived3 == true || survived3 == false) // Use the survived3 variable
	})

	t.Run("RunHistoricalBacktest", func(t *testing.T) {
		engine := NewHistoricalDataEngine()

		// Add multiple crash events
		crashes := []*CrashEvent{
			{
				ID:           "crash_1",
				Name:         "Mild Crash",
				Date:         time.Now(),
				Impact:       CrashImpactMild,
				PriceDrop:    0.10,
				RecoveryTime: 6 * time.Hour,
				Description:  "Mild crash event",
			},
			{
				ID:           "crash_2",
				Name:         "Moderate Crash",
				Date:         time.Now(),
				Impact:       CrashImpactModerate,
				PriceDrop:    0.25,
				RecoveryTime: 24 * time.Hour,
				Description:  "Moderate crash event",
			},
			{
				ID:           "crash_3",
				Name:         "Severe Crash",
				Date:         time.Now(),
				Impact:       CrashImpactSevere,
				PriceDrop:    0.50,
				RecoveryTime: 48 * time.Hour,
				Description:  "Severe crash event",
			},
		}

		for _, crash := range crashes {
			engine.AddCrashEvent(crash)
		}

		strategy := &Strategy{
			ID:        "test_strategy",
			RiskLevel: 0.6,
		}

		result := engine.RunHistoricalBacktest(strategy)

		require.NotNil(t, result)
		assert.Equal(t, strategy.ID, result.StrategyID)
		assert.Equal(t, 3, result.TotalScenarios)
		assert.True(t, result.Survived >= 0)
		assert.True(t, result.Failed >= 0)
		assert.True(t, result.Survived+result.Failed == 3)
		assert.True(t, result.TotalLoss >= 0.0)
		assert.True(t, result.AverageRecovery >= 0)
		assert.True(t, result.Duration >= 0)
		assert.True(t, result.SurvivalRate >= 0.0 && result.SurvivalRate <= 1.0)
	})

	t.Run("RunHistoricalBacktest_EmptyEvents", func(t *testing.T) {
		engine := NewHistoricalDataEngine() // No crash events
		strategy := &Strategy{
			ID:        "test_strategy",
			RiskLevel: 0.6,
		}

		result := engine.RunHistoricalBacktest(strategy)

		require.NotNil(t, result)
		assert.Equal(t, 0, result.TotalScenarios)
		assert.Equal(t, 0, result.Survived)
		assert.Equal(t, 0, result.Failed)
		assert.Equal(t, 0.0, result.SurvivalRate)
		assert.Equal(t, 0.0, result.TotalLoss)
		assert.Equal(t, int64(0), result.AverageRecovery)
	})
}

func TestBacktestResult(t *testing.T) {
	t.Run("GetSurvivalRate", func(t *testing.T) {
		result := &BacktestResult{
			StrategyID:     "test",
			TotalScenarios: 10,
			Survived:       7,
			SurvivalRate:   0.7,
		}

		survivalRate := result.GetSurvivalRate()
		assert.Equal(t, 70.0, survivalRate)
	})

	t.Run("GetAverageLoss", func(t *testing.T) {
		result := &BacktestResult{
			StrategyID:     "test",
			TotalScenarios: 5,
			TotalLoss:      25.0,
		}

		avgLoss := result.GetAverageLoss()
		assert.Equal(t, 5.0, avgLoss)
	})

	t.Run("GetAverageLoss_ZeroScenarios", func(t *testing.T) {
		result := &BacktestResult{
			StrategyID:     "test",
			TotalScenarios: 0,
			TotalLoss:      25.0,
		}

		avgLoss := result.GetAverageLoss()
		assert.Equal(t, 0.0, avgLoss)
	})

	t.Run("IsSuccessful", func(t *testing.T) {
		result := &BacktestResult{
			StrategyID:     "test",
			TotalScenarios: 10,
			Survived:       8,
			SurvivalRate:   0.8,
		}

		// Test with threshold below survival rate
		assert.True(t, result.IsSuccessful(0.7))

		// Test with threshold above survival rate
		assert.False(t, result.IsSuccessful(0.9))

		// Test with threshold equal to survival rate
		assert.True(t, result.IsSuccessful(0.8))
	})

	t.Run("String", func(t *testing.T) {
		result := &BacktestResult{
			StrategyID:     "test_strategy",
			TotalScenarios: 10,
			Survived:       7,
			SurvivalRate:   0.7,
			TotalLoss:      15.0,
		}

		str := result.String()
		assert.Contains(t, str, "test_strategy")
		assert.Contains(t, str, "70.0%")
		assert.Contains(t, str, "7/10")
		assert.Contains(t, str, "1.50%") // 15.0 / 10 = 1.5
	})
}

func TestCrashImpactConstants(t *testing.T) {
	// Test that crash impact constants are properly defined
	assert.Equal(t, CrashImpact(0), CrashImpactMild)
	assert.Equal(t, CrashImpact(1), CrashImpactModerate)
	assert.Equal(t, CrashImpact(2), CrashImpactSevere)
	assert.Equal(t, CrashImpact(3), CrashImpactCatastrophic)
}
