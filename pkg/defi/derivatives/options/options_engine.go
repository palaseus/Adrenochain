package options

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// OrderStatus represents the status of an options order
type OrderStatus int

const (
	Pending OrderStatus = iota
	Filled
	Cancelled
	Rejected
	Expired
)

// OrderSide represents the side of an options order
type OrderSide int

const (
	Buy OrderSide = iota
	Sell
)

// OrderType represents the type of options order
type OrderType int

const (
	Market OrderType = iota
	Limit
	Stop
	StopLimit
)

// OptionsOrder represents an options trading order
type OptionsOrder struct {
	ID             string
	Option         *Option
	Side           OrderSide
	Type           OrderType
	Quantity       *big.Float
	Price          *big.Float
	StopPrice      *big.Float
	Status         OrderStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	FilledAt       *time.Time
	FilledPrice    *big.Float
	FilledQuantity *big.Float
	UserID         string
	ClientOrderID  string
}

// Position represents an options position
type Position struct {
	ID            string
	Option        *Option
	Quantity      *big.Float
	AveragePrice  *big.Float
	UnrealizedPnL *big.Float
	RealizedPnL   *big.Float
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UserID        string
}

// Trade represents an executed options trade
type Trade struct {
	ID             string
	OrderID        string
	Option         *Option
	Side           OrderSide
	Quantity       *big.Float
	Price          *big.Float
	Timestamp      time.Time
	UserID         string
	CounterpartyID string
}

// OptionsEngine manages options trading operations
type OptionsEngine struct {
	mu          sync.RWMutex
	orders      map[string]*OptionsOrder
	positions   map[string]*Position
	trades      map[string]*Trade
	orderBook   *OptionsOrderBook
	riskManager *OptionsRiskManager
	nextOrderID int64
	nextTradeID int64
}

// OptionsOrderBook manages the order book for options
type OptionsOrderBook struct {
	mu         sync.RWMutex
	buyOrders  map[string]*OptionsOrder // OrderID -> Order
	sellOrders map[string]*OptionsOrder // OrderID -> Order
}

// OptionsRiskManager manages risk for options trading
type OptionsRiskManager struct {
	mu              sync.RWMutex
	maxPositionSize *big.Float
	maxLoss         *big.Float
	positionLimits  map[string]*big.Float // UserID -> Max Position Size
}

// NewOptionsEngine creates a new options trading engine
func NewOptionsEngine() *OptionsEngine {
	return &OptionsEngine{
		orders:      make(map[string]*OptionsOrder),
		positions:   make(map[string]*Position),
		trades:      make(map[string]*Trade),
		orderBook:   NewOptionsOrderBook(),
		riskManager: NewOptionsRiskManager(),
		nextOrderID: 1,
		nextTradeID: 1,
	}
}

// NewOptionsOrderBook creates a new options order book
func NewOptionsOrderBook() *OptionsOrderBook {
	return &OptionsOrderBook{
		buyOrders:  make(map[string]*OptionsOrder),
		sellOrders: make(map[string]*OptionsOrder),
	}
}

// NewOptionsRiskManager creates a new options risk manager
func NewOptionsRiskManager() *OptionsRiskManager {
	return &OptionsRiskManager{
		maxPositionSize: big.NewFloat(1000),  // Default max position size
		maxLoss:         big.NewFloat(10000), // Default max loss
		positionLimits:  make(map[string]*big.Float),
	}
}

// PlaceOrder places a new options order
func (oe *OptionsEngine) PlaceOrder(order *OptionsOrder) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	// Validate order
	if err := oe.validateOrder(order); err != nil {
		return err
	}

	// Check risk limits
	if err := oe.riskManager.CheckRiskLimits(order); err != nil {
		return err
	}

	// Generate order ID
	oe.mu.Lock()
	order.ID = fmt.Sprintf("OPT_%d", oe.nextOrderID)
	oe.nextOrderID++
	oe.mu.Unlock()

	// Set timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Status = Pending

	// Add to order book
	if err := oe.orderBook.AddOrder(order); err != nil {
		return err
	}

	// Store order
	oe.mu.Lock()
	oe.orders[order.ID] = order
	oe.mu.Unlock()

	// Try to match orders synchronously
	oe.matchOrders()

	return nil
}

// CancelOrder cancels an existing options order
func (oe *OptionsEngine) CancelOrder(orderID, userID string) error {
	oe.mu.RLock()
	order, exists := oe.orders[orderID]
	oe.mu.RUnlock()

	if !exists {
		return errors.New("order not found")
	}

	if order.UserID != userID {
		return errors.New("unauthorized to cancel order")
	}

	if order.Status != Pending {
		return errors.New("order cannot be cancelled")
	}

	// Update order status
	order.Status = Cancelled
	order.UpdatedAt = time.Now()

	// Remove from order book
	oe.orderBook.RemoveOrder(order)

	return nil
}

// GetOrder retrieves an options order by ID
func (oe *OptionsEngine) GetOrder(orderID string) (*OptionsOrder, error) {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	order, exists := oe.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

// GetUserOrders retrieves all orders for a specific user
func (oe *OptionsEngine) GetUserOrders(userID string) []*OptionsOrder {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	var userOrders []*OptionsOrder
	for _, order := range oe.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}

	return userOrders
}

// GetPosition retrieves a position by ID
func (oe *OptionsEngine) GetPosition(positionID string) (*Position, error) {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	position, exists := oe.positions[positionID]
	if !exists {
		return nil, errors.New("position not found")
	}

	return position, nil
}

// GetUserPositions retrieves all positions for a specific user
func (oe *OptionsEngine) GetUserPositions(userID string) []*Position {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	var userPositions []*Position
	for _, position := range oe.positions {
		if position.UserID == userID {
			userPositions = append(userPositions, position)
		}
	}

	return userPositions
}

// GetTrade retrieves a trade by ID
func (oe *OptionsEngine) GetTrade(tradeID string) (*Trade, error) {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	trade, exists := oe.trades[tradeID]
	if !exists {
		return nil, errors.New("trade not found")
	}

	return trade, nil
}

// GetUserTrades retrieves all trades for a specific user
func (oe *OptionsEngine) GetUserTrades(userID string) []*Trade {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	var userTrades []*Trade
	for _, trade := range oe.trades {
		if trade.UserID == userID || trade.CounterpartyID == userID {
			userTrades = append(userTrades, trade)
		}
	}

	return userTrades
}

// validateOrder validates an options order
func (oe *OptionsEngine) validateOrder(order *OptionsOrder) error {
	if order.Option == nil {
		return errors.New("option cannot be nil")
	}

	if order.Quantity == nil || order.Quantity.Sign() <= 0 {
		return errors.New("quantity must be positive")
	}

	if order.Price == nil || order.Price.Sign() < 0 {
		return errors.New("price must be non-negative")
	}

	if order.UserID == "" {
		return errors.New("user ID is required")
	}

	// Check if option has expired
	if order.Option.TimeToExpiry.Sign() <= 0 {
		return errors.New("option has expired")
	}

	return nil
}

// matchOrders attempts to match pending orders
func (oe *OptionsEngine) matchOrders() {
	// Get all pending orders without holding the main lock
	var pendingBuyOrders []*OptionsOrder
	var pendingSellOrders []*OptionsOrder

	oe.mu.RLock()
	for _, order := range oe.orders {
		if order.Status == Pending {
			if order.Side == Buy {
				pendingBuyOrders = append(pendingBuyOrders, order)
			} else {
				pendingSellOrders = append(pendingSellOrders, order)
			}
		}
	}
	oe.mu.RUnlock()

	// Simple matching logic (can be enhanced with more sophisticated algorithms)
	for _, buyOrder := range pendingBuyOrders {
		if buyOrder.Status != Pending {
			continue
		}

		for _, sellOrder := range pendingSellOrders {
			if sellOrder.Status != Pending {
				continue
			}

			// Check if orders can be matched
			if oe.canMatch(buyOrder, sellOrder) {
				oe.executeTrade(buyOrder, sellOrder)
				// Break after first match to avoid multiple matches
				break
			}
		}
	}
}

// canMatch checks if two orders can be matched
func (oe *OptionsEngine) canMatch(buyOrder, sellOrder *OptionsOrder) bool {
	// Check if options are the same
	if buyOrder.Option.Type != sellOrder.Option.Type ||
		buyOrder.Option.StrikePrice.Cmp(sellOrder.Option.StrikePrice) != 0 ||
		buyOrder.Option.TimeToExpiry.Cmp(sellOrder.Option.TimeToExpiry) != 0 {
		return false
	}

	// Check price compatibility
	if buyOrder.Type == Market || sellOrder.Type == Market {
		return true
	}

	// For limit orders, buy price must be >= sell price
	return buyOrder.Price.Cmp(sellOrder.Price) >= 0
}

// executeTrade executes a trade between two orders
func (oe *OptionsEngine) executeTrade(buyOrder, sellOrder *OptionsOrder) {
	// Determine execution price and quantity
	executionPrice := oe.determineExecutionPrice(buyOrder, sellOrder)
	executionQuantity := oe.determineExecutionQuantity(buyOrder, sellOrder)

	// Create trade record
	trade := &Trade{
		ID:             fmt.Sprintf("TRADE_%d", oe.nextTradeID),
		OrderID:        buyOrder.ID,
		Option:         buyOrder.Option,
		Side:           Buy,
		Quantity:       executionQuantity,
		Price:          executionPrice,
		Timestamp:      time.Now(),
		UserID:         buyOrder.UserID,
		CounterpartyID: sellOrder.UserID,
	}

	// Update trade ID counter
	oe.mu.Lock()
	oe.nextTradeID++
	oe.mu.Unlock()

	// Update orders
	buyOrder.Status = Filled
	buyOrder.FilledAt = &trade.Timestamp
	buyOrder.FilledPrice = executionPrice
	buyOrder.FilledQuantity = executionQuantity
	buyOrder.UpdatedAt = time.Now()

	sellOrder.Status = Filled
	sellOrder.FilledAt = &trade.Timestamp
	sellOrder.FilledPrice = executionPrice
	sellOrder.FilledQuantity = executionQuantity
	sellOrder.UpdatedAt = time.Now()

	// Store trade
	oe.mu.Lock()
	oe.trades[trade.ID] = trade
	oe.mu.Unlock()

	// Update positions
	oe.updatePositions(buyOrder, sellOrder, executionQuantity, executionPrice)

	// Remove filled orders from order book
	oe.orderBook.RemoveOrder(buyOrder)
	oe.orderBook.RemoveOrder(sellOrder)
}

// determineExecutionPrice determines the execution price for a trade
func (oe *OptionsEngine) determineExecutionPrice(buyOrder, sellOrder *OptionsOrder) *big.Float {
	// For market orders, use the limit order price
	// For limit orders, use the price that was placed first (price-time priority)
	if buyOrder.Type == Market {
		return new(big.Float).Copy(sellOrder.Price)
	}
	if sellOrder.Type == Market {
		return new(big.Float).Copy(buyOrder.Price)
	}

	// For limit orders, use the price that was placed first
	if buyOrder.CreatedAt.Before(sellOrder.CreatedAt) {
		return new(big.Float).Copy(buyOrder.Price)
	}
	return new(big.Float).Copy(sellOrder.Price)
}

// determineExecutionQuantity determines the execution quantity for a trade
func (oe *OptionsEngine) determineExecutionQuantity(buyOrder, sellOrder *OptionsOrder) *big.Float {
	// Use the smaller of the two quantities
	if buyOrder.Quantity.Cmp(sellOrder.Quantity) <= 0 {
		return new(big.Float).Copy(buyOrder.Quantity)
	}
	return new(big.Float).Copy(sellOrder.Quantity)
}

// updatePositions updates positions after a trade
func (oe *OptionsEngine) updatePositions(buyOrder, sellOrder *OptionsOrder, quantity, price *big.Float) {
	// Update buyer position
	buyerPositionID := fmt.Sprintf("%s_%d", buyOrder.UserID, buyOrder.Option.Type)
	
	oe.mu.Lock()
	buyerPosition, exists := oe.positions[buyerPositionID]

	if !exists {
		buyerPosition = &Position{
			ID:            buyerPositionID,
			Option:        buyOrder.Option,
			Quantity:      new(big.Float),
			AveragePrice:  new(big.Float),
			UnrealizedPnL: new(big.Float),
			RealizedPnL:   new(big.Float),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			UserID:        buyOrder.UserID,
		}
		oe.positions[buyerPositionID] = buyerPosition
	}

	// Update buyer position
	oldQuantity := new(big.Float).Copy(buyerPosition.Quantity)
	oldAveragePrice := new(big.Float).Copy(buyerPosition.AveragePrice)

	buyerPosition.Quantity.Add(buyerPosition.Quantity, quantity)

	// Calculate new average price
	totalCost := new(big.Float).Mul(price, quantity)
	oldCost := new(big.Float).Mul(oldAveragePrice, oldQuantity)
	totalCost.Add(totalCost, oldCost)
	buyerPosition.AveragePrice.Quo(totalCost, buyerPosition.Quantity)

	buyerPosition.UpdatedAt = time.Now()

	// Update seller position (short position)
	sellerPositionID := fmt.Sprintf("%s_%d", sellOrder.UserID, sellOrder.Option.Type)
	sellerPosition, exists := oe.positions[sellerPositionID]

	if !exists {
		sellerPosition = &Position{
			ID:            sellerPositionID,
			Option:        sellOrder.Option,
			Quantity:      new(big.Float),
			AveragePrice:  new(big.Float),
			UnrealizedPnL: new(big.Float),
			RealizedPnL:   new(big.Float),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			UserID:        sellOrder.UserID,
		}
		oe.positions[sellerPositionID] = sellerPosition
	}

	// Update seller position (reduce long position or increase short position)
	sellerPosition.Quantity.Sub(sellerPosition.Quantity, quantity)
	sellerPosition.UpdatedAt = time.Now()
	oe.mu.Unlock()
}

// AddOrder adds an order to the order book
func (ob *OptionsOrderBook) AddOrder(order *OptionsOrder) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == Buy {
		ob.buyOrders[order.ID] = order
	} else {
		ob.sellOrders[order.ID] = order
	}

	return nil
}

// RemoveOrder removes an order from the order book
func (ob *OptionsOrderBook) RemoveOrder(order *OptionsOrder) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Side == Buy {
		delete(ob.buyOrders, order.ID)
	} else {
		delete(ob.sellOrders, order.ID)
	}
}

// GetOrderBook returns the current order book
func (ob *OptionsOrderBook) GetOrderBook() ([]*OptionsOrder, []*OptionsOrder) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	buyOrders := make([]*OptionsOrder, 0, len(ob.buyOrders))
	sellOrders := make([]*OptionsOrder, 0, len(ob.sellOrders))

	for _, order := range ob.buyOrders {
		buyOrders = append(buyOrders, order)
	}

	for _, order := range ob.sellOrders {
		sellOrders = append(sellOrders, order)
	}

	return buyOrders, sellOrders
}

// CheckRiskLimits checks if an order violates risk limits
func (rm *OptionsRiskManager) CheckRiskLimits(order *OptionsOrder) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check position size limits
	if order.Quantity.Cmp(rm.maxPositionSize) > 0 {
		return errors.New("order quantity exceeds maximum position size")
	}

	// Check user-specific position limits
	if userLimit, exists := rm.positionLimits[order.UserID]; exists {
		if order.Quantity.Cmp(userLimit) > 0 {
			return errors.New("order quantity exceeds user position limit")
		}
	}

	return nil
}

// SetMaxPositionSize sets the maximum position size
func (rm *OptionsRiskManager) SetMaxPositionSize(size *big.Float) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.maxPositionSize = new(big.Float).Copy(size)
}

// SetMaxLoss sets the maximum loss limit
func (rm *OptionsRiskManager) SetMaxLoss(loss *big.Float) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.maxLoss = new(big.Float).Copy(loss)
}

// SetUserPositionLimit sets a position limit for a specific user
func (rm *OptionsRiskManager) SetUserPositionLimit(userID string, limit *big.Float) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.positionLimits[userID] = new(big.Float).Copy(limit)
}
