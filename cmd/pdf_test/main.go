package main

import (
	"log"
	"strings"
	"time"

	"github.com/palaseus/adrenochain/pkg/pdf"
)

func main() {
	log.Println("🚀 Starting Enhanced PDF Test Suite...")
	
	// Test 1: Basic multi-node PDF functionality
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🧪 TEST 1: Basic Multi-Node PDF Test")
	log.Println(strings.Repeat("=", 60))
	
	if err := pdf.RunMultiNodeTest(); err != nil {
		log.Printf("❌ Basic multi-node PDF test failed: %v", err)
	} else {
		log.Println("✅ Basic multi-node PDF test completed successfully!")
	}
	
	// Wait between tests
	time.Sleep(2 * time.Second)
	
	// Test 2: Enhanced multi-node PDF test with network simulation and consensus
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🚀 TEST 2: Enhanced Multi-Node PDF Test (Network Simulation + Blockchain Consensus)")
	log.Println(strings.Repeat("=", 60))
	
	if err := pdf.RunEnhancedMultiNodeTest(); err != nil {
		log.Printf("❌ Enhanced multi-node PDF test failed: %v", err)
		return
	}
	
	log.Println("✅ Enhanced multi-node PDF test completed successfully!")
	
	log.Println("\n🎉 All enhanced tests completed successfully!")
}
