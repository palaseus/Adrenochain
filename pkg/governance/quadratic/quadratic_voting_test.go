package quadratic

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestNewQuadraticVoting(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	if qv == nil {
		t.Fatal("Expected non-nil QuadraticVoting")
	}

	if qv.Config.MaxProposals != 1000 {
		t.Errorf("Expected MaxProposals to be 1000, got %d", qv.Config.MaxProposals)
	}

	if qv.Config.MaxVoters != 10000 {
		t.Errorf("Expected MaxVoters to be 10000, got %d", qv.Config.MaxVoters)
	}

	if qv.Config.MinQuorum.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected MinQuorum to be 1000, got %s", qv.Config.MinQuorum.String())
	}

	if len(qv.encryptionKey) != 32 {
		t.Errorf("Expected encryption key length to be 32, got %d", len(qv.encryptionKey))
	}
}

func TestStartStop(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	err := qv.Start()
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}

	if !qv.running {
		t.Error("Expected QuadraticVoting to be running after Start")
	}

	err = qv.Stop()
	if err != nil {
		t.Fatalf("Expected Stop to succeed, got error: %v", err)
	}

	if qv.running {
		t.Error("Expected QuadraticVoting to not be running after Stop")
	}
}

func TestCreateProposal(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	votingStart := time.Now().Add(time.Hour)
	votingEnd := time.Now().Add(time.Hour * 24)
	quorum := big.NewInt(2000)
	threshold := big.NewInt(10000)

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		quorum,
		threshold,
		nil,
	)

	if err != nil {
		t.Fatalf("Expected CreateProposal to succeed, got error: %v", err)
	}

	if proposal.Title != "Test Proposal" {
		t.Errorf("Expected Title to be 'Test Proposal', got %s", proposal.Title)
	}

	if proposal.Status != ProposalStatusDraft {
		t.Errorf("Expected Status to be Draft, got %v", proposal.Status)
	}

	if proposal.Quorum.Cmp(quorum) != 0 {
		t.Errorf("Expected Quorum to match, got %s vs %s", proposal.Quorum.String(), quorum.String())
	}
}

func TestCreateProposalValidation(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	votingStart := time.Now().Add(time.Hour)
	votingEnd := time.Now().Add(time.Hour * 24)

	// Test empty title
	_, err := qv.CreateProposal(
		"",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err == nil {
		t.Error("Expected error when title is empty")
	}

	// Test empty description
	_, err = qv.CreateProposal(
		"Test Proposal",
		"",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err == nil {
		t.Error("Expected error when description is empty")
	}

	// Test invalid voting times
	_, err = qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingEnd, // Start after end
		votingStart,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err == nil {
		t.Error("Expected error when voting start is after end")
	}

	// Test quorum below minimum
	_, err = qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(500), // Below minimum
		big.NewInt(10000),
		nil,
	)

	if err == nil {
		t.Error("Expected error when quorum is below minimum")
	}
}

func TestActivateProposal(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	votingStart := time.Now().Add(-time.Hour) // Start in the past
	votingEnd := time.Now().Add(time.Hour * 24)

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate proposal
	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Expected ActivateProposal to succeed, got error: %v", err)
	}

	// Verify status was updated
	activatedProposal, err := qv.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	if activatedProposal.Status != ProposalStatusActive {
		t.Errorf("Expected Status to be Active, got %v", activatedProposal.Status)
	}
}

func TestStartVoting(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	votingStart := time.Now().Add(-time.Hour) // Start in the past
	votingEnd := time.Now().Add(time.Hour * 24)

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Activate proposal first
	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	// Start voting
	err = qv.StartVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Expected StartVoting to succeed, got error: %v", err)
	}

	// Verify status was updated
	votingProposal, err := qv.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	if votingProposal.Status != ProposalStatusVoting {
		t.Errorf("Expected Status to be Voting, got %v", votingProposal.Status)
	}
}

func TestRegisterVoter(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	voter, err := qv.RegisterVoter("alice", big.NewInt(1000), nil)

	if err != nil {
		t.Fatalf("Expected RegisterVoter to succeed, got error: %v", err)
	}

	if voter.Address != "alice" {
		t.Errorf("Expected Address to be 'alice', got %s", voter.Address)
	}

	if voter.VotingPower.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected VotingPower to be 1000, got %s", voter.VotingPower.String())
	}

	if voter.Reputation.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected Reputation to be 100, got %s", voter.Reputation.String())
	}

	// Test duplicate registration
	_, err = qv.RegisterVoter("alice", big.NewInt(2000), nil)
	if err == nil {
		t.Error("Expected error when registering duplicate voter")
	}
}

func TestCastVote(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Create proposal and start voting
	votingStart := time.Now().Add(-time.Hour)
	votingEnd := time.Now().Add(time.Hour * 24)

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	err = qv.StartVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to start voting: %v", err)
	}

	// Register voter
	voter, err := qv.RegisterVoter("bob", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register voter: %v", err)
	}

	// Cast vote
	vote, err := qv.CastVote(proposal.ID, voter.Address, VoteTypeYes, big.NewInt(100), nil)

	if err != nil {
		t.Fatalf("Expected CastVote to succeed, got error: %v", err)
	}

	if vote.VoteType != VoteTypeYes {
		t.Errorf("Expected VoteType to be Yes, got %v", vote.VoteType)
	}

	if vote.VotePower.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected VotePower to be 100, got %s", vote.VotePower.String())
	}

	// Verify vote cost (quadratic formula: power^2)
	expectedCost := big.NewInt(10000) // 100^2
	if vote.VoteCost.Cmp(expectedCost) != 0 {
		t.Errorf("Expected VoteCost to be %s, got %s", expectedCost.String(), vote.VoteCost.String())
	}

	// Test duplicate voting
	_, err = qv.CastVote(proposal.ID, voter.Address, VoteTypeNo, big.NewInt(50), nil)
	if err == nil {
		t.Error("Expected error when voting twice on same proposal")
	}
}

func TestCloseVoting(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Create proposal and start voting
	votingStart := time.Now().Add(-time.Hour)
	votingEnd := time.Now().Add(time.Hour) // End in the future

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(1000), // Lower quorum to match total power
		big.NewInt(1000), // Lower threshold to match total power
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	err = qv.StartVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to start voting: %v", err)
	}

	// Register voters and cast votes
	voter1, err := qv.RegisterVoter("bob", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register voter1: %v", err)
	}

	voter2, err := qv.RegisterVoter("charlie", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register voter2: %v", err)
	}

	// Cast votes
	_, err = qv.CastVote(proposal.ID, voter1.Address, VoteTypeYes, big.NewInt(500), nil)
	if err != nil {
		t.Fatalf("Failed to cast vote1: %v", err)
	}

	_, err = qv.CastVote(proposal.ID, voter2.Address, VoteTypeYes, big.NewInt(500), nil)
	if err != nil {
		t.Fatalf("Failed to cast vote2: %v", err)
	}

	// Close voting
	result, err := qv.CloseVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Expected CloseVoting to succeed, got error: %v", err)
	}

	if result.TotalVotes != 2 {
		t.Errorf("Expected TotalVotes to be 2, got %d", result.TotalVotes)
	}

	if result.TotalPower.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("Expected TotalPower to be 1000, got %s", result.TotalPower.String())
	}

	if !result.QuorumReached {
		t.Error("Expected QuorumReached to be true")
	}
}

func TestExecuteProposal(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Create proposal with short execution delay
	votingStart := time.Now().Add(-time.Hour)
	votingEnd := time.Now().Add(time.Hour) // End in the future
	executionDelay := time.Millisecond * 100

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		executionDelay,
		big.NewInt(1000),
		big.NewInt(1000), // Lower threshold to match total power
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	err = qv.StartVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to start voting: %v", err)
	}

	// Register voter and cast vote
	voter, err := qv.RegisterVoter("bob", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register voter: %v", err)
	}

	_, err = qv.CastVote(proposal.ID, voter.Address, VoteTypeYes, big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Close voting
	_, err = qv.CloseVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to close voting: %v", err)
	}

	// Wait for execution delay
	time.Sleep(time.Millisecond * 200)

	// Execute proposal
	err = qv.ExecuteProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Expected ExecuteProposal to succeed, got error: %v", err)
	}

	// Verify status was updated
	executedProposal, err := qv.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	if executedProposal.Status != ProposalStatusExecuted {
		t.Errorf("Expected Status to be Executed, got %v", executedProposal.Status)
	}
}

func TestGetProposal(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		time.Now().Add(time.Hour),
		time.Now().Add(time.Hour*24),
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Retrieve the proposal
	retrievedProposal, err := qv.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Expected GetProposal to succeed, got error: %v", err)
	}

	if retrievedProposal.ID != proposal.ID {
		t.Errorf("Expected ID to match, got %s vs %s", retrievedProposal.ID, proposal.ID)
	}

	// Verify it's a deep copy
	originalTitle := retrievedProposal.Title
	retrievedProposal.Title = "Modified Title"

	storedProposal, _ := qv.GetProposal(proposal.ID)
	if storedProposal.Title == "Modified Title" {
		t.Error("Expected stored proposal to not be affected by external modifications")
	}

	if storedProposal.Title != originalTitle {
		t.Errorf("Expected stored proposal title to remain unchanged, got %s", storedProposal.Title)
	}
}

func TestGetVotingResult(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Create proposal and complete voting
	votingStart := time.Now().Add(-time.Hour)
	votingEnd := time.Now().Add(time.Hour) // End in the future

	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		votingStart,
		votingEnd,
		time.Hour,
		big.NewInt(1000),
		big.NewInt(5000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	err = qv.ActivateProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to activate proposal: %v", err)
	}

	err = qv.StartVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to start voting: %v", err)
	}

	// Register voter and cast vote
	voter, err := qv.RegisterVoter("bob", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	_, err = qv.CastVote(proposal.ID, voter.Address, VoteTypeYes, big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Close voting
	_, err = qv.CloseVoting(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to close voting: %v", err)
	}

	// Get voting result
	result, err := qv.GetVotingResult(proposal.ID)
	if err != nil {
		t.Fatalf("Expected GetVotingResult to succeed, got error: %v", err)
	}

	if result.ProposalID != proposal.ID {
		t.Errorf("Expected ProposalID to match, got %s vs %s", result.ProposalID, proposal.ID)
	}

	if result.TotalVotes != 1 {
		t.Errorf("Expected TotalVotes to be 1, got %d", result.TotalVotes)
	}
}

func TestGetVoter(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	voter, err := qv.RegisterVoter("alice", big.NewInt(1000), nil)
	if err != nil {
		t.Fatalf("Failed to register voter: %v", err)
	}

	// Retrieve the voter
	retrievedVoter, err := qv.GetVoter(voter.Address)
	if err != nil {
		t.Fatalf("Expected GetVoter to succeed, got error: %v", err)
	}

	if retrievedVoter.Address != voter.Address {
		t.Errorf("Expected Address to match, got %s vs %s", retrievedVoter.Address, voter.Address)
	}

	// Test non-existent voter
	_, err = qv.GetVoter("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent voter")
	}
}

func TestConcurrency(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Start the system
	err := qv.Start()
	if err != nil {
		t.Fatalf("Failed to start QuadraticVoting: %v", err)
	}
	defer qv.Stop()

	// Test concurrent proposal creation
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_, err := qv.CreateProposal(
				fmt.Sprintf("Proposal_%d", id),
				"Test Description",
				"alice",
				time.Now().Add(time.Hour),
				time.Now().Add(time.Hour*24),
				time.Hour,
				big.NewInt(2000),
				big.NewInt(10000),
				nil,
			)
			if err != nil {
				// Log error but continue (some might fail due to limits)
				return
			}
		}(i)
	}

	wg.Wait()

	// Verify some proposals were created
	if len(qv.Proposals) == 0 {
		t.Error("Expected some proposals to be created")
	}
}

func TestStringMethods(t *testing.T) {
	// Test VoteType.String()
	if VoteTypeYes.String() != "yes" {
		t.Errorf("Expected 'yes', got %s", VoteTypeYes.String())
	}

	if VoteTypeNo.String() != "no" {
		t.Errorf("Expected 'no', got %s", VoteTypeNo.String())
	}

	if VoteTypeAbstain.String() != "abstain" {
		t.Errorf("Expected 'abstain', got %s", VoteTypeAbstain.String())
	}

	// Test ProposalStatus.String()
	if ProposalStatusDraft.String() != "draft" {
		t.Errorf("Expected 'draft', got %s", ProposalStatusDraft.String())
	}

	if ProposalStatusVoting.String() != "voting" {
		t.Errorf("Expected 'voting', got %s", ProposalStatusVoting.String())
	}

	// Test SybilStatus.String()
	if SybilStatusVerified.String() != "verified" {
		t.Errorf("Expected 'verified', got %s", SybilStatusVerified.String())
	}

	if SybilStatusRejected.String() != "rejected" {
		t.Errorf("Expected 'rejected', got %s", SybilStatusRejected.String())
	}
}

func TestMemorySafety(t *testing.T) {
	qv := NewQuadraticVoting(QuadraticVotingConfig{})

	// Create a proposal
	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		time.Now().Add(time.Hour),
		time.Now().Add(time.Hour*24),
		time.Hour,
		big.NewInt(2000),
		big.NewInt(10000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Get the proposal and modify it
	retrievedProposal, err := qv.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("Failed to get proposal: %v", err)
	}

	originalTitle := retrievedProposal.Title
	retrievedProposal.Title = "Modified Title"

	// Verify internal state wasn't affected
	storedProposal, _ := qv.GetProposal(proposal.ID)
	if storedProposal.Title == "Modified Title" {
		t.Error("Expected internal state to not be affected by external modifications")
	}

	if storedProposal.Title != originalTitle {
		t.Errorf("Expected stored proposal title to remain unchanged, got %s", storedProposal.Title)
	}
}

func TestCleanupOldData(t *testing.T) {
	// Create QuadraticVoting with short timeout
	config := QuadraticVotingConfig{
		VotingTimeout:   time.Millisecond * 100,
		CleanupInterval: time.Millisecond * 50,
	}
	qv := NewQuadraticVoting(config)

	// Start the system
	err := qv.Start()
	if err != nil {
		t.Fatalf("Failed to start QuadraticVoting: %v", err)
	}
	defer qv.Stop()

	// Create a proposal
	proposal, err := qv.CreateProposal(
		"Test Proposal",
		"Test Description",
		"alice",
		time.Now().Add(-time.Hour),
		time.Now().Add(-time.Minute),
		time.Hour,
		big.NewInt(1000),
		big.NewInt(5000),
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create proposal: %v", err)
	}

	// Manually set status to rejected to trigger cleanup
	qv.mu.Lock()
	proposal.Status = ProposalStatusRejected
	qv.mu.Unlock()

	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)

	// Verify proposal was cleaned up
	_, err = qv.GetProposal(proposal.ID)
	if err == nil {
		t.Error("Expected rejected proposal to be cleaned up")
	}
}
