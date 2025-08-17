package risk

import (
	"math/big"
	"testing"
	"time"
)

func TestNewLiquidationConfig(t *testing.T) {
	tests := []struct {
		name                string
		gracePeriod         time.Duration
		liquidationFee      *big.Float
		minAuctionDuration  time.Duration
		maxAuctionDuration  time.Duration
		bidIncrement        *big.Float
		autoExtendThreshold *big.Float
		expectError         bool
	}{
		{
			name:                "Valid Config",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(0.05),
			minAuctionDuration:  1 * time.Hour,
			maxAuctionDuration:  24 * time.Hour,
			bidIncrement:        big.NewFloat(0.01),
			autoExtendThreshold: big.NewFloat(5),
			expectError:         false,
		},
		{
			name:                "Negative Liquidation Fee",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(-0.01),
			minAuctionDuration:  1 * time.Hour,
			maxAuctionDuration:  24 * time.Hour,
			bidIncrement:        big.NewFloat(0.01),
			autoExtendThreshold: big.NewFloat(5),
			expectError:         true,
		},
		{
			name:                "Zero Min Auction Duration",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(0.05),
			minAuctionDuration:  0,
			maxAuctionDuration:  24 * time.Hour,
			bidIncrement:        big.NewFloat(0.01),
			autoExtendThreshold: big.NewFloat(5),
			expectError:         true,
		},
		{
			name:                "Max Duration <= Min Duration",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(0.05),
			minAuctionDuration:  2 * time.Hour,
			maxAuctionDuration:  1 * time.Hour,
			bidIncrement:        big.NewFloat(0.01),
			autoExtendThreshold: big.NewFloat(5),
			expectError:         true,
		},
		{
			name:                "Zero Bid Increment",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(0.05),
			minAuctionDuration:  1 * time.Hour,
			maxAuctionDuration:  24 * time.Hour,
			bidIncrement:        big.NewFloat(0),
			autoExtendThreshold: big.NewFloat(5),
			expectError:         true,
		},
		{
			name:                "Negative Auto-Extend Threshold",
			gracePeriod:         5 * time.Minute,
			liquidationFee:      big.NewFloat(0.05),
			minAuctionDuration:  1 * time.Hour,
			maxAuctionDuration:  24 * time.Hour,
			bidIncrement:        big.NewFloat(0.01),
			autoExtendThreshold: big.NewFloat(-1),
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewLiquidationConfig(
				tt.gracePeriod,
				tt.liquidationFee,
				tt.minAuctionDuration,
				tt.maxAuctionDuration,
				tt.bidIncrement,
				tt.autoExtendThreshold,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Errorf("Expected config but got nil")
				return
			}

			if config.DefaultGracePeriod != tt.gracePeriod {
				t.Errorf("Expected grace period %v, got %v", tt.gracePeriod, config.DefaultGracePeriod)
			}

			if config.DefaultLiquidationFee.Cmp(tt.liquidationFee) != 0 {
				t.Errorf("Expected liquidation fee %v, got %v", tt.liquidationFee, config.DefaultLiquidationFee)
			}

			if config.MinimumAuctionDuration != tt.minAuctionDuration {
				t.Errorf("Expected min auction duration %v, got %v", tt.minAuctionDuration, config.MinimumAuctionDuration)
			}

			if config.MaximumAuctionDuration != tt.maxAuctionDuration {
				t.Errorf("Expected max auction duration %v, got %v", tt.maxAuctionDuration, config.MaximumAuctionDuration)
			}
		})
	}
}

func TestNewLiquidationTrigger(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		triggerType LiquidationTriggerType
		threshold   *big.Float
		gracePeriod time.Duration
		expectError bool
	}{
		{
			name:        "Valid Trigger",
			id:          "trigger1",
			triggerType: MarginCall,
			threshold:   big.NewFloat(0.8),
			gracePeriod: 5 * time.Minute,
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			triggerType: MarginCall,
			threshold:   big.NewFloat(0.8),
			gracePeriod: 5 * time.Minute,
			expectError: true,
		},
		{
			name:        "Zero Threshold",
			id:          "trigger1",
			triggerType: MarginCall,
			threshold:   big.NewFloat(0),
			gracePeriod: 5 * time.Minute,
			expectError: true,
		},
		{
			name:        "Negative Threshold",
			id:          "trigger1",
			triggerType: MarginCall,
			threshold:   big.NewFloat(-0.1),
			gracePeriod: 5 * time.Minute,
			expectError: true,
		},
		{
			name:        "Zero Grace Period",
			id:          "trigger1",
			triggerType: MarginCall,
			threshold:   big.NewFloat(0.8),
			gracePeriod: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, err := NewLiquidationTrigger(tt.id, tt.triggerType, tt.threshold, tt.gracePeriod)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if trigger == nil {
				t.Errorf("Expected trigger but got nil")
				return
			}

			if trigger.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, trigger.ID)
			}

			if trigger.Type != tt.triggerType {
				t.Errorf("Expected type %v, got %v", tt.triggerType, trigger.Type)
			}

			if trigger.Threshold.Cmp(tt.threshold) != 0 {
				t.Errorf("Expected threshold %v, got %v", tt.threshold, trigger.Threshold)
			}

			if trigger.GracePeriod != tt.gracePeriod {
				t.Errorf("Expected grace period %v, got %v", tt.gracePeriod, trigger.GracePeriod)
			}

			if !trigger.IsActive {
				t.Errorf("Expected trigger to be active")
			}
		})
	}
}

func TestNewLiquidationEvent(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		positionID      string
		userID          string
		triggerType     LiquidationTriggerType
		triggerValue    *big.Float
		threshold       *big.Float
		positionValue   *big.Float
		debtAmount      *big.Float
		collateralValue *big.Float
		expectError     bool
	}{
		{
			name:            "Valid Event",
			id:              "event1",
			positionID:      "position1",
			userID:          "user1",
			triggerType:     MarginCall,
			triggerValue:    big.NewFloat(0.75),
			threshold:       big.NewFloat(0.8),
			positionValue:   big.NewFloat(10000),
			debtAmount:      big.NewFloat(8000),
			collateralValue: big.NewFloat(12000),
			expectError:     false,
		},
		{
			name:            "Empty ID",
			id:              "",
			positionID:      "position1",
			userID:          "user1",
			triggerType:     MarginCall,
			triggerValue:    big.NewFloat(0.75),
			threshold:       big.NewFloat(0.8),
			positionValue:   big.NewFloat(10000),
			debtAmount:      big.NewFloat(8000),
			collateralValue: big.NewFloat(12000),
			expectError:     true,
		},
		{
			name:            "Empty Position ID",
			id:              "event1",
			positionID:      "",
			userID:          "user1",
			triggerType:     MarginCall,
			triggerValue:    big.NewFloat(0.75),
			threshold:       big.NewFloat(0.8),
			positionValue:   big.NewFloat(10000),
			debtAmount:      big.NewFloat(8000),
			collateralValue: big.NewFloat(12000),
			expectError:     true,
		},
		{
			name:            "Empty User ID",
			id:              "event1",
			positionID:      "position1",
			userID:          "",
			triggerType:     MarginCall,
			triggerValue:    big.NewFloat(0.75),
			threshold:       big.NewFloat(0.8),
			positionValue:   big.NewFloat(10000),
			debtAmount:      big.NewFloat(8000),
			collateralValue: big.NewFloat(12000),
			expectError:     true,
		},
		{
			name:            "Nil Trigger Value",
			id:              "event1",
			positionID:      "position1",
			userID:          "user1",
			triggerType:     MarginCall,
			triggerValue:    nil,
			threshold:       big.NewFloat(0.8),
			positionValue:   big.NewFloat(10000),
			debtAmount:      big.NewFloat(8000),
			collateralValue: big.NewFloat(12000),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewLiquidationEvent(
				tt.id,
				tt.positionID,
				tt.userID,
				tt.triggerType,
				tt.triggerValue,
				tt.threshold,
				tt.positionValue,
				tt.debtAmount,
				tt.collateralValue,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if event == nil {
				t.Errorf("Expected event but got nil")
				return
			}

			if event.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, event.ID)
			}

			if event.PositionID != tt.positionID {
				t.Errorf("Expected position ID %s, got %s", tt.positionID, event.PositionID)
			}

			if event.UserID != tt.userID {
				t.Errorf("Expected user ID %s, got %s", tt.userID, event.UserID)
			}

			if event.TriggerType != tt.triggerType {
				t.Errorf("Expected trigger type %v, got %v", tt.triggerType, event.TriggerType)
			}

			if event.Status != LiquidationTriggered {
				t.Errorf("Expected status LiquidationTriggered, got %v", event.Status)
			}

			if event.LiquidationFee.Sign() != 0 {
				t.Errorf("Expected zero liquidation fee, got %v", event.LiquidationFee)
			}
		})
	}
}

func TestNewAuction(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		liquidationID string
		assetID       string
		assetAmount   *big.Float
		startingPrice *big.Float
		minimumPrice  *big.Float
		reservePrice  *big.Float
		duration      time.Duration
		expectError   bool
	}{
		{
			name:          "Valid Auction",
			id:            "auction1",
			liquidationID: "liquidation1",
			assetID:       "asset1",
			assetAmount:   big.NewFloat(100),
			startingPrice: big.NewFloat(50),
			minimumPrice:  big.NewFloat(30),
			reservePrice:  big.NewFloat(40),
			duration:      2 * time.Hour,
			expectError:   false,
		},
		{
			name:          "Empty ID",
			id:            "",
			liquidationID: "liquidation1",
			assetID:       "asset1",
			assetAmount:   big.NewFloat(100),
			startingPrice: big.NewFloat(50),
			minimumPrice:  big.NewFloat(30),
			reservePrice:  big.NewFloat(40),
			duration:      2 * time.Hour,
			expectError:   true,
		},
		{
			name:          "Zero Asset Amount",
			id:            "auction1",
			liquidationID: "liquidation1",
			assetID:       "asset1",
			assetAmount:   big.NewFloat(0),
			startingPrice: big.NewFloat(50),
			minimumPrice:  big.NewFloat(30),
			reservePrice:  big.NewFloat(40),
			duration:      2 * time.Hour,
			expectError:   true,
		},
		{
			name:          "Zero Duration",
			id:            "auction1",
			liquidationID: "liquidation1",
			assetID:       "asset1",
			assetAmount:   big.NewFloat(100),
			startingPrice: big.NewFloat(50),
			minimumPrice:  big.NewFloat(30),
			reservePrice:  big.NewFloat(40),
			duration:      0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auction, err := NewAuction(
				tt.id,
				tt.liquidationID,
				tt.assetID,
				tt.assetAmount,
				tt.startingPrice,
				tt.minimumPrice,
				tt.reservePrice,
				tt.duration,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if auction == nil {
				t.Errorf("Expected auction but got nil")
				return
			}

			if auction.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, auction.ID)
			}

			if auction.LiquidationID != tt.liquidationID {
				t.Errorf("Expected liquidation ID %s, got %s", tt.liquidationID, auction.LiquidationID)
			}

			if auction.AssetID != tt.assetID {
				t.Errorf("Expected asset ID %s, got %s", tt.assetID, auction.AssetID)
			}

			if auction.Status != AuctionPending {
				t.Errorf("Expected status AuctionPending, got %v", auction.Status)
			}

			if len(auction.Bids) != 0 {
				t.Errorf("Expected empty bids list")
			}

			if auction.Winner != nil {
				t.Errorf("Expected nil winner")
			}
		})
	}
}

func TestNewBid(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		auctionID   string
		bidderID    string
		amount      *big.Float
		price       *big.Float
		expectError bool
	}{
		{
			name:        "Valid Bid",
			id:          "bid1",
			auctionID:   "auction1",
			bidderID:    "bidder1",
			amount:      big.NewFloat(50),
			price:       big.NewFloat(55),
			expectError: false,
		},
		{
			name:        "Empty ID",
			id:          "",
			auctionID:   "auction1",
			bidderID:    "bidder1",
			amount:      big.NewFloat(50),
			price:       big.NewFloat(55),
			expectError: true,
		},
		{
			name:        "Zero Amount",
			id:          "bid1",
			auctionID:   "auction1",
			bidderID:    "bidder1",
			amount:      big.NewFloat(0),
			price:       big.NewFloat(55),
			expectError: true,
		},
		{
			name:        "Zero Price",
			id:          "bid1",
			auctionID:   "auction1",
			bidderID:    "bidder1",
			amount:      big.NewFloat(50),
			price:       big.NewFloat(0),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid, err := NewBid(tt.id, tt.auctionID, tt.bidderID, tt.amount, tt.price)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if bid == nil {
				t.Errorf("Expected bid but got nil")
				return
			}

			if bid.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, bid.ID)
			}

			if bid.AuctionID != tt.auctionID {
				t.Errorf("Expected auction ID %s, got %s", tt.auctionID, bid.AuctionID)
			}

			if bid.BidderID != tt.bidderID {
				t.Errorf("Expected bidder ID %s, got %s", tt.bidderID, bid.BidderID)
			}

			if !bid.IsValid {
				t.Errorf("Expected bid to be valid")
			}
		})
	}
}

func TestLiquidationEngineBasicOperations(t *testing.T) {
	// Create liquidation config
	config, err := NewLiquidationConfig(
		5*time.Minute,
		big.NewFloat(0.05),
		1*time.Hour,
		24*time.Hour,
		big.NewFloat(0.01),
		big.NewFloat(5),
	)
	if err != nil {
		t.Fatalf("Failed to create liquidation config: %v", err)
	}

	// Create mock risk manager and insurance manager
	riskManager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}

	insuranceManager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Create liquidation engine
	engine, err := NewLiquidationEngine(riskManager, insuranceManager, config)
	if err != nil {
		t.Fatalf("Failed to create liquidation engine: %v", err)
	}

	t.Run("Add and Get Liquidation Trigger", func(t *testing.T) {
		trigger, err := NewLiquidationTrigger("trigger1", MarginCall, big.NewFloat(0.8), 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to create trigger: %v", err)
		}

		err = engine.AddLiquidationTrigger(trigger)
		if err != nil {
			t.Errorf("Failed to add trigger: %v", err)
			return
		}

		retrievedTrigger, err := engine.GetLiquidationTrigger("trigger1")
		if err != nil {
			t.Errorf("Failed to get trigger: %v", err)
			return
		}

		if retrievedTrigger != trigger {
			t.Errorf("Expected trigger %v, got %v", trigger, retrievedTrigger)
		}
	})

	t.Run("Add and Get Liquidation Event", func(t *testing.T) {
		event, err := NewLiquidationEvent(
			"event1",
			"position1",
			"user1",
			MarginCall,
			big.NewFloat(0.75),
			big.NewFloat(0.8),
			big.NewFloat(10000),
			big.NewFloat(8000),
			big.NewFloat(12000),
		)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		err = engine.AddLiquidationEvent(event)
		if err != nil {
			t.Errorf("Failed to add event: %v", err)
			return
		}

		retrievedEvent, err := engine.GetLiquidationEvent("event1")
		if err != nil {
			t.Errorf("Failed to get event: %v", err)
			return
		}

		if retrievedEvent != event {
			t.Errorf("Expected event %v, got %v", event, retrievedEvent)
		}
	})

	t.Run("Get Non-existent Items", func(t *testing.T) {
		_, err := engine.GetLiquidationTrigger("non_existent")
		if err == nil {
			t.Errorf("Expected error for non-existent trigger")
		}

		_, err = engine.GetLiquidationEvent("non_existent")
		if err == nil {
			t.Errorf("Expected error for non-existent event")
		}
	})
}

func TestLiquidationEngineStatistics(t *testing.T) {
	// Create liquidation config
	config, err := NewLiquidationConfig(
		5*time.Minute,
		big.NewFloat(0.05),
		1*time.Hour,
		24*time.Hour,
		big.NewFloat(0.01),
		big.NewFloat(5),
	)
	if err != nil {
		t.Fatalf("Failed to create liquidation config: %v", err)
	}

	// Create mock managers
	riskManager, err := NewAdvancedRiskManager(big.NewFloat(0.05), big.NewFloat(0.95))
	if err != nil {
		t.Fatalf("Failed to create risk manager: %v", err)
	}

	insuranceManager, err := NewInsuranceManager(big.NewFloat(0.05), big.NewFloat(0.02))
	if err != nil {
		t.Fatalf("Failed to create insurance manager: %v", err)
	}

	// Create liquidation engine
	engine, err := NewLiquidationEngine(riskManager, insuranceManager, config)
	if err != nil {
		t.Fatalf("Failed to create liquidation engine: %v", err)
	}

	t.Run("Empty Statistics", func(t *testing.T) {
		stats := engine.GetLiquidationStatistics()

		if stats["total_events"].(int) != 0 {
			t.Errorf("Expected 0 total events, got %d", stats["total_events"].(int))
		}

		if stats["total_auctions"].(int) != 0 {
			t.Errorf("Expected 0 total auctions, got %d", stats["total_auctions"].(int))
		}
	})

	t.Run("Statistics with Data", func(t *testing.T) {
		// Add some test events
		event1, _ := NewLiquidationEvent(
			"event1", "position1", "user1",
			MarginCall, big.NewFloat(0.75), big.NewFloat(0.8),
			big.NewFloat(10000), big.NewFloat(8000), big.NewFloat(12000),
		)
		event2, _ := NewLiquidationEvent(
			"event2", "position2", "user2",
			HealthFactor, big.NewFloat(0.6), big.NewFloat(0.7),
			big.NewFloat(15000), big.NewFloat(12000), big.NewFloat(18000),
		)

		engine.AddLiquidationEvent(event1)
		engine.AddLiquidationEvent(event2)

		stats := engine.GetLiquidationStatistics()

		if stats["total_events"].(int) != 2 {
			t.Errorf("Expected 2 total events, got %d", stats["total_events"].(int))
		}

		statusCounts := stats["events_by_status"].(map[LiquidationStatus]int)
		if statusCounts[LiquidationTriggered] != 2 {
			t.Errorf("Expected 2 triggered events, got %d", statusCounts[LiquidationTriggered])
		}
	})
}
