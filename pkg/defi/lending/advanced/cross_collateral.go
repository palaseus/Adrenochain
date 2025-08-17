package advanced

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/logger"
)

// CrossCollateralType represents the type of collateral
type CrossCollateralType string

const (
	CrossCollateralTypeCrypto     CrossCollateralType = "crypto"
	CrossCollateralTypeStablecoin CrossCollateralType = "stablecoin"
	CrossCollateralTypeNFT        CrossCollateralType = "nft"
	CrossCollateralTypeToken      CrossCollateralType = "token"
)

// CrossCollateralAsset represents a collateral asset in the cross-collateral system
type CrossCollateralAsset struct {
	ID               string              `json:"id"`
	Type             CrossCollateralType `json:"type"`
	Symbol           string              `json:"symbol"`
	Amount           *big.Int            `json:"amount"`
	Value            *big.Int            `json:"value"`
	LiquidationPrice *big.Int            `json:"liquidation_price"`
	Volatility       *big.Float          `json:"volatility"`
	LiquidityScore   *big.Float          `json:"liquidity_score"`
	RiskScore        *big.Float          `json:"risk_score"`
	PledgedAt        time.Time           `json:"pledged_at"`
	LastValuation    time.Time           `json:"last_valuation"`
}

// CrossCollateralPosition represents a borrowing position in the cross-collateral system
type CrossCollateralPosition struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	Asset                string     `json:"asset"`
	Amount               *big.Int   `json:"amount"`
	CollateralRatio      *big.Float `json:"collateral_ratio"`
	InterestRate         *big.Float `json:"interest_rate"`
	CreatedAt            time.Time  `json:"created_at"`
	MaturesAt            time.Time  `json:"matures_at"`
	Status               string     `json:"status"`
	CollateralAllocation []string   `json:"collateral_allocation"`
}

// CrossCollateralPortfolio represents a user's cross-collateral portfolio
type CrossCollateralPortfolio struct {
	UserID               string                              `json:"user_id"`
	TotalCollateralValue *big.Int                            `json:"total_collateral_value"`
	TotalBorrowedValue   *big.Int                            `json:"total_borrowed_value"`
	NetCollateralValue   *big.Int                            `json:"net_collateral_value"`
	CollateralRatio      *big.Float                          `json:"collateral_ratio"`
	MinCollateralRatio   *big.Float                          `json:"min_collateral_ratio"`
	RiskScore            *big.Float                          `json:"risk_score"`
	LiquidationThreshold *big.Float                          `json:"liquidation_threshold"`
	CollateralAssets     map[string]*CrossCollateralAsset    `json:"collateral_assets"`
	Positions            map[string]*CrossCollateralPosition `json:"positions"`
	RiskMetrics          *CrossCollateralRiskMetrics         `json:"risk_metrics"`
	LastRebalanced       time.Time                           `json:"last_rebalanced"`
	CreatedAt            time.Time                           `json:"created_at"`
	UpdatedAt            time.Time                           `json:"updated_at"`
}

// CrossCollateralRiskMetrics represents risk metrics for a portfolio
type CrossCollateralRiskMetrics struct {
	VaR95             *big.Float                       `json:"var_95"`
	Volatility        *big.Float                       `json:"volatility"`
	ConcentrationRisk *big.Float                       `json:"concentration_risk"`
	LiquidityRisk     *big.Float                       `json:"liquidity_risk"`
	CorrelationMatrix map[string]map[string]*big.Float `json:"correlation_matrix"`
}

// CrossCollateralManager manages cross-collateral portfolios
type CrossCollateralManager struct {
	portfolios map[string]*CrossCollateralPortfolio
	mu         sync.RWMutex
	logger     *logger.Logger
}

// NewCrossCollateralManager creates a new cross-collateral manager
func NewCrossCollateralManager() *CrossCollateralManager {
	return &CrossCollateralManager{
		portfolios: make(map[string]*CrossCollateralPortfolio),
		logger: logger.NewLogger(&logger.Config{
			Level:   logger.INFO,
			Prefix:  "cross_collateral_manager",
			UseJSON: false,
		}),
	}
}

// CreatePortfolio creates a new cross-collateral portfolio
func (ccm *CrossCollateralManager) CreatePortfolio(ctx context.Context, userID string, minCollateralRatio *big.Float) (*CrossCollateralPortfolio, error) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	if minCollateralRatio.Cmp(big.NewFloat(0)) <= 0 {
		return nil, errors.New("minimum collateral ratio must be positive")
	}

	portfolio := &CrossCollateralPortfolio{
		UserID:               userID,
		TotalCollateralValue: big.NewInt(0),
		TotalBorrowedValue:   big.NewInt(0),
		NetCollateralValue:   big.NewInt(0),
		CollateralRatio:      big.NewFloat(0),
		MinCollateralRatio:   new(big.Float).Copy(minCollateralRatio),
		RiskScore:            big.NewFloat(0),
		LiquidationThreshold: new(big.Float).Mul(minCollateralRatio, big.NewFloat(0.8)),
		CollateralAssets:     make(map[string]*CrossCollateralAsset),
		Positions:            make(map[string]*CrossCollateralPosition),
		RiskMetrics: &CrossCollateralRiskMetrics{
			CorrelationMatrix: make(map[string]map[string]*big.Float),
		},
		LastRebalanced: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	ccm.portfolios[userID] = portfolio
	ccm.logger.Info("Portfolio created - user: %s, min_collateral_ratio: %v", userID, minCollateralRatio)

	return portfolio, nil
}

// GetPortfolio retrieves a portfolio by user ID
func (ccm *CrossCollateralManager) GetPortfolio(userID string) (*CrossCollateralPortfolio, error) {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()

	portfolio, exists := ccm.portfolios[userID]
	if !exists {
		return nil, fmt.Errorf("portfolio for user %s not found", userID)
	}

	return portfolio, nil
}

// AddCollateral adds collateral to a portfolio
func (ccm *CrossCollateralManager) AddCollateral(ctx context.Context, userID string, asset *CrossCollateralAsset) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	portfolio, exists := ccm.portfolios[userID]
	if !exists {
		return fmt.Errorf("portfolio for user %s not found", userID)
	}

	// Validate asset
	if asset.Amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("collateral amount must be positive")
	}
	if asset.Value.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("collateral value must be positive")
	}

	// Set timestamps
	asset.PledgedAt = time.Now()
	asset.LastValuation = time.Now()

	// Add to portfolio
	portfolio.CollateralAssets[asset.ID] = asset

	// Update portfolio metrics
	ccm.updatePortfolioMetrics(portfolio)
	portfolio.UpdatedAt = time.Now()

	ccm.logger.Info("Collateral added - user: %s, asset: %s, amount: %s, value: %s",
		userID, asset.Symbol, asset.Amount.String(), asset.Value.String())

	return nil
}

// RemoveCollateral removes collateral from a portfolio
func (ccm *CrossCollateralManager) RemoveCollateral(ctx context.Context, userID string, assetID string, amount *big.Int) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	portfolio, exists := ccm.portfolios[userID]
	if !exists {
		return fmt.Errorf("portfolio for user %s not found", userID)
	}

	asset, exists := portfolio.CollateralAssets[assetID]
	if !exists {
		return fmt.Errorf("collateral asset %s not found", assetID)
	}

	if amount.Cmp(asset.Amount) > 0 {
		return fmt.Errorf("insufficient collateral: requested %s, available %s", amount.String(), asset.Amount.String())
	}

	// Log initial state for debugging
	ccm.logger.Info("Removing collateral - user: %s, asset: %s, current_amount: %s, current_value: %s, remove_amount: %s",
		userID, assetID, asset.Amount.String(), asset.Value.String(), amount.String())

	// Check if removal would violate minimum collateral ratio
	newAmount := new(big.Int).Sub(asset.Amount, amount)
	if newAmount.Cmp(big.NewInt(0)) == 0 {
		// Remove entire asset
		ccm.logger.Info("Removing entire asset - user: %s, asset: %s", userID, assetID)
		delete(portfolio.CollateralAssets, assetID)
	} else {
		// Update amount and recalculate value proportionally
		oldAmount := new(big.Int).Set(asset.Amount) // Store original amount before modification
		oldValue := new(big.Int).Set(asset.Value)   // Store original value before modification

		ccm.logger.Info("Partial removal - user: %s, asset: %s, old_amount: %s, old_value: %s, new_amount: %s",
			userID, assetID, oldAmount.String(), oldValue.String(), newAmount.String())

		asset.Amount = newAmount

		// Calculate new value proportionally: newValue = oldValue * (newAmount / oldAmount)
		if oldAmount.Cmp(big.NewInt(0)) > 0 {
			// Convert to big.Float for precise division
			oldAmountFloat := new(big.Float).SetInt(oldAmount)
			newAmountFloat := new(big.Float).SetInt(newAmount)
			oldValueFloat := new(big.Float).SetInt(oldValue)

			// Calculate ratio: newAmount / oldAmount
			ratio := new(big.Float).Quo(newAmountFloat, oldAmountFloat)

			// Calculate new value: oldValue * ratio
			newValueFloat := new(big.Float).Mul(oldValueFloat, ratio)

			// Convert back to big.Int
			newValueInt, _ := newValueFloat.Int(nil)
			asset.Value = newValueInt

			ccm.logger.Info("Value recalculated - user: %s, asset: %s, ratio: %v, new_value: %s",
				userID, assetID, ratio.String(), asset.Value.String())
		} else {
			ccm.logger.Warn("Invalid old amount detected - user: %s, asset: %s, old_amount: %s",
				userID, assetID, oldAmount.String())
			asset.Value = big.NewInt(0)
		}
	}

	// Log final asset state before portfolio update
	if asset, exists := portfolio.CollateralAssets[assetID]; exists {
		ccm.logger.Info("Asset state after removal - user: %s, asset: %s, final_amount: %s, final_value: %s",
			userID, assetID, asset.Amount.String(), asset.Value.String())
	}

	// Update portfolio metrics
	ccm.updatePortfolioMetrics(portfolio)
	portfolio.UpdatedAt = time.Now()

	// Log final portfolio state
	ccm.logger.Info("Portfolio updated after collateral removal - user: %s, total_collateral: %s, total_borrowed: %s, net_collateral: %s",
		userID, portfolio.TotalCollateralValue.String(), portfolio.TotalBorrowedValue.String(), portfolio.NetCollateralValue.String())

	return nil
}

// CreatePosition creates a new borrowing position
func (ccm *CrossCollateralManager) CreatePosition(ctx context.Context, userID string, asset string, amount *big.Int, collateralRatio *big.Float) (*CrossCollateralPosition, error) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	portfolio, exists := ccm.portfolios[userID]
	if !exists {
		return nil, fmt.Errorf("portfolio for user %s not found", userID)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("position amount must be positive")
	}

	if collateralRatio.Cmp(big.NewFloat(0)) <= 0 {
		return nil, errors.New("collateral ratio must be positive")
	}

	// Check if portfolio has sufficient collateral
	collateralRatioInt, _ := collateralRatio.Int(nil)
	requiredCollateral := new(big.Int).Mul(amount, collateralRatioInt)
	if portfolio.TotalCollateralValue.Cmp(requiredCollateral) < 0 {
		return nil, fmt.Errorf("insufficient collateral: required %s, available %s",
			requiredCollateral.String(), portfolio.TotalCollateralValue.String())
	}

	// Create position
	position := &CrossCollateralPosition{
		ID:                   fmt.Sprintf("pos_%s_%d", userID, time.Now().UnixNano()),
		UserID:               userID,
		Asset:                asset,
		Amount:               new(big.Int).Set(amount),
		CollateralRatio:      new(big.Float).Copy(collateralRatio),
		InterestRate:         big.NewFloat(0.08), // Default 8% interest rate
		CreatedAt:            time.Now(),
		MaturesAt:            time.Now().AddDate(0, 1, 0), // 1 month maturity
		Status:               "active",
		CollateralAllocation: make([]string, 0),
	}

	// Allocate collateral to position
	if err := ccm.allocateCollateralToPosition(portfolio, position, requiredCollateral); err != nil {
		return nil, fmt.Errorf("collateral allocation failed: %w", err)
	}

	// Add position to portfolio
	portfolio.Positions[position.ID] = position

	// Update portfolio metrics
	ccm.updatePortfolioMetrics(portfolio)
	portfolio.UpdatedAt = time.Now()

	ccm.logger.Info("Position created - user: %s, asset: %s, amount: %s, collateral_ratio: %v",
		userID, asset, amount.String(), collateralRatio)

	return position, nil
}

// allocateCollateralToPosition allocates collateral assets to a position
func (ccm *CrossCollateralManager) allocateCollateralToPosition(portfolio *CrossCollateralPortfolio, position *CrossCollateralPosition, requiredCollateral *big.Int) error {
	// Simple allocation strategy: use assets with highest liquidity scores first
	var assets []*CrossCollateralAsset
	for _, asset := range portfolio.CollateralAssets {
		assets = append(assets, asset)
	}

	allocatedValue := big.NewInt(0)
	for _, asset := range assets {
		if allocatedValue.Cmp(requiredCollateral) >= 0 {
			break
		}

		// Calculate how much of this asset to use
		remainingNeeded := new(big.Int).Sub(requiredCollateral, allocatedValue)
		assetValue := asset.Value
		if assetValue.Cmp(remainingNeeded) > 0 {
			assetValue = remainingNeeded
		}

		// Mark asset as allocated to position
		position.CollateralAllocation = append(position.CollateralAllocation, asset.ID)
		allocatedValue.Add(allocatedValue, assetValue)
	}

	if allocatedValue.Cmp(requiredCollateral) < 0 {
		return fmt.Errorf("insufficient collateral allocation: allocated %s, required %s",
			allocatedValue.String(), requiredCollateral.String())
	}

	return nil
}

// ClosePosition closes a borrowing position
func (ccm *CrossCollateralManager) ClosePosition(ctx context.Context, userID string, positionID string) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	portfolio, exists := ccm.portfolios[userID]
	if !exists {
		return fmt.Errorf("portfolio for user %s not found", userID)
	}

	position, exists := portfolio.Positions[positionID]
	if !exists {
		return fmt.Errorf("position %s not found", positionID)
	}

	if position.Status != "active" {
		return fmt.Errorf("position is not active (status: %s)", position.Status)
	}

	// Mark position as closed
	position.Status = "closed"

	// Release collateral allocation
	position.CollateralAllocation = make([]string, 0)

	// Update portfolio metrics
	ccm.updatePortfolioMetrics(portfolio)
	portfolio.UpdatedAt = time.Now()

	ccm.logger.Info("Position closed - user: %s, position: %s", userID, positionID)

	return nil
}

// updatePortfolioMetrics updates portfolio risk metrics and calculations
func (ccm *CrossCollateralManager) updatePortfolioMetrics(portfolio *CrossCollateralPortfolio) {
	ccm.logger.Info("Updating portfolio metrics for user: %s", portfolio.UserID)

	// Calculate total collateral value
	totalCollateral := big.NewInt(0)
	assetCount := 0
	for assetID, asset := range portfolio.CollateralAssets {
		if asset.Value.Cmp(big.NewInt(0)) < 0 {
			ccm.logger.Warn("Negative asset value detected - user: %s, asset: %s, value: %s",
				portfolio.UserID, assetID, asset.Value.String())
			// Reset negative values to 0
			asset.Value = big.NewInt(0)
		}
		totalCollateral.Add(totalCollateral, asset.Value)
		assetCount++
		ccm.logger.Debug("Asset contribution - user: %s, asset: %s, amount: %s, value: %s, total_so_far: %s",
			portfolio.UserID, assetID, asset.Amount.String(), asset.Value.String(), totalCollateral.String())
	}
	portfolio.TotalCollateralValue = totalCollateral

	ccm.logger.Info("Total collateral calculated - user: %s, asset_count: %d, total_collateral: %s",
		portfolio.UserID, assetCount, totalCollateral.String())

	// Calculate total borrowed value
	totalBorrowed := big.NewInt(0)
	activePositionCount := 0
	for positionID, position := range portfolio.Positions {
		if position.Status == "active" {
			totalBorrowed.Add(totalBorrowed, position.Amount)
			activePositionCount++
			ccm.logger.Debug("Active position - user: %s, position: %s, amount: %s, total_borrowed_so_far: %s",
				portfolio.UserID, positionID, position.Amount.String(), totalBorrowed.String())
		}
	}
	portfolio.TotalBorrowedValue = totalBorrowed

	ccm.logger.Info("Total borrowed calculated - user: %s, active_positions: %d, total_borrowed: %s",
		portfolio.UserID, activePositionCount, totalBorrowed.String())

	// Calculate net collateral value
	portfolio.NetCollateralValue = new(big.Int).Sub(totalCollateral, totalBorrowed)

	ccm.logger.Info("Net collateral calculated - user: %s, net_collateral: %s (total_collateral: %s - total_borrowed: %s)",
		portfolio.UserID, portfolio.NetCollateralValue.String(), totalCollateral.String(), totalBorrowed.String())

	// Calculate collateral ratio
	if totalBorrowed.Cmp(big.NewInt(0)) > 0 {
		portfolio.CollateralRatio = new(big.Float).Quo(
			new(big.Float).SetInt(totalCollateral),
			new(big.Float).SetInt(totalBorrowed),
		)
		ccm.logger.Info("Collateral ratio calculated - user: %s, ratio: %v",
			portfolio.UserID, portfolio.CollateralRatio.String())
	} else {
		portfolio.CollateralRatio = big.NewFloat(0)
		ccm.logger.Info("No borrowed value - user: %s, collateral_ratio set to 0", portfolio.UserID)
	}

	// Calculate risk metrics
	ccm.calculateRiskMetrics(portfolio)

	ccm.logger.Info("Portfolio metrics update completed - user: %s, total_collateral: %s, total_borrowed: %s, net_collateral: %s, ratio: %v",
		portfolio.UserID, portfolio.TotalCollateralValue.String(), portfolio.TotalBorrowedValue.String(),
		portfolio.NetCollateralValue.String(), portfolio.CollateralRatio.String())
}

// calculateRiskMetrics calculates comprehensive risk metrics for the portfolio
func (ccm *CrossCollateralManager) calculateRiskMetrics(portfolio *CrossCollateralPortfolio) {
	ccm.logger.Info("Calculating risk metrics for user: %s", portfolio.UserID)
	metrics := portfolio.RiskMetrics

	// Calculate VaR (Value at Risk) - simplified 95% confidence
	if portfolio.TotalCollateralValue.Cmp(big.NewInt(0)) > 0 {
		// Simple VaR calculation: 2% of total collateral value
		metrics.VaR95 = new(big.Float).Mul(
			new(big.Float).SetInt(portfolio.TotalCollateralValue),
			big.NewFloat(0.02),
		)
		ccm.logger.Debug("VaR calculated - user: %s, var_95: %v", portfolio.UserID, metrics.VaR95.String())
	} else {
		metrics.VaR95 = big.NewFloat(0)
		ccm.logger.Info("Zero collateral value - user: %s, var_95 set to 0", portfolio.UserID)
	}

	// Calculate volatility based on asset weights and individual volatilities
	metrics.Volatility = ccm.calculatePortfolioVolatility(portfolio)
	ccm.logger.Debug("Volatility calculated - user: %s, volatility: %v", portfolio.UserID, metrics.Volatility.String())

	// Calculate concentration risk
	metrics.ConcentrationRisk = ccm.calculateConcentrationRisk(portfolio)
	ccm.logger.Debug("Concentration risk calculated - user: %s, concentration_risk: %v", portfolio.UserID, metrics.ConcentrationRisk.String())

	// Calculate liquidity risk
	metrics.LiquidityRisk = ccm.calculateLiquidityRisk(portfolio)
	ccm.logger.Debug("Liquidity risk calculated - user: %s, liquidity_risk: %v", portfolio.UserID, metrics.LiquidityRisk.String())

	// Calculate correlation matrix
	ccm.updateCorrelationMatrix(portfolio)

	ccm.logger.Info("Risk metrics calculation completed for user: %s", portfolio.UserID)
}

// calculatePortfolioVolatility calculates portfolio volatility
func (ccm *CrossCollateralManager) calculatePortfolioVolatility(portfolio *CrossCollateralPortfolio) *big.Float {
	if len(portfolio.CollateralAssets) == 0 {
		return big.NewFloat(0)
	}

	// Check if total value is zero to avoid division by zero
	if portfolio.TotalCollateralValue.Cmp(big.NewInt(0)) == 0 {
		return big.NewFloat(0)
	}

	// Simple volatility calculation based on asset weights and individual volatilities
	totalValue := portfolio.TotalCollateralValue
	weightedVolatility := big.NewFloat(0)

	for _, asset := range portfolio.CollateralAssets {
		if asset.Volatility != nil && asset.Value.Cmp(big.NewInt(0)) > 0 {
			weight := new(big.Float).Quo(
				new(big.Float).SetInt(asset.Value),
				new(big.Float).SetInt(totalValue),
			)
			weightedVol := new(big.Float).Mul(weight, asset.Volatility)
			weightedVolatility.Add(weightedVolatility, weightedVol)
		}
	}

	return weightedVolatility
}

// calculateConcentrationRisk calculates concentration risk
func (ccm *CrossCollateralManager) calculateConcentrationRisk(portfolio *CrossCollateralPortfolio) *big.Float {
	if len(portfolio.CollateralAssets) == 0 {
		return big.NewFloat(0)
	}

	// Check if total value is zero to avoid division by zero
	if portfolio.TotalCollateralValue.Cmp(big.NewInt(0)) == 0 {
		return big.NewFloat(0)
	}

	// Calculate Herfindahl-Hirschman Index (HHI) for concentration
	totalValue := portfolio.TotalCollateralValue
	hhi := big.NewFloat(0)

	for _, asset := range portfolio.CollateralAssets {
		if asset.Value.Cmp(big.NewInt(0)) > 0 {
			weight := new(big.Float).Quo(
				new(big.Float).SetInt(asset.Value),
				new(big.Float).SetInt(totalValue),
			)
			weightSquared := new(big.Float).Mul(weight, weight)
			hhi.Add(hhi, weightSquared)
		}
	}

	// Normalize HHI to 0-1 scale
	numAssets := big.NewFloat(float64(len(portfolio.CollateralAssets)))
	minHHI := new(big.Float).Quo(big.NewFloat(1), numAssets)

	if hhi.Cmp(minHHI) == 0 {
		return big.NewFloat(0) // Perfect diversification
	}

	concentrationRisk := new(big.Float).Quo(
		new(big.Float).Sub(hhi, minHHI),
		new(big.Float).Sub(big.NewFloat(1), minHHI),
	)

	return concentrationRisk
}

// calculateLiquidityRisk calculates liquidity risk
func (ccm *CrossCollateralManager) calculateLiquidityRisk(portfolio *CrossCollateralPortfolio) *big.Float {
	if len(portfolio.CollateralAssets) == 0 {
		return big.NewFloat(0)
	}

	// Check if total value is zero to avoid division by zero
	if portfolio.TotalCollateralValue.Cmp(big.NewInt(0)) == 0 {
		return big.NewFloat(0)
	}

	// Calculate weighted average liquidity score
	totalValue := portfolio.TotalCollateralValue
	weightedLiquidity := big.NewFloat(0)

	for _, asset := range portfolio.CollateralAssets {
		if asset.LiquidityScore != nil && asset.Value.Cmp(big.NewInt(0)) > 0 {
			weight := new(big.Float).Quo(
				new(big.Float).SetInt(asset.Value),
				new(big.Float).SetInt(totalValue),
			)
			weightedLiq := new(big.Float).Mul(weight, asset.LiquidityScore)
			weightedLiquidity.Add(weightedLiquidity, weightedLiq)
		}
	}

	// Convert to risk score (lower liquidity = higher risk)
	liquidityRisk := new(big.Float).Sub(big.NewFloat(1), weightedLiquidity)
	return liquidityRisk
}

// updateCorrelationMatrix updates the correlation matrix between assets
func (ccm *CrossCollateralManager) updateCorrelationMatrix(portfolio *CrossCollateralPortfolio) {
	metrics := portfolio.RiskMetrics
	metrics.CorrelationMatrix = make(map[string]map[string]*big.Float)

	// Initialize correlation matrix
	for assetID1 := range portfolio.CollateralAssets {
		metrics.CorrelationMatrix[assetID1] = make(map[string]*big.Float)
		for assetID2 := range portfolio.CollateralAssets {
			if assetID1 == assetID2 {
				metrics.CorrelationMatrix[assetID1][assetID2] = big.NewFloat(1.0)
			} else {
				// In a real implementation, this would use historical data
				// For now, use simplified correlations based on asset types
				metrics.CorrelationMatrix[assetID1][assetID2] = big.NewFloat(0.3) // Default correlation
			}
		}
	}
}

// ValidatePortfolioState validates the portfolio state and returns any inconsistencies
func (ccm *CrossCollateralManager) ValidatePortfolioState(userID string) ([]string, error) {
	portfolio, err := ccm.GetPortfolio(userID)
	if err != nil {
		return nil, err
	}

	var issues []string

	// Check for negative values
	if portfolio.TotalCollateralValue.Cmp(big.NewInt(0)) < 0 {
		issues = append(issues, fmt.Sprintf("Total collateral value is negative: %s", portfolio.TotalCollateralValue.String()))
	}
	if portfolio.TotalBorrowedValue.Cmp(big.NewInt(0)) < 0 {
		issues = append(issues, fmt.Sprintf("Total borrowed value is negative: %s", portfolio.TotalBorrowedValue.String()))
	}
	if portfolio.NetCollateralValue.Cmp(big.NewInt(0)) < 0 {
		issues = append(issues, fmt.Sprintf("Net collateral value is negative: %s", portfolio.NetCollateralValue.String()))
	}

	// Check individual assets
	for assetID, asset := range portfolio.CollateralAssets {
		if asset.Amount.Cmp(big.NewInt(0)) < 0 {
			issues = append(issues, fmt.Sprintf("Asset %s has negative amount: %s", assetID, asset.Amount.String()))
		}
		if asset.Value.Cmp(big.NewInt(0)) < 0 {
			issues = append(issues, fmt.Sprintf("Asset %s has negative value: %s", assetID, asset.Value.String()))
		}
		if asset.Amount.Cmp(big.NewInt(0)) == 0 && asset.Value.Cmp(big.NewInt(0)) > 0 {
			issues = append(issues, fmt.Sprintf("Asset %s has zero amount but non-zero value: amount=%s, value=%s", 
				assetID, asset.Amount.String(), asset.Value.String()))
		}
	}

	// Check collateral ratio consistency
	if portfolio.TotalBorrowedValue.Cmp(big.NewInt(0)) > 0 {
		expectedRatio := new(big.Float).Quo(
			new(big.Float).SetInt(portfolio.TotalCollateralValue),
			new(big.Float).SetInt(portfolio.TotalBorrowedValue),
		)
		ratioDiff := new(big.Float).Sub(portfolio.CollateralRatio, expectedRatio)
		ratioDiff.Abs(ratioDiff)
		if ratioDiff.Cmp(big.NewFloat(0.0001)) > 0 {
			issues = append(issues, fmt.Sprintf("Collateral ratio mismatch: calculated=%v, stored=%v", 
				expectedRatio.String(), portfolio.CollateralRatio.String()))
		}
	}

	return issues, nil
}

// GetPortfolioAssetDetails returns detailed information about all assets in a portfolio
func (ccm *CrossCollateralManager) GetPortfolioAssetDetails(userID string) (map[string]interface{}, error) {
	portfolio, err := ccm.GetPortfolio(userID)
	if err != nil {
		return nil, err
	}

	assetDetails := make(map[string]interface{})
	for assetID, asset := range portfolio.CollateralAssets {
		assetDetails[assetID] = map[string]interface{}{
			"type":             asset.Type,
			"symbol":           asset.Symbol,
			"amount":           asset.Amount.String(),
			"value":            asset.Value.String(),
			"volatility":       asset.Volatility.String(),
			"liquidity_score":  asset.LiquidityScore.String(),
			"risk_score":       asset.RiskScore.String(),
			"pledged_at":       asset.PledgedAt,
			"last_valuation":   asset.LastValuation,
		}
	}

	return assetDetails, nil
}

// GetPortfolioStats returns comprehensive portfolio statistics
func (ccm *CrossCollateralManager) GetPortfolioStats(userID string) (map[string]interface{}, error) {
	portfolio, err := ccm.GetPortfolio(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":                portfolio.UserID,
		"total_collateral_value": portfolio.TotalCollateralValue.String(),
		"total_borrowed_value":   portfolio.TotalBorrowedValue.String(),
		"net_collateral_value":   portfolio.NetCollateralValue.String(),
		"collateral_ratio":       portfolio.CollateralRatio.String(),
		"min_collateral_ratio":   portfolio.MinCollateralRatio.String(),
		"risk_score":             portfolio.RiskScore.String(),
		"liquidation_threshold":  portfolio.LiquidationThreshold.String(),
		"collateral_asset_count": len(portfolio.CollateralAssets),
		"active_position_count":  len(portfolio.Positions),
		"last_rebalanced":        portfolio.LastRebalanced,
		"created_at":             portfolio.CreatedAt,
		"updated_at":             portfolio.UpdatedAt,
	}

	// Add risk metrics
	if portfolio.RiskMetrics != nil {
		riskStats := map[string]interface{}{
			"var_95":             portfolio.RiskMetrics.VaR95.String(),
			"volatility":         portfolio.RiskMetrics.Volatility.String(),
			"concentration_risk": portfolio.RiskMetrics.ConcentrationRisk.String(),
			"liquidity_risk":     portfolio.RiskMetrics.LiquidityRisk.String(),
		}
		stats["risk_metrics"] = riskStats
	}

	return stats, nil
}
