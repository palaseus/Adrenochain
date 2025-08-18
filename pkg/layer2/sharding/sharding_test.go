package sharding

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShardingManager(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.shards)
	assert.NotNil(t, sm.crossShardTxs)
	assert.NotNil(t, sm.shardSyncs)
	assert.NotNil(t, sm.metrics)
	assert.NotNil(t, sm.txQueue)
	assert.NotNil(t, sm.syncQueue)
	assert.NotNil(t, sm.metricsUpdater)
}

func TestCreateShard(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test valid shard creation
	shard := &Shard{
		Name:          "Test Shard",
		Description:   "A test shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}

	err := sm.CreateShard(shard)
	require.NoError(t, err)
	assert.NotEmpty(t, shard.ID)
	assert.Equal(t, ShardActive, shard.Status)
	assert.Equal(t, big.NewInt(0), shard.CurrentLoad)
	assert.Equal(t, uint64(0), shard.BlockHeight)
	assert.Empty(t, shard.LastBlockHash)
	assert.NotZero(t, shard.CreatedAt)
	assert.NotZero(t, shard.UpdatedAt)

	// Test shard retrieval
	retrievedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, shard.ID, retrievedShard.ID)
	assert.Equal(t, shard.Name, retrievedShard.Name)

	// Test metrics initialization
	metrics, err := sm.GetShardMetrics(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, shard.ID, metrics.ShardID)
	assert.Equal(t, 0.0, metrics.TPS)
	assert.Equal(t, 0, metrics.ValidatorCount)
}

func TestCreateShardValidation(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test missing name
	shard := &Shard{
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard name is required")

	// Test missing type
	shard = &Shard{
		Name:          "Test Shard",
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err = sm.CreateShard(shard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard type is required")

	// Test missing consensus type
	shard = &Shard{
		Name:     "Test Shard",
		Type:     ExecutionShard,
		Capacity: big.NewInt(1000),
	}
	err = sm.CreateShard(shard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consensus type is required")

	// Test invalid capacity
	shard = &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(0),
	}
	err = sm.CreateShard(shard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard must have positive capacity")

	// Test negative capacity
	shard = &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(-100),
	}
	err = sm.CreateShard(shard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard must have positive capacity")
}

func TestGetShard(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test getting non-existent shard
	_, err := sm.GetShard("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")

	// Test getting existing shard
	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err = sm.CreateShard(shard)
	require.NoError(t, err)

	retrievedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, shard.ID, retrievedShard.ID)
}

func TestGetShards(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create multiple shards
	shard1 := &Shard{
		Name:          "Execution Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Data Shard",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}
	shard3 := &Shard{
		Name:          "Consensus Shard",
		Type:          ConsensusShard,
		ConsensusType: DPoSConsensus,
		Capacity:      big.NewInt(200),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	err = sm.CreateShard(shard3)
	require.NoError(t, err)

	// Test getting all shards
	allShards := sm.GetShards("", "")
	assert.Len(t, allShards, 3)

	// Test filtering by status
	activeShards := sm.GetShards(ShardActive, "")
	assert.Len(t, activeShards, 3)

	// Test filtering by type
	executionShards := sm.GetShards("", ExecutionShard)
	assert.Len(t, executionShards, 1)
	assert.Equal(t, "Execution Shard", executionShards[0].Name)

	// Test filtering by both status and type
	activeExecutionShards := sm.GetShards(ShardActive, ExecutionShard)
	assert.Len(t, activeExecutionShards, 1)
	assert.Equal(t, "Execution Shard", activeExecutionShards[0].Name)
}

func TestUpdateShardStatus(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test updating status
	err = sm.UpdateShardStatus(shard.ID, ShardPaused)
	require.NoError(t, err)

	updatedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, ShardPaused, updatedShard.Status)

	// Test updating non-existent shard
	err = sm.UpdateShardStatus("non-existent", ShardActive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")
}

func TestAddValidator(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test adding validator
	err = sm.AddValidator(shard.ID, "validator1")
	require.NoError(t, err)

	updatedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Len(t, updatedShard.Validators, 1)
	assert.Equal(t, "validator1", updatedShard.Validators[0])

	// Test adding duplicate validator
	err = sm.AddValidator(shard.ID, "validator1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validator already exists")

	// Test adding validator to inactive shard
	err = sm.UpdateShardStatus(shard.ID, ShardPaused)
	require.NoError(t, err)

	err = sm.AddValidator(shard.ID, "validator2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard is not active")

	// Test adding validator to non-existent shard
	err = sm.AddValidator("non-existent", "validator3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")
}

func TestRemoveValidator(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Add validators
	err = sm.AddValidator(shard.ID, "validator1")
	require.NoError(t, err)
	err = sm.AddValidator(shard.ID, "validator2")
	require.NoError(t, err)

	// Test removing validator
	err = sm.RemoveValidator(shard.ID, "validator1")
	require.NoError(t, err)

	updatedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Len(t, updatedShard.Validators, 1)
	assert.Equal(t, "validator2", updatedShard.Validators[0])

	// Test removing non-existent validator
	err = sm.RemoveValidator(shard.ID, "validator3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validator not found")

	// Test removing validator from inactive shard
	err = sm.UpdateShardStatus(shard.ID, ShardPaused)
	require.NoError(t, err)

	err = sm.RemoveValidator(shard.ID, "validator2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard is not active")

	// Test removing validator from non-existent shard
	err = sm.RemoveValidator("non-existent", "validator1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")
}

func TestLinkShards(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create two shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)

	// Test linking shards
	err = sm.LinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Verify links were created
	updatedShard1, err := sm.GetShard(shard1.ID)
	require.NoError(t, err)
	updatedShard2, err := sm.GetShard(shard2.ID)
	require.NoError(t, err)

	assert.Len(t, updatedShard1.CrossShardLinks, 1)
	assert.Equal(t, shard2.ID, updatedShard1.CrossShardLinks[0])
	assert.Len(t, updatedShard2.CrossShardLinks, 1)
	assert.Equal(t, shard1.ID, updatedShard2.CrossShardLinks[0])

	// Test linking non-existent shard
	err = sm.LinkShards("non-existent", shard1.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard 1 not found")

	// Test linking inactive shard
	err = sm.UpdateShardStatus(shard1.ID, ShardPaused)
	require.NoError(t, err)

	err = sm.LinkShards(shard1.ID, shard2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "both shards must be active")
}

func TestUnlinkShards(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create and link two shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	err = sm.LinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Test unlinking shards
	err = sm.UnlinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Verify links were removed
	updatedShard1, err := sm.GetShard(shard1.ID)
	require.NoError(t, err)
	updatedShard2, err := sm.GetShard(shard2.ID)
	require.NoError(t, err)

	assert.Len(t, updatedShard1.CrossShardLinks, 0)
	assert.Len(t, updatedShard2.CrossShardLinks, 0)

	// Test unlinking non-existent shard
	err = sm.UnlinkShards("non-existent", shard1.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard 1 not found")
}

func TestCreateCrossShardTransaction(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create and link two shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	err = sm.LinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Test creating valid cross-shard transaction
	tx := &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
		Nonce:           1,
		GasLimit:        21000,
		GasPrice:        big.NewInt(20000000000),
		Data:            []byte("transfer"),
		Signature:       "0xSignature",
	}

	err = sm.CreateCrossShardTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)
	assert.NotZero(t, tx.CreatedAt)
	assert.NotZero(t, tx.UpdatedAt)

	// Test retrieving the transaction
	retrievedTx, err := sm.GetCrossShardTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.ID, retrievedTx.ID)
	assert.Equal(t, tx.FromShard, retrievedTx.FromShard)
	assert.Equal(t, tx.ToShard, retrievedTx.ToShard)

	// Status may have changed due to async processing, but should be valid
	assert.Contains(t, []CrossShardTxStatus{CrossShardTxPending, CrossShardTxProcessing, CrossShardTxConfirmed}, retrievedTx.Status)
}

func TestCreateCrossShardTransactionValidation(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create two shards (not linked)
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)

	// Test missing from shard
	tx := &CrossShardTransaction{
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from shard is required")

	// Test missing to shard
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to shard is required")

	// Test missing transaction hash
	tx = &CrossShardTransaction{
		FromShard: shard1.ID,
		ToShard:   shard2.ID,
		Amount:    big.NewInt(100),
		Asset:     "ETH",
		Sender:    "0xSender",
		Recipient: "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction hash is required")

	// Test missing amount
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be positive")

	// Test missing sender
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender is required")

	// Test missing recipient
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient is required")

	// Test non-existent from shard
	tx = &CrossShardTransaction{
		FromShard:       "non-existent",
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from shard not found")

	// Test non-existent to shard
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         "non-existent",
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to shard not found")

	// Test unlinked shards
	tx = &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shards are not linked")
}

func TestGetCrossShardTransactions(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create and link shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	err = sm.LinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Create multiple transactions
	tx1 := &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender1",
		Recipient:       "0xRecipient1",
	}
	tx2 := &CrossShardTransaction{
		FromShard:       shard2.ID,
		ToShard:         shard1.ID,
		TransactionHash: "0xfedcba0987654321",
		Amount:          big.NewInt(200),
		Asset:           "BTC",
		Sender:          "0xSender2",
		Recipient:       "0xRecipient2",
	}

	err = sm.CreateCrossShardTransaction(tx1)
	require.NoError(t, err)
	err = sm.CreateCrossShardTransaction(tx2)
	require.NoError(t, err)

	// Test getting all transactions
	allTxs := sm.GetCrossShardTransactions("", "", "")
	assert.Len(t, allTxs, 2)

	// Test filtering by from shard
	fromShard1Txs := sm.GetCrossShardTransactions("", shard1.ID, "")
	assert.Len(t, fromShard1Txs, 1)
	assert.Equal(t, shard1.ID, fromShard1Txs[0].FromShard)

	// Test filtering by to shard
	toShard2Txs := sm.GetCrossShardTransactions("", "", shard2.ID)
	assert.Len(t, toShard2Txs, 1)
	assert.Equal(t, shard2.ID, toShard2Txs[0].ToShard)

	// Test filtering by status (may vary due to async processing)
	statusTxs := sm.GetCrossShardTransactions(CrossShardTxPending, "", "")
	if len(statusTxs) > 0 {
		assert.Equal(t, CrossShardTxPending, statusTxs[0].Status)
	}

	// Test filtering by both status and from shard (may vary due to async processing)
	statusFromShard1Txs := sm.GetCrossShardTransactions(CrossShardTxPending, shard1.ID, "")
	if len(statusFromShard1Txs) > 0 {
		assert.Equal(t, shard1.ID, statusFromShard1Txs[0].FromShard)
	}
}

func TestUpdateShardBlock(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test updating block information
	err = sm.UpdateShardBlock(shard.ID, 100, "0xBlockHash100")
	require.NoError(t, err)

	updatedShard, err := sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), updatedShard.BlockHeight)
	assert.Equal(t, "0xBlockHash100", updatedShard.LastBlockHash)
	assert.NotZero(t, updatedShard.LastBlockTime)

	// Test updating with higher block height
	err = sm.UpdateShardBlock(shard.ID, 200, "0xBlockHash200")
	require.NoError(t, err)

	updatedShard, err = sm.GetShard(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(200), updatedShard.BlockHeight)
	assert.Equal(t, "0xBlockHash200", updatedShard.LastBlockHash)

	// Test updating with lower block height
	err = sm.UpdateShardBlock(shard.ID, 150, "0xBlockHash150")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "block height must be greater than current height")

	// Test updating non-existent shard
	err = sm.UpdateShardBlock("non-existent", 100, "0xBlockHash")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")
}

func TestGetShardMetrics(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test getting metrics
	metrics, err := sm.GetShardMetrics(shard.ID)
	require.NoError(t, err)
	assert.Equal(t, shard.ID, metrics.ShardID)
	assert.Equal(t, 0.0, metrics.TPS)
	assert.Equal(t, 0, metrics.ValidatorCount)
	assert.Equal(t, big.NewInt(0), metrics.StakeAmount)
	assert.Equal(t, uint64(0), metrics.CrossShardTxCount)

	// Test getting metrics for non-existent shard
	_, err = sm.GetShardMetrics("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard metrics not found")
}

func TestGetAllShardMetrics(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create multiple shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)

	// Test getting all metrics
	allMetrics := sm.GetAllShardMetrics()
	assert.Len(t, allMetrics, 2)

	// Verify each shard has metrics
	shard1Found := false
	shard2Found := false
	for _, metrics := range allMetrics {
		if metrics.ShardID == shard1.ID {
			shard1Found = true
		}
		if metrics.ShardID == shard2.ID {
			shard2Found = true
		}
	}
	assert.True(t, shard1Found)
	assert.True(t, shard2Found)
}

func TestConcurrency(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	const numGoroutines = 10
	const numShards = 5
	const numValidators = 3

	// Test concurrent shard creation
	shardChan := make(chan *Shard, numShards)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			shard := &Shard{
				Name:          fmt.Sprintf("Shard %d", id),
				Type:          ExecutionShard,
				ConsensusType: PoSConsensus,
				Capacity:      big.NewInt(1000),
			}
			err := sm.CreateShard(shard)
			if err == nil {
				shardChan <- shard
			}
		}(i)
	}

	// Collect created shards
	var shards []*Shard
	for i := 0; i < numShards; i++ {
		select {
		case shard := <-shardChan:
			shards = append(shards, shard)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for shard creation")
		}
	}

	// Verify all shards were created
	assert.Len(t, shards, numShards)

	// Test concurrent validator addition
	if len(shards) > 0 {
		validatorChan := make(chan bool, numValidators*len(shards))
		for _, shard := range shards {
			for j := 0; j < numValidators; j++ {
				go func(s *Shard, validatorID int) {
					validatorAddr := fmt.Sprintf("validator_%s_%d", s.ID, validatorID)
					err := sm.AddValidator(s.ID, validatorAddr)
					validatorChan <- (err == nil)
				}(shard, j)
			}
		}

		// Wait for all validators to be added
		successCount := 0
		for i := 0; i < numValidators*len(shards); i++ {
			select {
			case success := <-validatorChan:
				if success {
					successCount++
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for validator addition")
			}
		}

		// Verify validators were added
		assert.Equal(t, numValidators*len(shards), successCount)
	}
}

func TestMemorySafety(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	const numShards = 100
	const numTransactions = 50

	// Create many shards
	var shards []*Shard
	for i := 0; i < numShards; i++ {
		shard := &Shard{
			Name:          fmt.Sprintf("Shard %d", i),
			Type:          ExecutionShard,
			ConsensusType: PoSConsensus,
			Capacity:      big.NewInt(1000),
		}
		err := sm.CreateShard(shard)
		require.NoError(t, err)
		shards = append(shards, shard)
	}

	// Create many cross-shard transactions
	for i := 0; i < numTransactions; i++ {
		if i+1 < len(shards) {
			// Link shards
			err := sm.LinkShards(shards[i].ID, shards[i+1].ID)
			if err == nil {
				tx := &CrossShardTransaction{
					FromShard:       shards[i].ID,
					ToShard:         shards[i+1].ID,
					TransactionHash: fmt.Sprintf("0x%x", i),
					Amount:          big.NewInt(int64(i + 1)),
					Asset:           "ETH",
					Sender:          fmt.Sprintf("sender_%d", i),
					Recipient:       fmt.Sprintf("recipient_%d", i),
				}
				sm.CreateCrossShardTransaction(tx)
			}
		}
	}

	// Verify memory usage is reasonable
	allShards := sm.GetShards("", "")
	assert.Len(t, allShards, numShards)

	allTxs := sm.GetCrossShardTransactions("", "", "")
	assert.Len(t, allTxs, numTransactions)

	allMetrics := sm.GetAllShardMetrics()
	assert.Len(t, allMetrics, numShards)
}

func TestEdgeCases(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test creating shard with empty ID (should generate one)
	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)
	assert.NotEmpty(t, shard.ID)

	// Test creating shard with custom ID
	customID := ShardID("custom-shard-id")
	shard2 := &Shard{
		ID:            customID,
		Name:          "Custom Shard",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	assert.Equal(t, customID, shard2.ID)

	// Test creating transaction with empty ID (should generate one)
	// Link shards first
	err = sm.LinkShards(shard.ID, shard2.ID)
	require.NoError(t, err)

	tx := &CrossShardTransaction{
		FromShard:       shard.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}
	err = sm.CreateCrossShardTransaction(tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)

	// Test getting non-existent transaction
	_, err = sm.GetCrossShardTransaction("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cross-shard transaction not found")

	// Test getting non-existent shard
	_, err = sm.GetShard("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shard not found")
}

func TestCleanup(t *testing.T) {
	sm := NewShardingManager()

	// Create some test data
	shard := &Shard{
		Name:          "Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test cleanup
	err = sm.Close()
	require.NoError(t, err)

	// Verify channels are closed
	select {
	case <-sm.txQueue:
		// Channel is closed
	default:
		t.Error("txQueue should be closed")
	}

	select {
	case <-sm.syncQueue:
		// Channel is closed
	default:
		t.Error("syncQueue should be closed")
	}

	select {
	case <-sm.metricsUpdater:
		// Channel is closed
	default:
		t.Error("metricsUpdater should be closed")
	}
}

func TestGetRandomID(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test that random IDs are generated
	id1 := sm.GetRandomID()
	id2 := sm.GetRandomID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // 16 bytes = 32 hex chars
	assert.Len(t, id2, 32)
}

func TestShardTypes(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test all shard types
	shardTypes := []ShardType{
		ExecutionShard,
		DataShard,
		ConsensusShard,
		BridgeShard,
		CustomShard,
	}

	for i, shardType := range shardTypes {
		shard := &Shard{
			Name:          fmt.Sprintf("Shard Type %d", i),
			Type:          shardType,
			ConsensusType: PoSConsensus,
			Capacity:      big.NewInt(1000),
		}
		err := sm.CreateShard(shard)
		require.NoError(t, err)
		assert.Equal(t, shardType, shard.Type)
	}
}

func TestConsensusTypes(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Test all consensus types
	consensusTypes := []ConsensusType{
		PoWConsensus,
		PoSConsensus,
		PoAConsensus,
		DPoSConsensus,
		CustomConsensus,
	}

	for i, consensusType := range consensusTypes {
		shard := &Shard{
			Name:          fmt.Sprintf("Consensus Shard %d", i),
			Type:          ExecutionShard,
			ConsensusType: consensusType,
			Capacity:      big.NewInt(1000),
		}
		err := sm.CreateShard(shard)
		require.NoError(t, err)
		assert.Equal(t, consensusType, shard.ConsensusType)
	}
}

func TestShardStatusTransitions(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	shard := &Shard{
		Name:          "Status Test Shard",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	err := sm.CreateShard(shard)
	require.NoError(t, err)

	// Test status transitions
	statuses := []ShardStatus{
		ShardActive,
		ShardPaused,
		ShardSyncing,
		ShardInactive,
		ShardTesting,
		ShardDeprecated,
		ShardActive, // Back to active
	}

	for _, status := range statuses {
		err = sm.UpdateShardStatus(shard.ID, status)
		require.NoError(t, err)

		updatedShard, err := sm.GetShard(shard.ID)
		require.NoError(t, err)
		assert.Equal(t, status, updatedShard.Status)
	}
}

func TestCrossShardTransactionStatusTransitions(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	// Create and link shards
	shard1 := &Shard{
		Name:          "Shard 1",
		Type:          ExecutionShard,
		ConsensusType: PoSConsensus,
		Capacity:      big.NewInt(1000),
	}
	shard2 := &Shard{
		Name:          "Shard 2",
		Type:          DataShard,
		ConsensusType: PoAConsensus,
		Capacity:      big.NewInt(500),
	}

	err := sm.CreateShard(shard1)
	require.NoError(t, err)
	err = sm.CreateShard(shard2)
	require.NoError(t, err)
	err = sm.LinkShards(shard1.ID, shard2.ID)
	require.NoError(t, err)

	// Create transaction
	tx := &CrossShardTransaction{
		FromShard:       shard1.ID,
		ToShard:         shard2.ID,
		TransactionHash: "0x1234567890abcdef",
		Amount:          big.NewInt(100),
		Asset:           "ETH",
		Sender:          "0xSender",
		Recipient:       "0xRecipient",
	}

	err = sm.CreateCrossShardTransaction(tx)
	require.NoError(t, err)

	// Verify initial status
	assert.Equal(t, CrossShardTxPending, tx.Status)

	// Wait for processing to complete
	time.Sleep(200 * time.Millisecond)

	// Verify final status
	retrievedTx, err := sm.GetCrossShardTransaction(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, CrossShardTxConfirmed, retrievedTx.Status)
}

func TestPerformance(t *testing.T) {
	sm := NewShardingManager()
	defer sm.Close()

	const numShards = 50
	const numTransactions = 100

	start := time.Now()

	// Create many shards quickly
	for i := 0; i < numShards; i++ {
		shard := &Shard{
			Name:          fmt.Sprintf("Shard %d", i),
			Type:          ExecutionShard,
			ConsensusType: PoSConsensus,
			Capacity:      big.NewInt(1000),
		}
		err := sm.CreateShard(shard)
		require.NoError(t, err)
	}

	// Create many transactions quickly
	shards := sm.GetShards("", "")
	transactionCount := 0
	for i := 0; i < len(shards)-1 && transactionCount < numTransactions; i++ {
		// Link shards
		err := sm.LinkShards(shards[i].ID, shards[i+1].ID)
		if err == nil {
			tx := &CrossShardTransaction{
				FromShard:       shards[i].ID,
				ToShard:         shards[i+1].ID,
				TransactionHash: fmt.Sprintf("0x%x", i),
				Amount:          big.NewInt(int64(i + 1)),
				Asset:           "ETH",
				Sender:          fmt.Sprintf("sender_%d", i),
				Recipient:       fmt.Sprintf("recipient_%d", i),
			}
			sm.CreateCrossShardTransaction(tx)
			transactionCount++
		}
	}

	duration := time.Since(start)
	t.Logf("Created %d shards and %d transactions in %v", numShards, numTransactions, duration)

	// Verify all data was created
	allShards := sm.GetShards("", "")
	assert.Len(t, allShards, numShards)

	allTxs := sm.GetCrossShardTransactions("", "", "")
	// We can only create transactions between consecutive shards
	expectedTxCount := numShards - 1
	if expectedTxCount > numTransactions {
		expectedTxCount = numTransactions
	}
	assert.Len(t, allTxs, expectedTxCount)

	// Performance should be reasonable
	assert.Less(t, duration, 5*time.Second, "Performance test took too long")
}
