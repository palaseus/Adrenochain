# `pkg/xintent` — Cross-Chain Intent Framework (ERC-7683 aligned)

Intent-based cross-chain architecture, inspired by ERC-7683 (Cross Chain
Intents, Uniswap/Across Labs) and the solver/resolver model that dominated
cross-chain design in 2026.

## The problem it solves

Instead of users signing transactions that dictate every execution step,
users sign **intents** — declarative, signed messages expressing a desired
outcome (e.g. "deliver asset X on destination chain"). Competitive off-chain
actors called **solvers** race to fulfill the intent, earning a spread or fee.
A **resolver** normalizes protocol-specific orders into a common solver-facing
representation.

This shifts UX from "how to execute" to "what outcome I want," and enables
chain abstraction: the underlying chains disappear from the user's view.

## Components

| Component | Role |
|-----------|------|
| `CrossChainOrder` | ERC-7683-aligned solver-facing order (settlement contract, swapper, nonce, chains, deadlines, opaque order data, signature) |
| `Intent` | User-signed intent declaration with ID, type, fill reward, expiry, status |
| `Resolver` | Normalizes a protocol order into a `CrossChainOrder` |
| `Solver` | Evaluates intents (profit/risk) and fills them |
| `IntentRegistry` | Tracks open intents, records fills, handles cancellation |

## Intent lifecycle

```
Open  ──fill──▶ Filled
  │
  └─cancel──▶ Cancelled
  │
  └─expire──▶ Expired
```

## Example

```go
// 1. Build a cross-chain order (ERC-7683 aligned)
order := &xintent.CrossChainOrder{
    SettlementContract: settlementAddr,
    Swapper:            userAddr,
    Nonce:              big.NewInt(1),
    OriginChain:        1,        // Ethereum
    InitiationDeadline: deadline1,
    FillDeadline:       deadline2,
    OrderData:          orderPayload,
    Signature:          userSig,
}

// 2. Wrap in an intent
intent, err := xintent.NewIntent(order, xintent.IntentTypeTransfer, 100, expiry)

// 3. Register in the intent registry
reg := xintent.NewIntentRegistry()
_ = reg.Register(intent)

// 4. Solver evaluates and fills
si, err := solver.Evaluate(intent)
result, err := solver.Fill(intent, solverAddr)
_ = reg.Fill(result)
```

## Standards alignment

| Standard | Detail |
|----------|--------|
| ERC-7683 | Cross-Chain Intents (Draft) — solver-facing order normalization |
| ERC-7786 | Cross-Chain Messaging Gateway (Final) — cross-chain message delivery |
| ERC-7985 | Gateway Attributes (cancellation, timeout, retry) |

## Related

- `pkg/multichain/ccmp` — Cross-Chain Messaging Protocol (atomic messages)
- `pkg/multichain/chainregistry` — multi-chain registry + circuit breakers
- `pkg/compliance/travelrule` — FATF Travel Rule for cross-chain transfers
- `pkg/compliance/sanctions` — sanctions screening for intents
