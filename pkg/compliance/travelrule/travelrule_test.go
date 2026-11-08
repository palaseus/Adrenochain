package travelrule

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func asset(amount int64) *CryptoAsset {
	return &CryptoAsset{Symbol: "BTC", Amount: big.NewInt(amount), Decimals: 8}
}

func TestThresholds(t *testing.T) {
	// EU TFR: applies, no threshold.
	amt, cur, ok := EU.Threshold()
	assert.True(t, ok)
	assert.Nil(t, amt)
	assert.Equal(t, "", cur)

	// US FinCEN: > $3,000.
	amt, cur, ok = UnitedStates.Threshold()
	assert.True(t, ok)
	assert.Equal(t, int64(3000), amt.Int64())
	assert.Equal(t, "USD", cur)
}

func TestTransferComplete(t *testing.T) {
	v := NewValidator(EU)
	tx := &Transfer{
		TransferID:  "tx-1",
		Asset:       asset(50000000), // 0.5 BTC
		Originator:  &IVMS101Originator{Name: "Alice", AccountNumber: "0xalice"},
		Beneficiary: &IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob", BeneficiaryVASP: "vasp-b"},
	}
	res, err := v.CheckTransfer(tx)
	assert.NoError(t, err)
	assert.True(t, res.Valid)
	assert.True(t, res.TravelRuleApplies)
	assert.Contains(t, res.Reason, "complete")
}

func TestBelowThresholdUS(t *testing.T) {
	v := NewValidator(UnitedStates)
	tx := &Transfer{
		Asset:       asset(100), // $100 equivalent, below $3,000
		Originator:  &IVMS101Originator{Name: "Alice", AccountNumber: "0xalice"},
		Beneficiary: &IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob"},
	}
	res, err := v.CheckTransfer(tx)
	assert.NoError(t, err)
	assert.True(t, res.Valid)
	assert.False(t, res.TravelRuleApplies)
	assert.Contains(t, res.Reason, "threshold")
}

func TestMissingIdentity(t *testing.T) {
	v := NewValidator(EU)
	tx := &Transfer{Asset: asset(50000000)}
	res, err := v.CheckTransfer(tx)
	assert.NoError(t, err)
	assert.False(t, res.Valid)
	assert.Contains(t, res.Reason, "originator")
}

func TestSanctioned(t *testing.T) {
	v := NewValidator(EU)
	v.SanctionedAddresses["0xbad"] = true
	tx := &Transfer{
		Asset:       asset(50000000),
		Originator:  &IVMS101Originator{Name: "Bad", AccountNumber: "0xbad"},
		Beneficiary: &IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob"},
	}
	res, err := v.CheckTransfer(tx)
	assert.NoError(t, err)
	assert.False(t, res.Valid)
	assert.Contains(t, res.Reason, "sanctioned")
}

func TestUnhostedBeneficiary(t *testing.T) {
	v := NewValidator(EU)
	tx := &Transfer{
		Asset:       asset(50000000),
		Originator:  &IVMS101Originator{Name: "Alice", AccountNumber: "0xalice", OriginatorVASP: "vasp-a"},
		Beneficiary: &IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob"}, // no VASP
	}
	res, err := v.CheckTransfer(tx)
	assert.NoError(t, err)
	assert.True(t, res.Valid)
	assert.True(t, tx.Beneficiary.IsUnhosted())
	assert.Contains(t, res.Reason, "unhosted")
}

func TestFormatTransferID(t *testing.T) {
	id := FormatTransferID("tx", "BTC", big.NewInt(12345), 1000)
	assert.Contains(t, id, "BTC")
	assert.Contains(t, id, "12345")
}
