package api

import (
	"math/big"
	"testing"
)

func TestNewTradingAPI(t *testing.T) {
	api := NewTradingAPI()
	if api == nil {
		t.Fatal("Failed to create TradingAPI")
	}

	if api.orderBookManager == nil {
		t.Error("OrderBookManager is nil")
	}

	if api.tradingManager == nil {
		t.Error("TradingManager is nil")
	}

	if api.marketData == nil {
		t.Error("MarketData is nil")
	}
}

func TestNewMarketDataWebSocket(t *testing.T) {
	ws := NewMarketDataWebSocket()
	if ws == nil {
		t.Fatal("Failed to create MarketDataWebSocket")
	}

	if ws.clients == nil {
		t.Error("Clients map is nil")
	}

	if ws.broadcast == nil {
		t.Error("Broadcast channel is nil")
	}

	if ws.register == nil {
		t.Error("Register channel is nil")
	}

	if ws.unregister == nil {
		t.Error("Unregister channel is nil")
	}
}

func TestCreateOrderRequestValidation(t *testing.T) {
	// Test valid request
	validRequest := &CreateOrderRequest{
		TradingPair: "BTC/USDT",
		Side:        "buy",
		Type:        "limit",
		Quantity:    big.NewInt(1000000), // 0.001 BTC
		Price:       big.NewInt(50000000000), // $50,000
		UserID:      "user123",
	}

	if err := validRequest.Validate(); err != nil {
		t.Errorf("Valid request failed validation: %v", err)
	}

	// Test invalid request (missing trading pair)
	invalidRequest := &CreateOrderRequest{
		Side:     "buy",
		Type:     "limit",
		Quantity: big.NewInt(1000000),
		Price:    big.NewInt(50000000000),
		UserID:   "user123",
	}

	if err := invalidRequest.Validate(); err == nil {
		t.Error("Invalid request should have failed validation")
	}
}
