// Package compliance provides a unified compliance engine that combines FATF
// Travel Rule validation (travelrule) with sanctions screening (sanctions) for
// cross-border and multi-jurisdiction operation.
//
// The ComplianceEngine evaluates a proposed transfer against all applicable
// regulatory regimes based on originator/beneficiary jurisdictions, producing
// a single actionable decision: ALLOW, BLOCK, or REVIEW.
//
// See pkg/compliance/README.md for the compliance architecture.

package compliance

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/palaseus/adrenochain/pkg/compliance/sanctions"
	"github.com/palaseus/adrenochain/pkg/compliance/travelrule"
)

// Decision is the compliance engine's actionable output.
type Decision int

const (
	// DecisionAllow means the transfer passes all compliance checks.
	DecisionAllow Decision = iota
	// DecisionBlock means the transfer must be blocked (sanctions match, etc.).
	DecisionBlock
	// DecisionReview means the transfer requires manual review (unhosted
	// wallet, enhanced due diligence, threshold proximity).
	DecisionReview
)

// String returns the decision as a human-readable string.
func (d Decision) String() string {
	names := map[Decision]string{
		DecisionAllow:  "ALLOW",
		DecisionBlock:  "BLOCK",
		DecisionReview: "REVIEW",
	}
	if n, ok := names[d]; ok {
		return n
	}
	return "UNKNOWN"
}

// Result is the full compliance evaluation result.
type Result struct {
	Decision   Decision
	Reason     string
	TravelRule *travelrule.ValidationResult
	Sanctions  []sanctions.Match
	// Jurisdictions evaluated.
	Jurisdictions []string
}

// Engine evaluates transfers against Travel Rule and sanctions rules.
type Engine struct {
	// SanctionsChecker performs sanctions screening.
	SanctionsChecker *sanctions.SanctionsChecker
	// DefaultJurisdiction is used when a counterparty's jurisdiction is unset.
	DefaultJurisdiction travelrule.Jurisdiction
	// ReviewThresholdProximity determines whether transfers within 10% of the
	// Travel Rule threshold are flagged for review.
	ReviewThresholdProximity bool
}

// NewEngine creates a compliance engine.
func NewEngine(sanctionsChecker *sanctions.SanctionsChecker) *Engine {
	return &Engine{
		SanctionsChecker:         sanctionsChecker,
		DefaultJurisdiction:      travelrule.FATFGlobal,
		ReviewThresholdProximity: true,
	}
}

// Evaluate evaluates a transfer under the compliance engine.
func (e *Engine) Evaluate(originatorJurisdiction, beneficiaryJurisdiction travelrule.Jurisdiction, tx *travelrule.Transfer) (*Result, error) {
	if tx == nil {
		return nil, errors.New("nil transfer")
	}

	result := &Result{
		Jurisdictions: []string{string(originatorJurisdiction), string(beneficiaryJurisdiction)},
	}

	// Sanctions screening: originator and beneficiary wallet addresses.
	if e.SanctionsChecker != nil {
		if tx.Originator != nil && tx.Originator.AccountNumber != "" {
			originMatches, err := e.SanctionsChecker.ScreenWallet(tx.Originator.AccountNumber)
			if err != nil {
				return nil, fmt.Errorf("screen originator: %w", err)
			}
			result.Sanctions = append(result.Sanctions, originMatches...)
		}
		if tx.Beneficiary != nil && tx.Beneficiary.AccountNumber != "" {
			beneficiaryMatches, err := e.SanctionsChecker.ScreenWallet(tx.Beneficiary.AccountNumber)
			if err != nil {
				return nil, fmt.Errorf("screen beneficiary: %w", err)
			}
			result.Sanctions = append(result.Sanctions, beneficiaryMatches...)
		}
	}

	// Any sanctions match -> BLOCK.
	if len(result.Sanctions) > 0 {
		result.Decision = DecisionBlock
		result.Reason = "sanctions match on originator or beneficiary"
		return result, nil
	}

	// Travel Rule: use the stricter jurisdiction's threshold.
	trValidator := travelrule.NewValidator(stricterJurisdiction(originatorJurisdiction, beneficiaryJurisdiction))
	trResult, err := trValidator.CheckTransfer(tx)
	if err != nil {
		return nil, fmt.Errorf("travel rule check: %w", err)
	}
	result.TravelRule = trResult

	if !trResult.Valid {
		result.Decision = DecisionBlock
		result.Reason = trResult.Reason
		return result, nil
	}

	// Threshold proximity review: amounts within 90-100% of the threshold are
	// flagged regardless of whether the Transfer Rule strictly applies.
	if e.ReviewThresholdProximity && trResult.ThresholdAmount != nil && tx.Asset.Amount != nil {
		percent := new(big.Float).Quo(
			new(big.Float).SetInt(tx.Asset.Amount),
			new(big.Float).SetInt(trResult.ThresholdAmount),
		)
		if percent.Cmp(big.NewFloat(0.9)) >= 0 && percent.Cmp(big.NewFloat(1.0)) <= 0 {
			result.Decision = DecisionReview
			result.Reason = "amount within 90-100% of Travel Rule threshold"
			return result, nil
		}
	}

	// Unhosted wallet -> review (enhanced due diligence).
	if tx.Beneficiary.IsUnhosted() || tx.Originator.IsUnhosted() {
		if result.Decision == DecisionAllow {
			result.Decision = DecisionReview
			result.Reason = "unhosted counterparty: enhanced due diligence required"
		}
		return result, nil
	}

	result.Decision = DecisionAllow
	result.Reason = "compliance checks passed"
	return result, nil
}

// stricterJurisdiction returns the jurisdiction with the lower (or no)
// threshold, i.e. the one that imposes more Travel Rule obligations.
func stricterJurisdiction(a, b travelrule.Jurisdiction) travelrule.Jurisdiction {
	amtA, _, okA := a.Threshold()
	amtB, _, okB := b.Threshold()

	// No threshold = strictest (applies to all amounts).
	if !okA && okB {
		return a
	}
	if okA && !okB {
		return b
	}
	if !okA && !okB {
		return a
	}
	if amtA == nil && amtB != nil {
		return a
	}
	if amtA != nil && amtB == nil {
		return b
	}
	if amtA != nil && amtB != nil && amtA.Cmp(amtB) < 0 {
		return a
	}
	return b
}
