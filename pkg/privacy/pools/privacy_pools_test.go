package pools

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestNewPrivacyPools(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	if pp == nil {
		t.Fatal("Expected non-nil PrivacyPools")
	}
	
	if pp.Config.MaxPools != 100 {
		t.Errorf("Expected MaxPools to be 100, got %d", pp.Config.MaxPools)
	}
	
	if pp.Config.MinAnonymitySet != 3 {
		t.Errorf("Expected MinAnonymitySet to be 3, got %d", pp.Config.MinAnonymitySet)
	}
	
	if len(pp.encryptionKey) != 32 {
		t.Errorf("Expected encryption key length to be 32, got %d", len(pp.encryptionKey))
	}
}

func TestStartStop(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	err := pp.Start()
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}
	
	if !pp.running {
		t.Error("Expected PrivacyPools to be running after Start")
	}
	
	err = pp.Stop()
	if err != nil {
		t.Fatalf("Expected Stop to succeed, got error: %v", err)
	}
	
	if pp.running {
		t.Error("Expected PrivacyPools to not be running after Stop")
	}
}

func TestCreatePrivacyPool(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Expected CreatePrivacyPool to succeed, got error: %v", err)
	}
	
	if pool.Name != "Test Pool" {
		t.Errorf("Expected Name to be 'Test Pool', got %s", pool.Name)
	}
	
	if pool.Type != PoolTypeCoinMixing {
		t.Errorf("Expected Type to be CoinMixing, got %v", pool.Type)
	}
}

func TestJoinPool(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool first
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Join the pool
	participant, err := pp.JoinPool(pool.ID, "alice", big.NewInt(5000), nil)
	
	if err != nil {
		t.Fatalf("Expected JoinPool to succeed, got error: %v", err)
	}
	
	if participant.Address != "alice" {
		t.Errorf("Expected Address to be 'alice', got %s", participant.Address)
	}
	
	if participant.PoolID != pool.ID {
		t.Errorf("Expected PoolID to match, got %s vs %s", participant.PoolID, pool.ID)
	}
}

func TestStartMixingRound(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		if err != nil {
			t.Fatalf("Failed to join pool: %v", err)
		}
	}
	
	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	
	if err != nil {
		t.Fatalf("Expected StartMixingRound to succeed, got error: %v", err)
	}
	
	if round.PoolID != pool.ID {
		t.Errorf("Expected PoolID to match, got %s vs %s", round.PoolID, pool.ID)
	}
	
	if len(round.Participants) != 3 {
		t.Errorf("Expected 3 participants, got %d", len(round.Participants))
	}
}

func TestExecuteMixing(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool and start mixing round
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		if err != nil {
			t.Fatalf("Failed to join pool: %v", err)
		}
	}
	
	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	if err != nil {
		t.Fatalf("Failed to start mixing round: %v", err)
	}
	
	// Execute mixing
	err = pp.ExecuteMixing(round.RoundID)
	if err != nil {
		t.Fatalf("Expected ExecuteMixing to succeed, got error: %v", err)
	}
	
	// Verify round was completed
	completedRound, err := pp.GetMixingRound(round.RoundID)
	if err != nil {
		t.Fatalf("Failed to get mixing round: %v", err)
	}
	
	if completedRound.Status != MixingRoundStatusCompleted {
		t.Errorf("Expected Status to be Completed, got %v", completedRound.Status)
	}
}

func TestCreateSelectiveDisclosure(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool and participant
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	participant, err := pp.JoinPool(pool.ID, "alice", big.NewInt(5000), nil)
	if err != nil {
		t.Fatalf("Failed to join pool: %v", err)
	}
	
	// Create selective disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("5000"),
		time.Now().Add(time.Hour),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Expected CreateSelectiveDisclosure to succeed, got error: %v", err)
	}
	
	if disclosure.ParticipantID != participant.ID {
		t.Errorf("Expected ParticipantID to match, got %s vs %s", disclosure.ParticipantID, participant.ID)
	}
	
	if disclosure.DisclosureType != DisclosureTypeAmount {
		t.Errorf("Expected DisclosureType to be Amount, got %v", disclosure.DisclosureType)
	}
}

func TestValidateDisclosure(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool and participant
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	participant, err := pp.JoinPool(pool.ID, "alice", big.NewInt(5000), nil)
	if err != nil {
		t.Fatalf("Failed to join pool: %v", err)
	}
	
	// Create selective disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("5000"),
		time.Now().Add(time.Hour),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create disclosure: %v", err)
	}
	
	// Validate disclosure
	err = pp.ValidateDisclosure(disclosure.ID)
	if err != nil {
		t.Fatalf("Expected ValidateDisclosure to succeed, got error: %v", err)
	}
	
	// Verify status was updated
	validatedDisclosure, err := pp.GetDisclosure(disclosure.ID)
	if err != nil {
		t.Fatalf("Failed to get disclosure: %v", err)
	}
	
	if validatedDisclosure.Status != DisclosureStatusValidated {
		t.Errorf("Expected Status to be Validated, got %v", validatedDisclosure.Status)
	}
}

func TestGetPool(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool
	createdPool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Retrieve the pool
	retrievedPool, err := pp.GetPool(createdPool.ID)
	if err != nil {
		t.Fatalf("Expected GetPool to succeed, got error: %v", err)
	}
	
	if retrievedPool.ID != createdPool.ID {
		t.Errorf("Expected ID to match, got %s vs %s", retrievedPool.ID, createdPool.ID)
	}
	
	// Verify it's a deep copy
	originalName := retrievedPool.Name
	retrievedPool.Name = "Modified Name"
	
	storedPool, _ := pp.GetPool(createdPool.ID)
	if storedPool.Name == "Modified Name" {
		t.Error("Expected stored pool to not be affected by external modifications")
	}
	
	if storedPool.Name != originalName {
		t.Errorf("Expected stored pool name to remain unchanged, got %s", storedPool.Name)
	}
}

func TestGetPools(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create multiple pools
	_, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"BTC Pool",
		"BTC Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create first pool: %v", err)
	}
	
	_, err = pp.CreatePrivacyPool(
		PoolTypePrivacyPool,
		"ETH Pool",
		"ETH Description",
		"ETH",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create second pool: %v", err)
	}
	
	// Get all pools
	allPools := pp.GetPools(nil)
	if len(allPools) != 2 {
		t.Errorf("Expected 2 pools, got %d", len(allPools))
	}
	
	// Get pools with filter
	btcPools := pp.GetPools(func(p *PrivacyPool) bool {
		return p.Asset == "BTC"
	})
	
	if len(btcPools) != 1 {
		t.Errorf("Expected 1 BTC pool, got %d", len(btcPools))
	}
	
	if btcPools[0].Asset != "BTC" {
		t.Errorf("Expected filtered pool to be BTC, got %s", btcPools[0].Asset)
	}
}

func TestConcurrency(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Start the system
	err := pp.Start()
	if err != nil {
		t.Fatalf("Failed to start PrivacyPools: %v", err)
	}
	defer pp.Stop()
	
	// Test concurrent pool creation
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_, err := pp.CreatePrivacyPool(
				PoolTypeCoinMixing,
				fmt.Sprintf("Pool_%d", id),
				"Test Description",
				"BTC",
				big.NewInt(1000),
				big.NewInt(10000),
				3,
				10,
				big.NewInt(10),
				nil,
			)
			if err != nil {
				// Log error but continue (some might fail due to limits)
				return
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify some pools were created
	if len(pp.Pools) == 0 {
		t.Error("Expected some pools to be created")
	}
}

func TestEdgeCases(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Test with zero amounts
	_, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(0),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err == nil {
		t.Error("Expected error when min amount is zero")
	}
	
	// Test with invalid participant limits
	_, err = pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		1, // Below minimum anonymity set
		10,
		big.NewInt(10),
		nil,
	)
	
	if err == nil {
		t.Error("Expected error when min participants is below anonymity set")
	}
}

func TestStringMethods(t *testing.T) {
	// Test PoolType.String()
	if PoolTypeCoinMixing.String() != "coin_mixing" {
		t.Errorf("Expected 'coin_mixing', got %s", PoolTypeCoinMixing.String())
	}
	
	// Test PoolStatus.String()
	if PoolStatusActive.String() != "active" {
		t.Errorf("Expected 'active', got %s", PoolStatusActive.String())
	}
	
	// Test MixingRoundStatus.String()
	if MixingRoundStatusPending.String() != "pending" {
		t.Errorf("Expected 'pending', got %s", MixingRoundStatusPending.String())
	}
	
	// Test ParticipantStatus.String()
	if ParticipantStatusActive.String() != "active" {
		t.Errorf("Expected 'active', got %s", ParticipantStatusActive.String())
	}
	
	// Test DisclosureType.String()
	if DisclosureTypeAmount.String() != "amount" {
		t.Errorf("Expected 'amount', got %s", DisclosureTypeAmount.String())
	}
	
	// Test DisclosureStatus.String()
	if DisclosureStatusPending.String() != "pending" {
		t.Errorf("Expected 'pending', got %s", DisclosureStatusPending.String())
	}
}

func TestMemorySafety(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})
	
	// Create a pool
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Get the pool and modify it
	retrievedPool, err := pp.GetPool(pool.ID)
	if err != nil {
		t.Fatalf("Failed to get pool: %v", err)
	}
	
	originalName := retrievedPool.Name
	retrievedPool.Name = "Modified Name"
	
	// Verify internal state wasn't affected
	storedPool, _ := pp.GetPool(pool.ID)
	if storedPool.Name == "Modified Name" {
		t.Error("Expected internal state to not be affected by external modifications")
	}
	
	if storedPool.Name != originalName {
		t.Errorf("Expected stored pool name to remain unchanged, got %s", storedPool.Name)
	}
}

func TestCleanupOldData(t *testing.T) {
	// Create PrivacyPools with short timeout
	config := PrivacyPoolsConfig{
		MixingTimeout:   time.Millisecond * 100,
		CleanupInterval: time.Millisecond * 50,
	}
	pp := NewPrivacyPools(config)
	
	// Start the system
	err := pp.Start()
	if err != nil {
		t.Fatalf("Failed to start PrivacyPools: %v", err)
	}
	defer pp.Stop()
	
	// Create a pool and start mixing round
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Test Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	
	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		if err != nil {
			t.Fatalf("Failed to join pool: %v", err)
		}
	}
	
	// Start and complete mixing round
	round, err := pp.StartMixingRound(pool.ID)
	if err != nil {
		t.Fatalf("Failed to start mixing round: %v", err)
	}
	
	err = pp.ExecuteMixing(round.RoundID)
	if err != nil {
		t.Fatalf("Failed to execute mixing: %v", err)
	}
	
	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)
	
	// Verify mixing round was cleaned up
	_, err = pp.GetMixingRound(round.RoundID)
	if err == nil {
		t.Error("Expected completed mixing round to be cleaned up")
	}
}
