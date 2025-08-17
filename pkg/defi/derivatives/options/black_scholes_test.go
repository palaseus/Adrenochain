package options

import (
	"math"
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
	
	if math.Abs(expectedVal-actualVal) > tolerance {
		t.Errorf("%s: expected %.6f, got %.6f", message, expectedVal, actualVal)
	}
}

func TestNewBlackScholesModel(t *testing.T) {
	tests := []struct {
		name          string
		riskFreeRate  *big.Float
		volatility    *big.Float
		timeToExpiry  *big.Float
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid parameters",
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			timeToExpiry:  big.NewFloat(1.0),
			expectError:   false,
		},
		{
			name:          "Nil risk-free rate",
			riskFreeRate:  nil,
			volatility:    big.NewFloat(0.25),
			timeToExpiry:  big.NewFloat(1.0),
			expectError:   true,
			errorContains: "all parameters must be non-nil",
		},
		{
			name:          "Negative risk-free rate",
			riskFreeRate:  big.NewFloat(-0.01),
			volatility:    big.NewFloat(0.25),
			timeToExpiry:  big.NewFloat(1.0),
			expectError:   true,
			errorContains: "risk-free rate must be non-negative",
		},
		{
			name:          "Zero volatility",
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.0),
			timeToExpiry:  big.NewFloat(1.0),
			expectError:   true,
			errorContains: "volatility must be positive",
		},
		{
			name:          "Negative volatility",
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(-0.25),
			timeToExpiry:  big.NewFloat(1.0),
			expectError:   true,
			errorContains: "volatility must be positive",
		},
		{
			name:          "Zero time to expiry",
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			timeToExpiry:  big.NewFloat(0.0),
			expectError:   true,
			errorContains: "time to expiry must be positive",
		},
		{
			name:          "Negative time to expiry",
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			timeToExpiry:  big.NewFloat(-1.0),
			expectError:   true,
			errorContains: "time to expiry must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := NewBlackScholesModel(tt.riskFreeRate, tt.volatility, tt.timeToExpiry)
			
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
			
			if model == nil {
				t.Errorf("Expected model but got nil")
				return
			}
			
			// Verify parameters are copied correctly
			if model.RiskFreeRate.Cmp(tt.riskFreeRate) != 0 {
				t.Errorf("Risk-free rate not copied correctly")
			}
			if model.Volatility.Cmp(tt.volatility) != 0 {
				t.Errorf("Volatility not copied correctly")
			}
			if model.TimeToExpiry.Cmp(tt.timeToExpiry) != 0 {
				t.Errorf("Time to expiry not copied correctly")
			}
		})
	}
}

func TestNewOption(t *testing.T) {
	tests := []struct {
		name          string
		optionType    OptionType
		strikePrice   *big.Float
		currentPrice  *big.Float
		timeToExpiry  *big.Float
		riskFreeRate  *big.Float
		volatility    *big.Float
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid call option",
			optionType:    Call,
			strikePrice:   big.NewFloat(100.0),
			currentPrice:  big.NewFloat(110.0),
			timeToExpiry:  big.NewFloat(1.0),
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			expectError:   false,
		},
		{
			name:          "Valid put option",
			optionType:    Put,
			strikePrice:   big.NewFloat(100.0),
			currentPrice:  big.NewFloat(90.0),
			timeToExpiry:  big.NewFloat(1.0),
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			expectError:   false,
		},
		{
			name:          "Zero strike price",
			optionType:    Call,
			strikePrice:   big.NewFloat(0.0),
			currentPrice:  big.NewFloat(110.0),
			timeToExpiry:  big.NewFloat(1.0),
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			expectError:   true,
			errorContains: "strike price must be positive",
		},
		{
			name:          "Zero current price",
			optionType:    Call,
			strikePrice:   big.NewFloat(100.0),
			currentPrice:  big.NewFloat(0.0),
			timeToExpiry:  big.NewFloat(1.0),
			riskFreeRate:  big.NewFloat(0.05),
			volatility:    big.NewFloat(0.25),
			expectError:   true,
			errorContains: "current price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := NewOption(tt.optionType, tt.strikePrice, tt.currentPrice, tt.timeToExpiry, tt.riskFreeRate, tt.volatility)
			
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
			
			if option == nil {
				t.Errorf("Expected option but got nil")
				return
			}
			
			// Verify option properties
			if option.Type != tt.optionType {
				t.Errorf("Option type not set correctly")
			}
			if option.StrikePrice.Cmp(tt.strikePrice) != 0 {
				t.Errorf("Strike price not copied correctly")
			}
			if option.CurrentPrice.Cmp(tt.currentPrice) != 0 {
				t.Errorf("Current price not copied correctly")
			}
		})
	}
}

func TestBlackScholesPricing(t *testing.T) {
	// Test case: Known values from financial literature
	// S = 100, K = 100, T = 1, r = 0.05, σ = 0.25
	// Expected call price ≈ 12.34, put price ≈ 7.87
	
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		t.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	callOption, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create call option: %v", err)
	}
	
	putOption, err := NewOption(
		Put,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create put option: %v", err)
	}
	
	// Test call option pricing
	callPrice, err := model.Price(callOption)
	if err != nil {
		t.Fatalf("Failed to price call option: %v", err)
	}
	
	// Expected call price ≈ 12.34
	expectedCallPrice := big.NewFloat(12.34)
	compareBigFloat(t, expectedCallPrice, callPrice, 0.5, "Call option price")
	
	// Test put option pricing
	putPrice, err := model.Price(putOption)
	if err != nil {
		t.Fatalf("Failed to price put option: %v", err)
	}
	
	// Expected put price ≈ 7.87
	expectedPutPrice := big.NewFloat(7.87)
	compareBigFloat(t, expectedPutPrice, putPrice, 0.5, "Put option price")
	
	// Test put-call parity: C - P = S - K*e^(-rT)
	callPriceVal, _ := callPrice.Float64()
	putPriceVal, _ := putPrice.Float64()
	leftSide := callPriceVal - putPriceVal
	
	currentPrice, _ := callOption.CurrentPrice.Float64()
	strikePrice, _ := callOption.StrikePrice.Float64()
	riskFreeRate, _ := callOption.RiskFreeRate.Float64()
	timeToExpiry, _ := callOption.TimeToExpiry.Float64()
	
	discountFactor := math.Exp(-riskFreeRate * timeToExpiry)
	rightSide := currentPrice - strikePrice*discountFactor
	
	if math.Abs(leftSide-rightSide) > 0.01 {
		t.Errorf("Put-call parity violated: C-P=%.4f, S-K*e^(-rT)=%.4f", leftSide, rightSide)
	}
}

func TestBlackScholesGreeks(t *testing.T) {
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		t.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}
	
	// Test Delta
	delta, err := model.Delta(option)
	if err != nil {
		t.Fatalf("Failed to calculate Delta: %v", err)
	}
	
	// For ATM call option, delta should be approximately 0.5-0.7 depending on volatility and time
	// With 25% volatility and 1 year, delta should be around 0.63
	expectedDelta := big.NewFloat(0.63)
	compareBigFloat(t, expectedDelta, delta, 0.1, "Delta")
	
	// Test Gamma
	gamma, err := model.Gamma(option)
	if err != nil {
		t.Fatalf("Failed to calculate Gamma: %v", err)
	}
	
	// Gamma should be positive
	gammaVal, _ := gamma.Float64()
	if gammaVal <= 0 {
		t.Errorf("Gamma should be positive, got %f", gammaVal)
	}
	
	// Test Theta
	theta, err := model.Theta(option)
	if err != nil {
		t.Fatalf("Failed to calculate Theta: %v", err)
	}
	
	// Theta should be negative (time decay)
	thetaVal, _ := theta.Float64()
	if thetaVal >= 0 {
		t.Errorf("Theta should be negative for time decay, got %f", thetaVal)
	}
	
	// Test Vega
	vega, err := model.Vega(option)
	if err != nil {
		t.Fatalf("Failed to calculate Vega: %v", err)
	}
	
	// Vega should be positive
	vegaVal, _ := vega.Float64()
	if vegaVal <= 0 {
		t.Errorf("Vega should be positive, got %f", vegaVal)
	}
	
	// Test Rho
	rho, err := model.Rho(option)
	if err != nil {
		t.Fatalf("Failed to calculate Rho: %v", err)
	}
	
	// Rho should be positive for call options
	rhoVal, _ := rho.Float64()
	if rhoVal <= 0 {
		t.Errorf("Rho should be positive for call options, got %f", rhoVal)
	}
}

func TestBlackScholesEdgeCases(t *testing.T) {
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		t.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	// Test very short time to expiry
	shortTermOption, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(0.001), // 1 day to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create short-term option: %v", err)
	}
	
	shortTermPrice, err := model.Price(shortTermOption)
	if err != nil {
		t.Fatalf("Failed to price short-term option: %v", err)
	}
	
	// Short-term ATM option should have low extrinsic value
	shortTermPriceVal, _ := shortTermPrice.Float64()
	if shortTermPriceVal > 5.0 {
		t.Errorf("Short-term ATM option price too high: %f", shortTermPriceVal)
	}
	
	// Test very high volatility
	highVolOption, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // 1 year to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(1.0),   // 100% volatility
	)
	if err != nil {
		t.Fatalf("Failed to create high volatility option: %v", err)
	}
	
	highVolPrice, err := model.Price(highVolOption)
	if err != nil {
		t.Fatalf("Failed to price high volatility option: %v", err)
	}
	
	// High volatility should increase option price
	highVolPriceVal, _ := highVolPrice.Float64()
	if highVolPriceVal < 20.0 {
		t.Errorf("High volatility option price too low: %f", highVolPriceVal)
	}
}

func TestBlackScholesMathematicalProperties(t *testing.T) {
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		t.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	// Test monotonicity: option price should increase with underlying price
	prices := []float64{80.0, 90.0, 100.0, 110.0, 120.0}
	optionPrices := make([]float64, len(prices))
	
	for i, price := range prices {
		option, err := NewOption(
			Call,
			big.NewFloat(100.0), // Strike price
			big.NewFloat(price), // Current price
			big.NewFloat(1.0),   // Time to expiry
			big.NewFloat(0.05),  // Risk-free rate
			big.NewFloat(0.25),  // Volatility
		)
		if err != nil {
			t.Fatalf("Failed to create option with price %f: %v", price, err)
		}
		
		optionPrice, err := model.Price(option)
		if err != nil {
			t.Fatalf("Failed to price option with price %f: %v", price, err)
		}
		
		optionPrices[i], _ = optionPrice.Float64()
	}
	
	// Verify monotonicity
	for i := 1; i < len(optionPrices); i++ {
		if optionPrices[i] <= optionPrices[i-1] {
			t.Errorf("Option price should increase with underlying price: %f <= %f", optionPrices[i], optionPrices[i-1])
		}
	}
	
	// Test time decay: option price should decrease with time
	times := []float64{2.0, 1.5, 1.0, 0.5, 0.1}
	timePrices := make([]float64, len(times))
	
	for i, time := range times {
		option, err := NewOption(
			Call,
			big.NewFloat(100.0), // Strike price
			big.NewFloat(100.0), // Current price
			big.NewFloat(time),   // Time to expiry
			big.NewFloat(0.05),  // Risk-free rate
			big.NewFloat(0.25),  // Volatility
		)
		if err != nil {
			t.Fatalf("Failed to create option with time %f: %v", time, err)
		}
		
		optionPrice, err := model.Price(option)
		if err != nil {
			t.Fatalf("Failed to price option with time %f: %v", time, err)
		}
		
		timePrices[i], _ = optionPrice.Float64()
	}
	
	// Verify time decay (longer time = higher price)
	for i := 1; i < len(timePrices); i++ {
		if timePrices[i] >= timePrices[i-1] {
			t.Errorf("Option price should decrease with time: %f >= %f", timePrices[i], timePrices[i-1])
		}
	}
}

func BenchmarkBlackScholesPricing(b *testing.B) {
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		b.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		b.Fatalf("Failed to create option: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Price(option)
		if err != nil {
			b.Fatalf("Failed to price option: %v", err)
		}
	}
}

func BenchmarkBlackScholesGreeks(b *testing.B) {
	model, err := NewBlackScholesModel(
		big.NewFloat(0.05),  // 5% risk-free rate
		big.NewFloat(0.25),  // 25% volatility
		big.NewFloat(1.0),   // 1 year to expiry
	)
	if err != nil {
		b.Fatalf("Failed to create Black-Scholes model: %v", err)
	}
	
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(100.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		b.Fatalf("Failed to create option: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Calculate all Greeks
		_, err := model.Delta(option)
		if err != nil {
			b.Fatalf("Failed to calculate Delta: %v", err)
		}
		_, err = model.Gamma(option)
		if err != nil {
			b.Fatalf("Failed to calculate Gamma: %v", err)
		}
		_, err = model.Theta(option)
		if err != nil {
			b.Fatalf("Failed to calculate Theta: %v", err)
		}
		_, err = model.Vega(option)
		if err != nil {
			b.Fatalf("Failed to calculate Vega: %v", err)
		}
		_, err = model.Rho(option)
		if err != nil {
			b.Fatalf("Failed to calculate Rho: %v", err)
		}
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
