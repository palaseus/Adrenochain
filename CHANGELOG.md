# AdrenaChain Changelog
[Unreleased]

## Latest Changes (2026-11-08)
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
