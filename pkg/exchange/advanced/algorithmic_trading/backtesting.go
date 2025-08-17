package algorithmictrading

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// BacktestEngine performs backtesting of trading strategies
type BacktestEngine struct {
	config     BacktestConfig
	logger     *logger.Logger
	results    *BacktestResults
	portfolio  *Portfolio
	marketData []MarketData
}

// BacktestConfig contains configuration for backtesting
type BacktestConfig struct {
	InitialCapital  float64       `json:"initial_capital"`
	Commission      float64       `json:"commission"` // Commission per trade (e.g., 0.001 for 0.1%)
	Slippage        float64       `json:"slippage"`   // Slippage per trade (e.g., 0.0005 for 0.05%)
	StartDate       time.Time     `json:"start_date"`
	EndDate         time.Time     `json:"end_date"`
	DataInterval    time.Duration `json:"data_interval"`
	EnableShorting  bool          `json:"enable_shorting"`
	MaxLeverage     float64       `json:"max_leverage"`
	RiskFreeRate    float64       `json:"risk_free_rate"`   // Annual risk-free rate
	BenchmarkSymbol string        `json:"benchmark_symbol"` // Benchmark for comparison
}

// BacktestResults contains the results of a backtest
type BacktestResults struct {
	TotalReturn        float64            `json:"total_return"`
	AnnualizedReturn   float64            `json:"annualized_return"`
	SharpeRatio        float64            `json:"sharpe_ratio"`
	MaxDrawdown        float64            `json:"max_drawdown"`
	CalmarRatio        float64            `json:"calmar_ratio"`
	SortinoRatio       float64            `json:"sortino_ratio"`
	VaR                float64            `json:"var"`  // Value at Risk (95%)
	CVaR               float64            `json:"cvar"` // Conditional Value at Risk
	WinRate            float64            `json:"win_rate"`
	ProfitFactor       float64            `json:"profit_factor"`
	TotalTrades        int                `json:"total_trades"`
	WinningTrades      int                `json:"winning_trades"`
	LosingTrades       int                `json:"losing_trades"`
	AverageWin         float64            `json:"average_win"`
	AverageLoss        float64            `json:"average_loss"`
	LargestWin         float64            `json:"largest_win"`
	LargestLoss        float64            `json:"largest_loss"`
	EquityCurve        []EquityPoint      `json:"equity_curve"`
	DrawdownCurve      []DrawdownPoint    `json:"drawdown_curve"`
	MonthlyReturns     map[string]float64 `json:"monthly_returns"`
	TradeHistory       []BacktestTrade    `json:"trade_history"`
	BenchmarkReturns   []BenchmarkReturn  `json:"benchmark_returns"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
}

// Portfolio represents the trading portfolio during backtesting
type Portfolio struct {
	Cash          float64            `json:"cash"`
	Positions     map[string]float64 `json:"positions"`
	TotalValue    float64            `json:"total_value"`
	UnrealizedPnL float64            `json:"unrealized_pnl"`
	RealizedPnL   float64            `json:"realized_pnl"`
	EquityHistory []EquityPoint      `json:"equity_history"`
}

// EquityPoint represents a point on the equity curve
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Cash      float64   `json:"cash"`
	Positions float64   `json:"positions"`
}

// DrawdownPoint represents a point on the drawdown curve
type DrawdownPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Drawdown  float64   `json:"drawdown"`
	Peak      float64   `json:"peak"`
}

// BacktestTrade represents a trade during backtesting
type BacktestTrade struct {
	ID         string        `json:"id"`
	Symbol     string        `json:"symbol"`
	Type       SignalType    `json:"type"`
	Price      float64       `json:"price"`
	Quantity   float64       `json:"quantity"`
	Timestamp  time.Time     `json:"timestamp"`
	Commission float64       `json:"commission"`
	Slippage   float64       `json:"slippage"`
	TotalCost  float64       `json:"total_cost"`
	PnL        float64       `json:"pnl"`
	ExitPrice  float64       `json:"exit_price,omitempty"`
	ExitTime   time.Time     `json:"exit_time,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

// BenchmarkReturn represents benchmark performance data
type BenchmarkReturn struct {
	Timestamp time.Time `json:"timestamp"`
	Return    float64   `json:"return"`
	Value     float64   `json:"value"`
}

// PerformanceMetrics contains additional performance metrics
type PerformanceMetrics struct {
	Beta              float64 `json:"beta"`
	Alpha             float64 `json:"alpha"`
	InformationRatio  float64 `json:"information_ratio"`
	TreynorRatio      float64 `json:"treynor_ratio"`
	JensenAlpha       float64 `json:"jensen_alpha"`
	TrackingError     float64 `json:"tracking_error"`
	DownsideDeviation float64 `json:"downside_deviation"`
	UpsideCapture     float64 `json:"upside_capture"`
	DownsideCapture   float64 `json:"downside_capture"`
}

// NewBacktestEngine creates a new backtest engine
func NewBacktestEngine(config BacktestConfig) *BacktestEngine {
	return &BacktestEngine{
		config: config,
		logger: logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "backtest_engine"}),
		results: &BacktestResults{
			EquityCurve:      make([]EquityPoint, 0),
			DrawdownCurve:    make([]DrawdownPoint, 0),
			MonthlyReturns:   make(map[string]float64),
			TradeHistory:     make([]BacktestTrade, 0),
			BenchmarkReturns: make([]BenchmarkReturn, 0),
		},
		portfolio: &Portfolio{
			Cash:          config.InitialCapital,
			Positions:     make(map[string]float64),
			TotalValue:    config.InitialCapital,
			EquityHistory: make([]EquityPoint, 0),
		},
		marketData: make([]MarketData, 0),
	}
}

// LoadMarketData loads historical market data for backtesting
func (engine *BacktestEngine) LoadMarketData(data []MarketData) error {
	if len(data) == 0 {
		return fmt.Errorf("no market data provided")
	}

	// Sort data by timestamp
	sort.Slice(data, func(i, j int) bool {
		return data[i].Timestamp.Before(data[j].Timestamp)
	})

			// Filter data within date range (inclusive)
		filteredData := make([]MarketData, 0)
		for _, point := range data {
			// Include data points that are within or exactly at the start/end dates
			// Use a small tolerance for timestamp comparison to handle microsecond precision
			startTolerance := point.Timestamp.Sub(engine.config.StartDate)
			endTolerance := engine.config.EndDate.Sub(point.Timestamp)
			
			if startTolerance >= -time.Microsecond && endTolerance >= -time.Microsecond {
				filteredData = append(filteredData, point)
			}
		}

	if len(filteredData) == 0 {
		return fmt.Errorf("no market data within specified date range")
	}

	engine.marketData = filteredData
	engine.logger.Info("Market data loaded - data_points: %d, start_date: %s, end_date: %s", len(filteredData), engine.config.StartDate, engine.config.EndDate)
	return nil
}

// RunBacktest runs a complete backtest for a trading strategy
func (engine *BacktestEngine) RunBacktest(strategy TradingStrategy) (*BacktestResults, error) {
	if len(engine.marketData) == 0 {
		return nil, fmt.Errorf("no market data loaded")
	}

	engine.logger.Info("Starting backtest - strategy: %s, data_points: %d", strategy.GetStrategyType(), len(engine.marketData))

	// Initialize portfolio
	engine.portfolio.Cash = engine.config.InitialCapital
	engine.portfolio.TotalValue = engine.config.InitialCapital
	engine.portfolio.EquityHistory = []EquityPoint{{
		Timestamp: engine.config.StartDate,
		Value:     engine.config.InitialCapital,
		Cash:      engine.config.InitialCapital,
		Positions: 0,
	}}

	// Track peak value for drawdown calculation
	peakValue := engine.config.InitialCapital
	currentDrawdown := 0.0

	// Process each market data point
	for i, dataPoint := range engine.marketData {
		// Generate trading signals
		signals := strategy.GenerateSignals(dataPoint)

		// Execute signals
		for _, signal := range signals {
			if err := engine.executeSignal(signal, dataPoint, strategy); err != nil {
				engine.logger.Warn("Failed to execute signal - error: %v, signal: %+v", err, signal)
			}
		}

		// Update portfolio value
		engine.updatePortfolioValue(dataPoint)

		// Record equity point
		equityPoint := EquityPoint{
			Timestamp: dataPoint.Timestamp,
			Value:     engine.portfolio.TotalValue,
			Cash:      engine.portfolio.Cash,
			Positions: engine.portfolio.TotalValue - engine.portfolio.Cash,
		}
		engine.portfolio.EquityHistory = append(engine.portfolio.EquityHistory, equityPoint)

		// Update drawdown
		if engine.portfolio.TotalValue > peakValue {
			peakValue = engine.portfolio.TotalValue
		}
		currentDrawdown = (peakValue - engine.portfolio.TotalValue) / peakValue

		// Record drawdown point
		drawdownPoint := DrawdownPoint{
			Timestamp: dataPoint.Timestamp,
			Drawdown:  currentDrawdown,
			Peak:      peakValue,
		}
		engine.results.DrawdownCurve = append(engine.results.DrawdownCurve, drawdownPoint)

		// Update max drawdown
		if currentDrawdown > engine.results.MaxDrawdown {
			engine.results.MaxDrawdown = currentDrawdown
		}

		// Record monthly returns
		monthKey := dataPoint.Timestamp.Format("2006-01")
		if i > 0 {
			prevValue := engine.portfolio.EquityHistory[i-1].Value
			monthlyReturn := (engine.portfolio.TotalValue - prevValue) / prevValue
			engine.results.MonthlyReturns[monthKey] = monthlyReturn
		}
	}

	// Calculate final results
	engine.calculateResults()

	engine.logger.Info("Backtest completed - total_return: %.4f, max_drawdown: %.4f, total_trades: %d",
		engine.results.TotalReturn, engine.results.MaxDrawdown, engine.results.TotalTrades)

	return engine.results, nil
}

// executeSignal executes a trading signal during backtesting
func (engine *BacktestEngine) executeSignal(signal TradingSignal, marketData MarketData, strategy TradingStrategy) error {
	// Validate signal
	if err := strategy.ValidateSignal(signal); err != nil {
		return fmt.Errorf("invalid signal: %w", err)
	}

	// Calculate execution price with slippage
	executionPrice := signal.Price
	if signal.Type == SignalTypeBuy {
		executionPrice *= (1 + engine.config.Slippage)
	} else if signal.Type == SignalTypeSell {
		executionPrice *= (1 - engine.config.Slippage)
	}

	// Calculate quantity based on available capital
	availableCapital := engine.portfolio.Cash
	if signal.Type == SignalTypeSell {
		// For selling, check if we have the position
		if position, exists := engine.portfolio.Positions[signal.Symbol]; !exists || position < signal.Quantity {
			return fmt.Errorf("insufficient position to sell %s", signal.Symbol)
		}
		availableCapital = engine.portfolio.Positions[signal.Symbol] * executionPrice
	}

	// Calculate position size - use a simple calculation for backtesting
	positionSize := availableCapital * 0.02 // 2% risk per trade
	if positionSize > availableCapital {
		positionSize = availableCapital
	}

	// Calculate actual quantity
	quantity := positionSize / executionPrice

	// Calculate costs
	commission := positionSize * engine.config.Commission
	totalCost := positionSize + commission

	// Execute trade
	trade := BacktestTrade{
		ID:         fmt.Sprintf("trade_%d", len(engine.results.TradeHistory)+1),
		Symbol:     signal.Symbol,
		Type:       signal.Type,
		Price:      executionPrice,
		Quantity:   quantity,
		Timestamp:  marketData.Timestamp,
		Commission: commission,
		Slippage:   positionSize * engine.config.Slippage,
		TotalCost:  totalCost,
	}

	// Update portfolio
	if signal.Type == SignalTypeBuy {
		if engine.portfolio.Cash < totalCost {
			return fmt.Errorf("insufficient cash for buy order")
		}
		engine.portfolio.Cash -= totalCost
		engine.portfolio.Positions[signal.Symbol] += quantity
	} else if signal.Type == SignalTypeSell {
		if engine.portfolio.Positions[signal.Symbol] < quantity {
			return fmt.Errorf("insufficient position for sell order")
		}
		engine.portfolio.Cash += positionSize - commission
		engine.portfolio.Positions[signal.Symbol] -= quantity
	}

	// Record trade
	engine.results.TradeHistory = append(engine.results.TradeHistory, trade)
	engine.results.TotalTrades++

	return nil
}

// updatePortfolioValue updates the portfolio value based on current market prices
func (engine *BacktestEngine) updatePortfolioValue(marketData MarketData) {
	positionsValue := 0.0

	// Calculate current value of all positions
	for symbol, quantity := range engine.portfolio.Positions {
		if symbol == marketData.Symbol {
			positionsValue += quantity * marketData.Price
		}
	}

	engine.portfolio.TotalValue = engine.portfolio.Cash + positionsValue
	engine.portfolio.UnrealizedPnL = positionsValue - engine.portfolio.Cash
}

// calculateResults calculates all performance metrics
func (engine *BacktestEngine) calculateResults() {
	if len(engine.portfolio.EquityHistory) < 2 {
		return
	}

	// Calculate total return
	initialValue := engine.portfolio.EquityHistory[0].Value
	finalValue := engine.portfolio.EquityHistory[len(engine.portfolio.EquityHistory)-1].Value
	engine.results.TotalReturn = (finalValue - initialValue) / initialValue

	// Calculate annualized return
	duration := engine.config.EndDate.Sub(engine.config.StartDate)
	years := duration.Hours() / 8760 // 8760 hours in a year
	if years > 0 {
		engine.results.AnnualizedReturn = math.Pow(1+engine.results.TotalReturn, 1/years) - 1
	}

	// Calculate trade statistics
	engine.calculateTradeStatistics()

	// Calculate risk metrics
	engine.calculateRiskMetrics()

	// Calculate performance ratios
	engine.calculatePerformanceRatios()
}

// calculateTradeStatistics calculates trade-related statistics
func (engine *BacktestEngine) calculateTradeStatistics() {
	if len(engine.results.TradeHistory) == 0 {
		return
	}

	var totalPnL, totalWins, totalLosses float64
	var wins, losses int

	for _, trade := range engine.results.TradeHistory {
		if trade.PnL > 0 {
			totalWins += trade.PnL
			wins++
		} else if trade.PnL < 0 {
			totalLosses += math.Abs(trade.PnL)
			losses++
		}
		totalPnL += trade.PnL
	}

	engine.results.WinningTrades = wins
	engine.results.LosingTrades = losses
	engine.results.TotalTrades = len(engine.results.TradeHistory)

	if wins > 0 {
		engine.results.AverageWin = totalWins / float64(wins)
	}
	if losses > 0 {
		engine.results.AverageLoss = totalLosses / float64(losses)
	}

	engine.results.WinRate = float64(wins) / float64(engine.results.TotalTrades)

	if totalLosses > 0 {
		engine.results.ProfitFactor = totalWins / totalLosses
	}
}

// calculateRiskMetrics calculates risk-related metrics
func (engine *BacktestEngine) calculateRiskMetrics() {
	if len(engine.portfolio.EquityHistory) < 2 {
		return
	}

	// Calculate returns
	returns := make([]float64, 0, len(engine.portfolio.EquityHistory)-1)
	for i := 1; i < len(engine.portfolio.EquityHistory); i++ {
		prevValue := engine.portfolio.EquityHistory[i-1].Value
		currentValue := engine.portfolio.EquityHistory[i].Value
		return_ := (currentValue - prevValue) / prevValue
		returns = append(returns, return_)
	}

	// Calculate Sharpe ratio
	if len(returns) > 0 {
		meanReturn := calculateMean(returns)
		stdDev := calculateStdDev(returns, meanReturn)
		if stdDev > 0 {
			engine.results.SharpeRatio = (meanReturn - engine.config.RiskFreeRate/252) / stdDev * math.Sqrt(252)
		}
	}

	// Calculate VaR (95% confidence)
	if len(returns) > 0 {
		sort.Float64s(returns)
		varIndex := int(float64(len(returns)) * 0.05)
		if varIndex < len(returns) {
			engine.results.VaR = math.Abs(returns[varIndex])
		}
	}

	// Calculate Calmar ratio
	if engine.results.MaxDrawdown > 0 && engine.results.AnnualizedReturn > 0 {
		engine.results.CalmarRatio = engine.results.AnnualizedReturn / engine.results.MaxDrawdown
	}
}

// calculatePerformanceRatios calculates additional performance ratios
func (engine *BacktestEngine) calculatePerformanceRatios() {
	// Calculate Sortino ratio (downside deviation)
	if len(engine.portfolio.EquityHistory) > 1 {
		returns := make([]float64, 0, len(engine.portfolio.EquityHistory)-1)
		for i := 1; i < len(engine.portfolio.EquityHistory); i++ {
			prevValue := engine.portfolio.EquityHistory[i-1].Value
			currentValue := engine.portfolio.EquityHistory[i].Value
			return_ := (currentValue - prevValue) / prevValue
			returns = append(returns, return_)
		}

		meanReturn := calculateMean(returns)
		downsideDeviation := calculateDownsideDeviation(returns, meanReturn)

		if downsideDeviation > 0 {
			engine.results.SortinoRatio = (meanReturn - engine.config.RiskFreeRate/252) / downsideDeviation * math.Sqrt(252)
		}
	}
}

// Helper functions for calculations
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	return math.Sqrt(sum / float64(len(values)))
}

func calculateDownsideDeviation(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	count := 0
	for _, v := range values {
		if v < mean {
			sum += math.Pow(v-mean, 2)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(count))
}

// GetResults returns the backtest results
func (engine *BacktestEngine) GetResults() *BacktestResults {
	return engine.results
}

// GetPortfolio returns the current portfolio state
func (engine *BacktestEngine) GetPortfolio() *Portfolio {
	return engine.portfolio
}

// GetMarketData returns the loaded market data
func (engine *BacktestEngine) GetMarketData() []MarketData {
	return engine.marketData
}

// Reset resets the backtest engine for a new run
func (engine *BacktestEngine) Reset() {
	engine.results = &BacktestResults{
		EquityCurve:      make([]EquityPoint, 0),
		DrawdownCurve:    make([]DrawdownPoint, 0),
		MonthlyReturns:   make(map[string]float64),
		TradeHistory:     make([]BacktestTrade, 0),
		BenchmarkReturns: make([]BenchmarkReturn, 0),
	}
	engine.portfolio = &Portfolio{
		Cash:          engine.config.InitialCapital,
		Positions:     make(map[string]float64),
		TotalValue:    engine.config.InitialCapital,
		EquityHistory: make([]EquityPoint, 0),
	}
	engine.marketData = make([]MarketData, 0)
}
