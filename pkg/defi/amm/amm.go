package amm

import (
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// AMM represents an Automated Market Maker with constant product formula
type AMM struct {
	mu sync.RWMutex

	// Pool information
	PoolID      string
	TokenA      engine.Address
	TokenB      engine.Address
	ReserveA    *big.Int
	ReserveB    *big.Int
	TotalSupply *big.Int
	Fee         *big.Int // Fee in basis points (e.g., 30 = 0.3%)
	
	// Pool metadata
	Name        string
	Symbol      string
	Decimals    uint8
	Owner       engine.Address
	Paused      bool
	
	// Liquidity providers
	Providers map[engine.Address]*big.Int // address -> LP token balance
	
	// Events
	SwapEvents      []SwapEvent
	MintEvents      []MintEvent
	BurnEvents      []BurnEvent
	FeeCollectEvents []FeeCollectEvent
	
	// Statistics
	Volume24h     *big.Int
	Fees24h       *big.Int
	LastUpdate    time.Time
	SwapCount     uint64
	TotalFees     *big.Int
}

// NewAMM creates a new AMM pool
func NewAMM(
	poolID string,
	tokenA, tokenB engine.Address,
	name, symbol string,
	decimals uint8,
	owner engine.Address,
	fee *big.Int,
) *AMM {
	return &AMM{
		PoolID:      poolID,
		TokenA:      tokenA,
		TokenB:      tokenB,
		ReserveA:    big.NewInt(0),
		ReserveB:    big.NewInt(0),
		TotalSupply: big.NewInt(0),
		Fee:         new(big.Int).Set(fee),
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		Owner:       owner,
		Paused:      false,
		Providers:   make(map[engine.Address]*big.Int),
		SwapEvents:  make([]SwapEvent, 0),
		MintEvents:  make([]MintEvent, 0),
		BurnEvents:  make([]BurnEvent, 0),
		FeeCollectEvents: make([]FeeCollectEvent, 0),
		Volume24h:   big.NewInt(0),
		Fees24h:     big.NewInt(0),
		LastUpdate:  time.Now(),
		SwapCount:   0,
		TotalFees:   big.NewInt(0),
	}
}

// SwapEvent represents a token swap event
type SwapEvent struct {
	User        engine.Address
	TokenIn     engine.Address
	TokenOut    engine.Address
	AmountIn    *big.Int
	AmountOut   *big.Int
	Fee         *big.Int
	ReserveA    *big.Int
	ReserveB    *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// MintEvent represents a liquidity provision event
type MintEvent struct {
	Provider    engine.Address
	AmountA     *big.Int
	AmountB     *big.Int
	LPTokens    *big.Int
	ReserveA    *big.Int
	ReserveB    *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// BurnEvent represents a liquidity removal event
type BurnEvent struct {
	Provider    engine.Address
	AmountA     *big.Int
	AmountB     *big.Int
	LPTokens    *big.Int
	ReserveA    *big.Int
	ReserveB    *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// FeeCollectEvent represents a fee collection event
type FeeCollectEvent struct {
	Collector   engine.Address
	AmountA     *big.Int
	AmountB     *big.Int
	Timestamp   time.Time
	BlockNumber uint64
	TxHash      engine.Hash
}

// GetPoolInfo returns basic pool information
func (amm *AMM) GetPoolInfo() (string, engine.Address, engine.Address, *big.Int, *big.Int, *big.Int) {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return amm.PoolID, amm.TokenA, amm.TokenB, 
		   new(big.Int).Set(amm.ReserveA), 
		   new(big.Int).Set(amm.ReserveB), 
		   new(big.Int).Set(amm.TotalSupply)
}

// GetReserves returns current reserves
func (amm *AMM) GetReserves() (*big.Int, *big.Int) {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return new(big.Int).Set(amm.ReserveA), new(big.Int).Set(amm.ReserveB)
}

// GetFee returns the pool fee
func (amm *AMM) GetFee() *big.Int {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return new(big.Int).Set(amm.Fee)
}

// GetOwner returns the pool owner
func (amm *AMM) GetOwner() engine.Address {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return amm.Owner
}

// IsPaused returns whether the pool is paused
func (amm *AMM) IsPaused() bool {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return amm.Paused
}

// Pause pauses the pool
func (amm *AMM) Pause() error {
	amm.mu.Lock()
	defer amm.mu.Unlock()
	
	if amm.Paused {
		return ErrPoolAlreadyPaused
	}
	
	amm.Paused = true
	return nil
}

// Unpause resumes the pool
func (amm *AMM) Unpause() error {
	amm.mu.Lock()
	defer amm.mu.Unlock()
	
	if !amm.Paused {
		return ErrPoolNotPaused
	}
	
	amm.Paused = false
	return nil
}

// Swap performs a token swap
func (amm *AMM) Swap(
	user engine.Address,
	tokenIn engine.Address,
	amountIn *big.Int,
	minAmountOut *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) (*big.Int, error) {
	amm.mu.Lock()
	defer amm.mu.Unlock()
	
	// Check if pool is paused
	if amm.Paused {
		return nil, ErrPoolPaused
	}
	
	// Validate input
	if err := amm.validateSwapInput(tokenIn, amountIn); err != nil {
		return nil, err
	}
	
	// Determine token directions
	var tokenOut engine.Address
	var reserveIn, reserveOut *big.Int
	
	if tokenIn == amm.TokenA {
		tokenOut = amm.TokenB
		reserveIn = amm.ReserveA
		reserveOut = amm.ReserveB
	} else if tokenIn == amm.TokenB {
		tokenOut = amm.TokenA
		reserveIn = amm.ReserveB
		reserveOut = amm.ReserveA
	} else {
		return nil, ErrInvalidToken
	}
	
	// Calculate output amount using constant product formula
	amountOut := amm.calculateSwapOutput(amountIn, reserveIn, reserveOut)
	
	// Check slippage protection
	if amountOut.Cmp(minAmountOut) < 0 {
		return nil, ErrInsufficientOutputAmount
	}
	
	// Calculate fee
	fee := amm.calculateFee(amountIn)
	amountInAfterFee := new(big.Int).Sub(amountIn, fee)
	
	// Update reserves
	amm.ReserveA = new(big.Int).Add(amm.ReserveA, amountInAfterFee)
	amm.ReserveB = new(big.Int).Sub(amm.ReserveB, amountOut)
	
	// Update statistics
	amm.updateSwapStats(amountIn, fee)
	
	// Record event
	event := SwapEvent{
		User:        user,
		TokenIn:     tokenIn,
		TokenOut:    tokenOut,
		AmountIn:    new(big.Int).Set(amountIn),
		AmountOut:   new(big.Int).Set(amountOut),
		Fee:         new(big.Int).Set(fee),
		ReserveA:    new(big.Int).Set(amm.ReserveA),
		ReserveB:    new(big.Int).Set(amm.ReserveB),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	amm.SwapEvents = append(amm.SwapEvents, event)
	
	return amountOut, nil
}

// AddLiquidity adds liquidity to the pool
func (amm *AMM) AddLiquidity(
	provider engine.Address,
	amountA, amountB *big.Int,
	minLPTokens *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) (*big.Int, error) {
	amm.mu.Lock()
	defer amm.mu.Unlock()
	
	// Check if pool is paused
	if amm.Paused {
		return nil, ErrPoolPaused
	}
	
	// Validate input
	if err := amm.validateLiquidityInput(amountA, amountB); err != nil {
		return nil, err
	}
	
	// Calculate LP tokens to mint
	lpTokens := amm.calculateLPTokens(amountA, amountB)
	
	// Check slippage protection
	if lpTokens.Cmp(minLPTokens) < 0 {
		return nil, ErrInsufficientLPTokens
	}
	
	// Update reserves
	amm.ReserveA = new(big.Int).Add(amm.ReserveA, amountA)
	amm.ReserveB = new(big.Int).Add(amm.ReserveB, amountB)
	
	// Mint LP tokens
	amm.TotalSupply = new(big.Int).Add(amm.TotalSupply, lpTokens)
	
	// Update provider balance
	if amm.Providers[provider] == nil {
		amm.Providers[provider] = big.NewInt(0)
	}
	amm.Providers[provider] = new(big.Int).Add(amm.Providers[provider], lpTokens)
	
	// Record event
	event := MintEvent{
		Provider:    provider,
		AmountA:     new(big.Int).Set(amountA),
		AmountB:     new(big.Int).Set(amountB),
		LPTokens:    new(big.Int).Set(lpTokens),
		ReserveA:    new(big.Int).Set(amm.ReserveA),
		ReserveB:    new(big.Int).Set(amm.ReserveB),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	amm.MintEvents = append(amm.MintEvents, event)
	
	return lpTokens, nil
}

// RemoveLiquidity removes liquidity from the pool
func (amm *AMM) RemoveLiquidity(
	provider engine.Address,
	lpTokens *big.Int,
	minAmountA, minAmountB *big.Int,
	blockNumber uint64,
	txHash engine.Hash,
) (*big.Int, *big.Int, error) {
	amm.mu.Lock()
	defer amm.mu.Unlock()
	
	// Check if pool is paused
	if amm.Paused {
		return nil, nil, ErrPoolPaused
	}
	
	// Validate input
	if err := amm.validateLiquidityRemoval(provider, lpTokens); err != nil {
		return nil, nil, err
	}
	
	// Calculate amounts to return
	amountA := amm.calculateLiquidityAmount(lpTokens, amm.ReserveA)
	amountB := amm.calculateLiquidityAmount(lpTokens, amm.ReserveB)
	
	// Check slippage protection
	if amountA.Cmp(minAmountA) < 0 {
		return nil, nil, ErrInsufficientAmountA
	}
	if amountB.Cmp(minAmountB) < 0 {
		return nil, nil, ErrInsufficientAmountB
	}
	
	// Update reserves
	amm.ReserveA = new(big.Int).Sub(amm.ReserveA, amountA)
	amm.ReserveB = new(big.Int).Sub(amm.ReserveB, amountB)
	
	// Burn LP tokens
	amm.TotalSupply = new(big.Int).Sub(amm.TotalSupply, lpTokens)
	
	// Update provider balance
	amm.Providers[provider] = new(big.Int).Sub(amm.Providers[provider], lpTokens)
	
	// Record event
	event := BurnEvent{
		Provider:    provider,
		AmountA:     new(big.Int).Set(amountA),
		AmountB:     new(big.Int).Set(amountB),
		LPTokens:    new(big.Int).Set(lpTokens),
		ReserveA:    new(big.Int).Set(amm.ReserveA),
		ReserveB:    new(big.Int).Set(amm.ReserveB),
		Timestamp:   time.Now(),
		BlockNumber: blockNumber,
		TxHash:      txHash,
	}
	amm.BurnEvents = append(amm.BurnEvents, event)
	
	return amountA, amountB, nil
}

// GetLiquidityProviderBalance returns the LP token balance for a provider
func (amm *AMM) GetLiquidityProviderBalance(provider engine.Address) *big.Int {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	if balance, exists := amm.Providers[provider]; exists {
		return new(big.Int).Set(balance)
	}
	return big.NewInt(0)
}

// GetSwapEvents returns all swap events
func (amm *AMM) GetSwapEvents() []SwapEvent {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	events := make([]SwapEvent, len(amm.SwapEvents))
	copy(events, amm.SwapEvents)
	return events
}

// GetMintEvents returns all mint events
func (amm *AMM) GetMintEvents() []MintEvent {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	events := make([]MintEvent, len(amm.MintEvents))
	copy(events, amm.MintEvents)
	return events
}

// GetBurnEvents returns all burn events
func (amm *AMM) GetBurnEvents() []BurnEvent {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	events := make([]BurnEvent, len(amm.BurnEvents))
	copy(events, amm.BurnEvents)
	return events
}

// GetStatistics returns pool statistics
func (amm *AMM) GetStatistics() (uint64, *big.Int, *big.Int, *big.Int) {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	return amm.SwapCount,
		   new(big.Int).Set(amm.Volume24h),
		   new(big.Int).Set(amm.Fees24h),
		   new(big.Int).Set(amm.TotalFees)
}

// calculateSwapOutput calculates output amount using constant product formula
func (amm *AMM) calculateSwapOutput(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	if reserveIn.Sign() == 0 || reserveOut.Sign() == 0 {
		return big.NewInt(0)
	}
	
	// Calculate fee
	fee := amm.calculateFee(amountIn)
	amountInAfterFee := new(big.Int).Sub(amountIn, fee)
	
	// Constant product formula: (x + dx) * (y - dy) = x * y
	// dy = (y * dx) / (x + dx)
	numerator := new(big.Int).Mul(reserveOut, amountInAfterFee)
	denominator := new(big.Int).Add(reserveIn, amountInAfterFee)
	
	if denominator.Sign() == 0 {
		return big.NewInt(0)
	}
	
	return new(big.Int).Div(numerator, denominator)
}

// calculateFee calculates the swap fee
func (amm *AMM) calculateFee(amount *big.Int) *big.Int {
	// Fee is in basis points (e.g., 30 = 0.3%)
	feeNumerator := new(big.Int).Mul(amount, amm.Fee)
	feeDenominator := big.NewInt(10000) // 100%
	
	if feeDenominator.Sign() == 0 {
		return big.NewInt(0)
	}
	
	return new(big.Int).Div(feeNumerator, feeDenominator)
}

// calculateLPTokens calculates LP tokens to mint for given liquidity
func (amm *AMM) calculateLPTokens(amountA, amountB *big.Int) *big.Int {
	if amm.TotalSupply.Sign() == 0 {
		// First liquidity provision
		// LP tokens = sqrt(amountA * amountB)
		product := new(big.Int).Mul(amountA, amountB)
		return amm.sqrt(product)
	}
	
	// Subsequent liquidity provisions
	// LP tokens = min((amountA * totalSupply) / reserveA, (amountB * totalSupply) / reserveB)
	lpTokensA := new(big.Int).Mul(amountA, amm.TotalSupply)
	lpTokensA = new(big.Int).Div(lpTokensA, amm.ReserveA)
	
	lpTokensB := new(big.Int).Mul(amountB, amm.TotalSupply)
	lpTokensB = new(big.Int).Div(lpTokensB, amm.ReserveB)
	
	if lpTokensA.Cmp(lpTokensB) < 0 {
		return lpTokensA
	}
	return lpTokensB
}

// calculateLiquidityAmount calculates amount to return for LP tokens
func (amm *AMM) calculateLiquidityAmount(lpTokens, reserve *big.Int) *big.Int {
	if amm.TotalSupply.Sign() == 0 {
		return big.NewInt(0)
	}
	
	// amount = (lpTokens * reserve) / totalSupply
	numerator := new(big.Int).Mul(lpTokens, reserve)
	return new(big.Int).Div(numerator, amm.TotalSupply)
}

// sqrt calculates the integer square root
func (amm *AMM) sqrt(value *big.Int) *big.Int {
	if value.Sign() <= 0 {
		return big.NewInt(0)
	}
	
	// Simple integer square root using binary search
	left := big.NewInt(1)
	right := new(big.Int).Set(value)
	result := big.NewInt(0)
	
	for left.Cmp(right) <= 0 {
		mid := new(big.Int).Add(left, right)
		mid = new(big.Int).Div(mid, big.NewInt(2))
		
		midSquared := new(big.Int).Mul(mid, mid)
		
		if midSquared.Cmp(value) <= 0 {
			result = new(big.Int).Set(mid)
			left = new(big.Int).Add(mid, big.NewInt(1))
		} else {
			right = new(big.Int).Sub(mid, big.NewInt(1))
		}
	}
	
	return result
}

// updateSwapStats updates swap statistics
func (amm *AMM) updateSwapStats(amountIn, fee *big.Int) {
	amm.SwapCount++
	amm.Volume24h = new(big.Int).Add(amm.Volume24h, amountIn)
	amm.Fees24h = new(big.Int).Add(amm.Fees24h, fee)
	amm.TotalFees = new(big.Int).Add(amm.TotalFees, fee)
	amm.LastUpdate = time.Now()
}

// validateSwapInput validates swap input parameters
func (amm *AMM) validateSwapInput(tokenIn engine.Address, amountIn *big.Int) error {
	if amountIn == nil || amountIn.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if tokenIn != amm.TokenA && tokenIn != amm.TokenB {
		return ErrInvalidToken
	}
	
	return nil
}

// validateLiquidityInput validates liquidity input parameters
func (amm *AMM) validateLiquidityInput(amountA, amountB *big.Int) error {
	if amountA == nil || amountA.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	if amountB == nil || amountB.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	return nil
}

// validateLiquidityRemoval validates liquidity removal parameters
func (amm *AMM) validateLiquidityRemoval(provider engine.Address, lpTokens *big.Int) error {
	if lpTokens == nil || lpTokens.Sign() <= 0 {
		return ErrInvalidAmount
	}
	
	providerBalance := amm.Providers[provider]
	if providerBalance == nil || providerBalance.Cmp(lpTokens) < 0 {
		return ErrInsufficientLPTokens
	}
	
	return nil
}

// Clone creates a deep copy of the AMM
func (amm *AMM) Clone() *AMM {
	amm.mu.RLock()
	defer amm.mu.RUnlock()
	
	clone := &AMM{
		PoolID:      amm.PoolID,
		TokenA:      amm.TokenA,
		TokenB:      amm.TokenB,
		ReserveA:    new(big.Int).Set(amm.ReserveA),
		ReserveB:    new(big.Int).Set(amm.ReserveB),
		TotalSupply: new(big.Int).Set(amm.TotalSupply),
		Fee:         new(big.Int).Set(amm.Fee),
		Name:        amm.Name,
		Symbol:      amm.Symbol,
		Decimals:    amm.Decimals,
		Owner:       amm.Owner,
		Paused:      amm.Paused,
		Providers:   make(map[engine.Address]*big.Int),
		SwapEvents:  make([]SwapEvent, len(amm.SwapEvents)),
		MintEvents:  make([]MintEvent, len(amm.MintEvents)),
		BurnEvents:  make([]BurnEvent, len(amm.BurnEvents)),
		FeeCollectEvents: make([]FeeCollectEvent, len(amm.FeeCollectEvents)),
		Volume24h:   new(big.Int).Set(amm.Volume24h),
		Fees24h:     new(big.Int).Set(amm.Fees24h),
		LastUpdate:  amm.LastUpdate,
		SwapCount:   amm.SwapCount,
		TotalFees:   new(big.Int).Set(amm.TotalFees),
	}
	
	// Copy providers
	for addr, balance := range amm.Providers {
		clone.Providers[addr] = new(big.Int).Set(balance)
	}
	
	// Copy events
	copy(clone.SwapEvents, amm.SwapEvents)
	copy(clone.MintEvents, amm.MintEvents)
	copy(clone.BurnEvents, amm.BurnEvents)
	copy(clone.FeeCollectEvents, amm.FeeCollectEvents)
	
	return clone
}
