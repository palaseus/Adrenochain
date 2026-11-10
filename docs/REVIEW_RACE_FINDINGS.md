# Whole-project race-check findings

Command: `go test -race ./... -timeout 180s`.

This is an intermediate discovery run, not a claim that all listed defects remain. Mining lifecycle and security audit-log races were subsequently repaired and checked separately. Some other reports involve tests reading mutable state without synchronization. Reproduce each independently before changing its implementation.

Observed conflicting locations (deduplicated):

- `pkg/bridge/bridge_core_test.go:506` ↔ `pkg/bridge/bridge_core_test.go:485` ↔ `pkg/bridge/bridge_core.go:550` (1 report(s)).
- `pkg/bridge/bridge_core_test.go:507` ↔ `pkg/bridge/bridge_core_test.go:486` ↔ `pkg/bridge/bridge_core.go:550` (1 report(s)).
- `pkg/crosschain/validators/multi_chain_validator.go:332` ↔ `pkg/crosschain/validators/multi_chain_validator_test.go:196` (1 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:234` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` ↔ `pkg/defi/lending/advanced/flash_loans.go:117` (2 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:252` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` ↔ `pkg/defi/lending/advanced/flash_loans.go:234` (4 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:117` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` ↔ `pkg/defi/lending/advanced/flash_loans.go:252` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` (1 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:234` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` ↔ `pkg/defi/lending/advanced/flash_loans.go:252` (3 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:234` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` (7 report(s)).
- `pkg/defi/lending/advanced/flash_loans.go:252` ↔ `pkg/defi/lending/advanced/flash_loans.go:129` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:564` ↔ `pkg/defi/lending/advanced/flash_loans_test.go:573` (4 report(s)).
- `pkg/exchange/advanced/market_making/market_maker_extended_test.go:30` ↔ `pkg/exchange/advanced/market_making/market_maker.go:178` ↔ `pkg/exchange/advanced/market_making/market_maker.go:161` ↔ `pkg/exchange/advanced/market_making/market_maker.go:136` (2 report(s)).
- `pkg/exchange/advanced/market_making/market_maker_test.go:54` ↔ `pkg/exchange/advanced/market_making/market_maker.go:178` ↔ `pkg/exchange/advanced/market_making/market_maker.go:161` ↔ `pkg/exchange/advanced/market_making/market_maker.go:136` (2 report(s)).
- `pkg/exchange/advanced/market_making/market_maker_test.go:306` ↔ `pkg/exchange/advanced/market_making/market_maker.go:178` ↔ `pkg/exchange/advanced/market_making/market_maker.go:161` ↔ `pkg/exchange/advanced/market_making/market_maker.go:136` (2 report(s)).
- `pkg/exchange/api/api_test.go:2235` ↔ `pkg/exchange/api/market_data.go:62` ↔ `pkg/exchange/api/market_data.go:53` (1 report(s)).
- `pkg/exchange/api/api_test.go:2243` ↔ `pkg/exchange/api/market_data.go:69` ↔ `pkg/exchange/api/market_data.go:53` (1 report(s)).
- `pkg/exchange/api/api_test.go:2307` ↔ `pkg/exchange/api/market_data.go:69` ↔ `pkg/exchange/api/market_data.go:53` (1 report(s)).
- `pkg/governance/governance_test.go:523` ↔ `pkg/governance/governance_test.go:514` ↔ `pkg/governance/governance.go:335` (1 report(s)).
- `pkg/governance/governance_test.go:877` ↔ `pkg/governance/governance_test.go:868` ↔ `pkg/governance/governance.go:335` (1 report(s)).
- `pkg/governance/markets/proposal_markets_test.go:626` ↔ `pkg/governance/markets/proposal_markets.go:421` ↔ `pkg/governance/markets/proposal_markets.go:386` ↔ `pkg/governance/markets/proposal_markets.go:151` (1 report(s)).
- `pkg/governance/markets/proposal_markets_test.go:630` ↔ `pkg/governance/markets/proposal_markets.go:430` ↔ `pkg/governance/markets/proposal_markets.go:386` ↔ `pkg/governance/markets/proposal_markets.go:151` (1 report(s)).
- `pkg/governance/markets/proposal_markets_test.go:630` ↔ `pkg/governance/markets/proposal_markets.go:422` ↔ `pkg/governance/markets/proposal_markets.go:386` ↔ `pkg/governance/markets/proposal_markets.go:151` (1 report(s)).
- `pkg/layer2/sharding/sharding.go:575` ↔ `pkg/layer2/sharding/sharding.go:558` ↔ `pkg/layer2/sharding/sharding.go:165` ↔ `pkg/layer2/sharding/sharding_test.go:1158` (1 report(s)).
- `pkg/layer2/sidechains/sidechains.go:608` ↔ `pkg/layer2/sidechains/sidechains.go:591` ↔ `pkg/layer2/sidechains/sidechains.go:195` ↔ `pkg/layer2/sidechains/sidechains_test.go:1519` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:167` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:325` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:395` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:450` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:646` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:661` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:695` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner_test.go:883` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:1220` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` ↔ `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:1426` ↔ `pkg/miner/miner_test.go:1432` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:1493` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:1493` ↔ `pkg/miner/miner.go:343` ↔ `pkg/miner/miner.go:227` ↔ `pkg/miner/miner.go:193` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/miner/miner.go:97` ↔ `pkg/miner/miner_test.go:1578` ↔ `pkg/miner/miner_test.go:1587` ↔ `pkg/miner/miner.go:175` ↔ `pkg/miner/miner.go:100` (1 report(s)).
- `pkg/privacy/defi/private_defi.go:611` ↔ `pkg/privacy/defi/private_defi.go:452` ↔ `pkg/privacy/defi/private_defi_test.go:564` (1 report(s)).
- `pkg/security/real_security_validator.go:617` ↔ `pkg/security/real_security_validator.go:623` (1 report(s)).
- `pkg/security/real_security_validator.go:592` (2 report(s)).
- `pkg/security/real_security_validator.go:620` ↔ `pkg/security/real_security_validator.go:623` (1 report(s)).
- `pkg/security/security_hardening.go:269` ↔ `pkg/security/security_hardening_test.go:445` ↔ `pkg/security/security_hardening_test.go:451` (1 report(s)).
- `pkg/security/security_hardening.go:311` ↔ `pkg/security/security_hardening_test.go:448` ↔ `pkg/security/security_hardening_test.go:451` ↔ `pkg/security/security_hardening.go:269` ↔ `pkg/security/security_hardening_test.go:445` (1 report(s)).
- `pkg/testing/test_runner.go:292` ↔ `pkg/testing/test_runner.go:40` ↔ `pkg/testing/test_runner.go:54` ↔ `pkg/testing/unit_test_framework.go:322` ↔ `pkg/testing/comprehensive_test_suite.go:72` ↔ `pkg/testing/comprehensive_test_suite_test.go:102` (1 report(s)).
- `pkg/testing/test_runner.go:292` ↔ `pkg/testing/test_runner.go:40` ↔ `pkg/testing/test_runner.go:54` ↔ `pkg/testing/unit_test_framework.go:383` ↔ `pkg/testing/comprehensive_test_suite.go:77` ↔ `pkg/testing/comprehensive_test_suite_test.go:124` (1 report(s)).
