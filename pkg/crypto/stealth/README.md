# `pkg/crypto/stealth` — ERC-5564 Stealth Addresses

Private, non-interactive transfers per ERC-5564 (final Standards Track) and
ERC-6538 (stealth meta-address registry), scheme 1 (secp256k1).

## Protocol

A recipient publishes a **meta-address** `(spend_pub, view_pub)`. A sender who
wants to pay them:

1. Generates an **ephemeral** key pair `(e_priv, e_pub)`.
2. Computes `spend_shared_x = ECDH(e_priv, spend_pub)` and
   `spend_shared = SHA-512(e_pub || spend_pub || spend_shared_x)`.
3. Derives `scalar = spend_shared[0:32] mod n` and the **stealth address**
   `stealth_spend = spend_pub + scalar * G`.
4. Computes `view_shared_x = ECDH(e_priv, view_pub)` and
   `view_shared = SHA-512(e_pub || view_pub || view_shared_x)`;
   sets `view_tag = view_shared[32]`.
5. Publishes an **announcement** `(e_pub, stealth_spend, view_tag)` with the
   transfer.

The recipient:

* **Scans** announcements off-chain using the **viewing key**: recomputes
  `view_shared` from `ECDH(view_priv, e_pub)` and compares `view_tag`. Only
  the recipient recognizes payments to them — observers cannot link the
  stealth address to the meta-address.
* **Recovers spend authority** with the **spending key**:
  `stealth_priv = spend_priv + scalar(spend_shared[0:32])`, where
  `spend_shared` comes from `ECDH(spend_priv, e_pub)`.

## API

* `GenerateMetaAddress()` — recipient meta-address + spend/view private keys.
* `GenerateEphemeralKey()` — sender ephemeral key pair.
* `GenerateStealthAddress(meta, ephem)` — stealth address + announcement.
* `meta.Scan(ann, viewPriv)` — off-chain recognition (viewing key).
* `RecoverSpendPrivateKey(meta, spendPriv, ann)` — spend authority (spending key).

## Privacy properties

* Recipient address is **unlinkable** across transactions.
* Sender learns nothing about the recipient beyond the meta-address.
* Viewing key enables scanning without spend authority.
* Security: view tag reveals one byte (~124-bit privacy); the scalar derivation
  is ~128-bit against the discrete-log assumption on secp256k1.

## Related

* `pkg/crypto/pqc` — post-quantum hybrid KEM/signatures
* ERC-5564: https://eips.ethereum.org/EIPS/eip-5564
* ERC-6538: https://eips.ethereum.org/EIPS/eip-6538
