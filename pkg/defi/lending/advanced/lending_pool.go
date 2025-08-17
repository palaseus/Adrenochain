package advanced

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// LendingPool represents a lending pool for a specific asset
type LendingPool struct {
	ID                    string    `json:"id"`
	Asset                 string    `json:"asset"`
	TotalSupply           *big.Int  `json:"total_supply"`
	TotalBorrowed         *big.Int  `json:"total_borrowed"`
	AvailableLiquidity    *big.Int  `json:"available_liquidity"`
	UtilizationRate       *big.Int  `json:"utilization_rate"`
	SupplyRate            *big.Int  `json:"supply_rate"`
	BorrowRate            *big.Int  `json:"borrow_rate"`
	BaseRate              *big.Int  `json:"base_rate"`
	KinkRate              *big.Int  `json:"kink_rate"`
	Multiplier            *big.Int  `json:"multiplier"`
	JumpMultiplier        *big.Int  `json:"jump_multiplier"`
	ReserveFactor         *big.Int  `json:"reserve_factor"`
	CollateralFactor      *big.Int  `json:"collateral_factor"`
	LiquidationThreshold  *big.Int  `json:"liquidation_threshold"`
	LiquidationPenalty    *big.Int  `json:"liquidation_penalty"`
	MaxBorrowAmount       *big.Int  `json:"max_borrow_amount"`
	MinBorrowAmount       *big.Int  `json:"min_borrow_amount"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	LastInterestUpdate    time.Time `json:"last_interest_update"`
	IsActive              bool      `json:"is_active"`
	
	// Internal state
	mutex                 sync.RWMutex
	accounts              map[string]*Account
	interestAccrualIndex  *big.Int
	borrowIndex           *big.Int
	supplyIndex           *big.Int
}

// Account represents a user's position in the lending pool
type Account struct {
	UserID                string    `json:"user_id"`
	SupplyBalance         *big.Int  `json:"supply_balance"`
	BorrowBalance         *big.Int  `json:"borrow_balance"`
	SupplyIndex           *big.Int  `json:"supply_index"`
	BorrowIndex           *big.Int  `json:"borrow_index"`
	CollateralValue       *big.Int  `json:"collateral_value"`
	BorrowValue           *big.Int  `json:"borrow_value"`
	HealthFactor          *big.Int  `json:"health_factor"`
	LastUpdate            time.Time `json:"last_update"`
	IsLiquidatable        bool      `json:"is_liquidatable"`
}

// LendingPoolError represents lending pool specific errors
type LendingPoolError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	PoolID    string `json:"pool_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

func (e LendingPoolError) Error() string {
	return e.Operation + ": " + e.Message
}

// Lending pool errors
var (
	ErrPoolNotFound           = errors.New("lending pool not found")
	ErrPoolInactive           = errors.New("lending pool is inactive")
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")
	ErrInsufficientCollateral = errors.New("insufficient collateral")
	ErrBorrowLimitExceeded   = errors.New("borrow limit exceeded")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrAccountNotFound       = errors.New("account not found")
	ErrHealthFactorTooLow    = errors.New("health factor too low")
	ErrLiquidationNotAllowed = errors.New("liquidation not allowed")
)

// NewLendingPool creates a new lending pool
func NewLendingPool(
	asset string,
	baseRate, kinkRate, multiplier, jumpMultiplier, reserveFactor, collateralFactor, liquidationThreshold, liquidationPenalty *big.Int,
) *LendingPool {
	now := time.Now()
	
	return &LendingPool{
		ID:                   "pool_" + asset,
		Asset:                asset,
		TotalSupply:          big.NewInt(0),
		TotalBorrowed:        big.NewInt(0),
		AvailableLiquidity:   big.NewInt(0),
		UtilizationRate:      big.NewInt(0),
		SupplyRate:           big.NewInt(0),
		BorrowRate:           baseRate,
		BaseRate:             baseRate,
		KinkRate:             kinkRate,
		Multiplier:           multiplier,
		JumpMultiplier:       jumpMultiplier,
		ReserveFactor:        reserveFactor,
		CollateralFactor:     collateralFactor,
		LiquidationThreshold: liquidationThreshold,
		LiquidationPenalty:   liquidationPenalty,
		MaxBorrowAmount:      big.NewInt(0),
		MinBorrowAmount:      big.NewInt(1000), // Minimum 0.001 in smallest unit
		CreatedAt:            now,
		UpdatedAt:            now,
		LastInterestUpdate:   now,
		IsActive:             true,
		accounts:             make(map[string]*Account),
		interestAccrualIndex: big.NewInt(1000000000000000000), // 1e18 in wei
		borrowIndex:          big.NewInt(1000000000000000000),
		supplyIndex:          big.NewInt(1000000000000000000),
	}
}

// Supply adds assets to the lending pool
func (lp *LendingPool) Supply(userID string, amount *big.Int) error {
	if !lp.IsActive {
		return &LendingPoolError{Operation: "Supply", Message: ErrPoolInactive.Error(), PoolID: lp.ID}
	}
	
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return &LendingPoolError{Operation: "Supply", Message: ErrInvalidAmount.Error(), PoolID: lp.ID}
	}

	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	// Update interest rates before processing
	lp.updateInterestRates()

	// Get or create account
	account := lp.getOrCreateAccount(userID)

	// Calculate interest earned
	interestEarned := lp.calculateInterestEarned(account)
	if interestEarned.Cmp(big.NewInt(0)) > 0 {
		account.SupplyBalance.Add(account.SupplyBalance, interestEarned)
		lp.TotalSupply.Add(lp.TotalSupply, interestEarned)
	}

	// Update supply balance
	account.SupplyBalance.Add(account.SupplyBalance, amount)
	account.SupplyIndex.Set(lp.supplyIndex)
	account.LastUpdate = time.Now()

	// Update pool state
	lp.TotalSupply.Add(lp.TotalSupply, amount)
	lp.AvailableLiquidity.Add(lp.AvailableLiquidity, amount)
	lp.updateUtilizationRate()
	lp.UpdatedAt = time.Now()

	return nil
}

// Borrow borrows assets from the lending pool
func (lp *LendingPool) Borrow(userID string, amount *big.Int) error {
	if !lp.IsActive {
		return &LendingPoolError{Operation: "Borrow", Message: ErrPoolInactive.Error(), PoolID: lp.ID}
	}
	
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return &LendingPoolError{Operation: "Borrow", Message: ErrInvalidAmount.Error(), PoolID: lp.ID}
	}

	if amount.Cmp(lp.MinBorrowAmount) < 0 {
		return &LendingPoolError{Operation: "Borrow", Message: "amount below minimum borrow amount", PoolID: lp.ID}
	}

	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	// Update interest rates before processing
	lp.updateInterestRates()

	// Check liquidity
	if amount.Cmp(lp.AvailableLiquidity) > 0 {
		return &LendingPoolError{Operation: "Borrow", Message: ErrInsufficientLiquidity.Error(), PoolID: lp.ID}
	}

	// Get or create account
	account := lp.getOrCreateAccount(userID)

	// Calculate interest owed
	interestOwed := lp.calculateInterestOwed(account)
	if interestOwed.Cmp(big.NewInt(0)) > 0 {
		account.BorrowBalance.Add(account.BorrowBalance, interestOwed)
		lp.TotalBorrowed.Add(lp.TotalBorrowed, interestOwed)
	}

	// Check borrow limit
	if !lp.checkBorrowLimit(account, amount) {
		return &LendingPoolError{Operation: "Borrow", Message: ErrBorrowLimitExceeded.Error(), PoolID: lp.ID}
	}

	// Update borrow balance and value
	account.BorrowBalance.Add(account.BorrowBalance, amount)
	account.BorrowValue.Add(account.BorrowValue, amount)
	account.BorrowIndex.Set(lp.borrowIndex)
	account.LastUpdate = time.Now()

	// Update pool state
	lp.TotalBorrowed.Add(lp.TotalBorrowed, amount)
	lp.AvailableLiquidity.Sub(lp.AvailableLiquidity, amount)
	lp.updateUtilizationRate()
	lp.UpdatedAt = time.Now()

	// Update health factor
	lp.updateAccountHealthFactor(account)

	return nil
}

// Repay repays borrowed assets
func (lp *LendingPool) Repay(userID string, amount *big.Int) error {
	if !lp.IsActive {
		return &LendingPoolError{Operation: "Repay", Message: ErrPoolInactive.Error(), PoolID: lp.ID}
	}
	
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return &LendingPoolError{Operation: "Repay", Message: ErrInvalidAmount.Error(), PoolID: lp.ID}
	}

	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	// Update interest rates before processing
	lp.updateInterestRates()

	account, exists := lp.accounts[userID]
	if !exists {
		return &LendingPoolError{Operation: "Repay", Message: ErrAccountNotFound.Error(), PoolID: lp.ID, UserID: userID}
	}

	// Calculate interest owed
	interestOwed := lp.calculateInterestOwed(account)
	if interestOwed.Cmp(big.NewInt(0)) > 0 {
		account.BorrowBalance.Add(account.BorrowBalance, interestOwed)
		lp.TotalBorrowed.Add(lp.TotalBorrowed, interestOwed)
	}

	// Determine repayment amount
	repayAmount := new(big.Int).Set(amount)
	if repayAmount.Cmp(account.BorrowBalance) > 0 {
		repayAmount.Set(account.BorrowBalance)
	}

	// Update borrow balance and value
	account.BorrowBalance.Sub(account.BorrowBalance, repayAmount)
	account.BorrowValue.Sub(account.BorrowValue, repayAmount)
	account.BorrowIndex.Set(lp.borrowIndex)
	account.LastUpdate = time.Now()

	// Update pool state
	lp.TotalBorrowed.Sub(lp.TotalBorrowed, repayAmount)
	lp.AvailableLiquidity.Add(lp.AvailableLiquidity, repayAmount)
	lp.updateUtilizationRate()
	lp.UpdatedAt = time.Now()

	// Update health factor
	lp.updateAccountHealthFactor(account)

	return nil
}

// Withdraw withdraws supplied assets
func (lp *LendingPool) Withdraw(userID string, amount *big.Int) error {
	if !lp.IsActive {
		return &LendingPoolError{Operation: "Withdraw", Message: ErrPoolInactive.Error(), PoolID: lp.ID}
	}
	
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return &LendingPoolError{Operation: "Withdraw", Message: ErrInvalidAmount.Error(), PoolID: lp.ID}
	}

	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	// Update interest rates before processing
	lp.updateInterestRates()

	account, exists := lp.accounts[userID]
	if !exists {
		return &LendingPoolError{Operation: "Withdraw", Message: ErrAccountNotFound.Error(), PoolID: lp.ID, UserID: userID}
	}

	// Calculate interest earned
	interestEarned := lp.calculateInterestEarned(account)
	if interestEarned.Cmp(big.NewInt(0)) > 0 {
		account.SupplyBalance.Add(account.SupplyBalance, interestEarned)
		lp.TotalSupply.Add(lp.TotalSupply, interestEarned)
	}

	// Check withdrawal amount
	if amount.Cmp(account.SupplyBalance) > 0 {
		return &LendingPoolError{Operation: "Withdraw", Message: "insufficient supply balance", PoolID: lp.ID, UserID: userID}
	}

	// Check if withdrawal would make account unhealthy
	if !lp.checkBorrowLimit(account, big.NewInt(0)) {
		return &LendingPoolError{Operation: "Withdraw", Message: "withdrawal would make account unhealthy", PoolID: lp.ID, UserID: userID}
	}

	// Update supply balance
	account.SupplyBalance.Sub(account.SupplyBalance, amount)
	account.SupplyIndex.Set(lp.supplyIndex)
	account.LastUpdate = time.Now()

	// Update pool state
	lp.TotalSupply.Sub(lp.TotalSupply, amount)
	lp.AvailableLiquidity.Sub(lp.AvailableLiquidity, amount)
	lp.updateUtilizationRate()
	lp.UpdatedAt = time.Now()

	return nil
}

// updateInterestRates updates the interest rates based on utilization
func (lp *LendingPool) updateInterestRates() {
	now := time.Now()
	timeElapsed := now.Sub(lp.LastInterestUpdate)
	
	if timeElapsed < time.Minute {
		return // Update at most once per minute
	}

	// Calculate new borrow rate based on utilization
	utilization := lp.UtilizationRate
	borrowRate := new(big.Int)

	if utilization.Cmp(lp.KinkRate) <= 0 {
		// Below kink: base + (utilization * multiplier)
		utilizationMultiplier := new(big.Int).Mul(utilization, lp.Multiplier)
		borrowRate.Add(lp.BaseRate, utilizationMultiplier)
	} else {
		// Above kink: base + (kink * multiplier) + ((utilization - kink) * jumpMultiplier)
		kinkMultiplier := new(big.Int).Mul(lp.KinkRate, lp.Multiplier)
		excessUtilization := new(big.Int).Sub(utilization, lp.KinkRate)
		excessMultiplier := new(big.Int).Mul(excessUtilization, lp.JumpMultiplier)
		
		borrowRate.Add(lp.BaseRate, kinkMultiplier)
		borrowRate.Add(borrowRate, excessMultiplier)
	}

	lp.BorrowRate = borrowRate

	// Calculate supply rate: borrow rate * utilization * (1 - reserve factor)
	// Note: utilization is in basis points (10000 = 100%), so we need to convert to decimal
	utilizationDecimal := new(big.Int).Div(utilization, big.NewInt(10000)) // Convert to decimal
	if utilizationDecimal.Cmp(big.NewInt(0)) == 0 {
		// If utilization is very low, set supply rate to 0
		lp.SupplyRate = big.NewInt(0)
	} else {
		utilizationBorrowRate := new(big.Int).Mul(borrowRate, utilizationDecimal)
		reserveFactorAdjustment := new(big.Int).Sub(big.NewInt(10000), lp.ReserveFactor) // 10000 = 100%
		supplyRate := new(big.Int).Mul(utilizationBorrowRate, reserveFactorAdjustment)
		supplyRate.Div(supplyRate, big.NewInt(10000))
		lp.SupplyRate = supplyRate
	}
	lp.LastInterestUpdate = now
}

// updateUtilizationRate updates the utilization rate
func (lp *LendingPool) updateUtilizationRate() {
	if lp.TotalSupply.Cmp(big.NewInt(0)) == 0 {
		lp.UtilizationRate = big.NewInt(0)
		return
	}
	
	// Utilization = total borrowed / total supply
	utilization := new(big.Int).Mul(lp.TotalBorrowed, big.NewInt(10000)) // 10000 = 100%
	utilization.Div(utilization, lp.TotalSupply)
	lp.UtilizationRate = utilization
}

// getOrCreateAccount gets an existing account or creates a new one
func (lp *LendingPool) getOrCreateAccount(userID string) *Account {
	if account, exists := lp.accounts[userID]; exists {
		return account
	}
	
	account := &Account{
		UserID:        userID,
		SupplyBalance: big.NewInt(0),
		BorrowBalance: big.NewInt(0),
		SupplyIndex:   lp.supplyIndex,
		BorrowIndex:   lp.borrowIndex,
		CollateralValue: big.NewInt(0),
		BorrowValue:   big.NewInt(0),
		HealthFactor:  big.NewInt(0),
		LastUpdate:    time.Now(),
		IsLiquidatable: false,
	}
	
	lp.accounts[userID] = account
	return account
}

// calculateInterestEarned calculates interest earned on supply
func (lp *LendingPool) calculateInterestEarned(account *Account) *big.Int {
	if account.SupplyBalance.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Interest = supply balance * (current index - user index) / user index
	indexDelta := new(big.Int).Sub(lp.supplyIndex, account.SupplyIndex)
	interest := new(big.Int).Mul(account.SupplyBalance, indexDelta)
	interest.Div(interest, account.SupplyIndex)
	
	return interest
}

// calculateInterestOwed calculates interest owed on borrows
func (lp *LendingPool) calculateInterestOwed(account *Account) *big.Int {
	if account.BorrowBalance.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Interest = borrow balance * (current index - user index) / user index
	indexDelta := new(big.Int).Sub(lp.borrowIndex, account.BorrowIndex)
	interest := new(big.Int).Mul(account.BorrowBalance, indexDelta)
	interest.Div(interest, account.BorrowIndex)
	
	return interest
}

// checkBorrowLimit checks if an account can borrow the specified amount
func (lp *LendingPool) checkBorrowLimit(account *Account, borrowAmount *big.Int) bool {
	// Calculate total borrow value after this borrow
	totalBorrowValue := new(big.Int).Add(account.BorrowValue, borrowAmount)
	
	// Calculate maximum borrow value based on collateral
	maxBorrowValue := new(big.Int).Mul(account.CollateralValue, lp.CollateralFactor)
	maxBorrowValue.Div(maxBorrowValue, big.NewInt(10000)) // 10000 = 100%
	
	return totalBorrowValue.Cmp(maxBorrowValue) <= 0
}

// updateAccountHealthFactor updates the health factor for an account
func (lp *LendingPool) updateAccountHealthFactor(account *Account) {
	if account.BorrowValue.Cmp(big.NewInt(0)) == 0 {
		account.HealthFactor = big.NewInt(10000) // 100% healthy
		account.IsLiquidatable = false
		return
	}
	
	// Health factor = (collateral value * liquidation threshold) / borrow value
	healthFactor := new(big.Int).Mul(account.CollateralValue, lp.LiquidationThreshold)
	healthFactor.Div(healthFactor, account.BorrowValue)
	healthFactor.Div(healthFactor, big.NewInt(100)) // Convert to percentage
	
	account.HealthFactor = healthFactor
	account.IsLiquidatable = healthFactor.Cmp(big.NewInt(100)) < 0 // Below 100% = liquidatable
}

// GetAccount returns an account by user ID
func (lp *LendingPool) GetAccount(userID string) (*Account, error) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	account, exists := lp.accounts[userID]
	if !exists {
		return nil, &LendingPoolError{Operation: "GetAccount", Message: ErrAccountNotFound.Error(), PoolID: lp.ID, UserID: userID}
	}
	
	return account, nil
}

// GetPoolStats returns pool statistics
func (lp *LendingPool) GetPoolStats() map[string]interface{} {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	
	return map[string]interface{}{
		"id":                 lp.ID,
		"asset":              lp.Asset,
		"total_supply":       lp.TotalSupply.String(),
		"total_borrowed":     lp.TotalBorrowed.String(),
		"available_liquidity": lp.AvailableLiquidity.String(),
		"utilization_rate":   lp.UtilizationRate.String(),
		"supply_rate":        lp.SupplyRate.String(),
		"borrow_rate":        lp.BorrowRate.String(),
		"base_rate":          lp.BaseRate.String(),
		"is_active":          lp.IsActive,
		"account_count":      len(lp.accounts),
	}
}
