# `pkg/compliance/travelrule` — FATF Travel Rule / IVMS 101

FATF Recommendation 16 (Travel Rule) data structures and validation, aligned
with the InterVASP Messaging Standard (IVMS 101), the EU Transfer of Funds
Regulation (TFR, Reg. (EU) 2023/1113), and MiCA CASP obligations as of the
post-transitional environment (MiCA transitional period ended 1 July 2026;
FATF Rec 16 applied by 83% of surveyed jurisdictions as of July 2026).

## Regulatory context

* **EU (MiCA + TFR):** Travel Rule applies to all crypto-asset transfers with
  **no minimum threshold**. CASPs are obliged entities under the AML framework.
* **FATF (global):** Rec 16 — VASPs transmit originator and beneficiary
  information with transfers (USD/EUR 1,000 baseline threshold).
* **US (FinCEN):** > USD/EUR 3,000 (31 CFR 1010.410(f)).
* **Identity data:** IVMS 101 (name, account/wallet address, national ID,
  address, etc.).

## Components

* **`IVMS101Originator` / `IVMS101Beneficiary`** — IVMS 101 identity models
  (natural person / legal person, wallet address, VASP).
* **`Transfer`** — a virtual asset transfer subject to the Travel Rule.
* **`Validator`** — jurisdiction-aware validation:
  * threshold check (no threshold in EU/UAE/FATF baseline),
  * sanctions screening of originator/beneficiary addresses,
  * unhosted-wallet detection (flags enhanced due diligence).
* **`Jurisdiction.Threshold()`** — per-jurisdiction value threshold.

## Example

```go
v := travelrule.NewValidator(travelrule.EU)
res, err := v.CheckTransfer(&travelrule.Transfer{
    Asset:        &travelrule.CryptoAsset{Symbol: "BTC", Amount: big.NewInt(50000000)},
    Originator:   &travelrule.IVMS101Originator{Name: "Alice", AccountNumber: "0xalice"},
    Beneficiary:  &travelrule.IVMS101Beneficiary{Name: "Bob", AccountNumber: "0xbob", BeneficiaryVASP: "vasp-b"},
})
```

## Scope

This package provides **data models and validation logic only**. It does not
send/receive Travel Rule messages; integrate with a Travel Rule solution
provider (notaml, Notabene, Sygna, etc.) for the IVMS 101 messaging layer.

## Related

* `pkg/account/abstraction` — ERC-4337 account abstraction
* FATF Rec 16: https://www.fatf-gafi.org/en/publications/Fatfrecommendations/targeted-updated-virtualassets-vasps-2026.html
* IVMS 101: https://www.fatf-gafi.org/en/publications/Fatfrecommendations/ivms-101.html
