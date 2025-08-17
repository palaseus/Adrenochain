package testing

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/palaseus/adrenochain/pkg/exchange/orderbook"
	"github.com/palaseus/adrenochain/pkg/exchange/trading"
	"github.com/palaseus/adrenochain/pkg/bridge"
	"github.com/palaseus/adrenochain/pkg/governance"
)

// TestDataGenerators provides utilities for generating test data
type TestDataGenerators struct {
	orderCounter    int64
	proposalCounter int64
	txCounter       int64
}

// NewTestDataGenerators creates new test data generators
func NewTestDataGenerators() *TestDataGenerators {
	return &TestDataGenerators{
		orderCounter:    0,
		proposalCounter: 0,
		txCounter:       0,
	}
}

// GenerateTradingPair generates a random trading pair
func (tdg *TestDataGenerators) GenerateTradingPair() *trading.TradingPair {
	baseAssets := []string{"BTC", "ETH", "ADA", "DOT", "LINK", "UNI", "AAVE", "COMP"}
	quoteAssets := []string{"USDT", "USDC", "DAI", "BUSD", "TUSD", "FRAX", "GUSD", "HUSD"}
	
	baseAsset := baseAssets[tdg.randomInt(len(baseAssets))]
	quoteAsset := quoteAssets[tdg.randomInt(len(quoteAssets))]
	
	// Generate random parameters
	minQuantity := big.NewInt(int64(1000 + tdg.randomInt(9000)))
	maxQuantity := big.NewInt(int64(1000000 + tdg.randomInt(9000000)))
	minPrice := big.NewInt(int64(100 + tdg.randomInt(900)))
	maxPrice := big.NewInt(int64(100000 + tdg.randomInt(900000)))
	tickSize := big.NewInt(int64(1 + tdg.randomInt(10)))
	stepSize := big.NewInt(int64(1 + tdg.randomInt(10)))
	makerFee := big.NewInt(int64(50 + tdg.randomInt(150)))
	takerFee := big.NewInt(int64(100 + tdg.randomInt(200)))
	
	pair, err := trading.NewTradingPair(
		baseAsset, quoteAsset,
		minQuantity, maxQuantity,
		minPrice, maxPrice,
		tickSize, stepSize,
		makerFee, takerFee,
	)
	
	if err != nil {
		// Fallback to default values if generation fails
		pair, _ = trading.NewTradingPair(
			"BTC", "USDT",
			big.NewInt(1000), big.NewInt(1000000),
			big.NewInt(100), big.NewInt(100000),
			big.NewInt(1), big.NewInt(1),
			big.NewInt(100), big.NewInt(200),
		)
	}
	
	return pair
}

// GenerateOrder generates a random order
func (tdg *TestDataGenerators) GenerateOrder(tradingPair string) *orderbook.Order {
	tdg.orderCounter++
	
	sides := []string{"buy", "sell"}
	types := []string{"limit", "market", "stop_loss", "take_profit"}
	
	side := sides[tdg.randomInt(len(sides))]
	orderType := types[tdg.randomInt(len(types))]
	
	// Generate random quantities and prices
	quantity := big.NewInt(int64(1000 + tdg.randomInt(9000)))
	price := big.NewInt(int64(10000 + tdg.randomInt(90000)))
	
	// Market orders don't have prices
	if orderType == "market" {
		price = nil
	}
	
	order := &orderbook.Order{
		ID:          fmt.Sprintf("order_%d", tdg.orderCounter),
		TradingPair: tradingPair,
		Side:        orderbook.OrderSide(side),
		Type:        orderbook.OrderType(orderType),
		Quantity:    quantity,
		Price:       price,
		UserID:      fmt.Sprintf("user_%d", tdg.randomInt(100)),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      orderbook.OrderStatusPending,
	}
	
	return order
}

// GenerateOrderBook generates a populated order book
func (tdg *TestDataGenerators) GenerateOrderBook(tradingPair string, orderCount int) (*orderbook.OrderBook, error) {
	ob, err := orderbook.NewOrderBook(tradingPair)
	if err != nil {
		return nil, err
	}
	
	// Generate and add orders
	for i := 0; i < orderCount; i++ {
		order := tdg.GenerateOrder(tradingPair)
		if err := ob.AddOrder(order); err != nil {
			fmt.Printf("Warning: Failed to add order %s: %v\n", order.ID, err)
		}
	}
	
	return ob, nil
}

// GenerateBridgeTransaction generates a random bridge transaction
func (tdg *TestDataGenerators) GenerateBridgeTransaction() *bridge.CrossChainTransaction {
	tdg.txCounter++
	
	chains := []bridge.ChainID{bridge.ChainIDGoChain, bridge.ChainIDEthereum, bridge.ChainIDPolygon}
	assetTypes := []bridge.AssetType{bridge.AssetTypeNative, bridge.AssetTypeERC20, bridge.AssetTypeERC721}
	
	sourceChain := chains[tdg.randomInt(len(chains))]
	destChain := chains[tdg.randomInt(len(chains))]
	// Ensure different chains
	for destChain == sourceChain {
		destChain = chains[tdg.randomInt(len(chains))]
	}
	
	assetType := assetTypes[tdg.randomInt(len(assetTypes))]
	amount := big.NewInt(int64(100000000000000000 + tdg.randomInt(900000000000000000))) // 0.1 to 1.0 ETH
	
	tx := &bridge.CrossChainTransaction{
		ID:                fmt.Sprintf("bridge_tx_%d", tdg.txCounter),
		SourceChain:       sourceChain,
		DestinationChain:  destChain,
		SourceAddress:     tdg.generateRandomAddress(),
		DestinationAddress: tdg.generateRandomAddress(),
		AssetType:         assetType,
		Amount:            amount,
		Status:            bridge.TransactionStatusPending,
		ValidatorID:       fmt.Sprintf("validator_%d", tdg.randomInt(10)),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Fee:               big.NewInt(int64(1000000000000000 + tdg.randomInt(9000000000000000))), // 0.001 to 0.01 ETH
	}
	
	return tx
}

// GenerateProposal generates a random governance proposal
func (tdg *TestDataGenerators) GenerateProposal() *governance.Proposal {
	tdg.proposalCounter++
	
	proposalTypes := []governance.ProposalType{
		governance.ProposalTypeGeneral,
		governance.ProposalTypeParameterChange,
		governance.ProposalTypeTreasury,
		governance.ProposalTypeUpgrade,
		governance.ProposalTypeEmergency,
	}
	
	proposalType := proposalTypes[tdg.randomInt(len(proposalTypes))]
	quorumRequired := big.NewInt(int64(1000000000000000000 + tdg.randomInt(9000000000000000000))) // 1 to 10 ETH
	minVotingPower := big.NewInt(int64(100000000000000000 + tdg.randomInt(900000000000000000))) // 0.1 to 1.0 ETH
	
	proposal := &governance.Proposal{
		ID:             fmt.Sprintf("proposal_%d", tdg.proposalCounter),
		Title:          fmt.Sprintf("Test Proposal %d", tdg.proposalCounter),
		Description:    fmt.Sprintf("This is a test proposal number %d for testing purposes", tdg.proposalCounter),
		Proposer:       fmt.Sprintf("proposer_%d", tdg.randomInt(50)),
		ProposalType:   proposalType,
		Status:         governance.ProposalStatusDraft,
		VotingStart:    time.Now().Add(time.Duration(tdg.randomInt(3600)) * time.Second),
		VotingEnd:      time.Now().Add(time.Duration(3600+tdg.randomInt(86400)) * time.Second),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		QuorumRequired: quorumRequired,
		MinVotingPower: minVotingPower,
	}
	
	return proposal
}

// GenerateVote generates a random vote
func (tdg *TestDataGenerators) GenerateVote(proposalID string) *governance.Vote {
	voteChoices := []governance.VoteChoice{
		governance.VoteChoiceFor,
		governance.VoteChoiceAgainst,
		governance.VoteChoiceAbstain,
	}
	
	voteChoice := voteChoices[tdg.randomInt(len(voteChoices))]
	votingPower := big.NewInt(int64(100000000000000000 + tdg.randomInt(900000000000000000))) // 0.1 to 1.0 ETH
	
	vote := &governance.Vote{
		ProposalID: proposalID,
		Voter:      fmt.Sprintf("voter_%d", tdg.randomInt(100)),
		VoteChoice: voteChoice,
		VotingPower: votingPower,
		Reason:     fmt.Sprintf("Vote reason for proposal %s", proposalID),
		Timestamp:  time.Now(),
		IsDelegated: false,
	}
	
	return vote
}

// GenerateTreasuryTransaction generates a random treasury transaction
func (tdg *TestDataGenerators) GenerateTreasuryTransaction() *governance.TreasuryTransaction {
	transactionTypes := []governance.TreasuryTransactionType{
		governance.TreasuryTransactionTypeTransfer,
		governance.TreasuryTransactionTypeWithdrawal,
		governance.TreasuryTransactionTypeDeposit,
		governance.TreasuryTransactionTypeInvestment,
		governance.TreasuryTransactionTypeReward,
	}
	
	transactionType := transactionTypes[tdg.randomInt(len(transactionTypes))]
	amount := big.NewInt(int64(1000000000000000000 + tdg.randomInt(9000000000000000000))) // 1 to 10 ETH
	
	tx := &governance.TreasuryTransaction{
		ID:          fmt.Sprintf("treasury_tx_%d", tdg.randomInt(1000)),
		Type:        transactionType,
		Amount:      amount,
		Asset:       "ETH",
		From:        "treasury",
		To:          tdg.generateRandomAddress(),
		Description: fmt.Sprintf("Test treasury transaction of type %s", transactionType),
		Status:      governance.TransactionStatusPending,
		ExecutedBy:  fmt.Sprintf("executor_%d", tdg.randomInt(10)),
		CreatedAt:   time.Now(),
	}
	
	return tx
}

// GenerateValidator generates a random validator
func (tdg *TestDataGenerators) GenerateValidator() *bridge.Validator {
	stakeAmount := big.NewInt(int64(1000000000000000000 + tdg.randomInt(9000000000000000000))) // 1 to 10 ETH
	
	validator := &bridge.Validator{
		ID:             fmt.Sprintf("validator_%d", tdg.randomInt(100)),
		Address:        tdg.generateRandomAddress(),
		ChainID:        bridge.ChainIDGoChain,
		StakeAmount:    stakeAmount,
		IsActive:       true,
		LastHeartbeat:  time.Now().Add(-time.Duration(tdg.randomInt(3600)) * time.Second),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TotalValidated: int64(tdg.randomInt(1000)),
		SuccessRate:    float64(90 + tdg.randomInt(10)), // 90-100%
	}
	
	return validator
}

// GenerateAssetMapping generates a random asset mapping
func (tdg *TestDataGenerators) GenerateAssetMapping() *bridge.AssetMapping {
	chains := []bridge.ChainID{bridge.ChainIDGoChain, bridge.ChainIDEthereum, bridge.ChainIDPolygon}
	assetTypes := []bridge.AssetType{bridge.AssetTypeNative, bridge.AssetTypeERC20, bridge.AssetTypeERC721, bridge.AssetTypeERC1155}
	
	sourceChain := chains[tdg.randomInt(len(chains))]
	destChain := chains[tdg.randomInt(len(chains))]
	// Ensure different chains
	for destChain == sourceChain {
		destChain = chains[tdg.randomInt(len(chains))]
	}
	
	assetType := assetTypes[tdg.randomInt(len(assetTypes))]
	minAmount := big.NewInt(int64(1000000000000000 + tdg.randomInt(9000000000000000))) // 0.001 to 0.01 ETH
	maxAmount := big.NewInt(int64(1000000000000000000 + tdg.randomInt(9000000000000000000))) // 1 to 10 ETH
	dailyLimit := big.NewInt(int64(1000000000000000000 + tdg.randomInt(9000000000000000000))) // 1 to 10 ETH
	
	mapping := &bridge.AssetMapping{
		ID:               fmt.Sprintf("mapping_%d", tdg.randomInt(1000)),
		SourceChain:      sourceChain,
		DestinationChain: destChain,
		SourceAsset:      tdg.generateRandomAddress(),
		DestinationAsset: tdg.generateRandomAddress(),
		AssetType:        assetType,
		Decimals:         uint8(18),
		IsActive:         true,
		MinAmount:        minAmount,
		MaxAmount:        maxAmount,
		DailyLimit:       dailyLimit,
		DailyUsed:        big.NewInt(0),
		FeePercentage:    float64(1 + tdg.randomInt(5)) / 100.0, // 0.01% to 0.06%
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	return mapping
}

// GenerateTestDataset generates a comprehensive test dataset
func (tdg *TestDataGenerators) GenerateTestDataset() *TestDataset {
	fmt.Println("📊 Generating comprehensive test dataset...")
	
	dataset := &TestDataset{
		TradingPairs:    make([]*trading.TradingPair, 0),
		Orders:          make([]*orderbook.Order, 0),
		OrderBooks:      make([]*orderbook.OrderBook, 0),
		BridgeTransactions: make([]*bridge.CrossChainTransaction, 0),
		Proposals:       make([]*governance.Proposal, 0),
		Votes:           make([]*governance.Vote, 0),
		TreasuryTransactions: make([]*governance.TreasuryTransaction, 0),
		Validators:      make([]*bridge.Validator, 0),
		AssetMappings:   make([]*bridge.AssetMapping, 0),
	}
	
	// Generate trading pairs
	for i := 0; i < 10; i++ {
		pair := tdg.GenerateTradingPair()
		dataset.TradingPairs = append(dataset.TradingPairs, pair)
	}
	
	// Generate orders and order books
	for i := 0; i < 5; i++ {
		pair := dataset.TradingPairs[i]
		tradingPair := fmt.Sprintf("%s/%s", pair.BaseAsset, pair.QuoteAsset)
		
		// Generate order book with orders
		ob, err := tdg.GenerateOrderBook(tradingPair, 50)
		if err == nil {
			dataset.OrderBooks = append(dataset.OrderBooks, ob)
			
			// Get orders from order book
			// This would require access to internal order book methods
			// For now, we'll generate orders separately
			for j := 0; j < 50; j++ {
				order := tdg.GenerateOrder(tradingPair)
				dataset.Orders = append(dataset.Orders, order)
			}
		}
	}
	
	// Generate bridge transactions
	for i := 0; i < 20; i++ {
		tx := tdg.GenerateBridgeTransaction()
		dataset.BridgeTransactions = append(dataset.BridgeTransactions, tx)
	}
	
	// Generate governance proposals and votes
	for i := 0; i < 10; i++ {
		proposal := tdg.GenerateProposal()
		dataset.Proposals = append(dataset.Proposals, proposal)
		
		// Generate votes for this proposal
		for j := 0; j < 15; j++ {
			vote := tdg.GenerateVote(proposal.ID)
			dataset.Votes = append(dataset.Votes, vote)
		}
	}
	
	// Generate treasury transactions
	for i := 0; i < 15; i++ {
		tx := tdg.GenerateTreasuryTransaction()
		dataset.TreasuryTransactions = append(dataset.TreasuryTransactions, tx)
	}
	
	// Generate validators
	for i := 0; i < 8; i++ {
		validator := tdg.GenerateValidator()
		dataset.Validators = append(dataset.Validators, validator)
	}
	
	// Generate asset mappings
	for i := 0; i < 12; i++ {
		mapping := tdg.GenerateAssetMapping()
		dataset.AssetMappings = append(dataset.AssetMappings, mapping)
	}
	
	fmt.Printf("✅ Test dataset generated: %d pairs, %d orders, %d proposals, %d validators\n",
		len(dataset.TradingPairs), len(dataset.Orders), len(dataset.Proposals), len(dataset.Validators))
	
	return dataset
}

// Helper methods

// randomInt generates a random integer in range [0, max)
func (tdg *TestDataGenerators) randomInt(max int) int {
	// Simple random number generation for testing
	// In production, you'd use crypto/rand
	return int(time.Now().UnixNano()) % max
}

// generateRandomAddress generates a random Ethereum-style address
func (tdg *TestDataGenerators) generateRandomAddress() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return fmt.Sprintf("0x%x", bytes)
}

// TestDataset represents a comprehensive test dataset
type TestDataset struct {
	TradingPairs        []*trading.TradingPair
	Orders              []*orderbook.Order
	OrderBooks          []*orderbook.OrderBook
	BridgeTransactions  []*bridge.CrossChainTransaction
	Proposals           []*governance.Proposal
	Votes               []*governance.Vote
	TreasuryTransactions []*governance.TreasuryTransaction
	Validators          []*bridge.Validator
	AssetMappings       []*bridge.AssetMapping
}

// GetSummary returns a summary of the test dataset
func (td *TestDataset) GetSummary() string {
	return fmt.Sprintf(
		"Test Dataset Summary:\n"+
			"  Trading Pairs: %d\n"+
			"  Orders: %d\n"+
			"  Order Books: %d\n"+
			"  Bridge Transactions: %d\n"+
			"  Proposals: %d\n"+
			"  Votes: %d\n"+
			"  Treasury Transactions: %d\n"+
			"  Validators: %d\n"+
			"  Asset Mappings: %d",
		len(td.TradingPairs), len(td.Orders), len(td.OrderBooks),
		len(td.BridgeTransactions), len(td.Proposals), len(td.Votes),
		len(td.TreasuryTransactions), len(td.Validators), len(td.AssetMappings),
	)
}
