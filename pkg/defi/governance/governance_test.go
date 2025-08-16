package governance

import (
	"math/big"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// generateRandomAddress generates a random address for testing
func generateRandomAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 100)
	}
	return addr
}

// generateRandomHash generates a random hash for testing
func generateRandomHash() engine.Hash {
	hash := engine.Hash{}
	for i := 0; i < len(hash); i++ {
		hash[i] = byte(i + 200)
	}
	return hash
}

func TestNewGovernance(t *testing.T) {
	owner := generateRandomAddress()
	governanceToken := generateRandomAddress()
	minQuorum := big.NewInt(1000000)      // 1M tokens
	proposalThreshold := big.NewInt(500)   // 500 tokens
	votingPeriod := 7 * 24 * time.Hour    // 7 days
	executionDelay := 24 * time.Hour       // 1 day

	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		owner,
		governanceToken,
		minQuorum,
		proposalThreshold,
		votingPeriod,
		executionDelay,
	)

	if gov.GovernanceID != "test-governance" {
		t.Errorf("expected governance ID 'test-governance', got '%s'", gov.GovernanceID)
	}
	if gov.Name != "Test Governance" {
		t.Errorf("expected name 'Test Governance', got '%s'", gov.Name)
	}
	if gov.Symbol != "TEST" {
		t.Errorf("expected symbol 'TEST', got '%s'", gov.Symbol)
	}
	if gov.Decimals != 18 {
		t.Errorf("expected decimals 18, got %d", gov.Decimals)
	}
	if gov.Owner != owner {
		t.Errorf("expected owner %v, got %v", owner, gov.Owner)
	}
	if gov.GovernanceToken != governanceToken {
		t.Errorf("expected governance token %v, got %v", governanceToken, gov.GovernanceToken)
	}
	if gov.MinQuorum.Cmp(minQuorum) != 0 {
		t.Errorf("expected min quorum %v, got %v", minQuorum, gov.MinQuorum)
	}
	if gov.VotingPeriod != votingPeriod {
		t.Errorf("expected voting period %v, got %v", votingPeriod, gov.VotingPeriod)
	}
	if gov.ExecutionDelay != executionDelay {
		t.Errorf("expected execution delay %v, got %v", executionDelay, gov.ExecutionDelay)
	}
	if gov.ProposalThreshold.Cmp(proposalThreshold) != 0 {
		t.Errorf("expected proposal threshold %v, got %v", proposalThreshold, gov.ProposalThreshold)
	}
	if gov.Paused {
		t.Error("expected governance to not be paused")
	}
}

func TestGovernance_GetProposalInfo_NonExistent(t *testing.T) {
	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(1000000),
		big.NewInt(500),
		7*24*time.Hour,
		24*time.Hour,
	)

	// Get non-existent proposal
	proposal := gov.GetProposalInfo(999)

	if proposal != nil {
		t.Error("expected nil for non-existent proposal")
	}
}

func TestGovernance_GetUserInfo_NonExistent(t *testing.T) {
	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(1000000),
		big.NewInt(500),
		7*24*time.Hour,
		24*time.Hour,
	)

	user := generateRandomAddress()

	// Get user for non-existent user
	userInfo := gov.GetUserInfo(user)

	if userInfo != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestGovernance_GetGovernanceStats(t *testing.T) {
	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(1000000),
		big.NewInt(500),
		7*24*time.Hour,
		24*time.Hour,
	)

	// Get initial stats
	totalProposals, totalVotes, minQuorum := gov.GetGovernanceStats()
	
	if totalProposals != 0 {
		t.Errorf("expected initial total proposals 0, got %d", totalProposals)
	}
	if totalVotes != 0 {
		t.Errorf("expected initial total votes 0, got %d", totalVotes)
	}
	if minQuorum.Cmp(big.NewInt(1000000)) != 0 {
		t.Errorf("expected min quorum 1000000, got %v", minQuorum)
	}
}

func TestGovernance_PauseUnpause(t *testing.T) {
	owner := generateRandomAddress()
	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		owner,
		generateRandomAddress(),
		big.NewInt(1000000),
		big.NewInt(500),
		7*24*time.Hour,
		24*time.Hour,
	)

	// Test pause
	err := gov.Pause()
	if err != nil {
		t.Errorf("unexpected error pausing governance: %v", err)
	}
	if !gov.Paused {
		t.Error("governance should be paused")
	}

	// Test unpause
	err = gov.Unpause()
	if err != nil {
		t.Errorf("unexpected error unpausing governance: %v", err)
	}
	if gov.Paused {
		t.Error("governance should not be paused")
	}
}

func TestGovernance_Concurrency(t *testing.T) {
	gov := NewGovernance(
		"test-governance",
		"Test Governance",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(1000000),
		big.NewInt(500),
		7*24*time.Hour,
		24*time.Hour,
	)

	// Test concurrent access to governance stats
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Access governance stats concurrently
			totalProposals, totalVotes, minQuorum := gov.GetGovernanceStats()
			
			// Just verify we can access without errors
			if totalProposals < 0 {
				t.Errorf("invalid total proposals: %d", totalProposals)
			}
			if totalVotes < 0 {
				t.Errorf("invalid total votes: %d", totalVotes)
			}
			if minQuorum.Cmp(big.NewInt(0)) < 0 {
				t.Errorf("invalid min quorum: %v", minQuorum)
			}
			
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
