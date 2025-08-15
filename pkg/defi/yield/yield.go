package yield

import (
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// YieldFarm represents a yield farming protocol
type YieldFarm struct {
	mu sync.RWMutex

	// Farm information
	FarmID      string
	Name        string
	Symbol      string
	Decimals    uint8
	Owner       engine.Address
	Paused      bool
	
	// Reward token
	RewardToken engine.Address
	
	// Staking token (LP token or single token)
	StakingToken engine.Address
	
	// Reward distribution
	RewardPerSecond *big.Int // Reward tokens per second
	TotalAllocPoint *big.Int // Total allocation points
	StartTime       time.Time
	EndTime         time.Time
	
	// Pool information
	Pools map[uint64]*Pool
	
	// User information
	Users map[engine.Address]*User
	
	// Events
	DepositEvents   []DepositEvent
	WithdrawEvents  []WithdrawEvent
	HarvestEvents   []HarvestEvent
	AddPoolEvents   []AddPoolEvents
	UpdatePoolEvents []UpdatePoolEvent
	
	// Statistics
	TotalStaked     *big.Int
	TotalRewards    *big.Int
	LastUpdate      time.Time
	PoolCount       uint64
	UserCount       uint64
}

// Pool represents a staking pool
type Pool struct {
	PID           uint64
	StakingToken  engine.Address
	AllocPoint    *big.Int
	LastRewardTime time.Time
	AccRewardPerShare *big.Int
	TotalStaked   *big.Int
	Users         map[engine.Address]*UserPool
	Active        bool
}

// User represents a yield farming user
type User struct {
	Address     engine.Address
	Pools      map[uint64]*UserPool
	TotalStaked *big.Int
	TotalRewards *big.Int
	LastUpdate  time.Time
}

// UserPool represents a user's position in a pool
type UserPool struct {
	PID              uint64
	StakingToken     engine.Address
	StakedAmount     *big.Int
	RewardDebt       *big.Int
	PendingRewards   *big.Int
	LastUpdate       time.Time
}

// NewYieldFarm creates a new yield farm
func NewYieldFarm(
	farmID, name, symbol string,
	decimals uint8,
	owner engine.Address,
	rewardToken, stakingToken engine.Address,
	rewardPerSecond *big.Int,
	startTime, endTime time.Time,
) *YieldFarm {
	return &YieldFarm{
		FarmID:          farmID,
		Name:            name,
		Symbol:          symbol,
		Decimals:        decimals,
		Owner:           owner,
		Paused:          false,
		RewardToken:     rewardToken,
		StakingToken:    stakingToken,
		RewardPerSecond: new(big.Int).Set(rewardPerSecond),
		TotalAllocPoint: big.NewInt(0),
		StartTime:       startTime,
		EndTime:         endTime,
		Pools:           make(map[uint64]*Pool),
		Users:           make(map[engine.Address]*User),
		DepositEvents:   make([]DepositEvent, 0),
		WithdrawEvents:  make([]WithdrawEvent, 0),
		HarvestEvents:   make([]HarvestEvent, 0),
		AddPoolEvents:   make([]AddPoolEvents, 0),
		UpdatePoolEvents: make([]UpdatePoolEvent, 0),
		TotalStaked:     big.NewInt(0),
		TotalRewards:    big.NewInt(0),
		LastUpdate:      time.Now(),
		PoolCount:       0,
		UserCount:       0,
	}
}

// AddPool adds a new staking pool
func (yf *YieldFarm) AddPool(
	stakingToken engine.Address,
	allocPoint *big.Int,
) (uint64, error) {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	if yf.Paused {
		return 0, ErrFarmPaused
	}
	
	// Validate input
	if err := yf.validateAddPoolInput(stakingToken, allocPoint); err != nil {
		return 0, err
	}
	
	// Create new pool
	pid := yf.PoolCount
	pool := &Pool{
		PID:              pid,
		StakingToken:     stakingToken,
		AllocPoint:       new(big.Int).Set(allocPoint),
		LastRewardTime:   time.Now(),
		AccRewardPerShare: big.NewInt(0),
		TotalStaked:      big.NewInt(0),
		Users:            make(map[engine.Address]*UserPool),
		Active:           true,
	}
	
	yf.Pools[pid] = pool
	yf.TotalAllocPoint = new(big.Int).Add(yf.TotalAllocPoint, allocPoint)
	yf.PoolCount++
	
	// Record event
	event := AddPoolEvents{
		PID:         pid,
		StakingToken: stakingToken,
		AllocPoint:  new(big.Int).Set(allocPoint),
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
		TxHash:      engine.Hash{},
	}
	yf.AddPoolEvents = append(yf.AddPoolEvents, event)
	
	return pid, nil
}

// Deposit stakes tokens in a pool
func (yf *YieldFarm) Deposit(
	user engine.Address,
	pid uint64,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	// Check if farm is paused
	if yf.Paused {
		return ErrFarmPaused
	}
	
	// Validate input
	if err := yf.validateDepositInput(user, pid, amount); err != nil {
		return err
	}
	
	// Get pool
	pool := yf.Pools[pid]
	if pool == nil || !pool.Active {
		return ErrPoolNotFound
	}
	
	// Update pool rewards
	yf.updatePool(pid)
	
	// Get or create user
	if yf.Users[user] == nil {
		yf.Users[user] = &User{
			Address:     user,
			Pools:      make(map[uint64]*UserPool),
			TotalStaked: big.NewInt(0),
			TotalRewards: big.NewInt(0),
			LastUpdate:  time.Now(),
		}
		yf.UserCount++
	}
	
	// Get or create user pool
	if yf.Users[user].Pools[pid] == nil {
		yf.Users[user].Pools[pid] = &UserPool{
			PID:            pid,
			StakingToken:   pool.StakingToken,
			StakedAmount:   big.NewInt(0),
			RewardDebt:     big.NewInt(0),
			PendingRewards: big.NewInt(0),
			LastUpdate:     time.Now(),
		}
	}
	
	userPool := yf.Users[user].Pools[pid]
	
	// Calculate pending rewards
	if userPool.StakedAmount.Sign() > 0 {
		pending := new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
		pending = new(big.Int).Sub(pending, userPool.RewardDebt)
		userPool.PendingRewards = new(big.Int).Add(userPool.PendingRewards, pending)
	}
	
	// Update staked amount
	userPool.StakedAmount = new(big.Int).Add(userPool.StakedAmount, amount)
	userPool.RewardDebt = new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	
	// Update pool totals
	pool.TotalStaked = new(big.Int).Add(pool.TotalStaked, amount)
	
	// Update farm totals
	yf.TotalStaked = new(big.Int).Add(yf.TotalStaked, amount)
	yf.Users[user].TotalStaked = new(big.Int).Add(yf.Users[user].TotalStaked, amount)
	
	// Record event
	event := DepositEvent{
		User:        user,
		PID:         pid,
		Amount:      new(big.Int).Set(amount),
		TotalStaked: new(big.Int).Set(userPool.StakedAmount),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	yf.DepositEvents = append(yf.DepositEvents, event)
	
	return nil
}

// Withdraw unstakes tokens from a pool
func (yf *YieldFarm) Withdraw(
	user engine.Address,
	pid uint64,
	amount *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) error {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	// Check if farm is paused
	if yf.Paused {
		return ErrFarmPaused
	}
	
	// Validate input
	if err := yf.validateWithdrawInput(user, pid, amount); err != nil {
		return err
	}
	
	// Get pool
	pool := yf.Pools[pid]
	if pool == nil || !pool.Active {
		return ErrPoolNotFound
	}
	
	// Get user pool
	userPool := yf.Users[user].Pools[pid]
	if userPool == nil {
		return ErrUserNotInPool
	}
	
	// Check if user has sufficient staked amount
	if userPool.StakedAmount.Cmp(amount) < 0 {
		return ErrInsufficientStaked
	}
	
	// Update pool rewards
	yf.updatePool(pid)
	
	// Calculate pending rewards
	pending := new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	pending = new(big.Int).Sub(pending, userPool.RewardDebt)
	userPool.PendingRewards = new(big.Int).Add(userPool.PendingRewards, pending)
	
	// Update staked amount
	userPool.StakedAmount = new(big.Int).Sub(userPool.StakedAmount, amount)
	userPool.RewardDebt = new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	
	// Update pool totals
	pool.TotalStaked = new(big.Int).Sub(pool.TotalStaked, amount)
	
	// Update farm totals
	yf.TotalStaked = new(big.Int).Sub(yf.TotalStaked, amount)
	yf.Users[user].TotalStaked = new(big.Int).Sub(yf.Users[user].TotalStaked, amount)
	
	// Record event
	event := WithdrawEvent{
		User:        user,
		PID:         pid,
		Amount:      new(big.Int).Set(amount),
		TotalStaked: new(big.Int).Set(userPool.StakedAmount),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	yf.WithdrawEvents = append(yf.WithdrawEvents, event)
	
	return nil
}

// Harvest claims pending rewards from a pool
func (yf *YieldFarm) Harvest(
	user engine.Address,
	pid uint64,
	blockNumber uint64,
	txHash engine.Hash,
) (*big.Int, error) {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	// Check if farm is paused
	if yf.Paused {
		return nil, ErrFarmPaused
	}
	
	// Validate input
	if err := yf.validateHarvestInput(user, pid); err != nil {
		return nil, err
	}
	
	// Get pool
	pool := yf.Pools[pid]
	if pool == nil || !pool.Active {
		return nil, ErrPoolNotFound
	}
	
	// Get user pool
	userPool := yf.Users[user].Pools[pid]
	if userPool == nil {
		return nil, ErrUserNotInPool
	}
	
	// Update pool rewards
	yf.updatePool(pid)
	
	// Calculate pending rewards
	pending := new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	pending = new(big.Int).Sub(pending, userPool.RewardDebt)
	userPool.PendingRewards = new(big.Int).Add(userPool.PendingRewards, pending)
	
	// Reset reward debt
	userPool.RewardDebt = new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	
	// Get total pending rewards
	totalPending := new(big.Int).Set(userPool.PendingRewards)
	
	// Reset pending rewards
	userPool.PendingRewards = big.NewInt(0)
	
	// Update farm totals
	yf.TotalRewards = new(big.Int).Add(yf.TotalRewards, totalPending)
	yf.Users[user].TotalRewards = new(big.Int).Add(yf.Users[user].TotalRewards, totalPending)
	
	// Record event
	event := HarvestEvent{
		User:        user,
		PID:         pid,
		Amount:      new(big.Int).Set(totalPending),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	yf.HarvestEvents = append(yf.HarvestEvents, event)
	
	return totalPending, nil
}

// GetPendingRewards returns pending rewards for a user in a pool
func (yf *YieldFarm) GetPendingRewards(user engine.Address, pid uint64) (*big.Int, error) {
	yf.mu.RLock()
	defer yf.mu.RUnlock()
	
	// Get pool
	pool := yf.Pools[pid]
	if pool == nil || !pool.Active {
		return nil, ErrPoolNotFound
	}
	
	// Get user pool
	userPool := yf.Users[user].Pools[pid]
	if userPool == nil {
		return big.NewInt(0), nil
	}
	
	// Calculate pending rewards
	pending := new(big.Int).Mul(userPool.StakedAmount, pool.AccRewardPerShare)
	pending = new(big.Int).Sub(pending, userPool.RewardDebt)
	
	return new(big.Int).Add(userPool.PendingRewards, pending), nil
}

// GetUserInfo returns user information
func (yf *YieldFarm) GetUserInfo(user engine.Address) *User {
	yf.mu.RLock()
	defer yf.mu.RUnlock()
	
	if userInfo, exists := yf.Users[user]; exists {
		// Return a copy to avoid race conditions
		userCopy := &User{
			Address:     userInfo.Address,
			Pools:      make(map[uint64]*UserPool),
			TotalStaked: new(big.Int).Set(userInfo.TotalStaked),
			TotalRewards: new(big.Int).Set(userInfo.TotalRewards),
			LastUpdate:  userInfo.LastUpdate,
		}
		
		for pid, userPool := range userInfo.Pools {
			userCopy.Pools[pid] = &UserPool{
				PID:            userPool.PID,
				StakingToken:   userPool.StakingToken,
				StakedAmount:   new(big.Int).Set(userPool.StakedAmount),
				RewardDebt:     new(big.Int).Set(userPool.RewardDebt),
				PendingRewards: new(big.Int).Set(userPool.PendingRewards),
				LastUpdate:     userPool.LastUpdate,
			}
		}
		
		return userCopy
	}
	
	return nil
}

// GetPoolInfo returns pool information
func (yf *YieldFarm) GetPoolInfo(pid uint64) *Pool {
	yf.mu.RLock()
	defer yf.mu.RUnlock()
	
	if pool, exists := yf.Pools[pid]; exists {
		// Return a copy to avoid race conditions
		poolCopy := &Pool{
			PID:              pool.PID,
			StakingToken:     pool.StakingToken,
			AllocPoint:       new(big.Int).Set(pool.AllocPoint),
			LastRewardTime:   pool.LastRewardTime,
			AccRewardPerShare: new(big.Int).Set(pool.AccRewardPerShare),
			TotalStaked:      new(big.Int).Set(pool.TotalStaked),
			Users:            make(map[engine.Address]*UserPool),
			Active:           pool.Active,
		}
		
		for user, userPool := range pool.Users {
			poolCopy.Users[user] = &UserPool{
				PID:            userPool.PID,
				StakingToken:   userPool.StakingToken,
				StakedAmount:   new(big.Int).Set(userPool.StakedAmount),
				RewardDebt:     new(big.Int).Set(userPool.RewardDebt),
				PendingRewards: new(big.Int).Set(userPool.PendingRewards),
				LastUpdate:     userPool.LastUpdate,
			}
		}
		
		return poolCopy
	}
	
	return nil
}

// GetFarmStats returns farm statistics
func (yf *YieldFarm) GetFarmStats() (uint64, uint64, *big.Int, *big.Int) {
	yf.mu.RLock()
	defer yf.mu.RUnlock()
	
	return yf.PoolCount,
		   yf.UserCount,
		   new(big.Int).Set(yf.TotalStaked),
		   new(big.Int).Set(yf.TotalRewards)
}

// Pause pauses the farm
func (yf *YieldFarm) Pause() error {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	if yf.Paused {
		return ErrFarmAlreadyPaused
	}
	
	yf.Paused = true
	return nil
}

// Unpause resumes the farm
func (yf *YieldFarm) Unpause() error {
	yf.mu.Lock()
	defer yf.mu.Unlock()
	
	if !yf.Paused {
		return ErrFarmNotPaused
	}
	
	yf.Paused = false
	return nil
}

// updatePool updates pool rewards
func (yf *YieldFarm) updatePool(pid uint64) {
	pool := yf.Pools[pid]
	if pool == nil {
		return
	}
	
	// Check if pool should receive rewards
	if pool.TotalStaked.Sign() == 0 || pool.AllocPoint.Sign() == 0 {
		pool.LastRewardTime = time.Now()
		return
	}
	
	// Calculate time since last update
	now := time.Now()
	if now.Before(yf.StartTime) {
		now = yf.StartTime
	}
	if !yf.EndTime.IsZero() && now.After(yf.EndTime) {
		now = yf.EndTime
	}
	
	timeDiff := now.Sub(pool.LastRewardTime)
	if timeDiff <= 0 {
		return
	}
	
	// Calculate rewards
	seconds := big.NewInt(int64(timeDiff.Seconds()))
	rewards := new(big.Int).Mul(yf.RewardPerSecond, seconds)
	rewards = new(big.Int).Mul(rewards, pool.AllocPoint)
	
	if yf.TotalAllocPoint.Sign() > 0 {
		rewards = new(big.Int).Div(rewards, yf.TotalAllocPoint)
	}
	
	// Update pool
	pool.AccRewardPerShare = new(big.Int).Add(pool.AccRewardPerShare, rewards)
	pool.LastRewardTime = now
	
	// Record update event
	event := UpdatePoolEvent{
		PID:         pid,
		LastRewardTime: now,
		AccRewardPerShare: new(big.Int).Set(pool.AccRewardPerShare),
		TotalStaked: new(big.Int).Set(pool.TotalStaked),
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
		TxHash:      engine.Hash{},
	}
	yf.UpdatePoolEvents = append(yf.UpdatePoolEvents, event)
}

// Event types
type DepositEvent struct {
	User        engine.Address
	PID         uint64
	Amount      *big.Int
	TotalStaked *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

type WithdrawEvent struct {
	User        engine.Address
	PID         uint64
	Amount      *big.Int
	TotalStaked *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

type HarvestEvent struct {
	User        engine.Address
	PID         uint64
	Amount      *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

type AddPoolEvents struct {
	PID         uint64
	StakingToken engine.Address
	AllocPoint  *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

type UpdatePoolEvent struct {
	PID              uint64
	LastRewardTime   time.Time
	AccRewardPerShare *big.Int
	TotalStaked      *big.Int
	Timestamp        time.Time
	BlockNumber      uint64
	TxHash           engine.Hash
}

// Validation functions
func (yf *YieldFarm) validateAddPoolInput(stakingToken engine.Address, allocPoint *big.Int) error {
	if stakingToken == (engine.Address{}) {
		return ErrInvalidStakingToken
	}
	
	if allocPoint == nil || allocPoint.Sign() <= 0 {
		return ErrInvalidAllocPoint
	}
	
	return nil
}

func (yf *YieldFarm) validateDepositInput(user engine.Address, pid uint64, amount *big.Int) error {
	if user == (engine.Address{}) {
		return ErrInvalidUser
	}
	
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	return nil
}

func (yf *YieldFarm) validateWithdrawInput(user engine.Address, pid uint64, amount *big.Int) error {
	if user == (engine.Address{}) {
		return ErrInvalidUser
	}
	
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	return nil
}

func (yf *YieldFarm) validateHarvestInput(user engine.Address, pid uint64) error {
	if user == (engine.Address{}) {
		return ErrInvalidUser
	}
	
	return nil
}
