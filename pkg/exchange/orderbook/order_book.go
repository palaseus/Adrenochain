package orderbook

import (
	"container/heap"
	"errors"
	"math/big"
	"sync"
	"time"
)

// OrderBook represents a trading order book for a specific trading pair
type OrderBook struct {
	tradingPair string
	buyOrders   *OrderHeap  // Max heap for buy orders (highest price first)
	sellOrders  *OrderHeap  // Min heap for sell orders (lowest price first)
	orders      map[string]*Order // Map of order ID to order for quick lookup
	mutex       sync.RWMutex
	lastUpdate  time.Time
}

// OrderHeap implements heap.Interface for efficient order management
type OrderHeap []*Order

// OrderBookError represents order book specific errors
type OrderBookError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	OrderID   string `json:"order_id,omitempty"`
}

func (e OrderBookError) Error() string {
	return e.Operation + ": " + e.Message
}

// Order book errors
var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderBookEmpty    = errors.New("order book is empty")
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")
)

// NewOrderBook creates a new order book for a trading pair
func NewOrderBook(tradingPair string) (*OrderBook, error) {
	if tradingPair == "" {
		return nil, ErrInvalidTradingPair
	}

	return &OrderBook{
		tradingPair: tradingPair,
		buyOrders:   &OrderHeap{},
		sellOrders:  &OrderHeap{},
		orders:      make(map[string]*Order),
		lastUpdate:  time.Now(),
	}, nil
}

// AddOrder adds a new order to the order book
func (ob *OrderBook) AddOrder(order *Order) error {
	if order == nil {
		return &OrderBookError{Operation: "AddOrder", Message: "order is nil"}
	}

	if order.TradingPair != ob.tradingPair {
		return &OrderBookError{Operation: "AddOrder", Message: "trading pair mismatch"}
	}

	if err := order.Validate(); err != nil {
		return &OrderBookError{Operation: "AddOrder", Message: err.Error(), OrderID: order.ID}
	}

	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	// Check if order already exists
	if _, exists := ob.orders[order.ID]; exists {
		return &OrderBookError{Operation: "AddOrder", Message: ErrOrderAlreadyExists.Error(), OrderID: order.ID}
	}

	// Add order to the appropriate heap based on side
	switch order.Side {
	case OrderSideBuy:
		heap.Push(ob.buyOrders, order)
	case OrderSideSell:
		heap.Push(ob.sellOrders, order)
	default:
		return &OrderBookError{Operation: "AddOrder", Message: ErrInvalidOrderSide.Error(), OrderID: order.ID}
	}

	// Add order to the orders map
	ob.orders[order.ID] = order
	ob.lastUpdate = time.Now()

	return nil
}

// RemoveOrder removes an order from the order book
func (ob *OrderBook) RemoveOrder(orderID string) error {
	if orderID == "" {
		return &OrderBookError{Operation: "RemoveOrder", Message: "order ID is empty"}
	}

	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	order, exists := ob.orders[orderID]
	if !exists {
		return &OrderBookError{Operation: "RemoveOrder", Message: ErrOrderNotFound.Error(), OrderID: orderID}
	}

	// Remove from appropriate heap
	switch order.Side {
	case OrderSideBuy:
		ob.removeFromHeap(ob.buyOrders, orderID)
	case OrderSideSell:
		ob.removeFromHeap(ob.sellOrders, orderID)
	}

	// Remove from orders map
	delete(ob.orders, orderID)
	ob.lastUpdate = time.Now()

	return nil
}

// removeFromHeap removes an order from a heap by ID
func (ob *OrderBook) removeFromHeap(h *OrderHeap, orderID string) {
	for i, order := range *h {
		if order.ID == orderID {
			heap.Remove(h, i)
			break
		}
	}
}

// UpdateOrder updates an existing order in the order book
func (ob *OrderBook) UpdateOrder(updatedOrder *Order) error {
	if updatedOrder == nil {
		return &OrderBookError{Operation: "UpdateOrder", Message: "updated order is nil"}
	}

	if updatedOrder.TradingPair != ob.tradingPair {
		return &OrderBookError{Operation: "UpdateOrder", Message: "trading pair mismatch"}
	}

	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	// Check if order exists
	existingOrder, exists := ob.orders[updatedOrder.ID]
	if !exists {
		return &OrderBookError{Operation: "UpdateOrder", Message: ErrOrderNotFound.Error(), OrderID: updatedOrder.ID}
	}

	// Remove the old order from the heap
	switch existingOrder.Side {
	case OrderSideBuy:
		ob.removeFromHeap(ob.buyOrders, updatedOrder.ID)
	case OrderSideSell:
		ob.removeFromHeap(ob.sellOrders, updatedOrder.ID)
	}

	// Add the updated order to the appropriate heap
	switch updatedOrder.Side {
	case OrderSideBuy:
		heap.Push(ob.buyOrders, updatedOrder)
	case OrderSideSell:
		heap.Push(ob.sellOrders, updatedOrder)
	default:
		return &OrderBookError{Operation: "UpdateOrder", Message: ErrInvalidOrderSide.Error(), OrderID: updatedOrder.ID}
	}

	// Update the orders map
	ob.orders[updatedOrder.ID] = updatedOrder
	ob.lastUpdate = time.Now()

	return nil
}

// GetOrder retrieves an order by ID
func (ob *OrderBook) GetOrder(orderID string) (*Order, error) {
	if orderID == "" {
		return nil, &OrderBookError{Operation: "GetOrder", Message: "order ID is empty"}
	}

	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	order, exists := ob.orders[orderID]
	if !exists {
		return nil, &OrderBookError{Operation: "GetOrder", Message: ErrOrderNotFound.Error(), OrderID: orderID}
	}

	return order.Clone(), nil
}

// GetBestBid returns the best bid (highest buy price)
func (ob *OrderBook) GetBestBid() (*Order, error) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if ob.buyOrders.Len() == 0 {
		return nil, &OrderBookError{Operation: "GetBestBid", Message: ErrOrderBookEmpty.Error()}
	}

	bestBid := (*ob.buyOrders)[0]
	return bestBid.Clone(), nil
}

// GetBestAsk returns the best ask (lowest sell price)
func (ob *OrderBook) GetBestAsk() (*Order, error) {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	if ob.sellOrders.Len() == 0 {
		return nil, &OrderBookError{Operation: "GetBestAsk", Message: ErrOrderBookEmpty.Error()}
	}

	bestAsk := (*ob.sellOrders)[0]
	return bestAsk.Clone(), nil
}

// GetSpread returns the current spread (best ask - best bid)
func (ob *OrderBook) GetSpread() (*big.Int, error) {
	bestBid, err := ob.GetBestBid()
	if err != nil {
		return nil, err
	}

	bestAsk, err := ob.GetBestAsk()
	if err != nil {
		return nil, err
	}

	spread := new(big.Int).Sub(bestAsk.Price, bestBid.Price)
	return spread, nil
}

// GetMidPrice returns the mid price ((best ask + best bid) / 2)
func (ob *OrderBook) GetMidPrice() (*big.Int, error) {
	bestBid, err := ob.GetBestBid()
	if err != nil {
		return nil, err
	}

	bestAsk, err := ob.GetBestAsk()
	if err != nil {
		return nil, err
	}

	sum := new(big.Int).Add(bestAsk.Price, bestBid.Price)
	midPrice := new(big.Int).Div(sum, big.NewInt(2))
	return midPrice, nil
}

// GetOrderCount returns the total number of orders in the book
func (ob *OrderBook) GetOrderCount() int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return len(ob.orders)
}

// GetBuyOrderCount returns the number of buy orders
func (ob *OrderBook) GetBuyOrderCount() int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return ob.buyOrders.Len()
}

// GetSellOrderCount returns the number of sell orders
func (ob *OrderBook) GetSellOrderCount() int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return ob.sellOrders.Len()
}

// GetTotalVolume returns the total volume of all orders
func (ob *OrderBook) GetTotalVolume() *big.Int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	totalVolume := big.NewInt(0)
	for _, order := range ob.orders {
		totalVolume.Add(totalVolume, order.Quantity)
	}

	return totalVolume
}

// GetBuyVolume returns the total volume of buy orders
func (ob *OrderBook) GetBuyVolume() *big.Int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	buyVolume := big.NewInt(0)
	for _, order := range *ob.buyOrders {
		buyVolume.Add(buyVolume, order.RemainingQuantity)
	}

	return buyVolume
}

// GetSellVolume returns the total volume of sell orders
func (ob *OrderBook) GetSellVolume() *big.Int {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	sellVolume := big.NewInt(0)
	for _, order := range *ob.sellOrders {
		sellVolume.Add(sellVolume, order.RemainingQuantity)
	}

	return sellVolume
}

// GetDepth returns the order book depth at specified price levels
func (ob *OrderBook) GetDepth(levels int) ([]PriceLevel, error) {
	if levels <= 0 {
		return nil, &OrderBookError{Operation: "GetDepth", Message: "invalid number of levels"}
	}

	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	depth := make([]PriceLevel, 0, levels*2) // Buy + Sell levels

	// Get buy levels (highest to lowest)
	buyLevels := ob.getPriceLevels(ob.buyOrders, levels, OrderSideBuy)
	depth = append(depth, buyLevels...)

	// Get sell levels (lowest to highest)
	sellLevels := ob.getPriceLevels(ob.sellOrders, levels, OrderSideSell)
	depth = append(depth, sellLevels...)

	return depth, nil
}

// PriceLevel represents a price level with aggregated volume
type PriceLevel struct {
	Price   *big.Int `json:"price"`
	Volume  *big.Int `json:"volume"`
	Side    OrderSide `json:"side"`
	OrderCount int    `json:"order_count"`
}

// getPriceLevels aggregates orders by price level
func (ob *OrderBook) getPriceLevels(h *OrderHeap, levels int, side OrderSide) []PriceLevel {
	if h.Len() == 0 {
		return []PriceLevel{}
	}

	priceLevels := make(map[string]*PriceLevel)
	
	// Clone the heap to avoid modifying the original
	heapCopy := make(OrderHeap, h.Len())
	copy(heapCopy, *h)
	
	// Process orders and aggregate by price
	for heapCopy.Len() > 0 && len(priceLevels) < levels {
		order := heap.Pop(&heapCopy).(*Order)
		priceKey := order.Price.String()
		
		if level, exists := priceLevels[priceKey]; exists {
			level.Volume.Add(level.Volume, order.RemainingQuantity)
			level.OrderCount++
		} else {
			priceLevels[priceKey] = &PriceLevel{
				Price:      order.Price,
				Volume:     new(big.Int).Set(order.RemainingQuantity),
				Side:       side,
				OrderCount: 1,
			}
		}
	}

	// Convert map to slice and sort
	result := make([]PriceLevel, 0, len(priceLevels))
	for _, level := range priceLevels {
		result = append(result, *level)
	}

	// Sort by price (descending for buy, ascending for sell)
	if side == OrderSideBuy {
		// Sort buy orders by price (highest first)
		for i := 0; i < len(result)-1; i++ {
			for j := i + 1; j < len(result); j++ {
				if result[i].Price.Cmp(result[j].Price) < 0 {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
	} else {
		// Sort sell orders by price (lowest first)
		for i := 0; i < len(result)-1; i++ {
			for j := i + 1; j < len(result); j++ {
				if result[i].Price.Cmp(result[j].Price) > 0 {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
	}

	return result
}

// CancelExpiredOrders removes expired orders from the order book
func (ob *OrderBook) CancelExpiredOrders() int {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	cancelledCount := 0
	expiredOrders := make([]string, 0)

	// Find expired orders
	for orderID, order := range ob.orders {
		if order.IsExpired() {
			expiredOrders = append(expiredOrders, orderID)
		}
	}

	// Remove expired orders from the order book
	for _, orderID := range expiredOrders {
		order := ob.orders[orderID]
		
		// Remove from appropriate heap
		switch order.Side {
		case OrderSideBuy:
			ob.removeFromHeap(ob.buyOrders, orderID)
		case OrderSideSell:
			ob.removeFromHeap(ob.sellOrders, orderID)
		}
		
		// Remove from orders map
		delete(ob.orders, orderID)
		cancelledCount++
	}

	ob.lastUpdate = time.Now()
	return cancelledCount
}

// GetLastUpdate returns the last update time
func (ob *OrderBook) GetLastUpdate() time.Time {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return ob.lastUpdate
}

// GetTradingPair returns the trading pair
func (ob *OrderBook) GetTradingPair() string {
	return ob.tradingPair
}

// IsEmpty checks if the order book is empty
func (ob *OrderBook) IsEmpty() bool {
	ob.mutex.RLock()
	defer ob.mutex.RUnlock()

	return len(ob.orders) == 0
}

// Clear removes all orders from the order book
func (ob *OrderBook) Clear() {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	ob.buyOrders = &OrderHeap{}
	ob.sellOrders = &OrderHeap{}
	ob.orders = make(map[string]*Order)
	ob.lastUpdate = time.Now()
}

// Heap implementation for OrderHeap
func (h OrderHeap) Len() int { return len(h) }

func (h OrderHeap) Less(i, j int) bool {
	// For buy orders: higher price first, then earlier timestamp
	// For sell orders: lower price first, then earlier timestamp
	
	orderI := h[i]
	orderJ := h[j]
	
	// Handle nil prices (market orders)
	if orderI.Price == nil && orderJ.Price == nil {
		// Both are market orders, sort by timestamp
		return orderI.CreatedAt.Before(orderJ.CreatedAt)
	}
	if orderI.Price == nil {
		// orderI is market order, it should come first
		return true
	}
	if orderJ.Price == nil {
		// orderJ is market order, it should come first
		return false
	}
	
	if orderI.Side == OrderSideBuy {
		// Buy orders: higher price first
		priceCmp := orderI.Price.Cmp(orderJ.Price)
		if priceCmp != 0 {
			return priceCmp > 0
		}
		// If same price, earlier timestamp first
		return orderI.CreatedAt.Before(orderJ.CreatedAt)
	} else {
		// Sell orders: lower price first
		priceCmp := orderI.Price.Cmp(orderJ.Price)
		if priceCmp != 0 {
			return priceCmp < 0
		}
		// If same price, earlier timestamp first
		return orderI.CreatedAt.Before(orderJ.CreatedAt)
	}
}

func (h OrderHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *OrderHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *OrderHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
