package futures

import (
	"errors"
	"math/big"
	"time"
)

// FundingRate represents the funding rate for perpetual futures
type FundingRate struct {
	Rate      *big.Float
	Timestamp time.Time
	Interval  time.Duration // Usually 8 hours
}

// PerpetualContract represents a perpetual futures contract
type PerpetualContract struct {
	Symbol           string
	UnderlyingAsset  string
	ContractSize     *big.Float
	MarkPrice        *big.Float
	IndexPrice       *big.Float
	FundingRate      *FundingRate
	NextFundingTime  time.Time
	OpenInterest     *big.Float
	Volume24h        *big.Float
	High24h          *big.Float
	Low24h           *big.Float
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PerpetualPosition represents a position in perpetual futures
type PerpetualPosition struct {
	ID              string
	UserID          string
	Contract        *PerpetualContract
	Side            PositionSide
	Size            *big.Float
	EntryPrice      *big.Float
	MarkPrice      *big.Float
	UnrealizedPnL   *big.Float
	RealizedPnL     *big.Float
	FundingPaid     *big.Float
	LiquidationPrice *big.Float
	Margin          *big.Float
	Leverage        *big.Float
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PositionSide represents the side of a futures position
type PositionSide int

const (
	Long PositionSide = iota
	Short
)

// NewPerpetualContract creates a new perpetual futures contract
func NewPerpetualContract(symbol, underlyingAsset string, contractSize *big.Float) (*PerpetualContract, error) {
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}
	if underlyingAsset == "" {
		return nil, errors.New("underlying asset cannot be empty")
	}
	if contractSize == nil || contractSize.Sign() <= 0 {
		return nil, errors.New("contract size must be positive")
	}
	
	now := time.Now()
	nextFunding := now.Add(8 * time.Hour) // Default 8-hour funding interval
	
	return &PerpetualContract{
		Symbol:          symbol,
		UnderlyingAsset: underlyingAsset,
		ContractSize:    new(big.Float).Copy(contractSize),
		MarkPrice:       big.NewFloat(0),
		IndexPrice:      big.NewFloat(0),
		FundingRate: &FundingRate{
			Rate:      big.NewFloat(0),
			Timestamp: now,
			Interval:  8 * time.Hour,
		},
		NextFundingTime: nextFunding,
		OpenInterest:    big.NewFloat(0),
		Volume24h:       big.NewFloat(0),
		High24h:         big.NewFloat(0),
		Low24h:          big.NewFloat(0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// NewPerpetualPosition creates a new perpetual futures position
func NewPerpetualPosition(userID string, contract *PerpetualContract, side PositionSide, size, entryPrice, leverage *big.Float) (*PerpetualPosition, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if contract == nil {
		return nil, errors.New("contract cannot be nil")
	}
	if size == nil || size.Sign() <= 0 {
		return nil, errors.New("position size must be positive")
	}
	if entryPrice == nil || entryPrice.Sign() <= 0 {
		return nil, errors.New("entry price must be positive")
	}
	if leverage == nil || leverage.Sign() <= 0 {
		return nil, errors.New("leverage must be positive")
	}
	
	now := time.Now()
	
	return &PerpetualPosition{
		ID:              generatePositionID(userID, contract.Symbol),
		UserID:          userID,
		Contract:        contract,
		Side:            side,
		Size:            new(big.Float).Copy(size),
		EntryPrice:      new(big.Float).Copy(entryPrice),
		MarkPrice:       new(big.Float).Copy(entryPrice),
		UnrealizedPnL:   big.NewFloat(0),
		RealizedPnL:     big.NewFloat(0),
		FundingPaid:     big.NewFloat(0),
		LiquidationPrice: big.NewFloat(0),
		Margin:          big.NewFloat(0),
		Leverage:        new(big.Float).Copy(leverage),
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// UpdateMarkPrice updates the mark price of the contract
func (pc *PerpetualContract) UpdateMarkPrice(markPrice *big.Float) error {
	if markPrice == nil || markPrice.Sign() < 0 {
		return errors.New("mark price must be non-negative")
	}
	
	pc.MarkPrice = new(big.Float).Copy(markPrice)
	pc.UpdatedAt = time.Now()
	
	// Update 24h high/low
	if pc.High24h.Sign() == 0 || markPrice.Cmp(pc.High24h) > 0 {
		pc.High24h = new(big.Float).Copy(markPrice)
	}
	if pc.Low24h.Sign() == 0 || markPrice.Cmp(pc.Low24h) < 0 {
		pc.Low24h = new(big.Float).Copy(markPrice)
	}
	
	return nil
}

// UpdateIndexPrice updates the index price of the contract
func (pc *PerpetualContract) UpdateIndexPrice(indexPrice *big.Float) error {
	if indexPrice == nil || indexPrice.Sign() < 0 {
		return errors.New("index price must be non-negative")
	}
	
	pc.IndexPrice = new(big.Float).Copy(indexPrice)
	pc.UpdatedAt = time.Now()
	
	return nil
}

// UpdateFundingRate updates the funding rate
func (pc *PerpetualContract) UpdateFundingRate(rate *big.Float) error {
	if rate == nil {
		return errors.New("funding rate cannot be nil")
	}
	
	pc.FundingRate.Rate = new(big.Float).Copy(rate)
	pc.FundingRate.Timestamp = time.Now()
	pc.NextFundingTime = time.Now().Add(pc.FundingRate.Interval)
	pc.UpdatedAt = time.Now()
	
	return nil
}

// CalculateFundingPayment calculates the funding payment for a position
func (pc *PerpetualContract) CalculateFundingPayment(position *PerpetualPosition) *big.Float {
	if position == nil || pc.FundingRate.Rate.Sign() == 0 {
		return big.NewFloat(0)
	}
	
	// Funding payment = Position Size * Mark Price * Funding Rate
	fundingPayment := new(big.Float).Mul(position.Size, position.MarkPrice)
	fundingPayment.Mul(fundingPayment, pc.FundingRate.Rate)
	
	// Adjust for position side (long pays, short receives)
	if position.Side == Long {
		return fundingPayment
	} else {
		return new(big.Float).Neg(fundingPayment)
	}
}

// IsFundingTime checks if it's time for funding
func (pc *PerpetualContract) IsFundingTime() bool {
	return time.Now().After(pc.NextFundingTime)
}

// UpdatePosition updates a position's mark price and calculates PnL
func (pp *PerpetualPosition) UpdatePosition(markPrice *big.Float) error {
	if markPrice == nil || markPrice.Sign() < 0 {
		return errors.New("mark price must be non-negative")
	}
	
	pp.MarkPrice = new(big.Float).Copy(markPrice)
	pp.UpdatedAt = time.Now()
	
	// Calculate unrealized PnL
	pp.calculateUnrealizedPnL()
	
	return nil
}

// calculateUnrealizedPnL calculates the unrealized profit/loss
func (pp *PerpetualPosition) calculateUnrealizedPnL() {
	if pp.Side == Long {
		// Long position: PnL = (Mark Price - Entry Price) * Size
		priceDiff := new(big.Float).Sub(pp.MarkPrice, pp.EntryPrice)
		pp.UnrealizedPnL = new(big.Float).Mul(priceDiff, pp.Size)
	} else {
		// Short position: PnL = (Entry Price - Mark Price) * Size
		priceDiff := new(big.Float).Sub(pp.EntryPrice, pp.MarkPrice)
		pp.UnrealizedPnL = new(big.Float).Mul(priceDiff, pp.Size)
	}
}

// CalculateMargin calculates the required margin for the position
func (pp *PerpetualPosition) CalculateMargin() *big.Float {
	// Required margin = Position Value / Leverage
	positionValue := new(big.Float).Mul(pp.Size, pp.MarkPrice)
	margin := new(big.Float).Quo(positionValue, pp.Leverage)
	
	pp.Margin = margin
	return margin
}

// CalculateLiquidationPrice calculates the liquidation price
func (pp *PerpetualPosition) CalculateLiquidationPrice(maintenanceMargin *big.Float) *big.Float {
	if maintenanceMargin == nil || maintenanceMargin.Sign() <= 0 {
		return big.NewFloat(0)
	}
	
	// Liquidation price calculation depends on position side
	if pp.Side == Long {
		// For long: Liquidation Price = Entry Price * (1 - 1/Leverage + Maintenance Margin)
		leverageRatio := new(big.Float).Quo(big.NewFloat(1), pp.Leverage)
		marginBuffer := new(big.Float).Sub(big.NewFloat(1), leverageRatio)
		marginBuffer.Add(marginBuffer, maintenanceMargin)
		
		liquidationPrice := new(big.Float).Mul(pp.EntryPrice, marginBuffer)
		pp.LiquidationPrice = liquidationPrice
		return liquidationPrice
	} else {
		// For short: Liquidation Price = Entry Price * (1 + 1/Leverage - Maintenance Margin)
		leverageRatio := new(big.Float).Quo(big.NewFloat(1), pp.Leverage)
		marginBuffer := new(big.Float).Add(big.NewFloat(1), leverageRatio)
		marginBuffer.Sub(marginBuffer, maintenanceMargin)
		
		liquidationPrice := new(big.Float).Mul(pp.EntryPrice, marginBuffer)
		pp.LiquidationPrice = liquidationPrice
		return liquidationPrice
	}
}

// IsLiquidated checks if the position should be liquidated
func (pp *PerpetualPosition) IsLiquidated(maintenanceMargin *big.Float) bool {
	if maintenanceMargin == nil || maintenanceMargin.Sign() <= 0 {
		return false
	}
	
	// Calculate current margin ratio
	positionValue := new(big.Float).Mul(pp.Size, pp.MarkPrice)
	currentMargin := new(big.Float).Quo(pp.Margin, positionValue)
	
	// If current margin is below maintenance margin, position should be liquidated
	return currentMargin.Cmp(maintenanceMargin) < 0
}

// AddFundingPayment adds a funding payment to the position
func (pp *PerpetualPosition) AddFundingPayment(payment *big.Float) {
	if payment == nil {
		return
	}
	
	pp.FundingPaid.Add(pp.FundingPaid, payment)
	pp.UpdatedAt = time.Now()
}

// ClosePosition closes a position and calculates realized PnL
func (pp *PerpetualPosition) ClosePosition(exitPrice *big.Float) *big.Float {
	if exitPrice == nil || exitPrice.Sign() < 0 {
		return big.NewFloat(0)
	}
	
	// Calculate realized PnL
	var realizedPnL *big.Float
	if pp.Side == Long {
		// Long position: PnL = (Exit Price - Entry Price) * Size
		priceDiff := new(big.Float).Sub(exitPrice, pp.EntryPrice)
		realizedPnL = new(big.Float).Mul(priceDiff, pp.Size)
	} else {
		// Short position: PnL = (Entry Price - Exit Price) * Size
		priceDiff := new(big.Float).Sub(pp.EntryPrice, exitPrice)
		realizedPnL = new(big.Float).Mul(priceDiff, pp.Size)
	}
	
	// Subtract funding payments
	realizedPnL.Sub(realizedPnL, pp.FundingPaid)
	
	pp.RealizedPnL = realizedPnL
	pp.UpdatedAt = time.Now()
	
	return realizedPnL
}

// GetTotalPnL returns the total PnL (unrealized + realized)
func (pp *PerpetualPosition) GetTotalPnL() *big.Float {
	totalPnL := new(big.Float).Add(pp.UnrealizedPnL, pp.RealizedPnL)
	return totalPnL
}

// GetROI returns the return on investment percentage
func (pp *PerpetualPosition) GetROI() *big.Float {
	if pp.Margin.Sign() == 0 {
		return big.NewFloat(0)
	}
	
	totalPnL := pp.GetTotalPnL()
	roi := new(big.Float).Quo(totalPnL, pp.Margin)
	roi.Mul(roi, big.NewFloat(100)) // Convert to percentage
	
	return roi
}

// Helper function to generate position ID
func generatePositionID(userID, symbol string) string {
	return userID + "_" + symbol + "_" + time.Now().Format("20060102150405")
}
