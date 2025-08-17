package futures

import (
	"math/big"
	"testing"
	"time"
)

func TestNewStandardFuturesContract(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour) // 30 days from now
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour) // 2 days after expiration

	tests := []struct {
		name            string
		symbol          string
		underlyingAsset string
		contractSize    *big.Float
		strikePrice     *big.Float
		expirationDate  time.Time
		deliveryDate    time.Time
		contractType    ContractType
		settlementType  SettlementType
		expectError     bool
	}{
		{
			name:            "Valid Monthly Contract",
			symbol:          "BTC-2024-03",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     false,
		},
		{
			name:            "Valid Quarterly Contract",
			symbol:          "ETH-2024-Q2",
			underlyingAsset: "ETH",
			contractSize:    big.NewFloat(10),
			strikePrice:     big.NewFloat(3000),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Quarterly,
			settlementType:  PhysicalDelivery,
			expectError:     false,
		},
		{
			name:            "Empty Symbol",
			symbol:          "",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
		{
			name:            "Empty Underlying Asset",
			symbol:          "BTC-2024-03",
			underlyingAsset: "",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
		{
			name:            "Zero Contract Size",
			symbol:          "BTC-2024-03",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(0),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
		{
			name:            "Zero Strike Price",
			symbol:          "BTC-2024-03",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(0),
			expirationDate:  expirationDate,
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
		{
			name:            "Expiration Date in Past",
			symbol:          "BTC-2024-03",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  now.Add(-24 * time.Hour), // Yesterday
			deliveryDate:    deliveryDate,
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
		{
			name:            "Delivery Before Expiration",
			symbol:          "BTC-2024-03",
			underlyingAsset: "BTC",
			contractSize:    big.NewFloat(1),
			strikePrice:     big.NewFloat(50000),
			expirationDate:  expirationDate,
			deliveryDate:    now.Add(15 * 24 * time.Hour), // Before expiration
			contractType:    Monthly,
			settlementType:  CashSettlement,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, err := NewStandardFuturesContract(
				tt.symbol,
				tt.underlyingAsset,
				tt.contractSize,
				tt.strikePrice,
				tt.expirationDate,
				tt.deliveryDate,
				tt.contractType,
				tt.settlementType,
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

			if contract == nil {
				t.Errorf("Expected contract but got nil")
				return
			}

			if contract.Symbol != tt.symbol {
				t.Errorf("Expected symbol %s, got %s", tt.symbol, contract.Symbol)
			}

			if contract.UnderlyingAsset != tt.underlyingAsset {
				t.Errorf("Expected underlying asset %s, got %s", tt.underlyingAsset, contract.UnderlyingAsset)
			}

			if contract.ContractSize.Cmp(tt.contractSize) != 0 {
				t.Errorf("Expected contract size %v, got %v", tt.contractSize, contract.ContractSize)
			}

			if contract.StrikePrice.Cmp(tt.strikePrice) != 0 {
				t.Errorf("Expected strike price %v, got %v", tt.strikePrice, contract.StrikePrice)
			}

			if !contract.ExpirationDate.Equal(tt.expirationDate) {
				t.Errorf("Expected expiration date %v, got %v", tt.expirationDate, contract.ExpirationDate)
			}

			if !contract.DeliveryDate.Equal(tt.deliveryDate) {
				t.Errorf("Expected delivery date %v, got %v", tt.deliveryDate, contract.DeliveryDate)
			}

			if contract.ContractType != tt.contractType {
				t.Errorf("Expected contract type %v, got %v", tt.contractType, contract.ContractType)
			}

			if contract.SettlementType != tt.settlementType {
				t.Errorf("Expected settlement type %v, got %v", tt.settlementType, contract.SettlementType)
			}

			if !contract.IsActive {
				t.Errorf("Expected contract to be active")
			}
		})
	}
}

func TestStandardFuturesContractExpiration(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	t.Run("Not Expired", func(t *testing.T) {
		if contract.IsExpired() {
			t.Errorf("Expected contract to not be expired")
		}
	})

	t.Run("Days to Expiration", func(t *testing.T) {
		days := contract.DaysToExpiration()
		if days <= 0 {
			t.Errorf("Expected positive days to expiration, got %d", days)
		}
	})

	t.Run("Not Near Expiration", func(t *testing.T) {
		if contract.IsNearExpiration() {
			t.Errorf("Expected contract to not be near expiration")
		}
	})

	t.Run("Near Expiration", func(t *testing.T) {
		// Create a contract that expires in 5 days
		nearExpiration := now.Add(5 * 24 * time.Hour)
		nearDelivery := nearExpiration.Add(2 * 24 * time.Hour)

		nearContract, err := NewStandardFuturesContract(
			"BTC-2024-03-NEAR",
			"BTC",
			big.NewFloat(1),
			big.NewFloat(50000),
			nearExpiration,
			nearDelivery,
			Monthly,
			CashSettlement,
		)
		if err != nil {
			t.Fatalf("Failed to create near expiration contract: %v", err)
		}

		if !nearContract.IsNearExpiration() {
			t.Errorf("Expected contract to be near expiration")
		}
	})
}

func TestStandardFuturesContractPriceUpdates(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	t.Run("Update Mark Price", func(t *testing.T) {
		newPrice := big.NewFloat(55000)
		err := contract.UpdateMarkPrice(newPrice)
		if err != nil {
			t.Errorf("Failed to update mark price: %v", err)
		}

		if contract.MarkPrice.Cmp(newPrice) != 0 {
			t.Errorf("Expected mark price %v, got %v", newPrice, contract.MarkPrice)
		}

		// Check 24h high/low updates
		if contract.High24h.Cmp(newPrice) != 0 {
			t.Errorf("Expected 24h high %v, got %v", newPrice, contract.High24h)
		}

		if contract.Low24h.Cmp(newPrice) != 0 {
			t.Errorf("Expected 24h low %v, got %v", newPrice, contract.Low24h)
		}
	})

	t.Run("Update Index Price", func(t *testing.T) {
		newPrice := big.NewFloat(54000)
		err := contract.UpdateIndexPrice(newPrice)
		if err != nil {
			t.Errorf("Failed to update index price: %v", err)
		}

		if contract.IndexPrice.Cmp(newPrice) != 0 {
			t.Errorf("Expected index price %v, got %v", newPrice, contract.IndexPrice)
		}
	})

	t.Run("Update Volume", func(t *testing.T) {
		volume := big.NewFloat(1000)
		err := contract.UpdateVolume(volume)
		if err != nil {
			t.Errorf("Failed to update volume: %v", err)
		}

		if contract.Volume24h.Cmp(volume) != 0 {
			t.Errorf("Expected volume %v, got %v", volume, contract.Volume24h)
		}
	})

	t.Run("Update Open Interest", func(t *testing.T) {
		openInterest := big.NewFloat(500)
		err := contract.UpdateOpenInterest(openInterest)
		if err != nil {
			t.Errorf("Failed to update open interest: %v", err)
		}

		if contract.OpenInterest.Cmp(openInterest) != 0 {
			t.Errorf("Expected open interest %v, got %v", openInterest, contract.OpenInterest)
		}
	})

	t.Run("Invalid Price Updates", func(t *testing.T) {
		// Test nil price
		err := contract.UpdateMarkPrice(nil)
		if err == nil {
			t.Errorf("Expected error for nil price")
		}

		// Test negative price
		negativePrice := big.NewFloat(-1000)
		err = contract.UpdateMarkPrice(negativePrice)
		if err == nil {
			t.Errorf("Expected error for negative price")
		}
	})
}

func TestStandardFuturesContractValues(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	// Set mark price
	contract.UpdateMarkPrice(big.NewFloat(55000))

	t.Run("Contract Value", func(t *testing.T) {
		expectedValue := big.NewFloat(55000) // 1 * 55000
		contractValue := contract.GetContractValue()
		if contractValue.Cmp(expectedValue) != 0 {
			t.Errorf("Expected contract value %v, got %v", expectedValue, contractValue)
		}
	})

	t.Run("Strike Value", func(t *testing.T) {
		expectedValue := big.NewFloat(50000) // 1 * 50000
		strikeValue := contract.GetStrikeValue()
		if strikeValue.Cmp(expectedValue) != 0 {
			t.Errorf("Expected strike value %v, got %v", expectedValue, strikeValue)
		}
	})
}

func TestNewStandardFuturesPosition(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	tests := []struct {
		name        string
		userID      string
		contract    *StandardFuturesContract
		side        PositionSide
		size        *big.Float
		entryPrice  *big.Float
		leverage    *big.Float
		expectError bool
	}{
		{
			name:        "Valid Long Position",
			userID:      "user123",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(1),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: false,
		},
		{
			name:        "Valid Short Position",
			userID:      "user456",
			contract:    contract,
			side:        Short,
			size:        big.NewFloat(2),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(5),
			expectError: false,
		},
		{
			name:        "Empty User ID",
			userID:      "",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(1),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
		},
		{
			name:        "Nil Contract",
			userID:      "user123",
			contract:    nil,
			side:        Long,
			size:        big.NewFloat(1),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
		},
		{
			name:        "Zero Size",
			userID:      "user123",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(0),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(10),
			expectError: true,
		},
		{
			name:        "Zero Entry Price",
			userID:      "user123",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(1),
			entryPrice:  big.NewFloat(0),
			leverage:    big.NewFloat(10),
			expectError: true,
		},
		{
			name:        "Zero Leverage",
			userID:      "user123",
			contract:    contract,
			side:        Long,
			size:        big.NewFloat(1),
			entryPrice:  big.NewFloat(50000),
			leverage:    big.NewFloat(0),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, err := NewStandardFuturesPosition(
				tt.userID,
				tt.contract,
				tt.side,
				tt.size,
				tt.entryPrice,
				tt.leverage,
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

			if position == nil {
				t.Errorf("Expected position but got nil")
				return
			}

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

			if !position.IsOpen {
				t.Errorf("Expected position to be open")
			}

			// Check margin calculation
			expectedMargin := new(big.Float).Quo(tt.entryPrice, tt.leverage)
			if position.Margin.Cmp(expectedMargin) != 0 {
				t.Errorf("Expected margin %v, got %v", expectedMargin, position.Margin)
			}

			// Check liquidation price calculation
			if tt.side == Long {
				expectedLiquidation := new(big.Float).Mul(tt.entryPrice, big.NewFloat(0.8))
				if position.LiquidationPrice.Cmp(expectedLiquidation) != 0 {
					t.Errorf("Expected liquidation price %v, got %v", expectedLiquidation, position.LiquidationPrice)
				}
			} else {
				expectedLiquidation := new(big.Float).Mul(tt.entryPrice, big.NewFloat(1.2))
				if position.LiquidationPrice.Cmp(expectedLiquidation) != 0 {
					t.Errorf("Expected liquidation price %v, got %v", expectedLiquidation, position.LiquidationPrice)
				}
			}
		})
	}
}

func TestStandardFuturesPositionUpdates(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	position, err := NewStandardFuturesPosition(
		"user123",
		contract,
		Long,
		big.NewFloat(1),
		big.NewFloat(50000),
		big.NewFloat(10),
	)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	t.Run("Update Position - Price Increase", func(t *testing.T) {
		newPrice := big.NewFloat(55000)
		err := position.UpdatePosition(newPrice)
		if err != nil {
			t.Errorf("Failed to update position: %v", err)
		}

		if position.MarkPrice.Cmp(newPrice) != 0 {
			t.Errorf("Expected mark price %v, got %v", newPrice, position.MarkPrice)
		}

		// Check unrealized PnL calculation
		expectedPnL := big.NewFloat(5000) // (55000 - 50000) * 1
		if position.UnrealizedPnL.Cmp(expectedPnL) != 0 {
			t.Errorf("Expected unrealized PnL %v, got %v", expectedPnL, position.UnrealizedPnL)
		}
	})

	t.Run("Update Position - Price Decrease", func(t *testing.T) {
		newPrice := big.NewFloat(45000)
		err := position.UpdatePosition(newPrice)
		if err != nil {
			t.Errorf("Failed to update position: %v", err)
		}

		if position.MarkPrice.Cmp(newPrice) != 0 {
			t.Errorf("Expected mark price %v, got %v", newPrice, position.MarkPrice)
		}

		// Check unrealized PnL calculation
		expectedPnL := big.NewFloat(-5000) // (45000 - 50000) * 1
		if position.UnrealizedPnL.Cmp(expectedPnL) != 0 {
			t.Errorf("Expected unrealized PnL %v, got %v", expectedPnL, position.UnrealizedPnL)
		}
	})

	t.Run("Invalid Price Update", func(t *testing.T) {
		// Test nil price
		err := position.UpdatePosition(nil)
		if err == nil {
			t.Errorf("Expected error for nil price")
		}

		// Test negative price
		negativePrice := big.NewFloat(-1000)
		err = position.UpdatePosition(negativePrice)
		if err == nil {
			t.Errorf("Expected error for negative price")
		}
	})
}

func TestStandardFuturesPositionClose(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	position, err := NewStandardFuturesPosition(
		"user123",
		contract,
		Long,
		big.NewFloat(1),
		big.NewFloat(50000),
		big.NewFloat(10),
	)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	t.Run("Close Position - Profit", func(t *testing.T) {
		closePrice := big.NewFloat(55000)
		err := position.ClosePosition(closePrice)
		if err != nil {
			t.Errorf("Failed to close position: %v", err)
		}

		if position.IsOpen {
			t.Errorf("Expected position to be closed")
		}

		if position.CloseTime == nil {
			t.Errorf("Expected close time to be set")
		}

		// Check realized PnL
		expectedPnL := big.NewFloat(5000) // (55000 - 50000) * 1
		if position.RealizedPnL.Cmp(expectedPnL) != 0 {
			t.Errorf("Expected realized PnL %v, got %v", expectedPnL, position.RealizedPnL)
		}

		// Unrealized PnL should be zero after closing
		if position.UnrealizedPnL.Cmp(big.NewFloat(0)) != 0 {
			t.Errorf("Expected unrealized PnL to be 0, got %v", position.UnrealizedPnL)
		}
	})

	t.Run("Close Already Closed Position", func(t *testing.T) {
		closePrice := big.NewFloat(60000)
		err := position.ClosePosition(closePrice)
		if err == nil {
			t.Errorf("Expected error when closing already closed position")
		}
	})

	t.Run("Invalid Close Price", func(t *testing.T) {
		// Reset position to open state for testing
		position.IsOpen = true
		position.CloseTime = nil

		// Test nil price
		err := position.ClosePosition(nil)
		if err == nil {
			t.Errorf("Expected error for nil close price")
		}

		// Test negative price
		negativePrice := big.NewFloat(-1000)
		err = position.ClosePosition(negativePrice)
		if err == nil {
			t.Errorf("Expected error for negative close price")
		}
	})
}

func TestStandardFuturesPositionCalculations(t *testing.T) {
	now := time.Now()
	expirationDate := now.Add(30 * 24 * time.Hour)
	deliveryDate := expirationDate.Add(2 * 24 * time.Hour)

	contract, err := NewStandardFuturesContract(
		"BTC-2024-03",
		"BTC",
		big.NewFloat(1),
		big.NewFloat(50000),
		expirationDate,
		deliveryDate,
		Monthly,
		CashSettlement,
	)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	position, err := NewStandardFuturesPosition(
		"user123",
		contract,
		Long,
		big.NewFloat(1),
		big.NewFloat(50000),
		big.NewFloat(10),
	)
	if err != nil {
		t.Fatalf("Failed to create position: %v", err)
	}

	t.Run("Total PnL", func(t *testing.T) {
		// Set some unrealized PnL
		position.UnrealizedPnL = big.NewFloat(2000)
		position.RealizedPnL = big.NewFloat(1000)

		totalPnL := position.GetTotalPnL()
		expectedTotal := big.NewFloat(3000) // 2000 + 1000
		if totalPnL.Cmp(expectedTotal) != 0 {
			t.Errorf("Expected total PnL %v, got %v", expectedTotal, totalPnL)
		}
	})

	t.Run("ROI", func(t *testing.T) {
		// Set some PnL
		position.UnrealizedPnL = big.NewFloat(5000)
		position.RealizedPnL = big.NewFloat(0)

		roi := position.GetROI()
		// Expected ROI: 5000 / 5000 = 1.0 (100%)
		expectedROI := big.NewFloat(1.0)
		if roi.Cmp(expectedROI) != 0 {
			t.Errorf("Expected ROI %v, got %v", expectedROI, roi)
		}
	})

	t.Run("Margin Ratio", func(t *testing.T) {
		// Set mark price
		position.MarkPrice = big.NewFloat(55000)

		marginRatio := position.GetMarginRatio()
		// Expected: 5000 / (55000 * 1) = 5000 / 55000 ≈ 0.0909
		// Use a tolerance for floating point precision
		expectedRatio := big.NewFloat(0.0909)
		tolerance := big.NewFloat(0.0001)
		
		diff := new(big.Float).Sub(marginRatio, expectedRatio)
		absDiff := new(big.Float).Abs(diff)
		
		if absDiff.Cmp(tolerance) > 0 {
			t.Errorf("Expected margin ratio %v ± %v, got %v", expectedRatio, tolerance, marginRatio)
		}
	})

	t.Run("Liquidation Check", func(t *testing.T) {
		// Test liquidation for long position
		// Liquidation price is 80% of entry price = 40000
		position.MarkPrice = big.NewFloat(39000) // Below liquidation price

		if !position.IsLiquidated() {
			t.Errorf("Expected position to be liquidated")
		}

		// Test not liquidated
		position.MarkPrice = big.NewFloat(45000) // Above liquidation price
		if position.IsLiquidated() {
			t.Errorf("Expected position to not be liquidated")
		}
	})
}
