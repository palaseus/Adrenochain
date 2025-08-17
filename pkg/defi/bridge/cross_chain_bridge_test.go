package bridge

import (
	"math/big"
	"testing"
	"time"
)

func TestNewBridgeManager(t *testing.T) {
	bm := NewBridgeManager()
	
	if bm == nil {
		t.Fatal("NewBridgeManager returned nil")
	}
	
	if bm.Chains == nil {
		t.Error("Chains map not initialized")
	}
	if bm.Assets == nil {
		t.Error("Assets map not initialized")
	}
	if bm.BridgeAssets == nil {
		t.Error("BridgeAssets map not initialized")
	}
	if bm.Bridges == nil {
		t.Error("Bridges map not initialized")
	}
	if bm.Transfers == nil {
		t.Error("Transfers map not initialized")
	}
	if bm.Validators == nil {
		t.Error("Validators map not initialized")
	}
	
	// Check timestamps are set
	if bm.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if bm.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewChain(t *testing.T) {
	chain, err := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}
	
	if chain.ID != "eth" {
		t.Errorf("Expected ID 'eth', got '%s'", chain.ID)
	}
	if chain.Name != "Ethereum" {
		t.Errorf("Expected Name 'Ethereum', got '%s'", chain.Name)
	}
	if chain.Symbol != "ETH" {
		t.Errorf("Expected Symbol 'ETH', got '%s'", chain.Symbol)
	}
	if chain.ChainID != 1 {
		t.Errorf("Expected ChainID 1, got %d", chain.ChainID)
	}
	if chain.RPCURL != "https://eth.rpc" {
		t.Errorf("Expected RPCURL 'https://eth.rpc', got '%s'", chain.RPCURL)
	}
	if chain.ExplorerURL != "https://etherscan.io" {
		t.Errorf("Expected ExplorerURL 'https://etherscan.io', got '%s'", chain.ExplorerURL)
	}
	if !chain.IsActive {
		t.Error("Expected IsActive to be true")
	}
	
	// Check timestamps are set
	if chain.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if chain.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewChainValidation(t *testing.T) {
	// Test empty ID
	_, err := NewChain("", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty name
	_, err = NewChain("eth", "", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test empty symbol
	_, err = NewChain("eth", "Ethereum", "", 1, "https://eth.rpc", "https://etherscan.io")
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// Test empty RPC URL
	_, err = NewChain("eth", "Ethereum", "ETH", 1, "", "https://etherscan.io")
	if err == nil {
		t.Error("Expected error for empty RPC URL")
	}
}

func TestNewAsset(t *testing.T) {
	totalSupply := big.NewInt(1000000000) // 1B tokens
	asset, err := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	if err != nil {
		t.Fatalf("NewAsset failed: %v", err)
	}
	
	if asset.ID != "usdc" {
		t.Errorf("Expected ID 'usdc', got '%s'", asset.ID)
	}
	if asset.Symbol != "USDC" {
		t.Errorf("Expected Symbol 'USDC', got '%s'", asset.Symbol)
	}
	if asset.Name != "USD Coin" {
		t.Errorf("Expected Name 'USD Coin', got '%s'", asset.Name)
	}
	if asset.Decimals != 6 {
		t.Errorf("Expected Decimals 6, got %d", asset.Decimals)
	}
	if asset.TotalSupply.Cmp(totalSupply) != 0 {
		t.Errorf("Expected TotalSupply %v, got %v", totalSupply, asset.TotalSupply)
	}
	if !asset.IsActive {
		t.Error("Expected IsActive to be true")
	}
	
	// Check timestamps are set
	if asset.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if asset.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewAssetValidation(t *testing.T) {
	validTotalSupply := big.NewInt(1000000000)
	
	// Test empty ID
	_, err := NewAsset("", "USDC", "USD Coin", 6, validTotalSupply)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty symbol
	_, err = NewAsset("usdc", "", "USD Coin", 6, validTotalSupply)
	if err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// Test empty name
	_, err = NewAsset("usdc", "USDC", "", 6, validTotalSupply)
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test nil total supply
	_, err = NewAsset("usdc", "USDC", "USD Coin", 6, nil)
	if err == nil {
		t.Error("Expected error for nil total supply")
	}
	
	// Test negative total supply
	_, err = NewAsset("usdc", "USDC", "USD Coin", 6, big.NewInt(-1000000000))
	if err == nil {
		t.Error("Expected error for negative total supply")
	}
}

func TestNewBridgeAsset(t *testing.T) {
	balance := big.NewInt(500000000) // 500M tokens
	bridgeAsset, err := NewBridgeAsset("usdc_eth", "usdc", "eth", "0x1234567890abcdef", balance)
	if err != nil {
		t.Fatalf("NewBridgeAsset failed: %v", err)
	}
	
	if bridgeAsset.ID != "usdc_eth" {
		t.Errorf("Expected ID 'usdc_eth', got '%s'", bridgeAsset.ID)
	}
	if bridgeAsset.AssetID != "usdc" {
		t.Errorf("Expected AssetID 'usdc', got '%s'", bridgeAsset.AssetID)
	}
	if bridgeAsset.ChainID != "eth" {
		t.Errorf("Expected ChainID 'eth', got '%s'", bridgeAsset.ChainID)
	}
	if bridgeAsset.ContractAddr != "0x1234567890abcdef" {
		t.Errorf("Expected ContractAddr '0x1234567890abcdef', got '%s'", bridgeAsset.ContractAddr)
	}
	if bridgeAsset.Balance.Cmp(balance) != 0 {
		t.Errorf("Expected Balance %v, got %v", balance, bridgeAsset.Balance)
	}
	if !bridgeAsset.IsActive {
		t.Error("Expected IsActive to be true")
	}
	
	// Check timestamps are set
	if bridgeAsset.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if bridgeAsset.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewBridgeAssetValidation(t *testing.T) {
	validBalance := big.NewInt(500000000)
	
	// Test empty ID
	_, err := NewBridgeAsset("", "usdc", "eth", "0x1234567890abcdef", validBalance)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty asset ID
	_, err = NewBridgeAsset("usdc_eth", "", "eth", "0x1234567890abcdef", validBalance)
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
	
	// Test empty chain ID
	_, err = NewBridgeAsset("usdc_eth", "usdc", "", "0x1234567890abcdef", validBalance)
	if err == nil {
		t.Error("Expected error for empty chain ID")
	}
	
	// Test empty contract address
	_, err = NewBridgeAsset("usdc_eth", "usdc", "eth", "", validBalance)
	if err == nil {
		t.Error("Expected error for empty contract address")
	}
	
	// Test nil balance
	_, err = NewBridgeAsset("usdc_eth", "usdc", "eth", "0x1234567890abcdef", nil)
	if err == nil {
		t.Error("Expected error for nil balance")
	}
	
	// Test negative balance
	_, err = NewBridgeAsset("usdc_eth", "usdc", "eth", "0x1234567890abcdef", big.NewInt(-500000000))
	if err == nil {
		t.Error("Expected error for negative balance")
	}
}

func TestNewBridge(t *testing.T) {
	minTransfer := big.NewInt(1000000)   // 1M tokens
	maxTransfer := big.NewInt(100000000) // 100M tokens
	bridgeFee := big.NewInt(10000)       // 10K tokens
	
	bridge, err := NewBridge("eth_bsc", "ETH-BSC Bridge", "Bridge between Ethereum and BSC", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	if err != nil {
		t.Fatalf("NewBridge failed: %v", err)
	}
	
	if bridge.ID != "eth_bsc" {
		t.Errorf("Expected ID 'eth_bsc', got '%s'", bridge.ID)
	}
	if bridge.Name != "ETH-BSC Bridge" {
		t.Errorf("Expected Name 'ETH-BSC Bridge', got '%s'", bridge.Name)
	}
	if bridge.Description != "Bridge between Ethereum and BSC" {
		t.Errorf("Expected Description 'Bridge between Ethereum and BSC', got '%s'", bridge.Description)
	}
	if bridge.SourceChainID != "eth" {
		t.Errorf("Expected SourceChainID 'eth', got '%s'", bridge.SourceChainID)
	}
	if bridge.TargetChainID != "bsc" {
		t.Errorf("Expected TargetChainID 'bsc', got '%s'", bridge.TargetChainID)
	}
	if bridge.Status != BridgeActive {
		t.Errorf("Expected Status BridgeActive, got %v", bridge.Status)
	}
	if bridge.MinTransfer.Cmp(minTransfer) != 0 {
		t.Errorf("Expected MinTransfer %v, got %v", minTransfer, bridge.MinTransfer)
	}
	if bridge.MaxTransfer.Cmp(maxTransfer) != 0 {
		t.Errorf("Expected MaxTransfer %v, got %v", maxTransfer, bridge.MaxTransfer)
	}
	if bridge.BridgeFee.Cmp(bridgeFee) != 0 {
		t.Errorf("Expected BridgeFee %v, got %v", bridgeFee, bridge.BridgeFee)
	}
	if !bridge.IsActive {
		t.Error("Expected IsActive to be true")
	}
	
	// Check timestamps are set
	if bridge.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if bridge.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewBridgeValidation(t *testing.T) {
	minTransfer := big.NewInt(1000000)   // 1M tokens
	maxTransfer := big.NewInt(100000000) // 100M tokens
	bridgeFee := big.NewInt(10000)       // 10K tokens
	
	// Test empty ID
	_, err := NewBridge("", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty name
	_, err = NewBridge("eth_bsc", "", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test empty source chain ID
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "", "bsc", minTransfer, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for empty source chain ID")
	}
	
	// Test empty target chain ID
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "", minTransfer, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for empty target chain ID")
	}
	
	// Test same source and target chain
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "eth", minTransfer, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for same source and target chain")
	}
	
	// Test nil min transfer
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", nil, maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for nil min transfer")
	}
	
	// Test negative min transfer
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", big.NewInt(-1000000), maxTransfer, bridgeFee)
	if err == nil {
		t.Error("Expected error for negative min transfer")
	}
	
	// Test nil max transfer
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, nil, bridgeFee)
	if err == nil {
		t.Error("Expected error for nil max transfer")
	}
	
	// Test negative max transfer
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, big.NewInt(-100000000), bridgeFee)
	if err == nil {
		t.Error("Expected error for negative max transfer")
	}
	
	// Test min transfer greater than max transfer
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", big.NewInt(200000000), big.NewInt(100000000), bridgeFee)
	if err == nil {
		t.Error("Expected error for min transfer greater than max transfer")
	}
	
	// Test nil bridge fee
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, nil)
	if err == nil {
		t.Error("Expected error for nil bridge fee")
	}
	
	// Test negative bridge fee
	_, err = NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, big.NewInt(-10000))
	if err == nil {
		t.Error("Expected error for negative bridge fee")
	}
}

func TestNewTransfer(t *testing.T) {
	amount := big.NewInt(50000000) // 50M tokens
	
	transfer, err := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err != nil {
		t.Fatalf("NewTransfer failed: %v", err)
	}
	
	if transfer.ID != "tx1" {
		t.Errorf("Expected ID 'tx1', got '%s'", transfer.ID)
	}
	if transfer.SourceChainID != "eth" {
		t.Errorf("Expected SourceChainID 'eth', got '%s'", transfer.SourceChainID)
	}
	if transfer.TargetChainID != "bsc" {
		t.Errorf("Expected TargetChainID 'bsc', got '%s'", transfer.TargetChainID)
	}
	if transfer.AssetID != "usdc" {
		t.Errorf("Expected AssetID 'usdc', got '%s'", transfer.AssetID)
	}
	if transfer.Amount.Cmp(amount) != 0 {
		t.Errorf("Expected Amount %v, got %v", amount, transfer.Amount)
	}
	if transfer.Sender != "0x1234567890abcdef" {
		t.Errorf("Expected Sender '0x1234567890abcdef', got '%s'", transfer.Sender)
	}
	if transfer.Recipient != "0xfedcba0987654321" {
		t.Errorf("Expected Recipient '0xfedcba0987654321', got '%s'", transfer.Recipient)
	}
	if transfer.Status != TransferPending {
		t.Errorf("Expected Status TransferPending, got %v", transfer.Status)
	}
	if transfer.SourceTxHash != "" {
		t.Errorf("Expected empty SourceTxHash, got '%s'", transfer.SourceTxHash)
	}
	if transfer.TargetTxHash != "" {
		t.Errorf("Expected empty TargetTxHash, got '%s'", transfer.TargetTxHash)
	}
	if transfer.BridgeFee.Sign() != 0 {
		t.Errorf("Expected BridgeFee 0, got %v", transfer.BridgeFee)
	}
	if transfer.GasFee.Sign() != 0 {
		t.Errorf("Expected GasFee 0, got %v", transfer.GasFee)
	}
	if transfer.CompletedAt != nil {
		t.Error("Expected CompletedAt to be nil")
	}
	
	// Check timestamps are set
	if transfer.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if transfer.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewTransferValidation(t *testing.T) {
	validAmount := big.NewInt(50000000)
	
	// Test empty ID
	_, err := NewTransfer("", "eth", "bsc", "usdc", validAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty source chain ID
	_, err = NewTransfer("tx1", "", "bsc", "usdc", validAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for empty source chain ID")
	}
	
	// Test empty target chain ID
	_, err = NewTransfer("tx1", "eth", "", "usdc", validAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for empty target chain ID")
	}
	
	// Test same source and target chain
	_, err = NewTransfer("tx1", "eth", "eth", "usdc", validAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for same source and target chain")
	}
	
	// Test empty asset ID
	_, err = NewTransfer("tx1", "eth", "bsc", "", validAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
	
	// Test nil amount
	_, err = NewTransfer("tx1", "eth", "bsc", "usdc", nil, "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for nil amount")
	}
	
	// Test zero amount
	_, err = NewTransfer("tx1", "eth", "bsc", "usdc", big.NewInt(0), "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for zero amount")
	}
	
	// Test negative amount
	_, err = NewTransfer("tx1", "eth", "bsc", "usdc", big.NewInt(-50000000), "0x1234567890abcdef", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for negative amount")
	}
	
	// Test empty sender
	_, err = NewTransfer("tx1", "eth", "bsc", "usdc", validAmount, "", "0xfedcba0987654321")
	if err == nil {
		t.Error("Expected error for empty sender")
	}
	
	// Test empty recipient
	_, err = NewTransfer("tx1", "eth", "bsc", "usdc", validAmount, "0x1234567890abcdef", "")
	if err == nil {
		t.Error("Expected error for empty recipient")
	}
}

func TestNewValidator(t *testing.T) {
	stakeAmount := big.NewInt(1000000) // 1M tokens
	
	validator, err := NewValidator("val1", "0x1234567890abcdef", "eth", stakeAmount)
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}
	
	if validator.ID != "val1" {
		t.Errorf("Expected ID 'val1', got '%s'", validator.ID)
	}
	if validator.Address != "0x1234567890abcdef" {
		t.Errorf("Expected Address '0x1234567890abcdef', got '%s'", validator.Address)
	}
	if validator.ChainID != "eth" {
		t.Errorf("Expected ChainID 'eth', got '%s'", validator.ChainID)
	}
	if validator.StakeAmount.Cmp(stakeAmount) != 0 {
		t.Errorf("Expected StakeAmount %v, got %v", stakeAmount, validator.StakeAmount)
	}
	if !validator.IsActive {
		t.Error("Expected IsActive to be true")
	}
	
	// Check timestamps are set
	if validator.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if validator.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestNewValidatorValidation(t *testing.T) {
	validStakeAmount := big.NewInt(1000000)
	
	// Test empty ID
	_, err := NewValidator("", "0x1234567890abcdef", "eth", validStakeAmount)
	if err == nil {
		t.Error("Expected error for empty ID")
	}
	
	// Test empty address
	_, err = NewValidator("val1", "", "eth", validStakeAmount)
	if err == nil {
		t.Error("Expected error for empty address")
	}
	
	// Test empty chain ID
	_, err = NewValidator("val1", "0x1234567890abcdef", "", validStakeAmount)
	if err == nil {
		t.Error("Expected error for empty chain ID")
	}
	
	// Test nil stake amount
	_, err = NewValidator("val1", "0x1234567890abcdef", "eth", nil)
	if err == nil {
		t.Error("Expected error for nil stake amount")
	}
	
	// Test negative stake amount
	_, err = NewValidator("val1", "0x1234567890abcdef", "eth", big.NewInt(-1000000))
	if err == nil {
		t.Error("Expected error for negative stake amount")
	}
}

func TestBridgeManagerAddChain(t *testing.T) {
	bm := NewBridgeManager()
	chain, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	
	err := bm.AddChain(chain)
	if err != nil {
		t.Fatalf("AddChain failed: %v", err)
	}
	
	if _, exists := bm.Chains["eth"]; !exists {
		t.Error("Chain not added to manager")
	}
	
	// Test adding duplicate chain
	err = bm.AddChain(chain)
	if err == nil {
		t.Error("Expected error for duplicate chain")
	}
	
	// Test adding nil chain
	err = bm.AddChain(nil)
	if err == nil {
		t.Error("Expected error for nil chain")
	}
}

func TestBridgeManagerAddAsset(t *testing.T) {
	bm := NewBridgeManager()
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	
	err := bm.AddAsset(asset)
	if err != nil {
		t.Fatalf("AddAsset failed: %v", err)
	}
	
	if _, exists := bm.Assets["usdc"]; !exists {
		t.Error("Asset not added to manager")
	}
	
	// Test adding duplicate asset
	err = bm.AddAsset(asset)
	if err == nil {
		t.Error("Expected error for duplicate asset")
	}
	
	// Test adding nil asset
	err = bm.AddAsset(nil)
	if err == nil {
		t.Error("Expected error for nil asset")
	}
}

func TestBridgeManagerAddBridgeAsset(t *testing.T) {
	bm := NewBridgeManager()
	
	// Add required dependencies first
	chain, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	bm.AddChain(chain)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	balance := big.NewInt(500000000)
	bridgeAsset, _ := NewBridgeAsset("usdc_eth", "usdc", "eth", "0x1234567890abcdef", balance)
	
	err := bm.AddBridgeAsset(bridgeAsset)
	if err != nil {
		t.Fatalf("AddBridgeAsset failed: %v", err)
	}
	
	if _, exists := bm.BridgeAssets["usdc_eth"]; !exists {
		t.Error("Bridge asset not added to manager")
	}
	
	// Test adding bridge asset with non-existent asset
	invalidBridgeAsset, _ := NewBridgeAsset("invalid", "non_existent", "eth", "0x1234567890abcdef", balance)
	err = bm.AddBridgeAsset(invalidBridgeAsset)
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
	
	// Test adding bridge asset with non-existent chain
	invalidBridgeAsset2, _ := NewBridgeAsset("invalid2", "usdc", "non_existent", "0x1234567890abcdef", balance)
	err = bm.AddBridgeAsset(invalidBridgeAsset2)
	if err == nil {
		t.Error("Expected error for non-existent chain")
	}
}

func TestBridgeManagerAddBridge(t *testing.T) {
	bm := NewBridgeManager()
	
	// Add required dependencies first
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	
	err := bm.AddBridge(bridge)
	if err != nil {
		t.Fatalf("AddBridge failed: %v", err)
	}
	
	if _, exists := bm.Bridges["eth_bsc"]; !exists {
		t.Error("Bridge not added to manager")
	}
	
	// Test adding bridge with non-existent source chain
	invalidBridge1, _ := NewBridge("invalid", "Invalid Bridge", "Description", "non_existent", "bsc", minTransfer, maxTransfer, bridgeFee)
	err = bm.AddBridge(invalidBridge1)
	if err == nil {
		t.Error("Expected error for non-existent source chain")
	}
	
	// Test adding bridge with non-existent target chain
	invalidBridge2, _ := NewBridge("invalid2", "Invalid Bridge", "Description", "eth", "non_existent", minTransfer, maxTransfer, bridgeFee)
	err = bm.AddBridge(invalidBridge2)
	if err == nil {
		t.Error("Expected error for non-existent target chain")
	}
}

func TestBridgeManagerInitiateTransfer(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	amount := big.NewInt(50000000)
	transfer, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	
	err := bm.InitiateTransfer(transfer)
	if err != nil {
		t.Fatalf("InitiateTransfer failed: %v", err)
	}
	
	if _, exists := bm.Transfers["tx1"]; !exists {
		t.Error("Transfer not added to manager")
	}
	
	// Test transfer amount below minimum
	smallAmount := big.NewInt(500000) // Below min transfer
	smallTransfer, _ := NewTransfer("tx2", "eth", "bsc", "usdc", smallAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	err = bm.InitiateTransfer(smallTransfer)
	if err == nil {
		t.Error("Expected error for transfer amount below minimum")
	}
	
	// Test transfer amount above maximum
	largeAmount := big.NewInt(200000000) // Above max transfer
	largeTransfer, _ := NewTransfer("tx3", "eth", "bsc", "usdc", largeAmount, "0x1234567890abcdef", "0xfedcba0987654321")
	err = bm.InitiateTransfer(largeTransfer)
	if err == nil {
		t.Error("Expected error for transfer amount above maximum")
	}
}

func TestBridgeManagerTransferFlow(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	amount := big.NewInt(50000000)
	transfer, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	bm.InitiateTransfer(transfer)
	
	// Test confirming transfer
	err := bm.ConfirmTransfer("tx1", "0xabcdef1234567890")
	if err != nil {
		t.Fatalf("ConfirmTransfer failed: %v", err)
	}
	
	retrievedTransfer, _ := bm.GetTransfer("tx1")
	if retrievedTransfer.Status != TransferConfirmed {
		t.Errorf("Expected status TransferConfirmed, got %v", retrievedTransfer.Status)
	}
	if retrievedTransfer.SourceTxHash != "0xabcdef1234567890" {
		t.Errorf("Expected source tx hash '0xabcdef1234567890', got '%s'", retrievedTransfer.SourceTxHash)
	}
	
	// Test processing transfer
	err = bm.ProcessTransfer("tx1", "0xfedcba0987654321")
	if err != nil {
		t.Fatalf("ProcessTransfer failed: %v", err)
	}
	
	retrievedTransfer, _ = bm.GetTransfer("tx1")
	if retrievedTransfer.Status != TransferProcessing {
		t.Errorf("Expected status TransferProcessing, got %v", retrievedTransfer.Status)
	}
	if retrievedTransfer.TargetTxHash != "0xfedcba0987654321" {
		t.Errorf("Expected target tx hash '0xfedcba0987654321', got '%s'", retrievedTransfer.TargetTxHash)
	}
	
	// Test completing transfer
	err = bm.CompleteTransfer("tx1")
	if err != nil {
		t.Fatalf("CompleteTransfer failed: %v", err)
	}
	
	retrievedTransfer, _ = bm.GetTransfer("tx1")
	if retrievedTransfer.Status != TransferCompleted {
		t.Errorf("Expected status TransferCompleted, got %v", retrievedTransfer.Status)
	}
	if retrievedTransfer.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestBridgeManagerGetTransfersByStatus(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	// Add multiple transfers
	amount := big.NewInt(50000000)
	transfer1, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	transfer2, _ := NewTransfer("tx2", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	bm.InitiateTransfer(transfer1)
	bm.InitiateTransfer(transfer2)
	
	// Confirm one transfer
	bm.ConfirmTransfer("tx1", "0xabcdef1234567890")
	
	// Test getting pending transfers
	pendingTransfers := bm.GetTransfersByStatus(TransferPending)
	if len(pendingTransfers) != 1 {
		t.Errorf("Expected 1 pending transfer, got %d", len(pendingTransfers))
	}
	
	// Test getting confirmed transfers
	confirmedTransfers := bm.GetTransfersByStatus(TransferConfirmed)
	if len(confirmedTransfers) != 1 {
		t.Errorf("Expected 1 confirmed transfer, got %d", len(confirmedTransfers))
	}
}

func TestBridgeManagerGetTransfersByChain(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	// Add transfers
	amount := big.NewInt(50000000)
	transfer1, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	transfer2, _ := NewTransfer("tx2", "bsc", "eth", "usdc", amount, "0xfedcba0987654321", "0x1234567890abcdef")
	bm.InitiateTransfer(transfer1)
	bm.InitiateTransfer(transfer2)
	
	// Test getting transfers by source chain
	ethTransfers := bm.GetTransfersByChain("eth")
	if len(ethTransfers) != 1 {
		t.Errorf("Expected 1 transfer for ETH chain, got %d", len(ethTransfers))
	}
	
	// Test getting transfers by target chain
	bscTransfers := bm.GetTransfersByChain("bsc")
	if len(bscTransfers) != 1 {
		t.Errorf("Expected 1 transfer for BSC chain, got %d", len(bscTransfers))
	}
	
	// Test getting transfers by non-existent chain
	nonExistentTransfers := bm.GetTransfersByChain("non_existent")
	if len(nonExistentTransfers) != 0 {
		t.Errorf("Expected 0 transfers for non-existent chain, got %d", len(nonExistentTransfers))
	}
}

func TestBridgeManagerGetTotalVolumeAndFees(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	// Add and complete transfers
	amount := big.NewInt(50000000)
	transfer1, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	transfer2, _ := NewTransfer("tx2", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	bm.InitiateTransfer(transfer1)
	bm.InitiateTransfer(transfer2)
	
	// Complete both transfers
	bm.ConfirmTransfer("tx1", "0xabcdef1234567890")
	bm.ProcessTransfer("tx1", "0xfedcba0987654321")
	bm.CompleteTransfer("tx1")
	
	bm.ConfirmTransfer("tx2", "0xabcdef1234567890")
	bm.ProcessTransfer("tx2", "0xfedcba0987654321")
	bm.CompleteTransfer("tx2")
	
	// Test total volume
	totalVolume := bm.GetTotalVolume()
	expectedVolume := big.NewInt(100000000) // 2 * 50M
	if totalVolume.Cmp(expectedVolume) != 0 {
		t.Errorf("Expected total volume %v, got %v", expectedVolume, totalVolume)
	}
	
	// Test total fees
	totalFees := bm.GetTotalFees()
	expectedFees := big.NewInt(20000) // 2 * 10K
	if totalFees.Cmp(expectedFees) != 0 {
		t.Errorf("Expected total fees %v, got %v", expectedFees, totalFees)
	}
}

func TestGenerateTransferID(t *testing.T) {
	sourceChainID := "eth"
	targetChainID := "bsc"
	assetID := "usdc"
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	transferID := GenerateTransferID(sourceChainID, targetChainID, assetID, timestamp)
	
	if transferID == "" {
		t.Error("Generated transfer ID is empty")
	}
	
	// Test that same parameters generate same ID
	transferID2 := GenerateTransferID(sourceChainID, targetChainID, assetID, timestamp)
	if transferID != transferID2 {
		t.Error("Same parameters should generate same transfer ID")
	}
	
	// Test that different parameters generate different IDs
	transferID3 := GenerateTransferID("polygon", targetChainID, assetID, timestamp)
	if transferID == transferID3 {
		t.Error("Different parameters should generate different transfer IDs")
	}
}

func TestValidateTransfer(t *testing.T) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	amount := big.NewInt(50000000)
	transfer, _ := NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	
	// Test valid transfer
	err := bm.ValidateTransfer(transfer)
	if err != nil {
		t.Errorf("Expected no error for valid transfer, got %v", err)
	}
	
	// Test transfer with non-existent source chain
	invalidTransfer1, _ := NewTransfer("tx2", "non_existent", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	err = bm.ValidateTransfer(invalidTransfer1)
	if err == nil {
		t.Error("Expected error for non-existent source chain")
	}
	
	// Test transfer with non-existent asset
	invalidTransfer2, _ := NewTransfer("tx3", "eth", "bsc", "non_existent", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	err = bm.ValidateTransfer(invalidTransfer2)
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
}

// Benchmarks
func BenchmarkNewBridgeManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBridgeManager()
	}
}

func BenchmarkNewChain(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	}
}

func BenchmarkNewAsset(b *testing.B) {
	totalSupply := big.NewInt(1000000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	}
}

func BenchmarkNewBridge(b *testing.B) {
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	}
}

func BenchmarkNewTransfer(b *testing.B) {
	amount := big.NewInt(50000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTransfer("tx1", "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
	}
}

func BenchmarkBridgeManagerAddChain(b *testing.B) {
	bm := NewBridgeManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create unique chain for each iteration
		uniqueChain, _ := NewChain("eth"+string(rune(i)), "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
		bm.AddChain(uniqueChain)
	}
}

func BenchmarkBridgeManagerInitiateTransfer(b *testing.B) {
	bm := NewBridgeManager()
	
	// Setup test environment
	chain1, _ := NewChain("eth", "Ethereum", "ETH", 1, "https://eth.rpc", "https://etherscan.io")
	chain2, _ := NewChain("bsc", "BSC", "BNB", 56, "https://bsc.rpc", "https://bscscan.com")
	bm.AddChain(chain1)
	bm.AddChain(chain2)
	
	totalSupply := big.NewInt(1000000000)
	asset, _ := NewAsset("usdc", "USDC", "USD Coin", 6, totalSupply)
	bm.AddAsset(asset)
	
	minTransfer := big.NewInt(1000000)
	maxTransfer := big.NewInt(100000000)
	bridgeFee := big.NewInt(10000)
	bridge, _ := NewBridge("eth_bsc", "ETH-BSC Bridge", "Description", "eth", "bsc", minTransfer, maxTransfer, bridgeFee)
	bm.AddBridge(bridge)
	
	amount := big.NewInt(50000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transfer, _ := NewTransfer("tx"+string(rune(i)), "eth", "bsc", "usdc", amount, "0x1234567890abcdef", "0xfedcba0987654321")
		bm.InitiateTransfer(transfer)
	}
}

func BenchmarkGenerateTransferID(b *testing.B) {
	sourceChainID := "eth"
	targetChainID := "bsc"
	assetID := "usdc"
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateTransferID(sourceChainID, targetChainID, assetID, timestamp)
	}
}
