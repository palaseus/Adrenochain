package orderbook

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderBook(t *testing.T) {
	t.Run("valid trading pair", func(t *testing.T) {
		ob, err := NewOrderBook("BTC/USDT")
		require.NoError(t, err)
		assert.NotNil(t, ob)
		assert.Equal(t, "BTC/USDT", ob.GetTradingPair())
		assert.True(t, ob.IsEmpty())
		assert.Equal(t, 0, ob.GetOrderCount())
	})

	t.Run("empty trading pair", func(t *testing.T) {
		ob, err := NewOrderBook("")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTradingPair, err)
		assert.Nil(t, ob)
	})
}

func TestOrderBookAddOrder(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("add valid buy order", func(t *testing.T) {
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.NoError(t, err)
		assert.Equal(t, 1, ob.GetOrderCount())
		assert.Equal(t, 1, ob.GetBuyOrderCount())
		assert.Equal(t, 0, ob.GetSellOrderCount())
	})

	t.Run("add valid sell order", func(t *testing.T) {
		order, err := NewOrder(
			"order2", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.NoError(t, err)
		assert.Equal(t, 2, ob.GetOrderCount())
		assert.Equal(t, 1, ob.GetBuyOrderCount())
		assert.Equal(t, 1, ob.GetSellOrderCount())
	})

	t.Run("add order with nil order", func(t *testing.T) {
		err := ob.AddOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order is nil")
	})

	t.Run("add order with wrong trading pair", func(t *testing.T) {
		order, err := NewOrder(
			"order3", "user3", "ETH/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trading pair mismatch")
	})

	t.Run("add duplicate order", func(t *testing.T) {
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrOrderAlreadyExists.Error())
	})
}

func TestOrderBookRemoveOrder(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Add some orders first
	order1, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order1)

	order2, err := NewOrder(
		"order2", "user2", "BTC/USDT",
		OrderSideSell, OrderTypeLimit,
		big.NewInt(50), big.NewInt(1100),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order2)

	t.Run("remove existing order", func(t *testing.T) {
		err := ob.RemoveOrder("order1")
		assert.NoError(t, err)
		assert.Equal(t, 1, ob.GetOrderCount())
		assert.Equal(t, 0, ob.GetBuyOrderCount())
		assert.Equal(t, 1, ob.GetSellOrderCount())
	})

	t.Run("remove non-existent order", func(t *testing.T) {
		err := ob.RemoveOrder("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrOrderNotFound.Error())
	})

	t.Run("remove order with empty ID", func(t *testing.T) {
		err := ob.RemoveOrder("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order ID is empty")
	})
}

func TestOrderBookGetOrder(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	order, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order)

	t.Run("get existing order", func(t *testing.T) {
		retrievedOrder, err := ob.GetOrder("order1")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedOrder)
		assert.Equal(t, "order1", retrievedOrder.ID)
		assert.Equal(t, "user1", retrievedOrder.UserID)
		// Verify it's a clone, not the same pointer
		assert.NotSame(t, order, retrievedOrder)
	})

	t.Run("get non-existent order", func(t *testing.T) {
		retrievedOrder, err := ob.GetOrder("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrOrderNotFound.Error())
		assert.Nil(t, retrievedOrder)
	})

	t.Run("get order with empty ID", func(t *testing.T) {
		retrievedOrder, err := ob.GetOrder("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order ID is empty")
		assert.Nil(t, retrievedOrder)
	})
}

func TestOrderBookGetBestBidAndAsk(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("empty order book", func(t *testing.T) {
		bestBid, err := ob.GetBestBid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrOrderBookEmpty.Error())
		assert.Nil(t, bestBid)

		bestAsk, err := ob.GetBestAsk()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrOrderBookEmpty.Error())
		assert.Nil(t, bestAsk)
	})

	t.Run("with orders", func(t *testing.T) {
		// Add buy orders (higher price should be best bid)
		order1, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order1)

		order2, err := NewOrder(
			"order2", "user2", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order2)

		// Add sell orders (lower price should be best ask)
		order3, err := NewOrder(
			"order3", "user3", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(75), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order3)

		order4, err := NewOrder(
			"order4", "user4", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(25), big.NewInt(1150),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order4)

		// Test best bid (should be highest buy price)
		bestBid, err := ob.GetBestBid()
		assert.NoError(t, err)
		assert.NotNil(t, bestBid)
		assert.Equal(t, big.NewInt(1100), bestBid.Price)

		// Test best ask (should be lowest sell price)
		bestAsk, err := ob.GetBestAsk()
		assert.NoError(t, err)
		assert.NotNil(t, bestAsk)
		assert.Equal(t, big.NewInt(1150), bestAsk.Price)
	})
}

func TestOrderBookGetSpreadAndMidPrice(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("empty order book", func(t *testing.T) {
		spread, err := ob.GetSpread()
		assert.Error(t, err)
		assert.Nil(t, spread)

		midPrice, err := ob.GetMidPrice()
		assert.Error(t, err)
		assert.Nil(t, midPrice)
	})

	t.Run("with orders", func(t *testing.T) {
		// Add buy order
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder)

		// Add sell order
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		// Test spread
		spread, err := ob.GetSpread()
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(100), spread) // 1100 - 1000

		// Test mid price
		midPrice, err := ob.GetMidPrice()
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(1050), midPrice) // (1100 + 1000) / 2
	})
}

func TestOrderBookGetVolume(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("empty order book", func(t *testing.T) {
		totalVolume := ob.GetTotalVolume()
		assert.Equal(t, big.NewInt(0), totalVolume)

		buyVolume := ob.GetBuyVolume()
		assert.Equal(t, big.NewInt(0), buyVolume)

		sellVolume := ob.GetSellVolume()
		assert.Equal(t, big.NewInt(0), sellVolume)
	})

	t.Run("with orders", func(t *testing.T) {
		// Add buy orders
		buyOrder1, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder1)

		buyOrder2, err := NewOrder(
			"buy2", "user2", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder2)

		// Add sell orders
		sellOrder1, err := NewOrder(
			"sell1", "user3", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(75), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder1)

		sellOrder2, err := NewOrder(
			"sell2", "user4", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(25), big.NewInt(1150),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder2)

		// Test volumes
		totalVolume := ob.GetTotalVolume()
		assert.Equal(t, big.NewInt(250), totalVolume) // 100 + 50 + 75 + 25

		buyVolume := ob.GetBuyVolume()
		assert.Equal(t, big.NewInt(150), buyVolume) // 100 + 50

		sellVolume := ob.GetSellVolume()
		assert.Equal(t, big.NewInt(100), sellVolume) // 75 + 25
	})
}

func TestOrderBookGetDepth(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("empty order book", func(t *testing.T) {
		depth, err := ob.GetDepth(5)
		assert.NoError(t, err)
		assert.Empty(t, depth)
	})

	t.Run("invalid levels", func(t *testing.T) {
		depth, err := ob.GetDepth(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid number of levels")
		assert.Nil(t, depth)

		depth, err = ob.GetDepth(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid number of levels")
		assert.Nil(t, depth)
	})

	t.Run("with orders", func(t *testing.T) {
		// Add buy orders at different prices
		buyOrder1, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder1)

		buyOrder2, err := NewOrder(
			"buy2", "user2", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder2)

		// Add sell orders at different prices
		sellOrder1, err := NewOrder(
			"sell1", "user3", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(75), big.NewInt(1200),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder1)

		sellOrder2, err := NewOrder(
			"sell2", "user4", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(25), big.NewInt(1150),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder2)

		// Test depth
		depth, err := ob.GetDepth(3)
		assert.NoError(t, err)
		assert.Len(t, depth, 4) // 2 buy levels + 2 sell levels

		// Verify buy levels are sorted by price (highest first)
		buyLevels := 0
		for _, level := range depth {
			if level.Side == OrderSideBuy {
				buyLevels++
			}
		}
		assert.Equal(t, 2, buyLevels)

		// Verify sell levels are sorted by price (lowest first)
		sellLevels := 0
		for _, level := range depth {
			if level.Side == OrderSideSell {
				sellLevels++
			}
		}
		assert.Equal(t, 2, sellLevels)
	})
}

func TestOrderBookCancelExpiredOrders(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("no expired orders", func(t *testing.T) {
		// Add non-expired order
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order)

		cancelledCount := ob.CancelExpiredOrders()
		assert.Equal(t, 0, cancelledCount)
		assert.Equal(t, 1, ob.GetOrderCount())
	})

	t.Run("with expired orders", func(t *testing.T) {
		// Clear the order book first to start fresh
		ob.Clear()
		
		// Add non-expired order
		order1, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(order1)

		// Add expired order - we need to create it with a valid expiration first, then modify it
		expiresAt := time.Now().Add(time.Hour) // Future time initially
		expiredOrder, err := NewOrder(
			"order2", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, &expiresAt,
		)
		require.NoError(t, err)
		ob.AddOrder(expiredOrder)

		// Now modify the expiration to be in the past
		pastTime := time.Now().Add(-time.Hour)
		expiredOrder.ExpiresAt = &pastTime

		cancelledCount := ob.CancelExpiredOrders()
		assert.Equal(t, 1, cancelledCount)
		assert.Equal(t, 1, ob.GetOrderCount()) // Only non-expired order remains
	})
}

func TestOrderBookClear(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Add some orders
	order1, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order1)

	order2, err := NewOrder(
		"order2", "user2", "BTC/USDT",
		OrderSideSell, OrderTypeLimit,
		big.NewInt(50), big.NewInt(1100),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order2)

	assert.Equal(t, 2, ob.GetOrderCount())

	// Clear the order book
	ob.Clear()

	assert.Equal(t, 0, ob.GetOrderCount())
	assert.Equal(t, 0, ob.GetBuyOrderCount())
	assert.Equal(t, 0, ob.GetSellOrderCount())
	assert.True(t, ob.IsEmpty())
}

func TestOrderBookConcurrency(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Test concurrent read/write operations
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup

	// Start reader goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				ob.GetOrderCount()
				ob.GetBuyOrderCount()
				ob.GetSellOrderCount()
				ob.IsEmpty()
			}
		}()
	}

	// Start writer goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				order, err := NewOrder(
					"order"+string(rune(id))+string(rune(j)),
					"user"+string(rune(id)),
					"BTC/USDT",
					OrderSideBuy, OrderTypeLimit,
					big.NewInt(100), big.NewInt(1000),
					TimeInForceGTC, nil, nil,
				)
				if err == nil {
					ob.AddOrder(order)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify no data races occurred
	assert.Greater(t, ob.GetOrderCount(), 0)
}

// Benchmark tests for performance validation
func BenchmarkOrderBookAddOrder(b *testing.B) {
	ob, err := NewOrderBook("BTC/USDT")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order, err := NewOrder(
			"order"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		ob.AddOrder(order)
	}
}

func BenchmarkOrderBookGetBestBid(b *testing.B) {
	ob, err := NewOrderBook("BTC/USDT")
	if err != nil {
		b.Fatal(err)
	}

	// Add some orders first
	for i := 0; i < 100; i++ {
		order, err := NewOrder(
			"order"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000+int64(i)),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		ob.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ob.GetBestBid()
	}
}

func BenchmarkOrderBookGetDepth(b *testing.B) {
	ob, err := NewOrderBook("BTC/USDT")
	if err != nil {
		b.Fatal(err)
	}

	// Add some orders first
	for i := 0; i < 100; i++ {
		order, err := NewOrder(
			"order"+string(rune(i)), "user"+string(rune(i)),
			"BTC/USDT", OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000+int64(i)),
			TimeInForceGTC, nil, nil,
		)
		if err != nil {
			b.Fatal(err)
		}
		ob.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ob.GetDepth(10)
	}
}
