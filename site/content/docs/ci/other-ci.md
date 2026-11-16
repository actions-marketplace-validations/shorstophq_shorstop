---
description: >-
  Run ShorStop in GitLab CI, with Docker, or on any CI runner with a checksum-verified install script.
title: Other CI systems
weight: 2
---

Any CI system that runs shell commands can run ShorStop: install the binary,
scan, and act on the [exit code](../). Keep the reports as artifacts.

## Portable install

For a Linux runner with `curl` and `sha256sum`. It downloads a pinned release
and checks it against the release's checksums before using it:

```sh {filename="install-shorstop.sh"}
set -eu
V=0.1.3
case "$(uname -m)" in x86_64|amd64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; *) echo "unsupported: $(uname -m)" >&2; exit 2 ;; esac
base="https://github.com/shorstophq/shorstop/releases/download/v$V"
curl -fsSLO "$base/shorstop_${V}_linux_${ARCH}.tar.gz"
curl -fsSLO "$base/checksums.txt"
grep " shorstop_${V}_linux_${ARCH}.tar.gz\$" checksums.txt | sha256sum -c -
tar -xzf "shorstop_${V}_linux_${ARCH}.tar.gz" shorstop
./shorstop version
```

This checks the download against `checksums.txt`, fetched over HTTPS from the
same release. To also check that `checksums.txt` came from ShorStop's release
workflow, verify its Sigstore signature with cosign as shown in
[Getting started](../../getting-started/#install) (Binaries tab).

## GitLab CI

```yaml {filename=".gitlab-ci.yml"}
shorstop:
  image: debian:bookworm-slim
  before_script:
    - apt-get update -qq && apt-get install -y -qq curl ca-certificates > /dev/null
    - sh ci/install-shorstop.sh        # the portable install above
  script:
    - ./shorstop scan . --format cbom --output cbom.json
    - ./shorstop scan . --format markdown --output shorstop.md --fail-on high
  artifacts:
    when: always
    paths: [cbom.json, shorstop.md]
```

The first scan writes the full inventory and never fails the job. The second
writes a readable report and fails the job on high or critical findings;
`when: always` keeps both reports even when it does.

{{< callout type="info" >}}
The ShorStop container image can't be a GitLab job `image:`. It's built `FROM
scratch` and has no shell, and GitLab runs job scripts in a shell.
{{< /callout >}}

## With Docker

Where the CI system runs containers directly, as a step rather than as a job's
shell:

```sh
docker run --rm -v "$PWD:/src:ro" ghcr.io/shorstophq/shorstop:v0.1.3 scan --fail-on high
```

The image runs as a non-root user and reads the mounted directory. To write a
report, redirect standard output, since the mount is read-only:

```sh
docker run --rm -v "$PWD:/src:ro" ghcr.io/shorstophq/shorstop:v0.1.3 scan --format sarif > shorstop.sarif
```
