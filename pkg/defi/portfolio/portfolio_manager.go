package portfolio

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// Asset represents a financial asset in the portfolio
type Asset struct {
	ID          string
	Symbol      string
	Name        string
	Type        AssetType
	Price       *big.Float
	MarketCap   *big.Float
	Volume24h   *big.Float
	RiskScore   *big.Float
	LastUpdated time.Time
}

// AssetType represents the type of asset
type AssetType int

const (
	Cryptocurrency AssetType = iota
	Token
	LP_Token
	Stablecoin
	Derivative
	NFT
)

// Position represents a position in an asset
type Position struct {
	AssetID     string
	Quantity    *big.Float
	EntryPrice  *big.Float
	CurrentPrice *big.Float
	Value       *big.Float
	Pnl         *big.Float
	PnlPercent  *big.Float
	Weight      *big.Float
	LastUpdated time.Time
}

// Portfolio represents a DeFi portfolio
type Portfolio struct {
	ID                 string
	Name               string
	Description        string
	Owner              string
	TotalValue         *big.Float
	Positions          map[string]*Position
	RiskProfile        RiskProfile
	Strategy           Strategy
	CreatedAt          time.Time
	UpdatedAt          time.Time
	mu                 sync.RWMutex
	rebalancingTrades  []*RebalanceTrade
}

// RiskProfile represents the risk tolerance of a portfolio
type RiskProfile int

const (
	Conservative RiskProfile = iota
	Moderate
	Aggressive
	VeryAggressive
)

// Strategy represents the portfolio strategy
type Strategy int

const (
	BuyAndHold Strategy = iota
	ValueInvesting
	MomentumTrading
	Arbitrage
	YieldFarming
	LiquidityProviding
	Custom
)

// PortfolioManager manages multiple portfolios
type PortfolioManager struct {
	Portfolios map[string]*Portfolio
	Assets     map[string]*Asset
	mu         sync.RWMutex
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewPortfolioManager creates a new portfolio manager
func NewPortfolioManager() *PortfolioManager {
	now := time.Now()
	
	return &PortfolioManager{
		Portfolios: make(map[string]*Portfolio),
		Assets:     make(map[string]*Asset),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewPortfolio creates a new portfolio
func NewPortfolio(id, name, description, owner string, riskProfile RiskProfile, strategy Strategy) (*Portfolio, error) {
	if id == "" {
		return nil, errors.New("portfolio ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("portfolio name cannot be empty")
	}
	if owner == "" {
		return nil, errors.New("owner cannot be empty")
	}
	
	now := time.Now()
	
	return &Portfolio{
		ID:          id,
		Name:        name,
		Description: description,
		Owner:       owner,
		TotalValue:  big.NewFloat(0),
		Positions:   make(map[string]*Position),
		RiskProfile: riskProfile,
		Strategy:    strategy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewAsset creates a new asset
func NewAsset(id, symbol, name string, assetType AssetType, price, marketCap, volume24h, riskScore *big.Float) (*Asset, error) {
	if id == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("asset symbol cannot be empty")
	}
	if name == "" {
		return nil, errors.New("asset name cannot be empty")
	}
	if price == nil || price.Sign() < 0 {
		return nil, errors.New("asset price must be non-negative")
	}
	
	now := time.Now()
	
	return &Asset{
		ID:          id,
		Symbol:      symbol,
		Name:        name,
		Type:        assetType,
		Price:       new(big.Float).Copy(price),
		MarketCap:   marketCap,
		Volume24h:   volume24h,
		RiskScore:   riskScore,
		LastUpdated: now,
	}, nil
}

// NewPosition creates a new position
func NewPosition(assetID string, quantity, entryPrice *big.Float) (*Position, error) {
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if quantity == nil || quantity.Sign() <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if entryPrice == nil || entryPrice.Sign() <= 0 {
		return nil, errors.New("entry price must be positive")
	}
	
	now := time.Now()
	
	return &Position{
		AssetID:     assetID,
		Quantity:    new(big.Float).Copy(quantity),
		EntryPrice:  new(big.Float).Copy(entryPrice),
		CurrentPrice: new(big.Float).Copy(entryPrice),
		Value:       new(big.Float).Mul(quantity, entryPrice),
		Pnl:         big.NewFloat(0),
		PnlPercent:  big.NewFloat(0),
		Weight:      big.NewFloat(0),
		LastUpdated: now,
	}, nil
}

// AddPortfolio adds a portfolio to the manager
func (pm *PortfolioManager) AddPortfolio(portfolio *Portfolio) error {
	if portfolio == nil {
		return errors.New("portfolio cannot be nil")
	}
	
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.Portfolios[portfolio.ID]; exists {
		return errors.New("portfolio with this ID already exists")
	}
	
	pm.Portfolios[portfolio.ID] = portfolio
	pm.UpdatedAt = time.Now()
	
	return nil
}

// RemovePortfolio removes a portfolio from the manager
func (pm *PortfolioManager) RemovePortfolio(portfolioID string) error {
	if portfolioID == "" {
		return errors.New("portfolio ID cannot be empty")
	}
	
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.Portfolios[portfolioID]; !exists {
		return errors.New("portfolio not found")
	}
	
	delete(pm.Portfolios, portfolioID)
	pm.UpdatedAt = time.Now()
	
	return nil
}

// AddAsset adds an asset to the manager
func (pm *PortfolioManager) AddAsset(asset *Asset) error {
	if asset == nil {
		return errors.New("asset cannot be nil")
	}
	
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.Assets[asset.ID]; exists {
		return errors.New("asset with this ID already exists")
	}
	
	pm.Assets[asset.ID] = asset
	pm.UpdatedAt = time.Now()
	
	return nil
}

// UpdateAssetPrice updates the price of an asset
func (pm *PortfolioManager) UpdateAssetPrice(assetID string, newPrice *big.Float) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}
	if newPrice == nil || newPrice.Sign() < 0 {
		return errors.New("new price must be non-negative")
	}
	
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	asset, exists := pm.Assets[assetID]
	if !exists {
		return errors.New("asset not found")
	}
	
	asset.Price = new(big.Float).Copy(newPrice)
	asset.LastUpdated = time.Now()
	pm.UpdatedAt = time.Now()
	
	// Update all portfolio positions for this asset
	for _, portfolio := range pm.Portfolios {
		if _, exists := portfolio.Positions[assetID]; exists {
			portfolio.updatePositionPrice(assetID, newPrice)
			portfolio.UpdatedAt = time.Now()
		}
	}
	
	return nil
}

// AddPosition adds a position to a portfolio
func (p *Portfolio) AddPosition(position *Position) error {
	if position == nil {
		return errors.New("position cannot be nil")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if _, exists := p.Positions[position.AssetID]; exists {
		return errors.New("position for this asset already exists")
	}
	
	p.Positions[position.AssetID] = position
	p.updatePortfolioValue()
	p.UpdatedAt = time.Now()
	
	return nil
}

// UpdatePosition updates an existing position
func (p *Portfolio) UpdatePosition(assetID string, quantity, price *big.Float) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}
	if quantity == nil || quantity.Sign() <= 0 {
		return errors.New("quantity must be positive")
	}
	if price == nil || price.Sign() <= 0 {
		return errors.New("price must be positive")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	position, exists := p.Positions[assetID]
	if !exists {
		return errors.New("position not found")
	}
	
	// Update position
	position.Quantity = new(big.Float).Copy(quantity)
	position.EntryPrice = new(big.Float).Copy(price)
	position.CurrentPrice = new(big.Float).Copy(price)
	position.Value = new(big.Float).Mul(quantity, price)
	position.LastUpdated = time.Now()
	
	// Recalculate PnL
	p.calculatePositionPnL(assetID)
	
	// Update portfolio value and weights
	p.updatePortfolioValue()
	p.updatePositionWeights()
	p.UpdatedAt = time.Now()
	
	return nil
}

// RemovePosition removes a position from a portfolio
func (p *Portfolio) RemovePosition(assetID string) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if _, exists := p.Positions[assetID]; !exists {
		return errors.New("position not found")
	}
	
	delete(p.Positions, assetID)
	p.updatePortfolioValue()
	p.updatePositionWeights()
	p.UpdatedAt = time.Now()
	
	return nil
}

// updatePositionPrice updates the price of a position and recalculates PnL
func (p *Portfolio) updatePositionPrice(assetID string, newPrice *big.Float) {
	position, exists := p.Positions[assetID]
	if !exists {
		return
	}
	
	position.CurrentPrice = new(big.Float).Copy(newPrice)
	position.Value = new(big.Float).Mul(position.Quantity, newPrice)
	position.LastUpdated = time.Now()
	
	p.calculatePositionPnL(assetID)
}

// calculatePositionPnL calculates the profit/loss for a position
func (p *Portfolio) calculatePositionPnL(assetID string) {
	position, exists := p.Positions[assetID]
	if !exists {
		return
	}
	
	// Calculate PnL
	entryValue := new(big.Float).Mul(position.Quantity, position.EntryPrice)
	currentValue := new(big.Float).Mul(position.Quantity, position.CurrentPrice)
	
	position.Pnl = new(big.Float).Sub(currentValue, entryValue)
	
	// Calculate PnL percentage
	if entryValue.Sign() > 0 {
		position.PnlPercent = new(big.Float).Quo(position.Pnl, entryValue)
		position.PnlPercent.Mul(position.PnlPercent, big.NewFloat(100)) // Convert to percentage
	}
}

// updatePortfolioValue updates the total portfolio value
func (p *Portfolio) updatePortfolioValue() {
	totalValue := big.NewFloat(0)
	
	for _, position := range p.Positions {
		totalValue.Add(totalValue, position.Value)
	}
	
	p.TotalValue = totalValue
}

// updatePositionWeights updates the weight of each position
func (p *Portfolio) updatePositionWeights() {
	if p.TotalValue.Sign() <= 0 {
		return
	}
	
	for _, position := range p.Positions {
		position.Weight = new(big.Float).Quo(position.Value, p.TotalValue)
		position.Weight.Mul(position.Weight, big.NewFloat(100)) // Convert to percentage
	}
}

// GetPortfolio returns a portfolio by ID
func (pm *PortfolioManager) GetPortfolio(portfolioID string) (*Portfolio, error) {
	if portfolioID == "" {
		return nil, errors.New("portfolio ID cannot be empty")
	}
	
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	portfolio, exists := pm.Portfolios[portfolioID]
	if !exists {
		return nil, errors.New("portfolio not found")
	}
	
	return portfolio, nil
}

// GetAsset returns an asset by ID
func (pm *PortfolioManager) GetAsset(assetID string) (*Asset, error) {
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	asset, exists := pm.Assets[assetID]
	if !exists {
		return nil, errors.New("asset not found")
	}
	
	return asset, nil
}

// GetPortfoliosByOwner returns all portfolios for a specific owner
func (pm *PortfolioManager) GetPortfoliosByOwner(owner string) []*Portfolio {
	if owner == "" {
		return nil
	}
	
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	var portfolios []*Portfolio
	for _, portfolio := range pm.Portfolios {
		if portfolio.Owner == owner {
			portfolios = append(portfolios, portfolio)
		}
	}
	
	return portfolios
}

// GetPortfolioValue returns the total value of a portfolio
func (p *Portfolio) GetPortfolioValue() *big.Float {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return new(big.Float).Copy(p.TotalValue)
}

// GetPosition returns a position by asset ID
func (p *Portfolio) GetPosition(assetID string) (*Position, error) {
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	position, exists := p.Positions[assetID]
	if !exists {
		return nil, errors.New("position not found")
	}
	
	return position, nil
}

// GetTotalPnL returns the total profit/loss of the portfolio
func (p *Portfolio) GetTotalPnL() *big.Float {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	totalPnL := big.NewFloat(0)
	for _, position := range p.Positions {
		totalPnL.Add(totalPnL, position.Pnl)
	}
	
	return totalPnL
}

// GetTotalPnLPercent returns the total profit/loss percentage of the portfolio
func (p *Portfolio) GetTotalPnLPercent() *big.Float {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	totalPnL := p.GetTotalPnL()
	
	if p.TotalValue.Sign() <= 0 {
		return big.NewFloat(0)
	}
	
	totalPnLPercent := new(big.Float).Quo(totalPnL, p.TotalValue)
	totalPnLPercent.Mul(totalPnLPercent, big.NewFloat(100)) // Convert to percentage
	
	return totalPnLPercent
}

// GetAssetAllocation returns the asset allocation of the portfolio
func (p *Portfolio) GetAssetAllocation() map[string]*big.Float {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	allocation := make(map[string]*big.Float)
	for assetID, position := range p.Positions {
		allocation[assetID] = new(big.Float).Copy(position.Weight)
	}
	
	return allocation
}

// RebalancePortfolio rebalances the portfolio according to target weights
func (p *Portfolio) RebalancePortfolio(targetWeights map[string]*big.Float) error {
	if targetWeights == nil {
		return errors.New("target weights cannot be nil")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Validate target weights sum to 100%
	totalWeight := big.NewFloat(0)
	for _, weight := range targetWeights {
		totalWeight.Add(totalWeight, weight)
	}
	
	if totalWeight.Cmp(big.NewFloat(100)) != 0 {
		return errors.New("target weights must sum to 100%")
	}
	
	// Calculate target values for each asset
	targetValues := make(map[string]*big.Float)
	for assetID, weight := range targetWeights {
		targetValue := new(big.Float).Mul(p.TotalValue, weight)
		targetValue.Quo(targetValue, big.NewFloat(100))
		targetValues[assetID] = targetValue
	}
	
	// Calculate required trades
	var trades []*RebalanceTrade
	for assetID, targetValue := range targetValues {
		currentPosition, exists := p.Positions[assetID]
		if !exists {
			// Need to buy this asset
			if targetValue.Sign() > 0 {
				trade := &RebalanceTrade{
					AssetID: assetID,
					Action:  Buy,
					Value:   new(big.Float).Copy(targetValue),
				}
				trades = append(trades, trade)
			}
		} else {
			currentValue := new(big.Float).Copy(currentPosition.Value)
			if targetValue.Cmp(currentValue) > 0 {
				// Need to buy more
				buyValue := new(big.Float).Sub(targetValue, currentValue)
				trade := &RebalanceTrade{
					AssetID: assetID,
					Action:  Buy,
					Value:   buyValue,
				}
				trades = append(trades, trade)
			} else if targetValue.Cmp(currentValue) < 0 {
				// Need to sell some
				sellValue := new(big.Float).Sub(currentValue, targetValue)
				trade := &RebalanceTrade{
					AssetID: assetID,
					Action:  Sell,
					Value:   sellValue,
				}
				trades = append(trades, trade)
			}
		}
	}
	
	// Check for assets to sell that are not in target weights
	for assetID, position := range p.Positions {
		if _, exists := targetWeights[assetID]; !exists {
			// Need to sell this asset completely
			trade := &RebalanceTrade{
				AssetID: assetID,
				Action:  Sell,
				Value:   new(big.Float).Copy(position.Value),
			}
			trades = append(trades, trade)
		}
	}
	
	// Store rebalancing trades (in a real implementation, these would be executed)
	p.rebalancingTrades = trades
	p.UpdatedAt = time.Now()
	
	return nil
}

// RebalanceTrade represents a trade needed for rebalancing
type RebalanceTrade struct {
	AssetID string
	Action  TradeAction
	Value   *big.Float
}

// TradeAction represents the action to take
type TradeAction int

const (
	Buy TradeAction = iota
	Sell
	Hold
)



// GetRebalancingTrades returns the rebalancing trades
func (p *Portfolio) GetRebalancingTrades() []*RebalanceTrade {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.rebalancingTrades == nil {
		return nil
	}
	
	trades := make([]*RebalanceTrade, len(p.rebalancingTrades))
	for i, trade := range p.rebalancingTrades {
		trades[i] = &RebalanceTrade{
			AssetID: trade.AssetID,
			Action:  trade.Action,
			Value:   new(big.Float).Copy(trade.Value),
		}
	}
	
	return trades
}

// ClearRebalancingTrades clears the rebalancing trades
func (p *Portfolio) ClearRebalancingTrades() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.rebalancingTrades = nil
	p.UpdatedAt = time.Now()
}
