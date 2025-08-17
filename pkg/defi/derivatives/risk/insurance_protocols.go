package risk

import (
	"errors"
	"math/big"
	"time"
)

// InsuranceType represents the type of insurance coverage
type InsuranceType int

const (
	PortfolioInsurance InsuranceType = iota
	LiquidationInsurance
	SmartContractInsurance
	CrossChainInsurance
	YieldInsurance
)

// CoverageStatus represents the status of insurance coverage
type CoverageStatus int

const (
	Active CoverageStatus = iota
	Expired
	Cancelled
	Suspended
	UnderReview
)

// ClaimStatus represents the status of an insurance claim
type ClaimStatus int

const (
	Pending ClaimStatus = iota
	Approved
	Rejected
	UnderInvestigation
	Paid
)

// InsurancePool represents an insurance coverage pool
type InsurancePool struct {
	ID                string
	Name              string
	Type              InsuranceType
	Description       string
	TotalCapacity     *big.Float
	UsedCapacity      *big.Float
	AvailableCapacity *big.Float
	PremiumRate       *big.Float
	CoverageLimit     *big.Float
	MinCoverage       *big.Float
	MaxCoverage       *big.Float
	RiskScore         *big.Float
	Status            CoverageStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         time.Time
}

// InsurancePolicy represents an individual insurance policy
type InsurancePolicy struct {
	ID              string
	PoolID          string
	UserID          string
	CoverageAmount  *big.Float
	PremiumAmount   *big.Float
	PremiumRate     *big.Float
	CoverageType    InsuranceType
	RiskAssessment  *big.Float
	Status          CoverageStatus
	StartDate       time.Time
	EndDate         time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// InsuranceClaim represents a claim against an insurance policy
type InsuranceClaim struct {
	ID              string
	PolicyID        string
	UserID          string
	ClaimAmount     *big.Float
	Description     string
	Evidence        []string
	Status          ClaimStatus
	SubmittedAt     time.Time
	ProcessedAt     *time.Time
	ApprovedAt      *time.Time
	PaidAt          *time.Time
	RejectionReason *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PremiumCalculation represents the calculation of insurance premiums
type PremiumCalculation struct {
	BasePremium     *big.Float
	RiskMultiplier  *big.Float
	CoverageMultiplier *big.Float
	DurationMultiplier *big.Float
	FinalPremium    *big.Float
	RiskFactors     map[string]*big.Float
	CreatedAt       time.Time
}

// RiskAssessment represents a comprehensive risk assessment for insurance
type RiskAssessment struct {
	ID              string
	UserID          string
	PortfolioID     string
	RiskScore       *big.Float
	VaR             *big.Float
	CVaR            *big.Float
	Volatility      *big.Float
	Leverage        *big.Float
	LiquidityScore  *big.Float
	CorrelationRisk *big.Float
	MarketRisk      *big.Float
	CreditRisk      *big.Float
	OperationalRisk *big.Float
	OverallRisk     *big.Float
	Recommendations []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// InsuranceManager manages insurance pools, policies, and claims
type InsuranceManager struct {
	pools           map[string]*InsurancePool
	policies        map[string]*InsurancePolicy
	claims          map[string]*InsuranceClaim
	assessments     map[string]*RiskAssessment
	riskFreeRate    *big.Float
	basePremiumRate *big.Float
}

// NewInsuranceManager creates a new insurance manager
func NewInsuranceManager(riskFreeRate, basePremiumRate *big.Float) (*InsuranceManager, error) {
	if riskFreeRate == nil || basePremiumRate == nil {
		return nil, errors.New("risk-free rate and base premium rate cannot be nil")
	}

	if basePremiumRate.Sign() < 0 {
		return nil, errors.New("base premium rate cannot be negative")
	}

	return &InsuranceManager{
		pools:           make(map[string]*InsurancePool),
		policies:        make(map[string]*InsurancePolicy),
		claims:          make(map[string]*InsuranceClaim),
		assessments:     make(map[string]*RiskAssessment),
		riskFreeRate:    new(big.Float).Copy(riskFreeRate),
		basePremiumRate: new(big.Float).Copy(basePremiumRate),
	}, nil
}

// NewInsurancePool creates a new insurance coverage pool
func NewInsurancePool(
	id, name string,
	insuranceType InsuranceType,
	description string,
	totalCapacity, premiumRate, coverageLimit, minCoverage, maxCoverage *big.Float,
	expiresAt time.Time,
) (*InsurancePool, error) {
	if id == "" {
		return nil, errors.New("pool ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("pool name cannot be empty")
	}
	if totalCapacity == nil || totalCapacity.Sign() <= 0 {
		return nil, errors.New("total capacity must be positive")
	}
	if premiumRate == nil || premiumRate.Sign() < 0 {
		return nil, errors.New("premium rate cannot be negative")
	}
	if coverageLimit == nil || coverageLimit.Sign() <= 0 {
		return nil, errors.New("coverage limit must be positive")
	}
	if minCoverage == nil || minCoverage.Sign() < 0 {
		return nil, errors.New("minimum coverage cannot be negative")
	}
	if maxCoverage == nil || maxCoverage.Sign() <= 0 {
		return nil, errors.New("maximum coverage must be positive")
	}
	if minCoverage.Cmp(maxCoverage) >= 0 {
		return nil, errors.New("minimum coverage must be less than maximum coverage")
	}
	if maxCoverage.Cmp(totalCapacity) > 0 {
		return nil, errors.New("maximum coverage cannot exceed total capacity")
	}

	now := time.Now()
	if expiresAt.Before(now) {
		return nil, errors.New("expiration date cannot be in the past")
	}

	pool := &InsurancePool{
		ID:                id,
		Name:              name,
		Type:              insuranceType,
		Description:       description,
		TotalCapacity:     new(big.Float).Copy(totalCapacity),
		UsedCapacity:      big.NewFloat(0),
		AvailableCapacity: new(big.Float).Copy(totalCapacity),
		PremiumRate:       new(big.Float).Copy(premiumRate),
		CoverageLimit:     new(big.Float).Copy(coverageLimit),
		MinCoverage:       new(big.Float).Copy(minCoverage),
		MaxCoverage:       new(big.Float).Copy(maxCoverage),
		RiskScore:         big.NewFloat(0),
		Status:            Active,
		CreatedAt:         now,
		UpdatedAt:         now,
		ExpiresAt:         expiresAt,
	}

	return pool, nil
}

// NewInsurancePolicy creates a new insurance policy
func NewInsurancePolicy(
	id, poolID, userID string,
	coverageAmount *big.Float,
	coverageType InsuranceType,
	startDate, endDate time.Time,
) (*InsurancePolicy, error) {
	if id == "" {
		return nil, errors.New("policy ID cannot be empty")
	}
	if poolID == "" {
		return nil, errors.New("pool ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if coverageAmount == nil || coverageAmount.Sign() <= 0 {
		return nil, errors.New("coverage amount must be positive")
	}
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}
	if startDate.Before(time.Now()) {
		return nil, errors.New("start date cannot be in the past")
	}

	now := time.Now()
	policy := &InsurancePolicy{
		ID:             id,
		PoolID:         poolID,
		UserID:         userID,
		CoverageAmount: new(big.Float).Copy(coverageAmount),
		PremiumAmount:  big.NewFloat(0), // Will be calculated
		PremiumRate:    big.NewFloat(0), // Will be calculated
		CoverageType:   coverageType,
		RiskAssessment: big.NewFloat(0), // Will be calculated
		Status:         Active,
		StartDate:      startDate,
		EndDate:        endDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return policy, nil
}

// NewInsuranceClaim creates a new insurance claim
func NewInsuranceClaim(
	id, policyID, userID string,
	claimAmount *big.Float,
	description string,
	evidence []string,
) (*InsuranceClaim, error) {
	if id == "" {
		return nil, errors.New("claim ID cannot be empty")
	}
	if policyID == "" {
		return nil, errors.New("policy ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if claimAmount == nil || claimAmount.Sign() <= 0 {
		return nil, errors.New("claim amount must be positive")
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}

	now := time.Now()
	claim := &InsuranceClaim{
		ID:          id,
		PolicyID:    policyID,
		UserID:      userID,
		ClaimAmount: new(big.Float).Copy(claimAmount),
		Description: description,
		Evidence:    evidence,
		Status:      Pending,
		SubmittedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return claim, nil
}

// NewRiskAssessment creates a new risk assessment
func NewRiskAssessment(
	id, userID, portfolioID string,
	riskScore, varValue, cvarValue, volatility, leverage, liquidityScore, correlationRisk, marketRisk, creditRisk, operationalRisk *big.Float,
) (*RiskAssessment, error) {
	if id == "" {
		return nil, errors.New("assessment ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if portfolioID == "" {
		return nil, errors.New("portfolio ID cannot be empty")
	}

	now := time.Now()
	assessment := &RiskAssessment{
		ID:              id,
		UserID:          userID,
		PortfolioID:     portfolioID,
		RiskScore:       new(big.Float).Copy(riskScore),
		VaR:             new(big.Float).Copy(varValue),
		CVaR:            new(big.Float).Copy(cvarValue),
		Volatility:      new(big.Float).Copy(volatility),
		Leverage:        new(big.Float).Copy(leverage),
		LiquidityScore:  new(big.Float).Copy(liquidityScore),
		CorrelationRisk: new(big.Float).Copy(correlationRisk),
		MarketRisk:      new(big.Float).Copy(marketRisk),
		CreditRisk:      new(big.Float).Copy(creditRisk),
		OperationalRisk: new(big.Float).Copy(operationalRisk),
		OverallRisk:     big.NewFloat(0), // Will be calculated
		Recommendations: make([]string, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate overall risk score
	assessment.calculateOverallRisk()

	return assessment, nil
}

// calculateOverallRisk calculates the overall risk score based on individual risk factors
func (ra *RiskAssessment) calculateOverallRisk() {
	// Weighted average of risk factors
	weights := map[string]*big.Float{
		"var":             big.NewFloat(0.25),
		"cvar":            big.NewFloat(0.20),
		"volatility":      big.NewFloat(0.15),
		"leverage":        big.NewFloat(0.15),
		"liquidity":       big.NewFloat(0.10),
		"correlation":     big.NewFloat(0.05),
		"market":          big.NewFloat(0.05),
		"credit":          big.NewFloat(0.03),
		"operational":     big.NewFloat(0.02),
	}

	overallRisk := big.NewFloat(0)
	
	// Apply weights to each risk factor
	if ra.VaR != nil {
		weighted := new(big.Float).Mul(ra.VaR, weights["var"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.CVaR != nil {
		weighted := new(big.Float).Mul(ra.CVaR, weights["cvar"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.Volatility != nil {
		weighted := new(big.Float).Mul(ra.Volatility, weights["volatility"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.Leverage != nil {
		weighted := new(big.Float).Mul(ra.Leverage, weights["leverage"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.LiquidityScore != nil {
		weighted := new(big.Float).Mul(ra.LiquidityScore, weights["liquidity"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.CorrelationRisk != nil {
		weighted := new(big.Float).Mul(ra.CorrelationRisk, weights["correlation"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.MarketRisk != nil {
		weighted := new(big.Float).Mul(ra.MarketRisk, weights["market"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.CreditRisk != nil {
		weighted := new(big.Float).Mul(ra.CreditRisk, weights["credit"])
		overallRisk.Add(overallRisk, weighted)
	}
	
	if ra.OperationalRisk != nil {
		weighted := new(big.Float).Mul(ra.OperationalRisk, weights["operational"])
		overallRisk.Add(overallRisk, weighted)
	}

	ra.OverallRisk = overallRisk
	ra.UpdatedAt = time.Now()
}

// CalculatePremium calculates the insurance premium for a policy
func (im *InsuranceManager) CalculatePremium(
	policy *InsurancePolicy,
	pool *InsurancePool,
	assessment *RiskAssessment,
) (*PremiumCalculation, error) {
	if policy == nil {
		return nil, errors.New("policy cannot be nil")
	}
	if pool == nil {
		return nil, errors.New("pool cannot be nil")
	}
	if assessment == nil {
		return nil, errors.New("assessment cannot be nil")
	}

	// Base premium from pool
	basePremium := new(big.Float).Mul(pool.PremiumRate, policy.CoverageAmount)

	// Risk multiplier based on risk assessment
	riskMultiplier := im.calculateRiskMultiplier(assessment)

	// Coverage multiplier based on coverage amount relative to pool limits
	coverageMultiplier := im.calculateCoverageMultiplier(policy, pool)

	// Duration multiplier based on policy duration
	durationMultiplier := im.calculateDurationMultiplier(policy)

	// Calculate final premium
	finalPremium := new(big.Float).Mul(basePremium, riskMultiplier)
	finalPremium.Mul(finalPremium, coverageMultiplier)
	finalPremium.Mul(finalPremium, durationMultiplier)

	// Store risk factors for transparency
	riskFactors := map[string]*big.Float{
		"base_risk":      pool.PremiumRate,
		"risk_multiplier": riskMultiplier,
		"coverage_multiplier": coverageMultiplier,
		"duration_multiplier": durationMultiplier,
		"overall_risk_score": assessment.OverallRisk,
	}

	calculation := &PremiumCalculation{
		BasePremium:        basePremium,
		RiskMultiplier:     riskMultiplier,
		CoverageMultiplier: coverageMultiplier,
		DurationMultiplier: durationMultiplier,
		FinalPremium:       finalPremium,
		RiskFactors:        riskFactors,
		CreatedAt:          time.Now(),
	}

	// Update policy with calculated values
	policy.PremiumAmount = new(big.Float).Copy(finalPremium)
	policy.PremiumRate = new(big.Float).Quo(finalPremium, policy.CoverageAmount)
	policy.RiskAssessment = new(big.Float).Copy(assessment.OverallRisk)
	policy.UpdatedAt = time.Now()

	return calculation, nil
}

// calculateRiskMultiplier calculates the risk multiplier based on risk assessment
func (im *InsuranceManager) calculateRiskMultiplier(assessment *RiskAssessment) *big.Float {
	if assessment.OverallRisk == nil {
		return big.NewFloat(1.0)
	}

	// Convert to float64 for calculation
	riskScore, _ := assessment.OverallRisk.Float64()
	
	// Risk multiplier formula: 1 + (risk_score * 0.5)
	// This means a risk score of 0.1 results in a 1.05x multiplier
	multiplier := 1.0 + (riskScore * 0.5)
	
	// Cap the multiplier at 3.0x to prevent excessive premiums
	if multiplier > 3.0 {
		multiplier = 3.0
	}
	
	// Minimum multiplier of 0.5x for very low risk
	if multiplier < 0.5 {
		multiplier = 0.5
	}

	return big.NewFloat(multiplier)
}

// calculateCoverageMultiplier calculates the coverage multiplier based on coverage amount
func (im *InsuranceManager) calculateCoverageMultiplier(policy *InsurancePolicy, pool *InsurancePool) *big.Float {
	// Calculate coverage ratio relative to pool capacity
	coverageRatio := new(big.Float).Quo(policy.CoverageAmount, pool.TotalCapacity)
	
	// Convert to float64 for calculation
	ratio, _ := coverageRatio.Float64()
	
	// Coverage multiplier formula: 1 + (coverage_ratio * 0.3)
	// This means covering 10% of pool capacity results in a 1.03x multiplier
	multiplier := 1.0 + (ratio * 0.3)
	
	// Cap the multiplier at 2.0x
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	return big.NewFloat(multiplier)
}

// calculateDurationMultiplier calculates the duration multiplier based on policy duration
func (im *InsuranceManager) calculateDurationMultiplier(policy *InsurancePolicy) *big.Float {
	// Calculate policy duration in days
	duration := policy.EndDate.Sub(policy.StartDate)
	days := duration.Hours() / 24
	
	// Duration multiplier formula: 1 + (days / 365 * 0.2)
	// This means a 1-year policy gets a 1.2x multiplier
	multiplier := 1.0 + (days / 365.0 * 0.2)
	
	// Cap the multiplier at 1.5x
	if multiplier > 1.5 {
		multiplier = 1.5
	}

	return big.NewFloat(multiplier)
}

// AddInsurancePool adds an insurance pool to the manager
func (im *InsuranceManager) AddInsurancePool(pool *InsurancePool) error {
	if pool == nil {
		return errors.New("pool cannot be nil")
	}

	im.pools[pool.ID] = pool
	return nil
}

// GetInsurancePool retrieves an insurance pool by ID
func (im *InsuranceManager) GetInsurancePool(id string) (*InsurancePool, error) {
	if id == "" {
		return nil, errors.New("pool ID cannot be empty")
	}

	pool, exists := im.pools[id]
	if !exists {
		return nil, errors.New("pool not found")
	}

	return pool, nil
}

// AddInsurancePolicy adds an insurance policy to the manager
func (im *InsuranceManager) AddInsurancePolicy(policy *InsurancePolicy) error {
	if policy == nil {
		return errors.New("policy cannot be nil")
	}

	im.policies[policy.ID] = policy
	return nil
}

// GetInsurancePolicy retrieves an insurance policy by ID
func (im *InsuranceManager) GetInsurancePolicy(id string) (*InsurancePolicy, error) {
	if id == "" {
		return nil, errors.New("policy ID cannot be empty")
	}

	policy, exists := im.policies[id]
	if !exists {
		return nil, errors.New("policy not found")
	}

	return policy, nil
}

// AddInsuranceClaim adds an insurance claim to the manager
func (im *InsuranceManager) AddInsuranceClaim(claim *InsuranceClaim) error {
	if claim == nil {
		return errors.New("claim cannot be nil")
	}

	im.claims[claim.ID] = claim
	return nil
}

// GetInsuranceClaim retrieves an insurance claim by ID
func (im *InsuranceManager) GetInsuranceClaim(id string) (*InsuranceClaim, error) {
	if id == "" {
		return nil, errors.New("claim ID cannot be empty")
	}

	claim, exists := im.claims[id]
	if !exists {
		return nil, errors.New("claim not found")
	}

	return claim, nil
}

// AddRiskAssessment adds a risk assessment to the manager
func (im *InsuranceManager) AddRiskAssessment(assessment *RiskAssessment) error {
	if assessment == nil {
		return errors.New("assessment cannot be nil")
	}

	im.assessments[assessment.ID] = assessment
	return nil
}

// GetRiskAssessment retrieves a risk assessment by ID
func (im *InsuranceManager) GetRiskAssessment(id string) (*RiskAssessment, error) {
	if id == "" {
		return nil, errors.New("assessment ID cannot be empty")
	}

	assessment, exists := im.assessments[id]
	if !exists {
		return nil, errors.New("assessment not found")
	}

	return assessment, nil
}

// ProcessClaim processes an insurance claim
func (im *InsuranceManager) ProcessClaim(claimID string, approved bool, rejectionReason *string) error {
	claim, err := im.GetInsuranceClaim(claimID)
	if err != nil {
		return err
	}

	now := time.Now()
	claim.ProcessedAt = &now

	if approved {
		claim.Status = Approved
		claim.ApprovedAt = &now
	} else {
		claim.Status = Rejected
		claim.RejectionReason = rejectionReason
	}

	claim.UpdatedAt = now
	return nil
}

// PayClaim marks a claim as paid
func (im *InsuranceManager) PayClaim(claimID string) error {
	claim, err := im.GetInsuranceClaim(claimID)
	if err != nil {
		return err
	}

	if claim.Status != Approved {
		return errors.New("claim must be approved before payment")
	}

	now := time.Now()
	claim.Status = Paid
	claim.PaidAt = &now
	claim.UpdatedAt = now

	return nil
}

// GetPoolUtilization returns the utilization statistics for a pool
func (im *InsuranceManager) GetPoolUtilization(poolID string) (map[string]*big.Float, error) {
	pool, err := im.GetInsurancePool(poolID)
	if err != nil {
		return nil, err
	}

	utilization := map[string]*big.Float{
		"total_capacity":     new(big.Float).Copy(pool.TotalCapacity),
		"used_capacity":      new(big.Float).Copy(pool.UsedCapacity),
		"available_capacity": new(big.Float).Copy(pool.AvailableCapacity),
		"utilization_rate":   new(big.Float).Quo(pool.UsedCapacity, pool.TotalCapacity),
	}

	return utilization, nil
}

// GetUserPolicies returns all policies for a specific user
func (im *InsuranceManager) GetUserPolicies(userID string) ([]*InsurancePolicy, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var userPolicies []*InsurancePolicy
	for _, policy := range im.policies {
		if policy.UserID == userID {
			userPolicies = append(userPolicies, policy)
		}
	}

	return userPolicies, nil
}

// GetUserClaims returns all claims for a specific user
func (im *InsuranceManager) GetUserClaims(userID string) ([]*InsuranceClaim, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var userClaims []*InsuranceClaim
	for _, claim := range im.claims {
		if claim.UserID == userID {
			userClaims = append(userClaims, claim)
		}
	}

	return userClaims, nil
}
