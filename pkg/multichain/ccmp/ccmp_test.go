package ccmp

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageHashAndValidate(t *testing.T) {
	sender := [20]byte{1}
	recipient := [20]byte{2}
	expiry := time.Now().Add(time.Hour)
	msg := NewMessage(1, 2, MessageTypeTransfer, sender, recipient, []byte("hello"), big.NewInt(100), expiry)
	msg.Signature = []byte("origin-signature")

	assert.NotEqual(t, [32]byte{}, msg.ID)
	assert.Equal(t, uint64(1), msg.SourceChain)
	assert.Equal(t, uint64(2), msg.DestinationChain)
	assert.Equal(t, MessageTypeTransfer, msg.Type)
	assert.Equal(t, sender, msg.Sender)
	assert.Equal(t, recipient, msg.Recipient)
	assert.Equal(t, MessageStatusPending, msg.Status)
	assert.False(t, msg.IsExpired())
	assert.NotEmpty(t, msg.Hash())
	assert.NoError(t, msg.ValidateBasic())
}

func TestMessageExpiry(t *testing.T) {
	sender := [20]byte{1}
	recipient := [20]byte{2}
	expired := time.Now().Add(-time.Hour)
	msg := NewMessage(1, 2, MessageTypeTransfer, sender, recipient, []byte("x"), big.NewInt(1), expired)

	assert.True(t, msg.IsExpired())
	assert.Error(t, msg.ValidateBasic())
}

func TestRelayRegistry(t *testing.T) {
	rr := NewRelayRegistry()
	var id [32]byte
	copy(id[:], []byte("msg-1"))
	var relayer [20]byte
	copy(relayer[:], []byte("relayer-1"))

	assert.True(t, rr.Record(id, relayer))
	assert.False(t, rr.Record(id, relayer)) // duplicate not allowed
	assert.True(t, rr.HasRelay(id))
}
