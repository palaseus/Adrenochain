package orderbook

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderTypeConstants(t *testing.T) {
	t.Run("OrderType constants", func(t *testing.T) {
		assert.Equal(t, "limit", string(OrderTypeLimit))
		assert.Equal(t, "market", string(OrderTypeMarket))
		assert.Equal(t, "stop_loss", string(OrderTypeStopLoss))
		assert.Equal(t, "take_profit", string(OrderTypeTakeProfit))
	})

	t.Run("OrderSide constants", func(t *testing.T) {
		assert.Equal(t, "buy", string(OrderSideBuy))
		assert.Equal(t, "sell", string(OrderSideSell))
	})

	t.Run("OrderStatus constants", func(t *testing.T) {
		assert.Equal(t, "pending", string(OrderStatusPending))
		assert.Equal(t, "partial", string(OrderStatusPartial))
		assert.Equal(t, "filled", string(OrderStatusFilled))
		assert.Equal(t, "cancelled", string(OrderStatusCancelled))
		assert.Equal(t, "rejected", string(OrderStatusRejected))
	})

	t.Run("TimeInForce constants", func(t *testing.T) {
		assert.Equal(t, "gtc", string(TimeInForceGTC))
		assert.Equal(t, "ioc", string(TimeInForceIOC))
		assert.Equal(t, "fok", string(TimeInForceFOK))
	})
}

func TestNewOrder(t *testing.T) {
	t.Run("valid limit order", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "order1", order.ID)
		assert.Equal(t, "user1", order.UserID)
		assert.Equal(t, "BTC/USDT", order.TradingPair)
		assert.Equal(t, OrderSideBuy, order.Side)
		assert.Equal(t, OrderTypeLimit, order.Type)
		assert.Equal(t, OrderStatusPending, order.Status)
		assert.Equal(t, quantity, order.Quantity)
		assert.Equal(t, price, order.Price)
		assert.Equal(t, big.NewInt(0), order.FilledQuantity)
		assert.Equal(t, quantity, order.RemainingQuantity)
		assert.Nil(t, order.StopPrice)
		assert.Equal(t, TimeInForceGTC, order.TimeInForce)
		assert.False(t, order.CreatedAt.IsZero())
		assert.False(t, order.UpdatedAt.IsZero())
		assert.Nil(t, order.ExpiresAt)
	})

	t.Run("valid market order", func(t *testing.T) {
		quantity := big.NewInt(100)

		order, err := NewOrder(
			"order2", "user2", "ETH/USDT",
			OrderSideSell, OrderTypeMarket,
			quantity, nil,
			TimeInForceIOC, nil, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, OrderTypeMarket, order.Type)
		assert.Nil(t, order.Price)
		assert.Equal(t, TimeInForceIOC, order.TimeInForce)
	})

	t.Run("valid stop loss order", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)
		stopPrice := big.NewInt(950)

		order, err := NewOrder(
			"order3", "user3", "BTC/USDT",
			OrderSideSell, OrderTypeStopLoss,
			quantity, price,
			TimeInForceGTC, stopPrice, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, OrderTypeStopLoss, order.Type)
		assert.Equal(t, stopPrice, order.StopPrice)
	})

	t.Run("valid take profit order", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)
		stopPrice := big.NewInt(1050)

		order, err := NewOrder(
			"order4", "user4", "BTC/USDT",
			OrderSideBuy, OrderTypeTakeProfit,
			quantity, price,
			TimeInForceGTC, stopPrice, nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, OrderTypeTakeProfit, order.Type)
		assert.Equal(t, stopPrice, order.StopPrice)
	})

	t.Run("order with expiration", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)
		expiresAt := time.Now().Add(time.Hour)

		order, err := NewOrder(
			"order5", "user5", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, &expiresAt,
		)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, &expiresAt, order.ExpiresAt)
	})
}

func TestNewOrderValidationErrors(t *testing.T) {
	t.Run("invalid order ID", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidOrderID, err)
		assert.Nil(t, order)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUserID, err)
		assert.Nil(t, order)
	})

	t.Run("invalid trading pair", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTradingPair, err)
		assert.Nil(t, order)
	})

	t.Run("invalid order side", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			"invalid", OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidOrderSide, err)
		assert.Nil(t, order)
	})

	t.Run("invalid order type", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, "invalid",
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidOrderType, err)
		assert.Nil(t, order)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		price := big.NewInt(1000)

		// Test nil quantity
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			nil, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)
		assert.Nil(t, order)

		// Test zero quantity
		order, err = NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(0), price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)
		assert.Nil(t, order)

		// Test negative quantity
		order, err = NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			big.NewInt(-100), price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)
		assert.Nil(t, order)
	})

	t.Run("invalid price for limit order", func(t *testing.T) {
		quantity := big.NewInt(100)

		// Test nil price
		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, nil,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPrice, err)
		assert.Nil(t, order)

		// Test zero price
		order, err = NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, big.NewInt(0),
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPrice, err)
		assert.Nil(t, order)

		// Test negative price
		order, err = NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, big.NewInt(-1000),
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPrice, err)
		assert.Nil(t, order)
	})

	t.Run("market order with price", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeMarket,
			quantity, price,
			TimeInForceIOC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrMarketOrderWithPrice, err)
		assert.Nil(t, order)
	})

	t.Run("stop order without stop price", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideSell, OrderTypeStopLoss,
			quantity, price,
			TimeInForceGTC, nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrStopOrderWithoutStopPrice, err)
		assert.Nil(t, order)
	})

	t.Run("invalid time in force", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			"invalid", nil, nil,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTimeInForce, err)
		assert.Nil(t, order)
	})

	t.Run("expired order", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)
		expiresAt := time.Now().Add(-time.Hour) // Past time

		order, err := NewOrder(
			"order1", "user1", "BTC/USDT",
			OrderSideBuy, OrderTypeLimit,
			quantity, price,
			TimeInForceGTC, nil, &expiresAt,
		)

		assert.Error(t, err)
		assert.Equal(t, ErrExpiredOrder, err)
		assert.Nil(t, order)
	})
}

func TestOrderValidation(t *testing.T) {
	t.Run("valid order", func(t *testing.T) {
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       big.NewInt(1000),
			TimeInForce: TimeInForceGTC,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := order.Validate()
		assert.NoError(t, err)
		assert.True(t, order.IsValid())
	})

	t.Run("invalid order", func(t *testing.T) {
		order := &Order{
			ID:          "",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       big.NewInt(1000),
			TimeInForce: TimeInForceGTC,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := order.Validate()
		assert.Error(t, err)
		assert.False(t, order.IsValid())
	})
}

func TestOrderCanFill(t *testing.T) {
	t.Run("can fill pending order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.True(t, order.CanFill())
	})

	t.Run("cannot fill filled order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusFilled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(0),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.False(t, order.CanFill())
	})

	t.Run("cannot fill cancelled order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusCancelled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.False(t, order.CanFill())
	})

	t.Run("cannot fill rejected order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusRejected,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.False(t, order.CanFill())
	})

	t.Run("cannot fill expired order", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Hour) // Past time
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			ExpiresAt:         &expiresAt,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.False(t, order.CanFill())
	})

	t.Run("cannot fill order with zero remaining quantity", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(0),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		assert.False(t, order.CanFill())
	})
}

func TestOrderFill(t *testing.T) {
	t.Run("fill order successfully", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		fillQuantity := big.NewInt(50)
		fillPrice := big.NewInt(1000)

		err := order.Fill(fillQuantity, fillPrice)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(50), order.FilledQuantity)
		assert.Equal(t, big.NewInt(50), order.RemainingQuantity)
		assert.Equal(t, OrderStatusPartial, order.Status)
		assert.True(t, order.UpdatedAt.After(order.CreatedAt))
	})

	t.Run("fill order completely", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		fillQuantity := big.NewInt(100)
		fillPrice := big.NewInt(1000)

		err := order.Fill(fillQuantity, fillPrice)
		assert.NoError(t, err)
		assert.Equal(t, big.NewInt(100), order.FilledQuantity)
		assert.Equal(t, 0, order.RemainingQuantity.Cmp(big.NewInt(0)))
		assert.Equal(t, OrderStatusFilled, order.Status)
	})

	t.Run("fill order with invalid quantity", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Test nil quantity
		err := order.Fill(nil, big.NewInt(1000))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)

		// Test zero quantity
		err = order.Fill(big.NewInt(0), big.NewInt(1000))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)

		// Test negative quantity
		err = order.Fill(big.NewInt(-50), big.NewInt(1000))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)

		// Test quantity larger than remaining
		err = order.Fill(big.NewInt(150), big.NewInt(1000))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)
	})

	t.Run("fill order that cannot be filled", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusFilled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(100),
			RemainingQuantity: big.NewInt(0),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Fill(big.NewInt(50), big.NewInt(1000))
		assert.Error(t, err)
		assert.Equal(t, ErrOrderAlreadyFilled, err)
	})
}

func TestOrderCancel(t *testing.T) {
	t.Run("cancel pending order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Cancel()
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusCancelled, order.Status)
		assert.True(t, order.UpdatedAt.After(order.CreatedAt))
	})

	t.Run("cancel partial order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPartial,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(50),
			RemainingQuantity: big.NewInt(50),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Cancel()
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusCancelled, order.Status)
	})

	t.Run("cannot cancel filled order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusFilled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(100),
			RemainingQuantity: big.NewInt(0),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Cancel()
		assert.Error(t, err)
		assert.Equal(t, ErrOrderAlreadyFilled, err)
	})

	t.Run("cannot cancel already cancelled order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusCancelled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Cancel()
		assert.Error(t, err)
		assert.Equal(t, ErrOrderAlreadyCancelled, err)
	})

	t.Run("cannot cancel rejected order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusRejected,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Cancel()
		assert.Error(t, err)
		assert.Equal(t, ErrOrderAlreadyRejected, err)
	})
}

func TestOrderReject(t *testing.T) {
	t.Run("reject pending order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusPending,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			RemainingQuantity: big.NewInt(100),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Reject()
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusRejected, order.Status)
		assert.True(t, order.UpdatedAt.After(order.CreatedAt))
	})

	t.Run("cannot reject filled order", func(t *testing.T) {
		order := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeLimit,
			Status:            OrderStatusFilled,
			Quantity:          big.NewInt(100),
			Price:             big.NewInt(1000),
			FilledQuantity:    big.NewInt(100),
			RemainingQuantity: big.NewInt(0),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := order.Reject()
		assert.Error(t, err)
		assert.Equal(t, ErrOrderAlreadyFilled, err)
	})
}

func TestOrderGetFillPrice(t *testing.T) {
	t.Run("limit order fill price", func(t *testing.T) {
		price := big.NewInt(1000)
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       price,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		fillPrice := order.GetFillPrice()
		assert.Equal(t, price, fillPrice)
	})

	t.Run("market order fill price", func(t *testing.T) {
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeMarket,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		fillPrice := order.GetFillPrice()
		assert.Nil(t, fillPrice)
	})
}

func TestOrderIsExpired(t *testing.T) {
	t.Run("order without expiration", func(t *testing.T) {
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       big.NewInt(1000),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.False(t, order.IsExpired())
	})

	t.Run("order with future expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour)
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       big.NewInt(1000),
			ExpiresAt:   &expiresAt,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.False(t, order.IsExpired())
	})

	t.Run("expired order", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Hour) // Past time
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       big.NewInt(1000),
			ExpiresAt:   &expiresAt,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.True(t, order.IsExpired())
	})
}

func TestOrderGetPriority(t *testing.T) {
	t.Run("buy order priority", func(t *testing.T) {
		now := time.Now()
		price := big.NewInt(1000)
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       price,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		priority := order.GetPriority()
		expectedPriority := now.UnixNano() + price.Int64()*1000000
		assert.Equal(t, expectedPriority, priority)
	})

	t.Run("sell order priority", func(t *testing.T) {
		now := time.Now()
		price := big.NewInt(1000)
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideSell,
			Type:        OrderTypeLimit,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       price,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		priority := order.GetPriority()
		expectedPriority := now.UnixNano() - price.Int64()*1000000
		assert.Equal(t, expectedPriority, priority)
	})

	t.Run("market order priority", func(t *testing.T) {
		now := time.Now()
		order := &Order{
			ID:          "order1",
			UserID:      "user1",
			TradingPair: "BTC/USDT",
			Side:        OrderSideBuy,
			Type:        OrderTypeMarket,
			Status:      OrderStatusPending,
			Quantity:    big.NewInt(100),
			Price:       nil,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		priority := order.GetPriority()
		expectedPriority := now.UnixNano()
		assert.Equal(t, expectedPriority, priority)
	})
}

func TestOrderClone(t *testing.T) {
	t.Run("clone order", func(t *testing.T) {
		quantity := big.NewInt(100)
		price := big.NewInt(1000)
		stopPrice := big.NewInt(950)
		expiresAt := time.Now().Add(time.Hour)

		original := &Order{
			ID:                "order1",
			UserID:            "user1",
			TradingPair:       "BTC/USDT",
			Side:              OrderSideBuy,
			Type:              OrderTypeStopLoss,
			Status:            OrderStatusPending,
			Quantity:          quantity,
			Price:             price,
			FilledQuantity:    big.NewInt(0),
			RemainingQuantity: quantity,
			StopPrice:         stopPrice,
			TimeInForce:       TimeInForceGTC,
			ExpiresAt:         &expiresAt,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		clone := original.Clone()

		// Verify clone is not the same pointer
		assert.NotSame(t, original, clone)

		// Verify all fields are copied correctly
		assert.Equal(t, original.ID, clone.ID)
		assert.Equal(t, original.UserID, clone.UserID)
		assert.Equal(t, original.TradingPair, clone.TradingPair)
		assert.Equal(t, original.Side, clone.Side)
		assert.Equal(t, original.Type, clone.Type)
		assert.Equal(t, original.Status, clone.Status)
		assert.Equal(t, original.TimeInForce, clone.TimeInForce)
		assert.Equal(t, original.CreatedAt, clone.CreatedAt)
		assert.Equal(t, original.UpdatedAt, clone.UpdatedAt)
		assert.Equal(t, original.ExpiresAt, clone.ExpiresAt)

		// Verify big.Int fields are deep copied
		assert.Equal(t, original.Quantity, clone.Quantity)
		assert.Equal(t, original.Price, clone.Price)
		assert.Equal(t, original.FilledQuantity, clone.FilledQuantity)
		assert.Equal(t, original.RemainingQuantity, clone.RemainingQuantity)
		assert.Equal(t, original.StopPrice, clone.StopPrice)

		// Verify big.Int fields are not the same pointers
		assert.NotSame(t, original.Quantity, clone.Quantity)
		assert.NotSame(t, original.Price, clone.Price)
		assert.NotSame(t, original.FilledQuantity, clone.FilledQuantity)
		assert.NotSame(t, original.RemainingQuantity, clone.RemainingQuantity)
		assert.NotSame(t, original.StopPrice, clone.StopPrice)

		// Verify modifying clone doesn't affect original
		clone.Quantity.Add(clone.Quantity, big.NewInt(1))
		assert.NotEqual(t, original.Quantity, clone.Quantity)
	})
}

func TestOrderValidationError(t *testing.T) {
	t.Run("validation error formatting", func(t *testing.T) {
		err := OrderValidationError{
			Field:   "price",
			Message: "must be positive",
		}

		expected := "price: must be positive"
		assert.Equal(t, expected, err.Error())
	})
}

// Benchmark tests for performance validation
func BenchmarkOrderGetPriority(b *testing.B) {
	order := &Order{
		ID:          "order1",
		UserID:      "user1",
		TradingPair: "BTC/USDT",
		Side:        OrderSideBuy,
		Type:        OrderTypeLimit,
		Status:      OrderStatusPending,
		Quantity:    big.NewInt(100),
		Price:       big.NewInt(1000),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order.GetPriority()
	}
}

func BenchmarkOrderValidate(b *testing.B) {
	order := &Order{
		ID:          "order1",
		UserID:      "user1",
		TradingPair: "BTC/USDT",
		Side:        OrderSideBuy,
		Type:        OrderTypeLimit,
		Status:      OrderStatusPending,
		Quantity:    big.NewInt(100),
		Price:       big.NewInt(1000),
		TimeInForce: TimeInForceGTC,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order.Validate()
	}
}

func BenchmarkOrderClone(b *testing.B) {
	order := &Order{
		ID:                "order1",
		UserID:            "user1",
		TradingPair:       "BTC/USDT",
		Side:              OrderSideBuy,
		Type:              OrderTypeLimit,
		Status:            OrderStatusPending,
		Quantity:          big.NewInt(100),
		Price:             big.NewInt(1000),
		FilledQuantity:    big.NewInt(0),
		RemainingQuantity: big.NewInt(100),
		TimeInForce:       TimeInForceGTC,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order.Clone()
	}
}
