package yield

import (
	"math/big"
	"testing"
	"time"
)

func TestNewYieldManager(t *testing.T) {
	ym := NewYieldManager()

	if ym == nil {
		t.Fatal("NewYieldManager returned nil")
	}

	if ym.Farms == nil {
		t.Error("Farms map not initialized")
	}
	if ym.Positions == nil {
		t.Error("Positions map not initialized")
	}
	if ym.Calculator == nil {
		t.Error("Calculator not initialized")
	}

	// Check timestamps are set
	if ym.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if ym.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewFarm(t *testing.T) {
	apy := big.NewFloat(0.15)         // 15%
	apr := big.NewFloat(0.14)         // 14%
	tvl := big.NewFloat(1000000)      // 1M
	minStake := big.NewFloat(100)     // 100
	maxStake := big.NewFloat(10000)   // 10K
	lockPeriod := 30 * 24 * time.Hour // 30 days
	rewardTokens := []string{"REWARD", "BONUS"}
	stakeTokens := []string{"USDC", "USDT"}

	farm, err := NewFarm("farm1", "Test Farm", "Test Description", "Uniswap", "Ethereum", LiquidityProviding, Medium, apy, apr, tvl, minStake, maxStake, lockPeriod, rewardTokens, stakeTokens)
	if err != nil {
		t.Fatalf("NewFarm failed: %v", err)
	}

	if farm.ID != "farm1" {
		t.Errorf("Expected ID 'farm1', got '%s'", farm.ID)
	}
	if farm.Name != "Test Farm" {
		t.Errorf("Expected name 'Test Farm', got '%s'", farm.Name)
	}
	if farm.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", farm.Description)
	}
	if farm.Protocol != "Uniswap" {
		t.Errorf("Expected protocol 'Uniswap', got '%s'", farm.Protocol)
	}
	if farm.Chain != "Ethereum" {
		t.Errorf("Expected chain 'Ethereum', got '%s'", farm.Chain)
	}
	if farm.Strategy != LiquidityProviding {
		t.Errorf("Expected strategy LiquidityProviding, got %d", farm.Strategy)
	}
	if farm.RiskLevel != Medium {
		t.Errorf("Expected risk level Medium, got %d", farm.RiskLevel)
	}
	if farm.APY.Cmp(apy) != 0 {
		t.Error("APY not set correctly")
	}
	if farm.APR.Cmp(apr) != 0 {
		t.Error("APR not set correctly")
	}
	if farm.TVL.Cmp(tvl) != 0 {
		t.Error("TVL not set correctly")
	}
	if farm.MinStake.Cmp(minStake) != 0 {
		t.Error("MinStake not set correctly")
	}
	if farm.MaxStake.Cmp(maxStake) != 0 {
		t.Error("MaxStake not set correctly")
	}
	if farm.LockPeriod != lockPeriod {
		t.Errorf("Expected lock period %v, got %v", lockPeriod, farm.LockPeriod)
	}
	if len(farm.RewardTokens) != 2 {
		t.Errorf("Expected 2 reward tokens, got %d", len(farm.RewardTokens))
	}
	if len(farm.StakeTokens) != 2 {
		t.Errorf("Expected 2 stake tokens, got %d", len(farm.StakeTokens))
	}
	if !farm.IsActive {
		t.Error("Expected farm to be active")
	}
	if farm.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if farm.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewFarmValidation(t *testing.T) {
	validParams := map[string]*big.Float{
		"apy":      big.NewFloat(0.15),
		"apr":      big.NewFloat(0.14),
		"tvl":      big.NewFloat(1000000),
		"minStake": big.NewFloat(100),
		"maxStake": big.NewFloat(10000),
	}
	validLockPeriod := 30 * 24 * time.Hour
	validTokens := []string{"TOKEN"}

	// Test empty ID
	_, err := NewFarm("", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, validParams["apy"], validParams["apr"], validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test empty name
	_, err = NewFarm("farm1", "", "Description", "Protocol", "Chain", LiquidityProviding, Medium, validParams["apy"], validParams["apr"], validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test empty protocol
	_, err = NewFarm("farm1", "Test", "Description", "", "Chain", LiquidityProviding, Medium, validParams["apy"], validParams["apr"], validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for empty protocol")
	}

	// Test empty chain
	_, err = NewFarm("farm1", "Test", "Description", "Protocol", "", LiquidityProviding, Medium, validParams["apy"], validParams["apr"], validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for empty chain")
	}

	// Test negative APY
	_, err = NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(-0.15), validParams["apr"], validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for negative APY")
	}

	// Test negative APR
	_, err = NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, validParams["apy"], big.NewFloat(-0.14), validParams["tvl"], validParams["minStake"], validParams["maxStake"], validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for negative APR")
	}

	// Test min stake greater than max stake
	_, err = NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, validParams["apy"], validParams["apr"], validParams["tvl"], big.NewFloat(10000), big.NewFloat(100), validLockPeriod, validTokens, validTokens)
	if err == nil {
		t.Error("Expected error for min stake greater than max stake")
	}
}

func TestNewPosition(t *testing.T) {
	stakedAmount := big.NewFloat(1000)

	position, err := NewPosition("pos1", "farm1", "user1", stakedAmount)
	if err != nil {
		t.Fatalf("NewPosition failed: %v", err)
	}

	if position.ID != "pos1" {
		t.Errorf("Expected ID 'pos1', got '%s'", position.ID)
	}
	if position.FarmID != "farm1" {
		t.Errorf("Expected FarmID 'farm1', got '%s'", position.FarmID)
	}
	if position.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", position.UserID)
	}
	if position.StakedAmount.Cmp(stakedAmount) != 0 {
		t.Error("StakedAmount not set correctly")
	}
	if position.RewardsEarned.Sign() != 0 {
		t.Error("Expected initial rewards to be 0")
	}
	if position.StartTime.IsZero() {
		t.Error("StartTime not set")
	}
	if position.LastClaim.IsZero() {
		t.Error("LastClaim not set")
	}
	if !position.IsActive {
		t.Error("Expected position to be active")
	}
	if position.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if position.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewPositionValidation(t *testing.T) {
	validStakedAmount := big.NewFloat(1000)

	// Test empty ID
	_, err := NewPosition("", "farm1", "user1", validStakedAmount)
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test empty farm ID
	_, err = NewPosition("pos1", "", "user1", validStakedAmount)
	if err == nil {
		t.Error("Expected error for empty farm ID")
	}

	// Test empty user ID
	_, err = NewPosition("pos1", "farm1", "", validStakedAmount)
	if err == nil {
		t.Error("Expected error for empty user ID")
	}

	// Test nil staked amount
	_, err = NewPosition("pos1", "farm1", "user1", nil)
	if err == nil {
		t.Error("Expected error for nil staked amount")
	}

	// Test negative staked amount
	_, err = NewPosition("pos1", "farm1", "user1", big.NewFloat(-1000))
	if err == nil {
		t.Error("Expected error for negative staked amount")
	}

	// Test zero staked amount
	_, err = NewPosition("pos1", "farm1", "user1", big.NewFloat(0))
	if err == nil {
		t.Error("Expected error for zero staked amount")
	}
}

func TestYieldManagerAddFarm(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})

	err := ym.AddFarm(farm)
	if err != nil {
		t.Fatalf("AddFarm failed: %v", err)
	}

	if _, exists := ym.Farms["farm1"]; !exists {
		t.Error("Farm not added to manager")
	}

	// Test adding duplicate farm
	err = ym.AddFarm(farm)
	if err == nil {
		t.Error("Expected error for duplicate farm")
	}

	// Test adding nil farm
	err = ym.AddFarm(nil)
	if err == nil {
		t.Error("Expected error for nil farm")
	}
}

func TestYieldManagerRemoveFarm(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	err := ym.RemoveFarm("farm1")
	if err != nil {
		t.Fatalf("RemoveFarm failed: %v", err)
	}

	if _, exists := ym.Farms["farm1"]; exists {
		t.Error("Farm not removed from manager")
	}

	// Test removing non-existent farm
	err = ym.RemoveFarm("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent farm")
	}

	// Test removing with empty ID
	err = ym.RemoveFarm("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestYieldManagerStartFarming(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))

	err := ym.StartFarming(position)
	if err != nil {
		t.Fatalf("StartFarming failed: %v", err)
	}

	if _, exists := ym.Positions["pos1"]; !exists {
		t.Error("Position not added to manager")
	}

	// Test starting farming with non-existent farm
	position2, _ := NewPosition("pos2", "non_existent", "user1", big.NewFloat(1000))
	err = ym.StartFarming(position2)
	if err == nil {
		t.Error("Expected error for non-existent farm")
	}

	// Test starting farming with inactive farm
	farm.IsActive = false
	position3, _ := NewPosition("pos3", "farm1", "user1", big.NewFloat(1000))
	err = ym.StartFarming(position3)
	if err == nil {
		t.Error("Expected error for inactive farm")
	}

	// Test starting farming with stake below minimum
	farm.IsActive = true
	position4, _ := NewPosition("pos4", "farm1", "user1", big.NewFloat(50)) // Below min stake
	err = ym.StartFarming(position4)
	if err == nil {
		t.Error("Expected error for stake below minimum")
	}

	// Test starting farming with stake above maximum
	position5, _ := NewPosition("pos5", "farm1", "user1", big.NewFloat(20000)) // Above max stake
	err = ym.StartFarming(position5)
	if err == nil {
		t.Error("Expected error for stake above maximum")
	}

	// Test starting farming with nil position
	err = ym.StartFarming(nil)
	if err == nil {
		t.Error("Expected error for nil position")
	}
}

func TestYieldManagerStopFarming(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	ym.StartFarming(position)

	err := ym.StopFarming("pos1")
	if err != nil {
		t.Fatalf("StopFarming failed: %v", err)
	}

	if ym.Positions["pos1"].IsActive {
		t.Error("Position not deactivated")
	}

	// Test stopping non-existent position
	err = ym.StopFarming("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}

	// Test stopping already inactive position
	err = ym.StopFarming("pos1")
	if err == nil {
		t.Error("Expected error for already inactive position")
	}

	// Test stopping with empty ID
	err = ym.StopFarming("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestYieldManagerClaimRewards(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	ym.StartFarming(position)

	// Wait a bit to simulate time passing
	time.Sleep(10 * time.Millisecond)

	rewards, err := ym.ClaimRewards("pos1")
	if err != nil {
		t.Fatalf("ClaimRewards failed: %v", err)
	}

	if rewards.Sign() <= 0 {
		t.Error("Expected positive rewards")
	}

	// Test claiming from non-existent position
	_, err = ym.ClaimRewards("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}

	// Test claiming from inactive position
	ym.StopFarming("pos1")
	_, err = ym.ClaimRewards("pos1")
	if err == nil {
		t.Error("Expected error for inactive position")
	}

	// Test claiming with empty ID
	_, err = ym.ClaimRewards("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
}

func TestYieldCalculatorCalculateRewards(t *testing.T) {
	calculator := &YieldCalculator{}

	// Create a position with some time passed
	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	position.LastClaim = time.Now().Add(-1 * time.Hour) // 1 hour ago

	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})

	rewards := calculator.CalculateRewards(position, farm)

	if rewards.Sign() <= 0 {
		t.Error("Expected positive rewards")
	}

	// Test with no time passed
	position.LastClaim = time.Now()
	rewards = calculator.CalculateRewards(position, farm)
	// Allow for minimal time differences due to test execution
	if rewards.Sign() > 0 && rewards.Cmp(big.NewFloat(0.0001)) > 0 {
		t.Errorf("Expected minimal rewards for no time passed, got %v", rewards)
	}
}

func TestYieldManagerGetters(t *testing.T) {
	ym := NewYieldManager()

	// Test GetFarm
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	retrievedFarm, err := ym.GetFarm("farm1")
	if err != nil {
		t.Fatalf("GetFarm failed: %v", err)
	}
	if retrievedFarm != farm {
		t.Error("GetFarm returned different farm")
	}

	// Test GetFarm with non-existent ID
	_, err = ym.GetFarm("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent farm")
	}

	// Test GetFarm with empty ID
	_, err = ym.GetFarm("")
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test GetPosition
	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	ym.StartFarming(position)

	retrievedPosition, err := ym.GetPosition("pos1")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if retrievedPosition != position {
		t.Error("GetPosition returned different position")
	}

	// Test GetPosition with non-existent ID
	_, err = ym.GetPosition("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent position")
	}

	// Test GetFarmsByStrategy
	farms := ym.GetFarmsByStrategy(LiquidityProviding)
	if len(farms) != 1 {
		t.Errorf("Expected 1 farm for strategy, got %d", len(farms))
	}

	// Test GetFarmsByRiskLevel
	farms = ym.GetFarmsByRiskLevel(Medium)
	if len(farms) != 1 {
		t.Errorf("Expected 1 farm for risk level, got %d", len(farms))
	}

	// Test GetPositionsByUser
	positions := ym.GetPositionsByUser("user1")
	if len(positions) != 1 {
		t.Errorf("Expected 1 position for user, got %d", len(positions))
	}

	// Test GetActivePositions
	activePositions := ym.GetActivePositions()
	if len(activePositions) != 1 {
		t.Errorf("Expected 1 active position, got %d", len(activePositions))
	}
}

func TestYieldManagerUpdateFarmMetrics(t *testing.T) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	// Test UpdateFarmAPY
	newAPY := big.NewFloat(0.20)
	err := ym.UpdateFarmAPY("farm1", newAPY)
	if err != nil {
		t.Fatalf("UpdateFarmAPY failed: %v", err)
	}

	if farm.APY.Cmp(newAPY) != 0 {
		t.Error("Farm APY not updated")
	}

	// Test UpdateFarmAPR
	newAPR := big.NewFloat(0.19)
	err = ym.UpdateFarmAPR("farm1", newAPR)
	if err != nil {
		t.Fatalf("UpdateFarmAPR failed: %v", err)
	}

	if farm.APR.Cmp(newAPR) != 0 {
		t.Error("Farm APR not updated")
	}

	// Test UpdateFarmTVL
	newTVL := big.NewFloat(2000000)
	err = ym.UpdateFarmTVL("farm1", newTVL)
	if err != nil {
		t.Fatalf("UpdateFarmTVL failed: %v", err)
	}

	if farm.TVL.Cmp(newTVL) != 0 {
		t.Error("Farm TVL not updated")
	}

	// Test updating non-existent farm
	err = ym.UpdateFarmAPY("non_existent", newAPY)
	if err == nil {
		t.Error("Expected error for non-existent farm")
	}

	// Test updating with invalid values
	err = ym.UpdateFarmAPY("farm1", big.NewFloat(-0.20))
	if err == nil {
		t.Error("Expected error for negative APY")
	}
}

func TestYieldManagerAggregateMetrics(t *testing.T) {
	ym := NewYieldManager()

	// Add multiple farms
	farm1, _ := NewFarm("farm1", "Test1", "Description1", "Protocol1", "Chain1", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	farm2, _ := NewFarm("farm2", "Test2", "Description2", "Protocol2", "Chain2", Staking, High, big.NewFloat(0.25), big.NewFloat(0.24), big.NewFloat(2000000), big.NewFloat(200), big.NewFloat(20000), 60*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})

	ym.AddFarm(farm1)
	ym.AddFarm(farm2)

	// Add positions
	position1, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	position2, _ := NewPosition("pos2", "farm2", "user2", big.NewFloat(2000))

	ym.StartFarming(position1)
	ym.StartFarming(position2)

	// Test GetTotalStaked
	totalStaked := ym.GetTotalStaked()
	expectedTotal := big.NewFloat(3000)
	if totalStaked.Cmp(expectedTotal) != 0 {
		t.Errorf("Expected total staked %v, got %v", expectedTotal, totalStaked)
	}

	// Test GetTotalRewards
	totalRewards := ym.GetTotalRewards()
	if totalRewards.Sign() != 0 {
		t.Error("Expected initial total rewards to be 0")
	}

	// Test GetAverageAPY
	averageAPY := ym.GetAverageAPY()
	expectedAverage := big.NewFloat(0.20) // (0.15 + 0.25) / 2
	if averageAPY.Cmp(expectedAverage) != 0 {
		t.Errorf("Expected average APY %v, got %v", expectedAverage, averageAPY)
	}

	// Test GetTopPerformingFarms
	topFarms := ym.GetTopPerformingFarms(2)
	if len(topFarms) != 2 {
		t.Errorf("Expected 2 top farms, got %d", len(topFarms))
	}

	// Check they're sorted by APY (descending)
	if topFarms[0].APY.Cmp(topFarms[1].APY) <= 0 {
		t.Error("Top farms not sorted by APY")
	}

	// Test GetFarmsByChain
	ethereumFarms := ym.GetFarmsByChain("Ethereum")
	if len(ethereumFarms) != 0 {
		t.Error("Expected 0 farms for Ethereum chain")
	}

	chain1Farms := ym.GetFarmsByChain("Chain1")
	if len(chain1Farms) != 1 {
		t.Errorf("Expected 1 farm for Chain1, got %d", len(chain1Farms))
	}

	// Test GetFarmsByProtocol
	protocol1Farms := ym.GetFarmsByProtocol("Protocol1")
	if len(protocol1Farms) != 1 {
		t.Errorf("Expected 1 farm for Protocol1, got %d", len(protocol1Farms))
	}
}

// Benchmark tests for performance
func BenchmarkYieldManagerStartFarming(b *testing.B) {
	ym := NewYieldManager()
	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
	ym.AddFarm(farm)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		position, _ := NewPosition("pos"+string(rune(i)), "farm1", "user1", big.NewFloat(1000))
		ym.StartFarming(position)
	}
}

func BenchmarkYieldCalculatorCalculateRewards(b *testing.B) {
	calculator := &YieldCalculator{}
	position, _ := NewPosition("pos1", "farm1", "user1", big.NewFloat(1000))
	position.LastClaim = time.Now().Add(-1 * time.Hour)

	farm, _ := NewFarm("farm1", "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, big.NewFloat(0.15), big.NewFloat(0.14), big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculator.CalculateRewards(position, farm)
	}
}

func BenchmarkYieldManagerGetTopPerformingFarms(b *testing.B) {
	ym := NewYieldManager()

	// Add multiple farms
	for i := 0; i < 100; i++ {
		apy := big.NewFloat(0.10 + float64(i)*0.01)
		apr := big.NewFloat(0.09 + float64(i)*0.01)
		farm, _ := NewFarm("farm"+string(rune(i)), "Test", "Description", "Protocol", "Chain", LiquidityProviding, Medium, apy, apr, big.NewFloat(1000000), big.NewFloat(100), big.NewFloat(10000), 30*24*time.Hour, []string{"TOKEN"}, []string{"TOKEN"})
		ym.AddFarm(farm)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ym.GetTopPerformingFarms(10)
	}
}
