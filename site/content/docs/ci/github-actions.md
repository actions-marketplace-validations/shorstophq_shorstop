---
description: >-
  The ShorStop GitHub Action: Sigstore-verified install, quantum-vulnerable crypto findings in GitHub code scanning, inputs, outputs and recipes.
title: GitHub Actions
weight: 1
---

The [ShorStop Action](https://github.com/marketplace/actions/shorstop-pqc-scan)
installs a release, verifies its Sigstore signature, scans, and uploads SARIF to
GitHub code scanning. Findings appear in the repository's **Security** tab and
inline on pull requests.

```yaml {filename=".github/workflows/shorstop.yml"}
name: shorstop
on: [push, pull_request]
permissions:
  contents: read
  security-events: write   # upload results to code scanning
jobs:
  scan:
    runs-on: ubuntu-latest   # macOS and Windows runners work too
    steps:
      - uses: actions/checkout@v7
      - uses: shorstophq/shorstop@v0.1.3
```

It's a composite action, not a container one, so it runs on Linux, macOS and
Windows runners and starts in seconds.

## Inputs

All optional, under `with:`.

| Input | Default | |
|---|---|---|
| `path` | `.` | directory to scan, relative to the repository root |
| `fail-on` | `none` | fail the step at or above this severity: `critical`, `high`, `medium`, `low`. Results are uploaded first |
| `baseline` | | a previous SARIF or CBOM output; only findings not in it are reported |
| `config` | `.shorstop.yaml` in `path` | config file |
| `args` | | extra `shorstop scan` flags, split on whitespace (no shell quoting), e.g. `--exclude vendor/**` |
| `upload-sarif` | `true` | set `false` to only write the report |
| `sarif-file` | runner temp directory | where to write the SARIF report |
| `category` | `shorstop` | code scanning category, to keep several runs (one per directory, say) apart |
| `version` | `latest` | ShorStop release to install, e.g. `v0.1.3`. `latest` is the newest full release |
| `verify` | `true` | verify the release's signed checksums before using the download |
| `install` | `true` | set `false` to use a `shorstop` already on `PATH` |

## Outputs

| Output | |
|---|---|
| `version` | the installed ShorStop version |
| `sarif-file` | path of the SARIF report |
| `exit-code` | ShorStop's exit code: `0` clean, `1` findings at or above `fail-on` |

## Recipes

### Fail pull requests on new high-severity findings

```yaml
      - uses: shorstophq/shorstop@v0.1.3
        with:
          fail-on: high
          baseline: .shorstop/baseline.sarif
```

See [Baselines](../baselines/) for creating the baseline file.

### Scan several directories separately

Give each run its own `category`, or the second upload replaces the first:

```yaml
      - uses: shorstophq/shorstop@v0.1.3
        with:
          path: services/api
          category: shorstop-api
      - uses: shorstophq/shorstop@v0.1.3
        with:
          path: services/web
          category: shorstop-web
```

### Keep the report as an artifact

```yaml
      - uses: shorstophq/shorstop@v0.1.3
        id: shorstop
      - uses: actions/upload-artifact@v7
        with:
          name: shorstop-sarif
          path: ${{ steps.shorstop.outputs.sarif-file }}
```

## Pinning

`@v0.1.3` is a tag. To pin exactly, use the commit SHA with the version as a
comment, and set `version` so the binary is pinned too:

```yaml
      - uses: shorstophq/shorstop@<commit-sha> # v0.1.3
        with:
          version: v0.1.3
```

Dependabot updates both the SHA and the comment.

## Things to know

- **Code scanning** is free for public repositories. Private repositories need
  GitHub Code Security.
- **Pull requests from forks** get a read-only token, so GitHub refuses the
  upload. Use `upload-sarif: false` for them, with `fail-on` to still gate the
  pull request.
- **Severity in the Security tab** is ShorStop's own: `critical`, `high`,
  `medium` and `low` map to code scanning's levels of the same names. Code
  scanning takes severity from the rule, and one ShorStop rule can find both
  RSA-1024 (critical) and RSA-4096 (high), so alerts show the rule ID with the
  severity appended, like `go-rsa-generate-key/critical`.
