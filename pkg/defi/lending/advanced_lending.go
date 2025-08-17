package lending

import (
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/logger"
)

// RiskLevel represents the risk level of a lending pool
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// LendingPool represents an advanced lending pool with multiple assets
type LendingPool struct {
	ID              string
	Name            string
	Assets          map[string]*LendingAsset
	TotalLiquidity  float64
	TotalBorrowed   float64
	UtilizationRate float64
	APY             float64
	RiskLevel       RiskLevel
	CreatedAt       time.Time
	UpdatedAt       time.Time
	mu              sync.RWMutex
	logger          *logger.Logger
}

// LendingAsset represents a single asset in the lending pool
type LendingAsset struct {
	Symbol           string
	Address          string
	TotalSupply      float64
	TotalBorrowed    float64
	AvailableLiquidity float64
	SupplyAPY        float64
	BorrowAPY        float64
	CollateralFactor float64
	LiquidationThreshold float64
	ReserveFactor    float64
	LastUpdate       time.Time
}

// Loan represents a user's loan position
type Loan struct {
	ID              string
	UserID          string
	AssetSymbol     string
	Type            LoanType
	Amount          float64
	BorrowedAmount  float64
	RepaidAmount    float64
	InterestAccrued float64
	APY             float64
	StartDate       time.Time
	DueDate         time.Time
	Status          LoanStatus
	Collateral      []Collateral
	LastPayment     time.Time
	mu              sync.RWMutex
}

// LoanType represents the type of loan
type LoanType string

const (
	LoanTypeCollateralized LoanType = "collateralized"
	LoanTypeUncollateralized LoanType = "uncollateralized"
	LoanTypeFlash         LoanType = "flash"
	LoanTypeStable        LoanType = "stable"
)

// LoanStatus represents the status of a loan
type LoanStatus string

const (
	LoanStatusActive     LoanStatus = "active"
	LoanStatusRepaid     LoanStatus = "repaid"
	LoanStatusDefaulted  LoanStatus = "defaulted"
	LoanStatusLiquidated LoanStatus = "liquidated"
)

// Collateral represents collateral pledged for a loan
type Collateral struct {
	AssetSymbol string
	Amount      float64
	Value       float64
	LTV         float64 // Loan-to-Value ratio
	PledgedAt   time.Time
}

// LendingUser represents a lending platform user
type LendingUser struct {
	ID           string
	TotalBorrowed float64
	TotalSupplied float64
	CreditScore  float64
	RiskLevel    RiskLevel
	Loans        []string
	Collateral   map[string]float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// LendingService provides advanced lending and borrowing functionality
type LendingService struct {
	pools    map[string]*LendingPool
	loans    map[string]*Loan
	users    map[string]*LendingUser
	logger   *logger.Logger
	mu       sync.RWMutex
}

// NewLendingService creates a new lending service
func NewLendingService() *LendingService {
	return &LendingService{
		pools:   make(map[string]*LendingPool),
		loans:   make(map[string]*Loan),
		users:   make(map[string]*LendingUser),
		logger:  logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "lending_service"}),
	}
}

// CreatePool creates a new lending pool
func (ls *LendingService) CreatePool(id, name string, assets []string) (*LendingPool, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if _, exists := ls.pools[id]; exists {
		return nil, fmt.Errorf("pool with ID %s already exists", id)
	}

	pool := &LendingPool{
		ID:        id,
		Name:      name,
		Assets:    make(map[string]*LendingAsset),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		logger:    logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: fmt.Sprintf("lending_pool_%s", id)}),
	}

	// Initialize assets
	for _, asset := range assets {
		pool.Assets[asset] = &LendingAsset{
			Symbol:              asset,
			CollateralFactor:    0.8, // 80% LTV
			LiquidationThreshold: 0.85, // 85% liquidation threshold
			ReserveFactor:       0.1, // 10% reserve
			LastUpdate:          time.Now(),
		}
	}

	ls.pools[id] = pool
	ls.logger.Info("Created lending pool - id: %s, name: %s, assets: %v", id, name, assets)
	return pool, nil
}

// SupplyAsset supplies assets to a lending pool
func (ls *LendingService) SupplyAsset(poolID, assetSymbol, userID string, amount float64) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	pool, exists := ls.pools[poolID]
	if !exists {
		return fmt.Errorf("pool %s not found", poolID)
	}

	asset, exists := pool.Assets[assetSymbol]
	if !exists {
		return fmt.Errorf("asset %s not found in pool %s", assetSymbol, poolID)
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// Update asset supply
	asset.TotalSupply += amount
	asset.AvailableLiquidity += amount
	asset.LastUpdate = time.Now()

	// Update pool totals
	pool.TotalLiquidity += amount
	pool.UtilizationRate = pool.TotalBorrowed / pool.TotalLiquidity
	pool.UpdatedAt = time.Now()

	// Update user
	user := ls.getOrCreateUser(userID)
	user.TotalSupplied += amount
	if user.Collateral == nil {
		user.Collateral = make(map[string]float64)
	}
	user.Collateral[assetSymbol] += amount

	// Calculate new APY
	ls.updatePoolAPY(pool)

	ls.logger.Info("Asset supplied - pool: %s, asset: %s, user: %s, amount: %.2f", 
		poolID, assetSymbol, userID, amount)
	return nil
}

// BorrowAsset borrows assets from a lending pool
func (ls *LendingService) BorrowAsset(poolID, assetSymbol, userID string, amount float64, collateral []Collateral) (*Loan, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	pool, exists := ls.pools[poolID]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolID)
	}

	asset, exists := pool.Assets[assetSymbol]
	if !exists {
		return nil, fmt.Errorf("asset %s not found in pool %s", assetSymbol, poolID)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	if amount > asset.AvailableLiquidity {
		return nil, fmt.Errorf("insufficient liquidity: requested %.2f, available %.2f", 
			amount, asset.AvailableLiquidity)
	}

	// Validate collateral
	if err := ls.validateCollateral(collateral, amount, asset.CollateralFactor); err != nil {
		return nil, fmt.Errorf("collateral validation failed: %v", err)
	}

	// Create loan
	loan := &Loan{
		ID:             fmt.Sprintf("loan_%s_%s_%d", userID, assetSymbol, time.Now().Unix()),
		UserID:         userID,
		AssetSymbol:    assetSymbol,
		Type:           LoanTypeCollateralized,
		Amount:         amount,
		BorrowedAmount: amount,
		APY:            asset.BorrowAPY,
		StartDate:      time.Now(),
		DueDate:        time.Now().AddDate(0, 1, 0), // 1 month default
		Status:         LoanStatusActive,
		Collateral:     collateral,
		LastPayment:    time.Now(),
	}

	// Update asset borrowed amount
	asset.TotalBorrowed += amount
	asset.AvailableLiquidity -= amount
	asset.LastUpdate = time.Now()

	// Update pool totals
	pool.TotalBorrowed += amount
	pool.UtilizationRate = pool.TotalBorrowed / pool.TotalLiquidity
	pool.UpdatedAt = time.Now()

	// Update user
	user := ls.getOrCreateUser(userID)
	user.TotalBorrowed += amount
	user.Loans = append(user.Loans, loan.ID)

	// Store loan
	ls.loans[loan.ID] = loan

	// Calculate new APY
	ls.updatePoolAPY(pool)

	ls.logger.Info("Asset borrowed - pool: %s, asset: %s, user: %s, amount: %.2f, loan_id: %s", 
		poolID, assetSymbol, userID, amount, loan.ID)
	return loan, nil
}

// RepayLoan repays a portion or all of a loan
func (ls *LendingService) RepayLoan(loanID string, amount float64) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	loan, exists := ls.loans[loanID]
	if !exists {
		return fmt.Errorf("loan %s not found", loanID)
	}

	if loan.Status != LoanStatusActive {
		return fmt.Errorf("loan %s is not active", loanID)
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// Calculate interest
	interest := ls.calculateInterest(loan)
	totalOwed := loan.BorrowedAmount + interest



	if amount > totalOwed {
		amount = totalOwed
	}

	// Update loan
	loan.RepaidAmount += amount
	loan.LastPayment = time.Now()

	// Check if loan is fully repaid (with small tolerance for floating point precision)
	const tolerance = 0.01
	isFullyRepaid := loan.RepaidAmount >= (totalOwed - tolerance)
	
	if isFullyRepaid {
		loan.Status = LoanStatusRepaid
		loan.RepaidAmount = totalOwed
		loan.InterestAccrued = interest
		ls.logger.Info("Loan fully repaid - loan_id: %s, status: %s", loanID, loan.Status)
	}

	// Update pool
	pool := ls.findPoolByAsset(loan.AssetSymbol)
	if pool != nil {
		asset := pool.Assets[loan.AssetSymbol]
		asset.TotalBorrowed -= amount
		asset.AvailableLiquidity += amount
		asset.LastUpdate = time.Now()

		pool.TotalBorrowed -= amount
		pool.UtilizationRate = pool.TotalBorrowed / pool.TotalLiquidity
		pool.UpdatedAt = time.Now()

		ls.updatePoolAPY(pool)
	}

	// Update user
	user := ls.users[loan.UserID]
	if user != nil {
		user.TotalBorrowed -= amount
	}

	ls.logger.Info("Loan repaid - loan_id: %s, amount: %.2f, status: %s", 
		loanID, amount, loan.Status)
	return nil
}

// LiquidateLoan liquidates an underwater loan
func (ls *LendingService) LiquidateLoan(loanID string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	loan, exists := ls.loans[loanID]
	if !exists {
		return fmt.Errorf("loan %s not found", loanID)
	}

	if loan.Status != LoanStatusActive {
		return fmt.Errorf("loan %s is not active", loanID)
	}

	// Check if loan is underwater
	if !ls.isLoanUnderwater(loan) {
		return fmt.Errorf("loan %s is not underwater", loanID)
	}

	// Execute liquidation
	loan.Status = LoanStatusLiquidated

	// Return collateral to user (minus liquidation penalty)
	liquidationPenalty := 0.05 // 5% penalty
	for i := range loan.Collateral {
		loan.Collateral[i].Value *= (1 - liquidationPenalty)
	}

	// Update pool
	pool := ls.findPoolByAsset(loan.AssetSymbol)
	if pool != nil {
		asset := pool.Assets[loan.AssetSymbol]
		asset.TotalBorrowed -= loan.BorrowedAmount
		asset.AvailableLiquidity += loan.BorrowedAmount
		asset.LastUpdate = time.Now()

		pool.TotalBorrowed -= loan.BorrowedAmount
		pool.UtilizationRate = pool.TotalBorrowed / pool.TotalLiquidity
		pool.UpdatedAt = time.Now()

		ls.updatePoolAPY(pool)
	}

	// Update user
	user := ls.users[loan.UserID]
	if user != nil {
		user.TotalBorrowed -= loan.BorrowedAmount
	}

	ls.logger.Info("Loan liquidated - loan_id: %s, borrowed_amount: %.2f", 
		loanID, loan.BorrowedAmount)
	return nil
}

// GetPool returns a lending pool by ID
func (ls *LendingService) GetPool(poolID string) (*LendingPool, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	pool, exists := ls.pools[poolID]
	return pool, exists
}

// GetUser returns a user by ID
func (ls *LendingService) GetUser(userID string) (*LendingUser, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	user, exists := ls.users[userID]
	return user, exists
}

// GetLoan returns a loan by ID
func (ls *LendingService) GetLoan(loanID string) (*Loan, bool) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	loan, exists := ls.loans[loanID]
	return loan, exists
}

// GetPools returns all lending pools
func (ls *LendingService) GetPools() []*LendingPool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	pools := make([]*LendingPool, 0, len(ls.pools))
	for _, pool := range ls.pools {
		pools = append(pools, pool)
	}
	return pools
}

// GetUserLoans returns all loans for a user
func (ls *LendingService) GetUserLoans(userID string) []*Loan {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	var userLoans []*Loan
	for _, loan := range ls.loans {
		if loan.UserID == userID {
			userLoans = append(userLoans, loan)
		}
	}
	return userLoans
}

// Helper methods

func (ls *LendingService) getOrCreateUser(userID string) *LendingUser {
	if user, exists := ls.users[userID]; exists {
		return user
	}

	user := &LendingUser{
		ID:        userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	ls.users[userID] = user
	return user
}

func (ls *LendingService) validateCollateral(collateral []Collateral, borrowAmount float64, collateralFactor float64) error {
	totalCollateralValue := 0.0
	for _, col := range collateral {
		totalCollateralValue += col.Value
	}

	maxBorrowAmount := totalCollateralValue * collateralFactor
	if borrowAmount > maxBorrowAmount {
		return fmt.Errorf("borrow amount %.2f exceeds maximum allowed %.2f", 
			borrowAmount, maxBorrowAmount)
	}

	return nil
}

func (ls *LendingService) calculateInterest(loan *Loan) float64 {
	days := time.Since(loan.StartDate).Hours() / 24
	annualRate := loan.APY / 100
	dailyRate := annualRate / 365
	return loan.BorrowedAmount * dailyRate * days
}

func (ls *LendingService) isLoanUnderwater(loan *Loan) bool {
	totalCollateralValue := 0.0
	for _, col := range loan.Collateral {
		totalCollateralValue += col.Value
	}

	totalOwed := loan.BorrowedAmount + ls.calculateInterest(loan)
	return totalCollateralValue < totalOwed
}

func (ls *LendingService) findPoolByAsset(assetSymbol string) *LendingPool {
	for _, pool := range ls.pools {
		if _, exists := pool.Assets[assetSymbol]; exists {
			return pool
		}
	}
	return nil
}

func (ls *LendingService) updatePoolAPY(pool *LendingPool) {
	if pool.TotalLiquidity == 0 {
		pool.APY = 0
		return
	}

	// Simple APY calculation based on utilization
	baseRate := 0.05 // 5% base rate
	utilizationMultiplier := pool.UtilizationRate * 2
	pool.APY = baseRate + utilizationMultiplier

	// Update individual asset APYs
	for _, asset := range pool.Assets {
		if asset.TotalSupply > 0 {
			asset.SupplyAPY = pool.APY * 0.8 // 80% of pool APY
			asset.BorrowAPY = pool.APY * 1.2 // 120% of pool APY
		}
	}
}
