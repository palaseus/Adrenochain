package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/palaseus/adrenochain/pkg/exchange/orderbook"
	"math/big"
)

// MarketDataWebSocket handles WebSocket connections for real-time market data
type MarketDataWebSocket struct {
	clients    map[*Client]bool
	broadcast  chan interface{}
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

// Client represents a WebSocket client connection
type Client struct {
	hub      *MarketDataWebSocket
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	channels []string // Subscribed channels
	mutex    sync.RWMutex
}

// NewMarketDataWebSocket creates a new WebSocket market data service
func NewMarketDataWebSocket() *MarketDataWebSocket {
	return &MarketDataWebSocket{
		clients:   make(map[*Client]bool),
		broadcast: make(chan interface{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// Start starts the WebSocket service
func (mdws *MarketDataWebSocket) Start() {
	go mdws.run()
}

// run handles the main WebSocket event loop
func (mdws *MarketDataWebSocket) run() {
	for {
		select {
		case client := <-mdws.register:
			mdws.mutex.Lock()
			mdws.clients[client] = true
			mdws.mutex.Unlock()
			log.Printf("Client connected: %s", client.userID)

		case client := <-mdws.unregister:
			mdws.mutex.Lock()
			if _, ok := mdws.clients[client]; ok {
				delete(mdws.clients, client)
				close(client.send)
			}
			mdws.mutex.Unlock()
			log.Printf("Client disconnected: %s", client.userID)

		case message := <-mdws.broadcast:
			mdws.broadcastMessage(message)
		}
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (mdws *MarketDataWebSocket) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	conn, err := mdws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:    mdws,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle subscription messages
		c.handleSubscription(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscription handles client subscription requests
func (c *Client) handleSubscription(message []byte) {
	var subscription SubscriptionMessage
	if err := json.Unmarshal(message, &subscription); err != nil {
		c.sendError("Invalid subscription message")
		return
	}

	switch subscription.Action {
	case "subscribe":
		c.subscribe(subscription.Channel, subscription.TradingPair)
	case "unsubscribe":
		c.unsubscribe(subscription.Channel, subscription.TradingPair)
	default:
		c.sendError("Unknown action")
	}
}

// subscribe subscribes a client to a channel
func (c *Client) subscribe(channel, tradingPair string) {
	channelKey := fmt.Sprintf("%s:%s", channel, tradingPair)
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if already subscribed
	for _, ch := range c.channels {
		if ch == channelKey {
			return
		}
	}

	c.channels = append(c.channels, channelKey)
	c.sendMessage(SubscriptionResponse{
		Action:      "subscribed",
		Channel:     channel,
		TradingPair: tradingPair,
		Timestamp:   time.Now(),
	})
}

// unsubscribe unsubscribes a client from a channel
func (c *Client) unsubscribe(channel, tradingPair string) {
	channelKey := fmt.Sprintf("%s:%s", channel, tradingPair)
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, ch := range c.channels {
		if ch == channelKey {
			c.channels = append(c.channels[:i], c.channels[i+1:]...)
			break
		}
	}

	c.sendMessage(SubscriptionResponse{
		Action:      "unsubscribed",
		Channel:     channel,
		TradingPair: tradingPair,
		Timestamp:   time.Now(),
	})
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	c.sendMessage(ErrorMessage{
		Error:     message,
		Timestamp: time.Now(),
	})
}

// broadcastMessage broadcasts a message to all subscribed clients
func (mdws *MarketDataWebSocket) broadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	mdws.mutex.RLock()
	defer mdws.mutex.RUnlock()

	for client := range mdws.clients {
		// Check if client is subscribed to the relevant channel
		if mdws.shouldSendToClient(client, message) {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(mdws.clients, client)
			}
		}
	}
}

// shouldSendToClient checks if a message should be sent to a specific client
func (mdws *MarketDataWebSocket) shouldSendToClient(client *Client, message interface{}) bool {
	// This is a simplified implementation
	// In a real system, you'd check the message type and client subscriptions
	return true
}

// BroadcastOrderBookUpdate broadcasts order book updates
func (mdws *MarketDataWebSocket) BroadcastOrderBookUpdate(tradingPair string, orderBook *orderbook.OrderBook) {
	update := OrderBookUpdate{
		Type:         "orderbook_update",
		TradingPair:  tradingPair,
		Timestamp:    time.Now(),
		Bids:         []OrderBookEntry{}, // Would populate from order book
		Asks:         []OrderBookEntry{}, // Would populate from order book
	}

	mdws.broadcast <- update
}

// BroadcastTradeUpdate broadcasts trade updates
func (mdws *MarketDataWebSocket) BroadcastTradeUpdate(trade *orderbook.Trade) {
	update := TradeUpdate{
		Type:        "trade_update",
		TradingPair: trade.TradingPair,
		Price:       trade.Price,
		Quantity:    trade.Quantity,
		Timestamp:   trade.Timestamp,
	}

	mdws.broadcast <- update
}

// BroadcastMarketDataUpdate broadcasts market data updates
func (mdws *MarketDataWebSocket) BroadcastMarketDataUpdate(tradingPair string, marketData *MarketDataResponse) {
	update := MarketDataUpdate{
		Type:         "market_data_update",
		TradingPair:  tradingPair,
		LastPrice:    marketData.LastPrice,
		Volume24h:    marketData.Volume24h,
		PriceChange24h: marketData.PriceChange24h,
		Timestamp:    marketData.Timestamp,
	}

	mdws.broadcast <- update
}

// Message Types

// SubscriptionMessage represents a subscription request
type SubscriptionMessage struct {
	Action      string `json:"action"`
	Channel     string `json:"channel"`
	TradingPair string `json:"trading_pair"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	Action      string    `json:"action"`
	Channel     string    `json:"channel"`
	TradingPair string    `json:"trading_pair"`
	Timestamp   time.Time `json:"timestamp"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookUpdate represents an order book update
type OrderBookUpdate struct {
	Type         string           `json:"type"`
	TradingPair  string           `json:"trading_pair"`
	Timestamp    time.Time        `json:"timestamp"`
	Bids         []OrderBookEntry `json:"bids"`
	Asks         []OrderBookEntry `json:"asks"`
}

// TradeUpdate represents a trade update
type TradeUpdate struct {
	Type        string    `json:"type"`
	TradingPair string    `json:"trading_pair"`
	Price       *big.Int  `json:"price"`
	Quantity    *big.Int  `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
}

// MarketDataUpdate represents a market data update
type MarketDataUpdate struct {
	Type           string    `json:"type"`
	TradingPair    string    `json:"trading_pair"`
	LastPrice      *big.Int  `json:"last_price"`
	Volume24h      *big.Int  `json:"volume_24h"`
	PriceChange24h *big.Int  `json:"price_change_24h"`
	Timestamp      time.Time `json:"timestamp"`
}
