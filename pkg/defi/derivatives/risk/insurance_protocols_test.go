package risk

import (
	"math/big"
	"testing"
	"time"
)

func TestNewInsuranceManager(t *testing.T) {
	tests := []struct {
		name            string
		riskFreeRate    *big.Float
		basePremiumRate *big.Float
		expectError     bool
	}{
		{
			name:            "Valid Parameters",
			riskFreeRate:    big.NewFloat(0.05),
			basePremiumRate: big.NewFloat(0.02),
			expectError:     false,
		},
		{
			name:            "Nil Risk-Free Rate",
			riskFreeRate:    nil,
			basePremiumRate: big.NewFloat(0.02),
			expectError:     true,
		},
		{
			name:            "Nil Base Premium Rate",
			riskFreeRate:    big.NewFloat(0.05),
			basePremiumRate: nil,
			expectError:     true,
		},
		{
			name:            "Negative Base Premium Rate",
			riskFreeRate:    big.NewFloat(0.05),
			basePremiumRate: big.NewFloat(-0.01),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewInsuranceManager(tt.riskFreeRate, tt.basePremiumRate)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if manager == nil {
				t.Errorf("Expected manager but got nil")
				return
			}

			if len(manager.pools) != 0 {
				t.Errorf("Expected empty pools map")
			}

			if len(manager.policies) != 0 {
				t.Errorf("Expected empty policies map")
			}

			if len(manager.claims) != 0 {
				t.Errorf("Expected empty claims map")
			}
		})
	}
}

func TestNewInsurancePool(t *testing.T) {
	expiresAt := time.Now().AddDate(0, 1, 0) // 1 month from now

	tests := []struct {
		name          string
		id            string
		poolName      string
		insuranceType InsuranceType
		description   string
		totalCapacity *big.Float
		premiumRate   *big.Float
		coverageLimit *big.Float
		minCoverage   *big.Float
		maxCoverage   *big.Float
		expiresAt     time.Time
		expectError   bool
	}{
		{
			name:          "Valid Pool",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000), // 1M capacity
			premiumRate:   big.NewFloat(0.02),    // 2% premium rate
			coverageLimit: big.NewFloat(500000),  // 500K coverage limit
			minCoverage:   big.NewFloat(10000),   // 10K min coverage
			maxCoverage:   big.NewFloat(100000),  // 100K max coverage
			expiresAt:     expiresAt,
			expectError:   false,
		},
		{
			name:          "Empty ID",
			id:            "",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Empty Name",
			id:            "pool1",
			poolName:      "",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Zero Total Capacity",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(0),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Negative Premium Rate",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(-0.01),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Min Coverage >= Max Coverage",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(100000), // Same as max
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Max Coverage > Total Capacity",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(2000000), // 2M > 1M capacity
			expiresAt:     expiresAt,
			expectError:   true,
		},
		{
			name:          "Expired Date",
			id:            "pool1",
			poolName:      "DeFi Portfolio Insurance",
			insuranceType: PortfolioInsurance,
			description:   "Insurance for DeFi portfolio losses",
			totalCapacity: big.NewFloat(1000000),
			premiumRate:   big.NewFloat(0.02),
			coverageLimit: big.NewFloat(500000),
			minCoverage:   big.NewFloat(10000),
			maxCoverage:   big.NewFloat(100000),
			expiresAt:     time.Now().AddDate(0, 0, -1), // Yesterday
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewInsurancePool(
				tt.id,
				tt.poolName,
				tt.insuranceType,
				tt.description,
				tt.totalCapacity,
				tt.premiumRate,
				tt.coverageLimit,
				tt.minCoverage,
				tt.maxCoverage,
				tt.expiresAt,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pool == nil {
				t.Errorf("Expected pool but got nil")
				return
			}

			if pool.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, pool.ID)
			}

			if pool.Name != tt.poolName {
				t.Errorf("Expected name %s, got %s", tt.poolName, pool.Name)
			}

			if pool.Type != tt.insuranceType {
				t.Errorf("Expected type %v, got %v", tt.insuranceType, pool.Type)
			}

			if pool.TotalCapacity.Cmp(tt.totalCapacity) != 0 {
				t.Errorf("Expected total capacity %v, got %v", tt.totalCapacity, pool.TotalCapacity)
			}

			if pool.UsedCapacity.Sign() != 0 {
				t.Errorf("Expected zero used capacity, got %v", pool.UsedCapacity)
			}

			if pool.AvailableCapacity.Cmp(tt.totalCapacity) != 0 {
				t.Errorf("Expected available capacity %v, got %v", tt.totalCapacity, pool.AvailableCapacity)
			}

			if pool.Status != Active {
				t.Errorf("Expected active status, got %v", pool.Status)
			}
		})
	}
}

func TestNewInsurancePolicy(t *testing.T) {
	startDate := time.Now().AddDate(0, 0, 1)  // Tomorrow
	endDate := time.Now().AddDate(0, 1, 0)    // 1 month from now

	tests := []struct {
		name          string
		id            string
		poolID        string
		userID        string
		coverageAmount *big.Float
		coverageType   InsuranceType
		startDate      time.Time
		endDate        time.Time
		expectError    bool
	}{
		{
			name:           "Valid Policy",
			id:             "policy1",
			poolID:         "pool1",
			userID:         "user1",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      startDate,
			endDate:        endDate,
			expectError:    false,
		},
		{
			name:           "Empty ID",
			id:             "",
			poolID:         "pool1",
			userID:         "user1",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      startDate,
			endDate:        endDate,
			expectError:    true,
		},
		{
			name:           "Empty Pool ID",
			id:             "policy1",
			poolID:         "",
			userID:         "user1",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      startDate,
			endDate:        endDate,
			expectError:    true,
		},
		{
			name:           "Empty User ID",
			id:             "policy1",
			poolID:         "pool1",
			userID:         "",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      startDate,
			endDate:        endDate,
			expectError:    true,
		},
		{
			name:           "Zero Coverage Amount",
			id:             "policy1",
			poolID:         "pool1",
			userID:         "user1",
			coverageAmount: big.NewFloat(0),
			coverageType:   PortfolioInsurance,
			startDate:      startDate,
			endDate:        endDate,
			expectError:    true,
		},
		{
			name:           "Start Date After End Date",
			id:             "policy1",
			poolID:         "pool1",
			userID:         "user1",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      endDate,
			endDate:        startDate,
			expectError:    true,
		},
		{
			name:           "Start Date in Past",
			id:             "policy1",
			poolID:         "pool1",
			userID:         "user1",
			coverageAmount: big.NewFloat(50000),
			coverageType:   PortfolioInsurance,
			startDate:      time.Now().AddDate(0, 0, -1), // Yesterday
			endDate:        endDate,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := NewInsurancePolicy(
				tt.id,
				tt.poolID,
				tt.userID,
				tt.coverageAmount,
				tt.coverageType,
				tt.startDate,
				tt.endDate,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if policy == nil {
				t.Errorf("Expected policy but got nil")
				return
			}

			if policy.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, policy.ID)
			}

			if policy.PoolID != tt.poolID {
				t.Errorf("Expected pool ID %s, got %s", tt.poolID, policy.PoolID)
			}

			if policy.UserID != tt.userID {
				t.Errorf("Expected user ID %s, got %s", tt.userID, policy.UserID)
			}

			if policy.CoverageAmount.Cmp(tt.coverageAmount) != 0 {
				t.Errorf("Expected coverage amount %v, got %v", tt.coverageAmount, policy.CoverageAmount)
			}

			if policy.CoverageType != tt.coverageType {
				t.Errorf("Expected coverage type %v, got %v", tt.coverageType, policy.CoverageType)
			}

			if policy.Status != Active {
				t.Errorf("Expected active status, got %v", policy.Status)
			}

			if policy.PremiumAmount.Sign() != 0 {
				t.Errorf("Expected zero premium amount, got %v", policy.PremiumAmount)
			}
		})
	}
}

func TestInsuranceManagerPremiumCalculation(t *testing.T) {
	// Create insurance manager
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Create insurance pool
	expiresAt := time.Now().AddDate(0, 1, 0)
	pool, err := NewInsurancePool(
		"pool1",
		"DeFi Portfolio Insurance",
		PortfolioInsurance,
		"Insurance for DeFi portfolio losses",
		big.NewFloat(1000000), // 1M capacity
		big.NewFloat(0.02),    // 2% premium rate
		big.NewFloat(500000),  // 500K coverage limit
		big.NewFloat(10000),   // 10K min coverage
		big.NewFloat(100000),  // 100K max coverage
		expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create insurance pool: %v", err)
	}

	// Create insurance policy
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 1, 0)
	policy, err := NewInsurancePolicy(
		"policy1",
		"pool1",
		"user1",
		big.NewFloat(50000), // 50K coverage
		PortfolioInsurance,
		startDate,
		endDate,
	)
	if err != nil {
		t.Fatalf("Failed to create insurance policy: %v", err)
	}

	// Create risk assessment
	assessment, err := NewRiskAssessment(
		"assessment1",
		"user1",
		"portfolio1",
		big.NewFloat(0.15),  // Risk score
		big.NewFloat(0.05),  // VaR
		big.NewFloat(0.08),  // CVaR
		big.NewFloat(0.25),  // Volatility
		big.NewFloat(2.0),   // Leverage
		big.NewFloat(0.8),   // Liquidity score
		big.NewFloat(0.1),   // Correlation risk
		big.NewFloat(0.12),  // Market risk
		big.NewFloat(0.05),  // Credit risk
		big.NewFloat(0.03),  // Operational risk
	)
	if err != nil {
		t.Fatalf("Failed to create risk assessment: %v", err)
	}

	t.Run("Calculate Premium", func(t *testing.T) {
		calculation, err := manager.CalculatePremium(policy, pool, assessment)
		if err != nil {
			t.Errorf("Failed to calculate premium: %v", err)
			return
		}

		if calculation == nil {
			t.Errorf("Expected calculation but got nil")
			return
		}

		if calculation.BasePremium == nil {
			t.Errorf("Expected base premium but got nil")
		}

		if calculation.RiskMultiplier == nil {
			t.Errorf("Expected risk multiplier but got nil")
		}

		if calculation.CoverageMultiplier == nil {
			t.Errorf("Expected coverage multiplier but got nil")
		}

		if calculation.DurationMultiplier == nil {
			t.Errorf("Expected duration multiplier but got nil")
		}

		if calculation.FinalPremium == nil {
			t.Errorf("Expected final premium but got nil")
		}

		if len(calculation.RiskFactors) == 0 {
			t.Errorf("Expected risk factors but got none")
		}

		// Verify that final premium is calculated correctly
		expectedBasePremium := new(big.Float).Mul(pool.PremiumRate, policy.CoverageAmount)
		if calculation.BasePremium.Cmp(expectedBasePremium) != 0 {
			t.Errorf("Expected base premium %v, got %v", expectedBasePremium, calculation.BasePremium)
		}

		// Verify that policy was updated with calculated values
		if policy.PremiumAmount.Cmp(calculation.FinalPremium) != 0 {
			t.Errorf("Policy premium amount not updated correctly")
		}

		if policy.RiskAssessment.Cmp(assessment.OverallRisk) != 0 {
			t.Errorf("Policy risk assessment not updated correctly")
		}
	})

	t.Run("Nil Policy", func(t *testing.T) {
		_, err := manager.CalculatePremium(nil, pool, assessment)
		if err == nil {
			t.Errorf("Expected error for nil policy")
		}
	})

	t.Run("Nil Pool", func(t *testing.T) {
		_, err := manager.CalculatePremium(policy, nil, assessment)
		if err == nil {
			t.Errorf("Expected error for nil pool")
		}
	})

	t.Run("Nil Assessment", func(t *testing.T) {
		_, err := manager.CalculatePremium(policy, pool, nil)
		if err == nil {
			t.Errorf("Expected error for nil assessment")
		}
	})
}

func TestInsuranceManagerClaimsProcessing(t *testing.T) {
	// Create insurance manager
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Create insurance claim
	claim, err := NewInsuranceClaim(
		"claim1",
		"policy1",
		"user1",
		big.NewFloat(25000), // 25K claim
		"Portfolio loss due to market crash",
		[]string{"transaction_logs.pdf", "market_data.csv"},
	)
	if err != nil {
		t.Fatalf("Failed to create insurance claim: %v", err)
	}

	// Add claim to manager
	err = manager.AddInsuranceClaim(claim)
	if err != nil {
		t.Fatalf("Failed to add claim: %v", err)
	}

	t.Run("Process Claim - Approved", func(t *testing.T) {
		err := manager.ProcessClaim("claim1", true, nil)
		if err != nil {
			t.Errorf("Failed to process claim: %v", err)
			return
		}

		// Verify claim status
		processedClaim, err := manager.GetInsuranceClaim("claim1")
		if err != nil {
			t.Errorf("Failed to get processed claim: %v", err)
			return
		}

		if processedClaim.Status != Approved {
			t.Errorf("Expected approved status, got %v", processedClaim.Status)
		}

		if processedClaim.ProcessedAt == nil {
			t.Errorf("Expected processed timestamp")
		}

		if processedClaim.ApprovedAt == nil {
			t.Errorf("Expected approved timestamp")
		}
	})

	t.Run("Process Claim - Rejected", func(t *testing.T) {
		// Create a new claim for rejection test
		rejectedClaim, err := NewInsuranceClaim(
			"claim2",
			"policy1",
			"user1",
			big.NewFloat(25000),
			"Portfolio loss due to market crash",
			[]string{"transaction_logs.pdf"},
		)
		if err != nil {
			t.Fatalf("Failed to create rejected claim: %v", err)
		}

		err = manager.AddInsuranceClaim(rejectedClaim)
		if err != nil {
			t.Fatalf("Failed to add rejected claim: %v", err)
		}

		rejectionReason := "Insufficient evidence provided"
		err = manager.ProcessClaim("claim2", false, &rejectionReason)
		if err != nil {
			t.Errorf("Failed to process rejected claim: %v", err)
			return
		}

		// Verify claim status
		processedClaim, err := manager.GetInsuranceClaim("claim2")
		if err != nil {
			t.Errorf("Failed to get processed claim: %v", err)
			return
		}

		if processedClaim.Status != Rejected {
			t.Errorf("Expected rejected status, got %v", processedClaim.Status)
		}

		if processedClaim.RejectionReason == nil {
			t.Errorf("Expected rejection reason")
		}

		if *processedClaim.RejectionReason != rejectionReason {
			t.Errorf("Expected rejection reason %s, got %s", rejectionReason, *processedClaim.RejectionReason)
		}
	})

	t.Run("Pay Claim", func(t *testing.T) {
		// First approve the claim
		err := manager.ProcessClaim("claim1", true, nil)
		if err != nil {
			t.Errorf("Failed to approve claim: %v", err)
			return
		}

		// Then pay it
		err = manager.PayClaim("claim1")
		if err != nil {
			t.Errorf("Failed to pay claim: %v", err)
			return
		}

		// Verify claim status
		paidClaim, err := manager.GetInsuranceClaim("claim1")
		if err != nil {
			t.Errorf("Failed to get paid claim: %v", err)
			return
		}

		if paidClaim.Status != Paid {
			t.Errorf("Expected paid status, got %v", paidClaim.Status)
		}

		if paidClaim.PaidAt == nil {
			t.Errorf("Expected paid timestamp")
		}
	})

	t.Run("Pay Unapproved Claim", func(t *testing.T) {
		// Create a new pending claim
		pendingClaim, err := NewInsuranceClaim(
			"claim3",
			"policy1",
			"user1",
			big.NewFloat(25000),
			"Portfolio loss due to market crash",
			[]string{"transaction_logs.pdf"},
		)
		if err != nil {
			t.Fatalf("Failed to create pending claim: %v", err)
		}

		err = manager.AddInsuranceClaim(pendingClaim)
		if err != nil {
			t.Fatalf("Failed to add pending claim: %v", err)
		}

		// Try to pay without approval
		err = manager.PayClaim("claim3")
		if err == nil {
			t.Errorf("Expected error when paying unapproved claim")
		}
	})
}
