// Package oracle provides oracle price feed interfaces and implementations
// for decentralized price oracles (Pyth Network, Switchboard, Chainlink style).
//
// It provides:
//   - PriceFeed interface for reading price data
//   - Price struct with price, confidence interval, and expo (scaling)
//   - Pyth-style price account model
//   - Aggregated price from multiple oracles
//   - Staleness checks for freshness guarantees
//
// See pkg/oracle/README.md for usage examples.

package oracle

import (
	"fmt"
	"math"
	"time"
)

// ---------------------------------------------------------------------------
// Price
// ---------------------------------------------------------------------------

// Price represents a signed price with a confidence interval and scaling exponent.
// The actual price value is Price * 10^Expo.
// Confidence interval represents the standard deviation of the price.
type Price struct {
	// Price is the signed price integer.
	Price int64
	// Expo is the exponent scaling factor. Actual value = Price * 10^Expo.
	Expo int32
	// Conf is the confidence interval (standard deviation) in the same units as Price.
	Conf uint64
	// PublishTime is when the price was published (unix seconds).
	PublishTime int64
}

// Value returns the actual price value as Price * 10^Expo.
func (p *Price) Value() float64 {
	if p.Expo >= 0 {
		return float64(p.Price) * math.Pow10(int(p.Expo))
	}
	return float64(p.Price) / math.Pow10(int(-p.Expo))
}

// ---------------------------------------------------------------------------
// PriceFeed
// ---------------------------------------------------------------------------

// PriceFeed is a price oracle feed for a single asset pair.
type PriceFeed interface {
	// GetPrice returns the latest price for the asset pair.
	GetPrice() (*Price, error)
	// AssetPair returns the asset pair identifier (e.g. "BTC/USD").
	AssetPair() string
	// IsStale returns true if the price is older than maxAge.
	IsStale(maxAge time.Duration) bool
}

// PriceInfo contains price and metadata for an asset pair.
type PriceInfo struct {
	// Pair is the asset pair (e.g. "BTC/USD").
	Pair string
	// Price is the latest price.
	Price *Price
	// Feed is the oracle feed name (e.g. "pyth", "switchboard").
	Feed string
}

// Oracle aggregates prices from multiple price feeds.
type Oracle struct {
	feeds map[string]PriceFeed
}

// NewOracle creates an empty oracle aggregator.
func NewOracle() *Oracle {
	return &Oracle{feeds: make(map[string]PriceFeed)}
}

// RegisterFeed registers a price feed under a name.
func (o *Oracle) RegisterFeed(name string, feed PriceFeed) {
	o.feeds[name] = feed
}

// GetPrice returns the price from a specific feed.
func (o *Oracle) GetPrice(feedName string) (*Price, error) {
	feed, ok := o.feeds[feedName]
	if !ok {
		return nil, fmt.Errorf("feed not found: %s", feedName)
	}
	return feed.GetPrice()
}

// GetPriceInfo returns price info from a specific feed.
func (o *Oracle) GetPriceInfo(feedName string) (*PriceInfo, error) {
	price, err := o.GetPrice(feedName)
	if err != nil {
		return nil, err
	}
	return &PriceInfo{
		Pair:  o.feeds[feedName].AssetPair(),
		Price: price,
		Feed:  feedName,
	}, nil
}

// GetAggregatedPrice returns a median-aggregated price across all feeds
// for the same asset pair, filtered by freshness.
func (o *Oracle) GetAggregatedPrice(pair string, maxAge time.Duration) (*Price, error) {
	var prices []*Price
	for _, feed := range o.feeds {
		if feed.AssetPair() != pair {
			continue
		}
		if feed.IsStale(maxAge) {
			continue
		}
		p, err := feed.GetPrice()
		if err != nil {
			continue
		}
		prices = append(prices, p)
	}
	if len(prices) == 0 {
		return nil, fmt.Errorf("no fresh price feeds for pair %s", pair)
	}
	return medianPrice(prices), nil
}

// medianPrice returns the median price from a slice of prices.
func medianPrice(prices []*Price) *Price {
	if len(prices) == 0 {
		return nil
	}
	// Sort by price value
	for i := 0; i < len(prices); i++ {
		for j := i + 1; j < len(prices); j++ {
			if prices[i].Price > prices[j].Price {
				prices[i], prices[j] = prices[j], prices[i]
			}
		}
	}
	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		// Average of two middle values
		midPrice := (prices[mid-1].Price + prices[mid].Price) / 2
		midConf := (prices[mid-1].Conf + prices[mid].Conf) / 2
		return &Price{
			Price:       midPrice,
			Expo:        prices[mid].Expo,
			Conf:        midConf,
			PublishTime: prices[mid].PublishTime,
		}
	}
	return prices[mid]
}

// ---------------------------------------------------------------------------
// Pyth-style price account
// ---------------------------------------------------------------------------

// PythPriceAccount represents a Pyth Network price account structure.
// Pyth stores price data in on-chain accounts with the following layout.
type PythPriceAccount struct {
	// PriceType is the type of price (1 = price, 2 = twap, 3 = twacc).
	PriceType uint32
	// Expo is the exponent scaling factor.
	Expo int32
	// Price is the current price.
	Price int64
	// Conf is the confidence interval.
	Conf uint64
	// PublishTime is when the price was published.
	PublishTime int64
}

// ToPrice converts a PythPriceAccount to a generic Price.
func (pa *PythPriceAccount) ToPrice() *Price {
	return &Price{
		Price:       pa.Price,
		Expo:        pa.Expo,
		Conf:        pa.Conf,
		PublishTime: pa.PublishTime,
	}
}

// ---------------------------------------------------------------------------
// IsStale
// ---------------------------------------------------------------------------

// IsStale returns true if the price is older than maxAge.
func (p *Price) IsStale(maxAge time.Duration) bool {
	return time.Since(time.Unix(p.PublishTime, 0)) > maxAge
}

// AssetPair returns a placeholder pair for Pyth price accounts.
func (pa *PythPriceAccount) AssetPair() string {
	return "PYTH:" + fmt.Sprintf("%d", pa.PriceType)
}
