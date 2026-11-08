# `pkg/rwa` — Real-World Asset Tokenization (ERC-3643 / T-REX aligned)

Compliant real-world asset (RWA) tokenization with embedded on-chain compliance,
inspired by ERC-3643 (T-REX, Final, Dec 2023) — the dominant compliant
security-token standard for institutional RWA issuance ($38.6B+ tokenized RWA
market as of late 2026).

## The problem it solves

Plain ERC-20 tokens have no buyer-eligibility or transfer-restriction logic,
making them unsuitable for regulated assets (securities, Treasuries, real
estate, private credit). RWA tokens embed compliance directly into every
transfer: each transfer is gated on on-chain identity (ONCHAINID), investor
eligibility (KYC, accredited investor), sanctions clearance, and offering-level
rules.

## Components

| Component | Role |
|-----------|------|
| `Token` | RWA security token with draft/issued/paused/redeemed lifecycle, mint/burn/transfer |
| `IdentityRegistry` | ONCHAINID-aligned identity registry with claims (KYC, accredited, sanctions-clear) |
| `IdentityClaim` | Single identity claim issued by a registered claim issuer |
| `ComplianceEngine` | Transfer-compliance evaluation (sender/receiver KYC, sanctions, custom rules) |
| `Asset` | Link between token and underlying real-world asset (custodian, auditor, NAV, jurisdiction) |
| `ProofOfReserve` | Auditor attestation that on-chain token supply matches off-chain reserves |

## Compliance model (ERC-3643 aligned)

Every `Transfer` is gated:
1. Token must be `issued` (not draft/paused/redeemed)
2. Sender must have a valid KYC claim
3. Receiver must have a valid KYC claim **and** sanctions-clearance claim
4. Custom compliance rules (e.g. offering caps, holder limits, lockups) apply

A `ClaimTopic` system supports extensible identity claims beyond KYC.

## Example

```go
// 1. Set up identity registry and register a KYC provider
reg := rwa.NewIdentityRegistry()
kycProvider := rwa.NewIdentity([]byte{1})
reg.RegisterIssuer(kycProvider)

// 2. Create the RWA token
owner := rwa.NewIdentity([]byte{2})
token := rwa.NewToken("TKU", "Tokenized US Treasury", 18, owner, reg)
_ = token.Issue(owner)

// 3. Onboard an investor with KYC + sanctions clearance
investor := rwa.NewIdentity([]byte{5})
reg.AddClaim(investor, &rwa.IdentityClaim{
    Topic:     rwa.ClaimTopicKYC,
    Issuer:    kycProvider,
    Value:     []byte("kyc-attestation-1"),
    IssuedAt:  time.Now(),
})
reg.AddClaim(investor, &rwa.IdentityClaim{
    Topic:     rwa.ClaimTopicSanctionsClear,
    Issuer:    kycProvider,
    Value:     []byte("clear"),
    IssuedAt:  time.Now(),
})

// 4. Mint and transfer (compliance-checked)
_ = token.Mint(investor, big.NewInt(10000), owner)
_ = token.Transfer(investor, rwa.NewIdentity([]byte{6}), big.NewInt(100), investor)

// 5. Link to real-world asset and attest reserves
token.Asset = &rwa.Asset{
    Type:           rwa.AssetTypeTreasury,
    Description:    "US Treasury bill collateral",
    Custodian:      rwa.NewIdentity([]byte{7}),
    Auditor:        rwa.NewIdentity([]byte{8}),
    Jurisdiction:   "US",
}
por := &rwa.ProofOfReserve{
    Auditor:        token.Asset.Auditor,
    TokenSupply:    token.TotalSupplyOf(),
    ReserveValue:   token.TotalSupplyOf(),
    AttestedAt:     time.Now(),
    ReportReference: "ipfs://Qmreserve-report",
}
assert.True(t, por.IsValid())
```

## Standards alignment

| Standard | Detail |
|----------|--------|
| ERC-3643 (T-REX) | Compliant security token, ONCHAINID identity, transfer restrictions |
| ERC-1400 | Security token family (partitions, document hooks) |
| ERC-4626 | Tokenized vaults (yield-bearing RWA) |

## Related

- `pkg/compliance/travelrule` — FATF Travel Rule for cross-chain RWA transfers
- `pkg/compliance/sanctions` — OFAC/UN/EU sanctions screening
- `pkg/xintent` — Cross-chain intents for RWA swaps/settlement
- `pkg/multichain` — Multi-chain registry for RWA cross-chain operation
