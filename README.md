## GoChain

A modular educational blockchain implementation in Go. It demonstrates a simple blockchain with blocks, transactions, proof-of-work mining, a mempool, a basic wallet, improved UTXO integration, optional P2P networking (libp2p), and optional persistent storage (BadgerDB).

This codebase is designed for learning and experimentation rather than production use.

### Features
- Blocks and transactions with a Merkle root
- Very simple proof-of-work (target-based) mining
- In-memory blockchain and mempool
- Minimal wallet with ECDSA (P-256) keys, signing and verification
- **Persistent and encrypted wallet for secure storage of keys on disk.**
- Improved UTXO (Unspent Transaction Output) integration for tracking and managing transaction outputs.
- Integrated P2P networking using libp2p (GossipSub, DHT, mDNS discovery)
- Integrated persistent storage using BadgerDB
- CLI for running a full node, managing wallet, sending transactions, and inspecting chain state

### Architecture overview
- `cmd/gochain`: The main CLI entrypoint that orchestrates and connects all core components to run a full node.
- `pkg/block`: Core types for Block, Header, Transaction, inputs/outputs, and validation helpers
- `pkg/chain`: In-memory blockchain, genesis block creation, adding and validating blocks, difficulty calculation
- `pkg/mempool`: Transaction mempool and selection for block assembly
- `pkg/miner`: Periodic block assembly and proof-of-work hashing loop
- `pkg/wallet`: Key generation, simple address derivation, transaction creation and signing, verification. **Now includes functionality for persisting and encrypting wallet data to disk.**
- `pkg/net`: P2P networking (libp2p GossipSub, DHT-based discovery, MDNS). The `Network` struct implements `libp2p/core/network.Notifiee` and `libp2p/p2p/discovery/mdns.Notifee` for robust peer handling.
- `pkg/storage`: Persistence layer backed by BadgerDB. The `chain.NewChain` now explicitly takes a `*storage.Storage` instance.
- `pkg/consensus`: Stubs or simple helpers to extend later
- `pkg/utxo`: Manages the Unspent Transaction Output (UTXO) set, including adding and removing UTXOs, and tracking address balances. This is a foundational component for transaction validation and double-spend prevention.

Components are designed with dependency injection, allowing for flexible connections and testability.

### Build targets and Go versions
This repository supports two build modes to keep the default toolchain footprint small and allow development on older Go versions:

- Default (no tags):
  - Compiles `pkg/...` packages with Go 1.18+
  - P2P networking and storage are stubbed out
  - CLI requires Go 1.20+; on older Go versions the CLI prints a message and exits

- Full node (recommended): Go 1.20+ with tags `p2p` and `db`
  - Enables libp2p networking and BadgerDB storage, which are now fully integrated into the `cmd/gochain` application.
  - Builds the CLI to run a functional node.

Examples:

```bash
# Run tests for core packages (any supported Go)
go test ./pkg/...

# Build full node (requires Go 1.20+) with P2P and DB enabled
go build -tags='p2p db' ./cmd/gochain

# Run full node (requires Go 1.20+)
./gochain --port 30303 --mining
```

If you are on Go 1.18–1.19 and only want to work with the libraries (not the network/storage or CLI), you can build/test `./pkg/...` without tags.

### Install
```bash
# Clone
git clone https://github.com/gochain/gochain
cd gochain

# Optional: use Go 1.20+
# Build full node with networking and storage
go build -tags='p2p db' ./cmd/gochain
```

### Protobuf Generation

This project uses Protocol Buffers for defining network messages. Go code is generated from `.proto` files.

To generate the Go code from the `.proto` definitions, ensure you have `protoc` (the Protocol Buffer compiler) and `protoc-gen-go` (the Go plugin for `protoc`) installed and available in your system's PATH.

You can typically install `protoc-gen-go` using:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Once the tools are set up, you can generate the Go code using the provided `Makefile`:

```bash
make proto
```

This command will generate `message.pb.go` in the `pkg/proto/net/` directory.

### Configuration
An example config is provided at `config/config.yaml`. Relevant sections:
- `network`: listen port, bootstrap peers, MDNS, peer limits
- `blockchain`: genesis reward, block time, difficulty adjustment
- `mining`: threads, coinbase address, reward
- `mempool`: max size and minimum fee rate
- `wallet`: key type and passphrase (passphrase unused in current code)
- `storage`: data directory, db type (Badger is supported)
- `api`, `metrics`: placeholders for future expansion

CLI flags in `cmd/gochain` override config values:
- `--config`: path to YAML config
- `--port`: network port (0 for random)
- `--mining`: enable mining
- `--network`: string label (mainnet/testnet/devnet) used for logs only
- `--wallet-file`: path to wallet data file (default: `wallet.dat` in `./wallet_data`)
- `--passphrase`: passphrase for wallet encryption/decryption

### CLI usage
```bash
# Run the node (requires Go 1.20+; build with -tags='p2p db' for full features)
./gochain --config ./config/config.yaml --port 0 --mining

# Create or load a wallet (now persistent and encrypted)
# Use --wallet-file and --passphrase flags
./gochain wallet --wallet-file mywallet.dat --passphrase "your_secret_passphrase"

# Send a transaction (loads wallet from file)
./gochain send --from <address> --to <recipient> --amount 1000 --fee 10 --wallet-file mywallet.dat --passphrase "your_secret_passphrase"

# Query an address balance (loads wallet from file)
./gochain balance --address <address> --wallet-file mywallet.dat --passphrase "your_secret_passphrase"

# Get chain info (height, best block, etc.)
./gochain info
```

Note: The wallet now persists keys to disk and encrypts them using the provided passphrase.

### Testing
```bash
# Run all package tests (recommended)
go test ./...

# Run specific package
go test ./pkg/wallet -v

# Run the main application integration test
go test ./cmd/gochain -v
```

### Recent Updates

This section summarizes recent changes made to the codebase:

*   **Comprehensive Build and Test Fixes:**
    *   Resolved numerous build errors across `cmd/gochain`, `pkg/chain`, `pkg/wallet`, `pkg/miner`, and `pkg/net` by updating function signatures and imports to align with API changes in core packages.
    *   Corrected `chain.NewChain` and `miner.NewMiner` calls to properly integrate `consensus.ConsensusConfig`.
    *   Enhanced `pkg/chain/chain_test.go` by implementing proper block mining within `TestAddBlock` to satisfy proof-of-work requirements.
    *   Rectified `pkg/mempool`'s eviction logic in `mempool.go` and `mempool_test.go` by correctly implementing a min-heap for fee-based transaction eviction, ensuring `TestMempoolEviction` passes as expected.
    *   Cleaned up unused imports and variables across various packages, improving code hygiene.
    *   All existing tests now pass consistently.
*   **UTXO and Wallet Integration Fixes:**
    *   Resolved compilation errors and unused variable warnings in `pkg/wallet/wallet_test.go` related to `ecdsa.GenerateKey` and `wallet.generateAddress` usage.
    *   Corrected the `ScriptPubKey` and address handling logic in `pkg/utxo/utxo_test.go` to ensure consistency with the `UTXOSet`'s internal representation, leading to successful `TestProcessBlock` execution.
*   **Persistent and Encrypted Wallet Implementation:**
    *   Implemented persistent storage for wallet data using the existing file-based `pkg/storage` mechanism.
    *   Added AES-GCM encryption/decryption for wallet data, secured by a user-provided passphrase.
    *   Updated `pkg/wallet` to load and save wallet data automatically.
    *   Integrated wallet persistence into the CLI commands (`wallet`, `send`, `balance`) via new `--wallet-file` and `--passphrase` flags.
    *   Added comprehensive unit tests for wallet persistence and encryption/decryption.
*   **Test File Creation:**
    *   Added placeholder test files for `pkg/proto/net` and `proto/net` to ensure all packages have a basic test presence.
*   **Architectural Audit:**
    *   A high-level architectural audit was performed, confirming the project's modular design suitable for educational purposes. It also highlighted areas for future improvement in robustness, scalability, and security, consistent with the existing "Security/audit summary" and "Roadmap ideas".

If building/running with `-tags='p2p db'`, use Go 1.20+ to satisfy dependencies of libp2p, quic-go, and Badger.

### Developer notes
- Build tags:
  - `p2p`: enables the libp2p-backed `pkg/net` implementation
  - `db`: enables the Badger-backed `pkg/storage` implementation
- Without these tags, stub implementations are compiled to avoid pulling in heavy dependencies on older toolchains.
- The blockchain, miner, wallet, and mempool are intentionally simplified for clarity.

### Security/audit summary (high-level)
This codebase is educational and not hardened. Notable findings and recommendations:

- Wallet/crypto
  - Uses ECDSA P-256 from the Go stdlib, not secp256k1 (common in many chains).
  - Address derivation: SHA-256 of uncompressed public key, last 20 bytes. There is no checksum (e.g., no base58check/bech32). Collisions are highly unlikely but usability and safety (typo detection) are weak.
  - Transaction signing format is non-standard (stores public key concatenated with raw r||s). There is no DER encoding or canonical s enforcement; malleability is possible. Recommendation: adopt a canonical signature format and enforce low-s, or switch to an existing, vetted transaction format.
  - **The wallet now persists keys to disk and encrypts them using AES-GCM with a user-provided passphrase. However, the passphrase itself is not stored, and key derivation is a simple SHA-256 hash, which is not resistant to brute-force attacks. For production, a key derivation function like scrypt or Argon2 should be used.**

- Transaction model
  - Inputs reference a placeholder 32-byte prev hash with index 0; however, the `pkg/utxo` now provides basic UTXO tracking and balance validation within the `ProcessBlock` function. Double-spend prevention is still not fully implemented at the transaction validation level.
  - `CreateTransaction` currently omits balance checks and does not create change outputs. Recommendation: enhance `CreateTransaction` to integrate with the UTXO set for proper balance checks and to generate change outputs.

- Block validation and consensus
  - Proof-of-work target and difficulty adjustment are extremely simplified. The target representation is not endian- or field-validated, and difficulty retargeting is naive.
  - Timestamp validation is minimal (only monotonic check versus prev block). Recommendation: add drift windows and median-time-past style checks.
  - Merkle root is computed over transaction hashes with simple pair-duplication semantics, which is fine for demo but not optimized.

- Networking (p2p)
  - Relies on libp2p GossipSub and DHT discovery. There is no message validation beyond JSON decoding; no application-level signature checks on network messages.
  - No peer scoring or DoS protection. Recommendation: validate and rate-limit inbound gossip, add lightweight scoring and bans.

- Storage
  - Badger-backed storage is optional; default build stubs it out. There is no compaction strategy beyond a simple GC call nor any recovery routines.

- Concurrency and resource management
  - Mining loop uses simple tick-based generation and early returns on context/cancel signals. There is no backpressure if mempool is large.
  - Some maps are protected with RWMutex; access patterns are straightforward. Nevertheless, there is no shutdown orchestration across subsystems besides best-effort closes.

- Testing
  - Unit tests exist for `block`, `chain`, `wallet`, and `net`. Placeholder tests have been added for `pkg/proto/net` and `proto/net`. Test coverage for `pkg/utxo` has been significantly improved with the `TestProcessBlock` fix. Recommendation: expand test coverage for `miner`, `mempool`, and `storage`, and add more comprehensive tests for `net`, including fuzzing transaction encoding/decoding and signature verification.

Given these points, do not use this codebase for production networks or managing real value.

### Roadmap ideas
- Further enhance UTXO set or account-based state for comprehensive balance and nonce management.
- **Further enhance persistent and secure wallet (disk keystore, stronger encryption/key derivation, HD keys).**
- Standardized transaction serialization (e.g., RLP/Protobuf) and signature scheme
- Further enhance mempool policy and fee estimation (Initial work on eviction logic completed)
- Further strengthen PoW difficulty adjustment and consensus rules (Initial work on difficulty calculation completed)
- Enhance P2P message validation, signature checks, and peer scoring (Initial network test fixes completed)
- Expand test coverage, including end-to-end integration tests and property-based tests (Significant progress made on unit test stability)
- Monitoring/metrics and a JSON-RPC/REST API

### License
This repository is provided for educational purposes. If you plan to publish or redistribute, add a proper open-source license file (e.g., MIT, Apache-2.0) as appropriate.
