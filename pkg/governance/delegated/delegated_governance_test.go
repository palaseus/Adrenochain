package delegated

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestNewDelegatedGovernance(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	if dg == nil {
		t.Fatal("Expected non-nil DelegatedGovernance")
	}
	
	if dg.Config.MaxDelegators != 10000 {
		t.Errorf("Expected MaxDelegators to be 10000, got %d", dg.Config.MaxDelegators)
	}
	
	if dg.Config.MaxDelegates != 1000 {
		t.Errorf("Expected MaxDelegates to be 1000, got %d", dg.Config.MaxDelegates)
	}
	
	if dg.Config.MinDelegationPower.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected MinDelegationPower to be 100, got %s", dg.Config.MinDelegationPower.String())
	}
	
	if len(dg.encryptionKey) != 32 {
		t.Errorf("Expected encryption key length to be 32, got %d", len(dg.encryptionKey))
	}
}

func TestStartStop(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	err := dg.Start()
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}
	
	if !dg.running {
		t.Error("Expected DelegatedGovernance to be running after Start")
	}
	
	err = dg.Stop()
	if err != nil {
		t.Fatalf("Expected Stop to succeed, got error: %v", err)
	}
	
	if dg.running {
		t.Error("Expected DelegatedGovernance to not be running after Stop")
	}
}

func TestRegisterDelegator(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	
	if err != nil {
		t.Fatalf("Expected RegisterDelegator to succeed, got error: %v", err)
	}
	
	if delegator.Address != "alice" {
		t.Errorf("Expected Address to be 'alice', got %s", delegator.Address)
	}
	
	if delegator.TotalPower.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected TotalPower to be 1000, got %s", delegator.TotalPower.String())
	}
	
	if delegator.RetainedPower.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected RetainedPower to be 1000, got %s", delegator.RetainedPower.String())
	}
	
	if delegator.Reputation.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected Reputation to be 100, got %s", delegator.Reputation.String())
	}
	
	// Test duplicate registration
	_, err = dg.RegisterDelegator("alice", big.NewInt(2000), nil)
	if err == nil {
		t.Error("Expected error when registering duplicate delegator")
	}
}

func TestRegisterDelegatorValidation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Test empty address
	_, err := dg.RegisterDelegator("", big.NewInt(1000), nil)
	if err == nil {
		t.Error("Expected error when address is empty")
	}
	
	// Test zero power
	_, err = dg.RegisterDelegator("bob", big.NewInt(0), nil)
	if err == nil {
		t.Error("Expected error when total power is zero")
	}
	
	// Test negative power
	_, err = dg.RegisterDelegator("bob", big.NewInt(-100), nil)
	if err == nil {
		t.Error("Expected error when total power is negative")
	}
}

func TestRegisterDelegate(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	
	if err != nil {
		t.Fatalf("Expected RegisterDelegate to succeed, got error: %v", err)
	}
	
	if delegate.Address != "bob" {
		t.Errorf("Expected Address to be 'bob', got %s", delegate.Address)
	}
	
	if delegate.TotalDelegated.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected TotalDelegated to be 0, got %s", delegate.TotalDelegated.String())
	}
	
	if delegate.ActiveDelegations != 0 {
		t.Errorf("Expected ActiveDelegations to be 0, got %d", delegate.ActiveDelegations)
	}
	
	if delegate.Reputation.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected Reputation to be 100, got %s", delegate.Reputation.String())
	}
	
	// Test duplicate registration
	_, err = dg.RegisterDelegate("bob", nil)
	if err == nil {
		t.Error("Expected error when registering duplicate delegate")
	}
}

func TestRegisterDelegateValidation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Test empty address
	_, err := dg.RegisterDelegate("", nil)
	if err == nil {
		t.Error("Expected error when address is empty")
	}
}

func TestCreateDelegation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate first
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	
	if err != nil {
		t.Fatalf("Expected CreateDelegation to succeed, got error: %v", err)
	}
	
	if delegation.DelegatorID != delegator.ID {
		t.Errorf("Expected DelegatorID to match, got %s vs %s", delegation.DelegatorID, delegator.ID)
	}
	
	if delegation.DelegateID != delegate.ID {
		t.Errorf("Expected DelegateID to match, got %s vs %s", delegation.DelegateID, delegate.ID)
	}
	
	if delegation.Type != DelegationTypeFull {
		t.Errorf("Expected Type to be Full, got %v", delegation.Type)
	}
	
	if delegation.Power.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("Expected Power to be 500, got %s", delegation.Power.String())
	}
	
	if delegation.Status != DelegationStatusActive {
		t.Errorf("Expected Status to be Active, got %v", delegation.Status)
	}
	
	// Verify delegator and delegate were updated
	updatedDelegator, _ := dg.GetDelegator(delegator.Address)
	if updatedDelegator.DelegatedPower.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("Expected DelegatedPower to be 500, got %s", updatedDelegator.DelegatedPower.String())
	}
	
	updatedDelegate, _ := dg.GetDelegate(delegate.Address)
	if updatedDelegate.TotalDelegated.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("Expected TotalDelegated to be 500, got %s", updatedDelegate.TotalDelegated.String())
	}
	
	if updatedDelegate.ActiveDelegations != 1 {
		t.Errorf("Expected ActiveDelegations to be 1, got %d", updatedDelegate.ActiveDelegations)
	}
}

func TestCreateDelegationValidation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate first
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	expiresAt := time.Now().Add(time.Hour * 24)
	
	// Test delegator not found
	_, err = dg.CreateDelegation(
		"nonexistent",
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err == nil {
		t.Error("Expected error when delegator not found")
	}
	
	// Test delegate not found
	_, err = dg.CreateDelegation(
		delegator.Address,
		"nonexistent",
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err == nil {
		t.Error("Expected error when delegate not found")
	}
	
	// Test power below minimum
	_, err = dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(50), // Below minimum
		nil,
		expiresAt,
		nil,
	)
	if err == nil {
		t.Error("Expected error when power below minimum")
	}
	
	// Test power above maximum
	_, err = dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(2000000), // Above maximum
		nil,
		expiresAt,
		nil,
	)
	if err == nil {
		t.Error("Expected error when power above maximum")
	}
	
	// Test insufficient power
	_, err = dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(1500), // More than available
		nil,
		expiresAt,
		nil,
	)
	if err == nil {
		t.Error("Expected error when power exceeds available")
	}
}

func TestCastProxyVote(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Cast proxy vote
	proxyVote, err := dg.CastProxyVote(
		delegation.ID,
		"proposal_123",
		"yes",
		big.NewInt(300),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Expected CastProxyVote to succeed, got error: %v", err)
	}
	
	if proxyVote.DelegationID != delegation.ID {
		t.Errorf("Expected DelegationID to match, got %s vs %s", proxyVote.DelegationID, delegation.ID)
	}
	
	if proxyVote.ProposalID != "proposal_123" {
		t.Errorf("Expected ProposalID to be 'proposal_123', got %s", proxyVote.ProposalID)
	}
	
	if proxyVote.VoteType != "yes" {
		t.Errorf("Expected VoteType to be 'yes', got %s", proxyVote.VoteType)
	}
	
	if proxyVote.VotePower.Cmp(big.NewInt(300)) != 0 {
		t.Errorf("Expected VotePower to be 300, got %s", proxyVote.VotePower.String())
	}
	
	// Verify delegation was updated
	updatedDelegation, _ := dg.GetDelegation(delegation.ID)
	if updatedDelegation.LastUsed.Before(delegation.LastUsed) {
		t.Error("Expected LastUsed to be updated")
	}
	
	// Verify metrics were updated
	if metrics, exists := dg.Metrics[delegation.ID]; exists {
		if metrics.TotalVotes != 1 {
			t.Errorf("Expected TotalVotes to be 1, got %d", metrics.TotalVotes)
		}
	} else {
		t.Error("Expected metrics to be created")
	}
}

func TestCastProxyVoteValidation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Test delegation not found
	_, err = dg.CastProxyVote(
		"nonexistent",
		"proposal_123",
		"yes",
		big.NewInt(300),
		nil,
	)
	if err == nil {
		t.Error("Expected error when delegation not found")
	}
	
	// Test delegation not active
	dg.mu.Lock()
	delegation.Status = DelegationStatusPaused
	dg.mu.Unlock()
	
	_, err = dg.CastProxyVote(
		delegation.ID,
		"proposal_123",
		"yes",
		big.NewInt(300),
		nil,
	)
	if err == nil {
		t.Error("Expected error when delegation not active")
	}
	
	// Reset status for other tests
	dg.mu.Lock()
	delegation.Status = DelegationStatusActive
	dg.mu.Unlock()
	
	// Test vote power exceeds delegation power
	_, err = dg.CastProxyVote(
		delegation.ID,
		"proposal_123",
		"yes",
		big.NewInt(600), // More than delegation power
		nil,
	)
	if err == nil {
		t.Error("Expected error when vote power exceeds delegation power")
	}
}

func TestRevokeDelegation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Revoke delegation
	err = dg.RevokeDelegation(delegation.ID)
	if err != nil {
		t.Fatalf("Expected RevokeDelegation to succeed, got error: %v", err)
	}
	
	// Verify status was updated
	revokedDelegation, _ := dg.GetDelegation(delegation.ID)
	if revokedDelegation.Status != DelegationStatusRevoked {
		t.Errorf("Expected Status to be Revoked, got %v", revokedDelegation.Status)
	}
	
	// Verify delegator and delegate were updated
	updatedDelegator, _ := dg.GetDelegator(delegator.Address)
	if updatedDelegator.DelegatedPower.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected DelegatedPower to be 0, got %s", updatedDelegator.DelegatedPower.String())
	}
	
	updatedDelegate, _ := dg.GetDelegate(delegate.Address)
	if updatedDelegate.TotalDelegated.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected TotalDelegated to be 0, got %s", updatedDelegate.TotalDelegated.String())
	}
	
	if updatedDelegate.ActiveDelegations != 0 {
		t.Errorf("Expected ActiveDelegations to be 0, got %d", updatedDelegate.ActiveDelegations)
	}
}

func TestPauseResumeDelegation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Pause delegation
	err = dg.PauseDelegation(delegation.ID)
	if err != nil {
		t.Fatalf("Expected PauseDelegation to succeed, got error: %v", err)
	}
	
	// Verify status was updated
	pausedDelegation, _ := dg.GetDelegation(delegation.ID)
	if pausedDelegation.Status != DelegationStatusPaused {
		t.Errorf("Expected Status to be Paused, got %v", pausedDelegation.Status)
	}
	
	// Resume delegation
	err = dg.ResumeDelegation(delegation.ID)
	if err != nil {
		t.Fatalf("Expected ResumeDelegation to succeed, got error: %v", err)
	}
	
	// Verify status was updated
	resumedDelegation, _ := dg.GetDelegation(delegation.ID)
	if resumedDelegation.Status != DelegationStatusActive {
		t.Errorf("Expected Status to be Active, got %v", resumedDelegation.Status)
	}
}

func TestGetDelegator(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	// Retrieve the delegator
	retrievedDelegator, err := dg.GetDelegator(delegator.Address)
	if err != nil {
		t.Fatalf("Expected GetDelegator to succeed, got error: %v", err)
	}
	
	if retrievedDelegator.Address != delegator.Address {
		t.Errorf("Expected Address to match, got %s vs %s", retrievedDelegator.Address, delegator.Address)
	}
	
	// Test non-existent delegator
	_, err = dg.GetDelegator("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent delegator")
	}
}

func TestGetDelegate(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Retrieve the delegate
	retrievedDelegate, err := dg.GetDelegate(delegate.Address)
	if err != nil {
		t.Fatalf("Expected GetDelegate to succeed, got error: %v", err)
	}
	
	if retrievedDelegate.Address != delegate.Address {
		t.Errorf("Expected Address to match, got %s vs %s", retrievedDelegate.Address, delegate.Address)
	}
	
	// Test non-existent delegate
	_, err = dg.GetDelegate("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent delegate")
	}
}

func TestGetDelegation(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation
	expiresAt := time.Now().Add(time.Hour * 24)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Retrieve the delegation
	retrievedDelegation, err := dg.GetDelegation(delegation.ID)
	if err != nil {
		t.Fatalf("Expected GetDelegation to succeed, got error: %v", err)
	}
	
	if retrievedDelegation.ID != delegation.ID {
		t.Errorf("Expected ID to match, got %s vs %s", retrievedDelegation.ID, delegation.ID)
	}
	
	// Test non-existent delegation
	_, err = dg.GetDelegation("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent delegation")
	}
}

func TestGetDelegationsByDelegator(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegator and delegates
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate1, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate1: %v", err)
	}
	
	delegate2, err := dg.RegisterDelegate("charlie", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate2: %v", err)
	}
	
	// Create delegations
	expiresAt := time.Now().Add(time.Hour * 24)
	_, err = dg.CreateDelegation(
		delegator.Address,
		delegate1.Address,
		DelegationTypeFull,
		big.NewInt(300),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation1: %v", err)
	}
	
	_, err = dg.CreateDelegation(
		delegator.Address,
		delegate2.Address,
		DelegationTypePartial,
		big.NewInt(200),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation2: %v", err)
	}
	
	// Get delegations by delegator
	delegations, err := dg.GetDelegationsByDelegator(delegator.Address)
	if err != nil {
		t.Fatalf("Expected GetDelegationsByDelegator to succeed, got error: %v", err)
	}
	
	if len(delegations) != 2 {
		t.Errorf("Expected 2 delegations, got %d", len(delegations))
	}
}

func TestGetDelegationsByDelegate(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Register delegators and delegate
	delegator1, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator1: %v", err)
	}
	
	delegator2, err := dg.RegisterDelegator("david", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator2: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegations
	expiresAt := time.Now().Add(time.Hour * 24)
	_, err = dg.CreateDelegation(
		delegator1.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(300),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation1: %v", err)
	}
	
	_, err = dg.CreateDelegation(
		delegator2.Address,
		delegate.Address,
		DelegationTypePartial,
		big.NewInt(200),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation2: %v", err)
	}
	
	// Get delegations by delegate
	delegations, err := dg.GetDelegationsByDelegate(delegate.Address)
	if err != nil {
		t.Fatalf("Expected GetDelegationsByDelegate to succeed, got error: %v", err)
	}
	
	if len(delegations) != 2 {
		t.Errorf("Expected 2 delegations, got %d", len(delegations))
	}
}

func TestConcurrency(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Start the system
	err := dg.Start()
	if err != nil {
		t.Fatalf("Failed to start DelegatedGovernance: %v", err)
	}
	defer dg.Stop()
	
	// Test concurrent delegator registration
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_, err := dg.RegisterDelegator(
				fmt.Sprintf("delegator_%d", id),
				big.NewInt(1000),
				nil,
			)
			if err != nil {
				// Log error but continue (some might fail due to limits)
				return
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify some delegators were created
	if len(dg.Delegators) == 0 {
		t.Error("Expected some delegators to be created")
	}
}

func TestStringMethods(t *testing.T) {
	// Test DelegationType.String()
	if DelegationTypeFull.String() != "full" {
		t.Errorf("Expected 'full', got %s", DelegationTypeFull.String())
	}
	
	if DelegationTypePartial.String() != "partial" {
		t.Errorf("Expected 'partial', got %s", DelegationTypePartial.String())
	}
	
	if DelegationTypeConditional.String() != "conditional" {
		t.Errorf("Expected 'conditional', got %s", DelegationTypeConditional.String())
	}
	
	// Test DelegationStatus.String()
	if DelegationStatusActive.String() != "active" {
		t.Errorf("Expected 'active', got %s", DelegationStatusActive.String())
	}
	
	if DelegationStatusPaused.String() != "paused" {
		t.Errorf("Expected 'paused', got %s", DelegationStatusPaused.String())
	}
	
	// Test ConditionType.String()
	if ConditionTypeTimeBased.String() != "time_based" {
		t.Errorf("Expected 'time_based', got %s", ConditionTypeTimeBased.String())
	}
	
	if ConditionTypePerformanceBased.String() != "performance_based" {
		t.Errorf("Expected 'performance_based', got %s", ConditionTypePerformanceBased.String())
	}
	
	// Test ConditionOperator.String()
	if ConditionOperatorGreaterThan.String() != "greater_than" {
		t.Errorf("Expected 'greater_than', got %s", ConditionOperatorGreaterThan.String())
	}
	
	if ConditionOperatorLessThan.String() != "less_than" {
		t.Errorf("Expected 'less_than', got %s", ConditionOperatorLessThan.String())
	}
}

func TestMemorySafety(t *testing.T) {
	dg := NewDelegatedGovernance(DelegatedGovernanceConfig{})
	
	// Create a delegator
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	// Get the delegator and modify it
	retrievedDelegator, err := dg.GetDelegator(delegator.Address)
	if err != nil {
		t.Fatalf("Failed to get delegator: %v", err)
	}
	
	originalAddress := retrievedDelegator.Address
	retrievedDelegator.Address = "Modified Address"
	
	// Verify internal state wasn't affected
	storedDelegator, _ := dg.GetDelegator("alice")
	if storedDelegator.Address == "Modified Address" {
		t.Error("Expected internal state to not be affected by external modifications")
	}
	
	if storedDelegator.Address != originalAddress {
		t.Errorf("Expected stored delegator address to remain unchanged, got %s", storedDelegator.Address)
	}
}

func TestCleanupOldData(t *testing.T) {
	// Create DelegatedGovernance with short timeout
	config := DelegatedGovernanceConfig{
		DelegationTimeout:   time.Millisecond * 100,
		CleanupInterval:     time.Millisecond * 50,
	}
	dg := NewDelegatedGovernance(config)
	
	// Start the system
	err := dg.Start()
	if err != nil {
		t.Fatalf("Failed to start DelegatedGovernance: %v", err)
	}
	defer dg.Stop()
	
	// Register delegator and delegate
	delegator, err := dg.RegisterDelegator("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register delegator: %v", err)
	}
	
	delegate, err := dg.RegisterDelegate("bob", nil)
	if err != nil {
		t.Fatalf("Failed to register delegate: %v", err)
	}
	
	// Create delegation with past expiration
	expiresAt := time.Now().Add(-time.Millisecond * 50)
	delegation, err := dg.CreateDelegation(
		delegator.Address,
		delegate.Address,
		DelegationTypeFull,
		big.NewInt(500),
		nil,
		expiresAt,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create delegation: %v", err)
	}
	
	// Manually set status to expired to trigger cleanup
	dg.mu.Lock()
	delegation.Status = DelegationStatusExpired
	dg.mu.Unlock()
	
	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)
	
	// Verify delegation was cleaned up
	_, err = dg.GetDelegation(delegation.ID)
	if err == nil {
		t.Error("Expected expired delegation to be cleaned up")
	}
}
