# `pkg/zk` — Zero-Knowledge Proofs & Confidential Transfers

Zero-knowledge proof primitives and confidential transfers, inspired by
Solana's Token-2022 Confidential Transfer extension (Twisted ElGamal
encryption + ZKPs: Schnorr proofs, compact sigma proofs, Bulletproofs range
proofs) and privacy-preserving systems such as Aztec, Railgun and Aleo.

## What it does

Confidential transfers hide transfer amounts and balances from the public
ledger while remaining cryptographically verifiable. Issuers and auditors
retain selective disclosure through view keys — important for regulated
finance where transaction graphs must remain visible but amounts stay private.

## Components

| Component | Role |
|-----------|------|
| `Commitment` | Pedersen commitment `C = value·G + blinding·H` (hiding + binding) |
| `Ciphertext` | Encrypted amount (EC-ElGamal style placeholder) |
| `RangeProof` | Bulletproofs-style proof that a committed value is in [0, 2⁶⁴) |
| `ConfidentialBalance` | Encrypted balance + commitment + range proof + view key |
| `ConfidentialTransfer` | Confidential transfer: visible sender/recipient, encrypted amount, ZK proofs |
| `ViewKeyDerivation` | Per-asset view key derivation from a master key |

## Example

```go
// 1. Derive a per-asset view key
viewKey := zk.ViewKeyDerivation(masterKey, []byte("asset-1"))

// 2. Create a confidential balance
amount := big.NewInt(1000)
blinding := big.NewInt(7)
balance := zk.NewConfidentialBalance(amount, blinding, viewKey)
assert.True(t, balance.Verify())

// 3. Build a confidential transfer (amount hidden, sender/recipient visible)
transfer := zk.NewConfidentialTransfer(sender, recipient, amount,
    senderBlinding, recipientBlinding, viewKey)
assert.True(t, transfer.Verify())
```

## Privacy model

| What's hidden | What's visible |
|---------------|----------------|
| Transfer amount | Sender address |
| Balance | Recipient address |
| Blinding factors | Transfer nonce |
| | Validity proofs (ZKPs) |
| | View-key holders (auditors/regulators) |

## Standards alignment

| Standard / system | Detail |
|-------------------|--------|
| Solana Token-2022 | Confidential Transfer extension (Twisted ElGamal + ZKPs) |
| Bulletproofs | Short range proofs without trusted setup |
| Schnorr / compact sigma proofs | Proof of knowledge & composition |
| Aztec / Railgun / Aleo | Privacy-preserving blockchain systems |

## ⚠️ Security note

The elliptic-curve / pairing math here is a **structural placeholder**
implemented with the Go standard library. It demonstrates the confidential-
transfer architecture but is **not** an audited cryptographic construction.
Production use must replace these primitives with audited libraries
(e.g. arkworks BN254, circl) and undergo formal cryptographic review before
handling real funds.

## Related

- `pkg/xintent` — Cross-chain intents for private settlement
- `pkg/compliance/sanctions` — Sanctions screening (view-key disclosure to auditors)
- `pkg/compliance/travelrule` — FATF Travel Rule for confidential transfers
- `pkg/rwa` — Confidential RWA tokens (private balances for institutional assets)
