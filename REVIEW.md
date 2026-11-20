# Adrenochain review (November 2026)

This note records the baseline, what changed on this branch, and what is still worth doing. Measurements below were taken on this checkout.

## Baseline (before this branch)

Toolchain: Go 1.23.12. `go.mod` said `go 1.23.8` / `toolchain go1.23.12`.

| Check | Result |
| --- | --- |
| `go build ./...` | Failed. `pkg/config` is imported by chain, consensus, storage, logger, network, PDF, and benchmarking, and `.gitignore` excluded the whole directory, so the package was not in the tree. |
| `go test ./...` | Could not run. The missing package stops compilation of those dependents. |
| `go vet ./...` | Same compile failure. |
| CI | `.github/workflows/ci.yaml` built, vetted, and tested (including `-race`). It did not run a linter or `govulncheck`. |
| Dependencies | Direct modules were from 2024–2025 (`golang.org/x/crypto v0.41.0`, `quic-go v0.54.0`, `pion/dtls/v3 v3.0.6`, `webtransport-go v0.9.0`, libp2p 0.43). A `replace` pinned `klauspost/compress` to v1.15.15. |

Go 1.23 is outside the two-release support window. The current stable release is **Go 1.27.4**.

## After this branch

Toolchain: Go 1.27.4 (`go 1.27.0`, `toolchain go1.27.4`).

| Check | Result |
| --- | --- |
| `go build ./...` | Passes. |
| `go test -count=1 -timeout 180s ./...` | Passes. No package failures. |
| `go test -race` on chain, consensus, mempool, utxo, block, wallet, miner, net, storage | Passes. No data races reported. |
| `govulncheck ./...` (symbol mode, govulncheck v1.10.0) | One reachable finding remains: **GO-2024-3218** in `github.com/libp2p/go-libp2p-kad-dht@v0.42.2`. The database lists no fixed version. |
| `golangci-lint` v2.15.0, standard linters, `--new-from-rev=origin/main` | 0 issues. |

Before the libp2p bump, the same scan reported 9 reachable findings in `quic-go`, `webtransport-go`, `pion/dtls`, and `kad-dht`. Eight of those are gone after the upgrades below. Standard-library findings tied to Go 1.23 do not appear on 1.27.4.

A full-tree `golangci-lint` run (not limited to this diff) still reports a large pre-existing backlog, dominated by unchecked errors. CI lints the diff against `origin/main` so new code stays clean without pretending that backlog is gone. Counts from a previous full run on this repository shape were on the order of a thousand `errcheck` findings plus smaller `staticcheck`, `unused`, `ineffassign`, and `govet` sets.

## Findings

Severity is impact if the code is used as a live chain or exchange, not a style score.

### Fixed here

| Severity | Issue | Change |
| --- | --- | --- |
| Critical | Module did not build. `pkg/config` was imported and gitignored. | Restored `DefaultSystemConfig` with the values the consensus and network tests assert (12s target block time, finality depth 10, checkpoint interval 262974, 50 peers, 30s timeout, 1e9 genesis reward, 1MB blocks). Stopped ignoring the package. |
| High | `Consensus.GetStats` held `c.mu` and called `GetNextDifficulty`, which locks the same `RWMutex`. That deadlocks in Go. | `GetStats` calls an unlocked helper. |
| High | `Chain.AddBlock` persisted the new tip, then applied the UTXO set. A UTXO error left storage ahead of memory. | On UTXO failure the previous chain state is written back before the error is returned. The in-memory tip is not published. |
| High | AMM `Swap` always added input to `ReserveA` and subtracted output from `ReserveB`, including for a Token B to Token A swap. | Reserves follow the input token. |
| High | Yield `updatePool` added the whole interval reward to `AccRewardPerShare`. Pending rewards are `stake * acc - debt`, so each staker was paid `stake` times the emission. | The increment is `rewards / totalStaked`. |
| Medium | A crossing limit sell was recorded at the sell order's price, not the resting bid. Buys already used the resting ask. | The trade and both fills use `bestBid.Price`. |
| Medium | libp2p, quic-go, webtransport, and DTLS had reachable vulnerability-database findings. | libp2p 0.49.0, kad-dht 0.42.2, pubsub 0.18.0, quic-go 0.60.0, webtransport-go 0.11.1, pion/dtls v3.1.15. `dht.New` now takes the host first, matching v0.42. The test `network.Conn` mock implements `As`. |

`golang.org/x/crypto` is v0.59.0, `x/net` v0.63.0, cobra v1.10.2, viper v1.21.0, protobuf v1.36.12, testify v1.12.1, and klauspost/compress v1.20.5. The compress `replace` that forced v1.15.15 is gone.

### Still open

| Severity | Issue | Why it was left |
| --- | --- | --- |
| Critical | `Bridge.verifySignature` returns nil. Validator records have no public key, so a length check is the only other gate. | A real check needs a key on `Validator` and updated bridge tests. Doing that without keys would reject every current confirmation. |
| Critical | ZK verifiers in `pkg/security/zk_proofs.go` accept a proof whose verification key is `SHA-256(statement \|\| proof prefix)`. | Replacing this with a real proof system is a new dependency and a behavior change for every privacy caller. Treat these paths as research stubs. |
| High | `GET /api/v1/wallet/accounts` has no authentication. | The HTTP API is a local research server. Adding auth changes every client of that route. Bind it to localhost and put an authenticator in front before exposing it. |
| High | ERC-20 `Mint`, pause, and blacklist do not check `Owner`. | The methods have no caller argument. Adding one changes every token call site. |
| High | `AddBlock` still refuses a block that does not extend the canonical tip (`fork reorganization is unsupported`). There is no UTXO undo. | A correct reorg is a new subsystem (disconnect/connect to the common ancestor, bounded by `MaxReorgDepth`). The explicit error is safer than the old partial fork update. |
| Medium | GO-2024-3218, content censorship via Kademlia DHT abuse. | No fixed module version in the vulnerability database, including kad-dht v0.42.2. |
| Medium | Lending liquidation still debits the liquidator's supplied balance instead of seizing borrower collateral. `Borrow` and `checkCollateralRatio` can disagree if callers write `Borrows` by hand. | The tests describe the current accounting. Changing it is a product decision, not a one-line correction. |
| Medium | Oracle proof checks are length checks. Governance `SetVotingPower` has no authorization. Flash-loan code does not require repayment. | Same class of stub: the feature exists as a model, not as a custody or voting system. |
| Low | Custom logger is not `log/slog`. Many `math/rand` uses remain in simulations. | Output format is covered by logger tests. `slog` fits the logger package; switching it should be its own change. Simulations do not need `crypto/rand`. |

## Recommendations, in order

1. Keep CI's `govulncheck` exception limited to GO-2024-3218, and delete the exception when upstream publishes a fix.
2. Do not expose the HTTP API, bridge, or token admin methods beyond a trusted process until callers are authenticated and bridge/ZK checks are real.
3. If the chain is meant to follow a fork, implement connect/disconnect with UTXO undo up to `MaxReorgDepth`. Until then, the "unsupported" error is the behavior.
4. Burn down `errcheck` and `staticcheck` on non-test packages, then drop `--new-from-rev` so CI lints the whole tree.
5. Move the logger onto `log/slog` without changing the public `Logger` methods, and replace deprecated `rand.Seed` calls with a local `math/rand/v2` source.
