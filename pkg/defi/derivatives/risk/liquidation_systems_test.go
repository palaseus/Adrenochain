package risk

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLiquidationConfig(t *testing.T) {
	t.Run("valid_config", func(t *testing.T) {
		gracePeriod := 24 * time.Hour
		liquidationFee := big.NewFloat(0.05) // 5%
		minAuctionDuration := 1 * time.Hour
		maxAuctionDuration := 7 * 24 * time.Hour
		bidIncrement := big.NewFloat(0.01) // 1%
		autoExtendThreshold := big.NewFloat(0.1) // 10%

		config, err := NewLiquidationConfig(
			gracePeriod,
			liquidationFee,
			minAuctionDuration,
			maxAuctionDuration,
			bidIncrement,
			autoExtendThreshold,
		)

		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, gracePeriod, config.DefaultGracePeriod)
		assert.Equal(t, liquidationFee, config.DefaultLiquidationFee)
		assert.Equal(t, minAuctionDuration, config.MinimumAuctionDuration)
		assert.Equal(t, maxAuctionDuration, config.MaximumAuctionDuration)
		assert.Equal(t, bidIncrement, config.BidIncrement)
		assert.Equal(t, autoExtendThreshold, config.AutoExtendThreshold)
	})

	t.Run("negative_liquidation_fee", func(t *testing.T) {
		gracePeriod := 24 * time.Hour
		liquidationFee := big.NewFloat(-0.05) // negative fee
		minAuctionDuration := 1 * time.Hour
		maxAuctionDuration := 7 * 24 * time.Hour
		bidIncrement := big.NewFloat(0.01)
		autoExtendThreshold := big.NewFloat(0.1)

		config, err := NewLiquidationConfig(
			gracePeriod,
			liquidationFee,
			minAuctionDuration,
			maxAuctionDuration,
			bidIncrement,
			autoExtendThreshold,
		)

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "liquidation fee cannot be negative")
	})

	t.Run("nil_liquidation_fee", func(t *testing.T) {
		gracePeriod := 24 * time.Hour
		minAuctionDuration := 1 * time.Hour
		maxAuctionDuration := 7 * 24 * time.Hour
		bidIncrement := big.NewFloat(0.01)
		autoExtendThreshold := big.NewFloat(0.1)

		config, err := NewLiquidationConfig(
			gracePeriod,
			nil, // nil fee
			minAuctionDuration,
			maxAuctionDuration,
			bidIncrement,
			autoExtendThreshold,
		)

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "liquidation fee cannot be negative")
	})
}

func TestNewLiquidationEngine(t *testing.T) {
	// Create dependencies
	riskFreeRate := big.NewFloat(0.02) // 2%
	confidenceLevel := big.NewFloat(0.95) // 95%
	basePremiumRate := big.NewFloat(0.01) // 1%
	
	riskManager, err := NewAdvancedRiskManager(riskFreeRate, confidenceLevel)
	require.NoError(t, err)
	
	insuranceManager, err := NewInsuranceManager(riskFreeRate, basePremiumRate)
	require.NoError(t, err)

	config, err := NewLiquidationConfig(
		24*time.Hour,
		big.NewFloat(0.05),
		1*time.Hour,
		7*24*time.Hour,
		big.NewFloat(0.01),
		big.NewFloat(0.1),
	)
	require.NoError(t, err)

	t.Run("valid_engine", func(t *testing.T) {
		engine, err := NewLiquidationEngine(riskManager, insuranceManager, config)

		assert.NoError(t, err)
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.triggers)
		assert.NotNil(t, engine.events)
		assert.NotNil(t, engine.auctions)
		assert.NotNil(t, engine.recoveryMechanisms)
		assert.Equal(t, riskManager, engine.riskManager)
		assert.Equal(t, insuranceManager, engine.insuranceManager)
		assert.Equal(t, config, engine.config)
	})

	t.Run("nil_risk_manager", func(t *testing.T) {
		engine, err := NewLiquidationEngine(nil, insuranceManager, config)

		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "risk manager cannot be nil")
	})

	t.Run("nil_insurance_manager", func(t *testing.T) {
		engine, err := NewLiquidationEngine(riskManager, nil, config)

		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "insurance manager cannot be nil")
	})

	t.Run("nil_config", func(t *testing.T) {
		engine, err := NewLiquidationEngine(riskManager, insuranceManager, nil)

		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestNewLiquidationTrigger(t *testing.T) {
	t.Run("valid_trigger", func(t *testing.T) {
		triggerType := MarginCall
		threshold := big.NewFloat(1.5) // 150%
		gracePeriod := 24 * time.Hour

		trigger, err := NewLiquidationTrigger("trigger1", triggerType, threshold, gracePeriod)

		assert.NoError(t, err)
		assert.NotNil(t, trigger)
		assert.Equal(t, "trigger1", trigger.ID)
		assert.Equal(t, triggerType, trigger.Type)
		assert.Equal(t, threshold, trigger.Threshold)
		assert.Equal(t, gracePeriod, trigger.GracePeriod)
		assert.True(t, trigger.IsActive)
		assert.False(t, trigger.CreatedAt.IsZero())
		assert.False(t, trigger.UpdatedAt.IsZero())
	})

	t.Run("empty_id", func(t *testing.T) {
		triggerType := MarginCall
		threshold := big.NewFloat(1.5)
		gracePeriod := 24 * time.Hour

		trigger, err := NewLiquidationTrigger("", triggerType, threshold, gracePeriod)

		assert.Error(t, err)
		assert.Nil(t, trigger)
		assert.Contains(t, err.Error(), "trigger ID cannot be empty")
	})

	t.Run("nil_threshold", func(t *testing.T) {
		triggerType := MarginCall
		gracePeriod := 24 * time.Hour

		trigger, err := NewLiquidationTrigger("trigger1", triggerType, nil, gracePeriod)

		assert.Error(t, err)
		assert.Nil(t, trigger)
		assert.Contains(t, err.Error(), "threshold must be positive")
	})
}

func TestNewLiquidationEvent(t *testing.T) {
	t.Run("valid_event", func(t *testing.T) {
		positionID := "pos123"
		userID := "user456"
		triggerType := MarginCall
		triggerValue := big.NewFloat(1.2)
		threshold := big.NewFloat(1.5)
		positionValue := big.NewFloat(100000)
		debtAmount := big.NewFloat(80000)
		collateralValue := big.NewFloat(90000)

		event, err := NewLiquidationEvent(
			"event1",
			positionID,
			userID,
			triggerType,
			triggerValue,
			threshold,
			positionValue,
			debtAmount,
			collateralValue,
		)

		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, "event1", event.ID)
		assert.Equal(t, positionID, event.PositionID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, triggerType, event.TriggerType)
		assert.Equal(t, triggerValue, event.TriggerValue)
		assert.Equal(t, threshold, event.Threshold)
		assert.Equal(t, positionValue, event.PositionValue)
		assert.Equal(t, debtAmount, event.DebtAmount)
		assert.Equal(t, collateralValue, event.CollateralValue)
		assert.NotNil(t, event.LiquidationFee)
		assert.Equal(t, LiquidationTriggered, event.Status)
		assert.False(t, event.TriggeredAt.IsZero())
		assert.False(t, event.CreatedAt.IsZero())
		assert.False(t, event.UpdatedAt.IsZero())
		assert.Nil(t, event.ProcessedAt)
		assert.Nil(t, event.CompletedAt)
	})

	t.Run("empty_id", func(t *testing.T) {
		event, err := NewLiquidationEvent(
			"",
			"pos123",
			"user456",
			MarginCall,
			big.NewFloat(1.2),
			big.NewFloat(1.5),
			big.NewFloat(100000),
			big.NewFloat(80000),
			big.NewFloat(90000),
		)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "event ID cannot be empty")
	})

	t.Run("empty_position_id", func(t *testing.T) {
		event, err := NewLiquidationEvent(
			"event1",
			"",
			"user456",
			MarginCall,
			big.NewFloat(1.2),
			big.NewFloat(1.5),
			big.NewFloat(100000),
			big.NewFloat(80000),
			big.NewFloat(90000),
		)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "position ID cannot be empty")
	})

	t.Run("empty_user_id", func(t *testing.T) {
		event, err := NewLiquidationEvent(
			"event1",
			"pos123",
			"",
			MarginCall,
			big.NewFloat(1.2),
			big.NewFloat(1.5),
			big.NewFloat(100000),
			big.NewFloat(80000),
			big.NewFloat(90000),
		)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})
}

func TestNewAuction(t *testing.T) {
	t.Run("valid_auction", func(t *testing.T) {
		liquidationID := "liq123"
		assetID := "asset456"
		assetAmount := big.NewFloat(1000)
		startingPrice := big.NewFloat(100)
		minimumPrice := big.NewFloat(80)
		reservePrice := big.NewFloat(90)
		duration := 24 * time.Hour

		auction, err := NewAuction(
			"auction1",
			liquidationID,
			assetID,
			assetAmount,
			startingPrice,
			minimumPrice,
			reservePrice,
			duration,
		)

		assert.NoError(t, err)
		assert.NotNil(t, auction)
		assert.Equal(t, "auction1", auction.ID)
		assert.Equal(t, liquidationID, auction.LiquidationID)
		assert.Equal(t, assetID, auction.AssetID)
		assert.Equal(t, assetAmount, auction.AssetAmount)
		assert.Equal(t, startingPrice, auction.StartingPrice)
		assert.Equal(t, minimumPrice, auction.MinimumPrice)
		assert.Equal(t, startingPrice, auction.CurrentPrice) // Should start at starting price
		assert.Equal(t, reservePrice, auction.ReservePrice)
		assert.Equal(t, AuctionPending, auction.Status)
		assert.False(t, auction.StartTime.IsZero())
		assert.False(t, auction.EndTime.IsZero())
		assert.True(t, auction.EndTime.After(auction.StartTime))
		assert.NotNil(t, auction.Bids)
		assert.Empty(t, auction.Bids)
		assert.Nil(t, auction.Winner)
		assert.False(t, auction.CreatedAt.IsZero())
		assert.False(t, auction.UpdatedAt.IsZero())
	})

	t.Run("empty_id", func(t *testing.T) {
		auction, err := NewAuction(
			"",
			"liq123",
			"asset456",
			big.NewFloat(1000),
			big.NewFloat(100),
			big.NewFloat(80),
			big.NewFloat(90),
			24*time.Hour,
		)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "auction ID cannot be empty")
	})

	t.Run("nil_asset_amount", func(t *testing.T) {
		auction, err := NewAuction(
			"auction1",
			"liq123",
			"asset456",
			nil,
			big.NewFloat(100),
			big.NewFloat(80),
			big.NewFloat(90),
			24*time.Hour,
		)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "asset amount must be positive")
	})
}

func TestNewBid(t *testing.T) {
	t.Run("valid_bid", func(t *testing.T) {
		auctionID := "auction123"
		bidderID := "bidder456"
		bidPrice := big.NewFloat(105)
		bidAmount := big.NewFloat(500)

		bid, err := NewBid("bid1", auctionID, bidderID, bidAmount, bidPrice)

		assert.NoError(t, err)
		assert.NotNil(t, bid)
		assert.Equal(t, "bid1", bid.ID)
		assert.Equal(t, auctionID, bid.AuctionID)
		assert.Equal(t, bidderID, bid.BidderID)
		assert.Equal(t, bidPrice, bid.Price)
		assert.Equal(t, bidAmount, bid.Amount)
		assert.True(t, bid.IsValid)
		assert.False(t, bid.CreatedAt.IsZero())
	})

	t.Run("empty_id", func(t *testing.T) {
		bid, err := NewBid(
			"",
			"auction123",
			"bidder456",
			big.NewFloat(500),
			big.NewFloat(105),
		)

		assert.Error(t, err)
		assert.Nil(t, bid)
		assert.Contains(t, err.Error(), "bid ID cannot be empty")
	})

	t.Run("empty_auction_id", func(t *testing.T) {
		bid, err := NewBid(
			"bid1",
			"",
			"bidder456",
			big.NewFloat(500),
			big.NewFloat(105),
		)

		assert.Error(t, err)
		assert.Nil(t, bid)
		assert.Contains(t, err.Error(), "auction ID cannot be empty")
	})

	t.Run("nil_bid_price", func(t *testing.T) {
		bid, err := NewBid(
			"bid1",
			"auction123",
			"bidder456",
			big.NewFloat(500),
			nil,
		)

		assert.Error(t, err)
		assert.Nil(t, bid)
		assert.Contains(t, err.Error(), "bid price must be positive")
	})
}

func TestNewRecoveryMechanism(t *testing.T) {
	t.Run("valid_recovery_mechanism", func(t *testing.T) {
		recoveryType := PartialRecovery
		description := "Partial recovery mechanism for liquidation"
		parameters := map[string]*big.Float{
			"recoveryAmount": big.NewFloat(50000),
			"threshold":      big.NewFloat(0.8),
		}

		recovery, err := NewRecoveryMechanism(
			"recovery1",
			recoveryType,
			description,
			parameters,
		)

		assert.NoError(t, err)
		assert.NotNil(t, recovery)
		assert.Equal(t, "recovery1", recovery.ID)
		assert.Equal(t, recoveryType, recovery.Type)
		assert.Equal(t, description, recovery.Description)
		assert.Equal(t, parameters, recovery.Parameters)
		assert.True(t, recovery.IsActive)
		assert.False(t, recovery.CreatedAt.IsZero())
		assert.False(t, recovery.UpdatedAt.IsZero())
	})

	t.Run("empty_id", func(t *testing.T) {
		parameters := map[string]*big.Float{
			"recoveryAmount": big.NewFloat(50000),
		}

		recovery, err := NewRecoveryMechanism(
			"",
			PartialRecovery,
			"description",
			parameters,
		)

		assert.Error(t, err)
		assert.Nil(t, recovery)
		assert.Contains(t, err.Error(), "mechanism ID cannot be empty")
	})

	t.Run("nil_parameters", func(t *testing.T) {
		recovery, err := NewRecoveryMechanism(
			"recovery1",
			PartialRecovery,
			"description",
			nil,
		)

		assert.Error(t, err)
		assert.Nil(t, recovery)
		assert.Contains(t, err.Error(), "parameters cannot be nil")
	})
}

// Helper function to create a test liquidation engine
func createTestLiquidationEngine(t *testing.T) *LiquidationEngine {
	riskFreeRate := big.NewFloat(0.02) // 2%
	confidenceLevel := big.NewFloat(0.95) // 95%
	basePremiumRate := big.NewFloat(0.01) // 1%
	
	riskManager, err := NewAdvancedRiskManager(riskFreeRate, confidenceLevel)
	require.NoError(t, err)
	
	insuranceManager, err := NewInsuranceManager(riskFreeRate, basePremiumRate)
	require.NoError(t, err)

	config, err := NewLiquidationConfig(
		24*time.Hour,
		big.NewFloat(0.05),
		1*time.Hour,
		7*24*time.Hour,
		big.NewFloat(0.01),
		big.NewFloat(0.1),
	)
	require.NoError(t, err)

	engine, err := NewLiquidationEngine(riskManager, insuranceManager, config)
	require.NoError(t, err)

	return engine
}

func TestLiquidationEngine_AddLiquidationTrigger(t *testing.T) {
	engine := createTestLiquidationEngine(t)

	trigger, err := NewLiquidationTrigger(
		"trigger1",
		MarginCall,
		big.NewFloat(1.5),
		24*time.Hour,
	)
	require.NoError(t, err)

	t.Run("add_valid_trigger", func(t *testing.T) {
		err := engine.AddLiquidationTrigger(trigger)
		assert.NoError(t, err)

		// Verify trigger was added
		retrievedTrigger, err := engine.GetLiquidationTrigger("trigger1")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedTrigger)
		assert.Equal(t, trigger.ID, retrievedTrigger.ID)
	})

	t.Run("add_nil_trigger", func(t *testing.T) {
		err := engine.AddLiquidationTrigger(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trigger cannot be nil")
	})
}

func TestLiquidationEngine_GetLiquidationTrigger(t *testing.T) {
	engine := createTestLiquidationEngine(t)

	trigger, err := NewLiquidationTrigger(
		"trigger1",
		MarginCall,
		big.NewFloat(1.5),
		24*time.Hour,
	)
	require.NoError(t, err)

	err = engine.AddLiquidationTrigger(trigger)
	require.NoError(t, err)

	t.Run("get_existing_trigger", func(t *testing.T) {
		retrievedTrigger, err := engine.GetLiquidationTrigger("trigger1")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedTrigger)
		assert.Equal(t, trigger.ID, retrievedTrigger.ID)
	})

	t.Run("get_non_existing_trigger", func(t *testing.T) {
		retrievedTrigger, err := engine.GetLiquidationTrigger("non_existent")
		assert.Error(t, err)
		assert.Nil(t, retrievedTrigger)
	})
}
