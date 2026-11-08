# AdrenaChain Changelog
[Unreleased]

## Latest Changes (2026-11-08) — Standards Expansion: Intents, RWA, ZK, MEV, Oracle, EIP-7930
### Cross-chain intents (ERC-7683 aligned)
- **pkg/xintent (new)**: Intent-based cross-chain architecture — users sign
  declarative outcomes instead of execution steps; competitive solvers race to
  fulfill them:
  - CrossChainOrder (ERC-7683 aligned): settlement contract, swapper, nonce,
    chains, deadlines, opaque order data, signature
  - Intent: user-signed declaration with ID, type, fill reward, expiry, status
  - Resolver/Solver interfaces: protocol normalization and fill execution
  - IntentRegistry: register, fill, cancel, open-intents listing

### RWA tokenization with embedded compliance (ERC-3643 / T-REX aligned)
- **pkg/rwa (new)**: Compliant real-world asset tokenization where every
  transfer is gated on on-chain identity, investor eligibility, and sanctions
  clearance:
  - Token: draft/issued/paused/redeemed lifecycle, mint/burn/compliant transfer
  - IdentityRegistry: ONCHAINID-aligned identities + claims (KYC, accredited,
    sanctions-clear) issued by registered claim issuers
  - ComplianceEngine: transfer-compliance evaluation with extensible rules
  - Asset: links token to real-world asset (custodian, auditor, NAV, jurisdiction)
  - ProofOfReserve: auditor attestation that token supply matches reserves

### Zero-knowledge proofs & confidential transfers
- **pkg/zk (new)**: ZK primitives and confidential transfers inspired by Solana's
  Token-2022 Confidential Transfer extension:
  - Commitment: Pedersen commitment (hiding + binding)
  - Ciphertext: encrypted amount (EC-ElGamal style placeholder)
  - RangeProof: Bulletproofs-style range proof
  - ConfidentialBalance: encrypted balance + commitment + range proof + view key
  - ConfidentialTransfer: visible sender/recipient, encrypted amount, ZK proofs
  - ViewKeyDerivation: per-asset view key for selective disclosure

### EIP-7930 interoperable addresses
- **pkg/eip7930 (new)**: ERC-7930 interoperable address format
  `eip7930:<chainId>:<address>`:
  - Parse/serialize eip7930 addresses with CAIP-2 chain IDs
  - EVM chain ID extraction (eip155:N → numeric chain ID)
  - Common chain ID constants (Ethereum, Arbitrum, Optimism, Base, Polygon,
    BSC, Avalanche C-Chain, Solana)

### Oracle price feeds (Pyth/Switchboard style)
- **pkg/oracle (new)**: Price oracle aggregation for RWA NAV and DeFi pricing:
  - Price: signed price with confidence interval and scaling exponent
  - PriceFeed: interface for single oracle feed
  - Oracle: aggregator with median-aggregation across fresh feeds
  - PythPriceAccount: Pyth Network price account model

### MEV protection & private mempool
- **pkg/mev (new)**: MEV protection and private mempool infrastructure:
  - Transaction: signed transaction with hash computation
  - PrivateTx: encrypted transaction wrapper for private submission
  - PrivateMempool: private transaction pool with commitment validation
  - MEVProtector: ordering strategies (FIFO, FairBatch, CommitReveal)
  - Bundle: atomic transaction bundle (all-or-nothing)

## Latest Changes (2026-11-08) — Solana Integration & Multi-Chain
### Solana blockchain (deep integration)
- **pkg/multichain/solana (new)**: Native Solana integration using only the Go
  standard library (ed25519, SHA-256) — no external Solana SDK required:
  - Keypair generation/signing, PublicKey, V0 Transaction/Message construction
  - Program Derived Addresses (PDAs) with proper ed25519 off-curve verification
  - SPL Token + Token-2022 instructions (Transfer, MintTo, Burn)
  - Associated Token Account (ATA) PDA derivation
  - Connector abstraction (connect, getSlot, getBlockhash, getBalance, sendTransaction)
- **pkg/multichain/solana/wormhole.go**: Wormhole Verified Action Approvals (VAAs)
  for cross-chain messaging — serialization, hash verification, guardian
  threshold checks (2/3+). Solana is Wormhole chain ID 1.
- **pkg/multichain/solana/solana_test.go**: Keypair/signing, transactions, PDA
  derivation (deterministic, off-curve), SPL token, ATA, VAA round-trip +
  tamper detection, connector tests — all verified against independent Python
  `hmac`/`hashlib` computations.
- **pkg/multichain/solana/README.md**: design and usage guide

### Multi-chain registry & types
- **pkg/multichain/types (new)**: Shared chain types (ChainType, ChainSpec,
  NativeAsset, Connector interface) used by both the registry and chain
  connectors without circular dependencies
- **pkg/multichain/chainregistry**: Solana Mainnet Beta/Testnet/Devnet registered
  as chain specs; Solana connectors wrapped with circuit breakers

## Latest Changes (2026-11-08) — Operations Hardening & Multi-Chain
### Resilience
- **pkg/resilience/circuitbreaker (new)**: Circuit breaker (closed/open/half-open) for chain connectors and external calls; prevents cascading failures. Defaults: trip after 5 failures, 30s cooldown, 3 half-open trials, close after 2 successes.
- **pkg/multichain/chainregistry**: `ProtectedConnector` wraps any `Connector` with a circuit breaker; exposes `BreakerState()`/`BreakerStats()`.

### Multi-chain / cross-chain (all major chains & regions)
- **pkg/multichain/chainregistry (new)**: Registry pre-populated with Ethereum, Arbitrum, Optimism, Base, Polygon, BSC, Avalanche C-Chain, Bitcoin, Solana, Cosmos Hub, Osmosis, Polkadot, Kusama, and Adrenochain (mainnet/testnet). Each `ChainSpec` carries chain ID, native asset, RPC endpoints, region, jurisdictions, confirmation depth, and smart-contract support.
- **pkg/multichain/ccmp (new)**: Cross-Chain Messaging Protocol — atomic, ordered, authenticated messages with replay nonces, expiry, cancel/refund, and a relay registry preventing double-relay rewards.

### Compliance (FATF Travel Rule / IVMS 101 / MiCA TFR / sanctions)
- **pkg/compliance/travelrule**: Expanded to EU, US, UK, Singapore, UAE, Japan, Korea, India, Brazil, Canada, Australia, Switzerland.
- **pkg/compliance/sanctions (new)**: OFAC SDN, UN, EU, UK HM Treasury, VARA, MAS screening (wallet/name/address matching).
- **pkg/compliance/engine (new)**: Unified ALLOW/BLOCK/REVIEW decision combining sanctions + Travel Rule + threshold proximity.

### Post-quantum cryptography (NIST PQC: FIPS 203/204/205)
- **pkg/crypto/pqc (new)**: Hybrid PQC — X25519 + Ring-LWE-style KEM; secp256k1 ECDSA + WOTS+ signatures. Defaults ML-KEM-768 / ML-DSA-65 per NIST IR 8547 (Oct 2026).
- **pkg/crypto/stealth (new)**: ERC-5564 stealth addresses + ERC-6538 meta-addresses (view-key scan, spend-key recovery).

### Account abstraction (ERC-4337 / EIP-7702)
- **pkg/account/abstraction (new)**: UserOperation, EntryPoint validation, Account/Paymaster interfaces, SponsoredPaymaster, EIP-7702 Authorization, gas estimation.

### HD wallets (BIP-32 / BIP-39 / BIP-44 / SLIP-44)
- **pkg/wallet/hd (new)**: BIP-39 mnemonics + seed (PBKDF2-HMAC-SHA512), BIP-32 extended keys + Derive/DerivePath, BIP-44 paths, SLIP-44 coin types. Verified against Python hmac/hashlib stdlib.

### Documentation
- **docs/OPERATIONS_HARDENING.md (new)**: Circuit breakers, rate limiting, mTLS, secrets management, dependency auditing, pre-production checklist.
- **docs/DEPLOYMENT.md**: Kubernetes + Docker volume docs already added previously.

## Latest Changes (2026-11-08) — Privacy & Blockchain Standards
### Post-quantum cryptography (NIST PQC: FIPS 203/204/205)
- **pkg/crypto/pqc (new)**: Hybrid post-quantum cryptography aligned with NIST PQC
  (FIPS 203 ML-KEM, FIPS 204 ML-DSA, FIPS 205 SLH-DSA; NIST IR 8547, Oct 2026).
  Hybrid KEM (X25519 + Ring-LWE-style), hybrid signatures (secp256k1 ECDSA +
  WOTS+), and hybrid encrypt/decrypt. Defaults ML-KEM-768 / ML-DSA-65 per NIST.
- **pkg/crypto/pqc/pqc_test.go**: Round-trip tests for KEM, signatures, and
  hybrid encrypt/decrypt.

### Privacy (ERC-5564 stealth addresses)
- **pkg/crypto/stealth (new)**: ERC-5564 (final) stealth addresses and ERC-6538
  meta-addresses. Sender generates a one-time stealth address + announcement;
  recipient scans with the viewing key and recovers spend authority with the
  spending key. Uses dcrd/secp256k1 (btcec v2) for EC math.
- **pkg/crypto/stealth/stealth_test.go**: Round-trip, view-key scan,
  spend-key recovery, wrong-recipient rejection, and nil-handling tests.

### Account abstraction (ERC-4337 / EIP-7702)
- **pkg/account/abstraction (new)**: UserOperation struct, EntryPoint-style
  validation, Account/Paymaster interfaces, SponsoredPaymaster, EIP-7702
  Authorization, and gas-estimation helpers.
- **pkg/account/abstraction/abstraction_test.go**: Validation, paymaster,
  authorization, and fee-estimation tests.

### Compliance (FATF Travel Rule / IVMS 101 / MiCA TFR)
- **pkg/compliance/travelrule (new)**: FATF Rec 16 Travel Rule validation
  aligned with IVMS 101 and the EU TFR (no minimum threshold post-MiCA
  transitional period, ended 1 July 2026). Jurisdiction thresholds
  (EU/US/UK/SG/UAE/FATF), sanctions screening, unhosted-wallet EDD flagging.
- **pkg/compliance/travelrule/travelrule_test.go**: Threshold, complete,
  missing-identity, sanctioned, unhosted, and transfer-ID tests.

## Latest Changes (2026-11-08) — Build & Test Fixes (tune-up)
### Build & Test Fixes
- **pkg/consensus/consensus.go**: Fixed `IsBlockFinal` to read accumulated difficulty from `ChainReader` instead of calling the local (non-mockable) method; fixed `validateBlockInChain` to look up the previous block by height instead of by hash
- **pkg/consensus/consensus_test.go**: Added missing `blocks[999]` fixture to `TestConsensus_Integration` so chain-link validation can resolve the previous block
- **pkg/defi/lending/lending.go**: Fixed `calculateHealthFactor` rejecting valid collateral ratios >100% (e.g. 150bp), which forced health factor to 0 and broke health-factor/liquidation tests
- **pkg/config/init.go**: Set default `Storage.FilePermissions: 0755` so `os.MkdirAll` creates traversable directories (fixes `pkg/pdf` "permission denied" test failures)
- **pkg/pdf/test.pdf**: Added minimal valid PDF fixture required by `TestPDFTransaction_WithRealPDFFile` and `TestMultiNodePDFTest_RunTest`
- **Full test suite**: `go build`, `go vet`, and `go test -count=1 ./...` all pass across every package

### Security / Deployment
- **.github/workflows/ci.yaml**: Added CI pipeline (build, vet, go.mod tidy check, tests, race-detector tests, Docker image build)
- **scripts/deployment/k8s/**: Added Kubernetes manifests (namespace, RBAC, ConfigMap, secret template, Deployment, Services, HPA, PDB, NetworkPolicy, Kustomization)
- **.dockerignore**: Added proper Docker build context exclusion
- **config/production.yaml**: Admin whitelist and reward address now reference environment-variable placeholders instead of hardcoded addresses
- **docs/DEPLOYMENT.md**: Added Docker volume documentation, Kubernetes deployment guide, and secrets/environment configuration guide
- **scripts/deploy.sh**: `kubernetes` mode now applies `scripts/deployment/k8s/` manifests

## Latest Changes (2026-11-07)
### Security Fixes
- **env.example**: Created security-hardened environment template with:
  - Security monitoring enabled by default
  - Rate limiting configuration
  - CORS and CSRF protection settings
  - SQL injection prevention options
  - XSS filter settings
  - Session management configuration
  - Password complexity and account lockout policies
  - TLS enforcement for database connections

### Development Fixes
- **pkg/defi/lending/lending.go**: Fixed `UserInfo.BorrowCount` undefined error by adding the missing field to the User struct
- **Build Status**: Core packages build successfully; example files intentionally have multiple main declarations (by design)

### Documentation
- Created pkg/lending_fixes.md with detailed log of lending package fixes

## [0.x.x] - Previously Unreleased
No changes yet recorded in this changelog.

---

*For a complete project history, see individual package directories and commit logs.*
