package net

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	MaxMessageSize        int           // Maximum message size in bytes
	MaxPeerConnections    int           // Maximum number of peer connections
	RateLimitPerSecond    int           // Rate limit per second per peer
	TimestampTolerance    time.Duration // Allowed timestamp deviation
	MaxConcurrentRequests int           // Maximum concurrent requests per peer
	BlacklistThreshold    int           // Number of violations before blacklisting
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxMessageSize:        1024 * 1024, // 1MB
		MaxPeerConnections:    100,
		RateLimitPerSecond:    10,
		TimestampTolerance:    30 * time.Second,
		MaxConcurrentRequests: 5,
		BlacklistThreshold:    3,
	}
}

// MessageValidator validates network messages for security
type MessageValidator struct {
	config     *SecurityConfig
	peerStats  map[string]*PeerStats
	blacklist  map[string]bool
	rateLimits map[string]*RateLimiter
}

// PeerStats tracks peer behavior for security analysis
type PeerStats struct {
	MessageCount    int
	ViolationCount  int
	LastSeen        time.Time
	RequestCount    int
	LastRequestTime time.Time
}

// RateLimiter implements rate limiting for peers
type RateLimiter struct {
	requests    []time.Time
	lastReset   time.Time
	windowSize  time.Duration
	maxRequests int
}

// NewMessageValidator creates a new message validator
func NewMessageValidator(config *SecurityConfig) *MessageValidator {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return &MessageValidator{
		config:     config,
		peerStats:  make(map[string]*PeerStats),
		blacklist:  make(map[string]bool),
		rateLimits: make(map[string]*RateLimiter),
	}
}

// ValidateMessage validates a message from a peer
func (mv *MessageValidator) ValidateMessage(msg *Message, peerID string) error {
	// Check if peer is blacklisted
	if mv.IsPeerBlacklisted(peerID) {
		return fmt.Errorf("peer %s is blacklisted", peerID)
	}

	// Update peer stats first (count all messages, valid or invalid)
	mv.updatePeerStats(peerID)

	// Validate message size
	if err := mv.validateMessageSize(msg); err != nil {
		mv.recordViolation(peerID, "size_exceeded")
		return err
	}

	// Validate timestamp
	if err := mv.validateTimestamp(msg); err != nil {
		mv.recordViolation(peerID, "invalid_timestamp")
		return err
	}

	// Validate message content
	if err := mv.validateMessageContent(msg); err != nil {
		mv.recordViolation(peerID, "invalid_content")
		return err
	}

	// Check rate limiting
	if err := mv.checkRateLimit(peerID); err != nil {
		mv.recordViolation(peerID, "rate_limit_exceeded")
		return err
	}

	return nil
}

// validateMessageSize checks if message size is within limits
func (mv *MessageValidator) validateMessageSize(msg *Message) error {
	// Estimate message size (this is a simplified approach)
	// In a real implementation, you'd serialize the message to get exact size
	estimatedSize := 0

	// Add basic fields
	estimatedSize += 8 // TimestampUnixNano
	if msg.FromPeerId != nil {
		estimatedSize += len(msg.FromPeerId)
	}
	if msg.Signature != nil {
		estimatedSize += len(msg.Signature)
	}

	// Add content size
	if msg.Content != nil {
		switch content := msg.Content.(type) {
		case *Message_BlockMessage:
			if content.BlockMessage != nil && content.BlockMessage.BlockData != nil {
				estimatedSize += len(content.BlockMessage.BlockData)
			}
		case *Message_TransactionMessage:
			if content.TransactionMessage != nil && content.TransactionMessage.TransactionData != nil {
				estimatedSize += len(content.TransactionMessage.TransactionData)
			}
		}
	}

	if estimatedSize > mv.config.MaxMessageSize {
		return fmt.Errorf("message size %d exceeds maximum allowed size %d", estimatedSize, mv.config.MaxMessageSize)
	}

	return nil
}

// validateTimestamp checks if message timestamp is valid
func (mv *MessageValidator) validateTimestamp(msg *Message) error {
	now := time.Now().UnixNano()
	msgTime := msg.TimestampUnixNano

	// Check if timestamp is in the future
	if msgTime > now {
		return fmt.Errorf("message timestamp %d is in the future", msgTime)
	}

	// Check if timestamp is too old
	tolerance := mv.config.TimestampTolerance.Nanoseconds()
	if now-msgTime > tolerance {
		return fmt.Errorf("message timestamp %d is too old (tolerance: %d)", msgTime, tolerance)
	}

	return nil
}

// validateMessageContent validates the content of the message
func (mv *MessageValidator) validateMessageContent(msg *Message) error {
	if msg.Content == nil {
		return fmt.Errorf("message has no content")
	}

	// Validate based on content type
	switch content := msg.Content.(type) {
	case *Message_BlockMessage:
		return mv.validateBlockMessage(content.BlockMessage)
	case *Message_TransactionMessage:
		return mv.validateTransactionMessage(content.TransactionMessage)
	case *Message_HeadersRequest:
		return mv.validateHeadersRequest(content.HeadersRequest)
	case *Message_BlockRequest:
		return mv.validateBlockRequest(content.BlockRequest)
	case *Message_SyncRequest:
		return mv.validateSyncRequest(content.SyncRequest)
	case *Message_StateRequest:
		return mv.validateStateRequest(content.StateRequest)
	}

	return nil
}

// validateBlockMessage validates block message content
func (mv *MessageValidator) validateBlockMessage(blockMsg *BlockMessage) error {
	if blockMsg == nil {
		return fmt.Errorf("block message is nil")
	}

	if blockMsg.BlockData == nil {
		return fmt.Errorf("block data is nil")
	}

	if len(blockMsg.BlockData) == 0 {
		return fmt.Errorf("block data is empty")
	}

	// Additional validation could include:
	// - Block size limits
	// - Block format validation
	// - Cryptographic signature verification

	return nil
}

// validateTransactionMessage validates transaction message content
func (mv *MessageValidator) validateTransactionMessage(txMsg *TransactionMessage) error {
	if txMsg == nil {
		return fmt.Errorf("transaction message is nil")
	}

	if txMsg.TransactionData == nil {
		return fmt.Errorf("transaction data is nil")
	}

	if len(txMsg.TransactionData) == 0 {
		return fmt.Errorf("transaction data is empty")
	}

	return nil
}

// validateHeadersRequest validates headers request content
func (mv *MessageValidator) validateHeadersRequest(req *BlockHeadersRequest) error {
	if req == nil {
		return fmt.Errorf("headers request is nil")
	}

	// Validate count limits to prevent DoS
	if req.Count > 1000 {
		return fmt.Errorf("headers request count %d exceeds maximum allowed", req.Count)
	}

	return nil
}

// validateBlockRequest validates block request content
func (mv *MessageValidator) validateBlockRequest(req *BlockRequest) error {
	if req == nil {
		return fmt.Errorf("block request is nil")
	}

	// Validate that either hash or height is provided
	if req.BlockHash == nil && req.Height == 0 {
		return fmt.Errorf("block request must specify either hash or height")
	}

	return nil
}

// validateSyncRequest validates sync request content
func (mv *MessageValidator) validateSyncRequest(req *SyncRequest) error {
	if req == nil {
		return fmt.Errorf("sync request is nil")
	}

	// Validate known headers count to prevent DoS
	if len(req.KnownHeaders) > 1000 {
		return fmt.Errorf("sync request has too many known headers: %d", len(req.KnownHeaders))
	}

	return nil
}

// validateStateRequest validates state request content
func (mv *MessageValidator) validateStateRequest(req *StateRequest) error {
	if req == nil {
		return fmt.Errorf("state request is nil")
	}

	// Validate account hashes count to prevent DoS
	if len(req.AccountHashes) > 1000 {
		return fmt.Errorf("state request has too many account hashes: %d", len(req.AccountHashes))
	}

	return nil
}

// checkRateLimit checks if peer has exceeded rate limits
func (mv *MessageValidator) checkRateLimit(peerID string) error {
	limiter, exists := mv.rateLimits[peerID]
	if !exists {
		limiter = &RateLimiter{
			requests:    make([]time.Time, 0),
			lastReset:   time.Now(),
			windowSize:  time.Second,
			maxRequests: mv.config.RateLimitPerSecond,
		}
		mv.rateLimits[peerID] = limiter
	}

	now := time.Now()

	// Reset window if needed
	if now.Sub(limiter.lastReset) >= limiter.windowSize {
		limiter.requests = limiter.requests[:0]
		limiter.lastReset = now
	}

	// Check if rate limit exceeded
	if len(limiter.requests) >= limiter.maxRequests {
		return fmt.Errorf("rate limit exceeded for peer %s", peerID)
	}

	// Add current request
	limiter.requests = append(limiter.requests, now)

	return nil
}

// recordViolation records a security violation for a peer
func (mv *MessageValidator) recordViolation(peerID, violationType string) {
	stats, exists := mv.peerStats[peerID]
	if !exists {
		stats = &PeerStats{}
		mv.peerStats[peerID] = stats
	}

	stats.ViolationCount++
	stats.LastSeen = time.Now()

	// Blacklist peer if threshold exceeded
	if stats.ViolationCount >= mv.config.BlacklistThreshold {
		mv.blacklist[peerID] = true
	}
}

// updatePeerStats updates peer statistics
func (mv *MessageValidator) updatePeerStats(peerID string) {
	stats, exists := mv.peerStats[peerID]
	if !exists {
		stats = &PeerStats{}
		mv.peerStats[peerID] = stats
	}

	stats.MessageCount++
	stats.LastSeen = time.Now()
}

// IsPeerBlacklisted checks if a peer is blacklisted
func (mv *MessageValidator) IsPeerBlacklisted(peerID string) bool {
	return mv.blacklist[peerID]
}

// GetPeerStats returns statistics for a peer
func (mv *MessageValidator) GetPeerStats(peerID string) *PeerStats {
	return mv.peerStats[peerID]
}

// RemovePeer removes a peer from tracking
func (mv *MessageValidator) RemovePeer(peerID string) {
	delete(mv.peerStats, peerID)
	delete(mv.rateLimits, peerID)
	delete(mv.blacklist, peerID)
}

// GetBlacklist returns the current blacklist
func (mv *MessageValidator) GetBlacklist() map[string]bool {
	result := make(map[string]bool)
	for k, v := range mv.blacklist {
		result[k] = v
	}
	return result
}

// ClearBlacklist clears the blacklist
func (mv *MessageValidator) ClearBlacklist() {
	mv.blacklist = make(map[string]bool)
}

// CalculateMessageHash calculates a hash for message integrity
func (mv *MessageValidator) CalculateMessageHash(msg *Message) string {
	// Create a deterministic representation for hashing
	hashInput := fmt.Sprintf("%d:%x:%x",
		msg.TimestampUnixNano,
		msg.FromPeerId,
		msg.Signature)

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// ValidateMessageSignature validates message signature
func (mv *MessageValidator) ValidateMessageSignature(msg *Message, expectedPeerID []byte) error {
	if msg.Signature == nil {
		return fmt.Errorf("message has no signature")
	}

	if msg.FromPeerId == nil {
		return fmt.Errorf("message has no peer ID")
	}

	// In a real implementation, you would:
	// 1. Verify the signature cryptographically
	// 2. Check that the peer ID matches the expected one
	// 3. Validate the signature against the message content

	// For now, we'll do basic validation
	if len(msg.Signature) == 0 {
		return fmt.Errorf("message signature is empty")
	}

	if !bytesEqual(msg.FromPeerId, expectedPeerID) {
		return fmt.Errorf("peer ID mismatch: expected %x, got %x", expectedPeerID, msg.FromPeerId)
	}

	return nil
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
