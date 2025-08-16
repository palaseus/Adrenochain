package yield

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
