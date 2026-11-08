# `pkg/mev` — MEV Protection & Private Mempool

MEV (Maximal Extractable Value) protection and private mempool infrastructure,
aligned with 2026 blockchain privacy and fairness standards.

## What it does

MEV extraction (front-running, sandwich attacks, priority gas auctions) harms
users by reordering, inserting, or censoring transactions. This package
provides:

- **Private mempool**: transactions are submitted encrypted, preventing
  observers from seeing transaction contents before inclusion
- **Fair ordering**: FIFO ordering prevents priority gas auctions
- **Fair batch ordering**: transactions are batched and shuffled within each
  batch, preventing front-running within the batch
- **Commit-reveal**: transactions are ordered by commitment hash, preventing
  front-running entirely (adds latency)
- **Bundle**: atomic transaction bundles (all-or-nothing execution)

## Components

| Component | Role |
|-----------|------|
| `Transaction` | signed blockchain transaction with hash computation |
| `PrivateTx` | encrypted transaction wrapper for private submission |
| `PrivateMempool` | private transaction pool with commitment validation |
| `MEVProtector` | ordering strategy (FIFO, FairBatch, CommitReveal) |
| `Bundle` | atomic transaction bundle (all-or-nothing) |

## Example

```go
// 1. Create a private mempool and MEV protector
mempool := mev.NewPrivateMempool()
protector := mev.NewMEVProtector(mev.StrategyFairBatch, 10)

// 2. Create and submit a transaction privately
tx := mev.NewTransaction(sender, recipient, 1000, 100, 21000, 1, []byte("data"))
pt := mev.NewPrivateTx(tx)
mempool.Submit(tx, pt)

// 3. Get transactions in protected order
ordered := mempool.GetOrderedTransactions()
reordered := protector.Order(ordered)

// 4. Create an atomic bundle
bundle := mev.NewBundle([]byte("bundle1"), reordered, 3000)
```

## Protection strategies

| Strategy | Behavior |
|----------|----------|
| FIFO | First-in, first-out. Prevents priority gas auctions. |
| FairBatch | Transactions batched and shuffled within each batch. Prevents front-running within batches. |
| CommitReveal | Ordered by commitment hash. Prevents front-running entirely (adds latency). |

## Standards alignment

| Standard / concept | Detail |
|--------------------|--------|
| MEV protection | Front-running, sandwich attack prevention |
| Private mempool | Encrypted transaction submission |
| Fair ordering | FIFO and batch-fair ordering |

## Related

- `pkg/xintent` — Cross-chain intents can use MEV protection for settlement ordering
- `pkg/rwa` — RWA transfers benefit from MEV protection (prevent front-running large transfers)
