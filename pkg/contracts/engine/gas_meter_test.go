package engine

import (
	"fmt"
	"math"
	"sync"
	"testing"
)

func TestNewGasMeter(t *testing.T) {
	tests := []struct {
		name      string
		gasLimit  uint64
		expected  uint64
		shouldErr bool
	}{
		{"zero gas limit", 0, 0, false},
		{"normal gas limit", 1000, 1000, false},
		{"max uint64", math.MaxUint64, math.MaxUint64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGasMeter(tt.gasLimit)

			if gm.GasLimit() != tt.expected {
				t.Errorf("expected gas limit %d, got %d", tt.expected, gm.GasLimit())
			}

			if gm.GasConsumed() != 0 {
				t.Errorf("expected initial gas consumed 0, got %d", gm.GasConsumed())
			}

			if gm.GasRemaining() != tt.expected {
				t.Errorf("expected gas remaining %d, got %d", tt.expected, gm.GasRemaining())
			}
		})
	}
}

func TestGasMeterConsumeGas(t *testing.T) {
	tests := []struct {
		name          string
		initialLimit  uint64
		consumeAmount uint64
		operation     string
		shouldErr     bool
		expectedErr   error
	}{
		{"consume within limit", 1000, 500, "test_op", false, nil},
		{"consume exact limit", 1000, 1000, "test_op", false, nil},
		{"consume zero gas", 1000, 0, "test_op", false, nil},
		{"exceed gas limit", 1000, 1001, "test_op", true, ErrInsufficientGas},
		{"consume more than available", 1000, 1500, "test_op", true, ErrInsufficientGas},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGasMeter(tt.initialLimit)

			err := gm.ConsumeGas(tt.consumeAmount, tt.operation)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err == nil || !contains(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("expected error containing '%v', got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				expectedConsumed := tt.consumeAmount
				if gm.GasConsumed() != expectedConsumed {
					t.Errorf("expected gas consumed %d, got %d", expectedConsumed, gm.GasConsumed())
				}

				expectedRemaining := tt.initialLimit - tt.consumeAmount
				if gm.GasRemaining() != expectedRemaining {
					t.Errorf("expected gas remaining %d, got %d", expectedRemaining, gm.GasRemaining())
				}
			}
		})
	}
}

func TestGasMeterRefundGas(t *testing.T) {
	tests := []struct {
		name          string
		initialLimit  uint64
		consumeAmount uint64
		refundAmount  uint64
		expectedFinal uint64
	}{
		{"refund partial", 1000, 500, 200, 300},
		{"refund all consumed", 1000, 500, 500, 0},
		{"refund more than consumed", 1000, 500, 600, 0},
		{"refund zero", 1000, 500, 0, 500},
		{"no gas consumed", 1000, 0, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGasMeter(tt.initialLimit)

			// Consume some gas first
			if tt.consumeAmount > 0 {
				err := gm.ConsumeGas(tt.consumeAmount, "consume")
				if err != nil {
					t.Fatalf("failed to consume gas: %v", err)
				}
			}

			// Refund gas
			gm.RefundGas(tt.refundAmount, "refund")

			// Check final consumed amount
			if gm.GasConsumed() != tt.expectedFinal {
				t.Errorf("expected final gas consumed %d, got %d", tt.expectedFinal, gm.GasConsumed())
			}

			// Check refunded amount
			expectedRefunded := tt.refundAmount
			if tt.refundAmount > tt.consumeAmount {
				expectedRefunded = tt.consumeAmount
			}
			if gm.GetRefunded() != expectedRefunded {
				t.Errorf("expected refunded %d, got %d", expectedRefunded, gm.GetRefunded())
			}
		})
	}
}

func TestGasMeterReset(t *testing.T) {
	gm := NewGasMeter(1000)

	// Consume some gas
	err := gm.ConsumeGas(300, "test")
	if err != nil {
		t.Fatalf("failed to consume gas: %v", err)
	}

	// Refund some gas
	gm.RefundGas(100, "refund")

	// Reset
	gm.Reset()

	// Check reset state
	if gm.GasConsumed() != 0 {
		t.Errorf("expected gas consumed 0 after reset, got %d", gm.GasConsumed())
	}

	if gm.GetRefunded() != 0 {
		t.Errorf("expected refunded 0 after reset, got %d", gm.GetRefunded())
	}

	if gm.GasRemaining() != 1000 {
		t.Errorf("expected gas remaining 1000 after reset, got %d", gm.GasRemaining())
	}
}

func TestGasMeterResetWithLimit(t *testing.T) {
	gm := NewGasMeter(1000)

	// Consume some gas
	err := gm.ConsumeGas(300, "test")
	if err != nil {
		t.Fatalf("failed to consume gas: %v", err)
	}

	// Reset with new limit
	newLimit := uint64(2000)
	gm.ResetWithLimit(newLimit)

	// Check new state
	if gm.GasLimit() != newLimit {
		t.Errorf("expected new gas limit %d, got %d", newLimit, gm.GasLimit())
	}

	if gm.GasConsumed() != 0 {
		t.Errorf("expected gas consumed 0 after reset, got %d", gm.GasConsumed())
	}

	if gm.GasRemaining() != newLimit {
		t.Errorf("expected gas remaining %d after reset, got %d", newLimit, gm.GasRemaining())
	}
}

func TestGasMeterOperations(t *testing.T) {
	gm := NewGasMeter(1000)

	// Test multiple operations
	operations := []struct {
		amount    uint64
		operation string
	}{
		{100, "op1"},
		{200, "op2"},
		{150, "op3"},
	}

	for _, op := range operations {
		err := gm.ConsumeGas(op.amount, op.operation)
		if err != nil {
			t.Fatalf("failed to consume gas for %s: %v", op.operation, err)
		}
	}

	// Check operations
	ops := gm.GetOperations()
	if len(ops) != 3 {
		t.Errorf("expected 3 operations, got %d", len(ops))
	}

	// Check operation details
	for i, op := range ops {
		if op.Type != OpConsume {
			t.Errorf("operation %d: expected type OpConsume, got %v", i, op.Type)
		}
		if op.Operation != operations[i].operation {
			t.Errorf("operation %d: expected operation %s, got %s", i, operations[i].operation, op.Operation)
		}
		if op.Amount != operations[i].amount {
			t.Errorf("operation %d: expected amount %d, got %d", i, operations[i].amount, op.Amount)
		}
	}
}

func TestGasMeterConcurrency(t *testing.T) {
	gm := NewGasMeter(10000)

	// Test concurrent gas consumption
	var wg sync.WaitGroup
	numGoroutines := 10
	gasPerOp := uint64(100)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				err := gm.ConsumeGas(gasPerOp, fmt.Sprintf("goroutine_%d_op_%d", id, j))
				if err != nil {
					t.Errorf("goroutine %d failed to consume gas: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()

	expectedConsumed := uint64(numGoroutines * 5 * int(gasPerOp))
	if gm.GasConsumed() != expectedConsumed {
		t.Errorf("expected total gas consumed %d, got %d", expectedConsumed, gm.GasConsumed())
	}
}

func TestGasMeterEdgeCases(t *testing.T) {
	t.Run("max uint64 operations", func(t *testing.T) {
		gm := NewGasMeter(math.MaxUint64)

		// Consume max gas
		err := gm.ConsumeGas(math.MaxUint64, "max_consume")
		if err != nil {
			t.Errorf("failed to consume max gas: %v", err)
		}

		if gm.GasConsumed() != math.MaxUint64 {
			t.Errorf("expected gas consumed %d, got %d", uint64(math.MaxUint64), gm.GasConsumed())
		}

		if gm.GasRemaining() != 0 {
			t.Errorf("expected gas remaining 0, got %d", gm.GasRemaining())
		}
	})

	t.Run("zero gas limit operations", func(t *testing.T) {
		gm := NewGasMeter(0)

		// Try to consume gas
		err := gm.ConsumeGas(1, "test")
		if err == nil {
			t.Error("expected error when consuming gas with zero limit")
		}

		if gm.GasConsumed() != 0 {
			t.Errorf("expected gas consumed 0, got %d", gm.GasConsumed())
		}
	})

	t.Run("refund more than consumed", func(t *testing.T) {
		gm := NewGasMeter(1000)

		// Consume some gas
		err := gm.ConsumeGas(100, "consume")
		if err != nil {
			t.Fatalf("failed to consume gas: %v", err)
		}

		// Refund more than consumed
		gm.RefundGas(200, "refund")

		// Should only refund what was consumed
		if gm.GasConsumed() != 0 {
			t.Errorf("expected gas consumed 0, got %d", gm.GasConsumed())
		}

		if gm.GetRefunded() != 100 {
			t.Errorf("expected refunded 100, got %d", gm.GetRefunded())
		}
	})
}

func TestGasMeterString(t *testing.T) {
	gm := NewGasMeter(1000)

	// Consume some gas
	err := gm.ConsumeGas(300, "test")
	if err != nil {
		t.Fatalf("failed to consume gas: %v", err)
	}

	// Refund some gas
	gm.RefundGas(100, "refund")

	str := gm.String()

	// Check that string contains expected information
	expectedSubstrings := []string{
		"GasMeter",
		"consumed: 200",
		"limit: 1000",
		"remaining: 800",
		"refunded: 100",
		"usage: 20.00%",
	}

	for _, expected := range expectedSubstrings {
		if !contains(str, expected) {
			t.Errorf("expected string to contain '%s', got: %s", expected, str)
		}
	}
}

func TestGasMeterClone(t *testing.T) {
	gm := NewGasMeter(1000)

	// Consume some gas
	err := gm.ConsumeGas(300, "test")
	if err != nil {
		t.Fatalf("failed to consume gas: %v", err)
	}

	// Refund some gas
	gm.RefundGas(100, "refund")

	// Clone
	clone := gm.Clone()

	// Check clone has same values
	if clone.GasLimit() != gm.GasLimit() {
		t.Errorf("clone gas limit mismatch: expected %d, got %d", gm.GasLimit(), clone.GasLimit())
	}

	if clone.GasConsumed() != gm.GasConsumed() {
		t.Errorf("clone gas consumed mismatch: expected %d, got %d", gm.GasConsumed(), clone.GasConsumed())
	}

	if clone.GetRefunded() != gm.GetRefunded() {
		t.Errorf("clone refunded mismatch: expected %d, got %d", gm.GetRefunded(), clone.GetRefunded())
	}

	// Modify original and ensure clone is unaffected
	gm.ConsumeGas(100, "modify")

	if clone.GasConsumed() == gm.GasConsumed() {
		t.Error("clone should be independent of original")
	}
}

func TestGasMeterIsOutOfGas(t *testing.T) {
	tests := []struct {
		name          string
		gasLimit      uint64
		consumeAmount uint64
		expected      bool
	}{
		{"no gas consumed", 1000, 0, false},
		{"some gas consumed", 1000, 500, false},
		{"exact gas consumed", 1000, 1000, true},
		{"more gas consumed", 1000, 1001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGasMeter(tt.gasLimit)

			if tt.consumeAmount > 0 {
				err := gm.ConsumeGas(tt.consumeAmount, "test")
				if err != nil {
					// If gas consumption fails, we can't test IsOutOfGas
					// This is expected behavior for invalid operations
					return
				}
			}

			if gm.IsOutOfGas() != tt.expected {
				t.Errorf("expected IsOutOfGas() to return %v, got %v", tt.expected, gm.IsOutOfGas())
			}
		})
	}
}

func TestGasMeterGetGasUsagePercentage(t *testing.T) {
	tests := []struct {
		name          string
		gasLimit      uint64
		consumeAmount uint64
		expected      float64
	}{
		{"no gas consumed", 1000, 0, 0.0},
		{"half gas consumed", 1000, 500, 50.0},
		{"all gas consumed", 1000, 1000, 100.0},
		{"zero gas limit", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := NewGasMeter(tt.gasLimit)

			if tt.consumeAmount > 0 {
				err := gm.ConsumeGas(tt.consumeAmount, "test")
				if err != nil {
					t.Fatalf("failed to consume gas: %v", err)
				}
			}

			percentage := gm.GetGasUsagePercentage()
			if percentage != tt.expected {
				t.Errorf("expected gas usage percentage %.2f, got %.2f", tt.expected, percentage)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
