package defi

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

// TestNewCrossChainDeFi tests DeFi system creation
func TestNewCrossChainDeFi(t *testing.T) {
	config := DeFiConfig{
		MaxNetworks:       5,
		MaxLendingPools:   50,
		MaxYieldFarms:     25,
		MaxLiquidityPools: 50,
		EnableCrossChain:  true,
		SecurityLevel:     SecurityLevelHigh,
	}

	defi := NewCrossChainDeFi(config)

	if defi == nil {
		t.Fatal("Expected DeFi system to be created")
	}

	if defi.config.MaxNetworks != 5 {
		t.Errorf("Expected max networks 5, got %d", defi.config.MaxNetworks)
	}

	if defi.config.MaxLendingPools != 50 {
		t.Errorf("Expected max lending pools 50, got %d", defi.config.MaxLendingPools)
	}

	if defi.AssetRegistry == nil {
		t.Error("Expected asset registry to be created")
	}

	if defi.PriceOracle == nil {
		t.Error("Expected price oracle to be created")
	}

	if defi.RiskManager == nil {
		t.Error("Expected risk manager to be created")
	}
}

// TestNewBlockchainNetwork tests blockchain network creation
func TestNewBlockchainNetwork(t *testing.T) {
	config := NetworkConfig{
		MaxGasLimit:        30000000,
		MinConfirmations:   12,
		EnableFastFinality: true,
		SecurityLevel:      SecurityLevelHigh,
		CrossChainEnabled:  true,
	}

	network := NewBlockchainNetwork("Ethereum", 1, "ETH", config)

	if network == nil {
		t.Fatal("Expected network to be created")
	}

	if network.Name != "Ethereum" {
		t.Errorf("Expected name Ethereum, got %s", network.Name)
	}

	if network.ChainID != 1 {
		t.Errorf("Expected chain ID 1, got %d", network.ChainID)
	}

	if network.NativeToken != "ETH" {
		t.Errorf("Expected native token ETH, got %s", network.NativeToken)
	}

	if network.Status != NetworkStatusActive {
		t.Errorf("Expected status %d, got %d", NetworkStatusActive, network.Status)
	}

	if network.BlockTime != time.Second*15 {
		t.Errorf("Expected block time 15s, got %v", network.BlockTime)
	}
}

// TestAddNetwork tests adding networks to the DeFi system
func TestAddNetwork(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks: 3,
	})

	// Add networks
	networks := []*BlockchainNetwork{
		NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{}),
		NewBlockchainNetwork("Polygon", 137, "MATIC", NetworkConfig{}),
		NewBlockchainNetwork("BSC", 56, "BNB", NetworkConfig{}),
	}

	for i, network := range networks {
		err := defi.AddNetwork(network)
		if err != nil {
			t.Errorf("Failed to add network %d: %v", i, err)
		}
	}

	if len(defi.Networks) != 3 {
		t.Errorf("Expected 3 networks, got %d", len(defi.Networks))
	}

	if defi.metrics.TotalNetworks != 3 {
		t.Errorf("Expected 3 total networks in metrics, got %d", defi.metrics.TotalNetworks)
	}
}

// TestAddNetworkValidation tests network validation
func TestAddNetworkValidation(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks: 2,
	})

	// Add two networks
	network1 := NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{})
	network2 := NewBlockchainNetwork("Polygon", 137, "MATIC", NetworkConfig{})

	defi.AddNetwork(network1)
	defi.AddNetwork(network2)

	// Try to add a third network
	network3 := NewBlockchainNetwork("BSC", 56, "BNB", NetworkConfig{})
	err := defi.AddNetwork(network3)

	if err == nil {
		t.Error("Expected error when maximum networks reached")
	}

	if len(defi.Networks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(defi.Networks))
	}
}

// TestCreateLendingPool tests lending pool creation
func TestCreateLendingPool(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks:     5,
		MaxLendingPools: 10,
	})

	// Add a network first
	network := NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{})
	defi.AddNetwork(network)

	// Create lending pool config
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:           0.02,
			Multiplier:         0.1,
			JumpMultiplier:     0.5,
			OptimalUtilization: 0.8,
		},
		EnableFlashLoans:   true,
		MaxFlashLoanAmount: new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 1000 ETH
		SecurityLevel:      SecurityLevelHigh,
	}

	pool, err := defi.CreateLendingPool(network.ID, "USDC", config)
	if err != nil {
		t.Errorf("Failed to create lending pool: %v", err)
	}

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	if pool.NetworkID != network.ID {
		t.Errorf("Expected network ID %s, got %s", network.ID, pool.NetworkID)
	}

	if pool.Asset != "USDC" {
		t.Errorf("Expected asset USDC, got %s", pool.Asset)
	}

	if pool.Status != PoolStatusActive {
		t.Errorf("Expected status %d, got %d", PoolStatusActive, pool.Status)
	}

	if len(defi.LendingPools) != 1 {
		t.Errorf("Expected 1 lending pool, got %d", len(defi.LendingPools))
	}
}

// TestCreateYieldFarm tests yield farm creation
func TestCreateYieldFarm(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks:   5,
		MaxYieldFarms: 10,
	})

	// Add a network first
	network := NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{})
	defi.AddNetwork(network)

	// Create yield farm config
	config := YieldFarmConfig{
		MinStakeAmount:   big.NewInt(1000000000000000000),                                                          // 1 ETH
		MaxStakeAmount:   new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 100 ETH
		LockPeriod:       time.Hour * 24 * 7,                                                                       // 1 week
		EnableCompound:   true,
		RewardMultiplier: 1.5,
		SecurityLevel:    SecurityLevelHigh,
	}

	farm, err := defi.CreateYieldFarm(network.ID, "ETH", "REWARD", config)
	if err != nil {
		t.Errorf("Failed to create yield farm: %v", err)
	}

	if farm == nil {
		t.Fatal("Expected farm to be created")
	}

	if farm.NetworkID != network.ID {
		t.Errorf("Expected network ID %s, got %s", network.ID, farm.NetworkID)
	}

	if farm.StakingToken != "ETH" {
		t.Errorf("Expected staking token ETH, got %s", farm.StakingToken)
	}

	if farm.RewardToken != "REWARD" {
		t.Errorf("Expected reward token REWARD, got %s", farm.RewardToken)
	}

	if farm.Status != FarmStatusActive {
		t.Errorf("Expected status %d, got %d", FarmStatusActive, farm.Status)
	}

	if len(defi.YieldFarms) != 1 {
		t.Errorf("Expected 1 yield farm, got %d", len(defi.YieldFarms))
	}
}

// TestCreateLiquidityPool tests liquidity pool creation
func TestCreateLiquidityPool(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks:       5,
		MaxLiquidityPools: 10,
	})

	// Add a network first
	network := NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{})
	defi.AddNetwork(network)

	// Create liquidity pool config
	config := LiquidityPoolConfig{
		MinLiquidity:       new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)), // Reduced to 10^12
		MaxSlippage:        0.05,                                                                                   // 5%
		EnableRebalancing:  true,
		RebalanceThreshold: 0.1, // 10%
		SecurityLevel:      SecurityLevelHigh,
	}

	pool, err := defi.CreateLiquidityPool(network.ID, "ETH", "USDC", config)
	if err != nil {
		t.Errorf("Failed to create liquidity pool: %v", err)
	}

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	if pool.NetworkID != network.ID {
		t.Errorf("Expected network ID %s, got %s", network.ID, pool.NetworkID)
	}

	if pool.TokenA != "ETH" {
		t.Errorf("Expected token A ETH, got %s", pool.TokenA)
	}

	if pool.TokenB != "USDC" {
		t.Errorf("Expected token B USDC, got %s", pool.TokenB)
	}

	if pool.Fee != 0.003 {
		t.Errorf("Expected fee 0.003, got %f", pool.Fee)
	}

	if len(defi.LiquidityPools) != 1 {
		t.Errorf("Expected 1 liquidity pool, got %d", len(defi.LiquidityPools))
	}
}

// TestLendingPoolDeposit tests lending pool deposits
func TestLendingPoolDeposit(t *testing.T) {
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:           0.02,
			Multiplier:         0.1,
			JumpMultiplier:     0.5,
			OptimalUtilization: 0.8,
		},
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLendingPool("network-1", "USDC", config)

	// Test successful deposit
	depositAmount := big.NewInt(1000000000) // 1000 USDC
	err := pool.Deposit(depositAmount)
	if err != nil {
		t.Errorf("Failed to deposit: %v", err)
	}

	if pool.TotalSupply.Cmp(depositAmount) != 0 {
		t.Errorf("Expected total supply %s, got %s", depositAmount.String(), pool.TotalSupply.String())
	}

	if pool.UtilizationRate != 0.0 {
		t.Errorf("Expected utilization rate 0.0, got %f", pool.UtilizationRate)
	}

	// Test invalid deposit amount
	err = pool.Deposit(big.NewInt(0))
	if err == nil {
		t.Error("Expected error for zero deposit amount")
	}

	err = pool.Deposit(big.NewInt(-1000))
	if err == nil {
		t.Error("Expected error for negative deposit amount")
	}
}

// TestLendingPoolBorrow tests lending pool borrowing
func TestLendingPoolBorrow(t *testing.T) {
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:           0.02,
			Multiplier:         0.1,
			JumpMultiplier:     0.5,
			OptimalUtilization: 0.8,
		},
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLendingPool("network-1", "USDC", config)

	// Deposit first
	depositAmount := big.NewInt(1000000000) // 1000 USDC
	pool.Deposit(depositAmount)

	// Test successful borrow
	borrowAmount := big.NewInt(500000000) // 500 USDC
	err := pool.Borrow(borrowAmount)
	if err != nil {
		t.Errorf("Failed to borrow: %v", err)
	}

	if pool.TotalBorrowed.Cmp(borrowAmount) != 0 {
		t.Errorf("Expected total borrowed %s, got %s", borrowAmount.String(), pool.TotalBorrowed.String())
	}

	if pool.UtilizationRate != 0.5 {
		t.Errorf("Expected utilization rate 0.5, got %f", pool.UtilizationRate)
	}

	// Test borrow amount exceeding supply
	err = pool.Borrow(big.NewInt(600000000)) // 600 USDC
	if err == nil {
		t.Error("Expected error for borrow exceeding supply")
	}
}

// TestYieldFarmStaking tests yield farm staking
func TestYieldFarmStaking(t *testing.T) {
	config := YieldFarmConfig{
		MinStakeAmount:   big.NewInt(1000000000000000000),                                                          // 1 ETH
		MaxStakeAmount:   new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)), // 100 ETH
		LockPeriod:       time.Hour * 24 * 7,                                                                       // 1 week
		EnableCompound:   true,
		RewardMultiplier: 1.5,
		SecurityLevel:    SecurityLevelHigh,
	}

	farm := NewYieldFarm("network-1", "ETH", "REWARD", config)

	// Test successful staking
	stakeAmount := big.NewInt(5000000000000000000) // 5 ETH
	err := farm.Stake(stakeAmount)
	if err != nil {
		t.Errorf("Failed to stake: %v", err)
	}

	if farm.TotalStaked.Cmp(stakeAmount) != 0 {
		t.Errorf("Expected total staked %s, got %s", stakeAmount.String(), farm.TotalStaked.String())
	}

	if farm.metrics.TotalStakers != 1 {
		t.Errorf("Expected 1 total staker, got %d", farm.metrics.TotalStakers)
	}

	if farm.metrics.ActiveStakers != 1 {
		t.Errorf("Expected 1 active staker, got %d", farm.metrics.ActiveStakers)
	}

	// Test staking below minimum
	err = farm.Stake(big.NewInt(500000000000000000)) // 0.5 ETH
	if err == nil {
		t.Error("Expected error for stake below minimum")
	}

	// Test staking above maximum
	err = farm.Stake(new(big.Int).Mul(big.NewInt(200), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))) // 200 ETH
	if err == nil {
		t.Error("Expected error for stake above maximum")
	}
}

// TestLiquidityPoolAddLiquidity tests adding liquidity to pools
func TestLiquidityPoolAddLiquidity(t *testing.T) {
	config := LiquidityPoolConfig{
		MinLiquidity:       new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)), // Reduced to 10^12
		MaxSlippage:        0.05,                                                                                   // 5%
		EnableRebalancing:  true,
		RebalanceThreshold: 0.1, // 10%
		SecurityLevel:      SecurityLevelHigh,
	}

	pool := NewLiquidityPool("network-1", "ETH", "USDC", config)

	// Test adding initial liquidity
	amountA := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))   // 1 ETH
	amountB := new(big.Int).Mul(big.NewInt(2000), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 2000 USDC

	liquidityTokens, err := pool.AddLiquidity(amountA, amountB)
	if err != nil {
		t.Errorf("Failed to add liquidity: %v", err)
	}

	if liquidityTokens == nil {
		t.Fatal("Expected liquidity tokens to be returned")
	}

	if pool.ReserveA.Cmp(amountA) != 0 {
		t.Errorf("Expected reserve A %s, got %s", amountA.String(), pool.ReserveA.String())
	}

	if pool.ReserveB.Cmp(amountB) != 0 {
		t.Errorf("Expected reserve B %s, got %s", amountB.String(), pool.ReserveB.String())
	}

	if pool.TotalSupply.Cmp(liquidityTokens) != 0 {
		t.Errorf("Expected total supply %s, got %s", liquidityTokens.String(), pool.TotalSupply.String())
	}
}

// TestLiquidityPoolSwap tests swapping in liquidity pools
func TestLiquidityPoolSwap(t *testing.T) {
	config := LiquidityPoolConfig{
		MinLiquidity:       new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)), // Reduced to 10^12
		MaxSlippage:        0.05,                                                                                   // 5%
		EnableRebalancing:  true,
		RebalanceThreshold: 0.1, // 10%
		SecurityLevel:      SecurityLevelHigh,
	}

	pool := NewLiquidityPool("network-1", "ETH", "USDC", config)

	// Add initial liquidity
	amountA := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))   // 1 ETH
	amountB := new(big.Int).Mul(big.NewInt(2000), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 2000 USDC
	pool.AddLiquidity(amountA, amountB)

	// Test swapping ETH for USDC
	swapAmount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(17), nil)) // 0.1 ETH

	// Debug: Print pool state before swap
	t.Logf("Pool reserves before swap: A=%s, B=%s", pool.ReserveA.String(), pool.ReserveB.String())
	t.Logf("Swap amount: %s", swapAmount.String())

	amountOut, err := pool.Swap("ETH", swapAmount)
	if err != nil {
		t.Errorf("Failed to swap: %v", err)
	}

	if amountOut == nil {
		t.Fatal("Expected output amount to be returned")
	}

	if amountOut.Cmp(big.NewInt(0)) <= 0 {
		t.Errorf("Expected positive output amount, got %s", amountOut.String())
	}

	// Verify reserves were updated
	expectedReserveA := new(big.Int).Add(amountA, swapAmount)
	if pool.ReserveA.Cmp(expectedReserveA) != 0 {
		t.Errorf("Expected reserve A %s, got %s", expectedReserveA.String(), pool.ReserveA.String())
	}

	// Test swapping invalid token
	_, err = pool.Swap("INVALID", swapAmount)
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// TestUtilizationRateUpdate tests utilization rate updates
func TestUtilizationRateUpdate(t *testing.T) {
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:           0.02,
			Multiplier:         0.1,
			JumpMultiplier:     0.5,
			OptimalUtilization: 0.8,
		},
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLendingPool("network-1", "USDC", config)

	// Initial utilization should be 0
	if pool.UtilizationRate != 0.0 {
		t.Errorf("Expected initial utilization rate 0.0, got %f", pool.UtilizationRate)
	}

	// Deposit and borrow to test utilization calculation
	depositAmount := big.NewInt(1000000000) // 1000 USDC
	pool.Deposit(depositAmount)

	borrowAmount := big.NewInt(600000000) // 600 USDC
	pool.Borrow(borrowAmount)

	// Utilization should be 0.6 (600/1000)
	if pool.UtilizationRate != 0.6 {
		t.Errorf("Expected utilization rate 0.6, got %f", pool.UtilizationRate)
	}
}

// TestAPYUpdate tests APY updates
func TestAPYUpdate(t *testing.T) {
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:           0.02,
			Multiplier:         0.1,
			JumpMultiplier:     0.5,
			OptimalUtilization: 0.8,
		},
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLendingPool("network-1", "USDC", config)

	// Initial APY should be base rate
	if pool.APY != 0.0 {
		t.Errorf("Expected initial APY 0.0, got %f", pool.APY)
	}

	// Deposit and borrow to test APY calculation
	depositAmount := big.NewInt(1000000000) // 1000 USDC
	pool.Deposit(depositAmount)

	borrowAmount := big.NewInt(600000000) // 600 USDC
	pool.Borrow(borrowAmount)

	// APY should be baseRate + (multiplier * utilizationRate)
	expectedAPY := 0.02 + (0.1 * 0.6) // 0.02 + 0.06 = 0.08
	if pool.APY != expectedAPY {
		t.Errorf("Expected APY %f, got %f", expectedAPY, pool.APY)
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
		config := DeFiConfig{
			SecurityLevel: level,
		}

		defi := NewCrossChainDeFi(config)
		if defi.config.SecurityLevel != level {
			t.Errorf("Expected security level %d, got %d", level, defi.config.SecurityLevel)
		}
	}
}

// TestMetricsCollection tests metrics collection
func TestMetricsCollection(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks:       5,
		MaxLendingPools:   10,
		MaxYieldFarms:     5,
		MaxLiquidityPools: 10,
	})

	// Add a network
	network := NewBlockchainNetwork("Ethereum", 1, "ETH", NetworkConfig{})
	defi.AddNetwork(network)

	// Create a lending pool
	lendingConfig := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:   0.02,
			Multiplier: 0.1,
		},
		SecurityLevel: SecurityLevelHigh,
	}
	defi.CreateLendingPool(network.ID, "USDC", lendingConfig)

	// Create a yield farm
	farmConfig := YieldFarmConfig{
		MinStakeAmount: big.NewInt(1000000000000000000),
		MaxStakeAmount: new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		SecurityLevel:  SecurityLevelHigh,
	}
	defi.CreateYieldFarm(network.ID, "ETH", "REWARD", farmConfig)

	// Create a liquidity pool
	poolConfig := LiquidityPoolConfig{
		MinLiquidity:  new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		SecurityLevel: SecurityLevelHigh,
	}
	defi.CreateLiquidityPool(network.ID, "ETH", "USDC", poolConfig)

	// Check metrics
	metrics := defi.GetMetrics()
	if metrics.TotalNetworks != 1 {
		t.Errorf("Expected 1 network, got %d", metrics.TotalNetworks)
	}

	if metrics.TotalLendingPools != 1 {
		t.Errorf("Expected 1 lending pool, got %d", metrics.TotalLendingPools)
	}

	if metrics.TotalYieldFarms != 1 {
		t.Errorf("Expected 1 yield farm, got %d", metrics.TotalYieldFarms)
	}

	if metrics.TotalLiquidityPools != 1 {
		t.Errorf("Expected 1 liquidity pool, got %d", metrics.TotalLiquidityPools)
	}
}

// TestConcurrency tests concurrent operations
func TestConcurrency(t *testing.T) {
	defi := NewCrossChainDeFi(DeFiConfig{
		MaxNetworks:       10,
		MaxLendingPools:   50,
		MaxYieldFarms:     25,
		MaxLiquidityPools: 50,
	})

	// Test concurrent network addition
	done := make(chan bool, 5)
	errors := make(chan error, 5)
	timestamp := time.Now().UnixNano()
	for i := 0; i < 5; i++ {
		go func(index int) {
			network := NewBlockchainNetwork(fmt.Sprintf("Network%d_%d", index, timestamp), uint64(index+1), "TOKEN", NetworkConfig{})
			err := defi.AddNetwork(network)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check for any errors
	select {
	case err := <-errors:
		t.Errorf("Error adding network: %v", err)
	default:
		// No errors
	}

	// Add a small delay to ensure all operations complete
	time.Sleep(10 * time.Millisecond)

	if len(defi.Networks) != 5 {
		t.Errorf("Expected 5 networks, got %d", len(defi.Networks))
	}
}

// Benchmark tests for performance
func BenchmarkLendingPoolDeposit(b *testing.B) {
	config := LendingPoolConfig{
		MaxUtilizationRate:   0.95,
		MinCollateralRatio:   1.5,
		LiquidationThreshold: 1.1,
		InterestRateModel: InterestRateModel{
			BaseRate:   0.02,
			Multiplier: 0.1,
		},
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLendingPool("network-1", "USDC", config)
	amount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 1 USDC

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Deposit(amount)
	}
}

func BenchmarkLiquidityPoolSwap(b *testing.B) {
	config := LiquidityPoolConfig{
		MinLiquidity:  new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		SecurityLevel: SecurityLevelHigh,
	}

	pool := NewLiquidityPool("network-1", "ETH", "USDC", config)

	// Add initial liquidity
	amountA := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))   // 1 ETH
	amountB := new(big.Int).Mul(big.NewInt(2000), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 2000 USDC
	pool.AddLiquidity(amountA, amountB)

	swapAmount := new(big.Int).Mul(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(16), nil)) // 0.01 ETH

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Swap("ETH", swapAmount)
	}
}
