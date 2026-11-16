---
description: >-
  Run ShorStop in CI: exit codes, the GitHub Action, other CI systems, and baselines for existing codebases.
title: In CI
weight: 20
---

ShorStop is built to run on every push. It's a single static binary that makes
no network calls, so it needs nothing from the CI system except the files to
scan.

Exit codes are what CI acts on:

| Code | Meaning |
|---|---|
| `0` | no finding at or above `--fail-on` (or `--fail-on` not set) |
| `1` | at least one finding at or above `--fail-on` |
| `2` | a usage or scanner error |

```sh
shorstop scan . --fail-on high
```

{{< cards >}}
  {{< card link="github-actions/" title="GitHub Actions" subtitle="The ShorStop Action: verified install, findings in the Security tab and on pull requests." >}}
  {{< card link="other-ci/" title="Other CI systems" subtitle="GitLab CI, and a portable install for any runner." >}}
  {{< card link="baselines/" title="Baselines" subtitle="Adopt ShorStop on an existing codebase: fail only on new findings." >}}
{{< /cards >}}
