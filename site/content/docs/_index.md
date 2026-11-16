---
description: >-
  Documentation for ShorStop, the post-quantum cryptography scanner: install, CI, configuration, ratings, the CBOM and every rule.
title: Documentation
linkTitle: Docs
---

ShorStop is a post-quantum cryptography scanner. It finds the cryptography in a
repository that a quantum computer will break (RSA, elliptic curves,
Diffie-Hellman) and the crypto that's broken already (MD5, SHA-1, DES, RC4, short
keys, old TLS), rates each finding, and names the post-quantum replacement.

The name comes from Shor's algorithm, the quantum algorithm that breaks RSA and
elliptic-curve cryptography.

{{< callout type="info" >}}
**Early release (`v0.x`).** The scanner, rules and output formats work and are
tested. Rule IDs and output details may still change before `v1.0`.
{{< /callout >}}

{{< cards >}}
  {{< card link="getting-started/" title="Getting started" subtitle="Install, run your first scan, read the output." >}}
  {{< card link="ci/" title="In CI" subtitle="The GitHub Action, other CI systems, and baselines for existing code." >}}
  {{< card link="configuration/" title="Configuration" subtitle=".shorstop.yaml, suppressions and custom rules." >}}
  {{< card link="ratings/" title="How findings are rated" subtitle="Statuses, severities, and why." >}}
  {{< card link="cbom/" title="The CBOM" subtitle="What a Cryptographic Bill of Materials is, and what ShorStop puts in it." >}}
  {{< card link="rules/" title="Rules reference" subtitle="Every rule, with an example of what it finds." >}}
  {{< card link="limitations/" title="Known limitations" subtitle="What a static scanner can and can't see." >}}
{{< /cards >}}
