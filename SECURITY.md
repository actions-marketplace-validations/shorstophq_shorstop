# Security policy

## Supported versions

ShorStop is pre-1.0. Only the latest release gets security fixes; upgrade to
it before reporting.

## Reporting a vulnerability

Use GitHub's
[private vulnerability reporting](https://github.com/shorstophq/shorstop/security/advisories/new).
Please don't open a public issue, pull request or discussion for a
vulnerability.

Include the ShorStop version (`shorstop version`), how you installed it, and
the smallest input that reproduces the problem.

## What counts

In scope, for example:

- A crafted file or repository that makes ShorStop crash, hang, use
  unbounded memory, or read or write outside the directory it was asked to
  scan.
- Reports (SARIF, CBOM, Markdown) that let content from a scanned repository
  inject markup or break out of its field.
- Problems with the release chain: signatures, checksums, the container
  image, the Homebrew cask, or the GitHub Action's install and verify steps.

Not vulnerabilities (open a normal issue instead):

- ShorStop missing a use of cryptography, or rating one at a severity you
  disagree with. ShorStop is a static scanner and lists its known
  limitations in the README.
- Weak cryptography it finds in *your* code. That's the report working.
- Findings in the test fixtures under `testdata/`: the keys and certificates
  there are deliberately throwaway.

## Verifying releases

Release checksums and the container image are signed with keyless cosign
from this repository's release workflow. The README's install section shows
the exact `cosign verify-blob` and `cosign verify` commands.
