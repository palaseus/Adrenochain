package algorithmictrading

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// StrategyType represents different trading strategy types
type StrategyType string

const (
	StrategyTypeMeanReversion StrategyType = "mean_reversion"
	StrategyTypeMomentum      StrategyType = "momentum"
	StrategyTypeArbitrage     StrategyType = "arbitrage"
	StrategyTypeGrid          StrategyType = "grid"
	StrategyTypeDCA           StrategyType = "dollar_cost_averaging"
)

// TradingBot represents an automated trading bot
type TradingBot struct {
	ID          string
	Name        string
	Strategy    TradingStrategy
	Config      BotConfig
	State       BotState
	RiskManager RiskManager
	Logger      *logger.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// BotConfig contains configuration for a trading bot
type BotConfig struct {
	MaxPositionSize   float64       `json:"max_position_size"`
	MaxDrawdown       float64       `json:"max_drawdown"`
	StopLoss          float64       `json:"stop_loss"`
	TakeProfit        float64       `json:"take_profit"`
	MaxOrders         int           `json:"max_orders"`
	OrderTimeout      time.Duration `json:"order_timeout"`
	RiskPerTrade      float64       `json:"risk_per_trade"`
	MaxDailyLoss      float64       `json:"max_daily_loss"`
	RebalanceInterval time.Duration `json:"rebalance_interval"`
}

// BotState represents the current state of a trading bot
type BotState struct {
	IsActive        bool        `json:"is_active"`
	CurrentPosition float64     `json:"current_position"`
	TotalPnL        float64     `json:"total_pnl"`
	DailyPnL        float64     `json:"daily_pnl"`
	OpenOrders      int         `json:"open_orders"`
	LastTrade       time.Time   `json:"last_trade"`
	LastRebalance   time.Time   `json:"last_rebalance"`
	RiskMetrics     RiskMetrics `json:"risk_metrics"`
}

// RiskMetrics contains risk-related metrics
type RiskMetrics struct {
	CurrentDrawdown float64 `json:"current_drawdown"`
	MaxDrawdown     float64 `json:"max_drawdown"`
	SharpeRatio     float64 `json:"sharpe_ratio"`
	VaR             float64 `json:"var"`
	PositionSize    float64 `json:"position_size"`
}

// TradingStrategy defines the interface for trading strategies
type TradingStrategy interface {
	GenerateSignals(marketData MarketData) []TradingSignal
	CalculatePositionSize(signal TradingSignal, riskManager RiskManager) float64
	ValidateSignal(signal TradingSignal) error
	GetStrategyType() StrategyType
}

// TradingSignal represents a trading signal
type TradingSignal struct {
	Type       SignalType             `json:"type"`
	Symbol     string                 `json:"symbol"`
	Price      float64                `json:"price"`
	Quantity   float64                `json:"quantity"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
	Strategy   StrategyType           `json:"strategy"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SignalType represents the type of trading signal
type SignalType string

const (
	SignalTypeBuy  SignalType = "buy"
	SignalTypeSell SignalType = "sell"
	SignalTypeHold SignalType = "hold"
)

// MarketData contains market information
type MarketData struct {
	Symbol     string    `json:"symbol"`
	Price      float64   `json:"price"`
	Volume     float64   `json:"volume"`
	Timestamp  time.Time `json:"timestamp"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Spread     float64   `json:"spread"`
	Volatility float64   `json:"volatility"`
	Trend      float64   `json:"trend"`
}

// RiskManager defines the interface for risk management
type RiskManager interface {
	ValidateOrder(order TradingSignal) error
	CalculatePositionSize(signal TradingSignal, availableCapital float64) float64
	CheckRiskLimits(bot *TradingBot) error
	UpdateRiskMetrics(bot *TradingBot, trade TradeResult)
}

// TradeResult represents the result of a trade
type TradeResult struct {
	OrderID   string     `json:"order_id"`
	Symbol    string     `json:"symbol"`
	Type      SignalType `json:"type"`
	Price     float64    `json:"price"`
	Quantity  float64    `json:"quantity"`
	Timestamp time.Time  `json:"timestamp"`
	PnL       float64    `json:"pnl"`
	Fees      float64    `json:"fees"`
}

// NewTradingBot creates a new trading bot instance
func NewTradingBot(id, name string, strategy TradingStrategy, config BotConfig) *TradingBot {
	ctx, cancel := context.WithCancel(context.Background())

	return &TradingBot{
		ID:          id,
		Name:        name,
		Strategy:    strategy,
		Config:      config,
		State:       BotState{},
		RiskManager: NewDefaultRiskManager(),
		Logger: logger.NewLogger(&logger.Config{
			Level:   logger.INFO,
			Prefix:  fmt.Sprintf("trading_bot_%s", id),
			UseJSON: false,
		}),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start activates the trading bot
func (bot *TradingBot) Start() error {
	bot.mu.Lock()
	defer bot.mu.Unlock()

	if bot.State.IsActive {
		return fmt.Errorf("bot %s is already active", bot.ID)
	}

	bot.State.IsActive = true
	bot.Logger.Info("Trading bot started - bot_id: %s, strategy: %s", bot.ID, bot.Strategy.GetStrategyType())

	// Start the main trading loop
	go bot.tradingLoop()

	return nil
}

// Stop deactivates the trading bot
func (bot *TradingBot) Stop() error {
	bot.mu.Lock()
	defer bot.mu.Unlock()

	if !bot.State.IsActive {
		return fmt.Errorf("bot %s is not active", bot.ID)
	}

	bot.State.IsActive = false
	bot.cancel()
	bot.Logger.Info("Trading bot stopped - bot_id: %s", bot.ID)

	return nil
}

// tradingLoop is the main trading loop
func (bot *TradingBot) tradingLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bot.ctx.Done():
			return
		case <-ticker.C:
			if err := bot.executeTradingCycle(); err != nil {
				bot.Logger.Error("Error in trading cycle - error: %v, bot_id: %s", err, bot.ID)
			}
		}
	}
}

// executeTradingCycle executes one complete trading cycle
func (bot *TradingBot) executeTradingCycle() error {
	bot.mu.Lock()
	defer bot.mu.Unlock()

	// Check if bot is still active
	if !bot.State.IsActive {
		return nil
	}

	// Check risk limits
	if err := bot.RiskManager.CheckRiskLimits(bot); err != nil {
		bot.Logger.Warn("Risk limits exceeded, pausing trading - error: %v, bot_id: %s", err, bot.ID)
		return err
	}

	// Get market data (this would come from exchange)
	marketData := bot.getMarketData()

	// Generate trading signals
	signals := bot.Strategy.GenerateSignals(marketData)

	// Process signals
	for _, signal := range signals {
		if err := bot.processSignal(signal); err != nil {
			bot.Logger.Error("Error processing signal - error: %v, signal: %+v, bot_id: %s", err, signal, bot.ID)
		}
	}

	// Rebalance if needed
	if time.Since(bot.State.LastRebalance) >= bot.Config.RebalanceInterval {
		if err := bot.rebalance(); err != nil {
			bot.Logger.Error("Error during rebalancing - error: %v, bot_id: %s", err, bot.ID)
		}
	}

	return nil
}

// processSignal processes a single trading signal
func (bot *TradingBot) processSignal(signal TradingSignal) error {
	// Validate signal
	if err := bot.Strategy.ValidateSignal(signal); err != nil {
		return fmt.Errorf("invalid signal: %w", err)
	}

	// Check risk limits
	if err := bot.RiskManager.ValidateOrder(signal); err != nil {
		return fmt.Errorf("risk validation failed: %w", err)
	}

	// Calculate position size
	positionSize := bot.Strategy.CalculatePositionSize(signal, bot.RiskManager)

	// Execute order (this would interface with exchange)
	orderResult, err := bot.executeOrder(signal, positionSize)
	if err != nil {
		return fmt.Errorf("order execution failed: %w", err)
	}

	// Update bot state
	bot.updateState(orderResult)

	return nil
}

// executeOrder executes a trading order
func (bot *TradingBot) executeOrder(signal TradingSignal, positionSize float64) (*TradeResult, error) {
	// This is a placeholder - in real implementation, this would interface with exchange
	// For now, we'll simulate order execution

	orderResult := &TradeResult{
		OrderID:   fmt.Sprintf("order_%d", time.Now().UnixNano()),
		Symbol:    signal.Symbol,
		Type:      signal.Type,
		Price:     signal.Price,
		Quantity:  positionSize,
		Timestamp: time.Now(),
		PnL:       0,                                   // Would be calculated based on position
		Fees:      positionSize * signal.Price * 0.001, // 0.1% fee
	}

	bot.Logger.Info("Order executed - order_id: %s, symbol: %s, type: %s, price: %.2f, quantity: %.2f, bot_id: %s",
		orderResult.OrderID, orderResult.Symbol, orderResult.Type, orderResult.Price, orderResult.Quantity, bot.ID)

	return orderResult, nil
}

// updateState updates the bot's state based on trade results
func (bot *TradingBot) updateState(tradeResult *TradeResult) {
	// Update position
	switch tradeResult.Type {
	case SignalTypeBuy:
		bot.State.CurrentPosition += tradeResult.Quantity
	case SignalTypeSell:
		bot.State.CurrentPosition -= tradeResult.Quantity
	}

	// Update PnL
	bot.State.TotalPnL += tradeResult.PnL
	bot.State.DailyPnL += tradeResult.PnL

	// Update last trade time
	bot.State.LastTrade = tradeResult.Timestamp

	// Update risk metrics
	bot.RiskManager.UpdateRiskMetrics(bot, *tradeResult)
}

// rebalance performs portfolio rebalancing
func (bot *TradingBot) rebalance() error {
	bot.Logger.Info("Starting portfolio rebalancing - bot_id: %s", bot.ID)

	// This is a placeholder - in real implementation, this would:
	// 1. Calculate target allocations
	// 2. Generate rebalancing orders
	// 3. Execute orders to achieve target allocations

	bot.State.LastRebalance = time.Now()
	bot.Logger.Info("Portfolio rebalancing completed - bot_id: %s", bot.ID)

	return nil
}

// getMarketData retrieves current market data
func (bot *TradingBot) getMarketData() MarketData {
	// This is a placeholder - in real implementation, this would fetch from exchange
	return MarketData{
		Symbol:     "BTC/USDT",
		Price:      50000.0,
		Volume:     1000000.0,
		Timestamp:  time.Now(),
		Bid:        49999.0,
		Ask:        50001.0,
		Spread:     2.0,
		Volatility: 0.02,
		Trend:      0.01,
	}
}

// GetState returns a copy of the bot's current state
func (bot *TradingBot) GetState() BotState {
	bot.mu.RLock()
	defer bot.mu.RUnlock()
	return bot.State
}

// UpdateConfig updates the bot's configuration
func (bot *TradingBot) UpdateConfig(config BotConfig) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()

	// Validate new configuration
	if config.MaxPositionSize <= 0 {
		return fmt.Errorf("max position size must be positive")
	}
	if config.MaxDrawdown <= 0 || config.MaxDrawdown > 1 {
		return fmt.Errorf("max drawdown must be between 0 and 1")
	}

	bot.Config = config
	bot.Logger.Info("Configuration updated - bot_id: %s, config: %+v", bot.ID, config)

	return nil
}

// GetPerformanceMetrics returns performance metrics for the bot
func (bot *TradingBot) GetPerformanceMetrics() map[string]interface{} {
	bot.mu.RLock()
	defer bot.mu.RUnlock()

	return map[string]interface{}{
		"total_pnl":        bot.State.TotalPnL,
		"daily_pnl":        bot.State.DailyPnL,
		"current_position": bot.State.CurrentPosition,
		"open_orders":      bot.State.OpenOrders,
		"risk_metrics":     bot.State.RiskMetrics,
		"last_trade":       bot.State.LastTrade,
		"last_rebalance":   bot.State.LastRebalance,
	}
}

// DefaultRiskManager implements the RiskManager interface with basic risk controls
type DefaultRiskManager struct{}

// NewDefaultRiskManager creates a new default risk manager
func NewDefaultRiskManager() *DefaultRiskManager {
	return &DefaultRiskManager{}
}

// ValidateOrder validates if an order meets risk requirements
func (rm *DefaultRiskManager) ValidateOrder(order TradingSignal) error {
	if order.Price <= 0 {
		return fmt.Errorf("invalid price: must be positive")
	}
	if order.Quantity <= 0 {
		return fmt.Errorf("invalid quantity: must be positive")
	}
	if order.Confidence < 0 || order.Confidence > 1 {
		return fmt.Errorf("invalid confidence: must be between 0 and 1")
	}
	return nil
}

// CalculatePositionSize calculates the appropriate position size based on risk
func (rm *DefaultRiskManager) CalculatePositionSize(signal TradingSignal, availableCapital float64) float64 {
	// Basic position sizing based on confidence and available capital
	baseSize := availableCapital * 0.02 // 2% risk per trade
	confidenceMultiplier := signal.Confidence
	return baseSize * confidenceMultiplier
}

// CheckRiskLimits checks if the bot is within risk limits
func (rm *DefaultRiskManager) CheckRiskLimits(bot *TradingBot) error {
	// Check daily loss limit
	if bot.State.DailyPnL < -bot.Config.MaxDailyLoss {
		return fmt.Errorf("daily loss limit exceeded: %.2f", bot.State.DailyPnL)
	}

	// Check drawdown limit
	if bot.State.RiskMetrics.CurrentDrawdown > bot.Config.MaxDrawdown {
		return fmt.Errorf("drawdown limit exceeded: %.2f", bot.State.RiskMetrics.CurrentDrawdown)
	}

	// Check position size limit
	if math.Abs(bot.State.CurrentPosition) > bot.Config.MaxPositionSize {
		return fmt.Errorf("position size limit exceeded: %.2f", bot.State.CurrentPosition)
	}

	return nil
}

// UpdateRiskMetrics updates risk metrics based on trade results
func (rm *DefaultRiskManager) UpdateRiskMetrics(bot *TradingBot, trade TradeResult) {
	// Update drawdown calculation
	if trade.PnL < 0 {
		bot.State.RiskMetrics.CurrentDrawdown += math.Abs(trade.PnL)
		if bot.State.RiskMetrics.CurrentDrawdown > bot.State.RiskMetrics.MaxDrawdown {
			bot.State.RiskMetrics.MaxDrawdown = bot.State.RiskMetrics.CurrentDrawdown
		}
	} else {
		// Reduce drawdown on profitable trades
		bot.State.RiskMetrics.CurrentDrawdown = math.Max(0, bot.State.RiskMetrics.CurrentDrawdown-trade.PnL)
	}

	// Update position size
	bot.State.RiskMetrics.PositionSize = math.Abs(bot.State.CurrentPosition)

	// Simple VaR calculation (95% confidence)
	// In a real implementation, this would use historical data
	bot.State.RiskMetrics.VaR = bot.State.RiskMetrics.PositionSize * 0.02 // 2% daily VaR
}
