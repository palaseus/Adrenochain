// Package wormhole implements Wormhole Verified Action Approvals (VAAs) for
// cross-chain messaging to and from Solana.
//
// Wormhole is the dominant cross-chain messaging rail for Solana as of
// November 2026 (chain ID 1 for Solana). A VAA is a signed cross-chain message
// attested by Wormhole's guardian network. A VAA is considered valid when a
// threshold (2/3+) of guardians have signed its SHA-256 hash.
//
// Reference: https://wormhole.com/docs/protocol/core/vaa
//
// See pkg/multichain/solana/README.md for usage.

package solana

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// VAAVersion is the current VAA version.
const VAAVersion = 1

// ChainID is a Wormhole chain identifier.
type ChainID uint16

const (
	// ChainIDSolana is the Wormhole chain ID for Solana.
	ChainIDSolana ChainID = 1
	// ChainIDEthereum is the Wormhole chain ID for Ethereum.
	ChainIDEthereum ChainID = 2
)

// VAA is a Wormhole Verified Action Approval.
type VAA struct {
	// Version is the VAA version (must be 1).
	Version uint8

	// Signatures are the guardian signatures (65 bytes each: recovery ID + 64-byte sig).
	Signatures [][]byte

	// Timestamp is the VAA emission timestamp (Unix seconds).
	Timestamp uint32

	// Nonce is the guardian set nonce.
	Nonce byte

	// EmitterChain is the chain that emitted the message.
	EmitterChain ChainID

	// EmitterAddress is the emitter address on the source chain (32 bytes).
	EmitterAddress [32]byte

	// Sequence is the message sequence number.
	Sequence uint64

	// ConsistencyLevel is the Wormhole consistency level.
	ConsistencyLevel byte

	// Payload is the arbitrary cross-chain message payload.
	Payload []byte

	// Hash is the SHA-256 of the VAA body (everything before this field).
	Hash [32]byte
}

// NewVAA creates a new VAA with the given fields.
func NewVAA(timestamp uint32, nonce byte, emitterChain ChainID, emitterAddress [32]byte, sequence uint64, consistencyLevel byte, payload []byte) *VAA {
	return &VAA{
		Version:          VAAVersion,
		Timestamp:        timestamp,
		Nonce:            nonce,
		EmitterChain:     emitterChain,
		EmitterAddress:   emitterAddress,
		Sequence:         sequence,
		ConsistencyLevel: consistencyLevel,
		Payload:          payload,
	}
}

// ComputeHash computes the VAA's SHA-256 hash over the serialized body (version
// through payload, excluding the hash field itself).
func (v *VAA) ComputeHash() [32]byte {
	return sha256.Sum256(v.serializeBody())
}

// serializeBody serializes the VAA body (version through payload) for hashing
// and for the wire format.
func (v *VAA) serializeBody() []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte(v.Version)
	buf.WriteByte(byte(len(v.Signatures)))

	// Signatures.
	for _, sig := range v.Signatures {
		buf.Write(sig)
	}

	// Timestamp (4 bytes, big-endian).
	ts := make([]byte, 4)
	binary.BigEndian.PutUint32(ts, v.Timestamp)
	buf.Write(ts)

	// Nonce (1 byte).
	buf.WriteByte(v.Nonce)

	// Emitter chain (2 bytes, big-endian).
	ec := make([]byte, 2)
	binary.BigEndian.PutUint16(ec, uint16(v.EmitterChain))
	buf.Write(ec)

	// Emitter address (32 bytes).
	buf.Write(v.EmitterAddress[:])

	// Sequence (8 bytes, big-endian).
	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, v.Sequence)
	buf.Write(seq)

	// Consistency level (1 byte).
	buf.WriteByte(v.ConsistencyLevel)

	// Payload length (2 bytes, big-endian).
	pl := make([]byte, 2)
	binary.BigEndian.PutUint16(pl, uint16(len(v.Payload)))
	buf.Write(pl)

	// Payload.
	buf.Write(v.Payload)

	return buf.Bytes()
}

// Serialize serializes the complete VAA (body + hash) to wire format.
func (v *VAA) Serialize() ([]byte, error) {
	body := v.serializeBody()
	v.Hash = sha256.Sum256(body)

	buf := new(bytes.Buffer)
	buf.Write(body)
	buf.Write(v.Hash[:])
	return buf.Bytes(), nil
}

// Deserialize parses a VAA from wire format (body + hash).
func Deserialize(data []byte) (*VAA, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short for VAA")
	}

	v := &VAA{}
	off := 0

	v.Version = data[off]
	off++
	if v.Version != VAAVersion {
		return nil, fmt.Errorf("unsupported VAA version %d", v.Version)
	}

	sigLen := int(data[off])
	off++

	// Each signature is 65 bytes.
	if len(data) < off+sigLen*65 {
		return nil, errors.New("data too short for signatures")
	}
	v.Signatures = make([][]byte, sigLen)
	for i := 0; i < sigLen; i++ {
		v.Signatures[i] = make([]byte, 65)
		copy(v.Signatures[i], data[off:off+65])
		off += 65
	}

	// Timestamp.
	if len(data) < off+4 {
		return nil, errors.New("data too short for timestamp")
	}
	v.Timestamp = binary.BigEndian.Uint32(data[off : off+4])
	off += 4

	// Nonce.
	if len(data) < off+1 {
		return nil, errors.New("data too short for nonce")
	}
	v.Nonce = data[off]
	off++

	// Emitter chain.
	if len(data) < off+2 {
		return nil, errors.New("data too short for emitter chain")
	}
	v.EmitterChain = ChainID(binary.BigEndian.Uint16(data[off : off+2]))
	off += 2

	// Emitter address.
	if len(data) < off+32 {
		return nil, errors.New("data too short for emitter address")
	}
	copy(v.EmitterAddress[:], data[off:off+32])
	off += 32

	// Sequence.
	if len(data) < off+8 {
		return nil, errors.New("data too short for sequence")
	}
	v.Sequence = binary.BigEndian.Uint64(data[off : off+8])
	off += 8

	// Consistency level.
	if len(data) < off+1 {
		return nil, errors.New("data too short for consistency level")
	}
	v.ConsistencyLevel = data[off]
	off++

	// Payload length.
	if len(data) < off+2 {
		return nil, errors.New("data too short for payload length")
	}
	payloadLen := int(binary.BigEndian.Uint16(data[off : off+2]))
	off += 2

	// Payload.
	if len(data) < off+payloadLen {
		return nil, errors.New("data too short for payload")
	}
	v.Payload = make([]byte, payloadLen)
	copy(v.Payload, data[off:off+payloadLen])
	off += payloadLen

	// Hash.
	if len(data) < off+32 {
		return nil, errors.New("data too short for VAA hash")
	}
	copy(v.Hash[:], data[off:off+32])

	return v, nil
}

// VerifyHash verifies that the stored VAA hash matches the recomputed hash.
func (v *VAA) VerifyHash() bool {
	return v.Hash == v.ComputeHash()
}

// SignatureCount returns the number of guardian signatures.
func (v *VAA) SignatureCount() int {
	return len(v.Signatures)
}

// VerifyThreshold checks whether the VAA meets the Wormhole guardian threshold
// (2/3+ of the guardian set have signed). This verifies the hash and counts
// signatures against a provided guardian count.
func VerifyThreshold(sigCount int, guardianCount uint8) bool {
	if guardianCount == 0 {
		return false
	}
	// Wormhole threshold: signatures > 2/3 * guardianCount.
	required := (2 * int(guardianCount)) / 3
	return sigCount > required
}
