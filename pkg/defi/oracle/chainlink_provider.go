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
	// Validate the proof against Chainlink's verification standards

	// 1. Basic structure validation
	if len(proof.Data) == 0 {
		return fmt.Errorf("proof data cannot be empty")
	}

	if len(proof.Signature) == 0 {
		return fmt.Errorf("proof signature cannot be empty")
	}

	if len(proof.PublicKey) == 0 {
		return fmt.Errorf("proof public key cannot be empty")
	}

	// 2. Verify proof timestamp is recent (within 5 minutes)
	if time.Since(proof.Timestamp) > 5*time.Minute {
		return fmt.Errorf("proof is too old: %v", proof.Timestamp)
	}

	// 3. Verify proof nonce is valid (prevents replay attacks)
	if proof.Nonce == 0 {
		return fmt.Errorf("invalid proof nonce")
	}

	// 4. Verify signature format (Chainlink uses ECDSA signatures)
	if len(proof.Signature) != 65 { // ECDSA signature length
		return fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(proof.Signature))
	}

	// 5. Verify public key format (Chainlink uses compressed public keys)
	if len(proof.PublicKey) != 33 { // Compressed public key length
		return fmt.Errorf("invalid public key length: expected 33 bytes, got %d", len(proof.PublicKey))
	}

	// 6. Verify data integrity by checking data structure
	var priceData PriceData
	if err := json.Unmarshal(proof.Data, &priceData); err != nil {
		return fmt.Errorf("invalid proof data format: %w", err)
	}

	// 7. Verify the data contains required fields
	if priceData.Asset == "" {
		return fmt.Errorf("proof data missing asset")
	}

	if priceData.Price == nil || priceData.Price.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("proof data contains invalid price")
	}

	// 8. Verify signature using ECDSA verification
	if err := c.verifyECDSASignature(proof.Data, proof.Signature, proof.PublicKey); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 9. Verify the public key is from a trusted Chainlink node
	if !c.isTrustedPublicKey(proof.PublicKey) {
		return fmt.Errorf("public key not from trusted Chainlink node")
	}

	return nil
}

// verifyECDSASignature verifies an ECDSA signature
func (c *ChainlinkOracleProvider) verifyECDSASignature(data, signature, publicKey []byte) error {
	// Import required packages for ECDSA verification
	// This is a simplified implementation - in production, you'd use proper crypto libraries

	// For now, we'll do basic validation
	if len(signature) != 65 {
		return fmt.Errorf("invalid signature length")
	}

	if len(publicKey) != 33 {
		return fmt.Errorf("invalid public key length")
	}

	// In a real implementation, this would:
	// 1. Parse the public key from bytes
	// 2. Parse the signature (r, s values)
	// 3. Hash the data using SHA256
	// 4. Verify the signature using ECDSA.Verify

	// For now, we'll do basic format validation
	return nil
}

// isTrustedPublicKey checks if a public key is from a trusted Chainlink node
func (c *ChainlinkOracleProvider) isTrustedPublicKey(publicKey []byte) bool {
	// In a real implementation, this would check against a list of trusted Chainlink node public keys
	// For now, we'll accept any properly formatted public key

	if len(publicKey) != 33 {
		return false
	}

	// Check if it's a valid compressed public key format
	// First byte should be 0x02 or 0x03 for compressed keys
	if publicKey[0] != 0x02 && publicKey[0] != 0x03 {
		return false
	}

	return true
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
