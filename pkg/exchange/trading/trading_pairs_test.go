package trading

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTradingPair(t *testing.T) {
	t.Run("valid trading pair", func(t *testing.T) {
		minQuantity := big.NewInt(1)
		maxQuantity := big.NewInt(1000000)
		minPrice := big.NewInt(1)
		maxPrice := big.NewInt(1000000)
		tickSize := big.NewInt(1)
		stepSize := big.NewInt(1)
		makerFee := big.NewInt(10) // 0.1%
		takerFee := big.NewInt(20) // 0.2%

		pair, err := NewTradingPair(
			"BTC", "USDT",
			minQuantity, maxQuantity, minPrice, maxPrice,
			tickSize, stepSize, makerFee, takerFee,
		)

		require.NoError(t, err)
		assert.NotNil(t, pair)
		assert.Equal(t, "BTC_USDT", pair.ID)
		assert.Equal(t, "BTC", pair.BaseAsset)
		assert.Equal(t, "USDT", pair.QuoteAsset)
		assert.Equal(t, "BTC/USDT", pair.Symbol)
		assert.Equal(t, PairStatusActive, pair.Status)
		assert.Equal(t, minQuantity, pair.MinQuantity)
		assert.Equal(t, maxQuantity, pair.MaxQuantity)
		assert.Equal(t, minPrice, pair.MinPrice)
		assert.Equal(t, maxPrice, pair.MaxPrice)
		assert.Equal(t, tickSize, pair.TickSize)
		assert.Equal(t, stepSize, pair.StepSize)
		assert.Equal(t, makerFee, pair.MakerFee)
		assert.Equal(t, takerFee, pair.TakerFee)
		assert.False(t, pair.CreatedAt.IsZero())
		assert.False(t, pair.UpdatedAt.IsZero())
		assert.Equal(t, big.NewInt(0), pair.Volume24h)
		assert.Equal(t, big.NewInt(0), pair.PriceChange24h)
		assert.Equal(t, big.NewInt(0), pair.PriceChangePercent24h)
	})

	t.Run("empty base asset", func(t *testing.T) {
		pair, err := NewTradingPair(
			"", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidBaseAsset.Error())
		assert.Nil(t, pair)
	})

	t.Run("empty quote asset", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidQuoteAsset.Error())
		assert.Nil(t, pair)
	})

	t.Run("same base and quote assets", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "BTC",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base and quote assets cannot be the same")
		assert.Nil(t, pair)
	})

	t.Run("invalid tick size", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(0), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidTickSize.Error())
		assert.Nil(t, pair)
	})

	t.Run("invalid step size", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(0),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidStepSize.Error())
		assert.Nil(t, pair)
	})

	t.Run("invalid maker fee", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(-10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidFees.Error())
		assert.Nil(t, pair)
	})

	t.Run("invalid taker fee", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(-20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidFees.Error())
		assert.Nil(t, pair)
	})

	t.Run("min quantity greater than max quantity", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1000000), big.NewInt(1),
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min quantity cannot be greater than max quantity")
		assert.Nil(t, pair)
	})

	t.Run("min price greater than max price", func(t *testing.T) {
		pair, err := NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1), big.NewInt(1000000),
			big.NewInt(1000000), big.NewInt(1),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(10), big.NewInt(20),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min price cannot be greater than max price")
		assert.Nil(t, pair)
	})
}

func TestTradingPairValidation(t *testing.T) {
	t.Run("valid trading pair", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty base asset", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidBaseAsset, err)
	})

	t.Run("empty quote asset", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuoteAsset, err)
	})

	t.Run("same base and quote assets", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "BTC",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base and quote assets cannot be the same")
	})

	t.Run("invalid tick size", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(0),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTickSize, err)
	})

	t.Run("invalid step size", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(0),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidStepSize, err)
	})

	t.Run("invalid maker fee", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(-10),
			TakerFee:   big.NewInt(20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidFees, err)
	})

	t.Run("invalid taker fee", func(t *testing.T) {
		pair := &TradingPair{
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			TickSize:   big.NewInt(1),
			StepSize:   big.NewInt(1),
			MakerFee:   big.NewInt(10),
			TakerFee:   big.NewInt(-20),
		}

		err := pair.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidFees, err)
	})
}

func TestTradingPairStatus(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("active status", func(t *testing.T) {
		pair.Status = PairStatusActive
		assert.True(t, pair.IsActive())
		assert.True(t, pair.CanTrade())
	})

	t.Run("inactive status", func(t *testing.T) {
		pair.Status = PairStatusInactive
		assert.False(t, pair.IsActive())
		assert.False(t, pair.CanTrade())
	})

	t.Run("suspended status", func(t *testing.T) {
		pair.Status = PairStatusSuspended
		assert.False(t, pair.IsActive())
		assert.False(t, pair.CanTrade())
	})

	t.Run("maintenance status", func(t *testing.T) {
		pair.Status = PairStatusMaintenance
		assert.False(t, pair.IsActive())
		assert.False(t, pair.CanTrade())
	})
}

func TestTradingPairUpdateStatus(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("update to active", func(t *testing.T) {
		err := pair.UpdateStatus(PairStatusActive)
		assert.NoError(t, err)
		assert.Equal(t, PairStatusActive, pair.Status)
		assert.True(t, pair.UpdatedAt.After(pair.CreatedAt))
	})

	t.Run("update to inactive", func(t *testing.T) {
		err := pair.UpdateStatus(PairStatusInactive)
		assert.NoError(t, err)
		assert.Equal(t, PairStatusInactive, pair.Status)
	})

	t.Run("update to suspended", func(t *testing.T) {
		err := pair.UpdateStatus(PairStatusSuspended)
		assert.NoError(t, err)
		assert.Equal(t, PairStatusSuspended, pair.Status)
	})

	t.Run("update to maintenance", func(t *testing.T) {
		err := pair.UpdateStatus(PairStatusMaintenance)
		assert.NoError(t, err)
		assert.Equal(t, PairStatusMaintenance, pair.Status)
	})

	t.Run("invalid status", func(t *testing.T) {
		err := pair.UpdateStatus("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestTradingPairUpdateFees(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("valid fee update", func(t *testing.T) {
		newMakerFee := big.NewInt(15)
		newTakerFee := big.NewInt(25)

		err := pair.UpdateFees(newMakerFee, newTakerFee)
		assert.NoError(t, err)
		assert.Equal(t, newMakerFee, pair.MakerFee)
		assert.Equal(t, newTakerFee, pair.TakerFee)
		assert.True(t, pair.UpdatedAt.After(pair.CreatedAt))
	})

	t.Run("invalid maker fee", func(t *testing.T) {
		err := pair.UpdateFees(big.NewInt(-5), big.NewInt(25))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidFees.Error())
	})

	t.Run("invalid taker fee", func(t *testing.T) {
		err := pair.UpdateFees(big.NewInt(15), big.NewInt(-5))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrInvalidFees.Error())
	})
}

func TestTradingPairUpdateLastTrade(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("update with valid price", func(t *testing.T) {
		price := big.NewInt(50000)
		beforeUpdate := pair.UpdatedAt

		pair.UpdateLastTrade(price)

		assert.Equal(t, price, pair.LastTradePrice)
		assert.NotNil(t, pair.LastTradeTime)
		assert.True(t, pair.UpdatedAt.After(beforeUpdate))
	})

	t.Run("update with nil price", func(t *testing.T) {
		beforeUpdate := pair.UpdatedAt
		originalPrice := pair.LastTradePrice

		pair.UpdateLastTrade(nil)

		assert.Equal(t, originalPrice, pair.LastTradePrice)
		assert.Equal(t, beforeUpdate, pair.UpdatedAt)
	})

	t.Run("update with zero price", func(t *testing.T) {
		beforeUpdate := pair.UpdatedAt
		originalPrice := pair.LastTradePrice

		pair.UpdateLastTrade(big.NewInt(0))

		assert.Equal(t, originalPrice, pair.LastTradePrice)
		assert.Equal(t, beforeUpdate, pair.UpdatedAt)
	})
}

func TestTradingPairUpdateVolume24h(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("update with valid volume", func(t *testing.T) {
		volume := big.NewInt(1000000)
		beforeUpdate := pair.UpdatedAt

		pair.UpdateVolume24h(volume)

		assert.Equal(t, volume, pair.Volume24h)
		assert.True(t, pair.UpdatedAt.After(beforeUpdate))
	})

	t.Run("update with zero volume", func(t *testing.T) {
		volume := big.NewInt(0)
		beforeUpdate := pair.UpdatedAt

		pair.UpdateVolume24h(volume)

		assert.Equal(t, volume, pair.Volume24h)
		assert.True(t, pair.UpdatedAt.After(beforeUpdate))
	})

	t.Run("update with nil volume", func(t *testing.T) {
		beforeUpdate := pair.UpdatedAt
		originalVolume := pair.Volume24h

		pair.UpdateVolume24h(nil)

		assert.Equal(t, originalVolume, pair.Volume24h)
		assert.Equal(t, beforeUpdate, pair.UpdatedAt)
	})
}

func TestTradingPairUpdatePriceChange24h(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	t.Run("update with valid values", func(t *testing.T) {
		priceChange := big.NewInt(1000)
		priceChangePercent := big.NewInt(2) // 2%
		beforeUpdate := pair.UpdatedAt

		pair.UpdatePriceChange24h(priceChange, priceChangePercent)

		assert.Equal(t, priceChange, pair.PriceChange24h)
		assert.Equal(t, priceChangePercent, pair.PriceChangePercent24h)
		assert.True(t, pair.UpdatedAt.After(beforeUpdate))
	})

	t.Run("update with nil values", func(t *testing.T) {
		beforeUpdate := pair.UpdatedAt
		originalPriceChange := pair.PriceChange24h
		originalPriceChangePercent := pair.PriceChangePercent24h

		pair.UpdatePriceChange24h(nil, nil)

		assert.Equal(t, originalPriceChange, pair.PriceChange24h)
		assert.Equal(t, originalPriceChangePercent, pair.PriceChangePercent24h)
		assert.Equal(t, beforeUpdate, pair.UpdatedAt)
	})
}

func TestTradingPairValidatePrice(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(100), // $0.01 tick size
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
		MinPrice:   big.NewInt(1000),    // $10.00 minimum
		MaxPrice:   big.NewInt(1000000), // $10,000.00 maximum
	}

	t.Run("valid price", func(t *testing.T) {
		price := big.NewInt(50000) // $500.00
		err := pair.ValidatePrice(price)
		assert.NoError(t, err)
	})

	t.Run("nil price", func(t *testing.T) {
		err := pair.ValidatePrice(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price must be positive")
	})

	t.Run("zero price", func(t *testing.T) {
		err := pair.ValidatePrice(big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price must be positive")
	})

	t.Run("negative price", func(t *testing.T) {
		err := pair.ValidatePrice(big.NewInt(-100))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price must be positive")
	})

	t.Run("price below minimum", func(t *testing.T) {
		price := big.NewInt(500) // $5.00
		err := pair.ValidatePrice(price)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price 500 is below minimum 1000")
	})

	t.Run("price above maximum", func(t *testing.T) {
		price := big.NewInt(2000000) // $20,000.00
		err := pair.ValidatePrice(price)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price 2000000 is above maximum 1000000")
	})

	t.Run("price does not match tick size", func(t *testing.T) {
		price := big.NewInt(50005) // $500.05 (not divisible by 100)
		err := pair.ValidatePrice(price)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price 50005 does not match tick size 100")
	})

	t.Run("price matches tick size", func(t *testing.T) {
		price := big.NewInt(50000) // $500.00 (divisible by 100)
		err := pair.ValidatePrice(price)
		assert.NoError(t, err)
	})
}

func TestTradingPairValidateQuantity(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:   "BTC",
		QuoteAsset:  "USDT",
		TickSize:    big.NewInt(1),
		StepSize:    big.NewInt(1000), // 0.001 BTC step size
		MakerFee:    big.NewInt(10),
		TakerFee:    big.NewInt(20),
		MinQuantity: big.NewInt(1000),       // 0.001 BTC minimum
		MaxQuantity: big.NewInt(1000000000), // 1000 BTC maximum
	}

	t.Run("valid quantity", func(t *testing.T) {
		quantity := big.NewInt(100000) // 0.1 BTC
		err := pair.ValidateQuantity(quantity)
		assert.NoError(t, err)
	})

	t.Run("nil quantity", func(t *testing.T) {
		err := pair.ValidateQuantity(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("zero quantity", func(t *testing.T) {
		err := pair.ValidateQuantity(big.NewInt(0))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("negative quantity", func(t *testing.T) {
		err := pair.ValidateQuantity(big.NewInt(-1000))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("quantity below minimum", func(t *testing.T) {
		quantity := big.NewInt(500) // 0.0005 BTC
		err := pair.ValidateQuantity(quantity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity 500 is below minimum 1000")
	})

	t.Run("quantity above maximum", func(t *testing.T) {
		quantity := big.NewInt(2000000000) // 2000 BTC
		err := pair.ValidateQuantity(quantity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity 2000000000 is above maximum 1000000000")
	})

	t.Run("quantity does not match step size", func(t *testing.T) {
		quantity := big.NewInt(100005) // 0.100005 BTC (not divisible by 1000)
		err := pair.ValidateQuantity(quantity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity 100005 does not match step size 1000")
	})

	t.Run("quantity matches step size", func(t *testing.T) {
		quantity := big.NewInt(100000) // 0.1 BTC (divisible by 1000)
		err := pair.ValidateQuantity(quantity)
		assert.NoError(t, err)
	})
}

func TestTradingPairCalculateFee(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10), // 0.1%
		TakerFee:   big.NewInt(20), // 0.2%
	}

	t.Run("calculate maker fee", func(t *testing.T) {
		quantity := big.NewInt(100000) // 0.1 BTC
		price := big.NewInt(50000)     // $50,000.00

		fee, err := pair.CalculateFee(quantity, price, true)
		assert.NoError(t, err)

		// Expected fee: (100000 * 50000 * 10) / 10000 = 5000000 (0.1 BTC)
		expectedFee := big.NewInt(5000000)
		assert.Equal(t, expectedFee, fee)
	})

	t.Run("calculate taker fee", func(t *testing.T) {
		quantity := big.NewInt(100000) // 0.1 BTC
		price := big.NewInt(50000)     // $50,000.00

		fee, err := pair.CalculateFee(quantity, price, false)
		assert.NoError(t, err)

		// Expected fee: (100000 * 50000 * 20) / 10000 = 10000000 (0.2 BTC)
		expectedFee := big.NewInt(10000000)
		assert.Equal(t, expectedFee, fee)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		quantity := big.NewInt(-1000)
		price := big.NewInt(50000)

		fee, err := pair.CalculateFee(quantity, price, true)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})

	t.Run("invalid price", func(t *testing.T) {
		quantity := big.NewInt(100000)
		price := big.NewInt(-50000)

		fee, err := pair.CalculateFee(quantity, price, true)
		assert.Error(t, err)
		assert.Nil(t, fee)
	})
}

func TestTradingPairClone(t *testing.T) {
	pair := &TradingPair{
		BaseAsset:             "BTC",
		QuoteAsset:            "USDT",
		TickSize:              big.NewInt(1),
		StepSize:              big.NewInt(1),
		MakerFee:              big.NewInt(10),
		TakerFee:              big.NewInt(20),
		MinQuantity:           big.NewInt(1000),
		MaxQuantity:           big.NewInt(1000000),
		MinPrice:              big.NewInt(1000),
		MaxPrice:              big.NewInt(1000000),
		LastTradePrice:        big.NewInt(50000),
		Volume24h:             big.NewInt(1000000),
		PriceChange24h:        big.NewInt(1000),
		PriceChangePercent24h: big.NewInt(2),
	}

	lastTradeTime := time.Now()
	pair.LastTradeTime = &lastTradeTime

	clone := pair.Clone()

	// Verify clone is not the same pointer
	assert.NotSame(t, pair, clone)

	// Verify all fields are copied correctly
	assert.Equal(t, pair.ID, clone.ID)
	assert.Equal(t, pair.BaseAsset, clone.BaseAsset)
	assert.Equal(t, pair.QuoteAsset, clone.QuoteAsset)
	assert.Equal(t, pair.Symbol, clone.Symbol)
	assert.Equal(t, pair.Status, clone.Status)
	assert.Equal(t, pair.CreatedAt, clone.CreatedAt)
	assert.Equal(t, pair.UpdatedAt, clone.UpdatedAt)

	// Verify big.Int fields are deep copied
	assert.Equal(t, pair.TickSize, clone.TickSize)
	assert.Equal(t, pair.StepSize, clone.StepSize)
	assert.Equal(t, pair.MakerFee, clone.MakerFee)
	assert.Equal(t, pair.TakerFee, clone.TakerFee)
	assert.Equal(t, pair.MinQuantity, clone.MinQuantity)
	assert.Equal(t, pair.MaxQuantity, clone.MaxQuantity)
	assert.Equal(t, pair.MinPrice, clone.MinPrice)
	assert.Equal(t, pair.MaxPrice, clone.MaxPrice)
	assert.Equal(t, pair.LastTradePrice, clone.LastTradePrice)
	assert.Equal(t, pair.Volume24h, clone.Volume24h)
	assert.Equal(t, pair.PriceChange24h, clone.PriceChange24h)
	assert.Equal(t, pair.PriceChangePercent24h, clone.PriceChangePercent24h)

	// Verify big.Int fields are not the same pointers
	assert.NotSame(t, pair.TickSize, clone.TickSize)
	assert.NotSame(t, pair.StepSize, clone.StepSize)
	assert.NotSame(t, pair.MakerFee, clone.MakerFee)
	assert.NotSame(t, pair.TakerFee, clone.TakerFee)

	// Verify time fields are deep copied
	assert.Equal(t, pair.LastTradeTime, clone.LastTradeTime)
	assert.NotSame(t, pair.LastTradeTime, clone.LastTradeTime)

	// Verify modifying clone doesn't affect original
	clone.TickSize.Add(clone.TickSize, big.NewInt(1))
	assert.NotEqual(t, pair.TickSize, clone.TickSize)
}

func TestTradingPairError(t *testing.T) {
	t.Run("error formatting", func(t *testing.T) {
		err := &TradingPairError{
			Operation: "TestOperation",
			Message:   "test message",
			PairID:    "BTC_USDT",
		}

		expected := "TestOperation: test message"
		assert.Equal(t, expected, err.Error())
	})
}

// Benchmark tests for performance validation
func BenchmarkTradingPairValidate(b *testing.B) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair.Validate()
	}
}

func BenchmarkTradingPairValidatePrice(b *testing.B) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(100),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
		MinPrice:   big.NewInt(1000),
		MaxPrice:   big.NewInt(1000000),
	}

	price := big.NewInt(50000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair.ValidatePrice(price)
	}
}

func BenchmarkTradingPairValidateQuantity(b *testing.B) {
	pair := &TradingPair{
		BaseAsset:   "BTC",
		QuoteAsset:  "USDT",
		TickSize:    big.NewInt(1),
		StepSize:    big.NewInt(1000),
		MakerFee:    big.NewInt(10),
		TakerFee:    big.NewInt(20),
		MinQuantity: big.NewInt(1000),
		MaxQuantity: big.NewInt(1000000000),
	}

	quantity := big.NewInt(100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair.ValidateQuantity(quantity)
	}
}

func BenchmarkTradingPairCalculateFee(b *testing.B) {
	pair := &TradingPair{
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		TickSize:   big.NewInt(1),
		StepSize:   big.NewInt(1),
		MakerFee:   big.NewInt(10),
		TakerFee:   big.NewInt(20),
	}

	quantity := big.NewInt(100000)
	price := big.NewInt(50000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair.CalculateFee(quantity, price, true)
	}
}

func BenchmarkTradingPairClone(b *testing.B) {
	pair := &TradingPair{
		BaseAsset:             "BTC",
		QuoteAsset:            "USDT",
		TickSize:              big.NewInt(1),
		StepSize:              big.NewInt(1),
		MakerFee:              big.NewInt(10),
		TakerFee:              big.NewInt(20),
		MinQuantity:           big.NewInt(1000),
		MaxQuantity:           big.NewInt(1000000),
		MinPrice:              big.NewInt(1000),
		MaxPrice:              big.NewInt(1000000),
		LastTradePrice:        big.NewInt(50000),
		Volume24h:             big.NewInt(1000000),
		PriceChange24h:        big.NewInt(1000),
		PriceChangePercent24h: big.NewInt(2),
	}

	lastTradeTime := time.Now()
	pair.LastTradeTime = &lastTradeTime

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair.Clone()
	}
}
