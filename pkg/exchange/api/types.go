package api

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gochain/gochain/pkg/exchange/orderbook"
)

// CreateOrderRequest represents a request to create a new order
type CreateOrderRequest struct {
	TradingPair string         `json:"trading_pair"`
	Side        orderbook.OrderSide `json:"side"`
	Type        orderbook.OrderType `json:"type"`
	Quantity    *big.Int       `json:"quantity"`
	Price       *big.Int       `json:"price"`
	UserID      string         `json:"user_id"`
}

// Validate validates the create order request
func (req *CreateOrderRequest) Validate() error {
	if req.TradingPair == "" {
		return fmt.Errorf("trading pair is required")
	}
	if req.Side == "" {
		return fmt.Errorf("order side is required")
	}
	if req.Type == "" {
		return fmt.Errorf("order type is required")
	}
	if req.Quantity == nil || req.Quantity.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if req.Price == nil || req.Price.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

// CreateOrderResponse represents the response to creating an order
type CreateOrderResponse struct {
	OrderID   string                    `json:"order_id"`
	Status    orderbook.OrderStatus     `json:"status"`
	Execution *orderbook.TradeExecution `json:"execution,omitempty"`
	Timestamp time.Time                 `json:"timestamp"`
}

// CancelOrderResponse represents the response to cancelling an order
type CancelOrderResponse struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookResponse represents the order book data
type OrderBookResponse struct {
	TradingPair string              `json:"trading_pair"`
	Timestamp   time.Time           `json:"timestamp"`
	Bids        []*orderbook.Order  `json:"bids"`
	Asks        []*orderbook.Order  `json:"asks"`
}

// MarketDataResponse represents market data for a trading pair
type MarketDataResponse struct {
	TradingPair    string    `json:"trading_pair"`
	LastPrice      *big.Int  `json:"last_price"`
	Volume24h      *big.Int  `json:"volume_24h"`
	PriceChange24h *big.Int  `json:"price_change_24h"`
	Timestamp      time.Time `json:"timestamp"`
}

// OrderBookEntry represents a single entry in the order book
type OrderBookEntry struct {
	Price    *big.Int `json:"price"`
	Quantity *big.Int `json:"quantity"`
	Total    *big.Int `json:"total"`
}

// MarketSummary represents a summary of market data
type MarketSummary struct {
	TradingPair      string    `json:"trading_pair"`
	LastPrice        *big.Int  `json:"last_price"`
	Bid              *big.Int  `json:"bid"`
	Ask              *big.Int  `json:"ask"`
	Volume24h        *big.Int  `json:"volume_24h"`
	PriceChange24h   *big.Int  `json:"price_change_24h"`
	PriceChangePercent24h *big.Int `json:"price_change_percent_24h"`
	High24h          *big.Int  `json:"high_24h"`
	Low24h           *big.Int  `json:"low_24h"`
	Timestamp        time.Time `json:"timestamp"`
}
