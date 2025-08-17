package advancedorders

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/logger"
)

// OrderType represents different types of advanced orders
type OrderType string

const (
	OrderTypeConditional  OrderType = "conditional"
	OrderTypeTimeWeighted OrderType = "time_weighted"
	OrderTypeIceberg      OrderType = "iceberg"
	OrderTypeTWAP         OrderType = "twap"
	OrderTypeVWAP         OrderType = "vwap"
	OrderTypeStopLoss     OrderType = "stop_loss"
	OrderTypeTakeProfit   OrderType = "take_profit"
	OrderTypeTrailingStop OrderType = "trailing_stop"
)

// OrderStatus represents the status of an advanced order
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusActive          OrderStatus = "active"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusRejected        OrderStatus = "rejected"
	OrderStatusExpired         OrderStatus = "expired"
)

// AdvancedOrder represents an advanced order with complex execution logic
type AdvancedOrder struct {
	ID                string                 `json:"id"`
	Type              OrderType              `json:"type"`
	Symbol            string                 `json:"symbol"`
	Side              OrderSide              `json:"side"`
	Quantity          float64                `json:"quantity"`
	Price             float64                `json:"price"`
	Status            OrderStatus            `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	FilledQuantity    float64                `json:"filled_quantity"`
	RemainingQuantity float64                `json:"remaining_quantity"`
	AveragePrice      float64                `json:"average_price"`
	Metadata          map[string]interface{} `json:"metadata"`
	Conditions        []OrderCondition       `json:"conditions"`
	ExecutionPlan     ExecutionPlan          `json:"execution_plan"`
	mu                sync.RWMutex
}

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderCondition represents a condition that must be met for order execution
type OrderCondition interface {
	Evaluate(marketData MarketData) bool
	GetType() string
	GetDescription() string
}

// ExecutionPlan defines how the order should be executed
type ExecutionPlan interface {
	CalculateExecution(remainingQuantity float64, marketData MarketData) ExecutionInstruction
	GetType() string
	GetDescription() string
}

// ExecutionInstruction contains execution details
type ExecutionInstruction struct {
	Price     float64                `json:"price"`
	Quantity  float64                `json:"quantity"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MarketData contains market information for condition evaluation
type MarketData struct {
	Symbol     string    `json:"symbol"`
	Price      float64   `json:"price"`
	Volume     float64   `json:"volume"`
	Timestamp  time.Time `json:"timestamp"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Spread     float64   `json:"spread"`
	Volatility float64   `json:"volatility"`
	Trend      float64   `json:"trend"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Open       float64   `json:"open"`
}

// NewAdvancedOrder creates a new advanced order
func NewAdvancedOrder(id string, orderType OrderType, symbol string, side OrderSide, quantity, price float64) *AdvancedOrder {
	return &AdvancedOrder{
		ID:                id,
		Type:              orderType,
		Symbol:            symbol,
		Side:              side,
		Quantity:          quantity,
		Price:             price,
		Status:            OrderStatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		FilledQuantity:    0,
		RemainingQuantity: quantity,
		AveragePrice:      0,
		Metadata:          make(map[string]interface{}),
		Conditions:        make([]OrderCondition, 0),
	}
}

// AddCondition adds a condition to the order
func (order *AdvancedOrder) AddCondition(condition OrderCondition) {
	order.mu.Lock()
	defer order.mu.Unlock()
	order.Conditions = append(order.Conditions, condition)
}

// SetExecutionPlan sets the execution plan for the order
func (order *AdvancedOrder) SetExecutionPlan(plan ExecutionPlan) {
	order.mu.Lock()
	defer order.mu.Unlock()
	order.ExecutionPlan = plan
}

// EvaluateConditions evaluates all conditions for the order
func (order *AdvancedOrder) EvaluateConditions(marketData MarketData) bool {
	order.mu.RLock()
	defer order.mu.RUnlock()

	if len(order.Conditions) == 0 {
		return true // No conditions means always execute
	}

	for _, condition := range order.Conditions {
		if !condition.Evaluate(marketData) {
			return false
		}
	}
	return true
}

// UpdateStatus updates the order status
func (order *AdvancedOrder) UpdateStatus(status OrderStatus) {
	order.mu.Lock()
	defer order.mu.Unlock()
	order.Status = status
	order.UpdatedAt = time.Now()
}

// Fill updates the order with a fill
func (order *AdvancedOrder) Fill(quantity, price float64) {
	order.mu.Lock()
	defer order.mu.Unlock()

	if order.Status == OrderStatusFilled || order.Status == OrderStatusCancelled {
		return
	}

	// Calculate new average price
	totalValue := order.FilledQuantity*order.AveragePrice + quantity*price
	order.FilledQuantity += quantity
	order.AveragePrice = totalValue / order.FilledQuantity

	// Update remaining quantity
	order.RemainingQuantity = order.Quantity - order.FilledQuantity

	// Update status
	if order.RemainingQuantity <= 0 {
		order.Status = OrderStatusFilled
	} else {
		order.Status = OrderStatusPartiallyFilled
	}

	order.UpdatedAt = time.Now()
}

// Cancel cancels the order
func (order *AdvancedOrder) Cancel() {
	order.mu.Lock()
	defer order.mu.Unlock()

	if order.Status == OrderStatusFilled {
		return
	}

	order.Status = OrderStatusCancelled
	order.UpdatedAt = time.Now()
}

// IsActive checks if the order is active
func (order *AdvancedOrder) IsActive() bool {
	order.mu.RLock()
	defer order.mu.RUnlock()

	return order.Status == OrderStatusPending ||
		order.Status == OrderStatusActive ||
		order.Status == OrderStatusPartiallyFilled
}

// IsExpired checks if the order has expired
func (order *AdvancedOrder) IsExpired() bool {
	order.mu.RLock()
	defer order.mu.RUnlock()

	if order.ExpiresAt == nil {
		return false
	}

	return time.Now().After(*order.ExpiresAt)
}

// GetProgress returns the progress of order execution
func (order *AdvancedOrder) GetProgress() float64 {
	order.mu.RLock()
	defer order.mu.RUnlock()

	if order.Quantity <= 0 {
		return 0
	}

	return order.FilledQuantity / order.Quantity
}

// GetUnrealizedPnL calculates unrealized PnL for the order
func (order *AdvancedOrder) GetUnrealizedPnL(currentPrice float64) float64 {
	order.mu.RLock()
	defer order.mu.RUnlock()

	if order.Side == OrderSideBuy {
		return (currentPrice - order.AveragePrice) * order.FilledQuantity
	} else {
		return (order.AveragePrice - currentPrice) * order.FilledQuantity
	}
}

// Conditional Order Implementation

// PriceCondition represents a price-based condition
type PriceCondition struct {
	TargetPrice float64   `json:"target_price"`
	Operator    string    `json:"operator"` // "above", "below", "equals"
	Side        OrderSide `json:"side"`
}

func (pc *PriceCondition) Evaluate(marketData MarketData) bool {
	switch pc.Operator {
	case "above":
		if pc.Side == OrderSideBuy {
			return marketData.Price >= pc.TargetPrice
		} else {
			return marketData.Price <= pc.TargetPrice
		}
	case "below":
		if pc.Side == OrderSideBuy {
			return marketData.Price <= pc.TargetPrice
		} else {
			return marketData.Price >= pc.TargetPrice
		}
	case "equals":
		return math.Abs(marketData.Price-pc.TargetPrice) < 0.01
	default:
		return false
	}
}

func (pc *PriceCondition) GetType() string {
	return "price"
}

func (pc *PriceCondition) GetDescription() string {
	return fmt.Sprintf("Price %s %s %.2f", pc.Operator, pc.Side, pc.TargetPrice)
}

// VolumeCondition represents a volume-based condition
type VolumeCondition struct {
	MinVolume float64 `json:"min_volume"`
	MaxVolume float64 `json:"max_volume"`
}

func (vc *VolumeCondition) Evaluate(marketData MarketData) bool {
	return marketData.Volume >= vc.MinVolume && marketData.Volume <= vc.MaxVolume
}

func (vc *VolumeCondition) GetType() string {
	return "volume"
}

func (vc *VolumeCondition) GetDescription() string {
	return fmt.Sprintf("Volume between %.2f and %.2f", vc.MinVolume, vc.MaxVolume)
}

// TimeCondition represents a time-based condition
type TimeCondition struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

func (tc *TimeCondition) Evaluate(marketData MarketData) bool {
	now := marketData.Timestamp

	if tc.StartTime != nil && now.Before(*tc.StartTime) {
		return false
	}

	if tc.EndTime != nil && now.After(*tc.EndTime) {
		return false
	}

	return true
}

func (tc *TimeCondition) GetType() string {
	return "time"
}

func (tc *TimeCondition) GetDescription() string {
	if tc.StartTime != nil && tc.EndTime != nil {
		return fmt.Sprintf("Time between %s and %s", tc.StartTime.Format("15:04:05"), tc.EndTime.Format("15:04:05"))
	} else if tc.StartTime != nil {
		return fmt.Sprintf("After %s", tc.StartTime.Format("15:04:05"))
	} else if tc.EndTime != nil {
		return fmt.Sprintf("Before %s", tc.EndTime.Format("15:04:05"))
	}
	return "No time restrictions"
}

// VolatilityCondition represents a volatility-based condition
type VolatilityCondition struct {
	MinVolatility float64 `json:"min_volatility"`
	MaxVolatility float64 `json:"max_volatility"`
}

func (vc *VolatilityCondition) Evaluate(marketData MarketData) bool {
	return marketData.Volatility >= vc.MinVolatility && marketData.Volatility <= vc.MaxVolatility
}

func (vc *VolatilityCondition) GetType() string {
	return "volatility"
}

func (vc *VolatilityCondition) GetDescription() string {
	return fmt.Sprintf("Volatility between %.4f and %.4f", vc.MinVolatility, vc.MaxVolatility)
}

// Execution Plan Implementations

// MarketOrderExecution executes orders at market price
type MarketOrderExecution struct{}

func (moe *MarketOrderExecution) CalculateExecution(remainingQuantity float64, marketData MarketData) ExecutionInstruction {
	price := marketData.Price
	if remainingQuantity > 0 {
		price = marketData.Ask // Use ask price for buy orders
	} else {
		price = marketData.Bid // Use bid price for sell orders
	}

	return ExecutionInstruction{
		Price:     price,
		Quantity:  math.Abs(remainingQuantity),
		Timestamp: marketData.Timestamp,
		Metadata: map[string]interface{}{
			"execution_type": "market",
		},
	}
}

func (moe *MarketOrderExecution) GetType() string {
	return "market"
}

func (moe *MarketOrderExecution) GetDescription() string {
	return "Execute at market price"
}

// LimitOrderExecution executes orders at a specific limit price
type LimitOrderExecution struct {
	LimitPrice float64 `json:"limit_price"`
}

func (loe *LimitOrderExecution) CalculateExecution(remainingQuantity float64, marketData MarketData) ExecutionInstruction {
	// Check if limit price is favorable
	executable := false
	if remainingQuantity > 0 { // Buy order
		executable = marketData.Ask <= loe.LimitPrice
	} else { // Sell order
		executable = marketData.Bid >= loe.LimitPrice
	}

	if !executable {
		return ExecutionInstruction{
			Price:     0,
			Quantity:  0,
			Timestamp: marketData.Timestamp,
			Metadata: map[string]interface{}{
				"execution_type": "limit",
				"executable":     false,
			},
		}
	}

	return ExecutionInstruction{
		Price:     loe.LimitPrice,
		Quantity:  math.Abs(remainingQuantity),
		Timestamp: marketData.Timestamp,
		Metadata: map[string]interface{}{
			"execution_type": "limit",
			"executable":     true,
		},
	}
}

func (loe *LimitOrderExecution) GetType() string {
	return "limit"
}

func (loe *LimitOrderExecution) GetDescription() string {
	return fmt.Sprintf("Execute at limit price %.2f", loe.LimitPrice)
}

// OrderManager manages advanced orders
type OrderManager struct {
	orders map[string]*AdvancedOrder
	logger *logger.Logger
	mu     sync.RWMutex
}

// NewOrderManager creates a new order manager
func NewOrderManager() *OrderManager {
	return &OrderManager{
		orders: make(map[string]*AdvancedOrder),
		logger: logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "advanced_order_manager"}),
	}
}

// CreateOrder creates a new advanced order
func (om *OrderManager) CreateOrder(orderType OrderType, symbol string, side OrderSide, quantity, price float64) *AdvancedOrder {
	om.mu.Lock()
	defer om.mu.Unlock()

	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	order := NewAdvancedOrder(orderID, orderType, symbol, side, quantity, price)

	om.orders[orderID] = order
	om.logger.Info("Advanced order created - order_id: %s, type: %s, symbol: %s", orderID, orderType, symbol)

	return order
}

// GetOrder retrieves an order by ID
func (om *OrderManager) GetOrder(orderID string) (*AdvancedOrder, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	order, exists := om.orders[orderID]
	return order, exists
}

// CancelOrder cancels an order
func (om *OrderManager) CancelOrder(orderID string) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, exists := om.orders[orderID]
	if !exists {
		return fmt.Errorf("order %s not found", orderID)
	}

	order.Cancel()
	om.logger.Info("Order cancelled - order_id: %s", orderID)
	return nil
}

// GetActiveOrders returns all active orders
func (om *OrderManager) GetActiveOrders() []*AdvancedOrder {
	om.mu.RLock()
	defer om.mu.RUnlock()

	activeOrders := make([]*AdvancedOrder, 0)
	for _, order := range om.orders {
		if order.IsActive() {
			activeOrders = append(activeOrders, order)
		}
	}

	return activeOrders
}

// ProcessMarketData processes market data and evaluates order conditions
func (om *OrderManager) ProcessMarketData(marketData MarketData) {
	om.mu.Lock()
	defer om.mu.Unlock()

	for _, order := range om.orders {
		if !order.IsActive() || order.IsExpired() {
			continue
		}

		// Evaluate conditions
		if order.EvaluateConditions(marketData) {
			// Execute order if conditions are met
			if order.ExecutionPlan != nil {
				instruction := order.ExecutionPlan.CalculateExecution(order.RemainingQuantity, marketData)
				if instruction.Quantity > 0 {
					om.logger.Info("Order execution triggered - order_id: %s, instruction: %+v", order.ID, instruction)
					// Here you would typically send the instruction to the execution engine
				}
			}
		}
	}
}

// CleanupExpiredOrders removes expired orders
func (om *OrderManager) CleanupExpiredOrders() {
	om.mu.Lock()
	defer om.mu.Unlock()

	expiredOrders := make([]string, 0)
	for orderID, order := range om.orders {
		if order.IsExpired() {
			expiredOrders = append(expiredOrders, orderID)
		}
	}

	for _, orderID := range expiredOrders {
		delete(om.orders, orderID)
		om.logger.Info("Expired order removed - order_id: %s", orderID)
	}
}
