package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

// ChainlinkOracleProvider implements OracleProvider for Chainlink price feeds
type ChainlinkOracleProvider struct {
	name        string
	description string
	url         string
	active      bool
	reliability float64

	// Chainlink specific configuration
	apiKey     string
	timeout    time.Duration
	httpClient *http.Client

	// Statistics
	requests  uint64
	successes uint64
	failures  uint64
}

// ChainlinkResponse represents the response from Chainlink API
type ChainlinkResponse struct {
	Data struct {
		Base     string `json:"base"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	} `json:"data"`
}

// NewChainlinkOracleProvider creates a new Chainlink oracle provider
func NewChainlinkOracleProvider(name, description, url, apiKey string) *ChainlinkOracleProvider {
	return &ChainlinkOracleProvider{
		name:        name,
		description: description,
		url:         url,
		active:      true,
		reliability: 0.99, // High reliability for Chainlink
		apiKey:      apiKey,
		timeout:     30 * time.Second,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPrice retrieves the current price for a given asset from Chainlink
func (c *ChainlinkOracleProvider) GetPrice(ctx context.Context, asset string) (*PriceData, error) {
	c.requests++

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build request URL
	reqURL := fmt.Sprintf("%s/v1/price/%s", c.url, asset)
	req, err := http.NewRequestWithContext(reqCtx, "GET", reqURL, nil)
	if err != nil {
		c.failures++
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.failures++
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.failures++
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.failures++
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var chainlinkResp ChainlinkResponse
	if err := json.Unmarshal(body, &chainlinkResp); err != nil {
		c.failures++
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert amount to big.Int
	amount, ok := new(big.Int).SetString(chainlinkResp.Data.Amount, 10)
	if !ok {
		c.failures++
		return nil, fmt.Errorf("invalid amount format: %s", chainlinkResp.Data.Amount)
	}

	// Create price data
	priceData := &PriceData{
		Asset:       asset,
		Price:       amount,
		Timestamp:   time.Now(),
		BlockNumber: 0, // Would come from blockchain context
		Provider:    c.name,
		Confidence:  95, // High confidence for Chainlink
		Source:      "chainlink",
	}

	c.successes++
	return priceData, nil
}

// ValidateProof validates cryptographic proof for oracle data
func (c *ChainlinkOracleProvider) ValidateProof(ctx context.Context, proof *OracleProof) error {
	if proof == nil {
		return ErrInvalidProof
	}

	// Chainlink provides cryptographic proofs
	// In a real implementation, this would validate the proof against Chainlink's verification
	// For now, we'll do basic validation
	if len(proof.Data) == 0 {
		return ErrInvalidProof
	}

	if len(proof.Signature) == 0 {
		return ErrInvalidProof
	}

	if len(proof.PublicKey) == 0 {
		return ErrInvalidProof
	}

	// Implement actual cryptographic proof validation
	// This involves verifying the signature against the public key
	// and ensuring the proof is recent and valid

	// 1. Verify proof timestamp is recent (within 5 minutes)
	if time.Since(proof.Timestamp) > 5*time.Minute {
		return fmt.Errorf("proof is too old: %v", proof.Timestamp)
	}

	// 2. Verify proof nonce is valid (prevents replay attacks)
	if proof.Nonce == 0 {
		return fmt.Errorf("invalid proof nonce")
	}

	// 3. Verify signature using Chainlink's verification method
	// In a real implementation, this would use Chainlink's specific verification
	// For now, we'll do basic validation
	if len(proof.Signature) < 64 { // Minimum signature length
		return fmt.Errorf("invalid signature length")
	}

	// 4. Verify public key format
	if len(proof.PublicKey) < 32 { // Minimum public key length
		return fmt.Errorf("invalid public key length")
	}

	// 5. Verify data integrity
	if len(proof.Data) == 0 {
		return fmt.Errorf("proof data is empty")
	}

	// 6. In a real implementation, we would:
	// - Verify the signature against the public key
	// - Check that the public key is from a trusted Chainlink node
	// - Validate the proof data structure
	// - Ensure the proof hasn't been tampered with

	return nil
}

// UpdatePrice updates the price for a given asset
func (c *ChainlinkOracleProvider) UpdatePrice(ctx context.Context, asset string, price *big.Int, proof *OracleProof) error {
	// Chainlink is a read-only oracle - prices are updated by Chainlink nodes
	// This method is not applicable for Chainlink
	return fmt.Errorf("Chainlink oracle is read-only, cannot update prices")
}

// GetProviderInfo returns information about the oracle provider
func (c *ChainlinkOracleProvider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:        c.name,
		Description: c.description,
		URL:         c.url,
		PublicKey:   []byte("chainlink_public_key"), // Would be actual public key
		Active:      c.active,
		LastUpdate:  time.Now(),
		Reliability: c.reliability,
	}
}

// SetTimeout sets the request timeout
func (c *ChainlinkOracleProvider) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetActive sets whether the provider is active
func (c *ChainlinkOracleProvider) SetActive(active bool) {
	c.active = active
}

// GetStats returns provider statistics
func (c *ChainlinkOracleProvider) GetStats() (uint64, uint64, uint64) {
	return c.requests, c.successes, c.failures
}

// IsHealthy checks if the provider is healthy
func (c *ChainlinkOracleProvider) IsHealthy() bool {
	// Check if we can make a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get a simple price (e.g., ETH/USD)
	_, err := c.GetPrice(ctx, "ETH/USD")
	return err == nil
}
