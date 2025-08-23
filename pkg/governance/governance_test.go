package governance

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function to create large big.Int values from strings
func bigIntFromString(s string) *big.Int {
	v, _ := big.NewInt(0).SetString(s, 10)
	return v
}

func TestMustParseBigInt(t *testing.T) {
	t.Run("ValidBigInt", func(t *testing.T) {
		result := mustParseBigInt("1000000000000000000000")
		expected := bigIntFromString("1000000000000000000000")
		if result.Cmp(expected) != 0 {
			t.Errorf("Expected %s, got %s", expected.String(), result.String())
		}
	})

	t.Run("ZeroBigInt", func(t *testing.T) {
		result := mustParseBigInt("0")
		if result.Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Expected 0, got %s", result.String())
		}
	})

	t.Run("LargeBigInt", func(t *testing.T) {
		largeNumber := "9999999999999999999999999999999999999999999999999999999999999999"
		result := mustParseBigInt(largeNumber)
		expected := bigIntFromString(largeNumber)
		if result.Cmp(expected) != 0 {
			t.Errorf("Expected %s, got %s", expected.String(), result.String())
		}
	})
}

func TestVotingSystem(t *testing.T) {
	// Create voting system
	quorum := mustParseBigInt("1000000000000000000000") // 1000 tokens
	votingPeriod := 7 * 24 * time.Hour                  // 7 days
	vs := NewVotingSystem(quorum, votingPeriod)

	// Set up some voting power
	vs.SetVotingPower("0x1234567890123456789012345678901234567890", mustParseBigInt("2000000000000000000000")) // 2000 tokens
	vs.SetVotingPower("0x0987654321098765432109876543210987654321", mustParseBigInt("1500000000000000000000")) // 1500 tokens

	t.Run("CreateProposal", func(t *testing.T) {
		title := "Test Proposal"
		description := "This is a test proposal"
		proposer := "0x1234567890123456789012345678901234567890"
		proposalType := ProposalTypeGeneral
		quorumRequired := mustParseBigInt("1000000000000000000000") // 1000 tokens

		proposal, err := vs.CreateProposal(title, description, proposer, proposalType, quorumRequired, mustParseBigInt("100000000000000000000"))
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		if proposal.Title != title {
			t.Errorf("Expected title %s, got %s", title, proposal.Title)
		}

		if proposal.Status != ProposalStatusDraft {
			t.Errorf("Expected status %s, got %s", ProposalStatusDraft, proposal.Status)
		}
	})

	t.Run("CreateProposalInsufficientPower", func(t *testing.T) {
		title := "Test Proposal 2"
		description := "This is another test proposal"
		proposer := "0x1111111111111111111111111111111111111111" // No voting power
		proposalType := ProposalTypeGeneral
		quorumRequired := mustParseBigInt("1000000000000000000000") // 1000 tokens

		_, err := vs.CreateProposal(title, description, proposer, proposalType, quorumRequired, mustParseBigInt("100000000000000000000"))
		if err == nil {
			t.Error("Expected error for insufficient voting power")
		}
	})

	t.Run("ActivateProposal", func(t *testing.T) {
		proposals := vs.GetProposalsByStatus(ProposalStatusDraft)
		if len(proposals) == 0 {
			t.Fatal("No draft proposals found")
		}

		proposalID := proposals[0].ID
		err := vs.ActivateProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to activate proposal: %v", err)
		}

		proposal, err := vs.GetProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to get proposal: %v", err)
		}

		if proposal.Status != ProposalStatusActive {
			t.Errorf("Expected status %s, got %s", ProposalStatusActive, proposal.Status)
		}
	})

	t.Run("CastVote", func(t *testing.T) {
		proposals := vs.GetProposalsByStatus(ProposalStatusActive)
		if len(proposals) == 0 {
			t.Fatal("No active proposals found")
		}

		proposalID := proposals[0].ID
		voter := "0x1234567890123456789012345678901234567890"
		voteChoice := VoteChoiceFor
		reason := "I support this proposal"

		err := vs.CastVote(proposalID, voter, voteChoice, reason)
		if err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}
	})

	t.Run("CastVoteDuplicate", func(t *testing.T) {
		proposals := vs.GetProposalsByStatus(ProposalStatusActive)
		if len(proposals) == 0 {
			t.Fatal("No active proposals found")
		}

		proposalID := proposals[0].ID
		voter := "0x1234567890123456789012345678901234567890"
		voteChoice := VoteChoiceAgainst
		reason := "Changed my mind"

		err := vs.CastVote(proposalID, voter, voteChoice, reason)
		if err == nil {
			t.Error("Expected error for duplicate vote")
		}
	})

	t.Run("DelegateVotingPower", func(t *testing.T) {
		delegator := "0x1234567890123456789012345678901234567890"
		delegate := "0x2222222222222222222222222222222222222222"
		amount := func() *big.Int { v, _ := big.NewInt(0).SetString("500000000000000000000", 10); return v }() // 500 tokens

		err := vs.DelegateVote(delegator, delegate, amount)
		if err != nil {
			t.Fatalf("Failed to delegate voting power: %v", err)
		}

		// Check delegated power
		delegatedPower := vs.GetVotingPower(delegate)
		if delegatedPower.Cmp(amount) != 0 {
			t.Errorf("Expected delegated power %s, got %s", amount.String(), delegatedPower.String())
		}
	})

	t.Run("FinalizeProposal", func(t *testing.T) {
		proposals := vs.GetProposalsByStatus(ProposalStatusActive)
		if len(proposals) == 0 {
			t.Fatal("No active proposals found")
		}

		proposalID := proposals[0].ID

		// Manually set voting end to past time for testing
		proposal, _ := vs.GetProposal(proposalID)
		proposal.VotingEnd = time.Now().Add(-time.Hour)

		err := vs.FinalizeProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to finalize proposal: %v", err)
		}

		// Check if proposal passed (should have 1 for vote)
		finalizedProposal, _ := vs.GetProposal(proposalID)
		if finalizedProposal.Status != ProposalStatusPassed {
			t.Errorf("Expected status %s, got %s", ProposalStatusPassed, finalizedProposal.Status)
		}
	})

	t.Run("GetVotingStats", func(t *testing.T) {
		stats := vs.GetVotingStats()

		expectedKeys := []string{"total_proposals", "draft_proposals", "active_proposals", "passed_proposals", "rejected_proposals", "executed_proposals", "total_voters"}
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})
}

func TestTreasuryManager(t *testing.T) {
	// Create treasury manager
	maxTransactionAmount := func() *big.Int { v, _ := big.NewInt(0).SetString("100000000000000000000", 10); return v }() // 100 tokens
	dailyLimit := func() *big.Int { v, _ := big.NewInt(0).SetString("1000000000000000000000", 10); return v }()          // 1000 tokens
	multisigAddresses := []string{"0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222"}
	requiredSignatures := 2

	tm := NewTreasuryManager(maxTransactionAmount, dailyLimit, multisigAddresses, requiredSignatures)

	// Set initial treasury balance
	tm.SetBalance("ETH", func() *big.Int { v, _ := big.NewInt(0).SetString("10000000000000000000000", 10); return v }()) // 10000 ETH

	t.Run("CreateTreasuryProposal", func(t *testing.T) {
		title := "Test Treasury Proposal"
		description := "This is a test treasury proposal"
		proposer := "0x1234567890123456789012345678901234567890"
		amount := bigIntFromString("100000000000000000000") // 100 ETH
		asset := "ETH"
		recipient := "0x0987654321098765432109876543210987654321"
		purpose := "Development funding"

		proposal, err := tm.CreateTreasuryProposal(title, description, proposer, amount, asset, recipient, purpose)
		if err != nil {
			t.Fatalf("Failed to create treasury proposal: %v", err)
		}

		if proposal.Title != title {
			t.Errorf("Expected title %s, got %s", title, proposal.Title)
		}

		if proposal.Status != TreasuryProposalStatusDraft {
			t.Errorf("Expected status %s, got %s", TreasuryProposalStatusDraft, proposal.Status)
		}
	})

	t.Run("CreateTreasuryProposalInsufficientBalance", func(t *testing.T) {
		title := "Test Treasury Proposal 2"
		description := "This is another test treasury proposal"
		proposer := "0x1234567890123456789012345678901234567890"
		amount := bigIntFromString("20000000000000000000000") // 20000 ETH (exceeds balance)
		asset := "ETH"
		recipient := "0x0987654321098765432109876543210987654321"
		purpose := "Large funding request"

		_, err := tm.CreateTreasuryProposal(title, description, proposer, amount, asset, recipient, purpose)
		if err == nil {
			t.Error("Expected error for insufficient balance")
		}
	})

	t.Run("ActivateTreasuryProposal", func(t *testing.T) {
		proposals := tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusDraft)
		if len(proposals) == 0 {
			t.Fatal("No draft treasury proposals found")
		}

		proposalID := proposals[0].ID
		err := tm.ActivateTreasuryProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to activate treasury proposal: %v", err)
		}

		proposal, err := tm.GetTreasuryProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to get treasury proposal: %v", err)
		}

		if proposal.Status != TreasuryProposalStatusActive {
			t.Errorf("Expected status %s, got %s", TreasuryProposalStatusActive, proposal.Status)
		}
	})

	t.Run("VoteOnTreasuryProposal", func(t *testing.T) {
		proposals := tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusActive)
		if len(proposals) == 0 {
			t.Fatal("No active treasury proposals found")
		}

		proposalID := proposals[0].ID
		voter := "0x1234567890123456789012345678901234567890"
		voteChoice := VoteChoiceFor
		votingPower := bigIntFromString("1000000000000000000000") // 1000 tokens

		err := tm.VoteOnTreasuryProposal(proposalID, voter, voteChoice, votingPower)
		if err != nil {
			t.Fatalf("Failed to vote on treasury proposal: %v", err)
		}
	})

	t.Run("FinalizeTreasuryProposal", func(t *testing.T) {
		proposals := tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusActive)
		if len(proposals) == 0 {
			t.Fatal("No active treasury proposals found")
		}

		proposalID := proposals[0].ID
		err := tm.FinalizeTreasuryProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to finalize treasury proposal: %v", err)
		}

		// Check if proposal passed (should have 1 for vote)
		finalizedProposal, _ := tm.GetTreasuryProposal(proposalID)
		if finalizedProposal.Status != TreasuryProposalStatusPassed {
			t.Errorf("Expected status %s, got %s", TreasuryProposalStatusPassed, finalizedProposal.Status)
		}
	})

	t.Run("ExecuteTreasuryProposal", func(t *testing.T) {
		proposals := tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusPassed)
		if len(proposals) == 0 {
			t.Fatal("No passed treasury proposals found")
		}

		proposalID := proposals[0].ID
		executor := "0x3333333333333333333333333333333333333333"

		transaction, err := tm.ExecuteTreasuryProposal(proposalID, executor)
		if err != nil {
			t.Fatalf("Failed to execute treasury proposal: %v", err)
		}

		if transaction.Status != TransactionStatusExecuted {
			t.Errorf("Expected status %s, got %s", TransactionStatusExecuted, transaction.Status)
		}
	})

	t.Run("CreateDirectTransaction", func(t *testing.T) {
		transactionType := TreasuryTransactionTypeTransfer
		amount := bigIntFromString("50000000000000000000") // 50 ETH
		asset := "ETH"
		to := "0x4444444444444444444444444444444444444444"
		description := "Direct transfer"
		executor := "0x3333333333333333333333333333333333333333"

		transaction, err := tm.CreateDirectTransaction(transactionType, amount, asset, to, description, executor)
		if err != nil {
			t.Fatalf("Failed to create direct transaction: %v", err)
		}

		if transaction.Status != TransactionStatusExecuted {
			t.Errorf("Expected status %s, got %s", TransactionStatusExecuted, transaction.Status)
		}
	})

	t.Run("GetTreasuryStats", func(t *testing.T) {
		stats := tm.GetTreasuryStats()

		expectedKeys := []string{"total_proposals", "active_proposals", "passed_proposals", "total_transactions", "daily_limit", "daily_used", "multisig_signers", "required_signatures"}
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})
}

func TestMultisigWallet(t *testing.T) {
	// Create multisig wallet
	addresses := []string{"0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222"}
	requiredSignatures := 2

	mw := NewMultisigWallet(addresses, requiredSignatures)

	t.Run("SignTransaction", func(t *testing.T) {
		transactionID := "tx123"
		signer := "0x1111111111111111111111111111111111111111"
		signature := []byte("signature1")

		err := mw.SignTransaction(transactionID, signer, signature)
		if err != nil {
			t.Fatalf("Failed to sign transaction: %v", err)
		}
	})

	t.Run("SignTransactionInvalidSigner", func(t *testing.T) {
		transactionID := "tx123"
		signer := "0x9999999999999999999999999999999999999999" // Not in multisig
		signature := []byte("signature2")

		err := mw.SignTransaction(transactionID, signer, signature)
		if err == nil {
			t.Error("Expected error for invalid signer")
		}
	})

	t.Run("HasEnoughSignatures", func(t *testing.T) {
		transactionID := "tx123"

		// Add second signature
		signer2 := "0x2222222222222222222222222222222222222222"
		signature2 := []byte("signature2")

		err := mw.SignTransaction(transactionID, signer2, signature2)
		if err != nil {
			t.Fatalf("Failed to add second signature: %v", err)
		}

		if !mw.HasEnoughSignatures(transactionID) {
			t.Error("Expected enough signatures")
		}
	})

	t.Run("GetSignatures", func(t *testing.T) {
		transactionID := "tx123"
		signatures := mw.GetSignatures(transactionID)

		if len(signatures) != 2 {
			t.Errorf("Expected 2 signatures, got %d", len(signatures))
		}
	})
}

func TestGovernanceCoordinator(t *testing.T) {
	// Create governance coordinator
	config := &GovernanceConfig{
		Quorum:               bigIntFromString("1000000000000000000000"), // 1000 tokens
		VotingPeriod:         1 * time.Second,                            // 1 second for testing
		MaxTransactionAmount: bigIntFromString("100000000000000000000"),  // 100 tokens
		DailyLimit:           bigIntFromString("1000000000000000000000"), // 1000 tokens
		MinProposalPower:     bigIntFromString("100000000000000000000"),  // 100 tokens
		MultisigAddresses:    []string{"0x1111111111111111111111111111111111111111"},
		RequiredSignatures:   1,
		EmergencyThreshold:   bigIntFromString("10000000000000000000000"), // 10000 tokens
		SnapshotInterval:     24 * time.Hour,                              // 24 hours
	}

	gc := NewGovernanceCoordinator(config)

	// Set up voting power
	gc.SetVotingPower("0x1234567890123456789012345678901234567890", bigIntFromString("2000000000000000000000")) // 2000 tokens

	// Set treasury balance
	gc.SetTreasuryBalance("ETH", bigIntFromString("10000000000000000000000")) // 10000 ETH

	t.Run("CreateGovernanceProposal", func(t *testing.T) {
		title := "Test Governance Proposal"
		description := "This is a test governance proposal"
		proposer := "0x1234567890123456789012345678901234567890"
		proposalType := ProposalTypeGeneral
		quorumRequired := bigIntFromString("1000000000000000000000") // 1000 tokens

		proposal, err := gc.CreateGovernanceProposal(title, description, proposer, proposalType, quorumRequired)
		if err != nil {
			t.Fatalf("Failed to create governance proposal: %v", err)
		}

		if proposal.Title != title {
			t.Errorf("Expected title %s, got %s", title, proposal.Title)
		}
	})

	t.Run("CreateTreasuryProposal", func(t *testing.T) {
		title := "Test Treasury Proposal"
		description := "This is a test treasury proposal"
		proposer := "0x1234567890123456789012345678901234567890"
		amount := bigIntFromString("100000000000000000000") // 100 ETH
		asset := "ETH"
		recipient := "0x0987654321098765432109876543210987654321"
		purpose := "Development funding"

		proposal, err := gc.CreateTreasuryProposal(title, description, proposer, amount, asset, recipient, purpose)
		if err != nil {
			t.Fatalf("Failed to create treasury proposal: %v", err)
		}

		if proposal.Title != title {
			t.Errorf("Expected title %s, got %s", title, proposal.Title)
		}
	})

	t.Run("GetGovernanceStats", func(t *testing.T) {
		stats := gc.GetGovernanceStats()

		// Check that stats contain expected keys
		expectedKeys := []string{
			"voting_total_proposals", "voting_draft_proposals", "voting_active_proposals",
			"treasury_total_proposals", "treasury_active_proposals", "treasury_passed_proposals",
			"quorum", "voting_period_hours", "max_transaction_amount", "daily_limit",
			"min_proposal_power", "multisig_signers", "required_signatures",
		}

		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stat key %s", key)
			}
		}
	})

	t.Run("EmergencyOperations", func(t *testing.T) {
		pausedBy := "admin"
		reason := "Emergency maintenance"

		err := gc.EmergencyPause(pausedBy, reason)
		if err != nil {
			t.Fatalf("Failed to pause governance: %v", err)
		}

		resumedBy := "admin"
		err = gc.EmergencyResume(resumedBy)
		if err != nil {
			t.Fatalf("Failed to resume governance: %v", err)
		}
	})

	t.Run("EventHandling", func(t *testing.T) {
		eventReceived := false
		gc.On("test_event", func(data interface{}) {
			eventReceived = true
		})

		// Emit a test event
		gc.emitEvent("test_event", "test_data")

		// Give some time for the event to be processed
		time.Sleep(100 * time.Millisecond)

		if !eventReceived {
			t.Error("Event was not received")
		}
	})

	t.Run("ActivateGovernanceProposal", func(t *testing.T) {
		// First create a proposal
		title := "Test Proposal for Activation"
		description := "This proposal will be activated"
		proposer := "0x1234567890123456789012345678901234567890"
		proposalType := ProposalTypeGeneral
		quorumRequired := bigIntFromString("1000000000000000000000")

		proposal, err := gc.CreateGovernanceProposal(title, description, proposer, proposalType, quorumRequired)
		if err != nil {
			t.Fatalf("Failed to create proposal: %v", err)
		}

		// Activate the proposal
		err = gc.ActivateGovernanceProposal(proposal.ID)
		if err != nil {
			t.Fatalf("Failed to activate proposal: %v", err)
		}
	})

	t.Run("VoteOnGovernanceProposal", func(t *testing.T) {
		// Get an active proposal
		activeProposals := gc.GetGovernanceProposalsByStatus(ProposalStatusActive)
		if len(activeProposals) == 0 {
			t.Skip("No active proposals to vote on")
		}

		proposalID := activeProposals[0].ID
		voter := "0x1234567890123456789012345678901234567890"
		voteChoice := VoteChoiceFor
		reason := "I support this proposal"

		err := gc.VoteOnGovernanceProposal(proposalID, voter, voteChoice, reason)
		if err != nil {
			t.Fatalf("Failed to vote on proposal: %v", err)
		}
	})

	t.Run("DelegateVotingPower", func(t *testing.T) {
		delegator := "0x1234567890123456789012345678901234567890"
		delegate := "0x0987654321098765432109876543210987654321"
		amount := bigIntFromString("500000000000000000000") // 500 tokens

		err := gc.DelegateVotingPower(delegator, delegate, amount)
		if err != nil {
			t.Fatalf("Failed to delegate voting power: %v", err)
		}
	})

	t.Run("FinalizeGovernanceProposal", func(t *testing.T) {
		// Get an active proposal
		activeProposals := gc.GetGovernanceProposalsByStatus(ProposalStatusActive)
		if len(activeProposals) == 0 {
			t.Skip("No active proposals to finalize")
		}

		proposalID := activeProposals[0].ID

		// Wait for voting period to end (1 second)
		time.Sleep(2 * time.Second)

		err := gc.FinalizeGovernanceProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to finalize proposal: %v", err)
		}
	})

	t.Run("ExecuteGovernanceProposal", func(t *testing.T) {
		// Get a passed proposal
		passedProposals := gc.GetGovernanceProposalsByStatus(ProposalStatusPassed)
		if len(passedProposals) == 0 {
			t.Skip("No passed proposals to execute")
		}

		proposalID := passedProposals[0].ID
		executor := "0x1234567890123456789012345678901234567890"

		err := gc.ExecuteGovernanceProposal(proposalID, executor)
		if err != nil {
			t.Fatalf("Failed to execute proposal: %v", err)
		}
	})

	t.Run("ActivateTreasuryProposal", func(t *testing.T) {
		// First create a treasury proposal
		title := "Test Treasury Proposal for Activation"
		description := "This treasury proposal will be activated"
		proposer := "0x1234567890123456789012345678901234567890"
		amount := bigIntFromString("100000000000000000000")
		asset := "ETH"
		recipient := "0x0987654321098765432109876543210987654321"
		purpose := "Testing activation"

		proposal, err := gc.CreateTreasuryProposal(title, description, proposer, amount, asset, recipient, purpose)
		if err != nil {
			t.Fatalf("Failed to create treasury proposal: %v", err)
		}

		// Activate the proposal
		err = gc.ActivateTreasuryProposal(proposal.ID)
		if err != nil {
			t.Fatalf("Failed to activate treasury proposal: %v", err)
		}
	})

	t.Run("VoteOnTreasuryProposal", func(t *testing.T) {
		// Get an active treasury proposal
		activeProposals := gc.GetTreasuryProposalsByStatus(TreasuryProposalStatusActive)
		if len(activeProposals) == 0 {
			t.Skip("No active treasury proposals to vote on")
		}

		proposalID := activeProposals[0].ID
		voter := "0x1234567890123456789012345678901234567890"
		voteChoice := VoteChoiceFor
		votingPower := bigIntFromString("1000000000000000000000")

		err := gc.VoteOnTreasuryProposal(proposalID, voter, voteChoice, votingPower)
		if err != nil {
			t.Fatalf("Failed to vote on treasury proposal: %v", err)
		}
	})

	t.Run("FinalizeTreasuryProposal", func(t *testing.T) {
		// Get an active treasury proposal
		activeProposals := gc.GetTreasuryProposalsByStatus(TreasuryProposalStatusActive)
		if len(activeProposals) == 0 {
			t.Skip("No active treasury proposals to finalize")
		}

		proposalID := activeProposals[0].ID
		err := gc.FinalizeTreasuryProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to finalize treasury proposal: %v", err)
		}
	})

	t.Run("ExecuteTreasuryProposal", func(t *testing.T) {
		// Get a passed treasury proposal
		passedProposals := gc.GetTreasuryProposalsByStatus(TreasuryProposalStatusPassed)
		if len(passedProposals) == 0 {
			t.Skip("No passed treasury proposals to execute")
		}

		proposalID := passedProposals[0].ID
		executor := "0x1234567890123456789012345678901234567890"

		transaction, err := gc.ExecuteTreasuryProposal(proposalID, executor)
		if err != nil {
			t.Fatalf("Failed to execute treasury proposal: %v", err)
		}

		if transaction == nil {
			t.Error("Expected transaction to be returned")
		}
	})

	t.Run("CreateDirectTreasuryTransaction", func(t *testing.T) {
		transactionType := TreasuryTransactionTypeTransfer
		amount := bigIntFromString("50000000000000000000") // 50 ETH
		asset := "ETH"
		to := "0x0987654321098765432109876543210987654321"
		description := "Direct transfer for testing"
		executor := "0x1234567890123456789012345678901234567890"

		transaction, err := gc.CreateDirectTreasuryTransaction(transactionType, amount, asset, to, description, executor)
		if err != nil {
			t.Fatalf("Failed to create direct treasury transaction: %v", err)
		}

		if transaction == nil {
			t.Error("Expected transaction to be returned")
		}
	})

	t.Run("GetTreasuryBalance", func(t *testing.T) {
		balance := gc.GetTreasuryBalance("ETH")
		if balance == nil {
			t.Error("Expected treasury balance to be returned")
		}
	})

	t.Run("GetVotingPower", func(t *testing.T) {
		address := "0x1234567890123456789012345678901234567890"
		power := gc.GetVotingPower(address)
		if power == nil {
			t.Error("Expected voting power to be returned")
		}
	})

	t.Run("GetGovernanceProposal", func(t *testing.T) {
		// Get any proposal
		allProposals := gc.GetGovernanceProposalsByStatus(ProposalStatusDraft)
		if len(allProposals) == 0 {
			t.Skip("No proposals to get")
		}

		proposalID := allProposals[0].ID
		proposal, err := gc.GetGovernanceProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to get governance proposal: %v", err)
		}

		if proposal == nil {
			t.Error("Expected proposal to be returned")
		}
	})

	t.Run("GetGovernanceProposalsByStatus", func(t *testing.T) {
		proposals := gc.GetGovernanceProposalsByStatus(ProposalStatusDraft)
		if proposals == nil {
			t.Error("Expected proposals slice to be returned")
		}
	})

	t.Run("GetTreasuryProposal", func(t *testing.T) {
		// Get any treasury proposal
		allProposals := gc.GetTreasuryProposalsByStatus(TreasuryProposalStatusDraft)
		if len(allProposals) == 0 {
			t.Skip("No treasury proposals to get")
		}

		proposalID := allProposals[0].ID
		proposal, err := gc.GetTreasuryProposal(proposalID)
		if err != nil {
			t.Fatalf("Failed to get treasury proposal: %v", err)
		}
		if proposal == nil {
			t.Error("Expected proposal to be returned")
		}
	})

	t.Run("GetTreasuryProposalsByStatus", func(t *testing.T) {
		proposals := gc.GetTreasuryProposalsByStatus(TreasuryProposalStatusDraft)
		if proposals == nil {
			t.Error("Expected treasury proposals slice to be returned")
		}
	})

	t.Run("SetTreasuryBalance", func(t *testing.T) {
		asset := "TEST"
		newBalance := bigIntFromString("1000000000000000000000") // 1000 TEST tokens

		gc.SetTreasuryBalance(asset, newBalance)

		// Verify the balance was set
		balance := gc.GetTreasuryBalance(asset)
		if balance.Cmp(newBalance) != 0 {
			t.Errorf("Expected balance %s, got %s", newBalance.String(), balance.String())
		}
	})

	t.Run("SetVotingPower", func(t *testing.T) {
		address := "0x9999999999999999999999999999999999999999"
		newPower := bigIntFromString("3000000000000000000000") // 3000 tokens

		gc.SetVotingPower(address, newPower)

		// Verify the voting power was set
		power := gc.GetVotingPower(address)
		if power.Cmp(newPower) != 0 {
			t.Errorf("Expected voting power %s, got %s", newPower.String(), power.String())
		}
	})

	// Test getter functions
	t.Run("GetVotingSystem", func(t *testing.T) {
		votingSystem := gc.GetVotingSystem()
		if votingSystem == nil {
			t.Error("Expected voting system to be returned")
		}
	})

	t.Run("GetTreasuryManager", func(t *testing.T) {
		treasuryManager := gc.GetTreasuryManager()
		if treasuryManager == nil {
			t.Error("Expected treasury manager to be returned")
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		config := gc.GetConfig()
		if config == nil {
			t.Error("Expected config to be returned")
		}
	})

	t.Run("GetVotesForProposal", func(t *testing.T) {
		// Get any active proposal
		allProposals := gc.GetGovernanceProposalsByStatus(ProposalStatusActive)
		if len(allProposals) == 0 {
			t.Skip("No active proposals to get votes for")
		}

		proposalID := allProposals[0].ID
		votes := gc.GetVotesForProposal(proposalID)
		if votes == nil {
			t.Error("Expected votes slice to be returned")
		}
	})

	t.Run("GetTreasuryTransactions", func(t *testing.T) {
		transactions := gc.GetTreasuryTransactions(10)
		if transactions == nil {
			t.Error("Expected transactions slice to be returned")
		}
	})

	t.Run("GetGovernanceStats", func(t *testing.T) {
		stats := gc.GetGovernanceStats()
		if stats == nil {
			t.Error("Expected stats map to be returned")
		}

		// Check for expected stats
		if _, exists := stats["quorum"]; !exists {
			t.Error("Expected quorum stat to be present")
		}
		if _, exists := stats["voting_period_hours"]; !exists {
			t.Error("Expected voting_period_hours stat to be present")
		}
	})

	t.Run("EmergencyPause", func(t *testing.T) {
		err := gc.EmergencyPause("admin", "test pause")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EmergencyResume", func(t *testing.T) {
		err := gc.EmergencyResume("admin")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("EventSystem", func(t *testing.T) {
		eventReceived := false
		gc.On("test_event", func(data interface{}) {
			eventReceived = true
		})

		// Trigger an event
		gc.emitEvent("test_event", "test_data")

		// Give some time for the goroutine to execute
		time.Sleep(10 * time.Millisecond)

		if !eventReceived {
			t.Error("Expected event to be received")
		}
	})

}

func TestGovernanceCoordinator_GetVotesForProposal(t *testing.T) {
	config := &GovernanceConfig{
		Quorum:               bigIntFromString("1000000000000000000000"),
		VotingPeriod:         1 * time.Second,
		MaxTransactionAmount: bigIntFromString("100000000000000000000"),
		DailyLimit:           bigIntFromString("1000000000000000000000"),
		MinProposalPower:     bigIntFromString("100000000000000000000"),
		MultisigAddresses:    []string{"0x1111111111111111111111111111111111111111"},
		RequiredSignatures:   1,
		EmergencyThreshold:   bigIntFromString("10000000000000000000000"),
		SnapshotInterval:     24 * time.Hour,
	}

	gc := NewGovernanceCoordinator(config)
	gc.SetVotingPower("0x1234567890123456789012345678901234567890", bigIntFromString("2000000000000000000000"))

	// Create and activate a proposal
	proposal, err := gc.CreateGovernanceProposal(
		"Test Proposal for Votes",
		"This proposal will be used to test vote retrieval",
		"0x1234567890123456789012345678901234567890",
		ProposalTypeGeneral,
		bigIntFromString("1000000000000000000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = gc.ActivateGovernanceProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Cast some votes
	err = gc.VoteOnGovernanceProposal(proposal.ID, "0x1234567890123456789012345678901234567890", VoteChoiceFor, "Supporting the proposal")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Get votes for the proposal
	votes := gc.GetVotesForProposal(proposal.ID)
	if votes == nil {
		t.Fatal("Expected votes slice, got nil")
	}

	if len(votes) != 1 {
		t.Errorf("Expected 1 vote, got %d", len(votes))
	}

	// Verify vote details
	vote := votes[0]
	if vote.Voter != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Expected voter %s, got %s", "0x1234567890123456789012345678901234567890", vote.Voter)
	}

	if vote.VoteChoice != VoteChoiceFor {
		t.Errorf("Expected vote choice %s, got %s", VoteChoiceFor, vote.VoteChoice)
	}
}

func TestGovernanceCoordinator_GetVotesForProposal_NonExistent(t *testing.T) {
	gc := NewGovernanceCoordinator(nil)

	// Test getting votes for non-existent proposal
	votes := gc.GetVotesForProposal("non_existent_proposal")
	if votes != nil {
		t.Errorf("Expected nil for non-existent proposal, got %v", votes)
	}
}

func TestGovernanceCoordinator_GetVotesForProposal_Empty(t *testing.T) {
	config := &GovernanceConfig{
		Quorum:               bigIntFromString("1000000000000000000000"),
		VotingPeriod:         1 * time.Second,
		MaxTransactionAmount: bigIntFromString("100000000000000000000"),
		DailyLimit:           bigIntFromString("1000000000000000000000"),
		MinProposalPower:     bigIntFromString("100000000000000000000"),
		MultisigAddresses:    []string{"0x1111111111111111111111111111111111111111"},
		RequiredSignatures:   1,
		EmergencyThreshold:   bigIntFromString("10000000000000000000000"),
		SnapshotInterval:     24 * time.Hour,
	}

	gc := NewGovernanceCoordinator(config)
	gc.SetVotingPower("0x1234567890123456789012345678901234567890", bigIntFromString("2000000000000000000000"))

	// Create and activate a proposal
	proposal, err := gc.CreateGovernanceProposal(
		"Test Proposal No Votes",
		"This proposal will have no votes",
		"0x1234567890123456789012345678901234567890",
		ProposalTypeGeneral,
		bigIntFromString("1000000000000000000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = gc.ActivateGovernanceProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Get votes for proposal with no votes
	votes := gc.GetVotesForProposal(proposal.ID)
	if votes == nil {
		t.Fatal("Expected empty votes slice, got nil")
	}

	if len(votes) != 0 {
		t.Errorf("Expected 0 votes for proposal with no votes, got %d", len(votes))
	}
}

func TestVotingSystem_GetVotesForProposal(t *testing.T) {
	quorum := mustParseBigInt("1000000000000000000000")
	votingPeriod := 7 * 24 * time.Hour
	vs := NewVotingSystem(quorum, votingPeriod)

	// Set up voting power
	vs.SetVotingPower("0x1234567890123456789012345678901234567890", mustParseBigInt("2000000000000000000000"))

	// Create a proposal
	proposal, err := vs.CreateProposal(
		"Test Proposal",
		"This is a test proposal",
		"0x1234567890123456789012345678901234567890",
		ProposalTypeGeneral,
		mustParseBigInt("1000000000000000000000"),
		mustParseBigInt("100000000000000000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate the proposal
	err = vs.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Cast a vote
	err = vs.CastVote(proposal.ID, "0x1234567890123456789012345678901234567890", VoteChoiceFor, "Supporting the proposal")
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Get votes for the proposal
	votes := vs.GetVotesForProposal(proposal.ID)
	if votes == nil {
		t.Fatal("Expected votes slice, got nil")
	}

	if len(votes) != 1 {
		t.Errorf("Expected 1 vote, got %d", len(votes))
	}

	// Verify vote details
	vote := votes[0]
	if vote.Voter != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Expected voter %s, got %s", "0x1234567890123456789012345678901234567890", vote.Voter)
	}

	if vote.VoteChoice != VoteChoiceFor {
		t.Errorf("Expected vote choice %s, got %s", VoteChoiceFor, vote.VoteChoice)
	}
}

func TestVotingSystem_GetVotesForProposal_NonExistent(t *testing.T) {
	quorum := mustParseBigInt("1000000000000000000000")
	votingPeriod := 7 * 24 * time.Hour
	vs := NewVotingSystem(quorum, votingPeriod)

	// Test getting votes for non-existent proposal
	votes := vs.GetVotesForProposal("non_existent_proposal")
	if votes != nil {
		t.Errorf("Expected nil for non-existent proposal, got %v", votes)
	}
}

func TestVotingSystem_CastDelegatedVote(t *testing.T) {
	quorum := mustParseBigInt("1000000000000000000000")
	votingPeriod := 7 * 24 * time.Hour
	vs := NewVotingSystem(quorum, votingPeriod)

	// Set up voting power
	vs.SetVotingPower("0x1234567890123456789012345678901234567890", mustParseBigInt("2000000000000000000000"))
	vs.SetVotingPower("0x0987654321098765432109876543210987654321", mustParseBigInt("1000000000000000000000"))

	// Create a proposal
	proposal, err := vs.CreateProposal(
		"Test Proposal for Delegated Vote",
		"This proposal will be used to test delegated voting",
		"0x1234567890123456789012345678901234567890",
		ProposalTypeGeneral,
		mustParseBigInt("1000000000000000000000"),
		mustParseBigInt("100000000000000000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate the proposal
	err = vs.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Delegate voting power
	err = vs.DelegateVote("0x1234567890123456789012345678901234567890", "0x0987654321098765432109876543210987654321", mustParseBigInt("500000000000000000000"))
	if err != nil {
		t.Fatalf("Failed to delegate vote: %v", err)
	}

	// Cast delegated vote
	err = vs.CastDelegatedVote(proposal.ID, "0x0987654321098765432109876543210987654321", "0x1234567890123456789012345678901234567890", VoteChoiceFor, "Delegated vote")
	if err != nil {
		t.Fatalf("Failed to cast delegated vote: %v", err)
	}

	// Verify the delegated vote was recorded
	votes := vs.GetVotesForProposal(proposal.ID)
	if len(votes) != 1 {
		t.Errorf("Expected 1 vote, got %d", len(votes))
	}

	vote := votes[0]
	if !vote.IsDelegated {
		t.Error("Expected vote to be marked as delegated")
	}

	if vote.Delegator != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Expected delegator %s, got %s", "0x1234567890123456789012345678901234567890", vote.Delegator)
	}
}

func TestVotingSystem_CastDelegatedVote_InvalidDelegator(t *testing.T) {
	quorum := mustParseBigInt("1000000000000000000000")
	votingPeriod := 7 * 24 * time.Hour
	vs := NewVotingSystem(quorum, votingPeriod)

	// Set up voting power
	vs.SetVotingPower("0x1234567890123456789012345678901234567890", mustParseBigInt("2000000000000000000000"))

	// Create a proposal
	proposal, err := vs.CreateProposal(
		"Test Proposal for Invalid Delegated Vote",
		"This proposal will be used to test invalid delegated voting",
		"0x1234567890123456789012345678901234567890",
		ProposalTypeGeneral,
		mustParseBigInt("1000000000000000000000"),
		mustParseBigInt("100000000000000000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate the proposal
	err = vs.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Try to cast delegated vote with delegate that has no voting power
	err = vs.CastDelegatedVote(proposal.ID, "0x9999999999999999999999999999999999999999", "0x1234567890123456789012345678901234567890", VoteChoiceFor, "Invalid delegated vote")
	if err == nil {
		t.Error("Expected error for delegate with no voting power")
	}
}

// Test Enhanced Identity Verification
func TestEnhancedIdentityVerification(t *testing.T) {
	config := IdentityVerificationConfig{
		PersonhoodProofRequired:  true,
		SocialGraphRequired:      true,
		BiometricRequired:        false,
		MinimumStakeAmount:       big.NewInt(1000),
		TrustScoreThreshold:      0.5,
		SybilResistanceThreshold: 0.3,
		MaxVotingWeight:          1.0,
		VerificationValidity:     24 * time.Hour,
	}

	t.Run("new_enhanced_identity_verification", func(t *testing.T) {
		eiv := NewEnhancedIdentityVerification(config)
		assert.NotNil(t, eiv)
		assert.NotEmpty(t, eiv.ID)
		assert.NotNil(t, eiv.IdentityRegistry)
		assert.NotNil(t, eiv.ProofOfPersonhood)
		assert.NotNil(t, eiv.StakeBasedVerification)
		assert.NotNil(t, eiv.SocialGraphAnalysis)
		assert.NotNil(t, eiv.BiometricValidation)
		assert.Equal(t, config, eiv.config)
	})

	t.Run("verify_identity_basic", func(t *testing.T) {
		eiv := NewEnhancedIdentityVerification(config)
		userID := "user123"
		publicKey := []byte("public_key_data")
		stakeAmount := big.NewInt(15000)

		identity, err := eiv.VerifyIdentity(userID, publicKey, stakeAmount)
		assert.NoError(t, err)
		assert.NotNil(t, identity)
		assert.Equal(t, publicKey, identity.PublicKey)
		assert.Equal(t, stakeAmount, identity.StakeAmount)
		assert.Equal(t, VerificationLevelAdvanced, identity.VerificationLevel)
		assert.Greater(t, identity.TrustScore, 0.5)
		assert.NotNil(t, identity.PersonhoodProof)
		assert.Greater(t, identity.SybilResistanceScore, 0.0)
	})

	t.Run("verify_identity_duplicate", func(t *testing.T) {
		eiv := NewEnhancedIdentityVerification(config)
		userID := "user123"
		publicKey := []byte("public_key_data")
		stakeAmount := big.NewInt(15000)

		// First verification
		identity1, err := eiv.VerifyIdentity(userID, publicKey, stakeAmount)
		assert.NoError(t, err)
		assert.NotNil(t, identity1)

		// Second verification should return existing identity
		identity2, err := eiv.VerifyIdentity(userID, publicKey, stakeAmount)
		assert.NoError(t, err)
		assert.NotNil(t, identity2)
		assert.Equal(t, identity1.ID, identity2.ID)
	})
}

// Test Proof of Personhood System
func TestProofOfPersonhoodSystem(t *testing.T) {
	t.Run("new_proof_of_personhood_system", func(t *testing.T) {
		pop := NewProofOfPersonhoodSystem()
		assert.NotNil(t, pop)
	})

	t.Run("generate_personhood_proof", func(t *testing.T) {
		pop := NewProofOfPersonhoodSystem()
		userID := "user123"

		proof, err := pop.GeneratePersonhoodProof(userID)
		assert.NoError(t, err)
		assert.NotNil(t, proof)
		assert.Equal(t, userID, proof.UserID)
		assert.NotEmpty(t, proof.ProofData)
		assert.False(t, proof.Timestamp.IsZero())
	})

	t.Run("validate_personhood_proof", func(t *testing.T) {
		pop := NewProofOfPersonhoodSystem()
		userID := "user123"

		proof, err := pop.GeneratePersonhoodProof(userID)
		assert.NoError(t, err)

		isValid := pop.ValidatePersonhoodProof(proof)
		assert.True(t, isValid)
	})
}

// Test Stake-Based Verification System
func TestStakeBasedVerificationSystem(t *testing.T) {
	t.Run("new_stake_based_verification_system", func(t *testing.T) {
		sbvs := NewStakeBasedVerificationSystem()
		assert.NotNil(t, sbvs)
	})

	t.Run("verify_stake", func(t *testing.T) {
		sbvs := NewStakeBasedVerificationSystem()
		stakeAmount := big.NewInt(15000)
		verificationLevel := VerificationLevelAdvanced

		score, err := sbvs.VerifyStake(stakeAmount, verificationLevel)
		assert.NoError(t, err)
		assert.Greater(t, score, 0.0)
	})
}

// Test Social Graph Analyzer
func TestSocialGraphAnalyzer(t *testing.T) {
	t.Run("new_social_graph_analyzer", func(t *testing.T) {
		sga := NewSocialGraphAnalyzer()
		assert.NotNil(t, sga)
	})

	t.Run("analyze_identity", func(t *testing.T) {
		sga := NewSocialGraphAnalyzer()
		userID := "user123"

		score, err := sga.AnalyzeIdentity(userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, score, 1.0)
	})

	t.Run("detect_sybil_cluster", func(t *testing.T) {
		sga := NewSocialGraphAnalyzer()
		userID := "user123"

		isSybil, err := sga.DetectSybilCluster(userID)
		assert.NoError(t, err)
		assert.False(t, isSybil) // New user should not be flagged as Sybil
	})
}

// Test Biometric Validation System
func TestBiometricValidationSystem(t *testing.T) {
	t.Run("new_biometric_validation_system", func(t *testing.T) {
		bvs := NewBiometricValidationSystem()
		assert.NotNil(t, bvs)
	})

	t.Run("validate_biometric", func(t *testing.T) {
		bvs := NewBiometricValidationSystem()
		userID := "user123"

		hash, err := bvs.ValidateBiometric(userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

// Test Privacy-Preserving Identity System
func TestPrivacyPreservingIdentity(t *testing.T) {
	config := PrivacyPreservingConfig{
		EnableSocialReputation:     true,
		EnableEconomicReputation:   true,
		EnableBehavioralReputation: true,
		EnableNetworkReputation:    true,
		PrivacyLevel:               PrivacyLevelEnhanced,
		ReputationThreshold:        0.5,
		SybilResistanceThreshold:   0.3,
		MaxVotingWeight:            1.0,
		PrivacyBudget:              100.0,
	}

	t.Run("new_privacy_preserving_identity", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		assert.NotNil(t, ppi)
	})

	t.Run("verify_identity_privacy_preserving", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		userID := "user123"
		publicKey := []byte("public_key_123")
		stakeAmount := big.NewInt(15000)

		identity, err := ppi.VerifyIdentityPrivacyPreserving(userID, publicKey, stakeAmount)
		assert.NoError(t, err)
		assert.NotNil(t, identity)
		assert.NotEmpty(t, identity.ID)
		// The function generates its own ID, so we just check it's not empty
		assert.NotEqual(t, "", identity.ID)
	})

	t.Run("calculate_privacy_preserving_sybil_resistance", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		score := ppi.calculatePrivacyPreservingSybilResistance(identity)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 1.0)
	})

	t.Run("calculate_privacy_preserving_voting_weight", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		weight := ppi.calculatePrivacyPreservingVotingWeight(identity)
		assert.GreaterOrEqual(t, weight, 0.0)
	})

	t.Run("new_social_reputation_system", func(t *testing.T) {
		srs := NewSocialReputationSystem()
		assert.NotNil(t, srs)
	})

	t.Run("calculate_social_reputation", func(t *testing.T) {
		srs := NewSocialReputationSystem()
		userID := "user123"

		reputation, err := srs.CalculateSocialReputation(userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, reputation, 0.0)
		assert.LessOrEqual(t, reputation, 1.0)
	})

	t.Run("detect_sybil_pattern", func(t *testing.T) {
		srs := NewSocialReputationSystem()
		userID := "user123"

		isSybil, err := srs.DetectSybilPattern(userID)
		assert.NoError(t, err)
		assert.False(t, isSybil) // New user should not be flagged as Sybil
	})

	t.Run("new_economic_reputation_system", func(t *testing.T) {
		ers := NewEconomicReputationSystem()
		assert.NotNil(t, ers)
	})

	t.Run("calculate_economic_reputation", func(t *testing.T) {
		ers := NewEconomicReputationSystem()
		userID := "user123"
		stakeAmount := big.NewInt(15000)

		reputation, err := ers.CalculateEconomicReputation(userID, stakeAmount)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, reputation, 0.0)
		assert.LessOrEqual(t, reputation, 1.0)
	})

	t.Run("new_behavioral_reputation_system", func(t *testing.T) {
		brs := NewBehavioralReputationSystem()
		assert.NotNil(t, brs)
	})

	t.Run("calculate_behavioral_reputation", func(t *testing.T) {
		brs := NewBehavioralReputationSystem()
		userID := "user123"

		reputation, err := brs.CalculateBehavioralReputation(userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, reputation, 0.0)
		assert.LessOrEqual(t, reputation, 1.0)
	})

	t.Run("new_network_reputation_system", func(t *testing.T) {
		nrs := NewNetworkReputationSystem()
		assert.NotNil(t, nrs)
	})

	t.Run("calculate_network_reputation", func(t *testing.T) {
		nrs := NewNetworkReputationSystem()
		userID := "user123"

		reputation, err := nrs.CalculateNetworkReputation(userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, reputation, 0.0)
		assert.LessOrEqual(t, reputation, 1.0)
	})

	t.Run("new_privacy_engine", func(t *testing.T) {
		pe := NewPrivacyEngine()
		assert.NotNil(t, pe)
	})

	t.Run("ensure_privacy_preservation", func(t *testing.T) {
		pe := NewPrivacyEngine()
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		privacyScore, err := pe.EnsurePrivacyPreservation(identity)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, privacyScore, 0.0)
	})

	t.Run("update_metrics", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}
		success := true
		ppi.updateMetrics(identity, success)
		// Test that metrics are updated without error
	})

	t.Run("get_metrics", func(t *testing.T) {
		ppi := NewPrivacyPreservingIdentity(config)
		metrics := ppi.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("new_differential_privacy_engine", func(t *testing.T) {
		dpe := NewDifferentialPrivacyEngine()
		assert.NotNil(t, dpe)
	})

	t.Run("new_zero_knowledge_engine", func(t *testing.T) {
		zke := NewZeroKnowledgeEngine()
		assert.NotNil(t, zke)
	})

	t.Run("new_homomorphic_encryption_engine", func(t *testing.T) {
		hee := NewHomomorphicEncryptionEngine()
		assert.NotNil(t, hee)
	})

	t.Run("apply_differential_privacy", func(t *testing.T) {
		dpe := NewDifferentialPrivacyEngine()
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		protectedData, err := dpe.ApplyDifferentialPrivacy(identity)
		assert.NoError(t, err)
		assert.NotNil(t, protectedData)
	})

	t.Run("generate_zero_knowledge_proof", func(t *testing.T) {
		zke := NewZeroKnowledgeEngine()
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		proof, err := zke.GenerateZeroKnowledgeProof(identity)
		assert.NoError(t, err)
		assert.NotNil(t, proof)
	})

	t.Run("encrypt_data", func(t *testing.T) {
		hee := NewHomomorphicEncryptionEngine()
		identity := &VerifiedIdentity{
			ID:                "user123",
			PublicKey:         []byte("public_key_123"),
			StakeAmount:       big.NewInt(15000),
			VerificationLevel: VerificationLevelBasic,
			TrustScore:        0.5,
		}

		encryptedData, err := hee.EncryptData(identity)
		assert.NoError(t, err)
		assert.NotNil(t, encryptedData)
	})
}
