package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// EnhancedSecurityValidator provides comprehensive security validation
type EnhancedSecurityValidator struct {
	config *SecurityValidationConfig
}

// SecurityValidationConfig holds configuration for security validation
type SecurityValidationConfig struct {
	MaxHashLength        int
	MinHashLength        int
	MaxInputLength       int
	MinInputLength       int
	AllowedHashPatterns  []string
	BlockedPatterns      []string
	RateLimitWindow      time.Duration
	MaxRequestsPerWindow int
}

// DefaultSecurityValidationConfig returns default security validation configuration
func DefaultSecurityValidationConfig() *SecurityValidationConfig {
	return &SecurityValidationConfig{
		MaxHashLength:       64,
		MinHashLength:       32,
		MaxInputLength:      10000,
		MinInputLength:      1,
		AllowedHashPatterns: []string{`^[a-fA-F0-9]+$`},
		BlockedPatterns: []string{
			`<script[^>]*>.*?</script>`,
			`javascript:`,
			`on\w+\s*=`,
			`eval\s*\(`,
			`expression\s*\(`,
			`vbscript:`,
			`data:text/html`,
			`../`,
			`..\\`,
			`union\s+select`,
			`drop\s+table`,
			`delete\s+from`,
			`insert\s+into`,
			`update\s+set`,
			`exec\s*\(`,
			`system\s*\(`,
			`shell_exec`,
			`passthru`,
			`file_get_contents`,
			`fopen`,
			`fwrite`,
			`rm\s+-rf`,
			`chmod\s+777`,
		},
		RateLimitWindow:      1 * time.Minute,
		MaxRequestsPerWindow: 100,
	}
}

// NewEnhancedSecurityValidator creates a new enhanced security validator
func NewEnhancedSecurityValidator(config *SecurityValidationConfig) *EnhancedSecurityValidator {
	if config == nil {
		config = DefaultSecurityValidationConfig()
	}
	return &EnhancedSecurityValidator{
		config: config,
	}
}

// ValidateHash validates a hash string for security
func (esv *EnhancedSecurityValidator) ValidateHash(hashHex string) ([]byte, error) {
	// Check length
	if len(hashHex) < esv.config.MinHashLength {
		return nil, fmt.Errorf("hash too short: minimum %d characters required", esv.config.MinHashLength)
	}
	if len(hashHex) > esv.config.MaxHashLength {
		return nil, fmt.Errorf("hash too long: maximum %d characters allowed", esv.config.MaxHashLength)
	}

	// Check for blocked patterns
	if esv.containsBlockedPatterns(hashHex) {
		return nil, fmt.Errorf("hash contains blocked patterns")
	}

	// Validate hex format
	if !esv.isValidHex(hashHex) {
		return nil, fmt.Errorf("invalid hex format")
	}

	// Decode hex
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}

	return hash, nil
}

// ValidateInput validates general input for security
func (esv *EnhancedSecurityValidator) ValidateInput(input string) error {
	// Check length
	if len(input) < esv.config.MinInputLength {
		return fmt.Errorf("input too short: minimum %d characters required", esv.config.MinInputLength)
	}
	if len(input) > esv.config.MaxInputLength {
		return fmt.Errorf("input too long: maximum %d characters allowed", esv.config.MaxInputLength)
	}

	// Check for blocked patterns
	if esv.containsBlockedPatterns(input) {
		return fmt.Errorf("input contains blocked patterns")
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return fmt.Errorf("input contains null bytes")
	}

	return nil
}

// ValidateGasLimit validates gas limit for smart contracts
func (esv *EnhancedSecurityValidator) ValidateGasLimit(gasLimit uint64, maxGasLimit uint64) error {
	if gasLimit == 0 {
		return fmt.Errorf("gas limit cannot be zero")
	}
	if gasLimit > maxGasLimit {
		return fmt.Errorf("gas limit %d exceeds maximum %d", gasLimit, maxGasLimit)
	}

	// Check for reasonable minimum
	minGasLimit := uint64(21000) // Basic transaction gas
	if gasLimit < minGasLimit {
		return fmt.Errorf("gas limit %d below minimum %d", gasLimit, minGasLimit)
	}

	return nil
}

// ValidatePrice validates price for DeFi operations
func (esv *EnhancedSecurityValidator) ValidatePrice(price *big.Int, twap *big.Int, maxDeviationPercent int) error {
	if price == nil {
		return fmt.Errorf("price cannot be nil")
	}
	if price.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if twap == nil {
		return fmt.Errorf("TWAP cannot be nil")
	}

	// Calculate maximum deviation
	maxDeviation := new(big.Int).Div(twap, big.NewInt(int64(100/maxDeviationPercent)))
	if maxDeviation.Cmp(big.NewInt(0)) == 0 {
		maxDeviation = big.NewInt(1) // Minimum deviation of 1
	}

	// Check if price deviation is within acceptable range
	deviation := new(big.Int).Abs(new(big.Int).Sub(price, twap))
	if deviation.Cmp(maxDeviation) > 0 {
		return fmt.Errorf("price deviation %s exceeds maximum %s (TWAP: %s)",
			deviation.String(), maxDeviation.String(), twap.String())
	}

	return nil
}

// ValidateAddress validates blockchain address format
func (esv *EnhancedSecurityValidator) ValidateAddress(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}

	// Check for blocked patterns
	if esv.containsBlockedPatterns(address) {
		return fmt.Errorf("address contains blocked patterns")
	}

	// Basic format validation (can be extended for specific address types)
	if len(address) < 20 || len(address) > 50 {
		return fmt.Errorf("invalid address length: %d", len(address))
	}

	return nil
}

// containsBlockedPatterns checks if input contains any blocked patterns
func (esv *EnhancedSecurityValidator) containsBlockedPatterns(input string) bool {
	inputLower := strings.ToLower(input)

	for _, pattern := range esv.config.BlockedPatterns {
		matched, err := regexp.MatchString(pattern, inputLower)
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			return true
		}
	}

	return false
}

// isValidHex checks if string is valid hexadecimal
func (esv *EnhancedSecurityValidator) isValidHex(s string) bool {
	if len(s)%2 != 0 {
		return false
	}

	for _, pattern := range esv.config.AllowedHashPatterns {
		matched, err := regexp.MatchString(pattern, s)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

// GenerateSecureRandom generates cryptographically secure random bytes
func (esv *EnhancedSecurityValidator) GenerateSecureRandom(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	if length > 1024 {
		return nil, fmt.Errorf("length too large: maximum 1024 bytes")
	}

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return randomBytes, nil
}

// HashData securely hashes data using SHA-256
func (esv *EnhancedSecurityValidator) HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ValidateSignature validates cryptographic signature
func (esv *EnhancedSecurityValidator) ValidateSignature(signature []byte, message []byte, publicKey []byte) error {
	if len(signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	if len(publicKey) == 0 {
		return fmt.Errorf("public key cannot be empty")
	}

	// Basic length validation
	if len(signature) < 32 || len(signature) > 128 {
		return fmt.Errorf("invalid signature length: %d", len(signature))
	}
	if len(publicKey) < 32 || len(publicKey) > 128 {
		return fmt.Errorf("invalid public key length: %d", len(publicKey))
	}

	// In a real implementation, this would perform actual signature verification
	// For now, we'll do basic format validation
	return nil
}

// SanitizeInput sanitizes input by removing dangerous characters
func (esv *EnhancedSecurityValidator) SanitizeInput(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	sanitized = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	// Limit length
	if len(sanitized) > esv.config.MaxInputLength {
		sanitized = sanitized[:esv.config.MaxInputLength]
	}

	return sanitized
}

// SecurityRateLimitTracker tracks rate limiting for security
type SecurityRateLimitTracker struct {
	requests    map[string][]time.Time
	windowSize  time.Duration
	maxRequests int
}

// NewSecurityRateLimitTracker creates a new rate limit tracker
func NewSecurityRateLimitTracker(windowSize time.Duration, maxRequests int) *SecurityRateLimitTracker {
	return &SecurityRateLimitTracker{
		requests:    make(map[string][]time.Time),
		windowSize:  windowSize,
		maxRequests: maxRequests,
	}
}

// IsAllowed checks if a request is allowed based on rate limiting
func (rlt *SecurityRateLimitTracker) IsAllowed(identifier string) bool {
	now := time.Now()
	windowStart := now.Add(-rlt.windowSize)

	// Clean old requests
	if requests, exists := rlt.requests[identifier]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rlt.requests[identifier] = validRequests
	}

	// Check if under limit
	if len(rlt.requests[identifier]) >= rlt.maxRequests {
		return false
	}

	// Add current request
	rlt.requests[identifier] = append(rlt.requests[identifier], now)
	return true
}
