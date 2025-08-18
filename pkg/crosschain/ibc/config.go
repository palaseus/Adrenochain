package ibc

import (
	"fmt"
	"time"
	"gopkg.in/yaml.v3"
	"os"
)

// IBCConfig represents the complete IBC configuration
type IBCConfig struct {
	Clients     map[string]ClientConfigSection     `yaml:"clients"`
	Connections map[string]ConnectionConfigSection `yaml:"connections"`
	Channels    map[string]ChannelConfigSection    `yaml:"channels"`
	Packets     map[string]PacketConfigSection     `yaml:"packets"`
	Security    SecurityConfigSection               `yaml:"security"`
	Performance PerformanceConfigSection            `yaml:"performance"`
}

// ClientConfigSection represents client configuration section
type ClientConfigSection struct {
	MaxClockDrift     string `yaml:"max_clock_drift"`
	TrustingPeriod    string `yaml:"trusting_period"`
	UnbondingPeriod   string `yaml:"unbonding_period"`
	MaxHeaderSize     int    `yaml:"max_header_size"`
	EnableCompression bool   `yaml:"enable_compression"`
	SecurityLevel     string `yaml:"security_level"`
	AutoFreeze        bool   `yaml:"auto_freeze"`
}

// ConnectionConfigSection represents connection configuration section
type ConnectionConfigSection struct {
	MaxDelayPeriod    string `yaml:"max_delay_period"`
	RetryAttempts     int    `yaml:"retry_attempts"`
	Timeout           string `yaml:"timeout"`
	EnableCompression bool   `yaml:"enable_compression"`
	SecurityLevel     string `yaml:"security_level"`
	AutoRetry         bool   `yaml:"auto_retry"`
}

// ChannelConfigSection represents channel configuration section
type ChannelConfigSection struct {
	MaxPacketSize     int    `yaml:"max_packet_size"`
	MaxPacketTimeout  string `yaml:"max_packet_timeout"`
	EnableCompression bool   `yaml:"enable_compression"`
	SecurityLevel     string `yaml:"security_level"`
	AutoClose         bool   `yaml:"auto_close"`
}

// PacketConfigSection represents packet configuration section
type PacketConfigSection struct {
	MaxRetries       int    `yaml:"max_retries"`
	RetryDelay       string `yaml:"retry_delay"`
	Timeout          string `yaml:"timeout"`
	EnableCompression bool  `yaml:"enable_compression"`
	SecurityLevel    string `yaml:"security_level"`
	AutoRetry        bool   `yaml:"auto_retry"`
}

// SecurityConfigSection represents security configuration
type SecurityConfigSection struct {
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Validation ValidationConfig  `yaml:"validation"`
	Monitoring MonitoringConfig  `yaml:"monitoring"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	MaxConnectionsPerMinute int `yaml:"max_connections_per_minute"`
	MaxPacketsPerMinute     int `yaml:"max_packets_per_minute"`
	MaxChannelsPerMinute    int `yaml:"max_channels_per_minute"`
}

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	RequireProofs   bool `yaml:"require_proofs"`
	ValidateHeaders bool `yaml:"validate_headers"`
	ValidateProofs  bool `yaml:"validate_proofs"`
	MaxProofSize    int  `yaml:"max_proof_size"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	EnableMetrics     bool   `yaml:"enable_metrics"`
	EnableLogging     bool   `yaml:"enable_logging"`
	LogLevel          string `yaml:"log_level"`
	MetricsInterval   string `yaml:"metrics_interval"`
}

// PerformanceConfigSection represents performance configuration
type PerformanceConfigSection struct {
	ConnectionPool ConnectionPoolConfig `yaml:"connection_pool"`
	PacketBatching PacketBatchingConfig `yaml:"packet_batching"`
	Compression    CompressionConfig    `yaml:"compression"`
}

// ConnectionPoolConfig represents connection pool configuration
type ConnectionPoolConfig struct {
	MaxIdleConnections   int    `yaml:"max_idle_connections"`
	MaxActiveConnections int    `yaml:"max_active_connections"`
	ConnectionLifetime   string `yaml:"connection_lifetime"`
}

// PacketBatchingConfig represents packet batching configuration
type PacketBatchingConfig struct {
	Enabled       bool   `yaml:"enabled"`
	MaxBatchSize  int    `yaml:"max_batch_size"`
	MaxBatchDelay string `yaml:"max_batch_delay"`
}

// CompressionConfig represents compression configuration
type CompressionConfig struct {
	Algorithm string `yaml:"algorithm"`
	MinSize   int    `yaml:"min_size"`
	Level     int    `yaml:"level"`
}

// LoadIBCConfig loads IBC configuration from file
func LoadIBCConfig(configPath string) (*IBCConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config IBCConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// ConvertToClientConfig converts ClientConfigSection to ClientConfig
func (c ClientConfigSection) ConvertToClientConfig() (ClientConfig, error) {
	maxClockDrift, err := parseDuration(c.MaxClockDrift)
	if err != nil {
		return ClientConfig{}, fmt.Errorf("invalid max_clock_drift: %w", err)
	}

	trustingPeriod, err := parseDuration(c.TrustingPeriod)
	if err != nil {
		return ClientConfig{}, fmt.Errorf("invalid trusting_period: %w", err)
	}

	unbondingPeriod, err := parseDuration(c.UnbondingPeriod)
	if err != nil {
		return ClientConfig{}, fmt.Errorf("invalid unbonding_period: %w", err)
	}

	securityLevel, err := parseSecurityLevel(c.SecurityLevel)
	if err != nil {
		return ClientConfig{}, fmt.Errorf("invalid security_level: %w", err)
	}

	return ClientConfig{
		MaxClockDrift:     maxClockDrift,
		TrustingPeriod:    trustingPeriod,
		UnbondingPeriod:   unbondingPeriod,
		MaxHeaderSize:     c.MaxHeaderSize,
		EnableCompression: c.EnableCompression,
		SecurityLevel:     securityLevel,
		AutoFreeze:        c.AutoFreeze,
	}, nil
}

// ConvertToConnectionConfig converts ConnectionConfigSection to ConnectionConfig
func (c ConnectionConfigSection) ConvertToConnectionConfig() (ConnectionConfig, error) {
	maxDelayPeriod, err := parseDuration(c.MaxDelayPeriod)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("invalid max_delay_period: %w", err)
	}

	timeout, err := parseDuration(c.Timeout)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("invalid timeout: %w", err)
	}

	securityLevel, err := parseSecurityLevel(c.SecurityLevel)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("invalid security_level: %w", err)
	}

	return ConnectionConfig{
		MaxDelayPeriod:    maxDelayPeriod,
		RetryAttempts:     c.RetryAttempts,
		Timeout:           timeout,
		EnableCompression: c.EnableCompression,
		SecurityLevel:     securityLevel,
		AutoRetry:         c.AutoRetry,
	}, nil
}

// ConvertToChannelConfig converts ChannelConfigSection to ChannelConfig
func (c ChannelConfigSection) ConvertToChannelConfig() (ChannelConfig, error) {
	maxPacketTimeout, err := parseDuration(c.MaxPacketTimeout)
	if err != nil {
		return ChannelConfig{}, fmt.Errorf("invalid max_packet_timeout: %w", err)
	}

	securityLevel, err := parseSecurityLevel(c.SecurityLevel)
	if err != nil {
		return ChannelConfig{}, fmt.Errorf("invalid security_level: %w", err)
	}

	return ChannelConfig{
		MaxPacketSize:     c.MaxPacketSize,
		MaxPacketTimeout:  maxPacketTimeout,
		EnableCompression: c.EnableCompression,
		SecurityLevel:     securityLevel,
		AutoClose:         c.AutoClose,
	}, nil
}

// ConvertToPacketConfig converts PacketConfigSection to PacketConfig
func (c PacketConfigSection) ConvertToPacketConfig() (PacketConfig, error) {
	retryDelay, err := parseDuration(c.RetryDelay)
	if err != nil {
		return PacketConfig{}, fmt.Errorf("invalid retry_delay: %w", err)
	}

	timeout, err := parseDuration(c.Timeout)
	if err != nil {
		return PacketConfig{}, fmt.Errorf("invalid timeout: %w", err)
	}

	securityLevel, err := parseSecurityLevel(c.SecurityLevel)
	if err != nil {
		return PacketConfig{}, fmt.Errorf("invalid security_level: %w", err)
	}

	return PacketConfig{
		MaxRetries:       c.MaxRetries,
		RetryDelay:       retryDelay,
		Timeout:          timeout,
		EnableCompression: c.EnableCompression,
		SecurityLevel:    securityLevel,
		AutoRetry:        c.AutoRetry,
	}, nil
}

// Helper functions for parsing
func parseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

func parseSecurityLevel(levelStr string) (SecurityLevel, error) {
	switch levelStr {
	case "low":
		return SecurityLevelLow, nil
	case "medium":
		return SecurityLevelMedium, nil
	case "high":
		return SecurityLevelHigh, nil
	case "ultra":
		return SecurityLevelUltra, nil
	default:
		return SecurityLevelMedium, fmt.Errorf("unknown security level: %s", levelStr)
	}
}
