package governance

import (
	"math/big"
	"testing"
	"time"
)

// Helper function to create large big.Int values from strings
func bigIntFromString(s string) *big.Int {
	v, _ := big.NewInt(0).SetString(s, 10)
	return v
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
		VotingPeriod:         7 * 24 * time.Hour,                 // 7 days
		MaxTransactionAmount: bigIntFromString("100000000000000000000"),  // 100 tokens
		DailyLimit:           bigIntFromString("1000000000000000000000"), // 1000 tokens
		MinProposalPower:     bigIntFromString("100000000000000000000"),  // 100 tokens
		MultisigAddresses:    []string{"0x1111111111111111111111111111111111111111"},
		RequiredSignatures:   1,
		EmergencyThreshold:   bigIntFromString("10000000000000000000000"), // 10000 tokens
		SnapshotInterval:     24 * time.Hour,                      // 24 hours
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
}
