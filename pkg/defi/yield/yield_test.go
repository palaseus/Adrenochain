package yield

import (
	"math/big"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// generateRandomAddress generates a random address for testing
func generateRandomAddress() engine.Address {
	addr := engine.Address{}
	for i := 0; i < len(addr); i++ {
		addr[i] = byte(i + 100)
	}
	return addr
}

func TestNewYieldFarm(t *testing.T) {
	owner := generateRandomAddress()
	rewardToken := generateRandomAddress()
	stakingToken := generateRandomAddress()
	rewardPerSecond := big.NewInt(100) // 100 tokens per second
	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)

	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		owner,
		rewardToken,
		stakingToken,
		rewardPerSecond,
		startTime,
		endTime,
	)

	if yf.FarmID != "test-yield-farming" {
		t.Errorf("expected farm ID 'test-yield-farming', got '%s'", yf.FarmID)
	}
	if yf.Name != "Test Yield Farming" {
		t.Errorf("expected name 'Test Yield Farming', got '%s'", yf.Name)
	}
	if yf.Symbol != "TEST" {
		t.Errorf("expected symbol 'TEST', got '%s'", yf.Symbol)
	}
	if yf.Decimals != 18 {
		t.Errorf("expected decimals 18, got %d", yf.Decimals)
	}
	if yf.Owner != owner {
		t.Errorf("expected owner %v, got %v", owner, yf.Owner)
	}
	if yf.RewardToken != rewardToken {
		t.Errorf("expected reward token %v, got %v", rewardToken, yf.RewardToken)
	}
	if yf.StakingToken != stakingToken {
		t.Errorf("expected staking token %v, got %v", stakingToken, yf.StakingToken)
	}
	if yf.RewardPerSecond.Cmp(rewardPerSecond) != 0 {
		t.Errorf("expected reward per second %v, got %v", rewardPerSecond, yf.RewardPerSecond)
	}
	if yf.StartTime != startTime {
		t.Errorf("expected start time %v, got %v", startTime, yf.StartTime)
	}
	if yf.EndTime != endTime {
		t.Errorf("expected end time %v, got %v", endTime, yf.EndTime)
	}
	if yf.Paused {
		t.Error("expected yield farming to not be paused")
	}
}

func TestYieldFarm_AddPool(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	stakingToken := generateRandomAddress()
	allocPoint := big.NewInt(100)

	// Test successful pool addition
	pid, err := yf.AddPool(stakingToken, allocPoint)
	if err != nil {
		t.Errorf("unexpected error adding pool: %v", err)
	}
	if pid != 0 {
		t.Errorf("expected pool ID 0, got %d", pid)
	}
	if yf.PoolCount != 1 {
		t.Errorf("expected pool count 1, got %d", yf.PoolCount)
	}

	// Test adding second pool
	stakingToken2 := generateRandomAddress()
	allocPoint2 := big.NewInt(200)
	pid2, err := yf.AddPool(stakingToken2, allocPoint2)
	if err != nil {
		t.Errorf("unexpected error adding second pool: %v", err)
	}
	if pid2 != 1 {
		t.Errorf("expected pool ID 1, got %d", pid2)
	}
	if yf.PoolCount != 2 {
		t.Errorf("expected pool count 2, got %d", yf.PoolCount)
	}

	// Test adding pool to paused farm
	yf.Paused = true
	_, err = yf.AddPool(generateRandomAddress(), big.NewInt(50))
	if err != ErrFarmPaused {
		t.Errorf("expected ErrFarmPaused, got %v", err)
	}
}

func TestYieldFarm_AddPool_Validation(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Test with nil staking token
	_, err := yf.AddPool(engine.Address{}, big.NewInt(100))
	if err == nil {
		t.Error("expected error with nil staking token")
	}

	// Test with nil allocation point
	_, err = yf.AddPool(generateRandomAddress(), nil)
	if err == nil {
		t.Error("expected error with nil allocation point")
	}

	// Test with zero allocation point
	_, err = yf.AddPool(generateRandomAddress(), big.NewInt(0))
	if err == nil {
		t.Error("expected error with zero allocation point")
	}

	// Test with negative allocation point
	_, err = yf.AddPool(generateRandomAddress(), big.NewInt(-100))
	if err == nil {
		t.Error("expected error with negative allocation point")
	}
}

func TestYieldFarm_Deposit(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()
	amount := big.NewInt(1000)

	// Add a pool first
	pid, err := yf.AddPool(stakingToken, big.NewInt(100))
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}

	// Test successful deposit
	err = yf.Deposit(user, pid, amount, 1, engine.Hash{})
	if err != nil {
		t.Errorf("unexpected error depositing: %v", err)
	}

	// Verify user was created
	userInfo := yf.GetUserInfo(user)
	if userInfo == nil {
		t.Error("expected user info to be created")
	}
	if userInfo.TotalStaked.Cmp(amount) != 0 {
		t.Errorf("expected total staked %v, got %v", amount, userInfo.TotalStaked)
	}

	// Verify pool was updated
	pool := yf.GetPoolInfo(pid)
	if pool == nil {
		t.Error("expected pool info")
	}
	if pool.TotalStaked.Cmp(amount) != 0 {
		t.Errorf("expected pool total staked %v, got %v", amount, pool.TotalStaked)
	}
}

func TestYieldFarm_Deposit_Validation(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()
	amount := big.NewInt(1000)

	// Test deposit to non-existent pool
	err := yf.Deposit(user, 999, amount, 1, engine.Hash{})
	if err == nil {
		t.Error("expected error depositing to non-existent pool")
	}

	// Test deposit with nil amount
	err = yf.Deposit(user, 0, nil, 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with nil amount")
	}

	// Test deposit with zero amount
	err = yf.Deposit(user, 0, big.NewInt(0), 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with zero amount")
	}

	// Test deposit with negative amount
	err = yf.Deposit(user, 0, big.NewInt(-100), 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with negative amount")
	}

	// Test deposit to paused farm
	yf.Paused = true
	pid, _ := yf.AddPool(stakingToken, big.NewInt(100))
	err = yf.Deposit(user, pid, amount, 1, engine.Hash{})
	if err != ErrFarmPaused {
		t.Errorf("expected ErrFarmPaused, got %v", err)
	}
}

func TestYieldFarm_Withdraw(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()
	depositAmount := big.NewInt(1000)
	withdrawAmount := big.NewInt(500)

	// Add a pool and deposit
	pid, err := yf.AddPool(stakingToken, big.NewInt(100))
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}

	err = yf.Deposit(user, pid, depositAmount, 1, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}

	// Test successful withdrawal
	err = yf.Withdraw(user, pid, withdrawAmount, 2, engine.Hash{})
	if err != nil {
		t.Errorf("unexpected error withdrawing: %v", err)
	}

	// Verify user staked amount was reduced
	userInfo := yf.GetUserInfo(user)
	if userInfo.TotalStaked.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("expected total staked 500, got %v", userInfo.TotalStaked)
	}

	// Verify pool total staked was reduced
	pool := yf.GetPoolInfo(pid)
	if pool.TotalStaked.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("expected pool total staked 500, got %v", pool.TotalStaked)
	}
}

func TestYieldFarm_Withdraw_Validation(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()

	// Test withdraw from non-existent pool
	err := yf.Withdraw(user, 999, big.NewInt(100), 1, engine.Hash{})
	if err == nil {
		t.Error("expected error withdrawing from non-existent pool")
	}

	// Test withdraw with nil amount
	err = yf.Withdraw(user, 0, nil, 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with nil amount")
	}

	// Test withdraw with zero amount
	err = yf.Withdraw(user, 0, big.NewInt(0), 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with zero amount")
	}

	// Test withdraw with negative amount
	err = yf.Withdraw(user, 0, big.NewInt(-100), 1, engine.Hash{})
	if err == nil {
		t.Error("expected error with negative amount")
	}

	// Test withdraw from paused farm
	yf.Paused = true
	pid, _ := yf.AddPool(stakingToken, big.NewInt(100))
	err = yf.Withdraw(user, pid, big.NewInt(100), 1, engine.Hash{})
	if err != ErrFarmPaused {
		t.Errorf("expected ErrFarmPaused, got %v", err)
	}
}

func TestYieldFarm_Harvest(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()
	amount := big.NewInt(1000)

	// Add a pool and deposit
	pid, err := yf.AddPool(stakingToken, big.NewInt(100))
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}

	err = yf.Deposit(user, pid, amount, 1, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}

	// Test successful harvest
	rewards, err := yf.Harvest(user, pid, 2, engine.Hash{})
	if err != nil {
		t.Errorf("unexpected error harvesting: %v", err)
	}
	if rewards == nil {
		t.Error("expected non-nil rewards")
	}

	// Verify harvest event was recorded
	if len(yf.HarvestEvents) != 1 {
		t.Errorf("expected 1 harvest event, got %d", len(yf.HarvestEvents))
	}
}

func TestYieldFarm_Harvest_Validation(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()

	// Test harvest from non-existent pool
	_, err := yf.Harvest(user, 999, 1, engine.Hash{})
	if err == nil {
		t.Error("expected error harvesting from non-existent pool")
	}

	// Test harvest from paused farm
	yf.Paused = true
	pid, _ := yf.AddPool(generateRandomAddress(), big.NewInt(100))
	_, err = yf.Harvest(user, pid, 1, engine.Hash{})
	if err != ErrFarmPaused {
		t.Errorf("expected ErrFarmPaused, got %v", err)
	}
}

func TestYieldFarm_GetPendingRewards(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()
	stakingToken := generateRandomAddress()
	amount := big.NewInt(1000)

	// Add a pool and deposit
	pid, err := yf.AddPool(stakingToken, big.NewInt(100))
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}

	err = yf.Deposit(user, pid, amount, 1, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}

	// Get pending rewards
	rewards, err := yf.GetPendingRewards(user, pid)
	if err != nil {
		t.Errorf("unexpected error getting pending rewards: %v", err)
	}
	if rewards == nil {
		t.Error("expected non-nil rewards")
	}
}

func TestYieldFarm_GetPendingRewards_NonExistentPool(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()

	// Get pending rewards from non-existent pool
	rewards, err := yf.GetPendingRewards(user, 999)
	if err == nil {
		t.Error("expected error getting pending rewards from non-existent pool")
	}
	if rewards != nil {
		t.Error("expected nil rewards for non-existent pool")
	}
}

func TestYieldFarm_GetUserInfo_NonExistent(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	user := generateRandomAddress()

	// Get user for non-existent user
	userInfo := yf.GetUserInfo(user)

	if userInfo != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestYieldFarm_GetPoolInfo(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	stakingToken := generateRandomAddress()
	allocPoint := big.NewInt(100)

	// Add a pool
	pid, err := yf.AddPool(stakingToken, allocPoint)
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}

	// Get pool info
	pool := yf.GetPoolInfo(pid)
	if pool == nil {
		t.Error("expected pool info")
	}
	if pool.PID != pid {
		t.Errorf("expected pool ID %d, got %d", pid, pool.PID)
	}
	if pool.StakingToken != stakingToken {
		t.Errorf("expected staking token %v, got %v", stakingToken, pool.StakingToken)
	}
	if pool.AllocPoint.Cmp(allocPoint) != 0 {
		t.Errorf("expected allocation point %v, got %v", allocPoint, pool.AllocPoint)
	}
}

func TestYieldFarm_GetPoolInfo_NonExistent(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Get info for non-existent pool
	pool := yf.GetPoolInfo(999)
	if pool != nil {
		t.Error("expected nil for non-existent pool")
	}
}

func TestYieldFarm_GetFarmStats(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Get initial stats
	poolCount, userCount, totalStaked, totalRewards := yf.GetFarmStats()
	if poolCount != 0 {
		t.Errorf("expected pool count 0, got %d", poolCount)
	}
	if userCount != 0 {
		t.Errorf("expected user count 0, got %d", userCount)
	}
	if totalStaked.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected total staked 0, got %v", totalStaked)
	}
	if totalRewards.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected total rewards 0, got %v", totalRewards)
	}

	// Add a pool and some activity
	stakingToken := generateRandomAddress()
	pid, _ := yf.AddPool(stakingToken, big.NewInt(100))
	user := generateRandomAddress()
	yf.Deposit(user, pid, big.NewInt(1000), 1, engine.Hash{})

	// Get updated stats
	poolCount, userCount, totalStaked, totalRewards = yf.GetFarmStats()
	if poolCount != 1 {
		t.Errorf("expected pool count 1, got %d", poolCount)
	}
	if userCount != 1 {
		t.Errorf("expected user count 1, got %d", userCount)
	}
	if totalStaked.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("expected total staked 1000, got %v", totalStaked)
	}
}

func TestYieldFarm_Pause_Unpause(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Test initial state
	if yf.Paused {
		t.Error("expected farm to not be paused initially")
	}

	// Test pause
	err := yf.Pause()
	if err != nil {
		t.Errorf("unexpected error pausing: %v", err)
	}
	if !yf.Paused {
		t.Error("expected farm to be paused")
	}

	// Test pause when already paused
	err = yf.Pause()
	if err == nil {
		t.Error("expected error when pausing already paused farm")
	}

	// Test unpause
	err = yf.Unpause()
	if err != nil {
		t.Errorf("unexpected error unpausing: %v", err)
	}
	if yf.Paused {
		t.Error("expected farm to not be paused")
	}

	// Test unpause when not paused
	err = yf.Unpause()
	if err == nil {
		t.Error("expected error when unpausing non-paused farm")
	}
}

func TestYieldFarm_Concurrency(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Test concurrent access to farm data
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Access farm data concurrently
			if yf.FarmID != "test-yield-farming" {
				t.Errorf("invalid farm ID: %s", yf.FarmID)
			}
			if yf.Name != "Test Yield Farming" {
				t.Errorf("invalid name: %s", yf.Name)
			}
			if yf.Symbol != "TEST" {
				t.Errorf("invalid symbol: %s", yf.Symbol)
			}
			
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestYieldFarm_EdgeCases(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Test with very large numbers
	largeAmount := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	largeAllocPoint := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)

	// This should not panic
	pid, err := yf.AddPool(generateRandomAddress(), largeAllocPoint)
	if err != nil {
		t.Errorf("unexpected error with large allocation point: %v", err)
	}

	// Test deposit with large amount
	err = yf.Deposit(generateRandomAddress(), pid, largeAmount, 1, engine.Hash{})
	if err != nil {
		t.Errorf("unexpected error with large amount: %v", err)
	}
}

func TestYieldFarm_EventRecording(t *testing.T) {
	yf := NewYieldFarm(
		"test-yield-farming",
		"Test Yield Farming",
		"TEST",
		18,
		generateRandomAddress(),
		generateRandomAddress(),
		generateRandomAddress(),
		big.NewInt(100),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)

	// Test that events are recorded
	stakingToken := generateRandomAddress()
	pid, err := yf.AddPool(stakingToken, big.NewInt(100))
	if err != nil {
		t.Fatalf("failed to add pool: %v", err)
	}
	if len(yf.AddPoolEvents) != 1 {
		t.Errorf("expected 1 add pool event, got %d", len(yf.AddPoolEvents))
	}

	user := generateRandomAddress()
	err = yf.Deposit(user, pid, big.NewInt(1000), 1, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to deposit: %v", err)
	}
	if len(yf.DepositEvents) != 1 {
		t.Errorf("expected 1 deposit event, got %d", len(yf.DepositEvents))
	}

	err = yf.Withdraw(user, pid, big.NewInt(500), 2, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to withdraw: %v", err)
	}
	if len(yf.WithdrawEvents) != 1 {
		t.Errorf("expected 1 withdraw event, got %d", len(yf.WithdrawEvents))
	}

	_, err = yf.Harvest(user, pid, 3, engine.Hash{})
	if err != nil {
		t.Fatalf("failed to harvest: %v", err)
	}
	if len(yf.HarvestEvents) != 1 {
		t.Errorf("expected 1 harvest event, got %d", len(yf.HarvestEvents))
	}
}
