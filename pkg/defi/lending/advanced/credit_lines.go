package advanced

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// CreditLineStatus represents the status of a credit line
type CreditLineStatus string

const (
	CreditLineStatusActive    CreditLineStatus = "active"
	CreditLineStatusSuspended CreditLineStatus = "suspended"
	CreditLineStatusClosed    CreditLineStatus = "closed"
	CreditLineStatusDefaulted CreditLineStatus = "defaulted"
)

// CreditLineType represents the type of credit line
type CreditLineType string

const (
	CreditLineTypeRevolving CreditLineType = "revolving"
	CreditLineTypeTerm      CreditLineType = "term"
	CreditLineTypeOverdraft CreditLineType = "overdraft"
)

// CreditLine represents a revolving credit line
type CreditLine struct {
	ID                 string            `json:"id"`
	UserID             string            `json:"user_id"`
	Type               CreditLineType    `json:"type"`
	Status             CreditLineStatus  `json:"status"`
	CreditLimit        *big.Int          `json:"credit_limit"`
	AvailableCredit    *big.Int          `json:"available_credit"`
	OutstandingBalance *big.Int          `json:"outstanding_balance"`
	InterestRate       *big.Float        `json:"interest_rate"`
	OriginationFee     *big.Float        `json:"origination_fee"`
	AnnualFee          *big.Float        `json:"annual_fee"`
	CollateralRatio    *big.Float        `json:"collateral_ratio"`
	MinCollateralRatio *big.Float        `json:"min_collateral_ratio"`
	UtilizationRate    *big.Float        `json:"utilization_rate"`
	CreditScore        int               `json:"credit_score"`
	RiskLevel          CreditRiskLevel   `json:"risk_level"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	LastPaymentDate    time.Time         `json:"last_payment_date"`
	NextPaymentDate    time.Time         `json:"next_payment_date"`
	PaymentHistory     []Payment         `json:"payment_history"`
	DrawdownHistory    []Drawdown        `json:"drawdown_history"`
	CollateralAssets   []CollateralAsset `json:"collateral_assets"`
	InterestAccrued    *big.Int          `json:"interest_accrued"`
	FeesAccrued        *big.Int          `json:"fees_accrued"`
	GracePeriodDays    int               `json:"grace_period_days"`
	LatePaymentPenalty *big.Float        `json:"late_payment_penalty"`
	DefaultThreshold   *big.Float        `json:"default_threshold"`
	AutoRenewal        bool              `json:"auto_renewal"`
	RenewalTerms       *RenewalTerms     `json:"renewal_terms"`
}

// Payment represents a payment made on a credit line
type Payment struct {
	ID          string      `json:"id"`
	Amount      *big.Int    `json:"amount"`
	Type        PaymentType `json:"type"`
	Timestamp   time.Time   `json:"timestamp"`
	Description string      `json:"description"`
}

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypePrincipal PaymentType = "principal"
	PaymentTypeInterest  PaymentType = "interest"
	PaymentTypeFees      PaymentType = "fees"
	PaymentTypeLate      PaymentType = "late"
)

// Drawdown represents a drawdown from a credit line
type Drawdown struct {
	ID           string     `json:"id"`
	Amount       *big.Int   `json:"amount"`
	Timestamp    time.Time  `json:"timestamp"`
	Purpose      string     `json:"purpose"`
	InterestRate *big.Float `json:"interest_rate"`
}

// CollateralAsset represents collateral pledged for a credit line
type CollateralAsset struct {
	AssetID          string     `json:"asset_id"`
	Amount           *big.Int   `json:"amount"`
	Value            *big.Int   `json:"value"`
	CollateralRatio  *big.Float `json:"collateral_ratio"`
	LiquidationPrice *big.Int   `json:"liquidation_price"`
	PledgedAt        time.Time  `json:"pledged_at"`
}

// CreditRiskLevel represents the risk level of a credit line
type CreditRiskLevel string

const (
	CreditRiskLevelLow      CreditRiskLevel = "low"
	CreditRiskLevelMedium   CreditRiskLevel = "medium"
	CreditRiskLevelHigh     CreditRiskLevel = "high"
	CreditRiskLevelVeryHigh CreditRiskLevel = "very_high"
)

// RenewalTerms represents the terms for credit line renewal
type RenewalTerms struct {
	AutoRenewalEnabled bool       `json:"auto_renewal_enabled"`
	RenewalDate        time.Time  `json:"renewal_date"`
	NewCreditLimit     *big.Int   `json:"new_credit_limit"`
	NewInterestRate    *big.Float `json:"new_interest_rate"`
	RenewalFee         *big.Float `json:"renewal_fee"`
}

// CreditLineManager manages credit lines
type CreditLineManager struct {
	creditLines map[string]*CreditLine
	mu          sync.RWMutex
	logger      *logger.Logger
	riskManager RiskManager
}

// NewCreditLineManager creates a new credit line manager
func NewCreditLineManager(riskManager RiskManager) *CreditLineManager {
	return &CreditLineManager{
		creditLines: make(map[string]*CreditLine),
		logger: logger.NewLogger(&logger.Config{
			Level:   logger.INFO,
			Prefix:  "credit_line_manager",
			UseJSON: false,
		}),
		riskManager: riskManager,
	}
}

// CreateCreditLine creates a new credit line
func (clm *CreditLineManager) CreateCreditLine(ctx context.Context, userID string, creditLimit *big.Int, interestRate *big.Float, collateralRatio *big.Float) (*CreditLine, error) {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	// Validate inputs
	if creditLimit.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("credit limit must be positive")
	}
	if interestRate.Cmp(big.NewFloat(0)) < 0 {
		return nil, errors.New("interest rate cannot be negative")
	}
	if collateralRatio.Cmp(big.NewFloat(0)) <= 0 {
		return nil, errors.New("collateral ratio must be positive")
	}

	// Generate credit line ID
	creditLineID := fmt.Sprintf("cl_%s_%d", userID, time.Now().UnixNano())

	// Create credit line
	creditLine := &CreditLine{
		ID:                 creditLineID,
		UserID:             userID,
		Type:               CreditLineTypeRevolving,
		Status:             CreditLineStatusActive,
		CreditLimit:        new(big.Int).Set(creditLimit),
		AvailableCredit:    new(big.Int).Set(creditLimit),
		OutstandingBalance: big.NewInt(0),
		InterestRate:       new(big.Float).Copy(interestRate),
		OriginationFee:     big.NewFloat(0.01),  // 1% origination fee
		AnnualFee:          big.NewFloat(0.005), // 0.5% annual fee
		CollateralRatio:    new(big.Float).Copy(collateralRatio),
		MinCollateralRatio: new(big.Float).Mul(collateralRatio, big.NewFloat(0.8)), // 80% of required ratio
		UtilizationRate:    big.NewFloat(0),
		CreditScore:        700, // Default credit score
		RiskLevel:          CreditRiskLevelMedium,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastPaymentDate:    time.Now(),
		NextPaymentDate:    time.Now().AddDate(0, 1, 0), // Next month
		PaymentHistory:     make([]Payment, 0),
		DrawdownHistory:    make([]Drawdown, 0),
		CollateralAssets:   make([]CollateralAsset, 0),
		InterestAccrued:    big.NewInt(0),
		FeesAccrued:        big.NewInt(0),
		GracePeriodDays:    15,
		LatePaymentPenalty: big.NewFloat(0.05), // 5% late payment penalty
		DefaultThreshold:   big.NewFloat(0.9),  // 90% of credit limit
		AutoRenewal:        true,
		RenewalTerms: &RenewalTerms{
			AutoRenewalEnabled: true,
			RenewalDate:        time.Now().AddDate(1, 0, 0), // 1 year from now
			NewCreditLimit:     new(big.Int).Set(creditLimit),
			NewInterestRate:    new(big.Float).Copy(interestRate),
			RenewalFee:         big.NewFloat(0.0025), // 0.25% renewal fee
		},
	}

	clm.creditLines[creditLineID] = creditLine
	clm.logger.Info("Credit line created - id: %s, user: %s, limit: %s", creditLineID, userID, creditLimit.String())

	return creditLine, nil
}

// GetCreditLine retrieves a credit line by ID
func (clm *CreditLineManager) GetCreditLine(creditLineID string) (*CreditLine, error) {
	clm.mu.RLock()
	defer clm.mu.RUnlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return nil, fmt.Errorf("credit line %s not found", creditLineID)
	}

	return creditLine, nil
}

// GetUserCreditLines retrieves all credit lines for a user
func (clm *CreditLineManager) GetUserCreditLines(userID string) ([]*CreditLine, error) {
	clm.mu.RLock()
	defer clm.mu.RUnlock()

	var userCreditLines []*CreditLine
	for _, creditLine := range clm.creditLines {
		if creditLine.UserID == userID {
			userCreditLines = append(userCreditLines, creditLine)
		}
	}

	return userCreditLines, nil
}

// Drawdown draws funds from a credit line
func (clm *CreditLineManager) Drawdown(ctx context.Context, creditLineID string, amount *big.Int, purpose string) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if creditLine.Status != CreditLineStatusActive {
		return fmt.Errorf("credit line is not active (status: %s)", creditLine.Status)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("drawdown amount must be positive")
	}

	if amount.Cmp(creditLine.AvailableCredit) > 0 {
		return fmt.Errorf("insufficient available credit: requested %s, available %s", amount.String(), creditLine.AvailableCredit.String())
	}

	// Check collateral ratio
	if err := clm.validateCollateralRatio(creditLine); err != nil {
		return fmt.Errorf("collateral ratio validation failed: %w", err)
	}

	// Process drawdown
	creditLine.OutstandingBalance.Add(creditLine.OutstandingBalance, amount)
	creditLine.AvailableCredit.Sub(creditLine.AvailableCredit, amount)
	creditLine.UtilizationRate = new(big.Float).Quo(
		new(big.Float).SetInt(creditLine.OutstandingBalance),
		new(big.Float).SetInt(creditLine.CreditLimit),
	)

	// Record drawdown
	drawdown := Drawdown{
		ID:           fmt.Sprintf("dd_%s_%d", creditLineID, time.Now().UnixNano()),
		Amount:       new(big.Int).Set(amount),
		Timestamp:    time.Now(),
		Purpose:      purpose,
		InterestRate: new(big.Float).Copy(creditLine.InterestRate),
	}
	creditLine.DrawdownHistory = append(creditLine.DrawdownHistory, drawdown)

	creditLine.UpdatedAt = time.Now()
	clm.logger.Info("Drawdown processed - credit_line: %s, amount: %s, purpose: %s", creditLineID, amount.String(), purpose)

	return nil
}

// MakePayment makes a payment on a credit line
func (clm *CreditLineManager) MakePayment(ctx context.Context, creditLineID string, amount *big.Int, paymentType PaymentType, description string) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("payment amount must be positive")
	}

	// Record payment
	payment := Payment{
		ID:          fmt.Sprintf("pmt_%s_%d", creditLineID, time.Now().UnixNano()),
		Amount:      new(big.Int).Set(amount),
		Type:        paymentType,
		Timestamp:   time.Now(),
		Description: description,
	}
	creditLine.PaymentHistory = append(creditLine.PaymentHistory, payment)

	// Apply payment based on type
	switch paymentType {
	case PaymentTypePrincipal:
		if amount.Cmp(creditLine.OutstandingBalance) > 0 {
			// Overpayment - reduce outstanding balance to zero and increase available credit
			overpayment := new(big.Int).Sub(amount, creditLine.OutstandingBalance)
			creditLine.OutstandingBalance = big.NewInt(0)
			creditLine.AvailableCredit.Add(creditLine.AvailableCredit, overpayment)
		} else {
			// Regular principal payment
			creditLine.OutstandingBalance.Sub(creditLine.OutstandingBalance, amount)
			creditLine.AvailableCredit.Add(creditLine.AvailableCredit, amount)
		}
	case PaymentTypeInterest:
		creditLine.InterestAccrued.Sub(creditLine.InterestAccrued, amount)
	case PaymentTypeFees:
		creditLine.FeesAccrued.Sub(creditLine.FeesAccrued, amount)
	}

	// Update utilization rate
	if creditLine.CreditLimit.Cmp(big.NewInt(0)) > 0 {
		creditLine.UtilizationRate = new(big.Float).Quo(
			new(big.Float).SetInt(creditLine.OutstandingBalance),
			new(big.Float).SetInt(creditLine.CreditLimit),
		)
	}

	creditLine.LastPaymentDate = time.Now()
	creditLine.NextPaymentDate = time.Now().AddDate(0, 1, 0)
	creditLine.UpdatedAt = time.Now()

	clm.logger.Info("Payment processed - credit_line: %s, amount: %s, type: %s", creditLineID, amount.String(), paymentType)

	return nil
}

// AddCollateral adds collateral to a credit line
func (clm *CreditLineManager) AddCollateral(ctx context.Context, creditLineID string, assetID string, amount *big.Int, value *big.Int) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("collateral amount must be positive")
	}

	if value.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("collateral value must be positive")
	}

	// Calculate collateral ratio
	collateralRatio := new(big.Float).Quo(
		new(big.Float).SetInt(value),
		new(big.Float).SetInt(creditLine.OutstandingBalance),
	)

	// Add collateral asset
	collateralAsset := CollateralAsset{
		AssetID:          assetID,
		Amount:           new(big.Int).Set(amount),
		Value:            new(big.Int).Set(value),
		CollateralRatio:  collateralRatio,
		LiquidationPrice: big.NewInt(0), // Would be calculated based on market conditions
		PledgedAt:        time.Now(),
	}

	creditLine.CollateralAssets = append(creditLine.CollateralAssets, collateralAsset)

	// Update overall collateral ratio
	clm.updateCollateralRatio(creditLine)
	creditLine.UpdatedAt = time.Now()

	clm.logger.Info("Collateral added - credit_line: %s, asset: %s, amount: %s, value: %s",
		creditLineID, assetID, amount.String(), value.String())

	return nil
}

// RemoveCollateral removes collateral from a credit line
func (clm *CreditLineManager) RemoveCollateral(ctx context.Context, creditLineID string, assetID string, amount *big.Int) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	// Find and update collateral asset
	for i := range creditLine.CollateralAssets {
		asset := &creditLine.CollateralAssets[i]
		if asset.AssetID == assetID {
			if amount.Cmp(asset.Amount) > 0 {
				return fmt.Errorf("insufficient collateral: requested %s, available %s", amount.String(), asset.Amount.String())
			}

			// Check if removal would violate minimum collateral ratio
			originalAmount := new(big.Int).Set(asset.Amount)
			newAmount := new(big.Int).Sub(asset.Amount, amount)
			if newAmount.Cmp(big.NewInt(0)) == 0 {
				// Remove entire asset
				creditLine.CollateralAssets = append(creditLine.CollateralAssets[:i], creditLine.CollateralAssets[i+1:]...)
			} else {
				// Update amount
				asset.Amount = newAmount
				// Recalculate value proportionally
				ratio := new(big.Float).Quo(new(big.Float).SetInt(newAmount), new(big.Float).SetInt(originalAmount))
				newValue := new(big.Float).Mul(new(big.Float).SetInt(asset.Value), ratio)
				newValueInt, _ := newValue.Int(nil)
				asset.Value = newValueInt
			}

			// Update overall collateral ratio
			clm.updateCollateralRatio(creditLine)
			creditLine.UpdatedAt = time.Now()

			clm.logger.Info("Collateral removed - credit_line: %s, asset: %s, amount: %s",
				creditLineID, assetID, amount.String())

			return nil
		}
	}

	return fmt.Errorf("collateral asset %s not found", assetID)
}

// UpdateCreditLimit updates the credit limit of a credit line
func (clm *CreditLineManager) UpdateCreditLimit(ctx context.Context, creditLineID string, newLimit *big.Int) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if newLimit.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("credit limit must be positive")
	}

	if newLimit.Cmp(creditLine.OutstandingBalance) < 0 {
		return fmt.Errorf("new credit limit %s cannot be less than outstanding balance %s",
			newLimit.String(), creditLine.OutstandingBalance.String())
	}

	// Calculate difference
	limitDiff := new(big.Int).Sub(newLimit, creditLine.CreditLimit)

	// Update credit limit
	creditLine.CreditLimit = new(big.Int).Set(newLimit)

	// Update available credit
	creditLine.AvailableCredit.Add(creditLine.AvailableCredit, limitDiff)

	// Update utilization rate
	clm.updateUtilizationRate(creditLine)

	creditLine.UpdatedAt = time.Now()

	clm.logger.Info("Credit limit updated - credit_line: %s, old_limit: %s, new_limit: %s",
		creditLineID, creditLine.CreditLimit.String(), newLimit.String())

	return nil
}

// UpdateInterestRate updates the interest rate of a credit line
func (clm *CreditLineManager) UpdateInterestRate(ctx context.Context, creditLineID string, newRate *big.Float) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if newRate.Cmp(big.NewFloat(0)) < 0 {
		return errors.New("interest rate cannot be negative")
	}

	creditLine.InterestRate = new(big.Float).Copy(newRate)
	creditLine.UpdatedAt = time.Now()

	clm.logger.Info("Interest rate updated - credit_line: %s, new_rate: %v", creditLineID, newRate)

	return nil
}

// CloseCreditLine closes a credit line
func (clm *CreditLineManager) CloseCreditLine(ctx context.Context, creditLineID string) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	creditLine, exists := clm.creditLines[creditLineID]
	if !exists {
		return fmt.Errorf("credit line %s not found", creditLineID)
	}

	if creditLine.OutstandingBalance.Cmp(big.NewInt(0)) > 0 {
		return fmt.Errorf("cannot close credit line with outstanding balance: %s", creditLine.OutstandingBalance.String())
	}

	creditLine.Status = CreditLineStatusClosed
	creditLine.UpdatedAt = time.Now()

	clm.logger.Info("Credit line closed - id: %s", creditLineID)

	return nil
}

// validateCollateralRatio validates that the credit line meets minimum collateral requirements
func (clm *CreditLineManager) validateCollateralRatio(creditLine *CreditLine) error {
	if creditLine.OutstandingBalance.Cmp(big.NewInt(0)) == 0 {
		return nil // No outstanding balance, no collateral needed
	}

	if creditLine.CollateralRatio.Cmp(creditLine.MinCollateralRatio) < 0 {
		return fmt.Errorf("insufficient collateral ratio: current %v, minimum %v",
			creditLine.CollateralRatio, creditLine.MinCollateralRatio)
	}

	return nil
}

// updateCollateralRatio updates the overall collateral ratio for a credit line
func (clm *CreditLineManager) updateCollateralRatio(creditLine *CreditLine) {
	if creditLine.OutstandingBalance.Cmp(big.NewInt(0)) == 0 {
		creditLine.CollateralRatio = big.NewFloat(0)
		return
	}

	totalCollateralValue := big.NewInt(0)
	for _, asset := range creditLine.CollateralAssets {
		totalCollateralValue.Add(totalCollateralValue, asset.Value)
	}

	creditLine.CollateralRatio = new(big.Float).Quo(
		new(big.Float).SetInt(totalCollateralValue),
		new(big.Float).SetInt(creditLine.OutstandingBalance),
	)
}

// updateUtilizationRate updates the utilization rate for a credit line
func (clm *CreditLineManager) updateUtilizationRate(creditLine *CreditLine) {
	if creditLine.CreditLimit.Cmp(big.NewInt(0)) == 0 {
		creditLine.UtilizationRate = big.NewFloat(0)
		return
	}

	creditLine.UtilizationRate = new(big.Float).Quo(
		new(big.Float).SetInt(creditLine.OutstandingBalance),
		new(big.Float).SetInt(creditLine.CreditLimit),
	)
}

// GetCreditLineStats returns statistics for a credit line
func (clm *CreditLineManager) GetCreditLineStats(creditLineID string) (map[string]interface{}, error) {
	creditLine, err := clm.GetCreditLine(creditLineID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"id":                     creditLine.ID,
		"user_id":                creditLine.UserID,
		"status":                 creditLine.Status,
		"credit_limit":           creditLine.CreditLimit.String(),
		"available_credit":       creditLine.AvailableCredit.String(),
		"outstanding_balance":    creditLine.OutstandingBalance.String(),
		"interest_rate":          creditLine.InterestRate.String(),
		"collateral_ratio":       creditLine.CollateralRatio.String(),
		"utilization_rate":       creditLine.UtilizationRate.String(),
		"credit_score":           creditLine.CreditScore,
		"risk_level":             creditLine.RiskLevel,
		"created_at":             creditLine.CreatedAt,
		"last_payment_date":      creditLine.LastPaymentDate,
		"next_payment_date":      creditLine.NextPaymentDate,
		"payment_count":          len(creditLine.PaymentHistory),
		"drawdown_count":         len(creditLine.DrawdownHistory),
		"collateral_asset_count": len(creditLine.CollateralAssets),
		"interest_accrued":       creditLine.InterestAccrued.String(),
		"fees_accrued":           creditLine.FeesAccrued.String(),
	}

	return stats, nil
}

// RiskManager interface for credit risk assessment
type RiskManager interface {
	AssessCreditRisk(userID string, creditLimit *big.Int, collateralRatio *big.Float) (CreditRiskLevel, error)
	CalculateInterestRate(riskLevel CreditRiskLevel, marketRate *big.Float) *big.Float
	ValidateCollateral(assets []CollateralAsset, outstandingBalance *big.Int) error
}
