package advanced

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRiskAssessor(t *testing.T) {
	t.Run("valid risk assessor", func(t *testing.T) {
		lendingPool := createTestPool()
		updateInterval := 1 * time.Hour
		alertThreshold := 75

		assessor := NewRiskAssessor(lendingPool, updateInterval, alertThreshold)

		require.NotNil(t, assessor)
		assert.Equal(t, lendingPool, assessor.lendingPool)
		assert.Equal(t, updateInterval, assessor.updateInterval)
		assert.Equal(t, alertThreshold, assessor.alertThreshold)
		assert.NotNil(t, assessor.riskFactors)
		assert.NotNil(t, assessor.assessments)
		assert.Len(t, assessor.riskFactors, 6) // 6 default risk factors
	})
}

func TestRiskAssessorDefaultFactors(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("check default risk factors", func(t *testing.T) {
		// Check collateral ratio factor
		factor, err := assessor.GetRiskFactor("collateral_ratio")
		require.NoError(t, err)
		assert.Equal(t, "Collateral Ratio", factor.Name)
		assert.Equal(t, 0.25, factor.Weight)
		assert.Equal(t, 80, factor.Threshold)

		// Check liquidation proximity factor
		factor, err = assessor.GetRiskFactor("liquidation_proximity")
		require.NoError(t, err)
		assert.Equal(t, "Liquidation Proximity", factor.Name)
		assert.Equal(t, 0.20, factor.Weight)
		assert.Equal(t, 70, factor.Threshold)

		// Check market volatility factor
		factor, err = assessor.GetRiskFactor("market_volatility")
		require.NoError(t, err)
		assert.Equal(t, "Market Volatility", factor.Name)
		assert.Equal(t, 0.15, factor.Weight)
		assert.Equal(t, 75, factor.Threshold)
	})
}

func TestRiskAssessorAssessRisk(t *testing.T) {
	t.Run("assess risk for new user", func(t *testing.T) {
		lendingPool := createTestPool()
		assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

		// Supply some assets to create an account
		err := lendingPool.Supply("user1", big.NewInt(10000000)) // 10 USDC
		require.NoError(t, err)

		// Get account and set collateral value
		account, err := lendingPool.GetAccount("user1")
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(15000000) // 15 USDC collateral

		// Assess risk
		assessment, err := assessor.AssessRisk("user1")
		require.NoError(t, err)
		assert.NotNil(t, assessment)

		// Verify assessment data
		assert.Equal(t, "user1", assessment.UserID)
		assert.Equal(t, big.NewInt(15000000), assessment.CollateralValue)
		assert.Equal(t, big.NewInt(0), assessment.BorrowValue)
		assert.Equal(t, 1, assessment.AssessmentCount)
		assert.NotZero(t, assessment.LastAssessment)
		assert.NotZero(t, assessment.NextAssessment)

		// Verify risk score
		require.NotNil(t, assessment.RiskScore)
		assert.NotZero(t, assessment.RiskScore.LastUpdated)
		assert.NotEmpty(t, assessment.RiskScore.Recommendations)
	})

	t.Run("assess risk for user with borrows", func(t *testing.T) {
		lendingPool := createTestPool()
		assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

		// Supply and borrow
		err := lendingPool.Supply("user2", big.NewInt(10000000)) // 10 USDC
		require.NoError(t, err)

		account, err := lendingPool.GetAccount("user2")
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(15000000) // 15 USDC collateral

		// Borrow some assets
		err = lendingPool.Borrow("user2", big.NewInt(5000000)) // 5 USDC
		require.NoError(t, err)

		// Get updated account to ensure health factor is calculated
		account, err = lendingPool.GetAccount("user2")
		require.NoError(t, err)
		require.NotNil(t, account.HealthFactor)

		// Assess risk
		assessment, err := assessor.AssessRisk("user2")
		require.NoError(t, err)
		require.NotNil(t, assessment)

		// Verify risk metrics
		assert.True(t, assessment.LiquidationRisk > 0.0)
		assert.True(t, assessment.MarketExposure > 0.0)
		assert.True(t, assessment.MarketExposure < 100.0) // Should be percentage
	})

	t.Run("assess risk for non-existent user", func(t *testing.T) {
		lendingPool := createTestPool()
		assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

		assessment, err := assessor.AssessRisk("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, assessment)
		assert.Contains(t, err.Error(), "failed to get user account")
	})
}

func TestRiskAssessorScoreCalculation(t *testing.T) {
	lendingPool := createTestPool()
	assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

	t.Run("calculate collateral score", func(t *testing.T) {
		// Test different collateral ratios
		testCases := []struct {
			collateral *big.Int
			borrow     *big.Int
			expected   int
		}{
			{big.NewInt(20000000), big.NewInt(10000000), 0},  // 200% collateral
			{big.NewInt(15000000), big.NewInt(10000000), 20}, // 150% collateral
			{big.NewInt(12500000), big.NewInt(10000000), 40}, // 125% collateral
			{big.NewInt(11000000), big.NewInt(10000000), 60}, // 110% collateral
			{big.NewInt(10500000), big.NewInt(10000000), 80}, // 105% collateral
			{big.NewInt(9500000), big.NewInt(10000000), 100}, // 95% collateral
		}

		for _, tc := range testCases {
			account := &Account{
				CollateralValue: tc.collateral,
				BorrowValue:     tc.borrow,
			}
			score := assessor.calculateCollateralScore(account)
			assert.Equal(t, tc.expected, score, "collateral: %s, borrow: %s", tc.collateral, tc.borrow)
		}
	})

	t.Run("calculate credit score", func(t *testing.T) {
		// Test different health factors
		testCases := []struct {
			healthFactor *big.Int
			expected     int
		}{
			{big.NewInt(200), 80}, // 200% healthy
			{big.NewInt(150), 80}, // 150% healthy
			{big.NewInt(120), 70}, // 120% healthy
			{big.NewInt(100), 60}, // 100% healthy
			{big.NewInt(80), 40},  // 80% healthy
			{big.NewInt(60), 30},  // 60% healthy
			{big.NewInt(50), 20},  // 50% healthy
		}

		for _, tc := range testCases {
			account := &Account{
				HealthFactor: tc.healthFactor,
			}
			assessment := &RiskAssessment{
				AssessmentCount: 1,
				HistoricalData:  []HistoricalRiskData{},
			}
			score := assessor.calculateCreditScore(account, assessment)
			assert.Equal(t, tc.expected, score, "health factor: %s", tc.healthFactor)
		}
	})

	t.Run("calculate liquidity score", func(t *testing.T) {
		// Test different liquidity scenarios
		testCases := []struct {
			supplyBalance   *big.Int
			collateralValue *big.Int
			borrowValue     *big.Int
			expected        int
		}{
			{big.NewInt(0), big.NewInt(10000000), big.NewInt(5000000), 65},       // No supply, some excess
			{big.NewInt(1000000), big.NewInt(10000000), big.NewInt(5000000), 85}, // With supply, some excess
			{big.NewInt(2000000), big.NewInt(10000000), big.NewInt(5000000), 85}, // With supply, more excess
		}

		for _, tc := range testCases {
			account := &Account{
				SupplyBalance:   tc.supplyBalance,
				CollateralValue: tc.collateralValue,
				BorrowValue:     tc.borrowValue,
			}
			score := assessor.calculateLiquidityScore(account)
			assert.Equal(t, tc.expected, score)
		}
	})
}

func TestRiskAssessorRiskLevels(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("determine risk levels", func(t *testing.T) {
		testCases := []struct {
			score     int
			riskLevel RiskLevel
		}{
			{0, RiskLevelLow},
			{25, RiskLevelLow},
			{26, RiskLevelMedium},
			{50, RiskLevelMedium},
			{51, RiskLevelHigh},
			{75, RiskLevelHigh},
			{76, RiskLevelCritical},
			{100, RiskLevelCritical},
		}

		for _, tc := range testCases {
			riskLevel := assessor.determineRiskLevel(tc.score)
			assert.Equal(t, tc.riskLevel, riskLevel, "score: %d", tc.score)
		}
	})
}

func TestRiskAssessorRecommendations(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("generate recommendations for high risk", func(t *testing.T) {
		riskScore := &RiskScore{
			OverallScore:    85,
			CollateralScore: 80,
			CreditScore:     30,
			LiquidityScore:  60,
			MarketScore:     75,
			BehaviorScore:   70,
		}

		recommendations := assessor.generateRecommendations(riskScore)
		assert.NotEmpty(t, recommendations)
		assert.Contains(t, recommendations, "Consider reducing borrow amount to improve risk profile")
		assert.Contains(t, recommendations, "Add more collateral to reduce liquidation risk")
		assert.Contains(t, recommendations, "Monitor account health factor closely")
	})

	t.Run("generate recommendations for medium risk", func(t *testing.T) {
		riskScore := &RiskScore{
			OverallScore:    60,
			CollateralScore: 50,
			CreditScore:     60,
			LiquidityScore:  70,
			MarketScore:     60,
			BehaviorScore:   65,
		}

		recommendations := assessor.generateRecommendations(riskScore)
		assert.NotEmpty(t, recommendations)
		assert.Contains(t, recommendations, "Monitor collateral ratio and maintain healthy buffer")
		assert.Contains(t, recommendations, "Consider partial repayment to improve health factor")
	})

	t.Run("generate recommendations for low risk", func(t *testing.T) {
		riskScore := &RiskScore{
			OverallScore:    20,
			CollateralScore: 20,
			CreditScore:     80,
			LiquidityScore:  90,
			MarketScore:     30,
			BehaviorScore:   85,
		}

		recommendations := assessor.generateRecommendations(riskScore)
		// Low risk users should have fewer recommendations
		assert.Len(t, recommendations, 0)
	})
}

func TestRiskAssessorLiquidationRisk(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("calculate liquidation risk", func(t *testing.T) {
		testCases := []struct {
			healthFactor *big.Int
			expectedRisk float64
		}{
			{big.NewInt(200), 0.01}, // 200% healthy = 1% risk
			{big.NewInt(150), 0.05}, // 150% healthy = 5% risk
			{big.NewInt(120), 0.10}, // 120% healthy = 10% risk
			{big.NewInt(110), 0.20}, // 110% healthy = 20% risk
			{big.NewInt(105), 0.40}, // 105% healthy = 40% risk
			{big.NewInt(100), 0.60}, // 100% healthy = 60% risk
			{big.NewInt(95), 0.80},  // 95% healthy = 80% risk
			{big.NewInt(80), 0.95},  // 80% healthy = 95% risk
		}

		for _, tc := range testCases {
			account := &Account{
				HealthFactor: tc.healthFactor,
			}
			risk := assessor.calculateLiquidationRisk(account)
			assert.Equal(t, tc.expectedRisk, risk, "health factor: %s", tc.healthFactor)
		}
	})

	t.Run("calculate liquidation risk with nil health factor", func(t *testing.T) {
		account := &Account{
			HealthFactor: nil,
		}
		risk := assessor.calculateLiquidationRisk(account)
		assert.Equal(t, 0.0, risk)
	})
}

func TestRiskAssessorMarketExposure(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("calculate market exposure", func(t *testing.T) {
		testCases := []struct {
			collateralValue  *big.Int
			borrowValue      *big.Int
			expectedExposure float64
		}{
			{big.NewInt(10000000), big.NewInt(5000000), 50.0},   // 50% exposure
			{big.NewInt(10000000), big.NewInt(8000000), 80.0},   // 80% exposure
			{big.NewInt(10000000), big.NewInt(10000000), 100.0}, // 100% exposure
			{big.NewInt(10000000), big.NewInt(0), 0.0},          // 0% exposure
		}

		for _, tc := range testCases {
			account := &Account{
				CollateralValue: tc.collateralValue,
				BorrowValue:     tc.borrowValue,
			}
			exposure := assessor.calculateMarketExposure(account)
			assert.Equal(t, tc.expectedExposure, exposure)
		}
	})

	t.Run("calculate market exposure with zero collateral", func(t *testing.T) {
		account := &Account{
			CollateralValue: big.NewInt(0),
			BorrowValue:     big.NewInt(1000000),
		}
		exposure := assessor.calculateMarketExposure(account)
		assert.Equal(t, 0.0, exposure)
	})
}

func TestRiskAssessorRiskFactors(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("update risk factors", func(t *testing.T) {
		// Create a test assessment
		assessment := &RiskAssessment{
			RiskScore: &RiskScore{
				CollateralScore: 80,
				CreditScore:     60,
				LiquidityScore:  70,
				MarketScore:     65,
				BehaviorScore:   75,
			},
			RiskFactors: make(map[string]*RiskFactor),
		}

		// Update risk factors
		assessor.updateRiskFactors(assessment)

		// Verify factors were updated
		collateralFactor, err := assessor.GetRiskFactor("collateral_ratio")
		require.NoError(t, err)
		assert.Equal(t, 80, collateralFactor.Score)
		assert.True(t, collateralFactor.IsTriggered) // 80 >= 80 threshold

		liquidationFactor, err := assessor.GetRiskFactor("liquidation_proximity")
		require.NoError(t, err)
		assert.Equal(t, 40, liquidationFactor.Score)   // 100 - 60 = 40
		assert.False(t, liquidationFactor.IsTriggered) // 40 < 70 threshold
	})
}

func TestRiskAssessorManagement(t *testing.T) {
	lendingPool := createTestPool()
	assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

	t.Run("get risk assessment", func(t *testing.T) {
		// Create an assessment first
		err := lendingPool.Supply("user3", big.NewInt(10000000))
		require.NoError(t, err)

		assessment, err := assessor.AssessRisk("user3")
		require.NoError(t, err)

		// Retrieve the assessment
		retrievedAssessment, err := assessor.GetRiskAssessment("user3")
		require.NoError(t, err)
		assert.Equal(t, assessment.UserID, retrievedAssessment.UserID)
		assert.Equal(t, assessment.RiskScore.OverallScore, retrievedAssessment.RiskScore.OverallScore)
	})

	t.Run("get non-existent assessment", func(t *testing.T) {
		assessment, err := assessor.GetRiskAssessment("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, assessment)
		assert.Contains(t, err.Error(), "risk assessment not found")
	})

	t.Run("get high risk users", func(t *testing.T) {
		// Create a high-risk user
		err := lendingPool.Supply("user4", big.NewInt(10000000))
		require.NoError(t, err)

		account, err := lendingPool.GetAccount("user4")
		require.NoError(t, err)
		account.CollateralValue = big.NewInt(10000000) // 10 USDC collateral

		// Borrow very close to limit to create high risk
		err = lendingPool.Borrow("user4", big.NewInt(7400000)) // 7.4 USDC (74% of collateral, very close to 75% limit)
		require.NoError(t, err)

		// Create a new assessor with lower alert threshold to catch this user
		lowThresholdAssessor := NewRiskAssessor(lendingPool, 1*time.Hour, 60) // Lower threshold to 60

		// Assess risk with the new assessor
		_, err = lowThresholdAssessor.AssessRisk("user4")
		require.NoError(t, err)

		// Get high risk users with the lower threshold
		highRiskUsers := lowThresholdAssessor.GetHighRiskUsers()
		assert.Contains(t, highRiskUsers, "user4")
	})
}

func TestRiskAssessorFactorManagement(t *testing.T) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)

	t.Run("get risk factor", func(t *testing.T) {
		factor, err := assessor.GetRiskFactor("collateral_ratio")
		require.NoError(t, err)
		assert.Equal(t, "collateral_ratio", factor.ID)
		assert.Equal(t, "Collateral Ratio", factor.Name)
	})

	t.Run("get non-existent risk factor", func(t *testing.T) {
		factor, err := assessor.GetRiskFactor("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, factor)
		assert.Contains(t, err.Error(), "risk factor not found")
	})

	t.Run("update risk factor", func(t *testing.T) {
		// Update factor
		err := assessor.UpdateRiskFactor("collateral_ratio", 0.30, 85)
		assert.NoError(t, err)

		// Verify update
		factor, err := assessor.GetRiskFactor("collateral_ratio")
		require.NoError(t, err)
		assert.Equal(t, 0.30, factor.Weight)
		assert.Equal(t, 85, factor.Threshold)
	})

	t.Run("update risk factor with invalid weight", func(t *testing.T) {
		err := assessor.UpdateRiskFactor("collateral_ratio", 1.5, 85)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid weight value")
	})

	t.Run("update risk factor with invalid threshold", func(t *testing.T) {
		err := assessor.UpdateRiskFactor("collateral_ratio", 0.30, 150)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid threshold value")
	})
}

func TestRiskAssessorStatistics(t *testing.T) {
	lendingPool := createTestPool()
	assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

	t.Run("get risk stats", func(t *testing.T) {
		// Create some assessments
		err := lendingPool.Supply("user5", big.NewInt(10000000))
		require.NoError(t, err)

		_, err = assessor.AssessRisk("user5")
		require.NoError(t, err)

		// Get statistics
		stats := assessor.GetRiskStats()

		assert.Equal(t, 1, stats["total_assessments"])
		assert.Equal(t, 6, stats["total_factors"])
		assert.Equal(t, "1h0m0s", stats["update_interval"])
		assert.Equal(t, 75, stats["alert_threshold"])

		// Check risk level distribution
		distribution := stats["risk_level_distribution"].(map[string]int)
		assert.NotNil(t, distribution)
		assert.Contains(t, distribution, "low")
		assert.Contains(t, distribution, "medium")
		assert.Contains(t, distribution, "high")
		assert.Contains(t, distribution, "critical")
	})
}

func TestRiskAssessorConcurrency(t *testing.T) {
	lendingPool := createTestPool()
	assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

	t.Run("concurrent risk assessments", func(t *testing.T) {
		const numGoroutines = 5
		const numAssessments = 10

		var wg sync.WaitGroup

		// Start goroutines to perform risk assessments concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				userID := "user" + string(rune(id))

				// Supply assets
				err := lendingPool.Supply(userID, big.NewInt(10000000))
				if err != nil {
					return
				}

				for j := 0; j < numAssessments; j++ {
					_, err := assessor.AssessRisk(userID)
					assert.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify no data races occurred
		stats := assessor.GetRiskStats()
		totalAssessments := stats["total_assessments"].(int)
		assert.Equal(t, 5, totalAssessments) // 5 users
	})
}

// Benchmark tests for performance validation
func BenchmarkRiskAssessment(b *testing.B) {
	lendingPool := createTestPool()
	assessor := NewRiskAssessor(lendingPool, 1*time.Hour, 75)

	// Setup: supply assets
	lendingPool.Supply("benchmark_user", big.NewInt(10000000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assessor.AssessRisk("benchmark_user")
	}
}

func BenchmarkRiskScoreCalculation(b *testing.B) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)
	account := &Account{
		UserID:           "benchmark_user",
		SupplyBalance:    big.NewInt(10000000),
		BorrowBalance:    big.NewInt(5000000),
		SupplyIndex:      big.NewInt(1000000000000000000),
		BorrowIndex:      big.NewInt(1000000000000000000),
		CollateralValue:  big.NewInt(15000000),
		BorrowValue:      big.NewInt(5000000),
		HealthFactor:     big.NewInt(120),
		LastUpdate:       time.Now(),
		IsLiquidatable:   false,
	}
	assessment := &RiskAssessment{
		AssessmentCount: 5,
		HistoricalData:  []HistoricalRiskData{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assessor.calculateRiskScore(account, assessment)
	}
}

func BenchmarkCollateralScoreCalculation(b *testing.B) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)
	account := &Account{
		UserID:           "benchmark_user",
		SupplyBalance:    big.NewInt(10000000),
		BorrowBalance:    big.NewInt(5000000),
		SupplyIndex:      big.NewInt(1000000000000000000),
		BorrowIndex:      big.NewInt(1000000000000000000),
		CollateralValue:  big.NewInt(15000000),
		BorrowValue:      big.NewInt(5000000),
		HealthFactor:     big.NewInt(120),
		LastUpdate:       time.Now(),
		IsLiquidatable:   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assessor.calculateCollateralScore(account)
	}
}

func BenchmarkCreditScoreCalculation(b *testing.B) {
	assessor := NewRiskAssessor(createTestPool(), 1*time.Hour, 75)
	account := &Account{
		UserID:           "benchmark_user",
		SupplyBalance:    big.NewInt(10000000),
		BorrowBalance:    big.NewInt(5000000),
		SupplyIndex:      big.NewInt(1000000000000000000),
		BorrowIndex:      big.NewInt(1000000000000000000),
		CollateralValue:  big.NewInt(15000000),
		BorrowValue:      big.NewInt(5000000),
		HealthFactor:     big.NewInt(120),
		LastUpdate:       time.Now(),
		IsLiquidatable:   false,
	}
	assessment := &RiskAssessment{
		AssessmentCount: 5,
		HistoricalData:  []HistoricalRiskData{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assessor.calculateCreditScore(account, assessment)
	}
}
