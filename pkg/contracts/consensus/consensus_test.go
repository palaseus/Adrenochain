package consensus

import (
	"testing"
	"time"
)

func TestBlockValidatorCreation(t *testing.T) {
	// Test creating block validator
	config := BlockValidationConfig{
		EnableContractValidation: true,
		EnableStateValidation:    true,
		EnableGasValidation:      true,
		MaxValidationTime:        5 * time.Second,
	}

	validator := NewBlockValidator(config)

	if validator == nil {
		t.Fatal("BlockValidator should not be nil")
	}

	if validator.config != config {
		t.Error("Configuration should match")
	}
}

func TestGasAccountingCreation(t *testing.T) {
	// Test creating gas accounting
	gasAccounting := NewGasAccounting(1000000, 100000)

	if gasAccounting == nil {
		t.Fatal("GasAccounting should not be nil")
	}

	// Note: Fields are private, so we can't test them directly
	// In a real implementation, we would have getter methods
}

func TestValidationStatusValues(t *testing.T) {
	// Test validation status constants
	if ValidationStatusPending != 0 {
		t.Error("ValidationStatusPending should be 0")
	}

	if ValidationStatusValidating != 1 {
		t.Error("ValidationStatusValidating should be 1")
	}

	if ValidationStatusValid != 2 {
		t.Error("ValidationStatusValid should be 2")
	}

	if ValidationStatusInvalid != 3 {
		t.Error("ValidationStatusInvalid should be 3")
	}

	if ValidationStatusFailed != 4 {
		t.Error("ValidationStatusFailed should be 4")
	}
}

func TestIssueTypeValues(t *testing.T) {
	// Test issue type constants
	if IssueTypeContractExecution != 0 {
		t.Error("IssueTypeContractExecution should be 0")
	}

	if IssueTypeStateValidation != 1 {
		t.Error("IssueTypeStateValidation should be 1")
	}

	if IssueTypeGasAccounting != 2 {
		t.Error("IssueTypeGasAccounting should be 2")
	}

	if IssueTypeConsensus != 3 {
		t.Error("IssueTypeConsensus should be 3")
	}

	if IssueTypeOther != 4 {
		t.Error("IssueTypeOther should be 4")
	}
}

func TestIssueSeverityValues(t *testing.T) {
	// Test issue severity constants
	if IssueSeverityLow != 0 {
		t.Error("IssueSeverityLow should be 0")
	}

	if IssueSeverityMedium != 1 {
		t.Error("IssueSeverityMedium should be 1")
	}

	if IssueSeverityHigh != 2 {
		t.Error("IssueSeverityHigh should be 2")
	}

	if IssueSeverityCritical != 3 {
		t.Error("IssueSeverityCritical should be 3")
	}
}
