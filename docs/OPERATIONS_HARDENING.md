# Operations Hardening Guide

This document summarizes the production-hardening controls added to Adrenochain as
of November 2026, and how to operate them.

## 1. Circuit breakers (pkg/resilience/circuitbreaker)

Chain connectors and external-service calls are wrapped in circuit breakers so a
slow or failing downstream does not cascade into a node-wide outage.

- **Closed**: requests flow normally; failures are counted.
- **Open**: requests fail fast with `circuitbreaker.ErrCircuitOpen`.
- **Half-open**: a limited number of trial requests probe recovery.

Defaults: trip after 5 consecutive failures, stay open 30s, allow 3 half-open
trials, close after 2 successes.

Usage:
```go
conn := chainregistry.NewProtectedConnector(rawConn, circuitbreaker.DefaultConfig())
state := conn.BreakerState()   // closed/open/half-open
stats := conn.BreakerStats()
```

## 2. API rate limiting (pkg/api/server.go)

All HTTP API endpoints are protected by a per-client-IP rate limiter
(100 requests/minute by default). Excess requests receive `429 Too Many
Requests`. Client IP is resolved from `X-Forwarded-For` / `X-Real-IP` with
`RemoteAddr` fallback.

Tune via `api.NewRateLimiter(window, limit)` before constructing the server.

## 3. Multi-chain & cross-chain operation (pkg/multichain)

Adrenochain now operates across EVM chains (Ethereum, Arbitrum, Optimism, Base,
Polygon, BSC, Avalanche C-Chain), Bitcoin, Solana, Cosmos SDK, Substrate, and
native Adrenochain networks — mainnet and testnet for each.

Each `ChainSpec` is tagged with a `Region` and `Jurisdictions` list, so the
compliance layer applies the correct FATF Travel Rule / MiCA TFR thresholds.

Cross-chain messages use CCMP (pkg/multichain/ccmp): atomic, ordered,
authenticated messages with replay nonces, expiry, cancel/refund, and a relay
registry that prevents double-relay rewards.

## 4. Compliance & sanctions (pkg/compliance)

- **Travel Rule** (FATF Rec 16 / IVMS 101 / EU TFR): originator/beneficiary
  identity models, per-jurisdiction thresholds (EU/US/UK/SG/UAE/JP/KR/IN/BR/CA/AU/CH;
  no minimum threshold in EU per TFR), unhosted-wallet EDD flagging.
- **Sanctions screening** (pkg/compliance/sanctions): OFAC SDN, UN, EU, UK HM
  Treasury, VARA, MAS. Exact wallet matching, name/alias substring matching,
  address substring matching. Lists loaded offline via `LoadList` (refresh
  daily per FATF guidance).
- **Unified engine** (pkg/compliance/engine): single `ALLOW` / `BLOCK` /
  `REVIEW` decision combining sanctions + Travel Rule + threshold proximity.

## 5. Post-quantum cryptography (pkg/crypto/pqc)

Hybrid PQC aligned with NIST FIPS 203 (ML-KEM), FIPS 204 (ML-DSA), FIPS 205
(SLH-DSA), and NIST IR 8547 (Oct 2026):

- Hybrid KEM: X25519 (classical) + Ring-LWE-style (post-quantum)
- Hybrid signatures: secp256k1 ECDSA + WOTS+ (hash-based)
- Defaults: ML-KEM-768 / ML-DSA-65 per NIST guidance

Use for long-term confidentiality and signatures that must remain verifiable
after a cryptanalytically relevant quantum computer exists.

## 6. Transactional privacy (pkg/crypto/stealth)

ERC-5564 stealth addresses (final) + ERC-6538 meta-addresses. Senders generate
one-time destinations recipients alone can recognize (view-key scan) and spend
(spend-key recovery). Breaks on-chain linkability of payments.

## 7. Account abstraction (pkg/account/abstraction)

ERC-4337 UserOperation / EntryPoint / paymaster model plus EIP-7702
authorization, enabling smart-contract wallets, gas sponsorship, and batched
transactions without consensus-layer changes.

## 8. HD wallets (pkg/wallet/hd)

BIP-39 mnemonics, BIP-32 HD key derivation, BIP-44 accounts, SLIP-44 coin
types. Verified against the Python `hmac`/`hashlib` standard library (the
authoritative reference for HMAC-SHA512 / PBKDF2).

Pipeline: mnemonic -> PBKDF2 seed -> master node -> `m/44'/coin'/account'/change/index`.

## 9. mTLS / mutual authentication

Node-to-node P2P communication (libp2p) should use TLS with mutual
authentication. Configure the libp2p transport with `security.TLSConfig` using
a private CA, and require valid client certs on all RPC and P2P ports. Rotate
certificates at least every 90 days.

## 10. Secrets management

Never commit mnemonics, seeds, private keys, or API keys. All sensitive values
come from environment variables or a secrets manager (HashiCorp Vault, AWS
Secrets Manager, Kubernetes Secrets). `config/production.yaml` uses
`{{.Variable}}` placeholders for admin addresses and reward addresses; provide
them at runtime via `config/env.example` / root `env.example`.

## 11. Dependency auditing

Run `go mod tidy` and `go vet` in CI (see `.github/workflows/ci.yaml`).
Periodically run `govulncheck` (`go install golang.org/x/vuln/cmd/govulncheck@latest`)
and `gosec` (`make security`) to surface vulnerable dependencies.

## 12. Checklist before production deployment

- [ ] Circuit breakers tuned per-chain (tighter thresholds for testnet)
- [ ] API rate limits tuned for expected traffic
- [ ] Sanctions lists loaded and a daily refresh job scheduled
- [ ] mTLS configured on P2P + RPC ports
- [ ] Secrets injected via env/secret manager, not committed
- [ ] `go mod tidy` + `go vet` + `govulncheck` + `gosec` clean
- [ ] Travel Rule validator configured for each operating jurisdiction
- [ ] Backups of `adrenochain_data` (LevelDB) volume tested
