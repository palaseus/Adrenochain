package orderbook

import (
	"errors"
	"math/big"
	"time"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeLimit      OrderType = "limit"
	OrderTypeMarket     OrderType = "market"
	OrderTypeStopLoss   OrderType = "stop_loss"
	OrderTypeTakeProfit OrderType = "take_profit"
)

// OrderSide represents the side of the order (buy or sell)
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderStatus represents the current status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPartial   OrderStatus = "partial"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRejected  OrderStatus = "rejected"
)

// Order represents a trading order in the order book
type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	TradingPair   string      `json:"trading_pair"`
	Side          OrderSide   `json:"side"`
	Type          OrderType   `json:"type"`
	Status        OrderStatus `json:"status"`
	Quantity      *big.Int    `json:"quantity"`
	Price         *big.Int    `json:"price"`
	FilledQuantity *big.Int   `json:"filled_quantity"`
	RemainingQuantity *big.Int `json:"remaining_quantity"`
	StopPrice     *big.Int    `json:"stop_price,omitempty"`
	TimeInForce   TimeInForce `json:"time_in_force"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
}

// TimeInForce represents when the order should be executed
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "gtc" // Good Till Cancelled
	TimeInForceIOC TimeInForce = "ioc" // Immediate Or Cancel
	TimeInForceFOK TimeInForce = "fok" // Fill Or Kill
)

// OrderValidationError represents validation errors
type OrderValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e OrderValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validation errors
var (
	ErrInvalidOrderID          = errors.New("invalid order ID")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidTradingPair      = errors.New("invalid trading pair")
	ErrInvalidQuantity         = errors.New("invalid quantity")
	ErrInvalidPrice            = errors.New("invalid price")
	ErrInvalidStopPrice        = errors.New("invalid stop price")
	ErrInvalidTimeInForce      = errors.New("invalid time in force")
	ErrInvalidOrderType        = errors.New("invalid order type")
	ErrInvalidOrderSide        = errors.New("invalid order side")
	ErrMarketOrderWithPrice    = errors.New("market orders cannot have a price")
	ErrStopOrderWithoutStopPrice = errors.New("stop orders must have a stop price")
	ErrExpiredOrder            = errors.New("order has expired")
	ErrOrderAlreadyFilled      = errors.New("order is already filled")
	ErrOrderAlreadyCancelled   = errors.New("order is already cancelled")
	ErrOrderAlreadyRejected    = errors.New("order is already rejected")
)

// NewOrder creates a new order with validation
func NewOrder(
	id, userID, tradingPair string,
	side OrderSide,
	orderType OrderType,
	quantity, price *big.Int,
	timeInForce TimeInForce,
	stopPrice *big.Int,
	expiresAt *time.Time,
) (*Order, error) {
	order := &Order{
		ID:              id,
		UserID:          userID,
		TradingPair:     tradingPair,
		Side:            side,
		Type:            orderType,
		Status:          OrderStatusPending,
		Quantity:        quantity,
		Price:           price,
		FilledQuantity:  big.NewInt(0),
		RemainingQuantity: quantity,
		StopPrice:       stopPrice,
		TimeInForce:     timeInForce,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpiresAt:       expiresAt,
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// Validate validates the order
func (o *Order) Validate() error {
	// Validate required fields
	if o.ID == "" {
		return ErrInvalidOrderID
	}
	if o.UserID == "" {
		return ErrInvalidUserID
	}
	if o.TradingPair == "" {
		return ErrInvalidTradingPair
	}

	// Validate order side
	if o.Side != OrderSideBuy && o.Side != OrderSideSell {
		return ErrInvalidOrderSide
	}

	// Validate order type
	if o.Type != OrderTypeLimit && o.Type != OrderTypeMarket && 
	   o.Type != OrderTypeStopLoss && o.Type != OrderTypeTakeProfit {
		return ErrInvalidOrderType
	}

	// Validate quantity
	if o.Quantity == nil || o.Quantity.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidQuantity
	}

	// Validate price based on order type
	if o.Type == OrderTypeMarket {
		if o.Price != nil && o.Price.Cmp(big.NewInt(0)) > 0 {
			return ErrMarketOrderWithPrice
		}
	} else {
		if o.Price == nil || o.Price.Cmp(big.NewInt(0)) <= 0 {
			return ErrInvalidPrice
		}
	}

	// Validate stop price for stop orders
	if o.Type == OrderTypeStopLoss || o.Type == OrderTypeTakeProfit {
		if o.StopPrice == nil || o.StopPrice.Cmp(big.NewInt(0)) <= 0 {
			return ErrStopOrderWithoutStopPrice
		}
	}

	// Validate time in force
	if o.TimeInForce != TimeInForceGTC && o.TimeInForce != TimeInForceIOC && 
	   o.TimeInForce != TimeInForceFOK {
		return ErrInvalidTimeInForce
	}

	// Validate expiration
	if o.ExpiresAt != nil && o.ExpiresAt.Before(time.Now()) {
		return ErrExpiredOrder
	}

	return nil
}

// IsValid checks if the order is valid
func (o *Order) IsValid() bool {
	return o.Validate() == nil
}

// CanFill checks if the order can be filled
func (o *Order) CanFill() bool {
	if o.Status == OrderStatusFilled || 
	   o.Status == OrderStatusCancelled || 
	   o.Status == OrderStatusRejected {
		return false
	}

	if o.ExpiresAt != nil && time.Now().After(*o.ExpiresAt) {
		return false
	}

	return o.RemainingQuantity.Cmp(big.NewInt(0)) > 0
}

// Fill fills the order with the given quantity and price
func (o *Order) Fill(fillQuantity, fillPrice *big.Int) error {
	if !o.CanFill() {
		return ErrOrderAlreadyFilled
	}

	if fillQuantity == nil || fillQuantity.Cmp(big.NewInt(0)) <= 0 {
		return ErrInvalidQuantity
	}

	if fillQuantity.Cmp(o.RemainingQuantity) > 0 {
		return ErrInvalidQuantity
	}

	// Update filled quantity
	o.FilledQuantity.Add(o.FilledQuantity, fillQuantity)
	o.RemainingQuantity.Sub(o.RemainingQuantity, fillQuantity)
	o.UpdatedAt = time.Now()

	// Update status
	if o.RemainingQuantity.Cmp(big.NewInt(0)) == 0 {
		o.Status = OrderStatusFilled
	} else {
		o.Status = OrderStatusPartial
	}

	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusFilled {
		return ErrOrderAlreadyFilled
	}
	if o.Status == OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}
	if o.Status == OrderStatusRejected {
		return ErrOrderAlreadyRejected
	}

	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

// Reject rejects the order
func (o *Order) Reject() error {
	if o.Status == OrderStatusFilled {
		return ErrOrderAlreadyFilled
	}
	if o.Status == OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}
	if o.Status == OrderStatusRejected {
		return ErrOrderAlreadyRejected
	}

	o.Status = OrderStatusRejected
	o.UpdatedAt = time.Now()
	return nil
}

// GetFillPrice returns the effective fill price for the order
func (o *Order) GetFillPrice() *big.Int {
	if o.Type == OrderTypeMarket {
		// Market orders don't have a price, return nil
		return nil
	}
	return o.Price
}

// IsExpired checks if the order has expired
func (o *Order) IsExpired() bool {
	if o.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiresAt)
}

// GetPriority returns the priority value for order matching
// Higher priority orders are matched first
func (o *Order) GetPriority() int64 {
	// Base priority is the timestamp (earlier = higher priority)
	priority := o.CreatedAt.UnixNano()
	
	// For limit orders, price also affects priority
	if o.Type == OrderTypeLimit && o.Price != nil {
		// Convert price to priority (higher price for buy orders, lower for sell)
		pricePriority := o.Price.Int64()
		if o.Side == OrderSideBuy {
			priority += pricePriority * 1000000 // Buy orders: higher price = higher priority
		} else {
			priority -= pricePriority * 1000000 // Sell orders: lower price = higher priority
		}
	}
	
	return priority
}

// Clone creates a deep copy of the order
func (o *Order) Clone() *Order {
	clone := *o
	
	// Deep copy big.Int fields
	if o.Quantity != nil {
		clone.Quantity = new(big.Int).Set(o.Quantity)
	}
	if o.Price != nil {
		clone.Price = new(big.Int).Set(o.Price)
	}
	if o.FilledQuantity != nil {
		clone.FilledQuantity = new(big.Int).Set(o.FilledQuantity)
	}
	if o.RemainingQuantity != nil {
		clone.RemainingQuantity = new(big.Int).Set(o.RemainingQuantity)
	}
	if o.StopPrice != nil {
		clone.StopPrice = new(big.Int).Set(o.StopPrice)
	}
	
	return &clone
}
