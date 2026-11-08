package compliance

import (
	"math/big"
	"testing"

	"github.com/palaseus/adrenochain/pkg/compliance/sanctions"
	"github.com/palaseus/adrenochain/pkg/compliance/travelrule"
	"github.com/stretchr/testify/assert"
)

func asset(amount int64) *travelrule.CryptoAsset {
	return &travelrule.CryptoAsset{Symbol: "BTC", Amount: big.NewInt(amount), Decimals: 8}
}

func baseTransfer() *travelrule.Transfer {
	return &travelrule.Transfer{
		Asset:       asset(50000000),
		Originator:  &travelrule.IVMS101Originator{Name: "Alice", AccountNumber: "0xalice", OriginatorVASP: "vasp-a"},
		Beneficiary: &travelrule.IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob", BeneficiaryVASP: "vasp-b"},
	}
}

func TestEngineAllow(t *testing.T) {
	sc := sanctions.NewSanctionsChecker()
	e := NewEngine(sc)
	tx := baseTransfer()

	res, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionAllow, res.Decision)
	assert.Contains(t, res.Reason, "passed")
}

func TestEngineBlockSanctions(t *testing.T) {
	sc := sanctions.NewSanctionsChecker()
	_ = sc.LoadList(sanctions.ListOFACSDN, []sanctions.Entry{
		{ID: "1", List: sanctions.ListOFACSDN, Name: "Bad", WalletAddresses: []string{"0xbob"}},
	})
	e := NewEngine(sc)
	tx := baseTransfer()

	res, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionBlock, res.Decision)
	assert.Contains(t, res.Reason, "sanctions")
}

func TestEngineBlockInvalidTravelRule(t *testing.T) {
	sc := sanctions.NewSanctionsChecker()
	e := NewEngine(sc)
	tx := &travelrule.Transfer{Asset: asset(50000000)} // missing identity

	res, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionBlock, res.Decision)
}

func TestEngineReviewUnhosted(t *testing.T) {
	sc := sanctions.NewSanctionsChecker()
	e := NewEngine(sc)
	tx := baseTransfer()
	tx.Beneficiary.BeneficiaryVASP = "" // unhosted

	res, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionReview, res.Decision)
}

func TestEngineReviewThresholdProximity(t *testing.T) {
	sc := sanctions.NewSanctionsChecker()
	e := NewEngine(sc)
	// US threshold is 3000 USD; use 2800 (93%).
	tx := baseTransfer()
	tx.Asset.Amount = big.NewInt(2800)

	res, err := e.Evaluate(travelrule.UnitedStates, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionReview, res.Decision)
	assert.Contains(t, res.Reason, "90-100%")
}

func TestEngineNoSanctionsChecker(t *testing.T) {
	e := NewEngine(nil)
	tx := baseTransfer()

	res, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, tx)
	assert.NoError(t, err)
	assert.Equal(t, DecisionAllow, res.Decision)
}

func TestEngineNilTransfer(t *testing.T) {
	e := NewEngine(sanctions.NewSanctionsChecker())
	_, err := e.Evaluate(travelrule.EU, travelrule.UnitedStates, nil)
	assert.Error(t, err)
}
