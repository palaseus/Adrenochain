package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// explorerService implements the ExplorerService interface
type explorerService struct {
	dataProvider   BlockchainDataProvider
	cacheProvider  CacheProvider
	searchProvider SearchProvider
}

// NewExplorerService creates a new explorer service instance
func NewExplorerService(
	dataProvider BlockchainDataProvider,
	cacheProvider CacheProvider,
	searchProvider SearchProvider,
) ExplorerService {
	return &explorerService{
		dataProvider:   dataProvider,
		cacheProvider:  cacheProvider,
		searchProvider: searchProvider,
	}
}

// GetDashboard returns the main dashboard data
func (s *explorerService) GetDashboard(ctx context.Context) (*Dashboard, error) {
	// Get blockchain statistics
	stats, err := s.dataProvider.GetBlockchainStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain stats: %w", err)
	}

	// Get network information
	networkInfo, err := s.dataProvider.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	// Get latest block for recent activity
	latestBlock, err := s.dataProvider.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Create recent blocks list
	recentBlocks := []*BlockSummary{
		{
			Hash:          latestBlock.CalculateHash(),
			Height:        latestBlock.Header.Height,
			Timestamp:     latestBlock.Header.Timestamp,
			TxCount:       len(latestBlock.Transactions),
			Size:          uint64(len(latestBlock.Transactions) * 100), // Approximate size
			Difficulty:    latestBlock.Header.Difficulty,
			Confirmations: 0, // Latest block has 0 confirmations
		},
	}

	// Create recent transactions list
	var recentTxs []*TransactionSummary
	for _, tx := range latestBlock.Transactions {
		txSummary := &TransactionSummary{
			Hash:      tx.Hash,
			BlockHash: latestBlock.CalculateHash(),
			Height:    latestBlock.Header.Height,
			Timestamp: latestBlock.Header.Timestamp,
			Inputs:    len(tx.Inputs),
			Outputs:   len(tx.Outputs),
			Amount:    s.calculateTransactionAmount(tx),
			Fee:       tx.Fee,
			Status:    "confirmed",
		}
		recentTxs = append(recentTxs, txSummary)
	}

	dashboard := &Dashboard{
		Stats:        stats,
		RecentBlocks: recentBlocks,
		RecentTxs:    recentTxs,
		NetworkInfo:  networkInfo,
		LastUpdate:   time.Now(),
	}

	return dashboard, nil
}

// GetBlockDetails returns detailed information about a specific block
func (s *explorerService) GetBlockDetails(ctx context.Context, hash []byte) (*BlockDetails, error) {
	// Check cache first
	cacheKey := "block:" + string(hash)
	if cached, found := s.cacheProvider.Get(cacheKey); found {
		if blockDetails, ok := cached.(*BlockDetails); ok {
			return blockDetails, nil
		}
	}

	// Get block from blockchain
	block, err := s.dataProvider.GetBlock(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	// Get previous block
	var prevHash []byte
	if block.Header.Height > 0 {
		prevBlock, err := s.dataProvider.GetBlockByHeight(block.Header.Height - 1)
		if err == nil && prevBlock != nil {
			prevHash = prevBlock.CalculateHash()
		}
	}

	// Get next block
	var nextHash []byte
	nextBlock, err := s.dataProvider.GetBlockByHeight(block.Header.Height + 1)
	if err == nil && nextBlock != nil {
		nextHash = nextBlock.CalculateHash()
	}

	// Create block summary
	blockSummary := &BlockSummary{
		Hash:          block.CalculateHash(),
		Height:        block.Header.Height,
		Timestamp:     block.Header.Timestamp,
		TxCount:       len(block.Transactions),
		Size:          uint64(len(block.Transactions) * 100), // Approximate size
		Difficulty:    block.Header.Difficulty,
		Confirmations: s.calculateConfirmations(block.Header.Height),
	}

	// Create transaction summaries
	var txSummaries []*TransactionSummary
	for _, tx := range block.Transactions {
		txSummary := &TransactionSummary{
			Hash:      tx.Hash,
			BlockHash: block.CalculateHash(),
			Height:    block.Header.Height,
			Timestamp: block.Header.Timestamp,
			Inputs:    len(tx.Inputs),
			Outputs:   len(tx.Outputs),
			Amount:    s.calculateTransactionAmount(tx),
			Fee:       tx.Fee,
			Status:    "confirmed",
		}
		txSummaries = append(txSummaries, txSummary)
	}

	// Create block validation info
	validation := &BlockValidation{
		IsValid:      true, // Assuming blocks in the chain are valid
		Confirmations: s.calculateConfirmations(block.Header.Height),
		Finality:     s.determineFinality(block.Header.Height),
	}

	blockDetails := &BlockDetails{
		BlockSummary:  blockSummary,
		PrevHash:      prevHash,
		NextHash:      nextHash,
		MerkleRoot:    block.Header.MerkleRoot,
		Nonce:         block.Header.Nonce,
		Version:       block.Header.Version,
		Transactions:  txSummaries,
		Validation:    validation,
	}

	// Cache the result
	s.cacheProvider.Set(cacheKey, blockDetails, 5*time.Minute)

	return blockDetails, nil
}

// GetTransactionDetails returns detailed information about a specific transaction
func (s *explorerService) GetTransactionDetails(ctx context.Context, hash []byte) (*TransactionDetails, error) {
	// Check cache first
	cacheKey := "tx:" + string(hash)
	if cached, found := s.cacheProvider.Get(cacheKey); found {
		if txDetails, ok := cached.(*TransactionDetails); ok {
			return txDetails, nil
		}
	}

	// Get transaction from blockchain
	tx, err := s.dataProvider.GetTransaction(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Find the block containing this transaction
	block, err := s.findTransactionBlock(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to find transaction block: %w", err)
	}

	// Create transaction summary
	txSummary := &TransactionSummary{
		Hash:      tx.Hash,
		BlockHash: block.CalculateHash(),
		Height:    block.Header.Height,
		Timestamp: block.Header.Timestamp,
		Inputs:    len(tx.Inputs),
		Outputs:   len(tx.Outputs),
		Amount:    s.calculateTransactionAmount(tx),
		Fee:       tx.Fee,
		Status:    "confirmed",
	}

	// Create input details
	var inputDetails []*InputDetail
	for _, input := range tx.Inputs {
		inputDetail := &InputDetail{
			TxHash:  input.PrevTxHash,
			TxIndex: input.PrevTxIndex,
			Script:  input.ScriptSig,
			Address: s.extractAddressFromScript(input.ScriptSig),
			Amount:  0, // Would need to look up the actual amount from the previous output
		}
		inputDetails = append(inputDetails, inputDetail)
	}

	// Create output details
	var outputDetails []*OutputDetail
	for i, output := range tx.Outputs {
		outputDetail := &OutputDetail{
			Index:     uint32(i),
			Script:    output.ScriptPubKey,
			Address:   s.extractAddressFromScript(output.ScriptPubKey),
			Amount:    output.Value,
			Spent:     false, // Would need to check if this output has been spent
			SpentBy:   nil,   // Would need to track spending transactions
		}
		outputDetails = append(outputDetails, outputDetail)
	}

	// Create block info
	blockInfo := &BlockSummary{
		Hash:          block.CalculateHash(),
		Height:        block.Header.Height,
		Timestamp:     block.Header.Timestamp,
		TxCount:       len(block.Transactions),
		Size:          uint64(len(block.Transactions) * 100),
		Difficulty:    block.Header.Difficulty,
		Confirmations: s.calculateConfirmations(block.Header.Height),
	}

	txDetails := &TransactionDetails{
		TransactionSummary: txSummary,
		RawTx:             tx,
		InputDetails:      inputDetails,
		OutputDetails:     outputDetails,
		BlockInfo:         blockInfo,
	}

	// Cache the result
	s.cacheProvider.Set(cacheKey, txDetails, 5*time.Minute)

	return txDetails, nil
}

// GetAddressDetails returns detailed information about a specific address
func (s *explorerService) GetAddressDetails(ctx context.Context, address string) (*AddressDetails, error) {
	// Check cache first
	cacheKey := "address:" + address
	if cached, found := s.cacheProvider.Get(cacheKey); found {
		if addressDetails, ok := cached.(*AddressDetails); ok {
			return addressDetails, nil
		}
	}

	// Get address balance
	balance, err := s.dataProvider.GetAddressBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get address balance: %w", err)
	}

	// Get address UTXOs
	utxos, err := s.dataProvider.GetAddressUTXOs(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get address UTXOs: %w", err)
	}

	// Get address transactions
	transactions, err := s.dataProvider.GetAddressTransactions(address, 50, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get address transactions: %w", err)
	}

	// Create transaction summaries
	var txSummaries []*TransactionSummary
	for _, tx := range transactions {
		// Find the block containing this transaction
		block, err := s.findTransactionBlock(tx.Hash)
		if err != nil {
			continue // Skip transactions we can't find blocks for
		}

		txSummary := &TransactionSummary{
			Hash:      tx.Hash,
			BlockHash: block.CalculateHash(),
			Height:    block.Header.Height,
			Timestamp: block.Header.Timestamp,
			Inputs:    len(tx.Inputs),
			Outputs:   len(tx.Outputs),
			Amount:    s.calculateTransactionAmount(tx),
			Fee:       tx.Fee,
			Status:    "confirmed",
		}
		txSummaries = append(txSummaries, txSummary)
	}

	// Calculate first and last seen times
	var firstSeen, lastSeen time.Time
	if len(txSummaries) > 0 {
		firstSeen = txSummaries[len(txSummaries)-1].Timestamp
		lastSeen = txSummaries[0].Timestamp
	}

	// Calculate total received and sent (simplified)
	totalReceived := balance
	totalSent := uint64(0) // Would need more complex logic to calculate this

	addressSummary := &AddressSummary{
		Address:     address,
		Balance:     balance,
		TxCount:    len(txSummaries),
		FirstSeen:  firstSeen,
		LastSeen:   lastSeen,
	}

	addressDetails := &AddressDetails{
		AddressSummary: addressSummary,
		UTXOs:         utxos,
		Transactions:  txSummaries,
		TotalReceived: totalReceived,
		TotalSent:     totalSent,
	}

	// Cache the result
	s.cacheProvider.Set(cacheKey, addressDetails, 2*time.Minute)

	return addressDetails, nil
}

// GetBlocks returns a list of blocks with pagination
func (s *explorerService) GetBlocks(ctx context.Context, limit, offset int) ([]*BlockSummary, error) {
	height := s.dataProvider.GetBlockHeight()
	var blocks []*BlockSummary

	// Calculate start and end heights
	startHeight := height - uint64(offset)
	endHeight := startHeight - uint64(limit) + 1

	// Ensure we don't go below height 0
	if endHeight < 0 {
		endHeight = 0
	}

	// Get blocks from highest to lowest height
	for h := startHeight; h >= endHeight && h >= 0; h-- {
		block, err := s.dataProvider.GetBlockByHeight(h)
		if err != nil || block == nil {
			continue // Skip blocks we can't retrieve
		}

		blockSummary := &BlockSummary{
			Hash:          block.CalculateHash(),
			Height:        block.Header.Height,
			Timestamp:     block.Header.Timestamp,
			TxCount:       len(block.Transactions),
			Size:          uint64(len(block.Transactions) * 100),
			Difficulty:    block.Header.Difficulty,
			Confirmations: s.calculateConfirmations(block.Header.Height),
		}
		blocks = append(blocks, blockSummary)
	}

	return blocks, nil
}

// GetTransactions returns a list of transactions with pagination
func (s *explorerService) GetTransactions(ctx context.Context, limit, offset int) ([]*TransactionSummary, error) {
	height := s.dataProvider.GetBlockHeight()
	var transactions []*TransactionSummary

	// Calculate start and end heights
	startHeight := height - uint64(offset/5) // Assume average 5 transactions per block
	endHeight := startHeight - uint64(limit/5) + 1

	// Ensure we don't go below height 0
	if endHeight < 0 {
		endHeight = 0
	}

	// Get transactions from highest to lowest height
	for h := startHeight; h >= endHeight && h >= 0; h-- {
		block, err := s.dataProvider.GetBlockByHeight(h)
		if err != nil || block == nil {
			continue
		}

		for _, tx := range block.Transactions {
			if len(transactions) >= limit {
				break
			}

			txSummary := &TransactionSummary{
				Hash:      tx.Hash,
				BlockHash: block.CalculateHash(),
				Height:    block.Header.Height,
				Timestamp: block.Header.Timestamp,
				Inputs:    len(tx.Inputs),
				Outputs:   len(tx.Outputs),
				Amount:    s.calculateTransactionAmount(tx),
				Fee:       tx.Fee,
				Status:    "confirmed",
			}
			transactions = append(transactions, txSummary)
		}

		if len(transactions) >= limit {
			break
		}
	}

	return transactions, nil
}

// Search performs a search across the blockchain
func (s *explorerService) Search(ctx context.Context, query string) (*SearchResult, error) {
	// Check cache first
	cacheKey := "search:" + query
	if cached, found := s.cacheProvider.Get(cacheKey); found {
		if searchResult, ok := cached.(*SearchResult); ok {
			return searchResult, nil
		}
	}

	// Use search provider
	result, err := s.searchProvider.Search(query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Cache the result
	s.cacheProvider.Set(cacheKey, result, 1*time.Minute)

	return result, nil
}

// GetStatistics returns comprehensive blockchain statistics
func (s *explorerService) GetStatistics(ctx context.Context) (*Statistics, error) {
	// Get blockchain stats
	blockchainStats, err := s.dataProvider.GetBlockchainStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain stats: %w", err)
	}

	// Get network info
	networkInfo, err := s.dataProvider.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	// Get cache stats
	cacheStats := s.cacheProvider.GetStats()

	// Create performance stats
	performanceStats := &PerformanceStats{
		CacheHitRate: cacheStats.HitRate,
		// Other performance metrics would be calculated from actual measurements
	}

	statistics := &Statistics{
		Blockchain: blockchainStats,
		Network:    networkInfo,
		Performance: performanceStats,
		LastUpdate: time.Now(),
	}

	return statistics, nil
}

// Helper methods

// calculateTransactionAmount calculates the total output amount of a transaction
func (s *explorerService) calculateTransactionAmount(tx *block.Transaction) uint64 {
	var total uint64
	for _, output := range tx.Outputs {
		total += output.Value
	}
	return total
}

// calculateConfirmations calculates the number of confirmations for a block
func (s *explorerService) calculateConfirmations(blockHeight uint64) uint64 {
	currentHeight := s.dataProvider.GetBlockHeight()
	if blockHeight >= currentHeight {
		return 0
	}
	return currentHeight - blockHeight
}

// determineFinality determines the finality status of a block
func (s *explorerService) determineFinality(blockHeight uint64) string {
	confirmations := s.calculateConfirmations(blockHeight)
	if confirmations >= 100 {
		return "final"
	} else if confirmations >= 6 {
		return "likely_final"
	} else if confirmations >= 1 {
		return "pending"
	}
	return "unconfirmed"
}

// findTransactionBlock finds the block containing a specific transaction
func (s *explorerService) findTransactionBlock(txHash []byte) (*block.Block, error) {
	// This is a simplified implementation
	// In a real system, you'd have a transaction index
	height := s.dataProvider.GetBlockHeight()
	
	for h := uint64(0); h <= height; h++ {
		block, err := s.dataProvider.GetBlockByHeight(h)
		if err != nil || block == nil {
			continue
		}

		for _, tx := range block.Transactions {
			if string(tx.Hash) == string(txHash) {
				return block, nil
			}
		}
	}

	return nil, fmt.Errorf("transaction not found in any block")
}

// extractAddressFromScript extracts an address from a script
func (s *explorerService) extractAddressFromScript(script []byte) string {
	// This is a simplified implementation
	// In a real system, you'd properly parse the script and extract the address
	if len(script) == 0 {
		return "unknown"
	}
	
	// For now, just return a hex representation
	return hex.EncodeToString(script[:min(len(script), 8)])
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
