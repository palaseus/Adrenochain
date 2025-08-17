package trading

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// StrategyType represents the type of trading strategy
type StrategyType int

const (
	MarketMaking StrategyType = iota
	Arbitrage
	TrendFollowing
	MeanReversion
	StatisticalArbitrage
)

// Strategy represents a trading strategy
type Strategy struct {
	ID          string
	Type        StrategyType
	Name        string
	Description string
	Parameters  map[string]*big.Float
	Status      StrategyStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StrategyStatus represents the status of a strategy
type StrategyStatus int

const (
	Inactive StrategyStatus = iota
	Active
	Paused
	Stopped
)

// MarketMaker represents a market making strategy
type MarketMaker struct {
	Strategy
	SpreadPercentage   *big.Float
	InventoryTarget    *big.Float
	MaxInventory       *big.Float
	MinOrderSize       *big.Float
	MaxOrderSize       *big.Float
	RebalanceThreshold *big.Float
	LastRebalance      time.Time
}

// ArbitrageStrategy represents an arbitrage strategy
type ArbitrageStrategy struct {
	Strategy
	MinProfitThreshold  *big.Float
	MaxSlippage         *big.Float
	ExecutionDelay      time.Duration
	MaxConcurrentTrades int
}

// Trade represents a completed trade
type Trade struct {
	ID         string
	StrategyID string
	AssetID    string
	Side       TradeSide
	Size       *big.Float
	Price      *big.Float
	Value      *big.Float
	Timestamp  time.Time
	Status     TradeStatus
	Fee        *big.Float
}

// TradeSide represents the side of a trade
type TradeSide int

const (
	Buy TradeSide = iota
	Sell
)

// TradeStatus represents the status of a trade
type TradeStatus int

const (
	Pending TradeStatus = iota
	Executed
	Failed
	TradeCancelled
)

// Order represents a trading order
type Order struct {
	ID         string
	StrategyID string
	AssetID    string
	Side       TradeSide
	Size       *big.Float
	Price      *big.Float
	OrderType  OrderType
	Status     OrderStatus
	Timestamp  time.Time
	Expiry     time.Time
	UpdatedAt  time.Time
}

// OrderType represents the type of order
type OrderType int

const (
	Market OrderType = iota
	Limit
	Stop
	StopLimit
)

// OrderStatus represents the status of an order
type OrderStatus int

const (
	New OrderStatus = iota
	Submitted
	PartiallyFilled
	Filled
	Cancelled
	Rejected
)

// MarketData represents real-time market data
type MarketData struct {
	AssetID   string
	BidPrice  *big.Float
	AskPrice  *big.Float
	LastPrice *big.Float
	Volume    *big.Float
	Timestamp time.Time
	BidSize   *big.Float
	AskSize   *big.Float
}

// TradingEngine handles strategy execution and order management
type TradingEngine struct {
	Strategies map[string]*Strategy
	Orders     map[string]*Order
	Trades     map[string]*Trade
	MarketData map[string]*MarketData
	mu         sync.RWMutex
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewTradingEngine creates a new trading engine
func NewTradingEngine() *TradingEngine {
	now := time.Now()

	return &TradingEngine{
		Strategies: make(map[string]*Strategy),
		Orders:     make(map[string]*Order),
		Trades:     make(map[string]*Trade),
		MarketData: make(map[string]*MarketData),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewStrategy creates a new trading strategy
func NewStrategy(id, name, description string, strategyType StrategyType, parameters map[string]*big.Float) (*Strategy, error) {
	if id == "" {
		return nil, errors.New("strategy ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("strategy name cannot be empty")
	}
	if parameters == nil {
		parameters = make(map[string]*big.Float)
	}

	now := time.Now()

	return &Strategy{
		ID:          id,
		Type:        strategyType,
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Status:      Inactive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewMarketMaker creates a new market making strategy
func NewMarketMaker(id, name, description string, spreadPercentage, inventoryTarget, maxInventory, minOrderSize, maxOrderSize, rebalanceThreshold *big.Float) (*MarketMaker, error) {
	if spreadPercentage == nil || spreadPercentage.Sign() <= 0 {
		return nil, errors.New("spread percentage must be positive")
	}
	if inventoryTarget == nil {
		return nil, errors.New("inventory target cannot be nil")
	}
	if maxInventory == nil || maxInventory.Sign() <= 0 {
		return nil, errors.New("max inventory must be positive")
	}
	if minOrderSize == nil || minOrderSize.Sign() <= 0 {
		return nil, errors.New("min order size must be positive")
	}
	if maxOrderSize == nil || maxOrderSize.Sign() <= 0 {
		return nil, errors.New("max order size must be positive")
	}
	if rebalanceThreshold == nil || rebalanceThreshold.Sign() <= 0 {
		return nil, errors.New("rebalance threshold must be positive")
	}

	parameters := map[string]*big.Float{
		"spreadPercentage":   spreadPercentage,
		"inventoryTarget":    inventoryTarget,
		"maxInventory":       maxInventory,
		"minOrderSize":       minOrderSize,
		"maxOrderSize":       maxOrderSize,
		"rebalanceThreshold": rebalanceThreshold,
	}

	strategy, err := NewStrategy(id, name, description, MarketMaking, parameters)
	if err != nil {
		return nil, err
	}

	return &MarketMaker{
		Strategy:           *strategy,
		SpreadPercentage:   new(big.Float).Copy(spreadPercentage),
		InventoryTarget:    new(big.Float).Copy(inventoryTarget),
		MaxInventory:       new(big.Float).Copy(maxInventory),
		MinOrderSize:       new(big.Float).Copy(minOrderSize),
		MaxOrderSize:       new(big.Float).Copy(maxOrderSize),
		RebalanceThreshold: new(big.Float).Copy(rebalanceThreshold),
		LastRebalance:      time.Now(),
	}, nil
}

// NewArbitrageStrategy creates a new arbitrage strategy
func NewArbitrageStrategy(id, name, description string, minProfitThreshold, maxSlippage *big.Float, executionDelay time.Duration, maxConcurrentTrades int) (*ArbitrageStrategy, error) {
	if minProfitThreshold == nil || minProfitThreshold.Sign() <= 0 {
		return nil, errors.New("min profit threshold must be positive")
	}
	if maxSlippage == nil || maxSlippage.Sign() <= 0 {
		return nil, errors.New("max slippage must be positive")
	}
	if executionDelay < 0 {
		return nil, errors.New("execution delay cannot be negative")
	}
	if maxConcurrentTrades <= 0 {
		return nil, errors.New("max concurrent trades must be positive")
	}

	parameters := map[string]*big.Float{
		"minProfitThreshold":  minProfitThreshold,
		"maxSlippage":         maxSlippage,
		"executionDelay":      big.NewFloat(float64(executionDelay.Milliseconds())),
		"maxConcurrentTrades": big.NewFloat(float64(maxConcurrentTrades)),
	}

	strategy, err := NewStrategy(id, name, description, Arbitrage, parameters)
	if err != nil {
		return nil, err
	}

	return &ArbitrageStrategy{
		Strategy:            *strategy,
		MinProfitThreshold:  new(big.Float).Copy(minProfitThreshold),
		MaxSlippage:         new(big.Float).Copy(maxSlippage),
		ExecutionDelay:      executionDelay,
		MaxConcurrentTrades: maxConcurrentTrades,
	}, nil
}

// NewOrder creates a new trading order
func NewOrder(id, strategyID, assetID string, side TradeSide, size, price *big.Float, orderType OrderType) (*Order, error) {
	if id == "" {
		return nil, errors.New("order ID cannot be empty")
	}
	if strategyID == "" {
		return nil, errors.New("strategy ID cannot be empty")
	}
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if size == nil || size.Sign() <= 0 {
		return nil, errors.New("order size must be positive")
	}
	if price == nil || price.Sign() <= 0 {
		return nil, errors.New("order price must be positive")
	}

	now := time.Now()

	return &Order{
		ID:         id,
		StrategyID: strategyID,
		AssetID:    assetID,
		Side:       side,
		Size:       new(big.Float).Copy(size),
		Price:      new(big.Float).Copy(price),
		OrderType:  orderType,
		Status:     New,
		Timestamp:  now,
		Expiry:     now.Add(24 * time.Hour), // Default 24-hour expiry
		UpdatedAt:  now,
	}, nil
}

// AddStrategy adds a strategy to the trading engine
func (te *TradingEngine) AddStrategy(strategy *Strategy) error {
	if strategy == nil {
		return errors.New("strategy cannot be nil")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	if _, exists := te.Strategies[strategy.ID]; exists {
		return errors.New("strategy with this ID already exists")
	}

	te.Strategies[strategy.ID] = strategy
	te.UpdatedAt = time.Now()

	return nil
}

// RemoveStrategy removes a strategy from the trading engine
func (te *TradingEngine) RemoveStrategy(strategyID string) error {
	if strategyID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	if _, exists := te.Strategies[strategyID]; !exists {
		return errors.New("strategy not found")
	}

	delete(te.Strategies, strategyID)
	te.UpdatedAt = time.Now()

	return nil
}

// StartStrategy starts a trading strategy
func (te *TradingEngine) StartStrategy(strategyID string) error {
	if strategyID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	strategy, exists := te.Strategies[strategyID]
	if !exists {
		return errors.New("strategy not found")
	}

	strategy.Status = Active
	strategy.UpdatedAt = time.Now()
	te.UpdatedAt = time.Now()

	return nil
}

// StopStrategy stops a trading strategy
func (te *TradingEngine) StopStrategy(strategyID string) error {
	if strategyID == "" {
		return errors.New("strategy ID cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	strategy, exists := te.Strategies[strategyID]
	if !exists {
		return errors.New("strategy not found")
	}

	strategy.Status = Stopped
	strategy.UpdatedAt = time.Now()
	te.UpdatedAt = time.Now()

	return nil
}

// PlaceOrder places a new trading order
func (te *TradingEngine) PlaceOrder(order *Order) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	if _, exists := te.Orders[order.ID]; exists {
		return errors.New("order with this ID already exists")
	}

	// Validate strategy exists and is active
	strategy, exists := te.Strategies[order.StrategyID]
	if !exists {
		return errors.New("strategy not found")
	}
	if strategy.Status != Active {
		return errors.New("strategy is not active")
	}

	te.Orders[order.ID] = order
	te.UpdatedAt = time.Now()

	return nil
}

// CancelOrder cancels an existing order
func (te *TradingEngine) CancelOrder(orderID string) error {
	if orderID == "" {
		return errors.New("order ID cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	order, exists := te.Orders[orderID]
	if !exists {
		return errors.New("order not found")
	}

	if order.Status == Filled || order.Status == Cancelled {
		return errors.New("order cannot be cancelled")
	}

	order.Status = Cancelled
	order.UpdatedAt = time.Now()
	te.UpdatedAt = time.Now()

	return nil
}

// UpdateMarketData updates market data for an asset
func (te *TradingEngine) UpdateMarketData(marketData *MarketData) error {
	if marketData == nil {
		return errors.New("market data cannot be nil")
	}
	if marketData.AssetID == "" {
		return errors.New("asset ID cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	te.MarketData[marketData.AssetID] = marketData
	te.UpdatedAt = time.Now()

	return nil
}

// ExecuteMarketMaking executes market making logic
func (mm *MarketMaker) ExecuteMarketMaking(marketData *MarketData, currentInventory *big.Float) ([]*Order, error) {
	if marketData == nil {
		return nil, errors.New("market data cannot be nil")
	}
	if currentInventory == nil {
		return nil, errors.New("current inventory cannot be nil")
	}

	// Calculate bid and ask prices based on spread
	midPrice := new(big.Float).Add(marketData.BidPrice, marketData.AskPrice)
	midPrice.Quo(midPrice, big.NewFloat(2))

	spreadHalf := new(big.Float).Mul(midPrice, mm.SpreadPercentage)
	spreadHalf.Quo(spreadHalf, big.NewFloat(2))

	bidPrice := new(big.Float).Sub(midPrice, spreadHalf)
	askPrice := new(big.Float).Add(midPrice, spreadHalf)

	// Calculate order sizes based on inventory target
	inventoryDeviation := new(big.Float).Sub(currentInventory, mm.InventoryTarget)

	var orders []*Order

	// Place bid order if we can buy more
	if inventoryDeviation.Cmp(mm.MaxInventory) < 0 {
		bidSize := mm.calculateOrderSize(bidPrice, Buy)
		if bidSize.Sign() > 0 {
			bidOrder, _ := NewOrder(
				generateOrderID("BID", mm.ID),
				mm.ID,
				marketData.AssetID,
				Buy,
				bidSize,
				bidPrice,
				Limit,
			)
			orders = append(orders, bidOrder)
		}
	}

	// Place ask order if we can sell more
	if inventoryDeviation.Cmp(new(big.Float).Neg(mm.MaxInventory)) > 0 {
		askSize := mm.calculateOrderSize(askPrice, Sell)
		if askSize.Sign() > 0 {
			askOrder, _ := NewOrder(
				generateOrderID("ASK", mm.ID),
				mm.ID,
				marketData.AssetID,
				Sell,
				askSize,
				askPrice,
				Limit,
			)
			orders = append(orders, askOrder)
		}
	}

	return orders, nil
}

// calculateOrderSize calculates the appropriate order size
func (mm *MarketMaker) calculateOrderSize(price *big.Float, side TradeSide) *big.Float {
	// Simple size calculation - in practice, this would be more sophisticated
	baseSize := mm.MinOrderSize

	// Adjust size based on price volatility (simplified)
	if side == Buy {
		// Buy more when price is lower
		priceAdjustment := new(big.Float).Quo(big.NewFloat(1), price)
		baseSize.Mul(baseSize, priceAdjustment)
	} else {
		// Sell more when price is higher
		baseSize.Mul(baseSize, price)
	}

	// Ensure size is within bounds
	if baseSize.Cmp(mm.MinOrderSize) < 0 {
		baseSize = new(big.Float).Copy(mm.MinOrderSize)
	}
	if baseSize.Cmp(mm.MaxOrderSize) > 0 {
		baseSize = new(big.Float).Copy(mm.MaxOrderSize)
	}

	return baseSize
}

// DetectArbitrage detects arbitrage opportunities
func (as *ArbitrageStrategy) DetectArbitrage(marketData1, marketData2 *MarketData) (*ArbitrageOpportunity, error) {
	if marketData1 == nil || marketData2 == nil {
		return nil, errors.New("market data cannot be nil")
	}

	// Calculate potential profit
	// Buy at lower ask, sell at higher bid
	profit1 := new(big.Float).Sub(marketData2.BidPrice, marketData1.AskPrice)
	profit2 := new(big.Float).Sub(marketData1.BidPrice, marketData2.AskPrice)

	var opportunity *ArbitrageOpportunity

	// Check first direction
	if profit1.Cmp(as.MinProfitThreshold) > 0 {
		opportunity = &ArbitrageOpportunity{
			Asset1ID:  marketData1.AssetID,
			Asset2ID:  marketData2.AssetID,
			BuyAsset:  marketData1.AssetID,
			SellAsset: marketData2.AssetID,
			BuyPrice:  marketData1.AskPrice,
			SellPrice: marketData2.BidPrice,
			Profit:    profit1,
			Timestamp: time.Now(),
		}
	}

	// Check second direction
	if profit2.Cmp(as.MinProfitThreshold) > 0 {
		if opportunity == nil || profit2.Cmp(opportunity.Profit) > 0 {
			opportunity = &ArbitrageOpportunity{
				Asset1ID:  marketData1.AssetID,
				Asset2ID:  marketData2.AssetID,
				BuyAsset:  marketData2.AssetID,
				SellAsset: marketData1.AssetID,
				BuyPrice:  marketData2.AskPrice,
				SellPrice: marketData1.BidPrice,
				Profit:    profit2,
				Timestamp: time.Now(),
			}
		}
	}

	return opportunity, nil
}

// ArbitrageOpportunity represents an arbitrage opportunity
type ArbitrageOpportunity struct {
	Asset1ID  string
	Asset2ID  string
	BuyAsset  string
	SellAsset string
	BuyPrice  *big.Float
	SellPrice *big.Float
	Profit    *big.Float
	Timestamp time.Time
}

// ExecuteArbitrage executes an arbitrage trade
func (as *ArbitrageStrategy) ExecuteArbitrage(opportunity *ArbitrageOpportunity) ([]*Order, error) {
	if opportunity == nil {
		return nil, errors.New("arbitrage opportunity cannot be nil")
	}

	// Check if opportunity is still valid
	if time.Since(opportunity.Timestamp) > as.ExecutionDelay {
		return nil, errors.New("arbitrage opportunity expired")
	}

	var orders []*Order

	// Create buy order
	buyOrder, err := NewOrder(
		generateOrderID("ARB_BUY", as.ID),
		as.ID,
		opportunity.BuyAsset,
		Buy,
		big.NewFloat(1), // Fixed size for simplicity
		opportunity.BuyPrice,
		Market,
	)
	if err != nil {
		return nil, err
	}
	orders = append(orders, buyOrder)

	// Create sell order
	sellOrder, err := NewOrder(
		generateOrderID("ARB_SELL", as.ID),
		as.ID,
		opportunity.SellAsset,
		Sell,
		big.NewFloat(1), // Fixed size for simplicity
		opportunity.SellPrice,
		Market,
	)
	if err != nil {
		return nil, err
	}
	orders = append(orders, sellOrder)

	return orders, nil
}

// GetStrategy returns a strategy by ID
func (te *TradingEngine) GetStrategy(strategyID string) (*Strategy, error) {
	if strategyID == "" {
		return nil, errors.New("strategy ID cannot be empty")
	}

	te.mu.RLock()
	defer te.mu.RUnlock()

	strategy, exists := te.Strategies[strategyID]
	if !exists {
		return nil, errors.New("strategy not found")
	}

	return strategy, nil
}

// GetOrder returns an order by ID
func (te *TradingEngine) GetOrder(orderID string) (*Order, error) {
	if orderID == "" {
		return nil, errors.New("order ID cannot be empty")
	}

	te.mu.RLock()
	defer te.mu.RUnlock()

	order, exists := te.Orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

// GetOrdersByStrategy returns all orders for a specific strategy
func (te *TradingEngine) GetOrdersByStrategy(strategyID string) ([]*Order, error) {
	if strategyID == "" {
		return nil, errors.New("strategy ID cannot be empty")
	}

	te.mu.RLock()
	defer te.mu.RUnlock()

	var orders []*Order
	for _, order := range te.Orders {
		if order.StrategyID == strategyID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// GetActiveStrategies returns all active strategies
func (te *TradingEngine) GetActiveStrategies() []*Strategy {
	te.mu.RLock()
	defer te.mu.RUnlock()

	var activeStrategies []*Strategy
	for _, strategy := range te.Strategies {
		if strategy.Status == Active {
			activeStrategies = append(activeStrategies, strategy)
		}
	}

	return activeStrategies
}

// GetMarketData returns market data for an asset
func (te *TradingEngine) GetMarketData(assetID string) (*MarketData, error) {
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}

	te.mu.RLock()
	defer te.mu.RUnlock()

	marketData, exists := te.MarketData[assetID]
	if !exists {
		return nil, errors.New("market data not found")
	}

	return marketData, nil
}

// Helper function to generate order IDs
func generateOrderID(prefix, strategyID string) string {
	return prefix + "_" + strategyID + "_" + time.Now().Format("20060102150405")
}
