package algorithmictrading

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTradingStrategy implements TradingStrategy for testing
type MockTradingStrategy struct {
	signals []TradingSignal
}

func (m *MockTradingStrategy) GenerateSignals(marketData MarketData) []TradingSignal {
	// Generate a signal for each market data point
	signal := TradingSignal{
		Type:       SignalTypeBuy,
		Symbol:     marketData.Symbol,
		Price:      marketData.Price,
		Quantity:   0.1,
		Confidence: 0.8,
		Timestamp:  marketData.Timestamp,
		Strategy:   StrategyTypeMomentum,
	}
	return []TradingSignal{signal}
}

func (m *MockTradingStrategy) CalculatePositionSize(signal TradingSignal, riskManager RiskManager) float64 {
	return 100.0 // Mock position size
}

func (m *MockTradingStrategy) ValidateSignal(signal TradingSignal) error {
	if signal.Price <= 0 {
		return fmt.Errorf("invalid price")
	}
	return nil
}

func (m *MockTradingStrategy) GetStrategyType() StrategyType {
	return StrategyTypeMomentum
}

// MockRiskManager implements RiskManager for testing
type MockRiskManager struct {
	shouldFail bool
}

func (m *MockRiskManager) ValidateOrder(order TradingSignal) error {
	if m.shouldFail {
		return fmt.Errorf("risk validation failed")
	}
	return nil
}

func (m *MockRiskManager) CalculatePositionSize(signal TradingSignal, availableCapital float64) float64 {
	return availableCapital * 0.02
}

func (m *MockRiskManager) CheckRiskLimits(bot *TradingBot) error {
	if m.shouldFail {
		return fmt.Errorf("risk limits exceeded")
	}
	return nil
}

func (m *MockRiskManager) UpdateRiskMetrics(bot *TradingBot, trade TradeResult) {
	// Mock implementation
}

// TestTradingBot tests the TradingBot functionality
func TestTradingBot(t *testing.T) {
	t.Run("NewTradingBot", func(t *testing.T) {
		strategy := &MockTradingStrategy{}
		config := BotConfig{
			MaxPositionSize:   1000.0,
			MaxDrawdown:       0.1,
			StopLoss:          0.05,
			TakeProfit:        0.1,
			MaxOrders:         10,
			OrderTimeout:      30 * time.Second,
			RiskPerTrade:      0.02,
			MaxDailyLoss:      100.0,
			RebalanceInterval: 24 * time.Hour,
		}

		bot := NewTradingBot("test_bot", "Test Bot", strategy, config)
		require.NotNil(t, bot)
		assert.Equal(t, "test_bot", bot.ID)
		assert.Equal(t, "Test Bot", bot.Name)
		assert.Equal(t, strategy, bot.Strategy)
		assert.Equal(t, config, bot.Config)
		assert.NotNil(t, bot.RiskManager)
		assert.NotNil(t, bot.Logger)
	})

	t.Run("StartStop", func(t *testing.T) {
		strategy := &MockTradingStrategy{}
		config := BotConfig{
			MaxPositionSize: 1000.0,
			MaxDrawdown:     0.1,
		}

		bot := NewTradingBot("test_bot", "Test Bot", strategy, config)

		// Test start
		err := bot.Start()
		require.NoError(t, err)
		assert.True(t, bot.State.IsActive)

		// Test start again (should fail)
		err = bot.Start()
		assert.Error(t, err)

		// Test stop
		err = bot.Stop()
		require.NoError(t, err)
		assert.False(t, bot.State.IsActive)

		// Test stop again (should fail)
		err = bot.Stop()
		assert.Error(t, err)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		strategy := &MockTradingStrategy{}
		config := BotConfig{
			MaxPositionSize: 1000.0,
			MaxDrawdown:     0.1,
		}

		bot := NewTradingBot("test_bot", "Test Bot", strategy, config)

		newConfig := BotConfig{
			MaxPositionSize: 2000.0,
			MaxDrawdown:     0.2,
		}

		err := bot.UpdateConfig(newConfig)
		require.NoError(t, err)
		assert.Equal(t, newConfig, bot.Config)

		// Test invalid config
		invalidConfig := BotConfig{
			MaxPositionSize: -1000.0,
			MaxDrawdown:     0.1,
		}
		err = bot.UpdateConfig(invalidConfig)
		assert.Error(t, err)
	})

	t.Run("GetPerformanceMetrics", func(t *testing.T) {
		strategy := &MockTradingStrategy{}
		config := BotConfig{
			MaxPositionSize: 1000.0,
			MaxDrawdown:     0.1,
		}

		bot := NewTradingBot("test_bot", "Test Bot", strategy, config)
		bot.State.TotalPnL = 100.0
		bot.State.DailyPnL = 50.0
		bot.State.CurrentPosition = 500.0

		metrics := bot.GetPerformanceMetrics()
		assert.Equal(t, 100.0, metrics["total_pnl"])
		assert.Equal(t, 50.0, metrics["daily_pnl"])
		assert.Equal(t, 500.0, metrics["current_position"])
	})
}

// TestStrategyEngine tests the StrategyEngine functionality
func TestStrategyEngine(t *testing.T) {
	t.Run("NewStrategyEngine", func(t *testing.T) {
		config := EngineConfig{
			MaxConcurrentStrategies: 5,
			ExecutionInterval:       1 * time.Second,
			MaxRetries:              3,
			RetryDelay:              1 * time.Second,
			EnableBacktesting:       true,
			BacktestDataPath:        "/tmp/backtest",
		}

		engine := NewStrategyEngine(config)
		require.NotNil(t, engine)
		assert.Equal(t, config, engine.config)
		assert.NotNil(t, engine.logger)
		assert.Empty(t, engine.strategies)
		assert.Empty(t, engine.executors)
	})

	t.Run("RegisterUnregisterStrategy", func(t *testing.T) {
		engine := NewStrategyEngine(EngineConfig{})
		strategy := &MockTradingStrategy{}
		executorConfig := ExecutorConfig{
			Enabled:         true,
			MaxPositionSize: 1000.0,
			RiskPerTrade:    0.02,
		}

		// Test registration
		err := engine.RegisterStrategy("test_strategy", strategy, executorConfig)
		require.NoError(t, err)
		assert.Len(t, engine.strategies, 1)
		assert.Len(t, engine.executors, 1)

		// Test duplicate registration
		err = engine.RegisterStrategy("test_strategy", strategy, executorConfig)
		assert.Error(t, err)

		// Test unregistration
		err = engine.UnregisterStrategy("test_strategy")
		require.NoError(t, err)
		assert.Empty(t, engine.strategies)
		assert.Empty(t, engine.executors)

		// Test unregistering non-existent strategy
		err = engine.UnregisterStrategy("non_existent")
		assert.Error(t, err)
	})

	t.Run("StartStopEngine", func(t *testing.T) {
		engine := NewStrategyEngine(EngineConfig{
			ExecutionInterval: 100 * time.Millisecond,
		})
		strategy := &MockTradingStrategy{}
		executorConfig := ExecutorConfig{
			Enabled:         true,
			MaxPositionSize: 1000.0,
			RiskPerTrade:    0.02,
		}

		engine.RegisterStrategy("test_strategy", strategy, executorConfig)

		// Test start
		err := engine.StartEngine()
		require.NoError(t, err)

		// Wait a bit for execution
		time.Sleep(200 * time.Millisecond)

		// Test stop
		err = engine.StopEngine()
		require.NoError(t, err)
	})

	t.Run("GetStrategyStatus", func(t *testing.T) {
		engine := NewStrategyEngine(EngineConfig{})
		strategy := &MockTradingStrategy{}
		executorConfig := ExecutorConfig{
			Enabled:         true,
			MaxPositionSize: 1000.0,
			RiskPerTrade:    0.02,
		}

		engine.RegisterStrategy("test_strategy", strategy, executorConfig)

		status := engine.GetStrategyStatus()
		assert.Len(t, status, 1)
		assert.Contains(t, status, "test_strategy")
	})
}

// TestBacktestEngine tests the BacktestEngine functionality
func TestBacktestEngine(t *testing.T) {
	t.Run("NewBacktestEngine", func(t *testing.T) {
		config := BacktestConfig{
			InitialCapital:  10000.0,
			Commission:      0.001,
			Slippage:        0.0005,
			StartDate:       time.Now().AddDate(0, -1, 0),
			EndDate:         time.Now(),
			DataInterval:    1 * time.Hour,
			EnableShorting:  false,
			MaxLeverage:     1.0,
			RiskFreeRate:    0.02,
			BenchmarkSymbol: "BTC/USDT",
		}

		engine := NewBacktestEngine(config)
		require.NotNil(t, engine)
		assert.Equal(t, config, engine.config)
		assert.NotNil(t, engine.logger)
		assert.NotNil(t, engine.results)
		assert.NotNil(t, engine.portfolio)
		assert.Equal(t, config.InitialCapital, engine.portfolio.Cash)
	})

	t.Run("LoadMarketData", func(t *testing.T) {
		// Use fixed timestamps to avoid timing issues
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		config := BacktestConfig{
			InitialCapital: 10000.0,
			StartDate:      startTime,
			EndDate:        endTime,
		}

		engine := NewBacktestEngine(config)

		// Test empty data
		err := engine.LoadMarketData([]MarketData{})
		assert.Error(t, err)

		// Test valid data with fixed timestamps
		marketData := []MarketData{
			{
				Symbol:     "BTC/USDT",
				Price:      50000.0,
				Volume:     1000000.0,
				Timestamp:  startTime,
				Bid:        49999.0,
				Ask:        50001.0,
				Spread:     2.0,
				Volatility: 0.02,
				Trend:      0.01,
			},
			{
				Symbol:     "BTC/USDT",
				Price:      51000.0,
				Volume:     1100000.0,
				Timestamp:  endTime,
				Bid:        50999.0,
				Ask:        51001.0,
				Spread:     2.0,
				Volatility: 0.025,
				Trend:      0.015,
			},
		}

		err = engine.LoadMarketData(marketData)
		require.NoError(t, err)
		assert.Len(t, engine.GetMarketData(), 2)
	})

	t.Run("RunBacktest", func(t *testing.T) {
		// Use fixed timestamps to avoid timing issues
		startTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // Noon on day 1
		endTime := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)   // Noon on day 2

		config := BacktestConfig{
			InitialCapital: 10000.0,
			Commission:     0.001,
			Slippage:       0.0005,
			StartDate:      startTime,
			EndDate:        endTime,
		}

		engine := NewBacktestEngine(config)

		// Load market data with fixed timestamps
		marketData := []MarketData{
			{
				Symbol:     "BTC/USDT",
				Price:      50000.0,
				Volume:     1000000.0,
				Timestamp:  startTime,
				Bid:        49999.0,
				Ask:        50001.0,
				Spread:     2.0,
				Volatility: 0.02,
				Trend:      0.01,
			},
			{
				Symbol:     "BTC/USDT",
				Price:      51000.0,
				Volume:     1100000.0,
				Timestamp:  endTime,
				Bid:        50999.0,
				Ask:        51001.0,
				Spread:     2.0,
				Volatility: 0.025,
				Trend:      0.015,
			},
		}

		err := engine.LoadMarketData(marketData)
		require.NoError(t, err)

		// Create mock strategy
		strategy := &MockTradingStrategy{
			signals: []TradingSignal{
				{
					Type:       SignalTypeBuy,
					Symbol:     "BTC/USDT",
					Price:      50000.0,
					Quantity:   0.1,
					Confidence: 0.8,
					Timestamp:  startTime,
					Strategy:   StrategyTypeMomentum,
				},
			},
		}

		// Run backtest
		results, err := engine.RunBacktest(strategy)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Equal(t, 2, results.TotalTrades) // Now processing 2 data points due to improved date filtering
	})
}

// TestSignalGenerator tests the SignalGenerator functionality
func TestSignalGenerator(t *testing.T) {
	t.Run("NewSignalGenerator", func(t *testing.T) {
		config := SignalConfig{
			MinConfidence:       0.5,
			MaxSignalsPerDay:    10,
			SignalCooldown:      1 * time.Hour,
			EnableFilters:       true,
			RiskThreshold:       0.3,
			VolumeThreshold:     100000.0,
			VolatilityThreshold: 0.01,
		}

		generator := NewSignalGenerator(config)
		require.NotNil(t, generator)
		assert.Equal(t, config, generator.config)
		assert.NotNil(t, generator.logger)
		assert.Empty(t, generator.indicators)
	})

	t.Run("RegisterIndicator", func(t *testing.T) {
		generator := NewSignalGenerator(SignalConfig{})

		// Mock indicator
		indicator := &MockTechnicalIndicator{
			name:          "SMA_20",
			indicatorType: IndicatorTypeTrend,
		}

		err := generator.RegisterIndicator("SMA_20", indicator)
		require.NoError(t, err)
		assert.Len(t, generator.indicators, 1)

		// Test duplicate registration
		err = generator.RegisterIndicator("SMA_20", indicator)
		assert.Error(t, err)
	})

	t.Run("GenerateSignals", func(t *testing.T) {
		config := SignalConfig{
			MinConfidence:    0.5,
			MaxSignalsPerDay: 10,
			EnableFilters:    false,
		}

		generator := NewSignalGenerator(config)

		// Register mock indicators
		sma20 := &MockTechnicalIndicator{
			name:          "SMA_20",
			indicatorType: IndicatorTypeTrend,
			value:         50000.0,
		}
		sma50 := &MockTechnicalIndicator{
			name:          "SMA_50",
			indicatorType: IndicatorTypeTrend,
			value:         49000.0,
		}

		generator.RegisterIndicator("SMA_20", sma20)
		generator.RegisterIndicator("SMA_50", sma50)

		marketData := MarketData{
			Symbol:     "BTC/USDT",
			Price:      51000.0,
			Volume:     1000000.0,
			Timestamp:  time.Now(),
			Bid:        50999.0,
			Ask:        51001.0,
			Spread:     2.0,
			Volatility: 0.02,
			Trend:      0.01,
		}

		historicalData := []MarketData{marketData}

		signals := generator.GenerateSignals(marketData, historicalData)
		assert.NotEmpty(t, signals)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		generator := NewSignalGenerator(SignalConfig{})

		newConfig := SignalConfig{
			MinConfidence:    0.7,
			MaxSignalsPerDay: 20,
			EnableFilters:    true,
		}

		err := generator.UpdateConfig(newConfig)
		require.NoError(t, err)
		assert.Equal(t, newConfig, generator.config)

		// Test invalid config
		invalidConfig := SignalConfig{
			MinConfidence:    -0.1,
			MaxSignalsPerDay: 10,
		}
		err = generator.UpdateConfig(invalidConfig)
		assert.Error(t, err)
	})
}

// MockTechnicalIndicator implements TechnicalIndicator for testing
type MockTechnicalIndicator struct {
	name          string
	indicatorType IndicatorType
	value         float64
}

func (m *MockTechnicalIndicator) Calculate(data []MarketData) float64 {
	return m.value
}

func (m *MockTechnicalIndicator) GetName() string {
	return m.name
}

func (m *MockTechnicalIndicator) GetType() IndicatorType {
	return m.indicatorType
}

func (m *MockTechnicalIndicator) Validate(data []MarketData) error {
	if len(data) == 0 {
		return fmt.Errorf("no data")
	}
	return nil
}

// TestIntegration tests the integration between components
func TestIntegration(t *testing.T) {
	t.Run("CompleteTradingFlow", func(t *testing.T) {
		// Create strategy engine
		engineConfig := EngineConfig{
			MaxConcurrentStrategies: 2,
			ExecutionInterval:       100 * time.Millisecond,
			EnableBacktesting:       true,
		}
		engine := NewStrategyEngine(engineConfig)

		// Create trading strategy
		strategy := &MockTradingStrategy{
			signals: []TradingSignal{
				{
					Type:       SignalTypeBuy,
					Symbol:     "BTC/USDT",
					Price:      50000.0,
					Quantity:   0.1,
					Confidence: 0.8,
					Timestamp:  time.Now(),
					Strategy:   StrategyTypeMomentum,
				},
			},
		}

		// Create executor config
		executorConfig := ExecutorConfig{
			Enabled:         true,
			MaxPositionSize: 1000.0,
			RiskPerTrade:    0.02,
			ExecutionDelay:  0,
			SignalThreshold: 0.5,
			MaxOrders:       5,
			OrderTimeout:    30 * time.Second,
		}

		// Register strategy
		err := engine.RegisterStrategy("test_strategy", strategy, executorConfig)
		require.NoError(t, err)

		// Start engine
		err = engine.StartEngine()
		require.NoError(t, err)

		// Wait for execution
		time.Sleep(200 * time.Millisecond)

		// Check status
		status := engine.GetStrategyStatus()
		assert.Len(t, status, 1)
		assert.Contains(t, status, "test_strategy")

		// Stop engine
		err = engine.StopEngine()
		require.NoError(t, err)
	})

	t.Run("BacktestWithStrategy", func(t *testing.T) {
		// Create backtest engine
		backtestConfig := BacktestConfig{
			InitialCapital: 10000.0,
			Commission:     0.001,
			Slippage:       0.0005,
			StartDate:      time.Now().AddDate(0, -1, 0),
			EndDate:        time.Now(),
		}
		backtestEngine := NewBacktestEngine(backtestConfig)

		// Load market data
		marketData := []MarketData{
			{
				Symbol:     "BTC/USDT",
				Price:      50000.0,
				Volume:     1000000.0,
				Timestamp:  time.Now().AddDate(0, -1, 0),
				Bid:        49999.0,
				Ask:        50001.0,
				Spread:     2.0,
				Volatility: 0.02,
				Trend:      0.01,
			},
			{
				Symbol:     "BTC/USDT",
				Price:      51000.0,
				Volume:     1100000.0,
				Timestamp:  time.Now(),
				Bid:        50999.0,
				Ask:        51001.0,
				Spread:     2.0,
				Volatility: 0.025,
				Trend:      0.015,
			},
		}

		err := backtestEngine.LoadMarketData(marketData)
		require.NoError(t, err)

		// Create strategy with signals
		strategy := &MockTradingStrategy{
			signals: []TradingSignal{
				{
					Type:       SignalTypeBuy,
					Symbol:     "BTC/USDT",
					Price:      50000.0,
					Quantity:   0.1,
					Confidence: 0.8,
					Timestamp:  time.Now().AddDate(0, -1, 0),
					Strategy:   StrategyTypeMomentum,
				},
			},
		}

		// Run backtest
		results, err := backtestEngine.RunBacktest(strategy)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Equal(t, 1, results.TotalTrades)

		// Check portfolio
		portfolio := backtestEngine.GetPortfolio()
		assert.NotNil(t, portfolio)
		// After executing a buy order, cash should be reduced but total value should be maintained
		assert.True(t, portfolio.Cash < backtestConfig.InitialCapital, "Cash should be reduced after buy order")
		assert.True(t, portfolio.TotalValue > 0, "Total value should be positive")
	})
}

// TestPerformance tests performance characteristics
func TestPerformance(t *testing.T) {
	t.Run("ConcurrentStrategyExecution", func(t *testing.T) {
		engineConfig := EngineConfig{
			MaxConcurrentStrategies: 10,
			ExecutionInterval:       10 * time.Millisecond,
		}
		engine := NewStrategyEngine(engineConfig)

		// Register multiple strategies
		for i := 0; i < 5; i++ {
			strategy := &MockTradingStrategy{
				signals: []TradingSignal{
					{
						Type:       SignalTypeBuy,
						Symbol:     "BTC/USDT",
						Price:      50000.0,
						Quantity:   0.1,
						Confidence: 0.8,
						Timestamp:  time.Now(),
						Strategy:   StrategyTypeMomentum,
					},
				},
			}

			executorConfig := ExecutorConfig{
				Enabled:         true,
				MaxPositionSize: 1000.0,
				RiskPerTrade:    0.02,
				ExecutionDelay:  0,
				SignalThreshold: 0.5,
				MaxOrders:       5,
				OrderTimeout:    30 * time.Second,
			}

			err := engine.RegisterStrategy(fmt.Sprintf("strategy_%d", i), strategy, executorConfig)
			require.NoError(t, err)
		}

		// Start engine
		err := engine.StartEngine()
		require.NoError(t, err)

		// Wait for execution
		time.Sleep(100 * time.Millisecond)

		// Check status
		status := engine.GetStrategyStatus()
		assert.Len(t, status, 5)

		// Stop engine
		err = engine.StopEngine()
		require.NoError(t, err)
	})

	t.Run("LargeDatasetBacktest", func(t *testing.T) {
		backtestConfig := BacktestConfig{
			InitialCapital: 10000.0,
			Commission:     0.001,
			Slippage:       0.0005,
			StartDate:      time.Now().AddDate(0, -1, 0),
			EndDate:        time.Now(),
		}
		backtestEngine := NewBacktestEngine(backtestConfig)

		// Generate large dataset
		marketData := make([]MarketData, 1000)
		baseTime := time.Now().AddDate(0, -1, 0)
		basePrice := 50000.0

		for i := 0; i < 1000; i++ {
			marketData[i] = MarketData{
				Symbol:     "BTC/USDT",
				Price:      basePrice + float64(i)*10.0,
				Volume:     1000000.0 + float64(i)*1000.0,
				Timestamp:  baseTime.Add(time.Duration(i) * time.Hour),
				Bid:        basePrice + float64(i)*10.0 - 1.0,
				Ask:        basePrice + float64(i)*10.0 + 1.0,
				Spread:     2.0,
				Volatility: 0.02 + float64(i)*0.0001,
				Trend:      0.01 + float64(i)*0.0001,
			}
		}

		err := backtestEngine.LoadMarketData(marketData)
		require.NoError(t, err)

		// Create strategy
		strategy := &MockTradingStrategy{
			signals: []TradingSignal{
				{
					Type:       SignalTypeBuy,
					Symbol:     "BTC/USDT",
					Price:      50000.0,
					Quantity:   0.1,
					Confidence: 0.8,
					Timestamp:  time.Now(),
					Strategy:   StrategyTypeMomentum,
				},
			},
		}

		// Run backtest and measure performance
		start := time.Now()
		results, err := backtestEngine.RunBacktest(strategy)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, marketData, 1000)

		// Performance assertion: should complete within reasonable time
		assert.Less(t, duration, 5*time.Second, "Backtest should complete within 5 seconds")
	})
}
