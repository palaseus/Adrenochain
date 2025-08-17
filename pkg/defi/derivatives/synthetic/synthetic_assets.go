package synthetic

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// AssetType represents the type of underlying asset
type AssetType int

const (
	Token AssetType = iota
	Index
	BasketType
	Derivative
)

// Asset represents a base asset that can be used in synthetic products
type Asset struct {
	ID        string
	Symbol    string
	Type      AssetType
	Price     *big.Float
	Weight    *big.Float
	Decimals  int
	UpdatedAt time.Time
}

// Basket represents a weighted collection of assets
type Basket struct {
	ID                 string
	Name               string
	Description        string
	Assets             map[string]*Asset
	Weights            map[string]*big.Float
	TotalWeight        *big.Float
	RebalanceThreshold *big.Float
	LastRebalanced     time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SyntheticToken represents a synthetic token backed by underlying assets
type SyntheticToken struct {
	ID              string
	Symbol          string
	Basket          *Basket
	TotalSupply     *big.Float
	UnderlyingValue *big.Float
	CollateralRatio *big.Float
	MintFee         *big.Float
	RedeemFee       *big.Float
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RebalanceEvent represents a basket rebalancing event
type RebalanceEvent struct {
	ID         string
	BasketID   string
	OldWeights map[string]*big.Float
	NewWeights map[string]*big.Float
	Timestamp  time.Time
	Reason     string
}

// NewAsset creates a new asset
func NewAsset(id, symbol string, assetType AssetType, price *big.Float, decimals int) (*Asset, error) {
	if id == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("asset symbol cannot be empty")
	}
	if price == nil || price.Sign() < 0 {
		return nil, errors.New("asset price must be non-negative")
	}
	if decimals < 0 {
		return nil, errors.New("decimals must be non-negative")
	}

	now := time.Now()

	return &Asset{
		ID:        id,
		Symbol:    symbol,
		Type:      assetType,
		Price:     new(big.Float).Copy(price),
		Weight:    big.NewFloat(0),
		Decimals:  decimals,
		UpdatedAt: now,
	}, nil
}

// NewBasket creates a new basket of assets
func NewBasket(id, name, description string, rebalanceThreshold *big.Float) (*Basket, error) {
	if id == "" {
		return nil, errors.New("basket ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("basket name cannot be empty")
	}
	if rebalanceThreshold == nil || rebalanceThreshold.Sign() < 0 {
		return nil, errors.New("rebalance threshold must be non-negative")
	}

	now := time.Now()

	return &Basket{
		ID:                 id,
		Name:               name,
		Description:        description,
		Assets:             make(map[string]*Asset),
		Weights:            make(map[string]*big.Float),
		TotalWeight:        big.NewFloat(0),
		RebalanceThreshold: new(big.Float).Copy(rebalanceThreshold),
		LastRebalanced:     now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// NewSyntheticToken creates a new synthetic token
func NewSyntheticToken(id, symbol string, basket *Basket, mintFee, redeemFee *big.Float) (*SyntheticToken, error) {
	if id == "" {
		return nil, errors.New("token ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("token symbol cannot be empty")
	}
	if basket == nil {
		return nil, errors.New("basket cannot be nil")
	}
	if mintFee == nil || mintFee.Sign() < 0 {
		return nil, errors.New("mint fee must be non-negative")
	}
	if redeemFee == nil || redeemFee.Sign() < 0 {
		return nil, errors.New("redeem fee must be non-negative")
	}

	now := time.Now()

	return &SyntheticToken{
		ID:              id,
		Symbol:          symbol,
		Basket:          basket,
		TotalSupply:     big.NewFloat(0),
		UnderlyingValue: big.NewFloat(0),
		CollateralRatio: big.NewFloat(1),
		MintFee:         new(big.Float).Copy(mintFee),
		RedeemFee:       new(big.Float).Copy(redeemFee),
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// AddAsset adds an asset to the basket with a specified weight
func (b *Basket) AddAsset(asset *Asset, weight *big.Float) error {
	if asset == nil {
		return errors.New("asset cannot be nil")
	}
	if weight == nil || weight.Sign() <= 0 {
		return errors.New("weight must be positive")
	}

	// Check if asset already exists
	if _, exists := b.Assets[asset.ID]; exists {
		return errors.New("asset already exists in basket")
	}

	// Add asset to basket
	b.Assets[asset.ID] = asset
	b.Weights[asset.ID] = new(big.Float).Copy(weight)

	// Update total weight
	b.TotalWeight.Add(b.TotalWeight, weight)

	// Update asset weight
	asset.Weight = new(big.Float).Copy(weight)

	b.UpdatedAt = time.Now()

	return nil
}

// RemoveAsset removes an asset from the basket
func (b *Basket) RemoveAsset(assetID string) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}

	asset, exists := b.Assets[assetID]
	if !exists {
		return errors.New("asset not found in basket")
	}

	// Remove asset weight from total
	weight := b.Weights[assetID]
	b.TotalWeight.Sub(b.TotalWeight, weight)

	// Remove asset and weight
	delete(b.Assets, assetID)
	delete(b.Weights, assetID)

	// Reset asset weight
	asset.Weight = big.NewFloat(0)

	b.UpdatedAt = time.Now()

	return nil
}

// UpdateAssetWeight updates the weight of an asset in the basket
func (b *Basket) UpdateAssetWeight(assetID string, newWeight *big.Float) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}
	if newWeight == nil || newWeight.Sign() <= 0 {
		return errors.New("new weight must be positive")
	}

	oldWeight, exists := b.Weights[assetID]
	if !exists {
		return errors.New("asset not found in basket")
	}

	// Update total weight
	b.TotalWeight.Sub(b.TotalWeight, oldWeight)
	b.TotalWeight.Add(b.TotalWeight, newWeight)

	// Update weight
	b.Weights[assetID] = new(big.Float).Copy(newWeight)

	// Update asset weight
	if asset, exists := b.Assets[assetID]; exists {
		asset.Weight = new(big.Float).Copy(newWeight)
	}

	b.UpdatedAt = time.Now()

	return nil
}

// GetBasketValue calculates the total value of the basket
func (b *Basket) GetBasketValue() *big.Float {
	totalValue := big.NewFloat(0)

	for assetID, asset := range b.Assets {
		weight := b.Weights[assetID]
		assetValue := new(big.Float).Mul(asset.Price, weight)
		totalValue.Add(totalValue, assetValue)
	}

	return totalValue
}

// GetAssetAllocation returns the current allocation of assets in the basket
func (b *Basket) GetAssetAllocation() map[string]*big.Float {
	allocation := make(map[string]*big.Float)

	for assetID, asset := range b.Assets {
		weight := b.Weights[assetID]
		assetValue := new(big.Float).Mul(asset.Price, weight)
		basketValue := b.GetBasketValue()

		if basketValue.Sign() > 0 {
			allocation[assetID] = new(big.Float).Quo(assetValue, basketValue)
		} else {
			allocation[assetID] = big.NewFloat(0)
		}
	}

	return allocation
}

// CheckRebalanceNeeded checks if the basket needs rebalancing
func (b *Basket) CheckRebalanceNeeded() bool {
	if b.RebalanceThreshold.Sign() == 0 {
		return false
	}

	// Get current market allocation
	allocation := b.GetAssetAllocation()

	// Calculate target allocation based on weights
	targetAllocation := make(map[string]*big.Float)
	for assetID, weight := range b.Weights {
		if b.TotalWeight.Sign() > 0 {
			targetAllocation[assetID] = new(big.Float).Quo(weight, b.TotalWeight)
		}
	}

	// Check if any asset allocation deviates beyond threshold
	for assetID, currentAlloc := range allocation {
		if targetAlloc, exists := targetAllocation[assetID]; exists {
			deviation := new(big.Float).Sub(currentAlloc, targetAlloc)
			deviation.Abs(deviation)

			if deviation.Cmp(b.RebalanceThreshold) > 0 {
				return true
			}
		}
	}

	return false
}

// Rebalance rebalances the basket to target weights
func (b *Basket) Rebalance() (*RebalanceEvent, error) {
	if !b.CheckRebalanceNeeded() {
		return nil, errors.New("rebalancing not needed")
	}

	// Store old weights
	oldWeights := make(map[string]*big.Float)
	for assetID, weight := range b.Weights {
		oldWeights[assetID] = new(big.Float).Copy(weight)
	}

	// Calculate new weights based on current prices
	basketValue := b.GetBasketValue()
	if basketValue.Sign() <= 0 {
		return nil, errors.New("basket value must be positive for rebalancing")
	}

	// Rebalance to target weights
	for assetID, asset := range b.Assets {
		targetWeight := b.Weights[assetID]
		targetValue := new(big.Float).Mul(basketValue, targetWeight)
		targetValue.Quo(targetValue, b.TotalWeight)

		// Calculate new weight based on current price
		if asset.Price.Sign() > 0 {
			newWeight := new(big.Float).Quo(targetValue, asset.Price)
			b.Weights[assetID] = newWeight
			asset.Weight = new(big.Float).Copy(newWeight)
		}
	}

	// Create rebalance event
	event := &RebalanceEvent{
		ID:         fmt.Sprintf("REBALANCE_%s_%d", b.ID, time.Now().Unix()),
		BasketID:   b.ID,
		OldWeights: oldWeights,
		NewWeights: b.Weights,
		Timestamp:  time.Now(),
		Reason:     "Automatic rebalancing due to threshold breach",
	}

	b.LastRebalanced = time.Now()
	b.UpdatedAt = time.Now()

	return event, nil
}

// Mint mints new synthetic tokens
func (st *SyntheticToken) Mint(amount *big.Float, userID string) (*big.Float, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, errors.New("mint amount must be positive")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Calculate underlying value needed
	basketValue := st.Basket.GetBasketValue()
	if basketValue.Sign() <= 0 {
		return nil, errors.New("basket value must be positive")
	}

	// Calculate tokens to mint (considering mint fee)
	mintFeeAmount := new(big.Float).Mul(amount, st.MintFee)
	tokensToMint := new(big.Float).Add(amount, mintFeeAmount)

	// Update total supply
	st.TotalSupply.Add(st.TotalSupply, tokensToMint)

	// Update underlying value
	underlyingValue := new(big.Float).Mul(amount, basketValue)
	st.UnderlyingValue.Add(st.UnderlyingValue, underlyingValue)

	// Update collateral ratio
	if st.TotalSupply.Sign() > 0 {
		st.CollateralRatio.Quo(st.UnderlyingValue, st.TotalSupply)
	}

	st.UpdatedAt = time.Now()

	return tokensToMint, nil
}

// Redeem redeems synthetic tokens for underlying assets
func (st *SyntheticToken) Redeem(tokenAmount *big.Float, userID string) (*big.Float, error) {
	if tokenAmount == nil || tokenAmount.Sign() <= 0 {
		return nil, errors.New("redeem amount must be positive")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if tokenAmount.Cmp(st.TotalSupply) > 0 {
		return nil, errors.New("insufficient tokens for redemption")
	}

	// Calculate underlying value to redeem
	basketValue := st.Basket.GetBasketValue()
	if basketValue.Sign() <= 0 {
		return nil, errors.New("basket value must be positive")
	}

	// Calculate underlying value (considering redeem fee)
	redeemFeeAmount := new(big.Float).Mul(tokenAmount, st.RedeemFee)
	underlyingValue := new(big.Float).Sub(tokenAmount, redeemFeeAmount)

	// Calculate the proportion of underlying value to redeem
	if st.TotalSupply.Sign() > 0 {
		proportion := new(big.Float).Quo(underlyingValue, st.TotalSupply)
		underlyingValue.Mul(proportion, st.UnderlyingValue)
	}

	// Update total supply
	st.TotalSupply.Sub(st.TotalSupply, tokenAmount)

	// Update underlying value
	st.UnderlyingValue.Sub(st.UnderlyingValue, underlyingValue)

	// Update collateral ratio
	if st.TotalSupply.Sign() > 0 {
		st.CollateralRatio.Quo(st.UnderlyingValue, st.TotalSupply)
	} else {
		st.CollateralRatio = big.NewFloat(0)
	}

	st.UpdatedAt = time.Now()

	return underlyingValue, nil
}

// GetTokenPrice calculates the current price of the synthetic token
func (st *SyntheticToken) GetTokenPrice() *big.Float {
	if st.TotalSupply.Sign() <= 0 {
		return big.NewFloat(0)
	}

	basketValue := st.Basket.GetBasketValue()
	tokenPrice := new(big.Float).Quo(basketValue, st.TotalSupply)

	return tokenPrice
}

// GetCollateralizationRatio returns the current collateralization ratio
func (st *SyntheticToken) GetCollateralizationRatio() *big.Float {
	if st.TotalSupply.Sign() <= 0 {
		return big.NewFloat(0)
	}

	return new(big.Float).Copy(st.CollateralRatio)
}

// UpdateAssetPrice updates the price of an asset in the basket
func (b *Basket) UpdateAssetPrice(assetID string, newPrice *big.Float) error {
	if assetID == "" {
		return errors.New("asset ID cannot be empty")
	}
	if newPrice == nil || newPrice.Sign() < 0 {
		return errors.New("new price must be non-negative")
	}

	asset, exists := b.Assets[assetID]
	if !exists {
		return errors.New("asset not found in basket")
	}

	asset.Price = new(big.Float).Copy(newPrice)
	asset.UpdatedAt = time.Now()
	b.UpdatedAt = time.Now()

	return nil
}

// GetBasketPerformance calculates the performance metrics of the basket
func (b *Basket) GetBasketPerformance() map[string]*big.Float {
	performance := make(map[string]*big.Float)

	// Calculate weighted average return
	totalWeightedReturn := big.NewFloat(0)

	for assetID, asset := range b.Assets {
		weight := b.Weights[assetID]
		// This is a simplified calculation - in practice you'd need historical prices
		weightedReturn := new(big.Float).Mul(asset.Price, weight)
		totalWeightedReturn.Add(totalWeightedReturn, weightedReturn)
	}

	performance["totalValue"] = b.GetBasketValue()
	performance["weightedReturn"] = totalWeightedReturn

	return performance
}
