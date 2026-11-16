---
description: >-
  Adopt ShorStop on an existing codebase: record today's crypto findings in a baseline and fail CI only on new ones.
title: Baselines
weight: 3
---

A codebase that has used RSA for ten years will have findings on day one, and
they won't all be fixed this quarter. A baseline records today's findings, so
CI reports and fails on **new** ones only, while the inventory still lists
everything.

## Create one

Scan once and commit the result:

```sh
shorstop scan . --format sarif --output .shorstop/baseline.sarif
git add .shorstop/baseline.sarif
```

A CBOM (`--format cbom`) works as a baseline too. ShorStop tells the two apart
from the content.

## Use it

```sh
shorstop scan . --baseline .shorstop/baseline.sarif --fail-on high
```

Or set it once in [`.shorstop.yaml`](../../configuration/):

```yaml
baseline: .shorstop/baseline.sarif
```

In the GitHub Action, pass it as `baseline:`. Findings in the baseline are left
out of the table, SARIF, Markdown and `--fail-on`, and counted in the summary.
The CBOM still lists them, because the cryptography is still there.

## How findings are matched

Each finding has a fingerprint: a hash of the rule, the file, what was found,
and the source line's text, **without the line number**. So:

- **Moving code doesn't make old findings new.** Adding lines above a known
  finding leaves its fingerprint unchanged.
- **Editing the line does.** If the line with a known finding changes, ShorStop
  treats it as new, because it can't tell an edit from a new use. Review it and
  update the baseline.
- **Duplicates are counted.** If the baseline has one `import "crypto/md5"` in
  a file and the file now has two identical ones, one of them is new.
- **Renaming a file makes its findings new**, since the path is part of the
  fingerprint.

The same fingerprint is what GitHub code scanning uses to track an alert across
commits, so dismissed alerts stay dismissed when code moves.

## Keep it current

When you fix findings, they simply stop appearing; the baseline doesn't need
to change. Regenerate it when you deliberately accept new findings (or after
renames), and review the diff like any other change:

```sh
shorstop scan . --format sarif --output .shorstop/baseline.sarif
git diff --stat .shorstop/baseline.sarif
```

Writing over the baseline file means regenerating it, so ShorStop doesn't
apply the baseline in that run (it says so), and the new file records every
finding. The rest of `.shorstop.yaml` (excludes, custom and disabled rules)
still applies, so the baseline matches what CI compares against.

{{< callout type="info" >}}
On v0.1.2 and earlier, add `--baseline ""` to that command. Without it, a
`baseline:` set in `.shorstop.yaml` is applied while writing, and the new file
holds only the findings that weren't in the old one.
{{< /callout >}}

ShorStop never scans its own reports (SARIF, CBOM or Markdown, recognised by
their content, whatever they're called), so a baseline or an old report left
in the repository doesn't produce findings of its own. The summary line counts
them as skipped ShorStop reports.
