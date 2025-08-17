package synthetic

import (
	"fmt"
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

func TestNewAsset(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		symbol        string
		assetType     AssetType
		price         *big.Float
		decimals      int
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid asset",
			id:          "BTC",
			symbol:      "BTC",
			assetType:   Token,
			price:       big.NewFloat(50000),
			decimals:    8,
			expectError: false,
		},
		{
			name:          "Empty ID",
			id:            "",
			symbol:        "BTC",
			assetType:     Token,
			price:         big.NewFloat(50000),
			decimals:      8,
			expectError:   true,
			errorContains: "asset ID cannot be empty",
		},
		{
			name:          "Empty symbol",
			id:            "BTC",
			symbol:        "",
			assetType:     Token,
			price:         big.NewFloat(50000),
			decimals:      8,
			expectError:   true,
			errorContains: "asset symbol cannot be empty",
		},
		{
			name:          "Negative price",
			id:            "BTC",
			symbol:        "BTC",
			assetType:     Token,
			price:         big.NewFloat(-50000),
			decimals:      8,
			expectError:   true,
			errorContains: "asset price must be non-negative",
		},
		{
			name:          "Negative decimals",
			id:            "BTC",
			symbol:        "BTC",
			assetType:     Token,
			price:         big.NewFloat(50000),
			decimals:      -1,
			expectError:   true,
			errorContains: "decimals must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := NewAsset(tt.id, tt.symbol, tt.assetType, tt.price, tt.decimals)

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

			if asset == nil {
				t.Errorf("Expected asset but got nil")
				return
			}

			// Verify asset properties
			if asset.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, asset.ID)
			}
			if asset.Symbol != tt.symbol {
				t.Errorf("Expected symbol %s, got %s", tt.symbol, asset.Symbol)
			}
			if asset.Type != tt.assetType {
				t.Errorf("Expected type %v, got %v", tt.assetType, asset.Type)
			}
			if asset.Price.Cmp(tt.price) != 0 {
				t.Errorf("Expected price %v, got %v", tt.price, asset.Price)
			}
			if asset.Decimals != tt.decimals {
				t.Errorf("Expected decimals %d, got %d", tt.decimals, asset.Decimals)
			}

			// Verify initial values
			if asset.Weight.Sign() != 0 {
				t.Error("Initial weight should be zero")
			}
		})
	}
}

func TestNewBasket(t *testing.T) {
	tests := []struct {
		name               string
		id                 string
		nameStr            string
		description        string
		rebalanceThreshold *big.Float
		expectError        bool
		errorContains      string
	}{
		{
			name:               "Valid basket",
			id:                 "DEFI_BASKET",
			nameStr:            "DeFi Index",
			description:        "A basket of DeFi tokens",
			rebalanceThreshold: big.NewFloat(0.05),
			expectError:        false,
		},
		{
			name:               "Empty ID",
			id:                 "",
			nameStr:            "DeFi Index",
			description:        "A basket of DeFi tokens",
			rebalanceThreshold: big.NewFloat(0.05),
			expectError:        true,
			errorContains:      "basket ID cannot be empty",
		},
		{
			name:               "Empty name",
			id:                 "DEFI_BASKET",
			nameStr:            "",
			description:        "A basket of DeFi tokens",
			rebalanceThreshold: big.NewFloat(0.05),
			expectError:        true,
			errorContains:      "basket name cannot be empty",
		},
		{
			name:               "Negative threshold",
			id:                 "DEFI_BASKET",
			nameStr:            "DeFi Index",
			description:        "A basket of DeFi tokens",
			rebalanceThreshold: big.NewFloat(-0.05),
			expectError:        true,
			errorContains:      "rebalance threshold must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basket, err := NewBasket(tt.id, tt.nameStr, tt.description, tt.rebalanceThreshold)

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

			if basket == nil {
				t.Errorf("Expected basket but got nil")
				return
			}

			// Verify basket properties
			if basket.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, basket.ID)
			}
			if basket.Name != tt.nameStr {
				t.Errorf("Expected name %s, got %s", tt.nameStr, basket.Name)
			}
			if basket.Description != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, basket.Description)
			}
			if basket.RebalanceThreshold.Cmp(tt.rebalanceThreshold) != 0 {
				t.Errorf("Expected rebalance threshold %v, got %v", tt.rebalanceThreshold, basket.RebalanceThreshold)
			}

			// Verify initial values
			if len(basket.Assets) != 0 {
				t.Error("Initial assets should be empty")
			}
			if len(basket.Weights) != 0 {
				t.Error("Initial weights should be empty")
			}
			if basket.TotalWeight.Sign() != 0 {
				t.Error("Initial total weight should be zero")
			}
		})
	}
}

func TestNewSyntheticToken(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	tests := []struct {
		name          string
		id            string
		symbol        string
		basket        *Basket
		mintFee       *big.Float
		redeemFee     *big.Float
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid token",
			id:          "SYNTH_DEFI",
			symbol:      "sDEFI",
			basket:      basket,
			mintFee:     big.NewFloat(0.001),
			redeemFee:   big.NewFloat(0.001),
			expectError: false,
		},
		{
			name:          "Empty ID",
			id:            "",
			symbol:        "sDEFI",
			basket:        basket,
			mintFee:       big.NewFloat(0.001),
			redeemFee:     big.NewFloat(0.001),
			expectError:   true,
			errorContains: "token ID cannot be empty",
		},
		{
			name:          "Empty symbol",
			id:            "SYNTH_DEFI",
			symbol:        "",
			basket:        basket,
			mintFee:       big.NewFloat(0.001),
			redeemFee:     big.NewFloat(0.001),
			expectError:   true,
			errorContains: "token symbol cannot be empty",
		},
		{
			name:          "Nil basket",
			id:            "SYNTH_DEFI",
			symbol:        "sDEFI",
			basket:        nil,
			mintFee:       big.NewFloat(0.001),
			redeemFee:     big.NewFloat(0.001),
			expectError:   true,
			errorContains: "basket cannot be nil",
		},
		{
			name:          "Negative mint fee",
			id:            "SYNTH_DEFI",
			symbol:        "sDEFI",
			basket:        basket,
			mintFee:       big.NewFloat(-0.001),
			redeemFee:     big.NewFloat(0.001),
			expectError:   true,
			errorContains: "mint fee must be non-negative",
		},
		{
			name:          "Negative redeem fee",
			id:            "SYNTH_DEFI",
			symbol:        "sDEFI",
			basket:        basket,
			mintFee:       big.NewFloat(0.001),
			redeemFee:     big.NewFloat(-0.001),
			expectError:   true,
			errorContains: "redeem fee must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewSyntheticToken(tt.id, tt.symbol, tt.basket, tt.mintFee, tt.redeemFee)

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

			if token == nil {
				t.Errorf("Expected token but got nil")
				return
			}

			// Verify token properties
			if token.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, token.ID)
			}
			if token.Symbol != tt.symbol {
				t.Errorf("Expected symbol %s, got %s", tt.symbol, token.Symbol)
			}
			if token.Basket != tt.basket {
				t.Errorf("Expected basket %v, got %v", tt.basket, token.Basket)
			}
			if token.MintFee.Cmp(tt.mintFee) != 0 {
				t.Errorf("Expected mint fee %v, got %v", tt.mintFee, token.MintFee)
			}
			if token.RedeemFee.Cmp(tt.redeemFee) != 0 {
				t.Errorf("Expected redeem fee %v, got %v", tt.redeemFee, token.RedeemFee)
			}

			// Verify initial values
			if token.TotalSupply.Sign() != 0 {
				t.Error("Initial total supply should be zero")
			}
			if token.UnderlyingValue.Sign() != 0 {
				t.Error("Initial underlying value should be zero")
			}
			if token.CollateralRatio.Cmp(big.NewFloat(1)) != 0 {
				t.Error("Initial collateral ratio should be 1")
			}
		})
	}
}

func TestBasketAddAsset(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	asset, err := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	if err != nil {
		t.Fatalf("Failed to create test asset: %v", err)
	}

	// Test valid asset addition
	err = basket.AddAsset(asset, big.NewFloat(0.3))
	if err != nil {
		t.Fatalf("Failed to add asset: %v", err)
	}

	// Verify asset was added
	if _, exists := basket.Assets[asset.ID]; !exists {
		t.Error("Asset not found in basket")
	}

	if basket.Weights[asset.ID].Cmp(big.NewFloat(0.3)) != 0 {
		t.Error("Asset weight not set correctly")
	}

	if basket.TotalWeight.Cmp(big.NewFloat(0.3)) != 0 {
		t.Error("Total weight not updated correctly")
	}

	// Test adding duplicate asset
	err = basket.AddAsset(asset, big.NewFloat(0.5))
	if err == nil {
		t.Error("Expected error when adding duplicate asset")
	}

	// Test adding nil asset
	err = basket.AddAsset(nil, big.NewFloat(0.5))
	if err == nil {
		t.Error("Expected error when adding nil asset")
	}

	// Test adding asset with negative weight
	err = basket.AddAsset(asset, big.NewFloat(-0.5))
	if err == nil {
		t.Error("Expected error when adding asset with negative weight")
	}
}

func TestBasketRemoveAsset(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	asset, err := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	if err != nil {
		t.Fatalf("Failed to create test asset: %v", err)
	}

	// Add asset first
	err = basket.AddAsset(asset, big.NewFloat(0.3))
	if err != nil {
		t.Fatalf("Failed to add asset: %v", err)
	}

	// Test removing asset
	err = basket.RemoveAsset(asset.ID)
	if err != nil {
		t.Fatalf("Failed to remove asset: %v", err)
	}

	// Verify asset was removed
	if _, exists := basket.Assets[asset.ID]; exists {
		t.Error("Asset still exists in basket")
	}

	if basket.TotalWeight.Sign() != 0 {
		t.Error("Total weight not reset to zero")
	}

	// Test removing non-existent asset
	err = basket.RemoveAsset("NON_EXISTENT")
	if err == nil {
		t.Error("Expected error when removing non-existent asset")
	}

	// Test removing asset with empty ID
	err = basket.RemoveAsset("")
	if err == nil {
		t.Error("Expected error when removing asset with empty ID")
	}
}

func TestBasketGetBasketValue(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets with different prices and weights
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	// Calculate expected value: (20 * 0.6) + (300 * 0.4) = 12 + 120 = 132
	expectedValue := big.NewFloat(132)
	basketValue := basket.GetBasketValue()

	compareBigFloat(t, expectedValue, basketValue, 0.01, "Basket value")
}

func TestBasketGetAssetAllocation(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	allocation := basket.GetAssetAllocation()

	// UNI allocation: (20 * 0.6) / 132 = 0.0909...
	expectedUniAllocation := big.NewFloat(0.0909)
	compareBigFloat(t, expectedUniAllocation, allocation["UNI"], 0.01, "UNI allocation")

	// AAVE allocation: (300 * 0.4) / 132 = 0.909...
	expectedAaveAllocation := big.NewFloat(0.909)
	compareBigFloat(t, expectedAaveAllocation, allocation["AAVE"], 0.01, "AAVE allocation")
}

func TestBasketCheckRebalanceNeeded(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.1)) // 10% threshold
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets with weights that account for price differences
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	// Use weights that will result in roughly equal allocations
	// UNI: 20 * 15 = 300, AAVE: 300 * 1 = 300
	basket.AddAsset(uni, big.NewFloat(15))
	basket.AddAsset(aave, big.NewFloat(1))

	// Initially, rebalancing should be needed due to price differences
	if !basket.CheckRebalanceNeeded() {
		t.Error("Rebalancing should be needed initially due to price differences")
	}

	// Change UNI price significantly to trigger rebalancing
	basket.UpdateAssetPrice("UNI", big.NewFloat(40)) // Double the price

	// Now rebalancing should be needed
	if !basket.CheckRebalanceNeeded() {
		t.Error("Rebalancing should be needed after significant price change")
	}
}

func TestBasketRebalance(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.1))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets with weights that account for price differences
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	// Use weights that will result in roughly equal allocations
	// UNI: 20 * 15 = 300, AAVE: 300 * 1 = 300
	basket.AddAsset(uni, big.NewFloat(15))
	basket.AddAsset(aave, big.NewFloat(1))

	// Change UNI price to trigger rebalancing
	basket.UpdateAssetPrice("UNI", big.NewFloat(40))

	// Perform rebalancing
	event, err := basket.Rebalance()
	if err != nil {
		t.Fatalf("Failed to rebalance: %v", err)
	}

	if event == nil {
		t.Error("Rebalance event should not be nil")
	}

	if event.BasketID != basket.ID {
		t.Error("Rebalance event basket ID mismatch")
	}

	// Verify rebalancing was performed
	if !basket.LastRebalanced.After(event.Timestamp.Add(-time.Second)) {
		t.Error("Last rebalanced time not updated")
	}
}

func TestSyntheticTokenMint(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets to basket
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	token, err := NewSyntheticToken("SYNTH_DEFI", "sDEFI", basket, big.NewFloat(0.001), big.NewFloat(0.001))
	if err != nil {
		t.Fatalf("Failed to create synthetic token: %v", err)
	}

	// Mint tokens
	amount := big.NewFloat(100)
	tokensMinted, err := token.Mint(amount, "user1")
	if err != nil {
		t.Fatalf("Failed to mint tokens: %v", err)
	}

	// Verify tokens were minted
	if tokensMinted.Cmp(amount) <= 0 {
		t.Error("Tokens minted should be greater than amount due to fees")
	}

	if token.TotalSupply.Cmp(tokensMinted) != 0 {
		t.Error("Total supply not updated correctly")
	}

	// Verify underlying value was updated
	basketValue := basket.GetBasketValue()
	expectedUnderlyingValue := new(big.Float).Mul(amount, basketValue)
	compareBigFloat(t, expectedUnderlyingValue, token.UnderlyingValue, 0.01, "Underlying value")
}

func TestSyntheticTokenRedeem(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets to basket
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	token, err := NewSyntheticToken("SYNTH_DEFI", "sDEFI", basket, big.NewFloat(0.001), big.NewFloat(0.001))
	if err != nil {
		t.Fatalf("Failed to create synthetic token: %v", err)
	}

	// Mint tokens first
	amount := big.NewFloat(100)
	token.Mint(amount, "user1")

	// Redeem tokens
	redeemAmount := big.NewFloat(50)
	underlyingValue, err := token.Redeem(redeemAmount, "user1")
	if err != nil {
		t.Fatalf("Failed to redeem tokens: %v", err)
	}

	// Verify tokens were redeemed
	// Initial supply was 100 + 0.1 (fee) = 100.1, redeemed 50, so remaining should be 50.1
	expectedRemainingSupply := big.NewFloat(50.1)
	compareBigFloat(t, expectedRemainingSupply, token.TotalSupply, 0.01, "Total supply after redemption")

	// Verify underlying value was reduced
	if underlyingValue.Sign() <= 0 {
		t.Error("Underlying value should be positive")
	}
}

func TestSyntheticTokenGetTokenPrice(t *testing.T) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		t.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets to basket
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	token, err := NewSyntheticToken("SYNTH_DEFI", "sDEFI", basket, big.NewFloat(0.001), big.NewFloat(0.001))
	if err != nil {
		t.Fatalf("Failed to create synthetic token: %v", err)
	}

	// Initially, token price should be zero (no supply)
	tokenPrice := token.GetTokenPrice()
	if tokenPrice.Sign() != 0 {
		t.Error("Initial token price should be zero")
	}

	// Mint tokens
	amount := big.NewFloat(100)
	token.Mint(amount, "user1")

	// Now token price should equal basket value per token
	basketValue := basket.GetBasketValue()
	expectedTokenPrice := new(big.Float).Quo(basketValue, amount)
	tokenPrice = token.GetTokenPrice()

	compareBigFloat(t, expectedTokenPrice, tokenPrice, 0.01, "Token price")
}

func BenchmarkBasketGetBasketValue(b *testing.B) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		b.Fatalf("Failed to create test basket: %v", err)
	}

	// Add multiple assets
	for i := 0; i < 10; i++ {
		asset, _ := NewAsset(fmt.Sprintf("ASSET_%d", i), fmt.Sprintf("ASSET_%d", i), Token, big.NewFloat(float64(i+1)*10), 18)
		basket.AddAsset(asset, big.NewFloat(0.1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		basket.GetBasketValue()
	}
}

func BenchmarkSyntheticTokenMint(b *testing.B) {
	basket, err := NewBasket("DEFI_BASKET", "DeFi Index", "A basket of DeFi tokens", big.NewFloat(0.05))
	if err != nil {
		b.Fatalf("Failed to create test basket: %v", err)
	}

	// Add assets
	uni, _ := NewAsset("UNI", "UNI", Token, big.NewFloat(20), 18)
	aave, _ := NewAsset("AAVE", "AAVE", Token, big.NewFloat(300), 18)

	basket.AddAsset(uni, big.NewFloat(0.6))
	basket.AddAsset(aave, big.NewFloat(0.4))

	token, err := NewSyntheticToken("SYNTH_DEFI", "sDEFI", basket, big.NewFloat(0.001), big.NewFloat(0.001))
	if err != nil {
		b.Fatalf("Failed to create synthetic token: %v", err)
	}

	amount := big.NewFloat(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token.Mint(amount, "user1")
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
