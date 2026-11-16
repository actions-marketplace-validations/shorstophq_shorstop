---
title: Known limitations
linkTitle: Limitations
weight: 70
---

ShorStop is a static scanner: it reads files and never runs code. That makes it
fast, safe to run on any repository, and free of network calls, but it can't see
everything. These are the gaps we know about. A scanner that hides its blind
spots is trusted less, not more.

## What it can't see

- **Cryptography decided at run time.** An algorithm name read from an
  environment variable, a database or a remote config isn't in the repository.
- **Cryptography in binaries.** Compiled libraries and vendored `.so`/`.dll`
  files are skipped. Their source or their package-manager entry is what
  ShorStop sees.
- **Libraries' internal defaults.** A dependency is reported as present, but
  which algorithms it uses depends on how your code calls it. The calls are
  found by the language rules when ShorStop knows the API.

## Known false positives and negatives

- **Pattern rules match comments too.** Commented-out crypto is still worth
  seeing in an inventory. Use [`shorstop:ignore`](../configuration/#suppressing-a-finding)
  for the exceptions.
- **No type checking in Go.** A local variable that shadows a package name
  (`rsa := x; rsa.GenerateKey(...)`) is reported as a crypto call.
- **OpenSSL cipher sets that depend on the build** (`HIGH`, `MEDIUM`,
  `DEFAULT`) aren't expanded, so `MEDIUM` on an old OpenSSL that still includes
  3DES isn't caught.
- **Hybrid key exchange:** in `X25519MLKEM768`, the X25519 half is reported as
  quantum-vulnerable even though the hybrid protects it.
- **Maven `${property}` versions and Gradle version catalogs** aren't
  resolved, so those dependencies are reported without a version.
- **Lockfiles over 2 MB** are skipped by default. Raise the limit with
  `--max-file-size` or `max-file-size:` in `.shorstop.yaml`.
- **Report files in the scanned directory are scanned.** A SARIF or CBOM report
  left in the tree mentions algorithms by name, and those mentions are found.
  ShorStop skips the baseline it's using and the file it's writing; write other
  reports outside the scanned directory, or exclude them.

Found another one? [Open an issue](https://github.com/shorstophq/shorstop/issues/new/choose)
with the smallest input that shows it.
