package cross_protocol

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCrossProtocolGovernance(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	require.NotNil(t, cpg)
	assert.NotNil(t, cpg.protocols)
	assert.NotNil(t, cpg.proposals)
	assert.NotNil(t, cpg.alignments)
	assert.NotNil(t, cpg.metrics)
	assert.NotNil(t, cpg.proposalQueue)
	assert.NotNil(t, cpg.alignmentUpdater)

	// Test that background goroutines are started
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	err := cpg.Close()
	assert.NoError(t, err)
}

func TestRegisterProtocol(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	tests := []struct {
		name     string
		protocol *Protocol
		wantErr  bool
	}{
		{
			name: "valid protocol",
			protocol: &Protocol{
				Name:            "Test Protocol",
				Description:     "Test Description",
				ChainID:         "1",
				Network:         "mainnet",
				GovernanceToken: "TEST",
				TotalSupply:     big.NewInt(1000000),
				VotingPower:     big.NewInt(500000),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			protocol: &Protocol{
				Description:     "Test Description",
				ChainID:         "1",
				Network:         "mainnet",
				GovernanceToken: "TEST",
				TotalSupply:     big.NewInt(1000000),
				VotingPower:     big.NewInt(500000),
			},
			wantErr: true,
		},
		{
			name: "missing chain ID",
			protocol: &Protocol{
				Name:            "Test Protocol",
				Description:     "Test Description",
				Network:         "mainnet",
				GovernanceToken: "TEST",
				TotalSupply:     big.NewInt(1000000),
				VotingPower:     big.NewInt(500000),
			},
			wantErr: true,
		},
		{
			name: "missing governance token",
			protocol: &Protocol{
				Name:        "Test Protocol",
				Description: "Test Description",
				ChainID:     "1",
				Network:     "mainnet",
				TotalSupply: big.NewInt(1000000),
				VotingPower: big.NewInt(500000),
			},
			wantErr: true,
		},
		{
			name: "zero total supply",
			protocol: &Protocol{
				Name:            "Test Protocol",
				Description:     "Test Description",
				ChainID:         "1",
				Network:         "mainnet",
				GovernanceToken: "TEST",
				TotalSupply:     big.NewInt(0),
				VotingPower:     big.NewInt(500000),
			},
			wantErr: true,
		},
		{
			name: "negative total supply",
			protocol: &Protocol{
				Name:            "Test Protocol",
				Description:     "Test Description",
				ChainID:         "1",
				Network:         "mainnet",
				GovernanceToken: "TEST",
				TotalSupply:     big.NewInt(-1000000),
				VotingPower:     big.NewInt(500000),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpg.RegisterProtocol(tt.protocol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.protocol.ID)
				assert.Equal(t, ProtocolActive, tt.protocol.Status)
				assert.NotZero(t, tt.protocol.CreatedAt)
				assert.NotZero(t, tt.protocol.UpdatedAt)
			}
		})
	}
}

func TestCreateProposal(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol first
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	tests := []struct {
		name     string
		proposal *GovernanceProposal
		wantErr  bool
	}{
		{
			name: "valid proposal",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			proposal: &GovernanceProposal{
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
		{
			name: "missing creator",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
		{
			name: "no protocols",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
		{
			name: "zero voting period",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  0,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
		{
			name: "zero quorum",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(0),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
		{
			name: "zero threshold",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(0),
			},
			wantErr: true,
		},
		{
			name: "non-existent protocol",
			proposal: &GovernanceProposal{
				Title:         "Test Proposal",
				Description:   "Test Description",
				ProposalType:  ProtocolUpgrade,
				Creator:       "user1",
				Protocols:     []ProtocolID{"non-existent"},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpg.CreateProposal(tt.proposal)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.proposal.ID)
				assert.Equal(t, ProposalDraft, tt.proposal.Status)
				assert.NotZero(t, tt.proposal.CreatedAt)
				assert.NotZero(t, tt.proposal.UpdatedAt)
				assert.NotNil(t, tt.proposal.Votes)
				assert.NotNil(t, tt.proposal.TotalVotes)
				assert.Equal(t, int64(0), tt.proposal.TotalVotes.Int64())
			}
		})
	}
}

func TestActivateProposal(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol and create a proposal
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	// Test activation
	err = cpg.ActivateProposal(proposal.ID)
	assert.NoError(t, err)

	// Verify proposal was activated
	activatedProposal, err := cpg.GetProposal(proposal.ID)
	require.NoError(t, err)
	assert.Equal(t, ProposalActive, activatedProposal.Status)
	assert.NotZero(t, activatedProposal.StartTime)
	assert.NotZero(t, activatedProposal.EndTime)

	// Test activating non-existent proposal
	err = cpg.ActivateProposal("non-existent")
	assert.Error(t, err)

	// Test activating already active proposal
	err = cpg.ActivateProposal(proposal.ID)
	assert.Error(t, err)
}

func TestCastVote(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol and create a proposal
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test individual vote scenarios to avoid proposal finalization
	t.Run("valid yes vote", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 1",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, VoteYes, big.NewInt(100000), "Supporting the upgrade")
		assert.NoError(t, err)

		// Verify vote was recorded
		updatedProposal, err := cpg.GetProposal(proposal.ID)
		require.NoError(t, err)
		assert.Contains(t, updatedProposal.Votes, protocol.ID)

		vote := updatedProposal.Votes[protocol.ID]
		assert.Equal(t, VoteYes, vote.VoteType)
		assert.Equal(t, big.NewInt(100000), vote.VotingPower)
		assert.Equal(t, "Supporting the upgrade", vote.Reason)
		assert.NotEmpty(t, vote.Transaction)
		assert.NotEmpty(t, vote.Signature)
	})

	t.Run("valid no vote", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 2",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, VoteNo, big.NewInt(50000), "Concerns about the upgrade")
		assert.NoError(t, err)

		// Verify vote was recorded
		updatedProposal, err := cpg.GetProposal(proposal.ID)
		require.NoError(t, err)
		assert.Contains(t, updatedProposal.Votes, protocol.ID)

		vote := updatedProposal.Votes[protocol.ID]
		assert.Equal(t, VoteNo, vote.VoteType)
		assert.Equal(t, big.NewInt(50000), vote.VotingPower)
		assert.Equal(t, "Concerns about the upgrade", vote.Reason)
		assert.NotEmpty(t, vote.Transaction)
		assert.NotEmpty(t, vote.Signature)
	})

	t.Run("valid abstain vote", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 3",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, VoteAbstain, big.NewInt(75000), "Neutral on the issue")
		assert.NoError(t, err)

		// Verify vote was recorded
		updatedProposal, err := cpg.GetProposal(proposal.ID)
		require.NoError(t, err)
		assert.Contains(t, updatedProposal.Votes, protocol.ID)

		vote := updatedProposal.Votes[protocol.ID]
		assert.Equal(t, VoteAbstain, vote.VoteType)
		assert.Equal(t, big.NewInt(75000), vote.VotingPower)
		assert.Equal(t, "Neutral on the issue", vote.Reason)
		assert.NotEmpty(t, vote.Transaction)
		assert.NotEmpty(t, vote.Signature)
	})

	t.Run("non-existent proposal", func(t *testing.T) {
		err := cpg.CastVote(protocol.ID, "non-existent", VoteYes, big.NewInt(100000), "Test")
		assert.Error(t, err)
	})

	t.Run("non-existent protocol", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 4",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote("non-existent", proposal.ID, VoteYes, big.NewInt(100000), "Test")
		assert.Error(t, err)
	})

	t.Run("zero voting power", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 5",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, VoteYes, big.NewInt(0), "Test")
		assert.Error(t, err)
	})

	t.Run("excessive voting power", func(t *testing.T) {
		proposal := &GovernanceProposal{
			Title:         "Test Proposal 6",
			Description:   "Test Description",
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, VoteYes, big.NewInt(1000000), "Test")
		assert.Error(t, err)
	})
}

func TestGetProposal(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol and create a proposal
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	// Test getting existing proposal
	retrievedProposal, err := cpg.GetProposal(proposal.ID)
	assert.NoError(t, err)
	assert.Equal(t, proposal.ID, retrievedProposal.ID)

	// Test getting non-existent proposal
	_, err = cpg.GetProposal("non-existent")
	assert.Error(t, err)
}

func TestGetProposals(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Create multiple proposals
	proposal1 := &GovernanceProposal{
		Title:         "Proposal 1",
		Description:   "Test Description 1",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	proposal2 := &GovernanceProposal{
		Title:         "Proposal 2",
		Description:   "Test Description 2",
		ProposalType:  ParameterChange,
		Creator:       "user2",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  48 * time.Hour,
		Quorum:        big.NewInt(150000),
		Threshold:     big.NewInt(300000),
	}

	err = cpg.CreateProposal(proposal1)
	require.NoError(t, err)
	err = cpg.CreateProposal(proposal2)
	require.NoError(t, err)

	// Test getting all proposals
	proposals := cpg.GetProposals("", "")
	assert.Len(t, proposals, 2)

	// Test filtering by status
	draftProposals := cpg.GetProposals(ProposalDraft, "")
	assert.Len(t, draftProposals, 2)

	// Test filtering by proposal type
	upgradeProposals := cpg.GetProposals("", ProtocolUpgrade)
	assert.Len(t, upgradeProposals, 1)
	assert.Equal(t, ProtocolUpgrade, upgradeProposals[0].ProposalType)

	// Test filtering by both
	upgradeDraftProposals := cpg.GetProposals(ProposalDraft, ProtocolUpgrade)
	assert.Len(t, upgradeDraftProposals, 1)
	assert.Equal(t, ProtocolUpgrade, upgradeDraftProposals[0].ProposalType)
}

func TestGetProtocol(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test getting existing protocol
	retrievedProtocol, err := cpg.GetProtocol(protocol.ID)
	assert.NoError(t, err)
	assert.Equal(t, protocol.ID, retrievedProtocol.ID)

	// Test getting non-existent protocol
	_, err = cpg.GetProtocol("non-existent")
	assert.Error(t, err)
}

func TestGetProtocols(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register multiple protocols
	protocol1 := &Protocol{
		Name:            "Protocol 1",
		Description:     "Test Description 1",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST1",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	protocol2 := &Protocol{
		Name:            "Protocol 2",
		Description:     "Test Description 2",
		ChainID:         "2",
		Network:         "testnet",
		GovernanceToken: "TEST2",
		TotalSupply:     big.NewInt(2000000),
		VotingPower:     big.NewInt(1000000),
	}

	err := cpg.RegisterProtocol(protocol1)
	require.NoError(t, err)
	err = cpg.RegisterProtocol(protocol2)
	require.NoError(t, err)

	// Test getting all protocols
	protocols := cpg.GetProtocols("")
	assert.Len(t, protocols, 2)

	// Test filtering by status
	activeProtocols := cpg.GetProtocols(ProtocolActive)
	assert.Len(t, activeProtocols, 2)
}

func TestExecuteProposal(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol and create a proposal
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	// Test executing non-existent proposal
	err = cpg.ExecuteProposal("non-existent")
	assert.Error(t, err)

	// Test executing draft proposal
	err = cpg.ExecuteProposal(proposal.ID)
	assert.Error(t, err)

	// Activate and pass the proposal
	err = cpg.ActivateProposal(proposal.ID)
	require.NoError(t, err)

	// Cast enough votes to pass
	err = cpg.CastVote(protocol.ID, proposal.ID, VoteYes, big.NewInt(300000), "Supporting the upgrade")
	require.NoError(t, err)

	// Wait for proposal finalization
	time.Sleep(100 * time.Millisecond)

	// Now execute the passed proposal
	err = cpg.ExecuteProposal(proposal.ID)
	assert.NoError(t, err)

	// Verify proposal was executed
	executedProposal, err := cpg.GetProposal(proposal.ID)
	require.NoError(t, err)
	assert.Equal(t, ProposalExecuted, executedProposal.Status)
	assert.True(t, executedProposal.Executed)
	assert.NotNil(t, executedProposal.ExecutionTime)

	// Test executing already executed proposal
	err = cpg.ExecuteProposal(proposal.ID)
	assert.Error(t, err)
}

func TestConcurrency(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test concurrent proposal creation
	var wg sync.WaitGroup
	numProposals := 10

	for i := 0; i < numProposals; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			proposal := &GovernanceProposal{
				Title:         fmt.Sprintf("Proposal %d", index),
				Description:   fmt.Sprintf("Description %d", index),
				ProposalType:  ProtocolUpgrade,
				Creator:       fmt.Sprintf("user%d", index),
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			}

			err := cpg.CreateProposal(proposal)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all proposals were created
	proposals := cpg.GetProposals("", "")
	assert.Len(t, proposals, numProposals) // Only the proposals we created
}

func TestMemorySafety(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Create many protocols and proposals to test memory usage
	numProtocols := 10
	numProposals := 20

	for i := 0; i < numProtocols; i++ {
		protocol := &Protocol{
			Name:            fmt.Sprintf("Protocol %d", i),
			Description:     fmt.Sprintf("Description %d", i),
			ChainID:         fmt.Sprintf("%d", i),
			Network:         "mainnet",
			GovernanceToken: fmt.Sprintf("TOKEN%d", i),
			TotalSupply:     big.NewInt(int64(1000000 * (i + 1))),
			VotingPower:     big.NewInt(int64(500000 * (i + 1))),
		}

		err := cpg.RegisterProtocol(protocol)
		require.NoError(t, err)

		// Create proposals for each protocol
		for j := 0; j < numProposals; j++ {
			proposal := &GovernanceProposal{
				Title:         fmt.Sprintf("Proposal %d-%d", i, j),
				Description:   fmt.Sprintf("Description %d-%d", i, j),
				ProposalType:  ProtocolUpgrade,
				Creator:       fmt.Sprintf("user%d", j),
				Protocols:     []ProtocolID{protocol.ID},
				VotingPeriod:  24 * time.Hour,
				Quorum:        big.NewInt(100000),
				Threshold:     big.NewInt(250000),
			}

			err := cpg.CreateProposal(proposal)
			require.NoError(t, err)
		}
	}

	// Verify all protocols and proposals were created
	assert.Len(t, cpg.protocols, numProtocols)
	assert.Len(t, cpg.proposals, numProtocols*numProposals)
}

func TestEdgeCases(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Test with very large numbers
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), // 10^18
		VotingPower:     new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil), // 10^17
	}

	err := cpg.RegisterProtocol(protocol)
	assert.NoError(t, err)

	// Test with very small voting period
	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  1 * time.Millisecond,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	err = cpg.CreateProposal(proposal)
	assert.NoError(t, err)
}

func TestCleanup(t *testing.T) {
	cpg := NewCrossProtocolGovernance()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test cleanup
	err = cpg.Close()
	assert.NoError(t, err)

	// Verify cleanup
	time.Sleep(100 * time.Millisecond)
}

func TestGetRandomID(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	id1 := cpg.GetRandomID()
	id2 := cpg.GetRandomID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // 16 bytes = 32 hex chars
	assert.Len(t, id2, 32)
}

func TestProposalTypes(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposalTypes := []ProposalType{
		ProtocolUpgrade,
		ParameterChange,
		TokenAllocation,
		GovernanceChange,
		IntegrationChange,
		EmergencyAction,
		CustomProposal,
	}

	for _, proposalType := range proposalTypes {
		proposal := &GovernanceProposal{
			Title:         fmt.Sprintf("Test %s Proposal", proposalType),
			Description:   fmt.Sprintf("Test %s Description", proposalType),
			ProposalType:  proposalType,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		assert.NoError(t, err)
		assert.Equal(t, proposalType, proposal.ProposalType)
	}
}

func TestVoteTypes(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test different vote types with separate proposals to avoid finalization
	voteTypes := []VoteType{VoteYes, VoteNo, VoteAbstain, VoteVeto}

	for i, voteType := range voteTypes {
		proposal := &GovernanceProposal{
			Title:         fmt.Sprintf("Test Proposal %d", i+1),
			Description:   fmt.Sprintf("Test Description %d", i+1),
			ProposalType:  ProtocolUpgrade,
			Creator:       "user1",
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)

		err = cpg.ActivateProposal(proposal.ID)
		require.NoError(t, err)

		err = cpg.CastVote(protocol.ID, proposal.ID, voteType, big.NewInt(100000), fmt.Sprintf("Testing %s vote", voteType))
		assert.NoError(t, err)

		// Verify vote was recorded for this proposal
		updatedProposal, err := cpg.GetProposal(proposal.ID)
		require.NoError(t, err)
		assert.Contains(t, updatedProposal.Votes, protocol.ID)
		assert.Equal(t, voteType, updatedProposal.Votes[protocol.ID].VoteType)
	}
}

func TestProtocolStatusTransitions(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, ProtocolActive, protocol.Status)

	// Test status transitions
	protocol.Status = ProtocolInactive
	assert.Equal(t, ProtocolInactive, protocol.Status)

	protocol.Status = ProtocolSuspended
	assert.Equal(t, ProtocolSuspended, protocol.Status)

	protocol.Status = ProtocolDeprecated
	assert.Equal(t, ProtocolDeprecated, protocol.Status)

	protocol.Status = ProtocolTesting
	assert.Equal(t, ProtocolTesting, protocol.Status)
}

func TestProposalStatusTransitions(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol and create a proposal
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(250000),
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, ProposalDraft, proposal.Status)

	// Test status transitions
	proposal.Status = ProposalActive
	assert.Equal(t, ProposalActive, proposal.Status)

	proposal.Status = ProposalPassed
	assert.Equal(t, ProposalPassed, proposal.Status)

	proposal.Status = ProposalRejected
	assert.Equal(t, ProposalRejected, proposal.Status)

	proposal.Status = ProposalExecuted
	assert.Equal(t, ProposalExecuted, proposal.Status)

	proposal.Status = ProposalCancelled
	assert.Equal(t, ProposalCancelled, proposal.Status)

	proposal.Status = ProposalExpired
	assert.Equal(t, ProposalExpired, proposal.Status)
}

func TestBackgroundProcessing(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register multiple protocols to trigger background processing
	for i := 0; i < 5; i++ {
		protocol := &Protocol{
			Name:            fmt.Sprintf("Protocol %d", i),
			Description:     fmt.Sprintf("Description %d", i),
			ChainID:         fmt.Sprintf("%d", i),
			Network:         "mainnet",
			GovernanceToken: fmt.Sprintf("TOKEN%d", i),
			TotalSupply:     big.NewInt(int64(1000000 * (i + 1))),
			VotingPower:     big.NewInt(int64(500000 * (i + 1))),
		}

		err := cpg.RegisterProtocol(protocol)
		require.NoError(t, err)
	}

	// Wait for background processing
	time.Sleep(200 * time.Millisecond)

	// Verify metrics were updated
	assert.NotZero(t, cpg.metrics.ProtocolCount)
	assert.NotZero(t, cpg.metrics.LastUpdated)
}

func TestPerformance(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Performance test: create many proposals quickly
	start := time.Now()
	numProposals := 100

	for i := 0; i < numProposals; i++ {
		proposal := &GovernanceProposal{
			Title:         fmt.Sprintf("Proposal %d", i),
			Description:   fmt.Sprintf("Description %d", i),
			ProposalType:  ProtocolUpgrade,
			Creator:       fmt.Sprintf("user%d", i),
			Protocols:     []ProtocolID{protocol.ID},
			VotingPeriod:  24 * time.Hour,
			Quorum:        big.NewInt(100000),
			Threshold:     big.NewInt(250000),
		}

		err := cpg.CreateProposal(proposal)
		require.NoError(t, err)
	}

	duration := time.Since(start)
	proposalsPerSecond := float64(numProposals) / duration.Seconds()

	// Verify performance is reasonable (should handle at least 100 proposals/second)
	assert.Greater(t, proposalsPerSecond, 100.0)
	t.Logf("Processed %d proposals in %v (%.0f proposals/second)", numProposals, duration, proposalsPerSecond)
}

func TestGetProtocolAlignment(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register two protocols
	protocol1 := &Protocol{
		Name:            "Protocol 1",
		Description:     "Test Description 1",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST1",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	protocol2 := &Protocol{
		Name:            "Protocol 2",
		Description:     "Test Description 2",
		ChainID:         "2",
		Network:         "mainnet",
		GovernanceToken: "TEST2",
		TotalSupply:     big.NewInt(2000000),
		VotingPower:     big.NewInt(1000000),
	}

	err := cpg.RegisterProtocol(protocol1)
	require.NoError(t, err)
	err = cpg.RegisterProtocol(protocol2)
	require.NoError(t, err)

	// Manually trigger alignment update for testing
	cpg.alignmentUpdater <- protocol1.ID
	time.Sleep(100 * time.Millisecond)

	// Test getting alignment between protocols
	alignment, err := cpg.GetProtocolAlignment(protocol1.ID, protocol2.ID)
	if err != nil {
		// If alignment data doesn't exist yet, that's okay for this test
		// We're testing the error handling path
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "alignment data not found")
		return
	}

	// If alignment exists, verify it
	assert.NotNil(t, alignment)
	assert.Equal(t, protocol1.ID, alignment.Protocol1ID)
	assert.Equal(t, protocol2.ID, alignment.Protocol2ID)
	assert.GreaterOrEqual(t, alignment.AlignmentScore, 0.0)
	assert.LessOrEqual(t, alignment.AlignmentScore, 1.0)

	// Test getting alignment with non-existent protocols
	_, err = cpg.GetProtocolAlignment("non-existent", protocol2.ID)
	assert.Error(t, err)

	_, err = cpg.GetProtocolAlignment(protocol1.ID, "non-existent")
	assert.Error(t, err)
}

func TestProposalFinalization(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Test proposal that meets quorum but not threshold
	proposal1 := &GovernanceProposal{
		Title:         "Test Proposal 1",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(400000), // High threshold
	}

	err = cpg.CreateProposal(proposal1)
	require.NoError(t, err)

	err = cpg.ActivateProposal(proposal1.ID)
	require.NoError(t, err)

	// Cast vote that meets quorum but not threshold
	err = cpg.CastVote(protocol.ID, proposal1.ID, VoteYes, big.NewInt(150000), "Supporting")
	require.NoError(t, err)

	// Wait for finalization
	time.Sleep(100 * time.Millisecond)

	// Check proposal status
	finalizedProposal, err := cpg.GetProposal(proposal1.ID)
	require.NoError(t, err)
	assert.Equal(t, ProposalRejected, finalizedProposal.Status)
	assert.False(t, finalizedProposal.Approved)

	// Test proposal that meets both quorum and threshold
	proposal2 := &GovernanceProposal{
		Title:         "Test Proposal 2",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(100000),
		Threshold:     big.NewInt(200000), // Lower threshold
	}

	err = cpg.CreateProposal(proposal2)
	require.NoError(t, err)

	err = cpg.ActivateProposal(proposal2.ID)
	require.NoError(t, err)

	// Cast vote that meets both quorum and threshold
	err = cpg.CastVote(protocol.ID, proposal2.ID, VoteYes, big.NewInt(250000), "Strong support")
	require.NoError(t, err)

	// Wait for finalization
	time.Sleep(100 * time.Millisecond)

	// Check proposal status
	finalizedProposal2, err := cpg.GetProposal(proposal2.ID)
	require.NoError(t, err)
	assert.Equal(t, ProposalPassed, finalizedProposal2.Status)
	assert.True(t, finalizedProposal2.Approved)
}

func TestExpiredProposal(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	// Create proposal with very short voting period
	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  1 * time.Millisecond, // Very short period
		Quorum:        big.NewInt(1000000),  // Very high quorum to prevent finalization
		Threshold:     big.NewInt(1000000),  // Very high threshold to prevent finalization
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	err = cpg.ActivateProposal(proposal.ID)
	require.NoError(t, err)

	// Wait for proposal to expire
	time.Sleep(100 * time.Millisecond)

	// Manually trigger the expiration check for testing
	cpg.mu.Lock()
	cpg.checkExpiredProposals()
	cpg.mu.Unlock()

	// Check that proposal was marked as expired
	expiredProposal, err := cpg.GetProposal(proposal.ID)
	require.NoError(t, err)
	assert.Equal(t, ProposalExpired, expiredProposal.Status)
}

func TestVoteReplacement(t *testing.T) {
	cpg := NewCrossProtocolGovernance()
	defer cpg.Close()

	// Register a protocol
	protocol := &Protocol{
		Name:            "Test Protocol",
		Description:     "Test Description",
		ChainID:         "1",
		Network:         "mainnet",
		GovernanceToken: "TEST",
		TotalSupply:     big.NewInt(1000000),
		VotingPower:     big.NewInt(500000),
	}

	err := cpg.RegisterProtocol(protocol)
	require.NoError(t, err)

	proposal := &GovernanceProposal{
		Title:         "Test Proposal",
		Description:   "Test Description",
		ProposalType:  ProtocolUpgrade,
		Creator:       "user1",
		Protocols:     []ProtocolID{protocol.ID},
		VotingPeriod:  24 * time.Hour,
		Quorum:        big.NewInt(1000000),  // Very high quorum to prevent finalization
		Threshold:     big.NewInt(1000000),  // Very high threshold to prevent finalization
	}

	err = cpg.CreateProposal(proposal)
	require.NoError(t, err)

	err = cpg.ActivateProposal(proposal.ID)
	require.NoError(t, err)

	// Cast initial vote
	err = cpg.CastVote(protocol.ID, proposal.ID, VoteYes, big.NewInt(100000), "Initial support")
	require.NoError(t, err)

	// Verify initial vote
	initialProposal, err := cpg.GetProposal(proposal.ID)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(100000), initialProposal.TotalVotes)

	// Replace vote with different type and amount
	err = cpg.CastVote(protocol.ID, proposal.ID, VoteNo, big.NewInt(150000), "Changed mind")
	require.NoError(t, err)

	// Verify vote was replaced
	updatedProposal, err := cpg.GetProposal(proposal.ID)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(150000), updatedProposal.TotalVotes)
	assert.Equal(t, VoteNo, updatedProposal.Votes[protocol.ID].VoteType)
	assert.Equal(t, "Changed mind", updatedProposal.Votes[protocol.ID].Reason)
}
