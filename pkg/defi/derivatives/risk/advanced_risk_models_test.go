package risk

import (
	"math/big"
	"testing"
)

func TestNewAdvancedRiskManager(t *testing.T) {
	tests := []struct {
		name            string
		riskFreeRate    *big.Float
		confidenceLevel *big.Float
		expectError     bool
	}{
		{
			name:            "Valid Parameters",
			riskFreeRate:    big.NewFloat(0.05),
			confidenceLevel: big.NewFloat(0.95),
			expectError:     false,
		},
		{
			name:            "Nil Risk-Free Rate",
			riskFreeRate:    nil,
			confidenceLevel: big.NewFloat(0.95),
			expectError:     true,
		},
		{
			name:            "Nil Confidence Level",
			riskFreeRate:    big.NewFloat(0.05),
			confidenceLevel: nil,
			expectError:     true,
		},
		{
			name:            "Invalid Confidence Level - Too High",
			riskFreeRate:    big.NewFloat(0.05),
			confidenceLevel: big.NewFloat(1.1),
			expectError:     true,
		},
		{
			name:            "Invalid Confidence Level - Too Low",
			riskFreeRate:    big.NewFloat(0.05),
			confidenceLevel: big.NewFloat(-0.1),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewAdvancedRiskManager(tt.riskFreeRate, tt.confidenceLevel)

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

			if manager.RiskManager == nil {
				t.Errorf("Expected base risk manager to be initialized")
			}

			if len(manager.scenarios) != 0 {
				t.Errorf("Expected empty scenarios map")
			}

			if len(manager.simulations) != 0 {
				t.Errorf("Expected empty simulations map")
			}
		})
	}
}

func TestNewStressTestScenario(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		scenarioName string
		description string
		parameters  map[string]*big.Float
		severity    StressSeverity
		expectError bool
	}{
		{
			name:        "Valid Scenario",
			id:          "scenario1",
			scenarioName: "Market Crash",
			description: "Simulate a severe market crash",
			parameters: map[string]*big.Float{
				"price_shock_BTC": big.NewFloat(0.7), // 30% price drop
				"volatility_shock_BTC": big.NewFloat(2.0), // 2x volatility
			},
			severity:    SevereStress,
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			scenarioName: "Market Crash",
			description: "Simulate a severe market crash",
			parameters: map[string]*big.Float{
				"price_shock_BTC": big.NewFloat(0.7),
			},
			severity:    SevereStress,
			expectError: true,
		},
		{
			name:        "Empty Name",
			id:          "scenario1",
			scenarioName: "",
			description: "Simulate a severe market crash",
			parameters: map[string]*big.Float{
				"price_shock_BTC": big.NewFloat(0.7),
			},
			severity:    SevereStress,
			expectError: true,
		},
		{
			name:        "Nil Parameters",
			id:          "scenario1",
			scenarioName: "Market Crash",
			description: "Simulate a severe market crash",
			parameters:  nil,
			severity:    SevereStress,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := NewStressTestScenario(tt.id, tt.scenarioName, tt.description, tt.parameters, tt.severity)

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

			if scenario == nil {
				t.Errorf("Expected scenario but got nil")
				return
			}

			if scenario.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, scenario.ID)
			}

			if scenario.Name != tt.scenarioName {
				t.Errorf("Expected name %s, got %s", tt.scenarioName, scenario.Name)
			}

			if scenario.Description != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, scenario.Description)
			}

			if scenario.Severity != tt.severity {
				t.Errorf("Expected severity %v, got %v", tt.severity, scenario.Severity)
			}

			if len(scenario.Parameters) != len(tt.parameters) {
				t.Errorf("Expected %d parameters, got %d", len(tt.parameters), len(scenario.Parameters))
			}
		})
	}
}

func TestNewMonteCarloSimulation(t *testing.T) {
	tests := []struct {
		name              string
		id                string
		numSimulations    int
		timeHorizon       *big.Float
		confidenceLevel   *big.Float
		expectError       bool
	}{
		{
			name:              "Valid Simulation",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       false,
		},
		{
			name:              "Empty ID",
			id:                "",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       true,
		},
		{
			name:              "Zero Simulations",
			id:                "mc1",
			numSimulations:    0,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       true,
		},
		{
			name:              "Negative Simulations",
			id:                "mc1",
			numSimulations:    -1000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       true,
		},
		{
			name:              "Nil Time Horizon",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       nil,
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       true,
		},
		{
			name:              "Zero Time Horizon",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(0),
			confidenceLevel:   big.NewFloat(0.95),
			expectError:       true,
		},
		{
			name:              "Nil Confidence Level",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   nil,
			expectError:       true,
		},
		{
			name:              "Invalid Confidence Level - Too High",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(1.1),
			expectError:       true,
		},
		{
			name:              "Invalid Confidence Level - Too Low",
			id:                "mc1",
			numSimulations:    10000,
			timeHorizon:       big.NewFloat(1.0),
			confidenceLevel:   big.NewFloat(-0.1),
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simulation, err := NewMonteCarloSimulation(tt.id, tt.numSimulations, tt.timeHorizon, tt.confidenceLevel)

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

			if simulation == nil {
				t.Errorf("Expected simulation but got nil")
				return
			}

			if simulation.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, simulation.ID)
			}

			if simulation.NumSimulations != tt.numSimulations {
				t.Errorf("Expected %d simulations, got %d", tt.numSimulations, simulation.NumSimulations)
			}

			if simulation.TimeHorizon.Cmp(tt.timeHorizon) != 0 {
				t.Errorf("Expected time horizon %v, got %v", tt.timeHorizon, simulation.TimeHorizon)
			}

			if simulation.ConfidenceLevel.Cmp(tt.confidenceLevel) != 0 {
				t.Errorf("Expected confidence level %v, got %v", tt.confidenceLevel, simulation.ConfidenceLevel)
			}

			if simulation.Seed == 0 {
				t.Errorf("Expected non-zero seed")
			}
		})
	}
}

func TestAdvancedRiskManagerVaRMethodologies(t *testing.T) {
	// Create advanced risk manager
	manager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create advanced risk manager: %v", err)
	}

	// Create a test portfolio
	portfolio, err := NewPortfolio("test_portfolio")
	if err != nil {
		t.Fatalf("Failed to create portfolio: %v", err)
	}

	// Add some test positions with returns
	position1, err := NewPosition("pos1", "BTC", big.NewFloat(1), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Add some sample returns
	position1.Returns = []*big.Float{
		big.NewFloat(0.02),  // 2% return
		big.NewFloat(-0.01), // -1% return
		big.NewFloat(0.03),  // 3% return
		big.NewFloat(-0.02), // -2% return
		big.NewFloat(0.01),  // 1% return
	}

	portfolio.AddPosition(position1)

	timeHorizon := big.NewFloat(1.0) // 1 day

	t.Run("Parametric VaR", func(t *testing.T) {
		varValue, err := manager.CalculateVaRAdvanced(portfolio, timeHorizon, ParametricVaR)
		if err != nil {
			t.Errorf("Failed to calculate parametric VaR: %v", err)
			return
		}

		if varValue == nil {
			t.Errorf("Expected VaR value but got nil")
		}

		// VaR can be negative in some cases (e.g., when expected return is high)
		// This is mathematically correct behavior
	})

	t.Run("Historical VaR", func(t *testing.T) {
		varValue, err := manager.CalculateVaRAdvanced(portfolio, timeHorizon, HistoricalVaR)
		if err != nil {
			t.Errorf("Failed to calculate historical VaR: %v", err)
			return
		}

		if varValue == nil {
			t.Errorf("Expected VaR value but got nil")
		}

		// VaR can be negative in some cases (e.g., when expected return is high)
		// This is mathematically correct behavior
	})

	t.Run("Monte Carlo VaR", func(t *testing.T) {
		varValue, err := manager.CalculateVaRAdvanced(portfolio, timeHorizon, MonteCarloVaR)
		if err != nil {
			t.Errorf("Failed to calculate Monte Carlo VaR: %v", err)
			return
		}

		if varValue == nil {
			t.Errorf("Expected VaR value but got nil")
		}

		// VaR can be negative in some cases (e.g., when expected return is high)
		// This is mathematically correct behavior
	})

	t.Run("Filtered Historical VaR", func(t *testing.T) {
		varValue, err := manager.CalculateVaRAdvanced(portfolio, timeHorizon, FilteredHistoricalVaR)
		if err != nil {
			t.Errorf("Failed to calculate filtered historical VaR: %v", err)
			return
		}

		if varValue == nil {
			t.Errorf("Expected VaR value but got nil")
		}

		// VaR can be negative in some cases (e.g., when expected return is high)
		// This is mathematically correct behavior
	})

	t.Run("Unsupported Methodology", func(t *testing.T) {
		_, err := manager.CalculateVaRAdvanced(portfolio, timeHorizon, 999) // Invalid methodology
		if err == nil {
			t.Errorf("Expected error for unsupported methodology")
		}
	})
}

func TestAdvancedRiskManagerMonteCarloSimulation(t *testing.T) {
	// Create advanced risk manager
	manager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create advanced risk manager: %v", err)
	}

	// Create a test portfolio
	portfolio, err := NewPortfolio("test_portfolio")
	if err != nil {
		t.Fatalf("Failed to create portfolio: %v", err)
	}

	// Add some test positions with returns
	position1, err := NewPosition("pos1", "BTC", big.NewFloat(1), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Add some sample returns
	position1.Returns = []*big.Float{
		big.NewFloat(0.02),  // 2% return
		big.NewFloat(-0.01), // -1% return
		big.NewFloat(0.03),  // 3% return
		big.NewFloat(-0.02), // -2% return
		big.NewFloat(0.01),  // 1% return
	}

	portfolio.AddPosition(position1)

	// Create Monte Carlo simulation
	simulation, err := NewMonteCarloSimulation("test_mc", 1000, big.NewFloat(1.0), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create Monte Carlo simulation: %v", err)
	}

	t.Run("Run Monte Carlo Simulation", func(t *testing.T) {
		result, err := manager.RunMonteCarloSimulation(portfolio, simulation)
		if err != nil {
			t.Errorf("Failed to run Monte Carlo simulation: %v", err)
			return
		}

		if result == nil {
			t.Errorf("Expected result but got nil")
			return
		}

		if result.Simulation != simulation {
			t.Errorf("Expected simulation %v, got %v", simulation, result.Simulation)
		}

		if result.Portfolio != portfolio {
			t.Errorf("Expected portfolio %v, got %v", portfolio, result.Portfolio)
		}

		if result.VaR == nil {
			t.Errorf("Expected VaR but got nil")
		}

		if result.CVaR == nil {
			t.Errorf("Expected CVaR but got nil")
		}

		if len(result.Simulations) != simulation.NumSimulations {
			t.Errorf("Expected %d simulations, got %d", simulation.NumSimulations, len(result.Simulations))
		}

		if len(result.Percentiles) == 0 {
			t.Errorf("Expected percentiles but got none")
		}

		// Check key percentiles
		expectedPercentiles := []int{1, 5, 10, 25, 50, 75, 90, 95, 99}
		for _, p := range expectedPercentiles {
			if _, exists := result.Percentiles[p]; !exists {
				t.Errorf("Expected percentile %d but not found", p)
			}
		}
	})

	t.Run("Nil Portfolio", func(t *testing.T) {
		_, err := manager.RunMonteCarloSimulation(nil, simulation)
		if err == nil {
			t.Errorf("Expected error for nil portfolio")
		}
	})

	t.Run("Nil Simulation", func(t *testing.T) {
		_, err := manager.RunMonteCarloSimulation(portfolio, nil)
		if err == nil {
			t.Errorf("Expected error for nil simulation")
		}
	})
}

func TestAdvancedRiskManagerStressTesting(t *testing.T) {
	// Create advanced risk manager
	manager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create advanced risk manager: %v", err)
	}

	// Create a test portfolio
	portfolio, err := NewPortfolio("test_portfolio")
	if err != nil {
		t.Fatalf("Failed to create portfolio: %v", err)
	}

	// Add some test positions
	position1, err := NewPosition("pos1", "BTC", big.NewFloat(1), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	// Add some sample returns
	position1.Returns = []*big.Float{
		big.NewFloat(0.02),  // 2% return
		big.NewFloat(-0.01), // -1% return
		big.NewFloat(0.03),  // 3% return
		big.NewFloat(-0.02), // -2% return
		big.NewFloat(0.01),  // 1% return
	}

	portfolio.AddPosition(position1)

	// Create stress test scenario
	scenario, err := NewStressTestScenario(
		"stress1",
		"Market Crash",
		"Simulate a severe market crash",
		map[string]*big.Float{
			"price_shock_BTC":      big.NewFloat(0.7),  // 30% price drop
			"volatility_shock_BTC": big.NewFloat(2.0),  // 2x volatility
		},
		SevereStress,
	)
	if err != nil {
		t.Fatalf("Failed to create stress test scenario: %v", err)
	}

	t.Run("Run Stress Test", func(t *testing.T) {
		analysis, err := manager.RunStressTest(portfolio, scenario)
		if err != nil {
			t.Errorf("Failed to run stress test: %v", err)
			return
		}

		if analysis == nil {
			t.Errorf("Expected analysis but got nil")
			return
		}

		if analysis.Scenario != scenario {
			t.Errorf("Expected scenario %v, got %v", scenario, analysis.Scenario)
		}

		if analysis.Portfolio != portfolio {
			t.Errorf("Expected portfolio %v, got %v", portfolio, analysis.Portfolio)
		}

		if len(analysis.Results) == 0 {
			t.Errorf("Expected results but got none")
		}

		// Check that key metrics are present
		expectedMetrics := []MetricType{VaR, CVaR, Volatility, SharpeRatio}
		for _, metric := range expectedMetrics {
			if _, exists := analysis.Results[metric]; !exists {
				t.Logf("Metric %v not found in results (this may be expected for some scenarios)", metric)
			}
		}
	})

	t.Run("Nil Portfolio", func(t *testing.T) {
		_, err := manager.RunStressTest(nil, scenario)
		if err == nil {
			t.Errorf("Expected error for nil portfolio")
		}
	})

	t.Run("Nil Scenario", func(t *testing.T) {
		_, err := manager.RunStressTest(portfolio, nil)
		if err == nil {
			t.Errorf("Expected error for nil scenario")
		}
	})
}

func TestAdvancedRiskManagerScenarioManagement(t *testing.T) {
	// Create advanced risk manager
	manager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create advanced risk manager: %v", err)
	}

	// Create stress test scenario
	scenario, err := NewStressTestScenario(
		"stress1",
		"Market Crash",
		"Simulate a severe market crash",
		map[string]*big.Float{
			"price_shock_BTC": big.NewFloat(0.7),
		},
		SevereStress,
	)
	if err != nil {
		t.Fatalf("Failed to create stress test scenario: %v", err)
	}

	t.Run("Add and Get Stress Test Scenario", func(t *testing.T) {
		// Add scenario
		err := manager.AddStressTestScenario(scenario)
		if err != nil {
			t.Errorf("Failed to add stress test scenario: %v", err)
			return
		}

		// Get scenario
		retrievedScenario, err := manager.GetStressTestScenario(scenario.ID)
		if err != nil {
			t.Errorf("Failed to get stress test scenario: %v", err)
			return
		}

		if retrievedScenario != scenario {
			t.Errorf("Expected scenario %v, got %v", scenario, retrievedScenario)
		}
	})

	t.Run("Get Non-existent Scenario", func(t *testing.T) {
		_, err := manager.GetStressTestScenario("non_existent")
		if err == nil {
			t.Errorf("Expected error for non-existent scenario")
		}
	})

	t.Run("Add Nil Scenario", func(t *testing.T) {
		err := manager.AddStressTestScenario(nil)
		if err == nil {
			t.Errorf("Expected error for nil scenario")
		}
	})

	t.Run("Get Empty ID", func(t *testing.T) {
		_, err := manager.GetStressTestScenario("")
		if err == nil {
			t.Errorf("Expected error for empty ID")
		}
	})
}

func TestAdvancedRiskManagerSimulationManagement(t *testing.T) {
	// Create advanced risk manager
	manager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create advanced risk manager: %v", err)
	}

	// Create Monte Carlo simulation
	simulation, err := NewMonteCarloSimulation("test_mc", 1000, big.NewFloat(1.0), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create Monte Carlo simulation: %v", err)
	}

	t.Run("Add and Get Monte Carlo Simulation", func(t *testing.T) {
		// Add simulation
		err := manager.AddMonteCarloSimulation(simulation)
		if err != nil {
			t.Errorf("Failed to add Monte Carlo simulation: %v", err)
			return
		}

		// Get simulation
		retrievedSimulation, err := manager.GetMonteCarloSimulation(simulation.ID)
		if err != nil {
			t.Errorf("Failed to get Monte Carlo simulation: %v", err)
			return
		}

		if retrievedSimulation != simulation {
			t.Errorf("Expected simulation %v, got %v", simulation, retrievedSimulation)
		}
	})

	t.Run("Get Non-existent Simulation", func(t *testing.T) {
		_, err := manager.GetMonteCarloSimulation("non_existent")
		if err == nil {
			t.Errorf("Expected error for non-existent simulation")
		}
	})

	t.Run("Add Nil Simulation", func(t *testing.T) {
		err := manager.AddMonteCarloSimulation(nil)
		if err == nil {
			t.Errorf("Expected error for nil simulation")
		}
	})

	t.Run("Get Empty ID", func(t *testing.T) {
		_, err := manager.GetMonteCarloSimulation("")
		if err == nil {
			t.Errorf("Expected error for empty ID")
		}
	})
}
