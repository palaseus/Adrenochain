package risk

import (
	"errors"
	"math/big"
	"sort"
	"time"
)

// LiquidationTrigger represents the conditions that trigger liquidation
type LiquidationTrigger struct {
	ID          string
	Type        LiquidationTriggerType
	Threshold   *big.Float
	GracePeriod time.Duration
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LiquidationTriggerType represents the type of liquidation trigger
type LiquidationTriggerType int

const (
	MarginCall LiquidationTriggerType = iota
	HealthFactor
	VolatilitySpike
	CorrelationBreakdown
	LiquidityCrisis
	SmartContractRisk
)

// LiquidationEvent represents a liquidation event
type LiquidationEvent struct {
	ID              string
	PositionID      string
	UserID          string
	TriggerType     LiquidationTriggerType
	TriggerValue    *big.Float
	Threshold       *big.Float
	PositionValue   *big.Float
	DebtAmount      *big.Float
	CollateralValue *big.Float
	LiquidationFee  *big.Float
	Status          LiquidationStatus
	TriggeredAt     time.Time
	ProcessedAt     *time.Time
	CompletedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// LiquidationStatus represents the status of a liquidation event
type LiquidationStatus int

const (
	LiquidationTriggered LiquidationStatus = iota
	LiquidationProcessing
	LiquidationAuctioning
	LiquidationCompleted
	LiquidationFailed
	LiquidationCancelled
)

// Auction represents an auction for liquidated assets
type Auction struct {
	ID            string
	LiquidationID string
	AssetID       string
	AssetAmount   *big.Float
	StartingPrice *big.Float
	MinimumPrice  *big.Float
	CurrentPrice  *big.Float
	ReservePrice  *big.Float
	Status        AuctionStatus
	StartTime     time.Time
	EndTime       time.Time
	Bids          []*Bid
	Winner        *Bid
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AuctionStatus represents the status of an auction
type AuctionStatus int

const (
	AuctionPending AuctionStatus = iota
	AuctionActive
	AuctionEnded
	AuctionSold
	AuctionFailed
	AuctionCancelled
)

// Bid represents a bid in an auction
type Bid struct {
	ID        string
	AuctionID string
	BidderID  string
	Amount    *big.Float
	Price     *big.Float
	Timestamp time.Time
	IsValid   bool
	CreatedAt time.Time
}

// RecoveryMechanism represents a recovery mechanism for liquidated positions
type RecoveryMechanism struct {
	ID          string
	Type        RecoveryType
	Description string
	Parameters  map[string]*big.Float
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RecoveryType represents the type of recovery mechanism
type RecoveryType int

const (
	PartialRecovery RecoveryType = iota
	DebtRestructuring
	CollateralRelease
	PaymentPlan
	DebtForgiveness
)

// LiquidationEngine manages the liquidation process
type LiquidationEngine struct {
	triggers           map[string]*LiquidationTrigger
	events             map[string]*LiquidationEvent
	auctions           map[string]*Auction
	recoveryMechanisms map[string]*RecoveryMechanism
	riskManager        *AdvancedRiskManager
	insuranceManager   *InsuranceManager
	config             *LiquidationConfig
}

// LiquidationConfig represents the configuration for the liquidation engine
type LiquidationConfig struct {
	DefaultGracePeriod     time.Duration
	DefaultLiquidationFee  *big.Float
	MinimumAuctionDuration time.Duration
	MaximumAuctionDuration time.Duration
	BidIncrement           *big.Float
	AutoExtendThreshold    *big.Float
}

// NewLiquidationEngine creates a new liquidation engine
func NewLiquidationEngine(
	riskManager *AdvancedRiskManager,
	insuranceManager *InsuranceManager,
	config *LiquidationConfig,
) (*LiquidationEngine, error) {
	if riskManager == nil {
		return nil, errors.New("risk manager cannot be nil")
	}
	if insuranceManager == nil {
		return nil, errors.New("insurance manager cannot be nil")
	}
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	return &LiquidationEngine{
		triggers:           make(map[string]*LiquidationTrigger),
		events:             make(map[string]*LiquidationEvent),
		auctions:           make(map[string]*Auction),
		recoveryMechanisms: make(map[string]*RecoveryMechanism),
		riskManager:        riskManager,
		insuranceManager:   insuranceManager,
		config:             config,
	}, nil
}

// NewLiquidationConfig creates a new liquidation configuration
func NewLiquidationConfig(
	gracePeriod time.Duration,
	liquidationFee *big.Float,
	minAuctionDuration time.Duration,
	maxAuctionDuration time.Duration,
	bidIncrement *big.Float,
	autoExtendThreshold *big.Float,
) (*LiquidationConfig, error) {
	if liquidationFee == nil || liquidationFee.Sign() < 0 {
		return nil, errors.New("liquidation fee cannot be negative")
	}
	if minAuctionDuration <= 0 {
		return nil, errors.New("minimum auction duration must be positive")
	}
	if maxAuctionDuration <= minAuctionDuration {
		return nil, errors.New("maximum auction duration must be greater than minimum")
	}
	if bidIncrement == nil || bidIncrement.Sign() <= 0 {
		return nil, errors.New("bid increment must be positive")
	}
	if autoExtendThreshold == nil || autoExtendThreshold.Sign() < 0 {
		return nil, errors.New("auto-extend threshold cannot be negative")
	}

	return &LiquidationConfig{
		DefaultGracePeriod:     gracePeriod,
		DefaultLiquidationFee:  new(big.Float).Copy(liquidationFee),
		MinimumAuctionDuration: minAuctionDuration,
		MaximumAuctionDuration: maxAuctionDuration,
		BidIncrement:           new(big.Float).Copy(bidIncrement),
		AutoExtendThreshold:    new(big.Float).Copy(autoExtendThreshold),
	}, nil
}

// NewLiquidationTrigger creates a new liquidation trigger
func NewLiquidationTrigger(
	id string,
	triggerType LiquidationTriggerType,
	threshold *big.Float,
	gracePeriod time.Duration,
) (*LiquidationTrigger, error) {
	if id == "" {
		return nil, errors.New("trigger ID cannot be empty")
	}
	if threshold == nil || threshold.Sign() <= 0 {
		return nil, errors.New("threshold must be positive")
	}
	if gracePeriod <= 0 {
		return nil, errors.New("grace period must be positive")
	}

	now := time.Now()
	return &LiquidationTrigger{
		ID:          id,
		Type:        triggerType,
		Threshold:   new(big.Float).Copy(threshold),
		GracePeriod: gracePeriod,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewLiquidationEvent creates a new liquidation event
func NewLiquidationEvent(
	id, positionID, userID string,
	triggerType LiquidationTriggerType,
	triggerValue, threshold, positionValue, debtAmount, collateralValue *big.Float,
) (*LiquidationEvent, error) {
	if id == "" {
		return nil, errors.New("event ID cannot be empty")
	}
	if positionID == "" {
		return nil, errors.New("position ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if triggerValue == nil || threshold == nil || positionValue == nil || debtAmount == nil || collateralValue == nil {
		return nil, errors.New("all financial values must be provided")
	}

	now := time.Now()
	return &LiquidationEvent{
		ID:              id,
		PositionID:      positionID,
		UserID:          userID,
		TriggerType:     triggerType,
		TriggerValue:    new(big.Float).Copy(triggerValue),
		Threshold:       new(big.Float).Copy(threshold),
		PositionValue:   new(big.Float).Copy(positionValue),
		DebtAmount:      new(big.Float).Copy(debtAmount),
		CollateralValue: new(big.Float).Copy(collateralValue),
		LiquidationFee:  big.NewFloat(0), // Will be calculated
		Status:          LiquidationTriggered,
		TriggeredAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// NewAuction creates a new auction for liquidated assets
func NewAuction(
	id, liquidationID, assetID string,
	assetAmount, startingPrice, minimumPrice, reservePrice *big.Float,
	duration time.Duration,
) (*Auction, error) {
	if id == "" {
		return nil, errors.New("auction ID cannot be empty")
	}
	if liquidationID == "" {
		return nil, errors.New("liquidation ID cannot be empty")
	}
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if assetAmount == nil || assetAmount.Sign() <= 0 {
		return nil, errors.New("asset amount must be positive")
	}
	if startingPrice == nil || startingPrice.Sign() <= 0 {
		return nil, errors.New("starting price must be positive")
	}
	if minimumPrice == nil || minimumPrice.Sign() <= 0 {
		return nil, errors.New("minimum price must be positive")
	}
	if reservePrice == nil || reservePrice.Sign() <= 0 {
		return nil, errors.New("reserve price must be positive")
	}
	if duration <= 0 {
		return nil, errors.New("duration must be positive")
	}

	now := time.Now()
	startTime := now.Add(time.Minute) // Start in 1 minute
	endTime := startTime.Add(duration)

	return &Auction{
		ID:            id,
		LiquidationID: liquidationID,
		AssetID:       assetID,
		AssetAmount:   new(big.Float).Copy(assetAmount),
		StartingPrice: new(big.Float).Copy(startingPrice),
		MinimumPrice:  new(big.Float).Copy(minimumPrice),
		CurrentPrice:  new(big.Float).Copy(startingPrice),
		ReservePrice:  new(big.Float).Copy(reservePrice),
		Status:        AuctionPending,
		StartTime:     startTime,
		EndTime:       endTime,
		Bids:          make([]*Bid, 0),
		Winner:        nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// NewBid creates a new bid in an auction
func NewBid(id, auctionID, bidderID string, amount, price *big.Float) (*Bid, error) {
	if id == "" {
		return nil, errors.New("bid ID cannot be empty")
	}
	if auctionID == "" {
		return nil, errors.New("auction ID cannot be empty")
	}
	if bidderID == "" {
		return nil, errors.New("bidder ID cannot be empty")
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, errors.New("bid amount must be positive")
	}
	if price == nil || price.Sign() <= 0 {
		return nil, errors.New("bid price must be positive")
	}

	now := time.Now()
	return &Bid{
		ID:        id,
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    new(big.Float).Copy(amount),
		Price:     new(big.Float).Copy(price),
		Timestamp: now,
		IsValid:   true,
		CreatedAt: now,
	}, nil
}

// NewRecoveryMechanism creates a new recovery mechanism
func NewRecoveryMechanism(
	id string,
	recoveryType RecoveryType,
	description string,
	parameters map[string]*big.Float,
) (*RecoveryMechanism, error) {
	if id == "" {
		return nil, errors.New("mechanism ID cannot be empty")
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if parameters == nil {
		return nil, errors.New("parameters cannot be nil")
	}

	now := time.Now()
	return &RecoveryMechanism{
			ID:          id,
			Type:        recoveryType,
			Description: description,
			Parameters:  parameters,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		nil
}

// AddLiquidationTrigger adds a liquidation trigger to the engine
func (le *LiquidationEngine) AddLiquidationTrigger(trigger *LiquidationTrigger) error {
	if trigger == nil {
		return errors.New("trigger cannot be nil")
	}

	le.triggers[trigger.ID] = trigger
	return nil
}

// GetLiquidationTrigger retrieves a liquidation trigger by ID
func (le *LiquidationEngine) GetLiquidationTrigger(id string) (*LiquidationTrigger, error) {
	if id == "" {
		return nil, errors.New("trigger ID cannot be empty")
	}

	trigger, exists := le.triggers[id]
	if !exists {
		return nil, errors.New("trigger not found")
	}

	return trigger, nil
}

// AddLiquidationEvent adds a liquidation event to the engine
func (le *LiquidationEngine) AddLiquidationEvent(event *LiquidationEvent) error {
	if event == nil {
		return errors.New("event cannot be nil")
	}

	le.events[event.ID] = event
	return nil
}

// GetLiquidationEvent retrieves a liquidation event by ID
func (le *LiquidationEngine) GetLiquidationEvent(id string) (*LiquidationEvent, error) {
	if id == "" {
		return nil, errors.New("event ID cannot be empty")
	}

	event, exists := le.events[id]
	if !exists {
		return nil, errors.New("event not found")
	}

	return event, nil
}

// AddAuction adds an auction to the engine
func (le *LiquidationEngine) AddAuction(auction *Auction) error {
	if auction == nil {
		return errors.New("auction cannot be nil")
	}

	le.auctions[auction.ID] = auction
	return nil
}

// GetAuction retrieves an auction by ID
func (le *LiquidationEngine) GetAuction(id string) (*Auction, error) {
	if id == "" {
		return nil, errors.New("auction ID cannot be empty")
	}

	auction, exists := le.auctions[id]
	if !exists {
		return nil, errors.New("auction not found")
	}

	return auction, nil
}

// AddRecoveryMechanism adds a recovery mechanism to the engine
func (le *LiquidationEngine) AddRecoveryMechanism(mechanism *RecoveryMechanism) error {
	if mechanism == nil {
		return errors.New("mechanism cannot be nil")
	}

	le.recoveryMechanisms[mechanism.ID] = mechanism
	return nil
}

// GetRecoveryMechanism retrieves a recovery mechanism by ID
func (le *LiquidationEngine) GetRecoveryMechanism(id string) (*RecoveryMechanism, error) {
	if id == "" {
		return nil, errors.New("mechanism ID cannot be empty")
	}

	mechanism, exists := le.recoveryMechanisms[id]
	if !exists {
		return nil, errors.New("mechanism not found")
	}

	return mechanism, nil
}

// ProcessLiquidationEvent processes a liquidation event
func (le *LiquidationEngine) ProcessLiquidationEvent(eventID string) error {
	event, err := le.GetLiquidationEvent(eventID)
	if err != nil {
		return err
	}

	if event.Status != LiquidationTriggered {
		return errors.New("event is not in triggered status")
	}

	now := time.Now()
	event.Status = LiquidationProcessing
	event.ProcessedAt = &now
	event.UpdatedAt = now

	// Calculate liquidation fee
	le.calculateLiquidationFee(event)

	// Create auction for liquidated assets
	auction, err := le.createAuctionForEvent(event)
	if err != nil {
		event.Status = LiquidationFailed
		event.UpdatedAt = time.Now()
		return err
	}

	// Add auction to engine
	err = le.AddAuction(auction)
	if err != nil {
		event.Status = LiquidationFailed
		event.UpdatedAt = time.Now()
		return err
	}

	return nil
}

// calculateLiquidationFee calculates the liquidation fee for an event
func (le *LiquidationEngine) calculateLiquidationFee(event *LiquidationEvent) {
	// Base liquidation fee
	baseFee := new(big.Float).Mul(event.PositionValue, le.config.DefaultLiquidationFee)

	// Risk-based fee adjustment based on trigger type
	riskMultiplier := le.getRiskMultiplier(event.TriggerType)

	// Calculate final fee
	finalFee := new(big.Float).Mul(baseFee, riskMultiplier)

	// Cap the fee at 10% of position value
	maxFee := new(big.Float).Mul(event.PositionValue, big.NewFloat(0.10))
	if finalFee.Cmp(maxFee) > 0 {
		finalFee = maxFee
	}

	event.LiquidationFee = finalFee
}

// getRiskMultiplier returns the risk multiplier for a trigger type
func (le *LiquidationEngine) getRiskMultiplier(triggerType LiquidationTriggerType) *big.Float {
	switch triggerType {
	case MarginCall:
		return big.NewFloat(1.0)
	case HealthFactor:
		return big.NewFloat(1.2)
	case VolatilitySpike:
		return big.NewFloat(1.5)
	case CorrelationBreakdown:
		return big.NewFloat(1.8)
	case LiquidityCrisis:
		return big.NewFloat(2.0)
	case SmartContractRisk:
		return big.NewFloat(2.5)
	default:
		return big.NewFloat(1.0)
	}
}

// createAuctionForEvent creates an auction for a liquidation event
func (le *LiquidationEngine) createAuctionForEvent(event *LiquidationEvent) (*Auction, error) {
	// Calculate auction parameters
	startingPrice := new(big.Float).Mul(event.CollateralValue, big.NewFloat(0.8)) // Start at 80% of collateral value
	minimumPrice := new(big.Float).Mul(event.CollateralValue, big.NewFloat(0.5))  // Minimum 50% of collateral value
	reservePrice := new(big.Float).Mul(event.CollateralValue, big.NewFloat(0.6))  // Reserve 60% of collateral value

	// Determine auction duration based on asset value
	duration := le.calculateAuctionDuration(event.CollateralValue)

	auction, err := NewAuction(
		generateAuctionID(event.ID),
		event.ID,
		"collateral_asset", // This would be the actual asset ID
		event.CollateralValue,
		startingPrice,
		minimumPrice,
		reservePrice,
		duration,
	)

	return auction, err
}

// calculateAuctionDuration calculates the appropriate auction duration
func (le *LiquidationEngine) calculateAuctionDuration(collateralValue *big.Float) time.Duration {
	// Convert to float64 for calculation
	value, _ := collateralValue.Float64()

	// Base duration: 1 hour for every $10,000 in collateral value
	baseHours := value / 10000.0

	// Convert to duration
	duration := time.Duration(baseHours) * time.Hour

	// Apply min/max constraints
	if duration < le.config.MinimumAuctionDuration {
		duration = le.config.MinimumAuctionDuration
	}
	if duration > le.config.MaximumAuctionDuration {
		duration = le.config.MaximumAuctionDuration
	}

	return duration
}

// PlaceBid places a bid in an auction
func (le *LiquidationEngine) PlaceBid(auctionID, bidderID string, amount, price *big.Float) error {
	auction, err := le.GetAuction(auctionID)
	if err != nil {
		return err
	}

	if auction.Status != AuctionActive {
		return errors.New("auction is not active")
	}

	if time.Now().After(auction.EndTime) {
		return errors.New("auction has ended")
	}

	// Validate bid price
	if price.Cmp(auction.CurrentPrice) <= 0 {
		return errors.New("bid price must be higher than current price")
	}

	if price.Cmp(auction.MinimumPrice) < 0 {
		return errors.New("bid price is below minimum price")
	}

	// Create bid
	bid, err := NewBid(
		generateBidID(auctionID, bidderID),
		auctionID,
		bidderID,
		amount,
		price,
	)
	if err != nil {
		return err
	}

	// Add bid to auction
	auction.Bids = append(auction.Bids, bid)
	auction.CurrentPrice = new(big.Float).Copy(price)
	auction.UpdatedAt = time.Now()

	// Check if auction should be auto-extended
	le.checkAutoExtension(auction)

	return nil
}

// checkAutoExtension checks if an auction should be auto-extended
func (le *LiquidationEngine) checkAutoExtension(auction *Auction) {
	if len(auction.Bids) == 0 {
		return
	}

	// Get the last bid
	lastBid := auction.Bids[len(auction.Bids)-1]

	// Check if bid was placed within auto-extend threshold of end time
	timeToEnd := auction.EndTime.Sub(lastBid.Timestamp)
	thresholdFloat, _ := le.config.AutoExtendThreshold.Float64()
	threshold := time.Duration(thresholdFloat) * time.Minute

	if timeToEnd <= threshold {
		// Extend auction by 30 minutes
		auction.EndTime = auction.EndTime.Add(30 * time.Minute)
		auction.UpdatedAt = time.Now()
	}
}

// EndAuction ends an auction and determines the winner
func (le *LiquidationEngine) EndAuction(auctionID string) error {
	auction, err := le.GetAuction(auctionID)
	if err != nil {
		return err
	}

	if auction.Status != AuctionActive {
		return errors.New("auction is not active")
	}

	if time.Now().Before(auction.EndTime) {
		return errors.New("auction has not ended yet")
	}

	// Sort bids by price (highest first)
	sort.Slice(auction.Bids, func(i, j int) bool {
		return auction.Bids[i].Price.Cmp(auction.Bids[j].Price) > 0
	})

	// Find winning bid
	var winningBid *Bid
	for _, bid := range auction.Bids {
		if bid.IsValid && bid.Price.Cmp(auction.ReservePrice) >= 0 {
			winningBid = bid
			break
		}
	}

	if winningBid != nil {
		auction.Status = AuctionSold
		auction.Winner = winningBid
	} else {
		auction.Status = AuctionFailed
	}

	auction.UpdatedAt = time.Now()

	// Update liquidation event status
	le.updateLiquidationEventStatus(auction.LiquidationID, auction.Status)

	return nil
}

// updateLiquidationEventStatus updates the status of a liquidation event
func (le *LiquidationEngine) updateLiquidationEventStatus(liquidationID string, auctionStatus AuctionStatus) {
	// Find the liquidation event
	for _, event := range le.events {
		if event.ID == liquidationID {
			now := time.Now()
			event.UpdatedAt = now

			switch auctionStatus {
			case AuctionSold:
				event.Status = LiquidationCompleted
				event.CompletedAt = &now
			case AuctionFailed:
				event.Status = LiquidationFailed
			}
			break
		}
	}
}

// GetUserLiquidationEvents returns all liquidation events for a specific user
func (le *LiquidationEngine) GetUserLiquidationEvents(userID string) ([]*LiquidationEvent, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var userEvents []*LiquidationEvent
	for _, event := range le.events {
		if event.UserID == userID {
			userEvents = append(userEvents, event)
		}
	}

	return userEvents, nil
}

// GetActiveAuctions returns all active auctions
func (le *LiquidationEngine) GetActiveAuctions() []*Auction {
	var activeAuctions []*Auction
	for _, auction := range le.auctions {
		if auction.Status == AuctionActive {
			activeAuctions = append(activeAuctions, auction)
		}
	}
	return activeAuctions
}

// GetLiquidationStatistics returns statistics about liquidations
func (le *LiquidationEngine) GetLiquidationStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count events by status
	statusCounts := make(map[LiquidationStatus]int)
	for _, event := range le.events {
		statusCounts[event.Status]++
	}

	// Count auctions by status
	auctionStatusCounts := make(map[AuctionStatus]int)
	for _, auction := range le.auctions {
		auctionStatusCounts[auction.Status]++
	}

	// Calculate total liquidation fees
	totalFees := big.NewFloat(0)
	for _, event := range le.events {
		if event.LiquidationFee != nil {
			totalFees.Add(totalFees, event.LiquidationFee)
		}
	}

	stats["total_events"] = len(le.events)
	stats["events_by_status"] = statusCounts
	stats["total_auctions"] = len(le.auctions)
	stats["auctions_by_status"] = auctionStatusCounts
	stats["total_liquidation_fees"] = totalFees

	return stats
}

// generateAuctionID generates a unique auction ID
func generateAuctionID(liquidationID string) string {
	return "auction_" + liquidationID + "_" + time.Now().Format("20060102150405")
}

// generateBidID generates a unique bid ID
func generateBidID(auctionID, bidderID string) string {
	return "bid_" + auctionID + "_" + bidderID + "_" + time.Now().Format("20060102150405")
}
