package futures

import (
	"errors"
	"math/big"
	"time"
)

// StandardFuturesContract represents a standard futures contract with settlement
type StandardFuturesContract struct {
	Symbol          string
	UnderlyingAsset string
	ContractSize    *big.Float
	StrikePrice     *big.Float
	ExpirationDate  time.Time
	DeliveryDate    time.Time
	ContractType    ContractType
	SettlementType  SettlementType
	MarkPrice       *big.Float
	IndexPrice      *big.Float
	OpenInterest    *big.Float
	Volume24h       *big.Float
	High24h         *big.Float
	Low24h          *big.Float
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ContractType represents the type of futures contract
type ContractType int

const (
	Monthly ContractType = iota
	Quarterly
	Yearly
	Custom
)

// SettlementType represents how the futures contract is settled
type SettlementType int

const (
	PhysicalDelivery SettlementType = iota
	CashSettlement
	NetSettlement
)

// StandardFuturesPosition represents a position in standard futures
type StandardFuturesPosition struct {
	ID               string
	UserID           string
	Contract         *StandardFuturesContract
	Side             PositionSide
	Size             *big.Float
	EntryPrice       *big.Float
	MarkPrice        *big.Float
	UnrealizedPnL    *big.Float
	RealizedPnL      *big.Float
	Margin           *big.Float
	Leverage         *big.Float
	LiquidationPrice *big.Float
	IsOpen           bool
	EntryTime        time.Time
	UpdatedAt        time.Time
	CloseTime        *time.Time
}

// SettlementEvent represents a futures settlement event
type SettlementEvent struct {
	ID              string
	ContractID      string
	UserID          string
	PositionID      string
	SettlementType  SettlementType
	SettlementPrice *big.Float
	SettlementTime  time.Time
	Amount          *big.Float
	Status          SettlementStatus
	CreatedAt       time.Time
}

// SettlementStatus represents the status of settlement
type SettlementStatus int

const (
	SettlementPending SettlementStatus = iota
	SettlementProcessing
	SettlementCompleted
	SettlementFailed
)

// NewStandardFuturesContract creates a new standard futures contract
func NewStandardFuturesContract(
	symbol, underlyingAsset string,
	contractSize, strikePrice *big.Float,
	expirationDate, deliveryDate time.Time,
	contractType ContractType,
	settlementType SettlementType,
) (*StandardFuturesContract, error) {
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}
	if underlyingAsset == "" {
		return nil, errors.New("underlying asset cannot be empty")
	}
	if contractSize == nil || contractSize.Sign() <= 0 {
		return nil, errors.New("contract size must be positive")
	}
	if strikePrice == nil || strikePrice.Sign() <= 0 {
		return nil, errors.New("strike price must be positive")
	}
	if expirationDate.Before(time.Now()) {
		return nil, errors.New("expiration date cannot be in the past")
	}
	if deliveryDate.Before(expirationDate) {
		return nil, errors.New("delivery date cannot be before expiration date")
	}

	now := time.Now()

	return &StandardFuturesContract{
		Symbol:          symbol,
		UnderlyingAsset: underlyingAsset,
		ContractSize:    new(big.Float).Copy(contractSize),
		StrikePrice:     new(big.Float).Copy(strikePrice),
		ExpirationDate:  expirationDate,
		DeliveryDate:    deliveryDate,
		ContractType:    contractType,
		SettlementType:  settlementType,
		MarkPrice:       big.NewFloat(0),
		IndexPrice:      big.NewFloat(0),
		OpenInterest:    big.NewFloat(0),
		Volume24h:       big.NewFloat(0),
		High24h:         nil,
		Low24h:          nil,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// NewStandardFuturesPosition creates a new standard futures position
func NewStandardFuturesPosition(
	userID string,
	contract *StandardFuturesContract,
	side PositionSide,
	size, entryPrice, leverage *big.Float,
) (*StandardFuturesPosition, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if contract == nil {
		return nil, errors.New("contract cannot be nil")
	}
	if size == nil || size.Sign() <= 0 {
		return nil, errors.New("size must be positive")
	}
	if entryPrice == nil || entryPrice.Sign() <= 0 {
		return nil, errors.New("entry price must be positive")
	}
	if leverage == nil || leverage.Sign() <= 0 {
		return nil, errors.New("leverage must be positive")
	}

	now := time.Now()

	// Calculate initial margin
	margin := new(big.Float).Quo(entryPrice, leverage)

	// Calculate liquidation price (simplified - 80% of entry price for long, 120% for short)
	var liquidationPrice *big.Float
	if side == Long {
		liquidationPrice = new(big.Float).Mul(entryPrice, big.NewFloat(0.8))
	} else {
		liquidationPrice = new(big.Float).Mul(entryPrice, big.NewFloat(1.2))
	}

	return &StandardFuturesPosition{
		ID:               generatePositionID(userID, contract.Symbol),
		UserID:           userID,
		Contract:         contract,
		Side:             side,
		Size:             new(big.Float).Copy(size),
		EntryPrice:       new(big.Float).Copy(entryPrice),
		MarkPrice:        new(big.Float).Copy(entryPrice),
		UnrealizedPnL:    big.NewFloat(0),
		RealizedPnL:      big.NewFloat(0),
		Margin:           margin,
		Leverage:         new(big.Float).Copy(leverage),
		LiquidationPrice: liquidationPrice,
		IsOpen:           true,
		EntryTime:        now,
		UpdatedAt:        now,
		CloseTime:        nil,
	}, nil
}

// IsExpired checks if the futures contract has expired
func (sfc *StandardFuturesContract) IsExpired() bool {
	return time.Now().After(sfc.ExpirationDate)
}

// IsNearExpiration checks if the contract is near expiration (within 7 days)
func (sfc *StandardFuturesContract) IsNearExpiration() bool {
	sevenDays := 7 * 24 * time.Hour
	return time.Until(sfc.ExpirationDate) <= sevenDays
}

// DaysToExpiration returns the number of days until expiration
func (sfc *StandardFuturesContract) DaysToExpiration() int {
	return int(time.Until(sfc.ExpirationDate).Hours() / 24)
}

// UpdateMarkPrice updates the mark price of the contract
func (sfc *StandardFuturesContract) UpdateMarkPrice(newPrice *big.Float) error {
	if newPrice == nil || newPrice.Sign() < 0 {
		return errors.New("mark price must be non-negative")
	}

	sfc.MarkPrice = new(big.Float).Copy(newPrice)
	sfc.UpdatedAt = time.Now()

	// Update 24h high/low
	if sfc.High24h == nil || newPrice.Cmp(sfc.High24h) > 0 {
		sfc.High24h = new(big.Float).Copy(newPrice)
	}
	if sfc.Low24h == nil || newPrice.Cmp(sfc.Low24h) < 0 {
		sfc.Low24h = new(big.Float).Copy(newPrice)
	}

	return nil
}

// UpdateIndexPrice updates the index price of the contract
func (sfc *StandardFuturesContract) UpdateIndexPrice(newPrice *big.Float) error {
	if newPrice == nil || newPrice.Sign() < 0 {
		return errors.New("index price must be non-negative")
	}

	sfc.IndexPrice = new(big.Float).Copy(newPrice)
	sfc.UpdatedAt = time.Now()
	return nil
}

// UpdateVolume updates the 24h volume
func (sfc *StandardFuturesContract) UpdateVolume(volume *big.Float) error {
	if volume == nil || volume.Sign() < 0 {
		return errors.New("volume must be non-negative")
	}

	sfc.Volume24h = new(big.Float).Copy(volume)
	sfc.UpdatedAt = time.Now()
	return nil
}

// UpdateOpenInterest updates the open interest
func (sfc *StandardFuturesContract) UpdateOpenInterest(openInterest *big.Float) error {
	if openInterest == nil || openInterest.Sign() < 0 {
		return errors.New("open interest must be non-negative")
	}

	sfc.OpenInterest = new(big.Float).Copy(openInterest)
	sfc.UpdatedAt = time.Now()
	return nil
}

// GetContractValue returns the total contract value
func (sfc *StandardFuturesContract) GetContractValue() *big.Float {
	if sfc.MarkPrice == nil || sfc.ContractSize == nil {
		return big.NewFloat(0)
	}
	return new(big.Float).Mul(sfc.MarkPrice, sfc.ContractSize)
}

// GetStrikeValue returns the total strike value
func (sfc *StandardFuturesContract) GetStrikeValue() *big.Float {
	if sfc.StrikePrice == nil || sfc.ContractSize == nil {
		return big.NewFloat(0)
	}
	return new(big.Float).Mul(sfc.StrikePrice, sfc.ContractSize)
}

// UpdatePosition updates the position with new mark price and calculates PnL
func (sfp *StandardFuturesPosition) UpdatePosition(newMarkPrice *big.Float) error {
	if newMarkPrice == nil || newMarkPrice.Sign() < 0 {
		return errors.New("mark price must be non-negative")
	}

	sfp.MarkPrice = new(big.Float).Copy(newMarkPrice)
	sfp.UpdatedAt = time.Now()

	// Calculate unrealized PnL
	sfp.calculateUnrealizedPnL()

	return nil
}

// calculateUnrealizedPnL calculates the unrealized profit and loss
func (sfp *StandardFuturesPosition) calculateUnrealizedPnL() {
	if sfp.MarkPrice == nil || sfp.EntryPrice == nil || sfp.Size == nil {
		sfp.UnrealizedPnL = big.NewFloat(0)
		return
	}

	var pnl *big.Float
	if sfp.Side == Long {
		// Long position: (Mark Price - Entry Price) * Size
		pnl = new(big.Float).Sub(sfp.MarkPrice, sfp.EntryPrice)
	} else {
		// Short position: (Entry Price - Mark Price) * Size
		pnl = new(big.Float).Sub(sfp.EntryPrice, sfp.MarkPrice)
	}

	sfp.UnrealizedPnL = new(big.Float).Mul(pnl, sfp.Size)
}

// ClosePosition closes the position at the given price
func (sfp *StandardFuturesPosition) ClosePosition(closePrice *big.Float) error {
	if !sfp.IsOpen {
		return errors.New("position is already closed")
	}
	if closePrice == nil || closePrice.Sign() < 0 {
		return errors.New("close price must be non-negative")
	}

	// Calculate realized PnL
	var pnl *big.Float
	if sfp.Side == Long {
		// Long position: (Close Price - Entry Price) * Size
		pnl = new(big.Float).Sub(closePrice, sfp.EntryPrice)
	} else {
		// Short position: (Entry Price - Close Price) * Size
		pnl = new(big.Float).Sub(sfp.EntryPrice, closePrice)
	}

	sfp.RealizedPnL = new(big.Float).Mul(pnl, sfp.Size)
	sfp.UnrealizedPnL = big.NewFloat(0)
	sfp.IsOpen = false
	sfp.UpdatedAt = time.Now()
	now := time.Now()
	sfp.CloseTime = &now

	return nil
}

// GetTotalPnL returns the total profit and loss (realized + unrealized)
func (sfp *StandardFuturesPosition) GetTotalPnL() *big.Float {
	total := new(big.Float).Add(sfp.RealizedPnL, sfp.UnrealizedPnL)
	return total
}

// GetROI returns the return on investment
func (sfp *StandardFuturesPosition) GetROI() *big.Float {
	if sfp.Margin == nil || sfp.Margin.Sign() == 0 {
		return big.NewFloat(0)
	}

	totalPnL := sfp.GetTotalPnL()
	roi := new(big.Float).Quo(totalPnL, sfp.Margin)
	return roi
}

// IsLiquidated checks if the position should be liquidated
func (sfp *StandardFuturesPosition) IsLiquidated() bool {
	if !sfp.IsOpen {
		return false
	}

	if sfp.LiquidationPrice == nil || sfp.MarkPrice == nil {
		return false
	}

	if sfp.Side == Long {
		// Long position liquidated when mark price falls below liquidation price
		return sfp.MarkPrice.Cmp(sfp.LiquidationPrice) < 0
	} else {
		// Short position liquidated when mark price rises above liquidation price
		return sfp.MarkPrice.Cmp(sfp.LiquidationPrice) > 0
	}
}

// GetMarginRatio returns the current margin ratio
func (sfp *StandardFuturesPosition) GetMarginRatio() *big.Float {
	if sfp.Margin == nil || sfp.Margin.Sign() == 0 {
		return big.NewFloat(0)
	}

	// Calculate current margin value
	currentValue := new(big.Float).Mul(sfp.MarkPrice, sfp.Size)
	marginRatio := new(big.Float).Quo(sfp.Margin, currentValue)
	return marginRatio
}
