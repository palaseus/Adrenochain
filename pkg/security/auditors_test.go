package security

import (
	"context"
	"testing"
	"time"
)

func TestNewConsensusAuditor(t *testing.T) {
	auditor := NewConsensusAuditor()
	if auditor == nil {
		t.Fatal("NewConsensusAuditor returned nil")
	}
	if !auditor.enabled {
		t.Error("ConsensusAuditor should be enabled by default")
	}
}

func TestConsensusAuditor_Audit(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		expectError bool
	}{
		{
			name:        "enabled auditor",
			enabled:     true,
			expectError: false,
		},
		{
			name:        "disabled auditor",
			enabled:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := &ConsensusAuditor{enabled: tt.enabled}
			ctx := context.Background()

			result, err := auditor.Audit(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
				if err != ErrConsensusAuditingNotEnabled {
					t.Errorf("expected ErrConsensusAuditingNotEnabled, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result but got nil")
				}
				if result.ID != "consensus_audit" {
					t.Errorf("expected ID 'consensus_audit', got %s", result.ID)
				}
				if result.Type != AuditTypeConsensus {
					t.Errorf("expected Type AuditTypeConsensus, got %v", result.Type)
				}
				if result.Status != AuditStatusCompleted {
					t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
				}
				if result.Score != 95.0 {
					t.Errorf("expected Score 95.0, got %f", result.Score)
				}
				if result.Timestamp.IsZero() {
					t.Error("expected non-zero timestamp")
				}
			}
		})
	}
}

func TestNewContractAuditor(t *testing.T) {
	auditor := NewContractAuditor()
	if auditor == nil {
		t.Fatal("NewContractAuditor returned nil")
	}
	if !auditor.enabled {
		t.Error("ContractAuditor should be enabled by default")
	}
}

func TestContractAuditor_Audit(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		expectError bool
	}{
		{
			name:        "enabled auditor",
			enabled:     true,
			expectError: false,
		},
		{
			name:        "disabled auditor",
			enabled:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := &ContractAuditor{enabled: tt.enabled}
			ctx := context.Background()

			result, err := auditor.Audit(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
				if err != ErrContractAuditingNotEnabled {
					t.Errorf("expected ErrContractAuditingNotEnabled, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result but got nil")
				}
				if result.ID != "contract_audit" {
					t.Errorf("expected ID 'contract_audit', got %s", result.ID)
				}
				if result.Type != AuditTypeContract {
					t.Errorf("expected Type AuditTypeContract, got %v", result.Type)
				}
				if result.Status != AuditStatusCompleted {
					t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
				}
				if result.Score != 92.0 {
					t.Errorf("expected Score 92.0, got %f", result.Score)
				}
				if result.Timestamp.IsZero() {
					t.Error("expected non-zero timestamp")
				}
			}
		})
	}
}

func TestNewNetworkAuditor(t *testing.T) {
	auditor := NewNetworkAuditor()
	if auditor == nil {
		t.Fatal("NewNetworkAuditor returned nil")
	}
	if !auditor.enabled {
		t.Error("NetworkAuditor should be enabled by default")
	}
}

func TestNetworkAuditor_Audit(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		expectError bool
	}{
		{
			name:        "enabled auditor",
			enabled:     true,
			expectError: false,
		},
		{
			name:        "disabled auditor",
			enabled:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := &NetworkAuditor{enabled: tt.enabled}
			ctx := context.Background()

			result, err := auditor.Audit(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
				if err != ErrNetworkAuditingNotEnabled {
					t.Errorf("expected ErrNetworkAuditingNotEnabled, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result but got nil")
				}
				if result.ID != "network_audit" {
					t.Errorf("expected ID 'network_audit', got %s", result.ID)
				}
				if result.Type != AuditTypeNetwork {
					t.Errorf("expected Type AuditTypeNetwork, got %v", result.Type)
				}
				if result.Status != AuditStatusCompleted {
					t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
				}
				if result.Score != 88.0 {
					t.Errorf("expected Score 88.0, got %f", result.Score)
				}
				if result.Timestamp.IsZero() {
					t.Error("expected non-zero timestamp")
				}
			}
		})
	}
}

func TestNewEconomicAuditor(t *testing.T) {
	auditor := NewEconomicAuditor()
	if auditor == nil {
		t.Fatal("NewEconomicAuditor returned nil")
	}
	if !auditor.enabled {
		t.Error("EconomicAuditor should be enabled by default")
	}
}

func TestEconomicAuditor_Audit(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		expectError bool
	}{
		{
			name:        "enabled auditor",
			enabled:     true,
			expectError: false,
		},
		{
			name:        "disabled auditor",
			enabled:     false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := &EconomicAuditor{enabled: tt.enabled}
			ctx := context.Background()

			result, err := auditor.Audit(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
				if err != ErrEconomicAuditingNotEnabled {
					t.Errorf("expected ErrEconomicAuditingNotEnabled, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result but got nil")
				}
				if result.ID != "economic_audit" {
					t.Errorf("expected ID 'economic_audit', got %s", result.ID)
				}
				if result.Type != AuditTypeEconomic {
					t.Errorf("expected Type AuditTypeEconomic, got %v", result.Type)
				}
				if result.Status != AuditStatusCompleted {
					t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
				}
				if result.Score != 90.0 {
					t.Errorf("expected Score 90.0, got %f", result.Score)
				}
				if result.Timestamp.IsZero() {
					t.Error("expected non-zero timestamp")
				}
			}
		})
	}
}

func TestSecurityError_Error(t *testing.T) {
	message := "test security error"
	err := &SecurityError{Message: message}

	if err.Error() != message {
		t.Errorf("expected error message '%s', got '%s'", message, err.Error())
	}
}

func TestAuditorsWithContext(t *testing.T) {
	// Test with different context types
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	auditors := []struct {
		name   string
		create func() interface {
			Audit(context.Context) (*AuditResult, error)
		}
	}{
		{"ConsensusAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewConsensusAuditor()
		}},
		{"ContractAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewContractAuditor()
		}},
		{"NetworkAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewNetworkAuditor()
		}},
		{"EconomicAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewEconomicAuditor()
		}},
	}

	for _, auditor := range auditors {
		t.Run(auditor.name, func(t *testing.T) {
			aud := auditor.create()
			result, err := aud.Audit(ctx)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result but got nil")
			}
			if result.Timestamp.IsZero() {
				t.Error("expected non-zero timestamp")
			}
		})
	}
}

func TestAuditorsWithTimeoutContext(t *testing.T) {
	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	auditors := []struct {
		name   string
		create func() interface {
			Audit(context.Context) (*AuditResult, error)
		}
	}{
		{"ConsensusAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewConsensusAuditor()
		}},
		{"ContractAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewContractAuditor()
		}},
		{"NetworkAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewNetworkAuditor()
		}},
		{"EconomicAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewEconomicAuditor()
		}},
	}

	for _, auditor := range auditors {
		t.Run(auditor.name, func(t *testing.T) {
			aud := auditor.create()
			result, err := aud.Audit(ctx)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result but got nil")
			}
		})
	}
}

func TestAuditorsWithCancelledContext(t *testing.T) {
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	auditors := []struct {
		name   string
		create func() interface {
			Audit(context.Context) (*AuditResult, error)
		}
	}{
		{"ConsensusAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewConsensusAuditor()
		}},
		{"ContractAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewContractAuditor()
		}},
		{"NetworkAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewNetworkAuditor()
		}},
		{"EconomicAuditor", func() interface {
			Audit(context.Context) (*AuditResult, error)
		} {
			return NewEconomicAuditor()
		}},
	}

	for _, auditor := range auditors {
		t.Run(auditor.name, func(t *testing.T) {
			aud := auditor.create()
			result, err := aud.Audit(ctx)

			// Even with cancelled context, the audit should complete since it's synchronous
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result but got nil")
			}
		})
	}
}
