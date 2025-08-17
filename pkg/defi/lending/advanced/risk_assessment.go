package advanced

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

// RiskScore represents a risk assessment score
type RiskScore struct {
	OverallScore     int       `json:"overall_score"`      // 0-100, higher = riskier
	CreditScore      int       `json:"credit_score"`       // 0-100, higher = better credit
	CollateralScore  int       `json:"collateral_score"`   // 0-100, higher = better collateral
	LiquidityScore   int       `json:"liquidity_score"`    // 0-100, higher = better liquidity
	MarketScore      int       `json:"market_score"`       // 0-100, higher = better market conditions
	BehaviorScore    int       `json:"behavior_score"`     // 0-100, higher = better behavior
	LastUpdated      time.Time `json:"last_updated"`
	RiskLevel        RiskLevel `json:"risk_level"`
	RiskFactors      []string  `json:"risk_factors"`
	Recommendations  []string  `json:"recommendations"`
}

// RiskLevel represents the overall risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium  RiskLevel = "medium"
	RiskLevelHigh    RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RiskFactor represents a specific risk factor
type RiskFactor struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Weight      float64   `json:"weight"`       // 0.0-1.0, importance of this factor
	Score       int       `json:"score"`        // 0-100, calculated score
	Threshold   int       `json:"threshold"`    // Threshold for triggering alerts
	IsTriggered bool      `json:"is_triggered"` // Whether this factor is triggered
	LastUpdated time.Time `json:"last_updated"`
}

// RiskAssessment represents a comprehensive risk assessment
type RiskAssessment struct {
	UserID           string                 `json:"user_id"`
	RiskScore        *RiskScore             `json:"risk_score"`
	RiskFactors      map[string]*RiskFactor `json:"risk_factors"`
	HistoricalData   []HistoricalRiskData  `json:"historical_data"`
	CollateralValue  *big.Int              `json:"collateral_value"`
	BorrowValue      *big.Int              `json:"borrow_value"`
	LiquidationRisk  float64               `json:"liquidation_risk"`
	MarketExposure   float64               `json:"market_exposure"`
	LastAssessment   time.Time             `json:"last_assessment"`
	NextAssessment   time.Time             `json:"next_assessment"`
	AssessmentCount  int                   `json:"assessment_count"`
}

// HistoricalRiskData represents historical risk assessment data
type HistoricalRiskData struct {
	Timestamp   time.Time `json:"timestamp"`
	RiskScore   int       `json:"risk_score"`
	RiskLevel   RiskLevel `json:"risk_level"`
	Description string    `json:"description"`
}

// RiskAssessmentError represents risk assessment specific errors
type RiskAssessmentError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	UserID    string `json:"user_id,omitempty"`
}

func (e RiskAssessmentError) Error() string {
	return e.Operation + ": " + e.Message
}

// Risk assessment errors
var (
	ErrInvalidRiskScore     = errors.New("invalid risk score")
	ErrInvalidRiskFactor    = errors.New("invalid risk factor")
	ErrAssessmentNotFound   = errors.New("risk assessment not found")
	ErrInvalidThreshold     = errors.New("invalid threshold value")
	ErrInvalidWeight        = errors.New("invalid weight value")
)

// RiskAssessor manages risk assessment operations
type RiskAssessor struct {
	lendingPool     *LendingPool
	riskFactors     map[string]*RiskFactor
	assessments     map[string]*RiskAssessment
	mutex           sync.RWMutex
	updateInterval  time.Duration
	alertThreshold  int
}

// NewRiskAssessor creates a new risk assessor
func NewRiskAssessor(lendingPool *LendingPool, updateInterval time.Duration, alertThreshold int) *RiskAssessor {
	ra := &RiskAssessor{
		lendingPool:    lendingPool,
		riskFactors:    make(map[string]*RiskFactor),
		assessments:    make(map[string]*RiskAssessment),
		updateInterval: updateInterval,
		alertThreshold: alertThreshold,
	}

	// Initialize default risk factors
	ra.initializeDefaultRiskFactors()

	return ra
}

// initializeDefaultRiskFactors sets up default risk factors
func (ra *RiskAssessor) initializeDefaultRiskFactors() {
	defaultFactors := []*RiskFactor{
		{
			ID:          "collateral_ratio",
			Name:        "Collateral Ratio",
			Description: "Ratio of collateral value to borrow value",
			Weight:      0.25,
			Score:       0,
			Threshold:   80,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
		{
			ID:          "liquidation_proximity",
			Name:        "Liquidation Proximity",
			Description: "How close the account is to liquidation",
			Weight:      0.20,
			Score:       0,
			Threshold:   70,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
		{
			ID:          "market_volatility",
			Name:        "Market Volatility",
			Description: "Current market volatility impact",
			Weight:      0.15,
			Score:       0,
			Threshold:   75,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
		{
			ID:          "borrow_history",
			Name:        "Borrow History",
			Description: "Historical borrowing behavior",
			Weight:      0.15,
			Score:       0,
			Threshold:   60,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
		{
			ID:          "repayment_pattern",
			Name:        "Repayment Pattern",
			Description: "Pattern of loan repayments",
			Weight:      0.15,
			Score:       0,
			Threshold:   65,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
		{
			ID:          "liquidity_access",
			Name:        "Liquidity Access",
			Description: "Access to additional liquidity",
			Weight:      0.10,
			Score:       0,
			Threshold:   70,
			IsTriggered: false,
			LastUpdated: time.Now(),
		},
	}

	for _, factor := range defaultFactors {
		ra.riskFactors[factor.ID] = factor
	}
}

// AssessRisk performs a comprehensive risk assessment for a user
func (ra *RiskAssessor) AssessRisk(userID string) (*RiskAssessment, error) {
	ra.mutex.Lock()
	defer ra.mutex.Unlock()

	// Get user account from lending pool
	account, err := ra.lendingPool.GetAccount(userID)
	if err != nil {
		return nil, &RiskAssessmentError{
			Operation: "AssessRisk",
			Message:   "failed to get user account",
			UserID:    userID,
		}
	}

	// Create or get existing assessment
	assessment, exists := ra.assessments[userID]
	if !exists {
		assessment = &RiskAssessment{
			UserID:          userID,
			RiskFactors:     make(map[string]*RiskFactor),
			HistoricalData:  make([]HistoricalRiskData, 0),
			AssessmentCount: 0,
		}
		ra.assessments[userID] = assessment
	}

	// Update assessment data
	assessment.CollateralValue = account.CollateralValue
	assessment.BorrowValue = account.BorrowValue
	assessment.LastAssessment = time.Now()
	assessment.NextAssessment = time.Now().Add(ra.updateInterval)
	assessment.AssessmentCount++

	// Calculate risk scores
	riskScore := ra.calculateRiskScore(account, assessment)
	assessment.RiskScore = riskScore

	// Update risk factors
	ra.updateRiskFactors(assessment)

	// Calculate additional metrics
	assessment.LiquidationRisk = ra.calculateLiquidationRisk(account)
	assessment.MarketExposure = ra.calculateMarketExposure(account)

	// Add to historical data
	historicalData := HistoricalRiskData{
		Timestamp:   time.Now(),
		RiskScore:   riskScore.OverallScore,
		RiskLevel:   riskScore.RiskLevel,
		Description: "Regular risk assessment",
	}
	assessment.HistoricalData = append(assessment.HistoricalData, historicalData)

	// Keep only last 100 historical records
	if len(assessment.HistoricalData) > 100 {
		assessment.HistoricalData = assessment.HistoricalData[len(assessment.HistoricalData)-100:]
	}

	return assessment, nil
}

// calculateRiskScore calculates the overall risk score
func (ra *RiskAssessor) calculateRiskScore(account *Account, assessment *RiskAssessment) *RiskScore {
	riskScore := &RiskScore{
		LastUpdated: time.Now(),
		RiskFactors: make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Calculate collateral score
	riskScore.CollateralScore = ra.calculateCollateralScore(account)
	
	// Calculate credit score
	riskScore.CreditScore = ra.calculateCreditScore(account, assessment)
	
	// Calculate liquidity score
	riskScore.LiquidityScore = ra.calculateLiquidityScore(account)
	
	// Calculate market score
	riskScore.MarketScore = ra.calculateMarketScore(assessment)
	
	// Calculate behavior score
	riskScore.BehaviorScore = ra.calculateBehaviorScore(assessment)

	// Calculate overall score (weighted average)
	overallScore := (riskScore.CollateralScore*25 + 
		riskScore.CreditScore*20 + 
		riskScore.LiquidityScore*20 + 
		riskScore.MarketScore*15 + 
		riskScore.BehaviorScore*20) / 100

	riskScore.OverallScore = overallScore

	// Determine risk level
	riskScore.RiskLevel = ra.determineRiskLevel(overallScore)

	// Generate recommendations
	riskScore.Recommendations = ra.generateRecommendations(riskScore)

	return riskScore
}

// calculateCollateralScore calculates the collateral quality score
func (ra *RiskAssessor) calculateCollateralScore(account *Account) int {
	if account.CollateralValue.Cmp(big.NewInt(0)) == 0 {
		return 100 // No collateral = highest risk
	}

	if account.BorrowValue.Cmp(big.NewInt(0)) == 0 {
		return 0 // No borrows = lowest risk
	}

	// Calculate collateral ratio
	collateralRatio := new(big.Int).Mul(account.CollateralValue, big.NewInt(100))
	collateralRatio.Div(collateralRatio, account.BorrowValue)

	// Convert to int for scoring
	ratio := int(collateralRatio.Int64())

	// Score based on collateral ratio
	switch {
	case ratio >= 200: // 200%+ collateral
		return 0
	case ratio >= 150: // 150-199% collateral
		return 20
	case ratio >= 125: // 125-149% collateral
		return 40
	case ratio >= 110: // 110-124% collateral
		return 60
	case ratio >= 100: // 100-109% collateral
		return 80
	default: // Below 100% (liquidatable)
		return 100
	}
}

// calculateCreditScore calculates the creditworthiness score
func (ra *RiskAssessor) calculateCreditScore(account *Account, assessment *RiskAssessment) int {
	// Base score starts at 50
	score := 50

	// Adjust based on health factor
	if account.HealthFactor != nil {
		healthFactor := int(account.HealthFactor.Int64())
		switch {
		case healthFactor >= 150: // Very healthy
			score += 30
		case healthFactor >= 120: // Healthy
			score += 20
		case healthFactor >= 100: // Acceptable
			score += 10
		case healthFactor >= 80: // Risky
			score -= 10
		case healthFactor >= 60: // Very risky
			score -= 20
		default: // Critical
			score -= 30
		}
	}

	// Adjust based on assessment history
	if assessment.AssessmentCount > 10 {
		// Check if risk has been improving
		if len(assessment.HistoricalData) >= 5 {
			recent := assessment.HistoricalData[len(assessment.HistoricalData)-5:]
			improving := 0
			for i := 1; i < len(recent); i++ {
				if recent[i].RiskScore < recent[i-1].RiskScore {
					improving++
				}
			}
			if improving >= 3 {
				score += 10 // Bonus for improving risk profile
			}
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// calculateLiquidityScore calculates the liquidity access score
func (ra *RiskAssessor) calculateLiquidityScore(account *Account) int {
	// Base score starts at 50
	score := 50

	// Check if user has additional supply balance
	if account.SupplyBalance.Cmp(big.NewInt(0)) > 0 {
		score += 20 // Bonus for having additional funds
	}

	// Check if user has other assets (simplified)
	// In a real implementation, this would check multiple asset balances
	if account.CollateralValue.Cmp(account.BorrowValue) > 0 {
		excess := new(big.Int).Sub(account.CollateralValue, account.BorrowValue)
		if excess.Cmp(big.NewInt(1000000)) > 0 { // More than 1 USDC excess
			score += 15
		} else if excess.Cmp(big.NewInt(500000)) > 0 { // More than 0.5 USDC excess
			score += 10
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// calculateMarketScore calculates the market conditions score
func (ra *RiskAssessor) calculateMarketScore(assessment *RiskAssessment) int {
	// Base score starts at 50
	score := 50

	// Check pool utilization (higher utilization = higher risk)
	poolStats := ra.lendingPool.GetPoolStats()
	if utilizationStr, ok := poolStats["utilization_rate"].(string); ok {
		// Parse utilization rate (simplified)
		// In a real implementation, this would parse the actual value
		if len(utilizationStr) > 0 {
			// Higher utilization means higher market risk
			score += 20 // Conservative approach
		}
	}

	// Check if market is volatile (simplified)
	// In a real implementation, this would check actual market data
	score += 10 // Assume some market volatility

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// calculateBehaviorScore calculates the user behavior score
func (ra *RiskAssessor) calculateBehaviorScore(assessment *RiskAssessment) int {
	// Base score starts at 50
	score := 50

	// Check assessment history
	if assessment.AssessmentCount > 1 {
		// Check if user has been maintaining good risk profile
		recentAssessments := assessment.HistoricalData
		if len(recentAssessments) >= 3 {
			goodAssessments := 0
			for _, data := range recentAssessments[len(recentAssessments)-3:] {
				if data.RiskLevel == RiskLevelLow || data.RiskLevel == RiskLevelMedium {
					goodAssessments++
				}
			}
			
			if goodAssessments >= 2 {
				score += 20 // Bonus for consistent good behavior
			} else if goodAssessments == 0 {
				score -= 20 // Penalty for consistent poor behavior
			}
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return score
}

// determineRiskLevel determines the risk level based on score
func (ra *RiskAssessor) determineRiskLevel(score int) RiskLevel {
	switch {
	case score <= 25:
		return RiskLevelLow
	case score <= 50:
		return RiskLevelMedium
	case score <= 75:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}

// generateRecommendations generates risk mitigation recommendations
func (ra *RiskAssessor) generateRecommendations(riskScore *RiskScore) []string {
	var recommendations []string

	// Overall risk recommendations
	if riskScore.OverallScore > 75 {
		recommendations = append(recommendations, "Consider reducing borrow amount to improve risk profile")
		recommendations = append(recommendations, "Add more collateral to reduce liquidation risk")
		recommendations = append(recommendations, "Monitor account health factor closely")
	} else if riskScore.OverallScore > 50 {
		recommendations = append(recommendations, "Monitor collateral ratio and maintain healthy buffer")
		recommendations = append(recommendations, "Consider partial repayment to improve health factor")
	}

	// Specific factor recommendations
	if riskScore.CollateralScore > 70 {
		recommendations = append(recommendations, "Increase collateral to improve collateral ratio")
	}
	
	if riskScore.CreditScore < 40 {
		recommendations = append(recommendations, "Improve account health factor by reducing borrows or adding collateral")
	}

	if riskScore.LiquidityScore < 40 {
		recommendations = append(recommendations, "Maintain additional liquidity for emergency situations")
	}

	if riskScore.MarketScore > 70 {
		recommendations = append(recommendations, "Monitor market conditions and adjust positions accordingly")
	}

	if riskScore.BehaviorScore < 40 {
		recommendations = append(recommendations, "Maintain consistent good borrowing practices")
	}

	return recommendations
}

// calculateLiquidationRisk calculates the probability of liquidation
func (ra *RiskAssessor) calculateLiquidationRisk(account *Account) float64 {
	if account.HealthFactor == nil {
		return 0.0
	}

	healthFactor := float64(account.HealthFactor.Int64()) / 100.0

	// Convert health factor to liquidation risk
	switch {
	case healthFactor >= 2.0: // 200%+ healthy
		return 0.01 // 1% risk
	case healthFactor >= 1.5: // 150%+ healthy
		return 0.05 // 5% risk
	case healthFactor >= 1.2: // 120%+ healthy
		return 0.10 // 10% risk
	case healthFactor >= 1.1: // 110%+ healthy
		return 0.20 // 20% risk
	case healthFactor >= 1.05: // 105%+ healthy
		return 0.40 // 40% risk
	case healthFactor >= 1.0: // 100%+ healthy
		return 0.60 // 60% risk
	case healthFactor >= 0.95: // 95%+ healthy
		return 0.80 // 80% risk
	default: // Below 95%
		return 0.95 // 95% risk
	}
}

// calculateMarketExposure calculates the market exposure percentage
func (ra *RiskAssessor) calculateMarketExposure(account *Account) float64 {
	if account.CollateralValue.Cmp(big.NewInt(0)) == 0 {
		return 0.0
	}

	// Calculate exposure as percentage of total portfolio
	// This is a simplified calculation
	exposure := float64(account.BorrowValue.Int64()) / float64(account.CollateralValue.Int64())
	
	// Convert to percentage
	return exposure * 100.0
}

// updateRiskFactors updates the risk factors for an assessment
func (ra *RiskAssessor) updateRiskFactors(assessment *RiskAssessment) {
	// Update collateral ratio factor
	if factor, exists := ra.riskFactors["collateral_ratio"]; exists {
		factor.Score = assessment.RiskScore.CollateralScore
		factor.IsTriggered = factor.Score >= factor.Threshold
		factor.LastUpdated = time.Now()
		assessment.RiskFactors[factor.ID] = factor
	}

	// Update liquidation proximity factor
	if factor, exists := ra.riskFactors["liquidation_proximity"]; exists {
		factor.Score = 100 - assessment.RiskScore.CreditScore // Inverse of credit score
		factor.IsTriggered = factor.Score >= factor.Threshold
		factor.LastUpdated = time.Now()
		assessment.RiskFactors[factor.ID] = factor
	}

	// Update other factors similarly
	for id, factor := range ra.riskFactors {
		if id == "collateral_ratio" || id == "liquidation_proximity" {
			continue // Already updated
		}
		
		// Map factor to appropriate score
		switch id {
		case "market_volatility":
			factor.Score = assessment.RiskScore.MarketScore
		case "borrow_history":
			factor.Score = assessment.RiskScore.CreditScore
		case "repayment_pattern":
			factor.Score = assessment.RiskScore.BehaviorScore
		case "liquidity_access":
			factor.Score = assessment.RiskScore.LiquidityScore
		}
		
		factor.IsTriggered = factor.Score >= factor.Threshold
		factor.LastUpdated = time.Now()
		assessment.RiskFactors[factor.ID] = factor
	}
}

// GetRiskAssessment returns a risk assessment for a user
func (ra *RiskAssessor) GetRiskAssessment(userID string) (*RiskAssessment, error) {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	assessment, exists := ra.assessments[userID]
	if !exists {
		return nil, &RiskAssessmentError{
			Operation: "GetRiskAssessment",
			Message:   ErrAssessmentNotFound.Error(),
			UserID:    userID,
		}
	}

	return assessment, nil
}

// GetHighRiskUsers returns users with high risk scores
func (ra *RiskAssessor) GetHighRiskUsers() []string {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	var highRiskUsers []string
	for userID, assessment := range ra.assessments {
		if assessment.RiskScore != nil && assessment.RiskScore.OverallScore > ra.alertThreshold {
			highRiskUsers = append(highRiskUsers, userID)
		}
	}

	return highRiskUsers
}

// GetRiskFactor returns a specific risk factor
func (ra *RiskAssessor) GetRiskFactor(factorID string) (*RiskFactor, error) {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	factor, exists := ra.riskFactors[factorID]
	if !exists {
		return nil, &RiskAssessmentError{
			Operation: "GetRiskFactor",
			Message:   "risk factor not found",
		}
	}

	return factor, nil
}

// UpdateRiskFactor updates a risk factor configuration
func (ra *RiskAssessor) UpdateRiskFactor(factorID string, weight float64, threshold int) error {
	if weight < 0.0 || weight > 1.0 {
		return &RiskAssessmentError{
			Operation: "UpdateRiskFactor",
			Message:   ErrInvalidWeight.Error(),
		}
	}

	if threshold < 0 || threshold > 100 {
		return &RiskAssessmentError{
			Operation: "UpdateRiskFactor",
			Message:   ErrInvalidThreshold.Error(),
		}
	}

	ra.mutex.Lock()
	defer ra.mutex.Unlock()

	factor, exists := ra.riskFactors[factorID]
	if !exists {
		return &RiskAssessmentError{
			Operation: "UpdateRiskFactor",
			Message:   "risk factor not found",
		}
	}

	factor.Weight = weight
	factor.Threshold = threshold
	factor.LastUpdated = time.Now()

	return nil
}

// GetRiskStats returns risk assessment statistics
func (ra *RiskAssessor) GetRiskStats() map[string]interface{} {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_assessments": len(ra.assessments),
		"total_factors":     len(ra.riskFactors),
		"update_interval":   ra.updateInterval.String(),
		"alert_threshold":   ra.alertThreshold,
	}

	// Count users by risk level
	riskLevelCounts := make(map[string]int)
	riskLevelCounts[string(RiskLevelLow)] = 0
	riskLevelCounts[string(RiskLevelMedium)] = 0
	riskLevelCounts[string(RiskLevelHigh)] = 0
	riskLevelCounts[string(RiskLevelCritical)] = 0

	// Calculate average risk scores
	totalScore := 0
	validAssessments := 0

	for _, assessment := range ra.assessments {
		if assessment.RiskScore != nil {
			riskLevelCounts[string(assessment.RiskScore.RiskLevel)]++
			totalScore += assessment.RiskScore.OverallScore
			validAssessments++
		}
	}

	stats["risk_level_distribution"] = riskLevelCounts
	
	if validAssessments > 0 {
		stats["average_risk_score"] = totalScore / validAssessments
	} else {
		stats["average_risk_score"] = 0
	}

	// Count triggered risk factors
	triggeredFactors := 0
	for _, factor := range ra.riskFactors {
		if factor.IsTriggered {
			triggeredFactors++
		}
	}
	stats["triggered_factors"] = triggeredFactors

	return stats
}
