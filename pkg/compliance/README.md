# `pkg/compliance` — Unified Compliance Engine

Combines FATF Travel Rule validation (`travelrule`) with sanctions screening
(`sanctions`) into a single decision engine for cross-border and
multi-jurisdiction operation.

## Decision model

| Decision | Meaning |
|----------|---------|
| ALLOW | Transfer passes all compliance checks |
| BLOCK | Transfer must be blocked (sanctions match, invalid Travel Rule data) |
| REVIEW | Transfer requires manual review (unhosted wallet, threshold proximity) |

## Usage

```go
checker := sanctions.NewSanctionsChecker()
_ = checker.LoadList(sanctions.ListOFACSDN, sanctionedEntries)

engine := compliance.NewEngine(checker)
result, err := engine.Evaluate(travelrule.EU, travelrule.US, tx)

switch result.Decision {
case compliance.DecisionAllow:
    // proceed
case compliance.DecisionBlock:
    // halt transfer
case compliance.DecisionReview:
    // escalate to compliance officer
}
```

## Jurisdiction coverage

Travel Rule thresholds cover EU, US, UK, Singapore, UAE, Japan, Korea, India,
Brazil, Canada, Australia, and Switzerland as of November 2026.

## Related

- `pkg/compliance/travelrule` — FATF Rec 16 / IVMS 101 / EU TFR
- `pkg/compliance/sanctions` — OFAC, UN, EU, UK, VARA, MAS screening
