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
		
		condition := &TimeCondition{
			StartTime: &startTime,
			EndTime:   &endTime,
		}
		
		// Test within time range
		marketData := MarketData{Timestamp: time.Now()}
		result := condition.Evaluate(marketData)
		assert.True(t, result)
	})

	t.Run("VolatilityCondition", func(t *testing.T) {
		condition := &VolatilityCondition{
			MinVolatility: 0.02,
			MaxVolatility: 0.05,
		}
		
		// Test within range
		marketData := MarketData{Volatility: 0.03}
		result := condition.Evaluate(marketData)
		assert.True(t, result)
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
