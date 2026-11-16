---
description: >-
  Configure ShorStop with .shorstop.yaml, suppress accepted findings with shorstop:ignore, and write custom detection rules.
title: Configuration
weight: 30
---

## `.shorstop.yaml`

ShorStop reads `.shorstop.yaml` from the root of the scanned directory, or the
file `--config` names. `--no-config` ignores it. Every key is optional:

```yaml {filename=".shorstop.yaml"}
version: 1
exclude: [third_party/, "*.min.js"]
fail-on: high
baseline: .shorstop/baseline.sarif
max-file-size: 4MB
rules: [.shorstop/rules.yaml]    # your own rules, in the rule-pack format
disable: [generic-rsa-bits]      # turn a rule off for the whole repository
```

| Key | |
|---|---|
| `version` | always `1` |
| `exclude` | gitignore-style patterns to skip, on top of `.gitignore` |
| `fail-on` | `critical`, `high`, `medium`, `low` or `none` |
| `baseline` | a previous SARIF or CBOM output; see [Baselines](../ci/baselines/) |
| `max-file-size` | skip larger files, e.g. `512KB`, `4MB` (default `2MB`) |
| `rules` | extra rule files; see [Custom rules](#custom-rules) |
| `disable` | rule IDs to turn off; see the [rules reference](../rules/) |

- **Paths are relative to the config file**, so the same file works wherever
  ShorStop is run from.
- **Flags override the file.** `--fail-on`, `--baseline`, `--max-file-size`
  replace the configured value; `--exclude` and `--rules` add to it.
- **Mistakes are errors.** An unknown key or an unknown rule ID in `disable`
  stops the scan with exit code 2, so a typo can't silently do nothing.
- ShorStop never scans its own config, rule, baseline or output files.

## Suppressing a finding

When a finding is known and accepted, mark it where it is. Put
`shorstop:ignore` in a comment on the line, or on a line of its own just above
it, in whatever comment style the file uses:

```python
key = RSA.generate(2048)  # shorstop:ignore py-pycryptodome-rsa-generate reason="partner API until 2027"
```

```nginx
# shorstop:ignore tls-protocol-version reason="legacy clients, removed in Q3"
ssl_protocols TLSv1 TLSv1.2 TLSv1.3;
```

- Rule IDs are comma-separated: `shorstop:ignore generic-3des,generic-rc4`.
  A marker without IDs covers every finding on its line.
- `reason="..."` is optional but recommended: it's recorded with the
  suppression, so the next reader knows why.
- Suppressed findings leave the table, SARIF, Markdown and `--fail-on`, and
  are counted in the summary. The CBOM still lists them, because the crypto is
  still there.

Which to use: a **suppression** for one accepted place, **`disable:`** for a
rule that doesn't fit your repository at all, **`exclude:`** for code that
isn't yours (vendored, generated), and a **[baseline](../ci/baselines/)** for
everything that's already there when you adopt ShorStop.

## Custom rules

Rules use the same format as the
[built-in pack](https://github.com/shorstophq/shorstop/blob/main/internal/rules/rules.yaml).
Add your own for in-house wrappers, internal libraries or languages' APIs that
ShorStop doesn't know yet:

```yaml {filename=".shorstop/rules.yaml"}
version: 1
rules:
  - id: acme-legacy-cipher
    description: ACME's in-house wrapper around a 64-bit block cipher.
    detector: regex
    languages: [python]
    algorithm: 3DES
    primitive: block-cipher
    match:
      keywords: [legacy_cipher]
      pattern: 'legacy_cipher\('
      examples: ["legacy_cipher('blowfish')"]
```

| Field | |
|---|---|
| `id` | lower-case words joined by `-`; must not clash with a built-in rule |
| `description` | one sentence, shown in reports |
| `detector` | `regex` (any text), `goast` (Go imports and calls) or `deps` (packages) |
| `algorithm` | one of ShorStop's algorithm names, such as `RSA`, `ECDSA`, `AES`, `MD5`, `ML-KEM` |
| `primitive`, `key_size`, `mode` | optional, when the rule fixes them |
| `languages` | `regex` only: limit to these languages; empty means every text file |
| `match` | depends on the detector; below |

**`regex`** rules match line by line:

- `keywords`: literal words; the pattern runs only on lines that contain one,
  which keeps scans fast.
- `pattern`: an [RE2](https://github.com/google/re2/wiki/Syntax) pattern. A
  named group `(?P<bits>...)` supplies the key size, `(?P<mode>...)` the mode.
- `examples`: lines the pattern must match. **Required**, and checked every
  time the rules load, so a rule can't silently stop matching.

**`goast`** rules match Go code: `import` (an exact import path), optionally
`call` (a function of that package), and `key_size_arg` (which argument holds
the key size).

**`deps`** rules match dependencies: `ecosystem` (`npm`, `pypi`, `maven`,
`golang`, `cargo`, `composer`, `gem`, `nuget`) and `package`, spelled as the
ecosystem spells it (`org.bouncycastle:bcprov-jdk18on` for Maven). A trailing
`*` after a separator matches any suffix: `org.bouncycastle:bcprov-*`.

Check a rule file without scanning anything:

```sh
shorstop rules list --rules .shorstop/rules.yaml
```

A rule that would be useful to everyone is welcome upstream: see
[Adding a rule](https://github.com/shorstophq/shorstop/blob/main/CONTRIBUTING.md#adding-a-rule).
