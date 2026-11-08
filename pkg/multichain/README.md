# `pkg/multichain` — Multi-Chain & Cross-Chain Framework

This directory contains the multi-chain registry and Cross-Chain Messaging
Protocol (CCMP) that let Adrenochain operate across all major blockchain
networks and regions.

## Components

### `pkg/multichain/chainregistry`
Chain registry (`Registry`) with pre-populated specs for:
- **EVM**: Ethereum (mainnet/sepolia/holesky), Arbitrum, Optimism, Base,
  Polygon, BSC, Avalanche C-Chain (mainnet/testnet)
- **Bitcoin**: mainnet, testnet, signet, regtest
- **Solana**: mainnet-beta, testnet, devnet
- **Cosmos SDK**: Cosmos Hub, Osmosis, Injective
- **Substrate**: Polkadot, Kusama
- **Adrenochain**: mainnet and testnet

Each `ChainSpec` encodes chain ID, native asset, RPC endpoints, region,
jurisdictions, confirmation depth, and smart-contract support.

### `pkg/multichain/ccmp`
Cross-Chain Messaging Protocol providing:
- Atomic, ordered, authenticated cross-chain messages
- Per-chain nonce replay protection
- Origin-chain signatures and relayer attestations
- Expiry with cancel/refund lifecycle
- `RelayRegistry` to prevent double-relay rewards

## Region mapping

Chains are tagged with a `Region` and `Jurisdictions` list so the compliance
layer (`pkg/compliance/travelrule`) can route transfers according to the
applicable FATF Travel Rule and MiCA TFR thresholds for that region.

## Extending

To add a new chain:
1. Register a `ChainSpec` in `DefaultRegistry()` or via `Registry.RegisterChain`.
2. Implement the `Connector` interface (health check, block height, submit tx,
   get tx status) and register it with `Registry.RegisterConnector`.
3. Add region/jurisdiction tags so compliance rules apply automatically.
