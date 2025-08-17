package options

import (
	"math/big"
	"testing"
	"time"
)

func TestNewOptionsEngine(t *testing.T) {
	engine := NewOptionsEngine()
	
	if engine == nil {
		t.Fatal("Expected options engine but got nil")
	}
	
	if engine.orders == nil {
		t.Error("Orders map not initialized")
	}
	
	if engine.positions == nil {
		t.Error("Positions map not initialized")
	}
	
	if engine.trades == nil {
		t.Error("Trades map not initialized")
	}
	
	if engine.orderBook == nil {
		t.Error("Order book not initialized")
	}
	
	if engine.riskManager == nil {
		t.Error("Risk manager not initialized")
	}
}

func TestNewOptionsOrderBook(t *testing.T) {
	orderBook := NewOptionsOrderBook()
	
	if orderBook == nil {
		t.Fatal("Expected order book but got nil")
	}
	
	if orderBook.buyOrders == nil {
		t.Error("Buy orders map not initialized")
	}
	
	if orderBook.sellOrders == nil {
		t.Error("Sell orders map not initialized")
	}
}

func TestNewOptionsRiskManager(t *testing.T) {
	riskManager := NewOptionsRiskManager()
	
	if riskManager == nil {
		t.Fatal("Expected risk manager but got nil")
	}
	
	if riskManager.maxPositionSize == nil {
		t.Error("Max position size not initialized")
	}
	
	if riskManager.maxLoss == nil {
		t.Error("Max loss not initialized")
	}
	
	if riskManager.positionLimits == nil {
		t.Error("Position limits map not initialized")
	}
	
	// Check default values
	expectedMaxPositionSize := big.NewFloat(1000)
	if riskManager.maxPositionSize.Cmp(expectedMaxPositionSize) != 0 {
		t.Errorf("Expected max position size %v, got %v", expectedMaxPositionSize, riskManager.maxPositionSize)
	}
	
	expectedMaxLoss := big.NewFloat(10000)
	if riskManager.maxLoss.Cmp(expectedMaxLoss) != 0 {
		t.Errorf("Expected max loss %v, got %v", expectedMaxLoss, riskManager.maxLoss)
	}
}

func TestOptionsEnginePlaceOrder(t *testing.T) {
	engine := NewOptionsEngine()
	
	// Create a valid option
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(110.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}
	
	// Create a valid order
	order := &OptionsOrder{
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	// Place order
	err = engine.PlaceOrder(order)
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}
	
	// Verify order was stored
	if order.ID == "" {
		t.Error("Order ID not set")
	}
	
	if order.Status != Pending {
		t.Errorf("Expected order status Pending, got %v", order.Status)
	}
	
	if order.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	
	if order.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
	
	// Verify order is in the engine
	retrievedOrder, err := engine.GetOrder(order.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve order: %v", err)
	}
	
	if retrievedOrder != order {
		t.Error("Retrieved order is not the same as placed order")
	}
}

func TestOptionsEnginePlaceOrderValidation(t *testing.T) {
	engine := NewOptionsEngine()
	
	tests := []struct {
		name        string
		order       *OptionsOrder
		expectError bool
		errorContains string
	}{
		{
			name:        "Nil order",
			order:       nil,
			expectError: true,
			errorContains: "order cannot be nil",
		},
		{
			name: "Nil option",
			order: &OptionsOrder{
				Option:   nil,
				Side:     Buy,
				Type:     Limit,
				Quantity: big.NewFloat(10),
				Price:    big.NewFloat(15.0),
				UserID:   "user1",
			},
			expectError: true,
			errorContains: "option cannot be nil",
		},
		{
			name: "Zero quantity",
			order: &OptionsOrder{
				Option:   createTestOption(t),
				Side:     Buy,
				Type:     Limit,
				Quantity: big.NewFloat(0),
				Price:    big.NewFloat(15.0),
				UserID:   "user1",
			},
			expectError: true,
			errorContains: "quantity must be positive",
		},
		{
			name: "Negative quantity",
			order: &OptionsOrder{
				Option:   createTestOption(t),
				Side:     Buy,
				Type:     Limit,
				Quantity: big.NewFloat(-10),
				Price:    big.NewFloat(15.0),
				UserID:   "user1",
			},
			expectError: true,
			errorContains: "quantity must be positive",
		},
		{
			name: "Negative price",
			order: &OptionsOrder{
				Option:   createTestOption(t),
				Side:     Buy,
				Type:     Limit,
				Quantity: big.NewFloat(10),
				Price:    big.NewFloat(-15.0),
				UserID:   "user1",
			},
			expectError: true,
			errorContains: "price must be non-negative",
		},
		{
			name: "Empty user ID",
			order: &OptionsOrder{
				Option:   createTestOption(t),
				Side:     Buy,
				Type:     Limit,
				Quantity: big.NewFloat(10),
				Price:    big.NewFloat(15.0),
				UserID:   "",
			},
			expectError: true,
			errorContains: "user ID is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.PlaceOrder(tt.order)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOptionsEngineCancelOrder(t *testing.T) {
	engine := NewOptionsEngine()
	
	// Create and place an order
	option := createTestOption(t)
	order := &OptionsOrder{
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err := engine.PlaceOrder(order)
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}
	
	// Cancel the order
	err = engine.CancelOrder(order.ID, "user1")
	if err != nil {
		t.Fatalf("Failed to cancel order: %v", err)
	}
	
	// Verify order status was updated
	if order.Status != Cancelled {
		t.Errorf("Expected order status Cancelled, got %v", order.Status)
	}
	
	// Verify order was removed from order book
	buyOrders, sellOrders := engine.orderBook.GetOrderBook()
	if len(buyOrders) != 0 || len(sellOrders) != 0 {
		t.Error("Order not removed from order book")
	}
}

func TestOptionsEngineCancelOrderErrors(t *testing.T) {
	engine := NewOptionsEngine()
	
	// Create and place an order first for the unauthorized user test
	option := createTestOption(t)
	order := &OptionsOrder{
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err := engine.PlaceOrder(order)
	if err != nil {
		t.Fatalf("Failed to place order for test: %v", err)
	}
	
	tests := []struct {
		name        string
		orderID     string
		userID      string
		expectError bool
		errorContains string
	}{
		{
			name:        "Order not found",
			orderID:     "nonexistent",
			userID:      "user1",
			expectError: true,
			errorContains: "order not found",
		},
		{
			name:        "Unauthorized user",
			orderID:     order.ID,
			userID:      "user2",
			expectError: true,
			errorContains: "unauthorized to cancel order",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.CancelOrder(tt.orderID, tt.userID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOptionsEngineOrderMatching(t *testing.T) {
	engine := NewOptionsEngine()
	
	// Create a call option
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(110.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}
	
	// Place a buy order
	buyOrder := &OptionsOrder{
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "buyer",
	}
	
	err = engine.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}
	
	// Place a matching sell order
	sellOrder := &OptionsOrder{
		Option:   option,
		Side:     Sell,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(14.0), // Lower price, should match
		UserID:   "seller",
	}
	
	err = engine.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}
	
	// Wait a bit for order matching to complete
	time.Sleep(100 * time.Millisecond)
	
	// Verify both orders were filled
	if buyOrder.Status != Filled {
		t.Errorf("Buy order status should be Filled, got %v", buyOrder.Status)
	}
	
	if sellOrder.Status != Filled {
		t.Errorf("Sell order status should be Filled, got %v", sellOrder.Status)
	}
	
	// Verify trade was created
	userTrades := engine.GetUserTrades("buyer")
	if len(userTrades) != 1 {
		t.Errorf("Expected 1 trade for buyer, got %d", len(userTrades))
	}
	
	// Verify positions were created
	buyerPositions := engine.GetUserPositions("buyer")
	if len(buyerPositions) != 1 {
		t.Errorf("Expected 1 position for buyer, got %d", len(buyerPositions))
	}
	
	sellerPositions := engine.GetUserPositions("seller")
	if len(sellerPositions) != 1 {
		t.Errorf("Expected 1 position for seller, got %d", len(sellerPositions))
	}
}

func TestOptionsEnginePositionTracking(t *testing.T) {
	engine := NewOptionsEngine()
	
	// Create a call option
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(110.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}
	
	// Place and execute a trade
	buyOrder := &OptionsOrder{
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "buyer",
	}
	
	sellOrder := &OptionsOrder{
		Option:   option,
		Side:     Sell,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(14.0),
		UserID:   "seller",
	}
	
	engine.PlaceOrder(buyOrder)
	engine.PlaceOrder(sellOrder)
	
	// Wait for matching
	time.Sleep(100 * time.Millisecond)
	
	// Verify buyer position
	buyerPositions := engine.GetUserPositions("buyer")
	if len(buyerPositions) != 1 {
		t.Fatalf("Expected 1 position for buyer, got %d", len(buyerPositions))
	}
	
	buyerPosition := buyerPositions[0]
	if buyerPosition.Quantity.Cmp(big.NewFloat(10)) != 0 {
		t.Errorf("Expected buyer quantity 10, got %v", buyerPosition.Quantity)
	}
	
	// Verify seller position (short position)
	sellerPositions := engine.GetUserPositions("seller")
	if len(sellerPositions) != 1 {
		t.Fatalf("Expected 1 position for seller, got %d", len(sellerPositions))
	}
	
	sellerPosition := sellerPositions[0]
	if sellerPosition.Quantity.Cmp(big.NewFloat(-10)) != 0 {
		t.Errorf("Expected seller quantity -10, got %v", sellerPosition.Quantity)
	}
}

func TestOptionsRiskManager(t *testing.T) {
	riskManager := NewOptionsRiskManager()
	
	// Test default limits
	order := &OptionsOrder{
		Option:   createTestOption(t),
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(500), // Within default limit
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err := riskManager.CheckRiskLimits(order)
	if err != nil {
		t.Errorf("Order within limits should not be rejected: %v", err)
	}
	
	// Test exceeding default position size
	largeOrder := &OptionsOrder{
		Option:   createTestOption(t),
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(1500), // Exceeds default limit
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err = riskManager.CheckRiskLimits(largeOrder)
	if err == nil {
		t.Error("Order exceeding limits should be rejected")
	}
	
	// Test user-specific limits
	riskManager.SetUserPositionLimit("user1", big.NewFloat(200))
	
	userLimitOrder := &OptionsOrder{
		Option:   createTestOption(t),
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(150), // Within user limit
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err = riskManager.CheckRiskLimits(userLimitOrder)
	if err != nil {
		t.Errorf("Order within user limits should not be rejected: %v", err)
	}
	
	// Test exceeding user limit
	exceedingUserLimitOrder := &OptionsOrder{
		Option:   createTestOption(t),
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(250), // Exceeds user limit
		Price:    big.NewFloat(15.0),
		UserID:   "user1",
	}
	
	err = riskManager.CheckRiskLimits(exceedingUserLimitOrder)
	if err == nil {
		t.Error("Order exceeding user limits should be rejected")
	}
}

func TestOptionsOrderBook(t *testing.T) {
	orderBook := NewOptionsOrderBook()
	
	// Create test orders
	option := createTestOption(t)
	
	buyOrder := &OptionsOrder{
		ID:       "buy1",
		Option:   option,
		Side:     Buy,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(15.0),
		UserID:   "buyer",
	}
	
	sellOrder := &OptionsOrder{
		ID:       "sell1",
		Option:   option,
		Side:     Sell,
		Type:     Limit,
		Quantity: big.NewFloat(10),
		Price:    big.NewFloat(16.0),
		UserID:   "seller",
	}
	
	// Add orders
	err := orderBook.AddOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to add buy order: %v", err)
	}
	
	err = orderBook.AddOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to add sell order: %v", err)
	}
	
	// Get order book
	buyOrders, sellOrders := orderBook.GetOrderBook()
	
	if len(buyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(buyOrders))
	}
	
	if len(sellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(sellOrders))
	}
	
	// Remove orders
	orderBook.RemoveOrder(buyOrder)
	orderBook.RemoveOrder(sellOrder)
	
	// Verify orders were removed
	buyOrders, sellOrders = orderBook.GetOrderBook()
	if len(buyOrders) != 0 {
		t.Errorf("Expected 0 buy orders after removal, got %d", len(buyOrders))
	}
	
	if len(sellOrders) != 0 {
		t.Errorf("Expected 0 sell orders after removal, got %d", len(sellOrders))
	}
}

func BenchmarkOptionsEnginePlaceOrder(b *testing.B) {
	engine := NewOptionsEngine()
	option := createTestOption(b)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := &OptionsOrder{
			Option:   option,
			Side:     Buy,
			Type:     Limit,
			Quantity: big.NewFloat(10),
			Price:    big.NewFloat(15.0),
			UserID:   "user1",
		}
		
		err := engine.PlaceOrder(order)
		if err != nil {
			b.Fatalf("Failed to place order: %v", err)
		}
	}
}

func BenchmarkOptionsEngineOrderMatching(b *testing.B) {
	engine := NewOptionsEngine()
	option := createTestOption(b)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Place buy order
		buyOrder := &OptionsOrder{
			Option:   option,
			Side:     Buy,
			Type:     Limit,
			Quantity: big.NewFloat(10),
			Price:    big.NewFloat(15.0),
			UserID:   "buyer",
		}
		
		engine.PlaceOrder(buyOrder)
		
		// Place matching sell order
		sellOrder := &OptionsOrder{
			Option:   option,
			Side:     Sell,
			Type:     Limit,
			Quantity: big.NewFloat(10),
			Price:    big.NewFloat(14.0),
			UserID:   "seller",
		}
		
		engine.PlaceOrder(sellOrder)
	}
}

// Helper function to create a test option
func createTestOption(t testing.TB) *Option {
	option, err := NewOption(
		Call,
		big.NewFloat(100.0), // Strike price
		big.NewFloat(110.0), // Current price
		big.NewFloat(1.0),   // Time to expiry
		big.NewFloat(0.05),  // Risk-free rate
		big.NewFloat(0.25),  // Volatility
	)
	if err != nil {
		t.Fatalf("Failed to create test option: %v", err)
	}
	return option
}
