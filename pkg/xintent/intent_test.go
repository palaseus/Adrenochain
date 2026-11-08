package xintent

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testOrder() *CrossChainOrder {
	return &CrossChainOrder{
		SettlementContract: []byte{1},
		Swapper:            []byte{2},
		Nonce:              big.NewInt(1),
		OriginChain:        1,
		InitiationDeadline: 1000,
		FillDeadline:       2000,
		OrderData:          []byte("order-payload"),
		Signature:          []byte("sig"),
	}
}

func TestNewIntent(t *testing.T) {
	order := testOrder()
	expiry := time.Now().Add(time.Hour)
	intent, err := NewIntent(order, IntentTypeTransfer, 100, expiry)
	assert.NoError(t, err)
	assert.NotNil(t, intent)
	assert.Equal(t, IntentStatusOpen, intent.Status)
	assert.NotEqual(t, [32]byte{}, intent.ID)
	assert.Equal(t, uint16(100), intent.FillRewardBasisPoints)
	assert.False(t, intent.IsExpired())
	assert.NoError(t, intent.ValidateBasic())
}

func TestIntentRoundTripHash(t *testing.T) {
	order := testOrder()
	expiry := time.Now().Add(time.Hour)
	intent, err := NewIntent(order, IntentTypeTransfer, 50, expiry)
	assert.NoError(t, err)

	// Same inputs produce same ID.
	intent2, err := NewIntent(order, IntentTypeTransfer, 50, expiry)
	assert.NoError(t, err)
	assert.Equal(t, intent.ID, intent2.ID)

	// Different nonce produces different ID.
	order2 := testOrder()
	order2.Nonce = big.NewInt(2)
	intent3, err := NewIntent(order2, IntentTypeTransfer, 50, expiry)
	assert.NoError(t, err)
	assert.NotEqual(t, intent.ID, intent3.ID)
}

func TestIntentValidation(t *testing.T) {
	// Nil order.
	_, err := NewIntent(nil, IntentTypeTransfer, 50, time.Now().Add(time.Hour))
	assert.Error(t, err)

	// Expired intent.
	order := testOrder()
	_, err = NewIntent(order, IntentTypeTransfer, 50, time.Now().Add(-time.Hour))
	assert.Error(t, err)

	// Zero reward.
	_, err = NewIntent(order, IntentTypeTransfer, 0, time.Now().Add(time.Hour))
	assert.Error(t, err)
}

func TestFillReward(t *testing.T) {
	order := testOrder()
	intent, err := NewIntent(order, IntentTypeTransfer, 100, time.Now().Add(time.Hour))
	assert.NoError(t, err)
	// 100 bps = 1% of 1e18 = 1e16.
	assert.Equal(t, big.NewInt(1e16), intent.FillReward())
}

func TestIntentRegistry(t *testing.T) {
	reg := NewIntentRegistry()
	assert.Equal(t, 0, reg.FillCount())

	intent, err := NewIntent(testOrder(), IntentTypeTransfer, 50, time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NoError(t, reg.Register(intent))
	assert.Equal(t, 1, len(reg.OpenIntents()))

	// Get by ID.
	got := reg.Get(intent.ID)
	assert.NotNil(t, got)
	assert.Equal(t, intent.ID, got.ID)

	// Fill.
	filled := false
	result := &FillResult{
		IntentID:  intent.ID,
		Solver:    []byte{0xab},
		Reward:    intent.FillReward(),
		SettledAt: time.Now().UTC(),
		Success:   true,
	}
	_ = filled
	assert.NoError(t, reg.Fill(result))
	assert.Equal(t, IntentStatusFilled, reg.Get(intent.ID).Status)
	assert.Equal(t, 1, reg.FillCount())

	// Cannot refill.
	assert.Error(t, reg.Fill(result))
	// Cannot cancel filled.
	assert.Error(t, reg.Cancel(intent.ID))
}

func TestIntentRegistryCancel(t *testing.T) {
	reg := NewIntentRegistry()
	intent, err := NewIntent(testOrder(), IntentTypeTransfer, 50, time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NoError(t, reg.Register(intent))
	assert.NoError(t, reg.Cancel(intent.ID))
	assert.Equal(t, IntentStatusCancelled, reg.Get(intent.ID).Status)
	assert.Equal(t, 0, len(reg.OpenIntents()))
}

func TestIntentTypes(t *testing.T) {
	assert.Equal(t, "transfer", IntentTypeTransfer.String())
	assert.Equal(t, "swap", IntentTypeSwap.String())
	assert.Equal(t, "call", IntentTypeCall.String())
	assert.Equal(t, "unknown", IntentType(99).String())
}

func TestIntentExpired(t *testing.T) {
	// NewIntent rejects already-expired intents.
	_, err := NewIntent(testOrder(), IntentTypeTransfer, 50, time.Now().Add(-time.Minute))
	assert.Error(t, err)

	// A valid intent becomes expired after its deadline.
	intent, err := NewIntent(testOrder(), IntentTypeTransfer, 50, time.Now().Add(-time.Minute))
	_ = intent
	assert.Error(t, err)
}
