// Package rwa implements Real-World Asset tokenization with embedded
// compliance, inspired by ERC-3643 (T-REX, Final, Dec 2023) — the dominant
// compliant security-token standard for institutional RWA issuance.
//
// Unlike plain ERC-20 tokens, RWA tokens embed compliance directly into
// transfers: every transfer is gated on on-chain identity (ONCHAINID),
// investor eligibility, and offering-level rules. A legal wrapper (SPV/trust)
// holds the underlying asset; a custodian and auditor attest that token
// supply matches reserves (proof of reserve).
//
// See pkg/rwa/README.md for design, standards alignment, and usage.

package rwa

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// TokenVersion is the RWA token contract version.
	TokenVersion = 1

	// MaxClaimTopics is the maximum number of identity claim topics.
	MaxClaimTopics = 32

	// DefaultTransferFeeBasisPoints is the default transfer fee (0 bps).
	DefaultTransferFeeBasisPoints = 0
)

// TokenStatus is the lifecycle status of an RWA token.
type TokenStatus int

const (
	// TokenStatusDraft is a pre-issuance token (no transfers allowed).
	TokenStatusDraft TokenStatus = iota
	// TokenStatusIssued is a live, transferable token.
	TokenStatusIssued
	// TokenStatusPaused halts all transfers (emergency/compliance).
	TokenStatusPaused
	// TokenStatusRedeemed is a redeemed/retired token.
	TokenStatusRedeemed
)

// String returns the token status name.
func (s TokenStatus) String() string {
	names := map[TokenStatus]string{
		TokenStatusDraft:    "draft",
		TokenStatusIssued:   "issued",
		TokenStatusPaused:   "paused",
		TokenStatusRedeemed: "redeemed",
	}
	if n, ok := names[s]; ok {
		return n
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// Identity (ONCHAINID aligned)
// ---------------------------------------------------------------------------

// Identity is an on-chain identity identifier (ONCHAINID aligned).
type Identity [32]byte

// NewIdentity creates an identity from raw bytes (padded/truncated to 32).
func NewIdentity(b []byte) Identity {
	var id Identity
	copy(id[:], b)
	return id
}

// IsZero reports whether the identity is all zeros.
func (id Identity) IsZero() bool {
	return id == Identity{}
}

// Equal compares two identities.
func (id Identity) Equal(other Identity) bool {
	return id == other
}

// ClaimTopic is a category of identity claim (e.g. KYC, accredited investor).
type ClaimTopic uint8

const (
	// ClaimTopicKYC is a know-your-customer claim.
	ClaimTopicKYC ClaimTopic = 1
	// ClaimTopicAccreditedInvestor is an accredited-investor claim.
	ClaimTopicAccreditedInvestor ClaimTopic = 2
	// ClaimTopicSanctionsClear is a sanctions-screening-clear claim.
	ClaimTopicSanctionsClear ClaimTopic = 3
	// ClaimTopicQualifiedInvestor is a qualified-investor claim.
	ClaimTopicQualifiedInvestor ClaimTopic = 4
)

// String returns the claim topic name.
func (ct ClaimTopic) String() string {
	names := map[ClaimTopic]string{
		ClaimTopicKYC:                "kyc",
		ClaimTopicAccreditedInvestor: "accredited_investor",
		ClaimTopicSanctionsClear:     "sanctions_clear",
		ClaimTopicQualifiedInvestor:  "qualified_investor",
	}
	if n, ok := names[ct]; ok {
		return n
	}
	return fmt.Sprintf("topic_%d", uint8(ct))
}

// IdentityClaim is a single claim on an identity, issued by a claim issuer.
type IdentityClaim struct {
	// Topic is the claim category.
	Topic ClaimTopic
	// Issuer is the identity of the claim issuer (e.g. KYC provider).
	Issuer Identity
	// Value is the claim value (opaque, e.g. attestation reference).
	Value []byte
	// IssuedAt is when the claim was issued.
	IssuedAt time.Time
	// ExpiresAt is when the claim expires (zero = no expiry).
	ExpiresAt time.Time
}

// IsValid reports whether the claim is currently valid (not expired).
func (ic *IdentityClaim) IsValid() bool {
	if ic == nil {
		return false
	}
	if !ic.ExpiresAt.IsZero() && time.Now().UTC().After(ic.ExpiresAt) {
		return false
	}
	return true
}

// IdentityRegistry stores identities and their claims.
type IdentityRegistry struct {
	mu      sync.RWMutex
	claims  map[Identity][]*IdentityClaim
	issuers map[Identity]bool // registered claim issuers
}

// NewIdentityRegistry creates an empty identity registry.
func NewIdentityRegistry() *IdentityRegistry {
	return &IdentityRegistry{
		claims:  make(map[Identity][]*IdentityClaim),
		issuers: make(map[Identity]bool),
	}
}

// RegisterIssuer registers a claim issuer (e.g. a KYC provider).
func (ir *IdentityRegistry) RegisterIssuer(issuer Identity) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.issuers[issuer] = true
}

// IsRegisteredIssuer reports whether the issuer is registered.
func (ir *IdentityRegistry) IsRegisteredIssuer(issuer Identity) bool {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return ir.issuers[issuer]
}

// AddClaim adds (or replaces) a claim for an identity.
func (ir *IdentityRegistry) AddClaim(identity Identity, claim *IdentityClaim) error {
	if claim == nil {
		return errors.New("nil claim")
	}
	if !ir.IsRegisteredIssuer(claim.Issuer) {
		return errors.New("unregistered claim issuer")
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	// Replace existing claim for the same topic.
	existing := ir.claims[identity]
	found := false
	for i, c := range existing {
		if c.Topic == claim.Topic {
			existing[i] = claim
			found = true
			break
		}
	}
	if !found {
		ir.claims[identity] = append(existing, claim)
	}
	return nil
}

// GetClaims returns all valid claims for an identity.
func (ir *IdentityRegistry) GetClaims(identity Identity) []*IdentityClaim {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	claims := ir.claims[identity]
	var valid []*IdentityClaim
	for _, c := range claims {
		if c.IsValid() {
			valid = append(valid, c)
		}
	}
	return valid
}

// HasClaimTopic reports whether the identity has a valid claim on the topic
// issued by a registered issuer.
func (ir *IdentityRegistry) HasClaimTopic(identity Identity, topic ClaimTopic) bool {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	for _, c := range ir.claims[identity] {
		if c.Topic == topic && c.IsValid() && ir.issuers[c.Issuer] {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Compliance engine
// ---------------------------------------------------------------------------

// ComplianceRule is a transfer-compliance predicate.
type ComplianceRule func(token *Token, from, to Identity, amount *big.Int) error

// ComplianceEngine evaluates transfer-compliance rules (ERC-3643 aligned).
type ComplianceEngine struct {
	mu       sync.RWMutex
	rules    []ComplianceRule
	token    *Token
	registry *IdentityRegistry
}

// NewComplianceEngine creates a compliance engine for a token.
func NewComplianceEngine(token *Token, registry *IdentityRegistry) *ComplianceEngine {
	return &ComplianceEngine{
		token:    token,
		registry: registry,
	}
}

// AddRule adds a compliance rule.
func (ce *ComplianceEngine) AddRule(rule ComplianceRule) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.rules = append(ce.rules, rule)
}

// CanTransfer evaluates whether a transfer is compliant.
func (ce *ComplianceEngine) CanTransfer(from, to Identity, amount *big.Int) error {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if ce.token.Status != TokenStatusIssued {
		return errors.New("token not issued")
	}
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("invalid transfer amount")
	}
	if from.IsZero() || to.IsZero() {
		return errors.New("zero identity")
	}

	// Sender must have KYC.
	if !ce.registry.HasClaimTopic(from, ClaimTopicKYC) {
		return errors.New("sender lacks valid KYC claim")
	}
	// Receiver must have KYC.
	if !ce.registry.HasClaimTopic(to, ClaimTopicKYC) {
		return errors.New("receiver lacks valid KYC claim")
	}
	// Receiver must not be sanctioned.
	if ce.registry.HasClaimTopic(to, ClaimTopicSanctionsClear) == false {
		// Sanctions clear is a positive claim; absence means blocked.
		return errors.New("receiver lacks sanctions clearance")
	}

	for _, rule := range ce.rules {
		if err := rule(ce.token, from, to, amount); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Token
// ---------------------------------------------------------------------------

// Token is an RWA security token with embedded compliance.
type Token struct {
	mu                     sync.RWMutex
	Version                uint8
	Symbol                 string
	Name                   string
	Decimals               uint8
	Status                 TokenStatus
	TotalSupply            *big.Int
	Balances               map[Identity]*big.Int
	Owner                  Identity // issuer/owner identity
	Compliance             *ComplianceEngine
	Registry               *IdentityRegistry
	TransferFeeBasisPoints uint16
	Asset                  *Asset // linked real-world asset
}

// NewToken creates a new RWA token in draft status.
func NewToken(symbol, name string, decimals uint8, owner Identity, registry *IdentityRegistry) *Token {
	t := &Token{
		Version:                TokenVersion,
		Symbol:                 symbol,
		Name:                   name,
		Decimals:               decimals,
		Status:                 TokenStatusDraft,
		TotalSupply:            big.NewInt(0),
		Balances:               make(map[Identity]*big.Int),
		Owner:                  owner,
		Registry:               registry,
		TransferFeeBasisPoints: DefaultTransferFeeBasisPoints,
	}
	t.Compliance = NewComplianceEngine(t, registry)
	return t
}

// Issue transitions the token from draft to issued.
func (t *Token) Issue(caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !caller.Equal(t.Owner) {
		return errors.New("only owner can issue")
	}
	if t.Status != TokenStatusDraft {
		return errors.New("token already issued")
	}
	t.Status = TokenStatusIssued
	return nil
}

// Pause halts all transfers.
func (t *Token) Pause(caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !caller.Equal(t.Owner) {
		return errors.New("only owner can pause")
	}
	t.Status = TokenStatusPaused
	return nil
}

// Unpause resumes transfers.
func (t *Token) Unpause(caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !caller.Equal(t.Owner) {
		return errors.New("only owner can unpause")
	}
	if t.Status != TokenStatusPaused {
		return errors.New("token not paused")
	}
	t.Status = TokenStatusIssued
	return nil
}

// Mint mints tokens to an identity (owner-only).
func (t *Token) Mint(to Identity, amount *big.Int, caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !caller.Equal(t.Owner) {
		return errors.New("only owner can mint")
	}
	if t.Status != TokenStatusIssued {
		return errors.New("token not issued")
	}
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("invalid mint amount")
	}
	t.Balances[to] = new(big.Int).Add(balanceOrZero(t.Balances[to]), amount)
	t.TotalSupply = new(big.Int).Add(t.TotalSupply, amount)
	return nil
}

// balanceOrZero returns b if non-nil, else zero.
func balanceOrZero(b *big.Int) *big.Int {
	if b == nil {
		return big.NewInt(0)
	}
	return b
}

// Burn burns tokens from an identity (owner-only).
func (t *Token) Burn(from Identity, amount *big.Int, caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !caller.Equal(t.Owner) {
		return errors.New("only owner can burn")
	}
	if t.Status != TokenStatusIssued {
		return errors.New("token not issued")
	}
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("invalid burn amount")
	}
	bal := t.Balances[from]
	if bal == nil || bal.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}
	t.Balances[from] = new(big.Int).Sub(bal, amount)
	t.TotalSupply = new(big.Int).Sub(t.TotalSupply, amount)
	return nil
}

// Transfer moves tokens between identities, subject to compliance.
func (t *Token) Transfer(from, to Identity, amount *big.Int, caller Identity) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status != TokenStatusIssued {
		return errors.New("token not issued")
	}
	if amount == nil || amount.Sign() <= 0 {
		return errors.New("invalid transfer amount")
	}
	if from.IsZero() || to.IsZero() {
		return errors.New("zero identity")
	}

	// Compliance check (ERC-3643: gated on on-chain identity).
	if err := t.Compliance.CanTransfer(from, to, amount); err != nil {
		return err
	}

	bal := t.Balances[from]
	if bal == nil || bal.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}

	// Transfer fee (bps of amount).
	fee := new(big.Int).Mul(amount, big.NewInt(int64(t.TransferFeeBasisPoints)))
	fee = new(big.Int).Div(fee, big.NewInt(10000))
	netAmount := new(big.Int).Sub(amount, fee)

	t.Balances[from] = new(big.Int).Sub(bal, amount)
	t.Balances[to] = new(big.Int).Add(balanceOrZero(t.Balances[to]), netAmount)
	// Fee accrues to protocol (could be a fee wallet).
	_ = fee

	return nil
}

// BalanceOf returns the token balance of an identity.
func (t *Token) BalanceOf(id Identity) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if bal, ok := t.Balances[id]; ok {
		return new(big.Int).Set(bal)
	}
	return big.NewInt(0)
}

// TotalSupplyOf returns the total supply.
func (t *Token) TotalSupplyOf() *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return new(big.Int).Set(t.TotalSupply)
}

// ---------------------------------------------------------------------------
// Asset (real-world asset linked to the token)
// ---------------------------------------------------------------------------

// AssetType is the class of real-world asset.
type AssetType int

const (
	// AssetTypeTreasury is a tokenized treasury bill / government bond.
	AssetTypeTreasury AssetType = iota
	// AssetTypeRealEstate is tokenized real estate.
	AssetTypeRealEstate
	// AssetTypePrivateCredit is tokenized private credit / loans.
	AssetTypePrivateCredit
	// AssetTypeCommodity is tokenized gold / commodities.
	AssetTypeCommodity
	// AssetTypeEquity is tokenized equity / shares.
	AssetTypeEquity
	// AssetTypeFund is a tokenized fund share.
	AssetTypeFund
)

// String returns the asset type name.
func (at AssetType) String() string {
	names := map[AssetType]string{
		AssetTypeTreasury:      "treasury",
		AssetTypeRealEstate:    "real_estate",
		AssetTypePrivateCredit: "private_credit",
		AssetTypeCommodity:     "commodity",
		AssetTypeEquity:        "equity",
		AssetTypeFund:          "fund",
	}
	if n, ok := names[at]; ok {
		return n
	}
	return fmt.Sprintf("asset_%d", uint8(at))
}

// Asset links a token to its underlying real-world asset.
type Asset struct {
	// Type is the asset class.
	Type AssetType
	// Description is a human-readable asset description.
	Description string
	// Custodian holds the underlying asset.
	Custodian Identity
	// Auditor performs reserve/NAV attestations.
	Auditor Identity
	// LegalWrapper is the SPV/trust legal entity.
	LegalWrapper string
	// Jurisdiction is the governing legal jurisdiction.
	Jurisdiction string
	// NAV is the net asset value per token (minor units).
	NAV *big.Int
	// LastAttestation is when the auditor last attested reserves.
	LastAttestation time.Time
}

// ---------------------------------------------------------------------------
// ProofOfReserve
// ---------------------------------------------------------------------------

// ProofOfReserve is an auditor attestation that token supply matches reserves.
type ProofOfReserve struct {
	// Auditor is the attesting auditor identity.
	Auditor Identity
	// TokenSupply is the on-chain token supply at attestation time.
	TokenSupply *big.Int
	// ReserveValue is the verified off-chain reserve value.
	ReserveValue *big.Int
	// AttestedAt is when the attestation was made.
	AttestedAt time.Time
	// ReportReference is an off-chain report reference (URL/hash).
	ReportReference string
}

// IsValid reports whether the proof indicates supply matches reserves.
func (por *ProofOfReserve) IsValid() bool {
	if por == nil {
		return false
	}
	if por.TokenSupply == nil || por.ReserveValue == nil {
		return false
	}
	return por.TokenSupply.Cmp(por.ReserveValue) == 0 && por.TokenSupply.Sign() > 0
}
