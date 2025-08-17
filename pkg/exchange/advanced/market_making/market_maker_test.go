package marketmaking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMarketMaker tests the MarketMaker functionality
func TestMarketMaker(t *testing.T) {
	t.Run("NewMarketMaker", func(t *testing.T) {
		config := MarketMakerConfig{
			MaxPositionSize:  1.0,
			MaxSpread:        0.02,
			MinSpread:        0.001,
			QuoteSize:        0.1,
			QuoteRefreshRate: 1 * time.Second,
		}
		
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		
		mm := NewMarketMaker("mm_1", strategy, config)
		require.NotNil(t, mm)
		assert.Equal(t, "mm_1", mm.ID)
		assert.Equal(t, config, mm.Config)
		assert.Equal(t, strategy, mm.Strategy)
		assert.NotNil(t, mm.Logger)
		assert.NotNil(t, mm.Quotes)
		assert.NotNil(t, mm.Position)
	})

	t.Run("StartStop", func(t *testing.T) {
		config := MarketMakerConfig{
			MaxPositionSize:  1.0,
			MaxSpread:        0.02,
			MinSpread:        0.001,
			QuoteSize:        0.1,
			QuoteRefreshRate: 100 * time.Millisecond,
		}
		
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		mm := NewMarketMaker("mm_1", strategy, config)
		
		// Start market maker
		err := mm.Start()
		require.NoError(t, err)
		
		// Wait a bit for quotes to be generated
		time.Sleep(150 * time.Millisecond)
		
		// Check if quotes were generated
		quotes, exists := mm.Quotes["BTC/USDT"]
		assert.True(t, exists)
		assert.NotNil(t, quotes)
		
		// Stop market maker
		err = mm.Stop()
		require.NoError(t, err)
	})

	t.Run("UpdatePosition", func(t *testing.T) {
		config := MarketMakerConfig{
			MaxPositionSize:  1.0,
			MaxSpread:        0.02,
			MinSpread:        0.001,
			QuoteSize:        0.1,
			QuoteRefreshRate: 1 * time.Second,
		}
		
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		mm := NewMarketMaker("mm_1", strategy, config)
		
		// Initial position should be zero
		assert.Equal(t, 0.0, mm.Position.Quantity)
		assert.Equal(t, 0.0, mm.Position.AveragePrice)
		
		// Update position with a trade (buying)
		trade := Trade{
			Symbol:   "BTC/USDT",
			Side:     "buy",
			Quantity: 0.1,
			Price:    50000.0,
		}
		
		mm.UpdatePosition(trade)
		
		// Check position update
		assert.Equal(t, 0.1, mm.Position.Quantity)
		assert.Equal(t, 50000.0, mm.Position.AveragePrice)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		config := MarketMakerConfig{
			MaxPositionSize:  1.0,
			MaxSpread:        0.02,
			MinSpread:        0.001,
			QuoteSize:        0.1,
			QuoteRefreshRate: 1 * time.Second,
		}
		
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		mm := NewMarketMaker("mm_1", strategy, config)
		
		// Update configuration
		newConfig := MarketMakerConfig{
			MaxPositionSize:  2.0,
			MaxSpread:        0.03,
			MinSpread:        0.001,
			QuoteSize:        0.2,
			QuoteRefreshRate: 2 * time.Second,
		}
		
		err := mm.UpdateConfig(newConfig)
		require.NoError(t, err)
		assert.Equal(t, newConfig, mm.Config)
	})
}

// TestBasicMarketMaking tests the BasicMarketMaking strategy
func TestBasicMarketMaking(t *testing.T) {
	t.Run("NewBasicMarketMaking", func(t *testing.T) {
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		require.NotNil(t, strategy)
		assert.Equal(t, 1.5, strategy.SpreadMultiplier)
		assert.Equal(t, 1.0, strategy.PositionLimit)
		assert.Equal(t, 0.1, strategy.QuoteSize)
		assert.Equal(t, "basic_market_making", strategy.GetStrategyType())
	})

	t.Run("CalculateQuotes", func(t *testing.T) {
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		
		marketData := MarketData{
			Symbol: "BTC/USDT",
			Bid:    49999.0,
			Ask:    50001.0,
			MidPrice: 50000.0,
			Spread: 2.0,
			Volume: 1000000.0,
		}
		
		position := Position{
			Symbol:      "BTC/USDT",
			Quantity:    0.1,
			AveragePrice: 50000.0,
		}
		
		quotes := strategy.CalculateQuotes(marketData, position)
		require.NotNil(t, quotes)
		
		// Check bid quote
		assert.NotNil(t, quotes.Bid)
		assert.Equal(t, "bid", quotes.Bid.Side)
		assert.True(t, quotes.Bid.Price < marketData.MidPrice)
		assert.True(t, quotes.Bid.Quantity > 0)
		
		// Check ask quote
		assert.NotNil(t, quotes.Ask)
		assert.Equal(t, "ask", quotes.Ask.Side)
		assert.True(t, quotes.Ask.Price > marketData.MidPrice)
		assert.True(t, quotes.Ask.Quantity > 0)
	})

	t.Run("UpdateStrategy", func(t *testing.T) {
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		
		marketData := MarketData{
			Symbol: "BTC/USDT",
			Bid:    49999.0,
			Ask:    50001.0,
			MidPrice: 50000.0,
			Volume: 1000000.0,
		}
		
		trades := []Trade{
			{
				Symbol:   "BTC/USDT",
				Quantity: 0.1,
				Price:    50000.0,
			},
		}
		
		strategy.UpdateStrategy(marketData, trades)
		
		// Strategy should be updated (no specific assertions needed for this basic implementation)
		assert.NotNil(t, strategy)
	})

	t.Run("GetParameters", func(t *testing.T) {
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		
		params := strategy.GetParameters()
		assert.NotNil(t, params)
		assert.Equal(t, 1.5, params["spread_multiplier"])
		assert.Equal(t, 1.0, params["position_limit"])
		assert.Equal(t, 0.1, params["quote_size"])
	})
}

// TestAdaptiveMarketMaking tests the AdaptiveMarketMaking strategy
func TestAdaptiveMarketMaking(t *testing.T) {
	t.Run("NewAdaptiveMarketMaking", func(t *testing.T) {
		strategy := NewAdaptiveMarketMaking(0.02, 1.0, 0.1, 1.0, 0.1, 0.01)
		require.NotNil(t, strategy)
		assert.Equal(t, 0.02, strategy.BaseSpread)
		assert.Equal(t, 1.0, strategy.VolatilityMultiplier)
		assert.Equal(t, 0.1, strategy.VolumeMultiplier)
		assert.Equal(t, 1.0, strategy.PositionLimit)
		assert.Equal(t, 0.1, strategy.QuoteSize)
		assert.Equal(t, 0.01, strategy.LearningRate)
	})

	t.Run("CalculateQuotes", func(t *testing.T) {
		strategy := NewAdaptiveMarketMaking(0.02, 1.0, 0.1, 1.0, 0.1, 0.01)
		
		marketData := MarketData{
			Symbol:     "BTC/USDT",
			Bid:        49999.0,
			Ask:        50001.0,
			MidPrice:   50000.0,
			Spread:     2.0,
			Volume:     1000000.0,
			Volatility: 0.02,
		}
		
		position := Position{
			Symbol:      "BTC/USDT",
			Quantity:    0.1,
			AveragePrice: 50000.0,
		}
		
		quotes := strategy.CalculateQuotes(marketData, position)
		require.NotNil(t, quotes)
		
		// Check that quotes are generated
		assert.NotNil(t, quotes.Bid)
		assert.NotNil(t, quotes.Ask)
		assert.Equal(t, "bid", quotes.Bid.Side)
		assert.Equal(t, "ask", quotes.Ask.Side)
	})

	t.Run("UpdateStrategy", func(t *testing.T) {
		strategy := NewAdaptiveMarketMaking(0.02, 1.0, 0.1, 1.0, 0.1, 0.01)
		
		marketData := MarketData{
			Symbol:     "BTC/USDT",
			Bid:        49999.0,
			Ask:        50001.0,
			MidPrice:   50000.0,
			Volume:     1000000.0,
			Volatility: 0.02,
		}
		
		trades := []Trade{
			{
				Symbol:   "BTC/USDT",
				Quantity: 0.1,
				Price:    50000.0,
			},
		}
		
		strategy.UpdateStrategy(marketData, trades)
		
		// Strategy should be updated
		assert.NotNil(t, strategy)
	})

	t.Run("GetParameters", func(t *testing.T) {
		strategy := NewAdaptiveMarketMaking(0.02, 1.0, 0.1, 1.0, 0.1, 0.01)
		
		params := strategy.GetParameters()
		assert.NotNil(t, params)
		assert.Equal(t, 0.02, params["base_spread"])
		assert.Equal(t, 1.0, params["volatility_multiplier"])
		assert.Equal(t, 0.1, params["volume_multiplier"])
		assert.Equal(t, 1.0, params["position_limit"])
		assert.Equal(t, 0.1, params["quote_size"])
		assert.Equal(t, 0.01, params["learning_rate"])
	})
}

// TestIntegration tests the complete market making workflow
func TestIntegration(t *testing.T) {
	t.Run("CompleteMarketMakingWorkflow", func(t *testing.T) {
		config := MarketMakerConfig{
			MaxPositionSize:  1.0,
			MaxSpread:        0.02,
			MinSpread:        0.001,
			QuoteSize:        0.1,
			QuoteRefreshRate: 100 * time.Millisecond,
		}
		
		strategy := NewBasicMarketMaking(1.5, 1.0, 0.1)
		mm := NewMarketMaker("mm_1", strategy, config)
		
		// Start market maker
		err := mm.Start()
		require.NoError(t, err)
		
		// Wait for quotes to be generated
		time.Sleep(150 * time.Millisecond)
		
		// Check quotes
		quotes, exists := mm.Quotes["BTC/USDT"]
		assert.True(t, exists)
		assert.NotNil(t, quotes)
		assert.NotNil(t, quotes.Bid)
		assert.NotNil(t, quotes.Ask)
		
		// Simulate a trade (buying)
		trade := Trade{
			Symbol:   "BTC/USDT",
			Side:     "buy",
			Quantity: 0.1,
			Price:    50000.0,
		}
		
		mm.UpdatePosition(trade)
		
		// Check position update
		assert.Equal(t, 0.1, mm.Position.Quantity)
		assert.Equal(t, 50000.0, mm.Position.AveragePrice)
		
		// Stop market maker
		err = mm.Stop()
		require.NoError(t, err)
	})
}
