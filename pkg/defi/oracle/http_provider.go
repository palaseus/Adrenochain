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

// HTTPOracleProvider implements OracleProvider for generic HTTP API price feeds
type HTTPOracleProvider struct {
	name        string
	description string
	url         string
	active      bool
	reliability float64

	// HTTP specific configuration
	apiKey     string
	timeout    time.Duration
	httpClient *http.Client

	// Statistics
	requests  uint64
	successes uint64
	failures  uint64
}

// HTTPResponse represents the response from HTTP API
type HTTPResponse struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Timestamp int64  `json:"timestamp"`
}

// NewHTTPOracleProvider creates a new HTTP oracle provider
func NewHTTPOracleProvider(name, description, url, apiKey string) *HTTPOracleProvider {
	return &HTTPOracleProvider{
		name:        name,
		description: description,
		url:         url,
		active:      true,
		reliability: 0.95, // Good reliability for HTTP APIs
		apiKey:      apiKey,
		timeout:     30 * time.Second,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPrice retrieves the current price for a given asset from HTTP API
func (h *HTTPOracleProvider) GetPrice(ctx context.Context, asset string) (*PriceData, error) {
	h.requests++

	// Create request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Build request URL
	reqURL := fmt.Sprintf("%s/price/%s", h.url, asset)
	req, err := http.NewRequestWithContext(reqCtx, "GET", reqURL, nil)
	if err != nil {
		h.failures++
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.failures++
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		h.failures++
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.failures++
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var httpResp HTTPResponse
	if err := json.Unmarshal(body, &httpResp); err != nil {
		h.failures++
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert price to big.Int
	price, ok := new(big.Int).SetString(httpResp.Price, 10)
	if !ok {
		h.failures++
		return nil, fmt.Errorf("invalid price format: %s", httpResp.Price)
	}

	// Create price data
	priceData := &PriceData{
		Asset:       asset,
		Price:       price,
		Timestamp:   time.Unix(httpResp.Timestamp, 0),
		BlockNumber: 0, // Would come from blockchain context
		Provider:    h.name,
		Confidence:  85, // Good confidence for HTTP APIs
		Source:      "http_api",
	}

	h.successes++
	return priceData, nil
}

// ValidateProof validates cryptographic proof for oracle data
func (h *HTTPOracleProvider) ValidateProof(ctx context.Context, proof *OracleProof) error {
	if proof == nil {
		return ErrInvalidProof
	}

	// HTTP APIs typically don't provide cryptographic proofs
	// For now, we'll do basic validation
	if len(proof.Data) == 0 {
		return ErrInvalidProof
	}

	// HTTP APIs may not have signatures or public keys
	// This is a limitation of HTTP-based oracles

	return nil
}

// UpdatePrice updates the price for a given asset
func (h *HTTPOracleProvider) UpdatePrice(ctx context.Context, asset string, price *big.Int, proof *OracleProof) error {
	// HTTP APIs are typically read-only
	// This method is not applicable for HTTP APIs
	return fmt.Errorf("HTTP oracle is read-only, cannot update prices")
}

// GetProviderInfo returns information about the oracle provider
func (h *HTTPOracleProvider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:        h.name,
		Description: h.description,
		URL:         h.url,
		PublicKey:   []byte("http_api_public_key"), // Would be actual public key if available
		Active:      h.active,
		LastUpdate:  time.Now(),
		Reliability: h.reliability,
	}
}

// SetTimeout sets the request timeout
func (h *HTTPOracleProvider) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
	h.httpClient.Timeout = timeout
}

// SetActive sets whether the provider is active
func (h *HTTPOracleProvider) SetActive(active bool) {
	h.active = active
}

// GetStats returns provider statistics
func (h *HTTPOracleProvider) GetStats() (uint64, uint64, uint64) {
	return h.requests, h.successes, h.failures
}

// IsHealthy checks if the provider is healthy
func (h *HTTPOracleProvider) IsHealthy() bool {
	// Check if we can make a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get a simple price (e.g., ETH/USD)
	_, err := h.GetPrice(ctx, "ETH/USD")
	return err == nil
}
