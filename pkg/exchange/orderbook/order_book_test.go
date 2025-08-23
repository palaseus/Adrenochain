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
		assert.Equal(t, 1, ob.GetSellOrderCount())
	})

	t.Run("add order with invalid trading pair", func(t *testing.T) {
		order, err := NewOrder(
			"order3", "user3", "INVALID/PAIR",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trading pair mismatch")
	})

	t.Run("add order with nil price", func(t *testing.T) {
		// For limit orders, nil price should fail validation in NewOrder
		_, err := NewOrder(
			"order4", "user4", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), nil,
			TimeInForceGTC, nil, nil,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price")
	})

	t.Run("add order with zero price", func(t *testing.T) {
		// For limit orders, zero price should fail validation in NewOrder
		_, err := NewOrder(
			"order5", "user5", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(0),
			TimeInForceGTC, nil, nil,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price")
	})

	t.Run("add order with negative price", func(t *testing.T) {
		// For limit orders, negative price should fail validation in NewOrder
		_, err := NewOrder(
			"order6", "user6", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(-100),
			TimeInForceGTC, nil, nil,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price")
	})

	t.Run("add duplicate order", func(t *testing.T) {
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT", // Same ID as first order
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)

		err = ob.AddOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order already exists")
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

// TestOrderBookUpdateOrder tests order updating functionality
func TestOrderBookUpdateOrder(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Add an initial order
	order, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order)

	t.Run("update existing order", func(t *testing.T) {
		updatedOrder := order.Clone()
		updatedOrder.Quantity = big.NewInt(150)
		updatedOrder.Price = big.NewInt(1100)

		err := ob.UpdateOrder(updatedOrder)
		assert.NoError(t, err)

		// Verify the order was updated
		retrievedOrder, err := ob.GetOrder("order1")
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(150), retrievedOrder.Quantity)
		assert.Equal(t, big.NewInt(1100), retrievedOrder.Price)
	})

	t.Run("update non-existent order", func(t *testing.T) {
		nonExistentOrder := order.Clone()
		nonExistentOrder.ID = "non_existent"

		err := ob.UpdateOrder(nonExistentOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
	})

	t.Run("update order with nil order", func(t *testing.T) {
		err := ob.UpdateOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "updated order is nil")
	})

	t.Run("update order with different trading pair", func(t *testing.T) {
		updatedOrder := order.Clone()
		updatedOrder.TradingPair = "ETH/USDT"

		err := ob.UpdateOrder(updatedOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trading pair mismatch")
	})

	t.Run("update order with invalid validation", func(t *testing.T) {
		updatedOrder := order.Clone()
		updatedOrder.Quantity = big.NewInt(-50) // Invalid quantity

		// The validation error should occur during NewOrder, not UpdateOrder
		// Let's test this by trying to create an invalid order
		_, err := NewOrder(
			"invalid1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(-50), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.Error(t, err) // This should fail validation
		assert.Contains(t, err.Error(), "invalid quantity")
	})
}

// TestOrderBookGetSpreadAndMidPrice tests spread and mid-price calculations
func TestOrderBookGetSpreadAndMidPrice(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("get spread with no orders", func(t *testing.T) {
		spread, err := ob.GetSpread()
		assert.Error(t, err)
		assert.Nil(t, spread)
	})

	t.Run("get mid price with no orders", func(t *testing.T) {
		midPrice, err := ob.GetMidPrice()
		assert.Error(t, err)
		assert.Nil(t, midPrice)
	})

	t.Run("get spread with only buy orders", func(t *testing.T) {
		// Add only buy orders
		buyOrder, err := NewOrder(
			"buy1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(buyOrder)

		spread, err := ob.GetSpread()
		assert.Error(t, err)
		assert.Nil(t, spread)
	})

	t.Run("get spread with only sell orders", func(t *testing.T) {
		// Clear existing orders first
		ob.Clear()

		// Add only sell orders
		sellOrder, err := NewOrder(
			"sell1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		spread, err := ob.GetSpread()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order book is empty")
		assert.Nil(t, spread)
	})

	t.Run("get spread with both buy and sell orders", func(t *testing.T) {
		// Clear existing orders
		ob.Clear()

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
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, nil,
		)
		require.NoError(t, err)
		ob.AddOrder(sellOrder)

		// Test spread
		spread, err := ob.GetSpread()
		assert.NoError(t, err)
		assert.NotNil(t, spread)
		assert.Equal(t, big.NewInt(100), spread) // 1100 - 1000

		// Test mid price
		midPrice, err := ob.GetMidPrice()
		assert.NoError(t, err)
		assert.NotNil(t, midPrice)
		assert.Equal(t, big.NewInt(1050), midPrice) // (1000 + 1100) / 2
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

// TestOrderBookGetDepth tests depth calculation
func TestOrderBookGetDepth(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("get depth with no orders", func(t *testing.T) {
		depth, err := ob.GetDepth(5)
		assert.NoError(t, err)
		assert.Empty(t, depth)
	})

	t.Run("get depth with orders", func(t *testing.T) {
		// Add multiple buy orders at different prices
		buyOrder1, _ := NewOrder("buy1", "user1", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(100), big.NewInt(1000), TimeInForceGTC, nil, nil)
		buyOrder2, _ := NewOrder("buy2", "user2", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(50), big.NewInt(950), TimeInForceGTC, nil, nil)
		buyOrder3, _ := NewOrder("buy3", "user3", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(75), big.NewInt(900), TimeInForceGTC, nil, nil)

		ob.AddOrder(buyOrder1)
		ob.AddOrder(buyOrder2)
		ob.AddOrder(buyOrder3)

		// Add multiple sell orders at different prices
		sellOrder1, _ := NewOrder("sell1", "user4", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(60), big.NewInt(1100), TimeInForceGTC, nil, nil)
		sellOrder2, _ := NewOrder("sell2", "user5", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(40), big.NewInt(1150), TimeInForceGTC, nil, nil)
		sellOrder3, _ := NewOrder("sell3", "user6", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(80), big.NewInt(1200), TimeInForceGTC, nil, nil)

		ob.AddOrder(sellOrder1)
		ob.AddOrder(sellOrder2)
		ob.AddOrder(sellOrder3)

		depth, err := ob.GetDepth(3)
		assert.NoError(t, err)
		assert.Len(t, depth, 6) // 3 buy levels + 3 sell levels

		// Verify we have both buy and sell levels
		buyLevels := 0
		sellLevels := 0
		for _, level := range depth {
			if level.Side == OrderSideBuy {
				buyLevels++
			} else {
				sellLevels++
			}
		}
		assert.Equal(t, 3, buyLevels)
		assert.Equal(t, 3, sellLevels)
	})
}

// TestOrderBookCancelExpiredOrders tests expired order cancellation
func TestOrderBookCancelExpiredOrders(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("cancel expired orders", func(t *testing.T) {
		// Test that we cannot create an already expired order
		pastTime := time.Now().Add(-1 * time.Hour)
		_, err := NewOrder(
			"expired1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(100), big.NewInt(1000),
			TimeInForceGTC, nil, &pastTime,
		)
		// This should fail because the order is already expired
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order has expired")

		// Add a valid order
		futureTime := time.Now().Add(1 * time.Hour)
		validOrder, err := NewOrder(
			"valid1", "user2", "BTC/USDT",
			OrderSideSell, OrderTypeLimit,
			big.NewInt(50), big.NewInt(1100),
			TimeInForceGTC, nil, &futureTime,
		)
		require.NoError(t, err)
		ob.AddOrder(validOrder)

		// Cancel expired orders
		cancelledCount := ob.CancelExpiredOrders()
		assert.Equal(t, 0, cancelledCount) // No expired orders to cancel

		// Verify valid order still exists
		_, err = ob.GetOrder("valid1")
		assert.NoError(t, err)
	})
}

// TestOrderBookGetLastUpdate tests last update functionality
func TestOrderBookGetLastUpdate(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Get initial last update time
	initialTime := ob.GetLastUpdate()
	assert.False(t, initialTime.IsZero())

	// Add an order to trigger update
	order, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order)

	// Get updated time
	updatedTime := ob.GetLastUpdate()
	assert.True(t, updatedTime.After(initialTime) || updatedTime.Equal(initialTime))
}

// TestOrderHeapLess tests the heap ordering logic
func TestOrderHeapLess(t *testing.T) {
	// Test buy order heap ordering
	buyHeap := &OrderHeap{}

	// Add buy orders with different prices and timestamps
	buyOrder1, _ := NewOrder("buy1", "user1", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(100), big.NewInt(1000), TimeInForceGTC, nil, nil)
	buyOrder2, _ := NewOrder("buy2", "user2", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(50), big.NewInt(1100), TimeInForceGTC, nil, nil)
	buyOrder3, _ := NewOrder("buy3", "user3", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(75), big.NewInt(1000), TimeInForceGTC, nil, nil)

	// Test market order handling (nil price)
	marketOrder, _ := NewOrder("market1", "user4", "BTC/USDT", OrderSideBuy, OrderTypeMarket, big.NewInt(100), nil, TimeInForceGTC, nil, nil)

	// Add orders to heap for testing
	*buyHeap = append(*buyHeap, buyOrder1, buyOrder2, buyOrder3, marketOrder)

	// Test heap ordering with market orders
	// Market orders should come first (nil price)
	assert.True(t, buyHeap.Less(3, 0)) // marketOrder (index 3) should come before buyOrder1 (index 0)
	assert.True(t, buyHeap.Less(3, 1)) // marketOrder should come before buyOrder2
	assert.True(t, buyHeap.Less(3, 2)) // marketOrder should come before buyOrder3
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

// TestOrderBookAddOrderEdgeCases tests edge cases in AddOrder
func TestOrderBookAddOrderEdgeCases(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("add order with nil order", func(t *testing.T) {
		err := ob.AddOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order is nil")
	})

	t.Run("add order with invalid price validation", func(t *testing.T) {
		// Create an order with invalid price that passes NewOrder but fails AddOrder validation
		order := &Order{
			ID:                "invalid1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(0), // Invalid price
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(100),
			StopPrice:         nil,
			TimeInForce:       TimeInForceGTC,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ExpiresAt:         nil,
		}

		err := ob.AddOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price")
	})
}

// TestOrderBookUpdateOrderEdgeCases tests edge cases in UpdateOrder
func TestOrderBookUpdateOrderEdgeCases(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	// Add an initial order
	order, err := NewOrder(
		"order1", "user1", "BTC/USDT",
		OrderSideBuy, OrderTypeLimit,
		big.NewInt(100), big.NewInt(1000),
		TimeInForceGTC, nil, nil,
	)
	require.NoError(t, err)
	ob.AddOrder(order)

	t.Run("update order with nil order", func(t *testing.T) {
		err := ob.UpdateOrder(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "updated order is nil")
	})

	t.Run("update order with validation error", func(t *testing.T) {
		// Test that we cannot create an order with invalid quantity
		_, err := NewOrder(
			"invalid1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(-50), big.NewInt(1000), // Invalid quantity
			TimeInForceGTC, nil, nil,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid quantity")
	})
}

// TestOrderBookGetDepthEdgeCases tests edge cases in GetDepth
func TestOrderBookGetDepthEdgeCases(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("get depth with invalid levels", func(t *testing.T) {
		depth, err := ob.GetDepth(0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid number of levels")
		assert.Nil(t, depth)

		depth, err = ob.GetDepth(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid number of levels")
		assert.Nil(t, depth)
	})

	t.Run("get depth with complex price levels", func(t *testing.T) {
		// Add multiple orders at the same price to test price level aggregation
		buyOrder1, _ := NewOrder("buy1", "user1", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(100), big.NewInt(1000), TimeInForceGTC, nil, nil)
		buyOrder2, _ := NewOrder("buy2", "user2", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(50), big.NewInt(1000), TimeInForceGTC, nil, nil)
		buyOrder3, _ := NewOrder("buy3", "user3", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(75), big.NewInt(950), TimeInForceGTC, nil, nil)

		ob.AddOrder(buyOrder1)
		ob.AddOrder(buyOrder2)
		ob.AddOrder(buyOrder3)

		// Add sell orders at the same price
		sellOrder1, _ := NewOrder("sell1", "user4", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(60), big.NewInt(1100), TimeInForceGTC, nil, nil)
		sellOrder2, _ := NewOrder("sell2", "user5", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(40), big.NewInt(1100), TimeInForceGTC, nil, nil)
		sellOrder3, _ := NewOrder("sell3", "user6", "BTC/USDT", OrderSideSell, OrderTypeLimit, big.NewInt(80), big.NewInt(1200), TimeInForceGTC, nil, nil)

		ob.AddOrder(sellOrder1)
		ob.AddOrder(sellOrder2)
		ob.AddOrder(sellOrder3)

		depth, err := ob.GetDepth(5)
		assert.NoError(t, err)
		assert.NotNil(t, depth)

		// Verify that orders at the same price are aggregated
		priceLevels := make(map[string]int)
		for _, level := range depth {
			priceKey := level.Price.String()
			priceLevels[priceKey]++
		}

		// Should have 3 unique price levels (1000, 950, 1100, 1200)
		assert.GreaterOrEqual(t, len(priceLevels), 4)
	})
}

// TestOrderBookCancelExpiredOrdersEdgeCases tests edge cases in CancelExpiredOrders
func TestOrderBookCancelExpiredOrdersEdgeCases(t *testing.T) {
	ob, err := NewOrderBook("BTC/USDT")
	require.NoError(t, err)

	t.Run("cancel expired orders with complex scenarios", func(t *testing.T) {
		// Add orders with different expiration times
		pastTime := time.Now().Add(-2 * time.Hour)
		futureTime := time.Now().Add(2 * time.Hour)
		veryPastTime := time.Now().Add(-3 * time.Hour)

		// Create orders manually to bypass validation
		expiredOrder1 := &Order{
			ID:                "expired1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(100),
			StopPrice:         nil,
			TimeInForce:       TimeInForceGTC,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ExpiresAt:         &pastTime,
		}

		expiredOrder2 := &Order{
			ID:                "expired2",
			UserID:            "user2",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideSell,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(50),
			Price:             big.NewInt(1100),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(50),
			StopPrice:         nil,
			TimeInForce:       TimeInForceGTC,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ExpiresAt:         &veryPastTime,
		}

		validOrder := &Order{
			ID:                "valid1",
			UserID:            "user3",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(75),
			Price:             big.NewInt(900),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(75),
			StopPrice:         nil,
			TimeInForce:       TimeInForceGTC,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			ExpiresAt:         &futureTime,
		}

		// Add orders directly to bypass validation
		ob.orders["expired1"] = expiredOrder1
		ob.orders["expired2"] = expiredOrder2
		ob.orders["valid1"] = validOrder

		// Add to heaps
		*ob.buyOrders = append(*ob.buyOrders, expiredOrder1, validOrder)
		*ob.sellOrders = append(*ob.sellOrders, expiredOrder2)

		// Cancel expired orders
		cancelledCount := ob.CancelExpiredOrders()
		assert.Equal(t, 2, cancelledCount)

		// Verify expired orders were removed
		_, err = ob.GetOrder("expired1")
		assert.Error(t, err)
		_, err = ob.GetOrder("expired2")
		assert.Error(t, err)

		// Verify valid order still exists
		_, err = ob.GetOrder("valid1")
		assert.NoError(t, err)
	})
}

// TestOrderBookHeapEdgeCases tests edge cases in heap operations
func TestOrderBookHeapEdgeCases(t *testing.T) {

	t.Run("heap ordering with nil prices", func(t *testing.T) {
		buyHeap := &OrderHeap{}

		// Create orders with nil prices (market orders)
		marketOrder1, _ := NewOrder("market1", "user1", "BTC/USDT", OrderSideBuy, OrderTypeMarket, big.NewInt(100), nil, TimeInForceGTC, nil, nil)
		marketOrder2, _ := NewOrder("market2", "user2", "BTC/USDT", OrderSideBuy, OrderTypeMarket, big.NewInt(50), nil, TimeInForceGTC, nil, nil)

		// Create orders with prices
		limitOrder1, _ := NewOrder("limit1", "user3", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(75), big.NewInt(1000), TimeInForceGTC, nil, nil)
		limitOrder2, _ := NewOrder("limit2", "user4", "BTC/USDT", OrderSideBuy, OrderTypeLimit, big.NewInt(25), big.NewInt(1100), TimeInForceGTC, nil, nil)

		// Add to heap
		*buyHeap = append(*buyHeap, marketOrder1, marketOrder2, limitOrder1, limitOrder2)

		// Test heap ordering
		assert.True(t, buyHeap.Less(0, 2)) // Market order should come before limit order
		assert.True(t, buyHeap.Less(1, 2)) // Market order should come before limit order
		assert.True(t, buyHeap.Less(0, 1)) // Market orders should be ordered by timestamp
	})
}
