# `pkg/multichain/solana` — Solana Blockchain Integration

Deep Solana integration for Adrenochain, aligned with Solana documentation
current as of November 2026 (V1 transactions per SIMD-0385, 100M compute
units per block, ~60 second blockhash expiry).

## Components

### Core (`solana.go`)
- **Keypair**: ed25519 key generation, signing, and seed-based restoration
- **PublicKey**: 32-byte Solana address with base58 encoding
- **Transaction/Message**: V0 transaction construction, signing, and serialization
- **Connector**: RPC client abstraction (connect, getSlot, getBlockhash, getBalance,
  sendTransaction) with placeholder implementations for offline development

### Program Derived Addresses (`pda.go`)
- PDA derivation via `SHA256(program_id || SHA256(seeds) || bump)` with
  canonical bump search (255 → 0), ensuring the result is off the ed25519 curve
- Proper ed25519 on-curve check using the curve equation
  `-x² + y² = 1 + d·x²·y² (mod p)` with Euler's criterion

### SPL Token & Token-2022 (`spl_token.go`)
- SPL Token program instructions: Transfer, MintTo, Burn
- Token-2022 support (transfer hooks, interest-bearing, confidential transfers)
- Associated Token Account (ATA) PDA derivation:
  `FindProgramAddress([owner, token_program, mint], ASSOCIATED_TOKEN_PROGRAM)`

### Wormhole VAAs (`wormhole.go`)
- Wormhole Verified Action Approval (VAA) serialization/deserialization
- VAA hash verification and guardian threshold checks (2/3+)
- Solana is Wormhole chain ID 1

### Chain Registry (`pkg/multichain/chainregistry`)
- Solana Mainnet Beta, Testnet, and Devnet registered as chain specs
- Solana connectors wrapped with circuit breakers for resilience

## Example

```go
// Create a Solana keypair
kp, _ := solana.NewKeypair()

// Derive an Associated Token Account
token := solana.NewSPLToken()
owner := mustPublicKey("...")
mint := mustPublicKey("...")
ata, bump, _ := token.DeriveAssociatedTokenAccount(owner, mint)

// Build a SPL transfer instruction
instr, _ := token.BuildTransfer(source, dest, authority, 1_000_000)

// Build a cross-chain Wormhole VAA
vaa := solana.NewVAA(timestamp, nonce, solana.ChainIDSolana, emitterAddr, sequence, 1, payload)
data, _ := vaa.Serialize()
```

## Standards alignment

| Standard | Detail |
|----------|--------|
| V1 Transactions | SIMD-0385, 4096 byte max |
| Compute Units | 100M per block |
| Blockhash Expiry | ~60 seconds |
| Wormhole Chain ID | 1 (Solana) |
| PDA | Off-curve, program signs via invoke_signed |
| ATA | Deterministic owner+mint address |

## Related

- `pkg/multichain/chainregistry` — multi-chain registry + circuit breakers
- `pkg/multichain/types` — shared chain types
- `pkg/compliance/travelrule` — FATF Travel Rule for cross-chain transfers
- `pkg/compliance/sanctions` — sanctions screening for bridged assets
- `pkg/crypto/stealth` — ERC-5564 stealth addresses (transactional privacy)
