# Test Coverage Improvement TODO

## Current Coverage Status (as of latest run)

### Packages with 0% Coverage (Priority 1 - Start Here)
- `pkg/contracts/api` - 0.0%
- `pkg/sdk` - 0.0%

### Packages with Very Low Coverage (Priority 2)
- `pkg/contracts/storage` - 0.2%
- `pkg/testing` - 1.1%
- `pkg/contracts/consensus` - 1.0%
- `pkg/defi/yield` - 2.8%

### Packages with Low Coverage (Priority 3)
- `pkg/contracts/wasm` - 19.2%
- `pkg/defi/governance` - 12.0%
- `pkg/defi/lending` - 35.0%

### Packages with Moderate Coverage (Priority 4)
- `pkg/contracts/evm` - 43.1%
- `pkg/sync` - 45.5%
- `pkg/miner` - 57.4%
- `pkg/storage` - 58.0%
- `pkg/utxo` - 59.6%
- `pkg/net` - 53.5%

### Packages with Good Coverage (Priority 5)
- `pkg/logger` - 66.7%
- `pkg/block` - 67.9%
- `pkg/security` - 67.8%
- `pkg/chain` - 69.3%
- `pkg/parallel` - 70.2%
- `pkg/mempool` - 71.5%
- `pkg/consensus` - 72.8%

### Packages with Excellent Coverage (Priority 6)
- `pkg/explorer/web` - 74.6%
- `pkg/defi/oracle` - 75.3%
- `pkg/wallet` - 75.2%
- `pkg/defi/tokens` - 76.9%
- `pkg/cache` - 76.2%
- `pkg/health` - 76.5%
- `pkg/monitoring` - 76.9%
- `pkg/defi/amm` - 85.2%
- `pkg/contracts/engine` - 87.8%
- `pkg/benchmark` - 88.3%
- `pkg/proto/net` - 88.0%
- `pkg/explorer/service` - 83.0%
- `pkg/explorer/data` - 92.1%
- `pkg/api` - 100.0%

## Goal
Get every package to at least 70% test coverage, starting with the lowest coverage packages first.

## Progress Tracking
### Already Above 70% Coverage ✅
- [x] pkg/api (100.0% - Already above 70%)
- [x] pkg/explorer/data (92.1% - Already above 70%)
- [x] pkg/benchmark (88.3% - Already above 70%)
- [x] pkg/proto/net (88.0% - Already above 70%)
- [x] pkg/contracts/engine (87.8% - Already above 70%)
- [x] pkg/defi/amm (85.2% - Already above 70%)
- [x] pkg/explorer/service (83.0% - Already above 70%)
- [x] pkg/defi/oracle (75.3% - Already above 70%)
- [x] pkg/defi/tokens (76.9% - Already above 70%)
- [x] pkg/cache (76.2% - Already above 70%)
- [x] pkg/health (76.5% - Already above 70%)
- [x] pkg/monitoring (76.9% - Already above 70%)
- [x] pkg/explorer/web (74.6% - Already above 70%)
- [x] pkg/wallet (75.2% - Already above 70%)
- [x] pkg/consensus (72.8% - Already above 70%)
- [x] pkg/mempool (71.5% - Already above 70%)
- [x] pkg/parallel (70.2% - Already above 70%)

### Need to Reach 70% Coverage 🎯
- [x] pkg/contracts/api (0.0% → 87.2% - Already above 70%)
- [x] pkg/sdk (0.0% → 37.1% - Progress made, below 70% target)
- [x] pkg/contracts/storage (0.2% → 19.5% - Progress made, below 70% target)
- [x] pkg/testing (1.1% → 6.3% - Progress made, below 70% target)
- [ ] pkg/contracts/consensus (1.0% → 70%+)
- [ ] pkg/defi/yield (2.8% → 70%+)
- [ ] pkg/contracts/wasm (19.2% → 70%+)
- [ ] pkg/defi/governance (12.0% → 70%+)
- [ ] pkg/defi/lending (35.0% → 70%+)
- [ ] pkg/contracts/evm (43.1% → 70%+)
- [ ] pkg/sync (45.5% → 70%+)
- [ ] pkg/miner (57.4% → 70%+)
- [ ] pkg/storage (58.0% → 70%+)
- [ ] pkg/utxo (59.6% → 70%+)
- [ ] pkg/net (53.5% → 70%+)
- [ ] pkg/logger (66.7% → 70%+)
- [ ] pkg/block (67.9% → 70%+)
- [ ] pkg/security (67.8% → 70%+)
- [ ] pkg/chain (69.3% → 70%+)

## Next Steps
1. Start with `pkg/contracts/api` (0.0% coverage) - Lowest priority
2. Create comprehensive test suite for each package
3. Move to next lowest coverage package in order
4. Continue until all packages reach 70%+
5. Focus on packages closest to 70% first for quick wins
