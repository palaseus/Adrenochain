package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// SecurityConfig holds security configuration for production deployment
type SecurityConfig struct {
	// Cryptographic settings
	KeySize           int
	HashAlgorithm     string
	EncryptionMethod  string
	
	// Network security
	EnableTLS         bool
	TLSConfig         *tls.Config
	RateLimiting      bool
	MaxConnections    int
	
	// Access control
	EnableAuth        bool
	AdminWhitelist    []engine.Address
	APIKeyRequired    bool
	
	// Monitoring and logging
	EnableAuditLog    bool
	LogRetentionDays  int
	AlertThresholds   map[string]int
	
	// Threat detection
	EnableIntrusionDetection bool
	SuspiciousActivityThreshold int
	BlockedIPs        []string
}

// DefaultSecurityConfig returns production-ready security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		KeySize:           256,
		HashAlgorithm:     "SHA-256",
		EncryptionMethod:  "AES-256-GCM",
		EnableTLS:         true,
		RateLimiting:      true,
		MaxConnections:    1000,
		EnableAuth:        true,
		APIKeyRequired:    true,
		EnableAuditLog:    true,
		LogRetentionDays:  90,
		AlertThresholds: map[string]int{
			"failed_logins": 5,
			"suspicious_ips": 10,
			"api_abuse": 100,
		},
		EnableIntrusionDetection: true,
		SuspiciousActivityThreshold: 3,
	}
}

// SecurityManager handles security operations and monitoring
type SecurityManager struct {
	config     *SecurityConfig
	auditLog   []SecurityEvent
	blockedIPs map[string]time.Time
	rateLimit  map[string]*RateLimitTracker
	mu         sync.RWMutex
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time
	EventType   string
	Description string
	IPAddress   string
	UserID      string
	Severity    SecuritySeverity
	Metadata    map[string]interface{}
}

// SecuritySeverity indicates the severity of a security event
type SecuritySeverity int

const (
	SeverityLow SecuritySeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// RateLimitTracker tracks rate limiting for different entities
type RateLimitTracker struct {
	Requests    []time.Time
	WindowSize  time.Duration
	MaxRequests int
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config *SecurityConfig) *SecurityManager {
	return &SecurityManager{
		config:     config,
		auditLog:   make([]SecurityEvent, 0),
		blockedIPs: make(map[string]time.Time),
		rateLimit:  make(map[string]*RateLimitTracker),
	}
}

// GenerateSecureKey generates a cryptographically secure key
func (sm *SecurityManager) GenerateSecureKey() ([]byte, error) {
	key := make([]byte, sm.config.KeySize/8)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure key: %w", err)
	}
	return key, nil
}

// HashData securely hashes data using configured algorithm
func (sm *SecurityManager) HashData(data []byte) ([]byte, error) {
	switch sm.config.HashAlgorithm {
	case "SHA-256":
		hash := sha256.Sum256(data)
		return hash[:], nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", sm.config.HashAlgorithm)
	}
}

// ValidateAPIKey validates an API key
func (sm *SecurityManager) ValidateAPIKey(apiKey string) bool {
	if !sm.config.APIKeyRequired {
		return true
	}
	
	// In production, this would validate against a secure key store
	// For now, we'll use a simple validation
	return len(apiKey) >= 32 && len(apiKey) <= 128
}

// CheckRateLimit checks if a request should be rate limited
func (sm *SecurityManager) CheckRateLimit(identifier string, maxRequests int, window time.Duration) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	tracker, exists := sm.rateLimit[identifier]
	if !exists {
		tracker = &RateLimitTracker{
			Requests:    make([]time.Time, 0),
			WindowSize:  window,
			MaxRequests: maxRequests,
		}
		sm.rateLimit[identifier] = tracker
	}
	
	// Clean old requests outside the window
	now := time.Now()
	cutoff := now.Add(-window)
	
	var validRequests []time.Time
	for _, reqTime := range tracker.Requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	tracker.Requests = validRequests
	
	// Check if limit exceeded
	if len(tracker.Requests) >= maxRequests {
		return false // Rate limited
	}
	
	// Add current request
	tracker.Requests = append(tracker.Requests, now)
	return true // Allowed
}

// LogSecurityEvent logs a security event
func (sm *SecurityManager) LogSecurityEvent(eventType, description, ipAddress, userID string, severity SecuritySeverity, metadata map[string]interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	event := SecurityEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		IPAddress:   ipAddress,
		UserID:      userID,
		Severity:    severity,
		Metadata:    metadata,
	}
	
	sm.auditLog = append(sm.auditLog, event)
	
	// Clean old audit logs
	if sm.config.LogRetentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -sm.config.LogRetentionDays)
		var validLogs []SecurityEvent
		for _, log := range sm.auditLog {
			if log.Timestamp.After(cutoff) {
				validLogs = append(validLogs, log)
			}
		}
		sm.auditLog = validLogs
	}
	
	// Check alert thresholds
	sm.checkAlertThresholds(eventType, ipAddress)
}

// checkAlertThresholds checks if any alert thresholds have been exceeded
func (sm *SecurityManager) checkAlertThresholds(eventType, ipAddress string) {
	// Count recent events of this type
	recentEvents := 0
	cutoff := time.Now().Add(-time.Hour) // Check last hour
	
	for _, event := range sm.auditLog {
		if event.EventType == eventType && event.Timestamp.After(cutoff) {
			recentEvents++
		}
	}
	
	// Check if threshold exceeded
	if threshold, exists := sm.config.AlertThresholds[eventType]; exists && recentEvents >= threshold {
		sm.LogSecurityEvent(
			"ALERT_THRESHOLD_EXCEEDED",
			fmt.Sprintf("Threshold exceeded for %s: %d >= %d", eventType, recentEvents, threshold),
			ipAddress,
			"SYSTEM",
			SeverityHigh,
			map[string]interface{}{
				"event_type": eventType,
				"count":      recentEvents,
				"threshold":  threshold,
			},
		)
	}
}

// BlockIP blocks an IP address temporarily
func (sm *SecurityManager) BlockIP(ipAddress string, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.blockedIPs[ipAddress] = time.Now().Add(duration)
	
	sm.LogSecurityEvent(
		"IP_BLOCKED",
		fmt.Sprintf("IP address blocked for %v", duration),
		ipAddress,
		"SYSTEM",
		SeverityMedium,
		map[string]interface{}{
			"block_duration": duration.String(),
		},
	)
}

// IsIPBlocked checks if an IP address is currently blocked
func (sm *SecurityManager) IsIPBlocked(ipAddress string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if blockTime, exists := sm.blockedIPs[ipAddress]; exists {
		if time.Now().Before(blockTime) {
			return true // Still blocked
		}
		// Block expired, remove it
		delete(sm.blockedIPs, ipAddress)
	}
	return false
}

// GetAuditLog returns the audit log
func (sm *SecurityManager) GetAuditLog() []SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Return a copy to prevent external modification
	logs := make([]SecurityEvent, len(sm.auditLog))
	copy(logs, sm.auditLog)
	return logs
}

// GetSecurityStats returns security statistics
func (sm *SecurityManager) GetSecurityStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_events":     len(sm.auditLog),
		"blocked_ips":      len(sm.blockedIPs),
		"rate_limit_trackers": len(sm.rateLimit),
	}
	
	// Count events by severity
	severityCounts := make(map[SecuritySeverity]int)
	for _, event := range sm.auditLog {
		severityCounts[event.Severity]++
	}
	stats["severity_counts"] = severityCounts
	
	return stats
}

// CleanupExpiredBlocks removes expired IP blocks
func (sm *SecurityManager) CleanupExpiredBlocks() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	now := time.Now()
	for ip, blockTime := range sm.blockedIPs {
		if now.After(blockTime) {
			delete(sm.blockedIPs, ip)
		}
	}
}
