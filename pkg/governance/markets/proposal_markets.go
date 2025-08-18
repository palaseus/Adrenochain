package markets

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MarketType represents different types of proposal markets
type MarketType string

const (
	BinaryMarket       MarketType = "binary"      // Yes/No outcomes
	MultiOutcomeMarket MarketType = "multi"       // Multiple possible outcomes
	ScalarMarket       MarketType = "scalar"      // Numeric range outcomes
	FuturesMarket      MarketType = "futures"     // Time-based outcomes
	ConditionalMarket  MarketType = "conditional" // Conditional on other events
	CustomMarket       MarketType = "custom"      // Custom market structure
)

// MarketStatus represents the current status of a proposal market
type MarketStatus string

const (
	MarketDraft     MarketStatus = "draft"     // Market is being created
	MarketActive    MarketStatus = "active"    // Market is open for trading
	MarketSuspended MarketStatus = "suspended" // Market is temporarily suspended
	MarketClosed    MarketStatus = "closed"    // Market is closed for trading
	MarketResolved  MarketStatus = "resolved"  // Market outcome is determined
	MarketCancelled MarketStatus = "cancelled" // Market was cancelled
)

// Outcome represents a possible outcome in a proposal market
type Outcome struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Probability float64   `json:"probability"` // Current market probability
	Price       float64   `json:"price"`       // Current market price
	Volume      float64   `json:"volume"`      // Trading volume
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Market represents a proposal market for governance
type Market struct {
	ID                string                 `json:"id"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	MarketType        MarketType             `json:"market_type"`
	Status            MarketStatus           `json:"status"`
	Creator           string                 `json:"creator"`
	Outcomes          []*Outcome             `json:"outcomes"`
	TotalVolume       float64                `json:"total_volume"`
	TotalParticipants int                    `json:"total_participants"`
	CreationTime      time.Time              `json:"creation_time"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	ResolutionTime    *time.Time             `json:"resolution_time,omitempty"`
	ResolvedOutcome   *string                `json:"resolved_outcome,omitempty"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Order represents a trading order in the proposal market
type Order struct {
	ID        string      `json:"id"`
	MarketID  string      `json:"market_id"`
	UserID    string      `json:"user_id"`
	OutcomeID string      `json:"outcome_id"`
	Type      OrderType   `json:"type"`
	Amount    float64     `json:"amount"`
	Price     float64     `json:"price"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// OrderType represents the type of trading order
type OrderType string

const (
	BuyOrder  OrderType = "buy"
	SellOrder OrderType = "sell"
)

// OrderStatus represents the status of a trading order
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
	OrderExpired   OrderStatus = "expired"
)

// Position represents a user's position in a market outcome
type Position struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	MarketID  string    `json:"market_id"`
	OutcomeID string    `json:"outcome_id"`
	Amount    float64   `json:"amount"`
	AvgPrice  float64   `json:"avg_price"`
	Pnl       float64   `json:"pnl"` // Profit/Loss
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MarketMetrics tracks performance and trading metrics
type MarketMetrics struct {
	TotalTrades     int64     `json:"total_trades"`
	TotalVolume     float64   `json:"total_volume"`
	UniqueTraders   int       `json:"unique_traders"`
	PriceVolatility float64   `json:"price_volatility"`
	LiquidityIndex  float64   `json:"liquidity_index"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ProposalMarkets manages proposal markets and trading
type ProposalMarkets struct {
	markets    map[string]*Market
	orders     map[string]*Order
	positions  map[string]*Position
	metrics    map[string]*MarketMetrics
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	orderQueue chan *Order
}

// NewProposalMarkets creates a new ProposalMarkets instance
func NewProposalMarkets() *ProposalMarkets {
	ctx, cancel := context.WithCancel(context.Background())
	pm := &ProposalMarkets{
		markets:    make(map[string]*Market),
		orders:     make(map[string]*Order),
		positions:  make(map[string]*Position),
		metrics:    make(map[string]*MarketMetrics),
		ctx:        ctx,
		cancel:     cancel,
		orderQueue: make(chan *Order, 1000),
	}

	// Start background processing
	go pm.processOrders()
	go pm.updateMetrics()
	go pm.cleanupExpiredOrders()

	return pm
}

// CreateMarket creates a new proposal market
func (pm *ProposalMarkets) CreateMarket(market *Market) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if market.ID == "" {
		market.ID = uuid.New().String()
	}

	if market.Creator == "" {
		return fmt.Errorf("market creator is required")
	}

	if len(market.Outcomes) == 0 {
		return fmt.Errorf("market must have at least one outcome")
	}

	if market.StartTime.Before(time.Now()) {
		return fmt.Errorf("market start time must be in the future")
	}

	if market.EndTime.Before(market.StartTime) {
		return fmt.Errorf("market end time must be after start time")
	}

	market.CreatedAt = time.Now()
	market.UpdatedAt = time.Now()
	market.Status = MarketDraft

	// Initialize outcomes
	for _, outcome := range market.Outcomes {
		outcome.ID = uuid.New().String()
		outcome.CreatedAt = time.Now()
		outcome.UpdatedAt = time.Now()
		outcome.Probability = 1.0 / float64(len(market.Outcomes))
		outcome.Price = 1.0 / float64(len(market.Outcomes))
	}

	pm.markets[market.ID] = market
	pm.metrics[market.ID] = &MarketMetrics{
		LastUpdated: time.Now(),
	}

	return nil
}

// ActivateMarket activates a market for trading
func (pm *ProposalMarkets) ActivateMarket(marketID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	market, exists := pm.markets[marketID]
	if !exists {
		return fmt.Errorf("market not found")
	}

	if market.Status != MarketDraft {
		return fmt.Errorf("market must be in draft status to activate")
	}

	if time.Now().Before(market.StartTime) {
		return fmt.Errorf("market start time has not been reached")
	}

	market.Status = MarketActive
	market.UpdatedAt = time.Now()

	return nil
}

// PlaceOrder places a trading order in the market
func (pm *ProposalMarkets) PlaceOrder(order *Order) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	if order.MarketID == "" || order.UserID == "" || order.OutcomeID == "" {
		return fmt.Errorf("market ID, user ID, and outcome ID are required")
	}

	market, exists := pm.markets[order.MarketID]
	if !exists {
		return fmt.Errorf("market not found")
	}

	if market.Status != MarketActive {
		return fmt.Errorf("market is not active for trading")
	}

	if order.Amount <= 0 || order.Price <= 0 {
		return fmt.Errorf("amount and price must be positive")
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = OrderPending

	pm.orders[order.ID] = order

	// Send to order queue for processing
	select {
	case pm.orderQueue <- order:
	default:
		// Queue is full, process immediately
		go pm.processOrder(order)
	}

	return nil
}

// GetMarket retrieves a market by ID
func (pm *ProposalMarkets) GetMarket(marketID string) (*Market, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	market, exists := pm.markets[marketID]
	if !exists {
		return nil, fmt.Errorf("market not found")
	}

	return market, nil
}

// GetMarkets retrieves all markets with optional filtering
func (pm *ProposalMarkets) GetMarkets(status MarketStatus, marketType MarketType) []*Market {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var markets []*Market
	for _, market := range pm.markets {
		if status != "" && market.Status != status {
			continue
		}
		if marketType != "" && market.MarketType != marketType {
			continue
		}
		markets = append(markets, market)
	}

	return markets
}

// GetUserPositions retrieves all positions for a user
func (pm *ProposalMarkets) GetUserPositions(userID string) []*Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var positions []*Position
	for _, position := range pm.positions {
		if position.UserID == userID {
			positions = append(positions, position)
		}
	}

	return positions
}

// ResolveMarket resolves a market with a specific outcome
func (pm *ProposalMarkets) ResolveMarket(marketID string, outcomeID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	market, exists := pm.markets[marketID]
	if !exists {
		return fmt.Errorf("market not found")
	}

	if market.Status != MarketActive && market.Status != MarketClosed {
		return fmt.Errorf("market must be active or closed to resolve")
	}

	// Verify outcome exists
	outcomeExists := false
	for _, outcome := range market.Outcomes {
		if outcome.ID == outcomeID {
			outcomeExists = true
			break
		}
	}

	if !outcomeExists {
		return fmt.Errorf("outcome not found")
	}

	market.Status = MarketResolved
	market.ResolvedOutcome = &outcomeID
	now := time.Now()
	market.ResolutionTime = &now
	market.UpdatedAt = now

	// Calculate final PnL for all positions
	pm.calculateFinalPnL(marketID, outcomeID)

	return nil
}

// CloseMarket closes a market for trading
func (pm *ProposalMarkets) CloseMarket(marketID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	market, exists := pm.markets[marketID]
	if !exists {
		return fmt.Errorf("market not found")
	}

	if market.Status != MarketActive {
		return fmt.Errorf("market is not active")
	}

	if time.Now().Before(market.EndTime) {
		return fmt.Errorf("market end time has not been reached")
	}

	market.Status = MarketClosed
	market.UpdatedAt = time.Now()

	return nil
}

// processOrders processes orders from the queue
func (pm *ProposalMarkets) processOrders() {
	for {
		select {
		case order := <-pm.orderQueue:
			pm.processOrder(order)
		case <-pm.ctx.Done():
			return
		}
	}
}

// processOrder processes a single order
func (pm *ProposalMarkets) processOrder(order *Order) {
	if order == nil {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Find matching orders
	matchingOrders := pm.findMatchingOrders(order)
	if len(matchingOrders) == 0 {
		return
	}

	// Execute trades
	for _, matchingOrder := range matchingOrders {
		if order.Status != OrderPending || matchingOrder.Status != OrderPending {
			continue
		}

		tradeAmount := pm.min(order.Amount, matchingOrder.Amount)
		tradePrice := (order.Price + matchingOrder.Price) / 2

		// Execute trade
		pm.executeTrade(order, matchingOrder, tradeAmount, tradePrice)

		// Update order amounts
		order.Amount -= tradeAmount
		matchingOrder.Amount -= tradeAmount

		// Mark orders as filled if completely executed
		if order.Amount <= 0 {
			order.Status = OrderFilled
			order.UpdatedAt = time.Now()
		}
		if matchingOrder.Amount <= 0 {
			matchingOrder.Status = OrderFilled
			matchingOrder.UpdatedAt = time.Now()
		}
	}
}

// findMatchingOrders finds orders that can be matched with the given order
func (pm *ProposalMarkets) findMatchingOrders(order *Order) []*Order {
	var matchingOrders []*Order

	for _, existingOrder := range pm.orders {
		if existingOrder.ID == order.ID {
			continue
		}

		if existingOrder.MarketID != order.MarketID || existingOrder.OutcomeID != order.OutcomeID {
			continue
		}

		if existingOrder.Status != OrderPending {
			continue
		}

		// Check if orders can be matched
		if (order.Type == BuyOrder && existingOrder.Type == SellOrder && order.Price >= existingOrder.Price) ||
			(order.Type == SellOrder && existingOrder.Type == BuyOrder && order.Price <= existingOrder.Price) {
			matchingOrders = append(matchingOrders, existingOrder)
		}
	}

	return matchingOrders
}

// executeTrade executes a trade between two orders
func (pm *ProposalMarkets) executeTrade(order1, order2 *Order, amount, price float64) {
	// Update positions
	pm.updatePosition(order1.UserID, order1.MarketID, order1.OutcomeID, amount, price, order1.Type)
	pm.updatePosition(order2.UserID, order2.MarketID, order2.OutcomeID, amount, price, order2.Type)

	// Update market metrics
	pm.updateMarketMetrics(order1.MarketID, amount, price)
}

// updatePosition updates a user's position after a trade
func (pm *ProposalMarkets) updatePosition(userID, marketID, outcomeID string, amount, price float64, orderType OrderType) {
	positionKey := fmt.Sprintf("%s:%s:%s", userID, marketID, outcomeID)
	position, exists := pm.positions[positionKey]

	if !exists {
		position = &Position{
			ID:        uuid.New().String(),
			UserID:    userID,
			MarketID:  marketID,
			OutcomeID: outcomeID,
			CreatedAt: time.Now(),
		}
		pm.positions[positionKey] = position
	}

	if orderType == BuyOrder {
		// Buying increases position
		totalCost := position.Amount*position.AvgPrice + amount*price
		position.Amount += amount
		position.AvgPrice = totalCost / position.Amount
	} else {
		// Selling decreases position
		position.Amount -= amount
		if position.Amount < 0 {
			position.Amount = 0
		}
	}

	position.UpdatedAt = time.Now()
}

// updateMarketMetrics updates market metrics after a trade
func (pm *ProposalMarkets) updateMarketMetrics(marketID string, amount, price float64) {
	metrics, exists := pm.metrics[marketID]
	if !exists {
		return
	}

	metrics.TotalTrades++
	metrics.TotalVolume += amount * price
	metrics.LastUpdated = time.Now()

	// Update outcome prices based on trading activity
	pm.updateOutcomePrices(marketID)
}

// updateOutcomePrices updates outcome prices based on trading activity
func (pm *ProposalMarkets) updateOutcomePrices(marketID string) {
	market, exists := pm.markets[marketID]
	if !exists {
		return
	}

	// Simple price update based on volume
	for _, outcome := range market.Outcomes {
		// Calculate new price based on volume and market activity
		volumeWeight := outcome.Volume / market.TotalVolume
		if volumeWeight > 0 {
			outcome.Price = outcome.Price*0.9 + volumeWeight*0.1
			outcome.Probability = outcome.Price
			outcome.UpdatedAt = time.Now()
		}
	}
}

// calculateFinalPnL calculates final profit/loss for all positions when market is resolved
func (pm *ProposalMarkets) calculateFinalPnL(marketID, resolvedOutcomeID string) {
	for _, position := range pm.positions {
		if position.MarketID != marketID {
			continue
		}

		if position.OutcomeID == resolvedOutcomeID {
			// Winning outcome - calculate profit
			position.Pnl = position.Amount * (1.0 - position.AvgPrice)
		} else {
			// Losing outcome - calculate loss
			position.Pnl = -position.Amount * position.AvgPrice
		}

		position.UpdatedAt = time.Now()
	}
}

// updateMetrics updates market metrics periodically
func (pm *ProposalMarkets) updateMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.mu.Lock()
			for marketID, metrics := range pm.metrics {
				_, exists := pm.markets[marketID]
				if !exists {
					continue
				}

				// Update liquidity index
				metrics.LiquidityIndex = pm.calculateLiquidityIndex(marketID)
				metrics.LastUpdated = time.Now()
			}
			pm.mu.Unlock()
		case <-pm.ctx.Done():
			return
		}
	}
}

// calculateLiquidityIndex calculates the liquidity index for a market
func (pm *ProposalMarkets) calculateLiquidityIndex(marketID string) float64 {
	market, exists := pm.markets[marketID]
	if !exists {
		return 0
	}

	totalLiquidity := 0.0
	for _, outcome := range market.Outcomes {
		totalLiquidity += outcome.Volume
	}

	if market.TotalVolume > 0 {
		return totalLiquidity / market.TotalVolume
	}
	return 0
}

// cleanupExpiredOrders removes expired orders
func (pm *ProposalMarkets) cleanupExpiredOrders() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.mu.Lock()
			now := time.Now()
			for _, order := range pm.orders {
				if order.Status == OrderPending && now.Sub(order.CreatedAt) > 24*time.Hour {
					order.Status = OrderExpired
					order.UpdatedAt = now
				}
			}
			pm.mu.Unlock()
		case <-pm.ctx.Done():
			return
		}
	}
}

// min returns the minimum of two float64 values
func (pm *ProposalMarkets) min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Close shuts down the ProposalMarkets instance
func (pm *ProposalMarkets) Close() error {
	pm.cancel()
	close(pm.orderQueue)
	return nil
}

// GetRandomID generates a random ID for testing
func (pm *ProposalMarkets) GetRandomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
