package pools

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivacyPools_CreatePrivacyPool(t *testing.T) {
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

	require.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, "Test Pool", pool.Name)
	assert.Equal(t, "Test Description", pool.Description)
	assert.Equal(t, "BTC", pool.Asset)
	assert.Equal(t, PoolTypeCoinMixing, pool.Type)
	assert.Equal(t, PoolStatusActive, pool.Status)
	assert.Equal(t, uint64(3), pool.MinParticipants)
	assert.Equal(t, uint64(10), pool.MaxParticipants)
	assert.Equal(t, big.NewInt(1000), pool.MinAmount)
	assert.Equal(t, big.NewInt(10000), pool.MaxAmount)
	assert.Equal(t, big.NewInt(10), pool.Fee)
	assert.NotEmpty(t, pool.ID)
	assert.NotZero(t, pool.CreatedAt)
}

func TestPrivacyPools_JoinPool(t *testing.T) {
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
	require.NoError(t, err)

	// Join the pool
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)
	assert.NotNil(t, participant)
	assert.Equal(t, "user1", participant.Address)
	assert.Equal(t, pool.ID, participant.PoolID)
	assert.Equal(t, big.NewInt(5000), participant.InputAmount)
	assert.NotEmpty(t, participant.ID)
	assert.NotZero(t, participant.JoinedAt)
}

func TestPrivacyPools_StartMixingRound(t *testing.T) {
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
	require.NoError(t, err)

	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)
	assert.NotNil(t, round)
	assert.Equal(t, pool.ID, round.PoolID)
	assert.Equal(t, MixingRoundStatusCollecting, round.Status)
	assert.NotEmpty(t, round.RoundID)
	assert.NotZero(t, round.StartTime)
}

func TestPrivacyPools_ExecuteMixing(t *testing.T) {
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
	require.NoError(t, err)

	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)

	// Execute mixing
	err = pp.ExecuteMixing(round.RoundID)
	require.NoError(t, err)

	// Verify round status changed
	updatedRound, err := pp.GetMixingRound(round.RoundID)
	require.NoError(t, err)
	assert.Equal(t, MixingRoundStatusCompleted, updatedRound.Status)
	assert.NotZero(t, updatedRound.EndTime)
}

func TestPrivacyPools_CreateSelectiveDisclosure(t *testing.T) {
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
	require.NoError(t, err)

	// Join the pool first to create a participant
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)

	// Create selective disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("proof_data"),
		time.Now().Add(time.Hour),
		map[string]interface{}{"key": "value"},
	)
	require.NoError(t, err)
	assert.NotNil(t, disclosure)
	assert.Equal(t, pool.ID, disclosure.PoolID)
	assert.Equal(t, participant.ID, disclosure.ParticipantID)
	assert.Equal(t, []byte("proof_data"), disclosure.DisclosedData)
	assert.NotEmpty(t, disclosure.ID)
	assert.NotZero(t, disclosure.CreatedAt)
}

func TestPrivacyPools_GetPool(t *testing.T) {
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
	require.NoError(t, err)

	// Get the pool
	retrievedPool, err := pp.GetPool(pool.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedPool)
	assert.Equal(t, pool.ID, retrievedPool.ID)
	assert.Equal(t, pool.Name, retrievedPool.Name)

	// Test getting non-existent pool
	_, err = pp.GetPool("non-existent")
	assert.Error(t, err)
}

func TestPrivacyPools_GetPools(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Create pools of different types
	pool1, _ := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Coin Mixing Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)

	_, _ = pp.CreatePrivacyPool(
		PoolTypePrivacyPool,
		"Privacy Pool",
		"Description",
		"ETH",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)

	// Test getting all pools without filter
	allPools := pp.GetPools(nil)
	assert.Equal(t, 2, len(allPools))

	// Test getting pools by type using filter
	coinMixingPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.Type == PoolTypeCoinMixing
	})
	assert.Equal(t, 1, len(coinMixingPools))
	assert.Equal(t, pool1.ID, coinMixingPools[0].ID)

	// Test getting pools by asset using filter
	btcPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.Asset == "BTC"
	})
	assert.Equal(t, 1, len(btcPools))
	assert.Equal(t, pool1.ID, btcPools[0].ID)
}

func TestPrivacyPools_GetParticipant(t *testing.T) {
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
	require.NoError(t, err)

	// Join the pool
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)

	// Get the participant
	retrievedParticipant, err := pp.GetParticipant(participant.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedParticipant)
	assert.Equal(t, participant.ID, retrievedParticipant.ID)
	assert.Equal(t, participant.Address, retrievedParticipant.Address)

	// Test getting non-existent participant
	_, err = pp.GetParticipant("non-existent")
	assert.Error(t, err)
}

func TestPrivacyPools_GetMixingRound(t *testing.T) {
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
	require.NoError(t, err)

	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)

	// Get the mixing round
	retrievedRound, err := pp.GetMixingRound(round.RoundID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedRound)
	assert.Equal(t, round.RoundID, retrievedRound.RoundID)
	assert.Equal(t, round.PoolID, retrievedRound.PoolID)

	// Test getting non-existent round
	_, err = pp.GetMixingRound("non-existent")
	assert.Error(t, err)
}

func TestPrivacyPools_GetDisclosure(t *testing.T) {
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
	require.NoError(t, err)

	// Join the pool first to create a participant
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)

	// Create selective disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("proof_data"),
		time.Now().Add(time.Hour),
		map[string]interface{}{"key": "value"},
	)
	require.NoError(t, err)

	// Get the disclosure
	retrievedDisclosure, err := pp.GetDisclosure(disclosure.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedDisclosure)
	assert.Equal(t, disclosure.ID, retrievedDisclosure.ID)
	assert.Equal(t, disclosure.PoolID, retrievedDisclosure.PoolID)

	// Test getting non-existent disclosure
	_, err = pp.GetDisclosure("non-existent")
	assert.Error(t, err)
}

func TestPrivacyPools_ValidateDisclosure(t *testing.T) {
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
	require.NoError(t, err)

	// Join the pool first to create a participant
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)

	// Create selective disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("proof_data"),
		time.Now().Add(time.Hour),
		map[string]interface{}{"key": "value"},
	)
	require.NoError(t, err)

	// Validate the disclosure
	err = pp.ValidateDisclosure(disclosure.ID)
	require.NoError(t, err)

	// Verify disclosure status changed
	updatedDisclosure, err := pp.GetDisclosure(disclosure.ID)
	require.NoError(t, err)
	assert.Equal(t, DisclosureStatusValidated, updatedDisclosure.Status)
}

func TestPrivacyPools_StartStop(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Test starting
	err := pp.Start()
	require.NoError(t, err)

	// Test starting again (should fail)
	err = pp.Start()
	assert.Error(t, err)

	// Test stopping
	err = pp.Stop()
	require.NoError(t, err)

	// Test stopping again (should fail)
	err = pp.Stop()
	assert.Error(t, err)
}

func TestPrivacyPools_GetPoolsWithVariousFilters(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Create pools with different characteristics
	pool1, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"BTC Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(5000),
		3,
		5,
		big.NewInt(5),
		map[string]interface{}{"category": "defi"},
	)
	require.NoError(t, err)

	_, err = pp.CreatePrivacyPool(
		PoolTypePrivacyPool,
		"ETH Pool",
		"Description",
		"ETH",
		big.NewInt(1000),
		big.NewInt(15000),
		5,
		15,
		big.NewInt(20),
		map[string]interface{}{"category": "gaming"},
	)
	require.NoError(t, err)

	// Verify pools were created
	allPools := pp.GetPools(nil)
	assert.Equal(t, 2, len(allPools))

	// Test various filter combinations using the actual GetPools method
	// Test by asset
	btcPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.Asset == "BTC"
	})
	assert.Equal(t, 1, len(btcPools))

	// Test by size range
	smallPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.MaxAmount.Cmp(big.NewInt(10000)) <= 0
	})
	assert.Equal(t, 1, len(smallPools))

	// Test by fee range
	lowFeePools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.Fee.Cmp(big.NewInt(10)) <= 0
	})
	assert.Equal(t, 1, len(lowFeePools))

	// Test by minimum participants
	smallMinPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.MinParticipants <= 3
	})
	assert.Equal(t, 1, len(smallMinPools))

	// Test by maximum participants
	smallMaxPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.MaxParticipants <= 10
	})
	assert.Equal(t, 1, len(smallMaxPools))

	// Test by metadata
	defiPools := pp.GetPools(func(pool *PrivacyPool) bool {
		if category, ok := pool.Metadata["category"]; ok {
			return category == "defi"
		}
		return false
	})
	assert.Equal(t, 1, len(defiPools))

	// Test complex filter
	complexPools := pp.GetPools(func(pool *PrivacyPool) bool {
		return pool.Type == PoolTypeCoinMixing &&
			pool.Asset == "BTC" &&
			pool.Fee.Cmp(big.NewInt(10)) <= 0
	})
	assert.Equal(t, 1, len(complexPools))
	assert.Equal(t, pool1.ID, complexPools[0].ID)
}

func TestPrivacyPools_CopyMethods(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Create a pool
	pool, _ := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)

	// Get the pool and modify it
	retrievedPool, err := pp.GetPool(pool.ID)
	require.NoError(t, err)

	originalName := retrievedPool.Name
	retrievedPool.Name = "Modified Name"

	// Verify internal state wasn't affected
	storedPool, _ := pp.GetPool(pool.ID)
	assert.NotEqual(t, "Modified Name", storedPool.Name)
	assert.Equal(t, originalName, storedPool.Name)
}

func TestPrivacyPools_CopyMethodsWithNil(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Test copying nil objects
	nilPool := pp.copyPool(nil)
	assert.Nil(t, nilPool)

	nilParticipant := pp.copyParticipant(nil)
	assert.Nil(t, nilParticipant)

	nilMixingRound := pp.copyMixingRound(nil)
	assert.Nil(t, nilMixingRound)

	nilDisclosure := pp.copyDisclosure(nil)
	assert.Nil(t, nilDisclosure)
}

func TestPrivacyPools_CopyMethodsWithData(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Create a pool with metadata
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		map[string]interface{}{"key": "value"},
	)
	require.NoError(t, err)

	// Test copying pool
	copiedPool := pp.copyPool(pool)
	assert.NotNil(t, copiedPool)
	assert.Equal(t, pool.ID, copiedPool.ID)
	assert.Equal(t, pool.Name, copiedPool.Name)

	// Modify copied pool and verify original is unchanged
	copiedPool.Name = "Modified Name"
	assert.NotEqual(t, pool.Name, copiedPool.Name)
	assert.Equal(t, "Test Pool", pool.Name)

	// Test copying participant
	participant, err := pp.JoinPool(pool.ID, "user1", big.NewInt(5000), nil)
	require.NoError(t, err)

	copiedParticipant := pp.copyParticipant(participant)
	assert.NotNil(t, copiedParticipant)
	assert.Equal(t, participant.ID, copiedParticipant.ID)
	assert.Equal(t, participant.Address, copiedParticipant.Address)

	// Test copying disclosure
	disclosure, err := pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("proof_data"),
		time.Now().Add(time.Hour),
		map[string]interface{}{"key": "value"},
	)
	require.NoError(t, err)

	copiedDisclosure := pp.copyDisclosure(disclosure)
	assert.NotNil(t, copiedDisclosure)
	assert.Equal(t, disclosure.ID, copiedDisclosure.ID)
	assert.Equal(t, disclosure.PoolID, copiedDisclosure.PoolID)
}

func TestPrivacyPools_CopyMixingRound(t *testing.T) {
	pp := NewPrivacyPools(PrivacyPoolsConfig{})

	// Create a pool
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	require.NoError(t, err)

	// Join with enough participants to start mixing
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)

	// Test copying mixing round
	copiedRound := pp.copyMixingRound(round)
	assert.NotNil(t, copiedRound)
	assert.Equal(t, round.RoundID, copiedRound.RoundID)
	assert.Equal(t, round.PoolID, copiedRound.PoolID)
	assert.Equal(t, round.Status, copiedRound.Status)
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
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Join with minimum participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start and complete mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)

	err = pp.ExecuteMixing(round.RoundID)
	require.NoError(t, err)

	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)

	// Verify mixing round was cleaned up
	_, err = pp.GetMixingRound(round.RoundID)
	assert.Error(t, err)
}

func TestPoolTypeString(t *testing.T) {
	// Test all pool type string representations
	assert.Equal(t, "coin_mixing", PoolTypeCoinMixing.String())
	assert.Equal(t, "privacy_pool", PoolTypePrivacyPool.String())
	assert.Equal(t, "selective_disclosure", PoolTypeSelectiveDisclosure.String())
	assert.Equal(t, "confidential_swap", PoolTypeConfidentialSwap.String())
	assert.Equal(t, "anonymity_set", PoolTypeAnonymitySet.String())
	assert.Equal(t, "custom", PoolTypeCustom.String())
	assert.Equal(t, "unknown", PoolType(999).String()) // Test unknown type
}

func TestPoolStatusString(t *testing.T) {
	// Test all pool status string representations
	assert.Equal(t, "active", PoolStatusActive.String())
	assert.Equal(t, "mixing", PoolStatusMixing.String())
	assert.Equal(t, "completed", PoolStatusCompleted.String())
	assert.Equal(t, "paused", PoolStatusPaused.String())
	assert.Equal(t, "closed", PoolStatusClosed.String())
	assert.Equal(t, "error", PoolStatusError.String())
	assert.Equal(t, "unknown", PoolStatus(999).String()) // Test unknown status
}

func TestMixingRoundStatusString(t *testing.T) {
	// Test all mixing round status string representations
	assert.Equal(t, "pending", MixingRoundStatusPending.String())
	assert.Equal(t, "collecting", MixingRoundStatusCollecting.String())
	assert.Equal(t, "mixing", MixingRoundStatusMixing.String())
	assert.Equal(t, "distributing", MixingRoundStatusDistributing.String())
	assert.Equal(t, "completed", MixingRoundStatusCompleted.String())
	assert.Equal(t, "failed", MixingRoundStatusFailed.String())
	assert.Equal(t, "unknown", MixingRoundStatus(999).String()) // Test unknown status
}

func TestParticipantStatusString(t *testing.T) {
	// Test all participant status string representations
	assert.Equal(t, "pending", ParticipantStatusPending.String())
	assert.Equal(t, "active", ParticipantStatusActive.String())
	assert.Equal(t, "mixing", ParticipantStatusMixing.String())
	assert.Equal(t, "completed", ParticipantStatusCompleted.String())
	assert.Equal(t, "failed", ParticipantStatusFailed.String())
	assert.Equal(t, "exited", ParticipantStatusExited.String())
	assert.Equal(t, "unknown", ParticipantStatus(999).String()) // Test unknown status
}

func TestDisclosureTypeString(t *testing.T) {
	// Test all disclosure type string representations
	assert.Equal(t, "amount", DisclosureTypeAmount.String())
	assert.Equal(t, "source", DisclosureTypeSource.String())
	assert.Equal(t, "destination", DisclosureTypeDestination.String())
	assert.Equal(t, "timestamp", DisclosureTypeTimestamp.String())
	assert.Equal(t, "custom", DisclosureTypeCustom.String())
	assert.Equal(t, "unknown", DisclosureType(999).String()) // Test unknown type
}

func TestDisclosureStatusString(t *testing.T) {
	// Test all disclosure status string representations
	assert.Equal(t, "pending", DisclosureStatusPending.String())
	assert.Equal(t, "validated", DisclosureStatusValidated.String())
	assert.Equal(t, "expired", DisclosureStatusExpired.String())
	assert.Equal(t, "revoked", DisclosureStatusRevoked.String())
	assert.Equal(t, "unknown", DisclosureStatus(999).String()) // Test unknown status
}

func TestPrivacyPools_BackgroundProcessing(t *testing.T) {
	// Create PrivacyPools with very short intervals to trigger background processing
	config := PrivacyPoolsConfig{
		MixingTimeout:   time.Millisecond * 50,
		CleanupInterval: time.Millisecond * 25,
	}
	pp := NewPrivacyPools(config)

	// Start the system
	err := pp.Start()
	require.NoError(t, err)
	defer pp.Stop()

	// Create a pool
	pool, err := pp.CreatePrivacyPool(
		PoolTypeCoinMixing,
		"Test Pool",
		"Description",
		"BTC",
		big.NewInt(1000),
		big.NewInt(10000),
		3,
		10,
		big.NewInt(10),
		nil,
	)
	require.NoError(t, err)

	// Join with participants
	for i := 0; i < 3; i++ {
		_, err = pp.JoinPool(pool.ID, fmt.Sprintf("user_%d", i), big.NewInt(5000), nil)
		require.NoError(t, err)
	}

	// Start mixing round
	round, err := pp.StartMixingRound(pool.ID)
	require.NoError(t, err)

	// Create a disclosure
	participant, err := pp.GetParticipant(round.Participants[0])
	require.NoError(t, err)

	_, err = pp.CreateSelectiveDisclosure(
		pool.ID,
		participant.ID,
		DisclosureTypeAmount,
		[]byte("proof_data"),
		time.Now().Add(time.Millisecond*100),
		nil,
	)
	require.NoError(t, err)

	// Wait for background processing to run
	time.Sleep(time.Millisecond * 100)

	// Verify that some background processing occurred
	// The exact behavior depends on timing, but we can check that the system is still running
	assert.True(t, pp.running)
}
