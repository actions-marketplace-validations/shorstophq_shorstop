---
description: >-
  Install ShorStop with Homebrew, Docker, signed binaries or go install, run your first post-quantum cryptography scan, and read the results.
title: Getting started
weight: 10
---

## Install

{{< tabs >}}

{{< tab name="Homebrew" >}}
macOS and Linux:

```sh
brew install --cask shorstophq/tap/shorstop
```

Shell completions for bash, zsh and fish are installed with it.
{{< /tab >}}

{{< tab name="Docker" >}}
linux/amd64 and linux/arm64. The image holds only the binary and runs as a
non-root user. Mount the directory to scan at `/src`:

```sh
docker run --rm -v "$PWD:/src:ro" ghcr.io/shorstophq/shorstop
docker run --rm -v "$PWD:/src:ro" ghcr.io/shorstophq/shorstop scan --format sarif > shorstop.sarif
```
{{< /tab >}}

{{< tab name="Binaries" >}}
Linux and macOS (amd64, arm64) and Windows (amd64) archives are on the
[releases page](https://github.com/shorstophq/shorstop/releases). Each
release's `checksums.txt` is signed with [Sigstore](https://www.sigstore.dev/)
by the repository's release workflow, so there's no key to trust, only the
workflow's identity:

```sh
V=0.1.2 OS=linux ARCH=amd64   # or OS=darwin, ARCH=arm64
for f in "shorstop_${V}_${OS}_${ARCH}.tar.gz" checksums.txt checksums.txt.sigstore.json; do
  curl -fsSLO "https://github.com/shorstophq/shorstop/releases/download/v$V/$f"
done
cosign verify-blob checksums.txt --bundle checksums.txt.sigstore.json \
  --certificate-identity "https://github.com/shorstophq/shorstop/.github/workflows/release.yml@refs/tags/v$V" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
shasum -a 256 --ignore-missing -c checksums.txt   # or: sha256sum --ignore-missing -c checksums.txt
tar -xzf "shorstop_${V}_${OS}_${ARCH}.tar.gz"
```

The container image is signed the same way:
`cosign verify ghcr.io/shorstophq/shorstop:v0.1.2` with the same two
`--certificate-*` flags. Every archive also has an SBOM (`.sbom.json`).
{{< /tab >}}

{{< tab name="From source" >}}
Go 1.27 or newer:

```sh
go install github.com/shorstophq/shorstop/cmd/shorstop@latest
```
{{< /tab >}}

{{< /tabs >}}

## Your first scan

```sh
shorstop scan .
```

ShorStop walks the directory, honouring `.gitignore`, and prints a table grouped
by severity. Each group lists what was found, why it matters, what to use
instead, and every place it occurs with the ID of the rule that found it:

```text
CRITICAL · 3
  3DES-CBC · BROKEN · 1 place
    3DES's 64-bit block allows practical collision attacks (Sweet32), and NIST disallowed it after 2023.
    → Use AES-256-GCM (or ChaCha20-Poly1305).
      deploy/nginx.conf:4   tls-cipher
  MD5 · BROKEN · 1 place
    MD5 has practical collision attacks: signatures and integrity checks that rely on it can be forged.
    → Use SHA-256 or stronger (SHA-384, SHA-3).
      server/keys.go:7   go-md5-import
  RSA-1024 · BROKEN · 1 place
    1024-bit RSA is below the 2048-bit minimum (NIST SP 800-131A) and within reach of classical attacks.
    → Migrate to ML-KEM (FIPS 203) for key exchange and encryption, and ML-DSA (FIPS 204) for signatures; for TLS, enable X25519MLKEM768.
      server/keys.go:20   go-rsa-generate-key

HIGH · 10
  RSA-2048 · QUANTUM_VULNERABLE · 2 places
    RSA is broken by Shor's algorithm; if it protects key exchange or encryption, data encrypted today can be recorded and decrypted later.
    → Migrate to ML-KEM (FIPS 203) for key exchange and encryption, and ML-DSA (FIPS 204) for signatures; for TLS, enable X25519MLKEM768.
      scripts/sign.py:3   py-cryptography-rsa-generate
      server/keys.go:15   go-rsa-generate-key
  ...

Scanned 8 files · 22 findings · 3 critical · 10 high · 4 medium · 3 low (skipped: 1 ignored, ...) · 0 warnings
```

That's a scan of [shorstophq/shorstop-demo](https://github.com/shorstophq/shorstop-demo),
a small service with deliberately bad crypto. Clone it to try ShorStop on
something with findings.

The `BROKEN` and `QUANTUM_VULNERABLE` labels, and the severities, are explained
in [How findings are rated](../ratings/).

## Other output formats

```sh
shorstop scan . --format cbom     --output cbom.json      # CycloneDX 1.6 CBOM, every finding
shorstop scan . --format sarif    --output shorstop.sarif # SARIF 2.1.0 for code scanning
shorstop scan . --format markdown > report.md             # for a pull request comment or an audit
```

`--format json` is the same CBOM as `--format cbom`. The table is coloured only
on a terminal, and never when `NO_COLOR` is set.

## Useful flags

| Flag | |
|---|---|
| `--fail-on <severity>` | exit 1 if a finding is at or above `critical`, `high`, `medium` or `low` |
| `--exclude <pattern>` | skip paths, gitignore-style; repeatable |
| `--baseline <file>` | report only findings not in a previous SARIF or CBOM output |
| `--max-file-size <size>` | skip larger files (default `2MB`) |
| `--list-files` | print the files that would be scanned, and stop |
| `--workers <n>` | files scanned in parallel (default: the number of CPUs) |

Exit codes: `0` clean, `1` findings at or above `--fail-on`, `2` a usage or
scanner error. `shorstop scan --help` lists every flag.

## Next

- Run it on every push: [In CI](../ci/).
- Already have a lot of findings? Record them in a [baseline](../ci/baselines/)
  and fail only on new ones.
- Tune it for your repository: [Configuration](../configuration/).
