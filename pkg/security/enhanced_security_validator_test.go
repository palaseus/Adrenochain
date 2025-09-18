package security

import (
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestEnhancedSecurityValidator_ValidateHash(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name    string
		hashHex string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid hash",
			hashHex: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantErr: false,
		},
		{
			name:    "too short",
			hashHex: "abc",
			wantErr: true,
			errMsg:  "hash too short",
		},
		{
			name:    "too long",
			hashHex: strings.Repeat("a", 100),
			wantErr: true,
			errMsg:  "hash too long",
		},
		{
			name:    "invalid hex",
			hashHex: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr: true,
			errMsg:  "invalid hex format",
		},
		{
			name:    "script injection",
			hashHex: "<script>alert('xss')</script>",
			wantErr: true,
			errMsg:  "too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateHash(tt.hashHex)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateHash() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateHash() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateHash() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEnhancedSecurityValidator_ValidateInput(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid input",
			input:   "hello world",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
			errMsg:  "too short",
		},
		{
			name:    "too long input",
			input:   strings.Repeat("a", 20000),
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "script injection",
			input:   "<script>alert('xss')</script>",
			wantErr: true,
			errMsg:  "blocked patterns",
		},
		{
			name:    "sql injection",
			input:   "'; DROP TABLE users; --",
			wantErr: true,
			errMsg:  "blocked patterns",
		},
		{
			name:    "null bytes",
			input:   "hello\x00world",
			wantErr: true,
			errMsg:  "null bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateInput() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateInput() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateInput() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEnhancedSecurityValidator_ValidateGasLimit(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name        string
		gasLimit    uint64
		maxGasLimit uint64
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid gas limit",
			gasLimit:    100000,
			maxGasLimit: 1000000,
			wantErr:     false,
		},
		{
			name:        "zero gas limit",
			gasLimit:    0,
			maxGasLimit: 1000000,
			wantErr:     true,
			errMsg:      "cannot be zero",
		},
		{
			name:        "exceeds maximum",
			gasLimit:    2000000,
			maxGasLimit: 1000000,
			wantErr:     true,
			errMsg:      "exceeds maximum",
		},
		{
			name:        "below minimum",
			gasLimit:    10000,
			maxGasLimit: 1000000,
			wantErr:     true,
			errMsg:      "below minimum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateGasLimit(tt.gasLimit, tt.maxGasLimit)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateGasLimit() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateGasLimit() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateGasLimit() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEnhancedSecurityValidator_ValidatePrice(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name                string
		price               *big.Int
		twap                *big.Int
		maxDeviationPercent int
		wantErr             bool
		errMsg              string
	}{
		{
			name:                "valid price",
			price:               big.NewInt(100),
			twap:                big.NewInt(100),
			maxDeviationPercent: 10,
			wantErr:             false,
		},
		{
			name:                "nil price",
			price:               nil,
			twap:                big.NewInt(100),
			maxDeviationPercent: 10,
			wantErr:             true,
			errMsg:              "cannot be nil",
		},
		{
			name:                "zero price",
			price:               big.NewInt(0),
			twap:                big.NewInt(100),
			maxDeviationPercent: 10,
			wantErr:             true,
			errMsg:              "must be positive",
		},
		{
			name:                "price deviation too high",
			price:               big.NewInt(200),
			twap:                big.NewInt(100),
			maxDeviationPercent: 10,
			wantErr:             true,
			errMsg:              "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePrice(tt.price, tt.twap, tt.maxDeviationPercent)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePrice() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePrice() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePrice() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEnhancedSecurityValidator_GenerateSecureRandom(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name    string
		length  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid length",
			length:  32,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name:    "negative length",
			length:  -1,
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name:    "too large",
			length:  2000,
			wantErr: true,
			errMsg:  "too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.GenerateSecureRandom(tt.length)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateSecureRandom() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GenerateSecureRandom() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateSecureRandom() unexpected error = %v", err)
				}
				if len(result) != tt.length {
					t.Errorf("GenerateSecureRandom() length = %d, want %d", len(result), tt.length)
				}
			}
		})
	}
}

func TestSecurityRateLimitTracker(t *testing.T) {
	tracker := NewSecurityRateLimitTracker(1*time.Minute, 3)

	// Test normal operation
	if !tracker.IsAllowed("user1") {
		t.Error("First request should be allowed")
	}
	if !tracker.IsAllowed("user1") {
		t.Error("Second request should be allowed")
	}
	if !tracker.IsAllowed("user1") {
		t.Error("Third request should be allowed")
	}
	if tracker.IsAllowed("user1") {
		t.Error("Fourth request should be blocked")
	}

	// Test different users
	if !tracker.IsAllowed("user2") {
		t.Error("Different user should be allowed")
	}
}

func TestEnhancedSecurityValidator_SanitizeInput(t *testing.T) {
	validator := NewEnhancedSecurityValidator(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal input",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "null bytes",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "control characters",
			input:    "hello\x01\x02world",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkValidateHash(b *testing.B) {
	validator := NewEnhancedSecurityValidator(nil)
	hashHex := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateHash(hashHex)
	}
}

func BenchmarkValidateInput(b *testing.B) {
	validator := NewEnhancedSecurityValidator(nil)
	input := "hello world this is a test input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateInput(input)
	}
}
