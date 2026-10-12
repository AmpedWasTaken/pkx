<p align="center"><img src="assets/pkx-mark.svg" width="78" height="78" alt="pkx logo"></p>

<h1 align="center">pkx — Package eXchange</h1>
<p align="center"><strong>One command. Every ecosystem.</strong></p>

<p align="center">
  <a href="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/AmpedWasTaken/pkx/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/AmpedWasTaken/pkx"></a>
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-blue"></a>
  <a href="CONTRIBUTING.md"><img alt="Contributions welcome" src="https://img.shields.io/badge/contributions-welcome-54e2d1"></a>
</p>

**pkx** is a cross-platform, open-source CLI that finds package managers in your repository and uses the native tools you already know. No custom dependency resolver, replacement lockfile, or required account.

## Install in one line

Full instructions and copyable commands: **[pkx.ampedwastaken.dev](https://pkx.ampedwastaken.dev)** ([website source](https://github.com/AmpedWasTaken/pkx-website)).

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
pkx add axios                 # Add a dependency to a Node project
pkx add vite --dev            # Add development dependency
pkx run dev                   # Run a declared project script
pkx doctor                    # Diagnose missing tools
pkx install --dry-run         # Preview without making changes
```

**Supported installation managers:** npm, pnpm, Yarn, Bun, pip, uv, Poetry, Composer, Cargo, and Go modules. `add`, `remove`, and `run` currently support Node managers and Composer. See [compatibility details](docs/GETTING_STARTED.md#supported-managers-and-behavior).

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

**[Contributing guide](CONTRIBUTING.md)** · **[Good first issues](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)** · **[Architecture](docs/ARCHITECTURE.md)** · **[Changelog](CHANGELOG.md)** · **[Security](SECURITY.md)**

---
<sub>Made by <a href="https://github.com/AmpedWasTaken">AmpedWasTaken</a> and open-source contributors.</sub>
