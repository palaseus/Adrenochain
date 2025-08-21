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
	startDate := time.Now().AddDate(0, 0, 1) // Tomorrow
	endDate := time.Now().AddDate(0, 1, 0)   // 1 month from now

	tests := []struct {
		name           string
		id             string
		poolID         string
		userID         string
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
		big.NewFloat(0.15), // Risk score
		big.NewFloat(0.05), // VaR
		big.NewFloat(0.08), // CVaR
		big.NewFloat(0.25), // Volatility
		big.NewFloat(2.0),  // Leverage
		big.NewFloat(0.8),  // Liquidity score
		big.NewFloat(0.1),  // Correlation risk
		big.NewFloat(0.12), // Market risk
		big.NewFloat(0.05), // Credit risk
		big.NewFloat(0.03), // Operational risk
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

// TestInsuranceManagerPoolManagement tests insurance pool management functions
func TestInsuranceManagerPoolManagement(t *testing.T) {
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Test AddInsurancePool
	t.Run("AddInsurancePool", func(t *testing.T) {
		pool, err := NewInsurancePool(
			"pool1",
			"Test Pool",
			PortfolioInsurance,
			"Test pool description",
			big.NewFloat(1000000),
			big.NewFloat(0.05),
			big.NewFloat(100000),
			big.NewFloat(1000),
			big.NewFloat(100000),
			time.Now().Add(365*24*time.Hour),
		)
		if err != nil {
			t.Fatalf("Failed to create insurance pool: %v", err)
		}

		// Test adding valid pool
		err = manager.AddInsurancePool(pool)
		if err != nil {
			t.Errorf("Failed to add insurance pool: %v", err)
		}

		// Test adding nil pool
		err = manager.AddInsurancePool(nil)
		if err == nil {
			t.Errorf("Expected error when adding nil pool")
		}
	})

	// Test GetInsurancePool
	t.Run("GetInsurancePool", func(t *testing.T) {
		// Test getting existing pool
		pool, err := manager.GetInsurancePool("pool1")
		if err != nil {
			t.Errorf("Failed to get insurance pool: %v", err)
		}
		if pool == nil {
			t.Error("Expected pool but got nil")
		}
		if pool != nil && pool.ID != "pool1" {
			t.Errorf("Expected pool ID 'pool1', got '%s'", pool.ID)
		}

		// Test getting non-existent pool
		_, err = manager.GetInsurancePool("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent pool")
		}

		// Test getting pool with empty ID
		_, err = manager.GetInsurancePool("")
		if err == nil {
			t.Error("Expected error when getting pool with empty ID")
		}
	})
}

// TestInsuranceManagerPolicyManagement tests insurance policy management functions
func TestInsuranceManagerPolicyManagement(t *testing.T) {
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Test AddInsurancePolicy
	t.Run("AddInsurancePolicy", func(t *testing.T) {
		policy, err := NewInsurancePolicy(
			"policy1",
			"pool1",
			"user1",
			big.NewFloat(50000),
			PortfolioInsurance,
			time.Now().Add(time.Hour), // Future start date
			time.Now().Add(365*24*time.Hour),
		)
		if err != nil {
			t.Fatalf("Failed to create insurance policy: %v", err)
		}

		// Test adding valid policy
		err = manager.AddInsurancePolicy(policy)
		if err != nil {
			t.Errorf("Failed to add insurance policy: %v", err)
		}

		// Test adding nil policy
		err = manager.AddInsurancePolicy(nil)
		if err == nil {
			t.Errorf("Expected error when adding nil policy")
		}
	})

	// Test GetInsurancePolicy
	t.Run("GetInsurancePolicy", func(t *testing.T) {
		// Test getting existing policy
		policy, err := manager.GetInsurancePolicy("policy1")
		if err != nil {
			t.Errorf("Failed to get insurance policy: %v", err)
		}
		if policy == nil {
			t.Error("Expected policy but got nil")
		}
		if policy != nil && policy.ID != "policy1" {
			t.Errorf("Expected policy ID 'policy1', got '%s'", policy.ID)
		}

		// Test getting non-existent policy
		_, err = manager.GetInsurancePolicy("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent policy")
		}

		// Test getting policy with empty ID
		_, err = manager.GetInsurancePolicy("")
		if err == nil {
			t.Error("Expected error when getting policy with empty ID")
		}
	})
}

// TestInsuranceManagerRiskAssessmentManagement tests risk assessment management functions
func TestInsuranceManagerRiskAssessmentManagement(t *testing.T) {
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Test AddRiskAssessment
	t.Run("AddRiskAssessment", func(t *testing.T) {
		assessment, err := NewRiskAssessment(
			"assess1",
			"user1",
			"portfolio1",
			big.NewFloat(0.75),
			big.NewFloat(1000),
			big.NewFloat(1200),
			big.NewFloat(0.15),
			big.NewFloat(2.0),
			big.NewFloat(0.8),
			big.NewFloat(0.3),
			big.NewFloat(0.4),
			big.NewFloat(0.2),
			big.NewFloat(0.1),
		)
		if err != nil {
			t.Fatalf("Failed to create risk assessment: %v", err)
		}

		// Test adding valid assessment
		err = manager.AddRiskAssessment(assessment)
		if err != nil {
			t.Errorf("Failed to add risk assessment: %v", err)
		}

		// Test adding nil assessment
		err = manager.AddRiskAssessment(nil)
		if err == nil {
			t.Errorf("Expected error when adding nil assessment")
		}
	})

	// Test GetRiskAssessment
	t.Run("GetRiskAssessment", func(t *testing.T) {
		// Test getting existing assessment
		assessment, err := manager.GetRiskAssessment("assess1")
		if err != nil {
			t.Errorf("Failed to get risk assessment: %v", err)
		}
		if assessment == nil {
			t.Error("Expected assessment but got nil")
		}
		if assessment != nil && assessment.ID != "assess1" {
			t.Errorf("Expected assessment ID 'assess1', got '%s'", assessment.ID)
		}

		// Test getting non-existent assessment
		_, err = manager.GetRiskAssessment("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent assessment")
		}

		// Test getting assessment with empty ID
		_, err = manager.GetRiskAssessment("")
		if err == nil {
			t.Error("Expected error when getting assessment with empty ID")
		}
	})
}

// TestInsuranceManagerUserQueries tests user-specific query functions
func TestInsuranceManagerUserQueries(t *testing.T) {
	manager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Add test data
	pool, _ := NewInsurancePool("pool1", "Test Pool", PortfolioInsurance, "Test pool", big.NewFloat(1000000), big.NewFloat(0.05), big.NewFloat(100000), big.NewFloat(1000), big.NewFloat(100000), time.Now().Add(365*24*time.Hour))
	policy1, _ := NewInsurancePolicy("policy1", "pool1", "user1", big.NewFloat(50000), PortfolioInsurance, time.Now().Add(time.Hour), time.Now().Add(365*24*time.Hour))
	policy2, _ := NewInsurancePolicy("policy2", "pool1", "user1", big.NewFloat(30000), PortfolioInsurance, time.Now().Add(time.Hour), time.Now().Add(365*24*time.Hour))
	policy3, _ := NewInsurancePolicy("policy3", "pool1", "user2", big.NewFloat(40000), PortfolioInsurance, time.Now().Add(time.Hour), time.Now().Add(365*24*time.Hour))

	manager.AddInsurancePool(pool)
	manager.AddInsurancePolicy(policy1)
	manager.AddInsurancePolicy(policy2)
	manager.AddInsurancePolicy(policy3)

	// Test GetPoolUtilization
	t.Run("GetPoolUtilization", func(t *testing.T) {
		// Test with existing pool
		utilization, err := manager.GetPoolUtilization("pool1")
		if err != nil {
			t.Errorf("Failed to get pool utilization: %v", err)
		}
		if utilization == nil {
			t.Error("Expected utilization but got nil")
		}
		// utilization is a map[string]*big.Float, check if it contains the expected data
		if utilization != nil {
			if _, exists := utilization["utilization_rate"]; !exists {
				t.Error("Expected utilization_rate in utilization map")
			}
		}

		// Test with non-existent pool
		_, err = manager.GetPoolUtilization("nonexistent")
		if err == nil {
			t.Error("Expected error when getting utilization for non-existent pool")
		}

		// Test with empty pool ID
		_, err = manager.GetPoolUtilization("")
		if err == nil {
			t.Error("Expected error when getting utilization with empty pool ID")
		}
	})

	// Test GetUserPolicies
	t.Run("GetUserPolicies", func(t *testing.T) {
		// Test with user who has policies
		policies, err := manager.GetUserPolicies("user1")
		if err != nil {
			t.Errorf("Failed to get user policies: %v", err)
		}
		if len(policies) != 2 {
			t.Errorf("Expected 2 policies for user1, got %d", len(policies))
		}

		// Test with user who has no policies
		policies, err = manager.GetUserPolicies("user3")
		if err != nil {
			t.Errorf("Failed to get user policies: %v", err)
		}
		if len(policies) != 0 {
			t.Errorf("Expected 0 policies for user3, got %d", len(policies))
		}

		// Test with empty user ID
		policies, err = manager.GetUserPolicies("")
		if err == nil {
			t.Error("Expected error when getting policies with empty user ID")
		}
	})

	// Test GetUserClaims (after adding some claims)
	t.Run("GetUserClaims", func(t *testing.T) {
		// Add some claims
		claim1, _ := NewInsuranceClaim("claim1", "policy1", "user1", big.NewFloat(25000), "Equipment damage", []string{"photo1.jpg", "report.pdf"})
		claim2, _ := NewInsuranceClaim("claim2", "policy2", "user1", big.NewFloat(15000), "Property loss", []string{"photo2.jpg"})
		claim3, _ := NewInsuranceClaim("claim3", "policy3", "user2", big.NewFloat(20000), "Liability", []string{"witness_statement.pdf"})

		manager.AddInsuranceClaim(claim1)
		manager.AddInsuranceClaim(claim2)
		manager.AddInsuranceClaim(claim3)

		// Test with user who has claims
		claims, err := manager.GetUserClaims("user1")
		if err != nil {
			t.Errorf("Failed to get user claims: %v", err)
		}
		if len(claims) != 2 {
			t.Errorf("Expected 2 claims for user1, got %d", len(claims))
		}

		// Test with user who has no claims
		claims, err = manager.GetUserClaims("user3")
		if err != nil {
			t.Errorf("Failed to get user claims: %v", err)
		}
		if len(claims) != 0 {
			t.Errorf("Expected 0 claims for user3, got %d", len(claims))
		}

		// Test with empty user ID
		claims, err = manager.GetUserClaims("")
		if err == nil {
			t.Error("Expected error when getting claims with empty user ID")
		}
	})
}

// TestNewInsuranceClaim tests the NewInsuranceClaim constructor
func TestNewInsuranceClaim(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		policyID    string
		userID      string
		claimAmount *big.Float
		description string
		evidence    []string
		expectError bool
	}{
		{
			name:        "Valid Claim",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(50000),
			description: "Property damage",
			evidence:    []string{"photo.jpg", "report.pdf"},
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(50000),
			description: "Property damage",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Empty Policy ID",
			id:          "claim1",
			policyID:    "",
			userID:      "user1",
			claimAmount: big.NewFloat(50000),
			description: "Property damage",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Empty User ID",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "",
			claimAmount: big.NewFloat(50000),
			description: "Property damage",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Zero Claim Amount",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(0),
			description: "Property damage",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Negative Claim Amount",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(-1000),
			description: "Property damage",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Empty Description",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(50000),
			description: "",
			evidence:    []string{"photo.jpg"},
			expectError: true,
		},
		{
			name:        "Nil Evidence",
			id:          "claim1",
			policyID:    "policy1",
			userID:      "user1",
			claimAmount: big.NewFloat(50000),
			description: "Property damage",
			evidence:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim, err := NewInsuranceClaim(tt.id, tt.policyID, tt.userID, tt.claimAmount, tt.description, tt.evidence)

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

			if claim.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, claim.ID)
			}
			if claim.PolicyID != tt.policyID {
				t.Errorf("Expected PolicyID %s, got %s", tt.policyID, claim.PolicyID)
			}
			if claim.UserID != tt.userID {
				t.Errorf("Expected UserID %s, got %s", tt.userID, claim.UserID)
			}
			if claim.ClaimAmount.Cmp(tt.claimAmount) != 0 {
				t.Errorf("Expected ClaimAmount %s, got %s", tt.claimAmount.String(), claim.ClaimAmount.String())
			}
		})
	}
}
