package engine

import (
	"fmt"
	"sync"
)

// GasMeterImpl implements the GasMeter interface
type GasMeterImpl struct {
	gasConsumed uint64
	gasLimit    uint64
	refunded    uint64
	operations  []GasOperation
	mu          sync.RWMutex
}

// GasOperation represents a gas consumption or refund operation
type GasOperation struct {
	Type      OperationType
	Amount    uint64
	Operation string
	Timestamp int64
}

// OperationType indicates the type of gas operation
type OperationType int

const (
	OpConsume OperationType = iota
	OpRefund
	OpReset
)

// NewGasMeter creates a new gas meter with the specified gas limit
func NewGasMeter(gasLimit uint64) *GasMeterImpl {
	return &GasMeterImpl{
		gasLimit:   gasLimit,
		operations: make([]GasOperation, 0),
	}
}

// ConsumeGas consumes the specified amount of gas for an operation
func (gm *GasMeterImpl) ConsumeGas(amount uint64, operation string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if amount == 0 {
		return nil
	}

	// Check if we have enough gas
	if gm.gasConsumed+amount > gm.gasLimit {
		return fmt.Errorf("%w: requested %d, available %d", ErrInsufficientGas, amount, gm.gasLimit-gm.gasConsumed)
	}

	gm.gasConsumed += amount
	gm.operations = append(gm.operations, GasOperation{
		Type:      OpConsume,
		Amount:    amount,
		Operation: operation,
		Timestamp: getCurrentTimestamp(),
	})

	return nil
}

// RefundGas refunds the specified amount of gas
func (gm *GasMeterImpl) RefundGas(amount uint64, operation string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if amount == 0 {
		return
	}

	// Calculate actual refund amount (can't refund more than consumed)
	actualRefund := amount
	if gm.gasConsumed < amount {
		actualRefund = gm.gasConsumed
	}

	gm.gasConsumed -= actualRefund
	gm.refunded += actualRefund

	gm.operations = append(gm.operations, GasOperation{
		Type:      OpRefund,
		Amount:    actualRefund,
		Operation: operation,
		Timestamp: getCurrentTimestamp(),
	})
}

// GasConsumed returns the total amount of gas consumed
func (gm *GasMeterImpl) GasConsumed() uint64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.gasConsumed
}

// GasRemaining returns the amount of gas remaining
func (gm *GasMeterImpl) GasRemaining() uint64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.gasLimit - gm.gasConsumed
}

// GasLimit returns the total gas limit
func (gm *GasMeterImpl) GasLimit() uint64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.gasLimit
}

// Reset resets the gas meter to initial state
func (gm *GasMeterImpl) Reset() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.gasConsumed = 0
	gm.refunded = 0
	gm.operations = make([]GasOperation, 0)
}

// ResetWithLimit resets the gas meter with a new gas limit
func (gm *GasMeterImpl) ResetWithLimit(gasLimit uint64) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.gasLimit = gasLimit
	gm.gasConsumed = 0
	gm.refunded = 0
	gm.operations = make([]GasOperation, 0)
}

// GetOperations returns a copy of all gas operations
func (gm *GasMeterImpl) GetOperations() []GasOperation {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	operations := make([]GasOperation, len(gm.operations))
	copy(operations, gm.operations)
	return operations
}

// GetRefunded returns the total amount of gas refunded
func (gm *GasMeterImpl) GetRefunded() uint64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.refunded
}

// IsOutOfGas checks if the gas meter is out of gas
func (gm *GasMeterImpl) IsOutOfGas() bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.gasConsumed >= gm.gasLimit
}

// GetGasUsagePercentage returns the percentage of gas used
func (gm *GasMeterImpl) GetGasUsagePercentage() float64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gm.gasLimit == 0 {
		return 0.0
	}

	return float64(gm.gasConsumed) / float64(gm.gasLimit) * 100.0
}

// String returns a string representation of the gas meter
func (gm *GasMeterImpl) String() string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return fmt.Sprintf("GasMeter{consumed: %d, limit: %d, remaining: %d, refunded: %d, usage: %.2f%%}",
		gm.gasConsumed, gm.gasLimit, gm.gasLimit-gm.gasConsumed, gm.refunded, gm.GetGasUsagePercentage())
}

// Clone creates a deep copy of the gas meter
func (gm *GasMeterImpl) Clone() *GasMeterImpl {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	clone := &GasMeterImpl{
		gasLimit:    gm.gasLimit,
		gasConsumed: gm.gasConsumed,
		refunded:    gm.refunded,
		operations:  make([]GasOperation, len(gm.operations)),
	}

	copy(clone.operations, gm.operations)
	return clone
}

// getCurrentTimestamp returns the current timestamp in nanoseconds
func getCurrentTimestamp() int64 {
	// This would typically use a more precise time source in production
	// For now, using standard time package
	return 0 // Placeholder - would use actual timestamp in real implementation
}
