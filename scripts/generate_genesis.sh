#!/bin/bash

# 🏗️ Adrenochain Genesis Block Generation Script
# This script generates the genesis block for the adrenochain network

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GENESIS_DIR="$PROJECT_ROOT/genesis"
GENESIS_FILE="$GENESIS_DIR/genesis.json"
CONFIG_FILE="$GENESIS_DIR/genesis_config.yaml"

echo -e "${BLUE}🏗️  Generating Adrenochain Genesis Block...${NC}"

# Create genesis directory
echo -e "${BLUE}📁 Creating genesis directory...${NC}"
mkdir -p "$GENESIS_DIR"

# Create genesis configuration
echo -e "${BLUE}⚙️  Creating genesis configuration...${NC}"
cat > "$CONFIG_FILE" << 'EOF'
# Genesis Block Configuration
genesis:
  chain_id: "adrenochain-mainnet-1"
  timestamp: "2024-01-01T00:00:00Z"
  block_time: 10s
  difficulty: 1000000
  genesis_reward: 1000000000

# Initial validators
validators:
  - name: "genesis-validator-1"
    address: "0x1234567890123456789012345678901234567890"
    stake: 1000000000
    commission_rate: 0.1
  - name: "genesis-validator-2"
    address: "0x0987654321098765432109876543210987654321"
    stake: 1000000000
    commission_rate: 0.1
  - name: "genesis-validator-3"
    address: "0x1111111111111111111111111111111111111111"
    stake: 1000000000
    commission_rate: 0.1

# Initial token distribution
token_distribution:
  total_supply: 10000000000  # 10 billion tokens
  genesis_validators: 3000000000  # 3 billion to validators
  community_treasury: 2000000000  # 2 billion to community treasury
  development_fund: 2000000000    # 2 billion to development fund
  ecosystem_fund: 2000000000      # 2 billion to ecosystem fund
  reserve_fund: 1000000000        # 1 billion to reserve fund

# Network parameters
network_params:
  max_validators: 100
  min_stake: 1000000
  unbonding_period: 2592000  # 30 days in seconds
  max_commission_rate: 0.2
  min_commission_rate: 0.0
EOF

# Create genesis block generation tool
echo -e "${BLUE}🔧 Creating genesis block generation tool...${NC}"
cat > "$GENESIS_DIR/generate_genesis.go" << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/consensus"
)

type GenesisConfig struct {
	Genesis struct {
		ChainID        string `yaml:"chain_id"`
		Timestamp      string `yaml:"timestamp"`
		BlockTime      string `yaml:"block_time"`
		Difficulty     uint64 `yaml:"difficulty"`
		GenesisReward  uint64 `yaml:"genesis_reward"`
	} `yaml:"genesis"`
	
	Validators []struct {
		Name           string `yaml:"name"`
		Address        string `yaml:"address"`
		Stake          uint64 `yaml:"stake"`
		CommissionRate float64 `yaml:"commission_rate"`
	} `yaml:"validators"`
	
	TokenDistribution struct {
		TotalSupply        uint64 `yaml:"total_supply"`
		GenesisValidators  uint64 `yaml:"genesis_validators"`
		CommunityTreasury  uint64 `yaml:"community_treasury"`
		DevelopmentFund    uint64 `yaml:"development_fund"`
		EcosystemFund      uint64 `yaml:"ecosystem_fund"`
		ReserveFund        uint64 `yaml:"reserve_fund"`
	} `yaml:"token_distribution"`
}

type GenesisBlock struct {
	ChainID       string                 `json:"chain_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Validators    []Validator            `json:"validators"`
	Distribution  TokenDistribution      `json:"token_distribution"`
	NetworkParams NetworkParams          `json:"network_params"`
	Block         *block.Block           `json:"block"`
}

type Validator struct {
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Stake          uint64  `json:"stake"`
	CommissionRate float64 `json:"commission_rate"`
}

type TokenDistribution struct {
	TotalSupply        uint64 `json:"total_supply"`
	GenesisValidators  uint64 `json:"genesis_validators"`
	CommunityTreasury  uint64 `json:"community_treasury"`
	DevelopmentFund    uint64 `json:"development_fund"`
	EcosystemFund      uint64 `json:"ecosystem_fund"`
	ReserveFund        uint64 `json:"reserve_fund"`
}

type NetworkParams struct {
	MaxValidators      int    `json:"max_validators"`
	MinStake           uint64 `json:"min_stake"`
	UnbondingPeriod    uint64 `json:"unbonding_period"`
	MaxCommissionRate  float64 `json:"max_commission_rate"`
	MinCommissionRate  float64 `json:"min_commission_rate"`
}

func main() {
	fmt.Println("🏗️  Generating Adrenochain Genesis Block...")
	
	// Create genesis block
	genesisBlock := &GenesisBlock{
		ChainID:   "adrenochain-mainnet-1",
		Timestamp: time.Now(),
		Validators: []Validator{
			{
				Name:           "genesis-validator-1",
				Address:        "0x1234567890123456789012345678901234567890",
				Stake:          1000000000,
				CommissionRate: 0.1,
			},
			{
				Name:           "genesis-validator-2",
				Address:        "0x0987654321098765432109876543210987654321",
				Stake:          1000000000,
				CommissionRate: 0.1,
			},
			{
				Name:           "genesis-validator-3",
				Address:        "0x1111111111111111111111111111111111111111",
				Stake:          1000000000,
				CommissionRate: 0.1,
			},
		},
		Distribution: TokenDistribution{
			TotalSupply:       10000000000,
			GenesisValidators: 3000000000,
			CommunityTreasury: 2000000000,
			DevelopmentFund:   2000000000,
			EcosystemFund:     2000000000,
			ReserveFund:       1000000000,
		},
		NetworkParams: NetworkParams{
			MaxValidators:     100,
			MinStake:          1000000,
			UnbondingPeriod:   2592000,
			MaxCommissionRate: 0.2,
			MinCommissionRate: 0.0,
		},
	}
	
	// Create the actual blockchain genesis block
	genesisBlock.Block = createGenesisBlock(genesisBlock)
	
	// Write genesis block to file
	genesisJSON, err := json.MarshalIndent(genesisBlock, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling genesis block: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile("genesis.json", genesisJSON, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing genesis block: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Genesis block generated successfully!")
	fmt.Println("📁 Genesis block saved to: genesis.json")
	fmt.Printf("🔗 Chain ID: %s\n", genesisBlock.ChainID)
	fmt.Printf("⏰ Timestamp: %s\n", genesisBlock.Timestamp.Format(time.RFC3339))
	fmt.Printf("👥 Validators: %d\n", len(genesisBlock.Validators))
	fmt.Printf("💰 Total Supply: %d tokens\n", genesisBlock.Distribution.TotalSupply)
}

func createGenesisBlock(genesis *GenesisBlock) *block.Block {
	// Create genesis transaction
	genesisTx := &block.Transaction{
		Version: 1,
		Inputs:  []*block.TxInput{}, // No inputs for genesis
		Outputs: []*block.TxOutput{
			{
				Value:        genesis.Distribution.TotalSupply,
				ScriptPubKey: []byte("genesis_output"),
			},
		},
		LockTime: 0,
		Fee:      0,
		Hash:     []byte("genesis_transaction_hash"),
	}
	
	// Create genesis block header
	header := &block.Header{
		Version:       1,
		PrevBlockHash: make([]byte, 32), // Zero hash for genesis
		MerkleRoot:    genesisTx.CalculateHash(),
		Timestamp:     genesis.Timestamp,
		Difficulty:    1000000,
		Nonce:         0,
		Height:        0,
	}
	
	// Create genesis block
	genesisBlock := &block.Block{
		Header:       header,
		Transactions: []*block.Transaction{genesisTx},
		MerkleRoot:   genesisTx.CalculateHash(),
	}
	
	return genesisBlock
}
EOF

# Create genesis generation script
echo -e "${BLUE}🚀 Creating genesis generation script...${NC}"
cat > "$GENESIS_DIR/generate.sh" << 'EOF'
#!/bin/bash

# Generate genesis block
echo "🏗️  Generating genesis block..."

# Build and run genesis generator
cd "$(dirname "$0")"
go mod init genesis-generator 2>/dev/null || true
go mod tidy
go run generate_genesis.go

echo "✅ Genesis block generation completed!"
echo "📁 Files created:"
echo "   - genesis.json (genesis block)"
echo "   - genesis_config.yaml (configuration)"
EOF

chmod +x "$GENESIS_DIR/generate.sh"

# Create genesis validation script
echo -e "${BLUE}🔍 Creating genesis validation script...${NC}"
cat > "$GENESIS_DIR/validate_genesis.sh" << 'EOF'
#!/bin/bash

# Validate genesis block
echo "🔍 Validating genesis block..."

if [ ! -f "genesis.json" ]; then
    echo "❌ Genesis block not found. Run generate.sh first."
    exit 1
fi

# Validate JSON structure
if ! jq empty genesis.json 2>/dev/null; then
    echo "❌ Invalid JSON in genesis block"
    exit 1
fi

# Check required fields
required_fields=("chain_id" "timestamp" "validators" "token_distribution" "network_params" "block")
for field in "${required_fields[@]}"; do
    if ! jq -e ".$field" genesis.json > /dev/null; then
        echo "❌ Missing required field: $field"
        exit 1
    fi
done

# Validate token distribution
total_supply=$(jq -r '.token_distribution.total_supply' genesis.json)
distributed=$(jq -r '.token_distribution | .genesis_validators + .community_treasury + .development_fund + .ecosystem_fund + .reserve_fund' genesis.json)

if [ "$total_supply" != "$distributed" ]; then
    echo "❌ Token distribution mismatch: total=$total_supply, distributed=$distributed"
    exit 1
fi

# Validate validators
validator_count=$(jq '.validators | length' genesis.json)
if [ "$validator_count" -lt 1 ]; then
    echo "❌ No validators found in genesis block"
    exit 1
fi

echo "✅ Genesis block validation passed!"
echo "📊 Genesis block summary:"
echo "   Chain ID: $(jq -r '.chain_id' genesis.json)"
echo "   Validators: $validator_count"
echo "   Total Supply: $(jq -r '.token_distribution.total_supply' genesis.json)"
echo "   Timestamp: $(jq -r '.timestamp' genesis.json)"
EOF

chmod +x "$GENESIS_DIR/validate_genesis.sh"

# Create README for genesis
echo -e "${BLUE}📝 Creating genesis documentation...${NC}"
cat > "$GENESIS_DIR/README.md" << 'EOF'
# Adrenochain Genesis Block

This directory contains the genesis block generation and validation tools for adrenochain.

## Quick Start

1. **Generate Genesis Block**
   ```bash
   ./generate.sh
   ```

2. **Validate Genesis Block**
   ```bash
   ./validate_genesis.sh
   ```

## Files

- `genesis.json` - The generated genesis block
- `genesis_config.yaml` - Configuration for genesis block generation
- `generate_genesis.go` - Go program to generate the genesis block
- `generate.sh` - Script to build and run the genesis generator
- `validate_genesis.sh` - Script to validate the generated genesis block

## Genesis Block Structure

The genesis block contains:

- **Chain ID**: Unique identifier for the adrenochain network
- **Timestamp**: When the genesis block was created
- **Validators**: Initial set of validators with their stakes
- **Token Distribution**: How the initial token supply is distributed
- **Network Parameters**: Network-wide configuration parameters
- **Block**: The actual blockchain genesis block

## Token Distribution

- **Total Supply**: 10 billion tokens
- **Genesis Validators**: 3 billion tokens (30%)
- **Community Treasury**: 2 billion tokens (20%)
- **Development Fund**: 2 billion tokens (20%)
- **Ecosystem Fund**: 2 billion tokens (20%)
- **Reserve Fund**: 1 billion tokens (10%)

## Network Parameters

- **Max Validators**: 100
- **Min Stake**: 1,000,000 tokens
- **Unbonding Period**: 30 days
- **Max Commission Rate**: 20%
- **Min Commission Rate**: 0%
EOF

echo -e "${GREEN}✅ Genesis block generation setup completed!${NC}"
echo -e "${BLUE}📁 Genesis directory: $GENESIS_DIR${NC}"
echo -e "${BLUE}📝 Next steps:${NC}"
echo -e "   1. cd $GENESIS_DIR"
echo -e "   2. ./generate.sh"
echo -e "   3. ./validate_genesis.sh"
