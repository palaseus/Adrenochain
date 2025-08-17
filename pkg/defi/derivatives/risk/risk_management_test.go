package risk

import (
	"fmt"
	"math/big"
	"testing"
)

// Helper function to create big.Float from string to avoid overflow issues
func bigFloatFromString(s string) *big.Float {
	v, _ := new(big.Float).SetString(s)
	return v
}

// Helper function to compare big.Float values with tolerance
func compareBigFloat(t *testing.T, expected, actual *big.Float, tolerance float64, message string) {
	expectedVal, _ := expected.Float64()
	actualVal, _ := actual.Float64()
	
	if abs(expectedVal-actualVal) > tolerance {
		t.Errorf("%s: expected %.6f, got %.6f", message, expectedVal, actualVal)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestNewRiskManager(t *testing.T) {
	tests := []struct {
		name            string
		riskFreeRate    *big.Float
		confidenceLevel *big.Float
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid risk manager",
			riskFreeRate:    big.NewFloat(0.02),
			confidenceLevel: big.NewFloat(0.95),
			expectError:     false,
		},
		{
			name:            "Nil risk-free rate",
			riskFreeRate:    nil,
			confidenceLevel: big.NewFloat(0.95),
			expectError:     true,
			errorContains:   "risk-free rate and confidence level cannot be nil",
		},
		{
			name:            "Nil confidence level",
			riskFreeRate:    big.NewFloat(0.02),
			confidenceLevel: nil,
			expectError:     true,
			errorContains:   "risk-free rate and confidence level cannot be nil",
		},
		{
			name:            "Confidence level too low",
			riskFreeRate:    big.NewFloat(0.02),
			confidenceLevel: big.NewFloat(0),
			expectError:     true,
			errorContains:   "confidence level must be between 0 and 1",
		},
		{
			name:            "Confidence level too high",
			riskFreeRate:    big.NewFloat(0.02),
			confidenceLevel: big.NewFloat(1.0),
			expectError:     true,
			errorContains:   "confidence level must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm, err := NewRiskManager(tt.riskFreeRate, tt.confidenceLevel)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if rm == nil {
				t.Errorf("Expected risk manager but got nil")
				return
			}
			
			// Verify properties
			if rm.RiskFreeRate.Cmp(tt.riskFreeRate) != 0 {
				t.Errorf("Expected risk-free rate %v, got %v", tt.riskFreeRate, rm.RiskFreeRate)
			}
			if rm.ConfidenceLevel.Cmp(tt.confidenceLevel) != 0 {
				t.Errorf("Expected confidence level %v, got %v", tt.confidenceLevel, rm.ConfidenceLevel)
			}
		})
	}
}

func TestNewPortfolio(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
		errorContains string
	}{
		{
			name:        "Valid portfolio",
			id:          "PORTFOLIO_1",
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			expectError: true,
			errorContains: "portfolio ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portfolio, err := NewPortfolio(tt.id)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if portfolio == nil {
				t.Errorf("Expected portfolio but got nil")
				return
			}
			
			// Verify properties
			if portfolio.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, portfolio.ID)
			}
			if len(portfolio.Positions) != 0 {
				t.Error("Initial positions should be empty")
			}
			if len(portfolio.RiskMetrics) != 0 {
				t.Error("Initial risk metrics should be empty")
			}
		})
	}
}

func TestNewPosition(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		assetID     string
		size        *big.Float
		price       *big.Float
		expectError bool
		errorContains string
	}{
		{
			name:        "Valid position",
			id:          "POS_1",
			assetID:     "BTC",
			size:        big.NewFloat(10),
			price:       big.NewFloat(50000),
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			assetID:     "BTC",
			size:        big.NewFloat(10),
			price:       big.NewFloat(50000),
			expectError: true,
			errorContains: "position ID cannot be empty",
		},
		{
			name:        "Empty asset ID",
			id:          "POS_1",
			assetID:     "",
			size:        big.NewFloat(10),
			price:       big.NewFloat(50000),
			expectError: true,
			errorContains: "asset ID cannot be empty",
		},
		{
			name:        "Zero size",
			id:          "POS_1",
			assetID:     "BTC",
			size:        big.NewFloat(0),
			price:       big.NewFloat(50000),
			expectError: true,
			errorContains: "position size must be positive",
		},
		{
			name:        "Zero price",
			id:          "POS_1",
			assetID:     "BTC",
			size:        big.NewFloat(10),
			price:       big.NewFloat(0),
			expectError: true,
			errorContains: "position price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, err := NewPosition(tt.id, tt.assetID, tt.size, tt.price)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if position == nil {
				t.Errorf("Expected position but got nil")
				return
			}
			
			// Verify properties
			if position.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, position.ID)
			}
			if position.AssetID != tt.assetID {
				t.Errorf("Expected asset ID %s, got %s", tt.assetID, position.AssetID)
			}
			if position.Size.Cmp(tt.size) != 0 {
				t.Errorf("Expected size %v, got %v", tt.size, position.Size)
			}
			if position.Price.Cmp(tt.price) != 0 {
				t.Errorf("Expected price %v, got %v", tt.price, position.Price)
			}
			
			// Verify calculated value
			expectedValue := new(big.Float).Mul(tt.size, tt.price)
			if position.Value.Cmp(expectedValue) != 0 {
				t.Errorf("Expected value %v, got %v", expectedValue, position.Value)
			}
		})
	}
}

func TestPortfolioAddPosition(t *testing.T) {
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	position, err := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Test adding position
	err = portfolio.AddPosition(position)
	if err != nil {
		t.Fatalf("Failed to add position: %v", err)
	}
	
	// Verify position was added
	if _, exists := portfolio.Positions[position.ID]; !exists {
		t.Error("Position not found in portfolio")
	}
	
	// Verify weight was calculated
	if position.Weight.Sign() <= 0 {
		t.Error("Position weight not calculated")
	}
	
	// Test adding nil position
	err = portfolio.AddPosition(nil)
	if err == nil {
		t.Error("Expected error when adding nil position")
	}
}

func TestPortfolioRemovePosition(t *testing.T) {
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	position, err := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Add position first
	portfolio.AddPosition(position)
	
	// Test removing position
	err = portfolio.RemovePosition(position.ID)
	if err != nil {
		t.Fatalf("Failed to remove position: %v", err)
	}
	
	// Verify position was removed
	if _, exists := portfolio.Positions[position.ID]; exists {
		t.Error("Position still exists in portfolio")
	}
	
	// Test removing non-existent position
	err = portfolio.RemovePosition("NON_EXISTENT")
	if err == nil {
		t.Error("Expected error when removing non-existent position")
	}
}

func TestPortfolioGetTotalValue(t *testing.T) {
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Initially, portfolio should have zero value
	totalValue := portfolio.GetTotalValue()
	if totalValue.Sign() != 0 {
		t.Error("Empty portfolio should have zero value")
	}
	
	// Add positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate expected total value
	expectedValue := big.NewFloat(0)
	expectedValue.Add(expectedValue, position1.Value)
	expectedValue.Add(expectedValue, position2.Value)
	
	totalValue = portfolio.GetTotalValue()
	compareBigFloat(t, expectedValue, totalValue, 0.01, "Total portfolio value")
}

func TestRiskManagerCalculateVaR(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate VaR
	timeHorizon := big.NewFloat(1.0) // 1 day
	varValue, err := rm.CalculateVaR(portfolio, timeHorizon)
	if err != nil {
		t.Fatalf("Failed to calculate VaR: %v", err)
	}
	
	// VaR should be negative (representing potential loss)
	if varValue.Sign() >= 0 {
		t.Error("VaR should be negative for potential loss")
	}
	
	// Verify VaR was stored in portfolio
	if portfolio.RiskMetrics[VaR] == nil {
		t.Error("VaR metric not stored in portfolio")
	}
	
	// Test with nil portfolio
	_, err = rm.CalculateVaR(nil, timeHorizon)
	if err == nil {
		t.Error("Expected error when portfolio is nil")
	}
	
	// Test with invalid time horizon
	_, err = rm.CalculateVaR(portfolio, big.NewFloat(-1))
	if err == nil {
		t.Error("Expected error when time horizon is negative")
	}
}

func TestRiskManagerCalculateCVaR(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate CVaR
	timeHorizon := big.NewFloat(1.0)
	cvarValue, err := rm.CalculateCVaR(portfolio, timeHorizon)
	if err != nil {
		t.Fatalf("Failed to calculate CVaR: %v", err)
	}
	
	// CVaR should be negative (representing expected loss beyond VaR)
	if cvarValue.Sign() >= 0 {
		t.Error("CVaR should be negative for expected loss beyond VaR")
	}
	
	// Verify CVaR was stored in portfolio
	if portfolio.RiskMetrics[CVaR] == nil {
		t.Error("CVaR metric not stored in portfolio")
	}
}

func TestRiskManagerCalculateVolatility(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate volatility
	volatility, err := rm.CalculateVolatility(portfolio)
	if err != nil {
		t.Fatalf("Failed to calculate volatility: %v", err)
	}
	
	// Volatility should be positive
	if volatility.Sign() <= 0 {
		t.Error("Volatility should be positive")
	}
	
	// Verify volatility was stored in portfolio
	if portfolio.RiskMetrics[Volatility] == nil {
		t.Error("Volatility metric not stored in portfolio")
	}
}

func TestRiskManagerCalculateSharpeRatio(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate Sharpe ratio
	_, err = rm.CalculateSharpeRatio(portfolio)
	if err != nil {
		t.Fatalf("Failed to calculate Sharpe ratio: %v", err)
	}
	
	// Verify Sharpe ratio was stored in portfolio
	if portfolio.RiskMetrics[SharpeRatio] == nil {
		t.Error("Sharpe ratio metric not stored in portfolio")
	}
}

func TestRiskManagerCalculateMaxDrawdown(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Calculate max drawdown
	maxDrawdown, err := rm.CalculateMaxDrawdown(portfolio)
	if err != nil {
		t.Fatalf("Failed to calculate max drawdown: %v", err)
	}
	
	// Max drawdown should be positive (representing loss percentage)
	if maxDrawdown.Sign() < 0 {
		t.Error("Max drawdown should be positive")
	}
	
	// Verify max drawdown was stored in portfolio
	if portfolio.RiskMetrics[MaxDrawdown] == nil {
		t.Error("Max drawdown metric not stored in portfolio")
	}
}

func TestRiskManagerCalculateBeta(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Create benchmark returns
	benchmarkReturns := []*big.Float{
		big.NewFloat(0.01),
		big.NewFloat(-0.005),
		big.NewFloat(0.02),
		big.NewFloat(-0.01),
		big.NewFloat(0.015),
	}
	
	// Calculate beta
	_, err = rm.CalculateBeta(portfolio, benchmarkReturns)
	if err != nil {
		t.Fatalf("Failed to calculate beta: %v", err)
	}
	
	// Verify beta was stored in portfolio
	if portfolio.RiskMetrics[Beta] == nil {
		t.Error("Beta metric not stored in portfolio")
	}
	
	// Test with empty benchmark returns
	_, err = rm.CalculateBeta(portfolio, []*big.Float{})
	if err == nil {
		t.Error("Expected error when benchmark returns are empty")
	}
}

func TestRiskManagerStressTest(t *testing.T) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		t.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add some positions
	position1, _ := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	position2, _ := NewPosition("POS_2", "ETH", big.NewFloat(100), big.NewFloat(3000))
	
	portfolio.AddPosition(position1)
	portfolio.AddPosition(position2)
	
	// Create stress scenarios
	scenarios := []StressScenario{
		{
			Name: "BTC Crash",
			AssetShocks: map[string]*big.Float{
				"BTC": big.NewFloat(-0.5), // 50% drop in BTC
			},
		},
		{
			Name: "Market Crash",
			AssetShocks: map[string]*big.Float{
				"BTC": big.NewFloat(-0.3), // 30% drop in BTC
				"ETH": big.NewFloat(-0.4), // 40% drop in ETH
			},
		},
	}
	
	// Perform stress test
	results, err := rm.StressTest(portfolio, scenarios)
	if err != nil {
		t.Fatalf("Failed to perform stress test: %v", err)
	}
	
	// Verify results
	if len(results) != len(scenarios) {
		t.Errorf("Expected %d results, got %d", len(scenarios), len(results))
	}
	
	// BTC Crash scenario should show negative impact
	if btcCrashResult, exists := results["BTC Crash"]; exists {
		if btcCrashResult.Sign() >= 0 {
			t.Error("BTC crash scenario should show negative impact")
		}
	}
	
	// Test with nil portfolio
	_, err = rm.StressTest(nil, scenarios)
	if err == nil {
		t.Error("Expected error when portfolio is nil")
	}
	
	// Test with empty scenarios
	_, err = rm.StressTest(portfolio, []StressScenario{})
	if err == nil {
		t.Error("Expected error when scenarios are empty")
	}
}

func TestPositionUpdatePosition(t *testing.T) {
	position, err := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Update position price
	newPrice := big.NewFloat(55000)
	err = position.UpdatePosition(newPrice)
	if err != nil {
		t.Fatalf("Failed to update position: %v", err)
	}
	
	// Verify price was updated
	if position.Price.Cmp(newPrice) != 0 {
		t.Error("Position price not updated")
	}
	
	// Verify value was recalculated
	expectedValue := new(big.Float).Mul(position.Size, newPrice)
	if position.Value.Cmp(expectedValue) != 0 {
		t.Error("Position value not recalculated")
	}
	
	// Test with negative price
	err = position.UpdatePosition(big.NewFloat(-1000))
	if err == nil {
		t.Error("Expected error when updating with negative price")
	}
	
	// Test with nil price
	err = position.UpdatePosition(nil)
	if err == nil {
		t.Error("Expected error when updating with nil price")
	}
}

func TestPositionAddReturn(t *testing.T) {
	position, err := NewPosition("POS_1", "BTC", big.NewFloat(10), big.NewFloat(50000))
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	
	// Initially, no returns
	if len(position.Returns) != 0 {
		t.Error("Initial returns should be empty")
	}
	
	// Add returns
	return1 := big.NewFloat(0.05)
	return2 := big.NewFloat(-0.02)
	
	position.AddReturn(return1)
	position.AddReturn(return2)
	
	// Verify returns were added
	if len(position.Returns) != 2 {
		t.Error("Returns not added correctly")
	}
	
	// Verify return values
	if position.Returns[0].Cmp(return1) != 0 {
		t.Error("First return not stored correctly")
	}
	if position.Returns[1].Cmp(return2) != 0 {
		t.Error("Second return not stored correctly")
	}
	
	// Test adding nil return
	position.AddReturn(nil)
	if len(position.Returns) != 2 {
		t.Error("Nil return should not be added")
	}
}

func BenchmarkRiskManagerCalculateVaR(b *testing.B) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		b.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		b.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add multiple positions
	for i := 0; i < 10; i++ {
		position, _ := NewPosition(
			fmt.Sprintf("POS_%d", i),
			fmt.Sprintf("ASSET_%d", i),
			big.NewFloat(float64(i+1)*1000),
			big.NewFloat(float64(i+1)*100),
		)
		portfolio.AddPosition(position)
	}
	
	timeHorizon := big.NewFloat(1.0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.CalculateVaR(portfolio, timeHorizon)
	}
}

func BenchmarkRiskManagerCalculateVolatility(b *testing.B) {
	rm, err := NewRiskManager(big.NewFloat(0.02), big.NewFloat(0.95))
	if err != nil {
		b.Fatalf("Failed to create risk manager: %v", err)
	}
	
	portfolio, err := NewPortfolio("TEST_PORTFOLIO")
	if err != nil {
		b.Fatalf("Failed to create test portfolio: %v", err)
	}
	
	// Add multiple positions
	for i := 0; i < 10; i++ {
		position, _ := NewPosition(
			fmt.Sprintf("POS_%d", i),
			fmt.Sprintf("ASSET_%d", i),
			big.NewFloat(float64(i+1)*1000),
			big.NewFloat(float64(i+1)*100),
		)
		portfolio.AddPosition(position)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.CalculateVolatility(portfolio)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
