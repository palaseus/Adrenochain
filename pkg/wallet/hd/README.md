# `pkg/wallet/hd` — Hierarchical Deterministic Wallets (BIP-32 / BIP-39 / BIP-44)

Implementation of the HD wallet pipeline using only the Go standard library plus
the project's existing `btcec` dependency. Aligned with the standards current as
of November 2026.

## Pipeline

```
mnemonic (BIP-39)
    -> seed = PBKDF2-HMAC-SHA512(mnemonic, passphrase, 2048)  (BIP-39)
    -> master node = HMAC-SHA512("Bitcoin seed", seed)        (BIP-32)
    -> account node = m / purpose' / coin_type' / account'    (BIP-44 + SLIP-44)
    -> change/address node                                    (BIP-44)
    -> public key / address
```

## Standards

| Spec | What it covers |
|------|----------------|
| BIP-39 | Mnemonic phrases and seed generation |
| BIP-32 | Hierarchical deterministic key derivation |
| BIP-44 | Multi-account hierarchy (purpose `44'`) |
| SLIP-44 | Coin-type registry (e.g. Bitcoin `0'`, Ethereum `60'`) |

## Packages

- `pkg/wallet/hd/bip39.go` — mnemonic generation/validation, seed derivation
  (PBKDF2-HMAC-SHA512, 2048 iterations). Verified against the BIP-39 test vector
  and Python's `hashlib.pbkdf2_hmac`.
- `pkg/wallet/hd/bip32.go` — extended keys, `Derive`, `DerivePath` (supports
  hardened `'`/`h` notation). Verified against Python `hmac`/`hashlib`.
- `pkg/wallet/hd/pbkdf2.go` — PBKDF2-HMAC-SHA512 (verified against RFC 6070).

## Usage

```go
// 1. Generate a 12-word mnemonic
words, _ := hd.NewMnemonic(128)

// 2. Derive the seed (BIP-39)
seed, _ := hd.NewSeedFromMnemonic(words, "")

// 3. Create the master node (BIP-32)
master, _ := hd.NewMasterFromSeed(seed)

// 4. Derive an account: m / 44' / 60' / 0'  (Ethereum, account 0)
//    60' is the SLIP-44 coin type for Ethereum.
acct, _ := master.DerivePath("m/44'/60'/0'")

// 5. Derive an external address: m / 44' / 60' / 0' / 0 / 5
addrKey, _ := acct.DerivePath("0/5")
```

## Notes

* Fingerprint hashing uses a double-SHA-512 surrogate because the Go standard
  library lacks RIPEMD160 (BIP-32 specifies `RIPEMD160(SHA256(pubKey))`). This
  is structurally compatible but not byte-identical to reference implementations
  that use true RIPEMD160.
* Hardened derivation requires the parent extended key to contain a private key.

## Security

Never log or transmit mnemonics, seeds, or private keys. Store seeds encrypted
at rest (e.g. AES-GCM with an Argon2id-derived key).
