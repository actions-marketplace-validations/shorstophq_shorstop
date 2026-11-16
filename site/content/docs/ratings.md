---
title: How findings are rated
linkTitle: Ratings
weight: 40
---

Every finding gets two ratings:

- a **status**: *what kind of problem* it is. Broken today? Broken by a quantum
  computer? Fine?
- a **severity**: *how soon* to act, from `critical` to `info`. It's what
  `--fail-on` and code scanning use.

## Statuses

| Status | Meaning | Examples |
|---|---|---|
| `BROKEN` | Insecure today, quantum computer or not | MD5, SHA-1 signatures, DES, 3DES, RC4, RSA/DH under 2048 bits, SSLv3, TLS 1.0/1.1 |
| `INSECURE_MODE` | A sound cipher used in a mode that leaks | AES-ECB |
| `QUANTUM_VULNERABLE` | Secure today, broken by Shor's algorithm | RSA, ECDSA, ECDH, Ed25519, X25519, DH, DSA |
| `WEAKENED` | Grover's algorithm halves its strength | AES-128 |
| `ACCEPTABLE` | Fine now and after quantum computers | AES-256, SHA-256/384/512, SHA-3, ChaCha20, HMAC-SHA-1 |
| `QUANTUM_SAFE` | Post-quantum cryptography | ML-KEM, ML-DSA, SLH-DSA, sntrup761 |
| `UNKNOWN` | Not enough information to rate it | a crypto toolkit dependency, AES of unknown key size |

## Severities

| Severity | Given to |
|---|---|
| `critical` | broken crypto (MD5, SHA-1, DES, 3DES, RC4), public keys below today's minimum (RSA, DSA, DH under 2048 bits; elliptic curves under 224), SSLv2/SSLv3, and **committed private keys** |
| `high` | quantum-vulnerable key exchange and encryption, quantum-vulnerable keys whose use is unknown, AES-ECB, TLS 1.0 and 1.1 |
| `medium` | quantum-vulnerable signatures, AES-128 |
| `low` | dependencies, HMAC with MD5 or SHA-1, anything ShorStop can't rate |
| `info` | acceptable and quantum-safe crypto: nothing to do |

`info` findings are left out of the table, Markdown and SARIF by default (they'd
be noise), but they're counted in the summary and listed in the CBOM, where
they show migration progress.

## Why these ratings

**Key exchange is more urgent than signatures.** Traffic encrypted with RSA or
ECDH can be recorded today and decrypted once a large quantum computer exists:
"harvest now, decrypt later". A signature only has to be trustworthy when it's
checked, so signature keys can be replaced later, before such a computer
exists. That's why ECDH is `high` and ECDSA signatures are `medium`. When
ShorStop can't tell how a key is used (an `import "crypto/rsa"`, a
certificate's key), it assumes the more urgent case.

**Broken today beats broken someday.** RSA-1024, MD5 and 3DES are `critical`
because they're attackable now, without a quantum computer.

**TLS 1.0 and 1.1 are `high`, not `critical`.** They're deprecated (RFC 8996)
and refused by modern clients, but their known attacks are mitigated in current
TLS stacks. SSLv2 and SSLv3 are practically broken (DROWN, POODLE) and are
`critical`.

**HMAC-SHA-1 is `ACCEPTABLE`.** HMAC and key derivation rely on preimage
resistance, which MD5 and SHA-1 still have; it's collisions that are broken.
Calling HMAC-SHA-1 broken would overstate the risk, so it's `low`: worth
replacing, not urgent.

**Dependencies are capped at `low`.** A lockfile line says a crypto library is
present, not how it's used. The library is reported so it's in the inventory,
and the code that uses it is rated on its own findings.

**A committed private key is `critical` whatever its algorithm.** Anyone who
can read the repository has the key. Rotate it, then remove it from history.

**Hybrid key exchange.** The TLS group `X25519MLKEM768` counts as quantum-safe.
Its X25519 half is currently also reported as quantum-vulnerable; see
[Known limitations](../limitations/).

## The CBOM's extra detail

In the [CBOM](../cbom/), each algorithm also carries its classical security
level and its NIST post-quantum security category (`0` for anything Shor's
algorithm breaks), so other tools can apply their own policy.
