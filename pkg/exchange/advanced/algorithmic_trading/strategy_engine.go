package algorithmictrading

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/logger"
)

// StrategyEngine manages multiple trading strategies and their execution
type StrategyEngine struct {
	strategies map[string]TradingStrategy
	executors  map[string]*StrategyExecutor
	config     EngineConfig
	logger     *logger.Logger
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// EngineConfig contains configuration for the strategy engine
type EngineConfig struct {
	MaxConcurrentStrategies int           `json:"max_concurrent_strategies"`
	ExecutionInterval       time.Duration `json:"execution_interval"`
	MaxRetries              int           `json:"max_retries"`
	RetryDelay              time.Duration `json:"retry_delay"`
	EnableBacktesting       bool          `json:"enable_backtesting"`
	BacktestDataPath        string        `json:"backtest_data_path"`
}

// StrategyExecutor manages the execution of a single strategy
type StrategyExecutor struct {
	ID           string
	Strategy     TradingStrategy
	Config       ExecutorConfig
	State        ExecutorState
	MarketData   MarketDataProvider
	OrderManager OrderManager
	Logger       *logger.Logger
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// ExecutorConfig contains configuration for a strategy executor
type ExecutorConfig struct {
	Enabled         bool          `json:"enabled"`
	MaxPositionSize float64       `json:"max_position_size"`
	RiskPerTrade    float64       `json:"risk_per_trade"`
	ExecutionDelay  time.Duration `json:"execution_delay"`
	SignalThreshold float64       `json:"signal_threshold"`
	MaxOrders       int           `json:"max_orders"`
	OrderTimeout    time.Duration `json:"order_timeout"`
}

// ExecutorState represents the current state of a strategy executor
type ExecutorState struct {
	IsActive        bool      `json:"is_active"`
	LastExecution   time.Time `json:"last_execution"`
	TotalSignals    int       `json:"total_signals"`
	ExecutedSignals int       `json:"executed_signals"`
	FailedSignals   int       `json:"failed_signals"`
	CurrentPosition float64   `json:"current_position"`
	TotalPnL        float64   `json:"total_pnl"`
	OpenOrders      int       `json:"open_orders"`
}

// MarketDataProvider defines the interface for market data
type MarketDataProvider interface {
	GetMarketData(symbol string) (MarketData, error)
	GetHistoricalData(symbol string, start, end time.Time, interval time.Duration) ([]MarketData, error)
	SubscribeToUpdates(symbol string, callback func(MarketData)) error
	Unsubscribe(symbol string) error
}

// OrderManager defines the interface for order management
type OrderManager interface {
	PlaceOrder(order TradingSignal) (string, error)
	CancelOrder(orderID string) error
	GetOrderStatus(orderID string) (OrderStatus, error)
	GetOpenOrders() ([]OrderStatus, error)
}

// OrderStatus represents the status of an order
type OrderStatus struct {
	OrderID   string     `json:"order_id"`
	Symbol    string     `json:"symbol"`
	Type      SignalType `json:"type"`
	Price     float64    `json:"price"`
	Quantity  float64    `json:"quantity"`
	Status    string     `json:"status"`
	Filled    float64    `json:"filled"`
	Remaining float64    `json:"remaining"`
	Timestamp time.Time  `json:"timestamp"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewStrategyEngine creates a new strategy engine
func NewStrategyEngine(config EngineConfig) *StrategyEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &StrategyEngine{
		strategies: make(map[string]TradingStrategy),
		executors:  make(map[string]*StrategyExecutor),
		config:     config,
		logger:     logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "strategy_engine"}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterStrategy registers a new trading strategy
func (engine *StrategyEngine) RegisterStrategy(id string, strategy TradingStrategy, config ExecutorConfig) error {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if _, exists := engine.strategies[id]; exists {
		return fmt.Errorf("strategy with ID %s already exists", id)
	}

	engine.strategies[id] = strategy

	// Create executor for the strategy
	executor := NewStrategyExecutor(id, strategy, config)
	engine.executors[id] = executor

	engine.logger.Info("Strategy registered - strategy_id: %s, type: %s", id, strategy.GetStrategyType())
	return nil
}

// UnregisterStrategy removes a trading strategy
func (engine *StrategyEngine) UnregisterStrategy(id string) error {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if _, exists := engine.strategies[id]; !exists {
		return fmt.Errorf("strategy with ID %s not found", id)
	}

	// Stop executor if running
	if executor, exists := engine.executors[id]; exists {
		executor.Stop()
		delete(engine.executors, id)
	}

	delete(engine.strategies, id)
	engine.logger.Info("Strategy unregistered - strategy_id: %s", id)
	return nil
}

// StartEngine starts the strategy engine
func (engine *StrategyEngine) StartEngine() error {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.logger.Info("Starting strategy engine")

	// Start all enabled executors
	for id, executor := range engine.executors {
		if executor.Config.Enabled {
			if err := executor.Start(); err != nil {
				engine.logger.Error("Failed to start executor - strategy_id: %s, error: %v", id, err)
				continue
			}
		}
	}

	// Start the main execution loop
	go engine.executionLoop()

	return nil
}

// StopEngine stops the strategy engine
func (engine *StrategyEngine) StopEngine() error {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.logger.Info("Stopping strategy engine")

	// Stop all executors
	for id, executor := range engine.executors {
		if err := executor.Stop(); err != nil {
			engine.logger.Error("Failed to stop executor - strategy_id: %s, error: %v", id, err)
		}
	}

	engine.cancel()
	return nil
}

// executionLoop is the main execution loop for the strategy engine
func (engine *StrategyEngine) executionLoop() {
	ticker := time.NewTicker(engine.config.ExecutionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-engine.ctx.Done():
			return
		case <-ticker.C:
			engine.executeStrategies()
		}
	}
}

// executeStrategies executes all active strategies
func (engine *StrategyEngine) executeStrategies() {
	engine.mu.RLock()
	executors := make([]*StrategyExecutor, 0, len(engine.executors))
	for _, executor := range engine.executors {
		if executor.Config.Enabled && executor.State.IsActive {
			executors = append(executors, executor)
		}
	}
	engine.mu.RUnlock()

	// Execute strategies concurrently (with limit)
	semaphore := make(chan struct{}, engine.config.MaxConcurrentStrategies)
	var wg sync.WaitGroup

	for _, executor := range executors {
		wg.Add(1)
		go func(exec *StrategyExecutor) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := exec.Execute(); err != nil {
				engine.logger.Error("Strategy execution failed - strategy_id: %s, error: %v", exec.ID, err)
			}
		}(executor)
	}

	wg.Wait()
}

// GetStrategyStatus returns the status of all strategies
func (engine *StrategyEngine) GetStrategyStatus() map[string]ExecutorState {
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	status := make(map[string]ExecutorState)
	for id, executor := range engine.executors {
		status[id] = executor.GetState()
	}
	return status
}

// NewStrategyExecutor creates a new strategy executor
func NewStrategyExecutor(id string, strategy TradingStrategy, config ExecutorConfig) *StrategyExecutor {
	ctx, cancel := context.WithCancel(context.Background())

	return &StrategyExecutor{
		ID:       id,
		Strategy: strategy,
		Config:   config,
		State:    ExecutorState{},
		Logger:   logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: fmt.Sprintf("executor_%s", id)}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start activates the strategy executor
func (executor *StrategyExecutor) Start() error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	if executor.State.IsActive {
		return fmt.Errorf("executor %s is already active", executor.ID)
	}

	executor.State.IsActive = true
	executor.Logger.Info("Strategy executor started - strategy_id: %s", executor.ID)
	return nil
}

// Stop deactivates the strategy executor
func (executor *StrategyExecutor) Stop() error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	if !executor.State.IsActive {
		return fmt.Errorf("executor %s is not active", executor.ID)
	}

	executor.State.IsActive = false
	executor.cancel()
	executor.Logger.Info("Strategy executor stopped - strategy_id: %s", executor.ID)
	return nil
}

// Execute executes the strategy once
func (executor *StrategyExecutor) Execute() error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	if !executor.State.IsActive {
		return nil
	}

	// Check if enough time has passed since last execution
	if time.Since(executor.State.LastExecution) < executor.Config.ExecutionDelay {
		return nil
	}

	// Get market data
	marketData, err := executor.getMarketData()
	if err != nil {
		return fmt.Errorf("failed to get market data: %w", err)
	}

	// Generate trading signals
	signals := executor.Strategy.GenerateSignals(marketData)
	executor.State.TotalSignals += len(signals)

	// Process signals
	for _, signal := range signals {
		if err := executor.processSignal(signal); err != nil {
			executor.State.FailedSignals++
			executor.Logger.Error("Failed to process signal - error: %v, signal: %+v", err, signal)
		} else {
			executor.State.ExecutedSignals++
		}
	}

	executor.State.LastExecution = time.Now()
	return nil
}

// processSignal processes a single trading signal
func (executor *StrategyExecutor) processSignal(signal TradingSignal) error {
	// Validate signal
	if err := executor.Strategy.ValidateSignal(signal); err != nil {
		return fmt.Errorf("invalid signal: %w", err)
	}

	// Check signal threshold
	if signal.Confidence < executor.Config.SignalThreshold {
		executor.Logger.Debug("Signal below threshold, skipping - confidence: %.2f, threshold: %.2f", signal.Confidence, executor.Config.SignalThreshold)
		return nil
	}

	// Check if we can place more orders
	if executor.State.OpenOrders >= executor.Config.MaxOrders {
		executor.Logger.Warn("Maximum orders reached, skipping signal - open_orders: %d", executor.State.OpenOrders)
		return nil
	}

	// Place order
	if executor.OrderManager != nil {
		orderID, err := executor.OrderManager.PlaceOrder(signal)
		if err != nil {
			return fmt.Errorf("failed to place order: %w", err)
		}

		executor.State.OpenOrders++
		executor.Logger.Info("Order placed successfully - order_id: %s, signal: %+v", orderID, signal)
	}

	return nil
}

// getMarketData retrieves market data for the strategy
func (executor *StrategyExecutor) getMarketData() (MarketData, error) {
	// This is a placeholder - in real implementation, this would use MarketDataProvider
	// For now, return mock data
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
	}, nil
}

// GetState returns a copy of the executor's current state
func (executor *StrategyExecutor) GetState() ExecutorState {
	executor.mu.RLock()
	defer executor.mu.RUnlock()
	return executor.State
}

// UpdateConfig updates the executor's configuration
func (executor *StrategyExecutor) UpdateConfig(config ExecutorConfig) error {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	// Validate configuration
	if config.MaxPositionSize <= 0 {
		return fmt.Errorf("max position size must be positive")
	}
	if config.RiskPerTrade <= 0 || config.RiskPerTrade > 1 {
		return fmt.Errorf("risk per trade must be between 0 and 1")
	}

	executor.Config = config
	executor.Logger.Info("Configuration updated - strategy_id: %s, config: %+v", executor.ID, config)
	return nil
}

// SetMarketDataProvider sets the market data provider for the executor
func (executor *StrategyExecutor) SetMarketDataProvider(provider MarketDataProvider) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	executor.MarketData = provider
}

// SetOrderManager sets the order manager for the executor
func (executor *StrategyExecutor) SetOrderManager(manager OrderManager) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	executor.OrderManager = manager
}
