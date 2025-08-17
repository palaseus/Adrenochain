package advanced

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// FlashLoan represents a flash loan transaction
type FlashLoan struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Asset           string    `json:"asset"`
	Amount          *big.Int  `json:"amount"`
	Fee             *big.Int  `json:"fee"`
	FeeRate         *big.Int  `json:"fee_rate"`
	BorrowTime      time.Time `json:"borrow_time"`
	RepayTime       time.Time `json:"repay_time,omitempty"`
	Status          FlashLoanStatus `json:"status"`
	CallbackData    []byte    `json:"callback_data,omitempty"`
	GasUsed         *big.Int  `json:"gas_used,omitempty"`
	IsSuccessful    bool      `json:"is_successful"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// FlashLoanStatus represents the status of a flash loan
type FlashLoanStatus string

const (
	FlashLoanStatusPending   FlashLoanStatus = "pending"
	FlashLoanStatusActive    FlashLoanStatus = "active"
	FlashLoanStatusCompleted FlashLoanStatus = "completed"
	FlashLoanStatusFailed    FlashLoanStatus = "failed"
	FlashLoanStatusExpired   FlashLoanStatus = "expired"
)

// FlashLoanCallback represents the callback function for flash loan execution
type FlashLoanCallback func(flashLoan *FlashLoan) error

// FlashLoanError represents flash loan specific errors
type FlashLoanError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	LoanID    string `json:"loan_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

func (e FlashLoanError) Error() string {
	return e.Operation + ": " + e.Message
}

// Flash loan errors
var (
	ErrFlashLoanNotSupported = errors.New("flash loan not supported for this asset")
	ErrFlashLoanAmountTooHigh = errors.New("flash loan amount exceeds maximum allowed")
	ErrFlashLoanAmountTooLow = errors.New("flash loan amount below minimum required")
	ErrFlashLoanInsufficientLiquidity = errors.New("insufficient liquidity for flash loan")
	ErrFlashLoanCallbackFailed = errors.New("flash loan callback execution failed")
	ErrFlashLoanNotRepaid = errors.New("flash loan not repaid within deadline")
	ErrFlashLoanExpired = errors.New("flash loan has expired")
	ErrFlashLoanAlreadyActive = errors.New("flash loan is already active")
	ErrFlashLoanInvalidStatus = errors.New("invalid flash loan status for this operation")
)

// FlashLoanManager manages flash loan operations
type FlashLoanManager struct {
	lendingPool     *LendingPool
	maxFlashLoanAmount *big.Int
	minFlashLoanAmount *big.Int
	flashLoanFeeRate *big.Int
	maxFlashLoanDuration time.Duration
	activeLoans      map[string]*FlashLoan
	mutex            sync.RWMutex
	loanCounter      uint64
}

// NewFlashLoanManager creates a new flash loan manager
func NewFlashLoanManager(
	lendingPool *LendingPool,
	maxAmount, minAmount, feeRate *big.Int,
	maxDuration time.Duration,
) *FlashLoanManager {
	return &FlashLoanManager{
		lendingPool:         lendingPool,
		maxFlashLoanAmount:  maxAmount,
		minFlashLoanAmount:  minAmount,
		flashLoanFeeRate:    feeRate,
		maxFlashLoanDuration: maxDuration,
		activeLoans:         make(map[string]*FlashLoan),
		loanCounter:         0,
	}
}

// ExecuteFlashLoan executes a flash loan with the specified callback
func (flm *FlashLoanManager) ExecuteFlashLoan(
	userID, asset string,
	amount *big.Int,
	callback FlashLoanCallback,
	callbackData []byte,
) (*FlashLoan, error) {
	// Validate flash loan parameters
	if err := flm.validateFlashLoanRequest(userID, asset, amount); err != nil {
		return nil, err
	}

	// Check if lending pool supports flash loans
	if !flm.lendingPool.IsActive {
		return nil, &FlashLoanError{
			Operation: "ExecuteFlashLoan",
			Message:   ErrFlashLoanNotSupported.Error(),
			UserID:    userID,
		}
	}

	// Check liquidity
	if amount.Cmp(flm.lendingPool.AvailableLiquidity) > 0 {
		return nil, &FlashLoanError{
			Operation: "ExecuteFlashLoan",
			Message:   ErrFlashLoanInsufficientLiquidity.Error(),
			UserID:    userID,
		}
	}

	// Create flash loan
	flashLoan := flm.createFlashLoan(userID, asset, amount, callbackData)

	// Execute the flash loan
	if err := flm.executeFlashLoan(flashLoan, callback); err != nil {
		flashLoan.Status = FlashLoanStatusFailed
		flashLoan.ErrorMessage = err.Error()
		flashLoan.IsSuccessful = false
		return flashLoan, err
	}

	// Mark as completed
	flashLoan.Status = FlashLoanStatusCompleted
	flashLoan.IsSuccessful = true
	flashLoan.RepayTime = time.Now()

	return flashLoan, nil
}

// validateFlashLoanRequest validates flash loan request parameters
func (flm *FlashLoanManager) validateFlashLoanRequest(userID, asset string, amount *big.Int) error {
	if userID == "" {
		return &FlashLoanError{
			Operation: "ValidateFlashLoanRequest",
			Message:   "user ID cannot be empty",
			UserID:    userID,
		}
	}

	if asset == "" {
		return &FlashLoanError{
			Operation: "ValidateFlashLoanRequest",
			Message:   "asset cannot be empty",
			UserID:    userID,
		}
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return &FlashLoanError{
			Operation: "ValidateFlashLoanRequest",
			Message:   "amount must be positive",
			UserID:    userID,
		}
	}

	// Check minimum amount
	if amount.Cmp(flm.minFlashLoanAmount) < 0 {
		return &FlashLoanError{
			Operation: "ValidateFlashLoanRequest",
			Message:   ErrFlashLoanAmountTooLow.Error(),
			UserID:    userID,
		}
	}

	// Check maximum amount
	if amount.Cmp(flm.maxFlashLoanAmount) > 0 {
		return &FlashLoanError{
			Operation: "ValidateFlashLoanRequest",
			Message:   ErrFlashLoanAmountTooHigh.Error(),
			UserID:    userID,
		}
	}

	return nil
}

// createFlashLoan creates a new flash loan instance
func (flm *FlashLoanManager) createFlashLoan(userID, asset string, amount *big.Int, callbackData []byte) *FlashLoan {
	flm.mutex.Lock()
	defer flm.mutex.Unlock()

	flm.loanCounter++
	loanID := "flash_loan_" + string(rune(flm.loanCounter))

	// Calculate fee
	fee := flm.calculateFlashLoanFee(amount)

	flashLoan := &FlashLoan{
		ID:           loanID,
		UserID:       userID,
		Asset:        asset,
		Amount:       amount,
		Fee:          fee,
		FeeRate:      flm.flashLoanFeeRate,
		BorrowTime:   time.Now(),
		Status:       FlashLoanStatusPending,
		CallbackData: callbackData,
		IsSuccessful: false,
	}

	// Add to active loans
	flm.activeLoans[loanID] = flashLoan

	return flashLoan
}

// calculateFlashLoanFee calculates the fee for a flash loan
func (flm *FlashLoanManager) calculateFlashLoanFee(amount *big.Int) *big.Int {
	fee := new(big.Int).Mul(amount, flm.flashLoanFeeRate)
	fee.Div(fee, big.NewInt(10000)) // 10000 = 100%
	return fee
}

// executeFlashLoan executes the flash loan logic
func (flm *FlashLoanManager) executeFlashLoan(flashLoan *FlashLoan, callback FlashLoanCallback) error {
	// Mark as active
	flashLoan.Status = FlashLoanStatusActive

	// Temporarily reduce available liquidity
	flm.lendingPool.AvailableLiquidity.Sub(flm.lendingPool.AvailableLiquidity, flashLoan.Amount)

	// Execute callback
	if err := callback(flashLoan); err != nil {
		// Restore liquidity
		flm.lendingPool.AvailableLiquidity.Add(flm.lendingPool.AvailableLiquidity, flashLoan.Amount)
		return &FlashLoanError{
			Operation: "ExecuteFlashLoan",
			Message:   ErrFlashLoanCallbackFailed.Error(),
			LoanID:    flashLoan.ID,
		}
	}

	// Check if loan was repaid (this would typically be done by the callback)
	// For now, we'll assume the callback handles the repayment logic
	// In a real implementation, this would check the actual balance changes

	// Restore liquidity (assuming repayment)
	flm.lendingPool.AvailableLiquidity.Add(flm.lendingPool.AvailableLiquidity, flashLoan.Amount)

	return nil
}

// GetFlashLoan returns a flash loan by ID
func (flm *FlashLoanManager) GetFlashLoan(loanID string) (*FlashLoan, error) {
	flm.mutex.RLock()
	defer flm.mutex.RUnlock()

	flashLoan, exists := flm.activeLoans[loanID]
	if !exists {
		return nil, &FlashLoanError{
			Operation: "GetFlashLoan",
			Message:   "flash loan not found",
			LoanID:    loanID,
		}
	}

	return flashLoan, nil
}

// GetUserFlashLoans returns all flash loans for a user
func (flm *FlashLoanManager) GetUserFlashLoans(userID string) []*FlashLoan {
	flm.mutex.RLock()
	defer flm.mutex.RUnlock()

	var userLoans []*FlashLoan
	for _, loan := range flm.activeLoans {
		if loan.UserID == userID {
			userLoans = append(userLoans, loan)
		}
	}

	return userLoans
}

// GetActiveFlashLoans returns all active flash loans
func (flm *FlashLoanManager) GetActiveFlashLoans() []*FlashLoan {
	flm.mutex.RLock()
	defer flm.mutex.RUnlock()

	var activeLoans []*FlashLoan
	for _, loan := range flm.activeLoans {
		if loan.Status == FlashLoanStatusActive {
			activeLoans = append(activeLoans, loan)
		}
	}

	return activeLoans
}

// CancelFlashLoan cancels a pending flash loan
func (flm *FlashLoanManager) CancelFlashLoan(loanID, userID string) error {
	flm.mutex.Lock()
	defer flm.mutex.Unlock()

	flashLoan, exists := flm.activeLoans[loanID]
	if !exists {
		return &FlashLoanError{
			Operation: "CancelFlashLoan",
			Message:   "flash loan not found",
			LoanID:    loanID,
		}
	}

	// Check ownership
	if flashLoan.UserID != userID {
		return &FlashLoanError{
			Operation: "CancelFlashLoan",
			Message:   "user not authorized to cancel this flash loan",
			LoanID:    loanID,
			UserID:    userID,
		}
	}

	// Check status
	if flashLoan.Status != FlashLoanStatusPending {
		return &FlashLoanError{
			Operation: "CancelFlashLoan",
			Message:   ErrFlashLoanInvalidStatus.Error(),
			LoanID:    loanID,
		}
	}

	// Cancel the loan
	flashLoan.Status = FlashLoanStatusFailed
	flashLoan.ErrorMessage = "cancelled by user"
	flashLoan.IsSuccessful = false

	return nil
}

// CleanupExpiredLoans removes expired flash loans
func (flm *FlashLoanManager) CleanupExpiredLoans() {
	flm.mutex.Lock()
	defer flm.mutex.Unlock()

	now := time.Now()
	for _, loan := range flm.activeLoans {
		if loan.Status == FlashLoanStatusActive {
			timeElapsed := now.Sub(loan.BorrowTime)
			if timeElapsed > flm.maxFlashLoanDuration {
				loan.Status = FlashLoanStatusExpired
				loan.ErrorMessage = "flash loan expired"
				loan.IsSuccessful = false
				
				// Restore liquidity if loan was active
				if loan.Status == FlashLoanStatusActive {
					flm.lendingPool.AvailableLiquidity.Add(flm.lendingPool.AvailableLiquidity, loan.Amount)
				}
			}
		}
	}
}

// GetFlashLoanStats returns flash loan statistics
func (flm *FlashLoanManager) GetFlashLoanStats() map[string]interface{} {
	flm.mutex.RLock()
	defer flm.mutex.RUnlock()

	totalLoans := len(flm.activeLoans)
	activeLoans := 0
	completedLoans := 0
	failedLoans := 0
	totalVolume := big.NewInt(0)
	totalFees := big.NewInt(0)

	for _, loan := range flm.activeLoans {
		switch loan.Status {
		case FlashLoanStatusActive:
			activeLoans++
		case FlashLoanStatusCompleted:
			completedLoans++
			totalVolume.Add(totalVolume, loan.Amount)
			totalFees.Add(totalFees, loan.Fee)
		case FlashLoanStatusFailed:
			failedLoans++
		}
	}

	return map[string]interface{}{
		"total_loans":     totalLoans,
		"active_loans":    activeLoans,
		"completed_loans": completedLoans,
		"failed_loans":    failedLoans,
		"total_volume":    totalVolume.String(),
		"total_fees":      totalFees.String(),
		"fee_rate":        flm.flashLoanFeeRate.String(),
		"max_amount":      flm.maxFlashLoanAmount.String(),
		"min_amount":      flm.minFlashLoanAmount.String(),
		"max_duration":    flm.maxFlashLoanDuration.String(),
	}
}

// UpdateFlashLoanSettings updates flash loan manager settings
func (flm *FlashLoanManager) UpdateFlashLoanSettings(
	maxAmount, minAmount, feeRate *big.Int,
	maxDuration time.Duration,
) error {
	if maxAmount != nil && maxAmount.Cmp(big.NewInt(0)) > 0 {
		flm.maxFlashLoanAmount = maxAmount
	}

	if minAmount != nil && minAmount.Cmp(big.NewInt(0)) > 0 {
		flm.minFlashLoanAmount = minAmount
	}

	if feeRate != nil && feeRate.Cmp(big.NewInt(0)) >= 0 {
		flm.flashLoanFeeRate = feeRate
	}

	if maxDuration > 0 {
		flm.maxFlashLoanDuration = maxDuration
	}

	return nil
}
