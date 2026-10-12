<p align="center"><img src="assets/pkx-mark.svg" width="78" height="78" alt="pkx logo"></p>

<h1 align="center">pkx — Package eXchange</h1>
<p align="center"><strong>One command. Every ecosystem.</strong></p>

<p align="center">
  <a href="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/AmpedWasTaken/pkx/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/AmpedWasTaken/pkx"></a>
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-blue"></a>
  <a href="CONTRIBUTING.md"><img alt="Contributions welcome" src="https://img.shields.io/badge/contributions-welcome-54e2d1"></a>
</p>

**pkx** is a cross-platform, open-source CLI that finds package managers in mixed-language repositories and runs your existing native tools. One setup workflow; no new lockfile, account or custom dependency solver.

## Install in one line

**Safer manual option (no pipe-to-shell):** Download your OS archive **and `SHA256SUMS`** from [the official GitHub release](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0). Compare the archive's SHA256 with its matching checksum entry, extract `pkx` / `pkx.exe` and add its folder to your user `PATH`. [Full steps and troubleshooting](docs/GETTING_STARTED.md).

**Convenience installers:** [pkx.ampedwastaken.dev](https://pkx.ampedwastaken.dev) ([website source](https://github.com/AmpedWasTaken/pkx-website)). **Review the scripts before running them**: checksums verify the downloaded archive against the published checksum list, but they do not independently authenticate the remotely served installer or prevent a compromised release account from changing both files.

**Windows — CMD or PowerShell**

```cmd
powershell -NoProfile -ExecutionPolicy Bypass -Command "& ([scriptblock]::Create((irm 'https://pkx.ampedwastaken.dev/install.ps1'))) -AddToPath"
```

**macOS / Linux**

```sh
curl -fsSL https://pkx.ampedwastaken.dev/install.sh | sh -s -- --add-to-path
```

The installers fetch the official release binaries, check their SHA256 against the release checksum file, and set up a **user-level PATH**. Close the terminal and open a new one:

```sh
pkx version
```

pkx itself doesn't need Go or npm. Your projects still need their underlying package managers. **Review remotely hosted installers before executing them** or [download and verify a release manually](https://github.com/AmpedWasTaken/pkx/releases). See [installation and PATH troubleshooting](docs/GETTING_STARTED.md).

## Get started

Run these commands from the root of any supported repository:

```sh
pkx scan          # Detect manifests and package managers
pkx status        # Check your environment
pkx plan          # Preview exactly what will execute
pkx up            # Install dependencies using native tools
```

For everyday work:

```sh
pkx install http-server       # Detect language/manager and add package in this project
pkx install http-server --dry-run  # Preview the native command first
pkx add vite --dev            # Add development dependency
pkx run dev                   # Run a declared project script
pkx doctor                    # Diagnose missing tools
pkx install --dry-run         # Preview without making changes
```

**Current scope (important):**

| Capability | Supported managers |
| --- | --- |
| **Install / `pkx up`** | **10:** npm, pnpm, Yarn, Bun, pip, uv, Poetry, Composer, Cargo, Go modules |
| **`add` / `remove` / `run`** | **5:** npm, pnpm, Yarn, Bun, Composer (Node + PHP only) |

**Unreleased v0.6 development:** `pkx install PACKAGE` detects the target project and supports native package additions across Node, Composer, pip, uv, Poetry, Cargo and Go. For monorepos use `--project DIRECTORY`; there are **no required ecosystem prefixes**. In the published v0.5.0, `pkx install` still only bootstraps project dependencies. 

**Python, Rust and Go `add/remove/run` are not supported yet.** See [compatibility and limitations](docs/GETTING_STARTED.md#supported-managers-and-behavior). Marketplace discovery is in unreleased v0.6 development source; the published v0.5.0 binary does **not** have `pkx bundle search/info/fetch`.

> pkx invokes native package managers, which may execute third-party lifecycle scripts. Always inspect untrusted repositories and use `pkx plan` first. Runtime installation and the remote bundle marketplace are not implemented yet.

## Try the example project

Don't want to point pkx at a real repository yet? [`examples/mixed-language`](examples/mixed-language/README.md) is a small, committed fixture with a Node (`package.json`) and a PHP (`composer.json`) project side by side, no credentials and no generated dependency directories. Run `pkx scan`, `pkx status` and `pkx plan` from that directory to see what pkx detects and exactly what it would install.

## Roadmap

<p align="center">
  <a href="docs/ROADMAP.md"><img src="assets/roadmap.svg" width="100%" alt="pkx roadmap: core CLI shipped; reliability, developer experience, bundle marketplace and runtime bootstrap planned"></a>
</p>

The roadmap uses qualitative milestones, not made-up completion percentages. [See the full roadmap](docs/ROADMAP.md) or [help build an upcoming feature](https://github.com/AmpedWasTaken/pkx/issues).

## Contribute

pkx is MIT-licensed and community-driven. Contributions to tests, platform support, adapters, documentation and usability are welcome.

**[Contributing guide](CONTRIBUTING.md)** · **[Become an adapter author](examples/adapter-npm/README.md)** · **[Good first issues](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)** · **[Architecture](docs/ARCHITECTURE.md)** · **[Changelog](CHANGELOG.md)** · **[Security](SECURITY.md)**

---
<sub>Made by <a href="https://github.com/AmpedWasTaken">AmpedWasTaken</a> and open-source contributors.</sub>
