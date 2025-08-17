package orderbook

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// Trade represents a completed trade between two orders
type Trade struct {
	ID            string    `json:"id"`
	BuyOrderID    string    `json:"buy_order_id"`
	SellOrderID   string    `json:"sell_order_id"`
	TradingPair   string    `json:"trading_pair"`
	Quantity      *big.Int  `json:"quantity"`
	Price         *big.Int  `json:"price"`
	BuyUserID     string    `json:"buy_user_id"`
	SellUserID    string    `json:"sell_user_id"`
	Timestamp     time.Time `json:"timestamp"`
	Fee           *big.Int  `json:"fee"`
	FeeCurrency   string    `json:"fee_currency"`
}

// TradeExecution represents the result of executing a trade
type TradeExecution struct {
	Trade        *Trade   `json:"trade"`
	BuyOrder     *Order   `json:"buy_order"`
	SellOrder    *Order   `json:"sell_order"`
	PartialFills []*Trade `json:"partial_fills,omitempty"`
	RemainingBuy  *big.Int `json:"remaining_buy,omitempty"`
	RemainingSell *big.Int `json:"remaining_sell,omitempty"`
}

// MatchingEngine handles order matching and trade execution
type MatchingEngine struct {
	orderBook *OrderBook
	mutex     sync.RWMutex
	trades    []*Trade
	lastTradeID int64
}

// MatchingEngineError represents matching engine specific errors
type MatchingEngineError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	OrderID   string `json:"order_id,omitempty"`
}

func (e MatchingEngineError) Error() string {
	return e.Operation + ": " + e.Message
}

// Matching engine errors
var (
	ErrNoMatchingOrders = errors.New("no matching orders found")
	ErrInvalidTrade     = errors.New("invalid trade")
)

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(orderBook *OrderBook) *MatchingEngine {
	return &MatchingEngine{
		orderBook: orderBook,
		trades:    make([]*Trade, 0),
		lastTradeID: 0,
	}
}

// ProcessOrder processes a new order and attempts to match it with existing orders
func (me *MatchingEngine) ProcessOrder(order *Order) (*TradeExecution, error) {
	if order == nil {
		return nil, &MatchingEngineError{Operation: "ProcessOrder", Message: "order is nil"}
	}

	if err := order.Validate(); err != nil {
		return nil, &MatchingEngineError{Operation: "ProcessOrder", Message: err.Error(), OrderID: order.ID}
	}

	me.mutex.Lock()
	defer me.mutex.Unlock()

	// Add the order to the order book first
	if err := me.orderBook.AddOrder(order); err != nil {
		return nil, &MatchingEngineError{Operation: "ProcessOrder", Message: err.Error(), OrderID: order.ID}
	}

	// Attempt to match the order
	return me.matchOrder(order)
}

// matchOrder attempts to match an order with existing orders in the book
func (me *MatchingEngine) matchOrder(order *Order) (*TradeExecution, error) {
	var execution *TradeExecution
	var err error

	switch order.Side {
	case OrderSideBuy:
		execution, err = me.matchBuyOrder(order)
	case OrderSideSell:
		execution, err = me.matchSellOrder(order)
	default:
		return nil, &MatchingEngineError{Operation: "matchOrder", Message: "invalid order side", OrderID: order.ID}
	}

	if err != nil && err != ErrNoMatchingOrders {
		return nil, err
	}

	return execution, nil
}

// matchBuyOrder matches a buy order with existing sell orders
func (me *MatchingEngine) matchBuyOrder(buyOrder *Order) (*TradeExecution, error) {
	if buyOrder.Type == OrderTypeMarket {
		return me.matchMarketBuyOrder(buyOrder)
	}
	return me.matchLimitBuyOrder(buyOrder)
}

// matchSellOrder matches a sell order with existing buy orders
func (me *MatchingEngine) matchSellOrder(sellOrder *Order) (*TradeExecution, error) {
	if sellOrder.Type == OrderTypeMarket {
		return me.matchMarketSellOrder(sellOrder)
	}
	return me.matchLimitSellOrder(sellOrder)
}

// matchLimitBuyOrder matches a limit buy order with existing sell orders
func (me *MatchingEngine) matchLimitBuyOrder(buyOrder *Order) (*TradeExecution, error) {
	execution := &TradeExecution{
		BuyOrder: buyOrder.Clone(),
		SellOrder: nil,
		PartialFills: make([]*Trade, 0),
		RemainingBuy: new(big.Int).Set(buyOrder.RemainingQuantity),
	}

	// Get the best ask (lowest sell price)
	bestAsk, err := me.orderBook.GetBestAsk()
	if err != nil {
		return execution, ErrNoMatchingOrders
	}

	// Check if the buy order can match with the best ask
	if buyOrder.Price.Cmp(bestAsk.Price) < 0 {
		return execution, ErrNoMatchingOrders
	}

	// Match with sell orders until the buy order is filled or no more matches
	for execution.RemainingBuy.Cmp(big.NewInt(0)) > 0 {
		bestAsk, err := me.orderBook.GetBestAsk()
		if err != nil {
			break
		}

		// Check if we can still match
		if buyOrder.Price.Cmp(bestAsk.Price) < 0 {
			break
		}

		// Calculate the quantity to trade
		tradeQuantity := new(big.Int)
		if execution.RemainingBuy.Cmp(bestAsk.RemainingQuantity) <= 0 {
			tradeQuantity.Set(execution.RemainingBuy)
		} else {
			tradeQuantity.Set(bestAsk.RemainingQuantity)
		}

		// Execute the trade
		trade, err := me.executeTrade(buyOrder, bestAsk, tradeQuantity, bestAsk.Price)
		if err != nil {
			return execution, err
		}

		execution.PartialFills = append(execution.PartialFills, trade)
		execution.RemainingBuy.Sub(execution.RemainingBuy, tradeQuantity)

		// Update the buy order in the order book
		updatedBuyOrder := buyOrder.Clone()
		updatedBuyOrder.Fill(tradeQuantity, bestAsk.Price)
		me.orderBook.UpdateOrder(updatedBuyOrder)

		// Update the sell order in the order book
		updatedSellOrder := bestAsk.Clone()
		updatedSellOrder.Fill(tradeQuantity, bestAsk.Price)
		
		// If the sell order is completely filled, remove it
		if updatedSellOrder.RemainingQuantity.Cmp(big.NewInt(0)) == 0 {
			me.orderBook.RemoveOrder(bestAsk.ID)
		} else {
			me.orderBook.UpdateOrder(updatedSellOrder)
		}

		// Check if the buy order is completely filled
		if execution.RemainingBuy.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return execution, nil
}

// matchLimitSellOrder matches a limit sell order with existing buy orders
func (me *MatchingEngine) matchLimitSellOrder(sellOrder *Order) (*TradeExecution, error) {
	execution := &TradeExecution{
		BuyOrder: nil,
		SellOrder: sellOrder.Clone(),
		PartialFills: make([]*Trade, 0),
		RemainingSell: new(big.Int).Set(sellOrder.RemainingQuantity),
	}

	// Get the best bid (highest buy price)
	bestBid, err := me.orderBook.GetBestBid()
	if err != nil {
		return execution, ErrNoMatchingOrders
	}

	// Check if the sell order can match with the best bid
	if sellOrder.Price.Cmp(bestBid.Price) > 0 {
		return execution, ErrNoMatchingOrders
	}

	// Match with buy orders until the sell order is filled or no more matches
	for execution.RemainingSell.Cmp(big.NewInt(0)) > 0 {
		bestBid, err := me.orderBook.GetBestBid()
		if err != nil {
			break
		}

		// Check if we can still match
		if sellOrder.Price.Cmp(bestBid.Price) > 0 {
			break
		}

		// Calculate the quantity to trade
		tradeQuantity := new(big.Int)
		if execution.RemainingSell.Cmp(bestBid.RemainingQuantity) <= 0 {
			tradeQuantity.Set(execution.RemainingSell)
		} else {
			tradeQuantity.Set(bestBid.RemainingQuantity)
		}

		// Execute the trade
		trade, err := me.executeTrade(bestBid, sellOrder, tradeQuantity, sellOrder.Price)
		if err != nil {
			return execution, err
		}

		execution.PartialFills = append(execution.PartialFills, trade)
		execution.RemainingSell.Sub(execution.RemainingSell, tradeQuantity)

		// Update the sell order in the order book
		updatedSellOrder := sellOrder.Clone()
		updatedSellOrder.Fill(tradeQuantity, sellOrder.Price)
		me.orderBook.UpdateOrder(updatedSellOrder)

		// Update the buy order in the order book
		updatedBuyOrder := bestBid.Clone()
		updatedBuyOrder.Fill(tradeQuantity, sellOrder.Price)
		
		// If the buy order is completely filled, remove it
		if updatedBuyOrder.RemainingQuantity.Cmp(big.NewInt(0)) == 0 {
			me.orderBook.RemoveOrder(bestBid.ID)
		} else {
			me.orderBook.UpdateOrder(updatedBuyOrder)
		}

		// Check if the sell order is completely filled
		if execution.RemainingSell.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return execution, nil
}

// matchMarketBuyOrder matches a market buy order with existing sell orders
func (me *MatchingEngine) matchMarketBuyOrder(buyOrder *Order) (*TradeExecution, error) {
	execution := &TradeExecution{
		BuyOrder: buyOrder.Clone(),
		SellOrder: nil,
		PartialFills: make([]*Trade, 0),
		RemainingBuy: new(big.Int).Set(buyOrder.RemainingQuantity),
	}

	// Market orders match at the best available price
	for execution.RemainingBuy.Cmp(big.NewInt(0)) > 0 {
		bestAsk, err := me.orderBook.GetBestAsk()
		if err != nil {
			break
		}

		// Calculate the quantity to trade
		tradeQuantity := new(big.Int)
		if execution.RemainingBuy.Cmp(bestAsk.RemainingQuantity) <= 0 {
			tradeQuantity.Set(execution.RemainingBuy)
		} else {
			tradeQuantity.Set(bestAsk.RemainingQuantity)
		}

		// Execute the trade at the ask price
		trade, err := me.executeTrade(buyOrder, bestAsk, tradeQuantity, bestAsk.Price)
		if err != nil {
			return execution, err
		}

		execution.PartialFills = append(execution.PartialFills, trade)
		execution.RemainingBuy.Sub(execution.RemainingBuy, tradeQuantity)

		// Update the buy order in the order book
		updatedBuyOrder := buyOrder.Clone()
		updatedBuyOrder.Fill(tradeQuantity, bestAsk.Price)
		me.orderBook.UpdateOrder(updatedBuyOrder)

		// Update the sell order in the order book
		updatedSellOrder := bestAsk.Clone()
		updatedSellOrder.Fill(tradeQuantity, bestAsk.Price)
		
		// If the sell order is completely filled, remove it
		if updatedSellOrder.RemainingQuantity.Cmp(big.NewInt(0)) == 0 {
			me.orderBook.RemoveOrder(bestAsk.ID)
		} else {
			me.orderBook.UpdateOrder(updatedSellOrder)
		}

		// Check if the buy order is completely filled
		if execution.RemainingBuy.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return execution, nil
}

// matchMarketSellOrder matches a market sell order with existing buy orders
func (me *MatchingEngine) matchMarketSellOrder(sellOrder *Order) (*TradeExecution, error) {
	execution := &TradeExecution{
		BuyOrder: nil,
		SellOrder: sellOrder.Clone(),
		PartialFills: make([]*Trade, 0),
		RemainingSell: new(big.Int).Set(sellOrder.RemainingQuantity),
	}

	// Market orders match at the best available price
	for execution.RemainingSell.Cmp(big.NewInt(0)) > 0 {
		bestBid, err := me.orderBook.GetBestBid()
		if err != nil {
			break
		}

		// Calculate the quantity to trade
		tradeQuantity := new(big.Int)
		if execution.RemainingSell.Cmp(bestBid.RemainingQuantity) <= 0 {
			tradeQuantity.Set(execution.RemainingSell)
		} else {
			tradeQuantity.Set(bestBid.RemainingQuantity)
		}

		// Execute the trade at the bid price
		trade, err := me.executeTrade(bestBid, sellOrder, tradeQuantity, bestBid.Price)
		if err != nil {
			return execution, err
		}

		execution.PartialFills = append(execution.PartialFills, trade)
		execution.RemainingSell.Sub(execution.RemainingSell, tradeQuantity)

		// Update the sell order in the order book
		updatedSellOrder := sellOrder.Clone()
		updatedSellOrder.Fill(tradeQuantity, bestBid.Price)
		me.orderBook.UpdateOrder(updatedSellOrder)

		// Update the buy order in the order book
		updatedBuyOrder := bestBid.Clone()
		updatedBuyOrder.Fill(tradeQuantity, bestBid.Price)
		
		// If the buy order is completely filled, remove it
		if updatedBuyOrder.RemainingQuantity.Cmp(big.NewInt(0)) == 0 {
			me.orderBook.RemoveOrder(bestBid.ID)
		} else {
			me.orderBook.UpdateOrder(updatedBuyOrder)
		}

		// Check if the sell order is completely filled
		if execution.RemainingSell.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return execution, nil
}

// executeTrade executes a trade between two orders
func (me *MatchingEngine) executeTrade(buyOrder, sellOrder *Order, quantity, price *big.Int) (*Trade, error) {
	if quantity == nil || quantity.Cmp(big.NewInt(0)) <= 0 {
		return nil, &MatchingEngineError{Operation: "executeTrade", Message: "invalid quantity"}
	}

	if price == nil || price.Cmp(big.NewInt(0)) <= 0 {
		return nil, &MatchingEngineError{Operation: "executeTrade", Message: "invalid price"}
	}

	// Create the trade
	trade := &Trade{
		ID:          me.generateTradeID(),
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		TradingPair: buyOrder.TradingPair,
		Quantity:    new(big.Int).Set(quantity),
		Price:       new(big.Int).Set(price),
		BuyUserID:   buyOrder.UserID,
		SellUserID:  sellOrder.UserID,
		Timestamp:   time.Now(),
		Fee:         big.NewInt(0), // Fee calculation would be implemented separately
		FeeCurrency: "USDT",
	}

	// Add the trade to the trades list
	me.trades = append(me.trades, trade)

	return trade, nil
}

// generateTradeID generates a unique trade ID
func (me *MatchingEngine) generateTradeID() string {
	me.lastTradeID++
	return "trade_" + string(rune(me.lastTradeID))
}

// GetTrades returns all trades
func (me *MatchingEngine) GetTrades() []*Trade {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	trades := make([]*Trade, len(me.trades))
	copy(trades, me.trades)
	return trades
}

// GetTradesByTradingPair returns trades for a specific trading pair
func (me *MatchingEngine) GetTradesByTradingPair(tradingPair string) []*Trade {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	var filteredTrades []*Trade
	for _, trade := range me.trades {
		if trade.TradingPair == tradingPair {
			filteredTrades = append(filteredTrades, trade)
		}
	}

	return filteredTrades
}

// GetTradesByUser returns trades for a specific user
func (me *MatchingEngine) GetTradesByUser(userID string) []*Trade {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	var filteredTrades []*Trade
	for _, trade := range me.trades {
		if trade.BuyUserID == userID || trade.SellUserID == userID {
			filteredTrades = append(filteredTrades, trade)
		}
	}

	return filteredTrades
}

// GetTradeCount returns the total number of trades
func (me *MatchingEngine) GetTradeCount() int {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	return len(me.trades)
}

// ClearTrades clears all trades (useful for testing)
func (me *MatchingEngine) ClearTrades() {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	me.trades = make([]*Trade, 0)
	me.lastTradeID = 0
}
