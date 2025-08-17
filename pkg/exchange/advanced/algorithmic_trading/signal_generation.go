package algorithmictrading

import (
	"fmt"
	"math"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// SignalGenerator generates trading signals based on market analysis
type SignalGenerator struct {
	indicators map[string]TechnicalIndicator
	config     SignalConfig
	logger     *logger.Logger
}

// SignalConfig contains configuration for signal generation
type SignalConfig struct {
	MinConfidence       float64       `json:"min_confidence"`
	MaxSignalsPerDay    int           `json:"max_signals_per_day"`
	SignalCooldown      time.Duration `json:"signal_cooldown"`
	EnableFilters       bool          `json:"enable_filters"`
	RiskThreshold       float64       `json:"risk_threshold"`
	VolumeThreshold     float64       `json:"volume_threshold"`
	VolatilityThreshold float64       `json:"volatility_threshold"`
}

// TechnicalIndicator represents a technical analysis indicator
type TechnicalIndicator interface {
	Calculate(data []MarketData) float64
	GetName() string
	GetType() IndicatorType
	Validate(data []MarketData) error
}

// IndicatorType represents the type of technical indicator
type IndicatorType string

const (
	IndicatorTypeTrend      IndicatorType = "trend"
	IndicatorTypeMomentum   IndicatorType = "momentum"
	IndicatorTypeVolatility IndicatorType = "volatility"
	IndicatorTypeVolume     IndicatorType = "volume"
	IndicatorTypeOscillator IndicatorType = "oscillator"
)

// NewSignalGenerator creates a new signal generator
func NewSignalGenerator(config SignalConfig) *SignalGenerator {
	return &SignalGenerator{
		indicators: make(map[string]TechnicalIndicator),
		config:     config,
		logger:     logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "signal_generator"}),
	}
}

// RegisterIndicator registers a technical indicator
func (sg *SignalGenerator) RegisterIndicator(name string, indicator TechnicalIndicator) error {
	if _, exists := sg.indicators[name]; exists {
		return fmt.Errorf("indicator %s already registered", name)
	}

	sg.indicators[name] = indicator
	sg.logger.Info("Technical indicator registered - name: %s, type: %s", name, indicator.GetType())
	return nil
}

// GenerateSignals generates trading signals based on market data and indicators
func (sg *SignalGenerator) GenerateSignals(marketData MarketData, historicalData []MarketData) []TradingSignal {
	if len(historicalData) == 0 {
		return nil
	}

	signals := make([]TradingSignal, 0)

	// Calculate indicator values
	indicatorValues := sg.calculateIndicators(historicalData)

	// Generate signals based on indicator combinations
	signals = append(signals, sg.generateTrendSignals(marketData, indicatorValues)...)
	signals = append(signals, sg.generateMomentumSignals(marketData, indicatorValues)...)
	signals = append(signals, sg.generateMeanReversionSignals(marketData, indicatorValues)...)
	signals = append(signals, sg.generateBreakoutSignals(marketData, indicatorValues)...)

	// Apply filters
	if sg.config.EnableFilters {
		signals = sg.applyFilters(signals, marketData)
	}

	// Limit number of signals
	if len(signals) > sg.config.MaxSignalsPerDay {
		signals = signals[:sg.config.MaxSignalsPerDay]
	}

	sg.logger.Debug("Generated signals - count: %d, symbol: %s", len(signals), marketData.Symbol)
	return signals
}

// calculateIndicators calculates values for all registered indicators
func (sg *SignalGenerator) calculateIndicators(data []MarketData) map[string]float64 {
	values := make(map[string]float64)

	for name, indicator := range sg.indicators {
		if err := indicator.Validate(data); err != nil {
			sg.logger.Warn("Indicator validation failed - indicator: %s, error: %v", name, err)
			continue
		}

		value := indicator.Calculate(data)
		values[name] = value
	}

	return values
}

// generateTrendSignals generates signals based on trend indicators
func (sg *SignalGenerator) generateTrendSignals(marketData MarketData, indicators map[string]float64) []TradingSignal {
	signals := make([]TradingSignal, 0)

	// Example: Moving Average Crossover
	if sma20, exists := indicators["SMA_20"]; exists {
		if sma50, exists := indicators["SMA_50"]; exists {
			price := marketData.Price

			// Golden Cross (SMA 20 crosses above SMA 50)
			if price > sma20 && sma20 > sma50 {
				confidence := sg.calculateTrendConfidence(price, sma20, sma50)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeBuy,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator": "SMA_Crossover",
							"sma_20":    sma20,
							"sma_50":    sma50,
							"signal":    "golden_cross",
						},
					})
				}
			}

			// Death Cross (SMA 20 crosses below SMA 50)
			if price < sma20 && sma20 < sma50 {
				confidence := sg.calculateTrendConfidence(price, sma20, sma50)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeSell,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator": "SMA_Crossover",
							"sma_20":    sma20,
							"sma_50":    sma50,
							"signal":    "death_cross",
						},
					})
				}
			}
		}
	}

	return signals
}

// generateMomentumSignals generates signals based on momentum indicators
func (sg *SignalGenerator) generateMomentumSignals(marketData MarketData, indicators map[string]float64) []TradingSignal {
	signals := make([]TradingSignal, 0)

	// Example: RSI (Relative Strength Index)
	if rsi, exists := indicators["RSI"]; exists {
		price := marketData.Price

		// Oversold condition (RSI < 30)
		if rsi < 30 {
			confidence := sg.calculateRSIConfidence(rsi, 30)
			if confidence >= sg.config.MinConfidence {
				signals = append(signals, TradingSignal{
					Type:       SignalTypeBuy,
					Symbol:     marketData.Symbol,
					Price:      price,
					Quantity:   0, // Will be calculated by strategy
					Confidence: confidence,
					Timestamp:  marketData.Timestamp,
					Strategy:   StrategyTypeMeanReversion,
					Metadata: map[string]interface{}{
						"indicator": "RSI",
						"value":     rsi,
						"signal":    "oversold",
						"threshold": 30,
					},
				})
			}
		}

		// Overbought condition (RSI > 70)
		if rsi > 70 {
			confidence := sg.calculateRSIConfidence(rsi, 70)
			if confidence >= sg.config.MinConfidence {
				signals = append(signals, TradingSignal{
					Type:       SignalTypeSell,
					Symbol:     marketData.Symbol,
					Price:      price,
					Quantity:   0, // Will be calculated by strategy
					Confidence: confidence,
					Timestamp:  marketData.Timestamp,
					Strategy:   StrategyTypeMeanReversion,
					Metadata: map[string]interface{}{
						"indicator": "RSI",
						"value":     rsi,
						"signal":    "overbought",
						"threshold": 70,
					},
				})
			}
		}
	}

	// Example: MACD (Moving Average Convergence Divergence)
	if macd, exists := indicators["MACD"]; exists {
		if macdSignal, exists := indicators["MACD_Signal"]; exists {
			price := marketData.Price

			// MACD crosses above signal line
			if macd > macdSignal {
				confidence := sg.calculateMACDConfidence(macd, macdSignal)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeBuy,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator":   "MACD",
							"macd":        macd,
							"signal":      macdSignal,
							"signal_type": "bullish_cross",
						},
					})
				}
			}

			// MACD crosses below signal line
			if macd < macdSignal {
				confidence := sg.calculateMACDConfidence(macd, macdSignal)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeSell,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator":   "MACD",
							"macd":        macd,
							"signal":      macdSignal,
							"signal_type": "bearish_cross",
						},
					})
				}
			}
		}
	}

	return signals
}

// generateMeanReversionSignals generates signals based on mean reversion indicators
func (sg *SignalGenerator) generateMeanReversionSignals(marketData MarketData, indicators map[string]float64) []TradingSignal {
	signals := make([]TradingSignal, 0)

	// Example: Bollinger Bands
	if bbUpper, exists := indicators["BB_Upper"]; exists {
		if bbLower, exists := indicators["BB_Lower"]; exists {
			if bbMiddle, exists := indicators["BB_Middle"]; exists {
				price := marketData.Price

				// Price touches lower band (oversold)
				if price <= bbLower {
					confidence := sg.calculateBBConfidence(price, bbLower, bbMiddle)
					if confidence >= sg.config.MinConfidence {
						signals = append(signals, TradingSignal{
							Type:       SignalTypeBuy,
							Symbol:     marketData.Symbol,
							Price:      price,
							Quantity:   0, // Will be calculated by strategy
							Confidence: confidence,
							Timestamp:  marketData.Timestamp,
							Strategy:   StrategyTypeMeanReversion,
							Metadata: map[string]interface{}{
								"indicator": "Bollinger_Bands",
								"price":     price,
								"bb_lower":  bbLower,
								"bb_middle": bbMiddle,
								"signal":    "oversold",
							},
						})
					}
				}

				// Price touches upper band (overbought)
				if price >= bbUpper {
					confidence := sg.calculateBBConfidence(price, bbUpper, bbMiddle)
					if confidence >= sg.config.MinConfidence {
						signals = append(signals, TradingSignal{
							Type:       SignalTypeSell,
							Symbol:     marketData.Symbol,
							Price:      price,
							Quantity:   0, // Will be calculated by strategy
							Confidence: confidence,
							Timestamp:  marketData.Timestamp,
							Strategy:   StrategyTypeMeanReversion,
							Metadata: map[string]interface{}{
								"indicator": "Bollinger_Bands",
								"price":     price,
								"bb_upper":  bbUpper,
								"bb_middle": bbMiddle,
								"signal":    "overbought",
							},
						})
					}
				}
			}
		}
	}

	return signals
}

// generateBreakoutSignals generates signals based on breakout patterns
func (sg *SignalGenerator) generateBreakoutSignals(marketData MarketData, indicators map[string]float64) []TradingSignal {
	signals := make([]TradingSignal, 0)

	// Example: Support/Resistance Breakout
	if support, exists := indicators["Support_Level"]; exists {
		if resistance, exists := indicators["Resistance_Level"]; exists {
			price := marketData.Price

			// Breakout above resistance
			if price > resistance {
				confidence := sg.calculateBreakoutConfidence(price, resistance, marketData.Volume)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeBuy,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator":  "Support_Resistance",
							"price":      price,
							"resistance": resistance,
							"signal":     "breakout_up",
						},
					})
				}
			}

			// Breakdown below support
			if price < support {
				confidence := sg.calculateBreakoutConfidence(price, support, marketData.Volume)
				if confidence >= sg.config.MinConfidence {
					signals = append(signals, TradingSignal{
						Type:       SignalTypeSell,
						Symbol:     marketData.Symbol,
						Price:      price,
						Quantity:   0, // Will be calculated by strategy
						Confidence: confidence,
						Timestamp:  marketData.Timestamp,
						Strategy:   StrategyTypeMomentum,
						Metadata: map[string]interface{}{
							"indicator": "Support_Resistance",
							"price":     price,
							"support":   support,
							"signal":    "breakdown_down",
						},
					})
				}
			}
		}
	}

	return signals
}

// applyFilters applies various filters to the generated signals
func (sg *SignalGenerator) applyFilters(signals []TradingSignal, marketData MarketData) []TradingSignal {
	filtered := make([]TradingSignal, 0)

	for _, signal := range signals {
		// Volume filter
		if marketData.Volume < sg.config.VolumeThreshold {
			continue
		}

		// Volatility filter
		if marketData.Volatility < sg.config.VolatilityThreshold {
			continue
		}

		// Risk filter
		if signal.Confidence < sg.config.RiskThreshold {
			continue
		}

		filtered = append(filtered, signal)
	}

	return filtered
}

// Confidence calculation methods
func (sg *SignalGenerator) calculateTrendConfidence(price, sma20, sma50 float64) float64 {
	// Calculate confidence based on how far price is from moving averages
	priceToSMA20 := math.Abs(price-sma20) / sma20
	priceToSMA50 := math.Abs(price-sma50) / sma50

	// Higher confidence when price is closer to moving averages
	confidence := 1.0 - (priceToSMA20+priceToSMA50)/2.0
	return math.Max(0.1, math.Min(1.0, confidence))
}

func (sg *SignalGenerator) calculateRSIConfidence(rsi, threshold float64) float64 {
	// Calculate confidence based on how extreme the RSI value is
	distance := math.Abs(rsi-50) / 50.0
	confidence := distance * 1.5 // Scale to 0-1 range
	return math.Max(0.1, math.Min(1.0, confidence))
}

func (sg *SignalGenerator) calculateMACDConfidence(macd, signal float64) float64 {
	// Calculate confidence based on MACD divergence from signal line
	divergence := math.Abs(macd-signal) / math.Abs(signal)
	confidence := divergence * 2.0 // Scale to 0-1 range
	return math.Max(0.1, math.Min(1.0, confidence))
}

func (sg *SignalGenerator) calculateBBConfidence(price, band, middle float64) float64 {
	// Calculate confidence based on distance from middle band
	distance := math.Abs(price-middle) / middle
	confidence := distance * 2.0 // Scale to 0-1 range
	return math.Max(0.1, math.Min(1.0, confidence))
}

func (sg *SignalGenerator) calculateBreakoutConfidence(price, level, volume float64) float64 {
	// Calculate confidence based on breakout strength and volume
	breakoutStrength := math.Abs(price-level) / level
	volumeFactor := math.Min(volume/1000000.0, 1.0) // Normalize volume

	confidence := (breakoutStrength + volumeFactor) / 2.0
	return math.Max(0.1, math.Min(1.0, confidence))
}

// GetIndicators returns all registered indicators
func (sg *SignalGenerator) GetIndicators() map[string]TechnicalIndicator {
	return sg.indicators
}

// UpdateConfig updates the signal generator configuration
func (sg *SignalGenerator) UpdateConfig(config SignalConfig) error {
	if config.MinConfidence < 0 || config.MinConfidence > 1 {
		return fmt.Errorf("min confidence must be between 0 and 1")
	}
	if config.MaxSignalsPerDay <= 0 {
		return fmt.Errorf("max signals per day must be positive")
	}

	sg.config = config
	sg.logger.Info("Configuration updated - config: %+v", config)
	return nil
}
