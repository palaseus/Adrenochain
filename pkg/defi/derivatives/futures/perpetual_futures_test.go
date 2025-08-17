package futures

import (
	"math/big"
	"testing"
	"time"
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

func TestNewPerpetualContract(t *testing.T) {
	tests := []struct {
		name             string
		symbol           string
		underlyingAsset  string
		contractSize     *big.Float
		expectError      bool
		errorContains    string
	}{
		{
			name:            "Valid contract",
			symbol:          "BTC-PERP",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1.0),
			expectError:     false,
		},
		{
			name:            "Empty symbol",
			symbol:          "",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1.0),
			expectError:     true,
			errorContains:   "symbol cannot be empty",
		},
		{
			name:            "Empty underlying asset",
			symbol:          "BTC-PERP",
			underlyingAsset: "",
			contractSize:    big.NewFloat(1.0),
			expectError:     true,
			errorContains:   "underlying asset cannot be empty",
		},
		{
			name:            "Zero contract size",
			symbol:          "BTC-PERP",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(0),
			expectError:     true,
			errorContains:   "contract size must be positive",
		},
		{
			name:            "Negative contract size",
			symbol:          "BTC-PERP",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(-1.0),
			expectError:     true,
			errorContains:   "contract size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, err := NewPerpetualContract(tt.symbol, tt.underlyingAsset, tt.contractSize)
			
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
			
			if contract == nil {
				t.Errorf("Expected contract but got nil")
				return
			}
			
			// Verify contract properties
			if contract.Symbol != tt.symbol {
				t.Errorf("Expected symbol %s, got %s", tt.symbol, contract.Symbol)
			}
			if contract.UnderlyingAsset != tt.underlyingAsset {
				t.Errorf("Expected underlying asset %s, got %s", tt.underlyingAsset, contract.UnderlyingAsset)
			}
			if contract.ContractSize.Cmp(tt.contractSize) != 0 {
				t.Errorf("Expected contract size %v, got %v", tt.contractSize, contract.ContractSize)
			}
			
			// Verify funding rate is initialized
			if contract.FundingRate == nil {
				t.Error("Funding rate not initialized")
			}
			if contract.FundingRate.Rate.Sign() != 0 {
				t.Error("Initial funding rate should be zero")
			}
			if contract.FundingRate.Interval != 8*time.Hour {
				t.Error("Expected 8-hour funding interval")
			}
			
			// Verify next funding time is set
			if contract.NextFundingTime.Before(time.Now()) {
				t.Error("Next funding time should be in the future")
			}
		})
	}
}

func TestNewPerpetualPosition(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	tests := []struct {
		name        string
		userID      string
		contract    *PerpetualContract
		side        PositionSide
		size        *big.Float
		entryPrice  *big.Float
		leverage    *big.Float
		expectError bool
		errorContains string
	}{
		{
			name:        "Valid long position",
			userID:      "user1",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(10),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: false,
		},
		{
			name:        "Valid short position",
			userID:      "user2",
			contract:    contract,
			side:        Short,
			size:        big.NewFloat(5),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(5),
			expectError: false,
		},
		{
			name:        "Empty user ID",
			userID:      "",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(10),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
			errorContains: "user ID cannot be empty",
		},
		{
			name:        "Nil contract",
			userID:      "user1",
			contract:    nil,
			side:        Long,
			size:        big.NewFloat(10),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
			errorContains: "contract cannot be nil",
		},
		{
			name:        "Zero size",
			userID:      "user1",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(0),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
			errorContains: "position size must be positive",
		},
		{
			name:        "Zero entry price",
			userID:      "user1",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(10),
			entryPrice:  big.NewFloat(0),
			leverage:    big.NewFloat(10),
			expectError: true,
			errorContains: "entry price must be positive",
		},
		{
			name:        "Zero leverage",
			userID:      "user1",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(10),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(0),
			expectError: true,
			errorContains: "leverage must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, err := NewPerpetualPosition(tt.userID, tt.contract, tt.side, tt.size, tt.entryPrice, tt.leverage)
			
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
			
			// Verify position properties
			if position.UserID != tt.userID {
				t.Errorf("Expected user ID %s, got %s", tt.userID, position.UserID)
			}
			if position.Contract != tt.contract {
				t.Errorf("Expected contract %v, got %v", tt.contract, position.Contract)
			}
			if position.Side != tt.side {
				t.Errorf("Expected side %v, got %v", tt.side, position.Side)
			}
			if position.Size.Cmp(tt.size) != 0 {
				t.Errorf("Expected size %v, got %v", tt.size, position.Size)
			}
			if position.EntryPrice.Cmp(tt.entryPrice) != 0 {
				t.Errorf("Expected entry price %v, got %v", tt.entryPrice, position.EntryPrice)
			}
			if position.Leverage.Cmp(tt.leverage) != 0 {
				t.Errorf("Expected leverage %v, got %v", tt.leverage, position.Leverage)
			}
			
			// Verify initial values
			if position.UnrealizedPnL.Sign() != 0 {
				t.Error("Initial unrealized PnL should be zero")
			}
			if position.RealizedPnL.Sign() != 0 {
				t.Error("Initial realized PnL should be zero")
			}
			if position.FundingPaid.Sign() != 0 {
				t.Error("Initial funding paid should be zero")
			}
		})
	}
}

func TestPerpetualContractUpdateMarkPrice(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Test valid mark price update
	markPrice := big.NewFloat(50000)
	err = contract.UpdateMarkPrice(markPrice)
	if err != nil {
		t.Fatalf("Failed to update mark price: %v", err)
	}
	
	if contract.MarkPrice.Cmp(markPrice) != 0 {
		t.Errorf("Expected mark price %v, got %v", markPrice, contract.MarkPrice)
	}
	
	// Test 24h high/low updates
	higherPrice := big.NewFloat(51000)
	err = contract.UpdateMarkPrice(higherPrice)
	if err != nil {
		t.Fatalf("Failed to update mark price: %v", err)
	}
	
	if contract.High24h.Cmp(higherPrice) != 0 {
		t.Errorf("Expected 24h high %v, got %v", higherPrice, contract.High24h)
	}
	
	lowerPrice := big.NewFloat(49000)
	err = contract.UpdateMarkPrice(lowerPrice)
	if err != nil {
		t.Fatalf("Failed to update mark price: %v", err)
	}
	
	if contract.Low24h.Cmp(lowerPrice) != 0 {
		t.Errorf("Expected 24h low %v, got %v", lowerPrice, contract.Low24h)
	}
	
	// Test invalid mark price
	err = contract.UpdateMarkPrice(big.NewFloat(-1000))
	if err == nil {
		t.Error("Expected error for negative mark price")
	}
	
	err = contract.UpdateMarkPrice(nil)
	if err == nil {
		t.Error("Expected error for nil mark price")
	}
}

func TestPerpetualContractUpdateFundingRate(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Test valid funding rate update
	fundingRate := big.NewFloat(0.0001) // 0.01%
	err = contract.UpdateFundingRate(fundingRate)
	if err != nil {
		t.Fatalf("Failed to update funding rate: %v", err)
	}
	
	if contract.FundingRate.Rate.Cmp(fundingRate) != 0 {
		t.Errorf("Expected funding rate %v, got %v", fundingRate, contract.FundingRate.Rate)
	}
	
	// Verify next funding time was updated
	if contract.NextFundingTime.Before(time.Now()) {
		t.Error("Next funding time should be in the future")
	}
	
	// Test nil funding rate
	err = contract.UpdateFundingRate(nil)
	if err == nil {
		t.Error("Expected error for nil funding rate")
	}
}

func TestPerpetualContractCalculateFundingPayment(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Set funding rate
	fundingRate := big.NewFloat(0.0001) // 0.01%
	contract.UpdateFundingRate(fundingRate)
	
	// Create long position
	longPosition, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create long position: %v", err)
	}
	
	// Calculate funding payment for long position
	fundingPayment := contract.CalculateFundingPayment(longPosition)
	expectedPayment := big.NewFloat(50) // 10 * 50000 * 0.0001
	compareBigFloat(t, expectedPayment, fundingPayment, 0.01, "Long position funding payment")
	
	// Create short position
	shortPosition, err := NewPerpetualPosition("user2", contract, Short, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create short position: %v", err)
	}
	
	// Calculate funding payment for short position (should be negative)
	fundingPayment = contract.CalculateFundingPayment(shortPosition)
	expectedPayment = big.NewFloat(-50) // -10 * 50000 * 0.0001
	compareBigFloat(t, expectedPayment, fundingPayment, 0.01, "Short position funding payment")
	
	// Test with zero funding rate
	contract.UpdateFundingRate(big.NewFloat(0))
	fundingPayment = contract.CalculateFundingPayment(longPosition)
	if fundingPayment.Sign() != 0 {
		t.Error("Funding payment should be zero when funding rate is zero")
	}
}

func TestPerpetualPositionUpdatePosition(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create long position
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Update position with higher mark price (profit)
	newMarkPrice := big.NewFloat(51000)
	err = position.UpdatePosition(newMarkPrice)
	if err != nil {
		t.Fatalf("Failed to update position: %v", err)
	}
	
	// Verify unrealized PnL is positive
	if position.UnrealizedPnL.Sign() <= 0 {
		t.Error("Expected positive unrealized PnL for price increase")
	}
	
	// Verify PnL calculation: (51000 - 50000) * 10 = 10000
	expectedPnL := big.NewFloat(10000)
	compareBigFloat(t, expectedPnL, position.UnrealizedPnL, 0.01, "Long position PnL")
	
	// Update position with lower mark price (loss)
	newMarkPrice = big.NewFloat(49000)
	err = position.UpdatePosition(newMarkPrice)
	if err != nil {
		t.Fatalf("Failed to update position: %v", err)
	}
	
	// Verify unrealized PnL is negative
	if position.UnrealizedPnL.Sign() >= 0 {
		t.Error("Expected negative unrealized PnL for price decrease")
	}
	
	// Verify PnL calculation: (49000 - 50000) * 10 = -10000
	expectedPnL = big.NewFloat(-10000)
	compareBigFloat(t, expectedPnL, position.UnrealizedPnL, 0.01, "Long position PnL")
}

func TestPerpetualPositionCalculateMargin(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create position with 10x leverage
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Calculate margin
	margin := position.CalculateMargin()
	
	// Expected margin: (10 * 50000) / 10 = 50000
	expectedMargin := big.NewFloat(50000)
	compareBigFloat(t, expectedMargin, margin, 0.01, "Position margin")
	
	// Verify margin was stored
	if position.Margin.Cmp(margin) != 0 {
		t.Error("Margin not stored in position")
	}
}

func TestPerpetualPositionCalculateLiquidationPrice(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create long position with 10x leverage
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Set maintenance margin to 5%
	maintenanceMargin := big.NewFloat(0.05)
	
	// Calculate liquidation price
	liquidationPrice := position.CalculateLiquidationPrice(maintenanceMargin)
	
	// For long position: Entry Price * (1 - 1/Leverage + Maintenance Margin)
	// = 50000 * (1 - 1/10 + 0.05) = 50000 * 0.95 = 47500
	expectedLiquidationPrice := big.NewFloat(47500)
	compareBigFloat(t, expectedLiquidationPrice, liquidationPrice, 0.01, "Long position liquidation price")
	
	// Verify liquidation price was stored
	if position.LiquidationPrice.Cmp(liquidationPrice) != 0 {
		t.Error("Liquidation price not stored in position")
	}
	
	// Test short position
	shortPosition, err := NewPerpetualPosition("user2", contract, Short, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create short position: %v", err)
	}
	
	// Calculate liquidation price for short position
	liquidationPrice = shortPosition.CalculateLiquidationPrice(maintenanceMargin)
	
	// For short position: Entry Price * (1 + 1/Leverage - Maintenance Margin)
	// = 50000 * (1 + 1/10 - 0.05) = 50000 * 1.05 = 52500
	expectedLiquidationPrice = big.NewFloat(52500)
	compareBigFloat(t, expectedLiquidationPrice, liquidationPrice, 0.01, "Short position liquidation price")
}

func TestPerpetualPositionClosePosition(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create long position
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Add some funding payments
	position.AddFundingPayment(big.NewFloat(100))
	
	// Close position at profit
	exitPrice := big.NewFloat(51000)
	realizedPnL := position.ClosePosition(exitPrice)
	
	// Expected realized PnL: (51000 - 50000) * 10 - 100 = 9900
	expectedPnL := big.NewFloat(9900)
	compareBigFloat(t, expectedPnL, realizedPnL, 0.01, "Realized PnL")
	
	// Verify realized PnL was stored
	if position.RealizedPnL.Cmp(realizedPnL) != 0 {
		t.Error("Realized PnL not stored in position")
	}
	
	// Test short position
	shortPosition, err := NewPerpetualPosition("user2", contract, Short, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create short position: %v", err)
	}
	
	// Close short position at loss
	exitPrice = big.NewFloat(51000)
	realizedPnL = shortPosition.ClosePosition(exitPrice)
	
	// Expected realized PnL: (50000 - 51000) * 10 = -10000
	expectedPnL = big.NewFloat(-10000)
	compareBigFloat(t, expectedPnL, realizedPnL, 0.01, "Short position realized PnL")
}

func TestPerpetualPositionGetTotalPnL(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create position
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Update position to have unrealized PnL
	position.UpdatePosition(big.NewFloat(51000))
	
	// Close position to have realized PnL
	position.ClosePosition(big.NewFloat(51000))
	
	// Get total PnL
	totalPnL := position.GetTotalPnL()
	
	// Expected: Unrealized PnL (10000) + Realized PnL (10000) = 20000
	// Note: When closing at the same price as current mark price, unrealized PnL becomes realized
	expectedTotalPnL := big.NewFloat(20000)
	compareBigFloat(t, expectedTotalPnL, totalPnL, 0.01, "Total PnL")
}

func TestPerpetualPositionGetROI(t *testing.T) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}
	
	// Create position
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}
	
	// Calculate margin
	position.CalculateMargin()
	
	// Update position to have unrealized PnL
	position.UpdatePosition(big.NewFloat(51000))
	
	// Get ROI
	roi := position.GetROI()
	
	// Expected ROI: (10000 / 50000) * 100 = 20%
	expectedROI := big.NewFloat(20)
	compareBigFloat(t, expectedROI, roi, 0.01, "ROI")
}

func BenchmarkPerpetualContractUpdateMarkPrice(b *testing.B) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		b.Fatalf("Failed to create test contract: %v", err)
	}
	
	markPrice := big.NewFloat(50000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		markPrice.Add(markPrice, big.NewFloat(1))
		err := contract.UpdateMarkPrice(markPrice)
		if err != nil {
			b.Fatalf("Failed to update mark price: %v", err)
		}
	}
}

func BenchmarkPerpetualPositionUpdatePosition(b *testing.B) {
	contract, err := NewPerpetualContract("BTC-PERP", "BTC", big.NewFloat(1.0))
	if err != nil {
		b.Fatalf("Failed to create test contract: %v", err)
	}
	
	position, err := NewPerpetualPosition("user1", contract, Long, big.NewFloat(10), big.NewFloat(50000), big.NewFloat(10))
	if err != nil {
		b.Fatalf("Failed to create position: %v", err)
	}
	
	markPrice := big.NewFloat(50000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		markPrice.Add(markPrice, big.NewFloat(1))
		err := position.UpdatePosition(markPrice)
		if err != nil {
			b.Fatalf("Failed to update position: %v", err)
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
