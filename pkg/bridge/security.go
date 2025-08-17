package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// SecurityManager handles bridge security measures
type SecurityManager struct {
	bridge              *Bridge
	rateLimiters        map[string]*RateLimiter
	fraudDetector       *FraudDetector
	emergencyControls   *EmergencyControls
	securityEvents      []*SecurityEvent
	mutex               sync.RWMutex
	maxSecurityEvents   int
}

// RateLimiter implements rate limiting for bridge operations
type RateLimiter struct {
	address        string
	windowSize     time.Duration
	maxRequests    int
	requests       []time.Time
	mutex          sync.RWMutex
}

// FraudDetector detects suspicious bridge activities
type FraudDetector struct {
	suspiciousPatterns map[string]*SuspiciousPattern
	blacklistedAddresses map[string]bool
	anomalyThresholds map[string]float64
	mutex             sync.RWMutex
}

// SuspiciousPattern represents a pattern that might indicate fraud
type SuspiciousPattern struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	RiskScore   float64   `json:"risk_score"`
	Threshold   float64   `json:"threshold"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// EmergencyControls handles emergency bridge operations
type EmergencyControls struct {
	isEmergencyPaused bool
	pausedBy          string
	pausedAt          *time.Time
	pauseReason       string
	emergencyThreshold *big.Int
	mutex             sync.RWMutex
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string            `json:"id"`
	Type        SecurityEventType `json:"type"`
	Severity    SecuritySeverity  `json:"severity"`
	Description string            `json:"description"`
	Address     string            `json:"address,omitempty"`
	Amount      *big.Int          `json:"amount,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Resolved    bool              `json:"resolved"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	SecurityEventTypeRateLimitExceeded SecurityEventType = "rate_limit_exceeded"
	SecurityEventTypeSuspiciousActivity SecurityEventType = "suspicious_activity"
	SecurityEventTypeLargeTransfer      SecurityEventType = "large_transfer"
	SecurityEventTypeEmergencyPause     SecurityEventType = "emergency_pause"
	SecurityEventTypeFraudDetected      SecurityEventType = "fraud_detected"
)

// SecuritySeverity represents the severity of a security event
type SecuritySeverity string

const (
	SecuritySeverityLow    SecuritySeverity = "low"
	SecuritySeverityMedium SecuritySeverity = "medium"
	SecuritySeverityHigh   SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// NewSecurityManager creates a new security manager
func NewSecurityManager(bridge *Bridge) *SecurityManager {
	return &SecurityManager{
		bridge:            bridge,
		rateLimiters:      make(map[string]*RateLimiter),
		fraudDetector:     NewFraudDetector(),
		emergencyControls: NewEmergencyControls(),
		securityEvents:    make([]*SecurityEvent, 0),
		maxSecurityEvents: 1000,
	}
}

// NewRateLimiter creates a new rate limiter for an address
func NewRateLimiter(address string, windowSize time.Duration, maxRequests int) *RateLimiter {
	return &RateLimiter{
		address:     address,
		windowSize:  windowSize,
		maxRequests: maxRequests,
		requests:    make([]time.Time, 0),
	}
}

// NewFraudDetector creates a new fraud detector
func NewFraudDetector() *FraudDetector {
	return &FraudDetector{
		suspiciousPatterns: make(map[string]*SuspiciousPattern),
		blacklistedAddresses: make(map[string]bool),
		anomalyThresholds: make(map[string]float64),
	}
}

// NewEmergencyControls creates new emergency controls
func NewEmergencyControls() *EmergencyControls {
	return &EmergencyControls{
		isEmergencyPaused: false,
		emergencyThreshold: mustParseBigInt("10000000000000000000"), // 10 ETH
	}
}

// CheckRateLimit checks if an address has exceeded rate limits
func (sm *SecurityManager) CheckRateLimit(address string) error {
	sm.mutex.Lock()
	limiter, exists := sm.rateLimiters[address]
	if !exists {
		// Create new rate limiter for address
		limiter = NewRateLimiter(address, 1*time.Hour, 10) // 10 requests per hour
		sm.rateLimiters[address] = limiter
	}
	sm.mutex.Unlock()

	return limiter.CheckLimit()
}

// CheckTransferSecurity performs security checks on a transfer
func (sm *SecurityManager) CheckTransferSecurity(
	sourceAddress string,
	destinationAddress string,
	amount *big.Int,
	assetType AssetType,
) error {
	// Check if bridge is in emergency mode
	if sm.emergencyControls.IsEmergencyPaused() {
		return fmt.Errorf("bridge is in emergency mode")
	}

	// Check rate limits
	if err := sm.CheckRateLimit(sourceAddress); err != nil {
		sm.recordSecurityEvent(SecurityEventTypeRateLimitExceeded, SecuritySeverityMedium, 
			fmt.Sprintf("Rate limit exceeded for address %s", sourceAddress), sourceAddress, amount)
		return err
	}

	// Check for suspicious activity
	if sm.fraudDetector.IsSuspicious(sourceAddress, destinationAddress, amount, assetType) {
		sm.recordSecurityEvent(SecurityEventTypeSuspiciousActivity, SecuritySeverityHigh,
			fmt.Sprintf("Suspicious activity detected for transfer from %s to %s", sourceAddress, destinationAddress),
			sourceAddress, amount)
		return fmt.Errorf("suspicious activity detected")
	}

	// Check for large transfers
	if sm.isLargeTransfer(amount) {
		sm.recordSecurityEvent(SecurityEventTypeLargeTransfer, SecuritySeverityMedium,
			fmt.Sprintf("Large transfer detected: %s", amount.String()), sourceAddress, amount)
	}

	return nil
}

// PauseBridge pauses the bridge in emergency mode
func (sm *SecurityManager) PauseBridge(pausedBy string, reason string) error {
	return sm.emergencyControls.Pause(pausedBy, reason)
}

// ResumeBridge resumes the bridge from emergency mode
func (sm *SecurityManager) ResumeBridge(resumedBy string) error {
	return sm.emergencyControls.Resume(resumedBy)
}

// AddSuspiciousPattern adds a new suspicious pattern
func (sm *SecurityManager) AddSuspiciousPattern(
	pattern string,
	riskScore float64,
	threshold float64,
) error {
	return sm.fraudDetector.AddPattern(pattern, riskScore, threshold)
}

// BlacklistAddress blacklists an address
func (sm *SecurityManager) BlacklistAddress(address string, reason string) error {
	sm.fraudDetector.BlacklistAddress(address)
	
	sm.recordSecurityEvent(SecurityEventTypeFraudDetected, SecuritySeverityHigh,
		fmt.Sprintf("Address %s blacklisted: %s", address, reason), address, nil)
	
	return nil
}

// GetSecurityEvents returns security events
func (sm *SecurityManager) GetSecurityEvents(limit int) []*SecurityEvent {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if limit <= 0 || limit > len(sm.securityEvents) {
		limit = len(sm.securityEvents)
	}

	result := make([]*SecurityEvent, limit)
	copy(result, sm.securityEvents[len(sm.securityEvents)-limit:])

	return result
}

// GetSecurityStats returns security statistics
func (sm *SecurityManager) GetSecurityStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_security_events"] = len(sm.securityEvents)
	stats["emergency_paused"] = sm.emergencyControls.IsEmergencyPaused()
	stats["blacklisted_addresses"] = len(sm.fraudDetector.GetBlacklistedAddresses())
	stats["suspicious_patterns"] = len(sm.fraudDetector.GetSuspiciousPatterns())
	stats["rate_limiters"] = len(sm.rateLimiters)

	return stats
}

// isLargeTransfer checks if a transfer amount is considered large
func (sm *SecurityManager) isLargeTransfer(amount *big.Int) bool {
	threshold := sm.emergencyControls.GetEmergencyThreshold()
	return amount.Cmp(threshold) > 0
}

// recordSecurityEvent records a security event
func (sm *SecurityManager) recordSecurityEvent(
	eventType SecurityEventType,
	severity SecuritySeverity,
	description string,
	address string,
	amount *big.Int,
) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	event := &SecurityEvent{
		ID:          sm.generateEventID(),
		Type:        eventType,
		Severity:    severity,
		Description: description,
		Address:     address,
		Amount:      amount,
		Timestamp:   time.Now(),
		Resolved:    false,
	}

	sm.securityEvents = append(sm.securityEvents, event)

	// Limit the number of stored events
	if len(sm.securityEvents) > sm.maxSecurityEvents {
		sm.securityEvents = sm.securityEvents[1:]
	}

	// Emit event
	sm.bridge.emitEvent("security_event", event)
}

// generateEventID generates a unique event ID
func (sm *SecurityManager) generateEventID() string {
	data := fmt.Sprintf("security_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// Rate Limiter Methods

// CheckLimit checks if the rate limit has been exceeded
func (rl *RateLimiter) CheckLimit() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Remove expired requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range rl.requests {
		if now.Sub(reqTime) <= rl.windowSize {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests = validRequests

	// Check if limit exceeded
	if len(rl.requests) >= rl.maxRequests {
		return fmt.Errorf("rate limit exceeded: %d requests in %v", len(rl.requests), rl.windowSize)
	}

	// Add current request
	rl.requests = append(rl.requests, now)
	return nil
}

// Fraud Detector Methods

// IsSuspicious checks if an activity is suspicious
func (fd *FraudDetector) IsSuspicious(
	sourceAddress string,
	destinationAddress string,
	amount *big.Int,
	assetType AssetType,
) bool {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	// Check if addresses are blacklisted
	if fd.blacklistedAddresses[sourceAddress] || fd.blacklistedAddresses[destinationAddress] {
		return true
	}

	// Check suspicious patterns
	for _, pattern := range fd.suspiciousPatterns {
		if !pattern.IsActive {
			continue
		}

		// Simple pattern matching (could be enhanced with ML)
		if fd.matchesPattern(sourceAddress, destinationAddress, amount, assetType, pattern) {
			return true
		}
	}

	return false
}

// AddPattern adds a new suspicious pattern
func (fd *FraudDetector) AddPattern(pattern string, riskScore float64, threshold float64) error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	patternID := fmt.Sprintf("pattern_%d", time.Now().UnixNano())
	
	fd.suspiciousPatterns[patternID] = &SuspiciousPattern{
		ID:        patternID,
		Pattern:   pattern,
		RiskScore: riskScore,
		Threshold: threshold,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	return nil
}

// BlacklistAddress blacklists an address
func (fd *FraudDetector) BlacklistAddress(address string) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()
	fd.blacklistedAddresses[address] = true
}

// GetBlacklistedAddresses returns blacklisted addresses
func (fd *FraudDetector) GetBlacklistedAddresses() []string {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	var result []string
	for address := range fd.blacklistedAddresses {
		result = append(result, address)
	}

	return result
}

// GetSuspiciousPatterns returns suspicious patterns
func (fd *FraudDetector) GetSuspiciousPatterns() []*SuspiciousPattern {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	var result []*SuspiciousPattern
	for _, pattern := range fd.suspiciousPatterns {
		result = append(result, pattern)
	}

	return result
}

// matchesPattern checks if activity matches a suspicious pattern
func (fd *FraudDetector) matchesPattern(
	sourceAddress string,
	destinationAddress string,
	amount *big.Int,
	assetType AssetType,
	pattern *SuspiciousPattern,
) bool {
	// Simple pattern matching implementation
	// In a real system, this would use more sophisticated ML-based detection
	
	switch pattern.Pattern {
	case "high_frequency":
		// Check for high frequency transfers
		return false // Placeholder
	case "large_amount":
		// Check for unusually large amounts
		return false // Placeholder
	case "suspicious_pair":
		// Check for suspicious address pairs
		return false // Placeholder
	default:
		return false
	}
}

// Emergency Controls Methods

// IsEmergencyPaused checks if bridge is in emergency mode
func (ec *EmergencyControls) IsEmergencyPaused() bool {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	return ec.isEmergencyPaused
}

// Pause pauses the bridge in emergency mode
func (ec *EmergencyControls) Pause(pausedBy string, reason string) error {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()

	if ec.isEmergencyPaused {
		return fmt.Errorf("bridge is already paused")
	}

	ec.isEmergencyPaused = true
	ec.pausedBy = pausedBy
	ec.pauseReason = reason
	now := time.Now()
	ec.pausedAt = &now

	return nil
}

// Resume resumes the bridge from emergency mode
func (ec *EmergencyControls) Resume(resumedBy string) error {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()

	if !ec.isEmergencyPaused {
		return fmt.Errorf("bridge is not paused")
	}

	ec.isEmergencyPaused = false
	ec.pausedBy = ""
	ec.pauseReason = ""
	ec.pausedAt = nil

	return nil
}

// GetEmergencyThreshold returns the emergency threshold
func (ec *EmergencyControls) GetEmergencyThreshold() *big.Int {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	return ec.emergencyThreshold
}

// SetEmergencyThreshold sets the emergency threshold
func (ec *EmergencyControls) SetEmergencyThreshold(threshold *big.Int) {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	ec.emergencyThreshold = threshold
}
