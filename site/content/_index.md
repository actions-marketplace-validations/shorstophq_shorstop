---
description: >-
  Open-source scanner for quantum-vulnerable and broken cryptography in code, certificates, configs and dependencies. Writes a CycloneDX CBOM and SARIF.
title: "ShorStop: post-quantum cryptography scanner"
layout: hextra-home
---

{{< hextra/hero-badge link="https://github.com/shorstophq/shorstop/releases/latest" >}}
  Open source · Apache 2.0
{{< /hextra/hero-badge >}}

{{< hextra/hero-headline >}}
  Find the cryptography a quantum computer will break
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  ShorStop scans code, certificates, server configuration and dependencies for RSA, elliptic curves and Diffie-Hellman, and for crypto that's already broken. It writes a CycloneDX CBOM and SARIF for GitHub code scanning.
{{< /hextra/hero-subtitle >}}

{{< hextra/hero-button text="Get started" link="docs/getting-started/" >}}
{{< hextra/hero-button text="Rules reference" link="docs/rules/" style="background: transparent; color: inherit; border: 1px solid currentColor;" >}}

![ShorStop scanning a demo service: three critical findings (3DES, MD5, RSA-1024) and ten quantum-vulnerable ones, failing CI with --fail-on critical, and writing a CycloneDX CBOM](demo.gif)

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Five detectors, 225 rules"
    subtitle="Parsed Go, patterns for 11 more languages, certificates and keys, TLS and SSH configuration, and 8 package ecosystems."
    link="docs/rules/"
  >}}
  {{< hextra/feature-card
    title="A rating and a replacement"
    subtitle="Each finding says whether it's broken today or by a quantum computer, how urgent it is, and what to migrate to: ML-KEM, ML-DSA, SLH-DSA."
    link="docs/ratings/"
  >}}
  {{< hextra/feature-card
    title="A CBOM, not just a list"
    subtitle="A CycloneDX 1.6 Cryptographic Bill of Materials, validated against the official schema, for your inventory and auditors."
    link="docs/cbom/"
  >}}
  {{< hextra/feature-card
    title="Built for CI"
    subtitle="A GitHub Action with Sigstore-verified installs and code scanning upload, a fail-on threshold, baselines for existing code, and inline suppressions."
    link="docs/ci/"
  >}}
  {{< hextra/feature-card
    title="One static binary"
    subtitle="No network calls, no telemetry. About a second to scan the Kubernetes source on 8 cores. Homebrew, Docker, signed binaries or go install."
    link="docs/getting-started/"
  >}}
  {{< hextra/feature-card
    title="Quantum-safe counts too"
    subtitle="ML-KEM, ML-DSA and hybrid TLS groups like X25519MLKEM768 are reported as quantum-safe, so the scan shows migration progress."
    link="docs/ratings/"
  >}}
{{< /hextra/feature-grid >}}
