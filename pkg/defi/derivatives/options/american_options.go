package options

import (
	"errors"
	"math"
	"math/big"
	"time"
)

// AmericanOption represents an American-style option contract with early exercise
type AmericanOption struct {
	*Option
	EarlyExerciseAllowed bool
	ExerciseHistory      []*ExerciseEvent
	LastExerciseTime     *time.Time
}

// ExerciseEvent represents an early exercise event
type ExerciseEvent struct {
	Timestamp time.Time
	Price     *big.Float
	Quantity  *big.Float
	Reason    string
}

// AmericanOptionType represents the type of American option
type AmericanOptionType int

const (
	AmericanCall AmericanOptionType = iota
	AmericanPut
)

// NewAmericanOption creates a new American-style option
func NewAmericanOption(
	optionType AmericanOptionType,
	strikePrice, currentPrice, timeToExpiry, riskFreeRate, volatility *big.Float,
) (*AmericanOption, error) {
	// Convert AmericanOptionType to OptionType
	var baseType OptionType
	if optionType == AmericanCall {
		baseType = Call
	} else {
		baseType = Put
	}

	// Create base option
	baseOption, err := NewOption(baseType, strikePrice, currentPrice, timeToExpiry, riskFreeRate, volatility)
	if err != nil {
		return nil, err
	}

	return &AmericanOption{
		Option:               baseOption,
		EarlyExerciseAllowed: true,
		ExerciseHistory:      make([]*ExerciseEvent, 0),
		LastExerciseTime:     nil,
	}, nil
}

// CanExerciseEarly checks if the option can be exercised early
func (ao *AmericanOption) CanExerciseEarly() bool {
	if !ao.EarlyExerciseAllowed {
		return false
	}

	// Check if option has expired
	if ao.TimeToExpiry.Sign() <= 0 {
		return false
	}

	// For American calls, early exercise is rarely optimal unless there are dividends
	// For American puts, early exercise can be optimal when interest rates are high
	return true
}

// ShouldExerciseEarly determines if early exercise is optimal
func (ao *AmericanOption) ShouldExerciseEarly(currentMarketPrice *big.Float) (bool, error) {
	if !ao.CanExerciseEarly() {
		return false, nil
	}

	if currentMarketPrice == nil {
		return false, errors.New("current market price cannot be nil")
	}

	// For American calls, early exercise is optimal only if:
	// 1. There are significant dividends
	// 2. Interest rates are very low
	// 3. The option is deep in the money with high time value decay

	// For American puts, early exercise is optimal if:
	// 1. Interest rates are high
	// 2. The option is deep in the money
	// 3. The underlying asset has low volatility

	if ao.Type == Call {
		return ao.shouldExerciseCallEarly(currentMarketPrice)
	} else {
		return ao.shouldExercisePutEarly(currentMarketPrice)
	}
}

// shouldExerciseCallEarly determines if American call should be exercised early
func (ao *AmericanOption) shouldExerciseCallEarly(currentMarketPrice *big.Float) (bool, error) {
	// Convert to float64 for calculations
	currentPrice, _ := currentMarketPrice.Float64()
	strikePrice, _ := ao.StrikePrice.Float64()

	// Early exercise is rarely optimal for calls unless:
	// 1. Deep in the money (current price >> strike price)
	// 2. Very low time value remaining
	// 3. High dividends expected

	// Calculate intrinsic value
	intrinsicValue := math.Max(0, currentPrice-strikePrice)

	// Calculate time value using Black-Scholes
	bsModel := &BlackScholesModel{
		RiskFreeRate: ao.RiskFreeRate,
		Volatility:   ao.Volatility,
		TimeToExpiry: ao.TimeToExpiry,
	}

	// Create temporary option for pricing
	tempOption := &Option{
		Type:         ao.Type,
		StrikePrice:  ao.StrikePrice,
		CurrentPrice: currentMarketPrice,
		TimeToExpiry: ao.TimeToExpiry,
		RiskFreeRate: ao.RiskFreeRate,
		Volatility:   ao.Volatility,
	}

	optionPrice, err := bsModel.Price(tempOption)
	if err != nil {
		return false, err
	}

	optionPriceFloat, _ := optionPrice.Float64()
	timeValue := optionPriceFloat - intrinsicValue

	// Early exercise threshold: if time value is very small relative to intrinsic value
	timeValueThreshold := intrinsicValue * 0.05 // 5% threshold

	// Also consider if option is very deep in the money
	deepInTheMoney := currentPrice > strikePrice*1.5

	return timeValue < timeValueThreshold || deepInTheMoney, nil
}

// shouldExercisePutEarly determines if American put should be exercised early
func (ao *AmericanOption) shouldExercisePutEarly(currentMarketPrice *big.Float) (bool, error) {
	// Convert to float64 for calculations
	currentPrice, _ := currentMarketPrice.Float64()
	strikePrice, _ := ao.StrikePrice.Float64()
	timeToExpiry, _ := ao.TimeToExpiry.Float64()
	riskFreeRate, _ := ao.RiskFreeRate.Float64()

	// Early exercise is more likely for puts when:
	// 1. Interest rates are high
	// 2. Option is deep in the money
	// 3. Low volatility (low time value)

	// Calculate intrinsic value (not used in current logic but kept for future enhancement)
	_ = math.Max(0, strikePrice-currentPrice)

	// High interest rate threshold (5% annual)
	highInterestRate := riskFreeRate > 0.05

	// Deep in the money threshold
	deepInTheMoney := currentPrice < strikePrice*0.7

	// Short time to expiry
	shortTimeToExpiry := timeToExpiry < 0.1 // Less than 1 month

	return highInterestRate && (deepInTheMoney || shortTimeToExpiry), nil
}

// ExerciseEarly exercises the American option early
func (ao *AmericanOption) ExerciseEarly(
	quantity *big.Float,
	marketPrice *big.Float,
	reason string,
) (*ExerciseEvent, error) {
	if !ao.CanExerciseEarly() {
		return nil, errors.New("option cannot be exercised early")
	}

	if quantity == nil || quantity.Sign() <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	if marketPrice == nil {
		return nil, errors.New("market price cannot be nil")
	}

	// Check if early exercise is optimal
	shouldExercise, err := ao.ShouldExerciseEarly(marketPrice)
	if err != nil {
		return nil, err
	}

	if !shouldExercise {
		return nil, errors.New("early exercise is not optimal for this option")
	}

	// Create exercise event
	now := time.Now()
	exerciseEvent := &ExerciseEvent{
		Timestamp: now,
		Price:     new(big.Float).Copy(marketPrice),
		Quantity:  new(big.Float).Copy(quantity),
		Reason:    reason,
	}

	// Add to exercise history
	ao.ExerciseHistory = append(ao.ExerciseHistory, exerciseEvent)
	ao.LastExerciseTime = &now

	return exerciseEvent, nil
}

// GetExerciseHistory returns the exercise history
func (ao *AmericanOption) GetExerciseHistory() []*ExerciseEvent {
	return ao.ExerciseHistory
}

// GetLastExerciseTime returns the last exercise time
func (ao *AmericanOption) GetLastExerciseTime() *time.Time {
	return ao.LastExerciseTime
}

// IsExercised checks if the option has been exercised
func (ao *AmericanOption) IsExercised() bool {
	return len(ao.ExerciseHistory) > 0
}

// GetTotalExercisedQuantity returns the total quantity exercised
func (ao *AmericanOption) GetTotalExercisedQuantity() *big.Float {
	total := big.NewFloat(0)
	for _, event := range ao.ExerciseHistory {
		total.Add(total, event.Quantity)
	}
	return total
}

// GetRemainingQuantity returns the remaining quantity available for exercise
func (ao *AmericanOption) GetRemainingQuantity(originalQuantity *big.Float) *big.Float {
	if originalQuantity == nil {
		return big.NewFloat(0)
	}

	exercised := ao.GetTotalExercisedQuantity()
	remaining := new(big.Float).Sub(originalQuantity, exercised)

	if remaining.Sign() < 0 {
		return big.NewFloat(0)
	}

	return remaining
}

// CalculateEarlyExerciseValue calculates the value of early exercise
func (ao *AmericanOption) CalculateEarlyExerciseValue(
	currentMarketPrice *big.Float,
	quantity *big.Float,
) (*big.Float, error) {
	if currentMarketPrice == nil {
		return nil, errors.New("current market price cannot be nil")
	}

	if quantity == nil || quantity.Sign() <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	// Calculate intrinsic value
	var intrinsicValue *big.Float
	if ao.Type == Call {
		// Call: max(0, S - K)
		if currentMarketPrice.Cmp(ao.StrikePrice) > 0 {
			intrinsicValue = new(big.Float).Sub(currentMarketPrice, ao.StrikePrice)
		} else {
			intrinsicValue = big.NewFloat(0)
		}
	} else {
		// Put: max(0, K - S)
		if ao.StrikePrice.Cmp(currentMarketPrice) > 0 {
			intrinsicValue = new(big.Float).Sub(ao.StrikePrice, currentMarketPrice)
		} else {
			intrinsicValue = big.NewFloat(0)
		}
	}

	// Multiply by quantity
	totalValue := new(big.Float).Mul(intrinsicValue, quantity)
	return totalValue, nil
}

// GetEarlyExercisePremium calculates the premium for early exercise
func (ao *AmericanOption) GetEarlyExercisePremium(
	currentMarketPrice *big.Float,
	quantity *big.Float,
) (*big.Float, error) {
	if currentMarketPrice == nil {
		return nil, errors.New("current market price cannot be nil")
	}

	if quantity == nil || quantity.Sign() <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	// Get early exercise value
	earlyExerciseValue, err := ao.CalculateEarlyExerciseValue(currentMarketPrice, quantity)
	if err != nil {
		return nil, err
	}

	// Calculate Black-Scholes value for comparison
	bsModel := &BlackScholesModel{
		RiskFreeRate: ao.RiskFreeRate,
		Volatility:   ao.Volatility,
		TimeToExpiry: ao.TimeToExpiry,
	}

	tempOption := &Option{
		Type:         ao.Type,
		StrikePrice:  ao.StrikePrice,
		CurrentPrice: currentMarketPrice,
		TimeToExpiry: ao.TimeToExpiry,
		RiskFreeRate: ao.RiskFreeRate,
		Volatility:   ao.Volatility,
	}

	bsValue, err := bsModel.Price(tempOption)
	if err != nil {
		return nil, err
	}

	// Multiply BS value by quantity
	bsTotalValue := new(big.Float).Mul(bsValue, quantity)

	// Premium = Early Exercise Value - Black-Scholes Value
	premium := new(big.Float).Sub(earlyExerciseValue, bsTotalValue)
	return premium, nil
}
