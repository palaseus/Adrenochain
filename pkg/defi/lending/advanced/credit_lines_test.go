package advanced

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRiskManager implements RiskManager for testing
type MockRiskManager struct{}

func (m *MockRiskManager) AssessCreditRisk(userID string, creditLimit *big.Int, collateralRatio *big.Float) (CreditRiskLevel, error) {
	// Simple risk assessment based on credit limit
	if creditLimit.Cmp(big.NewInt(1000000)) > 0 {
		return CreditRiskLevelHigh, nil
	} else if creditLimit.Cmp(big.NewInt(500000)) > 0 {
		return CreditRiskLevelMedium, nil
	}
	return CreditRiskLevelLow, nil
}

func (m *MockRiskManager) CalculateInterestRate(riskLevel CreditRiskLevel, marketRate *big.Float) *big.Float {
	// Simple interest rate calculation based on risk
	baseRate := new(big.Float).Copy(marketRate)
	switch riskLevel {
	case CreditRiskLevelLow:
		return new(big.Float).Add(baseRate, big.NewFloat(0.02)) // 2% premium
	case CreditRiskLevelMedium:
		return new(big.Float).Add(baseRate, big.NewFloat(0.05)) // 5% premium
	case CreditRiskLevelHigh:
		return new(big.Float).Add(baseRate, big.NewFloat(0.10)) // 10% premium
	default:
		return new(big.Float).Add(baseRate, big.NewFloat(0.15)) // 15% premium
	}
}

func (m *MockRiskManager) ValidateCollateral(assets []CollateralAsset, outstandingBalance *big.Int) error {
	// Simple collateral validation
	if len(assets) == 0 && outstandingBalance.Cmp(big.NewInt(0)) > 0 {
		return assert.AnError
	}
	return nil
}

func TestNewCreditLineManager(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	assert.NotNil(t, clm)
	assert.NotNil(t, clm.creditLines)
	assert.Equal(t, riskManager, clm.riskManager)
}

func TestCreateCreditLine(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)   // 1M
	interestRate := big.NewFloat(0.08)   // 8%
	collateralRatio := big.NewFloat(1.5) // 150%

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)
	assert.NotNil(t, creditLine)

	// Verify credit line properties
	assert.Equal(t, userID, creditLine.UserID)
	assert.Equal(t, CreditLineTypeRevolving, creditLine.Type)
	assert.Equal(t, CreditLineStatusActive, creditLine.Status)
	assert.Equal(t, creditLimit, creditLine.CreditLimit)
	assert.Equal(t, creditLimit, creditLine.AvailableCredit)
	assert.Equal(t, big.NewInt(0), creditLine.OutstandingBalance)
	assert.Equal(t, interestRate, creditLine.InterestRate)
	assert.Equal(t, collateralRatio, creditLine.CollateralRatio)
	assert.Equal(t, 0, creditLine.UtilizationRate.Cmp(big.NewFloat(0)))
	assert.Equal(t, CreditRiskLevelMedium, creditLine.RiskLevel)
	assert.True(t, creditLine.CreatedAt.After(time.Now().Add(-1*time.Minute)))
	assert.True(t, creditLine.UpdatedAt.After(time.Now().Add(-1*time.Minute)))
}

func TestCreateCreditLineValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"

	// Test invalid credit limit
	_, err := clm.CreateCreditLine(ctx, userID, big.NewInt(0), big.NewFloat(0.08), big.NewFloat(1.5))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credit limit must be positive")

	// Test negative interest rate
	_, err = clm.CreateCreditLine(ctx, userID, big.NewInt(1000000), big.NewFloat(-0.01), big.NewFloat(1.5))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest rate cannot be negative")

	// Test invalid collateral ratio
	_, err = clm.CreateCreditLine(ctx, userID, big.NewInt(1000000), big.NewFloat(0.08), big.NewFloat(0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collateral ratio must be positive")
}

func TestGetCreditLine(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, creditLine.ID, retrieved.ID)

	// Test non-existent credit line
	_, err = clm.GetCreditLine("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetUserCreditLines(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	user1 := "user1"
	user2 := "user2"

	// Create credit lines for user1
	cl1, err := clm.CreateCreditLine(ctx, user1, big.NewInt(1000000), big.NewFloat(0.08), big.NewFloat(1.5))
	require.NoError(t, err)
	cl2, err := clm.CreateCreditLine(ctx, user1, big.NewInt(500000), big.NewFloat(0.06), big.NewFloat(2.0))
	require.NoError(t, err)

	// Create credit line for user2
	_, err = clm.CreateCreditLine(ctx, user2, big.NewInt(2000000), big.NewFloat(0.10), big.NewFloat(1.8))
	require.NoError(t, err)

	// Test user1 credit lines
	user1CreditLines, err := clm.GetUserCreditLines(user1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(user1CreditLines))

	// Verify both credit lines are returned
	ids := make(map[string]bool)
	for _, cl := range user1CreditLines {
		ids[cl.ID] = true
	}
	assert.True(t, ids[cl1.ID])
	assert.True(t, ids[cl2.ID])
}

func TestDrawdown(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test successful drawdown
	drawdownAmount := big.NewInt(500000)
	err = clm.Drawdown(ctx, creditLine.ID, drawdownAmount, "Business expansion")
	require.NoError(t, err)

	// Verify credit line state
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, drawdownAmount, updatedCL.OutstandingBalance)
	assert.Equal(t, big.NewInt(500000), updatedCL.AvailableCredit)
	assert.Equal(t, 0, updatedCL.UtilizationRate.Cmp(big.NewFloat(0.5)))
	assert.Equal(t, 1, len(updatedCL.DrawdownHistory))

	// Test drawdown validation
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(600000), "Too much")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient available credit")
}

func TestDrawdownValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test invalid amount
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(0), "Invalid amount")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "drawdown amount must be positive")

	// Test non-existent credit line
	err = clm.Drawdown(ctx, "non-existent", big.NewInt(100000), "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMakePayment(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown first
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(500000), "Initial drawdown")
	require.NoError(t, err)

	// Test principal payment
	paymentAmount := big.NewInt(200000)
	err = clm.MakePayment(ctx, creditLine.ID, paymentAmount, PaymentTypePrincipal, "Principal payment")
	require.NoError(t, err)

	// Verify credit line state
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(300000), updatedCL.OutstandingBalance)
	assert.Equal(t, big.NewInt(700000), updatedCL.AvailableCredit)
	// Check utilization rate with tolerance for floating-point precision
	expectedRate := big.NewFloat(0.3)
	diff := new(big.Float).Sub(updatedCL.UtilizationRate, expectedRate)
	diff.Abs(diff)
	assert.True(t, diff.Cmp(big.NewFloat(0.0001)) <= 0, "Utilization rate should be approximately 0.3, got %v", updatedCL.UtilizationRate)
	assert.Equal(t, 1, len(updatedCL.PaymentHistory))

	// Test overpayment
	overpaymentAmount := big.NewInt(400000)
	err = clm.MakePayment(ctx, creditLine.ID, overpaymentAmount, PaymentTypePrincipal, "Overpayment")
	require.NoError(t, err)

	// Verify overpayment handling
	overpaidCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), overpaidCL.OutstandingBalance)
	// After overpayment: 700k (from previous payment) + 100k (overpayment) = 800k
	assert.Equal(t, big.NewInt(800000), overpaidCL.AvailableCredit)
	// Check utilization rate with tolerance for floating-point precision
	expectedRate = big.NewFloat(0)
	diff = new(big.Float).Sub(overpaidCL.UtilizationRate, expectedRate)
	diff.Abs(diff)
	assert.True(t, diff.Cmp(big.NewFloat(0.0001)) <= 0, "Utilization rate should be approximately 0, got %v", overpaidCL.UtilizationRate)
}

func TestPaymentValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test invalid payment amount
	err = clm.MakePayment(ctx, creditLine.ID, big.NewInt(0), PaymentTypePrincipal, "Invalid amount")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment amount must be positive")

	// Test non-existent credit line
	err = clm.MakePayment(ctx, "non-existent", big.NewInt(100000), PaymentTypePrincipal, "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAddCollateral(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown first
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(500000), "Initial drawdown")
	require.NoError(t, err)

	// Add collateral
	assetID := "BTC"
	amount := big.NewInt(1000000000) // 1 BTC in satoshis
	value := big.NewInt(600000)      // $600k value
	err = clm.AddCollateral(ctx, creditLine.ID, assetID, amount, value)
	require.NoError(t, err)

	// Verify collateral was added
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(updatedCL.CollateralAssets))
	assert.Equal(t, assetID, updatedCL.CollateralAssets[0].AssetID)
	assert.Equal(t, amount, updatedCL.CollateralAssets[0].Amount)
	assert.Equal(t, value, updatedCL.CollateralAssets[0].Value)
	// Check collateral ratio with tolerance for floating-point precision
	expectedRatio := big.NewFloat(1.2)
	diff := new(big.Float).Sub(updatedCL.CollateralRatio, expectedRatio)
	diff.Abs(diff)
	assert.True(t, diff.Cmp(big.NewFloat(0.0001)) <= 0, "Collateral ratio should be approximately 1.2, got %v", updatedCL.CollateralRatio)
}

func TestAddCollateralValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test invalid amount
	err = clm.AddCollateral(ctx, creditLine.ID, "BTC", big.NewInt(0), big.NewInt(100000))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collateral amount must be positive")

	// Test invalid value
	err = clm.AddCollateral(ctx, creditLine.ID, "BTC", big.NewInt(1000000000), big.NewInt(0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collateral value must be positive")

	// Test non-existent credit line
	err = clm.AddCollateral(ctx, "non-existent", "BTC", big.NewInt(1000000000), big.NewInt(100000))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveCollateral(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown and add collateral
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(500000), "Initial drawdown")
	require.NoError(t, err)

	err = clm.AddCollateral(ctx, creditLine.ID, "BTC", big.NewInt(1000000000), big.NewInt(600000))
	require.NoError(t, err)

	// Remove partial collateral
	removeAmount := big.NewInt(500000000) // 0.5 BTC
	err = clm.RemoveCollateral(ctx, creditLine.ID, "BTC", removeAmount)
	require.NoError(t, err)

	// Verify collateral was partially removed
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(updatedCL.CollateralAssets))
	assert.Equal(t, big.NewInt(500000000), updatedCL.CollateralAssets[0].Amount)
	assert.Equal(t, big.NewInt(300000), updatedCL.CollateralAssets[0].Value) // Value reduced proportionally
}

func TestUpdateCreditLimit(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(300000), "Initial drawdown")
	require.NoError(t, err)

	// Increase credit limit
	newLimit := big.NewInt(1500000)
	err = clm.UpdateCreditLimit(ctx, creditLine.ID, newLimit)
	require.NoError(t, err)

	// Verify credit limit was updated
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, newLimit, updatedCL.CreditLimit)
	assert.Equal(t, big.NewInt(1200000), updatedCL.AvailableCredit) // 1.5M - 300k = 1.2M
	// Check utilization rate with tolerance for floating-point precision
	expectedRate := big.NewFloat(0.2)
	diff := new(big.Float).Sub(updatedCL.UtilizationRate, expectedRate)
	diff.Abs(diff)
	assert.True(t, diff.Cmp(big.NewFloat(0.0001)) <= 0, "Utilization rate should be approximately 0.2, got %v", updatedCL.UtilizationRate)
}

func TestUpdateCreditLimitValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(800000), "Initial drawdown")
	require.NoError(t, err)

	// Test invalid new limit (less than outstanding balance)
	err = clm.UpdateCreditLimit(ctx, creditLine.ID, big.NewInt(500000))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be less than outstanding balance")

	// Test invalid new limit (zero)
	err = clm.UpdateCreditLimit(ctx, creditLine.ID, big.NewInt(0))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credit limit must be positive")
}

func TestUpdateInterestRate(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Update interest rate
	newRate := big.NewFloat(0.10)
	err = clm.UpdateInterestRate(ctx, creditLine.ID, newRate)
	require.NoError(t, err)

	// Verify interest rate was updated
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, newRate, updatedCL.InterestRate)
}

func TestUpdateInterestRateValidation(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test negative interest rate
	err = clm.UpdateInterestRate(ctx, creditLine.ID, big.NewFloat(-0.01))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest rate cannot be negative")
}

func TestCloseCreditLine(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Close credit line
	err = clm.CloseCreditLine(ctx, creditLine.ID)
	require.NoError(t, err)

	// Verify credit line was closed
	updatedCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.Equal(t, CreditLineStatusClosed, updatedCL.Status)
}

func TestCloseCreditLineWithBalance(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(300000), "Initial drawdown")
	require.NoError(t, err)

	// Try to close credit line with outstanding balance
	err = clm.CloseCreditLine(ctx, creditLine.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot close credit line with outstanding balance")
}

func TestGetCreditLineStats(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Make a drawdown and add collateral
	err = clm.Drawdown(ctx, creditLine.ID, big.NewInt(500000), "Initial drawdown")
	require.NoError(t, err)

	err = clm.AddCollateral(ctx, creditLine.ID, "BTC", big.NewInt(1000000000), big.NewInt(600000))
	require.NoError(t, err)

	// Get stats
	stats, err := clm.GetCreditLineStats(creditLine.ID)
	require.NoError(t, err)

	// Verify key stats
	assert.Equal(t, creditLine.ID, stats["id"])
	assert.Equal(t, userID, stats["user_id"])
	assert.Equal(t, CreditLineStatusActive, stats["status"])
	assert.Equal(t, "1000000", stats["credit_limit"])
	assert.Equal(t, "500000", stats["available_credit"])
	assert.Equal(t, "500000", stats["outstanding_balance"])
	assert.Equal(t, "0.08", stats["interest_rate"])
	assert.Equal(t, "1.2", stats["collateral_ratio"])
	assert.Equal(t, "0.5", stats["utilization_rate"])
	assert.Equal(t, 1, stats["drawdown_count"])
	assert.Equal(t, 1, stats["collateral_asset_count"])
}

func TestConcurrentCreditLineOperations(t *testing.T) {
	riskManager := &MockRiskManager{}
	clm := NewCreditLineManager(riskManager)

	ctx := context.Background()
	userID := "user1"
	creditLimit := big.NewInt(1000000)
	interestRate := big.NewFloat(0.08)
	collateralRatio := big.NewFloat(1.5)

	creditLine, err := clm.CreateCreditLine(ctx, userID, creditLimit, interestRate, collateralRatio)
	require.NoError(t, err)

	// Test concurrent drawdowns
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			amount := big.NewInt(int64(10000 * (id + 1)))
			err := clm.Drawdown(ctx, creditLine.ID, amount, fmt.Sprintf("Concurrent drawdown %d", id))
			// Some drawdowns should fail due to insufficient credit
			if err != nil {
				// Expected for some concurrent operations
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify final state is consistent
	finalCL, err := clm.GetCreditLine(creditLine.ID)
	require.NoError(t, err)
	assert.True(t, finalCL.OutstandingBalance.Cmp(creditLimit) <= 0)
	assert.True(t, finalCL.AvailableCredit.Cmp(big.NewInt(0)) >= 0)
}
