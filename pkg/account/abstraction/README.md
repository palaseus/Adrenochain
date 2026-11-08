# `pkg/account/abstraction` — ERC-4337 Account Abstraction

Off-chain building blocks for smart-contract wallets, aligned with
ERC-4337 (final Standards Track; EntryPoint live on mainnet since March 2023,
26M+ smart accounts and 170M+ UserOperations as of 2026) and EIP-7702
(Pectra, May 2025).

## Components

* **`UserOperation`** — the pseudo-transaction struct (sender, nonce,
  initCode, callData, gas limits, fees, paymasterAndData, signature) with
  `Hash()` and `ValidateBasic()`.
* **`EntryPoint`** — registers sender accounts and validates UserOperations
  against them (`ValidateUserOp`), mirroring the on-chain EntryPoint.
* **`Account`** — interface a smart contract account implements to validate
  a UserOperation.
* **`Paymaster` / `SponsoredPaymaster`** — gas abstraction: a paymaster
  sponsors transactions for whitelisted senders.
* **`Authorization`** — EIP-7702 authorization for an EOA to delegate to
  contract code (`SignHash`).
* **`GasEstimate` / `MaxFeeForGas`** — fee estimation helpers.

## Design notes

* This package is **off-chain only** — it does not deploy or interact with
  on-chain EntryPoint/paymaster contracts. A bundler or wallet uses these
  structures to build, validate, and fee-estimate UserOperations.
* `UserOperation.Hash()` uses SHA-256 here as a placeholder; mainnet
  ERC-4337 uses keccak256. Replace with the project's hashing primitive for
  mainnet parity.

## Example

```go
ep := abstraction.NewEntryPoint()
ep.RegisterAccount(sender, myAccount)
res, err := ep.ValidateUserOp(uo)
```

## Related

* `pkg/account/...` — future account lifecycle management
* ERC-4337: https://eips.ethereum.org/EIPS/eip-4337
* EIP-7702: https://eips.ethereum.org/EIPS/eip-7702
