# Project review and repairs — 2026-11-09

## Assessment

The project has substantial Go implementation and test coverage, but is not production-ready. This review repaired concrete defects in the transaction, wallet, chain recovery, mempool, mining, API, storage, and security paths. It also found unfinished application flows, unimplemented monetary-policy checks, cryptographic placeholders, and concurrency failures outside those repaired paths. Passing the normal suite does not resolve those findings.

This is a code review and debugging pass, not a cryptographic audit or a certification of every subsystem. The initial working tree already contained EVM recovery changes; those were repaired in place rather than reverted. No commits, publication, deployments, or external messages were made.

## Repairs

| Area | Verified defect | Result |
|---|---|---|
| Wallet signing | Wallet and UTXO validation used different signing encodings. Wallet used DER while the validator expected fixed-width scalars. | Wallet uses the validator's shared signing digest and 65-byte public key + 32-byte R + 32-byte S, with low-S normalization. |
| Wallet verification | Verification trusted `tx.Hash` instead of recomputing the transaction contents. Signing serialization truncated values to a byte. | Verification recomputes both signing digest and transaction hash. Regression modifies an amount and fee by 256 and proves rejection. |
| Wallet coin selection | Processed UTXOs use hex public-key-hash addresses while wallet accounts use Base58Check. | Coin selection recognizes processed-chain addresses as well as existing wallet-address fixtures. Amount-plus-fee and available-balance overflow are rejected. |
| Block validation | Transactions were validated independently against the original state. Two transactions could spend the same input; spending an earlier transaction's output could fail. | Validation uses temporary state in block order, rejects duplicate transactions, extra coinbases, and overwriting an unspent output. Validation does not mutate live state. |
| UTXO amounts | Input/output additions could wrap `uint64`. Replacing an outpoint inflated cached balances. | Validation and fee calculation check sums. Replacements reverse the previous balance before adding the new one. |
| UTXO ownership | Caller-owned pointers and byte slices could modify stored outputs without updating balances. Public add/remove methods were unlocked. | Stored and returned outputs are copied; public mutations acquire the mutex. Internal helpers support callers already holding the lock. |
| UTXO application | A failure during a block could leave partially applied output/balance changes. | Touched entries are journaled and restored on application failure, including address-balance overflow. |
| Restart recovery | Only the tip was loaded. Genesis, intermediate blocks, balances, and work were not rebuilt. Read failures could silently reset the persisted chain. | Recovery follows parent hashes, checks heights and block identities, restores canonical history, replays UTXOs, and rebuilds work. Corruption returns an error without overwriting chain state. |
| Chain publication | The in-memory tip advanced before persistence succeeded. Cache maps and height lacked consistent synchronization. | Writers are serialized, public readers are synchronized, persisted tip succeeds before the new memory state is published, and returned tip/work values are copied. |
| Fork handling | Side branches could overwrite canonical height entries and be applied to the wrong UTXO state. Some paths recursively acquired the chain lock. | Incoming blocks must extend the canonical tip. Unsupported reorganizations are rejected explicitly. This is a correctness restriction, not a complete fork implementation. |
| Chain work and size | Unsigned difficulties were converted through `int64`. Block-size accounting omitted header bytes, hashes, and length prefixes. | Work uses unsigned big-integer conversion. Size enforcement matches the current binary encoding. |
| Mempool | Two heaps shared one index. Selection popped a shallow copy, changing live indices under a read lock. | Heaps have independent indices. Selection sorts an independent slice without mutating entries; removal and expiration preserve both heaps. |
| Mempool capacity | An oversized rejected transaction could evict existing transactions. Eviction requested the whole incoming size rather than the space deficit. | Capacity is checked first and only the deficit is evicted. Selection skips entries that do not fit. Production mempools reject coinbases. |
| Consensus | Proof of work used mutable local difficulty instead of the header's difficulty. Future-height finality could underflow. Checkpoint/finality paths nested read locks. | Proof of work and mining use header difficulty, reject unsupported bit ranges, future heights are rejected before subtraction, and finality avoids its outer nested lock. New checkpoints are compared before canonical insertion. |
| Mining | Replacing the stop channel raced with running workers, which could observe a later session's channel. | Each session captures its cancellation channel, mining work is serialized, and cancellation is checked before publication. Callback reads are synchronized. |
| Mining rewards | Height was absent from the coinbase transaction, making equal rewards reuse a transaction hash. Coinbase fees were calculated before selected transactions had been added. Confirmed transactions remained in the pool. | Height makes coinbases distinct; selected fees are included with checked addition; confirmed transactions leave the pool. Two consecutive mined blocks retain separate reward outputs. |
| EVM recovery | Recovery multiplied by S twice, used SHA3 rather than legacy Keccak, and read byte 96 after accepting a 96-byte input. The test chose the first non-error recovery ID. | Recovery uses the existing secp256k1 library, checks the input length and recovery ID, and hashes with legacy Keccak. A deterministic private-key-1 address vector replaces the random test. |
| EVM RIPEMD-160 | The advertised RIPEMD-160 precompile returned SHA-256 as a placeholder. | It now computes RIPEMD-160 and left-pads the digest to a 32-byte EVM word, checked against the empty-input vector. |
| Cross-chain IDs | IDs used only one time-derived byte, allowing at most 256 values per type and causing collisions. | Network, pool, farm, and system IDs use UUIDs. Network insertion checks nil/empty identity and capacity without narrowing the configured limit. |
| API | `/transactions/{hash}` shadowed `/transactions/pending`. Forwarding headers bypassed per-IP rate limits; IPv6 parsing was incorrect. | Literal route takes precedence. Identity uses the socket peer, correctly parsed for IPv4/IPv6. HTTP request/header/write/idle timeouts were added. |
| Security audit | `BlockIP` appended to the audit log after releasing its mutex. | Block event and blocked-IP mutation are protected by the same lock. |
| File storage | Direct creation truncated the previous committed file before a write completed. | Blocks and chain state use temporary files, file sync, atomic rename, and directory sync. New files are private. |
| Build tooling | Make targets referenced nonexistent `cmd/adrenochain`. | Build/install target the existing `cmd/gochain`. |

## Validation

- Final isolated `go test ./... -timeout 180s`: passed, 96 packages with tests and 12 packages without tests, including all new regressions.
- `go build ./...`: passed.
- `go vet ./...`: passed, including after the final rollback change.
- Focused race checks passed for API, block, chain, consensus, EVM recovery, cross-chain DeFi IDs, mempool, storage, sync, UTXO, and wallet, including the final rollback regression. Miner fixtures were subsequently isolated and passed race detection as well. The corrected RIPEMD-160 vector also passed the EVM package race check. The repaired audit-log path passed `go test -race ./pkg/security -run '^TestSecurityManager_Concurrency$' -count 1`. The complete security package still fails because its synthetic validator deliberately writes concurrently to shared memory; this is not suppressed.
- `go test ./pkg/block -run '^$' -fuzz '^FuzzMalformedWireData$' -fuzztime 10s -parallel 2`: passed, 190,232 executions. Block, transaction, header, input, and output decoders were exercised with arbitrary bytes.
- `git diff --check`: passed.
- The whole-project race discovery run failed. Its findings are retained in [REVIEW_RACE_FINDINGS.md](REVIEW_RACE_FINDINGS.md). Mining and security audit-log failures from that discovery run were subsequently repaired; other findings remain.
- Browser execution, JavaScript dependency installation, and SDK compilation were not performed. Static inspection established that the wallet is missing required entry/configuration files and seven imported components/pages. The TypeScript compiler and package dependencies were not installed for this review.

Existing tests were updated where they explicitly asserted unsafe or invalid behavior: corrupt-state reset, trusting arbitrary forwarding headers, a short previous-transaction hash, and mining fixtures that attempted effectively impossible proof-of-work. The mining test now joins its goroutine before inspecting the nonce. The concurrent storage test retains each stored block hash instead of rebuilding it with a new timestamp, removing a real second-boundary failure. No failing production path was hidden by skipping its tests.

Two overlapping Go invocations interfered with fixed test directories used by existing miner tests. Final verification is isolated; miner fixtures now use independent temporary directories. Other existing fixed-directory fixtures still require sequential test processes.

## Remaining critical findings

### P0 — Block issuance has no enforced monetary policy

`pkg/consensus/consensus.go:validateBlockTransactions` and `pkg/utxo/utxo.go:ValidateTransactionInBlock` check coinbase structure, but do not limit coinbase value to a network subsidy plus validated fees. A miner can choose an arbitrary positive reward within the numeric range. `GenesisBlockReward` and the miner's `CoinbaseReward` are distinct settings and even have different defaults.

Required work: define a consensus subsidy schedule and units, enforce the reward ceiling independently of the miner, validate fees against staged inputs, and test inflation attempts at boundaries. Choosing a monetary policy was not inferred from a genesis-only configuration field.

### P0 — Confidential-transfer proofs and authorization are placeholders

`pkg/zk/zk.go:RangeProof.Verify` accepts a proof by comparing embedded commitment bytes. `ConfidentialTransfer.Verify` checks that a signature is nonempty but does not verify sender authorization. This is not a zero-knowledge range proof or authenticated transfer protocol.

Required work: integrate a reviewed proof system and sender-signature scheme, or keep these interfaces explicitly unavailable for settlement. Tests of the current placeholders do not establish cryptographic security.

### P0 — Flash loans assume repayment

`pkg/defi/lending/advanced/flash_loans.go:executeFlashLoan` restores internal liquidity after a successful callback without checking repayment, fees, or an atomic ledger transition. The whole-project race run also shows concurrent `big.Int` mutation of available liquidity at lines 117, 234, and 252.

Required work: establish a transactional accounting boundary, reserve liquidity atomically, verify actual principal-plus-fee repayment, and roll back callback effects on failure. A mutex alone would not repair the accounting model.

### P0 — Browser wallet exposes secrets and fabricates success

`apps/web-wallet/src/hooks/useWallet.tsx:56` constructs an address from a substring of the private key. Line 64 writes the full private key to localStorage. Lines 93 and 112 fabricate balances and transaction hashes; sending reports success without submitting a signed transaction. The passphrase does not encrypt the stored key.

`apps/web-wallet/src/App.tsx` imports seven absent files: Layout, Send, Receive, DeFi, Settings, CreateWallet, and ImportWallet. The package also lacks `index.html`, TypeScript configuration, and Vite/Tailwind/PostCSS configuration.

Required work: complete the application, use the chain's actual key/address/signing rules, keep unlocked key material out of persistent plaintext storage, display authoritative balances, and report success only after actual submission. This requires implementation of the application/API contract, not cosmetic fixes to the existing simulation.

## Remaining high-priority findings

1. **Canonical fork choice and reorganizations remain unsupported.** `Chain.AddBlock` now rejects competing branches rather than mutating canonical state incorrectly. A network must implement parent-specific validation, cumulative work, bounded rollback/replay, and atomic canonical index updates before operating with real forks.
2. **Difficulty rules are inconsistent.** `Consensus.calculateExpectedDifficulty` and `adjustDifficulty` use different intervals/current-block contexts. Multiplying difficulty by actual/expected time moves the bit-difficulty target in the wrong direction for slow blocks. Configuration allows values much larger than the supported 1–256 proof-of-work bit range. Difficulty history and the next mining target are not fully reconstructed on restart. A deterministic, specified consensus algorithm is required.
3. **Finality uses an arbitrary work threshold.** `IsBlockFinal` requires `finalityDepth * 1000` work; at default difficulty 1, depth alone does not provide the documented behavior. The finality/security model needs definition and adversarial tests, beyond the underflow and locking repairs.
4. **Wallet balances are cached independently of chain state.** `Wallet.GetBalance` reads `Account.Balance`; the API delegates to that method. Chain UTXO replay does not make this cached field authoritative. Coinbase recipient scripts also need one explicit address-decoding rule rather than accepting arbitrary configured strings.
5. **Cryptographic wire formats need versioning.** EVM recovery deliberately retains this engine's documented custom 97-byte hash/R/S/V input and 20-byte output; it is not the standard Ethereum precompile ABI. Wallet signing was repaired to match the existing UTXO verifier. Old DER wallet transactions require rebuilding/re-signing. The legacy signing digest has a 32-bit lock time; larger values now fail explicitly instead of being truncated. Transaction/signing encodings still lack a fully specified, domain-separated, length-prefixed protocol.
6. **Storage encryption-key lifecycle is incomplete.** `Storage.NewStorage` can generate an encryption key when none is supplied without persisting a recoverable key. Reopening such an encrypted store cannot recover its key. Define explicit caller-provided keys or a durable key-management mechanism. Atomic block/state writes do not fix that lifecycle.
7. **SDK and server contracts differ.** `sdk/javascript/src/client.ts:getBlockchainInfo` calls `/api/v1/blockchain/info`, while the server exposes `/api/v1/chain/info`. Several SDK wallet/signing/submission operations have no corresponding server handlers. Define supported endpoints and run contract tests rather than promising absent capabilities.
8. **Concurrency failures remain in other packages.** Discovery reported bridge callback-test synchronization, validator removal, flash-loan liquidity, market-maker quote access, WebSocket client maps, governance callbacks/markets, shard/sidechain transaction status, privacy DeFi, and testing infrastructure. See the recorded conflict locations. Some performance assertions also fail under race instrumentation and should be assessed separately from data races.
9. **Mutable object exposure remains outside UTXO access.** Chain block pointers, mempool transaction pointers, and several subsystem maps are returned directly. Callers must not mutate shared objects. A wider ownership/copying contract is still required.
10. **Security validation includes simulations rather than subsystem checks.** `pkg/security/real_security_validator.go:testDataRaceConditions` deliberately writes concurrently to a slice and returns a result based on the iteration number. Its full-package race failure remains; it must become an isolated, explicit harness rather than an in-process race or a claim of subsystem validation.
11. **Operational lifecycle remains incomplete.** The API now has timeouts but does not expose graceful shutdown to the node. Rate-limit identity intentionally ignores forwarding headers until a trusted-proxy policy exists. Rate-limiter cardinality and background service shutdown deserve explicit load/lifecycle tests.

## Limits and tradeoffs of this patch

- UTXO block validation snapshots the current set. This favors correctness and isolation over scalability; large-chain validation should use a bounded overlay or transactional state backend.
- Block files and the tip are committed in order, not in one multi-record database transaction. Failed tip persistence can leave an unreferenced block file, which is safe to ignore. External mutation of the publicly exposed UTXO set is not part of the chain's writer transaction.
- Corrupt history now stops startup. Recovery is an explicit operator task; startup no longer silently discards the persisted chain.
- Coinbase height uniqueness changes newly mined coinbase transaction IDs. Wallet signature repair changes newly generated transactions. These are behavior changes that must be included in a network upgrade/migration decision.
- New UUID object IDs preserve string prefixes but change their format and length.
- No claim is made that the absent browser wallet flows, zero-knowledge proofs, bridge settlement, lending accounting, or network consensus are ready for financial use.

## Suggested implementation order

1. Specify and enforce issuance, difficulty, address/signature encoding, and fork-choice rules together with negative consensus vectors.
2. Implement transactional ledger state and verified settlement/repayment. Disable placeholder cryptographic settlement paths until real verification is available.
3. Resolve the remaining independently reproducible races, replacing sleep-based tests with channels or synchronization and adding snapshot getters for mutable state.
4. Complete the API/SDK contract and browser wallet, including secure key handling and real transaction submission.
5. Add restart/fork/load/crash-recovery integration tests and reproducible dependency/build workflows, then conduct an independent security review.
