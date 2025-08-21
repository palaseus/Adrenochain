package advancedorders

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdvancedOrder tests the AdvancedOrder functionality
func TestAdvancedOrder(t *testing.T) {
	t.Run("NewAdvancedOrder", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)
		require.NotNil(t, order)
		assert.Equal(t, "order_1", order.ID)
		assert.Equal(t, "BTC/USDT", order.Symbol)
		assert.Equal(t, OrderTypeConditional, order.Type)
		assert.Equal(t, OrderSideBuy, order.Side)
		assert.Equal(t, 100.0, order.Price)
		assert.Equal(t, 0.5, order.Quantity)
		assert.Equal(t, OrderStatusPending, order.Status)
		assert.True(t, order.IsActive())
		assert.False(t, order.IsExpired())
	})

	t.Run("AddCondition", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		// Add price condition
		priceCondition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}
		order.AddCondition(priceCondition)
		assert.Len(t, order.Conditions, 1)

		// Add time condition
		endTime := time.Now().Add(1 * time.Hour)
		timeCondition := &TimeCondition{
			EndTime: &endTime,
		}
		order.AddCondition(timeCondition)
		assert.Len(t, order.Conditions, 2)
	})

	t.Run("SetExecutionPlan", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		executionPlan := &MarketOrderExecution{}
		order.SetExecutionPlan(executionPlan)

		assert.NotNil(t, order.ExecutionPlan)
		assert.Equal(t, "market", order.ExecutionPlan.GetType())
	})

	t.Run("EvaluateConditions", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		// Add price condition
		priceCondition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}
		order.AddCondition(priceCondition)

		// Test with market data that meets condition
		marketData := MarketData{
			Symbol: "BTC/USDT",
			Price:  51000.0,
		}

		result := order.EvaluateConditions(marketData)
		assert.True(t, result)

		// Test with market data that doesn't meet condition
		marketData.Price = 49000.0
		result = order.EvaluateConditions(marketData)
		assert.False(t, result)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		order.UpdateStatus(OrderStatusActive)
		assert.Equal(t, OrderStatusActive, order.Status)

		order.UpdateStatus(OrderStatusFilled)
		assert.Equal(t, OrderStatusFilled, order.Status)
		assert.False(t, order.IsActive())
	})

	t.Run("Fill", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		order.Fill(0.3, 100.0)

		assert.Equal(t, 0.2, order.RemainingQuantity)
		assert.Equal(t, 0.3, order.FilledQuantity)
		assert.Equal(t, 0.6, order.GetProgress())
	})

	t.Run("Cancel", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		order.Cancel()
		assert.Equal(t, OrderStatusCancelled, order.Status)
		assert.False(t, order.IsActive())
	})

	t.Run("IsExpired", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		// Add time condition that expires in the past
		pastTime := time.Now().Add(-1 * time.Hour)
		order.ExpiresAt = &pastTime

		assert.True(t, order.IsExpired())
	})
}

// TestConditions tests the various condition types
func TestConditions(t *testing.T) {
	t.Run("PriceCondition", func(t *testing.T) {
		condition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}

		// Test above condition
		marketData := MarketData{Price: 51000.0}
		result := condition.Evaluate(marketData)
		assert.True(t, result)

		// Test below condition
		condition.Operator = "below"
		result = condition.Evaluate(marketData)
		assert.False(t, result)
	})

	t.Run("VolumeCondition", func(t *testing.T) {
		condition := &VolumeCondition{
			MinVolume: 1000000.0,
			MaxVolume: 2000000.0,
		}

		// Test within range
		marketData := MarketData{Volume: 1500000.0}
		result := condition.Evaluate(marketData)
		assert.True(t, result)

		// Test outside range
		marketData.Volume = 500000.0
		result = condition.Evaluate(marketData)
		assert.False(t, result)
	})

	t.Run("TimeCondition", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now().Add(1 * time.Hour)

		// Test with both start and end time
		condition := &TimeCondition{
			StartTime: &startTime,
			EndTime:   &endTime,
		}

		assert.Equal(t, "time", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Time between")

		// Test with only start time
		conditionStartOnly := &TimeCondition{
			StartTime: &startTime,
		}
		assert.Contains(t, conditionStartOnly.GetDescription(), "After")

		// Test with only end time
		conditionEndOnly := &TimeCondition{
			EndTime: &endTime,
		}
		assert.Contains(t, conditionEndOnly.GetDescription(), "Before")

		// Test with no time restrictions
		conditionNoTime := &TimeCondition{}
		assert.Equal(t, "No time restrictions", conditionNoTime.GetDescription())
	})

	t.Run("VolatilityCondition", func(t *testing.T) {
		condition := &VolatilityCondition{
			MinVolatility: 0.02,
			MaxVolatility: 0.05,
		}

		assert.Equal(t, "volatility", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Volatility between")
		assert.Contains(t, condition.GetDescription(), "0.0200")
		assert.Contains(t, condition.GetDescription(), "0.0500")
	})
}

// TestExecutionPlans tests the execution plan types
func TestExecutionPlans(t *testing.T) {
	t.Run("MarketOrderExecution", func(t *testing.T) {
		plan := &MarketOrderExecution{}

		assert.Equal(t, "market", plan.GetType())
		assert.Equal(t, "Execute at market price", plan.GetDescription())

		instruction := plan.CalculateExecution(0.5, MarketData{Price: 50000.0, Ask: 50001.0, Bid: 49999.0})
		assert.Equal(t, 0.5, instruction.Quantity)
		assert.Equal(t, 50001.0, instruction.Price) // Ask price for buy order
	})

	t.Run("LimitOrderExecution", func(t *testing.T) {
		plan := &LimitOrderExecution{
			LimitPrice: 50000.0,
		}

		assert.Equal(t, "limit", plan.GetType())
		assert.Equal(t, "Execute at limit price 50000.00", plan.GetDescription())

		instruction := plan.CalculateExecution(0.5, MarketData{Price: 51000.0})
		assert.Equal(t, 0.5, instruction.Quantity)
		assert.Equal(t, 50000.0, instruction.Price)
	})
}

// TestIntegration tests the complete order lifecycle
func TestIntegration(t *testing.T) {
	t.Run("CompleteOrderLifecycle", func(t *testing.T) {
		order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 100.0)

		// Add conditions
		priceCondition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}
		order.AddCondition(priceCondition)

		endTime := time.Now().Add(1 * time.Hour)
		timeCondition := &TimeCondition{
			EndTime: &endTime,
		}
		order.AddCondition(timeCondition)

		// Set execution plan
		executionPlan := &MarketOrderExecution{}
		order.SetExecutionPlan(executionPlan)

		// Test condition evaluation
		marketData := MarketData{
			Symbol: "BTC/USDT",
			Price:  51000.0,
			Volume: 1000000.0,
		}

		result := order.EvaluateConditions(marketData)
		assert.True(t, result)

		// Fill order
		order.Fill(0.3, 100.0)
		assert.Equal(t, 0.2, order.RemainingQuantity)
		assert.Equal(t, 0.3, order.FilledQuantity)

		// Complete order
		order.Fill(0.2, 100.0)
		assert.Equal(t, 0.0, order.RemainingQuantity)
		assert.Equal(t, 0.5, order.FilledQuantity)
		assert.Equal(t, 1.0, order.GetProgress())
	})
}

// TestGetUnrealizedPnL tests the GetUnrealizedPnL function
func TestGetUnrealizedPnL(t *testing.T) {
	order := NewAdvancedOrder("order_1", OrderTypeConditional, "BTC/USDT", OrderSideBuy, 1.0, 50000.0)

	// Fill part of the order
	order.Fill(0.5, 50000.0)

	// Test buy order PnL
	pnl := order.GetUnrealizedPnL(51000.0) // Price went up
	assert.Equal(t, 500.0, pnl)            // (51000 - 50000) * 0.5 = 500

	pnl = order.GetUnrealizedPnL(49000.0) // Price went down
	assert.Equal(t, -500.0, pnl)          // (49000 - 50000) * 0.5 = -500

	// Test sell order PnL
	sellOrder := NewAdvancedOrder("order_2", OrderTypeConditional, "BTC/USDT", OrderSideSell, 1.0, 50000.0)
	sellOrder.Fill(0.5, 50000.0)

	pnl = sellOrder.GetUnrealizedPnL(49000.0) // Price went down (good for sell)
	assert.Equal(t, 500.0, pnl)               // (50000 - 49000) * 0.5 = 500

	pnl = sellOrder.GetUnrealizedPnL(51000.0) // Price went up (bad for sell)
	assert.Equal(t, -500.0, pnl)              // (50000 - 51000) * 0.5 = -500
}

// TestConditionGetTypeAndDescription tests the GetType and GetDescription methods for different condition types
func TestConditionGetTypeAndDescription(t *testing.T) {
	t.Run("PriceCondition", func(t *testing.T) {
		condition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}

		assert.Equal(t, "price", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Price above")
		assert.Contains(t, condition.GetDescription(), "50000.00")
	})

	t.Run("VolumeCondition", func(t *testing.T) {
		condition := &VolumeCondition{
			MinVolume: 1000.0,
			MaxVolume: 5000.0,
		}

		assert.Equal(t, "volume", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Volume between")
		assert.Contains(t, condition.GetDescription(), "1000.00")
		assert.Contains(t, condition.GetDescription(), "5000.00")
	})

	t.Run("TimeCondition", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now().Add(1 * time.Hour)

		// Test with both start and end time
		condition := &TimeCondition{
			StartTime: &startTime,
			EndTime:   &endTime,
		}

		assert.Equal(t, "time", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Time between")

		// Test with only start time
		conditionStartOnly := &TimeCondition{
			StartTime: &startTime,
		}
		assert.Contains(t, conditionStartOnly.GetDescription(), "After")

		// Test with only end time
		conditionEndOnly := &TimeCondition{
			EndTime: &endTime,
		}
		assert.Contains(t, conditionEndOnly.GetDescription(), "Before")

		// Test with no time restrictions
		conditionNoTime := &TimeCondition{}
		assert.Equal(t, "No time restrictions", conditionNoTime.GetDescription())
	})

	t.Run("VolatilityCondition", func(t *testing.T) {
		condition := &VolatilityCondition{
			MinVolatility: 0.02,
			MaxVolatility: 0.05,
		}

		assert.Equal(t, "volatility", condition.GetType())
		assert.Contains(t, condition.GetDescription(), "Volatility between")
		assert.Contains(t, condition.GetDescription(), "0.0200")
		assert.Contains(t, condition.GetDescription(), "0.0500")
	})
}

// TestOrderManager tests the OrderManager functionality
func TestOrderManager(t *testing.T) {
	t.Run("NewOrderManager", func(t *testing.T) {
		manager := NewOrderManager()
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.orders)
		assert.NotNil(t, manager.logger)
	})

	t.Run("CreateOrder", func(t *testing.T) {
		manager := NewOrderManager()

		order := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)
		assert.NotNil(t, order)
		assert.Equal(t, "BTC/USDT", order.Symbol)
		assert.Equal(t, OrderSideBuy, order.Side)
		assert.Equal(t, 0.5, order.Quantity)
		assert.Equal(t, 50000.0, order.Price)

		// Verify order was added to manager
		assert.Len(t, manager.orders, 1)
	})

	t.Run("GetOrder", func(t *testing.T) {
		manager := NewOrderManager()

		order := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)

		// Get existing order
		retrievedOrder, exists := manager.GetOrder(order.ID)
		assert.True(t, exists)
		assert.Equal(t, order, retrievedOrder)

		// Get non-existent order
		_, exists = manager.GetOrder("non_existent")
		assert.False(t, exists)
	})

	t.Run("CancelOrder", func(t *testing.T) {
		manager := NewOrderManager()

		order := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)

		// Cancel existing order
		err := manager.CancelOrder(order.ID)
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusCancelled, order.Status)

		// Cancel non-existent order
		err = manager.CancelOrder("non_existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetActiveOrders", func(t *testing.T) {
		manager := NewOrderManager()

		// Create multiple orders
		order1 := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)
		order2 := manager.CreateOrder(OrderTypeConditional, "ETH/USDT", OrderSideSell, 1.0, 3000.0)

		// Cancel one order
		manager.CancelOrder(order1.ID)

		// Get active orders
		activeOrders := manager.GetActiveOrders()
		assert.Len(t, activeOrders, 1)
		assert.Equal(t, order2.ID, activeOrders[0].ID)
	})

	t.Run("ProcessMarketData", func(t *testing.T) {
		manager := NewOrderManager()

		// Create order with price condition
		order := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)
		priceCondition := &PriceCondition{
			TargetPrice: 50000.0,
			Operator:    "above",
			Side:        OrderSideBuy,
		}
		order.AddCondition(priceCondition)

		// Set execution plan
		executionPlan := &MarketOrderExecution{}
		order.SetExecutionPlan(executionPlan)

		// Process market data that meets condition
		marketData := MarketData{
			Symbol:    "BTC/USDT",
			Price:     51000.0,
			Timestamp: time.Now(),
		}

		manager.ProcessMarketData(marketData)
		// Note: We can't easily test the execution logging without mocking the logger
		// But we can verify the function doesn't panic
	})

	t.Run("CleanupExpiredOrders", func(t *testing.T) {
		manager := NewOrderManager()

		// Create order with explicit expiration time
		order := manager.CreateOrder(OrderTypeConditional, "BTC/USDT", OrderSideBuy, 0.5, 50000.0)

		// Set the order to expire in the past
		expiredTime := time.Now().Add(-1 * time.Hour)
		order.ExpiresAt = &expiredTime

		// Verify order is expired
		assert.True(t, order.IsExpired())

		// Cleanup expired orders
		manager.CleanupExpiredOrders()

		// Verify expired order was removed
		_, exists := manager.GetOrder(order.ID)
		assert.False(t, exists)
		assert.Len(t, manager.orders, 0)
	})
}
