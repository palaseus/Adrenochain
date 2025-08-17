package marketmaking

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/logger"
)

// MarketMakerStrategy represents a market making strategy
type MarketMakerStrategy interface {
	CalculateQuotes(marketData MarketData, position Position) QuotePair
	UpdateStrategy(marketData MarketData, trades []Trade)
	GetStrategyType() string
	GetParameters() map[string]interface{}
	ValidateParameters(params map[string]interface{}) error
}

// QuotePair represents a pair of bid and ask quotes
type QuotePair struct {
	Bid       Quote     `json:"bid"`
	Ask       Quote     `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
	Strategy  string    `json:"strategy"`
}

// Quote represents a single quote
type Quote struct {
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Side      string    `json:"side"` // "bid" or "ask"
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Position represents the current market making position
type Position struct {
	Symbol        string    `json:"symbol"`
	Quantity      float64   `json:"quantity"`
	AveragePrice  float64   `json:"average_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	RealizedPnL   float64   `json:"realized_pnl"`
	LastUpdate    time.Time `json:"last_update"`
}

// Trade represents a completed trade
type Trade struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	Timestamp    time.Time `json:"timestamp"`
	Counterparty string    `json:"counterparty"`
}

// MarketData contains market information for quote calculation
type MarketData struct {
	Symbol     string    `json:"symbol"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	MidPrice   float64   `json:"mid_price"`
	Spread     float64   `json:"spread"`
	Volume     float64   `json:"volume"`
	Volatility float64   `json:"volatility"`
	Timestamp  time.Time `json:"timestamp"`
	OrderBook  OrderBook `json:"order_book"`
}

// OrderBook represents the current order book state
type OrderBook struct {
	Bids []OrderBookLevel `json:"bids"`
	Asks []OrderBookLevel `json:"asks"`
}

// OrderBookLevel represents a level in the order book
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// MarketMaker manages market making operations
type MarketMaker struct {
	ID       string
	Strategy MarketMakerStrategy
	Config   MarketMakerConfig
	Position Position
	Quotes   map[string]QuotePair
	Logger   *logger.Logger
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// MarketMakerConfig contains configuration for market making
type MarketMakerConfig struct {
	MaxPositionSize  float64       `json:"max_position_size"`
	MaxSpread        float64       `json:"max_spread"`
	MinSpread        float64       `json:"min_spread"`
	QuoteSize        float64       `json:"quote_size"`
	QuoteRefreshRate time.Duration `json:"quote_refresh_rate"`
	RiskLimit        float64       `json:"risk_limit"`
	EnableHedging    bool          `json:"enable_hedging"`
	HedgingThreshold float64       `json:"hedging_threshold"`
	MaxQuotes        int           `json:"max_quotes"`
	QuoteTimeout     time.Duration `json:"quote_timeout"`
}

// NewMarketMaker creates a new market maker
func NewMarketMaker(id string, strategy MarketMakerStrategy, config MarketMakerConfig) *MarketMaker {
	ctx, cancel := context.WithCancel(context.Background())

	return &MarketMaker{
		ID:       id,
		Strategy: strategy,
		Config:   config,
		Position: Position{},
		Quotes:   make(map[string]QuotePair),
		Logger:   logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: fmt.Sprintf("market_maker_%s", id)}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the market maker
func (mm *MarketMaker) Start() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.Logger.Info("Market maker started - id: %s, strategy: %s", mm.ID, mm.Strategy.GetStrategyType())

	// Start quote refresh loop
	go mm.quoteRefreshLoop()

	return nil
}

// Stop stops the market maker
func (mm *MarketMaker) Stop() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.cancel()
	mm.Logger.Info("Market maker stopped - id: %s", mm.ID)
	return nil
}

// quoteRefreshLoop continuously refreshes quotes
func (mm *MarketMaker) quoteRefreshLoop() {
	ticker := time.NewTicker(mm.Config.QuoteRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.refreshQuotes()
		}
	}
}

// refreshQuotes refreshes all active quotes
func (mm *MarketMaker) refreshQuotes() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Get current market data (this would come from exchange)
	marketData := mm.getMarketData()

	// Calculate new quotes
	quotes := mm.Strategy.CalculateQuotes(marketData, mm.Position)

	// Update quotes
	mm.Quotes[marketData.Symbol] = quotes

	mm.Logger.Debug("Quotes refreshed - symbol: %s, bid: %.2f, ask: %.2f", marketData.Symbol, quotes.Bid.Price, quotes.Ask.Price)
}

// getMarketData retrieves current market data
func (mm *MarketMaker) getMarketData() MarketData {
	// This is a placeholder - in real implementation, this would fetch from exchange
	return MarketData{
		Symbol:     "BTC/USDT",
		Bid:        49999.0,
		Ask:        50001.0,
		MidPrice:   50000.0,
		Spread:     2.0,
		Volume:     1000000.0,
		Volatility: 0.02,
		Timestamp:  time.Now(),
		OrderBook: OrderBook{
			Bids: []OrderBookLevel{{Price: 49999.0, Quantity: 1.0}},
			Asks: []OrderBookLevel{{Price: 50001.0, Quantity: 1.0}},
		},
	}
}

// UpdatePosition updates the market maker's position
func (mm *MarketMaker) UpdatePosition(trade Trade) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Update position based on trade
	if trade.Side == "buy" {
		// We bought, so our position increases
		oldValue := mm.Position.Quantity * mm.Position.AveragePrice
		newValue := trade.Quantity * trade.Price
		totalQuantity := mm.Position.Quantity + trade.Quantity

		if totalQuantity > 0 {
			mm.Position.AveragePrice = (oldValue + newValue) / totalQuantity
		}
		mm.Position.Quantity = totalQuantity
	} else {
		// We sold, so our position decreases
		mm.Position.Quantity -= trade.Quantity

		// Calculate realized PnL
		if trade.Side == "sell" {
			realizedPnL := (trade.Price - mm.Position.AveragePrice) * trade.Quantity
			mm.Position.RealizedPnL += realizedPnL
		}
	}

	mm.Position.LastUpdate = time.Now()

	// Update strategy with trade information
	mm.Strategy.UpdateStrategy(mm.getMarketData(), []Trade{trade})

	mm.Logger.Info("Position updated - symbol: %s, quantity: %.2f, avg_price: %.2f", trade.Symbol, mm.Position.Quantity, mm.Position.AveragePrice)
}

// GetQuotes returns current quotes for a symbol
func (mm *MarketMaker) GetQuotes(symbol string) (QuotePair, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	quotes, exists := mm.Quotes[symbol]
	return quotes, exists
}

// GetPosition returns the current position
func (mm *MarketMaker) GetPosition() Position {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.Position
}

// UpdateConfig updates the market maker configuration
func (mm *MarketMaker) UpdateConfig(config MarketMakerConfig) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Validate configuration
	if config.MaxPositionSize <= 0 {
		return fmt.Errorf("max position size must be positive")
	}
	if config.MaxSpread <= config.MinSpread {
		return fmt.Errorf("max spread must be greater than min spread")
	}
	if config.QuoteSize <= 0 {
		return fmt.Errorf("quote size must be positive")
	}

	mm.Config = config
	mm.Logger.Info("Configuration updated - config: %+v", config)
	return nil
}

// Strategy Implementations

// BasicMarketMaking implements a basic market making strategy
type BasicMarketMaking struct {
	SpreadMultiplier float64 `json:"spread_multiplier"`
	PositionLimit    float64 `json:"position_limit"`
	QuoteSize        float64 `json:"quote_size"`
}

// NewBasicMarketMaking creates a new basic market making strategy
func NewBasicMarketMaking(spreadMultiplier, positionLimit, quoteSize float64) *BasicMarketMaking {
	return &BasicMarketMaking{
		SpreadMultiplier: spreadMultiplier,
		PositionLimit:    positionLimit,
		QuoteSize:        quoteSize,
	}
}

func (bmm *BasicMarketMaking) CalculateQuotes(marketData MarketData, position Position) QuotePair {
	midPrice := marketData.MidPrice
	spread := marketData.Spread * bmm.SpreadMultiplier

	// Adjust quotes based on position
	positionAdjustment := bmm.calculatePositionAdjustment(position, marketData)

	bidPrice := midPrice - spread/2 + positionAdjustment
	askPrice := midPrice + spread/2 + positionAdjustment

	// Ensure minimum spread
	minSpread := marketData.Spread * 1.1
	if askPrice-bidPrice < minSpread {
		midPoint := (bidPrice + askPrice) / 2
		bidPrice = midPoint - minSpread/2
		askPrice = midPoint + minSpread/2
	}

	// Calculate quote sizes based on position
	bidSize := bmm.calculateQuoteSize(position, "bid")
	askSize := bmm.calculateQuoteSize(position, "ask")

	now := time.Now()
	expiry := now.Add(30 * time.Second)

	return QuotePair{
		Bid: Quote{
			Price:     bidPrice,
			Quantity:  bidSize,
			Side:      "bid",
			Valid:     true,
			ExpiresAt: expiry,
		},
		Ask: Quote{
			Price:     askPrice,
			Quantity:  askSize,
			Side:      "ask",
			Valid:     true,
			ExpiresAt: expiry,
		},
		Timestamp: now,
		Strategy:  "basic_market_making",
	}
}

func (bmm *BasicMarketMaking) calculatePositionAdjustment(position Position, marketData MarketData) float64 {
	if position.Quantity == 0 {
		return 0
	}

	// Adjust quotes to encourage position reduction
	positionRatio := position.Quantity / bmm.PositionLimit
	adjustment := positionRatio * marketData.Spread * 0.5

	if position.Quantity > 0 {
		// Long position: lower bid, higher ask
		return -adjustment
	} else {
		// Short position: higher bid, lower ask
		return adjustment
	}
}

func (bmm *BasicMarketMaking) calculateQuoteSize(position Position, side string) float64 {
	baseSize := bmm.QuoteSize

	// Adjust size based on position
	if side == "bid" && position.Quantity > 0 {
		// Reduce bid size when long
		baseSize *= (1 - math.Abs(position.Quantity)/bmm.PositionLimit)
	} else if side == "ask" && position.Quantity < 0 {
		// Reduce ask size when short
		baseSize *= (1 - math.Abs(position.Quantity)/bmm.PositionLimit)
	}

	return math.Max(baseSize*0.1, baseSize) // Minimum 10% of base size
}

func (bmm *BasicMarketMaking) UpdateStrategy(marketData MarketData, trades []Trade) {
	// Basic strategy doesn't need complex updates
}

func (bmm *BasicMarketMaking) GetStrategyType() string {
	return "basic_market_making"
}

func (bmm *BasicMarketMaking) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"spread_multiplier": bmm.SpreadMultiplier,
		"position_limit":    bmm.PositionLimit,
		"quote_size":        bmm.QuoteSize,
	}
}

func (bmm *BasicMarketMaking) ValidateParameters(params map[string]interface{}) error {
	if spreadMultiplier, exists := params["spread_multiplier"]; exists {
		if val, ok := spreadMultiplier.(float64); !ok || val <= 0 {
			return fmt.Errorf("spread multiplier must be a positive number")
		}
	}
	if positionLimit, exists := params["position_limit"]; exists {
		if val, ok := positionLimit.(float64); !ok || val <= 0 {
			return fmt.Errorf("position limit must be a positive number")
		}
	}
	if quoteSize, exists := params["quote_size"]; exists {
		if val, ok := quoteSize.(float64); !ok || val <= 0 {
			return fmt.Errorf("quote size must be a positive number")
		}
	}
	return nil
}

// AdaptiveMarketMaking implements an adaptive market making strategy
type AdaptiveMarketMaking struct {
	BaseSpread           float64 `json:"base_spread"`
	VolatilityMultiplier float64 `json:"volatility_multiplier"`
	VolumeMultiplier     float64 `json:"volume_multiplier"`
	PositionLimit        float64 `json:"position_limit"`
	QuoteSize            float64 `json:"quote_size"`
	LearningRate         float64 `json:"learning_rate"`
	historicalData       []MarketData
}

// NewAdaptiveMarketMaking creates a new adaptive market making strategy
func NewAdaptiveMarketMaking(baseSpread, volatilityMultiplier, volumeMultiplier, positionLimit, quoteSize, learningRate float64) *AdaptiveMarketMaking {
	return &AdaptiveMarketMaking{
		BaseSpread:           baseSpread,
		VolatilityMultiplier: volatilityMultiplier,
		VolumeMultiplier:     volumeMultiplier,
		PositionLimit:        positionLimit,
		QuoteSize:            quoteSize,
		LearningRate:         learningRate,
		historicalData:       make([]MarketData, 0),
	}
}

func (amm *AdaptiveMarketMaking) CalculateQuotes(marketData MarketData, position Position) QuotePair {
	// Add to historical data
	amm.historicalData = append(amm.historicalData, marketData)
	if len(amm.historicalData) > 1000 {
		amm.historicalData = amm.historicalData[1:]
	}

	// Calculate adaptive spread
	adaptiveSpread := amm.calculateAdaptiveSpread(marketData)

	// Calculate position adjustment
	positionAdjustment := amm.calculatePositionAdjustment(position, marketData)

	midPrice := marketData.MidPrice
	bidPrice := midPrice - adaptiveSpread/2 + positionAdjustment
	askPrice := midPrice + adaptiveSpread/2 + positionAdjustment

	// Calculate quote sizes
	bidSize := amm.calculateQuoteSize(position, "bid", marketData)
	askSize := amm.calculateQuoteSize(position, "ask", marketData)

	now := time.Now()
	expiry := now.Add(30 * time.Second)

	return QuotePair{
		Bid: Quote{
			Price:     bidPrice,
			Quantity:  bidSize,
			Side:      "bid",
			Valid:     true,
			ExpiresAt: expiry,
		},
		Ask: Quote{
			Price:     askPrice,
			Quantity:  askSize,
			Side:      "ask",
			Valid:     true,
			ExpiresAt: expiry,
		},
		Timestamp: now,
		Strategy:  "adaptive_market_making",
	}
}

func (amm *AdaptiveMarketMaking) calculateAdaptiveSpread(marketData MarketData) float64 {
	baseSpread := amm.BaseSpread

	// Adjust for volatility
	volatilityAdjustment := marketData.Volatility * amm.VolatilityMultiplier

	// Adjust for volume (lower spread for higher volume)
	volumeAdjustment := -math.Log(marketData.Volume/1000000.0) * amm.VolumeMultiplier

	// Calculate adaptive spread
	adaptiveSpread := baseSpread + volatilityAdjustment + volumeAdjustment

	// Ensure minimum spread
	minSpread := marketData.Spread * 1.05
	return math.Max(adaptiveSpread, minSpread)
}

func (amm *AdaptiveMarketMaking) calculatePositionAdjustment(position Position, marketData MarketData) float64 {
	if position.Quantity == 0 {
		return 0
	}

	// More sophisticated position adjustment based on market conditions
	positionRatio := position.Quantity / amm.PositionLimit
	volatilityFactor := marketData.Volatility * 100 // Scale volatility

	adjustment := positionRatio * volatilityFactor * marketData.Spread

	if position.Quantity > 0 {
		return -adjustment
	} else {
		return adjustment
	}
}

func (amm *AdaptiveMarketMaking) calculateQuoteSize(position Position, side string, marketData MarketData) float64 {
	baseSize := amm.QuoteSize

	// Adjust size based on volatility
	volatilityAdjustment := 1 + marketData.Volatility*10

	// Adjust size based on position
	positionAdjustment := 1.0
	if side == "bid" && position.Quantity > 0 {
		positionAdjustment = 1 - math.Abs(position.Quantity)/amm.PositionLimit
	} else if side == "ask" && position.Quantity < 0 {
		positionAdjustment = 1 - math.Abs(position.Quantity)/amm.PositionLimit
	}

	adjustedSize := baseSize * volatilityAdjustment * positionAdjustment
	return math.Max(adjustedSize*0.1, adjustedSize)
}

func (amm *AdaptiveMarketMaking) UpdateStrategy(marketData MarketData, trades []Trade) {
	// Update strategy parameters based on market performance
	if len(trades) == 0 {
		return
	}

	// Calculate performance metrics
	var totalPnL float64
	for _, trade := range trades {
		// Calculate PnL for this trade
		// This is a simplified calculation
		totalPnL += trade.Price * trade.Quantity
	}

	// Adjust parameters based on performance
	if totalPnL > 0 {
		// Profitable: reduce spread to capture more volume
		amm.BaseSpread *= (1 - amm.LearningRate)
	} else {
		// Unprofitable: increase spread to reduce risk
		amm.BaseSpread *= (1 + amm.LearningRate)
	}

	// Ensure parameters stay within reasonable bounds
	amm.BaseSpread = math.Max(amm.BaseSpread, 0.0001)
	amm.BaseSpread = math.Min(amm.BaseSpread, 0.1)
}

func (amm *AdaptiveMarketMaking) GetStrategyType() string {
	return "adaptive_market_making"
}

func (amm *AdaptiveMarketMaking) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"base_spread":           amm.BaseSpread,
		"volatility_multiplier": amm.VolatilityMultiplier,
		"volume_multiplier":     amm.VolumeMultiplier,
		"position_limit":        amm.PositionLimit,
		"quote_size":            amm.QuoteSize,
		"learning_rate":         amm.LearningRate,
	}
}

func (amm *AdaptiveMarketMaking) ValidateParameters(params map[string]interface{}) error {
	// Validate all parameters
	requiredParams := []string{"base_spread", "volatility_multiplier", "volume_multiplier", "position_limit", "quote_size", "learning_rate"}

	for _, param := range requiredParams {
		if val, exists := params[param]; !exists {
			return fmt.Errorf("missing required parameter: %s", param)
		} else if floatVal, ok := val.(float64); !ok || floatVal <= 0 {
			return fmt.Errorf("parameter %s must be a positive number", param)
		}
	}

	return nil
}
