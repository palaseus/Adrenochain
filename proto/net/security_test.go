package net

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSecurityConfig tests security configuration
func TestSecurityConfig(t *testing.T) {
	t.Run("DefaultSecurityConfig", func(t *testing.T) {
		config := DefaultSecurityConfig()
		assert.NotNil(t, config)
		assert.Equal(t, 1024*1024, config.MaxMessageSize)
		assert.Equal(t, 100, config.MaxPeerConnections)
		assert.Equal(t, 10, config.RateLimitPerSecond)
		assert.Equal(t, 30*time.Second, config.TimestampTolerance)
		assert.Equal(t, 5, config.MaxConcurrentRequests)
		assert.Equal(t, 3, config.BlacklistThreshold)
	})

	t.Run("CustomSecurityConfig", func(t *testing.T) {
		config := &SecurityConfig{
			MaxMessageSize:        2048 * 1024,
			MaxPeerConnections:    50,
			RateLimitPerSecond:    20,
			TimestampTolerance:    60 * time.Second,
			MaxConcurrentRequests: 10,
			BlacklistThreshold:    5,
		}
		assert.Equal(t, 2048*1024, config.MaxMessageSize)
		assert.Equal(t, 50, config.MaxPeerConnections)
		assert.Equal(t, 20, config.RateLimitPerSecond)
		assert.Equal(t, 60*time.Second, config.TimestampTolerance)
		assert.Equal(t, 10, config.MaxConcurrentRequests)
		assert.Equal(t, 5, config.BlacklistThreshold)
	})
}

// TestMessageValidator tests message validator functionality
func TestMessageValidator(t *testing.T) {
	t.Run("NewMessageValidator", func(t *testing.T) {
		validator := NewMessageValidator(nil)
		assert.NotNil(t, validator)
		assert.NotNil(t, validator.config)
		assert.Equal(t, DefaultSecurityConfig().MaxMessageSize, validator.config.MaxMessageSize)
	})

	t.Run("NewMessageValidatorWithConfig", func(t *testing.T) {
		config := &SecurityConfig{
			MaxMessageSize:     512 * 1024,
			BlacklistThreshold: 2,
		}
		validator := NewMessageValidator(config)
		assert.NotNil(t, validator)
		assert.Equal(t, config, validator.config)
		assert.Equal(t, 512*1024, validator.config.MaxMessageSize)
		assert.Equal(t, 2, validator.config.BlacklistThreshold)
	})
}

// TestMessageValidation tests message validation
func TestMessageValidation(t *testing.T) {
	validator := NewMessageValidator(nil)
	peerID := "test-peer-123"

	t.Run("ValidBlockMessage", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})

	t.Run("ValidTransactionMessage", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_TransactionMessage{
				TransactionMessage: &TransactionMessage{
					TransactionData: []byte("transaction data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})

	t.Run("ValidHeadersRequest", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_HeadersRequest{
				HeadersRequest: &BlockHeadersRequest{
					StartHeight: 100,
					Count:       50,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})

	t.Run("ValidBlockRequest", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockRequest{
				BlockRequest: &BlockRequest{
					BlockHash: []byte("block-hash"),
					Height:    100,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})

	t.Run("ValidSyncRequest", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_SyncRequest{
				SyncRequest: &SyncRequest{
					CurrentHeight: 100,
					BestBlockHash: []byte("best-hash"),
					KnownHeaders:  [][]byte{[]byte("header1"), []byte("header2")},
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})

	t.Run("ValidStateRequest", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_StateRequest{
				StateRequest: &StateRequest{
					StateRoot:     []byte("state-root"),
					AccountHashes: [][]byte{[]byte("account1"), []byte("account2")},
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})
}

// TestMessageValidationErrors tests message validation error cases
func TestMessageValidationErrors(t *testing.T) {
	config := &SecurityConfig{
		MaxMessageSize:     1024 * 1024,
		TimestampTolerance: 30 * time.Second,
		BlacklistThreshold: 100, // Very high threshold to avoid blacklisting
	}
	validator := NewMessageValidator(config)
	peerID := "test-peer-errors"

	t.Run("MessageTooLarge", func(t *testing.T) {
		// Create a message that exceeds the size limit
		largeData := make([]byte, 2*1024*1024) // 2MB
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: largeData,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message size")
		assert.Contains(t, err.Error(), "exceeds maximum allowed size")
	})

	t.Run("FutureTimestamp", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().Add(1 * time.Hour).UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp")
		assert.Contains(t, err.Error(), "in the future")
	})

	t.Run("OldTimestamp", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().Add(-2 * time.Minute).UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp")
		assert.Contains(t, err.Error(), "too old")
	})

	t.Run("NoContent", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content:           nil,
			Signature:         []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message has no content")
	})

	t.Run("NilBlockMessage", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: nil,
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "block message is nil")
	})

	t.Run("EmptyBlockData", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte{},
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "block data is empty")
	})

	t.Run("HeadersRequestCountExceeded", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_HeadersRequest{
				HeadersRequest: &BlockHeadersRequest{
					StartHeight: 100,
					Count:       1500, // Exceeds limit of 1000
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "headers request count")
		assert.Contains(t, err.Error(), "exceeds maximum allowed")
	})

	t.Run("BlockRequestNoHashOrHeight", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockRequest{
				BlockRequest: &BlockRequest{
					BlockHash: nil,
					Height:    0,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must specify either hash or height")
	})

	t.Run("SyncRequestTooManyHeaders", func(t *testing.T) {
		// Create too many known headers
		knownHeaders := make([][]byte, 1500) // Exceeds limit of 1000
		for i := range knownHeaders {
			knownHeaders[i] = []byte(fmt.Sprintf("header-%d", i))
		}

		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_SyncRequest{
				SyncRequest: &SyncRequest{
					CurrentHeight: 100,
					BestBlockHash: []byte("best-hash"),
					KnownHeaders:  knownHeaders,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many known headers")
	})

	t.Run("StateRequestTooManyAccounts", func(t *testing.T) {
		// Create too many account hashes
		accountHashes := make([][]byte, 1500) // Exceeds limit of 1000
		for i := range accountHashes {
			accountHashes[i] = []byte(fmt.Sprintf("account-%d", i))
		}

		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_StateRequest{
				StateRequest: &StateRequest{
					StateRoot:     []byte("state-root"),
					AccountHashes: accountHashes,
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many account hashes")
	})
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	config := &SecurityConfig{
		RateLimitPerSecond: 3,                // Allow only 3 requests per second
		MaxMessageSize:     1024 * 1024,      // Set reasonable message size limit
		TimestampTolerance: 30 * time.Second, // Set proper timestamp tolerance
		BlacklistThreshold: 10,               // Set high threshold to avoid blacklisting during test
	}
	validator := NewMessageValidator(config)
	peerID := "test-peer-rate-limit"

	t.Run("RateLimitNotExceeded", func(t *testing.T) {
		// Send 3 messages (within limit)
		for i := 0; i < 3; i++ {
			msg := &Message{
				TimestampUnixNano: time.Now().UnixNano(),
				FromPeerId:        []byte("peer-id"),
				Content: &Message_BlockMessage{
					BlockMessage: &BlockMessage{
						BlockData: []byte("block data"),
					},
				},
				Signature: []byte("signature"),
			}

			err := validator.ValidateMessage(msg, peerID)
			assert.NoError(t, err)
		}
	})

	t.Run("RateLimitExceeded", func(t *testing.T) {
		// Send 4th message (exceeds limit)
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("RateLimitResetAfterWindow", func(t *testing.T) {
		// Wait for rate limit window to reset
		time.Sleep(1100 * time.Millisecond)

		// Should be able to send messages again
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)
	})
}

// TestBlacklisting tests blacklisting functionality
func TestBlacklisting(t *testing.T) {
	config := &SecurityConfig{
		BlacklistThreshold: 2, // Blacklist after 2 violations
	}
	validator := NewMessageValidator(config)
	peerID := "test-peer-123"

	t.Run("PeerNotBlacklistedInitially", func(t *testing.T) {
		assert.False(t, validator.IsPeerBlacklisted(peerID))
	})

	t.Run("PeerBlacklistedAfterViolations", func(t *testing.T) {
		// Send message with future timestamp (violation 1)
		msg1 := &Message{
			TimestampUnixNano: time.Now().Add(1 * time.Hour).UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg1, peerID)
		assert.Error(t, err)
		assert.False(t, validator.IsPeerBlacklisted(peerID)) // Not blacklisted yet

		// Send message with old timestamp (violation 2)
		msg2 := &Message{
			TimestampUnixNano: time.Now().Add(-2 * time.Minute).UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err = validator.ValidateMessage(msg2, peerID)
		assert.Error(t, err)
		assert.True(t, validator.IsPeerBlacklisted(peerID)) // Now blacklisted
	})

	t.Run("BlacklistedPeerRejected", func(t *testing.T) {
		// Try to send a valid message from blacklisted peer
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is blacklisted")
	})
}

// TestPeerStats tests peer statistics tracking
func TestPeerStats(t *testing.T) {
	validator := NewMessageValidator(nil)
	peerID := "test-peer-123"

	t.Run("PeerStatsInitialization", func(t *testing.T) {
		stats := validator.GetPeerStats(peerID)
		assert.Nil(t, stats) // No stats initially
	})

	t.Run("PeerStatsUpdatedAfterValidMessage", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)

		stats := validator.GetPeerStats(peerID)
		assert.NotNil(t, stats)
		assert.Equal(t, 1, stats.MessageCount)
		assert.Equal(t, 0, stats.ViolationCount)
		assert.True(t, stats.LastSeen.After(time.Now().Add(-1*time.Second)))
	})

	t.Run("PeerStatsUpdatedAfterViolation", func(t *testing.T) {
		// Send invalid message
		msg := &Message{
			TimestampUnixNano: time.Now().Add(1 * time.Hour).UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.Error(t, err)

		stats := validator.GetPeerStats(peerID)
		assert.NotNil(t, stats)
		assert.Equal(t, 2, stats.MessageCount) // 1 valid + 1 invalid = 2 total
		assert.Equal(t, 1, stats.ViolationCount)
	})
}

// TestPeerManagement tests peer management functionality
func TestPeerManagement(t *testing.T) {
	validator := NewMessageValidator(nil)
	peerID := "test-peer-123"

	t.Run("RemovePeer", func(t *testing.T) {
		// First, add some stats for the peer
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer-id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peerID)
		assert.NoError(t, err)

		// Verify peer has stats
		stats := validator.GetPeerStats(peerID)
		assert.NotNil(t, stats)

		// Remove peer
		validator.RemovePeer(peerID)

		// Verify peer is removed
		stats = validator.GetPeerStats(peerID)
		assert.Nil(t, stats)
		assert.False(t, validator.IsPeerBlacklisted(peerID))
	})

	t.Run("ClearBlacklist", func(t *testing.T) {
		// Add a peer to blacklist
		validator.blacklist[peerID] = true
		assert.True(t, validator.IsPeerBlacklisted(peerID))

		// Clear blacklist
		validator.ClearBlacklist()
		assert.False(t, validator.IsPeerBlacklisted(peerID))
	})

	t.Run("GetBlacklist", func(t *testing.T) {
		// Add a peer to blacklist
		validator.blacklist[peerID] = true

		// Get blacklist
		blacklist := validator.GetBlacklist()
		assert.NotNil(t, blacklist)
		assert.True(t, blacklist[peerID])

		// Verify it's a copy, not the original
		delete(blacklist, peerID)
		assert.True(t, validator.IsPeerBlacklisted(peerID)) // Original unchanged
	})
}

// TestMessageHash tests message hash calculation
func TestMessageHash(t *testing.T) {
	validator := NewMessageValidator(nil)

	t.Run("CalculateMessageHash", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: 1234567890,
			FromPeerId:        []byte("peer-id"),
			Signature:         []byte("signature"),
		}

		hash1 := validator.CalculateMessageHash(msg)
		hash2 := validator.CalculateMessageHash(msg)

		// Same message should produce same hash
		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
		assert.Len(t, hash1, 64) // SHA256 hex string length
	})

	t.Run("DifferentMessagesDifferentHashes", func(t *testing.T) {
		msg1 := &Message{
			TimestampUnixNano: 1234567890,
			FromPeerId:        []byte("peer-id-1"),
			Signature:         []byte("signature-1"),
		}

		msg2 := &Message{
			TimestampUnixNano: 1234567890,
			FromPeerId:        []byte("peer-id-2"),
			Signature:         []byte("signature-2"),
		}

		hash1 := validator.CalculateMessageHash(msg1)
		hash2 := validator.CalculateMessageHash(msg2)

		// Different messages should produce different hashes
		assert.NotEqual(t, hash1, hash2)
	})
}

// TestMessageSignatureValidation tests message signature validation
func TestMessageSignatureValidation(t *testing.T) {
	validator := NewMessageValidator(nil)

	t.Run("ValidateMessageSignature", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("expected-peer-id"),
			Signature:         []byte("valid-signature"),
		}

		err := validator.ValidateMessageSignature(msg, []byte("expected-peer-id"))
		assert.NoError(t, err)
	})

	t.Run("ValidateMessageSignatureNoSignature", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("expected-peer-id"),
			Signature:         nil,
		}

		err := validator.ValidateMessageSignature(msg, []byte("expected-peer-id"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message has no signature")
	})

	t.Run("ValidateMessageSignatureNoPeerID", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        nil,
			Signature:         []byte("valid-signature"),
		}

		err := validator.ValidateMessageSignature(msg, []byte("expected-peer-id"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message has no peer ID")
	})

	t.Run("ValidateMessageSignatureEmptySignature", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("expected-peer-id"),
			Signature:         []byte{},
		}

		err := validator.ValidateMessageSignature(msg, []byte("expected-peer-id"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message signature is empty")
	})

	t.Run("ValidateMessageSignaturePeerIDMismatch", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("actual-peer-id"),
			Signature:         []byte("valid-signature"),
		}

		err := validator.ValidateMessageSignature(msg, []byte("expected-peer-id"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "peer ID mismatch")
	})
}

// TestAntiEclipseProtection tests anti-eclipse attack protection
func TestAntiEclipseProtection(t *testing.T) {
	config := &SecurityConfig{
		MaxPeerConnections: 5,                // Limit to 5 peer connections
		BlacklistThreshold: 2,                // Blacklist after 2 violations
		MaxMessageSize:     1024 * 1024,      // Set reasonable message size limit
		TimestampTolerance: 30 * time.Second, // Set proper timestamp tolerance
		RateLimitPerSecond: 2,                // Allow only 2 messages per second for testing
	}
	validator := NewMessageValidator(config)

	t.Run("MultiplePeersRateLimiting", func(t *testing.T) {
		// Test that multiple peers are tracked separately
		peer1 := "peer-1"
		peer2 := "peer-2"

		// Send messages from peer1 with small delays to ensure rate limiting works
		for i := 0; i < 3; i++ {
			msg := &Message{
				TimestampUnixNano: time.Now().UnixNano(),
				FromPeerId:        []byte(peer1),
				Content: &Message_BlockMessage{
					BlockMessage: &BlockMessage{
						BlockData: []byte("block data"),
					},
				},
				Signature: []byte("signature"),
			}

			err := validator.ValidateMessage(msg, peer1)
			if i < 2 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err) // Rate limit exceeded
				assert.Contains(t, err.Error(), "rate limit exceeded")
			}

			// Small delay to ensure messages are processed sequentially
			time.Sleep(10 * time.Millisecond)
		}

		// Peer2 should still be able to send messages
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte(peer2),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(msg, peer2)
		assert.NoError(t, err)
	})

	t.Run("PeerIsolation", func(t *testing.T) {
		// Test that blacklisted peers are isolated
		peerID := "malicious-peer"

		// Send invalid messages to trigger blacklisting
		for i := 0; i < 3; i++ {
			msg := &Message{
				TimestampUnixNano: time.Now().Add(1 * time.Hour).UnixNano(), // Future timestamp
				FromPeerId:        []byte(peerID),
				Content: &Message_BlockMessage{
					BlockMessage: &BlockMessage{
						BlockData: []byte("block data"),
					},
				},
				Signature: []byte("signature"),
			}

			validator.ValidateMessage(msg, peerID)
		}

		// Verify peer is blacklisted
		assert.True(t, validator.IsPeerBlacklisted(peerID))

		// Try to send valid message from blacklisted peer
		validMsg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte(peerID),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{
					BlockData: []byte("block data"),
				},
			},
			Signature: []byte("signature"),
		}

		err := validator.ValidateMessage(validMsg, peerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is blacklisted")
	})
}
