---
title: The CBOM
weight: 50
---

A **Cryptographic Bill of Materials** (CBOM) is an inventory of the
cryptography a system uses: which algorithms, keys, certificates and protocols,
and where each one appears. It's what a post-quantum migration starts from:
you can't replace RSA until you know where it is.

ShorStop writes its CBOM in [CycloneDX 1.6](https://cyclonedx.org/capabilities/cbom/),
the OWASP standard that added cryptographic assets to the CycloneDX SBOM
format, so the same tools that read your SBOMs can read it.

```sh
shorstop scan . --format cbom --output cbom.json
```

Unlike the table, SARIF and Markdown reports, the CBOM lists **everything**:
suppressed and baselined findings, and acceptable and quantum-safe crypto too.
It's an inventory, not a to-do list.

ShorStop's tests validate its CBOM against the official CycloneDX 1.6 JSON
schema, and its output is deterministic apart from the timestamp and serial
number, so successive CBOMs can be diffed.

## What's in it

The examples below come from a scan of
[shorstophq/shorstop-demo](https://github.com/shorstophq/shorstop-demo).

### Algorithms

Each distinct algorithm, with its key size or parameter set, is one
`cryptographic-asset` component. Every place it was found is an
**occurrence**, with the file, line, rule and source line:

```json
{
  "bom-ref": "crypto:algorithm:RSA-1024:unknown",
  "type": "cryptographic-asset",
  "name": "RSA-1024",
  "cryptoProperties": {
    "assetType": "algorithm",
    "algorithmProperties": {
      "primitive": "unknown",
      "parameterSetIdentifier": "1024",
      "classicalSecurityLevel": 80,
      "nistQuantumSecurityLevel": 0
    }
  },
  "evidence": {
    "occurrences": [
      {
        "bom-ref": "shorstop:fp:016e4a08225f73bc9c3b5c8bf04eec30#1",
        "location": "server/keys.go",
        "line": 20,
        "additionalContext": "go-rsa-generate-key: return rsa.GenerateKey(rand.Reader, 1024)"
      }
    ]
  },
  "properties": [
    { "name": "shorstop:status", "value": "BROKEN" },
    { "name": "shorstop:severity", "value": "critical" },
    { "name": "shorstop:reason", "value": "1024-bit RSA is below the 2048-bit minimum (NIST SP 800-131A) and within reach of classical attacks." },
    { "name": "shorstop:recommendation", "value": "Migrate to ML-KEM (FIPS 203) for key exchange and encryption, and ML-DSA (FIPS 204) for signatures; for TLS, enable X25519MLKEM768." },
    { "name": "shorstop:confidence", "value": "high" },
    { "name": "shorstop:rules", "value": "go-rsa-generate-key" }
  ]
}
```

- `primitive` is the CycloneDX primitive (`signature`, `key-agree`, `hash`,
  `block-cipher`...) when ShorStop can tell it, else `unknown`.
- `nistQuantumSecurityLevel` is the NIST post-quantum security category:
  `0` for anything Shor's algorithm breaks, up to `5` for AES-256 or
  ML-KEM-1024.
- The `shorstop:*` properties carry the [rating](../ratings/) and advice.
  `shorstop:confidence` says how the finding was made: `high` when the code or
  certificate was parsed, `medium` for a pattern match, `low` for a dependency.
- Each occurrence's `bom-ref` is the finding's fingerprint, which is what lets
  a CBOM serve as a [baseline](../ci/baselines/).

### Certificates

A certificate is its own asset, linked to its key and signature algorithms:

```json
{
  "name": "CN=demo.example",
  "cryptoProperties": {
    "assetType": "certificate",
    "certificateProperties": {
      "subjectName": "CN=demo.example",
      "issuerName": "CN=demo.example",
      "notValidBefore": "2026-11-13T16:30:42Z",
      "notValidAfter": "2036-11-10T16:30:42Z",
      "signatureAlgorithmRef": "crypto:algorithm:RSA:signature",
      "subjectPublicKeyRef": "crypto:algorithm:RSA-2048:unknown",
      "certificateFormat": "X.509"
    }
  }
}
```

### Protocols

Enabled TLS versions from server configuration:

```json
{
  "name": "TLSv1.0",
  "cryptoProperties": {
    "assetType": "protocol",
    "protocolProperties": { "type": "tls", "version": "1.0" }
  }
}
```

### Libraries

Crypto dependencies are `library` components with a
[package URL](https://github.com/package-url/purl-spec), as in an SBOM. The
`dependencies` section records that the project depends on them and, where
ShorStop knows, which algorithm a library provides:

```json
"dependencies": [
  { "ref": "project", "dependsOn": ["pkg:pypi/cryptography", "pkg:pypi/rsa"] },
  { "ref": "pkg:pypi/rsa", "provides": ["crypto:algorithm:RSA:unknown"] }
]
```

## Using it

- **Track migration.** Diff this month's CBOM against last month's: RSA
  occurrences going down, ML-KEM going up.
- **Answer auditors.** Many post-quantum readiness requirements start with a
  cryptographic inventory; the CBOM is that inventory, in a standard format.
- **Query it.** It's JSON. For example, every quantum-vulnerable algorithm and
  how often it occurs:

```sh
jq -r '.components[]
  | select(any(.properties[]?; .name == "shorstop:status" and .value == "QUANTUM_VULNERABLE"))
  | "\(.evidence.occurrences | length)\t\(.name)"' cbom.json | sort -rn
```
