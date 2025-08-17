package yield

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// FarmingStrategy represents a yield farming strategy
type FarmingStrategy int

const (
	LiquidityProviding FarmingStrategy = iota
	Staking
	Lending
	YieldAggregator
	CrossChainFarming
	LeveragedFarming
	CustomStrategy
)

// RiskLevel represents the risk level of a farming strategy
type RiskLevel int

const (
	Low RiskLevel = iota
	Medium
	High
	VeryHigh
)

// Farm represents a yield farming opportunity
type Farm struct {
	ID              string
	Name            string
	Description     string
	Protocol        string
	Chain           string
	Strategy        FarmingStrategy
	RiskLevel       RiskLevel
	APY             *big.Float
	APR             *big.Float
	TVL             *big.Float
	MinStake        *big.Float
	MaxStake        *big.Float
	LockPeriod      time.Duration
	RewardTokens    []string
	StakeTokens     []string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Position represents a user's position in a farm
type Position struct {
	ID          string
	FarmID      string
	UserID      string
	StakedAmount *big.Float
	RewardsEarned *big.Float
	StartTime   time.Time
	LastClaim   time.Time
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// YieldCalculator calculates yield metrics
type YieldCalculator struct {
	mu sync.RWMutex
}

// YieldManager manages yield farming operations
type YieldManager struct {
	Farms     map[string]*Farm
	Positions map[string]*Position
	Calculator *YieldCalculator
	mu        sync.RWMutex
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewYieldManager creates a new yield manager
func NewYieldManager() *YieldManager {
	now := time.Now()
	
	return &YieldManager{
		Farms:      make(map[string]*Farm),
		Positions:  make(map[string]*Position),
		Calculator: &YieldCalculator{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewFarm creates a new farming opportunity
func NewFarm(id, name, description, protocol, chain string, strategy FarmingStrategy, riskLevel RiskLevel, apy, apr, tvl, minStake, maxStake *big.Float, lockPeriod time.Duration, rewardTokens, stakeTokens []string) (*Farm, error) {
	if id == "" {
		return nil, errors.New("farm ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("farm name cannot be empty")
	}
	if protocol == "" {
		return nil, errors.New("protocol cannot be empty")
	}
	if chain == "" {
		return nil, errors.New("chain cannot be empty")
	}
	if apy == nil || apy.Sign() < 0 {
		return nil, errors.New("APY must be non-negative")
	}
	if apr == nil || apr.Sign() < 0 {
		return nil, errors.New("APR must be non-negative")
	}
	if minStake == nil || minStake.Sign() < 0 {
		return nil, errors.New("min stake must be non-negative")
	}
	if maxStake == nil || maxStake.Sign() < 0 {
		return nil, errors.New("max stake must be non-negative")
	}
	if minStake.Sign() > 0 && maxStake.Sign() > 0 && minStake.Cmp(maxStake) > 0 {
		return nil, errors.New("min stake cannot be greater than max stake")
	}
	if rewardTokens == nil {
		rewardTokens = []string{}
	}
	if stakeTokens == nil {
		stakeTokens = []string{}
	}
	
	now := time.Now()
	
	return &Farm{
		ID:           id,
		Name:         name,
		Description:  description,
		Protocol:     protocol,
		Chain:        chain,
		Strategy:     strategy,
		RiskLevel:    riskLevel,
		APY:          new(big.Float).Copy(apy),
		APR:          new(big.Float).Copy(apr),
		TVL:          tvl,
		MinStake:     new(big.Float).Copy(minStake),
		MaxStake:     new(big.Float).Copy(maxStake),
		LockPeriod:   lockPeriod,
		RewardTokens: rewardTokens,
		StakeTokens:  stakeTokens,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewPosition creates a new farming position
func NewPosition(id, farmID, userID string, stakedAmount *big.Float) (*Position, error) {
	if id == "" {
		return nil, errors.New("position ID cannot be empty")
	}
	if farmID == "" {
		return nil, errors.New("farm ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if stakedAmount == nil || stakedAmount.Sign() <= 0 {
		return nil, errors.New("staked amount must be positive")
	}
	
	now := time.Now()
	
	return &Position{
		ID:           id,
		FarmID:       farmID,
		UserID:       userID,
		StakedAmount: new(big.Float).Copy(stakedAmount),
		RewardsEarned: big.NewFloat(0),
		StartTime:    now,
		LastClaim:    now,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// AddFarm adds a farm to the yield manager
func (ym *YieldManager) AddFarm(farm *Farm) error {
	if farm == nil {
		return errors.New("farm cannot be nil")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	if _, exists := ym.Farms[farm.ID]; exists {
		return errors.New("farm with this ID already exists")
	}
	
	ym.Farms[farm.ID] = farm
	ym.UpdatedAt = time.Now()
	
	return nil
}

// RemoveFarm removes a farm from the yield manager
func (ym *YieldManager) RemoveFarm(farmID string) error {
	if farmID == "" {
		return errors.New("farm ID cannot be empty")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	if _, exists := ym.Farms[farmID]; !exists {
		return errors.New("farm not found")
	}
	
	// Check if there are active positions
	for _, position := range ym.Positions {
		if position.FarmID == farmID && position.IsActive {
			return errors.New("cannot remove farm with active positions")
		}
	}
	
	delete(ym.Farms, farmID)
	ym.UpdatedAt = time.Now()
	
	return nil
}

// StartFarming starts a farming position
func (ym *YieldManager) StartFarming(position *Position) error {
	if position == nil {
		return errors.New("position cannot be nil")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	// Validate farm exists and is active
	farm, exists := ym.Farms[position.FarmID]
	if !exists {
		return errors.New("farm not found")
	}
	if !farm.IsActive {
		return errors.New("farm is not active")
	}
	
	// Validate stake amount
	if position.StakedAmount.Cmp(farm.MinStake) < 0 {
		return errors.New("staked amount below minimum")
	}
	if farm.MaxStake.Sign() > 0 && position.StakedAmount.Cmp(farm.MaxStake) > 0 {
		return errors.New("staked amount above maximum")
	}
	
	// Check if position already exists
	if _, exists := ym.Positions[position.ID]; exists {
		return errors.New("position with this ID already exists")
	}
	
	ym.Positions[position.ID] = position
	ym.UpdatedAt = time.Now()
	
	return nil
}

// StopFarming stops a farming position
func (ym *YieldManager) StopFarming(positionID string) error {
	if positionID == "" {
		return errors.New("position ID cannot be empty")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	position, exists := ym.Positions[positionID]
	if !exists {
		return errors.New("position not found")
	}
	
	if !position.IsActive {
		return errors.New("position is already inactive")
	}
	
	position.IsActive = false
	position.UpdatedAt = time.Now()
	ym.UpdatedAt = time.Now()
	
	return nil
}

// ClaimRewards claims rewards from a farming position
func (ym *YieldManager) ClaimRewards(positionID string) (*big.Float, error) {
	if positionID == "" {
		return nil, errors.New("position ID cannot be empty")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	position, exists := ym.Positions[positionID]
	if !exists {
		return nil, errors.New("position not found")
	}
	
	if !position.IsActive {
		return nil, errors.New("position is not active")
	}
	
	// Calculate rewards since last claim
	farm, exists := ym.Farms[position.FarmID]
	if !exists {
		return nil, errors.New("farm not found")
	}
	
	rewards := ym.Calculator.CalculateRewards(position, farm)
	
	// Update position
	position.RewardsEarned.Add(position.RewardsEarned, rewards)
	position.LastClaim = time.Now()
	position.UpdatedAt = time.Now()
	ym.UpdatedAt = time.Now()
	
	return rewards, nil
}

// CalculateRewards calculates rewards for a position
func (yc *YieldCalculator) CalculateRewards(position *Position, farm *Farm) *big.Float {
	yc.mu.RLock()
	defer yc.mu.RUnlock()
	
	// Calculate time since last claim
	timeSinceClaim := time.Since(position.LastClaim)
	if timeSinceClaim <= 0 {
		return big.NewFloat(0)
	}
	
	// Convert to years
	years := timeSinceClaim.Hours() / 8760 // 8760 hours in a year
	
	// Calculate rewards using APR (more accurate for short periods)
	rewards := new(big.Float).Mul(position.StakedAmount, farm.APR)
	rewards.Mul(rewards, big.NewFloat(years))
	
	return rewards
}

// GetFarm returns a farm by ID
func (ym *YieldManager) GetFarm(farmID string) (*Farm, error) {
	if farmID == "" {
		return nil, errors.New("farm ID cannot be empty")
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	farm, exists := ym.Farms[farmID]
	if !exists {
		return nil, errors.New("farm not found")
	}
	
	return farm, nil
}

// GetPosition returns a position by ID
func (ym *YieldManager) GetPosition(positionID string) (*Position, error) {
	if positionID == "" {
		return nil, errors.New("position ID cannot be empty")
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	position, exists := ym.Positions[positionID]
	if !exists {
		return nil, errors.New("position not found")
	}
	
	return position, nil
}

// GetFarmsByStrategy returns all farms for a specific strategy
func (ym *YieldManager) GetFarmsByStrategy(strategy FarmingStrategy) []*Farm {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var farms []*Farm
	for _, farm := range ym.Farms {
		if farm.Strategy == strategy && farm.IsActive {
			farms = append(farms, farm)
		}
	}
	
	return farms
}

// GetFarmsByRiskLevel returns all farms for a specific risk level
func (ym *YieldManager) GetFarmsByRiskLevel(riskLevel RiskLevel) []*Farm {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var farms []*Farm
	for _, farm := range ym.Farms {
		if farm.RiskLevel == riskLevel && farm.IsActive {
			farms = append(farms, farm)
		}
	}
	
	return farms
}

// GetPositionsByUser returns all positions for a specific user
func (ym *YieldManager) GetPositionsByUser(userID string) []*Position {
	if userID == "" {
		return nil
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var positions []*Position
	for _, position := range ym.Positions {
		if position.UserID == userID {
			positions = append(positions, position)
		}
	}
	
	return positions
}

// GetActivePositions returns all active positions
func (ym *YieldManager) GetActivePositions() []*Position {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var positions []*Position
	for _, position := range ym.Positions {
		if position.IsActive {
			positions = append(positions, position)
		}
	}
	
	return positions
}

// UpdateFarmAPY updates the APY of a farm
func (ym *YieldManager) UpdateFarmAPY(farmID string, newAPY *big.Float) error {
	if farmID == "" {
		return errors.New("farm ID cannot be empty")
	}
	if newAPY == nil || newAPY.Sign() < 0 {
		return errors.New("APY must be non-negative")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	farm, exists := ym.Farms[farmID]
	if !exists {
		return errors.New("farm not found")
	}
	
	farm.APY = new(big.Float).Copy(newAPY)
	farm.UpdatedAt = time.Now()
	ym.UpdatedAt = time.Now()
	
	return nil
}

// UpdateFarmAPR updates the APR of a farm
func (ym *YieldManager) UpdateFarmAPR(farmID string, newAPR *big.Float) error {
	if farmID == "" {
		return errors.New("farm ID cannot be empty")
	}
	if newAPR == nil || newAPR.Sign() < 0 {
		return errors.New("APR must be non-negative")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	farm, exists := ym.Farms[farmID]
	if !exists {
		return errors.New("farm not found")
	}
	
	farm.APR = new(big.Float).Copy(newAPR)
	farm.UpdatedAt = time.Now()
	ym.UpdatedAt = time.Now()
	
	return nil
}

// UpdateFarmTVL updates the TVL of a farm
func (ym *YieldManager) UpdateFarmTVL(farmID string, newTVL *big.Float) error {
	if farmID == "" {
		return errors.New("farm ID cannot be empty")
	}
	if newTVL == nil || newTVL.Sign() < 0 {
		return errors.New("TVL must be non-negative")
	}
	
	ym.mu.Lock()
	defer ym.mu.Unlock()
	
	farm, exists := ym.Farms[farmID]
	if !exists {
		return errors.New("farm not found")
	}
	
	farm.TVL = new(big.Float).Copy(newTVL)
	farm.UpdatedAt = time.Now()
	ym.UpdatedAt = time.Now()
	
	return nil
}

// GetTotalStaked returns the total amount staked across all farms
func (ym *YieldManager) GetTotalStaked() *big.Float {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	totalStaked := big.NewFloat(0)
	for _, position := range ym.Positions {
		if position.IsActive {
			totalStaked.Add(totalStaked, position.StakedAmount)
		}
	}
	
	return totalStaked
}

// GetTotalRewards returns the total rewards earned across all positions
func (ym *YieldManager) GetTotalRewards() *big.Float {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	totalRewards := big.NewFloat(0)
	for _, position := range ym.Positions {
		totalRewards.Add(totalRewards, position.RewardsEarned)
	}
	
	return totalRewards
}

// GetAverageAPY returns the average APY across all active farms
func (ym *YieldManager) GetAverageAPY() *big.Float {
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	if len(ym.Farms) == 0 {
		return big.NewFloat(0)
	}
	
	totalAPY := big.NewFloat(0)
	activeFarms := 0
	
	for _, farm := range ym.Farms {
		if farm.IsActive {
			totalAPY.Add(totalAPY, farm.APY)
			activeFarms++
		}
	}
	
	if activeFarms == 0 {
		return big.NewFloat(0)
	}
	
	averageAPY := new(big.Float).Quo(totalAPY, big.NewFloat(float64(activeFarms)))
	return averageAPY
}

// GetTopPerformingFarms returns the top performing farms by APY
func (ym *YieldManager) GetTopPerformingFarms(limit int) []*Farm {
	if limit <= 0 {
		limit = 10 // Default to top 10
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var farms []*Farm
	for _, farm := range ym.Farms {
		if farm.IsActive {
			farms = append(farms, farm)
		}
	}
	
	// Sort by APY (descending)
	for i := 0; i < len(farms)-1; i++ {
		for j := i + 1; j < len(farms); j++ {
			if farms[i].APY.Cmp(farms[j].APY) < 0 {
				farms[i], farms[j] = farms[j], farms[i]
			}
		}
	}
	
	// Return top N farms
	if len(farms) > limit {
		return farms[:limit]
	}
	
	return farms
}

// GetFarmsByChain returns all farms for a specific blockchain
func (ym *YieldManager) GetFarmsByChain(chain string) []*Farm {
	if chain == "" {
		return nil
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var farms []*Farm
	for _, farm := range ym.Farms {
		if farm.Chain == chain && farm.IsActive {
			farms = append(farms, farm)
		}
	}
	
	return farms
}

// GetFarmsByProtocol returns all farms for a specific protocol
func (ym *YieldManager) GetFarmsByProtocol(protocol string) []*Farm {
	if protocol == "" {
		return nil
	}
	
	ym.mu.RLock()
	defer ym.mu.RUnlock()
	
	var farms []*Farm
	for _, farm := range ym.Farms {
		if farm.Protocol == protocol && farm.IsActive {
			farms = append(farms, farm)
		}
	}
	
	return farms
}
