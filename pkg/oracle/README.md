# `pkg/oracle` — Price Oracle Aggregation (Pyth/Switchboard style)

Price oracle interfaces and aggregation for decentralized price feeds, aligned
with Pyth Network and Switchboard patterns as of November 2026.

## Components

| Component | Role |
|-----------|------|
| `Price` | Signed price with confidence interval and scaling exponent (`Price * 10^Expo`) |
| `PriceFeed` | Interface for a single oracle feed (get price, asset pair, staleness) |
| `Oracle` | Aggregator that registers multiple feeds and returns median prices |
| `PythPriceAccount` | Pyth Network price account model (price type, expo, price, conf, publish time) |

## Example

```go
// 1. Create an oracle aggregator
oracle := oracle.NewOracle()

// 2. Register price feeds
oracle.RegisterFeed("pyth", &myPythFeed{})
oracle.RegisterFeed("switchboard", &mySwitchboardFeed{})

// 3. Get aggregated price (median across fresh feeds)
price, err := oracle.GetAggregatedPrice("BTC/USD", 5*time.Minute)
if err != nil {
    // no fresh feeds
    log.Fatal(err)
}
fmt.Printf("BTC/USD price: %f\n", price.Value())
```

## Price model

```
actual value = Price * 10^Expo
```

Confidence interval (`Conf`) represents the standard deviation in the same
units as `Price`. Staleness is checked via `PublishTime` against a max age.

## Standards alignment

| Standard / system | Detail |
|-------------------|--------|
| Pyth Network | Price account model (price type, expo, price, conf, publish time) |
| Switchboard | Aggregated oracle feeds with confidence intervals |
| Chainlink | Price feed concept (single asset pair, latest value) |

## Related

- `pkg/rwa` — RWA tokenization uses oracle prices for NAV computation
- `pkg/compliance/travelrule` — Travel Rule uses asset valuations
