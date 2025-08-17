package orderbook

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMatchingEngine(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)
	assert.NotNil(t, me)
	assert.Equal(t, ob, me.orderBook)
	assert.Empty(t, me.trades)
	assert.Equal(t, int64(0), me.lastTradeID)
}

func TestMatchingEngineProcessOrder(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("process nil order", func(t *testing.T) {
		execution, err := me.ProcessOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order is nil")
		assert.Nil(t, execution)
	})

	t.Run("process invalid order", func(t *testing.T) {
		invalidOrder := &Order{
			ID:            "",
			UserID:        "user1",
			TradingPair:   "BTC/USDT",
			Side:          OrderSideBuy,
			Type:          OrderTypeLimit,
			Status:        OrderStatusPending,
			Quantity:      big.NewInt(100),
			Price:         big.NewInt(1000),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		execution, err := me.ProcessOrder(invalidOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid order ID")
		assert.Nil(t, execution)
	})

	t.Run("process valid order with no matches", func(t *testing.T) {
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(order)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Equal(t, order.ID, execution.BuyOrder.ID)
		assert.Nil(t, execution.SellOrder)
		assert.Empty(t, execution.PartialFills)
		assert.Equal(t, big.NewInt(100), execution.RemainingBuy)
		assert.Nil(t, execution.RemainingSell)
	})
}

func TestMatchingEngineMatchLimitOrders(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("match limit buy order with sell order", func(t *testing.T) {
		// Add a sell order first
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		// Process a buy order that should match
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 1)
		assert.Equal(t, big.NewInt(50), execution.RemainingBuy) // 100 - 50 = 50 remaining

		// Verify the trade was created
		trades := me.GetTrades()
		assert.Len(t, trades, 1)
		assert.Equal(t, "buy1", trades[0].BuyOrderID)
		assert.Equal(t, "sell1", trades[0].SellOrderID)
		assert.Equal(t, big.NewInt(50), trades[0].Quantity)
		assert.Equal(t, big.NewInt(1100), trades[0].Price) // Should match at sell price
	})

	t.Run("match limit sell order with buy order", func(t *testing.T) {
		// Clear previous orders
		ob.Clear()
		me.ClearTrades()

		// Add a buy order first
		buyOrder, err := NewOrder(
			"buy2", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(75), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder)

		// Process a sell order that should match
		sellOrder, err := NewOrder(
			"sell2", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(sellOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 1)
		assert.Equal(t, big.NewInt(25), execution.RemainingSell) // 100 - 75 = 25 remaining

		// Verify the trade was created
		trades := me.GetTrades()
		assert.Len(t, trades, 1)
		assert.Equal(t, "buy2", trades[0].BuyOrderID)
		assert.Equal(t, "sell2", trades[0].SellOrderID)
		assert.Equal(t, big.NewInt(75), trades[0].Quantity)
		assert.Equal(t, big.NewInt(1100), trades[0].Price) // Should match at sell price
	})
}

func TestMatchingEngineMatchMarketOrders(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("match market buy order with sell orders", func(t *testing.T) {
		// Add sell orders at different prices
		sellOrder1, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(30), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder1)

		sellOrder2, err := NewOrder(
			"sell2", "user3", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(40), big.NewInt(1150),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder2)

		// Process a market buy order
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeMarket,
			big.NewInt(50), nil,
			TimeInForceIOC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 2) // Should match with both sell orders

		// Verify trades were created
		trades := me.GetTrades()
		assert.Len(t, trades, 2)

		// First trade should be at 1100 (best price)
		assert.Equal(t, big.NewInt(30), trades[0].Quantity)
		assert.Equal(t, big.NewInt(1100), trades[0].Price)

		// Second trade should be at 1150 (next best price)
		assert.Equal(t, big.NewInt(20), trades[1].Quantity) // 50 - 30 = 20
		assert.Equal(t, big.NewInt(1150), trades[1].Price)
	})

	t.Run("match market sell order with buy orders", func(t *testing.T) {
		// Clear previous orders
		ob.Clear()
		me.ClearTrades()

		// Add buy orders at different prices
		buyOrder1, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(25), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder1)

		buyOrder2, err := NewOrder(
			"buy2", "user4", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(35), big.NewInt(1250),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder2)

		// Process a market sell order
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeMarket,
			big.NewInt(50), nil,
			TimeInForceIOC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(sellOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 2) // Should match with both buy orders

		// Verify trades were created
		trades := me.GetTrades()
		assert.Len(t, trades, 2)

		// First trade should be at 1250 (best price)
		assert.Equal(t, big.NewInt(35), trades[0].Quantity)
		assert.Equal(t, big.NewInt(1250), trades[0].Price)

		// Second trade should be at 1200 (next best price)
		assert.Equal(t, big.NewInt(15), trades[1].Quantity) // 50 - 35 = 15
		assert.Equal(t, big.NewInt(1200), trades[1].Price)
	})
}

func TestMatchingEnginePartialFills(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("partial fill of buy order", func(t *testing.T) {
		// Add a sell order with smaller quantity
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(30), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		// Process a buy order with larger quantity
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 1)
		assert.Equal(t, big.NewInt(70), execution.RemainingBuy) // 100 - 30 = 70 remaining

		// Verify the buy order is still in the book with partial fill
		remainingOrder, err := ob.GetOrder("buy1")
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusPartial, remainingOrder.Status)
		assert.Equal(t, big.NewInt(30), remainingOrder.FilledQuantity)
		assert.Equal(t, big.NewInt(70), remainingOrder.RemainingQuantity)
	})

	t.Run("partial fill of sell order", func(t *testing.T) {
		// Clear previous orders
		ob.Clear()
		me.ClearTrades()

		// Add a buy order with smaller quantity
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(25), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder)

		// Process a sell order with larger quantity
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(sellOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Len(t, execution.PartialFills, 1)
		assert.Equal(t, big.NewInt(75), execution.RemainingSell) // 100 - 25 = 75 remaining

		// Verify the sell order is still in the book with partial fill
		remainingOrder, err := ob.GetOrder("sell1")
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusPartial, remainingOrder.Status)
		assert.Equal(t, big.NewInt(25), remainingOrder.FilledQuantity)
		assert.Equal(t, big.NewInt(75), remainingOrder.RemainingQuantity)
	})
}

func TestMatchingEngineNoMatches(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("buy order below best ask", func(t *testing.T) {
		// Add a sell order
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		// Process a buy order below the ask price
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Empty(t, execution.PartialFills)
		assert.Equal(t, big.NewInt(100), execution.RemainingBuy) // No matches, full quantity remaining

		// Verify no trades were created
		trades := me.GetTrades()
		assert.Empty(t, trades)
	})

	t.Run("sell order above best bid", func(t *testing.T) {
		// Clear previous orders
		ob.Clear()
		me.ClearTrades()

		// Add a buy order
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder)

		// Process a sell order above the bid price
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(sellOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		assert.Empty(t, execution.PartialFills)
		assert.Equal(t, big.NewInt(100), execution.RemainingSell) // No matches, full quantity remaining

		// Verify no trades were created
		trades := me.GetTrades()
		assert.Empty(t, trades)
	})
}

func TestMatchingEngineTradeManagement(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	t.Run("get all trades", func(t *testing.T) {
		// Add some orders and create trades
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)

		// Get all trades
		trades := me.GetTrades()
		assert.Len(t, trades, 1)
		assert.Equal(t, "buy1", trades[0].BuyOrderID)
		assert.Equal(t, "sell1", trades[0].SellOrderID)
	})

	t.Run("get trades by trading pair", func(t *testing.T) {
		// Clear and add trades for different pairs
		ob.Clear()
		me.ClearTrades()

		// Add trades for BTC/USDT
		sellOrder1, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder1)

		buyOrder1, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution1, err := me.ProcessOrder(buyOrder1)
		assert.NoError(t, err)
		assert.NotNil(t, execution1)

		// Get trades by trading pair
		btcTrades := me.GetTradesByTradingPair("BTC/USDT")
		assert.Len(t, btcTrades, 1)

		ethTrades := me.GetTradesByTradingPair("ETH/USDT")
		assert.Empty(t, ethTrades)
	})

	t.Run("get trades by user", func(t *testing.T) {
		// Clear and add trades for different users
		ob.Clear()
		me.ClearTrades()

		// Add trades
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)

		// Get trades by user
		user1Trades := me.GetTradesByUser("user1")
		assert.Len(t, user1Trades, 1)

		user2Trades := me.GetTradesByUser("user2")
		assert.Len(t, user2Trades, 1)

		user3Trades := me.GetTradesByUser("user3")
		assert.Empty(t, user3Trades)
	})

	t.Run("get trade count", func(t *testing.T) {
		// Clear and add some trades
		ob.Clear()
		me.ClearTrades()

		assert.Equal(t, 0, me.GetTradeCount())

		// Add a trade
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		execution, err := me.ProcessOrder(buyOrder)
		assert.NoError(t, err)
		assert.NotNil(t, execution)

		assert.Equal(t, 1, me.GetTradeCount())
	})

	t.Run("clear trades", func(t *testing.T) {
		// Clear trades
		me.ClearTrades()

		assert.Equal(t, 0, me.GetTradeCount())
		assert.Empty(t, me.GetTrades())
	})
}

func TestMatchingEngineConcurrency(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	me := NewMatchingEngine(ob)

	// Test concurrent order processing
	const numGoroutines = 5
	const numOrders = 20

	var wg sync.WaitGroup

	// Start goroutines to add buy orders concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOrders; j++ {
				order, err := NewOrder(
					"buy"+string(rune(id))+string(rune(j)),
					"user"+string(rune(id)),
					"BTC/USDT",
					OrderSideBuy, OrderTypeLimit,
					big.NewInt(10), big.NewInt(1100+int64(j)),
					TimeInForceGTC, nil, nil,
				)
				if err == nil {
					me.ProcessOrder(order)
				}
			}
		}(i)
	}

	// Start goroutines to add sell orders concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOrders; j++ {
				order, err := NewOrder(
					"sell"+string(rune(id))+string(rune(j)),
					"user"+string(rune(id+100)),
					"BTC/USDT",
					OrderSideSell, OrderTypeLimit,
					big.NewInt(10), big.NewInt(1000+int64(j)),
					TimeInForceGTC, nil, nil,
				)
				if err == nil {
					me.ProcessOrder(order)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify no data races occurred and some trades were created
	assert.Greater(t, me.GetTradeCount(), 0)
}

// Benchmark tests for performance validation
func BenchmarkMatchingEngineProcessOrder(b *testing.B) {
	ob, err := NewOrderBook("BTC/USDT")
	if err != nil {
		b.Fatal(err)
	}

	me := NewMatchingEngine(ob)

	// Add some existing orders to match against
	for i := 0; i < 100; i++ {
		order, err := NewOrder(
			"sell"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideSell, OrderTypeLimit,
			big.NewInt(10), big.NewInt(1000+int64(i)),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		ob.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order, err := NewOrder(
			"buy"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideBuy, OrderTypeLimit,
			big.NewInt(10), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		me.ProcessOrder(order)
	}
}

func BenchmarkMatchingEngineGetTrades(b *testing.B) {
	ob, err := NewOrderBook("BTC/USDT")
	if err != nil {
		b.Fatal(err)
	}

	me := NewMatchingEngine(ob)

	// Add some trades first
	for i := 0; i < 1000; i++ {
		order, err := NewOrder(
			"order"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideBuy, OrderTypeLimit,
			big.NewInt(10), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		me.ProcessOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		me.GetTrades()
	}
}
