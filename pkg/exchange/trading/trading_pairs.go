package trading

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// TradingPair represents a trading pair (e.g., BTC/USDT)
type TradingPair struct {
	ID              string    `json:"id"`
	BaseAsset       string    `json:"base_asset"`
	QuoteAsset      string    `json:"quote_asset"`
	Symbol          string    `json:"symbol"`
	Status          PairStatus `json:"status"`
	MinQuantity     *big.Int  `json:"min_quantity"`
	MaxQuantity     *big.Int  `json:"max_quantity"`
	MinPrice        *big.Int  `json:"min_price"`
	MaxPrice        *big.Int  `json:"max_price"`
	TickSize        *big.Int  `json:"tick_size"`
	StepSize        *big.Int  `json:"step_size"`
	MakerFee        *big.Int  `json:"maker_fee"`
	TakerFee        *big.Int  `json:"taker_fee"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastTradePrice  *big.Int  `json:"last_trade_price,omitempty"`
	LastTradeTime   *time.Time `json:"last_trade_time,omitempty"`
	Volume24h       *big.Int  `json:"volume_24h"`
	PriceChange24h  *big.Int  `json:"price_change_24h"`
	PriceChangePercent24h *big.Int `json:"price_change_percent_24h"`
}

// PairStatus represents the status of a trading pair
type PairStatus string

const (
	PairStatusActive   PairStatus = "active"
	PairStatusInactive PairStatus = "inactive"
	PairStatusSuspended PairStatus = "suspended"
	PairStatusMaintenance PairStatus = "maintenance"
)

// TradingPairError represents trading pair specific errors
type TradingPairError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	PairID    string `json:"pair_id,omitempty"`
}

func (e TradingPairError) Error() string {
	return e.Operation + ": " + e.Message
}

// Trading pair errors
var (
	ErrInvalidBaseAsset  = errors.New("invalid base asset")
	ErrInvalidQuoteAsset = errors.New("invalid quote asset")
	ErrInvalidSymbol     = errors.New("invalid symbol")
	ErrInvalidTickSize   = errors.New("invalid tick size")
	ErrInvalidStepSize   = errors.New("invalid step size")
	ErrInvalidFees       = errors.New("invalid fees")
	ErrPairAlreadyExists = errors.New("trading pair already exists")
	ErrPairNotFound      = errors.New("trading pair not found")
	ErrPairInactive      = errors.New("trading pair is inactive")
)

// NewTradingPair creates a new trading pair
func NewTradingPair(
	baseAsset, quoteAsset string,
	minQuantity, maxQuantity, minPrice, maxPrice, tickSize, stepSize, makerFee, takerFee *big.Int,
) (*TradingPair, error) {
	if baseAsset == "" {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidBaseAsset.Error()}
	}
	if quoteAsset == "" {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidQuoteAsset.Error()}
	}
	if baseAsset == quoteAsset {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: "base and quote assets cannot be the same"}
	}

	// Validate tick size and step size
	if tickSize == nil || tickSize.Cmp(big.NewInt(0)) <= 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidTickSize.Error()}
	}
	if stepSize == nil || stepSize.Cmp(big.NewInt(0)) <= 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidStepSize.Error()}
	}

	// Validate fees
	if makerFee == nil || makerFee.Cmp(big.NewInt(0)) < 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidFees.Error()}
	}
	if takerFee == nil || takerFee.Cmp(big.NewInt(0)) < 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: ErrInvalidFees.Error()}
	}

	// Validate quantities and prices
	if minQuantity != nil && maxQuantity != nil && minQuantity.Cmp(maxQuantity) > 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: "min quantity cannot be greater than max quantity"}
	}
	if minPrice != nil && maxPrice != nil && minPrice.Cmp(maxPrice) > 0 {
		return nil, &TradingPairError{Operation: "NewTradingPair", Message: "min price cannot be greater than max price"}
	}

	pair := &TradingPair{
		ID:            fmt.Sprintf("%s_%s", baseAsset, quoteAsset),
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Symbol:        fmt.Sprintf("%s/%s", baseAsset, quoteAsset),
		Status:        PairStatusActive,
		MinQuantity:   minQuantity,
		MaxQuantity:   maxQuantity,
		MinPrice:      minPrice,
		MaxPrice:      maxPrice,
		TickSize:      tickSize,
		StepSize:      stepSize,
		MakerFee:      makerFee,
		TakerFee:      takerFee,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Volume24h:     big.NewInt(0),
		PriceChange24h: big.NewInt(0),
		PriceChangePercent24h: big.NewInt(0),
	}

	return pair, nil
}

// Validate validates the trading pair
func (tp *TradingPair) Validate() error {
	if tp.BaseAsset == "" {
		return ErrInvalidBaseAsset
	}
	if tp.QuoteAsset == "" {
		return ErrInvalidQuoteAsset
	}
	if tp.BaseAsset == tp.QuoteAsset {
		return errors.New("base and quote assets cannot be the same")
	}
	if tp.TickSize == nil || tp.TickSize.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidTickSize
	}
	if tp.StepSize == nil || tp.StepSize.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidStepSize
	}
	if tp.MakerFee == nil || tp.MakerFee.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidFees
	}
	if tp.TakerFee == nil || tp.TakerFee.Cmp(big.NewInt(0)) < 0 {
		return ErrInvalidFees
	}
	return nil
}

// IsActive checks if the trading pair is active
func (tp *TradingPair) IsActive() bool {
	return tp.Status == PairStatusActive
}

// CanTrade checks if the trading pair can be traded
func (tp *TradingPair) CanTrade() bool {
	return tp.IsActive() && tp.Status != PairStatusSuspended && tp.Status != PairStatusMaintenance
}

// UpdateStatus updates the status of the trading pair
func (tp *TradingPair) UpdateStatus(status PairStatus) error {
	if status != PairStatusActive && status != PairStatusInactive && 
	   status != PairStatusSuspended && status != PairStatusMaintenance {
		return &TradingPairError{Operation: "UpdateStatus", Message: "invalid status", PairID: tp.ID}
	}

	tp.Status = status
	tp.UpdatedAt = time.Now()
	return nil
}

// UpdateFees updates the fees for the trading pair
func (tp *TradingPair) UpdateFees(makerFee, takerFee *big.Int) error {
	if makerFee == nil || makerFee.Cmp(big.NewInt(0)) < 0 {
		return &TradingPairError{Operation: "UpdateFees", Message: ErrInvalidFees.Error(), PairID: tp.ID}
	}
	if takerFee == nil || takerFee.Cmp(big.NewInt(0)) < 0 {
		return &TradingPairError{Operation: "UpdateFees", Message: ErrInvalidFees.Error(), PairID: tp.ID}
	}

	tp.MakerFee = makerFee
	tp.TakerFee = takerFee
	tp.UpdatedAt = time.Now()
	return nil
}

// UpdateLastTrade updates the last trade information
func (tp *TradingPair) UpdateLastTrade(price *big.Int) {
	if price != nil && price.Cmp(big.NewInt(0)) > 0 {
		tp.LastTradePrice = price
		now := time.Now()
		tp.LastTradeTime = &now
		tp.UpdatedAt = now
	}
}

// UpdateVolume24h updates the 24-hour volume
func (tp *TradingPair) UpdateVolume24h(volume *big.Int) {
	if volume != nil && volume.Cmp(big.NewInt(0)) >= 0 {
		tp.Volume24h = volume
		tp.UpdatedAt = time.Now()
	}
}

// UpdatePriceChange24h updates the 24-hour price change
func (tp *TradingPair) UpdatePriceChange24h(priceChange, priceChangePercent *big.Int) {
	updated := false
	
	if priceChange != nil {
		tp.PriceChange24h = priceChange
		updated = true
	}
	if priceChangePercent != nil {
		tp.PriceChangePercent24h = priceChangePercent
		updated = true
	}
	
	// Only update timestamp if actual changes were made
	if updated {
		tp.UpdatedAt = time.Now()
	}
}

// ValidatePrice validates if a price is valid for this trading pair
func (tp *TradingPair) ValidatePrice(price *big.Int) error {
	if price == nil || price.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("price must be positive")
	}

	// Check if price is within allowed range
	if tp.MinPrice != nil && price.Cmp(tp.MinPrice) < 0 {
		return fmt.Errorf("price %s is below minimum %s", price.String(), tp.MinPrice.String())
	}
	if tp.MaxPrice != nil && price.Cmp(tp.MaxPrice) > 0 {
		return fmt.Errorf("price %s is above maximum %s", price.String(), tp.MaxPrice.String())
	}

	// Check if price matches tick size
	if tp.TickSize != nil {
		remainder := new(big.Int).Rem(price, tp.TickSize)
		if remainder.Cmp(big.NewInt(0)) != 0 {
			return fmt.Errorf("price %s does not match tick size %s", price.String(), tp.TickSize.String())
		}
	}

	return nil
}

// ValidateQuantity validates if a quantity is valid for this trading pair
func (tp *TradingPair) ValidateQuantity(quantity *big.Int) error {
	if quantity == nil || quantity.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("quantity must be positive")
	}

	// Check if quantity is within allowed range
	if tp.MinQuantity != nil && quantity.Cmp(tp.MinQuantity) < 0 {
		return fmt.Errorf("quantity %s is below minimum %s", quantity.String(), tp.MinQuantity.String())
	}
	if tp.MaxQuantity != nil && quantity.Cmp(tp.MaxQuantity) > 0 {
		return fmt.Errorf("quantity %s is above maximum %s", quantity.String(), tp.MaxQuantity.String())
	}

	// Check if quantity matches step size
	if tp.StepSize != nil {
		remainder := new(big.Int).Rem(quantity, tp.StepSize)
		if remainder.Cmp(big.NewInt(0)) != 0 {
			return fmt.Errorf("quantity %s does not match step size %s", quantity.String(), tp.StepSize.String())
		}
	}

	return nil
}

// CalculateFee calculates the fee for a trade
func (tp *TradingPair) CalculateFee(quantity, price *big.Int, isMaker bool) (*big.Int, error) {
	if err := tp.ValidateQuantity(quantity); err != nil {
		return nil, err
	}
	if err := tp.ValidatePrice(price); err != nil {
		return nil, err
	}

	// Calculate trade value
	tradeValue := new(big.Int).Mul(quantity, price)

	// Apply appropriate fee rate
	var feeRate *big.Int
	if isMaker {
		feeRate = tp.MakerFee
	} else {
		feeRate = tp.TakerFee
	}

	// Calculate fee (assuming fee is in basis points, e.g., 10 = 0.1%)
	fee := new(big.Int).Mul(tradeValue, feeRate)
	fee = fee.Div(fee, big.NewInt(10000)) // Divide by 10000 to get percentage

	return fee, nil
}

// Clone creates a deep copy of the trading pair
func (tp *TradingPair) Clone() *TradingPair {
	clone := *tp

	// Deep copy big.Int fields
	if tp.MinQuantity != nil {
		clone.MinQuantity = new(big.Int).Set(tp.MinQuantity)
	}
	if tp.MaxQuantity != nil {
		clone.MaxQuantity = new(big.Int).Set(tp.MaxQuantity)
	}
	if tp.MinPrice != nil {
		clone.MinPrice = new(big.Int).Set(tp.MinPrice)
	}
	if tp.MaxPrice != nil {
		clone.MaxPrice = new(big.Int).Set(tp.MaxPrice)
	}
	if tp.TickSize != nil {
		clone.TickSize = new(big.Int).Set(tp.TickSize)
	}
	if tp.StepSize != nil {
		clone.StepSize = new(big.Int).Set(tp.StepSize)
	}
	if tp.MakerFee != nil {
		clone.MakerFee = new(big.Int).Set(tp.MakerFee)
	}
	if tp.TakerFee != nil {
		clone.TakerFee = new(big.Int).Set(tp.TakerFee)
	}
	if tp.LastTradePrice != nil {
		clone.LastTradePrice = new(big.Int).Set(tp.LastTradePrice)
	}
	if tp.Volume24h != nil {
		clone.Volume24h = new(big.Int).Set(tp.Volume24h)
	}
	if tp.PriceChange24h != nil {
		clone.PriceChange24h = new(big.Int).Set(tp.PriceChange24h)
	}
	if tp.PriceChangePercent24h != nil {
		clone.PriceChangePercent24h = new(big.Int).Set(tp.PriceChangePercent24h)
	}

	// Deep copy time fields
	if tp.LastTradeTime != nil {
		lastTradeTime := *tp.LastTradeTime
		clone.LastTradeTime = &lastTradeTime
	}

	return &clone
}
