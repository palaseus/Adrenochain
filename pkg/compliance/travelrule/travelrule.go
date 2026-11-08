// Package travelrule implements FATF Recommendation 16 (Travel Rule) data
// structures and validation aligned with the InterVASP Messaging Standard
// (IVMS 101), the EU Transfer of Funds Regulation (TFR, Reg. (EU) 2023/1113),
// and MiCA CASP obligations as of the post-transitional-period environment
// (MiCA transitional period ended 1 July 2026; no minimum threshold in the EU).
//
// It provides originator/beneficiary identity models, threshold checks,
// unhosted-wallet handling, and transfer validation. It does not perform
// network messaging; a VASP integrates this with its chosen Travel Rule
// solution provider.
//
// See pkg/compliance/travelrule/README.md for regulatory context.

package travelrule

import (
	"errors"
	"fmt"
	"math/big"
)

// ---------------------------------------------------------------------------
// Jurisdiction & threshold
// ---------------------------------------------------------------------------

// Jurisdiction identifies a regulatory jurisdiction for Travel Rule purposes.
type Jurisdiction string

const (
	FATFGlobal    Jurisdiction = "FATF"
	EU            Jurisdiction = "EU" // TFR: no minimum threshold
	UnitedStates  Jurisdiction = "US" // FinCEN: > USD/EUR 3,000
	UnitedKingdom Jurisdiction = "UK" // FCA
	Singapore     Jurisdiction = "SG" // MAS
	UAE           Jurisdiction = "AE" // VARA
	Japan         Jurisdiction = "JP" // FSA: no minimum threshold for crypto
	Korea         Jurisdiction = "KR" // FIU: > KRW 1 million (approx USD 750)
	India         Jurisdiction = "IN" // FIU-IND: all virtual asset transfers
	Brazil        Jurisdiction = "BR" // Bacen/CVMB: no minimum threshold
	Canada        Jurisdiction = "CA" // FINTRAC: > CAD 10,000
	Australia     Jurisdiction = "AU" // AUSTRAC: > AUD 10,000
	Switzerland   Jurisdiction = "CH" // FINMA: no minimum threshold
)

// Threshold returns the Travel Rule value threshold for a jurisdiction.
func (j Jurisdiction) Threshold() (amount *big.Int, currency string, ok bool) {
	switch j {
	case EU:
		// TFR applies to all crypto transfers, no minimum threshold.
		return nil, "", true
	case UnitedStates:
		return big.NewInt(3000), "USD", true
	case UnitedKingdom:
		return big.NewInt(1000), "GBP", true
	case Singapore:
		return big.NewInt(1500), "SGD", true
	case UAE:
		return nil, "", true // VARA: all virtual asset transfers
	case Japan:
		return nil, "", true // FSA: no minimum threshold for crypto transfers
	case Korea:
		return big.NewInt(1000000), "KRW", true // FIU: > KRW 1 million
	case India:
		return nil, "", true // FIU-IND: all virtual asset transfers
	case Brazil:
		return nil, "", true // Bacen: all virtual asset transfers
	case Canada:
		return big.NewInt(10000), "CAD", true // FINTRAC: > CAD 10,000
	case Australia:
		return big.NewInt(10000), "AUD", true // AUSTRAC: > AUD 10,000
	case Switzerland:
		return nil, "", true // FINMA: no minimum threshold
	case FATFGlobal:
		return big.NewInt(1000), "USD/EUR", true
	default:
		return nil, "", false
	}
}

// ---------------------------------------------------------------------------
// IVMS 101 identity data
// ---------------------------------------------------------------------------

// IVMS101NaturalPerson is an individual originator/beneficiary per IVMS 101.
type IVMS101NaturalPerson struct {
	// Personal details
	Name             string `json:"name"`
	NationalIDNumber string `json:"national_id_number,omitempty"`
	NationalIDType   string `json:"national_id_type,omitempty"`
	Country          string `json:"country,omitempty"`
	BirthDate        string `json:"birth_date,omitempty"`
	BirthPlace       string `json:"birth_place,omitempty"`
	Address          string `json:"address,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	PostalCode       string `json:"postal_code,omitempty"`
}

// IVMS101LegalPerson is a legal entity per IVMS 101.
type IVMS101LegalPerson struct {
	Name                string `json:"name"`
	RegistrationNumber  string `json:"registration_number,omitempty"`
	RegistrationCountry string `json:"registration_country,omitempty"`
	Address             string `json:"address,omitempty"`
	City                string `json:"city,omitempty"`
	State               string `json:"state,omitempty"`
	PostalCode          string `json:"postal_code,omitempty"`
}

// IVMS101Originator is the transfer originator.
type IVMS101Originator struct {
	// Required: name and account number / wallet address.
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"` // wallet address or account
	// Optional IVMS 101 fields.
	Person        *IVMS101NaturalPerson `json:"person,omitempty"`
	LegalPerson   *IVMS101LegalPerson   `json:"legal_person,omitempty"`
	WalletAddress string                `json:"wallet_address,omitempty"`
	// For the originator VASP.
	OriginatorVASP string `json:"originator_vasp,omitempty"`
}

// IVMS101Beneficiary is the transfer beneficiary.
type IVMS101Beneficiary struct {
	Name            string                `json:"name"`
	AccountNumber   string                `json:"account_number"`
	Person          *IVMS101NaturalPerson `json:"person,omitempty"`
	LegalPerson     *IVMS101LegalPerson   `json:"legal_person,omitempty"`
	WalletAddress   string                `json:"wallet_address,omitempty"`
	BeneficiaryVASP string                `json:"beneficiary_vasp,omitempty"`
}

// IsUnhosted reports whether the account appears to be an unhosted wallet
// (self-custodied, no VASP).
func (b *IVMS101Beneficiary) IsUnhosted() bool {
	return b.BeneficiaryVASP == ""
}

// IsUnhostedOriginator reports whether the originator is unhosted.
func (o *IVMS101Originator) IsUnhosted() bool {
	return o.OriginatorVASP == ""
}

// ---------------------------------------------------------------------------
// Transfer & validation
// ---------------------------------------------------------------------------

// CryptoAsset identifies the virtual asset being transferred.
type CryptoAsset struct {
	Symbol   string   // e.g. BTC, ETH, USDT
	Address  string   // contract address for tokens, empty for native
	Chain    string   // e.g. ethereum, bitcoin
	Amount   *big.Int // minor units (wei/satoshis)
	Decimals uint8
}

// Transfer is a virtual asset transfer subject to the Travel Rule.
type Transfer struct {
	TransferID        string
	Asset             *CryptoAsset
	Originator        *IVMS101Originator
	Beneficiary       *IVMS101Beneficiary
	Timestamp         int64 // unix seconds
	TravelRuleApplies bool
}

// ValidationResult is the outcome of a Travel Rule check.
type ValidationResult struct {
	Valid             bool
	TravelRuleApplies bool
	Reason            string
	ThresholdAmount   *big.Int
	ThresholdCurrency string
}

// Validator checks transfers against Travel Rule rules for a jurisdiction.
type Validator struct {
	Jurisdiction Jurisdiction
	// SanctionedAddresses is an optional sanctions list (hex addresses).
	SanctionedAddresses map[string]bool
}

// NewValidator creates a Travel Rule validator for a jurisdiction.
func NewValidator(j Jurisdiction) *Validator {
	return &Validator{
		Jurisdiction:        j,
		SanctionedAddresses: make(map[string]bool),
	}
}

// CheckTransfer validates a Transfer under the jurisdiction's Travel Rule.
func (v *Validator) CheckTransfer(t *Transfer) (*ValidationResult, error) {
	if t == nil {
		return nil, errors.New("nil transfer")
	}
	if t.Asset == nil || t.Asset.Amount == nil {
		return nil, errors.New("nil asset or amount")
	}
	if t.Originator == nil || t.Originator.Name == "" || t.Originator.AccountNumber == "" {
		return &ValidationResult{Valid: false, Reason: "missing originator identity"}, nil
	}
	if t.Beneficiary == nil || t.Beneficiary.Name == "" || t.Beneficiary.AccountNumber == "" {
		return &ValidationResult{Valid: false, Reason: "missing beneficiary identity"}, nil
	}

	amount, currency, ok := v.Jurisdiction.Threshold()
	result := &ValidationResult{
		ThresholdAmount:   amount,
		ThresholdCurrency: currency,
	}

	// Determine whether the Travel Rule applies.
	applies := true
	if v.Jurisdiction == EU || v.Jurisdiction == UAE || v.Jurisdiction == FATFGlobal {
		applies = true // TFR / VARA / FATF baseline: applies regardless of amount
	} else if ok && amount != nil {
		applies = t.Asset.Amount.Cmp(amount) >= 0
	}
	result.TravelRuleApplies = applies

	if !applies {
		result.Valid = true
		result.Reason = "below Travel Rule threshold"
		return result, nil
	}

	// Sanctions screening (originator + beneficiary).
	if v.SanctionedAddresses != nil {
		if v.SanctionedAddresses[t.Originator.AccountNumber] {
			return &ValidationResult{Valid: false, Reason: "originator sanctioned"}, nil
		}
		if v.SanctionedAddresses[t.Beneficiary.AccountNumber] {
			return &ValidationResult{Valid: false, Reason: "beneficiary sanctioned"}, nil
		}
	}

	// Unhosted-wallet handling: enhanced due diligence flagged.
	if t.Beneficiary.IsUnhosted() {
		result.Valid = true
		result.Reason = "unhosted beneficiary: enhanced due diligence required"
		return result, nil
	}

	result.Valid = true
	result.Reason = "Travel Rule data complete"
	return result, nil
}

// FormatTransferID builds a deterministic transfer ID.
func FormatTransferID(prefix, assetSymbol string, amount *big.Int, timestamp int64) string {
	return fmt.Sprintf("%s:%s:%s:%d", prefix, assetSymbol, amount.String(), timestamp)
}
