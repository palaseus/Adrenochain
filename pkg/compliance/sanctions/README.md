# `pkg/compliance/sanctions` — Sanctions Screening

Sanctions screening aligned with FATF Recommendation 15 and major sanctions
regimes as of November 2026.

## Supported regimes

| List | Regime |
|------|--------|
| OFAC SDN | US Office of Foreign Assets Control |
| UNSC | UN Security Council Consolidated List |
| EU | EU Consolidated List |
| UK | HM Treasury Consolidated List |
| VARA | UAE Virtual Assets Regulatory Authority |
| MAS | Singapore Monetary Authority |

## Usage

```go
checker := sanctions.NewSanctionsChecker()
_ = checker.LoadList(sanctions.ListOFACSDN, []sanctions.Entry{
    {ID: "12345", List: sanctions.ListOFACSDN, Name: "Blocked Party", WalletAddresses: []string{"0xdeadbeef..."}},
})

matches, _ := checker.ScreenWallet("0xdeadbeef...")
if len(matches) > 0 {
    // block the transfer, file a suspicious activity report
}
```

Screening supports:
- Exact wallet address matching
- Exact and substring name matching (including aliases)
- Substring address matching

## Offline data

This package does not fetch sanctions lists automatically. Operators must load
official lists from the relevant authority (OFAC, UN, EU, UK, VARA, MAS) and
call `LoadList`. Lists should be refreshed at least daily per FATF guidance.
