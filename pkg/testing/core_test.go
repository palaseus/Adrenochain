package testing

import (
	"testing"
)

func TestPerformanceBenchmarks(t *testing.T) {
	benchmarks := NewPerformanceBenchmarks()
	results := benchmarks.RunAllBenchmarks()
	
	if len(results) == 0 {
		t.Error("Expected benchmark results, got none")
	}
	
	// Verify we have results for key components
	expectedComponents := []string{
		"orderbook_add_orders",
		"matching_engine",
		"trading_pairs",
		"bridge_validators",
		"validator_consensus",
		"voting_system",
		"treasury_operations",
		"lending_protocols",
		"amm_operations",
	}
	
	for _, component := range expectedComponents {
		if _, exists := results[component]; !exists {
			t.Errorf("Expected benchmark result for %s", component)
		}
	}
}

func TestSecurityTestFramework(t *testing.T) {
	securityTests := NewSecurityTestFramework()
	results := securityTests.RunAllSecurityTests()
	
	if len(results) == 0 {
		t.Error("Expected security test results, got none")
	}
	
	// Verify we have security tests for key components
	expectedComponents := []string{
		"OrderBook_order_id",
		"MatchingEngine_order_manipulation",
		"Bridge_validator_collusion",
		"VotingSystem_vote_manipulation",
		"Treasury_fund_theft",
	}
	
	for _, component := range expectedComponents {
		if _, exists := results[component]; !exists {
			t.Errorf("Expected security test result for %s", component)
		}
	}
}

func TestTestDataGenerators(t *testing.T) {
	generators := NewTestDataGenerators()
	
	// Test trading pair generation
	pair := generators.GenerateTradingPair()
	if pair == nil {
		t.Error("Failed to generate trading pair")
	}
	
	// Test order generation
	order := generators.GenerateOrder("BTC/USDT")
	if order == nil {
		t.Error("Failed to generate order")
	}
	
	// Test proposal generation
	proposal := generators.GenerateProposal()
	if proposal == nil {
		t.Error("Failed to generate proposal")
	}
	
	// Test vote generation
	vote := generators.GenerateVote(proposal.ID)
	if vote == nil {
		t.Error("Failed to generate vote")
	}
	
	// Test dataset generation
	dataset := generators.GenerateTestDataset()
	if dataset == nil {
		t.Error("Failed to generate test dataset")
	}
	
	// Verify dataset has content
	if len(dataset.TradingPairs) == 0 {
		t.Error("Expected trading pairs in dataset")
	}
	
	if len(dataset.Proposals) == 0 {
		t.Error("Expected proposals in dataset")
	}
	
	if len(dataset.Validators) == 0 {
		t.Error("Expected validators in dataset")
	}
}
