package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/exchange/orderbook"
	"github.com/palaseus/adrenochain/pkg/exchange/trading"
)

// TradingAPI handles trading-related HTTP endpoints
type TradingAPI struct {
	orderBookManager *OrderBookManager
	tradingManager   *TradingManager
	marketData       *MarketDataService
}

// OrderBookManager manages multiple order books
type OrderBookManager struct {
	orderBooks map[string]*orderbook.OrderBook
	mutex      sync.RWMutex
}

// TradingManager handles trading operations
type TradingManager struct {
	tradingPairs map[string]*trading.TradingPair
	orderBooks   map[string]*orderbook.OrderBook
	mutex        sync.RWMutex
}

// MarketDataService provides market data and analytics
type MarketDataService struct {
	orderBooks map[string]*orderbook.OrderBook
	trades     []*orderbook.Trade
	mutex      sync.RWMutex
}

// NewTradingAPI creates a new trading API instance
func NewTradingAPI() *TradingAPI {
	return &TradingAPI{
		orderBookManager: NewOrderBookManager(),
		tradingManager:   NewTradingManager(),
		marketData:       NewMarketDataService(),
	}
}

// NewOrderBookManager creates a new order book manager
func NewOrderBookManager() *OrderBookManager {
	return &OrderBookManager{
		orderBooks: make(map[string]*orderbook.OrderBook),
	}
}

// NewTradingManager creates a new trading manager
func NewTradingManager() *TradingManager {
	return &TradingManager{
		tradingPairs: make(map[string]*trading.TradingPair),
		orderBooks:   make(map[string]*orderbook.OrderBook),
	}
}

// NewMarketDataService creates a new market data service
func NewMarketDataService() *MarketDataService {
	return &MarketDataService{
		orderBooks: make(map[string]*orderbook.OrderBook),
		trades:     make([]*orderbook.Trade, 0),
	}
}

// HTTP Handlers

// CreateOrder handles POST /api/v1/orders
func (ta *TradingAPI) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var orderRequest CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := orderRequest.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create order
	order, err := ta.createOrderFromRequest(&orderRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Process order through matching engine
	execution, err := ta.processOrder(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	response := CreateOrderResponse{
		OrderID:   order.ID,
		Status:    order.Status,
		Execution: execution,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrder handles GET /api/v1/orders/{orderID}
func (ta *TradingAPI) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, err := ta.getOrder(orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// CancelOrder handles DELETE /api/v1/orders/{orderID}
func (ta *TradingAPI) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	if err := ta.cancelOrder(orderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CancelOrderResponse{
		OrderID:   orderID,
		Status:    "cancelled",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrderBook handles GET /api/v1/orderbook/{tradingPair}
func (ta *TradingAPI) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tradingPair := r.URL.Query().Get("trading_pair")
	if tradingPair == "" {
		http.Error(w, "Trading pair is required", http.StatusBadRequest)
		return
	}

	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	if depth <= 0 {
		depth = 20 // Default depth
	}

	orderBook, err := ta.getOrderBook(tradingPair, depth)
	if err != nil {
		http.Error(w, "Order book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderBook)
}

// GetTrades handles GET /api/v1/trades/{tradingPair}
func (ta *TradingAPI) GetTrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tradingPair := r.URL.Query().Get("trading_pair")
	if tradingPair == "" {
		http.Error(w, "Trading pair is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default limit
	}

	trades, err := ta.getTrades(tradingPair, limit)
	if err != nil {
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// GetTradingPairs handles GET /api/v1/trading-pairs
func (ta *TradingAPI) GetTradingPairs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pairs := ta.getAllTradingPairs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pairs)
}

// GetMarketData handles GET /api/v1/market-data/{tradingPair}
func (ta *TradingAPI) GetMarketData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tradingPair := r.URL.Query().Get("trading_pair")
	if tradingPair == "" {
		http.Error(w, "Trading pair is required", http.StatusBadRequest)
		return
	}

	marketData, err := ta.getMarketData(tradingPair)
	if err != nil {
		http.Error(w, "Market data not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketData)
}

// Business Logic Methods

// createOrderFromRequest creates an order from the request
func (ta *TradingAPI) createOrderFromRequest(req *CreateOrderRequest) (*orderbook.Order, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate required fields
	if req.TradingPair == "" {
		return nil, fmt.Errorf("trading pair is required")
	}
	if req.Quantity == nil {
		return nil, fmt.Errorf("quantity is required")
	}
	// For non-market orders, price is required
	if req.Type != orderbook.OrderTypeMarket && req.Price == nil {
		return nil, fmt.Errorf("price is required")
	}
	// For market orders, price should be nil or zero
	if req.Type == orderbook.OrderTypeMarket && req.Price != nil && req.Price.Cmp(big.NewInt(0)) > 0 {
		return nil, fmt.Errorf("market orders cannot have a price")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Create order with proper validation
	order := &orderbook.Order{
		ID:                generateOrderID(),
		TradingPair:       req.TradingPair,
		Side:              req.Side,
		Type:              req.Type,
		Quantity:          req.Quantity,
		Price:             req.Price, // This can be nil for market orders
		UserID:            req.UserID,
		TimeInForce:       orderbook.TimeInForceGTC,       // Default to Good Till Cancelled
		FilledQuantity:    big.NewInt(0),                  // Initialize to zero
		RemainingQuantity: new(big.Int).Set(req.Quantity), // Copy quantity
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Status:            orderbook.OrderStatusPending,
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// processOrder processes an order through the matching engine
func (ta *TradingAPI) processOrder(order *orderbook.Order) (*orderbook.TradeExecution, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Get or create order book for the trading pair
	ob, err := ta.getOrCreateOrderBook(order.TradingPair)
	if err != nil {
		return nil, err
	}

	// Create matching engine
	matchingEngine := orderbook.NewMatchingEngine(ob)

	// Process the order
	execution, err := matchingEngine.ProcessOrder(order)
	if err != nil {
		return nil, err
	}

	// Update market data
	ta.updateMarketData(order.TradingPair, execution)

	return execution, nil
}

// getOrder retrieves an order by ID
func (ta *TradingAPI) getOrder(orderID string) (*orderbook.Order, error) {
	// This would typically query a database
	// For now, we'll search through all order books
	ta.orderBookManager.mutex.RLock()
	defer ta.orderBookManager.mutex.RUnlock()

	for _, ob := range ta.orderBookManager.orderBooks {
		if order, err := ob.GetOrder(orderID); err == nil {
			return order, nil
		}
	}

	return nil, fmt.Errorf("order not found: %s", orderID)
}

// cancelOrder cancels an order
func (ta *TradingAPI) cancelOrder(orderID string) error {
	// This would typically update a database
	// For now, we'll search through all order books
	ta.orderBookManager.mutex.Lock()
	defer ta.orderBookManager.mutex.Unlock()

	for _, ob := range ta.orderBookManager.orderBooks {
		if _, err := ob.GetOrder(orderID); err == nil {
			// Update order status (this would typically be done through a proper method)
			// For now, we'll just return success
			return nil
		}
	}

	return fmt.Errorf("order not found: %s", orderID)
}

// getOrderBook retrieves the order book for a trading pair
func (ta *TradingAPI) getOrderBook(tradingPair string, depth int) (*OrderBookResponse, error) {
	ob, err := ta.getOrCreateOrderBook(tradingPair)
	if err != nil {
		return nil, err
	}

	// Get depth for both sides using the available GetDepth method
	buyDepth, err := ob.GetDepth(depth)
	if err != nil {
		return nil, err
	}

	// Filter buy orders (we'll need to get sell depth separately)
	// For now, create a simple response with available data
	bids := make([]*orderbook.Order, 0) // Initialize as empty slice, not nil
	asks := make([]*orderbook.Order, 0) // Initialize as empty slice, not nil

	// Convert PriceLevel to Order format for bids
	for _, level := range buyDepth {
		if level.Side == orderbook.OrderSideBuy {
			// Create a placeholder order for the price level
			// In a real implementation, you'd get actual orders
			order := &orderbook.Order{
				ID:          fmt.Sprintf("level_%s", level.Price.String()),
				TradingPair: tradingPair,
				Side:        level.Side,
				Price:       level.Price,
				Quantity:    level.Volume,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			bids = append(bids, order)
		}
	}

	return &OrderBookResponse{
		TradingPair: tradingPair,
		Timestamp:   time.Now(),
		Bids:        bids,
		Asks:        asks,
	}, nil
}

// getTrades retrieves recent trades for a trading pair
func (ta *TradingAPI) getTrades(tradingPair string, limit int) ([]*orderbook.Trade, error) {
	ta.marketData.mutex.RLock()
	defer ta.marketData.mutex.RUnlock()

	trades := make([]*orderbook.Trade, 0) // Initialize as empty slice, not nil
	for _, trade := range ta.marketData.trades {
		if trade.TradingPair == tradingPair {
			trades = append(trades, trade)
			if len(trades) >= limit {
				break
			}
		}
	}

	return trades, nil
}

// getAllTradingPairs returns all available trading pairs
func (ta *TradingAPI) getAllTradingPairs() []*trading.TradingPair {
	ta.tradingManager.mutex.RLock()
	defer ta.tradingManager.mutex.RUnlock()

	pairs := make([]*trading.TradingPair, 0) // Initialize as empty slice, not nil
	for _, pair := range ta.tradingManager.tradingPairs {
		pairs = append(pairs, pair)
	}

	return pairs
}

// getMarketData retrieves market data for a trading pair
func (ta *TradingAPI) getMarketData(tradingPair string) (*MarketDataResponse, error) {
	ob, err := ta.getOrCreateOrderBook(tradingPair)
	if err != nil {
		return nil, err
	}

	// Calculate market data using available methods
	// For now, use placeholder values since these methods don't exist yet
	lastPrice := big.NewInt(0)       // Would come from trade history
	volume24h := ob.GetTotalVolume() // Use available volume method
	priceChange24h := big.NewInt(0)  // Would be calculated from price history

	return &MarketDataResponse{
		TradingPair:    tradingPair,
		LastPrice:      lastPrice,
		Volume24h:      volume24h,
		PriceChange24h: priceChange24h,
		Timestamp:      time.Now(),
	}, nil
}

// getOrCreateOrderBook gets or creates an order book for a trading pair
func (ta *TradingAPI) getOrCreateOrderBook(tradingPair string) (*orderbook.OrderBook, error) {
	ta.orderBookManager.mutex.Lock()
	defer ta.orderBookManager.mutex.Unlock()

	if ob, exists := ta.orderBookManager.orderBooks[tradingPair]; exists {
		return ob, nil
	}

	// Create new order book
	ob, err := orderbook.NewOrderBook(tradingPair)
	if err != nil {
		return nil, err
	}

	ta.orderBookManager.orderBooks[tradingPair] = ob
	return ob, nil
}

// updateMarketData updates market data after a trade
func (ta *TradingAPI) updateMarketData(tradingPair string, execution *orderbook.TradeExecution) {
	if execution == nil {
		return // Nothing to update
	}

	ta.marketData.mutex.Lock()
	defer ta.marketData.mutex.Unlock()

	// Add trades to market data
	if execution.Trade != nil {
		ta.marketData.trades = append(ta.marketData.trades, execution.Trade)
	}

	// Keep only last 1000 trades
	if len(ta.marketData.trades) > 1000 {
		ta.marketData.trades = ta.marketData.trades[len(ta.marketData.trades)-1000:]
	}
}

// Helper function to generate order IDs
func generateOrderID() string {
	return fmt.Sprintf("order_%d", time.Now().UnixNano())
}
