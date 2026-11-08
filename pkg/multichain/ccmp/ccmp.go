// Package ccmp implements the Cross-Chain Messaging Protocol (CCMP) for
// Adrenochain. CCMP enables atomic, ordered, and authenticated messages to be
// relayed between Adrenochain and all registered chains in the multi-chain
// registry.
//
// Design goals:
//   - Atomicity: a message is either delivered on both chains or neither.
//   - Ordering: per-source-chain message sequencing.
//   - Authentication: origin chain validates the sender; destination chain
//     verifies the relayer signature and message hash.
//   - Replay protection: nonces per source chain + destination pair.
//   - Timeout/refund: expired messages can be cancelled or refunded.
//
// See pkg/multichain/README.md for the full architecture.

package ccmp

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"sync"
	"time"
)

// MessageType categorizes a CCMP message.
type MessageType int

const (
	MessageTypeTransfer MessageType = iota
	MessageTypeExecution
	MessageTypeGovernance
	MessageTypeRefund
	MessageTypeCancel
)

// MessageStatus tracks the lifecycle of a cross-chain message.
type MessageStatus int

const (
	MessageStatusPending MessageStatus = iota
	MessageStatusRelayed
	MessageStatusExecuted
	MessageStatusFailed
	MessageStatusExpired
	MessageStatusRefunded
)

// Message is a cross-chain CCMP message.
type Message struct {
	// ID is a globally unique message identifier (32 bytes).
	ID [32]byte
	// SourceChain and DestinationChain identify the participating chains.
	SourceChain      uint64
	DestinationChain uint64
	// Type is the message category.
	Type MessageType
	// Nonce prevents replay per (source, destination) pair.
	Nonce *big.Int
	// Sender is the 20-byte sender identifier on the source chain.
	Sender [20]byte
	// Recipient is the 20-byte recipient identifier on the destination chain.
	Recipient [20]byte
	// Payload is the opaque message payload.
	Payload []byte
	// Fee is the relay fee paid to relayers.
	Fee *big.Int
	// Expiry is the absolute deadline after which the message is invalid.
	Expiry time.Time
	// Status is the current lifecycle status.
	Status MessageStatus
	// Signature is the origin-chain validator signature over the message hash.
	Signature []byte
	// RelayerSignature is the relayer's signature confirming delivery.
	RelayerSignature []byte
	// CreatedAt records when the message was created.
	CreatedAt time.Time
}

// NewMessage creates a new CCMP message with a random ID and nonce.
func NewMessage(source, dest uint64, msgType MessageType, sender, recipient [20]byte, payload []byte, fee *big.Int, expiry time.Time) *Message {
	var id [32]byte
	_, _ = rand.Read(id[:])
	return &Message{
		ID:               id,
		SourceChain:      source,
		DestinationChain: dest,
		Type:             msgType,
		Nonce:            new(big.Int).SetBytes(id[:]),
		Sender:           sender,
		Recipient:        recipient,
		Payload:          payload,
		Fee:              fee,
		Expiry:           expiry,
		Status:           MessageStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
}

// Hash returns the SHA-256 digest of the message content (excluding signatures).
func (m *Message) Hash() []byte {
	h := sha256.New()
	h.Write(m.ID[:])
	h.Write(encodeUint64(m.SourceChain))
	h.Write(encodeUint64(m.DestinationChain))
	h.Write([]byte{byte(m.Type)})
	h.Write(m.Nonce.Bytes())
	h.Write(m.Sender[:])
	h.Write(m.Recipient[:])
	h.Write(m.Payload)
	h.Write(m.Fee.Bytes())
	h.Write(m.Expiry.UTC().AppendFormat(nil, time.RFC3339Nano))
	return h.Sum(nil)
}

// IsExpired reports whether the message is past its expiry.
func (m *Message) IsExpired() bool {
	return time.Now().UTC().After(m.Expiry)
}

// ValidateBasic checks structural invariants.
func (m *Message) ValidateBasic() error {
	if m.Nonce == nil || m.Nonce.Sign() <= 0 {
		return errors.New("invalid nonce")
	}
	if m.Fee == nil || m.Fee.Sign() < 0 {
		return errors.New("invalid fee")
	}
	if m.IsExpired() {
		return errors.New("message expired")
	}
	if len(m.Payload) == 0 {
		return errors.New("empty payload")
	}
	if len(m.Signature) == 0 {
		return errors.New("missing origin signature")
	}
	return nil
}

// Relay is a relayer's attestation that a message was delivered.
type Relay struct {
	MessageID [32]byte
	Relayer   [20]byte
	Signature []byte
	Timestamp time.Time
}

// RelayRegistry tracks relays to prevent double-relay rewards.
type RelayRegistry struct {
	mu  sync.RWMutex
	rel map[[32]byte][]byte // messageID -> relayer
}

// NewRelayRegistry creates an empty relay registry.
func NewRelayRegistry() *RelayRegistry {
	return &RelayRegistry{rel: make(map[[32]byte][]byte)}
}

// Record records a relay for a message, returning false if already recorded.
func (rr *RelayRegistry) Record(msgID [32]byte, relayer [20]byte) bool {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	if _, exists := rr.rel[msgID]; exists {
		return false
	}
	rr.rel[msgID] = relayer[:]
	return true
}

// HasRelay reports whether a message has been relayed.
func (rr *RelayRegistry) HasRelay(msgID [32]byte) bool {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	_, ok := rr.rel[msgID]
	return ok
}

func encodeUint64(v uint64) []byte {
	b := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		b[7-i] = byte(v >> (8 * i))
	}
	return b
}
