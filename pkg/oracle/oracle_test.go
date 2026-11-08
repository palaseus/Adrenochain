package oracle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockFeed is a test price feed.
type mockFeed struct {
	pair   string
	price  *Price
	err    error
	maxAge time.Duration
}

func (f *mockFeed) GetPrice() (*Price, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.price, nil
}

func (f *mockFeed) AssetPair() string {
	return f.pair
}

func (f *mockFeed) IsStale(maxAge time.Duration) bool {
	return time.Since(time.Unix(f.price.PublishTime, 0)) > maxAge
}

func TestPriceValue(t *testing.T) {
	// Positive exponent: 1234 * 10^2 = 123400
	p := &Price{Price: 1234, Expo: 2}
	assert.Equal(t, 123400.0, p.Value())

	// Negative exponent: 1234 * 10^-2 = 12.34
	p = &Price{Price: 1234, Expo: -2}
	assert.Equal(t, 12.34, p.Value())

	// Zero exponent
	p = &Price{Price: 500, Expo: 0}
	assert.Equal(t, 500.0, p.Value())
}

func TestPriceIsStale(t *testing.T) {
	now := time.Now().Unix()
	p := &Price{Price: 100, PublishTime: now}

	// Not stale
	assert.False(t, p.IsStale(time.Minute))

	// Stale
	p.PublishTime = now - 300 // 5 minutes ago
	assert.True(t, p.IsStale(time.Minute))
}

func TestOracleRegisterAndGet(t *testing.T) {
	o := NewOracle()
	feed := &mockFeed{
		pair:  "BTC/USD",
		price: &Price{Price: 50000, Expo: 2, Conf: 10, PublishTime: time.Now().Unix()},
	}
	o.RegisterFeed("pyth", feed)

	price, err := o.GetPrice("pyth")
	assert.NoError(t, err)
	assert.Equal(t, int64(50000), price.Price)

	// Unknown feed
	_, err = o.GetPrice("unknown")
	assert.Error(t, err)
}

func TestOracleGetPriceInfo(t *testing.T) {
	o := NewOracle()
	feed := &mockFeed{
		pair:  "ETH/USD",
		price: &Price{Price: 3000, Expo: 2, Conf: 5, PublishTime: time.Now().Unix()},
	}
	o.RegisterFeed("pyth", feed)

	info, err := o.GetPriceInfo("pyth")
	assert.NoError(t, err)
	assert.Equal(t, "ETH/USD", info.Pair)
	assert.Equal(t, int64(3000), info.Price.Price)
	assert.Equal(t, "pyth", info.Feed)
}

func TestOracleAggregatedPrice(t *testing.T) {
	now := time.Now().Unix()
	o := NewOracle()

	// Register three feeds for BTC/USD
	o.RegisterFeed("pyth", &mockFeed{
		pair:  "BTC/USD",
		price: &Price{Price: 50000, Expo: 2, Conf: 10, PublishTime: now},
	})
	o.RegisterFeed("switchboard", &mockFeed{
		pair:  "BTC/USD",
		price: &Price{Price: 50100, Expo: 2, Conf: 10, PublishTime: now},
	})
	o.RegisterFeed("chainlink", &mockFeed{
		pair:  "BTC/USD",
		price: &Price{Price: 49900, Expo: 2, Conf: 10, PublishTime: now},
	})

	price, err := o.GetAggregatedPrice("BTC/USD", time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(50000), price.Price) // median of 49900, 50000, 50100

	// Stale feed should be excluded
	staleFeed := &mockFeed{
		pair:  "BTC/USD",
		price: &Price{Price: 999999, Expo: 2, Conf: 10, PublishTime: now - 3600}, // 1 hour ago
	}
	o.RegisterFeed("stale", staleFeed)
	price, err = o.GetAggregatedPrice("BTC/USD", time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(50000), price.Price) // stale excluded

	// No fresh feeds
	emptyOracle := NewOracle()
	_, err = emptyOracle.GetAggregatedPrice("BTC/USD", time.Minute)
	assert.Error(t, err)
}

func TestOracleAggregatedPriceEvenCount(t *testing.T) {
	now := time.Now().Unix()
	o := NewOracle()

	o.RegisterFeed("a", &mockFeed{
		pair:  "X/USD",
		price: &Price{Price: 100, Expo: 2, Conf: 10, PublishTime: now},
	})
	o.RegisterFeed("b", &mockFeed{
		pair:  "X/USD",
		price: &Price{Price: 200, Expo: 2, Conf: 10, PublishTime: now},
	})

	price, err := o.GetAggregatedPrice("X/USD", time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(150), price.Price) // median of 100, 200 = 150
}

func TestPythPriceAccount(t *testing.T) {
	pa := &PythPriceAccount{
		PriceType:   1,
		Expo:        2,
		Price:       50000,
		Conf:        10,
		PublishTime: time.Now().Unix(),
	}

	p := pa.ToPrice()
	assert.Equal(t, int64(50000), p.Price)
	assert.Equal(t, int32(2), p.Expo)
	assert.Equal(t, uint64(10), p.Conf)
	assert.False(t, p.IsStale(time.Minute)) // just published, not stale

	// Stale price
	pa.PublishTime = time.Now().Add(-time.Hour).Unix()
	p = pa.ToPrice()
	assert.True(t, p.IsStale(time.Minute)) // 1 hour ago, stale for 1 minute
	assert.Equal(t, "PYTH:1", pa.AssetPair())
}
