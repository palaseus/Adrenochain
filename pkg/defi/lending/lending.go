package lending

import (
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// LendingProtocol represents a DeFi lending protocol
type LendingProtocol struct {
	mu sync.RWMutex

	// Protocol information
	ProtocolID   string
	Name         string
	Symbol       string
	Decimals     uint8
	Owner        engine.Address
	Paused       bool
	
	// Assets
	Assets map[engine.Address]*Asset
	
	// Users
	Users map[engine.Address]*User
	
	// Interest rate model
	InterestRateModel InterestRateModel
	
	// Liquidation settings
	LiquidationThreshold *big.Int // Collateral ratio threshold for liquidation
	LiquidationBonus     *big.Int // Bonus for liquidators (basis points)
	
	// Events
	SupplyEvents     []SupplyEvent
	BorrowEvents     []BorrowEvent
	RepayEvents      []RepayEvent
	WithdrawEvents   []WithdrawEvent
	LiquidateEvents  []LiquidateEvent
	InterestEvents   []InterestEvent
	
	// Statistics
	TotalSupply     *big.Int
	TotalBorrow     *big.Int
	TotalReserves   *big.Int
	LastUpdate      time.Time
	SupplyCount     uint64
	BorrowCount     uint64
}

// Asset represents a lending asset
type Asset struct {
	Token           engine.Address
	Symbol          string
	Decimals        uint8
	TotalSupply     *big.Int
	TotalBorrow     *big.Int
	Reserves        *big.Int
	BorrowRate      *big.Int // Annual rate in basis points
	SupplyRate      *big.Int // Annual rate in basis points
	CollateralRatio *big.Int // Collateral ratio in basis points
	MaxLTV          *big.Int // Maximum loan-to-value ratio
	LiquidationThreshold *big.Int
	Paused          bool
}

// User represents a lending protocol user
type User struct {
	Address     engine.Address
	Assets      map[engine.Address]*UserAsset
	Collateral  map[engine.Address]*big.Int
	Borrows     map[engine.Address]*big.Int
	LastUpdate  time.Time
}

// UserAsset represents a user's position in an asset
type UserAsset struct {
	Token           engine.Address
	Balance         *big.Int
	BorrowBalance   *big.Int
	CollateralValue *big.Int
	BorrowValue     *big.Int
	LastUpdate      time.Time
}

// InterestRateModel defines interest rate calculation
type InterestRateModel interface {
	CalculateBorrowRate(utilizationRate *big.Int) *big.Int
	CalculateSupplyRate(utilizationRate *big.Int, borrowRate *big.Int) *big.Int
}

// DefaultInterestRateModel implements a simple interest rate model
type DefaultInterestRateModel struct {
	BaseRate      *big.Int // Base rate in basis points
	Multiplier    *big.Int // Multiplier for utilization
	JumpMultiplier *big.Int // Jump multiplier for high utilization
	Kink          *big.Int // Utilization rate at which jump occurs
}

// NewDefaultInterestRateModel creates a new interest rate model
func NewDefaultInterestRateModel(baseRate, multiplier, jumpMultiplier, kink *big.Int) *DefaultInterestRateModel {
	return &DefaultInterestRateModel{
		BaseRate:      new(big.Int).Set(baseRate),
		Multiplier:    new(big.Int).Set(multiplier),
		JumpMultiplier: new(big.Int).Set(jumpMultiplier),
		Kink:          new(big.Int).Set(kink),
	}
}

// CalculateBorrowRate calculates the borrow rate based on utilization
func (irm *DefaultInterestRateModel) CalculateBorrowRate(utilizationRate *big.Int) *big.Int {
	if utilizationRate.Cmp(irm.Kink) <= 0 {
		// Below kink: baseRate + (utilizationRate * multiplier)
		rate := new(big.Int).Mul(utilizationRate, irm.Multiplier)
		rate = new(big.Int).Div(rate, big.NewInt(10000)) // Convert from basis points
		return new(big.Int).Add(irm.BaseRate, rate)
	} else {
		// Above kink: baseRate + (kink * multiplier) + ((utilizationRate - kink) * jumpMultiplier)
		normalRate := new(big.Int).Mul(irm.Kink, irm.Multiplier)
		normalRate = new(big.Int).Div(normalRate, big.NewInt(10000))
		
		excessUtilization := new(big.Int).Sub(utilizationRate, irm.Kink)
		jumpRate := new(big.Int).Mul(excessUtilization, irm.JumpMultiplier)
		jumpRate = new(big.Int).Div(jumpRate, big.NewInt(10000))
		
		totalRate := new(big.Int).Add(irm.BaseRate, normalRate)
		return new(big.Int).Add(totalRate, jumpRate)
	}
}

// CalculateSupplyRate calculates the supply rate based on utilization and borrow rate
func (irm *DefaultInterestRateModel) CalculateSupplyRate(utilizationRate *big.Int, borrowRate *big.Int) *big.Int {
	// Supply rate = borrow rate * utilization rate
	rate := new(big.Int).Mul(borrowRate, utilizationRate)
	return new(big.Int).Div(rate, big.NewInt(10000))
}

// NewLendingProtocol creates a new lending protocol
func NewLendingProtocol(
	protocolID, name, symbol string,
	decimals uint8,
	owner engine.Address,
	liquidationThreshold, liquidationBonus *big.Int,
) *LendingProtocol {
	return &LendingProtocol{
		ProtocolID:          protocolID,
		Name:                name,
		Symbol:              symbol,
		Decimals:            decimals,
		Owner:               owner,
		Paused:              false,
		Assets:              make(map[engine.Address]*Asset),
		Users:               make(map[engine.Address]*User),
		InterestRateModel:   NewDefaultInterestRateModel(
			big.NewInt(200),   // 2% base rate
			big.NewInt(1000),  // 10% multiplier
			big.NewInt(2000),  // 20% jump multiplier
			big.NewInt(8000),  // 80% kink
		),
		LiquidationThreshold: new(big.Int).Set(liquidationThreshold),
		LiquidationBonus:     new(big.Int).Set(liquidationBonus),
		SupplyEvents:         make([]SupplyEvent, 0),
		BorrowEvents:         make([]BorrowEvent, 0),
		RepayEvents:          make([]RepayEvent, 0),
		WithdrawEvents:       make([]WithdrawEvent, 0),
		LiquidateEvents:      make([]LiquidateEvent, 0),
		InterestEvents:       make([]InterestEvent, 0),
		TotalSupply:          big.NewInt(0),
		TotalBorrow:          big.NewInt(0),
		TotalReserves:        big.NewInt(0),
		LastUpdate:           time.Now(),
		SupplyCount:          0,
		BorrowCount:          0,
	}
}

// SupplyEvent represents a supply event
type SupplyEvent struct {
	User        engine.Address
	Asset       engine.Address
	Amount      *big.Int
	Balance     *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// BorrowEvent represents a borrow event
type BorrowEvent struct {
	User        engine.Address
	Asset       engine.Address
	Amount      *big.Int
	Balance     *big.Int
	BorrowRate  *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// RepayEvent represents a repay event
type RepayEvent struct {
	User        engine.Address
	Asset       engine.Address
	Amount      *big.Int
	Balance     *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// WithdrawEvent represents a withdraw event
type WithdrawEvent struct {
	User        engine.Address
	Asset       engine.Address
	Amount      *big.Int
	Balance     *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// LiquidateEvent represents a liquidation event
type LiquidateEvent struct {
	Liquidator  engine.Address
	Borrower    engine.Address
	Asset       engine.Address
	Amount      *big.Int
	Bonus       *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// InterestEvent represents an interest accrual event
type InterestEvent struct {
	Asset       engine.Address
	BorrowRate  *big.Int
	SupplyRate  *big.Int
	Utilization *big.Int
	Timestamp   time.Time
	BlockNumber uint64
}

// AddAsset adds a new asset to the lending protocol
func (lp *LendingProtocol) AddAsset(
	token engine.Address,
	symbol string,
	decimals uint8,
	collateralRatio, maxLTV *big.Int,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	if _, exists := lp.Assets[token]; exists {
		return ErrAssetAlreadyExists
	}
	
	asset := &Asset{
		Token:                token,
		Symbol:               symbol,
		Decimals:             decimals,
		TotalSupply:          big.NewInt(0),
		TotalBorrow:          big.NewInt(0),
		Reserves:             big.NewInt(0),
		BorrowRate:           big.NewInt(0),
		SupplyRate:           big.NewInt(0),
		CollateralRatio:      new(big.Int).Set(collateralRatio),
		MaxLTV:               new(big.Int).Set(maxLTV),
		LiquidationThreshold: new(big.Int).Set(lp.LiquidationThreshold),
		Paused:               false,
	}
	
	lp.Assets[token] = asset
	return nil
}

// Supply supplies assets to the lending protocol
func (lp *LendingProtocol) Supply(
	user engine.Address,
	asset engine.Address,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Check if protocol is paused
	if lp.Paused {
		return ErrProtocolPaused
	}
	
	// Validate input
	if err := lp.validateSupplyInput(asset, amount); err != nil {
		return err
	}
	
	// Get or create user
	if lp.Users[user] == nil {
		lp.Users[user] = &User{
			Address:    user,
			Assets:     make(map[engine.Address]*UserAsset),
			Collateral: make(map[engine.Address]*big.Int),
			Borrows:    make(map[engine.Address]*big.Int),
			LastUpdate: time.Now(),
		}
	}
	
	// Get or create user asset
	if lp.Users[user].Assets[asset] == nil {
		lp.Users[user].Assets[asset] = &UserAsset{
			Token:           asset,
			Balance:         big.NewInt(0),
			BorrowBalance:   big.NewInt(0),
			CollateralValue: big.NewInt(0),
			BorrowValue:     big.NewInt(0),
			LastUpdate:      time.Now(),
		}
	}
	
	// Update balances
	lp.Assets[asset].TotalSupply = new(big.Int).Add(lp.Assets[asset].TotalSupply, amount)
	lp.Users[user].Assets[asset].Balance = new(big.Int).Add(lp.Users[user].Assets[asset].Balance, amount)
	
	// Update protocol totals
	lp.TotalSupply = new(big.Int).Add(lp.TotalSupply, amount)
	lp.SupplyCount++
	
	// Update interest rates
	lp.updateInterestRates(asset)
	
	// Record event
	event := SupplyEvent{
		User:        user,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		Balance:     new(big.Int).Set(lp.Users[user].Assets[asset].Balance),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	lp.SupplyEvents = append(lp.SupplyEvents, event)
	
	return nil
}

// Withdraw withdraws assets from the lending protocol
func (lp *LendingProtocol) Withdraw(
	user engine.Address,
	asset engine.Address,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Check if protocol is paused
	if lp.Paused {
		return ErrProtocolPaused
	}
	
	// Validate input
	if err := lp.validateWithdrawInput(user, asset, amount); err != nil {
		return err
	}
	
	// Check if user has sufficient balance
	userAsset := lp.Users[user].Assets[asset]
	if userAsset.Balance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}
	
	// Update balances
	lp.Assets[asset].TotalSupply = new(big.Int).Sub(lp.Assets[asset].TotalSupply, amount)
	userAsset.Balance = new(big.Int).Sub(userAsset.Balance, amount)
	
	// Update protocol totals
	lp.TotalSupply = new(big.Int).Sub(lp.TotalSupply, amount)
	
	// Update interest rates
	lp.updateInterestRates(asset)
	
	// Record event
	event := WithdrawEvent{
		User:        user,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		Balance:     new(big.Int).Set(userAsset.Balance),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	lp.WithdrawEvents = append(lp.WithdrawEvents, event)
	
	return nil
}

// Borrow borrows assets from the lending protocol
func (lp *LendingProtocol) Borrow(
	user engine.Address,
	asset engine.Address,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Check if protocol is paused
	if lp.Paused {
		return ErrProtocolPaused
	}
	
	// Validate input
	if err := lp.validateBorrowInput(asset, amount); err != nil {
		return err
	}
	
	// Check if user has sufficient collateral
	if err := lp.checkCollateralRatio(user, asset, amount); err != nil {
		return err
	}
	
	// Get or create user
	if lp.Users[user] == nil {
		lp.Users[user] = &User{
			Address:    user,
			Assets:     make(map[engine.Address]*UserAsset),
			Collateral: make(map[engine.Address]*big.Int),
			Borrows:    make(map[engine.Address]*big.Int),
			LastUpdate: time.Now(),
		}
	}
	
	// Get or create user asset
	if lp.Users[user].Assets[asset] == nil {
		lp.Users[user].Assets[asset] = &UserAsset{
			Token:           asset,
			Balance:         big.NewInt(0),
			BorrowBalance:   big.NewInt(0),
			CollateralValue: big.NewInt(0),
			BorrowValue:     big.NewInt(0),
			LastUpdate:      time.Now(),
		}
	}
	
	// Update balances
	lp.Assets[asset].TotalBorrow = new(big.Int).Add(lp.Assets[asset].TotalBorrow, amount)
	lp.Users[user].Assets[asset].BorrowBalance = new(big.Int).Add(lp.Users[user].Assets[asset].BorrowBalance, amount)
	
	// Update protocol totals
	lp.TotalBorrow = new(big.Int).Add(lp.TotalBorrow, amount)
	lp.BorrowCount++
	
	// Update interest rates
	lp.updateInterestRates(asset)
	
	// Record event
	event := BorrowEvent{
		User:        user,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		Balance:     new(big.Int).Set(lp.Users[user].Assets[asset].BorrowBalance),
		BorrowRate:  new(big.Int).Set(lp.Assets[asset].BorrowRate),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	lp.BorrowEvents = append(lp.BorrowEvents, event)
	
	return nil
}

// Repay repays borrowed assets
func (lp *LendingProtocol) Repay(
	user engine.Address,
	asset engine.Address,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Check if protocol is paused
	if lp.Paused {
		return ErrProtocolPaused
	}
	
	// Validate input
	if err := lp.validateRepayInput(user, asset, amount); err != nil {
		return err
	}
	
	// Check if user has sufficient borrow balance
	userAsset := lp.Users[user].Assets[asset]
	if userAsset.BorrowBalance.Cmp(amount) < 0 {
		return ErrInsufficientBorrowBalance
	}
	
	// Update balances
	lp.Assets[asset].TotalBorrow = new(big.Int).Sub(lp.Assets[asset].TotalBorrow, amount)
	userAsset.BorrowBalance = new(big.Int).Sub(userAsset.BorrowBalance, amount)
	
	// Update protocol totals
	lp.TotalBorrow = new(big.Int).Sub(lp.TotalBorrow, amount)
	
	// Update interest rates
	lp.updateInterestRates(asset)
	
	// Record event
	event := RepayEvent{
		User:        user,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		Balance:     new(big.Int).Set(userAsset.BorrowBalance),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	lp.RepayEvents = append(lp.RepayEvents, event)
	
	return nil
}

// Liquidate liquidates a user's position
func (lp *LendingProtocol) Liquidate(
	liquidator engine.Address,
	borrower engine.Address,
	asset engine.Address,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Check if protocol is paused
	if lp.Paused {
		return ErrProtocolPaused
	}
	
	// Validate input
	if err := lp.validateLiquidateInput(liquidator, borrower, asset, amount); err != nil {
		return err
	}
	
	// Check if liquidation is necessary
	if err := lp.checkLiquidationEligibility(borrower, asset); err != nil {
		return err
	}
	
	// Calculate liquidation bonus
	bonus := lp.calculateLiquidationBonus(amount)
	totalAmount := new(big.Int).Add(amount, bonus)
	
	// Check if liquidator has sufficient collateral
	liquidatorAsset := lp.Users[liquidator].Assets[asset]
	if liquidatorAsset == nil || liquidatorAsset.Balance.Cmp(totalAmount) < 0 {
		return ErrInsufficientLiquidationCollateral
	}
	
	// Update borrower balances
	borrowerAsset := lp.Users[borrower].Assets[asset]
	borrowerAsset.BorrowBalance = new(big.Int).Sub(borrowerAsset.BorrowBalance, amount)
	
	// Update liquidator balances
	liquidatorAsset.Balance = new(big.Int).Sub(liquidatorAsset.Balance, totalAmount)
	
	// Update asset totals
	lp.Assets[asset].TotalBorrow = new(big.Int).Sub(lp.Assets[asset].TotalBorrow, amount)
	lp.Assets[asset].TotalSupply = new(big.Int).Sub(lp.Assets[asset].TotalSupply, bonus)
	
	// Update protocol totals
	lp.TotalBorrow = new(big.Int).Sub(lp.TotalBorrow, amount)
	lp.TotalSupply = new(big.Int).Sub(lp.TotalSupply, bonus)
	
	// Record event
	event := LiquidateEvent{
		Liquidator:  liquidator,
		Borrower:    borrower,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		Bonus:       new(big.Int).Set(bonus),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	lp.LiquidateEvents = append(lp.LiquidateEvents, event)
	
	return nil
}

// GetUserInfo returns user information
func (lp *LendingProtocol) GetUserInfo(user engine.Address) *User {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	if userInfo, exists := lp.Users[user]; exists {
		// Return a copy to avoid race conditions
		userCopy := &User{
			Address:    userInfo.Address,
			Assets:     make(map[engine.Address]*UserAsset),
			Collateral: make(map[engine.Address]*big.Int),
			Borrows:    make(map[engine.Address]*big.Int),
			LastUpdate: userInfo.LastUpdate,
		}
		
		for token, asset := range userInfo.Assets {
			userCopy.Assets[token] = &UserAsset{
				Token:           asset.Token,
				Balance:         new(big.Int).Set(asset.Balance),
				BorrowBalance:   new(big.Int).Set(asset.BorrowBalance),
				CollateralValue: new(big.Int).Set(asset.CollateralValue),
				BorrowValue:     new(big.Int).Set(asset.BorrowValue),
				LastUpdate:      asset.LastUpdate,
			}
		}
		
		for token, collateral := range userInfo.Collateral {
			userCopy.Collateral[token] = new(big.Int).Set(collateral)
		}
		
		for token, borrow := range userInfo.Borrows {
			userCopy.Borrows[token] = new(big.Int).Set(borrow)
		}
		
		return userCopy
	}
	
	return nil
}

// GetAssetInfo returns asset information
func (lp *LendingProtocol) GetAssetInfo(asset engine.Address) *Asset {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	if assetInfo, exists := lp.Assets[asset]; exists {
		// Return a copy to avoid race conditions
		assetCopy := &Asset{
			Token:                assetInfo.Token,
			Symbol:               assetInfo.Symbol,
			Decimals:             assetInfo.Decimals,
			TotalSupply:          new(big.Int).Set(assetInfo.TotalSupply),
			TotalBorrow:          new(big.Int).Set(assetInfo.TotalBorrow),
			Reserves:             new(big.Int).Set(assetInfo.Reserves),
			BorrowRate:           new(big.Int).Set(assetInfo.BorrowRate),
			SupplyRate:           new(big.Int).Set(assetInfo.SupplyRate),
			CollateralRatio:      new(big.Int).Set(assetInfo.CollateralRatio),
			MaxLTV:               new(big.Int).Set(assetInfo.MaxLTV),
			LiquidationThreshold: new(big.Int).Set(assetInfo.LiquidationThreshold),
			Paused:               assetInfo.Paused,
		}
		
		return assetCopy
	}
	
	return nil
}

// GetProtocolStats returns protocol statistics
func (lp *LendingProtocol) GetProtocolStats() (uint64, uint64, *big.Int, *big.Int, *big.Int) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	return lp.SupplyCount,
		   lp.BorrowCount,
		   new(big.Int).Set(lp.TotalSupply),
		   new(big.Int).Set(lp.TotalBorrow),
		   new(big.Int).Set(lp.TotalReserves)
}

// Pause pauses the protocol
func (lp *LendingProtocol) Pause() error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	if lp.Paused {
		return ErrProtocolAlreadyPaused
	}
	
	lp.Paused = true
	return nil
}

// Unpause resumes the protocol
func (lp *LendingProtocol) Unpause() error {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	if !lp.Paused {
		return ErrProtocolNotPaused
	}
	
	lp.Paused = false
	return nil
}

// updateInterestRates updates interest rates for an asset
func (lp *LendingProtocol) updateInterestRates(asset engine.Address) {
	assetInfo := lp.Assets[asset]
	if assetInfo == nil {
		return
	}
	
	// Calculate utilization rate
	utilizationRate := lp.calculateUtilizationRate(assetInfo)
	
	// Calculate new rates
	borrowRate := lp.InterestRateModel.CalculateBorrowRate(utilizationRate)
	supplyRate := lp.InterestRateModel.CalculateSupplyRate(utilizationRate, borrowRate)
	
	// Update asset rates
	assetInfo.BorrowRate = borrowRate
	assetInfo.SupplyRate = supplyRate
	
	// Record interest event
	event := InterestEvent{
		Asset:       asset,
		BorrowRate:  new(big.Int).Set(borrowRate),
		SupplyRate:  new(big.Int).Set(supplyRate),
		Utilization: new(big.Int).Set(utilizationRate),
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
	}
	lp.InterestEvents = append(lp.InterestEvents, event)
}

// calculateUtilizationRate calculates the utilization rate for an asset
func (lp *LendingProtocol) calculateUtilizationRate(asset *Asset) *big.Int {
	if asset.TotalSupply.Sign() == 0 {
		return big.NewInt(0)
	}
	
	// Utilization = TotalBorrow / TotalSupply
	utilization := new(big.Int).Mul(asset.TotalBorrow, big.NewInt(10000))
	return new(big.Int).Div(utilization, asset.TotalSupply)
}

// calculateLiquidationBonus calculates the liquidation bonus
func (lp *LendingProtocol) calculateLiquidationBonus(amount *big.Int) *big.Int {
	bonus := new(big.Int).Mul(amount, lp.LiquidationBonus)
	return new(big.Int).Div(bonus, big.NewInt(10000))
}

// checkCollateralRatio checks if user has sufficient collateral
func (lp *LendingProtocol) checkCollateralRatio(user engine.Address, asset engine.Address, amount *big.Int) error {
	userInfo := lp.Users[user]
	if userInfo == nil {
		return ErrInsufficientCollateral
	}
	
	// Calculate total collateral value
	totalCollateralValue := big.NewInt(0)
	for token, collateral := range userInfo.Collateral {
		if assetInfo, exists := lp.Assets[token]; exists {
			collateralValue := new(big.Int).Mul(collateral, assetInfo.CollateralRatio)
			collateralValue = new(big.Int).Div(collateralValue, big.NewInt(10000))
			totalCollateralValue = new(big.Int).Add(totalCollateralValue, collateralValue)
		}
	}
	
	// Calculate total borrow value
	totalBorrowValue := big.NewInt(0)
	for token, borrow := range userInfo.Borrows {
		if assetInfo, exists := lp.Assets[token]; exists {
			borrowValue := new(big.Int).Mul(borrow, big.NewInt(10000))
			borrowValue = new(big.Int).Div(borrowValue, assetInfo.MaxLTV)
			totalBorrowValue = new(big.Int).Add(totalBorrowValue, borrowValue)
		}
	}
	
	// Add new borrow amount
	newBorrowValue := new(big.Int).Mul(amount, big.NewInt(10000))
	if assetInfo, exists := lp.Assets[asset]; exists {
		newBorrowValue = new(big.Int).Div(newBorrowValue, assetInfo.MaxLTV)
	}
	totalBorrowValue = new(big.Int).Add(totalBorrowValue, newBorrowValue)
	
	// Check if collateral is sufficient
	if totalCollateralValue.Cmp(totalBorrowValue) < 0 {
		return ErrInsufficientCollateral
	}
	
	return nil
}

// checkLiquidationEligibility checks if a user's position can be liquidated
func (lp *LendingProtocol) checkLiquidationEligibility(borrower engine.Address, asset engine.Address) error {
	userInfo := lp.Users[borrower]
	if userInfo == nil {
		return ErrUserNotFound
	}
	
	// Calculate health factor
	healthFactor := lp.calculateHealthFactor(borrower)
	if healthFactor.Cmp(big.NewInt(10000)) >= 0 {
		return ErrNotEligibleForLiquidation
	}
	
	return nil
}

// calculateHealthFactor calculates the health factor for a user
func (lp *LendingProtocol) calculateHealthFactor(user engine.Address) *big.Int {
	userInfo := lp.Users[user]
	if userInfo == nil {
		return big.NewInt(0)
	}
	
	// Calculate total collateral value
	totalCollateralValue := big.NewInt(0)
	for token, collateral := range userInfo.Collateral {
		if assetInfo, exists := lp.Assets[token]; exists {
			collateralValue := new(big.Int).Mul(collateral, assetInfo.CollateralRatio)
			collateralValue = new(big.Int).Div(collateralValue, big.NewInt(10000))
			totalCollateralValue = new(big.Int).Add(totalCollateralValue, collateralValue)
		}
	}
	
	// Calculate total borrow value
	totalBorrowValue := big.NewInt(0)
	for token, borrow := range userInfo.Borrows {
		if assetInfo, exists := lp.Assets[token]; exists {
			borrowValue := new(big.Int).Mul(borrow, big.NewInt(10000))
			borrowValue = new(big.Int).Div(borrowValue, assetInfo.MaxLTV)
			totalBorrowValue = new(big.Int).Add(totalBorrowValue, borrowValue)
		}
	}
	
	if totalBorrowValue.Sign() == 0 {
		return big.NewInt(10000) // 100% health factor
	}
	
	// Health factor = (total collateral value / total borrow value) * 10000
	healthFactor := new(big.Int).Mul(totalCollateralValue, big.NewInt(10000))
	return new(big.Int).Div(healthFactor, totalBorrowValue)
}

// Validation functions
func (lp *LendingProtocol) validateSupplyInput(asset engine.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if assetInfo, exists := lp.Assets[asset]; !exists || assetInfo.Paused {
		return ErrAssetNotSupported
	}
	
	return nil
}

func (lp *LendingProtocol) validateWithdrawInput(user engine.Address, asset engine.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if assetInfo, exists := lp.Assets[asset]; !exists || assetInfo.Paused {
		return ErrAssetNotSupported
	}
	
	if lp.Users[user] == nil || lp.Users[user].Assets[asset] == nil {
		return ErrUserNotFound
	}
	
	return nil
}

func (lp *LendingProtocol) validateBorrowInput(asset engine.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if assetInfo, exists := lp.Assets[asset]; !exists || assetInfo.Paused {
		return ErrAssetNotSupported
	}
	
	return nil
}

func (lp *LendingProtocol) validateRepayInput(user engine.Address, asset engine.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if assetInfo, exists := lp.Assets[asset]; !exists || assetInfo.Paused {
		return ErrAssetNotSupported
	}
	
	if lp.Users[user] == nil || lp.Users[user].Assets[asset] == nil {
		return ErrUserNotFound
	}
	
	return nil
}

func (lp *LendingProtocol) validateLiquidateInput(liquidator engine.Address, borrower engine.Address, asset engine.Address, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if assetInfo, exists := lp.Assets[asset]; !exists || assetInfo.Paused {
		return ErrAssetNotSupported
	}
	
	if liquidator == borrower {
		return ErrCannotLiquidateSelf
	}
	
	return nil
}
