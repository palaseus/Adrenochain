# `pkg/crypto/pqc` — Post-Quantum Cryptography

Hybrid post-quantum cryptography aligned with the NIST PQC standards finalized
August 2024 and reaffirmed by NIST IR 8547 (October 2026):

| Standard | Algorithm | Role | Default parameter |
|----------|-----------|------|-------------------|
| FIPS 203 | ML-KEM (CRYSTALS-Kyber) | Key encapsulation | ML-KEM-768 |
| FIPS 204 | ML-DSA (CRYSTALS-Dilithium) | Digital signatures | ML-DSA-65 |
| FIPS 205 | SLH-DSA (SPHINCS+) | Hash-based signatures | SLH-DSA-SHA2-128 |

## What this package provides

Because no production-grade PQC library is an allowed dependency in this
offline build, the package implements **hybrid** cryptography using only the Go
standard library:

* **Hybrid KEM** (`GenerateKEMKeyPair` / `Encapsulate` / `Decapsulate`) — an
  X25519 classical ECDH share combined with a Ring-LWE-style lattice share.
  Output key = `SHA-256(x25519_share || lattice_share || ephemeral || static)`.
  Breaking either layer is required to recover the session key.
* **Hybrid signatures** (`GenerateSigKeyPair` / `Sign` / `Verify`) — secp256k1
  ECDSA combined with a WOTS+ hash-based signature (FIPS 205 aligned).
  A forgery must break both schemes.
* **Hybrid encrypt/decrypt** (`HybridEncrypt` / `HybridDecrypt`) — KEM-derived
  key bound to the plaintext.

## Security notes

* This is a **research/development implementation**, not a FIPS-validated
  module. For production, replace the standard-library placeholder math with a
  vetted PQC library (e.g. `github.com/cloudflare/circl` for ML-KEM/ML-DSA).
* NIST IR 8547: 112-bit security is deprecated after 2030 and disallowed after
  2035; Adrenochain's legacy secp256k1/ECDSA is quantum-vulnerable and should
  be paired with a PQC layer for long-term confidentiality.
* Default parameters follow NIST guidance: ML-KEM-768 and ML-DSA-65 for civil
  systems; ML-KEM-1024 / ML-DSA-87 for CNSA 2.0 environments.

## Related

* `pkg/crypto/stealth` — ERC-5564 stealth addresses (transactional privacy)
* `pkg/security/quantum_resistant.go` — PQC algorithm metadata registry
