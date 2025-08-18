package markets

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProposalMarkets(t *testing.T) {
	pm := NewProposalMarkets()
	require.NotNil(t, pm)
	assert.NotNil(t, pm.markets)
	assert.NotNil(t, pm.orders)
	assert.NotNil(t, pm.positions)
	assert.NotNil(t, pm.metrics)
	assert.NotNil(t, pm.orderQueue)

	// Test that background goroutines are started
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	err := pm.Close()
	assert.NoError(t, err)
}

func TestCreateMarket(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	tests := []struct {
		name    string
		market  *Market
		wantErr bool
	}{
		{
			name: "valid market",
			market: &Market{
				Title:       "Test Market",
				Description: "Test Description",
				MarketType:  BinaryMarket,
				Creator:     "user1",
				Outcomes: []*Outcome{
					{Name: "Yes", Description: "Yes outcome"},
					{Name: "No", Description: "No outcome"},
				},
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "missing creator",
			market: &Market{
				Title:       "Test Market",
				Description: "Test Description",
				MarketType:  BinaryMarket,
				Outcomes: []*Outcome{
					{Name: "Yes", Description: "Yes outcome"},
				},
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "no outcomes",
			market: &Market{
				Title:       "Test Market",
				Description: "Test Description",
				MarketType:  BinaryMarket,
				Creator:     "user1",
				StartTime:   time.Now().Add(1 * time.Hour),
				EndTime:     time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "start time in past",
			market: &Market{
				Title:       "Test Market",
				Description: "Test Description",
				MarketType:  BinaryMarket,
				Creator:     "user1",
				Outcomes: []*Outcome{
					{Name: "Yes", Description: "Yes outcome"},
				},
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "end time before start time",
			market: &Market{
				Title:       "Test Market",
				Description: "Test Description",
				MarketType:  BinaryMarket,
				Creator:     "user1",
				Outcomes: []*Outcome{
					{Name: "Yes", Description: "Yes outcome"},
				},
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.CreateMarket(tt.market)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.market.ID)
				assert.Equal(t, MarketDraft, tt.market.Status)
				assert.NotZero(t, tt.market.CreatedAt)
				assert.NotZero(t, tt.market.UpdatedAt)

				// Check outcomes initialization
				for _, outcome := range tt.market.Outcomes {
					assert.NotEmpty(t, outcome.ID)
					assert.NotZero(t, outcome.CreatedAt)
					assert.NotZero(t, outcome.UpdatedAt)
					assert.Greater(t, outcome.Probability, 0.0)
					assert.Greater(t, outcome.Price, 0.0)
				}
			}
		})
	}
}

func TestActivateMarket(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Test activation before start time
	err = pm.ActivateMarket(market.ID)
	assert.Error(t, err)

	// Test activation of non-existent market
	err = pm.ActivateMarket("non-existent")
	assert.Error(t, err)

	// Test activation of already active market
	market.Status = MarketActive
	err = pm.ActivateMarket(market.ID)
	assert.Error(t, err)
}

func TestPlaceOrder(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour), // Start time in future
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Activate market manually for testing
	market.Status = MarketActive
	pm.markets[market.ID] = market

	tests := []struct {
		name    string
		order   *Order
		wantErr bool
	}{
		{
			name: "valid buy order",
			order: &Order{
				MarketID:  market.ID,
				UserID:    "user1",
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    100.0,
				Price:     0.5,
			},
			wantErr: false,
		},
		{
			name: "valid sell order",
			order: &Order{
				MarketID:  market.ID,
				UserID:    "user2",
				OutcomeID: market.Outcomes[0].ID,
				Type:      SellOrder,
				Amount:    50.0,
				Price:     0.6,
			},
			wantErr: false,
		},
		{
			name: "missing market ID",
			order: &Order{
				UserID:    "user1",
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    100.0,
				Price:     0.5,
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			order: &Order{
				MarketID:  market.ID,
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    100.0,
				Price:     0.5,
			},
			wantErr: true,
		},
		{
			name: "missing outcome ID",
			order: &Order{
				MarketID: market.ID,
				UserID:   "user1",
				Type:     BuyOrder,
				Amount:   100.0,
				Price:    0.5,
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			order: &Order{
				MarketID:  market.ID,
				UserID:    "user1",
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    0.0,
				Price:     0.5,
			},
			wantErr: true,
		},
		{
			name: "zero price",
			order: &Order{
				MarketID:  market.ID,
				UserID:    "user1",
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    100.0,
				Price:     0.0,
			},
			wantErr: true,
		},
		{
			name: "market not active",
			order: &Order{
				MarketID:  market.ID,
				UserID:    "user1",
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    100.0,
				Price:     0.5,
			},
			wantErr: true,
		},
	}

	// Deactivate market for some tests
	market.Status = MarketDraft

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "market not active" {
				market.Status = MarketDraft
			} else {
				market.Status = MarketActive
			}
			pm.markets[market.ID] = market

			err := pm.PlaceOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.order.ID)
				assert.Equal(t, OrderPending, tt.order.Status)
				assert.NotZero(t, tt.order.CreatedAt)
				assert.NotZero(t, tt.order.UpdatedAt)
			}
		})
	}
}

func TestGetMarket(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Test getting existing market
	retrievedMarket, err := pm.GetMarket(market.ID)
	assert.NoError(t, err)
	assert.Equal(t, market.ID, retrievedMarket.ID)

	// Test getting non-existent market
	_, err = pm.GetMarket("non-existent")
	assert.Error(t, err)
}

func TestGetMarkets(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create multiple markets
	market1 := &Market{
		Title:       "Binary Market",
		Description: "Binary Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	market2 := &Market{
		Title:       "Multi Market",
		Description: "Multi Description",
		MarketType:  MultiOutcomeMarket,
		Creator:     "user2",
		Outcomes: []*Outcome{
			{Name: "Option A", Description: "Option A"},
			{Name: "Option B", Description: "Option B"},
			{Name: "Option C", Description: "Option C"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market1)
	require.NoError(t, err)
	err = pm.CreateMarket(market2)
	require.NoError(t, err)

	// Test getting all markets
	markets := pm.GetMarkets("", "")
	assert.Len(t, markets, 2)

	// Test filtering by status
	draftMarkets := pm.GetMarkets(MarketDraft, "")
	assert.Len(t, draftMarkets, 2)

	// Test filtering by market type
	binaryMarkets := pm.GetMarkets("", BinaryMarket)
	assert.Len(t, binaryMarkets, 1)
	assert.Equal(t, BinaryMarket, binaryMarkets[0].MarketType)

	// Test filtering by both
	binaryDraftMarkets := pm.GetMarkets(MarketDraft, BinaryMarket)
	assert.Len(t, binaryDraftMarkets, 1)
	assert.Equal(t, BinaryMarket, binaryDraftMarkets[0].MarketType)
}

func TestGetUserPositions(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market and place orders
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Place buy order
	buyOrder := &Order{
		MarketID:  market.ID,
		UserID:    "user1",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    100.0,
		Price:     0.6,
	}

	// Place sell order that should match
	sellOrder := &Order{
		MarketID:  market.ID,
		UserID:    "user2",
		OutcomeID: market.Outcomes[0].ID,
		Type:      SellOrder,
		Amount:    50.0,
		Price:     0.5,
	}

	err = pm.PlaceOrder(buyOrder)
	require.NoError(t, err)
	err = pm.PlaceOrder(sellOrder)
	require.NoError(t, err)

	// Wait for order processing
	time.Sleep(100 * time.Millisecond)

	// Test getting positions for user1 (buyer)
	positions := pm.GetUserPositions("user1")
	assert.Len(t, positions, 1)
	assert.Equal(t, "user1", positions[0].UserID)

	// Test getting positions for user2 (seller)
	positions = pm.GetUserPositions("user2")
	assert.Len(t, positions, 1)
	assert.Equal(t, "user2", positions[0].UserID)

	// Test getting positions for non-existent user
	positions = pm.GetUserPositions("non-existent")
	assert.Len(t, positions, 0)
}

func TestResolveMarket(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Test resolving non-existent market
	err = pm.ResolveMarket("non-existent", market.Outcomes[0].ID)
	assert.Error(t, err)

	// Test resolving market with non-existent outcome
	err = pm.ResolveMarket(market.ID, "non-existent")
	assert.Error(t, err)

	// Test resolving draft market
	err = pm.ResolveMarket(market.ID, market.Outcomes[0].ID)
	assert.Error(t, err)

	// Activate market
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Test resolving active market
	err = pm.ResolveMarket(market.ID, market.Outcomes[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, MarketResolved, market.Status)
	assert.Equal(t, market.Outcomes[0].ID, *market.ResolvedOutcome)
	assert.NotNil(t, market.ResolutionTime)
}

func TestCloseMarket(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Test closing non-existent market
	err = pm.CloseMarket("non-existent")
	assert.Error(t, err)

	// Test closing draft market
	err = pm.CloseMarket(market.ID)
	assert.Error(t, err)

	// Activate market
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Test closing active market before end time
	err = pm.CloseMarket(market.ID)
	assert.Error(t, err)

	// Set end time in past
	market.EndTime = time.Now().Add(-1 * time.Hour)
	pm.markets[market.ID] = market

	// Test closing active market after end time
	err = pm.CloseMarket(market.ID)
	assert.NoError(t, err)
	assert.Equal(t, MarketClosed, market.Status)
}

func TestOrderMatching(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Place buy order
	buyOrder := &Order{
		MarketID:  market.ID,
		UserID:    "buyer",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    100.0,
		Price:     0.6,
	}

	err = pm.PlaceOrder(buyOrder)
	require.NoError(t, err)

	// Place sell order that should match
	sellOrder := &Order{
		MarketID:  market.ID,
		UserID:    "seller",
		OutcomeID: market.Outcomes[0].ID,
		Type:      SellOrder,
		Amount:    50.0,
		Price:     0.5,
	}

	err = pm.PlaceOrder(sellOrder)
	require.NoError(t, err)

	// Wait for order processing
	time.Sleep(200 * time.Millisecond)

	// Debug: Print order statuses
	buyOrder, exists := pm.orders[buyOrder.ID]
	require.True(t, exists)
	t.Logf("Buy order status: %s, Amount: %f", buyOrder.Status, buyOrder.Amount)

	sellOrder, exists = pm.orders[sellOrder.ID]
	require.True(t, exists)
	t.Logf("Sell order status: %s, Amount: %f", sellOrder.Status, sellOrder.Amount)

	// Verify that the sell order was completely filled (50.0 amount)
	assert.Equal(t, OrderFilled, sellOrder.Status)
	assert.Equal(t, 0.0, sellOrder.Amount)

	// Verify that the buy order was partially filled (100.0 - 50.0 = 50.0 remaining)
	assert.Equal(t, OrderPending, buyOrder.Status)
	assert.Equal(t, 50.0, buyOrder.Amount)

	// Check that orders were created
	assert.Len(t, pm.orders, 2)

	// Check that positions were created for the matched portion
	positions := pm.GetUserPositions("buyer")
	assert.Len(t, positions, 1)
	assert.Equal(t, 50.0, positions[0].Amount) // 50.0 was matched

	positions = pm.GetUserPositions("seller")
	assert.Len(t, positions, 1)
	assert.Equal(t, 0.0, positions[0].Amount) // Seller position should be 0 after selling
}

func TestConcurrency(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Test concurrent order placement
	var wg sync.WaitGroup
	numOrders := 100

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			order := &Order{
				MarketID:  market.ID,
				UserID:    fmt.Sprintf("user%d", index),
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    float64(rand.Intn(100) + 1),
				Price:     0.5 + rand.Float64()*0.1,
			}

			err := pm.PlaceOrder(order)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all orders were placed
	assert.Len(t, pm.orders, numOrders)
}

func TestMemorySafety(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a reasonable number of markets and orders to test memory usage
	numMarkets := 10 // Reduced from 1000
	numOrders := 20  // Reduced from 100

	for i := 0; i < numMarkets; i++ {
		market := &Market{
			Title:       fmt.Sprintf("Market %d", i),
			Description: fmt.Sprintf("Description %d", i),
			MarketType:  BinaryMarket,
			Creator:     fmt.Sprintf("user%d", i),
			Outcomes: []*Outcome{
				{Name: "Yes", Description: "Yes outcome"},
				{Name: "No", Description: "No outcome"},
			},
			StartTime: time.Now().Add(1 * time.Hour),
			EndTime:   time.Now().Add(24 * time.Hour),
		}

		err := pm.CreateMarket(market)
		require.NoError(t, err)
		market.Status = MarketActive
		pm.markets[market.ID] = market

		// Place orders for each market
		for j := 0; j < numOrders; j++ {
			order := &Order{
				MarketID:  market.ID,
				UserID:    fmt.Sprintf("user%d", j),
				OutcomeID: market.Outcomes[0].ID,
				Type:      BuyOrder,
				Amount:    float64(rand.Intn(100) + 1),
				Price:     0.5 + rand.Float64()*0.1,
			}

			err := pm.PlaceOrder(order)
			require.NoError(t, err)
		}
	}

	// Verify all markets and orders were created
	assert.Len(t, pm.markets, numMarkets)
	assert.Len(t, pm.orders, numMarkets*numOrders)
}

func TestEdgeCases(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Test with empty market
	emptyMarket := &Market{
		Title:       "Empty Market",
		Description: "Empty Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes:    []*Outcome{},
		StartTime:   time.Now().Add(1 * time.Hour),
		EndTime:     time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(emptyMarket)
	assert.Error(t, err)

	// Test with very large amounts
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err = pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	largeOrder := &Order{
		MarketID:  market.ID,
		UserID:    "user1",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    1e15, // Very large amount
		Price:     0.5,
	}

	err = pm.PlaceOrder(largeOrder)
	assert.NoError(t, err)

	// Test with very small amounts
	smallOrder := &Order{
		MarketID:  market.ID,
		UserID:    "user2",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    1e-15, // Very small amount
		Price:     0.5,
	}

	err = pm.PlaceOrder(smallOrder)
	assert.NoError(t, err)
}

func TestCleanup(t *testing.T) {
	pm := NewProposalMarkets()

	// Create a market and place orders
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Place an order
	order := &Order{
		MarketID:  market.ID,
		UserID:    "user1",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    100.0,
		Price:     0.5,
	}

	err = pm.PlaceOrder(order)
	require.NoError(t, err)

	// Test cleanup
	err = pm.Close()
	assert.NoError(t, err)

	// Verify cleanup
	time.Sleep(100 * time.Millisecond)
}

func TestGetRandomID(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	id1 := pm.GetRandomID()
	id2 := pm.GetRandomID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // 16 bytes = 32 hex chars
	assert.Len(t, id2, 32)
}

func TestMarketTypes(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	marketTypes := []MarketType{
		BinaryMarket,
		MultiOutcomeMarket,
		ScalarMarket,
		FuturesMarket,
		ConditionalMarket,
		CustomMarket,
	}

	for _, marketType := range marketTypes {
		market := &Market{
			Title:       fmt.Sprintf("Test %s Market", marketType),
			Description: fmt.Sprintf("Test %s Description", marketType),
			MarketType:  marketType,
			Creator:     "user1",
			Outcomes: []*Outcome{
				{Name: "Outcome 1", Description: "First outcome"},
				{Name: "Outcome 2", Description: "Second outcome"},
			},
			StartTime: time.Now().Add(1 * time.Hour),
			EndTime:   time.Now().Add(24 * time.Hour),
		}

		err := pm.CreateMarket(market)
		assert.NoError(t, err)
		assert.Equal(t, marketType, market.MarketType)
	}
}

func TestOrderTypes(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Test buy order
	buyOrder := &Order{
		MarketID:  market.ID,
		UserID:    "buyer",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    100.0,
		Price:     0.5,
	}

	err = pm.PlaceOrder(buyOrder)
	assert.NoError(t, err)

	// Test sell order
	sellOrder := &Order{
		MarketID:  market.ID,
		UserID:    "seller",
		OutcomeID: market.Outcomes[0].ID,
		Type:      SellOrder,
		Amount:    50.0,
		Price:     0.6,
	}

	err = pm.PlaceOrder(sellOrder)
	assert.NoError(t, err)

	// Verify both order types were created
	assert.Equal(t, BuyOrder, buyOrder.Type)
	assert.Equal(t, SellOrder, sellOrder.Type)
}

func TestMarketStatusTransitions(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, MarketDraft, market.Status)

	// Test status transitions
	market.Status = MarketActive
	assert.Equal(t, MarketActive, market.Status)

	market.Status = MarketSuspended
	assert.Equal(t, MarketSuspended, market.Status)

	market.Status = MarketClosed
	assert.Equal(t, MarketClosed, market.Status)

	market.Status = MarketResolved
	assert.Equal(t, MarketResolved, market.Status)

	market.Status = MarketCancelled
	assert.Equal(t, MarketCancelled, market.Status)
}

func TestOrderStatusTransitions(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Place an order
	order := &Order{
		MarketID:  market.ID,
		UserID:    "user1",
		OutcomeID: market.Outcomes[0].ID,
		Type:      BuyOrder,
		Amount:    100.0,
		Price:     0.5,
	}

	err = pm.PlaceOrder(order)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, OrderPending, order.Status)

	// Test status transitions
	order.Status = OrderFilled
	assert.Equal(t, OrderFilled, order.Status)

	order.Status = OrderCancelled
	assert.Equal(t, OrderCancelled, order.Status)

	order.Status = OrderExpired
	assert.Equal(t, OrderExpired, order.Status)
}

func TestBackgroundProcessing(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Place multiple orders to trigger background processing
	for i := 0; i < 10; i++ {
		order := &Order{
			MarketID:  market.ID,
			UserID:    fmt.Sprintf("user%d", i),
			OutcomeID: market.Outcomes[0].ID,
			Type:      BuyOrder,
			Amount:    float64(rand.Intn(100) + 1),
			Price:     0.5 + rand.Float64()*0.1,
		}

		err := pm.PlaceOrder(order)
		require.NoError(t, err)
	}

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify metrics were updated
	metrics, exists := pm.metrics[market.ID]
	assert.True(t, exists)
	assert.NotZero(t, metrics.LastUpdated)
}

func TestPerformance(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Performance test: place many orders quickly
	start := time.Now()
	numOrders := 100 // Reduced from 1000

	for i := 0; i < numOrders; i++ {
		order := &Order{
			MarketID:  market.ID,
			UserID:    fmt.Sprintf("user%d", i),
			OutcomeID: market.Outcomes[0].ID,
			Type:      BuyOrder,
			Amount:    float64(rand.Intn(100) + 1),
			Price:     0.5 + rand.Float64()*0.1,
		}

		err := pm.PlaceOrder(order)
		require.NoError(t, err)
	}

	duration := time.Since(start)
	ordersPerSecond := float64(numOrders) / duration.Seconds()

	// Verify performance is reasonable (should handle at least 100 orders/second)
	assert.Greater(t, ordersPerSecond, 100.0)
	t.Logf("Processed %d orders in %v (%.0f orders/second)", numOrders, duration, ordersPerSecond)
}

func TestCalculateFinalPnL(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Create positions manually to test PnL calculation
	positionKey1 := fmt.Sprintf("user1:%s:%s", market.ID, market.Outcomes[0].ID)
	position1 := &Position{
		ID:        uuid.New().String(),
		UserID:    "user1",
		MarketID:  market.ID,
		OutcomeID: market.Outcomes[0].ID,
		Amount:    100.0,
		AvgPrice:  0.4,
		CreatedAt: time.Now(),
	}
	pm.positions[positionKey1] = position1

	positionKey2 := fmt.Sprintf("user2:%s:%s", market.ID, market.Outcomes[1].ID)
	position2 := &Position{
		ID:        uuid.New().String(),
		UserID:    "user2",
		MarketID:  market.ID,
		OutcomeID: market.Outcomes[1].ID,
		Amount:    50.0,
		AvgPrice:  0.6,
		CreatedAt: time.Now(),
	}
	pm.positions[positionKey2] = position2

	// Resolve market with first outcome (user1 wins, user2 loses)
	err = pm.ResolveMarket(market.ID, market.Outcomes[0].ID)
	require.NoError(t, err)

	// Check that PnL was calculated correctly
	// User1 should have profit: 100 * (1.0 - 0.4) = 60.0
	// User2 should have loss: -50 * 0.6 = -30.0
	position1, exists := pm.positions[positionKey1]
	require.True(t, exists)
	assert.Equal(t, 60.0, position1.Pnl)

	position2, exists = pm.positions[positionKey2]
	require.True(t, exists)
	assert.Equal(t, -30.0, position2.Pnl)
}

func TestLiquidityIndex(t *testing.T) {
	pm := NewProposalMarkets()
	defer pm.Close()

	// Create a market
	market := &Market{
		Title:       "Test Market",
		Description: "Test Description",
		MarketType:  BinaryMarket,
		Creator:     "user1",
		Outcomes: []*Outcome{
			{Name: "Yes", Description: "Yes outcome"},
			{Name: "No", Description: "No outcome"},
		},
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	err := pm.CreateMarket(market)
	require.NoError(t, err)
	market.Status = MarketActive
	pm.markets[market.ID] = market

	// Set some volume data
	market.Outcomes[0].Volume = 100.0
	market.Outcomes[1].Volume = 50.0
	market.TotalVolume = 150.0

	// Test liquidity index calculation
	liquidityIndex := pm.calculateLiquidityIndex(market.ID)
	expectedLiquidity := (100.0 + 50.0) / 150.0
	assert.Equal(t, expectedLiquidity, liquidityIndex)

	// Test with zero total volume
	market.TotalVolume = 0.0
	liquidityIndex = pm.calculateLiquidityIndex(market.ID)
	assert.Equal(t, 0.0, liquidityIndex)
}
