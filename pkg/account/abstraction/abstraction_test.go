package abstraction

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type alwaysValidAccount struct{}

func (a *alwaysValidAccount) ValidateUserOp(uo *UserOperation, caller []byte) (*ValidationResult, error) {
	return &ValidationResult{Valid: true, GasUsed: big.NewInt(21000)}, nil
}

func TestUserOperationHashAndBasic(t *testing.T) {
	uo := &UserOperation{
		Sender:               []byte{0xab},
		Nonce:                big.NewInt(1),
		CallGasLimit:         big.NewInt(100000),
		VerificationGasLimit: big.NewInt(50000),
		PreVerificationGas:   big.NewInt(10000),
		MaxFeePerGas:         big.NewInt(100),
		MaxPriorityFeePerGas: big.NewInt(20),
	}
	assert.NoError(t, uo.ValidateBasic())
	assert.NotEmpty(t, uo.Hash())

	// priority fee > max fee should fail.
	uo2 := &UserOperation{Sender: []byte{0xab}, Nonce: big.NewInt(1),
		MaxFeePerGas: big.NewInt(10), MaxPriorityFeePerGas: big.NewInt(20)}
	assert.Error(t, uo2.ValidateBasic())
}

func TestEntryPointValidate(t *testing.T) {
	ep := NewEntryPoint()
	acct := &alwaysValidAccount{}
	ep.RegisterAccount([]byte("alice"), acct)

	uo := &UserOperation{
		Sender:               []byte("alice"),
		Nonce:                big.NewInt(1),
		CallGasLimit:         big.NewInt(100000),
		VerificationGasLimit: big.NewInt(50000),
		PreVerificationGas:   big.NewInt(10000),
		MaxFeePerGas:         big.NewInt(100),
		MaxPriorityFeePerGas: big.NewInt(20),
	}
	res, err := ep.ValidateUserOp(uo)
	assert.NoError(t, err)
	assert.True(t, res.Valid)
	assert.NotNil(t, res.GasUsed)

	// unknown sender
	uo2 := &UserOperation{Sender: []byte("bob"), Nonce: big.NewInt(1),
		CallGasLimit: big.NewInt(100000), VerificationGasLimit: big.NewInt(50000),
		PreVerificationGas: big.NewInt(10000), MaxFeePerGas: big.NewInt(100),
		MaxPriorityFeePerGas: big.NewInt(20)}
	res2, _ := ep.ValidateUserOp(uo2)
	assert.False(t, res2.Valid)
	assert.Contains(t, res2.Reason, "not registered")
}

func TestPaymaster(t *testing.T) {
	pm := &SponsoredPaymaster{Senders: map[string]bool{"alice": true}}
	uo := &UserOperation{Sender: []byte("alice"), Nonce: big.NewInt(1)}
	res, err := pm.ValidatePaymasterUserOp(uo)
	assert.NoError(t, err)
	assert.True(t, res.Valid)

	uo2 := &UserOperation{Sender: []byte("bob"), Nonce: big.NewInt(1)}
	res2, _ := pm.ValidatePaymasterUserOp(uo2)
	assert.False(t, res2.Valid)
}

func TestAuthorization(t *testing.T) {
	auth := &Authorization{
		ContractAddress: []byte{0xcd},
		ChainID:         big.NewInt(1),
		Nonce:           big.NewInt(5),
	}
	assert.NotEmpty(t, auth.SignHash())
}

func TestGasEstimate(t *testing.T) {
	est := EstimateGas(big.NewInt(10000), big.NewInt(50000), big.NewInt(21000))
	assert.Equal(t, big.NewInt(81000), est.Total)
	fee := MaxFeeForGas(est.Total, big.NewInt(100))
	assert.Equal(t, big.NewInt(8100000), fee)
}
