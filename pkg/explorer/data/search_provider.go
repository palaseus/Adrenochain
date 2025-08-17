package data

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/explorer/service"
)

// SimpleSearchProvider implements SearchProvider interface with basic search functionality
type SimpleSearchProvider struct {
	dataProvider service.BlockchainDataProvider
}

// NewSimpleSearchProvider creates a new simple search provider
func NewSimpleSearchProvider(dataProvider service.BlockchainDataProvider) *SimpleSearchProvider {
	return &SimpleSearchProvider{
		dataProvider: dataProvider,
	}
}

// Search performs a global search across the blockchain
func (s *SimpleSearchProvider) Search(query string) (*service.SearchResult, error) {
	// Clean the query
	query = strings.TrimSpace(query)
	if query == "" {
		return &service.SearchResult{
			Query: query,
			Error: "empty search query",
		}, nil
	}

	// Try to determine the type of query and search accordingly

	// Check if it's a block hash (64 hex characters)
	if len(query) == 64 && isHexString(query) {
		hash, err := hex.DecodeString(query)
		if err == nil {
			// Try to find as block
			block, err := s.dataProvider.GetBlock(hash)
			if err == nil && block != nil {
				blockSummary := &service.BlockSummary{
					Hash:          block.CalculateHash(),
					Height:        block.Header.Height,
					Timestamp:     block.Header.Timestamp,
					TxCount:       len(block.Transactions),
					Size:          uint64(len(block.Transactions) * 100),
					Difficulty:    block.Header.Difficulty,
					Confirmations: 0, // Would need to calculate this
				}

				return &service.SearchResult{
					Query: query,
					Type:  "block",
					Block: blockSummary,
				}, nil
			}
		}
	}

	// Check if it's a transaction hash (64 hex characters)
	if len(query) == 64 && isHexString(query) {
		hash, err := hex.DecodeString(query)
		if err == nil {
			// Try to find as transaction
			tx, err := s.dataProvider.GetTransaction(hash)
			if err == nil && tx != nil {
				// Find the block containing this transaction
				block, err := s.findTransactionBlock(hash)
				if err == nil && block != nil {
					txSummary := &service.TransactionSummary{
						Hash:      tx.Hash,
						BlockHash: block.CalculateHash(),
						Height:    block.Header.Height,
						Timestamp: block.Header.Timestamp,
						Inputs:    len(tx.Inputs),
						Outputs:   len(tx.Outputs),
						Amount:    0, // Would need to calculate this
						Fee:       tx.Fee,
						Status:    "confirmed",
					}

					return &service.SearchResult{
						Query:       query,
						Type:        "transaction",
						Transaction: txSummary,
					}, nil
				}
			}
		}
	}

	// Check if it's an address (base58 format, typically 26-35 characters)
	if len(query) >= 26 && len(query) <= 35 && isBase58String(query) {
		// Try to find as address
		balance, err := s.dataProvider.GetAddressBalance(query)
		if err == nil {
			addressSummary := &service.AddressSummary{
				Address:   query,
				Balance:   balance,
				TxCount:   0,           // Would need to calculate this
				FirstSeen: time.Time{}, // Would need to track this
				LastSeen:  time.Time{}, // Would need to track this
			}

			return &service.SearchResult{
				Query:   query,
				Type:    "address",
				Address: addressSummary,
			}, nil
		}
	}

	// Check if it's a block height (numeric)
	if isNumericString(query) {
		height, err := parseUint64(query)
		if err == nil {
			block, err := s.dataProvider.GetBlockByHeight(height)
			if err == nil && block != nil {
				blockSummary := &service.BlockSummary{
					Hash:          block.CalculateHash(),
					Height:        block.Header.Height,
					Timestamp:     block.Header.Timestamp,
					TxCount:       len(block.Transactions),
					Size:          uint64(len(block.Transactions) * 100),
					Difficulty:    block.Header.Difficulty,
					Confirmations: 0, // Would need to calculate this
				}

				return &service.SearchResult{
					Query: query,
					Type:  "block",
					Block: blockSummary,
				}, nil
			}
		}
	}

	// If no exact match found, provide suggestions
	suggestions := s.generateSuggestions(query)

	return &service.SearchResult{
		Query:       query,
		Type:        "unknown",
		Suggestions: suggestions,
	}, nil
}

// SearchBlocks searches for blocks matching the query
func (s *SimpleSearchProvider) SearchBlocks(query string, limit, offset int) ([]*service.BlockSummary, error) {
	// This is a simplified implementation
	// In production, you'd want proper indexing and search algorithms

	var results []*service.BlockSummary
	height := s.dataProvider.GetBlockHeight()

	// Simple search: look for blocks containing the query in their hash
	startHeight := height - uint64(offset)
	endHeight := startHeight - uint64(limit)

	if endHeight < 0 {
		endHeight = 0
	}

	for h := startHeight; h >= endHeight && h >= 0; h-- {
		block, err := s.dataProvider.GetBlockByHeight(h)
		if err != nil || block == nil {
			continue
		}

		// Check if block matches query
		if s.blockMatchesQuery(block, query) {
			blockSummary := &service.BlockSummary{
				Hash:          block.CalculateHash(),
				Height:        block.Header.Height,
				Timestamp:     block.Header.Timestamp,
				TxCount:       len(block.Transactions),
				Size:          uint64(len(block.Transactions) * 100),
				Difficulty:    block.Header.Difficulty,
				Confirmations: 0, // Would need to calculate this
			}

			results = append(results, blockSummary)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// SearchTransactions searches for transactions matching the query
func (s *SimpleSearchProvider) SearchTransactions(query string, limit, offset int) ([]*service.TransactionSummary, error) {
	// This is a simplified implementation
	// In production, you'd want proper indexing and search algorithms

	var results []*service.TransactionSummary
	height := s.dataProvider.GetBlockHeight()

	// Simple search: look for transactions containing the query
	startHeight := height - uint64(offset/5) // Assume 5 transactions per block
	endHeight := startHeight - uint64(limit/5)

	if endHeight < 0 {
		endHeight = 0
	}

	for h := startHeight; h >= endHeight && h >= 0; h-- {
		block, err := s.dataProvider.GetBlockByHeight(h)
		if err != nil || block == nil {
			continue
		}

		for _, tx := range block.Transactions {
			if s.transactionMatchesQuery(tx, query) {
				txSummary := &service.TransactionSummary{
					Hash:      tx.Hash,
					BlockHash: block.CalculateHash(),
					Height:    block.Header.Height,
					Timestamp: block.Header.Timestamp,
					Inputs:    len(tx.Inputs),
					Outputs:   len(tx.Outputs),
					Amount:    0, // Would need to calculate this
					Fee:       tx.Fee,
					Status:    "confirmed",
				}

				results = append(results, txSummary)
				if len(results) >= limit {
					break
				}
			}
		}

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// SearchAddresses searches for addresses matching the query
func (s *SimpleSearchProvider) SearchAddresses(query string, limit, offset int) ([]*service.AddressSummary, error) {
	// This is a simplified implementation
	// In production, you'd want proper indexing and search algorithms

	// For now, return empty results
	// This would need proper address indexing
	return []*service.AddressSummary{}, nil
}

// Helper methods

// findTransactionBlock finds the block containing a specific transaction
func (s *SimpleSearchProvider) findTransactionBlock(txHash []byte) (*block.Block, error) {
	// This is inefficient and should be replaced with proper transaction indexing
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

	return nil, fmt.Errorf("transaction not found")
}

// blockMatchesQuery checks if a block matches the search query
func (s *SimpleSearchProvider) blockMatchesQuery(block *block.Block, query string) bool {
	// Empty query matches all blocks
	if query == "" {
		return true
	}
	// Check if query appears in block hash
	hashHex := hex.EncodeToString(block.CalculateHash())
	if strings.Contains(strings.ToLower(hashHex), strings.ToLower(query)) {
		return true
	}
	// Check if query matches block height
	heightStr := fmt.Sprintf("%d", block.Header.Height)
	if strings.Contains(heightStr, query) {
		return true
	}
	return false
}

// transactionMatchesQuery checks if a transaction matches the search query
func (s *SimpleSearchProvider) transactionMatchesQuery(tx *block.Transaction, query string) bool {
	// Empty query matches all transactions
	if query == "" {
		return true
	}
	// Check if query appears in transaction hash (hex representation)
	hashHex := hex.EncodeToString(tx.Hash)
	if strings.Contains(strings.ToLower(hashHex), strings.ToLower(query)) {
		return true
	}
	// Check if query appears in transaction hash (string representation)
	hashStr := string(tx.Hash)
	if strings.Contains(strings.ToLower(hashStr), strings.ToLower(query)) {
		return true
	}
	return false
}

// generateSuggestions generates search suggestions
func (s *SimpleSearchProvider) generateSuggestions(query string) []string {
	suggestions := []string{
		"Try searching for:",
		"- Block hash (64 hex characters)",
		"- Transaction hash (64 hex characters)",
		"- Address (base58 format)",
		"- Block height (number)",
	}

	// Add the query to suggestions if it's not empty
	if query != "" {
		suggestions = append(suggestions, query)
	}

	return suggestions
}

// Utility functions

// isHexString checks if a string contains only hex characters
func isHexString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// isBase58String checks if a string contains only base58 characters
func isBase58String(s string) bool {
	if s == "" {
		return false
	}
	base58Chars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for _, r := range s {
		if !strings.ContainsRune(base58Chars, r) {
			return false
		}
	}
	return true
}

// isNumericString checks if a string contains only numeric characters
func isNumericString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parseUint64 parses a string to uint64
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be parsed as uint64")
	}
	var result uint64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("non-numeric character in string: %c", r)
		}
		result = result*10 + uint64(r-'0')
	}
	return result, nil
}
