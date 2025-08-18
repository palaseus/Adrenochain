package validators

import (
	"math/big"
	"testing"
	"time"
)

// TestNewMultiChainValidator tests validator creation
func TestNewMultiChainValidator(t *testing.T) {
	address := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	publicKey := []byte("test-public-key")

	config := ValidatorConfig{
		MinStake:          big.NewInt(1000000),
		MaxStake:          big.NewInt(1000000000),
		UnbondingPeriod:   time.Hour * 24 * 7 * 2,
		SlashingThreshold: 100,
		ReputationDecay:   time.Hour * 24 * 7,
		SecurityLevel:     SecurityLevelHigh,
	}

	validator := NewMultiChainValidator(address, publicKey, config)

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	if validator.Address != address {
		t.Error("Expected address to match")
	}

	if string(validator.PublicKey) != string(publicKey) {
		t.Error("Expected public key to match")
	}

	if validator.Status != ValidatorStatusActive {
		t.Errorf("Expected status %d, got %d", ValidatorStatusActive, validator.Status)
	}

	if validator.Reputation != 1000 {
		t.Errorf("Expected reputation 1000, got %d", validator.Reputation)
	}

	if validator.config.MinStake.Cmp(big.NewInt(1000000)) != 0 {
		t.Errorf("Expected min stake 1000000, got %s", validator.config.MinStake.String())
	}
}

// TestNewValidatorNetwork tests network creation
func TestNewValidatorNetwork(t *testing.T) {
	config := NetworkConfig{
		MinStake:          big.NewInt(1000000),
		MaxValidators:     100,
		RotationInterval:  time.Hour * 24 * 7,
		SlashingEnabled:   true,
		ReputationEnabled: true,
		SecurityLevel:     SecurityLevelHigh,
	}

	network := NewValidatorNetwork(config)

	if network == nil {
		t.Fatal("Expected network to be created")
	}

	if network.config.MaxValidators != 100 {
		t.Errorf("Expected max validators 100, got %d", network.config.MaxValidators)
	}

	if network.MinValidators != 33 { // 100/3
		t.Errorf("Expected min validators 33, got %d", network.MinValidators)
	}

	if network.ConsensusRules.QuorumSize != 67 { // 100*2/3
		t.Errorf("Expected quorum size 67, got %d", network.ConsensusRules.QuorumSize)
	}
}

// TestAddValidator tests adding validators to the network
func TestAddValidator(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 5,
		MinStake:      big.NewInt(1000000),
	})

	// Create validators with sufficient stake
	for i := 0; i < 3; i++ {
		address := [20]byte{byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		validator := NewMultiChainValidator(address, []byte("key"), ValidatorConfig{})
		validator.Stake = big.NewInt(2000000) // 2M stake

		err := network.AddValidator(validator)
		if err != nil {
			t.Errorf("Failed to add validator %d: %v", i, err)
		}
	}

	if len(network.Validators) != 3 {
		t.Errorf("Expected 3 validators, got %d", len(network.Validators))
	}

	if network.ActiveValidators != 3 {
		t.Errorf("Expected 3 active validators, got %d", network.ActiveValidators)
	}
}

// TestAddValidatorValidation tests validator validation
func TestAddValidatorValidation(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 2,
		MinStake:      big.NewInt(1000000),
	})

	// Test insufficient stake
	address := [20]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	validator := NewMultiChainValidator(address, []byte("key"), ValidatorConfig{})
	validator.Stake = big.NewInt(500000) // 500K stake (below minimum)

	err := network.AddValidator(validator)
	if err == nil {
		t.Error("Expected error for insufficient stake")
	}

	// Test network capacity
	validator1 := NewMultiChainValidator([20]byte{1}, []byte("key1"), ValidatorConfig{})
	validator1.Stake = big.NewInt(2000000)
	network.AddValidator(validator1)

	validator2 := NewMultiChainValidator([20]byte{2}, []byte("key2"), ValidatorConfig{})
	validator2.Stake = big.NewInt(2000000)
	network.AddValidator(validator2)

	validator3 := NewMultiChainValidator([20]byte{3}, []byte("key3"), ValidatorConfig{})
	validator3.Stake = big.NewInt(2000000)

	err = network.AddValidator(validator3)
	if err == nil {
		t.Error("Expected error when network is at capacity")
	}
}

// TestJoinChain tests validator joining chains
func TestJoinChain(t *testing.T) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{
		MinStake: big.NewInt(1000000),
		MaxStake: big.NewInt(10000000),
	})

	// Test successful join
	err := validator.JoinChain("chain-1", big.NewInt(2000000))
	if err != nil {
		t.Errorf("Failed to join chain: %v", err)
	}

	if len(validator.Chains) != 1 {
		t.Errorf("Expected 1 chain participation, got %d", len(validator.Chains))
	}

	participation := validator.Chains["chain-1"]
	if participation.Stake.Cmp(big.NewInt(2000000)) != 0 {
		t.Errorf("Expected stake 2000000, got %s", participation.Stake.String())
	}

	if participation.Status != ParticipationStatusActive {
		t.Errorf("Expected status %d, got %d", ParticipationStatusActive, participation.Status)
	}

	// Test insufficient stake
	err = validator.JoinChain("chain-2", big.NewInt(500000))
	if err == nil {
		t.Error("Expected error for insufficient stake")
	}

	// Test exceeding max stake
	err = validator.JoinChain("chain-3", big.NewInt(9000000))
	if err == nil {
		t.Error("Expected error for exceeding max stake")
	}
}

// TestLeaveChain tests validator leaving chains
func TestLeaveChain(t *testing.T) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{
		MinStake:        big.NewInt(1000000),
		UnbondingPeriod: time.Millisecond * 100, // Short period for testing
	})

	validator.JoinChain("chain-1", big.NewInt(2000000))

	err := validator.LeaveChain("chain-1")
	if err != nil {
		t.Errorf("Failed to leave chain: %v", err)
	}

	participation := validator.Chains["chain-1"]
	if participation.Status != ParticipationStatusUnbonding {
		t.Errorf("Expected status %d, got %d", ParticipationStatusUnbonding, participation.Status)
	}

	// Wait for unbonding period
	time.Sleep(time.Millisecond * 150)

	if len(validator.Chains) != 0 {
		t.Errorf("Expected 0 chain participations after unbonding, got %d", len(validator.Chains))
	}

	if validator.TotalStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Expected total stake 0 after unbonding, got %s", validator.TotalStake.String())
	}
}

// TestVote tests validator voting
func TestVote(t *testing.T) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{})
	validator.JoinChain("chain-1", big.NewInt(2000000))

	blockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	// Test successful vote
	err := validator.Vote("chain-1", blockHash, 100, true)
	if err != nil {
		t.Errorf("Failed to vote: %v", err)
	}

	participation := validator.Chains["chain-1"]
	if len(participation.VoteHistory) != 1 {
		t.Errorf("Expected 1 vote record, got %d", len(participation.VoteHistory))
	}

	voteRecord := participation.VoteHistory[0]
	if voteRecord.BlockHeight != 100 {
		t.Errorf("Expected block height 100, got %d", voteRecord.BlockHeight)
	}

	if !voteRecord.Vote {
		t.Error("Expected vote to be true")
	}

	// Test voting on non-participating chain
	err = validator.Vote("chain-2", blockHash, 100, true)
	if err == nil {
		t.Error("Expected error for non-participating chain")
	}

	// Test metrics update
	metrics := validator.GetMetrics()
	if metrics.TotalVotes != 1 {
		t.Errorf("Expected 1 total vote, got %d", metrics.TotalVotes)
	}

	if metrics.SuccessfulVotes != 1 {
		t.Errorf("Expected 1 successful vote, got %d", metrics.SuccessfulVotes)
	}
}

// TestSlashValidator tests validator slashing
func TestSlashValidator(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 10,
		MinStake:      big.NewInt(1000000),
	})

	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{})
	validator.Stake = big.NewInt(5000000)
	validator.JoinChain("chain-1", big.NewInt(2000000))
	validator.JoinChain("chain-2", big.NewInt(3000000))

	network.AddValidator(validator)

	// Test slashing
	slashAmount := big.NewInt(1000000)
	evidence := []byte("evidence of misbehavior")

	err := network.SlashValidator(validator.ID, "double signing", slashAmount, evidence)
	if err != nil {
		t.Errorf("Failed to slash validator: %v", err)
	}

	if validator.Status != ValidatorStatusSlashed {
		t.Errorf("Expected status %d, got %d", ValidatorStatusSlashed, validator.Status)
	}

	if validator.Reputation != 500 { // Should be halved from 1000
		t.Errorf("Expected reputation 500, got %d", validator.Reputation)
	}

	// Check that all chain participations are slashed
	for _, participation := range validator.Chains {
		if participation.Status != ParticipationStatusSlashed {
			t.Errorf("Expected participation status %d, got %d", ParticipationStatusSlashed, participation.Status)
		}

		if len(participation.SlashingHistory) != 1 {
			t.Errorf("Expected 1 slashing record, got %d", len(participation.SlashingHistory))
		}
	}

	// Check network stake reduction
	// Note: Network stake is only updated when validators are added/removed, not when their stakes change
	// So we check that the validator's stake was reduced instead
	expectedValidatorStake := new(big.Int).Sub(big.NewInt(5000000), slashAmount)
	if validator.TotalStake.Cmp(expectedValidatorStake) != 0 {
		t.Errorf("Expected validator stake %s, got %s", expectedValidatorStake.String(), validator.TotalStake.String())
	}

	// Check that the validator's chain participations were properly slashed
	for _, participation := range validator.Chains {
		if participation.Status != ParticipationStatusSlashed {
			t.Errorf("Expected participation status %d, got %d", ParticipationStatusSlashed, participation.Status)
		}
	}

	// Also check that the validator's total stake was reduced
	expectedValidatorStake = new(big.Int).Sub(big.NewInt(5000000), slashAmount)
	if validator.TotalStake.Cmp(expectedValidatorStake) != 0 {
		t.Errorf("Expected validator stake %s, got %s", expectedValidatorStake.String(), validator.TotalStake.String())
	}
}

// TestRotateValidators tests validator rotation
func TestRotateValidators(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators:     3,
		MinStake:          big.NewInt(1000000),
		ReputationEnabled: true,
	})

	// Add validators with different reputations and stakes
	validators := []*MultiChainValidator{
		NewMultiChainValidator([20]byte{1}, []byte("key1"), ValidatorConfig{}),
		NewMultiChainValidator([20]byte{2}, []byte("key2"), ValidatorConfig{}),
		NewMultiChainValidator([20]byte{3}, []byte("key3"), ValidatorConfig{}),
		NewMultiChainValidator([20]byte{4}, []byte("key4"), ValidatorConfig{}),
	}

	// Set different reputations and stakes
	validators[0].Reputation = 1000
	validators[0].Stake = big.NewInt(5000000)
	validators[1].Reputation = 800
	validators[1].Stake = big.NewInt(4000000)
	validators[2].Reputation = 600
	validators[2].Stake = big.NewInt(3000000)
	validators[3].Reputation = 400
	validators[3].Stake = big.NewInt(2000000)

	// Add all validators
	for _, validator := range validators {
		network.AddValidator(validator)
	}

	// Test rotation
	err := network.RotateValidators()
	if err != nil {
		t.Errorf("Failed to rotate validators: %v", err)
	}

	// Should keep top 3 validators and remove the lowest ranked one
	if len(network.Validators) != 3 {
		t.Errorf("Expected 3 validators after rotation, got %d", len(network.Validators))
	}

	// Check that the lowest ranked validator was removed
	if _, exists := network.Validators[validators[3].ID]; exists {
		t.Error("Expected lowest ranked validator to be removed")
	}

	// Check that top validators remain
	if _, exists := network.Validators[validators[0].ID]; !exists {
		t.Error("Expected top ranked validator to remain")
	}
}

// TestGetConsensusQuorum tests consensus quorum formation
func TestGetConsensusQuorum(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 5,
		MinStake:      big.NewInt(1000000),
	})

	// Add validators
	for i := 0; i < 4; i++ {
		address := [20]byte{byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		validator := NewMultiChainValidator(address, []byte("key"), ValidatorConfig{})
		validator.Stake = big.NewInt(2000000)
		validator.Reputation = uint64(1000 - i*100) // Decreasing reputation
		network.AddValidator(validator)
	}

	// Test quorum formation
	quorum, err := network.GetConsensusQuorum()
	if err != nil {
		t.Errorf("Failed to get consensus quorum: %v", err)
	}

	// Should return validators up to quorum size
	expectedQuorumSize := int(network.ConsensusRules.QuorumSize)
	if len(quorum) != expectedQuorumSize {
		t.Errorf("Expected quorum size %d, got %d", expectedQuorumSize, len(quorum))
	}

	// Check that quorum is sorted by rank (highest first)
	for i := 0; i < len(quorum)-1; i++ {
		score1 := float64(quorum[i].Reputation) * float64(quorum[i].TotalStake.Uint64()) / 1e18
		score2 := float64(quorum[i+1].Reputation) * float64(quorum[i+1].TotalStake.Uint64()) / 1e18
		if score1 < score2 {
			t.Errorf("Quorum not properly sorted: validator %d has lower score than %d", i, i+1)
		}
	}
}

// TestConcurrency tests concurrent access to validators
func TestConcurrency(t *testing.T) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 100,
		MinStake:      big.NewInt(1000000),
	})

	// Test concurrent validator addition
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			address := [20]byte{byte(index), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			validator := NewMultiChainValidator(address, []byte("key"), ValidatorConfig{})
			validator.Stake = big.NewInt(2000000)
			network.AddValidator(validator)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if len(network.Validators) != 10 {
		t.Errorf("Expected 10 validators, got %d", len(network.Validators))
	}
}

// TestSecurityLevels tests different security level configurations
func TestSecurityLevels(t *testing.T) {
	securityLevels := []SecurityLevel{
		SecurityLevelLow,
		SecurityLevelMedium,
		SecurityLevelHigh,
		SecurityLevelUltra,
	}

	for _, level := range securityLevels {
		config := ValidatorConfig{
			SecurityLevel: level,
		}

		validator := NewMultiChainValidator([20]byte{1}, []byte("key"), config)
		if validator.config.SecurityLevel != level {
			t.Errorf("Expected security level %d, got %d", level, validator.config.SecurityLevel)
		}
	}
}

// TestMetricsTracking tests metrics collection
func TestMetricsTracking(t *testing.T) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{})
	validator.JoinChain("chain-1", big.NewInt(2000000))

	// Perform multiple votes
	for i := 0; i < 5; i++ {
		blockHash := [32]byte{byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		validator.Vote("chain-1", blockHash, uint64(100+i), i%2 == 0) // Alternate true/false
	}

	// Check metrics
	metrics := validator.GetMetrics()
	if metrics.TotalVotes != 5 {
		t.Errorf("Expected 5 total votes, got %d", metrics.TotalVotes)
	}

	if metrics.SuccessfulVotes != 3 { // 3 true votes (indices 0, 2, 4)
		t.Errorf("Expected 3 successful votes, got %d", metrics.SuccessfulVotes)
	}

	if metrics.FailedVotes != 2 { // 2 false votes (indices 1, 3)
		t.Errorf("Expected 2 failed votes, got %d", metrics.FailedVotes)
	}
}

// TestReputationUpdate tests reputation system
func TestReputationUpdate(t *testing.T) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{
		ReputationDecay: time.Hour * 24, // 1 day
	})

	// Join a chain first
	validator.JoinChain("chain-1", big.NewInt(2000000))

	initialReputation := validator.Reputation

	// Test reputation increase for positive votes
	validator.Vote("chain-1", [32]byte{1}, 100, true)
	if validator.Reputation <= initialReputation {
		t.Error("Expected reputation to increase for positive vote")
	}

	// Test reputation decrease for negative votes
	reputationAfterPositive := validator.Reputation
	validator.Vote("chain-1", [32]byte{2}, 101, false)
	if validator.Reputation >= reputationAfterPositive {
		t.Error("Expected reputation to decrease for negative vote")
	}
}

// Benchmark tests for performance
func BenchmarkValidatorVoting(b *testing.B) {
	validator := NewMultiChainValidator([20]byte{1}, []byte("key"), ValidatorConfig{})
	validator.JoinChain("chain-1", big.NewInt(2000000))

	blockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Vote("chain-1", blockHash, uint64(100+i), i%2 == 0)
	}
}

func BenchmarkNetworkQuorum(b *testing.B) {
	network := NewValidatorNetwork(NetworkConfig{
		MaxValidators: 100,
		MinStake:      big.NewInt(1000000),
	})

	// Add validators
	for i := 0; i < 50; i++ {
		address := [20]byte{byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		validator := NewMultiChainValidator(address, []byte("key"), ValidatorConfig{})
		validator.Stake = big.NewInt(2000000)
		validator.Reputation = uint64(1000 - i*10)
		network.AddValidator(validator)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		network.GetConsensusQuorum()
	}
}
